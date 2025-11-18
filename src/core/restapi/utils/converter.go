/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"atom-engine/src/core/restapi/models"
)

// Converter provides conversion utilities between different formats
type Converter struct{}

// NewConverter creates new converter instance
func NewConverter() *Converter {
	return &Converter{}
}

// MapToVariables converts map[string]interface{} to gRPC variables format
func (c *Converter) MapToVariables(vars map[string]interface{}) (map[string]string, error) {
	if vars == nil {
		return nil, nil
	}

	result := make(map[string]string)
	for key, value := range vars {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variable %s: %w", key, err)
		}
		result[key] = string(jsonValue)
	}

	return result, nil
}

// VariablesToMap converts gRPC variables format to map[string]interface{}
func (c *Converter) VariablesToMap(vars map[string]string) (map[string]interface{}, error) {
	if vars == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for key, value := range vars {
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err != nil {
			// If not valid JSON, treat as string
			result[key] = value
		} else {
			result[key] = jsonValue
		}
	}

	return result, nil
}

// StringToInt32 converts string to int32 with validation
func (c *Converter) StringToInt32(value string) (int32, error) {
	if value == "" {
		return 0, nil
	}

	result, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", value)
	}

	return int32(result), nil
}

// StringToInt64 converts string to int64 with validation
func (c *Converter) StringToInt64(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}

	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", value)
	}

	return result, nil
}

// StringToBool converts string to bool with validation
func (c *Converter) StringToBool(value string) (bool, error) {
	if value == "" {
		return false, nil
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}

	return result, nil
}

// TimestampToTime converts Unix timestamp to time.Time
func (c *Converter) TimestampToTime(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0)
}

// TimeToTimestamp converts time.Time to Unix timestamp
func (c *Converter) TimeToTimestamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

// ParseTimestamp parses various timestamp formats
func (c *Converter) ParseTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid timestamp format: %s", value)
}

// FormatTimestamp formats time.Time to RFC3339 string
func (c *Converter) FormatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// HTTPStatusFromGRPCError converts gRPC error to HTTP status code
func (c *Converter) HTTPStatusFromGRPCError(err error) int {
	if err == nil {
		return 200
	}

	errMsg := err.Error()

	switch {
	case contains(errMsg, "not found"):
		return 404
	case contains(errMsg, "already exists"):
		return 409
	case contains(errMsg, "invalid"):
		return 400
	case contains(errMsg, "unauthorized"):
		return 401
	case contains(errMsg, "forbidden"):
		return 403
	case contains(errMsg, "rate limit"):
		return 429
	default:
		return 500
	}
}

// GRPCErrorToAPIError converts gRPC error to API error
func (c *Converter) GRPCErrorToAPIError(err error) *models.APIError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case contains(errMsg, "not found"):
		return models.NotFoundError(errMsg)
	case contains(errMsg, "already exists"):
		return models.ConflictError(errMsg)
	case contains(errMsg, "invalid"):
		return models.BadRequestError(errMsg)
	case contains(errMsg, "unauthorized"):
		return models.UnauthorizedError(errMsg)
	case contains(errMsg, "forbidden"):
		return models.ForbiddenError(errMsg)
	case contains(errMsg, "rate limit"):
		return models.RateLimitedError(errMsg)
	default:
		return models.InternalServerError(errMsg)
	}
}

// ConvertProcessStatus converts process status between formats
func (c *Converter) ConvertProcessStatus(status string) string {
	statusMap := map[string]string{
		"ACTIVE":    "active",
		"COMPLETED": "completed",
		"CANCELLED": "cancelled",
		"FAILED":    "failed",
	}

	if converted, exists := statusMap[status]; exists {
		return converted
	}

	return status
}

// ConvertJobState converts job state between formats
func (c *Converter) ConvertJobState(state string) string {
	stateMap := map[string]string{
		"ACTIVATABLE": "activatable",
		"ACTIVATED":   "activated",
		"COMPLETED":   "completed",
		"FAILED":      "failed",
		"CANCELLED":   "cancelled",
	}

	if converted, exists := stateMap[state]; exists {
		return converted
	}

	return state
}

// ConvertIncidentType converts incident type between formats
func (c *Converter) ConvertIncidentType(incidentType string) string {
	typeMap := map[string]string{
		"INCIDENT_TYPE_JOB_FAILURE":      "job_failure",
		"INCIDENT_TYPE_BPMN_ERROR":       "bpmn_error",
		"INCIDENT_TYPE_EXPRESSION_ERROR": "expression_error",
		"INCIDENT_TYPE_PROCESS_ERROR":    "process_error",
		"INCIDENT_TYPE_TIMER_ERROR":      "timer_error",
		"INCIDENT_TYPE_MESSAGE_ERROR":    "message_error",
		"INCIDENT_TYPE_SYSTEM_ERROR":     "system_error",
	}

	if converted, exists := typeMap[incidentType]; exists {
		return converted
	}

	return incidentType
}

// ConvertIncidentStatus converts incident status between formats
func (c *Converter) ConvertIncidentStatus(status string) string {
	statusMap := map[string]string{
		"INCIDENT_STATUS_OPEN":      "open",
		"INCIDENT_STATUS_RESOLVED":  "resolved",
		"INCIDENT_STATUS_DISMISSED": "dismissed",
	}

	if converted, exists := statusMap[status]; exists {
		return converted
	}

	return status
}

// IsValidJSON checks if string is valid JSON
func IsValidJSON(str string) bool {
	var js interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsIgnoreCase(s, substr)))
}

// containsIgnoreCase performs case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

// ConvertArrayToStringSlice converts []interface{} to []string
func (c *Converter) ConvertArrayToStringSlice(arr []interface{}) []string {
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}

// MergeStringMaps merges multiple string maps
func (c *Converter) MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
