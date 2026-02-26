package handlers

import (
	"crypto/rand"
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
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}
	client, ok := store.Clients.GetByClientID(clientID)
	if !ok || client.ClientSecret != clientSecret {
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}
	_ = client // 검증 완료

	switch r.FormValue("grant_type") {
	case "authorization_code":
		handleAuthorizationCode(w, r, clientID)
	case "refresh_token":
		// Refresh Token Grant: 만료된 access token을 refresh token으로 갱신
		// 장점: 사용자가 다시 로그인하지 않아도 됨
		handleRefreshToken(w, r, clientID)
	default:
		tokenError(w, "unsupported_grant_type", "지원하지 않는 grant_type", http.StatusBadRequest)
	}
}

func handleAuthorizationCode(w http.ResponseWriter, r *http.Request, clientID string) {
	code        := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")

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

	issueTokens(w, ac.UserID, ac.ClientID, ac.Scope)
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

	issueTokens(w, rt.UserID, rt.ClientID, rt.Scope)
}

// issueTokens: access token + refresh token 동시 발급
func issueTokens(w http.ResponseWriter, userID, clientID, scope string) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    900, // 15분 (초 단위)
		"refresh_token": refreshToken,
	})
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
