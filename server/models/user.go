package models

import "time"

// User: 글로벌 user pool 의 1 row.
// Phase-R 부터 DB users 테이블에 영속. ID 는 UUID.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// seedPasswordHash: 학습용 시드 사용자 공통 비밀번호("password123") 의 bcrypt(cost=12) hash.
// 실제 운영에서는 각 사용자마다 고유 salt 가 적용된 hash 가 저장된다.
const seedPasswordHash = "$2a$12$dC674FROOYvOF.OPHSWntuvU2QjhGHtIUe2LvUFJepN.DmSXQibvq"

// TestUsers: Phase 1 학습용 다중 사용자.
// 모두 비밀번호 "password123".
// Phase-R R-3 부터 /oauth/login 은 이 map 대신 DB 의 users 테이블을 조회한다.
// 본 map 은 Postgres 시드용으로 사용된 후 R-7 에서 제거 예정.
var TestUsers = map[string]User{
	"alice": {Username: "alice", PasswordHash: seedPasswordHash, DisplayName: "Alice"},
	"bob":   {Username: "bob", PasswordHash: seedPasswordHash, DisplayName: "Bob"},
	"carol": {Username: "carol", PasswordHash: seedPasswordHash, DisplayName: "Carol"},
}

// DummyPasswordHash: timing attack 완화용.
// 존재하지 않는 ID 로 로그인 시도 시에도 동일 cost 의 bcrypt 검증을 수행해
// 응답 시간으로 사용자 enumeration 이 불가능하도록 만든다.
const DummyPasswordHash = seedPasswordHash
