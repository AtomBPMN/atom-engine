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
)

// NewHierarchicalTimingWheel creates new hierarchical timing wheel
// Создает новый иерархический timing wheel
func NewHierarchicalTimingWheel(config Config, responseChannel chan<- string) (*HierarchicalTimingWheel, error) {
	if len(config.Levels) == 0 {
		return nil, ErrInvalidConfig
	}

	htw := &HierarchicalTimingWheel{
		levels:          make([]*TimingWheelLevel, 0, len(config.Levels)),
		timerIndex:      make(map[string]*TimerLocation),
		running:         false,
		stopChan:        make(chan struct{}),
		responseChannel: responseChannel,
	}

	// Create levels
	// Создаем уровни
	for i, levelConfig := range config.Levels {
		tick, err := time.ParseDuration(levelConfig.Tick)
		if err != nil {
			return nil, fmt.Errorf("invalid tick duration %s: %w", levelConfig.Tick, err)
		}

		level := NewTimingWheelLevel(i, tick, levelConfig.Size)
		htw.levels = append(htw.levels, level)
	}

	return htw, nil
}

// Start starts the timing wheel
// Запускает timing wheel
func (htw *HierarchicalTimingWheel) Start() error {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	if htw.running {
		return ErrWheelAlreadyRunning
	}

	htw.running = true
	htw.startTime = time.Now()

	// Start with smallest tick interval
	// Запускаем с наименьшим интервалом тика
	if len(htw.levels) > 0 {
		htw.ticker = time.NewTicker(htw.levels[0].tick)
		htw.wg.Add(1)
		go htw.run()
	}

	return nil
}

// Stop stops the timing wheel
// Останавливает timing wheel
func (htw *HierarchicalTimingWheel) Stop() error {
	htw.mu.Lock()
	defer htw.mu.Unlock()

	if !htw.running {
		return ErrWheelNotRunning
	}

	htw.running = false
	close(htw.stopChan)

	if htw.ticker != nil {
		htw.ticker.Stop()
	}

	htw.wg.Wait()
	return nil
}

// findLevelForDelay finds appropriate level for given delay
// Находит подходящий уровень для заданной задержки
func (htw *HierarchicalTimingWheel) findLevelForDelay(delay time.Duration) *TimingWheelLevel {
	// Simple and correct algorithm: select first level that can fit the timer
	// Простой и правильный алгоритм: выбираем первый уровень, который может вместить таймер
	for _, level := range htw.levels {
		if delay <= level.GetHorizon() {
			return level
		}
	}
	return nil
}

// getLevelIndex returns index of level
// Возвращает индекс уровня
func (htw *HierarchicalTimingWheel) getLevelIndex(targetLevel *TimingWheelLevel) int {
	if targetLevel == nil {
		return len(htw.levels) - 1
	}

	for i, level := range htw.levels {
		if level == targetLevel {
			return i
		}
	}
	return len(htw.levels) - 1
}
