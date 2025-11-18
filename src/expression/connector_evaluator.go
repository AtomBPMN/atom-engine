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
	"regexp"
	"strings"

	"atom-engine/src/core/logger"
)

// ConnectorExpressionEvaluator processes FEEL expressions in connector parameters
// Обрабатывает FEEL expressions в параметрах коннекторов
type ConnectorExpressionEvaluator struct {
	logger logger.ComponentLogger
	// Regular expressions for finding different types of expressions
	// Регулярные выражения для поиска различных видов выражений
	feelVariableRegex   *regexp.Regexp // for finding variables like "application_id"
	feelExpressionRegex *regexp.Regexp // for finding expressions like "=variable"
}

// NewConnectorExpressionEvaluator creates new connector expression evaluator
// Создает новый обработчик выражений для коннекторов
func NewConnectorExpressionEvaluator(logger logger.ComponentLogger) *ConnectorExpressionEvaluator {
	return &ConnectorExpressionEvaluator{
		logger: logger,
		// Regular expression for finding variables in strings
		// Search for words that can be variables
		// Регулярное выражение для поиска переменных в строках
		// Ищет слова, которые могут быть переменными
		feelVariableRegex: regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`),
		// Regular expression for FEEL expressions with "=" prefix
		// Регулярное выражение для FEEL выражений с префиксом "="
		feelExpressionRegex: regexp.MustCompile(`^=(.+)$`),
	}
}

// EvaluateStringParameter processes string parameter, performing variable substitution
// Обрабатывает строковый параметр, выполняя подстановку переменных
func (cee *ConnectorExpressionEvaluator) EvaluateStringParameter(
	paramName, value string,
	variables map[string]interface{},
) interface{} {
	cee.logger.Debug("Evaluating parameter",
		logger.String("param", paramName),
		logger.String("value", value))

	// Check if value starts with "=" (FEEL expression)
	// Проверяем, начинается ли значение с "=" (FEEL expression)
	if matches := cee.feelExpressionRegex.FindStringSubmatch(value); len(matches) > 1 {
		return cee.evaluateFeelExpression(paramName, matches[1], variables)
	}

	// Check if string contains variables (JSON-like syntax)
	// Проверяем, содержит ли строка переменные (JSON-подобный синтаксис)
	if cee.containsVariables(value) {
		return cee.evaluateJSONString(paramName, value, variables)
	}

	// Return as is
	// Возвращаем как есть
	return value
}

// EvaluateConnectorParameters processes all connector parameters
// Обрабатывает все параметры коннектора
func (cee *ConnectorExpressionEvaluator) EvaluateConnectorParameters(
	connectorParams, processVariables map[string]interface{},
) map[string]interface{} {
	cee.logger.Debug("Evaluating connector parameters",
		logger.Int("variables_count", len(processVariables)))

	result := make(map[string]interface{})

	for paramName, paramValue := range connectorParams {
		switch typedValue := paramValue.(type) {
		case string:
			// Process string parameters
			// Обрабатываем строковые параметры
			evaluated := cee.EvaluateStringParameter(paramName, typedValue, processVariables)
			result[paramName] = evaluated
			cee.logger.Debug("Parameter processed",
				logger.String("param", paramName),
				logger.String("original", typedValue),
				logger.Any("evaluated", evaluated))

		default:
			// For non-string parameters just copy
			// Для не-строковых параметров просто копируем
			result[paramName] = paramValue
		}
	}

	return result
}

// evaluateFeelExpression processes FEEL expression (starting with "=")
// Обрабатывает FEEL expression (начинающееся с "=")
func (cee *ConnectorExpressionEvaluator) evaluateFeelExpression(
	paramName, expression string,
	variables map[string]interface{},
) interface{} {
	cee.logger.Debug("Evaluating FEEL expression",
		logger.String("expression", fmt.Sprintf("=%s", expression)))

	// Simple processing: take variable name directly
	// Простая обработка: берем имя переменной напрямую
	varName := strings.TrimSpace(expression)

	if value, exists := variables[varName]; exists {
		cee.logger.Debug("FEEL variable resolved",
			logger.String("variable", varName),
			logger.Any("value", value))
		return value
	}

	cee.logger.Warn("FEEL variable not found",
		logger.String("variable", varName))
	return expression
}

// evaluateJSONString processes JSON-like strings with variables
// Обрабатывает JSON-подобные строки с переменными
func (cee *ConnectorExpressionEvaluator) evaluateJSONString(
	paramName, jsonStr string,
	variables map[string]interface{},
) interface{} {
	cee.logger.Debug("Evaluating JSON string",
		logger.String("json", jsonStr))

	result := jsonStr
	variablesReplaced := false

	// Simple approach: find variables that are NOT in quotes and replace them
	// Простой подход: ищем переменные, которые НЕ находятся в кавычках и заменяем их
	for varName, varValue := range variables {
		// Find cases where varName stands as value without quotes
		// Pattern: ": varName" but not ": "varName""
		// Ищем варианты где varName стоит как значение без кавычек
		// Паттерн: ": varName" но не ": "varName""
		bareVariablePattern := fmt.Sprintf(`:\s*%s\b`, regexp.QuoteMeta(varName))
		bareVariableRegex := regexp.MustCompile(bareVariablePattern)

		// Check that this is not a string value in quotes
		// Проверяем что это не строковое значение в кавычках
		quotedPattern := fmt.Sprintf(`:\s*["\']%s["\']`, regexp.QuoteMeta(varName))
		quotedRegex := regexp.MustCompile(quotedPattern)

		if bareVariableRegex.MatchString(result) && !quotedRegex.MatchString(result) {
			// Convert variable value to JSON-compatible string
			// Преобразуем значение переменной в JSON-совместимую строку
			var replacement string
			switch v := varValue.(type) {
			case string:
				replacement = fmt.Sprintf("\"%s\"", v)
			case int, int64, float64, bool:
				replacement = fmt.Sprintf("%v", v)
			default:
				replacement = fmt.Sprintf("\"%v\"", v)
			}

			// Replace variable with its value
			// Заменяем переменную на её значение
			replacePattern := fmt.Sprintf(`(:\s*)%s\b`, regexp.QuoteMeta(varName))
			replaceRegex := regexp.MustCompile(replacePattern)
			result = replaceRegex.ReplaceAllString(result, "${1}"+replacement)
			variablesReplaced = true

			cee.logger.Debug("Variable replaced",
				logger.String("variable", varName),
				logger.String("replacement", replacement))
		}
	}

	// Try to parse result as JSON
	// Пытаемся распарсить результат как JSON
	if variablesReplaced || cee.looksLikeValidJSON(result) {
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(result), &jsonObj); err == nil {
			cee.logger.Debug("JSON parsed successfully",
				logger.String("result", result))
			return jsonObj
		}
		cee.logger.Debug("Failed to parse as JSON, returning as string",
			logger.String("result", result))
	}

	return result
}

// containsVariables checks if string contains variables for replacement
// Проверяет, содержит ли строка переменные для замены
func (cee *ConnectorExpressionEvaluator) containsVariables(str string) bool {
	// Check for JSON-like syntax
	// Проверяем на JSON-подобный синтаксис
	if !strings.Contains(str, "{") || !strings.Contains(str, "}") {
		return false
	}

	// Find pattern ": variable_name" (variable as value without quotes)
	// Ищем паттерн ": variable_name" (переменная как значение без кавычек)
	bareValuePattern := regexp.MustCompile(`:\s*[a-zA-Z_][a-zA-Z0-9_]*\b`)

	// Check if there are variables as values
	// Проверяем есть ли переменные как значения
	if !bareValuePattern.MatchString(str) {
		return false
	}

	// Check that these are NOT only string values in quotes
	// Проверяем что это НЕ только строковые значения в кавычках
	quotedValuePattern := regexp.MustCompile(`:\s*["'][^"']*["']`)
	allMatches := bareValuePattern.FindAllString(str, -1)
	quotedMatches := quotedValuePattern.FindAllString(str, -1)

	// If there are bareValue matches, but NOT all of them are in quotes - means there are variables
	// Если есть bareValue совпадения, но НЕ все из них в кавычках - значит есть переменные
	return len(allMatches) > len(quotedMatches)
}

// looksLikeValidJSON checks if string looks like valid JSON
// Проверяет, похожа ли строка на валидный JSON
func (cee *ConnectorExpressionEvaluator) looksLikeValidJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}
