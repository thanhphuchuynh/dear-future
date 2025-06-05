// Package composition handles dependency injection and application setup
package composition

import (
	"context"
	"errors"
	"fmt"

	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// App represents the complete application with all dependencies
type App struct {
	config *config.Config

	// Services (side effects)
	database    effects.Database
	auth        effects.AuthService
	email       effects.EmailService
	storage     effects.StorageService
	scheduling  effects.SchedulingService
	cache       effects.CacheService
	maintenance effects.MaintenanceService

	// Application services (business logic coordinators)
	userService    *UserService
	messageService *MessageService
	authHandler    *AuthHandler
	fileHandler    *FileHandler
}

// UserService coordinates user-related operations
type UserService struct {
	db     effects.Database
	auth   effects.AuthService
	config *config.Config
}

// MessageService coordinates message-related operations
type MessageService struct {
	db         effects.Database
	email      effects.EmailService
	storage    effects.StorageService
	scheduling effects.SchedulingService
	config     *config.Config
}

// AuthHandler handles authentication operations
type AuthHandler struct {
	auth   effects.AuthService
	db     effects.Database
	config *config.Config
}

// FileHandler handles file operations
type FileHandler struct {
	storage effects.StorageService
	db      effects.Database
	config  *config.Config
}

// AppConfig contains configuration for app creation
type AppConfig struct {
	Config *config.Config

	// Optional service implementations
	Database    effects.Database
	Auth        effects.AuthService
	Email       effects.EmailService
	Storage     effects.StorageService
	Scheduling  effects.SchedulingService
	Cache       effects.CacheService
	Maintenance effects.MaintenanceService
}

// NewApp creates a new application instance with dependency injection
func NewApp(ctx context.Context, appConfig AppConfig) common.Result[*App] {
	// Validate required dependencies
	if appConfig.Config == nil {
		return common.Err[*App](errors.New("config is required"))
	}

	if appConfig.Database == nil {
		return common.Err[*App](errors.New("database is required"))
	}

	if appConfig.Auth == nil {
		return common.Err[*App](errors.New("auth service is required"))
	}

	// Create application services
	userService := &UserService{
		db:     appConfig.Database,
		auth:   appConfig.Auth,
		config: appConfig.Config,
	}

	messageService := &MessageService{
		db:         appConfig.Database,
		email:      appConfig.Email,
		storage:    appConfig.Storage,
		scheduling: appConfig.Scheduling,
		config:     appConfig.Config,
	}

	authHandler := &AuthHandler{
		auth:   appConfig.Auth,
		db:     appConfig.Database,
		config: appConfig.Config,
	}

	fileHandler := &FileHandler{
		storage: appConfig.Storage,
		db:      appConfig.Database,
		config:  appConfig.Config,
	}

	app := &App{
		config:         appConfig.Config,
		database:       appConfig.Database,
		auth:           appConfig.Auth,
		email:          appConfig.Email,
		storage:        appConfig.Storage,
		scheduling:     appConfig.Scheduling,
		cache:          appConfig.Cache,
		maintenance:    appConfig.Maintenance,
		userService:    userService,
		messageService: messageService,
		authHandler:    authHandler,
		fileHandler:    fileHandler,
	}

	// Initialize the application
	return initializeApp(ctx, app)
}

// initializeApp performs application initialization
func initializeApp(ctx context.Context, app *App) common.Result[*App] {
	// Health check all services
	healthChecks := []common.Result[bool]{
		app.database.Ping(ctx),
	}

	// Check optional services
	if app.email != nil {
		healthChecks = append(healthChecks, app.email.ValidateEmailConfiguration(ctx))
	}

	// Validate all health checks
	for i, check := range healthChecks {
		if check.IsErr() {
			errorMsg := fmt.Sprintf("service %d health check failed: %v", i, check.Error())
			return common.Err[*App](errors.New(errorMsg))
		}
	}

	return common.Ok(app)
}

// Getters for app components

func (a *App) Config() *config.Config {
	return a.config
}

func (a *App) Database() effects.Database {
	return a.database
}

func (a *App) UserService() *UserService {
	return a.userService
}

func (a *App) MessageService() *MessageService {
	return a.messageService
}

func (a *App) AuthHandler() *AuthHandler {
	return a.authHandler
}

func (a *App) FileHandler() *FileHandler {
	return a.fileHandler
}

// Service getters

func (us *UserService) Database() effects.Database {
	return us.db
}

func (us *UserService) Auth() effects.AuthService {
	return us.auth
}

func (us *UserService) Config() *config.Config {
	return us.config
}

func (ms *MessageService) Database() effects.Database {
	return ms.db
}

func (ms *MessageService) Email() effects.EmailService {
	return ms.email
}

func (ms *MessageService) Storage() effects.StorageService {
	return ms.storage
}

func (ms *MessageService) Scheduling() effects.SchedulingService {
	return ms.scheduling
}

func (ms *MessageService) Config() *config.Config {
	return ms.config
}

func (ah *AuthHandler) Auth() effects.AuthService {
	return ah.auth
}

func (ah *AuthHandler) Database() effects.Database {
	return ah.db
}

func (ah *AuthHandler) Config() *config.Config {
	return ah.config
}

func (fh *FileHandler) Storage() effects.StorageService {
	return fh.storage
}

func (fh *FileHandler) Database() effects.Database {
	return fh.db
}

func (fh *FileHandler) Config() *config.Config {
	return fh.config
}

// Lifecycle management

// Start starts the application and all its services
func (a *App) Start(ctx context.Context) common.Result[bool] {
	// Log application startup
	if a.config.Debug {
		// In a real implementation, you'd use a proper logger
		println("Starting Dear Future application...")
		println("Environment:", a.config.Environment)
		println("Platform:", a.config.PlatformName)
	}

	// Start background services if needed
	if a.scheduling != nil {
		// Start message scheduler
		// This would typically start a goroutine for scheduled processing
	}

	return common.Ok(true)
}

// Stop gracefully stops the application and all its services
func (a *App) Stop(ctx context.Context) common.Result[bool] {
	if a.config.Debug {
		println("Stopping Dear Future application...")
	}

	// Stop background services
	// Close database connections
	// Cleanup resources

	return common.Ok(true)
}

// Health returns the health status of the application
func (a *App) Health(ctx context.Context) common.Result[AppHealth] {
	health := AppHealth{
		Status:   "healthy",
		Services: make(map[string]ServiceHealth),
	}

	// Check database health
	dbHealth := a.database.Ping(ctx)
	health.Services["database"] = ServiceHealth{
		Status:  getHealthStatus(dbHealth),
		Message: getHealthMessage(dbHealth),
	}

	// Check other services
	if a.email != nil {
		emailHealth := a.email.ValidateEmailConfiguration(ctx)
		health.Services["email"] = ServiceHealth{
			Status:  getHealthStatus(emailHealth),
			Message: getHealthMessage(emailHealth),
		}
	}

	// Determine overall health
	for _, service := range health.Services {
		if service.Status != "healthy" {
			health.Status = "degraded"
			break
		}
	}

	return common.Ok(health)
}

// AppHealth represents the health status of the application
type AppHealth struct {
	Status   string                   `json:"status"`
	Services map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Helper functions

func getHealthStatus(result common.Result[bool]) string {
	if result.IsOk() && result.Value() {
		return "healthy"
	}
	return "unhealthy"
}

func getHealthMessage(result common.Result[bool]) string {
	if result.IsErr() {
		return result.Error().Error()
	}
	return ""
}

// Factory functions for creating apps with different configurations

// NewDevelopmentApp creates an app configured for development
func NewDevelopmentApp(ctx context.Context, config *config.Config) common.Result[*App] {
	// In development, we might use mock services or local implementations
	appConfig := AppConfig{
		Config: config,
		// Database: mockDatabase, // Would be implemented
		// Auth: mockAuth,         // Would be implemented
		// Email: mockEmail,       // Would be implemented
	}

	return NewApp(ctx, appConfig)
}

// NewProductionApp creates an app configured for production
func NewProductionApp(ctx context.Context, config *config.Config) common.Result[*App] {
	// In production, we use real implementations
	appConfig := AppConfig{
		Config: config,
		// Database: supabaseDatabase, // Would be implemented
		// Auth: supabaseAuth,         // Would be implemented
		// Email: sesEmail,            // Would be implemented
		// Storage: s3Storage,         // Would be implemented
	}

	return NewApp(ctx, appConfig)
}

// NewTestApp creates an app configured for testing
func NewTestApp(ctx context.Context) common.Result[*App] {
	// Load test configuration
	testConfig := &config.Config{
		Environment: "test",
		Debug:       true,
		LogLevel:    "debug",
		// Add other test-specific settings
	}

	appConfig := AppConfig{
		Config: testConfig,
		// Database: inMemoryDatabase, // Would be implemented
		// Auth: mockAuth,             // Would be implemented
		// Email: mockEmail,           // Would be implemented
	}

	return NewApp(ctx, appConfig)
}

// Functional composition helpers

// WithDatabase adds a database to the app configuration
func WithDatabase(db effects.Database) func(*AppConfig) {
	return func(config *AppConfig) {
		config.Database = db
	}
}

// WithAuth adds an auth service to the app configuration
func WithAuth(auth effects.AuthService) func(*AppConfig) {
	return func(config *AppConfig) {
		config.Auth = auth
	}
}

// WithEmail adds an email service to the app configuration
func WithEmail(email effects.EmailService) func(*AppConfig) {
	return func(config *AppConfig) {
		config.Email = email
	}
}

// WithStorage adds a storage service to the app configuration
func WithStorage(storage effects.StorageService) func(*AppConfig) {
	return func(config *AppConfig) {
		config.Storage = storage
	}
}

// NewAppWithOptions creates an app using functional options pattern
func NewAppWithOptions(ctx context.Context, config *config.Config, options ...func(*AppConfig)) common.Result[*App] {
	appConfig := AppConfig{
		Config: config,
	}

	// Apply all options
	for _, option := range options {
		option(&appConfig)
	}

	return NewApp(ctx, appConfig)
}
