package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	// 학습/데모용 시드 client. 추후 phase 의 어드민 콘솔이 등록 UI 로 대체한다.
	// 이번 phase 의 핵심 시연은 다음 매핑:
	//   - app1, app2 ∈ group-a (SSO ON) → 같은 그룹 silent SSO 시연
	//   - app3       ∈ group-b (SSO OFF) → 다른 그룹은 다시 로그인 (Realm 경계)
	Clients.register(seedClient("app1", "App 1", "group-a", 8011, 5181))
	Clients.register(seedClient("app2", "App 2", "group-a", 8012, 5182))
	Clients.register(seedClient("app3", "App 3", "group-b", 8013, 5183))
}

// seedClient: 학습용 헬퍼. ClientSecret 도 식별 가능한 형태로 고정한다
// (다음 phase 의 어드민 UI 가 도입되면 secret 은 랜덤 생성으로 바뀐다).
func seedClient(id, name, groupID string, backendPort, frontendPort int) *models.Client {
	return &models.Client{
		ID:           "seed-" + id,
		ClientID:     id,
		ClientSecret: id + "-secret",
		Name:         name,
		Description:  "examples/" + id + " 예제 앱",
		MainURL:      fmt.Sprintf("http://localhost:%d", frontendPort),
		ServerURLs:   []string{fmt.Sprintf("http://localhost:%d", backendPort)},
		RedirectURIs: []string{fmt.Sprintf("http://localhost:%d/callback", backendPort)},
		OwnerID:      "",
		CreatedAt:    time.Now(),
		GroupID:      groupID,
		SSOOverride:  models.OverrideInherit,
	}
}

// GetByClientID: client_id 로 클라이언트 조회
func (s *clientStore) GetByClientID(clientID string) (*models.Client, bool) {
	clientMutex.RLock()
	defer clientMutex.RUnlock()
	c, ok := s.byID[clientID]
	if !ok {
		return nil, false
	}
	return c, true
}

// register: 클라이언트 등록 (내부용, 추후 웹 등록 API 에서 호출)
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

// Register: 새 클라이언트 등록 (웹 등록 API 에서 사용)
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

