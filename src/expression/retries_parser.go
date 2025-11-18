/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"strconv"
	"strings"

	"atom-engine/src/core/logger"
)

// RetriesParser retries count parser
// Парсер количества повторов
type RetriesParser struct {
	logger            logger.ComponentLogger
	variableEvaluator *VariableEvaluator
}

// NewRetriesParser creates new retries parser
// Создает новый парсер повторов
func NewRetriesParser(logger logger.ComponentLogger) *RetriesParser {
	return &RetriesParser{
		logger:            logger,
		variableEvaluator: NewVariableEvaluator(logger),
	}
}

// NewRetriesParserWithVariableEvaluator creates new retries parser with shared VariableEvaluator
// Создает новый парсер повторов с общим VariableEvaluator
func NewRetriesParserWithVariableEvaluator(
	logger logger.ComponentLogger,
	variableEvaluator *VariableEvaluator,
) *RetriesParser {
	return &RetriesParser{
		logger:            logger,
		variableEvaluator: variableEvaluator,
	}
}

// ParseRetries parses retries count from string
// Парсит количество повторов из строки
func (rp *RetriesParser) ParseRetries(retriesStr string) (int, error) {
	return rp.ParseRetriesWithVariables(retriesStr, nil)
}

// ParseRetriesWithVariables parses retries count from string with variable support
// Парсит количество повторов из строки с поддержкой переменных
func (rp *RetriesParser) ParseRetriesWithVariables(retriesStr string, variables map[string]interface{}) (int, error) {
	if retriesStr == "" {
		rp.logger.Debug("Empty retries string, using default")
		return 3, nil // default value
	}

	// If this is expression ${variable} or #{variable}, evaluate it
	// Если это выражение ${variable} или #{variable}, вычисляем его
	if (strings.HasPrefix(retriesStr, "${") && strings.HasSuffix(retriesStr, "}")) ||
		(strings.HasPrefix(retriesStr, "#{") && strings.HasSuffix(retriesStr, "}")) {

		if variables == nil {
			rp.logger.Warn("Expression-based retries provided but no variables available, using default",
				logger.String("expression", retriesStr))
			return 3, nil
		}

		// Evaluate expression using VariableEvaluator
		// Вычисляем выражение используя VariableEvaluator
		result, err := rp.variableEvaluator.EvaluateVariable(retriesStr, variables)
		if err != nil {
			rp.logger.Warn("Failed to evaluate retries expression, using default",
				logger.String("expression", retriesStr),
				logger.String("error", err.Error()))
			return 3, nil
		}

		// Convert result to int
		// Преобразуем результат в int
		switch v := result.(type) {
		case int:
			return rp.validateRetries(v), nil
		case int64:
			return rp.validateRetries(int(v)), nil
		case float64:
			return rp.validateRetries(int(v)), nil
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				return rp.validateRetries(parsed), nil
			}
			rp.logger.Warn("Expression result is not a valid number, using default",
				logger.String("expression", retriesStr),
				logger.String("result", v))
			return 3, nil
		default:
			rp.logger.Warn("Expression result is not a number, using default",
				logger.String("expression", retriesStr),
				logger.Any("result", result))
			return 3, nil
		}
	}

	// Parse as direct number
	// Парсим как прямое число
	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		rp.logger.Warn("Invalid retries value, using default",
			logger.String("value", retriesStr),
			logger.String("error", err.Error()))
		return 3, nil
	}

	return rp.validateRetries(retries), nil
}

// validateRetries validates and limits retries value
// Валидирует и ограничивает значение retries
func (rp *RetriesParser) validateRetries(retries int) int {
	if retries < 0 {
		rp.logger.Warn("Negative retries value, using 0",
			logger.Int("value", retries))
		return 0
	}

	if retries > 100 {
		rp.logger.Warn("Too many retries, limiting to 100",
			logger.Int("value", retries))
		return 100
	}

	rp.logger.Debug("Retries validated successfully",
		logger.Int("result", retries))
	return retries
}
