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

// APIResponse represents standard API response format
type APIResponse struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *APIError    `json:"error,omitempty"`
	Meta    ResponseMeta `json:"meta"`
}

// ResponseMeta contains response metadata
type ResponseMeta struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id"`
}

// PaginatedResponse represents paginated API response
type PaginatedResponse struct {
	Success    bool            `json:"success"`
	Data       interface{}     `json:"data,omitempty"`
	Error      *APIError       `json:"error,omitempty"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Meta       ResponseMeta    `json:"meta"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasNext bool `json:"has_next"`
	HasPrev bool `json:"has_prev"`
}

// SuccessResponse creates successful API response
func SuccessResponse(data interface{}, requestID string) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Meta: ResponseMeta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	}
}

// ErrorResponse creates error API response
func ErrorResponse(err *APIError, requestID string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error:   err,
		Meta: ResponseMeta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	}
}

// PaginatedSuccessResponse creates successful paginated API response
func PaginatedSuccessResponse(data interface{}, pagination *PaginationInfo, requestID string) *PaginatedResponse {
	return &PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Meta: ResponseMeta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	}
}

// PaginatedErrorResponse creates error paginated API response
func PaginatedErrorResponse(err *APIError, requestID string) *PaginatedResponse {
	return &PaginatedResponse{
		Success: false,
		Error:   err,
		Meta: ResponseMeta{
			Timestamp: time.Now(),
			RequestID: requestID,
		},
	}
}

// StatusResponse represents daemon status response
type StatusResponse struct {
	Status    string    `json:"status"`
	Uptime    int64     `json:"uptime_seconds"`
	Version   string    `json:"version,omitempty"`
	StartedAt time.Time `json:"started_at"`
	IsHealthy bool      `json:"is_healthy"`
}

// StatsResponse represents generic statistics response
type StatsResponse struct {
	TotalCount int                    `json:"total_count"`
	Counts     map[string]int         `json:"counts"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ListResponse represents generic list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int         `json:"total_count"`
}

// CreateResponse represents resource creation response
type CreateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

// UpdateResponse represents resource update response
type UpdateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

// DeleteResponse represents resource deletion response
type DeleteResponse struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Checks    map[string]interface{} `json:"checks"`
	Timestamp time.Time              `json:"timestamp"`
}
