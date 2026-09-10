package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (r *UserRepository) IncrementFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	const query = `
		UPDATE users
		SET failed_attempts = failed_attempts + 1
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("incrementing failed attempts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking failed attempts update: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) LockUser(
	ctx context.Context,
	userID int,
	lockedUntil time.Time,
) error {
	const query = `
		UPDATE users
		SET locked_until = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		lockedUntil,
		userID,
	)
	if err != nil {
		return fmt.Errorf("locking user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking lock update: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ResetFailedAttempts(
	ctx context.Context,
	userID int,
) error {
	const query = `
		UPDATE users
		SET failed_attempts = 0,
		    locked_until = NULL
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("resetting failed attempts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking failed attempts reset: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(
	ctx context.Context,
	userID int,
	loginTime time.Time,
) error {
	const query = `
		UPDATE users
		SET last_login = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		loginTime,
		userID,
	)
	if err != nil {
		return fmt.Errorf("updating last login: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking last login update: %w", err)
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}
