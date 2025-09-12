/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/grpc"
	"atom-engine/src/core/logger"
	coremodels "atom-engine/src/core/models"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// TimerRequest represents timewheel timer request for JSON messaging
type TimerRequest struct {
	ElementID         string                          `json:"element_id"`
	TokenID           string                          `json:"token_id"`
	ProcessInstanceID string                          `json:"process_instance_id"`
	TimerType         coremodels.TimerType            `json:"timer_type"`
	ProcessContext    *coremodels.TimerProcessContext `json:"process_context"`
	TimeDate          *string                         `json:"time_date,omitempty"`
	TimeDuration      *string                         `json:"time_duration,omitempty"`
	TimeCycle         *string                         `json:"time_cycle,omitempty"`
	AttachedToRef     *string                         `json:"attached_to_ref,omitempty"`
	CancelActivity    *bool                           `json:"cancel_activity,omitempty"`
}

// TimerHandler handles timer management HTTP requests
type TimerHandler struct {
	coreInterface TimerCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// TimerCoreInterface defines methods needed for timer operations
type TimerCoreInterface interface {
	GetTimewheelComponent() grpc.TimewheelComponentInterface
	GetTimewheelStats() (*timewheelpb.GetTimeWheelStatsResponse, error)
	GetTimersList(statusFilter string, limit int32) (*timewheelpb.ListTimersResponse, error)
}

// TimewheelComponentInterface defines timewheel component interface
type TimewheelComponentInterface interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
	GetResponseChannel() <-chan string
	GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool)
}

// Timer response types
type TimewheelStatsResponse struct {
	TotalTimers     int32            `json:"total_timers"`
	PendingTimers   int32            `json:"pending_timers"`
	FiredTimers     int32            `json:"fired_timers"`
	CancelledTimers int32            `json:"cancelled_timers"`
	CurrentTick     int64            `json:"current_tick"`
	SlotsCount      int32            `json:"slots_count"`
	TimerTypes      map[string]int32 `json:"timer_types"`
}

type TimersListResponse struct {
	Timers     []TimerInfo `json:"timers"`
	TotalCount int32       `json:"total_count"`
}

type TimerInfo struct {
	TimerID           string `json:"timer_id"`
	ElementID         string `json:"element_id"`
	ProcessInstanceID string `json:"process_instance_id"`
	TimerType         string `json:"timer_type"`
	Status            string `json:"status"`
	ScheduledAt       int64  `json:"scheduled_at"`
	CreatedAt         int64  `json:"created_at"`
	TimeDuration      string `json:"time_duration"`
	TimeCycle         string `json:"time_cycle"`
	RemainingSeconds  int64  `json:"remaining_seconds"`
	WheelLevel        int32  `json:"wheel_level"`
}

type TimerCreateResponse struct {
	TimerID     string `json:"timer_id"`
	ScheduledAt int64  `json:"scheduled_at"`
	Status      string `json:"status"`
}

// NewTimerHandler creates new timer handler
func NewTimerHandler(coreInterface TimerCoreInterface) *TimerHandler {
	return &TimerHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers timer routes
func (h *TimerHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	timers := router.Group("/timers")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		timers.Use(authMiddleware.RequirePermission("timer"))
	}

	{
		timers.POST("", h.CreateTimer)
		timers.GET("", h.ListTimers)
		timers.GET("/:id", h.GetTimer)
		timers.DELETE("/:id", h.DeleteTimer)
		timers.GET("/stats", h.GetStats)
	}
}

// CreateTimer handles POST /api/v1/timers
// @Summary Create timer
// @Description Create a new timer with ISO 8601 duration format
// @Tags timers
// @Accept json
// @Produce json
// @Param request body models.AddTimerRequest true "Timer creation request"
// @Success 201 {object} models.APIResponse{data=TimerCreateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 409 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/timers [post]
func (h *TimerHandler) CreateTimer(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.AddTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse create timer request",
			logger.String("request_id", requestID),
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

	// Additional validations
	validationErrors := h.validator.ValidateMultiple(
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.TimerID, "timer_id")
		},
		func() *models.ValidationError {
			return h.validator.ValidateISO8601Duration(req.Duration, "duration")
		},
		func() *models.ValidationError {
			if req.Repeating && req.Interval != "" {
				return h.validator.ValidateISO8601Duration(req.Interval, "interval")
			}
			return nil
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Creating timer",
		logger.String("request_id", requestID),
		logger.String("timer_id", req.TimerID),
		logger.String("duration", req.Duration),
		logger.Bool("repeating", req.Repeating))

	// Get timewheel component
	timewheelComp := h.coreInterface.GetTimewheelComponent()
	if timewheelComp == nil {
		logger.Error("Timewheel component not available",
			logger.String("request_id", requestID))

		apiErr := models.InternalServerError("Timer service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Create timer request message
	timerRequest := TimerRequest{
		ElementID:         req.TimerID,                   // Use TimerID as ElementID for user timers
		TokenID:           "rest-token-" + req.TimerID,   // Generate token ID for user timers
		ProcessInstanceID: "rest-process-" + req.TimerID, // Generate process instance ID for user timers
		TimerType:         coremodels.TimerTypeEvent,     // Use EVENT for user timers
		ProcessContext:    nil,                           // No process context for user timers
	}

	// Handle repeating vs one-time timers
	if req.Repeating && req.Interval != "" {
		timerRequest.TimeCycle = &req.Interval
	} else {
		timerRequest.TimeDuration = &req.Duration
	}

	timerReq := struct {
		Type    string       `json:"type"`
		Request TimerRequest `json:"request"`
	}{
		Type:    "schedule_timer",
		Request: timerRequest,
	}

	reqJSON, err := json.Marshal(timerReq)
	if err != nil {
		logger.Error("Failed to marshal timer request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to process timer request")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Sending timer request to timewheel",
		logger.String("request_id", requestID),
		logger.String("json_message", string(reqJSON)))

	// Send timer creation request
	err = timewheelComp.ProcessMessage(context.Background(), string(reqJSON))
	if err != nil {
		logger.Error("Failed to send timer request",
			logger.String("request_id", requestID),
			logger.String("timer_id", req.TimerID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)

		// Check for specific timer errors
		if strings.Contains(err.Error(), "already exists") {
			apiErr = models.ConflictError("Timer with this ID already exists")
		} else if strings.Contains(err.Error(), "invalid duration") {
			apiErr = models.NewAPIError(models.ErrorCodeInvalidDuration, "Invalid timer duration format")
		}

		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Timer creation successful (no error returned)
	timerResp := &TimerCreateResponse{
		TimerID:     req.TimerID,
		ScheduledAt: 0, // We don't have exact scheduled time from ProcessMessage
		Status:      "scheduled",
	}

	logger.Info("Timer created successfully",
		logger.String("request_id", requestID),
		logger.String("timer_id", req.TimerID))

	c.JSON(http.StatusCreated, models.SuccessResponse(timerResp, requestID))
}

// ListTimers handles GET /api/v1/timers
// @Summary List timers
// @Description Get list of timers with filtering and pagination
// @Tags timers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Status filter (scheduled, fired, cancelled)"
// @Success 200 {object} models.PaginatedResponse{data=[]TimerInfo}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/timers [get]
func (h *TimerHandler) ListTimers(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")

	// Parse and validate pagination
	paginationHelper := utils.NewPaginationHelper()
	params, apiErr := paginationHelper.ParseAndValidate(pageStr, limitStr)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate status filter
	if status != "" {
		validStatuses := []string{"scheduled", "fired", "cancelled"}
		if apiErr := h.validator.ValidateStringEnum(status, "status", validStatuses); apiErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.NewValidationError("Invalid status filter", []models.ValidationError{*apiErr}),
				requestID))
			return
		}
	}

	logger.Debug("Listing timers",
		logger.String("request_id", requestID),
		logger.Int("page", params.Page),
		logger.Int("limit", params.Limit),
		logger.String("status", status))

	// Get timers list from core
	timersResp, err := h.coreInterface.GetTimersList(status, int32(params.Limit*params.Page))
	if err != nil {
		logger.Error("Failed to get timers list",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Apply client-side pagination
	timers := timersResp.Timers
	totalCount := int(timersResp.TotalCount)

	paginatedTimers, paginationInfo := utils.ApplyPagination(timers, params.Page, params.Limit)

	logger.Info("Timers listed",
		logger.String("request_id", requestID),
		logger.Int("count", len(timers)),
		logger.Int("total", totalCount))

	paginatedResp := models.PaginatedSuccessResponse(paginatedTimers, paginationInfo, requestID)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetTimer handles GET /api/v1/timers/:id
// @Summary Get timer details
// @Description Get detailed information about a specific timer
// @Tags timers
// @Produce json
// @Param id path string true "Timer ID"
// @Success 200 {object} models.APIResponse{data=TimerInfo}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/timers/{id} [get]
func (h *TimerHandler) GetTimer(c *gin.Context) {
	requestID := h.getRequestID(c)
	timerID := c.Param("id")

	if timerID == "" {
		apiErr := models.BadRequestError("Timer ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting timer details",
		logger.String("request_id", requestID),
		logger.String("timer_id", timerID))

	// Get timewheel component
	timewheelComp := h.coreInterface.GetTimewheelComponent()
	if timewheelComp == nil {
		apiErr := models.InternalServerError("Timer service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Get timer info
	level, remainingSeconds, found := timewheelComp.GetTimerInfo(timerID)
	if !found {
		logger.Warn("Timer not found",
			logger.String("request_id", requestID),
			logger.String("timer_id", timerID))

		apiErr := models.TimerNotFoundError(timerID)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Create timer info response
	timerInfo := &TimerInfo{
		TimerID:          timerID,
		Status:           "scheduled", // Default status
		RemainingSeconds: remainingSeconds,
		WheelLevel:       int32(level),
	}

	logger.Info("Timer details retrieved",
		logger.String("request_id", requestID),
		logger.String("timer_id", timerID),
		logger.Int64("remaining_seconds", remainingSeconds))

	c.JSON(http.StatusOK, models.SuccessResponse(timerInfo, requestID))
}

// DeleteTimer handles DELETE /api/v1/timers/:id
// @Summary Delete timer
// @Description Cancel and remove a timer
// @Tags timers
// @Produce json
// @Param id path string true "Timer ID"
// @Success 200 {object} models.APIResponse{data=models.DeleteResponse}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/timers/{id} [delete]
func (h *TimerHandler) DeleteTimer(c *gin.Context) {
	requestID := h.getRequestID(c)
	timerID := c.Param("id")

	if timerID == "" {
		apiErr := models.BadRequestError("Timer ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Deleting timer",
		logger.String("request_id", requestID),
		logger.String("timer_id", timerID))

	// Get timewheel component
	timewheelComp := h.coreInterface.GetTimewheelComponent()
	if timewheelComp == nil {
		logger.Error("Timewheel component not available",
			logger.String("request_id", requestID))

		apiErr := models.InternalServerError("Timer service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Create timer removal request message
	timerReq := map[string]interface{}{
		"type":     "cancel_timer",
		"timer_id": timerID,
	}

	reqJSON, err := json.Marshal(timerReq)
	if err != nil {
		logger.Error("Failed to marshal timer removal request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to process timer removal request")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Send timer removal request
	err = timewheelComp.ProcessMessage(context.Background(), string(reqJSON))
	if err != nil {
		logger.Error("Failed to send timer removal request",
			logger.String("request_id", requestID),
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)

		// Check for specific timer errors
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			apiErr = models.TimerNotFoundError(timerID)
		}

		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Timer removal successful (no error returned)
	logger.Info("Timer removed successfully",
		logger.String("request_id", requestID),
		logger.String("timer_id", timerID))

	// Return success response with minimal data
	deleteResp := map[string]interface{}{
		"timer_id": timerID,
		"message":  "Timer removed successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(deleteResp, requestID))
}

// GetStats handles GET /api/v1/timers/stats
// @Summary Get timer statistics
// @Description Get timewheel statistics and metrics
// @Tags timers
// @Produce json
// @Success 200 {object} models.APIResponse{data=TimewheelStatsResponse}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/timers/stats [get]
func (h *TimerHandler) GetStats(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting timer statistics",
		logger.String("request_id", requestID))

	// Get timewheel stats from core
	stats, err := h.coreInterface.GetTimewheelStats()
	if err != nil {
		logger.Error("Failed to get timewheel stats",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Timer statistics retrieved",
		logger.String("request_id", requestID),
		logger.Any("total_timers", stats.TotalTimers),
		logger.Any("pending_timers", stats.PendingTimers))

	c.JSON(http.StatusOK, models.SuccessResponse(stats, requestID))
}

// Helper methods

func (h *TimerHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("timer")
}
