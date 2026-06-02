package store

import (
	"errors"
	"sync"
	"time"

	"github.com/ftery0/ouath/server/models"
)

// errGroupNotFound: 인메모리 구현체에서 그룹 미존재 시 반환.
var errGroupNotFound = errors.New("group not found")

// defaultMemoryGroups: 인메모리 기본 인스턴스 (시드 init() 적용 대상).
var defaultMemoryGroups = &groupStore{byID: make(map[string]*models.ProjectGroup)}

// Groups: 외부 노출. main 이 DATABASE_URL 보고 Postgres 구현체로 교체할 수 있다.
var Groups GroupStore = defaultMemoryGroups

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

// All: 모든 그룹 슬라이스로 반환 (어드민 read 용도).
func (s *groupStore) All() []*models.ProjectGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.ProjectGroup, 0, len(s.byID))
	for _, g := range s.byID {
		out = append(out, g)
	}
	return out
}

// UpdateSSODefault: 기존 그룹의 SSO 기본 정책 변경.
func (s *groupStore) UpdateSSODefault(id string, ssoDefault models.SSODefault) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.byID[id]
	if !ok {
		return errGroupNotFound
	}
	g.SSODefault = ssoDefault
	return nil
}

// Register: 그룹 등록. CreatedAt 자동 채움.
func (s *groupStore) Register(g *models.ProjectGroup) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now()
	}
	s.byID[g.ID] = g
	return nil
}

// 학습/데모용 시드 그룹. 인메모리 인스턴스에만 자동 적용.
// Postgres 사용 시 main 에서 SeedGroups 호출.
func init() {
	defaultMemoryGroups.Register(seedGroupA())
	defaultMemoryGroups.Register(seedGroupB())
}

// SeedGroups: 시드 그룹들을 임의의 GroupStore 에 등록 (Postgres SeedIfEmpty 용).
func SeedGroups(s GroupStore) error {
	for _, g := range []*models.ProjectGroup{seedGroupA(), seedGroupB()} {
		if err := s.Register(g); err != nil {
			return err
		}
	}
	return nil
}

func seedGroupA() *models.ProjectGroup {
	return &models.ProjectGroup{
		ID:          "group-a",
		Name:        "그룹 A",
		Description: "그룹 내 앱들은 silent SSO 가 기본 ON (app1, app2 가 소속)",
		SSODefault:  models.SSODefaultON,
	}
}

func seedGroupB() *models.ProjectGroup {
	return &models.ProjectGroup{
		ID:          "group-b",
		Name:        "그룹 B",
		Description: "그룹 내 앱들은 기본적으로 매번 로그인 필요 (app3 가 소속)",
		SSODefault:  models.SSODefaultOFF,
	}
}
