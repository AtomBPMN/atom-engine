/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import (
	"context"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// handleTimerFired handles timer when it fires
// Обрабатывает таймер когда он срабатывает
func (m *Manager) handleTimerFired(ctx context.Context, timer *models.Timer) error {
	// Timer response is already sent by the wheel via JSON channel
	// Ответ таймера уже отправлен колесом через JSON канал

	// Debug: log all timer variables for boundary timers
	// Отладка: логируем все переменные таймера для boundary таймеров
	if timer.Type == models.TimerTypeBoundary {
		logger.Debug("Boundary timer fired",
			logger.String("timer_id", timer.ID),
			logger.Any("variables", timer.Variables))
	}

	// Update timer status in storage to FIRED
	// Обновляем статус таймера в storage на FIRED
	if m.storage != nil {
		// Load existing timer record to preserve all fields
		// Загружаем существующую запись таймера чтобы сохранить все поля
		existingRecord, err := m.storage.LoadTimer(timer.ID)
		if err != nil {
			logger.Error("Failed to load existing timer for update",
				logger.String("timer_id", timer.ID),
				logger.String("error", err.Error()))
			return nil
		}

		// Update only the state and timestamp, preserve everything else
		// Обновляем только статус и timestamp, сохраняем все остальное
		existingRecord.State = "FIRED"
		existingRecord.UpdatedAt = time.Now()

		err = m.storage.SaveTimer(existingRecord)
		if err != nil {
			// Log error but don't fail the timer processing
			// Логируем ошибку но не прерываем обработку таймера
			logger.Error("Failed to update timer status in storage",
				logger.String("timer_id", timer.ID),
				logger.String("error", err.Error()))
		}
	}

	// Handle cycle timers for rescheduling
	// Обрабатываем циклические таймеры для переplanирования
	if cycleStr, ok := timer.Variables["time_cycle"].(string); ok {
		logger.Debug("Found time_cycle for timer - handling cycle",
			logger.String("timer_id", timer.ID),
			logger.String("time_cycle", cycleStr))
		return m.handleCycleTimer(timer, cycleStr)
	} else {
		if timer.Type == models.TimerTypeBoundary {
			logger.Debug("No time_cycle found for boundary timer",
				logger.String("timer_id", timer.ID))
		}
	}

	return nil
}

// handleCycleTimer handles cycle timer rescheduling
// Обрабатывает переplanирование циклического таймера
func (m *Manager) handleCycleTimer(timer *models.Timer, cycleStr string) error {
	repeatCount, ok := timer.Variables["repeat_count"].(int)
	if !ok {
		return nil // Not a cycle timer
	}

	currentIteration, ok := timer.Variables["current_iteration"].(int)
	if !ok {
		currentIteration = 1
	}

	// Check if we need to reschedule
	// Проверяем нужно ли переplanировать
	if repeatCount == -1 || currentIteration < repeatCount {
		// For BOUNDARY timers, check if parent scope is still active
		// Для BOUNDARY таймеров проверяем активен ли еще родительский scope
		if timer.Type == models.TimerTypeBoundary {
			if !m.isParentScopeActive(timer) {
				logger.Info("Parent scope ended for boundary timer - canceling repeats",
					logger.String("timer_id", timer.ID),
					logger.String("process_instance_id", timer.ProcessInstanceID))
				return nil // Parent scope ended, don't create more iterations
			}
		}

		// Parse interval
		// Парсим интервал
		_, interval, err := m.parser.ParseRepeatingInterval(cycleStr)
		if err != nil {
			return err
		}

		// Create new timer for next iteration
		// Создаем новый таймер для следующей итерации
		nextTimer := *timer
		nextTimer.ID = models.GenerateID()
		nextTimer.DueDate = time.Now().Add(interval)
		nextTimer.State = models.TimerStateScheduled
		nextTimer.CreatedAt = time.Now()
		nextTimer.UpdatedAt = time.Now()

		// Ensure Variables is initialized before assignment
		// Убеждаемся что Variables инициализирован перед присваиванием
		if nextTimer.Variables == nil {
			nextTimer.Variables = make(map[string]interface{})
			// Copy original variables if they existed
			if timer.Variables != nil {
				for k, v := range timer.Variables {
					nextTimer.Variables[k] = v
				}
			}
		}
		nextTimer.Variables["current_iteration"] = currentIteration + 1

		// Clear anchor from previous timer
		// Очищаем якорь от предыдущего таймера
		delete(nextTimer.Variables, "_anchor")

		// Save next timer to storage before scheduling
		// Сохраняем следующий таймер в storage перед планированием
		if m.storage != nil {
			timerRecord := &storage.TimerRecord{
				ID:                nextTimer.ID,
				ElementID:         nextTimer.ElementID,
				ProcessInstanceID: nextTimer.ProcessInstanceID,
				TokenID:           nextTimer.ExecutionTokenID,
				TimerType:         string(nextTimer.Type),
				State:             string(nextTimer.State),
				ScheduledAt:       nextTimer.CreatedAt,
				CreatedAt:         nextTimer.CreatedAt,
				UpdatedAt:         nextTimer.UpdatedAt,
				Variables:         nextTimer.Variables,
			}

			// Set timer definition
			if cycle, exists := nextTimer.Variables["time_cycle"]; exists {
				if cycleStr, ok := cycle.(string); ok {
					timerRecord.TimeCycle = &cycleStr
				}
			}

			// Set process context
			if nextTimer.ProcessContext != nil {
				timerRecord.ProcessContext = map[string]interface{}{
					"process_key":      nextTimer.ProcessContext.ProcessKey,
					"process_version":  nextTimer.ProcessContext.ProcessVersion,
					"process_name":     nextTimer.ProcessContext.ProcessName,
					"component_source": nextTimer.ProcessContext.ComponentSource,
				}
			}

			if err := m.storage.SaveTimer(timerRecord); err != nil {
				logger.Error("Failed to save repeat timer to storage",
					logger.String("timer_id", nextTimer.ID),
					logger.Int("iteration", currentIteration+1),
					logger.String("error", err.Error()))
			} else {
				logger.Debug("Repeat timer saved to storage",
					logger.String("timer_id", nextTimer.ID),
					logger.Int("iteration", currentIteration+1))
			}
		}

		// Schedule next iteration
		// Планируем следующую итерацию
		handler := TimerHandlerFunc(m.handleTimerFired)
		return m.wheel.AddTimer(&nextTimer, handler)
	}

	return nil
}

// isParentScopeActive checks if parent scope is still active for boundary timer
// Проверяет активен ли родительский scope для boundary таймера
func (m *Manager) isParentScopeActive(timer *models.Timer) bool {
	logger.Debug("Checking parent scope activity for boundary timer",
		logger.String("timer_id", timer.ID),
		logger.String("execution_token_id", timer.ExecutionTokenID),
		logger.String("process_instance_id", timer.ProcessInstanceID))

	// Strategy 1: Check execution token if available
	// Стратегия 1: Проверяем execution token если доступен
	if timer.ExecutionTokenID != "" && m.storage != nil {
		if token, err := m.storage.LoadToken(timer.ExecutionTokenID); err == nil {
			// For boundary timers, both ACTIVE and WAITING tokens indicate active scope
			// Для boundary таймеров и ACTIVE и WAITING токены указывают на активный scope
			isActive := token.State == models.TokenStateActive || token.State == models.TokenStateWaiting
			logger.Debug("Boundary timer parent scope check via execution token",
				logger.String("timer_id", timer.ID),
				logger.String("token_id", timer.ExecutionTokenID),
				logger.String("token_state", string(token.State)),
				logger.Bool("is_active", isActive))
			return isActive
		} else {
			logger.Warn("Failed to load execution token for boundary timer",
				logger.String("timer_id", timer.ID),
				logger.String("token_id", timer.ExecutionTokenID),
				logger.String("error", err.Error()))
		}
	}

	// Strategy 2: Check attached activity via process instance tokens
	// Стратегия 2: Проверяем привязанную активность через токены экземпляра процесса
	if attachedToRef, exists := timer.Variables["attached_to_ref"]; exists && m.storage != nil {
		if attachedElementID, ok := attachedToRef.(string); ok && attachedElementID != "" {
			return m.checkAttachedActivityStatus(timer.ProcessInstanceID, attachedElementID, timer.ID)
		}
	}

	// Strategy 3: Fallback to process instance activity check
	// Стратегия 3: Возврат к проверке активности экземпляра процесса
	if timer.ProcessInstanceID != "" && m.storage != nil {
		activeTokens, err := m.storage.LoadTokensByProcessInstance(timer.ProcessInstanceID)
		if err == nil {
			// If there are any active or waiting tokens in the process instance, assume scope is active
			// Если есть активные или ожидающие токены в экземпляре процесса, считаем scope активным
			for _, token := range activeTokens {
				// For boundary timers, both ACTIVE and WAITING tokens indicate active scope
				// Для boundary таймеров и ACTIVE и WAITING токены указывают на активный scope
				if token.State == models.TokenStateActive || token.State == models.TokenStateWaiting {
					logger.Debug("Found active/waiting tokens in process instance - scope considered active",
						logger.String("timer_id", timer.ID),
						logger.String("process_instance_id", timer.ProcessInstanceID),
						logger.String("token_state", string(token.State)),
						logger.Int("active_tokens_count", len(activeTokens)))
					return true
				}
			}
			logger.Debug("No active/waiting tokens found in process instance - scope considered inactive",
				logger.String("timer_id", timer.ID),
				logger.String("process_instance_id", timer.ProcessInstanceID))
			return false
		} else {
			logger.Error("Failed to load process instance tokens for boundary timer scope check",
				logger.String("timer_id", timer.ID),
				logger.String("process_instance_id", timer.ProcessInstanceID),
				logger.String("error", err.Error()))
		}
	}

	// Final fallback: assume inactive to prevent runaway timers
	// Финальный возврат: считаем неактивным чтобы предотвратить бесконтрольные таймеры
	logger.Warn("Unable to determine parent scope activity - assuming inactive to prevent runaway timers",
		logger.String("timer_id", timer.ID))
	return false
}

// checkAttachedActivityStatus checks if the attached activity is still active
// Проверяет активность привязанной активности
func (m *Manager) checkAttachedActivityStatus(processInstanceID, attachedElementID, timerID string) bool {
	tokens, err := m.storage.LoadTokensByProcessInstance(processInstanceID)
	if err != nil {
		logger.Error("Failed to load tokens for attached activity check",
			logger.String("timer_id", timerID),
			logger.String("process_instance_id", processInstanceID),
			logger.String("attached_element_id", attachedElementID),
			logger.String("error", err.Error()))
		return false
	}

	// Look for active or waiting tokens on the attached activity
	// Ищем активные или ожидающие токены на привязанной активности
	for _, token := range tokens {
		// For boundary timers, both ACTIVE and WAITING tokens indicate active scope
		// Для boundary таймеров и ACTIVE и WAITING токены указывают на активный scope
		if (token.State == models.TokenStateActive || token.State == models.TokenStateWaiting) &&
			token.CurrentElementID == attachedElementID {
			logger.Debug("Found active/waiting token on attached activity",
				logger.String("timer_id", timerID),
				logger.String("attached_element_id", attachedElementID),
				logger.String("token_id", token.TokenID),
				logger.String("token_state", string(token.State)))
			return true
		}
	}

	logger.Debug("No active/waiting tokens found on attached activity",
		logger.String("timer_id", timerID),
		logger.String("attached_element_id", attachedElementID),
		logger.Int("total_tokens_checked", len(tokens)))
	return false
}
