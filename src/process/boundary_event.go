/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// BoundaryEventExecutor executes boundary events
// Исполнитель граничных событий
type BoundaryEventExecutor struct {
	processComponent ComponentInterface
}

// NewBoundaryEventExecutor creates new boundary event executor
// Создает новый исполнитель граничного события
func NewBoundaryEventExecutor(processComponent ComponentInterface) *BoundaryEventExecutor {
	return &BoundaryEventExecutor{
		processComponent: processComponent,
	}
}

// Execute executes boundary event
// Выполняет граничное событие
func (bee *BoundaryEventExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing boundary event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get event name for logging
	eventName, _ := element["name"].(string)
	if eventName == "" {
		eventName = token.CurrentElementID
	}

	// Check for attached activity
	attachedTo, hasAttached := element["attached_to_ref"]
	if !hasAttached {
		logger.Error("Boundary event missing attachedToRef",
			logger.String("element_id", token.CurrentElementID))
		return &ExecutionResult{
			Success:   false,
			Error:     "boundary event has no attached activity",
			Completed: false,
		}, nil
	}

	// Check if boundary event is interrupting
	cancelActivity := true // Default is interrupting
	if cancelActivityAttr, exists := element["cancel_activity"]; exists {
		if cancelActivityBool, ok := cancelActivityAttr.(bool); ok {
			cancelActivity = cancelActivityBool
		} else if cancelActivityStr, ok := cancelActivityAttr.(string); ok {
			cancelActivity = cancelActivityStr != "false"
		}
	}

	// Check for event definitions to determine boundary event type
	// Проверяем event definitions чтобы определить тип граничного события
	if eventDefinitions, hasEventDefs := element["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					eventType, _ := eventDefMap["type"].(string)

					// Handle message boundary events
					if eventType == "messageEventDefinition" {
						return bee.handleMessageBoundaryEvent(token, element, eventDefMap, cancelActivity)
					}

					// Handle timer boundary events (already implemented in boundary timer manager)
					if eventType == "timerEventDefinition" {
						logger.Info("Timer boundary event - handled by boundary timer manager",
							logger.String("token_id", token.TokenID),
							logger.String("element_id", token.CurrentElementID))
						// This is handled by the boundary timer manager
						return bee.executeRegularBoundaryEvent(token, element, cancelActivity)
					}

					// Handle signal boundary events
					if eventType == "signalEventDefinition" {
						return bee.handleSignalBoundaryEvent(token, element, eventDefMap, cancelActivity)
					}

					// Handle error boundary events
					if eventType == "errorEventDefinition" {
						return bee.handleErrorBoundaryEvent(token, element, eventDefMap, cancelActivity)
					}
				}
			}
		}
	}

	// Regular boundary event
	// Обычное граничное событие
	logger.Info("Regular boundary event triggered",
		logger.String("token_id", token.TokenID),
		logger.String("event_name", eventName),
		logger.String("attached_to", attachedTo.(string)))

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		// Boundary event without outgoing flows completes the token
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	return bee.executeRegularBoundaryEvent(token, element, true)
}

// handleMessageBoundaryEvent handles message boundary events
// Обрабатывает граничные события сообщений
func (bee *BoundaryEventExecutor) handleMessageBoundaryEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}, cancelActivity bool) (*ExecutionResult, error) {
	logger.Info("Handling message boundary event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Bool("cancel_activity", cancelActivity))

	// Extract message information from event definition
	// Извлекаем информацию о сообщении из event definition
	messageName := ""
	correlationKey := ""

	if messageRef, exists := eventDef["message_ref"]; exists {
		if messageRefStr, ok := messageRef.(string); ok {
			messageName = messageRefStr
			logger.Info("Message boundary event message reference",
				logger.String("message_ref", messageRefStr))
		}
	}

	// Extract correlation key from token variables or use default
	// Извлекаем correlation key из переменных токена или используем default
	if corrKey, exists := token.Variables["correlationKey"]; exists {
		if corrKeyStr, ok := corrKey.(string); ok {
			correlationKey = corrKeyStr
		}
	}

	// Create message subscription for this boundary event
	// Создаем подписку на сообщение для этого граничного события
	if bee.processComponent != nil && messageName != "" {
		subscription := &models.ProcessMessageSubscription{
			ID:                   models.GenerateID(),
			TenantID:             "", // TODO: Get tenant from context
			ProcessDefinitionKey: token.ProcessKey,
			StartEventID:         token.CurrentElementID, // Use current element as reference
			MessageName:          messageName,
			CorrelationKey:       correlationKey,
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := bee.processComponent.CreateMessageSubscription(subscription); err != nil {
			logger.Error("Failed to create message subscription for boundary event",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Message subscription created for boundary event",
				logger.String("subscription_id", subscription.ID),
				logger.String("message_name", messageName))
		}
	}

	// Get outgoing flows for later continuation
	// Получаем исходящие потоки для последующего продолжения
	var nextElements []string
	if outgoing, exists := element["outgoing"]; exists {
		if outgoingList, ok := outgoing.([]interface{}); ok {
			for _, item := range outgoingList {
				if flowID, ok := item.(string); ok {
					nextElements = append(nextElements, flowID)
				}
			}
		} else if outgoingStr, ok := outgoing.(string); ok {
			nextElements = append(nextElements, outgoingStr)
		}
	}

	// Message boundary events wait for message correlation
	// Граничные события сообщений ожидают корреляции сообщения
	logger.Info("Message boundary event waiting for correlation",
		logger.String("token_id", token.TokenID),
		logger.String("message_name", messageName),
		logger.Bool("cancel_activity", cancelActivity))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: nextElements,
		WaitingFor:   fmt.Sprintf("message:%s", messageName),
		Completed:    false,
		// TODO: Implement activity cancellation logic for interrupting events
	}, nil
}

// handleSignalBoundaryEvent handles signal boundary events
// Обрабатывает граничные события сигналов
func (bee *BoundaryEventExecutor) handleSignalBoundaryEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}, cancelActivity bool) (*ExecutionResult, error) {
	logger.Info("Handling signal boundary event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Bool("cancel_activity", cancelActivity))

	// TODO: Implement signal subscription
	// ТОДО: Реализовать подписку на сигнал
	logger.Info("Signal boundary event - signal subscription not yet implemented")

	return bee.executeRegularBoundaryEvent(token, element, cancelActivity)
}

// handleErrorBoundaryEvent handles error boundary events
// Обрабатывает граничные события ошибок
func (bee *BoundaryEventExecutor) handleErrorBoundaryEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}, cancelActivity bool) (*ExecutionResult, error) {
	logger.Info("Handling error boundary event during setup",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Bool("cancel_activity", cancelActivity))

	// Error boundary events are passive - they don't execute immediately
	// Instead, they are registered as subscriptions when the attached activity starts
	// The actual error handling happens in job failure callbacks
	//
	// Граничные события ошибок пассивны - они не выполняются сразу
	// Вместо этого они регистрируются как подписки при запуске прикрепленной активности
	// Фактическая обработка ошибок происходит в callback'ах провалов job'ов

	logger.Info("Error boundary event is passive - handled via job failure callbacks",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// For error boundary events, we don't execute them directly
	// They are handled by the error boundary registry system
	// Return a "waiting" state to indicate this boundary event is monitoring
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: []string{}, // No immediate next elements
		WaitingFor:   fmt.Sprintf("error_boundary:%s", token.CurrentElementID),
		Completed:    false, // Remains active until error occurs or parent completes
	}, nil
}

// executeRegularBoundaryEvent executes regular boundary event flow
// Выполняет поток обычного граничного события
func (bee *BoundaryEventExecutor) executeRegularBoundaryEvent(token *models.Token, element map[string]interface{}, cancelActivity bool) (*ExecutionResult, error) {
	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	logger.Info("Regular boundary event executed",
		logger.String("token_id", token.TokenID),
		logger.Bool("cancel_activity", cancelActivity),
		logger.Int("next_elements", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
		// TODO: Handle activity cancellation for interrupting boundary events
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (bee *BoundaryEventExecutor) GetElementType() string {
	return "boundaryEvent"
}
