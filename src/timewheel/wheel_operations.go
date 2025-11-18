/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"fmt"
	"time"

	"atom-engine/src/core/models"
)

// AddTimer adds timer to timing wheel
// Добавляет таймер в timing wheel
func (htw *HierarchicalTimingWheel) AddTimer(timer *models.Timer, handler TimerHandler) error {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	// Check if timer already exists
	// Проверяем существует ли таймер
	if _, exists := htw.timerIndex[timer.ID]; exists {
		return ErrTimerAlreadyExists
	}

	// Calculate delay
	// Вычисляем задержку
	delay := time.Until(timer.DueDate)
	if delay <= 0 {
		// Timer should fire immediately
		// Таймер должен сработать немедленно
		go htw.fireTimer(timer, handler)
		return nil
	}

	// Find appropriate level
	// Находим подходящий уровень
	level := htw.findLevelForDelay(delay)
	if level == nil {
		return ErrTimerTooFar
	}

	// Add to level
	// Добавляем в уровень
	anchor, err := level.AddTimer(timer, handler, delay)
	if err != nil {
		return err
	}

	// Store in index
	// Сохраняем в индексе
	htw.timerIndex[timer.ID] = &TimerLocation{
		Level: anchor.Level,
		Slot:  anchor.Slot,
	}

	// Store anchor in timer variables for removal
	// Сохраняем якорь в переменных таймера для удаления
	if timer.Variables == nil {
		timer.Variables = make(map[string]interface{})
	}
	timer.Variables["_anchor"] = anchor

	return nil
}

// RemoveTimer removes timer from timing wheel
// Удаляет таймер из timing wheel
func (htw *HierarchicalTimingWheel) RemoveTimer(timer *models.Timer) error {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	// Remove from index
	// Удаляем из индекса
	delete(htw.timerIndex, timer.ID)

	// Get anchor from timer variables
	// Получаем якорь из переменных таймера
	if timer.Variables == nil {
		return ErrTimerNotFound
	}

	anchorInterface, exists := timer.Variables["_anchor"]
	if !exists {
		return ErrTimerNotFound
	}

	anchor, ok := anchorInterface.(*TimerAnchor)
	if !ok {
		return ErrInvalidAnchor
	}

	// Remove from level
	// Удаляем из уровня
	if anchor.Level >= len(htw.levels) {
		return ErrInvalidAnchor
	}

	return htw.levels[anchor.Level].RemoveTimer(anchor)
}

// RemoveTimerByID removes timer by ID using index lookup
// Удаляет таймер по ID используя поиск в индексе
func (htw *HierarchicalTimingWheel) RemoveTimerByID(timerID string) error {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	// Find timer location in index
	// Находим местоположение таймера в индексе
	location, exists := htw.timerIndex[timerID]
	if !exists {
		return ErrTimerNotFound
	}

	// Remove from index
	// Удаляем из индекса
	delete(htw.timerIndex, timerID)

	// Remove from level using direct slot access
	// Удаляем из уровня используя прямой доступ к слоту
	if location.Level >= len(htw.levels) {
		return ErrInvalidAnchor
	}

	level := htw.levels[location.Level]
	return level.RemoveTimerBySlotAndID(location.Slot, timerID)
}

// GetTimerLocation returns timer location in wheel
// Возвращает местоположение таймера в колесе
func (htw *HierarchicalTimingWheel) GetTimerLocation(timerID string) (*TimerLocation, bool) {
	htw.mu.RLock()
	defer htw.mu.RUnlock()

	location, exists := htw.timerIndex[timerID]
	return location, exists
}

// GetRemainingTime calculates remaining time until timer fires
// Вычисляет оставшееся время до срабатывания таймера
func (htw *HierarchicalTimingWheel) GetRemainingTime(timerID string) (time.Duration, error) {
	htw.mu.RLock()
	defer htw.mu.RUnlock()

	location, exists := htw.timerIndex[timerID]
	if !exists {
		return 0, fmt.Errorf("timer not found: %s", timerID)
	}

	if !htw.running {
		return 0, fmt.Errorf("timewheel not running")
	}

	// Find the timer in the appropriate level and slot to get DueDate
	// Находим таймер в соответствующем уровне и слоте чтобы получить DueDate
	level := htw.levels[location.Level]
	level.mu.RLock()
	defer level.mu.RUnlock()

	// Search through the slot to find the timer
	// Ищем в слоте чтобы найти таймер
	if location.Slot >= 0 && location.Slot < len(level.slots) {
		slot := level.slots[location.Slot]
		for e := slot.Front(); e != nil; e = e.Next() {
			if entry, ok := e.Value.(*TimerEntry); ok && entry.Timer.ID == timerID {
				// Use precise calculation based on DueDate
				// Используем точный расчет на основе DueDate
				remainingTime := time.Until(entry.Timer.DueDate)
				if remainingTime < 0 {
					return 0, nil // Timer should fire now
				}
				return remainingTime, nil
			}
		}
	}

	// Fallback to slot-based calculation if timer not found in slot
	// Запасной способ через слоты если таймер не найден
	currentSlot := level.currentSlot
	targetSlot := location.Slot

	slotsDiff := targetSlot - currentSlot
	if slotsDiff < 0 {
		// Timer should have fired already, return 0
		// Таймер должен был уже сработать, возвращаем 0
		return 0, nil
	}
	if slotsDiff == 0 {
		// Same slot means timer fires within current tick
		// Тот же слот означает что таймер сработает в текущем тике
		return time.Second, nil
	}

	remainingTime := time.Duration(slotsDiff) * level.tick
	return remainingTime, nil
}

// GetStats returns timing wheel statistics
// Возвращает статистику timing wheel
func (htw *HierarchicalTimingWheel) GetStats() Stats {
	htw.mu.RLock()
	defer htw.mu.RUnlock()

	stats := Stats{
		LevelsCount: len(htw.levels),
		LevelStats:  make([]LevelStats, 0, len(htw.levels)),
		Running:     htw.running,
		StartTime:   htw.startTime,
	}

	totalTimers := 0
	for _, level := range htw.levels {
		levelStats := level.GetStats()
		stats.LevelStats = append(stats.LevelStats, levelStats)
		totalTimers += levelStats.TotalTimers
	}

	stats.TotalTimers = totalTimers
	if len(htw.levels) > 0 {
		stats.TotalHorizon = htw.levels[len(htw.levels)-1].GetHorizon()
	}

	return stats
}

// removeFromIndex removes timer from index
// Удаляет таймер из индекса
func (htw *HierarchicalTimingWheel) removeFromIndex(timerID string) {
	htw.mu.Lock()
	defer htw.mu.Unlock()
	delete(htw.timerIndex, timerID)
}

// rescheduleTimer reschedules timer to higher level
// Перепланирует таймер на более высокий уровень
func (htw *HierarchicalTimingWheel) rescheduleTimer(entry *TimerEntry) {
	delay := time.Until(entry.Timer.DueDate)
	level := htw.findLevelForDelay(delay)
	if level != nil {
		level.AddTimer(entry.Timer, entry.Handler, delay)
	}
}
