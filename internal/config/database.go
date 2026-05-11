// internal/config/database.go
package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// InitDB initializes and returns a database connection
func InitDB(cfg *Config) (*sql.DB, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSLMode)

	log.Printf("Connecting to database: %s/%s", cfg.DB.Host, cfg.DB.DBName)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(25)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")

	// Run Migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed default data
	if err := seed(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to seed database: %w", err)
	}

	return db, nil
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB) error {
	log.Println("Running migrations")

	migrationFiles := []string{
		"migrations/0001_01_01_000000_create_default_tables.sql",
		"migrations/0003_device_tokens.sql",
		"migrations/0004_notification_logs.sql",
	}

	for _, file := range migrationFiles {
		schemaBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		_, err = db.Exec(string(schemaBytes))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
		log.Printf("Migration applied: %s", file)
	}

	log.Println("All migrations applied successfully")
	return nil
}

// Seed default data
func seed(db *sql.DB) error {
	log.Println("Seeding database")

	// ONLY RUN IF DATABASE IS EMPTY
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	if count > 0 {
		log.Println("Database already seeded, skipping")
		return nil
	}
	log.Println("Database is empty, seeding")

	// Seed default user
	// Read the SQL schema file.
	seedBytes, err := os.ReadFile("migrations/0002_default_seed.sql")
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	seedQuery := string(seedBytes)

	// Execute the schema query
	_, err = db.Exec(seedQuery)
	if err != nil {
		return fmt.Errorf("failed to execute seed query: %w", err)
	}
	log.Println("Default data seeded successfully")

	return nil
}
