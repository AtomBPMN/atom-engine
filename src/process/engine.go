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
	"strconv"
	"strings"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// Engine represents the BPMN process execution engine
// Представляет движок выполнения BPMN процессов
type Engine struct {
	storage            storage.Storage
	component          ComponentInterface
	executorRegistry   *ExecutorRegistry
	executionProcessor *ExecutionProcessor
}

// NewEngine creates new process engine
// Создает новый движок процессов
func NewEngine(storage storage.Storage, component ComponentInterface) *Engine {
	engine := &Engine{
		storage:   storage,
		component: component,
	}

	// Initialize sub-components
	engine.executorRegistry = NewExecutorRegistry(component)
	engine.executionProcessor = NewExecutionProcessor(storage, component)

	// Register built-in element executors
	engine.executorRegistry.registerExecutors()

	return engine
}

// Init initializes process engine
// Инициализирует движок процессов
func (e *Engine) Init() error {
	logger.Info("Initializing process engine")

	if e.storage == nil {
		return fmt.Errorf("storage not provided")
	}

	logger.Info("Process engine initialized")
	return nil
}

// Start starts process engine
// Запускает движок процессов
func (e *Engine) Start() error {
	logger.Info("Starting process engine")
	logger.Info("Process engine started")
	return nil
}

// Stop stops process engine
// Останавливает движок процессов
func (e *Engine) Stop() error {
	logger.Info("Stopping process engine")
	logger.Info("Process engine stopped")
	return nil
}

// ExecuteToken executes token at current element
// Выполняет токен на текущем элементе
func (e *Engine) ExecuteToken(token *models.Token) error {
	logger.Info("=== EXECUTING TOKEN ===",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.String("token_state", string(token.State)),
		logger.String("process_instance_id", token.ProcessInstanceID))

	// Load process definition
	logger.Info("Loading process definition",
		logger.String("token_id", token.TokenID),
		logger.String("process_key", token.ProcessKey))

	processData, err := e.storage.LoadBPMNProcess(token.ProcessKey)
	if err != nil {
		logger.Error("Failed to load process definition",
			logger.String("process_key", token.ProcessKey),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to load process definition: %w", err)
	}

	logger.Info("Process definition loaded successfully",
		logger.String("process_key", token.ProcessKey),
		logger.Int("data_length", len(processData)))

	// DEBUG: Output raw JSON data from database
	logger.Debug("Raw process JSON from database",
		logger.String("process_key", token.ProcessKey),
		logger.String("json_data", string(processData)))

	var bpmnProcess models.BPMNProcess
	if err := json.Unmarshal(processData, &bpmnProcess); err != nil {
		logger.Error("Failed to parse process definition",
			logger.String("process_key", token.ProcessKey),
			logger.String("parse_error", err.Error()),
			logger.String("raw_json", string(processData)))
		return fmt.Errorf("failed to parse process definition: %w", err)
	}

	logger.Info("Process definition parsed",
		logger.String("process_key", token.ProcessKey),
		logger.String("process_id", bpmnProcess.ProcessID),
		logger.String("process_name", bpmnProcess.ProcessName),
		logger.Int("elements_count", len(bpmnProcess.Elements)))

	// DEBUG: Output all elements structure
	for elementID := range bpmnProcess.Elements {
		if elementData, err := json.Marshal(bpmnProcess.Elements[elementID]); err == nil {
			logger.Debug("Process element structure",
				logger.String("element_id", elementID),
				logger.String("element_json", string(elementData)))
		}
	}

	// Get current element
	logger.Info("Looking for element",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	element, exists := bpmnProcess.Elements[token.CurrentElementID]
	if !exists {
		// Debug: show all available elements
		availableElements := make([]string, 0, len(bpmnProcess.Elements))
		for elementID := range bpmnProcess.Elements {
			availableElements = append(availableElements, elementID)
		}
		logger.Error("Element not found in process",
			logger.String("element_id", token.CurrentElementID),
			logger.String("available_elements", fmt.Sprintf("%v", availableElements)))
		return fmt.Errorf("element not found: %s", token.CurrentElementID)
	}

	// Check if this is a sequence flow - handle it directly
	// Проверяем является ли это sequence flow - обрабатываем напрямую
	elementMap, ok := element.(map[string]interface{})
	if ok {
		elementType, typeExists := elementMap["type"].(string)
		if typeExists && elementType == "sequenceFlow" {
			// Handle sequence flow directly by getting target_ref
			// Обрабатываем sequence flow напрямую получая target_ref
			targetRef, targetExists := elementMap["target_ref"]
			if !targetExists {
				return fmt.Errorf("sequence flow missing target_ref: %s", token.CurrentElementID)
			}

			targetElementID, ok := targetRef.(string)
			if !ok {
				return fmt.Errorf("invalid target_ref format in sequence flow: %s", token.CurrentElementID)
			}
			logger.Info("Token moving via sequence flow to target element",
				logger.String("token_id", token.TokenID),
				logger.String("flow_id", token.CurrentElementID),
				logger.String("target_element", targetElementID))

			// Move token to target element and execute it
			token.MoveTo(targetElementID)
			if err := e.storage.UpdateToken(token); err != nil {
				return fmt.Errorf("failed to update token: %w", err)
			}
			return e.component.ExecuteToken(token)
		}
	}

	if !ok {
		return fmt.Errorf("invalid element structure: %s", token.CurrentElementID)
	}

	// Get element type
	elementType, typeExists := elementMap["type"].(string)
	if !typeExists {
		return fmt.Errorf("element type not found: %s", token.CurrentElementID)
	}

	// Find executor for element type
	executor, executorExists := e.executorRegistry.GetExecutor(elementType)
	if !executorExists {
		return fmt.Errorf("no executor found for element type: %s", elementType)
	}

	// Execute element
	result, err := executor.Execute(token, elementMap)
	if err != nil {
		logger.Error("Element execution failed",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("element_type", elementType),
			logger.String("error", err.Error()))

		// Cancel boundary timers before marking token as failed
		// Отменяем boundary таймеры перед отметкой токена как провалившегося
		if err := e.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
			logger.Error("Failed to cancel boundary timers for failed token",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
		}

		// Mark token as failed
		token.SetState(models.TokenStateFailed)
		if updateErr := e.storage.UpdateToken(token); updateErr != nil {
			logger.Error("Failed to update failed token", logger.String("error", updateErr.Error()))
		}

		return fmt.Errorf("element execution failed: %w", err)
	}

	// Process execution result
	if err := e.executionProcessor.processExecutionResult(token, result, &bpmnProcess); err != nil {
		return fmt.Errorf("failed to process execution result: %w", err)
	}

	logger.Info("=== TOKEN EXECUTION COMPLETED ===",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID),
		logger.String("element_type", elementType),
		logger.Bool("success", result.Success),
		logger.Bool("completed", result.Completed),
		logger.Int("next_elements_count", len(result.NextElements)),
		logger.String("waiting_for", result.WaitingFor))

	return nil
}

// RegisterExecutor registers element executor
// Регистрирует исполнитель элемента
func (e *Engine) RegisterExecutor(executor ElementExecutor) {
	e.executorRegistry.RegisterExecutor(executor)
}

// GetExecutor gets executor for element type
// Получает исполнитель для типа элемента
func (e *Engine) GetExecutor(elementType string) (ElementExecutor, bool) {
	return e.executorRegistry.GetExecutor(elementType)
}

// HandleMessageCallback handles message correlation callback
// Обрабатывает callback корреляции сообщения
func (e *Engine) HandleMessageCallback(messageID, messageName, correlationKey, tokenID string, variables map[string]interface{}) error {
	logger.Info("Engine handling message callback",
		logger.String("message_id", messageID),
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey),
		logger.String("token_id", tokenID))

	if e.storage == nil {
		return fmt.Errorf("storage not available")
	}

	// Check if this is Message Start Event callback (empty token_id)
	// Проверяем является ли это callback для Message Start Event (пустой token_id)
	if tokenID == "" {
		logger.Info("Message Start Event callback detected - creating new process instance",
			logger.String("message_id", messageID),
			logger.String("message_name", messageName))
		return e.handleMessageStartEventCallback(messageID, messageName, correlationKey, variables)
	}

	// Load the specific token that is waiting for this message (for intermediate catch events)
	// Загружаем конкретный токен который ожидает это сообщение (для intermediate catch events)
	token, err := e.storage.LoadToken(tokenID)
	if err != nil {
		logger.Error("Failed to load token for message callback",
			logger.String("message_id", messageID),
			logger.String("token_id", tokenID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to load token %s: %w", tokenID, err)
	}

	// Check if token is waiting for this message
	expectedWaitingFor := fmt.Sprintf("message:%s", messageName)
	if !token.IsWaiting() || token.WaitingFor != expectedWaitingFor {
		logger.Warn("Token is not waiting for this message",
			logger.String("message_id", messageID),
			logger.String("token_id", tokenID),
			logger.String("message_name", messageName),
			logger.String("token_state", string(token.State)),
			logger.String("token_waiting_for", token.WaitingFor),
			logger.String("expected_waiting_for", expectedWaitingFor))
		return fmt.Errorf("token %s is not waiting for message %s", tokenID, messageName)
	}

	logger.Info("Token confirmed waiting for message",
		logger.String("token_id", tokenID),
		logger.String("message_name", messageName))

	// Clear waiting state and merge message variables
	// Очищаем состояние ожидания и объединяем переменные сообщения
	token.ClearWaitingFor()
	if variables != nil {
		token.MergeVariables(variables)
		logger.Info("Message variables merged to token",
			logger.String("token_id", tokenID),
			logger.Any("variables", variables))
	}

	// Mark token as message correlated for future intermediate catch event detection
	// Отмечаем токен как активированный через message correlation для обнаружения в intermediate catch events
	if token.Variables == nil {
		token.Variables = make(map[string]interface{})
	}
	token.Variables["_message_correlated"] = true

	// Continue token execution from current element
	// Продолжаем выполнение токена с текущего элемента
	logger.Info("Continuing token execution after message correlation",
		logger.String("token_id", tokenID),
		logger.String("element_id", token.CurrentElementID))

	return e.ExecuteToken(token)
}

// handleMessageStartEventCallback handles Message Start Event callback
// Обрабатывает callback для Message Start Event
func (e *Engine) handleMessageStartEventCallback(messageID, messageName, correlationKey string, variables map[string]interface{}) error {
	logger.Info("Handling Message Start Event callback",
		logger.String("message_id", messageID),
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey))

	// Find subscription to get process key and start event ID
	// We need to access subscriptions through storage since we don't have direct messages component access
	// Нам нужно получить подписки через storage поскольку нет прямого доступа к messages component
	subscriptions, err := e.storage.ListProcessMessageSubscriptions(context.Background(), "", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var targetSubscription *models.ProcessMessageSubscription
	for _, sub := range subscriptions {
		if sub.MessageName == messageName && sub.IsActive {
			// For Message Start Events, we don't need exact correlation key match
			// since we're creating new instances, not finding existing tokens
			targetSubscription = sub
			break
		}
	}

	if targetSubscription == nil {
		return fmt.Errorf("no active subscription found for message %s", messageName)
	}

	logger.Info("Found subscription for Message Start Event",
		logger.String("subscription_id", targetSubscription.ID),
		logger.String("process_key", targetSubscription.ProcessDefinitionKey),
		logger.String("start_event_id", targetSubscription.StartEventID))

	// Create new process instance for Message Start Event
	// Создаем новый process instance для Message Start Event
	processInstance := models.NewProcessInstance(
		extractProcessIDFromKey(targetSubscription.ProcessDefinitionKey),
		"", // Process name will be loaded from definition
		extractVersionFromKey(targetSubscription.ProcessDefinitionKey),
		targetSubscription.ProcessDefinitionKey,
	)

	// Mark instance as active since it received trigger message
	// Отмечаем экземпляр как активный поскольку получил сообщение-триггер
	processInstance.State = models.ProcessInstanceStateActive

	// Set variables from message
	// Устанавливаем переменные из сообщения
	if variables != nil {
		processInstance.SetVariables(variables)
	}

	// Save process instance
	// Сохраняем process instance
	if err := e.storage.SaveProcessInstance(processInstance); err != nil {
		return fmt.Errorf("failed to save process instance: %w", err)
	}

	// Create initial token at start event
	// Создаем начальный токен на start event
	token := models.NewToken(processInstance.InstanceID, targetSubscription.ProcessDefinitionKey, targetSubscription.StartEventID)
	token.SetVariables(processInstance.Variables)

	logger.Info("Created new token for Message Start Event",
		logger.String("token_id", token.TokenID),
		logger.String("process_instance_id", processInstance.InstanceID),
		logger.String("start_event_id", targetSubscription.StartEventID))

	// Save token
	// Сохраняем токен
	if err := e.storage.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Execute token to start the process
	// Выполняем токен чтобы запустить процесс
	logger.Info("Starting Message Start Event process execution",
		logger.String("token_id", token.TokenID),
		logger.String("process_key", targetSubscription.ProcessDefinitionKey))

	return e.ExecuteToken(token)
}

// Helper functions
// Функции-помощники

// extractProcessIDFromKey extracts process ID from process key
func extractProcessIDFromKey(processKey string) string {
	if strings.Contains(processKey, ":v") {
		return strings.Split(processKey, ":v")[0]
	}
	return processKey
}

// extractVersionFromKey extracts version from process key
func extractVersionFromKey(processKey string) int {
	if strings.Contains(processKey, ":v") {
		parts := strings.Split(processKey, ":v")
		if len(parts) > 1 {
			if version, err := strconv.Atoi(parts[1]); err == nil {
				return version
			}
		}
	}
	return 1
}
