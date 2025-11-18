package types

import (
	"time"
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	MessageStatusPending    MessageStatus = "PENDING"
	MessageStatusCorrelated MessageStatus = "CORRELATED"
	MessageStatusExpired    MessageStatus = "EXPIRED"
	MessageStatusBuffered   MessageStatus = "BUFFERED"
)

// MessageType represents the type of a message
type MessageType string

const (
	MessageTypeStart        MessageType = "START"
	MessageTypeIntermediate MessageType = "INTERMEDIATE"
	MessageTypeBoundary     MessageType = "BOUNDARY"
	MessageTypeEnd          MessageType = "END"
)

// MessageVariables represents variables associated with a message
type MessageVariables map[string]interface{}

// MessageInfo represents detailed information about a message
type MessageInfo struct {
	MessageID         string           `json:"message_id"`
	Name              string           `json:"name"`
	CorrelationKey    string           `json:"correlation_key,omitempty"`
	TenantID          string           `json:"tenant_id"`
	Variables         MessageVariables `json:"variables,omitempty"`
	TTL               *time.Duration   `json:"ttl,omitempty"`
	Status            MessageStatus    `json:"status"`
	Type              MessageType      `json:"type,omitempty"`
	ElementID         string           `json:"element_id,omitempty"`
	ProcessInstanceID string           `json:"process_instance_id,omitempty"`
	PublishedAt       time.Time        `json:"published_at"`
	ExpiresAt         *time.Time       `json:"expires_at,omitempty"`
	CorrelatedAt      *time.Time       `json:"correlated_at,omitempty"`
	Reason            string           `json:"reason,omitempty"`
}

// BufferedMessageInfo represents a buffered message
type BufferedMessageInfo struct {
	MessageID      string           `json:"message_id"`
	Name           string           `json:"name"`
	CorrelationKey string           `json:"correlation_key,omitempty"`
	TenantID       string           `json:"tenant_id"`
	Variables      MessageVariables `json:"variables,omitempty"`
	ElementID      string           `json:"element_id,omitempty"`
	PublishedAt    time.Time        `json:"published_at"`
	ExpiresAt      time.Time        `json:"expires_at"`
	Reason         string           `json:"reason"`
	BufferedAt     time.Time        `json:"buffered_at"`
}

// MessageSubscriptionInfo represents a message subscription
type MessageSubscriptionInfo struct {
	SubscriptionID       string     `json:"subscription_id"`
	ProcessDefinitionKey string     `json:"process_definition_key"`
	Version              int32      `json:"version"`
	MessageName          string     `json:"message_name"`
	StartEventID         string     `json:"start_event_id,omitempty"`
	CorrelationKey       string     `json:"correlation_key,omitempty"`
	TenantID             string     `json:"tenant_id"`
	IsActive             bool       `json:"is_active"`
	CreatedAt            time.Time  `json:"created_at"`
	LastUsedAt           *time.Time `json:"last_used_at,omitempty"`
}

// MessageCorrelationInfo represents the result of message correlation
type MessageCorrelationInfo struct {
	MessageID           string                `json:"message_id"`
	CorrelationKey      string                `json:"correlation_key,omitempty"`
	CorrelatedInstances []ProcessInstanceInfo `json:"correlated_instances"`
	StartedInstances    []ProcessInstanceInfo `json:"started_instances"`
	CorrelationCount    int32                 `json:"correlation_count"`
	Success             bool                  `json:"success"`
	ErrorMessage        string                `json:"error_message,omitempty"`
	CorrelatedAt        time.Time             `json:"correlated_at"`
}

// ProcessInstanceInfo represents basic process instance information
type ProcessInstanceInfo struct {
	InstanceID string    `json:"instance_id"`
	ProcessKey string    `json:"process_key"`
	Version    int32     `json:"version"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
}

// MessageStats represents statistics about messages in the system
type MessageStats struct {
	TotalMessages          int64            `json:"total_messages"`
	PendingMessages        int64            `json:"pending_messages"`
	CorrelatedMessages     int64            `json:"correlated_messages"`
	ExpiredMessages        int64            `json:"expired_messages"`
	BufferedMessages       int64            `json:"buffered_messages"`
	TotalSubscriptions     int64            `json:"total_subscriptions"`
	ActiveSubscriptions    int64            `json:"active_subscriptions"`
	MessagesByName         map[string]int64 `json:"messages_by_name"`
	MessagesByTenant       map[string]int64 `json:"messages_by_tenant"`
	CorrelationRate        float64          `json:"correlation_rate"`
	AverageCorrelationTime time.Duration    `json:"average_correlation_time"`
	LastMessagePublished   *time.Time       `json:"last_message_published,omitempty"`
	LastCorrelation        *time.Time       `json:"last_correlation,omitempty"`
}

// MessagePublishRequest represents a request to publish a message
type MessagePublishRequest struct {
	Name           string           `json:"name" validate:"required"`
	CorrelationKey string           `json:"correlation_key,omitempty"`
	TenantID       string           `json:"tenant_id,omitempty"`
	Variables      MessageVariables `json:"variables,omitempty"`
	TTL            *time.Duration   `json:"ttl,omitempty"`
	ElementID      string           `json:"element_id,omitempty"`
}

// MessagePublishResponse represents the response from publishing a message
type MessagePublishResponse struct {
	MessageID   string                  `json:"message_id"`
	Success     bool                    `json:"success"`
	Message     string                  `json:"message"`
	Correlation *MessageCorrelationInfo `json:"correlation,omitempty"`
	PublishedAt time.Time               `json:"published_at"`
}

// MessageListRequest represents a request to list messages
type MessageListRequest struct {
	TenantID       *string        `json:"tenant_id,omitempty"`
	Name           *string        `json:"name,omitempty"`
	Status         *MessageStatus `json:"status,omitempty"`
	CorrelationKey *string        `json:"correlation_key,omitempty"`
	Limit          int32          `json:"limit,omitempty"`
	Offset         int32          `json:"offset,omitempty"`
	IncludeExpired bool           `json:"include_expired,omitempty"`
}

// MessageListResponse represents a response with a list of messages
type MessageListResponse struct {
	Messages   []MessageInfo `json:"messages"`
	TotalCount int32         `json:"total_count"`
	HasMore    bool          `json:"has_more"`
}

// BufferedMessageListRequest represents a request to list buffered messages
type BufferedMessageListRequest struct {
	TenantID *string `json:"tenant_id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Limit    int32   `json:"limit,omitempty"`
	Offset   int32   `json:"offset,omitempty"`
}

// BufferedMessageListResponse represents a response with buffered messages
type BufferedMessageListResponse struct {
	Messages   []BufferedMessageInfo `json:"messages"`
	TotalCount int32                 `json:"total_count"`
	HasMore    bool                  `json:"has_more"`
}

// MessageSubscriptionListRequest represents a request to list subscriptions
type MessageSubscriptionListRequest struct {
	TenantID    *string `json:"tenant_id,omitempty"`
	MessageName *string `json:"message_name,omitempty"`
	ProcessKey  *string `json:"process_key,omitempty"`
	ActiveOnly  bool    `json:"active_only,omitempty"`
	Limit       int32   `json:"limit,omitempty"`
	Offset      int32   `json:"offset,omitempty"`
}

// MessageSubscriptionListResponse represents a response with subscriptions
type MessageSubscriptionListResponse struct {
	Subscriptions []MessageSubscriptionInfo `json:"subscriptions"`
	TotalCount    int32                     `json:"total_count"`
	HasMore       bool                      `json:"has_more"`
}

// MessageCleanupRequest represents a request to cleanup expired messages
type MessageCleanupRequest struct {
	TenantID    *string        `json:"tenant_id,omitempty"`
	MaxAge      *time.Duration `json:"max_age,omitempty"`
	MessageName *string        `json:"message_name,omitempty"`
	DryRun      bool           `json:"dry_run,omitempty"`
}

// MessageCleanupResponse represents the response from cleanup operation
type MessageCleanupResponse struct {
	CleanedCount   int32         `json:"cleaned_count"`
	DryRun         bool          `json:"dry_run"`
	Success        bool          `json:"success"`
	ErrorMessage   string        `json:"error_message,omitempty"`
	CleanedAt      time.Time     `json:"cleaned_at"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// MessageTestRequest represents a request to test message functionality
type MessageTestRequest struct {
	TestType    string `json:"test_type" validate:"required"`
	MessageName string `json:"message_name,omitempty"`
	TenantID    string `json:"tenant_id,omitempty"`
}

// MessageTestResponse represents the response from message testing
type MessageTestResponse struct {
	TestType     string                 `json:"test_type"`
	Success      bool                   `json:"success"`
	Results      map[string]interface{} `json:"results"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	TestedAt     time.Time              `json:"tested_at"`
}

// Helper methods for MessageVariables
func (mv MessageVariables) GetString(key string) (string, bool) {
	if val, exists := mv[key]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (mv MessageVariables) GetInt64(key string) (int64, bool) {
	if val, exists := mv[key]; exists {
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

func (mv MessageVariables) GetBool(key string) (bool, bool) {
	if val, exists := mv[key]; exists {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// Helper methods for MessageInfo
func (mi *MessageInfo) IsPending() bool {
	return mi.Status == MessageStatusPending
}

func (mi *MessageInfo) IsCorrelated() bool {
	return mi.Status == MessageStatusCorrelated
}

func (mi *MessageInfo) IsExpired() bool {
	return mi.Status == MessageStatusExpired
}

func (mi *MessageInfo) IsBuffered() bool {
	return mi.Status == MessageStatusBuffered
}

func (mi *MessageInfo) HasTTL() bool {
	return mi.TTL != nil
}

func (mi *MessageInfo) IsExpiredNow() bool {
	return mi.ExpiresAt != nil && time.Now().After(*mi.ExpiresAt)
}

// Helper methods for MessageStats
func (ms *MessageStats) GetCorrelationRate() float64 {
	if ms.TotalMessages == 0 {
		return 0
	}
	return float64(ms.CorrelatedMessages) / float64(ms.TotalMessages) * 100
}

func (ms *MessageStats) GetBufferingRate() float64 {
	if ms.TotalMessages == 0 {
		return 0
	}
	return float64(ms.BufferedMessages) / float64(ms.TotalMessages) * 100
}

func (ms *MessageStats) GetSubscriptionEfficiency() float64 {
	if ms.TotalSubscriptions == 0 {
		return 0
	}
	return float64(ms.ActiveSubscriptions) / float64(ms.TotalSubscriptions) * 100
}

// Helper methods for MessageCorrelationInfo
func (mci *MessageCorrelationInfo) HasCorrelations() bool {
	return mci.CorrelationCount > 0
}

func (mci *MessageCorrelationInfo) GetTotalAffectedInstances() int {
	return len(mci.CorrelatedInstances) + len(mci.StartedInstances)
}

func (mci *MessageCorrelationInfo) IsSuccessful() bool {
	return mci.Success && mci.CorrelationCount > 0
}
