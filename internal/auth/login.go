package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account is temporarily locked")
)

type LoginService struct {
	users            UserRepository
	lockoutThreshold int
	lockoutDuration  time.Duration
}

func NewLoginService(
	users UserRepository,
	lockoutThreshold int,
	lockoutDuration time.Duration,
) *LoginService {
	return &LoginService{
		users:            users,
		lockoutThreshold: lockoutThreshold,
		lockoutDuration:  lockoutDuration,
	}
}

func (s *LoginService) Login(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	username = strings.TrimSpace(username)

	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("looking up user: %w", err)
	}

	now := time.Now()

	if user.LockedUntil != nil {
		if now.Before(*user.LockedUntil) {
			return nil, ErrAccountLocked
		}

		// The previous lock has expired.
		// Clear the lock state before attempting authentication.
		if err := s.users.ResetFailedAttempts(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("resetting expired lock: %w", err)
		}

		user.FailedAttempts = 0
		user.LockedUntil = nil
	}

	if err := VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, s.handleFailedLogin(ctx, user)
	}

	if err := s.users.ResetFailedAttempts(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("resetting failed attempts: %w", err)
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, fmt.Errorf("updating last login: %w", err)
	}

	user.FailedAttempts = 0
	user.LockedUntil = nil
	user.LastLogin = &now

	return user, nil
}

func (s *LoginService) handleFailedLogin(
	ctx context.Context,
	user *db.User,
) error {
	newFailedAttempts := user.FailedAttempts + 1

	if err := s.users.IncrementFailedAttempts(
		ctx,
		user.ID,
	); err != nil {
		return fmt.Errorf("recording failed login: %w", err)
	}

	if s.lockoutThreshold > 0 &&
		newFailedAttempts >= s.lockoutThreshold {

		lockedUntil := time.Now().Add(s.lockoutDuration)

		if err := s.users.LockUser(
			ctx,
			user.ID,
			lockedUntil,
		); err != nil {
			return fmt.Errorf("locking account: %w", err)
		}
	}

	return ErrInvalidCredentials
}
