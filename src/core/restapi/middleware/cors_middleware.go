/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/logger"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Request-ID", "X-API-Key", "User-Agent",
		},
		ExposedHeaders: []string{
			"X-Request-ID", "X-Rate-Limit-Remaining", "X-Rate-Limit-Reset",
		},
		AllowCredentials: false,
		MaxAge:           3600, // 1 hour
	}
}

// CORSMiddleware provides CORS handling middleware
type CORSMiddleware struct {
	config *CORSConfig
}

// NewCORSMiddleware creates new CORS middleware
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig()
	}

	// Set defaults for empty fields
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = DefaultCORSConfig().AllowedMethods
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = DefaultCORSConfig().AllowedHeaders
	}
	if config.MaxAge == 0 {
		config.MaxAge = DefaultCORSConfig().MaxAge
	}

	return &CORSMiddleware{
		config: config,
	}
}

// Handler provides Gin middleware for CORS
func (cm *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cm.config.Enabled {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			cm.handlePreflight(c, origin)
			return
		}

		// Handle actual requests
		cm.handleActualRequest(c, origin)
		c.Next()
	}
}

// handlePreflight handles CORS preflight requests
func (cm *CORSMiddleware) handlePreflight(c *gin.Context, origin string) {
	if !cm.isOriginAllowed(origin) {
		logger.Debug("CORS preflight: origin not allowed",
			logger.String("origin", origin))
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Set CORS headers for preflight
	cm.setOriginHeader(c, origin)
	cm.setMethodsHeader(c)
	cm.setHeadersHeader(c)
	cm.setCredentialsHeader(c)
	cm.setMaxAgeHeader(c)

	logger.Debug("CORS preflight handled",
		logger.String("origin", origin),
		logger.String("method", c.GetHeader("Access-Control-Request-Method")))

	c.AbortWithStatus(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func (cm *CORSMiddleware) handleActualRequest(c *gin.Context, origin string) {
	if !cm.isOriginAllowed(origin) {
		logger.Debug("CORS request: origin not allowed",
			logger.String("origin", origin))
		return
	}

	// Set CORS headers for actual request
	cm.setOriginHeader(c, origin)
	cm.setExposedHeadersHeader(c)
	cm.setCredentialsHeader(c)

	logger.Debug("CORS request handled",
		logger.String("origin", origin),
		logger.String("method", c.Request.Method))
}

// isOriginAllowed checks if origin is allowed
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Allow requests without Origin header (e.g., same-origin)
	}

	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// setOriginHeader sets Access-Control-Allow-Origin header
func (cm *CORSMiddleware) setOriginHeader(c *gin.Context, origin string) {
	if cm.hasWildcardOrigin() && !cm.config.AllowCredentials {
		c.Header("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	}
}

// setMethodsHeader sets Access-Control-Allow-Methods header
func (cm *CORSMiddleware) setMethodsHeader(c *gin.Context) {
	if len(cm.config.AllowedMethods) > 0 {
		methods := strings.Join(cm.config.AllowedMethods, ", ")
		c.Header("Access-Control-Allow-Methods", methods)
	}
}

// setHeadersHeader sets Access-Control-Allow-Headers header
func (cm *CORSMiddleware) setHeadersHeader(c *gin.Context) {
	requestedHeaders := c.GetHeader("Access-Control-Request-Headers")

	if requestedHeaders != "" {
		// Check if all requested headers are allowed
		requestedList := parseHeaderList(requestedHeaders)
		allowedList := make([]string, 0)

		for _, header := range requestedList {
			if cm.isHeaderAllowed(header) {
				allowedList = append(allowedList, header)
			}
		}

		if len(allowedList) > 0 {
			headers := strings.Join(allowedList, ", ")
			c.Header("Access-Control-Allow-Headers", headers)
		}
	} else if len(cm.config.AllowedHeaders) > 0 {
		headers := strings.Join(cm.config.AllowedHeaders, ", ")
		c.Header("Access-Control-Allow-Headers", headers)
	}
}

// setExposedHeadersHeader sets Access-Control-Expose-Headers header
func (cm *CORSMiddleware) setExposedHeadersHeader(c *gin.Context) {
	if len(cm.config.ExposedHeaders) > 0 {
		headers := strings.Join(cm.config.ExposedHeaders, ", ")
		c.Header("Access-Control-Expose-Headers", headers)
	}
}

// setCredentialsHeader sets Access-Control-Allow-Credentials header
func (cm *CORSMiddleware) setCredentialsHeader(c *gin.Context) {
	if cm.config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}
}

// setMaxAgeHeader sets Access-Control-Max-Age header
func (cm *CORSMiddleware) setMaxAgeHeader(c *gin.Context) {
	if cm.config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", strings.Join([]string{string(rune(cm.config.MaxAge + '0'))}, ""))
	}
}

// isHeaderAllowed checks if header is allowed
func (cm *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(strings.TrimSpace(header))

	// Always allow simple headers
	simpleHeaders := []string{
		"accept", "accept-language", "content-language", "content-type",
	}

	for _, simple := range simpleHeaders {
		if header == simple {
			return true
		}
	}

	// Check configured allowed headers
	for _, allowed := range cm.config.AllowedHeaders {
		if strings.ToLower(allowed) == header {
			return true
		}
	}

	return false
}

// hasWildcardOrigin checks if wildcard origin is configured
func (cm *CORSMiddleware) hasWildcardOrigin() bool {
	for _, origin := range cm.config.AllowedOrigins {
		if origin == "*" {
			return true
		}
	}
	return false
}

// parseHeaderList parses comma-separated header list
func parseHeaderList(headerValue string) []string {
	headers := strings.Split(headerValue, ",")
	result := make([]string, 0, len(headers))

	for _, header := range headers {
		trimmed := strings.TrimSpace(header)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// GetConfig returns CORS configuration
func (cm *CORSMiddleware) GetConfig() *CORSConfig {
	return cm.config
}

// UpdateConfig updates CORS configuration
func (cm *CORSMiddleware) UpdateConfig(config *CORSConfig) {
	if config != nil {
		cm.config = config
		logger.Info("CORS configuration updated",
			logger.Bool("enabled", config.Enabled),
			logger.Any("allowed_origins", config.AllowedOrigins))
	}
}
