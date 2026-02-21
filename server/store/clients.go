package store

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ftery0/ouath/server/models"
)

var (
	Clients     = &clientStore{byID: make(map[string]*models.Client)}
	clientMutex sync.RWMutex
)

type clientStore struct {
	byID map[string]*models.Client // key: client_id (OAuth client identifier)
}

func init() {
	// 개발/테스트용 시드 데이터 (나중에 웹 등록으로 대체)
	Clients.register(&models.Client{
		ID:           "seed-example-app",
		ClientID:     "example-app",
		ClientSecret: "secret",
		Name:         "Example App",
		Description:  "OAuth 테스트용 예제 앱",
		MainURL:      "http://localhost:8081",
		ServerURLs:   []string{"http://localhost:8081"},
		RedirectURIs: []string{"http://localhost:8081/callback"},
		OwnerID:      "",
		CreatedAt:    time.Now(),
	})
}

// GetByClientID: client_id로 클라이언트 조회
func (s *clientStore) GetByClientID(clientID string) (*models.Client, bool) {
	clientMutex.RLock()
	defer clientMutex.RUnlock()
	c, ok := s.byID[clientID]
	if !ok {
		return nil, false
	}
	return c, true
}

// register: 클라이언트 등록 (내부용, 추후 웹 등록 API에서 호출)
func (s *clientStore) register(c *models.Client) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if c.ID == "" {
		c.ID = "client-" + randomHex(8)
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	s.byID[c.ClientID] = c
}

// Register: 새 클라이언트 등록 (웹 등록 API에서 사용)
func (s *clientStore) Register(c *models.Client) error {
	if c.ClientID == "" {
		c.ClientID = "client-" + randomHex(8)
	}
	if c.ClientSecret == "" {
		c.ClientSecret = randomHex(32)
	}
	s.register(c)
	return nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
