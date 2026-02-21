package handlers

import (
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

	if r.FormValue("grant_type") != "authorization_code" {
		tokenError(w, "unsupported_grant_type", "지원하지 않는 grant_type", http.StatusBadRequest)
		return
	}

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

	accessToken, err := token.Create(ac.UserID, ac.ClientID, ac.Scope)
	if err != nil {
		tokenError(w, "server_error", "토큰 생성 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   900,
	})
}
