/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"time"

	"atom-engine/src/core/grpc"
	"atom-engine/src/core/interfaces"
	"atom-engine/src/core/models"
	"atom-engine/src/core/types"
	"atom-engine/src/process"
)

// GetProcessComponent returns process component for gRPC
// Возвращает process компонент для gRPC
func (c *Core) GetProcessComponent() grpc.ProcessComponentInterface {
	if c.processComp == nil {
		return nil
	}
	return &processComponentAdapter{comp: c.processComp}
}

// processComponentAdapter adapts process component to gRPC interface
// Адаптирует process компонент к gRPC интерфейсу
type processComponentAdapter struct {
	comp *process.Component
}

// StartProcessInstance starts new process instance
// Запускает новый экземпляр процесса
func (a *processComponentAdapter) StartProcessInstance(processKey string, variables map[string]interface{}) (*interfaces.ProcessInstanceResult, error) {
	instance, err := a.comp.StartProcessInstance(processKey, variables)
	if err != nil {
		return nil, err
	}

	return &grpc.ProcessInstanceResult{
		InstanceID:  instance.InstanceID,
		ProcessID:   instance.ProcessID,
		ProcessName: instance.ProcessName,
		State:       string(instance.State),
		StartedAt:   instance.StartedAt.Unix(),
		Variables:   instance.Variables,
	}, nil
}

// GetProcessInstanceStatus gets process instance status
// Получает статус экземпляра процесса
func (a *processComponentAdapter) GetProcessInstanceStatus(instanceID string) (*interfaces.ProcessInstanceStatus, error) {
	instance, err := a.comp.GetProcessInstanceStatus(instanceID)
	if err != nil {
		return nil, err
	}

	var completedAtStr string
	if instance.CompletedAt != nil {
		completedAtStr = instance.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &interfaces.ProcessInstanceStatus{
		InstanceID:      instance.InstanceID,
		ProcessID:       instance.ProcessID,
		ProcessName:     instance.ProcessName,
		Status:          string(instance.State),
		State:           string(instance.State),
		CurrentActivity: instance.CurrentActivity,
		StartedAt:       instance.StartedAt.Unix(),
		UpdatedAt:       instance.UpdatedAt.Unix(),
		CompletedAt:     completedAtStr,
		Variables:       instance.Variables,
		CreatedAt:       instance.StartedAt.Format("2006-01-02T15:04:05Z07:00"), // Use StartedAt as CreatedAt
	}, nil
}

// CancelProcessInstance cancels process instance
// Отменяет экземпляр процесса
func (a *processComponentAdapter) CancelProcessInstance(instanceID string, reason string) error {
	return a.comp.CancelProcessInstance(instanceID, reason)
}

// ListProcessInstances lists process instances with optional filters
// Получает список экземпляров процессов с опциональными фильтрами
func (a *processComponentAdapter) ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*interfaces.ProcessInstanceStatus, error) {
	instances, err := a.comp.ListProcessInstances(statusFilter, processKeyFilter, limit)
	if err != nil {
		return nil, err
	}

	var results []*interfaces.ProcessInstanceStatus
	for _, instance := range instances {
		var completedAtStr string
		if instance.CompletedAt != nil {
			completedAtStr = instance.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		result := &interfaces.ProcessInstanceStatus{
			InstanceID:      instance.InstanceID,
			ProcessID:       instance.ProcessID,
			ProcessName:     instance.ProcessName,
			Status:          string(instance.State),
			State:           string(instance.State),
			CurrentActivity: instance.CurrentActivity,
			StartedAt:       instance.StartedAt.Unix(),
			UpdatedAt:       instance.UpdatedAt.Unix(),
			CompletedAt:     completedAtStr,
			Variables:       instance.Variables,
			CreatedAt:       instance.StartedAt.Format("2006-01-02T15:04:05Z07:00"), // Use StartedAt as CreatedAt
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTokensByProcessInstance gets tokens for process instance
// Получает токены для экземпляра процесса
func (a *processComponentAdapter) GetTokensByProcessInstance(instanceID string) ([]*models.Token, error) {
	return a.comp.GetTokensByProcessInstance(instanceID)
}

// GetActiveTokens gets active tokens for process instance
// Получает активные токены для экземпляра процесса
func (a *processComponentAdapter) GetActiveTokens(instanceID string) ([]*models.Token, error) {
	return a.comp.GetActiveTokens(instanceID)
}

// ProcessComponentTypedInterface implementation
// Реализация ProcessComponentTypedInterface

// StartProcessInstanceTyped starts process instance with typed request
// Запускает экземпляр процесса с типизированным запросом
func (a *processComponentAdapter) StartProcessInstanceTyped(processKey string, variables types.ProcessVariables) (*types.ProcessInstanceDetails, error) {
	// Convert typed variables to legacy format
	legacyVars := make(map[string]interface{})
	for k, v := range variables {
		legacyVars[k] = v
	}

	instance, err := a.comp.StartProcessInstance(processKey, legacyVars)
	if err != nil {
		return nil, err
	}

	// Convert to typed response
	now := time.Now()
	var duration *time.Duration
	if instance.CompletedAt != nil {
		d := instance.CompletedAt.Sub(instance.StartedAt)
		duration = &d
	}

	return &types.ProcessInstanceDetails{
		InstanceID:          instance.InstanceID,
		ProcessKey:          processKey, // Use the key from request
		ProcessDefinitionID: instance.ProcessID,
		Version:             1,                         // TODO: get from instance when available
		Status:              types.ProcessStatusActive, // Convert from instance.State
		Variables:           variables,
		StartedAt:           instance.StartedAt,
		UpdatedAt:           now,
		CompletedAt:         instance.CompletedAt,
		Duration:            duration,
		CurrentActivity:     instance.CurrentActivity,
		ActiveTokens:        0,  // TODO: calculate from tokens
		CompletedTokens:     0,  // TODO: calculate from tokens
		ErrorMessage:        "", // TODO: get from instance errors
		Metadata: map[string]interface{}{
			"original_variables": legacyVars,
		},
	}, nil
}

// GetProcessInstanceStatusTyped gets process instance status with typed response
// Получает статус экземпляра процесса с типизированным ответом
func (a *processComponentAdapter) GetProcessInstanceStatusTyped(instanceID string) (*types.ProcessInstanceDetails, error) {
	instance, err := a.comp.GetProcessInstanceStatus(instanceID)
	if err != nil {
		return nil, err
	}

	// Convert variables to typed format
	variables := make(types.ProcessVariables)
	for k, v := range instance.Variables {
		variables[k] = v
	}

	// Get active tokens count
	activeTokens, err := a.comp.GetActiveTokens(instanceID)
	if err != nil {
		activeTokens = []*models.Token{} // Default to empty
	}

	var duration *time.Duration
	if instance.CompletedAt != nil {
		d := instance.CompletedAt.Sub(instance.StartedAt)
		duration = &d
	}

	return &types.ProcessInstanceDetails{
		InstanceID:          instance.InstanceID,
		ProcessKey:          "", // TODO: get from instance
		ProcessDefinitionID: instance.ProcessID,
		Version:             1, // TODO: get from instance
		Status:              types.ProcessStatus(instance.State),
		Variables:           variables,
		StartedAt:           instance.StartedAt,
		UpdatedAt:           instance.UpdatedAt,
		CompletedAt:         instance.CompletedAt,
		Duration:            duration,
		CurrentActivity:     instance.CurrentActivity,
		ActiveTokens:        int32(len(activeTokens)),
		CompletedTokens:     0,  // TODO: calculate completed tokens
		ErrorMessage:        "", // TODO: get error message
		Metadata: map[string]interface{}{
			"legacy_state": instance.State,
		},
	}, nil
}

// ListProcessInstancesTyped lists process instances with typed request/response
// Получает список экземпляров процессов с типизированным запросом/ответом
func (a *processComponentAdapter) ListProcessInstancesTyped(req *types.ProcessListRequest) (*types.ProcessListResponse, error) {
	// Convert typed request to legacy parameters
	statusFilter := ""
	if req.Status != nil {
		statusFilter = string(*req.Status)
	}

	processKeyFilter := ""
	if req.ProcessKey != nil {
		processKeyFilter = *req.ProcessKey
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	instances, err := a.comp.ListProcessInstances(statusFilter, processKeyFilter, limit)
	if err != nil {
		return nil, err
	}

	// Convert to typed response
	typedInstances := make([]types.ProcessInstanceDetails, 0, len(instances))
	for _, instance := range instances {
		// Convert variables
		variables := make(types.ProcessVariables)
		for k, v := range instance.Variables {
			variables[k] = v
		}

		var duration *time.Duration
		if instance.CompletedAt != nil {
			d := instance.CompletedAt.Sub(instance.StartedAt)
			duration = &d
		}

		typedInstance := types.ProcessInstanceDetails{
			InstanceID:          instance.InstanceID,
			ProcessKey:          "", // TODO: get from instance
			ProcessDefinitionID: instance.ProcessID,
			Version:             1, // TODO: get from instance
			Status:              types.ProcessStatus(instance.State),
			Variables:           variables,
			StartedAt:           instance.StartedAt,
			UpdatedAt:           instance.UpdatedAt,
			CompletedAt:         instance.CompletedAt,
			Duration:            duration,
			CurrentActivity:     instance.CurrentActivity,
			ActiveTokens:        0,  // TODO: calculate
			CompletedTokens:     0,  // TODO: calculate
			ErrorMessage:        "", // TODO: get error message
		}
		typedInstances = append(typedInstances, typedInstance)
	}

	return &types.ProcessListResponse{
		Instances:  typedInstances,
		TotalCount: int32(len(typedInstances)),
		HasMore:    false, // TODO: implement proper pagination
	}, nil
}

// GetProcessStats returns process statistics
// Возвращает статистику процессов
func (a *processComponentAdapter) GetProcessStats() (*types.ProcessStats, error) {
	// TODO: Implement proper stats gathering from process component
	return &types.ProcessStats{
		TotalInstances:        0,
		ActiveInstances:       0,
		CompletedInstances:    0,
		CancelledInstances:    0,
		FailedInstances:       0,
		SuspendedInstances:    0,
		TotalDefinitions:      0,
		ActiveDefinitions:     0,
		InstancesByProcess:    make(map[string]int64),
		InstancesByStatus:     make(map[types.ProcessStatus]int64),
		InstancesByTenant:     make(map[string]int64),
		AverageExecutionTime:  0,
		TotalTokens:           0,
		ActiveTokens:          0,
		CompletedTokens:       0,
		LastInstanceStarted:   nil,
		LastInstanceCompleted: nil,
	}, nil
}

// GetTokensTyped returns tokens with typed request/response
// Возвращает токены с типизированным запросом/ответом
func (a *processComponentAdapter) GetTokensTyped(req *types.TokenListRequest) (*types.TokenListResponse, error) {
	// Get tokens based on request
	var tokens []*models.Token
	var err error

	if req.ProcessInstanceID != nil && *req.ProcessInstanceID != "" {
		if req.ActiveOnly {
			tokens, err = a.comp.GetActiveTokens(*req.ProcessInstanceID)
		} else {
			tokens, err = a.comp.GetTokensByProcessInstance(*req.ProcessInstanceID)
		}
	} else {
		// TODO: implement getting all tokens when ProcessInstanceID is empty
		tokens = []*models.Token{}
	}

	if err != nil {
		return nil, err
	}

	// Convert to typed response
	tokenInfos := make([]types.TokenInfo, 0, len(tokens))
	for _, token := range tokens {
		tokenInfo := types.TokenInfo{
			TokenID:           token.TokenID,
			ProcessInstanceID: token.ProcessInstanceID,
			ElementID:         token.CurrentElementID,
			Status:            string(token.State),
			Variables:         types.ProcessVariables(token.Variables),
			CreatedAt:         token.CreatedAt,
			UpdatedAt:         token.UpdatedAt,
			CompletedAt:       token.CompletedAt,
			ParentTokenID:     token.ParentTokenID,
			IsActive:          token.State == models.TokenStateActive,
		}
		tokenInfos = append(tokenInfos, tokenInfo)
	}

	return &types.TokenListResponse{
		Tokens:     tokenInfos,
		TotalCount: int32(len(tokenInfos)),
		HasMore:    false, // TODO: implement proper pagination
	}, nil
}

// TraceProcessExecution returns process execution trace
// Возвращает трассировку выполнения процесса
func (a *processComponentAdapter) TraceProcessExecution(req *types.ProcessTraceRequest) (*types.ProcessTraceResponse, error) {
	// Get all tokens for the process instance
	tokens, err := a.comp.GetTokensByProcessInstance(req.ProcessInstanceID)
	if err != nil {
		return nil, err
	}

	// Convert tokens to TokenInfo for trace
	tokenInfos := make([]types.TokenInfo, 0, len(tokens))
	for _, token := range tokens {
		tokenInfo := types.TokenInfo{
			TokenID:           token.TokenID,
			ProcessInstanceID: token.ProcessInstanceID,
			ElementID:         token.CurrentElementID,
			Status:            string(token.State),
			Variables:         types.ProcessVariables(token.Variables),
			CreatedAt:         token.CreatedAt,
			UpdatedAt:         token.UpdatedAt,
			CompletedAt:       token.CompletedAt,
			ParentTokenID:     token.ParentTokenID,
			IsActive:          token.State == models.TokenStateActive,
		}
		tokenInfos = append(tokenInfos, tokenInfo)
	}

	// Build execution path from tokens
	executionPath := make([]string, 0)
	for _, token := range tokens {
		executionPath = append(executionPath, token.CurrentElementID)
	}

	var totalDuration *time.Duration
	if len(tokens) > 0 {
		// Find the earliest and latest timestamps
		start := tokens[0].CreatedAt
		var end *time.Time

		for _, token := range tokens {
			if token.CreatedAt.Before(start) {
				start = token.CreatedAt
			}
			if token.CompletedAt != nil && (end == nil || token.CompletedAt.After(*end)) {
				end = token.CompletedAt
			}
		}

		if end != nil {
			d := end.Sub(start)
			totalDuration = &d
		}
	}

	return &types.ProcessTraceResponse{
		ProcessInstanceID: req.ProcessInstanceID,
		ProcessKey:        "",                        // TODO: get from instance
		Status:            types.ProcessStatusActive, // TODO: get real status
		Tokens:            tokenInfos,
		ExecutionPath:     executionPath,
		StartedAt:         time.Now(), // TODO: get real start time
		CompletedAt:       nil,        // TODO: get completion time if completed
		Duration:          totalDuration,
		TotalTokens:       int32(len(tokenInfos)),
		CompletedTokens:   0, // TODO: count completed tokens
	}, nil
}
