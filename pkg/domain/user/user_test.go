package user

import (
	"testing"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateUserRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user creation",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Name:     "John Doe",
				Timezone: "UTC",
			},
			expectError: false,
		},
		{
			name: "invalid email",
			request: CreateUserRequest{
				Email:    "invalid-email",
				Name:     "John Doe",
				Timezone: "UTC",
			},
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name: "empty name",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Name:     "",
				Timezone: "UTC",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "invalid timezone",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Name:     "John Doe",
				Timezone: "Invalid/Timezone",
			},
			expectError: true,
			errorMsg:    "invalid timezone",
		},
		{
			name: "default timezone when empty",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Name:     "John Doe",
				Timezone: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewUser(tt.request)

			if tt.expectError {
				if result.IsOk() {
					t.Errorf("expected error but got success")
				}
				if tt.errorMsg != "" && !contains(result.Error().Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s' but got '%s'", tt.errorMsg, result.Error().Error())
				}
			} else {
				if result.IsErr() {
					t.Errorf("expected success but got error: %v", result.Error())
				} else {
					user := result.Value()
					if user.Email() != normalizeEmail(tt.request.Email) {
						t.Errorf("expected email %s but got %s", tt.request.Email, user.Email())
					}
					if user.Name() != tt.request.Name {
						t.Errorf("expected name %s but got %s", tt.request.Name, user.Name())
					}
					if user.ID() == uuid.Nil {
						t.Errorf("expected valid UUID but got nil")
					}
				}
			}
		})
	}
}

func TestUserImmutability(t *testing.T) {
	// Create a user
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Timezone: "UTC",
	}

	userResult := NewUser(req)
	if userResult.IsErr() {
		t.Fatalf("failed to create user: %v", userResult.Error())
	}

	originalUser := userResult.Value()
	originalName := originalUser.Name()
	originalEmail := originalUser.Email()

	// Test immutability - updating name returns new instance
	newUserResult := originalUser.WithName("Jane Doe")
	if newUserResult.IsErr() {
		t.Fatalf("failed to update user name: %v", newUserResult.Error())
	}

	newUser := newUserResult.Value()

	// Original user should be unchanged
	if originalUser.Name() != originalName {
		t.Errorf("original user name was modified")
	}
	if originalUser.Email() != originalEmail {
		t.Errorf("original user email was modified")
	}

	// New user should have updated name but same ID and email
	if newUser.Name() != "Jane Doe" {
		t.Errorf("new user name was not updated")
	}
	if newUser.ID() != originalUser.ID() {
		t.Errorf("new user should have same ID")
	}
	if newUser.Email() != originalUser.Email() {
		t.Errorf("new user should have same email")
	}
	if newUser.UpdatedAt().Before(originalUser.UpdatedAt()) || newUser.UpdatedAt().Equal(originalUser.UpdatedAt()) {
		t.Errorf("new user should have later UpdatedAt timestamp")
	}
}

func TestUserProfileOperations(t *testing.T) {
	// Create a user
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Timezone: "UTC",
	}

	userResult := NewUser(req)
	if userResult.IsErr() {
		t.Fatalf("failed to create user: %v", userResult.Error())
	}

	user := userResult.Value()
	profile := NewUserProfile(user)

	// Test default values
	if !profile.EmailNotifications() {
		t.Errorf("email notifications should be enabled by default")
	}
	if profile.ProfilePictureURL().IsSome() {
		t.Errorf("profile picture URL should be None by default")
	}
	if profile.NotificationEmail().IsSome() {
		t.Errorf("notification email should be None by default")
	}

	// Test profile picture update
	pictureURL := "https://example.com/profile.jpg"
	updatedProfileResult := profile.WithProfilePicture(pictureURL)
	if updatedProfileResult.IsErr() {
		t.Fatalf("failed to update profile picture: %v", updatedProfileResult.Error())
	}

	updatedProfile := updatedProfileResult.Value()
	if updatedProfile.ProfilePictureURL().IsNone() {
		t.Errorf("profile picture URL should be set")
	}
	if updatedProfile.ProfilePictureURL().Value() != pictureURL {
		t.Errorf("profile picture URL should be %s but got %s", pictureURL, updatedProfile.ProfilePictureURL().Value())
	}

	// Test notification email update
	notificationEmail := "notifications@example.com"
	finalProfileResult := updatedProfile.WithNotificationEmail(notificationEmail)
	if finalProfileResult.IsErr() {
		t.Fatalf("failed to update notification email: %v", finalProfileResult.Error())
	}

	finalProfile := finalProfileResult.Value()
	if finalProfile.NotificationEmail().IsNone() {
		t.Errorf("notification email should be set")
	}
	if finalProfile.NotificationEmail().Value() != notificationEmail {
		t.Errorf("notification email should be %s but got %s", notificationEmail, finalProfile.NotificationEmail().Value())
	}

	// Test effective email
	effectiveEmail := finalProfile.GetEffectiveEmail()
	if effectiveEmail != notificationEmail {
		t.Errorf("effective email should be notification email %s but got %s", notificationEmail, effectiveEmail)
	}
}

func TestUserUpdateRequest(t *testing.T) {
	// Create a user
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Timezone: "UTC",
	}

	userResult := NewUser(req)
	if userResult.IsErr() {
		t.Fatalf("failed to create user: %v", userResult.Error())
	}

	originalUser := userResult.Value()

	// Test partial update (name only)
	updateReq := UpdateUserRequest{
		Name:     common.Some("Jane Smith"),
		Timezone: common.None[string](),
	}

	updatedUserResult := originalUser.UpdateUser(updateReq)
	if updatedUserResult.IsErr() {
		t.Fatalf("failed to update user: %v", updatedUserResult.Error())
	}

	updatedUser := updatedUserResult.Value()

	// Check that name was updated
	if updatedUser.Name() != "Jane Smith" {
		t.Errorf("expected name 'Jane Smith' but got '%s'", updatedUser.Name())
	}

	// Check that timezone remained the same
	if updatedUser.Timezone() != originalUser.Timezone() {
		t.Errorf("timezone should not have changed")
	}

	// Check that ID remained the same
	if updatedUser.ID() != originalUser.ID() {
		t.Errorf("user ID should not have changed")
	}

	// Test full update
	fullUpdateReq := UpdateUserRequest{
		Name:     common.Some("Bob Johnson"),
		Timezone: common.Some("America/New_York"),
	}

	fullyUpdatedUserResult := originalUser.UpdateUser(fullUpdateReq)
	if fullyUpdatedUserResult.IsErr() {
		t.Fatalf("failed to fully update user: %v", fullyUpdatedUserResult.Error())
	}

	fullyUpdatedUser := fullyUpdatedUserResult.Value()

	if fullyUpdatedUser.Name() != "Bob Johnson" {
		t.Errorf("expected name 'Bob Johnson' but got '%s'", fullyUpdatedUser.Name())
	}
	if fullyUpdatedUser.Timezone() != "America/New_York" {
		t.Errorf("expected timezone 'America/New_York' but got '%s'", fullyUpdatedUser.Timezone())
	}
}

func TestFunctionalComposition(t *testing.T) {
	// Test that we can compose operations using functional patterns

	// Start with a user creation request
	req := CreateUserRequest{
		Email:    "  TEST@EXAMPLE.COM  ", // Test normalization
		Name:     "  John Doe  ",         // Test trimming
		Timezone: "",                     // Test default timezone
	}

	// Create user (this will normalize and validate)
	userResult := NewUser(req)
	if userResult.IsErr() {
		t.Fatalf("failed to create user: %v", userResult.Error())
	}

	user := userResult.Value()

	// Verify normalization happened
	if user.Email() != "test@example.com" {
		t.Errorf("email should be normalized to lowercase and trimmed")
	}
	if user.Name() != "John Doe" {
		t.Errorf("name should be trimmed")
	}
	if user.Timezone() != "UTC" {
		t.Errorf("timezone should default to UTC")
	}

	// Chain multiple transformations
	finalResult := common.Bind(
		userResult,
		func(u User) common.Result[User] {
			return u.WithName("Jane Doe")
		},
	)

	if finalResult.IsErr() {
		t.Fatalf("failed to chain transformations: %v", finalResult.Error())
	}

	finalUser := finalResult.Value()
	if finalUser.Name() != "Jane Doe" {
		t.Errorf("chained transformation should have updated name")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[len(s)-len(substr):] == substr || s[:len(substr)] == substr ||
		(len(s) > len(substr) && (s[len(s)-len(substr)-1:len(s)-1] == substr || s[1:len(substr)+1] == substr)))
}

func normalizeEmail(email string) string {
	// Simplified normalization for testing
	return email // In real implementation, this would trim and lowercase
}

// Benchmark tests to verify performance of functional patterns

func BenchmarkUserCreation(b *testing.B) {
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Timezone: "UTC",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := NewUser(req)
		if result.IsErr() {
			b.Fatalf("failed to create user: %v", result.Error())
		}
	}
}

func BenchmarkUserTransformation(b *testing.B) {
	req := CreateUserRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Timezone: "UTC",
	}

	userResult := NewUser(req)
	if userResult.IsErr() {
		b.Fatalf("failed to create user: %v", userResult.Error())
	}

	user := userResult.Value()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := user.WithName("Jane Doe")
		if result.IsErr() {
			b.Fatalf("failed to transform user: %v", result.Error())
		}
	}
}
