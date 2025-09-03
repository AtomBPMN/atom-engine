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
	"os"
	"strconv"
	"time"

	"atom-engine/proto/parser/parserpb"
	"atom-engine/src/core/logger"
)

// BPMNParse parses BPMN file via gRPC
// Парсит BPMN файл через gRPC
func (d *DaemonCommand) BPMNParse() error {
	logger.Debug("Parsing BPMN file")

	if len(os.Args) < 4 {
		logger.Error("Invalid BPMN parse arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd bpmn parse <file.bpmn> [process_id] [--force|-f]")
	}

	filename := os.Args[3]
	var processID string
	var force bool

	// Parse optional arguments
	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--force" || arg == "-f" {
			force = true
		} else if processID == "" {
			processID = arg
		}
	}

	logger.Debug("BPMN parse request",
		logger.String("filename", filename),
		logger.String("process_id", processID),
		logger.Bool("force", force))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN parse",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.ParseBPMNFile(ctx, &parserpb.ParseBPMNFileRequest{
		FilePath:  filename,
		ProcessId: processID,
		Force:     force,
	})
	if err != nil {
		logger.Error("Failed to parse BPMN file", logger.String("error", err.Error()))
		return fmt.Errorf("failed to parse BPMN file: %w", err)
	}

	logger.Debug("BPMN parse completed",
		logger.Bool("success", resp.Success),
		logger.String("bpmn_id", resp.BpmnId),
		logger.Int("total_elements", int(resp.TotalElements)))

	fmt.Printf("BPMN Parse Results\n")
	fmt.Printf("==================\n")
	fmt.Printf("File: %s\n", filename)
	fmt.Printf("Success: %t\n", resp.Success)
	fmt.Printf("Message: %s\n", resp.Message)
	if resp.Success {
		fmt.Printf("BPMN ID: %s\n", resp.BpmnId)
		fmt.Printf("Process ID: %s\n", resp.ProcessId)
		fmt.Printf("Process Name: %s\n", resp.ProcessName)
		fmt.Printf("Total Elements: %d\n", resp.TotalElements)
		fmt.Printf("Successful: %d\n", resp.SuccessfulElements)
		fmt.Printf("Generic: %d\n", resp.GenericElements)
		fmt.Printf("Failed: %d\n", resp.FailedElements)
	}

	return nil
}

// BPMNList lists BPMN processes via gRPC
// Выводит список BPMN процессов через gRPC
func (d *DaemonCommand) BPMNList() error {
	logger.Debug("Listing BPMN processes")

	// Parse limit from arguments
	var limit int32 = 0
	if len(os.Args) > 3 {
		if l, err := strconv.Atoi(os.Args[3]); err == nil {
			limit = int32(l)
		}
	}

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN list",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListBPMNProcesses(ctx, &parserpb.ListBPMNProcessesRequest{
		Limit: limit,
	})
	if err != nil {
		logger.Error("Failed to list BPMN processes", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list BPMN processes: %w", err)
	}

	logger.Debug("BPMN processes listed", logger.Int("count", len(resp.Processes)))

	fmt.Printf("BPMN Process List\n")
	fmt.Printf("================\n")
	printBPMNProcessesTable(resp.Processes, resp.TotalCount)

	return nil
}

// BPMNShow shows BPMN process details via gRPC
// Показывает детали BPMN процесса через gRPC
func (d *DaemonCommand) BPMNShow() error {
	logger.Debug("Showing BPMN process details")

	if len(os.Args) < 4 {
		logger.Error("Invalid BPMN show arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd bpmn show <process_key>")
	}

	processKey := os.Args[3]
	logger.Debug("BPMN show request", logger.String("process_key", processKey))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN show",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetBPMNProcess(ctx, &parserpb.GetBPMNProcessRequest{
		ProcessKey: processKey,
	})
	if err != nil {
		logger.Error("Failed to get BPMN process", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		return nil
	}

	process := resp.Process
	logger.Debug("BPMN process details retrieved",
		logger.String("process_key", process.ProcessKey),
		logger.String("process_name", process.ProcessName))

	fmt.Printf("BPMN Process Details\n")
	fmt.Printf("===================\n")
	fmt.Printf("Process Key: %s\n", process.ProcessKey)
	fmt.Printf("Process ID: %s\n", process.ProcessId)
	fmt.Printf("Name: %s\n", process.ProcessName)
	fmt.Printf("Version: %s\n", process.Version)
	fmt.Printf("Process Version: %d\n", process.ProcessVersion)
	fmt.Printf("Status: %s\n", colorizeStatus(process.Status))
	fmt.Printf("Total Elements: %d\n", process.TotalElements)
	fmt.Printf("Content Hash: %s\n", process.ContentHash)
	fmt.Printf("Original File: %s\n", process.OriginalFile)
	fmt.Printf("Created At: %s\n", process.CreatedAt)
	fmt.Printf("Updated At: %s\n", process.UpdatedAt)
	fmt.Printf("Parsed At: %s\n", process.ParsedAt)

	if len(process.ElementCounts) > 0 {
		fmt.Printf("\nElement Counts:\n")
		for elementType, count := range process.ElementCounts {
			fmt.Printf("  %s: %d\n", elementType, count)
		}
	}

	return nil
}

// BPMNDelete deletes BPMN process via gRPC
// Удаляет BPMN процесс через gRPC
func (d *DaemonCommand) BPMNDelete() error {
	logger.Debug("Deleting BPMN process")

	if len(os.Args) < 4 {
		logger.Error("Invalid BPMN delete arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd bpmn delete <process_id>")
	}

	processID := os.Args[3]
	logger.Debug("BPMN delete request", logger.String("process_id", processID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN delete",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.DeleteBPMNProcess(ctx, &parserpb.DeleteBPMNProcessRequest{
		ProcessId: processID,
	})
	if err != nil {
		logger.Error("Failed to delete BPMN process", logger.String("error", err.Error()))
		return fmt.Errorf("failed to delete BPMN process: %w", err)
	}

	logger.Debug("BPMN process delete completed",
		logger.Bool("success", resp.Success),
		logger.String("message", resp.Message))

	fmt.Printf("BPMN Process Delete\n")
	fmt.Printf("==================\n")
	fmt.Printf("Process ID: %s\n", processID)
	fmt.Printf("Success: %t\n", resp.Success)
	fmt.Printf("Message: %s\n", resp.Message)

	return nil
}

// BPMNStats shows BPMN statistics via gRPC
// Показывает статистику BPMN через gRPC
func (d *DaemonCommand) BPMNStats() error {
	logger.Debug("Getting BPMN statistics")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN stats",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetBPMNStats(ctx, &parserpb.GetBPMNStatsRequest{})
	if err != nil {
		logger.Error("Failed to get BPMN stats", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get BPMN stats: %w", err)
	}

	logger.Debug("BPMN stats retrieved",
		logger.Int("total_processes", int(resp.TotalProcesses)),
		logger.Int("active_processes", int(resp.ActiveProcesses)))

	fmt.Printf("BPMN Statistics\n")
	fmt.Printf("===============\n")
	fmt.Printf("Total Processes: %d\n", resp.TotalProcesses)
	fmt.Printf("Active Processes: %d\n", resp.ActiveProcesses)
	fmt.Printf("Total Elements Parsed: %d\n", resp.TotalElementsParsed)
	fmt.Printf("Successful Elements: %d\n", resp.SuccessfulElements)
	fmt.Printf("Generic Elements: %d\n", resp.GenericElements)
	fmt.Printf("Failed Elements: %d\n", resp.FailedElements)
	if resp.LastParsedAt != "" {
		fmt.Printf("Last Parsed At: %s\n", resp.LastParsedAt)
	}

	if len(resp.ElementTypeCounts) > 0 {
		fmt.Printf("\nElement Type Counts:\n")
		for elementType, count := range resp.ElementTypeCounts {
			fmt.Printf("  %s: %d\n", elementType, count)
		}
	}

	return nil
}

// BPMNJson shows BPMN process JSON data via gRPC
// Показывает JSON данные BPMN процесса через gRPC
func (d *DaemonCommand) BPMNJson() error {
	logger.Debug("Getting BPMN process JSON")

	if len(os.Args) < 4 {
		logger.Error("Invalid BPMN json arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd bpmn json <process_key>")
	}

	processKey := os.Args[3]
	logger.Debug("BPMN json request", logger.String("process_key", processKey))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for BPMN json",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := parserpb.NewParserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetBPMNProcessJSON(ctx, &parserpb.GetBPMNProcessJSONRequest{
		ProcessKey: processKey,
	})
	if err != nil {
		logger.Error("Failed to get BPMN process JSON", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get BPMN process JSON: %w", err)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Message)
		return nil
	}

	logger.Debug("BPMN process JSON retrieved",
		logger.String("process_key", processKey),
		logger.Int("json_length", len(resp.JsonData)))

	fmt.Printf("BPMN Process JSON\n")
	fmt.Printf("================\n")
	fmt.Printf("%s\n", resp.JsonData)

	return nil
}
