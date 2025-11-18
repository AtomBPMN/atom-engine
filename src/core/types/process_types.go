package types

import (
	"time"
)

// ProcessStatus represents the status of a process instance
type ProcessStatus string

const (
	ProcessStatusActive    ProcessStatus = "ACTIVE"
	ProcessStatusCompleted ProcessStatus = "COMPLETED"
	ProcessStatusCancelled ProcessStatus = "CANCELLED"
	ProcessStatusFailed    ProcessStatus = "FAILED"
	ProcessStatusSuspended ProcessStatus = "SUSPENDED"
)

// ProcessElementType represents the type of a BPMN process element
type ProcessElementType string

const (
	ElementTypeStartEvent             ProcessElementType = "startEvent"
	ElementTypeEndEvent               ProcessElementType = "endEvent"
	ElementTypeIntermediateCatchEvent ProcessElementType = "intermediateCatchEvent"
	ElementTypeIntermediateThrowEvent ProcessElementType = "intermediateThrowEvent"
	ElementTypeBoundaryEvent          ProcessElementType = "boundaryEvent"
	ElementTypeServiceTask            ProcessElementType = "serviceTask"
	ElementTypeUserTask               ProcessElementType = "userTask"
	ElementTypeScriptTask             ProcessElementType = "scriptTask"
	ElementTypeSendTask               ProcessElementType = "sendTask"
	ElementTypeReceiveTask            ProcessElementType = "receiveTask"
	ElementTypeCallActivity           ProcessElementType = "callActivity"
	ElementTypeSubProcess             ProcessElementType = "subProcess"
	ElementTypeExclusiveGateway       ProcessElementType = "exclusiveGateway"
	ElementTypeParallelGateway        ProcessElementType = "parallelGateway"
	ElementTypeInclusiveGateway       ProcessElementType = "inclusiveGateway"
	ElementTypeEventBasedGateway      ProcessElementType = "eventBasedGateway"
	ElementTypeSequenceFlow           ProcessElementType = "sequenceFlow"
)

// ProcessVariables represents variables associated with a process instance
type ProcessVariables map[string]interface{}

// ProcessMetadata represents metadata for a process
type ProcessMetadata map[string]interface{}

// ProcessElementInfo represents information about a process element
type ProcessElementInfo struct {
	ElementID  string             `json:"element_id"`
	Type       ProcessElementType `json:"type"`
	Name       string             `json:"name,omitempty"`
	Properties ProcessMetadata    `json:"properties,omitempty"`
	Incoming   []string           `json:"incoming,omitempty"`
	Outgoing   []string           `json:"outgoing,omitempty"`
	AttachedTo string             `json:"attached_to,omitempty"`
}

// ProcessInstanceDetails represents detailed information about a process instance
type ProcessInstanceDetails struct {
	InstanceID          string           `json:"instance_id"`
	ProcessKey          string           `json:"process_key"`
	ProcessDefinitionID string           `json:"process_definition_id"`
	Version             int32            `json:"version"`
	Status              ProcessStatus    `json:"status"`
	Variables           ProcessVariables `json:"variables,omitempty"`
	Metadata            ProcessMetadata  `json:"metadata,omitempty"`
	TenantID            string           `json:"tenant_id,omitempty"`
	ParentInstanceID    string           `json:"parent_instance_id,omitempty"`
	CalledFromActivity  string           `json:"called_from_activity,omitempty"`
	StartedAt           time.Time        `json:"started_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	CompletedAt         *time.Time       `json:"completed_at,omitempty"`
	Duration            *time.Duration   `json:"duration,omitempty"`
	CurrentActivity     string           `json:"current_activity,omitempty"`
	ActiveTokens        int32            `json:"active_tokens"`
	CompletedTokens     int32            `json:"completed_tokens"`
	ErrorMessage        string           `json:"error_message,omitempty"`
}

// ProcessDefinitionInfo represents information about a process definition
type ProcessDefinitionInfo struct {
	ProcessKey          string               `json:"process_key"`
	ProcessDefinitionID string               `json:"process_definition_id"`
	Name                string               `json:"name,omitempty"`
	Version             int32                `json:"version"`
	VersionTag          string               `json:"version_tag,omitempty"`
	TenantID            string               `json:"tenant_id,omitempty"`
	Elements            []ProcessElementInfo `json:"elements,omitempty"`
	ElementsCount       int32                `json:"elements_count"`
	IsExecutable        bool                 `json:"is_executable"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	ContentHash         string               `json:"content_hash,omitempty"`
	FileName            string               `json:"file_name,omitempty"`
	Metadata            ProcessMetadata      `json:"metadata,omitempty"`
}

// TokenInfo represents information about a process token
type TokenInfo struct {
	TokenID           string           `json:"token_id"`
	ProcessInstanceID string           `json:"process_instance_id"`
	ElementID         string           `json:"element_id"`
	Status            string           `json:"status"`
	Variables         ProcessVariables `json:"variables,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty"`
	ParentTokenID     string           `json:"parent_token_id,omitempty"`
	IsActive          bool             `json:"is_active"`
}

// ProcessStats represents statistics about processes in the system
type ProcessStats struct {
	TotalInstances        int64                   `json:"total_instances"`
	ActiveInstances       int64                   `json:"active_instances"`
	CompletedInstances    int64                   `json:"completed_instances"`
	CancelledInstances    int64                   `json:"cancelled_instances"`
	FailedInstances       int64                   `json:"failed_instances"`
	SuspendedInstances    int64                   `json:"suspended_instances"`
	TotalDefinitions      int64                   `json:"total_definitions"`
	ActiveDefinitions     int64                   `json:"active_definitions"`
	InstancesByProcess    map[string]int64        `json:"instances_by_process"`
	InstancesByStatus     map[ProcessStatus]int64 `json:"instances_by_status"`
	InstancesByTenant     map[string]int64        `json:"instances_by_tenant"`
	AverageExecutionTime  time.Duration           `json:"average_execution_time"`
	TotalTokens           int64                   `json:"total_tokens"`
	ActiveTokens          int64                   `json:"active_tokens"`
	CompletedTokens       int64                   `json:"completed_tokens"`
	LastInstanceStarted   *time.Time              `json:"last_instance_started,omitempty"`
	LastInstanceCompleted *time.Time              `json:"last_instance_completed,omitempty"`
}

// ProcessStartRequest represents a request to start a process instance
type ProcessStartRequest struct {
	ProcessKey        string           `json:"process_key" validate:"required"`
	Version           *int32           `json:"version,omitempty"`
	Variables         ProcessVariables `json:"variables,omitempty"`
	TenantID          string           `json:"tenant_id,omitempty"`
	BusinessKey       string           `json:"business_key,omitempty"`
	StartInstructions []string         `json:"start_instructions,omitempty"`
}

// ProcessStartResponse represents the response from starting a process
type ProcessStartResponse struct {
	InstanceID string           `json:"instance_id"`
	ProcessKey string           `json:"process_key"`
	Version    int32            `json:"version"`
	Status     ProcessStatus    `json:"status"`
	Success    bool             `json:"success"`
	Message    string           `json:"message"`
	StartedAt  time.Time        `json:"started_at"`
	Variables  ProcessVariables `json:"variables,omitempty"`
}

// ProcessCancelRequest represents a request to cancel a process instance
type ProcessCancelRequest struct {
	InstanceID string `json:"instance_id" validate:"required"`
	Reason     string `json:"reason,omitempty"`
	Force      bool   `json:"force,omitempty"`
}

// ProcessCancelResponse represents the response from cancelling a process
type ProcessCancelResponse struct {
	InstanceID  string    `json:"instance_id"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// ProcessListRequest represents a request to list process instances
type ProcessListRequest struct {
	ProcessKey    *string        `json:"process_key,omitempty"`
	Status        *ProcessStatus `json:"status,omitempty"`
	TenantID      *string        `json:"tenant_id,omitempty"`
	BusinessKey   *string        `json:"business_key,omitempty"`
	StartedAfter  *time.Time     `json:"started_after,omitempty"`
	StartedBefore *time.Time     `json:"started_before,omitempty"`
	Limit         int32          `json:"limit,omitempty"`
	Offset        int32          `json:"offset,omitempty"`
}

// ProcessListResponse represents a response with a list of process instances
type ProcessListResponse struct {
	Instances  []ProcessInstanceDetails `json:"instances"`
	TotalCount int32                    `json:"total_count"`
	HasMore    bool                     `json:"has_more"`
}

// ProcessDefinitionListRequest represents a request to list process definitions
type ProcessDefinitionListRequest struct {
	ProcessKey     *string `json:"process_key,omitempty"`
	TenantID       *string `json:"tenant_id,omitempty"`
	VersionTag     *string `json:"version_tag,omitempty"`
	LatestOnly     bool    `json:"latest_only,omitempty"`
	ExecutableOnly bool    `json:"executable_only,omitempty"`
	Limit          int32   `json:"limit,omitempty"`
	Offset         int32   `json:"offset,omitempty"`
}

// ProcessDefinitionListResponse represents a response with process definitions
type ProcessDefinitionListResponse struct {
	Definitions []ProcessDefinitionInfo `json:"definitions"`
	TotalCount  int32                   `json:"total_count"`
	HasMore     bool                    `json:"has_more"`
}

// TokenListRequest represents a request to list tokens
type TokenListRequest struct {
	ProcessInstanceID *string `json:"process_instance_id,omitempty"`
	ElementID         *string `json:"element_id,omitempty"`
	Status            *string `json:"status,omitempty"`
	ActiveOnly        bool    `json:"active_only,omitempty"`
	Limit             int32   `json:"limit,omitempty"`
	Offset            int32   `json:"offset,omitempty"`
}

// TokenListResponse represents a response with a list of tokens
type TokenListResponse struct {
	Tokens     []TokenInfo `json:"tokens"`
	TotalCount int32       `json:"total_count"`
	HasMore    bool        `json:"has_more"`
}

// ProcessTraceRequest represents a request to trace process execution
type ProcessTraceRequest struct {
	ProcessInstanceID string `json:"process_instance_id" validate:"required"`
	IncludeVariables  bool   `json:"include_variables,omitempty"`
	IncludeMetadata   bool   `json:"include_metadata,omitempty"`
}

// ProcessTraceResponse represents the execution trace of a process
type ProcessTraceResponse struct {
	ProcessInstanceID string         `json:"process_instance_id"`
	ProcessKey        string         `json:"process_key"`
	Status            ProcessStatus  `json:"status"`
	Tokens            []TokenInfo    `json:"tokens"`
	ExecutionPath     []string       `json:"execution_path"`
	StartedAt         time.Time      `json:"started_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
	Duration          *time.Duration `json:"duration,omitempty"`
	TotalTokens       int32          `json:"total_tokens"`
	CompletedTokens   int32          `json:"completed_tokens"`
}

// Helper methods for ProcessVariables
func (pv ProcessVariables) GetString(key string) (string, bool) {
	if val, exists := pv[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (pv ProcessVariables) GetInt64(key string) (int64, bool) {
	if val, exists := pv[key]; exists {
		switch v := val.(type) {
		case int64:
			return v, true
		case int:
			return int64(v), true
		case float64:
			return int64(v), true
		}
	}
	return 0, false
}

func (pv ProcessVariables) GetBool(key string) (bool, bool) {
	if val, exists := pv[key]; exists {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

func (pv ProcessVariables) Set(key string, value interface{}) {
	pv[key] = value
}

func (pv ProcessVariables) Delete(key string) {
	delete(pv, key)
}

func (pv ProcessVariables) Keys() []string {
	keys := make([]string, 0, len(pv))
	for k := range pv {
		keys = append(keys, k)
	}
	return keys
}

// Helper methods for ProcessInstanceDetails
func (pid *ProcessInstanceDetails) IsActive() bool {
	return pid.Status == ProcessStatusActive
}

func (pid *ProcessInstanceDetails) IsCompleted() bool {
	return pid.Status == ProcessStatusCompleted
}

func (pid *ProcessInstanceDetails) IsCancelled() bool {
	return pid.Status == ProcessStatusCancelled
}

func (pid *ProcessInstanceDetails) IsFailed() bool {
	return pid.Status == ProcessStatusFailed
}

func (pid *ProcessInstanceDetails) IsSuspended() bool {
	return pid.Status == ProcessStatusSuspended
}

func (pid *ProcessInstanceDetails) IsFinished() bool {
	return pid.Status == ProcessStatusCompleted ||
		pid.Status == ProcessStatusCancelled ||
		pid.Status == ProcessStatusFailed
}

func (pid *ProcessInstanceDetails) GetExecutionTime() time.Duration {
	if pid.Duration != nil {
		return *pid.Duration
	}
	if pid.CompletedAt != nil {
		return pid.CompletedAt.Sub(pid.StartedAt)
	}
	return time.Since(pid.StartedAt)
}

func (pid *ProcessInstanceDetails) GetTokenProgress() float64 {
	total := pid.ActiveTokens + pid.CompletedTokens
	if total == 0 {
		return 0
	}
	return float64(pid.CompletedTokens) / float64(total) * 100
}

// Helper methods for ProcessStats
func (ps *ProcessStats) GetCompletionRate() float64 {
	if ps.TotalInstances == 0 {
		return 0
	}
	return float64(ps.CompletedInstances) / float64(ps.TotalInstances) * 100
}

func (ps *ProcessStats) GetFailureRate() float64 {
	if ps.TotalInstances == 0 {
		return 0
	}
	return float64(ps.FailedInstances) / float64(ps.TotalInstances) * 100
}

func (ps *ProcessStats) GetActiveRate() float64 {
	if ps.TotalInstances == 0 {
		return 0
	}
	return float64(ps.ActiveInstances) / float64(ps.TotalInstances) * 100
}

func (ps *ProcessStats) GetTokenCompletionRate() float64 {
	if ps.TotalTokens == 0 {
		return 0
	}
	return float64(ps.CompletedTokens) / float64(ps.TotalTokens) * 100
}

// Helper methods for TokenInfo
func (ti *TokenInfo) IsCompleted() bool {
	return ti.CompletedAt != nil
}

func (ti *TokenInfo) GetExecutionTime() time.Duration {
	if ti.CompletedAt != nil {
		return ti.CompletedAt.Sub(ti.CreatedAt)
	}
	return time.Since(ti.CreatedAt)
}
