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

// TimerCallbackManagerInterface handles timer operations
// Интерфейс менеджера таймеров
type TimerCallbackManagerInterface interface {
	// Timer operations
	CreateTimer(timerRequest *TimerRequest) error
	HandleTimerCallback(timerID, elementID, tokenID string) error

	// Boundary timer operations
	CreateBoundaryTimer(timerRequest *TimerRequest) error
	CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error)
	LinkBoundaryTimerToToken(tokenID, timerID string) error
	CancelBoundaryTimersForToken(tokenID string) error

	// Process timer operations
	CancelAllTimersForProcessInstance(instanceID string) error

	// Helper operations
	GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error)
}
