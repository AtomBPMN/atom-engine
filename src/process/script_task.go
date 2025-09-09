/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/models"
)

// ScriptTaskExecutor executes script tasks
// Исполнитель скриптовых задач
type ScriptTaskExecutor struct{}

// Execute executes script task
// Выполняет скриптовую задачу
func (ste *ScriptTaskExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	// Script tasks execute inline scripts
	// In real implementation would execute script code
	return executeBasicFlowElement(token, element, "script task")
}

// GetElementType returns element type
// Возвращает тип элемента
func (ste *ScriptTaskExecutor) GetElementType() string {
	return "scriptTask"
}
