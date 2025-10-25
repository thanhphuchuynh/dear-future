// Package user contains validation functions for user domain
package user

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// Email validation regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateCreateUserRequest validates the create user request
func validateCreateUserRequest(req CreateUserRequest) common.Result[CreateUserRequest] {
	// Validate email
	emailResult := validateEmail(req.Email)
	if emailResult.IsErr() {
		return common.Err[CreateUserRequest](emailResult.Error())
	}

	// Validate name
	nameResult := validateName(req.Name)
	if nameResult.IsErr() {
		return common.Err[CreateUserRequest](nameResult.Error())
	}

	// Validate timezone
	timezoneResult := validateTimezone(req.Timezone)
	if timezoneResult.IsErr() {
		return common.Err[CreateUserRequest](timezoneResult.Error())
	}

	return common.Ok(CreateUserRequest{
		Email:    emailResult.Value(),
		Name:     nameResult.Value(),
		Timezone: timezoneResult.Value(),
		UserID:   req.UserID,
	})
}

// validateEmail validates email format
func validateEmail(email string) common.Result[string] {
	if email == "" {
		return common.Err[string](errors.New("email cannot be empty"))
	}

	email = strings.TrimSpace(strings.ToLower(email))

	if !emailRegex.MatchString(email) {
		return common.Err[string](errors.New("invalid email format"))
	}

	if len(email) > 254 {
		return common.Err[string](errors.New("email is too long"))
	}

	return common.Ok(email)
}

// validateName validates user name
func validateName(name string) common.Result[string] {
	if name == "" {
		return common.Err[string](errors.New("name cannot be empty"))
	}

	name = strings.TrimSpace(name)

	if len(name) < 1 {
		return common.Err[string](errors.New("name cannot be empty after trimming"))
	}

	if len(name) > 100 {
		return common.Err[string](errors.New("name is too long (max 100 characters)"))
	}

	// Check for valid characters (letters, numbers, spaces, hyphens, apostrophes)
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-'\.]+$`)
	if !validNameRegex.MatchString(name) {
		return common.Err[string](errors.New("name contains invalid characters"))
	}

	return common.Ok(name)
}

// validateTimezone validates timezone string
func validateTimezone(timezone string) common.Result[string] {
	if timezone == "" {
		return common.Ok("UTC") // Default to UTC
	}

	timezone = strings.TrimSpace(timezone)

	// Try to load the timezone to validate it
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return common.Err[string](errors.New("invalid timezone: " + timezone))
	}

	return common.Ok(timezone)
}

// validateProfilePictureURL validates profile picture URL
func validateProfilePictureURL(pictureURL string) common.Result[string] {
	if pictureURL == "" {
		return common.Err[string](errors.New("profile picture URL cannot be empty"))
	}

	pictureURL = strings.TrimSpace(pictureURL)

	// Parse URL to validate format
	parsedURL, err := url.Parse(pictureURL)
	if err != nil {
		return common.Err[string](errors.New("invalid URL format"))
	}

	// Must be HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return common.Err[string](errors.New("profile picture URL must use HTTP or HTTPS"))
	}

	// Must have a host
	if parsedURL.Host == "" {
		return common.Err[string](errors.New("profile picture URL must have a valid host"))
	}

	// Check URL length
	if len(pictureURL) > 2048 {
		return common.Err[string](errors.New("profile picture URL is too long"))
	}

	// Check for common image file extensions
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	path := strings.ToLower(parsedURL.Path)
	hasValidExtension := false

	for _, ext := range validExtensions {
		if strings.HasSuffix(path, ext) {
			hasValidExtension = true
			break
		}
	}

	// Allow URLs without extensions (could be dynamic image URLs)
	if !hasValidExtension && strings.Contains(path, ".") {
		return common.Err[string](errors.New("profile picture URL must point to a valid image file"))
	}

	return common.Ok(pictureURL)
}

// validateUser validates a complete user object
func validateUser(user User) common.Result[User] {
	// Validate email
	emailResult := validateEmail(user.email)
	if emailResult.IsErr() {
		return common.Err[User](emailResult.Error())
	}

	// Validate name
	nameResult := validateName(user.name)
	if nameResult.IsErr() {
		return common.Err[User](nameResult.Error())
	}

	// Validate timezone
	timezoneResult := validateTimezone(user.timezone)
	if timezoneResult.IsErr() {
		return common.Err[User](timezoneResult.Error())
	}

	return common.Ok(user)
}

// normalizeUser normalizes user data (trims whitespace, converts email to lowercase, etc.)
func normalizeUser(user User) common.Result[User] {
	normalizedEmail := strings.TrimSpace(strings.ToLower(user.email))
	normalizedName := strings.TrimSpace(user.name)
	normalizedTimezone := strings.TrimSpace(user.timezone)

	if normalizedTimezone == "" {
		normalizedTimezone = "UTC"
	}

	normalized := User{
		id:        user.id,
		email:     normalizedEmail,
		name:      normalizedName,
		timezone:  normalizedTimezone,
		createdAt: user.createdAt,
		updatedAt: user.updatedAt,
	}

	return common.Ok(normalized)
}

// Validation helper functions for common patterns

// IsValidEmailDomain checks if email domain is in allowed list (if needed)
func IsValidEmailDomain(email string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		return true // No restrictions
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])
	for _, allowedDomain := range allowedDomains {
		if domain == strings.ToLower(allowedDomain) {
			return true
		}
	}

	return false
}

// IsStrongName checks if name meets strength requirements
func IsStrongName(name string) bool {
	name = strings.TrimSpace(name)

	// At least 2 characters
	if len(name) < 2 {
		return false
	}

	// Contains at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(name)

	return hasLetter
}

// GetSupportedTimezones returns a list of commonly supported timezones
func GetSupportedTimezones() []string {
	return []string{
		"UTC",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Kolkata",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
}

// IsCommonTimezone checks if timezone is in the common list
func IsCommonTimezone(timezone string) bool {
	supported := GetSupportedTimezones()
	for _, tz := range supported {
		if tz == timezone {
			return true
		}
	}
	return false
}
