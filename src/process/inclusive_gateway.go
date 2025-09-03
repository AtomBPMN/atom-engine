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

// InclusiveGatewayExecutor executes inclusive gateways
// Исполнитель включающих шлюзов
type InclusiveGatewayExecutor struct{}

// Execute executes inclusive gateway
// Выполняет включающий шлюз
func (ige *InclusiveGatewayExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing inclusive gateway",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get gateway name for logging
	gatewayName, _ := element["name"].(string)
	if gatewayName == "" {
		gatewayName = token.CurrentElementID
	}

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:   false,
			Error:     "inclusive gateway has no outgoing sequence flows",
			Completed: false,
		}, nil
	}

	var outgoingFlows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				outgoingFlows = append(outgoingFlows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		outgoingFlows = append(outgoingFlows, outgoingStr)
	}

	if len(outgoingFlows) == 0 {
		return &ExecutionResult{
			Success:   false,
			Error:     "inclusive gateway has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// For inclusive gateway, evaluate conditions and select matching flows
	// For now, select all flows (in real implementation would evaluate conditions)
	selectedFlows := outgoingFlows

	logger.Info("Inclusive gateway executed",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.Int("selected_flows", len(selectedFlows)),
		logger.Int("total_flows", len(outgoingFlows)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: selectedFlows,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ige *InclusiveGatewayExecutor) GetElementType() string {
	return "inclusiveGateway"
}
