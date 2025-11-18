/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"

	"atom-engine/src/core/restapi/handlers"
	"atom-engine/src/jobs"
	"atom-engine/src/parser"
)

// REST API adapter methods
// Методы-адаптеры для REST API

// GetTimewheelStatsForREST returns timewheel stats adapted for REST API
func (c *Core) GetTimewheelStatsForREST() (*handlers.TimewheelStatsResponse, error) {
	grpcStats, err := c.GetTimewheelStats()
	if err != nil {
		return nil, err
	}

	return &handlers.TimewheelStatsResponse{
		TotalTimers:     grpcStats.TotalTimers,
		PendingTimers:   grpcStats.PendingTimers,
		FiredTimers:     grpcStats.FiredTimers,
		CancelledTimers: grpcStats.CancelledTimers,
		CurrentTick:     grpcStats.CurrentTick,
		SlotsCount:      grpcStats.SlotsCount,
		TimerTypes:      grpcStats.TimerTypes,
	}, nil
}

// GetTimersListForREST returns timers list adapted for REST API
func (c *Core) GetTimersListForREST(statusFilter string, limit int32) (*handlers.TimersListResponse, error) {
	grpcList, err := c.GetTimersList(statusFilter, limit)
	if err != nil {
		return nil, err
	}

	// Convert gRPC timer info to REST timer info
	restTimers := make([]handlers.TimerInfo, len(grpcList.Timers))
	for i, grpcTimer := range grpcList.Timers {
		restTimers[i] = handlers.TimerInfo{
			TimerID:           grpcTimer.TimerId,
			ElementID:         grpcTimer.ElementId,
			ProcessInstanceID: grpcTimer.ProcessInstanceId,
			TimerType:         grpcTimer.TimerType,
			Status:            grpcTimer.Status,
			ScheduledAt:       grpcTimer.ScheduledAt,
			CreatedAt:         grpcTimer.CreatedAt,
			TimeDuration:      grpcTimer.TimeDuration,
			TimeCycle:         grpcTimer.TimeCycle,
			RemainingSeconds:  grpcTimer.RemainingSeconds,
			WheelLevel:        grpcTimer.WheelLevel,
		}
	}

	return &handlers.TimersListResponse{
		Timers:     restTimers,
		TotalCount: grpcList.TotalCount,
	}, nil
}

// GetProcessInfoForREST returns complete process information adapted for REST API
func (c *Core) GetProcessInfoForREST(instanceID string) (map[string]interface{}, error) {
	// Get process status
	processComp := c.GetProcessComponent()
	if processComp == nil {
		return nil, fmt.Errorf("process component not available")
	}

	processStatus, err := processComp.GetProcessInstanceStatus(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	// Get BPMN Process Key by finding process definition
	bpmnProcessKey := ""
	processKey := processStatus.ProcessKey

	// If ProcessKey is empty, use ProcessID to find the latest version
	if processStatus.ProcessID != "" {
		processID := processStatus.ProcessID

		// Get parser component to find BPMN process
		if parserComp := c.GetParserComponent(); parserComp != nil {
			// Try to cast to the proper interface
			if typedParserComp, ok := parserComp.(interface {
				ListBPMNProcesses(limit int) ([]*parser.ProcessInfo, error)
			}); ok {
				// Get all BPMN processes and find matching one
				if processes, err := typedParserComp.ListBPMNProcesses(100); err == nil {
					// Find the latest version for this process ID
					var latestProcess *parser.ProcessInfo
					for _, process := range processes {
						if process.ProcessID == processID {
							if latestProcess == nil || process.ProcessVersion > latestProcess.ProcessVersion {
								latestProcess = process
							}
						}
					}
					if latestProcess != nil {
						bpmnProcessKey = latestProcess.BPMNID
						processKey = fmt.Sprintf("%s:v%d", latestProcess.ProcessID, latestProcess.ProcessVersion)
					}
				}
			}
		}
	}

	// Build complete process info including external services
	processInfo := map[string]interface{}{
		"instance_id":       processStatus.InstanceID,
		"process_key":       processKey,
		"bpmn_process_key":  bpmnProcessKey,
		"process_name":      processStatus.ProcessName,
		"state":             processStatus.State,
		"created_at":        processStatus.CreatedAt,
		"updated_at":        processStatus.UpdatedAt,
		"variables":         processStatus.Variables,
		"external_services": c.buildExternalServicesForREST(instanceID, processStatus.ProcessKey),
	}

	return processInfo, nil
}

// buildExternalServicesForREST builds external services info for REST API
func (c *Core) buildExternalServicesForREST(instanceID, processKey string) map[string]interface{} {
	externalServices := map[string]interface{}{
		"timers":                []map[string]interface{}{},
		"jobs":                  []map[string]interface{}{},
		"message_subscriptions": []map[string]interface{}{},
		"buffered_messages":     []map[string]interface{}{},
		"incidents":             []map[string]interface{}{},
	}

	// Get timers using existing method
	if timersResp, err := c.GetTimersList("", 1000); err == nil {
		var timers []map[string]interface{}
		for _, timer := range timersResp.Timers {
			if timer.ProcessInstanceId == instanceID {
				timerInfo := map[string]interface{}{
					"timer_id":          timer.TimerId,
					"element_id":        timer.ElementId,
					"timer_type":        timer.TimerType,
					"status":            timer.Status,
					"scheduled_at":      timer.ScheduledAt,
					"remaining_seconds": timer.RemainingSeconds,
					"time_duration":     timer.TimeDuration,
					"time_cycle":        timer.TimeCycle,
				}
				timers = append(timers, timerInfo)
			}
		}
		externalServices["timers"] = timers
	}

	// Get jobs using jobs component - cast to jobs.Component
	if jobsComp, ok := c.GetJobsComponent().(*jobs.Component); jobsComp != nil && ok {
		if jobInfos, _, err := jobsComp.ListJobs("", "", instanceID, "", 1000, 0); err == nil {
			var jobsList []map[string]interface{}
			for _, jobInfo := range jobInfos {
				jobMap := map[string]interface{}{
					"key":           jobInfo.Key,
					"type":          jobInfo.Type,
					"worker":        jobInfo.Worker,
					"element_id":    "", // Not available in JobInfo
					"status":        jobInfo.Status,
					"retries":       jobInfo.Retries,
					"created_at":    jobInfo.CreatedAt,
					"error_message": jobInfo.ErrorMessage,
				}
				jobsList = append(jobsList, jobMap)
			}
			externalServices["jobs"] = jobsList
		}
	}

	return externalServices
}
