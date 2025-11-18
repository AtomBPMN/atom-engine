/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package incidents

import (
	"time"
)

// IncidentType represents the type of incident
// Представляет тип инцидента
type IncidentType string

const (
	// Job-related incidents
	IncidentTypeJobFailure IncidentType = "JOB_FAILURE"

	// BPMN-related incidents
	IncidentTypeBPMNError IncidentType = "BPMN_ERROR"

	// Expression-related incidents
	IncidentTypeExpressionError IncidentType = "EXPRESSION_ERROR"

	// Process-related incidents
	IncidentTypeProcessError IncidentType = "PROCESS_ERROR"

	// Timer-related incidents
	IncidentTypeTimerError IncidentType = "TIMER_ERROR"

	// Message-related incidents
	IncidentTypeMessageError IncidentType = "MESSAGE_ERROR"

	// General system incidents
	IncidentTypeSystemError IncidentType = "SYSTEM_ERROR"
)

// IncidentStatus represents the status of an incident
// Представляет статус инцидента
type IncidentStatus string

const (
	IncidentStatusOpen      IncidentStatus = "OPEN"
	IncidentStatusResolved  IncidentStatus = "RESOLVED"
	IncidentStatusDismissed IncidentStatus = "DISMISSED"
)

// ResolveAction represents the action to take when resolving an incident
// Представляет действие при разрешении инцидента
type ResolveAction string

const (
	ResolveActionRetry   ResolveAction = "RETRY"
	ResolveActionDismiss ResolveAction = "DISMISS"
)

// Incident represents a single incident in the system
// Представляет отдельный инцидент в системе
type Incident struct {
	// Core incident fields
	ID        string         `json:"id"`
	Type      IncidentType   `json:"type"`
	Status    IncidentStatus `json:"status"`
	Message   string         `json:"message"`
	ErrorCode string         `json:"error_code,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	// Process context
	ProcessInstanceID string `json:"process_instance_id,omitempty"`
	ProcessKey        string `json:"process_key,omitempty"`
	ElementID         string `json:"element_id,omitempty"`
	ElementType       string `json:"element_type,omitempty"`

	// Job context (for job-related incidents)
	JobKey   string `json:"job_key,omitempty"`
	JobType  string `json:"job_type,omitempty"`
	WorkerID string `json:"worker_id,omitempty"`

	// Timer context (for timer-related incidents)
	TimerID string `json:"timer_id,omitempty"`

	// Message context (for message-related incidents)
	MessageName    string `json:"message_name,omitempty"`
	CorrelationKey string `json:"correlation_key,omitempty"`

	// Resolution fields
	ResolvedAt     *time.Time    `json:"resolved_at,omitempty"`
	ResolvedBy     string        `json:"resolved_by,omitempty"`
	ResolveAction  ResolveAction `json:"resolve_action,omitempty"`
	ResolveComment string        `json:"resolve_comment,omitempty"`

	// Retry context (for job failures)
	OriginalRetries int `json:"original_retries,omitempty"`
	NewRetries      int `json:"new_retries,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentFilter represents filters for incident queries
// Представляет фильтры для запросов инцидентов
type IncidentFilter struct {
	Status            []IncidentStatus `json:"status,omitempty"`
	Type              []IncidentType   `json:"type,omitempty"`
	ProcessInstanceID string           `json:"process_instance_id,omitempty"`
	ProcessKey        string           `json:"process_key,omitempty"`
	ElementID         string           `json:"element_id,omitempty"`
	JobKey            string           `json:"job_key,omitempty"`
	WorkerID          string           `json:"worker_id,omitempty"`
	CreatedAfter      *time.Time       `json:"created_after,omitempty"`
	CreatedBefore     *time.Time       `json:"created_before,omitempty"`
	Limit             int              `json:"limit,omitempty"`
	Offset            int              `json:"offset,omitempty"`
}

// IncidentStats represents incident statistics
// Представляет статистику инцидентов
type IncidentStats struct {
	TotalIncidents     int                    `json:"total_incidents"`
	OpenIncidents      int                    `json:"open_incidents"`
	ResolvedIncidents  int                    `json:"resolved_incidents"`
	DismissedIncidents int                    `json:"dismissed_incidents"`
	IncidentsByType    map[IncidentType]int   `json:"incidents_by_type"`
	IncidentsByStatus  map[IncidentStatus]int `json:"incidents_by_status"`
	RecentIncidents    int                    `json:"recent_incidents_24h"`
}

// CreateIncidentRequest represents a request to create an incident
// Представляет запрос на создание инцидента
type CreateIncidentRequest struct {
	Type              IncidentType           `json:"type"`
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

// ResolveIncidentRequest represents a request to resolve an incident
// Представляет запрос на разрешение инцидента
type ResolveIncidentRequest struct {
	IncidentID string        `json:"incident_id"`
	Action     ResolveAction `json:"action"`
	Comment    string        `json:"comment,omitempty"`
	ResolvedBy string        `json:"resolved_by,omitempty"`
	NewRetries int           `json:"new_retries,omitempty"` // For retry action
}

// NewIncident creates a new incident with default values
// Создает новый инцидент со значениями по умолчанию
func NewIncident(incidentType IncidentType, message string) *Incident {
	now := time.Now()
	return &Incident{
		Type:      incidentType,
		Status:    IncidentStatusOpen,
		Message:   message,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]interface{}),
	}
}

// IsOpen returns true if the incident is open
// Возвращает true если инцидент открыт
func (i *Incident) IsOpen() bool {
	return i.Status == IncidentStatusOpen
}

// IsResolved returns true if the incident is resolved
// Возвращает true если инцидент разрешен
func (i *Incident) IsResolved() bool {
	return i.Status == IncidentStatusResolved
}

// IsDismissed returns true if the incident is dismissed
// Возвращает true если инцидент отклонен
func (i *Incident) IsDismissed() bool {
	return i.Status == IncidentStatusDismissed
}

// Resolve marks the incident as resolved
// Отмечает инцидент как разрешенный
func (i *Incident) Resolve(action ResolveAction, resolvedBy, comment string) {
	now := time.Now()
	i.Status = IncidentStatusResolved
	i.ResolvedAt = &now
	i.ResolvedBy = resolvedBy
	i.ResolveAction = action
	i.ResolveComment = comment
	i.UpdatedAt = now
}

// Dismiss marks the incident as dismissed
// Отмечает инцидент как отклоненный
func (i *Incident) Dismiss(resolvedBy, comment string) {
	now := time.Now()
	i.Status = IncidentStatusDismissed
	i.ResolvedAt = &now
	i.ResolvedBy = resolvedBy
	i.ResolveAction = ResolveActionDismiss
	i.ResolveComment = comment
	i.UpdatedAt = now
}

// IsJobRelated returns true if the incident is related to a job
// Возвращает true если инцидент связан с job
func (i *Incident) IsJobRelated() bool {
	return i.JobKey != ""
}

// IsProcessRelated returns true if the incident is related to a process
// Возвращает true если инцидент связан с процессом
func (i *Incident) IsProcessRelated() bool {
	return i.ProcessInstanceID != ""
}

// GetDisplayName returns a human-readable name for the incident
// Возвращает понятное человеку имя инцидента
func (i *Incident) GetDisplayName() string {
	switch i.Type {
	case IncidentTypeJobFailure:
		if i.JobType != "" {
			return "Job Failure: " + i.JobType
		}
		return "Job Failure"
	case IncidentTypeBPMNError:
		if i.ElementID != "" {
			return "BPMN Error: " + i.ElementID
		}
		return "BPMN Error"
	case IncidentTypeExpressionError:
		return "Expression Error"
	case IncidentTypeProcessError:
		return "Process Error"
	case IncidentTypeTimerError:
		return "Timer Error"
	case IncidentTypeMessageError:
		return "Message Error"
	case IncidentTypeSystemError:
		return "System Error"
	default:
		return string(i.Type)
	}
}
