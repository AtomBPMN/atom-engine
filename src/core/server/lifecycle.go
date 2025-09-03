/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// Start initializes and starts all components
// Инициализирует и запускает все компоненты
func (c *Core) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("core is already running")
	}

	// Initialize logger first
	err := logger.Init(&c.config.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	c.loggerReady = true
	logger.Info("Logger initialized successfully")

	// Create PID file
	err = c.createPIDFile()
	if err != nil {
		logger.Error("Failed to create PID file", logger.String("error", err.Error()))
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	// Initialize storage
	err = c.storage.Init()
	if err != nil {
		logger.Error("Failed to initialize storage", logger.String("error", err.Error()))
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Start storage
	err = c.storage.Start()
	if err != nil {
		logger.Error("Failed to start storage", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start storage: %w", err)
	}

	// Wait for storage to be ready
	if !c.storage.IsReady() {
		logger.Error("Storage is not ready")
		return fmt.Errorf("storage is not ready")
	}

	// Initialize and start timewheel component
	// Инициализируем и запускаем timewheel компонент
	err = c.timewheelComp.Initialize("") // Use default config
	if err != nil {
		logger.Error("Failed to initialize timewheel", logger.String("error", err.Error()))
		return fmt.Errorf("failed to initialize timewheel: %w", err)
	}

	err = c.timewheelComp.Start()
	if err != nil {
		logger.Error("Failed to start timewheel", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start timewheel: %w", err)
	}

	// Initialize and start process component
	// Инициализируем и запускаем process компонент

	// Set core interface for timer management
	// Устанавливаем интерфейс core для управления таймерами
	c.processComp.SetCore(c)

	err = c.processComp.Init()
	if err != nil {
		logger.Error("Failed to initialize process component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to initialize process component: %w", err)
	}

	err = c.processComp.Start()
	if err != nil {
		logger.Error("Failed to start process component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start process component: %w", err)
	}

	// Initialize and start parser component
	// Инициализируем и запускаем parser компонент
	err = c.parserComp.Init()
	if err != nil {
		logger.Error("Failed to initialize parser component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to initialize parser component: %w", err)
	}

	err = c.parserComp.Start()
	if err != nil {
		logger.Error("Failed to start parser component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start parser component: %w", err)
	}

	// Initialize and start jobs component
	// Инициализируем и запускаем jobs компонент
	err = c.jobsComp.Start()
	if err != nil {
		logger.Error("Failed to start jobs component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start jobs component: %w", err)
	}

	// Initialize and start messages component
	// Инициализируем и запускаем messages компонент
	err = c.messagesComp.Start()
	if err != nil {
		logger.Error("Failed to start messages component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start messages component: %w", err)
	}

	// Initialize and start expression component
	// Инициализируем и запускаем expression компонент
	err = c.expressionComp.Init()
	if err != nil {
		logger.Error("Failed to initialize expression component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to initialize expression component: %w", err)
	}

	err = c.expressionComp.Start()
	if err != nil {
		logger.Error("Failed to start expression component", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start expression component: %w", err)
	}

	// Log startup event
	err = c.storage.LogSystemEvent(models.EventTypeStartup, models.StatusInProgress, "Starting Atom Engine")
	if err != nil {
		logger.Warn("Failed to log startup event to storage", logger.String("error", err.Error()))
	}

	// Start gRPC server
	err = c.startGRPCServer()
	if err != nil {
		logger.Error("Failed to start gRPC server", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	// Start timewheel response processor
	// Запускаем обработчик ответов timewheel
	go c.processTimewheelResponses()

	// Start jobs response processor - TEMPORARILY DISABLED for gRPC response testing
	// Запускаем обработчик ответов jobs - ВРЕМЕННО ОТКЛЮЧЕН для тестирования gRPC ответов
	// go c.processJobsResponses()

	// Start messages response processor - TEMPORARILY DISABLED for gRPC response testing
	// Запускаем обработчик ответов messages - ВРЕМЕННО ОТКЛЮЧЕН для тестирования gRPC ответов
	// go c.processMessagesResponses()

	c.running = true
	logger.Info("Atom Engine started successfully")

	// Log successful startup
	err = c.storage.LogSystemEvent(models.EventTypeStartup, models.StatusSuccess, "Atom Engine started successfully")
	if err != nil {
		logger.Warn("Failed to log startup success to storage", logger.String("error", err.Error()))
	}

	// Restore timers from storage after everything is initialized
	// Восстанавливаем таймеры из storage после полной инициализации
	logger.Info("Restoring timers from storage")
	err = c.timewheelComp.RestoreTimers()
	if err != nil {
		logger.Error("Failed to restore timers", logger.String("error", err.Error()))
		// Don't fail startup - just warn about timer restoration
		logger.Warn("Timer restoration failed, continuing without restored timers")
	} else {
		logger.Info("Timer restoration completed")
	}

	return nil
}

// Stop gracefully stops all components
// Корректно останавливает все компоненты
func (c *Core) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return fmt.Errorf("core is not running")
	}

	logger.Info("Shutting down Atom Engine")

	// Log shutdown event
	err := c.storage.LogSystemEvent(models.EventTypeShutdown, models.StatusInProgress, "Shutting down Atom Engine")
	if err != nil {
		logger.Warn("Failed to log shutdown event to storage", logger.String("error", err.Error()))
	}

	// Stop gRPC server
	c.stopGRPCServer()

	// Stop expression component
	// Останавливаем expression компонент
	if c.expressionComp != nil {
		err := c.expressionComp.Stop()
		if err != nil {
			logger.Error("Failed to stop expression component", logger.String("error", err.Error()))
		} else {
			logger.Info("Expression component stopped")
		}
	}

	// Stop messages component
	// Останавливаем messages компонент
	if c.messagesComp != nil {
		err := c.messagesComp.Stop()
		if err != nil {
			logger.Error("Failed to stop messages component", logger.String("error", err.Error()))
		} else {
			logger.Info("Messages component stopped")
		}
	}

	// Stop jobs component
	// Останавливаем jobs компонент
	if c.jobsComp != nil {
		err := c.jobsComp.Stop()
		if err != nil {
			logger.Error("Failed to stop jobs component", logger.String("error", err.Error()))
		} else {
			logger.Info("Jobs component stopped")
		}
	}

	// Stop process component
	// Останавливаем process компонент
	if c.processComp != nil {
		err := c.processComp.Stop()
		if err != nil {
			logger.Error("Failed to stop process component", logger.String("error", err.Error()))
		} else {
			logger.Info("Process component stopped")
		}
	}

	// Stop parser component
	// Останавливаем parser компонент
	if c.parserComp != nil {
		err := c.parserComp.Stop()
		if err != nil {
			logger.Error("Failed to stop parser component", logger.String("error", err.Error()))
		} else {
			logger.Info("Parser component stopped")
		}
	}

	// Stop timewheel component
	// Останавливаем timewheel компонент
	if c.timewheelComp != nil {
		err := c.timewheelComp.Stop()
		if err != nil {
			logger.Error("Failed to stop timewheel", logger.String("error", err.Error()))
		} else {
			logger.Info("Timewheel component stopped")
		}
	}

	// Stop storage
	err = c.storage.Stop()
	if err != nil {
		logger.Error("Failed to stop storage", logger.String("error", err.Error()))
		c.logEvent(models.EventTypeShutdown, models.StatusFailed, fmt.Sprintf("Storage stop failed: %v", err))
		return fmt.Errorf("failed to stop storage: %w", err)
	}

	c.running = false
	logger.Info("Atom Engine shutdown completed")

	// Remove PID file
	if pidErr := c.removePIDFile(); pidErr != nil {
		logger.Warn("Failed to remove PID file", logger.String("error", pidErr.Error()))
	}

	// Close logger last
	logger.Close()

	return nil
}

// IsRunning returns core running status
// Возвращает статус работы core
func (c *Core) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// logEvent logs system event (helper method)
// Логирует системное событие (вспомогательный метод)
func (c *Core) logEvent(eventType, status, message string) error {
	if c.storage != nil && c.storage.IsReady() {
		return c.storage.LogSystemEvent(eventType, status, message)
	}
	return fmt.Errorf("storage not available")
}
