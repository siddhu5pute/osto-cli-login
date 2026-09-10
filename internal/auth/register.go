package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
)

var ErrInvalidUsername = errors.New("invalid username")

type RegistrationService struct {
	users *repository.UserRepository
}

func NewRegistrationService(users *repository.UserRepository) *RegistrationService {
	return &RegistrationService{
		users: users,
	}
}

func (s *RegistrationService) Register(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return nil, ErrInvalidUsername
	}

	if len(username) < 3 || len(username) > 50 {
		return nil, ErrInvalidUsername
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("preparing password: %w", err)
	}

	user, err := s.users.CreateUser(
		ctx,
		username,
		passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("registering user: %w", err)
	}

	return user, nil
}
