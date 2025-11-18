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

	"atom-engine/src/core/logger"
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
func NewManager(
	config Config,
	requestChannel <-chan string,
	responseChannel chan<- string,
	storage StorageInterface,
) (*Manager, error) {
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
	// Try to remove timer from wheel by ID (using timerIndex lookup)
	// Пытаемся удалить таймер из wheel по ID (используя поиск в timerIndex)
	if err := m.wheel.RemoveTimerByID(timerID); err != nil {
		// Timer not in wheel (possibly already fired or not scheduled)
		// Just mark as cancelled in storage if available
		// Таймер не в wheel (возможно уже сработал или не запланирован)
		// Просто помечаем как отмененный в storage если доступен
		if m.storage != nil {
			timerRecord, loadErr := m.storage.LoadTimer(timerID)
			if loadErr == nil {
				timerRecord.State = "CANCELLED"
				_ = m.storage.UpdateTimer(timerRecord)
			}
		}
		return nil // Don't fail the operation
	}

	// Also mark as cancelled in storage
	// Также помечаем как отмененный в storage
	if m.storage != nil {
		timerRecord, loadErr := m.storage.LoadTimer(timerID)
		if loadErr == nil {
			timerRecord.State = "CANCELLED"
			_ = m.storage.UpdateTimer(timerRecord)
		}
	}

	return nil
}

// processRequests processes incoming JSON requests
// Обрабатывает входящие JSON запросы
func (m *Manager) processRequests() {
	for {
		select {
		case jsonStr := <-m.requestChannel:
			if response, err := m.ProcessJSONRequest(jsonStr); err != nil {
				logger.Error("Failed to process timewheel JSON request",
					logger.String("request", jsonStr),
					logger.String("error", err.Error()))

				// Send error response back through channel if available
				// Отправляем ответ об ошибке обратно через канал если доступен
				if m.responseChannel != nil {
					errorResponse := fmt.Sprintf(`{"success":false,"error":"%s","request_id":"unknown"}`, err.Error())
					select {
					case m.responseChannel <- errorResponse:
						// Error response sent successfully
					default:
						logger.Warn("Failed to send error response - response channel full or closed",
							logger.String("original_error", err.Error()))
					}
				}
				continue
			} else if response != "" && m.responseChannel != nil {
				// Send successful response if available
				// Отправляем успешный ответ если доступен
				select {
				case m.responseChannel <- response:
					// Response sent successfully
				default:
					logger.Warn("Failed to send response - response channel full or closed",
						logger.String("response", response))
				}
			}
		case <-m.stopChan:
			return
		}
	}
}
