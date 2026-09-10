package totp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrInvalidSecret = errors.New("invalid TOTP secret")
	ErrInvalidCode   = errors.New("invalid TOTP code")
)

type Service struct {
	issuer string
}

func NewService(issuer string) *Service {
	return &Service{
		issuer: issuer,
	}
}

type Setup struct {
	Secret       string
	ProvisionURI string
}

func (s *Service) GenerateSecret(username string) (*Setup, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return nil, errors.New("username is required")
	}

	key, err := totp.Generate(
		totp.GenerateOpts{
			Issuer:      s.issuer,
			AccountName: username,
			Algorithm:   otp.AlgorithmSHA1,
			Period:      30,
			SecretSize:  20,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("generating TOTP secret: %w", err)
	}

	return &Setup{
		Secret:       key.Secret(),
		ProvisionURI: key.URL(),
	}, nil
}

func (s *Service) VerifyCode(secret, code string) error {
	secret = strings.TrimSpace(secret)
	code = strings.TrimSpace(code)

	if secret == "" {
		return ErrInvalidSecret
	}

	if code == "" {
		return ErrInvalidCode
	}

	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
	if err != nil {
		return fmt.Errorf("validating TOTP code: %w", err)
	}

	if !valid {
		return ErrInvalidCode
	}

	return nil
}
