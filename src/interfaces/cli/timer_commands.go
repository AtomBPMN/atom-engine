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

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/logger"
)

// TimerAdd adds new timer via gRPC
// Добавляет новый таймер через gRPC
func (d *DaemonCommand) TimerAdd() error {
	logger.Debug("Adding new timer")

	if len(os.Args) < 5 {
		logger.Error("Invalid timer add arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd timer add <id> <duration_or_cycle>")
	}

	timerID := os.Args[3]
	durationOrCycle := os.Args[4]

	logger.Debug("Timer add request",
		logger.String("timer_id", timerID),
		logger.String("duration_or_cycle", durationOrCycle))

	// Determine if argument is duration or cycle
	// Определяем является ли аргумент duration или cycle
	var duration, cycle string
	if strings.HasPrefix(strings.ToUpper(durationOrCycle), "R") {
		// It's a repeating cycle like R3/PT10S or r3/pt10s
		// Это повторяющийся цикл типа R3/PT10S или r3/pt10s
		cycle = durationOrCycle
		logger.Debug("Detected repeating cycle", logger.String("cycle", cycle))
	} else {
		// It's a one-time duration like PT5S or pt5s
		// Это однократная длительность типа PT5S или pt5s
		duration = durationOrCycle
		logger.Debug("Detected one-time duration", logger.String("duration", duration))
	}

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for timer add", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running: %w", err)
	}
	defer conn.Close()

	client := timewheelpb.NewTimeWheelServiceClient(conn)

	// Create request with ISO 8601 duration or cycle
	req := &timewheelpb.AddTimerRequest{
		TimerId:      timerID,
		Duration:     duration,
		CallbackData: fmt.Sprintf("CLI timer %s", timerID),
		Repeating:    cycle != "",
		Interval:     cycle,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.AddTimer(ctx, req)
	if err != nil {
		logger.Error("Failed to add timer via gRPC",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to add timer: %w", err)
	}

	if resp.Success {
		logger.Info("Timer added successfully",
			logger.String("timer_id", resp.TimerId),
			logger.Int64("scheduled_at", resp.ScheduledAt))
		fmt.Printf("Timer '%s' added successfully\n", resp.TimerId)
		fmt.Printf("Scheduled at: %s\n", time.Unix(resp.ScheduledAt, 0).Format("2006-01-02 15:04:05"))
		if cycle != "" {
			fmt.Printf("Repeating cycle: %s\n", cycle)
		}
	} else {
		logger.Warn("Timer add failed",
			logger.String("timer_id", timerID),
			logger.String("message", resp.Message))
		fmt.Printf("Failed to add timer: %s\n", resp.Message)
	}

	return nil
}

// TimerRemove removes timer via gRPC
// Удаляет таймер через gRPC
func (d *DaemonCommand) TimerRemove() error {
	logger.Debug("Removing timer")

	if len(os.Args) < 4 {
		logger.Error("Invalid timer remove arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd timer remove <id>")
	}

	timerID := os.Args[3]
	logger.Debug("Timer remove request", logger.String("timer_id", timerID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for timer remove", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running: %w", err)
	}
	defer conn.Close()

	client := timewheelpb.NewTimeWheelServiceClient(conn)

	req := &timewheelpb.RemoveTimerRequest{
		TimerId: timerID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RemoveTimer(ctx, req)
	if err != nil {
		logger.Error("Failed to remove timer via gRPC",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to remove timer: %w", err)
	}

	if resp.Success {
		logger.Info("Timer removed successfully", logger.String("timer_id", resp.TimerId))
		fmt.Printf("Timer '%s' removed successfully\n", resp.TimerId)
	} else {
		logger.Warn("Timer remove failed",
			logger.String("timer_id", timerID),
			logger.String("message", resp.Message))
		fmt.Printf("Failed to remove timer: %s\n", resp.Message)
	}

	return nil
}

// TimerStatus gets timer status via gRPC
// Получает статус таймера через gRPC
func (d *DaemonCommand) TimerStatus() error {
	logger.Debug("Getting timer status")

	if len(os.Args) < 4 {
		logger.Error("Invalid timer status arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd timer status <id>")
	}

	timerID := os.Args[3]
	logger.Debug("Timer status request", logger.String("timer_id", timerID))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for timer status", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running: %w", err)
	}
	defer conn.Close()

	client := timewheelpb.NewTimeWheelServiceClient(conn)

	req := &timewheelpb.GetTimerStatusRequest{
		TimerId: timerID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetTimerStatus(ctx, req)
	if err != nil {
		logger.Error("Failed to get timer status via gRPC",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to get timer status: %w", err)
	}

	logger.Debug("Timer status retrieved",
		logger.String("timer_id", resp.TimerId),
		logger.String("status", resp.Status))

	fmt.Printf("Timer ID: %s\n", resp.TimerId)
	fmt.Printf("Status: %s\n", colorizeStatus(resp.Status))
	fmt.Printf("Scheduled at: %s\n", time.Unix(resp.ScheduledAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Remaining: %d ms\n", resp.RemainingMs)
	fmt.Printf("Repeating: %t\n", resp.IsRepeating)

	return nil
}

// TimerStats gets timewheel statistics via gRPC
// Получает статистику timewheel через gRPC
func (d *DaemonCommand) TimerStats() error {
	logger.Debug("Getting timewheel statistics")

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for timer stats", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running: %w", err)
	}
	defer conn.Close()

	client := timewheelpb.NewTimeWheelServiceClient(conn)

	req := &timewheelpb.GetTimeWheelStatsRequest{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetTimeWheelStats(ctx, req)
	if err != nil {
		logger.Error("Failed to get timewheel stats via gRPC", logger.String("error", err.Error()))
		return fmt.Errorf("failed to get timewheel stats: %w", err)
	}

	logger.Debug("Timewheel stats retrieved",
		logger.Int("total_timers", int(resp.TotalTimers)),
		logger.Int("pending_timers", int(resp.PendingTimers)))

	fmt.Printf("TimeWheel Statistics:\n")
	fmt.Printf("Total timers: %d\n", resp.TotalTimers)
	fmt.Printf("Pending timers: %d\n", resp.PendingTimers)
	fmt.Printf("Fired timers: %d\n", resp.FiredTimers)
	fmt.Printf("Cancelled timers: %d\n", resp.CancelledTimers)
	fmt.Printf("Current tick: %d\n", resp.CurrentTick)
	fmt.Printf("Slots count: %d\n", resp.SlotsCount)

	if len(resp.TimerTypes) > 0 {
		fmt.Println("Timer types:")
		for timerType, count := range resp.TimerTypes {
			fmt.Printf("  %s: %d\n", timerType, count)
		}
	}

	return nil
}

// TimerList lists all timers via gRPC
// Выводит список всех таймеров через gRPC
func (d *DaemonCommand) TimerList() error {
	logger.Debug("Listing timers")

	// Parse arguments for filtering and pagination
	var statusFilter string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd timer list"

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
			}
		}
	}

	logger.Debug("Timer list request",
		logger.String("status_filter", statusFilter),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for timer list", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running: %w", err)
	}
	defer conn.Close()

	client := timewheelpb.NewTimeWheelServiceClient(conn)

	req := &timewheelpb.ListTimersRequest{
		StatusFilter: statusFilter,
		Limit:        0, // Use pagination instead
		PageSize:     pageSize,
		Page:         page,
		SortBy:       "created_at",
		SortOrder:    "DESC",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListTimers(ctx, req)
	if err != nil {
		logger.Error("Failed to list timers via gRPC", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list timers: %w", err)
	}

	logger.Debug("Timer list retrieved",
		logger.Int("timers_count", len(resp.Timers)),
		logger.Int("total_count", int(resp.TotalCount)))

	// Print pagination info if multiple pages exist
	if resp.TotalPages > 1 {
		fmt.Printf("Timer List - Page %d of %d (Total: %d timers, Showing: %d)\n\n",
			resp.Page, resp.TotalPages, resp.TotalCount, len(resp.Timers))
	} else {
		fmt.Printf("Timer List - Found %d timer(s):\n\n", resp.TotalCount)
	}

	printTimersTable(resp.Timers, resp.TotalCount)

	// Show navigation hints for pagination
	if resp.TotalPages > 1 {
		fmt.Printf("\nNavigation:\n")

		// Previous page
		if resp.Page > 1 {
			prevPageCmd := fmt.Sprintf("atomd timer list")
			if statusFilter != "" {
				prevPageCmd += fmt.Sprintf(" %s", statusFilter)
			}
			prevPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page-1, resp.PageSize)
			fmt.Printf("Previous page: %s\n", prevPageCmd)
		}

		// Next page
		if resp.Page < resp.TotalPages {
			nextPageCmd := fmt.Sprintf("atomd timer list")
			if statusFilter != "" {
				nextPageCmd += fmt.Sprintf(" %s", statusFilter)
			}
			nextPageCmd += fmt.Sprintf(" --page %d --page-size %d", resp.Page+1, resp.PageSize)
			fmt.Printf("Next page: %s\n", nextPageCmd)
		}
	}

	return nil
}
