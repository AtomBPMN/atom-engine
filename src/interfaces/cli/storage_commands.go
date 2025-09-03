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
	"time"

	"atom-engine/proto/storage/storagepb"
	"atom-engine/src/core/logger"
)

// StorageStatus shows storage status via gRPC
// Показывает статус storage через gRPC
func (d *DaemonCommand) StorageStatus() error {
	logger.Debug("Getting storage status")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for storage status",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	// Create real gRPC client
	client := storagepb.NewStorageServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.GetStorageStatus(ctx, &storagepb.GetStorageStatusRequest{})
	if err != nil {
		logger.Error("Failed to get storage status", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get storage status: %w", err)
	}

	logger.Debug("Storage status retrieved successfully",
		logger.Bool("connected", response.IsConnected),
		logger.Bool("healthy", response.IsHealthy))

	fmt.Println("Storage Status:")
	fmt.Println("===============")
	connectedStr, healthyStr := ColorizeConnectionStatus(response.IsConnected, response.IsHealthy)
	fmt.Printf("Connected:    %s\n", connectedStr)
	fmt.Printf("Healthy:      %s\n", healthyStr)
	fmt.Printf("Status:       %s\n", colorizeStatus(response.Status))
	if response.UptimeSeconds > 0 {
		fmt.Printf("Uptime:       %s\n", time.Duration(response.UptimeSeconds)*time.Second)
	}

	return nil
}

// StorageInfo shows storage information via gRPC
// Показывает информацию storage через gRPC
func (d *DaemonCommand) StorageInfo() error {
	logger.Debug("Getting storage info")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for storage info",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	// Create real gRPC client
	client := storagepb.NewStorageServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.GetStorageInfo(ctx, &storagepb.GetStorageInfoRequest{})
	if err != nil {
		logger.Error("Failed to get storage info", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get storage info: %w", err)
	}

	logger.Debug("Storage info retrieved successfully",
		logger.String("path", response.DatabasePath),
		logger.Int64("total_keys", response.TotalKeys))

	fmt.Println("Storage Information:")
	fmt.Println("===================")
	fmt.Printf("Database Path:    %s\n", response.DatabasePath)
	fmt.Printf("Total Keys:       %d\n", response.TotalKeys)
	fmt.Printf("Used Size:        %s\n", formatBytes(response.UsedSizeBytes))
	if response.TotalSizeBytes > 0 {
		fmt.Printf("Total Size:       %s\n", formatBytes(response.TotalSizeBytes))
	}
	if response.FreeSizeBytes > 0 {
		fmt.Printf("Free Size:        %s\n", formatBytes(response.FreeSizeBytes))
	}

	if len(response.Statistics) > 0 {
		fmt.Println("\nStatistics:")
		for key, value := range response.Statistics {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}
