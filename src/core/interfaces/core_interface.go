/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package interfaces

import (
	"context"
	"time"

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/models"
	"atom-engine/src/core/types"
	"atom-engine/src/storage"
)

// CoreInterface defines the unified interface that core must provide to all consumers
// Определяет единый интерфейс, который core должен предоставлять всем потребителям
type CoreInterface interface {
	// Storage operations
	// Операции с хранилищем
	GetStorageStatus() (*StorageStatusResponse, error)
	GetStorageInfo() (*StorageInfoResponse, error)

	// Component access - typed interfaces
	// Доступ к компонентам - типизированные интерфейсы
	GetProcessComponent() ProcessComponentInterface
	GetTimewheelComponent() TimewheelComponentInterface
	GetStorageComponent() StorageComponentInterface

	// Component access - legacy interface{} methods for REST handlers compatibility
	// Доступ к компонентам - legacy методы interface{} для совместимости с REST handlers
	GetMessagesComponent() interface{}
	GetJobsComponent() interface{}
	GetParserComponent() interface{}
	GetExpressionComponent() interface{}
	GetIncidentsComponent() interface{}
	GetAuthComponent() interface{}
	GetStorage() interface{}

	// Component access - strictly typed interfaces for new code
	// Доступ к компонентам - строго типизированные интерфейсы для нового кода
	GetProcessComponentTyped() ProcessComponentTypedInterface
	GetMessagesComponentTyped() MessagesComponentInterface
	GetJobsComponentTyped() JobsComponentInterface
	GetParserComponentTyped() ParserComponentInterface
	GetExpressionComponentTyped() ExpressionComponentInterface
	GetIncidentsComponentTyped() IncidentsComponentInterface
	GetAuthComponentTyped() AuthComponentInterface
	GetStorageTyped() StorageInterface

	// Timer management
	// Управление таймерами
	GetTimewheelStats() (*timewheelpb.GetTimeWheelStatsResponse, error)
	GetTimersList(statusFilter string, limit int32) (*timewheelpb.ListTimersResponse, error)

	// JSON Message Routing
	// Маршрутизация JSON сообщений
	SendMessage(componentName, messageJSON string) error

	// Response Handling
	// Обработка ответов
	WaitForParserResponse(timeoutMs int) (string, error)
	WaitForJobsResponse(timeoutMs int) (string, error)
	WaitForMessagesResponse(timeoutMs int) (string, error)
	WaitForIncidentsResponse(timeoutMs int) (string, error)
}

// CoreTypedInterface defines strongly typed system-wide methods
// Определяет строго типизированные системные методы
type CoreTypedInterface interface {
	CoreInterface

	// New strongly typed system-wide methods
	// Новые строго типизированные системные методы
	GetSystemStatus() (*types.SystemStatus, error)
	GetSystemInfo() (*types.SystemInfo, error)
	GetSystemMetrics() (*types.SystemMetrics, error)
	ListComponents(req *types.ComponentListRequest) (*types.ComponentListResponse, error)
	GetComponentStatus(componentName string) (*types.ComponentInfo, error)
	HealthCheck(req *types.ComponentHealthCheckRequest) (*types.ComponentHealthCheckResponse, error)

	// Strongly typed process operations
	// Строго типизированные операции с процессами
	StartProcessTyped(req *types.ProcessStartRequest) (*types.ProcessStartResponse, error)
	CancelProcessTyped(req *types.ProcessCancelRequest) (*types.ProcessCancelResponse, error)

	// REST API adapter methods
	// Методы адаптера для REST API
	GetProcessInfoForREST(instanceID string) (map[string]interface{}, error)

	// Strongly typed operations results
	// Строго типизированные результаты операций
	ExecuteOperation(operationName string, params types.Variables) (*types.OperationResult, error)
	ExecuteBatchOperation(operations []string, params []types.Variables) (*types.BatchOperationResult, error)

	// gRPC connection access for REST handlers
	// Доступ к gRPC соединению для REST обработчиков
	GetGRPCConnection() (interface{}, error)
}

// StorageStatusResponse represents storage status
// Представляет статус хранилища
type StorageStatusResponse struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	DatabasePath  string `json:"database_path"`
	Status        string `json:"status"`
	LastError     string `json:"last_error,omitempty"`
	ErrorCount    int    `json:"error_count"`
	Uptime        string `json:"uptime"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	LastOperation string `json:"last_operation"`
}

// StorageInfoResponse represents storage information
// Представляет информацию о хранилище
type StorageInfoResponse struct {
	DatabasePath   string            `json:"database_path"`
	DatabaseSize   int64             `json:"database_size"`
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	KeyCount       int               `json:"key_count"`
	TotalKeys      int64             `json:"total_keys"`
	LastCompaction string            `json:"last_compaction"`
	MemoryUsage    int64             `json:"memory_usage"`
	DiskUsage      int64             `json:"disk_usage"`
	Configuration  map[string]string `json:"configuration"`
	Statistics     map[string]int64  `json:"statistics"`
	Health         HealthInfo        `json:"health"`
}

// HealthInfo represents health information
// Представляет информацию о здоровье системы
type HealthInfo struct {
	Status      string `json:"status"`
	LastCheck   string `json:"last_check"`
	Errors      int    `json:"errors"`
	Warnings    int    `json:"warnings"`
	Uptime      string `json:"uptime"`
	Performance string `json:"performance"`
}

// TimewheelComponentInterface defines timewheel component interface
// Определяет интерфейс timewheel компонента
type TimewheelComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool)
}

// StorageComponentInterface defines storage component interface
// Определяет интерфейс storage компонента
type StorageComponentInterface interface {
	LoadAllTokens() ([]*models.Token, error)
	LoadTokensByState(state models.TokenState) ([]*models.Token, error)
	LoadToken(tokenID string) (*models.Token, error)
}

// ProcessComponentInterface defines basic process component interface
// Определяет базовый интерфейс process компонента
type ProcessComponentInterface interface {
	// Legacy methods for backward compatibility
	// Устаревшие методы для обратной совместимости
	StartProcessInstance(processKey string, variables map[string]interface{}) (*ProcessInstanceResult, error)
	GetProcessInstanceStatus(instanceID string) (*ProcessInstanceStatus, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*ProcessInstanceStatus, error)
	GetTokensByProcessInstance(instanceID string) ([]*models.Token, error)
	GetActiveTokens(instanceID string) ([]*models.Token, error)
}

// ProcessComponentTypedInterface defines strongly typed process methods
// Определяет строго типизированные методы process
type ProcessComponentTypedInterface interface {
	ProcessComponentInterface

	// New strongly typed methods
	// Новые строго типизированные методы
	StartProcessInstanceTyped(processKey string, variables types.ProcessVariables) (*types.ProcessInstanceDetails, error)
	GetProcessInstanceStatusTyped(instanceID string) (*types.ProcessInstanceDetails, error)
	ListProcessInstancesTyped(req *types.ProcessListRequest) (*types.ProcessListResponse, error)
	GetProcessStats() (*types.ProcessStats, error)
	GetTokensTyped(req *types.TokenListRequest) (*types.TokenListResponse, error)
	TraceProcessExecution(req *types.ProcessTraceRequest) (*types.ProcessTraceResponse, error)
}

// ProcessInstanceResult represents process instance creation result
// Представляет результат создания экземпляра процесса
type ProcessInstanceResult struct {
	InstanceID      string                 `json:"instance_id"`
	ProcessKey      string                 `json:"process_key"`
	ProcessID       string                 `json:"process_id"`
	ProcessName     string                 `json:"process_name"`
	Version         int32                  `json:"version"`
	Variables       map[string]interface{} `json:"variables"`
	Status          string                 `json:"status"`
	State           string                 `json:"state"`
	CurrentActivity string                 `json:"current_activity"`
	CreatedAt       string                 `json:"created_at"`
	StartedAt       int64                  `json:"started_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	CompletedAt     int64                  `json:"completed_at,omitempty"`
}

// ProcessInstanceStatus represents process instance status
// Представляет статус экземпляра процесса
type ProcessInstanceStatus struct {
	InstanceID      string                 `json:"instance_id"`
	ProcessKey      string                 `json:"process_key"`
	ProcessID       string                 `json:"process_id"`
	ProcessName     string                 `json:"process_name"`
	Status          string                 `json:"status"`
	State           string                 `json:"state"`
	CurrentActivity string                 `json:"current_activity"`
	Variables       map[string]interface{} `json:"variables"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	StartedAt       int64                  `json:"started_at"`
	CompletedAt     string                 `json:"completed_at,omitempty"`
}

// ProcessInstanceList represents list of process instances
// Представляет список экземпляров процессов
type ProcessInstanceList struct {
	Instances  []*ProcessInstanceStatus `json:"instances"`
	TotalCount int32                    `json:"total_count"`
	PageSize   int32                    `json:"page_size"`
	PageNumber int32                    `json:"page_number"`
}

// Token type alias for models.Token
// Псевдоним типа для models.Token
type Token = models.Token
type TokenState = models.TokenState

// Strict component interfaces to replace interface{}
// Строгие интерфейсы компонентов для замены interface{}

// MessagesComponentInterface defines basic messages component interface
// Определяет базовый интерфейс messages компонента
type MessagesComponentInterface interface {
	// JSON message processing methods
	// Методы обработки JSON сообщений
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string

	// Legacy typed methods for backward compatibility
	// Устаревшие типизированные методы для обратной совместимости
	PublishMessage(ctx context.Context, tenantID, messageName, correlationKey, elementID string, variables map[string]interface{}, ttl *time.Duration) (*models.MessageCorrelationResult, error)
	ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error)
	ListMessageSubscriptions(ctx context.Context, tenantID string, limit, offset int) ([]*models.ProcessMessageSubscription, error)
	CleanupExpiredMessages(ctx context.Context) (int, error)
}

// MessagesComponentTypedInterface defines strongly typed messages methods
// Определяет строго типизированные методы messages
type MessagesComponentTypedInterface interface {
	MessagesComponentInterface

	// New strongly typed methods
	// Новые строго типизированные методы
	PublishMessageTyped(ctx context.Context, req *types.MessagePublishRequest) (*types.MessagePublishResponse, error)
	ListBufferedMessagesTyped(ctx context.Context, req *types.BufferedMessageListRequest) (*types.BufferedMessageListResponse, error)
	ListMessageSubscriptionsTyped(ctx context.Context, req *types.MessageSubscriptionListRequest) (*types.MessageSubscriptionListResponse, error)
	CleanupExpiredMessagesTyped(ctx context.Context, req *types.MessageCleanupRequest) (*types.MessageCleanupResponse, error)
	GetMessageStats() (*types.MessageStats, error)
	TestMessageFunctionality(ctx context.Context, req *types.MessageTestRequest) (*types.MessageTestResponse, error)
}

// JobsComponentInterface defines basic jobs component interface
// Определяет базовый интерфейс jobs компонента
type JobsComponentInterface interface {
	// JSON message processing methods
	// Методы обработки JSON сообщений
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string

	// Legacy typed methods for backward compatibility
	// Устаревшие типизированные методы для обратной совместимости
	CreateJob(jobType, processInstanceID string, variables map[string]interface{}) (string, error)
}

// JobsComponentTypedInterface defines strongly typed jobs methods
// Определяет строго типизированные методы jobs
type JobsComponentTypedInterface interface {
	JobsComponentInterface

	// New strongly typed methods
	// Новые строго типизированные методы
	CreateJobTyped(req *types.JobCreateRequest) (*types.JobOperationResult, error)
	ActivateJobs(req *types.JobActivateRequest) (*types.JobActivateResponse, error)
	CompleteJob(req *types.JobCompleteRequest) (*types.JobOperationResult, error)
	FailJob(req *types.JobFailRequest) (*types.JobOperationResult, error)
	CancelJob(req *types.JobCancelRequest) (*types.JobOperationResult, error)
	ThrowJobError(req *types.JobThrowErrorRequest) (*types.JobOperationResult, error)
	ListJobs(req *types.JobListRequest) (*types.JobListResponse, error)
	GetJobInfo(jobKey string) (*types.JobInfo, error)
	GetJobStats() (*types.JobStats, error)
}

// ParserComponentInterface defines parser component interface
// Определяет интерфейс parser компонента
type ParserComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	DeleteBPMNProcess(processID string) error // Real signature without context
	// Note: Other methods return component-specific types
}

// ExpressionComponentInterface defines basic expression component interface
// Определяет базовый интерфейс expression компонента
type ExpressionComponentInterface interface {
	// Legacy methods for backward compatibility
	// Устаревшие методы для обратной совместимости
	EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	ParseRetries(retriesStr string) (int, error)
}

// ExpressionComponentTypedInterface defines strongly typed expression methods
// Определяет строго типизированные методы expression
type ExpressionComponentTypedInterface interface {
	ExpressionComponentInterface

	// New strongly typed methods
	// Новые строго типизированные методы
	EvaluateExpressionTyped(req *types.ExpressionRequest) (*types.ExpressionResult, error)
	EvaluateBatch(req *types.ExpressionBatchRequest) (*types.ExpressionBatchResult, error)
	ValidateExpression(req *types.ExpressionValidationRequest) (*types.ExpressionValidationResult, error)
	ExtractVariables(expression string) ([]string, error)
	GetExpressionStats() (*types.ExpressionStats, error)
	GetAvailableFunctions(category string) ([]*types.ExpressionFunctionInfo, error)
}

// IncidentsComponentInterface defines basic incidents component interface
// Определяет базовый интерфейс incidents компонента
type IncidentsComponentInterface interface {
	// JSON message processing methods
	// Методы обработки JSON сообщений
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
}

// IncidentsComponentTypedInterface defines strongly typed incidents methods
// Определяет строго типизированные методы incidents
type IncidentsComponentTypedInterface interface {
	IncidentsComponentInterface

	// New strongly typed methods using types.CoreError
	// Новые строго типизированные методы использующие types.CoreError
	CreateIncident(incident *types.CoreError) (string, error)
	ResolveIncident(incidentID string, retry bool, retries int32, comment string) (*types.OperationResult, error)
	ListIncidents(status, incidentType string, limit int32) ([]*types.CoreError, error)
	GetIncidentInfo(incidentID string) (*types.CoreError, error)
	GetIncidentsStats() (*types.ErrorStats, error)
}

// AuthComponentInterface defines basic auth component interface
// Определяет базовый интерфейс auth компонента
type AuthComponentInterface interface {
	// Basic component state methods
	// Основные методы состояния компонента
	IsEnabled() bool
	IsReady() bool
}

// AuthComponentTypedInterface defines strongly typed auth methods
// Определяет строго типизированные методы auth
type AuthComponentTypedInterface interface {
	AuthComponentInterface

	// New strongly typed methods for system monitoring
	// Новые строго типизированные методы для системного мониторинга
	GetComponentInfo() (*types.ComponentInfo, error)
	GetHealthStatus() (*types.ComponentHealthCheckResponse, error)
	GetComponentStats() (*types.ComponentMetrics, error)
}

// StorageInterface is a type alias for storage.Storage
// StorageInterface является псевдонимом типа для storage.Storage
type StorageInterface = storage.Storage

// Data transfer objects for strict typing
// Объекты передачи данных для строгой типизации

// Note: Most component-specific types are defined in their respective packages
// Большинство типов специфичных для компонентов определены в их собственных пакетах

// StorageInfo represents storage information
// Представляет информацию о хранилище
type StorageInfo struct {
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	TotalKeys      int64             `json:"total_keys"`
	DatabasePath   string            `json:"database_path"`
	Statistics     map[string]string `json:"statistics"`
}

// StorageStatus represents storage status
// Представляет статус хранилища
type StorageStatus struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}
