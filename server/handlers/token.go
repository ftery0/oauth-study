package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// Phase 1 하드코딩 클라이언트 시크릿
var testClientSecret = "secret"

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	// 클라이언트 앱은 HTTP Basic Auth로 자신을 증명: Authorization: Basic base64(client_id:secret)
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok || clientSecret != testClientSecret {
		http.Error(w, "client 인증 실패", http.StatusUnauthorized)
		return
	}

	if r.FormValue("grant_type") != "authorization_code" {
		http.Error(w, "지원하지 않는 grant_type", http.StatusBadRequest)
		return
	}

	code        := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")

	// LoadAndDelete: 꺼내는 동시에 삭제 → auth code 재사용 방지
	val, ok := store.AuthCodes.LoadAndDelete(code)
	if !ok {
		http.Error(w, "유효하지 않은 code", http.StatusBadRequest)
		return
	}

	ac := val.(models.AuthCode)

	if time.Now().After(ac.ExpiresAt) || ac.ClientID != clientID || ac.RedirectURI != redirectURI {
		http.Error(w, "code 검증 실패", http.StatusBadRequest)
		return
	}

	accessToken, err := token.Create(ac.UserID, ac.ClientID, ac.Scope)
	if err != nil {
		http.Error(w, "토큰 생성 실패", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   900,
	})
}
