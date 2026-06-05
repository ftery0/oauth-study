package models

import "time"

type AuthCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
	ExpiresAt   time.Time

	// PKCE (RFC 7636) — public client 의 code interception 방어용.
	// 클라이언트가 /authorize 에 code_challenge 를 제공한 경우에만 채워짐.
	// /token 에서 code_verifier 를 받아 S256(verifier) == challenge 검증.
	CodeChallenge       string
	CodeChallengeMethod string // S256 만 허용 (plain 거부)

	// Nonce (OIDC) — ID Token replay 방어. /authorize 시점에 클라이언트가 제공,
	// /token 응답의 ID Token claim 으로 그대로 echo.
	Nonce string

	// AuthTime — 사용자가 IdP 에서 실제 로그인 (또는 silent SSO 통과) 한 시각.
	// ID Token auth_time claim 으로 전달.
	AuthTime time.Time
}
