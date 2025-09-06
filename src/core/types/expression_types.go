package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// ExpressionValueType represents the type of an expression value
type ExpressionValueType string

const (
	ExpressionTypeString  ExpressionValueType = "string"
	ExpressionTypeNumber  ExpressionValueType = "number"
	ExpressionTypeBoolean ExpressionValueType = "boolean"
	ExpressionTypeNull    ExpressionValueType = "null"
	ExpressionTypeArray   ExpressionValueType = "array"
	ExpressionTypeObject  ExpressionValueType = "object"
	ExpressionTypeDate    ExpressionValueType = "date"
	ExpressionTypeError   ExpressionValueType = "error"
)

// ExpressionValue represents a strongly typed value from expression evaluation
type ExpressionValue struct {
	Type     ExpressionValueType `json:"type"`
	Value    interface{}         `json:"value"`
	RawValue interface{}         `json:"raw_value,omitempty"` // Original value before type conversion
	IsValid  bool                `json:"is_valid"`
	ErrorMsg string              `json:"error_message,omitempty"`
}

// NewExpressionValue creates a new ExpressionValue from interface{}
func NewExpressionValue(value interface{}) *ExpressionValue {
	if value == nil {
		return &ExpressionValue{
			Type:    ExpressionTypeNull,
			Value:   nil,
			IsValid: true,
		}
	}

	ev := &ExpressionValue{
		RawValue: value,
		IsValid:  true,
	}

	switch v := value.(type) {
	case string:
		ev.Type = ExpressionTypeString
		ev.Value = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ev.Type = ExpressionTypeNumber
		ev.Value = v
	case float32, float64:
		ev.Type = ExpressionTypeNumber
		ev.Value = v
	case bool:
		ev.Type = ExpressionTypeBoolean
		ev.Value = v
	case time.Time:
		ev.Type = ExpressionTypeDate
		ev.Value = v
	case []interface{}, []string, []int, []float64, []bool:
		ev.Type = ExpressionTypeArray
		ev.Value = v
	case map[string]interface{}:
		ev.Type = ExpressionTypeObject
		ev.Value = v
	case error:
		ev.Type = ExpressionTypeError
		ev.Value = v.Error()
		ev.ErrorMsg = v.Error()
		ev.IsValid = false
	default:
		// Try to handle unknown types
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			ev.Type = ExpressionTypeArray
			ev.Value = value
		case reflect.Map, reflect.Struct:
			ev.Type = ExpressionTypeObject
			ev.Value = value
		default:
			ev.Type = ExpressionTypeString
			ev.Value = fmt.Sprintf("%v", value)
		}
	}

	return ev
}

// String returns the string representation of the value
func (ev *ExpressionValue) String() string {
	if !ev.IsValid {
		return ev.ErrorMsg
	}
	return fmt.Sprintf("%v", ev.Value)
}

// AsString returns the value as a string
func (ev *ExpressionValue) AsString() (string, bool) {
	if !ev.IsValid {
		return "", false
	}
	if ev.Type == ExpressionTypeString {
		return ev.Value.(string), true
	}
	return fmt.Sprintf("%v", ev.Value), true
}

// AsNumber returns the value as a float64
func (ev *ExpressionValue) AsNumber() (float64, bool) {
	if !ev.IsValid || ev.Type != ExpressionTypeNumber {
		return 0, false
	}

	switch v := ev.Value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// AsBoolean returns the value as a boolean
func (ev *ExpressionValue) AsBoolean() (bool, bool) {
	if !ev.IsValid {
		return false, false
	}

	switch ev.Type {
	case ExpressionTypeBoolean:
		return ev.Value.(bool), true
	case ExpressionTypeString:
		str := ev.Value.(string)
		return str != "", true
	case ExpressionTypeNumber:
		num, ok := ev.AsNumber()
		return num != 0, ok
	case ExpressionTypeNull:
		return false, true
	default:
		return ev.Value != nil, true
	}
}

// AsArray returns the value as an array
func (ev *ExpressionValue) AsArray() ([]interface{}, bool) {
	if !ev.IsValid || ev.Type != ExpressionTypeArray {
		return nil, false
	}

	if arr, ok := ev.Value.([]interface{}); ok {
		return arr, true
	}
	return nil, false
}

// AsObject returns the value as a map
func (ev *ExpressionValue) AsObject() (map[string]interface{}, bool) {
	if !ev.IsValid || ev.Type != ExpressionTypeObject {
		return nil, false
	}

	if obj, ok := ev.Value.(map[string]interface{}); ok {
		return obj, true
	}
	return nil, false
}

// AsDate returns the value as a time.Time
func (ev *ExpressionValue) AsDate() (time.Time, bool) {
	if !ev.IsValid || ev.Type != ExpressionTypeDate {
		return time.Time{}, false
	}

	if date, ok := ev.Value.(time.Time); ok {
		return date, true
	}
	return time.Time{}, false
}

// ToJSON converts the expression value to JSON
func (ev *ExpressionValue) ToJSON() ([]byte, error) {
	return json.Marshal(ev.Value)
}

// ExpressionRequest represents a request to evaluate an expression
type ExpressionRequest struct {
	Expression string                 `json:"expression" validate:"required"`
	Context    map[string]interface{} `json:"context,omitempty"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	Schema     *ExpressionSchema      `json:"schema,omitempty"`
}

// ExpressionBatchRequest represents a request to evaluate multiple expressions
type ExpressionBatchRequest struct {
	Expressions []ExpressionItem       `json:"expressions" validate:"required,min=1"`
	Context     map[string]interface{} `json:"context,omitempty"`
	TenantID    string                 `json:"tenant_id,omitempty"`
}

// ExpressionItem represents a single expression in a batch
type ExpressionItem struct {
	ID         string `json:"id" validate:"required"`
	Expression string `json:"expression" validate:"required"`
}

// ExpressionResult represents the result of expression evaluation
type ExpressionResult struct {
	ID           string                 `json:"id,omitempty"`
	Expression   string                 `json:"expression"`
	Value        *ExpressionValue       `json:"value"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	EvaluatedAt  time.Time              `json:"evaluated_at"`
	Duration     time.Duration          `json:"duration"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// ExpressionBatchResult represents the result of batch expression evaluation
type ExpressionBatchResult struct {
	Results      []ExpressionResult `json:"results"`
	TotalCount   int                `json:"total_count"`
	SuccessCount int                `json:"success_count"`
	FailureCount int                `json:"failure_count"`
	Duration     time.Duration      `json:"duration"`
}

// ExpressionValidationRequest represents a request to validate an expression
type ExpressionValidationRequest struct {
	Expression string            `json:"expression" validate:"required"`
	Schema     *ExpressionSchema `json:"schema,omitempty"`
}

// ExpressionValidationResult represents the result of expression validation
type ExpressionValidationResult struct {
	IsValid      bool              `json:"is_valid"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Syntax       *ExpressionSyntax `json:"syntax,omitempty"`
	Variables    []string          `json:"variables,omitempty"`
}

// ExpressionSchema represents the schema for expression validation
type ExpressionSchema struct {
	RequiredVariables []string                       `json:"required_variables,omitempty"`
	VariableTypes     map[string]ExpressionValueType `json:"variable_types,omitempty"`
	ReturnType        ExpressionValueType            `json:"return_type,omitempty"`
}

// ExpressionSyntax represents the parsed syntax of an expression
type ExpressionSyntax struct {
	Tokens    []string `json:"tokens,omitempty"`
	Functions []string `json:"functions,omitempty"`
	Variables []string `json:"variables,omitempty"`
	Operators []string `json:"operators,omitempty"`
	IsComplex bool     `json:"is_complex"`
	Depth     int      `json:"depth"`
}

// ExpressionStats represents statistics about expression evaluations
type ExpressionStats struct {
	TotalEvaluations      int64                         `json:"total_evaluations"`
	SuccessfulEvaluations int64                         `json:"successful_evaluations"`
	FailedEvaluations     int64                         `json:"failed_evaluations"`
	AverageEvaluationTime time.Duration                 `json:"average_evaluation_time"`
	ExpressionsByType     map[ExpressionValueType]int64 `json:"expressions_by_type"`
	MostUsedVariables     map[string]int64              `json:"most_used_variables"`
	LastEvaluated         *time.Time                    `json:"last_evaluated,omitempty"`
}

// ExpressionFunctionInfo represents information about an available function
type ExpressionFunctionInfo struct {
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	Description string                `json:"description"`
	Parameters  []ExpressionParameter `json:"parameters"`
	ReturnType  ExpressionValueType   `json:"return_type"`
	Examples    []string              `json:"examples,omitempty"`
}

// ExpressionParameter represents a function parameter
type ExpressionParameter struct {
	Name        string              `json:"name"`
	Type        ExpressionValueType `json:"type"`
	Required    bool                `json:"required"`
	Description string              `json:"description,omitempty"`
}

// Helper functions for ExpressionResult
func (er *ExpressionResult) IsSuccessful() bool {
	return er.Success && er.Value != nil && er.Value.IsValid
}

func (er *ExpressionResult) GetStringValue() (string, bool) {
	if !er.IsSuccessful() {
		return "", false
	}
	return er.Value.AsString()
}

func (er *ExpressionResult) GetBooleanValue() (bool, bool) {
	if !er.IsSuccessful() {
		return false, false
	}
	return er.Value.AsBoolean()
}

func (er *ExpressionResult) GetNumericValue() (float64, bool) {
	if !er.IsSuccessful() {
		return 0, false
	}
	return er.Value.AsNumber()
}

// Helper functions for ExpressionBatchResult
func (ebr *ExpressionBatchResult) GetSuccessRate() float64 {
	if ebr.TotalCount == 0 {
		return 0
	}
	return float64(ebr.SuccessCount) / float64(ebr.TotalCount) * 100
}

func (ebr *ExpressionBatchResult) HasFailures() bool {
	return ebr.FailureCount > 0
}

func (ebr *ExpressionBatchResult) GetFailedResults() []ExpressionResult {
	var failed []ExpressionResult
	for _, result := range ebr.Results {
		if !result.Success {
			failed = append(failed, result)
		}
	}
	return failed
}
