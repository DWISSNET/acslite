package handlers

import (
	"net/http"
	"strconv"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

// DeviceHandler handles device REST API endpoints
type DeviceHandler struct {
	devices *services.DeviceService
	rpc     *services.RPCService
}

func NewDeviceHandler(devices *services.DeviceService, rpc *services.RPCService) *DeviceHandler {
	return &DeviceHandler{devices: devices, rpc: rpc}
}

// List handles GET /api/devices
func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	devices, total, err := h.devices.List(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}
	if devices == nil {
		devices = []*models.Device{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": devices,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
	})
}

// Get handles GET /api/devices/:serial
func (h *DeviceHandler) Get(c *gin.Context) {
	serial := c.Param("serial")
	device, err := h.devices.GetBySerial(serial)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"device": device})
}

// Reboot handles POST /api/devices/:serial/reboot
func (h *DeviceHandler) Reboot(c *gin.Context) {
	serial := c.Param("serial")
	_, err := h.devices.GetBySerial(serial)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	cmd, err := h.rpc.Enqueue(serial, "Reboot", models.JSONMap{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue reboot"})
		return
	}
	h.devices.LogEvent(serial, "reboot_requested", "Admin requested reboot via API")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Reboot command queued", "id": cmd.ID})
}

// FactoryReset handles POST /api/devices/:serial/factory-reset
func (h *DeviceHandler) FactoryReset(c *gin.Context) {
	serial := c.Param("serial")
	_, err := h.devices.GetBySerial(serial)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	cmd, err := h.rpc.Enqueue(serial, "FactoryReset", models.JSONMap{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue factory reset"})
		return
	}
	h.devices.LogEvent(serial, "factory_reset_requested", "Admin requested factory reset via API")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Factory reset command queued", "id": cmd.ID})
}

// SetParams handles PUT /api/devices/:serial/params
func (h *DeviceHandler) SetParams(c *gin.Context) {
	serial := c.Param("serial")
	var req struct {
		Parameters map[string]interface{} `json:"parameters" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameters map required"})
		return
	}

	_, err := h.devices.GetBySerial(serial)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Queue SetParameterValues RPC
	params := models.JSONMap{}
	for k, v := range req.Parameters {
		params[k] = v
	}
	cmd, err := h.rpc.Enqueue(serial, "SetParameterValues", params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue set parameters"})
		return
	}
	h.devices.LogEvent(serial, "set_params_requested", "Admin queued parameter update via API")
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "SetParameterValues queued", "id": cmd.ID, "count": len(params)})
}

// GetEvents handles GET /api/devices/:serial/events
func (h *DeviceHandler) GetEvents(c *gin.Context) {
	serial := c.Param("serial")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	events, err := h.devices.GetEvents(serial, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	if events == nil {
		events = []models.DeviceEvent{}
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

// Delete handles DELETE /api/devices/:serial
func (h *DeviceHandler) Delete(c *gin.Context) {
	serial := c.Param("serial")
	if err := h.devices.Delete(serial); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
