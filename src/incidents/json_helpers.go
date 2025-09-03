/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// IncidentRequest represents a JSON request for incident operations
// Представляет JSON запрос для операций с инцидентами
type IncidentRequest struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	RequestID string                 `json:"request_id,omitempty"`
}

// IncidentResponse represents a JSON response for incident operations
// Представляет JSON ответ для операций с инцидентами
type IncidentResponse struct {
	Type      string                 `json:"type"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// CreateIncidentPayload represents payload for incident creation
// Представляет полезную нагрузку для создания инцидента
type CreateIncidentPayload struct {
	Type              string                 `json:"type"`
	Message           string                 `json:"message"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id,omitempty"`
	ProcessKey        string                 `json:"process_key,omitempty"`
	ElementID         string                 `json:"element_id,omitempty"`
	ElementType       string                 `json:"element_type,omitempty"`
	JobKey            string                 `json:"job_key,omitempty"`
	JobType           string                 `json:"job_type,omitempty"`
	WorkerID          string                 `json:"worker_id,omitempty"`
	TimerID           string                 `json:"timer_id,omitempty"`
	MessageName       string                 `json:"message_name,omitempty"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	OriginalRetries   int                    `json:"original_retries,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ResolveIncidentPayload represents payload for incident resolution
// Представляет полезную нагрузку для разрешения инцидента
type ResolveIncidentPayload struct {
	IncidentID string `json:"incident_id"`
	Action     string `json:"action"`
	Comment    string `json:"comment,omitempty"`
	ResolvedBy string `json:"resolved_by,omitempty"`
	NewRetries int    `json:"new_retries,omitempty"`
}

// GetIncidentPayload represents payload for getting incident
// Представляет полезную нагрузку для получения инцидента
type GetIncidentPayload struct {
	IncidentID string `json:"incident_id"`
}

// ListIncidentsPayload represents payload for listing incidents
// Представляет полезную нагрузку для получения списка инцидентов
type ListIncidentsPayload struct {
	Status            []string `json:"status,omitempty"`
	Type              []string `json:"type,omitempty"`
	ProcessInstanceID string   `json:"process_instance_id,omitempty"`
	ProcessKey        string   `json:"process_key,omitempty"`
	ElementID         string   `json:"element_id,omitempty"`
	JobKey            string   `json:"job_key,omitempty"`
	WorkerID          string   `json:"worker_id,omitempty"`
	Limit             int      `json:"limit,omitempty"`
	Offset            int      `json:"offset,omitempty"`
}

// CreateIncidentMessage creates JSON message for incident creation
// Создает JSON сообщение для создания инцидента
func CreateIncidentMessage(payload CreateIncidentPayload) (string, error) {
	request := IncidentRequest{
		Type:    "create_incident",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateResolveIncidentMessage creates JSON message for incident resolution
// Создает JSON сообщение для разрешения инцидента
func CreateResolveIncidentMessage(payload ResolveIncidentPayload) (string, error) {
	request := IncidentRequest{
		Type:    "resolve_incident",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetIncidentMessage creates JSON message for getting incident
// Создает JSON сообщение для получения инцидента
func CreateGetIncidentMessage(payload GetIncidentPayload) (string, error) {
	request := IncidentRequest{
		Type:    "get_incident",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateListIncidentsMessage creates JSON message for listing incidents
// Создает JSON сообщение для получения списка инцидентов
func CreateListIncidentsMessage(payload ListIncidentsPayload) (string, error) {
	request := IncidentRequest{
		Type:    "list_incidents",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetIncidentStatsMessage creates JSON message for getting incident stats
// Создает JSON сообщение для получения статистики инцидентов
func CreateGetIncidentStatsMessage() (string, error) {
	request := IncidentRequest{
		Type:    "get_incident_stats",
		Payload: make(map[string]interface{}),
	}
	return marshalRequest(request)
}

// CreateIncidentSuccessResponse creates successful incident response
// Создает успешный ответ об инциденте
func CreateIncidentSuccessResponse(responseType string, incident *Incident) string {
	response := IncidentResponse{
		Type:    responseType,
		Success: true,
		Data:    structToMap(incident),
	}

	if data, err := json.Marshal(response); err == nil {
		return string(data)
	}
	return ""
}

// CreateIncidentListResponse creates incident list response
// Создает ответ со списком инцидентов
func CreateIncidentListResponse(incidents []*Incident, total int) string {
	response := IncidentResponse{
		Type:    "list_incidents_response",
		Success: true,
		Data: map[string]interface{}{
			"incidents": incidents,
			"total":     total,
		},
	}

	if data, err := json.Marshal(response); err == nil {
		return string(data)
	}
	return ""
}

// CreateIncidentStatsResponse creates incident stats response
// Создает ответ со статистикой инцидентов
func CreateIncidentStatsResponse(stats *IncidentStats) string {
	response := IncidentResponse{
		Type:    "get_incident_stats_response",
		Success: true,
		Data:    structToMap(stats),
	}

	if data, err := json.Marshal(response); err == nil {
		return string(data)
	}
	return ""
}

// CreateIncidentErrorResponse creates error response
// Создает ответ об ошибке
func CreateIncidentErrorResponse(responseType, requestID, errorMsg string) string {
	response := IncidentResponse{
		Type:      responseType,
		Success:   false,
		Error:     errorMsg,
		RequestID: requestID,
	}

	if data, err := json.Marshal(response); err == nil {
		return string(data)
	}
	return fmt.Sprintf(`{"type":"%s","success":false,"error":"Failed to marshal error response"}`, responseType)
}

// Helper functions

// structToMap converts a struct to map[string]interface{}
// Конвертирует структуру в map[string]interface{}
func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if data, err := json.Marshal(obj); err == nil {
		json.Unmarshal(data, &result)
	}

	return result
}

// mapToStruct converts map[string]interface{} to a struct
// Конвертирует map[string]interface{} в структуру
func mapToStruct(data map[string]interface{}, target interface{}) error {
	if jsonData, err := json.Marshal(data); err != nil {
		return err
	} else {
		return json.Unmarshal(jsonData, target)
	}
}

// marshalRequest marshals request to JSON string
// Маршалит запрос в JSON строку
func marshalRequest(request IncidentRequest) (string, error) {
	if data, err := json.Marshal(request); err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	} else {
		return string(data), nil
	}
}

// isZeroValue checks if a value is zero value of its type
// Проверяет является ли значение нулевым для своего типа
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
