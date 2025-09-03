/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"atom-engine/proto/messages/messagespb"
	"atom-engine/src/core/logger"
)

// MessagePublish publishes a message via gRPC
// Публикует сообщение через gRPC
func (d *DaemonCommand) MessagePublish() error {
	logger.Debug("Publishing message")

	if len(os.Args) < 4 {
		logger.Error("Invalid message publish arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd message publish <name> [correlation_key] [variables] [ttl]")
	}

	// Parse arguments
	name := os.Args[3]
	var correlationKey string
	var variables string
	var ttlSeconds int64

	if len(os.Args) > 4 {
		correlationKey = os.Args[4]
	}
	if len(os.Args) > 5 {
		variables = os.Args[5]
	}
	if len(os.Args) > 6 {
		if ttl, err := fmt.Sscanf(os.Args[6], "%d", &ttlSeconds); err == nil && ttl == 1 {
			// TTL parsed successfully
		}
	}

	logger.Debug("Message publish request",
		logger.String("name", name),
		logger.String("correlation_key", correlationKey),
		logger.String("variables", variables))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for message publish", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create messages gRPC client
	client := messagespb.NewMessagesServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse variables if provided
	variablesMap := make(map[string]string)
	if variables != "" {
		// For now, simple implementation - store as single variable
		// TODO: Implement proper JSON parsing
		// Пока что простая реализация - сохраняем как одну переменную
		// ТОДО: Реализовать правильный парсинг JSON
		variablesMap["data"] = variables
	}

	// Make gRPC request
	request := &messagespb.PublishMessageRequest{
		TenantId:       "", // Default tenant
		MessageName:    name,
		CorrelationKey: correlationKey,
		Variables:      variablesMap,
		TtlSeconds:     ttlSeconds,
	}

	response, err := client.PublishMessage(ctx, request)
	if err != nil {
		logger.Error("Message publish failed", logger.String("error", err.Error()))
		return fmt.Errorf("message publish failed: %w", err)
	}

	fmt.Printf("Message Publish\n")
	fmt.Printf("===============\n")
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Correlation Key: %s\n", correlationKey)
	fmt.Printf("Message ID: %s\n", response.MessageId)
	fmt.Printf("Success: %t\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)

	if response.Success {
		logger.Info("Message published successfully", logger.String("message_id", response.MessageId))
	}

	return nil
}

// MessageList lists correlation results via gRPC
// Выводит список результатов корреляции через gRPC
func (d *DaemonCommand) MessageList() error {
	logger.Debug("Listing messages")

	// Parse arguments
	var tenantID string
	var limit int32 = 50
	if len(os.Args) > 3 {
		tenantID = os.Args[3]
	}

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for message list",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := messagespb.NewMessagesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListBufferedMessages(ctx, &messagespb.ListBufferedMessagesRequest{
		TenantId: tenantID,
		Limit:    limit,
	})
	if err != nil {
		logger.Error("Failed to list buffered messages", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list buffered messages: %w", err)
	}

	logger.Debug("Messages listed", logger.Int("count", len(resp.Messages)))

	fmt.Printf("Buffered Message List\n")
	fmt.Printf("====================\n")
	printMessagesTable(resp.Messages, resp.TotalCount)

	return nil
}

// MessageSubscriptions lists message subscriptions via gRPC
// Выводит список подписок на сообщения через gRPC
func (d *DaemonCommand) MessageSubscriptions() error {
	logger.Debug("Listing message subscriptions")

	// Parse arguments
	var tenantID string
	var limit int32 = 50
	if len(os.Args) > 3 {
		tenantID = os.Args[3]
	}

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for message subscriptions",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := messagespb.NewMessagesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListMessageSubscriptions(ctx, &messagespb.ListMessageSubscriptionsRequest{
		TenantId: tenantID,
		Limit:    limit,
	})
	if err != nil {
		logger.Error("Failed to list message subscriptions", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list message subscriptions: %w", err)
	}

	logger.Debug("Message subscriptions listed", logger.Int("count", len(resp.Subscriptions)))

	fmt.Printf("Message Subscriptions\n")
	fmt.Printf("====================\n")
	printMessageSubscriptionsTable(resp.Subscriptions, resp.TotalCount)

	return nil
}

// MessageBuffered lists buffered messages via gRPC
// Выводит список буферизованных сообщений через gRPC
func (d *DaemonCommand) MessageBuffered() error {
	logger.Debug("Listing buffered messages")

	fmt.Printf("Buffered Messages\n")
	fmt.Printf("=================\n")
	fmt.Printf("Note: Buffered messages functionality needs to be implemented\n")

	return nil
}

// MessageCleanup cleans up expired messages via gRPC
// Очищает просроченные сообщения через gRPC
func (d *DaemonCommand) MessageCleanup() error {
	logger.Debug("Cleaning up expired messages")

	fmt.Printf("Message Cleanup\n")
	fmt.Printf("===============\n")
	fmt.Printf("Note: Message cleanup functionality needs to be implemented\n")

	return nil
}

// MessageStats shows message statistics via gRPC
// Показывает статистику сообщений через gRPC
func (d *DaemonCommand) MessageStats() error {
	logger.Debug("Getting message statistics")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for message stats",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := messagespb.NewMessagesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetMessageStats(ctx, &messagespb.GetMessageStatsRequest{})
	if err != nil {
		logger.Error("Failed to get message stats", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get message stats: %w", err)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		return nil
	}

	stats := resp.Stats
	logger.Debug("Message stats retrieved",
		logger.Int("total_messages", int(stats.TotalMessages)),
		logger.Int("buffered_messages", int(stats.BufferedMessages)))

	fmt.Printf("Message Statistics\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
	fmt.Printf("Buffered Messages: %d\n", stats.BufferedMessages)
	fmt.Printf("Expired Messages: %d\n", stats.ExpiredMessages)
	fmt.Printf("Published Today: %d\n", stats.PublishedToday)
	fmt.Printf("Instances Created Today: %d\n", stats.InstancesCreatedToday)

	return nil
}

// MessageTest tests message functionality via gRPC
// Тестирует функциональность сообщений через gRPC
func (d *DaemonCommand) MessageTest() error {
	logger.Debug("Testing message functionality")

	fmt.Printf("Message Test\n")
	fmt.Printf("============\n")
	fmt.Printf("Note: Message test functionality needs to be implemented\n")

	return nil
}
