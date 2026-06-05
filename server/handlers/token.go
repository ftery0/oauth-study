package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// tokenError: OAuth 2.0 RFC 6749 Section 5.2 - 토큰 엔드포인트 에러는 JSON으로 반환
func tokenError(w http.ResponseWriter, errorCode, description string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	// 클라이언트 앱은 HTTP Basic Auth로 자신을 증명
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		AuditWarn(r, "token.client_auth_missing")
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}
	client, ok := store.Clients.GetByClientID(clientID)
	if !ok || !store.VerifySecret(client, clientSecret) {
		AuditWarn(r, "token.client_auth_failed", "client_id", clientID)
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}
	_ = client // 검증 완료

	switch r.FormValue("grant_type") {
	case "authorization_code":
		handleAuthorizationCode(w, r, clientID)
	case "refresh_token":
		handleRefreshToken(w, r, clientID)
	default:
		AuditWarn(r, "token.unsupported_grant_type", "client_id", clientID, "grant_type", r.FormValue("grant_type"))
		tokenError(w, "unsupported_grant_type", "지원하지 않는 grant_type", http.StatusBadRequest)
	}
}

func handleAuthorizationCode(w http.ResponseWriter, r *http.Request, clientID string) {
	code        := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || redirectURI == "" {
		tokenError(w, "invalid_request", "code 또는 redirect_uri가 필요합니다", http.StatusBadRequest)
		return
	}

	// LoadAndDelete: 꺼내는 동시에 삭제 → auth code 재사용 방지
	val, ok := store.AuthCodes.LoadAndDelete(code)
	if !ok {
		tokenError(w, "invalid_grant", "유효하지 않은 code", http.StatusBadRequest)
		return
	}

	ac := val.(models.AuthCode)
	if time.Now().After(ac.ExpiresAt) || ac.ClientID != clientID || ac.RedirectURI != redirectURI {
		tokenError(w, "invalid_grant", "code 검증 실패", http.StatusBadRequest)
		return
	}

	// PKCE (RFC 7636): /authorize 에서 challenge 가 있었다면 verifier 검증 필수.
	if ac.CodeChallenge != "" {
		if codeVerifier == "" {
			tokenError(w, "invalid_request", "code_verifier 필요", http.StatusBadRequest)
			return
		}
		if !verifyPKCE(ac.CodeChallenge, ac.CodeChallengeMethod, codeVerifier) {
			tokenError(w, "invalid_grant", "PKCE 검증 실패", http.StatusBadRequest)
			return
		}
	}

	auditTokenIssued(r, "authorization_code", ac.UserID, ac.ClientID, ac.Scope)
	issueTokens(w, ac.UserID, ac.ClientID, ac.Scope, ac.Nonce, ac.AuthTime)
}

// verifyPKCE: S256(verifier) == challenge.
// method 가 plain (또는 비어있음) 일 때는 그대로 비교. 학습용 호환성 — 운영에선 S256 강제 권장.
func verifyPKCE(challenge, method, verifier string) bool {
	switch method {
	case "S256":
		sum := sha256.Sum256([]byte(verifier))
		got := base64.RawURLEncoding.EncodeToString(sum[:])
		return got == challenge
	case "plain", "":
		return verifier == challenge
	default:
		return false
	}
}

func handleRefreshToken(w http.ResponseWriter, r *http.Request, clientID string) {
	rtStr := r.FormValue("refresh_token")

	// LoadAndDelete: refresh token도 1회용으로 처리 (Token Rotation)
	// 매 갱신마다 새 refresh token을 발급 → 탈취된 토큰 감지 가능
	val, ok := store.RefreshTokens.LoadAndDelete(rtStr)
	if !ok {
		tokenError(w, "invalid_grant", "유효하지 않은 refresh_token", http.StatusBadRequest)
		return
	}

	rt := val.(models.RefreshToken)
	if time.Now().After(rt.ExpiresAt) || rt.ClientID != clientID {
		http.Error(w, "refresh_token 검증 실패", http.StatusBadRequest)
		return
	}

	auditTokenIssued(r, "refresh_token", rt.UserID, rt.ClientID, rt.Scope)
	// refresh 시점엔 신선한 nonce 없음. auth_time 도 그대로 유지 (이 grant 에선 시점 기록 X).
	issueTokens(w, rt.UserID, rt.ClientID, rt.Scope, "", time.Time{})
}

// issueTokens: access token + refresh token 동시 발급.
// scope 에 openid 가 있으면 ID Token 도 같이 발급 (P3.1).
func issueTokens(w http.ResponseWriter, userID, clientID, scope, nonce string, authTime time.Time) {
	accessToken, err := token.Create(userID, clientID, scope)
	if err != nil {
		http.Error(w, "액세스 토큰 생성 실패", http.StatusInternalServerError)
		return
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		http.Error(w, "리프레시 토큰 생성 실패", http.StatusInternalServerError)
		return
	}

	// Refresh token은 장기 유효 (7일), access token보다 훨씬 긺
	store.RefreshTokens.Store(refreshToken, models.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ClientID:  clientID,
		Scope:     scope,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	resp := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    900, // 15분 (초 단위)
		"refresh_token": refreshToken,
		"scope":         scope,
	}

	// OIDC: openid scope 가 있으면 ID Token 발급.
	if hasOpenIDScope(scope) {
		idToken, err := token.CreateIDToken(userID, clientID, nonce, authTime)
		if err != nil {
			http.Error(w, "ID 토큰 생성 실패", http.StatusInternalServerError)
			return
		}
		resp["id_token"] = idToken
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// auditTokenIssued: handleAuthorizationCode / handleRefreshToken 에서 발급 직후 호출.
func auditTokenIssued(r *http.Request, grant, userID, clientID, scope string) {
	AuditEvent(r, "token.issued",
		"grant", grant,
		"sub", userID,
		"client_id", clientID,
		"scope", scope,
	)
}

func hasOpenIDScope(scope string) bool {
	for _, s := range splitScope(scope) {
		if s == "openid" {
			return true
		}
	}
	return false
}

func splitScope(scope string) []string {
	out := []string{}
	start := 0
	for i := 0; i <= len(scope); i++ {
		if i == len(scope) || scope[i] == ' ' {
			if i > start {
				out = append(out, scope[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
