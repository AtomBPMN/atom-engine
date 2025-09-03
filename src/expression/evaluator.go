/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"atom-engine/src/core/logger"
)

// ExpressionEvaluator main expression evaluator
// Главный оценщик выражений
type ExpressionEvaluator struct {
	logger             logger.ComponentLogger
	variableEvaluator  *VariableEvaluator
	conditionEvaluator *ConditionEvaluator
	retriesParser      *RetriesParser
	engineEvaluator    *EngineEvaluator
	connectorEvaluator *ConnectorExpressionEvaluator
}

// NewExpressionEvaluator creates new expression evaluator
// Создает новый оценщик выражений
func NewExpressionEvaluator() *ExpressionEvaluator {
	logger := logger.NewComponentLogger("expression-evaluator")

	return &ExpressionEvaluator{
		logger:             logger,
		variableEvaluator:  NewVariableEvaluator(logger),
		conditionEvaluator: NewConditionEvaluator(logger),
		retriesParser:      NewRetriesParser(logger),
		engineEvaluator:    NewEngineEvaluator(logger),
		connectorEvaluator: NewConnectorExpressionEvaluator(logger),
	}
}

// EvaluateExpression evaluates expression in parameters
// Вычисляет выражение в параметрах
func (ee *ExpressionEvaluator) EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error) {
	return ee.variableEvaluator.EvaluateVariable(expression, variables)
}

// EvaluateCondition evaluates conditional expression
// Вычисляет условное выражение
func (ee *ExpressionEvaluator) EvaluateCondition(variables map[string]interface{}, condition string) (bool, error) {
	return ee.conditionEvaluator.EvaluateCondition(variables, condition)
}

// EvaluateExpressionEngine full expression engine
// Полноценный движок выражений
func (ee *ExpressionEvaluator) EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error) {
	return ee.engineEvaluator.EvaluateExpressionEngine(expression, variables)
}

// ParseRetries parses retries count from string
// Парсит количество повторов из строки
func (ee *ExpressionEvaluator) ParseRetries(retriesStr string) (int, error) {
	return ee.retriesParser.ParseRetries(retriesStr)
}

// GetConnectorEvaluator returns connector expression evaluator
// Возвращает обработчик выражений для коннекторов
func (ee *ExpressionEvaluator) GetConnectorEvaluator() *ConnectorExpressionEvaluator {
	return ee.connectorEvaluator
}

// GetVariableEvaluator returns variable evaluator
// Возвращает оценщик переменных
func (ee *ExpressionEvaluator) GetVariableEvaluator() *VariableEvaluator {
	return ee.variableEvaluator
}

// GetConditionEvaluator returns condition evaluator
// Возвращает оценщик условий
func (ee *ExpressionEvaluator) GetConditionEvaluator() *ConditionEvaluator {
	return ee.conditionEvaluator
}

// GetEngineEvaluator returns engine evaluator
// Возвращает движок выражений
func (ee *ExpressionEvaluator) GetEngineEvaluator() *EngineEvaluator {
	return ee.engineEvaluator
}

// GetRetriesParser returns retries parser
// Возвращает парсер повторов
func (ee *ExpressionEvaluator) GetRetriesParser() *RetriesParser {
	return ee.retriesParser
}
