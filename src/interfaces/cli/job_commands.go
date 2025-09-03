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

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/src/core/logger"
)

// JobList lists jobs via gRPC
// Список работ через gRPC
func (d *DaemonCommand) JobList() error {
	logger.Debug("Listing jobs")

	// Parse arguments
	var jobType, worker, processInstanceID, processKey, state string
	var limit int32 = 50
	if len(os.Args) > 3 {
		jobType = os.Args[3]
	}
	if len(os.Args) > 4 {
		worker = os.Args[4]
	}
	if len(os.Args) > 5 {
		if l, err := strconv.Atoi(os.Args[5]); err == nil {
			limit = int32(l)
		}
	}

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job list",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.ListJobs(ctx, &jobspb.ListJobsRequest{
		Type:              jobType,
		Worker:            worker,
		ProcessInstanceId: processInstanceID,
		ProcessKey:        processKey,
		State:             state,
		Limit:             limit,
		IncludeVariables:  false,
	})
	if err != nil {
		logger.Error("Failed to list jobs", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	logger.Debug("Jobs listed", logger.Int("count", len(resp.Jobs)))

	fmt.Printf("Job List\n")
	fmt.Printf("========\n")
	printJobsTable(resp.Jobs, resp.TotalCount)

	return nil
}

// JobShow shows job details via gRPC
// Показывает детали работы через gRPC
func (d *DaemonCommand) JobShow() error {
	logger.Debug("Showing job details")

	if len(os.Args) < 4 {
		logger.Error("Invalid job show arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd job show <job_key>")
	}

	jobKey := os.Args[3]
	logger.Debug("Job show request", logger.String("job_key", jobKey))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job show",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetJob(ctx, &jobspb.GetJobRequest{
		JobKey: jobKey,
	})
	if err != nil {
		logger.Error("Failed to get job", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get job: %w", err)
	}

	fmt.Printf("Job Details\n")
	fmt.Printf("===========\n")

	if !resp.Found {
		fmt.Printf("Job not found: %s\n", jobKey)
		return nil
	}

	job := resp.Job
	fmt.Printf("Job Key: %s\n", job.Key)
	fmt.Printf("Type: %s\n", job.Type)
	fmt.Printf("Process Instance: %s\n", job.ProcessInstanceKey)
	fmt.Printf("Worker: %s\n", job.Worker)
	fmt.Printf("Status: %s\n", colorizeJobStatus(job.Status))
	fmt.Printf("Retries: %d\n", job.Retries)
	fmt.Printf("Created At: %s\n", time.Unix(job.CreatedAt, 0).Format("2006-01-02 15:04:05"))
	if job.ErrorMessage != "" {
		fmt.Printf("Error Message: %s\n", job.ErrorMessage)
	}
	if len(job.Variables) > 0 {
		fmt.Printf("Variables:\n")
		for k, v := range job.Variables {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}

// JobStats shows job statistics via gRPC
// Показывает статистику работ через gRPC
func (d *DaemonCommand) JobStats() error {
	logger.Debug("Getting job statistics")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for job stats",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	client := jobspb.NewJobsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.GetJobStats(ctx, &jobspb.GetJobStatsRequest{})
	if err != nil {
		logger.Error("Failed to get job stats", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get job stats: %w", err)
	}

	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.ErrorMessage)
		return nil
	}

	stats := resp.Stats
	logger.Debug("Job stats retrieved",
		logger.Int("total_jobs", int(stats.TotalJobs)),
		logger.Int("active_jobs", int(stats.ActiveJobs)))

	fmt.Printf("Job Statistics\n")
	fmt.Printf("==============\n")
	fmt.Printf("Total Jobs: %d\n", stats.TotalJobs)
	fmt.Printf("Active Jobs: %d\n", stats.ActiveJobs)
	fmt.Printf("Completed Jobs: %d\n", stats.CompletedJobs)
	fmt.Printf("Failed Jobs: %d\n", stats.FailedJobs)
	fmt.Printf("Activated Today: %d\n", stats.ActivatedToday)
	fmt.Printf("Completed Today: %d\n", stats.CompletedToday)

	return nil
}
