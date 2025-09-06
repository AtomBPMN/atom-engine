package types

import (
	"fmt"
	"time"
)

// ErrorType represents the type/category of an error
type ErrorType string

const (
	// System level errors
	ErrorTypeSystem        ErrorType = "SYSTEM"
	ErrorTypeConfiguration ErrorType = "CONFIGURATION"
	ErrorTypeNetwork       ErrorType = "NETWORK"
	ErrorTypeStorage       ErrorType = "STORAGE"
	ErrorTypeTimeout       ErrorType = "TIMEOUT"

	// Authentication and authorization errors
	ErrorTypeAuth       ErrorType = "AUTHENTICATION"
	ErrorTypePermission ErrorType = "PERMISSION"
	ErrorTypeRateLimit  ErrorType = "RATE_LIMIT"

	// Validation errors
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeSchema     ErrorType = "SCHEMA"
	ErrorTypeFormat     ErrorType = "FORMAT"

	// BPMN specific errors
	ErrorTypeBPMN       ErrorType = "BPMN"
	ErrorTypeProcess    ErrorType = "PROCESS"
	ErrorTypeJob        ErrorType = "JOB"
	ErrorTypeMessage    ErrorType = "MESSAGE"
	ErrorTypeExpression ErrorType = "EXPRESSION"
	ErrorTypeTimer      ErrorType = "TIMER"
	ErrorTypeIncident   ErrorType = "INCIDENT"

	// Business logic errors
	ErrorTypeBusiness  ErrorType = "BUSINESS"
	ErrorTypeWorkflow  ErrorType = "WORKFLOW"
	ErrorTypeCondition ErrorType = "CONDITION"

	// External service errors
	ErrorTypeExternal    ErrorType = "EXTERNAL"
	ErrorTypeIntegration ErrorType = "INTEGRATION"
	ErrorTypeAPI         ErrorType = "API"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Generic error codes
	ErrorCodeUnknown       ErrorCode = "ERR_UNKNOWN"
	ErrorCodeInternal      ErrorCode = "ERR_INTERNAL"
	ErrorCodeNotFound      ErrorCode = "ERR_NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ERR_ALREADY_EXISTS"
	ErrorCodeInvalidInput  ErrorCode = "ERR_INVALID_INPUT"
	ErrorCodeUnauthorized  ErrorCode = "ERR_UNAUTHORIZED"
	ErrorCodeForbidden     ErrorCode = "ERR_FORBIDDEN"
	ErrorCodeTimeout       ErrorCode = "ERR_TIMEOUT"
	ErrorCodeRateLimit     ErrorCode = "ERR_RATE_LIMIT"

	// BPMN specific error codes
	ErrorCodeProcessNotFound      ErrorCode = "ERR_PROCESS_NOT_FOUND"
	ErrorCodeProcessInvalidState  ErrorCode = "ERR_PROCESS_INVALID_STATE"
	ErrorCodeJobNotFound          ErrorCode = "ERR_JOB_NOT_FOUND"
	ErrorCodeJobTimeout           ErrorCode = "ERR_JOB_TIMEOUT"
	ErrorCodeJobFailed            ErrorCode = "ERR_JOB_FAILED"
	ErrorCodeMessageNotCorrelated ErrorCode = "ERR_MESSAGE_NOT_CORRELATED"
	ErrorCodeTimerNotFound        ErrorCode = "ERR_TIMER_NOT_FOUND"
	ErrorCodeExpressionInvalid    ErrorCode = "ERR_EXPRESSION_INVALID"
	ErrorCodeIncidentActive       ErrorCode = "ERR_INCIDENT_ACTIVE"

	// Storage error codes
	ErrorCodeStorageConnection  ErrorCode = "ERR_STORAGE_CONNECTION"
	ErrorCodeStorageTransaction ErrorCode = "ERR_STORAGE_TRANSACTION"
	ErrorCodeStorageCorrupted   ErrorCode = "ERR_STORAGE_CORRUPTED"

	// Configuration error codes
	ErrorCodeConfigInvalid ErrorCode = "ERR_CONFIG_INVALID"
	ErrorCodeConfigMissing ErrorCode = "ERR_CONFIG_MISSING"
)

// CoreError represents a standardized error structure for the system
type CoreError struct {
	Type       ErrorType              `json:"type"`
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Severity   ErrorSeverity          `json:"severity"`
	Component  string                 `json:"component,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	Retryable  bool                   `json:"retryable"`
	InnerError *CoreError             `json:"inner_error,omitempty"`
	StackTrace []string               `json:"stack_trace,omitempty"`
	HTTPStatus int                    `json:"http_status,omitempty"`
}

// NewCoreError creates a new CoreError with basic information
func NewCoreError(errorType ErrorType, code ErrorCode, message string) *CoreError {
	return &CoreError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Severity:  ErrorSeverityMedium,
		Timestamp: time.Now(),
		Retryable: false,
	}
}

// NewSystemError creates a system-level error
func NewSystemError(code ErrorCode, message string) *CoreError {
	return NewCoreError(ErrorTypeSystem, code, message).
		WithSeverity(ErrorSeverityHigh).
		WithComponent("system")
}

// NewValidationError creates a validation error
func NewValidationError(message string, field string) *CoreError {
	err := NewCoreError(ErrorTypeValidation, ErrorCodeInvalidInput, message).
		WithSeverity(ErrorSeverityLow)
	if field != "" {
		err.Context = map[string]interface{}{"field": field}
	}
	return err
}

// NewBPMNError creates a BPMN-specific error
func NewBPMNError(code ErrorCode, message string) *CoreError {
	return NewCoreError(ErrorTypeBPMN, code, message).
		WithSeverity(ErrorSeverityMedium).
		WithComponent("bpmn")
}

// NewProcessError creates a process-specific error
func NewProcessError(code ErrorCode, message string, processInstanceID string) *CoreError {
	err := NewCoreError(ErrorTypeProcess, code, message).
		WithSeverity(ErrorSeverityMedium).
		WithComponent("process")
	if processInstanceID != "" {
		err.Context = map[string]interface{}{"process_instance_id": processInstanceID}
	}
	return err
}

// NewJobError creates a job-specific error
func NewJobError(code ErrorCode, message string, jobKey string) *CoreError {
	err := NewCoreError(ErrorTypeJob, code, message).
		WithSeverity(ErrorSeverityMedium).
		WithComponent("jobs")
	if jobKey != "" {
		err.Context = map[string]interface{}{"job_key": jobKey}
	}
	return err
}

// NewAuthError creates an authentication/authorization error
func NewAuthError(code ErrorCode, message string) *CoreError {
	return NewCoreError(ErrorTypeAuth, code, message).
		WithSeverity(ErrorSeverityHigh).
		WithComponent("auth").
		WithHTTPStatus(401)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string, id string) *CoreError {
	message := fmt.Sprintf("%s not found", resource)
	if id != "" {
		message = fmt.Sprintf("%s with id '%s' not found", resource, id)
	}

	err := NewCoreError(ErrorTypeValidation, ErrorCodeNotFound, message).
		WithSeverity(ErrorSeverityLow).
		WithHTTPStatus(404)

	if resource != "" {
		err.Context = map[string]interface{}{
			"resource": resource,
			"id":       id,
		}
	}
	return err
}

// Builder methods for CoreError
func (e *CoreError) WithSeverity(severity ErrorSeverity) *CoreError {
	e.Severity = severity
	return e
}

func (e *CoreError) WithComponent(component string) *CoreError {
	e.Component = component
	return e
}

func (e *CoreError) WithOperation(operation string) *CoreError {
	e.Operation = operation
	return e
}

func (e *CoreError) WithDetails(details string) *CoreError {
	e.Details = details
	return e
}

func (e *CoreError) WithContext(key string, value interface{}) *CoreError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func (e *CoreError) WithContextMap(context map[string]interface{}) *CoreError {
	e.Context = context
	return e
}

func (e *CoreError) WithRequestID(requestID string) *CoreError {
	e.RequestID = requestID
	return e
}

func (e *CoreError) WithUserID(userID string) *CoreError {
	e.UserID = userID
	return e
}

func (e *CoreError) WithTenantID(tenantID string) *CoreError {
	e.TenantID = tenantID
	return e
}

func (e *CoreError) WithRetryable(retryable bool) *CoreError {
	e.Retryable = retryable
	return e
}

func (e *CoreError) WithInnerError(innerError *CoreError) *CoreError {
	e.InnerError = innerError
	return e
}

func (e *CoreError) WithHTTPStatus(status int) *CoreError {
	e.HTTPStatus = status
	return e
}

func (e *CoreError) WithStackTrace(trace []string) *CoreError {
	e.StackTrace = trace
	return e
}

// Error implements the error interface
func (e *CoreError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s - %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// GetHTTPStatus returns appropriate HTTP status code
func (e *CoreError) GetHTTPStatus() int {
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}

	// Default HTTP status codes based on error type/code
	switch e.Code {
	case ErrorCodeNotFound:
		return 404
	case ErrorCodeUnauthorized:
		return 401
	case ErrorCodeForbidden:
		return 403
	case ErrorCodeInvalidInput:
		return 400
	case ErrorCodeRateLimit:
		return 429
	case ErrorCodeTimeout:
		return 408
	case ErrorCodeAlreadyExists:
		return 409
	default:
		switch e.Type {
		case ErrorTypeValidation, ErrorTypeFormat, ErrorTypeSchema:
			return 400
		case ErrorTypeAuth:
			return 401
		case ErrorTypePermission:
			return 403
		case ErrorTypeTimeout:
			return 408
		default:
			return 500
		}
	}
}

// IsCritical returns true if the error is critical
func (e *CoreError) IsCritical() bool {
	return e.Severity == ErrorSeverityCritical
}

// IsRetryable returns true if the operation can be retried
func (e *CoreError) IsRetryable() bool {
	return e.Retryable
}

// ToMap converts the error to a map for JSON serialization
func (e *CoreError) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"type":      e.Type,
		"code":      e.Code,
		"message":   e.Message,
		"severity":  e.Severity,
		"timestamp": e.Timestamp,
		"retryable": e.Retryable,
	}

	if e.Details != "" {
		result["details"] = e.Details
	}
	if e.Context != nil {
		result["context"] = e.Context
	}
	if e.Component != "" {
		result["component"] = e.Component
	}
	if e.Operation != "" {
		result["operation"] = e.Operation
	}
	if e.RequestID != "" {
		result["request_id"] = e.RequestID
	}
	if e.UserID != "" {
		result["user_id"] = e.UserID
	}
	if e.TenantID != "" {
		result["tenant_id"] = e.TenantID
	}
	if e.InnerError != nil {
		result["inner_error"] = e.InnerError.ToMap()
	}
	if len(e.StackTrace) > 0 {
		result["stack_trace"] = e.StackTrace
	}
	if e.HTTPStatus > 0 {
		result["http_status"] = e.HTTPStatus
	}

	return result
}

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Success      bool       `json:"success"`
	Error        *CoreError `json:"error"`
	ErrorMessage string     `json:"error_message"`
	RequestID    string     `json:"request_id,omitempty"`
	Timestamp    time.Time  `json:"timestamp"`
	Path         string     `json:"path,omitempty"`
	Method       string     `json:"method,omitempty"`
	StatusCode   int        `json:"status_code,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *CoreError) *ErrorResponse {
	return &ErrorResponse{
		Success:      false,
		Error:        err,
		ErrorMessage: err.Message,
		RequestID:    err.RequestID,
		Timestamp:    time.Now(),
		StatusCode:   err.GetHTTPStatus(),
	}
}

// ErrorStats represents statistics about errors in the system
type ErrorStats struct {
	TotalErrors       int64                   `json:"total_errors"`
	ErrorsByType      map[ErrorType]int64     `json:"errors_by_type"`
	ErrorsByCode      map[ErrorCode]int64     `json:"errors_by_code"`
	ErrorsBySeverity  map[ErrorSeverity]int64 `json:"errors_by_severity"`
	ErrorsByComponent map[string]int64        `json:"errors_by_component"`
	CriticalErrors    int64                   `json:"critical_errors"`
	RetryableErrors   int64                   `json:"retryable_errors"`
	LastError         *time.Time              `json:"last_error,omitempty"`
	ErrorRate         float64                 `json:"error_rate"`
}

// Helper functions for common error patterns
func WrapError(originalError error, errorType ErrorType, code ErrorCode, message string) *CoreError {
	coreErr := NewCoreError(errorType, code, message)
	if originalError != nil {
		coreErr.Details = originalError.Error()
	}
	return coreErr
}

func WrapSystemError(originalError error, message string) *CoreError {
	return WrapError(originalError, ErrorTypeSystem, ErrorCodeInternal, message).
		WithSeverity(ErrorSeverityHigh)
}

func WrapValidationError(originalError error, message string) *CoreError {
	return WrapError(originalError, ErrorTypeValidation, ErrorCodeInvalidInput, message).
		WithSeverity(ErrorSeverityLow)
}
