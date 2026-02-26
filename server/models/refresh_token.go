package models

import "time"

type RefreshToken struct {
	Token     string
	UserID    string
	ClientID  string
	Scope     string
	ExpiresAt time.Time
}
