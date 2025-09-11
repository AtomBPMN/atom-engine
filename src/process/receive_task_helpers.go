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
)

// extractErrorInfo extracts error code and name from error event definition
// Извлекает код ошибки и имя из определения события ошибки
func (rte *ReceiveTaskExecutor) extractErrorInfo(eventDef map[string]interface{}, bpmnProcess interface{}) (string, string) {
	// Get error reference from event definition
	errorRef, exists := eventDef["reference"] // Changed from "error_ref" to "reference"
	if !exists {
		return "GENERAL_ERROR", "General Error"
	}

	errorRefStr, ok := errorRef.(string)
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	// Get the complete BPMN structure with all elements
	bpmnProcessMap, ok := bpmnProcess.(map[string]interface{})
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	// Look for the error definition in the elements map (not error_definitions array)
	if elements, exists := bpmnProcessMap["elements"]; exists {
		if elementsMap, ok := elements.(map[string]interface{}); ok {
			// Look for the specific error element by ID
			if errorElement, exists := elementsMap[errorRefStr]; exists {
				if errorDefMap, ok := errorElement.(map[string]interface{}); ok {
					errorCode := "GENERAL_ERROR"
					errorName := "General Error"

					// Extract error_code from the error element
					if code, exists := errorDefMap["error_code"]; exists {
						if codeStr, ok := code.(string); ok {
							errorCode = codeStr
						}
					}

					// Extract name from the error element
					if name, exists := errorDefMap["name"]; exists {
						if nameStr, ok := name.(string); ok {
							errorName = nameStr
						}
					}

					logger.Info("Resolved error definition from elements for receive task",
						logger.String("error_ref", errorRefStr),
						logger.String("error_code", errorCode),
						logger.String("error_name", errorName))

					return errorCode, errorName
				}
			}
		}
	}

	logger.Warn("Could not resolve error definition for receive task, using default",
		logger.String("error_ref", errorRefStr))
	return "GENERAL_ERROR", "General Error"
}

// getOutgoingFlows extracts outgoing sequence flows from boundary event
// Извлекает исходящие потоки последовательности из граничного события
func (rte *ReceiveTaskExecutor) getOutgoingFlows(boundaryEvent map[string]interface{}) []string {
	outgoing, exists := boundaryEvent["outgoing"]
	if !exists {
		return []string{}
	}

	var flows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				flows = append(flows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		flows = append(flows, outgoingStr)
	}

	return flows
}

// evaluateTimerExpression evaluates timer expressions using expression component
// Вычисляет timer expressions используя expression компонент
func (rte *ReceiveTaskExecutor) evaluateTimerExpression(expression string, token *models.Token) (interface{}, error) {
	// If not a FEEL expression (doesn't start with =), return as is
	// Если не FEEL expression (не начинается с =), возвращаем как есть
	if expression == "" || len(expression) == 0 || expression[0] != '=' {
		return expression, nil
	}

	// Get expression component through process component
	// Получаем expression компонент через process компонент
	if rte.processComponent == nil {
		return nil, fmt.Errorf("process component not available for expression evaluation")
	}

	// Get core interface
	core := rte.processComponent.GetCore()
	if core == nil {
		return nil, fmt.Errorf("core interface not available for expression evaluation")
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		return nil, fmt.Errorf("expression component not available")
	}

	// Cast to expression evaluator interface with EvaluateExpressionEngine method
	// Приводим к интерфейсу expression evaluator с методом EvaluateExpressionEngine
	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		return nil, fmt.Errorf("failed to cast expression component to ExpressionEvaluator interface")
	}

	// Evaluate FEEL expression using expression engine
	// Вычисляем FEEL expression используя expression engine
	result, err := expressionComp.EvaluateExpressionEngine(expression, token.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate FEEL expression '%s': %w", expression, err)
	}

	logger.Debug("Boundary timer expression evaluated successfully for receive task",
		logger.String("token_id", token.TokenID),
		logger.String("original_expression", expression),
		logger.Any("evaluated_result", result),
		logger.Any("token_variables", token.Variables))

	return result, nil
}
