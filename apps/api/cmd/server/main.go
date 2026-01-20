package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anthropics/pickle-go/apps/api/internal/config"
	"github.com/anthropics/pickle-go/apps/api/internal/handler"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "pickle-go-api",
			"version": "0.1.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/line/callback", handler.LineCallback)
			auth.POST("/refresh", middleware.AuthRequired(), handler.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(), handler.Logout)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.GET("/me", middleware.AuthRequired(), handler.GetCurrentUser)
			users.GET("/me/events", middleware.AuthRequired(), handler.GetMyEvents)
			users.GET("/me/registrations", middleware.AuthRequired(), handler.GetMyRegistrations)
			users.GET("/me/notifications", middleware.AuthRequired(), handler.GetMyNotifications)
		}

		// Event routes
		events := v1.Group("/events")
		{
			events.GET("", handler.ListEvents)
			events.GET("/:id", handler.GetEvent)
			events.POST("", middleware.AuthRequired(), handler.CreateEvent)
			events.PUT("/:id", middleware.AuthRequired(), handler.UpdateEvent)
			events.DELETE("/:id", middleware.AuthRequired(), handler.DeleteEvent)

			// Registration routes
			events.POST("/:id/register", middleware.AuthRequired(), handler.RegisterEvent)
			events.DELETE("/:id/register", middleware.AuthRequired(), handler.CancelRegistration)
			events.GET("/:id/registrations", handler.GetEventRegistrations)
		}
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
