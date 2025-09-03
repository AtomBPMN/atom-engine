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

// StartEventExecutor executes start events
// Исполнитель стартовых событий
type StartEventExecutor struct {
	processComponent ComponentInterface
}

// NewStartEventExecutor creates new start event executor
// Создает новый исполнитель стартового события
func NewStartEventExecutor(processComponent ComponentInterface) *StartEventExecutor {
	return &StartEventExecutor{
		processComponent: processComponent,
	}
}

// Execute executes start event
// Выполняет стартовое событие
func (se *StartEventExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing start event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Debug: show all element keys
	elementKeys := make([]string, 0, len(element))
	for k := range element {
		elementKeys = append(elementKeys, k)
	}
	logger.Info("StartEvent element structure",
		logger.String("available_keys", fmt.Sprintf("%v", elementKeys)))

	// Check for event definitions to determine start event type
	// Проверяем event definitions чтобы определить тип стартового события
	if eventDefinitions, hasEventDefs := element["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					eventType, _ := eventDefMap["type"].(string)

					// Handle message start events
					if eventType == "messageEventDefinition" {
						return se.handleMessageStartEvent(token, element, eventDefMap)
					}

					// Handle timer start events
					if eventType == "timerEventDefinition" {
						return se.handleTimerStartEvent(token, element, eventDefMap)
					}

					// Handle signal start events
					if eventType == "signalEventDefinition" {
						return se.handleSignalStartEvent(token, element, eventDefMap)
					}
				}
			}
		}
	}

	// Regular start event - simply pass the token to outgoing sequence flows
	// Обычное стартовое событие - просто передаем токен к исходящим sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		logger.Error("StartEvent missing outgoing flows",
			logger.String("element_id", token.CurrentElementID),
			logger.String("available_keys", fmt.Sprintf("%v", elementKeys)))
		return &ExecutionResult{
			Success:   false,
			Error:     "start event has no outgoing sequence flows",
			Completed: false,
		}, nil
	}

	logger.Info("StartEvent outgoing data found",
		logger.String("outgoing_type", fmt.Sprintf("%T", outgoing)),
		logger.String("outgoing_value", fmt.Sprintf("%v", outgoing)))

	// Get outgoing sequence flows
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

	if len(nextElements) == 0 {
		return &ExecutionResult{
			Success:   false,
			Error:     "start event has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	logger.Info("StartEvent executed - moving to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Int("next_elements", len(nextElements)),
		logger.String("next_list", fmt.Sprintf("%v", nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// handleMessageStartEvent handles message start events
// Обрабатывает стартовые события сообщений
func (se *StartEventExecutor) handleMessageStartEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Handling message start event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Check if this token was created by message correlation (auto-start)
	// Проверяем был ли этот токен создан через message correlation (автозапуск)
	if se.isAutoStartedToken(token) {
		logger.Info("Message Start Event auto-started by correlation - proceeding to next elements",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))

		// For auto-started tokens, proceed directly to outgoing sequence flows
		// Для автозапущенных токенов переходим сразу к исходящим sequence flows
		return se.executeRegularStartEvent(token, element)
	}

	// This is initial process registration - create subscription and wait
	// Это первичная регистрация процесса - создаем подписку и ждем
	logger.Info("Message Start Event registration - creating subscription",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract message information from event definition
	// Извлекаем информацию о сообщении из event definition
	messageName := ""
	correlationKey := ""

	if messageRef, exists := eventDef["message_ref"]; exists {
		if messageRefStr, ok := messageRef.(string); ok {
			messageName = messageRefStr
			logger.Info("Message start event message reference",
				logger.String("message_ref", messageRefStr))
		}
	}

	// TODO: Extract correlation key from message definition or token context
	// ТОДО: Извлечь correlation key из message definition или контекста токена

	// Create message subscription for this start event
	// Создаем подписку на сообщение для этого стартового события
	if se.processComponent != nil && messageName != "" {
		subscription := &models.ProcessMessageSubscription{
			ID:                   models.GenerateID(),
			TenantID:             "", // TODO: Get tenant from context
			ProcessDefinitionKey: token.ProcessKey,
			StartEventID:         token.CurrentElementID,
			MessageName:          messageName,
			CorrelationKey:       correlationKey,
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := se.processComponent.CreateMessageSubscription(subscription); err != nil {
			logger.Error("Failed to create message subscription for start event",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Message subscription created for start event",
				logger.String("subscription_id", subscription.ID),
				logger.String("message_name", messageName))
		}
	}

	// Message start events wait for message correlation
	// Стартовые события сообщений ожидают корреляции сообщения
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		WaitingFor:   fmt.Sprintf("message:%s", messageName),
		Completed:    false,
	}, nil
}

// handleTimerStartEvent handles timer start events
// Обрабатывает стартовые события таймера
func (se *StartEventExecutor) handleTimerStartEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Handling timer start event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Timer start events are typically triggered by external scheduler
	// For now, treat as regular start event
	// Стартовые события таймера обычно запускаются внешним планировщиком
	// Пока что обрабатываем как обычное стартовое событие
	return se.executeRegularStartEvent(token, element)
}

// handleSignalStartEvent handles signal start events
// Обрабатывает стартовые события сигнала
func (se *StartEventExecutor) handleSignalStartEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Handling signal start event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Signal start events wait for signal broadcast
	// For now, treat as regular start event
	// Стартовые события сигнала ожидают broadcast сигнала
	// Пока что обрабатываем как обычное стартовое событие
	return se.executeRegularStartEvent(token, element)
}

// executeRegularStartEvent executes regular start event flow
// Выполняет поток обычного стартового события
func (se *StartEventExecutor) executeRegularStartEvent(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:   false,
			Error:     "start event has no outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// Get outgoing sequence flows
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

	if len(nextElements) == 0 {
		return &ExecutionResult{
			Success:   false,
			Error:     "start event has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	logger.Info("Regular start event executed",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Int("next_elements_count", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (se *StartEventExecutor) GetElementType() string {
	return "startEvent"
}

// isAutoStartedToken checks if token was created by message correlation auto-start
// Проверяет был ли токен создан через message correlation автозапуск
func (se *StartEventExecutor) isAutoStartedToken(token *models.Token) bool {
	// Check if token variables contain message correlation data
	// Проверяем содержат ли переменные токена данные message correlation
	if _, hasData := token.Variables["data"]; hasData {
		logger.Info("Token contains message correlation data - auto-started",
			logger.String("token_id", token.TokenID))
		return true
	}

	// Check if process instance was created within last few seconds (likely auto-started)
	// Проверяем был ли process instance создан в последние несколько секунд (вероятно автозапуск)
	if token.CreatedAt.Add(time.Second * 5).After(time.Now()) {
		logger.Info("Token created recently - likely auto-started",
			logger.String("token_id", token.TokenID))
		return true
	}

	logger.Info("Token appears to be from initial registration",
		logger.String("token_id", token.TokenID))
	return false
}
