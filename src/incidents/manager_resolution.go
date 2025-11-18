/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"context"
	"encoding/json"
	"fmt"

	"atom-engine/src/core/logger"
)

// Incident resolution operations
// Операции разрешения инцидентов

// RetryIncident retries an incident
// Повторяет инцидент
func (im *IncidentManager) RetryIncident(
	ctx context.Context,
	incidentID, resolvedBy string,
	newRetries int,
	comment string,
) (*Incident, error) {
	request := &ResolveIncidentRequest{
		IncidentID: incidentID,
		Action:     ResolveActionRetry,
		ResolvedBy: resolvedBy,
		NewRetries: newRetries,
		Comment:    comment,
	}

	return im.ResolveIncident(ctx, request)
}

// DismissIncident dismisses an incident
// Отклоняет инцидент
func (im *IncidentManager) DismissIncident(
	ctx context.Context,
	incidentID, resolvedBy, comment string,
) (*Incident, error) {
	request := &ResolveIncidentRequest{
		IncidentID: incidentID,
		Action:     ResolveActionDismiss,
		ResolvedBy: resolvedBy,
		Comment:    comment,
	}

	return im.ResolveIncident(ctx, request)
}

// RetryJobIncident retries a job failure incident with new retries
// Повторяет инцидент отказа job с новым количеством попыток
func (im *IncidentManager) RetryJobIncident(
	ctx context.Context,
	incidentID, resolvedBy string,
	newRetries int,
) (*Incident, error) {
	im.logger.Info("Retrying job incident",
		logger.String("incident_id", incidentID),
		logger.String("resolved_by", resolvedBy),
		logger.Int("new_retries", newRetries))

	// Get incident to validate it's a job failure
	incident, err := im.GetIncident(ctx, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident for retry: %w", err)
	}

	if incident.Type != IncidentTypeJobFailure {
		return nil, fmt.Errorf("incident %s is not a job failure incident", incidentID)
	}

	if incident.JobKey == "" {
		return nil, fmt.Errorf("incident %s has no job key", incidentID)
	}

	comment := fmt.Sprintf("Job retried with %d retries", newRetries)
	return im.RetryIncident(ctx, incidentID, resolvedBy, newRetries, comment)
}

// DismissBPMNErrorIncident dismisses a BPMN error incident
// Отклоняет инцидент BPMN ошибки
func (im *IncidentManager) DismissBPMNErrorIncident(
	ctx context.Context,
	incidentID, resolvedBy, reason string,
) (*Incident, error) {
	im.logger.Info("Dismissing BPMN error incident",
		logger.String("incident_id", incidentID),
		logger.String("resolved_by", resolvedBy),
		logger.String("reason", reason))

	// Get incident to validate it's a BPMN error
	incident, err := im.GetIncident(ctx, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident for dismissal: %w", err)
	}

	if incident.Type != IncidentTypeBPMNError {
		return nil, fmt.Errorf("incident %s is not a BPMN error incident", incidentID)
	}

	comment := fmt.Sprintf("BPMN error dismissed: %s", reason)
	return im.DismissIncident(ctx, incidentID, resolvedBy, comment)
}

// ResolveByType resolves incidents by type with batch operation
// Разрешает инциденты по типу пакетной операцией
func (im *IncidentManager) ResolveByType(
	ctx context.Context,
	incidentType IncidentType,
	action ResolveAction,
	resolvedBy, comment string,
	newRetries int,
) ([]*Incident, error) {
	im.logger.Info("Resolving incidents by type",
		logger.String("type", string(incidentType)),
		logger.String("action", string(action)),
		logger.String("resolved_by", resolvedBy))

	// Get incidents of specified type that are open
	filter := &IncidentFilter{
		Type:   []IncidentType{incidentType},
		Status: []IncidentStatus{IncidentStatusOpen},
	}

	incidents, _, err := im.ListIncidents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents for batch resolution: %w", err)
	}

	var resolvedIncidents []*Incident
	var errors []string

	for _, incident := range incidents {
		request := &ResolveIncidentRequest{
			IncidentID: incident.ID,
			Action:     action,
			ResolvedBy: resolvedBy,
			Comment:    comment,
			NewRetries: newRetries,
		}

		resolvedIncident, err := im.ResolveIncident(ctx, request)
		if err != nil {
			errors = append(errors, fmt.Sprintf("incident %s: %v", incident.ID, err))
			continue
		}

		resolvedIncidents = append(resolvedIncidents, resolvedIncident)
	}

	if len(errors) > 0 {
		im.logger.Warn("Some incidents failed to resolve",
			logger.Int("failed_count", len(errors)),
			logger.Int("success_count", len(resolvedIncidents)))
	}

	im.logger.Info("Batch resolution completed",
		logger.Int("resolved_count", len(resolvedIncidents)),
		logger.Int("failed_count", len(errors)))

	return resolvedIncidents, nil
}

// ResolveByProcessInstance resolves all incidents for a process instance
// Разрешает все инциденты для экземпляра процесса
func (im *IncidentManager) ResolveByProcessInstance(
	ctx context.Context,
	processInstanceID string,
	action ResolveAction,
	resolvedBy, comment string,
) ([]*Incident, error) {
	im.logger.Info("Resolving incidents by process instance",
		logger.String("process_instance_id", processInstanceID),
		logger.String("action", string(action)),
		logger.String("resolved_by", resolvedBy))

	// Get incidents for specified process instance that are open
	filter := &IncidentFilter{
		ProcessInstanceID: processInstanceID,
		Status:            []IncidentStatus{IncidentStatusOpen},
	}

	incidents, _, err := im.ListIncidents(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents for process instance resolution: %w", err)
	}

	if len(incidents) == 0 {
		im.logger.Info("No open incidents found for process instance",
			logger.String("process_instance_id", processInstanceID))
		return []*Incident{}, nil
	}

	var resolvedIncidents []*Incident
	var errors []string

	for _, incident := range incidents {
		request := &ResolveIncidentRequest{
			IncidentID: incident.ID,
			Action:     action,
			ResolvedBy: resolvedBy,
			Comment:    fmt.Sprintf("%s (process instance cleanup)", comment),
		}

		resolvedIncident, err := im.ResolveIncident(ctx, request)
		if err != nil {
			errors = append(errors, fmt.Sprintf("incident %s: %v", incident.ID, err))
			continue
		}

		resolvedIncidents = append(resolvedIncidents, resolvedIncident)
	}

	if len(errors) > 0 {
		im.logger.Warn("Some incidents failed to resolve for process instance",
			logger.String("process_instance_id", processInstanceID),
			logger.Int("failed_count", len(errors)),
			logger.Int("success_count", len(resolvedIncidents)))
	}

	im.logger.Info("Process instance incidents resolution completed",
		logger.String("process_instance_id", processInstanceID),
		logger.Int("resolved_count", len(resolvedIncidents)),
		logger.Int("failed_count", len(errors)))

	return resolvedIncidents, nil
}

// updateJobRetries updates job retries through core interface
// Обновляет retries job через интерфейс core
func (im *IncidentManager) updateJobRetries(ctx context.Context, jobKey string, newRetries int) error {
	if im.core == nil {
		im.logger.Warn("Core interface not set, cannot update job retries")
		return fmt.Errorf("core interface not set")
	}

	im.logger.Info("Updating job retries through core",
		logger.String("job_key", jobKey),
		logger.Int("new_retries", newRetries))

	// Import jobs package at the top of the file if not already done
	// Create payload for update job retries message
	payload := map[string]interface{}{
		"job_key":     jobKey,
		"new_retries": newRetries,
	}

	// Create JSON message
	request := map[string]interface{}{
		"type":    "update_job_retries",
		"payload": payload,
	}

	messageJSON, err := json.Marshal(request)
	if err != nil {
		im.logger.Error("Failed to marshal update job retries message", logger.String("error", err.Error()))
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message to jobs component
	if err := im.core.SendMessage("jobs", string(messageJSON)); err != nil {
		im.logger.Error("Failed to send update job retries message", logger.String("error", err.Error()))
		return fmt.Errorf("failed to send message to jobs component: %w", err)
	}

	im.logger.Info("Job retries update message sent successfully",
		logger.String("job_key", jobKey),
		logger.Int("new_retries", newRetries))

	return nil
}
