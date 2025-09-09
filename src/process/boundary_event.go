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

	// Extract and evaluate correlation key from token variables
	// Извлекаем и вычисляем correlation key из переменных токена
	correlationKey = bee.evaluateCorrelationKey(token)

	// Create message subscription for this boundary event
	// Создаем подписку на сообщение для этого граничного события
	if bee.processComponent != nil && messageName != "" {
		subscription := &models.ProcessMessageSubscription{
			ID:                   models.GenerateID(),
			TenantID:             "", // Default tenant
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
		// Activity cancellation for interrupting events not implemented
	}, nil
}

// handleSignalBoundaryEvent handles signal boundary events
// Обрабатывает граничные события сигналов
func (bee *BoundaryEventExecutor) handleSignalBoundaryEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}, cancelActivity bool) (*ExecutionResult, error) {
	logger.Info("Handling signal boundary event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Bool("cancel_activity", cancelActivity))

	// Extract signal name from event definition
	signalName := ""
	if signalRef, exists := eventDef["signal_ref"]; exists {
		if signalRefStr, ok := signalRef.(string); ok {
			signalName = signalRefStr
		}
	}

	// Fallback: use element ID as signal name if no signal_ref
	if signalName == "" {
		signalName = token.CurrentElementID + "_signal"
		logger.Warn("No signal_ref found, using element ID as signal name",
			logger.String("signal_name", signalName))
	}

	// Subscribe to signal using process component
	if bee.processComponent != nil {
		variables := make(map[string]interface{})
		if token.Variables != nil {
			variables = token.Variables
		}

		err := bee.processComponent.SubscribeToSignal(signalName, token.TokenID, token.CurrentElementID, cancelActivity, variables)
		if err != nil {
			logger.Error("Failed to subscribe to signal",
				logger.String("signal_name", signalName),
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("failed to subscribe to signal: %v", err),
				Completed: false,
			}, err
		}

		logger.Info("Successfully subscribed to signal",
			logger.String("signal_name", signalName),
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))

		// Return waiting state - signal will trigger via callback when received
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: false,
			NextElements: []string{},
			WaitingFor:   fmt.Sprintf("signal:%s", signalName),
			Completed:    false,
		}, nil
	}

	// Fallback if no process component available
	logger.Warn("Process component not available, treating as regular boundary event")
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
		// Activity cancellation for interrupting boundary events not implemented
	}, nil
}

// evaluateCorrelationKey evaluates FEEL expressions in correlation key
// Вычисляет FEEL expressions в correlation key
func (bee *BoundaryEventExecutor) evaluateCorrelationKey(token *models.Token) string {
	correlationKey := ""

	// Extract correlation key from token variables
	// Извлекаем correlation key из переменных токена
	if corrKey, exists := token.Variables["correlationKey"]; exists {
		if corrKeyStr, ok := corrKey.(string); ok {
			// Check if this is a FEEL expression
			// Проверяем является ли это FEEL expression
			if len(corrKeyStr) > 0 && corrKeyStr[0] == '=' {
				// Evaluate FEEL expression
				// Вычисляем FEEL expression
				if evaluatedKey := bee.evaluateFEELExpression(corrKeyStr, token); evaluatedKey != "" {
					correlationKey = evaluatedKey
				} else {
					// Fallback to original value without "="
					correlationKey = corrKeyStr[1:]
				}
			} else {
				// Not a FEEL expression - use as is
				correlationKey = corrKeyStr
			}
		}
	}

	return correlationKey
}

// evaluateFEELExpression evaluates FEEL expression using expression component
// Вычисляет FEEL expression используя expression компонент
func (bee *BoundaryEventExecutor) evaluateFEELExpression(expression string, token *models.Token) string {
	// Get expression component through process component
	// Получаем expression компонент через process компонент
	if bee.processComponent == nil {
		return ""
	}

	// Get core interface
	core := bee.processComponent.GetCore()
	if core == nil {
		return ""
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		return ""
	}

	// Cast to expression evaluator interface
	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		return ""
	}

	// Evaluate FEEL expression
	result, err := expressionComp.EvaluateExpressionEngine(expression, token.Variables)
	if err != nil {
		logger.Error("Failed to evaluate FEEL expression in correlation key",
			logger.String("token_id", token.TokenID),
			logger.String("expression", expression),
			logger.String("error", err.Error()))
		return ""
	}

	// Convert result to string
	if resultStr := fmt.Sprintf("%v", result); resultStr != "" {
		logger.Debug("Correlation key FEEL expression evaluated",
			logger.String("token_id", token.TokenID),
			logger.String("original", expression),
			logger.String("evaluated", resultStr))
		return resultStr
	}

	return ""
}

// GetElementType returns element type
// Возвращает тип элемента
func (bee *BoundaryEventExecutor) GetElementType() string {
	return "boundaryEvent"
}
