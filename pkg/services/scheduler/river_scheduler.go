package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/thanhphuchuynh/dear-future/pkg/config"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// RiverScheduler provides a River-based message scheduling engine
type RiverScheduler struct {
	client *river.Client[pgx.Tx]
	db     effects.Database
	email  effects.EmailService
	cfg    *config.Config
}

// DeliverMessageArgs are the arguments for the message delivery job
type DeliverMessageArgs struct {
	MessageID uuid.UUID `json:"message_id"`
}

// Kind returns the unique name for this job type
func (DeliverMessageArgs) Kind() string {
	return "deliver_message"
}

// DeliverMessageWorker processes message delivery jobs
type DeliverMessageWorker struct {
	river.WorkerDefaults[DeliverMessageArgs]
	db     effects.Database
	email  effects.EmailService
	client *river.Client[pgx.Tx]
}

// Work processes a single message delivery job
func (w *DeliverMessageWorker) Work(ctx context.Context, job *river.Job[DeliverMessageArgs]) error {
	slog.Info("river: processing message delivery", "message_id", job.Args.MessageID)

	// Load the message
	msgResult := w.db.FindMessageByID(ctx, job.Args.MessageID)
	if msgResult.IsErr() {
		slog.Error("river: failed to load message", "message_id", job.Args.MessageID, "error", msgResult.Error())
		return msgResult.Error()
	}

	msg := msgResult.Value()

	// Check if message is still scheduled
	if msg.Status() != message.StatusScheduled {
		slog.Info("river: message no longer scheduled, skipping", "message_id", job.Args.MessageID, "status", msg.Status())
		return nil
	}

	// Load user
	userResult := w.db.FindUserByID(ctx, msg.UserID())
	if userResult.IsErr() {
		slog.Error("river: failed to load user", "message_id", job.Args.MessageID, "error", userResult.Error())
		return userResult.Error()
	}

	// Prepare delivery info
	profile := user.NewUserProfile(userResult.Value())
	deliveryInfoResult := message.ProcessMessageDelivery(msg, profile)
	if deliveryInfoResult.IsErr() {
		return w.failMessage(ctx, msg, deliveryInfoResult.Error())
	}

	// Send email
	emailResult := w.email.SendMessage(ctx, deliveryInfoResult.Value())
	if emailResult.IsErr() {
		return w.failMessage(ctx, msg, emailResult.Error())
	}

	result := emailResult.Value()
	slog.Info("river: email sent", "message_id", job.Args.MessageID, "status", result.Status)
	if result.Status != effects.EmailStatusSent {
		return w.failMessage(ctx, msg, errors.New("email not sent"))
	}

	// Handle completion
	return w.completeMessage(ctx, msg)
}

func (w *DeliverMessageWorker) failMessage(ctx context.Context, msg message.Message, err error) error {
	slog.Error("river: delivery failed", "message_id", msg.ID(), "error", err)

	statusResult := msg.WithStatus(message.StatusFailed)
	if statusResult.IsErr() {
		slog.Error("river: failed to update message status", "message_id", msg.ID(), "error", statusResult.Error())
		return statusResult.Error()
	}

	saveResult := w.db.UpdateMessage(ctx, statusResult.Value())
	if saveResult.IsErr() {
		slog.Error("river: failed to persist message status", "message_id", msg.ID(), "error", saveResult.Error())
		return saveResult.Error()
	}

	return err
}

func (w *DeliverMessageWorker) completeMessage(ctx context.Context, msg message.Message) error {
	if msg.HasRecurrence() {
		// For recurring messages, keep status as scheduled and update delivery date
		nextMessage, err := w.prepareNextOccurrence(msg)
		if err != nil {
			slog.Error("river: failed to prepare recurring message", "message_id", msg.ID(), "error", err)
			return err
		}

		saveResult := w.db.UpdateMessage(ctx, nextMessage)
		if saveResult.IsErr() {
			slog.Error("river: failed to persist message updates", "message_id", msg.ID(), "error", saveResult.Error())
			return saveResult.Error()
		}

		// Schedule the next occurrence job
		if w.client != nil {
			jobArgs := DeliverMessageArgs{MessageID: msg.ID()}
			_, err := w.client.Insert(ctx, jobArgs, &river.InsertOpts{
				ScheduledAt: nextMessage.DeliveryDate(),
				Queue:       river.QueueDefault,
				MaxAttempts: 5,
				UniqueOpts: river.UniqueOpts{
					ByArgs: true,
				},
			})
			if err != nil {
				slog.Error("river: failed to schedule next occurrence", "message_id", msg.ID(), "error", err)
				return err
			}
		}

		slog.Info("river: recurring message delivered, next occurrence scheduled",
			"message_id", msg.ID(),
			"next_delivery", nextMessage.DeliveryDate())
	} else {
		// For one-time messages, mark as delivered
		statusResult := msg.WithStatus(message.StatusDelivered)
		if statusResult.IsErr() {
			slog.Error("river: failed to mark message as delivered", "message_id", msg.ID(), "error", statusResult.Error())
			return statusResult.Error()
		}

		saveResult := w.db.UpdateMessage(ctx, statusResult.Value())
		if saveResult.IsErr() {
			slog.Error("river: failed to persist message updates", "message_id", msg.ID(), "error", saveResult.Error())
			return saveResult.Error()
		}

		slog.Info("river: message delivered successfully", "message_id", msg.ID())
	}

	return nil
}

func (w *DeliverMessageWorker) prepareNextOccurrence(msg message.Message) (message.Message, error) {
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

// NewRiverScheduler creates a new River-based scheduler
func NewRiverScheduler(pool *pgxpool.Pool, db effects.Database, email effects.EmailService, cfg *config.Config) (*RiverScheduler, error) {
	// Get configuration values
	maxWorkers := 10
	queueName := river.QueueDefault
	fetchPollInterval := 1 * time.Second

	if cfg != nil {
		if cfg.Scheduling.River.MaxWorkers > 0 {
			maxWorkers = cfg.Scheduling.River.MaxWorkers
		}
		if cfg.Scheduling.River.QueueName != "" {
			queueName = cfg.Scheduling.River.QueueName
		}
		if cfg.Scheduling.River.PollInterval != "" {
			if duration, err := time.ParseDuration(cfg.Scheduling.River.PollInterval); err == nil {
				fetchPollInterval = duration
			}
		}
	}

	// Register the message delivery worker
	workers := river.NewWorkers()

	// Create River client with workers
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			queueName: {
				MaxWorkers:        maxWorkers,
				FetchPollInterval: fetchPollInterval,
			},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	// Add worker with client reference after client is created
	river.AddWorker(workers, &DeliverMessageWorker{
		db:     db,
		email:  email,
		client: riverClient,
	})

	slog.Info("river: scheduler configured",
		"max_workers", maxWorkers,
		"queue", queueName,
		"fetch_poll_interval", fetchPollInterval)

	return &RiverScheduler{
		client: riverClient,
		db:     db,
		email:  email,
		cfg:    cfg,
	}, nil
}

// ScheduleMessage schedules a message for delivery using River
func (s *RiverScheduler) ScheduleMessage(ctx context.Context, messageID uuid.UUID, deliveryTime time.Time) common.Result[effects.ScheduleResult] {
	msgResult := s.db.FindMessageByID(ctx, messageID)
	if msgResult.IsErr() {
		return common.Err[effects.ScheduleResult](msgResult.Error())
	}

	msg := msgResult.Value()

	// Update message delivery time and status
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

	// Get max attempts from config
	maxAttempts := 5
	queueName := river.QueueDefault

	if s.cfg != nil {
		if s.cfg.Scheduling.River.MaxAttempts > 0 {
			maxAttempts = s.cfg.Scheduling.River.MaxAttempts
		}
		if s.cfg.Scheduling.River.QueueName != "" {
			queueName = s.cfg.Scheduling.River.QueueName
		}
	}

	// Schedule the job with River using unique key to prevent duplicates
	jobArgs := DeliverMessageArgs{MessageID: messageID}
	_, err := s.client.Insert(ctx, jobArgs, &river.InsertOpts{
		ScheduledAt: deliveryTime,
		Queue:       queueName,
		MaxAttempts: maxAttempts,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true, // Unique by message ID
		},
	})
	if err != nil {
		return common.Err[effects.ScheduleResult](err)
	}

	return common.Ok(effects.ScheduleResult{
		MessageID:    messageID,
		ScheduledFor: deliveryTime,
		ScheduleID:   messageID.String(),
		Status:       effects.ScheduleStatusActive,
	})
}

// CancelScheduledMessage marks a scheduled message as cancelled
func (s *RiverScheduler) CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool] {
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

	// Note: River jobs are handled by checking message status in the worker
	// Cancelled messages will be skipped when the job runs

	return common.Ok(true)
}

// RescheduleMessage changes an existing schedule to a new delivery time
func (s *RiverScheduler) RescheduleMessage(ctx context.Context, messageID uuid.UUID, newDeliveryTime time.Time) common.Result[effects.ScheduleResult] {
	return s.ScheduleMessage(ctx, messageID, newDeliveryTime)
}

// GetScheduledMessages returns scheduled messages within the provided window
func (s *RiverScheduler) GetScheduledMessages(ctx context.Context, from, to time.Time) common.Result[[]effects.ScheduledMessage] {
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

// Start begins the River worker
func (s *RiverScheduler) Start(ctx context.Context) common.Result[bool] {
	if s.client == nil {
		return common.Err[bool](errors.New("river client not configured"))
	}

	if err := s.client.Start(ctx); err != nil {
		return common.Err[bool](err)
	}

	slog.Info("river: scheduler started")
	return common.Ok(true)
}

// Stop terminates the River worker
func (s *RiverScheduler) Stop(ctx context.Context) common.Result[bool] {
	if s.client == nil {
		return common.Ok(true)
	}

	if err := s.client.Stop(ctx); err != nil {
		return common.Err[bool](err)
	}

	slog.Info("river: scheduler stopped")
	return common.Ok(true)
}
