// Package user contains the user domain types and business logic
package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// User represents an immutable user entity
type User struct {
	id        uuid.UUID
	email     string
	name      string
	timezone  string
	createdAt time.Time
	updatedAt time.Time
}

// UserProfile represents the user's profile information
type UserProfile struct {
	user               User
	profilePictureURL  common.Option[string]
	emailNotifications bool
	notificationEmail  common.Option[string]
}

// CreateUserRequest contains data needed to create a new user
type CreateUserRequest struct {
	Email    string
	Name     string
	Timezone string
}

// UpdateUserRequest contains data for updating user information
type UpdateUserRequest struct {
	Name     common.Option[string]
	Timezone common.Option[string]
}

// UpdateProfileRequest contains data for updating user profile
type UpdateProfileRequest struct {
	ProfilePictureURL  common.Option[string]
	EmailNotifications common.Option[bool]
	NotificationEmail  common.Option[string]
}

// NewUser creates a new User instance with validation
func NewUser(req CreateUserRequest) common.Result[User] {
	// Validate the request
	validReq := validateCreateUserRequest(req)
	if validReq.IsErr() {
		return common.Err[User](validReq.Error())
	}

	// Create the user
	now := time.Now()
	user := User{
		id:        uuid.New(),
		email:     validReq.Value().Email,
		name:      validReq.Value().Name,
		timezone:  validReq.Value().Timezone,
		createdAt: now,
		updatedAt: now,
	}

	// Validate and normalize the user
	validUser := validateUser(user)
	if validUser.IsErr() {
		return validUser
	}

	return normalizeUser(validUser.Value())
}

// NewUserProfile creates a new UserProfile with default settings
func NewUserProfile(user User) UserProfile {
	return UserProfile{
		user:               user,
		profilePictureURL:  common.None[string](),
		emailNotifications: true,
		notificationEmail:  common.None[string](),
	}
}

// Getters for User (immutable access)
func (u User) ID() uuid.UUID {
	return u.id
}

func (u User) Email() string {
	return u.email
}

func (u User) Name() string {
	return u.name
}

func (u User) Timezone() string {
	return u.timezone
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Getters for UserProfile
func (up UserProfile) User() User {
	return up.user
}

func (up UserProfile) ProfilePictureURL() common.Option[string] {
	return up.profilePictureURL
}

func (up UserProfile) EmailNotifications() bool {
	return up.emailNotifications
}

func (up UserProfile) NotificationEmail() common.Option[string] {
	return up.notificationEmail
}

// Pure transformation functions (return new instances)

// WithName returns a new User with updated name
func (u User) WithName(name string) common.Result[User] {
	if name == "" {
		return common.Err[User](errors.New("name cannot be empty"))
	}

	updated := User{
		id:        u.id,
		email:     u.email,
		name:      name,
		timezone:  u.timezone,
		createdAt: u.createdAt,
		updatedAt: time.Now(),
	}
	return common.Ok(updated)
}

// WithTimezone returns a new User with updated timezone
func (u User) WithTimezone(timezone string) common.Result[User] {
	validTz := validateTimezone(timezone)
	if validTz.IsErr() {
		return common.Err[User](validTz.Error())
	}

	updated := User{
		id:        u.id,
		email:     u.email,
		name:      u.name,
		timezone:  validTz.Value(),
		createdAt: u.createdAt,
		updatedAt: time.Now(),
	}
	return common.Ok(updated)
}

// UpdateUser applies updates to a user
func (u User) UpdateUser(req UpdateUserRequest) common.Result[User] {
	result := common.Ok(u)

	// Apply name update if provided
	if req.Name.IsSome() {
		result = common.Bind(result, func(user User) common.Result[User] {
			return user.WithName(req.Name.Value())
		})
	}

	// Apply timezone update if provided
	if req.Timezone.IsSome() {
		result = common.Bind(result, func(user User) common.Result[User] {
			return user.WithTimezone(req.Timezone.Value())
		})
	}

	return result
}

// WithProfilePicture returns a new UserProfile with updated profile picture
func (up UserProfile) WithProfilePicture(url string) common.Result[UserProfile] {
	validURL := validateProfilePictureURL(url)
	if validURL.IsErr() {
		return common.Err[UserProfile](validURL.Error())
	}

	updated := UserProfile{
		user:               up.user,
		profilePictureURL:  common.Some(validURL.Value()),
		emailNotifications: up.emailNotifications,
		notificationEmail:  up.notificationEmail,
	}
	return common.Ok(updated)
}

// WithEmailNotifications returns a new UserProfile with updated notification settings
func (up UserProfile) WithEmailNotifications(enabled bool) UserProfile {
	return UserProfile{
		user:               up.user,
		profilePictureURL:  up.profilePictureURL,
		emailNotifications: enabled,
		notificationEmail:  up.notificationEmail,
	}
}

// WithNotificationEmail returns a new UserProfile with updated notification email
func (up UserProfile) WithNotificationEmail(email string) common.Result[UserProfile] {
	validEmail := validateEmail(email)
	if validEmail.IsErr() {
		return common.Err[UserProfile](validEmail.Error())
	}

	updated := UserProfile{
		user:               up.user,
		profilePictureURL:  up.profilePictureURL,
		emailNotifications: up.emailNotifications,
		notificationEmail:  common.Some(validEmail.Value()),
	}
	return common.Ok(updated)
}

// UpdateProfile applies updates to a user profile
func (up UserProfile) UpdateProfile(req UpdateProfileRequest) common.Result[UserProfile] {
	result := common.Ok(up)

	// Apply profile picture update if provided
	if req.ProfilePictureURL.IsSome() {
		result = common.Bind(result, func(profile UserProfile) common.Result[UserProfile] {
			return profile.WithProfilePicture(req.ProfilePictureURL.Value())
		})
	}

	// Apply email notifications update if provided
	if req.EmailNotifications.IsSome() {
		result = common.Map(result, func(profile UserProfile) UserProfile {
			return profile.WithEmailNotifications(req.EmailNotifications.Value())
		})
	}

	// Apply notification email update if provided
	if req.NotificationEmail.IsSome() {
		result = common.Bind(result, func(profile UserProfile) common.Result[UserProfile] {
			return profile.WithNotificationEmail(req.NotificationEmail.Value())
		})
	}

	return result
}

// GetDisplayName returns the user's display name or email if name is empty
func (u User) GetDisplayName() string {
	if u.name != "" {
		return u.name
	}
	return u.email
}

// GetEffectiveEmail returns the notification email if set, otherwise the user's email
func (up UserProfile) GetEffectiveEmail() string {
	return up.notificationEmail.ValueOr(up.user.email)
}

// IsEmailNotificationsEnabled returns true if email notifications are enabled
func (up UserProfile) IsEmailNotificationsEnabled() bool {
	return up.emailNotifications
}
