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
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
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
	Title           string `json:"title"`
	Content         string `json:"content"`
	DeliveryDate    string `json:"delivery_date"` // ISO 8601 format
	Timezone        string `json:"timezone"`
	DeliveryMethod  string `json:"delivery_method"`
	Recurrence      string `json:"recurrence"`
	ReminderMinutes *int   `json:"reminder_minutes"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID              string               `json:"id"`
	UserID          string               `json:"user_id"`
	Title           string               `json:"title"`
	Content         string               `json:"content"`
	DeliveryDate    string               `json:"delivery_date"`
	Timezone        string               `json:"timezone"`
	Status          string               `json:"status"`
	DeliveryMethod  string               `json:"delivery_method"`
	Recurrence      string               `json:"recurrence"`
	ReminderMinutes *int                 `json:"reminder_minutes,omitempty"`
	AttachmentCount int                  `json:"attachment_count"`
	Attachments     []AttachmentResponse `json:"attachments,omitempty"`
	CreatedAt       string               `json:"created_at"`
	UpdatedAt       string               `json:"updated_at"`
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

	recurrence := message.RecurrencePattern(req.Recurrence)
	reminderOption := common.None[int]()
	if req.ReminderMinutes != nil {
		reminderOption = common.Some(*req.ReminderMinutes)
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
		UserID:          userID,
		Title:           req.Title,
		Content:         req.Content,
		DeliveryDate:    deliveryDate,
		Timezone:        req.Timezone,
		DeliveryMethod:  deliveryMethod,
		Recurrence:      recurrence,
		ReminderMinutes: reminderOption,
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

	// Schedule the message for delivery using River Queue
	messageService := h.app.MessageService()
	if messageService != nil && messageService.Scheduling() != nil {
		scheduleResult := messageService.Scheduling().ScheduleMessage(
			r.Context(),
			savedMsg.ID(),
			savedMsg.DeliveryDate(),
		)
		if scheduleResult.IsErr() {
			slog.Error("Failed to schedule message", "message_id", savedMsg.ID(), "error", scheduleResult.Error())
			// Don't fail the request - message is saved, scheduling can be retried
		} else {
			slog.Info("Message scheduled successfully",
				"message_id", savedMsg.ID(),
				"scheduled_for", savedMsg.DeliveryDate())
		}
	}

	respondWithJSON(w, http.StatusCreated, buildMessageResponse(savedMsg))
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
		response = append(response, buildMessageResponse(msg))
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

	response := buildMessageResponse(msg)

	attachmentsResult := h.app.Database().FindAttachmentsByMessageID(r.Context(), msg.ID())
	if attachmentsResult.IsOk() {
		attachments := attachmentsResult.Value()
		response.AttachmentCount = len(attachments)
		for _, att := range attachments {
			response.Attachments = append(response.Attachments, attachmentToResponse(r.Context(), h.app, att))
		}
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
		Title           *string `json:"title"`
		Content         *string `json:"content"`
		DeliveryDate    *string `json:"delivery_date"`
		Timezone        *string `json:"timezone"`
		Recurrence      *string `json:"recurrence"`
		ReminderMinutes *int    `json:"reminder_minutes"`
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

	if req.Recurrence != nil {
		updateResult := updatedMsg.WithRecurrence(message.RecurrencePattern(*req.Recurrence))
		if updateResult.IsErr() {
			respondWithError(w, http.StatusBadRequest, updateResult.Error().Error())
			return
		}
		updatedMsg = updateResult.Value()
	}

	if req.ReminderMinutes != nil {
		var reminderOption common.Option[int]
		if *req.ReminderMinutes > 0 {
			reminderOption = common.Some(*req.ReminderMinutes)
		} else {
			reminderOption = common.None[int]()
		}

		updateResult := updatedMsg.WithReminderMinutes(reminderOption)
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

	// Reschedule the message if delivery date changed
	messageService := h.app.MessageService()
	if req.DeliveryDate != nil && messageService != nil && messageService.Scheduling() != nil {
		rescheduleResult := messageService.Scheduling().RescheduleMessage(
			r.Context(),
			savedMsg.ID(),
			savedMsg.DeliveryDate(),
		)
		if rescheduleResult.IsErr() {
			slog.Error("Failed to reschedule message", "message_id", savedMsg.ID(), "error", rescheduleResult.Error())
			// Don't fail the request - message is updated, rescheduling can be retried
		} else {
			slog.Info("Message rescheduled successfully",
				"message_id", savedMsg.ID(),
				"new_delivery_date", savedMsg.DeliveryDate())
		}
	}

	respondWithJSON(w, http.StatusOK, buildMessageResponse(savedMsg))
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

func buildMessageResponse(msg message.Message) MessageResponse {
	response := MessageResponse{
		ID:              msg.ID().String(),
		UserID:          msg.UserID().String(),
		Title:           msg.Title(),
		Content:         msg.Content(),
		DeliveryDate:    msg.DeliveryDate().Format(time.RFC3339),
		Timezone:        msg.Timezone(),
		Status:          string(msg.Status()),
		DeliveryMethod:  string(msg.DeliveryMethod()),
		Recurrence:      string(msg.Recurrence()),
		AttachmentCount: 0,
		CreatedAt:       msg.CreatedAt().Format(time.RFC3339),
		UpdatedAt:       msg.UpdatedAt().Format(time.RFC3339),
	}

	if reminder := msg.ReminderMinutes(); reminder.IsSome() {
		value := reminder.Value()
		response.ReminderMinutes = &value
	}

	return response
}
