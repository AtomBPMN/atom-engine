/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package messages

// MessageRequest base structure for all message requests
// Базовая структура для всех запросов сообщений
type MessageRequest struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
}

// MessageResponse base structure for all message responses
// Базовая структура для всех ответов сообщений
type MessageResponse struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// PublishMessagePayload payload for publishing a message
// Payload для публикации сообщения
type PublishMessagePayload struct {
	TenantID       string                 `json:"tenant_id,omitempty"`
	MessageName    string                 `json:"message_name"`
	CorrelationKey string                 `json:"correlation_key,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	TTLSeconds     int                    `json:"ttl_seconds,omitempty"`
}

// CorrelateMessagePayload payload for correlating a message
// Payload для корреляции сообщения
type CorrelateMessagePayload struct {
	TenantID          string                 `json:"tenant_id,omitempty"`
	MessageName       string                 `json:"message_name"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
}

// CreateSubscriptionPayload payload for creating a message subscription
// Payload для создания подписки на сообщение
type CreateSubscriptionPayload struct {
	TenantID          string                 `json:"tenant_id,omitempty"`
	MessageName       string                 `json:"message_name"`
	ProcessKey        string                 `json:"process_key,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id,omitempty"`
	ElementID         string                 `json:"element_id"`
	TokenID           string                 `json:"token_id,omitempty"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	SubscriptionType  string                 `json:"subscription_type"` // PERMANENT or TEMPORARY
	Variables         map[string]interface{} `json:"variables,omitempty"`
	IsInterrupting    bool                   `json:"is_interrupting,omitempty"`
}

// DeleteSubscriptionPayload payload for deleting a message subscription
// Payload для удаления подписки на сообщение
type DeleteSubscriptionPayload struct {
	SubscriptionID string `json:"subscription_id"`
}

// ListSubscriptionsPayload payload for listing message subscriptions
// Payload для списка подписок на сообщения
type ListSubscriptionsPayload struct {
	TenantID string `json:"tenant_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// ListBufferedMessagesPayload payload for listing buffered messages
// Payload для списка буферизованных сообщений
type ListBufferedMessagesPayload struct {
	TenantID string `json:"tenant_id,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Offset   int    `json:"offset,omitempty"`
}

// CleanupExpiredPayload payload for cleaning up expired messages
// Payload для очистки просроченных сообщений
type CleanupExpiredPayload struct {
	TenantID string `json:"tenant_id,omitempty"`
}

// GetStatsPayload payload for getting message statistics
// Payload для получения статистики сообщений
type GetStatsPayload struct {
	TenantID string `json:"tenant_id,omitempty"`
}

// MessageResult result structure for message operations
// Структура результата для операций с сообщениями
type MessageResult struct {
	MessageID         string                 `json:"message_id,omitempty"`
	CorrelationID     string                 `json:"correlation_id,omitempty"`
	Success           bool                   `json:"success"`
	Message           string                 `json:"message,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id,omitempty"`
	Timestamp         int64                  `json:"timestamp,omitempty"`
}

// SubscriptionResult result structure for subscription operations
// Структура результата для операций с подписками
type SubscriptionResult struct {
	SubscriptionID string `json:"subscription_id,omitempty"`
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty"`
}

// CleanupResult result structure for cleanup operations
// Структура результата для операций очистки
type CleanupResult struct {
	ExpiredCount int    `json:"expired_count"`
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
}
