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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"atom-engine/src/core/logger"
)

// GRPCClient handles gRPC connections
// gRPC клиент для подключений к демону
type GRPCClient struct {
	address string
}

// NewGRPCClient creates new gRPC client instance
// Создает новый экземпляр gRPC клиента
func NewGRPCClient() *GRPCClient {
	return &GRPCClient{
		address: "localhost:9090",
	}
}

// Connect establishes connection to gRPC server
// Устанавливает соединение с gRPC сервером
func (g *GRPCClient) Connect() (*grpc.ClientConn, error) {
	logger.Debug("Connecting to gRPC server", logger.String("address", g.address))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, g.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		logger.Error("Failed to connect to gRPC server",
			logger.String("address", g.address),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", g.address, err)
	}

	logger.Debug("Successfully connected to gRPC server", logger.String("address", g.address))

	return conn, nil
}
