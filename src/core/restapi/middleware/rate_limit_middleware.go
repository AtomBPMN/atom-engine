/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/auth"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/models"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled            bool          `yaml:"enabled"`
	RequestsPerMinute  int           `yaml:"requests_per_minute"`
	BurstSize          int           `yaml:"burst_size"`
	WindowSize         time.Duration `yaml:"window_size"`
	SkipPaths          []string      `yaml:"skip_paths"`
	UseAuthRateLimiter bool          `yaml:"use_auth_rate_limiter"`
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:            true,
		RequestsPerMinute:  100,
		BurstSize:          10,
		WindowSize:         time.Minute,
		SkipPaths:          []string{"/health", "/metrics"},
		UseAuthRateLimiter: true, // Use auth component's rate limiter
	}
}

// clientInfo holds information about client requests
type clientInfo struct {
	requests  []time.Time
	lastReset time.Time
	mutex     sync.Mutex
}

// RateLimitMiddleware provides HTTP rate limiting
type RateLimitMiddleware struct {
	config        *RateLimitConfig
	authComponent auth.Component
	clients       map[string]*clientInfo
	clientsMutex  sync.RWMutex
}

// NewRateLimitMiddleware creates new rate limit middleware
func NewRateLimitMiddleware(config *RateLimitConfig, authComponent auth.Component) *RateLimitMiddleware {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return &RateLimitMiddleware{
		config:        config,
		authComponent: authComponent,
		clients:       make(map[string]*clientInfo),
	}
}

// Handler provides Gin middleware for rate limiting
func (rlm *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rlm.config.Enabled {
			c.Next()
			return
		}

		// Skip rate limiting for configured paths
		if rlm.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract client identifier
		clientID := rlm.getClientIdentifier(c)

		// Check rate limit
		if rlm.config.UseAuthRateLimiter && rlm.authComponent != nil {
			// Use auth component's rate limiter
			if !rlm.checkAuthRateLimit(c, clientID) {
				return
			}
		} else {
			// Use built-in rate limiter (if implemented)
			if !rlm.checkBuiltinRateLimit(c, clientID) {
				return
			}
		}

		// Add rate limit headers
		rlm.addRateLimitHeaders(c, clientID)

		c.Next()
	}
}

// checkAuthRateLimit checks rate limit using auth component
func (rlm *RateLimitMiddleware) checkAuthRateLimit(c *gin.Context, clientID string) bool {
	if rlm.authComponent == nil || !rlm.authComponent.IsReady() {
		// If auth component not available, allow request
		return true
	}

	// Extract API key from request
	apiKey := rlm.extractAPIKey(c)

	// Check rate limit using auth component's rate limiter
	rateLimiter := rlm.authComponent.GetRateLimiter()
	if rateLimiter == nil {
		return true
	}

	// Check if rate limit is exceeded
	if !rateLimiter.CheckLimit(clientID, apiKey) {
		logger.Warn("Rate limit exceeded",
			logger.String("client_id", clientID),
			logger.String("api_key_prefix", maskAPIKey(apiKey)),
			logger.String("path", c.Request.URL.Path))

		// Log rate limit violation to audit
		if auditLogger := rlm.authComponent.GetAuditLogger(); auditLogger != nil {
			authCtx := auth.CreateAuthContextFromHTTP(
				clientID,
				c.GetHeader("User-Agent"),
				c.Request.Method,
				c.Request.URL.Path,
				c.GetHeader("Authorization"),
			)
			auditLogger.LogAuthFailure(authCtx, "Rate limit exceeded")
		}

		apiErr := models.RateLimitedError("Rate limit exceeded")
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse(apiErr, getRequestID(c)))
		c.Abort()
		return false
	}

	// Record the request
	rateLimiter.RecordRequest(clientID, apiKey)

	return true
}

// checkBuiltinRateLimit checks rate limit using built-in limiter
func (rlm *RateLimitMiddleware) checkBuiltinRateLimit(c *gin.Context, clientID string) bool {
	now := time.Now()
	
	// Get or create client info
	rlm.clientsMutex.RLock()
	client, exists := rlm.clients[clientID]
	rlm.clientsMutex.RUnlock()
	
	if !exists {
		// Create new client info
		client = &clientInfo{
			requests:  make([]time.Time, 0),
			lastReset: now,
		}
		rlm.clientsMutex.Lock()
		rlm.clients[clientID] = client
		rlm.clientsMutex.Unlock()
	}
	
	// Lock client for update
	client.mutex.Lock()
	defer client.mutex.Unlock()
	
	// Clean old requests (sliding window)
	cutoff := now.Add(-rlm.config.WindowSize)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests
	
	// Check if limit exceeded
	if len(client.requests) >= rlm.config.RequestsPerMinute {
		logger.Warn("Rate limit exceeded",
			logger.String("client_id", clientID),
			logger.String("path", c.Request.URL.Path),
			logger.Int("requests_count", len(client.requests)),
			logger.Int("limit", rlm.config.RequestsPerMinute))
		return false
	}
	
	// Record this request
	client.requests = append(client.requests, now)
	
	logger.Debug("Rate limit check passed",
		logger.String("client_id", clientID),
		logger.String("path", c.Request.URL.Path),
		logger.Int("requests_count", len(client.requests)),
		logger.Int("limit", rlm.config.RequestsPerMinute))
	
	return true
}

// addRateLimitHeaders adds rate limit information to response headers
func (rlm *RateLimitMiddleware) addRateLimitHeaders(c *gin.Context, clientID string) {
	// Add standard rate limit headers
	c.Header("X-RateLimit-Limit", strconv.Itoa(rlm.config.RequestsPerMinute))

	// Calculate remaining requests for built-in limiter
	remaining := rlm.config.RequestsPerMinute
	resetTime := time.Now().Add(rlm.config.WindowSize)

	if !rlm.config.UseAuthRateLimiter {
		// Get current request count for built-in limiter
		rlm.clientsMutex.RLock()
		if client, exists := rlm.clients[clientID]; exists {
			client.mutex.Lock()
			// Clean old requests for accurate count
			now := time.Now()
			cutoff := now.Add(-rlm.config.WindowSize)
			validCount := 0
			for _, reqTime := range client.requests {
				if reqTime.After(cutoff) {
					validCount++
				}
			}
			remaining = rlm.config.RequestsPerMinute - validCount
			if remaining < 0 {
				remaining = 0
			}
			client.mutex.Unlock()
		}
		rlm.clientsMutex.RUnlock()
	} else if rlm.authComponent != nil {
		// Try to get info from auth component if available
		if rateLimiter := rlm.authComponent.GetRateLimiter(); rateLimiter != nil {
			// This would require extending the RateLimiter interface for detailed info
			// For now, keep default calculation
		}
	}

	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
}

// getClientIdentifier extracts client identifier for rate limiting
func (rlm *RateLimitMiddleware) getClientIdentifier(c *gin.Context) string {
	// Try to get API key first
	if apiKey := rlm.extractAPIKey(c); apiKey != "" {
		return "api:" + apiKey
	}

	// Fallback to IP address
	return "ip:" + c.ClientIP()
}

// extractAPIKey extracts API key from request
func (rlm *RateLimitMiddleware) extractAPIKey(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Extract Bearer token
		const bearerPrefix = "Bearer "
		if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
			return authHeader[len(bearerPrefix):]
		}
	}

	// Check X-API-Key header
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check query parameter
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// shouldSkipPath checks if path should be skipped from rate limiting
func (rlm *RateLimitMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range rlm.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// maskAPIKey masks API key for logging
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}

	if len(apiKey) <= 8 {
		return "***"
	}

	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

// GetConfig returns rate limit configuration
func (rlm *RateLimitMiddleware) GetConfig() *RateLimitConfig {
	return rlm.config
}

// UpdateConfig updates rate limit configuration
func (rlm *RateLimitMiddleware) UpdateConfig(config *RateLimitConfig) {
	if config != nil {
		rlm.config = config
		logger.Info("Rate limit middleware configuration updated",
			logger.Bool("enabled", config.Enabled),
			logger.Int("requests_per_minute", config.RequestsPerMinute),
			logger.Bool("use_auth_rate_limiter", config.UseAuthRateLimiter))
	}
}

// AddSkipPath adds a path to skip rate limiting
func (rlm *RateLimitMiddleware) AddSkipPath(path string) {
	rlm.config.SkipPaths = append(rlm.config.SkipPaths, path)
}

// RateLimitInfo provides rate limit information for clients
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
}

// GetRateLimitInfo returns current rate limit information for client
func (rlm *RateLimitMiddleware) GetRateLimitInfo(clientID string) *RateLimitInfo {
	info := &RateLimitInfo{
		Limit:  rlm.config.RequestsPerMinute,
		Window: rlm.config.WindowSize,
		Reset:  time.Now().Add(rlm.config.WindowSize),
	}

	// If using auth rate limiter, try to get actual remaining count
	if rlm.config.UseAuthRateLimiter && rlm.authComponent != nil {
		if rateLimiter := rlm.authComponent.GetRateLimiter(); rateLimiter != nil {
			// This would require extending the RateLimiter interface to get remaining count
			// For now, set to unknown
			info.Remaining = -1 // -1 indicates unknown
		}
	}

	return info
}

// CleanupOldClients removes inactive clients to prevent memory leaks
func (rlm *RateLimitMiddleware) CleanupOldClients() {
	now := time.Now()
	cleanupCutoff := now.Add(-rlm.config.WindowSize * 2) // Clean clients with no activity for 2 windows
	
	rlm.clientsMutex.Lock()
	defer rlm.clientsMutex.Unlock()
	
	for clientID, client := range rlm.clients {
		client.mutex.Lock()
		// Check if client has any recent activity
		hasRecentActivity := false
		for _, reqTime := range client.requests {
			if reqTime.After(cleanupCutoff) {
				hasRecentActivity = true
				break
			}
		}
		client.mutex.Unlock()
		
		// Remove client if no recent activity
		if !hasRecentActivity {
			delete(rlm.clients, clientID)
			logger.Debug("Cleaned up inactive rate limit client",
				logger.String("client_id", clientID))
		}
	}
}

// StartCleanupWorker starts a background worker to clean up old clients
func (rlm *RateLimitMiddleware) StartCleanupWorker() {
	go func() {
		ticker := time.NewTicker(rlm.config.WindowSize)
		defer ticker.Stop()
		
		for range ticker.C {
			rlm.CleanupOldClients()
		}
	}()
}
