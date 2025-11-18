/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package jobs

import (
	"encoding/json"
	"fmt"
)

// CreateJobMessage creates JSON message for job creation
// Создает JSON сообщение для создания job'а
func CreateJobMessage(payload CreateJobPayload) (string, error) {
	request := JobRequest{
		Type:    "create_job",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateActivateJobsMessage creates JSON message for job activation
// Создает JSON сообщение для активации job'ов
func CreateActivateJobsMessage(payload ActivateJobsPayload) (string, error) {
	request := JobRequest{
		Type:    "activate_jobs",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateCompleteJobMessage creates JSON message for job completion
// Создает JSON сообщение для завершения job'а
func CreateCompleteJobMessage(payload CompleteJobPayload) (string, error) {
	request := JobRequest{
		Type:    "complete_job",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateFailJobMessage creates JSON message for job failure
// Создает JSON сообщение для провала job'а
func CreateFailJobMessage(payload FailJobPayload) (string, error) {
	request := JobRequest{
		Type:    "fail_job",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateCancelJobMessage creates JSON message for job cancellation
// Создает JSON сообщение для отмены job'а
func CreateCancelJobMessage(payload CancelJobPayload) (string, error) {
	request := JobRequest{
		Type:    "cancel_job",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateUpdateJobRetriesMessage creates JSON message for updating job retries
// Создает JSON сообщение для обновления retries job'а
func CreateUpdateJobRetriesMessage(payload UpdateJobRetriesPayload) (string, error) {
	request := JobRequest{
		Type:    "update_job_retries",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateListJobsMessage creates JSON message for listing jobs
// Создает JSON сообщение для списка job'ов
func CreateListJobsMessage(payload ListJobsPayload) (string, error) {
	request := JobRequest{
		Type:    "list_jobs",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetJobMessage creates JSON message for getting job
// Создает JSON сообщение для получения job'а
func CreateGetJobMessage(payload GetJobPayload) (string, error) {
	request := JobRequest{
		Type:    "get_job",
		Payload: structToMap(payload),
	}
	return marshalRequest(request)
}

// CreateGetStatsMessage creates JSON message for getting job statistics
// Создает JSON сообщение для получения статистики job'ов
func CreateGetStatsMessage() (string, error) {
	request := JobRequest{
		Type:    "get_stats",
		Payload: make(map[string]interface{}),
	}
	return marshalRequest(request)
}

// Helper functions
// Вспомогательные функции

// marshalRequest marshals JobRequest to JSON string
// Маршалит JobRequest в JSON строку
func marshalRequest(request JobRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job request: %w", err)
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

// CreateJobResponse creates a successful job response
// Создает успешный ответ job'а
func CreateJobResponse(responseType, requestID string, result interface{}) JobResponse {
	return JobResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   true,
		Result:    result,
	}
}

// CreateJobErrorResponse creates an error job response
// Создает ответ job'а с ошибкой
func CreateJobErrorResponse(responseType, requestID, errorMsg string) JobResponse {
	return JobResponse{
		Type:      responseType,
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
}
