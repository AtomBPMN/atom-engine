/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

// JobCallbackManagerInterface handles job callback operations
// Интерфейс менеджера job callback операций
type JobCallbackManagerInterface interface {
	// Job callback operations
	HandleJobCallback(jobID, elementID, tokenID, status, errorMessage string, variables map[string]interface{}) error
}
