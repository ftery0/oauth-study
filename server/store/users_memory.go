package store

import (
	"context"
	"sync"

	"github.com/ftery0/ouath/server/models"
)

// memoryUserStore: Postgres 가 없을 때 fallback + 테스트용 인메모리 UserStore.
// 시드 TestUsers (alice/bob/carol) 가 자동 등록된다.
type memoryUserStore struct {
	mu         sync.RWMutex
	byUsername map[string]*models.User
	byID       map[string]*models.User
}

var defaultMemoryUsers = &memoryUserStore{
	byUsername: make(map[string]*models.User),
	byID:       make(map[string]*models.User),
}

// init: store.Users 의 기본값을 인메모리로 설정 + 시드.
// main 이 Postgres 모드로 가면 store.Users 가 교체된다.
func init() {
	Users = defaultMemoryUsers
	for _, u := range models.TestUsers {
		cp := u
		if cp.ID == "" {
			cp.ID = "u-" + cp.Username
		}
		defaultMemoryUsers.byUsername[cp.Username] = &cp
		defaultMemoryUsers.byID[cp.ID] = &cp
	}
}

func (s *memoryUserStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byUsername[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (s *memoryUserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (s *memoryUserStore) Create(ctx context.Context, u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byUsername[u.Username]; exists {
		return ErrUserAlreadyExists
	}
	if u.ID == "" {
		u.ID = "u-" + u.Username
	}
	cp := *u
	s.byUsername[u.Username] = &cp
	s.byID[u.ID] = &cp
	return nil
}
