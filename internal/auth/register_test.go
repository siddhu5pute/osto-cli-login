package auth

import (
	"context"
	"testing"
)

func TestRegistrationRejectsEmptyUsername(t *testing.T) {
	service := NewRegistrationService(nil)

	_, err := service.Register(
		context.Background(),
		"",
		"StrongPassword123!",
	)

	if err != ErrInvalidUsername {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestRegistrationRejectsShortUsername(t *testing.T) {
	service := NewRegistrationService(nil)

	_, err := service.Register(
		context.Background(),
		"ab",
		"StrongPassword123!",
	)

	if err != ErrInvalidUsername {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestRegistrationRejectsLongUsername(t *testing.T) {
	service := NewRegistrationService(nil)

	username := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	_, err := service.Register(
		context.Background(),
		username,
		"StrongPassword123!",
	)

	if err != ErrInvalidUsername {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestRegistrationRejectsEmptyPassword(t *testing.T) {
	service := NewRegistrationService(nil)

	_, err := service.Register(
		context.Background(),
		"testuser",
		"",
	)

	if err == nil {
		t.Fatal("expected error for empty password")
	}
}
