/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"atom-engine/src/core/restapi/models"
)

// Validator provides request validation utilities
type Validator struct{}

// NewValidator creates new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired validates required fields
func (v *Validator) ValidateRequired(value interface{}, fieldName string) *models.ValidationError {
	if value == nil {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s is required", fieldName),
		}
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			return &models.ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("%s cannot be empty", fieldName),
			}
		}
	case []interface{}:
		if len(val) == 0 {
			return &models.ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("%s cannot be empty", fieldName),
			}
		}
	case map[string]interface{}:
		if len(val) == 0 {
			return &models.ValidationError{
				Field:   fieldName,
				Value:   value,
				Message: fmt.Sprintf("%s cannot be empty", fieldName),
			}
		}
	}

	return nil
}

// ValidateStringLength validates string length constraints
func (v *Validator) ValidateStringLength(value string, fieldName string, minLen, maxLen int) *models.ValidationError {
	length := utf8.RuneCountInString(value)

	if minLen > 0 && length < minLen {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be at least %d characters long", fieldName, minLen),
		}
	}

	if maxLen > 0 && length > maxLen {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be at most %d characters long", fieldName, maxLen),
		}
	}

	return nil
}

// ValidatePattern validates string against regex pattern
func (v *Validator) ValidatePattern(value, fieldName, pattern, patternName string) *models.ValidationError {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("invalid pattern validation for %s", fieldName),
		}
	}

	if !matched {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must match %s format", fieldName, patternName),
		}
	}

	return nil
}

// ValidateID validates ID format (NanoID)
func (v *Validator) ValidateID(value, fieldName string) *models.ValidationError {
	// NanoID pattern: 4-char prefix + hyphen + 18-char NanoID
	pattern := `^[a-zA-Z0-9]{4}-[a-zA-Z0-9_-]{18}$`
	return v.ValidatePattern(value, fieldName, pattern, "ID")
}

// ValidateProcessKey validates process key format
func (v *Validator) ValidateProcessKey(value, fieldName string) *models.ValidationError {
	if err := v.ValidateRequired(value, fieldName); err != nil {
		return err
	}

	if err := v.ValidateStringLength(value, fieldName, 1, 255); err != nil {
		return err
	}

	// Process key can contain letters, numbers, underscores, hyphens
	pattern := `^[a-zA-Z0-9_-]+$`
	return v.ValidatePattern(value, fieldName, pattern, "process key")
}

// ValidateISO8601Duration validates ISO 8601 duration format
func (v *Validator) ValidateISO8601Duration(value, fieldName string) *models.ValidationError {
	if value == "" {
		return nil // Allow empty for optional fields
	}

	// Basic ISO 8601 duration pattern
	patterns := []string{
		`^P(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`, // PT30S, P1D, etc.
		`^R\d*/P.*$`, // R5/PT30S - repeating
		`^R/P.*$`,    // R/PT30S - infinite repeating
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   fieldName,
		Value:   value,
		Message: fmt.Sprintf("%s must be valid ISO 8601 duration format (e.g., PT30S, P1D, R5/PT30S)", fieldName),
	}
}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(value, fieldName string) *models.ValidationError {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return v.ValidatePattern(value, fieldName, pattern, "email")
}

// ValidateURL validates URL format
func (v *Validator) ValidateURL(value, fieldName string) *models.ValidationError {
	if value == "" {
		return nil // Allow empty for optional fields
	}

	_, err := url.Parse(value)
	if err != nil {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be a valid URL", fieldName),
		}
	}

	return nil
}

// ValidateEnum validates value against allowed enum values
func (v *Validator) ValidateEnum(
	value interface{},
	fieldName string,
	allowedValues []interface{},
) *models.ValidationError {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   fieldName,
		Value:   value,
		Message: fmt.Sprintf("%s must be one of: %v", fieldName, allowedValues),
	}
}

// ValidateStringEnum validates string value against allowed enum values
func (v *Validator) ValidateStringEnum(value, fieldName string, allowedValues []string) *models.ValidationError {
	if value == "" {
		return nil // Allow empty for optional fields
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   fieldName,
		Value:   value,
		Message: fmt.Sprintf("%s must be one of: %v", fieldName, allowedValues),
	}
}

// ValidateRange validates numeric value range
func (v *Validator) ValidateRange(value interface{}, fieldName string, min, max float64) *models.ValidationError {
	var numValue float64

	switch val := value.(type) {
	case int:
		numValue = float64(val)
	case int32:
		numValue = float64(val)
	case int64:
		numValue = float64(val)
	case float32:
		numValue = float64(val)
	case float64:
		numValue = val
	default:
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be a numeric value", fieldName),
		}
	}

	if numValue < min {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be at least %g", fieldName, min),
		}
	}

	if numValue > max {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be at most %g", fieldName, max),
		}
	}

	return nil
}

// ValidateTimestamp validates timestamp format
func (v *Validator) ValidateTimestamp(value, fieldName string) *models.ValidationError {
	if value == "" {
		return nil // Allow empty for optional fields
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if _, err := time.Parse(layout, value); err == nil {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   fieldName,
		Value:   value,
		Message: fmt.Sprintf("%s must be a valid timestamp (RFC3339 format)", fieldName),
	}
}

// ValidateJSON validates JSON string
func (v *Validator) ValidateJSON(value, fieldName string) *models.ValidationError {
	if value == "" {
		return nil // Allow empty for optional fields
	}

	// Try to parse as JSON
	if !IsValidJSON(value) {
		return &models.ValidationError{
			Field:   fieldName,
			Value:   value,
			Message: fmt.Sprintf("%s must be valid JSON", fieldName),
		}
	}

	return nil
}

// ValidateMultiple validates multiple constraints and returns all errors
func (v *Validator) ValidateMultiple(validations ...func() *models.ValidationError) []models.ValidationError {
	var errors []models.ValidationError

	for _, validation := range validations {
		if err := validation(); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// CreateValidationError creates validation error response
func (v *Validator) CreateValidationError(errors []models.ValidationError) *models.APIError {
	if len(errors) == 0 {
		return nil
	}

	message := fmt.Sprintf("Validation failed for %d field(s)", len(errors))
	return models.NewValidationError(message, errors)
}
