package db

import (
	"context"
	"fmt"
)

// schemaSQL: Phase 2-C 의 시작 스키마.
// 부팅 시 CREATE TABLE IF NOT EXISTS 로 idempotent 하게 적용한다.
//
// 본격 운영 단계에서는 goose / migrate 같은 마이그레이션 도구를 도입할 가치.
// 학습 단계에서는 inline + idempotent 로 충분.
const schemaSQL = `
CREATE TABLE IF NOT EXISTS groups (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    sso_default  TEXT NOT NULL CHECK (sso_default IN ('ON', 'OFF')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clients (
    id              TEXT PRIMARY KEY,
    client_id       TEXT NOT NULL UNIQUE,
    client_secret   TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    main_url        TEXT NOT NULL DEFAULT '',
    server_urls     TEXT[] NOT NULL DEFAULT '{}',
    redirect_uris   TEXT[] NOT NULL DEFAULT '{}',
    owner_id        TEXT NOT NULL DEFAULT '',
    group_id        TEXT REFERENCES groups(id) ON DELETE SET NULL,
    sso_override    TEXT NOT NULL DEFAULT 'INHERIT'
                    CHECK (sso_override IN ('INHERIT', 'FORCE_ON', 'FORCE_OFF')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clients_group_id ON clients(group_id);
`

// RunMigrations: schema 를 멱등하게 적용. main 부팅 시 db.Connect 후 호출.
func RunMigrations(ctx context.Context) error {
	if Pool == nil {
		return fmt.Errorf("db pool not initialized")
	}
	if _, err := Pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
