/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/src/core/config"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// Component handles job management operations
type Component struct {
	config          *config.Config
	logger          logger.ComponentLogger
	storage         storage.Storage
	manager         *JobManager
	isRunning       bool
	responseChannel chan string
}

// NewComponent creates new jobs component
func NewComponent(cfg *config.Config, storage storage.Storage) *Component {
	comp := &Component{
		config:          cfg,
		logger:          logger.NewComponentLogger("jobs"),
		storage:         storage,
		responseChannel: make(chan string, 100), // Buffered channel for job callbacks
	}
	comp.manager = NewJobManager(storage, logger.NewComponentLogger("job-manager"), comp)
	return comp
}

// Start initializes and starts the jobs component
func Start(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	storageConfig := &storage.Config{
		Path: cfg.Database.Path,
	}
	storageInstance := storage.NewStorage(storageConfig)

	component := NewComponent(cfg, storageInstance)
	return component.Start()
}

// Start starts the component
func (c *Component) Start() error {
	c.logger.Info("Starting jobs component")

	// Start job manager
	if err := c.manager.Start(); err != nil {
		return fmt.Errorf("failed to start job manager: %w", err)
	}

	c.isRunning = true
	c.logger.Info("Jobs component started successfully")

	// Send ready status to stdout for core
	fmt.Println(`{"status": "ready", "component": "jobs"}`)

	return nil
}

// Stop stops the component
func (c *Component) Stop() error {
	c.logger.Info("Stopping jobs component")

	c.isRunning = false

	// Stop job manager
	if c.manager != nil {
		c.manager.Stop()
	}

	c.logger.Info("Jobs component stopped")
	return nil
}

// IsReady returns component readiness status
func (c *Component) IsReady() bool {
	return c.isRunning && c.manager != nil && c.manager.IsRunning()
}

// CreateJob creates a new job
func (c *Component) CreateJob(jobType, processInstanceID string, variables map[string]interface{}) (string, error) {
	return c.CreateJobWithDetails(jobType, processInstanceID, "", nil, variables)
}

// CreateJobWithDetails creates a new job with custom headers and element ID
func (c *Component) CreateJobWithDetails(jobType, processInstanceID, elementID string, customHeaders map[string]string, variables map[string]interface{}) (string, error) {
	c.logger.Info("Creating job",
		logger.String("type", jobType),
		logger.String("processInstanceId", processInstanceID),
		logger.String("elementId", elementID))

	// Extract token ID from variables if available
	var tokenID string
	if variables != nil {
		if tid, ok := variables["_tokenID"].(string); ok {
			tokenID = tid
		}
	}

	// Create job model
	job := &models.Job{
		ID:                models.GenerateID(),
		Type:              jobType,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		TokenID:           tokenID,
		CustomHeaders:     customHeaders,
		Variables:         variables,
		Status:            models.JobStatusPending,
		Retries:           3,
		MaxRetries:        3,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if customHeaders == nil {
		job.CustomHeaders = make(map[string]string)
	}

	// Delegate to job manager
	if err := c.manager.CreateJob(context.Background(), job); err != nil {
		return "", err
	}

	c.logger.Info("Job created successfully",
		logger.String("jobId", job.ID),
		logger.String("type", jobType),
		logger.String("elementId", elementID))

	return job.ID, nil
}

// ActivateJobs activates jobs for worker
func (c *Component) ActivateJobs(workerName, jobType string, maxJobs int) ([]JobInfo, error) {
	c.logger.Info("Activating jobs", logger.String("worker", workerName), logger.String("type", jobType), logger.Int("maxJobs", maxJobs))

	// Delegate to job manager
	timeout := 30 * time.Second
	jobs, err := c.manager.ActivateJobs(context.Background(), jobType, workerName, maxJobs, timeout)
	if err != nil {
		return nil, err
	}

	// Convert to JobInfo
	jobInfos := make([]JobInfo, len(jobs))
	for i, job := range jobs {
		jobInfos[i] = JobInfo{
			Key:               job.ID,
			Type:              job.Type,
			ProcessInstanceID: job.ProcessInstanceID,
			Variables:         job.Variables,
			Worker:            job.WorkerID,
			Retries:           job.Retries,
			CreatedAt:         job.CreatedAt.Unix(),
		}
	}

	return jobInfos, nil
}

// ActivateJobsWithTimeout activates jobs for worker with custom timeout
func (c *Component) ActivateJobsWithTimeout(workerName, jobType string, maxJobs int, timeoutMs int32) ([]JobInfo, error) {
	c.logger.Info("Activating jobs with timeout",
		logger.String("worker", workerName),
		logger.String("type", jobType),
		logger.Int("maxJobs", maxJobs),
		logger.Int("timeoutMs", int(timeoutMs)))

	// Convert milliseconds to time.Duration
	timeout := time.Duration(timeoutMs) * time.Millisecond
	jobs, err := c.manager.ActivateJobs(context.Background(), jobType, workerName, maxJobs, timeout)
	if err != nil {
		return nil, err
	}

	// Convert to JobInfo
	jobInfos := make([]JobInfo, len(jobs))
	for i, job := range jobs {
		jobInfos[i] = JobInfo{
			Key:               job.ID,
			Type:              job.Type,
			ProcessInstanceID: job.ProcessInstanceID,
			Variables:         job.Variables,
			Worker:            job.WorkerID,
			Retries:           job.Retries,
			CreatedAt:         job.CreatedAt.Unix(),
		}
	}

	return jobInfos, nil
}

// CompleteJob completes a job
func (c *Component) CompleteJob(jobKey string, variables map[string]interface{}) error {
	c.logger.Info("Completing job", logger.String("jobKey", jobKey))

	// Delegate to job manager
	return c.manager.CompleteJob(context.Background(), jobKey, variables)
}

// FailJob fails a job
func (c *Component) FailJob(jobKey string, retries int, errorMessage string) error {
	c.logger.Info("Failing job", logger.String("jobKey", jobKey), logger.Int("retries", retries))

	// Delegate to job manager
	retryBackoff := 5 * time.Second
	return c.manager.FailJob(context.Background(), jobKey, retries, errorMessage, retryBackoff)
}

// GetJobStats returns job statistics
func (c *Component) GetJobStats() (interface{}, error) {
	c.logger.Debug("Getting job stats")

	// For now return empty stats - would need implementation in JobManager
	return &JobStats{
		TotalJobs:      0,
		ActiveJobs:     0,
		CompletedJobs:  0,
		FailedJobs:     0,
		ActivatedToday: 0,
		CompletedToday: 0,
	}, nil
}

// ListJobs lists jobs with filtering
func (c *Component) ListJobs(jobType, worker, processInstanceID, state string, limit, offset int) ([]JobInfo, int, error) {
	c.logger.Debug("Listing jobs", logger.String("type", jobType), logger.String("worker", worker))

	// Create filter - ListJobsFilter is defined in manager.go
	filter := &ListJobsFilter{
		Type:              jobType,
		Worker:            worker,
		ProcessInstanceID: processInstanceID,
		State:             state,
		Limit:             limit,
		Offset:            offset,
		IncludeVariables:  true,
	}

	// Delegate to job manager
	jobs, total, err := c.manager.ListJobs(context.Background(), filter)
	if err != nil {
		return nil, 0, err
	}

	// Convert to JobInfo
	jobInfos := make([]JobInfo, len(jobs))
	for i, job := range jobs {
		jobInfos[i] = JobInfo{
			Key:               job.ID,
			Type:              job.Type,
			ProcessInstanceID: job.ProcessInstanceID,
			Variables:         job.Variables,
			Worker:            job.WorkerID,
			Retries:           job.Retries,
			CreatedAt:         job.CreatedAt.Unix(),
			Status:            string(job.Status),
			ErrorMessage:      job.ErrorMessage,
		}
	}

	return jobInfos, total, nil
}

// GetJob gets job by ID
func (c *Component) GetJob(jobID string) (*JobInfo, error) {
	c.logger.Debug("Getting job", logger.String("jobID", jobID))

	// Delegate to job manager
	job, err := c.manager.GetJob(context.Background(), jobID)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return nil, nil // Job not found
	}

	// Convert to JobInfo
	jobInfo := &JobInfo{
		Key:               job.ID,
		Type:              job.Type,
		ProcessInstanceID: job.ProcessInstanceID,
		Variables:         job.Variables,
		Worker:            job.WorkerID,
		Retries:           job.Retries,
		CreatedAt:         job.CreatedAt.Unix(),
		Status:            string(job.Status),
		ErrorMessage:      job.ErrorMessage,
	}

	return jobInfo, nil
}

// GetResponseChannel returns response channel for job callbacks
// Возвращает канал ответов для callback'ов job'ов
func (c *Component) GetResponseChannel() <-chan string {
	return c.responseChannel
}

// SendJobCallback sends job callback response
// Отправляет callback ответ job'а
func (c *Component) SendJobCallback(response string) {
	if c.responseChannel != nil {
		select {
		case c.responseChannel <- response:
		default:
			c.logger.Warn("Job response channel full, callback dropped")
		}
	}
}

// CancelJob cancels a job
func (c *Component) CancelJob(jobID, reason string) error {
	c.logger.Info("Canceling job", logger.String("jobID", jobID), logger.String("reason", reason))

	// Delegate to job manager
	return c.manager.CancelJob(context.Background(), jobID)
}

// JobInfo represents job information
type JobInfo struct {
	Key               string                 `json:"key"`
	Type              string                 `json:"type"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	Variables         map[string]interface{} `json:"variables"`
	Worker            string                 `json:"worker"`
	Retries           int                    `json:"retries"`
	CreatedAt         int64                  `json:"created_at"`
	Status            string                 `json:"status"`
	ErrorMessage      string                 `json:"error_message"`
}

// JobStats represents job statistics
type JobStats struct {
	TotalJobs      int32 `json:"total_jobs"`
	ActiveJobs     int32 `json:"active_jobs"`
	CompletedJobs  int32 `json:"completed_jobs"`
	FailedJobs     int32 `json:"failed_jobs"`
	ActivatedToday int32 `json:"activated_today"`
	CompletedToday int32 `json:"completed_today"`
}

// ProcessMessage processes JSON message from core engine
// Обрабатывает JSON сообщение от core engine
func (c *Component) ProcessMessage(ctx context.Context, messageJSON string) error {
	if !c.IsReady() {
		return fmt.Errorf("jobs component not ready")
	}

	// Parse message to determine type
	// Парсим сообщение для определения типа
	var request JobRequest
	if err := json.Unmarshal([]byte(messageJSON), &request); err != nil {
		return fmt.Errorf("failed to parse job message: %w", err)
	}

	c.logger.Debug("Processing job message", logger.String("type", request.Type), logger.String("request_id", request.RequestID))

	switch request.Type {
	case "create_job":
		return c.handleCreateJob(ctx, request)
	case "activate_jobs":
		return c.handleActivateJobs(ctx, request)
	case "complete_job":
		return c.handleCompleteJob(ctx, request)
	case "fail_job":
		return c.handleFailJob(ctx, request)
	case "cancel_job":
		return c.handleCancelJob(ctx, request)
	case "list_jobs":
		return c.handleListJobs(ctx, request)
	case "get_job":
		return c.handleGetJob(ctx, request)
	case "get_stats":
		return c.handleGetStats(ctx, request)
	default:
		return fmt.Errorf("unknown job message type: %s", request.Type)
	}
}

// handleCreateJob handles job creation request
// Обрабатывает запрос создания job'а
func (c *Component) handleCreateJob(ctx context.Context, request JobRequest) error {
	var payload CreateJobPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("create_job_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	jobID, err := c.CreateJobWithDetails(
		payload.JobType,
		payload.ProcessInstanceID,
		payload.ElementID,
		payload.CustomHeaders,
		payload.Variables)

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("create_job_response", request.RequestID, err.Error())
	} else {
		result := JobResult{
			JobID:     jobID,
			Success:   true,
			Timestamp: time.Now().Unix(),
		}
		response = CreateJobResponse("create_job_response", request.RequestID, result)
	}

	return c.sendResponse(response)
}

// handleActivateJobs handles job activation request
// Обрабатывает запрос активации job'ов
func (c *Component) handleActivateJobs(ctx context.Context, request JobRequest) error {
	var payload ActivateJobsPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("activate_jobs_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	var jobs []JobInfo
	var err error

	if payload.TimeoutMs > 0 {
		jobs, err = c.ActivateJobsWithTimeout(payload.WorkerName, payload.JobType, payload.MaxJobs, payload.TimeoutMs)
	} else {
		jobs, err = c.ActivateJobs(payload.WorkerName, payload.JobType, payload.MaxJobs)
	}

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("activate_jobs_response", request.RequestID, err.Error())
	} else {
		response = CreateJobResponse("activate_jobs_response", request.RequestID, jobs)
	}

	return c.sendResponse(response)
}

// handleCompleteJob handles job completion request
// Обрабатывает запрос завершения job'а
func (c *Component) handleCompleteJob(ctx context.Context, request JobRequest) error {
	var payload CompleteJobPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("complete_job_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	err := c.CompleteJob(payload.JobKey, payload.Variables)

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("complete_job_response", request.RequestID, err.Error())
	} else {
		result := JobResult{
			JobKey:    payload.JobKey,
			Success:   true,
			Message:   "Job completed successfully",
			Timestamp: time.Now().Unix(),
		}
		response = CreateJobResponse("complete_job_response", request.RequestID, result)
	}

	return c.sendResponse(response)
}

// handleFailJob handles job failure request
// Обрабатывает запрос провала job'а
func (c *Component) handleFailJob(ctx context.Context, request JobRequest) error {
	var payload FailJobPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("fail_job_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	err := c.FailJob(payload.JobKey, payload.Retries, payload.ErrorMessage)

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("fail_job_response", request.RequestID, err.Error())
	} else {
		result := JobResult{
			JobKey:    payload.JobKey,
			Success:   true,
			Message:   "Job failed with retry",
			Timestamp: time.Now().Unix(),
		}
		response = CreateJobResponse("fail_job_response", request.RequestID, result)
	}

	return c.sendResponse(response)
}

// handleCancelJob handles job cancellation request
// Обрабатывает запрос отмены job'а
func (c *Component) handleCancelJob(ctx context.Context, request JobRequest) error {
	var payload CancelJobPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("cancel_job_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	err := c.CancelJob(payload.JobKey, "Canceled via JSON API")

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("cancel_job_response", request.RequestID, err.Error())
	} else {
		result := JobResult{
			JobKey:    payload.JobKey,
			Success:   true,
			Message:   "Job canceled successfully",
			Timestamp: time.Now().Unix(),
		}
		response = CreateJobResponse("cancel_job_response", request.RequestID, result)
	}

	return c.sendResponse(response)
}

// handleListJobs handles job listing request
// Обрабатывает запрос списка job'ов
func (c *Component) handleListJobs(ctx context.Context, request JobRequest) error {
	var payload ListJobsPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("list_jobs_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	jobs, total, err := c.ListJobs(
		payload.JobType,
		payload.Worker,
		payload.ProcessInstanceID,
		payload.State,
		payload.Limit,
		payload.Offset)

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("list_jobs_response", request.RequestID, err.Error())
	} else {
		result := JobListResult{
			Jobs:   jobs,
			Total:  total,
			Limit:  payload.Limit,
			Offset: payload.Offset,
		}
		response = CreateJobResponse("list_jobs_response", request.RequestID, result)
	}

	return c.sendResponse(response)
}

// handleGetJob handles get job request
// Обрабатывает запрос получения job'а
func (c *Component) handleGetJob(ctx context.Context, request JobRequest) error {
	var payload GetJobPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateJobErrorResponse("get_job_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	job, err := c.GetJob(payload.JobID)

	var response JobResponse
	if err != nil {
		response = CreateJobErrorResponse("get_job_response", request.RequestID, err.Error())
	} else {
		response = CreateJobResponse("get_job_response", request.RequestID, job)
	}

	return c.sendResponse(response)
}

// handleGetStats handles get statistics request
// Обрабатывает запрос получения статистики
func (c *Component) handleGetStats(ctx context.Context, request JobRequest) error {
	// Create basic job statistics
	// Создаем базовую статистику job'ов
	stats := JobStatsResult{
		TotalJobs:     0, // TODO: Implement real statistics
		PendingJobs:   0,
		ActiveJobs:    0,
		CompletedJobs: 0,
		FailedJobs:    0,
		CanceledJobs:  0,
	}

	response := CreateJobResponse("get_stats_response", request.RequestID, stats)
	return c.sendResponse(response)
}

// sendResponse sends job response through response channel
// Отправляет ответ job'а через канал ответов
func (c *Component) sendResponse(response JobResponse) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		c.logger.Error("Failed to marshal job response", logger.String("error", err.Error()))
		return fmt.Errorf("failed to marshal job response: %w", err)
	}

	c.logger.Debug("Sending job response", logger.String("type", response.Type), logger.String("request_id", response.RequestID))

	if c.responseChannel != nil {
		select {
		case c.responseChannel <- string(responseJSON):
		default:
			c.logger.Warn("Job response channel full, response dropped")
			return fmt.Errorf("job response channel full")
		}
	}

	return nil
}
