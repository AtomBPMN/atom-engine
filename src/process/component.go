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
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// CoreInterface defines core methods needed by process component
// Определяет методы core необходимые для процессного компонента
type CoreInterface interface {
	GetTimewheelComponentInterface() interface{} // Returns TimewheelComponentInterface
	GetJobsComponent() interface{}               // Returns JobsComponentInterface
	GetMessagesComponent() interface{}           // Returns MessagesComponentInterface
	GetExpressionComponent() interface{}         // Returns ExpressionComponentInterface
	GetIncidentsComponent() interface{}          // Returns IncidentsComponentInterface
	GetAuthComponent() interface{}               // Returns AuthComponentInterface
	SendMessage(componentName, messageJSON string) error
}

// ComponentInterface defines process component interface (legacy compatibility)
// Определяет интерфейс компонента процессов (для обратной совместимости)
type ComponentInterface interface {
	// Lifecycle
	Init() error
	Start() error
	Stop() error
	IsReady() bool

	// Process management
	StartProcessInstance(processKey string, variables map[string]interface{}) (*models.ProcessInstance, error)
	GetProcessInstanceStatus(instanceID string) (*models.ProcessInstance, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*models.ProcessInstance, error)

	// Token management
	GetActiveTokens(instanceID string) ([]*models.Token, error)
	GetTokensByProcessInstance(instanceID string) ([]*models.Token, error)
	GetAllTokens() ([]*models.Token, error)
	ExecuteToken(token *models.Token) error
	ContinueExecution(instanceID string) error
	UpdateToken(token *models.Token) error

	// Timer management
	CreateTimer(timerRequest *TimerRequest) error
	HandleTimerCallback(timerID, elementID, tokenID string) error
	CreateBoundaryTimer(timerRequest *TimerRequest) error
	CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error)
	LinkBoundaryTimerToToken(tokenID, timerID string) error
	CancelBoundaryTimersForToken(tokenID string) error

	// Job management
	HandleJobCallback(jobID, elementID, tokenID, status, errorMessage string, variables map[string]interface{}) error
	CancelJobForToken(tokenID string) error
	CancelJobByID(jobID string) error

	// Message management
	HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error
	HandleEngineMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error
	CheckBufferedMessages(messageName, correlationKey string) (*models.BufferedMessage, error)
	ProcessBufferedMessage(message *models.BufferedMessage, token *models.Token) error
	CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error
	DeleteMessageSubscription(subscriptionID string) error
	PublishMessage(messageName, correlationKey string, variables map[string]interface{}) (*models.MessageCorrelationResult, error)
	PublishMessageWithElementID(messageName, correlationKey, elementID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error)
	CorrelateMessage(messageName, correlationKey, processInstanceID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error)

	// Helper methods
	GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error)

	// Gateway synchronization methods
	SaveGatewaySyncState(state *models.GatewaySyncState) error
	LoadGatewaySyncState(gatewayID, processInstanceID string) (*models.GatewaySyncState, error)
	DeleteGatewaySyncState(gatewayID, processInstanceID string) error

	// Error boundary management
	RegisterErrorBoundary(subscription *ErrorBoundarySubscription)
	GetErrorBoundariesForToken(tokenID string) []*ErrorBoundarySubscription
	FindMatchingErrorBoundary(tokenID, errorCode string) *ErrorBoundarySubscription
	RemoveErrorBoundariesForToken(tokenID string)

	// Signal management
	SubscribeToSignal(signalName, tokenID, elementID string, cancelActivity bool, variables map[string]interface{}) error
	BroadcastSignal(signalName string, variables map[string]interface{}) error
	UnsubscribeSignalsByToken(tokenID string) error

	// Legacy compatibility (will be removed in future)
	GetJobsComponent() interface{}
	GetMessagesComponent() interface{}
	GetCore() CoreInterface
}

// TimerRequest represents timer creation request from flow element
// Представляет запрос создания таймера от flow элемента
type TimerRequest struct {
	ElementID         string `json:"element_id"`
	TokenID           string `json:"token_id"`
	ProcessInstanceID string `json:"process_instance_id"`
	ProcessKey        string `json:"process_key"`

	// Timer definitions (only one should be set)
	TimeDuration *string `json:"time_duration,omitempty"` // "PT30S"
	TimeDate     *string `json:"time_date,omitempty"`     // "2025-12-31T23:59:59Z"
	TimeCycle    *string `json:"time_cycle,omitempty"`    // "R3/PT20S"

	// Boundary timer specific metadata
	AttachedToRef  *string `json:"attached_to_ref,omitempty"` // Element ID this boundary timer is attached to
	CancelActivity *bool   `json:"cancel_activity,omitempty"` // Whether this is interrupting boundary timer
}

// ExecutionResult represents result of element execution
// Представляет результат выполнения элемента
type ExecutionResult struct {
	Success      bool                   `json:"success"`
	TokenUpdated bool                   `json:"token_updated"`
	NextElements []string               `json:"next_elements"`
	NewTokens    []*models.Token        `json:"new_tokens,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	WaitingFor   string                 `json:"waiting_for,omitempty"`
	Completed    bool                   `json:"completed"`
	Error        string                 `json:"error,omitempty"`

	// Timer request for intermediate catch events and boundary events
	TimerRequest *TimerRequest `json:"timer_request,omitempty"`

	// Timer callback context flag - indicates this execution is from timer callback
	// Флаг контекста timer callback - указывает что выполнение от timer callback
	IsTimerCallback bool `json:"is_timer_callback,omitempty"`
}

// ElementExecutor defines interface for BPMN element executors
// Определяет интерфейс для исполнителей BPMN элементов
type ElementExecutor interface {
	Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error)
	GetElementType() string
}

// Component represents the process execution component with SRP-compliant architecture
// Представляет компонент выполнения процессов с архитектурой соблюдающей SRP
type Component struct {
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

	// Error boundary management
	errorBoundaryRegistry *ErrorBoundaryRegistry

	// Signal management
	signalManager *SignalManager

	// Component state
	ready  bool
	ctx    context.Context
	cancel context.CancelFunc
}

// NewComponent creates new process component with SRP architecture
// Создает новый компонент процессов с SRP архитектурой
func NewComponent(storage storage.Storage) *Component {
	logger.Info("DEBUG: NewComponent called")
	ctx, cancel := context.WithCancel(context.Background())

	comp := &Component{
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

	// Initialize error boundary management
	comp.errorBoundaryRegistry = NewErrorBoundaryRegistry()

	// Initialize signal management
	comp.signalManager = NewSignalManager(comp)

	// Initialize core components
	logger.Info("DEBUG: About to create BPMNHelper")
	comp.bpmnHelper = NewBPMNHelper(storage)
	logger.Info("DEBUG: About to create Engine")
	comp.engine = NewEngine(storage, comp)
	logger.Info("DEBUG: Engine created successfully")

	return comp
}

// SetCore sets core interface for external dependencies
// Устанавливает интерфейс core для внешних зависимостей
func (c *Component) SetCore(core CoreInterface) {
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
func (c *Component) GetCore() CoreInterface {
	return c.core
}

// ComponentLifecycleInterface implementation
// Реализация ComponentLifecycleInterface

// Init initializes process component
// Инициализирует компонент процессов
func (c *Component) Init() error {
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
func (c *Component) Start() error {
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
func (c *Component) Stop() error {
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
func (c *Component) IsReady() bool {
	return c.ready && c.storage != nil && c.storage.IsReady()
}

// ProcessManagerInterface delegation
// Делегирование ProcessManagerInterface

func (c *Component) StartProcessInstance(processKey string, variables map[string]interface{}) (*models.ProcessInstance, error) {
	return c.processManager.StartProcessInstance(processKey, variables)
}

func (c *Component) GetProcessInstanceStatus(instanceID string) (*models.ProcessInstance, error) {
	return c.processManager.GetProcessInstanceStatus(instanceID)
}

func (c *Component) CancelProcessInstance(instanceID string, reason string) error {
	return c.processManager.CancelProcessInstance(instanceID, reason)
}

func (c *Component) ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*models.ProcessInstance, error) {
	return c.processManager.ListProcessInstances(statusFilter, processKeyFilter, limit)
}

// TokenManagerInterface delegation
// Делегирование TokenManagerInterface

func (c *Component) GetActiveTokens(instanceID string) ([]*models.Token, error) {
	return c.tokenManager.GetActiveTokens(instanceID)
}

func (c *Component) GetTokensByProcessInstance(instanceID string) ([]*models.Token, error) {
	return c.tokenManager.GetTokensByProcessInstance(instanceID)
}

func (c *Component) GetAllTokens() ([]*models.Token, error) {
	return c.storage.LoadAllTokens()
}

func (c *Component) ExecuteToken(token *models.Token) error {
	if !c.IsReady() {
		return fmt.Errorf("process component not ready")
	}
	return c.engine.ExecuteToken(token)
}

func (c *Component) ContinueExecution(instanceID string) error {
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

func (c *Component) CreateTimer(timerRequest *TimerRequest) error {
	return c.timerManager.CreateTimer(timerRequest)
}

func (c *Component) HandleTimerCallback(timerID, elementID, tokenID string) error {
	return c.timerManager.HandleTimerCallback(timerID, elementID, tokenID)
}

func (c *Component) CreateBoundaryTimer(timerRequest *TimerRequest) error {
	return c.timerManager.CreateBoundaryTimer(timerRequest)
}

func (c *Component) CreateBoundaryTimerWithID(timerRequest *TimerRequest) (string, error) {
	return c.timerManager.CreateBoundaryTimerWithID(timerRequest)
}

func (c *Component) LinkBoundaryTimerToToken(tokenID, timerID string) error {
	return c.timerManager.LinkBoundaryTimerToToken(tokenID, timerID)
}

func (c *Component) CancelBoundaryTimersForToken(tokenID string) error {
	return c.timerManager.CancelBoundaryTimersForToken(tokenID)
}

func (c *Component) GetBPMNProcessForToken(token *models.Token) (map[string]interface{}, error) {
	return c.timerManager.GetBPMNProcessForToken(token)
}

// Gateway synchronization methods implementation
// Реализация методов синхронизации шлюзов
func (c *Component) SaveGatewaySyncState(state *models.GatewaySyncState) error {
	return c.storage.SaveGatewaySyncState(state)
}

func (c *Component) LoadGatewaySyncState(gatewayID, processInstanceID string) (*models.GatewaySyncState, error) {
	return c.storage.LoadGatewaySyncState(gatewayID, processInstanceID)
}

func (c *Component) DeleteGatewaySyncState(gatewayID, processInstanceID string) error {
	return c.storage.DeleteGatewaySyncState(gatewayID, processInstanceID)
}

// JobCallbackManagerInterface delegation
// Делегирование JobCallbackManagerInterface

func (c *Component) HandleJobCallback(jobID, elementID, tokenID, status, errorMessage string, variables map[string]interface{}) error {
	return c.jobManager.HandleJobCallback(jobID, elementID, tokenID, status, errorMessage, variables)
}

func (c *Component) CancelJobForToken(tokenID string) error {
	// Check if jobManager supports job cancellation via CancelJobForToken
	if cancelMethod, ok := c.jobManager.(interface {
		CancelJobForToken(tokenID string) error
	}); ok {
		return cancelMethod.CancelJobForToken(tokenID)
	}

	// Fallback: delegate to jobManager directly - caller is responsible for validation
	if jobCallbacks, ok := c.jobManager.(*JobCallbacks); ok {
		// Load token to get jobID
		token, err := c.storage.LoadToken(tokenID)
		if err != nil {
			return fmt.Errorf("failed to load token %s: %w", tokenID, err)
		}

		if token.IsWaiting() && strings.HasPrefix(token.WaitingFor, "job:") {
			jobID := strings.TrimPrefix(token.WaitingFor, "job:")
			return jobCallbacks.cancelJob(jobID)
		}
	}

	return nil
}

func (c *Component) CancelJobByID(jobID string) error {
	// Direct job cancellation by ID - more efficient when jobID is already known
	if jobCallbacks, ok := c.jobManager.(*JobCallbacks); ok {
		return jobCallbacks.cancelJob(jobID)
	}

	return fmt.Errorf("job manager does not support job cancellation")
}

// MessageCallbackManagerInterface delegation
// Делегирование MessageCallbackManagerInterface

func (c *Component) HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error {
	return c.messageManager.HandleMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
}

func (c *Component) HandleEngineMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error {
	return c.engine.HandleMessageCallback(messageID, messageName, correlationKey, tokenID, variables)
}

func (c *Component) CheckBufferedMessages(messageName, correlationKey string) (*models.BufferedMessage, error) {
	return c.messageManager.CheckBufferedMessages(messageName, correlationKey)
}

func (c *Component) ProcessBufferedMessage(message *models.BufferedMessage, token *models.Token) error {
	return c.messageManager.ProcessBufferedMessage(message, token)
}

func (c *Component) CreateMessageSubscription(subscription *models.ProcessMessageSubscription) error {
	return c.messageManager.CreateMessageSubscription(subscription)
}

func (c *Component) DeleteMessageSubscription(subscriptionID string) error {
	return c.messageManager.DeleteMessageSubscription(subscriptionID)
}

func (c *Component) PublishMessage(messageName, correlationKey string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	return c.messageManager.PublishMessage(messageName, correlationKey, variables)
}

func (c *Component) PublishMessageWithElementID(messageName, correlationKey, elementID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	return c.messageManager.PublishMessageWithElementID(messageName, correlationKey, elementID, variables)
}

func (c *Component) CorrelateMessage(messageName, correlationKey, processInstanceID string, variables map[string]interface{}) (*models.MessageCorrelationResult, error) {
	return c.messageManager.CorrelateMessage(messageName, correlationKey, processInstanceID, variables)
}

// Legacy support methods for backward compatibility
// Методы для обратной совместимости

func (c *Component) GetJobsComponent() interface{} {
	if c.core == nil {
		return nil
	}
	return c.core.GetJobsComponent()
}

func (c *Component) GetMessagesComponent() interface{} {
	if c.core == nil {
		return nil
	}
	return c.core.GetMessagesComponent()
}

// main function for standalone component execution
// Основная функция для автономного выполнения компонента
func main() {
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

	// Initialize storage for autonomous operation
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

// ErrorBoundaryRegistry delegation
// Делегирование ErrorBoundaryRegistry

func (c *Component) RegisterErrorBoundary(subscription *ErrorBoundarySubscription) {
	c.errorBoundaryRegistry.RegisterErrorBoundary(subscription)
}

func (c *Component) GetErrorBoundariesForToken(tokenID string) []*ErrorBoundarySubscription {
	return c.errorBoundaryRegistry.GetErrorBoundariesForToken(tokenID)
}

func (c *Component) FindMatchingErrorBoundary(tokenID, errorCode string) *ErrorBoundarySubscription {
	return c.errorBoundaryRegistry.FindMatchingErrorBoundary(tokenID, errorCode)
}

func (c *Component) RemoveErrorBoundariesForToken(tokenID string) {
	c.errorBoundaryRegistry.RemoveErrorBoundariesForToken(tokenID)
}

// SubscribeToSignal subscribes a token to a signal
// Подписывает токен на сигнал
func (c *Component) SubscribeToSignal(signalName, tokenID, elementID string, cancelActivity bool, variables map[string]interface{}) error {
	if c.signalManager == nil {
		return fmt.Errorf("signal manager not initialized")
	}
	return c.signalManager.Subscribe(signalName, tokenID, elementID, cancelActivity, variables)
}

// BroadcastSignal broadcasts a signal to all subscribers
// Рассылает сигнал всем подписчикам
func (c *Component) BroadcastSignal(signalName string, variables map[string]interface{}) error {
	if c.signalManager == nil {
		return fmt.Errorf("signal manager not initialized")
	}
	return c.signalManager.BroadcastSignal(signalName, variables)
}

// UnsubscribeSignalsByToken removes all signal subscriptions for a token
// Удаляет все подписки на сигналы для токена
func (c *Component) UnsubscribeSignalsByToken(tokenID string) error {
	if c.signalManager == nil {
		return fmt.Errorf("signal manager not initialized")
	}
	return c.signalManager.UnsubscribeByToken(tokenID)
}

// UpdateToken updates token in storage
// Обновляет токен в storage
func (c *Component) UpdateToken(token *models.Token) error {
	if c.storage == nil {
		return fmt.Errorf("storage not available")
	}
	return c.storage.UpdateToken(token)
}
