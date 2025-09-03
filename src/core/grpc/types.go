/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"atom-engine/proto/storage/storagepb"
)

// StorageStatusResponse represents storage status response for internal use
// Представляет ответ со статусом storage для внутреннего использования
type StorageStatusResponse struct {
	IsConnected   bool   `json:"is_connected"`
	IsHealthy     bool   `json:"is_healthy"`
	Status        string `json:"status"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// StorageInfoResponse represents storage info response for internal use
// Представляет ответ с информацией storage для внутреннего использования
type StorageInfoResponse struct {
	TotalSizeBytes int64             `json:"total_size_bytes"`
	UsedSizeBytes  int64             `json:"used_size_bytes"`
	FreeSizeBytes  int64             `json:"free_size_bytes"`
	TotalKeys      int64             `json:"total_keys"`
	DatabasePath   string            `json:"database_path"`
	Statistics     map[string]string `json:"statistics"`
}

// Type aliases for generated proto types
// Псевдонимы типов для сгенерированных proto типов
type GetStorageStatusRequest = storagepb.GetStorageStatusRequest
type GetStorageStatusResponse = storagepb.GetStorageStatusResponse
type GetStorageInfoRequest = storagepb.GetStorageInfoRequest
type GetStorageInfoResponse = storagepb.GetStorageInfoResponse
type StorageServiceServer = storagepb.StorageServiceServer
type UnimplementedStorageServiceServer = storagepb.UnimplementedStorageServiceServer

// RegisterStorageServiceServer wraps the generated registration function
// Обертка для сгенерированной функции регистрации
var RegisterStorageServiceServer = storagepb.RegisterStorageServiceServer
