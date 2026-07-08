package handlers

import (
	"net/http"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/pkg"
	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

// ProvisioningHandler handles bulk provisioning and ISP template APIs
type ProvisioningHandler struct {
	devices *services.DeviceService
}

func NewProvisioningHandler(devices *services.DeviceService) *ProvisioningHandler {
	return &ProvisioningHandler{devices: devices}
}

// BulkImport handles POST /api/provisioning/bulk
func (h *ProvisioningHandler) BulkImport(c *gin.Context) {
	var req struct {
		Devices []struct {
			Serial       string `json:"serial"`
			MAC          string `json:"mac"`
			Manufacturer string `json:"manufacturer"`
			Model        string `json:"model"`
			ISP          string `json:"isp"`
			CustomerName string `json:"customer_name"`
			Phone        string `json:"phone"`
			Package      string `json:"package"`
			City         string `json:"city"`
			Province     string `json:"province"`
		} `json:"devices" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "devices array required"})
		return
	}

	results := make([]gin.H, 0, len(req.Devices))
	for _, dev := range req.Devices {
		d := &models.Device{
			ID:              pkg.NewUUID(),
			SerialNumber:    dev.Serial,
			DeviceID:        pkg.BuildDeviceID("000000", dev.Serial),
			Manufacturer:    dev.Manufacturer,
			ModelName:       dev.Model,
			MACAddress:      dev.MAC,
			ISP:             dev.ISP,
			CustomerName:    dev.CustomerName,
			CustomerPhone:   dev.Phone,
			CustomerPackage: dev.Package,
			LocationCity:    dev.City,
			LocationProvince: dev.Province,
			ConnectionStatus: "offline",
			Parameters:      models.JSONMap{},
		}
		if err := h.devices.Upsert(d); err != nil {
			results = append(results, gin.H{"serial": dev.Serial, "status": "error", "error": err.Error()})
		} else {
			results = append(results, gin.H{"serial": dev.Serial, "status": "queued"})
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "processed": len(results), "results": results})
}

// Templates handles GET /api/provisioning/templates
func (h *ProvisioningHandler) Templates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"templates": []string{"indihome", "firstmedia", "biznet", "mncplay", "xlhome", "telkomsel"},
	})
}

// GetTemplate handles GET /api/provisioning/template/:isp
func (h *ProvisioningHandler) GetTemplate(c *gin.Context) {
	isp := c.Param("isp")
	tpl, ok := ISPTemplates[isp]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found for this ISP"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

// ISPTemplates holds pre-configured ISP provisioning templates
var ISPTemplates = map[string]interface{}{
	"indihome": map[string]interface{}{
		"name": "Indihome Default Template",
		"isp":  "Indihome",
		"wan": map[string]interface{}{
			"pppoe": map[string]string{"username": "%SESSION%@indihome", "password": "%PASSWORD%"},
			"vlan":  35,
		},
		"iptv": map[string]interface{}{"vlan": 4000, "igmp": true, "multicast": true},
		"voice": map[string]interface{}{"sipServer": "sip.indihome.co.id", "port": 5060},
		"wifi":  map[string]interface{}{"ssidPrefix": "INDIHOME-", "band2g": true, "band5g": true},
	},
	"firstmedia": map[string]interface{}{
		"name": "First Media Template",
		"isp":  "First Media",
		"wan":  map[string]interface{}{"dhcp": true, "vlan": 100},
		"wifi": map[string]interface{}{"ssidPrefix": "FIRSTMEDIA-", "band2g": true, "band5g": true},
	},
	"biznet": map[string]interface{}{
		"name": "Biznet Home Template",
		"isp":  "Biznet",
		"wan":  map[string]interface{}{"pppoe": map[string]string{"username": "%USER%", "password": "%PASS%"}, "vlan": 100},
	},
	"mncplay": map[string]interface{}{
		"name": "MNC Play Template",
		"isp":  "MNC Play",
		"wan":  map[string]interface{}{"dhcp": true, "vlan": 200},
	},
	"xlhome": map[string]interface{}{
		"name":    "XL Home LTE Template",
		"isp":     "XL Axiata",
		"apn":     "xlhome",
		"pincode": "0000",
	},
	"telkomsel": map[string]interface{}{
		"name":    "Telkomsel Template",
		"isp":     "Telkomsel",
		"apn":     "internet",
		"pincode": "0000",
	},
}
