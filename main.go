// Package main provides the main entry point for the Dear Future application
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/thanhphuchuynh/dear-future/pkg/adapters/database"
	"github.com/thanhphuchuynh/dear-future/pkg/adapters/email"
	"github.com/thanhphuchuynh/dear-future/pkg/adapters/storage"
	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/mocks"
	"github.com/thanhphuchuynh/dear-future/pkg/server"
	"github.com/thanhphuchuynh/dear-future/pkg/services/scheduler"
)

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

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
	log.Printf("🚀 Starting Dear Future server on %s", cfg.GetServerAddress())
	log.Printf("📊 Environment: %s", cfg.Environment)
	log.Printf("🏗️  Platform: %s", cfg.PlatformName)

	if cfg.IsDevelopment() {
		log.Printf("🔗 Homepage: http://localhost%s", cfg.GetServerAddress())
		log.Printf("❤️  Health Check: http://localhost%s/health", cfg.GetServerAddress())
		log.Printf("🛠️  API: http://localhost%s/api/v1/", cfg.GetServerAddress())
	}

	if cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		log.Printf("🔒 Starting HTTPS server...")
		if err := httpServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		log.Printf("🌐 Starting HTTP server...")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}

	log.Println("✅ Server stopped")
}

// initializeApplication creates and configures the application
func initializeApplication(ctx context.Context, cfg *config.Config) common.Result[*composition.App] {
	appConfig := composition.AppConfig{
		Config: cfg,
	}

	// Create mock services for development
	if cfg.IsDevelopment() {
		return createDevelopmentApp(ctx, appConfig)
	}

	// Create production services
	return createProductionApp(ctx, appConfig)
}

// createDevelopmentApp creates an app with real services for development
func createDevelopmentApp(ctx context.Context, appConfig composition.AppConfig) common.Result[*composition.App] {
	log.Println("🧪 Initializing development environment...")

	cfg := appConfig.Config

	// Initialize Database (PostgreSQL) - same as production
	if cfg.Database.URL != "" {
		log.Println("📊 Connecting to PostgreSQL database...")
		dbConfig := database.PostgresConfig{
			DatabaseURL:  cfg.Database.URL,
			MaxConns:     cfg.Database.MaxConns,
			MaxIdleConns: cfg.Database.MaxIdleConns,
			ConnLifetime: cfg.DatabaseConnLifetime,
		}

		db, err := database.NewSimplePostgresDB(dbConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize database: %v", err)
			log.Println("⚠️  Falling back to mock database")
			appConfig.Database = mocks.NewMockDatabase()
		} else {
			log.Println("✅ PostgreSQL database connected")
			appConfig.Database = db
		}
	} else {
		log.Println("⚠️  No database URL configured, using mock database")
		appConfig.Database = mocks.NewMockDatabase()
	}

	// Initialize Email Service (SMTP) - optional in dev
	if cfg.SMTP.Host != "" && cfg.SMTP.Username != "" {
		log.Printf("📧 Configuring SMTP email service (%s)...", cfg.SMTP.Host)
		emailConfig := email.SMTPConfig{
			Host:          cfg.SMTP.Host,
			Port:          cfg.SMTP.Port,
			Username:      cfg.SMTP.Username,
			Password:      cfg.SMTP.Password,
			FromEmail:     cfg.SMTP.FromEmail,
			FromName:      cfg.SMTP.FromName,
			UseTLS:        cfg.SMTP.UseTLS,
			SkipTLSVerify: cfg.SMTP.SkipTLSVerify,
		}

		emailService, err := email.NewSMTPEmailService(emailConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize SMTP email service: %v", err)
			log.Println("⚠️  Using mock email service")
			appConfig.Email = mocks.NewMockEmailService()
		} else {
			log.Println("✅ SMTP email service configured")
			appConfig.Email = emailService
		}
	} else {
		log.Println("⚠️  No SMTP configuration, using mock email service")
		appConfig.Email = mocks.NewMockEmailService()
	}

	// Initialize Storage Service (R2) - optional in dev
	if cfg.R2Storage.AccountID != "" && cfg.R2Storage.BucketName != "" {
		log.Println("☁️  Configuring Cloudflare R2 storage...")
		storageConfig := storage.R2StorageConfig{
			AccountID:       cfg.R2Storage.AccountID,
			AccessKeyID:     cfg.R2Storage.AccessKeyID,
			SecretAccessKey: cfg.R2Storage.SecretAccessKey,
			BucketName:      cfg.R2Storage.BucketName,
			PublicURL:       cfg.R2Storage.PublicURL,
		}

		r2Storage, err := storage.NewR2Storage(storageConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize R2 storage: %v", err)
			log.Println("⚠️  Using mock storage service")
			appConfig.Storage = mocks.NewMockStorageService()
		} else {
			log.Println("✅ R2 storage configured")
			appConfig.Storage = r2Storage
		}
	} else {
		log.Println("⚠️  No R2 configuration, using mock storage service")
		appConfig.Storage = mocks.NewMockStorageService()
	}

	// JWT authentication (active in handlers)
	log.Println("🔐 JWT authentication active (password + token-based)")
	appConfig.Auth = mocks.NewMockAuthService()

	// Initialize scheduling service with River Queue
	if cfg.Database.URL != "" {
		log.Println("📋 Initializing River Queue scheduler...")
		pool, err := pgxpool.New(ctx, cfg.Database.URL)
		if err != nil {
			log.Printf("❌ Failed to create pgxpool for River: %v", err)
			log.Println("⚠️  Falling back to simple scheduler")
			appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
		} else {
			riverScheduler, err := scheduler.NewRiverScheduler(pool, appConfig.Database, appConfig.Email, cfg)
			if err != nil {
				log.Printf("❌ Failed to initialize River scheduler: %v", err)
				log.Println("⚠️  Falling back to simple scheduler")
				appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
			} else {
				log.Println("✅ River Queue scheduler initialized")
				appConfig.Scheduling = riverScheduler
			}
		}
	} else {
		log.Println("⚠️  No database URL, using simple scheduler")
		appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
	}

	return composition.NewApp(ctx, appConfig)
}

// createProductionApp creates an app with real services for production
func createProductionApp(ctx context.Context, appConfig composition.AppConfig) common.Result[*composition.App] {
	log.Println("🏭 Initializing production environment...")

	cfg := appConfig.Config

	// Initialize Database (PostgreSQL)
	if cfg.Database.URL != "" {
		log.Println("📊 Connecting to PostgreSQL database...")
		dbConfig := database.PostgresConfig{
			DatabaseURL:  cfg.Database.URL,
			MaxConns:     cfg.Database.MaxConns,
			MaxIdleConns: cfg.Database.MaxIdleConns,
			ConnLifetime: cfg.DatabaseConnLifetime,
		}

		db, err := database.NewSimplePostgresDB(dbConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize database: %v", err)
			log.Println("⚠️  Falling back to mock database")
			appConfig.Database = mocks.NewMockDatabase()
		} else {
			log.Println("✅ PostgreSQL database connected")
			appConfig.Database = db
		}
	} else {
		log.Println("⚠️  No database URL configured, using mock database")
		appConfig.Database = mocks.NewMockDatabase()
	}

	// Initialize Email Service (SMTP)
	if cfg.SMTP.Host != "" && cfg.SMTP.Username != "" {
		log.Printf("📧 Configuring SMTP email service (%s)...", cfg.SMTP.Host)
		emailConfig := email.SMTPConfig{
			Host:          cfg.SMTP.Host,
			Port:          cfg.SMTP.Port,
			Username:      cfg.SMTP.Username,
			Password:      cfg.SMTP.Password,
			FromEmail:     cfg.SMTP.FromEmail,
			FromName:      cfg.SMTP.FromName,
			UseTLS:        cfg.SMTP.UseTLS,
			SkipTLSVerify: cfg.SMTP.SkipTLSVerify,
		}

		emailService, err := email.NewSMTPEmailService(emailConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize SMTP email service: %v", err)
			log.Println("⚠️  Falling back to mock email service")
			appConfig.Email = mocks.NewMockEmailService()
		} else {
			log.Println("✅ SMTP email service configured")
			appConfig.Email = emailService
		}
	} else {
		log.Println("⚠️  No SMTP configuration found, using mock email service")
		appConfig.Email = mocks.NewMockEmailService()
	}

	// Initialize Storage Service (R2)
	if cfg.R2Storage.AccountID != "" && cfg.R2Storage.BucketName != "" {
		log.Println("☁️  Configuring Cloudflare R2 storage...")
		storageConfig := storage.R2StorageConfig{
			AccountID:       cfg.R2Storage.AccountID,
			AccessKeyID:     cfg.R2Storage.AccessKeyID,
			SecretAccessKey: cfg.R2Storage.SecretAccessKey,
			BucketName:      cfg.R2Storage.BucketName,
			PublicURL:       cfg.R2Storage.PublicURL,
		}

		r2Storage, err := storage.NewR2Storage(storageConfig)
		if err != nil {
			log.Printf("❌ Failed to initialize R2 storage: %v", err)
			log.Println("⚠️  Falling back to mock storage service")
			appConfig.Storage = mocks.NewMockStorageService()
		} else {
			log.Println("✅ R2 storage configured")
			appConfig.Storage = r2Storage
		}
	} else {
		log.Println("⚠️  No R2 configuration found, using mock storage service")
		appConfig.Storage = mocks.NewMockStorageService()
	}

	// Auth service (legacy abstraction - JWT auth is implemented in handlers)
	// Note: Real JWT authentication is active via handlers and middleware
	// This Auth service is for alternative providers (Supabase, OAuth, etc.)
	log.Println("🔐 JWT authentication active (password + token-based)")
	appConfig.Auth = mocks.NewMockAuthService()

	// Initialize scheduling service with River Queue
	if cfg.Database.URL != "" {
		log.Println("📋 Initializing River Queue scheduler...")
		pool, err := pgxpool.New(ctx, cfg.Database.URL)
		if err != nil {
			log.Printf("❌ Failed to create pgxpool for River: %v", err)
			log.Println("⚠️  Falling back to simple scheduler")
			appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
		} else {
			riverScheduler, err := scheduler.NewRiverScheduler(pool, appConfig.Database, appConfig.Email, cfg)
			if err != nil {
				log.Printf("❌ Failed to initialize River scheduler: %v", err)
				log.Println("⚠️  Falling back to simple scheduler")
				appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
			} else {
				log.Println("✅ River Queue scheduler initialized")
				appConfig.Scheduling = riverScheduler
			}
		}
	} else {
		log.Println("⚠️  No database URL, using simple scheduler")
		appConfig.Scheduling = scheduler.NewSimpleScheduler(appConfig.Database, appConfig.Email, cfg)
	}

	return composition.NewApp(ctx, appConfig)
}

// setupGracefulShutdown handles graceful shutdown of the server
func setupGracefulShutdown(ctx context.Context, cancel context.CancelFunc, server *http.Server, app *composition.App) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("⏹️  Shutting down server gracefully...")

		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("❌ Server forced to shutdown: %v", err)
		}

		stopResult := app.Stop(shutdownCtx)
		if stopResult.IsErr() {
			log.Printf("❌ Error stopping application: %v", stopResult.Error())
		}

		log.Println("🛑 Server shutdown complete")
	}()
}
