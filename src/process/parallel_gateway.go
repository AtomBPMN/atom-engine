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
type ParallelGatewayExecutor struct {
	processComponent ComponentInterface
}

// NewParallelGatewayExecutor creates new parallel gateway executor
// Создает новый исполнитель параллельного шлюза
func NewParallelGatewayExecutor(processComponent ComponentInterface) *ParallelGatewayExecutor {
	return &ParallelGatewayExecutor{
		processComponent: processComponent,
	}
}

// Execute executes parallel gateway
// Выполняет параллельный шлюз
func (pge *ParallelGatewayExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
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
		} else if incomingList, ok := incoming.([]string); ok {
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
		} else if outgoingList, ok := outgoing.([]string); ok {
			outgoingFlows = append(outgoingFlows, outgoingList...)
		} else if outgoingStr, ok := outgoing.(string); ok {
			outgoingFlows = append(outgoingFlows, outgoingStr)
		}
	}

	logger.Info("Gateway flow analysis",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_id", token.CurrentElementID),
		logger.Int("incoming_count", incomingCount),
		logger.Int("outgoing_count", len(outgoingFlows)),
		logger.Bool("has_incoming", hasIncoming),
		logger.Bool("has_outgoing", hasOutgoing))

	if incomingCount > 1 && len(outgoingFlows) >= 1 {
		// This is a join gateway - wait for all incoming tokens
		logger.Info("Parallel gateway join detected",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_name", gatewayName),
			logger.Int("incoming_count", incomingCount))

		return pge.handleJoinGateway(token, token.CurrentElementID, incomingCount, outgoingFlows)

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

// handleJoinGateway handles token synchronization for join gateway
// Обрабатывает синхронизацию токенов для join gateway
func (pge *ParallelGatewayExecutor) handleJoinGateway(
	token *models.Token,
	gatewayID string,
	expectedCount int,
	outgoingFlows []string,
) (*ExecutionResult, error) {
	// Load or create gateway synchronization state
	syncState, err := pge.processComponent.LoadGatewaySyncState(gatewayID, token.ProcessInstanceID)
	if err != nil {
		logger.Error("Failed to load gateway sync state",
			logger.String("gateway_id", gatewayID),
			logger.String("process_instance_id", token.ProcessInstanceID),
			logger.String("error", err.Error()))
		return &ExecutionResult{Success: false}, err
	}

	// Create new sync state if not exists
	if syncState == nil {
		syncState = models.NewGatewaySyncState(gatewayID, token.ProcessInstanceID, expectedCount)
		logger.Info("Created new gateway sync state",
			logger.String("gateway_id", gatewayID),
			logger.String("process_instance_id", token.ProcessInstanceID),
			logger.Int("expected_count", expectedCount))
	}

	// Check if this token already arrived (prevent duplicates)
	if syncState.HasToken(token.TokenID) {
		logger.Warn("Token already processed by gateway",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_id", gatewayID))
		// Complete the token without further processing
		return &ExecutionResult{
			Success:   true,
			Completed: true,
		}, nil
	}

	// Add this token to arrived tokens
	syncState.AddToken(token.TokenID)

	logger.Info("Token arrived at join gateway",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_id", gatewayID),
		logger.Int("arrived_count", len(syncState.ArrivedTokens)),
		logger.Int("expected_count", syncState.ExpectedTokenCount))

	// Save updated sync state
	if err := pge.processComponent.SaveGatewaySyncState(syncState); err != nil {
		logger.Error("Failed to save gateway sync state",
			logger.String("gateway_id", gatewayID),
			logger.String("error", err.Error()))
		return &ExecutionResult{Success: false}, err
	}

	// Complete current token (it will be completed and removed from active tokens)
	token.SetState(models.TokenStateCompleted)

	// Check if all tokens have arrived
	if syncState.IsComplete() {
		logger.Info("All tokens arrived at join gateway - proceeding to next elements",
			logger.String("gateway_id", gatewayID),
			logger.String("process_instance_id", token.ProcessInstanceID),
			logger.Int("total_tokens", len(syncState.ArrivedTokens)))

		// Clean up sync state
		if err := pge.processComponent.DeleteGatewaySyncState(gatewayID, token.ProcessInstanceID); err != nil {
			logger.Error("Failed to delete gateway sync state",
				logger.String("gateway_id", gatewayID),
				logger.String("error", err.Error()))
			// Continue anyway - this is not critical
		}

		// Create new token for next elements
		newToken := token.Clone()
		newToken.SetState(models.TokenStateActive)

		// Return execution result to proceed to next elements
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: false,
			NextElements: outgoingFlows,
			Completed:    false,
			NewTokens:    []*models.Token{newToken},
		}, nil
	} else {
		// Not all tokens arrived yet - wait
		logger.Info("Waiting for more tokens at join gateway",
			logger.String("gateway_id", gatewayID),
			logger.Int("arrived_count", len(syncState.ArrivedTokens)),
			logger.Int("expected_count", syncState.ExpectedTokenCount))

		// Complete current token and wait
		return &ExecutionResult{
			Success:   true,
			Completed: true,
		}, nil
	}
}
