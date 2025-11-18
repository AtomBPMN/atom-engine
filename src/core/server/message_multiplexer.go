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

// messageMultiplexer implements MessageMultiplexerInterface
type messageMultiplexer struct {
	// Core components
	channelManager ChannelManager
	router         MessageRouter
	logger         logger.ComponentLogger

	// Input source
	sourceChannel <-chan string
	componentName string

	// State management
	isRunning bool
	stopChan  chan struct{}
	doneChan  chan struct{}
	mutex     sync.RWMutex

	// Metrics
	startTime   time.Time
	lastMessage time.Time
	config      ChannelConfig
}

// NewMessageMultiplexer creates a new message multiplexer
func NewMessageMultiplexer(
	sourceChannel <-chan string,
	componentName string,
	logger logger.ComponentLogger,
) MessageMultiplexerInterface {

	// Create channel manager
	channelManager := NewChannelManager(logger)

	// Create message router
	router := NewMessageRouter(channelManager, logger)

	return &messageMultiplexer{
		channelManager: channelManager,
		router:         router,
		logger:         logger,
		sourceChannel:  sourceChannel,
		componentName:  componentName,
		stopChan:       make(chan struct{}),
		doneChan:       make(chan struct{}),
		config:         DefaultChannelConfig(),
	}
}

// Start begins message routing
func (mm *messageMultiplexer) Start() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.isRunning {
		return fmt.Errorf("message multiplexer is already running")
	}

	// Initialize channels
	if err := mm.channelManager.CreateChannels(mm.config); err != nil {
		return fmt.Errorf("failed to create channels: %w", err)
	}

	// Set running state
	mm.isRunning = true
	mm.startTime = time.Now()

	// Start routing goroutine
	go mm.routingLoop()

	mm.logger.Info("Message multiplexer started",
		logger.String("component", mm.componentName),
		logger.Int("buffer_size", mm.config.BufferSize),
		logger.String("timeout", mm.config.Timeout.String()))

	return nil
}

// Stop gracefully shuts down the multiplexer
func (mm *messageMultiplexer) Stop() error {
	mm.mutex.Lock()
	if !mm.isRunning {
		mm.mutex.Unlock()
		return fmt.Errorf("message multiplexer is not running")
	}
	mm.mutex.Unlock()

	mm.logger.Info("Stopping message multiplexer",
		logger.String("component", mm.componentName))

	// Signal stop
	close(mm.stopChan)

	// Wait for routing loop to finish
	select {
	case <-mm.doneChan:
		mm.logger.Debug("Routing loop stopped")
	case <-time.After(5 * time.Second):
		mm.logger.Warn("Timeout waiting for routing loop to stop")
	}

	// Close all channels
	mm.channelManager.CloseChannels()

	mm.mutex.Lock()
	mm.isRunning = false
	mm.mutex.Unlock()

	mm.logger.Info("Message multiplexer stopped")
	return nil
}

// IsRunning returns whether the multiplexer is active
func (mm *messageMultiplexer) IsRunning() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.isRunning
}

// GetAPIResponseChannel returns channel for API responses
func (mm *messageMultiplexer) GetAPIResponseChannel() <-chan string {
	if cm, ok := mm.channelManager.(*channelManager); ok {
		return cm.GetAPIResponseChannel()
	}
	return nil
}

// GetJobCallbackChannel returns channel for job callbacks
func (mm *messageMultiplexer) GetJobCallbackChannel() <-chan string {
	if cm, ok := mm.channelManager.(*channelManager); ok {
		return cm.GetJobCallbackChannel()
	}
	return nil
}

// GetBPMNErrorChannel returns channel for BPMN errors
func (mm *messageMultiplexer) GetBPMNErrorChannel() <-chan string {
	if cm, ok := mm.channelManager.(*channelManager); ok {
		return cm.GetBPMNErrorChannel()
	}
	return nil
}

// GetMetrics returns current routing metrics
func (mm *messageMultiplexer) GetMetrics() MultiplexerMetrics {
	// Get router metrics
	metrics := mm.router.GetMetrics()

	// Add timing information
	mm.mutex.RLock()
	metrics.LastMessageTime = mm.lastMessage.Unix()
	mm.mutex.RUnlock()

	return metrics
}

// routingLoop is the main message processing loop
func (mm *messageMultiplexer) routingLoop() {
	defer close(mm.doneChan)

	mm.logger.Debug("Message routing loop started",
		logger.String("component", mm.componentName))

	for {
		select {
		case message, ok := <-mm.sourceChannel:
			if !ok {
				mm.logger.Info("Source channel closed, stopping routing loop")
				return
			}

			// Update last message time
			mm.mutex.Lock()
			mm.lastMessage = time.Now()
			mm.mutex.Unlock()

			// Route the message
			if err := mm.router.Route(message); err != nil {
				mm.logger.Error("Failed to route message",
					logger.String("error", err.Error()),
					logger.Int("message_length", len(message)),
					logger.String("component", mm.componentName))
			}

		case <-mm.stopChan:
			mm.logger.Debug("Stop signal received, shutting down routing loop")
			return
		}
	}
}

// SetChannelConfig updates channel configuration (only when stopped)
func (mm *messageMultiplexer) SetChannelConfig(config ChannelConfig) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.isRunning {
		return fmt.Errorf("cannot change configuration while multiplexer is running")
	}

	mm.config = config
	mm.logger.Info("Channel configuration updated",
		logger.Int("buffer_size", config.BufferSize),
		logger.String("timeout", config.Timeout.String()))

	return nil
}

// GetChannelStatus returns detailed status for all channels
func (mm *messageMultiplexer) GetChannelStatus() map[MessageType]ChannelStatus {
	return mm.channelManager.GetChannelStatus()
}

// GetComponentName returns the name of the source component
func (mm *messageMultiplexer) GetComponentName() string {
	return mm.componentName
}

// GetUptime returns how long the multiplexer has been running
func (mm *messageMultiplexer) GetUptime() time.Duration {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	if !mm.isRunning {
		return 0
	}

	return time.Since(mm.startTime)
}
