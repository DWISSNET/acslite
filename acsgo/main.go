package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DWISSNET/acsgo/config"
	"github.com/DWISSNET/acsgo/db"
	"github.com/DWISSNET/acsgo/handlers"
	"github.com/DWISSNET/acsgo/middleware"
	"github.com/DWISSNET/acsgo/seeds"
	"github.com/DWISSNET/acsgo/services"
	"github.com/gin-gonic/gin"
)

//go:embed web/dashboard.html
var dashboardHTML embed.FS

func main() {
	// -------------------------------------------------------------------------
	// Configuration
	// -------------------------------------------------------------------------
	cfg := config.Load()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[ACSGO] ")

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// -------------------------------------------------------------------------
	// Database
	// -------------------------------------------------------------------------
	database, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database, cfg.DBType); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database schema ready")

	// -------------------------------------------------------------------------
	// Services
	// -------------------------------------------------------------------------
	deviceSvc := services.NewDeviceService(database)
	paramSvc := services.NewParameterService(database)
	rpcSvc := services.NewRPCService(database)
	wsHub := services.NewWSHub()
	cwmpSvc := services.NewCWMPService(deviceSvc, rpcSvc, wsHub)

	// Seed parameter templates (idempotent)
	if err := seeds.SeedParameters(paramSvc); err != nil {
		log.Printf("⚠️  Seed warning: %v", err)
	}

	// -------------------------------------------------------------------------
	// Background: Mark offline devices every 5 minutes
	// -------------------------------------------------------------------------
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			n, err := deviceSvc.MarkOffline(15 * time.Minute)
			if err != nil {
				log.Printf("⚠️  MarkOffline error: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("📴 Marked %d devices offline", n)
				wsHub.Broadcast(services.WSMessage{Event: "device:offline", Data: map[string]interface{}{"count": n}})
			}
		}
	}()

	// -------------------------------------------------------------------------
	// CWMP server (port 7547)
	// -------------------------------------------------------------------------
	cwmpMux := http.NewServeMux()
	cwmpMux.HandleFunc("/cwmp", cwmpSvc.HandleRequest)
	cwmpMux.HandleFunc("/acs", cwmpSvc.HandleRequest)     // alternate path
	cwmpMux.HandleFunc("/tr069", cwmpSvc.HandleRequest)   // alternate path
	cwmpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			cwmpSvc.HandleRequest(w, r)
			return
		}
		fmt.Fprintln(w, "ACSGO CWMP/TR-069 Server")
	})

	cwmpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.CWMPPort),
		Handler:      cwmpMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🔌 CWMP/TR-069 server listening on 0.0.0.0:%d", cfg.CWMPPort)
		if err := cwmpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ CWMP server error: %v", err)
		}
	}()

	// -------------------------------------------------------------------------
	// REST API + Dashboard (port 7548)
	// -------------------------------------------------------------------------
	r := gin.New()
	r.Use(middleware.Logger(), middleware.CORS())

	// Serve dashboard
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})
	r.GET("/dashboard", func(c *gin.Context) {
		data, _ := dashboardHTML.ReadFile("web/dashboard.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	r.GET("/login", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})

	// WebSocket
	r.GET("/ws", func(c *gin.Context) {
		wsHub.ServeWS(c.Writer, c.Request)
	})

	// Handlers
	authH := handlers.NewAuthHandler()
	deviceH := handlers.NewDeviceHandler(deviceSvc, rpcSvc)
	paramH := handlers.NewParameterHandler(paramSvc)
	provH := handlers.NewProvisioningHandler(deviceSvc)
	sysH := handlers.NewSystemHandler(deviceSvc, wsHub)

	// Auth routes (public)
	api := r.Group("/api")
	{
		authG := api.Group("/auth")
		authG.POST("/login", authH.Login)
		authG.POST("/logout", authH.Logout)

		// Health is public
		api.GET("/health", sysH.Health)
	}

	// Protected API routes
	protected := api.Group("", middleware.AuthRequired())
	{
		protected.GET("/auth/me", authH.Me)
		protected.GET("/stats", sysH.Stats)

		// Devices
		devG := protected.Group("/devices")
		devG.GET("", deviceH.List)
		devG.GET("/:serial", deviceH.Get)
		devG.POST("/:serial/reboot", deviceH.Reboot)
		devG.POST("/:serial/factory-reset", deviceH.FactoryReset)
		devG.PUT("/:serial/params", deviceH.SetParams)
		devG.GET("/:serial/events", deviceH.GetEvents)
		devG.DELETE("/:serial", deviceH.Delete)

		// Parameters
		// GET /api/parameters            → list (supports ?category= and ?vendor= filters)
		// GET /api/parameters/by-path    → fetch by path (pass ?path=Device.WiFi.SSID.1.SSID)
		// GET /api/parameters/category/:cat
		// GET /api/parameters/vendor/:vendor
		paramG := protected.Group("/parameters")
		paramG.GET("", paramH.List)
		paramG.GET("/by-path", paramH.GetByQuery)
		paramG.GET("/category/:cat", paramH.ByCategory)
		paramG.GET("/vendor/:vendor", paramH.ByVendor)

		// Provisioning
		provG := protected.Group("/provisioning")
		provG.POST("/bulk", provH.BulkImport)
		provG.GET("/templates", provH.Templates)
		provG.GET("/template/:isp", provH.GetTemplate)
	}

	webServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.WebPort),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🌐 Web/API server starting on http://0.0.0.0:%d", cfg.WebPort)
	log.Printf("📊 Dashboard: http://0.0.0.0:%d/dashboard", cfg.WebPort)
	log.Printf("🔑 Admin login: %s (password set in .env)", cfg.AdminEmail)
	log.Printf("📡 CWMP: 0.0.0.0:%d", cfg.CWMPPort)

	if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Web server error: %v", err)
	}
}
