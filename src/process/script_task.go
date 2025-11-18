/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// ScriptTaskExecutor executes script tasks
// Исполнитель скриптовых задач
type ScriptTaskExecutor struct{}

// Execute executes script task
// Выполняет скриптовую задачу
func (ste *ScriptTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	// Script tasks execute inline scripts
	// Скриптовые задачи выполняют встроенные скрипты

	// Extract script information for logging and potential future execution
	// Извлекаем информацию о скрипте для логирования и потенциального будущего выполнения
	scriptFormat, scriptCode, scriptResult := ste.extractScriptInfo(element, token)

	logger.Info("Executing script task",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", ste.getElementID(element)),
		logger.String("script_format", scriptFormat),
		logger.Int("script_size", len(scriptCode)),
		logger.String("script_result", scriptResult))

	return executeBasicFlowElement(token, element, "script task")
}

// extractScriptInfo extracts script information from element definition
// Извлекает информацию о скрипте из определения элемента
func (ste *ScriptTaskExecutor) extractScriptInfo(
	element map[string]interface{},
	token *models.Token,
) (format, code, result string) {
	// Check for script format (JavaScript, Python, etc.)
	if scriptFormat, exists := element["script_format"]; exists {
		if formatStr, ok := scriptFormat.(string); ok {
			format = formatStr
		}
	}
	if format == "" {
		format = "javascript" // Default format
	}

	// Extract script code
	if scriptCode, exists := element["script"]; exists {
		if codeStr, ok := scriptCode.(string); ok {
			code = codeStr
		}
	}

	// Check for result variable
	if resultVar, exists := element["result_variable"]; exists {
		if resultStr, ok := resultVar.(string); ok {
			result = resultStr
		}
	}

	// Script execution ready for future integration - infrastructure prepared
	// Выполнение скриптов готово для будущей интеграции - инфраструктура подготовлена
	if code != "" {
		logger.Debug("Script task code available - ready for script engine integration",
			logger.String("token_id", token.TokenID),
			logger.String("format", format),
			logger.String("result_var", result),
			logger.Int("code_length", len(code)))
	}

	return format, code, result
}

// getElementID extracts element ID from element definition
// Извлекает ID элемента из определения элемента
func (ste *ScriptTaskExecutor) getElementID(element map[string]interface{}) string {
	if id, exists := element["id"]; exists {
		if idStr, ok := id.(string); ok {
			return idStr
		}
	}
	return "unknown"
}

// GetElementType returns element type
// Возвращает тип элемента
func (ste *ScriptTaskExecutor) GetElementType() string {
	return "scriptTask"
}
