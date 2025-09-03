/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/proto/incidents/incidentspb"
	"atom-engine/src/core/logger"
	"atom-engine/src/incidents"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// incidentsServiceServer implements incidents gRPC service
type incidentsServiceServer struct {
	incidentspb.UnimplementedIncidentsServiceServer
	core CoreInterface
}

// getIncidentsComponent helper function for direct component access
// helper функция для прямого доступа к компоненту incidents
func getIncidentsComponent(core CoreInterface) (*incidents.Component, error) {
	componentIf := core.GetIncidentsComponent()
	if componentIf == nil {
		return nil, fmt.Errorf("incidents component not available")
	}

	component, ok := componentIf.(*incidents.Component)
	if !ok {
		return nil, fmt.Errorf("incidents component type assertion failed")
	}

	return component, nil
}

// CreateIncident creates a new incident
func (s *incidentsServiceServer) CreateIncident(ctx context.Context, req *incidentspb.CreateIncidentRequest) (*incidentspb.CreateIncidentResponse, error) {
	logger.Info("CreateIncident gRPC request",
		logger.String("type", req.Type.String()),
		logger.String("message", req.Message),
		logger.String("process_instance_id", req.ProcessInstanceId))

	// Convert protobuf metadata to map
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create JSON message for incidents component
	payload := incidents.CreateIncidentPayload{
		Type:              convertProtoIncidentType(req.Type),
		Message:           req.Message,
		ErrorCode:         req.ErrorCode,
		ProcessInstanceID: req.ProcessInstanceId,
		ProcessKey:        req.ProcessKey,
		ElementID:         req.ElementId,
		ElementType:       req.ElementType,
		JobKey:            req.JobKey,
		JobType:           req.JobType,
		WorkerID:          req.WorkerId,
		TimerID:           req.TimerId,
		MessageName:       req.MessageName,
		CorrelationKey:    req.CorrelationKey,
		OriginalRetries:   int(req.OriginalRetries),
		Metadata:          metadata,
	}

	message, err := incidents.CreateIncidentMessage(payload)
	if err != nil {
		logger.Error("Failed to create incident message", logger.String("error", err.Error()))
		return &incidentspb.CreateIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to create incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	if err := s.core.SendMessage("incidents", message); err != nil {
		logger.Error("Failed to send incident message", logger.String("error", err.Error()))
		return &incidentspb.CreateIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to send incident message: %w", err)
	}

	// For now, we return success immediately
	// In a more sophisticated implementation, we'd wait for response
	logger.Info("Incident creation request sent successfully")

	// Create a basic response (we'll improve this when we add response handling)
	response := &incidentspb.CreateIncidentResponse{
		Incident: &incidentspb.Incident{
			Type:              req.Type,
			Status:            incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN,
			Message:           req.Message,
			ErrorCode:         req.ErrorCode,
			ProcessInstanceId: req.ProcessInstanceId,
			ProcessKey:        req.ProcessKey,
			ElementId:         req.ElementId,
			ElementType:       req.ElementType,
			JobKey:            req.JobKey,
			JobType:           req.JobType,
			WorkerId:          req.WorkerId,
			TimerId:           req.TimerId,
			MessageName:       req.MessageName,
			CorrelationKey:    req.CorrelationKey,
			OriginalRetries:   req.OriginalRetries,
			Metadata:          req.Metadata,
		},
	}

	return response, nil
}

// ResolveIncident resolves an incident
func (s *incidentsServiceServer) ResolveIncident(ctx context.Context, req *incidentspb.ResolveIncidentRequest) (*incidentspb.ResolveIncidentResponse, error) {
	logger.Info("ResolveIncident gRPC request",
		logger.String("incident_id", req.IncidentId),
		logger.String("action", req.Action.String()),
		logger.String("resolved_by", req.ResolvedBy))

	// Create JSON message for incidents component
	payload := incidents.ResolveIncidentPayload{
		IncidentID: req.IncidentId,
		Action:     convertProtoResolveAction(req.Action),
		Comment:    req.Comment,
		ResolvedBy: req.ResolvedBy,
		NewRetries: int(req.NewRetries),
	}

	message, err := incidents.CreateResolveIncidentMessage(payload)
	if err != nil {
		logger.Error("Failed to create resolve incident message", logger.String("error", err.Error()))
		return &incidentspb.ResolveIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to create resolve incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	if err := s.core.SendMessage("incidents", message); err != nil {
		logger.Error("Failed to send resolve incident message", logger.String("error", err.Error()))
		return &incidentspb.ResolveIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to send resolve incident message: %w", err)
	}

	logger.Info("Incident resolution request sent successfully")

	// Create a basic response
	response := &incidentspb.ResolveIncidentResponse{
		Incident: &incidentspb.Incident{
			Id:         req.IncidentId,
			Status:     convertActionToStatus(req.Action),
			ResolvedBy: req.ResolvedBy,
			NewRetries: req.NewRetries,
		},
	}

	return response, nil
}

// GetIncident retrieves an incident by ID
func (s *incidentsServiceServer) GetIncident(ctx context.Context, req *incidentspb.GetIncidentRequest) (*incidentspb.GetIncidentResponse, error) {
	logger.Info("GetIncident gRPC request",
		logger.String("incident_id", req.IncidentId))

	// Create JSON message for incidents component
	payload := incidents.GetIncidentPayload{
		IncidentID: req.IncidentId,
	}

	message, err := incidents.CreateGetIncidentMessage(payload)
	if err != nil {
		logger.Error("Failed to create get incident message", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to create get incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	if err := s.core.SendMessage("incidents", message); err != nil {
		logger.Error("Failed to send get incident message", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to send get incident message: %w", err)
	}

	logger.Info("Get incident request sent successfully")

	// Wait for response from incidents component
	// Ожидаем ответ от компонента incidents
	responseJSON, err := s.core.WaitForIncidentsResponse(5000) // 5 second timeout
	if err != nil {
		logger.Error("Failed to get incidents response", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to get incidents response: %w", err)
	}

	logger.Debug("Received incidents response", logger.String("response", responseJSON))

	// Parse JSON response
	var response struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
		Data    struct {
			ID                string                 `json:"id"`
			Type              string                 `json:"type"`
			Status            string                 `json:"status"`
			Message           string                 `json:"message"`
			ErrorCode         string                 `json:"error_code"`
			ProcessInstanceID string                 `json:"process_instance_id"`
			ProcessKey        string                 `json:"process_key"`
			ElementID         string                 `json:"element_id"`
			ElementType       string                 `json:"element_type"`
			JobKey            string                 `json:"job_key"`
			JobType           string                 `json:"job_type"`
			WorkerID          string                 `json:"worker_id"`
			TimerID           string                 `json:"timer_id"`
			MessageName       string                 `json:"message_name"`
			CorrelationKey    string                 `json:"correlation_key"`
			CreatedAt         string                 `json:"created_at"`
			UpdatedAt         string                 `json:"updated_at"`
			ResolvedAt        *string                `json:"resolved_at"`
			ResolvedBy        string                 `json:"resolved_by"`
			OriginalRetries   int                    `json:"original_retries"`
			NewRetries        int                    `json:"new_retries"`
			Metadata          map[string]interface{} `json:"metadata"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		logger.Error("Failed to parse incident response", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("failed to parse incident response: %w", err)
	}

	if !response.Success {
		return &incidentspb.GetIncidentResponse{
			Incident: nil,
		}, fmt.Errorf("incident request failed")
	}

	// Convert to protobuf incident
	incident := &incidentspb.Incident{
		Id:                response.Data.ID,
		Type:              convertStringToIncidentType(response.Data.Type),
		Status:            convertStringToIncidentStatus(response.Data.Status),
		Message:           response.Data.Message,
		ErrorCode:         response.Data.ErrorCode,
		ProcessInstanceId: response.Data.ProcessInstanceID,
		ProcessKey:        response.Data.ProcessKey,
		ElementId:         response.Data.ElementID,
		ElementType:       response.Data.ElementType,
		JobKey:            response.Data.JobKey,
		JobType:           response.Data.JobType,
		WorkerId:          response.Data.WorkerID,
		TimerId:           response.Data.TimerID,
		MessageName:       response.Data.MessageName,
		CorrelationKey:    response.Data.CorrelationKey,
		OriginalRetries:   int32(response.Data.OriginalRetries),
		NewRetries:        int32(response.Data.NewRetries),
		ResolvedBy:        response.Data.ResolvedBy,
	}

	// Convert metadata
	incident.Metadata = make(map[string]string)
	for k, v := range response.Data.Metadata {
		if str, ok := v.(string); ok {
			incident.Metadata[k] = str
		}
	}

	// Parse timestamps if available
	if response.Data.CreatedAt != "" {
		if ts, err := parseTimestamp(response.Data.CreatedAt); err == nil {
			incident.CreatedAt = ts
		}
	}
	if response.Data.UpdatedAt != "" {
		if ts, err := parseTimestamp(response.Data.UpdatedAt); err == nil {
			incident.UpdatedAt = ts
		}
	}
	if response.Data.ResolvedAt != nil && *response.Data.ResolvedAt != "" {
		if ts, err := parseTimestamp(*response.Data.ResolvedAt); err == nil {
			incident.ResolvedAt = ts
		}
	}

	return &incidentspb.GetIncidentResponse{
		Incident: incident,
	}, nil
}

// ListIncidents lists incidents with filtering
func (s *incidentsServiceServer) ListIncidents(ctx context.Context, req *incidentspb.ListIncidentsRequest) (*incidentspb.ListIncidentsResponse, error) {
	logger.Info("ListIncidents gRPC request")

	// Create JSON message for incidents component
	filter := req.Filter
	payload := incidents.ListIncidentsPayload{
		Status:            convertProtoIncidentStatusArray(filter.Status),
		Type:              convertProtoIncidentTypeArray(filter.Type),
		ProcessInstanceID: filter.ProcessInstanceId,
		ProcessKey:        filter.ProcessKey,
		ElementID:         filter.ElementId,
		JobKey:            filter.JobKey,
		WorkerID:          filter.WorkerId,
		Limit:             int(filter.Limit),
		Offset:            int(filter.Offset),
	}

	message, err := incidents.CreateListIncidentsMessage(payload)
	if err != nil {
		logger.Error("Failed to create list incidents message", logger.String("error", err.Error()))
		return &incidentspb.ListIncidentsResponse{
			Incidents: nil,
			Total:     0,
		}, fmt.Errorf("failed to create list incidents message: %w", err)
	}

	// Send JSON message to incidents component through Core
	if err := s.core.SendMessage("incidents", message); err != nil {
		logger.Error("Failed to send list incidents message", logger.String("error", err.Error()))
		return &incidentspb.ListIncidentsResponse{
			Incidents: nil,
			Total:     0,
		}, fmt.Errorf("failed to send list incidents message: %w", err)
	}

	logger.Info("List incidents request sent successfully")

	// Wait for response from incidents component
	// Ожидаем ответ от компонента incidents
	responseJSON, err := s.core.WaitForIncidentsResponse(5000) // 5 second timeout
	if err != nil {
		logger.Error("Failed to get incidents response", logger.String("error", err.Error()))
		return &incidentspb.ListIncidentsResponse{
			Incidents: nil,
			Total:     0,
		}, fmt.Errorf("failed to get incidents response: %w", err)
	}

	logger.Debug("Received incidents response", logger.String("response", responseJSON))

	// Parse JSON response
	var response struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
		Data    struct {
			Incidents []struct {
				ID                string                 `json:"id"`
				Type              string                 `json:"type"`
				Status            string                 `json:"status"`
				Message           string                 `json:"message"`
				ErrorCode         string                 `json:"error_code"`
				ProcessInstanceID string                 `json:"process_instance_id"`
				ProcessKey        string                 `json:"process_key"`
				ElementID         string                 `json:"element_id"`
				ElementType       string                 `json:"element_type"`
				JobKey            string                 `json:"job_key"`
				JobType           string                 `json:"job_type"`
				WorkerID          string                 `json:"worker_id"`
				TimerID           string                 `json:"timer_id"`
				MessageName       string                 `json:"message_name"`
				CorrelationKey    string                 `json:"correlation_key"`
				CreatedAt         string                 `json:"created_at"`
				UpdatedAt         string                 `json:"updated_at"`
				ResolvedAt        *string                `json:"resolved_at"`
				ResolvedBy        string                 `json:"resolved_by"`
				OriginalRetries   int                    `json:"original_retries"`
				NewRetries        int                    `json:"new_retries"`
				Metadata          map[string]interface{} `json:"metadata"`
			} `json:"incidents"`
			Total int `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		logger.Error("Failed to parse incidents response", logger.String("error", err.Error()))
		return &incidentspb.ListIncidentsResponse{
			Incidents: nil,
			Total:     0,
		}, fmt.Errorf("failed to parse incidents response: %w", err)
	}

	if !response.Success {
		return &incidentspb.ListIncidentsResponse{
			Incidents: nil,
			Total:     0,
		}, fmt.Errorf("incidents request failed")
	}

	// Convert to protobuf incidents
	var protoIncidents []*incidentspb.Incident
	for _, incident := range response.Data.Incidents {
		protoIncident := &incidentspb.Incident{
			Id:                incident.ID,
			Type:              convertStringToIncidentType(incident.Type),
			Status:            convertStringToIncidentStatus(incident.Status),
			Message:           incident.Message,
			ErrorCode:         incident.ErrorCode,
			ProcessInstanceId: incident.ProcessInstanceID,
			ProcessKey:        incident.ProcessKey,
			ElementId:         incident.ElementID,
			ElementType:       incident.ElementType,
			JobKey:            incident.JobKey,
			JobType:           incident.JobType,
			WorkerId:          incident.WorkerID,
			TimerId:           incident.TimerID,
			MessageName:       incident.MessageName,
			CorrelationKey:    incident.CorrelationKey,
			OriginalRetries:   int32(incident.OriginalRetries),
			NewRetries:        int32(incident.NewRetries),
		}

		// Convert metadata
		protoIncident.Metadata = make(map[string]string)
		for k, v := range incident.Metadata {
			if str, ok := v.(string); ok {
				protoIncident.Metadata[k] = str
			}
		}

		// Parse timestamps if available
		if incident.CreatedAt != "" {
			if ts, err := parseTimestamp(incident.CreatedAt); err == nil {
				protoIncident.CreatedAt = ts
			}
		}
		if incident.UpdatedAt != "" {
			if ts, err := parseTimestamp(incident.UpdatedAt); err == nil {
				protoIncident.UpdatedAt = ts
			}
		}
		if incident.ResolvedAt != nil && *incident.ResolvedAt != "" {
			if ts, err := parseTimestamp(*incident.ResolvedAt); err == nil {
				protoIncident.ResolvedAt = ts
			}
		}

		protoIncidents = append(protoIncidents, protoIncident)
	}

	return &incidentspb.ListIncidentsResponse{
		Incidents: protoIncidents,
		Total:     int32(response.Data.Total),
	}, nil
}

// GetIncidentStats retrieves incident statistics
func (s *incidentsServiceServer) GetIncidentStats(ctx context.Context, req *incidentspb.GetIncidentStatsRequest) (*incidentspb.GetIncidentStatsResponse, error) {
	logger.Info("GetIncidentStats gRPC request")

	// Create JSON message for incidents component
	message, err := incidents.CreateGetIncidentStatsMessage()
	if err != nil {
		logger.Error("Failed to create get incident stats message", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentStatsResponse{
			Stats: nil,
		}, fmt.Errorf("failed to create get incident stats message: %w", err)
	}

	// Send JSON message to incidents component through Core
	if err := s.core.SendMessage("incidents", message); err != nil {
		logger.Error("Failed to send get incident stats message", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentStatsResponse{
			Stats: nil,
		}, fmt.Errorf("failed to send get incident stats message: %w", err)
	}

	logger.Info("Get incident stats request sent successfully")

	// Wait for response from incidents component
	// Ожидаем ответ от компонента incidents
	responseJSON, err := s.core.WaitForIncidentsResponse(5000) // 5 second timeout
	if err != nil {
		logger.Error("Failed to get incidents response", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentStatsResponse{
			Stats: nil,
		}, fmt.Errorf("failed to get incidents response: %w", err)
	}

	logger.Debug("Received incidents response", logger.String("response", responseJSON))

	// Parse JSON response
	var response struct {
		Type    string `json:"type"`
		Success bool   `json:"success"`
		Data    struct {
			TotalIncidents     int            `json:"total_incidents"`
			OpenIncidents      int            `json:"open_incidents"`
			ResolvedIncidents  int            `json:"resolved_incidents"`
			DismissedIncidents int            `json:"dismissed_incidents"`
			RecentIncidents24H int            `json:"recent_incidents_24h"`
			IncidentsByType    map[string]int `json:"incidents_by_type"`
			IncidentsByStatus  map[string]int `json:"incidents_by_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		logger.Error("Failed to parse incident stats response", logger.String("error", err.Error()))
		return &incidentspb.GetIncidentStatsResponse{
			Stats: nil,
		}, fmt.Errorf("failed to parse incident stats response: %w", err)
	}

	if !response.Success {
		return &incidentspb.GetIncidentStatsResponse{
			Stats: nil,
		}, fmt.Errorf("incident stats request failed")
	}

	// Convert to protobuf stats
	stats := &incidentspb.IncidentStats{
		TotalIncidents:      int32(response.Data.TotalIncidents),
		OpenIncidents:       int32(response.Data.OpenIncidents),
		ResolvedIncidents:   int32(response.Data.ResolvedIncidents),
		DismissedIncidents:  int32(response.Data.DismissedIncidents),
		RecentIncidents_24H: int32(response.Data.RecentIncidents24H),
		IncidentsByType:     make(map[string]int32),
		IncidentsByStatus:   make(map[string]int32),
	}

	// Convert type stats
	for typeStr, count := range response.Data.IncidentsByType {
		stats.IncidentsByType[typeStr] = int32(count)
	}

	// Convert status stats
	for statusStr, count := range response.Data.IncidentsByStatus {
		stats.IncidentsByStatus[statusStr] = int32(count)
	}

	return &incidentspb.GetIncidentStatsResponse{
		Stats: stats,
	}, nil
}

// Helper functions for protobuf conversion

// convertProtoIncidentType converts protobuf incident type to string
func convertProtoIncidentType(protoType incidentspb.IncidentType) string {
	switch protoType {
	case incidentspb.IncidentType_INCIDENT_TYPE_JOB_FAILURE:
		return "JOB_FAILURE"
	case incidentspb.IncidentType_INCIDENT_TYPE_BPMN_ERROR:
		return "BPMN_ERROR"
	case incidentspb.IncidentType_INCIDENT_TYPE_EXPRESSION_ERROR:
		return "EXPRESSION_ERROR"
	case incidentspb.IncidentType_INCIDENT_TYPE_PROCESS_ERROR:
		return "PROCESS_ERROR"
	case incidentspb.IncidentType_INCIDENT_TYPE_TIMER_ERROR:
		return "TIMER_ERROR"
	case incidentspb.IncidentType_INCIDENT_TYPE_MESSAGE_ERROR:
		return "MESSAGE_ERROR"
	case incidentspb.IncidentType_INCIDENT_TYPE_SYSTEM_ERROR:
		return "SYSTEM_ERROR"
	default:
		return "SYSTEM_ERROR"
	}
}

// convertProtoResolveAction converts protobuf resolve action to string
func convertProtoResolveAction(protoAction incidentspb.ResolveAction) string {
	switch protoAction {
	case incidentspb.ResolveAction_RESOLVE_ACTION_RETRY:
		return "RETRY"
	case incidentspb.ResolveAction_RESOLVE_ACTION_DISMISS:
		return "DISMISS"
	default:
		return "DISMISS"
	}
}

// convertActionToStatus converts resolve action to incident status
func convertActionToStatus(action incidentspb.ResolveAction) incidentspb.IncidentStatus {
	switch action {
	case incidentspb.ResolveAction_RESOLVE_ACTION_RETRY:
		return incidentspb.IncidentStatus_INCIDENT_STATUS_RESOLVED
	case incidentspb.ResolveAction_RESOLVE_ACTION_DISMISS:
		return incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED
	default:
		return incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED
	}
}

// convertProtoIncidentStatusArray converts protobuf status array to string array
func convertProtoIncidentStatusArray(protoStatuses []incidentspb.IncidentStatus) []string {
	var statuses []string
	for _, status := range protoStatuses {
		switch status {
		case incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN:
			statuses = append(statuses, "OPEN")
		case incidentspb.IncidentStatus_INCIDENT_STATUS_RESOLVED:
			statuses = append(statuses, "RESOLVED")
		case incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED:
			statuses = append(statuses, "DISMISSED")
		}
	}
	return statuses
}

// convertProtoIncidentTypeArray converts protobuf type array to string array
func convertProtoIncidentTypeArray(protoTypes []incidentspb.IncidentType) []string {
	var types []string
	for _, incidentType := range protoTypes {
		types = append(types, convertProtoIncidentType(incidentType))
	}
	return types
}

// convertStringToIncidentType converts string to protobuf incident type
func convertStringToIncidentType(typeStr string) incidentspb.IncidentType {
	switch typeStr {
	case "JOB_FAILURE":
		return incidentspb.IncidentType_INCIDENT_TYPE_JOB_FAILURE
	case "BPMN_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_BPMN_ERROR
	case "EXPRESSION_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_EXPRESSION_ERROR
	case "PROCESS_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_PROCESS_ERROR
	case "TIMER_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_TIMER_ERROR
	case "MESSAGE_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_MESSAGE_ERROR
	case "SYSTEM_ERROR":
		return incidentspb.IncidentType_INCIDENT_TYPE_SYSTEM_ERROR
	default:
		return incidentspb.IncidentType_INCIDENT_TYPE_SYSTEM_ERROR
	}
}

// convertStringToIncidentStatus converts string to protobuf incident status
func convertStringToIncidentStatus(statusStr string) incidentspb.IncidentStatus {
	switch statusStr {
	case "OPEN":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN
	case "RESOLVED":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_RESOLVED
	case "DISMISSED":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED
	default:
		return incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN
	}
}

// parseTimestamp converts string timestamp to protobuf timestamp
func parseTimestamp(timestampStr string) (*timestamppb.Timestamp, error) {
	// Try parsing common timestamp formats
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			return timestamppb.New(t), nil
		}
	}

	return nil, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}
