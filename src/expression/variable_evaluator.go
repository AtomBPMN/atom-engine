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

// VariableEvaluator variable processor
// Обработчик переменных
type VariableEvaluator struct {
	logger        logger.ComponentLogger
	pathNavigator *PathNavigator
}

// NewVariableEvaluator creates new variable processor
// Создает новый обработчик переменных
func NewVariableEvaluator(logger logger.ComponentLogger) *VariableEvaluator {
	return &VariableEvaluator{
		logger:        logger,
		pathNavigator: NewPathNavigator(logger),
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
		
		// Check if it's a logical expression (contains and, or, not)
		// Проверяем является ли это логическим выражением (содержит and, or, not)
		if ve.isLogicalExpression(feelExpr) {
			result, err := ve.evaluateLogicalExpression(feelExpr, variables)
			if err != nil {
				ve.logger.Warn("Logical expression evaluation failed",
					logger.String("expression", feelExpr),
					logger.String("error", err.Error()))
				return false, err
			}
			ve.logger.Debug("Logical expression evaluation successful",
				logger.String("expression", feelExpr),
				logger.Bool("result", result))
			return result, nil
		}
		
		// Check if it's a comparison expression (contains ==, !=, >=, <=, >, <)
		// Проверяем является ли это выражением сравнения (содержит ==, !=, >=, <=, >, <)
		if ve.isComparisonExpression(feelExpr) {
			result, err := ve.evaluateComparison(feelExpr, variables)
			if err != nil {
				ve.logger.Warn("Comparison evaluation failed",
					logger.String("expression", feelExpr),
					logger.String("error", err.Error()))
				return false, err
			}
			ve.logger.Debug("Comparison evaluation successful",
				logger.String("expression", feelExpr),
				logger.Any("result", result))
			return result, nil
		}
		
		// Check if it's a path expression vs string with variables
		// Различаем path выражения и строки с переменными
		// Path expression: response.body.data (no /)
		// String with variables: api_url/nodes/params.newid (has /)
		if (strings.Contains(feelExpr, ".") || strings.Contains(feelExpr, "[")) && !strings.Contains(feelExpr, "/") {
			// Use PathNavigator for complex paths (no slashes)
			// Используем PathNavigator для сложных путей (без слешей)
			result, err := ve.pathNavigator.NavigatePath(feelExpr, variables)
			if err != nil {
				ve.logger.Warn("Path navigation failed",
					logger.String("path", feelExpr),
					logger.String("error", err.Error()))
				// Fallback to existing logic
				// Откатываемся к существующей логике
			} else {
				ve.logger.Debug("Path navigation successful",
					logger.String("path", feelExpr),
					logger.Any("result", result),
					logger.String("result_type", fmt.Sprintf("%T", result)))
				return result, nil
			}
		}
		
		// Handle simple variable access in FEEL
		// Обрабатываем простой доступ к переменным в FEEL
		if value, exists := variables[feelExpr]; exists {
			ve.logger.Debug("FEEL variable found",
				logger.String("variable", feelExpr),
				logger.Any("value", value))
			return value, nil
		}
		// Try to replace variables in string expression
		// Пытаемся заменить переменные в строковом выражении
		replaced := ve.replaceVariablesInString(feelExpr, variables)
		if replaced != feelExpr {
			ve.logger.Debug("FEEL expression with variables replaced",
				logger.String("original", feelExpr),
				logger.String("replaced", replaced))
			return replaced, nil
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

// isWordChar checks if character is a word character (letter, digit, underscore)
// Проверяет является ли символ словесным (буква, цифра, подчеркивание)
func (ve *VariableEvaluator) isWordChar(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_'
}

// replaceVariablesInString replaces variable names and paths in string with their values
// Заменяет имена переменных и пути в строке на их значения
func (ve *VariableEvaluator) replaceVariablesInString(
	str string,
	variables map[string]interface{},
) string {
	result := ""
	i := 0

	for i < len(str) {
		// Check if this is the start of a variable path
		// Проверяем является ли это началом пути переменной
		if ve.isVarStartChar(str[i]) {
			// Scan the full path (including dots for nested access)
			// Сканируем полный путь (включая точки для вложенного доступа)
			pathStart := i
			path := ve.scanVariablePath(str, i)
			pathEnd := pathStart + len(path)

			// Check word boundary before
			// Проверяем границу слова до
			beforeOK := pathStart == 0 || !ve.isWordChar(str[pathStart-1])

			// Check word boundary after (dot is OK for paths)
			// Проверяем границу слова после (точка допустима для путей)
			afterOK := pathEnd >= len(str) || !ve.isWordChar(str[pathEnd])

			if beforeOK && afterOK && path != "" {
				// Try to resolve the path
				// Пытаемся разрешить путь
				value, found := ve.resolveVariablePath(path, variables)
				if found {
					result += ve.formatValueForString(value)
					ve.logger.Debug("Variable path replaced in string",
						logger.String("path", path),
						logger.Any("value", value),
						logger.String("position", fmt.Sprintf("%d-%d", pathStart, pathEnd)))
					i = pathEnd
					continue
				}
			}
		}

		// No variable found at this position, keep the character
		// Переменная не найдена на этой позиции, сохраняем символ
		result += string(str[i])
		i++
	}

	return result
}

// isVarStartChar checks if character can start a variable name
// Проверяет может ли символ начинать имя переменной
func (ve *VariableEvaluator) isVarStartChar(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		char == '_'
}

// scanVariablePath scans a variable path from the given position
// Сканирует путь переменной с заданной позиции
func (ve *VariableEvaluator) scanVariablePath(str string, start int) string {
	i := start
	path := ""

	// Scan the first part (variable name)
	// Сканируем первую часть (имя переменной)
	for i < len(str) && ve.isWordChar(str[i]) {
		path += string(str[i])
		i++
	}

	// Continue scanning through dots and subsequent parts
	// Продолжаем сканирование через точки и последующие части
	for i < len(str) {
		if str[i] == '.' {
			// Check if there's a valid identifier after the dot
			// Проверяем есть ли валидный идентификатор после точки
			if i+1 < len(str) && ve.isVarStartChar(str[i+1]) {
				path += "."
				i++
				// Scan the next part
				// Сканируем следующую часть
				for i < len(str) && ve.isWordChar(str[i]) {
					path += string(str[i])
					i++
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	return path
}

// resolveVariablePath resolves a variable path to its value
// Разрешает путь переменной в значение
func (ve *VariableEvaluator) resolveVariablePath(
	path string,
	variables map[string]interface{},
) (interface{}, bool) {
	// Check if it's a simple variable first
	// Сначала проверяем является ли это простой переменной
	if !strings.Contains(path, ".") {
		value, exists := variables[path]
		return value, exists
	}

	// Use PathNavigator for complex paths
	// Используем PathNavigator для сложных путей
	value, err := ve.pathNavigator.NavigatePath(path, variables)
	if err != nil {
		ve.logger.Debug("Failed to resolve variable path",
			logger.String("path", path),
			logger.String("error", err.Error()))
		return nil, false
	}

	return value, true
}

// isLogicalExpression checks if expression contains logical operators
// Проверяет содержит ли выражение логические операторы
func (ve *VariableEvaluator) isLogicalExpression(expr string) bool {
	// Check for logical operators with word boundaries
	// Проверяем логические операторы с границами слов
	return strings.Contains(expr, " and ") ||
		strings.Contains(expr, " or ") ||
		strings.HasPrefix(expr, "not ") ||
		strings.Contains(expr, " not ") ||
		strings.Contains(expr, "(") // Parentheses indicate complex expression
}

// isComparisonExpression checks if expression contains comparison operators
// Проверяет содержит ли выражение операторы сравнения
func (ve *VariableEvaluator) isComparisonExpression(expr string) bool {
	return strings.Contains(expr, "==") ||
		strings.Contains(expr, "!=") ||
		strings.Contains(expr, ">=") ||
		strings.Contains(expr, "<=") ||
		strings.Contains(expr, ">") ||
		strings.Contains(expr, "<")
}

// evaluateComparison evaluates comparison expression
// Вычисляет выражение сравнения
func (ve *VariableEvaluator) evaluateComparison(
	expr string,
	variables map[string]interface{},
) (bool, error) {
	ve.logger.Debug("Evaluating comparison expression",
		logger.String("expression", expr))

	// Try operators in order: ==, !=, >=, <=, >, < (longer first to avoid partial matches)
	// Пробуем операторы по порядку: ==, !=, >=, <=, >, < (длинные первыми чтобы избежать частичных совпадений)
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}
	
	for _, op := range operators {
		if strings.Contains(expr, op) {
			parts := strings.SplitN(expr, op, 2)
			if len(parts) != 2 {
				continue
			}

			leftExpr := strings.TrimSpace(parts[0])
			rightExpr := strings.TrimSpace(parts[1])

			ve.logger.Debug("Comparison parts identified",
				logger.String("operator", op),
				logger.String("left", leftExpr),
				logger.String("right", rightExpr))

			// Evaluate left side
			// Вычисляем левую часть
			leftValue, err := ve.evaluateExpressionPart(leftExpr, variables)
			if err != nil {
				return false, fmt.Errorf("failed to evaluate left side '%s': %w", leftExpr, err)
			}

			// Evaluate right side
			// Вычисляем правую часть
			rightValue, err := ve.evaluateExpressionPart(rightExpr, variables)
			if err != nil {
				return false, fmt.Errorf("failed to evaluate right side '%s': %w", rightExpr, err)
			}

			ve.logger.Debug("Comparison values evaluated",
				logger.String("operator", op),
				logger.Any("left_value", leftValue),
				logger.String("left_type", fmt.Sprintf("%T", leftValue)),
				logger.Any("right_value", rightValue),
				logger.String("right_type", fmt.Sprintf("%T", rightValue)))

			// Perform comparison
			// Выполняем сравнение
			result, err := ve.compareValues(leftValue, rightValue, op)
			if err != nil {
				return false, fmt.Errorf("comparison failed: %w", err)
			}

			ve.logger.Info("Comparison result",
				logger.String("expression", expr),
				logger.String("operator", op),
				logger.Any("left", leftValue),
				logger.Any("right", rightValue),
				logger.Bool("result", result))

			return result, nil
		}
	}

	return false, fmt.Errorf("no comparison operator found in expression: %s", expr)
}

// evaluateExpressionPart evaluates a single part of comparison expression
// Вычисляет одну часть выражения сравнения
func (ve *VariableEvaluator) evaluateExpressionPart(
	expr string,
	variables map[string]interface{},
) (interface{}, error) {
	// Check if it's a path expression (contains . or [)
	// Проверяем является ли это path выражением (содержит . или [)
	if strings.Contains(expr, ".") || strings.Contains(expr, "[") {
		result, err := ve.pathNavigator.NavigatePath(expr, variables)
		if err != nil {
			// Real error occurred during navigation
			// Произошла реальная ошибка во время навигации
			return nil, err
		}
		// Return result (can be nil for missing fields per FEEL standard)
		// Возвращаем результат (может быть nil для отсутствующих полей по стандарту FEEL)
		return result, nil
	}

	// Check if it's a simple variable
	// Проверяем является ли это простой переменной
	if value, exists := variables[expr]; exists {
		return value, nil
	}

	// Check if it's a number literal
	// Проверяем является ли это числовым литералом
	if num, ok := ve.parseNumber(expr); ok {
		return num, nil
	}

	// Check if it's a boolean literal
	// Проверяем является ли это булевым литералом
	switch strings.ToLower(expr) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	// Check if it's a string literal (with quotes)
	// Проверяем является ли это строковым литералом (с кавычками)
	if (strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"")) ||
		(strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'")) {
		return expr[1 : len(expr)-1], nil
	}

	// Return as string literal
	// Возвращаем как строковый литерал
	return expr, nil
}

// parseNumber tries to parse string as number
// Пытается распарсить строку как число
func (ve *VariableEvaluator) parseNumber(str string) (interface{}, bool) {
	// Try to parse as integer
	// Пытаемся распарсить как целое число
	var intVal int
	if _, err := fmt.Sscanf(str, "%d", &intVal); err == nil {
		return intVal, true
	}

	// Try to parse as float
	// Пытаемся распарсить как число с плавающей точкой
	var floatVal float64
	if _, err := fmt.Sscanf(str, "%f", &floatVal); err == nil {
		return floatVal, true
	}

	return nil, false
}

// compareValues compares two values using given operator
// Сравнивает два значения используя заданный оператор
func (ve *VariableEvaluator) compareValues(
	left, right interface{},
	operator string,
) (bool, error) {
	switch operator {
	case "==":
		return ve.compareEqual(left, right), nil
	case "!=":
		return !ve.compareEqual(left, right), nil
	case ">":
		return ve.compareGreater(left, right)
	case "<":
		return ve.compareLess(left, right)
	case ">=":
		return ve.compareGreaterOrEqual(left, right)
	case "<=":
		return ve.compareLessOrEqual(left, right)
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// compareEqual checks if two values are equal with FEEL null semantics
// Проверяет равенство двух значений с FEEL null семантикой
func (ve *VariableEvaluator) compareEqual(left, right interface{}) bool {
	// Handle null values per FEEL standard
	// Обрабатываем null значения по стандарту FEEL
	if left == nil && right == nil {
		return true // null == null → true
	}
	if left == nil || right == nil {
		return false // null == value → false
	}

	// Direct comparison
	// Прямое сравнение
	if left == right {
		return true
	}

	// String comparison
	// Сравнение как строки
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return leftStr == rightStr
}

// compareGreater checks if left > right with FEEL null semantics
// Проверяет если left > right с FEEL null семантикой
func (ve *VariableEvaluator) compareGreater(left, right interface{}) (bool, error) {
	// Handle null values per FEEL standard: null > value → false
	// Обрабатываем null значения по стандарту FEEL: null > value → false
	if left == nil || right == nil {
		return false, nil
	}

	leftNum, rightNum, err := ve.convertToNumbers(left, right)
	if err != nil {
		return false, err
	}
	return leftNum > rightNum, nil
}

// compareLess checks if left < right with FEEL null semantics
// Проверяет если left < right с FEEL null семантикой
func (ve *VariableEvaluator) compareLess(left, right interface{}) (bool, error) {
	// Handle null values per FEEL standard: null < value → false
	// Обрабатываем null значения по стандарту FEEL: null < value → false
	if left == nil || right == nil {
		return false, nil
	}

	leftNum, rightNum, err := ve.convertToNumbers(left, right)
	if err != nil {
		return false, err
	}
	return leftNum < rightNum, nil
}

// compareGreaterOrEqual checks if left >= right with FEEL null semantics
// Проверяет если left >= right с FEEL null семантикой
func (ve *VariableEvaluator) compareGreaterOrEqual(left, right interface{}) (bool, error) {
	// Handle null values per FEEL standard
	// Обрабатываем null значения по стандарту FEEL
	if left == nil && right == nil {
		return true, nil // null >= null → true
	}
	if left == nil || right == nil {
		return false, nil // null >= value → false
	}

	leftNum, rightNum, err := ve.convertToNumbers(left, right)
	if err != nil {
		return false, err
	}
	return leftNum >= rightNum, nil
}

// compareLessOrEqual checks if left <= right with FEEL null semantics
// Проверяет если left <= right с FEEL null семантикой
func (ve *VariableEvaluator) compareLessOrEqual(left, right interface{}) (bool, error) {
	// Handle null values per FEEL standard
	// Обрабатываем null значения по стандарту FEEL
	if left == nil && right == nil {
		return true, nil // null <= null → true
	}
	if left == nil || right == nil {
		return false, nil // null <= value → false
	}

	leftNum, rightNum, err := ve.convertToNumbers(left, right)
	if err != nil {
		return false, err
	}
	return leftNum <= rightNum, nil
}

// convertToNumbers converts two values to float64 for numeric comparison
// Конвертирует два значения в float64 для числового сравнения
func (ve *VariableEvaluator) convertToNumbers(left, right interface{}) (float64, float64, error) {
	leftNum, err := ve.toFloat64(left)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert left value to number: %w", err)
	}

	rightNum, err := ve.toFloat64(right)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert right value to number: %w", err)
	}

	return leftNum, rightNum, nil
}

// toFloat64 converts value to float64
// Конвертирует значение в float64
func (ve *VariableEvaluator) toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("cannot parse string '%s' as number", v)
	default:
		return 0, fmt.Errorf("cannot convert type %T to number", value)
	}
}

// formatValueForString formats value for string concatenation without scientific notation
// Форматирует значение для конкатенации строк без научной нотации
func (ve *VariableEvaluator) formatValueForString(value interface{}) string {
	switch v := value.(type) {
	case float64:
		// Check if it's an integer value
		// Проверяем является ли это целым числом
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%f", v)
	case float32:
		if v == float32(int32(v)) {
			return fmt.Sprintf("%d", int32(v))
		}
		return fmt.Sprintf("%f", v)
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return v
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// tokenType represents type of token in logical expression
// Тип токена в логическом выражении
type tokenType int

const (
	tokenOperand    tokenType = iota // Operand (comparison or value)
	tokenAnd                         // and operator
	tokenOr                          // or operator
	tokenNot                         // not operator
	tokenLeftParen                   // (
	tokenRightParen                  // )
)

// token represents a token in logical expression
// Токен в логическом выражении
type token struct {
	typ   tokenType
	value string // Original expression for operands
}

// evaluateLogicalExpression evaluates logical expression with and/or/not
// Вычисляет логическое выражение с and/or/not
func (ve *VariableEvaluator) evaluateLogicalExpression(
	expr string,
	variables map[string]interface{},
) (bool, error) {
	ve.logger.Debug("Evaluating logical expression",
		logger.String("expression", expr))

	// Tokenize expression
	// Токенизируем выражение
	tokens, err := ve.tokenizeLogicalExpression(expr)
	if err != nil {
		return false, fmt.Errorf("tokenization failed: %w", err)
	}

	ve.logger.Debug("Expression tokenized",
		logger.Int("tokens_count", len(tokens)))

	// Convert infix to RPN using shunting-yard algorithm
	// Конвертируем инфиксную нотацию в постфиксную используя алгоритм сортировочной станции
	rpn, err := ve.infixToRPN(tokens)
	if err != nil {
		return false, fmt.Errorf("infix to RPN conversion failed: %w", err)
	}

	ve.logger.Debug("Expression converted to RPN",
		logger.Int("rpn_tokens_count", len(rpn)))

	// Evaluate RPN expression
	// Вычисляем RPN выражение
	result, err := ve.evaluateRPN(rpn, variables)
	if err != nil {
		return false, fmt.Errorf("RPN evaluation failed: %w", err)
	}

	// Convert null to false in boolean context per FEEL standard
	// Конвертируем null в false в булевом контексте по стандарту FEEL
	if result == nil {
		ve.logger.Debug("Result is null, converting to false in boolean context")
		return false, nil
	}

	// Convert result to boolean
	// Конвертируем результат в boolean
	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expected boolean result, got %T", result)
	}

	ve.logger.Debug("Logical expression evaluated",
		logger.Bool("result", boolResult))

	return boolResult, nil
}

// tokenizeLogicalExpression splits expression into tokens
// Разбивает выражение на токены
func (ve *VariableEvaluator) tokenizeLogicalExpression(expr string) ([]token, error) {
	tokens := make([]token, 0)
	i := 0
	expr = strings.TrimSpace(expr)

	for i < len(expr) {
		// Skip whitespace
		// Пропускаем пробелы
		for i < len(expr) && expr[i] == ' ' {
			i++
		}

		if i >= len(expr) {
			break
		}

		// Check for parentheses
		// Проверяем скобки
		if expr[i] == '(' {
			tokens = append(tokens, token{typ: tokenLeftParen, value: "("})
			i++
			continue
		}

		if expr[i] == ')' {
			tokens = append(tokens, token{typ: tokenRightParen, value: ")"})
			i++
			continue
		}

		// Check for logical operators
		// Проверяем логические операторы
		remaining := expr[i:]

		if strings.HasPrefix(remaining, "not ") || strings.HasPrefix(remaining, "not(") {
			tokens = append(tokens, token{typ: tokenNot, value: "not"})
			i += 3
			continue
		}

		if strings.HasPrefix(remaining, "and ") {
			tokens = append(tokens, token{typ: tokenAnd, value: "and"})
			i += 3
			continue
		}

		if strings.HasPrefix(remaining, "or ") {
			tokens = append(tokens, token{typ: tokenOr, value: "or"})
			i += 2
			continue
		}

		// Extract operand (everything until next operator or parenthesis)
		// Извлекаем операнд (все до следующего оператора или скобки)
		operandStart := i
		parenDepth := 0

		for i < len(expr) {
			char := expr[i]

			if char == '(' {
				parenDepth++
				i++
				continue
			}

			if char == ')' {
				if parenDepth > 0 {
					parenDepth--
					i++
					continue
				}
				// End of operand
				// Конец операнда
				break
			}

			// Check if we reached a logical operator at depth 0
			// Проверяем достигли ли мы логического оператора на уровне 0
			if parenDepth == 0 {
				rest := expr[i:]
				if strings.HasPrefix(rest, " and ") || strings.HasPrefix(rest, " or ") {
					break
				}
			}

			i++
		}

		operand := strings.TrimSpace(expr[operandStart:i])
		if operand != "" {
			tokens = append(tokens, token{typ: tokenOperand, value: operand})
		}
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	ve.logger.Debug("Tokenization complete",
		logger.Int("tokens_count", len(tokens)))

	return tokens, nil
}

// getOperatorPrecedence returns operator precedence per FEEL standard
// Возвращает приоритет оператора по стандарту FEEL
// Higher number = higher precedence: not(3) > and(2) > or(1)
func (ve *VariableEvaluator) getOperatorPrecedence(typ tokenType) int {
	switch typ {
	case tokenNot:
		return 3 // Highest precedence
	case tokenAnd:
		return 2 // Medium precedence
	case tokenOr:
		return 1 // Lowest precedence
	default:
		return 0
	}
}

// infixToRPN converts infix notation to postfix (RPN) using shunting-yard algorithm
// Конвертирует инфиксную нотацию в постфиксную (RPN) используя алгоритм сортировочной станции
func (ve *VariableEvaluator) infixToRPN(tokens []token) ([]token, error) {
	output := make([]token, 0)
	operatorStack := make([]token, 0)

	for _, tok := range tokens {
		switch tok.typ {
		case tokenOperand:
			// Operands go directly to output
			// Операнды идут напрямую в выход
			output = append(output, tok)

		case tokenNot, tokenAnd, tokenOr:
			// Pop operators with higher or equal precedence from stack
			// Извлекаем операторы с более высоким или равным приоритетом из стека
			for len(operatorStack) > 0 {
				top := operatorStack[len(operatorStack)-1]
				
				// Stop if we hit a left parenthesis
				// Останавливаемся если достигли левой скобки
				if top.typ == tokenLeftParen {
					break
				}

				// For 'not' (right-associative) use >, for 'and'/'or' (left-associative) use >=
				// Для 'not' (правоассоциативный) используем >, для 'and'/'or' (левоассоциативный) используем >=
				topPrecedence := ve.getOperatorPrecedence(top.typ)
				currentPrecedence := ve.getOperatorPrecedence(tok.typ)

				// 'not' is right-associative, others are left-associative
				// 'not' правоассоциативный, остальные левоассоциативные
				shouldPop := false
				if tok.typ == tokenNot {
					shouldPop = topPrecedence > currentPrecedence
				} else {
					shouldPop = topPrecedence >= currentPrecedence
				}

				if !shouldPop {
					break
				}

				// Pop operator to output
				// Извлекаем оператор в выход
				output = append(output, top)
				operatorStack = operatorStack[:len(operatorStack)-1]
			}

			// Push current operator to stack
			// Помещаем текущий оператор в стек
			operatorStack = append(operatorStack, tok)

		case tokenLeftParen:
			// Left parenthesis goes to stack
			// Левая скобка идет в стек
			operatorStack = append(operatorStack, tok)

		case tokenRightParen:
			// Pop operators until we find matching left parenthesis
			// Извлекаем операторы пока не найдем соответствующую левую скобку
			foundLeftParen := false
			for len(operatorStack) > 0 {
				top := operatorStack[len(operatorStack)-1]
				operatorStack = operatorStack[:len(operatorStack)-1]

				if top.typ == tokenLeftParen {
					foundLeftParen = true
					break
				}

				output = append(output, top)
			}

			if !foundLeftParen {
				return nil, fmt.Errorf("mismatched parentheses: no matching left parenthesis")
			}
		}
	}

	// Pop remaining operators from stack
	// Извлекаем оставшиеся операторы из стека
	for len(operatorStack) > 0 {
		top := operatorStack[len(operatorStack)-1]
		operatorStack = operatorStack[:len(operatorStack)-1]

		if top.typ == tokenLeftParen {
			return nil, fmt.Errorf("mismatched parentheses: unclosed left parenthesis")
		}

		output = append(output, top)
	}

	ve.logger.Debug("Infix to RPN conversion complete",
		logger.Int("input_tokens", len(tokens)),
		logger.Int("output_tokens", len(output)))

	return output, nil
}

// evaluateRPN evaluates RPN expression with FEEL null semantics
// Вычисляет RPN выражение с FEEL null семантикой
func (ve *VariableEvaluator) evaluateRPN(
	rpn []token,
	variables map[string]interface{},
) (interface{}, error) {
	stack := make([]interface{}, 0)

	for i, tok := range rpn {
		ve.logger.Debug("Processing RPN token",
			logger.Int("position", i),
			logger.String("type", fmt.Sprintf("%d", tok.typ)),
			logger.String("value", tok.value))

		switch tok.typ {
		case tokenOperand:
			// Evaluate operand and push to stack
			// Вычисляем операнд и помещаем в стек
			result, err := ve.evaluateOperand(tok.value, variables)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate operand '%s': %w", tok.value, err)
			}
			stack = append(stack, result)

			ve.logger.Debug("Operand evaluated",
				logger.String("operand", tok.value),
				logger.Any("result", result))

		case tokenNot:
			// Unary operator - pop one operand
			// Унарный оператор - извлекаем один операнд
			if len(stack) < 1 {
				return nil, fmt.Errorf("not operator requires 1 operand, got %d", len(stack))
			}

			operand := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			result := ve.applyLogicalOperator(tokenNot, operand, nil)
			stack = append(stack, result)

			ve.logger.Debug("NOT operator applied",
				logger.Any("operand", operand),
				logger.Any("result", result))

		case tokenAnd, tokenOr:
			// Binary operator - pop two operands
			// Бинарный оператор - извлекаем два операнда
			if len(stack) < 2 {
				return nil, fmt.Errorf("binary operator requires 2 operands, got %d", len(stack))
			}

			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			result := ve.applyLogicalOperator(tok.typ, left, right)
			stack = append(stack, result)

			ve.logger.Debug("Binary operator applied",
				logger.String("operator", tok.value),
				logger.Any("left", left),
				logger.Any("right", right),
				logger.Any("result", result))

		default:
			return nil, fmt.Errorf("unexpected token type in RPN: %d", tok.typ)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid RPN expression: expected 1 result, got %d", len(stack))
	}

	result := stack[0]
	ve.logger.Debug("RPN evaluation complete",
		logger.Any("result", result))

	return result, nil
}

// evaluateOperand evaluates single operand (can be comparison, path, or literal)
// Вычисляет единичный операнд (может быть сравнением, путем, или литералом)
func (ve *VariableEvaluator) evaluateOperand(
	operand string,
	variables map[string]interface{},
) (interface{}, error) {
	operand = strings.TrimSpace(operand)

	ve.logger.Debug("Evaluating operand",
		logger.String("operand", operand))

	// Check if it's a comparison expression
	// Проверяем является ли это выражением сравнения
	if ve.isComparisonExpression(operand) {
		result, err := ve.evaluateComparison(operand, variables)
		if err != nil {
			return nil, fmt.Errorf("comparison evaluation failed: %w", err)
		}
		return result, nil
	}

	// Check if it's a path expression
	// Проверяем является ли это path выражением
	if strings.Contains(operand, ".") || strings.Contains(operand, "[") {
		result, err := ve.pathNavigator.NavigatePath(operand, variables)
		if err != nil {
			return nil, fmt.Errorf("path navigation failed: %w", err)
		}
		return result, nil
	}

	// Try to get as variable
	// Пытаемся получить как переменную
	if value, exists := variables[operand]; exists {
		return value, nil
	}

	// Try to parse as boolean literal
	// Пытаемся распарсить как булев литерал
	if operand == "true" {
		return true, nil
	}
	if operand == "false" {
		return false, nil
	}

	// Try to parse as null
	// Пытаемся распарсить как null
	if operand == "null" {
		return nil, nil
	}

	// Try to parse as number
	// Пытаемся распарсить как число
	if num, err := ve.toFloat64(operand); err == nil {
		return num, nil
	}

	// Try to parse as string literal (quoted)
	// Пытаемся распарсить как строковый литерал (в кавычках)
	if (strings.HasPrefix(operand, "\"") && strings.HasSuffix(operand, "\"")) ||
		(strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'")) {
		return operand[1 : len(operand)-1], nil
	}

	// Return operand as string
	// Возвращаем операнд как строку
	return operand, nil
}

// applyLogicalOperator applies logical operator with FEEL null semantics
// Применяет логический оператор с FEEL null семантикой
func (ve *VariableEvaluator) applyLogicalOperator(
	op tokenType,
	left interface{},
	right interface{},
) interface{} {
	// Convert operands to boolean (with null support)
	// Конвертируем операнды в boolean (с поддержкой null)
	leftBool := ve.toBooleanOrNull(left)
	rightBool := ve.toBooleanOrNull(right)

	switch op {
	case tokenNot:
		// not null → null
		if leftBool == nil {
			return nil
		}
		return !leftBool.(bool)

	case tokenAnd:
		// FEEL null semantics for AND:
		// false and null → false
		// true and null → null
		// null and false → false
		// null and true → null
		// null and null → null
		if leftBool == nil && rightBool == nil {
			return nil
		}
		if leftBool == nil {
			if rightBool.(bool) == false {
				return false
			}
			return nil
		}
		if rightBool == nil {
			if leftBool.(bool) == false {
				return false
			}
			return nil
		}
		return leftBool.(bool) && rightBool.(bool)

	case tokenOr:
		// FEEL null semantics for OR:
		// true or null → true
		// false or null → null
		// null or true → true
		// null or false → null
		// null or null → null
		if leftBool == nil && rightBool == nil {
			return nil
		}
		if leftBool == nil {
			if rightBool.(bool) == true {
				return true
			}
			return nil
		}
		if rightBool == nil {
			if leftBool.(bool) == true {
				return true
			}
			return nil
		}
		return leftBool.(bool) || rightBool.(bool)

	default:
		return nil
	}
}

// toBooleanOrNull converts value to boolean or returns nil for null
// Конвертирует значение в boolean или возвращает nil для null
func (ve *VariableEvaluator) toBooleanOrNull(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool:
		return v
	case int, int32, int64, float32, float64:
		// Non-zero numbers are true
		// Ненулевые числа это true
		num, _ := ve.toFloat64(v)
		return num != 0
	case string:
		// Non-empty strings are true
		// Непустые строки это true
		return v != ""
	default:
		// Other types: non-nil is true
		// Другие типы: не-nil это true
		return true
	}
}
