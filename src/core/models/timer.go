/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"time"
)

// TimerType defines type of timer
// Определяет тип таймера
type TimerType string

const (
	TimerTypeStart    TimerType = "START"    // Start event timer
	TimerTypeBoundary TimerType = "BOUNDARY" // Boundary event timer
	TimerTypeEvent    TimerType = "EVENT"    // Intermediate timer event
)

// TimerState defines state of timer
// Определяет состояние таймера
type TimerState string

const (
	TimerStateScheduled TimerState = "SCHEDULED" // Timer is scheduled
	TimerStateFired     TimerState = "FIRED"     // Timer has fired
	TimerStateCanceled  TimerState = "CANCELED"  // Timer was canceled
)

// Timer represents a timer in the system
// Представляет таймер в системе
type Timer struct {
	ID                string                 `json:"id"`
	ElementID         string                 `json:"element_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ExecutionTokenID  string                 `json:"execution_token_id"`
	Type              TimerType              `json:"type"`
	State             TimerState             `json:"state"`
	DueDate           time.Time              `json:"due_date"`
	Variables         map[string]interface{} `json:"variables"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`

	// Process context for returning results
	// Контекст процесса для возврата результатов
	ProcessContext *TimerProcessContext `json:"process_context,omitempty"`
}

// TimerProcessContext contains process metadata for timer callbacks
// Содержит метаданные процесса для колбеков таймера
type TimerProcessContext struct {
	ProcessKey      string `json:"process_key"`      // BPMN process definition key
	ProcessVersion  int    `json:"process_version"`  // Process version
	ProcessName     string `json:"process_name"`     // Human readable name
	ComponentSource string `json:"component_source"` // Component that created timer
}
