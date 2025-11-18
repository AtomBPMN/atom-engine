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
)

// IntermediateThrowEventExecutor executes intermediate throw events
// Исполнитель промежуточных событий бросания
type IntermediateThrowEventExecutor struct {
	processComponent ComponentInterface
}

// NewIntermediateThrowEventExecutor creates new intermediate throw event executor
// Создает новый исполнитель промежуточного события бросания
func NewIntermediateThrowEventExecutor(processComponent ComponentInterface) *IntermediateThrowEventExecutor {
	return &IntermediateThrowEventExecutor{
		processComponent: processComponent,
	}
}

// Execute executes intermediate throw event
// Выполняет промежуточное событие бросания
func (itee *IntermediateThrowEventExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing intermediate throw event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get event name for logging
	eventName, _ := element["name"].(string)
	if eventName == "" {
		eventName = token.CurrentElementID
	}

	// Check for event definitions to determine throw event type
	// Проверяем event definitions чтобы определить тип события бросания
	if eventDefinitions, hasEventDefs := element["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					eventType, _ := eventDefMap["type"].(string)

					// Handle message events
					if eventType == "messageEventDefinition" {
						return itee.handleMessageThrowEvent(token, element, eventDefMap)
					}

					// Handle signal events
					if eventType == "signalEventDefinition" {
						return itee.handleSignalThrowEvent(token, element, eventDefMap)
					}

					// Handle other event types...
				}
			}
		}
	}

	// Regular intermediate throw event - no specific action
	// Обычное промежуточное событие бросания - никаких специальных действий
	logger.Info("Regular intermediate throw event executed",
		logger.String("token_id", token.TokenID),
		logger.String("event_name", eventName))

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		// Throw events without outgoing flows complete the token
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// handleMessageThrowEvent handles message intermediate throw events
// Обрабатывает промежуточные события бросания сообщений
func (itee *IntermediateThrowEventExecutor) handleMessageThrowEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Handling message intermediate throw event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract message information from event definition
	// Извлекаем информацию о сообщении из event definition
	messageName := ""
	correlationKey := ""

	// Try multiple places to find messageRef
	// Пытаемся найти messageRef в разных местах
	messageRef := ""
	if reference, exists := eventDef["reference"]; exists {
		if refStr, ok := reference.(string); ok {
			messageRef = refStr
			logger.Info("Message throw event reference found",
				logger.String("reference", refStr))
		}
	} else if msgRef, exists := eventDef["message_ref"]; exists {
		if msgRefStr, ok := msgRef.(string); ok {
			messageRef = msgRefStr
			logger.Info("Message throw event message_ref found",
				logger.String("message_ref", msgRefStr))
		}
	}

	// If we have messageRef, resolve it to actual message name
	// Если у нас есть messageRef, разрешаем его в настоящее имя сообщения
	if messageRef != "" {
		actualMessageName := itee.getMessageNameByReference(token, messageRef)
		if actualMessageName != "" {
			messageName = actualMessageName
			logger.Info("Message throw event reference resolved to name",
				logger.String("reference", messageRef),
				logger.String("message_name", actualMessageName))
		} else {
			messageName = messageRef
			logger.Info("Message throw event using reference as name (fallback)",
				logger.String("reference", messageRef))
		}

		// Extract correlation key from message definition
		// Извлекаем correlation key из определения сообщения
		correlationKey = itee.extractCorrelationKeyFromMessage(token, messageRef)
	}

	// If no messageRef found, try to extract message name from extension elements
	// Если messageRef не найден, пытаемся извлечь имя сообщения из extension elements
	if messageName == "" {
		if extensionElements, exists := element["extension_elements"]; exists {
			if extList, ok := extensionElements.([]interface{}); ok {
				for _, ext := range extList {
					if extMap, ok := ext.(map[string]interface{}); ok {
						if extensions, exists := extMap["extensions"]; exists {
							if extensionsList, ok := extensions.([]interface{}); ok {
								for _, extension := range extensionsList {
									if extensionMap, ok := extension.(map[string]interface{}); ok {
										if extensionType, exists := extensionMap["type"]; exists && extensionType == "taskDefinition" {
											if attributes, exists := extensionMap["attributes"]; exists {
												if attrsMap, ok := attributes.(map[string]interface{}); ok {
													if msgType, exists := attrsMap["type"]; exists {
														if msgTypeStr, ok := msgType.(string); ok {
															messageName = msgTypeStr
															logger.Info("Message name extracted from taskDefinition",
																logger.String("message_name", msgTypeStr))
															break
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
		}
	}

	// Fallback: Extract correlation key from token variables if not found in message definition
	// Резерв: Извлекаем correlation key из переменных токена если не найден в определении сообщения
	if correlationKey == "" {
		if corrKey, exists := token.Variables["correlationKey"]; exists {
			if corrKeyStr, ok := corrKey.(string); ok {
				correlationKey = corrKeyStr
				logger.Info("Using correlation key from token variables",
					logger.String("correlation_key", corrKeyStr))
			}
		}
	}

	// Publish message through process component
	// Публикуем сообщение через process component с element ID
	if itee.processComponent != nil && messageName != "" {
		result, err := itee.processComponent.PublishMessageWithElementID(
			messageName,
			correlationKey,
			token.CurrentElementID,
			token.Variables,
		)
		if err != nil {
			logger.Error("Failed to publish message from throw event",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("element_id", token.CurrentElementID),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Message published from throw event",
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey),
				logger.String("element_id", token.CurrentElementID),
				logger.Bool("instance_created", result != nil && result.InstanceCreated))
		}
	}

	// Get outgoing sequence flows and continue
	// Получаем исходящие sequence flows и продолжаем
	outgoing, exists := element["outgoing"]
	if !exists {
		// Message throw events without outgoing flows complete the token
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	logger.Info("Message throw event continuing execution",
		logger.String("token_id", token.TokenID),
		logger.String("message_name", messageName),
		logger.Int("next_elements", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// handleSignalThrowEvent handles signal intermediate throw events
// Обрабатывает промежуточные события бросания сигналов
func (itee *IntermediateThrowEventExecutor) handleSignalThrowEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Handling signal intermediate throw event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract signal name from event definition
	signalName := ""
	if signalRef, exists := eventDef["signal_ref"]; exists {
		if signalRefStr, ok := signalRef.(string); ok {
			signalName = signalRefStr
		}
	}

	// Fallback: use element ID as signal name if no signal_ref
	if signalName == "" {
		signalName = token.CurrentElementID + "_signal"
		logger.Warn("No signal_ref found, using element ID as signal name",
			logger.String("signal_name", signalName))
	}

	// Broadcast signal using process component
	if itee.processComponent != nil {
		variables := make(map[string]interface{})
		if token.Variables != nil {
			variables = token.Variables
		}

		err := itee.processComponent.BroadcastSignal(signalName, variables)
		if err != nil {
			logger.Error("Failed to broadcast signal",
				logger.String("signal_name", signalName),
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("failed to broadcast signal: %v", err),
				Completed: false,
			}, err
		}

		logger.Info("Successfully broadcast signal",
			logger.String("signal_name", signalName),
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))
	} else {
		logger.Warn("Process component not available, cannot broadcast signal")
	}

	// Continue with regular flow after broadcasting signal
	// Продолжаем с обычным потоком после broadcasting сигнала
	return itee.executeRegularThrowEvent(token, element)
}

// executeRegularThrowEvent executes regular throw event flow
// Выполняет поток обычного события бросания
func (itee *IntermediateThrowEventExecutor) executeRegularThrowEvent(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		// Throw events without outgoing flows complete the token
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (itee *IntermediateThrowEventExecutor) GetElementType() string {
	return "intermediateThrowEvent"
}

// getMessageNameByReference gets message name by reference ID for throw event
// Получает имя сообщения по ID ссылки для throw event
func (itee *IntermediateThrowEventExecutor) getMessageNameByReference(token *models.Token, messageRef string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := itee.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for message name resolution",
			logger.String("token_id", token.TokenID),
			logger.String("message_ref", messageRef),
			logger.String("error", err.Error()))
		return ""
	}

	// Extract elements from process map
	// Извлекаем элементы из карты процесса
	elements, ok := bpmnProcess["elements"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid elements structure in BPMN process for message name resolution",
			logger.String("token_id", token.TokenID),
			logger.String("message_ref", messageRef))
		return ""
	}

	// Find message definition by ID
	// Ищем определение сообщения по ID
	if element, exists := elements[messageRef]; exists {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "message" {
				if name, exists := elementMap["name"]; exists {
					if nameStr, ok := name.(string); ok {
						logger.Info("Message name resolved from reference",
							logger.String("message_ref", messageRef),
							logger.String("message_name", nameStr))
						return nameStr
					}
				}
			}
		}
	}

	logger.Warn("Message name not found for reference",
		logger.String("message_ref", messageRef))
	return ""
}

// extractCorrelationKeyFromMessage extracts correlation key from message definition for throw event
// Извлекает correlation key из определения сообщения для throw event
func (itee *IntermediateThrowEventExecutor) extractCorrelationKeyFromMessage(
	token *models.Token,
	messageID string,
) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := itee.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for correlation key extraction",
			logger.String("token_id", token.TokenID),
			logger.String("message_id", messageID),
			logger.String("error", err.Error()))
		return ""
	}

	// Extract elements from process map
	// Извлекаем элементы из карты процесса
	elements, ok := bpmnProcess["elements"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid elements structure in BPMN process for correlation key extraction",
			logger.String("token_id", token.TokenID),
			logger.String("message_id", messageID))
		return ""
	}

	// Find message definition by ID
	// Ищем определение сообщения по ID
	if element, exists := elements[messageID]; exists {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "message" {
				// Found message definition, extract correlation key
				// Нашли определение сообщения, извлекаем correlation key
				if extensionElements, exists := elementMap["extension_elements"]; exists {
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
																	logger.Info("Correlation key extracted from message definition for throw event",
																		logger.String("message_id", messageID),
																		logger.String("correlation_key", corrKeyStr))
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
				}

				logger.Info("Message definition found but no correlation key in extensions for throw event",
					logger.String("message_id", messageID))
				return ""
			}
		}
	}

	logger.Warn("Message definition not found for correlation key extraction in throw event",
		logger.String("message_id", messageID))
	return ""
}
