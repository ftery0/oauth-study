package db

import (
	"context"
	"fmt"
)

// schemaSQL: 부팅 시 CREATE TABLE IF NOT EXISTS / ALTER TABLE ADD COLUMN IF NOT EXISTS
// 로 멱등하게 적용한다.
//
// Phase 변화:
//   - Phase 2-C : groups + clients 생성
//   - Phase-R   : users 신규 + clients.silent_sso 컬럼 추가
//                 (그룹 관련 컬럼/테이블 DROP 은 R-7 cleanup 단계에서)
//
// 본격 운영 단계에서는 goose / migrate 같은 마이그레이션 도구를 도입할 가치.
// 학습 단계에서는 inline + idempotent 로 충분.
const schemaSQL = `
-- Phase 2-C: 그룹 (Phase-R 의 R-7 에서 제거 예정 — 일단 유지)
CREATE TABLE IF NOT EXISTS groups (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    sso_default  TEXT NOT NULL CHECK (sso_default IN ('ON', 'OFF')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Phase 2-C: 클라이언트
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

-- Phase-R R-1: clients 에 silent_sso 컬럼 추가 (기본 true = 참여).
-- 글로벌 user pool 모델로 가면서 그룹 정책 대신 client 단위 토글로.
ALTER TABLE clients ADD COLUMN IF NOT EXISTS silent_sso BOOLEAN NOT NULL DEFAULT true;

-- Phase-R R-1: users 글로벌 테이블.
-- 코드 하드코딩 TestUsers 를 대체. 이메일/실명은 학습 범위 외라 username + hash 만.
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
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
