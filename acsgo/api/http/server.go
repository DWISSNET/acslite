// Package http provides the Gin-based REST API server for ACSGO.
package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/api/http/handlers"
	"github.com/DWISSNET/acsgo/api/ws"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/queue"
)

// Server is the REST API + WebSocket + Dashboard HTTP server.
type Server struct {
	deviceHandlers *handlers.DeviceHandlers
	authHandlers   *handlers.AuthHandlers
	wsHub          *ws.Hub
	jwtSecret      string
	logger         *zap.Logger
	httpServer     *http.Server
}

// NewServer creates and configures the Gin HTTP server.
func NewServer(
	port int,
	deviceRepo *repository.DeviceRepository,
	commandRepo *repository.CommandRepository,
	queuePub queue.CommandPublisher,
	wsHub *ws.Hub,
	jwtSecret string,
	logger *zap.Logger,
	metricsEnabled bool,
) *Server {
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	deviceH := handlers.NewDeviceHandlers(deviceRepo, commandRepo, queuePub, logger)
	authH := handlers.NewAuthHandlers(jwtSecret, logger)

	s := &Server{
		deviceHandlers: deviceH,
		authHandlers:   authH,
		wsHub:          wsHub,
		jwtSecret:      jwtSecret,
		logger:         logger,
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))
	r.Use(corsMiddleware())

	// Static assets
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/dashboard", "./web/index.html")
	r.StaticFile("/login", "./web/login.html")
	r.Static("/assets", "./web/assets")

	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authH.Login)
	}

	// Health & metrics
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "OK") })
	r.GET("/readyz", func(c *gin.Context) { c.String(http.StatusOK, "OK") })
	if metricsEnabled {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		wsHub.HandleWS(c.Writer, c.Request)
	})

	// Protected API routes
	api := r.Group("/api", s.authMiddleware())
	{
		// Devices
		devices := api.Group("/devices")
		{
			devices.GET("", deviceH.ListDevices)
			devices.GET("/:id", deviceH.GetDevice)
			devices.PUT("/:id", deviceH.UpdateDevice)
			devices.GET("/:id/parameters", deviceH.GetDeviceParameters)
			devices.POST("/:id/command", deviceH.SendCommand)
			devices.POST("/bulk-command", deviceH.BulkSendCommand)
		}
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

// Start begins listening for HTTP connections.
func (s *Server) Start() error {
	s.logger.Info("API server starting", zap.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// authMiddleware validates the JWT bearer token.
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		claims, err := handlers.ValidateJWT(parts[1], s.jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("HTTP",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
