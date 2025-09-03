/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"encoding/json"
	"fmt"

	"atom-engine/src/storage"
)

// StorageInterface defines storage operations for timers
// Определяет операции storage для таймеров
type StorageInterface interface {
	SaveTimer(timer *storage.TimerRecord) error
	LoadTimer(timerID string) (*storage.TimerRecord, error)
	LoadAllTimers() ([]*storage.TimerRecord, error)
	DeleteTimer(timerID string) error
}

// Component represents timewheel component for core integration
// Компонент представляет timewheel компонент для интеграции с core
type Component struct {
	manager         *Manager
	storage         StorageInterface
	requestChannel  chan string
	responseChannel chan string
	ready           bool
}

// NewComponent creates new timewheel component
// Создает новый timewheel компонент
func NewComponent() *Component {
	return &Component{
		requestChannel:  make(chan string, 100), // Buffered for async processing
		responseChannel: make(chan string, 100), // Buffered for timer responses
		ready:           false,
	}
}

// NewComponentWithStorage creates new timewheel component with storage
// Создает новый timewheel компонент с storage
func NewComponentWithStorage(storage StorageInterface) *Component {
	return &Component{
		storage:         storage,
		requestChannel:  make(chan string, 100), // Buffered for async processing
		responseChannel: make(chan string, 100), // Buffered for timer responses
		ready:           false,
	}
}

// Initialize initializes the component with configuration
// Инициализирует компонент с конфигурацией
func (c *Component) Initialize(configJSON string) error {
	// Parse configuration
	// Парсим конфигурацию
	var config Config
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return fmt.Errorf("failed to parse timewheel config: %w", err)
		}
	} else {
		// Use default configuration
		// Используем конфигурацию по умолчанию
		config = DefaultConfig
	}

	// Create manager with storage
	// Создаем менеджер с storage
	manager, err := NewManager(config, c.requestChannel, c.responseChannel, c.storage)
	if err != nil {
		return fmt.Errorf("failed to create timewheel manager: %w", err)
	}

	c.manager = manager
	c.ready = true
	return nil
}

// Start starts the timewheel component
// Запускает timewheel компонент
func (c *Component) Start() error {
	if !c.ready {
		return fmt.Errorf("component not initialized")
	}

	if err := c.manager.Start(); err != nil {
		return fmt.Errorf("failed to start timewheel manager: %w", err)
	}

	return nil
}

// Stop stops the timewheel component
// Останавливает timewheel компонент
func (c *Component) Stop() error {
	if c.manager == nil {
		return nil
	}

	return c.manager.Stop()
}

// IsReady returns component ready status
// Возвращает статус готовности компонента
func (c *Component) IsReady() bool {
	return c.ready
}

// GetResponseChannel returns channel for timer responses
// Возвращает канал для ответов таймеров
func (c *Component) GetResponseChannel() <-chan string {
	return c.responseChannel
}

// GetStats returns current timewheel statistics directly
// Возвращает текущую статистику timewheel напрямую
func (c *Component) GetStats() (Stats, error) {
	if c.manager == nil {
		return Stats{}, fmt.Errorf("timewheel manager not initialized")
	}
	return c.manager.GetStats(), nil
}

// GetTimerInfo gets timer location and remaining time
// Получает местоположение таймера и оставшееся время
func (c *Component) GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool) {
	if c.manager == nil {
		return 0, 0, false
	}

	location, exists := c.manager.GetTimerLocation(timerID)
	if !exists {
		return 0, 0, false
	}

	// Get remaining time from manager
	// Получаем оставшееся время от manager
	remaining, err := c.manager.GetRemainingTime(timerID)
	if err != nil {
		return location.Level, 0, true // Found but no remaining time
	}

	return location.Level, int64(remaining.Seconds()), true
}
