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

// Helper functions to reduce code duplication in BPMN element executors
// Вспомогательные функции для уменьшения дублирования кода в BPMN element executors

// extractElementName extracts element name from BPMN element for logging
// Извлекает имя элемента из BPMN элемента для логирования
func extractElementName(element map[string]interface{}, elementID string) string {
	if name, exists := element["name"].(string); exists && name != "" {
		return name
	}
	return elementID
}

// extractOutgoingFlows extracts outgoing sequence flows from BPMN element
// Извлекает исходящие sequence flows из BPMN элемента
func extractOutgoingFlows(element map[string]interface{}) []string {
	outgoing, exists := element["outgoing"]
	if !exists {
		return []string{} // No outgoing flows
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

	return nextElements
}

// createSuccessResult creates a standard success execution result
// Создает стандартный успешный результат выполнения
func createSuccessResult(nextElements []string, completed bool) *ExecutionResult {
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    completed,
	}
}

// createCompletionResult creates result for elements with no outgoing flows
// Создает результат для элементов без исходящих flows
func createCompletionResult() *ExecutionResult {
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		Completed:    true,
	}
}

// createWaitingResult creates result for elements that need to wait
// Создает результат для элементов которые должны ждать
func createWaitingResult(waitingFor string) *ExecutionResult {
	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		WaitingFor:   waitingFor,
		Completed:    false,
	}
}

// executeBasicFlowElement executes basic flow element (tasks without special logic)
// Выполняет базовый flow элемент (задачи без специальной логики)
func executeBasicFlowElement(
	token *models.Token,
	element map[string]interface{},
	elementType string,
) (*ExecutionResult, error) {
	elementName := extractElementName(element, token.CurrentElementID)

	logger.Info("Executing "+elementType,
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.String("element_name", elementName))

	nextElements := extractOutgoingFlows(element)

	if len(nextElements) == 0 {
		logger.Info(elementType+" completed - no outgoing flows",
			logger.String("token_id", token.TokenID),
			logger.String("element_name", elementName))
		return createCompletionResult(), nil
	}

	logger.Info(elementType+" completed - moving to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_name", elementName),
		logger.Int("next_elements_count", len(nextElements)))

	return createSuccessResult(nextElements, false), nil
}

// logElementExecution logs standard element execution info
// Логирует стандартную информацию о выполнении элемента
func logElementExecution(token *models.Token, element map[string]interface{}, elementType string) {
	elementName := extractElementName(element, token.CurrentElementID)

	logger.Info("Executing "+elementType,
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.String("element_type", elementType),
		logger.String("element_name", elementName))
}
