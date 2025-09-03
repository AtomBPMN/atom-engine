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

// JSON Communication Implementation
// Реализация JSON коммуникации

// processMessages processes incoming JSON messages
// Обрабатывает входящие JSON сообщения
func (c *Component) processMessages() {
	c.logger.Info("Started incidents message processing")

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Stopping incidents message processing")
			return
		case message := <-c.requestChannel:
			c.handleMessage(c.ctx, message)
		}
	}
}

// handleMessage handles single JSON message
// Обрабатывает одно JSON сообщение
func (c *Component) handleMessage(ctx context.Context, message string) {
	var request IncidentRequest
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		c.logger.Error("Failed to parse incident request",
			logger.String("message", message),
			logger.String("error", err.Error()))
		response := CreateIncidentErrorResponse("error_response", request.RequestID, fmt.Sprintf("invalid JSON: %v", err))
		c.sendResponse(response)
		return
	}

	c.logger.Debug("Processing incident request",
		logger.String("type", request.Type),
		logger.String("request_id", request.RequestID))

	switch request.Type {
	case "create_incident":
		c.handleCreateIncident(ctx, request)
	case "resolve_incident":
		c.handleResolveIncident(ctx, request)
	case "get_incident":
		c.handleGetIncident(ctx, request)
	case "list_incidents":
		c.handleListIncidents(ctx, request)
	case "get_incident_stats":
		c.handleGetIncidentStats(ctx, request)
	default:
		c.logger.Warn("Unknown incident request type", logger.String("type", request.Type))
		response := CreateIncidentErrorResponse("error_response", request.RequestID, fmt.Sprintf("unknown request type: %s", request.Type))
		c.sendResponse(response)
	}
}

// handleCreateIncident handles incident creation request
// Обрабатывает запрос создания инцидента
func (c *Component) handleCreateIncident(ctx context.Context, request IncidentRequest) {
	var payload CreateIncidentPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateIncidentErrorResponse("create_incident_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		c.sendResponse(response)
		return
	}

	// Convert to create request
	createRequest := &CreateIncidentRequest{
		Type:              IncidentType(payload.Type),
		Message:           payload.Message,
		ErrorCode:         payload.ErrorCode,
		ProcessInstanceID: payload.ProcessInstanceID,
		ProcessKey:        payload.ProcessKey,
		ElementID:         payload.ElementID,
		ElementType:       payload.ElementType,
		JobKey:            payload.JobKey,
		JobType:           payload.JobType,
		WorkerID:          payload.WorkerID,
		TimerID:           payload.TimerID,
		MessageName:       payload.MessageName,
		CorrelationKey:    payload.CorrelationKey,
		OriginalRetries:   payload.OriginalRetries,
		Metadata:          payload.Metadata,
	}

	incident, err := c.manager.CreateIncident(ctx, createRequest)
	if err != nil {
		response := CreateIncidentErrorResponse("create_incident_response", request.RequestID, err.Error())
		c.sendResponse(response)
		return
	}

	response := CreateIncidentSuccessResponse("create_incident_response", incident)
	c.sendResponse(response)
}

// handleResolveIncident handles incident resolution request
// Обрабатывает запрос разрешения инцидента
func (c *Component) handleResolveIncident(ctx context.Context, request IncidentRequest) {
	var payload ResolveIncidentPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateIncidentErrorResponse("resolve_incident_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		c.sendResponse(response)
		return
	}

	// Convert to resolve request
	resolveRequest := &ResolveIncidentRequest{
		IncidentID: payload.IncidentID,
		Action:     ResolveAction(payload.Action),
		Comment:    payload.Comment,
		ResolvedBy: payload.ResolvedBy,
		NewRetries: payload.NewRetries,
	}

	incident, err := c.manager.ResolveIncident(ctx, resolveRequest)
	if err != nil {
		response := CreateIncidentErrorResponse("resolve_incident_response", request.RequestID, err.Error())
		c.sendResponse(response)
		return
	}

	response := CreateIncidentSuccessResponse("resolve_incident_response", incident)
	c.sendResponse(response)
}

// handleGetIncident handles get incident request
// Обрабатывает запрос получения инцидента
func (c *Component) handleGetIncident(ctx context.Context, request IncidentRequest) {
	var payload GetIncidentPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateIncidentErrorResponse("get_incident_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		c.sendResponse(response)
		return
	}

	incident, err := c.manager.GetIncident(ctx, payload.IncidentID)
	if err != nil {
		response := CreateIncidentErrorResponse("get_incident_response", request.RequestID, err.Error())
		c.sendResponse(response)
		return
	}

	response := CreateIncidentSuccessResponse("get_incident_response", incident)
	c.sendResponse(response)
}

// handleListIncidents handles list incidents request
// Обрабатывает запрос получения списка инцидентов
func (c *Component) handleListIncidents(ctx context.Context, request IncidentRequest) {
	var payload ListIncidentsPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateIncidentErrorResponse("list_incidents_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		c.sendResponse(response)
		return
	}

	// Convert payload to filter
	filter := &IncidentFilter{
		ProcessInstanceID: payload.ProcessInstanceID,
		ProcessKey:        payload.ProcessKey,
		ElementID:         payload.ElementID,
		JobKey:            payload.JobKey,
		WorkerID:          payload.WorkerID,
		Limit:             payload.Limit,
		Offset:            payload.Offset,
	}

	// Convert string arrays to typed arrays
	for _, status := range payload.Status {
		filter.Status = append(filter.Status, IncidentStatus(status))
	}
	for _, incidentType := range payload.Type {
		filter.Type = append(filter.Type, IncidentType(incidentType))
	}

	incidents, total, err := c.manager.ListIncidents(ctx, filter)
	if err != nil {
		response := CreateIncidentErrorResponse("list_incidents_response", request.RequestID, err.Error())
		c.sendResponse(response)
		return
	}

	response := CreateIncidentListResponse(incidents, total)
	c.sendResponse(response)
}

// handleGetIncidentStats handles get incident stats request
// Обрабатывает запрос получения статистики инцидентов
func (c *Component) handleGetIncidentStats(ctx context.Context, request IncidentRequest) {
	stats, err := c.manager.GetIncidentStats(ctx)
	if err != nil {
		response := CreateIncidentErrorResponse("get_incident_stats_response", request.RequestID, err.Error())
		c.sendResponse(response)
		return
	}

	response := CreateIncidentStatsResponse(stats)
	c.sendResponse(response)
}

// sendResponse sends response to response channel
// Отправляет ответ в канал ответов
func (c *Component) sendResponse(response string) {
	select {
	case c.responseChannel <- response:
		// Response sent successfully
	default:
		c.logger.Warn("Incidents response channel full, response dropped")
	}
}

// ProcessMessage processes JSON message
// Обрабатывает JSON сообщение
func (c *Component) ProcessMessage(ctx context.Context, messageJSON string) error {
	if !c.IsReady() {
		return fmt.Errorf("incidents component is not ready")
	}

	// Send message to processing channel
	select {
	case c.requestChannel <- messageJSON:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("incidents component request channel is full")
	}
}

// GetResponseChannel returns response channel for JSON communication
// Возвращает канал ответов для JSON коммуникации
func (c *Component) GetResponseChannel() <-chan string {
	return c.responseChannel
}
