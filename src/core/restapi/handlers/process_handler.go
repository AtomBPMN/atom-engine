/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/grpc"
	"atom-engine/src/core/interfaces"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/core/restapi/middleware"
	restmodels "atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
	"atom-engine/src/core/types"
)

// ProcessHandler handles process management HTTP requests
type ProcessHandler struct {
	coreInterface ProcessCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// ProcessCoreInterface defines methods needed for process operations
type ProcessCoreInterface interface {
	// Legacy interface access
	GetProcessComponent() grpc.ProcessComponentInterface

	// New typed interface access
	GetProcessComponentTyped() interfaces.ProcessComponentTypedInterface

	// Core typed methods for process operations
	StartProcessTyped(req *types.ProcessStartRequest) (*types.ProcessStartResponse, error)
	CancelProcessTyped(req *types.ProcessCancelRequest) (*types.ProcessCancelResponse, error)
	GetSystemStatus() (*types.SystemStatus, error)
	GetSystemMetrics() (*types.SystemMetrics, error)
}

// ProcessComponentInterface defines process component interface
type ProcessComponentInterface interface {
	StartProcessInstance(processKey string, variables map[string]interface{}) (*ProcessInstanceResult, error)
	GetProcessInstanceStatus(instanceID string) (*ProcessInstanceResult, error)
	CancelProcessInstance(instanceID string, reason string) error
	ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*ProcessInstanceResult, error)
	GetActiveTokens(instanceID string) ([]*models.Token, error)
	GetTokensByProcessInstance(instanceID string) ([]*models.Token, error)
}

// Process data types
type ProcessInstanceResult struct {
	InstanceID      string                 `json:"instance_id"`
	ProcessID       string                 `json:"process_id"`
	ProcessName     string                 `json:"process_name"`
	State           string                 `json:"state"`
	CurrentActivity string                 `json:"current_activity"`
	StartedAt       int64                  `json:"started_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	CompletedAt     int64                  `json:"completed_at,omitempty"`
	Variables       map[string]interface{} `json:"variables"`
}

type Token struct {
	ID                string                 `json:"id"`
	State             TokenState             `json:"state"`
	ElementID         string                 `json:"element_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
	Variables         map[string]interface{} `json:"variables"`
}

type TokenState string

const (
	TokenStateActive    TokenState = "ACTIVE"
	TokenStateCompleted TokenState = "COMPLETED"
	TokenStateCancelled TokenState = "CANCELLED"
)

// NewProcessHandler creates new process handler
func NewProcessHandler(coreInterface ProcessCoreInterface) *ProcessHandler {
	return &ProcessHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers process routes
func (h *ProcessHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	processes := router.Group("/processes")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		processes.Use(authMiddleware.RequirePermission("process"))
	}

	{
		// Legacy v1 endpoints - maintain backward compatibility
		processes.POST("", h.StartProcess)
		processes.GET("", h.ListProcesses)
		processes.GET("/:id", h.GetProcessStatus)
		processes.GET("/:id/info", h.GetProcessInfo)
		processes.DELETE("/:id", h.CancelProcess)
		processes.GET("/:id/tokens", h.GetProcessTokens)
		processes.GET("/:id/tokens/trace", h.GetTokenTrace)

		// New typed endpoints for enhanced functionality
		processes.POST("/typed", h.StartProcessTyped)
		processes.GET("/typed", h.ListProcessesTyped)
		processes.GET("/:id/typed", h.GetProcessStatusTyped)
		processes.DELETE("/:id/typed", h.CancelProcessTyped)
		processes.GET("/:id/tokens/typed", h.GetProcessTokensTyped)
		processes.GET("/:id/trace/typed", h.TraceProcessExecutionTyped)
		processes.GET("/stats", h.GetProcessStatsHandler)
	}
}

// StartProcess handles POST /api/v1/processes
// @Summary Start process instance
// @Description Start a new process instance with optional variables
// @Tags processes
// @Accept json
// @Produce json
// @Param request body restmodels.StartProcessRequest true "Process start request"
// @Success 201 {object} restmodels.APIResponse{data=ProcessInstanceResult}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 401 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 403 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes [post]
func (h *ProcessHandler) StartProcess(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req restmodels.StartProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse start process request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := restmodels.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if apiErr, ok := err.(*restmodels.APIError); ok {
			c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		} else {
			c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(restmodels.BadRequestError(err.Error()), requestID))
		}
		return
	}

	logger.Debug("Starting process instance",
		logger.String("request_id", requestID),
		logger.String("process_key", req.ProcessKey),
		logger.String("client_ip", c.ClientIP()))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		logger.Error("Process component not available",
			logger.String("request_id", requestID))

		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Start process instance
	result, err := processComp.StartProcessInstance(req.ProcessKey, req.Variables)
	if err != nil {
		logger.Error("Failed to start process instance",
			logger.String("request_id", requestID),
			logger.String("process_key", req.ProcessKey),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Process instance started",
		logger.String("request_id", requestID),
		logger.String("process_key", req.ProcessKey),
		logger.String("instance_id", result.InstanceID))

	c.JSON(http.StatusCreated, restmodels.SuccessResponse(result, requestID))
}

// ListProcesses handles GET /api/v1/processes
// @Summary List process instances
// @Description Get list of process instances with filtering and pagination
// @Tags processes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Status filter (active, completed, cancelled)"
// @Param process_key query string false "Process key filter"
// @Param tenant_id query string false "Tenant ID filter"
// @Success 200 {object} restmodels.PaginatedResponse{data=[]ProcessInstanceResult}
// @Failure 401 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 403 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes [get]
func (h *ProcessHandler) ListProcesses(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	processKey := c.Query("process_key")
	_ = c.Query("tenant_id") // tenantID for future implementation

	// Parse and validate pagination
	paginationHelper := utils.NewPaginationHelper()
	params, apiErr := paginationHelper.ParseAndValidate(pageStr, limitStr)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate status filter
	if status != "" {
		validStatuses := []string{"active", "completed", "cancelled"}
		if apiErr := h.validator.ValidateStringEnum(status, "status", validStatuses); apiErr != nil {
			c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(
				restmodels.NewValidationError("Invalid status filter", []restmodels.ValidationError{*apiErr}),
				requestID))
			return
		}
	}

	logger.Debug("Listing process instances",
		logger.String("request_id", requestID),
		logger.Int("page", params.Page),
		logger.Int("limit", params.Limit),
		logger.String("status", status),
		logger.String("process_key", processKey))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// List process instances (load all for sorting)
	instances, err := processComp.ListProcessInstances(status, processKey, 0)
	if err != nil {
		logger.Error("Failed to list process instances",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Apply sorting by started_at DESC (consistent with gRPC/CLI behavior)
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].StartedAt > instances[j].StartedAt
	})

	// Apply client-side pagination after sorting
	paginatedInstances, paginationInfo := utils.ApplyPagination(instances, params.Page, params.Limit)

	logger.Info("Process instances listed",
		logger.String("request_id", requestID),
		logger.Int("count", len(instances)),
		logger.Int("page", params.Page))

	paginatedResp := restmodels.PaginatedSuccessResponse(paginatedInstances, paginationInfo, requestID)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetProcessStatus handles GET /api/v1/processes/:id
// @Summary Get process instance status
// @Description Get detailed status of a specific process instance
// @Tags processes
// @Produce json
// @Param id path string true "Process instance ID"
// @Success 200 {object} restmodels.APIResponse{data=ProcessInstanceResult}
// @Failure 401 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 403 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id} [get]
func (h *ProcessHandler) GetProcessStatus(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate instance ID format
	if apiErr := h.validator.ValidateID(instanceID, "instance_id"); apiErr != nil {
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(
			restmodels.NewValidationError("Invalid instance ID format", []restmodels.ValidationError{*apiErr}),
			requestID))
		return
	}

	logger.Debug("Getting process instance status",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get process status
	result, err := processComp.GetProcessInstanceStatus(instanceID)
	if err != nil {
		logger.Error("Failed to get process instance status",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		if apiErr.Code == restmodels.ErrorCodeResourceNotFound {
			apiErr = restmodels.ProcessNotFoundError(instanceID)
		}
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Process instance status retrieved",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.String("state", result.State))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetProcessInfo handles GET /api/v1/processes/:id/info
// @Summary Get complete process instance information
// @Description Get detailed information about a process instance including tokens, timers, jobs, messages, and incidents
// @Tags processes
// @Produce json
// @Param id path string true "Process instance ID"
// @Success 200 {object} restmodels.APIResponse{data=map[string]interface{}}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 401 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 403 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id}/info [get]
func (h *ProcessHandler) GetProcessInfo(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate instance ID format
	if apiErr := h.validator.ValidateID(instanceID, "instance_id"); apiErr != nil {
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(
			restmodels.NewValidationError("Invalid instance ID format", []restmodels.ValidationError{*apiErr}),
			requestID))
		return
	}

	logger.Debug("Getting complete process instance information",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	// Use adapter to get complete process info
	coreInterface, ok := h.coreInterface.(interfaces.CoreTypedInterface)
	if !ok {
		logger.Error("Core typed interface not available",
			logger.String("request_id", requestID))
		apiErr := restmodels.InternalServerError("Service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get complete process info through adapter
	processInfo, err := coreInterface.GetProcessInfoForREST(instanceID)
	if err != nil {
		logger.Error("Failed to get complete process information",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Process information retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(processInfo, requestID))
}

func (h *ProcessHandler) CancelProcess(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse optional request body
	var req restmodels.CancelProcessRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Failed to parse cancel request body, using defaults",
				logger.String("request_id", requestID),
				logger.String("error", err.Error()))
		}
	}

	logger.Debug("Cancelling process instance",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.String("reason", req.Reason))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Cancel process instance
	err := processComp.CancelProcessInstance(instanceID, req.Reason)
	if err != nil {
		logger.Error("Failed to cancel process instance",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		if apiErr.Code == restmodels.ErrorCodeResourceNotFound {
			apiErr = restmodels.ProcessNotFoundError(instanceID)
		}
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	response := &restmodels.DeleteResponse{
		ID:      instanceID,
		Message: "Process instance cancelled successfully",
	}

	logger.Info("Process instance cancelled",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(response, requestID))
}

// GetProcessTokens handles GET /api/v1/processes/:id/tokens
func (h *ProcessHandler) GetProcessTokens(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	logger.Debug("Getting process tokens",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		logger.Error("Process component not available",
			logger.String("request_id", requestID))

		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get active tokens for the process instance
	tokens, err := processComp.GetActiveTokens(instanceID)
	if err != nil {
		logger.Error("Failed to get process tokens",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		if apiErr.Code == restmodels.ErrorCodeResourceNotFound {
			apiErr = restmodels.ProcessNotFoundError(instanceID)
		}
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert to REST API token format
	restTokens := make([]*Token, len(tokens))
	for i, token := range tokens {
		restTokens[i] = &Token{
			ID:                token.TokenID,
			State:             TokenState(token.State),
			ElementID:         token.CurrentElementID,
			ProcessInstanceID: token.ProcessInstanceID,
			CreatedAt:         token.CreatedAt.Unix(),
			UpdatedAt:         token.UpdatedAt.Unix(),
			Variables:         token.Variables,
		}
	}

	logger.Info("Process tokens retrieved",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Int("tokens_count", len(restTokens)))

	pagination := &restmodels.PaginationInfo{
		Page:    1,
		Limit:   len(restTokens),
		Total:   len(restTokens),
		Pages:   1,
		HasNext: false,
		HasPrev: false,
	}

	c.JSON(http.StatusOK, restmodels.PaginatedSuccessResponse(restTokens, pagination, requestID))
}

// GetTokenTrace handles GET /api/v1/processes/:id/tokens/trace
func (h *ProcessHandler) GetTokenTrace(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	logger.Debug("Getting token trace",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	// Get process component
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		logger.Error("Process component not available",
			logger.String("request_id", requestID))

		apiErr := restmodels.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get all tokens for the process instance (for trace)
	tokens, err := processComp.GetTokensByProcessInstance(instanceID)
	if err != nil {
		logger.Error("Failed to get token trace",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		if apiErr.Code == restmodels.ErrorCodeResourceNotFound {
			apiErr = restmodels.ProcessNotFoundError(instanceID)
		}
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert to REST API token format and sort by creation time
	restTokens := make([]*Token, len(tokens))
	for i, token := range tokens {
		restTokens[i] = &Token{
			ID:                token.TokenID,
			State:             TokenState(token.State),
			ElementID:         token.CurrentElementID,
			ProcessInstanceID: token.ProcessInstanceID,
			CreatedAt:         token.CreatedAt.Unix(),
			UpdatedAt:         token.UpdatedAt.Unix(),
			Variables:         token.Variables,
		}
	}

	logger.Info("Token trace retrieved",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Int("tokens_count", len(restTokens)))

	pagination := &restmodels.PaginationInfo{
		Page:    1,
		Limit:   len(restTokens),
		Total:   len(restTokens),
		Pages:   1,
		HasNext: false,
		HasPrev: false,
	}

	c.JSON(http.StatusOK, restmodels.PaginatedSuccessResponse(restTokens, pagination, requestID))
}

// Helper methods

func (h *ProcessHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("process")
}

// ProcessStats provides process statistics
type ProcessStats struct {
	TotalInstances       int64            `json:"total_instances"`
	ActiveInstances      int64            `json:"active_instances"`
	CompletedInstances   int64            `json:"completed_instances"`
	CancelledInstances   int64            `json:"cancelled_instances"`
	InstancesByStatus    map[string]int64 `json:"instances_by_status"`
	InstancesByProcess   map[string]int64 `json:"instances_by_process"`
	AverageExecutionTime float64          `json:"average_execution_time_ms"`
}

// GetProcessStats returns process statistics
func (h *ProcessHandler) GetProcessStats() (*ProcessStats, error) {
	processComp := h.coreInterface.GetProcessComponent()
	if processComp == nil {
		return nil, fmt.Errorf("process component not available")
	}

	// Get all instances to calculate stats
	allInstances, err := processComp.ListProcessInstances("", "", 0)
	if err != nil {
		return nil, err
	}

	stats := &ProcessStats{
		TotalInstances:     int64(len(allInstances)),
		InstancesByStatus:  make(map[string]int64),
		InstancesByProcess: make(map[string]int64),
	}

	// Calculate statistics
	for _, instance := range allInstances {
		switch instance.State {
		case "ACTIVE":
			stats.ActiveInstances++
		case "COMPLETED":
			stats.CompletedInstances++
		case "CANCELLED":
			stats.CancelledInstances++
		}

		stats.InstancesByStatus[instance.State]++
		stats.InstancesByProcess[instance.ProcessID]++
	}

	return stats, nil
}

// Typed API handlers for enhanced functionality
// Typed API обработчики для расширенной функциональности

// StartProcessTyped handles POST /api/v1/processes/typed
// @Summary Start process instance with typed request/response
// @Description Start a new process instance using strongly typed API
// @Tags processes
// @Accept json
// @Produce json
// @Param request body types.ProcessStartRequest true "Typed process start request"
// @Success 201 {object} restmodels.APIResponse{data=types.ProcessStartResponse}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/typed [post]
func (h *ProcessHandler) StartProcessTyped(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse typed request body
	var req types.ProcessStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse typed start process request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := restmodels.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Starting process instance with typed API",
		logger.String("request_id", requestID),
		logger.String("process_key", req.ProcessKey),
		logger.String("client_ip", c.ClientIP()))

	// Use Core typed method
	result, err := h.coreInterface.StartProcessTyped(&req)
	if err != nil {
		logger.Error("Failed to start process instance via typed API",
			logger.String("request_id", requestID),
			logger.String("process_key", req.ProcessKey),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Process instance started via typed API",
		logger.String("request_id", requestID),
		logger.String("process_key", req.ProcessKey),
		logger.String("instance_id", result.InstanceID))

	c.JSON(http.StatusCreated, restmodels.SuccessResponse(result, requestID))
}

// ListProcessesTyped handles GET /api/v1/processes/typed
// @Summary List process instances with typed response
// @Description Get list of process instances using strongly typed API
// @Tags processes
// @Produce json
// @Param process_key query string false "Process key filter"
// @Param status query string false "Status filter"
// @Param tenant_id query string false "Tenant ID filter"
// @Param limit query int false "Items per page" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} restmodels.APIResponse{data=types.ProcessListResponse}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/typed [get]
func (h *ProcessHandler) ListProcessesTyped(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters into typed request
	req := &types.ProcessListRequest{
		Limit:  20, // Default
		Offset: 0,  // Default
	}

	if processKey := c.Query("process_key"); processKey != "" {
		req.ProcessKey = &processKey
	}

	if status := c.Query("status"); status != "" {
		processStatus := types.ProcessStatus(status)
		req.Status = &processStatus
	}

	if tenantID := c.Query("tenant_id"); tenantID != "" {
		req.TenantID = &tenantID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = int32(limit)
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = int32(offset)
		}
	}

	logger.Debug("Listing process instances with typed API",
		logger.String("request_id", requestID),
		logger.Any("request", req))

	// Get typed process component
	processComp := h.coreInterface.GetProcessComponentTyped()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Typed process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// List process instances using typed method
	result, err := processComp.ListProcessInstancesTyped(req)
	if err != nil {
		logger.Error("Failed to list process instances via typed API",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Listed process instances via typed API",
		logger.String("request_id", requestID),
		logger.Int("count", int(result.TotalCount)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetProcessStatusTyped handles GET /api/v1/processes/:id/typed
// @Summary Get process instance status with typed response
// @Description Get detailed process instance information using strongly typed API
// @Tags processes
// @Produce json
// @Param id path string true "Process instance ID"
// @Success 200 {object} restmodels.APIResponse{data=types.ProcessInstanceDetails}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id}/typed [get]
func (h *ProcessHandler) GetProcessStatusTyped(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting process status with typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	// Get typed process component
	processComp := h.coreInterface.GetProcessComponentTyped()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Typed process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get process status using typed method
	result, err := processComp.GetProcessInstanceStatusTyped(instanceID)
	if err != nil {
		logger.Error("Failed to get process status via typed API",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got process status via typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.String("status", string(result.Status)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// CancelProcessTyped handles DELETE /api/v1/processes/:id/typed
// @Summary Cancel process instance with typed request/response
// @Description Cancel a process instance using strongly typed API
// @Tags processes
// @Accept json
// @Produce json
// @Param id path string true "Process instance ID"
// @Param request body types.ProcessCancelRequest true "Typed cancel request"
// @Success 200 {object} restmodels.APIResponse{data=types.ProcessCancelResponse}
// @Failure 400 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id}/typed [delete]
func (h *ProcessHandler) CancelProcessTyped(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse typed request body
	var req types.ProcessCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body - just set instance ID from path
		req.InstanceID = instanceID
	} else {
		// Override instance ID from path parameter
		req.InstanceID = instanceID
	}

	logger.Debug("Cancelling process with typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.String("reason", req.Reason))

	// Use Core typed method
	result, err := h.coreInterface.CancelProcessTyped(&req)
	if err != nil {
		logger.Error("Failed to cancel process via typed API",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Process cancelled via typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetProcessTokensTyped handles GET /api/v1/processes/:id/tokens/typed
// @Summary Get process tokens with typed response
// @Description Get tokens for a process instance using strongly typed API
// @Tags processes
// @Produce json
// @Param id path string true "Process instance ID"
// @Param active_only query bool false "Return only active tokens" default(false)
// @Param limit query int false "Items per page" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} restmodels.APIResponse{data=types.TokenListResponse}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id}/tokens/typed [get]
func (h *ProcessHandler) GetProcessTokensTyped(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse query parameters
	req := &types.TokenListRequest{
		ProcessInstanceID: &instanceID,
		ActiveOnly:        c.Query("active_only") == "true",
		Limit:             50, // Default
		Offset:            0,  // Default
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = int32(limit)
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = int32(offset)
		}
	}

	logger.Debug("Getting process tokens with typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Bool("active_only", req.ActiveOnly))

	// Get typed process component
	processComp := h.coreInterface.GetProcessComponentTyped()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Typed process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get tokens using typed method
	result, err := processComp.GetTokensTyped(req)
	if err != nil {
		logger.Error("Failed to get process tokens via typed API",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got process tokens via typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Int("count", int(result.TotalCount)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// TraceProcessExecutionTyped handles GET /api/v1/processes/:id/trace/typed
// @Summary Trace process execution with typed response
// @Description Get detailed execution trace for a process instance using strongly typed API
// @Tags processes
// @Produce json
// @Param id path string true "Process instance ID"
// @Param include_variables query bool false "Include variable details" default(true)
// @Param include_metadata query bool false "Include metadata" default(false)
// @Success 200 {object} restmodels.APIResponse{data=types.ProcessTraceResponse}
// @Failure 404 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/{id}/trace/typed [get]
func (h *ProcessHandler) TraceProcessExecutionTyped(c *gin.Context) {
	requestID := h.getRequestID(c)
	instanceID := c.Param("id")

	if instanceID == "" {
		apiErr := restmodels.BadRequestError("Process instance ID is required")
		c.JSON(http.StatusBadRequest, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse query parameters
	req := &types.ProcessTraceRequest{
		ProcessInstanceID: instanceID,
		IncludeVariables:  c.DefaultQuery("include_variables", "true") == "true",
		IncludeMetadata:   c.Query("include_metadata") == "true",
	}

	logger.Debug("Tracing process execution with typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Bool("include_variables", req.IncludeVariables),
		logger.Bool("include_metadata", req.IncludeMetadata))

	// Get typed process component
	processComp := h.coreInterface.GetProcessComponentTyped()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Typed process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get execution trace using typed method
	result, err := processComp.TraceProcessExecution(req)
	if err != nil {
		logger.Error("Failed to trace process execution via typed API",
			logger.String("request_id", requestID),
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Traced process execution via typed API",
		logger.String("request_id", requestID),
		logger.String("instance_id", instanceID),
		logger.Int("total_tokens", int(result.TotalTokens)))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}

// GetProcessStatsHandler handles GET /api/v1/processes/stats
// @Summary Get process statistics with typed response
// @Description Get comprehensive process statistics using strongly typed API
// @Tags processes
// @Produce json
// @Success 200 {object} restmodels.APIResponse{data=types.ProcessStats}
// @Failure 500 {object} restmodels.APIResponse{error=restmodels.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/processes/stats [get]
func (h *ProcessHandler) GetProcessStatsHandler(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting process statistics with typed API",
		logger.String("request_id", requestID))

	// Get typed process component
	processComp := h.coreInterface.GetProcessComponentTyped()
	if processComp == nil {
		apiErr := restmodels.InternalServerError("Typed process service not available")
		c.JSON(http.StatusInternalServerError, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	// Get process statistics using typed method
	result, err := processComp.GetProcessStats()
	if err != nil {
		logger.Error("Failed to get process statistics via typed API",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := restmodels.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, restmodels.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Got process statistics via typed API",
		logger.String("request_id", requestID),
		logger.Int64("total_instances", result.TotalInstances))

	c.JSON(http.StatusOK, restmodels.SuccessResponse(result, requestID))
}
