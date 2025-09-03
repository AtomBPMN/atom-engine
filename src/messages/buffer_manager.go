/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package messages

import (
	"context"
	"fmt"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// BufferManager manages message buffering
type BufferManager struct {
	storage        storage.Storage
	logger         logger.ComponentLogger
	correlationMgr *CorrelationManager
	isRunning      bool
	stopChan       chan struct{}
}

// NewBufferManager creates new buffer manager
func NewBufferManager(storage storage.Storage, logger logger.ComponentLogger) *BufferManager {
	return &BufferManager{
		storage:  storage,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// SetCorrelationManager sets correlation manager reference for processing
func (bm *BufferManager) SetCorrelationManager(cm *CorrelationManager) {
	bm.correlationMgr = cm
}

// Start starts the buffer manager
func (bm *BufferManager) Start() error {
	bm.logger.Info("Starting buffer manager")
	bm.isRunning = true

	// Start cleanup goroutine
	go bm.cleanupExpiredMessages()

	bm.logger.Info("Buffer manager started")
	return nil
}

// Stop stops the buffer manager
func (bm *BufferManager) Stop() {
	bm.logger.Info("Stopping buffer manager")
	bm.isRunning = false
	close(bm.stopChan)
	bm.logger.Info("Buffer manager stopped")
}

// BufferMessage buffers a message
func (bm *BufferManager) BufferMessage(ctx context.Context, message *models.BufferedMessage) error {
	bm.logger.Info("Buffering message", logger.String("name", message.Name), logger.String("reason", message.Reason))

	if err := bm.storage.SaveBufferedMessage(ctx, message); err != nil {
		return fmt.Errorf("failed to buffer message: %w", err)
	}

	bm.logger.Info("Message buffered successfully")
	return nil
}

// ListBufferedMessages lists buffered messages
func (bm *BufferManager) ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error) {
	bm.logger.Debug("Listing buffered messages", logger.Int("limit", limit), logger.Int("offset", offset))

	messages, err := bm.storage.ListBufferedMessages(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	// Apply offset and limit
	start := offset
	if start > len(messages) {
		start = len(messages)
	}

	end := start + limit
	if end > len(messages) {
		end = len(messages)
	}

	result := messages[start:end]
	bm.logger.Debug("Listed buffered messages", logger.Int("returned", len(result)))

	return result, nil
}

// GetBufferedMessage gets buffered message by ID
func (bm *BufferManager) GetBufferedMessage(ctx context.Context, messageID string) (*models.BufferedMessage, error) {
	bm.logger.Debug("Getting buffered message", logger.String("messageID", messageID))

	message, err := bm.storage.GetBufferedMessage(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buffered message: %w", err)
	}

	if message != nil {
		bm.logger.Debug("Buffered message found", logger.String("name", message.Name))
	} else {
		bm.logger.Debug("Buffered message not found")
	}

	return message, nil
}

// DeleteBufferedMessage deletes buffered message
func (bm *BufferManager) DeleteBufferedMessage(ctx context.Context, messageID string) error {
	bm.logger.Info("Deleting buffered message", logger.String("messageID", messageID))

	if err := bm.storage.DeleteBufferedMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete buffered message: %w", err)
	}

	bm.logger.Info("Buffered message deleted")
	return nil
}

// CleanupExpiredMessages cleans up expired buffered messages
func (bm *BufferManager) CleanupExpiredMessages(ctx context.Context) (int, error) {
	bm.logger.Info("Cleaning up expired buffered messages")

	messages, err := bm.storage.ListBufferedMessages(ctx, "", 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	cleanedCount := 0
	for _, message := range messages {
		if message.IsExpired() {
			if err := bm.storage.DeleteBufferedMessage(ctx, message.ID); err != nil {
				bm.logger.Error("Failed to delete expired message", logger.String("error", err.Error()))
				continue
			}
			cleanedCount++
			bm.logger.Debug("Deleted expired message", logger.String("name", message.Name))
		}
	}

	bm.logger.Info("Expired messages cleaned up", logger.Int("cleanedCount", cleanedCount))
	return cleanedCount, nil
}

// ProcessBufferedMessages processes buffered messages against new subscriptions
func (bm *BufferManager) ProcessBufferedMessages(ctx context.Context, subscription *models.ProcessMessageSubscription) (int, error) {
	bm.logger.Info("Processing buffered messages for subscription", logger.String("messageName", subscription.MessageName))

	messages, err := bm.storage.ListBufferedMessages(ctx, subscription.TenantID, 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	processedCount := 0
	for _, message := range messages {
		// Check if message matches subscription
		if message.Name != subscription.MessageName {
			continue
		}

		// Check correlation key match if specified
		if subscription.CorrelationKey != "" && message.CorrelationKey != subscription.CorrelationKey {
			continue
		}

		// Check if message is expired
		if message.IsExpired() {
			continue
		}

		// For intermediate catch events, trigger message correlation through correlation manager
		// Для intermediate catch events запускаем корреляцию сообщений через correlation manager
		if bm.correlationMgr != nil {
			correlationResult, err := bm.correlationMgr.PublishMessage(ctx, message.TenantID, message.Name, message.CorrelationKey, message.Variables, nil)
			if err != nil {
				bm.logger.Error("Failed to correlate buffered message",
					logger.String("message_id", message.ID),
					logger.String("error", err.Error()))
				continue
			}
			bm.logger.Info("Buffered message correlated successfully",
				logger.String("message_id", message.ID),
				logger.String("correlation_result_id", correlationResult.ID))
		} else {
			bm.logger.Warn("Correlation manager not available for buffered message processing",
				logger.String("message_id", message.ID))
		}

		// Delete processed message from buffer
		if err := bm.storage.DeleteBufferedMessage(ctx, message.ID); err != nil {
			bm.logger.Error("Failed to delete processed message", logger.String("error", err.Error()))
			continue
		}

		processedCount++
	}

	if processedCount > 0 {
		bm.logger.Info("Processed buffered messages", logger.String("subscriptionID", subscription.ID), logger.Int("processedCount", processedCount))
	}

	return processedCount, nil
}

// GetBufferedMessagesByName gets buffered messages by name
func (bm *BufferManager) GetBufferedMessagesByName(ctx context.Context, tenantID, messageName string) ([]*models.BufferedMessage, error) {
	bm.logger.Debug("Getting buffered messages by name", logger.String("messageName", messageName))

	messages, err := bm.storage.ListBufferedMessages(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	var matchingMessages []*models.BufferedMessage
	for _, message := range messages {
		if message.Name == messageName && !message.IsExpired() {
			matchingMessages = append(matchingMessages, message)
		}
	}

	bm.logger.Debug("Found buffered messages by name", logger.String("messageName", messageName), logger.Int("count", len(messages)))
	return matchingMessages, nil
}

// GetBufferedMessagesByCorrelationKey gets buffered messages by correlation key
func (bm *BufferManager) GetBufferedMessagesByCorrelationKey(ctx context.Context, tenantID, correlationKey string) ([]*models.BufferedMessage, error) {
	bm.logger.Debug("Getting buffered messages by correlation key", logger.String("correlationKey", correlationKey))

	messages, err := bm.storage.ListBufferedMessages(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	var matchingMessages []*models.BufferedMessage
	for _, message := range messages {
		if message.CorrelationKey == correlationKey && !message.IsExpired() {
			matchingMessages = append(matchingMessages, message)
		}
	}

	bm.logger.Debug("Found buffered messages by correlation key", logger.String("correlationKey", correlationKey), logger.Int("count", len(messages)))
	return matchingMessages, nil
}

// cleanupExpiredMessages runs periodic cleanup
func (bm *BufferManager) cleanupExpiredMessages() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if _, err := bm.CleanupExpiredMessages(ctx); err != nil {
				bm.logger.Error("Failed to cleanup expired messages", logger.String("error", err.Error()))
			}
			cancel()
		case <-bm.stopChan:
			return
		}
	}
}
