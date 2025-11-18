/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"time"
)

// ChannelConfig represents configuration for a message channel
type ChannelConfig struct {
	BufferSize int           // Channel buffer size
	Timeout    time.Duration // Timeout for channel operations
}

// DefaultChannelConfig returns default channel configuration
func DefaultChannelConfig() ChannelConfig {
	return ChannelConfig{
		BufferSize: 100,
		Timeout:    30 * time.Second,
	}
}

// MessageMultiplexerInterface defines the contract for message routing
type MessageMultiplexerInterface interface {
	// Start begins message routing
	Start() error

	// Stop gracefully shuts down the multiplexer
	Stop() error

	// IsRunning returns whether the multiplexer is active
	IsRunning() bool

	// GetAPIResponseChannel returns channel for API responses
	GetAPIResponseChannel() <-chan string

	// GetJobCallbackChannel returns channel for job callbacks
	GetJobCallbackChannel() <-chan string

	// GetBPMNErrorChannel returns channel for BPMN errors
	GetBPMNErrorChannel() <-chan string

	// GetMetrics returns current routing metrics
	GetMetrics() MultiplexerMetrics
}

// MultiplexerMetrics contains statistics about message routing
type MultiplexerMetrics struct {
	TotalMessages   uint64 `json:"total_messages"`
	APIResponses    uint64 `json:"api_responses"`
	JobCallbacks    uint64 `json:"job_callbacks"`
	BPMNErrors      uint64 `json:"bpmn_errors"`
	UnknownMessages uint64 `json:"unknown_messages"`
	DroppedMessages uint64 `json:"dropped_messages"`
	LastMessageTime int64  `json:"last_message_time"`
	RoutingErrors   uint64 `json:"routing_errors"`
}

// MessageRouter defines interface for message classification and routing
type MessageRouter interface {
	// Route classifies and routes a message to appropriate channel
	Route(message string) error

	// ClassifyMessage returns the type of a message
	ClassifyMessage(message string) MessageType

	// GetMetrics returns current routing metrics
	GetMetrics() MultiplexerMetrics
}

// ChannelManager manages multiple output channels
type ChannelManager interface {
	// GetChannel returns a channel for the specified message type
	GetChannel(messageType MessageType) chan<- string

	// CreateChannels initializes all channels with configuration
	CreateChannels(config ChannelConfig) error

	// CloseChannels gracefully closes all channels
	CloseChannels()

	// GetChannelStatus returns status for all channels
	GetChannelStatus() map[MessageType]ChannelStatus
}

// ChannelStatus represents the status of a message channel
type ChannelStatus struct {
	BufferSize    int    `json:"buffer_size"`
	CurrentLength int    `json:"current_length"`
	IsClosed      bool   `json:"is_closed"`
	MessageCount  uint64 `json:"message_count"`
	LastActivity  int64  `json:"last_activity"`
}

// ComponentChannelProvider defines interface for components that provide channels
type ComponentChannelProvider interface {
	// GetSourceChannel returns the source channel for messages
	GetSourceChannel() <-chan string

	// GetComponentName returns the name of the component
	GetComponentName() string
}
