/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"container/list"
	"context"
	"sync"
	"time"

	"atom-engine/src/core/models"
)

// TimerRequest JSON message for scheduling timer
// JSON сообщение для планирования таймера
type TimerRequest struct {
	// Basic parameters
	// Основные параметры
	ElementID         string           `json:"element_id"`
	TokenID           string           `json:"token_id"`
	ProcessInstanceID string           `json:"process_instance_id"`
	TimerType         models.TimerType `json:"timer_type"`

	// Process context for callbacks
	// Контекст процесса для колбеков
	ProcessContext *models.TimerProcessContext `json:"process_context"`

	// BPMN timer definitions (only one should be set)
	// Определения BPMN таймеров (только одно должно быть установлено)
	TimeDate     *string `json:"time_date,omitempty"`     // "2025-12-31T23:59:59Z"
	TimeDuration *string `json:"time_duration,omitempty"` // "PT30S"
	TimeCycle    *string `json:"time_cycle,omitempty"`    // "R3/PT20S"

	// Boundary timer specific
	// Специфично для boundary таймеров
	AttachedToRef  *string `json:"attached_to_ref,omitempty"`
	CancelActivity *bool   `json:"cancel_activity,omitempty"`

	// Restoration specific - if set, use this ID instead of generating new one
	// Для восстановления - если установлен, используем этот ID вместо генерации нового
	RestoreTimerID *string `json:"restore_timer_id,omitempty"`

	// Restoration specific - if set, use this DueDate instead of calculating from time definitions
	// Для восстановления - если установлен, используем этот DueDate вместо расчета из определений времени
	RestoreDueDate *time.Time `json:"restore_due_date,omitempty"`

	// Base time for consistent calculation - if set, use this instead of time.Now()
	// Базовое время для консистентного расчета - если установлен, используем его вместо time.Now()
	BaseTime *time.Time `json:"base_time,omitempty"`
}

// TimerResponse JSON message when timer fires
// JSON сообщение при срабатывании таймера
type TimerResponse struct {
	TimerID           string                      `json:"timer_id"`
	ElementID         string                      `json:"element_id"`
	TokenID           string                      `json:"token_id"`
	ProcessInstanceID string                      `json:"process_instance_id"`
	TimerType         models.TimerType            `json:"timer_type"`
	ProcessContext    *models.TimerProcessContext `json:"process_context"`
	FiredAt           time.Time                   `json:"fired_at"`
	Variables         map[string]interface{}      `json:"variables,omitempty"`
}

// TimerHandler interface for processing timers
// Интерфейс для обработки таймеров
type TimerHandler interface {
	HandleTimer(ctx context.Context, timer *models.Timer) error
}

// TimerHandlerFunc adapter for functions
// Адаптер функций
type TimerHandlerFunc func(ctx context.Context, timer *models.Timer) error

func (f TimerHandlerFunc) HandleTimer(ctx context.Context, timer *models.Timer) error {
	return f(ctx, timer)
}

// TimerAnchor for O(1) timer removal
// Якорь для O(1) удаления таймера
type TimerAnchor struct {
	Level   int           `json:"level"`
	Slot    int           `json:"slot"`
	Element *list.Element `json:"-"` // Not serializable
}

// TimerEntry entry in timing wheel slot
// Запись в слоте timing wheel
type TimerEntry struct {
	Timer   *models.Timer
	Handler TimerHandler
	Anchor  *TimerAnchor
}

// TimingWheelLevel represents single level in hierarchical wheel
// Представляет один уровень в иерархическом колесе
type TimingWheelLevel struct {
	levelID     int
	tick        time.Duration
	size        int
	currentSlot int
	slots       []*list.List
	horizon     time.Duration
	mu          sync.RWMutex
}

// HierarchicalTimingWheel main timing wheel component
// Основной компонент timing wheel
type HierarchicalTimingWheel struct {
	levels     []*TimingWheelLevel
	timerIndex map[string]*TimerLocation // TimerID -> Location
	running    bool
	startTime  time.Time
	ticker     *time.Ticker
	stopChan   chan struct{}
	mu         sync.RWMutex
	wg         sync.WaitGroup

	// JSON communication channel with core
	// Канал JSON связи с core
	responseChannel chan<- string
}

// TimerLocation location of timer in timing wheel
// Местоположение таймера в timing wheel
type TimerLocation struct {
	Level    int `json:"level"`
	Slot     int `json:"slot"`
	Position int `json:"position"`
}

// LevelConfig configuration for timing wheel level
// Конфигурация уровня timing wheel
type LevelConfig struct {
	Tick string `json:"tick"` // Duration string like "1s", "1m"
	Size int    `json:"size"` // Number of slots
}

// Config configuration for timing wheel
// Конфигурация timing wheel
type Config struct {
	Levels []LevelConfig `json:"levels"`
}

// DefaultConfig default timing wheel configuration
// Конфигурация timing wheel по умолчанию
var DefaultConfig = Config{
	Levels: []LevelConfig{
		{Tick: "1s", Size: 60},     // L0: 1s * 60 = 1m
		{Tick: "1m", Size: 60},     // L1: 1m * 60 = 1h
		{Tick: "1h", Size: 24},     // L2: 1h * 24 = 1d
		{Tick: "24h", Size: 365},   // L3: 24h * 365 = 1y
		{Tick: "8760h", Size: 100}, // L4: 8760h * 100 = 100y
	},
}

// Stats statistics for timing wheel
// Статистика timing wheel
type Stats struct {
	LevelsCount  int           `json:"levels_count"`
	LevelStats   []LevelStats  `json:"level_stats"`
	TotalTimers  int           `json:"total_timers"`
	TotalHorizon time.Duration `json:"total_horizon"`
	Running      bool          `json:"running"`
	StartTime    time.Time     `json:"start_time"`
}

// LevelStats statistics for single level
// Статистика одного уровня
type LevelStats struct {
	LevelID     int           `json:"level_id"`
	Tick        time.Duration `json:"tick"`
	Size        int           `json:"size"`
	CurrentSlot int           `json:"current_slot"`
	TotalTimers int           `json:"total_timers"`
	Horizon     time.Duration `json:"horizon"`
}
