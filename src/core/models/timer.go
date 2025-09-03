/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"crypto/rand"
	"math/big"
	"strings"
	"sync"
	"time"
)

// TimerType defines type of timer
// Определяет тип таймера
type TimerType string

const (
	TimerTypeStart    TimerType = "START"    // Start event timer
	TimerTypeBoundary TimerType = "BOUNDARY" // Boundary event timer
	TimerTypeEvent    TimerType = "EVENT"    // Intermediate timer event
)

// TimerState defines state of timer
// Определяет состояние таймера
type TimerState string

const (
	TimerStateScheduled TimerState = "SCHEDULED" // Timer is scheduled
	TimerStateFired     TimerState = "FIRED"     // Timer has fired
	TimerStateCanceled  TimerState = "CANCELED"  // Timer was canceled
)

// Timer represents a timer in the system
// Представляет таймер в системе
type Timer struct {
	ID                string                 `json:"id"`
	ElementID         string                 `json:"element_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ExecutionTokenID  string                 `json:"execution_token_id"`
	Type              TimerType              `json:"type"`
	State             TimerState             `json:"state"`
	DueDate           time.Time              `json:"due_date"`
	Variables         map[string]interface{} `json:"variables"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`

	// Process context for returning results
	// Контекст процесса для возврата результатов
	ProcessContext *TimerProcessContext `json:"process_context,omitempty"`
}

// TimerProcessContext contains process metadata for timer callbacks
// Содержит метаданные процесса для колбеков таймера
type TimerProcessContext struct {
	ProcessKey      string `json:"process_key"`      // BPMN process definition key
	ProcessVersion  int    `json:"process_version"`  // Process version
	ProcessName     string `json:"process_name"`     // Human readable name
	ComponentSource string `json:"component_source"` // Component that created timer
}

// Global instance name for ID generation
// Глобальное имя инстанса для генерации ID
var (
	instanceName string
	instanceMu   sync.RWMutex
)

// SetInstanceName sets instance name for ID generation
// Устанавливает имя инстанса для генерации ID
func SetInstanceName(name string) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	instanceName = name
}

// GetInstanceName returns current instance name
// Возвращает текущее имя инстанса
func GetInstanceName() string {
	instanceMu.RLock()
	defer instanceMu.RUnlock()
	return instanceName
}

// GenerateID generates unique ID with node prefix + NanoID
// Генерирует уникальный ID с префиксом узла + NanoID
func GenerateID() string {
	nodePrefix := getNodePrefix()
	nanoID := generateNanoID(18)
	return nodePrefix + "-" + nanoID
}

// getNodePrefix gets 4-character node prefix from instance name
// Получает 4-символьный префикс узла из имени инстанса
func getNodePrefix() string {
	instanceMu.RLock()
	name := instanceName
	instanceMu.RUnlock()

	if name == "" {
		name = "unkn" // fallback if not set
	}

	// Clean instance name and take first 4 chars
	// Очищаем имя инстанса и берем первые 4 символа
	cleaned := strings.ToLower(strings.ReplaceAll(name, ".", ""))
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "_", "")

	if len(cleaned) >= 4 {
		return cleaned[:4]
	}

	// Pad with zeros if name is too short
	// Дополняем нулями если имя слишком короткое
	for len(cleaned) < 4 {
		cleaned += "0"
	}

	return cleaned
}

// generateNanoID generates NanoID-like string with custom alphabet
// Генерирует NanoID-подобную строку с кастомным алфавитом
func generateNanoID(length int) string {
	// URL-safe alphabet like NanoID
	// URL-safe алфавит как в NanoID
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to time-based if crypto/rand fails
			// Фоллбэк на time-based если crypto/rand не работает
			result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			result[i] = charset[num.Int64()]
		}
	}

	return string(result)
}
