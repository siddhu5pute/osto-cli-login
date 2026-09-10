package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	SessionTimeout   time.Duration
	LockoutThreshold int
	LockoutDuration  time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "siddhu5pute"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "osto_auth"),
	}

	timeoutMin, err := strconv.Atoi(getEnv("SESSION_TIMEOUT_MIN", "30"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TIMEOUT_MIN: %w", err)
	}
	cfg.SessionTimeout = time.Duration(timeoutMin) * time.Minute

	threshold, err := strconv.Atoi(getEnv("LOCKOUT_THRESHOLD", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid LOCKOUT_THRESHOLD: %w", err)
	}
	cfg.LockoutThreshold = threshold

	lockoutMin, err := strconv.Atoi(getEnv("LOCKOUT_DURATION_MIN", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid LOCKOUT_DURATION_MIN: %w", err)
	}
	cfg.LockoutDuration = time.Duration(lockoutMin) * time.Minute

	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD must be set")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
