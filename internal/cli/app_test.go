package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/totp"
)

type fakePrompt struct {
	lines     []string
	password  string
	lineIndex int
}

func (p *fakePrompt) ReadLine() (string, error) {
	if p.lineIndex >= len(p.lines) {
		return "", errors.New("no more input")
	}

	line := p.lines[p.lineIndex]
	p.lineIndex++

	return line, nil
}

func (p *fakePrompt) ReadPassword(prompt string) (string, error) {
	return p.password, nil
}

type fakeAuthService struct {
	user *db.User
	err  error
}

func (s *fakeAuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	return s.user, s.err
}

func (s *fakeAuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	return s.user, s.err
}

type fakeSessionService struct {
	createCalls int
	session     *db.Session
	err         error
}

func (s *fakeSessionService) Create(
	ctx context.Context,
	userID int,
) (*db.Session, error) {
	s.createCalls++

	if s.err != nil {
		return nil, s.err
	}

	return s.session, nil
}

func (s *fakeSessionService) Validate(
	ctx context.Context,
	sessionID string,
) (*db.Session, error) {
	return s.session, nil
}

func (s *fakeSessionService) Delete(
	ctx context.Context,
	sessionID string,
) error {
	return nil
}

type fakeUserService struct{}

func (s *fakeUserService) EnableTOTP(
	ctx context.Context,
	userID int,
	secret string,
) error {
	return nil
}

func (s *fakeUserService) DisableTOTP(
	ctx context.Context,
	userID int,
) error {
	return nil
}

type fakeTOTPService struct {
	verifyCalls int
	verifyErr   error
}

func (s *fakeTOTPService) GenerateSecret(
	username string,
) (*totp.Setup, error) {
	return &totp.Setup{
		Secret:       "test-secret",
		ProvisionURI: "otpauth://test",
	}, nil
}

func (s *fakeTOTPService) VerifyCode(
	secret string,
	code string,
) error {
	s.verifyCalls++

	return s.verifyErr
}

func newTestApp(
	prompt PromptService,
	authService AuthService,
	sessionService SessionService,
	totpService TOTPService,
) *App {
	return NewApp(
		prompt,
		&bytes.Buffer{},
		authService,
		sessionService,
		&fakeUserService{},
		totpService,
	)
}

func testUser(totpEnabled bool) *db.User {
	secret := "test-secret"

	user := &db.User{
		ID:          1,
		Username:    "testuser",
		TOTPEnabled: totpEnabled,
		CreatedAt:   time.Now(),
	}

	if totpEnabled {
		user.TOTPSecret = &secret
	}

	return user
}

func TestHandleLoginWithout2FA(t *testing.T) {
	prompt := &fakePrompt{
		lines:    []string{"testuser"},
		password: "password",
	}

	authService := &fakeAuthService{
		user: testUser(false),
	}

	sessionService := &fakeSessionService{
		session: &db.Session{
			ID:        "session-123",
			UserID:    1,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	totpService := &fakeTOTPService{}

	app := newTestApp(
		prompt,
		authService,
		sessionService,
		totpService,
	)

	err := app.handleLogin()
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	if app.currentUser == nil {
		t.Fatal("expected current user to be set")
	}

	if app.sessionID != "session-123" {
		t.Fatalf("expected session ID %q, got %q", "session-123", app.sessionID)
	}

	if totpService.verifyCalls != 0 {
		t.Fatalf("expected TOTP verification not to be called, got %d calls", totpService.verifyCalls)
	}

	if sessionService.createCalls != 1 {
		t.Fatalf("expected session creation once, got %d calls", sessionService.createCalls)
	}
}

func TestHandleLoginWith2FASucceeds(t *testing.T) {
	prompt := &fakePrompt{
		lines:    []string{"testuser", "123456"},
		password: "password",
	}

	authService := &fakeAuthService{
		user: testUser(true),
	}

	sessionService := &fakeSessionService{
		session: &db.Session{
			ID:        "session-123",
			UserID:    1,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	totpService := &fakeTOTPService{}

	app := newTestApp(
		prompt,
		authService,
		sessionService,
		totpService,
	)

	err := app.handleLogin()
	if err != nil {
		t.Fatalf("expected 2FA login to succeed, got error: %v", err)
	}

	if totpService.verifyCalls != 1 {
		t.Fatalf("expected TOTP verification once, got %d calls", totpService.verifyCalls)
	}

	if sessionService.createCalls != 1 {
		t.Fatalf("expected session creation once, got %d calls", sessionService.createCalls)
	}

	if app.currentUser == nil {
		t.Fatal("expected current user to be set")
	}

	if app.sessionID != "session-123" {
		t.Fatalf("expected session ID %q, got %q", "session-123", app.sessionID)
	}
}

func TestHandleLoginWith2FARejectsInvalidCode(t *testing.T) {
	prompt := &fakePrompt{
		lines:    []string{"testuser", "000000"},
		password: "password",
	}

	authService := &fakeAuthService{
		user: testUser(true),
	}

	sessionService := &fakeSessionService{
		session: &db.Session{
			ID:        "session-123",
			UserID:    1,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	totpService := &fakeTOTPService{
		verifyErr: errors.New("invalid TOTP code"),
	}

	app := newTestApp(
		prompt,
		authService,
		sessionService,
		totpService,
	)

	err := app.handleLogin()
	if err == nil {
		t.Fatal("expected login to fail with invalid 2FA code")
	}

	if totpService.verifyCalls != 1 {
		t.Fatalf("expected TOTP verification once, got %d calls", totpService.verifyCalls)
	}

	if sessionService.createCalls != 0 {
		t.Fatalf(
			"expected session not to be created after invalid 2FA code, got %d calls",
			sessionService.createCalls,
		)
	}

	if app.currentUser != nil {
		t.Fatal("expected current user to remain nil")
	}

	if app.sessionID != "" {
		t.Fatalf("expected session ID to remain empty, got %q", app.sessionID)
	}
}
