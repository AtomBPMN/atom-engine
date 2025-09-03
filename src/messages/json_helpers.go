/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package messages

import (
	"encoding/json"
	"fmt"
)

// CreatePublishMessageMessage creates JSON message for message publishing
// Создает JSON сообщение для публикации сообщения
func CreatePublishMessageMessage(payload PublishMessagePayload) (string, error) {
	request := MessageRequest{
		Type:    "publish_message",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateCorrelateMessageMessage creates JSON message for message correlation
// Создает JSON сообщение для корреляции сообщения
func CreateCorrelateMessageMessage(payload CorrelateMessagePayload) (string, error) {
	request := MessageRequest{
		Type:    "correlate_message",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateSubscriptionMessage creates JSON message for subscription creation
// Создает JSON сообщение для создания подписки
func CreateSubscriptionMessage(payload CreateSubscriptionPayload) (string, error) {
	request := MessageRequest{
		Type:    "create_subscription",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateDeleteSubscriptionMessage creates JSON message for subscription deletion
// Создает JSON сообщение для удаления подписки
func CreateDeleteSubscriptionMessage(payload DeleteSubscriptionPayload) (string, error) {
	request := MessageRequest{
		Type:    "delete_subscription",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateListSubscriptionsMessage creates JSON message for listing subscriptions
// Создает JSON сообщение для списка подписок
func CreateListSubscriptionsMessage(payload ListSubscriptionsPayload) (string, error) {
	request := MessageRequest{
		Type:    "list_subscriptions",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateListBufferedMessagesMessage creates JSON message for listing buffered messages
// Создает JSON сообщение для списка буферизованных сообщений
func CreateListBufferedMessagesMessage(payload ListBufferedMessagesPayload) (string, error) {
	request := MessageRequest{
		Type:    "list_buffered_messages",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateCleanupExpiredMessage creates JSON message for cleanup expired messages
// Создает JSON сообщение для очистки просроченных сообщений
func CreateCleanupExpiredMessage(payload CleanupExpiredPayload) (string, error) {
	request := MessageRequest{
		Type:    "cleanup_expired",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetMessageStatsMessage creates JSON message for getting message statistics
// Создает JSON сообщение для получения статистики сообщений
func CreateGetMessageStatsMessage(payload GetStatsPayload) (string, error) {
	request := MessageRequest{
		Type:    "get_stats",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// Helper functions
// Вспомогательные функции

// marshalRequest marshals MessageRequest to JSON string
// Маршалит MessageRequest в JSON строку
func marshalRequest(request MessageRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message request: %w", err)
	}
	return string(jsonData), nil
}

// structToMap converts struct to map[string]interface{}
// Конвертирует структуру в map[string]interface{}
func structToMap(v interface{}) map[string]interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		return make(map[string]interface{})
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return make(map[string]interface{})
	}

	return result
}

// mapToStruct converts map[string]interface{} to struct
// Конвертирует map[string]interface{} в структуру
func mapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to struct: %w", err)
	}

	return nil
}

// CreateMessageResponse creates a successful message response
// Создает успешный ответ сообщения
func CreateMessageResponse(responseType, requestID string, result interface{}) MessageResponse {
	return MessageResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   true,
		Result:    result,
	}
}

// CreateMessageErrorResponse creates an error message response
// Создает ответ сообщения с ошибкой
func CreateMessageErrorResponse(responseType, requestID, errorMsg string) MessageResponse {
	return MessageResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
}
