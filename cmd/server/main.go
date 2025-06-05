// Package main provides the main entry point for the Dear Future server
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/mocks"
	"github.com/thanhphuchuynh/dear-future/pkg/server"
)

func main() {
	// Load configuration (check for CONFIG_FILE environment variable)
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	configResult := config.LoadWithPath(configFile)
	if configResult.IsErr() {
		log.Fatalf("Failed to load configuration: %v", configResult.Error())
	}

	cfg := configResult.Value()

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application with dependencies
	appResult := initializeApplication(ctx, cfg)
	if appResult.IsErr() {
		log.Fatalf("Failed to initialize application: %v", appResult.Error())
	}

	app := appResult.Value()

	// Start the application
	startResult := app.Start(ctx)
	if startResult.IsErr() {
		log.Fatalf("Failed to start application: %v", startResult.Error())
	}

	// Create HTTP server
	serverResult := server.NewServer(cfg, app)
	if serverResult.IsErr() {
		log.Fatalf("Failed to create server: %v", serverResult.Error())
	}

	httpServer := serverResult.Value()

	// Setup graceful shutdown
	setupGracefulShutdown(ctx, cancel, httpServer, app)

	// Start the server
	log.Printf("Starting Dear Future server on %s", cfg.GetServerAddress())
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Platform: %s", cfg.PlatformName)

	if cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		log.Printf("Starting HTTPS server...")
		if err := httpServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		log.Printf("Starting HTTP server...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}

	log.Println("Server stopped")
}

// initializeApplication creates and configures the application
func initializeApplication(ctx context.Context, cfg *config.Config) common.Result[*composition.App] {
	// For now, we'll create a minimal app configuration
	// In a real implementation, this would wire up actual services
	appConfig := composition.AppConfig{
		Config: cfg,
		// Database: createDatabaseService(cfg),
		// Auth: createAuthService(cfg),
		// Email: createEmailService(cfg),
		// Storage: createStorageService(cfg),
	}

	// Create mock services for development
	if cfg.IsDevelopment() {
		return createDevelopmentApp(ctx, appConfig)
	}

	// Create production services
	return createProductionApp(ctx, appConfig)
}

// createDevelopmentApp creates an app with mock services for development
func createDevelopmentApp(ctx context.Context, appConfig composition.AppConfig) common.Result[*composition.App] {
	// For development, we'll use mock implementations
	log.Println("Initializing development environment with mock services...")

	appConfig.Database = mocks.NewMockDatabase()
	appConfig.Auth = mocks.NewMockAuthService()
	appConfig.Email = mocks.NewMockEmailService()
	appConfig.Storage = mocks.NewMockStorageService()

	return composition.NewApp(ctx, appConfig)
}

// createProductionApp creates an app with real services for production
func createProductionApp(ctx context.Context, appConfig composition.AppConfig) common.Result[*composition.App] {
	log.Println("Initializing production environment...")

	// TODO: Implement real services
	// appConfig.Database = supabase.NewDatabase(cfg.GetDatabaseConfig())
	// appConfig.Auth = supabase.NewAuthService(cfg.GetDatabaseConfig())
	// appConfig.Email = ses.NewEmailService(cfg.GetAWSConfig())
	// appConfig.Storage = s3.NewStorageService(cfg.GetAWSConfig())

	// For now, use mock services
	appConfig.Database = mocks.NewMockDatabase()
	appConfig.Auth = mocks.NewMockAuthService()
	appConfig.Email = mocks.NewMockEmailService()
	appConfig.Storage = mocks.NewMockStorageService()

	return composition.NewApp(ctx, appConfig)
}

// setupGracefulShutdown handles graceful shutdown of the server
func setupGracefulShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server, app *composition.App) {
	// Create channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server gracefully...")

		// Cancel the application context
		cancel()

		// Give the server 30 seconds to finish handling requests
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Shutdown the HTTP server
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}

		// Stop the application
		stopResult := app.Stop(shutdownCtx)
		if stopResult.IsErr() {
			log.Printf("Error stopping application: %v", stopResult.Error())
		}

		log.Println("Server shutdown complete")
	}()
}
