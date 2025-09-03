/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/src/core/models"
)

// ProcessJSONRequest processes JSON timer request
// Обрабатывает JSON запрос таймера
func (m *Manager) ProcessJSONRequest(jsonStr string) (string, error) {
	var req TimerRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return "", fmt.Errorf("failed to parse JSON request: %w", err)
	}

	return m.ScheduleTimerRequest(context.Background(), req)
}

// ScheduleTimerRequest schedules timer from request
// Планирует таймер из запроса
func (m *Manager) ScheduleTimerRequest(ctx context.Context, req TimerRequest) (string, error) {
	// Validate request
	// Валидируем запрос
	if err := m.validateRequest(req); err != nil {
		return "", err
	}

	// Create timer model
	// Создаем модель таймера
	var timerID string
	if req.RestoreTimerID != nil && *req.RestoreTimerID != "" {
		// Use existing ID for restoration
		// Используем существующий ID для восстановления
		timerID = *req.RestoreTimerID
	} else {
		// Generate new ID for new timer
		// Генерируем новый ID для нового таймера
		timerID = models.GenerateID()
	}

	timer := &models.Timer{
		ID:                timerID,
		ElementID:         req.ElementID,
		ProcessInstanceID: req.ProcessInstanceID,
		ExecutionTokenID:  req.TokenID,
		Type:              req.TimerType,
		State:             models.TimerStateScheduled,
		Variables:         make(map[string]interface{}),
		ProcessContext:    req.ProcessContext,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Process timer definition
	// Обрабатываем определение таймера
	var err error
	if req.RestoreDueDate != nil {
		// Use provided DueDate for restoration - don't recalculate
		// Используем предоставленный DueDate для восстановления - не пересчитываем
		timer.DueDate = *req.RestoreDueDate
	} else if req.TimeDate != nil {
		err = m.processTimeDate(timer, *req.TimeDate)
	} else if req.TimeDuration != nil {
		err = m.processTimeDuration(timer, *req.TimeDuration, req.BaseTime)
	} else if req.TimeCycle != nil {
		err = m.processTimeCycle(timer, *req.TimeCycle, req.BaseTime)
	}

	if err != nil {
		return "", fmt.Errorf("failed to process timer definition: %w", err)
	}

	// Add boundary timer metadata
	// Добавляем метаданные boundary таймера
	if req.TimerType == models.TimerTypeBoundary {
		m.addBoundaryTimerMetadata(timer, req)
	}

	// Add timer to wheel
	// Добавляем таймер в колесо
	handler := TimerHandlerFunc(m.handleTimerFired)
	if err := m.wheel.AddTimer(timer, handler); err != nil {
		return "", fmt.Errorf("failed to add timer to wheel: %w", err)
	}

	return timer.ID, nil
}

// validateRequest validates timer request
// Валидирует запрос таймера
func (m *Manager) validateRequest(req TimerRequest) error {
	if req.ElementID == "" {
		return ErrInvalidTimerRequest("element_id is required")
	}
	if req.TokenID == "" {
		return ErrInvalidTimerRequest("token_id is required")
	}
	if req.ProcessInstanceID == "" {
		return ErrInvalidTimerRequest("process_instance_id is required")
	}

	// Check that exactly one timer definition is provided
	// Проверяем что предоставлено ровно одно определение таймера
	definitionCount := 0
	if req.TimeDate != nil && *req.TimeDate != "" {
		definitionCount++
	}
	if req.TimeDuration != nil && *req.TimeDuration != "" {
		definitionCount++
	}
	if req.TimeCycle != nil && *req.TimeCycle != "" {
		definitionCount++
	}

	if definitionCount == 0 {
		return ErrInvalidTimerRequest("one timer definition must be provided")
	}
	if definitionCount > 1 {
		return ErrInvalidTimerRequest("only one timer definition can be provided")
	}

	return nil
}

// processTimeDate processes date-based timer
// Обрабатывает таймер на основе даты
func (m *Manager) processTimeDate(timer *models.Timer, dateStr string) error {
	dueDate, err := m.parser.ParseDate(dateStr)
	if err != nil {
		return err
	}

	timer.DueDate = dueDate
	timer.Variables["time_date"] = dateStr
	return nil
}

// processTimeDuration processes duration-based timer
// Обрабатывает таймер на основе длительности
func (m *Manager) processTimeDuration(timer *models.Timer, durationStr string, baseTime *time.Time) error {
	duration, err := m.parser.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	// Use provided base time or fallback to current time
	// Используем предоставленное базовое время или возвращаемся к текущему времени
	var startTime time.Time
	if baseTime != nil {
		startTime = *baseTime
	} else {
		startTime = time.Now()
	}

	timer.DueDate = startTime.Add(duration)
	timer.Variables["time_duration"] = durationStr
	return nil
}

// processTimeCycle processes cycle-based timer
// Обрабатывает циклический таймер
func (m *Manager) processTimeCycle(timer *models.Timer, cycleStr string, baseTime *time.Time) error {
	repeatCount, interval, err := m.parser.ParseRepeatingInterval(cycleStr)
	if err != nil {
		return err
	}

	// Use provided base time or fallback to current time
	// Используем предоставленное базовое время или возвращаемся к текущему времени
	var startTime time.Time
	if baseTime != nil {
		startTime = *baseTime
	} else {
		startTime = time.Now()
	}

	// For first execution
	// Для первого выполнения
	timer.DueDate = startTime.Add(interval)
	timer.Variables["time_cycle"] = cycleStr
	timer.Variables["repeat_count"] = repeatCount
	timer.Variables["interval"] = interval.String()
	timer.Variables["current_iteration"] = 1

	return nil
}

// addBoundaryTimerMetadata adds boundary timer specific metadata
// Добавляет специфичные метаданные boundary таймера
func (m *Manager) addBoundaryTimerMetadata(timer *models.Timer, req TimerRequest) {
	if req.AttachedToRef != nil {
		timer.Variables["attached_to_ref"] = *req.AttachedToRef
	}
	if req.CancelActivity != nil {
		timer.Variables["cancel_activity"] = *req.CancelActivity
	}
}
