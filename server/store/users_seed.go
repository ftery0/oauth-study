package store

import (
	"context"
	"errors"

	"github.com/ftery0/ouath/server/models"
)

// SeedUsers: 학습용 시드 사용자 (alice/bob/carol) 를 UserStore 에 멱등 등록.
// 이미 존재하면 무시 (ErrUserAlreadyExists 흡수).
// main 부팅 시 Postgres 사용 시 호출.
func SeedUsers(ctx context.Context, s UserStore) error {
	for _, u := range models.TestUsers {
		// 시드는 username 을 user_id 로도 쓰지 않고 DB 가 UUID 할당하도록 ID 비움.
		cp := u // copy — map iteration variable 의 주소를 안 잡도록
		if err := s.Create(ctx, &cp); err != nil {
			if errors.Is(err, ErrUserAlreadyExists) {
				continue
			}
			return err
		}
	}
	return nil
}
