package main

import (
	"bytes"
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

func TestDashboardUsesTextContent(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	// The safe DOM approach must be present; direct innerHTML row-building must not be.
	if !strings.Contains(body, "td.textContent") {
		t.Error("dashboard should use textContent for safe rendering")
	}
	if strings.Contains(body, "innerHTML = rows") {
		t.Error("dashboard must not use raw innerHTML row concatenation (XSS risk)")
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

func TestReadEndpointsMethodRestriction(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	readOnlyEndpoints := []string{"/api/stats", "/api/devices"}
	disallowedMethods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, endpoint := range readOnlyEndpoints {
		for _, method := range disallowedMethods {
			req := httptest.NewRequest(method, endpoint, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s: status = %d, want %d", method, endpoint, rr.Code, http.StatusMethodNotAllowed)
			}
		}
	}
}

func TestImportMethodRestriction(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/devices/import", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /api/devices/import: status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestImportBodySizeLimit(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	// Build a payload larger than 1 MiB
	oversized := make([]byte, 2<<20) // 2 MiB of zeros
	req := httptest.NewRequest(http.MethodPost, "/api/devices/import", bytes.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Error("oversized body should not return 200 OK")
	}
}

func TestHealthEndpoint(t *testing.T) {
	db, h := newTestServer(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("health status field = %q, want %q", resp["status"], "ok")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg := loadConfig()
	if cfg.Port == "" {
		t.Error("Port should have a default value")
	}
	if cfg.DBType == "" {
		t.Error("DBType should have a default value")
	}
	// Ensure default password is not the old insecure "admin123"
	if cfg.AdminPassword == "admin123" {
		t.Error("default AdminPassword must not be 'admin123'")
	}
}
