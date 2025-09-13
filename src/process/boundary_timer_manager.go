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
	"strings"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
)

// BoundaryTimerManager manages boundary timer operations
// Управляет операциями boundary таймеров
type BoundaryTimerManager struct {
	storage    storage.Storage
	component  ComponentInterface
	core       CoreInterface
	bpmnHelper *BPMNHelper
}

// NewBoundaryTimerManager creates new boundary timer manager
// Создает новый менеджер boundary таймеров
func NewBoundaryTimerManager(storage storage.Storage, component ComponentInterface) *BoundaryTimerManager {
	return &BoundaryTimerManager{
		storage:    storage,
		component:  component,
		bpmnHelper: NewBPMNHelper(storage),
	}
}

// SetCore sets core interface for timer management
// Устанавливает интерфейс core для управления таймерами
func (btm *BoundaryTimerManager) SetCore(core CoreInterface) {
	btm.core = core
}

// Init initializes boundary timer manager
// Инициализирует менеджер boundary таймеров
func (btm *BoundaryTimerManager) Init() error {
	logger.Info("Initializing boundary timer manager")
	return nil
}

// CreateBoundaryTimer creates boundary timer with BOUNDARY timer type
// Создает boundary таймер с типом BOUNDARY
func (btm *BoundaryTimerManager) CreateBoundaryTimer(timerRequest *TimerRequest) error {
	if btm.core == nil {
		return fmt.Errorf("core interface not set")
	}

	timewheelComp := btm.core.GetTimewheelComponentInterface()
	if timewheelComp == nil {
		return fmt.Errorf("timewheel component not available")
	}

	// Get process version from ProcessInstanceID
	processVersion := 1 // Default fallback
	if btm.storage != nil {
		if instance, err := btm.storage.LoadProcessInstance(timerRequest.ProcessInstanceID); err == nil && instance != nil {
			processVersion = instance.ProcessVersion
		}
	}

	// Create timewheel timer request for boundary timer
	twRequest := timewheel.TimerRequest{
		ElementID:         timerRequest.ElementID,
		TokenID:           timerRequest.TokenID, // Parent token ID for boundary context
		ProcessInstanceID: timerRequest.ProcessInstanceID,
		TimerType:         models.TimerTypeBoundary, // Boundary timer type
		ProcessContext: &models.TimerProcessContext{
			ProcessKey:      timerRequest.ProcessKey,
			ProcessVersion:  processVersion, // Use actual version from process instance
			ProcessName:     "Boundary Timer",
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
		return fmt.Errorf("failed to create boundary timer message: %w", err)
	}

	// Process timer message via timewheel component
	if processMsgMethod, ok := timewheelComp.(interface {
		ProcessMessage(context.Context, string) error
	}); ok {
		ctx := context.Background()
		if err := processMsgMethod.ProcessMessage(ctx, messageJSON); err != nil {
			return fmt.Errorf("failed to process boundary timer message: %w", err)
		}
	} else {
		return fmt.Errorf("timewheel component does not support ProcessMessage")
	}

	logger.Info("Boundary timer created successfully",
		logger.String("element_id", timerRequest.ElementID),
		logger.String("parent_token_id", timerRequest.TokenID),
		logger.String("timer_type", "BOUNDARY"))

	return nil
}

// CreateBoundaryTimerWithID creates boundary timer and returns timer ID
// Создает boundary таймер и возвращает ID таймера
func (btm *BoundaryTimerManager) CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error) {
	if err := btm.CreateBoundaryTimer(timerRequest); err != nil {
		return "", err
	}

	// Find the real timer ID from storage that was just created
	// Находим реальный timer ID из storage который только что был создан
	allTimers, err := btm.storage.LoadAllTimers()
	if err != nil {
		logger.Error("Failed to load timers to find boundary timer ID",
			logger.String("element_id", timerRequest.ElementID),
			logger.String("parent_token_id", timerRequest.TokenID),
			logger.String("error", err.Error()))
		return "", fmt.Errorf("failed to load timers: %w", err)
	}

	// Find the timer we just created for this token and element
	// Находим таймер который мы только что создали для данного токена и элемента
	for _, timerRecord := range allTimers {
		if timerRecord.TimerType == "BOUNDARY" &&
			timerRecord.TokenID == timerRequest.TokenID &&
			timerRecord.ElementID == timerRequest.ElementID &&
			timerRecord.State == "SCHEDULED" {

			logger.Info("Real boundary timer ID found in storage",
				logger.String("element_id", timerRequest.ElementID),
				logger.String("parent_token_id", timerRequest.TokenID),
				logger.String("timer_id", timerRecord.ID))

			return timerRecord.ID, nil
		}
	}

	// Fallback: timer not found in storage, generate ID but log warning
	// Fallback: таймер не найден в storage, генерируем ID но логируем предупреждение
	timerID := models.GenerateID()
	logger.Warn("Boundary timer not found in storage, using generated ID",
		logger.String("element_id", timerRequest.ElementID),
		logger.String("parent_token_id", timerRequest.TokenID),
		logger.String("timer_id", timerID))

	return timerID, nil
}

// LinkBoundaryTimerToToken links boundary timer to parent token
// Связывает boundary таймер с родительским токеном
func (btm *BoundaryTimerManager) LinkBoundaryTimerToToken(tokenID, timerID string) error {
	// Load parent token
	// Загружаем родительский токен
	token, err := btm.storage.LoadToken(tokenID)
	if err != nil {
		return fmt.Errorf("failed to load token: %w", err)
	}

	// Add boundary timer ID to token
	// Добавляем ID boundary таймера к токену
	token.AddBoundaryTimer(timerID)

	// Save updated token
	// Сохраняем обновленный токен
	if err := btm.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	logger.Info("Boundary timer linked to token",
		logger.String("token_id", tokenID),
		logger.String("timer_id", timerID))

	return nil
}

// CancelBoundaryTimersForToken cancels all boundary timers for token
// Отменяет все boundary таймеры для токена
func (btm *BoundaryTimerManager) CancelBoundaryTimersForToken(tokenID string) error {
	// Load all timers and find boundary timers for this token
	// Загружаем все таймеры и находим boundary таймеры для данного токена
	allTimers, err := btm.storage.LoadAllTimers()
	if err != nil {
		return fmt.Errorf("failed to load timers: %w", err)
	}

	// Find boundary timers for this token
	// Находим boundary таймеры для данного токена
	var boundaryTimers []string
	for _, timerRecord := range allTimers {
		if timerRecord.TimerType == "BOUNDARY" && timerRecord.TokenID == tokenID && timerRecord.State == "SCHEDULED" {
			boundaryTimers = append(boundaryTimers, timerRecord.ID)
		}
	}

	if len(boundaryTimers) == 0 {
		return nil // No boundary timers to cancel
	}

	logger.Info("Canceling boundary timers for token",
		logger.String("token_id", tokenID),
		logger.Int("timer_count", len(boundaryTimers)),
		logger.String("boundary_timer_ids", fmt.Sprintf("%v", boundaryTimers)))

	// Cancel each boundary timer with panic recovery
	// Отменяем каждый boundary таймер с восстановлением от паник
	for _, timerID := range boundaryTimers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic occurred while canceling boundary timer",
						logger.String("token_id", tokenID),
						logger.String("timer_id", timerID),
						logger.String("panic", fmt.Sprintf("%v", r)))
				}
			}()

			if err := btm.cancelTimer(timerID); err != nil {
				logger.Error("Failed to cancel boundary timer",
					logger.String("token_id", tokenID),
					logger.String("timer_id", timerID),
					logger.String("error", err.Error()))
				// Continue with other timers
			} else {
				logger.Info("Boundary timer canceled",
					logger.String("token_id", tokenID),
					logger.String("timer_id", timerID))
			}
		}()
	}

	// Update token to clear boundary timer IDs (if needed)
	// Обновляем токен чтобы очистить ID boundary таймеров (если нужно)
	token, err := btm.storage.LoadToken(tokenID)
	if err == nil && token.HasBoundaryTimers() {
		token.BoundaryTimerIDs = make([]string, 0)
		if err := btm.storage.UpdateToken(token); err != nil {
			logger.Error("Failed to update token after canceling boundary timers",
				logger.String("token_id", tokenID),
				logger.String("error", err.Error()))
		}
	}

	return nil
}

// HandleBoundaryTimerCallback handles boundary timer callbacks
// Обрабатывает callbacks boundary таймеров
func (btm *BoundaryTimerManager) HandleBoundaryTimerCallback(timerID, elementID, tokenID string, timerRecord interface{}) error {
	logger.Info("Processing boundary timer callback",
		logger.String("timer_id", timerID),
		logger.String("boundary_event_id", elementID),
		logger.String("parent_token_id", tokenID))

	// Load parent token to check if it's still active in the activity
	// Загружаем родительский токен чтобы проверить активен ли он в активности
	parentToken, err := btm.storage.LoadToken(tokenID)
	if err != nil {
		logger.Error("Failed to load parent token for boundary timer",
			logger.String("parent_token_id", tokenID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to load parent token %s: %w", tokenID, err)
	}

	// Check if parent token is still active
	// Проверяем активен ли еще родительский токен
	if parentToken.IsCompleted() {
		logger.Info("Parent token is no longer active - canceling boundary timer",
			logger.String("parent_token_id", tokenID),
			logger.String("parent_token_state", string(parentToken.State)),
			logger.String("boundary_event_id", elementID),
			logger.String("timer_id", timerID))

		// Cancel the timer since parent activity is completed
		// Отменяем таймер поскольку родительская activity завершена
		if err := btm.cancelTimer(timerID); err != nil {
			logger.Error("Failed to cancel orphaned boundary timer",
				logger.String("timer_id", timerID),
				logger.String("parent_token_id", tokenID),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Orphaned boundary timer canceled successfully",
				logger.String("timer_id", timerID),
				logger.String("parent_token_id", tokenID))
		}

		return nil // Parent scope ended, boundary timer no longer relevant
	}

	// Load BPMN process elements using helper
	// Загружаем элементы BPMN процесса с помощью helper
	elements, err := btm.bpmnHelper.LoadProcessElements(parentToken.ProcessKey)
	if err != nil {
		return fmt.Errorf("failed to load process elements: %w", err)
	}

	boundaryEvent, exists := elements[elementID].(map[string]interface{})
	if !exists {
		return fmt.Errorf("boundary event %s not found in process", elementID)
	}

	// Check if this is non-interrupting boundary event
	// Проверяем является ли это non-interrupting boundary событием
	cancelActivity := true // default is interrupting
	if cancelActivityValue, exists := boundaryEvent["cancel_activity"]; exists {
		if cancelActivityBool, ok := cancelActivityValue.(bool); ok {
			cancelActivity = cancelActivityBool
		}
	}

	logger.Info("Boundary event properties determined",
		logger.String("boundary_event_id", elementID),
		logger.String("parent_token_id", tokenID),
		logger.Bool("cancel_activity", cancelActivity))

	if cancelActivity {
		// Interrupting boundary event - interrupt parent token and move it to boundary event
		// Прерывающее boundary событие - прерываем родительский токен и перемещаем на boundary событие
		logger.Info("Processing interrupting boundary event",
			logger.String("boundary_event_id", elementID),
			logger.String("parent_token_id", tokenID))

		// Cancel any jobs the parent token is waiting for
		// Отменяем любые jobs которые ждет родительский токен
		if parentToken.IsWaiting() && strings.HasPrefix(parentToken.WaitingFor, "job:") {
			jobID := strings.TrimPrefix(parentToken.WaitingFor, "job:")
			logger.Info("Canceling job for interrupted token",
				logger.String("token_id", tokenID),
				logger.String("job_id", jobID),
				logger.String("boundary_event_id", elementID))

			// Cancel job directly by ID - more efficient than CancelJobForToken
			// Отменяем job напрямую по ID - эффективнее чем CancelJobForToken
			if err := btm.component.CancelJobByID(jobID); err != nil {
				logger.Error("Failed to cancel job for interrupted token",
					logger.String("token_id", tokenID),
					logger.String("job_id", jobID),
					logger.String("error", err.Error()))
				// Continue execution even if job cancellation fails
			} else {
				logger.Info("Job canceled successfully for interrupted token",
					logger.String("token_id", tokenID),
					logger.String("job_id", jobID))
			}

			// Clear waiting state
			// Очищаем состояние ожидания
			parentToken.ClearWaitingFor()
		}

		// Move parent token to boundary event
		parentToken.MoveTo(elementID)
		if err := btm.storage.UpdateToken(parentToken); err != nil {
			return fmt.Errorf("failed to update parent token: %w", err)
		}

		logger.Info("Parent token interrupted and moved to boundary event",
			logger.String("token_id", tokenID),
			logger.String("boundary_event_id", elementID))

		// Execute boundary event with parent token
		return btm.component.ExecuteToken(parentToken)
	} else {
		// Non-interrupting boundary event - create new token on boundary event, keep parent active
		// Non-interrupting boundary событие - создаем новый токен на boundary событии, оставляем родительский активным
		logger.Info("Processing non-interrupting boundary event",
			logger.String("boundary_event_id", elementID),
			logger.String("parent_token_id", tokenID))

		// Create new token for boundary event
		// Создаем новый токен для boundary события
		boundaryToken := models.NewToken(parentToken.ProcessInstanceID, parentToken.ProcessKey, elementID)
		boundaryToken.SetVariables(parentToken.Variables) // Copy parent variables

		// Save boundary token
		if err := btm.storage.SaveToken(boundaryToken); err != nil {
			return fmt.Errorf("failed to save boundary token: %w", err)
		}

		logger.Info("Non-interrupting boundary token created",
			logger.String("boundary_token_id", boundaryToken.TokenID),
			logger.String("boundary_event_id", elementID),
			logger.String("parent_token_id", tokenID))

		// Execute boundary event with new token
		return btm.component.ExecuteToken(boundaryToken)
	}
}

// cancelTimer cancels specific timer via timewheel
// Отменяет конкретный таймер через timewheel
func (btm *BoundaryTimerManager) cancelTimer(timerID string) error {
	if btm.core == nil {
		return fmt.Errorf("core interface not set")
	}

	timewheelComp := btm.core.GetTimewheelComponentInterface()
	if timewheelComp == nil {
		return fmt.Errorf("timewheel component not available")
	}

	// Type assertion to get timewheel component with ProcessMessage method
	// Приведение типа чтобы получить timewheel компонент с методом ProcessMessage
	type TimewheelComponent interface {
		ProcessMessage(ctx context.Context, messageJSON string) error
	}

	logger.Info("Attempting to cast timewheel component for ProcessMessage",
		logger.String("timer_id", timerID))

	twComp, ok := timewheelComp.(TimewheelComponent)
	if !ok {
		logger.Error("Timewheel component does not implement ProcessMessage interface",
			logger.String("timer_id", timerID),
			logger.String("component_type", fmt.Sprintf("%T", timewheelComp)))
		return fmt.Errorf("timewheel component does not implement ProcessMessage")
	}

	// Create cancel timer message using timewheel helper
	// Создаем сообщение отмены таймера используя timewheel helper
	cancelMessage, err := timewheel.CreateCancelTimerMessage(timerID)
	if err != nil {
		logger.Error("Failed to create cancel timer message",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to create cancel timer message: %w", err)
	}

	logger.Info("Sending cancel message to timewheel",
		logger.String("timer_id", timerID),
		logger.String("cancel_message", cancelMessage))

	// Send cancel message to timewheel - this removes timer from wheel
	// Отправляем сообщение отмены в timewheel - это удаляет таймер из колеса
	ctx := context.Background()
	if err := twComp.ProcessMessage(ctx, cancelMessage); err != nil {
		logger.Error("Failed to cancel timer in timewheel",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		// Continue to mark as canceled in storage even if timewheel fails
		// Продолжаем помечать как отмененный в storage даже если timewheel не удалось
	} else {
		logger.Info("Timer successfully canceled in timewheel",
			logger.String("timer_id", timerID))
	}

	// Load timer and mark as canceled in storage
	// Загружаем таймер и отмечаем как отмененный в storage
	if timer, err := btm.storage.LoadTimer(timerID); err == nil {
		timer.State = "CANCELED"
		if err := btm.storage.SaveTimer(timer); err != nil {
			return fmt.Errorf("failed to save canceled timer: %w", err)
		}
	}

	return nil
}
