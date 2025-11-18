/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"time"

	"atom-engine/src/core/config"
)

// Import config types
type (
	AuthConfig      = config.AuthConfig
	APIKey          = config.APIKeyConfig
	RateLimitConfig = config.RateLimitConfig
	AuditConfig     = config.AuditConfig
)

// AuthContext represents protocol-agnostic authentication context
type AuthContext struct {
	ClientIP    string
	APIKey      string
	UserAgent   string
	RequestPath string
	Method      string
	Protocol    string // "grpc" or "http"
	Timestamp   time.Time
}

// AuthResult represents the result of authentication
type AuthResult struct {
	Authenticated bool
	APIKeyName    string
	Permissions   []string
	Reason        string // Reason for failure if not authenticated
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	ClientIP    string    `json:"client_ip"`
	APIKey      string    `json:"api_key,omitempty"`
	APIKeyName  string    `json:"api_key_name,omitempty"`
	Protocol    string    `json:"protocol"`
	Method      string    `json:"method"`
	RequestPath string    `json:"request_path"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Result      string    `json:"result"` // "success", "failed", "blocked"
	Reason      string    `json:"reason,omitempty"`
}

// Permission constants for common permissions
const (
	PermissionAll        = "*"
	PermissionRead       = "read"
	PermissionWrite      = "write"
	PermissionAdmin      = "admin"
	PermissionStorage    = "storage"
	PermissionProcess    = "process"
	PermissionTimer      = "timer"
	PermissionJob        = "job"
	PermissionMessage    = "message"
	PermissionIncident   = "incident"
	PermissionExpression = "expression"
	PermissionBPMN       = "bpmn"
)

// HasPermission checks if the given permissions include the required permission
func HasPermission(permissions []string, required string) bool {
	// Check for wildcard permission
	for _, perm := range permissions {
		if perm == PermissionAll {
			return true
		}
		if perm == required {
			return true
		}
		// Support for read-only checks
		if required == "read" && perm == PermissionRead {
			return true
		}
	}
	return false
}

// IsLocalhost checks if the given IP is localhost
func IsLocalhost(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}
