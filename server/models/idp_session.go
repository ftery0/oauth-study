package models

import "time"

// IdPSession: IdP 가 사용자별로 유지하는 글로벌 세션.
// SSO 의 핵심 invariant 가 이 구조체에 담긴다:
//
//   - UserID    : silent SSO 시 발급될 auth code 의 주인이 누구인가
//   - LastGroupID: cross-group silent 차단 키 — 세션이 만들어진 그룹과
//     요청 client.GroupID 가 일치할 때만 silent 통과
//   - ExpiresAt : sweep goroutine 이 메모리 누수를 막는 기준
type IdPSession struct {
	SessionID   string    // crypto/rand 32바이트 (256bit) hex
	UserID      string    // 예: "alice"
	LastGroupID string    // 직전 로그인 시점의 그룹 ID
	LoginAt     time.Time
	ExpiresAt   time.Time
}

// Expired: 만료 여부 확인.
func (s *IdPSession) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}
