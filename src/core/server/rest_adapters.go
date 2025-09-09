/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"context"

	"atom-engine/src/core/models"
	"atom-engine/src/core/restapi/handlers"
)

// REST API adapter methods
// Методы-адаптеры для REST API

// GetProcessComponentForREST returns process component adapted for REST API
func (c *Core) GetProcessComponentForREST() handlers.ProcessComponentInterface {
	grpcComp := c.GetProcessComponent()
	if grpcComp == nil {
		return nil
	}
	return &processComponentRESTAdapter{grpcComp: grpcComp}
}

// GetTimewheelComponentForREST returns timewheel component adapted for REST API
func (c *Core) GetTimewheelComponentForREST() handlers.TimewheelComponentInterface {
	grpcComp := c.GetTimewheelComponent()
	return &timewheelComponentRESTAdapter{grpcComp: grpcComp}
}

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

// Adapter for process component
type processComponentRESTAdapter struct {
	grpcComp interface{} // grpc.ProcessComponentInterface
}

func (a *processComponentRESTAdapter) StartProcessInstance(processKey string, variables map[string]interface{}) (*handlers.ProcessInstanceResult, error) {
	// This would need to call the actual gRPC component method and convert the result
	// For now, return a mock result
	return &handlers.ProcessInstanceResult{
		InstanceID:      "mock-instance-id",
		ProcessID:       processKey,
		ProcessName:     "Mock Process",
		State:           "ACTIVE",
		CurrentActivity: "start",
		StartedAt:       0,
		UpdatedAt:       0,
		Variables:       variables,
	}, nil
}

func (a *processComponentRESTAdapter) GetProcessInstanceStatus(instanceID string) (*handlers.ProcessInstanceResult, error) {
	// Mock implementation
	return &handlers.ProcessInstanceResult{
		InstanceID:      instanceID,
		ProcessID:       "mock-process",
		ProcessName:     "Mock Process",
		State:           "ACTIVE",
		CurrentActivity: "running",
		StartedAt:       0,
		UpdatedAt:       0,
		Variables:       make(map[string]interface{}),
	}, nil
}

func (a *processComponentRESTAdapter) CancelProcessInstance(instanceID string, reason string) error {
	// Mock implementation
	return nil
}

func (a *processComponentRESTAdapter) ListProcessInstances(statusFilter string, processKeyFilter string, limit int) ([]*handlers.ProcessInstanceResult, error) {
	// Mock implementation
	return []*handlers.ProcessInstanceResult{}, nil
}

func (a *processComponentRESTAdapter) GetActiveTokens(instanceID string) ([]*models.Token, error) {
	// Mock implementation
	return []*models.Token{}, nil
}

func (a *processComponentRESTAdapter) GetTokensByProcessInstance(instanceID string) ([]*models.Token, error) {
	// Mock implementation - for trace endpoint
	return []*models.Token{}, nil
}

// Adapter for timewheel component
type timewheelComponentRESTAdapter struct {
	grpcComp interface{} // grpc.TimewheelComponentInterface
}

func (a *timewheelComponentRESTAdapter) ProcessMessage(ctx context.Context, messageJSON string) error {
	// This would need to call the actual gRPC timewheel component
	// For now, return success
	return nil
}

func (a *timewheelComponentRESTAdapter) GetResponseChannel() <-chan string {
	// This would need to return the actual gRPC component response channel
	// For now, return a mock channel
	ch := make(chan string, 1)
	ch <- `{"success": true}`
	return ch
}

func (a *timewheelComponentRESTAdapter) GetTimerInfo(timerID string) (level int, remainingSeconds int64, found bool) {
	// This would need to call the actual gRPC component
	// For now, return mock data
	return 0, 0, false
}
