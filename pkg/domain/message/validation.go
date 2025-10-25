// Package message contains validation functions for message domain
package message

import (
	"errors"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// validateCreateMessageRequest validates the create message request
func validateCreateMessageRequest(req CreateMessageRequest) common.Result[CreateMessageRequest] {
	// Validate title
	titleResult := validateTitle(req.Title)
	if titleResult.IsErr() {
		return common.Err[CreateMessageRequest](titleResult.Error())
	}

	// Validate content
	contentResult := validateContent(req.Content)
	if contentResult.IsErr() {
		return common.Err[CreateMessageRequest](contentResult.Error())
	}

	// Validate delivery date
	deliveryResult := validateDeliveryDate(req.DeliveryDate, req.Timezone, false)
	if deliveryResult.IsErr() {
		return common.Err[CreateMessageRequest](deliveryResult.Error())
	}

	// Validate delivery method
	methodResult := validateDeliveryMethod(req.DeliveryMethod)
	if methodResult.IsErr() {
		return common.Err[CreateMessageRequest](methodResult.Error())
	}

	// Validate recurrence pattern
	recurrenceResult := validateRecurrence(req.Recurrence)
	if recurrenceResult.IsErr() {
		return common.Err[CreateMessageRequest](recurrenceResult.Error())
	}

	// Validate reminder offset
	reminderResult := validateReminderMinutes(req.ReminderMinutes)
	if reminderResult.IsErr() {
		return common.Err[CreateMessageRequest](reminderResult.Error())
	}

	return common.Ok(CreateMessageRequest{
		UserID:          req.UserID,
		Title:           titleResult.Value(),
		Content:         contentResult.Value(),
		DeliveryDate:    req.DeliveryDate,
		Timezone:        req.Timezone,
		DeliveryMethod:  methodResult.Value(),
		Recurrence:      recurrenceResult.Value(),
		ReminderMinutes: reminderResult.Value(),
	})
}

// validateTitle validates message title
func validateTitle(title string) common.Result[string] {
	if title == "" {
		return common.Err[string](errors.New("title cannot be empty"))
	}

	title = strings.TrimSpace(title)

	if len(title) < 1 {
		return common.Err[string](errors.New("title cannot be empty after trimming"))
	}

	if len(title) > 200 {
		return common.Err[string](errors.New("title is too long (max 200 characters)"))
	}

	// Check for reasonable characters
	if strings.Contains(title, "\n") || strings.Contains(title, "\r") {
		return common.Err[string](errors.New("title cannot contain line breaks"))
	}

	return common.Ok(title)
}

// validateContent validates message content
func validateContent(content string) common.Result[string] {
	if content == "" {
		return common.Err[string](errors.New("content cannot be empty"))
	}

	content = strings.TrimSpace(content)

	if len(content) < 1 {
		return common.Err[string](errors.New("content cannot be empty after trimming"))
	}

	if len(content) > 10000 {
		return common.Err[string](errors.New("content is too long (max 10,000 characters)"))
	}

	return common.Ok(content)
}

// validateDeliveryDate validates the delivery date and timezone
func validateDeliveryDate(deliveryDate time.Time, timezone string, allowPast bool) common.Result[time.Time] {
	// Validate timezone first
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return common.Err[time.Time](errors.New("invalid timezone: " + timezone))
	}

	// Convert delivery date to the specified timezone for validation
	deliveryInTz := deliveryDate.In(loc)
	now := time.Now().In(loc)

	// Must be in the future unless explicitly allowed (used when rehydrating existing messages)
	if !allowPast && deliveryInTz.Before(now) {
		return common.Err[time.Time](errors.New("delivery date must be in the future"))
	}

	// Must be within reasonable limits (e.g., not more than 50 years in the future)
	maxFuture := now.AddDate(50, 0, 0)
	if deliveryInTz.After(maxFuture) {
		return common.Err[time.Time](errors.New("delivery date is too far in the future"))
	}

	if !allowPast {
		// Must be at least 1 minute in the future to allow for processing
		minFuture := now.Add(1 * time.Minute)
		if deliveryInTz.Before(minFuture) {
			return common.Err[time.Time](errors.New("delivery date must be at least 1 minute in the future"))
		}
	}

	return common.Ok(deliveryDate)
}

// validateDeliveryMethod validates the delivery method
func validateDeliveryMethod(method DeliveryMethod) common.Result[DeliveryMethod] {
	switch method {
	case DeliveryEmail, DeliveryPush:
		return common.Ok(method)
	case "":
		return common.Ok(DeliveryEmail) // Default to email
	default:
		return common.Err[DeliveryMethod](errors.New("invalid delivery method"))
	}
}

// validateRecurrence validates recurrence pattern
func validateRecurrence(pattern RecurrencePattern) common.Result[RecurrencePattern] {
	if pattern == "" {
		return common.Ok(RecurrenceNone)
	}

	switch pattern {
	case RecurrenceNone, RecurrenceDaily, RecurrenceWeekly, RecurrenceMonthly, RecurrenceYearly:
		return common.Ok(pattern)
	default:
		return common.Err[RecurrencePattern](errors.New("invalid recurrence pattern"))
	}
}

// validateReminderMinutes validates reminder offsets
func validateReminderMinutes(reminder common.Option[int]) common.Result[common.Option[int]] {
	if reminder.IsNone() {
		return common.Ok(common.None[int]())
	}

	value := reminder.Value()
	if value == 0 {
		return common.Ok(common.None[int]())
	}
	if value < 0 {
		return common.Err[common.Option[int]](errors.New("reminder minutes must be positive"))
	}

	// Limit reminders to 30 days (43200 minutes) to prevent unrealistic values
	if value > 43200 {
		return common.Err[common.Option[int]](errors.New("reminder minutes cannot exceed 30 days"))
	}

	return common.Ok(common.Some(value))
}

// validateFileName validates attachment file name
func validateFileName(fileName string) common.Result[string] {
	if fileName == "" {
		return common.Err[string](errors.New("file name cannot be empty"))
	}

	fileName = strings.TrimSpace(fileName)

	if len(fileName) < 1 {
		return common.Err[string](errors.New("file name cannot be empty after trimming"))
	}

	if len(fileName) > 255 {
		return common.Err[string](errors.New("file name is too long (max 255 characters)"))
	}

	// Check for invalid characters in filename
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(fileName, char) {
			return common.Err[string](errors.New("file name contains invalid characters"))
		}
	}

	// Must have an extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		return common.Err[string](errors.New("file name must have an extension"))
	}

	return common.Ok(fileName)
}

// validateFileType validates attachment file type
func validateFileType(fileType string) common.Result[string] {
	if fileType == "" {
		return common.Err[string](errors.New("file type cannot be empty"))
	}

	fileType = strings.TrimSpace(strings.ToLower(fileType))

	// List of allowed file types
	allowedTypes := map[string]bool{
		"image/jpeg":         true,
		"image/jpg":          true,
		"image/png":          true,
		"image/gif":          true,
		"image/webp":         true,
		"image/svg+xml":      true,
		"application/pdf":    true,
		"text/plain":         true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/zip":              true,
		"application/x-zip-compressed": true,
		"audio/mpeg":                   true,
		"audio/mp3":                    true,
		"audio/wav":                    true,
		"video/mp4":                    true,
		"video/mpeg":                   true,
		"video/quicktime":              true,
	}

	if !allowedTypes[fileType] {
		return common.Err[string](errors.New("file type not allowed: " + fileType))
	}

	return common.Ok(fileType)
}

// validateMessage validates a complete message object
func validateMessage(message Message) common.Result[Message] {
	// Validate title
	titleResult := validateTitle(message.title)
	if titleResult.IsErr() {
		return common.Err[Message](titleResult.Error())
	}

	// Validate content
	contentResult := validateContent(message.content)
	if contentResult.IsErr() {
		return common.Err[Message](contentResult.Error())
	}

	// Validate delivery date
	deliveryResult := validateDeliveryDate(message.deliveryDate, message.timezone, true)
	if deliveryResult.IsErr() {
		return common.Err[Message](deliveryResult.Error())
	}

	// Validate delivery method
	methodResult := validateDeliveryMethod(message.deliveryMethod)
	if methodResult.IsErr() {
		return common.Err[Message](methodResult.Error())
	}

	// Validate status
	statusResult := validateMessageStatus(message.status)
	if statusResult.IsErr() {
		return common.Err[Message](statusResult.Error())
	}

	// Validate recurrence
	recurrenceResult := validateRecurrence(message.recurrence)
	if recurrenceResult.IsErr() {
		return common.Err[Message](recurrenceResult.Error())
	}

	// Validate reminder option
	reminderResult := validateReminderMinutes(message.reminderOffset)
	if reminderResult.IsErr() {
		return common.Err[Message](reminderResult.Error())
	}

	return common.Ok(message)
}

// validateMessageStatus validates message status
func validateMessageStatus(status MessageStatus) common.Result[MessageStatus] {
	switch status {
	case StatusScheduled, StatusDelivered, StatusFailed, StatusCancelled:
		return common.Ok(status)
	default:
		return common.Err[MessageStatus](errors.New("invalid message status"))
	}
}

// validateStatusTransition validates that a status transition is allowed
func validateStatusTransition(currentStatus, newStatus MessageStatus) common.Result[MessageStatus] {
	// Define valid transitions
	validTransitions := map[MessageStatus][]MessageStatus{
		StatusScheduled: {StatusDelivered, StatusFailed, StatusCancelled},
		StatusDelivered: {},                                 // No transitions from delivered
		StatusFailed:    {StatusScheduled, StatusCancelled}, // Can retry or cancel
		StatusCancelled: {},                                 // No transitions from cancelled
	}

	allowedNextStates, exists := validTransitions[currentStatus]
	if !exists {
		return common.Err[MessageStatus](errors.New("invalid current status"))
	}

	for _, allowedStatus := range allowedNextStates {
		if newStatus == allowedStatus {
			return common.Ok(newStatus)
		}
	}

	return common.Err[MessageStatus](errors.New("invalid status transition from " + string(currentStatus) + " to " + string(newStatus)))
}

// normalizeMessage normalizes message data
func normalizeMessage(message Message) common.Result[Message] {
	normalizedTitle := strings.TrimSpace(message.title)
	normalizedContent := strings.TrimSpace(message.content)
	normalizedTimezone := strings.TrimSpace(message.timezone)
	normalizedRecurrence := message.recurrence

	if normalizedTimezone == "" {
		normalizedTimezone = "UTC"
	}
	if normalizedRecurrence == "" {
		normalizedRecurrence = RecurrenceNone
	}

	normalized := Message{
		id:             message.id,
		userID:         message.userID,
		title:          normalizedTitle,
		content:        normalizedContent,
		deliveryDate:   message.deliveryDate,
		timezone:       normalizedTimezone,
		status:         message.status,
		deliveryMethod: message.deliveryMethod,
		recurrence:     normalizedRecurrence,
		reminderOffset: message.reminderOffset,
		createdAt:      message.createdAt,
		updatedAt:      message.updatedAt,
		deliveredAt:    message.deliveredAt,
	}

	return common.Ok(normalized)
}

// Business validation functions

// IsValidDeliveryTime checks if the delivery time is optimal
func IsValidDeliveryTime(deliveryDate time.Time, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false
	}

	deliveryInTz := deliveryDate.In(loc)
	hour := deliveryInTz.Hour()

	// Reasonable delivery hours (8 AM to 10 PM)
	return hour >= 8 && hour <= 22
}

// GetOptimalDeliveryTime suggests an optimal delivery time based on the requested time
func GetOptimalDeliveryTime(requestedTime time.Time, timezone string) common.Result[time.Time] {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return common.Err[time.Time](err)
	}

	requestedInTz := requestedTime.In(loc)
	hour := requestedInTz.Hour()

	// If it's a reasonable time, return as-is
	if hour >= 8 && hour <= 22 {
		return common.Ok(requestedTime)
	}

	// Adjust to reasonable time
	var adjustedTime time.Time
	if hour < 8 {
		// Move to 9 AM same day
		adjustedTime = time.Date(requestedInTz.Year(), requestedInTz.Month(), requestedInTz.Day(),
			9, 0, 0, 0, loc)
	} else {
		// Move to 9 AM next day
		nextDay := requestedInTz.AddDate(0, 0, 1)
		adjustedTime = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(),
			9, 0, 0, 0, loc)
	}

	return common.Ok(adjustedTime.UTC())
}

// ValidateAttachmentLimits checks if attachment limits are respected
func ValidateAttachmentLimits(attachments []MessageAttachment, newFileSize int64) common.Result[bool] {
	// Check count limit
	if len(attachments) >= 10 {
		return common.Err[bool](errors.New("maximum number of attachments (10) reached"))
	}

	// Check total size limit
	var totalSize int64
	for _, attachment := range attachments {
		totalSize += attachment.fileSize
	}

	totalSize += newFileSize
	maxSize := int64(100 * 1024 * 1024) // 100MB

	if totalSize > maxSize {
		return common.Err[bool](errors.New("total attachment size exceeds 100MB limit"))
	}

	// Check individual file size limit
	maxFileSize := int64(50 * 1024 * 1024) // 50MB per file
	if newFileSize > maxFileSize {
		return common.Err[bool](errors.New("individual file size exceeds 50MB limit"))
	}

	return common.Ok(true)
}

// GetFileTypeFromExtension determines file type from file extension
func GetFileTypeFromExtension(fileName string) string {
	ext := filepath.Ext(fileName)
	return mime.TypeByExtension(ext)
}

// IsImageFile checks if the file is an image
func IsImageFile(fileType string) bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/svg+xml",
	}

	for _, imgType := range imageTypes {
		if fileType == imgType {
			return true
		}
	}
	return false
}

// IsDocumentFile checks if the file is a document
func IsDocumentFile(fileType string) bool {
	docTypes := []string{
		"application/pdf", "text/plain", "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	for _, docType := range docTypes {
		if fileType == docType {
			return true
		}
	}
	return false
}
