/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"context"
	"fmt"
)

// Direct API Methods for other components
// Прямые API методы для других компонентов

// checkReady validates component readiness and returns error if not ready
// Проверяет готовность компонента и возвращает ошибку если не готов
func (c *Component) checkReady() error {
	if !c.IsReady() {
		return fmt.Errorf("incidents component is not ready")
	}
	return nil
}

// CreateIncident creates a new incident
// Создает новый инцидент
func (c *Component) CreateIncident(ctx context.Context, request *CreateIncidentRequest) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateIncident(ctx, request)
}

// ResolveIncident resolves an incident
// Разрешает инцидент
func (c *Component) ResolveIncident(ctx context.Context, request *ResolveIncidentRequest) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.ResolveIncident(ctx, request)
}

// GetIncident retrieves an incident by ID
// Получает инцидент по ID
func (c *Component) GetIncident(ctx context.Context, incidentID string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.GetIncident(ctx, incidentID)
}

// ListIncidents lists incidents with filtering
// Получает список инцидентов с фильтрацией
func (c *Component) ListIncidents(ctx context.Context, filter *IncidentFilter) ([]*Incident, int, error) {
	if err := c.checkReady(); err != nil {
		return nil, 0, err
	}
	return c.manager.ListIncidents(ctx, filter)
}

// GetIncidentStats retrieves incident statistics
// Получает статистику инцидентов
func (c *Component) GetIncidentStats(ctx context.Context) (*IncidentStats, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.GetIncidentStats(ctx)
}

// Convenience Methods for creating specific incident types
// Удобные методы для создания специфичных типов инцидентов

// CreateJobFailureIncident creates a job failure incident
// Создает инцидент отказа job
func (c *Component) CreateJobFailureIncident(ctx context.Context, jobKey, elementID, processInstanceID, message string, retries int) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateJobFailureIncident(ctx, jobKey, elementID, processInstanceID, message, retries)
}

// CreateBPMNErrorIncident creates a BPMN error incident
// Создает инцидент BPMN ошибки
func (c *Component) CreateBPMNErrorIncident(ctx context.Context, elementID, processInstanceID, errorCode, message string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateBPMNErrorIncident(ctx, elementID, processInstanceID, errorCode, message)
}

// CreateExpressionErrorIncident creates an expression error incident
// Создает инцидент ошибки выражения
func (c *Component) CreateExpressionErrorIncident(ctx context.Context, elementID, processInstanceID, expression, message string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateExpressionErrorIncident(ctx, elementID, processInstanceID, expression, message)
}

// CreateTimerErrorIncident creates a timer error incident
// Создает инцидент ошибки таймера
func (c *Component) CreateTimerErrorIncident(ctx context.Context, timerID, elementID, processInstanceID, message string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateTimerErrorIncident(ctx, timerID, elementID, processInstanceID, message)
}

// CreateMessageErrorIncident creates a message error incident
// Создает инцидент ошибки сообщения
func (c *Component) CreateMessageErrorIncident(ctx context.Context, messageName, correlationKey, processInstanceID, message string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateMessageErrorIncident(ctx, messageName, correlationKey, processInstanceID, message)
}

// Advanced Convenience Methods for complex scenarios
// Продвинутые удобные методы для сложных сценариев

// CreateIncidentFromJobError creates incident from job error details
// Создает инцидент на основе деталей ошибки job
func (c *Component) CreateIncidentFromJobError(ctx context.Context, jobKey, jobType, workerID, elementID, processInstanceID, errorMessage string, retries int) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateIncidentFromJobError(ctx, jobKey, jobType, workerID, elementID, processInstanceID, errorMessage, retries)
}

// CreateIncidentFromBPMNError creates incident from BPMN error details
// Создает инцидент на основе деталей BPMN ошибки
func (c *Component) CreateIncidentFromBPMNError(ctx context.Context, elementID, elementType, processInstanceID, processKey, errorCode, errorMessage string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateIncidentFromBPMNError(ctx, elementID, elementType, processInstanceID, processKey, errorCode, errorMessage)
}

// CreateIncidentFromTimerError creates incident from timer error details
// Создает инцидент на основе деталей ошибки таймера
func (c *Component) CreateIncidentFromTimerError(ctx context.Context, timerID, elementID, processInstanceID, timerType, errorMessage string, scheduledAt interface{}) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateIncidentFromTimerError(ctx, timerID, elementID, processInstanceID, timerType, errorMessage, scheduledAt)
}

// CreateIncidentFromExpressionError creates incident from expression evaluation error
// Создает инцидент на основе ошибки вычисления выражения
func (c *Component) CreateIncidentFromExpressionError(ctx context.Context, elementID, processInstanceID, expression, context, errorMessage string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.CreateIncidentFromExpressionError(ctx, elementID, processInstanceID, expression, context, errorMessage)
}

// Resolution Methods delegation to manager
// Делегирование методов разрешения в менеджер

// RetryIncident retries an incident
// Повторяет инцидент
func (c *Component) RetryIncident(ctx context.Context, incidentID, resolvedBy string, newRetries int, comment string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.RetryIncident(ctx, incidentID, resolvedBy, newRetries, comment)
}

// DismissIncident dismisses an incident
// Отклоняет инцидент
func (c *Component) DismissIncident(ctx context.Context, incidentID, resolvedBy, comment string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.DismissIncident(ctx, incidentID, resolvedBy, comment)
}

// RetryJobIncident retries a job failure incident with new retries
// Повторяет инцидент отказа job с новым количеством попыток
func (c *Component) RetryJobIncident(ctx context.Context, incidentID, resolvedBy string, newRetries int) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.RetryJobIncident(ctx, incidentID, resolvedBy, newRetries)
}

// DismissBPMNErrorIncident dismisses a BPMN error incident
// Отклоняет инцидент BPMN ошибки
func (c *Component) DismissBPMNErrorIncident(ctx context.Context, incidentID, resolvedBy, reason string) (*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.DismissBPMNErrorIncident(ctx, incidentID, resolvedBy, reason)
}

// Batch Operations delegation to manager
// Делегирование пакетных операций в менеджер

// ResolveByType resolves incidents by type with batch operation
// Разрешает инциденты по типу пакетной операцией
func (c *Component) ResolveByType(ctx context.Context, incidentType IncidentType, action ResolveAction, resolvedBy, comment string, newRetries int) ([]*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.ResolveByType(ctx, incidentType, action, resolvedBy, comment, newRetries)
}

// ResolveByProcessInstance resolves all incidents for a process instance
// Разрешает все инциденты для экземпляра процесса
func (c *Component) ResolveByProcessInstance(ctx context.Context, processInstanceID string, action ResolveAction, resolvedBy, comment string) ([]*Incident, error) {
	if err := c.checkReady(); err != nil {
		return nil, err
	}
	return c.manager.ResolveByProcessInstance(ctx, processInstanceID, action, resolvedBy, comment)
}
