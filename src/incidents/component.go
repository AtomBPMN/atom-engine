/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"context"
	"fmt"
	"os"

	"atom-engine/src/core/config"
	"atom-engine/src/core/logger"
	"atom-engine/src/storage"
)

// ComponentInterface defines the interface for incidents component
// Определяет интерфейс для компонента инцидентов
type ComponentInterface interface {
	// Component lifecycle
	Init() error
	Start() error
	Stop() error
	IsReady() bool

	// JSON communication interface
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string

	// Core incident operations interface (for other components)
	CreateIncident(ctx context.Context, request *CreateIncidentRequest) (*Incident, error)
	ResolveIncident(ctx context.Context, request *ResolveIncidentRequest) (*Incident, error)
	GetIncident(ctx context.Context, incidentID string) (*Incident, error)
	ListIncidents(ctx context.Context, filter *IncidentFilter) ([]*Incident, int, error)
	GetIncidentStats(ctx context.Context) (*IncidentStats, error)

	// Convenience methods for other components
	CreateJobFailureIncident(ctx context.Context, jobKey, elementID, processInstanceID, message string, retries int) (*Incident, error)
	CreateBPMNErrorIncident(ctx context.Context, elementID, processInstanceID, errorCode, message string) (*Incident, error)
	CreateExpressionErrorIncident(ctx context.Context, elementID, processInstanceID, expression, message string) (*Incident, error)
	CreateTimerErrorIncident(ctx context.Context, timerID, elementID, processInstanceID, message string) (*Incident, error)
	CreateMessageErrorIncident(ctx context.Context, messageName, correlationKey, processInstanceID, message string) (*Incident, error)
}

// Component represents the incidents component
// Представляет компонент инцидентов
type Component struct {
	config  *config.Config
	storage storage.Storage
	logger  logger.ComponentLogger

	// Component state
	ready  bool
	ctx    context.Context
	cancel context.CancelFunc

	// JSON communication channels
	requestChannel  chan string
	responseChannel chan string

	// Core manager
	manager IncidentManagerInterface
}

// NewComponent creates new incidents component
// Создает новый компонент инцидентов
func NewComponent(cfg *config.Config, storage storage.Storage) *Component {
	ctx, cancel := context.WithCancel(context.Background())

	return &Component{
		config:          cfg,
		storage:         storage,
		logger:          logger.NewComponentLogger("incidents"),
		ctx:             ctx,
		cancel:          cancel,
		requestChannel:  make(chan string, 100), // Buffered for async processing
		responseChannel: make(chan string, 100), // Buffered for responses
		manager:         NewIncidentManager(storage),
	}
}

// Init initializes incidents component
// Инициализирует компонент инцидентов
func (c *Component) Init() error {
	c.logger.Info("Initializing incidents component")

	if c.storage == nil {
		return fmt.Errorf("storage is required for incidents component")
	}

	if !c.storage.IsReady() {
		return fmt.Errorf("storage is not ready")
	}

	c.logger.Info("Incidents component initialized successfully")
	return nil
}

// Start starts incidents component
// Запускает компонент инцидентов
func (c *Component) Start() error {
	c.logger.Info("Starting incidents component")

	// Start JSON message processing goroutine
	go c.processMessages()

	c.ready = true
	c.logger.Info("Incidents component started successfully")
	return nil
}

// Stop stops incidents component
// Останавливает компонент инцидентов
func (c *Component) Stop() error {
	c.logger.Info("Stopping incidents component")

	c.ready = false
	c.cancel()

	// Close channels
	close(c.requestChannel)
	close(c.responseChannel)

	c.logger.Info("Incidents component stopped successfully")
	return nil
}

// IsReady returns component ready status
// Возвращает статус готовности компонента
func (c *Component) IsReady() bool {
	return c.ready && c.storage != nil && c.storage.IsReady()
}

// Component lifecycle and JSON communication methods are implemented above
// JSON message processing is in component_json.go
// Direct API methods are in component_api.go

// Start starts incidents component as separate service (CLI command)
// Запускает компонент инцидентов как отдельный сервис (команда CLI)
func Start() error {
	// Read configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	storageConfig := &storage.Config{
		Path: cfg.Storage.Directory,
	}
	storageInstance := storage.NewStorage(storageConfig)

	component := NewComponent(cfg, storageInstance)
	return component.Start()
}
