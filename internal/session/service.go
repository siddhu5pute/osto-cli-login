package session

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
)

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrSessionExpired = errors.New("session expired")
)

type Service struct {
	repository Repository
	timeout    time.Duration
}

func NewService(
	repository Repository,
	timeout time.Duration,
) *Service {
	return &Service{
		repository: repository,
		timeout:    timeout,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID int,
) (*db.Session, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	if s.timeout <= 0 {
		return nil, fmt.Errorf("session timeout must be positive")
	}

	expiresAt := time.Now().Add(s.timeout)

	session, err := s.repository.CreateSession(
		ctx,
		userID,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	return session, nil
}

func (s *Service) Validate(
	ctx context.Context,
	sessionID string,
) (*db.Session, error) {
	sessionID = strings.TrimSpace(sessionID)

	if sessionID == "" {
		return nil, ErrInvalidSession
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return nil, ErrInvalidSession
	}

	session, err := s.repository.GetSession(
		ctx,
		sessionID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrInvalidSession
		}

		return nil, fmt.Errorf("getting session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.repository.DeleteSession(ctx, sessionID)

		return nil, ErrSessionExpired
	}

	return session, nil
}

func (s *Service) Delete(
	ctx context.Context,
	sessionID string,
) error {
	sessionID = strings.TrimSpace(sessionID)

	if sessionID == "" {
		return ErrInvalidSession
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return ErrInvalidSession
	}

	err := s.repository.DeleteSession(
		ctx,
		sessionID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrInvalidSession
		}

		return fmt.Errorf("deleting session: %w", err)
	}

	return nil
}
