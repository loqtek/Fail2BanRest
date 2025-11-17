package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/auth"
	"github.com/fail2rest/v2/internal/config"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/handlers"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize components
	f2bClient := fail2ban.NewClient(cfg.Fail2ban.ClientPath)
	
	tokenExpiry, err := cfg.GetTokenExpiry()
	if err != nil {
		log.Fatalf("Invalid token expiry: %v", err)
	}
	
	authService := auth.NewAuthService(cfg.Auth.JWTSecret, tokenExpiry)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	statusHandler := handlers.NewStatusHandler(f2bClient)
	jailHandler := handlers.NewJailHandler(f2bClient)
	ipHandler := handlers.NewIPHandler(f2bClient)
	statsHandler := handlers.NewStatsHandler(f2bClient)

	// Setup router
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "fail2rest-v2",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.POST("/auth/login", authHandler.Login)

		// Protected routes
		protected := api.Group("")
		protected.Use(authService.Middleware())
		{
			// Status
			protected.GET("/status", statusHandler.GetStatus)

			// Jails
			protected.GET("/jails", jailHandler.GetJails)
			protected.GET("/jails/:name", jailHandler.GetJail)
			protected.GET("/jails/:name/status", jailHandler.GetJailStatus)
			protected.POST("/jails/:name/start", jailHandler.StartJail)
			protected.POST("/jails/:name/stop", jailHandler.StopJail)
			protected.POST("/jails/:name/restart", jailHandler.RestartJail)
			protected.POST("/jails/:name/reload", jailHandler.ReloadJail)

			// IP Management
			protected.GET("/jails/:name/banned", ipHandler.GetBannedIPs)
			protected.POST("/jails/:name/ban", ipHandler.BanIP)
			protected.POST("/jails/:name/unban", ipHandler.UnbanIP)

			// Statistics
			protected.GET("/stats", statsHandler.GetStats)
			protected.GET("/jails/:name/stats", statsHandler.GetJailStats)
		}
	}

	// Start server
	address := cfg.GetAddress()
	
	if cfg.Server.TLS.Enabled {
		if _, err := os.Stat(cfg.Server.TLS.CertFile); os.IsNotExist(err) {
			log.Fatalf("TLS certificate file not found: %s", cfg.Server.TLS.CertFile)
		}
		if _, err := os.Stat(cfg.Server.TLS.KeyFile); os.IsNotExist(err) {
			log.Fatalf("TLS key file not found: %s", cfg.Server.TLS.KeyFile)
		}
		
		log.Printf("Starting server with TLS on %s", address)
		if err := router.RunTLS(address, cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	} else {
		log.Printf("Starting server on %s", address)
		if err := router.Run(address); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

