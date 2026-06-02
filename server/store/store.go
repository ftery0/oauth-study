// Package store 는 client / group / IdP session 의 영속 인터페이스와 구현체를 모은다.
//
// Phase 1: sync.Map / Mutex 기반 인메모리 (clients.go, groups.go, idp_sessions.go)
// Phase 2-C: 인터페이스 추출 + Postgres 구현체 (postgres/ 서브 패키지)
//
// 호출자(handlers, policy 등) 는 항상 인터페이스를 통해 접근한다.
package store

import "github.com/ftery0/ouath/server/models"

// ClientStore: OAuth client(서비스) 영속 인터페이스.
type ClientStore interface {
	GetByClientID(clientID string) (*models.Client, bool)
	All() []*models.Client
	Register(c *models.Client) error
}

// GroupStore: ProjectGroup 영속 인터페이스.
type GroupStore interface {
	Get(id string) (*models.ProjectGroup, bool)
	All() []*models.ProjectGroup
	Register(g *models.ProjectGroup) error
	UpdateSSODefault(id string, ssoDefault models.SSODefault) error
}

// 컴파일 타임 인터페이스 충족 검증.
// 새 구현체를 추가하면 여기에도 한 줄 더 추가해서 type-check 한다.
var (
	_ ClientStore = (*clientStore)(nil)
	_ GroupStore  = (*groupStore)(nil)
)
