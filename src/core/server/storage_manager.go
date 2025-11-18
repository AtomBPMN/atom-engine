/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"
	"strconv"

	"atom-engine/src/core/grpc"
)

// GetStorageStatus returns storage status for gRPC
// Возвращает статус storage для gRPC
func (c *Core) GetStorageStatus() (*grpc.StorageStatusResponse, error) {
	if c.storage == nil {
		return &grpc.StorageStatusResponse{
			IsConnected:   false,
			IsHealthy:     false,
			Status:        "not_initialized",
			UptimeSeconds: 0,
		}, nil
	}

	status, err := c.storage.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage status: %w", err)
	}

	return &grpc.StorageStatusResponse{
		IsConnected:   status.IsConnected,
		IsHealthy:     status.IsHealthy,
		Status:        status.Status,
		UptimeSeconds: status.UptimeSeconds,
	}, nil
}

// GetStorageInfo returns storage info for gRPC
// Возвращает информацию storage для gRPC
func (c *Core) GetStorageInfo() (*grpc.StorageInfoResponse, error) {
	if c.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}

	info, err := c.storage.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage info: %w", err)
	}

	// Convert statistics from string to int64
	statistics := make(map[string]int64)
	for k, v := range info.Statistics {
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			statistics[k] = val
		} else {
			// If conversion fails, use 0 as default
			statistics[k] = 0
		}
	}

	return &grpc.StorageInfoResponse{
		TotalSizeBytes: info.TotalSizeBytes,
		UsedSizeBytes:  info.UsedSizeBytes,
		FreeSizeBytes:  info.FreeSizeBytes,
		TotalKeys:      info.TotalKeys,
		DatabasePath:   info.DatabasePath,
		Statistics:     statistics,
	}, nil
}
