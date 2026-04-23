# Contributing to Medication Tracker API

Thank you for your interest in contributing! This document describes how to set up your development environment, follow the project's conventions, and submit a great pull request.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Development Setup](#development-setup)
- [Project Conventions](#project-conventions)
  - [Architecture Rules](#architecture-rules)
  - [Go Package Conventions](#go-package-conventions)
  - [Code Style](#code-style)
  - [Error Handling](#error-handling)
  - [Database Changes](#database-changes)
  - [API Changes](#api-changes)
- [Git Workflow](#git-workflow)
  - [Branch Naming](#branch-naming)
  - [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

---

## Code of Conduct

This project is a welcoming environment for all contributors. We expect everyone to:

- Be respectful and constructive in all interactions
- Focus feedback on code, not people
- Welcome differing viewpoints and experiences
- Gracefully accept constructive criticism

---

## Ways to Contribute

| Type | How |
|---|---|
| 🐛 Bug fix | Open an issue, then submit a PR with the fix |
| ✨ New feature | Open a discussion issue first to align on design |
| 📖 Documentation | Edit the relevant `.md` files or Go doc comments |
| 🧹 Refactor | Must keep existing behaviour intact; include rationale |
| 🧪 Tests | Always welcome — we aim to increase test coverage |
| 🗃 Schema change | Requires a new migration file; see [Database Changes](#database-changes) |

---

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repo on GitHub first, then:
git clone https://github.com/<your-username>/medication-tracker-api.git
cd medication-tracker-api
git remote add upstream https://github.com/jaygaha/medication-tracker-api.git
```

### 2. Environment

```bash
cp .env.example .env
# Edit .env — set at minimum POSTGRES_PASSWORD
```

### 3. Start Infrastructure

```bash
docker compose up -d postgres redis
```

### 4. Install Go Tools

```bash
# Swagger doc generator
go install github.com/swaggo/swag/cmd/swag@latest

# Hot reload (optional but recommended)
go install github.com/air-verse/air@latest
```

### 5. Run the Server

```bash
# With hot reload
air -c .air.toml

# Or directly
go run ./cmd/server/main.go
```

The server starts at `http://localhost:5010`.  
Swagger UI: `http://localhost:5010/swagger/index.html`

---

## Project Conventions

### Architecture Rules

This project uses a strict **layered architecture**. Respect these rules:

```
Handler → Service → Repository
```

| Rule | Rationale |
|---|---|
| **Handlers must not contain business logic.** Parse the request, call a service method, return the response. | Keeps controllers thin and testable |
| **Services must not execute SQL.** All database access goes through a repository method. | Allows swapping the data layer independently |
| **Repositories must return domain models, not raw `sql.Row`.** | Keeps DB details contained |
| **No global state.** All dependencies are injected via constructors. | Testability and clarity |
| **No ad-hoc route registration in `main.go`.** All routes go in `internal/routes`. | Single source of truth for the API surface |

### Go Package Conventions

This project follows the [standard Go project layout](https://github.com/golang-standards/project-layout) and the conventions described in [Effective Go](https://go.dev/doc/effective_go).

#### Module & Import Path

The module is declared as:

```
module github.com/jaygaha/medication-tracker-api
```

All internal imports must use this full path. Never use relative imports (e.g., `./internal/models`).

```go
// ✅ correct
import "github.com/jaygaha/medication-tracker-api/internal/models"

// ❌ wrong
import "./internal/models"
```

#### Package Naming

| Rule | Example |
|---|---|
| Package names are **lowercase**, single words — no underscores or camelCase | `package handler` not `package medicationHandler` |
| Package name should match the directory name | `internal/handler/` → `package handler` |
| Avoid stutter between package and exported name | `errors.NotFoundError` not `errors.ErrorNotFound` |
| Test files use `_test` suffix on the package name for black-box tests | `package handler_test` |

#### Directory Layout Rules

| Directory | Purpose | Visibility |
|---|---|---|
| `cmd/server/` | Binary entry point only — wiring + `main()` | Public |
| `internal/` | All application code — **cannot** be imported by external modules | Private |
| `internal/config/` | Environment & infrastructure init | Internal |
| `internal/models/` | Domain structs shared across layers | Internal |
| `internal/errors/` | Typed application errors | Internal |
| `internal/handler/` | HTTP layer (Gin handlers) | Internal |
| `internal/service/` | Business logic | Internal |
| `internal/repository/` | Data access (raw SQL) | Internal |
| `internal/routes/` | Route registration | Internal |
| `internal/middleware/` | Gin middleware | Internal |
| `migrations/` | SQL migration files | Data only |
| `docs/` | Auto-generated Swagger output — **do not edit manually** | Generated |

> **Why `internal/`?** The `internal` directory is a Go toolchain enforcement mechanism. No external module can import packages inside `internal/`, protecting the API boundary.

#### Interfaces

Define interfaces at the **consumer** side (in the package that uses the behaviour), not in the package that implements it. This keeps dependencies pointing inward and makes testing with mocks trivial.

```go
// internal/service/medication_service.go — define what the service NEEDS
type MedicationStore interface {
    Create(ctx context.Context, m *models.Medication) error
    GetByID(ctx context.Context, id string) (*models.Medication, error)
    List(ctx context.Context, userID string, limit, offset int, orderBy, orderDir string) ([]*models.Medication, error)
    Update(ctx context.Context, m *models.Medication) error
    Delete(ctx context.Context, id string) error
}
```

---

### Code Style

- Format all code with **`goimports`** (a superset of `gofmt` that also manages imports) before committing.
- Run **`go vet ./...`** — fix all reported issues.
- Run **`staticcheck ./...`** for deeper linting (`go install honnef.co/go/tools/cmd/staticcheck@latest`).
- Keep function and method bodies short; prefer early returns (guard clauses).
- Name interfaces by what they *do*, not what they *are* (e.g., `MedicationStore`, not `IMedicationRepository`).
- Export only what needs to be exported — when in doubt, keep it unexported.
- Use `context.Context` as the **first** parameter of every function that touches the database or network.
- Group imports in three blocks separated by blank lines: stdlib → external → internal.

```go
import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jaygaha/medication-tracker-api/internal/errors"
	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/service"
)
```

### Error Handling

Use the typed errors in `internal/errors`:

```go
// Not found — returns a 404 in the handler
return errors.NewNotFoundError("medication", id)

// Validation — returns a 400 in the handler
return errors.NewValidationError("name is required")
```

**Never** use `fmt.Errorf("not found")` as a raw string for sentinel errors — it makes type-checking in handlers impossible.

Wrap infrastructure errors with context:

```go
if err != nil {
    return nil, fmt.Errorf("MedicationRepository.GetByID: %w", err)
}
```

### Database Changes

All schema changes must be expressed as new numbered SQL migration files in the `migrations/` directory.

**Naming convention:**

```
migrations/NNNN_<timestamp>_<short_description>.sql
```

Example:
```
migrations/0003_01_01_000000_add_refill_reminders.sql
```

Rules:
- Use `CREATE TABLE IF NOT EXISTS` and `DO $$ BEGIN IF NOT EXISTS ... END $$` guards for idempotency.
- Never drop a column in a migration — use `ALTER TABLE ... ALTER COLUMN` to deprecate.
- Add a corresponding `ROLLBACK` section as a comment if the change is reversible.
- Update `migrations/0002_default_seed.sql` only if the seed data needs updating for the new schema.

### API Changes

- All new endpoints must have Swaggo annotations (`// @Summary`, `// @Router`, etc.)
- Run `swag init -g cmd/server/main.go` after adding or changing annotations.
- New endpoints must be registered in `internal/routes/api.go` and documented in `README.md`.
- Breaking changes (removed fields, changed types) require a version bump (e.g., `/api/v2/...`).

---

## Git Workflow

### Branch Naming

| Type | Pattern | Example |
|---|---|---|
| Feature | `feat/<short-description>` | `feat/add-schedule-endpoint` |
| Bug fix | `fix/<short-description>` | `fix/pagination-offset-calculation` |
| Documentation | `docs/<short-description>` | `docs/update-api-reference` |
| Refactor | `refactor/<short-description>` | `refactor/extract-schedule-service` |
| Database | `db/<short-description>` | `db/add-refill-reminders-table` |

Always branch from an up-to-date `main`:

```bash
git fetch upstream
git checkout -b feat/my-feature upstream/main
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer]
```

**Types:** `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `db`

**Examples:**

```
feat(medications): add paginated list endpoint with sort support

fix(handler): return 404 instead of 500 for missing medication

db(schedule): add schedule_days table for weekday-specific schedules

docs(readme): add API reference table for medication endpoints
```

- Subject line must be ≤ 72 characters.
- Use the imperative mood: *"add"* not *"added"* or *"adds"*.
- Reference issues in the footer: `Closes #42`

---

## Pull Request Process

1. **Ensure your branch is up to date** with `upstream/main`.
2. **All code compiles** — `go build ./...` must succeed with zero errors.
3. **Generate docs** — if you changed any handler: `swag init -g cmd/server/main.go`.
4. **Self-review** — go through your diff as if you were the reviewer.
5. **Open the PR** against `main` with:
   - A clear title following the commit convention
   - A description explaining *what* changed and *why*
   - Reference to any related issues (e.g., `Closes #42`)
   - Screenshots or `curl` examples for API-visible changes

### PR Checklist

```
- [ ] Code compiles: go build ./... passes
- [ ] go vet ./... passes with zero issues
- [ ] goimports has been run on all modified files
- [ ] go mod tidy has been run (go.mod / go.sum are clean)
- [ ] All internal imports use the full module path (github.com/jaygaha/medication-tracker-api/...)
- [ ] New endpoints have Swaggo annotations (@Summary, @Router, etc.)
- [ ] swag init -g cmd/server/main.go has been re-run (docs/ is up to date)
- [ ] New DB changes are in a new numbered migration file under migrations/
- [ ] README.md updated if the public API surface changed
- [ ] No secrets, credentials, or .env files committed
```

PRs that do not meet the checklist will be asked to revise before review.

---

## Reporting Bugs

Open a [GitHub Issue](https://github.com/jaygaha/medication-tracker-api/issues) with:

- **Title**: Short, specific summary (`Pagination returns wrong offset on page 1`)
- **Environment**: Go version, OS, Docker version
- **Steps to reproduce**: Exact `curl` command or request body that triggers the bug
- **Expected behaviour**: What should happen
- **Actual behaviour**: What happens instead, including any error messages or logs

---

## Suggesting Features

Open a [GitHub Issue](https://github.com/jaygaha/medication-tracker-api/issues) with the label `enhancement` and include:

- The **use case** — who needs this and why
- Your **proposed approach** — how it might fit within the existing architecture
- Any **alternatives** you have considered

For significant changes (new domain entities, new endpoints affecting multiple layers), open a discussion issue *before* writing code. This saves everyone time.

---

Thank you for helping make Medication Tracker better! 🙏
