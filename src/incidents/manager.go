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
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// IncidentManagerInterface defines the interface for incident management
// Определяет интерфейс для управления инцидентами
type IncidentManagerInterface interface {
	// Core incident operations
	CreateIncident(ctx context.Context, request *CreateIncidentRequest) (*Incident, error)
	ResolveIncident(ctx context.Context, request *ResolveIncidentRequest) (*Incident, error)
	GetIncident(ctx context.Context, incidentID string) (*Incident, error)
	ListIncidents(ctx context.Context, filter *IncidentFilter) ([]*Incident, int, error)
	GetIncidentStats(ctx context.Context) (*IncidentStats, error)

	// Specialized creation methods for common incident types
	CreateJobFailureIncident(ctx context.Context, jobKey, elementID, processInstanceID, message string, retries int) (*Incident, error)
	CreateBPMNErrorIncident(ctx context.Context, elementID, processInstanceID, errorCode, message string) (*Incident, error)
	CreateExpressionErrorIncident(ctx context.Context, elementID, processInstanceID, expression, message string) (*Incident, error)
	CreateTimerErrorIncident(ctx context.Context, timerID, elementID, processInstanceID, message string) (*Incident, error)
	CreateMessageErrorIncident(ctx context.Context, messageName, correlationKey, processInstanceID, message string) (*Incident, error)

	// Incident resolution operations
	RetryIncident(ctx context.Context, incidentID, resolvedBy string, newRetries int, comment string) (*Incident, error)
	DismissIncident(ctx context.Context, incidentID, resolvedBy, comment string) (*Incident, error)

	// Advanced creation methods with detailed context
	CreateIncidentFromJobError(ctx context.Context, jobKey, jobType, workerID, elementID, processInstanceID, errorMessage string, retries int) (*Incident, error)
	CreateIncidentFromBPMNError(ctx context.Context, elementID, elementType, processInstanceID, processKey, errorCode, errorMessage string) (*Incident, error)
	CreateIncidentFromTimerError(ctx context.Context, timerID, elementID, processInstanceID, timerType, errorMessage string, scheduledAt interface{}) (*Incident, error)
	CreateIncidentFromExpressionError(ctx context.Context, elementID, processInstanceID, expression, context, errorMessage string) (*Incident, error)

	// Advanced resolution operations
	RetryJobIncident(ctx context.Context, incidentID, resolvedBy string, newRetries int) (*Incident, error)
	DismissBPMNErrorIncident(ctx context.Context, incidentID, resolvedBy, reason string) (*Incident, error)
	ResolveByType(ctx context.Context, incidentType IncidentType, action ResolveAction, resolvedBy, comment string, newRetries int) ([]*Incident, error)
	ResolveByProcessInstance(ctx context.Context, processInstanceID string, action ResolveAction, resolvedBy, comment string) ([]*Incident, error)
}

// IncidentManager implements incident management operations
// Реализует операции управления инцидентами
type IncidentManager struct {
	storage storage.Storage
	logger  logger.ComponentLogger
}

// NewIncidentManager creates new incident manager
// Создает новый менеджер инцидентов
func NewIncidentManager(storage storage.Storage) *IncidentManager {
	return &IncidentManager{
		storage: storage,
		logger:  logger.NewComponentLogger("incident-manager"),
	}
}

// CreateIncident creates a new incident
// Создает новый инцидент
func (im *IncidentManager) CreateIncident(ctx context.Context, request *CreateIncidentRequest) (*Incident, error) {
	// Validate request
	if err := im.validateIncidentRequest(request); err != nil {
		return nil, fmt.Errorf("invalid incident request: %w", err)
	}

	im.logger.Info("Creating incident",
		logger.String("type", string(request.Type)),
		logger.String("message", request.Message),
		logger.String("process_instance_id", request.ProcessInstanceID),
		logger.String("element_id", request.ElementID))

	// Generate unique incident ID
	incidentID := models.GenerateID()

	// Create incident
	incident := NewIncident(request.Type, request.Message)
	incident.ID = incidentID
	incident.ErrorCode = request.ErrorCode
	incident.ProcessInstanceID = request.ProcessInstanceID
	incident.ProcessKey = request.ProcessKey
	incident.ElementID = request.ElementID
	incident.ElementType = request.ElementType
	incident.JobKey = request.JobKey
	incident.JobType = request.JobType
	incident.WorkerID = request.WorkerID
	incident.TimerID = request.TimerID
	incident.MessageName = request.MessageName
	incident.CorrelationKey = request.CorrelationKey
	incident.OriginalRetries = request.OriginalRetries

	// Copy metadata
	if request.Metadata != nil {
		incident.Metadata = make(map[string]interface{})
		for k, v := range request.Metadata {
			incident.Metadata[k] = v
		}
	}

	// Enrich incident metadata
	im.enrichIncidentMetadata(incident)

	// Sanitize data before storage
	im.sanitizeIncidentData(incident)

	// Save to storage
	if err := im.storage.SaveIncident(incident); err != nil {
		im.logger.Error("Failed to save incident",
			logger.String("incident_id", incidentID),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to save incident: %w", err)
	}

	im.logger.Info("Incident created successfully",
		logger.String("incident_id", incidentID),
		logger.String("type", string(request.Type)))

	return incident, nil
}

// ResolveIncident resolves an incident
// Разрешает инцидент
func (im *IncidentManager) ResolveIncident(ctx context.Context, request *ResolveIncidentRequest) (*Incident, error) {
	// Validate request
	if err := im.validateResolveRequest(request); err != nil {
		return nil, fmt.Errorf("invalid resolve request: %w", err)
	}

	im.logger.Info("Resolving incident",
		logger.String("incident_id", request.IncidentID),
		logger.String("action", string(request.Action)),
		logger.String("resolved_by", request.ResolvedBy))

	// Get incident
	incidentData, err := im.storage.GetIncident(request.IncidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	if incidentData == nil {
		return nil, fmt.Errorf("incident not found: %s", request.IncidentID)
	}

	// Convert to Incident struct
	incident, err := im.convertToIncident(incidentData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert incident data: %w", err)
	}

	if !incident.IsOpen() {
		return nil, fmt.Errorf("incident is already resolved: %s", request.IncidentID)
	}

	// Resolve incident based on action
	switch request.Action {
	case ResolveActionRetry:
		incident.Resolve(ResolveActionRetry, request.ResolvedBy, request.Comment)
		incident.NewRetries = request.NewRetries

		// If this is a job failure incident, update job retries
		if incident.Type == IncidentTypeJobFailure && incident.JobKey != "" {
			if err := im.updateJobRetries(ctx, incident.JobKey, request.NewRetries); err != nil {
				im.logger.Warn("Failed to update job retries",
					logger.String("job_key", incident.JobKey),
					logger.Int("new_retries", request.NewRetries),
					logger.String("error", err.Error()))
				// Continue with incident resolution even if job update fails
			}
		}

	case ResolveActionDismiss:
		incident.Dismiss(request.ResolvedBy, request.Comment)

	default:
		return nil, fmt.Errorf("unknown resolve action: %s", request.Action)
	}

	// Save updated incident
	if err := im.storage.SaveIncident(incident); err != nil {
		return nil, fmt.Errorf("failed to save resolved incident: %w", err)
	}

	im.logger.Info("Incident resolved successfully",
		logger.String("incident_id", request.IncidentID),
		logger.String("action", string(request.Action)))

	return incident, nil
}

// GetIncident retrieves an incident by ID
// Получает инцидент по ID
func (im *IncidentManager) GetIncident(ctx context.Context, incidentID string) (*Incident, error) {
	incidentData, err := im.storage.GetIncident(incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	if incidentData == nil {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	return im.convertToIncident(incidentData)
}

// ListIncidents lists incidents with filtering
// Получает список инцидентов с фильтрацией
func (im *IncidentManager) ListIncidents(ctx context.Context, filter *IncidentFilter) ([]*Incident, int, error) {
	incidentsData, total, err := im.storage.ListIncidents(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list incidents: %w", err)
	}

	incidents, err := im.convertToIncidentList(incidentsData)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert incidents data: %w", err)
	}

	return incidents, total, nil
}

// GetIncidentStats retrieves incident statistics
// Получает статистику инцидентов
func (im *IncidentManager) GetIncidentStats(ctx context.Context) (*IncidentStats, error) {
	// Get all incidents for stats calculation
	filter := &IncidentFilter{Limit: 0} // No limit for stats
	incidentsData, total, err := im.storage.ListIncidents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get incidents for stats: %w", err)
	}

	incidents, err := im.convertToIncidentList(incidentsData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert incidents data: %w", err)
	}

	stats := &IncidentStats{
		TotalIncidents:    total,
		IncidentsByType:   make(map[IncidentType]int),
		IncidentsByStatus: make(map[IncidentStatus]int),
	}

	// Calculate stats
	recentThreshold := time.Now().Add(-24 * time.Hour)

	for _, incident := range incidents {
		// Count by status
		switch incident.Status {
		case IncidentStatusOpen:
			stats.OpenIncidents++
		case IncidentStatusResolved:
			stats.ResolvedIncidents++
		case IncidentStatusDismissed:
			stats.DismissedIncidents++
		}

		// Count by type
		stats.IncidentsByType[incident.Type]++

		// Count by status for map
		stats.IncidentsByStatus[incident.Status]++

		// Count recent incidents
		if incident.CreatedAt.After(recentThreshold) {
			stats.RecentIncidents++
		}
	}

	return stats, nil
}

// Core CRUD operations are implemented above
// Specialized methods are in manager_creation.go
// Resolution methods are in manager_resolution.go
// Helper functions are in manager_helpers.go
