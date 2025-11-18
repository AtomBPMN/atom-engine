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

		// Convert to string map for protobuf, preserving JSON structure
		for key, value := range jsonVars {
			// For complex objects, serialize back to JSON
			// Для сложных объектов сериализуем обратно в JSON
			switch v := value.(type) {
			case map[string]interface{}, []interface{}, map[string]string, []string:
				// Serialize complex types back to JSON
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					logger.Warn("Failed to serialize variable to JSON, using string representation",
						logger.String("key", key),
						logger.String("error", err.Error()))
					variablesMap[key] = fmt.Sprintf("%v", value)
				} else {
					variablesMap[key] = string(jsonBytes)
				}
			default:
				// For simple types, use string representation
				variablesMap[key] = fmt.Sprintf("%v", value)
			}
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

	// Parse arguments for filtering and pagination
	var statusFilter string
	var processKeyFilter string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd process list"

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
			if statusFilter == "" {
				statusFilter = arg
			} else if processKeyFilter == "" {
				processKeyFilter = arg
			}
		}
	}

	logger.Debug("Process list request",
		logger.String("status_filter", statusFilter),
		logger.String("process_key_filter", processKeyFilter),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

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
		ProcessKeyFilter: processKeyFilter,
		Limit:            0, // Use pagination instead
		PageSize:         pageSize,
		Page:             page,
		SortBy:           "started_at",
		SortOrder:        "DESC",
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

	// Print pagination info if multiple pages exist
	if response.TotalPages > 1 {
		fmt.Printf("Page %d of %d (Total: %d instances, Showing: %d)\n\n",
			response.Page, response.TotalPages, response.TotalCount, len(response.Instances))
	} else {
		fmt.Printf("Found %d instance(s):\n\n", response.TotalCount)
	}

	printProcessInstancesTable(response.Instances, response.TotalCount)

	// Show navigation hints for pagination
	if response.TotalPages > 1 {
		fmt.Printf("\nNavigation:\n")

		// Previous page
		if response.Page > 1 {
			prevPageCmd := fmt.Sprintf("atomd process list")
			if statusFilter != "" {
				prevPageCmd += fmt.Sprintf(" %s", statusFilter)
			}
			if processKeyFilter != "" {
				prevPageCmd += fmt.Sprintf(" %s", processKeyFilter)
			}
			prevPageCmd += fmt.Sprintf(" --page %d --page-size %d", response.Page-1, response.PageSize)
			fmt.Printf("Previous page: %s\n", prevPageCmd)
		}

		// Next page
		if response.Page < response.TotalPages {
			nextPageCmd := fmt.Sprintf("atomd process list")
			if statusFilter != "" {
				nextPageCmd += fmt.Sprintf(" %s", statusFilter)
			}
			if processKeyFilter != "" {
				nextPageCmd += fmt.Sprintf(" %s", processKeyFilter)
			}
			nextPageCmd += fmt.Sprintf(" --page %d --page-size %d", response.Page+1, response.PageSize)
			fmt.Printf("Next page: %s\n", nextPageCmd)
		}
	}

	return nil
}

// ProcessInfo gets complete process instance information via gRPC
// Получает полную информацию об экземпляре процесса через gRPC
func (d *DaemonCommand) ProcessInfo() error {
	logger.Debug("Getting process info")

	if len(os.Args) < 4 {
		logger.Error("Invalid process info arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd process info <instance_id>")
	}

	instanceID := os.Args[3]
	logger.Debug("Process info request", logger.String("instance_id", instanceID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for process info", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := client.GetProcessInstanceInfo(ctx, &processpb.GetProcessInstanceInfoRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		logger.Error("Failed to get process info via gRPC",
			logger.String("instance_id", instanceID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to get process instance info: %w", err)
	}

	if !response.Success {
		logger.Warn("Process info failed",
			logger.String("instance_id", instanceID),
			logger.String("message", response.Message))
		return fmt.Errorf("failed to get process instance info: %s", response.Message)
	}

	logger.Debug("Process info retrieved",
		logger.String("instance_id", response.InstanceId),
		logger.String("status", response.Status))

	// Print process basic information
	fmt.Printf("Process Instance Complete Information\n")
	fmt.Printf("====================================\n\n")

	fmt.Printf("BASIC INFORMATION:\n")
	fmt.Printf("  Instance ID:      %s\n", response.InstanceId)
	fmt.Printf("  Process Key:      %s\n", response.ProcessKey)
	if response.BpmnProcessKey != "" {
		fmt.Printf("  BPMN Process Key: %s\n", response.BpmnProcessKey)
	}
	fmt.Printf("  Status:           %s\n", colorizeStatus(response.Status))
	fmt.Printf("  Current Activity: %s\n", response.CurrentActivity)
	fmt.Printf("  Started At:       %s\n", time.Unix(response.StartedAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("  Updated At:       %s\n", time.Unix(response.UpdatedAt, 0).Format("2006-01-02 15:04:05"))

	// Print variables
	if len(response.Variables) > 0 {
		fmt.Printf("\nVARIABLES:\n")
		for key, value := range response.Variables {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	// Print tokens information
	if len(response.Tokens) > 0 {
		fmt.Printf("\nTOKENS (%d):\n", len(response.Tokens))
		for _, token := range response.Tokens {
			fmt.Printf("  ID: %s\n", token.TokenId)
			fmt.Printf("    Element: %s\n", token.CurrentElementId)
			fmt.Printf("    State: %s\n", colorizeStatus(token.State))
			if token.WaitingFor != "" {
				fmt.Printf("    Waiting For: %s\n", token.WaitingFor)
			}
			fmt.Printf("    Created: %s\n", time.Unix(token.CreatedAt, 0).Format("2006-01-02 15:04:05"))
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("\nTOKENS: None\n")
	}

	// Print external services details
	if response.ExternalServices != nil {
		// Print timers
		if len(response.ExternalServices.Timers) > 0 {
			fmt.Printf("TIMERS (%d):\n", len(response.ExternalServices.Timers))
			for _, timer := range response.ExternalServices.Timers {
				fmt.Printf("  ID: %s\n", timer.TimerId)
				fmt.Printf("    Element: %s\n", timer.ElementId)
				fmt.Printf("    Type: %s\n", timer.TimerType)
				fmt.Printf("    Status: %s\n", colorizeStatus(timer.Status))
				if timer.TimeDuration != "" {
					fmt.Printf("    Duration: %s\n", timer.TimeDuration)
				}
				if timer.TimeCycle != "" {
					fmt.Printf("    Cycle: %s\n", timer.TimeCycle)
				}
				fmt.Printf("    Scheduled: %s\n", time.Unix(timer.ScheduledAt, 0).Format("2006-01-02 15:04:05"))
				if timer.RemainingSeconds > 0 {
					fmt.Printf("    Remaining: %ds\n", timer.RemainingSeconds)
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("TIMERS: None\n\n")
		}

		// Print jobs
		if len(response.ExternalServices.Jobs) > 0 {
			fmt.Printf("JOBS (%d):\n", len(response.ExternalServices.Jobs))
			for _, job := range response.ExternalServices.Jobs {
				fmt.Printf("  Key: %s\n", job.Key)
				fmt.Printf("    Type: %s\n", job.Type)
				fmt.Printf("    Element: %s\n", job.ElementId)
				fmt.Printf("    Status: %s\n", colorizeStatus(job.Status))
				if job.Worker != "" {
					fmt.Printf("    Worker: %s\n", job.Worker)
				}
				fmt.Printf("    Retries: %d\n", job.Retries)
				fmt.Printf("    Created: %s\n", time.Unix(job.CreatedAt, 0).Format("2006-01-02 15:04:05"))
				if job.ErrorMessage != "" {
					fmt.Printf("    Error: %s\n", job.ErrorMessage)
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("JOBS: None\n\n")
		}

		// Print message subscriptions
		if len(response.ExternalServices.MessageSubscriptions) > 0 {
			fmt.Printf("MESSAGE SUBSCRIPTIONS (%d):\n", len(response.ExternalServices.MessageSubscriptions))
			for _, sub := range response.ExternalServices.MessageSubscriptions {
				fmt.Printf("  ID: %s\n", sub.Id)
				fmt.Printf("    Message: %s\n", sub.MessageName)
				if sub.CorrelationKey != "" {
					fmt.Printf("    Correlation Key: %s\n", sub.CorrelationKey)
				}
				fmt.Printf("    Start Event: %s\n", sub.StartEventId)
				fmt.Printf("    Active: %t\n", sub.IsActive)
				fmt.Printf("    Created: %s\n", time.Unix(sub.CreatedAt, 0).Format("2006-01-02 15:04:05"))
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("MESSAGE SUBSCRIPTIONS: None\n\n")
		}

		// Print buffered messages
		if len(response.ExternalServices.BufferedMessages) > 0 {
			fmt.Printf("BUFFERED MESSAGES (%d):\n", len(response.ExternalServices.BufferedMessages))
			for _, msg := range response.ExternalServices.BufferedMessages {
				fmt.Printf("  ID: %s\n", msg.Id)
				fmt.Printf("    Name: %s\n", msg.Name)
				if msg.CorrelationKey != "" {
					fmt.Printf("    Correlation Key: %s\n", msg.CorrelationKey)
				}
				if msg.ElementId != "" {
					fmt.Printf("    Element: %s\n", msg.ElementId)
				}
				fmt.Printf("    Published: %s\n", time.Unix(msg.PublishedAt, 0).Format("2006-01-02 15:04:05"))
				fmt.Printf("    Expires: %s\n", time.Unix(msg.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
				if msg.Reason != "" {
					fmt.Printf("    Reason: %s\n", msg.Reason)
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("BUFFERED MESSAGES: None\n\n")
		}

		// Print incidents
		if len(response.ExternalServices.Incidents) > 0 {
			fmt.Printf("INCIDENTS (%d):\n", len(response.ExternalServices.Incidents))
			for _, incident := range response.ExternalServices.Incidents {
				fmt.Printf("  ID: %s\n", incident.Id)
				fmt.Printf("    Type: %s\n", incident.Type)
				fmt.Printf("    Status: %s\n", colorizeStatus(incident.Status))
				fmt.Printf("    Message: %s\n", incident.Message)
				if incident.ErrorCode != "" {
					fmt.Printf("    Error Code: %s\n", incident.ErrorCode)
				}
				if incident.ElementId != "" {
					fmt.Printf("    Element: %s\n", incident.ElementId)
				}
				if incident.JobKey != "" {
					fmt.Printf("    Job Key: %s\n", incident.JobKey)
				}
				fmt.Printf("    Created: %s\n", time.Unix(incident.CreatedAt, 0).Format("2006-01-02 15:04:05"))
				fmt.Printf("\n")
			}
		} else {
			fmt.Printf("INCIDENTS: None\n\n")
		}
	}

	fmt.Printf("\n")

	return nil
}
