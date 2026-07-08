package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()
	cfg := config{
		Port:     "7548",
		CWMPPort: "7547",
		DBType:   "sqlite",
		DBURL:    ":memory:",
	}
	db, err := openDB(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db, newMux(cfg, db)
}

func TestDashboardRoute(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "ACSGO Dashboard") {
		t.Fatalf("dashboard body missing expected title")
	}
}

func TestDeviceImportUpdatesStatsAndList(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	importReq := httptest.NewRequest(http.MethodPost, "/api/devices/import", strings.NewReader(`{"devices":[{"id":"router-01","name":"Main","status":"online"},{"id":"router-02","name":"Backup","status":"offline"}]}`))
	importReq.Header.Set("Content-Type", "application/json")
	importRR := httptest.NewRecorder()
	h.ServeHTTP(importRR, importReq)
	if importRR.Code != http.StatusOK {
		t.Fatalf("import status = %d, want %d", importRR.Code, http.StatusOK)
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	statsRR := httptest.NewRecorder()
	h.ServeHTTP(statsRR, statsReq)
	if statsRR.Code != http.StatusOK {
		t.Fatalf("stats status = %d, want %d", statsRR.Code, http.StatusOK)
	}

	var stats map[string]string
	if err := json.Unmarshal(statsRR.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if stats["total_devices"] != "2" || stats["online_devices"] != "1" {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	devicesReq := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	devicesRR := httptest.NewRecorder()
	h.ServeHTTP(devicesRR, devicesReq)
	if devicesRR.Code != http.StatusOK {
		t.Fatalf("devices status = %d, want %d", devicesRR.Code, http.StatusOK)
	}

	var payload struct {
		Devices []device `json:"devices"`
	}
	if err := json.Unmarshal(devicesRR.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode devices: %v", err)
	}
	if len(payload.Devices) != 2 {
		t.Fatalf("devices count = %d, want 2", len(payload.Devices))
	}
}
