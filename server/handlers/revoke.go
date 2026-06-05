package handlers

import (
	"net/http"
	"sync"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// revokedAccessTokens: access token 이 JWT 라 stateless 검증이지만,
// revoke 호출 시점부터 만료(<=15분)까지는 블로킹 필요. 메모리 set 으로 충분 (TTL = 만료시각).
var (
	revokedAccessTokens   = make(map[string]struct{})
	revokedAccessTokensMu sync.RWMutex
)

// IsAccessTokenRevoked: userinfo/introspect 가 부르는 헬퍼.
func IsAccessTokenRevoked(tokenStr string) bool {
	revokedAccessTokensMu.RLock()
	defer revokedAccessTokensMu.RUnlock()
	_, ok := revokedAccessTokens[tokenStr]
	return ok
}

// RevokeHandler: POST /oauth/revoke (RFC 7009).
// form: token, token_type_hint=access_token|refresh_token (선택).
// 항상 200 반환 (정보 노출 방지). 클라이언트 인증 필수.
func RevokeHandler(w http.ResponseWriter, r *http.Request) {
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

	// refresh token 시도 → 본인 client 의 것만 삭제 (다른 client 가 남의 토큰 폐기 못 함)
	tryRefresh := func() bool {
		val, ok := store.RefreshTokens.Load(tokStr)
		if !ok {
			return false
		}
		rt, ok := val.(models.RefreshToken)
		if !ok || rt.ClientID != clientID {
			return false
		}
		store.RefreshTokens.Delete(tokStr)
		return true
	}

	// access token (JWT) 시도 → 본인 client 의 것이면 blocklist
	tryAccess := func() bool {
		claims, err := token.Parse(tokStr)
		if err != nil || claims.ClientID != clientID {
			return false
		}
		revokedAccessTokensMu.Lock()
		revokedAccessTokens[tokStr] = struct{}{}
		revokedAccessTokensMu.Unlock()
		return true
	}

	switch hint {
	case "refresh_token":
		if !tryRefresh() {
			_ = tryAccess()
		}
	case "access_token":
		if !tryAccess() {
			_ = tryRefresh()
		}
	default:
		if !tryRefresh() {
			_ = tryAccess()
		}
	}

	AuditEvent(r, "token.revoked", "client_id", clientID, "hint", hint)
	// RFC 7009: revoke 는 토큰 존재 여부와 무관하게 200
	w.WriteHeader(http.StatusOK)
}
