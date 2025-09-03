/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// TaskExecutor executes tasks
// Исполнитель задач
type TaskExecutor struct{}

// Execute executes task
// Выполняет задачу
func (te *TaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing task",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get task name for logging
	taskName, _ := element["name"].(string)
	if taskName == "" {
		taskName = token.CurrentElementID
	}

	logger.Info("Task executed - moving to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.String("task_name", taskName))

	// For basic task, we just pass through to next elements
	// In real implementation, this would create jobs, call services, etc.
	outgoing, exists := element["outgoing"]
	if !exists {
		// Task with no outgoing flows - complete the token
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

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (te *TaskExecutor) GetElementType() string {
	return "task"
}

// NOTE: ServiceTaskExecutor moved to src/process/flow/elements/task/service_task.go
// Old ServiceTaskExecutor removed - now using new version with Jobs integration

// UserTaskExecutor executes user tasks
// Исполнитель пользовательских задач
type UserTaskExecutor struct{}

// Execute executes user task
// Выполняет пользовательскую задачу
func (ute *UserTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing user task",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get task name for logging
	taskName, _ := element["name"].(string)
	if taskName == "" {
		taskName = token.CurrentElementID
	}

	// User tasks typically wait for external completion
	// For now, we'll put the token in waiting state
	logger.Info("User task waiting for completion",
		logger.String("token_id", token.TokenID),
		logger.String("task_name", taskName))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		WaitingFor:   "user_task_completion",
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ute *UserTaskExecutor) GetElementType() string {
	return "userTask"
}
