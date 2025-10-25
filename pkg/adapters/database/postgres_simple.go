// Package database provides simplified database adapter implementation
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/message"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/user"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// PostgresConfig holds PostgreSQL database configuration
type PostgresConfig struct {
	DatabaseURL  string
	MaxConns     int
	MaxIdleConns int
	ConnLifetime time.Duration
}

// SimplePostgresDB is a simplified PostgreSQL database adapter
type SimplePostgresDB struct {
	db *sql.DB
}

// NewSimplePostgresDB creates a new simplified PostgreSQL database adapter
func NewSimplePostgresDB(config PostgresConfig) (*SimplePostgresDB, error) {
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if config.MaxConns > 0 {
		db.SetMaxOpenConns(config.MaxConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnLifetime)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SimplePostgresDB{db: db}, nil
}

// Ping checks if the database is reachable
func (p *SimplePostgresDB) Ping(ctx context.Context) common.Result[bool] {
	if err := p.db.PingContext(ctx); err != nil {
		return common.Err[bool](fmt.Errorf("database ping failed: %w", err))
	}
	return common.Ok(true)
}

// Helper function to reconstruct User from database row
func userFromDB(id uuid.UUID, email, name, timezone string, createdAt, updatedAt time.Time) common.Result[user.User] {
	result := user.NewUser(user.CreateUserRequest{
		Email:    email,
		Name:     name,
		Timezone: timezone,
		UserID:   id,
	})
	return result
}

// SaveUser inserts or updates a user in the database
func (p *SimplePostgresDB) SaveUser(ctx context.Context, u user.User) common.Result[user.User] {
	query := `
		INSERT INTO user_profiles (id, email, name, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO UPDATE
		SET name = EXCLUDED.name,
			timezone = EXCLUDED.timezone,
			updated_at = EXCLUDED.updated_at
		RETURNING id, email, name, timezone, created_at, updated_at
	`

	var id uuid.UUID
	var email, name, timezone string
	var createdAt, updatedAt time.Time

	err := p.db.QueryRowContext(
		ctx,
		query,
		u.ID(),
		u.Email(),
		u.Name(),
		u.Timezone(),
		u.CreatedAt(),
		u.UpdatedAt(),
	).Scan(&id, &email, &name, &timezone, &createdAt, &updatedAt)

	if err != nil {
		return common.Err[user.User](fmt.Errorf("failed to save user: %w", err))
	}

	return userFromDB(id, email, name, timezone, createdAt, updatedAt)
}

// FindUserByID finds a user by ID
func (p *SimplePostgresDB) FindUserByID(ctx context.Context, userID uuid.UUID) common.Result[user.User] {
	query := `
		SELECT id, email, COALESCE(name, ''), COALESCE(timezone, 'UTC'), created_at, updated_at
		FROM user_profiles
		WHERE id = $1
	`

	var id uuid.UUID
	var email, name, timezone string
	var createdAt, updatedAt time.Time

	err := p.db.QueryRowContext(ctx, query, userID).Scan(&id, &email, &name, &timezone, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return common.Err[user.User](fmt.Errorf("user not found"))
	}
	if err != nil {
		return common.Err[user.User](fmt.Errorf("failed to find user: %w", err))
	}

	return userFromDB(id, email, name, timezone, createdAt, updatedAt)
}

// FindUserByEmail finds a user by email
func (p *SimplePostgresDB) FindUserByEmail(ctx context.Context, email string) common.Result[user.User] {
	query := `
		SELECT id, email, COALESCE(name, ''), COALESCE(timezone, 'UTC'), created_at, updated_at
		FROM user_profiles
		WHERE email = $1
	`

	var id uuid.UUID
	var dbEmail, name, timezone string
	var createdAt, updatedAt time.Time

	err := p.db.QueryRowContext(ctx, query, email).Scan(&id, &dbEmail, &name, &timezone, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return common.Err[user.User](fmt.Errorf("user not found"))
	}
	if err != nil {
		return common.Err[user.User](fmt.Errorf("failed to find user: %w", err))
	}

	fmt.Println("Found user:", dbEmail, id)

	return userFromDB(id, dbEmail, name, timezone, createdAt, updatedAt)
}

// UpdateUser updates an existing user
func (p *SimplePostgresDB) UpdateUser(ctx context.Context, u user.User) common.Result[user.User] {
	query := `
		UPDATE user_profiles
		SET name = $2, timezone = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, email, name, timezone, created_at, updated_at
	`

	var id uuid.UUID
	var email, name, timezone string
	var createdAt, updatedAt time.Time

	err := p.db.QueryRowContext(
		ctx,
		query,
		u.ID(),
		u.Name(),
		u.Timezone(),
		time.Now(),
	).Scan(&id, &email, &name, &timezone, &createdAt, &updatedAt)

	if err != nil {
		return common.Err[user.User](fmt.Errorf("failed to update user: %w", err))
	}

	return userFromDB(id, email, name, timezone, createdAt, updatedAt)
}

// DeleteUser deletes a user by ID
func (p *SimplePostgresDB) DeleteUser(ctx context.Context, userID uuid.UUID) common.Result[bool] {
	query := `DELETE FROM user_profiles WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, userID)
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to delete user: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to get rows affected: %w", err))
	}

	return common.Ok(rowsAffected > 0)
}

// SaveUserProfile - placeholder implementation
func (p *SimplePostgresDB) SaveUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile] {
	return common.Ok(profile)
}

// FindUserProfile - placeholder implementation
func (p *SimplePostgresDB) FindUserProfile(ctx context.Context, userID uuid.UUID) common.Result[user.UserProfile] {
	return common.Err[user.UserProfile](fmt.Errorf("not implemented yet"))
}

// UpdateUserProfile - placeholder implementation
func (p *SimplePostgresDB) UpdateUserProfile(ctx context.Context, profile user.UserProfile) common.Result[user.UserProfile] {
	return common.Ok(profile)
}

// Helper to reconstruct Message from database
func messageFromDB(id, userID uuid.UUID, title, content string, deliveryDate time.Time, timezone, status, deliveryMethod string, createdAt, updatedAt time.Time) common.Result[message.Message] {
	var msgStatus message.MessageStatus
	switch status {
	case "scheduled":
		msgStatus = message.StatusScheduled
	case "sent", "delivered":
		msgStatus = message.StatusDelivered
	case "failed":
		msgStatus = message.StatusFailed
	case "draft":
		msgStatus = message.StatusScheduled
	default:
		msgStatus = message.StatusScheduled
	}

	var method message.DeliveryMethod
	switch deliveryMethod {
	case "email":
		method = message.DeliveryEmail
	case "push":
		method = message.DeliveryPush
	default:
		method = message.DeliveryEmail
	}

	// Create message with time.Time delivery date
	result := message.NewMessage(message.CreateMessageRequest{
		UserID:         userID,
		Title:          title,
		Content:        content,
		DeliveryDate:   deliveryDate,
		Timezone:       timezone,
		DeliveryMethod: method,
	})

	if result.IsErr() {
		return result
	}

	// Update status if different from scheduled
	if msgStatus != message.StatusScheduled {
		statusResult := result.Value().WithStatus(msgStatus)
		if statusResult.IsErr() {
			// If status transition fails, just return original message
			return result
		}
		return statusResult
	}

	return result
}

// SaveMessage inserts a new message
func (p *SimplePostgresDB) SaveMessage(ctx context.Context, msg message.Message) common.Result[message.Message] {
	metadata := map[string]interface{}{
		"timezone":        msg.Timezone(),
		"delivery_method": string(msg.DeliveryMethod()),
	}
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO messages (id, user_id, subject, content, scheduled_for, status, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, subject, content, scheduled_for, status, created_at, updated_at
	`

	var id, userID uuid.UUID
	var title, content, status string
	var scheduledFor, createdAt, updatedAt time.Time
	fmt.Println(msg.UserID())
	err := p.db.QueryRowContext(
		ctx,
		query,
		msg.ID(),
		msg.UserID(),
		msg.Title(),
		msg.Content(),
		msg.DeliveryDate(),
		string(msg.Status()),
		msg.CreatedAt(),
		msg.UpdatedAt(),
		metadataJSON,
	).Scan(&id, &userID, &title, &content, &scheduledFor, &status, &createdAt, &updatedAt)

	if err != nil {
		return common.Err[message.Message](fmt.Errorf("failed to save message: %w", err))
	}

	return messageFromDB(id, userID, title, content, scheduledFor, msg.Timezone(), status, string(msg.DeliveryMethod()), createdAt, updatedAt)
}

// FindMessageByID finds a message by ID
func (p *SimplePostgresDB) FindMessageByID(ctx context.Context, messageID uuid.UUID) common.Result[message.Message] {
	query := `
		SELECT id, user_id, subject, content, scheduled_for, status, created_at, updated_at, COALESCE(metadata, '{}'::jsonb)
		FROM messages
		WHERE id = $1
	`

	var id, userID uuid.UUID
	var title, content, status string
	var scheduledFor, createdAt, updatedAt time.Time
	var metadataJSON []byte

	err := p.db.QueryRowContext(ctx, query, messageID).Scan(&id, &userID, &title, &content, &scheduledFor, &status, &createdAt, &updatedAt, &metadataJSON)
	if err == sql.ErrNoRows {
		return common.Err[message.Message](fmt.Errorf("message not found"))
	}
	if err != nil {
		return common.Err[message.Message](fmt.Errorf("failed to find message: %w", err))
	}

	var metadata map[string]interface{}
	json.Unmarshal(metadataJSON, &metadata)

	timezone := "UTC"
	deliveryMethod := "email"
	if tz, ok := metadata["timezone"].(string); ok {
		timezone = tz
	}
	if dm, ok := metadata["delivery_method"].(string); ok {
		deliveryMethod = dm
	}

	return messageFromDB(id, userID, title, content, scheduledFor, timezone, status, deliveryMethod, createdAt, updatedAt)
}

// FindMessagesByUserID finds all messages for a user
func (p *SimplePostgresDB) FindMessagesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) common.Result[[]message.Message] {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, user_id, subject, content, scheduled_for, status, created_at, updated_at, COALESCE(metadata, '{}'::jsonb)
		FROM messages
		WHERE user_id = $1
		ORDER BY scheduled_for DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return common.Err[[]message.Message](fmt.Errorf("failed to find messages: %w", err))
	}
	defer rows.Close()

	var messages []message.Message
	for rows.Next() {
		var id, uid uuid.UUID
		var title, content, status string
		var scheduledFor, createdAt, updatedAt time.Time
		var metadataJSON []byte

		err := rows.Scan(&id, &uid, &title, &content, &scheduledFor, &status, &createdAt, &updatedAt, &metadataJSON)
		if err != nil {
			return common.Err[[]message.Message](fmt.Errorf("failed to scan message: %w", err))
		}

		var metadata map[string]interface{}
		json.Unmarshal(metadataJSON, &metadata)

		timezone := "UTC"
		deliveryMethod := "email"
		if tz, ok := metadata["timezone"].(string); ok {
			timezone = tz
		}
		if dm, ok := metadata["delivery_method"].(string); ok {
			deliveryMethod = dm
		}

		msgResult := messageFromDB(id, uid, title, content, scheduledFor, timezone, status, deliveryMethod, createdAt, updatedAt)
		if msgResult.IsOk() {
			messages = append(messages, msgResult.Value())
		}
	}

	return common.Ok(messages)
}

// FindMessagesByStatus - placeholder implementation
func (p *SimplePostgresDB) FindMessagesByStatus(ctx context.Context, status message.MessageStatus, limit int) common.Result[[]message.Message] {
	return common.Ok([]message.Message{})
}

// FindDueMessages - placeholder implementation
func (p *SimplePostgresDB) FindDueMessages(ctx context.Context, before time.Time, limit int) common.Result[[]message.Message] {
	return common.Ok([]message.Message{})
}

// UpdateMessage updates an existing message
func (p *SimplePostgresDB) UpdateMessage(ctx context.Context, msg message.Message) common.Result[message.Message] {
	metadata := map[string]interface{}{
		"timezone":        msg.Timezone(),
		"delivery_method": string(msg.DeliveryMethod()),
	}
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		UPDATE messages
		SET subject = $2, content = $3, scheduled_for = $4, status = $5, updated_at = $6, metadata = $7
		WHERE id = $1
		RETURNING id, user_id, subject, content, scheduled_for, status, created_at, updated_at
	`

	var id, userID uuid.UUID
	var title, content, status string
	var scheduledFor, createdAt, updatedAt time.Time

	err := p.db.QueryRowContext(
		ctx,
		query,
		msg.ID(),
		msg.Title(),
		msg.Content(),
		msg.DeliveryDate(),
		string(msg.Status()),
		time.Now(),
		metadataJSON,
	).Scan(&id, &userID, &title, &content, &scheduledFor, &status, &createdAt, &updatedAt)

	if err != nil {
		return common.Err[message.Message](fmt.Errorf("failed to update message: %w", err))
	}

	return messageFromDB(id, userID, title, content, scheduledFor, msg.Timezone(), status, string(msg.DeliveryMethod()), createdAt, updatedAt)
}

// DeleteMessage deletes a message by ID
func (p *SimplePostgresDB) DeleteMessage(ctx context.Context, messageID uuid.UUID) common.Result[bool] {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to delete message: %w", err))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to get rows affected: %w", err))
	}

	return common.Ok(rowsAffected > 0)
}

// SaveMessageAttachment - placeholder implementation
func (p *SimplePostgresDB) SaveMessageAttachment(ctx context.Context, attachment message.MessageAttachment) common.Result[message.MessageAttachment] {
	return common.Ok(attachment)
}

// FindAttachmentsByMessageID - placeholder implementation
func (p *SimplePostgresDB) FindAttachmentsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]message.MessageAttachment] {
	return common.Ok([]message.MessageAttachment{})
}

// DeleteAttachment - placeholder implementation
func (p *SimplePostgresDB) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) common.Result[bool] {
	return common.Ok(true)
}

// SaveDeliveryLog - placeholder implementation
func (p *SimplePostgresDB) SaveDeliveryLog(ctx context.Context, log effects.DeliveryLog) common.Result[effects.DeliveryLog] {
	return common.Ok(log)
}

// FindDeliveryLogsByMessageID - placeholder implementation
func (p *SimplePostgresDB) FindDeliveryLogsByMessageID(ctx context.Context, messageID uuid.UUID) common.Result[[]effects.DeliveryLog] {
	return common.Ok([]effects.DeliveryLog{})
}

// Close closes the database connection
func (p *SimplePostgresDB) Close() error {
	return p.db.Close()
}
