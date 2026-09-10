package cli

import (
	"context"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
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
