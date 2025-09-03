/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"atom-engine/proto/messages/messagespb"
	"atom-engine/src/core/logger"
	"atom-engine/src/messages"
)

// messagesServiceServer implements messages gRPC service
type messagesServiceServer struct {
	messagespb.UnimplementedMessagesServiceServer
	core CoreInterface
}

// PublishMessage publishes a message
func (s *messagesServiceServer) PublishMessage(ctx context.Context, req *messagespb.PublishMessageRequest) (*messagespb.PublishMessageResponse, error) {
	logger.Info("PublishMessage gRPC request",
		logger.String("tenant_id", req.TenantId),
		logger.String("message_name", req.MessageName),
		logger.String("correlation_key", req.CorrelationKey))

	// Convert variables
	variables := make(map[string]interface{})
	for k, v := range req.Variables {
		variables[k] = v
	}

	// Create JSON message for messages component
	payload := messages.PublishMessagePayload{
		TenantID:       req.TenantId,
		MessageName:    req.MessageName,
		CorrelationKey: req.CorrelationKey,
		Variables:      variables,
		TTLSeconds:     int(req.TtlSeconds),
	}

	message, err := messages.CreatePublishMessageMessage(payload)
	if err != nil {
		logger.Error("Failed to create publish message", logger.String("error", err.Error()))
		return &messagespb.PublishMessageResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create publish message: %v", err),
		}, nil
	}

	// Send JSON message to messages component through Core
	if err := s.core.SendMessage("messages", message); err != nil {
		logger.Error("Failed to send publish message", logger.String("error", err.Error()))
		return &messagespb.PublishMessageResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	logger.Info("Message publish request sent successfully")

	// Wait for response from messages component
	// Ожидаем ответ от компонента messages
	responseJSON, err := s.core.WaitForMessagesResponse(5000) // 5 second timeout
	if err != nil {
		logger.Error("Failed to get messages response", logger.String("error", err.Error()))
		return &messagespb.PublishMessageResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get messages response: %v", err),
		}, nil
	}

	// Parse JSON response
	// Парсим JSON ответ
	var messagesResponse messages.MessageResponse
	if err := json.Unmarshal([]byte(responseJSON), &messagesResponse); err != nil {
		logger.Error("Failed to parse messages response", logger.String("error", err.Error()))
		return &messagespb.PublishMessageResponse{
			Success: false,
			Message: fmt.Sprintf("failed to parse response JSON: %v", err),
		}, nil
	}

	if !messagesResponse.Success {
		return &messagespb.PublishMessageResponse{
			Success: false,
			Message: messagesResponse.Error,
		}, nil
	}

	// Extract message ID from response
	// Извлекаем message ID из ответа
	messageID := "unknown"
	if resultData, ok := messagesResponse.Result.(map[string]interface{}); ok {
		if id, ok := resultData["message_id"].(string); ok {
			messageID = id
		}
	}

	return &messagespb.PublishMessageResponse{
		MessageId: messageID,
		Success:   true,
		Message:   "message published successfully",
	}, nil
}

// ListBufferedMessages lists buffered messages
func (s *messagesServiceServer) ListBufferedMessages(ctx context.Context, req *messagespb.ListBufferedMessagesRequest) (*messagespb.ListBufferedMessagesResponse, error) {
	logger.Info("ListBufferedMessages gRPC request",
		logger.String("tenant_id", req.TenantId),
		logger.Int("limit", int(req.Limit)),
		logger.Int("offset", int(req.Offset)))

	// Get messages component from core
	componentIf := s.core.GetMessagesComponent()
	if componentIf == nil {
		return &messagespb.ListBufferedMessagesResponse{
			Success: false,
			Message: "messages component not available",
		}, nil
	}

	// Cast to messages component
	messageComp, ok := componentIf.(*messages.Component)
	if !ok {
		return &messagespb.ListBufferedMessagesResponse{
			Success: false,
			Message: "messages component type assertion failed",
		}, nil
	}

	// Set defaults
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)

	// List buffered messages
	messages, err := messageComp.ListBufferedMessages(ctx, req.TenantId, limit, offset)
	if err != nil {
		logger.Error("Failed to list buffered messages", logger.String("error", err.Error()))
		return &messagespb.ListBufferedMessagesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Convert to protobuf format
	pbMessages := make([]*messagespb.BufferedMessage, len(messages))
	for i, msg := range messages {
		expiresAt := int64(0)
		if msg.ExpiresAt != nil {
			expiresAt = msg.ExpiresAt.Unix()
		}

		// Convert variables
		variables := make(map[string]string)
		for k, v := range msg.Variables {
			if str, ok := v.(string); ok {
				variables[k] = str
			} else {
				variables[k] = fmt.Sprintf("%v", v)
			}
		}

		pbMessages[i] = &messagespb.BufferedMessage{
			Id:             msg.ID,
			TenantId:       msg.TenantID,
			Name:           msg.Name,
			CorrelationKey: msg.CorrelationKey,
			Variables:      variables,
			PublishedAt:    msg.PublishedAt.Unix(),
			BufferedAt:     msg.BufferedAt.Unix(),
			ExpiresAt:      expiresAt,
			Reason:         msg.Reason,
		}
	}

	return &messagespb.ListBufferedMessagesResponse{
		Messages:   pbMessages,
		TotalCount: int32(len(pbMessages)),
		Success:    true,
		Message:    "buffered messages retrieved successfully",
	}, nil
}

// ListMessageSubscriptions lists message subscriptions
func (s *messagesServiceServer) ListMessageSubscriptions(ctx context.Context, req *messagespb.ListMessageSubscriptionsRequest) (*messagespb.ListMessageSubscriptionsResponse, error) {
	logger.Info("ListMessageSubscriptions gRPC request",
		logger.String("tenant_id", req.TenantId),
		logger.Int("limit", int(req.Limit)),
		logger.Int("offset", int(req.Offset)))

	// Get messages component from core
	componentIf := s.core.GetMessagesComponent()
	if componentIf == nil {
		return &messagespb.ListMessageSubscriptionsResponse{
			Success: false,
			Message: "messages component not available",
		}, nil
	}

	// Cast to messages component
	messageComp, ok := componentIf.(*messages.Component)
	if !ok {
		return &messagespb.ListMessageSubscriptionsResponse{
			Success: false,
			Message: "messages component type assertion failed",
		}, nil
	}

	// Set defaults
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)

	// List subscriptions
	subscriptions, err := messageComp.ListMessageSubscriptions(ctx, req.TenantId, limit, offset)
	if err != nil {
		logger.Error("Failed to list message subscriptions", logger.String("error", err.Error()))
		return &messagespb.ListMessageSubscriptionsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Convert to protobuf format
	pbSubscriptions := make([]*messagespb.MessageSubscription, len(subscriptions))
	for i, sub := range subscriptions {
		pbSubscriptions[i] = &messagespb.MessageSubscription{
			Id:                   sub.ID,
			TenantId:             sub.TenantID,
			ProcessDefinitionKey: sub.ProcessDefinitionKey,
			ProcessVersion:       sub.ProcessVersion,
			StartEventId:         sub.StartEventID,
			MessageName:          sub.MessageName,
			MessageRef:           sub.MessageRef,
			CorrelationKey:       sub.CorrelationKey,
			IsActive:             sub.IsActive,
			CreatedAt:            sub.CreatedAt.Unix(),
			UpdatedAt:            sub.UpdatedAt.Unix(),
		}
	}

	return &messagespb.ListMessageSubscriptionsResponse{
		Subscriptions: pbSubscriptions,
		TotalCount:    int32(len(pbSubscriptions)),
		Success:       true,
		Message:       "message subscriptions retrieved successfully",
	}, nil
}

// GetMessageStats gets message statistics
func (s *messagesServiceServer) GetMessageStats(ctx context.Context, req *messagespb.GetMessageStatsRequest) (*messagespb.GetMessageStatsResponse, error) {
	logger.Info("GetMessageStats gRPC request", logger.String("tenant_id", req.TenantId))

	// Get messages component from core
	componentIf := s.core.GetMessagesComponent()
	if componentIf == nil {
		return &messagespb.GetMessageStatsResponse{
			Success: false,
			Message: "messages component not available",
		}, nil
	}

	// Cast to messages component
	messageComp, ok := componentIf.(*messages.Component)
	if !ok {
		return &messagespb.GetMessageStatsResponse{
			Success: false,
			Message: "messages component type assertion failed",
		}, nil
	}

	// Get stats
	stats, err := messageComp.GetMessageStats(ctx, req.TenantId)
	if err != nil {
		logger.Error("Failed to get message stats", logger.String("error", err.Error()))
		return &messagespb.GetMessageStatsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &messagespb.GetMessageStatsResponse{
		Stats: &messagespb.MessageStats{
			TotalMessages:         int32(stats.TotalMessages),
			BufferedMessages:      int32(stats.BufferedMessages),
			ExpiredMessages:       int32(stats.ExpiredMessages),
			PublishedToday:        int32(stats.PublishedToday),
			InstancesCreatedToday: int32(stats.InstancesCreatedToday),
		},
		Success: true,
		Message: "message statistics retrieved successfully",
	}, nil
}

// CleanupExpiredMessages cleans up expired messages
func (s *messagesServiceServer) CleanupExpiredMessages(ctx context.Context, req *messagespb.CleanupExpiredMessagesRequest) (*messagespb.CleanupExpiredMessagesResponse, error) {
	logger.Info("CleanupExpiredMessages gRPC request", logger.String("tenant_id", req.TenantId))

	// Get messages component from core
	componentIf := s.core.GetMessagesComponent()
	if componentIf == nil {
		return &messagespb.CleanupExpiredMessagesResponse{
			Success: false,
			Message: "messages component not available",
		}, nil
	}

	// Cast to messages component
	messageComp, ok := componentIf.(*messages.Component)
	if !ok {
		return &messagespb.CleanupExpiredMessagesResponse{
			Success: false,
			Message: "messages component type assertion failed",
		}, nil
	}

	// Cleanup expired messages
	cleanedCount, err := messageComp.CleanupExpiredMessages(ctx)
	if err != nil {
		logger.Error("Failed to cleanup expired messages", logger.String("error", err.Error()))
		return &messagespb.CleanupExpiredMessagesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	logger.Info("Cleaned up expired messages", logger.Int("count", cleanedCount))

	return &messagespb.CleanupExpiredMessagesResponse{
		CleanedCount: int32(cleanedCount),
		Success:      true,
		Message:      "expired messages cleaned successfully",
	}, nil
}
