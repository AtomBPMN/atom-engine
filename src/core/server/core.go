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

	"atom-engine/src/core/auth"
	"atom-engine/src/core/config"
	"atom-engine/src/core/grpc"
	"atom-engine/src/core/interfaces"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/core/restapi"
	"atom-engine/src/core/restapi/handlers"
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
	restServer    *restapi.Server
	timewheelComp *timewheel.Component

	processComp    *process.Component
	parserComp     *parser.Component
	jobsComp       *jobs.Component
	messagesComp   *messages.Component
	expressionComp *expression.Component
	incidentsComp  *incidents.Component
	authComp       auth.Component
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
		Path:    cfg.Database.Path,
		Options: convertStorageOptions(&cfg.Storage.Options),
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

	// Initialize auth component
	// Инициализируем auth компонент
	authComp := auth.NewComponent()

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
		authComp:       authComp,
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

// GetAuthComponent returns auth component
func (c *Core) GetAuthComponent() interface{} {
	return c.authComp
}

// GetStorage returns storage instance
func (c *Core) GetStorage() interface{} {
	return c.storage
}

// Typed component getters for strict type safety
// Типизированные геттеры компонентов для строгой типобезопасности

// GetMessagesComponentTyped returns typed messages component
func (c *Core) GetMessagesComponentTyped() interfaces.MessagesComponentInterface {
	return c.messagesComp
}

// GetJobsComponentTyped returns typed jobs component
func (c *Core) GetJobsComponentTyped() interfaces.JobsComponentInterface {
	return c.jobsComp
}

// GetParserComponentTyped returns typed parser component
func (c *Core) GetParserComponentTyped() interfaces.ParserComponentInterface {
	return c.parserComp
}

// GetExpressionComponentTyped returns typed expression component
func (c *Core) GetExpressionComponentTyped() interfaces.ExpressionComponentInterface {
	return c.expressionComp
}

// GetIncidentsComponentTyped returns typed incidents component
func (c *Core) GetIncidentsComponentTyped() interfaces.IncidentsComponentInterface {
	return c.incidentsComp
}

// GetAuthComponentTyped returns typed auth component
func (c *Core) GetAuthComponentTyped() interfaces.AuthComponentInterface {
	return c.authComp
}

// GetStorageTyped returns typed storage interface
func (c *Core) GetStorageTyped() interfaces.StorageInterface {
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

// REST API interface implementations
// Реализации интерфейсов для REST API

// REST API implementations that override gRPC methods for handlers
// These methods provide REST-specific implementations

// Storage interfaces
func (c *Core) GetStorageStatusForREST() (*handlers.StorageStatusResponse, error) {
	grpcStatus, err := c.GetStorageStatus()
	if err != nil {
		return nil, err
	}
	return &handlers.StorageStatusResponse{
		IsConnected:   grpcStatus.IsConnected,
		IsHealthy:     grpcStatus.IsHealthy,
		Status:        grpcStatus.Status,
		UptimeSeconds: grpcStatus.UptimeSeconds,
	}, nil
}

func (c *Core) GetStorageInfoForREST() (*handlers.StorageInfoResponse, error) {
	grpcInfo, err := c.GetStorageInfo()
	if err != nil {
		return nil, err
	}
	// Convert statistics from int64 to string for REST handlers
	statistics := make(map[string]string)
	for k, v := range grpcInfo.Statistics {
		statistics[k] = fmt.Sprintf("%d", v)
	}

	return &handlers.StorageInfoResponse{
		TotalSizeBytes: grpcInfo.TotalSizeBytes,
		UsedSizeBytes:  grpcInfo.UsedSizeBytes,
		FreeSizeBytes:  grpcInfo.FreeSizeBytes,
		TotalKeys:      grpcInfo.TotalKeys,
		DatabasePath:   grpcInfo.DatabasePath,
		Statistics:     statistics,
	}, nil
}

// convertStorageOptions converts config storage options to storage package format
// Конвертирует настройки storage из config в формат пакета storage
func convertStorageOptions(configOptions *config.StorageOptionsConfig) *storage.StorageOptionsConfig {
	if configOptions == nil {
		return nil
	}

	options := &storage.StorageOptionsConfig{
		SyncWrites:       configOptions.SyncWrites,
		ValueLogFileSize: configOptions.ValueLogFileSize,
	}

	if configOptions.Performance != nil {
		options.Performance = &storage.BadgerPerformanceConfig{
			MemTableSize:            configOptions.Performance.MemTableSize,
			NumMemtables:            configOptions.Performance.NumMemtables,
			NumLevelZeroTables:      configOptions.Performance.NumLevelZeroTables,
			NumLevelZeroTablesStall: configOptions.Performance.NumLevelZeroTablesStall,
			ValueCacheSize:          configOptions.Performance.ValueCacheSize,
			BlockCacheSize:          configOptions.Performance.BlockCacheSize,
			IndexCacheSize:          configOptions.Performance.IndexCacheSize,
			BaseTableSize:           configOptions.Performance.BaseTableSize,
			MaxTableSize:            configOptions.Performance.MaxTableSize,
			LevelSizeMultiplier:     configOptions.Performance.LevelSizeMultiplier,
			NumCompactors:           configOptions.Performance.NumCompactors,
			CompactL0OnClose:        configOptions.Performance.CompactL0OnClose,
			TableLoadingMode:        configOptions.Performance.TableLoadingMode,
			ValueLogLoadingMode:     configOptions.Performance.ValueLogLoadingMode,
			BloomFalsePositive:      configOptions.Performance.BloomFalsePositive,
			DetectConflicts:         configOptions.Performance.DetectConflicts,
			ManageTxns:              configOptions.Performance.ManageTxns,
			MaxBatchCount:           configOptions.Performance.MaxBatchCount,
			MaxBatchSize:            configOptions.Performance.MaxBatchSize,
		}
	}

	return options
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

	logger.Debug("Waiting for parser response", logger.Int("timeout_ms", timeoutMs))
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		logger.Debug("Received parser response", logger.String("response_length", fmt.Sprintf("%d", len(response))))
		return response, nil
	case <-time.After(timeout):
		logger.Warn("Parser response timeout", logger.Int("timeout_ms", timeoutMs))
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

	logger.Debug("Waiting for jobs response", logger.Int("timeout_ms", timeoutMs))
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		logger.Debug("Received jobs response", logger.String("response_length", fmt.Sprintf("%d", len(response))))
		return response, nil
	case <-time.After(timeout):
		logger.Warn("Jobs response timeout", logger.Int("timeout_ms", timeoutMs))
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
