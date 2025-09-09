/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"
	"sync"
	"time"

	"atom-engine/src/core/logger"
)

// channelManager implements ChannelManager interface
type channelManager struct {
	apiResponseChan chan string
	jobCallbackChan chan string
	bpmnErrorChan   chan string

	// Metrics and status tracking
	channels      map[MessageType]chan string
	channelStatus map[MessageType]*channelStats
	mutex         sync.RWMutex
	logger        logger.ComponentLogger
	config        ChannelConfig
}

// channelStats tracks statistics for a single channel
type channelStats struct {
	messageCount uint64
	lastActivity int64
	mutex        sync.RWMutex
}

// NewChannelManager creates a new channel manager
func NewChannelManager(logger logger.ComponentLogger) ChannelManager {
	return &channelManager{
		channels:      make(map[MessageType]chan string),
		channelStatus: make(map[MessageType]*channelStats),
		logger:        logger,
		config:        DefaultChannelConfig(),
	}
}

// CreateChannels initializes all channels with configuration
func (cm *channelManager) CreateChannels(config ChannelConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.config = config

	// Create API response channel
	cm.apiResponseChan = make(chan string, config.BufferSize)
	cm.channels[MessageTypeAPIResponse] = cm.apiResponseChan
	cm.channelStatus[MessageTypeAPIResponse] = &channelStats{}

	// Create job callback channel
	cm.jobCallbackChan = make(chan string, config.BufferSize)
	cm.channels[MessageTypeJobCallback] = cm.jobCallbackChan
	cm.channelStatus[MessageTypeJobCallback] = &channelStats{}

	// Create BPMN error channel
	cm.bpmnErrorChan = make(chan string, config.BufferSize)
	cm.channels[MessageTypeBPMNError] = cm.bpmnErrorChan
	cm.channelStatus[MessageTypeBPMNError] = &channelStats{}

	cm.logger.Info("Channels created successfully",
		logger.Int("buffer_size", config.BufferSize),
		logger.String("timeout", config.Timeout.String()),
		logger.Int("channel_count", len(cm.channels)))

	return nil
}

// GetChannel returns a channel for the specified message type
func (cm *channelManager) GetChannel(messageType MessageType) chan<- string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if ch, exists := cm.channels[messageType]; exists {
		return ch
	}

	cm.logger.Warn("Requested channel for unknown message type",
		logger.String("message_type", messageType.String()))

	return nil
}

// CloseChannels gracefully closes all channels
func (cm *channelManager) CloseChannels() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Close all channels
	for msgType, ch := range cm.channels {
		if ch != nil {
			close(ch)
			cm.logger.Debug("Channel closed",
				logger.String("message_type", msgType.String()))
		}
	}

	// Clear the maps
	cm.channels = make(map[MessageType]chan string)
	cm.channelStatus = make(map[MessageType]*channelStats)

	cm.logger.Info("All channels closed")
}

// GetChannelStatus returns status for all channels
func (cm *channelManager) GetChannelStatus() map[MessageType]ChannelStatus {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	status := make(map[MessageType]ChannelStatus)

	for msgType, ch := range cm.channels {
		stats := cm.channelStatus[msgType]
		if stats == nil {
			continue
		}

		stats.mutex.RLock()
		channelStatus := ChannelStatus{
			BufferSize:    cm.config.BufferSize,
			CurrentLength: len(ch),
			IsClosed:      cm.isChannelClosed(ch),
			MessageCount:  stats.messageCount,
			LastActivity:  stats.lastActivity,
		}
		stats.mutex.RUnlock()

		status[msgType] = channelStatus
	}

	return status
}

// SendMessage attempts to send a message to the specified channel type
func (cm *channelManager) SendMessage(messageType MessageType, message string) error {
	ch := cm.GetChannel(messageType)
	if ch == nil {
		return fmt.Errorf("no channel available for message type: %s", messageType.String())
	}

	// Update statistics
	cm.updateChannelStats(messageType)

	// Try to send with timeout
	select {
	case ch <- message:
		cm.logger.Debug("Message sent to channel",
			logger.String("message_type", messageType.String()),
			logger.Int("message_length", len(message)))
		return nil

	case <-time.After(cm.config.Timeout):
		cm.logger.Warn("Channel send timeout",
			logger.String("message_type", messageType.String()),
			logger.String("timeout", cm.config.Timeout.String()))
		return fmt.Errorf("timeout sending message to %s channel", messageType.String())

	default:
		// Channel is full
		cm.logger.Warn("Channel is full, message dropped",
			logger.String("message_type", messageType.String()),
			logger.Int("buffer_size", cm.config.BufferSize))
		return fmt.Errorf("channel %s is full", messageType.String())
	}
}

// GetAPIResponseChannel returns the API response channel for reading
func (cm *channelManager) GetAPIResponseChannel() <-chan string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.apiResponseChan
}

// GetJobCallbackChannel returns the job callback channel for reading
func (cm *channelManager) GetJobCallbackChannel() <-chan string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.jobCallbackChan
}

// GetBPMNErrorChannel returns the BPMN error channel for reading
func (cm *channelManager) GetBPMNErrorChannel() <-chan string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.bpmnErrorChan
}

// updateChannelStats updates statistics for a channel
func (cm *channelManager) updateChannelStats(messageType MessageType) {
	cm.mutex.RLock()
	stats := cm.channelStatus[messageType]
	cm.mutex.RUnlock()

	if stats != nil {
		stats.mutex.Lock()
		stats.messageCount++
		stats.lastActivity = time.Now().Unix()
		stats.mutex.Unlock()
	}
}

// isChannelClosed checks if a channel is closed
func (cm *channelManager) isChannelClosed(ch chan string) bool {
	select {
	case _, ok := <-ch:
		if !ok {
			return true // Channel is closed
		}
		// Put the value back (channel state detection method)
		// In practice, this method should be used carefully
		return false
	default:
		return false // Channel is open and has no data
	}
}
