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

// EndEventExecutor executes end events
// Исполнитель конечных событий
type EndEventExecutor struct {
	processComponent ComponentInterface
}

// NewEndEventExecutor creates new end event executor
// Создает новый исполнитель конечного события
func NewEndEventExecutor(processComponent ComponentInterface) *EndEventExecutor {
	return &EndEventExecutor{
		processComponent: processComponent,
	}
}

// Execute executes end event
// Выполняет конечное событие
func (ee *EndEventExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing end event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Check for event definitions to determine end event type
	// Проверяем event definitions чтобы определить тип конечного события
	if eventDefinitions, hasEventDefs := element["event_definitions"]; hasEventDefs {
		if eventDefList, ok := eventDefinitions.([]interface{}); ok {
			for _, eventDef := range eventDefList {
				if eventDefMap, ok := eventDef.(map[string]interface{}); ok {
					eventType, _ := eventDefMap["type"].(string)

					// Handle message end events
					if eventType == "messageEventDefinition" {
						return ee.handleMessageEndEvent(token, element, eventDefMap)
					}

					// Handle signal end events
					if eventType == "signalEventDefinition" {
						return ee.handleSignalEndEvent(token, element, eventDefMap)
					}

					// Handle error end events
					if eventType == "errorEventDefinition" {
						return ee.handleErrorEndEvent(token, element, eventDefMap)
					}
				}
			}
		}
	}

	// Regular end event - just complete the token
	// Обычное конечное событие - просто завершаем токен
	logger.Info("Regular end event completed",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		Completed:    true,
	}, nil
}

// handleMessageEndEvent handles message end events
// Обрабатывает конечные события сообщений
func (ee *EndEventExecutor) handleMessageEndEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Handling message end event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract message information from event definition
	// Извлекаем информацию о сообщении из event definition
	messageName := ""
	correlationKey := ""

	// Try multiple places to find messageRef (similar to intermediate catch event)
	// Пытаемся найти messageRef в разных местах (аналогично intermediate catch event)
	messageRef := ""
	if reference, exists := eventDef["reference"]; exists {
		if refStr, ok := reference.(string); ok {
			messageRef = refStr
			logger.Info("Message end event reference found",
				logger.String("reference", refStr))
		}
	} else if msgRef, exists := eventDef["message_ref"]; exists {
		if msgRefStr, ok := msgRef.(string); ok {
			messageRef = msgRefStr
			logger.Info("Message end event message_ref found",
				logger.String("message_ref", msgRefStr))
		}
	} else if attributes, exists := eventDef["attributes"]; exists {
		if attrsMap, ok := attributes.(map[string]interface{}); ok {
			if msgRef, exists := attrsMap["messageRef"]; exists {
				if msgRefStr, ok := msgRef.(string); ok {
					messageRef = msgRefStr
					logger.Info("Message end event messageRef in attributes found",
						logger.String("messageRef", msgRefStr))
				}
			}
		}
	}

	// If we have messageRef, resolve it to actual message name
	// Если у нас есть messageRef, разрешаем его в настоящее имя сообщения
	if messageRef != "" {
		actualMessageName := ee.getMessageNameByReference(token, messageRef)
		if actualMessageName != "" {
			messageName = actualMessageName
			logger.Info("Message end event reference resolved to name",
				logger.String("reference", messageRef),
				logger.String("message_name", actualMessageName))
		} else {
			messageName = messageRef
			logger.Info("Message end event using reference as name (fallback)",
				logger.String("reference", messageRef))
		}

		// Extract correlation key from message definition
		// Извлекаем correlation key из определения сообщения
		correlationKey = ee.extractCorrelationKeyFromMessage(token, messageRef)
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

	// Publish message through process component before completing
	// Публикуем сообщение через process component перед завершением
	if ee.processComponent != nil && messageName != "" {
		result, err := ee.processComponent.PublishMessageWithElementID(
			messageName,
			correlationKey,
			token.CurrentElementID,
			token.Variables,
		)
		if err != nil {
			logger.Error("Failed to publish message from end event",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
		} else {
			logger.Info("Message published from end event",
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey),
				logger.Bool("instance_created", result != nil && result.InstanceCreated))
		}
	}

	// Complete the token - end events always complete
	// Завершаем токен - конечные события всегда завершают
	logger.Info("Message end event completed",
		logger.String("token_id", token.TokenID),
		logger.String("message_name", messageName))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		Completed:    true,
	}, nil
}

// handleSignalEndEvent handles signal end events
// Обрабатывает конечные события сигналов
func (ee *EndEventExecutor) handleSignalEndEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Handling signal end event",
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
	if ee.processComponent != nil {
		variables := make(map[string]interface{})
		if token.Variables != nil {
			variables = token.Variables
		}

		err := ee.processComponent.BroadcastSignal(signalName, variables)
		if err != nil {
			logger.Error("Failed to broadcast signal from end event",
				logger.String("signal_name", signalName),
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
			// Continue execution even if broadcast fails
		} else {
			logger.Info("Successfully broadcast signal from end event",
				logger.String("signal_name", signalName),
				logger.String("token_id", token.TokenID))
		}
	} else {
		logger.Warn("Process component not available, cannot broadcast signal")
	}

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		Completed:    true,
	}, nil
}

// handleErrorEndEvent handles error end events
// Обрабатывает конечные события ошибок
func (ee *EndEventExecutor) handleErrorEndEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Handling error end event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract error code and message from event definition
	errorCode := "GENERAL_ERROR" // Default error code
	errorMessage := "Error end event triggered"

	if errorRef, exists := eventDef["error_ref"]; exists {
		if errorRefStr, ok := errorRef.(string); ok {
			errorCode = errorRefStr
		}
	}

	// Check for error message in variables or element name
	if token.Variables != nil {
		if errMsg, exists := token.Variables["errorMessage"]; exists {
			if errMsgStr, ok := errMsg.(string); ok {
				errorMessage = errMsgStr
			}
		}
	}

	// Use element name as error message if available
	if elementName, exists := element["name"].(string); exists && elementName != "" {
		errorMessage = elementName
	}

	logger.Info("Error end event propagating error",
		logger.String("token_id", token.TokenID),
		logger.String("error_code", errorCode),
		logger.String("error_message", errorMessage))

	// Check for error boundary events in process component
	if ee.processComponent != nil {
		errorBoundary := ee.processComponent.FindMatchingErrorBoundary(token.TokenID, errorCode)
		if errorBoundary != nil {
			logger.Info("Found matching error boundary for error end event",
				logger.String("token_id", token.TokenID),
				logger.String("error_boundary_id", errorBoundary.ElementID),
				logger.String("error_code", errorCode))

			// Remove error boundary subscription
			ee.processComponent.RemoveErrorBoundariesForToken(token.TokenID)

			// Cancel current token and activate error boundary
			token.SetState(models.TokenStateCanceled)
			if err := ee.processComponent.UpdateToken(token); err != nil {
				logger.Error("Failed to update canceled token",
					logger.String("token_id", token.TokenID),
					logger.String("error", err.Error()))
			}

			// Create new token for error boundary flow
			return ee.activateErrorBoundaryFlow(token, errorBoundary, errorCode, errorMessage)
		}
	}

	// No error boundary found - this is unhandled error
	logger.Warn("No error boundary found for error end event, treating as unhandled error",
		logger.String("token_id", token.TokenID),
		logger.String("error_code", errorCode))

	// Mark token as failed with error info
	token.SetState(models.TokenStateFailed)
	if token.Variables == nil {
		token.Variables = make(map[string]interface{})
	}
	token.Variables["errorCode"] = errorCode
	token.Variables["errorMessage"] = errorMessage

	return &ExecutionResult{
		Success:   false,
		Error:     fmt.Sprintf("BPMN Error %s: %s", errorCode, errorMessage),
		Completed: true,
	}, fmt.Errorf("BPMN Error %s: %s", errorCode, errorMessage)
}

// activateErrorBoundaryFlow activates error boundary flow from error end event
// Активирует поток граничного события ошибки из error end event
func (ee *EndEventExecutor) activateErrorBoundaryFlow(
	originalToken *models.Token,
	errorBoundary *ErrorBoundarySubscription,
	errorCode, errorMessage string,
) (*ExecutionResult, error) {
	logger.Info("Activating error boundary flow from error end event",
		logger.String("original_token_id", originalToken.TokenID),
		logger.String("boundary_element_id", errorBoundary.ElementID),
		logger.String("error_code", errorCode),
		logger.Int("outgoing_flows_count", len(errorBoundary.OutgoingFlows)))

	// Prepare variables for error boundary
	variables := make(map[string]interface{})
	if originalToken.Variables != nil {
		for k, v := range originalToken.Variables {
			variables[k] = v
		}
	}
	variables["errorCode"] = errorCode
	variables["errorMessage"] = errorMessage

	// Continue execution with outgoing flows from error boundary event
	if len(errorBoundary.OutgoingFlows) > 0 {
		logger.Info("Error boundary has outgoing flows, continuing execution",
			logger.String("boundary_element_id", errorBoundary.ElementID),
			logger.Int("flows_count", len(errorBoundary.OutgoingFlows)))

		// Create new tokens for outgoing flows
		nextElements := make([]string, 0, len(errorBoundary.OutgoingFlows))
		for _, flow := range errorBoundary.OutgoingFlows {
			nextElements = append(nextElements, flow)
		}

		// Update original token variables
		originalToken.SetVariables(variables)

		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: nextElements,
			Completed:    false,
		}, nil
	}

	logger.Info("Error boundary event has no outgoing flows - process ends",
		logger.String("boundary_element_id", errorBoundary.ElementID))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: []string{},
		Completed:    true,
	}, nil
}

// GetElementType returns element type
// Возвращает тип элемента
func (ee *EndEventExecutor) GetElementType() string {
	return "endEvent"
}

// getMessageNameByReference gets message name by reference ID for end event
// Получает имя сообщения по ID ссылки для конечного события
func (ee *EndEventExecutor) getMessageNameByReference(token *models.Token, messageRef string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := ee.processComponent.GetBPMNProcessForToken(token)
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

// extractCorrelationKeyFromMessage extracts correlation key from message definition for end event
// Извлекает correlation key из определения сообщения для конечного события
func (ee *EndEventExecutor) extractCorrelationKeyFromMessage(token *models.Token, messageID string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := ee.processComponent.GetBPMNProcessForToken(token)
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
																	logger.Info("Correlation key extracted from message definition for end event",
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

				logger.Info("Message definition found but no correlation key in extensions for end event",
					logger.String("message_id", messageID))
				return ""
			}
		}
	}

	logger.Warn("Message definition not found for correlation key extraction in end event",
		logger.String("message_id", messageID))
	return ""
}
