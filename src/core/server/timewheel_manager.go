/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/proto/timewheel/timewheelpb"
	"atom-engine/src/core/grpc"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
	"atom-engine/src/timewheel"
)

// GetTimewheelComponent returns timewheel component for gRPC
// Возвращает timewheel компонент для gRPC
func (c *Core) GetTimewheelComponent() grpc.TimewheelComponentInterface {
	return c.timewheelComp
}

// GetTimewheelComponentInterface returns timewheel component as interface for process component
// Возвращает timewheel компонент как interface для процессного компонента
func (c *Core) GetTimewheelComponentInterface() interface{} {
	return c.timewheelComp
}

// GetTimewheelStats returns timewheel statistics for gRPC
// Возвращает статистику timewheel для gRPC
func (c *Core) GetTimewheelStats() (*timewheelpb.GetTimeWheelStatsResponse, error) {
	if c.timewheelComp == nil {
		return &timewheelpb.GetTimeWheelStatsResponse{
			TotalTimers:   0,
			PendingTimers: 0,
			CurrentTick:   0,
		}, nil
	}

	// Calculate real statistics from storage
	// Вычисляем реальную статистику из storage
	var totalTimers, pendingTimers, firedTimers, cancelledTimers int32
	timerTypes := make(map[string]int32)

	if c.storage != nil {
		timers, err := c.storage.LoadAllTimers()
		if err != nil {
			logger.Error("Failed to load timers for stats", logger.String("error", err.Error()))
		} else {
			totalTimers = int32(len(timers))
			for _, timer := range timers {
				switch timer.State {
				case "SCHEDULED", "ACTIVE":
					pendingTimers++
				case "FIRED":
					firedTimers++
				case "CANCELLED":
					cancelledTimers++
				}
				timerTypes[timer.TimerType]++
			}
		}
	}

	return &timewheelpb.GetTimeWheelStatsResponse{
		TotalTimers:     totalTimers,
		PendingTimers:   pendingTimers,
		FiredTimers:     firedTimers,
		CancelledTimers: cancelledTimers,
		CurrentTick:     time.Now().Unix(),
		SlotsCount:      60, // Default first level slots
		TimerTypes:      timerTypes,
	}, nil
}

// GetTimersList returns list of timers for gRPC
// Возвращает список таймеров для gRPC
func (c *Core) GetTimersList(statusFilter string, limit int32) (*timewheelpb.ListTimersResponse, error) {
	// Use the internal implementation that actually loads timers from storage
	// Используем внутреннюю реализацию которая загружает таймеры из storage
	return c.GetTimersListInternal(statusFilter, limit)
}

// GetTimersListInternal returns list of timers for internal use
func (c *Core) GetTimersListInternal(statusFilter string, limit int32) (*timewheelpb.ListTimersResponse, error) {
	if c.storage == nil {
		return &timewheelpb.ListTimersResponse{
			Timers:     []*timewheelpb.TimerInfo{},
			TotalCount: 0,
		}, nil
	}

	// Load all timers from storage
	// Загружаем все таймеры из storage
	timers, err := c.storage.LoadAllTimers()
	if err != nil {
		logger.Error("Failed to load timers from storage", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load timers: %w", err)
	}

	logger.Debug("GetTimersListInternal called",
		logger.String("status_filter", statusFilter),
		logger.Int("limit", int(limit)),
		logger.Int("loaded_timers_count", len(timers)))

	var filteredTimers []*timewheelpb.TimerInfo

	for _, timer := range timers {
		// Apply status filter if specified
		// Применяем фильтр по статусу если указан
		if statusFilter != "" && timer.State != statusFilter {
			continue
		}

		// Get timer info from timewheel if available
		// Получаем информацию о таймере из timewheel если доступно
		var wheelLevel int32 = -1
		var remainingSeconds int64 = -1

		if c.timewheelComp != nil {
			level, remaining, found := c.timewheelComp.GetTimerInfo(timer.ID)
			if found {
				wheelLevel = int32(level)
				remainingSeconds = remaining
			}
		}

		// Calculate actual due time for display
		// Вычисляем фактическое время срабатывания для отображения
		scheduledAt := timer.ScheduledAt.Unix()
		if timer.State == "SCHEDULED" {
			// For scheduled timers, calculate due time from timer definition
			// Для запланированных таймеров вычисляем время срабатывания из определения
			if dueTime, err := c.calculateTimerDueTime(timer); err == nil {
				scheduledAt = dueTime.Unix()
			}
		} else if timer.State == "FIRED" {
			// For fired timers, calculate what was the original due time
			// Для сработавших таймеров вычисляем какое было оригинальное время срабатывания
			if dueTime, err := c.calculateTimerDueTime(timer); err == nil {
				scheduledAt = dueTime.Unix()
			}
		}

		timerInfo := &timewheelpb.TimerInfo{
			TimerId:           timer.ID,
			ElementId:         timer.ElementID,
			ProcessInstanceId: timer.ProcessInstanceID,
			TimerType:         timer.TimerType,
			Status:            timer.State,
			ScheduledAt:       scheduledAt,
			CreatedAt:         timer.CreatedAt.Unix(),
			TimeDuration:      getStringValue(timer.TimeDuration),
			TimeCycle:         getStringValue(timer.TimeCycle),
			RemainingSeconds:  remainingSeconds,
			WheelLevel:        wheelLevel,
		}

		filteredTimers = append(filteredTimers, timerInfo)

		// Apply limit if specified
		// Применяем лимит если указан
		if limit > 0 && int32(len(filteredTimers)) >= limit {
			break
		}
	}

	return &timewheelpb.ListTimersResponse{
		Timers:     filteredTimers,
		TotalCount: int32(len(filteredTimers)),
	}, nil
}

// getStringValue safely gets string value from pointer
// Безопасно получает строковое значение из указателя
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// calculateTimerDueTime calculates timer due time from timer definition
// Вычисляет время срабатывания таймера из определения таймера
func (c *Core) calculateTimerDueTime(timer *storage.TimerRecord) (time.Time, error) {
	if c.timewheelComp == nil {
		return time.Time{}, fmt.Errorf("timewheel component not available")
	}

	// Use ScheduledAt as base time for calculation
	// Используем ScheduledAt как базовое время для расчета
	baseTime := timer.ScheduledAt

	if timer.TimeDate != nil {
		// Absolute date timer - parse the date string
		// Абсолютный таймер даты - парсим строку даты
		parser := timewheel.NewISO8601DurationParser()
		return parser.ParseDate(*timer.TimeDate)
	} else if timer.TimeDuration != nil {
		// Duration-based timer
		// Таймер на основе длительности
		parser := timewheel.NewISO8601DurationParser()
		duration, err := parser.ParseDuration(*timer.TimeDuration)
		if err != nil {
			return time.Time{}, err
		}
		return baseTime.Add(duration), nil
	} else if timer.TimeCycle != nil {
		// Cycle-based timer - get first execution time
		// Циклический таймер - получаем время первого выполнения
		parser := timewheel.NewISO8601DurationParser()
		_, interval, err := parser.ParseRepeatingInterval(*timer.TimeCycle)
		if err != nil {
			return time.Time{}, err
		}
		return baseTime.Add(interval), nil
	}

	return time.Time{}, fmt.Errorf("no timer definition found")
}

// processTimewheelResponses processes timewheel responses in background
// Обрабатывает ответы timewheel в фоне
func (c *Core) processTimewheelResponses() {
	if c.timewheelComp == nil {
		return
	}

	responseChannel := c.timewheelComp.GetResponseChannel()

	for {
		select {
		case response := <-responseChannel:
			c.handleTimewheelResponse(response)
		}
	}
}

// handleTimewheelResponse handles single timewheel response
// Обрабатывает один ответ timewheel
func (c *Core) handleTimewheelResponse(response string) {
	// Parse timer response for readable logging
	// Парсим ответ таймера для читаемого логирования
	var timerResp struct {
		TimerID           string `json:"timer_id"`
		ElementID         string `json:"element_id"`
		TokenID           string `json:"token_id"`
		ProcessInstanceID string `json:"process_instance_id"`
		FiredAt           string `json:"fired_at"`
	}

	if err := json.Unmarshal([]byte(response), &timerResp); err == nil {
		logger.Info("CLI Timer Callback",
			logger.String("element_id", timerResp.ElementID),
			logger.String("timer_id", timerResp.TimerID),
			logger.String("token_id", timerResp.TokenID),
			logger.String("process_instance_id", timerResp.ProcessInstanceID),
			logger.String("fired_at", timerResp.FiredAt))

		// Forward timer callback to process component with token ID
		// Передаем timer callback в process component с token ID
		if c.processComp != nil {
			if err := c.processComp.HandleTimerCallback(timerResp.TimerID, timerResp.ElementID, timerResp.TokenID); err != nil {
				logger.Error("Failed to handle timer callback in process component",
					logger.String("timer_id", timerResp.TimerID),
					logger.String("element_id", timerResp.ElementID),
					logger.String("token_id", timerResp.TokenID),
					logger.String("error", err.Error()))
			} else {
				logger.Info("Timer callback processed successfully",
					logger.String("timer_id", timerResp.TimerID),
					logger.String("element_id", timerResp.ElementID),
					logger.String("token_id", timerResp.TokenID))
			}
		}
	}

	// Also log full JSON for debugging
	// Также логируем полный JSON для отладки
	logger.Debug("Timer fired", logger.String("response", response))

	// Log timer response to storage
	// Логируем ответ таймера в storage
	err := c.storage.LogSystemEvent(models.EventTypeReady, models.StatusSuccess, fmt.Sprintf("Timer fired: %s", response))
	if err != nil {
		logger.Warn("Failed to log timer response to storage", logger.String("error", err.Error()))
	}
}
