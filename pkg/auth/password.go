// Package auth provides authentication and authorization utilities
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

const (
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8

	// MaxPasswordLength is the maximum password length
	MaxPasswordLength = 72 // bcrypt limitation

	// BcryptCost is the cost factor for bcrypt hashing
	// Cost of 10 provides a good balance between security and performance
	BcryptCost = 10
)

// PasswordHasher provides password hashing and verification
type PasswordHasher struct {
	cost int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: BcryptCost,
	}
}

// HashPassword hashes a plaintext password using bcrypt
func (h *PasswordHasher) HashPassword(password string) common.Result[string] {
	// Validate password length
	if len(password) < MinPasswordLength {
		return common.Err[string](fmt.Errorf("password must be at least %d characters", MinPasswordLength))
	}

	if len(password) > MaxPasswordLength {
		return common.Err[string](fmt.Errorf("password must not exceed %d characters", MaxPasswordLength))
	}

	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return common.Err[string](fmt.Errorf("failed to hash password: %w", err))
	}

	return common.Ok(string(hashedBytes))
}

// VerifyPassword verifies a plaintext password against a bcrypt hash
func (h *PasswordHasher) VerifyPassword(password, hash string) common.Result[bool] {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return common.Ok(false) // Password doesn't match, but not an error
		}
		return common.Err[bool](fmt.Errorf("failed to verify password: %w", err))
	}

	return common.Ok(true)
}

// ValidatePassword validates password strength
func (h *PasswordHasher) ValidatePassword(password string) common.Result[bool] {
	if len(password) < MinPasswordLength {
		return common.Err[bool](fmt.Errorf("password must be at least %d characters", MinPasswordLength))
	}

	if len(password) > MaxPasswordLength {
		return common.Err[bool](fmt.Errorf("password must not exceed %d characters", MaxPasswordLength))
	}

	// Check for at least one letter and one number (basic strength check)
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
		if hasLetter && hasNumber {
			break
		}
	}

	if !hasLetter || !hasNumber {
		return common.Err[bool](fmt.Errorf("password must contain at least one letter and one number"))
	}

	return common.Ok(true)
}
