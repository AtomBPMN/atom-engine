package types

import (
	"time"
)

// ComponentStatus represents the status of a system component
type ComponentStatus string

const (
	ComponentStatusStarting    ComponentStatus = "STARTING"
	ComponentStatusReady       ComponentStatus = "READY"
	ComponentStatusRunning     ComponentStatus = "RUNNING"
	ComponentStatusStopping    ComponentStatus = "STOPPING"
	ComponentStatusStopped     ComponentStatus = "STOPPED"
	ComponentStatusError       ComponentStatus = "ERROR"
	ComponentStatusDegraded    ComponentStatus = "DEGRADED"
	ComponentStatusMaintenance ComponentStatus = "MAINTENANCE"
)

// ComponentType represents the type of a system component
type ComponentType string

const (
	ComponentTypeCore       ComponentType = "CORE"
	ComponentTypeAuth       ComponentType = "AUTH"
	ComponentTypeConfig     ComponentType = "CONFIG"
	ComponentTypeLogger     ComponentType = "LOGGER"
	ComponentTypeStorage    ComponentType = "STORAGE"
	ComponentTypeGRPC       ComponentType = "GRPC"
	ComponentTypeRESTAPI    ComponentType = "RESTAPI"
	ComponentTypeProcess    ComponentType = "PROCESS"
	ComponentTypeParser     ComponentType = "PARSER"
	ComponentTypeJobs       ComponentType = "JOBS"
	ComponentTypeMessages   ComponentType = "MESSAGES"
	ComponentTypeTimewheel  ComponentType = "TIMEWHEEL"
	ComponentTypeExpression ComponentType = "EXPRESSION"
	ComponentTypeIncidents  ComponentType = "INCIDENTS"
)

// ComponentHealth represents the health status of a component
type ComponentHealth string

const (
	ComponentHealthHealthy   ComponentHealth = "HEALTHY"
	ComponentHealthUnhealthy ComponentHealth = "UNHEALTHY"
	ComponentHealthUnknown   ComponentHealth = "UNKNOWN"
	ComponentHealthDegraded  ComponentHealth = "DEGRADED"
)

// ComponentInfo represents detailed information about a system component
type ComponentInfo struct {
	Name            string                 `json:"name"`
	Type            ComponentType          `json:"type"`
	Status          ComponentStatus        `json:"status"`
	Health          ComponentHealth        `json:"health"`
	Version         string                 `json:"version,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	Endpoints       []ComponentEndpoint    `json:"endpoints,omitempty"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
	Metrics         ComponentMetrics       `json:"metrics,omitempty"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	LastHealthCheck *time.Time             `json:"last_health_check,omitempty"`
	Uptime          *time.Duration         `json:"uptime,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	LastError       *time.Time             `json:"last_error,omitempty"`
	RestartCount    int32                  `json:"restart_count"`
	IsEnabled       bool                   `json:"is_enabled"`
	ReadyFlag       bool                   `json:"is_ready"`
	Tags            []string               `json:"tags,omitempty"`
}

// ComponentEndpoint represents an endpoint exposed by a component
type ComponentEndpoint struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"` // HTTP, gRPC, WebSocket, etc.
	Address   string            `json:"address"`
	Port      int32             `json:"port,omitempty"`
	Path      string            `json:"path,omitempty"`
	IsSecure  bool              `json:"is_secure"`
	IsHealthy bool              `json:"is_healthy"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ComponentMetrics represents metrics for a component
type ComponentMetrics struct {
	RequestCount        int64         `json:"request_count"`
	ErrorCount          int64         `json:"error_count"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastRequestAt       *time.Time    `json:"last_request_at,omitempty"`
	MemoryUsage         int64         `json:"memory_usage,omitempty"`
	CPUUsage            float64       `json:"cpu_usage,omitempty"`
	ConnectionCount     int32         `json:"connection_count,omitempty"`
	QueueSize           int32         `json:"queue_size,omitempty"`
	ThroughputPerSecond float64       `json:"throughput_per_second,omitempty"`
}

// SystemStatus represents the overall status of the system
type SystemStatus struct {
	Status          ComponentStatus        `json:"status"`
	Health          ComponentHealth        `json:"health"`
	Version         string                 `json:"version"`
	StartedAt       time.Time              `json:"started_at"`
	Uptime          time.Duration          `json:"uptime"`
	Components      []ComponentInfo        `json:"components"`
	ComponentsTotal int32                  `json:"components_total"`
	ComponentsReady int32                  `json:"components_ready"`
	ComponentsError int32                  `json:"components_error"`
	LastHealthCheck time.Time              `json:"last_health_check"`
	SystemMetrics   SystemMetrics          `json:"system_metrics"`
	Configuration   map[string]interface{} `json:"configuration,omitempty"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	TotalErrors         int64         `json:"total_errors"`
	ErrorRate           float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`
	MemoryUsage         int64         `json:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage"`
	DiskUsage           int64         `json:"disk_usage"`
	NetworkIn           int64         `json:"network_in"`
	NetworkOut          int64         `json:"network_out"`
	ActiveConnections   int32         `json:"active_connections"`
	Goroutines          int32         `json:"goroutines"`
}

// ComponentStartRequest represents a request to start a component
type ComponentStartRequest struct {
	ComponentName string                 `json:"component_name" validate:"required"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Force         bool                   `json:"force,omitempty"`
}

// ComponentStopRequest represents a request to stop a component
type ComponentStopRequest struct {
	ComponentName string         `json:"component_name" validate:"required"`
	Graceful      bool           `json:"graceful,omitempty"`
	Timeout       *time.Duration `json:"timeout,omitempty"`
}

// ComponentRestartRequest represents a request to restart a component
type ComponentRestartRequest struct {
	ComponentName string                 `json:"component_name" validate:"required"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Graceful      bool                   `json:"graceful,omitempty"`
}

// ComponentOperationResponse represents the response from component operations
type ComponentOperationResponse struct {
	ComponentName string          `json:"component_name"`
	Operation     string          `json:"operation"`
	Success       bool            `json:"success"`
	Message       string          `json:"message"`
	Status        ComponentStatus `json:"status"`
	ExecutedAt    time.Time       `json:"executed_at"`
	Duration      time.Duration   `json:"duration"`
}

// ComponentHealthCheckRequest represents a request to check component health
type ComponentHealthCheckRequest struct {
	ComponentName string         `json:"component_name,omitempty"`
	Deep          bool           `json:"deep,omitempty"`
	Timeout       *time.Duration `json:"timeout,omitempty"`
}

// ComponentHealthCheckResponse represents the response from health check
type ComponentHealthCheckResponse struct {
	ComponentName string          `json:"component_name,omitempty"`
	Health        ComponentHealth `json:"health"`
	Status        ComponentStatus `json:"status"`
	Message       string          `json:"message,omitempty"`
	Checks        []HealthCheck   `json:"checks,omitempty"`
	CheckedAt     time.Time       `json:"checked_at"`
	Duration      time.Duration   `json:"duration"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Name     string                 `json:"name"`
	Status   ComponentHealth        `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComponentListRequest represents a request to list components
type ComponentListRequest struct {
	Type           *ComponentType   `json:"type,omitempty"`
	Status         *ComponentStatus `json:"status,omitempty"`
	Health         *ComponentHealth `json:"health,omitempty"`
	EnabledOnly    bool             `json:"enabled_only,omitempty"`
	ReadyOnly      bool             `json:"ready_only,omitempty"`
	IncludeMetrics bool             `json:"include_metrics,omitempty"`
}

// ComponentListResponse represents a response with component list
type ComponentListResponse struct {
	Components []ComponentInfo  `json:"components"`
	TotalCount int32            `json:"total_count"`
	Summary    ComponentSummary `json:"summary"`
}

// ComponentSummary represents a summary of component statuses
type ComponentSummary struct {
	Total     int32                     `json:"total"`
	ByStatus  map[ComponentStatus]int32 `json:"by_status"`
	ByHealth  map[ComponentHealth]int32 `json:"by_health"`
	ByType    map[ComponentType]int32   `json:"by_type"`
	Ready     int32                     `json:"ready"`
	Enabled   int32                     `json:"enabled"`
	HasErrors int32                     `json:"has_errors"`
}

// Helper methods for ComponentInfo
func (ci *ComponentInfo) IsStarting() bool {
	return ci.Status == ComponentStatusStarting
}

func (ci *ComponentInfo) IsReady() bool {
	return ci.Status == ComponentStatusReady
}

func (ci *ComponentInfo) IsRunning() bool {
	return ci.Status == ComponentStatusRunning
}

func (ci *ComponentInfo) IsStopped() bool {
	return ci.Status == ComponentStatusStopped
}

func (ci *ComponentInfo) HasError() bool {
	return ci.Status == ComponentStatusError
}

func (ci *ComponentInfo) IsDegraded() bool {
	return ci.Status == ComponentStatusDegraded
}

func (ci *ComponentInfo) IsHealthy() bool {
	return ci.Health == ComponentHealthHealthy
}

func (ci *ComponentInfo) IsUnhealthy() bool {
	return ci.Health == ComponentHealthUnhealthy
}

func (ci *ComponentInfo) GetUptime() time.Duration {
	if ci.Uptime != nil {
		return *ci.Uptime
	}
	if ci.StartedAt != nil {
		return time.Since(*ci.StartedAt)
	}
	return 0
}

func (ci *ComponentInfo) GetErrorRate() float64 {
	if ci.Metrics.RequestCount == 0 {
		return 0
	}
	return float64(ci.Metrics.ErrorCount) / float64(ci.Metrics.RequestCount) * 100
}

func (ci *ComponentInfo) HasDependencies() bool {
	return len(ci.Dependencies) > 0
}

func (ci *ComponentInfo) HasEndpoints() bool {
	return len(ci.Endpoints) > 0
}

func (ci *ComponentInfo) GetHealthyEndpointsCount() int {
	count := 0
	for _, endpoint := range ci.Endpoints {
		if endpoint.IsHealthy {
			count++
		}
	}
	return count
}

// Helper methods for SystemStatus
func (ss *SystemStatus) IsHealthy() bool {
	return ss.Health == ComponentHealthHealthy
}

func (ss *SystemStatus) IsReady() bool {
	return ss.Status == ComponentStatusReady || ss.Status == ComponentStatusRunning
}

func (ss *SystemStatus) GetReadyPercentage() float64 {
	if ss.ComponentsTotal == 0 {
		return 0
	}
	return float64(ss.ComponentsReady) / float64(ss.ComponentsTotal) * 100
}

func (ss *SystemStatus) GetErrorPercentage() float64 {
	if ss.ComponentsTotal == 0 {
		return 0
	}
	return float64(ss.ComponentsError) / float64(ss.ComponentsTotal) * 100
}

func (ss *SystemStatus) HasCriticalErrors() bool {
	return ss.ComponentsError > 0
}

// Helper methods for ComponentMetrics
func (cm *ComponentMetrics) GetErrorRate() float64 {
	if cm.RequestCount == 0 {
		return 0
	}
	return float64(cm.ErrorCount) / float64(cm.RequestCount) * 100
}

func (cm *ComponentMetrics) IsPerformant() bool {
	// Consider component performant if error rate < 5% and response time < 1s
	return cm.GetErrorRate() < 5.0 && cm.AverageResponseTime < time.Second
}

func (cm *ComponentMetrics) HasRecentActivity() bool {
	if cm.LastRequestAt == nil {
		return false
	}
	return time.Since(*cm.LastRequestAt) < time.Minute*5
}

// Helper functions for component operations
func NewComponentInfo(name string, componentType ComponentType) *ComponentInfo {
	now := time.Now()
	return &ComponentInfo{
		Name:            name,
		Type:            componentType,
		Status:          ComponentStatusStarting,
		Health:          ComponentHealthUnknown,
		IsEnabled:       true,
		ReadyFlag:       false,
		RestartCount:    0,
		StartedAt:       &now,
		LastHealthCheck: &now,
		Metrics:         ComponentMetrics{},
		Configuration:   make(map[string]interface{}),
		Dependencies:    make([]string, 0),
		Endpoints:       make([]ComponentEndpoint, 0),
		Tags:            make([]string, 0),
	}
}

func NewComponentEndpoint(name, endpointType, address string, port int32) ComponentEndpoint {
	return ComponentEndpoint{
		Name:      name,
		Type:      endpointType,
		Address:   address,
		Port:      port,
		IsSecure:  false,
		IsHealthy: true,
		Metadata:  make(map[string]string),
	}
}
