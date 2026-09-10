package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) CreateSession(
	ctx context.Context,
	userID int,
	expiresAt time.Time,
) (*db.Session, error) {
	sessionID := uuid.New()

	session := &db.Session{
		ID:        sessionID.String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO sessions (
			id,
			user_id,
			created_at,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		`,
		session.ID,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	return session, nil
}

func (r *SessionRepository) GetSession(
	ctx context.Context,
	sessionID string,
) (*db.Session, error) {
	session := &db.Session{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			user_id,
			created_at,
			expires_at
		FROM sessions
		WHERE id = $1
		`,
		sessionID,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	return session, nil
}

func (r *SessionRepository) DeleteSession(
	ctx context.Context,
	sessionID string,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`
		DELETE FROM sessions
		WHERE id = $1
		`,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted session: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) DeleteExpiredSessions(
	ctx context.Context,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		DELETE FROM sessions
		WHERE expires_at <= $1
		`,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("deleting expired sessions: %w", err)
	}

	return nil
}
