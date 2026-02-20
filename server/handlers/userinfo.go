package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ftery0/ouath/server/token"
)

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Authorization: Bearer <token> 에서 토큰 부분만 추출
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sub":       claims.UserID,
		"client_id": claims.ClientID,
		"scope":     claims.Scope,
	})
}
