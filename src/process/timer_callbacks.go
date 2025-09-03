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

// TimerCallbacks handles timer-related operations and callbacks
// Обрабатывает операции и callbacks связанные с таймерами
type TimerCallbacks struct {
	storage        storage.Storage
	component      ComponentInterface
	core           CoreInterface
	callbackHelper *CallbackHelper
}

// NewTimerCallbacks creates new timer callbacks handler
// Создает новый обработчик callbacks таймеров
func NewTimerCallbacks(storage storage.Storage, component ComponentInterface) *TimerCallbacks {
	return &TimerCallbacks{
		storage:        storage,
		component:      component,
		callbackHelper: NewCallbackHelper(storage, component),
	}
}

// SetCore sets core interface for timer management
// Устанавливает интерфейс core для управления таймерами
func (tc *TimerCallbacks) SetCore(core CoreInterface) {
	tc.core = core
}

// Init initializes timer callbacks handler
// Инициализирует обработчик callbacks таймеров
func (tc *TimerCallbacks) Init() error {
	logger.Info("Initializing timer callbacks handler")
	return nil
}

// CreateTimer creates timer for intermediate catch events
// Создает таймер для промежуточных событий ловли
func (tc *TimerCallbacks) CreateTimer(timerRequest *TimerRequest) error {
	if tc.core == nil {
		return fmt.Errorf("core interface not set")
	}

	timewheelComp := tc.core.GetTimewheelComponentInterface()
	if timewheelComp == nil {
		return fmt.Errorf("timewheel component not available")
	}

	// Create timewheel timer request
	twRequest := timewheel.TimerRequest{
		ElementID:         timerRequest.ElementID,
		TokenID:           timerRequest.TokenID,
		ProcessInstanceID: timerRequest.ProcessInstanceID,
		TimerType:         models.TimerTypeEvent,
		ProcessContext: &models.TimerProcessContext{
			ProcessKey:      timerRequest.ProcessKey,
			ProcessVersion:  1, // TODO: Extract from process key
			ProcessName:     "Process Timer",
			ComponentSource: "process",
		},
	}

	// Set timer definition
	if timerRequest.TimeDuration != nil {
		twRequest.TimeDuration = timerRequest.TimeDuration
	} else if timerRequest.TimeDate != nil {
		twRequest.TimeDate = timerRequest.TimeDate
	} else if timerRequest.TimeCycle != nil {
		twRequest.TimeCycle = timerRequest.TimeCycle
	} else {
		return fmt.Errorf("no timer definition provided")
	}

	// Create schedule timer message
	messageJSON, err := timewheel.CreateScheduleTimerMessage(twRequest)
	if err != nil {
		return fmt.Errorf("failed to create timer message: %w", err)
	}

	// Process timer message via timewheel component
	if processMsgMethod, ok := timewheelComp.(interface {
		ProcessMessage(context.Context, string) error
	}); ok {
		ctx := context.Background()
		if err := processMsgMethod.ProcessMessage(ctx, messageJSON); err != nil {
			return fmt.Errorf("failed to process timer message: %w", err)
		}
	} else {
		return fmt.Errorf("timewheel component does not support ProcessMessage")
	}

	logger.Info("Timer created successfully",
		logger.String("element_id", timerRequest.ElementID),
		logger.String("token_id", timerRequest.TokenID))

	return nil
}

// HandleTimerCallback handles timer callback from timewheel component
// Обрабатывает callback таймера от timewheel компонента
func (tc *TimerCallbacks) HandleTimerCallback(timerID, elementID, tokenID string) error {
	if !tc.component.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	logger.Info("Handling timer callback",
		logger.String("timer_id", timerID),
		logger.String("element_id", elementID),
		logger.String("token_id", tokenID))

	// Load timer from storage to get timer type
	// Загружаем таймер из storage чтобы получить тип таймера
	timerRecord, err := tc.storage.LoadTimer(timerID)
	if err != nil {
		logger.Error("Failed to load timer for callback",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to load timer %s: %w", timerID, err)
	}

	logger.Info("Timer callback type determined",
		logger.String("timer_id", timerID),
		logger.String("timer_type", timerRecord.TimerType))

	// Handle timer callback (only EVENT type timers should reach here via component routing)
	// Обрабатываем callback таймера (только EVENT таймеры должны попадать сюда через роутинг в component)
	return tc.handleEventTimerCallback(timerID, elementID, tokenID)
}

// handleEventTimerCallback handles intermediate catch event timer callbacks
// Обрабатывает callbacks таймеров промежуточных событий ловли
func (tc *TimerCallbacks) handleEventTimerCallback(timerID, elementID, tokenID string) error {
	// Load and validate token using helper
	expectedWaitingFor := fmt.Sprintf("timer:%s", elementID)
	token, err := tc.callbackHelper.LoadAndValidateToken(tokenID, expectedWaitingFor)
	if err != nil {
		return err
	}

	// Process callback and continue execution using helper
	return tc.callbackHelper.ProcessCallbackAndContinue(token, elementID, nil)
}
