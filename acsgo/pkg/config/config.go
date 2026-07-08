// Package config provides centralized configuration management for ACSGO.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server
	Env         string
	WebPort     int
	CWMPPort    int
	GRPCPort    int
	MetricsPort int

	// PostgreSQL
	DatabaseURL string

	// Redis
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// RabbitMQ
	RabbitMQURL string

	// JWT
	JWTSecret  string
	JWTExpiry  time.Duration

	// CWMP
	ACSUsername string
	ACSPassword string
	ACSURL      string

	// Firmware
	FirmwareBaseURL     string
	FirmwareStoragePath string

	// gRPC
	GRPCNodeList []string

	// Logging
	LogLevel string

	// Metrics
	PrometheusEnabled bool

	// Automation
	WebhookSecret string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Env:         getEnv("ENV", "development"),
		WebPort:     getEnvInt("PORT", 7548),
		CWMPPort:    getEnvInt("CWMP_PORT", 7547),
		GRPCPort:    getEnvInt("GRPC_PORT", 50051),
		MetricsPort: getEnvInt("METRICS_PORT", 9090),

		DatabaseURL: getEnv("DATABASE_URL", "******localhost:5432/acsgo?sslmode=disable"),

		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		RabbitMQURL: getEnv("RABBITMQ_URL", "******localhost:5672/"),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiry: getEnvDuration("JWT_EXPIRY", 24*time.Hour),

		ACSUsername: getEnv("ACS_USERNAME", "acsadmin"),
		ACSPassword: getEnv("ACS_PASSWORD", ""),
		ACSURL:      getEnv("ACS_URL", "http://localhost:7547/cwmp"),

		FirmwareBaseURL:     getEnv("FIRMWARE_BASE_URL", "http://localhost:7548/firmware/"),
		FirmwareStoragePath: getEnv("FIRMWARE_STORAGE_PATH", "/opt/acsgo/firmware"),

		LogLevel:          getEnv("LOG_LEVEL", "info"),
		PrometheusEnabled: getEnvBool("PROMETHEUS_ENABLED", true),
		WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
