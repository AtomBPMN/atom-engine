/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/interfaces"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// StorageHandler handles storage-related HTTP requests
type StorageHandler struct {
	coreInterface CoreInterface
	converter     *utils.Converter
}

// CoreInterface defines methods needed for storage operations
type CoreInterface interface {
	GetStorageStatus() (*interfaces.StorageStatusResponse, error)
	GetStorageInfo() (*interfaces.StorageInfoResponse, error)
}

// Response types for storage operations
type StorageStatusResponse struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type StorageInfoResponse struct {
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	TotalKeys      int64             `json:"total_keys"`
	DatabasePath   string            `json:"database_path"`
	Statistics     map[string]string `json:"statistics"`
}

// NewStorageHandler creates new storage handler
func NewStorageHandler(coreInterface CoreInterface) *StorageHandler {
	return &StorageHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
	}
}

// RegisterRoutes registers storage routes
func (h *StorageHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	storage := router.Group("/storage")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		storage.Use(authMiddleware.RequirePermission("storage"))
	}

	{
		storage.GET("/status", h.GetStatus)
		storage.GET("/info", h.GetInfo)
	}
}

// GetStatus handles GET /api/v1/storage/status
// @Summary Get storage status
// @Description Get current storage connection and health status
// @Tags storage
// @Produce json
// @Success 200 {object} models.APIResponse{data=StorageStatusResponse}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/storage/status [get]
func (h *StorageHandler) GetStatus(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting storage status",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Get storage status from core
	status, err := h.coreInterface.GetStorageStatus()
	if err != nil {
		logger.Error("Failed to get storage status",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert to response format
	response := &models.StatusResponse{
		Status:    status.Status,
		Uptime:    status.UptimeSeconds,
		IsHealthy: status.IsHealthy,
	}

	logger.Info("Storage status retrieved",
		logger.String("request_id", requestID),
		logger.String("status", status.Status),
		logger.Bool("is_healthy", status.IsHealthy))

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

// GetInfo handles GET /api/v1/storage/info
// @Summary Get storage information
// @Description Get detailed storage information including size and statistics
// @Tags storage
// @Produce json
// @Success 200 {object} models.APIResponse{data=StorageInfoResponse}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/storage/info [get]
func (h *StorageHandler) GetInfo(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting storage info",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Get storage info from core
	info, err := h.coreInterface.GetStorageInfo()
	if err != nil {
		logger.Error("Failed to get storage info",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Storage info is already in correct format for response
	response := info

	logger.Info("Storage info retrieved",
		logger.String("request_id", requestID),
		logger.Int64("total_size", info.TotalSizeBytes),
		logger.Int64("used_size", info.UsedSizeBytes),
		logger.Int64("total_keys", info.TotalKeys))

	c.JSON(http.StatusOK, models.SuccessResponse(response, requestID))
}

// getRequestID extracts request ID from context
func (h *StorageHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return h.generateRequestID()
}

// generateRequestID generates a simple request ID
func (h *StorageHandler) generateRequestID() string {
	return utils.GenerateSecureRequestID("storage")
}

// generateRandomString generates random string of given length

// GetHealthStatus provides storage health information for health checks
func (h *StorageHandler) GetHealthStatus() map[string]interface{} {
	status, err := h.coreInterface.GetStorageStatus()
	if err != nil {
		return map[string]interface{}{
			"storage": map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			},
		}
	}

	return map[string]interface{}{
		"storage": map[string]interface{}{
			"status":       status.Status,
			"is_connected": status.IsConnected,
			"is_healthy":   status.IsHealthy,
			"uptime":       status.UptimeSeconds,
		},
	}
}

// ValidateStoragePermissions checks if user has required storage permissions
func (h *StorageHandler) ValidateStoragePermissions(c *gin.Context, operation string) *models.APIError {
	authResult, exists := middleware.GetAuthResult(c)
	if !exists {
		return models.UnauthorizedError("Authentication required")
	}

	// Define permission requirements for different operations
	var requiredPermission string
	switch operation {
	case "read":
		requiredPermission = "storage:read"
	case "write":
		requiredPermission = "storage:write"
	case "admin":
		requiredPermission = "storage:admin"
	default:
		requiredPermission = "storage"
	}

	// Check if user has required permission
	hasPermission := false
	for _, permission := range authResult.Permissions {
		if permission == requiredPermission || permission == "storage" || permission == "*" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return models.ForbiddenError("Insufficient storage permissions")
	}

	return nil
}

// StorageStats provides storage statistics for monitoring
type StorageStats struct {
	TotalRequests    int64   `json:"total_requests"`
	SuccessfulOps    int64   `json:"successful_operations"`
	FailedOps        int64   `json:"failed_operations"`
	AverageLatency   float64 `json:"average_latency_ms"`
	LastHealthCheck  int64   `json:"last_health_check"`
	DatabaseSize     int64   `json:"database_size_bytes"`
	KeyCount         int64   `json:"key_count"`
	ConnectionStatus string  `json:"connection_status"`
}

// GetStorageStats returns storage operation statistics
func (h *StorageHandler) GetStorageStats() (*StorageStats, error) {
	// Get current storage info
	info, err := h.coreInterface.GetStorageInfo()
	if err != nil {
		return nil, err
	}

	status, err := h.coreInterface.GetStorageStatus()
	if err != nil {
		return nil, err
	}

	stats := &StorageStats{
		DatabaseSize:     info.TotalSizeBytes,
		KeyCount:         info.TotalKeys,
		ConnectionStatus: status.Status,
		LastHealthCheck:  status.UptimeSeconds,
	}

	return stats, nil
}
