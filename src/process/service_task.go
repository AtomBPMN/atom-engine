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

// ServiceTaskExecutor executes service tasks
// Исполнитель сервисных задач
type ServiceTaskExecutor struct {
	processComponent ComponentInterface
}

// ComponentInterface defines interface for process component access
// Определяет интерфейс для доступа к process компоненту

// JobComponentInterface defines interface for job creation and management
// Определяет интерфейс для создания и управления заданиями
type JobComponentInterface interface {
	CreateJobWithDetails(jobType, processInstanceID, elementID string, customHeaders map[string]string, variables map[string]interface{}) (string, error)
}

// NewServiceTaskExecutor creates new service task executor
// Создает новый исполнитель сервисных задач
func NewServiceTaskExecutor(processComponent ComponentInterface) *ServiceTaskExecutor {
	return &ServiceTaskExecutor{
		processComponent: processComponent,
	}
}

// Execute executes service task
// Выполняет сервисную задачу
func (ste *ServiceTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing service task",
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

	// Extract task definition from extension elements
	taskDefinition, err := ste.extractTaskDefinition(element)
	if err != nil {
		logger.Error("Failed to extract task definition",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to extract task definition: %v", err),
			Completed: false,
		}, nil
	}

	logger.Info("Service task definition extracted",
		logger.String("token_id", token.TokenID),
		logger.String("task_name", taskName),
		logger.String("job_type", taskDefinition.Type),
		logger.Int("retries", taskDefinition.Retries))

	// Extract custom headers from task definition
	customHeaders := ste.extractCustomHeaders(element)

	// Add token ID to variables for job callback
	jobVariables := make(map[string]interface{})
	for k, v := range token.Variables {
		jobVariables[k] = v
	}
	jobVariables["_tokenID"] = token.TokenID

	// Get job component dynamically from process component
	var jobComponent JobComponentInterface
	if ste.processComponent != nil {
		if jobComp := ste.processComponent.GetJobsComponent(); jobComp != nil {
			if jc, ok := jobComp.(JobComponentInterface); ok {
				jobComponent = jc
			}
		}
	}

	// Create job for external worker
	logger.Info("Checking job component availability",
		logger.String("token_id", token.TokenID),
		logger.Bool("hasJobComponent", jobComponent != nil))

	if jobComponent != nil {
		logger.Info("Attempting to create job",
			logger.String("token_id", token.TokenID),
			logger.String("job_type", taskDefinition.Type))

		jobID, err := jobComponent.CreateJobWithDetails(
			taskDefinition.Type,
			token.ProcessInstanceID,
			token.CurrentElementID,
			customHeaders,
			jobVariables,
		)
		if err != nil {
			logger.Error("Failed to create job for service task",
				logger.String("token_id", token.TokenID),
				logger.String("task_name", taskName),
				logger.String("job_type", taskDefinition.Type),
				logger.String("error", err.Error()))
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("failed to create job: %v", err),
				Completed: false,
			}, nil
		}

		logger.Info("Job created for service task",
			logger.String("token_id", token.TokenID),
			logger.String("task_name", taskName),
			logger.String("job_id", jobID),
			logger.String("job_type", taskDefinition.Type))

		// Set token to wait for job completion
		waitingFor := fmt.Sprintf("job:%s", jobID)
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			WaitingFor:   waitingFor,
			Completed:    false,
		}, nil
	}

	// Fallback: if no job component available, simulate immediate completion
	logger.Warn("No job component available, simulating service task completion",
		logger.String("token_id", token.TokenID),
		logger.String("task_name", taskName))

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

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ste *ServiceTaskExecutor) GetElementType() string {
	return "serviceTask"
}

// TaskDefinition represents service task definition
// Представляет определение сервисной задачи
type TaskDefinition struct {
	Type    string `json:"type"`
	Retries int    `json:"retries"`
}

// extractTaskDefinition extracts task definition from element
// Извлекает определение задачи из элемента
func (ste *ServiceTaskExecutor) extractTaskDefinition(element map[string]interface{}) (*TaskDefinition, error) {
	// Look for extension elements
	extensionElements, exists := element["extension_elements"]
	if !exists {
		return nil, fmt.Errorf("no extension elements found")
	}

	// Parse extension elements as array
	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		return nil, fmt.Errorf("extension elements is not an array")
	}

	// Find taskDefinition in extension elements
	for _, extElement := range extElementsList {
		extElementMap, ok := extElement.(map[string]interface{})
		if !ok {
			continue
		}

		extensions, exists := extElementMap["extensions"]
		if !exists {
			continue
		}

		extensionsList, ok := extensions.([]interface{})
		if !ok {
			continue
		}

		for _, ext := range extensionsList {
			extMap, ok := ext.(map[string]interface{})
			if !ok {
				continue
			}

			extType, exists := extMap["type"]
			if !exists || extType != "taskDefinition" {
				continue
			}

			// Found taskDefinition - extract data
			taskDef, exists := extMap["task_definition"]
			if !exists {
				continue
			}

			taskDefMap, ok := taskDef.(map[string]interface{})
			if !ok {
				continue
			}

			jobType, _ := taskDefMap["type"].(string)
			if jobType == "" {
				return nil, fmt.Errorf("task definition missing type")
			}

			retries := 3 // default retries
			if retriesVal, exists := taskDefMap["retries"]; exists {
				if retriesInt, ok := retriesVal.(int); ok {
					retries = retriesInt
				}
			}

			return &TaskDefinition{
				Type:    jobType,
				Retries: retries,
			}, nil
		}
	}

	return nil, fmt.Errorf("taskDefinition not found in extension elements")
}

// extractCustomHeaders extracts custom headers from element
// Извлекает пользовательские заголовки из элемента
func (ste *ServiceTaskExecutor) extractCustomHeaders(element map[string]interface{}) map[string]string {
	customHeaders := make(map[string]string)

	// Look for extension elements with properties
	extensionElements, exists := element["extension_elements"]
	if !exists {
		return customHeaders
	}

	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		return customHeaders
	}

	for _, extElement := range extElementsList {
		extElementMap, ok := extElement.(map[string]interface{})
		if !ok {
			continue
		}

		extensions, exists := extElementMap["extensions"]
		if !exists {
			continue
		}

		extensionsList, ok := extensions.([]interface{})
		if !ok {
			continue
		}

		for _, ext := range extensionsList {
			extMap, ok := ext.(map[string]interface{})
			if !ok {
				continue
			}

			extType, exists := extMap["type"]
			if !exists || extType != "properties" {
				continue
			}

			// Extract properties as custom headers
			// This would contain task headers if they existed in BPMN
			logger.Debug("Found properties extension",
				logger.String("extension_data", fmt.Sprintf("%v", extMap)))
		}
	}

	return customHeaders
}

// createBoundaryTimers creates boundary timers for activity
// Создает boundary таймеры для активности
func (ste *ServiceTaskExecutor) createBoundaryTimers(token *models.Token, element map[string]interface{}) error {
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

	logger.Info("Found boundary events for activity",
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

// findBoundaryEventsForActivity finds boundary events attached to activity
// Находит boundary события прикрепленные к активности
func (ste *ServiceTaskExecutor) findBoundaryEventsForActivity(activityID string, bpmnProcess map[string]interface{}) map[string]map[string]interface{} {
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
func (ste *ServiceTaskExecutor) createBoundaryTimerForEvent(token *models.Token, eventID string, boundaryEvent map[string]interface{}) error {
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

		// Set timer definition based on type
		// Устанавливаем timer определение в зависимости от типа
		if duration, exists := timerMap["duration"]; exists {
			if durationStr, ok := duration.(string); ok {
				timerRequest.TimeDuration = &durationStr
			}
		} else if cycle, exists := timerMap["cycle"]; exists {
			if cycleStr, ok := cycle.(string); ok {
				timerRequest.TimeCycle = &cycleStr
			}
		} else if date, exists := timerMap["date"]; exists {
			if dateStr, ok := date.(string); ok {
				timerRequest.TimeDate = &dateStr
			}
		}

		// Create boundary timer via process component
		// Создаем boundary таймер через process компонент
		timerID, err := ste.processComponent.CreateBoundaryTimerWithID(timerRequest)
		if err != nil {
			return fmt.Errorf("failed to create boundary timer: %w", err)
		}

		logger.Info("Boundary timer created",
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
