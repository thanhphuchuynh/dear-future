// Package config handles application configuration
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// Config holds all application configuration
type Config struct {
	// Environment
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
	LogLevel    string `yaml:"log_level"`

	// Server configuration
	Server ServerConfig `yaml:"server"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// Authentication
	Auth AuthConfig `yaml:"auth"`

	// AWS Configuration
	AWS AWSConfig `yaml:"aws"`

	// Supabase Configuration
	Supabase SupabaseConfig `yaml:"supabase"`

	// R2 Storage Configuration (Cloudflare R2 - S3 compatible)
	R2Storage R2StorageConfig `yaml:"r2_storage"`

	// SMTP Email Configuration (Gmail, etc.)
	SMTP SMTPConfig `yaml:"smtp"`

	// File upload limits
	FileUpload FileUploadConfig `yaml:"file_upload"`

	// Message limits
	Message MessageConfig `yaml:"message"`

	// Email configuration
	Email EmailConfig `yaml:"email"`

	// Scheduling configuration
	Scheduling SchedulingConfig `yaml:"scheduling"`

	// Cache configuration
	Cache CacheConfig `yaml:"cache"`

	// Security
	Security SecurityConfig `yaml:"security"`

	// Monitoring
	Monitoring MonitoringConfig `yaml:"monitoring"`

	// Feature flags
	Features FeatureFlags `yaml:"features"`

	// Deployment specific
	Deployment DeploymentConfig `yaml:"deployment"`

	// Runtime fields (not from YAML)
	IsLambda     bool   `yaml:"-"`
	IsContainer  bool   `yaml:"-"`
	PlatformName string `yaml:"-"`

	// Legacy compatibility fields for existing code
	Port                   string        `yaml:"-"`
	ReadTimeout            time.Duration `yaml:"-"`
	WriteTimeout           time.Duration `yaml:"-"`
	IdleTimeout            time.Duration `yaml:"-"`
	DatabaseURL            string        `yaml:"-"`
	DatabaseMaxConns       int           `yaml:"-"`
	DatabaseMaxIdleConns   int           `yaml:"-"`
	DatabaseConnLifetime   time.Duration `yaml:"-"`
	JWTSecret              string        `yaml:"-"`
	JWTExpirationTime      time.Duration `yaml:"-"`
	RefreshTokenLifetime   time.Duration `yaml:"-"`
	PasswordMinLength      int           `yaml:"-"`
	AWSRegion              string        `yaml:"-"`
	S3Bucket               string        `yaml:"-"`
	S3BucketRegion         string        `yaml:"-"`
	SESRegion              string        `yaml:"-"`
	SESFromEmail           string        `yaml:"-"`
	SESFromName            string        `yaml:"-"`
	SupabaseURL            string        `yaml:"-"`
	SupabaseAnonKey        string        `yaml:"-"`
	SupabaseServiceKey     string        `yaml:"-"`
	MaxFileSize            int64         `yaml:"-"`
	MaxAttachments         int           `yaml:"-"`
	AllowedFileTypes       []string      `yaml:"-"`
	MaxMessageLength       int           `yaml:"-"`
	MaxTitleLength         int           `yaml:"-"`
	MessageQuotaPerUser    int           `yaml:"-"`
	EmailTemplatesPath     string        `yaml:"-"`
	EmailRateLimit         int           `yaml:"-"`
	SchedulerInterval      time.Duration `yaml:"-"`
	MaxRetryAttempts       int           `yaml:"-"`
	RetryBackoffMultiplier float64       `yaml:"-"`
	CacheEnabled           bool          `yaml:"-"`
	CacheURL               string        `yaml:"-"`
	CacheTTL               time.Duration `yaml:"-"`
	CacheKeyPrefix         string        `yaml:"-"`
	CORSOrigins            []string      `yaml:"-"`
	TrustedProxies         []string      `yaml:"-"`
	RateLimitPerMinute     int           `yaml:"-"`
	EnableHTTPS            bool          `yaml:"-"`
	CertFile               string        `yaml:"-"`
	KeyFile                string        `yaml:"-"`
	MetricsEnabled         bool          `yaml:"-"`
	MetricsPort            string        `yaml:"-"`
	HealthCheckURL         string        `yaml:"-"`
}

// Nested configuration structures
type ServerConfig struct {
	Port         string `yaml:"port"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

type DatabaseConfig struct {
	URL          string `yaml:"url"`
	MaxConns     int    `yaml:"max_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	ConnLifetime string `yaml:"conn_lifetime"`
}

type AuthConfig struct {
	JWTSecret            string `yaml:"jwt_secret"`
	JWTExpiration        string `yaml:"jwt_expiration"`
	RefreshTokenLifetime string `yaml:"refresh_token_lifetime"`
	PasswordMinLength    int    `yaml:"password_min_length"`
}

type AWSConfig struct {
	Region       string `yaml:"region"`
	S3Bucket     string `yaml:"s3_bucket"`
	S3Region     string `yaml:"s3_bucket_region"`
	SESRegion    string `yaml:"ses_region"`
	SESFromEmail string `yaml:"ses_from_email"`
	SESFromName  string `yaml:"ses_from_name"`
}

type SupabaseConfig struct {
	URL        string `yaml:"url"`
	AnonKey    string `yaml:"anon_key"`
	ServiceKey string `yaml:"service_key"`
}

type R2StorageConfig struct {
	AccountID       string `yaml:"account_id"`        // Cloudflare account ID
	AccessKeyID     string `yaml:"access_key_id"`     // R2 API token ID
	SecretAccessKey string `yaml:"secret_access_key"` // R2 API token secret
	BucketName      string `yaml:"bucket_name"`       // R2 bucket name
	PublicURL       string `yaml:"public_url"`        // Optional public URL
}

type SMTPConfig struct {
	Host          string `yaml:"host"`            // SMTP server host (smtp.gmail.com)
	Port          string `yaml:"port"`            // SMTP port (587 or 465)
	Username      string `yaml:"username"`        // SMTP username (email)
	Password      string `yaml:"password"`        // SMTP password (app password)
	FromEmail     string `yaml:"from_email"`      // From email address
	FromName      string `yaml:"from_name"`       // From display name
	UseTLS        bool   `yaml:"use_tls"`         // Use TLS
	SkipTLSVerify bool   `yaml:"skip_tls_verify"` // Skip TLS verify (dev only)
}

type FileUploadConfig struct {
	MaxFileSize      int64    `yaml:"max_file_size"`
	MaxAttachments   int      `yaml:"max_attachments"`
	AllowedFileTypes []string `yaml:"allowed_file_types"`
}

type MessageConfig struct {
	MaxMessageLength int `yaml:"max_message_length"`
	MaxTitleLength   int `yaml:"max_title_length"`
	QuotaPerUser     int `yaml:"quota_per_user"`
}

type EmailConfig struct {
	TemplatesPath string `yaml:"templates_path"`
	RateLimit     int    `yaml:"rate_limit"`
}

type SchedulingConfig struct {
	Interval               string           `yaml:"interval"`
	MaxRetryAttempts       int              `yaml:"max_retry_attempts"`
	RetryBackoffMultiplier float64          `yaml:"retry_backoff_multiplier"`
	River                  RiverQueueConfig `yaml:"river"`
}

type RiverQueueConfig struct {
	Enabled      bool   `yaml:"enabled"`       // Enable River Queue (vs simple scheduler)
	MaxWorkers   int    `yaml:"max_workers"`   // Max concurrent workers
	QueueName    string `yaml:"queue_name"`    // Queue name (default: "default")
	MaxAttempts  int    `yaml:"max_attempts"`  // Max delivery attempts per job
	PollInterval string `yaml:"poll_interval"` // How often to poll for jobs
}

type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	URL       string `yaml:"url"`
	TTL       string `yaml:"ttl"`
	KeyPrefix string `yaml:"key_prefix"`
}

type SecurityConfig struct {
	CORSOrigins        []string `yaml:"cors_origins"`
	TrustedProxies     []string `yaml:"trusted_proxies"`
	RateLimitPerMinute int      `yaml:"rate_limit_per_minute"`
	EnableHTTPS        bool     `yaml:"enable_https"`
	CertFile           string   `yaml:"cert_file"`
	KeyFile            string   `yaml:"key_file"`
}

type MonitoringConfig struct {
	MetricsEnabled bool   `yaml:"metrics_enabled"`
	MetricsPort    string `yaml:"metrics_port"`
	HealthCheckURL string `yaml:"health_check_url"`
}

type FeatureFlags struct {
	EnablePushNotifications  bool `yaml:"push_notifications"`
	EnableWebhooks           bool `yaml:"webhooks"`
	EnableMessageTemplates   bool `yaml:"message_templates"`
	EnableBatchProcessing    bool `yaml:"batch_processing"`
	EnableAdvancedScheduling bool `yaml:"advanced_scheduling"`
	EnableFileAttachments    bool `yaml:"file_attachments"`
	EnableEmailReminders     bool `yaml:"email_reminders"`
	EnableAnalytics          bool `yaml:"analytics"`
}

type DeploymentConfig struct {
	PlatformName string `yaml:"platform_name"`
	Container    bool   `yaml:"container"`
}

// LoadConfig loads configuration from YAML file and environment variables
func Load() common.Result[*Config] {
	return LoadWithPath("config.yaml")
}

// LoadWithPath loads configuration from a specific YAML file path
func LoadWithPath(configPath string) common.Result[*Config] {
	// First, load from YAML file
	config, err := loadFromYAML(configPath)
	if err != nil {
		// If YAML file doesn't exist, start with defaults
		config = getDefaultConfig()
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(config)

	// Map to legacy fields for backward compatibility
	mapToLegacyFields(config)

	// Detect deployment environment
	detectDeploymentEnvironment(config)

	// Validate configuration
	return validateConfig(config)
}

// loadFromYAML loads configuration from YAML file
func loadFromYAML(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return config, nil
}

// getDefaultConfig returns default configuration values
func getDefaultConfig() *Config {
	return &Config{
		Environment: "development",
		Debug:       true,
		LogLevel:    "info",
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  "15s",
			WriteTimeout: "15s",
			IdleTimeout:  "60s",
		},
		Database: DatabaseConfig{
			MaxConns:     25,
			MaxIdleConns: 5,
			ConnLifetime: "5m",
		},
		Auth: AuthConfig{
			JWTSecret:            "your-secret-key-change-in-production",
			JWTExpiration:        "15m",
			RefreshTokenLifetime: "168h",
			PasswordMinLength:    8,
		},
		AWS: AWSConfig{
			Region:       "us-east-1",
			S3Region:     "us-east-1",
			SESRegion:    "us-east-1",
			SESFromEmail: "noreply@dearfuture.app",
			SESFromName:  "Dear Future",
		},
		FileUpload: FileUploadConfig{
			MaxFileSize:    52428800, // 50MB
			MaxAttachments: 10,
			AllowedFileTypes: []string{
				"image/jpeg",
				"image/png",
				"application/pdf",
				"text/plain",
			},
		},
		Message: MessageConfig{
			MaxMessageLength: 10000,
			MaxTitleLength:   200,
			QuotaPerUser:     100,
		},
		Email: EmailConfig{
			TemplatesPath: "./templates",
			RateLimit:     100,
		},
		Scheduling: SchedulingConfig{
			Interval:               "1m",
			MaxRetryAttempts:       5,
			RetryBackoffMultiplier: 2.0,
			River: RiverQueueConfig{
				Enabled:      true,
				MaxWorkers:   10,
				QueueName:    "default",
				MaxAttempts:  5,
				PollInterval: "1s",
			},
		},
		Cache: CacheConfig{
			Enabled:   false,
			TTL:       "10m",
			KeyPrefix: "dear-future:",
		},
		Security: SecurityConfig{
			CORSOrigins: []string{
				"http://localhost:3000",
				"http://localhost:8080",
			},
			TrustedProxies:     []string{},
			RateLimitPerMinute: 60,
			EnableHTTPS:        false,
		},
		Monitoring: MonitoringConfig{
			MetricsEnabled: true,
			MetricsPort:    "9090",
			HealthCheckURL: "/health",
		},
		Features: FeatureFlags{
			EnableBatchProcessing: true,
			EnableFileAttachments: true,
		},
		Deployment: DeploymentConfig{
			PlatformName: "development",
			Container:    false,
		},
	}
}

// applyEnvironmentOverrides applies environment variable overrides
func applyEnvironmentOverrides(config *Config) {
	// Environment
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}
	if debug := os.Getenv("DEBUG"); debug != "" {
		config.Debug = getBoolFromEnv("DEBUG", config.Debug)
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Server
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}
	if timeout := os.Getenv("READ_TIMEOUT"); timeout != "" {
		config.Server.ReadTimeout = timeout
	}
	if timeout := os.Getenv("WRITE_TIMEOUT"); timeout != "" {
		config.Server.WriteTimeout = timeout
	}
	if timeout := os.Getenv("IDLE_TIMEOUT"); timeout != "" {
		config.Server.IdleTimeout = timeout
	}

	// Database
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if maxConns := os.Getenv("DATABASE_MAX_CONNS"); maxConns != "" {
		config.Database.MaxConns = getIntFromEnv("DATABASE_MAX_CONNS", config.Database.MaxConns)
	}

	// Auth
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Auth.JWTSecret = jwtSecret
	}
	if jwtExp := os.Getenv("JWT_EXPIRATION"); jwtExp != "" {
		config.Auth.JWTExpiration = jwtExp
	}

	// AWS
	if region := os.Getenv("AWS_REGION"); region != "" {
		config.AWS.Region = region
	}
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		config.AWS.S3Bucket = bucket
	}

	// Supabase
	if supabaseURL := os.Getenv("SUPABASE_URL"); supabaseURL != "" {
		config.Supabase.URL = supabaseURL
	}
	if anonKey := os.Getenv("SUPABASE_ANON_KEY"); anonKey != "" {
		config.Supabase.AnonKey = anonKey
	}
	if serviceKey := os.Getenv("SUPABASE_SERVICE_KEY"); serviceKey != "" {
		config.Supabase.ServiceKey = serviceKey
	}

	// R2 Storage (Cloudflare)
	if r2AccountID := os.Getenv("R2_ACCOUNT_ID"); r2AccountID != "" {
		config.R2Storage.AccountID = r2AccountID
	}
	if r2AccessKeyID := os.Getenv("R2_ACCESS_KEY_ID"); r2AccessKeyID != "" {
		config.R2Storage.AccessKeyID = r2AccessKeyID
	}
	if r2SecretKey := os.Getenv("R2_SECRET_ACCESS_KEY"); r2SecretKey != "" {
		config.R2Storage.SecretAccessKey = r2SecretKey
	}
	if r2Bucket := os.Getenv("R2_BUCKET_NAME"); r2Bucket != "" {
		config.R2Storage.BucketName = r2Bucket
	}
	if r2PublicURL := os.Getenv("R2_PUBLIC_URL"); r2PublicURL != "" {
		config.R2Storage.PublicURL = r2PublicURL
	}

	// SMTP Email
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.SMTP.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		config.SMTP.Port = smtpPort
	}
	if smtpUsername := os.Getenv("SMTP_USERNAME"); smtpUsername != "" {
		config.SMTP.Username = smtpUsername
	}
	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		config.SMTP.Password = smtpPassword
	}
	if smtpFromEmail := os.Getenv("SMTP_FROM_EMAIL"); smtpFromEmail != "" {
		config.SMTP.FromEmail = smtpFromEmail
	}
	if smtpFromName := os.Getenv("SMTP_FROM_NAME"); smtpFromName != "" {
		config.SMTP.FromName = smtpFromName
	}
	if smtpUseTLS := os.Getenv("SMTP_USE_TLS"); smtpUseTLS != "" {
		config.SMTP.UseTLS = getBoolFromEnv("SMTP_USE_TLS", true)
	}

	// File upload
	if maxSize := os.Getenv("MAX_FILE_SIZE"); maxSize != "" {
		config.FileUpload.MaxFileSize = getInt64FromEnv("MAX_FILE_SIZE", config.FileUpload.MaxFileSize)
	}
	if maxAttachments := os.Getenv("MAX_ATTACHMENTS"); maxAttachments != "" {
		config.FileUpload.MaxAttachments = getIntFromEnv("MAX_ATTACHMENTS", config.FileUpload.MaxAttachments)
	}

	// Cache
	if cacheURL := os.Getenv("CACHE_URL"); cacheURL != "" {
		config.Cache.URL = cacheURL
		config.Cache.Enabled = true
	}

	// Security
	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		config.Security.CORSOrigins = strings.Split(corsOrigins, ",")
	}

	// Feature flags
	if featureAttachments := os.Getenv("FEATURE_FILE_ATTACHMENTS"); featureAttachments != "" {
		config.Features.EnableFileAttachments = getBoolFromEnv("FEATURE_FILE_ATTACHMENTS", config.Features.EnableFileAttachments)
	}
}

// mapToLegacyFields maps new config structure to legacy fields for backward compatibility
func mapToLegacyFields(config *Config) {
	// Server
	config.Port = config.Server.Port
	config.ReadTimeout = parseDuration(config.Server.ReadTimeout, 15*time.Second)
	config.WriteTimeout = parseDuration(config.Server.WriteTimeout, 15*time.Second)
	config.IdleTimeout = parseDuration(config.Server.IdleTimeout, 60*time.Second)

	// Database
	config.DatabaseURL = config.Database.URL
	config.DatabaseMaxConns = config.Database.MaxConns
	config.DatabaseMaxIdleConns = config.Database.MaxIdleConns
	config.DatabaseConnLifetime = parseDuration(config.Database.ConnLifetime, 5*time.Minute)

	// Auth
	config.JWTSecret = config.Auth.JWTSecret
	config.JWTExpirationTime = parseDuration(config.Auth.JWTExpiration, 15*time.Minute)
	config.RefreshTokenLifetime = parseDuration(config.Auth.RefreshTokenLifetime, 7*24*time.Hour)
	config.PasswordMinLength = config.Auth.PasswordMinLength

	// AWS
	config.AWSRegion = config.AWS.Region
	config.S3Bucket = config.AWS.S3Bucket
	config.S3BucketRegion = config.AWS.S3Region
	config.SESRegion = config.AWS.SESRegion
	config.SESFromEmail = config.AWS.SESFromEmail
	config.SESFromName = config.AWS.SESFromName

	// Supabase
	config.SupabaseURL = config.Supabase.URL
	config.SupabaseAnonKey = config.Supabase.AnonKey
	config.SupabaseServiceKey = config.Supabase.ServiceKey

	// File upload
	config.MaxFileSize = config.FileUpload.MaxFileSize
	config.MaxAttachments = config.FileUpload.MaxAttachments
	config.AllowedFileTypes = config.FileUpload.AllowedFileTypes

	// Message
	config.MaxMessageLength = config.Message.MaxMessageLength
	config.MaxTitleLength = config.Message.MaxTitleLength
	config.MessageQuotaPerUser = config.Message.QuotaPerUser

	// Email
	config.EmailTemplatesPath = config.Email.TemplatesPath
	config.EmailRateLimit = config.Email.RateLimit

	// Scheduling
	config.SchedulerInterval = parseDuration(config.Scheduling.Interval, 1*time.Minute)
	config.MaxRetryAttempts = config.Scheduling.MaxRetryAttempts
	config.RetryBackoffMultiplier = config.Scheduling.RetryBackoffMultiplier

	// Cache
	config.CacheEnabled = config.Cache.Enabled
	config.CacheURL = config.Cache.URL
	config.CacheTTL = parseDuration(config.Cache.TTL, 10*time.Minute)
	config.CacheKeyPrefix = config.Cache.KeyPrefix

	// Security
	config.CORSOrigins = config.Security.CORSOrigins
	config.TrustedProxies = config.Security.TrustedProxies
	config.RateLimitPerMinute = config.Security.RateLimitPerMinute
	config.EnableHTTPS = config.Security.EnableHTTPS
	config.CertFile = config.Security.CertFile
	config.KeyFile = config.Security.KeyFile

	// Monitoring
	config.MetricsEnabled = config.Monitoring.MetricsEnabled
	config.MetricsPort = config.Monitoring.MetricsPort
	config.HealthCheckURL = config.Monitoring.HealthCheckURL

	// Deployment
	config.PlatformName = config.Deployment.PlatformName
	config.IsContainer = config.Deployment.Container
}

// detectDeploymentEnvironment detects the deployment environment
func detectDeploymentEnvironment(config *Config) {
	// Detect Lambda
	config.IsLambda = os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""

	// Detect container
	if os.Getenv("CONTAINER") != "" {
		config.IsContainer = getBoolFromEnv("CONTAINER", false)
	}

	// Override platform name if set
	if platformName := os.Getenv("PLATFORM_NAME"); platformName != "" {
		config.PlatformName = platformName
	}
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) common.Result[*Config] {
	// Check required fields
	if config.Environment == "" {
		return common.Err[*Config](errors.New("ENVIRONMENT is required"))
	}

	if config.DatabaseURL == "" && config.SupabaseURL == "" {
		return common.Err[*Config](errors.New("Either DATABASE_URL or SUPABASE_URL is required"))
	}

	if config.JWTSecret == "your-secret-key-change-in-production" && config.Environment == "production" {
		return common.Err[*Config](errors.New("JWT_SECRET must be changed in production"))
	}

	if config.S3Bucket == "" && config.Features.EnableFileAttachments {
		return common.Err[*Config](errors.New("S3_BUCKET is required when file attachments are enabled"))
	}

	// Validate numeric ranges
	if config.Port == "" {
		return common.Err[*Config](errors.New("PORT cannot be empty"))
	}

	if config.MaxFileSize <= 0 {
		return common.Err[*Config](errors.New("MAX_FILE_SIZE must be positive"))
	}

	if config.MaxAttachments <= 0 {
		return common.Err[*Config](errors.New("MAX_ATTACHMENTS must be positive"))
	}

	// Validate timeouts
	if config.ReadTimeout <= 0 {
		return common.Err[*Config](errors.New("read timeout must be positive"))
	}

	if config.WriteTimeout <= 0 {
		return common.Err[*Config](errors.New("write timeout must be positive"))
	}

	return common.Ok(config)
}

// Helper functions

func getBoolFromEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntFromEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getInt64FromEnv(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed
	}
	return defaultValue
}

// Configuration helpers

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.Environment == "staging" || c.Environment == "stage"
}

// GetServerAddress returns the server address to bind to
func (c *Config) GetServerAddress() string {
	if c.IsLambda {
		return "" // Lambda doesn't need server address
	}
	return ":" + c.Port
}

// GetDatabaseConfig returns database-specific configuration
func (c *Config) GetDatabaseConfig() DatabaseConfigInfo {
	return DatabaseConfigInfo{
		URL:          c.DatabaseURL,
		MaxConns:     c.DatabaseMaxConns,
		MaxIdleConns: c.DatabaseMaxIdleConns,
		ConnLifetime: c.DatabaseConnLifetime,
		SupabaseURL:  c.SupabaseURL,
		SupabaseKey:  c.SupabaseServiceKey,
	}
}

// GetAWSConfig returns AWS-specific configuration
func (c *Config) GetAWSConfig() AWSConfigInfo {
	return AWSConfigInfo{
		Region:       c.AWSRegion,
		S3Bucket:     c.S3Bucket,
		S3Region:     c.S3BucketRegion,
		SESRegion:    c.SESRegion,
		SESFromEmail: c.SESFromEmail,
		SESFromName:  c.SESFromName,
	}
}

// Configuration info structures for backward compatibility
type DatabaseConfigInfo struct {
	URL          string
	MaxConns     int
	MaxIdleConns int
	ConnLifetime time.Duration
	SupabaseURL  string
	SupabaseKey  string
}

type AWSConfigInfo struct {
	Region       string
	S3Bucket     string
	S3Region     string
	SESRegion    string
	SESFromEmail string
	SESFromName  string
}

// LoadConfigFromFile loads configuration from a specific file path
func LoadConfigFromFile(filePath string) common.Result[*Config] {
	// Make path absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return common.Err[*Config](fmt.Errorf("invalid config path: %w", err))
	}

	return LoadWithPath(absPath)
}

// SaveConfig saves the current configuration to a YAML file
func SaveConfig(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
