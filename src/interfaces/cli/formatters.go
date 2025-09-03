/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"atom-engine/proto/jobs/jobspb"
	"atom-engine/proto/messages/messagespb"
	"atom-engine/proto/parser/parserpb"
	"atom-engine/proto/process/processpb"
	"atom-engine/proto/timewheel/timewheelpb"
)

// formatBytes formats byte size to human readable format
// Форматирует размер в байтах в удобочитаемый формат
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats seconds into human readable duration
// Форматирует секунды в читаемую длительность
func formatDuration(seconds int64) string {
	if seconds == -1 {
		return "-1"
	}
	if seconds < 0 {
		return "N/A"
	}

	if seconds == 0 {
		return "0s"
	}

	years := seconds / (365 * 24 * 3600)
	seconds %= (365 * 24 * 3600)

	days := seconds / (24 * 3600)
	seconds %= (24 * 3600)

	hours := seconds / 3600
	seconds %= 3600

	minutes := seconds / 60
	seconds %= 60

	var parts []string

	if years > 0 {
		parts = append(parts, fmt.Sprintf("%dy", years))
	}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	// Show max 2 most significant parts
	// Показываем максимум 2 наиболее значимые части
	if len(parts) > 2 {
		parts = parts[:2]
	}

	return strings.Join(parts, " ")
}

// formatWheelLevel formats wheel level into readable string
// Форматирует уровень колеса в читаемую строку
func formatWheelLevel(level int32) string {
	if level == -1 {
		return "-1"
	}
	if level < 0 {
		return "N/A"
	}

	levelNames := []string{"Sec", "Min", "Hour", "Day", "Year"}
	if int(level) < len(levelNames) {
		return fmt.Sprintf("L%d:%s", level, levelNames[level])
	}

	return fmt.Sprintf("L%d", level)
}

// printTimersTable prints timers in a formatted table
// Выводит таймеры в форматированной таблице
func printTimersTable(timers []*timewheelpb.TimerInfo, totalCount int32) {
	if len(timers) == 0 {
		fmt.Println("No timers found.")
		return
	}

	// Sort by scheduled time (newest first)
	// Сортируем по времени запланированного выполнения (новые первыми)
	sort.Slice(timers, func(i, j int) bool {
		return timers[i].ScheduledAt > timers[j].ScheduledAt
	})

	// Print header
	fmt.Printf("Found %d timer(s):\n\n", totalCount)

	// Table headers - full width without truncation
	fmt.Printf("%-25s %-20s %-25s %-12s %-10s %-12s %-15s %-10s %-20s\n",
		"TIMER ID", "ELEMENT ID", "PROCESS INSTANCE", "TYPE", "STATUS", "DURATION", "REMAINING", "LEVEL", "SCHEDULED AT")
	fmt.Printf("%-25s %-20s %-25s %-12s %-10s %-12s %-15s %-10s %-20s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 20),
		strings.Repeat("-", 25),
		strings.Repeat("-", 12),
		strings.Repeat("-", 10),
		strings.Repeat("-", 12),
		strings.Repeat("-", 15),
		strings.Repeat("-", 10),
		strings.Repeat("-", 20))

	// Print timer rows - show full data without truncation
	for _, timer := range timers {
		// Handle zero timestamp for fired timers
		// Обрабатываем нулевой timestamp для сработавших таймеров
		var scheduledTime string
		if timer.ScheduledAt <= 0 {
			if timer.Status == "FIRED" {
				scheduledTime = "FIRED"
			} else {
				scheduledTime = "N/A"
			}
		} else {
			scheduledTime = time.Unix(timer.ScheduledAt, 0).Format("2006-01-02 15:04:05")
		}

		// Show full data without truncation
		timerID := timer.TimerId
		elementID := timer.ElementId
		processID := timer.ProcessInstanceId
		timerType := timer.TimerType
		status := timer.Status
		duration := timer.TimeDuration

		// Format remaining time and wheel level
		remaining := formatDuration(timer.RemainingSeconds)
		wheelLevel := formatWheelLevel(timer.WheelLevel)

		fmt.Printf("%-25s %-20s %-25s %-12s %-10s %-12s %-15s %-10s %-20s\n",
			timerID, elementID, processID, colorizeTimerType(timerType), colorizeStatus(status), duration, remaining, colorizeLevel(wheelLevel), scheduledTime)
	}

	fmt.Println()
}

// printSeparator prints a separator line
// Выводит разделительную линию
func printSeparator(length int) {
	fmt.Println(strings.Repeat("-", length))
}

// printBPMNProcessesTable prints BPMN processes in a formatted table sorted by creation time (newest first)
// Выводит BPMN процессы в форматированной таблице, отсортированной по времени создания (новые первыми)
func printBPMNProcessesTable(processes []*parserpb.BPMNProcessSummary, totalCount int32) {
	if len(processes) == 0 {
		fmt.Println("No BPMN processes found.")
		return
	}

	// Sort by creation time (newest first)
	// Сортируем по времени создания (новые первыми)
	sort.Slice(processes, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, processes[i].CreatedAt)
		timeJ, errJ := time.Parse(time.RFC3339, processes[j].CreatedAt)
		if errI != nil || errJ != nil {
			return processes[i].CreatedAt > processes[j].CreatedAt // Fallback to string comparison
		}
		return timeI.After(timeJ)
	})

	fmt.Printf("Found %d BPMN process(es):\n\n", totalCount)

	// Table headers with full width
	fmt.Printf("%-30s %-25s %-30s %-8s %-10s %-8s %-20s\n",
		"PROCESS KEY", "PROCESS ID", "NAME", "VERSION", "STATUS", "ELEMENTS", "CREATED")
	fmt.Printf("%-30s %-25s %-30s %-8s %-10s %-8s %-20s\n",
		strings.Repeat("-", 30),
		strings.Repeat("-", 25),
		strings.Repeat("-", 30),
		strings.Repeat("-", 8),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 20))

	// Print process rows
	for _, process := range processes {
		// Format created time
		var createdTime string
		if parsedTime, err := time.Parse(time.RFC3339, process.CreatedAt); err == nil {
			createdTime = parsedTime.Format("2006-01-02 15:04:05")
		} else {
			createdTime = process.CreatedAt
		}

		// Handle empty name
		name := process.ProcessName
		if name == "" {
			name = "<no name>"
		}

		fmt.Printf("%-30s %-25s %-30s %-8s %-10s %-8d %-20s\n",
			process.ProcessKey,
			process.ProcessId,
			name,
			process.Version,
			colorizeStatus(process.Status),
			process.TotalElements,
			createdTime)
	}

	fmt.Println()
}

// printJobsTable prints jobs in a formatted table sorted by creation time (newest first)
// Выводит задания в форматированной таблице, отсортированной по времени создания (новые первыми)
func printJobsTable(jobs []*jobspb.JobInfo, totalCount int32) {
	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return
	}

	// Sort by creation time (newest first)
	// Сортируем по времени создания (новые первыми)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt > jobs[j].CreatedAt
	})

	fmt.Printf("Found %d job(s):\n\n", totalCount)

	// Table headers
	fmt.Printf("%-25s %-15s %-15s %-8s %-12s %-25s %-20s %-20s\n",
		"JOB KEY", "TYPE", "WORKER", "RETRIES", "STATUS", "PROCESS INSTANCE", "ELEMENT ID", "CREATED")
	fmt.Printf("%-25s %-15s %-15s %-8s %-12s %-25s %-20s %-20s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 15),
		strings.Repeat("-", 15),
		strings.Repeat("-", 8),
		strings.Repeat("-", 12),
		strings.Repeat("-", 25),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20))

	// Print job rows
	for _, job := range jobs {
		// Format created time
		var createdTime string
		if job.CreatedAt > 0 {
			createdTime = time.Unix(job.CreatedAt, 0).Format("2006-01-02 15:04:05")
		} else {
			createdTime = "N/A"
		}

		// Format retries with color
		retriesInfo := colorizeRetries(job.Retries, job.MaxRetries)

		fmt.Printf("%-25s %-15s %-15s %-8s %-12s %-25s %-20s %-20s\n",
			job.Key,
			job.Type,
			job.Worker,
			retriesInfo,
			colorizeJobStatus(job.Status),
			job.ProcessInstanceKey,
			job.ElementId,
			createdTime)
	}

	fmt.Println()
}

// printProcessInstancesTable prints process instances in a formatted table sorted by creation time (newest first)
// Выводит экземпляры процессов в форматированной таблице, отсортированной по времени создания (новые первыми)
func printProcessInstancesTable(instances []*processpb.ProcessInstanceInfo, totalCount int32) {
	if len(instances) == 0 {
		fmt.Println("No process instances found.")
		return
	}

	// Sort by creation time (newest first)
	// Сортируем по времени создания (новые первыми)
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].StartedAt > instances[j].StartedAt
	})

	fmt.Printf("Found %d process instance(s):\n\n", totalCount)

	// Table headers
	fmt.Printf("%-25s %-25s %-8s %-12s %-20s %-20s %-20s\n",
		"INSTANCE ID", "PROCESS KEY", "VERSION", "STATUS", "STARTED", "UPDATED", "DURATION")
	fmt.Printf("%-25s %-25s %-8s %-12s %-20s %-20s %-20s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 25),
		strings.Repeat("-", 8),
		strings.Repeat("-", 12),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20))

	// Print instance rows
	for _, instance := range instances {
		// Format times
		var startTime, updatedTime, duration string
		if instance.StartedAt > 0 {
			startTime = time.Unix(instance.StartedAt, 0).Format("2006-01-02 15:04:05")
		} else {
			startTime = "N/A"
		}

		if instance.UpdatedAt > 0 {
			updatedTime = time.Unix(instance.UpdatedAt, 0).Format("2006-01-02 15:04:05")
			if instance.StartedAt > 0 {
				durationSec := instance.UpdatedAt - instance.StartedAt
				duration = formatDuration(durationSec)
			} else {
				duration = "N/A"
			}
		} else {
			updatedTime = "Running"
			if instance.StartedAt > 0 {
				durationSec := time.Now().Unix() - instance.StartedAt
				duration = formatDuration(durationSec)
			} else {
				duration = "N/A"
			}
		}

		fmt.Printf("%-25s %-25s %-8s %-12s %-20s %-20s %-20s\n",
			instance.InstanceId,
			instance.ProcessKey,
			"1",
			colorizeStatus(instance.Status),
			startTime,
			updatedTime,
			duration)
	}

	fmt.Println()
}

// printMessagesTable prints buffered messages in a formatted table sorted by published time (newest first)
// Выводит буферизованные сообщения в форматированной таблице, отсортированной по времени публикации (новые первыми)
func printMessagesTable(messages []*messagespb.BufferedMessage, totalCount int32) {
	if len(messages) == 0 {
		fmt.Println("No buffered messages found.")
		return
	}

	// Sort by published time (newest first)
	// Сортируем по времени публикации (новые первыми)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].PublishedAt > messages[j].PublishedAt
	})

	fmt.Printf("Found %d buffered message(s):\n\n", totalCount)

	// Table headers
	fmt.Printf("%-25s %-20s %-20s %-15s %-20s %-20s %-30s\n",
		"MESSAGE ID", "NAME", "CORRELATION KEY", "TENANT ID", "PUBLISHED", "EXPIRES", "REASON")
	fmt.Printf("%-25s %-20s %-20s %-15s %-20s %-20s %-30s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 15),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 30))

	// Print message rows
	for _, msg := range messages {
		// Format times
		var publishedTime, expiresTime string
		if msg.PublishedAt > 0 {
			publishedTime = time.Unix(msg.PublishedAt/1000, 0).Format("2006-01-02 15:04:05")
		} else {
			publishedTime = "N/A"
		}

		if msg.ExpiresAt > 0 {
			expiresTime = time.Unix(msg.ExpiresAt/1000, 0).Format("2006-01-02 15:04:05")
		} else {
			expiresTime = "Never"
		}

		// Handle empty fields
		correlationKey := msg.CorrelationKey
		if correlationKey == "" {
			correlationKey = "<none>"
		}

		tenantID := msg.TenantId
		if tenantID == "" {
			tenantID = "<default>"
		}

		reason := msg.Reason
		if reason == "" {
			reason = "<no reason>"
		}

		fmt.Printf("%-25s %-20s %-20s %-15s %-20s %-20s %-30s\n",
			msg.Id,
			msg.Name,
			correlationKey,
			tenantID,
			publishedTime,
			expiresTime,
			reason)
	}

	fmt.Println()
}

// printMessageSubscriptionsTable prints message subscriptions in a formatted table sorted by creation time (newest first)
// Выводит подписки на сообщения в форматированной таблице, отсортированной по времени создания (новые первыми)
func printMessageSubscriptionsTable(subscriptions []*messagespb.MessageSubscription, totalCount int32) {
	if len(subscriptions) == 0 {
		fmt.Println("No message subscriptions found.")
		return
	}

	// Sort by creation time (newest first)
	// Сортируем по времени создания (новые первыми)
	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].CreatedAt > subscriptions[j].CreatedAt
	})

	fmt.Printf("Found %d message subscription(s):\n\n", totalCount)

	// Table headers
	fmt.Printf("%-25s %-25s %-8s %-20s %-20s %-20s %-8s %-20s\n",
		"SUBSCRIPTION ID", "PROCESS DEF KEY", "VERSION", "MESSAGE NAME", "START EVENT ID", "CORRELATION KEY", "ACTIVE", "CREATED")
	fmt.Printf("%-25s %-25s %-8s %-20s %-20s %-20s %-8s %-20s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 25),
		strings.Repeat("-", 8),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 8),
		strings.Repeat("-", 20))

	// Print subscription rows
	for _, sub := range subscriptions {
		// Format created time
		var createdTime string
		if sub.CreatedAt > 0 {
			createdTime = time.Unix(sub.CreatedAt, 0).Format("2006-01-02 15:04:05")
		} else {
			createdTime = "N/A"
		}

		// Format active status with color
		activeStatus := colorizeBool(sub.IsActive)

		// Handle empty fields
		correlationKey := sub.CorrelationKey
		if correlationKey == "" {
			correlationKey = "<none>"
		}

		fmt.Printf("%-25s %-25s %-8d %-20s %-20s %-20s %-8s %-20s\n",
			sub.Id,
			sub.ProcessDefinitionKey,
			sub.ProcessVersion,
			sub.MessageName,
			sub.StartEventId,
			correlationKey,
			activeStatus,
			createdTime)
	}

	fmt.Println()
}
