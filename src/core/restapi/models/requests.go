/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"encoding/json"
	"time"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `json:"page" form:"page" binding:"min=1"`
	Limit int `json:"limit" form:"limit" binding:"min=1,max=1000"`
}

// GetDefaultPagination returns default pagination params
func GetDefaultPagination() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 20,
	}
}

// Process Management Requests

// StartProcessRequest represents process start request
type StartProcessRequest struct {
	ProcessKey string                 `json:"process_key" binding:"required"`
	Version    *int32                 `json:"version,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
}

// ListProcessInstancesRequest represents process instances list request
type ListProcessInstancesRequest struct {
	Status     string `json:"status" form:"status"`
	ProcessKey string `json:"process_key" form:"process_key"`
	TenantID   string `json:"tenant_id" form:"tenant_id"`
	PaginationParams
}

// CancelProcessRequest represents process cancellation request
type CancelProcessRequest struct {
	Reason string `json:"reason,omitempty"`
}

// Timer Management Requests

// AddTimerRequest represents timer creation request
type AddTimerRequest struct {
	TimerID      string `json:"timer_id" binding:"required"`
	Duration     string `json:"duration" binding:"required"`
	CallbackData string `json:"callback_data,omitempty"`
	Repeating    bool   `json:"repeating,omitempty"`
	Interval     string `json:"interval,omitempty"`
}

// ListTimersRequest represents timers list request
type ListTimersRequest struct {
	Status string `json:"status" form:"status"`
	PaginationParams
}

// Job Management Requests

// CreateJobRequest represents job creation request
type CreateJobRequest struct {
	Type              string                 `json:"type" binding:"required"`
	ProcessInstanceID string                 `json:"process_instance_id" binding:"required"`
	ElementID         string                 `json:"element_id" binding:"required"`
	ElementInstanceID string                 `json:"element_instance_id,omitempty"`
	CustomHeaders     map[string]string      `json:"custom_headers,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	Retries           int32                  `json:"retries,omitempty"`
	TimeoutMs         int64                  `json:"timeout_ms,omitempty"`
}

// ActivateJobsRequest represents job activation request
type ActivateJobsRequest struct {
	Type           string   `json:"type" binding:"required"`
	Worker         string   `json:"worker" binding:"required"`
	MaxJobs        int32    `json:"max_jobs,omitempty"`
	TimeoutMs      int64    `json:"timeout_ms,omitempty"`
	FetchVariables []string `json:"fetch_variables,omitempty"`
}

// CompleteJobRequest represents job completion request
type CompleteJobRequest struct {
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// FailJobRequest represents job failure request
type FailJobRequest struct {
	Retries      int32  `json:"retries" binding:"required"`
	ErrorMessage string `json:"error_message,omitempty"`
	BackoffMs    int64  `json:"backoff_ms,omitempty"`
}

// ThrowErrorRequest represents job error throwing request
type ThrowErrorRequest struct {
	ErrorCode    string                 `json:"error_code" binding:"required"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
}

// ListJobsRequest represents jobs list request
type ListJobsRequest struct {
	Type   string `json:"type" form:"type"`
	Worker string `json:"worker" form:"worker"`
	State  string `json:"state" form:"state"`
	PaginationParams
}

// UpdateJobRetriesRequest represents job retries update request
type UpdateJobRetriesRequest struct {
	Retries int32 `json:"retries" binding:"required,min=0,max=100"`
}

// CancelJobRequest represents job cancellation request
type CancelJobRequest struct {
	Reason string `json:"reason,omitempty"`
}

// UpdateJobTimeoutRequest represents job timeout update request
type UpdateJobTimeoutRequest struct {
	TimeoutMs int64 `json:"timeout_ms" binding:"required,min=1000,max=86400000"` // 1 second to 24 hours
}

// Message Management Requests

// PublishMessageRequest represents message publishing request
type PublishMessageRequest struct {
	TenantID       string                 `json:"tenant_id,omitempty"`
	MessageName    string                 `json:"message_name" binding:"required"`
	CorrelationKey string                 `json:"correlation_key,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	TTLSeconds     int64                  `json:"ttl_seconds,omitempty"`
}

// ListMessagesRequest represents messages list request
type ListMessagesRequest struct {
	TenantID string `json:"tenant_id" form:"tenant_id"`
	PaginationParams
}

// Expression Management Requests

// EvaluateExpressionRequest represents expression evaluation request
type EvaluateExpressionRequest struct {
	Expression string                 `json:"expression" binding:"required"`
	Context    map[string]interface{} `json:"context,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
}

// ValidateExpressionRequest represents expression validation request
type ValidateExpressionRequest struct {
	Expression string `json:"expression" binding:"required"`
	Schema     string `json:"schema,omitempty"`
}

// ParseExpressionRequest represents expression parsing request
type ParseExpressionRequest struct {
	Expression string `json:"expression" binding:"required"`
}

// TestExpressionRequest represents expression testing request
type TestExpressionRequest struct {
	Expression string                   `json:"expression" binding:"required"`
	TestCases  []map[string]interface{} `json:"test_cases" binding:"required"`
}

// BPMN Management Requests

// ParseBPMNRequest represents BPMN parsing request
type ParseBPMNRequest struct {
	ProcessID string `json:"process_id,omitempty"`
	Force     bool   `json:"force,omitempty"`
}

// ListBPMNProcessesRequest represents BPMN processes list request
type ListBPMNProcessesRequest struct {
	TenantID string `json:"tenant_id" form:"tenant_id"`
	PaginationParams
}

// Incident Management Requests

// ListIncidentsRequest represents incidents list request
type ListIncidentsRequest struct {
	Status            []string   `json:"status" form:"status"`
	Type              []string   `json:"type" form:"type"`
	ProcessInstanceID string     `json:"process_instance_id" form:"process_instance_id"`
	ProcessKey        string     `json:"process_key" form:"process_key"`
	ElementID         string     `json:"element_id" form:"element_id"`
	JobKey            string     `json:"job_key" form:"job_key"`
	WorkerID          string     `json:"worker_id" form:"worker_id"`
	CreatedAfter      *time.Time `json:"created_after" form:"created_after"`
	CreatedBefore     *time.Time `json:"created_before" form:"created_before"`
	PaginationParams
}

// ResolveIncidentRequest represents incident resolution request
type ResolveIncidentRequest struct {
	Action     string `json:"action" binding:"required,oneof=retry dismiss"`
	Comment    string `json:"comment,omitempty"`
	ResolvedBy string `json:"resolved_by,omitempty"`
	NewRetries int32  `json:"new_retries,omitempty"`
}

// Token Management Requests

// ListTokensRequest represents tokens list request
type ListTokensRequest struct {
	InstanceID string `json:"instance_id" form:"instance_id"`
	State      string `json:"state" form:"state"`
	PaginationParams
}

// Generic filter request
type FilterRequest struct {
	Filters map[string]interface{} `json:"filters,omitempty"`
	PaginationParams
}

// Unmarshal JSON with validation
func (r *StartProcessRequest) UnmarshalJSON(data []byte) error {
	type Alias StartProcessRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, &aux)
}

// Validate request fields
func (r *StartProcessRequest) Validate() error {
	if r.ProcessKey == "" {
		return BadRequestError("process_key is required")
	}
	return nil
}

func (r *AddTimerRequest) Validate() error {
	if r.TimerID == "" {
		return BadRequestError("timer_id is required")
	}
	if r.Duration == "" {
		return BadRequestError("duration is required")
	}
	return nil
}

func (r *PublishMessageRequest) Validate() error {
	if r.MessageName == "" {
		return BadRequestError("message_name is required")
	}
	return nil
}

func (r *FailJobRequest) Validate() error {
	if r.Retries < 0 {
		return BadRequestError("retries cannot be negative")
	}
	return nil
}

func (r *ThrowErrorRequest) Validate() error {
	if r.ErrorCode == "" {
		return BadRequestError("error_code is required")
	}
	return nil
}

func (r *UpdateJobRetriesRequest) Validate() error {
	if r.Retries < 0 {
		return BadRequestError("retries cannot be negative")
	}
	if r.Retries > 100 {
		return BadRequestError("retries cannot exceed 100")
	}
	return nil
}

func (r *UpdateJobTimeoutRequest) Validate() error {
	if r.TimeoutMs < 1000 {
		return BadRequestError("timeout must be at least 1000ms (1 second)")
	}
	if r.TimeoutMs > 86400000 {
		return BadRequestError("timeout cannot exceed 86400000ms (24 hours)")
	}
	return nil
}
