package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
)

type fakeUserRepository struct {
	user *db.User

	incrementFailedAttemptsCalls int
	lockUserCalls                int
	resetFailedAttemptsCalls     int
	updateLastLoginCalls         int

	lastLockedUntil time.Time
	lastLoginTime   time.Time

	getUserErr error
}

func (f *fakeUserRepository) EnableTOTP(
	ctx context.Context,
	userID int,
	secret string,
) error {
	return nil
}

func (f *fakeUserRepository) DisableTOTP(
	ctx context.Context,
	userID int,
) error {
	return nil
}

func (f *fakeUserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*db.User, error) {
	if f.getUserErr != nil {
		return nil, f.getUserErr
	}

	if f.user == nil {
		return nil, repository.ErrUserNotFound
	}

	return f.user, nil
}

func (f *fakeUserRepository) IncrementFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	f.incrementFailedAttemptsCalls++
	f.user.FailedAttempts++

	return nil
}

func (f *fakeUserRepository) LockUser(
	ctx context.Context,
	userID int,
	lockedUntil time.Time,
) error {
	f.lockUserCalls++
	f.lastLockedUntil = lockedUntil
	f.user.LockedUntil = &lockedUntil

	return nil
}

func (f *fakeUserRepository) ResetFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	f.resetFailedAttemptsCalls++
	f.user.FailedAttempts = 0
	f.user.LockedUntil = nil

	return nil
}

func (f *fakeUserRepository) UpdateLastLogin(
	ctx context.Context,
	userID int,
	loginTime time.Time,
) error {
	f.updateLastLoginCalls++
	f.lastLoginTime = loginTime
	f.user.LastLogin = &loginTime

	return nil
}

func newTestUser(t *testing.T, password string) *db.User {
	t.Helper()

	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	return &db.User{
		ID:             1,
		Username:       "testuser",
		PasswordHash:   passwordHash,
		FailedAttempts: 0,
		TOTPEnabled:    false,
	}
}

func TestLoginSuccess(t *testing.T) {
	user := newTestUser(t, "StrongPassword123!")

	repo := &fakeUserRepository{
		user: user,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	ctx := context.Background()

	result, err := service.Login(
		ctx,
		"testuser",
		"StrongPassword123!",
	)

	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}

	if result.Username != "testuser" {
		t.Fatalf(
			"expected username testuser, got %q",
			result.Username,
		)
	}

	if repo.resetFailedAttemptsCalls != 1 {
		t.Fatalf(
			"expected reset failed attempts to be called once, got %d",
			repo.resetFailedAttemptsCalls,
		)
	}

	if repo.updateLastLoginCalls != 1 {
		t.Fatalf(
			"expected last login to be updated once, got %d",
			repo.updateLastLoginCalls,
		)
	}

	if result.LastLogin == nil {
		t.Fatal("expected LastLogin to be set")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	user := newTestUser(t, "StrongPassword123!")

	repo := &fakeUserRepository{
		user: user,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"testuser",
		"WrongPassword!",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if repo.incrementFailedAttemptsCalls != 1 {
		t.Fatalf(
			"expected one failed attempt, got %d",
			repo.incrementFailedAttemptsCalls,
		)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	repo := &fakeUserRepository{}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"unknown",
		"SomePassword123!",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginLocksAccountAfterThreshold(t *testing.T) {
	user := newTestUser(t, "StrongPassword123!")
	user.FailedAttempts = 4

	repo := &fakeUserRepository{
		user: user,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"testuser",
		"WrongPassword!",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if repo.incrementFailedAttemptsCalls != 1 {
		t.Fatalf(
			"expected increment to be called once, got %d",
			repo.incrementFailedAttemptsCalls,
		)
	}

	if repo.lockUserCalls != 1 {
		t.Fatalf(
			"expected lock to be called once, got %d",
			repo.lockUserCalls,
		)
	}

	if repo.lastLockedUntil.IsZero() {
		t.Fatal("expected account to have a lock expiration")
	}

	if !repo.lastLockedUntil.After(time.Now()) {
		t.Fatal("expected lock expiration to be in the future")
	}
}

func TestLoginRejectsLockedAccount(t *testing.T) {
	user := newTestUser(t, "StrongPassword123!")

	lockedUntil := time.Now().Add(10 * time.Minute)
	user.LockedUntil = &lockedUntil

	repo := &fakeUserRepository{
		user: user,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"testuser",
		"StrongPassword123!",
	)

	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf(
			"expected ErrAccountLocked, got %v",
			err,
		)
	}

	if repo.incrementFailedAttemptsCalls != 0 {
		t.Fatal("locked account should not increment failed attempts")
	}

	if repo.updateLastLoginCalls != 0 {
		t.Fatal("locked account should not update last login")
	}
}

func TestLoginAllowsExpiredLock(t *testing.T) {
	user := newTestUser(t, "StrongPassword123!")

	lockedUntil := time.Now().Add(-1 * time.Minute)
	user.LockedUntil = &lockedUntil
	user.FailedAttempts = 5

	repo := &fakeUserRepository{
		user: user,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	result, err := service.Login(
		context.Background(),
		"testuser",
		"StrongPassword123!",
	)

	if err != nil {
		t.Fatalf(
			"expected login after expired lock, got %v",
			err,
		)
	}

	if result.LockedUntil != nil {
		t.Fatal("expected expired lock to be cleared")
	}

	if result.FailedAttempts != 0 {
		t.Fatalf(
			"expected failed attempts to be reset, got %d",
			result.FailedAttempts,
		)
	}
}

func TestLoginRejectsEmptyUsername(t *testing.T) {
	service := NewLoginService(
		&fakeUserRepository{},
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"",
		"StrongPassword123!",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginRejectsEmptyPassword(t *testing.T) {
	service := NewLoginService(
		&fakeUserRepository{},
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"testuser",
		"",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}
}

func TestLoginReturnsRepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &fakeUserRepository{
		getUserErr: expectedErr,
	}

	service := NewLoginService(
		repo,
		5,
		15*time.Minute,
	)

	_, err := service.Login(
		context.Background(),
		"testuser",
		"StrongPassword123!",
	)

	if err == nil {
		t.Fatal("expected repository error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected repository error to be wrapped, got %v",
			err,
		)
	}
}
