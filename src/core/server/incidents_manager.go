/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"encoding/json"
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// processIncidentsResponses processes incidents responses in background
// Обрабатывает ответы incidents в фоне
func (c *Core) processIncidentsResponses() {
	if c.incidentsComp == nil {
		logger.Warn("Incidents component is nil in processIncidentsResponses")
		return
	}

	responseChannel := c.incidentsComp.GetResponseChannel()
	if responseChannel == nil {
		logger.Warn("Response channel is nil from incidents component")
		return
	}

	logger.Info("Incidents response processor started")

	for {
		select {
		case response := <-responseChannel:
			logger.Debug("Received incident response", logger.String("response", response))
			c.handleIncidentsResponse(response)
		}
	}
}

// handleIncidentsResponse handles single incidents response
// Обрабатывает один ответ incidents
func (c *Core) handleIncidentsResponse(response string) {
	logger.Debug("Processing incident response", logger.String("response", response))

	// Try to parse as incident response for logging
	var incidentResponse struct {
		Type       string      `json:"type"`
		RequestID  string      `json:"request_id"`
		Success    bool        `json:"success"`
		Error      string      `json:"error,omitempty"`
		Data       interface{} `json:"data,omitempty"`
		IncidentID string      `json:"incident_id,omitempty"`
	}

	if err := json.Unmarshal([]byte(response), &incidentResponse); err != nil {
		logger.Warn("Failed to parse incident response JSON",
			logger.String("response", response),
			logger.String("error", err.Error()))
		return
	}

	// Log based on response type
	switch incidentResponse.Type {
	case "create_incident_response":
		if incidentResponse.Success {
			logger.Info("Incident created successfully",
				logger.String("incident_id", incidentResponse.IncidentID),
				logger.String("request_id", incidentResponse.RequestID))
		} else {
			logger.Error("Failed to create incident",
				logger.String("error", incidentResponse.Error),
				logger.String("request_id", incidentResponse.RequestID))
		}

	case "list_incidents_response":
		if incidentResponse.Success {
			logger.Debug("Incidents listed successfully",
				logger.String("request_id", incidentResponse.RequestID))
		} else {
			logger.Error("Failed to list incidents",
				logger.String("error", incidentResponse.Error),
				logger.String("request_id", incidentResponse.RequestID))
		}

	case "get_incident_response":
		if incidentResponse.Success {
			logger.Debug("Incident retrieved successfully",
				logger.String("request_id", incidentResponse.RequestID))
		} else {
			logger.Error("Failed to get incident",
				logger.String("error", incidentResponse.Error),
				logger.String("request_id", incidentResponse.RequestID))
		}

	case "resolve_incident_response":
		if incidentResponse.Success {
			logger.Info("Incident resolved successfully",
				logger.String("incident_id", incidentResponse.IncidentID),
				logger.String("request_id", incidentResponse.RequestID))
		} else {
			logger.Error("Failed to resolve incident",
				logger.String("error", incidentResponse.Error),
				logger.String("request_id", incidentResponse.RequestID))
		}

	case "get_incident_stats_response":
		if incidentResponse.Success {
			logger.Debug("Incident stats retrieved successfully",
				logger.String("request_id", incidentResponse.RequestID))
		} else {
			logger.Error("Failed to get incident stats",
				logger.String("error", incidentResponse.Error),
				logger.String("request_id", incidentResponse.RequestID))
		}

	default:
		logger.Debug("Unknown incident response type",
			logger.String("type", incidentResponse.Type),
			logger.String("request_id", incidentResponse.RequestID))
	}

	// Log system event for successful incident operations
	if incidentResponse.Success && incidentResponse.Type == "create_incident_response" {
		err := c.storage.LogSystemEvent(models.EventTypeReady, models.StatusSuccess,
			fmt.Sprintf("Incident created: %s", response))
		if err != nil {
			logger.Warn("Failed to log incident creation system event",
				logger.String("error", err.Error()))
		}
	}
}
