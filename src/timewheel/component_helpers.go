/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"encoding/json"
	"fmt"
)

// Helper functions for core engine integration
// Вспомогательные функции для интеграции с core engine

// CreateScheduleTimerMessage creates JSON message for timer scheduling
// Создает JSON сообщение для планирования таймера
func CreateScheduleTimerMessage(request TimerRequest) (string, error) {
	message := struct {
		Type    string       `json:"type"`
		Request TimerRequest `json:"request"`
	}{
		Type:    "schedule_timer",
		Request: request,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schedule timer message: %w", err)
	}

	return string(jsonData), nil
}

// CreateCancelTimerMessage creates JSON message for timer cancellation
// Создает JSON сообщение для отмены таймера
func CreateCancelTimerMessage(timerID string) (string, error) {
	message := struct {
		Type    string `json:"type"`
		TimerID string `json:"timer_id"`
	}{
		Type:    "cancel_timer",
		TimerID: timerID,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cancel timer message: %w", err)
	}

	return string(jsonData), nil
}

// CreateGetStatsMessage creates JSON message for statistics request
// Создает JSON сообщение для запроса статистики
func CreateGetStatsMessage() (string, error) {
	message := struct {
		Type string `json:"type"`
	}{
		Type: "get_stats",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal get stats message: %w", err)
	}

	return string(jsonData), nil
}
