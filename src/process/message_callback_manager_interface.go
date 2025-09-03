/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/models"
)

// MessageCallbackManagerInterface handles message operations
// Интерфейс менеджера сообщений
type MessageCallbackManagerInterface interface {
	// Message callback operations
	HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error
	CheckBufferedMessages(messageName, correlationKey string) (*models.BufferedMessage, error)
	ProcessBufferedMessage(message *models.BufferedMessage, token *models.Token) error

	// Message subscription operations
	CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error
	DeleteMessageSubscription(subscriptionID string) error
	PublishMessage(messageName, correlationKey string, variables map[string]interface{}) (*models.MessageCorrelationResult, error)
	CorrelateMessage(messageName, correlationKey, processInstanceID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error)
}
