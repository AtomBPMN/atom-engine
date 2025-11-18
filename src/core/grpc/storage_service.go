/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"fmt"
)

// storageServiceServer implements StorageService gRPC interface
// Реализует gRPC интерфейс StorageService
type storageServiceServer struct {
	UnimplementedStorageServiceServer
	core CoreInterface
}

// GetStorageStatus returns storage status via gRPC
// Возвращает статус storage через gRPC
func (s *storageServiceServer) GetStorageStatus(
	ctx context.Context,
	req *GetStorageStatusRequest,
) (*GetStorageStatusResponse, error) {
	status, err := s.core.GetStorageStatus()
	if err != nil {
		return nil, err
	}

	return &GetStorageStatusResponse{
		IsConnected:   status.IsConnected,
		IsHealthy:     status.IsHealthy,
		Status:        status.Status,
		UptimeSeconds: status.UptimeSeconds,
	}, nil
}

// GetStorageInfo returns storage info via gRPC
// Возвращает информацию storage через gRPC
func (s *storageServiceServer) GetStorageInfo(
	ctx context.Context,
	req *GetStorageInfoRequest,
) (*GetStorageInfoResponse, error) {
	info, err := s.core.GetStorageInfo()
	if err != nil {
		return nil, err
	}

	// Convert statistics from int64 to string
	statistics := make(map[string]string)
	for k, v := range info.Statistics {
		statistics[k] = fmt.Sprintf("%d", v)
	}

	return &GetStorageInfoResponse{
		TotalSizeBytes: info.TotalSizeBytes,
		UsedSizeBytes:  info.UsedSizeBytes,
		FreeSizeBytes:  info.FreeSizeBytes,
		TotalKeys:      info.TotalKeys,
		DatabasePath:   info.DatabasePath,
		Statistics:     statistics,
	}, nil
}
