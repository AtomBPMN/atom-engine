/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// UnifiedTimerManager implements TimerCallbackManagerInterface
// Объединенный менеджер таймеров
type UnifiedTimerManager struct {
	storage              storage.Storage
	component            ComponentInterface
	timerCallbacks       *TimerCallbacks
	boundaryTimerManager *BoundaryTimerManager
	bpmnHelper           *BPMNHelper
}

// NewUnifiedTimerManager creates new unified timer manager
// Создает новый объединенный менеджер таймеров
func NewUnifiedTimerManager(storage storage.Storage, component ComponentInterface) *UnifiedTimerManager {
	return &UnifiedTimerManager{
		storage:              storage,
		component:            component,
		timerCallbacks:       NewTimerCallbacks(storage, component),
		boundaryTimerManager: NewBoundaryTimerManager(storage, component),
		bpmnHelper:           NewBPMNHelper(storage),
	}
}

// SetCore sets core interface
// Устанавливает интерфейс core
func (utm *UnifiedTimerManager) SetCore(core CoreInterface) {
	utm.timerCallbacks.SetCore(core)
	utm.boundaryTimerManager.SetCore(core)
}

// Init initializes unified timer manager
// Инициализирует объединенный менеджер таймеров
func (utm *UnifiedTimerManager) Init() error {
	if err := utm.timerCallbacks.Init(); err != nil {
		return fmt.Errorf("failed to initialize timer callbacks: %w", err)
	}

	if err := utm.boundaryTimerManager.Init(); err != nil {
		return fmt.Errorf("failed to initialize boundary timer manager: %w", err)
	}

	return nil
}

// CreateTimer creates timer through timer callbacks
// Создает таймер через timer callbacks
func (utm *UnifiedTimerManager) CreateTimer(timerRequest *TimerRequest) error {
	return utm.timerCallbacks.CreateTimer(timerRequest)
}

// HandleTimerCallback handles timer callback by routing to appropriate manager
// Обрабатывает timer callback маршрутизируя в подходящий менеджер
func (utm *UnifiedTimerManager) HandleTimerCallback(timerID, elementID, tokenID string) error {
	// Load timer to determine type and route to appropriate handler
	timerRecord, err := utm.storage.LoadTimer(timerID)
	if err != nil {
		return fmt.Errorf("failed to load timer %s: %w", timerID, err)
	}

	// Route to appropriate handler based on timer type
	switch timerRecord.TimerType {
	case "BOUNDARY":
		return utm.boundaryTimerManager.HandleBoundaryTimerCallback(timerID, elementID, tokenID, timerRecord)
	case "EVENT":
		return utm.timerCallbacks.HandleTimerCallback(timerID, elementID, tokenID)
	default:
		return utm.timerCallbacks.HandleTimerCallback(timerID, elementID, tokenID)
	}
}

// CreateBoundaryTimer creates boundary timer
// Создает boundary таймер
func (utm *UnifiedTimerManager) CreateBoundaryTimer(timerRequest *TimerRequest) error {
	return utm.boundaryTimerManager.CreateBoundaryTimer(timerRequest)
}

// CreateBoundaryTimerWithID creates boundary timer with ID
// Создает boundary таймер с ID
func (utm *UnifiedTimerManager) CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error) {
	return utm.boundaryTimerManager.CreateBoundaryTimerWithID(timerRequest)
}

// LinkBoundaryTimerToToken links boundary timer to token
// Связывает boundary таймер с токеном
func (utm *UnifiedTimerManager) LinkBoundaryTimerToToken(tokenID, timerID string) error {
	return utm.boundaryTimerManager.LinkBoundaryTimerToToken(tokenID, timerID)
}

// CancelBoundaryTimersForToken cancels boundary timers for token
// Отменяет boundary таймеры для токена
func (utm *UnifiedTimerManager) CancelBoundaryTimersForToken(tokenID string) error {
	return utm.boundaryTimerManager.CancelBoundaryTimersForToken(tokenID)
}

// GetBPMNProcessForToken loads BPMN process for token
// Загружает BPMN процесс для токена
func (utm *UnifiedTimerManager) GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error) {
	return utm.bpmnHelper.GetBPMNProcessForToken(token.ProcessKey)
}
