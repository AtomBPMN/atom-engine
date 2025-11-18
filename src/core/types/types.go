// Package types provides strongly typed data structures for the Atom Engine core system.
// This package replaces the widespread use of interface{} with strict typing to improve
// type safety, code clarity, and development experience.
//
// The types are organized into several categories:
// - Job types (job_types.go): Job management and execution
// - Expression types (expression_types.go): Expression evaluation and FEEL
// - Message types (message_types.go): Message publishing and correlation
// - Process types (process_types.go): BPMN process and instance management
// - Error types (error_types.go): Unified error handling and reporting
// - Component types (component_types.go): System component status and health
package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Common type aliases for convenience
type (
	// Variables represents a collection of variables
	Variables = map[string]interface{}

	// Metadata represents metadata information
	Metadata = map[string]interface{}

	// Headers represents header information
	Headers = map[string]string

	// Properties represents configuration properties
	Properties = map[string]interface{}
)

// Common status values used across the system
const (
	// Generic statuses
	StatusActive    = "ACTIVE"
	StatusInactive  = "INACTIVE"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusStopped   = "STOPPED"
	StatusError     = "ERROR"

	// Health statuses
	HealthHealthy   = "HEALTHY"
	HealthUnhealthy = "UNHEALTHY"
	HealthDegraded  = "DEGRADED"
	HealthUnknown   = "UNKNOWN"
)

// Common time constants
const (
	DefaultTimeout      = 30 * time.Second
	DefaultRetryBackoff = 5 * time.Second
	DefaultTTL          = time.Hour
	MaxRetries          = 3
)

// Pagination represents pagination parameters
type Pagination struct {
	Limit  int32 `json:"limit,omitempty"`
	Offset int32 `json:"offset,omitempty"`
	Page   int32 `json:"page,omitempty"`
	Size   int32 `json:"size,omitempty"`
}

// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	CurrentPage int32 `json:"current_page"`
	TotalPages  int32 `json:"total_pages"`
	TotalCount  int32 `json:"total_count"`
	PageSize    int32 `json:"page_size"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}

// Filter represents a generic filter
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, gte, lte, in, like
	Value    interface{} `json:"value"`
}

// SearchOptions represents search and filtering options
type SearchOptions struct {
	Query      string        `json:"query,omitempty"`
	Filters    []Filter      `json:"filters,omitempty"`
	Sort       []SortOptions `json:"sort,omitempty"`
	Pagination Pagination    `json:"pagination,omitempty"`
	TimeRange  *TimeRange    `json:"time_range,omitempty"`
}

// OperationResult represents the result of an operation
type OperationResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Data        interface{}            `json:"data,omitempty"`
	Error       *CoreError             `json:"error,omitempty"`
	OperationID string                 `json:"operation_id,omitempty"`
	ExecutedAt  time.Time              `json:"executed_at"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchOperationResult represents the result of a batch operation
type BatchOperationResult struct {
	TotalCount   int32                  `json:"total_count"`
	SuccessCount int32                  `json:"success_count"`
	FailureCount int32                  `json:"failure_count"`
	Results      []OperationResult      `json:"results"`
	Summary      map[string]interface{} `json:"summary,omitempty"`
	ExecutedAt   time.Time              `json:"executed_at"`
	Duration     time.Duration          `json:"duration"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CheckedAt time.Time              `json:"checked_at"`
	Duration  time.Duration          `json:"duration"`
}

// SystemInfo represents general system information
type SystemInfo struct {
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	BuildTime     time.Time              `json:"build_time"`
	GitCommit     string                 `json:"git_commit,omitempty"`
	Environment   string                 `json:"environment"`
	StartedAt     time.Time              `json:"started_at"`
	Uptime        time.Duration          `json:"uptime"`
	HostInfo      HostInfo               `json:"host_info"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

// HostInfo represents host system information
type HostInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	CPUCores     int32  `json:"cpu_cores"`
	MemoryTotal  int64  `json:"memory_total"`
	DiskTotal    int64  `json:"disk_total"`
}

// APIInfo represents API information
type APIInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoints   []EndpointInfo    `json:"endpoints"`
	Schemas     map[string]string `json:"schemas,omitempty"`
}

// EndpointInfo represents API endpoint information
type EndpointInfo struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// Common utility functions

// NewPagination creates a new pagination with defaults
func NewPagination(limit, offset int32) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{
		Limit:  limit,
		Offset: offset,
		Page:   (offset / limit) + 1,
		Size:   limit,
	}
}

// NewPaginationInfo calculates pagination information
func NewPaginationInfo(totalCount, limit, offset int32) PaginationInfo {
	if limit <= 0 {
		limit = 10
	}

	totalPages := (totalCount + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		PageSize:    limit,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}
}

// NewTimeRange creates a new time range
func NewTimeRange(start, end time.Time) *TimeRange {
	return &TimeRange{
		StartTime: &start,
		EndTime:   &end,
	}
}

// NewOperationResult creates a successful operation result
func NewOperationResult(message string, data interface{}) *OperationResult {
	return &OperationResult{
		Success:    true,
		Message:    message,
		Data:       data,
		ExecutedAt: time.Now(),
	}
}

// NewOperationError creates a failed operation result
func NewOperationError(message string, err *CoreError) *OperationResult {
	return &OperationResult{
		Success:    false,
		Message:    message,
		Error:      err,
		ExecutedAt: time.Now(),
	}
}

// NewSystemInfo creates system information
func NewSystemInfo(name, version, environment string) *SystemInfo {
	now := time.Now()
	return &SystemInfo{
		Name:        name,
		Version:     version,
		Environment: environment,
		StartedAt:   now,
		Uptime:      0,
		HostInfo: HostInfo{
			OS:           "linux",
			Architecture: "amd64",
		},
		Configuration: make(map[string]interface{}),
	}
}

// JSON utility functions

// ToJSON converts any value to JSON string
func ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON converts JSON string to specified type
func FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// ToPrettyJSON converts any value to pretty-formatted JSON string
func ToPrettyJSON(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Time utility functions

// FormatDuration formats duration in human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// IsValidTimeRange checks if time range is valid
func (tr *TimeRange) IsValid() bool {
	if tr.StartTime == nil || tr.EndTime == nil {
		return false
	}
	return tr.StartTime.Before(*tr.EndTime)
}

// Contains checks if time is within the range
func (tr *TimeRange) Contains(t time.Time) bool {
	if !tr.IsValid() {
		return false
	}
	return (tr.StartTime.Before(t) || tr.StartTime.Equal(t)) &&
		(tr.EndTime.After(t) || tr.EndTime.Equal(t))
}

// Validation utility functions

// IsEmptyString checks if string is empty or contains only whitespace
func IsEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsValidID checks if ID format is valid (basic validation)
func IsValidID(id string) bool {
	return !IsEmptyString(id) && len(id) >= 3 && len(id) <= 100
}

// IsValidEmail checks if email format is valid (basic validation)
func IsValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// IsValidURL checks if URL format is valid (basic validation)
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// Helper functions for common operations

// GetStringFromMap safely gets string value from map
func GetStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetIntFromMap safely gets int value from map
func GetIntFromMap(m map[string]interface{}, key string, defaultValue int64) int64 {
	if val, exists := m[key]; exists {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

// GetBoolFromMap safely gets bool value from map
func GetBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := m[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// MergeVariables merges two variable maps
func MergeVariables(base, override Variables) Variables {
	result := make(Variables)

	// Copy base variables
	for k, v := range base {
		result[k] = v
	}

	// Override with new variables
	for k, v := range override {
		result[k] = v
	}

	return result
}

// CloneVariables creates a deep copy of variables
func CloneVariables(vars Variables) Variables {
	result := make(Variables)
	for k, v := range vars {
		result[k] = v
	}
	return result
}
