package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
)

type fakeRepository struct {
	session *db.Session

	createCalls int
	getCalls    int
	deleteCalls int

	getErr    error
	createErr error
	deleteErr error
}

func (f *fakeRepository) CreateSession(
	ctx context.Context,
	userID int,
	expiresAt time.Time,
) (*db.Session, error) {
	f.createCalls++

	if f.createErr != nil {
		return nil, f.createErr
	}

	session := &db.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	f.session = session

	return session, nil
}

func (f *fakeRepository) GetSession(
	ctx context.Context,
	sessionID string,
) (*db.Session, error) {
	f.getCalls++

	if f.getErr != nil {
		return nil, f.getErr
	}

	if f.session == nil {
		return nil, repository.ErrSessionNotFound
	}

	return f.session, nil
}

func (f *fakeRepository) DeleteSession(
	ctx context.Context,
	sessionID string,
) error {
	f.deleteCalls++

	if f.deleteErr != nil {
		return f.deleteErr
	}

	f.session = nil

	return nil
}

func (f *fakeRepository) DeleteExpiredSessions(
	ctx context.Context,
) error {
	return nil
}

func TestCreateSession(t *testing.T) {
	repo := &fakeRepository{}

	service := NewService(
		repo,
		30*time.Minute,
	)

	before := time.Now()

	session, err := service.Create(
		context.Background(),
		1,
	)

	if err != nil {
		t.Fatalf("expected session creation to succeed: %v", err)
	}

	after := time.Now()

	if session.UserID != 1 {
		t.Fatalf(
			"expected user ID 1, got %d",
			session.UserID,
		)
	}

	if session.ID == "" {
		t.Fatal("expected session ID")
	}

	if session.ExpiresAt.Before(
		before.Add(30 * time.Minute),
	) {
		t.Fatal("session expires too early")
	}

	if session.ExpiresAt.After(
		after.Add(30 * time.Minute),
	) {
		t.Fatal("session expires too late")
	}

	if repo.createCalls != 1 {
		t.Fatalf(
			"expected one create call, got %d",
			repo.createCalls,
		)
	}
}

func TestCreateSessionRejectsInvalidUser(t *testing.T) {
	service := NewService(
		&fakeRepository{},
		30*time.Minute,
	)

	_, err := service.Create(
		context.Background(),
		0,
	)

	if err == nil {
		t.Fatal("expected invalid user error")
	}
}

func TestValidateSession(t *testing.T) {
	sessionID := uuid.New().String()

	repo := &fakeRepository{
		session: &db.Session{
			ID:        sessionID,
			UserID:    1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		},
	}

	service := NewService(
		repo,
		30*time.Minute,
	)

	session, err := service.Validate(
		context.Background(),
		sessionID,
	)

	if err != nil {
		t.Fatalf(
			"expected valid session, got %v",
			err,
		)
	}

	if session.UserID != 1 {
		t.Fatalf(
			"expected user ID 1, got %d",
			session.UserID,
		)
	}
}

func TestValidateSessionRejectsUnknownSession(t *testing.T) {
	repo := &fakeRepository{}

	service := NewService(
		repo,
		30*time.Minute,
	)

	_, err := service.Validate(
		context.Background(),
		uuid.New().String(),
	)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf(
			"expected ErrInvalidSession, got %v",
			err,
		)
	}
}

func TestValidateSessionRejectsExpiredSession(t *testing.T) {
	sessionID := uuid.New().String()

	repo := &fakeRepository{
		session: &db.Session{
			ID:        sessionID,
			UserID:    1,
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		},
	}

	service := NewService(
		repo,
		30*time.Minute,
	)

	_, err := service.Validate(
		context.Background(),
		sessionID,
	)

	if !errors.Is(err, ErrSessionExpired) {
		t.Fatalf(
			"expected ErrSessionExpired, got %v",
			err,
		)
	}

	if repo.deleteCalls != 1 {
		t.Fatalf(
			"expected expired session to be deleted, got %d delete calls",
			repo.deleteCalls,
		)
	}
}

func TestValidateSessionRejectsMalformedSessionID(t *testing.T) {
	service := NewService(
		&fakeRepository{},
		30*time.Minute,
	)

	_, err := service.Validate(
		context.Background(),
		"not-a-valid-uuid",
	)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf(
			"expected ErrInvalidSession, got %v",
			err,
		)
	}
}

func TestDeleteSession(t *testing.T) {
	sessionID := uuid.New().String()

	repo := &fakeRepository{
		session: &db.Session{
			ID:        sessionID,
			UserID:    1,
			ExpiresAt: time.Now().Add(30 * time.Minute),
		},
	}

	service := NewService(
		repo,
		30*time.Minute,
	)

	err := service.Delete(
		context.Background(),
		sessionID,
	)

	if err != nil {
		t.Fatalf(
			"expected session deletion to succeed: %v",
			err,
		)
	}

	if repo.deleteCalls != 1 {
		t.Fatalf(
			"expected one delete call, got %d",
			repo.deleteCalls,
		)
	}
}

func TestDeleteSessionRejectsMalformedID(t *testing.T) {
	service := NewService(
		&fakeRepository{},
		30*time.Minute,
	)

	err := service.Delete(
		context.Background(),
		"invalid-session",
	)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf(
			"expected ErrInvalidSession, got %v",
			err,
		)
	}
}
