package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// IntrospectHandler: POST /oauth/introspect (RFC 7662).
// form: token, token_type_hint=access_token|refresh_token (선택).
// 클라이언트 인증 필수. 응답: {active: bool, sub, client_id, scope, exp, iat, token_type}
// active=false 케이스에서는 active 만 반환 (정보 노출 방지).
func IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}
	client, ok := store.Clients.GetByClientID(clientID)
	if !ok || !store.VerifySecret(client, clientSecret) {
		tokenError(w, "invalid_client", "client 인증 실패", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		tokenError(w, "invalid_request", "form parse 실패", http.StatusBadRequest)
		return
	}
	tokStr := r.FormValue("token")
	hint := r.FormValue("token_type_hint")

	resp := map[string]any{"active": false}
	w.Header().Set("Content-Type", "application/json")

	tryAccess := func() bool {
		if IsAccessTokenRevoked(tokStr) {
			return false
		}
		claims, err := token.Parse(tokStr)
		if err != nil {
			return false
		}
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return false
		}
		resp["active"] = true
		resp["sub"] = claims.UserID
		resp["client_id"] = claims.ClientID
		resp["scope"] = claims.Scope
		resp["token_type"] = "Bearer"
		if claims.ExpiresAt != nil {
			resp["exp"] = claims.ExpiresAt.Unix()
		}
		if claims.IssuedAt != nil {
			resp["iat"] = claims.IssuedAt.Unix()
		}
		return true
	}

	tryRefresh := func() bool {
		val, ok := store.RefreshTokens.Load(tokStr)
		if !ok {
			return false
		}
		rt, ok := val.(models.RefreshToken)
		if !ok {
			return false
		}
		if rt.ExpiresAt.Before(time.Now()) {
			return false
		}
		resp["active"] = true
		resp["sub"] = rt.UserID
		resp["client_id"] = rt.ClientID
		resp["scope"] = rt.Scope
		resp["token_type"] = "refresh_token"
		resp["exp"] = rt.ExpiresAt.Unix()
		return true
	}

	switch hint {
	case "refresh_token":
		if !tryRefresh() {
			_ = tryAccess()
		}
	default:
		if !tryAccess() {
			_ = tryRefresh()
		}
	}

	json.NewEncoder(w).Encode(resp)
}
