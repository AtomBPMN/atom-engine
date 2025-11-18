/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"encoding/json"
	"time"
)

// ProcessInstanceState defines state of process instance
// Определяет состояние экземпляра процесса
type ProcessInstanceState string

const (
	ProcessInstanceStateActive    ProcessInstanceState = "ACTIVE"
	ProcessInstanceStateMessages  ProcessInstanceState = "MESSAGES"
	ProcessInstanceStateCompleted ProcessInstanceState = "COMPLETED"
	ProcessInstanceStateCanceled  ProcessInstanceState = "CANCELED"
	ProcessInstanceStateFailed    ProcessInstanceState = "FAILED"
	ProcessInstanceStateSuspended ProcessInstanceState = "SUSPENDED"
)

// ProcessInstance represents running instance of BPMN process
// Представляет выполняющийся экземпляр BPMN процесса
type ProcessInstance struct {
	InstanceID      string                 `json:"instance_id"`
	ProcessID       string                 `json:"process_id"`      // Process definition ID
	ProcessName     string                 `json:"process_name"`    // Human readable name
	ProcessVersion  int                    `json:"process_version"` // Version of process definition
	ProcessKey      string                 `json:"process_key"`     // Unique process key (BPMN ID)
	State           ProcessInstanceState   `json:"state"`
	Variables       map[string]interface{} `json:"variables"`        // Process variables
	CurrentActivity string                 `json:"current_activity"` // Current active element ID
	StartedAt       time.Time              `json:"started_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`

	// Metadata for process execution
	// Метаданные для выполнения процесса
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewProcessInstance creates new process instance
// Создает новый экземпляр процесса
func NewProcessInstance(processID, processName string, processVersion int, processKey string) *ProcessInstance {
	now := time.Now()
	return &ProcessInstance{
		InstanceID:     GenerateID(),
		ProcessID:      processID,
		ProcessName:    processName,
		ProcessVersion: processVersion,
		ProcessKey:     processKey,
		State:          ProcessInstanceStateActive,
		Variables:      make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		StartedAt:      now,
		UpdatedAt:      now,
	}
}

// ToJSON converts process instance to JSON
// Конвертирует экземпляр процесса в JSON
func (pi *ProcessInstance) ToJSON() ([]byte, error) {
	return json.Marshal(pi)
}

// FromJSON creates process instance from JSON
// Создает экземпляр процесса из JSON
func (pi *ProcessInstance) FromJSON(data []byte) error {
	return json.Unmarshal(data, pi)
}

// SetVariable sets process variable
// Устанавливает переменную процесса
func (pi *ProcessInstance) SetVariable(key string, value interface{}) {
	if pi.Variables == nil {
		pi.Variables = make(map[string]interface{})
	}
	pi.Variables[key] = value
	pi.UpdatedAt = time.Now()
}

// GetVariable gets process variable
// Получает переменную процесса
func (pi *ProcessInstance) GetVariable(key string) (interface{}, bool) {
	value, exists := pi.Variables[key]
	return value, exists
}

// SetVariables sets multiple process variables
// Устанавливает множественные переменные процесса
func (pi *ProcessInstance) SetVariables(variables map[string]interface{}) {
	if pi.Variables == nil {
		pi.Variables = make(map[string]interface{})
	}
	for key, value := range variables {
		pi.Variables[key] = value
	}
	pi.UpdatedAt = time.Now()
}

// SetCurrentActivity sets current active element
// Устанавливает текущий активный элемент
func (pi *ProcessInstance) SetCurrentActivity(elementID string) {
	pi.CurrentActivity = elementID
	pi.UpdatedAt = time.Now()
}

// SetState sets process instance state
// Устанавливает состояние экземпляра процесса
func (pi *ProcessInstance) SetState(state ProcessInstanceState) {
	pi.State = state
	pi.UpdatedAt = time.Now()

	if state == ProcessInstanceStateCompleted ||
		state == ProcessInstanceStateCanceled ||
		state == ProcessInstanceStateFailed {
		now := time.Now()
		pi.CompletedAt = &now
	}
}

// AddMetadata adds metadata field
// Добавляет поле метаданных
func (pi *ProcessInstance) AddMetadata(key string, value interface{}) {
	if pi.Metadata == nil {
		pi.Metadata = make(map[string]interface{})
	}
	pi.Metadata[key] = value
	pi.UpdatedAt = time.Now()
}

// GetMetadata gets metadata field
// Получает поле метаданных
func (pi *ProcessInstance) GetMetadata(key string) (interface{}, bool) {
	value, exists := pi.Metadata[key]
	return value, exists
}

// IsActive checks if process instance is active
// Проверяет активен ли экземпляр процесса
func (pi *ProcessInstance) IsActive() bool {
	return pi.State == ProcessInstanceStateActive
}

// IsWaitingForMessage checks if process instance is waiting for message
// Проверяет ожидает ли экземпляр процесса сообщение
func (pi *ProcessInstance) IsWaitingForMessage() bool {
	return pi.State == ProcessInstanceStateMessages
}

// IsCompleted checks if process instance is completed
// Проверяет завершен ли экземпляр процесса
func (pi *ProcessInstance) IsCompleted() bool {
	return pi.State == ProcessInstanceStateCompleted ||
		pi.State == ProcessInstanceStateCanceled ||
		pi.State == ProcessInstanceStateFailed
}
