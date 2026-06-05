package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ftery0/ouath/server/config"
)

// DiscoveryHandler: GET /.well-known/openid-configuration (OIDC Discovery 1.0).
// 표준 OIDC 클라이언트 (NextAuth, Spring Security 등) 가 자동 설정에 사용.
func DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	issuer := config.IssuerForDiscovery()

	resp := map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/oauth/token",
		"userinfo_endpoint":                     issuer + "/oauth/userinfo",
		"jwks_uri":                              issuer + "/oauth/jwks",
		"registration_endpoint":                 issuer + "/oauth/register",
		"revocation_endpoint":                   issuer + "/oauth/revoke",
		"introspection_endpoint":                issuer + "/oauth/introspect",
		"end_session_endpoint":                  issuer + "/oauth/logout",
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
		"code_challenge_methods_supported":      []string{"S256"},
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce",
			"name", "preferred_username", "email", "email_verified",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(resp)
}
