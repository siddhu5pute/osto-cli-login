package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/siddhu5pute/osto-cli-login/internal/db"
)

var ErrExit = errors.New("exit requested")

type App struct {
	prompt         *Prompt
	output         io.Writer
	authService    AuthService
	sessionService SessionService
	userService    UserService
	totpService    TOTPService

	currentUser *db.User
	sessionID   string
	expiresAt   time.Time
}

func NewApp(
	prompt *Prompt,
	output io.Writer,
	authService AuthService,
	sessionService SessionService,
	userService UserService,
	totpService TOTPService,
) *App {
	return &App{
		prompt:         prompt,
		output:         output,
		authService:    authService,
		sessionService: sessionService,
		userService:    userService,
		totpService:    totpService,
	}
}

func (a *App) Run() error {
	fmt.Fprintln(a.output, "Welcome to Osto CLI Login System")
	fmt.Fprintln(a.output, "Type 'help' to see available commands.")

	for {
		line, err := a.prompt.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("reading command: %w", err)
		}

		if line == "" {
			continue
		}

		if err := a.executeCommand(line); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}

			fmt.Fprintf(a.output, "Error: %v\n", err)
		}
	}
}

func (a *App) executeCommand(line string) error {
	authenticated := a.isAuthenticated()

	command := findCommand(line, authenticated)
	if command == nil {
		return fmt.Errorf("unknown command: %s", line)
	}

	switch command.Name {
	case "register":
		return a.handleRegister()

	case "login":
		return a.handleLogin()

	case "whoami":
		return a.handleWhoAmI()

	case "enable-2fa":
		return fmt.Errorf("2FA setup is not available yet")

	case "disable-2fa":
		return fmt.Errorf("2FA setup is not available yet")

	case "logout":
		return a.handleLogout()

	case "help":
		a.printHelp()

	case "exit":
		fmt.Fprintln(a.output, "Goodbye!")
		return ErrExit
	}

	return nil
}

func (a *App) handleRegister() error {
	username, err := a.promptValue("Username: ")
	if err != nil {
		return err
	}

	password, err := a.prompt.ReadPassword("Password: ")
	if err != nil {
		return err
	}

	user, err := a.authService.Register(
		context.Background(),
		username,
		password,
	)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		a.output,
		"Registration successful. Welcome, %s!\n",
		user.Username,
	)

	return nil
}

func (a *App) handleLogin() error {
	if a.isAuthenticated() {
		return errors.New("already logged in")
	}

	username, err := a.promptValue("Username: ")
	if err != nil {
		return err
	}

	password, err := a.prompt.ReadPassword("Password: ")
	if err != nil {
		return err
	}

	user, err := a.authService.Login(
		context.Background(),
		username,
		password,
	)
	if err != nil {
		return err
	}

	session, err := a.sessionService.Create(
		context.Background(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("creating login session: %w", err)
	}

	a.currentUser = user
	a.sessionID = session.ID
	a.expiresAt = session.ExpiresAt

	fmt.Fprintf(
		a.output,
		"\nLogin successful. Welcome, %s!\n\n",
		user.Username,
	)

	a.printUserDetails()

	return nil
}

func (a *App) handleWhoAmI() error {
	if err := a.requireAuthentication(); err != nil {
		return err
	}

	if err := a.validateCurrentSession(); err != nil {
		return err
	}

	a.printUserDetails()

	return nil
}

func (a *App) handleLogout() error {
	if err := a.requireAuthentication(); err != nil {
		return err
	}

	err := a.sessionService.Delete(
		context.Background(),
		a.sessionID,
	)
	if err != nil {
		return fmt.Errorf("ending session: %w", err)
	}

	username := a.currentUser.Username

	a.clearSession()

	fmt.Fprintf(
		a.output,
		"Logout successful. Goodbye, %s!\n",
		username,
	)

	return nil
}

func (a *App) validateCurrentSession() error {
	if a.sessionID == "" {
		a.clearSession()
		return errors.New("session is no longer valid")
	}

	session, err := a.sessionService.Validate(
		context.Background(),
		a.sessionID,
	)
	if err != nil {
		a.clearSession()

		return fmt.Errorf("session is no longer valid: %w", err)
	}

	a.expiresAt = session.ExpiresAt

	return nil
}

func (a *App) requireAuthentication() error {
	if !a.isAuthenticated() {
		return errors.New("you must be logged in")
	}

	return nil
}

func (a *App) isAuthenticated() bool {
	return a.currentUser != nil && a.sessionID != ""
}

func (a *App) clearSession() {
	a.currentUser = nil
	a.sessionID = ""
	a.expiresAt = time.Time{}
}

func (a *App) printUserDetails() {
	if a.currentUser == nil {
		return
	}

	fmt.Fprintln(a.output, "User details:")
	fmt.Fprintf(a.output, "  Username: %s\n", a.currentUser.Username)
	fmt.Fprintf(
		a.output,
		"  Registration date: %s\n",
		a.currentUser.CreatedAt.Format(time.RFC1123),
	)

	mfaStatus := "disabled"
	if a.currentUser.TOTPEnabled {
		mfaStatus = "enabled"
	}

	fmt.Fprintf(a.output, "  MFA status: %s\n", mfaStatus)
	fmt.Fprintf(
		a.output,
		"  Session expiration: %s\n",
		a.expiresAt.Format(time.RFC1123),
	)

	if a.currentUser.LastLogin != nil {
		fmt.Fprintf(
			a.output,
			"  Last login: %s\n",
			a.currentUser.LastLogin.Format(time.RFC1123),
		)
	} else {
		fmt.Fprintln(a.output, "  Last login: never")
	}

	fmt.Fprintln(a.output)
}

func (a *App) promptValue(prompt string) (string, error) {
	fmt.Fprint(a.output, prompt)

	value, err := a.prompt.ReadLine()
	if err != nil {
		return "", err
	}

	return value, nil
}

func (a *App) printHelp() {
	if a.isAuthenticated() {
		fmt.Fprintln(a.output, "Available commands:")
		for _, command := range authenticatedCommands {
			fmt.Fprintf(
				a.output,
				"  %-11s - %s\n",
				command.Name,
				command.Description,
			)
		}
		return
	}

	fmt.Fprintln(a.output, "Available commands:")
	for _, command := range commands {
		fmt.Fprintf(
			a.output,
			"  %-11s - %s\n",
			command.Name,
			command.Description,
		)
	}
}
