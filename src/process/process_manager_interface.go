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

// ProcessManagerInterface handles process instance operations
// Интерфейс менеджера процессов
type ProcessManagerInterface interface {
	// Process instance lifecycle
	StartProcessInstance(processKey string, variables map[string]interface{}) (*models.ProcessInstance, error)
	GetProcessInstanceStatus(instanceID string) (*models.ProcessInstance, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*models.ProcessInstance, error)
}
