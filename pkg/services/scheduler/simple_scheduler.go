package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// SimpleScheduler provides a basic in-process message scheduling engine.
type SimpleScheduler struct {
	db       effects.Database
	email    effects.EmailService
	cfg      *config.Config
	interval time.Duration

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewSimpleScheduler creates a scheduler that polls the database for due messages.
func NewSimpleScheduler(db effects.Database, email effects.EmailService, cfg *config.Config) *SimpleScheduler {
	interval := time.Minute
	if cfg != nil && cfg.SchedulerInterval > 0 {
		interval = cfg.SchedulerInterval
	}

	return &SimpleScheduler{
		db:       db,
		email:    email,
		cfg:      cfg,
		interval: interval,
	}
}

// ScheduleMessage updates a message's delivery time and re-queues it.
func (s *SimpleScheduler) ScheduleMessage(ctx context.Context, messageID uuid.UUID, deliveryTime time.Time) common.Result[effects.ScheduleResult] {
	msgResult := s.db.FindMessageByID(ctx, messageID)
	if msgResult.IsErr() {
		return common.Err[effects.ScheduleResult](msgResult.Error())
	}

	msg := msgResult.Value()

	updatedMsgResult := msg.WithDeliveryDate(deliveryTime, msg.Timezone())
	if updatedMsgResult.IsErr() {
		return common.Err[effects.ScheduleResult](updatedMsgResult.Error())
	}

	updatedMsg := updatedMsgResult.Value()
	if msg.Status() != message.StatusScheduled {
		statusResult := updatedMsg.WithStatus(message.StatusScheduled)
		if statusResult.IsErr() {
			return common.Err[effects.ScheduleResult](statusResult.Error())
		}
		updatedMsg = statusResult.Value()
	}

	saveResult := s.db.UpdateMessage(ctx, updatedMsg)
	if saveResult.IsErr() {
		return common.Err[effects.ScheduleResult](saveResult.Error())
	}

	return common.Ok(effects.ScheduleResult{
		MessageID:    messageID,
		ScheduledFor: deliveryTime,
		ScheduleID:   messageID.String(),
		Status:       effects.ScheduleStatusActive,
	})
}

// CancelScheduledMessage marks a scheduled message as cancelled.
func (s *SimpleScheduler) CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool] {
	msgResult := s.db.FindMessageByID(ctx, messageID)
	if msgResult.IsErr() {
		return common.Err[bool](msgResult.Error())
	}

	statusResult := msgResult.Value().WithStatus(message.StatusCancelled)
	if statusResult.IsErr() {
		return common.Err[bool](statusResult.Error())
	}

	saveResult := s.db.UpdateMessage(ctx, statusResult.Value())
	if saveResult.IsErr() {
		return common.Err[bool](saveResult.Error())
	}

	return common.Ok(true)
}

// RescheduleMessage changes an existing schedule to a new delivery time.
func (s *SimpleScheduler) RescheduleMessage(ctx context.Context, messageID uuid.UUID, newDeliveryTime time.Time) common.Result[effects.ScheduleResult] {
	return s.ScheduleMessage(ctx, messageID, newDeliveryTime)
}

// GetScheduledMessages returns scheduled messages within the provided window.
func (s *SimpleScheduler) GetScheduledMessages(ctx context.Context, from, to time.Time) common.Result[[]effects.ScheduledMessage] {
	messagesResult := s.db.FindMessagesByStatus(ctx, message.StatusScheduled, 200)
	if messagesResult.IsErr() {
		return common.Err[[]effects.ScheduledMessage](messagesResult.Error())
	}

	var scheduled []effects.ScheduledMessage
	for _, msg := range messagesResult.Value() {
		if msg.DeliveryDate().Before(from) || msg.DeliveryDate().After(to) {
			continue
		}

		scheduled = append(scheduled, effects.ScheduledMessage{
			MessageID:    msg.ID(),
			ScheduledFor: msg.DeliveryDate(),
			ScheduleID:   msg.ID().String(),
			Status:       effects.ScheduleStatusActive,
			CreatedAt:    msg.CreatedAt(),
		})
	}

	return common.Ok(scheduled)
}

// Start begins the background polling loop.
func (s *SimpleScheduler) Start(ctx context.Context) common.Result[bool] {
	if s == nil || s.db == nil {
		return common.Err[bool](errors.New("scheduler not configured"))
	}

	if s.interval <= 0 {
		s.interval = time.Minute
	}

	s.stop = make(chan struct{})
	s.wg.Add(1)

	go s.run(ctx)

	return common.Ok(true)
}

// Stop terminates the background loop.
func (s *SimpleScheduler) Stop(ctx context.Context) common.Result[bool] {
	if s.stop != nil {
		close(s.stop)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return common.Ok(true)
	case <-ctx.Done():
		return common.Err[bool](ctx.Err())
	}
}

func (s *SimpleScheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case <-ticker.C:
			s.executeCycle(ctx)
		}
	}
}

func (s *SimpleScheduler) executeCycle(ctx context.Context) {
	dueResult := s.db.FindDueMessages(ctx, time.Now(), 100)
	if dueResult.IsErr() {
		slog.Error("scheduler: failed to load due messages", "error", dueResult.Error())
		return
	}

	for _, msg := range dueResult.Value() {
		s.processMessage(ctx, msg)
	}
}

func (s *SimpleScheduler) processMessage(ctx context.Context, msg message.Message) {
	if s.email == nil {
		slog.Warn("scheduler: email service not configured, skipping delivery", "message_id", msg.ID())
		return
	}

	userResult := s.db.FindUserByID(ctx, msg.UserID())
	if userResult.IsErr() {
		slog.Error("scheduler: failed to load user", "message_id", msg.ID(), "error", userResult.Error())
		return
	}

	profile := user.NewUserProfile(userResult.Value())
	deliveryInfoResult := message.ProcessMessageDelivery(msg, profile)
	if deliveryInfoResult.IsErr() {
		s.failMessage(ctx, msg, deliveryInfoResult.Error())
		return
	}

	emailResult := s.email.SendMessage(ctx, deliveryInfoResult.Value())
	if emailResult.IsErr() {
		s.failMessage(ctx, msg, emailResult.Error())
		return
	}

	result := emailResult.Value()
	if result.Status != effects.EmailStatusSent {
		s.failMessage(ctx, msg, errors.New("email not sent"))
		return
	}

	s.completeMessage(ctx, msg)
}

func (s *SimpleScheduler) failMessage(ctx context.Context, msg message.Message, err error) {
	slog.Error("scheduler: delivery failed", "message_id", msg.ID(), "error", err)

	statusResult := msg.WithStatus(message.StatusFailed)
	if statusResult.IsErr() {
		slog.Error("scheduler: failed to update message status", "message_id", msg.ID(), "error", statusResult.Error())
		return
	}

	saveResult := s.db.UpdateMessage(ctx, statusResult.Value())
	if saveResult.IsErr() {
		slog.Error("scheduler: failed to persist message status", "message_id", msg.ID(), "error", saveResult.Error())
	}
}

func (s *SimpleScheduler) completeMessage(ctx context.Context, msg message.Message) {
	var nextMessage message.Message
	var err error

	if msg.HasRecurrence() {
		nextMessage, err = s.prepareNextOccurrence(msg)
		if err != nil {
			slog.Error("scheduler: failed to prepare recurring message", "message_id", msg.ID(), "error", err)
			return
		}
	} else {
		statusResult := msg.WithStatus(message.StatusDelivered)
		if statusResult.IsErr() {
			slog.Error("scheduler: failed to mark message as delivered", "message_id", msg.ID(), "error", statusResult.Error())
			return
		}
		nextMessage = statusResult.Value()
	}

	saveResult := s.db.UpdateMessage(ctx, nextMessage)
	if saveResult.IsErr() {
		slog.Error("scheduler: failed to persist message updates", "message_id", msg.ID(), "error", saveResult.Error())
	}
}

func (s *SimpleScheduler) prepareNextOccurrence(msg message.Message) (message.Message, error) {
	nextTimeResult := msg.NextRecurrenceTime()
	if nextTimeResult.IsErr() {
		return msg, nextTimeResult.Error()
	}

	nextDelivery := nextTimeResult.Value()
	updatedResult := msg.WithDeliveryDate(nextDelivery, msg.Timezone())
	if updatedResult.IsErr() {
		return msg, updatedResult.Error()
	}

	// Set status back to scheduled for the next occurrence
	statusResult := updatedResult.Value().WithStatus(message.StatusScheduled)
	if statusResult.IsErr() {
		return msg, statusResult.Error()
	}

	return statusResult.Value(), nil
}
