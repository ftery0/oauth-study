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

// UserInfoHandler: /oauth/userinfo.
//
// 응답 필드:
//   - sub                 : 사용자 ID (UUID). OIDC 표준 필드
//   - preferred_username  : users.username.        OIDC 표준 (scope=profile 일 때만)
//   - name                : users.display_name.    OIDC 표준 (scope=profile 일 때만)
//   - client_id, scope    : 비표준 (학습용 디버그). 토큰 claim 을 그대로 노출
//
// scope 에 "profile" 이 포함되지 않으면 username/name 은 응답하지 않는다 (OIDC 표준 동작).
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		http.Error(w, "Authorization 헤더 없음", http.StatusUnauthorized)
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

	if hasProfileScope(claims.Scope) && store.Users != nil {
		u, err := store.Users.GetByID(r.Context(), claims.UserID)
		if err == nil && u != nil {
			resp["preferred_username"] = u.Username
			resp["name"] = u.DisplayName
		} else if err != nil && !errors.Is(err, store.ErrUserNotFound) {
			log.Printf("[userinfo] GetByID failed sub=%s err=%v", claims.UserID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func hasProfileScope(scope string) bool {
	for _, s := range strings.Fields(scope) {
		if s == "profile" {
			return true
		}
	}
	return false
}
