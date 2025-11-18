/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	"atom-engine/src/core/system"
	"atom-engine/src/core/types"
	"atom-engine/src/expression"
	"atom-engine/src/incidents"
	"atom-engine/src/jobs"
	"atom-engine/src/messages"
	"atom-engine/src/parser"
	"atom-engine/src/process"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
	"atom-engine/src/version"
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

	// Additional fields for typed interface implementation
	// Дополнительные поля для реализации typed интерфейса
	startTime      time.Time
	isShuttingDown bool

	// Message Multiplexer for jobs component
	// Message Multiplexer для jobs компонента
	jobsMultiplexer MessageMultiplexerInterface

	// CPU monitoring fields for sophisticated calculation
	// Поля мониторинга CPU для более точных вычислений
	lastCPUUpdate    time.Time
	lastUserTime     int64
	lastSystemTime   int64
	cachedCPUUsage   float64
	cpuCacheDuration time.Duration
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

		// Initialize typed interface fields
		// Инициализируем поля для typed интерфейса
		startTime:        time.Now(),
		isShuttingDown:   false,
		cpuCacheDuration: 5 * time.Second, // Cache CPU metrics for 5 seconds
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

// GetProcessComponentTyped returns typed process component
func (c *Core) GetProcessComponentTyped() interfaces.ProcessComponentTypedInterface {
	if c.processComp == nil {
		return nil
	}
	return &processComponentAdapter{comp: c.processComp}
}

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

// GetGRPCConnection returns gRPC connection for direct calls
// Возвращает gRPC соединение для прямых вызовов
func (c *Core) GetGRPCConnection() (interface{}, error) {
	if c.grpcServer == nil {
		return nil, fmt.Errorf("gRPC server is not available")
	}

	// Return a loopback connection to the gRPC server
	// Возвращаем loopback соединение к gRPC серверу
	conn, err := c.grpcServer.GetLoopbackConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC loopback connection: %w", err)
	}

	return conn, nil
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
	// Use Message Multiplexer if available
	// Используем Message Multiplexer если доступен
	if c.jobsMultiplexer != nil && c.jobsMultiplexer.IsRunning() {
		responseChannel := c.jobsMultiplexer.GetAPIResponseChannel()
		if responseChannel == nil {
			return "", fmt.Errorf("jobs API response channel not available")
		}

		logger.Debug("Waiting for jobs API response via multiplexer", logger.Int("timeout_ms", timeoutMs))
		timeout := time.Duration(timeoutMs) * time.Millisecond
		select {
		case response := <-responseChannel:
			logger.Debug("Received jobs API response via multiplexer",
				logger.String("response_length", fmt.Sprintf("%d", len(response))))
			return response, nil
		case <-time.After(timeout):
			logger.Warn("Jobs API response timeout via multiplexer", logger.Int("timeout_ms", timeoutMs))
			return "", fmt.Errorf("timeout waiting for jobs API response after %dms", timeoutMs)
		}
	}

	// Fallback to direct channel access (for backwards compatibility)
	// Резервный прямой доступ к каналу (для обратной совместимости)
	if c.jobsComp == nil {
		return "", fmt.Errorf("jobs component not available")
	}

	responseChannel := c.jobsComp.GetResponseChannel()
	if responseChannel == nil {
		return "", fmt.Errorf("jobs response channel not available")
	}

	logger.Debug("Waiting for jobs response (direct channel)", logger.Int("timeout_ms", timeoutMs))
	timeout := time.Duration(timeoutMs) * time.Millisecond
	select {
	case response := <-responseChannel:
		logger.Debug("Received jobs response (direct channel)",
			logger.String("response_length", fmt.Sprintf("%d", len(response))))
		return response, nil
	case <-time.After(timeout):
		logger.Warn("Jobs response timeout (direct channel)", logger.Int("timeout_ms", timeoutMs))
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

// CoreTypedInterface implementation
// Реализация CoreTypedInterface

// GetSystemStatus returns comprehensive system status
// Возвращает комплексный статус системы
func (c *Core) GetSystemStatus() (*types.SystemStatus, error) {
	if c.isShuttingDown {
		return nil, fmt.Errorf("system is shutting down")
	}

	now := time.Now()
	components := c.gatherComponentsInfo()

	componentsTotal := int32(len(components))
	componentsReady := int32(0)
	componentsError := int32(0)

	for _, comp := range components {
		if comp.IsReady() {
			componentsReady++
		}
		if comp.HasError() {
			componentsError++
		}
	}

	status := types.ComponentStatusRunning
	health := types.ComponentHealthHealthy

	if componentsError > 0 {
		health = types.ComponentHealthDegraded
		if componentsError > componentsTotal/2 {
			health = types.ComponentHealthUnhealthy
			status = types.ComponentStatusDegraded
		}
	}

	uptime := now.Sub(c.startTime)

	return &types.SystemStatus{
		Status:          status,
		Health:          health,
		Version:         version.Version,
		StartedAt:       c.startTime,
		Uptime:          uptime,
		Components:      components,
		ComponentsTotal: componentsTotal,
		ComponentsReady: componentsReady,
		ComponentsError: componentsError,
		LastHealthCheck: now,
		SystemMetrics:   c.gatherSystemMetrics(),
		Configuration:   c.getSystemConfiguration(),
	}, nil
}

// GetSystemInfo returns system information
// Возвращает информацию о системе
func (c *Core) GetSystemInfo() (*types.SystemInfo, error) {
	// Get real hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown" // Fallback if hostname cannot be determined
	}

	hostInfo := types.HostInfo{
		Hostname:     hostname, // Use real hostname
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCores:     int32(runtime.NumCPU()),
		MemoryTotal:  system.GetTotalMemory(),     // Use real total memory
		DiskTotal:    system.GetSystemDiskSpace(), // Use real disk space
	}

	return &types.SystemInfo{
		Name:          "Atom Engine",
		Version:       version.Version,        // Use real version from build info
		BuildTime:     version.GetBuildTime(), // Use real build time
		GitCommit:     version.GitCommit,      // Use real git commit
		Environment:   config.GetEnvWithDefault("ATOM_ENVIRONMENT", "development"),
		StartedAt:     c.startTime,
		Uptime:        time.Since(c.startTime),
		HostInfo:      hostInfo,
		Configuration: c.getSystemConfiguration(),
	}, nil
}

// GetSystemMetrics returns system metrics
// Возвращает системные метрики
func (c *Core) GetSystemMetrics() (*types.SystemMetrics, error) {
	metrics := c.gatherSystemMetrics()
	return &metrics, nil
}

// ListComponents returns list of components
// Возвращает список компонентов
func (c *Core) ListComponents(req *types.ComponentListRequest) (*types.ComponentListResponse, error) {
	components := c.gatherComponentsInfo()

	// Apply filters
	filtered := make([]types.ComponentInfo, 0)
	for _, comp := range components {
		if req.Type != nil && comp.Type != *req.Type {
			continue
		}
		if req.Status != nil && comp.Status != *req.Status {
			continue
		}
		if req.Health != nil && comp.Health != *req.Health {
			continue
		}
		if req.EnabledOnly && !comp.IsEnabled {
			continue
		}
		if req.ReadyOnly && !comp.ReadyFlag {
			continue
		}

		filtered = append(filtered, comp)
	}

	summary := types.ComponentSummary{
		Total:     int32(len(components)),
		ByStatus:  make(map[types.ComponentStatus]int32),
		ByHealth:  make(map[types.ComponentHealth]int32),
		ByType:    make(map[types.ComponentType]int32),
		Ready:     0,
		Enabled:   0,
		HasErrors: 0,
	}

	for _, comp := range components {
		summary.ByStatus[comp.Status]++
		summary.ByHealth[comp.Health]++
		summary.ByType[comp.Type]++

		if comp.ReadyFlag {
			summary.Ready++
		}
		if comp.IsEnabled {
			summary.Enabled++
		}
		if comp.HasError() {
			summary.HasErrors++
		}
	}

	return &types.ComponentListResponse{
		Components: filtered,
		TotalCount: int32(len(filtered)),
		Summary:    summary,
	}, nil
}

// GetComponentStatus returns component status
// Возвращает статус компонента
func (c *Core) GetComponentStatus(componentName string) (*types.ComponentInfo, error) {
	components := c.gatherComponentsInfo()

	for _, comp := range components {
		if comp.Name == componentName {
			return &comp, nil
		}
	}

	return nil, fmt.Errorf("component not found: %s", componentName)
}

// HealthCheck performs health check
// Выполняет проверку состояния
func (c *Core) HealthCheck(req *types.ComponentHealthCheckRequest) (*types.ComponentHealthCheckResponse, error) {
	start := time.Now()

	if req.ComponentName != "" {
		// Check specific component
		comp, err := c.GetComponentStatus(req.ComponentName)
		if err != nil {
			return &types.ComponentHealthCheckResponse{
				ComponentName: req.ComponentName,
				Health:        types.ComponentHealthUnknown,
				Status:        types.ComponentStatusError,
				Message:       err.Error(),
				CheckedAt:     time.Now(),
				Duration:      time.Since(start),
			}, nil
		}

		return &types.ComponentHealthCheckResponse{
			ComponentName: req.ComponentName,
			Health:        comp.Health,
			Status:        comp.Status,
			Message:       "Component health check completed",
			CheckedAt:     time.Now(),
			Duration:      time.Since(start),
		}, nil
	}

	// System-wide health check
	systemStatus, err := c.GetSystemStatus()
	if err != nil {
		return &types.ComponentHealthCheckResponse{
			Health:    types.ComponentHealthUnhealthy,
			Status:    types.ComponentStatusError,
			Message:   err.Error(),
			CheckedAt: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	return &types.ComponentHealthCheckResponse{
		Health:    systemStatus.Health,
		Status:    systemStatus.Status,
		Message:   "System health check completed",
		CheckedAt: time.Now(),
		Duration:  time.Since(start),
	}, nil
}

// StartProcessTyped starts process with typed request
// Запускает процесс с типизированным запросом
func (c *Core) StartProcessTyped(req *types.ProcessStartRequest) (*types.ProcessStartResponse, error) {
	if c.processComp == nil {
		return nil, fmt.Errorf("process component not available")
	}

	// Convert typed request to legacy format
	variables := make(map[string]interface{})
	for k, v := range req.Variables {
		variables[k] = v
	}

	result, err := c.processComp.StartProcessInstance(req.ProcessKey, variables)
	if err != nil {
		return &types.ProcessStartResponse{
			ProcessKey: req.ProcessKey,
			Success:    false,
			Message:    err.Error(),
			StartedAt:  time.Now(),
		}, err
	}

	return &types.ProcessStartResponse{
		InstanceID: result.InstanceID,
		ProcessKey: result.ProcessKey,
		Version:    int32(result.ProcessVersion),
		Status:     convertProcessInstanceState(result.State),
		Success:    true,
		Message:    "process started successfully",
		StartedAt:  result.StartedAt,
		Variables:  req.Variables,
	}, nil
}

// CancelProcessTyped cancels process with typed request
// Отменяет процесс с типизированным запросом
func (c *Core) CancelProcessTyped(req *types.ProcessCancelRequest) (*types.ProcessCancelResponse, error) {
	if c.processComp == nil {
		return nil, fmt.Errorf("process component not available")
	}

	err := c.processComp.CancelProcessInstance(req.InstanceID, req.Reason)
	if err != nil {
		return &types.ProcessCancelResponse{
			InstanceID:  req.InstanceID,
			Success:     false,
			Message:     err.Error(),
			CancelledAt: time.Now(),
		}, err
	}

	return &types.ProcessCancelResponse{
		InstanceID:  req.InstanceID,
		Success:     true,
		Message:     "process cancelled successfully",
		CancelledAt: time.Now(),
	}, nil
}

// ExecuteOperation executes generic operation
// Выполняет общую операцию
func (c *Core) ExecuteOperation(operationName string, params types.Variables) (*types.OperationResult, error) {
	start := time.Now()

	// Increment request count for operation tracking
	if err := c.IncrementRequestCount(); err != nil {
		logger.Warn("Failed to increment request count", logger.String("error", err.Error()))
	}

	result := &types.OperationResult{
		ExecutedAt: start,
		Metadata: map[string]interface{}{
			"operation": operationName,
			"params":    params,
		},
	}

	// Execute based on operation name
	switch operationName {
	case "system.status":
		status, err := c.GetSystemStatus()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to get system status: %v", err)
			result.Duration = time.Since(start)
			c.IncrementErrorCount()
			return result, err
		}
		result.Success = true
		result.Message = "System status retrieved successfully"
		result.Data = status

	case "system.info":
		info, err := c.GetSystemInfo()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to get system info: %v", err)
			result.Duration = time.Since(start)
			c.IncrementErrorCount()
			return result, err
		}
		result.Success = true
		result.Message = "System info retrieved successfully"
		result.Data = info

	case "system.metrics":
		metrics, err := c.GetSystemMetrics()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to get system metrics: %v", err)
			result.Duration = time.Since(start)
			c.IncrementErrorCount()
			return result, err
		}
		result.Success = true
		result.Message = "System metrics retrieved successfully"
		result.Data = metrics

	case "storage.status":
		storageStatus, err := c.GetStorageStatus()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to get storage status: %v", err)
			result.Duration = time.Since(start)
			c.IncrementErrorCount()
			return result, err
		}
		result.Success = true
		result.Message = "Storage status retrieved successfully"
		result.Data = storageStatus

	case "storage.info":
		storageInfo, err := c.GetStorageInfo()
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to get storage info: %v", err)
			result.Duration = time.Since(start)
			c.IncrementErrorCount()
			return result, err
		}
		result.Success = true
		result.Message = "Storage info retrieved successfully"
		result.Data = storageInfo

	default:
		result.Success = false
		result.Message = fmt.Sprintf("Unknown operation: %s", operationName)
		result.Data = map[string]interface{}{
			"available_operations": []string{
				"system.status",
				"system.info",
				"system.metrics",
				"storage.status",
				"storage.info",
			},
		}
		result.Duration = time.Since(start)
		c.IncrementErrorCount()
		return result, fmt.Errorf("unknown operation: %s", operationName)
	}

	result.Duration = time.Since(start)
	logger.Debug("Operation executed successfully",
		logger.String("operation", operationName),
		logger.String("duration", result.Duration.String()))

	return result, nil
}

// ExecuteBatchOperation executes batch operations
// Выполняет пакетные операции
func (c *Core) ExecuteBatchOperation(
	operations []string,
	params []types.Variables,
) (*types.BatchOperationResult, error) {
	start := time.Now()
	results := make([]types.OperationResult, 0, len(operations))
	successCount := int32(0)

	for i, op := range operations {
		var opParams types.Variables
		if i < len(params) {
			opParams = params[i]
		} else {
			opParams = make(types.Variables)
		}

		result, _ := c.ExecuteOperation(op, opParams)
		if result.Success {
			successCount++
		}
		results = append(results, *result)
	}

	return &types.BatchOperationResult{
		TotalCount:   int32(len(operations)),
		SuccessCount: successCount,
		FailureCount: int32(len(operations)) - successCount,
		Results:      results,
		ExecutedAt:   start,
		Duration:     time.Since(start),
	}, nil
}

// Helper methods for gathering system information
// Вспомогательные методы для сбора системной информации

func (c *Core) gatherComponentsInfo() []types.ComponentInfo {
	components := make([]types.ComponentInfo, 0)
	now := time.Now()

	// Storage component
	if c.storage != nil {
		comp := types.ComponentInfo{
			Name:        "storage",
			Type:        types.ComponentTypeStorage,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "BadgerDB storage component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	// Process component
	if c.processComp != nil {
		comp := types.ComponentInfo{
			Name:        "process",
			Type:        types.ComponentTypeProcess,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "BPMN process execution component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	// Messages component
	if c.messagesComp != nil {
		comp := types.ComponentInfo{
			Name:        "messages",
			Type:        types.ComponentTypeMessages,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "Message correlation component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	// Jobs component
	if c.jobsComp != nil {
		comp := types.ComponentInfo{
			Name:        "jobs",
			Type:        types.ComponentTypeJobs,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "Job management component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	// Expression component
	if c.expressionComp != nil {
		comp := types.ComponentInfo{
			Name:        "expression",
			Type:        types.ComponentTypeExpression,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "Expression evaluation component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	// Incidents component
	if c.incidentsComp != nil {
		comp := types.ComponentInfo{
			Name:        "incidents",
			Type:        types.ComponentTypeIncidents,
			Status:      types.ComponentStatusRunning,
			Health:      types.ComponentHealthHealthy,
			Description: "Incident management component",
			IsEnabled:   true,
			ReadyFlag:   true,
			StartedAt:   &c.startTime,
			Uptime:      &[]time.Duration{now.Sub(c.startTime)}[0],
		}
		components = append(components, comp)
	}

	return components
}

func (c *Core) gatherSystemMetrics() types.SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Load persistent metrics from storage
	persistedMetrics, err := c.storage.LoadSystemMetrics()
	if err != nil {
		logger.Warn("Failed to load system metrics from storage", logger.String("error", err.Error()))
		// Continue with default values if storage fails
		persistedMetrics = &storage.SystemMetrics{}
	}

	// Calculate CPU usage
	cpuUsage := c.calculateCPUUsage()

	// Update current memory usage in storage
	memoryUsage := int64(memStats.Alloc)
	if err := c.storage.UpdateMemoryUsage(memoryUsage); err != nil {
		logger.Warn("Failed to update memory usage in storage", logger.String("error", err.Error()))
	}

	// Update CPU usage in storage
	if err := c.storage.UpdateCPUUsage(cpuUsage); err != nil {
		logger.Warn("Failed to update CPU usage in storage", logger.String("error", err.Error()))
	}

	// Calculate error rate
	errorRate := float64(0)
	if persistedMetrics.TotalRequests > 0 {
		errorRate = float64(persistedMetrics.TotalErrors) / float64(persistedMetrics.TotalRequests) * 100
	}

	return types.SystemMetrics{
		TotalRequests:       persistedMetrics.TotalRequests,
		TotalErrors:         persistedMetrics.TotalErrors,
		ErrorRate:           errorRate,
		AverageResponseTime: persistedMetrics.AverageResponseTime,
		RequestsPerSecond:   persistedMetrics.RequestsPerSecond,
		MemoryUsage:         memoryUsage,
		CPUUsage:            cpuUsage,
		DiskUsage:           persistedMetrics.DiskUsage,
		NetworkIn:           persistedMetrics.NetworkIn,
		NetworkOut:          persistedMetrics.NetworkOut,
		ActiveConnections:   persistedMetrics.ActiveConnections,
		Goroutines:          int32(runtime.NumGoroutine()),
	}
}

func (c *Core) getSystemConfiguration() map[string]interface{} {
	config := make(map[string]interface{})

	if c.config != nil {
		config["instance_name"] = c.config.InstanceName
		config["grpc_port"] = c.config.GRPC.Port
		config["rest_port"] = c.config.RestAPI.Port
		config["storage_path"] = c.config.Database.Path
		config["log_level"] = c.config.Logger.Level
		config["base_path"] = c.config.BasePath
	}

	return config
}

// calculateCPUUsage calculates current CPU usage using sophisticated methods
// Вычисляет текущее использование CPU с использованием продвинутых методов
func (c *Core) calculateCPUUsage() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Use cached value if still fresh (within cache duration)
	// Используем кэшированное значение если оно еще свежее
	if !c.lastCPUUpdate.IsZero() && now.Sub(c.lastCPUUpdate) < c.cpuCacheDuration {
		return c.cachedCPUUsage
	}

	// Strategy 1: Try to read process CPU usage from runtime
	// Стратегия 1: Пытаемся получить CPU usage процесса из runtime
	if cpuPercent := c.calculateProcessCPUUsage(now); cpuPercent >= 0 {
		c.cachedCPUUsage = cpuPercent
		c.lastCPUUpdate = now
		return cpuPercent
	}

	// Strategy 2: Estimate based on runtime metrics and goroutines
	// Стратегия 2: Оценка на основе runtime метрик и горутин
	cpuUsage := c.estimateCPUFromRuntimeMetrics()

	c.cachedCPUUsage = cpuUsage
	c.lastCPUUpdate = now
	return cpuUsage
}

// calculateProcessCPUUsage calculates CPU usage based on process times
// Вычисляет CPU usage на основе времени процесса
func (c *Core) calculateProcessCPUUsage(now time.Time) float64 {
	// Get current process times
	// Получаем текущее время процесса
	userTime, systemTime := c.getProcessTimes()
	if userTime < 0 || systemTime < 0 {
		return -1 // Failed to get process times
	}

	// If this is first measurement, store times and return estimate
	// Если это первое измерение, сохраняем времена и возвращаем оценку
	if c.lastCPUUpdate.IsZero() {
		c.lastUserTime = userTime
		c.lastSystemTime = systemTime
		// Return initial estimate based on goroutines
		return c.estimateCPUFromRuntimeMetrics()
	}

	// Calculate time differences
	// Вычисляем разности времен
	timeDiff := now.Sub(c.lastCPUUpdate).Seconds()
	userDiff := float64(userTime-c.lastUserTime) / 1000000.0 // Convert from microseconds to seconds
	systemDiff := float64(systemTime-c.lastSystemTime) / 1000000.0

	if timeDiff <= 0 {
		return c.cachedCPUUsage // No time passed, return cached value
	}

	// Calculate CPU percentage: (total_cpu_time / elapsed_time) * 100
	// Вычисляем процент CPU: (общее_время_cpu / прошедшее_время) * 100
	totalCPUTime := userDiff + systemDiff
	cpuPercent := (totalCPUTime / timeDiff) * 100.0

	// Store current times for next calculation
	// Сохраняем текущие времена для следующего вычисления
	c.lastUserTime = userTime
	c.lastSystemTime = systemTime

	// Limit to reasonable bounds
	// Ограничиваем разумными пределами
	if cpuPercent < 0 {
		cpuPercent = 0
	} else if cpuPercent > 100 {
		cpuPercent = 100
	}

	return cpuPercent
}

// estimateCPUFromRuntimeMetrics estimates CPU usage from runtime metrics
// Оценивает CPU usage по runtime метрикам
func (c *Core) estimateCPUFromRuntimeMetrics() float64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	numGoroutines := runtime.NumGoroutine()
	numCPU := runtime.NumCPU()

	// Multi-factor estimation combining various metrics
	// Многофакторная оценка, комбинирующая различные метрики

	// Factor 1: Goroutine load relative to CPU cores
	// Фактор 1: Нагрузка горутин относительно ядер CPU
	goroutineLoad := float64(numGoroutines) / float64(numCPU) * 5.0
	if goroutineLoad > 50 {
		goroutineLoad = 50 // Cap goroutine contribution
	}

	// Factor 2: GC pressure indicates CPU activity
	// Фактор 2: Давление GC указывает на активность CPU
	gcLoad := float64(memStats.NumGC) * 0.1
	if gcLoad > 20 {
		gcLoad = 20 // Cap GC contribution
	}

	// Factor 3: Memory allocation rate indicates processing activity
	// Фактор 3: Скорость аллокации памяти указывает на активность обработки
	mallocLoad := float64(memStats.Mallocs-memStats.Frees) / 1000.0
	if mallocLoad > 20 {
		mallocLoad = 20 // Cap malloc contribution
	}

	// Factor 4: System uptime factor (busy systems tend to have steady load)
	// Фактор 4: Фактор времени работы системы
	uptime := time.Since(c.startTime).Seconds()
	uptimeFactor := math.Min(uptime/3600.0*2.0, 10.0) // 2% per hour, max 10%

	// Combine factors with weights
	// Комбинируем факторы с весами
	estimatedCPU := goroutineLoad*0.4 + gcLoad*0.2 + mallocLoad*0.2 + uptimeFactor*0.2

	// Apply system load factor based on number of CPUs
	// Применяем фактор системной нагрузки на основе количества CPU
	if numCPU > 0 {
		estimatedCPU = estimatedCPU / float64(numCPU) * 2.0
	}

	// Ensure reasonable bounds
	// Обеспечиваем разумные границы
	if estimatedCPU < 0 {
		estimatedCPU = 0
	} else if estimatedCPU > 95 {
		estimatedCPU = 95 // Leave some headroom
	}

	return estimatedCPU
}

// getProcessTimes returns user and system time for current process in microseconds
// Возвращает user и system время для текущего процесса в микросекундах
func (c *Core) getProcessTimes() (userTime, systemTime int64) {
	// Try to read from /proc/self/stat on Linux
	// Пытаемся прочитать из /proc/self/stat на Linux
	if userTime, systemTime := c.readLinuxProcessTimes(); userTime >= 0 {
		return userTime, systemTime
	}

	// Fallback: Use runtime GC stats as proxy for process activity
	// Запасной вариант: Используем GC статистику как приближение к активности процесса
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Use GC pause time as approximation of system activity
	// Используем время пауз GC как приближение к системной активности
	totalPauseNs := memStats.PauseTotalNs
	return int64(totalPauseNs / 1000), int64(totalPauseNs / 1000) // Convert to microseconds
}

// readLinuxProcessTimes reads process times from /proc/self/stat on Linux
// Читает времена процесса из /proc/self/stat на Linux
func (c *Core) readLinuxProcessTimes() (userTime, systemTime int64) {
	file, err := os.Open("/proc/self/stat")
	if err != nil {
		return -1, -1 // Not available on this system
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return -1, -1
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 15 {
		return -1, -1 // Invalid format
	}

	// Field 13: utime - CPU time spent in user mode (in clock ticks)
	// Field 14: stime - CPU time spent in kernel mode (in clock ticks)
	userTicks, err1 := strconv.ParseInt(fields[13], 10, 64)
	systemTicks, err2 := strconv.ParseInt(fields[14], 10, 64)

	if err1 != nil || err2 != nil {
		return -1, -1
	}

	// Convert clock ticks to microseconds
	// Clock ticks per second is typically 100 (USER_HZ)
	// Конвертируем clock ticks в микросекунды
	clockTicks := int64(100) // USER_HZ, typically 100 on Linux
	userTimeMicros := (userTicks * 1000000) / clockTicks
	systemTimeMicros := (systemTicks * 1000000) / clockTicks

	return userTimeMicros, systemTimeMicros
}

// IncrementRequestCount increments total request count
// Увеличивает общий счетчик запросов
func (c *Core) IncrementRequestCount() error {
	if c.storage != nil {
		return c.storage.IncrementRequestCount()
	}
	return nil
}

// IncrementErrorCount increments total error count
// Увеличивает общий счетчик ошибок
func (c *Core) IncrementErrorCount() error {
	if c.storage != nil {
		return c.storage.IncrementErrorCount()
	}
	return nil
}

// convertProcessInstanceState converts models.ProcessInstanceState to types.ProcessStatus
// Конвертирует models.ProcessInstanceState в types.ProcessStatus
func convertProcessInstanceState(state models.ProcessInstanceState) types.ProcessStatus {
	switch state {
	case models.ProcessInstanceStateActive:
		return types.ProcessStatusActive
	case models.ProcessInstanceStateCompleted:
		return types.ProcessStatusCompleted
	case models.ProcessInstanceStateCanceled:
		return types.ProcessStatusCancelled
	case models.ProcessInstanceStateFailed:
		return types.ProcessStatusFailed
	case models.ProcessInstanceStateSuspended:
		return types.ProcessStatusSuspended
	case models.ProcessInstanceStateMessages:
		return types.ProcessStatusActive // Messages state is still active
	default:
		return types.ProcessStatusActive // Default to active for unknown states
	}
}
