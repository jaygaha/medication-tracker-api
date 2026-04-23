package config

import (
	"os"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
}

// Server Config
type ServerConfig struct {
	Name string
	Host string
	Port string
	Env  string
}

// DB Config
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Redis Config
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// LoadConfig Load the application configuration
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Name: getEnv("APP_NAME", "Medication System"),
			Host: getEnv("API_HOST", "localhost"),
			Port: getEnv("API_PORT", "5010"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "k_med"),
			Password: getEnv("DB_PASSWORD", "MedAdmin123#"),
			DBName:   getEnv("DB_NAME", "med_sys"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}
