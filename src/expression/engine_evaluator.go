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

// EngineEvaluator full expression engine
// Полноценный движок выражений
type EngineEvaluator struct {
	logger            logger.ComponentLogger
	variableEvaluator *VariableEvaluator
}

// NewEngineEvaluator creates new expression engine
// Создает новый движок выражений
func NewEngineEvaluator(logger logger.ComponentLogger) *EngineEvaluator {
	return &EngineEvaluator{
		logger:            logger,
		variableEvaluator: NewVariableEvaluator(logger),
	}
}

// EvaluateExpressionEngine full expression engine
// Полноценный движок выражений
func (ee *EngineEvaluator) EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error) {
	switch expr := expression.(type) {
	case string:
		// Process string expressions
		// Обработка строковых выражений
		if strings.HasPrefix(expr, "${") && strings.HasSuffix(expr, "}") {
			varName := strings.TrimSuffix(strings.TrimPrefix(expr, "${"), "}")
			if value, exists := variables[varName]; exists {
				ee.logger.Debug("Engine variable found",
					logger.String("variable", varName),
					logger.Any("value", value))
				return value, nil
			}
			ee.logger.Warn("Engine variable not found",
				logger.String("variable", varName))
			return expr, nil
		}

		// Handle Camunda-style expressions #{variable}
		// Обрабатываем выражения в стиле Camunda #{variable}
		if strings.HasPrefix(expr, "#{") && strings.HasSuffix(expr, "}") {
			varName := strings.TrimSuffix(strings.TrimPrefix(expr, "#{"), "}")
			if value, exists := variables[varName]; exists {
				ee.logger.Debug("Engine Camunda variable found",
					logger.String("variable", varName),
					logger.Any("value", value))
				return value, nil
			}
			ee.logger.Warn("Engine Camunda variable not found",
				logger.String("variable", varName))
			return expr, nil
		}

		// Handle FEEL expressions starting with "="
		// Обрабатываем FEEL выражения начинающиеся с "="
		if strings.HasPrefix(expr, "=") {
			feelExpr := expr[1:] // Remove "="
			// For now, handle simple variable access
			// Пока обрабатываем простой доступ к переменным
			if value, exists := variables[feelExpr]; exists {
				ee.logger.Debug("Engine FEEL variable found",
					logger.String("variable", feelExpr),
					logger.Any("value", value))
				return value, nil
			}
			ee.logger.Debug("Engine FEEL expression as literal",
				logger.String("expression", feelExpr))
			return feelExpr, nil
		}

		ee.logger.Debug("Engine string literal",
			logger.String("value", expr))
		return expr, nil

	case int, int32, int64:
		ee.logger.Debug("Engine integer",
			logger.Any("value", expr))
		return expr, nil

	case float32, float64:
		ee.logger.Debug("Engine float",
			logger.Any("value", expr))
		return expr, nil

	case bool:
		ee.logger.Debug("Engine boolean",
			logger.Bool("value", expr))
		return expr, nil

	case map[string]interface{}:
		// Process map expressions - recursively evaluate each value
		// Обрабатываем map выражения - рекурсивно оцениваем каждое значение
		result := make(map[string]interface{})
		for key, value := range expr {
			evaluatedValue, err := ee.EvaluateExpressionEngine(value, variables)
			if err != nil {
				ee.logger.Warn("Failed to evaluate map value",
					logger.String("key", key),
					logger.String("error", err.Error()))
				result[key] = value // Keep original value on error
			} else {
				result[key] = evaluatedValue
			}
		}
		ee.logger.Debug("Engine map processed",
			logger.Int("keys_count", len(result)))
		return result, nil

	case []interface{}:
		// Process array expressions - recursively evaluate each element
		// Обрабатываем массивы - рекурсивно оцениваем каждый элемент
		result := make([]interface{}, len(expr))
		for i, element := range expr {
			evaluatedElement, err := ee.EvaluateExpressionEngine(element, variables)
			if err != nil {
				ee.logger.Warn("Failed to evaluate array element",
					logger.Int("index", i),
					logger.String("error", err.Error()))
				result[i] = element // Keep original element on error
			} else {
				result[i] = evaluatedElement
			}
		}
		ee.logger.Debug("Engine array processed",
			logger.Int("elements_count", len(result)))
		return result, nil

	default:
		ee.logger.Warn("Unsupported expression type",
			logger.String("type", fmt.Sprintf("%T", expr)))
		return expr, nil
	}
}
