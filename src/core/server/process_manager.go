/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"
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
func (a *processComponentAdapter) StartProcessInstance(
	processKey string,
	variables map[string]interface{},
) (*interfaces.ProcessInstanceResult, error) {
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
func (a *processComponentAdapter) GetProcessInstanceStatus(
	instanceID string,
) (*interfaces.ProcessInstanceStatus, error) {
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
func (a *processComponentAdapter) ListProcessInstances(
	statusFilter string,
	processKeyFilter string,
	limit int,
) ([]*interfaces.ProcessInstanceStatus, error) {
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
func (a *processComponentAdapter) StartProcessInstanceTyped(
	processKey string,
	variables types.ProcessVariables,
) (*types.ProcessInstanceDetails, error) {
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
	// Get active tokens for this instance
	activeTokens, err := a.comp.GetActiveTokens(instance.InstanceID)
	if err != nil {
		activeTokens = []*models.Token{} // Default to empty
	}

	// Get all tokens for completed count
	allTokens, err := a.comp.GetTokensByProcessInstance(instance.InstanceID)
	if err != nil {
		allTokens = []*models.Token{} // Default to empty
	}

	// Count completed tokens
	completedCount := 0
	for _, token := range allTokens {
		if token.IsCompleted() {
			completedCount++
		}
	}

	var duration *time.Duration
	if instance.CompletedAt != nil {
		d := instance.CompletedAt.Sub(instance.StartedAt)
		duration = &d
	}

	return &types.ProcessInstanceDetails{
		InstanceID:          instance.InstanceID,
		ProcessKey:          instance.ProcessKey, // Use actual process key from instance
		ProcessDefinitionID: instance.ProcessID,
		Version:             int32(instance.ProcessVersion), // Use actual version from instance
		Status:              types.ProcessStatus(instance.State),
		Variables:           variables,
		StartedAt:           instance.StartedAt,
		UpdatedAt:           now,
		CompletedAt:         instance.CompletedAt,
		Duration:            duration,
		CurrentActivity:     instance.CurrentActivity,
		ActiveTokens:        int32(len(activeTokens)),
		CompletedTokens:     int32(completedCount),
		ErrorMessage:        "", // Could extract from instance metadata if available
		Metadata: map[string]interface{}{
			"original_variables": legacyVars,
			"process_key":        instance.ProcessKey,
		},
	}, nil
}

// GetProcessInstanceStatusTyped gets process instance status with typed response
// Получает статус экземпляра процесса с типизированным ответом
func (a *processComponentAdapter) GetProcessInstanceStatusTyped(
	instanceID string,
) (*types.ProcessInstanceDetails, error) {
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

	// Get all tokens for completed count
	allTokens, err := a.comp.GetTokensByProcessInstance(instanceID)
	if err != nil {
		allTokens = []*models.Token{} // Default to empty
	}

	// Count completed tokens
	completedCount := 0
	for _, token := range allTokens {
		if token.IsCompleted() {
			completedCount++
		}
	}

	var duration *time.Duration
	if instance.CompletedAt != nil {
		d := instance.CompletedAt.Sub(instance.StartedAt)
		duration = &d
	}

	return &types.ProcessInstanceDetails{
		InstanceID:          instance.InstanceID,
		ProcessKey:          instance.ProcessKey, // Use actual process key from instance
		ProcessDefinitionID: instance.ProcessID,
		Version:             int32(instance.ProcessVersion), // Use actual version from instance
		Status:              types.ProcessStatus(instance.State),
		Variables:           variables,
		StartedAt:           instance.StartedAt,
		UpdatedAt:           instance.UpdatedAt,
		CompletedAt:         instance.CompletedAt,
		Duration:            duration,
		CurrentActivity:     instance.CurrentActivity,
		ActiveTokens:        int32(len(activeTokens)),
		CompletedTokens:     int32(completedCount), // Use real completed tokens count
		ErrorMessage:        "",                    // Could extract from instance metadata if available
		Metadata: map[string]interface{}{
			"legacy_state": instance.State,
			"process_key":  instance.ProcessKey,
		},
	}, nil
}

// ListProcessInstancesTyped lists process instances with typed request/response
// Получает список экземпляров процессов с типизированным запросом/ответом
func (a *processComponentAdapter) ListProcessInstancesTyped(
	req *types.ProcessListRequest,
) (*types.ProcessListResponse, error) {
	// Convert typed request to legacy parameters
	statusFilter := ""
	if req.Status != nil {
		statusFilter = string(*req.Status)
	}

	processKeyFilter := ""
	if req.ProcessKey != nil {
		processKeyFilter = *req.ProcessKey
	}

	// Load ALL instances for proper pagination (pass 0 for no limit)
	instances, err := a.comp.ListProcessInstances(statusFilter, processKeyFilter, 0)
	if err != nil {
		return nil, err
	}

	// Store total count before pagination
	totalCount := len(instances)

	// Set pagination defaults
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20 // Default page size
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Apply pagination
	var paginatedInstances []*models.ProcessInstance
	if offset < len(instances) {
		end := offset + limit
		if end > len(instances) {
			end = len(instances)
		}
		paginatedInstances = instances[offset:end]
	}

	// Convert to typed response
	typedInstances := make([]types.ProcessInstanceDetails, 0, len(paginatedInstances))
	for _, instance := range paginatedInstances {
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
			ProcessKey:          instance.ProcessKey, // Use actual process key from instance
			ProcessDefinitionID: instance.ProcessID,
			Version:             int32(instance.ProcessVersion), // Use actual version from instance
			Status:              types.ProcessStatus(instance.State),
			Variables:           variables,
			StartedAt:           instance.StartedAt,
			UpdatedAt:           instance.UpdatedAt,
			CompletedAt:         instance.CompletedAt,
			Duration:            duration,
			CurrentActivity:     instance.CurrentActivity,
			ActiveTokens:        0,  // Not calculated for list performance
			CompletedTokens:     0,  // Not calculated for list performance
			ErrorMessage:        "", // Could extract from instance metadata if available
		}
		typedInstances = append(typedInstances, typedInstance)
	}

	// Calculate if there are more items
	hasMore := offset+limit < totalCount

	return &types.ProcessListResponse{
		Instances:  typedInstances,
		TotalCount: int32(totalCount),
		HasMore:    hasMore,
	}, nil
}

// GetProcessStats returns process statistics
// Возвращает статистику процессов
func (a *processComponentAdapter) GetProcessStats() (*types.ProcessStats, error) {
	// Get real statistics from storage and components
	// Получаем реальную статистику из storage и компонентов

	// Load all process instances for statistics
	instances, err := a.comp.ListProcessInstances("", "", -1) // Get all instances
	if err != nil {
		return nil, fmt.Errorf("failed to load process instances: %w", err)
	}

	// Get token statistics from token manager
	// Since GetTokenStatistics is not exposed in Component interface,
	// we'll get token counts by loading tokens by state
	tokenStats := make(map[string]int)

	// Count tokens by state using existing Component methods
	if allTokens, err := a.comp.GetAllTokens(); err == nil {
		for _, token := range allTokens {
			stateKey := string(token.State)
			tokenStats[stateKey]++
		}
	}

	// Calculate process statistics by status
	var totalInstances, activeInstances, completedInstances, cancelledInstances int64
	instancesByProcess := make(map[string]int64)
	instancesByStatus := make(map[types.ProcessStatus]int64)
	instancesByTenant := make(map[string]int64)

	processKeys := make(map[string]bool)

	for _, instance := range instances {
		totalInstances++

		// Count by process key
		instancesByProcess[instance.ProcessKey]++
		processKeys[instance.ProcessKey] = true

		// Count by tenant (using process key as tenant for now)
		instancesByTenant[instance.ProcessKey]++

		// Count by status
		switch instance.State {
		case models.ProcessInstanceStateActive:
			activeInstances++
			instancesByStatus[types.ProcessStatusActive]++
		case models.ProcessInstanceStateCompleted:
			completedInstances++
			instancesByStatus[types.ProcessStatusCompleted]++
		case models.ProcessInstanceStateCanceled:
			cancelledInstances++
			instancesByStatus[types.ProcessStatusCancelled]++
		case models.ProcessInstanceStateFailed:
			instancesByStatus[types.ProcessStatusFailed]++
		case models.ProcessInstanceStateSuspended:
			instancesByStatus[types.ProcessStatusSuspended]++
		}
	}

	// Extract token counts from token statistics
	var totalTokens, activeTokens, completedTokens int64
	if count, exists := tokenStats[string(models.TokenStateActive)]; exists {
		activeTokens = int64(count)
		totalTokens += int64(count)
	}
	if count, exists := tokenStats[string(models.TokenStateCompleted)]; exists {
		completedTokens = int64(count)
		totalTokens += int64(count)
	}
	if count, exists := tokenStats[string(models.TokenStateCanceled)]; exists {
		totalTokens += int64(count)
	}
	if count, exists := tokenStats[string(models.TokenStateFailed)]; exists {
		totalTokens += int64(count)
	}
	if count, exists := tokenStats[string(models.TokenStateWaiting)]; exists {
		totalTokens += int64(count)
	}

	return &types.ProcessStats{
		TotalInstances:        totalInstances,
		ActiveInstances:       activeInstances,
		CompletedInstances:    completedInstances,
		CancelledInstances:    cancelledInstances,
		FailedInstances:       instancesByStatus[types.ProcessStatusFailed],
		SuspendedInstances:    instancesByStatus[types.ProcessStatusSuspended],
		TotalDefinitions:      int64(len(processKeys)),
		ActiveDefinitions:     int64(len(processKeys)), // Assume all loaded definitions are active
		InstancesByProcess:    instancesByProcess,
		InstancesByStatus:     instancesByStatus,
		InstancesByTenant:     instancesByTenant,
		AverageExecutionTime:  0, // Complex calculation, would need execution time tracking
		TotalTokens:           totalTokens,
		ActiveTokens:          activeTokens,
		CompletedTokens:       completedTokens,
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
		// Get all tokens from storage
		tokens, err = a.comp.GetAllTokens()
	}

	if err != nil {
		return nil, err
	}

	// Store total count before pagination
	totalCount := len(tokens)

	// Set pagination defaults
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default page size for tokens
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Apply pagination
	var paginatedTokens []*models.Token
	if offset < len(tokens) {
		end := offset + limit
		if end > len(tokens) {
			end = len(tokens)
		}
		paginatedTokens = tokens[offset:end]
	}

	// Convert to typed response
	tokenInfos := make([]types.TokenInfo, 0, len(paginatedTokens))
	for _, token := range paginatedTokens {
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

	// Calculate if there are more items
	hasMore := offset+limit < totalCount

	return &types.TokenListResponse{
		Tokens:     tokenInfos,
		TotalCount: int32(totalCount),
		HasMore:    hasMore,
	}, nil
}

// TraceProcessExecution returns process execution trace
// Возвращает трассировку выполнения процесса
func (a *processComponentAdapter) TraceProcessExecution(
	req *types.ProcessTraceRequest,
) (*types.ProcessTraceResponse, error) {
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

	// Get process instance for additional data
	instance, err := a.comp.GetProcessInstanceStatus(req.ProcessInstanceID)
	if err != nil {
		// Return trace without instance data if not found
		return &types.ProcessTraceResponse{
			ProcessInstanceID: req.ProcessInstanceID,
			ProcessKey:        "",
			Status:            types.ProcessStatusActive,
			Tokens:            tokenInfos,
			ExecutionPath:     executionPath,
			StartedAt:         time.Now(),
			CompletedAt:       nil,
			Duration:          totalDuration,
			TotalTokens:       int32(len(tokenInfos)),
			CompletedTokens:   0,
		}, nil
	}

	// Count completed tokens
	completedCount := int32(0)
	for _, token := range tokens {
		if token.IsCompleted() {
			completedCount++
		}
	}

	return &types.ProcessTraceResponse{
		ProcessInstanceID: req.ProcessInstanceID,
		ProcessKey:        instance.ProcessKey,
		Status:            types.ProcessStatus(instance.State),
		Tokens:            tokenInfos,
		ExecutionPath:     executionPath,
		StartedAt:         instance.StartedAt,
		CompletedAt:       instance.CompletedAt,
		Duration:          totalDuration,
		TotalTokens:       int32(len(tokenInfos)),
		CompletedTokens:   completedCount,
	}, nil
}
