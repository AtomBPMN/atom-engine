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

// IntermediateCatchEventExecutor executes intermediate catch events
// Исполнитель промежуточных событий ловли
type IntermediateCatchEventExecutor struct {
	processComponent ComponentInterface
	timerHandler     *IntermediateCatchTimerHandler
	messageHandler   *IntermediateCatchMessageHandler
}

// NewIntermediateCatchEventExecutor creates new intermediate catch event executor
// Создает новый исполнитель промежуточного события ловли
func NewIntermediateCatchEventExecutor(processComponent ComponentInterface) *IntermediateCatchEventExecutor {
	return &IntermediateCatchEventExecutor{
		processComponent: processComponent,
		timerHandler:     NewIntermediateCatchTimerHandler(processComponent),
		messageHandler:   NewIntermediateCatchMessageHandler(processComponent),
	}
}

// Execute executes intermediate catch event
// Выполняет промежуточное событие ловли
func (icee *IntermediateCatchEventExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing intermediate catch event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get event name for logging
	eventName, _ := element["name"].(string)
	if eventName == "" {
		eventName = token.CurrentElementID
	}

	// Check for event definitions to determine event type
	eventDefinitions, hasEventDefs := element["event_definitions"]
	if hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					eventType, _ := eventDefMap["type"].(string)

					// Handle timer events
					if eventType == "timerEventDefinition" {
						return icee.timerHandler.HandleTimerEvent(token, element, eventDefMap)
					}

					// Handle message events
					if eventType == "messageEventDefinition" {
						return icee.messageHandler.HandleMessageEvent(token, element, eventDefMap)
					}

					// Handle signal events
					if eventType == "signalEventDefinition" {
						return icee.handleSignalEvent(token, element, eventDefMap)
					}
				}
			}
		}
	}

	// No specific event definition or unsupported type - immediate trigger
	logger.Info("Intermediate catch event triggered immediately",
		logger.String("token_id", token.TokenID),
		logger.String("event_name", eventName))

	return icee.handleDefaultEvent(token, element)
}

// handleSignalEvent handles signal intermediate catch events
// Обрабатывает signal промежуточные catch события
func (icee *IntermediateCatchEventExecutor) handleSignalEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Signal event detected - waiting for signal",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// TODO: Implement signal subscription
	// For now, proceed immediately as signals are not implemented
	logger.Warn("Signal events not yet implemented - proceeding immediately",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	return icee.handleDefaultEvent(token, element)
}

// handleDefaultEvent handles default intermediate catch event (no specific event definition)
// Обрабатывает default промежуточное catch событие (без specific event definition)
func (icee *IntermediateCatchEventExecutor) handleDefaultEvent(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	// Get outgoing sequence flows
	var nextElements []string
	if outgoing, exists := element["outgoing"]; exists {
		if outgoingList, ok := outgoing.([]interface{}); ok {
			for _, item := range outgoingList {
				if flowID, ok := item.(string); ok {
					nextElements = append(nextElements, flowID)
				}
			}
		} else if outgoingStr, ok := outgoing.(string); ok {
			nextElements = append(nextElements, outgoingStr)
		}
	}

	if len(nextElements) == 0 {
		// No outgoing flows - complete the token
		logger.Info("Intermediate catch event has no outgoing flows - completing token",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	// Continue to next elements
	logger.Info("Intermediate catch event proceeding to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.Int("next_elements_count", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (icee *IntermediateCatchEventExecutor) GetElementType() string {
	return "intermediateCatchEvent"
}
