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
	"time"

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
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
func (s *timewheelServiceServer) AddTimer(ctx context.Context, req *timewheelpb.AddTimerRequest) (*timewheelpb.AddTimerResponse, error) {
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
func (s *timewheelServiceServer) RemoveTimer(ctx context.Context, req *timewheelpb.RemoveTimerRequest) (*timewheelpb.RemoveTimerResponse, error) {
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
func (s *timewheelServiceServer) GetTimerStatus(ctx context.Context, req *timewheelpb.GetTimerStatusRequest) (*timewheelpb.GetTimerStatusResponse, error) {
	logger.Info("GetTimerStatus gRPC request", logger.String("timer_id", req.TimerId))

	component := s.core.GetTimewheelComponent()
	if component == nil {
		return &timewheelpb.GetTimerStatusResponse{
			TimerId: req.TimerId,
			Status:  "error",
		}, fmt.Errorf("timewheel component not initialized")
	}

	// For now return basic status - can be enhanced with actual timer lookup
	// Пока возвращаем базовый статус - можно улучшить реальным поиском таймера
	return &timewheelpb.GetTimerStatusResponse{
		TimerId:     req.TimerId,
		Status:      "pending", // pending, fired, cancelled
		ScheduledAt: time.Now().Unix(),
		RemainingMs: 0,
		IsRepeating: false,
	}, nil
}

// GetTimeWheelStats gets time wheel statistics
// Получает статистику time wheel
func (s *timewheelServiceServer) GetTimeWheelStats(ctx context.Context, req *timewheelpb.GetTimeWheelStatsRequest) (*timewheelpb.GetTimeWheelStatsResponse, error) {
	logger.Info("GetTimeWheelStats gRPC request")

	return s.core.GetTimewheelStats()
}

// ListTimers lists all timers
// Возвращает список всех таймеров
func (s *timewheelServiceServer) ListTimers(ctx context.Context, req *timewheelpb.ListTimersRequest) (*timewheelpb.ListTimersResponse, error) {
	logger.Info("ListTimers gRPC request",
		logger.String("status_filter", req.StatusFilter),
		logger.Int("limit", int(req.Limit)))

	return s.core.GetTimersList(req.StatusFilter, req.Limit)
}
