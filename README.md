# 💊 Medication Tracker API

> A backend API server for a medication tracking system, inspired by Apple Health's medication feature. Built with Go, Gin, and PostgreSQL — designed to help users manage their prescriptions, schedules, and adherence logs in a structured, privacy-first way.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Running with Docker](#running-with-docker)
  - [Running Locally](#running-locally)
- [API Reference](#api-reference)
  - [Health Check](#health-check)
  - [Medications](#medications)
- [Project Structure](#project-structure)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Medication Tracker is a RESTful backend API that replicates the core logic behind Apple Health's medication tracking module. The system allows a user to:

- **Add medications** with clinical details (name, form, strength, prescription number)
- **Visualize medications** with shapes and colors to aid recognition
- **Create flexible schedules** — daily, interval-based, day-of-week, or as-needed
- **Log adherence** by marking each scheduled dose as *taken* or *skipped*
- **Track drug interactions** with severity warnings and acknowledgement support

The API is stateless, timezone-aware, and designed to back a native mobile or web client.

---

## Features

### ✅ Implemented

| Feature | Description |
|---|---|
| 💊 **Medication Management** | Full CRUD for medications with clinical metadata |
| 🎨 **Medication Visuals** | CRUD endpoints for `medication_visuals` (shape, colors) via nested models |
| 🗓 **Schedules API** | Full CRUD for `schedules`, including nested `schedule_days` and `schedule_times` |
| 📋 **Adherence Logs** | Record and retrieve taken/skipped dose history (`medication_logs`) |
| 🔍 **Paginated Listing** | Sortable, paginated medication lists |
| 🩺 **Swagger Docs** | Interactive API documentation at `/swagger/index.html` |
| 🐳 **Docker Ready** | Full Docker Compose setup with hot reload via Air |
| 🔒 **Soft Deletes** | Medications use `deleted_at` — history is preserved |

### 🚧 Planned (TODO)

The following features are **not yet implemented** and will be added incrementally. Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

| # | Feature | Description |
|---|---|---|
| 1 | ⚠️ **Drug Interactions API** | Store, query, and acknowledge interaction warnings (`drug_interactions`) |
| 2 | 🔔 **Push Notifications** | Dose reminder alerts via **APNs** (iOS) and **FCM** (Android/Web) |
| 3 | 🧠 **Smart Notification Scheduling** | Generate notification trigger times from schedules, respecting user timezone |
| 4 | 👤 **User Auth** | Registration, login, JWT issuance, and token refresh |
| 5 | 🌏 **Timezone Management** | User-facing endpoint to update timezone preference |
| 6 | ⚡ **Redis Caching** | Cache frequently read data (medication lists, schedules) using the provisioned Redis service |
| 7 | 📊 **Adherence Stats** | Aggregated adherence rates and streaks for charts/calendars |

> **Next up:** Item **1** (Drug Interactions API) followed by **2–3** (notification alerts). See the [open issues](https://github.com/jaygaha/medication-tracker-api/issues) for progress.

---

## Architecture

The project follows a clean **layered architecture** pattern, separating concerns across distinct packages:

```
Request → Router → Middleware → Handler → Service → Repository → PostgreSQL
```

| Layer | Package | Responsibility |
|---|---|---|
| **Entry Point** | `cmd/server` | App bootstrap, DI wiring, server start |
| **Config** | `internal/config` | Environment loading, DB init, migrations, seeding |
| **Routes** | `internal/routes` | Route definitions and middleware attachment |
| **Middleware** | `internal/middleware` | Request enrichment (e.g., user context injection) |
| **Handler** | `internal/handler` | HTTP request parsing, response formatting |
| **Service** | `internal/service` | Business logic and validation |
| **Repository** | `internal/repository` | Database queries (raw SQL via `database/sql`) |
| **Models** | `internal/models` | Domain structs, request/response types |
| **Errors** | `internal/errors` | Typed application errors (NotFound, Validation) |

### Dependency Injection

Dependencies flow top-down through explicit constructor injection — there is no global state or service locator. The wiring happens in `main.go`:

```go
medRepo    := repository.NewMedicationRepository(db)
medService := service.NewMedicationService(medRepo)
medHandler := handler.NewMedicationHandler(medService)
router     := routes.SetupRouter(medHandler)
```

---

## Database Schema

The schema is inspired by Apple Health's medication model and is designed for clinical accuracy and extensibility. All tables use UUID primary keys and `TIMESTAMP WITH TIME ZONE` columns.

```
┌─────────────────────────────────────────────────────────────┐
│                           users                             │
│  id · first_name · last_name · email · timezone · ...       │
└─────────────────────────┬───────────────────────────────────┘
                          │ 1
                          │
          ┌───────────────┼──────────────────────┐
          │               │                      │
          ▼ N             ▼ N                    ▼ N
┌─────────────────┐ ┌───────────────┐  ┌──────────────────────┐
│   medications   │ │ medication_   │  │   drug_interactions  │
│                 │ │    logs       │  │                      │
│ name · form     │ │               │  │ med1 · med2          │
│ strength · rx   │ │ status · dose │  │ severity · ack       │
│ status · notes  │ │ timestamps    │  └──────────────────────┘
└────────┬────────┘ └───────────────┘
         │ 1
         ├──────────────────────┐
         │                      │
         ▼ 1                    ▼ 1
┌─────────────────┐    ┌──────────────────┐
│ medication_     │    │    schedules     │
│   visuals       │    │                  │
│                 │    │ type · interval  │
│ shape · colors  │    │ start · end      │
└─────────────────┘    └────────┬─────────┘
                                │ 1
                    ┌───────────┴──────────┐
                    │                      │
                    ▼ N                    ▼ N
           ┌──────────────┐    ┌──────────────────┐
           │ schedule_    │    │  schedule_times  │
           │    days      │    │                  │
           │              │    │ time_of_day      │
           │ day_of_week  │    │ dose_amount      │
           └──────────────┘    └──────────────────┘
```

### Enums

| Type | Values |
|---|---|
| `medication_form` | `tablet`, `capsule`, `liquid`, `topical`, `injection`, `drops`, `inhaler`, `powder`, `device`, `other` |
| `medication_status` | `active`, `archived`, `discontinued` |
| `frequency_type` | `every_day`, `regular_intervals`, `specific_days`, `as_needed` |
| `log_status` | `taken`, `skipped` |
| `interaction_severity` | `minor`, `moderate`, `severe`, `critical` |

---

## Tech Stack

| Component | Technology |
|---|---|
| **Language** | [Go 1.26+](https://golang.org/) |
| **Web Framework** | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| **Database** | [PostgreSQL 18](https://www.postgresql.org/) |
| **DB Driver** | [lib/pq](https://github.com/lib/pq) (native `database/sql`) |
| **Cache** | [Redis 8](https://redis.io/) (provisioned, future use) |
| **API Docs** | [Swaggo/Swag](https://github.com/swaggo/swag) + Gin Swagger UI |
| **Config** | [godotenv](https://github.com/joho/godotenv) |
| **Hot Reload** | [Air](https://github.com/air-verse/air) |
| **Containerisation** | Docker + Docker Compose |

---

## Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (recommended)
- Or: a local PostgreSQL 14+ instance

### Environment Variables

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|---|---|---|
| `APP_NAME` | `Medication Tracker` | Application name |
| `API_PORT` | `5010` | Port the API listens on |
| `ENV` | `development` | Runtime environment |
| `POSTGRES_USER` | `postgres` | PostgreSQL username |
| `POSTGRES_PASSWORD` | *(required)* | PostgreSQL password |
| `POSTGRES_DB` | `med_sys` | PostgreSQL database name |
| `POSTGRES_PORT` | `5432` | PostgreSQL port inside container |
| `POSTGRES_PUBLISHED_PORT` | `5402` | PostgreSQL port exposed on host |
| `REDIS_PORT` | `6379` | Redis port inside container |
| `REDIS_PUBLISHED_PORT` | `6302` | Redis port exposed on host |
| `REDIS_PASSWORD` | *(empty)* | Redis auth password |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,...` | Comma-separated allowed origins |
| `APNs_KEY_ID` | *(optional)* | Apple Push Notification key ID |
| `APNs_TEAM_ID` | *(optional)* | Apple Push Notification team ID |
| `FCM_API_KEY` | *(optional)* | Firebase Cloud Messaging API key |

### Running with Docker

This is the recommended approach. It starts the API, PostgreSQL, and Redis with a single command, and enables live-reload via Air.

```bash
# 1. Clone the repository
git clone https://github.com/jaygaha/medication-tracker-api.git
cd medication-tracker-api

# 2. Set up environment
cp .env.example .env
# Edit .env with your values

# 3. Start all services
docker compose up --build

# API will be available at: http://localhost:5010
# Swagger UI:               http://localhost:5010/swagger/index.html
```

To run in detached mode:

```bash
docker compose up --build -d
```

To stop and remove volumes:

```bash
docker compose down -v
```

### Running Locally

If you prefer to run the API directly without Docker (you still need a running PostgreSQL instance):

```bash
# Install dependencies
go mod download

# Install swag CLI for doc generation
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
swag init -g cmd/server/main.go

# Run the server
go run ./cmd/server/main.go
```

With hot reload (requires [Air](https://github.com/air-verse/air)):

```bash
# Install Air
go install github.com/air-verse/air@latest

# Start with hot reload
air -c .air.toml
```

**Database migrations** are run automatically on startup. The seed data (sample user, medications, schedules, and logs) is inserted only on a fresh, empty database.

---

## API Reference

The base path for all API endpoints is `/api/v1`. Interactive documentation is available at:

```
http://localhost:5010/swagger/index.html
```

### Health Check

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/health` | Returns `{ "status": "healthy" }` |

### Medications

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/medications` | Create a new medication |
| `GET` | `/api/v1/medications` | List all medications (paginated) |
| `GET` | `/api/v1/medications/:id` | Get a single medication by UUID |
| `PUT` | `/api/v1/medications/:id` | Update a medication |
| `DELETE` | `/api/v1/medications/:id` | Soft-delete a medication |

#### List Query Parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| `page` | `int` | `1` | Page number |
| `limit` | `int` | `20` | Items per page |
| `order-by` | `string` | `created_at` | Sort field (`name`, `form`, `status`, `created_at`, etc.) |
| `order-dir` | `string` | `desc` | Sort direction: `asc` or `desc` |

#### Create / Update Request Body

```json
{
  "name": "Ibuprofen",
  "form": "tablet",
  "strength_value": 200,
  "strength_unit": "mg",
  "rx_number": "RX-123456",
  "notes": "Take with food to prevent upset stomach.",
  "status": "active"
}
```

#### Medication Object Response

```json
{
  "id": "22222222-2222-2222-2222-222222222222",
  "name": "Ibuprofen",
  "form": "tablet",
  "strength_value": 200,
  "strength_unit": "mg",
  "rx_number": "RX-123456",
  "notes": "Take with food to prevent upset stomach.",
  "status": "active",
  "created_at": "2026-04-23T03:45:00Z",
  "updated_at": "2026-04-23T03:45:00Z"
}
```

---

## Project Structure

```
medication-tracker-api/
│
├── cmd/
│   └── server/
│       └── main.go              # Entry point: DI wiring & server start
│
├── internal/
│   ├── config/
│   │   ├── config.go            # Environment variable loading
│   │   └── database.go          # DB init, connection pool, migrations, seeding
│   │
│   ├── errors/
│   │   └── *.go                 # Typed errors (NotFoundError, ValidationError)
│   │
│   ├── handler/
│   │   └── medication_handler.go # HTTP handlers (controllers)
│   │
│   ├── middleware/
│   │   └── attach-user.go       # Injects user_id into Gin context
│   │
│   ├── models/
│   │   ├── medication.go        # Medication domain struct & request types
│   │   └── user.go              # User domain struct
│   │
│   ├── repository/
│   │   └── medication_repo.go   # Raw SQL queries against PostgreSQL
│   │
│   ├── routes/
│   │   └── api.go               # Route group definitions
│   │
│   └── service/
│       └── medication_service.go # Business logic layer
│
├── migrations/
│   ├── 0001_..._create_default_tables.sql  # Full schema (enums, tables, indexes)
│   └── 0002_default_seed.sql               # Sample data for development
│
├── docs/                        # Auto-generated Swagger documentation
├── .air.toml                    # Air hot-reload configuration
├── .env.example                 # Environment variable template
├── compose.yml                  # Docker Compose (api + postgres + redis)
├── Dockerfile                   # Multi-stage build with Air + Swag
├── go.mod
└── go.sum
```

---

## Development

### Generating Swagger Docs

API docs are generated from Go source annotations using `swag`. Run this after changing any handler signature or adding a new route:

```bash
swag init -g cmd/server/main.go
```

### Running Migrations

Migrations are applied automatically at startup. To manually inspect or extend the schema, edit `migrations/0001_..._create_default_tables.sql`.

> **Note:** The current migration runner reads and executes the SQL file on every startup using idempotent `CREATE TABLE IF NOT EXISTS` and `CREATE TYPE IF NOT EXISTS` guards — no versioning table is used yet.

### Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Keep handlers thin — put logic in the service layer
- Use typed errors from `internal/errors` for consistent error handling
- All routes go through the `internal/routes` package — no ad-hoc route registration in `main.go`

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
