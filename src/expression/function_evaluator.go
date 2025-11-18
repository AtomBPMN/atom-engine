/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/timewheel"
)

// FunctionEvaluator evaluates FEEL functions
// Оценщик FEEL функций
type FunctionEvaluator struct {
	logger            logger.ComponentLogger
	durationParser    *timewheel.ISO8601DurationParser
	functionCallRegex *regexp.Regexp
}

// NewFunctionEvaluator creates new function evaluator
// Создает новый оценщик функций
func NewFunctionEvaluator(logger logger.ComponentLogger) *FunctionEvaluator {
	return &FunctionEvaluator{
		logger:            logger,
		durationParser:    timewheel.NewISO8601DurationParser(),
		functionCallRegex: regexp.MustCompile(`^([a-z]+)\((.*)\)$`),
	}
}

// IsFunctionCall checks if expression is a function call
// Проверяет является ли выражение вызовом функции
func (fe *FunctionEvaluator) IsFunctionCall(expr string) bool {
	// Remove leading = if present (FEEL expression prefix)
	// Убираем ведущий = если есть (префикс FEEL выражения)
	trimmed := strings.TrimSpace(expr)
	if strings.HasPrefix(trimmed, "=") {
		trimmed = strings.TrimSpace(trimmed[1:])
	}
	return fe.functionCallRegex.MatchString(trimmed)
}

// EvaluateFunction evaluates FEEL function call
// Вычисляет вызов FEEL функции
func (fe *FunctionEvaluator) EvaluateFunction(
	expression string,
	variables map[string]interface{},
) (interface{}, error) {
	fe.logger.Debug("Evaluating function",
		logger.String("expression", expression))

	// Remove leading = if present (FEEL expression prefix) and trim spaces
	// Убираем ведущий = если есть (префикс FEEL выражения) и пробелы
	trimmed := strings.TrimSpace(expression)
	if strings.HasPrefix(trimmed, "=") {
		trimmed = strings.TrimSpace(trimmed[1:])
	}

	// Parse function call
	// Парсим вызов функции
	funcName, args, err := fe.parseFunctionCall(trimmed)
	if err != nil {
		return nil, err
	}

	fe.logger.Debug("Function parsed",
		logger.String("function", funcName),
		logger.Int("args_count", len(args)))

	// Evaluate arguments recursively
	// Вычисляем аргументы рекурсивно
	evaluatedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		evaluatedArg, err := fe.evaluateArgument(arg, variables)
		if err != nil {
			fe.logger.Warn("Failed to evaluate argument",
				logger.Int("arg_index", i),
				logger.String("arg", arg),
				logger.String("error", err.Error()))
			return nil, fmt.Errorf("failed to evaluate argument %d: %w", i, err)
		}
		evaluatedArgs[i] = evaluatedArg
	}

	// Execute function
	// Выполняем функцию
	switch funcName {
	case "duration":
		return fe.executeDuration(evaluatedArgs)
	case "subtract":
		return fe.executeSubtract(evaluatedArgs)
	case "add":
		return fe.executeAdd(evaluatedArgs)
	default:
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
}

// parseFunctionCall parses function name and arguments
// Парсит имя функции и аргументы
func (fe *FunctionEvaluator) parseFunctionCall(expr string) (string, []string, error) {
	matches := fe.functionCallRegex.FindStringSubmatch(expr)
	if matches == nil {
		return "", nil, fmt.Errorf("invalid function call syntax: %s", expr)
	}

	funcName := matches[1]
	argsString := strings.TrimSpace(matches[2])

	// Parse arguments (simple comma split, handling nested calls)
	// Парсим аргументы (простое разделение по запятой, обработка вложенных вызовов)
	args := fe.splitArguments(argsString)

	return funcName, args, nil
}

// splitArguments splits function arguments by comma, respecting nested calls and quotes
// Разделяет аргументы функции по запятой, учитывая вложенные вызовы и кавычки
func (fe *FunctionEvaluator) splitArguments(argsString string) []string {
	if argsString == "" {
		return []string{}
	}

	var args []string
	var currentArg strings.Builder
	depth := 0
	inQuotes := false
	quoteChar := rune(0)

	for _, char := range argsString {
		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			currentArg.WriteRune(char)

		case char == '(' && !inQuotes:
			depth++
			currentArg.WriteRune(char)

		case char == ')' && !inQuotes:
			depth--
			currentArg.WriteRune(char)

		case char == ',' && depth == 0 && !inQuotes:
			args = append(args, strings.TrimSpace(currentArg.String()))
			currentArg.Reset()

		default:
			currentArg.WriteRune(char)
		}
	}

	// Add last argument
	// Добавляем последний аргумент
	if currentArg.Len() > 0 {
		args = append(args, strings.TrimSpace(currentArg.String()))
	}

	return args
}

// evaluateArgument evaluates single function argument
// Вычисляет один аргумент функции
func (fe *FunctionEvaluator) evaluateArgument(arg string, variables map[string]interface{}) (interface{}, error) {
	// If argument is a function call, evaluate it recursively
	// Если аргумент - вызов функции, вычисляем рекурсивно
	if fe.IsFunctionCall(arg) {
		return fe.EvaluateFunction(arg, variables)
	}

	// If argument is a quoted string, remove quotes
	// Если аргумент - строка в кавычках, убираем кавычки
	if (strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"")) ||
		(strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) {
		return arg[1 : len(arg)-1], nil
	}

	// If argument is a variable name, get from variables
	// Если аргумент - имя переменной, получаем из переменных
	if value, exists := variables[arg]; exists {
		return value, nil
	}

	// Return as literal string
	// Возвращаем как литеральную строку
	return arg, nil
}

// executeDuration executes duration() function
// Выполняет функцию duration()
func (fe *FunctionEvaluator) executeDuration(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("duration() requires exactly 1 argument, got %d", len(args))
	}

	durationStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("duration() argument must be string, got %T", args[0])
	}

	// Validate duration format by parsing
	// Валидируем формат длительности парсингом
	_, err := fe.durationParser.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	fe.logger.Debug("Duration parsed",
		logger.String("duration", durationStr))

	// Return duration string as is
	// Возвращаем строку длительности как есть
	return durationStr, nil
}

// executeSubtract executes subtract() function
// Выполняет функцию subtract()
func (fe *FunctionEvaluator) executeSubtract(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("subtract() requires exactly 2 arguments, got %d", len(args))
	}

	// Parse datetime
	// Парсим дату-время
	datetimeStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("subtract() first argument must be datetime string, got %T", args[0])
	}

	datetime, err := fe.parseISO8601DateTime(datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %w", err)
	}

	// Parse duration
	// Парсим длительность
	durationStr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("subtract() second argument must be duration string, got %T", args[1])
	}

	duration, err := fe.durationParser.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	// Subtract duration from datetime
	// Вычитаем длительность из даты-времени
	result := datetime.Add(-duration)

	resultStr := fe.formatISO8601DateTime(result)

	fe.logger.Debug("Subtract executed",
		logger.String("datetime", datetimeStr),
		logger.String("duration", durationStr),
		logger.String("result", resultStr))

	return resultStr, nil
}

// executeAdd executes add() function
// Выполняет функцию add()
func (fe *FunctionEvaluator) executeAdd(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("add() requires exactly 2 arguments, got %d", len(args))
	}

	// Parse datetime
	// Парсим дату-время
	datetimeStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("add() first argument must be datetime string, got %T", args[0])
	}

	datetime, err := fe.parseISO8601DateTime(datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %w", err)
	}

	// Parse duration
	// Парсим длительность
	durationStr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("add() second argument must be duration string, got %T", args[1])
	}

	duration, err := fe.durationParser.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	// Add duration to datetime
	// Добавляем длительность к дате-времени
	result := datetime.Add(duration)

	resultStr := fe.formatISO8601DateTime(result)

	fe.logger.Debug("Add executed",
		logger.String("datetime", datetimeStr),
		logger.String("duration", durationStr),
		logger.String("result", resultStr))

	return resultStr, nil
}

// parseISO8601DateTime parses ISO 8601 datetime string
// Парсит строку даты-времени ISO 8601
func (fe *FunctionEvaluator) parseISO8601DateTime(dateStr string) (time.Time, error) {
	// Try RFC3339Nano first (with milliseconds)
	// Сначала пробуем RFC3339Nano (с миллисекундами)
	t, err := time.Parse(time.RFC3339Nano, dateStr)
	if err == nil {
		return t, nil
	}

	// Try RFC3339 (without milliseconds)
	// Пробуем RFC3339 (без миллисекунд)
	t, err = time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid ISO 8601 datetime format: %s", dateStr)
}

// formatISO8601DateTime formats time to ISO 8601 string
// Форматирует время в строку ISO 8601
func (fe *FunctionEvaluator) formatISO8601DateTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

