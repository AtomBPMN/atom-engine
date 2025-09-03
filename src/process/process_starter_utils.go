/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// handleMessageStartEvent handles Message Start Event startup
// Обрабатывает запуск Message Start Event
func (ps *ProcessStarter) handleMessageStartEvent(instance *models.ProcessInstance, bpmnProcess *models.BPMNProcess, startEventID, processKey string, variables map[string]interface{}) error {
	logger.Info("Process has Message Start Event - creating permanent subscription instead of starting",
		logger.String("process_id", instance.ProcessID),
		logger.String("start_event_id", startEventID),
		logger.String("storage_key", processKey))

	if err := ps.createMessageStartEventSubscription(bpmnProcess, startEventID, processKey, variables); err != nil {
		logger.Error("Failed to create Message Start Event subscription",
			logger.String("process_id", instance.ProcessID),
			logger.String("start_event_id", startEventID),
			logger.String("error", err.Error()))
		// Don't fail the command, subscription might already exist
	}

	// Don't create token or start process instance - just mark as waiting for messages
	// Не создаем токен или экземпляр процесса - просто отмечаем как ожидающий сообщения
	instance.State = models.ProcessInstanceStateMessages
	instance.Variables["_message_start_registered"] = true

	// Save updated instance state to storage
	// Сохраняем обновленное состояние процесса в storage
	if err := ps.storage.SaveProcessInstance(instance); err != nil {
		logger.Error("Failed to save process instance with MESSAGES state",
			logger.String("instance_id", instance.InstanceID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to save process instance: %w", err)
	}

	logger.Info("Message Start Event subscription created, process registered for messages",
		logger.String("process_id", instance.ProcessID),
		logger.String("instance_id", instance.InstanceID),
		logger.String("state", string(instance.State)),
		logger.String("storage_key", processKey))

	return nil
}

// handleRegularStartEvent handles regular start event startup
// Обрабатывает запуск обычного стартового события
func (ps *ProcessStarter) handleRegularStartEvent(instance *models.ProcessInstance, processKey, startEventID string) error {
	logger.Info("Creating initial token for regular start event",
		logger.String("instance_id", instance.InstanceID),
		logger.String("process_key", processKey),
		logger.String("start_event_id", startEventID))

	token := models.NewToken(instance.InstanceID, processKey, startEventID)
	token.SetVariables(instance.Variables) // Copy process variables to token

	logger.Info("Initial token created",
		logger.String("token_id", token.TokenID),
		logger.String("process_key", token.ProcessKey),
		logger.String("element_id", token.CurrentElementID))

	// Save token
	if err := ps.storage.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save initial token: %w", err)
	}

	// Create collaboration message subscriptions for subprocess Message Start Events
	// Создаем collaboration подписки для Message Start Events в подпроцессах
	bpmnProcess, err := ps.bpmnHelper.LoadBPMNProcess(processKey)
	if err != nil {
		logger.Warn("Failed to load BPMN process for collaboration subscriptions",
			logger.String("error", err.Error()))
	} else {
		if err := ps.collaborationManager.CreateCollaborationMessageSubscriptions(bpmnProcess, processKey); err != nil {
			logger.Warn("Failed to create collaboration message subscriptions",
				logger.String("error", err.Error()))
			// Don't fail the start, continue without collaboration subscriptions
		}
	}

	// Execute token to start the process
	if err := ps.component.ExecuteToken(token); err != nil {
		logger.Error("Failed to execute initial token", logger.String("error", err.Error()))
		// Don't fail the start, the token will be processed later
	}

	return nil
}

// findStartEvent finds start event in process definition
// Находит стартовое событие в определении процесса
func (ps *ProcessStarter) findStartEvent(bpmnProcess *models.BPMNProcess) (string, error) {
	// First, try to find the target process in elements
	var targetProcessFlowNodes []string

	// Look for the main process by processID
	for _, element := range bpmnProcess.Elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "process" {
				if processID, exists := elementMap["id"]; exists && processID == bpmnProcess.ProcessID {
					// Found our target process, get its flow nodes
					if flowNodeIds, exists := elementMap["flow_node_ids"]; exists {
						if flowNodeList, ok := flowNodeIds.([]interface{}); ok {
							for _, nodeID := range flowNodeList {
								if nodeIDStr, ok := nodeID.(string); ok {
									targetProcessFlowNodes = append(targetProcessFlowNodes, nodeIDStr)
								}
							}
						}
					}
					break
				}
			}
		}
	}

	logger.Info("Finding start event",
		logger.String("process_id", bpmnProcess.ProcessID),
		logger.Int("flow_nodes_count", len(targetProcessFlowNodes)),
		logger.Any("flow_nodes", targetProcessFlowNodes))

	// Search within target process flow nodes
	if len(targetProcessFlowNodes) > 0 {
		for _, flowNodeID := range targetProcessFlowNodes {
			if element, exists := bpmnProcess.Elements[flowNodeID]; exists {
				if elementMap, ok := element.(map[string]interface{}); ok {
					if elementType, exists := elementMap["type"]; exists && elementType == "startEvent" {
						logger.Info("Found start event for target process",
							logger.String("start_event_id", flowNodeID),
							logger.String("process_id", bpmnProcess.ProcessID))
						return flowNodeID, nil
					}
				}
			}
		}
	}

	// Fallback: search all elements
	logger.Warn("No process-specific flow nodes found, searching all elements",
		logger.String("process_id", bpmnProcess.ProcessID))

	for elementID, element := range bpmnProcess.Elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "startEvent" {
				logger.Info("Found start event (fallback search)",
					logger.String("start_event_id", elementID))
				return elementID, nil
			}
		}
	}

	return "", fmt.Errorf("no start event found in process")
}

// isMessageStartEvent checks if start event is Message Start Event
// Проверяет является ли start event Message Start Event
func (ps *ProcessStarter) isMessageStartEvent(bpmnProcess *models.BPMNProcess, startEventID string) (bool, error) {
	startElement, exists := bpmnProcess.Elements[startEventID]
	if !exists {
		return false, fmt.Errorf("start event not found: %s", startEventID)
	}

	elementMap, ok := startElement.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid start event structure")
	}

	// Check for event definitions
	if eventDefinitions, hasEventDefs := elementMap["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					if eventType, exists := eventDefMap["type"]; exists && eventType == "messageEventDefinition" {
						logger.Info("Found Message Start Event",
							logger.String("start_event_id", startEventID))
						return true, nil
					}
				}
			}
		}
	}

	logger.Info("Regular Start Event (not message)",
		logger.String("start_event_id", startEventID))
	return false, nil
}

// createMessageStartEventSubscription creates permanent subscription for Message Start Event
// Создает постоянную подписку для Message Start Event
func (ps *ProcessStarter) createMessageStartEventSubscription(bpmnProcess *models.BPMNProcess, startEventID, processKey string, variables map[string]interface{}) error {
	startElement, exists := bpmnProcess.Elements[startEventID]
	if !exists {
		return fmt.Errorf("start event not found: %s", startEventID)
	}

	elementMap, ok := startElement.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid start event structure")
	}

	// Extract message information from start event
	messageName, correlationKey, messageRef := ps.extractMessageInfo(elementMap)

	if messageRef == "" {
		return fmt.Errorf("no messageRef found in Message Start Event: %s", startEventID)
	}

	// Resolve message name and correlation key from message definition
	if messageName == "" || correlationKey == "" {
		resolvedName, resolvedKey := ps.resolveMessageData(bpmnProcess, messageRef)
		if messageName == "" {
			messageName = resolvedName
		}
		if correlationKey == "" {
			correlationKey = resolvedKey
		}
	}

	if messageName == "" {
		messageName = messageRef // Fallback to using messageRef as name
	}

	logger.Info("Creating permanent Message Start Event subscription",
		logger.String("start_event_id", startEventID),
		logger.String("message_ref", messageRef),
		logger.String("message_name", messageName),
		logger.String("correlation_key", correlationKey),
		logger.String("process_key", processKey))

	// Create permanent subscription
	subscription := &models.ProcessMessageSubscription{
		ID:                   models.GenerateID(),
		TenantID:             "", // TODO: Get tenant from context
		ProcessDefinitionKey: processKey,
		StartEventID:         startEventID,
		MessageName:          messageName,
		CorrelationKey:       correlationKey,
		IsActive:             true,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := ps.component.CreateMessageSubscription(subscription); err != nil {
		return fmt.Errorf("failed to create message subscription: %w", err)
	}

	logger.Info("Permanent Message Start Event subscription created",
		logger.String("subscription_id", subscription.ID),
		logger.String("process_key", processKey),
		logger.String("message_name", messageName))

	return nil
}

// extractMessageInfo extracts message information from start event element
// Извлекает информацию о сообщении из элемента стартового события
func (ps *ProcessStarter) extractMessageInfo(elementMap map[string]interface{}) (string, string, string) {
	messageName := ""
	correlationKey := ""
	messageRef := ""

	if eventDefinitions, hasEventDefs := elementMap["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					if eventType, exists := eventDefMap["type"]; exists && eventType == "messageEventDefinition" {
						// Extract messageRef
						if reference, exists := eventDefMap["reference"]; exists {
							if refStr, ok := reference.(string); ok {
								messageRef = refStr
							}
						} else if msgRef, exists := eventDefMap["message_ref"]; exists {
							if msgRefStr, ok := msgRef.(string); ok {
								messageRef = msgRefStr
							}
						}
						break
					}
				}
			}
		}
	}

	return messageName, correlationKey, messageRef
}

// resolveMessageData resolves message name and correlation key from message definition
// Разрешает имя сообщения и correlation key из определения сообщения
func (ps *ProcessStarter) resolveMessageData(bpmnProcess *models.BPMNProcess, messageRef string) (string, string) {
	messageName := ""
	correlationKey := ""

	// Find message definition by ID
	if element, exists := bpmnProcess.Elements[messageRef]; exists {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "message" {
				// Extract message name
				if name, exists := elementMap["name"]; exists {
					if nameStr, ok := name.(string); ok {
						messageName = nameStr
					}
				}

				// Extract correlation key from extension elements
				if extensionElements, exists := elementMap["extension_elements"]; exists {
					correlationKey = ps.extractCorrelationKeyFromExtensions(extensionElements)
				}

				logger.Info("Message data resolved for Message Start Event",
					logger.String("message_ref", messageRef),
					logger.String("message_name", messageName),
					logger.String("correlation_key", correlationKey))
			}
		}
	}

	return messageName, correlationKey
}

// extractCorrelationKeyFromExtensions extracts correlation key from extension elements
// Извлекает correlation key из extension elements
func (ps *ProcessStarter) extractCorrelationKeyFromExtensions(extensionElements interface{}) string {
	if extList, ok := extensionElements.([]interface{}); ok {
		for _, ext := range extList {
			if extMap, ok := ext.(map[string]interface{}); ok {
				if extensions, exists := extMap["extensions"]; exists {
					if extensionsList, ok := extensions.([]interface{}); ok {
						for _, extension := range extensionsList {
							if extensionMap, ok := extension.(map[string]interface{}); ok {
								if extensionType, exists := extensionMap["type"]; exists && extensionType == "subscription" {
									if attributes, exists := extensionMap["attributes"]; exists {
										if attrsMap, ok := attributes.(map[string]interface{}); ok {
											if corrKey, exists := attrsMap["correlationKey"]; exists {
												if corrKeyStr, ok := corrKey.(string); ok {
													return corrKeyStr
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}
