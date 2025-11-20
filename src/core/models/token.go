/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"encoding/json"
	"time"
)

// TokenState defines state of execution token
// Определяет состояние токена выполнения
type TokenState string

const (
	TokenStateActive    TokenState = "ACTIVE"
	TokenStateCompleted TokenState = "COMPLETED"
	TokenStateCanceled  TokenState = "CANCELED"
	TokenStateFailed    TokenState = "FAILED"
	TokenStateWaiting   TokenState = "WAITING"
)

// TokenType defines type of execution token
// Определяет тип токена выполнения
type TokenType string

const (
	TokenTypeExecution TokenType = "EXECUTION" // Normal execution token
	TokenTypeEvent     TokenType = "EVENT"     // Event-based token
	TokenTypeTimer     TokenType = "TIMER"     // Timer-based token
)

// ExecutionContext keys
// Ключи контекста выполнения
const (
	ContextKeyTimerCallback = "timer_callback" // Indicates token execution from timer callback
)

// Token represents execution token moving through process
// Представляет токен выполнения движущийся по процессу
type Token struct {
	TokenID           string                 `json:"token_id"`
	ProcessInstanceID string                 `json:"process_instance_id"`
	ProcessKey        string                 `json:"process_key"`
	CurrentElementID  string                 `json:"current_element_id"`
	PreviousElementID string                 `json:"previous_element_id,omitempty"`
	State             TokenState             `json:"state"`
	Type              TokenType              `json:"type"`
	Variables map[string]interface{} `json:"variables"` // Token-specific variables
	// What token is waiting for (job, message, timer)
	WaitingFor string `json:"waiting_for,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`

	// Execution context
	// Контекст выполнения
	ExecutionContext map[string]interface{} `json:"execution_context,omitempty"`

	// Parent token for sub-processes or parallel flows
	// Родительский токен для подпроцессов или параллельных потоков
	ParentTokenID string `json:"parent_token_id,omitempty"`

	// SubProcess element ID this token is executing inside
	// ID элемента SubProcess внутри которого выполняется токен
	SubProcessID string `json:"subprocess_id,omitempty"`

	// Child tokens for parallel execution
	// Дочерние токены для параллельного выполнения
	ChildTokenIDs []string `json:"child_token_ids,omitempty"`

	// Boundary timer IDs attached to this token
	// ID boundary таймеров прикрепленных к данному токену
	BoundaryTimerIDs []string `json:"boundary_timer_ids,omitempty"`
}

// NewToken creates new execution token
// Создает новый токен выполнения
func NewToken(processInstanceID, processKey, elementID string) *Token {
	now := time.Now()
	return &Token{
		TokenID:           GenerateID(),
		ProcessInstanceID: processInstanceID,
		ProcessKey:        processKey,
		CurrentElementID:  elementID,
		State:             TokenStateActive,
		Type:              TokenTypeExecution,
		Variables:         make(map[string]interface{}),
		ExecutionContext:  make(map[string]interface{}),
		ChildTokenIDs:     make([]string, 0),
		BoundaryTimerIDs:  make([]string, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// NewEventToken creates new event-based token
// Создает новый токен на основе события
func NewEventToken(processInstanceID, processKey, elementID string) *Token {
	token := NewToken(processInstanceID, processKey, elementID)
	token.Type = TokenTypeEvent
	return token
}

// NewTimerToken creates new timer-based token
// Создает новый токен на основе таймера
func NewTimerToken(processInstanceID, processKey, elementID string) *Token {
	token := NewToken(processInstanceID, processKey, elementID)
	token.Type = TokenTypeTimer
	return token
}

// ToJSON converts token to JSON
// Конвертирует токен в JSON
func (t *Token) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON creates token from JSON
// Создает токен из JSON
func (t *Token) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// MoveTo moves token to next element
// Перемещает токен к следующему элементу
func (t *Token) MoveTo(elementID string) {
	t.PreviousElementID = t.CurrentElementID
	t.CurrentElementID = elementID
	t.UpdatedAt = time.Now()
}

// SetState sets token state
// Устанавливает состояние токена
func (t *Token) SetState(state TokenState) {
	t.State = state
	t.UpdatedAt = time.Now()

	if state == TokenStateCompleted ||
		state == TokenStateCanceled ||
		state == TokenStateFailed {
		now := time.Now()
		t.CompletedAt = &now
	}
}

// SetVariable sets token variable
// Устанавливает переменную токена
func (t *Token) SetVariable(key string, value interface{}) {
	if t.Variables == nil {
		t.Variables = make(map[string]interface{})
	}
	t.Variables[key] = value
	t.UpdatedAt = time.Now()
}

// GetVariable gets token variable
// Получает переменную токена
func (t *Token) GetVariable(key string) (interface{}, bool) {
	value, exists := t.Variables[key]
	return value, exists
}

// SetVariables sets multiple token variables
// Устанавливает множественные переменные токена
func (t *Token) SetVariables(variables map[string]interface{}) {
	if t.Variables == nil {
		t.Variables = make(map[string]interface{})
	}
	for key, value := range variables {
		t.Variables[key] = value
	}
	t.UpdatedAt = time.Now()
}

// MergeVariables merges variables from another source
// Объединяет переменные из другого источника
func (t *Token) MergeVariables(variables map[string]interface{}) {
	if t.Variables == nil {
		t.Variables = make(map[string]interface{})
	}
	for key, value := range variables {
		t.Variables[key] = value
	}
	t.UpdatedAt = time.Now()
}

// SetExecutionContext sets execution context field
// Устанавливает поле контекста выполнения
func (t *Token) SetExecutionContext(key string, value interface{}) {
	if t.ExecutionContext == nil {
		t.ExecutionContext = make(map[string]interface{})
	}
	t.ExecutionContext[key] = value
	t.UpdatedAt = time.Now()
}

// GetExecutionContext gets execution context field
// Получает поле контекста выполнения
func (t *Token) GetExecutionContext(key string) (interface{}, bool) {
	if t.ExecutionContext == nil {
		return nil, false
	}
	value, exists := t.ExecutionContext[key]
	return value, exists
}

// SetWaitingFor sets what token is waiting for
// Устанавливает чего ожидает токен
func (t *Token) SetWaitingFor(waitingFor string) {
	t.WaitingFor = waitingFor
	t.State = TokenStateWaiting
	t.UpdatedAt = time.Now()
}

// ClearWaitingFor clears waiting state
// Очищает состояние ожидания
func (t *Token) ClearWaitingFor() {
	t.WaitingFor = ""
	if t.State == TokenStateWaiting {
		t.State = TokenStateActive
	}
	t.UpdatedAt = time.Now()
}

// AddChildToken adds child token ID
// Добавляет ID дочернего токена
func (t *Token) AddChildToken(childTokenID string) {
	t.ChildTokenIDs = append(t.ChildTokenIDs, childTokenID)
	t.UpdatedAt = time.Now()
}

// RemoveChildToken removes child token ID
// Удаляет ID дочернего токена
func (t *Token) RemoveChildToken(childTokenID string) {
	for i, id := range t.ChildTokenIDs {
		if id == childTokenID {
			t.ChildTokenIDs = append(t.ChildTokenIDs[:i], t.ChildTokenIDs[i+1:]...)
			break
		}
	}
	t.UpdatedAt = time.Now()
}

// HasChildTokens checks if token has child tokens
// Проверяет есть ли у токена дочерние токены
func (t *Token) HasChildTokens() bool {
	return len(t.ChildTokenIDs) > 0
}

// AddBoundaryTimer adds boundary timer ID to token
// Добавляет ID boundary таймера к токену
func (t *Token) AddBoundaryTimer(timerID string) {
	t.BoundaryTimerIDs = append(t.BoundaryTimerIDs, timerID)
	t.UpdatedAt = time.Now()
}

// RemoveBoundaryTimer removes boundary timer ID from token
// Удаляет ID boundary таймера из токена
func (t *Token) RemoveBoundaryTimer(timerID string) {
	for i, id := range t.BoundaryTimerIDs {
		if id == timerID {
			t.BoundaryTimerIDs = append(t.BoundaryTimerIDs[:i], t.BoundaryTimerIDs[i+1:]...)
			break
		}
	}
	t.UpdatedAt = time.Now()
}

// HasBoundaryTimers checks if token has boundary timers
// Проверяет есть ли у токена boundary таймеры
func (t *Token) HasBoundaryTimers() bool {
	return len(t.BoundaryTimerIDs) > 0
}

// GetBoundaryTimers returns boundary timer IDs
// Возвращает ID boundary таймеров
func (t *Token) GetBoundaryTimers() []string {
	return append([]string{}, t.BoundaryTimerIDs...) // Return copy
}

// IsActive checks if token is active
// Проверяет активен ли токен
func (t *Token) IsActive() bool {
	return t.State == TokenStateActive
}

// IsWaiting checks if token is waiting
// Проверяет ожидает ли токен
func (t *Token) IsWaiting() bool {
	return t.State == TokenStateWaiting
}

// IsCompleted checks if token is completed
// Проверяет завершен ли токен
func (t *Token) IsCompleted() bool {
	return t.State == TokenStateCompleted ||
		t.State == TokenStateCanceled ||
		t.State == TokenStateFailed
}

// SetTimerCallback marks token as coming from timer callback
// Отмечает токен как пришедший от timer callback
func (t *Token) SetTimerCallback() {
	t.SetExecutionContext(ContextKeyTimerCallback, true)
}

// ClearTimerCallback clears timer callback flag
// Очищает флаг timer callback
func (t *Token) ClearTimerCallback() {
	delete(t.ExecutionContext, ContextKeyTimerCallback)
	t.UpdatedAt = time.Now()
}

// IsFromTimerCallback checks if token execution is from timer callback
// Проверяет пришло ли выполнение токена от timer callback
func (t *Token) IsFromTimerCallback() bool {
	value, exists := t.GetExecutionContext(ContextKeyTimerCallback)
	if !exists {
		return false
	}
	isCallback, ok := value.(bool)
	return ok && isCallback
}

// Clone creates a copy of token for parallel execution
// Создает копию токена для параллельного выполнения
func (t *Token) Clone() *Token {
	now := time.Now()
	clone := &Token{
		TokenID:           GenerateID(),
		ProcessInstanceID: t.ProcessInstanceID,
		ProcessKey:        t.ProcessKey,
		CurrentElementID:  t.CurrentElementID,
		PreviousElementID: t.PreviousElementID,
		State:             t.State,
		Type:              t.Type,
		Variables:         make(map[string]interface{}),
		ExecutionContext:  make(map[string]interface{}),
		ParentTokenID:     t.TokenID, // Set original as parent
		ChildTokenIDs:     make([]string, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Copy variables
	for key, value := range t.Variables {
		clone.Variables[key] = value
	}

	// Copy execution context
	for key, value := range t.ExecutionContext {
		clone.ExecutionContext[key] = value
	}

	return clone
}
