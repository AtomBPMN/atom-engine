/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// Component represents the process execution component with SRP-compliant architecture
// Представляет компонент выполнения процессов с архитектурой соблюдающей SRP
type ComponentNew struct {
	storage storage.Storage

	// Specialized managers implementing dedicated interfaces
	processManager ProcessManagerInterface
	tokenManager   TokenManagerInterface
	timerManager   TimerCallbackManagerInterface
	jobManager     JobCallbackManagerInterface
	messageManager MessageCallbackManagerInterface

	// Core components
	engine     *Engine
	bpmnHelper *BPMNHelper
	core       CoreInterface

	// Component state
	ready  bool
	ctx    context.Context
	cancel context.CancelFunc
}

// NewComponentNew creates new process component with SRP architecture
// Создает новый компонент процессов с SRP архитектурой
func NewComponentNew(storage storage.Storage) *ComponentNew {
	ctx, cancel := context.WithCancel(context.Background())

	comp := &ComponentNew{
		storage: storage,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Initialize specialized managers
	comp.processManager = NewProcessInstanceManager(storage, comp)
	comp.tokenManager = NewTokenManager(storage)
	comp.timerManager = NewUnifiedTimerManager(storage, comp)
	comp.jobManager = NewJobCallbacks(storage, comp)
	comp.messageManager = NewUnifiedMessageManager(storage, comp)

	// Initialize core components
	comp.bpmnHelper = NewBPMNHelper(storage)
	comp.engine = NewEngine(storage, comp)

	return comp
}

// SetCore sets core interface for external dependencies
// Устанавливает интерфейс core для внешних зависимостей
func (c *ComponentNew) SetCore(core CoreInterface) {
	c.core = core

	// Pass core to managers that need external dependencies
	if utm, ok := c.timerManager.(*UnifiedTimerManager); ok {
		utm.SetCore(core)
	}

	if umm, ok := c.messageManager.(*UnifiedMessageManager); ok {
		umm.SetCore(core)
	}

	if jcb, ok := c.jobManager.(*JobCallbacks); ok {
		jcb.SetCore(core)
	}
}

// GetCore returns core interface
// Возвращает интерфейс core
func (c *ComponentNew) GetCore() CoreInterface {
	return c.core
}

// ComponentLifecycleInterface implementation
// Реализация ComponentLifecycleInterface

// Init initializes process component
// Инициализирует компонент процессов
func (c *ComponentNew) Init() error {
	logger.Info("Initializing process component")

	if c.storage == nil {
		return fmt.Errorf("storage not provided")
	}

	// Initialize engine
	if err := c.engine.Init(); err != nil {
		return fmt.Errorf("failed to initialize engine: %w", err)
	}

	// Initialize managers through their concrete implementations
	// Инициализируем менеджеры через их конкретные реализации
	if tokenMgr, ok := c.tokenManager.(*TokenManager); ok {
		if err := tokenMgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize token manager: %w", err)
		}
	}

	if processMgr, ok := c.processManager.(*ProcessInstanceManager); ok {
		if err := processMgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize process instance manager: %w", err)
		}
	}

	if timerMgr, ok := c.timerManager.(*UnifiedTimerManager); ok {
		if err := timerMgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize timer manager: %w", err)
		}
	}

	if jobMgr, ok := c.jobManager.(*JobCallbacks); ok {
		if err := jobMgr.Init(); err != nil {
			return fmt.Errorf("failed to initialize job callbacks: %w", err)
		}
	}

	logger.Info("Process component initialized")
	return nil
}

// Start starts process component
// Запускает компонент процессов
func (c *ComponentNew) Start() error {
	logger.Info("Starting process component")

	// Start engine
	if err := c.engine.Start(); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}

	// Start token manager
	if tokenMgr, ok := c.tokenManager.(*TokenManager); ok {
		if err := tokenMgr.Start(); err != nil {
			return fmt.Errorf("failed to start token manager: %w", err)
		}
	}

	c.ready = true
	logger.Info("Process component started")

	// Restore active process instances and tokens AFTER component is ready
	if processMgr, ok := c.processManager.(*ProcessInstanceManager); ok {
		if err := processMgr.RestoreActiveProcesses(); err != nil {
			logger.Error("Failed to restore active processes", logger.String("error", err.Error()))
			// Don't fail startup, just log the error
		}
	}
	return nil
}

// Stop stops process component
// Останавливает компонент процессов
func (c *ComponentNew) Stop() error {
	logger.Info("Stopping process component")

	c.ready = false
	c.cancel()

	// Stop token manager
	if tokenMgr, ok := c.tokenManager.(*TokenManager); ok {
		if err := tokenMgr.Stop(); err != nil {
			logger.Error("Failed to stop token manager", logger.String("error", err.Error()))
		}
	}

	// Stop engine
	if err := c.engine.Stop(); err != nil {
		logger.Error("Failed to stop engine", logger.String("error", err.Error()))
	}

	logger.Info("Process component stopped")
	return nil
}

// IsReady returns component ready status
// Возвращает статус готовности компонента
func (c *ComponentNew) IsReady() bool {
	return c.ready && c.storage != nil && c.storage.IsReady()
}

// ProcessManagerInterface delegation
// Делегирование ProcessManagerInterface

func (c *ComponentNew) StartProcessInstance(processKey string, variables map[string]interface{}) (*models.ProcessInstance, error) {
	return c.processManager.StartProcessInstance(processKey, variables)
}

func (c *ComponentNew) GetProcessInstanceStatus(instanceID string) (*models.ProcessInstance, error) {
	return c.processManager.GetProcessInstanceStatus(instanceID)
}

func (c *ComponentNew) CancelProcessInstance(instanceID string, reason string) error {
	return c.processManager.CancelProcessInstance(instanceID, reason)
}

func (c *ComponentNew) ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*models.ProcessInstance, error) {
	return c.processManager.ListProcessInstances(statusFilter, processKeyFilter, limit)
}

// TokenManagerInterface delegation
// Делегирование TokenManagerInterface

func (c *ComponentNew) GetActiveTokens(instanceID string) ([]*models.Token, error) {
	return c.tokenManager.GetActiveTokens(instanceID)
}

func (c *ComponentNew) ExecuteToken(token *models.Token) error {
	if !c.IsReady() {
		return fmt.Errorf("process component not ready")
	}
	return c.engine.ExecuteToken(token)
}

func (c *ComponentNew) ContinueExecution(instanceID string) error {
	if !c.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	// Get active tokens and execute them
	activeTokens, err := c.tokenManager.GetActiveTokens(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get active tokens: %w", err)
	}

	// Execute each active token
	for _, token := range activeTokens {
		if err := c.ExecuteToken(token); err != nil {
			logger.Error("Failed to execute token during continuation",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
		}
	}

	return nil
}

// TimerCallbackManagerInterface delegation
// Делегирование TimerCallbackManagerInterface

func (c *ComponentNew) CreateTimer(timerRequest *TimerRequest) error {
	return c.timerManager.CreateTimer(timerRequest)
}

func (c *ComponentNew) HandleTimerCallback(timerID, elementID, tokenID string) error {
	return c.timerManager.HandleTimerCallback(timerID, elementID, tokenID)
}

func (c *ComponentNew) CreateBoundaryTimer(timerRequest *TimerRequest) error {
	return c.timerManager.CreateBoundaryTimer(timerRequest)
}

func (c *ComponentNew) CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error) {
	return c.timerManager.CreateBoundaryTimerWithID(timerRequest)
}

func (c *ComponentNew) LinkBoundaryTimerToToken(tokenID, timerID string) error {
	return c.timerManager.LinkBoundaryTimerToToken(tokenID, timerID)
}

func (c *ComponentNew) CancelBoundaryTimersForToken(tokenID string) error {
	return c.timerManager.CancelBoundaryTimersForToken(tokenID)
}

func (c *ComponentNew) GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error) {
	return c.timerManager.GetBPMNProcessForToken(token)
}

// JobCallbackManagerInterface delegation
// Делегирование JobCallbackManagerInterface

func (c *ComponentNew) HandleJobCallback(jobID, elementID, tokenID string, variables map[string]interface{}) error {
	return c.jobManager.HandleJobCallback(jobID, elementID, tokenID, variables)
}

// MessageCallbackManagerInterface delegation
// Делегирование MessageCallbackManagerInterface

func (c *ComponentNew) HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error {
	return c.messageManager.HandleMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
}

func (c *ComponentNew) CheckBufferedMessages(messageName, correlationKey string) (*models.BufferedMessage, error) {
	return c.messageManager.CheckBufferedMessages(messageName, correlationKey)
}

func (c *ComponentNew) ProcessBufferedMessage(message *models.BufferedMessage, token *models.Token) error {
	return c.messageManager.ProcessBufferedMessage(message, token)
}

func (c *ComponentNew) CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error {
	return c.messageManager.CreateMessageSubscription(subscription)
}

func (c *ComponentNew) DeleteMessageSubscription(subscriptionID string) error {
	return c.messageManager.DeleteMessageSubscription(subscriptionID)
}

func (c *ComponentNew) PublishMessage(messageName, correlationKey string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	return c.messageManager.PublishMessage(messageName, correlationKey, variables)
}

func (c *ComponentNew) CorrelateMessage(messageName, correlationKey, processInstanceID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	return c.messageManager.CorrelateMessage(messageName, correlationKey, processInstanceID, variables)
}

// Legacy support methods for backward compatibility
// Методы для обратной совместимости

func (c *ComponentNew) GetJobsComponent() interface{} {
	if c.core == nil {
		return nil
	}
	return c.core.GetJobsComponent()
}

func (c *ComponentNew) GetMessagesComponent() interface{} {
	if c.core == nil {
		return nil
	}
	return c.core.GetMessagesComponent()
}

// main function for standalone component execution
// Основная функция для автономного выполнения компонента
func mainNew() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: component <json_request>")
		os.Exit(1)
	}

	// Parse JSON request
	var request map[string]interface{}
	if err := json.Unmarshal([]byte(os.Args[1]), &request); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage (would need proper configuration in real usage)
	// This is a placeholder for autonomous component operation
	fmt.Println("Process component started in autonomous mode")

	// Process the request and return JSON response
	response := map[string]interface{}{
		"success":   true,
		"message":   "process component ready",
		"timestamp": time.Now().Unix(),
	}

	responseJSON, _ := json.Marshal(response)
	fmt.Println(string(responseJSON))
}
