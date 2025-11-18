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
	"strings"
	"time"

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/src/core/logger"
)

// JobList lists jobs via gRPC
// Список работ через gRPC
func (d *DaemonCommand) JobList() error {
	logger.Debug("Listing jobs")

	// Parse arguments for filtering and pagination
	var jobType, worker, processInstanceID, processKey, state string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd job list"

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
			if jobType == "" {
				jobType = arg
			} else if worker == "" {
				worker = arg
			} else if processInstanceID == "" {
				processInstanceID = arg
			} else if processKey == "" {
				processKey = arg
			} else if state == "" {
				state = arg
			}
		}
	}

	logger.Debug("Job list request",
		logger.String("job_type", jobType),
		logger.String("worker", worker),
		logger.String("process_instance_id", processInstanceID),
		logger.String("process_key", processKey),
		logger.String("state", state),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

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
		Limit:             0, // Use pagination instead
		PageSize:          pageSize,
		Page:              page,
		SortBy:            "created_at",
		SortOrder:         "DESC",
		IncludeVariables:  false,
	})
	if err != nil {
		logger.Error("Failed to list jobs", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	logger.Debug("Jobs listed", logger.Int("count", len(resp.Jobs)))

	fmt.Printf("Job List\n")
	fmt.Printf("========\n")

	// Print pagination info if multiple pages exist
	if resp.TotalPages > 1 {
		fmt.Printf("Page %d of %d (Total: %d jobs, Showing: %d)\n\n",
			resp.Page, resp.TotalPages, resp.TotalCount, len(resp.Jobs))
	} else {
		fmt.Printf("Found %d job(s):\n\n", resp.TotalCount)
	}

	printJobsTable(resp.Jobs, resp.TotalCount)

	// Show navigation hints for pagination
	if resp.TotalPages > 1 {
		fmt.Printf("\nNavigation:\n")

		// Previous page
		if resp.Page > 1 {
			prevPageCmd := fmt.Sprintf("atomd job list")
			if jobType != "" {
				prevPageCmd += fmt.Sprintf(" %s", jobType)
			}
			if worker != "" {
				prevPageCmd += fmt.Sprintf(" %s", worker)
			}
			prevPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page-1, resp.PageSize)
			fmt.Printf("Previous page: %s\n", prevPageCmd)
		}

		// Next page
		if resp.Page < resp.TotalPages {
			nextPageCmd := fmt.Sprintf("atomd job list")
			if jobType != "" {
				nextPageCmd += fmt.Sprintf(" %s", jobType)
			}
			if worker != "" {
				nextPageCmd += fmt.Sprintf(" %s", worker)
			}
			nextPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page+1, resp.PageSize)
			fmt.Printf("Next page: %s\n", nextPageCmd)
		}
	}

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
