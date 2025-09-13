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

// ReceiveTaskExecutor executes receive tasks
// Исполнитель задач получения
type ReceiveTaskExecutor struct {
	processComponent ComponentInterface
	messageHandler   *IntermediateCatchMessageHandler
}

// NewReceiveTaskExecutor creates new receive task executor
// Создает новый исполнитель задач получения
func NewReceiveTaskExecutor(processComponent ComponentInterface) *ReceiveTaskExecutor {
	return &ReceiveTaskExecutor{
		processComponent: processComponent,
		messageHandler:   NewIntermediateCatchMessageHandler(processComponent),
	}
}

// Execute executes receive task by creating message subscription and waiting
// Выполняет задачу получения создавая подписку на сообщение и ожидая
func (rte *ReceiveTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing receive task",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get task name for logging
	taskName, _ := element["name"].(string)
	if taskName == "" {
		taskName = token.CurrentElementID
	}

	// Create boundary timers when token enters activity
	// Создаем boundary таймеры когда токен входит в активность
	if err := rte.createBoundaryTimers(token, element); err != nil {
		logger.Error("Failed to create boundary timers",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - boundary timer creation is not critical
		// Продолжаем выполнение - создание boundary таймеров не критично
	}

	// Create error boundary subscriptions when token enters activity
	// Создаем подписки на граничные события ошибок когда токен входит в активность
	if err := rte.createErrorBoundaries(token, element); err != nil {
		logger.Error("Failed to create error boundary subscriptions",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - error boundary creation is not critical
		// Продолжаем выполнение - создание граничных событий ошибок не критично
	}

	// Check if this token was activated by message correlation
	// Проверяем был ли этот токен активирован через message correlation
	if rte.isMessageCorrelatedToken(token) {
		logger.Info("Receive Task already correlated - proceeding to next elements",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))

		// Token was already activated by message correlation, proceed to next elements
		// Токен уже активирован через message correlation, переходим к следующим элементам
		return rte.proceedToNextElements(element)
	}

	// Extract message information from receive_task section
	// Извлекаем информацию о сообщении из секции receive_task
	messageName := ""
	messageRef := ""

	if receiveTaskData, exists := element["receive_task"]; exists {
		if receiveTaskMap, ok := receiveTaskData.(map[string]interface{}); ok {
			if msgRef, exists := receiveTaskMap["message_ref"]; exists {
				if msgRefStr, ok := msgRef.(string); ok {
					messageRef = msgRefStr
					logger.Info("Receive task messageRef found",
						logger.String("message_ref", msgRefStr))
				}
			}
		}
	}

	// Resolve messageRef to actual message name
	// Разрешаем messageRef в настоящее имя сообщения
	if messageRef != "" {
		actualMessageName := rte.getMessageNameByReference(messageRef, token)
		if actualMessageName != "" {
			messageName = actualMessageName
			logger.Info("Receive task message name resolved",
				logger.String("message_ref", messageRef),
				logger.String("message_name", messageName))
		}
	}

	// Extract correlation key from message definition
	// Извлекаем correlation key из определения сообщения
	correlationKey := rte.extractCorrelationKeyFromMessage(token, messageName)

	// Get outgoing flows for later use
	// Получаем исходящие потоки для последующего использования
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

	// Check for buffered messages first
	// Сначала проверяем буферизованные сообщения
	if messageName != "" && rte.processComponent != nil {
		bufferedMessage, err := rte.processComponent.CheckBufferedMessages(messageName, correlationKey)
		if err != nil {
			logger.Error("Failed to check buffered messages",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
		} else if bufferedMessage != nil {
			logger.Info("Found buffered message for receive task - processing immediately",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey))

			// Process buffered message and continue
			rte.processComponent.ProcessBufferedMessage(bufferedMessage, token)

			return &ExecutionResult{
				Success:      true,
				TokenUpdated: false,
				NextElements: nextElements,
				Completed:    false,
			}, nil
		}
	}

	// No buffered message found, create subscription and wait
	// Буферизованное сообщение не найдено, создаем подписку и ждем
	if messageName != "" {
		logger.Info("Creating message subscription for receive task",
			logger.String("token_id", token.TokenID),
			logger.String("task_name", taskName),
			logger.String("message_name", messageName),
			logger.String("correlation_key", correlationKey))

		// Extract process version from token's ProcessKey
		processVersion := extractVersionFromKey(token.ProcessKey)

		// Create message subscription
		// Создаем подписку на сообщение
		subscription := &models.ProcessMessageSubscription{
			ID:                   models.GenerateID(),
			TenantID:             "DEFAULT_TENANT",
			ProcessDefinitionKey: token.ProcessKey,
			ProcessVersion:       int32(processVersion), // Use actual version from ProcessKey
			StartEventID:         token.CurrentElementID, // This is the receive task ID
			MessageName:          messageName,
			CorrelationKey:       correlationKey,
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := rte.processComponent.CreateMessageSubscription(subscription); err != nil {
			logger.Error("Failed to create message subscription for receive task",
				logger.String("token_id", token.TokenID),
				logger.String("task_name", taskName),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("failed to create message subscription: %v", err),
				Completed: false,
			}, nil
		}

		logger.Info("Message subscription created for receive task - waiting for correlation",
			logger.String("token_id", token.TokenID),
			logger.String("task_name", taskName),
			logger.String("subscription_id", subscription.ID),
			logger.String("message_name", messageName))

		// Set token to waiting state
		// Устанавливаем токен в состояние ожидания
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: nextElements,
			WaitingFor:   fmt.Sprintf("message:%s", messageName),
			Completed:    false,
		}, nil
	}

	// No message name found - cannot create subscription
	// Имя сообщения не найдено - нельзя создать подписку
	logger.Error("Cannot create message subscription - no message name found",
		logger.String("token_id", token.TokenID),
		logger.String("task_name", taskName),
		logger.String("element_id", token.CurrentElementID))

	return &ExecutionResult{
		Success:   false,
		Error:     "receive task: no message name found in task definition",
		Completed: false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (rte *ReceiveTaskExecutor) GetElementType() string {
	return "receiveTask"
}

// isMessageCorrelatedToken checks if token was activated by message correlation
// Проверяет был ли токен активирован через message correlation
func (rte *ReceiveTaskExecutor) isMessageCorrelatedToken(token *models.Token) bool {
	// Check if token has correlation data indicating it came from message correlation
	// Проверяем есть ли у токена correlation данные указывающие что он пришел из message correlation
	if token.Variables != nil {
		if correlatedBy, exists := token.Variables["_correlatedBy"]; exists {
			if correlatedByStr, ok := correlatedBy.(string); ok {
				return correlatedByStr == "message"
			}
		}
	}
	return false
}

// proceedToNextElements proceeds to next elements after message correlation
// Переходит к следующим элементам после message correlation
func (rte *ReceiveTaskExecutor) proceedToNextElements(element map[string]interface{}) (*ExecutionResult, error) {
	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		// Receive task without outgoing flows completes the token
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

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// createBoundaryTimers creates boundary timers for activity
// Создает boundary таймеры для активности
func (rte *ReceiveTaskExecutor) createBoundaryTimers(token *models.Token, element map[string]interface{}) error {
	if rte.processComponent == nil {
		return nil // No process component available
	}

	// Get BPMN process for this token
	// Получаем BPMN процесс для данного токена
	bpmnProcess, err := rte.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this activity
	// Находим boundary события прикрепленные к данной активности
	boundaryEvents := rte.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil // No boundary events found
	}

	logger.Info("Found boundary events for receive task",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create timers for timer boundary events
	// Создаем таймеры для timer boundary событий
	for eventID, boundaryEvent := range boundaryEvents {
		if err := rte.createBoundaryTimerForEvent(token, eventID, boundaryEvent); err != nil {
			logger.Error("Failed to create boundary timer",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue // Continue with other events
		}
	}

	return nil
}

// createErrorBoundaries creates error boundary subscriptions for activity
// Создает подписки на граничные события ошибок для активности
func (rte *ReceiveTaskExecutor) createErrorBoundaries(token *models.Token, element map[string]interface{}) error {
	if rte.processComponent == nil {
		return nil // No process component available
	}

	// Get BPMN process for this token
	// Получаем BPMN процесс для данного токена
	bpmnProcess, err := rte.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this activity
	// Находим boundary события прикрепленные к данной активности
	boundaryEvents := rte.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil // No boundary events found
	}

	logger.Info("Found boundary events for receive task error boundary registration",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create error boundary subscriptions for error boundary events
	// Создаем подписки на граничные события ошибок для error boundary событий
	for eventID, boundaryEvent := range boundaryEvents {
		if err := rte.createErrorBoundaryForEvent(token, eventID, boundaryEvent, bpmnProcess); err != nil {
			logger.Error("Failed to create error boundary subscription",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue // Continue with other events
		}
	}

	return nil
}

// Helper methods for boundary events (similar to SendTaskExecutor)
// Вспомогательные методы для boundary событий (аналогично SendTaskExecutor)

func (rte *ReceiveTaskExecutor) findBoundaryEventsForActivity(activityID string, bpmnProcess map[string]interface{}) map[string]map[string]interface{} {
	boundaryEvents := make(map[string]map[string]interface{})

	elements, exists := bpmnProcess["elements"]
	if !exists {
		return boundaryEvents
	}

	elementsMap, ok := elements.(map[string]interface{})
	if !ok {
		return boundaryEvents
	}

	// Search through all elements for boundary events
	// Ищем среди всех элементов boundary события
	for elementID, element := range elementsMap {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}

		elementType, exists := elementMap["type"]
		if !exists || elementType != "boundaryEvent" {
			continue
		}

		// Check if this boundary event is attached to our activity
		// Проверяем прикреплено ли данное boundary событие к нашей активности
		attachedToRef, exists := elementMap["attached_to_ref"]
		if exists && attachedToRef == activityID {
			boundaryEvents[elementID] = elementMap
		}
	}

	return boundaryEvents
}

func (rte *ReceiveTaskExecutor) createBoundaryTimerForEvent(token *models.Token, eventID string, boundaryEvent map[string]interface{}) error {
	logger.Info("Creating boundary timer for receive task",
		logger.String("token_id", token.TokenID),
		logger.String("event_id", eventID),
		logger.String("activity_id", token.CurrentElementID))

	// Check if this boundary event has timer definition
	// Проверяем есть ли у данного boundary события timer определение
	eventDefinitions, exists := boundaryEvent["event_definitions"]
	if !exists {
		return nil // No event definitions
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil // Invalid event definitions format
	}

	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is timer event definition
		// Проверяем является ли это timer event определением
		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "timerEventDefinition" {
			continue
		}

		// Extract timer data
		// Извлекаем timer данные
		timerData, exists := eventDefMap["timer"]
		if !exists {
			continue
		}

		timerMap, ok := timerData.(map[string]interface{})
		if !ok {
			continue
		}

		// Create timer request
		// Создаем запрос таймера
		timerRequest := &TimerRequest{
			ElementID:         eventID,
			TokenID:           token.TokenID, // Parent token ID for boundary context
			ProcessInstanceID: token.ProcessInstanceID,
			ProcessKey:        token.ProcessKey,
		}

		// Set timer definition based on type with FEEL expression evaluation
		// Устанавливаем timer определение в зависимости от типа с evaluation FEEL expressions
		if duration, exists := timerMap["duration"]; exists {
			if durationStr, ok := duration.(string); ok {
				evaluatedDuration, err := rte.evaluateTimerExpression(durationStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer duration expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", durationStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer duration: %w", err)
				}
				evaluatedDurationStr := fmt.Sprintf("%v", evaluatedDuration)
				timerRequest.TimeDuration = &evaluatedDurationStr
				logger.Debug("Boundary timer duration evaluated",
					logger.String("original", durationStr),
					logger.String("evaluated", evaluatedDurationStr))
			}
		} else if cycle, exists := timerMap["cycle"]; exists {
			if cycleStr, ok := cycle.(string); ok {
				evaluatedCycle, err := rte.evaluateTimerExpression(cycleStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer cycle expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", cycleStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer cycle: %w", err)
				}
				evaluatedCycleStr := fmt.Sprintf("%v", evaluatedCycle)
				timerRequest.TimeCycle = &evaluatedCycleStr
				logger.Debug("Boundary timer cycle evaluated",
					logger.String("original", cycleStr),
					logger.String("evaluated", evaluatedCycleStr))
			}
		} else if date, exists := timerMap["date"]; exists {
			if dateStr, ok := date.(string); ok {
				evaluatedDate, err := rte.evaluateTimerExpression(dateStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer date expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", dateStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer date: %w", err)
				}
				evaluatedDateStr := fmt.Sprintf("%v", evaluatedDate)
				timerRequest.TimeDate = &evaluatedDateStr
				logger.Debug("Boundary timer date evaluated",
					logger.String("original", dateStr),
					logger.String("evaluated", evaluatedDateStr))
			}
		}

		// Create boundary timer via process component
		// Создаем boundary таймер через process компонент
		timerID, err := rte.processComponent.CreateBoundaryTimerWithID(timerRequest)
		if err != nil {
			return fmt.Errorf("failed to create boundary timer: %w", err)
		}

		logger.Info("Boundary timer created for receive task",
			logger.String("parent_token_id", token.TokenID),
			logger.String("timer_id", timerID),
			logger.String("event_id", eventID),
			logger.String("activity_id", token.CurrentElementID))

		// Associate boundary timer with parent token
		// Связываем boundary таймер с родительским токеном
		if err := rte.processComponent.LinkBoundaryTimerToToken(token.TokenID, timerID); err != nil {
			logger.Error("Failed to link boundary timer to token",
				logger.String("parent_token_id", token.TokenID),
				logger.String("timer_id", timerID),
				logger.String("error", err.Error()))
			// Continue execution - linking is not critical
		}
	}

	return nil
}

func (rte *ReceiveTaskExecutor) createErrorBoundaryForEvent(token *models.Token, eventID string, boundaryEvent interface{}, bpmnProcess map[string]interface{}) error {
	boundaryEventMap, ok := boundaryEvent.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid boundary event structure")
	}

	// Check if this is an error boundary event
	eventDefinitions, exists := boundaryEventMap["event_definitions"]
	if !exists {
		return nil // No event definitions - skip
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil // Invalid event definitions structure - skip
	}

	// Look for errorEventDefinition
	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "errorEventDefinition" {
			continue // Not an error event definition
		}

		// This is an error boundary event - create subscription
		logger.Info("Creating error boundary subscription for receive task",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("activity_id", token.CurrentElementID))

		// Extract error reference and resolve error code
		errorCode, errorName := rte.extractErrorInfo(eventDefMap, bpmnProcess)

		// Check if this boundary event is interrupting
		cancelActivity := true // Default is interrupting
		if cancelActivityAttr, exists := boundaryEventMap["cancel_activity"]; exists {
			if cancelActivityBool, ok := cancelActivityAttr.(bool); ok {
				cancelActivity = cancelActivityBool
			} else if cancelActivityStr, ok := cancelActivityAttr.(string); ok {
				cancelActivity = cancelActivityStr != "false"
			}
		}

		// Get outgoing sequence flows from boundary event
		outgoingFlows := rte.getOutgoingFlows(boundaryEventMap)

		// Create error boundary subscription
		subscription := &ErrorBoundarySubscription{
			TokenID:       token.TokenID,
			ElementID:     eventID,
			AttachedToRef: token.CurrentElementID,
			// ErrorRef:       "", // DEAD CODE: ErrorRef field not used anywhere in codebase
			ErrorCode:      errorCode,
			ErrorName:      errorName,
			CancelActivity: cancelActivity,
			OutgoingFlows:  outgoingFlows,
		}

		// Register error boundary subscription
		rte.processComponent.RegisterErrorBoundary(subscription)

		logger.Info("Error boundary subscription created for receive task",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("error_code", errorCode),
			logger.Bool("cancel_activity", cancelActivity))

		return nil
	}

	return nil // No error event definition found
}

// getMessageNameByReference gets message name by reference ID
// Получает имя сообщения по ID ссылки
func (rte *ReceiveTaskExecutor) getMessageNameByReference(messageRef string, token *models.Token) string {
	if rte.processComponent == nil {
		return ""
	}

	// Get BPMN process for this token
	bpmnProcess, err := rte.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for message reference",
			logger.String("message_ref", messageRef),
			logger.String("error", err.Error()))
		return ""
	}

	// Look for message definition
	elements, exists := bpmnProcess["elements"]
	if !exists {
		return ""
	}

	elementsMap, ok := elements.(map[string]interface{})
	if !ok {
		return ""
	}

	// Find message by reference ID
	if messageElement, exists := elementsMap[messageRef]; exists {
		if messageMap, ok := messageElement.(map[string]interface{}); ok {
			if messageType, exists := messageMap["type"]; exists && messageType == "message" {
				if messageName, exists := messageMap["name"]; exists {
					if messageNameStr, ok := messageName.(string); ok {
						return messageNameStr
					}
				}
			}
		}
	}

	return ""
}

// extractCorrelationKeyFromMessage extracts correlation key from message definition
// Извлекает correlation key из определения сообщения
func (rte *ReceiveTaskExecutor) extractCorrelationKeyFromMessage(token *models.Token, messageName string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := rte.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for correlation key extraction",
			logger.String("token_id", token.TokenID),
			logger.String("message_name", messageName),
			logger.String("error", err.Error()))
		return ""
	}

	// Look for message definition with matching name
	elements, exists := bpmnProcess["elements"]
	if !exists {
		return ""
	}

	elementsMap, ok := elements.(map[string]interface{})
	if !ok {
		return ""
	}

	// Search for message definition
	for _, element := range elementsMap {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}

		elementType, exists := elementMap["type"]
		if !exists || elementType != "message" {
			continue
		}

		elementName, exists := elementMap["name"]
		if !exists || elementName != messageName {
			continue
		}

		// Found matching message - extract correlation key
		if extensionElements, exists := elementMap["extension_elements"]; exists {
			if extElementsList, ok := extensionElements.([]interface{}); ok {
				for _, extElement := range extElementsList {
					if extElementMap, ok := extElement.(map[string]interface{}); ok {
						if extensions, exists := extElementMap["extensions"]; exists {
							if extensionsList, ok := extensions.([]interface{}); ok {
								for _, extension := range extensionsList {
									if extensionMap, ok := extension.(map[string]interface{}); ok {
										if extensionType, exists := extensionMap["type"]; exists && extensionType == "subscription" {
											if attributes, exists := extensionMap["attributes"]; exists {
												if attributesMap, ok := attributes.(map[string]interface{}); ok {
													if correlationKey, exists := attributesMap["correlationKey"]; exists {
														if correlationKeyStr, ok := correlationKey.(string); ok {
															return correlationKeyStr
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}
