/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"atom-engine/src/core/logger"
)

// PathNavigator navigates through nested structures using FEEL path expressions
// Навигатор по вложенным структурам используя FEEL path выражения
type PathNavigator struct {
	logger logger.ComponentLogger
}

// NewPathNavigator creates new path navigator
// Создает новый навигатор путей
func NewPathNavigator(logger logger.ComponentLogger) *PathNavigator {
	return &PathNavigator{
		logger: logger,
	}
}

// PathSegment represents a segment in a path
// Сегмент пути
type PathSegment struct {
	Field      string // Field name for map/struct access
	Index      int    // Array index (if IsIndex = true)
	IndexExpr  string // Expression for index (if IsIndexExpr = true)
	IsIndex    bool   // True if this is array index access [N]
	IsIndexExpr bool  // True if this is expression-based index [expr]
}

// NavigatePath navigates through nested structures using path expression
// Навигация по вложенным структурам используя path выражение
// Examples:
//   - "response.body.data" -> extracts response["body"]["data"]
//   - "items[0].name" -> extracts items[0]["name"]
//   - "users[i].emails[0]" -> extracts users[i]["emails"][0] where i is variable
func (pn *PathNavigator) NavigatePath(
	path string,
	variables map[string]interface{},
) (interface{}, error) {
	pn.logger.Debug("Navigating path",
		logger.String("path", path))

	// Parse path into segments
	// Парсим путь на сегменты
	segments, err := pn.parsePathSegments(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}

	pn.logger.Debug("Path parsed into segments",
		logger.String("path", path),
		logger.Int("segments_count", len(segments)))

	if len(segments) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	// Start with root variable
	// Начинаем с корневой переменной
	rootSegment := segments[0]
	if rootSegment.IsIndex || rootSegment.IsIndexExpr {
		return nil, fmt.Errorf("path cannot start with index: %s", path)
	}

	current, exists := variables[rootSegment.Field]
	if !exists {
		// Return nil (null) for missing root variable per FEEL standard
		// Возвращаем nil (null) для отсутствующей корневой переменной по стандарту FEEL
		pn.logger.Debug("Root variable not found, returning null",
			logger.String("variable", rootSegment.Field))
		return nil, nil
	}

	pn.logger.Debug("Root variable found",
		logger.String("variable", rootSegment.Field),
		logger.String("type", fmt.Sprintf("%T", current)))

	// Navigate through remaining segments
	// Проходим через оставшиеся сегменты
	for i := 1; i < len(segments); i++ {
		segment := segments[i]
		
		pn.logger.Debug("Processing segment",
			logger.Int("index", i),
			logger.String("field", segment.Field),
			logger.Bool("is_index", segment.IsIndex),
			logger.Bool("is_index_expr", segment.IsIndexExpr))

		var err error
		if segment.IsIndex {
			// Array index access
			// Доступ по индексу массива
			current, err = pn.navigateArrayIndex(current, segment.Index)
		} else if segment.IsIndexExpr {
			// Dynamic index/key access (variable-based)
			// Динамический доступ по индексу/ключу (на основе переменной)
			current, err = pn.navigateDynamicAccess(current, segment.IndexExpr, variables)
		} else {
			// Map/struct field access
			// Доступ к полю map/struct
			current, err = pn.navigateMapField(current, segment.Field)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to navigate segment %d (%s): %w", i, segment.Field, err)
		}
	}

	pn.logger.Debug("Path navigation successful",
		logger.String("path", path),
		logger.Any("result", current),
		logger.String("result_type", fmt.Sprintf("%T", current)))

	return current, nil
}

// parsePathSegments parses path string into segments
// Парсит строку пути на сегменты
// Examples:
//   - "response.body.data" -> [{Field: "response"}, {Field: "body"}, {Field: "data"}]
//   - "items[0].name" -> [{Field: "items"}, {Index: 0, IsIndex: true}, {Field: "name"}]
//   - "data[key]" -> [{Field: "data"}, {IndexExpr: "key", IsIndexExpr: true}]
func (pn *PathNavigator) parsePathSegments(path string) ([]PathSegment, error) {
	segments := make([]PathSegment, 0)
	
	current := ""
	inBracket := false
	bracketContent := ""

	for i := 0; i < len(path); i++ {
		char := path[i]

		switch char {
		case '.':
			if inBracket {
				bracketContent += string(char)
			} else {
				if current != "" {
					segments = append(segments, PathSegment{Field: current})
					current = ""
				}
			}

		case '[':
			if inBracket {
				return nil, fmt.Errorf("nested brackets not supported at position %d", i)
			}
			if current != "" {
				segments = append(segments, PathSegment{Field: current})
				current = ""
			}
			inBracket = true
			bracketContent = ""

		case ']':
			if !inBracket {
				return nil, fmt.Errorf("unexpected closing bracket at position %d", i)
			}
			// Process bracket content
			// Обрабатываем содержимое скобок
			if bracketContent == "" {
				return nil, fmt.Errorf("empty bracket at position %d", i)
			}

			// Try to parse as integer index
			// Пытаемся распарсить как целочисленный индекс
			if index, err := strconv.Atoi(bracketContent); err == nil {
				segments = append(segments, PathSegment{
					Index:   index,
					IsIndex: true,
				})
			} else {
				// It's an expression (variable name)
				// Это выражение (имя переменной)
				segments = append(segments, PathSegment{
					IndexExpr:   bracketContent,
					IsIndexExpr: true,
				})
			}

			inBracket = false
			bracketContent = ""

		default:
			if inBracket {
				bracketContent += string(char)
			} else {
				current += string(char)
			}
		}
	}

	if inBracket {
		return nil, fmt.Errorf("unclosed bracket")
	}

	if current != "" {
		segments = append(segments, PathSegment{Field: current})
	}

	return segments, nil
}

// navigateMapField navigates to a field in map or struct
// Навигация к полю в map или struct
func (pn *PathNavigator) navigateMapField(obj interface{}, field string) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot access field '%s' on nil object", field)
	}

	// Handle map[string]interface{}
	// Обрабатываем map[string]interface{}
	if mapObj, ok := obj.(map[string]interface{}); ok {
		value, exists := mapObj[field]
		if !exists {
			// Return nil (null) for missing field per FEEL standard
			// Возвращаем nil (null) для отсутствующего поля по стандарту FEEL
			pn.logger.Debug("Field not found in map, returning null",
				logger.String("field", field))
			return nil, nil
		}
		return value, nil
	}

	// Handle struct using reflection
	// Обрабатываем struct используя рефлексию
	val := reflect.ValueOf(obj)
	
	// Dereference pointer if needed
	// Разыменовываем указатель если нужно
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("cannot access field '%s' on nil pointer", field)
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		// Try to find field by name (case-insensitive)
		// Пытаемся найти поле по имени (без учета регистра)
		fieldVal := val.FieldByName(field)
		if !fieldVal.IsValid() {
			// Try case-insensitive search
			// Пробуем поиск без учета регистра
			typ := val.Type()
			for i := 0; i < typ.NumField(); i++ {
				if strings.EqualFold(typ.Field(i).Name, field) {
					fieldVal = val.Field(i)
					break
				}
			}
		}

		if !fieldVal.IsValid() {
			return nil, fmt.Errorf("field '%s' not found in struct %T", field, obj)
		}

		return fieldVal.Interface(), nil
	}

	return nil, fmt.Errorf("cannot access field '%s' on type %T", field, obj)
}

// navigateArrayIndex navigates to an array element by index
// Навигация к элементу массива по индексу
func (pn *PathNavigator) navigateArrayIndex(obj interface{}, index int) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot access index %d on nil object", index)
	}

	// Handle []interface{}
	// Обрабатываем []interface{}
	if slice, ok := obj.([]interface{}); ok {
		if index < 0 || index >= len(slice) {
			return nil, fmt.Errorf("index %d out of bounds (length: %d)", index, len(slice))
		}
		return slice[index], nil
	}

	// Handle reflection for other slice types
	// Обрабатываем рефлексию для других типов slice
	val := reflect.ValueOf(obj)
	
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		if index < 0 || index >= val.Len() {
			return nil, fmt.Errorf("index %d out of bounds (length: %d)", index, val.Len())
		}
		return val.Index(index).Interface(), nil
	}

	return nil, fmt.Errorf("cannot access index %d on non-array type %T", index, obj)
}

// navigateDynamicAccess navigates using dynamic variable-based access
// Навигация используя динамический доступ на основе переменной
// Handles both array index access (when variable is integer) and map key access (when variable is string)
func (pn *PathNavigator) navigateDynamicAccess(obj interface{}, expr string, variables map[string]interface{}) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot access dynamic index/key '%s' on nil object", expr)
	}

	// Get variable value
	// Получаем значение переменной
	value, exists := variables[expr]
	if !exists {
		return nil, fmt.Errorf("variable '%s' not found", expr)
	}

	// Check if obj is a map - use value as string key
	// Проверяем является ли obj map - используем значение как строковый ключ
	if mapObj, ok := obj.(map[string]interface{}); ok {
		// Use value as string key
		// Используем значение как строковый ключ
		keyStr := fmt.Sprintf("%v", value)
		result, exists := mapObj[keyStr]
		if !exists {
			// Return nil (null) for missing key per FEEL standard
			// Возвращаем nil (null) для отсутствующего ключа по стандарту FEEL
			pn.logger.Debug("Key not found in map, returning null",
				logger.String("key", keyStr))
			return nil, nil
		}
		return result, nil
	}

	// Check if obj is array/slice - convert value to integer index
	// Проверяем является ли obj массивом/slice - конвертируем значение в целочисленный индекс
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		// Convert to integer
		// Конвертируем в integer
		var index int
		switch v := value.(type) {
		case int:
			index = v
		case int32:
			index = int(v)
		case int64:
			index = int(v)
		case float64:
			index = int(v)
		case float32:
			index = int(v)
		case string:
			// Try to parse string as integer
			// Пытаемся распарсить строку как integer
			var err error
			index, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("cannot convert string '%s' to integer for array access: %w", v, err)
			}
		default:
			return nil, fmt.Errorf("cannot convert %T to integer index for array access", value)
		}

		return pn.navigateArrayIndex(obj, index)
	}

	return nil, fmt.Errorf("cannot access dynamic index/key '%s' on type %T", expr, obj)
}

