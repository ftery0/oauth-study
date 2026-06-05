package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// UserInfoHandler: /oauth/userinfo (OIDC Core §5.3).
//
// 응답 필드 (scope 별로 필터링):
//   - sub                 : 항상 — 사용자 ID (UUID)
//   - preferred_username  : scope=profile — users.username
//   - name                : scope=profile — users.display_name
//   - email               : scope=email — users.email
//   - email_verified      : scope=email — users.email_verified
//   - client_id, scope    : 비표준 (학습용 디버그). 토큰 claim 그대로
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		http.Error(w, "Authorization 헤더 없음", http.StatusUnauthorized)
		return
	}

	if IsAccessTokenRevoked(tokenStr) {
		http.Error(w, "토큰이 폐기됨", http.StatusUnauthorized)
		return
	}

	claims, err := token.Parse(tokenStr)
	if err != nil {
		http.Error(w, "유효하지 않은 토큰", http.StatusUnauthorized)
		return
	}

	resp := map[string]any{
		"sub":       claims.UserID,
		"client_id": claims.ClientID,
		"scope":     claims.Scope,
	}

	scopes := strings.Fields(claims.Scope)
	needsUserLookup := contains(scopes, "profile") || contains(scopes, "email")

	if needsUserLookup && store.Users != nil {
		u, err := store.Users.GetByID(r.Context(), claims.UserID)
		if err == nil && u != nil {
			if contains(scopes, "profile") {
				resp["preferred_username"] = u.Username
				resp["name"] = u.DisplayName
			}
			if contains(scopes, "email") && u.Email != "" {
				resp["email"] = u.Email
				resp["email_verified"] = u.EmailVerified
			}
		} else if err != nil && !errors.Is(err, store.ErrUserNotFound) {
			log.Printf("[userinfo] GetByID failed sub=%s err=%v", claims.UserID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}
