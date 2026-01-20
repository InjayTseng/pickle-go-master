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
	"github.com/anthropics/pickle-go/apps/api/internal/database"
	"github.com/anthropics/pickle-go/apps/api/internal/handler"
	"github.com/anthropics/pickle-go/apps/api/internal/middleware"
	"github.com/anthropics/pickle-go/apps/api/internal/repository"
	"github.com/anthropics/pickle-go/apps/api/pkg/line"
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

	// Initialize database connection
	dbCfg := database.DefaultConfig(cfg.DatabaseURL)
	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Initialize Line client
	lineClient := line.NewClient(line.Config{
		ChannelID:     cfg.LineChannelID,
		ChannelSecret: cfg.LineChannelSecret,
		RedirectURI:   cfg.LineRedirectURI,
	})

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, lineClient)
	userHandler := handler.NewUserHandler(userRepo, eventRepo, registrationRepo, notificationRepo)
	eventHandler := handler.NewEventHandler(eventRepo, userRepo, registrationRepo)
	registrationHandler := handler.NewRegistrationHandler(registrationRepo, eventRepo, notificationRepo)

	// Initialize router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Rate limiting (only in production)
	if cfg.Environment == "production" {
		router.Use(middleware.RateLimit())
	}

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
		// Auth routes - with strict rate limiting for security
		auth := v1.Group("/auth")
		if cfg.Environment == "production" {
			auth.Use(middleware.RateLimitStrict())
		}
		{
			auth.POST("/line/callback", authHandler.LineCallback)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(), authHandler.Logout)
		}

		// User routes
		users := v1.Group("/users")
		{
			users.GET("/me", middleware.AuthRequired(), authHandler.GetCurrentUser)
			users.GET("/me/events", middleware.AuthRequired(), userHandler.GetMyEvents)
			users.GET("/me/registrations", middleware.AuthRequired(), userHandler.GetMyRegistrations)
			users.GET("/me/notifications", middleware.AuthRequired(), userHandler.GetMyNotifications)
		}

		// Event routes
		events := v1.Group("/events")
		{
			events.GET("", eventHandler.ListEvents)
			events.GET("/by-code/:code", eventHandler.GetEventByCode)
			events.GET("/:id", eventHandler.GetEvent)
			events.POST("", middleware.AuthRequired(), eventHandler.CreateEvent)
			events.PUT("/:id", middleware.AuthRequired(), eventHandler.UpdateEvent)
			events.DELETE("/:id", middleware.AuthRequired(), eventHandler.DeleteEvent)

			// Registration routes
			events.POST("/:id/register", middleware.AuthRequired(), registrationHandler.RegisterEvent)
			events.DELETE("/:id/register", middleware.AuthRequired(), registrationHandler.CancelRegistration)
			events.GET("/:id/registrations", registrationHandler.GetEventRegistrations)
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
