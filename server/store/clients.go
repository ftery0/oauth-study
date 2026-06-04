package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ftery0/ouath/server/models"
)

// defaultMemoryClients: 인메모리 기본 인스턴스. main 이 DATABASE_URL 보고
// Clients 를 Postgres 구현체로 교체할 수 있다. 시드 init() 은 이 인스턴스에만
// 적용되므로, Postgres 사용 시 별도 SeedIfEmpty 가 필요.
var (
	defaultMemoryClients = &clientStore{byID: make(map[string]*models.Client)}
	// Clients: 외부 노출. main 에서 교체 가능.
	Clients     ClientStore = defaultMemoryClients
	clientMutex sync.RWMutex
)

type clientStore struct {
	byID map[string]*models.Client // key: client_id (OAuth client identifier)
}

func init() {
	// 학습/데모용 시드 client. 인메모리 인스턴스에만 자동 적용.
	// Postgres 사용 시 main 에서 SeedClients 를 명시적으로 호출.
	defaultMemoryClients.register(seedClient("app1", "App 1", 8011, 5181))
	defaultMemoryClients.register(seedClient("app2", "App 2", 8012, 5182))
	defaultMemoryClients.register(seedClient("app3", "App 3", 8013, 5183))
}

// SeedClients: 시드 client 들을 임의의 ClientStore 에 등록한다 (Postgres SeedIfEmpty 용).
// 같은 client_id 가 이미 있으면 무시 (Register 가 ON CONFLICT DO NOTHING).
func SeedClients(s ClientStore) error {
	seeds := []*models.Client{
		seedClient("app1", "App 1", 8011, 5181),
		seedClient("app2", "App 2", 8012, 5182),
		seedClient("app3", "App 3", 8013, 5183),
	}
	for _, c := range seeds {
		if err := s.Register(c); err != nil {
			return err
		}
	}
	return nil
}

// seedClient: 학습용 헬퍼. ClientSecret 도 식별 가능한 형태로 고정한다
// (다음 phase 의 어드민 UI 가 도입되면 secret 은 랜덤 생성으로 바뀐다).
func seedClient(id, name string, backendPort, frontendPort int) *models.Client {
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
		SilentSSO:    true,
	}
}

// All: 모든 client 슬라이스로 반환 (어드민 read 용도).
func (s *clientStore) All() []*models.Client {
	clientMutex.RLock()
	defer clientMutex.RUnlock()
	out := make([]*models.Client, 0, len(s.byID))
	for _, c := range s.byID {
		out = append(out, c)
	}
	return out
}

// UpdateSilentSSO: silent_sso 토글 (인메모리).
func (s *clientStore) UpdateSilentSSO(clientID string, silentSSO bool) error {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	c, ok := s.byID[clientID]
	if !ok {
		return errors.New("client not found")
	}
	c.SilentSSO = silentSSO
	return nil
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

