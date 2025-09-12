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

// JobsHandler handles job management HTTP requests
type JobsHandler struct {
	coreInterface JobsCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// JobsCoreInterface defines methods needed for jobs operations
type JobsCoreInterface interface {
	// JSON Message Routing to jobs component
	SendMessage(componentName, messageJSON string) error
	WaitForJobsResponse(timeoutMs int) (string, error)
	GetJobsComponent() interface{}
}

// Job data types
type Job struct {
	Key                 string                 `json:"key"`
	Type                string                 `json:"type"`
	ProcessInstanceID   string                 `json:"process_instance_id"`
	ProcessDefinitionID string                 `json:"process_definition_id"`
	ElementID           string                 `json:"element_id"`
	ElementInstanceID   string                 `json:"element_instance_id"`
	CustomHeaders       map[string]string      `json:"custom_headers"`
	Variables           map[string]interface{} `json:"variables"`
	Retries             int32                  `json:"retries"`
	Deadline            int64                  `json:"deadline"`
	Worker              string                 `json:"worker,omitempty"`
	State               string                 `json:"state"`
	CreatedAt           int64                  `json:"created_at"`
	UpdatedAt           int64                  `json:"updated_at"`
}

type JobActivationResponse struct {
	Jobs []Job `json:"jobs"`
}

type JobStats struct {
	TotalJobs        int64            `json:"total_jobs"`
	ActiveJobs       int64            `json:"active_jobs"`
	CompletedJobs    int64            `json:"completed_jobs"`
	FailedJobs       int64            `json:"failed_jobs"`
	JobsByType       map[string]int64 `json:"jobs_by_type"`
	JobsByWorker     map[string]int64 `json:"jobs_by_worker"`
	AverageLatency   float64          `json:"average_latency_ms"`
	ThroughputPerMin int64            `json:"throughput_per_minute"`
}

// NewJobsHandler creates new jobs handler
func NewJobsHandler(coreInterface JobsCoreInterface) *JobsHandler {
	return &JobsHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers job routes
func (h *JobsHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	jobs := router.Group("/jobs")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		jobs.Use(authMiddleware.RequirePermission("job"))
	}

	{
		jobs.POST("", h.CreateJob)
		jobs.GET("", h.ListJobs)
		jobs.GET("/:key", h.GetJob)
		jobs.POST("/activate", h.ActivateJobs)
		jobs.PUT("/:key/complete", h.CompleteJob)
		jobs.PUT("/:key/fail", h.FailJob)
		jobs.POST("/:key/throw-error", h.ThrowError)
		jobs.PUT("/:key/retries", h.UpdateJobRetries)
		jobs.DELETE("/:key", h.CancelJob)
		jobs.PUT("/:key/timeout", h.UpdateJobTimeout)
		jobs.GET("/stats", h.GetJobStats)
	}
}

// CreateJob handles POST /api/v1/jobs
// @Summary Create job
// @Description Create a new job for service task execution
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body models.CreateJobRequest true "Job creation request"
// @Success 201 {object} models.APIResponse{data=models.CreateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs [post]
func (h *JobsHandler) CreateJob(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse create job request",
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
			return h.validator.ValidateRequired(req.ProcessInstanceID, "process_instance_id")
		},
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.ElementID, "element_id")
		},
		func() *models.ValidationError {
			return h.validator.ValidateRange(req.Retries, "retries", 0, 100)
		},
		func() *models.ValidationError {
			return h.validator.ValidateRange(req.TimeoutMs, "timeout_ms", 0, 86400000) // 24 hours max
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Creating job",
		logger.String("request_id", requestID),
		logger.String("type", req.Type),
		logger.String("process_instance_id", req.ProcessInstanceID),
		logger.String("element_id", req.ElementID))

	// Create job request message
	jobReq := map[string]interface{}{
		"type":       "create_job",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_type":            req.Type,
			"process_instance_id": req.ProcessInstanceID,
			"element_id":          req.ElementID,
			"element_instance_id": req.ElementInstanceID,
			"custom_headers":      req.CustomHeaders,
			"variables":           req.Variables,
			"retries":             req.Retries,
			"timeout_ms":          req.TimeoutMs,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(jobReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Extract job key from response
	var jobKey string
	if result, exists := response["result"]; exists {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if jid, ok := resultMap["job_id"].(string); ok {
				jobKey = jid
			}
		}
	}

	if jobKey == "" {
		apiErr := models.InternalServerError("Job created but key not returned")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	createResp := &models.CreateResponse{
		ID:      jobKey,
		Message: "Job created successfully",
	}

	logger.Info("Job created successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.String("type", req.Type))

	c.JSON(http.StatusCreated, models.SuccessResponse(createResp, requestID))
}

// ActivateJobs handles POST /api/v1/jobs/activate
// @Summary Activate jobs for worker
// @Description Activate available jobs for a specific worker
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body models.ActivateJobsRequest true "Job activation request"
// @Success 200 {object} models.APIResponse{data=JobActivationResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/activate [post]
func (h *JobsHandler) ActivateJobs(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.ActivateJobsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
			return h.validator.ValidateRequired(req.Worker, "worker")
		},
		func() *models.ValidationError {
			if req.MaxJobs <= 0 {
				req.MaxJobs = 10 // Default value
			}
			return h.validator.ValidateRange(req.MaxJobs, "max_jobs", 1, 1000)
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Activating jobs for worker",
		logger.String("request_id", requestID),
		logger.String("type", req.Type),
		logger.String("worker", req.Worker),
		logger.Any("max_jobs", req.MaxJobs))

	// Create activation request
	activateReq := map[string]interface{}{
		"type":       "activate_jobs",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_type":    req.Type,
			"worker_name": req.Worker,
			"max_jobs":    req.MaxJobs,
			"timeout_ms":  req.TimeoutMs,
		},
	}

	// Send to jobs component and get response
	response, err := h.sendJobsRequest(activateReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse jobs from response
	jobs := h.parseJobsFromResponse(response)

	activationResp := &JobActivationResponse{
		Jobs: jobs,
	}

	logger.Info("Jobs activated for worker",
		logger.String("request_id", requestID),
		logger.String("worker", req.Worker),
		logger.Int("activated_count", len(jobs)))

	c.JSON(http.StatusOK, models.SuccessResponse(activationResp, requestID))
}

// ListJobs handles GET /api/v1/jobs
// @Summary List jobs
// @Description Get list of jobs with filtering and pagination
// @Tags jobs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Job type filter"
// @Param worker query string false "Worker filter"
// @Param state query string false "State filter (activatable, activated, completed, failed)"
// @Success 200 {object} models.PaginatedResponse{data=[]Job}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs [get]
func (h *JobsHandler) ListJobs(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	jobType := c.Query("type")
	worker := c.Query("worker")
	state := c.Query("state")

	// Parse and validate pagination
	paginationHelper := utils.NewPaginationHelper()
	params, apiErr := paginationHelper.ParseAndValidate(pageStr, limitStr)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate state filter
	if state != "" {
		validStates := []string{"activatable", "activated", "completed", "failed", "cancelled"}
		if apiErr := h.validator.ValidateStringEnum(state, "state", validStates); apiErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.NewValidationError("Invalid state filter", []models.ValidationError{*apiErr}),
				requestID))
			return
		}
	}

	logger.Debug("Listing jobs",
		logger.String("request_id", requestID),
		logger.Int("page", params.Page),
		logger.Int("limit", params.Limit),
		logger.String("type", jobType),
		logger.String("worker", worker),
		logger.String("state", state))

	// Create list request (load all for sorting)
	listReq := map[string]interface{}{
		"type":       "list_jobs",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_type": jobType,
			"worker":   worker,
			"state":    state,
			"limit":    0, // Load all for sorting
			"offset":   0,
		},
	}

	// Send to jobs component and get response
	response, err := h.sendJobsRequest(listReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse jobs from response
	jobs := h.parseJobsFromResponse(response)
	totalCount := len(jobs)

	// Apply sorting by created_at DESC (consistent with gRPC/CLI behavior)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt > jobs[j].CreatedAt // DESC order
	})

	// Apply client-side pagination after sorting
	paginatedJobs, paginationInfo := utils.ApplyPagination(jobs, params.Page, params.Limit)

	logger.Info("Jobs listed",
		logger.String("request_id", requestID),
		logger.Int("count", len(jobs)),
		logger.Int("total", totalCount))

	paginatedResp := models.PaginatedSuccessResponse(paginatedJobs, paginationInfo, requestID)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetJob handles GET /api/v1/jobs/:key
// @Summary Get job details
// @Description Get detailed information about a specific job
// @Tags jobs
// @Produce json
// @Param key path string true "Job key"
// @Success 200 {object} models.APIResponse{data=Job}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key} [get]
func (h *JobsHandler) GetJob(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	if jobKey == "" {
		apiErr := models.BadRequestError("Job key is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting job details",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	// Create get request
	getReq := map[string]interface{}{
		"type":       "get_job",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_id": jobKey,
		},
	}

	// Send to jobs component and get response
	response, err := h.sendJobsRequest(getReq, requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apiErr := models.JobNotFoundError(jobKey)
			c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		} else {
			apiErr := h.converter.GRPCErrorToAPIError(err)
			statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
			c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		}
		return
	}

	// Parse job from response
	job := h.parseJobFromResponse(response)
	if job == nil {
		apiErr := models.JobNotFoundError(jobKey)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Job details retrieved",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.String("type", job.Type),
		logger.String("state", job.State))

	c.JSON(http.StatusOK, models.SuccessResponse(job, requestID))
}

// CompleteJob handles PUT /api/v1/jobs/:key/complete
// @Summary Complete job
// @Description Mark job as completed with optional variables
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.CompleteJobRequest false "Job completion request"
// @Success 200 {object} models.APIResponse{data=models.UpdateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key}/complete [put]
func (h *JobsHandler) CompleteJob(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	if jobKey == "" {
		apiErr := models.BadRequestError("Job key is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse optional request body
	var req models.CompleteJobRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("Failed to parse complete job request body, using defaults",
				logger.String("request_id", requestID),
				logger.String("error", err.Error()))
		}
	}

	logger.Debug("Completing job",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	// Create complete request
	completeReq := map[string]interface{}{
		"type":       "complete_job",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key":   jobKey,
			"variables": req.Variables,
		},
	}

	// Send to jobs component and get response
	_, err := h.sendJobsRequest(completeReq, requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			apiErr := models.JobNotFoundError(jobKey)
			c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		} else {
			apiErr := h.converter.GRPCErrorToAPIError(err)
			statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
			c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		}
		return
	}

	updateResp := &models.UpdateResponse{
		Message: "Job completed successfully",
	}

	logger.Info("Job completed successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	c.JSON(http.StatusOK, models.SuccessResponse(updateResp, requestID))
}

// FailJob handles PUT /api/v1/jobs/:key/fail
// @Summary Fail job
// @Description Mark a job as failed with retry information
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.FailJobRequest true "Job failure request"
// @Success 200 {object} models.APIResponse{data=models.SuccessResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key}/fail [put]
func (h *JobsHandler) FailJob(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	// Parse request body
	var req models.FailJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse fail job request",
			logger.String("request_id", requestID),
			logger.String("job_key", jobKey),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(models.BadRequestError(err.Error()), requestID))
		}
		return
	}

	logger.Debug("Failing job",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.Int("retries", int(req.Retries)))

	// Create fail job request
	failReq := map[string]interface{}{
		"type":       "fail_job",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key":       jobKey,
			"retries":       req.Retries,
			"error_message": req.ErrorMessage,
			"backoff_ms":    req.BackoffMs,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(failReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if success, ok := response["success"].(bool); !ok || !success {
		message := "Job failure operation failed"
		if msg, exists := response["message"].(string); exists {
			message = msg
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Job failed successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	successResp := &models.UpdateResponse{
		ID:      jobKey,
		Message: "Job failed successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(successResp, requestID))
}

// ThrowError handles POST /api/v1/jobs/:key/throw-error
// @Summary Throw BPMN error for job
// @Description Throw a BPMN error for a job that will trigger error handling
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.ThrowErrorRequest true "Error throwing request"
// @Success 200 {object} models.APIResponse{data=models.SuccessResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key}/throw-error [post]
func (h *JobsHandler) ThrowError(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	// Parse request body
	var req models.ThrowErrorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse throw error request",
			logger.String("request_id", requestID),
			logger.String("job_key", jobKey),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(models.BadRequestError(err.Error()), requestID))
		}
		return
	}

	logger.Debug("Throwing error for job",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.String("error_code", req.ErrorCode))

	// Create throw error request
	throwReq := map[string]interface{}{
		"type":       "throw_error",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key":       jobKey,
			"error_code":    req.ErrorCode,
			"error_message": req.ErrorMessage,
			"variables":     req.Variables,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(throwReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if success, ok := response["success"].(bool); !ok || !success {
		message := "Error throwing operation failed"
		if msg, exists := response["message"].(string); exists {
			message = msg
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Error thrown for job successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.String("error_code", req.ErrorCode))

	successResp := &models.UpdateResponse{
		ID:      jobKey,
		Message: "BPMN error thrown successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(successResp, requestID))
}

// UpdateJobRetries handles PUT /api/v1/jobs/:key/retries
// @Summary Update job retries
// @Description Update the number of retries for a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.UpdateJobRetriesRequest true "Job retries update request"
// @Success 200 {object} models.APIResponse{data=models.SuccessResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key}/retries [put]
func (h *JobsHandler) UpdateJobRetries(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	// Parse request body
	var req models.UpdateJobRetriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse update retries request",
			logger.String("request_id", requestID),
			logger.String("job_key", jobKey),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(models.BadRequestError(err.Error()), requestID))
		}
		return
	}

	logger.Debug("Updating job retries",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.Int("retries", int(req.Retries)))

	// Create update retries request
	updateReq := map[string]interface{}{
		"type":       "update_job_retries",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key": jobKey,
			"retries": req.Retries,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(updateReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if success, ok := response["success"].(bool); !ok || !success {
		message := "Job retries update failed"
		if msg, exists := response["message"].(string); exists {
			message = msg
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Job retries updated successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	successResp := &models.UpdateResponse{
		ID:      jobKey,
		Message: "Job retries updated successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(successResp, requestID))
}

// CancelJob handles DELETE /api/v1/jobs/:key
// @Summary Cancel job
// @Description Cancel a job with optional reason
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.CancelJobRequest false "Job cancellation request"
// @Success 200 {object} models.APIResponse{data=models.SuccessResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key} [delete]
func (h *JobsHandler) CancelJob(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	// Parse optional request body
	var req models.CancelJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Body is optional for DELETE, ignore bind errors
		req = models.CancelJobRequest{}
	}

	logger.Debug("Cancelling job",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.String("reason", req.Reason))

	// Create cancel job request
	cancelReq := map[string]interface{}{
		"type":       "cancel_job",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key": jobKey,
			"reason":  req.Reason,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(cancelReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if success, ok := response["success"].(bool); !ok || !success {
		message := "Job cancellation failed"
		if msg, exists := response["message"].(string); exists {
			message = msg
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Job cancelled successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	successResp := &models.DeleteResponse{
		ID:      jobKey,
		Message: "Job cancelled successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(successResp, requestID))
}

// UpdateJobTimeout handles PUT /api/v1/jobs/:key/timeout
// @Summary Update job timeout
// @Description Update the timeout for a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param key path string true "Job key"
// @Param request body models.UpdateJobTimeoutRequest true "Job timeout update request"
// @Success 200 {object} models.APIResponse{data=models.SuccessResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/{key}/timeout [put]
func (h *JobsHandler) UpdateJobTimeout(c *gin.Context) {
	requestID := h.getRequestID(c)
	jobKey := c.Param("key")

	// Parse request body
	var req models.UpdateJobTimeoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse update timeout request",
			logger.String("request_id", requestID),
			logger.String("job_key", jobKey),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if apiErr, ok := err.(*models.APIError); ok {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(models.BadRequestError(err.Error()), requestID))
		}
		return
	}

	logger.Debug("Updating job timeout",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey),
		logger.Int64("timeout_ms", req.TimeoutMs))

	// Create update timeout request
	updateReq := map[string]interface{}{
		"type":       "update_job_timeout",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"job_key":    jobKey,
			"timeout_ms": req.TimeoutMs,
		},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(updateReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if success, ok := response["success"].(bool); !ok || !success {
		message := "Job timeout update failed"
		if msg, exists := response["message"].(string); exists {
			message = msg
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Job timeout updated successfully",
		logger.String("request_id", requestID),
		logger.String("job_key", jobKey))

	successResp := &models.UpdateResponse{
		ID:      jobKey,
		Message: "Job timeout updated successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(successResp, requestID))
}

// GetJobStats handles GET /api/v1/jobs/stats
// @Summary Get job statistics
// @Description Get comprehensive job statistics
// @Tags jobs
// @Produce json
// @Success 200 {object} models.APIResponse{data=JobStats}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/jobs/stats [get]
func (h *JobsHandler) GetJobStats(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting job statistics",
		logger.String("request_id", requestID))

	// Create get stats request
	statsReq := map[string]interface{}{
		"type":       "get_stats",
		"request_id": requestID,
		"payload":    map[string]interface{}{},
	}

	// Send to jobs component
	response, err := h.sendJobsRequest(statsReq, requestID)
	if err != nil {
		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse stats from response
	stats := &JobStats{
		TotalJobs:        0,
		ActiveJobs:       0,
		CompletedJobs:    0,
		FailedJobs:       0,
		JobsByType:       make(map[string]int64),
		JobsByWorker:     make(map[string]int64),
		AverageLatency:   0.0,
		ThroughputPerMin: 0,
	}

	// Extract stats from response
	if statsData, exists := response["stats"]; exists {
		if statsMap, ok := statsData.(map[string]interface{}); ok {
			if totalJobs, ok := statsMap["total_jobs"].(float64); ok {
				stats.TotalJobs = int64(totalJobs)
			}
			if activeJobs, ok := statsMap["active_jobs"].(float64); ok {
				stats.ActiveJobs = int64(activeJobs)
			}
			if completedJobs, ok := statsMap["completed_jobs"].(float64); ok {
				stats.CompletedJobs = int64(completedJobs)
			}
			if failedJobs, ok := statsMap["failed_jobs"].(float64); ok {
				stats.FailedJobs = int64(failedJobs)
			}
		}
	}

	logger.Info("Job statistics retrieved",
		logger.String("request_id", requestID),
		logger.Int64("total_jobs", stats.TotalJobs))

	c.JSON(http.StatusOK, models.SuccessResponse(stats, requestID))
}

// Helper methods

func (h *JobsHandler) sendJobsRequest(req map[string]interface{}, requestID string) (map[string]interface{}, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	err = h.coreInterface.SendMessage("jobs", string(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	respJSON, err := h.coreInterface.WaitForJobsResponse(30000)
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

func (h *JobsHandler) parseJobsFromResponse(response map[string]interface{}) []Job {
	var jobs []Job

	// Extract result from response
	resultData, exists := response["result"]
	if !exists {
		return jobs
	}

	resultMap, ok := resultData.(map[string]interface{})
	if !ok {
		return jobs
	}

	// Extract jobs array from result
	jobsData, exists := resultMap["jobs"]
	if !exists {
		return jobs
	}

	jobsArray, ok := jobsData.([]interface{})
	if !ok {
		return jobs
	}

	// Parse each job
	for _, jobData := range jobsArray {
		jobMap, ok := jobData.(map[string]interface{})
		if !ok {
			continue
		}

		job := h.parseJobFromMap(jobMap)
		if job != nil {
			jobs = append(jobs, *job)
		}
	}

	return jobs
}

func (h *JobsHandler) parseJobFromResponse(response map[string]interface{}) *Job {
	// Extract result from response
	resultData, exists := response["result"]
	if !exists {
		return nil
	}

	jobMap, ok := resultData.(map[string]interface{})
	if !ok {
		return nil
	}

	return h.parseJobFromMap(jobMap)
}

func (h *JobsHandler) parseJobFromMap(jobMap map[string]interface{}) *Job {
	job := &Job{}

	// Parse string fields
	if key, ok := jobMap["key"].(string); ok {
		job.Key = key
	}
	if jobType, ok := jobMap["type"].(string); ok {
		job.Type = jobType
	}
	if processInstanceID, ok := jobMap["process_instance_id"].(string); ok {
		job.ProcessInstanceID = processInstanceID
	}
	if worker, ok := jobMap["worker"].(string); ok {
		job.Worker = worker
	}
	if status, ok := jobMap["status"].(string); ok {
		job.State = status
	}

	// Parse numeric fields
	if retries, ok := jobMap["retries"].(float64); ok {
		job.Retries = int32(retries)
	}
	if createdAt, ok := jobMap["created_at"].(float64); ok {
		job.CreatedAt = int64(createdAt)
	}

	// Parse variables
	if variables, ok := jobMap["variables"].(map[string]interface{}); ok {
		job.Variables = variables
	}

	// Initialize empty maps if nil
	if job.CustomHeaders == nil {
		job.CustomHeaders = make(map[string]string)
	}
	if job.Variables == nil {
		job.Variables = make(map[string]interface{})
	}

	return job
}

func (h *JobsHandler) extractTotalCount(response map[string]interface{}) int {
	// Extract result from response
	resultData, exists := response["result"]
	if !exists {
		return 0
	}

	resultMap, ok := resultData.(map[string]interface{})
	if !ok {
		return 0
	}

	// Extract total from result
	if total, ok := resultMap["total"].(float64); ok {
		return int(total)
	}

	return 0
}

func (h *JobsHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("jobs")
}
