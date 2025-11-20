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
	"atom-engine/src/storage"
)

// ExecutionProcessor handles execution result processing
// Обрабатывает результаты выполнения
type ExecutionProcessor struct {
	storage   storage.Storage
	component ComponentInterface
}

// NewExecutionProcessor creates new execution processor
// Создает новый процессор выполнения
func NewExecutionProcessor(storage storage.Storage, component ComponentInterface) *ExecutionProcessor {
	return &ExecutionProcessor{
		storage:   storage,
		component: component,
	}
}

// processExecutionResult processes the result of element execution
// Обрабатывает результат выполнения элемента
func (ep *ExecutionProcessor) processExecutionResult(
	token *models.Token,
	result *ExecutionResult,
	bpmnProcess *models.BPMNProcess,
) error {
	// Update token variables if provided
	if result.Variables != nil {
		token.MergeVariables(result.Variables)
	}

	// Handle timer request from intermediate catch events
	if result.TimerRequest != nil {
		logger.Info("Processing timer request",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", result.TimerRequest.ElementID))

		if err := ep.component.CreateTimer(result.TimerRequest); err != nil {
			logger.Error("Failed to create timer",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", result.TimerRequest.ElementID),
				logger.String("error", err.Error()))
			return fmt.Errorf("failed to create timer: %w", err)
		}
	}

	// Handle waiting state
	if result.WaitingFor != "" {
		token.SetWaitingFor(result.WaitingFor)
		return ep.storage.UpdateToken(token)
	}

	// Handle completion
	if result.Completed {
		token.SetState(models.TokenStateCompleted)
		if err := ep.storage.UpdateToken(token); err != nil {
			return fmt.Errorf("failed to update completed token: %w", err)
		}

		// Cancel boundary timers for completed token
		// Отменяем boundary таймеры для завершенного токена
		if err := ep.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
			logger.Error("Failed to cancel boundary timers for completed token",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
			// Continue execution - boundary timer cancellation is not critical
		}

		// Check if process instance should be completed
		return ep.checkProcessCompletion(token.ProcessInstanceID)
	}

	// Handle token movement to next elements
	if len(result.NextElements) > 0 {
		return ep.moveTokenToNextElements(token, result.NextElements, bpmnProcess)
	}

	// Handle new tokens creation (for parallel flows)
	if len(result.NewTokens) > 0 {
		for _, newToken := range result.NewTokens {
			if err := ep.storage.SaveToken(newToken); err != nil {
				logger.Error("Failed to save new token", logger.String("error", err.Error()))
				continue
			}

			// Execute new token asynchronously
			go func(t *models.Token) {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Panic in token execution goroutine",
							logger.String("token_id", t.TokenID),
							logger.Any("panic", r))
					}
				}()
				if err := ep.component.ExecuteToken(t); err != nil {
					logger.Error("Failed to execute new token", logger.String("error", err.Error()))
				}
			}(newToken)
		}
	}

	// Update original token if needed
	if result.TokenUpdated {
		return ep.storage.UpdateToken(token)
	}

	return nil
}

// moveTokenToNextElements moves token to next elements in the process
// Перемещает токен к следующим элементам в процессе
func (ep *ExecutionProcessor) moveTokenToNextElements(
	token *models.Token,
	nextElements []string,
	bpmnProcess *models.BPMNProcess,
) error {
	if len(nextElements) == 0 {
		return nil
	}

	// Cancel boundary timers if token is leaving an activity
	// Отменяем boundary таймеры если токен покидает activity
	// Boundary timers are bound to specific activity and must be cancelled when token leaves that activity
	// Boundary таймеры привязаны к конкретной activity и должны отменяться когда токен покидает эту activity
	if ep.isActivityElement(token.CurrentElementID, bpmnProcess) {
		logger.Info("Token leaving activity - canceling boundary timers",
			logger.String("token_id", token.TokenID),
			logger.String("current_element_id", token.CurrentElementID))

		if err := ep.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
			logger.Error("Failed to cancel boundary timers when leaving activity",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", token.CurrentElementID),
				logger.String("error", err.Error()))
			// Continue execution - boundary timer cancellation is not critical
		} else {
			logger.Info("Boundary timers canceled when leaving activity",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", token.CurrentElementID))
		}
	}

	// Find target elements by flow IDs
	var targetElements []string
	for _, flowID := range nextElements {
		targetElementID := ep.findTargetElementByFlowID(flowID, bpmnProcess)
		if targetElementID != "" {
			targetElements = append(targetElements, targetElementID)
		} else {
			logger.Error("Target element not found for flow",
				logger.String("flow_id", flowID),
				logger.String("token_id", token.TokenID))
		}
	}

	if len(targetElements) == 0 {
		return fmt.Errorf("no target elements found for flows: %v", nextElements)
	}

	if len(targetElements) == 1 {
		// Simple case: move token to single target element
		token.MoveTo(targetElements[0])
		if err := ep.storage.UpdateToken(token); err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}

		// Continue execution at target element
		return ep.component.ExecuteToken(token)
	}

	// Multiple target elements: create parallel tokens
	token.SetState(models.TokenStateCompleted) // Original token is completed
	if err := ep.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update original token: %w", err)
	}

	// Create new tokens for each target element
	for _, targetElementID := range targetElements {
		newToken := token.Clone()
		newToken.MoveTo(targetElementID)

		if err := ep.storage.SaveToken(newToken); err != nil {
			logger.Error("Failed to save parallel token", logger.String("error", err.Error()))
			continue
		}

		// Execute new token asynchronously
		go func(t *models.Token) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic in parallel token execution goroutine",
						logger.String("token_id", t.TokenID),
						logger.Any("panic", r))
				}
			}()
			if err := ep.component.ExecuteToken(t); err != nil {
				logger.Error("Failed to execute parallel token", logger.String("error", err.Error()))
			}
		}(newToken)
	}

	return nil
}

// findTargetElementByFlowID finds target element by sequence flow ID
// Находит целевой элемент по ID sequence flow
func (ep *ExecutionProcessor) findTargetElementByFlowID(flowID string, bpmnProcess *models.BPMNProcess) string {
	// Search through all elements to find one with this flow in incoming
	for elementID, element := range bpmnProcess.Elements {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}

		incoming, exists := elementMap["incoming"]
		if !exists {
			continue
		}

		// Check if this element has the flow in incoming array
		if incomingList, ok := incoming.([]interface{}); ok {
			for _, item := range incomingList {
				if incomingFlow, ok := item.(string); ok && incomingFlow == flowID {
					return elementID
				}
			}
		} else if incomingStr, ok := incoming.(string); ok && incomingStr == flowID {
			return elementID
		}
	}

	return ""
}

// isActivityElement checks if element is an activity type that can have boundary timers
// Проверяет является ли элемент типом activity который может иметь boundary таймеры
func (ep *ExecutionProcessor) isActivityElement(elementID string, bpmnProcess *models.BPMNProcess) bool {
	element, exists := bpmnProcess.Elements[elementID]
	if !exists {
		return false
	}

	elementMap, ok := element.(map[string]interface{})
	if !ok {
		return false
	}

	elementType, exists := elementMap["type"]
	if !exists {
		return false
	}

	elementTypeStr, ok := elementType.(string)
	if !ok {
		return false
	}

	// Activity types that can have boundary timers
	// Типы activity которые могут иметь boundary таймеры
	activityTypes := []string{
		"serviceTask",
		"userTask",
		"scriptTask",
		"sendTask",
		"receiveTask",
		"manualTask",
		"businessRuleTask",
		"callActivity",
		"subProcess",
		"task", // Generic task type
	}

	for _, activityType := range activityTypes {
		if elementTypeStr == activityType {
			return true
		}
	}

	return false
}

// checkProcessCompletion checks if process instance should be completed
// Проверяет должен ли экземпляр процесса быть завершен
func (ep *ExecutionProcessor) checkProcessCompletion(instanceID string) error {
	// Load all tokens for process instance
	tokens, err := ep.storage.LoadTokensByProcessInstance(instanceID)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	// Check if all tokens are completed
	allCompleted := true
	for _, token := range tokens {
		if !token.IsCompleted() {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		// Load and update process instance
		instance, err := ep.storage.LoadProcessInstance(instanceID)
		if err != nil {
			return fmt.Errorf("failed to load process instance: %w", err)
		}

		instance.SetState(models.ProcessInstanceStateCompleted)
		if err := ep.storage.UpdateProcessInstance(instance); err != nil {
			return fmt.Errorf("failed to update process instance: %w", err)
		}

		logger.Info("Process instance completed", logger.String("instance_id", instanceID))

		// Check for call activity parent tokens waiting for this process
		if err := ep.handleCallActivityCompletion(instanceID); err != nil {
			logger.Error("Failed to handle call activity completion",
				logger.String("instance_id", instanceID),
				logger.String("error", err.Error()))
			// Don't fail the process completion, log error and continue
		}
	}

	return nil
}

// handleCallActivityCompletion handles completion of child process for call activity
// Обрабатывает завершение дочернего процесса для call activity
func (ep *ExecutionProcessor) handleCallActivityCompletion(childInstanceID string) error {
	// Find all tokens waiting for this child process completion
	waitingFor := fmt.Sprintf("call_activity:%s", childInstanceID)

	// Search through waiting tokens to find ones waiting for this child process
	// Optimized: Load only waiting tokens instead of all tokens
	waitingTokens, err := ep.storage.LoadTokensByState(models.TokenStateWaiting)
	if err != nil {
		return fmt.Errorf("failed to load waiting tokens: %w", err)
	}

	var parentTokens []*models.Token
	for _, token := range waitingTokens {
		if token.WaitingFor == waitingFor {
			parentTokens = append(parentTokens, token)
		}
	}

	logger.Info("Found parent tokens waiting for child process completion",
		logger.String("child_instance_id", childInstanceID),
		logger.String("waiting_for", waitingFor),
		logger.Int("parent_tokens_count", len(parentTokens)))

	// Get child process variables for propagation
	childInstance, err := ep.storage.LoadProcessInstance(childInstanceID)
	if err != nil {
		logger.Warn("Failed to load child process instance for variable propagation",
			logger.String("child_instance_id", childInstanceID),
			logger.String("error", err.Error()))
		// Continue without variable propagation
		childInstance = nil
	}

	// Continue execution for each parent token
	for _, parentToken := range parentTokens {
		logger.Info("Continuing call activity parent token execution",
			logger.String("parent_token_id", parentToken.TokenID),
			logger.String("child_instance_id", childInstanceID))

		// Merge child process variables if available
		if childInstance != nil && childInstance.Variables != nil {
			parentToken.MergeVariables(childInstance.Variables)
			logger.Debug("Merged child process variables to parent token",
				logger.String("parent_token_id", parentToken.TokenID),
				logger.Int("variables_count", len(childInstance.Variables)))
		}

		// Clear waiting state
		parentToken.ClearWaitingFor()

		// Update token in storage
		if err := ep.storage.UpdateToken(parentToken); err != nil {
			logger.Error("Failed to update parent token",
				logger.String("parent_token_id", parentToken.TokenID),
				logger.String("error", err.Error()))
			continue
		}

		// Continue token execution
		if err := ep.component.ExecuteToken(parentToken); err != nil {
			logger.Error("Failed to execute parent token",
				logger.String("parent_token_id", parentToken.TokenID),
				logger.String("error", err.Error()))
			// Continue with other tokens
		}
	}

	return nil
}
