package handlers

import (
	"net/http"

	"github.com/DWISSNET/acsgo/config"
	"github.com/DWISSNET/acsgo/pkg"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles login/logout
type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}

	cfg := config.App
	if req.Email != cfg.AdminEmail {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare with the configured password (bcrypt-aware)
	if err := bcrypt.CompareHashAndPassword([]byte(cfg.AdminPassword), []byte(req.Password)); err != nil {
		// Fallback: plain-text comparison (when password is stored in plain text in .env)
		if req.Password != cfg.AdminPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
	}

	token, err := pkg.GenerateToken(req.Email, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user":    gin.H{"email": req.Email},
	})
}

// Logout handles POST /api/auth/logout (stateless — client discards token)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logged out"})
}

// Me handles GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	email, _ := c.Get("email")
	c.JSON(http.StatusOK, gin.H{"email": email})
}
