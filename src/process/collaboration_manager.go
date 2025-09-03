/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// CollaborationManager handles collaboration message subscriptions
// Управляет collaboration подписками на сообщения
type CollaborationManager struct {
	storage    storage.Storage
	component  ComponentInterface
	bpmnHelper *BPMNHelper
}

// NewCollaborationManager creates new collaboration manager
// Создает новый менеджер collaboration
func NewCollaborationManager(storage storage.Storage, component ComponentInterface) *CollaborationManager {
	return &CollaborationManager{
		storage:    storage,
		component:  component,
		bpmnHelper: NewBPMNHelper(storage),
	}
}

// CreateCollaborationMessageSubscriptions creates subscriptions for Message Start Events
// in all processes within the same BPMN file (collaboration)
// Создает подписки для Message Start Events во всех процессах того же BPMN файла (collaboration)
func (cm *CollaborationManager) CreateCollaborationMessageSubscriptions(bpmnProcess *models.BPMNProcess, currentProcessKey string) error {
	logger.Info("Creating collaboration message subscriptions",
		logger.String("current_process_key", currentProcessKey))

	// Find all processes in the BPMN definition
	allProcesses := cm.findAllProcesses(bpmnProcess, currentProcessKey)

	logger.Info("Found collaboration processes",
		logger.Int("process_count", len(allProcesses)))

	// For each process, find Message Start Events and create subscriptions
	for _, processElement := range allProcesses {
		if err := cm.createSubscriptionsForProcess(bpmnProcess, processElement, currentProcessKey); err != nil {
			logger.Warn("Failed to create subscriptions for collaboration process",
				logger.String("error", err.Error()))
			// Continue with other processes even if one fails
		}
	}

	return nil
}

// findAllProcesses finds all executable processes in BPMN definition except current
// Находит все executable процессы в BPMN определении кроме текущего
func (cm *CollaborationManager) findAllProcesses(bpmnProcess *models.BPMNProcess, currentProcessKey string) []map[string]interface{} {
	var allProcesses []map[string]interface{}

	for elementID, element := range bpmnProcess.Elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "process" {
				// Check if this is executable and not the current process
				if isExecutable, exists := elementMap["is_executable"]; exists {
					if executable, ok := isExecutable.(bool); ok && executable {
						// Create process key for this subprocess
						subprocessKey := fmt.Sprintf("%s:v1", elementID)
						if subprocessKey != currentProcessKey {
							logger.Info("Found collaboration subprocess",
								logger.String("subprocess_id", elementID),
								logger.String("subprocess_key", subprocessKey))
							allProcesses = append(allProcesses, elementMap)
						}
					}
				}
			}
		}
	}

	return allProcesses
}

// createSubscriptionsForProcess creates subscriptions for Message Start Events in a specific process
// Создает подписки для Message Start Events в конкретном процессе
func (cm *CollaborationManager) createSubscriptionsForProcess(bpmnProcess *models.BPMNProcess, processElement map[string]interface{}, mainProcessKey string) error {
	processID, exists := processElement["id"].(string)
	if !exists {
		return fmt.Errorf("process ID not found")
	}

	// Use the main BPMN process key since all collaboration processes are stored as one BPMN
	processKey := mainProcessKey

	logger.Info("Searching for Message Start Events in collaboration process",
		logger.String("process_id", processID),
		logger.String("process_key", processKey))

	// Find Message Start Events in this process
	startEventIDs := cm.findMessageStartEvents(bpmnProcess, processElement)

	// Create subscriptions for each Message Start Event found
	for _, startEventID := range startEventIDs {
		if err := cm.createSubscriptionForMessageStartEvent(bpmnProcess, startEventID, processKey, processID); err != nil {
			logger.Warn("Failed to create subscription for collaboration Message Start Event",
				logger.String("process_id", processID),
				logger.String("start_event_id", startEventID),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Created subscription for collaboration Message Start Event",
				logger.String("process_id", processID),
				logger.String("start_event_id", startEventID),
				logger.String("process_key", processKey))
		}
	}

	return nil
}

// findMessageStartEvents finds Message Start Events in process flow nodes
// Находит Message Start Events в flow nodes процесса
func (cm *CollaborationManager) findMessageStartEvents(bpmnProcess *models.BPMNProcess, processElement map[string]interface{}) []string {
	var startEventIDs []string

	if flowNodeIds, exists := processElement["flow_node_ids"]; exists {
		if nodeList, ok := flowNodeIds.([]interface{}); ok {
			for _, nodeID := range nodeList {
				if nodeIDStr, ok := nodeID.(string); ok {
					// Check if this is a start event
					if element, exists := bpmnProcess.Elements[nodeIDStr]; exists {
						if elementMap, ok := element.(map[string]interface{}); ok {
							if elementType, exists := elementMap["type"]; exists && elementType == "startEvent" {
								// Check if this is a Message Start Event
								if cm.isMessageStartEventByID(bpmnProcess, nodeIDStr) {
									startEventIDs = append(startEventIDs, nodeIDStr)
									logger.Info("Found Message Start Event in collaboration process",
										logger.String("start_event_id", nodeIDStr))
								}
							}
						}
					}
				}
			}
		}
	}

	return startEventIDs
}

// isMessageStartEventByID checks if start event is Message Start Event by ID
// Проверяет является ли start event Message Start Event по ID
func (cm *CollaborationManager) isMessageStartEventByID(bpmnProcess *models.BPMNProcess, startEventID string) bool {
	startElement, exists := bpmnProcess.Elements[startEventID]
	if !exists {
		return false
	}

	elementMap, ok := startElement.(map[string]interface{})
	if !ok {
		return false
	}

	// Check for event definitions
	if eventDefinitions, hasEventDefs := elementMap["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					if eventType, exists := eventDefMap["type"]; exists && eventType == "messageEventDefinition" {
						return true
					}
				}
			}
		}
	}

	return false
}

// createSubscriptionForMessageStartEvent creates subscription for specific Message Start Event
// Создает подписку для конкретного Message Start Event
func (cm *CollaborationManager) createSubscriptionForMessageStartEvent(bpmnProcess *models.BPMNProcess, startEventID, processKey, processID string) error {
	// For now, delegate to ProcessStarter - this creates some coupling but avoids duplication
	// We can extract common subscription creation logic later if needed
	starter := NewProcessStarter(cm.storage, cm.component)
	return starter.createMessageStartEventSubscription(bpmnProcess, startEventID, processKey, map[string]interface{}{})
}
