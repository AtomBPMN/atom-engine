/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// ProcessStarter handles process instance creation and startup logic
// Обрабатывает создание экземпляров процессов и логику запуска
type ProcessStarter struct {
	storage              storage.Storage
	component            ComponentInterface
	bpmnHelper           *BPMNHelper
	collaborationManager *CollaborationManager
}

// NewProcessStarter creates new process starter
// Создает новый стартер процессов
func NewProcessStarter(storage storage.Storage, component ComponentInterface) *ProcessStarter {
	return &ProcessStarter{
		storage:              storage,
		component:            component,
		bpmnHelper:           NewBPMNHelper(storage),
		collaborationManager: NewCollaborationManager(storage, component),
	}
}

// StartProcessInstance starts new process instance
// Запускает новый экземпляр процесса
func (ps *ProcessStarter) StartProcessInstance(
	processKey string,
	variables map[string]interface{},
) (*models.ProcessInstance, error) {
	logger.Info("Starting process instance",
		logger.String("process_key", processKey))

	if !ps.component.IsReady() {
		return nil, fmt.Errorf("process component not ready")
	}

	// Parse process key to get process ID and version
	processID, version := ps.parseProcessKey(processKey)

	// Load process definition from storage
	processData, actualStorageKey, err := ps.loadProcessDefinition(processID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to load process definition: %w", err)
	}

	// Parse process definition
	bpmnProcess, err := ps.parseProcessDefinition(processData, actualStorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse process definition: %w", err)
	}

	// Create process instance
	instance := ps.createProcessInstance(bpmnProcess, actualStorageKey, variables)

	// Save to storage first (sets InstanceID)
	if err := ps.storage.SaveProcessInstance(instance); err != nil {
		return nil, fmt.Errorf("failed to save process instance: %w", err)
	}

	logger.Info("Process instance created",
		logger.String("instance_id", instance.InstanceID),
		logger.String("process_id", instance.ProcessID),
		logger.String("process_key", processKey),
		logger.String("state", string(instance.State)))

	// Start execution
	if err := ps.startExecution(instance, bpmnProcess, actualStorageKey, variables); err != nil {
		logger.Error("Failed to start process execution",
			logger.String("instance_id", instance.InstanceID),
			logger.String("error", err.Error()))
		return instance, fmt.Errorf("failed to start process execution: %w", err)
	}

	logger.Info("Process instance started successfully",
		logger.String("instance_id", instance.InstanceID),
		logger.String("process_key", processKey))

	return instance, nil
}

// parseProcessKey parses process key to extract process ID and version
// Парсит ключ процесса для извлечения ID процесса и версии
func (ps *ProcessStarter) parseProcessKey(processKey string) (string, int) {
	// Default version
	version := -1

	// Try to extract version from process key format "processID:version"
	if strings.Contains(processKey, ":") {
		parts := strings.Split(processKey, ":")
		if len(parts) == 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				version = v
			}
		}
		return parts[0], version
	}

	// No version specified, return process key as-is
	return processKey, version
}

// loadProcessDefinition loads process definition from storage
// Загружает определение процесса из storage
func (ps *ProcessStarter) loadProcessDefinition(processID string, version int) ([]byte, string, error) {
	logger.Info("Loading process for new instance",
		logger.String("process_id", processID),
		logger.Int("version", version))

	processData, actualStorageKey, err := ps.storage.LoadBPMNProcessByProcessID(processID, version)
	if err != nil {
		logger.Error("Failed to load process for new instance",
			logger.String("process_id", processID),
			logger.Int("version", version),
			logger.String("error", err.Error()))
		return nil, "", fmt.Errorf("failed to load process definition: %w", err)
	}

	return processData, actualStorageKey, nil
}

// parseProcessDefinition parses BPMN process definition
// Парсит определение BPMN процесса
func (ps *ProcessStarter) parseProcessDefinition(processData []byte, storageKey string) (*models.BPMNProcess, error) {
	var bpmnProcess models.BPMNProcess
	if err := json.Unmarshal(processData, &bpmnProcess); err != nil {
		logger.Error("Failed to parse process for new instance",
			logger.String("storage_key", storageKey),
			logger.String("parse_error", err.Error()),
			logger.String("raw_json", string(processData)))
		return nil, fmt.Errorf("failed to parse process definition: %w", err)
	}

	return &bpmnProcess, nil
}

// createProcessInstance creates new process instance
// Создает новый экземпляр процесса
func (ps *ProcessStarter) createProcessInstance(
	bpmnProcess *models.BPMNProcess,
	processKey string,
	variables map[string]interface{},
) *models.ProcessInstance {
	instance := models.NewProcessInstance(
		bpmnProcess.ProcessID,
		bpmnProcess.ProcessName,
		bpmnProcess.ProcessVersion,
		processKey,
	)

	// Set variables if provided
	if variables != nil {
		instance.SetVariables(variables)
	}

	return instance
}

// startExecution starts process execution
// Запускает выполнение процесса
func (ps *ProcessStarter) startExecution(
	instance *models.ProcessInstance,
	bpmnProcess *models.BPMNProcess,
	processKey string,
	variables map[string]interface{},
) error {
	// Find start event
	startEventID, err := ps.findStartEvent(bpmnProcess)
	if err != nil {
		return fmt.Errorf("failed to find start event: %w", err)
	}

	// Check if start event is Message Start Event
	isMessageStartEvent, err := ps.isMessageStartEvent(bpmnProcess, startEventID)
	if err != nil {
		return fmt.Errorf("failed to check start event type: %w", err)
	}

	if isMessageStartEvent {
		return ps.handleMessageStartEvent(instance, bpmnProcess, startEventID, processKey, variables)
	} else {
		return ps.handleRegularStartEvent(instance, processKey, startEventID)
	}
}
