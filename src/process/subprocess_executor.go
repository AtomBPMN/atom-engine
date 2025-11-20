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

// SubProcessExecutor executes embedded subprocesses
// Исполнитель встроенных подпроцессов
type SubProcessExecutor struct {
	component ComponentInterface
}

// StartEventInfo contains information about subprocess start event
// Информация о start event подпроцесса
type StartEventInfo struct {
	ID   string // Start event ID
	Type string // Start event type: none, message, timer, signal
}

// NewSubProcessExecutor creates new subprocess executor
// Создает новый исполнитель subprocess
func NewSubProcessExecutor(component ComponentInterface) *SubProcessExecutor {
	return &SubProcessExecutor{
		component: component,
	}
}

// Execute executes subprocess
// Выполняет subprocess
func (spe *SubProcessExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing subprocess",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get subprocess name for logging
	subprocessName, _ := element["name"].(string)
	if subprocessName == "" {
		subprocessName = token.CurrentElementID
	}

	// Check if subprocess was already executed for this token
	subprocessKey := fmt.Sprintf("subprocess_executed:%s", token.CurrentElementID)
	if executed, exists := token.GetExecutionContext(subprocessKey); exists && executed == true {
		logger.Info("Subprocess already executed, continuing to next elements",
			logger.String("token_id", token.TokenID),
			logger.String("subprocess_name", subprocessName),
			logger.String("element_id", token.CurrentElementID))

		// Subprocess completed, continue to next elements
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

	// Create boundary timers when token enters subprocess
	// Создаем boundary таймеры когда токен входит в subprocess
	if err := spe.createBoundaryTimers(token, element); err != nil {
		logger.Error("Failed to create boundary timers",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - boundary timer creation is not critical
	}

	// Create error boundary subscriptions when token enters subprocess
	// Создаем подписки на граничные события ошибок когда токен входит в subprocess
	if err := spe.createErrorBoundaries(token, element); err != nil {
		logger.Error("Failed to create error boundary subscriptions",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		// Continue execution - error boundary creation is not critical
	}

	// Get BPMN process to find subprocess internal startEvents
	bpmnProcess, err := spe.component.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for subprocess",
			logger.String("token_id", token.TokenID),
			logger.String("subprocess_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get BPMN process: %v", err),
		}, nil
	}

	// Find all internal startEvents of subprocess
	startEvents, err := spe.findSubProcessStartEvents(token.CurrentElementID, bpmnProcess)
	if err != nil {
		logger.Error("Failed to find subprocess start events",
			logger.String("token_id", token.TokenID),
			logger.String("subprocess_id", token.CurrentElementID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to find subprocess start events: %v", err),
		}, nil
	}

	logger.Info("Found subprocess start events",
		logger.String("token_id", token.TokenID),
		logger.String("subprocess_id", token.CurrentElementID),
		logger.Int("count", len(startEvents)))

	// Apply input variable mapping if exists
	subprocessVariables := spe.applyInputMapping(token.Variables, element)

	// Mark subprocess as executed for this element
	token.SetExecutionContext(subprocessKey, true)

	// Set parent token to wait for subprocess completion
	waitingFor := fmt.Sprintf("subprocess:%s", token.CurrentElementID)

	logger.Info("Subprocess starting, parent token waiting",
		logger.String("parent_token_id", token.TokenID),
		logger.String("waiting_for", waitingFor))

	// Process each start event based on its type
	for _, startEventInfo := range startEvents {
		switch startEventInfo.Type {
		case "none":
			// Handle none start event - create token and execute immediately
			if err := spe.handleNoneStartEvent(token, startEventInfo, subprocessVariables, subprocessName); err != nil {
				logger.Error("Failed to handle none start event",
					logger.String("subprocess_token_id", token.TokenID),
					logger.String("start_event_id", startEventInfo.ID),
					logger.String("error", err.Error()))
				return &ExecutionResult{
					Success: false,
					Error:   fmt.Sprintf("failed to handle none start event: %v", err),
				}, nil
			}
		case "message":
			// Handle message start event - create subscription
			if err := spe.handleMessageStartEvent(token, startEventInfo, element, subprocessVariables); err != nil {
				logger.Error("Failed to handle message start event",
					logger.String("subprocess_token_id", token.TokenID),
					logger.String("start_event_id", startEventInfo.ID),
					logger.String("error", err.Error()))
				// Continue with other start events
			}
		case "timer":
			// Handle timer start event - schedule timer
			if err := spe.handleTimerStartEvent(token, startEventInfo, element, subprocessVariables); err != nil {
				logger.Error("Failed to handle timer start event",
					logger.String("subprocess_token_id", token.TokenID),
					logger.String("start_event_id", startEventInfo.ID),
					logger.String("error", err.Error()))
				// Continue with other start events
			}
		case "signal":
			// Handle signal start event - create subscription
			if err := spe.handleSignalStartEvent(token, startEventInfo, element, subprocessVariables); err != nil {
				logger.Error("Failed to handle signal start event",
					logger.String("subprocess_token_id", token.TokenID),
					logger.String("start_event_id", startEventInfo.ID),
					logger.String("error", err.Error()))
				// Continue with other start events
			}
		default:
			logger.Warn("Unknown start event type in subprocess",
				logger.String("start_event_id", startEventInfo.ID),
				logger.String("event_type", startEventInfo.Type))
		}
	}

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		WaitingFor:   waitingFor,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (spe *SubProcessExecutor) GetElementType() string {
	return "subProcess"
}

// findSubProcessStartEvent finds startEvent inside subprocess
// Находит startEvent внутри subprocess
func (spe *SubProcessExecutor) findSubProcessStartEvents(
	subprocessID string,
	bpmnProcess map[string]interface{},
) ([]StartEventInfo, error) {
	elements, exists := bpmnProcess["elements"]
	if !exists {
		return nil, fmt.Errorf("no elements in BPMN process")
	}

	elementsMap, ok := elements.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid elements structure")
	}

	startEvents := []StartEventInfo{}

	// Search through all elements to find startEvents with subprocess as parent_scope
	for elementID, element := range elementsMap {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}

		elementType, exists := elementMap["type"]
		if !exists || elementType != "startEvent" {
			continue
		}

		// Check parent_scope - must match subprocess ID
		parentScope, hasParentScope := elementMap["parent_scope"]
		if !hasParentScope {
			continue
		}

		parentScopeStr, ok := parentScope.(string)
		if !ok || parentScopeStr != subprocessID {
			continue
		}

		// Determine start event type
		eventType := spe.determineStartEventType(elementMap)

		logger.Info("Found start event in subprocess",
			logger.String("subprocess_id", subprocessID),
			logger.String("start_event_id", elementID),
			logger.String("event_type", eventType))

		startEvents = append(startEvents, StartEventInfo{
			ID:   elementID,
			Type: eventType,
		})
	}

	if len(startEvents) == 0 {
		return nil, fmt.Errorf("no start events found for subprocess %s", subprocessID)
	}

	logger.Info("Found start events for subprocess",
		logger.String("subprocess_id", subprocessID),
		logger.Int("count", len(startEvents)))

	return startEvents, nil
}

// determineStartEventType determines type of start event from event definitions
// Определяет тип start event из event definitions
func (spe *SubProcessExecutor) determineStartEventType(elementMap map[string]interface{}) string {
	eventDefinitions, hasEventDefs := elementMap["event_definitions"]
	if !hasEventDefs {
		return "none"
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok || len(eventDefList) == 0 {
		return "none"
	}

	// Check first event definition
	eventDef, ok := eventDefList[0].(map[string]interface{})
	if !ok {
		return "none"
	}

	eventType, exists := eventDef["type"]
	if !exists {
		return "none"
	}

	eventTypeStr, ok := eventType.(string)
	if !ok {
		return "none"
	}

	// Map BPMN event definition types to simple types
	switch eventTypeStr {
	case "messageEventDefinition":
		return "message"
	case "timerEventDefinition":
		return "timer"
	case "signalEventDefinition":
		return "signal"
	default:
		return "none"
	}
}

// handleNoneStartEvent handles none start event in subprocess
// Обрабатывает none start event в подпроцессе
func (spe *SubProcessExecutor) handleNoneStartEvent(
	parentToken *models.Token,
	startEventInfo StartEventInfo,
	subprocessVariables map[string]interface{},
	subprocessName string,
) error {
	// Create new token inside subprocess at startEvent
	subprocessToken := models.NewToken(
		parentToken.ProcessInstanceID,
		parentToken.ProcessKey,
		startEventInfo.ID,
	)
	subprocessToken.Variables = subprocessVariables
	subprocessToken.ParentTokenID = parentToken.TokenID
	subprocessToken.SubProcessID = parentToken.CurrentElementID

	logger.Info("Creating subprocess token for none start event",
		logger.String("parent_token_id", parentToken.TokenID),
		logger.String("subprocess_token_id", subprocessToken.TokenID),
		logger.String("subprocess_id", parentToken.CurrentElementID),
		logger.String("start_event_id", startEventInfo.ID))

	// Save subprocess token to storage
	if err := spe.component.GetStorage().SaveToken(subprocessToken); err != nil {
		logger.Error("Failed to save subprocess token",
			logger.String("parent_token_id", parentToken.TokenID),
			logger.String("subprocess_token_id", subprocessToken.TokenID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to save subprocess token: %w", err)
	}

	// Execute subprocess start event immediately
	if err := spe.component.ExecuteToken(subprocessToken); err != nil {
		logger.Error("Failed to execute subprocess start event",
			logger.String("subprocess_token_id", subprocessToken.TokenID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to execute subprocess start event: %w", err)
	}

	return nil
}

// handleMessageStartEvent handles message start event in subprocess
// Обрабатывает message start event в подпроцессе
func (spe *SubProcessExecutor) handleMessageStartEvent(
	parentToken *models.Token,
	startEventInfo StartEventInfo,
	element map[string]interface{},
	subprocessVariables map[string]interface{},
) error {
	logger.Info("Message start event in subprocess - creating subscription",
		logger.String("parent_token_id", parentToken.TokenID),
		logger.String("subprocess_id", parentToken.CurrentElementID),
		logger.String("start_event_id", startEventInfo.ID))

	// TODO: Implement message subscription for subprocess start event
	// For now, treat as none start event
	return spe.handleNoneStartEvent(parentToken, startEventInfo, subprocessVariables, "")
}

// handleTimerStartEvent handles timer start event in subprocess
// Обрабатывает timer start event в подпроцессе
func (spe *SubProcessExecutor) handleTimerStartEvent(
	parentToken *models.Token,
	startEventInfo StartEventInfo,
	element map[string]interface{},
	subprocessVariables map[string]interface{},
) error {
	logger.Info("Timer start event in subprocess - scheduling timer",
		logger.String("parent_token_id", parentToken.TokenID),
		logger.String("subprocess_id", parentToken.CurrentElementID),
		logger.String("start_event_id", startEventInfo.ID))

	// TODO: Implement timer scheduling for subprocess start event
	// For now, treat as none start event
	return spe.handleNoneStartEvent(parentToken, startEventInfo, subprocessVariables, "")
}

// handleSignalStartEvent handles signal start event in subprocess
// Обрабатывает signal start event в подпроцессе
func (spe *SubProcessExecutor) handleSignalStartEvent(
	parentToken *models.Token,
	startEventInfo StartEventInfo,
	element map[string]interface{},
	subprocessVariables map[string]interface{},
) error {
	logger.Info("Signal start event in subprocess - creating subscription",
		logger.String("parent_token_id", parentToken.TokenID),
		logger.String("subprocess_id", parentToken.CurrentElementID),
		logger.String("start_event_id", startEventInfo.ID))

	// TODO: Implement signal subscription for subprocess start event
	// For now, treat as none start event
	return spe.handleNoneStartEvent(parentToken, startEventInfo, subprocessVariables, "")
}

// applyInputMapping applies input variable mapping to subprocess variables
// Применяет input variable mapping к переменным subprocess
func (spe *SubProcessExecutor) applyInputMapping(
	variables map[string]interface{},
	element map[string]interface{},
) map[string]interface{} {
	// For now, pass all variables to subprocess
	// Full implementation would parse zeebe:ioMapping from extension_elements
	subprocessVars := make(map[string]interface{})
	for k, v := range variables {
		subprocessVars[k] = v
	}

	logger.Debug("Applied input mapping to subprocess",
		logger.Int("input_variables", len(variables)),
		logger.Int("subprocess_variables", len(subprocessVars)))

	return subprocessVars
}

// applyOutputMapping applies output variable mapping from subprocess to parent
// Применяет output variable mapping от subprocess к родителю
func (spe *SubProcessExecutor) applyOutputMapping(
	subprocessVars map[string]interface{},
	parentVars map[string]interface{},
	element map[string]interface{},
) map[string]interface{} {
	// For now, merge all subprocess variables into parent
	// Full implementation would parse zeebe:ioMapping from extension_elements
	result := make(map[string]interface{})
	
	// Copy parent variables
	for k, v := range parentVars {
		result[k] = v
	}
	
	// Merge subprocess variables
	for k, v := range subprocessVars {
		result[k] = v
	}

	logger.Debug("Applied output mapping from subprocess",
		logger.Int("subprocess_variables", len(subprocessVars)),
		logger.Int("parent_variables", len(parentVars)),
		logger.Int("result_variables", len(result)))

	return result
}

// createBoundaryTimers creates boundary timers for subprocess
// Создает boundary таймеры для subprocess
func (spe *SubProcessExecutor) createBoundaryTimers(
	token *models.Token,
	element map[string]interface{},
) error {
	if spe.component == nil {
		return nil
	}

	// Get BPMN process for this token
	bpmnProcess, err := spe.component.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this subprocess
	boundaryEvents := spe.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil
	}

	logger.Info("Found boundary events for subprocess",
		logger.String("subprocess_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create timers for timer boundary events
	for eventID, boundaryEvent := range boundaryEvents {
		if err := spe.createBoundaryTimerForEvent(token, eventID, boundaryEvent); err != nil {
			logger.Error("Failed to create boundary timer",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue
		}
	}

	return nil
}

// findBoundaryEventsForActivity finds boundary events attached to activity
// Находит boundary события прикрепленные к активности
func (spe *SubProcessExecutor) findBoundaryEventsForActivity(
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
		attachedToRef, exists := elementMap["attached_to_ref"]
		if exists && attachedToRef == activityID {
			boundaryEvents[elementID] = elementMap
		}
	}

	return boundaryEvents
}

// createBoundaryTimerForEvent creates timer for boundary event if it has timer definition
// Создает таймер для boundary события если у него есть timer определение
func (spe *SubProcessExecutor) createBoundaryTimerForEvent(
	token *models.Token,
	eventID string,
	boundaryEvent map[string]interface{},
) error {
	// Check if this boundary event has timer definition
	eventDefinitions, exists := boundaryEvent["event_definitions"]
	if !exists {
		return nil
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil
	}

	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is timer event definition
		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "timerEventDefinition" {
			continue
		}

		// Extract timer data
		timerData, exists := eventDefMap["timer"]
		if !exists {
			continue
		}

		timerMap, ok := timerData.(map[string]interface{})
		if !ok {
			continue
		}

		// Create timer request
		timerRequest := &TimerRequest{
			ElementID:         eventID,
			TokenID:           token.TokenID,
			ProcessInstanceID: token.ProcessInstanceID,
			ProcessKey:        token.ProcessKey,
		}

		// Extract boundary event metadata
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

		// Extract timer expression and set appropriate field
		if timeDuration, exists := timerMap["time_duration"]; exists {
			if durationStr, ok := timeDuration.(string); ok {
				timerRequest.TimeDuration = &durationStr
			}
		} else if timeCycle, exists := timerMap["time_cycle"]; exists {
			if cycleStr, ok := timeCycle.(string); ok {
				timerRequest.TimeCycle = &cycleStr
			}
		} else if timeDate, exists := timerMap["time_date"]; exists {
			if dateStr, ok := timeDate.(string); ok {
				timerRequest.TimeDate = &dateStr
			}
		}

		if timerRequest.TimeDuration == nil && timerRequest.TimeCycle == nil && timerRequest.TimeDate == nil {
			continue
		}

		// Create boundary timer via process component
		timerID, err := spe.component.CreateBoundaryTimerWithID(timerRequest)
		if err != nil {
			return fmt.Errorf("failed to create boundary timer: %w", err)
		}

		logger.Info("Boundary timer created for subprocess",
			logger.String("parent_token_id", token.TokenID),
			logger.String("timer_id", timerID),
			logger.String("event_id", eventID),
			logger.String("subprocess_id", token.CurrentElementID))

		// Associate boundary timer with parent token
		if err := spe.component.LinkBoundaryTimerToToken(token.TokenID, timerID); err != nil {
			logger.Error("Failed to link boundary timer to token",
				logger.String("parent_token_id", token.TokenID),
				logger.String("timer_id", timerID),
				logger.String("error", err.Error()))
		}
	}

	return nil
}

// createErrorBoundaries creates error boundary subscriptions for subprocess
// Создает подписки на граничные события ошибок для subprocess
func (spe *SubProcessExecutor) createErrorBoundaries(
	token *models.Token,
	element map[string]interface{},
) error {
	if spe.component == nil {
		return nil
	}

	// Get BPMN process for this token
	bpmnProcess, err := spe.component.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	// Find boundary events attached to this subprocess
	boundaryEvents := spe.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil
	}

	logger.Info("Found boundary events for error boundary registration",
		logger.String("subprocess_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	// Create error boundary subscriptions
	for eventID, boundaryEvent := range boundaryEvents {
		if err := spe.createErrorBoundaryForEvent(token, eventID, boundaryEvent, bpmnProcess); err != nil {
			logger.Error("Failed to create error boundary subscription",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue
		}
	}

	return nil
}

// createErrorBoundaryForEvent creates error boundary subscription for specific event
// Создает подписку на граничное событие ошибки для конкретного события
func (spe *SubProcessExecutor) createErrorBoundaryForEvent(
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
		return nil
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil
	}

	// Look for errorEventDefinition
	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "errorEventDefinition" {
			continue
		}

		// This is an error boundary event
		logger.Info("Creating error boundary subscription for subprocess",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("subprocess_id", token.CurrentElementID))

		// Extract error info
		errorCode, errorName := spe.extractErrorInfo(eventDefMap, bpmnProcess)

		// Check if interrupting
		cancelActivity := true
		if cancelActivityAttr, exists := boundaryEventMap["cancel_activity"]; exists {
			if cancelActivityBool, ok := cancelActivityAttr.(bool); ok {
				cancelActivity = cancelActivityBool
			} else if cancelActivityStr, ok := cancelActivityAttr.(string); ok {
				cancelActivity = cancelActivityStr != "false"
			}
		}

		// Get outgoing flows
		outgoingFlows := spe.getOutgoingFlows(boundaryEventMap)

		// Create subscription
		subscription := &ErrorBoundarySubscription{
			TokenID:        token.TokenID,
			ElementID:      eventID,
			AttachedToRef:  token.CurrentElementID,
			ErrorCode:      errorCode,
			ErrorName:      errorName,
			CancelActivity: cancelActivity,
			OutgoingFlows:  outgoingFlows,
		}

		spe.component.RegisterErrorBoundary(subscription)

		logger.Info("Error boundary subscription created for subprocess",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("error_code", errorCode),
			logger.Bool("cancel_activity", cancelActivity))

		return nil
	}

	return nil
}

// extractErrorInfo extracts error code and name from error event definition
// Извлекает код ошибки и имя из определения события ошибки
func (spe *SubProcessExecutor) extractErrorInfo(
	eventDef map[string]interface{},
	bpmnProcess interface{},
) (string, string) {
	errorRef, exists := eventDef["reference"]
	if !exists {
		return "GENERAL_ERROR", "General Error"
	}

	errorRefStr, ok := errorRef.(string)
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	bpmnProcessMap, ok := bpmnProcess.(map[string]interface{})
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	if elements, exists := bpmnProcessMap["elements"]; exists {
		if elementsMap, ok := elements.(map[string]interface{}); ok {
			if errorElement, exists := elementsMap[errorRefStr]; exists {
				if errorDefMap, ok := errorElement.(map[string]interface{}); ok {
					errorCode := "GENERAL_ERROR"
					errorName := "General Error"

					if code, exists := errorDefMap["error_code"]; exists {
						if codeStr, ok := code.(string); ok {
							errorCode = codeStr
						}
					}

					if name, exists := errorDefMap["name"]; exists {
						if nameStr, ok := name.(string); ok {
							errorName = nameStr
						}
					}

					return errorCode, errorName
				}
			}
		}
	}

	return "GENERAL_ERROR", "General Error"
}

// getOutgoingFlows gets outgoing sequence flows from boundary event
// Получает исходящие sequence flows из boundary события
func (spe *SubProcessExecutor) getOutgoingFlows(boundaryEvent map[string]interface{}) []string {
	var outgoingFlows []string

	outgoing, exists := boundaryEvent["outgoing"]
	if !exists {
		return outgoingFlows
	}

	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				outgoingFlows = append(outgoingFlows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		outgoingFlows = append(outgoingFlows, outgoingStr)
	}

	return outgoingFlows
}

