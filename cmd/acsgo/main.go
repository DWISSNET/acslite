package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type config struct {
	Port          string
	CWMPPort      string
	DBType        string
	DBURL         string
	AdminEmail    string
	AdminPassword string
	JWTSecret     string
}

func main() {
	cfg := loadConfig()
	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, fmt.Sprintf("database unreachable: %v", err), http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"db_type":   cfg.DBType,
			"cwmp_port": cfg.CWMPPort,
		})
	})

	addr := ":" + cfg.Port
	log.Printf("acsgo listening on %s (cwmp_port=%s db_type=%s)", addr, cfg.CWMPPort, cfg.DBType)
	if err := http.ListenAndServe(addr, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadConfig() config {
	cfg := config{
		Port:          getEnv("PORT", "7548"),
		CWMPPort:      getEnv("CWMP_PORT", "7547"),
		DBType:        strings.ToLower(getEnv("DB_TYPE", "sqlite")),
		DBURL:         getEnv("DB_URL", "./data/acsgo.db"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@acsgo.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me"),
	}
	if cfg.DBType == "" {
		cfg.DBType = "sqlite"
	}
	if cfg.DBURL == "" {
		cfg.DBURL = "./data/acsgo.db"
	}
	return cfg
}

func openDB(cfg config) (*sql.DB, error) {
	switch cfg.DBType {
	case "postgres", "postgresql":
		if cfg.DBURL == "" {
			return nil, errors.New("DB_URL is required when DB_TYPE=postgres")
		}
		return sql.Open("postgres", cfg.DBURL)
	case "sqlite", "sqlite3", "":
		ensureSQLitePath(cfg.DBURL)
		return sql.Open("sqlite", cfg.DBURL)
	default:
		return nil, fmt.Errorf("unsupported DB_TYPE %q (supported: sqlite, postgres)", cfg.DBType)
	}
}

func ensureSQLitePath(dbURL string) {
	path := dbURL
	if strings.HasPrefix(dbURL, "file:") {
		u, err := url.Parse(dbURL)
		if err == nil {
			path = u.Path
		}
	}
	if path == "" || path == ":memory:" {
		return
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("warning: failed creating sqlite directory %s: %v", dir, err)
	}
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
