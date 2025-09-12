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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"atom-engine/proto/parser/parserpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// ParserHandler handles BPMN parsing HTTP requests
type ParserHandler struct {
	coreInterface ParserCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// ParserCoreInterface defines methods needed for BPMN operations
type ParserCoreInterface interface {
	// JSON Message Routing to parser component
	SendMessage(componentName, messageJSON string) error
	WaitForParserResponse(timeoutMs int) (string, error)
	// gRPC connection for direct calls
	GetGRPCConnection() (interface{}, error)
}

// BPMN response types
type BPMNProcess struct {
	ID           string                 `json:"id"`
	Key          string                 `json:"key"`
	Name         string                 `json:"name"`
	Version      int32                  `json:"version"`
	Description  string                 `json:"description"`
	CreatedAt    int64                  `json:"created_at"`
	UpdatedAt    int64                  `json:"updated_at"`
	ElementCount int32                  `json:"element_count"`
	IsDeployable bool                   `json:"is_deployable"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type BPMNProcessDetails struct {
	ProcessKey     string           `json:"process_key"`
	ProcessID      string           `json:"process_id"`
	ProcessName    string           `json:"process_name"`
	Version        string           `json:"version"`
	ProcessVersion int32            `json:"process_version"`
	Status         string           `json:"status"`
	TotalElements  int32            `json:"total_elements"`
	ElementCounts  map[string]int32 `json:"element_counts"`
	ContentHash    string           `json:"content_hash"`
	OriginalFile   string           `json:"original_file"`
	CreatedAt      string           `json:"created_at"`
	UpdatedAt      string           `json:"updated_at"`
	ParsedAt       string           `json:"parsed_at"`
}

type BPMNStats struct {
	TotalProcesses   int32            `json:"total_processes"`
	ActiveProcesses  int32            `json:"active_processes"`
	ProcessesByType  map[string]int32 `json:"processes_by_type"`
	TotalElements    int32            `json:"total_elements"`
	ElementsByType   map[string]int32 `json:"elements_by_type"`
	LastParsed       int64            `json:"last_parsed"`
	ParseSuccessRate float64          `json:"parse_success_rate"`
}

// NewParserHandler creates new parser handler
func NewParserHandler(coreInterface ParserCoreInterface) *ParserHandler {
	return &ParserHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers BPMN routes
func (h *ParserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	bpmn := router.Group("/bpmn")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		bpmn.Use(authMiddleware.RequirePermission("bpmn"))
	}

	{
		bpmn.POST("/parse", h.ParseBPMN)
		bpmn.GET("/processes", h.ListProcesses)
		bpmn.GET("/processes/:key", h.GetProcess)
		bpmn.DELETE("/processes/:id", h.DeleteBPMNProcess)
		bpmn.GET("/processes/:key/json", h.GetBPMNProcessJSON)
		bpmn.GET("/processes/:key/xml", h.GetBPMNProcessXML)
		bpmn.GET("/stats", h.GetBPMNStats)
	}
}

// ParseBPMN handles POST /api/v1/bpmn/parse
// @Summary Parse BPMN file
// @Description Parse and store BPMN process definition
// @Tags bpmn
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "BPMN file"
// @Param process_id formData string false "Process ID"
// @Param force formData boolean false "Force overwrite existing process"
// @Success 201 {object} models.APIResponse{data=models.CreateResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 409 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/parse [post]
func (h *ParserHandler) ParseBPMN(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Parsing BPMN file",
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		logger.Error("Failed to parse multipart form",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid multipart form data")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Get BPMN file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("No BPMN file provided",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("BPMN file is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer file.Close()

	// Validate file type
	if !h.isValidBPMNFile(header) {
		apiErr := models.BadRequestError("Invalid file type. Only .bpmn and .xml files are allowed")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Read file content
	bpmnContent, err := h.readFileContent(file)
	if err != nil {
		logger.Error("Failed to read BPMN file",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to read BPMN file")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Get optional parameters
	processID := c.Request.FormValue("process_id")
	forceStr := c.Request.FormValue("force")
	force, _ := strconv.ParseBool(forceStr)

	// Create parse request
	parseReq := map[string]interface{}{
		"type":       "parse_bpmn_content",
		"request_id": requestID,
		"payload": map[string]interface{}{
			"bpmn_content": bpmnContent,
			"process_id":   processID,
			"force":        force,
		},
	}

	// Send to parser component
	reqJSON, err := json.Marshal(parseReq)
	if err != nil {
		logger.Error("Failed to marshal parse request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to process request")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	err = h.coreInterface.SendMessage("parser", string(reqJSON))
	if err != nil {
		logger.Error("Failed to send message to parser",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to communicate with parser service")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Wait for response
	respJSON, err := h.coreInterface.WaitForParserResponse(30000) // 30 seconds timeout
	if err != nil {
		logger.Error("Failed to get parser response",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service timeout")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse response
	var parseResp map[string]interface{}
	err = json.Unmarshal([]byte(respJSON), &parseResp)
	if err != nil {
		logger.Error("Failed to parse parser response",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Invalid parser response")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if parsing was successful
	success, _ := parseResp["success"].(bool)
	if !success {
		errorMsg, _ := parseResp["error"].(string)
		if errorMsg == "" {
			errorMsg = "BPMN parsing failed"
		}

		logger.Warn("BPMN parsing failed",
			logger.String("request_id", requestID),
			logger.String("error", errorMsg))

		// Determine appropriate error type
		var apiErr *models.APIError
		if strings.Contains(strings.ToLower(errorMsg), "already exists") {
			apiErr = models.ConflictError(errorMsg)
		} else if strings.Contains(strings.ToLower(errorMsg), "invalid") ||
			strings.Contains(strings.ToLower(errorMsg), "validation") {
			apiErr = models.NewAPIError(models.ErrorCodeBPMNValidationError, errorMsg)
		} else {
			apiErr = models.NewAPIError(models.ErrorCodeBPMNParseError, errorMsg)
		}

		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Extract process information from response
	processKey, _ := parseResp["process_key"].(string)
	processName, _ := parseResp["process_name"].(string)
	if processKey == "" {
		processKey = processID
	}

	response := &models.CreateResponse{
		ID:      processKey,
		Message: fmt.Sprintf("BPMN process '%s' parsed successfully", processName),
	}

	logger.Info("BPMN file parsed successfully",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey),
		logger.String("file_name", header.Filename))

	c.JSON(http.StatusCreated, models.SuccessResponse(response, requestID))
}

// ListProcesses handles GET /api/v1/bpmn/processes
// @Summary List BPMN processes
// @Description Get list of all BPMN processes with pagination
// @Tags bpmn
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param tenant_id query string false "Tenant ID filter"
// @Success 200 {object} models.PaginatedResponse{data=[]BPMNProcess}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/processes [get]
func (h *ParserHandler) ListProcesses(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	_ = c.Query("tenant_id") // tenantID for future implementation

	paginationHelper := utils.NewPaginationHelper()
	params, apiErr := paginationHelper.ParseAndValidate(pageStr, limitStr)
	if apiErr != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Listing BPMN processes",
		logger.String("request_id", requestID),
		logger.Int("page", params.Page),
		logger.Int("limit", params.Limit))

	// Get gRPC client (same as other methods in this handler)
	client, conn, err := h.getParserGRPCClient()
	if err != nil {
		logger.Error("Failed to get Parser gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC ListBPMNProcesses method (same as CLI)
	grpcReq := &parserpb.ListBPMNProcessesRequest{
		Limit:     0, // Use pagination instead
		PageSize:  int32(params.Limit),
		Page:      int32(params.Page),
		SortBy:    "created_at",
		SortOrder: "DESC",
	}

	resp, err := client.ListBPMNProcesses(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to list BPMN processes via gRPC",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "Failed to list BPMN processes"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert gRPC response to REST API format
	processes := h.convertGRPCProcessesToREST(resp.Processes)

	logger.Info("BPMN processes listed",
		logger.String("request_id", requestID),
		logger.Int("count", len(processes)),
		logger.Int("total", int(resp.TotalCount)))

	// Create pagination info from gRPC response
	paginationInfo := &models.PaginationInfo{
		Page:    int(resp.Page),
		Limit:   int(resp.PageSize),
		Total:   int(resp.TotalCount),
		Pages:   int(resp.TotalPages),
		HasNext: resp.Page < resp.TotalPages,
		HasPrev: resp.Page > 1,
	}

	paginatedResp := models.PaginatedSuccessResponse(processes, paginationInfo, requestID)
	c.JSON(http.StatusOK, paginatedResp)
}

// GetProcess handles GET /api/v1/bpmn/processes/:key
func (h *ParserHandler) GetProcess(c *gin.Context) {
	requestID := h.getRequestID(c)
	processKey := c.Param("key")

	if processKey == "" {
		apiErr := models.BadRequestError("Process key is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting BPMN process",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey))

	// Get gRPC connection
	connInterface, err := h.coreInterface.GetGRPCConnection()
	if err != nil {
		logger.Error("Failed to get gRPC connection", logger.String("error", err.Error()))
		apiErr := models.InternalServerError("Internal service error")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	conn, ok := connInterface.(*grpc.ClientConn)
	if !ok {
		logger.Error("Invalid gRPC connection type")
		apiErr := models.InternalServerError("Internal service error")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Create gRPC client and call GetBPMNProcess
	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetBPMNProcess(ctx, &parserpb.GetBPMNProcessRequest{
		ProcessKey: processKey,
	})
	if err != nil {
		logger.Error("Failed to get BPMN process from gRPC",
			logger.String("process_key", processKey),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Failed to retrieve process information")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	if !resp.Success {
		logger.Warn("BPMN process not found",
			logger.String("process_key", processKey),
			logger.String("message", resp.Message))

		apiErr := models.ProcessNotFoundError(fmt.Sprintf("Process with key '%s' not found", processKey))
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert gRPC response to REST API format
	processDetails := h.convertGRPCProcessDetailsToREST(resp.Process)

	logger.Info("BPMN process retrieved successfully",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey),
		logger.String("process_name", processDetails.ProcessName))

	c.JSON(http.StatusOK, models.SuccessResponse(processDetails, requestID))
}

// Helper methods

func (h *ParserHandler) isValidBPMNFile(header *multipart.FileHeader) bool {
	filename := strings.ToLower(header.Filename)
	return strings.HasSuffix(filename, ".bpmn") || strings.HasSuffix(filename, ".xml")
}

func (h *ParserHandler) readFileContent(file multipart.File) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (h *ParserHandler) sendParserRequest(req map[string]interface{}, requestID string) (map[string]interface{}, error) {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	err = h.coreInterface.SendMessage("parser", string(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	respJSON, err := h.coreInterface.WaitForParserResponse(30000)
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

// convertGRPCProcessesToREST converts gRPC BPMNProcessSummary to REST API BPMNProcess format
func (h *ParserHandler) convertGRPCProcessesToREST(grpcProcesses []*parserpb.BPMNProcessSummary) []BPMNProcess {
	processes := make([]BPMNProcess, len(grpcProcesses))

	for i, grpcProcess := range grpcProcesses {
		// Parse created_at time
		var createdAt int64
		if parsedTime, err := time.Parse(time.RFC3339, grpcProcess.CreatedAt); err == nil {
			createdAt = parsedTime.Unix()
		}

		// Parse updated_at time
		var updatedAt int64
		if parsedTime, err := time.Parse(time.RFC3339, grpcProcess.UpdatedAt); err == nil {
			updatedAt = parsedTime.Unix()
		}

		// Convert version string to int32
		var version int32 = 1
		if v, err := strconv.Atoi(strings.TrimPrefix(grpcProcess.Version, "v")); err == nil {
			version = int32(v)
		}

		processes[i] = BPMNProcess{
			ID:           grpcProcess.ProcessId,
			Key:          grpcProcess.ProcessKey,
			Name:         grpcProcess.ProcessName,
			Version:      version,
			Description:  "", // Not available in gRPC summary
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			ElementCount: grpcProcess.TotalElements,
			IsDeployable: grpcProcess.Status == "active",
			Metadata: map[string]interface{}{
				"status":         grpcProcess.Status,
				"version_string": grpcProcess.Version,
				"total_elements": grpcProcess.TotalElements,
			},
		}
	}

	return processes
}

func (h *ParserHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("parser")
}

// DeleteBPMNProcess handles DELETE /api/v1/bpmn/processes/:id
// @Summary Delete BPMN process
// @Description Delete a BPMN process by process ID
// @Tags bpmn
// @Produce json
// @Param id path string true "Process ID"
// @Success 200 {object} models.APIResponse{data=models.DeleteResponse}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/processes/{id} [delete]
func (h *ParserHandler) DeleteBPMNProcess(c *gin.Context) {
	requestID := h.getRequestID(c)
	processID := c.Param("id")

	if processID == "" {
		apiErr := models.BadRequestError("Process ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Deleting BPMN process",
		logger.String("request_id", requestID),
		logger.String("process_id", processID))

	// Get gRPC client
	client, conn, err := h.getParserGRPCClient()
	if err != nil {
		logger.Error("Failed to get Parser gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call gRPC DeleteBPMNProcess method
	grpcReq := &parserpb.DeleteBPMNProcessRequest{
		ProcessId: processID,
	}

	resp, err := client.DeleteBPMNProcess(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to delete BPMN process via gRPC",
			logger.String("request_id", requestID),
			logger.String("process_id", processID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "BPMN process deletion failed"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("BPMN process deleted successfully",
		logger.String("request_id", requestID),
		logger.String("process_id", processID))

	deleteResp := &models.DeleteResponse{
		ID:      processID,
		Message: "BPMN process deleted successfully",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(deleteResp, requestID))
}

// GetBPMNStats handles GET /api/v1/bpmn/stats
// @Summary Get BPMN statistics
// @Description Get statistics about BPMN parsing and processes
// @Tags bpmn
// @Produce json
// @Success 200 {object} models.APIResponse{data=BPMNStats}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/stats [get]
func (h *ParserHandler) GetBPMNStats(c *gin.Context) {
	requestID := h.getRequestID(c)

	logger.Debug("Getting BPMN stats",
		logger.String("request_id", requestID))

	// Get gRPC client
	client, conn, err := h.getParserGRPCClient()
	if err != nil {
		logger.Error("Failed to get Parser gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC GetBPMNStats method
	grpcReq := &parserpb.GetBPMNStatsRequest{}

	resp, err := client.GetBPMNStats(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to get BPMN stats via gRPC",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "Failed to get BPMN stats"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.InternalServerError(message)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert gRPC response to REST API format
	stats := &BPMNStats{
		TotalProcesses:   resp.TotalProcesses,
		ActiveProcesses:  resp.ActiveProcesses,
		TotalElements:    resp.TotalElementsParsed,
		ParseSuccessRate: float64(resp.SuccessfulElements) / float64(resp.TotalElementsParsed) * 100,
		LastParsed:       0, // Convert from string if needed
	}

	// Convert element counts from map[string]int32 to map[string]int32
	if resp.ElementTypeCounts != nil {
		stats.ElementsByType = make(map[string]int32)
		for k, v := range resp.ElementTypeCounts {
			stats.ElementsByType[k] = v
		}
	}

	// Add processes by type placeholder
	stats.ProcessesByType = make(map[string]int32)
	stats.ProcessesByType["total"] = resp.TotalProcesses

	logger.Info("BPMN stats retrieved",
		logger.String("request_id", requestID),
		logger.Int("total_processes", int(stats.TotalProcesses)))

	c.JSON(http.StatusOK, models.SuccessResponse(stats, requestID))
}

// GetBPMNProcessJSON handles GET /api/v1/bpmn/processes/:key/json
// @Summary Get BPMN process JSON
// @Description Get JSON data of a BPMN process by process key
// @Tags bpmn
// @Produce json
// @Param key path string true "Process Key"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/processes/{key}/json [get]
func (h *ParserHandler) GetBPMNProcessJSON(c *gin.Context) {
	requestID := h.getRequestID(c)
	processKey := c.Param("key")

	if processKey == "" {
		apiErr := models.BadRequestError("Process key is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting BPMN process JSON",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey))

	// Get gRPC client
	client, conn, err := h.getParserGRPCClient()
	if err != nil {
		logger.Error("Failed to get Parser gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC GetBPMNProcessJSON method
	grpcReq := &parserpb.GetBPMNProcessJSONRequest{
		ProcessKey: processKey,
	}

	resp, err := client.GetBPMNProcessJSON(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to get BPMN process JSON via gRPC",
			logger.String("request_id", requestID),
			logger.String("process_key", processKey),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "BPMN process not found"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.NotFoundError(message)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Parse JSON data
	var jsonData interface{}
	if err := json.Unmarshal([]byte(resp.JsonData), &jsonData); err != nil {
		logger.Error("Failed to parse BPMN JSON data",
			logger.String("request_id", requestID),
			logger.String("process_key", processKey),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Invalid JSON data in BPMN process")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("BPMN process JSON retrieved",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey))

	c.JSON(http.StatusOK, models.SuccessResponse(jsonData, requestID))
}

// GetBPMNProcessXML handles GET /api/v1/bpmn/processes/:key/xml
// @Summary Get BPMN process original XML
// @Description Get original XML content of a BPMN process by process key
// @Tags bpmn
// @Produce text/xml
// @Param key path string true "Process Key"
// @Success 200 {string} string "Original BPMN XML content"
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/bpmn/processes/{key}/xml [get]
func (h *ParserHandler) GetBPMNProcessXML(c *gin.Context) {
	requestID := h.getRequestID(c)
	processKey := c.Param("key")

	if processKey == "" {
		apiErr := models.BadRequestError("Process key is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting BPMN process XML",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey))

	// Get gRPC client
	client, conn, err := h.getParserGRPCClient()
	if err != nil {
		logger.Error("Failed to get Parser gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Parser service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC GetBPMNProcessXML method
	grpcReq := &parserpb.GetBPMNProcessXMLRequest{
		ProcessKey: processKey,
	}

	resp, err := client.GetBPMNProcessXML(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to get BPMN process XML via gRPC",
			logger.String("request_id", requestID),
			logger.String("process_key", processKey),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "BPMN process XML not found"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.NotFoundError(message)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("BPMN process XML retrieved",
		logger.String("request_id", requestID),
		logger.String("process_key", processKey),
		logger.Int("file_size", int(resp.FileSize)))

	// Set appropriate headers for XML content
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", resp.Filename))
	c.Header("Content-Length", fmt.Sprintf("%d", resp.FileSize))

	// Return raw XML content
	c.String(http.StatusOK, resp.XmlData)
}

// Helper method to get Parser gRPC client
func (h *ParserHandler) getParserGRPCClient() (parserpb.ParserServiceClient, *grpc.ClientConn, error) {
	conn, err := h.coreInterface.GetGRPCConnection()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get gRPC connection: %w", err)
	}

	grpcConn, ok := conn.(*grpc.ClientConn)
	if !ok {
		return nil, nil, fmt.Errorf("invalid gRPC connection type")
	}

	client := parserpb.NewParserServiceClient(grpcConn)
	return client, grpcConn, nil
}

// convertGRPCProcessDetailsToREST converts gRPC BPMNProcessDetails to REST API BPMNProcessDetails format
func (h *ParserHandler) convertGRPCProcessDetailsToREST(grpcDetails *parserpb.BPMNProcessDetails) *BPMNProcessDetails {
	if grpcDetails == nil {
		return nil
	}

	return &BPMNProcessDetails{
		ProcessKey:     grpcDetails.ProcessKey,
		ProcessID:      grpcDetails.ProcessId,
		ProcessName:    grpcDetails.ProcessName,
		Version:        grpcDetails.Version,
		ProcessVersion: grpcDetails.ProcessVersion,
		Status:         grpcDetails.Status,
		TotalElements:  grpcDetails.TotalElements,
		ElementCounts:  grpcDetails.ElementCounts,
		ContentHash:    grpcDetails.ContentHash,
		OriginalFile:   grpcDetails.OriginalFile,
		CreatedAt:      grpcDetails.CreatedAt,
		UpdatedAt:      grpcDetails.UpdatedAt,
		ParsedAt:       grpcDetails.ParsedAt,
	}
}
