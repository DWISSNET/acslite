package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

type device struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	LastSeenAt string `json:"last_seen_at"`
}

type serverState struct {
	mu      sync.RWMutex
	started time.Time
	devices []device
}

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ACSGO Dashboard</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 1.5rem; line-height: 1.4; }
    h1, h2 { margin-bottom: .5rem; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit,minmax(220px,1fr)); gap: .75rem; margin-bottom: 1rem; }
    .card { border: 1px solid #ddd; border-radius: 8px; padding: .75rem; }
    table { width: 100%; border-collapse: collapse; }
    th, td { border: 1px solid #ddd; padding: .5rem; text-align: left; }
    textarea { width: 100%; min-height: 100px; }
    .muted { color: #666; font-size: .9rem; }
  </style>
</head>
<body>
  <h1>ACSGO Dashboard</h1>
  <p class="muted">Live updates via API polling every 2s.</p>
  <div class="grid">
    <div class="card"><strong>Status:</strong> <span id="status">-</span></div>
    <div class="card"><strong>Total devices:</strong> <span id="total">0</span></div>
    <div class="card"><strong>Online devices:</strong> <span id="online">0</span></div>
    <div class="card"><strong>Uptime (s):</strong> <span id="uptime">0</span></div>
  </div>

  <h2>Bulk Import Devices</h2>
  <p class="muted">One device per line: <code>id,name,status</code></p>
  <textarea id="bulk" placeholder="router-01,Main Router,online"></textarea>
  <p><button id="importBtn">Import</button> <span id="msg" class="muted"></span></p>

  <h2>Device List</h2>
  <table>
    <thead><tr><th>ID</th><th>Name</th><th>Status</th><th>Last Seen</th></tr></thead>
    <tbody id="devices"></tbody>
  </table>

  <script>
    async function loadStats() {
      const r = await fetch('/api/stats');
      const s = await r.json();
      document.getElementById('status').textContent = s.status;
      document.getElementById('total').textContent = s.total_devices;
      document.getElementById('online').textContent = s.online_devices;
      document.getElementById('uptime').textContent = s.uptime_seconds;
    }
    async function loadDevices() {
      const r = await fetch('/api/devices');
      const payload = await r.json();
      const tbody = document.getElementById('devices');
      tbody.innerHTML = '';
      const list = payload.devices || [];
      if (list.length === 0) {
        const tr = document.createElement('tr');
        const td = document.createElement('td');
        td.setAttribute('colspan', '4');
        td.textContent = 'No devices';
        tr.appendChild(td);
        tbody.appendChild(tr);
      } else {
        list.forEach(function(d) {
          const tr = document.createElement('tr');
          [d.id, d.name, d.status, d.last_seen_at].forEach(function(val) {
            const td = document.createElement('td');
            td.textContent = val || '';
            tr.appendChild(td);
          });
          tbody.appendChild(tr);
        });
      }
    }
    function parseBulkInput(input) {
      return input.split('\n').map(function(line) { return line.trim(); }).filter(Boolean).map(function(line) {
        const parts = line.split(',');
        return { id: (parts[0] || '').trim(), name: (parts[1] || '').trim(), status: (parts[2] || 'offline').trim() };
      });
    }
    async function importDevices() {
      const raw = document.getElementById('bulk').value;
      const devices = parseBulkInput(raw).filter(function(d) { return d.id.length > 0; });
      if (!devices.length) {
        document.getElementById('msg').textContent = 'No valid lines to import.';
        return;
      }
      const r = await fetch('/api/devices/import', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({devices: devices})
      });
      const p = await r.json();
      document.getElementById('msg').textContent = 'Imported: ' + (p.imported || 0);
      await refresh();
    }
    async function refresh() {
      await Promise.all([loadStats(), loadDevices()]);
    }
    document.getElementById('importBtn').addEventListener('click', importDevices);
    refresh();
    setInterval(refresh, 2000);
  </script>
</body>
</html>`

func main() {
	cfg := loadConfig()
	warnInsecureDefaults(cfg)
	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("database setup failed: %v", err)
	}
	defer db.Close()

	mux := newMux(cfg, db)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("acsgo listening on %s (cwmp_port=%s db_type=%s)", srv.Addr, cfg.CWMPPort, cfg.DBType)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

func newMux(cfg config, db *sql.DB) *http.ServeMux {
	state := &serverState{
		started: time.Now(),
		devices: make([]device, 0),
	}

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
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, dashboardHTML)
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		state.mu.RLock()
		total := len(state.devices)
		online := 0
		for _, d := range state.devices {
			if strings.EqualFold(d.Status, "online") {
				online++
			}
		}
		started := state.started
		state.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":         "ok",
			"db_type":        cfg.DBType,
			"cwmp_port":      cfg.CWMPPort,
			"total_devices":  strconv.Itoa(total),
			"online_devices": strconv.Itoa(online),
			"uptime_seconds": strconv.FormatInt(int64(time.Since(started).Seconds()), 10),
		})
	})
	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		state.mu.RLock()
		devices := append([]device(nil), state.devices...)
		state.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string][]device{"devices": devices})
	})
	mux.HandleFunc("/api/devices/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB limit
		var payload struct {
			Devices []device `json:"devices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		imported := 0
		state.mu.Lock()
		for _, d := range payload.Devices {
			id := strings.TrimSpace(d.ID)
			if id == "" {
				continue
			}
			item := device{
				ID:         id,
				Name:       strings.TrimSpace(d.Name),
				Status:     strings.TrimSpace(d.Status),
				LastSeenAt: now,
			}
			if item.Status == "" {
				item.Status = "offline"
			}
			replaced := false
			for i := range state.devices {
				if state.devices[i].ID == item.ID {
					state.devices[i] = item
					replaced = true
					break
				}
			}
			if !replaced {
				state.devices = append(state.devices, item)
			}
			imported++
		}
		state.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"imported": imported})
	})
	return mux
}

func loadConfig() config {
	cfg := config{
		Port:          getEnv("PORT", "7548"),
		CWMPPort:      getEnv("CWMP_PORT", "7547"),
		DBType:        strings.ToLower(getEnv("DB_TYPE", "sqlite")),
		DBURL:         getEnv("DB_URL", "./data/acsgo.db"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@acsgo.local"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "change-me"),
		JWTSecret:     getEnv("JWT_SECRET", "change-this-secret"),
	}
	if cfg.DBType == "" {
		cfg.DBType = "sqlite"
	}
	if cfg.DBURL == "" {
		cfg.DBURL = "./data/acsgo.db"
	}
	return cfg
}

func warnInsecureDefaults(cfg config) {
	if cfg.JWTSecret == "change-me" || cfg.JWTSecret == "change-this-secret" {
		log.Print("WARNING: JWT_SECRET is set to the default value; set a strong random secret before production use")
	}
	if cfg.AdminPassword == "admin123" || cfg.AdminPassword == "change-me" {
		log.Print("WARNING: ADMIN_PASSWORD is set to the default value; change it before production use")
	}
}

func openDB(cfg config) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)
	switch cfg.DBType {
	case "postgres", "postgresql":
		if cfg.DBURL == "" {
			return nil, errors.New("DB_URL is required when DB_TYPE=postgres")
		}
		db, err = sql.Open("postgres", cfg.DBURL)
	case "sqlite", "sqlite3", "":
		ensureSQLitePath(cfg.DBURL)
		db, err = sql.Open("sqlite", cfg.DBURL)
	default:
		return nil, fmt.Errorf("unsupported DB_TYPE %q (supported: sqlite, postgres)", cfg.DBType)
	}
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if pingErr := db.PingContext(ctx); pingErr != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", pingErr)
	}
	return db, nil
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
