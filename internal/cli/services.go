package cli

import (
	"context"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/totp"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*db.User, error)
	Login(ctx context.Context, username, password string) (*db.User, error)
}

type SessionService interface {
	Create(ctx context.Context, userID int) (*db.Session, error)
	Validate(ctx context.Context, sessionID string) (*db.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type UserService interface {
	EnableTOTP(ctx context.Context, userID int, secret string) error
	DisableTOTP(ctx context.Context, userID int) error
}

type TOTPService interface {
	GenerateSecret(username string) (*totp.Setup, error)
	VerifyCode(secret, code string) error
}
