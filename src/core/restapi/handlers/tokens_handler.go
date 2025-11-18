/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"atom-engine/proto/process/processpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/restapi/middleware"
	"atom-engine/src/core/restapi/models"
	"atom-engine/src/core/restapi/utils"
)

// TokensHandler handles token management HTTP requests
type TokensHandler struct {
	coreInterface TokensCoreInterface
	converter     *utils.Converter
	validator     *utils.Validator
}

// TokensCoreInterface defines methods needed for tokens operations
type TokensCoreInterface interface {
	// gRPC connection for direct calls
	GetGRPCConnection() (interface{}, error)
}

// TokenInfo represents token information for REST API
type TokenInfo struct {
	ID                string                 `json:"id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ProcessKey        string                 `json:"process_key"`
	CurrentElementID  string                 `json:"current_element_id"`
	State             string                 `json:"state"`
	WaitingFor        string                 `json:"waiting_for,omitempty"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
}

// NewTokensHandler creates new tokens handler
func NewTokensHandler(coreInterface TokensCoreInterface) *TokensHandler {
	return &TokensHandler{
		coreInterface: coreInterface,
		converter:     utils.NewConverter(),
		validator:     utils.NewValidator(),
	}
}

// RegisterRoutes registers token routes
func (h *TokensHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	tokens := router.Group("/tokens")

	// Apply auth middleware with required permissions
	if authMiddleware != nil {
		tokens.Use(authMiddleware.RequirePermission("token"))
	}

	{
		tokens.GET("/:id", h.GetTokenStatus)
	}
}

// GetTokenStatus handles GET /api/v1/tokens/:id
// @Summary Get token status
// @Description Get detailed information about a specific token
// @Tags tokens
// @Produce json
// @Param id path string true "Token ID"
// @Success 200 {object} models.APIResponse{data=TokenInfo}
// @Failure 400 {object} models.APIResponse{error=models.APIError}
// @Failure 401 {object} models.APIResponse{error=models.APIError}
// @Failure 403 {object} models.APIResponse{error=models.APIError}
// @Failure 404 {object} models.APIResponse{error=models.APIError}
// @Failure 500 {object} models.APIResponse{error=models.APIError}
// @Security ApiKeyAuth
// @Router /api/v1/tokens/{id} [get]
func (h *TokensHandler) GetTokenStatus(c *gin.Context) {
	requestID := h.getRequestID(c)
	tokenID := c.Param("id")

	if tokenID == "" {
		apiErr := models.BadRequestError("Token ID is required")
		c.JSON(http.StatusBadRequest, models.ErrorResponse(apiErr, requestID))
		return
	}

	logger.Debug("Getting token status",
		logger.String("request_id", requestID),
		logger.String("token_id", tokenID))

	// Get gRPC client
	client, conn, err := h.getProcessGRPCClient()
	if err != nil {
		logger.Error("Failed to get Process gRPC client",
			logger.String("request_id", requestID),
			logger.String("error", err.Error()))

		apiErr := models.InternalServerError("Process service not available")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(apiErr, requestID))
		return
	}
	defer conn.Close()

	// Create gRPC context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC GetTokenStatus method
	grpcReq := &processpb.GetTokenStatusRequest{
		TokenId: tokenID,
	}

	resp, err := client.GetTokenStatus(ctx, grpcReq)
	if err != nil {
		logger.Error("Failed to get token status via gRPC",
			logger.String("request_id", requestID),
			logger.String("token_id", tokenID),
			logger.String("error", err.Error()))

		apiErr := h.converter.GRPCErrorToAPIError(err)
		statusCode := models.HTTPStatusFromErrorCode(apiErr.Code)
		c.JSON(statusCode, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Check if operation succeeded
	if !resp.Success {
		message := "Token not found"
		if resp.Message != "" {
			message = resp.Message
		}
		apiErr := models.NotFoundError(message)
		c.JSON(http.StatusNotFound, models.ErrorResponse(apiErr, requestID))
		return
	}

	// Convert gRPC token to REST API format
	token := &TokenInfo{
		ID:                resp.Token.TokenId,
		ProcessInstanceID: resp.Token.ProcessInstanceId,
		ProcessKey:        resp.Token.ProcessKey,
		CurrentElementID:  resp.Token.CurrentElementId,
		State:             resp.Token.State,
		WaitingFor:        resp.Token.WaitingFor,
		CreatedAt:         resp.Token.CreatedAt,
		UpdatedAt:         resp.Token.UpdatedAt,
	}

	// Convert variables from map[string]string to map[string]interface{}
	if resp.Token.Variables != nil {
		token.Variables = make(map[string]interface{})
		for k, v := range resp.Token.Variables {
			token.Variables[k] = v
		}
	}

	logger.Info("Token status retrieved",
		logger.String("request_id", requestID),
		logger.String("token_id", tokenID),
		logger.String("state", token.State))

	c.JSON(http.StatusOK, models.SuccessResponse(token, requestID))
}

// Helper methods

func (h *TokensHandler) getProcessGRPCClient() (processpb.ProcessServiceClient, *grpc.ClientConn, error) {
	conn, err := h.coreInterface.GetGRPCConnection()
	if err != nil {
		return nil, nil, err
	}

	grpcConn, ok := conn.(*grpc.ClientConn)
	if !ok {
		return nil, nil, fmt.Errorf("invalid gRPC connection type")
	}

	client := processpb.NewProcessServiceClient(grpcConn)
	return client, grpcConn, nil
}

func (h *TokensHandler) getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	return utils.GenerateSecureRequestID("tokens")
}
