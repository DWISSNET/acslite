package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	CWMPPort int
	WebPort  int

	// Database
	DBType string // "sqlite" or "postgres"
	DBURL  string // SQLite file path or PostgreSQL DSN

	// Auth
	JWTSecret     string
	AdminEmail    string
	AdminPassword string

	// App
	Env string
}

var App *Config

// Load reads configuration from environment (and optional .env file)
func Load() *Config {
	// Load .env file if it exists (ignore error if absent)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cwmpPort, _ := strconv.Atoi(getEnv("CWMP_PORT", "7547"))
	webPort, _ := strconv.Atoi(getEnv("PORT", "7548"))

	App = &Config{
		CWMPPort:      cwmpPort,
		WebPort:       webPort,
		DBType:        getEnv("DB_TYPE", "sqlite"),
		DBURL:         getEnv("DB_URL", "./data/acsgo.db"),
		JWTSecret:     getEnv("JWT_SECRET", "acsgo-secret-key-2026"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@acsgo.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		Env:           getEnv("ENV", "production"),
	}

	return App
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
