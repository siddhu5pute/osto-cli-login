package db

import "time"

type User struct {
	ID             int
	Username       string
	PasswordHash   string
	TOTPSecret     *string
	TOTPEnabled    bool
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	LastLogin      *time.Time
}

type Session struct {
	ID        string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
}
