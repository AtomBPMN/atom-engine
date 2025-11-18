package types

import (
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusRunning   JobStatus = "RUNNING" 
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
)

// JobType represents the type of a job
type JobType string

// Common job types
const (
	JobTypeServiceTask JobType = "service-task"
	JobTypeUserTask    JobType = "user-task"
	JobTypeScriptTask  JobType = "script-task"
	JobTypeSendTask    JobType = "send-task"
	JobTypeReceiveTask JobType = "receive-task"
)

// JobVariables represents variables associated with a job
type JobVariables map[string]interface{}

// JobHeaders represents custom headers for a job
type JobHeaders map[string]string

// JobInfo represents detailed information about a job
type JobInfo struct {
	JobKey           string       `json:"job_key"`
	JobType          JobType      `json:"job_type"`
	ProcessInstanceID string      `json:"process_instance_id"`
	ElementID        string       `json:"element_id"`
	Worker           string       `json:"worker,omitempty"`
	Retries          int32        `json:"retries"`
	MaxRetries       int32        `json:"max_retries"`
	Status           JobStatus    `json:"status"`
	Variables        JobVariables `json:"variables,omitempty"`
	CustomHeaders    JobHeaders   `json:"custom_headers,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at,omitempty"`
	ActivatedAt      *time.Time   `json:"activated_at,omitempty"`
	CompletedAt      *time.Time   `json:"completed_at,omitempty"`
	ErrorMessage     string       `json:"error_message,omitempty"`
	Timeout          *time.Duration `json:"timeout,omitempty"`
}

// JobStats represents statistics about jobs in the system
type JobStats struct {
	TotalJobs       int64            `json:"total_jobs"`
	PendingJobs     int64            `json:"pending_jobs"`
	RunningJobs     int64            `json:"running_jobs"`
	CompletedJobs   int64            `json:"completed_jobs"`
	FailedJobs      int64            `json:"failed_jobs"`
	CancelledJobs   int64            `json:"cancelled_jobs"`
	JobsByType      map[JobType]int64 `json:"jobs_by_type"`
	JobsByWorker    map[string]int64  `json:"jobs_by_worker"`
	AverageRetries  float64          `json:"average_retries"`
	LastJobCreated  *time.Time       `json:"last_job_created,omitempty"`
	LastJobCompleted *time.Time      `json:"last_job_completed,omitempty"`
}

// JobCreateRequest represents a request to create a new job
type JobCreateRequest struct {
	JobType           JobType      `json:"job_type" validate:"required"`
	ProcessInstanceID string       `json:"process_instance_id" validate:"required"`
	ElementID         string       `json:"element_id" validate:"required"`
	Variables         JobVariables `json:"variables,omitempty"`
	CustomHeaders     JobHeaders   `json:"custom_headers,omitempty"`
	Retries           int32        `json:"retries,omitempty"`
	Timeout           *time.Duration `json:"timeout,omitempty"`
}

// JobActivateRequest represents a request to activate jobs
type JobActivateRequest struct {
	JobType    JobType       `json:"job_type" validate:"required"`
	Worker     string        `json:"worker" validate:"required"`
	MaxJobs    int32         `json:"max_jobs,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty"`
	FetchVariables []string  `json:"fetch_variables,omitempty"`
}

// JobCompleteRequest represents a request to complete a job
type JobCompleteRequest struct {
	JobKey    string       `json:"job_key" validate:"required"`
	Variables JobVariables `json:"variables,omitempty"`
}

// JobFailRequest represents a request to fail a job
type JobFailRequest struct {
	JobKey       string `json:"job_key" validate:"required"`
	Retries      int32  `json:"retries"`
	ErrorMessage string `json:"error_message,omitempty"`
	RetryBackoff *time.Duration `json:"retry_backoff,omitempty"`
}

// JobCancelRequest represents a request to cancel a job
type JobCancelRequest struct {
	JobKey string `json:"job_key" validate:"required"`
	Reason string `json:"reason,omitempty"`
}

// JobThrowErrorRequest represents a request to throw a BPMN error for a job
type JobThrowErrorRequest struct {
	JobKey      string `json:"job_key" validate:"required"`
	ErrorCode   string `json:"error_code" validate:"required"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// JobListRequest represents a request to list jobs
type JobListRequest struct {
	JobType           *JobType  `json:"job_type,omitempty"`
	Worker            *string   `json:"worker,omitempty"`
	Status            *JobStatus `json:"status,omitempty"`
	ProcessInstanceID *string   `json:"process_instance_id,omitempty"`
	Limit             int32     `json:"limit,omitempty"`
	Offset            int32     `json:"offset,omitempty"`
}

// JobListResponse represents a response with a list of jobs
type JobListResponse struct {
	Jobs       []JobInfo `json:"jobs"`
	TotalCount int32     `json:"total_count"`
	HasMore    bool      `json:"has_more"`
}

// JobActivateResponse represents a response from job activation
type JobActivateResponse struct {
	Jobs    []JobInfo `json:"jobs"`
	Count   int32     `json:"count"`
	Message string    `json:"message,omitempty"`
}

// JobOperationResult represents the result of a job operation
type JobOperationResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	JobKey    string `json:"job_key,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// Helper methods for JobVariables
func (jv JobVariables) GetString(key string) (string, bool) {
	if val, exists := jv[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (jv JobVariables) GetInt64(key string) (int64, bool) {
	if val, exists := jv[key]; exists {
		switch v := val.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		}
	}
	return 0, false
}

func (jv JobVariables) GetBool(key string) (bool, bool) {
	if val, exists := jv[key]; exists {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// Helper methods for JobInfo
func (ji *JobInfo) IsPending() bool {
	return ji.Status == JobStatusPending
}

func (ji *JobInfo) IsRunning() bool {
	return ji.Status == JobStatusRunning
}

func (ji *JobInfo) IsCompleted() bool {
	return ji.Status == JobStatusCompleted
}

func (ji *JobInfo) IsFailed() bool {
	return ji.Status == JobStatusFailed
}

func (ji *JobInfo) IsCancelled() bool {
	return ji.Status == JobStatusCancelled
}

func (ji *JobInfo) IsActive() bool {
	return ji.Status == JobStatusPending || ji.Status == JobStatusRunning
}

func (ji *JobInfo) IsFinished() bool {
	return ji.Status == JobStatusCompleted || ji.Status == JobStatusFailed || ji.Status == JobStatusCancelled
}
