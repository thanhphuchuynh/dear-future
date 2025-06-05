// Package message contains pure business logic functions for message domain
package message

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
)

// ScheduleMessage schedules a message for delivery (pure business logic)
func ScheduleMessage(message Message, userProfile user.UserProfile) common.Result[ScheduleResult] {
	// Validate that the message is schedulable
	if !message.IsEditable() {
		return common.Err[ScheduleResult](errors.New("message is not in a schedulable state"))
	}

	// Calculate optimal delivery time
	optimalTime := GetOptimalDeliveryTime(message.DeliveryDate(), message.Timezone())
	if optimalTime.IsErr() {
		return common.Err[ScheduleResult](optimalTime.Error())
	}

	// Calculate estimated delivery time (accounting for processing time)
	estimatedDelivery := optimalTime.Value().Add(1 * time.Minute)

	result := ScheduleResult{
		Message:           message,
		ScheduledFor:      optimalTime.Value(),
		EstimatedDelivery: estimatedDelivery,
	}

	return common.Ok(result)
}

// ProcessMessageDelivery processes a message for delivery (pure business logic)
func ProcessMessageDelivery(message Message, recipient user.UserProfile) common.Result[MessageDeliveryInfo] {
	// Check if message is ready for delivery
	if !message.IsDeliverable() {
		return common.Err[MessageDeliveryInfo](errors.New("message is not ready for delivery"))
	}

	// Check if user has email notifications enabled
	if !recipient.IsEmailNotificationsEnabled() && message.DeliveryMethod() == DeliveryEmail {
		return common.Err[MessageDeliveryInfo](errors.New("user has email notifications disabled"))
	}

	// Get effective recipient email
	recipientEmail := recipient.GetEffectiveEmail()
	if recipientEmail == "" {
		return common.Err[MessageDeliveryInfo](errors.New("no valid recipient email"))
	}

	// Create delivery info
	deliveryInfo := MessageDeliveryInfo{
		Message:        message,
		RecipientEmail: recipientEmail,
		DeliveryMethod: message.DeliveryMethod(),
		Subject:        generateEmailSubject(message),
		Body:           generateEmailBody(message, recipient.User()),
		ScheduledTime:  message.DeliveryDate(),
		ProcessedAt:    time.Now(),
	}

	return common.Ok(deliveryInfo)
}

// MessageDeliveryInfo contains information needed for message delivery
type MessageDeliveryInfo struct {
	Message        Message
	RecipientEmail string
	DeliveryMethod DeliveryMethod
	Subject        string
	Body           string
	ScheduledTime  time.Time
	ProcessedAt    time.Time
}

// Getters for MessageDeliveryInfo
func (mdi MessageDeliveryInfo) GetMessage() Message {
	return mdi.Message
}

func (mdi MessageDeliveryInfo) GetRecipientEmail() string {
	return mdi.RecipientEmail
}

func (mdi MessageDeliveryInfo) GetDeliveryMethod() DeliveryMethod {
	return mdi.DeliveryMethod
}

func (mdi MessageDeliveryInfo) GetSubject() string {
	return mdi.Subject
}

func (mdi MessageDeliveryInfo) GetBody() string {
	return mdi.Body
}

func (mdi MessageDeliveryInfo) GetScheduledTime() time.Time {
	return mdi.ScheduledTime
}

func (mdi MessageDeliveryInfo) GetProcessedAt() time.Time {
	return mdi.ProcessedAt
}

// CalculateNextRetryTime calculates when to retry a failed delivery
func CalculateNextRetryTime(attemptCount int, lastAttempt time.Time) common.Result[time.Time] {
	if attemptCount <= 0 {
		return common.Err[time.Time](errors.New("attempt count must be positive"))
	}

	if attemptCount > 5 {
		return common.Err[time.Time](errors.New("maximum retry attempts exceeded"))
	}

	// Exponential backoff: 5 minutes, 15 minutes, 45 minutes, 2.25 hours, 6.75 hours
	backoffMinutes := []int{5, 15, 45, 135, 405}

	if attemptCount > len(backoffMinutes) {
		return common.Err[time.Time](errors.New("no more retry intervals available"))
	}

	retryDelay := time.Duration(backoffMinutes[attemptCount-1]) * time.Minute
	nextRetry := lastAttempt.Add(retryDelay)

	return common.Ok(nextRetry)
}

// FilterDueMessages filters messages that are due for delivery
func FilterDueMessages(messages []Message, currentTime time.Time) []Message {
	return common.FilterSlice(messages, func(m Message) bool {
		return m.IsDeliverable() && currentTime.After(m.DeliveryDate())
	})
}

// GroupMessagesByDeliveryMethod groups messages by their delivery method
func GroupMessagesByDeliveryMethod(messages []Message) map[DeliveryMethod][]Message {
	return common.GroupBySlice(messages, func(m Message) DeliveryMethod {
		return m.DeliveryMethod()
	})
}

// SortMessagesByDeliveryDate sorts messages by delivery date (earliest first)
func SortMessagesByDeliveryDate(messages []Message) []Message {
	// Simple bubble sort for demonstration (use sort.Slice in production)
	sorted := make([]Message, len(messages))
	copy(sorted, messages)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].DeliveryDate().After(sorted[j+1].DeliveryDate()) {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// BatchMessages groups messages into batches for processing
func BatchMessages(messages []Message, batchSize int) [][]Message {
	if batchSize <= 0 {
		return [][]Message{}
	}

	return common.ChunkSlice(messages, batchSize)
}

// ValidateMessageQuota checks if user can create more messages
func ValidateMessageQuota(userMessages []Message, quotaLimit int) common.Result[bool] {
	if quotaLimit <= 0 {
		return common.Ok(true) // No limit
	}

	// Count active (non-cancelled, non-delivered) messages
	activeMessages := common.FilterSlice(userMessages, func(m Message) bool {
		return m.Status() == StatusScheduled || m.Status() == StatusFailed
	})

	if len(activeMessages) >= quotaLimit {
		return common.Err[bool](errors.New("message quota exceeded"))
	}

	return common.Ok(true)
}

// CalculateDeliveryStats calculates statistics for message deliveries
func CalculateDeliveryStats(messages []Message) MessageStats {
	total := len(messages)

	scheduled := len(common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusScheduled
	}))

	delivered := len(common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusDelivered
	}))

	failed := len(common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusFailed
	}))

	cancelled := len(common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusCancelled
	}))

	var deliveryRate float64
	if total > 0 {
		deliveryRate = float64(delivered) / float64(total) * 100
	}

	return MessageStats{
		Total:        total,
		Scheduled:    scheduled,
		Delivered:    delivered,
		Failed:       failed,
		Cancelled:    cancelled,
		DeliveryRate: deliveryRate,
	}
}

// MessageStats contains statistics about messages
type MessageStats struct {
	Total        int
	Scheduled    int
	Delivered    int
	Failed       int
	Cancelled    int
	DeliveryRate float64
}

// GetUpcomingMessages returns messages that will be delivered within the specified duration
func GetUpcomingMessages(messages []Message, within time.Duration) []Message {
	cutoff := time.Now().Add(within)

	return common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusScheduled &&
			m.DeliveryDate().Before(cutoff) &&
			m.DeliveryDate().After(time.Now())
	})
}

// GetOverdueMessages returns messages that should have been delivered but weren't
func GetOverdueMessages(messages []Message) []Message {
	now := time.Now()

	return common.FilterSlice(messages, func(m Message) bool {
		return m.Status() == StatusScheduled && m.DeliveryDate().Before(now)
	})
}

// Pure helper functions for email generation

// generateEmailSubject creates an email subject for the message
func generateEmailSubject(message Message) string {
	if message.Title() != "" {
		return "Message from your past: " + message.Title()
	}
	return "You have a message from your past self"
}

// generateEmailBody creates an email body for the message
func generateEmailBody(message Message, sender user.User) string {
	body := "Hello " + sender.GetDisplayName() + ",\n\n"
	body += "You scheduled this message to be delivered to your future self.\n\n"

	if message.Title() != "" {
		body += "Subject: " + message.Title() + "\n\n"
	}

	body += "Message:\n"
	body += message.Content() + "\n\n"

	// Format the original send date
	originalDate := message.CreatedAt().Format("January 2, 2006 at 3:04 PM")
	body += "Originally written on: " + originalDate + "\n"

	// Format the delivery date
	deliveryDate := message.DeliveryDate().Format("January 2, 2006 at 3:04 PM")
	body += "Scheduled for delivery on: " + deliveryDate + "\n\n"

	body += "Best regards,\n"
	body += "Your Past Self\n\n"
	body += "---\n"
	body += "This message was delivered by Dear Future - Your Message to Tomorrow"

	return body
}

// Business rule functions

// CanUserEditMessage checks if a user can edit a specific message
func CanUserEditMessage(message Message, userID uuid.UUID) bool {
	return message.UserID() == userID && message.IsEditable()
}

// CanUserDeleteMessage checks if a user can delete a specific message
func CanUserDeleteMessage(message Message, userID uuid.UUID) bool {
	return message.UserID() == userID && message.IsDeletable()
}

// CanUserCancelMessage checks if a user can cancel a specific message
func CanUserCancelMessage(message Message, userID uuid.UUID) bool {
	return message.UserID() == userID && message.Status() == StatusScheduled
}

// GetMessageAccessLevel determines what level of access a user has to a message
func GetMessageAccessLevel(message Message, userID uuid.UUID) MessageAccessLevel {
	if message.UserID() != userID {
		return AccessNone
	}

	switch message.Status() {
	case StatusScheduled:
		return AccessFull
	case StatusDelivered:
		return AccessReadOnly
	case StatusFailed:
		return AccessRetryOnly
	case StatusCancelled:
		return AccessReadOnly
	default:
		return AccessNone
	}
}

// MessageAccessLevel represents the level of access a user has to a message
type MessageAccessLevel int

const (
	AccessNone      MessageAccessLevel = iota // No access
	AccessReadOnly                            // Can only view
	AccessRetryOnly                           // Can view and retry
	AccessFull                                // Can view, edit, delete, cancel
)

// String returns a string representation of the access level
func (mal MessageAccessLevel) String() string {
	switch mal {
	case AccessNone:
		return "none"
	case AccessReadOnly:
		return "read_only"
	case AccessRetryOnly:
		return "retry_only"
	case AccessFull:
		return "full"
	default:
		return "unknown"
	}
}

// CanRead checks if the access level allows reading
func (mal MessageAccessLevel) CanRead() bool {
	return mal >= AccessReadOnly
}

// CanEdit checks if the access level allows editing
func (mal MessageAccessLevel) CanEdit() bool {
	return mal == AccessFull
}

// CanDelete checks if the access level allows deletion
func (mal MessageAccessLevel) CanDelete() bool {
	return mal == AccessFull
}

// CanRetry checks if the access level allows retrying
func (mal MessageAccessLevel) CanRetry() bool {
	return mal == AccessRetryOnly || mal == AccessFull
}
