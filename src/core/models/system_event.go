/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import "time"

// SystemEvent represents system lifecycle events
// Представляет события жизненного цикла системы
type SystemEvent struct {
	ID        int       `json:"id"`
	EventType string    `json:"event_type"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// System event types
// Типы системных событий
const (
	EventTypeStartup    = "startup"
	EventTypeShutdown   = "shutdown"
	EventTypeReady      = "ready"
	EventTypeBPMNParse  = "bpmn_parse"
	EventTypeBPMNDelete = "bpmn_delete"
	EventTypeError      = "error"
)

// System event statuses
// Статусы системных событий
const (
	StatusSuccess    = "success"
	StatusFailed     = "failed"
	StatusInProgress = "in_progress"
)

// NewSystemEvent creates new system event
// Создает новое системное событие
func NewSystemEvent(eventType, status, message string) *SystemEvent {
	return &SystemEvent{
		EventType: eventType,
		Status:    status,
		Message:   message,
		CreatedAt: time.Now(),
	}
}
