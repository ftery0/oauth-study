package models

type User struct {
	ID           string
	PasswordHash string
	DisplayName  string
}

// seedPasswordHash: 학습용 시드 사용자 공통 비밀번호("password123") 의 bcrypt(cost=12) hash.
// 실제 운영에서는 각 사용자마다 고유 salt 가 적용된 hash 가 저장된다.
const seedPasswordHash = "$2a$12$dC674FROOYvOF.OPHSWntuvU2QjhGHtIUe2LvUFJepN.DmSXQibvq"

// TestUsers: Phase 1 학습용 다중 사용자.
// 모두 비밀번호 "password123".
var TestUsers = map[string]User{
	"alice": {ID: "alice", PasswordHash: seedPasswordHash, DisplayName: "Alice"},
	"bob":   {ID: "bob", PasswordHash: seedPasswordHash, DisplayName: "Bob"},
	"carol": {ID: "carol", PasswordHash: seedPasswordHash, DisplayName: "Carol"},
}

// DummyPasswordHash: timing attack 완화용.
// 존재하지 않는 ID 로 로그인 시도 시에도 동일 cost 의 bcrypt 검증을 수행해
// 응답 시간으로 사용자 enumeration 이 불가능하도록 만든다.
const DummyPasswordHash = seedPasswordHash
