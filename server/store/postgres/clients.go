// Package postgres 는 store.ClientStore / store.UserStore 의 Postgres 구현체.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ftery0/ouath/server/models"
)

// ClientStore: pgxpool 기반 영속 구현체.
type ClientStore struct {
	pool *pgxpool.Pool
}

func NewClientStore(pool *pgxpool.Pool) *ClientStore {
	return &ClientStore{pool: pool}
}

func (s *ClientStore) GetByClientID(clientID string) (*models.Client, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx, `
		SELECT id, client_id, client_secret, name, description,
		       main_url, server_urls, redirect_uris, owner_id,
		       silent_sso, created_at
		FROM clients WHERE client_id = $1
	`, clientID)

	c, err := scanClient(row)
	if err != nil {
		return nil, false
	}
	return c, true
}

func (s *ClientStore) All() []*models.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx, `
		SELECT id, client_id, client_secret, name, description,
		       main_url, server_urls, redirect_uris, owner_id,
		       silent_sso, created_at
		FROM clients ORDER BY created_at ASC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]*models.Client, 0)
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (s *ClientStore) Register(c *models.Client) error {
	if c.ClientID == "" {
		return errors.New("client_id required")
	}
	if c.ID == "" {
		c.ID = "client-" + c.ClientID
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx, `
		INSERT INTO clients (
		    id, client_id, client_secret, name, description,
		    main_url, server_urls, redirect_uris, owner_id,
		    silent_sso, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (client_id) DO NOTHING
	`,
		c.ID, c.ClientID, c.ClientSecret, c.Name, c.Description,
		c.MainURL, c.ServerURLs, c.RedirectURIs, c.OwnerID,
		c.SilentSSO, c.CreatedAt,
	)
	return err
}

// UpdateSilentSSO: silent_sso 토글 (어드민 모달에서 사용).
func (s *ClientStore) UpdateSilentSSO(clientID string, silentSSO bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := s.pool.Exec(ctx, `
		UPDATE clients SET silent_sso = $1 WHERE client_id = $2
	`, silentSSO, clientID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("client not found")
	}
	return nil
}

// scanClient: pgx.Row / pgx.Rows 둘 다 받아서 Client 채움.
func scanClient(row pgx.Row) (*models.Client, error) {
	var c models.Client
	if err := row.Scan(
		&c.ID, &c.ClientID, &c.ClientSecret, &c.Name, &c.Description,
		&c.MainURL, &c.ServerURLs, &c.RedirectURIs, &c.OwnerID,
		&c.SilentSSO, &c.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &c, nil
}
