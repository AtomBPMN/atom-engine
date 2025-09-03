/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"context"
)

// Specialized incident creation methods for common incident types
// Специализированные методы создания инцидентов для распространенных типов

// CreateJobFailureIncident creates a job failure incident
// Создает инцидент отказа job
func (im *IncidentManager) CreateJobFailureIncident(ctx context.Context, jobKey, elementID, processInstanceID, message string, retries int) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeJobFailure,
		Message:           message,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		JobKey:            jobKey,
		OriginalRetries:   retries,
	}

	return im.CreateIncident(ctx, request)
}

// CreateBPMNErrorIncident creates a BPMN error incident
// Создает инцидент BPMN ошибки
func (im *IncidentManager) CreateBPMNErrorIncident(ctx context.Context, elementID, processInstanceID, errorCode, message string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeBPMNError,
		Message:           message,
		ErrorCode:         errorCode,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
	}

	return im.CreateIncident(ctx, request)
}

// CreateExpressionErrorIncident creates an expression error incident
// Создает инцидент ошибки выражения
func (im *IncidentManager) CreateExpressionErrorIncident(ctx context.Context, elementID, processInstanceID, expression, message string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeExpressionError,
		Message:           message,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		Metadata: map[string]interface{}{
			"expression": expression,
		},
	}

	return im.CreateIncident(ctx, request)
}

// CreateTimerErrorIncident creates a timer error incident
// Создает инцидент ошибки таймера
func (im *IncidentManager) CreateTimerErrorIncident(ctx context.Context, timerID, elementID, processInstanceID, message string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeTimerError,
		Message:           message,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		TimerID:           timerID,
	}

	return im.CreateIncident(ctx, request)
}

// CreateMessageErrorIncident creates a message error incident
// Создает инцидент ошибки сообщения
func (im *IncidentManager) CreateMessageErrorIncident(ctx context.Context, messageName, correlationKey, processInstanceID, message string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeMessageError,
		Message:           message,
		ProcessInstanceID: processInstanceID,
		MessageName:       messageName,
		CorrelationKey:    correlationKey,
	}

	return im.CreateIncident(ctx, request)
}

// CreateProcessErrorIncident creates a process error incident
// Создает инцидент ошибки процесса
func (im *IncidentManager) CreateProcessErrorIncident(ctx context.Context, elementID, processInstanceID, processKey, message string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeProcessError,
		Message:           message,
		ProcessInstanceID: processInstanceID,
		ProcessKey:        processKey,
		ElementID:         elementID,
	}

	return im.CreateIncident(ctx, request)
}

// CreateSystemErrorIncident creates a system error incident
// Создает инцидент системной ошибки
func (im *IncidentManager) CreateSystemErrorIncident(ctx context.Context, message string, metadata map[string]interface{}) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:     IncidentTypeSystemError,
		Message:  message,
		Metadata: metadata,
	}

	return im.CreateIncident(ctx, request)
}

// CreateIncidentFromJobError creates incident from job error details
// Создает инцидент на основе деталей ошибки job
func (im *IncidentManager) CreateIncidentFromJobError(ctx context.Context, jobKey, jobType, workerID, elementID, processInstanceID, errorMessage string, retries int) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeJobFailure,
		Message:           errorMessage,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		JobKey:            jobKey,
		JobType:           jobType,
		WorkerID:          workerID,
		OriginalRetries:   retries,
		Metadata: map[string]interface{}{
			"failure_source": "job_worker",
			"worker_id":      workerID,
		},
	}

	return im.CreateIncident(ctx, request)
}

// CreateIncidentFromBPMNError creates incident from BPMN error details
// Создает инцидент на основе деталей BPMN ошибки
func (im *IncidentManager) CreateIncidentFromBPMNError(ctx context.Context, elementID, elementType, processInstanceID, processKey, errorCode, errorMessage string) (*Incident, error) {
	request := &CreateIncidentRequest{
		Type:              IncidentTypeBPMNError,
		Message:           errorMessage,
		ErrorCode:         errorCode,
		ProcessInstanceID: processInstanceID,
		ProcessKey:        processKey,
		ElementID:         elementID,
		ElementType:       elementType,
		Metadata: map[string]interface{}{
			"error_source": "bpmn_execution",
			"element_type": elementType,
			"no_boundary":  "true", // Indicates no boundary event was found
		},
	}

	return im.CreateIncident(ctx, request)
}

// CreateIncidentFromTimerError creates incident from timer error details
// Создает инцидент на основе деталей ошибки таймера
func (im *IncidentManager) CreateIncidentFromTimerError(ctx context.Context, timerID, elementID, processInstanceID, timerType, errorMessage string, scheduledAt interface{}) (*Incident, error) {
	metadata := map[string]interface{}{
		"timer_type":   timerType,
		"scheduled_at": scheduledAt,
		"error_source": "timer_execution",
	}

	request := &CreateIncidentRequest{
		Type:              IncidentTypeTimerError,
		Message:           errorMessage,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		TimerID:           timerID,
		Metadata:          metadata,
	}

	return im.CreateIncident(ctx, request)
}

// CreateIncidentFromExpressionError creates incident from expression evaluation error
// Создает инцидент на основе ошибки вычисления выражения
func (im *IncidentManager) CreateIncidentFromExpressionError(ctx context.Context, elementID, processInstanceID, expression, context, errorMessage string) (*Incident, error) {
	metadata := map[string]interface{}{
		"expression":   expression,
		"context":      context,
		"error_source": "expression_evaluation",
	}

	request := &CreateIncidentRequest{
		Type:              IncidentTypeExpressionError,
		Message:           errorMessage,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		Metadata:          metadata,
	}

	return im.CreateIncident(ctx, request)
}
