// Package handlers provides HTTP request handlers
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/middleware"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	app *composition.App
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(app *composition.App) *MessageHandler {
	return &MessageHandler{
		app: app,
	}
}

// CreateMessageRequest represents a message creation request
type CreateMessageRequest struct {
	Title          string `json:"title"`
	Content        string `json:"content"`
	DeliveryDate   string `json:"delivery_date"` // ISO 8601 format
	Timezone       string `json:"timezone"`
	DeliveryMethod string `json:"delivery_method"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	DeliveryDate   string `json:"delivery_date"`
	Timezone       string `json:"timezone"`
	Status         string `json:"status"`
	DeliveryMethod string `json:"delivery_method"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateMessage creates a new message
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Title == "" || req.Content == "" || req.DeliveryDate == "" {
		respondWithError(w, http.StatusBadRequest, "title, content, and delivery_date are required")
		return
	}

	// Parse delivery date
	deliveryDate, err := time.Parse(time.RFC3339, req.DeliveryDate)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid delivery_date format (use RFC3339)")
		return
	}

	// Set defaults
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	if req.DeliveryMethod == "" {
		req.DeliveryMethod = "email"
	}

	// Parse delivery method
	var deliveryMethod message.DeliveryMethod
	switch req.DeliveryMethod {
	case "email":
		deliveryMethod = message.DeliveryEmail
	case "push":
		deliveryMethod = message.DeliveryPush
	default:
		respondWithError(w, http.StatusBadRequest, "invalid delivery_method (must be 'email' or 'push')")
		return
	}

	// Create message in domain
	createMsgReq := message.CreateMessageRequest{
		UserID:         userID,
		Title:          req.Title,
		Content:        req.Content,
		DeliveryDate:   deliveryDate,
		Timezone:       req.Timezone,
		DeliveryMethod: deliveryMethod,
	}

	msgResult := message.NewMessage(createMsgReq)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusBadRequest, msgResult.Error().Error())
		return
	}

	newMsg := msgResult.Value()

	// Save message to database
	saveResult := h.app.Database().SaveMessage(r.Context(), newMsg)
	if saveResult.IsErr() {
		slog.Error("Failed to save message", "error", saveResult.Error())
		respondWithError(w, http.StatusInternalServerError, "failed to create message")
		return
	}

	savedMsg := saveResult.Value()

	response := MessageResponse{
		ID:             savedMsg.ID().String(),
		UserID:         savedMsg.UserID().String(),
		Title:          savedMsg.Title(),
		Content:        savedMsg.Content(),
		DeliveryDate:   savedMsg.DeliveryDate().Format(time.RFC3339),
		Timezone:       savedMsg.Timezone(),
		Status:         string(savedMsg.Status()),
		DeliveryMethod: string(savedMsg.DeliveryMethod()),
		CreatedAt:      savedMsg.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      savedMsg.UpdatedAt().Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// GetMessages returns all messages for the current user
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get messages from database
	messagesResult := h.app.Database().FindMessagesByUserID(r.Context(), userID, limit, offset)
	if messagesResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to retrieve messages")
		return
	}

	messages := messagesResult.Value()

	// Convert to response format
	var response []MessageResponse
	for _, msg := range messages {
		response = append(response, MessageResponse{
			ID:             msg.ID().String(),
			UserID:         msg.UserID().String(),
			Title:          msg.Title(),
			Content:        msg.Content(),
			DeliveryDate:   msg.DeliveryDate().Format(time.RFC3339),
			Timezone:       msg.Timezone(),
			Status:         string(msg.Status()),
			DeliveryMethod: string(msg.DeliveryMethod()),
			CreatedAt:      msg.CreatedAt().Format(time.RFC3339),
			UpdatedAt:      msg.UpdatedAt().Format(time.RFC3339),
		})
	}

	if response == nil {
		response = []MessageResponse{} // Return empty array instead of null
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetMessage returns a specific message by ID
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	messageIDStr := r.URL.Query().Get("id")
	if messageIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "message id is required")
		return
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	// Get message from database
	msgResult := h.app.Database().FindMessageByID(r.Context(), messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusNotFound, "message not found")
		return
	}

	msg := msgResult.Value()

	// Verify ownership
	if msg.UserID() != userID {
		respondWithError(w, http.StatusForbidden, "access denied")
		return
	}

	response := MessageResponse{
		ID:             msg.ID().String(),
		UserID:         msg.UserID().String(),
		Title:          msg.Title(),
		Content:        msg.Content(),
		DeliveryDate:   msg.DeliveryDate().Format(time.RFC3339),
		Timezone:       msg.Timezone(),
		Status:         string(msg.Status()),
		DeliveryMethod: string(msg.DeliveryMethod()),
		CreatedAt:      msg.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      msg.UpdatedAt().Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// UpdateMessage updates a message
func (h *MessageHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	messageIDStr := r.URL.Query().Get("id")
	if messageIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "message id is required")
		return
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	// Get current message
	msgResult := h.app.Database().FindMessageByID(r.Context(), messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusNotFound, "message not found")
		return
	}

	currentMsg := msgResult.Value()

	// Verify ownership
	if currentMsg.UserID() != userID {
		respondWithError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if message is editable
	if !currentMsg.IsEditable() {
		respondWithError(w, http.StatusBadRequest, "message cannot be edited (already delivered or cancelled)")
		return
	}

	// Parse update request
	var req struct {
		Title        *string `json:"title"`
		Content      *string `json:"content"`
		DeliveryDate *string `json:"delivery_date"`
		Timezone     *string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Apply updates
	updatedMsg := currentMsg

	if req.Title != nil {
		updateResult := updatedMsg.WithTitle(*req.Title)
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedMsg = updateResult.Value()
	}

	if req.Content != nil {
		updateResult := updatedMsg.WithContent(*req.Content)
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedMsg = updateResult.Value()
	}

	if req.DeliveryDate != nil {
		deliveryDate, err := time.Parse(time.RFC3339, *req.DeliveryDate)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid delivery_date format")
			return
		}

		timezone := updatedMsg.Timezone()
		if req.Timezone != nil {
			timezone = *req.Timezone
		}

		updateResult := updatedMsg.WithDeliveryDate(deliveryDate, timezone)
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedMsg = updateResult.Value()
	}

	// Save updated message
	saveResult := h.app.Database().UpdateMessage(r.Context(), updatedMsg)
	if saveResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to update message")
		return
	}

	savedMsg := saveResult.Value()

	response := MessageResponse{
		ID:             savedMsg.ID().String(),
		UserID:         savedMsg.UserID().String(),
		Title:          savedMsg.Title(),
		Content:        savedMsg.Content(),
		DeliveryDate:   savedMsg.DeliveryDate().Format(time.RFC3339),
		Timezone:       savedMsg.Timezone(),
		Status:         string(savedMsg.Status()),
		DeliveryMethod: string(savedMsg.DeliveryMethod()),
		CreatedAt:      savedMsg.CreatedAt().Format(time.RFC3339),
		UpdatedAt:      savedMsg.UpdatedAt().Format(time.RFC3339),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// DeleteMessage deletes a message
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	messageIDStr := r.URL.Query().Get("id")
	if messageIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "message id is required")
		return
	}

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	// Get message to verify ownership
	msgResult := h.app.Database().FindMessageByID(r.Context(), messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusNotFound, "message not found")
		return
	}

	msg := msgResult.Value()

	// Verify ownership
	if msg.UserID() != userID {
		respondWithError(w, http.StatusForbidden, "access denied")
		return
	}

	// Check if message is deletable
	if !msg.IsDeletable() {
		respondWithError(w, http.StatusBadRequest, "message cannot be deleted (already delivered)")
		return
	}

	// Delete message
	deleteResult := h.app.Database().DeleteMessage(r.Context(), messageID)
	if deleteResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "message deleted successfully"})
}
