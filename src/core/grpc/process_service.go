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
	"sort"
	"strconv"
	"strings"

	"atom-engine/proto/incidents/incidentspb"
	"atom-engine/proto/jobs/jobspb"
	"atom-engine/proto/messages/messagespb"
	"atom-engine/proto/parser/parserpb"
	"atom-engine/proto/process/processpb"
	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/interfaces"
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
		ProcessId:       result.ProcessID,
		ProcessKey:      result.ProcessKey,
		ProcessVersion:  int32(extractVersionFromKey(result.ProcessKey)), // Extract version from ProcessKey
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

// ListProcessInstances lists process instances
// Получает список экземпляров процессов
func (s *processServiceServer) ListProcessInstances(ctx context.Context, req *processpb.ListProcessInstancesRequest) (*processpb.ListProcessInstancesResponse, error) {
	// Set defaults for pagination and sorting parameters
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}
	page := req.Page
	if page <= 0 {
		page = 1 // Default page
	}
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "started_at" // Default sort field
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC" // Default sort order
	}

	logger.Info("ListProcessInstances request",
		logger.String("status_filter", req.StatusFilter),
		logger.String("process_key_filter", req.ProcessKeyFilter),
		logger.Int("limit", int(req.Limit)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder))

	// Get process component
	processComp := s.core.GetProcessComponent()
	if processComp == nil {
		return &processpb.ListProcessInstancesResponse{
			Success: false,
			Message: "process component not available",
		}, nil
	}

	// Call process component (load all for sorting/pagination)
	instances, err := processComp.ListProcessInstances(req.StatusFilter, req.ProcessKeyFilter, 0)
	if err != nil {
		logger.Error("Failed to list process instances", logger.String("error", err.Error()))
		return &processpb.ListProcessInstancesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Store total count before pagination
	totalCount := len(instances)

	// Apply sorting
	sort.Slice(instances, func(i, j int) bool {
		switch sortBy {
		case "started_at":
			if sortOrder == "ASC" {
				return instances[i].StartedAt < instances[j].StartedAt
			}
			return instances[i].StartedAt > instances[j].StartedAt
		case "updated_at":
			if sortOrder == "ASC" {
				return instances[i].UpdatedAt < instances[j].UpdatedAt
			}
			return instances[i].UpdatedAt > instances[j].UpdatedAt
		case "instance_id":
			if sortOrder == "ASC" {
				return instances[i].InstanceID < instances[j].InstanceID
			}
			return instances[i].InstanceID > instances[j].InstanceID
		default:
			// Default to started_at DESC
			return instances[i].StartedAt > instances[j].StartedAt
		}
	})

	// Calculate pagination
	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)
	offset := (int(page) - 1) * int(pageSize)

	// Apply pagination
	var paginatedInstances []*interfaces.ProcessInstanceStatus
	if offset < len(instances) {
		end := offset + int(pageSize)
		if end > len(instances) {
			end = len(instances)
		}
		paginatedInstances = instances[offset:end]
	}

	// Use paginated instances for new pagination system or legacy limit for old system
	if req.PageSize > 0 || (req.PageSize == 0 && req.Limit == 0) {
		// New pagination system (also default when no parameters specified)
		instances = paginatedInstances
	} else if req.Limit > 0 && req.PageSize <= 0 {
		// Legacy limit system for backward compatibility
		if len(instances) > int(req.Limit) {
			instances = instances[:req.Limit]
			totalCount = len(instances)
			totalPages = 1
		}
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

	logger.Info("Process instances listed successfully",
		logger.Int("count", len(protoInstances)),
		logger.Int("total_count", totalCount),
		logger.Int("page", int(page)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("total_pages", totalPages))

	return &processpb.ListProcessInstancesResponse{
		Instances:  protoInstances,
		Success:    true,
		Message:    fmt.Sprintf("found %d process instances (page %d of %d)", len(protoInstances), page, totalPages),
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

// ListTokens lists tokens
// Получает список токенов
func (s *processServiceServer) ListTokens(ctx context.Context, req *processpb.ListTokensRequest) (*processpb.ListTokensResponse, error) {
	// Set defaults for pagination and sorting parameters
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}
	page := req.Page
	if page <= 0 {
		page = 1 // Default page
	}
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at" // Default sort field
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC" // Default sort order
	}

	logger.Info("ListTokens request",
		logger.String("instance_id_filter", req.InstanceIdFilter),
		logger.String("state_filter", req.StateFilter),
		logger.Int("limit", int(req.Limit)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder))

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

	// Get storage component for token loading
	storageComp := s.core.GetStorageComponent()
	if storageComp == nil {
		return &processpb.ListTokensResponse{
			Success: false,
			Message: "storage component not available",
		}, nil
	}

	if req.InstanceIdFilter != "" {
		// Filter by process instance - load ALL tokens for this instance (including FAILED)
		tokens, err = processComp.GetTokensByProcessInstance(req.InstanceIdFilter)
	} else {
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
			case "FAILED":
				state = models.TokenStateFailed
			case "WAITING":
				state = models.TokenStateWaiting
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

	// Apply additional state filter if both instance and state filters are provided
	if req.InstanceIdFilter != "" && req.StateFilter != "" {
		var filteredTokens []*models.Token
		for _, token := range tokens {
			if string(token.State) == req.StateFilter {
				filteredTokens = append(filteredTokens, token)
			}
		}
		tokens = filteredTokens
	}

	if err != nil {
		logger.Error("Failed to load tokens", logger.String("error", err.Error()))
		return &processpb.ListTokensResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Store total count before pagination
	totalCount := len(tokens)

	// Apply sorting
	sort.Slice(tokens, func(i, j int) bool {
		switch sortBy {
		case "created_at":
			if sortOrder == "ASC" {
				return tokens[i].CreatedAt.Before(tokens[j].CreatedAt)
			}
			return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
		case "updated_at":
			if sortOrder == "ASC" {
				return tokens[i].UpdatedAt.Before(tokens[j].UpdatedAt)
			}
			return tokens[i].UpdatedAt.After(tokens[j].UpdatedAt)
		case "token_id":
			if sortOrder == "ASC" {
				return tokens[i].TokenID < tokens[j].TokenID
			}
			return tokens[i].TokenID > tokens[j].TokenID
		default:
			// Default to created_at DESC
			return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
		}
	})

	// Calculate pagination
	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)
	offset := (int(page) - 1) * int(pageSize)

	// Apply pagination
	var paginatedTokens []*models.Token
	if offset < len(tokens) {
		end := offset + int(pageSize)
		if end > len(tokens) {
			end = len(tokens)
		}
		paginatedTokens = tokens[offset:end]
	}

	// Use paginated tokens for new pagination system or legacy limit for old system
	if req.PageSize > 0 || (req.PageSize == 0 && req.Limit == 0) {
		// New pagination system (also default when no parameters specified)
		tokens = paginatedTokens
	} else if req.Limit > 0 && req.PageSize <= 0 {
		// Legacy limit system for backward compatibility
		if len(tokens) > int(req.Limit) {
			tokens = tokens[:req.Limit]
			totalCount = len(tokens)
			totalPages = 1
		}
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

	logger.Info("Tokens listed successfully",
		logger.Int("count", len(protoTokens)),
		logger.Int("total_count", totalCount),
		logger.Int("page", int(page)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("total_pages", totalPages))

	return &processpb.ListTokensResponse{
		Tokens:     protoTokens,
		Success:    true,
		Message:    fmt.Sprintf("found %d tokens (page %d of %d)", len(protoTokens), page, totalPages),
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
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

// GetProcessInstanceInfo gets complete process instance information
// Получает полную информацию об экземпляре процесса
func (s *processServiceServer) GetProcessInstanceInfo(ctx context.Context, req *processpb.GetProcessInstanceInfoRequest) (*processpb.GetProcessInstanceInfoResponse, error) {
	logger.Info("GetProcessInstanceInfo request",
		logger.String("instance_id", req.InstanceId))

	// Get process instance status first
	statusResp, err := s.GetProcessInstanceStatus(ctx, &processpb.GetProcessInstanceStatusRequest{
		InstanceId: req.InstanceId,
	})
	if err != nil {
		logger.Error("Failed to get process instance status",
			logger.String("instance_id", req.InstanceId),
			logger.String("error", err.Error()))
		return &processpb.GetProcessInstanceInfoResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get process status: %v", err),
		}, nil
	}

	// Get tokens for this process instance
	tokensResp, err := s.ListTokens(ctx, &processpb.ListTokensRequest{
		InstanceIdFilter: req.InstanceId,
		PageSize:         1000,
	})
	if err != nil {
		logger.Error("Failed to get tokens for process instance",
			logger.String("instance_id", req.InstanceId),
			logger.String("error", err.Error()))
		tokensResp = &processpb.ListTokensResponse{
			Tokens:  []*processpb.TokenInfo{},
			Success: false,
		}
	}

	// Get process key from status (formatted as "ProcessID:vVersion")
	processKey := req.InstanceId
	if statusResp.ProcessKey != "" {
		processKey = statusResp.ProcessKey
	} else if statusResp.ProcessId != "" && statusResp.ProcessVersion > 0 {
		// Format: "ProcessID:vVersion"
		processKey = fmt.Sprintf("%s:v%d", statusResp.ProcessId, statusResp.ProcessVersion)
	}

	// Find BPMN Process Key by process ID and version from status
	bpmnProcessKey := ""
	if statusResp.ProcessId != "" && statusResp.ProcessVersion > 0 {
		processID := statusResp.ProcessId
		version := fmt.Sprintf("v%d", statusResp.ProcessVersion)

		// Get parser service to find BPMN process
		parserService := &ParserService{core: s.core}
		// Get all BPMN processes and find matching one
		listResp, err := parserService.ListBPMNProcesses(ctx, &parserpb.ListBPMNProcessesRequest{})
		if err == nil && listResp != nil && listResp.Success {
			for _, process := range listResp.Processes {
				if process.ProcessId == processID && process.Version == version {
					bpmnProcessKey = process.ProcessKey
					break
				}
			}
		}
	}

	// Get external services information
	externalServices := &processpb.ExternalServicesInfo{}

	// Get timers for this process instance
	timewheelService := &timewheelServiceServer{core: s.core}
	if timewheelService != nil {
		timersResp, err := timewheelService.ListTimers(ctx, &timewheelpb.ListTimersRequest{
			PageSize: 1000,
		})
		if err == nil && timersResp != nil {
			for _, timer := range timersResp.Timers {
				if timer.ProcessInstanceId == req.InstanceId {
					processTimer := &processpb.ProcessTimerInfo{
						TimerId:          timer.TimerId,
						ElementId:        timer.ElementId,
						TimerType:        timer.TimerType,
						Status:           timer.Status,
						ScheduledAt:      timer.ScheduledAt,
						RemainingSeconds: timer.RemainingSeconds,
						TimeDuration:     timer.TimeDuration,
						TimeCycle:        timer.TimeCycle,
					}
					externalServices.Timers = append(externalServices.Timers, processTimer)
				}
			}
		}
	}

	// Get jobs for this process instance
	jobsService := &jobsServiceServer{core: s.core}
	if jobsService != nil {
		jobsResp, err := jobsService.ListJobs(ctx, &jobspb.ListJobsRequest{
			ProcessInstanceId: req.InstanceId,
			PageSize:          1000,
		})
		if err == nil && jobsResp != nil {
			for _, job := range jobsResp.Jobs {
				processJob := &processpb.ProcessJobInfo{
					Key:          job.Key,
					Type:         job.Type,
					Worker:       job.Worker,
					ElementId:    job.ElementId,
					Status:       job.Status,
					Retries:      job.Retries,
					CreatedAt:    job.CreatedAt,
					ErrorMessage: job.ErrorMessage,
				}
				externalServices.Jobs = append(externalServices.Jobs, processJob)
			}
		}
	}

	// Get message subscriptions
	messagesService := &messagesServiceServer{core: s.core}
	if messagesService != nil {
		subsResp, err := messagesService.ListMessageSubscriptions(ctx, &messagespb.ListMessageSubscriptionsRequest{
			PageSize: 1000,
		})
		if err == nil && subsResp != nil {
			for _, sub := range subsResp.Subscriptions {
				if sub.ProcessDefinitionKey == processKey {
					processSub := &processpb.ProcessMessageSubscriptionInfo{
						Id:             sub.Id,
						MessageName:    sub.MessageName,
						CorrelationKey: sub.CorrelationKey,
						StartEventId:   sub.StartEventId,
						IsActive:       sub.IsActive,
						CreatedAt:      sub.CreatedAt,
					}
					externalServices.MessageSubscriptions = append(externalServices.MessageSubscriptions, processSub)
				}
			}
		}
	}

	// Get buffered messages
	// Reusing messagesService from above
	if messagesService != nil {
		bufferedResp, err := messagesService.ListBufferedMessages(ctx, &messagespb.ListBufferedMessagesRequest{
			PageSize: 1000,
		})
		if err == nil && bufferedResp != nil {
			for _, msg := range bufferedResp.Messages {
				processMsg := &processpb.ProcessBufferedMessageInfo{
					Id:             msg.Id,
					Name:           msg.Name,
					CorrelationKey: msg.CorrelationKey,
					ElementId:      msg.ElementId,
					PublishedAt:    msg.PublishedAt,
					ExpiresAt:      msg.ExpiresAt,
					Reason:         msg.Reason,
				}
				externalServices.BufferedMessages = append(externalServices.BufferedMessages, processMsg)
			}
		}
	}

	// Get incidents for this process instance
	incidentsService := &incidentsServiceServer{core: s.core}
	if incidentsService != nil {
		filter := &incidentspb.IncidentFilter{
			ProcessInstanceId: req.InstanceId,
			PageSize:          1000,
		}
		incidentsResp, err := incidentsService.ListIncidents(ctx, &incidentspb.ListIncidentsRequest{
			Filter: filter,
		})
		if err == nil && incidentsResp != nil {
			for _, incident := range incidentsResp.Incidents {
				processIncident := &processpb.ProcessIncidentInfo{
					Id:        incident.Id,
					Type:      incident.Type.String(),
					Status:    incident.Status.String(),
					Message:   incident.Message,
					ErrorCode: incident.ErrorCode,
					ElementId: incident.ElementId,
					JobKey:    incident.JobKey,
					CreatedAt: incident.CreatedAt.Seconds,
				}
				externalServices.Incidents = append(externalServices.Incidents, processIncident)
			}
		}
	}

	// Build response
	response := &processpb.GetProcessInstanceInfoResponse{
		Success:          true,
		Message:          "process instance information retrieved successfully",
		InstanceId:       statusResp.InstanceId,
		ProcessKey:       processKey,
		BpmnProcessKey:   bpmnProcessKey,
		Status:           statusResp.Status,
		CurrentActivity:  statusResp.CurrentActivity,
		StartedAt:        statusResp.StartedAt,
		UpdatedAt:        statusResp.UpdatedAt,
		Variables:        statusResp.Variables,
		Tokens:           tokensResp.Tokens,
		ExternalServices: externalServices,
	}

	logger.Info("Process instance info retrieved successfully",
		logger.String("instance_id", req.InstanceId),
		logger.String("status", response.Status),
		logger.Int("tokens_count", len(response.Tokens)),
		logger.Int("timers_count", len(externalServices.Timers)),
		logger.Int("jobs_count", len(externalServices.Jobs)),
		logger.Int("incidents_count", len(externalServices.Incidents)),
		logger.Int("subscriptions_count", len(externalServices.MessageSubscriptions)),
		logger.Int("buffered_messages_count", len(externalServices.BufferedMessages)))

	return response, nil
}

// extractVersionFromKey extracts version from process key
func extractVersionFromKey(processKey string) int {
	if strings.Contains(processKey, ":v") {
		parts := strings.Split(processKey, ":v")
		if len(parts) > 1 {
			if version, err := strconv.Atoi(parts[1]); err == nil {
				return version
			}
		}
	}
	return 1
}
