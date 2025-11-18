/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// UnifiedMessageManager implements MessageCallbackManagerInterface
// Объединенный менеджер сообщений
type UnifiedMessageManager struct {
	storage        storage.Storage
	component      ComponentInterface
	core           CoreInterface
	processor      *BufferedMessageProcessor
	callbackHelper *CallbackHelper
}

// NewUnifiedMessageManager creates new unified message manager
// Создает новый объединенный менеджер сообщений
func NewUnifiedMessageManager(storage storage.Storage, component ComponentInterface) *UnifiedMessageManager {
	umm := &UnifiedMessageManager{
		storage:        storage,
		component:      component,
		callbackHelper: NewCallbackHelper(storage, component),
	}

	return umm
}

// SetCore sets core interface
// Устанавливает интерфейс core
func (umm *UnifiedMessageManager) SetCore(core CoreInterface) {
	umm.core = core
	umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
}

// HandleMessageCallback handles message callback following proper architectural patterns
// Обрабатывает message callback следуя правильным архитектурным паттернам
func (umm *UnifiedMessageManager) HandleMessageCallback(
	messageID, messageName, correlationKey, tokenID string,
	variables map[string]interface{},
) error {
	if !umm.component.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	logger.Info("UnifiedMessageManager handling message callback",
		logger.String("message_id", messageID),
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey),
		logger.String("token_id", tokenID))

	// Check if this is Message Start Event callback (empty token_id)
	// Message Start Events create new process instances - delegate to engine
	if tokenID == "" {
		logger.Info("Message Start Event callback detected - delegating to engine",
			logger.String("message_id", messageID),
			logger.String("message_name", messageName))

		// Delegate Message Start Event handling to engine
		return umm.component.HandleEngineMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
	}

	// Handle Intermediate Catch Message Events using CallbackHelper pattern
	// Обрабатываем Intermediate Catch Message Events используя паттерн CallbackHelper
	return umm.handleIntermediateCatchMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
}

// handleIntermediateCatchMessageCallback handles intermediate catch message events
// Обрабатывает intermediate catch message события
func (umm *UnifiedMessageManager) handleIntermediateCatchMessageCallback(
	messageID, messageName, correlationKey, tokenID string,
	variables map[string]interface{},
) error {
	logger.Info("Handling Intermediate Catch Message Event callback",
		logger.String("message_id", messageID),
		logger.String("message_name", messageName),
		logger.String("token_id", tokenID))

	// Load and validate token using CallbackHelper (same pattern as TimerCallbacks and JobCallbacks)
	expectedWaitingFor := fmt.Sprintf("message:%s", messageName)
	token, err := umm.callbackHelper.LoadAndValidateToken(tokenID, expectedWaitingFor)
	if err != nil {
		logger.Error("Failed to load and validate token for message callback",
			logger.String("message_id", messageID),
			logger.String("token_id", tokenID),
			logger.String("expected_waiting_for", expectedWaitingFor),
			logger.String("error", err.Error()))
		return err
	}

	logger.Info("Token validated for message callback - proceeding with callback processing",
		logger.String("message_id", messageID),
		logger.String("token_id", tokenID),
		logger.String("message_name", messageName))

	// Process callback and continue execution using CallbackHelper (same pattern as other managers)
	return umm.callbackHelper.ProcessCallbackAndContinue(token, token.CurrentElementID, variables)
}

// CheckBufferedMessages checks for buffered messages
// Проверяет буферизованные сообщения
func (umm *UnifiedMessageManager) CheckBufferedMessages(
	messageName, correlationKey string,
) (*models.BufferedMessage, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.CheckBufferedMessages(messageName, correlationKey)
}

// ProcessBufferedMessage processes buffered message
// Обрабатывает буферизованное сообщение
func (umm *UnifiedMessageManager) ProcessBufferedMessage(message *models.BufferedMessage, token *models.Token) error {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.ProcessBufferedMessage(message, token)
}

// CreateMessageSubscription creates message subscription
// Создает подписку на сообщение
func (umm *UnifiedMessageManager) CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.CreateMessageSubscription(subscription)
}

// DeleteMessageSubscription deletes message subscription
// Удаляет подписку на сообщение
func (umm *UnifiedMessageManager) DeleteMessageSubscription(subscriptionID string) error {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.DeleteMessageSubscription(subscriptionID)
}

// PublishMessage publishes message
// Публикует сообщение
func (umm *UnifiedMessageManager) PublishMessage(
	messageName, correlationKey string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.PublishMessage(messageName, correlationKey, variables)
}

func (umm *UnifiedMessageManager) PublishMessageWithElementID(
	messageName, correlationKey, elementID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.PublishMessageWithElementID(messageName, correlationKey, elementID, variables)
}

// CorrelateMessage correlates message
// Коррелирует сообщение
func (umm *UnifiedMessageManager) CorrelateMessage(
	messageName, correlationKey, processInstanceID string,
	variables map[string]interface{},
) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.CorrelateMessage(messageName, correlationKey, processInstanceID, variables)
}
