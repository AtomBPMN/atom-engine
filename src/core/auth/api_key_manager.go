/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"crypto/subtle"
	"strings"

	"atom-engine/src/core/logger"
)

// apiKeyManager implements APIKeyValidator interface
type apiKeyManager struct {
	apiKeys map[string]*APIKey // map[key]APIKey for fast lookup
}

// NewAPIKeyManager creates a new API key manager
func NewAPIKeyManager(apiKeys []APIKey) APIKeyValidator {
	keyMap := make(map[string]*APIKey)
	for i := range apiKeys {
		keyMap[apiKeys[i].Key] = &apiKeys[i]
	}

	return &apiKeyManager{
		apiKeys: keyMap,
	}
}

// ValidateAPIKey validates an API key and returns associated configuration
func (m *apiKeyManager) ValidateAPIKey(key string) (*APIKey, bool) {
	if key == "" {
		return nil, false
	}

	// Use constant-time comparison to prevent timing attacks
	for storedKey, apiKey := range m.apiKeys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(storedKey)) == 1 {
			logger.Debug("API key validated successfully",
				logger.String("key_name", apiKey.Name),
				logger.String("key_prefix", maskAPIKey(key)))
			return apiKey, true
		}
	}

	logger.Debug("API key validation failed",
		logger.String("key_prefix", maskAPIKey(key)))
	return nil, false
}

// GetAPIKeys returns all configured API keys
func (m *apiKeyManager) GetAPIKeys() []APIKey {
	keys := make([]APIKey, 0, len(m.apiKeys))
	for _, apiKey := range m.apiKeys {
		keys = append(keys, *apiKey)
	}
	return keys
}

// UpdateAPIKeys updates the API keys configuration
func (m *apiKeyManager) UpdateAPIKeys(apiKeys []APIKey) {
	keyMap := make(map[string]*APIKey)
	for i := range apiKeys {
		keyMap[apiKeys[i].Key] = &apiKeys[i]
	}
	m.apiKeys = keyMap

	logger.Info("API keys updated", logger.Int("count", len(apiKeys)))
}

// ValidatePermission checks if API key has required permission
func (m *apiKeyManager) ValidatePermission(apiKey *APIKey, permission string) bool {
	if apiKey == nil {
		return false
	}

	return HasPermission(apiKey.Permissions, permission)
}

// GetAPIKeyStats returns statistics about API keys
func (m *apiKeyManager) GetAPIKeyStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_keys"] = len(m.apiKeys)

	// Count keys by permission types
	permissionCounts := make(map[string]int)
	for _, apiKey := range m.apiKeys {
		for _, perm := range apiKey.Permissions {
			permissionCounts[perm]++
		}
	}
	stats["permissions"] = permissionCounts

	// Count keys with host restrictions
	keysWithHostRestrictions := 0
	for _, apiKey := range m.apiKeys {
		if len(apiKey.AllowedHosts) > 0 {
			keysWithHostRestrictions++
		}
	}
	stats["keys_with_host_restrictions"] = keysWithHostRestrictions

	return stats
}

// maskAPIKey masks API key for logging (shows only first 8 characters)
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + "..."
}

// ValidateAPIKeyFormat validates API key format (basic validation)
func ValidateAPIKeyFormat(key string) bool {
	// Basic validation rules
	if len(key) < 16 {
		return false
	}

	// Check for reasonable characters (alphanumeric, underscore, hyphen)
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}
