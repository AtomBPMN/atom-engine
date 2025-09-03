/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"context"
	"fmt"
	"time"

	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// handleTimerFired handles timer when it fires
// Обрабатывает таймер когда он срабатывает
func (m *Manager) handleTimerFired(ctx context.Context, timer *models.Timer) error {
	// Timer response is already sent by the wheel via JSON channel
	// Ответ таймера уже отправлен колесом через JSON канал

	// Debug: log all timer variables for boundary timers
	// Отладка: логируем все переменные таймера для boundary таймеров
	if timer.Type == models.TimerTypeBoundary {
		fmt.Printf("BOUNDARY TIMER FIRED: %s, Variables: %+v\n", timer.ID, timer.Variables)
	}

	// Update timer status in storage to FIRED
	// Обновляем статус таймера в storage на FIRED
	if m.storage != nil {
		// Load existing timer record to preserve all fields
		// Загружаем существующую запись таймера чтобы сохранить все поля
		existingRecord, err := m.storage.LoadTimer(timer.ID)
		if err != nil {
			fmt.Printf("Failed to load existing timer for update: %v\n", err)
			return nil
		}

		// Update only the state and timestamp, preserve everything else
		// Обновляем только статус и timestamp, сохраняем все остальное
		existingRecord.State = "FIRED"
		existingRecord.UpdatedAt = time.Now()

		err = m.storage.SaveTimer(existingRecord)
		if err != nil {
			// Log error but don't fail the timer processing
			// Логируем ошибку но не прерываем обработку таймера
			fmt.Printf("Failed to update timer status in storage: %v\n", err)
		}
	}

	// Handle cycle timers for rescheduling
	// Обрабатываем циклические таймеры для переplanирования
	if cycleStr, ok := timer.Variables["time_cycle"].(string); ok {
		fmt.Printf("Found time_cycle for timer %s: %s, handling cycle\n", timer.ID, cycleStr)
		return m.handleCycleTimer(timer, cycleStr)
	} else {
		if timer.Type == models.TimerTypeBoundary {
			fmt.Printf("No time_cycle found for boundary timer %s\n", timer.ID)
		}
	}

	return nil
}

// handleCycleTimer handles cycle timer rescheduling
// Обрабатывает переplanирование циклического таймера
func (m *Manager) handleCycleTimer(timer *models.Timer, cycleStr string) error {
	repeatCount, ok := timer.Variables["repeat_count"].(int)
	if !ok {
		return nil // Not a cycle timer
	}

	currentIteration, ok := timer.Variables["current_iteration"].(int)
	if !ok {
		currentIteration = 1
	}

	// Check if we need to reschedule
	// Проверяем нужно ли переplanировать
	if repeatCount == -1 || currentIteration < repeatCount {
		// For BOUNDARY timers, check if parent scope is still active
		// Для BOUNDARY таймеров проверяем активен ли еще родительский scope
		if timer.Type == models.TimerTypeBoundary {
			if !m.isParentScopeActive(timer) {
				fmt.Printf("Parent scope ended for boundary timer %s - canceling repeats\n", timer.ID)
				return nil // Parent scope ended, don't create more iterations
			}
		}

		// Parse interval
		// Парсим интервал
		_, interval, err := m.parser.ParseRepeatingInterval(cycleStr)
		if err != nil {
			return err
		}

		// Create new timer for next iteration
		// Создаем новый таймер для следующей итерации
		nextTimer := *timer
		nextTimer.ID = models.GenerateID()
		nextTimer.DueDate = time.Now().Add(interval)
		nextTimer.State = models.TimerStateScheduled
		nextTimer.CreatedAt = time.Now()
		nextTimer.UpdatedAt = time.Now()
		nextTimer.Variables["current_iteration"] = currentIteration + 1

		// Clear anchor from previous timer
		// Очищаем якорь от предыдущего таймера
		delete(nextTimer.Variables, "_anchor")

		// Save next timer to storage before scheduling
		// Сохраняем следующий таймер в storage перед планированием
		if m.storage != nil {
			timerRecord := &storage.TimerRecord{
				ID:                nextTimer.ID,
				ElementID:         nextTimer.ElementID,
				ProcessInstanceID: nextTimer.ProcessInstanceID,
				TokenID:           nextTimer.ExecutionTokenID,
				TimerType:         string(nextTimer.Type),
				State:             string(nextTimer.State),
				ScheduledAt:       nextTimer.CreatedAt,
				CreatedAt:         nextTimer.CreatedAt,
				UpdatedAt:         nextTimer.UpdatedAt,
				Variables:         nextTimer.Variables,
			}

			// Set timer definition
			if cycle, exists := nextTimer.Variables["time_cycle"]; exists {
				if cycleStr, ok := cycle.(string); ok {
					timerRecord.TimeCycle = &cycleStr
				}
			}

			// Set process context
			if nextTimer.ProcessContext != nil {
				timerRecord.ProcessContext = map[string]interface{}{
					"process_key":      nextTimer.ProcessContext.ProcessKey,
					"process_version":  nextTimer.ProcessContext.ProcessVersion,
					"process_name":     nextTimer.ProcessContext.ProcessName,
					"component_source": nextTimer.ProcessContext.ComponentSource,
				}
			}

			if err := m.storage.SaveTimer(timerRecord); err != nil {
				fmt.Printf("Failed to save repeat timer to storage: %v\n", err)
			} else {
				fmt.Printf("Repeat timer saved to storage: %s (iteration %d)\n", nextTimer.ID, currentIteration+1)
			}
		}

		// Schedule next iteration
		// Планируем следующую итерацию
		handler := TimerHandlerFunc(m.handleTimerFired)
		return m.wheel.AddTimer(&nextTimer, handler)
	}

	return nil
}

// isParentScopeActive checks if parent scope is still active for boundary timer
// Проверяет активен ли родительский scope для boundary таймера
func (m *Manager) isParentScopeActive(timer *models.Timer) bool {
	// For now, return true - real implementation would check with process component
	// Пока возвращаем true - реальная реализация проверяла бы через process компонент
	// This is a simplification - in production we'd need a callback to process component
	// Это упрощение - в продакшене нужен был бы колбек в process компонент

	fmt.Printf("Checking parent scope for boundary timer %s - assuming active\n", timer.ID)
	return true
}
