/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"time"
)

// ProcessMessageSubscription represents process message subscription
type ProcessMessageSubscription struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	ProcessDefinitionKey string    `json:"process_definition_key"`
	ProcessVersion       int32     `json:"process_version"`
	StartEventID         string    `json:"start_event_id"`
	MessageName          string    `json:"message_name"`
	MessageRef           string    `json:"message_ref"`
	CorrelationKey       string    `json:"correlation_key,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// BufferedMessage represents a buffered message
type BufferedMessage struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	Name           string                 `json:"name"`
	CorrelationKey string                 `json:"correlation_key,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	PublishedAt    time.Time              `json:"published_at"`
	BufferedAt     time.Time              `json:"buffered_at"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	Reason         string                 `json:"reason"`
}

// MessageCorrelationResult represents message correlation result
type MessageCorrelationResult struct {
	ID                string                 `json:"id"`
	MessageID         string                 `json:"message_id"`
	TenantID          string                 `json:"tenant_id"`
	MessageName       string                 `json:"message_name"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	InstanceCreated   bool                   `json:"instance_created"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
}

// IsExpired checks if buffered message is expired
func (bm *BufferedMessage) IsExpired() bool {
	if bm.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*bm.ExpiresAt)
}

// NewProcessMessageSubscription creates new process message subscription
func NewProcessMessageSubscription(tenantID, processKey, startEventID, messageName string) *ProcessMessageSubscription {
	now := time.Now()
	return &ProcessMessageSubscription{
		ID:                   GenerateID(),
		TenantID:             tenantID,
		ProcessDefinitionKey: processKey,
		StartEventID:         startEventID,
		MessageName:          messageName,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// NewBufferedMessage creates new buffered message
func NewBufferedMessage(tenantID, name, correlationKey string, variables map[string]interface{}, reason string) *BufferedMessage {
	now := time.Now()
	return &BufferedMessage{
		ID:             GenerateID(),
		TenantID:       tenantID,
		Name:           name,
		CorrelationKey: correlationKey,
		Variables:      variables,
		PublishedAt:    now,
		BufferedAt:     now,
		Reason:         reason,
	}
}

// NewMessageCorrelationResult creates new message correlation result
func NewMessageCorrelationResult(messageID, tenantID, messageName, correlationKey string) *MessageCorrelationResult {
	return &MessageCorrelationResult{
		ID:             GenerateID(),
		MessageID:      messageID,
		TenantID:       tenantID,
		MessageName:    messageName,
		CorrelationKey: correlationKey,
		CreatedAt:      time.Now(),
	}
}
