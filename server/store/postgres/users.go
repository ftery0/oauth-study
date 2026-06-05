package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// UserStore: pgxpool 기반 글로벌 user pool 영속 구현체.
type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, display_name, COALESCE(email, ''), email_verified, created_at, updated_at
		FROM users WHERE username = $1
	`, username)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrUserNotFound
	}
	return u, err
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, username, password_hash, display_name, COALESCE(email, ''), email_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrUserNotFound
	}
	return u, err
}

// Create: INSERT users. username UNIQUE 충돌 시 ErrUserAlreadyExists.
// 호출자가 u.ID 채워 보내면 그것 사용, 빈 값이면 DB default(uuid) 적용.
func (s *UserStore) Create(ctx context.Context, u *models.User) error {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}

	var (
		row pgx.Row
		err error
	)
	emailArg := any(nil)
	if u.Email != "" {
		emailArg = u.Email
	}
	if u.ID != "" {
		row = s.pool.QueryRow(ctx, `
			INSERT INTO users (id, username, password_hash, display_name, email, email_verified, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id::text
		`, u.ID, u.Username, u.PasswordHash, u.DisplayName, emailArg, u.EmailVerified, u.CreatedAt, u.UpdatedAt)
	} else {
		row = s.pool.QueryRow(ctx, `
			INSERT INTO users (username, password_hash, display_name, email, email_verified, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id::text
		`, u.Username, u.PasswordHash, u.DisplayName, emailArg, u.EmailVerified, u.CreatedAt, u.UpdatedAt)
	}

	if err = row.Scan(&u.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return store.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	if err := row.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.Email, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &u, nil
}
