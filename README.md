[![CI](https://github.com/siddhu5pute/osto-cli-login/actions/workflows/ci.yml/badge.svg)](https://github.com/siddhu5pute/osto-cli-login/actions/workflows/ci.yml)

# Osto CLI Login System

A secure, containerized command-line authentication system built with **Go** and **PostgreSQL** — with password hashing, account lockout, session management, and optional TOTP-based two-factor authentication.

Module path: `github.com/siddhu5pute/osto-cli-login`

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Technology Stack](#technology-stack)
- [Getting Started](#getting-started)
- [CLI Usage](#cli-usage)
- [Security Design](#security-design)
- [Known Limitations](#known-limitations)
- [Database](#database)
- [Docker](#docker)
- [Testing](#testing)
- [Continuous Integration](#continuous-integration)
- [Design Decisions](#design-decisions)

## Features

- User registration and password-based authentication
- Passwords hashed with bcrypt (cost factor 12) — never stored in plaintext
- Failed-login tracking with temporary, auto-expiring account lockout
- Database-backed sessions with explicit expiration
- TOTP-based two-factor authentication (enable/disable, authenticator-app provisioning via QR-compatible URI)
- Interactive CLI with context-aware commands (separate command sets before and after login)
- PostgreSQL persistence via Docker Compose, with a health check gating app startup
- Automated unit and integration tests
- GitHub Actions CI: format check, tests against a live Postgres service container, and a `gosec` security scan

## Architecture

```
CLI Client
    │
    ▼
CLI Layer (internal/cli)
    │
    ▼
Auth & MFA Layer (internal/auth, internal/totp)
    │
    ├──► Session Layer (internal/session)
    │
    ▼
Repository Layer (internal/repository)
    │
    ▼
PostgreSQL
```

Each layer depends on an interface, not a concrete implementation — `LoginService` and `MFAService` take a `UserRepository` interface, and `session.Service` takes a `Repository` interface. This is what makes the unit tests possible without a live database.

## Project Structure

```
.
├── cmd/cli/                   # Application entry point
├── internal/
│   ├── auth/                  # Registration, login, password hashing, MFA
│   ├── cli/                   # Interactive CLI: commands, prompt loop, app wiring
│   ├── config/                # Environment-based configuration
│   ├── db/                    # Connection setup, models, migrations
│   ├── repository/            # Postgres queries for users and sessions
│   ├── session/                # Session creation, validation, expiration
│   └── totp/                   # TOTP secret generation and verification
├── .github/workflows/ci.yml   # CI pipeline
├── .gitignore                  # Excludes .env, binaries, logs, CLI history
├── Dockerfile                   # Multi-stage Go build
├── docker-compose.yml           # App + PostgreSQL
├── go.mod / go.sum
```

## Technology Stack

| Technology | Purpose |
|---|---|
| Go 1.26 | Application development |
| PostgreSQL 16 (Alpine) | Persistent storage |
| Docker / Docker Compose | Containerization and orchestration |
| `github.com/pquerna/otp` | TOTP generation and verification |
| `golang.org/x/crypto/bcrypt` | Password hashing |
| `github.com/chzyer/readline` | Interactive CLI prompt |
| `github.com/google/uuid` | Session ID generation |
| `github.com/lib/pq` | PostgreSQL driver |
| GitHub Actions | Continuous integration |
| `gosec` | Static security analysis |

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.26+ (for local development/testing)
- Git

### 1. Set the database password

The app reads its configuration from environment variables (`internal/config/config.go`) and **refuses to start** if `DB_PASSWORD` isn't set. `.gitignore` already excludes `.env` and `*.env`, so set it there locally:

```bash
echo "DB_PASSWORD=your-password-here" > .env
```

Full set of variables the app reads, with their defaults:

| Variable | Default | Notes |
|---|---|---|
| `DB_HOST` | `localhost` | Overridden to `postgres` inside Docker Compose |
| `DB_PORT` | `5432` | |
| `DB_USER` | `siddhu5pute` | |
| `DB_PASSWORD` | *(none — required)* | App exits with an error if unset |
| `DB_NAME` | `osto_auth` | |
| `SESSION_TIMEOUT_MIN` | `30` | Session lifetime, in minutes |
| `LOCKOUT_THRESHOLD` | `5` | Consecutive failed attempts before lockout |
| `LOCKOUT_DURATION_MIN` | `15` | Lockout duration, in minutes |

### 2. Start PostgreSQL

```bash
docker compose up -d postgres
docker compose ps
```

Postgres should report a `healthy` status — the `app` service is configured to wait for this before starting. Note it's exposed on host port `5433` (mapped to `5432` inside the container) to avoid clashing with a local Postgres install.

### 3. Start the application

```bash
docker compose up -d app
docker compose logs app
```

### 4. Attach to the CLI

The app is interactive, so attach your terminal to the running container:

```bash
docker compose attach app
```

You should see an `osto>` prompt.

## CLI Usage

Before login:

| Command | Description |
|---|---|
| `register` | Create a new user |
| `login` | Authenticate with username and password |
| `help` | Show available commands |
| `exit` | Quit the application |

After login:

| Command | Description |
|---|---|
| `whoami` | Show current user and session details |
| `enable-2fa` | Enable TOTP-based two-factor authentication |
| `disable-2fa` | Disable two-factor authentication |
| `logout` | End the current session |
| `help` | Show available commands |

### Register

```
osto> register
osto> testuser
Password:
Registration successful. Welcome, testuser!
```

### Login

```
osto> login
osto> testuser
Password:
Login successful. Welcome, testuser!
```

A successful login creates a database-backed session and displays account details (registration date, MFA status, session expiration, last login).

### Two-factor authentication

```
osto> enable-2fa
2FA enabled successfully.
Provisioning URI: otpauth://totp/...
```

Scan the provisioning URI with an authenticator app. From then on, login requires both the password and a 6-digit TOTP code:

```
osto> login
osto> testuser
Password:
osto> 287578
Login successful. Welcome, testuser!
```

### Error cases

```
osto> login
osto> testuser
Password: [wrong password]
Error: invalid username or password

osto> login
osto> testuser
Password: [correct password, after 5 failed attempts]
Error: account is temporarily locked

osto> login
osto> testuser
Password:
osto> 000000
Error: invalid 2FA code: invalid TOTP code
```

## Security Design

- **Password hashing** — bcrypt, cost factor 12. Empty passwords and empty hashes are explicitly rejected before comparison rather than relying on bcrypt's own error.
- **Account lockout** — failed logins increment a counter per user; once `LOCKOUT_THRESHOLD` is reached, the account is locked until `LOCKOUT_DURATION_MIN` has elapsed. On login, an expired lock is detected and cleared automatically before authentication proceeds, so users aren't permanently stuck without an admin step.
- **Sessions** — each session gets a UUID primary key, an explicit `expires_at`, and is tied to a user via a foreign key with `ON DELETE CASCADE`. Sessions are validated (not just checked for existence) before authenticated commands run.
- **TOTP (2FA)** — standard RFC 6238 TOTP: SHA1, 30-second period, 6 digits, ±1 step of clock skew tolerance. Enabling/disabling MFA are separate, explicit operations that reject redundant calls (`MFA is already enabled` / `MFA is not enabled`) rather than silently no-op-ing.
- **Secrets hygiene** — `.gitignore` excludes `.env`, `*.env`, and the CLI's own history file (`.osto_history`), so credentials and command history don't end up committed.

## Known Limitations

- No password complexity requirements are enforced at registration beyond "non-empty."
- Lockout is per-account only — there's no IP-based or global rate limiting in front of `login`.
- Sessions are validated against the database on each use; there's no in-memory or cache layer, so session checks scale with the database, not independently of it.
- The CLI is single-user per process — there's no multi-session-per-terminal handling.
- Test coverage is not uniform: `internal/config` (environment parsing and validation) and `internal/repository/sessions.go` (session queries) don't have dedicated test files, even though the rest of the codebase — including all the core auth, TOTP, session-service, and user-repository logic — does.

## Database

PostgreSQL is used for persistence, with two tables defined in `internal/db/migrations/001_init.sql`:

- **`users`** — `id`, `username` (unique), `password_hash`, `totp_secret`, `totp_enabled`, `failed_attempts`, `locked_until`, `created_at`, `last_login`.
- **`sessions`** — `id` (UUID), `user_id` (foreign key to `users`, cascading delete), `created_at`, `expires_at`, with an index on `user_id`.

## Docker

The `Dockerfile` is a two-stage build: dependencies are downloaded and the binary compiled in a `golang:1.26-alpine` build stage, then just the compiled binary is copied into a minimal `alpine:3.22` runtime image — keeping the final image free of the Go toolchain and source.

`docker-compose.yml` wires the app to a `postgres:16-alpine` service with a health check (`pg_isready`), so the app container only starts once the database is actually accepting connections — not just once the container process has started.

## Testing

Run the full test suite:

```bash
go test ./...
```

Most of the codebase has dedicated test files covering the core logic: password hashing, login and lockout behavior, registration, MFA, TOTP, session service, CLI commands, CLI app wiring, CLI prompt handling, user repository queries, and database connection setup. See [Known Limitations](#known-limitations) for the two files not yet covered.

Run CLI tests only:

```bash
go test ./internal/cli/... -v
```

## Continuous Integration

GitHub Actions (`.github/workflows/ci.yml`) runs on pushes to `main` and `addtests`, and on pull requests targeting `main`:

1. Checkout and set up Go 1.26
2. Download dependencies
3. Verify formatting: `test -z "$(gofmt -l .)"`
4. Run `go test ./... -v` against a live `postgres:16-alpine` service container
5. Install and run `gosec ./...` for static security analysis

Every step must pass for the workflow to succeed — there's no `continue-on-error` on the security scan or the format check.

## Design Decisions

- **Sessions over stateless tokens** — sessions are stored server-side (Postgres) rather than as signed JWTs, so a session can be invalidated immediately (e.g. on logout) without needing a revocation list. The trade-off is a DB lookup on every authenticated command.
- **Lockout by count + time, not IP** — lockout tracks failed attempts per account, not per source IP, which is simpler to reason about and test but means it doesn't distinguish a single attacker from many legitimate users sharing a network.
- **TOTP over SMS-based 2FA** — avoids any dependency on an SMS provider or phone number collection, at the cost of requiring the user to have an authenticator app.
- **Config fails closed** — `config.Load()` returns an error (and the app won't start) if `DB_PASSWORD` is missing, rather than defaulting to an empty or well-known password.
