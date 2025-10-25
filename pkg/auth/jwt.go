// Package auth provides authentication and authorization utilities
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// Claims represents JWT claims for authentication
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID uuid.UUID, email string) common.Result[TokenPair] {
	now := time.Now()
	expiresAt := now.Add(j.accessTokenExpiry)
	fmt.Println("Generating tokens for user:", userID, "expires at:", expiresAt)
	// Create access token
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secretKey)
	if err != nil {
		return common.Err[TokenPair](fmt.Errorf("failed to sign access token: %w", err))
	}

	// Create refresh token
	refreshClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secretKey)
	if err != nil {
		return common.Err[TokenPair](fmt.Errorf("failed to sign refresh token: %w", err))
	}

	return common.Ok(TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	})
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) common.Result[Claims] {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return common.Err[Claims](fmt.Errorf("invalid token: %w", err))
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return common.Ok(*claims)
	}

	return common.Err[Claims](fmt.Errorf("invalid token claims"))
}

// RefreshAccessToken generates a new access token from a refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string) common.Result[TokenPair] {
	// Validate refresh token
	claimsResult := j.ValidateToken(refreshTokenString)
	if claimsResult.IsErr() {
		return common.Err[TokenPair](fmt.Errorf("invalid refresh token: %w", claimsResult.Error()))
	}

	claims := claimsResult.Value()

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email)
}

// ExtractUserID extracts user ID from a token without full validation
func (j *JWTService) ExtractUserID(tokenString string) common.Result[uuid.UUID] {
	claimsResult := j.ValidateToken(tokenString)
	if claimsResult.IsErr() {
		return common.Err[uuid.UUID](claimsResult.Error())
	}

	return common.Ok(claimsResult.Value().UserID)
}
