// Package handlers provides HTTP request handlers
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thanhphuchuynh/dear-future/pkg/auth"
	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
	"github.com/thanhphuchuynh/dear-future/pkg/middleware"
)

// UserHandler handles user-related requests
type UserHandler struct {
	app            *composition.App
	jwtService     *auth.JWTService
	passwordHasher *auth.PasswordHasher
}

// NewUserHandler creates a new user handler
func NewUserHandler(app *composition.App, jwtService *auth.JWTService) *UserHandler {
	return &UserHandler{
		app:            app,
		jwtService:     jwtService,
		passwordHasher: auth.NewPasswordHasher(),
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Timezone  string `json:"timezone"`
	CreatedAt string `json:"created_at"`
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "email, name, and password are required")
		return
	}

	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	// Validate password strength
	validationResult := h.passwordHasher.ValidatePassword(req.Password)
	if validationResult.IsErr() {
		respondWithError(w, http.StatusBadRequest, validationResult.Error().Error())
		return
	}

	// Hash password
	hashedPasswordResult := h.passwordHasher.HashPassword(req.Password)
	if hashedPasswordResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to process password")
		return
	}
	hashedPassword := hashedPasswordResult.Value()

	// Create user in domain
	createUserReq := user.CreateUserRequest{
		Email:    req.Email,
		Name:     req.Name,
		Timezone: req.Timezone,
	}

	userResult := user.NewUser(createUserReq)
	if userResult.IsErr() {
		respondWithError(w, http.StatusBadRequest, userResult.Error().Error())
		return
	}

	newUser := userResult.Value()

	// Save user to database
	// TODO: In production, save hashedPassword alongside user
	// For now, using mock database which doesn't persist passwords
	saveResult := h.app.Database().SaveUser(r.Context(), newUser)
	if saveResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	savedUser := saveResult.Value()

	// In production, store the mapping: userID -> hashedPassword
	// For now, we're using mock authentication
	_ = hashedPassword // Acknowledge we hashed it, but mocks don't persist it

	// Generate JWT tokens
	tokenPairResult := h.jwtService.GenerateTokenPair(savedUser.ID(), savedUser.Email())
	if tokenPairResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	tokenPair := tokenPairResult.Value()

	// Prepare response
	response := AuthResponse{
		User: UserResponse{
			ID:        savedUser.ID().String(),
			Email:     savedUser.Email(),
			Name:      savedUser.Name(),
			Timezone:  savedUser.Timezone(),
			CreatedAt: savedUser.CreatedAt().Format("2006-01-02T15:04:05Z"),
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// Login handles user login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Find user by email
	userResult := h.app.Database().FindUserByEmail(r.Context(), req.Email)
	if userResult.IsErr() {
		respondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	foundUser := userResult.Value()

	// TODO: Verify password hash (for now, mock authentication is used)
	// In production, you would:
	// - Hash the provided password
	// - Compare with stored hash
	// - Return error if they don't match

	// Generate JWT tokens
	tokenPairResult := h.jwtService.GenerateTokenPair(foundUser.ID(), foundUser.Email())
	if tokenPairResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	tokenPair := tokenPairResult.Value()

	// Prepare response
	response := AuthResponse{
		User: UserResponse{
			ID:        foundUser.ID().String(),
			Email:     foundUser.Email(),
			Name:      foundUser.Name(),
			Timezone:  foundUser.Timezone(),
			CreatedAt: foundUser.CreatedAt().Format("2006-01-02T15:04:05Z"),
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Find user
	userResult := h.app.Database().FindUserByID(r.Context(), userID)
	if userResult.IsErr() {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	foundUser := userResult.Value()

	response := UserResponse{
		ID:        foundUser.ID().String(),
		Email:     foundUser.Email(),
		Name:      foundUser.Name(),
		Timezone:  foundUser.Timezone(),
		CreatedAt: foundUser.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name     *string `json:"name"`
		Timezone *string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Find user
	userResult := h.app.Database().FindUserByID(r.Context(), userID)
	if userResult.IsErr() {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	currentUser := userResult.Value()

	// Apply updates
	updatedUser := currentUser
	if req.Name != nil {
		updateResult := updatedUser.WithName(*req.Name)
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedUser = updateResult.Value()
	}

	if req.Timezone != nil {
		updateResult := updatedUser.WithTimezone(*req.Timezone)
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedUser = updateResult.Value()
	}

	// Save updated user
	saveResult := h.app.Database().UpdateUser(r.Context(), updatedUser)
	if saveResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	savedUser := saveResult.Value()

	response := UserResponse{
		ID:        savedUser.ID().String(),
		Email:     savedUser.Email(),
		Name:      savedUser.Name(),
		Timezone:  savedUser.Timezone(),
		CreatedAt: savedUser.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// RefreshToken generates a new access token from a refresh token
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondWithError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Validate and generate new tokens
	tokenPairResult := h.jwtService.RefreshAccessToken(req.RefreshToken)
	if tokenPairResult.IsErr() {
		respondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	tokenPair := tokenPairResult.Value()

	response := map[string]interface{}{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Helper functions

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
