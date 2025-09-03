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

// Manager manages timing wheel and handles JSON communication
// Менеджер управляет timing wheel и обрабатывает JSON коммуникацию
type Manager struct {
	wheel           *HierarchicalTimingWheel
	parser          *ISO8601DurationParser
	requestChannel  <-chan string
	responseChannel chan<- string
	running         bool
	stopChan        chan struct{}
	storage         StorageInterface // For updating timer status
}

// NewManager creates new timing wheel manager
// Создает новый менеджер timing wheel
func NewManager(config Config, requestChannel <-chan string, responseChannel chan<- string, storage StorageInterface) (*Manager, error) {
	wheel, err := NewHierarchicalTimingWheel(config, responseChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to create timing wheel: %w", err)
	}

	return &Manager{
		wheel:           wheel,
		parser:          NewISO8601DurationParser(),
		requestChannel:  requestChannel,
		responseChannel: responseChannel,
		running:         false,
		stopChan:        make(chan struct{}),
		storage:         storage,
	}, nil
}

// Start starts the manager
// Запускает менеджер
func (m *Manager) Start() error {
	if m.running {
		return fmt.Errorf("manager already running")
	}

	// Start timing wheel
	// Запускаем timing wheel
	if err := m.wheel.Start(); err != nil {
		return fmt.Errorf("failed to start timing wheel: %w", err)
	}

	m.running = true

	// Start request processing goroutine
	// Запускаем горутину обработки запросов
	go m.processRequests()

	return nil
}

// Stop stops the manager
// Останавливает менеджер
func (m *Manager) Stop() error {
	if !m.running {
		return fmt.Errorf("manager not running")
	}

	m.running = false
	close(m.stopChan)

	// Stop timing wheel
	// Останавливаем timing wheel
	return m.wheel.Stop()
}

// GetStats returns timing wheel statistics
// Возвращает статистику timing wheel
func (m *Manager) GetStats() Stats {
	return m.wheel.GetStats()
}

// GetTimerLocation returns timer location in wheel
// Возвращает местоположение таймера в колесе
func (m *Manager) GetTimerLocation(timerID string) (*TimerLocation, bool) {
	return m.wheel.GetTimerLocation(timerID)
}

// GetRemainingTime calculates remaining time until timer fires
// Вычисляет оставшееся время до срабатывания таймера
func (m *Manager) GetRemainingTime(timerID string) (time.Duration, error) {
	return m.wheel.GetRemainingTime(timerID)
}

// CancelTimer cancels timer by ID
// Отменяет таймер по ID
func (m *Manager) CancelTimer(timerID string) error {
	// Load timer from storage to get anchor for wheel removal
	// Загружаем timer из storage чтобы получить anchor для удаления из wheel
	if m.storage == nil {
		return fmt.Errorf("timer cancellation requires storage to load timer anchor")
	}

	// Load timer record from storage
	// Загружаем запись timer из storage
	timerRecord, err := m.storage.LoadTimer(timerID)
	if err != nil {
		return fmt.Errorf("failed to load timer for cancellation: %w", err)
	}

	// Convert timer record to models.Timer with anchor
	// Конвертируем запись timer в models.Timer с anchor
	timer := &models.Timer{
		ID:                timerRecord.ID,
		ElementID:         timerRecord.ElementID,
		ProcessInstanceID: timerRecord.ProcessInstanceID,
		ExecutionTokenID:  timerRecord.TokenID,
		Variables:         timerRecord.Variables,
	}

	// Remove timer from wheel using anchor
	// Удаляем timer из wheel используя anchor
	if err := m.wheel.RemoveTimer(timer); err != nil {
		return fmt.Errorf("failed to remove timer from wheel: %w", err)
	}

	return nil
}

// processRequests processes incoming JSON requests
// Обрабатывает входящие JSON запросы
func (m *Manager) processRequests() {
	for {
		select {
		case jsonStr := <-m.requestChannel:
			if _, err := m.ProcessJSONRequest(jsonStr); err != nil {
				// In production, you'd log this error
				// В продакшене вы бы логировали эту ошибку
				continue
			}
		case <-m.stopChan:
			return
		}
	}
}
