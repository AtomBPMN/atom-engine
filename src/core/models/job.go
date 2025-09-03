/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"time"
)

// JobStatus represents job status
type JobStatus string

const (
	JobStatusPending     JobStatus = "PENDING"
	JobStatusRunning     JobStatus = "RUNNING"
	JobStatusCompleted   JobStatus = "COMPLETED"
	JobStatusFailed      JobStatus = "FAILED"
	JobStatusCanceled    JobStatus = "CANCELED"
	JobStatusDeferred    JobStatus = "DEFERRED"
	JobStatusTransferred JobStatus = "TRANSFERRED"
)

// Job represents a job in the system
type Job struct {
	// Basic fields
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Status     JobStatus `json:"status"`
	WorkerID   string    `json:"worker_id"`
	Retries    int       `json:"retries"`
	MaxRetries int       `json:"max_retries"`

	// Related entities
	ProcessInstanceID string `json:"process_instance_id"`
	ElementID         string `json:"element_id"`
	ElementInstanceID string `json:"element_instance_id"`
	TokenID           string `json:"token_id"` // Token that created this job

	// Job data
	CustomHeaders map[string]string      `json:"custom_headers"`
	Variables     map[string]interface{} `json:"variables"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Scheduling
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Priority    int        `json:"priority"`

	// Metadata
	ErrorMessage string            `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewJob creates a new job
func NewJob(jobType, processInstanceID, elementID string) *Job {
	now := time.Now()
	return &Job{
		ID:                GenerateID(),
		Type:              jobType,
		Status:            JobStatusPending,
		ProcessInstanceID: processInstanceID,
		ElementID:         elementID,
		CustomHeaders:     make(map[string]string),
		Variables:         make(map[string]interface{}),
		Metadata:          make(map[string]string),
		CreatedAt:         now,
		UpdatedAt:         now,
		Priority:          0,
		MaxRetries:        3,
	}
}

// IsActive checks if job is active
func (j *Job) IsActive() bool {
	return j.Status == JobStatusPending || j.Status == JobStatusRunning
}

// IsCompleted checks if job is completed
func (j *Job) IsCompleted() bool {
	return j.Status == JobStatusCompleted || j.Status == JobStatusFailed || j.Status == JobStatusCanceled
}

// CanRetry checks if job can be retried
func (j *Job) CanRetry() bool {
	return j.Status == JobStatusFailed && j.Retries > 0
}

// MarkAsStarted marks job as started
func (j *Job) MarkAsStarted(workerID string) {
	now := time.Now()
	j.Status = JobStatusRunning
	j.WorkerID = workerID
	j.StartedAt = &now
	j.UpdatedAt = now
}

// MarkAsCompleted marks job as completed
func (j *Job) MarkAsCompleted() {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkAsFailed marks job as failed
func (j *Job) MarkAsFailed(errorMessage string) {
	now := time.Now()
	j.Status = JobStatusFailed
	j.ErrorMessage = errorMessage
	j.Retries++
	j.CompletedAt = &now
	j.UpdatedAt = now
}

// MarkAsDeferred marks job as deferred
func (j *Job) MarkAsDeferred(scheduledAt time.Time) {
	j.Status = JobStatusDeferred
	j.ScheduledAt = &scheduledAt
	j.UpdatedAt = time.Now()
}
