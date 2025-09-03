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
)

// Helper functions for type conversion and data manipulation
// Вспомогательные функции для конверсии типов и манипуляции данными

// convertToIncident converts interface{} to *Incident
// Конвертирует interface{} в *Incident
func (im *IncidentManager) convertToIncident(data interface{}) (*Incident, error) {
	if data == nil {
		return nil, nil
	}

	// Try direct type assertion first
	if incident, ok := data.(*Incident); ok {
		return incident, nil
	}

	// Convert through JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incident data: %w", err)
	}

	var incident Incident
	if err := json.Unmarshal(jsonData, &incident); err != nil {
		return nil, fmt.Errorf("failed to unmarshal incident data: %w", err)
	}

	return &incident, nil
}

// convertToIncidentList converts interface{} to []*Incident
// Конвертирует interface{} в []*Incident
func (im *IncidentManager) convertToIncidentList(data interface{}) ([]*Incident, error) {
	if data == nil {
		return nil, nil
	}

	// Try direct type assertion first
	if incidents, ok := data.([]*Incident); ok {
		return incidents, nil
	}

	// Try interface{} slice
	if dataSlice, ok := data.([]interface{}); ok {
		var incidents []*Incident
		for _, item := range dataSlice {
			incident, err := im.convertToIncident(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert incident item: %w", err)
			}
			if incident != nil {
				incidents = append(incidents, incident)
			}
		}
		return incidents, nil
	}

	// Try map slice (from storage)
	if dataSlice, ok := data.([]map[string]interface{}); ok {
		var incidents []*Incident
		for _, item := range dataSlice {
			incident, err := im.convertToIncident(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert incident item: %w", err)
			}
			if incident != nil {
				incidents = append(incidents, incident)
			}
		}
		return incidents, nil
	}

	// Convert through JSON marshalling/unmarshalling
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal incidents data: %w", err)
	}

	var incidents []*Incident
	if err := json.Unmarshal(jsonData, &incidents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal incidents data: %w", err)
	}

	return incidents, nil
}

// validateIncidentRequest validates incident creation request
// Валидирует запрос создания инцидента
func (im *IncidentManager) validateIncidentRequest(request *CreateIncidentRequest) error {
	if request == nil {
		return fmt.Errorf("incident request is required")
	}

	if request.Type == "" {
		return fmt.Errorf("incident type is required")
	}

	if request.Message == "" {
		return fmt.Errorf("incident message is required")
	}

	// Type-specific validation
	switch request.Type {
	case IncidentTypeJobFailure:
		if request.JobKey == "" {
			return fmt.Errorf("job key is required for job failure incidents")
		}
	case IncidentTypeBPMNError:
		if request.ElementID == "" {
			return fmt.Errorf("element ID is required for BPMN error incidents")
		}
	case IncidentTypeTimerError:
		if request.TimerID == "" {
			return fmt.Errorf("timer ID is required for timer error incidents")
		}
	case IncidentTypeMessageError:
		if request.MessageName == "" {
			return fmt.Errorf("message name is required for message error incidents")
		}
	}

	return nil
}

// validateResolveRequest validates incident resolution request
// Валидирует запрос разрешения инцидента
func (im *IncidentManager) validateResolveRequest(request *ResolveIncidentRequest) error {
	if request == nil {
		return fmt.Errorf("resolve request is required")
	}

	if request.IncidentID == "" {
		return fmt.Errorf("incident ID is required")
	}

	if request.Action == "" {
		return fmt.Errorf("resolve action is required")
	}

	if request.ResolvedBy == "" {
		return fmt.Errorf("resolved by is required")
	}

	// Action-specific validation
	switch request.Action {
	case ResolveActionRetry:
		if request.NewRetries < 0 {
			return fmt.Errorf("new retries must be non-negative for retry action")
		}
	case ResolveActionDismiss:
		// No additional validation needed for dismiss
	default:
		return fmt.Errorf("unknown resolve action: %s", request.Action)
	}

	return nil
}

// enrichIncidentMetadata adds additional metadata to incident
// Обогащает метаданные инцидента дополнительной информацией
func (im *IncidentManager) enrichIncidentMetadata(incident *Incident) {
	if incident.Metadata == nil {
		incident.Metadata = make(map[string]interface{})
	}

	// Add creation timestamp as string
	incident.Metadata["created_at_string"] = incident.CreatedAt.Format("2006-01-02 15:04:05")

	// Add display name
	incident.Metadata["display_name"] = incident.GetDisplayName()

	// Add context flags
	incident.Metadata["is_job_related"] = incident.IsJobRelated()
	incident.Metadata["is_process_related"] = incident.IsProcessRelated()

	// Add type-specific metadata
	switch incident.Type {
	case IncidentTypeJobFailure:
		incident.Metadata["can_retry"] = incident.IsOpen()
		incident.Metadata["has_retries"] = incident.OriginalRetries > 0
	case IncidentTypeBPMNError:
		incident.Metadata["has_error_code"] = incident.ErrorCode != ""
		incident.Metadata["error_category"] = im.categorizeErrorCode(incident.ErrorCode)
	case IncidentTypeTimerError:
		incident.Metadata["timer_related"] = true
	case IncidentTypeMessageError:
		incident.Metadata["message_related"] = true
		incident.Metadata["has_correlation"] = incident.CorrelationKey != ""
	}
}

// categorizeErrorCode categorizes error code for better understanding
// Категоризирует код ошибки для лучшего понимания
func (im *IncidentManager) categorizeErrorCode(errorCode string) string {
	switch errorCode {
	case "404":
		return "not_found"
	case "400":
		return "bad_request"
	case "401":
		return "unauthorized"
	case "403":
		return "forbidden"
	case "500":
		return "internal_error"
	case "502":
		return "bad_gateway"
	case "503":
		return "service_unavailable"
	case "504":
		return "gateway_timeout"
	default:
		if errorCode == "" {
			return "unknown"
		}
		return "custom"
	}
}

// buildIncidentSummary builds a human-readable summary of the incident
// Создает читаемое человеком резюме инцидента
func (im *IncidentManager) buildIncidentSummary(incident *Incident) string {
	summary := fmt.Sprintf("%s in ", incident.GetDisplayName())

	if incident.ProcessInstanceID != "" {
		summary += fmt.Sprintf("process %s", incident.ProcessInstanceID)
	} else {
		summary += "system"
	}

	if incident.ElementID != "" {
		summary += fmt.Sprintf(" at element %s", incident.ElementID)
	}

	if incident.ErrorCode != "" {
		summary += fmt.Sprintf(" (error: %s)", incident.ErrorCode)
	}

	return summary
}

// extractContextFromMetadata extracts relevant context from incident metadata
// Извлекает релевантный контекст из метаданных инцидента
func (im *IncidentManager) extractContextFromMetadata(incident *Incident) map[string]string {
	context := make(map[string]string)

	if incident.Metadata == nil {
		return context
	}

	// Extract string values from metadata
	for key, value := range incident.Metadata {
		if strValue, ok := value.(string); ok {
			context[key] = strValue
		}
	}

	return context
}

// sanitizeIncidentData sanitizes incident data for safe storage/transmission
// Очищает данные инцидента для безопасного хранения/передачи
func (im *IncidentManager) sanitizeIncidentData(incident *Incident) {
	// Truncate message if too long
	maxMessageLength := 1000
	if len(incident.Message) > maxMessageLength {
		incident.Message = incident.Message[:maxMessageLength] + "... [truncated]"
	}

	// Sanitize metadata
	if incident.Metadata != nil {
		for key, value := range incident.Metadata {
			if strValue, ok := value.(string); ok {
				if len(strValue) > 500 { // Limit metadata string values
					incident.Metadata[key] = strValue[:500] + "... [truncated]"
				}
			}
		}
	}

	// Ensure error code is uppercase
	if incident.ErrorCode != "" {
		// Keep original case for now, as error codes can be case-sensitive
	}
}
