/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"atom-engine/proto/storage/storagepb"
	"atom-engine/src/core/interfaces"
)

// Removed old type definitions, now using unified interfaces

// Type aliases for generated proto types
// Псевдонимы типов для сгенерированных proto типов
type GetStorageStatusRequest = storagepb.GetStorageStatusRequest
type GetStorageStatusResponse = storagepb.GetStorageStatusResponse
type GetStorageInfoRequest = storagepb.GetStorageInfoRequest
type GetStorageInfoResponse = storagepb.GetStorageInfoResponse

// Type aliases for interfaces package to maintain compatibility
// Псевдонимы типов из пакета interfaces для поддержания совместимости
type StorageStatusResponse = interfaces.StorageStatusResponse
type StorageInfoResponse = interfaces.StorageInfoResponse
type ProcessInstanceResult = interfaces.ProcessInstanceResult
type StorageServiceServer = storagepb.StorageServiceServer
type UnimplementedStorageServiceServer = storagepb.UnimplementedStorageServiceServer

// RegisterStorageServiceServer wraps the generated registration function
// Обертка для сгенерированной функции регистрации
var RegisterStorageServiceServer = storagepb.RegisterStorageServiceServer
