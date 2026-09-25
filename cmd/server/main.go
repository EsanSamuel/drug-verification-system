package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drug-verification/server/internal/auth"
	"github.com/drug-verification/server/internal/config"
	"github.com/drug-verification/server/internal/database"
	"github.com/drug-verification/server/internal/drugs"
	"github.com/drug-verification/server/internal/manufacturers"
	"github.com/drug-verification/server/internal/middleware"
	"github.com/drug-verification/server/internal/verification"
	jwtpkg "github.com/drug-verification/server/pkg/jwt"
	"github.com/drug-verification/server/pkg/qr"
)

func main() {
	// Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("configuration loaded", "env", cfg.AppEnv, "port", cfg.Server.Port)

	// Database connection
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("database connected")

	// Gin setup
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))

	// Request size limit (10 MB)
	router.MaxMultipartMemory = 10 << 20

	// CORS middleware
	router.Use(corsMiddleware(cfg.CORS.Origin))

	// Register routes
	registerRoutes(router, pool, cfg, logger)

	// HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down server", "signal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

// corsMiddleware returns a simple CORS middleware.
func corsMiddleware(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// registerRoutes sets up all API route groups.
func registerRoutes(router *gin.Engine, pool *pgxpool.Pool, cfg *config.Config, logger *slog.Logger) {
	// Initialize core utilities
	jwtManager := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.Expiration)
	qrGen := qr.NewGenerator(cfg.Verify.BaseURL)

	// Initialize repositories
	authRepo := auth.NewPostgresRepository(pool)
	mfgRepo := manufacturers.NewPostgresRepository(pool)
	drugsRepo := drugs.NewPostgresRepository(pool)
	verifRepo := verification.NewPostgresRepository(pool)

	// Initialize services
	authService := auth.NewService(authRepo, jwtManager, logger)
	mfgService := manufacturers.NewService(mfgRepo, logger)
	drugsService := drugs.NewService(drugsRepo, logger)
	verifService := verification.NewService(verifRepo, logger)

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	mfgHandler := manufacturers.NewHandler(mfgService)
	drugsHandler := drugs.NewHandler(drugsService, qrGen)
	verifHandler := verification.NewHandler(verifService)

	// API v1 route group
	api := router.Group("/api/v1")

	// Public routes
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler.RegisterRoutes(api)
	verifHandler.RegisterPublicRoutes(api)

	// Protected routes (require JWT authentication)
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	{
		authHandler.RegisterProtectedRoutes(protected)
		mfgHandler.RegisterRoutes(protected)
		drugsHandler.RegisterRoutes(protected)
		verifHandler.RegisterProtectedRoutes(protected)
	}
}
