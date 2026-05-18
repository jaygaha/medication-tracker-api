# 💊 Medication Tracker API

> Your personal health assistant's intelligent backend. Inspired by Apple Health's medication feature, this API is built with Go, Gin, and PostgreSQL to help users manage their prescriptions, schedules, and adherence logs securely and effortlessly.

---

## 🌟 Welcome!

Managing medications shouldn't be a headache. The Medication Tracker API provides a robust, stateless, and timezone-aware foundation for building beautiful mobile or web applications that keep users on track with their health. 

With this API, your application can empower users to:
- **Easily add medications** with all the necessary clinical details (name, form, strength, prescription number).
- **Visually identify pills** by saving shapes and colors, making the medicine cabinet less confusing.
- **Set up flexible schedules** that fit their life—whether it's daily, every few days, specific days of the week, or just "as needed".
- **Track their adherence** by logging every dose as *taken* or *skipped* to build healthy streaks.
- **Stay safe** with automated drug interaction warnings and severity flags.
- **Never miss a dose** with intelligent push notifications delivered right when they need them.

---

## ✨ Features

### ✅ What's Built (Ready to Use)

| Feature | What it does |
|---|---|
| 💊 **Medication Management** | Full CRUD for medications with detailed clinical metadata. |
| 🎨 **Medication Visuals** | Help users recognize their meds by storing shape and color data. |
| 🗓 **Smart Schedules** | Flexible scheduling system (`daily`, `intervals`, `specific_days`, `as_needed`). |
| 📋 **Adherence Logs** | Keep a historical record of taken and skipped doses. |
| 📊 **Adherence Stats** | Track progress with taken/skipped counts, adherence rates, and streaks! |
| ⚠️ **Drug Interactions** | Store, query, and acknowledge interaction warnings. |
| 🔔 **Push Notifications** | Dose reminder alerts via **APNs** and **FCM** with strict deduplication and user timezone support. |
| 🧠 **Intelligent Scheduler** | Background worker that accurately calculates localized trigger times based on user schedules. |
| 🔍 **Paginated Listing** | Fast, sortable, and paginated medication lists. |
| 🩺 **Swagger Docs** | Interactive, beautifully generated API documentation. |
| 🐳 **Docker Ready** | Quick setup with Docker Compose and hot-reloading via Air. |
| 🔒 **Soft Deletes** | History is preserved even when records are deleted. |
| 👤 **User Profiles** | View and update basic user information and timezone settings. |

### 🚧 What's Next (Planned)

We're always looking to improve. Here is what is on the roadmap:

| # | Feature | Description |
|---|---|---|
| 1 | 👤 **User Authentication** | Full registration, login, JWT issuance, and token refresh flows. |
| 2 | ⚡ **Redis Caching** | Supercharge read performance for medication lists and schedules. |
| 3 | 🔔 **Notification Upgrades** | User preferences and real SDK integration for Push Notifications. |
| 4 | 🤖 **AI Prompt Generator** | Transform initial user thoughts into structured AI prompts (Simple, Advanced, Expert). |

---

## 🏗 How It's Built (Architecture)

We believe in clean, maintainable code. The project follows a strict **layered architecture**, separating concerns to make scaling and testing a breeze:

```text
Request → Router → Middleware → Handler → Service → Repository → PostgreSQL
```

| Layer | Where it lives | What it does |
|---|---|---|
| **Entry Point** | `cmd/server` | Bootstraps the app, wires up dependencies, and starts the engine. |
| **Config** | `internal/config` | Loads your environment, initializes the DB, and runs migrations. |
| **Routes** | `internal/routes` | Maps URLs to the right handlers. |
| **Middleware** | `internal/middleware` | Enriches requests (like injecting the user context). |
| **Handler** | `internal/handler` | Parses incoming HTTP requests and formats outgoing responses. |
| **Service** | `internal/service` | The brain of the operation! All business logic lives here. |
| **Repository**| `internal/repository`| Talks to the database using raw, optimized SQL queries. |
| **Models** | `internal/models` | Defines the data structures and shapes. |
| **Errors** | `internal/errors` | Keeps our error handling typed and consistent. |

Dependencies flow cleanly from top to bottom through constructor injection—no messy global state!

---

## 💾 The Database

Our schema is designed for clinical accuracy and future growth. We use `UUID`s for all primary keys and `TIMESTAMP WITH TIME ZONE` to handle time globally.

```text
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

---

## 🛠 Tech Stack

We use a modern, blazing-fast stack:

- **Language:** [Go 1.26+](https://golang.org/)
- **Web Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Database:** [PostgreSQL 18](https://www.postgresql.org/) (via `database/sql` & `lib/pq`)
- **Cache:** [Redis 8](https://redis.io/) (provisioned for future use)
- **API Docs:** [Swaggo/Swag](https://github.com/swaggo/swag)
- **Dev Tools:** [Air](https://github.com/air-verse/air) (Hot Reloading), Docker

---

## 🚀 Getting Started

Ready to run this locally? Let's get you set up!

### Prerequisites
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Highly recommended for the smoothest experience)
- Or: Go 1.26+ and a local PostgreSQL instance.

### 1. Configure your environment

Copy the example environment file and tweak it if needed:

```bash
cp .env.example .env
```

### 2. Fire it up with Docker! 🐳

This is the easiest way to run the API, Database, and Redis all at once. Plus, it includes live-reloading!

```bash
# Start all services
docker compose up --build

# That's it! 
# Your API is running at:      http://localhost:5010
# Interactive API Docs are at: http://localhost:5010/swagger/index.html
```

*(To run silently in the background, just add `-d`: `docker compose up --build -d`)*

---

## 📚 API Reference

Explore and test the API directly from your browser using our beautifully generated Swagger UI:
👉 **[http://localhost:5010/swagger/index.html](http://localhost:5010/swagger/index.html)**

### Quick Glance:
- **`GET /api/v1/health`** - Check if the system is alive and kicking!
- **`POST /api/v1/medications`** - Add a new medication.
- **`GET /api/v1/medications`** - List medications (supports pagination and sorting).
- **`POST /api/v1/device-tokens`** - Register a device to start receiving push notifications!
- **`GET /api/v1/users/profile`** - Retrieve the currently authenticated user's profile.
- **`PATCH /api/v1/users/profile`** - Update basic information and timezone settings.

---

## 👨‍💻 For Developers

### Updating the API Docs
Whenever you add a new endpoint or change a response, regenerate the swagger docs:
```bash
swag init -g cmd/server/main.go
```

### Database Migrations
Migrations are smart and run automatically when the server starts! Need to tweak the schema? Just update `migrations/0001_..._create_default_tables.sql` or add a new migration file to the folder.

### Code Style Guidelines
- **Keep it Go-idiomatic:** Use `gofmt` and keep things simple.
- **Thin Handlers, Fat Services:** HTTP Handlers should just parse requests and format responses. All the actual "thinking" belongs in the Service layer.
- **Error Handling:** Always use our custom typed errors from the `internal/errors` package so the frontend gets consistent, helpful error messages.

---

## 🤝 Let's Build Together

Found a bug? Have a great idea for a feature? We'd love your help! Check out [CONTRIBUTING.md](CONTRIBUTING.md) to see how you can get involved.

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for all the legal details.
