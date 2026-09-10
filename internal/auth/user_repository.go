package auth

import (
	"context"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*db.User, error)

	IncrementFailedAttempts(ctx context.Context, userID int) error

	LockUser(ctx context.Context, userID int, lockedUntil time.Time) error

	ResetFailedAttempts(ctx context.Context, userID int) error

	UpdateLastLogin(ctx context.Context, userID int, loginTime time.Time) error

	EnableTOTP(ctx context.Context, userID int, secret string) error

	DisableTOTP(ctx context.Context, userID int) error
}
