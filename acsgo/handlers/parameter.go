package handlers

import (
	"net/http"

	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

// ParameterHandler handles parameter template API endpoints
type ParameterHandler struct {
	params *services.ParameterService
}

func NewParameterHandler(params *services.ParameterService) *ParameterHandler {
	return &ParameterHandler{params: params}
}

// List handles GET /api/parameters
func (h *ParameterHandler) List(c *gin.Context) {
	category := c.Query("category")
	vendor := c.Query("vendor")

	params, err := h.params.List(category, vendor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parameters"})
		return
	}

	// Group by category
	grouped := make(map[string]interface{})
	for _, p := range params {
		if _, ok := grouped[p.Category]; !ok {
			grouped[p.Category] = []interface{}{}
		}
		grouped[p.Category] = append(grouped[p.Category].([]interface{}), p)
	}

	cats, _ := h.params.Categories()
	c.JSON(http.StatusOK, gin.H{
		"total":      len(params),
		"categories": cats,
		"parameters": grouped,
	})
}

// GetByQuery handles GET /api/parameters/by-path?path=Device.WiFi.SSID.1.SSID
func (h *ParameterHandler) GetByQuery(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter required"})
		return
	}
	param, err := h.params.GetByPath(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parameter not found"})
		return
	}
	c.JSON(http.StatusOK, param)
}

// Get handles GET /api/parameters/:path (kept for backwards compat — use /by-path in practice)
func (h *ParameterHandler) Get(c *gin.Context) {
	path := c.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	param, err := h.params.GetByPath(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parameter not found"})
		return
	}
	c.JSON(http.StatusOK, param)
}

// ByCategory handles GET /api/parameters/category/:cat
func (h *ParameterHandler) ByCategory(c *gin.Context) {
	cat := c.Param("cat")
	params, err := h.params.List(cat, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parameters"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat, "parameters": params, "total": len(params)})
}

// ByVendor handles GET /api/parameters/vendor/:vendor
func (h *ParameterHandler) ByVendor(c *gin.Context) {
	vendor := c.Param("vendor")
	params, err := h.params.List("", vendor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parameters"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vendor": vendor, "parameters": params, "total": len(params)})
}
