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

// CallActivityExecutor executes call activities
// Исполнитель вызываемых активностей
type CallActivityExecutor struct{}

// Execute executes call activity
// Выполняет вызываемую активность
func (cae *CallActivityExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing call activity",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get activity name for logging
	activityName, _ := element["name"].(string)
	if activityName == "" {
		activityName = token.CurrentElementID
	}

	// Call activities invoke external processes or services
	// In real implementation would start child process instance
	logger.Info("Call activity executed",
		logger.String("token_id", token.TokenID),
		logger.String("activity_name", activityName))

	// For now, simulate immediate completion
	// In real implementation would wait for child process completion

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
func (cae *CallActivityExecutor) GetElementType() string {
	return "callActivity"
}
