/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"encoding/json"
	"fmt"
	"strings"

	"atom-engine/src/core/logger"
)

// EngineEvaluator full expression engine
// Полноценный движок выражений
type EngineEvaluator struct {
	logger            logger.ComponentLogger
	variableEvaluator *VariableEvaluator
	functionEvaluator *FunctionEvaluator
}

// NewEngineEvaluator creates new expression engine
// Создает новый движок выражений
func NewEngineEvaluator(logger logger.ComponentLogger) *EngineEvaluator {
	return &EngineEvaluator{
		logger:            logger,
		variableEvaluator: NewVariableEvaluator(logger),
		functionEvaluator: NewFunctionEvaluator(logger),
	}
}

// NewEngineEvaluatorWithVariableEvaluator creates new expression engine with shared VariableEvaluator
// Создает новый движок выражений с общим VariableEvaluator
func NewEngineEvaluatorWithVariableEvaluator(
	logger logger.ComponentLogger,
	variableEvaluator *VariableEvaluator,
) *EngineEvaluator {
	return &EngineEvaluator{
		logger:            logger,
		variableEvaluator: variableEvaluator,
		functionEvaluator: NewFunctionEvaluator(logger),
	}
}

// NewEngineEvaluatorWithEvaluators creates new expression engine with shared evaluators
// Создает новый движок выражений с общими evaluators
func NewEngineEvaluatorWithEvaluators(
	logger logger.ComponentLogger,
	variableEvaluator *VariableEvaluator,
	functionEvaluator *FunctionEvaluator,
) *EngineEvaluator {
	return &EngineEvaluator{
		logger:            logger,
		variableEvaluator: variableEvaluator,
		functionEvaluator: functionEvaluator,
	}
}

// EvaluateExpressionEngine full expression engine
// Полноценный движок выражений
func (ee *EngineEvaluator) EvaluateExpressionEngine(
	expression interface{},
	variables map[string]interface{},
) (interface{}, error) {
	switch expr := expression.(type) {
	case string:
		// Check if it's a FEEL function call
		// Проверяем является ли это вызовом FEEL функции
		if ee.functionEvaluator != nil && ee.functionEvaluator.IsFunctionCall(expr) {
			result, err := ee.functionEvaluator.EvaluateFunction(expr, variables)
			if err == nil {
				ee.logger.Debug("Function evaluated successfully",
					logger.Any("result", result))
				return result, nil
			}
			// Fallback to variable evaluation on error
			// Возвращаемся к оценке переменных при ошибке
			ee.logger.Debug("Function evaluation failed, trying variable evaluation",
				logger.String("error", err.Error()))
		}

		// Use VariableEvaluator for all string expression processing
		// Используем VariableEvaluator для всей обработки строковых выражений
		result, err := ee.variableEvaluator.EvaluateVariable(expr, variables)
		if err != nil {
			return expr, err
		}

		// If result is still string and looks like JSON, try to parse it
		// Если результат все еще строка и похож на JSON, пытаемся распарсить
		if strResult, ok := result.(string); ok && strResult != expr {
			if (strings.HasPrefix(strResult, "{") && strings.HasSuffix(strResult, "}")) ||
				(strings.HasPrefix(strResult, "[") && strings.HasSuffix(strResult, "]")) {
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(strResult), &jsonValue); err == nil {
					ee.logger.Debug("Engine parsed JSON from variable",
						logger.Any("parsed_value", jsonValue))
					return jsonValue, nil
				}
			}
		}

		ee.logger.Debug("Engine string processed",
			logger.Any("result", result))
		return result, nil

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
