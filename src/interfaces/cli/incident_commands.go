/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"atom-engine/proto/incidents/incidentspb"
	"atom-engine/src/core/logger"
)

// IncidentList lists incidents with optional filtering
// Получает список инцидентов с дополнительной фильтрацией
func (d *DaemonCommand) IncidentList() error {
	logger.Debug("Listing incidents")

	// Parse arguments: atomd incident list [status] [type] [limit]
	args := os.Args[3:] // Skip "atomd incident list"

	// Default filter values
	var statusFilter []incidentspb.IncidentStatus
	var typeFilter []incidentspb.IncidentType
	var limit int32 = 10

	// Parse arguments
	for i, arg := range args {
		switch i {
		case 0: // status filter
			if arg != "" && arg != "all" {
				status := parseIncidentStatus(arg)
				if status != incidentspb.IncidentStatus_INCIDENT_STATUS_UNSPECIFIED {
					statusFilter = append(statusFilter, status)
				}
			}
		case 1: // type filter
			if arg != "" && arg != "all" {
				incidentType := parseIncidentType(arg)
				if incidentType != incidentspb.IncidentType_INCIDENT_TYPE_UNSPECIFIED {
					typeFilter = append(typeFilter, incidentType)
				}
			}
		case 2: // limit
			if limitVal, err := strconv.Atoi(arg); err == nil && limitVal > 0 {
				limit = int32(limitVal)
			}
		}
	}

	logger.Debug("Incident list request",
		logger.Int("status_filters", len(statusFilter)),
		logger.Int("type_filters", len(typeFilter)),
		logger.Int("limit", int(limit)))

	// Connect to gRPC server
	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for incident list", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := incidentspb.NewIncidentsServiceClient(conn)

	// Create request
	req := &incidentspb.ListIncidentsRequest{
		Filter: &incidentspb.IncidentFilter{
			Status: statusFilter,
			Type:   typeFilter,
			Limit:  limit,
		},
	}

	// Execute request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListIncidents(ctx, req)
	if err != nil {
		logger.Error("Failed to list incidents", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list incidents: %w", err)
	}

	// Display results
	if len(resp.Incidents) == 0 {
		fmt.Println("No incidents found")
		return nil
	}

	fmt.Printf("Found %d incidents (total: %d)\n\n", len(resp.Incidents), resp.Total)

	// Display table header
	fmt.Printf("%-25s %-15s %-15s %-40s %-25s\n", "ID", "TYPE", "STATUS", "MESSAGE", "CREATED")
	fmt.Println(strings.Repeat("-", 120))

	// Display incidents
	for _, incident := range resp.Incidents {
		message := incident.Message
		if len(message) > 35 {
			message = message[:32] + "..."
		}

		createdAt := "N/A"
		if incident.CreatedAt != nil {
			createdAt = incident.CreatedAt.AsTime().Format("2006-01-02 15:04")
		}

		fmt.Printf("%-25s %-15s %-15s %-40s %-25s\n",
			truncateString(incident.Id, 24),
			formatIncidentType(incident.Type),
			formatIncidentStatus(incident.Status),
			message,
			createdAt)
	}

	fmt.Printf("\nShowing %d of %d incidents\n", len(resp.Incidents), resp.Total)
	return nil
}

// IncidentShow displays detailed information about an incident
// Показывает детальную информацию об инциденте
func (d *DaemonCommand) IncidentShow() error {
	logger.Debug("Showing incident details")

	if len(os.Args) < 4 {
		return fmt.Errorf("usage: atomd incident show <incident_id>")
	}

	incidentID := os.Args[3]

	logger.Debug("Incident show request", logger.String("incident_id", incidentID))

	// Connect to gRPC server
	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for incident show", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := incidentspb.NewIncidentsServiceClient(conn)

	// Execute request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetIncident(ctx, &incidentspb.GetIncidentRequest{
		IncidentId: incidentID,
	})
	if err != nil {
		logger.Error("Failed to get incident", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get incident: %w", err)
	}

	if resp.Incident == nil {
		fmt.Printf("Incident not found: %s\n", incidentID)
		return nil
	}

	// Display detailed information
	incident := resp.Incident
	fmt.Printf("Incident Details\n")
	fmt.Printf("================\n\n")

	fmt.Printf("ID:           %s\n", incident.Id)
	fmt.Printf("Type:         %s\n", formatIncidentType(incident.Type))
	fmt.Printf("Status:       %s\n", formatIncidentStatus(incident.Status))
	fmt.Printf("Message:      %s\n", incident.Message)

	if incident.ErrorCode != "" {
		fmt.Printf("Error Code:   %s\n", incident.ErrorCode)
	}

	if incident.ProcessInstanceId != "" {
		fmt.Printf("Process:      %s\n", incident.ProcessInstanceId)
	}

	if incident.ProcessKey != "" {
		fmt.Printf("Process Key:  %s\n", incident.ProcessKey)
	}

	if incident.ElementId != "" {
		fmt.Printf("Element:      %s\n", incident.ElementId)
	}

	if incident.JobKey != "" {
		fmt.Printf("Job Key:      %s\n", incident.JobKey)
	}

	if incident.JobType != "" {
		fmt.Printf("Job Type:     %s\n", incident.JobType)
	}

	if incident.WorkerId != "" {
		fmt.Printf("Worker:       %s\n", incident.WorkerId)
	}

	if incident.TimerId != "" {
		fmt.Printf("Timer ID:     %s\n", incident.TimerId)
	}

	if incident.MessageName != "" {
		fmt.Printf("Message:      %s\n", incident.MessageName)
	}

	if incident.CorrelationKey != "" {
		fmt.Printf("Correlation:  %s\n", incident.CorrelationKey)
	}

	// Timestamps
	if incident.CreatedAt != nil {
		fmt.Printf("Created:      %s\n", incident.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}

	if incident.UpdatedAt != nil {
		fmt.Printf("Updated:      %s\n", incident.UpdatedAt.AsTime().Format("2006-01-02 15:04:05"))
	}

	// Resolution info
	if incident.Status != incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN {
		if incident.ResolvedAt != nil {
			fmt.Printf("Resolved:     %s\n", incident.ResolvedAt.AsTime().Format("2006-01-02 15:04:05"))
		}
		if incident.ResolvedBy != "" {
			fmt.Printf("Resolved By:  %s\n", incident.ResolvedBy)
		}
		if incident.ResolveAction != incidentspb.ResolveAction_RESOLVE_ACTION_UNSPECIFIED {
			fmt.Printf("Action:       %s\n", formatResolveAction(incident.ResolveAction))
		}
		if incident.NewRetries > 0 {
			fmt.Printf("New Retries:  %d\n", incident.NewRetries)
		}
	}

	// Metadata
	if len(incident.Metadata) > 0 {
		fmt.Printf("\nMetadata:\n")
		for key, value := range incident.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

// IncidentResolve resolves an incident with retry or dismiss action
// Разрешает инцидент с действием retry или dismiss
func (d *DaemonCommand) IncidentResolve() error {
	logger.Debug("Resolving incident")

	if len(os.Args) < 5 {
		return fmt.Errorf("usage: atomd incident resolve <incident_id> <retry|dismiss> [retries] [comment]")
	}

	incidentID := os.Args[3]
	actionStr := strings.ToLower(os.Args[4])

	var action incidentspb.ResolveAction
	var newRetries int32 = 0
	var comment string = ""

	// Parse action
	switch actionStr {
	case "retry":
		action = incidentspb.ResolveAction_RESOLVE_ACTION_RETRY
		// Parse retries if provided
		if len(os.Args) > 5 {
			if retries, err := strconv.Atoi(os.Args[5]); err == nil && retries >= 0 {
				newRetries = int32(retries)
			} else {
				return fmt.Errorf("invalid retries value: %s", os.Args[5])
			}
		} else {
			newRetries = 3 // Default retries
		}
		// Parse comment if provided
		if len(os.Args) > 6 {
			comment = strings.Join(os.Args[6:], " ")
		}
	case "dismiss":
		action = incidentspb.ResolveAction_RESOLVE_ACTION_DISMISS
		// Parse comment if provided
		if len(os.Args) > 5 {
			comment = strings.Join(os.Args[5:], " ")
		}
	default:
		return fmt.Errorf("invalid action: %s. Use 'retry' or 'dismiss'", actionStr)
	}

	logger.Debug("Incident resolve request",
		logger.String("incident_id", incidentID),
		logger.String("action", actionStr),
		logger.Int("new_retries", int(newRetries)),
		logger.String("comment", comment))

	// Connect to gRPC server
	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for incident resolve", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := incidentspb.NewIncidentsServiceClient(conn)

	// Execute request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ResolveIncident(ctx, &incidentspb.ResolveIncidentRequest{
		IncidentId: incidentID,
		Action:     action,
		Comment:    comment,
		ResolvedBy: "cli-user", // TODO: Get actual user from config
		NewRetries: newRetries,
	})
	if err != nil {
		logger.Error("Failed to resolve incident", logger.String("error", err.Error()))
		return fmt.Errorf("failed to resolve incident: %w", err)
	}

	if resp.Incident == nil {
		fmt.Printf("Failed to resolve incident: %s\n", incidentID)
		return nil
	}

	// Display success message
	fmt.Printf("Incident %s resolved successfully\n", incidentID)
	fmt.Printf("Action: %s\n", formatResolveAction(action))
	if newRetries > 0 {
		fmt.Printf("New retries: %d\n", newRetries)
	}
	if comment != "" {
		fmt.Printf("Comment: %s\n", comment)
	}

	return nil
}

// IncidentStats displays incident statistics
// Показывает статистику инцидентов
func (d *DaemonCommand) IncidentStats() error {
	logger.Debug("Getting incident statistics")

	// Connect to gRPC server
	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for incident stats", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := incidentspb.NewIncidentsServiceClient(conn)

	// Execute request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetIncidentStats(ctx, &incidentspb.GetIncidentStatsRequest{})
	if err != nil {
		logger.Error("Failed to get incident stats", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get incident stats: %w", err)
	}

	if resp.Stats == nil {
		fmt.Println("No incident statistics available")
		return nil
	}

	// Display statistics
	stats := resp.Stats
	fmt.Printf("Incident Statistics\n")
	fmt.Printf("==================\n\n")

	fmt.Printf("Total Incidents:      %d\n", stats.TotalIncidents)
	fmt.Printf("Open Incidents:       %d\n", stats.OpenIncidents)
	fmt.Printf("Resolved Incidents:   %d\n", stats.ResolvedIncidents)
	fmt.Printf("Dismissed Incidents:  %d\n", stats.DismissedIncidents)
	fmt.Printf("Recent (24h):         %d\n", stats.RecentIncidents_24H)

	// By Type
	if len(stats.IncidentsByType) > 0 {
		fmt.Printf("\nBy Type:\n")
		for typeStr, count := range stats.IncidentsByType {
			fmt.Printf("  %-20s %d\n", typeStr, count)
		}
	}

	// By Status
	if len(stats.IncidentsByStatus) > 0 {
		fmt.Printf("\nBy Status:\n")
		for statusStr, count := range stats.IncidentsByStatus {
			fmt.Printf("  %-20s %d\n", statusStr, count)
		}
	}

	return nil
}

// Helper functions for parsing and formatting

// parseIncidentStatus converts string to incident status enum
func parseIncidentStatus(status string) incidentspb.IncidentStatus {
	switch strings.ToUpper(status) {
	case "OPEN":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN
	case "RESOLVED":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_RESOLVED
	case "DISMISSED":
		return incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED
	default:
		return incidentspb.IncidentStatus_INCIDENT_STATUS_UNSPECIFIED
	}
}

// parseIncidentType converts string to incident type enum
func parseIncidentType(incidentType string) incidentspb.IncidentType {
	switch strings.ToUpper(incidentType) {
	case "JOB_FAILURE", "JOB":
		return incidentspb.IncidentType_INCIDENT_TYPE_JOB_FAILURE
	case "BPMN_ERROR", "BPMN":
		return incidentspb.IncidentType_INCIDENT_TYPE_BPMN_ERROR
	case "EXPRESSION_ERROR", "EXPRESSION":
		return incidentspb.IncidentType_INCIDENT_TYPE_EXPRESSION_ERROR
	case "PROCESS_ERROR", "PROCESS":
		return incidentspb.IncidentType_INCIDENT_TYPE_PROCESS_ERROR
	case "TIMER_ERROR", "TIMER":
		return incidentspb.IncidentType_INCIDENT_TYPE_TIMER_ERROR
	case "MESSAGE_ERROR", "MESSAGE":
		return incidentspb.IncidentType_INCIDENT_TYPE_MESSAGE_ERROR
	case "SYSTEM_ERROR", "SYSTEM":
		return incidentspb.IncidentType_INCIDENT_TYPE_SYSTEM_ERROR
	default:
		return incidentspb.IncidentType_INCIDENT_TYPE_UNSPECIFIED
	}
}

// formatIncidentType formats incident type enum for display
func formatIncidentType(incidentType incidentspb.IncidentType) string {
	switch incidentType {
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
		return "UNKNOWN"
	}
}

// formatIncidentStatus formats incident status enum for display
func formatIncidentStatus(status incidentspb.IncidentStatus) string {
	switch status {
	case incidentspb.IncidentStatus_INCIDENT_STATUS_OPEN:
		return "OPEN"
	case incidentspb.IncidentStatus_INCIDENT_STATUS_RESOLVED:
		return "RESOLVED"
	case incidentspb.IncidentStatus_INCIDENT_STATUS_DISMISSED:
		return "DISMISSED"
	default:
		return "UNKNOWN"
	}
}

// formatResolveAction formats resolve action enum for display
func formatResolveAction(action incidentspb.ResolveAction) string {
	switch action {
	case incidentspb.ResolveAction_RESOLVE_ACTION_RETRY:
		return "RETRY"
	case incidentspb.ResolveAction_RESOLVE_ACTION_DISMISS:
		return "DISMISS"
	default:
		return "UNKNOWN"
	}
}

// truncateString truncates string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
