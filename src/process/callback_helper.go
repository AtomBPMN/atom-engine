/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// CallbackHelper provides common callback processing functionality
// Предоставляет общую функциональность обработки callbacks
type CallbackHelper struct {
	storage       storage.Storage
	component     ComponentInterface
	tokenMovement *TokenMovement
}

// NewCallbackHelper creates new callback helper
// Создает новый callback helper
func NewCallbackHelper(storage storage.Storage, component ComponentInterface) *CallbackHelper {
	return &CallbackHelper{
		storage:       storage,
		component:     component,
		tokenMovement: NewTokenMovement(storage, component),
	}
}

// LoadAndValidateToken loads token and validates it's waiting for expected condition
// Загружает токен и проверяет что он ожидает ожидаемое условие
func (ch *CallbackHelper) LoadAndValidateToken(tokenID, expectedWaitingFor string) (*models.Token, error) {
	// Load the specific token
	token, err := ch.storage.LoadToken(tokenID)
	if err != nil {
		logger.Error("Failed to load token for callback",
			logger.String("token_id", tokenID),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load token %s: %w", tokenID, err)
	}

	// Check if token is waiting for expected condition
	if !token.IsWaiting() || token.WaitingFor != expectedWaitingFor {
		logger.Warn("Token is not waiting for expected condition",
			logger.String("token_id", tokenID),
			logger.String("token_state", string(token.State)),
			logger.String("token_waiting_for", token.WaitingFor),
			logger.String("expected_waiting_for", expectedWaitingFor))
		return nil, fmt.Errorf("token %s is not waiting for %s", tokenID, expectedWaitingFor)
	}

	logger.Info("Token confirmed waiting for condition",
		logger.String("token_id", tokenID),
		logger.String("waiting_for", expectedWaitingFor))

	return token, nil
}

// ProcessCallbackAndContinue processes callback and continues token execution
// Обрабатывает callback и продолжает выполнение токена
func (ch *CallbackHelper) ProcessCallbackAndContinue(token *models.Token, elementID string, variables map[string]interface{}) error {
	// Clear waiting state and merge variables if provided
	token.ClearWaitingFor()
	if variables != nil {
		token.MergeVariables(variables)
	}

	// Cancel boundary timers when token leaves activity (Service Task, etc.)
	// Отменяем boundary таймеры когда токен покидает activity (Service Task, и т.д.)
	if token.HasBoundaryTimers() {
		logger.Info("Canceling boundary timers for token leaving activity",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", elementID),
			logger.Int("timer_count", len(token.GetBoundaryTimers())))

		if err := ch.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
			logger.Error("Failed to cancel boundary timers for token leaving activity",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", elementID),
				logger.String("error", err.Error()))
			// Continue execution - boundary timer cancellation is not critical
		}
	}

	// Update token in storage first
	if err := ch.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	// Move token to next elements using existing logic
	logger.Info("DEBUG: About to move token to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID))

	if err := ch.tokenMovement.MoveTokenToNextElements(token, elementID); err != nil {
		logger.Error("DEBUG: Failed to move token to next elements",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", elementID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to move token to next elements: %w", err)
	}

	logger.Info("DEBUG: Successfully moved token to next elements",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID))

	logger.Info("Callback processed successfully - token execution continued",
		logger.String("element_id", elementID),
		logger.String("token_id", token.TokenID))

	return nil
}
