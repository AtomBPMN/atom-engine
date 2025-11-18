/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"encoding/json"
	"sync"

	"atom-engine/src/core/logger"
)

// auditLogger implements AuditLogger interface
type auditLogger struct {
	config       AuditConfig
	recentEvents []AuditEvent
	mutex        sync.RWMutex
	maxEvents    int // Maximum number of recent events to keep in memory
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditConfig) AuditLogger {
	return &auditLogger{
		config:       config,
		recentEvents: make([]AuditEvent, 0),
		maxEvents:    1000, // Keep last 1000 events in memory
	}
}

// LogEvent logs a security audit event
func (al *auditLogger) LogEvent(event AuditEvent) {
	if !al.config.Enabled {
		return
	}

	// Add to recent events (in memory)
	al.addToRecentEvents(event)

	// Log to structured logger
	eventJSON, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal audit event",
			logger.String("error", err.Error()))
		return
	}

	// Log with appropriate level based on result
	switch event.Result {
	case "success":
		if al.config.LogSuccessfulAuth {
			logger.Info("Auth success",
				logger.String("audit_event", string(eventJSON)))
		}
	case "failed", "blocked":
		if al.config.LogFailedAttempts {
			logger.Warn("Auth failure",
				logger.String("audit_event", string(eventJSON)))
		}
	default:
		logger.Info("Auth event",
			logger.String("audit_event", string(eventJSON)))
	}
}

// LogAuthSuccess logs successful authentication
func (al *auditLogger) LogAuthSuccess(ctx AuthContext, result *AuthResult) {
	event := AuditEvent{
		Timestamp:   ctx.Timestamp,
		ClientIP:    ctx.ClientIP,
		APIKey:      maskAPIKey(ctx.APIKey),
		APIKeyName:  result.APIKeyName,
		Protocol:    ctx.Protocol,
		Method:      ctx.Method,
		RequestPath: ctx.RequestPath,
		UserAgent:   ctx.UserAgent,
		Result:      "success",
		Reason:      "Authentication successful",
	}

	al.LogEvent(event)
}

// LogAuthFailure logs failed authentication attempt
func (al *auditLogger) LogAuthFailure(ctx AuthContext, reason string) {
	event := AuditEvent{
		Timestamp:   ctx.Timestamp,
		ClientIP:    ctx.ClientIP,
		APIKey:      maskAPIKey(ctx.APIKey),
		Protocol:    ctx.Protocol,
		Method:      ctx.Method,
		RequestPath: ctx.RequestPath,
		UserAgent:   ctx.UserAgent,
		Result:      "failed",
		Reason:      reason,
	}

	al.LogEvent(event)
}

// LogIPBlocked logs blocked IP attempt
func (al *auditLogger) LogIPBlocked(ctx AuthContext, reason string) {
	event := AuditEvent{
		Timestamp:   ctx.Timestamp,
		ClientIP:    ctx.ClientIP,
		APIKey:      maskAPIKey(ctx.APIKey),
		Protocol:    ctx.Protocol,
		Method:      ctx.Method,
		RequestPath: ctx.RequestPath,
		UserAgent:   ctx.UserAgent,
		Result:      "blocked",
		Reason:      reason,
	}

	al.LogEvent(event)
}

// GetRecentEvents returns recent audit events
func (al *auditLogger) GetRecentEvents(limit int) []AuditEvent {
	al.mutex.RLock()
	defer al.mutex.RUnlock()

	if limit <= 0 || limit > len(al.recentEvents) {
		limit = len(al.recentEvents)
	}

	// Return the most recent events (from the end of the slice)
	start := len(al.recentEvents) - limit
	if start < 0 {
		start = 0
	}

	events := make([]AuditEvent, limit)
	copy(events, al.recentEvents[start:])

	// Reverse to get newest first
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events
}

// addToRecentEvents adds event to in-memory recent events list
func (al *auditLogger) addToRecentEvents(event AuditEvent) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	// Add new event
	al.recentEvents = append(al.recentEvents, event)

	// Trim if exceeding max events
	if len(al.recentEvents) > al.maxEvents {
		// Remove oldest events, keep the most recent ones
		al.recentEvents = al.recentEvents[len(al.recentEvents)-al.maxEvents:]
	}
}

// GetStats returns audit logger statistics
func (al *auditLogger) GetStats() map[string]interface{} {
	al.mutex.RLock()
	defer al.mutex.RUnlock()

	stats := map[string]interface{}{
		"enabled":             al.config.Enabled,
		"log_failed_attempts": al.config.LogFailedAttempts,
		"log_successful_auth": al.config.LogSuccessfulAuth,
		"recent_events_count": len(al.recentEvents),
		"max_events":          al.maxEvents,
	}

	// Count events by result type
	if len(al.recentEvents) > 0 {
		resultCounts := make(map[string]int)
		for _, event := range al.recentEvents {
			resultCounts[event.Result]++
		}
		stats["result_counts"] = resultCounts
	}

	return stats
}

// UpdateConfig updates audit logger configuration
func (al *auditLogger) UpdateConfig(config AuditConfig) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	al.config = config

	logger.Info("Audit logger configuration updated",
		logger.Bool("enabled", config.Enabled),
		logger.Bool("log_failed_attempts", config.LogFailedAttempts),
		logger.Bool("log_successful_auth", config.LogSuccessfulAuth))
}
