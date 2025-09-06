/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	restmodels "atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
	"atom-engine/src/core/types"
)

// SystemHandler handles system monitoring and management HTTP requests
type SystemHandler struct {
	coreInterface SystemCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// SystemCoreInterface defines methods needed for system operations
type SystemCoreInterface interface {
	// Typed system methods
	GetSystemStatus() (*types.SystemStatus, error)
	GetSystemInfo() (*types.SystemInfo, error)
	GetSystemMetrics() (*types.SystemMetrics, error)
	ListComponents(req *types.ComponentListRequest) (*types.ComponentListResponse, error)
	GetComponentStatus(componentName string) (*types.ComponentInfo, error)
	HealthCheck(req *types.ComponentHealthCheckRequest) (*types.ComponentHealthCheckResponse, error)
}

// NewSystemHandler creates new system handler
func NewSystemHandler(coreInterface SystemCoreInterface) *SystemHandler {
	return &SystemHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers system routes
func (h *SystemHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	system := router.Group("/system")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		system.Use(authMiddleware.RequirePermission("system"))
	}

	{
		// System information endpoints
		system.GET("/status", h.GetSystemStatus)
		system.GET("/info", h.GetSystemInfo)
		system.GET("/metrics", h.GetSystemMetrics)
		system.GET("/health", h.SystemHealthCheck)

		// Component management endpoints
		system.GET("/components", h.ListComponents)
		system.GET("/components/:name", h.GetComponentStatus)
		system.GET("/components/:name/health", h.ComponentHealthCheck)
	}
}

// GetSystemStatus handles GET /api/v1/system/status
// @Summary Get system status
// @Description Get comprehensive system status information
// @Tags system
// @Produce json
// @Success 200 {object} restmodels.APIResponse{data=types.SystemStatus}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/status [get]
func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting system status",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Get system status using typed method
	result, err := h.coreInterface.GetSystemStatus()
	if err != nil {
		logger.Error("Failed to get system status",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got system status",
		logger.String("request_id", requestID),
		logger.String("status", string(result.Status)),
		logger.String("health", string(result.Health)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetSystemInfo handles GET /api/v1/system/info
// @Summary Get system information
// @Description Get detailed system information including host details
// @Tags system
// @Produce json
// @Success 200 {object} restmodels.APIResponse{data=types.SystemInfo}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/info [get]
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting system info",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Get system info using typed method
	result, err := h.coreInterface.GetSystemInfo()
	if err != nil {
		logger.Error("Failed to get system info",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got system info",
		logger.String("request_id", requestID),
		logger.String("name", result.Name),
		logger.String("version", result.Version))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetSystemMetrics handles GET /api/v1/system/metrics
// @Summary Get system metrics
// @Description Get real-time system performance metrics
// @Tags system
// @Produce json
// @Success 200 {object} restmodels.APIResponse{data=types.SystemMetrics}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/metrics [get]
func (h *SystemHandler) GetSystemMetrics(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting system metrics",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Get system metrics using typed method
	result, err := h.coreInterface.GetSystemMetrics()
	if err != nil {
		logger.Error("Failed to get system metrics",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got system metrics",
		logger.String("request_id", requestID),
		logger.Int64("memory_usage", result.MemoryUsage),
		logger.Int("goroutines", int(result.Goroutines)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// SystemHealthCheck handles GET /api/v1/system/health
// @Summary Perform system health check
// @Description Perform comprehensive system health check
// @Tags system
// @Produce json
// @Param deep query bool false "Deep health check" default(false)
// @Success 200 {object} restmodels.APIResponse{data=types.ComponentHealthCheckResponse}
// @Failure 503 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/health [get]
func (h *SystemHandler) SystemHealthCheck(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters
	deepCheck := c.Query("deep") == "true"

	timeout := 30 * time.Second
	req := &types.ComponentHealthCheckRequest{
		ComponentName: "", // Empty for system-wide check
		Deep:          deepCheck,
		Timeout:       &timeout,
	}

	logger.Debug("Performing system health check",
		logger.String("request_id", requestID),
		logger.Bool("deep_check", deepCheck),
		logger.String("client_ip", c.ClientIP()))

	// Perform health check using typed method
	result, err := h.coreInterface.HealthCheck(req)
	if err != nil {
		logger.Error("Failed to perform system health check",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Determine HTTP status code based on health
	var httpStatus int
	switch result.Health {
	case types.ComponentHealthHealthy:
		httpStatus = http.StatusOK
	case types.ComponentHealthDegraded:
		httpStatus = http.StatusOK // Still operational
	case types.ComponentHealthUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	logger.Debug("System health check completed",
		logger.String("request_id", requestID),
		logger.String("health", string(result.Health)),
		logger.String("status", string(result.Status)))

	c.JSON(httpStatus, restmodels.SuccessResponse(result, requestID))
}

// ListComponents handles GET /api/v1/system/components
// @Summary List system components
// @Description Get list of system components with filtering options
// @Tags system
// @Produce json
// @Param type query string false "Component type filter"
// @Param status query string false "Component status filter"
// @Param health query string false "Component health filter"
// @Param enabled_only query bool false "Show only enabled components" default(false)
// @Param ready_only query bool false "Show only ready components" default(false)
// @Success 200 {object} restmodels.APIResponse{data=types.ComponentListResponse}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/components [get]
func (h *SystemHandler) ListComponents(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters into typed request
	req := &types.ComponentListRequest{
		EnabledOnly: c.Query("enabled_only") == "true",
		ReadyOnly:   c.Query("ready_only") == "true",
	}

	if compType := c.Query("type"); compType != "" {
		componentType := types.ComponentType(compType)
		req.Type = &componentType
	}

	if status := c.Query("status"); status != "" {
		componentStatus := types.ComponentStatus(status)
		req.Status = &componentStatus
	}

	if health := c.Query("health"); health != "" {
		componentHealth := types.ComponentHealth(health)
		req.Health = &componentHealth
	}

	logger.Debug("Listing system components",
		logger.String("request_id", requestID),
		logger.Any("request", req),
		logger.String("client_ip", c.ClientIP()))

	// List components using typed method
	result, err := h.coreInterface.ListComponents(req)
	if err != nil {
		logger.Error("Failed to list system components",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Listed system components",
		logger.String("request_id", requestID),
		logger.Int("count", int(result.TotalCount)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetComponentStatus handles GET /api/v1/system/components/:name
// @Summary Get component status
// @Description Get detailed status information for a specific component
// @Tags system
// @Produce json
// @Param name path string true "Component name"
// @Success 200 {object} restmodels.APIResponse{data=types.ComponentInfo}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/components/{name} [get]
func (h *SystemHandler) GetComponentStatus(c *gin.Context) {
	requestID := h.getRequestID(c)
	componentName := c.Param("name")

	if componentName == "" {
		apiErr := restmodels.BadRequestError("Component name is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting component status",
		logger.String("request_id", requestID),
		logger.String("component_name", componentName),
		logger.String("client_ip", c.ClientIP()))

	// Get component status using typed method
	result, err := h.coreInterface.GetComponentStatus(componentName)
	if err != nil {
		logger.Error("Failed to get component status",
			logger.String("request_id", requestID),
			logger.String("component_name", componentName),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got component status",
		logger.String("request_id", requestID),
		logger.String("component_name", componentName),
		logger.String("status", string(result.Status)),
		logger.String("health", string(result.Health)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// ComponentHealthCheck handles GET /api/v1/system/components/:name/health
// @Summary Perform component health check
// @Description Perform health check for a specific component
// @Tags system
// @Produce json
// @Param name path string true "Component name"
// @Param deep query bool false "Deep health check" default(false)
// @Param timeout query int false "Timeout in seconds" default(10)
// @Success 200 {object} restmodels.APIResponse{data=types.ComponentHealthCheckResponse}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 503 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/system/components/{name}/health [get]
func (h *SystemHandler) ComponentHealthCheck(c *gin.Context) {
	requestID := h.getRequestID(c)
	componentName := c.Param("name")

	if componentName == "" {
		apiErr := restmodels.BadRequestError("Component name is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse query parameters
	deepCheck := c.Query("deep") == "true"
	timeout := 10 // Default timeout
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = t
		}
	}

	timeoutDuration := time.Duration(timeout) * time.Second
	req := &types.ComponentHealthCheckRequest{
		ComponentName: componentName,
		Deep:          deepCheck,
		Timeout:       &timeoutDuration,
	}

	logger.Debug("Performing component health check",
		logger.String("request_id", requestID),
		logger.String("component_name", componentName),
		logger.Bool("deep_check", deepCheck),
		logger.Int("timeout", timeout),
		logger.String("client_ip", c.ClientIP()))

	// Perform health check using typed method
	result, err := h.coreInterface.HealthCheck(req)
	if err != nil {
		logger.Error("Failed to perform component health check",
			logger.String("request_id", requestID),
			logger.String("component_name", componentName),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Determine HTTP status code based on health
	var httpStatus int
	switch result.Health {
	case types.ComponentHealthHealthy:
		httpStatus = http.StatusOK
	case types.ComponentHealthDegraded:
		httpStatus = http.StatusOK // Still operational
	case types.ComponentHealthUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	logger.Debug("Component health check completed",
		logger.String("request_id", requestID),
		logger.String("component_name", componentName),
		logger.String("health", string(result.Health)),
		logger.String("status", string(result.Status)))

	c.JSON(httpStatus, restmodels.SuccessResponse(result, requestID))
}

// Helper methods

// getRequestID extracts request ID from gin context
func (h *SystemHandler) getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return "unknown"
}
