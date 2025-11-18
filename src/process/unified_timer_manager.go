/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"context"
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
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

// CancelAllTimersForProcessInstance cancels all scheduled timers for process instance
// Отменяет все запланированные таймеры для экземпляра процесса
func (utm *UnifiedTimerManager) CancelAllTimersForProcessInstance(instanceID string) error {
	// Get core interface to access timewheel
	// Получаем core interface для доступа к timewheel
	core := utm.component.GetCore()
	if core == nil {
		return fmt.Errorf("core interface not available")
	}

	timewheelComp := core.GetTimewheelComponentInterface()
	if timewheelComp == nil {
		return fmt.Errorf("timewheel component not available")
	}

	// Type assertion to get timewheel component with ProcessMessage method
	// Приведение типа чтобы получить timewheel компонент с методом ProcessMessage
	type TimewheelComponent interface {
		ProcessMessage(ctx context.Context, messageJSON string) error
	}

	twComp, ok := timewheelComp.(TimewheelComponent)
	if !ok {
		return fmt.Errorf("timewheel component does not implement ProcessMessage")
	}

	// Load all timers from storage
	// Загружаем все таймеры из storage
	allTimers, err := utm.storage.LoadAllTimers()
	if err != nil {
		return fmt.Errorf("failed to load timers: %w", err)
	}

	// Filter timers by process instance ID and scheduled state
	// Фильтруем таймеры по ID экземпляра процесса и статусу SCHEDULED
	var timersToCancel []*storage.TimerRecord
	for _, timer := range allTimers {
		if timer.ProcessInstanceID == instanceID && timer.State == "SCHEDULED" {
			timersToCancel = append(timersToCancel, timer)
		}
	}

	if len(timersToCancel) == 0 {
		logger.Debug("No scheduled timers found for process instance",
			logger.String("instance_id", instanceID))
		return nil
	}

	logger.Info("Canceling process timers",
		logger.String("instance_id", instanceID),
		logger.Int("timer_count", len(timersToCancel)))

	// Cancel each timer with panic recovery
	// Отменяем каждый таймер с восстановлением от паник
	ctx := context.Background()
	for _, timer := range timersToCancel {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic occurred while canceling process timer",
						logger.String("instance_id", instanceID),
						logger.String("timer_id", timer.ID),
						logger.String("panic", fmt.Sprintf("%v", r)))
				}
			}()

			// Create cancel timer message
			// Создаем сообщение отмены таймера
			cancelMessage, err := timewheel.CreateCancelTimerMessage(timer.ID)
			if err != nil {
				logger.Error("Failed to create cancel timer message",
					logger.String("timer_id", timer.ID),
					logger.String("error", err.Error()))
				return
			}

			// Send cancel message to timewheel
			// Отправляем сообщение отмены в timewheel
			if err := twComp.ProcessMessage(ctx, cancelMessage); err != nil {
				logger.Error("Failed to cancel timer in timewheel",
					logger.String("timer_id", timer.ID),
					logger.String("error", err.Error()))
				// Continue to mark as canceled in storage even if timewheel fails
				// Продолжаем помечать как отмененный в storage даже если timewheel не удалось
			} else {
				logger.Debug("Timer canceled successfully",
					logger.String("timer_id", timer.ID),
					logger.String("instance_id", instanceID))
			}

			// Mark timer as canceled in storage
			// Помечаем таймер как отмененный в storage
			timer.State = "CANCELLED"
			if err := utm.storage.UpdateTimer(timer); err != nil {
				logger.Error("Failed to update timer state in storage",
					logger.String("timer_id", timer.ID),
					logger.String("error", err.Error()))
			}
		}()
	}

	logger.Info("Process timers cancellation completed",
		logger.String("instance_id", instanceID),
		logger.Int("canceled_count", len(timersToCancel)))

	return nil
}

// GetBPMNProcessForToken loads BPMN process for token
// Загружает BPMN процесс для токена
func (utm *UnifiedTimerManager) GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error) {
	return utm.bpmnHelper.GetBPMNProcessForToken(token.ProcessKey)
}
