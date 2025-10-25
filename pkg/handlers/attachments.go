package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/composition"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
	"github.com/thanhphuchuynh/dear-future/pkg/middleware"
)

// AttachmentHandler manages file attachments for messages.
type AttachmentHandler struct {
	app *composition.App
}

// AttachmentResponse represents attachment metadata returned to clients.
type AttachmentResponse struct {
	ID           string `json:"id"`
	MessageID    string `json:"message_id"`
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSize     int64  `json:"file_size"`
	DownloadURL  string `json:"download_url,omitempty"`
	UploadedAt   string `json:"uploaded_at"`
}

// NewAttachmentHandler creates a new handler instance.
func NewAttachmentHandler(app *composition.App) *AttachmentHandler {
	return &AttachmentHandler{app: app}
}

// Upload handles POST /api/v1/messages/attachments
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if !h.app.Config().Features.EnableFileAttachments {
		respondWithError(w, http.StatusBadRequest, "file attachments are disabled")
		return
	}

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID, err := parseUUIDParam(r, "message_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	msgResult := h.verifyMessageOwnership(r, userID, messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusForbidden, msgResult.Error().Error())
		return
	}

	maxSize := h.app.Config().FileUpload.MaxFileSize
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024
	}

	if err := r.ParseMultipartForm(maxSize); err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse upload form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to read file data")
		return
	}

	if int64(len(data)) > maxSize {
		respondWithError(w, http.StatusBadRequest, "file exceeds maximum allowed size")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	storage := h.app.FileHandler().Storage()
	if storage == nil {
		respondWithError(w, http.StatusInternalServerError, "storage service unavailable")
		return
	}

	uploadMetadata := map[string]string{
		"message_id":  messageID.String(),
		"uploaded_by": userID.String(),
	}

	uploadResult := storage.UploadFile(r.Context(), effects.FileUpload{
		FileName:    header.Filename,
		ContentType: contentType,
		Data:        data,
		Size:        int64(len(data)),
		Metadata:    uploadMetadata,
	})
	if uploadResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to upload file")
		return
	}

	attachmentResult := message.NewMessageAttachment(messageID, header.Filename, contentType, uploadResult.Value().Key, int64(len(data)))
	if attachmentResult.IsErr() {
		respondWithError(w, http.StatusBadRequest, attachmentResult.Error().Error())
		return
	}

	saveResult := h.app.Database().SaveMessageAttachment(r.Context(), attachmentResult.Value())
	if saveResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to save attachment metadata")
		return
	}

	response := attachmentToResponse(r.Context(), h.app, saveResult.Value())
	respondWithJSON(w, http.StatusCreated, response)
}

// List handles GET /api/v1/messages/attachments
func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID, err := parseUUIDParam(r, "message_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	msgResult := h.verifyMessageOwnership(r, userID, messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusForbidden, msgResult.Error().Error())
		return
	}

	attachmentsResult := h.app.Database().FindAttachmentsByMessageID(r.Context(), messageID)
	if attachmentsResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to load attachments")
		return
	}

	var responses []AttachmentResponse
	for _, attachment := range attachmentsResult.Value() {
		responses = append(responses, attachmentToResponse(r.Context(), h.app, attachment))
	}

	respondWithJSON(w, http.StatusOK, responses)
}

// Delete handles DELETE /api/v1/messages/attachments
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID, err := parseUUIDParam(r, "message_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid message_id")
		return
	}

	attachmentID, err := parseUUIDParam(r, "attachment_id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid attachment_id")
		return
	}

	msgResult := h.verifyMessageOwnership(r, userID, messageID)
	if msgResult.IsErr() {
		respondWithError(w, http.StatusForbidden, msgResult.Error().Error())
		return
	}

	attachmentsResult := h.app.Database().FindAttachmentsByMessageID(r.Context(), messageID)
	if attachmentsResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to load attachments")
		return
	}

	var target *message.MessageAttachment
	for _, attachment := range attachmentsResult.Value() {
		if attachment.ID() == attachmentID {
			target = &attachment
			break
		}
	}

	if target == nil {
		respondWithError(w, http.StatusNotFound, "attachment not found")
		return
	}

	storage := h.app.FileHandler().Storage()
	if storage != nil {
		deleteResult := storage.DeleteFile(r.Context(), target.S3Key())
		if deleteResult.IsErr() {
			respondWithError(w, http.StatusInternalServerError, "failed to remove file from storage")
			return
		}
	}

	removeResult := h.app.Database().DeleteAttachment(r.Context(), attachmentID)
	if removeResult.IsErr() {
		respondWithError(w, http.StatusInternalServerError, "failed to delete attachment")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AttachmentHandler) verifyMessageOwnership(r *http.Request, userID uuid.UUID, messageID uuid.UUID) common.Result[message.Message] {
	msgResult := h.app.Database().FindMessageByID(r.Context(), messageID)
	if msgResult.IsErr() {
		return common.Err[message.Message](msgResult.Error())
	}

	msg := msgResult.Value()
	if msg.UserID() != userID {
		return common.Err[message.Message](errors.New("access denied"))
	}

	return common.Ok(msg)
}

func attachmentToResponse(ctx context.Context, app *composition.App, attachment message.MessageAttachment) AttachmentResponse {
	response := AttachmentResponse{
		ID:         attachment.ID().String(),
		MessageID:  attachment.MessageID().String(),
		FileName:   attachment.FileName(),
		FileType:   attachment.FileType(),
		FileSize:   attachment.FileSize(),
		UploadedAt: attachment.UploadedAt().Format(time.RFC3339),
	}

	if fileHandler := app.FileHandler(); fileHandler != nil {
		if storage := fileHandler.Storage(); storage != nil {
			urlResult := storage.GeneratePresignedURL(ctx, attachment.S3Key(), 15*time.Minute)
			if urlResult.IsOk() {
				response.DownloadURL = urlResult.Value()
			}
		}
	}

	return response
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return uuid.Nil, errors.New("missing parameter")
	}
	return uuid.Parse(value)
}
