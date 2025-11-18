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
	"strconv"
	"strings"
	"time"

	"atom-engine/src/core/config"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// Component handles message operations
type Component struct {
	config          *config.Config
	logger          logger.ComponentLogger
	storage         storage.Storage
	correlationMgr  *CorrelationManager
	subscriptionMgr *SubscriptionManager
	bufferMgr       *BufferManager
	responseChannel chan string
	isRunning       bool
}

// NewComponent creates new messages component
func NewComponent(cfg *config.Config, storage storage.Storage) *Component {
	return &Component{
		config:          cfg,
		logger:          logger.NewComponentLogger("messages"),
		storage:         storage,
		responseChannel: make(chan string, 100),
	}
}

// Start initializes and starts the messages component
func Start(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	storageConfig := &storage.Config{
		Path: cfg.Database.Path,
	}
	storageInstance := storage.NewStorage(storageConfig)

	component := NewComponent(cfg, storageInstance)
	return component.Start()
}

// Start starts the component
func (c *Component) Start() error {
	c.logger.Info("Starting messages component", logger.String("component", "messages"))

	// Initialize managers
	c.correlationMgr = NewCorrelationManager(c.storage, c.logger, c.responseChannel)
	c.subscriptionMgr = NewSubscriptionManager(c.storage, c.logger)
	c.bufferMgr = NewBufferManager(c.storage, c.logger)

	// Set correlation manager reference in buffer manager
	// Устанавливаем ссылку на correlation manager в buffer manager
	c.bufferMgr.SetCorrelationManager(c.correlationMgr)

	// Start managers
	if err := c.correlationMgr.Start(); err != nil {
		return fmt.Errorf("failed to start correlation manager: %w", err)
	}

	if err := c.subscriptionMgr.Start(); err != nil {
		return fmt.Errorf("failed to start subscription manager: %w", err)
	}

	if err := c.bufferMgr.Start(); err != nil {
		return fmt.Errorf("failed to start buffer manager: %w", err)
	}

	c.isRunning = true
	c.logger.Info("Messages component started successfully", logger.String("component", "messages"))

	return nil
}

// Stop stops the component
func (c *Component) Stop() error {
	c.logger.Info("Stopping messages component", logger.String("component", "messages"))

	if !c.isRunning {
		return nil
	}

	// Stop managers
	if c.correlationMgr != nil {
		c.correlationMgr.Stop()
	}

	if c.subscriptionMgr != nil {
		c.subscriptionMgr.Stop()
	}

	if c.bufferMgr != nil {
		c.bufferMgr.Stop()
	}

	c.isRunning = false
	c.logger.Info("Messages component stopped", logger.String("component", "messages"))

	return nil
}

// IsRunning returns if component is running
func (c *Component) IsRunning() bool {
	return c.isRunning
}

// PublishMessage publishes a message for correlation
func (c *Component) PublishMessage(
	ctx context.Context,
	tenantID, messageName, correlationKey, elementID string,
	variables map[string]interface{},
	ttl *time.Duration,
) (*models.MessageCorrelationResult, error) {
	c.logger.Info("Publishing message",
		logger.String("messageName", messageName),
		logger.String("correlationKey", correlationKey),
		logger.String("elementID", elementID),
	)

	return c.correlationMgr.PublishMessage(ctx, tenantID, messageName, correlationKey, elementID, variables, ttl)
}

func (c *Component) PublishMessageWithElementID(
	ctx context.Context,
	tenantID, messageName, correlationKey, elementID string,
	variables map[string]interface{},
	ttl *time.Duration,
) (*models.MessageCorrelationResult, error) {
	// Backward compatibility - delegate to main method
	return c.PublishMessage(ctx, tenantID, messageName, correlationKey, elementID, variables, ttl)
}

// CorrelateMessage correlates message with specific process instance
func (c *Component) CorrelateMessage(
	ctx context.Context,
	tenantID, messageName, correlationKey, processInstanceID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	c.logger.Info("Correlating message",
		logger.String("messageName", messageName),
		logger.String("correlationKey", correlationKey),
		logger.String("processInstanceId", processInstanceID),
	)

	return c.correlationMgr.CorrelateMessage(ctx, tenantID, messageName, correlationKey, processInstanceID, variables)
}

// CreateMessageSubscription creates a message subscription
func (c *Component) CreateMessageSubscription(
	ctx context.Context,
	subscription *models.ProcessMessageSubscription,
) error {
	c.logger.Info("Creating message subscription", logger.String("messageName", subscription.MessageName))

	// Create subscription first
	subscriptionCreated := true
	if err := c.subscriptionMgr.CreateSubscription(ctx, subscription); err != nil {
		// Check if error is due to existing subscription
		// Проверяем является ли ошибка из-за существующей подписки
		if strings.Contains(err.Error(), "subscription already exists") {
			c.logger.Info("Subscription already exists - will process buffered messages anyway",
				logger.String("message_name", subscription.MessageName),
				logger.String("process_key", subscription.ProcessDefinitionKey))
			subscriptionCreated = false
		} else {
			return err
		}
	}

	// Process any buffered messages that match this subscription
	// Even if subscription already exists, we should check for buffered messages
	// Обрабатываем любые буферизованные сообщения которые подходят этой подписке
	// Даже если подписка уже существует, мы должны проверить буферизованные сообщения
	if processedCount, err := c.bufferMgr.ProcessBufferedMessages(ctx, subscription); err != nil {
		c.logger.Warn("Failed to process buffered messages for subscription",
			logger.String("subscription_id", subscription.ID),
			logger.String("error", err.Error()))
	} else if processedCount > 0 {
		status := "new"
		if !subscriptionCreated {
			status = "existing"
		}
		c.logger.Info("Processed buffered messages for subscription",
			logger.String("subscription_id", subscription.ID),
			logger.String("status", status),
			logger.Int("processed_count", processedCount))
	}

	return nil
}

// DeleteMessageSubscription deletes a message subscription
func (c *Component) DeleteMessageSubscription(ctx context.Context, subscriptionID string) error {
	c.logger.Info("Deleting message subscription", logger.String("subscriptionID", subscriptionID))

	return c.subscriptionMgr.DeleteSubscription(ctx, subscriptionID)
}

// ListMessageSubscriptions lists message subscriptions
func (c *Component) ListMessageSubscriptions(
	ctx context.Context,
	tenantID string,
	limit, offset int,
) ([]*models.ProcessMessageSubscription, error) {
	c.logger.Debug("Listing message subscriptions")

	return c.subscriptionMgr.ListSubscriptions(ctx, tenantID, limit, offset)
}

// GetMessageSubscription gets message subscription by ID
func (c *Component) GetMessageSubscription(
	ctx context.Context,
	subscriptionID string,
) (*models.ProcessMessageSubscription, error) {
	c.logger.Debug("Getting message subscription", logger.String("subscriptionID", subscriptionID))

	return c.subscriptionMgr.GetSubscriptionByID(ctx, subscriptionID)
}

// ListBufferedMessages lists buffered messages
func (c *Component) ListBufferedMessages(
	ctx context.Context,
	tenantID string,
	limit, offset int,
) ([]*models.BufferedMessage, error) {
	c.logger.Debug("Listing buffered messages")

	return c.bufferMgr.ListBufferedMessages(ctx, tenantID, limit, offset)
}

// CleanupExpiredMessages cleans up expired buffered messages
func (c *Component) CleanupExpiredMessages(ctx context.Context) (int, error) {
	c.logger.Info("Cleaning up expired messages")

	return c.bufferMgr.CleanupExpiredMessages(ctx)
}

// GetMessageStats returns message statistics
func (c *Component) GetMessageStats(ctx context.Context, tenantID string) (*MessageStats, error) {
	c.logger.Debug("Getting message stats")

	return c.correlationMgr.GetStats(ctx, tenantID)
}

// GetCorrelationManager returns correlation manager for internal use
func (c *Component) GetCorrelationManager() *CorrelationManager {
	return c.correlationMgr
}

// GetSubscriptionManager returns subscription manager for internal use
func (c *Component) GetSubscriptionManager() *SubscriptionManager {
	return c.subscriptionMgr
}

// GetBufferManager returns buffer manager for internal use
func (c *Component) GetBufferManager() *BufferManager {
	return c.bufferMgr
}

// GetResponseChannel returns response channel for message callbacks
// Возвращает канал ответов для callback'ов сообщений
func (c *Component) GetResponseChannel() <-chan string {
	return c.responseChannel
}

// SendMessageCallback sends message callback response
// Отправляет callback ответ сообщения
func (c *Component) SendMessageCallback(response string) {
	if c.responseChannel != nil {
		select {
		case c.responseChannel <- response:
		default:
			c.logger.Warn("Message response channel full, callback dropped")
		}
	}
}

// MessageStats represents message statistics
type MessageStats struct {
	TotalMessages         int `json:"total_messages"`
	BufferedMessages      int `json:"buffered_messages"`
	ExpiredMessages       int `json:"expired_messages"`
	PublishedToday        int `json:"published_today"`
	InstancesCreatedToday int `json:"instances_created_today"`
}

// ProcessMessage processes JSON message from core engine
// Обрабатывает JSON сообщение от core engine
func (c *Component) ProcessMessage(ctx context.Context, messageJSON string) error {
	if !c.IsRunning() {
		return fmt.Errorf("messages component not running")
	}

	// Parse message to determine type
	// Парсим сообщение для определения типа
	var request MessageRequest
	if err := json.Unmarshal([]byte(messageJSON), &request); err != nil {
		return fmt.Errorf("failed to parse message request: %w", err)
	}

	c.logger.Debug("Processing message request",
		logger.String("type", request.Type),
		logger.String("request_id", request.RequestID),
	)

	switch request.Type {
	case "publish_message":
		return c.handlePublishMessage(ctx, request)
	case "correlate_message":
		return c.handleCorrelateMessage(ctx, request)
	case "create_subscription":
		return c.handleCreateSubscription(ctx, request)
	case "delete_subscription":
		return c.handleDeleteSubscription(ctx, request)
	case "list_subscriptions":
		return c.handleListSubscriptions(ctx, request)
	case "list_buffered_messages":
		return c.handleListBufferedMessages(ctx, request)
	case "cleanup_expired":
		return c.handleCleanupExpired(ctx, request)
	case "get_stats":
		return c.handleGetStats(ctx, request)
	default:
		return fmt.Errorf("unknown message request type: %s", request.Type)
	}
}

// handlePublishMessage handles message publishing request
// Обрабатывает запрос публикации сообщения
func (c *Component) handlePublishMessage(ctx context.Context, request MessageRequest) error {
	var payload PublishMessagePayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"publish_message_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	// Set TTL if provided
	var ttl *time.Duration
	if payload.TTLSeconds > 0 {
		duration := time.Duration(payload.TTLSeconds) * time.Second
		ttl = &duration
	}

	result, err := c.PublishMessage(
		ctx,
		payload.TenantID,
		payload.MessageName,
		payload.CorrelationKey,
		"",
		payload.Variables,
		ttl,
	)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("publish_message_response", request.RequestID, err.Error())
	} else {
		messageResult := MessageResult{
			MessageID:         result.MessageID,
			CorrelationID:     result.ID,
			Success:           true,
			ProcessInstanceID: result.ProcessInstanceID,
			Variables:         result.Variables,
			Timestamp:         time.Now().Unix(),
		}
		response = CreateMessageResponse("publish_message_response", request.RequestID, messageResult)
	}

	return c.sendResponse(response)
}

// handleCorrelateMessage handles message correlation request
// Обрабатывает запрос корреляции сообщения
func (c *Component) handleCorrelateMessage(ctx context.Context, request MessageRequest) error {
	var payload CorrelateMessagePayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"correlate_message_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	result, err := c.CorrelateMessage(
		ctx,
		payload.TenantID,
		payload.MessageName,
		payload.CorrelationKey,
		payload.ProcessInstanceID,
		payload.Variables,
	)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("correlate_message_response", request.RequestID, err.Error())
	} else {
		messageResult := MessageResult{
			MessageID:         result.MessageID,
			CorrelationID:     result.ID,
			Success:           true,
			ProcessInstanceID: result.ProcessInstanceID,
			Variables:         result.Variables,
			Timestamp:         time.Now().Unix(),
		}
		response = CreateMessageResponse("correlate_message_response", request.RequestID, messageResult)
	}

	return c.sendResponse(response)
}

// handleCreateSubscription handles subscription creation request
// Обрабатывает запрос создания подписки
func (c *Component) handleCreateSubscription(ctx context.Context, request MessageRequest) error {
	var payload CreateSubscriptionPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"create_subscription_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	// Extract process version from ProcessKey
	processVersion := extractVersionFromKey(payload.ProcessKey)

	// Create ProcessMessageSubscription from payload
	subscription := &models.ProcessMessageSubscription{
		ID:                   models.GenerateID(),
		TenantID:             payload.TenantID,
		ProcessDefinitionKey: payload.ProcessKey,
		ProcessVersion:       int32(processVersion), // Use actual version from ProcessKey
		StartEventID:         payload.ElementID,
		MessageName:          payload.MessageName,
		MessageRef:           payload.MessageName,
		CorrelationKey:       payload.CorrelationKey,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := c.CreateMessageSubscription(ctx, subscription)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("create_subscription_response", request.RequestID, err.Error())
	} else {
		subscResult := SubscriptionResult{
			SubscriptionID: subscription.ID, // Use ID from subscription object
			Success:        true,
			Message:        "Subscription created successfully",
			Timestamp:      time.Now().Unix(),
		}
		response = CreateMessageResponse("create_subscription_response", request.RequestID, subscResult)
	}

	return c.sendResponse(response)
}

// handleDeleteSubscription handles subscription deletion request
// Обрабатывает запрос удаления подписки
func (c *Component) handleDeleteSubscription(ctx context.Context, request MessageRequest) error {
	var payload DeleteSubscriptionPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"delete_subscription_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	err := c.DeleteMessageSubscription(ctx, payload.SubscriptionID)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("delete_subscription_response", request.RequestID, err.Error())
	} else {
		subscResult := SubscriptionResult{
			SubscriptionID: payload.SubscriptionID,
			Success:        true,
			Message:        "Subscription deleted successfully",
			Timestamp:      time.Now().Unix(),
		}
		response = CreateMessageResponse("delete_subscription_response", request.RequestID, subscResult)
	}

	return c.sendResponse(response)
}

// handleListSubscriptions handles subscription listing request
// Обрабатывает запрос списка подписок
func (c *Component) handleListSubscriptions(ctx context.Context, request MessageRequest) error {
	var payload ListSubscriptionsPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"list_subscriptions_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	subscriptions, err := c.ListMessageSubscriptions(ctx, payload.TenantID, payload.Limit, payload.Offset)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("list_subscriptions_response", request.RequestID, err.Error())
	} else {
		response = CreateMessageResponse("list_subscriptions_response", request.RequestID, subscriptions)
	}

	return c.sendResponse(response)
}

// handleListBufferedMessages handles buffered messages listing request
// Обрабатывает запрос списка буферизованных сообщений
func (c *Component) handleListBufferedMessages(ctx context.Context, request MessageRequest) error {
	var payload ListBufferedMessagesPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"list_buffered_messages_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	messages, err := c.ListBufferedMessages(ctx, payload.TenantID, payload.Limit, payload.Offset)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("list_buffered_messages_response", request.RequestID, err.Error())
	} else {
		response = CreateMessageResponse("list_buffered_messages_response", request.RequestID, messages)
	}

	return c.sendResponse(response)
}

// handleCleanupExpired handles expired messages cleanup request
// Обрабатывает запрос очистки просроченных сообщений
func (c *Component) handleCleanupExpired(ctx context.Context, request MessageRequest) error {
	var payload CleanupExpiredPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"cleanup_expired_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	expiredCount, err := c.CleanupExpiredMessages(ctx)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("cleanup_expired_response", request.RequestID, err.Error())
	} else {
		cleanupResult := CleanupResult{
			ExpiredCount: expiredCount,
			Success:      true,
			Message:      fmt.Sprintf("Cleaned up %d expired messages", expiredCount),
			Timestamp:    time.Now().Unix(),
		}
		response = CreateMessageResponse("cleanup_expired_response", request.RequestID, cleanupResult)
	}

	return c.sendResponse(response)
}

// handleGetStats handles get statistics request
// Обрабатывает запрос получения статистики
func (c *Component) handleGetStats(ctx context.Context, request MessageRequest) error {
	var payload GetStatsPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateMessageErrorResponse(
			"get_stats_response",
			request.RequestID,
			fmt.Sprintf("invalid payload: %v", err),
		)
		return c.sendResponse(response)
	}

	stats, err := c.GetMessageStats(ctx, payload.TenantID)

	var response MessageResponse
	if err != nil {
		response = CreateMessageErrorResponse("get_stats_response", request.RequestID, err.Error())
	} else {
		response = CreateMessageResponse("get_stats_response", request.RequestID, stats)
	}

	return c.sendResponse(response)
}

// sendResponse sends message response through response channel
// Отправляет ответ сообщения через канал ответов
func (c *Component) sendResponse(response MessageResponse) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		c.logger.Error("Failed to marshal message response", logger.String("error", err.Error()))
		return fmt.Errorf("failed to marshal message response: %w", err)
	}

	c.logger.Debug("Sending message response",
		logger.String("type", response.Type),
		logger.String("request_id", response.RequestID),
	)

	if c.responseChannel != nil {
		select {
		case c.responseChannel <- string(responseJSON):
		default:
			c.logger.Warn("Message response channel full, response dropped")
			return fmt.Errorf("message response channel full")
		}
	}

	return nil
}

// extractVersionFromKey extracts version from process key
func extractVersionFromKey(processKey string) int {
	if strings.Contains(processKey, ":v") {
		parts := strings.Split(processKey, ":v")
		if len(parts) > 1 {
			if version, err := strconv.Atoi(parts[1]); err == nil {
				return version
			}
		}
	}
	return 1
}

