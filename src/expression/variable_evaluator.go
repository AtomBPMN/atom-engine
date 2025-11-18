/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"strings"

	"atom-engine/src/core/logger"
)

// VariableEvaluator variable processor
// Обработчик переменных
type VariableEvaluator struct {
	logger logger.ComponentLogger
}

// NewVariableEvaluator creates new variable processor
// Создает новый обработчик переменных
func NewVariableEvaluator(logger logger.ComponentLogger) *VariableEvaluator {
	return &VariableEvaluator{
		logger: logger,
	}
}

// EvaluateVariable evaluates variable from expression
// Вычисляет переменную из выражения
func (ve *VariableEvaluator) EvaluateVariable(
	expression string,
	variables map[string]interface{},
) (interface{}, error) {
	// Handle variables in format ${variableName}
	// Обрабатываем переменные в формате ${variableName}
	if strings.HasPrefix(expression, "${") && strings.HasSuffix(expression, "}") {
		varName := strings.TrimSuffix(strings.TrimPrefix(expression, "${"), "}")
		if value, exists := variables[varName]; exists {
			ve.logger.Debug("Variable found",
				logger.String("variable", varName),
				logger.Any("value", value))
			return value, nil
		}
		ve.logger.Warn("Variable not found",
			logger.String("variable", varName))
		return expression, nil
	}

	// Handle variables in format #{expression} (Camunda style)
	// Обрабатываем переменные в формате #{expression} (стиль Camunda)
	if strings.HasPrefix(expression, "#{") && strings.HasSuffix(expression, "}") {
		varName := strings.TrimSuffix(strings.TrimPrefix(expression, "#{"), "}")
		if value, exists := variables[varName]; exists {
			ve.logger.Debug("Camunda variable found",
				logger.String("variable", varName),
				logger.Any("value", value))
			return value, nil
		}
		ve.logger.Warn("Camunda variable not found",
			logger.String("variable", varName))
		return expression, nil
	}

	// Handle FEEL expressions starting with "="
	// Обрабатываем FEEL выражения начинающиеся с "="
	if strings.HasPrefix(expression, "=") {
		feelExpr := expression[1:] // Remove "="
		// Handle simple variable access in FEEL
		// Обрабатываем простой доступ к переменным в FEEL
		if value, exists := variables[feelExpr]; exists {
			ve.logger.Debug("FEEL variable found",
				logger.String("variable", feelExpr),
				logger.Any("value", value))
			return value, nil
		}
		ve.logger.Debug("FEEL expression as literal",
			logger.String("expression", feelExpr))
		return feelExpr, nil
	}

	// Handle simple variable name without brackets
	// Обрабатываем простое имя переменной без скобок
	if ve.isSimpleVariableName(expression) {
		if value, exists := variables[expression]; exists {
			ve.logger.Debug("Simple variable found",
				logger.String("variable", expression),
				logger.Any("value", value))
			return value, nil
		}
		ve.logger.Debug("Simple variable not found, returning as literal",
			logger.String("expression", expression))
	}

	ve.logger.Debug("Expression returned as literal",
		logger.String("expression", expression))
	return expression, nil
}

// isSimpleVariableName checks if string is a simple variable name
// Проверяет является ли строка простым именем переменной
func (ve *VariableEvaluator) isSimpleVariableName(str string) bool {
	// Simple validation: variable name should contain only letters, numbers, underscores
	// and start with letter or underscore
	// Простая валидация: имя переменной должно содержать только буквы, цифры, подчеркивания
	// и начинаться с буквы или подчеркивания
	if len(str) == 0 {
		return false
	}

	// Must start with letter or underscore
	// Должно начинаться с буквы или подчеркивания
	first := str[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest can be letters, numbers, underscores
	// Остальное может быть буквами, цифрами, подчеркиваниями
	for i := 1; i < len(str); i++ {
		char := str[i]
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}
