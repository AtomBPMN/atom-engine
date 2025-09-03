/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/src/storage"
)

// ProcessMessage processes JSON message from core engine
// Обрабатывает JSON сообщение от core engine
func (c *Component) ProcessMessage(ctx context.Context, messageJSON string) error {
	if !c.ready {
		return fmt.Errorf("component not ready")
	}

	// Parse message to determine type
	// Парсим сообщение для определения типа
	var baseMessage struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(messageJSON), &baseMessage); err != nil {
		return fmt.Errorf("failed to parse message type: %w", err)
	}

	switch baseMessage.Type {
	case "schedule_timer":
		return c.handleScheduleTimer(messageJSON)
	case "cancel_timer":
		return c.handleCancelTimer(messageJSON)
	case "get_stats":
		return c.handleGetStats()
	default:
		return fmt.Errorf("unknown message type: %s", baseMessage.Type)
	}
}

// handleScheduleTimer handles timer scheduling request
// Обрабатывает запрос планирования таймера
func (c *Component) handleScheduleTimer(messageJSON string) error {
	var message struct {
		Type    string       `json:"type"`
		Request TimerRequest `json:"request"`
	}

	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return fmt.Errorf("failed to parse schedule timer message: %w", err)
	}

	// Process timer synchronously to get the generated timer ID
	// Обрабатываем таймер синхронно чтобы получить сгенерированный ID
	timerID, err := c.manager.ScheduleTimerRequest(context.Background(), message.Request)
	if err != nil {
		return fmt.Errorf("failed to schedule timer: %w", err)
	}

	// Save timer to storage with the actual timer ID from manager
	// Сохраняем таймер в storage с реальным ID от manager
	// Skip saving for restored timers to preserve original ScheduledAt
	// Пропускаем сохранение для восстановленных таймеров чтобы сохранить оригинальный ScheduledAt
	if c.storage != nil && message.Request.RestoreTimerID == nil {
		timerRecord := c.timerRequestToRecord(&message.Request, timerID)
		if err := c.storage.SaveTimer(timerRecord); err != nil {
			return fmt.Errorf("failed to save timer to storage: %w", err)
		}
	}

	return nil
}

// handleCancelTimer handles timer cancellation request
// Обрабатывает запрос отмены таймера
func (c *Component) handleCancelTimer(messageJSON string) error {
	var message struct {
		Type    string `json:"type"`
		TimerID string `json:"timer_id"`
	}

	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return fmt.Errorf("failed to parse cancel timer message: %w", err)
	}

	// Cancel timer in timewheel
	err := c.manager.CancelTimer(message.TimerID)

	// Delete from storage if available
	// Удаляем из storage если доступен
	if c.storage != nil {
		if delErr := c.storage.DeleteTimer(message.TimerID); delErr != nil {
			// Log error but don't fail the operation
			// Логируем ошибку но не проваливаем операцию
			return fmt.Errorf("timer cancelled but failed to delete from storage: %w", delErr)
		}
	}

	return err
}

// handleGetStats handles statistics request
// Обрабатывает запрос статистики
func (c *Component) handleGetStats() error {
	stats := c.manager.GetStats()
	statsJSON, err := json.Marshal(struct {
		Type  string `json:"type"`
		Stats Stats  `json:"stats"`
	}{
		Type:  "stats_response",
		Stats: stats,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	select {
	case c.responseChannel <- string(statsJSON):
		return nil
	default:
		return fmt.Errorf("response channel is full")
	}
}

// timerRequestToRecord converts TimerRequest to storage.TimerRecord
// Конвертирует TimerRequest в storage.TimerRecord
func (c *Component) timerRequestToRecord(req *TimerRequest, timerID string) *storage.TimerRecord {
	now := time.Now()

	// Convert ProcessContext to map
	processContext := make(map[string]interface{})
	if req.ProcessContext != nil {
		processContext["process_key"] = req.ProcessContext.ProcessKey
		processContext["process_version"] = req.ProcessContext.ProcessVersion
		processContext["process_name"] = req.ProcessContext.ProcessName
		processContext["component_source"] = req.ProcessContext.ComponentSource
	}

	return &storage.TimerRecord{
		ID:                timerID,
		ElementID:         req.ElementID,
		TokenID:           req.TokenID,
		ProcessInstanceID: req.ProcessInstanceID,
		TimerType:         string(req.TimerType),
		ScheduledAt:       now, // Will be updated when actually scheduled
		TimeDate:          req.TimeDate,
		TimeDuration:      req.TimeDuration,
		TimeCycle:         req.TimeCycle,
		ProcessContext:    processContext,
		CreatedAt:         now,
		UpdatedAt:         now,
		State:             "SCHEDULED",
	}
}
