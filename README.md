# aloh-SSH

> A real-time messaging server built over the SSH protocol.

**aloh-SSH** is a real-time messaging and presence server built entirely on top of the SSH protocol. It uses SSH for both authentication and transport, providing secure, encrypted communication with a built-in friend system, presence notifications, and user management — all over standard SSH connections.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Layers](#layers)
- [Features](#features)
  - [Authentication & Registration](#authentication--registration)
  - [User Profiles](#user-profiles)
  - [Friendship System](#friendship-system)
  - [Presence & Real-Time Events](#presence--real-time-events)
  - [Blocking System](#blocking-system)
  - [Session Management](#session-management)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Running with Docker Compose](#running-with-docker-compose)
  - [Quick Start](#quick-start)
  - [Docker Compose Services](#docker-compose-services)
  - [Docker Compose Commands](#docker-compose-commands)
- [Running Locally](#running-locally)
- [Database Migrations](#database-migrations)
  - [Migration Files](#migration-files)
  - [Creating a New Migration](#creating-a-new-migration)
- [Project Structure](#project-structure)
- [Domain Model](#domain-model)
  - [Database Schema](#database-schema)
  - [Key Relationships](#key-relationships)
- [SSH Protocol](#ssh-protocol)
  - [Client Versions](#client-versions)
  - [Authentication Flow](#authentication-flow)
  - [SSH Request Handlers](#ssh-request-handlers)
  - [Event Channel](#event-channel)
- [Events](#events)
  - [Event Types](#event-types)
  - [Event Data Types](#event-data-types)
- [Error Handling](#error-handling)
  - [Error Types](#error-types-1)
  - [SSH Error Responses](#ssh-error-responses)
- [Logging](#logging)
- [Environment Variables](#environment-variables)
- [CI/CD](#cicd)
- [License](#license)

---

## Overview

aloh-SSH is a Go-based server that implements a real-time messaging and presence system on top of the SSH protocol. Users connect via SSH (using public key authentication or password authentication), register with a unique nickname, add friends, block users, customize personal information (tagline, color, nickname), and receive real-time notifications when their friends come online, go offline, or update their profiles.

The server leverages the [Charmbracelet SSH](https://github.com/charmbracelet/ssh) library for SSH server functionality and [pgx](https://github.com/jackc/pgx) for high-performance PostgreSQL connection pooling.

---

## Architecture

The project follows a clean, layered architectural pattern:

```text
┌─────────────────────────────────────────────────────────┐
│                    cmd/app/main.go                      │
│                     (Entry Point)                       │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────┴──────────────────────────────┐
│                   internal/app/app.go                   │
│            (Application Bootstrap / Wiring)             │
└────┬─────────┬──────────────┬─────────────┬─────────────┘
     │         │              │             │
     ▼         ▼              ▼             ▼
┌────────┐ ┌─────────┐ ┌──────────────┐ ┌───────────┐
│ config │ │pkg/     │ │internal/     │ │internal/  │
│        │ │logger   │ │domain        │ │server     │
│        │ │(slog)   │ │services      │ │(SSH)      │
└────────┘ └─────────┘ └──────┬───────┘ └─────┬─────┘
                              │               │
                              ▼               │
                       ┌──────────────┐       │
                       │internal/     │       │
                       │domain repos  │       │
                       └──────┬───────┘       │
                              │               │
                              ▼               │
                ┌───────────────────────────┐ │
                │    PostgreSQL Database    │◄┘
                │(users, friends, blocked..)│
                └───────────────────────────┘
```

### Layers

| Layer | Package | Description |
| :--- | :--- | :--- |
| **Entry Point** | `cmd/app/main.go` | Application entry point invoking `app.Run()` |
| **App Bootstrap** | `internal/app/app.go` | Initializes config, logger, storage, repositories, services, and the SSH server |
| **Config** | `internal/config/config.go` | YAML configuration parsing via Viper with environment variable expansion |
| **Server** | `internal/server/` | SSH server with authentication, request routing, and channel handlers |
| **Domain Services** | `internal/domain/services/` | Business logic for users, sessions, friendships, and blocked accounts |
| **Domain Repositories** | `internal/domain/repositories/` | Data access layer (PostgreSQL + in-memory session store) |
| **Domain Models** | `internal/domain/models/` | Core domain entities: `User`, `Session`, `Event`, `Friend`, `Identity` |
| **Storage** | `pkg/storage/` | PostgreSQL connection pool wrapper using `pgx` |
| **Logger** | `pkg/logger/` | Structured logging using `slog` (supporting `local`, `dev`, and `prod` outputs) |
| **Errors** | `pkg/errs/` | Custom error types with operation context wrapping |
| **Utils** | `internal/utils/` | Shared utilities (SSH key fingerprinting, etc.) |
| **Public API** | `event.go` | Public type aliases and event constants re-exported at the module root |

---

## Features

### Authentication & Registration

- **Public Key Registration**: New users register by connecting with SSH client version `SSH-2.0-aloh-register`, supplying an SSH public key.
- **Password Authentication**: Existing users authenticate with their nickname and password using client version `SSH-2.0-aloh-login`.
- **Password Management**: Users can set an initial password after key registration and change it at any time by providing both the old and new passwords.

### User Profiles

- **Nickname**: Unique identifier (max 80 chars). Users can rename their accounts.
- **Tagline**: Short personal status message (max 28 chars) visible to friends.
- **Color**: Hex code color preference (7 chars, e.g., `#FF5733`) for client UI customization.
- **Public Key**: SSH public key stored alongside its SHA256 fingerprint for fast lookups and verification.

### Friendship System

- **Send Requests**: Send friend requests by nickname. The target receives a `NEW_FRIEND_REQ` event.
- **Accept/Deny**: Recipients can accept or decline pending incoming requests.
- **Remove Friends**: Either party can remove an active friendship at any time.
- **Pending Requests Delivery**: Pending requests are preserved and delivered upon user login.

### Presence & Real-Time Events

- **Online Presence**: When a user connects, friends receive `FRIEND_ONLINE` notifications with connection metadata.
- **Offline Notifications**: When a user disconnects, friends immediately receive `FRIEND_OFFLINE` events.
- **Connection Tracking**: Tracks active SSH channel connections per user and broadcasts `FRIEND_CONNECTIONS` updates.
- **Live Profile Sync**: Profile updates (tagline, color, nickname) are broadcast to all connected friends instantly.

### Blocking System

- **Block User**: Users can block others by nickname. Blocking removes friendships and broadcasts the event to the target's event channel.
- **Unblock User**: Users can unblock previously restricted users.
- **Cross-Reference Sync**: Nickname modifications notify both accepted friends and blocked accounts to prevent impersonation.

### Session Management

- **In-Memory Sessions**: Tracked in-memory with concurrent `sync.Map` stores managing active connections, channels, and pipelines.
- **Graceful Shutdown**: On client disconnect, friends are notified, sessions are drained, and contexts are cancelled.
- **Keep-Alive**: Automatic keep-alive ticker packets are exchanged to prevent timeouts over NAT/routers.

---

## Prerequisites

- **Go 1.26+** (for local source builds)
- **Docker** and **Docker Compose** (recommended for deployment)
- **PostgreSQL** (can run via the bundled Compose service)
- **goose** (for running migrations locally)

---

## Configuration

Configuration is located at `configs/config.yaml` and supports environment variable expansion (`${VAR}` syntax).

```yaml
app:
  name: "aloh-ssh-server"
  env: "local"           # local | dev | prod | test
  version: "1.0.0"
  logPath: ""            # Optional: file path for log output

server:
  host: "0.0.0.0"
  port: "${SERVER_PORT}"
  timeout: 10s
  idleTimeout: 60s
  keepAliveTimeout: 10s

storage:
  user: "${POSTGRES_USER}"
  password: "${POSTGRES_PASSWORD}"
  database: "${POSTGRES_DB}"
  timezone: "${TZ}"
  host: "postgres"
  port: "5432"
  sslMode: "${POSTGRES_SSL_MODE}"
  connectTimeout: 10s
  pingTimeout: 5s
  amountOfConns: 5        # Maximum connection pool size
```

> **Note:** The configuration path is specified via the `CONFIG_PATH` environment variable.

---

## Running with Docker Compose

This is the recommended deployment method for both production and local development.

### Quick Start

```bash
# Clone the repository
git clone https://github.com/kiryuhakipyatok/aloh-ssh.git
cd aloh-ssh

# Setup environment variables
cp .env .env.local  # or edit .env directly

# Run the full stack (app + postgres + migrations)
make docker-run-app
```

### Docker Compose Services

| Service | Description |
| :--- | :--- |
| `alohssh` | The primary aloh-SSH application server |
| `postgres` | PostgreSQL database with built-in health checks |
| `migrate` | Database migration runner using `goose` |

### Docker Compose Commands

```bash
# Start the app server
docker compose up --build alohssh

# Run database migrations
make docker-migrate-up

# Rollback database migrations
make docker-migrate-down

# Create a new migration file
make create-migra name=add_some_column
```

---

## Running Locally

### Prerequisites

Ensure you have a PostgreSQL instance running and the required environment variables set.

### Steps

```bash
# 1. Export required environment variables
export POSTGRES_USER=alohssh
export POSTGRES_PASSWORD=root
export POSTGRES_DB=alohssh
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_SSL_MODE=disable
export TZ=Europe/Minsk
export SERVER_PORT=2222
export CONFIG_PATH=./configs

# 2. Run database migrations
goose -dir=./internal/migrations postgres "postgresql://alohssh:root@localhost:5432/alohssh?sslmode=disable" up

# 3. Build and execute
go build -o ./bin/app ./cmd/app/main.go
./bin/app
```

---

## Database Migrations

Database migrations are managed with [goose](https://github.com/pressly/goose) and stored under `internal/migrations/`.

### Migration Files

| File | Description |
| :--- | :--- |
| `20260330125510_create_users.sql` | Creates `users` table with UUID PK, nickname, key, fingerprint, password, and timestamp |
| `20260516190442_new_index_nickname.sql` | Adds index on `nickname` for fast lookups |
| `20260520190353_create_friends.sql` | Creates `friends` table with `friendship_status` enum (`pending` / `active`) |
| `20260523130134_create_unique_friendship_idx.sql` | Adds unique index preventing duplicate bidirectional friendships |
| `20260524181928_create_blocked.sql` | Creates `blocked_users` table |
| `20260706171123_add_users_tagline.sql` | Adds `tagline` column to `users` table |
| `20260720123553_add_users_color.sql` | Adds `color` column to `users` table (defaults to `#random`) |

### Creating a New Migration

```bash
make create-migra name=add_new_column
```

---

## Project Structure

```text
aloh-ssh/
├── cmd/
│   └── app/
│       └── main.go                    # Application entry point
├── configs/
│   └── config.yaml                    # YAML configuration
├── internal/
│   ├── app/
│   │   └── app.go                     # Application bootstrap
│   ├── config/
│   │   └── config.go                  # Configuration loading
│   ├── domain/
│   │   ├── models/
│   │   │   ├── event.go               # Event types and constructors
│   │   │   ├── session.go             # Session model
│   │   │   └── user.go                # User, Identity, Friend models
│   │   ├── repositories/
│   │   │   ├── blocked-repo.go        # Blocked users repository
│   │   │   ├── friendship-repo.go     # Friendships repository
│   │   │   ├── session-repo.go        # In-memory session repository
│   │   │   └── user-repo.go           # Users repository
│   │   └── services/
│   │       ├── blocked-serv.go        # Blocked users service
│   │       ├── friendship-serv.go     # Friendships service
│   │       ├── session-serv.go        # Sessions service
│   │       └── user-serv.go           # Users service
│   ├── migrations/                    # Database migrations (goose format)
│   ├── server/
│   │   ├── channel.go                 # SSH event channel handler
│   │   ├── errs.go                    # Error casting utilities
│   │   ├── requests.go                # SSH request handlers
│   │   └── server.go                  # SSH server setup
│   └── utils/
│       └── utils.go                   # Utilities (fingerprint generation)
├── pkg/
│   ├── errs/
│   │   └── errs.go                    # Custom error types
│   ├── logger/
│   │   └── logger.go                  # Structured logging (slog)
│   └── storage/
│       ├── errs.go                    # Storage error helpers
│       └── postgres.go                # PostgreSQL connection pool
├── .dockerignore
├── .env                               # Environment variables
├── .gitlab-ci.yml                     # CI/CD pipeline
├── Dockerfile
├── docker-compose.yaml
├── event.go                           # Public type aliases and constants
├── go.mod
├── go.sum
└── Makefile
```

---

## Domain Model

### Database Schema

```text
┌─────────────────────────────────────────────────┐
│                     users                       │
├─────────────────────────────────────────────────┤
│ id            UUID PK  DEFAULT gen_random_uuid()│
│ nickname      VARCHAR(80)  UNIQUE  NOT NULL     │
│ key           VARCHAR(256) UNIQUE  NOT NULL     │
│ fingerprint   VARCHAR(256) UNIQUE  NOT NULL     │
│ register_time TIMESTAMPTZ DEFAULT CURRENT_TIME │
│ password      BYTEA  UNIQUE                     │
│ tagline       VARCHAR(28) DEFAULT ''            │
│ color         VARCHAR(7)  DEFAULT '#random'     │
└─────────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│                  friends                   │
├────────────────────────────────────────────┤
│ user_id1    UUID  FK → users.id            │
│ user_id2    UUID  FK → users.id            │
│ req_time    TIMESTAMPTZ DEFAULT NOW()      │
│ status      friendship_status              │
│             ('pending' | 'active')         │
│ PRIMARY KEY (user_id1, user_id2)           │
│ CHECK (user_id1 <> user_id2)               │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│               blocked_users                │
├────────────────────────────────────────────┤
│ blocker_id  UUID  FK → users.id            │
│ blocked_id  UUID  FK → users.id            │
│ block_time  TIMESTAMPTZ DEFAULT NOW()      │
│ PRIMARY KEY (blocker_id, blocked_id)       │
│ CHECK (blocker_id <> blocked_id)           │
└────────────────────────────────────────────┘
```

### Key Relationships

- **Users $\leftrightarrow$ Friends**: Many-to-many self-referencing relationship via `friends`.
- **Users $\leftrightarrow$ Blocked Users**: Many-to-many self-referencing relationship via `blocked_users`.
- **Friendship Status**: Requests originate as `pending`; acceptance transitions the row to `active`.
- **Bidirectional Uniqueness**: The `unique_friendship_idx` guarantees that no two users can create duplicate relations in either directional order.

---

## SSH Protocol

aloh-SSH leverages SSH client version strings to determine authentication and operation modes:

### Client Versions

| Version String | Mode | Description |
| :--- | :--- | :--- |
| `SSH-2.0-aloh-login` | **Login** | Password-based authentication for existing accounts |
| `SSH-2.0-aloh-register` | **Register** | Public key-based registration for new users |
| `SSH-2.0-aloh-default` | **Default** | Public key-based login for existing accounts |

### Authentication Flow

1. **New User Registration:**
   - Client connects with version `SSH-2.0-aloh-register`.
   - The server triggers public-key hooks to insert a new user profile.
   - The user then sets an initial password via the `set-password` SSH request.
2. **Existing User Login (Key-Based):**
   - Client connects with version `SSH-2.0-aloh-default`.
   - The server verifies that the presented public key matches the record on file.
3. **Existing User Login (Password-Based):**
   - Client connects with version `SSH-2.0-aloh-login`.
   - The server verifies the user's password using the stored bcrypt hash.

### SSH Request Handlers

| Request Type | Handler | Description |
| :--- | :--- | :--- |
| `set-password` | `setPasswordRequest()` | Set the initial password for an authenticated user |
| `key` | `setNewKeyRequest()` | Update the user's SSH public key |
| `new-friend` | `newFriendRequest()` | Send a friend request by nickname |
| `accept-friend` | `acceptFriendshipRequest()` | Accept a pending friend request |
| `deny-friend` | `denyFriendshipRequest()` | Decline a pending friend request |
| `personal-data` | `fetchPersonalRequest()` | Fetch user profile, friends, and block lists |
| `delete-friend` | `deleteFromFriendsRequest()` | Remove an existing friend |
| `block-user` | `blockUserRequest()` | Block another user by nickname |
| `unblock-user` | `unblockUserRequest()` | Unblock a user by nickname |
| `conns-update` | `updateCurOnlineRequest()` | Update the count of current active sessions |
| `set-tagline` | `setTaglineRequest()` | Update user personal tagline |
| `set-color` | `setColorRequest()` | Update user display color |
| `new-nickname` | `newNicknameRequest()` | Rename account (requires password verification) |
| `new-password` | `newPasswordRequest()` | Change password (requires old password) |

### Event Channel

The server exposes an `event-channel` SSH channel handler that:
1. Spawns a session for the authenticated user.
2. Broadcasts online presence events to connected friends.
3. Emits current connection stats.
4. Maintains an active keep-alive heartbeat.
5. Proxies domain events from Go channels straight into the SSH stream.
6. Notifies friends of offline status and cleans up resources upon disconnect.

---

## Events

Events are encoded as JSON payloads and piped over the SSH event channel. Each message contains a `type` (`uint`) and `data` (`json.RawMessage`).

### Event Types

| Constant | Value | Description |
| :--- | :---: | :--- |
| `NEW_FRIEND_REQ` | `0` | A new friend request was received |
| `ACCEPT_FRIEND` | `1` | A friend request was accepted |
| `DENY_FRIEND` | `2` | A friend request was denied |
| `DELETE_FRIEND` | `3` | A friend was removed |
| `BLOCK_USER` | `4` | The current user was blocked by someone |
| `UNBLOCK_USER` | `5` | A user was unblocked |
| `FRIEND_ONLINE` | `6` | A friend came online (includes connection data) |
| `FRIEND_OFFLINE` | `7` | A friend went offline |
| `FRIEND_CONNECTIONS` | `8` | Friend active connection count update |
| `UPDATE_HARD_DENOISE` | `9` | Hard denoise preference updated |
| `UPDATE_SOFT_DENOISE` | `10` | Soft denoise preference updated |
| `UPDATE_TAGLINE` | `11` | A friend updated their tagline |
| `UPDATE_NICKNAME` | `12` | A friend updated their nickname |
| `UPDATE_COLOR` | `13` | A friend updated their profile color |

### Event Data Types

| Type | Fields |
| :--- | :--- |
| `Identity` | `id` (UUID), `nickname` (string) |
| `FriendConnsData` | `identity` (Identity), `connects` ([]Identity) |
| `TaglineData` | `identity` (Identity), `tagline` (string) |
| `ColorData` | `identity` (Identity), `color` (string) |
| `NicknameData` | `identity` (Identity), `nickname` (string) |

---

## Error Handling

### Error Types

Custom errors defined under `pkg/errs/errs.go`:

| Error Type | Description |
| :--- | :--- |
| `ErrNotFoundBase` | Target entity does not exist |
| `ErrAlreadyExistsBase` | Resource conflict or duplicate record |
| `ErrRequestTimeoutBase` | Context deadline exceeded during operation |
| `ErrInvalidTypeBase` | Type assertion or serialization failure |

All errors attach an operation (`op`) string identifier and can be matched using standard `errors.Is()` and `errors.As()`.

### SSH Error Responses

Domain errors are converted to single-byte status codes over SSH responses:

| Code | Constant | Meaning |
| :---: | :--- | :--- |
| `0` | `SUCCESS` | Operation succeeded |
| `1` | `NOT_FOUND` | Resource not found |
| `2` | `ALREADY_EXISTS` | Resource already exists |
| `3` | `SERVER_ERROR` | Internal server error |

---

## Logging

Structured logging is powered by Go's standard `log/slog` with environment presets:

| Environment | Handler | Log Level |
| :--- | :--- | :--- |
| `local` | `TextHandler` | `Debug` |
| `dev` | `JSONHandler` | `Debug` |
| `prod` | `JSONHandler` | `Info` |

Log records include: `type="app"`, `env`, `app`, `version`, and `op` (operation call chain). Logs can be written concurrently to `stdout` and an optional file via `app.logPath`.

---

## Environment Variables

Configured in your `.env` file:

### PostgreSQL

| Variable | Default | Description |
| :--- | :--- | :--- |
| `POSTGRES_USER` | `alohssh` | Database username |
| `POSTGRES_PASSWORD` | `root` | Database password |
| `POSTGRES_DB` | `alohssh` | Database name |
| `POSTGRES_HOST` | `postgres` | Database host |
| `POSTGRES_PORT` | `5432` | Database port |
| `POSTGRES_SSL_MODE` | `disable` | SSL mode for database connection |

### Server

| Variable | Default | Description |
| :--- | :--- | :--- |
| `SERVER_PORT` | `2222` | SSH listening port |

### General

| Variable | Default | Description |
| :--- | :--- | :--- |
| `TZ` | `Europe/Minsk` | Server timezone |
| `CONFIG_PATH` | `../../configs` | Path to directory containing `config.yaml` |
| `GOOSE_DRIVER` | `postgres` | Migration driver for goose |
| `GOOSE_DBSTRING` | *(derived)* | Connection string for goose |
| `MIGRATIONS_PATH` | `./internal/migrations` | Path to SQL migration files |

---

## CI/CD

The GitLab CI pipeline (`.gitlab-ci.yml`) executes through four stages:

```text
┌─────────┐     ┌───────────┐     ┌──────────┐     ┌──────────────┐
│  Build  │ ──► │  Migrate  │ ──► │  Deploy  │ ──► │ Health Check │
└─────────┘     └───────────┘     └──────────┘     └──────────────┘
```

1. **Build**: Builds production Docker images.
2. **Migrate**: Spins up temporary PostgreSQL services and applies migrations.
3. **Deploy**: Gracefully halts previous instances and updates containers.
4. **Health Check**: Validates server responsiveness and SSH handshake availability.

---

## License

This project is licensed under the [MIT License](LICENSE).