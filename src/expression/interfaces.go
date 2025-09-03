/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

// ExpressionEvaluatorInterface interface for expression evaluation
// Интерфейс для оценки выражений
type ExpressionEvaluatorInterface interface {
	EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	ParseRetries(retriesStr string) (int, error)
}

// VariableEvaluatorInterface interface for variable processing
// Интерфейс для работы с переменными
type VariableEvaluatorInterface interface {
	EvaluateVariable(expression string, variables map[string]interface{}) (interface{}, error)
}

// ConditionEvaluatorInterface interface for condition processing
// Интерфейс для работы с условиями
type ConditionEvaluatorInterface interface {
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateFeelExpression(expression string, variables map[string]interface{}) (bool, error)
}

// RetriesParserInterface interface for retries parsing
// Интерфейс для парсинга повторов
type RetriesParserInterface interface {
	ParseRetries(retriesStr string) (int, error)
}

// ConnectorExpressionEvaluatorInterface interface for FEEL expressions in connectors
// Интерфейс для обработки FEEL expressions в коннекторах
type ConnectorExpressionEvaluatorInterface interface {
	EvaluateStringParameter(paramName, value string, variables map[string]interface{}) interface{}
	EvaluateConnectorParameters(connectorParams, processVariables map[string]interface{}) map[string]interface{}
}

// EngineEvaluatorInterface interface for full expression engine
// Интерфейс для полноценного движка выражений
type EngineEvaluatorInterface interface {
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
}

// EvaluationHelperInterface interface for evaluation helper
// Интерфейс для хелпера оценки выражений
type EvaluationHelperInterface interface {
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	ParseRetries(retriesStr string) (int, error)
}
