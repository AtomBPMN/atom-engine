/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"context"
	"time"

	"atom-engine/src/core/models"

	"github.com/dgraph-io/badger/v3"
)

// Storage interface defines storage operations
// Интерфейс Storage определяет операции storage
type Storage interface {
	Init() error
	Start() error
	Stop() error
	IsReady() bool
	LogSystemEvent(eventType, status, message string) error
	GetStatus() (*StorageStatus, error)
	GetInfo() (*StorageInfo, error)

	// Timer persistence methods
	// Методы персистентности таймеров
	SaveTimer(timer *TimerRecord) error
	LoadTimer(timerID string) (*TimerRecord, error)
	LoadAllTimers() ([]*TimerRecord, error)
	DeleteTimer(timerID string) error

	// BPMN persistence methods
	// Методы персистентности BPMN
	SaveBPMNProcess(processID string, data []byte) error
	LoadBPMNProcess(processID string) ([]byte, error)
	LoadBPMNProcessByProcessID(processID string, version int) ([]byte, string, error)
	LoadBPMNProcessByBPMNID(bpmnID string) ([]byte, error)
	LoadAllBPMNProcesses() (map[string][]byte, error)
	GetMaxProcessVersionByProcessID(processID string) (int, error)
	DeleteBPMNProcess(processID string) error
	// Note: SaveBPMNFile and LoadBPMNFile removed - XML files saved to filesystem only
	// Примечание: SaveBPMNFile и LoadBPMNFile удалены - XML файлы сохраняются только в файловую систему

	// Process Instance persistence methods
	// Методы персистентности экземпляров процессов
	SaveProcessInstance(instance *models.ProcessInstance) error
	LoadProcessInstance(instanceID string) (*models.ProcessInstance, error)
	LoadProcessInstancesByProcessKey(processKey string) ([]*models.ProcessInstance, error)
	LoadAllProcessInstances() ([]*models.ProcessInstance, error)
	UpdateProcessInstance(instance *models.ProcessInstance) error
	DeleteProcessInstance(instanceID string) error

	// Token persistence methods
	// Методы персистентности токенов
	SaveToken(token *models.Token) error
	LoadToken(tokenID string) (*models.Token, error)
	LoadTokensByProcessInstance(processInstanceID string) ([]*models.Token, error)
	LoadActiveTokens() ([]*models.Token, error)
	LoadTokensByState(state models.TokenState) ([]*models.Token, error)
	LoadAllTokens() ([]*models.Token, error)
	UpdateToken(token *models.Token) error
	DeleteToken(tokenID string) error

	// Job persistence methods
	// Методы персистентности заданий
	SaveJob(ctx context.Context, job *models.Job) error
	GetJob(ctx context.Context, jobID string) (*models.Job, error)
	ListJobsByType(ctx context.Context, jobType string, status models.JobStatus, limit int) ([]*models.Job, error)

	// Message persistence methods
	// Методы персистентности сообщений
	SaveProcessMessageSubscription(ctx context.Context, subscription *models.ProcessMessageSubscription) error
	GetProcessMessageSubscription(ctx context.Context, tenantID, processKey, startEventID string) (*models.ProcessMessageSubscription, error)
	ListProcessMessageSubscriptions(ctx context.Context, tenantID string, limit, offset int) ([]*models.ProcessMessageSubscription, error)
	DeleteProcessMessageSubscription(ctx context.Context, subscriptionID string) error
	SaveBufferedMessage(ctx context.Context, message *models.BufferedMessage) error
	GetBufferedMessage(ctx context.Context, messageID string) (*models.BufferedMessage, error)
	ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error)
	DeleteBufferedMessage(ctx context.Context, messageID string) error
	SaveMessageCorrelationResult(ctx context.Context, result *models.MessageCorrelationResult) error
	ListMessageCorrelationResults(ctx context.Context, tenantID, messageName, processKey string, limit, offset int) ([]*models.MessageCorrelationResult, error)
	DeleteMessageCorrelationResult(ctx context.Context, resultID string) error

	// Gateway synchronization persistence methods
	// Методы персистентности синхронизации шлюзов
	SaveGatewaySyncState(state *models.GatewaySyncState) error
	LoadGatewaySyncState(gatewayID, processInstanceID string) (*models.GatewaySyncState, error)
	DeleteGatewaySyncState(gatewayID, processInstanceID string) error

	// Incident persistence methods
	// Методы персистентности инцидентов
	SaveIncident(incident interface{}) error
	GetIncident(incidentID string) (interface{}, error)
	ListIncidents(filter interface{}) (interface{}, int, error)

	// Batch operations methods
	// Методы пакетных операций
	SaveBufferedMessagesBatch(ctx context.Context, messages []*models.BufferedMessage) error
	SaveTokensBatch(tokens []*models.Token) error
	DeleteMessagesBatch(ctx context.Context, messageIDs []string) error
	CleanupExpiredMessagesBatch(ctx context.Context, batchSize int) (int, error)
	GetBatchConfig() (maxBatchCount int, maxBatchSize int64)
}

// BadgerStorage implements Storage interface
// Реализация Storage для BadgerDB
type BadgerStorage struct {
	db        *badger.DB
	config    *Config
	ready     bool
	startTime time.Time
}

// Config holds database configuration
// Конфигурация базы данных
type Config struct {
	Path    string
	Options *StorageOptionsConfig
}

// StorageOptionsConfig holds storage options
// Настройки опций хранилища
type StorageOptionsConfig struct {
	SyncWrites       *bool
	ValueLogFileSize *int64
	Performance      *BadgerPerformanceConfig
}

// BadgerPerformanceConfig holds BadgerDB performance settings
// Настройки производительности BadgerDB
type BadgerPerformanceConfig struct {
	// Memory settings
	MemTableSize            *int64
	NumMemtables            *int
	NumLevelZeroTables      *int
	NumLevelZeroTablesStall *int

	// Cache settings
	ValueCacheSize *int64
	BlockCacheSize *int64
	IndexCacheSize *int64

	// Table and file settings
	BaseTableSize       *int64
	MaxTableSize        *int64
	LevelSizeMultiplier *int

	// Compaction settings
	NumCompactors    *int
	CompactL0OnClose *bool

	// I/O settings
	TableLoadingMode    *string
	ValueLogLoadingMode *string

	// Advanced settings
	BloomFalsePositive *float64
	DetectConflicts    *bool
	ManageTxns         *bool

	// Batch processing
	MaxBatchCount *int
	MaxBatchSize  *int64
}

// StorageStatus represents current status
// Представляет текущий статус
type StorageStatus struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// StorageInfo represents storage information
// Представляет информацию о storage
type StorageInfo struct {
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	TotalKeys      int64             `json:"total_keys"`
	DatabasePath   string            `json:"database_path"`
	Statistics     map[string]string `json:"statistics"`
}

// SystemEventRecord represents system event record
// Представляет запись системного события
type SystemEventRecord struct {
	ID        string    `json:"id"`
	EventType string    `json:"event_type"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TimerRecord represents timer in database
// Представляет таймер в базе данных
type TimerRecord struct {
	ID                string                 `json:"id"`
	ElementID         string                 `json:"element_id"`
	TokenID           string                 `json:"token_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	TimerType         string                 `json:"timer_type"`
	ScheduledAt       time.Time              `json:"scheduled_at"`
	TimeDate          *string                `json:"time_date,omitempty"`
	TimeDuration      *string                `json:"time_duration,omitempty"`
	TimeCycle         *string                `json:"time_cycle,omitempty"`
	ProcessContext    map[string]interface{} `json:"process_context,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	State             string                 `json:"state"` // SCHEDULED, FIRED, CANCELLED
}
