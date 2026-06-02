package store

import (
	"sync"
	"time"

	"github.com/ftery0/ouath/server/models"
)

var Groups = &groupStore{byID: make(map[string]*models.ProjectGroup)}

type groupStore struct {
	mu   sync.RWMutex
	byID map[string]*models.ProjectGroup
}

// Get: 그룹 ID 로 조회. 미존재 시 nil + false.
func (s *groupStore) Get(id string) (*models.ProjectGroup, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.byID[id]
	if !ok {
		return nil, false
	}
	return g, true
}

// Register: 그룹 등록. CreatedAt 자동 채움.
func (s *groupStore) Register(g *models.ProjectGroup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now()
	}
	s.byID[g.ID] = g
}

// 학습/데모용 시드 그룹. 그룹명은 도메인 의미 없는 generic 이름으로 둔다.
// 추후 phase 의 어드민 콘솔이 도입되면 동적 등록으로 대체된다.
func init() {
	Groups.Register(&models.ProjectGroup{
		ID:          "group-a",
		Name:        "그룹 A",
		Description: "그룹 내 앱들은 silent SSO 가 기본 ON (app1, app2 가 소속)",
		SSODefault:  models.SSODefaultON,
	})
	Groups.Register(&models.ProjectGroup{
		ID:          "group-b",
		Name:        "그룹 B",
		Description: "그룹 내 앱들은 기본적으로 매번 로그인 필요 (app3 가 소속)",
		SSODefault:  models.SSODefaultOFF,
	})
}
