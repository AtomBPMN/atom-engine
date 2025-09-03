/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package jobs

// JobRequest base structure for all job requests
// Базовая структура для всех запросов job'ов
type JobRequest struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
}

// JobResponse base structure for all job responses
// Базовая структура для всех ответов job'ов
type JobResponse struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// CreateJobPayload payload for creating a job
// Payload для создания job'а
type CreateJobPayload struct {
	JobType           string                 `json:"job_type"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ElementID         string                 `json:"element_id,omitempty"`
	CustomHeaders     map[string]string      `json:"custom_headers,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
}

// ActivateJobsPayload payload for activating jobs
// Payload для активации job'ов
type ActivateJobsPayload struct {
	WorkerName string `json:"worker_name"`
	JobType    string `json:"job_type"`
	MaxJobs    int    `json:"max_jobs"`
	TimeoutMs  int32  `json:"timeout_ms,omitempty"`
}

// CompleteJobPayload payload for completing a job
// Payload для завершения job'а
type CompleteJobPayload struct {
	JobKey    string                 `json:"job_key"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// FailJobPayload payload for failing a job
// Payload для провала job'а
type FailJobPayload struct {
	JobKey       string `json:"job_key"`
	Retries      int    `json:"retries"`
	ErrorMessage string `json:"error_message,omitempty"`
	RetryBackoff int64  `json:"retry_backoff,omitempty"`
}

// CancelJobPayload payload for canceling a job
// Payload для отмены job'а
type CancelJobPayload struct {
	JobKey string `json:"job_key"`
}

// ListJobsPayload payload for listing jobs
// Payload для списка job'ов
type ListJobsPayload struct {
	JobType           string `json:"job_type,omitempty"`
	Worker            string `json:"worker,omitempty"`
	ProcessInstanceID string `json:"process_instance_id,omitempty"`
	State             string `json:"state,omitempty"`
	Limit             int    `json:"limit,omitempty"`
	Offset            int    `json:"offset,omitempty"`
}

// GetJobPayload payload for getting a specific job
// Payload для получения конкретного job'а
type GetJobPayload struct {
	JobID string `json:"job_id"`
}

// JobResult result structure for job operations
// Структура результата для операций с job'ами
type JobResult struct {
	JobID     string `json:"job_id,omitempty"`
	JobKey    string `json:"job_key,omitempty"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// JobListResult result structure for job list operations
// Структура результата для операций списка job'ов
type JobListResult struct {
	Jobs   []JobInfo `json:"jobs"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// JobStatsResult result structure for job statistics
// Структура результата для статистики job'ов
type JobStatsResult struct {
	TotalJobs     int `json:"total_jobs"`
	PendingJobs   int `json:"pending_jobs"`
	ActiveJobs    int `json:"active_jobs"`
	CompletedJobs int `json:"completed_jobs"`
	FailedJobs    int `json:"failed_jobs"`
	CanceledJobs  int `json:"canceled_jobs"`
}
