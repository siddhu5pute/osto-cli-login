package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{
		db: database,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	username string,
	passwordHash string,
) (*db.User, error) {
	const query = `
		INSERT INTO users (
			username,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING
			id,
			username,
			password_hash,
			totp_secret,
			totp_enabled,
			failed_attempts,
			locked_until,
			created_at,
			last_login
	`

	user := &db.User{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		username,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.FailedAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUsernameExists
		}

		return nil, fmt.Errorf("creating user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*db.User, error) {
	const query = `
		SELECT
			id,
			username,
			password_hash,
			totp_secret,
			totp_enabled,
			failed_attempts,
			locked_until,
			created_at,
			last_login
		FROM users
		WHERE username = $1
	`

	user := &db.User{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.FailedAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.LastLogin,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("getting user by username: %w", err)
	}

	return user, nil
}
