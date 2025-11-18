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
)

// SendTaskExecutor executes send tasks
// Исполнитель задач отправки
type SendTaskExecutor struct {
	processComponent ComponentInterface
}

// NewSendTaskExecutor creates new send task executor
// Создает новый исполнитель задач отправки
func NewSendTaskExecutor(processComponent ComponentInterface) *SendTaskExecutor {
	return &SendTaskExecutor{
		processComponent: processComponent,
	}
}

// Execute executes send task with instant message publishing
// Выполняет задачу отправки с мгновенной публикацией сообщения
func (ste *SendTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing send task",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get task name for logging
	taskName, _ := element["name"].(string)
	if taskName == "" {
		taskName = token.CurrentElementID
	}

	// Create boundary timers when token enters activity
	// Создаем boundary таймеры когда токен входит в активность
	if err := ste.createBoundaryTimers(token, element); err != nil {
		logger.Error("Failed to create boundary timers",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - boundary timer creation is not critical
		// Продолжаем выполнение - создание boundary таймеров не критично
	}

	// Create error boundary subscriptions when token enters activity
	// Создаем подписки на граничные события ошибок когда токен входит в активность
	if err := ste.createErrorBoundaries(token, element); err != nil {
		logger.Error("Failed to create error boundary subscriptions",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - error boundary creation is not critical
		// Продолжаем выполнение - создание граничных событий ошибок не критично
	}

	// Extract message information from send_task section
	// Извлекаем информацию о сообщении из секции send_task
	messageName := ""
	logger.Info("DEBUG: Send task element data",
		logger.Any("element", element))

	if sendTaskData, exists := element["send_task"]; exists {
		logger.Info("DEBUG: Found send_task data",
			logger.Any("send_task_data", sendTaskData))

		if sendTaskMap, ok := sendTaskData.(map[string]interface{}); ok {
			if taskType, exists := sendTaskMap["task_type"]; exists {
				if taskTypeStr, ok := taskType.(string); ok {
					messageName = taskTypeStr
					logger.Info("Send task message name extracted from task_type",
						logger.String("message_name", messageName))
				} else {
					logger.Warn("DEBUG: task_type is not string",
						logger.Any("task_type", taskType))
				}
			} else {
				logger.Warn("DEBUG: task_type not found in send_task")
			}
		} else {
			logger.Warn("DEBUG: send_task_data is not map[string]interface{}")
		}
	} else {
		logger.Warn("DEBUG: send_task not found in element")
	}

	// Fallback: try to extract from messageRef if present
	// Запасной вариант: пытаемся извлечь из messageRef если присутствует
	if messageName == "" {
		if sendTaskData, exists := element["send_task"]; exists {
			if sendTaskMap, ok := sendTaskData.(map[string]interface{}); ok {
				if msgRef, exists := sendTaskMap["message_ref"]; exists {
					if msgRefStr, ok := msgRef.(string); ok {
						actualMessageName := ste.getMessageNameByReference(msgRefStr, token)
						if actualMessageName != "" {
							messageName = actualMessageName
							logger.Info("Send task message name resolved from messageRef",
								logger.String("message_ref", msgRefStr),
								logger.String("message_name", messageName))
						}
					}
				}
			}
		}
	}

	// Generate correlation key from token variables or use message name
	// Генерируем ключ корреляции из переменных токена или используем имя сообщения
	correlationKey := messageName
	if token.Variables != nil {
		if corrKey, exists := token.Variables["correlationKey"]; exists {
			if corrKeyStr, ok := corrKey.(string); ok {
				correlationKey = corrKeyStr
			}
		}
	}

	// Publish message instantly through process component
	// Мгновенно публикуем сообщение через process component
	logger.Info("DEBUG: About to publish message",
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey),
		logger.Bool("has_process_component", ste.processComponent != nil))

	if ste.processComponent != nil && messageName != "" {
		result, err := ste.processComponent.PublishMessageWithElementID(
			messageName,
			correlationKey,
			token.CurrentElementID,
			token.Variables,
		)
		if err != nil {
			logger.Error("Failed to publish message from send task",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("element_id", token.CurrentElementID),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Message published from send task",
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey),
				logger.String("element_id", token.CurrentElementID),
				logger.Bool("instance_created", result != nil && result.InstanceCreated))
		}
	} else if messageName == "" {
		logger.Warn("Send task has no message name - skipping message publishing",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))
	}

	// Get outgoing sequence flows and continue immediately
	// Получаем исходящие sequence flows и продолжаем немедленно
	outgoing, exists := element["outgoing"]
	if !exists {
		// Send task without outgoing flows completes the token
		logger.Info("Send task completed - no outgoing flows",
			logger.String("token_id", token.TokenID),
			logger.String("task_name", taskName))
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
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

	logger.Info("Send task continuing execution immediately",
		logger.String("token_id", token.TokenID),
		logger.String("task_name", taskName),
		logger.String("message_name", messageName),
		logger.Int("next_elements", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ste *SendTaskExecutor) GetElementType() string {
	return "sendTask"
}

// createBoundaryTimers creates boundary timers for activity
// Создает boundary таймеры для активности
func (ste *SendTaskExecutor) createBoundaryTimers(token *models.Token, element map[string]interface{}) error {
	if ste.processComponent == nil {
		return nil // No process component available
	}

	// Get BPMN process for this token
	// Получаем BPMN процесс для данного токена
	bpmnProcess, err := ste.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this activity
	// Находим boundary события прикрепленные к данной активности
	boundaryEvents := ste.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil // No boundary events found
	}

	logger.Info("Found boundary events for send task",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create timers for timer boundary events
	// Создаем таймеры для timer boundary событий
	for eventID, boundaryEvent := range boundaryEvents {
		if err := ste.createBoundaryTimerForEvent(token, eventID, boundaryEvent); err != nil {
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
func (ste *SendTaskExecutor) createErrorBoundaries(token *models.Token, element map[string]interface{}) error {
	if ste.processComponent == nil {
		return nil // No process component available
	}

	// Get BPMN process for this token
	// Получаем BPMN процесс для данного токена
	bpmnProcess, err := ste.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this activity
	// Находим boundary события прикрепленные к данной активности
	boundaryEvents := ste.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil // No boundary events found
	}

	logger.Info("Found boundary events for send task error boundary registration",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create error boundary subscriptions for error boundary events
	// Создаем подписки на граничные события ошибок для error boundary событий
	for eventID, boundaryEvent := range boundaryEvents {
		if err := ste.createErrorBoundaryForEvent(token, eventID, boundaryEvent, bpmnProcess); err != nil {
			logger.Error("Failed to create error boundary subscription",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue // Continue with other events
		}
	}

	return nil
}

// findBoundaryEventsForActivity finds boundary events attached to activity
// Находит boundary события прикрепленные к активности
func (ste *SendTaskExecutor) findBoundaryEventsForActivity(
	activityID string,
	bpmnProcess map[string]interface{},
) map[string]map[string]interface{} {
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

// createBoundaryTimerForEvent creates timer for boundary event if it has timer definition
// Создает таймер для boundary события если у него есть timer определение
func (ste *SendTaskExecutor) createBoundaryTimerForEvent(
	token *models.Token,
	eventID string,
	boundaryEvent map[string]interface{},
) error {
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

		// Extract boundary event metadata for proper scope tracking
		// Извлекаем метаданные boundary события для правильного отслеживания scope
		if attachedToRef, exists := boundaryEvent["attached_to_ref"]; exists {
			if attachedStr, ok := attachedToRef.(string); ok {
				timerRequest.AttachedToRef = &attachedStr
			}
		}

		if cancelActivity, exists := boundaryEvent["cancel_activity"]; exists {
			if cancelBool, ok := cancelActivity.(bool); ok {
				timerRequest.CancelActivity = &cancelBool
			}
		}

		// Set timer definition based on type with FEEL expression evaluation
		// Устанавливаем timer определение в зависимости от типа с evaluation FEEL expressions
		if duration, exists := timerMap["duration"]; exists {
			if durationStr, ok := duration.(string); ok {
				evaluatedDuration, err := ste.evaluateTimerExpression(durationStr, token)
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
				evaluatedCycle, err := ste.evaluateTimerExpression(cycleStr, token)
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
				evaluatedDate, err := ste.evaluateTimerExpression(dateStr, token)
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
		timerID, err := ste.processComponent.CreateBoundaryTimerWithID(timerRequest)
		if err != nil {
			return fmt.Errorf("failed to create boundary timer: %w", err)
		}

		logger.Info("Boundary timer created for send task",
			logger.String("parent_token_id", token.TokenID),
			logger.String("timer_id", timerID),
			logger.String("event_id", eventID),
			logger.String("activity_id", token.CurrentElementID))

		// Associate boundary timer with parent token
		// Связываем boundary таймер с родительским токеном
		if err := ste.processComponent.LinkBoundaryTimerToToken(token.TokenID, timerID); err != nil {
			logger.Error("Failed to link boundary timer to token",
				logger.String("parent_token_id", token.TokenID),
				logger.String("timer_id", timerID),
				logger.String("error", err.Error()))
			// Continue execution - linking is not critical
		}
	}

	return nil
}

// createErrorBoundaryForEvent creates error boundary subscription for specific event
// Создает подписку на граничное событие ошибки для конкретного события
func (ste *SendTaskExecutor) createErrorBoundaryForEvent(
	token *models.Token,
	eventID string,
	boundaryEvent interface{},
	bpmnProcess interface{},
) error {
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
		logger.Info("Creating error boundary subscription for send task",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("activity_id", token.CurrentElementID))

		// Extract error reference and resolve error code
		errorCode, errorName := ste.extractErrorInfo(eventDefMap, bpmnProcess)

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
		outgoingFlows := ste.getOutgoingFlows(boundaryEventMap)

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
		ste.processComponent.RegisterErrorBoundary(subscription)

		logger.Info("Error boundary subscription created for send task",
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
func (ste *SendTaskExecutor) getMessageNameByReference(messageRef string, token *models.Token) string {
	if ste.processComponent == nil {
		return ""
	}

	// Get BPMN process for this token
	bpmnProcess, err := ste.processComponent.GetBPMNProcessForToken(token)
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

// extractErrorInfo extracts error code and name from error event definition
// Извлекает код ошибки и имя из определения события ошибки
func (ste *SendTaskExecutor) extractErrorInfo(
	eventDef map[string]interface{},
	bpmnProcess interface{},
) (string, string) {
	// Get error reference from event definition
	errorRef, exists := eventDef["reference"] // Changed from "error_ref" to "reference"
	if !exists {
		return "GENERAL_ERROR", "General Error"
	}

	errorRefStr, ok := errorRef.(string)
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	// Get the complete BPMN structure with all elements
	bpmnProcessMap, ok := bpmnProcess.(map[string]interface{})
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	// Look for the error definition in the elements map (not error_definitions array)
	if elements, exists := bpmnProcessMap["elements"]; exists {
		if elementsMap, ok := elements.(map[string]interface{}); ok {
			// Look for the specific error element by ID
			if errorElement, exists := elementsMap[errorRefStr]; exists {
				if errorDefMap, ok := errorElement.(map[string]interface{}); ok {
					errorCode := "GENERAL_ERROR"
					errorName := "General Error"

					// Extract error_code from the error element
					if code, exists := errorDefMap["error_code"]; exists {
						if codeStr, ok := code.(string); ok {
							errorCode = codeStr
						}
					}

					// Extract name from the error element
					if name, exists := errorDefMap["name"]; exists {
						if nameStr, ok := name.(string); ok {
							errorName = nameStr
						}
					}

					logger.Info("Resolved error definition from elements for send task",
						logger.String("error_ref", errorRefStr),
						logger.String("error_code", errorCode),
						logger.String("error_name", errorName))

					return errorCode, errorName
				}
			}
		}
	}

	logger.Warn("Could not resolve error definition for send task, using default",
		logger.String("error_ref", errorRefStr))
	return "GENERAL_ERROR", "General Error"
}

// getOutgoingFlows extracts outgoing sequence flows from boundary event
// Извлекает исходящие потоки последовательности из граничного события
func (ste *SendTaskExecutor) getOutgoingFlows(boundaryEvent map[string]interface{}) []string {
	outgoing, exists := boundaryEvent["outgoing"]
	if !exists {
		return []string{}
	}

	var flows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				flows = append(flows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		flows = append(flows, outgoingStr)
	}

	return flows
}

// evaluateTimerExpression evaluates timer expressions using expression component
// Вычисляет timer expressions используя expression компонент
func (ste *SendTaskExecutor) evaluateTimerExpression(expression string, token *models.Token) (interface{}, error) {
	// If not a FEEL expression (doesn't start with =), return as is
	// Если не FEEL expression (не начинается с =), возвращаем как есть
	if expression == "" || len(expression) == 0 || expression[0] != '=' {
		return expression, nil
	}

	// Get expression component through process component
	// Получаем expression компонент через process компонент
	if ste.processComponent == nil {
		return nil, fmt.Errorf("process component not available for expression evaluation")
	}

	// Get core interface
	core := ste.processComponent.GetCore()
	if core == nil {
		return nil, fmt.Errorf("core interface not available for expression evaluation")
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		return nil, fmt.Errorf("expression component not available")
	}

	// Cast to expression evaluator interface with EvaluateExpressionEngine method
	// Приводим к интерфейсу expression evaluator с методом EvaluateExpressionEngine
	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		return nil, fmt.Errorf("failed to cast expression component to ExpressionEvaluator interface")
	}

	// Evaluate FEEL expression using expression engine
	// Вычисляем FEEL expression используя expression engine
	result, err := expressionComp.EvaluateExpressionEngine(expression, token.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate FEEL expression '%s': %w", expression, err)
	}

	logger.Debug("Boundary timer expression evaluated successfully for send task",
		logger.String("token_id", token.TokenID),
		logger.String("original_expression", expression),
		logger.Any("evaluated_result", result),
		logger.Any("token_variables", token.Variables))

	return result, nil
}
