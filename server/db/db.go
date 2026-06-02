// Package db 는 PostgreSQL 연결 풀을 관리한다.
//
// Phase 2-B 단계 — 연결만 한다 (실제 store 가 Postgres 를 사용하는 것은 P2-C).
// 다음 phase 가 store 인터페이스를 분리하고 Postgres 구현체를 끼우면 그때 본격 사용.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool: 패키지 전역 연결 풀. main 에서 Connect 호출 후 사용.
var Pool *pgxpool.Pool

// Connect: connection string 으로 풀 생성 + ping.
// 호출자(main) 는 실패 시 종료/경고 정책을 결정.
func Connect(ctx context.Context, dsn string) error {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	// 학습 단계: 합리적인 풀 크기. 운영에서는 부하 측정 후 조정.
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return fmt.Errorf("ping postgres: %w", err)
	}

	Pool = pool
	return nil
}

// Close: 풀 종료 (main 종료 시 또는 graceful shutdown 시 호출).
func Close() {
	if Pool != nil {
		Pool.Close()
		Pool = nil
	}
}
