/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"fmt"

	"atom-engine/proto/process/processpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// processServiceServer implements process gRPC service
// Реализация process gRPC сервиса
type processServiceServer struct {
	processpb.UnimplementedProcessServiceServer
	core CoreInterface
}

// StartProcessInstance starts new process instance
// Запускает новый экземпляр процесса
func (s *processServiceServer) StartProcessInstance(ctx context.Context, req *processpb.StartProcessInstanceRequest) (*processpb.StartProcessInstanceResponse, error) {
	logger.Info("=== gRPC StartProcessInstance RECEIVED ===",
		logger.String("process_id", req.ProcessId),
		logger.String("variables_count", fmt.Sprintf("%d", len(req.Variables))))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.StartProcessInstanceResponse{
			Success: false,
			Message: "process component not available",
		}, nil
	}

	// Convert variables from protobuf map to Go map
	variables := make(map[string]interface{})
	for key, value := range req.Variables {
		variables[key] = value
	}

	// Start process instance
	result, err := processComp.StartProcessInstance(req.ProcessId, variables)
	if err != nil {
		logger.Error("Failed to start process instance",
			logger.String("process_id", req.ProcessId),
			logger.String("error", err.Error()))

		return &processpb.StartProcessInstanceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	logger.Info("Process instance started successfully",
		logger.String("instance_id", result.InstanceID),
		logger.String("process_id", req.ProcessId))

	return &processpb.StartProcessInstanceResponse{
		InstanceId: result.InstanceID,
		Status:     result.State,
		Success:    true,
		Message:    "process instance started successfully",
	}, nil
}

// GetProcessInstanceStatus gets process instance status
// Получает статус экземпляра процесса
func (s *processServiceServer) GetProcessInstanceStatus(ctx context.Context, req *processpb.GetProcessInstanceStatusRequest) (*processpb.GetProcessInstanceStatusResponse, error) {
	logger.Info("GetProcessInstanceStatus request",
		logger.String("instance_id", req.InstanceId))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.GetProcessInstanceStatusResponse{},
			fmt.Errorf("process component not available")
	}

	// Get process instance status
	result, err := processComp.GetProcessInstanceStatus(req.InstanceId)
	if err != nil {
		logger.Error("Failed to get process instance status",
			logger.String("instance_id", req.InstanceId),
			logger.String("error", err.Error()))

		return &processpb.GetProcessInstanceStatusResponse{}, err
	}

	// Convert variables to protobuf map
	variables := make(map[string]string)
	for key, value := range result.Variables {
		if strValue, ok := value.(string); ok {
			variables[key] = strValue
		} else {
			variables[key] = fmt.Sprintf("%v", value)
		}
	}

	logger.Info("Process instance status retrieved",
		logger.String("instance_id", req.InstanceId),
		logger.String("status", result.State))

	return &processpb.GetProcessInstanceStatusResponse{
		InstanceId:      result.InstanceID,
		Status:          result.State,
		CurrentActivity: result.CurrentActivity,
		Variables:       variables,
		StartedAt:       result.StartedAt,
		UpdatedAt:       result.UpdatedAt,
	}, nil
}

// CancelProcessInstance cancels process instance
// Отменяет экземпляр процесса
func (s *processServiceServer) CancelProcessInstance(ctx context.Context, req *processpb.CancelProcessInstanceRequest) (*processpb.CancelProcessInstanceResponse, error) {
	logger.Info("CancelProcessInstance request",
		logger.String("instance_id", req.InstanceId),
		logger.String("reason", req.Reason))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.CancelProcessInstanceResponse{
			Success: false,
			Message: "process component not available",
		}, nil
	}

	// Cancel process instance
	err := processComp.CancelProcessInstance(req.InstanceId, req.Reason)
	if err != nil {
		logger.Error("Failed to cancel process instance",
			logger.String("instance_id", req.InstanceId),
			logger.String("error", err.Error()))

		return &processpb.CancelProcessInstanceResponse{
			InstanceId: req.InstanceId,
			Success:    false,
			Message:    err.Error(),
		}, nil
	}

	logger.Info("Process instance canceled successfully",
		logger.String("instance_id", req.InstanceId))

	return &processpb.CancelProcessInstanceResponse{
		InstanceId: req.InstanceId,
		Success:    true,
		Message:    "process instance canceled successfully",
	}, nil
}

// DeployProcess deploys BPMN process (placeholder implementation)
// Разворачивает BPMN процесс (заглушка)
func (s *processServiceServer) DeployProcess(ctx context.Context, req *processpb.DeployProcessRequest) (*processpb.DeployProcessResponse, error) {
	logger.Info("DeployProcess request",
		logger.String("process_name", req.ProcessName))

	// This is a placeholder - in real implementation, this would:
	// 1. Parse BPMN content
	// 2. Validate process
	// 3. Store process definition
	// 4. Return deployment information

	return &processpb.DeployProcessResponse{
		ProcessId: "placeholder_process_id",
		Version:   req.Version,
		Success:   true,
		Message:   "process deployment not implemented yet",
	}, nil
}

// ListProcessInstances lists process instances
// Получает список экземпляров процессов
func (s *processServiceServer) ListProcessInstances(ctx context.Context, req *processpb.ListProcessInstancesRequest) (*processpb.ListProcessInstancesResponse, error) {
	logger.Info("ListProcessInstances request",
		logger.String("status_filter", req.StatusFilter),
		logger.String("process_key_filter", req.ProcessKeyFilter),
		logger.Int("limit", int(req.Limit)))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.ListProcessInstancesResponse{
			Success: false,
			Message: "process component not available",
		}, nil
	}

	// Call process component
	instances, err := processComp.ListProcessInstances(req.StatusFilter, req.ProcessKeyFilter, int(req.Limit))
	if err != nil {
		logger.Error("Failed to list process instances", logger.String("error", err.Error()))
		return &processpb.ListProcessInstancesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Convert to protobuf format
	var protoInstances []*processpb.ProcessInstanceInfo
	for _, instance := range instances {
		// Convert variables map
		variables := make(map[string]string)
		for key, value := range instance.Variables {
			if strValue, ok := value.(string); ok {
				variables[key] = strValue
			} else {
				variables[key] = fmt.Sprintf("%v", value)
			}
		}

		protoInstance := &processpb.ProcessInstanceInfo{
			InstanceId:      instance.InstanceID,
			ProcessKey:      instance.ProcessID,
			Status:          instance.State,
			CurrentActivity: instance.CurrentActivity,
			StartedAt:       instance.StartedAt,
			UpdatedAt:       instance.UpdatedAt,
			Variables:       variables,
		}
		protoInstances = append(protoInstances, protoInstance)
	}

	logger.Info("Process instances listed successfully", logger.Int("count", len(protoInstances)))

	return &processpb.ListProcessInstancesResponse{
		Instances: protoInstances,
		Success:   true,
		Message:   fmt.Sprintf("found %d process instances", len(protoInstances)),
	}, nil
}

// ListTokens lists tokens
// Получает список токенов
func (s *processServiceServer) ListTokens(ctx context.Context, req *processpb.ListTokensRequest) (*processpb.ListTokensResponse, error) {
	logger.Info("ListTokens request",
		logger.String("instance_id_filter", req.InstanceIdFilter),
		logger.String("state_filter", req.StateFilter),
		logger.Int("limit", int(req.Limit)))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.ListTokensResponse{
			Success: false,
			Message: "process component not available",
		}, nil
	}

	// Load tokens based on filters
	var tokens []*models.Token
	var err error

	if req.InstanceIdFilter != "" {
		// Filter by process instance
		tokens, err = processComp.GetActiveTokens(req.InstanceIdFilter)
	} else {
		// Get all tokens via storage component
		storageComp := s.core.GetStorageComponent()
		if storageComp == nil {
			return &processpb.ListTokensResponse{
				Success: false,
				Message: "storage component not available",
			}, nil
		}

		if req.StateFilter != "" {
			// Filter by state
			var state models.TokenState
			switch req.StateFilter {
			case "ACTIVE":
				state = models.TokenStateActive
			case "COMPLETED":
				state = models.TokenStateCompleted
			case "CANCELLED":
				state = models.TokenStateCanceled
			default:
				return &processpb.ListTokensResponse{
					Success: false,
					Message: fmt.Sprintf("invalid state filter: %s", req.StateFilter),
				}, nil
			}
			tokens, err = storageComp.LoadTokensByState(state)
		} else {
			// Load all tokens
			tokens, err = storageComp.LoadAllTokens()
		}
	}

	if err != nil {
		logger.Error("Failed to load tokens", logger.String("error", err.Error()))
		return &processpb.ListTokensResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Apply limit if specified
	if req.Limit > 0 && len(tokens) > int(req.Limit) {
		tokens = tokens[:req.Limit]
	}

	// Convert to protobuf format
	var protoTokens []*processpb.TokenInfo
	for _, token := range tokens {
		// Convert variables map
		variables := make(map[string]string)
		for key, value := range token.Variables {
			if strValue, ok := value.(string); ok {
				variables[key] = strValue
			} else {
				variables[key] = fmt.Sprintf("%v", value)
			}
		}

		protoToken := &processpb.TokenInfo{
			TokenId:           token.TokenID,
			ProcessInstanceId: token.ProcessInstanceID,
			ProcessKey:        token.ProcessKey,
			CurrentElementId:  token.CurrentElementID,
			State:             string(token.State),
			WaitingFor:        token.WaitingFor,
			CreatedAt:         token.CreatedAt.Unix(),
			UpdatedAt:         token.UpdatedAt.Unix(),
			Variables:         variables,
		}
		protoTokens = append(protoTokens, protoToken)
	}

	logger.Info("Tokens listed successfully", logger.Int("count", len(protoTokens)))

	return &processpb.ListTokensResponse{
		Tokens:  protoTokens,
		Success: true,
		Message: fmt.Sprintf("found %d tokens", len(protoTokens)),
	}, nil
}

// GetTokenStatus gets token status
// Получает статус токена
func (s *processServiceServer) GetTokenStatus(ctx context.Context, req *processpb.GetTokenStatusRequest) (*processpb.GetTokenStatusResponse, error) {
	logger.Info("GetTokenStatus request", logger.String("token_id", req.TokenId))

	// Get storage component
	storageComp := s.core.GetStorageComponent()
	if storageComp == nil {
		return &processpb.GetTokenStatusResponse{
			Success: false,
			Message: "storage component not available",
		}, nil
	}

	// Load token
	token, err := storageComp.LoadToken(req.TokenId)
	if err != nil {
		logger.Error("Failed to load token", logger.String("error", err.Error()))
		return &processpb.GetTokenStatusResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Convert variables map
	variables := make(map[string]string)
	for key, value := range token.Variables {
		if strValue, ok := value.(string); ok {
			variables[key] = strValue
		} else {
			variables[key] = fmt.Sprintf("%v", value)
		}
	}

	protoToken := &processpb.TokenInfo{
		TokenId:           token.TokenID,
		ProcessInstanceId: token.ProcessInstanceID,
		ProcessKey:        token.ProcessKey,
		CurrentElementId:  token.CurrentElementID,
		State:             string(token.State),
		WaitingFor:        token.WaitingFor,
		CreatedAt:         token.CreatedAt.Unix(),
		UpdatedAt:         token.UpdatedAt.Unix(),
		Variables:         variables,
	}

	logger.Info("Token status retrieved successfully", logger.String("token_id", req.TokenId))

	return &processpb.GetTokenStatusResponse{
		Token:   protoToken,
		Success: true,
		Message: "token status retrieved successfully",
	}, nil
}
