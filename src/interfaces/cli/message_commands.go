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
	"strings"
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
		// Simple implementation - store as single variable
		// Простая реализация - сохраняем как одну переменную
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

	// Parse arguments for filtering and pagination
	var tenantID string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd message list"

	// Parse arguments: handle flags and positional arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--page" || arg == "-p" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &page); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if arg == "--page-size" || arg == "-s" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &pageSize); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			// Positional arguments
			if tenantID == "" {
				tenantID = arg
			}
		}
	}

	logger.Debug("Message list request",
		logger.String("tenant_id", tenantID),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

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
		TenantId:  tenantID,
		Limit:     0, // Use pagination instead
		PageSize:  pageSize,
		Page:      page,
		SortBy:    "published_at",
		SortOrder: "DESC",
	})
	if err != nil {
		logger.Error("Failed to list buffered messages", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list buffered messages: %w", err)
	}

	logger.Debug("Messages listed", logger.Int("count", len(resp.Messages)))

	fmt.Printf("Buffered Message List\n")
	fmt.Printf("====================\n")

	// Print pagination info if multiple pages exist
	if resp.TotalPages > 1 {
		fmt.Printf("Page %d of %d (Total: %d messages, Showing: %d)\n\n",
			resp.Page, resp.TotalPages, resp.TotalCount, len(resp.Messages))
	} else {
		fmt.Printf("Found %d message(s):\n\n", resp.TotalCount)
	}

	printMessagesTable(resp.Messages, resp.TotalCount)

	// Show navigation hints for pagination
	if resp.TotalPages > 1 {
		fmt.Printf("\nNavigation:\n")

		// Previous page
		if resp.Page > 1 {
			prevPageCmd := fmt.Sprintf("atomd message list")
			if tenantID != "" {
				prevPageCmd += fmt.Sprintf(" %s", tenantID)
			}
			prevPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page-1, resp.PageSize)
			fmt.Printf("Previous page: %s\n", prevPageCmd)
		}

		// Next page
		if resp.Page < resp.TotalPages {
			nextPageCmd := fmt.Sprintf("atomd message list")
			if tenantID != "" {
				nextPageCmd += fmt.Sprintf(" %s", tenantID)
			}
			nextPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page+1, resp.PageSize)
			fmt.Printf("Next page: %s\n", nextPageCmd)
		}
	}

	return nil
}

// MessageSubscriptions lists message subscriptions via gRPC
// Выводит список подписок на сообщения через gRPC
func (d *DaemonCommand) MessageSubscriptions() error {
	logger.Debug("Listing message subscriptions")

	// Parse arguments for filtering and pagination
	var tenantID string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd message subscriptions"

	// Parse arguments: handle flags and positional arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--page" || arg == "-p" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &page); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if arg == "--page-size" || arg == "-s" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &pageSize); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			// Positional arguments
			if tenantID == "" {
				tenantID = arg
			}
		}
	}

	logger.Debug("Message subscriptions request",
		logger.String("tenant_id", tenantID),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

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
		TenantId:  tenantID,
		Limit:     0, // Use pagination instead
		PageSize:  pageSize,
		Page:      page,
		SortBy:    "created_at",
		SortOrder: "DESC",
	})
	if err != nil {
		logger.Error("Failed to list message subscriptions", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list message subscriptions: %w", err)
	}

	logger.Debug("Message subscriptions listed", logger.Int("count", len(resp.Subscriptions)))

	fmt.Printf("Message Subscriptions\n")
	fmt.Printf("====================\n")

	// Print pagination info if multiple pages exist
	if resp.TotalPages > 1 {
		fmt.Printf("Page %d of %d (Total: %d subscriptions, Showing: %d)\n\n",
			resp.Page, resp.TotalPages, resp.TotalCount, len(resp.Subscriptions))
	} else {
		fmt.Printf("Found %d subscription(s):\n\n", resp.TotalCount)
	}

	printMessageSubscriptionsTable(resp.Subscriptions, resp.TotalCount)

	// Show navigation hints for pagination
	if resp.TotalPages > 1 {
		fmt.Printf("\nNavigation:\n")

		// Previous page
		if resp.Page > 1 {
			prevPageCmd := fmt.Sprintf("atomd message subscriptions")
			if tenantID != "" {
				prevPageCmd += fmt.Sprintf(" %s", tenantID)
			}
			prevPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page-1, resp.PageSize)
			fmt.Printf("Previous page: %s\n", prevPageCmd)
		}

		// Next page
		if resp.Page < resp.TotalPages {
			nextPageCmd := fmt.Sprintf("atomd message subscriptions")
			if tenantID != "" {
				nextPageCmd += fmt.Sprintf(" %s", tenantID)
			}
			nextPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page+1, resp.PageSize)
			fmt.Printf("Next page: %s\n", nextPageCmd)
		}
	}

	return nil
}

// MessageBuffered lists buffered messages via gRPC
// Выводит список буферизованных сообщений через gRPC
func (d *DaemonCommand) MessageBuffered() error {
	logger.Debug("Listing buffered messages")

	fmt.Printf("Buffered Messages\n")
	fmt.Printf("=================\n")
	fmt.Printf("Use: atomd message list [tenant_id] [--page N] [--page-size N]\n")
	fmt.Printf("The 'message list' command shows buffered messages with pagination support.\n")

	return nil
}

// MessageCleanup cleans up expired messages via gRPC
// Очищает просроченные сообщения через gRPC
func (d *DaemonCommand) MessageCleanup() error {
	logger.Debug("Cleaning up expired messages")

	// Parse arguments for tenant filter
	var tenantID string
	args := os.Args[3:] // Skip "atomd message cleanup"

	// Parse arguments: handle positional arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			// Positional arguments
			if tenantID == "" {
				tenantID = arg
			}
		}
	}

	logger.Debug("Message cleanup request", logger.String("tenant_id", tenantID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for message cleanup",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := messagespb.NewMessagesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CleanupExpiredMessages(ctx, &messagespb.CleanupExpiredMessagesRequest{
		TenantId: tenantID,
	})
	if err != nil {
		logger.Error("Failed to cleanup expired messages", logger.String("error", err.Error()))
		return fmt.Errorf("failed to cleanup expired messages: %w", err)
	}

	fmt.Printf("Message Cleanup\n")
	fmt.Printf("===============\n")
	fmt.Printf("Success: %t\n", resp.Success)
	fmt.Printf("Message: %s\n", resp.Message)
	if resp.Success {
		fmt.Printf("Cleaned up messages: %d\n", resp.CleanedCount)
	}

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
	fmt.Printf("Testing basic message functionality...\n\n")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for message test",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := messagespb.NewMessagesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Publish a test message
	fmt.Printf("1. Testing message publish...")
	testMessageName := "test_message"
	testCorrelationKey := "test_key_123"
	testVariables := map[string]string{"test": "value"}

	pubResp, err := client.PublishMessage(ctx, &messagespb.PublishMessageRequest{
		TenantId:       "",
		MessageName:    testMessageName,
		CorrelationKey: testCorrelationKey,
		Variables:      testVariables,
		TtlSeconds:     60, // 1 minute TTL
	})

	if err != nil {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", err.Error())
	} else if !pubResp.Success {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", pubResp.Message)
	} else {
		fmt.Printf(" PASSED\n")
		fmt.Printf("   Message ID: %s\n", pubResp.MessageId)
	}

	// Test 2: List buffered messages
	fmt.Printf("\n2. Testing message list...")
	listResp, err := client.ListBufferedMessages(ctx, &messagespb.ListBufferedMessagesRequest{
		TenantId:  "",
		PageSize:  10,
		Page:      1,
		SortBy:    "published_at",
		SortOrder: "DESC",
	})

	if err != nil {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", err.Error())
	} else if !listResp.Success {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", listResp.Message)
	} else {
		fmt.Printf(" PASSED\n")
		fmt.Printf("   Found %d buffered messages\n", listResp.TotalCount)
	}

	// Test 3: Get message stats
	fmt.Printf("\n3. Testing message stats...")
	statsResp, err := client.GetMessageStats(ctx, &messagespb.GetMessageStatsRequest{})

	if err != nil {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", err.Error())
	} else if !statsResp.Success {
		fmt.Printf(" FAILED\n")
		fmt.Printf("   Error: %s\n", statsResp.Message)
	} else {
		fmt.Printf(" PASSED\n")
		fmt.Printf("   Total messages: %d, Buffered: %d\n",
			statsResp.Stats.TotalMessages, statsResp.Stats.BufferedMessages)
	}

	fmt.Printf("\nMessage test completed.\n")
	fmt.Printf("For detailed testing, use individual commands:\n")
	fmt.Printf("  atomd message publish <name> [correlation_key] [variables]\n")
	fmt.Printf("  atomd message list [tenant_id]\n")
	fmt.Printf("  atomd message stats\n")

	return nil
}
