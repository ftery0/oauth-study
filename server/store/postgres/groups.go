package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ftery0/ouath/server/models"
)

// GroupStore: pgxpool 기반 ProjectGroup 영속 구현체.
type GroupStore struct {
	pool *pgxpool.Pool
}

func NewGroupStore(pool *pgxpool.Pool) *GroupStore {
	return &GroupStore{pool: pool}
}

func (s *GroupStore) Get(id string) (*models.ProjectGroup, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, name, description, sso_default, created_at
		FROM groups WHERE id = $1
	`, id)

	g, err := scanGroup(row)
	if err != nil {
		return nil, false
	}
	return g, true
}

func (s *GroupStore) All() []*models.ProjectGroup {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, description, sso_default, created_at
		FROM groups ORDER BY created_at ASC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]*models.ProjectGroup, 0)
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			continue
		}
		out = append(out, g)
	}
	return out
}

// UpdateSSODefault: 기존 그룹의 SSO 기본 정책 변경.
func (s *GroupStore) UpdateSSODefault(id string, ssoDefault models.SSODefault) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := s.pool.Exec(ctx, `
		UPDATE groups SET sso_default = $1 WHERE id = $2
	`, string(ssoDefault), id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("group not found")
	}
	return nil
}

func (s *GroupStore) Register(g *models.ProjectGroup) error {
	if g.ID == "" {
		return errors.New("group id required")
	}
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO groups (id, name, description, sso_default, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, g.ID, g.Name, g.Description, string(g.SSODefault), g.CreatedAt)
	return err
}

func scanGroup(row pgx.Row) (*models.ProjectGroup, error) {
	var g models.ProjectGroup
	var ssoDefault string
	if err := row.Scan(
		&g.ID, &g.Name, &g.Description, &ssoDefault, &g.CreatedAt,
	); err != nil {
		return nil, err
	}
	g.SSODefault = models.SSODefault(ssoDefault)
	return &g, nil
}
