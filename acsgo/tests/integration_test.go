package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DWISSNET/acsgo/config"
	"github.com/DWISSNET/acsgo/handlers"
	"github.com/DWISSNET/acsgo/middleware"
	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	// Minimal config for test
	config.App = &config.Config{
		JWTSecret:     "test-secret",
		AdminEmail:    "test@test.local",
		AdminPassword: "testpass123",
	}
}

func setupTestRouter(t *testing.T) (*gin.Engine, *services.DeviceService, *services.RPCService) {
	t.Helper()
	db := setupTestDB(t)
	deviceSvc := services.NewDeviceService(db)
	rpcSvc := services.NewRPCService(db)
	paramSvc := services.NewParameterService(db)
	wsHub := services.NewWSHub()

	r := gin.New()
	r.Use(middleware.CORS())

	authH := handlers.NewAuthHandler()
	deviceH := handlers.NewDeviceHandler(deviceSvc, rpcSvc)
	paramH := handlers.NewParameterHandler(paramSvc)
	provH := handlers.NewProvisioningHandler(deviceSvc)
	sysH := handlers.NewSystemHandler(deviceSvc, wsHub)

	r.POST("/api/auth/login", authH.Login)
	r.GET("/api/health", sysH.Health)

	protected := r.Group("/api", middleware.AuthRequired())
	protected.GET("/stats", sysH.Stats)
	protected.GET("/auth/me", authH.Me)
	protected.GET("/devices", deviceH.List)
	protected.GET("/devices/:serial", deviceH.Get)
	protected.POST("/devices/:serial/reboot", deviceH.Reboot)
	protected.DELETE("/devices/:serial", deviceH.Delete)
	protected.GET("/parameters", paramH.List)
	protected.POST("/provisioning/bulk", provH.BulkImport)
	protected.GET("/provisioning/templates", provH.Templates)
	protected.GET("/provisioning/template/:isp", provH.GetTemplate)

	return r, deviceSvc, rpcSvc
}

func getAuthToken(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": "test@test.local", "password": "testpass123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["token"].(string)
}

// ======== Auth Tests ========

func TestLoginSuccess(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	body, _ := json.Marshal(map[string]string{"email": "test@test.local", "password": "testpass123"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Error("expected token in response")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	body, _ := json.Marshal(map[string]string{"email": "test@test.local", "password": "wrongpassword"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUnauthorizedWithoutToken(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/devices", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}

// ======== Health ========

func TestHealthEndpoint(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ======== Device API Tests ========

func TestListDevicesEmpty(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/devices", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["data"] == nil {
		t.Error("expected data field in response")
	}
}

func TestGetDeviceNotFound(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/devices/NOSUCHSERIAL", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRebootDevice(t *testing.T) {
	r, deviceSvc, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	// Create a device first
	deviceSvc.Upsert(&models.Device{
		SerialNumber:     "TESTSERIAL001",
		DeviceID:         "DEVID001",
		ConnectionStatus: "online",
		Parameters:       models.JSONMap{},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/devices/TESTSERIAL001/reboot", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Error("expected success=true")
	}
}

func TestBulkImport(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	devices := []map[string]string{
		{"serial": "BULK001", "mac": "AA:BB:CC:DD:EE:01", "manufacturer": "Huawei", "model": "HG8245H", "isp": "Indihome"},
		{"serial": "BULK002", "mac": "AA:BB:CC:DD:EE:02", "manufacturer": "ZTE", "model": "F609", "isp": "Indihome"},
	}
	body, _ := json.Marshal(map[string]interface{}{"devices": devices})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/provisioning/bulk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["processed"].(float64) != 2 {
		t.Errorf("expected processed=2, got %v", resp["processed"])
	}
}

func TestGetISPTemplate(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/provisioning/template/indihome", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["isp"] != "Indihome" {
		t.Errorf("expected isp=Indihome, got %v", resp["isp"])
	}
}

func TestGetISPTemplateNotFound(t *testing.T) {
	r, _, _ := setupTestRouter(t)
	token := getAuthToken(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/provisioning/template/nonexistentisp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
