/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"encoding/json"
	"fmt"
)

// CreateParseBPMNFileMessage creates JSON message for BPMN file parsing
// Создает JSON сообщение для парсинга BPMN файла
func CreateParseBPMNFileMessage(payload ParseBPMNFilePayload) (string, error) {
	request := ParserRequest{
		Type:    "parse_bpmn_file",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateParseBPMNContentMessage creates JSON message for BPMN content parsing
// Создает JSON сообщение для парсинга содержимого BPMN
func CreateParseBPMNContentMessage(payload ParseBPMNContentPayload) (string, error) {
	request := ParserRequest{
		Type:    "parse_bpmn_content",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateValidateBPMNMessage creates JSON message for BPMN validation
// Создает JSON сообщение для валидации BPMN
func CreateValidateBPMNMessage(payload ValidateBPMNPayload) (string, error) {
	request := ParserRequest{
		Type:    "validate_bpmn",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetProcessInfoMessage creates JSON message for getting process info
// Создает JSON сообщение для получения информации о процессе
func CreateGetProcessInfoMessage(payload GetProcessInfoPayload) (string, error) {
	request := ParserRequest{
		Type:    "get_process_info",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateListProcessesMessage creates JSON message for listing processes
// Создает JSON сообщение для списка процессов
func CreateListProcessesMessage(payload ListProcessesPayload) (string, error) {
	request := ParserRequest{
		Type:    "list_processes",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateDeleteProcessMessage creates JSON message for deleting process
// Создает JSON сообщение для удаления процесса
func CreateDeleteProcessMessage(payload DeleteProcessPayload) (string, error) {
	request := ParserRequest{
		Type:    "delete_process",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetParserStatsMessage creates JSON message for getting parser statistics
// Создает JSON сообщение для получения статистики парсера
func CreateGetParserStatsMessage() (string, error) {
	request := ParserRequest{
		Type:    "get_stats",
		Payload: make(map[string]interface{}),
	}
	return marshalRequest(request)
}

// Helper functions
// Вспомогательные функции

// marshalRequest marshals ParserRequest to JSON string
// Маршалит ParserRequest в JSON строку
func marshalRequest(request ParserRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal parser request: %w", err)
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

// CreateParserResponse creates a successful parser response
// Создает успешный ответ парсера
func CreateParserResponse(responseType, requestID string, result interface{}) ParserResponse {
	return ParserResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   true,
		Result:    result,
	}
}

// CreateParserErrorResponse creates an error parser response
// Создает ответ парсера с ошибкой
func CreateParserErrorResponse(responseType, requestID, errorMsg string) ParserResponse {
	return ParserResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
}
