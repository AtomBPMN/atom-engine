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
	"time"

	"atom-engine/src/core/models"
)

// run main timing wheel loop
// Основной цикл timing wheel
func (htw *HierarchicalTimingWheel) run() {
	defer htw.wg.Done()

	for {
		select {
		case <-htw.ticker.C:
			htw.processTick()
		case <-htw.stopChan:
			return
		}
	}
}

// processTick processes one tick
// Обрабатывает один тик
func (htw *HierarchicalTimingWheel) processTick() {
	if len(htw.levels) == 0 {
		return
	}

	now := time.Now()

	// Process L0 (most frequent level)
	// Обрабатываем L0 (самый частый уровень)
	expiredTimers, overflowTimers := htw.levels[0].Tick()

	// Fire expired timers
	// Запускаем истекшие таймеры
	for _, entry := range expiredTimers {
		go htw.fireTimer(entry.Timer, entry.Handler)
		htw.removeFromIndex(entry.Timer.ID)
	}

	// Reschedule overflow timers
	// Перепланируем переполненные таймеры
	for _, entry := range overflowTimers {
		htw.rescheduleTimer(entry)
	}

	// Check if higher levels need to cascade
	// Проверяем нужно ли каскадирование более высоких уровней
	htw.processCascade()

	// Proactive cascade - move timers to correct levels
	// Проактивное каскадирование - перемещаем таймеры на правильные уровни
	htw.proactiveCascade(now)
}

// fireTimer fires a timer by sending JSON response
// Запускает таймер отправкой JSON ответа
func (htw *HierarchicalTimingWheel) fireTimer(timer *models.Timer, handler TimerHandler) {
	// Update timer state
	// Обновляем состояние таймера
	timer.State = models.TimerStateFired
	timer.UpdatedAt = time.Now()

	// Create response
	// Создаем ответ
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

	// Send JSON response to core engine
	// Отправляем JSON ответ в core engine
	if htw.responseChannel != nil {
		if jsonData, err := json.Marshal(response); err == nil {
			select {
			case htw.responseChannel <- string(jsonData):
			default:
				// Channel is full, log error but continue
				// Канал заполнен, логируем ошибку но продолжаем
			}
		}
	}

	// Also call handler if provided
	// Также вызываем обработчик если предоставлен
	if handler != nil {
		ctx := context.Background()
		handler.HandleTimer(ctx, timer)
	}
}

// processCascade processes cascade from higher levels
// Обрабатывает каскадирование с более высоких уровней
func (htw *HierarchicalTimingWheel) processCascade() {
	// Implementation of cascade logic for higher levels
	// Реализация логики каскадирования для более высоких уровней
	// This is simplified - full implementation would track ticks per level
	// Это упрощенная версия - полная реализация отслеживала бы тики на уровень
}

// proactiveCascade proactively cascades timers based on remaining time
// Проактивно каскадирует таймеры на основе оставшегося времени
func (htw *HierarchicalTimingWheel) proactiveCascade(now time.Time) {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	// Go through levels from highest to lowest (L1, L2, L3...)
	// Проходим по уровням от высших к низшим (L1, L2, L3...)
	for levelIndex := len(htw.levels) - 1; levelIndex >= 1; levelIndex-- {
		level := htw.levels[levelIndex]

		// Get all timers from current level
		// Получаем все таймеры с текущего уровня
		allTimers := level.GetAllTimers()
		if len(allTimers) == 0 {
			continue
		}

		var toReschedule []*TimerEntry

		// Check each timer for cascade necessity
		// Проверяем каждый таймер на необходимость каскадирования
		for _, entry := range allTimers {
			delta := entry.Timer.DueDate.Sub(now)

			if delta <= 0 {
				// Overdue timer, execute immediately
				// Просроченный таймер, выполняем немедленно
				toReschedule = append(toReschedule, entry)
				continue
			}

			// Determine correct level for remaining time
			// Определяем правильный уровень для оставшегося времени
			correctLevel := htw.findLevelForDelay(delta)
			correctLevelIndex := htw.getLevelIndex(correctLevel)

			// If timer should be on a lower level
			// Если таймер должен быть на более низком уровне
			if correctLevelIndex < levelIndex {
				toReschedule = append(toReschedule, entry)
			}
		}

		// Move timers that need to be cascaded
		// Перемещаем таймеры, которые нужно каскадировать
		for _, entry := range toReschedule {
			// Remove timer from current level
			// Удаляем таймер с текущего уровня
			if anchor, ok := entry.Timer.Variables["_anchor"].(*TimerAnchor); ok {
				level.RemoveTimer(anchor)
				delete(htw.timerIndex, entry.Timer.ID)
			}

			// Cascade to correct level
			// Каскадируем на правильный уровень
			delta := entry.Timer.DueDate.Sub(now)
			if delta <= 0 {
				// Execute immediately
				// Выполняем немедленно
				go htw.fireTimer(entry.Timer, entry.Handler)
			} else {
				// Add to correct level
				// Добавляем на правильный уровень
				targetLevel := htw.findLevelForDelay(delta)
				if targetLevel != nil {
					anchor, err := targetLevel.AddTimer(entry.Timer, entry.Handler, delta)
					if err == nil && anchor != nil {
						entry.Timer.Variables["_anchor"] = anchor
						htw.timerIndex[entry.Timer.ID] = &TimerLocation{
							Level: anchor.Level,
							Slot:  anchor.Slot,
						}
					}
				}
			}
		}
	}
}
