// Package message contains the message domain types and business logic
package message

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
)

// MessageStatus represents the current status of a message
type MessageStatus string

const (
	StatusScheduled MessageStatus = "scheduled"
	StatusDelivered MessageStatus = "delivered"
	StatusFailed    MessageStatus = "failed"
	StatusCancelled MessageStatus = "cancelled"
)

// DeliveryMethod represents how the message will be delivered
type DeliveryMethod string

const (
	DeliveryEmail DeliveryMethod = "email"
	DeliveryPush  DeliveryMethod = "push"
)

// Message represents an immutable message entity
type Message struct {
	id             uuid.UUID
	userID         uuid.UUID
	title          string
	content        string
	deliveryDate   time.Time
	timezone       string
	status         MessageStatus
	deliveryMethod DeliveryMethod
	createdAt      time.Time
	updatedAt      time.Time
	deliveredAt    common.Option[time.Time]
}

// MessageAttachment represents a file attached to a message
type MessageAttachment struct {
	id         uuid.UUID
	messageID  uuid.UUID
	fileName   string
	fileType   string
	s3Key      string
	fileSize   int64
	uploadedAt time.Time
}

// CreateMessageRequest contains data needed to create a new message
type CreateMessageRequest struct {
	UserID         uuid.UUID
	Title          string
	Content        string
	DeliveryDate   time.Time
	Timezone       string
	DeliveryMethod DeliveryMethod
}

// UpdateMessageRequest contains data for updating a message
type UpdateMessageRequest struct {
	Title        common.Option[string]
	Content      common.Option[string]
	DeliveryDate common.Option[time.Time]
	Timezone     common.Option[string]
}

// MessageWithAttachments represents a message with its attachments
type MessageWithAttachments struct {
	message     Message
	attachments []MessageAttachment
}

// ScheduleResult represents the result of scheduling a message
type ScheduleResult struct {
	Message           Message
	ScheduledFor      time.Time
	EstimatedDelivery time.Time
}

// DeliveryResult represents the result of delivering a message
type DeliveryResult struct {
	MessageID   uuid.UUID
	Status      MessageStatus
	DeliveredAt time.Time
	Error       common.Option[string]
}

// NewMessage creates a new Message instance with validation
func NewMessage(req CreateMessageRequest) common.Result[Message] {
	// Validate the request
	validReq := validateCreateMessageRequest(req)
	if validReq.IsErr() {
		return common.Err[Message](validReq.Error())
	}

	// Create the message
	now := time.Now()
	message := Message{
		id:             uuid.New(),
		userID:         validReq.Value().UserID,
		title:          validReq.Value().Title,
		content:        validReq.Value().Content,
		deliveryDate:   validReq.Value().DeliveryDate,
		timezone:       validReq.Value().Timezone,
		status:         StatusScheduled,
		deliveryMethod: validReq.Value().DeliveryMethod,
		createdAt:      now,
		updatedAt:      now,
		deliveredAt:    common.None[time.Time](),
	}

	// Validate and normalize the message
	validMessage := validateMessage(message)
	if validMessage.IsErr() {
		return validMessage
	}

	return normalizeMessage(validMessage.Value())
}

// NewMessageAttachment creates a new MessageAttachment
func NewMessageAttachment(messageID uuid.UUID, fileName, fileType, s3Key string, fileSize int64) common.Result[MessageAttachment] {
	// Validate inputs
	if messageID == uuid.Nil {
		return common.Err[MessageAttachment](errors.New("message ID cannot be nil"))
	}

	validFileName := validateFileName(fileName)
	if validFileName.IsErr() {
		return common.Err[MessageAttachment](validFileName.Error())
	}

	validFileType := validateFileType(fileType)
	if validFileType.IsErr() {
		return common.Err[MessageAttachment](validFileType.Error())
	}

	if s3Key == "" {
		return common.Err[MessageAttachment](errors.New("S3 key cannot be empty"))
	}

	if fileSize <= 0 {
		return common.Err[MessageAttachment](errors.New("file size must be positive"))
	}

	attachment := MessageAttachment{
		id:         uuid.New(),
		messageID:  messageID,
		fileName:   validFileName.Value(),
		fileType:   validFileType.Value(),
		s3Key:      s3Key,
		fileSize:   fileSize,
		uploadedAt: time.Now(),
	}

	return common.Ok(attachment)
}

// Getters for Message (immutable access)
func (m Message) ID() uuid.UUID {
	return m.id
}

func (m Message) UserID() uuid.UUID {
	return m.userID
}

func (m Message) Title() string {
	return m.title
}

func (m Message) Content() string {
	return m.content
}

func (m Message) DeliveryDate() time.Time {
	return m.deliveryDate
}

func (m Message) Timezone() string {
	return m.timezone
}

func (m Message) Status() MessageStatus {
	return m.status
}

func (m Message) DeliveryMethod() DeliveryMethod {
	return m.deliveryMethod
}

func (m Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m Message) UpdatedAt() time.Time {
	return m.updatedAt
}

func (m Message) DeliveredAt() common.Option[time.Time] {
	return m.deliveredAt
}

// Getters for MessageAttachment
func (ma MessageAttachment) ID() uuid.UUID {
	return ma.id
}

func (ma MessageAttachment) MessageID() uuid.UUID {
	return ma.messageID
}

func (ma MessageAttachment) FileName() string {
	return ma.fileName
}

func (ma MessageAttachment) FileType() string {
	return ma.fileType
}

func (ma MessageAttachment) S3Key() string {
	return ma.s3Key
}

func (ma MessageAttachment) FileSize() int64 {
	return ma.fileSize
}

func (ma MessageAttachment) UploadedAt() time.Time {
	return ma.uploadedAt
}

// Getters for MessageWithAttachments
func (mwa MessageWithAttachments) Message() Message {
	return mwa.message
}

func (mwa MessageWithAttachments) Attachments() []MessageAttachment {
	return mwa.attachments
}

// NewMessageWithAttachments creates a MessageWithAttachments
func NewMessageWithAttachments(message Message, attachments []MessageAttachment) MessageWithAttachments {
	return MessageWithAttachments{
		message:     message,
		attachments: attachments,
	}
}

// Pure transformation functions (return new instances)

// WithTitle returns a new Message with updated title
func (m Message) WithTitle(title string) common.Result[Message] {
	validTitle := validateTitle(title)
	if validTitle.IsErr() {
		return common.Err[Message](validTitle.Error())
	}

	updated := Message{
		id:             m.id,
		userID:         m.userID,
		title:          validTitle.Value(),
		content:        m.content,
		deliveryDate:   m.deliveryDate,
		timezone:       m.timezone,
		status:         m.status,
		deliveryMethod: m.deliveryMethod,
		createdAt:      m.createdAt,
		updatedAt:      time.Now(),
		deliveredAt:    m.deliveredAt,
	}
	return common.Ok(updated)
}

// WithContent returns a new Message with updated content
func (m Message) WithContent(content string) common.Result[Message] {
	validContent := validateContent(content)
	if validContent.IsErr() {
		return common.Err[Message](validContent.Error())
	}

	updated := Message{
		id:             m.id,
		userID:         m.userID,
		title:          m.title,
		content:        validContent.Value(),
		deliveryDate:   m.deliveryDate,
		timezone:       m.timezone,
		status:         m.status,
		deliveryMethod: m.deliveryMethod,
		createdAt:      m.createdAt,
		updatedAt:      time.Now(),
		deliveredAt:    m.deliveredAt,
	}
	return common.Ok(updated)
}

// WithDeliveryDate returns a new Message with updated delivery date
func (m Message) WithDeliveryDate(deliveryDate time.Time, timezone string) common.Result[Message] {
	validDelivery := validateDeliveryDate(deliveryDate, timezone)
	if validDelivery.IsErr() {
		return common.Err[Message](validDelivery.Error())
	}

	updated := Message{
		id:             m.id,
		userID:         m.userID,
		title:          m.title,
		content:        m.content,
		deliveryDate:   deliveryDate,
		timezone:       timezone,
		status:         m.status,
		deliveryMethod: m.deliveryMethod,
		createdAt:      m.createdAt,
		updatedAt:      time.Now(),
		deliveredAt:    m.deliveredAt,
	}
	return common.Ok(updated)
}

// WithStatus returns a new Message with updated status
func (m Message) WithStatus(status MessageStatus) common.Result[Message] {
	validStatus := validateStatusTransition(m.status, status)
	if validStatus.IsErr() {
		return common.Err[Message](validStatus.Error())
	}

	now := time.Now()
	deliveredAt := m.deliveredAt

	// Set delivered time if status is changing to delivered
	if status == StatusDelivered && m.status != StatusDelivered {
		deliveredAt = common.Some(now)
	}

	updated := Message{
		id:             m.id,
		userID:         m.userID,
		title:          m.title,
		content:        m.content,
		deliveryDate:   m.deliveryDate,
		timezone:       m.timezone,
		status:         status,
		deliveryMethod: m.deliveryMethod,
		createdAt:      m.createdAt,
		updatedAt:      now,
		deliveredAt:    deliveredAt,
	}
	return common.Ok(updated)
}

// UpdateMessage applies updates to a message
func (m Message) UpdateMessage(req UpdateMessageRequest) common.Result[Message] {
	// Can only update scheduled messages
	if m.status != StatusScheduled {
		return common.Err[Message](errors.New("can only update scheduled messages"))
	}

	result := common.Ok(m)

	// Apply title update if provided
	if req.Title.IsSome() {
		result = common.Bind(result, func(message Message) common.Result[Message] {
			return message.WithTitle(req.Title.Value())
		})
	}

	// Apply content update if provided
	if req.Content.IsSome() {
		result = common.Bind(result, func(message Message) common.Result[Message] {
			return message.WithContent(req.Content.Value())
		})
	}

	// Apply delivery date update if provided
	if req.DeliveryDate.IsSome() {
		timezone := m.timezone
		if req.Timezone.IsSome() {
			timezone = req.Timezone.Value()
		}
		result = common.Bind(result, func(message Message) common.Result[Message] {
			return message.WithDeliveryDate(req.DeliveryDate.Value(), timezone)
		})
	}

	return result
}

// Business logic functions

// IsEditable returns true if the message can be edited
func (m Message) IsEditable() bool {
	return m.status == StatusScheduled
}

// IsDeletable returns true if the message can be deleted
func (m Message) IsDeletable() bool {
	return m.status == StatusScheduled || m.status == StatusFailed
}

// IsDeliverable returns true if the message is ready for delivery
func (m Message) IsDeliverable() bool {
	return m.status == StatusScheduled && time.Now().After(m.deliveryDate)
}

// GetDeliveryTimeInTimezone returns the delivery time in the message's timezone
func (m Message) GetDeliveryTimeInTimezone() common.Result[time.Time] {
	loc, err := time.LoadLocation(m.timezone)
	if err != nil {
		return common.Err[time.Time](err)
	}
	return common.Ok(m.deliveryDate.In(loc))
}

// GetTimeUntilDelivery returns the duration until delivery
func (m Message) GetTimeUntilDelivery() time.Duration {
	return time.Until(m.deliveryDate)
}

// HasAttachments returns true if the message has attachments
func (mwa MessageWithAttachments) HasAttachments() bool {
	return len(mwa.attachments) > 0
}

// GetAttachmentCount returns the number of attachments
func (mwa MessageWithAttachments) GetAttachmentCount() int {
	return len(mwa.attachments)
}

// GetTotalAttachmentSize returns the total size of all attachments
func (mwa MessageWithAttachments) GetTotalAttachmentSize() int64 {
	var total int64
	for _, attachment := range mwa.attachments {
		total += attachment.fileSize
	}
	return total
}

// AddAttachment returns a new MessageWithAttachments with an additional attachment
func (mwa MessageWithAttachments) AddAttachment(attachment MessageAttachment) common.Result[MessageWithAttachments] {
	// Validate that attachment belongs to this message
	if attachment.messageID != mwa.message.id {
		return common.Err[MessageWithAttachments](errors.New("attachment does not belong to this message"))
	}

	// Check attachment limits
	if len(mwa.attachments) >= 10 { // Max 10 attachments per message
		return common.Err[MessageWithAttachments](errors.New("maximum number of attachments reached"))
	}

	newTotalSize := mwa.GetTotalAttachmentSize() + attachment.fileSize
	if newTotalSize > 100*1024*1024 { // Max 100MB total
		return common.Err[MessageWithAttachments](errors.New("total attachment size exceeds limit"))
	}

	newAttachments := make([]MessageAttachment, len(mwa.attachments)+1)
	copy(newAttachments, mwa.attachments)
	newAttachments[len(mwa.attachments)] = attachment

	updated := MessageWithAttachments{
		message:     mwa.message,
		attachments: newAttachments,
	}

	return common.Ok(updated)
}

// RemoveAttachment returns a new MessageWithAttachments with the attachment removed
func (mwa MessageWithAttachments) RemoveAttachment(attachmentID uuid.UUID) MessageWithAttachments {
	filteredAttachments := common.FilterSlice(mwa.attachments, func(a MessageAttachment) bool {
		return a.id != attachmentID
	})

	return MessageWithAttachments{
		message:     mwa.message,
		attachments: filteredAttachments,
	}
}
