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

// IntermediateCatchMessageHandler handles message events for intermediate catch events
// Обработчик message событий для промежуточных catch событий
type IntermediateCatchMessageHandler struct {
	processComponent ComponentInterface
}

// NewIntermediateCatchMessageHandler creates new message handler
// Создает новый обработчик message событий
func NewIntermediateCatchMessageHandler(processComponent ComponentInterface) *IntermediateCatchMessageHandler {
	return &IntermediateCatchMessageHandler{
		processComponent: processComponent,
	}
}

// HandleMessageEvent handles message intermediate catch events
// Обрабатывает message промежуточные catch события
func (icmh *IntermediateCatchMessageHandler) HandleMessageEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Handling message intermediate catch event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Check if this token was activated by message correlation
	// Проверяем был ли этот токен активирован через message correlation
	if icmh.isMessageCorrelatedToken(token) {
		logger.Info("Message Intermediate Catch Event already correlated - proceeding to next elements",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))

		// Token was already activated by message correlation, proceed to next elements
		// Токен уже активирован через message correlation, переходим к следующим элементам
		return icmh.proceedToNextElements(element)
	}

	// Extract message information from event definition
	// Извлекаем информацию о сообщении из event definition
	messageName := ""
	correlationKey := ""

	// Try multiple places to find messageRef
	// Пытаемся найти messageRef в разных местах
	if reference, exists := eventDef["reference"]; exists {
		if refStr, ok := reference.(string); ok {
			// reference is message ID, need to find the actual message name
			// reference это ID сообщения, нужно найти настоящее имя сообщения
			actualMessageName := icmh.getMessageNameByReference(token, refStr)
			if actualMessageName != "" {
				messageName = actualMessageName
				logger.Info("Message catch event reference resolved to name",
					logger.String("reference", refStr),
					logger.String("message_name", actualMessageName))
			} else {
				messageName = refStr
				logger.Info("Message catch event reference found (fallback to ID)",
					logger.String("reference", refStr))
			}
		}
	} else if messageRef, exists := eventDef["message_ref"]; exists {
		if messageRefStr, ok := messageRef.(string); ok {
			messageName = messageRefStr
			logger.Info("Message catch event message_ref found",
				logger.String("message_ref", messageRefStr))
		}
	} else if attributes, exists := eventDef["attributes"]; exists {
		if attrsMap, ok := attributes.(map[string]interface{}); ok {
			if messageRef, exists := attrsMap["messageRef"]; exists {
				if messageRefStr, ok := messageRef.(string); ok {
					messageName = messageRefStr
					logger.Info("Message catch event messageRef in attributes found",
						logger.String("messageRef", messageRefStr))
				}
			}
		}
	} else if message, exists := eventDef["message"]; exists {
		if msgMap, ok := message.(map[string]interface{}); ok {
			if messageRef, exists := msgMap["message_ref"]; exists {
				if messageRefStr, ok := messageRef.(string); ok {
					messageName = messageRefStr
					logger.Info("Message catch event message_ref in message found",
						logger.String("message_ref", messageRefStr))
				}
			}
		}
	}

	// Extract correlation key from message definition
	// Извлекаем correlation key из message definition
	// For correlation key extraction, we need the message ID, not name
	// Для извлечения correlation key нужен ID сообщения, не имя
	messageID := ""
	if reference, exists := eventDef["reference"]; exists {
		if refStr, ok := reference.(string); ok {
			messageID = refStr
		}
	}
	if messageID != "" {
		correlationKey = icmh.extractCorrelationKeyFromMessage(token, messageID)

		// Evaluate FEEL expressions in correlation key BEFORE checking buffered messages
		// Вычисляем FEEL expressions в correlation key ПЕРЕД проверкой буферизованных сообщений
		correlationKey = icmh.evaluateCorrelationKeyExpression(correlationKey, token)
	}

	// Get outgoing flows for later continuation
	// Получаем исходящие потоки для последующего продолжения
	var nextElements []string
	if outgoing, exists := element["outgoing"]; exists {
		if outgoingList, ok := outgoing.([]interface{}); ok {
			for _, item := range outgoingList {
				if flowID, ok := item.(string); ok {
					nextElements = append(nextElements, flowID)
				}
			}
		} else if outgoingStr, ok := outgoing.(string); ok {
			nextElements = append(nextElements, outgoingStr)
		}
	}

	// FIRST: Check if there are buffered messages that match this catch event
	// СНАЧАЛА: Проверяем есть ли буферизованные сообщения соответствующие этому catch event
	if messageName != "" {
		logger.Info("Checking for buffered messages",
			logger.String("token_id", token.TokenID),
			logger.String("message_name", messageName),
			logger.String("correlation_key", correlationKey))

		bufferedMessage, err := icmh.processComponent.CheckBufferedMessages(messageName, correlationKey)
		if err != nil {
			logger.Error("Failed to check buffered messages",
				logger.String("error", err.Error()))
		}

		if bufferedMessage != nil {
			logger.Info("Found buffered message for token",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("correlation_key", correlationKey))

			// Process buffered message (merge variables with token)
			// Обрабатываем буферизованное сообщение (объединяем переменные с токеном)
			if err := icmh.processComponent.ProcessBufferedMessage(bufferedMessage, token); err != nil {
				logger.Error("Failed to process buffered message", logger.String("error", err.Error()))
				return &ExecutionResult{
					Success:   false,
					Error:     fmt.Sprintf("failed to process buffered message: %v", err),
					Completed: false,
				}, nil
			}

			// Continue to next elements after processing message
			// Переходим к следующим элементам после обработки сообщения
			logger.Info("Proceeding to next elements after processing buffered message",
				logger.String("token_id", token.TokenID),
				logger.Int("next_elements_count", len(nextElements)))

			return &ExecutionResult{
				Success:      true,
				TokenUpdated: true, // Token was updated with message variables
				NextElements: nextElements,
				Completed:    false,
			}, nil
		}
	}

	// SECOND: No buffered message found, create subscription and wait
	// ВТОРОЕ: Буферизованное сообщение не найдено, создаем подписку и ждем
	if messageName != "" {
		logger.Info("Creating message subscription for intermediate catch event",
			logger.String("token_id", token.TokenID),
			logger.String("message_name", messageName),
			logger.String("correlation_key", correlationKey))

		// Create message subscription
		// Создаем подписку на сообщение
		subscription := &models.ProcessMessageSubscription{
			ID:                   models.GenerateID(),
			TenantID:             "DEFAULT_TENANT",
			ProcessDefinitionKey: token.ProcessKey,
			ProcessVersion:       1, // Default version
			StartEventID:         token.CurrentElementID,
			MessageName:          messageName,
			CorrelationKey:       correlationKey,
			IsActive:             true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := icmh.processComponent.CreateMessageSubscription(subscription); err != nil {
			logger.Error("Failed to create message subscription",
				logger.String("token_id", token.TokenID),
				logger.String("message_name", messageName),
				logger.String("error", err.Error()))
			return &ExecutionResult{
				Success:   false,
				Error:     fmt.Sprintf("failed to create message subscription: %v", err),
				Completed: false,
			}, nil
		}

		logger.Info("Message subscription created successfully",
			logger.String("token_id", token.TokenID),
			logger.String("subscription_id", subscription.ID),
			logger.String("message_name", messageName))

		// Set token to waiting state
		// Устанавливаем токен в состояние ожидания
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: nextElements,
			WaitingFor:   fmt.Sprintf("message:%s", messageName),
			Completed:    false,
		}, nil
	}

	// No message name found - cannot create subscription
	// Имя сообщения не найдено - нельзя создать подписку
	logger.Error("Cannot create message subscription - no message name found",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	return &ExecutionResult{
		Success:   false,
		Error:     "message intermediate catch event: no message name found in event definition",
		Completed: false,
	}, nil
}

// extractCorrelationKeyFromMessage extracts correlation key from message definition
// Извлекает correlation key из определения сообщения
func (icmh *IntermediateCatchMessageHandler) extractCorrelationKeyFromMessage(token *models.Token, messageName string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := icmh.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		logger.Error("Failed to get BPMN process for correlation key extraction",
			logger.String("token_id", token.TokenID),
			logger.String("message_name", messageName),
			logger.String("error", err.Error()))
		return ""
	}

	// Extract elements from process map
	// Извлекаем элементы из карты процесса
	elements, ok := bpmnProcess["elements"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid elements structure in BPMN process",
			logger.String("token_id", token.TokenID),
			logger.String("message_name", messageName))
		return ""
	}

	// Find message definition by ID (not name)
	// Ищем определение сообщения по ID (не по имени)
	logger.Info("Searching for message definition",
		logger.String("looking_for", messageName),
		logger.Int("total_elements", len(elements)))

	for elementID, element := range elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "message" {
				logger.Info("Found message element",
					logger.String("element_id", elementID),
					logger.String("looking_for", messageName))
				if elementID == messageName {
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
																		logger.Info("Correlation key extracted from message definition",
																			logger.String("message_name", messageName),
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

					logger.Warn("Message definition found but no correlation key in extensions",
						logger.String("message_name", messageName),
						logger.String("element_id", elementID))
					return ""
				}
			}
		}
	}

	logger.Warn("Message definition not found for correlation key extraction",
		logger.String("message_name", messageName))
	return ""
}

// getMessageNameByReference gets message name by reference ID
// Получает имя сообщения по ID ссылки
func (icmh *IntermediateCatchMessageHandler) getMessageNameByReference(token *models.Token, messageRef string) string {
	// Get full BPMN process definition
	// Получаем полное определение BPMN процесса
	bpmnProcess, err := icmh.processComponent.GetBPMNProcessForToken(token)
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
	for elementID, element := range elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "message" {
				if elementID == messageRef {
					// Found message definition, extract name
					// Нашли определение сообщения, извлекаем имя
					if name, exists := elementMap["name"]; exists {
						if nameStr, ok := name.(string); ok {
							logger.Info("Message name resolved by reference",
								logger.String("message_ref", messageRef),
								logger.String("message_name", nameStr))
							return nameStr
						}
					}
					// Fallback to ID if no name
					// Fallback на ID если нет имени
					logger.Info("Message name fallback to ID",
						logger.String("message_ref", messageRef))
					return messageRef
				}
			}
		}
	}

	logger.Warn("Message definition not found for name resolution",
		logger.String("message_ref", messageRef))
	return ""
}

// isMessageCorrelatedToken checks if token was activated by message correlation
// Проверяет был ли токен активирован через message correlation
func (icmh *IntermediateCatchMessageHandler) isMessageCorrelatedToken(token *models.Token) bool {
	// Check if token variables contain message correlation marker
	// Проверяем содержат ли переменные токена маркер message correlation
	if hasMarker, exists := token.Variables["_message_correlated"]; exists && hasMarker == true {
		logger.Info("Token has message correlation marker - activated by message correlation",
			logger.String("token_id", token.TokenID))
		return true
	}

	// No correlation marker found - this is first-time execution of intermediate catch event
	// Маркер корреляции не найден - это первое выполнение intermediate catch event
	logger.Info("Token appears to be first-time execution of intermediate catch event",
		logger.String("token_id", token.TokenID))
	return false
}

// proceedToNextElements proceeds to next elements without waiting
// Переходит к следующим элементам без ожидания
func (icmh *IntermediateCatchMessageHandler) proceedToNextElements(element map[string]interface{}) (*ExecutionResult, error) {
	// Get outgoing flows
	// Получаем исходящие потоки
	var nextElements []string
	if outgoing, exists := element["outgoing"]; exists {
		if outgoingList, ok := outgoing.([]interface{}); ok {
			for _, item := range outgoingList {
				if flowID, ok := item.(string); ok {
					nextElements = append(nextElements, flowID)
				}
			}
		} else if outgoingStr, ok := outgoing.(string); ok {
			nextElements = append(nextElements, outgoingStr)
		}
	}

	logger.Info("Message Intermediate Catch Event proceeding to next elements",
		logger.Int("next_elements_count", len(nextElements)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// evaluateCorrelationKeyExpression evaluates FEEL expressions in correlation key for message correlation
// Вычисляет FEEL expressions в correlation key для корреляции сообщений
func (icmh *IntermediateCatchMessageHandler) evaluateCorrelationKeyExpression(correlationKey string, token *models.Token) string {
	// If not a FEEL expression (doesn't start with =), return as is
	// Если не FEEL expression (не начинается с =), возвращаем как есть
	if correlationKey == "" || correlationKey[0] != '=' {
		return correlationKey
	}

	// Get expression component through process component
	// Получаем expression компонент через process компонент
	if icmh.processComponent == nil {
		logger.Warn("Process component not available for correlation key evaluation",
			logger.String("token_id", token.TokenID),
			logger.String("correlation_key", correlationKey))
		return correlationKey[1:] // Fallback to remove "=" prefix
	}

	// Get core interface
	core := icmh.processComponent.GetCore()
	if core == nil {
		logger.Warn("Core interface not available for correlation key evaluation",
			logger.String("token_id", token.TokenID),
			logger.String("correlation_key", correlationKey))
		return correlationKey[1:] // Fallback to remove "=" prefix
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		logger.Warn("Expression component not available for correlation key evaluation",
			logger.String("token_id", token.TokenID),
			logger.String("correlation_key", correlationKey))
		return correlationKey[1:] // Fallback to remove "=" prefix
	}

	// Cast to expression evaluator interface
	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		logger.Warn("Failed to cast expression component for correlation key evaluation",
			logger.String("token_id", token.TokenID),
			logger.String("correlation_key", correlationKey))
		return correlationKey[1:] // Fallback to remove "=" prefix
	}

	// Evaluate FEEL expression
	result, err := expressionComp.EvaluateExpressionEngine(correlationKey, token.Variables)
	if err != nil {
		logger.Error("Failed to evaluate FEEL expression in correlation key",
			logger.String("token_id", token.TokenID),
			logger.String("expression", correlationKey),
			logger.String("error", err.Error()))
		return correlationKey[1:] // Fallback to remove "=" prefix on error
	}

	// Convert result to string
	evaluatedKey := fmt.Sprintf("%v", result)
	if evaluatedKey != "" {
		logger.Info("Message correlation key FEEL expression evaluated successfully",
			logger.String("token_id", token.TokenID),
			logger.String("original_expression", correlationKey),
			logger.String("evaluated_key", evaluatedKey))
		return evaluatedKey
	}

	// Fallback if result is empty
	logger.Warn("FEEL expression evaluation resulted in empty value",
		logger.String("token_id", token.TokenID),
		logger.String("expression", correlationKey))
	return correlationKey[1:] // Fallback to remove "=" prefix
}
