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
	"sync"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// JobCallback represents job completion callback
// Представляет callback завершения job'а
type JobCallback struct {
	JobID             string                 `json:"job_id"`
	ElementID         string                 `json:"element_id"`
	TokenID           string                 `json:"token_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	Status            string                 `json:"status"` // "COMPLETED", "FAILED", "ERROR"
	Variables         map[string]interface{} `json:"variables,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"` // For BPMN errors
	CompletedAt       time.Time              `json:"completed_at"`
}

// JobManager manages job lifecycle and operations
type JobManager struct {
	storage   storage.Storage
	logger    logger.ComponentLogger
	workers   map[string]*WorkerInfo
	mutex     sync.RWMutex
	isRunning bool
	stopChan  chan struct{}
	component JobsComponentInterface
}

// JobsComponentInterface defines interface for job callback handling
// Определяет интерфейс для обработки callback'ов job'ов
type JobsComponentInterface interface {
	SendJobCallback(response string)
}

// WorkerInfo contains information about job worker
type WorkerInfo struct {
	ID           string
	JobType      string
	LastPing     time.Time
	ActiveJobs   int
	MaxJobs      int
	Timeout      time.Duration
	FetchTimeout time.Duration
}

// ListJobsFilter contains filtering options for listing jobs
type ListJobsFilter struct {
	Type              string
	Worker            string
	ProcessInstanceID string
	ProcessKey        string
	State             string
	TenantID          string
	IncludeVariables  bool
	Limit             int
	Offset            int
}

// NewJobManager creates new job manager
func NewJobManager(storage storage.Storage, logger logger.ComponentLogger, component JobsComponentInterface) *JobManager {
	return &JobManager{
		storage:   storage,
		logger:    logger,
		workers:   make(map[string]*WorkerInfo),
		stopChan:  make(chan struct{}),
		component: component,
	}
}

// Start starts the job manager
func (jm *JobManager) Start() error {
	jm.logger.Info("Starting job manager")

	jm.isRunning = true

	// Start cleanup goroutine for expired jobs
	go jm.cleanupExpiredJobs()

	// Start worker health check
	go jm.monitorWorkers()

	jm.logger.Info("Job manager started successfully")
	return nil
}

// Stop stops the job manager
func (jm *JobManager) Stop() {
	jm.logger.Info("Stopping job manager")

	jm.isRunning = false
	close(jm.stopChan)

	jm.logger.Info("Job manager stopped")
}

// IsRunning returns job manager running status
func (jm *JobManager) IsRunning() bool {
	return jm.isRunning
}

// CreateJob creates a new job
func (jm *JobManager) CreateJob(ctx context.Context, job *models.Job) error {
	jm.logger.Info("Creating job", logger.String("type", job.Type), logger.String("id", job.ID))

	// Save job to storage
	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	jm.logger.Info("Job created successfully")
	return nil
}

// ActivateJobs activates jobs for worker
func (jm *JobManager) ActivateJobs(ctx context.Context, jobType, workerID string, maxJobs int, timeout time.Duration) ([]*models.Job, error) {
	jm.logger.Info("Activating jobs", logger.String("worker", workerID), logger.Int("maxJobs", maxJobs))

	// Register or update worker info
	jm.registerWorker(workerID, jobType, maxJobs, timeout)

	// Get available jobs
	jobs, err := jm.storage.ListJobsByType(ctx, jobType, models.JobStatusPending, maxJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	jm.logger.Debug("Found jobs for activation",
		logger.String("jobType", jobType),
		logger.String("status", string(models.JobStatusPending)),
		logger.Int("count", len(jobs)))

	var activatedJobs []*models.Job
	for _, job := range jobs {
		jm.logger.Debug("Processing job for activation",
			logger.String("jobID", job.ID),
			logger.String("currentStatus", string(job.Status)),
			logger.String("currentWorker", job.WorkerID))

		// Double-check job is actually pending (race condition protection)
		if job.Status != models.JobStatusPending {
			jm.logger.Warn("Skipping job - not pending",
				logger.String("jobID", job.ID),
				logger.String("status", string(job.Status)))
			continue
		}

		// Re-read job from storage to check if still pending (avoid race condition)
		freshJob, err := jm.storage.GetJob(ctx, job.ID)
		if err != nil {
			jm.logger.Error("Failed to re-read job", logger.String("error", err.Error()))
			continue
		}
		
		if freshJob == nil || freshJob.Status != models.JobStatusPending {
			jm.logger.Warn("Job no longer pending - skipping",
				logger.String("jobID", job.ID),
				logger.String("freshStatus", string(freshJob.Status)))
			continue
		}

		// Mark job as running
		freshJob.MarkAsStarted(workerID)

		// Set lease expiry
		leaseExpiry := time.Now().Add(timeout)
		freshJob.ScheduledAt = &leaseExpiry

		jm.logger.Debug("Marking job as started",
			logger.String("jobID", freshJob.ID),
			logger.String("newWorker", workerID),
			logger.String("newStatus", string(freshJob.Status)),
			logger.String("timeout", timeout.String()),
			logger.String("scheduledAt", leaseExpiry.Format("15:04:05.000")))

		if err := jm.storage.SaveJob(ctx, freshJob); err != nil {
			jm.logger.Error("Failed to save activated job", logger.String("error", err.Error()))
			continue
		}

		// Verify job was saved correctly
		savedJob, err := jm.storage.GetJob(ctx, freshJob.ID)
		if err == nil && savedJob != nil {
			jm.logger.Debug("Job saved verification",
				logger.String("jobID", savedJob.ID),
				logger.String("savedStatus", string(savedJob.Status)),
				logger.String("savedWorker", savedJob.WorkerID))
		}

		activatedJobs = append(activatedJobs, freshJob)

		if len(activatedJobs) >= maxJobs {
			break
		}
	}

	jm.logger.Info("Jobs activated", logger.String("worker", workerID), logger.Int("count", len(activatedJobs)))
	return activatedJobs, nil
}

// CompleteJob completes a job
func (jm *JobManager) CompleteJob(ctx context.Context, jobID string, variables map[string]interface{}) error {
	jm.logger.Info("Completing job", logger.String("jobID", jobID))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != models.JobStatusRunning {
		return fmt.Errorf("job is not running: %s", jobID)
	}

	// Update job variables if provided
	if variables != nil {
		if job.Variables == nil {
			job.Variables = make(map[string]interface{})
		}
		for k, v := range variables {
			job.Variables[k] = v
		}
	}

	job.MarkAsCompleted()

	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save completed job: %w", err)
	}

	// Update worker info
	jm.updateWorkerActiveJobs(job.WorkerID, -1)

	// Send job completion callback
	callback := JobCallback{
		JobID:             job.ID,
		ElementID:         job.ElementID,
		TokenID:           job.TokenID,
		ProcessInstanceID: job.ProcessInstanceID,
		Status:            "COMPLETED",
		Variables:         variables,
		CompletedAt:       time.Now(),
	}

	if jm.component != nil {
		if callbackJSON, err := json.Marshal(callback); err == nil {
			jm.component.SendJobCallback(string(callbackJSON))
			jm.logger.Info("Job completion callback sent",
				logger.String("jobID", job.ID),
				logger.String("elementID", job.ElementID))
		}
	}

	jm.logger.Info("Job completed successfully")
	return nil
}

// FailJob fails a job
func (jm *JobManager) FailJob(ctx context.Context, jobID string, retries int, errorMessage string, retryBackoff time.Duration) error {
	jm.logger.Info("Failing job", logger.String("jobID", jobID), logger.Int("retries", retries), logger.String("error", errorMessage))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update retries and mark as failed
	now := time.Now()
	job.Status = models.JobStatusFailed
	job.ErrorMessage = errorMessage
	job.Retries = retries // Set explicit retries value from CLI
	job.CompletedAt = &now
	job.UpdatedAt = now

	// Check if can retry BEFORE changing status to DEFERRED
	canRetry := job.CanRetry()

	// Schedule retry if retries available
	if canRetry && retryBackoff > 0 {
		retryTime := time.Now().Add(retryBackoff)
		job.Status = models.JobStatusDeferred
		job.ScheduledAt = &retryTime
	}

	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save failed job: %w", err)
	}

	// Update worker info
	jm.updateWorkerActiveJobs(job.WorkerID, -1)

	// Send job failure callback only if cannot retry anymore
	if !canRetry {
		callback := JobCallback{
			JobID:             job.ID,
			ElementID:         job.ElementID,
			TokenID:           job.TokenID,
			ProcessInstanceID: job.ProcessInstanceID,
			Status:            "FAILED",
			ErrorMessage:      errorMessage,
			CompletedAt:       time.Now(),
		}

		if jm.component != nil {
			if callbackJSON, err := json.Marshal(callback); err == nil {
				jm.component.SendJobCallback(string(callbackJSON))
				jm.logger.Info("Job failure callback sent",
					logger.String("jobID", job.ID),
					logger.String("elementID", job.ElementID))
			}
		}
	}

	jm.logger.Info("Job failed", logger.String("jobID", jobID), logger.Bool("canRetry", job.CanRetry()))
	return nil
}

// ThrowError throws error for job
func (jm *JobManager) ThrowError(ctx context.Context, jobID, errorCode, errorMessage string, variables map[string]interface{}) error {
	jm.logger.Info("Throwing error for job", logger.String("jobID", jobID), logger.String("errorCode", errorCode))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Update job variables if provided
	if variables != nil {
		if job.Variables == nil {
			job.Variables = make(map[string]interface{})
		}
		for k, v := range variables {
			job.Variables[k] = v
		}
	}

	// Add error information to metadata
	if job.Metadata == nil {
		job.Metadata = make(map[string]string)
	}
	job.Metadata["errorCode"] = errorCode
	job.Metadata["errorType"] = "BPMN_ERROR"

	job.MarkAsFailed(errorMessage)

	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save job with error: %w", err)
	}

	// Update worker info
	jm.updateWorkerActiveJobs(job.WorkerID, -1)

	jm.logger.Info("Error thrown for job", logger.String("errorCode", errorCode))
	return nil
}

// UpdateJobRetries updates job retries
func (jm *JobManager) UpdateJobRetries(ctx context.Context, jobID string, retries int) error {
	jm.logger.Info("Updating job retries", logger.String("jobID", jobID), logger.Int("retries", retries))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Retries = retries
	job.UpdatedAt = time.Now()

	// If job was failed but now has retries, make it pending again
	if job.Status == models.JobStatusFailed && retries > 0 {
		job.Status = models.JobStatusPending
		job.ErrorMessage = ""
		job.CompletedAt = nil
	}

	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	jm.logger.Info("Job retries updated", logger.Int("retries", retries))
	return nil
}

// CancelJob cancels a job
func (jm *JobManager) CancelJob(ctx context.Context, jobID string) error {
	jm.logger.Info("Canceling job", logger.String("jobID", jobID))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.IsCompleted() {
		return fmt.Errorf("job is already completed: %s", jobID)
	}

	job.Status = models.JobStatusCanceled
	job.UpdatedAt = time.Now()
	now := time.Now()
	job.CompletedAt = &now

	if err := jm.storage.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to save canceled job: %w", err)
	}

	// Update worker info
	if job.WorkerID != "" {
		jm.updateWorkerActiveJobs(job.WorkerID, -1)
	}

	jm.logger.Info("Job canceled")
	return nil
}

// ListJobs lists jobs with filtering
func (jm *JobManager) ListJobs(ctx context.Context, filter *ListJobsFilter) ([]*models.Job, int, error) {
	jm.logger.Debug("Listing jobs with filter", logger.String("worker", filter.Worker))

	// Convert state filter to JobStatus
	var status models.JobStatus
	if filter.State != "" {
		status = models.JobStatus(filter.State)
	}

	jobs, err := jm.storage.ListJobsByType(ctx, filter.Type, status, filter.Limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Apply additional filters
	var filteredJobs []*models.Job
	for _, job := range jobs {
		// Filter by worker
		if filter.Worker != "" && job.WorkerID != filter.Worker {
			continue
		}

		// Filter by process instance
		if filter.ProcessInstanceID != "" && job.ProcessInstanceID != filter.ProcessInstanceID {
			continue
		}

		// Filter by process key (would need process definition lookup)
		// TODO: implement process key filtering

		filteredJobs = append(filteredJobs, job)

		// Apply limit
		if len(filteredJobs) >= filter.Limit {
			break
		}
	}

	// Apply offset
	start := filter.Offset
	if start > len(filteredJobs) {
		start = len(filteredJobs)
	}

	end := start + filter.Limit
	if end > len(filteredJobs) {
		end = len(filteredJobs)
	}

	result := filteredJobs[start:end]
	total := len(filteredJobs)

	jm.logger.Debug("Jobs listed", logger.Int("returned", len(result)))
	return result, total, nil
}

// GetJob gets job by ID
func (jm *JobManager) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	return jm.storage.GetJob(ctx, jobID)
}

// UpdateJobTimeout updates job timeout
func (jm *JobManager) UpdateJobTimeout(ctx context.Context, jobID string, timeout time.Duration) error {
	jm.logger.Info("Updating job timeout", logger.String("jobID", jobID), logger.String("timeout", timeout.String()))

	job, err := jm.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == models.JobStatusRunning && job.ScheduledAt != nil {
		// Extend lease expiry
		newExpiry := time.Now().Add(timeout)
		job.ScheduledAt = &newExpiry
		job.UpdatedAt = time.Now()

		if err := jm.storage.SaveJob(ctx, job); err != nil {
			return fmt.Errorf("failed to save job: %w", err)
		}
	}

	jm.logger.Info("Job timeout updated", logger.String("jobID", jobID))
	return nil
}

// registerWorker registers or updates worker information
func (jm *JobManager) registerWorker(workerID, jobType string, maxJobs int, timeout time.Duration) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	worker, exists := jm.workers[workerID]
	if !exists {
		worker = &WorkerInfo{
			ID:      workerID,
			JobType: jobType,
			MaxJobs: maxJobs,
			Timeout: timeout,
		}
		jm.workers[workerID] = worker
	}

	worker.LastPing = time.Now()
	worker.JobType = jobType
	worker.MaxJobs = maxJobs
	worker.Timeout = timeout
}

// updateWorkerActiveJobs updates worker active job count
func (jm *JobManager) updateWorkerActiveJobs(workerID string, delta int) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	if worker, exists := jm.workers[workerID]; exists {
		worker.ActiveJobs += delta
		if worker.ActiveJobs < 0 {
			worker.ActiveJobs = 0
		}
	}
}

// cleanupExpiredJobs runs cleanup for expired jobs
func (jm *JobManager) cleanupExpiredJobs() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jm.performCleanup()
		case <-jm.stopChan:
			return
		}
	}
}

// performCleanup performs cleanup of expired jobs
func (jm *JobManager) performCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all running jobs
	jobs, err := jm.storage.ListJobsByType(ctx, "", models.JobStatusRunning, 1000)
	if err != nil {
		jm.logger.Error("Failed to list jobs for cleanup", logger.String("error", err.Error()))
		return
	}

	now := time.Now()
	expiredCount := 0

	for _, job := range jobs {
		// Check if job lease has expired
		if job.ScheduledAt != nil && now.After(*job.ScheduledAt) {
			jm.logger.Debug("Job expired",
				logger.String("jobID", job.ID),
				logger.String("now", now.Format("15:04:05.000")),
				logger.String("scheduledAt", job.ScheduledAt.Format("15:04:05.000")))

			// Reset job to pending for retry
			job.Status = models.JobStatusPending
			job.WorkerID = ""
			job.ScheduledAt = nil
			job.UpdatedAt = now

			if err := jm.storage.SaveJob(ctx, job); err != nil {
				jm.logger.Error("Failed to reset expired job", logger.String("error", err.Error()))
				continue
			}

			expiredCount++
			jm.logger.Info("Reset expired job", logger.String("type", job.Type))
		}
	}

	if expiredCount > 0 {
		jm.logger.Info("Cleaned up expired jobs", logger.Int("cleanedCount", expiredCount))
	}
}

// monitorWorkers monitors worker health
func (jm *JobManager) monitorWorkers() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jm.checkWorkerHealth()
		case <-jm.stopChan:
			return
		}
	}
}

// checkWorkerHealth checks worker health and removes inactive workers
func (jm *JobManager) checkWorkerHealth() {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	now := time.Now()
	inactiveThreshold := 5 * time.Minute

	for workerID, worker := range jm.workers {
		if now.Sub(worker.LastPing) > inactiveThreshold {
			jm.logger.Info("Removing inactive worker", logger.String("workerID", workerID), logger.String("lastPing", worker.LastPing.String()))
			delete(jm.workers, workerID)
		}
	}
}
