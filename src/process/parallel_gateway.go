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

// ParallelGatewayExecutor executes parallel gateways
// Исполнитель параллельных шлюзов
type ParallelGatewayExecutor struct{}

// Execute executes parallel gateway
// Выполняет параллельный шлюз
func (pge *ParallelGatewayExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing parallel gateway",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get gateway name for logging
	gatewayName, _ := element["name"].(string)
	if gatewayName == "" {
		gatewayName = token.CurrentElementID
	}

	// Check if this is a fork or join
	incoming, hasIncoming := element["incoming"]
	outgoing, hasOutgoing := element["outgoing"]

	// Get incoming flows count
	var incomingCount int
	if hasIncoming {
		if incomingList, ok := incoming.([]interface{}); ok {
			incomingCount = len(incomingList)
		} else if _, ok := incoming.(string); ok {
			incomingCount = 1
		}
	}

	// Get outgoing flows
	var outgoingFlows []string
	if hasOutgoing {
		if outgoingList, ok := outgoing.([]interface{}); ok {
			for _, item := range outgoingList {
				if flowID, ok := item.(string); ok {
					outgoingFlows = append(outgoingFlows, flowID)
				}
			}
		} else if outgoingStr, ok := outgoing.(string); ok {
			outgoingFlows = append(outgoingFlows, outgoingStr)
		}
	}

	if incomingCount > 1 && len(outgoingFlows) == 1 {
		// This is a join gateway - wait for all incoming tokens
		logger.Info("Parallel gateway join detected",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_name", gatewayName),
			logger.Int("incoming_count", incomingCount))

		// TODO: Implement proper token synchronization
		// For now, just pass through
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: false,
			NextElements: outgoingFlows,
			Completed:    false,
		}, nil

	} else if incomingCount == 1 && len(outgoingFlows) > 1 {
		// This is a fork gateway - create parallel tokens
		logger.Info("Parallel gateway fork detected",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_name", gatewayName),
			logger.Int("outgoing_count", len(outgoingFlows)))

		return &ExecutionResult{
			Success:      true,
			TokenUpdated: false,
			NextElements: outgoingFlows, // Engine will create parallel tokens
			Completed:    false,
		}, nil

	} else {
		// Simple pass-through
		logger.Info("Parallel gateway pass-through",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_name", gatewayName))

		return &ExecutionResult{
			Success:      true,
			TokenUpdated: false,
			NextElements: outgoingFlows,
			Completed:    false,
		}, nil
	}
}

// GetElementType returns element type
// Возвращает тип элемента
func (pge *ParallelGatewayExecutor) GetElementType() string {
	return "parallelGateway"
}
