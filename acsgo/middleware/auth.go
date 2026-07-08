package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DWISSNET/acsgo/config"
	"github.com/DWISSNET/acsgo/pkg"
	"github.com/gin-gonic/gin"
)

// AuthRequired verifies JWT tokens on protected routes.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		claims, err := pkg.ValidateToken(parts[1], config.App.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("email", claims.Email)
		c.Next()
	}
}

// CORS adds permissive CORS headers for development/internal use.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger is a minimal structured request logger.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()

		logFn := log.Printf
		if status >= 500 {
			logFn = log.Printf
		}
		logFn("%s %s %d %s", c.Request.Method, c.Request.URL.Path, status, latency)
	}
}
