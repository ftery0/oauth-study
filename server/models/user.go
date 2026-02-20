package models

type User struct {
	ID           string
	PasswordHash string
}

// Phase 1 테스트용 유저 ("password123" 의 bcrypt 해시)
var TestUser = User{
	ID:           "alice",
	PasswordHash: "$2a$12$u4s6Z3MZnGE0k6FEDRYKiuSJYWOsVRQdajwv4LWQrmLxkoY7FHvBq",
}
