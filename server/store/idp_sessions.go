package store

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/ftery0/ouath/server/models"
)

// IdPSessionTTL: 사용자별 IdP 세션 유효시간.
// 너무 길면 탈취 위험, 너무 짧으면 사용자 불편. 학습용 8시간.
const IdPSessionTTL = 8 * time.Hour

// idpSessionStore: 단순 map + Mutex.
//
// sync.Map 이 아니라 Mutex 를 쓰는 이유:
//   - Get → expired 검사 → 삭제 같은 조합 연산이 atomic 해야 한다
//   - 이걸 sync.Map 단독으로 하면 race 로 좀비 세션 부활 가능
type idpSessionStore struct {
	mu sync.Mutex
	m  map[string]*models.IdPSession
}

// IdPSessions: 패키지 진입점. main 에서 StartCleanup 으로 청소 goroutine 기동.
var IdPSessions = &idpSessionStore{m: make(map[string]*models.IdPSession)}

// Create: 새 sessionID 발급 + 저장. 세션 고정 공격 방어를 위해 매 로그인마다 호출.
// 기존 sid 가 있었다면 호출자(login handler) 가 Delete 로 명시적으로 폐기해야 한다.
func (s *idpSessionStore) Create(userID, groupID string) (string, error) {
	sid, err := randomHex32()
	if err != nil {
		return "", err
	}
	now := time.Now()
	s.mu.Lock()
	s.m[sid] = &models.IdPSession{
		SessionID:   sid,
		UserID:      userID,
		LastGroupID: groupID,
		LoginAt:     now,
		ExpiresAt:   now.Add(IdPSessionTTL),
	}
	s.mu.Unlock()
	return sid, nil
}

// Get: sid 로 세션 조회. 만료된 세션은 자동으로 제거하고 (nil, false) 반환.
// 복사본을 반환해 호출자의 변형이 store 를 오염시키지 않도록 한다.
func (s *idpSessionStore) Get(sid string) (*models.IdPSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.m[sid]
	if !ok {
		return nil, false
	}
	if sess.Expired() {
		delete(s.m, sid)
		return nil, false
	}
	clone := *sess
	return &clone, true
}

// Touch: 같은 sid 의 LastGroupID 와 만료시간을 갱신.
// 같은 사용자가 다른 그룹의 앱에서 다시 로그인할 때 사용.
func (s *idpSessionStore) Touch(sid, lastGroupID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.m[sid]; ok && !sess.Expired() {
		sess.LastGroupID = lastGroupID
		sess.ExpiresAt = time.Now().Add(IdPSessionTTL)
	}
}

// Delete: 명시적 세션 폐기 (logout / 세션 고정 방어).
func (s *idpSessionStore) Delete(sid string) {
	s.mu.Lock()
	delete(s.m, sid)
	s.mu.Unlock()
}

// SweepExpired: 만료된 세션 일괄 정리. 청소 goroutine 이 주기적으로 호출.
func (s *idpSessionStore) SweepExpired() int {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for sid, sess := range s.m {
		if now.After(sess.ExpiresAt) {
			delete(s.m, sid)
			removed++
		}
	}
	return removed
}

// StartCleanup: 5분 주기로 SweepExpired 실행하는 goroutine 기동.
// main 에서 한 번 호출. 학습용으로 context 종료는 따로 처리하지 않는다.
func (s *idpSessionStore) StartCleanup() {
	go func() {
		t := time.NewTicker(5 * time.Minute)
		defer t.Stop()
		for range t.C {
			if n := s.SweepExpired(); n > 0 {
				log.Printf("[idp_sessions] swept %d expired sessions", n)
			}
		}
	}()
}

func randomHex32() (string, error) {
	b := make([]byte, 32) // 32바이트 = 256bit 엔트로피
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
