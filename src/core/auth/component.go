/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"errors"
	"fmt"

	"atom-engine/src/core/logger"
)

// component implements Component interface
type component struct {
	config        *AuthConfig
	apiKeyManager APIKeyValidator
	ipValidator   IPValidator
	rateLimiter   RateLimiter
	auditLogger   AuditLogger
	initialized   bool
	running       bool
}

// NewComponent creates a new auth component
func NewComponent() Component {
	return &component{
		initialized: false,
		running:     false,
	}
}

// Initialize initializes the auth component with configuration
func (c *component) Initialize(config *AuthConfig) error {
	if config == nil {
		return errors.New("auth config cannot be nil")
	}

	c.config = config

	// Initialize sub-components
	c.apiKeyManager = NewAPIKeyManager(config.APIKeys)
	c.ipValidator = NewIPValidator(config.AllowedHosts)
	c.rateLimiter = NewRateLimiter(config.RateLimit.Enabled, config.RateLimit.RequestsPerMinute)
	c.auditLogger = NewAuditLogger(config.Audit)

	c.initialized = true

	logger.Info("Auth component initialized",
		logger.Bool("enabled", config.Enabled),
		logger.Int("api_keys_count", len(config.APIKeys)),
		logger.Int("allowed_hosts_count", len(config.AllowedHosts)))

	return nil
}

// Start starts the auth component
func (c *component) Start() error {
	if !c.initialized {
		return errors.New("auth component not initialized")
	}

	c.running = true
	logger.Info("Auth component started")
	return nil
}

// Stop stops the auth component
func (c *component) Stop() error {
	if !c.running {
		return nil
	}

	// Stop rate limiter if it's running
	if rl, ok := c.rateLimiter.(*rateLimiter); ok {
		rl.Stop()
	}

	c.running = false
	logger.Info("Auth component stopped")
	return nil
}

// IsReady returns true if component is ready
func (c *component) IsReady() bool {
	return c.initialized && c.running
}

// IsEnabled returns true if authentication is enabled
func (c *component) IsEnabled() bool {
	if c.config == nil {
		return false
	}
	return c.config.Enabled
}

// GetConfig returns current auth configuration
func (c *component) GetConfig() *AuthConfig {
	return c.config
}

// Authenticate validates the authentication context
func (c *component) Authenticate(ctx AuthContext) (*AuthResult, error) {
	if !c.IsEnabled() {
		// Authentication disabled - allow all requests
		return &AuthResult{
			Authenticated: true,
			APIKeyName:    "auth_disabled",
			Permissions:   []string{PermissionAll},
			Reason:        "Authentication disabled",
		}, nil
	}

	if !c.IsReady() {
		return nil, errors.New("auth component not ready")
	}

	// Check if localhost - allow without API key for CLI compatibility
	if IsLocalhost(ctx.ClientIP) {
		// Localhost bypass - allow CLI commands without API key
		return &AuthResult{
			Authenticated: true,
			APIKeyName:    "localhost_bypass",
			Permissions:   []string{PermissionAll},
			Reason:        "Localhost bypass",
		}, nil
	}

	// Check rate limit first
	if !c.rateLimiter.CheckLimit(ctx.ClientIP, ctx.APIKey) {
		c.auditLogger.LogAuthFailure(ctx, "Rate limit exceeded")
		return &AuthResult{
			Authenticated: false,
			Reason:        "Rate limit exceeded",
		}, nil
	}

	// Record the request for rate limiting
	c.rateLimiter.RecordRequest(ctx.ClientIP, ctx.APIKey)

	// Validate API key
	apiKey, valid := c.apiKeyManager.ValidateAPIKey(ctx.APIKey)
	if !valid {
		c.auditLogger.LogAuthFailure(ctx, "Invalid API key")
		return &AuthResult{
			Authenticated: false,
			Reason:        "Invalid API key",
		}, nil
	}

	// Check IP whitelist
	if !c.ipValidator.ValidateIP(ctx.ClientIP, apiKey.AllowedHosts) {
		c.auditLogger.LogIPBlocked(ctx, fmt.Sprintf("IP %s not in whitelist", ctx.ClientIP))
		return &AuthResult{
			Authenticated: false,
			Reason:        fmt.Sprintf("IP %s not allowed", ctx.ClientIP),
		}, nil
	}

	// Authentication successful
	result := &AuthResult{
		Authenticated: true,
		APIKeyName:    apiKey.Name,
		Permissions:   apiKey.Permissions,
		Reason:        "Authentication successful",
	}

	c.auditLogger.LogAuthSuccess(ctx, result)
	return result, nil
}

// CheckPermission validates if authenticated context has required permission
func (c *component) CheckPermission(result *AuthResult, permission string) bool {
	if result == nil || !result.Authenticated {
		return false
	}

	return HasPermission(result.Permissions, permission)
}

// GetAPIKeyValidator returns API key validator
func (c *component) GetAPIKeyValidator() APIKeyValidator {
	return c.apiKeyManager
}

// GetIPValidator returns IP validator
func (c *component) GetIPValidator() IPValidator {
	return c.ipValidator
}

// GetRateLimiter returns rate limiter
func (c *component) GetRateLimiter() RateLimiter {
	return c.rateLimiter
}

// GetAuditLogger returns audit logger
func (c *component) GetAuditLogger() AuditLogger {
	return c.auditLogger
}

// SetStorage sets storage for persistent auth operations
// Устанавливает storage для персистентных auth операций
func (c *component) SetStorage(storage StorageInterface) error {
	if c.rateLimiter != nil {
		c.rateLimiter.SetStorage(storage)
		// Load existing state from storage
		if err := c.rateLimiter.LoadState(); err != nil {
			logger.Warn("Failed to load rate limiter state from storage", logger.String("error", err.Error()))
			return err
		}
		logger.Info("Storage set for auth component rate limiter")
	}
	return nil
}

// UpdateConfig updates the auth configuration
func (c *component) UpdateConfig(config *AuthConfig) error {
	if config == nil {
		return errors.New("auth config cannot be nil")
	}

	oldEnabled := c.IsEnabled()
	c.config = config

	// Update sub-components
	if akm, ok := c.apiKeyManager.(*apiKeyManager); ok {
		akm.UpdateAPIKeys(config.APIKeys)
	}

	if ipv, ok := c.ipValidator.(*ipValidator); ok {
		ipv.UpdateAllowedHosts(config.AllowedHosts)
	}

	if rl, ok := c.rateLimiter.(*rateLimiter); ok {
		rl.UpdateConfig(config.RateLimit.Enabled, config.RateLimit.RequestsPerMinute)
	}

	if al, ok := c.auditLogger.(*auditLogger); ok {
		al.UpdateConfig(config.Audit)
	}

	logger.Info("Auth component configuration updated",
		logger.Bool("was_enabled", oldEnabled),
		logger.Bool("now_enabled", config.Enabled),
		logger.Int("api_keys_count", len(config.APIKeys)))

	return nil
}

// GetStats returns comprehensive auth component statistics
func (c *component) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     c.IsEnabled(),
		"initialized": c.initialized,
		"running":     c.running,
	}

	if c.apiKeyManager != nil {
		if akm, ok := c.apiKeyManager.(*apiKeyManager); ok {
			stats["api_keys"] = akm.GetAPIKeyStats()
		}
	}

	if c.rateLimiter != nil {
		stats["rate_limiter"] = c.rateLimiter.GetStats()
	}

	if c.auditLogger != nil {
		if al, ok := c.auditLogger.(*auditLogger); ok {
			stats["audit"] = al.GetStats()
		}
	}

	if c.ipValidator != nil {
		if ipv, ok := c.ipValidator.(*ipValidator); ok {
			stats["allowed_hosts_count"] = len(ipv.GetAllowedHosts())
		}
	}

	return stats
}
