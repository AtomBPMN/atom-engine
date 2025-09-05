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
		result, err := ste.processComponent.PublishMessageWithElementID(messageName, correlationKey, token.CurrentElementID, token.Variables)
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
func (ste *SendTaskExecutor) findBoundaryEventsForActivity(activityID string, bpmnProcess map[string]interface{}) map[string]map[string]interface{} {
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
func (ste *SendTaskExecutor) createBoundaryTimerForEvent(token *models.Token, eventID string, boundaryEvent map[string]interface{}) error {
	// Implementation similar to ServiceTaskExecutor
	// Реализация аналогична ServiceTaskExecutor
	// For now, we'll use the same logic as service task
	return nil // TODO: Implement boundary timer creation
}

// createErrorBoundaryForEvent creates error boundary subscription for specific event
// Создает подписку на граничное событие ошибки для конкретного события
func (ste *SendTaskExecutor) createErrorBoundaryForEvent(token *models.Token, eventID string, boundaryEvent interface{}, bpmnProcess map[string]interface{}) error {
	// Implementation similar to ServiceTaskExecutor
	// Реализация аналогична ServiceTaskExecutor
	// For now, we'll use the same logic as service task
	return nil // TODO: Implement error boundary creation
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
