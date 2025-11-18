/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"strconv"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// EvaluationHelper helper for expression evaluation
// Помощник для вычисления выражений и условий
type EvaluationHelper struct {
	expressionEvaluator *ExpressionEvaluator
	logger              logger.ComponentLogger
}

// NewEvaluationHelper creates new evaluation helper
// Создает новый помощник для вычислений
func NewEvaluationHelper(expressionEvaluator *ExpressionEvaluator) *EvaluationHelper {
	return &EvaluationHelper{
		expressionEvaluator: expressionEvaluator,
		logger:              logger.NewComponentLogger("evaluation-helper"),
	}
}

// EvaluateCondition evaluates conditional expression with token variables
// Вычисляет условное выражение с переменными токена
func (eh *EvaluationHelper) EvaluateCondition(token *models.Token, condition string) (bool, error) {
	if token == nil {
		eh.logger.Warn("Token is nil, cannot evaluate condition")
		return false, nil
	}

	variables := token.Variables
	if variables == nil {
		variables = make(map[string]interface{})
	}

	eh.logger.Debug("Evaluating condition",
		logger.String("token_id", token.TokenID),
		logger.String("condition", condition),
		logger.Int("variables_count", len(variables)))

	return eh.expressionEvaluator.EvaluateCondition(variables, condition)
}

// EvaluateConditionWithVariables evaluates conditional expression with provided variables
// Вычисляет условное выражение с предоставленными переменными
func (eh *EvaluationHelper) EvaluateConditionWithVariables(
	variables map[string]interface{},
	condition string,
) (bool, error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	eh.logger.Debug("Evaluating condition with variables",
		logger.String("condition", condition),
		logger.Int("variables_count", len(variables)))

	return eh.expressionEvaluator.EvaluateCondition(variables, condition)
}

// EvaluateExpression evaluates expression in parameters
// Вычисляет выражение в параметрах
func (eh *EvaluationHelper) EvaluateExpression(
	expression string,
	variables map[string]interface{},
) (interface{}, error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	eh.logger.Debug("Evaluating expression",
		logger.String("expression", expression),
		logger.Int("variables_count", len(variables)))

	return eh.expressionEvaluator.EvaluateExpression(expression, variables)
}

// EvaluateExpressionEngine full expression engine
// Полноценный движок выражений
func (eh *EvaluationHelper) EvaluateExpressionEngine(
	expression interface{},
	variables map[string]interface{},
) (interface{}, error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	eh.logger.Debug("Evaluating expression with engine",
		logger.Any("expression", expression),
		logger.Int("variables_count", len(variables)))

	return eh.expressionEvaluator.EvaluateExpressionEngine(expression, variables)
}

// ParseRetries parses retries count from string
// Парсит количество повторов из строки
func (eh *EvaluationHelper) ParseRetries(retriesStr string) (int, error) {
	if retriesStr == "" {
		return 3, nil // Default value
	}

	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		eh.logger.Warn("Failed to parse retries, using default",
			logger.String("retries", retriesStr),
			logger.String("error", err.Error()))
		return 3, err // Return default value on error
	}

	if retries < 0 {
		return 0, nil
	}

	return retries, nil
}

// EvaluateTokenVariables evaluates all variables in token using expression engine
// Оценивает все переменные в токене используя expression engine
func (eh *EvaluationHelper) EvaluateTokenVariables(token *models.Token) error {
	if token == nil || token.Variables == nil {
		return nil
	}

	eh.logger.Debug("Evaluating token variables",
		logger.String("token_id", token.TokenID),
		logger.Int("variables_count", len(token.Variables)))

	evaluatedVariables := make(map[string]interface{})

	for key, value := range token.Variables {
		evaluatedValue, err := eh.expressionEvaluator.EvaluateExpressionEngine(value, token.Variables)
		if err != nil {
			eh.logger.Warn("Failed to evaluate token variable",
				logger.String("token_id", token.TokenID),
				logger.String("variable", key),
				logger.String("error", err.Error()))
			evaluatedVariables[key] = value // Keep original value on error
		} else {
			evaluatedVariables[key] = evaluatedValue
		}
	}

	// Update token variables with evaluated values
	// Обновляем переменные токена оцененными значениями
	token.Variables = evaluatedVariables
	token.UpdatedAt = time.Now()

	eh.logger.Debug("Token variables evaluated successfully",
		logger.String("token_id", token.TokenID))

	return nil
}

// GetConnectorEvaluator returns connector expression evaluator
// Возвращает обработчик выражений для коннекторов
func (eh *EvaluationHelper) GetConnectorEvaluator() *ConnectorExpressionEvaluator {
	return eh.expressionEvaluator.GetConnectorEvaluator()
}
