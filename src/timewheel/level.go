/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"container/list"
	"time"

	"atom-engine/src/core/models"
)

// NewTimingWheelLevel creates new timing wheel level
// Создает новый уровень timing wheel
func NewTimingWheelLevel(levelID int, tick time.Duration, size int) *TimingWheelLevel {
	level := &TimingWheelLevel{
		levelID:     levelID,
		tick:        tick,
		size:        size,
		currentSlot: 0,
		slots:       make([]*list.List, size),
		horizon:     tick * time.Duration(size),
	}

	// Initialize all slots
	// Инициализируем все слоты
	for i := 0; i < size; i++ {
		level.slots[i] = list.New()
	}

	return level
}

// AddTimer adds timer to appropriate slot
// Добавляет таймер в соответствующий слот
func (twl *TimingWheelLevel) AddTimer(
	timer *models.Timer,
	handler TimerHandler,
	delay time.Duration,
) (*TimerAnchor, error) {
	twl.mu.Lock()
	defer twl.mu.Unlock()

	// Calculate slot position using relative delay from current position
	// Вычисляем позицию слота используя относительную задержку от текущей позиции
	slots := int(delay / twl.tick)
	slotIndex := (twl.currentSlot + slots) % twl.size
	entry := &TimerEntry{
		Timer:   timer,
		Handler: handler,
	}

	// Add to slot
	// Добавляем в слот
	element := twl.slots[slotIndex].PushBack(entry)

	// Create anchor for O(1) removal
	// Создаем якорь для O(1) удаления
	anchor := &TimerAnchor{
		Level:   twl.levelID,
		Slot:    slotIndex,
		Element: element,
	}

	entry.Anchor = anchor
	return anchor, nil
}

// RemoveTimer removes timer by anchor
// Удаляет таймер по якорю
func (twl *TimingWheelLevel) RemoveTimer(anchor *TimerAnchor) error {
	twl.mu.Lock()
	defer twl.mu.Unlock()

	if anchor.Level != twl.levelID {
		return ErrInvalidAnchor
	}

	if anchor.Slot < 0 || anchor.Slot >= twl.size {
		return ErrInvalidAnchor
	}

	if anchor.Element == nil {
		return ErrInvalidAnchor
	}

	twl.slots[anchor.Slot].Remove(anchor.Element)
	return nil
}

// RemoveTimerBySlotAndID removes timer by slot and ID
// Удаляет таймер по слоту и ID
func (twl *TimingWheelLevel) RemoveTimerBySlotAndID(slot int, timerID string) error {
	twl.mu.Lock()
	defer twl.mu.Unlock()

	if slot < 0 || slot >= twl.size {
		return ErrInvalidAnchor
	}

	slotList := twl.slots[slot]
	for element := slotList.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*TimerEntry)
		if entry.Timer.ID == timerID {
			slotList.Remove(element)
			return nil
		}
	}

	return ErrTimerNotFound
}

// Tick advances level by one tick and returns expired timers
// Продвигает уровень на один тик и возвращает истекшие таймеры
func (twl *TimingWheelLevel) Tick(now time.Time) ([]*TimerEntry, []*TimerEntry) {
	twl.mu.Lock()
	defer twl.mu.Unlock()

	var expiredTimers []*TimerEntry
	var overflowTimers []*TimerEntry

	// Get current slot
	// Получаем текущий слот
	currentSlotList := twl.slots[twl.currentSlot]

	// Process all timers in current slot
	// Обрабатываем все таймеры в текущем слоте
	for currentSlotList.Len() > 0 {
		element := currentSlotList.Front()
		entry := element.Value.(*TimerEntry)
		currentSlotList.Remove(element)

		// Check if timer should fire now using consistent time
		// Проверяем должен ли таймер сработать сейчас используя единое время
		if now.After(entry.Timer.DueDate) || now.Equal(entry.Timer.DueDate) {
			expiredTimers = append(expiredTimers, entry)
		} else {
			// Timer needs to be rescheduled to higher level
			// Таймер нужно перепланировать на более высокий уровень
			overflowTimers = append(overflowTimers, entry)
		}
	}

	// Advance to next slot
	// Переходим к следующему слоту
	twl.currentSlot = (twl.currentSlot + 1) % twl.size

	return expiredTimers, overflowTimers
}

// GetCurrentSlot returns current slot index
// Возвращает индекс текущего слота
func (twl *TimingWheelLevel) GetCurrentSlot() int {
	twl.mu.RLock()
	defer twl.mu.RUnlock()
	return twl.currentSlot
}

// GetTotalTimers returns total number of timers in level
// Возвращает общее количество таймеров в уровне
func (twl *TimingWheelLevel) GetTotalTimers() int {
	twl.mu.RLock()
	defer twl.mu.RUnlock()

	total := 0
	for _, slot := range twl.slots {
		total += slot.Len()
	}
	return total
}

// GetAllTimers returns all timer entries from level
// Возвращает все записи таймеров с уровня
func (twl *TimingWheelLevel) GetAllTimers() []*TimerEntry {
	twl.mu.RLock()
	defer twl.mu.RUnlock()

	var allTimers []*TimerEntry
	for _, slot := range twl.slots {
		for element := slot.Front(); element != nil; element = element.Next() {
			if entry, ok := element.Value.(*TimerEntry); ok {
				allTimers = append(allTimers, entry)
			}
		}
	}
	return allTimers
}

// GetHorizon returns time horizon of this level
// Возвращает временной горизонт этого уровня
func (twl *TimingWheelLevel) GetHorizon() time.Duration {
	return twl.horizon
}

// GetStats returns level statistics
// Возвращает статистику уровня
func (twl *TimingWheelLevel) GetStats() LevelStats {
	twl.mu.RLock()
	defer twl.mu.RUnlock()

	return LevelStats{
		LevelID:     twl.levelID,
		Tick:        twl.tick,
		Size:        twl.size,
		CurrentSlot: twl.currentSlot,
		TotalTimers: twl.GetTotalTimers(),
		Horizon:     twl.horizon,
	}
}
