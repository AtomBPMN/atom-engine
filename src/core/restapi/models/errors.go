/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"fmt"
	"net/http"
)

// Error codes for API responses
const (
	// General errors
	ErrorCodeInternalError   = "INTERNAL_ERROR"
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeNotFound        = "NOT_FOUND"
	ErrorCodeConflict        = "CONFLICT"
	ErrorCodeValidationError = "VALIDATION_ERROR"

	// Authentication errors
	ErrorCodeUnauthorized            = "UNAUTHORIZED"
	ErrorCodeForbidden               = "FORBIDDEN"
	ErrorCodeInvalidAPIKey           = "INVALID_API_KEY"
	ErrorCodeMissingAPIKey           = "MISSING_API_KEY"
	ErrorCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// Rate limiting
	ErrorCodeRateLimited = "RATE_LIMITED"
	ErrorCodeIPBlocked   = "IP_BLOCKED"

	// Resource errors
	ErrorCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrorCodeResourceConflict = "RESOURCE_CONFLICT"
	ErrorCodeResourceLocked   = "RESOURCE_LOCKED"

	// Process errors
	ErrorCodeProcessNotFound  = "PROCESS_NOT_FOUND"
	ErrorCodeProcessFailed    = "PROCESS_FAILED"
	ErrorCodeInstanceNotFound = "INSTANCE_NOT_FOUND"

	// Job errors
	ErrorCodeJobNotFound    = "JOB_NOT_FOUND"
	ErrorCodeJobFailed      = "JOB_FAILED"
	ErrorCodeWorkerNotFound = "WORKER_NOT_FOUND"

	// Timer errors
	ErrorCodeTimerNotFound   = "TIMER_NOT_FOUND"
	ErrorCodeTimerFailed     = "TIMER_FAILED"
	ErrorCodeInvalidDuration = "INVALID_DURATION"

	// Message errors
	ErrorCodeMessageFailed     = "MESSAGE_FAILED"
	ErrorCodeCorrelationFailed = "CORRELATION_FAILED"

	// Expression errors
	ErrorCodeExpressionError = "EXPRESSION_ERROR"
	ErrorCodeSyntaxError     = "SYNTAX_ERROR"

	// Storage errors
	ErrorCodeStorageError  = "STORAGE_ERROR"
	ErrorCodeDatabaseError = "DATABASE_ERROR"

	// BPMN errors
	ErrorCodeBPMNParseError      = "BPMN_PARSE_ERROR"
	ErrorCodeBPMNValidationError = "BPMN_VALIDATION_ERROR"
)

// APIError represents API error response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// NewAPIError creates new API error
func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithDetails creates new API error with details
func NewAPIErrorWithDetails(code, message string, details map[string]interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewValidationError creates validation error
func NewValidationError(message string, errors []ValidationError) *APIError {
	details := map[string]interface{}{
		"validation_errors": errors,
	}
	return &APIError{
		Code:    ErrorCodeValidationError,
		Message: message,
		Details: details,
	}
}

// HTTPStatusFromErrorCode maps error codes to HTTP status codes
func HTTPStatusFromErrorCode(code string) int {
	switch code {
	case ErrorCodeBadRequest, ErrorCodeValidationError, ErrorCodeInvalidDuration,
		ErrorCodeBPMNParseError, ErrorCodeBPMNValidationError, ErrorCodeSyntaxError:
		return http.StatusBadRequest

	case ErrorCodeUnauthorized, ErrorCodeInvalidAPIKey, ErrorCodeMissingAPIKey:
		return http.StatusUnauthorized

	case ErrorCodeForbidden, ErrorCodeInsufficientPermissions, ErrorCodeIPBlocked:
		return http.StatusForbidden

	case ErrorCodeNotFound, ErrorCodeResourceNotFound, ErrorCodeProcessNotFound,
		ErrorCodeInstanceNotFound, ErrorCodeJobNotFound, ErrorCodeTimerNotFound,
		ErrorCodeWorkerNotFound:
		return http.StatusNotFound

	case ErrorCodeConflict, ErrorCodeResourceConflict:
		return http.StatusConflict

	case ErrorCodeResourceLocked:
		return http.StatusLocked

	case ErrorCodeRateLimited:
		return http.StatusTooManyRequests

	case ErrorCodeInternalError, ErrorCodeProcessFailed, ErrorCodeJobFailed,
		ErrorCodeTimerFailed, ErrorCodeMessageFailed, ErrorCodeCorrelationFailed,
		ErrorCodeExpressionError, ErrorCodeStorageError, ErrorCodeDatabaseError:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func BadRequestError(message string) *APIError {
	return NewAPIError(ErrorCodeBadRequest, message)
}

func NotFoundError(message string) *APIError {
	return NewAPIError(ErrorCodeNotFound, message)
}

func UnauthorizedError(message string) *APIError {
	return NewAPIError(ErrorCodeUnauthorized, message)
}

func ForbiddenError(message string) *APIError {
	return NewAPIError(ErrorCodeForbidden, message)
}

func InternalServerError(message string) *APIError {
	return NewAPIError(ErrorCodeInternalError, message)
}

func ConflictError(message string) *APIError {
	return NewAPIError(ErrorCodeConflict, message)
}

func RateLimitedError(message string) *APIError {
	return NewAPIError(ErrorCodeRateLimited, message)
}

func ProcessNotFoundError(processID string) *APIError {
	return NewAPIErrorWithDetails(
		ErrorCodeProcessNotFound,
		"Process not found",
		map[string]interface{}{"process_id": processID},
	)
}

func JobNotFoundError(jobKey string) *APIError {
	return NewAPIErrorWithDetails(
		ErrorCodeJobNotFound,
		"Job not found",
		map[string]interface{}{"job_key": jobKey},
	)
}

func TimerNotFoundError(timerID string) *APIError {
	return NewAPIErrorWithDetails(
		ErrorCodeTimerNotFound,
		"Timer not found",
		map[string]interface{}{"timer_id": timerID},
	)
}
