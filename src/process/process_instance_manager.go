/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// ProcessInstanceManager manages process instance lifecycle
// Управляет жизненным циклом экземпляров процессов
type ProcessInstanceManager struct {
	storage        storage.Storage
	component      ComponentInterface
	processStarter *ProcessStarter
}

// NewProcessInstanceManager creates new process instance manager
// Создает новый менеджер экземпляров процессов
func NewProcessInstanceManager(storage storage.Storage, component ComponentInterface) *ProcessInstanceManager {
	return &ProcessInstanceManager{
		storage:        storage,
		component:      component,
		processStarter: NewProcessStarter(storage, component),
	}
}

// Init initializes process instance manager
// Инициализирует менеджер экземпляров процессов
func (pim *ProcessInstanceManager) Init() error {
	logger.Info("Initializing process instance manager")
	return nil
}

// StartProcessInstance starts new process instance
// Запускает новый экземпляр процесса
func (pim *ProcessInstanceManager) StartProcessInstance(processKey string, variables map[string]interface{}) (*models.ProcessInstance, error) {
	return pim.processStarter.StartProcessInstance(processKey, variables)
}

// GetProcessInstanceStatus gets process instance status
// Получает статус экземпляра процесса
func (pim *ProcessInstanceManager) GetProcessInstanceStatus(instanceID string) (*models.ProcessInstance, error) {
	if !pim.component.IsReady() {
		return nil, fmt.Errorf("process component not ready")
	}

	return pim.storage.LoadProcessInstance(instanceID)
}

// CancelProcessInstance cancels process instance
// Отменяет экземпляр процесса
func (pim *ProcessInstanceManager) CancelProcessInstance(instanceID string, reason string) error {
	logger.Info("Canceling process instance",
		logger.String("instance_id", instanceID),
		logger.String("reason", reason))

	if !pim.component.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	// Load process instance
	instance, err := pim.storage.LoadProcessInstance(instanceID)
	if err != nil {
		return fmt.Errorf("failed to load process instance: %w", err)
	}

	// Set state to canceled
	instance.SetState(models.ProcessInstanceStateCanceled)
	instance.AddMetadata("cancel_reason", reason)

	// Update process instance
	if err := pim.storage.UpdateProcessInstance(instance); err != nil {
		return fmt.Errorf("failed to update process instance: %w", err)
	}

	// Cancel all active tokens
	tokens, err := pim.storage.LoadTokensByProcessInstance(instanceID)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	for _, token := range tokens {
		if token.IsActive() || token.IsWaiting() {
			// Cancel boundary timers before setting token state
			if err := pim.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
				logger.Error("Failed to cancel boundary timers for token",
					logger.String("token_id", token.TokenID),
					logger.String("error", err.Error()))
			}

			token.SetState(models.TokenStateCanceled)
			if err := pim.storage.UpdateToken(token); err != nil {
				logger.Error("Failed to cancel token",
					logger.String("token_id", token.TokenID),
					logger.String("error", err.Error()))
			}
		}
	}

	logger.Info("Process instance canceled", logger.String("instance_id", instanceID))
	return nil
}

// ListProcessInstances lists process instances with optional filters
// Получает список экземпляров процессов с опциональными фильтрами
func (pim *ProcessInstanceManager) ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*models.ProcessInstance, error) {
	if !pim.component.IsReady() {
		return nil, fmt.Errorf("process component not ready")
	}

	var instances []*models.ProcessInstance
	var err error

	// Load instances based on filters
	if processKeyFilter != "" {
		instances, err = pim.storage.LoadProcessInstancesByProcessKey(processKeyFilter)
	} else {
		instances, err = pim.storage.LoadAllProcessInstances()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load process instances: %w", err)
	}

	// Apply status filter if provided
	if statusFilter != "" {
		var filteredInstances []*models.ProcessInstance
		for _, instance := range instances {
			if string(instance.State) == statusFilter {
				filteredInstances = append(filteredInstances, instance)
			}
		}
		instances = filteredInstances
	}

	// Apply limit if provided
	if limit > 0 && len(instances) > limit {
		instances = instances[:limit]
	}

	return instances, nil
}

// RestoreActiveProcesses restores active processes after restart
// Восстанавливает активные процессы после перезапуска
func (pim *ProcessInstanceManager) RestoreActiveProcesses() error {
	logger.Info("Restoring active processes")

	// Load all active tokens
	activeTokens, err := pim.storage.LoadActiveTokens()
	if err != nil {
		return fmt.Errorf("failed to load active tokens: %w", err)
	}

	logger.Info("Found active tokens to restore", logger.Int("count", len(activeTokens)))

	// Group tokens by process instance
	instanceTokens := make(map[string][]*models.Token)
	for _, token := range activeTokens {
		instanceTokens[token.ProcessInstanceID] = append(instanceTokens[token.ProcessInstanceID], token)
	}

	// Continue execution for each process instance
	for instanceID, tokens := range instanceTokens {
		logger.Info("Restoring process instance",
			logger.String("instance_id", instanceID),
			logger.Int("token_count", len(tokens)))

		if err := pim.component.ContinueExecution(instanceID); err != nil {
			logger.Error("Failed to restore process instance",
				logger.String("instance_id", instanceID),
				logger.String("error", err.Error()))
		}
	}

	return nil
}
