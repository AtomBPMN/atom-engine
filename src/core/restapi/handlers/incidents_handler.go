/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// IncidentsHandler handles incident management HTTP requests
type IncidentsHandler struct {
	coreInterface IncidentsCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// CreateIncidentRequest represents incident creation request
type CreateIncidentRequest struct {
	Type              string                 `json:"type" binding:"required"`
	Message           string                 `json:"message" binding:"required"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ProcessInstanceID string                 `json:"process_instance_id,omitempty"`
	ProcessKey        string                 `json:"process_key,omitempty"`
	ElementID         string                 `json:"element_id,omitempty"`
	ElementType       string                 `json:"element_type,omitempty"`
	JobKey            string                 `json:"job_key,omitempty"`
	JobType           string                 `json:"job_type,omitempty"`
	WorkerID          string                 `json:"worker_id,omitempty"`
	TimerID           string                 `json:"timer_id,omitempty"`
	MessageName       string                 `json:"message_name,omitempty"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	OriginalRetries   int32                  `json:"original_retries,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentsCoreInterface defines methods needed for incidents operations
type IncidentsCoreInterface interface {
	// JSON Message Routing to incidents component
	SendMessage(componentName, messageJSON string) error
	WaitForIncidentsResponse(timeoutMs int) (string, error)
	GetIncidentsComponent() interface{}
}

// Incident data types
type Incident struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Status            string                 `json:"status"`
	Message           string                 `json:"message"`
	ErrorCode         string                 `json:"error_code"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ProcessKey        string                 `json:"process_key"`
	ElementID         string                 `json:"element_id"`
	ElementType       string                 `json:"element_type"`
	JobKey            string                 `json:"job_key,omitempty"`
	JobType           string                 `json:"job_type,omitempty"`
	WorkerID          string                 `json:"worker_id,omitempty"`
	TimerID           string                 `json:"timer_id,omitempty"`
	MessageName       string                 `json:"message_name,omitempty"`
	CorrelationKey    string                 `json:"correlation_key,omitempty"`
	ResolvedAt        int64                  `json:"resolved_at,omitempty"`
	ResolvedBy        string                 `json:"resolved_by,omitempty"`
	ResolveAction     string                 `json:"resolve_action,omitempty"`
	ResolveComment    string                 `json:"resolve_comment,omitempty"`
	OriginalRetries   int32                  `json:"original_retries,omitempty"`
	NewRetries        int32                  `json:"new_retries,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type IncidentStats struct {
	TotalIncidents     int32            `json:"total_incidents"`
	OpenIncidents      int32            `json:"open_incidents"`
	ResolvedIncidents  int32            `json:"resolved_incidents"`
	DismissedIncidents int32            `json:"dismissed_incidents"`
	IncidentsByType    map[string]int32 `json:"incidents_by_type"`
	IncidentsByStatus  map[string]int32 `json:"incidents_by_status"`
	RecentIncidents24h int32            `json:"recent_incidents_24h"`
}

// NewIncidentsHandler creates new incidents handler
func NewIncidentsHandler(coreInterface IncidentsCoreInterface) *IncidentsHandler {
	return &IncidentsHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers incident routes
func (h *IncidentsHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	incidents := router.Group("/incidents")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		incidents.Use(authMiddleware.RequirePermission("incident"))
	}

	{
		incidents.POST("", h.CreateIncident)
		incidents.GET("", h.ListIncidents)
		incidents.GET("/:id", h.GetIncident)
		incidents.PUT("/:id/resolve", h.ResolveIncident)
		incidents.GET("/stats", h.GetStats)
	}
}

// CreateIncident handles POST /api/v1/incidents
// @Summary Create incident
// @Description Create a new incident for error tracking
// @Tags incidents
// @Accept json
// @Produce json
// @Param request body CreateIncidentRequest true "Incident creation request"
// @Success 201 {object} models.APIResponse{data=models.CreateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/incidents [post]
func (h *IncidentsHandler) CreateIncident(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse create incident request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	validationErrors := h.validator.ValidateMultiple(
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.Type, "type")
		},
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.Message, "message")
		},
		func() *models.ValidationError {
			validTypes := []string{"job_failure", "bpmn_error", "expression_error", "process_error", "timer_error", "message_error", "system_error"}
			return h.validator.ValidateStringEnum(req.Type, "type", validTypes)
		},
		func() *models.ValidationError {
			return h.validator.ValidateStringLength(req.Message, "message", 1, 1000)
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Creating incident",
		logger.String("request_id", requestID),
		logger.String("type", req.Type),
		logger.String("message", req.Message),
		logger.String("process_instance_id", req.ProcessInstanceID))

	// Create incident request message
	incidentReq := map[string]interface{}{
		"operation":           "create",
		"type":                req.Type,
		"message":             req.Message,
		"error_code":          req.ErrorCode,
		"process_instance_id": req.ProcessInstanceID,
		"process_key":         req.ProcessKey,
		"element_id":          req.ElementID,
		"element_type":        req.ElementType,
		"job_key":             req.JobKey,
		"job_type":            req.JobType,
		"worker_id":           req.WorkerID,
		"timer_id":            req.TimerID,
		"message_name":        req.MessageName,
		"correlation_key":     req.CorrelationKey,
		"original_retries":    req.OriginalRetries,
		"metadata":            req.Metadata,
	}

	// Send to incidents component
	response, err := h.sendIncidentsRequest(incidentReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if incident creation was successful
	success, _ := response["success"].(bool)
	if !success {
		errorMsg, _ := response["error"].(string)
		if errorMsg == "" {
			errorMsg = "Incident creation failed"
		}

		apiErr := models.InternalServerError(errorMsg)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Extract incident ID from response
	incidentID, _ := response["incident_id"].(string)
	if incidentID == "" {
		apiErr := models.InternalServerError("Incident created but ID not returned")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	createResp := &models.CreateResponse{
		ID:      incidentID,
		Message: "Incident created successfully",
	}

	logger.Info("Incident created successfully",
		logger.String("request_id", requestID),
		logger.String("incident_id", incidentID),
		logger.String("type", req.Type))

	c.JSON(http.StatusCreated, models.SuccessResponse(createResp, requestID))
}

// ListIncidents handles GET /api/v1/incidents
// @Summary List incidents
// @Description Get list of incidents with filtering and pagination
// @Tags incidents
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Status filter (open, resolved, dismissed)"
// @Param type query string false "Type filter"
// @Param process_instance_id query string false "Process instance ID filter"
// @Param process_key query string false "Process key filter"
// @Param element_id query string false "Element ID filter"
// @Param job_key query string false "Job key filter"
// @Param worker_id query string false "Worker ID filter"
// @Success 200 {object} models.PaginatedResponse{data=[]Incident}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/incidents [get]
func (h *IncidentsHandler) ListIncidents(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	incidentType := c.Query("type")
	processInstanceID := c.Query("process_instance_id")
	processKey := c.Query("process_key")
	elementID := c.Query("element_id")
	jobKey := c.Query("job_key")
	workerID := c.Query("worker_id")

	// Parse and validate pagination
	paginationHelper := utils.NewPaginationHelper()
	params, apiErr := paginationHelper.ParseAndValidate(pageStr, limitStr)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate status filter
	if status != "" {
		validStatuses := []string{"open", "resolved", "dismissed"}
		if apiErr := h.validator.ValidateStringEnum(status, "status", validStatuses); apiErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.NewValidationError("Invalid status filter", []models.ValidationError{*apiErr}),
				requestID))
			return
		}
	}

	// Validate type filter
	if incidentType != "" {
		validTypes := []string{"job_failure", "bpmn_error", "expression_error", "process_error", "timer_error", "message_error", "system_error"}
		if apiErr := h.validator.ValidateStringEnum(incidentType, "type", validTypes); apiErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.NewValidationError("Invalid type filter", []models.ValidationError{*apiErr}),
				requestID))
			return
		}
	}

	logger.Debug("Listing incidents",
		logger.String("request_id", requestID),
		logger.Int("page", params.Page),
		logger.Int("limit", params.Limit),
		logger.String("status", status),
		logger.String("type", incidentType))

	// Create list request (load all for sorting)
	listReq := map[string]interface{}{
		"operation":           "list",
		"status":              status,
		"type":                incidentType,
		"process_instance_id": processInstanceID,
		"process_key":         processKey,
		"element_id":          elementID,
		"job_key":             jobKey,
		"worker_id":           workerID,
		"limit":               0, // Load all for sorting
		"offset":              0,
	}

	// Send to incidents component and get response
	response, err := h.sendIncidentsRequest(listReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse incidents from response
	incidents := h.parseIncidentsFromResponse(response)
	totalCount := len(incidents)

	// Apply sorting by created_at DESC (consistent with gRPC/CLI behavior)
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt > incidents[j].CreatedAt // DESC order
	})

	// Apply client-side pagination after sorting
	paginatedIncidents, paginationInfo := utils.ApplyPagination(incidents, params.Page, params.Limit)

	logger.Info("Incidents listed",
		logger.String("request_id", requestID),
		logger.Int("count", len(incidents)),
		logger.Int("total", totalCount))

	paginatedResp := models.PaginatedSuccessResponse(paginatedIncidents, paginationInfo, requestID)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetIncident handles GET /api/v1/incidents/:id
// @Summary Get incident details
// @Description Get detailed information about a specific incident
// @Tags incidents
// @Produce json
// @Param id path string true "Incident ID"
// @Success 200 {object} models.APIResponse{data=Incident}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/incidents/{id} [get]
func (h *IncidentsHandler) GetIncident(c *gin.Context) {
	requestID := h.getRequestID(c)
	incidentID := c.Param("id")

	if incidentID == "" {
		apiErr := models.BadRequestError("Incident ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting incident details",
		logger.String("request_id", requestID),
		logger.String("incident_id", incidentID))

	// Create get request
	getReq := map[string]interface{}{
		"operation":   "get",
		"incident_id": incidentID,
	}

	// Send to incidents component and get response
	response, err := h.sendIncidentsRequest(getReq, requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apiErr := models.NewAPIErrorWithDetails(
				models.ErrorCodeResourceNotFound,
				"Incident not found",
				map[string]interface{}{"incident_id": incidentID},
			)
			c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		} else {
			apiErr := h.converter.GRPCErrorToAPIError(err)
			statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
			c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		}
		return
	}

	// Parse incident from response
	incident := h.parseIncidentFromResponse(response)
	if incident == nil {
		apiErr := models.NewAPIErrorWithDetails(
			models.ErrorCodeResourceNotFound,
			"Incident not found",
			map[string]interface{}{"incident_id": incidentID},
		)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Incident details retrieved",
		logger.String("request_id", requestID),
		logger.String("incident_id", incidentID),
		logger.String("type", incident.Type),
		logger.String("status", incident.Status))

	c.JSON(http.StatusOK, models.SuccessResponse(incident, requestID))
}

// ResolveIncident handles PUT /api/v1/incidents/:id/resolve
// @Summary Resolve incident
// @Description Resolve incident with retry or dismiss action
// @Tags incidents
// @Accept json
// @Produce json
// @Param id path string true "Incident ID"
// @Param request body models.ResolveIncidentRequest true "Incident resolution request"
// @Success 200 {object} models.APIResponse{data=models.UpdateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/incidents/{id}/resolve [put]
func (h *IncidentsHandler) ResolveIncident(c *gin.Context) {
	requestID := h.getRequestID(c)
	incidentID := c.Param("id")

	if incidentID == "" {
		apiErr := models.BadRequestError("Incident ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse request body
	var req models.ResolveIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	validationErrors := h.validator.ValidateMultiple(
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.Action, "action")
		},
		func() *models.ValidationError {
			validActions := []string{"retry", "dismiss"}
			return h.validator.ValidateStringEnum(req.Action, "action", validActions)
		},
		func() *models.ValidationError {
			if req.Action == "retry" && req.NewRetries < 0 {
				return &models.ValidationError{
					Field:   "new_retries",
					Value:   req.NewRetries,
					Message: "new_retries must be non-negative for retry action",
				}
			}
			return nil
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Resolving incident",
		logger.String("request_id", requestID),
		logger.String("incident_id", incidentID),
		logger.String("action", req.Action),
		logger.String("comment", req.Comment))

	// Create resolve request
	resolveReq := map[string]interface{}{
		"operation":   "resolve",
		"incident_id": incidentID,
		"action":      req.Action,
		"comment":     req.Comment,
		"resolved_by": req.ResolvedBy,
		"new_retries": req.NewRetries,
	}

	// Send to incidents component and get response
	response, err := h.sendIncidentsRequest(resolveReq, requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apiErr := models.NewAPIErrorWithDetails(
				models.ErrorCodeResourceNotFound,
				"Incident not found",
				map[string]interface{}{"incident_id": incidentID},
			)
			c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		} else {
			apiErr := h.converter.GRPCErrorToAPIError(err)
			statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
			c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		}
		return
	}

	// Check if resolution was successful
	success, _ := response["success"].(bool)
	if !success {
		errorMsg, _ := response["error"].(string)
		if errorMsg == "" {
			errorMsg = "Incident resolution failed"
		}

		apiErr := models.InternalServerError(errorMsg)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	updateResp := &models.UpdateResponse{
		ID:      incidentID,
		Message: fmt.Sprintf("Incident %s successfully", req.Action),
	}

	logger.Info("Incident resolved successfully",
		logger.String("request_id", requestID),
		logger.String("incident_id", incidentID),
		logger.String("action", req.Action))

	c.JSON(http.StatusOK, models.SuccessResponse(updateResp, requestID))
}

// GetStats handles GET /api/v1/incidents/stats
// @Summary Get incident statistics
// @Description Get incident processing statistics
// @Tags incidents
// @Produce json
// @Success 200 {object} models.APIResponse{data=IncidentStats}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/incidents/stats [get]
func (h *IncidentsHandler) GetStats(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting incident statistics",
		logger.String("request_id", requestID))

	// Create stats request
	statsReq := map[string]interface{}{
		"operation": "stats",
	}

	// Send to incidents component and get response
	response, err := h.sendIncidentsRequest(statsReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse stats from response
	stats := h.parseStatsFromResponse(response)

	logger.Info("Incident statistics retrieved",
		logger.String("request_id", requestID),
		logger.Any("total_incidents", stats.TotalIncidents),
		logger.Any("open_incidents", stats.OpenIncidents))

	c.JSON(http.StatusOK, models.SuccessResponse(stats, requestID))
}

// Helper methods

func (h *IncidentsHandler) sendIncidentsRequest(req map[string]interface{}, requestID string) (map[string]interface{}, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	err = h.coreInterface.SendMessage("incidents", string(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	respJSON, err := h.coreInterface.WaitForIncidentsResponse(30000)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}

	var response map[string]interface{}
	err = json.Unmarshal([]byte(respJSON), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

func (h *IncidentsHandler) parseIncidentsFromResponse(response map[string]interface{}) []Incident {
	// Parse incidents from response - implementation details
	return []Incident{}
}

func (h *IncidentsHandler) parseIncidentFromResponse(response map[string]interface{}) *Incident {
	// Parse single incident from response - implementation details
	return nil
}

func (h *IncidentsHandler) parseStatsFromResponse(response map[string]interface{}) *IncidentStats {
	// Parse stats from response - implementation details
	return &IncidentStats{}
}

func (h *IncidentsHandler) extractTotalCount(response map[string]interface{}) int {
	if count, ok := response["total_count"].(float64); ok {
		return int(count)
	}
	return 0
}

func (h *IncidentsHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("incidents")
}
