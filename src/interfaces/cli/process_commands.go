/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"atom-engine/proto/process/processpb"
	"atom-engine/src/core/logger"
)

// ProcessStart starts process instance via gRPC
// Запускает экземпляр процесса через gRPC
func (d *DaemonCommand) ProcessStart() error {
	logger.Debug("Starting process instance")

	if len(os.Args) < 4 {
		logger.Error("Invalid process start arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd process start <process_key> [-v version] [-d variables]")
	}

	// Parse arguments and flags
	var processKey string
	var version string
	var variables string

	args := os.Args[3:] // Skip "atomd process start"
	for i, arg := range args {
		if arg == "-v" || arg == "--version" {
			if i+1 < len(args) {
				version = args[i+1]
			}
		} else if arg == "-d" || arg == "--data" {
			if i+1 < len(args) {
				variables = args[i+1]
			}
		} else if processKey == "" && !strings.HasPrefix(arg, "-") {
			processKey = arg
		}
	}

	if processKey == "" {
		logger.Error("Process key not provided")
		return fmt.Errorf("usage: atomd process start <process_key> [-v version] [-d variables]")
	}

	logger.Debug("Process start request",
		logger.String("process_key", processKey),
		logger.String("version", version),
		logger.String("variables", variables))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for process start", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create process gRPC client
	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse variables if provided
	variablesMap := make(map[string]string)
	if variables != "" {
		var jsonVars map[string]interface{}
		if err := json.Unmarshal([]byte(variables), &jsonVars); err != nil {
			logger.Error("Failed to parse JSON variables",
				logger.String("variables", variables),
				logger.String("error", err.Error()))
			return fmt.Errorf("invalid JSON variables: %w", err)
		}

		// Convert to string map for protobuf
		for key, value := range jsonVars {
			variablesMap[key] = fmt.Sprintf("%v", value)
		}
		logger.Debug("Parsed variables", logger.Int("var_count", len(variablesMap)))
	}

	// Construct final process key with version if specified
	finalProcessKey := processKey
	if version != "" {
		finalProcessKey = fmt.Sprintf("%s:v%s", processKey, version)
	}

	logger.Debug("Starting process with final key", logger.String("final_process_key", finalProcessKey))

	response, err := client.StartProcessInstance(ctx, &processpb.StartProcessInstanceRequest{
		ProcessId: finalProcessKey,
		Variables: variablesMap,
	})
	if err != nil {
		logger.Error("Failed to start process instance via gRPC",
			logger.String("process_key", finalProcessKey),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to start process instance: %w", err)
	}

	if !response.Success {
		logger.Warn("Process start failed",
			logger.String("process_key", finalProcessKey),
			logger.String("message", response.Message))
		return fmt.Errorf("process start failed: %s", response.Message)
	}

	logger.Info("Process instance started successfully",
		logger.String("instance_id", response.InstanceId),
		logger.String("status", response.Status))

	fmt.Printf("Process instance started successfully\n")
	fmt.Printf("Instance ID: %s\n", response.InstanceId)
	fmt.Printf("Status: %s\n", colorizeStatus(response.Status))
	fmt.Printf("Message: %s\n", response.Message)

	return nil
}

// ProcessStatus gets process instance status via gRPC
// Получает статус экземпляра процесса через gRPC
func (d *DaemonCommand) ProcessStatus() error {
	logger.Debug("Getting process status")

	if len(os.Args) < 4 {
		logger.Error("Invalid process status arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd process status <instance_id>")
	}

	instanceID := os.Args[3]
	logger.Debug("Process status request", logger.String("instance_id", instanceID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for process status", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.GetProcessInstanceStatus(ctx, &processpb.GetProcessInstanceStatusRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		logger.Error("Failed to get process status via gRPC",
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to get process instance status: %w", err)
	}

	logger.Debug("Process status retrieved",
		logger.String("instance_id", response.InstanceId),
		logger.String("status", response.Status))

	fmt.Printf("Process Instance Status\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Instance ID:      %s\n", response.InstanceId)
	fmt.Printf("Status:           %s\n", colorizeStatus(response.Status))
	fmt.Printf("Current Activity: %s\n", response.CurrentActivity)
	fmt.Printf("Started At:       %s\n", time.Unix(response.StartedAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:       %s\n", time.Unix(response.UpdatedAt, 0).Format("2006-01-02 15:04:05"))

	if len(response.Variables) > 0 {
		fmt.Printf("\nVariables:\n")
		for key, value := range response.Variables {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

// ProcessCancel cancels process instance via gRPC
// Отменяет экземпляр процесса через gRPC
func (d *DaemonCommand) ProcessCancel() error {
	logger.Debug("Cancelling process instance")

	if len(os.Args) < 4 {
		logger.Error("Invalid process cancel arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd process cancel <instance_id> [reason]")
	}

	instanceID := os.Args[3]
	reason := "Canceled by user"
	if len(os.Args) >= 5 {
		reason = os.Args[4]
	}

	logger.Debug("Process cancel request",
		logger.String("instance_id", instanceID),
		logger.String("reason", reason))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for process cancel", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.CancelProcessInstance(ctx, &processpb.CancelProcessInstanceRequest{
		InstanceId: instanceID,
		Reason:     reason,
	})
	if err != nil {
		logger.Error("Failed to cancel process via gRPC",
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to cancel process instance: %w", err)
	}

	if !response.Success {
		logger.Warn("Process cancel failed",
			logger.String("instance_id", instanceID),
			logger.String("message", response.Message))
		return fmt.Errorf("process cancel failed: %s", response.Message)
	}

	logger.Info("Process instance cancelled successfully",
		logger.String("instance_id", response.InstanceId),
		logger.String("reason", reason))

	fmt.Printf("Process instance canceled successfully\n")
	fmt.Printf("Instance ID: %s\n", response.InstanceId)
	fmt.Printf("Message: %s\n", response.Message)

	return nil
}

// ProcessList lists process instances via gRPC
// Выводит список экземпляров процессов через gRPC
func (d *DaemonCommand) ProcessList() error {
	logger.Debug("Listing process instances")

	// Parse arguments for optional filters
	var statusFilter string
	var processKeyFilter string
	var limit int32

	args := os.Args[3:] // Skip "atomd process list"
	for i, arg := range args {
		switch {
		case i == 0 && !strings.HasPrefix(arg, "-"):
			// First positional argument is status filter
			statusFilter = arg
		case i == 1 && !strings.HasPrefix(arg, "-"):
			// Second positional argument is limit
			if l, err := strconv.Atoi(arg); err == nil {
				limit = int32(l)
			}
		case i == 0 && arg != "":
			// If first arg is empty string, it's still valid
			statusFilter = arg
		}
	}

	logger.Debug("Process list request",
		logger.String("status_filter", statusFilter),
		logger.String("process_key_filter", processKeyFilter),
		logger.Int("limit", int(limit)))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for process list", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create process gRPC client
	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.ListProcessInstances(ctx, &processpb.ListProcessInstancesRequest{
		StatusFilter:     statusFilter,
		Limit:            limit,
		ProcessKeyFilter: processKeyFilter,
	})
	if err != nil {
		logger.Error("Failed to list process instances via gRPC", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list process instances: %w", err)
	}

	if !response.Success {
		logger.Warn("Process list failed", logger.String("message", response.Message))
		return fmt.Errorf("failed to list process instances: %s", response.Message)
	}

	logger.Debug("Process list retrieved",
		logger.Int("instances_count", len(response.Instances)))

	fmt.Printf("Process Instance List\n")
	fmt.Printf("====================\n")
	printProcessInstancesTable(response.Instances, int32(len(response.Instances)))

	return nil
}
