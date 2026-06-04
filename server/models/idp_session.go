package models

import "time"

// IdPSession: IdP 가 사용자별로 유지하는 글로벌 세션.
//
// Phase-R 단순화: LastGroupID 필드 제거 (글로벌 user pool 로 가면서
// cross-group 차단 개념이 사라짐). silent SSO 는 client 단위 토글로 제어.
type IdPSession struct {
	SessionID string    // crypto/rand 32바이트 (256bit) hex
	UserID    string    // users.id (UUID)
	LoginAt   time.Time
	ExpiresAt time.Time
}

// Expired: 만료 여부 확인.
func (s *IdPSession) Expired() bool {
	return time.Now().After(s.ExpiresAt)
}
