package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Connect(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening db connection: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return conn, nil
}

func RunMigrations(conn *sql.DB) error {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}
	for _, entry := range entries {
		content, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
		}
		if _, err := conn.Exec(string(content)); err != nil {
			return fmt.Errorf("applying migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
