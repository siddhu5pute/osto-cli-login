package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/totp"
)

var (
	ErrMFAAlreadyEnabled = errors.New("MFA is already enabled")
	ErrMFANotEnabled     = errors.New("MFA is not enabled")
)

type MFAService struct {
	users UserRepository
	totp  *totp.Service
}

func NewMFAService(
	users UserRepository,
	totpService *totp.Service,
) *MFAService {
	return &MFAService{
		users: users,
		totp:  totpService,
	}
}

func (s *MFAService) Enable(
	ctx context.Context,
	user *db.User,
) (*totp.Setup, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}

	if user.TOTPEnabled {
		return nil, ErrMFAAlreadyEnabled
	}

	setup, err := s.totp.GenerateSecret(user.Username)
	if err != nil {
		return nil, fmt.Errorf("generating MFA secret: %w", err)
	}

	if err := s.users.EnableTOTP(
		ctx,
		user.ID,
		setup.Secret,
	); err != nil {
		return nil, fmt.Errorf("enabling MFA: %w", err)
	}

	return setup, nil
}

func (s *MFAService) Disable(
	ctx context.Context,
	user *db.User,
) error {
	if user == nil {
		return errors.New("user is required")
	}

	if !user.TOTPEnabled {
		return ErrMFANotEnabled
	}

	if err := s.users.DisableTOTP(ctx, user.ID); err != nil {
		return fmt.Errorf("disabling MFA: %w", err)
	}

	return nil
}
