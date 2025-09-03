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
	"strings"
	"time"

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/src/core/logger"
)

// JobActivate activates jobs for a worker via gRPC
// Активирует работы для воркера через gRPC
func (d *DaemonCommand) JobActivate() error {
	logger.Debug("Activating jobs")

	if len(os.Args) < 5 {
		logger.Error("Invalid job activate arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job activate <type> <worker> [-j max_jobs] [-t timeout_ms]")
	}

	jobType := os.Args[3]
	worker := os.Args[4]

	// Default values
	var maxJobs int32 = 1
	var timeout int64 = 30000

	// Parse flags and remaining positional arguments (for backward compatibility)
	args := os.Args[5:] // Skip "atomd job activate type worker"

	// Parse flags first
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-j" && i+1 < len(args) {
			// Parse max jobs flag
			if mj, err := strconv.Atoi(args[i+1]); err == nil {
				maxJobs = int32(mj)
			} else {
				return fmt.Errorf("invalid value for -j flag: %s", args[i+1])
			}
			i++ // Skip the value
		} else if arg == "-t" && i+1 < len(args) {
			// Parse timeout flag
			if to, err := strconv.Atoi(args[i+1]); err == nil {
				timeout = int64(to)
			} else {
				return fmt.Errorf("invalid value for -t flag: %s", args[i+1])
			}
			i++ // Skip the value
		} else if !strings.HasPrefix(arg, "-") {
			// Unknown positional argument
			return fmt.Errorf("unknown argument: %s. Use -j for max_jobs or -t for timeout", arg)
		} else {
			// Unknown flag
			return fmt.Errorf("unknown flag: %s. Supported flags: -j (max_jobs), -t (timeout)", arg)
		}
	}

	logger.Debug("Job activate request",
		logger.String("job_type", jobType),
		logger.String("worker", worker),
		logger.Int("max_jobs", int(maxJobs)),
		logger.Int("timeout", int(timeout)))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job activate",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+5000)*time.Millisecond)
	defer cancel()

	// Call the streaming activation
	stream, err := client.ActivateJobs(ctx, &jobspb.ActivateJobsRequest{
		Type:              jobType,
		Worker:            worker,
		MaxJobsToActivate: maxJobs,
		Timeout:           int32(timeout),
	})
	if err != nil {
		logger.Error("Failed to activate jobs", logger.String("error", err.Error()))
		return fmt.Errorf("failed to activate jobs: %w", err)
	}

	fmt.Printf("Job Activation\n")
	fmt.Printf("==============\n")
	fmt.Printf("Job Type: %s\n", jobType)
	fmt.Printf("Worker: %s\n", worker)
	fmt.Printf("\n")

	activatedCount := 0
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}

		for _, job := range resp.Jobs {
			fmt.Printf("Activated Job:\n")
			fmt.Printf("  Key: %s\n", job.Key)
			fmt.Printf("  Type: %s\n", job.Type)
			fmt.Printf("  Process Instance: %s\n", job.ProcessInstanceKey)
			fmt.Printf("  Worker: %s\n", job.Worker)
			fmt.Printf("  Retries: %d\n", job.Retries)
			fmt.Printf("  Variables: %s\n", job.Variables)
			fmt.Printf("\n")
			activatedCount++
		}
	}

	fmt.Printf("Total activated: %d jobs\n", activatedCount)

	return nil
}

// JobComplete completes a job via gRPC
// Завершает работу через gRPC
func (d *DaemonCommand) JobComplete() error {
	logger.Debug("Completing job")

	if len(os.Args) < 4 {
		logger.Error("Invalid job complete arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job complete <job_key> [variables]")
	}

	jobKey := os.Args[3]
	var variables string
	if len(os.Args) > 4 {
		variables = os.Args[4]
	}

	logger.Debug("Job complete request",
		logger.String("job_key", jobKey),
		logger.String("variables", variables))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job complete",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.CompleteJob(ctx, &jobspb.CompleteJobRequest{
		JobKey:    jobKey,
		Variables: variables,
	})
	if err != nil {
		logger.Error("Failed to complete job", logger.String("error", err.Error()))
		return fmt.Errorf("failed to complete job: %w", err)
	}

	fmt.Printf("Job Completion\n")
	fmt.Printf("==============\n")
	fmt.Printf("Job Key: %s\n", jobKey)

	if resp.Success {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("SUCCESS"))
	} else {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("FAILED"))
		fmt.Printf("Error: %s\n", resp.ErrorMessage)
	}

	return nil
}

// JobFail fails a job via gRPC
// Провальная работа через gRPC
func (d *DaemonCommand) JobFail() error {
	logger.Debug("Failing job")

	if len(os.Args) < 5 {
		logger.Error("Invalid job fail arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job fail <job_key> <retries> [error] [backoff]")
	}

	jobKey := os.Args[3]
	retriesStr := os.Args[4]
	retries, err := strconv.Atoi(retriesStr)
	if err != nil {
		return fmt.Errorf("invalid retries value: %s", retriesStr)
	}

	var errorMessage string
	if len(os.Args) > 5 {
		errorMessage = os.Args[5]
	}

	logger.Debug("Job fail request",
		logger.String("job_key", jobKey),
		logger.Int("retries", retries),
		logger.String("error_message", errorMessage))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job fail",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.FailJob(ctx, &jobspb.FailJobRequest{
		JobKey:       jobKey,
		Retries:      int32(retries),
		ErrorMessage: errorMessage,
	})
	if err != nil {
		logger.Error("Failed to fail job", logger.String("error", err.Error()))
		return fmt.Errorf("failed to fail job: %w", err)
	}

	fmt.Printf("Job Failure\n")
	fmt.Printf("===========\n")
	fmt.Printf("Job Key: %s\n", jobKey)
	fmt.Printf("Retries: %d\n", retries)

	if resp.Success {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("SUCCESS"))
		fmt.Printf("Job failed successfully\n")
	} else {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("FAILED"))
		fmt.Printf("Error: %s\n", resp.ErrorMessage)
	}

	return nil
}

// JobCancel cancels a job via gRPC
// Отменяет работу через gRPC
func (d *DaemonCommand) JobCancel() error {
	logger.Debug("Cancelling job")

	if len(os.Args) < 4 {
		logger.Error("Invalid job cancel arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job cancel <job_key>")
	}

	jobKey := os.Args[3]

	logger.Debug("Job cancel request", logger.String("job_key", jobKey))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job cancel",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.CancelJob(ctx, &jobspb.CancelJobRequest{
		JobKey: jobKey,
	})
	if err != nil {
		logger.Error("Failed to cancel job", logger.String("error", err.Error()))
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	fmt.Printf("Job Cancellation\n")
	fmt.Printf("================\n")
	fmt.Printf("Job Key: %s\n", jobKey)

	if resp.Success {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("SUCCESS"))
	} else {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("FAILED"))
		fmt.Printf("Error: %s\n", resp.ErrorMessage)
	}

	return nil
}

// JobCreate creates a new job via gRPC
// Создает новую работу через gRPC
func (d *DaemonCommand) JobCreate() error {
	logger.Debug("Creating job")

	if len(os.Args) < 5 {
		logger.Error("Invalid job create arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job create <type> <process_instance_id> [variables]")
	}

	jobType := os.Args[3]
	processInstanceID := os.Args[4]
	var variables string
	if len(os.Args) > 5 {
		variables = os.Args[5]
	}

	logger.Debug("Job create request",
		logger.String("job_type", jobType),
		logger.String("process_instance_id", processInstanceID),
		logger.String("variables", variables))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job create",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.CreateJob(ctx, &jobspb.CreateJobRequest{
		Type:              jobType,
		ProcessInstanceId: processInstanceID,
		Variables:         variables,
	})
	if err != nil {
		logger.Error("Failed to create job", logger.String("error", err.Error()))
		return fmt.Errorf("failed to create job: %w", err)
	}

	fmt.Printf("Job Creation\n")
	fmt.Printf("============\n")
	fmt.Printf("Job Type: %s\n", jobType)
	fmt.Printf("Process Instance: %s\n", processInstanceID)

	if resp.Success {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("SUCCESS"))
		fmt.Printf("Job Key: %s\n", resp.JobKey)
	} else {
		fmt.Printf("Status: %s\n", ColorizeOperationStatus("FAILED"))
		fmt.Printf("Error: %s\n", resp.ErrorMessage)
	}

	return nil
}
