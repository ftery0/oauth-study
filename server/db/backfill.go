package db

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BackfillClientSecretHashes: client_secret_hash 가 비어있고 client_secret 평문이 남아있는 행에 대해
// bcrypt 해시를 생성해 채워넣고, 평문 컬럼은 NULL 로 비운다.
// 부팅 시 1 회만 의미 있게 동작 (이후 호출은 no-op).
func BackfillClientSecretHashes(ctx context.Context) error {
	if Pool == nil {
		return fmt.Errorf("db pool not initialized")
	}

	rows, err := Pool.Query(ctx, `
		SELECT id, client_secret
		FROM clients
		WHERE client_secret_hash IS NULL AND client_secret IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("scan plaintext clients: %w", err)
	}

	type row struct {
		id, secret string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.secret); err != nil {
			rows.Close()
			return fmt.Errorf("scan row: %w", err)
		}
		pending = append(pending, r)
	}
	rows.Close()

	for _, r := range pending {
		hash, err := bcrypt.GenerateFromPassword([]byte(r.secret), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash %s: %w", r.id, err)
		}
		if _, err := Pool.Exec(ctx, `
			UPDATE clients SET client_secret_hash = $1, client_secret = NULL WHERE id = $2
		`, string(hash), r.id); err != nil {
			return fmt.Errorf("update %s: %w", r.id, err)
		}
	}
	return nil
}
