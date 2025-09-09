/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// UnifiedMessageManager implements MessageCallbackManagerInterface
// Объединенный менеджер сообщений
type UnifiedMessageManager struct {
	storage   storage.Storage
	component ComponentInterface
	core      CoreInterface
	processor *BufferedMessageProcessor
}

// NewUnifiedMessageManager creates new unified message manager
// Создает новый объединенный менеджер сообщений
func NewUnifiedMessageManager(storage storage.Storage, component ComponentInterface) *UnifiedMessageManager {
	umm := &UnifiedMessageManager{
		storage:   storage,
		component: component,
	}

	return umm
}

// SetCore sets core interface
// Устанавливает интерфейс core
func (umm *UnifiedMessageManager) SetCore(core CoreInterface) {
	umm.core = core
	umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
}

// HandleMessageCallback handles message callback via component interface
// Обрабатывает message callback через интерфейс компонента
func (umm *UnifiedMessageManager) HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error {
	// Use component interface instead of direct engine access
	if umm.component != nil {
		return umm.component.HandleMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
	}
	return fmt.Errorf("component not available for message callback")
}

// CheckBufferedMessages checks for buffered messages
// Проверяет буферизованные сообщения
func (umm *UnifiedMessageManager) CheckBufferedMessages(messageName, correlationKey string) (*models.BufferedMessage, error) {
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
func (umm *UnifiedMessageManager) PublishMessage(messageName, correlationKey string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.PublishMessage(messageName, correlationKey, variables)
}

func (umm *UnifiedMessageManager) PublishMessageWithElementID(messageName, correlationKey, elementID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.PublishMessageWithElementID(messageName, correlationKey, elementID, variables)
}

// CorrelateMessage correlates message
// Коррелирует сообщение
func (umm *UnifiedMessageManager) CorrelateMessage(messageName, correlationKey, processInstanceID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	if umm.processor == nil {
		umm.processor = NewBufferedMessageProcessor(umm.storage, umm.core)
	}
	return umm.processor.CorrelateMessage(messageName, correlationKey, processInstanceID, variables)
}
