package models

import "time"

type AuthCode struct {
	Code        string
	ClientID    string
	UserID      string
	RedirectURI string
	Scope       string
	ExpiresAt   time.Time
}
