/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package auth

import (
	"sync"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/storage"
)

// requestInfo stores request information for rate limiting
type requestInfo struct {
	count     int
	resetTime time.Time
}

// rateLimiter implements RateLimiter interface
type rateLimiter struct {
	enabled           bool
	requestsPerMinute int
	requests          map[string]*requestInfo // map[clientIP_or_apiKey]requestInfo
	mutex             sync.RWMutex
	cleanupTicker     *time.Ticker
	stopCleanup       chan bool
	storage           StorageInterface
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(enabled bool, requestsPerMinute int) RateLimiter {
	rl := &rateLimiter{
		enabled:           enabled,
		requestsPerMinute: requestsPerMinute,
		requests:          make(map[string]*requestInfo),
		stopCleanup:       make(chan bool),
	}

	if enabled {
		// Start cleanup goroutine to remove expired entries
		rl.cleanupTicker = time.NewTicker(1 * time.Minute)
		go rl.cleanupExpiredEntries()
	}

	return rl
}

// CheckLimit verifies if request is within rate limits
func (rl *rateLimiter) CheckLimit(clientIP string, apiKey string) bool {
	if !rl.enabled {
		return true
	}

	// Use API key as identifier if available, otherwise use IP
	identifier := clientIP
	if apiKey != "" {
		identifier = "key:" + maskAPIKey(apiKey)
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	resetTime := now.Add(1 * time.Minute)

	info, exists := rl.requests[identifier]
	if !exists {
		// First request from this identifier
		rl.requests[identifier] = &requestInfo{
			count:     1,
			resetTime: resetTime,
		}
		// Save to storage if available
		rl.saveToStorage(identifier, 1, resetTime)
		return true
	}

	// Check if reset time has passed
	if now.After(info.resetTime) {
		// Reset the counter
		info.count = 1
		info.resetTime = resetTime
		// Save to storage if available
		rl.saveToStorage(identifier, info.count, info.resetTime)
		return true
	}

	// Check if limit exceeded
	if info.count >= rl.requestsPerMinute {
		logger.Debug("Rate limit exceeded",
			logger.String("identifier", identifier),
			logger.Int("count", info.count),
			logger.Int("limit", rl.requestsPerMinute))
		return false
	}

	return true
}

// RecordRequest records a request for rate limiting
func (rl *rateLimiter) RecordRequest(clientIP string, apiKey string) {
	if !rl.enabled {
		return
	}

	// Use API key as identifier if available, otherwise use IP
	identifier := clientIP
	if apiKey != "" {
		identifier = "key:" + maskAPIKey(apiKey)
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	resetTime := now.Add(1 * time.Minute)

	info, exists := rl.requests[identifier]
	if !exists {
		rl.requests[identifier] = &requestInfo{
			count:     1,
			resetTime: resetTime,
		}
		// Save to storage if available
		rl.saveToStorage(identifier, 1, resetTime)
		return
	}

	// Check if reset time has passed
	if now.After(info.resetTime) {
		// Reset the counter
		info.count = 1
		info.resetTime = resetTime
	} else {
		// Increment counter
		info.count++
	}

	// Save to storage if available
	rl.saveToStorage(identifier, info.count, info.resetTime)
}

// GetStats returns current rate limiting statistics
func (rl *rateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := map[string]interface{}{
		"enabled":             rl.enabled,
		"requests_per_minute": rl.requestsPerMinute,
		"active_identifiers":  len(rl.requests),
	}

	if rl.enabled {
		// Count requests by type
		ipRequests := 0
		apiKeyRequests := 0
		totalRequests := 0

		for identifier, info := range rl.requests {
			totalRequests += info.count
			if identifier[:4] == "key:" {
				apiKeyRequests++
			} else {
				ipRequests++
			}
		}

		stats["ip_based_identifiers"] = ipRequests
		stats["api_key_based_identifiers"] = apiKeyRequests
		stats["total_requests"] = totalRequests
	}

	return stats
}

// cleanupExpiredEntries removes expired entries from the rate limiter
func (rl *rateLimiter) cleanupExpiredEntries() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.performCleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// performCleanup removes expired entries
func (rl *rateLimiter) performCleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for identifier, info := range rl.requests {
		if now.After(info.resetTime) {
			delete(rl.requests, identifier)
			expiredCount++
		}
	}

	// Also cleanup expired entries from storage
	if rl.storage != nil {
		if err := rl.storage.CleanupExpiredRateLimitInfo(); err != nil {
			logger.Warn("Failed to cleanup expired rate limit info from storage", logger.String("error", err.Error()))
		}
	}

	if expiredCount > 0 {
		logger.Debug("Rate limiter cleanup completed",
			logger.Int("expired_entries", expiredCount),
			logger.Int("remaining_entries", len(rl.requests)))
	}
}

// Stop stops the rate limiter and cleanup goroutine
func (rl *rateLimiter) Stop() {
	if rl.enabled && rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
		close(rl.stopCleanup)
	}
}

// UpdateConfig updates rate limiter configuration
func (rl *rateLimiter) UpdateConfig(enabled bool, requestsPerMinute int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	oldEnabled := rl.enabled
	rl.enabled = enabled
	rl.requestsPerMinute = requestsPerMinute

	// Start or stop cleanup based on enabled status
	if enabled && !oldEnabled {
		rl.stopCleanup = make(chan bool)
		rl.cleanupTicker = time.NewTicker(1 * time.Minute)
		go rl.cleanupExpiredEntries()
	} else if !enabled && oldEnabled {
		rl.cleanupTicker.Stop()
		close(rl.stopCleanup)
	}

	logger.Info("Rate limiter configuration updated",
		logger.Bool("enabled", enabled),
		logger.Int("requests_per_minute", requestsPerMinute))
}

// SetStorage sets storage for persistent rate limiting
// Устанавливает storage для персистентного rate limiting
func (rl *rateLimiter) SetStorage(storage StorageInterface) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.storage = storage

	logger.Debug("Storage set for rate limiter")
}

// LoadState loads rate limiter state from storage
// Загружает состояние rate limiter из storage
func (rl *rateLimiter) LoadState() error {
	if rl.storage == nil {
		return nil // No storage available, skip loading
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Load all rate limit info from storage
	storedInfo, err := rl.storage.LoadAllRateLimitInfo()
	if err != nil {
		logger.Warn("Failed to load rate limit state from storage", logger.String("error", err.Error()))
		return err
	}

	// Convert storage format to internal format
	for identifier, info := range storedInfo {
		// Skip expired entries
		if time.Now().After(info.ResetTime) {
			continue
		}

		rl.requests[identifier] = &requestInfo{
			count:     info.Count,
			resetTime: info.ResetTime,
		}
	}

	logger.Info("Rate limiter state loaded from storage", logger.Int("entries", len(storedInfo)))
	return nil
}

// saveToStorage saves rate limit info to storage (called with mutex held)
// Сохраняет информацию о rate limit в storage (вызывается с захваченным mutex)
func (rl *rateLimiter) saveToStorage(identifier string, count int, resetTime time.Time) {
	if rl.storage == nil {
		return
	}

	info := &storage.RateLimitInfo{
		Identifier: identifier,
		Count:      count,
		ResetTime:  resetTime,
		LastAccess: time.Now(),
	}

	// Save asynchronously to avoid blocking rate limiter operations
	go func() {
		if err := rl.storage.SaveRateLimitInfo(identifier, info); err != nil {
			logger.Warn("Failed to save rate limit info to storage",
				logger.String("identifier", identifier),
				logger.String("error", err.Error()))
		}
	}()
}
