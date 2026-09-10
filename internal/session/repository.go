package session

import (
	"context"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

type Repository interface {
	CreateSession(
		ctx context.Context,
		userID int,
		expiresAt time.Time,
	) (*db.Session, error)

	GetSession(
		ctx context.Context,
		sessionID string,
	) (*db.Session, error)

	DeleteSession(
		ctx context.Context,
		sessionID string,
	) error

	DeleteExpiredSessions(
		ctx context.Context,
	) error
}
