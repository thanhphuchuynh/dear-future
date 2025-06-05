// Package effects defines interfaces for side effects (I/O operations)
// This package isolates all side effects from pure business logic
package effects

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
)

// Database interface defines all database operations
type Database interface {
	// User operations
	SaveUser(ctx context.Context, user user.User) common.Result[user.User]
	FindUserByID(ctx context.Context, userID uuid.UUID) common.Result[user.User]
	FindUserByEmail(ctx context.Context, email string) common.Result[user.User]
	UpdateUser(ctx context.Context, user user.User) common.Result[user.User]
	DeleteUser(ctx context.Context, userID uuid.UUID) common.Result[bool]

	// User profile operations
	SaveUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile]
	FindUserProfile(ctx context.Context, userID uuid.UUID) common.Result[user.UserProfile]
	UpdateUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile]

	// Message operations
	SaveMessage(ctx context.Context, msg message.Message) common.Result[message.Message]
	FindMessageByID(ctx context.Context, messageID uuid.UUID) common.Result[message.Message]
	FindMessagesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) common.Result[[]message.Message]
	FindMessagesByStatus(ctx context.Context, status message.MessageStatus, limit int) common.Result[[]message.Message]
	FindDueMessages(ctx context.Context, before time.Time, limit int) common.Result[[]message.Message]
	UpdateMessage(ctx context.Context, msg message.Message) common.Result[message.Message]
	DeleteMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool]

	// Message attachment operations
	SaveMessageAttachment(ctx context.Context, attachment message.MessageAttachment) common.Result[message.MessageAttachment]
	FindAttachmentsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]message.MessageAttachment]
	DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) common.Result[bool]

	// Delivery log operations
	SaveDeliveryLog(ctx context.Context, log DeliveryLog) common.Result[DeliveryLog]
	FindDeliveryLogsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]DeliveryLog]

	// Health check
	Ping(ctx context.Context) common.Result[bool]
}

// EmailService interface defines email delivery operations
type EmailService interface {
	SendMessage(ctx context.Context, deliveryInfo message.MessageDeliveryInfo) common.Result[EmailResult]
	SendVerificationEmail(ctx context.Context, email, verificationToken string) common.Result[EmailResult]
	SendPasswordResetEmail(ctx context.Context, email, resetToken string) common.Result[EmailResult]
	ValidateEmailConfiguration(ctx context.Context) common.Result[bool]
}

// StorageService interface defines file storage operations
type StorageService interface {
	UploadFile(ctx context.Context, fileData FileUpload) common.Result[FileUploadResult]
	DownloadFile(ctx context.Context, key string) common.Result[FileDownload]
	GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) common.Result[string]
	DeleteFile(ctx context.Context, key string) common.Result[bool]
	GetFileMetadata(ctx context.Context, key string) common.Result[FileMetadata]
}

// AuthService interface defines authentication operations
type AuthService interface {
	CreateUser(ctx context.Context, email, password string) common.Result[AuthResult]
	AuthenticateUser(ctx context.Context, email, password string) common.Result[AuthResult]
	ValidateToken(ctx context.Context, token string) common.Result[TokenValidationResult]
	RefreshToken(ctx context.Context, refreshToken string) common.Result[AuthResult]
	RevokeToken(ctx context.Context, token string) common.Result[bool]
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) common.Result[bool]
	RequestPasswordReset(ctx context.Context, email string) common.Result[string] // Returns reset token
	ResetPassword(ctx context.Context, resetToken, newPassword string) common.Result[bool]
}

// SchedulingService interface defines message scheduling operations
type SchedulingService interface {
	ScheduleMessage(ctx context.Context, messageID uuid.UUID, deliveryTime time.Time) common.Result[ScheduleResult]
	CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool]
	RescheduleMessage(ctx context.Context, messageID uuid.UUID, newDeliveryTime time.Time) common.Result[ScheduleResult]
	GetScheduledMessages(ctx context.Context, from, to time.Time) common.Result[[]ScheduledMessage]
}

// NotificationService interface defines notification operations
type NotificationService interface {
	SendPushNotification(ctx context.Context, userID uuid.UUID, notification PushNotification) common.Result[NotificationResult]
	SendWebhookNotification(ctx context.Context, webhookURL string, payload interface{}) common.Result[NotificationResult]
	GetNotificationPreferences(ctx context.Context, userID uuid.UUID) common.Result[NotificationPreferences]
	UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs NotificationPreferences) common.Result[bool]
}

// CacheService interface defines caching operations
type CacheService interface {
	Get(ctx context.Context, key string) common.Result[[]byte]
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) common.Result[bool]
	Delete(ctx context.Context, key string) common.Result[bool]
	Exists(ctx context.Context, key string) common.Result[bool]
	FlushAll(ctx context.Context) common.Result[bool]
}

// Data structures for side effects

// DeliveryLog represents a log entry for message delivery attempts
type DeliveryLog struct {
	ID          uuid.UUID
	MessageID   uuid.UUID
	Status      message.MessageStatus
	ErrorMsg    common.Option[string]
	AttemptedAt time.Time
	Metadata    map[string]interface{}
}

// EmailResult represents the result of an email operation
type EmailResult struct {
	MessageID string
	Status    EmailStatus
	SentAt    time.Time
	Error     common.Option[string]
	Recipient string
	Subject   string
}

// EmailStatus represents the status of an email delivery
type EmailStatus string

const (
	EmailStatusSent     EmailStatus = "sent"
	EmailStatusFailed   EmailStatus = "failed"
	EmailStatusBounced  EmailStatus = "bounced"
	EmailStatusRejected EmailStatus = "rejected"
)

// FileUpload represents a file to be uploaded
type FileUpload struct {
	FileName    string
	ContentType string
	Data        []byte
	Size        int64
	Metadata    map[string]string
}

// FileUploadResult represents the result of a file upload
type FileUploadResult struct {
	Key         string
	URL         string
	Size        int64
	ContentType string
	UploadedAt  time.Time
	ETag        string
}

// FileDownload represents a downloaded file
type FileDownload struct {
	FileName     string
	ContentType  string
	Data         []byte
	Size         int64
	LastModified time.Time
}

// FileMetadata represents metadata about a file
type FileMetadata struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
	Metadata     map[string]string
}

// AuthResult represents the result of an authentication operation
type AuthResult struct {
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	TokenType    string
}

// TokenValidationResult represents the result of token validation
type TokenValidationResult struct {
	UserID    uuid.UUID
	Valid     bool
	ExpiresAt time.Time
	Scopes    []string
}

// ScheduleResult represents the result of scheduling a message
type ScheduleResult struct {
	MessageID    uuid.UUID
	ScheduledFor time.Time
	ScheduleID   string
	Status       ScheduleStatus
}

// ScheduleStatus represents the status of a scheduled message
type ScheduleStatus string

const (
	ScheduleStatusActive    ScheduleStatus = "active"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
	ScheduleStatusExecuted  ScheduleStatus = "executed"
)

// ScheduledMessage represents a scheduled message
type ScheduledMessage struct {
	MessageID    uuid.UUID
	ScheduledFor time.Time
	ScheduleID   string
	Status       ScheduleStatus
	CreatedAt    time.Time
}

// PushNotification represents a push notification
type PushNotification struct {
	Title string
	Body  string
	Data  map[string]string
	Sound common.Option[string]
	Badge common.Option[int]
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	ID        string
	Status    NotificationStatus
	SentAt    time.Time
	Error     common.Option[string]
	Recipient string
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusPending NotificationStatus = "pending"
)

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	EmailEnabled      bool
	PushEnabled       bool
	WebhookURL        common.Option[string]
	DeliveryReminders bool
	MarketingEmails   bool
	SecurityAlerts    bool
	WeeklyDigest      bool
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Service   string
	Healthy   bool
	Latency   time.Duration
	Error     common.Option[string]
	CheckedAt time.Time
}

// Batch operations for efficiency

// BatchOperation represents a batch operation
type BatchOperation[T any] struct {
	Operation string
	Data      T
	ID        string
}

// BatchResult represents the result of a batch operation
type BatchResult[T any] struct {
	ID      string
	Success bool
	Result  common.Option[T]
	Error   common.Option[string]
}

// DatabaseBatch interface for batch operations
type DatabaseBatch interface {
	SaveUsersBatch(ctx context.Context, users []user.User) common.Result[[]BatchResult[user.User]]
	SaveMessagesBatch(ctx context.Context, messages []message.Message) common.Result[[]BatchResult[message.Message]]
	UpdateMessagesBatch(ctx context.Context, messages []message.Message) common.Result[[]BatchResult[message.Message]]
}

// EmailBatch interface for batch email operations
type EmailBatch interface {
	SendMessagesBatch(ctx context.Context, deliveryInfos []message.MessageDeliveryInfo) common.Result[[]BatchResult[EmailResult]]
}

// Migration and maintenance interfaces

// DatabaseMigration interface for database migrations
type DatabaseMigration interface {
	GetVersion(ctx context.Context) common.Result[int]
	Migrate(ctx context.Context, targetVersion int) common.Result[bool]
	Rollback(ctx context.Context, targetVersion int) common.Result[bool]
	GetMigrationHistory(ctx context.Context) common.Result[[]MigrationRecord]
}

// MigrationRecord represents a migration record
type MigrationRecord struct {
	Version   int
	Name      string
	AppliedAt time.Time
	Checksum  string
}

// MaintenanceService interface for system maintenance
type MaintenanceService interface {
	BackupDatabase(ctx context.Context) common.Result[BackupResult]
	RestoreDatabase(ctx context.Context, backupID string) common.Result[bool]
	CleanupExpiredData(ctx context.Context, before time.Time) common.Result[CleanupResult]
	OptimizeDatabase(ctx context.Context) common.Result[bool]
	GetSystemStats(ctx context.Context) common.Result[SystemStats]
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackupID  string
	Size      int64
	CreatedAt time.Time
	Location  string
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	DeletedMessages    int
	DeletedAttachments int
	DeletedLogs        int
	FreedSpace         int64
}

// SystemStats represents system statistics
type SystemStats struct {
	TotalUsers        int
	TotalMessages     int
	ScheduledMessages int
	DeliveredMessages int
	FailedMessages    int
	TotalAttachments  int
	DatabaseSize      int64
	StorageUsed       int64
	UptimeSeconds     int64
	LastBackup        common.Option[time.Time]
}
