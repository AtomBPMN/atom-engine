/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import "atom-engine/src/storage"

// StorageInterface defines minimal storage interface needed by auth components
type StorageInterface interface {
	SaveRateLimitInfo(identifier string, info *storage.RateLimitInfo) error
	LoadRateLimitInfo(identifier string) (*storage.RateLimitInfo, error)
	LoadAllRateLimitInfo() (map[string]*storage.RateLimitInfo, error)
	DeleteRateLimitInfo(identifier string) error
	CleanupExpiredRateLimitInfo() error
}

// AuthManager defines the main authentication manager interface
type AuthManager interface {
	// Authenticate validates the authentication context
	Authenticate(ctx AuthContext) (*AuthResult, error)

	// CheckPermission validates if authenticated context has required permission
	CheckPermission(result *AuthResult, permission string) bool

	// IsEnabled returns true if authentication is enabled
	IsEnabled() bool

	// GetConfig returns current auth configuration
	GetConfig() *AuthConfig
}

// APIKeyValidator defines interface for API key validation
type APIKeyValidator interface {
	// ValidateAPIKey validates an API key and returns associated configuration
	ValidateAPIKey(key string) (*APIKey, bool)

	// GetAPIKeys returns all configured API keys
	GetAPIKeys() []APIKey
}

// IPValidator defines interface for IP whitelist validation
type IPValidator interface {
	// ValidateIP checks if IP is allowed based on configuration
	ValidateIP(ip string, allowedHosts []string) bool

	// IsAllowedGlobally checks if IP is in global whitelist
	IsAllowedGlobally(ip string) bool
}

// RateLimiter defines interface for request rate limiting
type RateLimiter interface {
	// CheckLimit verifies if request is within rate limits
	CheckLimit(clientIP string, apiKey string) bool

	// RecordRequest records a request for rate limiting
	RecordRequest(clientIP string, apiKey string)

	// GetStats returns current rate limiting statistics
	GetStats() map[string]interface{}

	// SetStorage sets storage for persistent rate limiting
	SetStorage(storage StorageInterface)

	// LoadState loads rate limiter state from storage
	LoadState() error
}

// AuditLogger defines interface for security event logging
type AuditLogger interface {
	// LogEvent logs a security audit event
	LogEvent(event AuditEvent)

	// LogAuthSuccess logs successful authentication
	LogAuthSuccess(ctx AuthContext, result *AuthResult)

	// LogAuthFailure logs failed authentication attempt
	LogAuthFailure(ctx AuthContext, reason string)

	// LogIPBlocked logs blocked IP attempt
	LogIPBlocked(ctx AuthContext, reason string)

	// GetRecentEvents returns recent audit events
	GetRecentEvents(limit int) []AuditEvent
}

// Component defines the main auth component interface
type Component interface {
	AuthManager

	// Initialize initializes the auth component with configuration
	Initialize(config *AuthConfig) error

	// Start starts the auth component
	Start() error

	// Stop stops the auth component
	Stop() error

	// IsReady returns true if component is ready
	IsReady() bool

	// GetAPIKeyValidator returns API key validator
	GetAPIKeyValidator() APIKeyValidator

	// GetIPValidator returns IP validator
	GetIPValidator() IPValidator

	// GetRateLimiter returns rate limiter
	GetRateLimiter() RateLimiter

	// GetAuditLogger returns audit logger
	GetAuditLogger() AuditLogger

	// SetStorage sets storage for persistent auth operations
	SetStorage(storage StorageInterface) error
}
