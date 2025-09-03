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
	"atom-engine/src/storage"
)

// RestoreTimers restores timers from storage
// Восстанавливает таймеры из storage
func (c *Component) RestoreTimers() error {
	if c.storage == nil {
		return nil // No storage configured, skip restore
	}

	timers, err := c.storage.LoadAllTimers()
	if err != nil {
		return fmt.Errorf("failed to load timers from storage: %w", err)
	}

	restoredCount := 0
	firedCount := 0

	for _, timerRecord := range timers {
		if timerRecord.State != "SCHEDULED" {
			continue // Skip non-scheduled timers
		}

		// Calculate DueDate from original timer definition
		// Вычисляем DueDate из оригинального определения таймера
		dueDate, err := c.calculateOriginalDueDate(timerRecord)
		if err != nil {
			continue // Skip invalid timer
		}

		// Check if timer is overdue
		// Проверяем просрочен ли таймер
		now := time.Now()
		if dueDate.Before(now) || dueDate.Equal(now) {
			// Timer is overdue - fire it immediately
			// Таймер просрочен - запускаем немедленно
			if err := c.fireOverdueTimer(timerRecord, dueDate); err == nil {
				firedCount++
			}
			continue
		}

		// Timer is still valid - restore it to timewheel with correct DueDate
		// Таймер еще валиден - восстанавливаем в timewheel с правильным DueDate
		timerReq := c.timerRecordToRequest(timerRecord)
		timerReq.RestoreDueDate = &dueDate // Set calculated DueDate for restoration
		scheduleMessage := struct {
			Type    string       `json:"type"`
			Request TimerRequest `json:"request"`
		}{
			Type:    "schedule_timer",
			Request: timerReq,
		}

		// Schedule timer directly via ProcessMessage
		// Планируем таймер напрямую через ProcessMessage
		reqJSON, err := json.Marshal(scheduleMessage)
		if err != nil {
			continue // Skip invalid timer
		}

		ctx := context.Background()
		if err := c.ProcessMessage(ctx, string(reqJSON)); err != nil {
			// Log error but continue with other timers
			// Логируем ошибку но продолжаем с другими таймерами
			continue
		}

		restoredCount++
	}

	// Log how many timers were restored and fired
	// Логируем сколько таймеров было восстановлено и запущено
	if restoredCount > 0 {
		fmt.Printf("Restored %d timer(s) from storage\n", restoredCount)
	}
	if firedCount > 0 {
		fmt.Printf("Fired %d overdue timer(s) immediately\n", firedCount)
	}

	return nil
}

// timerRecordToRequest converts storage.TimerRecord to TimerRequest for restoration
// Конвертирует storage.TimerRecord в TimerRequest для восстановления
func (c *Component) timerRecordToRequest(record *storage.TimerRecord) TimerRequest {
	// Convert ProcessContext back to models format
	processContext := &models.TimerProcessContext{}
	if record.ProcessContext != nil {
		if pk, ok := record.ProcessContext["process_key"].(string); ok {
			processContext.ProcessKey = pk
		}
		if pv, ok := record.ProcessContext["process_version"].(float64); ok {
			processContext.ProcessVersion = int(pv)
		}
		if pn, ok := record.ProcessContext["process_name"].(string); ok {
			processContext.ProcessName = pn
		}
		if cs, ok := record.ProcessContext["component_source"].(string); ok {
			processContext.ComponentSource = cs
		}
	}

	// For restoration, we need to use the existing timer ID and calculate the correct timing
	// Для восстановления нужно использовать существующий ID таймера и правильно рассчитать время
	req := TimerRequest{
		ElementID:         record.ElementID,
		TokenID:           record.TokenID,
		ProcessInstanceID: record.ProcessInstanceID,
		TimerType:         models.TimerType(record.TimerType),
		ProcessContext:    processContext,
		TimeDate:          record.TimeDate,
		TimeDuration:      record.TimeDuration,
		TimeCycle:         record.TimeCycle,
		RestoreTimerID:    &record.ID, // CRITICAL: Use existing timer ID for restoration
	}

	return req
}

// calculateOriginalDueDate calculates DueDate from timer record based on original schedule time
// Вычисляет DueDate из записи таймера на основе оригинального времени планирования
func (c *Component) calculateOriginalDueDate(record *storage.TimerRecord) (time.Time, error) {
	parser := NewISO8601DurationParser()

	// Use ScheduledAt as base time for calculation
	// Используем ScheduledAt как базовое время для расчета
	baseTime := record.ScheduledAt

	if record.TimeDate != nil {
		// Absolute date timer
		// Абсолютный таймер даты
		return parser.ParseDate(*record.TimeDate)
	} else if record.TimeDuration != nil {
		// Duration-based timer
		// Таймер на основе длительности
		duration, err := parser.ParseDuration(*record.TimeDuration)
		if err != nil {
			return time.Time{}, err
		}
		return baseTime.Add(duration), nil
	} else if record.TimeCycle != nil {
		// Cycle-based timer - get first execution time
		// Циклический таймер - получаем время первого выполнения
		_, interval, err := parser.ParseRepeatingInterval(*record.TimeCycle)
		if err != nil {
			return time.Time{}, err
		}
		return baseTime.Add(interval), nil
	}

	return time.Time{}, fmt.Errorf("no timer definition found")
}

// fireOverdueTimer fires an overdue timer immediately and updates storage
// Запускает просроченный таймер немедленно и обновляет storage
func (c *Component) fireOverdueTimer(record *storage.TimerRecord, originalDueDate time.Time) error {
	// Convert to Timer model for firing
	// Конвертируем в модель Timer для запуска
	timer := &models.Timer{
		ID:                record.ID,
		ElementID:         record.ElementID,
		ProcessInstanceID: record.ProcessInstanceID,
		ExecutionTokenID:  record.TokenID,
		Type:              models.TimerType(record.TimerType),
		State:             models.TimerStateFired,
		DueDate:           originalDueDate,
		Variables:         make(map[string]interface{}),
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         time.Now(),
	}

	// Convert ProcessContext back to models format
	// Конвертируем ProcessContext обратно в формат models
	if record.ProcessContext != nil {
		processContext := &models.TimerProcessContext{}
		if pk, ok := record.ProcessContext["process_key"].(string); ok {
			processContext.ProcessKey = pk
		}
		if pv, ok := record.ProcessContext["process_version"].(float64); ok {
			processContext.ProcessVersion = int(pv)
		}
		if pn, ok := record.ProcessContext["process_name"].(string); ok {
			processContext.ProcessName = pn
		}
		if cs, ok := record.ProcessContext["component_source"].(string); ok {
			processContext.ComponentSource = cs
		}
		timer.ProcessContext = processContext
	}

	// Create timer response for immediate firing
	// Создаем ответ таймера для немедленного запуска
	response := TimerResponse{
		TimerID:           timer.ID,
		ElementID:         timer.ElementID,
		TokenID:           timer.ExecutionTokenID,
		ProcessInstanceID: timer.ProcessInstanceID,
		TimerType:         timer.Type,
		ProcessContext:    timer.ProcessContext,
		FiredAt:           time.Now(),
		Variables:         timer.Variables,
	}

	// Send response via channel
	// Отправляем ответ через канал
	if c.responseChannel != nil {
		if jsonData, err := json.Marshal(response); err == nil {
			select {
			case c.responseChannel <- string(jsonData):
			default:
				// Channel full, timer will still be marked as fired
				// Канал заполнен, таймер все равно будет помечен как запущенный
			}
		}
	}

	// Update timer status in storage to FIRED
	// Обновляем статус таймера в storage на FIRED
	if c.storage != nil {
		updatedRecord := *record
		updatedRecord.State = "FIRED"
		updatedRecord.UpdatedAt = time.Now()
		return c.storage.SaveTimer(&updatedRecord)
	}

	return nil
}
