/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/src/core/logger"
	"atom-engine/src/jobs"
)

// jobsServiceServer implements jobs gRPC service
type jobsServiceServer struct {
	jobspb.UnimplementedJobsServiceServer
	core CoreInterface
}

// getJobsComponent helper function deprecated - use JSON communication instead
// helper функция getJobsComponent устарела - используйте JSON коммуникацию
func getJobsComponent(core CoreInterface) (*jobs.Component, error) {
	componentIf := core.GetJobsComponent()
	if componentIf == nil {
		return nil, fmt.Errorf("jobs component not available")
	}

	component, ok := componentIf.(*jobs.Component)
	if !ok {
		return nil, fmt.Errorf("jobs component type assertion failed")
	}

	return component, nil
}

// CreateJob creates a new job
func (s *jobsServiceServer) CreateJob(
	ctx context.Context,
	req *jobspb.CreateJobRequest,
) (*jobspb.CreateJobResponse, error) {
	logger.Info("CreateJob gRPC request",
		logger.String("type", req.Type),
		logger.String("process_instance_id", req.ProcessInstanceId))

	// Parse variables from JSON string
	variables := make(map[string]interface{})
	if req.Variables != "" {
		if err := json.Unmarshal([]byte(req.Variables), &variables); err != nil {
			logger.Error("Failed to parse variables JSON", logger.String("error", err.Error()))
			return &jobspb.CreateJobResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("invalid variables JSON: %v", err),
			}, nil
		}
	}

	// Create JSON message for jobs component
	payload := jobs.CreateJobPayload{
		JobType:           req.Type,
		ProcessInstanceID: req.ProcessInstanceId,
		ElementID:         req.ElementId,
		Variables:         variables,
	}

	message, err := jobs.CreateJobMessage(payload)
	if err != nil {
		logger.Error("Failed to create job message", logger.String("error", err.Error()))
		return &jobspb.CreateJobResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create job message: %v", err),
		}, nil
	}

	// Send JSON message to jobs component through Core
	if err := s.core.SendMessage("jobs", message); err != nil {
		logger.Error("Failed to send job message", logger.String("error", err.Error()))
		return &jobspb.CreateJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	logger.Info("Job creation request sent successfully")

	// Wait for response from jobs component
	// Ожидаем ответ от компонента jobs
	responseJSON, err := s.core.WaitForJobsResponse(5000) // 5 second timeout
	if err != nil {
		logger.Error("Failed to get jobs response", logger.String("error", err.Error()))
		return &jobspb.CreateJobResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to get jobs response: %v", err),
		}, nil
	}

	// Parse JSON response
	// Парсим JSON ответ
	var jobsResponse jobs.JobResponse
	if err := json.Unmarshal([]byte(responseJSON), &jobsResponse); err != nil {
		logger.Error("Failed to parse jobs response", logger.String("error", err.Error()))
		return &jobspb.CreateJobResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to parse response JSON: %v", err),
		}, nil
	}

	if !jobsResponse.Success {
		return &jobspb.CreateJobResponse{
			Success:      false,
			ErrorMessage: jobsResponse.Error,
		}, nil
	}

	// Extract job key from response
	// Извлекаем job key из ответа
	jobKey := "unknown"
	if resultData, ok := jobsResponse.Result.(map[string]interface{}); ok {
		if key, ok := resultData["job_id"].(string); ok {
			jobKey = key
		}
	}

	return &jobspb.CreateJobResponse{
		Success: true,
		JobKey:  jobKey,
	}, nil
}

// ActivateJobs activates jobs for worker (streaming)
func (s *jobsServiceServer) ActivateJobs(
	req *jobspb.ActivateJobsRequest,
	stream jobspb.JobsService_ActivateJobsServer,
) error {
	logger.Info("ActivateJobs gRPC request",
		logger.String("worker", req.Worker),
		logger.String("type", req.Type),
		logger.Int("max_jobs", int(req.MaxJobsToActivate)))

	// Create JSON message for jobs component
	payload := jobs.ActivateJobsPayload{
		WorkerName: req.Worker,
		JobType:    req.Type,
		MaxJobs:    int(req.MaxJobsToActivate),
		TimeoutMs:  req.Timeout,
	}

	message, err := jobs.CreateActivateJobsMessage(payload)
	if err != nil {
		logger.Error("Failed to create activate jobs message", logger.String("error", err.Error()))
		return fmt.Errorf("failed to create activate jobs message: %w", err)
	}

	// Send JSON message to jobs component through Core
	if err := s.core.SendMessage("jobs", message); err != nil {
		logger.Error("Failed to send activate jobs message", logger.String("error", err.Error()))
		return fmt.Errorf("failed to send activate jobs message: %w", err)
	}

	// Wait for response from jobs component
	// Ожидаем ответ от компонента jobs
	responseJSON, err := s.core.WaitForJobsResponse(2000) // 2 second timeout - reduced from 10 seconds
	if err != nil {
		logger.Error("Failed to get jobs response", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get jobs response: %w", err)
	}

	// Parse JSON response
	// Парсим JSON ответ
	var jobsResponse jobs.JobResponse
	if err := json.Unmarshal([]byte(responseJSON), &jobsResponse); err != nil {
		logger.Error("Failed to parse jobs response", logger.String("error", err.Error()))
		return fmt.Errorf("failed to parse response JSON: %w", err)
	}

	var activatedJobs []jobs.JobInfo
	if !jobsResponse.Success {
		logger.Error("Jobs activation failed", logger.String("error", jobsResponse.Error))
		activatedJobs = []jobs.JobInfo{}
	} else {
		// Extract jobs from response
		if jobsList, ok := jobsResponse.Result.([]interface{}); ok {
			for _, jobData := range jobsList {
				if jobMap, ok := jobData.(map[string]interface{}); ok {
					job := jobs.JobInfo{}
					if key, ok := jobMap["key"].(string); ok {
						job.Key = key
					}
					if jobType, ok := jobMap["type"].(string); ok {
						job.Type = jobType
					}
					if worker, ok := jobMap["worker"].(string); ok {
						job.Worker = worker
					}
					if processInstanceID, ok := jobMap["process_instance_id"].(string); ok {
						job.ProcessInstanceID = processInstanceID
					}
					if variables, ok := jobMap["variables"].(map[string]interface{}); ok {
						job.Variables = variables
					}
					if retries, ok := jobMap["retries"].(float64); ok {
						job.Retries = int(retries)
					}
					activatedJobs = append(activatedJobs, job)
				}
			}
		}
	}

	// Stream activated jobs
	for _, job := range activatedJobs {
		// Convert variables to JSON string
		variablesJSON := ""
		if job.Variables != nil {
			if jsonBytes, err := json.Marshal(job.Variables); err == nil {
				variablesJSON = string(jsonBytes)
			}
		}

		activatedJob := &jobspb.ActivatedJob{
			Key:                job.Key,
			Type:               job.Type,
			ProcessInstanceKey: job.ProcessInstanceID,
			Variables:          variablesJSON,
			Worker:             job.Worker,
			Retries:            int32(job.Retries),
			Deadline:           job.CreatedAt + 30000, // 30 second deadline
		}

		response := &jobspb.ActivateJobsResponse{
			Jobs: []*jobspb.ActivatedJob{activatedJob},
		}

		if err := stream.Send(response); err != nil {
			logger.Error("Failed to send job", logger.String("error", err.Error()))
			return err
		}
	}

	logger.Info("Jobs activated successfully", logger.Int("count", len(activatedJobs)))
	return nil
}

// CompleteJob completes a job
func (s *jobsServiceServer) CompleteJob(
	ctx context.Context,
	req *jobspb.CompleteJobRequest,
) (*jobspb.CompleteJobResponse, error) {
	logger.Info("CompleteJob gRPC request", logger.String("job_key", req.JobKey))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.CompleteJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Parse variables from JSON string
	variables := make(map[string]interface{})
	if req.Variables != "" {
		if err := json.Unmarshal([]byte(req.Variables), &variables); err != nil {
			logger.Error("Failed to parse variables JSON", logger.String("error", err.Error()))
			return &jobspb.CompleteJobResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("invalid variables JSON: %v", err),
			}, nil
		}
	}

	// Complete job through component
	if err := component.CompleteJob(req.JobKey, variables); err != nil {
		logger.Error("Failed to complete job", logger.String("error", err.Error()))
		return &jobspb.CompleteJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	logger.Info("Job completed successfully", logger.String("job_key", req.JobKey))

	return &jobspb.CompleteJobResponse{
		Success: true,
	}, nil
}

// FailJob fails a job
func (s *jobsServiceServer) FailJob(ctx context.Context, req *jobspb.FailJobRequest) (*jobspb.FailJobResponse, error) {
	logger.Info("FailJob gRPC request",
		logger.String("job_key", req.JobKey),
		logger.Int("retries", int(req.Retries)))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.FailJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Fail job through component
	if err := component.FailJob(req.JobKey, int(req.Retries), req.ErrorMessage); err != nil {
		logger.Error("Failed to fail job", logger.String("error", err.Error()))
		return &jobspb.FailJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	logger.Info("Job failed successfully", logger.String("job_key", req.JobKey))

	return &jobspb.FailJobResponse{
		Success: true,
	}, nil
}

// ThrowError throws BPMN error for job
func (s *jobsServiceServer) ThrowError(
	ctx context.Context,
	req *jobspb.ThrowErrorRequest,
) (*jobspb.ThrowErrorResponse, error) {
	logger.Info("ThrowError gRPC request",
		logger.String("job_key", req.JobKey),
		logger.String("error_code", req.ErrorCode),
		logger.String("error_message", req.ErrorMessage))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.ThrowErrorResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Call ThrowError on component
	err = component.ThrowError(req.JobKey, req.ErrorCode, req.ErrorMessage)
	if err != nil {
		logger.Error("Failed to throw error for job", logger.String("error", err.Error()))
		return &jobspb.ThrowErrorResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	logger.Info("Job error thrown successfully",
		logger.String("job_key", req.JobKey),
		logger.String("error_code", req.ErrorCode))

	return &jobspb.ThrowErrorResponse{
		Success: true,
	}, nil
}

// GetJobStats gets job statistics
func (s *jobsServiceServer) GetJobStats(
	ctx context.Context,
	req *jobspb.GetJobStatsRequest,
) (*jobspb.GetJobStatsResponse, error) {
	logger.Info("GetJobStats gRPC request")

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.GetJobStatsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Get job stats from component
	componentStatsIf, err := component.GetJobStats()
	if err != nil {
		logger.Error("Failed to get job stats", logger.String("error", err.Error()))
		return &jobspb.GetJobStatsResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Convert to protobuf format - for now use empty stats until proper integration
	componentStats, ok := componentStatsIf.(*jobs.JobStats)
	var stats *jobspb.JobStats
	if ok {
		stats = &jobspb.JobStats{
			TotalJobs:      componentStats.TotalJobs,
			ActiveJobs:     componentStats.ActiveJobs,
			CompletedJobs:  componentStats.CompletedJobs,
			FailedJobs:     componentStats.FailedJobs,
			ActivatedToday: componentStats.ActivatedToday,
			CompletedToday: componentStats.CompletedToday,
		}
	} else {
		stats = &jobspb.JobStats{
			TotalJobs:      0,
			ActiveJobs:     0,
			CompletedJobs:  0,
			FailedJobs:     0,
			ActivatedToday: 0,
			CompletedToday: 0,
		}
	}

	return &jobspb.GetJobStatsResponse{
		Success: true,
		Stats:   stats,
	}, nil
}

// ListJobs lists jobs
func (s *jobsServiceServer) ListJobs(
	ctx context.Context,
	req *jobspb.ListJobsRequest,
) (*jobspb.ListJobsResponse, error) {
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

	logger.Info("ListJobs gRPC request",
		logger.String("type", req.Type),
		logger.String("worker", req.Worker),
		logger.Int("limit", int(req.Limit)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.ListJobsResponse{
			Jobs:       []*jobspb.JobInfo{},
			TotalCount: 0,
		}, nil
	}

	// List all jobs for sorting/pagination
	jobInfos, total, err := component.ListJobs(req.Type, req.Worker, req.ProcessInstanceId, req.State, 0, 0)
	if err != nil {
		logger.Error("Failed to list jobs", logger.String("error", err.Error()))
		return &jobspb.ListJobsResponse{
			Jobs:       []*jobspb.JobInfo{},
			TotalCount: 0,
		}, nil
	}

	// Store total count before pagination
	totalCount := total

	// Apply sorting
	sort.Slice(jobInfos, func(i, j int) bool {
		switch sortBy {
		case "created_at":
			if sortOrder == "ASC" {
				return jobInfos[i].CreatedAt < jobInfos[j].CreatedAt
			}
			return jobInfos[i].CreatedAt > jobInfos[j].CreatedAt
		case "key":
			if sortOrder == "ASC" {
				return jobInfos[i].Key < jobInfos[j].Key
			}
			return jobInfos[i].Key > jobInfos[j].Key
		case "type":
			if sortOrder == "ASC" {
				return jobInfos[i].Type < jobInfos[j].Type
			}
			return jobInfos[i].Type > jobInfos[j].Type
		default:
			// Default to created_at DESC
			return jobInfos[i].CreatedAt > jobInfos[j].CreatedAt
		}
	})

	// Calculate pagination
	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)
	offset := (int(page) - 1) * int(pageSize)

	// Apply pagination
	var paginatedJobs []jobs.JobInfo
	if offset < len(jobInfos) {
		end := offset + int(pageSize)
		if end > len(jobInfos) {
			end = len(jobInfos)
		}
		paginatedJobs = jobInfos[offset:end]
	}

	// Use paginated jobs for new pagination system or legacy limit for old system
	if req.PageSize > 0 || (req.PageSize == 0 && req.Limit == 0) {
		// New pagination system (also default when no parameters specified)
		jobInfos = paginatedJobs
	} else if req.Limit > 0 && req.PageSize <= 0 {
		// Legacy limit system for backward compatibility
		if len(jobInfos) > int(req.Limit) {
			jobInfos = jobInfos[:req.Limit]
			totalCount = len(jobInfos)
			totalPages = 1
		}
	}

	// Convert to protobuf format
	protoJobs := make([]*jobspb.JobInfo, len(jobInfos))
	for i, job := range jobInfos {
		// Convert variables to protobuf map format
		variables := make(map[string]string)
		for k, v := range job.Variables {
			if str, ok := v.(string); ok {
				variables[k] = str
			} else {
				variables[k] = fmt.Sprintf("%v", v)
			}
		}

		protoJobs[i] = &jobspb.JobInfo{
			Key:                job.Key,
			Type:               job.Type,
			ProcessInstanceKey: job.ProcessInstanceID,
			Variables:          variables,
			Worker:             job.Worker,
			Retries:            int32(job.Retries),
			CreatedAt:          job.CreatedAt,
			Status:             job.Status,
			ErrorMessage:       job.ErrorMessage,
		}
	}

	logger.Info("Jobs listed successfully",
		logger.Int("count", len(protoJobs)),
		logger.Int("total_count", totalCount),
		logger.Int("page", int(page)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("total_pages", totalPages))

	return &jobspb.ListJobsResponse{
		Jobs:       protoJobs,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}

// GetJob gets job details
func (s *jobsServiceServer) GetJob(
	ctx context.Context,
	req *jobspb.GetJobRequest,
) (*jobspb.GetJobResponse, error) {
	logger.Info("GetJob gRPC request", logger.String("job_key", req.JobKey))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.GetJobResponse{
			Job:   nil,
			Found: false,
		}, nil
	}

	// Get job through component
	jobInfo, err := component.GetJob(req.JobKey)
	if err != nil {
		logger.Error("Failed to get job", logger.String("error", err.Error()))
		return &jobspb.GetJobResponse{
			Job:   nil,
			Found: false,
		}, nil
	}

	if jobInfo == nil {
		logger.Info("Job not found", logger.String("job_key", req.JobKey))
		return &jobspb.GetJobResponse{
			Job:   nil,
			Found: false,
		}, nil
	}

	// Convert variables to protobuf map format
	variables := make(map[string]string)
	for k, v := range jobInfo.Variables {
		if str, ok := v.(string); ok {
			variables[k] = str
		} else {
			variables[k] = fmt.Sprintf("%v", v)
		}
	}

	// Convert to protobuf format
	protoJob := &jobspb.JobInfo{
		Key:                jobInfo.Key,
		Type:               jobInfo.Type,
		ProcessInstanceKey: jobInfo.ProcessInstanceID,
		Variables:          variables,
		Worker:             jobInfo.Worker,
		Retries:            int32(jobInfo.Retries),
		CreatedAt:          jobInfo.CreatedAt,
		Status:             jobInfo.Status,
		ErrorMessage:       jobInfo.ErrorMessage,
	}

	logger.Info("Job found successfully", logger.String("job_key", req.JobKey))

	return &jobspb.GetJobResponse{
		Job:   protoJob,
		Found: true,
	}, nil
}

// CancelJob cancels a job
func (s *jobsServiceServer) CancelJob(
	ctx context.Context,
	req *jobspb.CancelJobRequest,
) (*jobspb.CancelJobResponse, error) {
	logger.Info("CancelJob gRPC request",
		logger.String("job_key", req.JobKey),
		logger.String("reason", req.Reason))

	// Get jobs component from core
	component, err := getJobsComponent(s.core)
	if err != nil {
		return &jobspb.CancelJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Cancel job through component
	if err := component.CancelJob(req.JobKey, req.Reason); err != nil {
		logger.Error("Failed to cancel job", logger.String("error", err.Error()))
		return &jobspb.CancelJobResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	logger.Info("Job canceled successfully", logger.String("job_key", req.JobKey))

	return &jobspb.CancelJobResponse{
		Success: true,
	}, nil
}
