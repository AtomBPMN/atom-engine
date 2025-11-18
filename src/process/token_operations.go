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

// TokenOperations handles complex token operations
// Обрабатывает сложные операции с токенами
type TokenOperations struct {
	storage storage.Storage
}

// NewTokenOperations creates new token operations handler
// Создает новый обработчик операций с токенами
func NewTokenOperations(storage storage.Storage) *TokenOperations {
	return &TokenOperations{
		storage: storage,
	}
}

// CloneToken creates a clone of token for parallel execution
// Создает клон токена для параллельного выполнения
func (to *TokenOperations) CloneToken(tokenID string) (*models.Token, error) {
	originalToken, err := to.storage.LoadToken(tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to load original token: %w", err)
	}

	clonedToken := originalToken.Clone()

	if err := to.storage.SaveToken(clonedToken); err != nil {
		return nil, fmt.Errorf("failed to save cloned token: %w", err)
	}

	// Add child token to original
	originalToken.AddChildToken(clonedToken.TokenID)
	if err := to.storage.UpdateToken(originalToken); err != nil {
		logger.Error("Failed to update original token with child reference", logger.String("error", err.Error()))
	}

	logger.Info("Token cloned",
		logger.String("original_token_id", tokenID),
		logger.String("cloned_token_id", clonedToken.TokenID))

	return clonedToken, nil
}

// MergeTokens merges multiple tokens back into one (for joining parallel flows)
// Объединяет множественные токены обратно в один (для соединения параллельных потоков)
func (to *TokenOperations) MergeTokens(tokenIDs []string, targetElementID string) (*models.Token, error) {
	if len(tokenIDs) == 0 {
		return nil, fmt.Errorf("no tokens provided for merge")
	}

	// Load all tokens
	var tokens []*models.Token
	var processInstanceID, processKey string
	mergedVariables := make(map[string]interface{})

	for _, tokenID := range tokenIDs {
		token, err := to.storage.LoadToken(tokenID)
		if err != nil {
			return nil, fmt.Errorf("failed to load token %s: %w", tokenID, err)
		}

		tokens = append(tokens, token)

		// Set process info from first token
		if processInstanceID == "" {
			processInstanceID = token.ProcessInstanceID
			processKey = token.ProcessKey
		}

		// Merge variables from all tokens
		for key, value := range token.Variables {
			mergedVariables[key] = value
		}
	}

	// Create new merged token
	mergedToken := models.NewToken(processInstanceID, processKey, targetElementID)
	mergedToken.SetVariables(mergedVariables)

	if err := to.storage.SaveToken(mergedToken); err != nil {
		return nil, fmt.Errorf("failed to save merged token: %w", err)
	}

	// Complete all original tokens
	for _, token := range tokens {
		token.SetState(models.TokenStateCompleted)
		if err := to.storage.UpdateToken(token); err != nil {
			logger.Error("Failed to complete merged token",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
		}
	}

	logger.Info("Tokens merged",
		logger.Int("token_count", len(tokenIDs)),
		logger.String("merged_token_id", mergedToken.TokenID),
		logger.String("target_element", targetElementID))

	return mergedToken, nil
}

// GetTokenStatistics gets token statistics
// Получает статистику токенов
func (to *TokenOperations) GetTokenStatistics() (map[string]int, error) {
	stats := make(map[string]int)

	// Get tokens by each state
	states := []models.TokenState{
		models.TokenStateActive,
		models.TokenStateCompleted,
		models.TokenStateCanceled,
		models.TokenStateFailed,
		models.TokenStateWaiting,
	}

	for _, state := range states {
		tokens, err := to.storage.LoadTokensByState(state)
		if err != nil {
			return nil, fmt.Errorf("failed to load tokens for state %s: %w", state, err)
		}
		stats[string(state)] = len(tokens)
	}

	return stats, nil
}
