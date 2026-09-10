package totp

import (
	"errors"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func TestGenerateSecret(t *testing.T) {
	service := NewService("Osto CLI Login")

	setup, err := service.GenerateSecret("alice")
	if err != nil {
		t.Fatalf("expected secret generation to succeed: %v", err)
	}

	if setup.Secret == "" {
		t.Fatal("expected TOTP secret")
	}

	if setup.ProvisionURI == "" {
		t.Fatal("expected provisioning URI")
	}
}

func TestGenerateSecretRejectsEmptyUsername(t *testing.T) {
	service := NewService("Osto CLI Login")

	_, err := service.GenerateSecret("")

	if err == nil {
		t.Fatal("expected username validation error")
	}
}

func TestVerifyCode(t *testing.T) {
	service := NewService("Osto CLI Login")

	setup, err := service.GenerateSecret("alice")
	if err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	now := time.Now()

	code, err := totp.GenerateCodeCustom(
		setup.Secret,
		now,
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	if err != nil {
		t.Fatalf("failed to generate test code: %v", err)
	}

	// The generated code is based on the current time, so
	// verification should succeed immediately.
	if err := service.VerifyCode(setup.Secret, code); err != nil {
		t.Fatalf("expected valid code, got %v", err)
	}
}

func TestVerifyCodeRejectsWrongCode(t *testing.T) {
	service := NewService("Osto CLI Login")

	setup, err := service.GenerateSecret("alice")
	if err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	err = service.VerifyCode(setup.Secret, "000000")

	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}

func TestVerifyCodeRejectsEmptySecret(t *testing.T) {
	service := NewService("Osto CLI Login")

	err := service.VerifyCode("", "123456")

	if !errors.Is(err, ErrInvalidSecret) {
		t.Fatalf("expected ErrInvalidSecret, got %v", err)
	}
}

func TestVerifyCodeRejectsEmptyCode(t *testing.T) {
	service := NewService("Osto CLI Login")

	err := service.VerifyCode("secret", "")

	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}
