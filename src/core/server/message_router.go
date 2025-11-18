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
	"sync/atomic"

	"atom-engine/src/core/logger"
)

// messageRouter implements MessageRouter interface
type messageRouter struct {
	classifier     *MessageClassifier
	channelManager ChannelManager
	logger         logger.ComponentLogger

	// Metrics
	totalMessages   uint64
	apiResponses    uint64
	jobCallbacks    uint64
	bpmnErrors      uint64
	unknownMessages uint64
	droppedMessages uint64
	routingErrors   uint64

	mutex sync.RWMutex
}

// NewMessageRouter creates a new message router
func NewMessageRouter(channelManager ChannelManager, logger logger.ComponentLogger) MessageRouter {
	return &messageRouter{
		classifier:     NewMessageClassifier(),
		channelManager: channelManager,
		logger:         logger,
	}
}

// Route classifies and routes a message to appropriate channel
func (mr *messageRouter) Route(message string) error {
	if len(message) == 0 {
		atomic.AddUint64(&mr.routingErrors, 1)
		mr.logger.Warn("Attempted to route empty message")
		return fmt.Errorf("cannot route empty message")
	}

	// Increment total message counter
	atomic.AddUint64(&mr.totalMessages, 1)

	// Classify the message
	messageType := mr.ClassifyMessage(message)

	// Update metrics based on classification
	mr.updateMetrics(messageType)

	// Route to appropriate channel
	if err := mr.sendToChannel(messageType, message); err != nil {
		atomic.AddUint64(&mr.droppedMessages, 1)
		mr.logger.Error("Failed to route message",
			logger.String("message_type", messageType.String()),
			logger.String("error", err.Error()),
			logger.Int("message_length", len(message)))
		return fmt.Errorf("failed to route %s message: %w", messageType.String(), err)
	}

	mr.logger.Debug("Message routed successfully",
		logger.String("message_type", messageType.String()),
		logger.Int("message_length", len(message)))

	return nil
}

// ClassifyMessage returns the type of a message
func (mr *messageRouter) ClassifyMessage(message string) MessageType {
	return mr.classifier.ClassifyMessage(message)
}

// sendToChannel sends a message to the appropriate channel based on its type
func (mr *messageRouter) sendToChannel(messageType MessageType, message string) error {
	// Cast ChannelManager to access SendMessage method
	if cm, ok := mr.channelManager.(*channelManager); ok {
		return cm.SendMessage(messageType, message)
	}

	// Fallback: try to get channel directly and send
	ch := mr.channelManager.GetChannel(messageType)
	if ch == nil {
		return fmt.Errorf("no channel available for message type: %s", messageType.String())
	}

	select {
	case ch <- message:
		return nil
	default:
		return fmt.Errorf("channel %s is full or closed", messageType.String())
	}
}

// updateMetrics updates routing metrics based on message type
func (mr *messageRouter) updateMetrics(messageType MessageType) {
	switch messageType {
	case MessageTypeAPIResponse:
		atomic.AddUint64(&mr.apiResponses, 1)
	case MessageTypeJobCallback:
		atomic.AddUint64(&mr.jobCallbacks, 1)
	case MessageTypeBPMNError:
		atomic.AddUint64(&mr.bpmnErrors, 1)
	case MessageTypeUnknown:
		atomic.AddUint64(&mr.unknownMessages, 1)
		mr.logger.Warn("Unknown message type encountered")
	}
}

// GetMetrics returns current routing metrics
func (mr *messageRouter) GetMetrics() MultiplexerMetrics {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	return MultiplexerMetrics{
		TotalMessages:   atomic.LoadUint64(&mr.totalMessages),
		APIResponses:    atomic.LoadUint64(&mr.apiResponses),
		JobCallbacks:    atomic.LoadUint64(&mr.jobCallbacks),
		BPMNErrors:      atomic.LoadUint64(&mr.bpmnErrors),
		UnknownMessages: atomic.LoadUint64(&mr.unknownMessages),
		DroppedMessages: atomic.LoadUint64(&mr.droppedMessages),
		RoutingErrors:   atomic.LoadUint64(&mr.routingErrors),
		LastMessageTime: 0, // Will be updated by multiplexer
	}
}

// ResetMetrics resets all routing metrics to zero
func (mr *messageRouter) ResetMetrics() {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	atomic.StoreUint64(&mr.totalMessages, 0)
	atomic.StoreUint64(&mr.apiResponses, 0)
	atomic.StoreUint64(&mr.jobCallbacks, 0)
	atomic.StoreUint64(&mr.bpmnErrors, 0)
	atomic.StoreUint64(&mr.unknownMessages, 0)
	atomic.StoreUint64(&mr.droppedMessages, 0)
	atomic.StoreUint64(&mr.routingErrors, 0)

	mr.logger.Info("Routing metrics reset")
}

// GetMessageTypeStats returns detailed statistics for each message type
func (mr *messageRouter) GetMessageTypeStats() map[MessageType]uint64 {
	return map[MessageType]uint64{
		MessageTypeAPIResponse: atomic.LoadUint64(&mr.apiResponses),
		MessageTypeJobCallback: atomic.LoadUint64(&mr.jobCallbacks),
		MessageTypeBPMNError:   atomic.LoadUint64(&mr.bpmnErrors),
		MessageTypeUnknown:     atomic.LoadUint64(&mr.unknownMessages),
	}
}
