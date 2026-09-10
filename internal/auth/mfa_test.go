package auth

import (
	"context"
	"testing"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/totp"
)

type fakeMFARepository struct {
	enableTOTPCalled  bool
	disableTOTPCalled bool

	secret string
	userID int

	enableErr  error
	disableErr error
}

func (f *fakeMFARepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*db.User, error) {
	return nil, nil
}

func (f *fakeMFARepository) IncrementFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	return nil
}

func (f *fakeMFARepository) LockUser(
	ctx context.Context,
	userID int,
	lockedUntil time.Time,
) error {
	return nil
}

func (f *fakeMFARepository) ResetFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	return nil
}

func (f *fakeMFARepository) UpdateLastLogin(
	ctx context.Context,
	userID int,
	loginTime time.Time,
) error {
	return nil
}

func (f *fakeMFARepository) EnableTOTP(
	ctx context.Context,
	userID int,
	secret string,
) error {
	f.enableTOTPCalled = true
	f.userID = userID
	f.secret = secret
	return f.enableErr
}

func (f *fakeMFARepository) DisableTOTP(
	ctx context.Context,
	userID int,
) error {
	f.disableTOTPCalled = true
	f.userID = userID
	return f.disableErr
}

func TestMFAServiceEnable(t *testing.T) {
	repo := &fakeMFARepository{}
	totpService := totp.NewService("Osto")

	service := NewMFAService(repo, totpService)

	user := &db.User{
		ID:       1,
		Username: "testuser",
	}

	setup, err := service.Enable(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if setup == nil {
		t.Fatal("expected TOTP setup")
	}

	if setup.Secret == "" {
		t.Fatal("expected TOTP secret")
	}

	if setup.ProvisionURI == "" {
		t.Fatal("expected provisioning URI")
	}

	if !repo.enableTOTPCalled {
		t.Fatal("expected EnableTOTP to be called")
	}

	if repo.userID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, repo.userID)
	}

	if repo.secret != setup.Secret {
		t.Fatal("expected repository to receive generated secret")
	}
}

func TestMFAServiceEnableRejectsAlreadyEnabled(t *testing.T) {
	repo := &fakeMFARepository{}
	service := NewMFAService(repo, totp.NewService("Osto"))

	user := &db.User{
		ID:          1,
		Username:    "testuser",
		TOTPEnabled: true,
	}

	_, err := service.Enable(context.Background(), user)

	if err != ErrMFAAlreadyEnabled {
		t.Fatalf("expected ErrMFAAlreadyEnabled, got %v", err)
	}

	if repo.enableTOTPCalled {
		t.Fatal("expected EnableTOTP not to be called")
	}
}

func TestMFAServiceDisable(t *testing.T) {
	repo := &fakeMFARepository{}
	service := NewMFAService(repo, totp.NewService("Osto"))

	user := &db.User{
		ID:          1,
		Username:    "testuser",
		TOTPEnabled: true,
	}

	err := service.Disable(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repo.disableTOTPCalled {
		t.Fatal("expected DisableTOTP to be called")
	}

	if repo.userID != user.ID {
		t.Fatalf("expected user ID %d, got %d", user.ID, repo.userID)
	}
}

func TestMFAServiceDisableRejectsNotEnabled(t *testing.T) {
	repo := &fakeMFARepository{}
	service := NewMFAService(repo, totp.NewService("Osto"))

	user := &db.User{
		ID:          1,
		Username:    "testuser",
		TOTPEnabled: false,
	}

	err := service.Disable(context.Background(), user)

	if err != ErrMFANotEnabled {
		t.Fatalf("expected ErrMFANotEnabled, got %v", err)
	}

	if repo.disableTOTPCalled {
		t.Fatal("expected DisableTOTP not to be called")
	}
}
