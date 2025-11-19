/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"fmt"
	"strings"

	"atom-engine/src/core/logger"
)

// ConditionEvaluator condition processor
// Обработчик условий
type ConditionEvaluator struct {
	logger            logger.ComponentLogger
	variableEvaluator *VariableEvaluator
}

// NewConditionEvaluator creates new condition processor
// Создает новый обработчик условий
func NewConditionEvaluator(logger logger.ComponentLogger) *ConditionEvaluator {
	return &ConditionEvaluator{
		logger:            logger,
		variableEvaluator: NewVariableEvaluator(logger),
	}
}

// NewConditionEvaluatorWithVariableEvaluator creates new condition processor with shared VariableEvaluator
// Создает новый обработчик условий с общим VariableEvaluator
func NewConditionEvaluatorWithVariableEvaluator(
	logger logger.ComponentLogger,
	variableEvaluator *VariableEvaluator,
) *ConditionEvaluator {
	return &ConditionEvaluator{
		logger:            logger,
		variableEvaluator: variableEvaluator,
	}
}

// EvaluateCondition evaluates conditional expression
// Вычисляет условное выражение
func (ce *ConditionEvaluator) EvaluateCondition(variables map[string]interface{}, condition string) (bool, error) {
	// FEEL syntax: expressions starting with '=' are processed as FEEL
	// FEEL синтаксис: выражения начинающиеся с '=' обрабатываются как FEEL
	if strings.HasPrefix(condition, "=") {
		return ce.EvaluateFeelExpression(condition[1:], variables) // remove '=' at the beginning
	}

	// Simple condition implementation for backward compatibility
	// Простая реализация условий для обратной совместимости
	// Example: condition = "${status} == 'approved'"
	if strings.Contains(condition, "==") {
		// Check for incorrect operators (triple equality, multiple operators)
		// Проверяем на некорректные операторы (тройное равенство, множественные операторы)
		if strings.Contains(condition, "===") {
			ce.logger.Warn("Unsupported operator '===' in condition",
				logger.String("condition", condition))
			return false, fmt.Errorf("unsupported operator '===' in condition: %s", condition)
		}

		parts := strings.Split(condition, "==")

		// Check for multiple operators (more than 2 parts means multiple '==')
		// Проверяем на множественные операторы (больше 2 частей означает несколько '==')
		if len(parts) > 2 {
			ce.logger.Warn("Multiple '==' operators in condition",
				logger.String("condition", condition))
			return false, fmt.Errorf("multiple '==' operators not supported in condition: %s", condition)
		}

		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Validation: left and right parts should not be empty
			// Валидация: левая и правая части не должны быть пустыми
			if left == "" {
				ce.logger.Warn("Empty left side in condition",
					logger.String("condition", condition))
				return false, fmt.Errorf("empty left side in condition: %s", condition)
			}

			if right == "" {
				ce.logger.Warn("Empty right side in condition",
					logger.String("condition", condition))
				return false, fmt.Errorf("empty right side in condition: %s", condition)
			}

			// Remove quotes from right part
			// Убираем кавычки из правой части
			right = strings.Trim(right, "'\"")

			// Evaluate left part
			// Вычисляем левую часть
			leftValue, err := ce.variableEvaluator.EvaluateVariable(left, variables)
			if err != nil {
				return false, err
			}

			result := fmt.Sprintf("%v", leftValue) == right
			ce.logger.Debug("Simple condition evaluated",
				logger.String("left", fmt.Sprintf("%v", leftValue)),
				logger.String("right", right),
				logger.Bool("result", result))
			return result, nil
		}
	}

	// For unsupported conditions return error
	// Для неподдерживаемых условий возвращаем ошибку
	ce.logger.Warn("Could not evaluate condition",
		logger.String("condition", condition))
	return false, fmt.Errorf("unsupported condition format: %s", condition)
}

// EvaluateFeelExpression processes FEEL expressions
// Обрабатывает FEEL выражения
func (ce *ConditionEvaluator) EvaluateFeelExpression(
	expression string,
	variables map[string]interface{},
) (bool, error) {
	ce.logger.Debug("Evaluating FEEL expression",
		logger.String("expression", expression))

	// Use VariableEvaluator for all FEEL expressions
	// Используем VariableEvaluator для всех FEEL выражений
	result, err := ce.variableEvaluator.EvaluateVariable("="+expression, variables)
	if err != nil {
		ce.logger.Warn("Failed to evaluate FEEL expression",
			logger.String("expression", expression),
			logger.String("error", err.Error()))
		return false, err
	}

	// Convert result to boolean
	// Конвертируем результат в boolean
	if boolVal, ok := result.(bool); ok {
		ce.logger.Debug("FEEL expression result",
			logger.String("expression", expression),
			logger.Bool("result", boolVal))
		return boolVal, nil
	}

	// If result is not boolean, treat as string and check for "true"
	// Если результат не boolean, обрабатываем как строку и проверяем на "true"
	strVal := fmt.Sprintf("%v", result)
	boolResult := strings.ToLower(strVal) == "true"
	ce.logger.Debug("FEEL expression converted to boolean",
		logger.String("expression", expression),
		logger.String("result_str", strVal),
		logger.Bool("result_bool", boolResult))
	return boolResult, nil
}
