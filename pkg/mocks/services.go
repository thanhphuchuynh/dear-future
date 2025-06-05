// Package mocks provides mock implementations for testing and development
package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// MockDatabase implements effects.Database interface for testing
type MockDatabase struct{}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

func (m *MockDatabase) Ping(ctx context.Context) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockDatabase) SaveUser(ctx context.Context, u user.User) common.Result[user.User] {
	return common.Ok(u)
}

func (m *MockDatabase) FindUserByID(ctx context.Context, userID uuid.UUID) common.Result[user.User] {
	return common.Err[user.User](NewError("user not found"))
}

func (m *MockDatabase) FindUserByEmail(ctx context.Context, email string) common.Result[user.User] {
	return common.Err[user.User](NewError("user not found"))
}

func (m *MockDatabase) UpdateUser(ctx context.Context, u user.User) common.Result[user.User] {
	return common.Ok(u)
}

func (m *MockDatabase) DeleteUser(ctx context.Context, userID uuid.UUID) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockDatabase) SaveUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile] {
	return common.Ok(profile)
}

func (m *MockDatabase) FindUserProfile(ctx context.Context, userID uuid.UUID) common.Result[user.UserProfile] {
	return common.Err[user.UserProfile](NewError("profile not found"))
}

func (m *MockDatabase) UpdateUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile] {
	return common.Ok(profile)
}

func (m *MockDatabase) SaveMessage(ctx context.Context, msg message.Message) common.Result[message.Message] {
	return common.Ok(msg)
}

func (m *MockDatabase) FindMessageByID(ctx context.Context, messageID uuid.UUID) common.Result[message.Message] {
	return common.Err[message.Message](NewError("message not found"))
}

func (m *MockDatabase) FindMessagesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) common.Result[[]message.Message] {
	return common.Ok([]message.Message{})
}

func (m *MockDatabase) FindMessagesByStatus(ctx context.Context, status message.MessageStatus, limit int) common.Result[[]message.Message] {
	return common.Ok([]message.Message{})
}

func (m *MockDatabase) FindDueMessages(ctx context.Context, before time.Time, limit int) common.Result[[]message.Message] {
	return common.Ok([]message.Message{})
}

func (m *MockDatabase) UpdateMessage(ctx context.Context, msg message.Message) common.Result[message.Message] {
	return common.Ok(msg)
}

func (m *MockDatabase) DeleteMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockDatabase) SaveMessageAttachment(ctx context.Context, attachment message.MessageAttachment) common.Result[message.MessageAttachment] {
	return common.Ok(attachment)
}

func (m *MockDatabase) FindAttachmentsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]message.MessageAttachment] {
	return common.Ok([]message.MessageAttachment{})
}

func (m *MockDatabase) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockDatabase) SaveDeliveryLog(ctx context.Context, log effects.DeliveryLog) common.Result[effects.DeliveryLog] {
	return common.Ok(log)
}

func (m *MockDatabase) FindDeliveryLogsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]effects.DeliveryLog] {
	return common.Ok([]effects.DeliveryLog{})
}

// MockAuthService implements effects.AuthService interface
type MockAuthService struct{}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{}
}

func (m *MockAuthService) CreateUser(ctx context.Context, email, password string) common.Result[effects.AuthResult] {
	result := effects.AuthResult{
		UserID:       uuid.New(),
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}
	return common.Ok(result)
}

func (m *MockAuthService) AuthenticateUser(ctx context.Context, email, password string) common.Result[effects.AuthResult] {
	result := effects.AuthResult{
		UserID:       uuid.New(),
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}
	return common.Ok(result)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) common.Result[effects.TokenValidationResult] {
	result := effects.TokenValidationResult{
		UserID:    uuid.New(),
		Valid:     true,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Scopes:    []string{"read", "write"},
	}
	return common.Ok(result)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) common.Result[effects.AuthResult] {
	result := effects.AuthResult{
		UserID:       uuid.New(),
		AccessToken:  "new-mock-access-token",
		RefreshToken: "new-mock-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}
	return common.Ok(result)
}

func (m *MockAuthService) RevokeToken(ctx context.Context, token string) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockAuthService) RequestPasswordReset(ctx context.Context, email string) common.Result[string] {
	return common.Ok("mock-reset-token")
}

func (m *MockAuthService) ResetPassword(ctx context.Context, resetToken, newPassword string) common.Result[bool] {
	return common.Ok(true)
}

// MockEmailService implements effects.EmailService interface
type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (m *MockEmailService) SendMessage(ctx context.Context, deliveryInfo message.MessageDeliveryInfo) common.Result[effects.EmailResult] {
	result := effects.EmailResult{
		MessageID: "mock-message-id",
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: deliveryInfo.GetRecipientEmail(),
		Subject:   deliveryInfo.GetSubject(),
	}
	return common.Ok(result)
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, email, verificationToken string) common.Result[effects.EmailResult] {
	result := effects.EmailResult{
		MessageID: "mock-verification-id",
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: email,
		Subject:   "Email Verification",
	}
	return common.Ok(result)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, email, resetToken string) common.Result[effects.EmailResult] {
	result := effects.EmailResult{
		MessageID: "mock-reset-id",
		Status:    effects.EmailStatusSent,
		SentAt:    time.Now(),
		Error:     common.None[string](),
		Recipient: email,
		Subject:   "Password Reset",
	}
	return common.Ok(result)
}

func (m *MockEmailService) ValidateEmailConfiguration(ctx context.Context) common.Result[bool] {
	return common.Ok(true)
}

// MockStorageService implements effects.StorageService interface
type MockStorageService struct{}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{}
}

func (m *MockStorageService) UploadFile(ctx context.Context, fileData effects.FileUpload) common.Result[effects.FileUploadResult] {
	result := effects.FileUploadResult{
		Key:         "mock-file-key-" + fileData.FileName,
		URL:         "https://mock-storage.com/" + fileData.FileName,
		Size:        fileData.Size,
		ContentType: fileData.ContentType,
		UploadedAt:  time.Now(),
		ETag:        "mock-etag",
	}
	return common.Ok(result)
}

func (m *MockStorageService) DownloadFile(ctx context.Context, key string) common.Result[effects.FileDownload] {
	result := effects.FileDownload{
		FileName:     "mock-file.txt",
		ContentType:  "text/plain",
		Data:         []byte("mock file content"),
		Size:         18,
		LastModified: time.Now(),
	}
	return common.Ok(result)
}

func (m *MockStorageService) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) common.Result[string] {
	url := "https://mock-presigned.com/" + key + "?expires=" + time.Now().Add(expiration).Format(time.RFC3339)
	return common.Ok(url)
}

func (m *MockStorageService) DeleteFile(ctx context.Context, key string) common.Result[bool] {
	return common.Ok(true)
}

func (m *MockStorageService) GetFileMetadata(ctx context.Context, key string) common.Result[effects.FileMetadata] {
	result := effects.FileMetadata{
		Key:          key,
		Size:         1024,
		ContentType:  "application/octet-stream",
		LastModified: time.Now(),
		ETag:         "mock-etag",
		Metadata:     map[string]string{"source": "mock"},
	}
	return common.Ok(result)
}

// NewError creates a simple error for mocking
func NewError(message string) error {
	return &mockError{message: message}
}

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}
