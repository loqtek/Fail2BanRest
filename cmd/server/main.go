package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fail2rest/v2/internal/auth"
	"github.com/fail2rest/v2/internal/config"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/handlers"
	"github.com/fail2rest/v2/internal/middleware"
	"github.com/gin-gonic/gin"
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
	f2bClient := fail2ban.NewClient(cfg.Fail2ban.ClientPath, cfg.Fail2ban.UseSudo)

	// Test fail2ban connection at startup
	log.Println("Testing fail2ban connection...")
	if _, err := f2bClient.GetStatus(); err != nil {
		log.Printf("WARNING: Failed to connect to fail2ban: %v", err)
		log.Println("The server will start, but fail2ban operations may fail.")
		log.Println("See README.md for permission setup instructions.")
	} else {
		log.Println("Successfully connected to fail2ban")
	}

	tokenExpiry, err := cfg.GetTokenExpiry()
	if err != nil {
		log.Fatalf("Invalid token expiry: %v", err)
	}

	// Convert config users to auth format
	userMap := make(map[string]string)
	for _, user := range cfg.Auth.Users {
		if user.Username != "" && user.Password != "" {
			userMap[user.Username] = user.Password
		}
	}

	authConfig := auth.AuthConfig{
		APIKeys: cfg.Auth.APIKeys,
		Users:   userMap,
	}

	authService := auth.NewAuthService(cfg.Auth.JWTSecret, tokenExpiry, authConfig)

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

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.BodySizeLimit(1024 * 1024)) // 1MB limit
	router.Use(middleware.Timeout(30 * time.Second))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		// Test fail2ban connectivity
		status := "ok"
		if _, err := f2bClient.GetStatus(); err != nil {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":  status,
			"service": "fail2ban-rest",
			"time":    time.Now().Unix(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes with rate limiting
		api.POST("/auth/login", middleware.RateLimiter("10-M"), authHandler.Login)

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

	// Start server with graceful shutdown
	address := cfg.GetAddress()

	srv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if cfg.Server.TLS.Enabled {
			if _, err := os.Stat(cfg.Server.TLS.CertFile); os.IsNotExist(err) {
				log.Fatalf("TLS certificate file not found: %s", cfg.Server.TLS.CertFile)
			}
			if _, err := os.Stat(cfg.Server.TLS.KeyFile); os.IsNotExist(err) {
				log.Fatalf("TLS key file not found: %s", cfg.Server.TLS.KeyFile)
			}

			log.Printf("Starting server with TLS on %s", address)
			if err := srv.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		} else {
			log.Printf("Starting server on %s", address)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
