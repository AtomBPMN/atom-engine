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

// EventBasedGatewayExecutor executes event-based gateways
// Исполнитель событийных шлюзов
type EventBasedGatewayExecutor struct{}

// Execute executes event-based gateway
// Выполняет событийный шлюз
func (ebge *EventBasedGatewayExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing event-based gateway",
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
			Error:     "event-based gateway has no outgoing sequence flows",
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
			Error:     "event-based gateway has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// Event-based gateway creates waiting tokens for all outgoing events
	// The first event to occur will continue, others will be canceled
	logger.Info("Event-based gateway creating waiting tokens",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.Int("outgoing_flows", len(outgoingFlows)))

	// For now, just pass to all outgoing flows
	// In real implementation, would create competing event subscriptions
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: outgoingFlows,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ebge *EventBasedGatewayExecutor) GetElementType() string {
	return "eventBasedGateway"
}
