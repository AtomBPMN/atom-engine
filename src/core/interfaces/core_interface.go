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
	DatabasePath    string            `json:"database_path"`
	DatabaseSize    int64             `json:"database_size"`
	TotalSizeBytes  int64             `json:"total_size_bytes"`
	UsedSizeBytes   int64             `json:"used_size_bytes"`
	FreeSizeBytes   int64             `json:"free_size_bytes"`
	KeyCount        int               `json:"key_count"`
	TotalKeys       int64             `json:"total_keys"`
	LastCompaction  string            `json:"last_compaction"`
	MemoryUsage     int64             `json:"memory_usage"`
	DiskUsage       int64             `json:"disk_usage"`
	Configuration   map[string]string `json:"configuration"`
	Statistics      map[string]int64  `json:"statistics"`
	Health          HealthInfo        `json:"health"`
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

// ProcessComponentInterface defines process component interface
// Определяет интерфейс process компонента
type ProcessComponentInterface interface {
	StartProcessInstance(processKey string, variables map[string]interface{}) (*ProcessInstanceResult, error)
	GetProcessInstanceStatus(instanceID string) (*ProcessInstanceStatus, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*ProcessInstanceStatus, error)
	GetTokensByProcessInstance(instanceID string) ([]*models.Token, error)
	GetActiveTokens(instanceID string) ([]*models.Token, error)
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

// MessagesComponentInterface defines messages component interface
// Определяет интерфейс messages компонента
type MessagesComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	PublishMessage(ctx context.Context, tenantID, messageName, correlationKey, elementID string, variables map[string]interface{}, ttl *time.Duration) (*models.MessageCorrelationResult, error)
	ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error)
	ListMessageSubscriptions(ctx context.Context, tenantID string, limit, offset int) ([]*models.ProcessMessageSubscription, error)
	CleanupExpiredMessages(ctx context.Context) (int, error)
	// Note: GetMessageStats returns component-specific type, not interfaces type
}

// JobsComponentInterface defines jobs component interface
// Определяет интерфейс jobs компонента
type JobsComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	CreateJob(jobType, processInstanceID string, variables map[string]interface{}) (string, error)
	// Note: Real methods return component-specific types, not interfaces types
}

// ParserComponentInterface defines parser component interface  
// Определяет интерфейс parser компонента
type ParserComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	DeleteBPMNProcess(processID string) error // Real signature without context
	// Note: Other methods return component-specific types
}

// ExpressionComponentInterface defines expression component interface
// Определяет интерфейс expression компонента
type ExpressionComponentInterface interface {
	EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	ParseRetries(retriesStr string) (int, error)
	// Note: Expression component does not have JSON message handling methods
}

// IncidentsComponentInterface defines incidents component interface
// Определяет интерфейс incidents компонента
type IncidentsComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	// Note: Component uses its own types for CreateIncidentRequest, Incident, etc.
}

// AuthComponentInterface defines auth component interface
// Определяет интерфейс auth компонента  
type AuthComponentInterface interface {
	IsEnabled() bool
	IsReady() bool
	// Note: Component uses its own AuthContext and AuthResult types
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
