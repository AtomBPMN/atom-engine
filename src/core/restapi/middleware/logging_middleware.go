/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/logger"
)

// LoggingConfig holds logging middleware configuration
type LoggingConfig struct {
	Enabled              bool          `yaml:"enabled"`
	LogRequests          bool          `yaml:"log_requests"`
	LogResponses         bool          `yaml:"log_responses"`
	LogBodies            bool          `yaml:"log_bodies"`
	MaxBodySize          int           `yaml:"max_body_size"`
	SkipPaths            []string      `yaml:"skip_paths"`
	LogSlowRequests      bool          `yaml:"log_slow_requests"`
	SlowRequestThreshold time.Duration `yaml:"slow_request_threshold"`
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Enabled:              true,
		LogRequests:          true,
		LogResponses:         true,
		LogBodies:            false, // Disabled by default for security
		MaxBodySize:          1024,  // 1KB max body logging
		SkipPaths:            []string{"/health", "/metrics"},
		LogSlowRequests:      true,
		SlowRequestThreshold: 1 * time.Second,
	}
}

// LoggingMiddleware provides HTTP request/response logging
type LoggingMiddleware struct {
	config *LoggingConfig
}

// NewLoggingMiddleware creates new logging middleware
func NewLoggingMiddleware(config *LoggingConfig) *LoggingMiddleware {
	if config == nil {
		config = DefaultLoggingConfig()
	}

	return &LoggingMiddleware{
		config: config,
	}
}

// Handler provides Gin middleware for request logging
func (lm *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !lm.config.Enabled {
			c.Next()
			return
		}

		// Skip logging for configured paths
		if lm.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		// Capture request details
		reqInfo := lm.captureRequest(c)

		// Log request if enabled
		if lm.config.LogRequests {
			lm.logRequest(reqInfo)
		}

		// Capture response using custom writer
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
			config:         lm.config,
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Capture response details
		respInfo := lm.captureResponse(c, responseWriter, duration)

		// Log response if enabled
		if lm.config.LogResponses {
			lm.logResponse(reqInfo, respInfo)
		}

		// Log slow requests if enabled
		if lm.config.LogSlowRequests && duration > lm.config.SlowRequestThreshold {
			lm.logSlowRequest(reqInfo, respInfo)
		}
	}
}

// RequestInfo holds request information
type RequestInfo struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body,omitempty"`
	ClientIP  string            `json:"client_ip"`
	UserAgent string            `json:"user_agent"`
	RequestID string            `json:"request_id"`
	Timestamp time.Time         `json:"timestamp"`
}

// ResponseInfo holds response information
type ResponseInfo struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body,omitempty"`
	Size       int               `json:"size"`
	Duration   time.Duration     `json:"duration"`
}

// captureRequest captures request information
func (lm *LoggingMiddleware) captureRequest(c *gin.Context) *RequestInfo {
	reqInfo := &RequestInfo{
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		Query:     c.Request.URL.RawQuery,
		Headers:   make(map[string]string),
		ClientIP:  c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		RequestID: c.GetHeader("X-Request-ID"),
		Timestamp: time.Now(),
	}

	// Capture headers (excluding sensitive ones)
	for name, values := range c.Request.Header {
		if !lm.isSensitiveHeader(name) && len(values) > 0 {
			reqInfo.Headers[name] = values[0]
		}
	}

	// Capture request body if enabled
	if lm.config.LogBodies && c.Request.Body != nil {
		body, err := lm.readRequestBody(c)
		if err == nil {
			reqInfo.Body = body
		}
	}

	return reqInfo
}

// captureResponse captures response information
func (lm *LoggingMiddleware) captureResponse(
	c *gin.Context,
	writer *responseWriter,
	duration time.Duration,
) *ResponseInfo {
	respInfo := &ResponseInfo{
		StatusCode: c.Writer.Status(),
		Headers:    make(map[string]string),
		Size:       c.Writer.Size(),
		Duration:   duration,
	}

	// Capture response headers
	for name, values := range c.Writer.Header() {
		if len(values) > 0 {
			respInfo.Headers[name] = values[0]
		}
	}

	// Capture response body if enabled
	if lm.config.LogBodies && writer.body != nil {
		body := writer.body.String()
		if len(body) <= lm.config.MaxBodySize {
			respInfo.Body = body
		}
	}

	return respInfo
}

// logRequest logs HTTP request
func (lm *LoggingMiddleware) logRequest(reqInfo *RequestInfo) {
	fields := []logger.Field{
		logger.String("type", "http_request"),
		logger.String("method", reqInfo.Method),
		logger.String("path", reqInfo.Path),
		logger.String("client_ip", reqInfo.ClientIP),
		logger.String("user_agent", reqInfo.UserAgent),
		logger.String("request_id", reqInfo.RequestID),
	}

	if reqInfo.Query != "" {
		fields = append(fields, logger.String("query", reqInfo.Query))
	}

	if reqInfo.Body != "" {
		fields = append(fields, logger.String("body", reqInfo.Body))
	}

	logger.Info("HTTP Request", fields...)
}

// logResponse logs HTTP response
func (lm *LoggingMiddleware) logResponse(reqInfo *RequestInfo, respInfo *ResponseInfo) {
	level := logger.Info
	if respInfo.StatusCode >= 400 {
		level = logger.Warn
	}
	if respInfo.StatusCode >= 500 {
		level = logger.Error
	}

	fields := []logger.Field{
		logger.String("type", "http_response"),
		logger.String("method", reqInfo.Method),
		logger.String("path", reqInfo.Path),
		logger.Int("status_code", respInfo.StatusCode),
		logger.Int("response_size", respInfo.Size),
		logger.Any("duration", respInfo.Duration),
		logger.String("client_ip", reqInfo.ClientIP),
		logger.String("request_id", reqInfo.RequestID),
	}

	if respInfo.Body != "" {
		fields = append(fields, logger.String("response_body", respInfo.Body))
	}

	level("HTTP Response", fields...)
}

// logSlowRequest logs slow HTTP requests
func (lm *LoggingMiddleware) logSlowRequest(reqInfo *RequestInfo, respInfo *ResponseInfo) {
	logger.Warn("Slow HTTP Request",
		logger.String("type", "slow_request"),
		logger.String("method", reqInfo.Method),
		logger.String("path", reqInfo.Path),
		logger.Int("status_code", respInfo.StatusCode),
		logger.Any("duration", respInfo.Duration),
		logger.Any("threshold", lm.config.SlowRequestThreshold),
		logger.String("client_ip", reqInfo.ClientIP),
		logger.String("request_id", reqInfo.RequestID))
}

// shouldSkipPath checks if path should be skipped from logging
func (lm *LoggingMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range lm.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// isSensitiveHeader checks if header contains sensitive information
func (lm *LoggingMiddleware) isSensitiveHeader(headerName string) bool {
	sensitiveHeaders := []string{
		"Authorization", "Cookie", "Set-Cookie", "X-API-Key",
		"X-Auth-Token", "Proxy-Authorization",
	}

	for _, sensitive := range sensitiveHeaders {
		if headerName == sensitive {
			return true
		}
	}
	return false
}

// readRequestBody reads and restores request body
func (lm *LoggingMiddleware) readRequestBody(c *gin.Context) (string, error) {
	if c.Request.Body == nil {
		return "", nil
	}

	// Read body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}

	// Restore body for downstream handlers
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Limit body size for logging
	if len(bodyBytes) > lm.config.MaxBodySize {
		return string(bodyBytes[:lm.config.MaxBodySize]) + "...[truncated]", nil
	}

	return string(bodyBytes), nil
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	config *LoggingConfig
}

// Write captures response body
func (rw *responseWriter) Write(data []byte) (int, error) {
	// Write to original writer
	n, err := rw.ResponseWriter.Write(data)

	// Capture body if logging is enabled
	if rw.config.LogBodies && rw.body != nil {
		// Limit captured body size
		if rw.body.Len()+len(data) <= rw.config.MaxBodySize {
			rw.body.Write(data)
		}
	}

	return n, err
}

// GetConfig returns logging configuration
func (lm *LoggingMiddleware) GetConfig() *LoggingConfig {
	return lm.config
}

// UpdateConfig updates logging configuration
func (lm *LoggingMiddleware) UpdateConfig(config *LoggingConfig) {
	if config != nil {
		lm.config = config
		logger.Info("Logging middleware configuration updated",
			logger.Bool("enabled", config.Enabled),
			logger.Bool("log_requests", config.LogRequests),
			logger.Bool("log_responses", config.LogResponses))
	}
}
