package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// SystemHandler handles health/stats endpoints
type SystemHandler struct {
	devices *services.DeviceService
	hub     *services.WSHub
}

func NewSystemHandler(devices *services.DeviceService, hub *services.WSHub) *SystemHandler {
	return &SystemHandler{devices: devices, hub: hub}
}

// Health handles GET /api/health
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "1.0.0",
		"uptime":  time.Since(startTime).String(),
	})
}

// Stats handles GET /api/stats
func (h *SystemHandler) Stats(c *gin.Context) {
	stats, err := h.devices.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	stats["uptime"] = time.Since(startTime).String()
	stats["ws_clients"] = h.hub.ClientCount()
	stats["go_goroutines"] = runtime.NumGoroutine()
	stats["memory_mb"] = mem.Alloc / 1024 / 1024

	c.JSON(http.StatusOK, stats)
}
