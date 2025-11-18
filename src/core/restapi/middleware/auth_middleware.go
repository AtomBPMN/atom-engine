/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/auth"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// AuthMiddleware provides HTTP authentication middleware
type AuthMiddleware struct {
	authComponent auth.Component
	bypassPaths   []string // Paths that bypass authentication
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(authComponent auth.Component) *AuthMiddleware {
	return &AuthMiddleware{
		authComponent: authComponent,
		bypassPaths: []string{
			"/health",
			"/api/health",
			"/api/v1/health",
			"/metrics",
			"/api/docs",
			"/api/v1/docs",
		},
	}
}

// Authenticate provides Gin middleware for authentication
func (am *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if auth is enabled
		if !am.authComponent.IsEnabled() {
			// Auth disabled - allow all requests
			c.Next()
			return
		}

		// Check if path should bypass auth
		if am.shouldBypassAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract client IP
		clientIP := am.extractClientIP(c)

		// Extract authorization header
		authHeader := c.GetHeader("Authorization")

		// Extract user agent
		userAgent := c.GetHeader("User-Agent")

		// Create auth context from HTTP request
		authCtx := auth.CreateAuthContextFromHTTP(
			clientIP,
			userAgent,
			c.Request.Method,
			c.Request.URL.Path,
			authHeader,
		)

		// Validate auth context
		if err := auth.ValidateAuthContext(authCtx); err != nil {
			logger.Warn("Invalid auth context",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("error", err.Error()))

			apiErr := models.BadRequestError("Invalid request context")
			c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		// Authenticate
		authResult, err := am.authComponent.Authenticate(authCtx)
		if err != nil {
			logger.Error("Authentication error",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("error", err.Error()))

			apiErr := models.InternalServerError("Authentication failed")
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		if !authResult.Authenticated {
			logger.Warn("Authentication failed",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("client_ip", authCtx.ClientIP),
				logger.String("reason", authResult.Reason))

			// Return appropriate error based on reason
			var apiErr *models.APIError
			var statusCode int

			switch {
			case strings.Contains(authResult.Reason, "Rate limit"):
				apiErr = models.RateLimitedError("Rate limit exceeded")
				statusCode = http.StatusTooManyRequests
			case strings.Contains(authResult.Reason, "IP"):
				apiErr = models.ForbiddenError("IP address not allowed")
				statusCode = http.StatusForbidden
			case strings.Contains(authResult.Reason, "API key"):
				apiErr = models.UnauthorizedError("Invalid or missing API key")
				statusCode = http.StatusUnauthorized
			default:
				apiErr = models.UnauthorizedError("Authentication failed")
				statusCode = http.StatusUnauthorized
			}

			c.JSON(statusCode, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		// Add auth result to context for downstream use
		c.Set("auth_result", authResult)
		c.Set("auth_context", authCtx)

		logger.Debug("HTTP Request authenticated",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", authCtx.ClientIP),
			logger.String("api_key_name", authResult.APIKeyName))

		c.Next()
	}
}

// RequirePermission middleware that checks for specific permission
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth result from context
		authResult, exists := c.Get("auth_result")
		if !exists {
			apiErr := models.InternalServerError("Authentication context not found")
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		result, ok := authResult.(*auth.AuthResult)
		if !ok {
			apiErr := models.InternalServerError("Invalid authentication context")
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		if !auth.HasPermission(result.Permissions, permission) {
			logger.Warn("Insufficient permissions",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("required_permission", permission),
				logger.Any("user_permissions", result.Permissions))

			apiErr := models.ForbiddenError("Insufficient permissions")
			c.JSON(http.StatusForbidden, models.ErrorResponse(apiErr, getRequestID(c)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// shouldBypassAuth checks if path should bypass authentication
func (am *AuthMiddleware) shouldBypassAuth(path string) bool {
	for _, bypassPath := range am.bypassPaths {
		if path == bypassPath || strings.HasPrefix(path, bypassPath) {
			return true
		}
	}
	return false
}

// extractClientIP extracts client IP from request
func (am *AuthMiddleware) extractClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for load balancers)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return c.ClientIP()
}

// AddBypassPath adds a path to bypass authentication
func (am *AuthMiddleware) AddBypassPath(path string) {
	am.bypassPaths = append(am.bypassPaths, path)
}

// GetAuthResult extracts auth result from Gin context
func GetAuthResult(c *gin.Context) (*auth.AuthResult, bool) {
	if authResult, exists := c.Get("auth_result"); exists {
		if result, ok := authResult.(*auth.AuthResult); ok {
			return result, true
		}
	}
	return nil, false
}

// GetAuthContext extracts auth context from Gin context
func GetAuthContext(c *gin.Context) (*auth.AuthContext, bool) {
	if authCtx, exists := c.Get("auth_context"); exists {
		if ctx, ok := authCtx.(*auth.AuthContext); ok {
			return ctx, true
		}
	}
	return nil, false
}

// getRequestID gets or generates request ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate simple request ID if not provided
	return generateRequestID()
}

// generateRequestID generates a cryptographically secure request ID
func generateRequestID() string {
	return utils.GenerateSecureRequestID("req")
}

// AuthContextKey is the context key for auth result
type AuthContextKey string

const (
	AuthResultKey AuthContextKey = "auth_result"
	AuthCtxKey    AuthContextKey = "auth_context"
)

// SetAuthContext sets auth context in standard context
func SetAuthContext(ctx context.Context, authResult *auth.AuthResult) context.Context {
	return context.WithValue(ctx, AuthResultKey, authResult)
}

// GetAuthResultFromContext extracts auth result from standard context
func GetAuthResultFromContext(ctx context.Context) (*auth.AuthResult, bool) {
	authResult, ok := ctx.Value(AuthResultKey).(*auth.AuthResult)
	return authResult, ok
}
