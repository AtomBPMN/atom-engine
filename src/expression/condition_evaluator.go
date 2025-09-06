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
func NewConditionEvaluatorWithVariableEvaluator(logger logger.ComponentLogger, variableEvaluator *VariableEvaluator) *ConditionEvaluator {
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
func (ce *ConditionEvaluator) EvaluateFeelExpression(expression string, variables map[string]interface{}) (bool, error) {
	ce.logger.Debug("Evaluating FEEL expression",
		logger.String("expression", expression))

	// Handle boolean expressions like "gate=true", "gate=false"
	// Обработка булевых выражений типа "gate=true", "gate=false"
	if strings.Contains(expression, "=") {
		parts := strings.Split(expression, "=")
		if len(parts) == 2 {
			varName := strings.TrimSpace(parts[0])
			expectedValue := strings.TrimSpace(parts[1])

			ce.logger.Debug("FEEL boolean comparison",
				logger.String("variable", varName),
				logger.String("expected", expectedValue))

			// Get variable value
			// Получаем значение переменной
			value, exists := variables[varName]
			if !exists {
				ce.logger.Warn("FEEL variable not found",
					logger.String("variable", varName))
				return false, fmt.Errorf("variable '%s' not found", varName)
			}

			// Convert expected value to needed type
			// Преобразуем ожидаемое значение в нужный тип
			switch strings.ToLower(expectedValue) {
			case "true":
				if boolVal, ok := value.(bool); ok {
					ce.logger.Debug("FEEL boolean result",
						logger.String("variable", varName),
						logger.Bool("value", boolVal),
						logger.Bool("result", boolVal))
					return boolVal, nil
				}
				// Check string representation
				// Проверяем строковое представление
				strVal := fmt.Sprintf("%v", value)
				result := strings.ToLower(strVal) == "true"
				ce.logger.Debug("FEEL string->boolean result",
					logger.String("variable", varName),
					logger.String("value", strVal),
					logger.Bool("result", result))
				return result, nil

			case "false":
				if boolVal, ok := value.(bool); ok {
					ce.logger.Debug("FEEL boolean result",
						logger.String("variable", varName),
						logger.Bool("value", boolVal),
						logger.Bool("result", !boolVal))
					return !boolVal, nil
				}
				// Check string representation
				// Проверяем строковое представление
				strVal := fmt.Sprintf("%v", value)
				result := strings.ToLower(strVal) == "false"
				ce.logger.Debug("FEEL string->boolean result",
					logger.String("variable", varName),
					logger.String("value", strVal),
					logger.Bool("result", result))
				return result, nil

			default:
				// String comparison
				// Сравнение как строки
				actualStr := fmt.Sprintf("%v", value)

				// Remove quotes from string literals if present
				// Удаляем кавычки из строковых литералов если они есть
				cleanExpectedValue := expectedValue
				if len(expectedValue) >= 2 &&
					((expectedValue[0] == '"' && expectedValue[len(expectedValue)-1] == '"') ||
						(expectedValue[0] == '\'' && expectedValue[len(expectedValue)-1] == '\'')) {
					cleanExpectedValue = expectedValue[1 : len(expectedValue)-1]
					ce.logger.Debug("FEEL removed quotes from expected value",
						logger.String("original", expectedValue),
						logger.String("cleaned", cleanExpectedValue))
				}

				result := actualStr == cleanExpectedValue
				ce.logger.Debug("FEEL string comparison result",
					logger.String("variable", varName),
					logger.String("actual", actualStr),
					logger.String("expected", cleanExpectedValue),
					logger.Bool("result", result))
				return result, nil
			}
		}
	}

	// Handle simple boolean variables like "gate"
	// Обработка простых булевых переменных типа "gate"
	if varName := strings.TrimSpace(expression); varName != "" {
		value, exists := variables[varName]
		if !exists {
			ce.logger.Warn("FEEL variable not found",
				logger.String("variable", varName))
			return false, fmt.Errorf("variable '%s' not found", varName)
		}

		// If variable is boolean, return its value
		// Если переменная булевая, возвращаем её значение
		if boolVal, ok := value.(bool); ok {
			ce.logger.Debug("FEEL simple boolean",
				logger.String("variable", varName),
				logger.Bool("value", boolVal))
			return boolVal, nil
		}

		// If variable is string, check for "true"
		// Если переменная строковая, проверяем на "true"
		strVal := fmt.Sprintf("%v", value)
		result := strings.ToLower(strVal) == "true"
		ce.logger.Debug("FEEL string->boolean",
			logger.String("variable", varName),
			logger.String("value", strVal),
			logger.Bool("result", result))
		return result, nil
	}

	ce.logger.Warn("Could not evaluate FEEL expression, returning false",
		logger.String("expression", expression))
	return false, fmt.Errorf("unsupported FEEL expression: %s", expression)
}
