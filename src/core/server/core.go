/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"atom-engine/src/core/config"
	"atom-engine/src/core/grpc"
	"atom-engine/src/core/models"
	"atom-engine/src/expression"
	"atom-engine/src/incidents"
	"atom-engine/src/jobs"
	"atom-engine/src/messages"
	"atom-engine/src/parser"
	"atom-engine/src/process"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
)

// Core manages all system components
// Управляет всеми компонентами системы
type Core struct {
	config        *config.Config
	storage       storage.Storage
	grpcServer    *grpc.Server
	timewheelComp *timewheel.Component

	processComp    *process.Component
	parserComp     *parser.Component
	jobsComp       *jobs.Component
	messagesComp   *messages.Component
	expressionComp *expression.Component
	incidentsComp  *incidents.Component
	loggerReady    bool
	mu             sync.RWMutex
	running        bool
}

// NewCore creates new core instance
// Создает новый экземпляр core
func NewCore(configPath string) (*Core, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return NewCoreWithConfig(cfg)
}

// NewCoreWithConfig creates new core instance with provided config
// Создает новый экземпляр core с предоставленной конфигурацией
func NewCoreWithConfig(cfg *config.Config) (*Core, error) {
	// Set instance name for ID generation
	// Устанавливаем имя инстанса для генерации ID
	models.SetInstanceName(cfg.InstanceName)

	storageConfig := &storage.Config{
		Path: cfg.Database.Path,
	}

	storageInstance := storage.NewStorage(storageConfig)

	// Initialize timewheel component with storage
	// Инициализируем timewheel компонент с storage
	timewheelComp := timewheel.NewComponentWithStorage(storageInstance)

	// Initialize process component with storage
	// Инициализируем process компонент с storage
	processComp := process.NewComponent(storageInstance)

	// Initialize parser component with config and storage
	// Инициализируем parser компонент с конфигурацией и storage
	parserComp := parser.NewComponent(cfg, storageInstance)

	// Initialize jobs component with storage
	// Инициализируем jobs компонент с storage
	jobsComp := jobs.NewComponent(cfg, storageInstance)

	// Initialize messages component with storage
	// Инициализируем messages компонент с storage
	messagesComp := messages.NewComponent(cfg, storageInstance)

	// Initialize expression component
	// Инициализируем expression компонент
	expressionComp := expression.NewComponent()

	// Initialize incidents component with storage
	// Инициализируем incidents компонент с storage
	incidentsComp := incidents.NewComponent(cfg, storageInstance)

	return &Core{
		config:        cfg,
		storage:       storageInstance,
		timewheelComp: timewheelComp,

		processComp:    processComp,
		parserComp:     parserComp,
		jobsComp:       jobsComp,
		messagesComp:   messagesComp,
		expressionComp: expressionComp,
		incidentsComp:  incidentsComp,
		loggerReady:    false,
		running:        false,
	}, nil
}

// GetMessagesComponent returns messages component
func (c *Core) GetMessagesComponent() interface{} {
	return c.messagesComp
}

// GetJobsComponent returns jobs component
func (c *Core) GetJobsComponent() interface{} {
	return c.jobsComp
}

// GetExpressionComponent returns expression component
func (c *Core) GetExpressionComponent() interface{} {
	return c.expressionComp
}

// GetIncidentsComponent returns incidents component
func (c *Core) GetIncidentsComponent() interface{} {
	return c.incidentsComp
}

// GetParserComponent returns parser component
func (c *Core) GetParserComponent() interface{} {
	return c.parserComp
}

// GetStorage returns storage instance
func (c *Core) GetStorage() interface{} {
	return c.storage
}

// GetStorageComponent returns storage component
func (c *Core) GetStorageComponent() grpc.StorageComponentInterface {
	return c.storage
}

// SendMessage sends JSON message to specified component
// Отправляет JSON сообщение указанному компоненту
func (c *Core) SendMessage(componentName, messageJSON string) error {
	component := c.getComponentByName(componentName)
	if component == nil {
		return fmt.Errorf("component not found: %s", componentName)
	}

	processor, ok := component.(models.JSONMessageProcessor)
	if !ok {
		return fmt.Errorf("component %s does not support JSON messages", componentName)
	}

	return processor.ProcessMessage(context.Background(), messageJSON)
}

// getComponentByName returns component by name
// Возвращает компонент по имени
func (c *Core) getComponentByName(componentName string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch componentName {
	case "jobs":
		return c.jobsComp
	case "messages":
		return c.messagesComp
	case "parser":
		return c.parserComp
	case "process":
		return c.processComp
	case "timewheel":
		return c.timewheelComp
	case "expression":
		return c.expressionComp
	case "incidents":
		return c.incidentsComp
	case "storage":
		return c.storage
	default:
		return nil
	}
}

// WaitForParserResponse waits for parser response with timeout
// Ожидает ответ от парсера с таймаутом
func (c *Core) WaitForParserResponse(timeoutMs int) (string, error) {
	if c.parserComp == nil {
		return "", fmt.Errorf("parser component not available")
	}

	responseChannel := c.parserComp.GetResponseChannel()
	if responseChannel == nil {
		return "", fmt.Errorf("parser response channel not available")
	}

	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for parser response after %dms", timeoutMs)
	}
}

// WaitForJobsResponse waits for jobs response with timeout
// Ожидает ответ от jobs компонента с таймаутом
func (c *Core) WaitForJobsResponse(timeoutMs int) (string, error) {
	if c.jobsComp == nil {
		return "", fmt.Errorf("jobs component not available")
	}

	responseChannel := c.jobsComp.GetResponseChannel()
	if responseChannel == nil {
		return "", fmt.Errorf("jobs response channel not available")
	}

	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for jobs response after %dms", timeoutMs)
	}
}

// WaitForMessagesResponse waits for messages response with timeout
// Ожидает ответ от messages компонента с таймаутом
func (c *Core) WaitForMessagesResponse(timeoutMs int) (string, error) {
	if c.messagesComp == nil {
		return "", fmt.Errorf("messages component not available")
	}

	responseChannel := c.messagesComp.GetResponseChannel()
	if responseChannel == nil {
		return "", fmt.Errorf("messages response channel not available")
	}

	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for messages response after %dms", timeoutMs)
	}
}

// WaitForIncidentsResponse waits for incidents response with timeout
// Ожидает ответ от incidents компонента с таймаутом
func (c *Core) WaitForIncidentsResponse(timeoutMs int) (string, error) {
	if c.incidentsComp == nil {
		return "", fmt.Errorf("incidents component not available")
	}

	responseChannel := c.incidentsComp.GetResponseChannel()
	if responseChannel == nil {
		return "", fmt.Errorf("incidents response channel not available")
	}

	// Clear any old responses from channel (non-blocking)
	// Очищаем старые ответы из канала (неблокирующе)
	for {
		select {
		case <-responseChannel:
			// Discard old response
		default:
			// Channel is empty, proceed to wait for new response
			goto waitForResponse
		}
	}

waitForResponse:
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for incidents response after %dms", timeoutMs)
	}
}
