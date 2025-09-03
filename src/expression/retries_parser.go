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
	logger logger.ComponentLogger
}

// NewRetriesParser creates new retries parser
// Создает новый парсер повторов
func NewRetriesParser(logger logger.ComponentLogger) *RetriesParser {
	return &RetriesParser{
		logger: logger,
	}
}

// ParseRetries parses retries count from string
// Парсит количество повторов из строки
func (rp *RetriesParser) ParseRetries(retriesStr string) (int, error) {
	if retriesStr == "" {
		rp.logger.Debug("Empty retries string, using default")
		return 3, nil // default value
	}

	// If this is expression ${variable}, return default for now
	// Если это выражение ${variable}, то пока возвращаем значение по умолчанию
	if strings.HasPrefix(retriesStr, "${") && strings.HasSuffix(retriesStr, "}") {
		rp.logger.Warn("Expression-based retries not fully implemented, using default",
			logger.String("expression", retriesStr))
		return 3, nil
	}

	// If this is Camunda expression #{variable}, return default for now
	// Если это выражение Camunda #{variable}, то пока возвращаем значение по умолчанию
	if strings.HasPrefix(retriesStr, "#{") && strings.HasSuffix(retriesStr, "}") {
		rp.logger.Warn("Camunda expression-based retries not fully implemented, using default",
			logger.String("expression", retriesStr))
		return 3, nil
	}

	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		rp.logger.Warn("Invalid retries value, using default",
			logger.String("value", retriesStr),
			logger.String("error", err.Error()))
		return 3, nil
	}

	if retries < 0 {
		rp.logger.Warn("Negative retries value, using 0",
			logger.Int("value", retries))
		return 0, nil
	}

	if retries > 100 {
		rp.logger.Warn("Too many retries, limiting to 100",
			logger.Int("value", retries))
		return 100, nil
	}

	rp.logger.Debug("Retries parsed successfully",
		logger.String("input", retriesStr),
		logger.Int("result", retries))
	return retries, nil
}
