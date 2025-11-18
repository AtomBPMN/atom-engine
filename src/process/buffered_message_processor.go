/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"context"
	"fmt"
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/messages"
	"atom-engine/src/storage"
)

// BufferedMessageProcessor handles buffered message operations
// Обрабатывает операции с буферизованными сообщениями
type BufferedMessageProcessor struct {
	storage storage.Storage
	core    CoreInterface
}

// NewBufferedMessageProcessor creates new buffered message processor
// Создает новый обработчик буферизованных сообщений
func NewBufferedMessageProcessor(storage storage.Storage, core CoreInterface) *BufferedMessageProcessor {
	return &BufferedMessageProcessor{
		storage: storage,
		core:    core,
	}
}

// CheckBufferedMessages checks if there are buffered messages that match the criteria
// Проверяет есть ли буферизованные сообщения которые соответствуют критериям
func (bmp *BufferedMessageProcessor) CheckBufferedMessages(
	messageName, correlationKey string,
) (*models.BufferedMessage, error) {
	logger.Info("Checking buffered messages",
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey))

	// Get messages component to check buffer
	if bmp.core == nil {
		return nil, fmt.Errorf("core interface not available")
	}

	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return nil, fmt.Errorf("messages component not available")
	}

	// Type assert to get actual messages component
	msgComponent, ok := messagesComponent.(*messages.Component)
	if !ok {
		return nil, fmt.Errorf("invalid messages component type")
	}

	// List all buffered messages
	bufferedMessages, err := msgComponent.ListBufferedMessages(context.Background(), "", 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list buffered messages: %w", err)
	}

	// Find matching message
	return bmp.findMatchingBufferedMessage(bufferedMessages, messageName, correlationKey)
}

// findMatchingBufferedMessage finds buffered message that matches criteria
// Находит буферизованное сообщение которое соответствует критериям
func (bmp *BufferedMessageProcessor) findMatchingBufferedMessage(
	messages []*models.BufferedMessage,
	messageName, correlationKey string,
) (*models.BufferedMessage, error) {
	for _, message := range messages {
		// Check message name match
		if message.Name != messageName {
			continue
		}

		// Check correlation key match (empty correlation key matches any)
		if correlationKey != "" {
			// Note: FEEL expressions in correlation keys are now evaluated BEFORE calling this method
			// in intermediate_catch_message_handler.go:evaluateCorrelationKeyExpression()
			// Примечание: FEEL expressions в correlation keys теперь вычисляются ДО вызова этого метода
			// в intermediate_catch_message_handler.go:evaluateCorrelationKeyExpression()
			expectedKey := correlationKey
			if strings.HasPrefix(correlationKey, "=") {
				// This should not happen anymore, but keep fallback for safety
				// Это больше не должно происходить, но оставляем fallback для безопасности
				expectedKey = correlationKey[1:]
				logger.Warn("Unexpected FEEL expression in correlation key - should be pre-evaluated",
					logger.String("expression", correlationKey),
					logger.String("fallback_key", expectedKey))
			}

			// Special handling for buffered messages with empty correlation key
			// Старые буферизованные сообщения могут иметь пустой correlation key
			if message.CorrelationKey == "" || message.CorrelationKey == "<none>" {
				// For empty correlation key in buffered message, check if expectedKey matches message name
				// Для пустого correlation key в буферизованном сообщении проверяем совпадает ли expectedKey с именем сообщения
				if expectedKey == messageName {
					logger.Info("Found buffered message with empty correlation key matching expected key",
						logger.String("message_id", message.ID),
						logger.String("expected_key", expectedKey),
						logger.String("message_name", messageName))
					// Continue to other checks (expiry, etc.)
				} else {
					logger.Debug("Correlation key mismatch - empty correlation key but expectedKey doesn't match message name",
						logger.String("message_correlation_key", message.CorrelationKey),
						logger.String("expected_key", expectedKey),
						logger.String("message_name", messageName),
						logger.String("original_key", correlationKey))
					continue
				}
			} else if message.CorrelationKey != expectedKey {
				logger.Debug("Correlation key mismatch",
					logger.String("message_correlation_key", message.CorrelationKey),
					logger.String("expected_key", expectedKey),
					logger.String("original_key", correlationKey))
				continue
			}
		}

		// Check if message is not expired
		if message.IsExpired() {
			continue
		}

		logger.Info("Found matching buffered message",
			logger.String("message_id", message.ID),
			logger.String("message_name", message.Name),
			logger.String("correlation_key", message.CorrelationKey))

		return message, nil
	}

	logger.Info("No matching buffered message found",
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey))

	return nil, nil
}

// ProcessBufferedMessage processes a buffered message for a token
// Обрабатывает буферизованное сообщение для токена
func (bmp *BufferedMessageProcessor) ProcessBufferedMessage(
	message *models.BufferedMessage,
	token *models.Token,
) error {
	logger.Info("Processing buffered message for token",
		logger.String("message_id", message.ID),
		logger.String("token_id", token.TokenID))

	// Add message variables to token variables
	if message.Variables != nil {
		if token.Variables == nil {
			token.Variables = make(map[string]interface{})
		}

		// Add message data with a special key to mark it as correlated
		token.Variables["data"] = message.Variables
		token.Variables["_message_correlated"] = true
		token.Variables["_message_id"] = message.ID
		token.Variables["_correlation_key"] = message.CorrelationKey
	}

	// Update token in storage
	if err := bmp.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update token with message data: %w", err)
	}

	// Delete the processed message from buffer
	if err := bmp.deleteProcessedMessage(message); err != nil {
		// Don't fail the whole operation for this
		logger.Warn("Failed to delete processed buffered message", logger.String("error", err.Error()))
	}

	logger.Info("Buffered message processed successfully",
		logger.String("message_id", message.ID),
		logger.String("token_id", token.TokenID))

	return nil
}

// deleteProcessedMessage deletes processed message from buffer
// Удаляет обработанное сообщение из буфера
func (bmp *BufferedMessageProcessor) deleteProcessedMessage(message *models.BufferedMessage) error {
	if bmp.core == nil {
		return fmt.Errorf("core interface not available")
	}

	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return fmt.Errorf("messages component not available")
	}

	if msgComponent, ok := messagesComponent.(*messages.Component); ok {
		bufferManager := msgComponent.GetBufferManager()
		if err := bufferManager.DeleteBufferedMessage(context.Background(), message.ID); err != nil {
			return err
		}
		logger.Info("Buffered message deleted after processing",
			logger.String("message_id", message.ID))
	}

	return nil
}

// CreateMessageSubscription creates a message subscription
// Создает подписку на сообщение
func (bmp *BufferedMessageProcessor) CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error {
	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return fmt.Errorf("messages component not available")
	}

	// Cast to messages component and create subscription
	if messageComp, ok := messagesComponent.(*messages.Component); ok {
		ctx := context.Background()
		if err := messageComp.CreateMessageSubscription(ctx, subscription); err != nil {
			logger.Error("Failed to create message subscription",
				logger.String("error", err.Error()),
				logger.String("message_name", subscription.MessageName))
			return fmt.Errorf("failed to create message subscription: %w", err)
		}

		logger.Info("Message subscription created successfully",
			logger.String("message_name", subscription.MessageName),
			logger.String("subscription_id", subscription.ID))
		return nil
	}

	return fmt.Errorf("messages component type assertion failed")
}

// DeleteMessageSubscription deletes a message subscription
// Удаляет подписку на сообщение
func (bmp *BufferedMessageProcessor) DeleteMessageSubscription(subscriptionID string) error {
	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return fmt.Errorf("messages component not available")
	}

	// Cast to messages component and delete subscription
	if messageComp, ok := messagesComponent.(*messages.Component); ok {
		ctx := context.Background()
		if err := messageComp.DeleteMessageSubscription(ctx, subscriptionID); err != nil {
			logger.Error("Failed to delete message subscription",
				logger.String("error", err.Error()),
				logger.String("subscription_id", subscriptionID))
			return fmt.Errorf("failed to delete message subscription: %w", err)
		}

		logger.Info("Message subscription deleted successfully",
			logger.String("subscription_id", subscriptionID))
		return nil
	}

	return fmt.Errorf("messages component type assertion failed")
}

// PublishMessage publishes a message for correlation
// Публикует сообщение для корреляции
func (bmp *BufferedMessageProcessor) PublishMessage(
	messageName, correlationKey string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	return bmp.PublishMessageWithElementID(messageName, correlationKey, "", variables)
}

// PublishMessageWithElementID publishes message with element ID for correlation
// Публикует сообщение с element ID для корреляции
func (bmp *BufferedMessageProcessor) PublishMessageWithElementID(
	messageName, correlationKey, elementID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return nil, fmt.Errorf("messages component not available")
	}

	// Cast to messages component and publish message
	if messageComp, ok := messagesComponent.(*messages.Component); ok {
		ctx := context.Background()
		ttl := 300 * time.Second // Default TTL 5 minutes

		result, err := messageComp.PublishMessageWithElementID(
			ctx,
			"",
			messageName,
			correlationKey,
			elementID,
			variables,
			&ttl,
		)
		if err != nil {
			logger.Error("Failed to publish message",
				logger.String("error", err.Error()),
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey),
				logger.String("element_id", elementID))
			return nil, fmt.Errorf("failed to publish message: %w", err)
		}

		logger.Info("Message published successfully",
			logger.String("message_name", messageName),
			logger.String("correlation_key", correlationKey),
			logger.String("element_id", elementID),
			logger.Bool("instance_created", result.InstanceCreated))
		return result, nil
	}

	return nil, fmt.Errorf("messages component type assertion failed")
}

// CorrelateMessage correlates message with specific process instance
// Коррелирует сообщение с конкретным экземпляром процесса
func (bmp *BufferedMessageProcessor) CorrelateMessage(
	messageName, correlationKey, processInstanceID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	messagesComponent := bmp.core.GetMessagesComponent()
	if messagesComponent == nil {
		return nil, fmt.Errorf("messages component not available")
	}

	// Cast to messages component and correlate message
	if messageComp, ok := messagesComponent.(*messages.Component); ok {
		ctx := context.Background()

		result, err := messageComp.CorrelateMessage(ctx, "", messageName, correlationKey, processInstanceID, variables)
		if err != nil {
			logger.Error("Failed to correlate message",
				logger.String("error", err.Error()),
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey),
				logger.String("process_instance_id", processInstanceID))
			return nil, fmt.Errorf("failed to correlate message: %w", err)
		}

		logger.Info("Message correlated successfully",
			logger.String("message_name", messageName),
			logger.String("correlation_key", correlationKey),
			logger.String("process_instance_id", processInstanceID))
		return result, nil
	}

	return nil, fmt.Errorf("messages component type assertion failed")
}
