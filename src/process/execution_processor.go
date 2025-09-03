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
func (ep *ExecutionProcessor) processExecutionResult(token *models.Token, result *ExecutionResult, bpmnProcess *models.BPMNProcess) error {
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
func (ep *ExecutionProcessor) moveTokenToNextElements(token *models.Token, nextElements []string, bpmnProcess *models.BPMNProcess) error {
	if len(nextElements) == 0 {
		return nil
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
	}

	return nil
}
