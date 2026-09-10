package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

func TestUserRepository_CreateAndGetUser(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping PostgreSQL integration test")
	}

	database, err := db.Connect(dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repo := NewUserRepository(database)

	username := fmt.Sprintf("test_%s", uuid.NewString())
	passwordHash := "bcrypt-test-hash"

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	user, err := repo.CreateUser(
		ctx,
		username,
		passwordHash,
	)
	if err != nil {
		t.Fatalf("CreateUser() returned error: %v", err)
	}

	if user.Username != username {
		t.Fatalf(
			"expected username %q, got %q",
			username,
			user.Username,
		)
	}

	if user.PasswordHash != passwordHash {
		t.Fatal("stored password hash does not match supplied hash")
	}

	found, err := repo.GetUserByUsername(ctx, username)
	if err != nil {
		t.Fatalf("GetUserByUsername() returned error: %v", err)
	}

	if found.ID != user.ID {
		t.Fatalf(
			"expected user ID %d, got %d",
			user.ID,
			found.ID,
		)
	}

	if found.Username != username {
		t.Fatalf(
			"expected username %q, got %q",
			username,
			found.Username,
		)
	}
}

func TestUserRepository_GetUserByUsername_NotFound(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping PostgreSQL integration test")
	}

	database, err := db.Connect(dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer database.Close()

	repo := NewUserRepository(database)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	_, err = repo.GetUserByUsername(
		ctx,
		"definitely-does-not-exist",
	)

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf(
			"expected ErrUserNotFound, got %v",
			err,
		)
	}
}
