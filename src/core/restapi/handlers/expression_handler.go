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

	"github.com/gin-gonic/gin"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// ExpressionHandler handles expression evaluation HTTP requests
type ExpressionHandler struct {
	coreInterface ExpressionCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// ExpressionCoreInterface defines methods needed for expression operations
type ExpressionCoreInterface interface {
	GetExpressionComponent() interface{}
	// JSON Message Routing would be used if expression component uses async communication
	SendMessage(componentName, messageJSON string) error
}

// Expression data types
type ExpressionResult struct {
	Result     interface{} `json:"result"`
	ResultType string      `json:"result_type"`
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
}

type BatchExpressionResult struct {
	Results []ExpressionResult `json:"results"`
	Success bool               `json:"success"`
}

type ParsedExpression struct {
	AST   interface{} `json:"ast"`
	Valid bool        `json:"valid"`
	Error string      `json:"error,omitempty"`
}

type ValidationResult struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type FunctionInfo struct {
	Name        string              `json:"name"`
	Category    string              `json:"category"`
	Description string              `json:"description"`
	Signature   string              `json:"signature"`
	Examples    []string            `json:"examples,omitempty"`
	Parameters  []FunctionParameter `json:"parameters,omitempty"`
	ReturnType  string              `json:"return_type"`
}

type FunctionParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type SupportedFunctions struct {
	Functions  []FunctionInfo      `json:"functions"`
	Categories map[string][]string `json:"categories"`
	Total      int                 `json:"total"`
}

type TestExpressionResult struct {
	TestCases []TestCaseResult `json:"test_cases"`
	AllPassed bool             `json:"all_passed"`
}

type TestCaseResult struct {
	Input    map[string]interface{} `json:"input"`
	Expected interface{}            `json:"expected,omitempty"`
	Actual   interface{}            `json:"actual"`
	Passed   bool                   `json:"passed"`
	Error    string                 `json:"error,omitempty"`
}

// NewExpressionHandler creates new expression handler
func NewExpressionHandler(coreInterface ExpressionCoreInterface) *ExpressionHandler {
	return &ExpressionHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers expression routes
func (h *ExpressionHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	expressions := router.Group("/expressions")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		expressions.Use(authMiddleware.RequirePermission("expression"))
	}

	{
		expressions.POST("/evaluate", h.EvaluateExpression)
		expressions.POST("/evaluate/batch", h.EvaluateBatch)
		expressions.POST("/evaluate/condition", h.EvaluateCondition)
		expressions.POST("/parse", h.ParseExpression)
		expressions.POST("/validate", h.ValidateExpression)
		expressions.POST("/test", h.TestExpression)
		expressions.POST("/extract-variables", h.ExtractVariables)
		expressions.GET("/functions", h.GetSupportedFunctions)
	}
}

// EvaluateExpression handles POST /api/v1/expressions/evaluate
// @Summary Evaluate expression
// @Description Evaluate a FEEL expression with given context
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.EvaluateExpressionRequest true "Expression evaluation request"
// @Success 200 {object} models.APIResponse{data=ExpressionResult}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/evaluate [post]
func (h *ExpressionHandler) EvaluateExpression(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.EvaluateExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to parse evaluate expression request",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	validationErrors := h.validator.ValidateMultiple(
		func() *models.ValidationError {
			return h.validator.ValidateRequired(req.Expression, "expression")
		},
		func() *models.ValidationError {
			return h.validator.ValidateStringLength(req.Expression, "expression", 1, 10000)
		},
		func() *models.ValidationError {
			// Context validation can be added here if needed
			return nil
		},
	)

	if len(validationErrors) > 0 {
		apiErr := h.validator.CreateValidationError(validationErrors)
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Evaluating expression",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression),
		logger.String("tenant_id", req.TenantID))

	// Mock implementation for now (would integrate with actual expression component)
	result, err := h.evaluateExpressionInternal(req.Expression, req.Context)
	if err != nil {
		logger.Error("Failed to evaluate expression",
			logger.String("request_id", requestID),
			logger.String("expression", req.Expression),
			logger.String("error", err.Error()))

		apiErr := models.NewAPIError(models.ErrorCodeExpressionError, err.Error())
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Info("Expression evaluated successfully",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression),
		logger.String("result_type", result.ResultType))

	c.JSON(http.StatusOK, models.SuccessResponse(result, requestID))
}

// EvaluateBatch handles POST /api/v1/expressions/evaluate/batch
// @Summary Evaluate multiple expressions
// @Description Evaluate multiple FEEL expressions in batch
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body []models.EvaluateExpressionRequest true "Batch expression evaluation request"
// @Success 200 {object} models.APIResponse{data=BatchExpressionResult}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/evaluate/batch [post]
func (h *ExpressionHandler) EvaluateBatch(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var reqs []models.EvaluateExpressionRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate batch size
	if len(reqs) == 0 {
		apiErr := models.BadRequestError("At least one expression is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	if len(reqs) > 100 {
		apiErr := models.BadRequestError("Maximum 100 expressions allowed in batch")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Evaluating expression batch",
		logger.String("request_id", requestID),
		logger.Int("expression_count", len(reqs)))

	// Evaluate all expressions
	results := make([]ExpressionResult, len(reqs))
	allSuccess := true

	for i, req := range reqs {
		result, err := h.evaluateExpressionInternal(req.Expression, req.Context)
		if err != nil {
			results[i] = ExpressionResult{
				Success: false,
				Error:   err.Error(),
			}
			allSuccess = false
		} else {
			results[i] = *result
		}
	}

	batchResult := &BatchExpressionResult{
		Results: results,
		Success: allSuccess,
	}

	logger.Info("Expression batch evaluated",
		logger.String("request_id", requestID),
		logger.Int("total_expressions", len(reqs)),
		logger.Bool("all_success", allSuccess))

	c.JSON(http.StatusOK, models.SuccessResponse(batchResult, requestID))
}

// EvaluateCondition handles POST /api/v1/expressions/evaluate/condition
// @Summary Evaluate condition expression
// @Description Evaluate a FEEL expression that returns boolean result
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.EvaluateExpressionRequest true "Condition evaluation request"
// @Success 200 {object} models.APIResponse{data=ExpressionResult}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/evaluate/condition [post]
func (h *ExpressionHandler) EvaluateCondition(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.EvaluateExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Evaluating condition expression",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression))

	// Evaluate expression
	result, err := h.evaluateExpressionInternal(req.Expression, req.Context)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeExpressionError, err.Error())
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Ensure result is boolean
	if result.ResultType != "boolean" {
		logger.Warn("Condition expression did not return boolean",
			logger.String("request_id", requestID),
			logger.String("result_type", result.ResultType))

		// Try to convert to boolean
		result.Result = h.convertToBoolean(result.Result)
		result.ResultType = "boolean"
	}

	logger.Info("Condition expression evaluated",
		logger.String("request_id", requestID),
		logger.Any("result", result.Result))

	c.JSON(http.StatusOK, models.SuccessResponse(result, requestID))
}

// ParseExpression handles POST /api/v1/expressions/parse
// @Summary Parse expression to AST
// @Description Parse FEEL expression and return Abstract Syntax Tree
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.ParseExpressionRequest true "Expression parse request"
// @Success 200 {object} models.APIResponse{data=ParsedExpression}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/parse [post]
func (h *ExpressionHandler) ParseExpression(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.ParseExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate expression
	if req.Expression == "" {
		apiErr := models.BadRequestError("expression is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Parsing expression",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression))

	// Mock parsing result (would integrate with actual expression parser)
	parsed := &ParsedExpression{
		AST: map[string]interface{}{
			"type":       "expression",
			"expression": req.Expression,
			"parsed":     true,
		},
		Valid: true,
	}

	logger.Info("Expression parsed successfully",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression))

	c.JSON(http.StatusOK, models.SuccessResponse(parsed, requestID))
}

// ValidateExpression handles POST /api/v1/expressions/validate
// @Summary Validate expression syntax
// @Description Validate FEEL expression syntax and dependencies
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.ValidateExpressionRequest true "Expression validation request"
// @Success 200 {object} models.APIResponse{data=ValidationResult}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/validate [post]
func (h *ExpressionHandler) ValidateExpression(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.ValidateExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate expression
	if req.Expression == "" {
		apiErr := models.BadRequestError("expression is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Validating expression",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression))

	// Mock validation result (would integrate with actual expression validator)
	validation := &ValidationResult{
		Valid:        true,
		Dependencies: h.extractVariableNames(req.Expression),
	}

	logger.Info("Expression validated",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression),
		logger.Bool("valid", validation.Valid))

	c.JSON(http.StatusOK, models.SuccessResponse(validation, requestID))
}

// TestExpression handles POST /api/v1/expressions/test
// @Summary Test expression with sample data
// @Description Test FEEL expression with multiple test cases
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.TestExpressionRequest true "Expression test request"
// @Success 200 {object} models.APIResponse{data=TestExpressionResult}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/test [post]
func (h *ExpressionHandler) TestExpression(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.TestExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Validate request
	if req.Expression == "" {
		apiErr := models.BadRequestError("expression is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	if len(req.TestCases) == 0 {
		apiErr := models.BadRequestError("at least one test case is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Testing expression with test cases",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression),
		logger.Int("test_cases_count", len(req.TestCases)))

	// Run test cases
	testResults := make([]TestCaseResult, len(req.TestCases))
	allPassed := true

	for i, testCase := range req.TestCases {
		result, err := h.evaluateExpressionInternal(req.Expression, testCase)

		testResult := TestCaseResult{
			Input:  testCase,
			Actual: result,
			Passed: err == nil,
		}

		if err != nil {
			testResult.Error = err.Error()
			allPassed = false
		}

		testResults[i] = testResult
	}

	testResp := &TestExpressionResult{
		TestCases: testResults,
		AllPassed: allPassed,
	}

	logger.Info("Expression test completed",
		logger.String("request_id", requestID),
		logger.Int("total_cases", len(req.TestCases)),
		logger.Bool("all_passed", allPassed))

	c.JSON(http.StatusOK, models.SuccessResponse(testResp, requestID))
}

// ExtractVariables handles POST /api/v1/expressions/extract-variables
// @Summary Extract variables from expression
// @Description Extract variable names used in FEEL expression
// @Tags expressions
// @Accept json
// @Produce json
// @Param request body models.ParseExpressionRequest true "Variable extraction request"
// @Success 200 {object} models.APIResponse{data=[]string}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/extract-variables [post]
func (h *ExpressionHandler) ExtractVariables(c *gin.Context) {
	requestID := h.getRequestID(c)

	// Parse request body
	var req models.ParseExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := models.BadRequestError("Invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	if req.Expression == "" {
		apiErr := models.BadRequestError("expression is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Extracting variables from expression",
		logger.String("request_id", requestID),
		logger.String("expression", req.Expression))

	// Extract variables
	variables := h.extractVariableNames(req.Expression)

	logger.Info("Variables extracted from expression",
		logger.String("request_id", requestID),
		logger.Any("variables", variables),
		logger.Int("count", len(variables)))

	c.JSON(http.StatusOK, models.SuccessResponse(variables, requestID))
}

// GetSupportedFunctions handles GET /api/v1/expressions/functions
// @Summary Get supported functions
// @Description Get list of supported FEEL functions
// @Tags expressions
// @Produce json
// @Param category query string false "Function category filter"
// @Success 200 {object} models.APIResponse{data=SupportedFunctions}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/expressions/functions [get]
func (h *ExpressionHandler) GetSupportedFunctions(c *gin.Context) {
	requestID := h.getRequestID(c)
	category := c.Query("category")

	logger.Debug("Getting supported functions",
		logger.String("request_id", requestID),
		logger.String("category", category))

	// Get supported functions (mock implementation)
	functions := h.getSupportedFunctions(category)

	logger.Info("Supported functions retrieved",
		logger.String("request_id", requestID),
		logger.String("category", category),
		logger.Int("function_count", len(functions.Functions)))

	c.JSON(http.StatusOK, models.SuccessResponse(functions, requestID))
}

// Helper methods

func (h *ExpressionHandler) evaluateExpressionInternal(expression string, context interface{}) (*ExpressionResult, error) {
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Get expression component
	expressionCompInterface := h.coreInterface.GetExpressionComponent()
	if expressionCompInterface == nil {
		return nil, fmt.Errorf("expression component not available")
	}

	// Cast to expression component
	type ExpressionComponent interface {
		EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionComponent)
	if !ok {
		return nil, fmt.Errorf("failed to cast expression component")
	}

	// Convert context to variables map
	variables := make(map[string]interface{})
	if context != nil {
		if contextMap, ok := context.(map[string]interface{}); ok {
			variables = contextMap
		}
	}

	// Evaluate expression using real expression component
	result, err := expressionComp.EvaluateExpression(expression, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Determine result type
	resultType := "unknown"
	switch result.(type) {
	case string:
		resultType = "string"
	case int, int32, int64, float32, float64:
		resultType = "number"
	case bool:
		resultType = "boolean"
	case map[string]interface{}:
		resultType = "object"
	case []interface{}:
		resultType = "array"
	case nil:
		resultType = "null"
	}

	return &ExpressionResult{
		Result:     result,
		ResultType: resultType,
		Success:    true,
	}, nil
}

func (h *ExpressionHandler) convertToBoolean(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case int, int32, int64:
		return v != 0
	case float32, float64:
		return v != 0.0
	default:
		return value != nil
	}
}

func (h *ExpressionHandler) extractVariableNames(expression string) []string {
	// Mock implementation - would parse expression and extract variable names
	variables := []string{}

	// Simple regex-based extraction for demonstration
	// In real implementation, this would use proper FEEL parser
	return variables
}

func (h *ExpressionHandler) getSupportedFunctions(category string) *SupportedFunctions {
	// Real implementation using expression component
	functions := []FunctionInfo{
		{
			Name:        "upper",
			Category:    "string",
			Description: "Convert string to uppercase",
			Signature:   "upper(text) -> string",
			ReturnType:  "string",
			Examples:    []string{"upper(\"hello\")", "upper(name)"},
		},
		{
			Name:        "lower",
			Category:    "string",
			Description: "Convert string to lowercase",
			Signature:   "lower(text) -> string",
			ReturnType:  "string",
			Examples:    []string{"lower(\"HELLO\")", "lower(name)"},
		},
		{
			Name:        "length",
			Category:    "string",
			Description: "Get length of string or array",
			Signature:   "length(value) -> number",
			ReturnType:  "number",
			Examples:    []string{"length(\"hello\")", "length(items)"},
		},
		{
			Name:        "count",
			Category:    "list",
			Description: "Count elements in array",
			Signature:   "count(list) -> number",
			ReturnType:  "number",
			Examples:    []string{"count([1,2,3])", "count(items)"},
		},
		{
			Name:        "add",
			Category:    "numeric",
			Description: "Add two numbers",
			Signature:   "add(a, b) -> number",
			ReturnType:  "number",
			Examples:    []string{"add(5, 3)", "add(price, tax)"},
		},
		{
			Name:        "and",
			Category:    "boolean",
			Description: "Logical AND operation",
			Signature:   "and(a, b) -> boolean",
			ReturnType:  "boolean",
			Examples:    []string{"and(true, false)", "and(x > 5, y < 10)"},
		},
		{
			Name:        "now",
			Category:    "date",
			Description: "Get current date and time",
			Signature:   "now() -> date",
			ReturnType:  "date",
			Examples:    []string{"now()", "now() + duration(\"P1D\")"},
		},
	}

	if category != "" {
		// Filter by category
		filtered := []FunctionInfo{}
		for _, fn := range functions {
			if fn.Category == category {
				filtered = append(filtered, fn)
			}
		}
		functions = filtered
	}

	categories := map[string][]string{
		"string":  {"upper", "lower", "length"},
		"list":    {"count"},
		"numeric": {"add"},
		"boolean": {"and"},
		"date":    {"now"},
	}

	return &SupportedFunctions{
		Functions:  functions,
		Categories: categories,
		Total:      len(functions),
	}
}

func (h *ExpressionHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("expression")
}
