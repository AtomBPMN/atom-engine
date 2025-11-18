/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// CorrelationManager manages message correlation
type CorrelationManager struct {
	storage         storage.Storage
	logger          logger.ComponentLogger
	responseChannel chan string
	isRunning       bool
	stopChan        chan struct{}
}

// NewCorrelationManager creates new correlation manager
func NewCorrelationManager(
	storage storage.Storage,
	logger logger.ComponentLogger,
	responseChannel chan string,
) *CorrelationManager {
	return &CorrelationManager{
		storage:         storage,
		logger:          logger,
		responseChannel: responseChannel,
		stopChan:        make(chan struct{}),
	}
}

// Start starts the correlation manager
func (cm *CorrelationManager) Start() error {
	cm.logger.Info("Starting correlation manager")
	cm.isRunning = true

	// Start background cleanup
	go cm.cleanupExpiredData()

	cm.logger.Info("Correlation manager started")
	return nil
}

// Stop stops the correlation manager
func (cm *CorrelationManager) Stop() {
	cm.logger.Info("Stopping correlation manager")
	cm.isRunning = false
	close(cm.stopChan)
	cm.logger.Info("Correlation manager stopped")
}

// PublishMessage publishes a message for correlation
func (cm *CorrelationManager) PublishMessage(
	ctx context.Context,
	tenantID, messageName, correlationKey, elementID string,
	variables map[string]interface{},
	ttl *time.Duration,
) (*models.MessageCorrelationResult, error) {
	cm.logger.Info("Publishing message for correlation",
		logger.String("messageName", messageName),
		logger.String("correlationKey", correlationKey),
		logger.String("elementID", elementID),
	)

	// Create message ID
	messageID := models.GenerateID()

	// Try to find active subscription
	subscriptions, err := cm.storage.ListProcessMessageSubscriptions(ctx, tenantID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var targetSubscription *models.ProcessMessageSubscription
	for _, sub := range subscriptions {
		if sub.MessageName == messageName && sub.IsActive {
			// Check correlation key match if specified
			if correlationKey != "" && sub.CorrelationKey != "" {
				// Handle FEEL expressions in subscription correlation key
				// Обрабатываем FEEL выражения в correlation key подписки
				subscriptionKey := sub.CorrelationKey

				// If subscription correlation key starts with "=", it's a FEEL expression
				// Если correlation key подписки начинается с "=", это FEEL выражение
				if strings.HasPrefix(subscriptionKey, "=") {
					// For now, simple FEEL literal evaluation: ="value" or =value
					// Пока что простая оценка FEEL литералов: ="value" или =value
					feelExpression := strings.TrimPrefix(subscriptionKey, "=")

					// If FEEL expression is quoted string literal, remove quotes
					// Если FEEL выражение это строковый литерал в кавычках, убираем кавычки
					if strings.HasPrefix(feelExpression, "\"") && strings.HasSuffix(feelExpression, "\"") {
						subscriptionKey = strings.Trim(feelExpression, "\"")
					} else {
						// Treat as string literal without quotes
						// Рассматриваем как строковый литерал без кавычек
						subscriptionKey = feelExpression
					}

					cm.logger.Info("FEEL correlation key evaluated",
						logger.String("original", sub.CorrelationKey),
						logger.String("evaluated", subscriptionKey),
						logger.String("incoming", correlationKey))
				}

				if subscriptionKey != correlationKey {
					continue
				}
			}
			targetSubscription = sub
			break
		}
	}

	result := &models.MessageCorrelationResult{
		ID:              models.GenerateID(),
		MessageID:       messageID,
		TenantID:        tenantID,
		MessageName:     messageName,
		CorrelationKey:  correlationKey,
		Variables:       variables,
		CreatedAt:       time.Now(),
		InstanceCreated: false,
	}

	if targetSubscription != nil {
		// Check if this is intermediate catch event or start event
		// Проверяем является ли это intermediate catch event или start event
		isIntermediateCatchEvent := cm.isIntermediateCatchEvent(targetSubscription.StartEventID)

		if isIntermediateCatchEvent {
			// For intermediate catch events, find waiting token and activate it
			// Для intermediate catch events находим ожидающий токен и активируем его
			waitingToken, err := cm.findWaitingToken(targetSubscription.StartEventID, messageName)
			if err != nil {
				return nil, fmt.Errorf("failed to find waiting token: %w", err)
			}

			if waitingToken != nil {
				result.ProcessInstanceID = waitingToken.ProcessInstanceID
				result.InstanceCreated = false

				cm.logger.Info("Message correlated with waiting token",
					logger.String("token_id", waitingToken.TokenID),
					logger.String("process_instance_id", waitingToken.ProcessInstanceID),
					logger.String("subscriptionID", targetSubscription.ID))
			} else {
				cm.logger.Warn("No waiting token found for intermediate catch event",
					logger.String("element_id", targetSubscription.StartEventID),
					logger.String("message_name", messageName))
				return result, nil
			}
		} else {
			// For start events, create new process instance
			// Для start events создаем новый экземпляр процесса
			processInstanceID := models.GenerateID()

			// NOTE: Process instance creation should be integrated with process engine
			// For now, just set the ID
			result.ProcessInstanceID = processInstanceID
			result.InstanceCreated = true

			cm.logger.Info("Message correlated, process instance created",
				logger.String("processInstanceID", processInstanceID),
				logger.String("subscriptionID", targetSubscription.ID),
			)
		}

		// Send correlation callback if response channel is available
		// Отправляем correlation callback если канал ответов доступен
		if cm.responseChannel != nil {
			callback := map[string]interface{}{
				"event_type":          "correlation",
				"message_id":          messageID,
				"message_name":        messageName,
				"correlation_key":     correlationKey,
				"process_instance_id": result.ProcessInstanceID,
				"subscription_id":     targetSubscription.ID,
				"variables":           variables,
				"correlated_at":       time.Now().Format(time.RFC3339),
			}

			// For intermediate catch events, include token_id
			// Для intermediate catch events включаем token_id
			if isIntermediateCatchEvent {
				if waitingToken, err := cm.findWaitingToken(
					targetSubscription.StartEventID,
					messageName,
				); err == nil && waitingToken != nil {
					callback["token_id"] = waitingToken.TokenID
				}
			}

			if callbackJSON, err := json.Marshal(callback); err == nil {
				cm.logger.Info("Sending callback to response channel",
					logger.String("message_name", messageName),
					logger.String("callback_json", string(callbackJSON)))
				select {
				case cm.responseChannel <- string(callbackJSON):
					cm.logger.Info("Message correlation callback sent successfully",
						logger.String("message_name", messageName),
						logger.String("process_instance_id", result.ProcessInstanceID))
				default:
					cm.logger.Warn("Message response channel full, correlation callback dropped")
				}
			} else {
				cm.logger.Error("Failed to marshal callback JSON",
					logger.String("error", err.Error()))
			}
		}

		// Delete subscription after successful correlation for intermediate catch events
		// Удаляем подписку после успешной корреляции для intermediate catch events
		if isIntermediateCatchEvent {
			if err := cm.storage.DeleteProcessMessageSubscription(ctx, targetSubscription.ID); err != nil {
				cm.logger.Error("Failed to delete subscription after correlation",
					logger.String("subscription_id", targetSubscription.ID),
					logger.String("error", err.Error()))
			} else {
				cm.logger.Info("Subscription deleted after successful correlation",
					logger.String("subscription_id", targetSubscription.ID),
					logger.String("message_name", messageName))
			}
		}
	} else {
		// Buffer message if no active subscription found
		bufferedMessage := &models.BufferedMessage{
			ID:             messageID,
			TenantID:       tenantID,
			Name:           messageName,
			CorrelationKey: correlationKey,
			Variables:      variables,
			PublishedAt:    time.Now(),
			BufferedAt:     time.Now(),
			Reason:         "No active subscription found",
			ElementID:      elementID,
		}

		if ttl != nil {
			expiresAt := time.Now().Add(*ttl)
			bufferedMessage.ExpiresAt = &expiresAt
		}

		if err := cm.storage.SaveBufferedMessage(ctx, bufferedMessage); err != nil {
			cm.logger.Error("Failed to buffer message", logger.String("error", err.Error()))
			result.ErrorMessage = fmt.Sprintf("failed to buffer message: %v", err)
		} else {
			cm.logger.Info("Message buffered", logger.String("reason", bufferedMessage.Reason))
		}
	}

	// Save correlation result
	if err := cm.storage.SaveMessageCorrelationResult(ctx, result); err != nil {
		cm.logger.Error("Failed to save correlation result", logger.String("error", err.Error()))
	}

	return result, nil
}

// CorrelateMessage correlates message with specific process instance
func (cm *CorrelationManager) CorrelateMessage(
	ctx context.Context,
	tenantID, messageName, correlationKey, processInstanceID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	cm.logger.Info("Correlating message with process instance",
		logger.String("correlationKey", correlationKey),
		logger.String("processInstanceID", processInstanceID),
	)

	messageID := models.GenerateID()

	result := &models.MessageCorrelationResult{
		ID:                messageID,
		MessageID:         messageID,
		TenantID:          tenantID,
		MessageName:       messageName,
		CorrelationKey:    correlationKey,
		ProcessInstanceID: processInstanceID,
		Variables:         variables,
		CreatedAt:         time.Now(),
		InstanceCreated:   false, // Not creating new instance, correlating with existing
	}

	// NOTE: Complete message correlation with running process instances
	// would involve finding intermediate catch events waiting for this message
	// and triggering them with the provided variables

	// For now, just save the correlation result
	if err := cm.storage.SaveMessageCorrelationResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save correlation result: %w", err)
	}

	cm.logger.Info("Message correlated successfully", logger.String("messageId", messageID))
	return result, nil
}

// GetStats returns message statistics
func (cm *CorrelationManager) GetStats(ctx context.Context, tenantID string) (*MessageStats, error) {
	cm.logger.Debug("Getting message stats")

	// Get active subscriptions
	subscriptions, err := cm.storage.ListProcessMessageSubscriptions(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	activeSubscriptions := 0
	for _, sub := range subscriptions {
		if sub.IsActive {
			activeSubscriptions++
		}
	}

	// Get buffered messages
	bufferedMessages, err := cm.storage.ListBufferedMessages(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get buffered messages: %w", err)
	}

	expiredMessages := 0
	for _, msg := range bufferedMessages {
		if msg.IsExpired() {
			expiredMessages++
		}
	}

	// Get today's correlation results
	today := time.Now().Format("2006-01-02")
	correlationResults, err := cm.storage.ListMessageCorrelationResults(ctx, tenantID, "", "", 1000, 0)
	if err != nil {
		cm.logger.Warn("Failed to get correlation results for stats", logger.String("error", err.Error()))
		correlationResults = []*models.MessageCorrelationResult{}
	}

	publishedToday := 0
	instancesCreatedToday := 0
	for _, result := range correlationResults {
		if result.CreatedAt.Format("2006-01-02") == today {
			publishedToday++
			if result.InstanceCreated {
				instancesCreatedToday++
			}
		}
	}

	stats := &MessageStats{
		TotalMessages:         len(correlationResults),
		BufferedMessages:      len(bufferedMessages),
		ExpiredMessages:       expiredMessages,
		PublishedToday:        publishedToday,
		InstancesCreatedToday: instancesCreatedToday,
	}

	cm.logger.Debug("Message stats calculated",
		logger.Int("bufferedMessages", stats.BufferedMessages),
		logger.Int("publishedToday", stats.PublishedToday),
	)

	return stats, nil
}

// CleanupExpiredData cleans up expired correlation data
func (cm *CorrelationManager) CleanupExpiredData(ctx context.Context) (int, error) {
	cm.logger.Info("Cleaning up expired correlation data")

	// Clean up old correlation results (older than 30 days)
	cutoffDate := time.Now().AddDate(0, 0, -30)

	// Get all correlation results
	results, err := cm.storage.ListMessageCorrelationResults(ctx, "", "", "", 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list correlation results: %w", err)
	}

	cleanedCount := 0
	for _, result := range results {
		if result.CreatedAt.Before(cutoffDate) {
			if err := cm.storage.DeleteMessageCorrelationResult(ctx, result.ID); err != nil {
				cm.logger.Error("Failed to delete correlation result", logger.String("error", err.Error()))
				continue
			}
			cleanedCount++
		}
	}

	cm.logger.Info("Expired correlation data cleaned", logger.Int("cleanedCount", cleanedCount))
	return cleanedCount, nil
}

// cleanupExpiredData runs periodic cleanup
func (cm *CorrelationManager) cleanupExpiredData() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if _, err := cm.CleanupExpiredData(ctx); err != nil {
				cm.logger.Error("Failed to cleanup expired data", logger.String("error", err.Error()))
			}
			cancel()
		case <-cm.stopChan:
			return
		}
	}
}

// isIntermediateCatchEvent checks if element ID is intermediate catch event
// Проверяет является ли element ID intermediate catch event
func (cm *CorrelationManager) isIntermediateCatchEvent(elementID string) bool {
	// Better heuristic: check if there are actually waiting tokens for this element
	// Лучшая эвристика: проверяем есть ли реально ожидающие токены для этого элемента

	// If there are waiting tokens for this element, it's an intermediate catch event
	// Если есть ожидающие токены для этого элемента, это intermediate catch event
	waitingTokens, err := cm.storage.LoadTokensByState(models.TokenStateWaiting)
	if err != nil {
		cm.logger.Error("Failed to load waiting tokens for element type detection",
			logger.String("element_id", elementID),
			logger.String("error", err.Error()))
		// Fallback to simple heuristic
		// Откат к простой эвристике
		return !strings.HasPrefix(elementID, "Start_")
	}

	// Check if any token is waiting on this element
	// Проверяем ждет ли какой-то токен на этом элементе
	for _, token := range waitingTokens {
		if token.CurrentElementID == elementID {
			cm.logger.Info("Found waiting token - this is intermediate catch event",
				logger.String("element_id", elementID),
				logger.String("token_id", token.TokenID))
			return true
		}
	}

	// No waiting tokens found - this is likely a start event
	// Ожидающих токенов не найдено - это скорее всего start event
	cm.logger.Info("No waiting tokens found - this is start event",
		logger.String("element_id", elementID))
	return false
}

// findWaitingToken finds token waiting for message on specific element
// Находит токен ожидающий сообщение на определенном элементе
func (cm *CorrelationManager) findWaitingToken(elementID, messageName string) (*models.Token, error) {
	// Load all waiting tokens
	// Загружаем все ожидающие токены
	waitingTokens, err := cm.storage.LoadTokensByState(models.TokenStateWaiting)
	if err != nil {
		return nil, fmt.Errorf("failed to load waiting tokens: %w", err)
	}

	// Find token waiting for this message on this element
	// Находим токен ожидающий это сообщение на этом элементе
	for _, token := range waitingTokens {
		if token.CurrentElementID == elementID {
			// Check if token is waiting for this message
			// Проверяем ждет ли токен это сообщение
			expectedWaiting := fmt.Sprintf("message:%s", messageName)
			if token.WaitingFor == expectedWaiting {
				cm.logger.Info("Found waiting token for message",
					logger.String("token_id", token.TokenID),
					logger.String("element_id", elementID),
					logger.String("message_name", messageName))
				return token, nil
			}
		}
	}

	cm.logger.Warn("No waiting token found",
		logger.String("element_id", elementID),
		logger.String("message_name", messageName))
	return nil, nil
}
