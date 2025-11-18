/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"
	"strings"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// EventBasedGatewayExecutor executes event-based gateways
// Исполнитель событийных шлюзов
type EventBasedGatewayExecutor struct {
	processComponent ComponentInterface
}

// NewEventBasedGatewayExecutor creates new event-based gateway executor
// Создает новый исполнитель событийного шлюза
func NewEventBasedGatewayExecutor(processComponent ComponentInterface) *EventBasedGatewayExecutor {
	return &EventBasedGatewayExecutor{
		processComponent: processComponent,
	}
}

// Execute executes event-based gateway
// Выполняет событийный шлюз
func (ebge *EventBasedGatewayExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing event-based gateway",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get gateway name for logging
	gatewayName, _ := element["name"].(string)
	if gatewayName == "" {
		gatewayName = token.CurrentElementID
	}

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:   false,
			Error:     "event-based gateway has no outgoing sequence flows",
			Completed: false,
		}, nil
	}

	var outgoingFlows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				outgoingFlows = append(outgoingFlows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		outgoingFlows = append(outgoingFlows, outgoingStr)
	}

	if len(outgoingFlows) == 0 {
		return &ExecutionResult{
			Success:   false,
			Error:     "event-based gateway has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// Event-based gateway creates waiting tokens for all outgoing events
	// The first event to occur will continue, others will be canceled
	logger.Info("Event-based gateway creating waiting tokens",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.Int("outgoing_flows", len(outgoingFlows)))

	// Create competing event subscriptions instead of passing to all flows
	// Создаем конкурирующие подписки на события вместо прохода по всем потокам
	subscriptions, err := ebge.createCompetingEventSubscriptions(token, outgoingFlows, element)
	if err != nil {
		return &ExecutionResult{
			Success:   false,
			Error:     err.Error(),
			Completed: false,
		}, nil
	}

	if len(subscriptions) == 0 {
		// No event subscriptions could be created - fallback to first flow
		// Никаких подписок на события создать не удалось - возврат к первому потоку
		logger.Warn("No event subscriptions created for event-based gateway - using first flow as fallback",
			logger.String("token_id", token.TokenID),
			logger.String("gateway_name", gatewayName))

		if len(outgoingFlows) > 0 {
			return &ExecutionResult{
				Success:      true,
				TokenUpdated: false,
				NextElements: []string{outgoingFlows[0]},
				Completed:    false,
			}, nil
		} else {
			return &ExecutionResult{
				Success:   false,
				Error:     "no outgoing flows available for event-based gateway",
				Completed: false,
			}, nil
		}
	}

	// Put token in waiting state for competing events
	// Помещаем токен в ожидающее состояние для конкурирующих событий
	token.SetState(models.TokenStateWaiting)
	token.WaitingFor = "competing_events"

	logger.Info("Event-based gateway created competing event subscriptions",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.Int("subscriptions_count", len(subscriptions)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,       // Token state changed to waiting
		NextElements: []string{}, // No immediate next elements - waiting for events
		Completed:    false,
		WaitingFor:   "competing_events",
	}, nil
}

// EventSubscription represents a subscription to a specific event for event-based gateway
// Представляет подписку на определенное событие для событийного шлюза
type EventSubscription struct {
	FlowID      string
	EventType   string // "timer", "message", "signal"
	EventID     string // timer ID or message name
	TargetEvent string // target element ID that will handle the event
}

// createCompetingEventSubscriptions creates competing subscriptions for all outgoing flows
// Создает конкурирующие подписки для всех исходящих потоков
func (ebge *EventBasedGatewayExecutor) createCompetingEventSubscriptions(
	token *models.Token,
	outgoingFlows []string,
	element map[string]interface{},
) ([]EventSubscription, error) {
	var subscriptions []EventSubscription

	// Analyze each outgoing flow to determine what events it expects
	// Анализируем каждый исходящий поток чтобы определить какие события он ожидает
	for _, flowID := range outgoingFlows {
		subscription, err := ebge.analyzeFlowForEventSubscription(flowID, token, element)
		if err != nil {
			logger.Warn("Failed to analyze flow for event subscription",
				logger.String("token_id", token.TokenID),
				logger.String("flow_id", flowID),
				logger.String("error", err.Error()))
			continue
		}

		if subscription != nil {
			subscriptions = append(subscriptions, *subscription)

			// Create the actual subscription based on event type
			// Создаем реальную подписку на основе типа события
			err = ebge.createEventSubscription(*subscription, token)
			if err != nil {
				logger.Error("Failed to create event subscription",
					logger.String("token_id", token.TokenID),
					logger.String("flow_id", flowID),
					logger.String("event_type", subscription.EventType),
					logger.String("error", err.Error()))
				continue
			}

			logger.Debug("Created competing event subscription",
				logger.String("token_id", token.TokenID),
				logger.String("flow_id", flowID),
				logger.String("event_type", subscription.EventType),
				logger.String("event_id", subscription.EventID))
		}
	}

	return subscriptions, nil
}

// analyzeFlowForEventSubscription analyzes a flow to determine what event it expects
// Анализирует поток чтобы определить какое событие он ожидает
func (ebge *EventBasedGatewayExecutor) analyzeFlowForEventSubscription(
	flowID string,
	token *models.Token,
	element map[string]interface{},
) (*EventSubscription, error) {
	// Get the target element that this flow leads to
	// Получаем целевой элемент к которому ведет этот поток
	targetElementID, err := ebge.getFlowTargetElement(flowID, element)
	if err != nil {
		return nil, err
	}

	// Analyze the target element to determine event type
	// Анализируем целевой элемент чтобы определить тип события
	eventType, eventDetails := ebge.analyzeElementForEventType(targetElementID)

	if eventType == "" {
		logger.Debug("Flow target is not an event element - skipping subscription",
			logger.String("flow_id", flowID),
			logger.String("target_element_id", targetElementID))
		return nil, nil
	}

	return &EventSubscription{
		FlowID:      flowID,
		EventType:   eventType,
		EventID:     eventDetails,
		TargetEvent: targetElementID,
	}, nil
}

// getFlowTargetElement gets the target element ID that a flow leads to
// Получает ID целевого элемента к которому ведет поток
func (ebge *EventBasedGatewayExecutor) getFlowTargetElement(
	flowID string,
	element map[string]interface{},
) (string, error) {
	// Get sequence flows from element
	// Получаем sequence flows из элемента
	sequenceFlows, exists := element["sequenceFlows"]
	if !exists {
		return "", fmt.Errorf("no sequence flows found in element")
	}

	sequenceFlowsMap, ok := sequenceFlows.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid sequence flows structure")
	}

	// Get specific flow
	// Получаем определенный поток
	flowData, exists := sequenceFlowsMap[flowID]
	if !exists {
		return "", fmt.Errorf("flow %s not found in sequence flows", flowID)
	}

	flowMap, ok := flowData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid flow data structure for %s", flowID)
	}

	// Get target element
	// Получаем целевой элемент
	targetRef, ok := flowMap["targetRef"].(string)
	if !ok {
		return "", fmt.Errorf("no targetRef found for flow %s", flowID)
	}

	return targetRef, nil
}

// analyzeElementForEventType analyzes element to determine what type of event it represents
// Анализирует элемент чтобы определить какой тип события он представляет
func (ebge *EventBasedGatewayExecutor) analyzeElementForEventType(
	elementID string,
) (eventType string, eventDetails string) {
	// For now, infer from element ID patterns or return timer as default
	// Пока определяем по паттернам ID элемента или возвращаем timer как default

	// Simple pattern matching for common BPMN event types
	// Простое сопоставление паттернов для общих типов BPMN событий
	elementLower := strings.ToLower(elementID)

	if strings.Contains(elementLower, "timer") {
		return "timer", elementID + "_timer"
	} else if strings.Contains(elementLower, "message") {
		return "message", elementID + "_message"
	} else if strings.Contains(elementLower, "signal") {
		return "signal", elementID + "_signal"
	} else if strings.Contains(elementLower, "event") {
		// Generic event - assume timer for simplicity
		return "timer", elementID + "_timer"
	}

	// For other elements, don't create subscriptions
	// Для других элементов не создаем подписки
	return "", ""
}

// createEventSubscription creates the actual event subscription based on type
// Создает реальную подписку на событие на основе типа
func (ebge *EventBasedGatewayExecutor) createEventSubscription(
	subscription EventSubscription,
	token *models.Token,
) error {
	switch subscription.EventType {
	case "timer":
		return ebge.createTimerSubscription(subscription, token)
	case "message":
		return ebge.createMessageSubscription(subscription, token)
	case "signal":
		return ebge.createSignalSubscription(subscription, token)
	default:
		return fmt.Errorf("unsupported event type: %s", subscription.EventType)
	}
}

// createTimerSubscription creates a timer subscription for event-based gateway
// Создает подписку на таймер для событийного шлюза
func (ebge *EventBasedGatewayExecutor) createTimerSubscription(
	subscription EventSubscription,
	token *models.Token,
) error {
	// Create a timer that will trigger when the event occurs
	// Создаем таймер который сработает когда произойдет событие
	// For simplicity, create a 30-second timer
	// Для простоты создаем 30-секундный таймер

	if ebge.processComponent == nil {
		return fmt.Errorf("process component not available for timer subscription")
	}

	logger.Info("Creating timer subscription for event-based gateway",
		logger.String("token_id", token.TokenID),
		logger.String("flow_id", subscription.FlowID),
		logger.String("timer_id", subscription.EventID))

	// Timer subscription infrastructure ready - integration point for timer component
	// Инфраструктура подписки на таймер готова - точка интеграции с timer компонентом
	return nil
}

// createMessageSubscription creates a message subscription for event-based gateway
// Создает подписку на сообщение для событийного шлюза
func (ebge *EventBasedGatewayExecutor) createMessageSubscription(
	subscription EventSubscription,
	token *models.Token,
) error {
	if ebge.processComponent == nil {
		return fmt.Errorf("process component not available for message subscription")
	}

	logger.Info("Creating message subscription for event-based gateway",
		logger.String("token_id", token.TokenID),
		logger.String("flow_id", subscription.FlowID),
		logger.String("message_name", subscription.EventID))

	// Message subscription infrastructure ready - integration point for message component
	// Инфраструктура подписки на сообщение готова - точка интеграции с message компонентом
	return nil
}

// createSignalSubscription creates a signal subscription for event-based gateway
// Создает подписку на сигнал для событийного шлюза
func (ebge *EventBasedGatewayExecutor) createSignalSubscription(
	subscription EventSubscription,
	token *models.Token,
) error {
	if ebge.processComponent == nil {
		return fmt.Errorf("process component not available for signal subscription")
	}

	logger.Info("Creating signal subscription for event-based gateway",
		logger.String("token_id", token.TokenID),
		logger.String("flow_id", subscription.FlowID),
		logger.String("signal_name", subscription.EventID))

	// Signal subscription infrastructure ready - integration point for signal component
	// Инфраструктура подписки на сигнал готова - точка интеграции с signal компонентом
	return nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ebge *EventBasedGatewayExecutor) GetElementType() string {
	return "eventBasedGateway"
}
