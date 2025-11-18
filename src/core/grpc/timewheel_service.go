/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"fmt"
	"sort"
	"time"

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
)

// timewheelServiceServer implements TimeWheel gRPC service
// Реализует TimeWheel gRPC сервис
type timewheelServiceServer struct {
	timewheelpb.UnimplementedTimeWheelServiceServer
	core CoreInterface
}

// AddTimer adds timer to time wheel
// Добавляет таймер в time wheel
func (s *timewheelServiceServer) AddTimer(
	ctx context.Context,
	req *timewheelpb.AddTimerRequest,
) (*timewheelpb.AddTimerResponse, error) {
	logger.Info("AddTimer gRPC request",
		logger.String("timer_id", req.TimerId),
		logger.Int64("delay_ms", req.DelayMs),
		logger.Bool("repeating", req.Repeating))

	component := s.core.GetTimewheelComponent()
	if component == nil {
		return &timewheelpb.AddTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: "timewheel component not initialized",
		}, nil
	}

	// Create timer request from gRPC request
	// Создаем запрос таймера из gRPC запроса
	timerReq := timewheel.TimerRequest{
		ElementID:         req.TimerId,
		TokenID:           fmt.Sprintf("cli-token-%s", req.TimerId),
		ProcessInstanceID: fmt.Sprintf("cli-process-%s", req.TimerId),
		TimerType:         models.TimerTypeStart, // CLI timer type
		ProcessContext: &models.TimerProcessContext{
			ProcessKey:      "cli-timer",
			ProcessVersion:  1,
			ProcessName:     "CLI Timer",
			ComponentSource: "logs", // Send response to logs
		},
	}

	// Set timer definition - prefer cycle for repeating timers, duration for one-time
	// Устанавливаем определение таймера - cycle для повторяющихся, duration для однократных
	if req.Interval != "" {
		// Repeating timer - use TimeCycle only
		// Повторяющийся таймер - используем только TimeCycle
		timerReq.TimeCycle = &req.Interval
	} else if req.Duration != "" {
		// One-time timer - use TimeDuration
		// Однократный таймер - используем TimeDuration
		timerReq.TimeDuration = &req.Duration
	} else if req.Repeating && req.IntervalMs > 0 {
		// Legacy repeating timer
		// Устаревший повторяющийся таймер
		cycle := fmt.Sprintf("R/PT%dS", req.IntervalMs/1000)
		timerReq.TimeCycle = &cycle
	} else if req.DelayMs > 0 {
		// Legacy one-time timer
		// Устаревший однократный таймер
		duration := fmt.Sprintf("PT%dS", req.DelayMs/1000)
		timerReq.TimeDuration = &duration
	}

	// Create JSON message for timer scheduling
	// Создаем JSON сообщение для планирования таймера
	messageJSON, err := timewheel.CreateScheduleTimerMessage(timerReq)
	if err != nil {
		logger.Error("Failed to create timer message", logger.String("error", err.Error()))
		return &timewheelpb.AddTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: fmt.Sprintf("failed to create timer message: %v", err),
		}, nil
	}

	// Process timer message
	// Обрабатываем сообщение таймера
	err = component.ProcessMessage(ctx, messageJSON)
	if err != nil {
		logger.Error("Failed to process timer message", logger.String("error", err.Error()))
		return &timewheelpb.AddTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: fmt.Sprintf("failed to process timer: %v", err),
		}, nil
	}

	// Use consistent base time for all calculations
	// Используем консистентное базовое время для всех расчетов
	baseTime := time.Now()
	timerReq.BaseTime = &baseTime

	// Calculate correct scheduled time for both legacy and ISO 8601 formats
	// Вычисляем правильное время срабатывания для legacy и ISO 8601 форматов
	var scheduledAt int64
	if req.DelayMs > 0 {
		// Legacy format with milliseconds
		// Устаревший формат с миллисекундами
		scheduledAt = baseTime.Add(time.Duration(req.DelayMs) * time.Millisecond).Unix()
	} else {
		// ISO 8601 format - try to parse and calculate from the timer request
		// ISO 8601 формат - пытаемся парсить и вычислить из запроса таймера
		if req.Duration != "" {
			// Parse ISO duration and calculate scheduled time
			// Парсим ISO длительность и вычисляем время срабатывания
			if parser := timewheel.NewISO8601DurationParser(); parser != nil {
				if duration, err := parser.ParseDuration(req.Duration); err == nil {
					scheduledAt = baseTime.Add(duration).Unix()
				}
			}
		} else if req.Interval != "" {
			// Parse ISO cycle and get first execution time
			// Парсим ISO цикл и получаем время первого выполнения
			if parser := timewheel.NewISO8601DurationParser(); parser != nil {
				if _, interval, err := parser.ParseRepeatingInterval(req.Interval); err == nil {
					scheduledAt = baseTime.Add(interval).Unix()
				}
			}
		}

		// Fallback to base time if parsing failed
		// Запасной вариант - базовое время если парсинг неудачен
		if scheduledAt == 0 {
			scheduledAt = baseTime.Unix()
		}
	}

	logger.Info("Timer scheduled successfully",
		logger.String("timer_id", req.TimerId),
		logger.Int64("scheduled_at", scheduledAt))

	return &timewheelpb.AddTimerResponse{
		TimerId:     req.TimerId,
		Success:     true,
		Message:     "Timer scheduled successfully",
		ScheduledAt: scheduledAt,
	}, nil
}

// RemoveTimer removes timer from time wheel
// Удаляет таймер из time wheel
func (s *timewheelServiceServer) RemoveTimer(
	ctx context.Context,
	req *timewheelpb.RemoveTimerRequest,
) (*timewheelpb.RemoveTimerResponse, error) {
	logger.Info("RemoveTimer gRPC request", logger.String("timer_id", req.TimerId))

	component := s.core.GetTimewheelComponent()
	if component == nil {
		return &timewheelpb.RemoveTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: "timewheel component not initialized",
		}, nil
	}

	// Create cancel timer message
	// Создаем сообщение отмены таймера
	messageJSON, err := timewheel.CreateCancelTimerMessage(req.TimerId)
	if err != nil {
		logger.Error("Failed to create cancel timer message", logger.String("error", err.Error()))
		return &timewheelpb.RemoveTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: fmt.Sprintf("failed to create cancel message: %v", err),
		}, nil
	}

	// Process cancel message
	// Обрабатываем сообщение отмены
	err = component.ProcessMessage(ctx, messageJSON)
	if err != nil {
		logger.Error("Failed to cancel timer", logger.String("error", err.Error()))
		return &timewheelpb.RemoveTimerResponse{
			TimerId: req.TimerId,
			Success: false,
			Message: fmt.Sprintf("failed to cancel timer: %v", err),
		}, nil
	}

	logger.Info("Timer cancelled successfully", logger.String("timer_id", req.TimerId))

	return &timewheelpb.RemoveTimerResponse{
		TimerId: req.TimerId,
		Success: true,
		Message: "Timer cancelled successfully",
	}, nil
}

// GetTimerStatus gets timer status
// Получает статус таймера
func (s *timewheelServiceServer) GetTimerStatus(
	ctx context.Context,
	req *timewheelpb.GetTimerStatusRequest,
) (*timewheelpb.GetTimerStatusResponse, error) {
	logger.Info("GetTimerStatus gRPC request", logger.String("timer_id", req.TimerId))

	component := s.core.GetTimewheelComponent()
	if component == nil {
		return &timewheelpb.GetTimerStatusResponse{
			TimerId: req.TimerId,
			Status:  "error",
		}, fmt.Errorf("timewheel component not initialized")
	}

	// Get real timer status from storage and timewheel
	// Получаем реальный статус таймера из storage и timewheel

	// First try to get timer info from timewheel (for active timers)
	_, remainingSeconds, foundInWheel := component.GetTimerInfo(req.TimerId)

	// Get timer record from storage for status and metadata
	storageInterface := s.core.GetStorage()
	storageComp, ok := storageInterface.(storage.Storage)
	if !ok {
		return &timewheelpb.GetTimerStatusResponse{
			TimerId: req.TimerId,
			Status:  "error",
		}, fmt.Errorf("storage component not available")
	}

	timerRecord, err := storageComp.LoadTimer(req.TimerId)
	if err != nil {
		// Timer not found in storage - it may not exist
		return &timewheelpb.GetTimerStatusResponse{
			TimerId: req.TimerId,
			Status:  "not_found",
		}, nil
	}

	// Determine status based on storage state and wheel presence
	var status string
	var remainingMs int64
	var isRepeating bool

	switch timerRecord.State {
	case "SCHEDULED":
		if foundInWheel {
			status = "pending"
			remainingMs = remainingSeconds * 1000 // Convert to milliseconds
		} else {
			// Scheduled in storage but not in wheel - likely system restart scenario
			status = "pending"
			remainingMs = 0
		}
	case "FIRED":
		status = "fired"
		remainingMs = 0
	case "CANCELLED":
		status = "cancelled"
		remainingMs = 0
	default:
		status = "unknown"
		remainingMs = 0
	}

	// Check if timer is repeating (has time_cycle)
	if timerRecord.TimeCycle != nil && *timerRecord.TimeCycle != "" {
		isRepeating = true
	}

	return &timewheelpb.GetTimerStatusResponse{
		TimerId:     req.TimerId,
		Status:      status,
		ScheduledAt: timerRecord.ScheduledAt.Unix(),
		RemainingMs: remainingMs,
		IsRepeating: isRepeating,
	}, nil
}

// GetTimeWheelStats gets time wheel statistics
// Получает статистику time wheel
func (s *timewheelServiceServer) GetTimeWheelStats(
	ctx context.Context,
	req *timewheelpb.GetTimeWheelStatsRequest,
) (*timewheelpb.GetTimeWheelStatsResponse, error) {
	logger.Info("GetTimeWheelStats gRPC request")

	return s.core.GetTimewheelStats()
}

// ListTimers lists all timers
// Возвращает список всех таймеров
func (s *timewheelServiceServer) ListTimers(
	ctx context.Context,
	req *timewheelpb.ListTimersRequest,
) (*timewheelpb.ListTimersResponse, error) {
	// Set defaults for pagination and sorting parameters
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}
	page := req.Page
	if page <= 0 {
		page = 1 // Default page
	}
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at" // Default sort field
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC" // Default sort order
	}

	logger.Info("ListTimers gRPC request",
		logger.String("status_filter", req.StatusFilter),
		logger.Int("limit", int(req.Limit)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder))

	// Get all timers first for sorting/pagination
	allTimersResponse, err := s.core.GetTimersList(req.StatusFilter, 0)
	if err != nil {
		logger.Error("Failed to get timers list", logger.String("error", err.Error()))
		return &timewheelpb.ListTimersResponse{
			Timers:     []*timewheelpb.TimerInfo{},
			TotalCount: 0,
		}, err
	}

	// Store total count before pagination
	totalCount := len(allTimersResponse.Timers)

	// Apply sorting
	sort.Slice(allTimersResponse.Timers, func(i, j int) bool {
		switch sortBy {
		case "created_at":
			if sortOrder == "ASC" {
				return allTimersResponse.Timers[i].CreatedAt < allTimersResponse.Timers[j].CreatedAt
			}
			return allTimersResponse.Timers[i].CreatedAt > allTimersResponse.Timers[j].CreatedAt
		case "scheduled_at":
			if sortOrder == "ASC" {
				return allTimersResponse.Timers[i].ScheduledAt < allTimersResponse.Timers[j].ScheduledAt
			}
			return allTimersResponse.Timers[i].ScheduledAt > allTimersResponse.Timers[j].ScheduledAt
		case "timer_id":
			if sortOrder == "ASC" {
				return allTimersResponse.Timers[i].TimerId < allTimersResponse.Timers[j].TimerId
			}
			return allTimersResponse.Timers[i].TimerId > allTimersResponse.Timers[j].TimerId
		default:
			// Default to created_at DESC
			return allTimersResponse.Timers[i].CreatedAt > allTimersResponse.Timers[j].CreatedAt
		}
	})

	// Calculate pagination
	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)
	offset := (int(page) - 1) * int(pageSize)

	// Apply pagination
	var paginatedTimers []*timewheelpb.TimerInfo
	if offset < len(allTimersResponse.Timers) {
		end := offset + int(pageSize)
		if end > len(allTimersResponse.Timers) {
			end = len(allTimersResponse.Timers)
		}
		paginatedTimers = allTimersResponse.Timers[offset:end]
	}

	// Use paginated timers for new pagination system or legacy limit for old system
	var finalTimers []*timewheelpb.TimerInfo
	if req.PageSize > 0 || (req.PageSize == 0 && req.Limit == 0) {
		// New pagination system (also default when no parameters specified)
		finalTimers = paginatedTimers
	} else if req.Limit > 0 && req.PageSize <= 0 {
		// Legacy limit system for backward compatibility
		if len(allTimersResponse.Timers) > int(req.Limit) {
			finalTimers = allTimersResponse.Timers[:req.Limit]
			totalCount = len(finalTimers)
			totalPages = 1
		} else {
			finalTimers = allTimersResponse.Timers
		}
	} else {
		finalTimers = paginatedTimers
	}

	logger.Info("Timers listed successfully",
		logger.Int("count", len(finalTimers)),
		logger.Int("total_count", totalCount),
		logger.Int("page", int(page)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("total_pages", totalPages))

	return &timewheelpb.ListTimersResponse{
		Timers:     finalTimers,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}, nil
}
