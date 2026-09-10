package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/siddhu5pute/osto-cli-login/internal/auth"
	"github.com/siddhu5pute/osto-cli-login/internal/cli"
	"github.com/siddhu5pute/osto-cli-login/internal/config"
	"github.com/siddhu5pute/osto-cli-login/internal/db"
	"github.com/siddhu5pute/osto-cli-login/internal/repository"
	"github.com/siddhu5pute/osto-cli-login/internal/session"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	conn, err := db.Connect(cfg.DSN())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(conn); err != nil {
		return fmt.Errorf("running database migrations: %w", err)
	}

	userRepository := repository.NewUserRepository(conn)
	sessionRepository := repository.NewSessionRepository(conn)

	registrationService := auth.NewRegistrationService(userRepository)

	loginService := auth.NewLoginService(
		userRepository,
		cfg.LockoutThreshold,
		cfg.LockoutDuration,
	)

	sessionService := session.NewService(
		sessionRepository,
		cfg.SessionTimeout,
	)

	prompt, err := cli.NewPrompt()
	if err != nil {
		return fmt.Errorf("initializing CLI: %w", err)
	}
	defer prompt.Close()

	app := cli.NewApp(
		prompt,
		os.Stdout,
		cliAuthService{
			registration: registrationService,
			login:        loginService,
		},
		sessionService,
	)

	return app.Run()
}

type cliAuthService struct {
	registration *auth.RegistrationService
	login        *auth.LoginService
}

func (s cliAuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	return s.registration.Register(ctx, username, password)
}

func (s cliAuthService) Login(
	ctx context.Context,
	username string,
	password string,
) (*db.User, error) {
	return s.login.Login(ctx, username, password)
}
