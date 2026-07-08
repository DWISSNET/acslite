// Package handlers provides Gin HTTP handlers for the ACSGO REST API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/queue"
)

// DeviceHandlers groups device-related HTTP handlers.
type DeviceHandlers struct {
	deviceRepo  *repository.DeviceRepository
	commandRepo *repository.CommandRepository
	queue       queue.CommandPublisher
	logger      *zap.Logger
}

// NewDeviceHandlers creates new DeviceHandlers.
func NewDeviceHandlers(
	deviceRepo *repository.DeviceRepository,
	commandRepo *repository.CommandRepository,
	queuePub queue.CommandPublisher,
	logger *zap.Logger,
) *DeviceHandlers {
	return &DeviceHandlers{
		deviceRepo:  deviceRepo,
		commandRepo: commandRepo,
		queue:       queuePub,
		logger:      logger,
	}
}

// ListDevices godoc
// @Summary     List devices
// @Description Returns a paginated list of devices with optional filters
// @Tags        devices
// @Produce     json
// @Param       manufacturer query string false "Filter by manufacturer"
// @Param       model        query string false "Filter by model name"
// @Param       isp          query string false "Filter by ISP"
// @Param       status       query string false "Filter by status (online/offline)"
// @Param       search       query string false "Full-text search on device_id/serial/customer"
// @Param       limit        query int    false "Page size (default 50)"
// @Param       offset       query int    false "Offset (default 0)"
// @Success     200 {object} map[string]interface{}
// @Router      /api/devices [get]
func (h *DeviceHandlers) ListDevices(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := repository.DeviceFilter{
		Manufacturer: c.Query("manufacturer"),
		ModelName:    c.Query("model"),
		ISP:          c.Query("isp"),
		Status:       c.Query("status"),
		Search:       c.Query("search"),
		Limit:        limit,
		Offset:       offset,
	}

	devices, total, err := h.deviceRepo.List(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("ListDevices failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   devices,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetDevice returns a single device by UUID.
func (h *DeviceHandlers) GetDevice(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}
	c.JSON(http.StatusOK, device)
}

// GetDeviceParameters returns all TR-069 parameters for a device.
func (h *DeviceHandlers) GetDeviceParameters(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	params, err := h.deviceRepo.GetParameters(c.Request.Context(), device.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get parameters"})
		return
	}
	c.JSON(http.StatusOK, params)
}

// SendCommand dispatches an RPC command to a device.
func (h *DeviceHandlers) SendCommand(c *gin.Context) {
	deviceID := c.Param("id")

	var req struct {
		Type       models.CommandType `json:"type" binding:"required"`
		Parameters map[string]string  `json:"parameters"`
		MaxRetries int                `json:"max_retries"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}

	// Look up device
	device, err := h.deviceRepo.GetByID(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	userID := ""
	if u, exists := c.Get("user_id"); exists {
		userID, _ = u.(string)
	}

	cmd := &models.Command{
		DeviceID:    device.DeviceID,
		CommandType: req.Type,
		Parameters:  req.Parameters,
		Status:      models.CommandPending,
		MaxRetries:  req.MaxRetries,
		CreatedBy:   userID,
	}

	if err := h.commandRepo.Create(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Create command failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	// Publish to RabbitMQ for async delivery
	_ = h.queue.Publish(c.Request.Context(), cmd)

	c.JSON(http.StatusCreated, cmd)
}

// BulkSendCommand sends the same command to multiple devices.
func (h *DeviceHandlers) BulkSendCommand(c *gin.Context) {
	var req struct {
		DeviceIDs  []string           `json:"device_ids" binding:"required"`
		Type       models.CommandType `json:"type" binding:"required"`
		Parameters map[string]string  `json:"parameters"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ""
	if u, exists := c.Get("user_id"); exists {
		userID, _ = u.(string)
	}

	success := 0
	for _, devID := range req.DeviceIDs {
		cmd := &models.Command{
			DeviceID:    devID,
			CommandType: req.Type,
			Parameters:  req.Parameters,
			Status:      models.CommandPending,
			MaxRetries:  3,
			CreatedBy:   userID,
		}
		if err := h.commandRepo.Create(c.Request.Context(), cmd); err == nil {
			_ = h.queue.Publish(c.Request.Context(), cmd)
			success++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"queued":     success,
		"total":      len(req.DeviceIDs),
		"command":    req.Type,
	})
}

// UpdateDevice updates mutable device fields (customer, location, tags).
func (h *DeviceHandlers) UpdateDevice(c *gin.Context) {
	id := c.Param("id")
	device, err := h.deviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	var req struct {
		ISP             string   `json:"isp"`
		Tags            []string `json:"tags"`
		CustomerName    string   `json:"customer_name"`
		CustomerPhone   string   `json:"customer_phone"`
		CustomerEmail   string   `json:"customer_email"`
		CustomerPackage string   `json:"customer_package"`
		City            string   `json:"city"`
		Province        string   `json:"province"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ISP != "" {
		device.ISP = req.ISP
	}
	if len(req.Tags) > 0 {
		device.Tags = req.Tags
	}
	if req.CustomerName != "" {
		device.CustomerName = req.CustomerName
	}
	device.CustomerPhone = req.CustomerPhone
	device.CustomerEmail = req.CustomerEmail
	device.CustomerPackage = req.CustomerPackage
	if req.City != "" {
		device.City = req.City
	}
	if req.Province != "" {
		device.Province = req.Province
	}

	if err := h.deviceRepo.Upsert(c.Request.Context(), device); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, device)
}

// ---- Auth Handlers ----

// AuthHandlers groups authentication HTTP handlers.
type AuthHandlers struct {
	jwtSecret string
	logger    *zap.Logger
}

// NewAuthHandlers creates new AuthHandlers.
func NewAuthHandlers(jwtSecret string, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{jwtSecret: jwtSecret, logger: logger}
}

// LoginRequest is the JSON body for /api/auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login authenticates a user and returns a JWT.
func (h *AuthHandlers) Login(c *gin.Context) {
	// Simplified — production should look up user from DB
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Demo: accept admin/admin123
	if req.Username != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte("$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4tbQBqY4Fy"),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := generateJWT(req.Username, "admin", h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "username": req.Username, "role": "admin"})
}
