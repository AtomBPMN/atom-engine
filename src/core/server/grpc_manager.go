/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"

	"atom-engine/src/core/grpc"
	"atom-engine/src/core/logger"
)

// startGRPCServer starts gRPC server
// Запускает gRPC сервер
func (c *Core) startGRPCServer() error {
	grpcConfig := &grpc.Config{
		Port: c.config.GRPC.Port,
	}

	if grpcConfig.Port == 0 {
		grpcConfig.Port = 27500 // Default port
	}

	server := grpc.NewServer(grpcConfig, c)
	err := server.Start()
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	c.grpcServer = server
	return nil
}

// stopGRPCServer stops gRPC server
// Останавливает gRPC сервер
func (c *Core) stopGRPCServer() {
	if c.grpcServer != nil {
		err := c.grpcServer.Stop()
		if err != nil {
			logger.Error("Failed to stop gRPC server", logger.String("error", err.Error()))
		} else {
			logger.Info("gRPC server stopped")
		}
		c.grpcServer = nil
	}
}
