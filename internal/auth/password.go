package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcryptCost,
	)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}

	return string(hash), nil
}

func VerifyPassword(password, hash string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if hash == "" {
		return fmt.Errorf("password hash cannot be empty")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	); err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}
