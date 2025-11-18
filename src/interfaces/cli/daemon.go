/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"os"

	"atom-engine/src/core/server"
)

// DaemonCommand handles daemon operations
// Обработчик операций демона
type DaemonCommand struct {
	executable string
	core       *server.Core
	grpcClient *GRPCClient
}

// NewDaemonCommand creates new daemon command handler
// Создает новый обработчик команд демона
func NewDaemonCommand() *DaemonCommand {
	executable, _ := os.Executable()

	// Try to create gRPC client from config, fallback to default if config fails
	grpcClient, err := NewGRPCClientFromConfig()
	if err != nil {
		// Use default client if config loading fails
		grpcClient = NewGRPCClient()
	}

	return &DaemonCommand{
		executable: executable,
		grpcClient: grpcClient,
	}
}

// All command implementations have been moved to separate files:
// - daemon_commands.go - daemon lifecycle management (start, stop, run, status, events)
// - storage_commands.go - storage operations (status, info)
// - timer_commands.go - timer management (add, remove, status, list, stats)
// - process_commands.go - process instance management (start, status, cancel, list)
// - token_commands.go - token operations (list, show, trace)
// - job_commands.go - job management (list, show, activate, complete, fail, cancel, create, stats)
// - message_commands.go - message operations (publish, list, subscriptions, buffered, cleanup, stats, test)
// - bpmn_commands.go - BPMN process management (parse, list, show, delete, stats, json)
// - expression_commands.go - expression evaluation (eval, validate, parse, functions, test)
// - grpc_client.go - common gRPC connection handling
// - formatters.go - output formatting utilities
// - help.go - help message functions
