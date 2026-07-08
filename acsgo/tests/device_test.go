package tests

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/services"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	// Execute each statement
	stmts := splitSQL(string(schema))
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Logf("schema stmt warning: %v (stmt: %.60s)", err, stmt)
		}
	}
	return db
}

func splitSQL(s string) []string {
	var stmts []string
	var cur []byte
	for _, b := range []byte(s) {
		cur = append(cur, b)
		if b == ';' {
			stmt := trimSpace(string(cur))
			if stmt != "" && stmt != ";" {
				stmts = append(stmts, stmt)
			}
			cur = nil
		}
	}
	return stmts
}

func trimSpace(s string) string {
	start, end := 0, len(s)-1
	for start < len(s) && (s[start] == ' ' || s[start] == '\n' || s[start] == '\r' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end] == ' ' || s[end] == '\n' || s[end] == '\r' || s[end] == '\t') {
		end--
	}
	if start > end {
		return ""
	}
	return s[start : end+1]
}

// ========== Device Service Tests ==========

func TestDeviceUpsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	now := time.Now()
	d := &models.Device{
		SerialNumber:     "TESTHWT00001",
		DeviceID:         "000000-TESTHWT00001",
		Manufacturer:     "huawei",
		ModelName:        "HG8245H",
		IPAddress:        "192.168.1.100",
		ConnectionStatus: "online",
		LastInform:       &now,
		Parameters:       models.JSONMap{"test": "value"},
	}

	if err := svc.Upsert(d); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if d.ID == "" {
		t.Error("expected ID to be set after upsert")
	}

	// Retrieve
	got, err := svc.GetBySerial("TESTHWT00001")
	if err != nil {
		t.Fatalf("GetBySerial failed: %v", err)
	}
	if got.Manufacturer != "huawei" {
		t.Errorf("expected manufacturer=huawei, got %s", got.Manufacturer)
	}
	if got.ModelName != "HG8245H" {
		t.Errorf("expected model=HG8245H, got %s", got.ModelName)
	}
	if got.Parameters["test"] != "value" {
		t.Errorf("expected parameter test=value, got %v", got.Parameters["test"])
	}
}

func TestDeviceList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	for i := 0; i < 5; i++ {
		svc.Upsert(&models.Device{
			SerialNumber:     "SN" + string(rune('A'+i)),
			DeviceID:         "ID" + string(rune('A'+i)),
			Manufacturer:     "zte",
			ConnectionStatus: "offline",
			Parameters:       models.JSONMap{},
		})
	}

	devices, total, err := svc.List(1, 10, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total=5, got %d", total)
	}
	if len(devices) != 5 {
		t.Errorf("expected 5 devices, got %d", len(devices))
	}
}

func TestDeviceSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	svc.Upsert(&models.Device{SerialNumber: "HUAWEI001", DeviceID: "ID1", Manufacturer: "huawei", ModelName: "HG8245H", ConnectionStatus: "online", Parameters: models.JSONMap{}})
	svc.Upsert(&models.Device{SerialNumber: "ZTE0001", DeviceID: "ID2", Manufacturer: "zte", ModelName: "F609", ConnectionStatus: "offline", Parameters: models.JSONMap{}})

	devices, total, err := svc.List(1, 10, "HUAWEI")
	if err != nil {
		t.Fatalf("List search failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 result for search=HUAWEI, got %d", total)
	}
	if len(devices) > 0 && devices[0].Manufacturer != "huawei" {
		t.Errorf("wrong device returned: %s", devices[0].Manufacturer)
	}
}

func TestDeviceMarkOffline(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	old := time.Now().Add(-30 * time.Minute)
	svc.Upsert(&models.Device{
		SerialNumber:     "OLD001",
		DeviceID:         "IDOLD",
		Manufacturer:     "huawei",
		ConnectionStatus: "online",
		LastInform:       &old,
		Parameters:       models.JSONMap{},
	})

	n, err := svc.MarkOffline(15 * time.Minute)
	if err != nil {
		t.Fatalf("MarkOffline failed: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 device marked offline, got %d", n)
	}
}

func TestDeviceLogEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	svc.LogEvent("SN-TEST", "test_event", "Test description")

	events, err := svc.GetEvents("SN-TEST", 10)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "test_event" {
		t.Errorf("expected event_type=test_event, got %s", events[0].EventType)
	}
}

func TestDeviceStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := services.NewDeviceService(db)
	now := time.Now()
	svc.Upsert(&models.Device{SerialNumber: "S1", DeviceID: "D1", ConnectionStatus: "online", LastInform: &now, Parameters: models.JSONMap{}})
	svc.Upsert(&models.Device{SerialNumber: "S2", DeviceID: "D2", ConnectionStatus: "offline", Parameters: models.JSONMap{}})

	stats, err := svc.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats["total"].(int) != 2 {
		t.Errorf("expected total=2, got %v", stats["total"])
	}
	if stats["online"].(int) != 1 {
		t.Errorf("expected online=1, got %v", stats["online"])
	}
}

func TestDetectModel(t *testing.T) {
	modem := services.DetectModel("Huawei Technologies Co., Ltd", "HG8245H", "HWTHG8245H001")
	if modem.Model != "HG8245H" {
		t.Errorf("expected model=HG8245H, got %s", modem.Model)
	}
	if modem.Manufacturer != "huawei" {
		t.Errorf("expected manufacturer=huawei, got %s", modem.Manufacturer)
	}
}
