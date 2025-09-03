/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// TokenManager manages token lifecycle and operations
// Управляет жизненным циклом и операциями токенов
type TokenManager struct {
	storage         storage.Storage
	tokenOperations *TokenOperations
}

// NewTokenManager creates new token manager
// Создает новый менеджер токенов
func NewTokenManager(storage storage.Storage) *TokenManager {
	tm := &TokenManager{
		storage: storage,
	}

	// Initialize token operations
	tm.tokenOperations = NewTokenOperations(storage)

	return tm
}

// Init initializes token manager
// Инициализирует менеджер токенов
func (tm *TokenManager) Init() error {
	logger.Info("Initializing token manager")

	if tm.storage == nil {
		return fmt.Errorf("storage not provided")
	}

	logger.Info("Token manager initialized")
	return nil
}

// Start starts token manager
// Запускает менеджер токенов
func (tm *TokenManager) Start() error {
	logger.Info("Starting token manager")
	logger.Info("Token manager started")
	return nil
}

// Stop stops token manager
// Останавливает менеджер токенов
func (tm *TokenManager) Stop() error {
	logger.Info("Stopping token manager")
	logger.Info("Token manager stopped")
	return nil
}

// CreateToken creates new token
// Создает новый токен
func (tm *TokenManager) CreateToken(processInstanceID, processKey, elementID string) (*models.Token, error) {
	token := models.NewToken(processInstanceID, processKey, elementID)

	if err := tm.storage.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	logger.Info("Token created",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID))

	return token, nil
}

// CreateEventToken creates new event-based token
// Создает новый токен на основе события
func (tm *TokenManager) CreateEventToken(processInstanceID, processKey, elementID string) (*models.Token, error) {
	token := models.NewEventToken(processInstanceID, processKey, elementID)

	if err := tm.storage.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save event token: %w", err)
	}

	logger.Info("Event token created",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID))

	return token, nil
}

// CreateTimerToken creates new timer-based token
// Создает новый токен на основе таймера
func (tm *TokenManager) CreateTimerToken(processInstanceID, processKey, elementID string) (*models.Token, error) {
	token := models.NewTimerToken(processInstanceID, processKey, elementID)

	if err := tm.storage.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save timer token: %w", err)
	}

	logger.Info("Timer token created",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID))

	return token, nil
}

// GetToken gets token by ID
// Получает токен по ID
func (tm *TokenManager) GetToken(tokenID string) (*models.Token, error) {
	return tm.storage.LoadToken(tokenID)
}

// UpdateToken updates token in storage
// Обновляет токен в storage
func (tm *TokenManager) UpdateToken(token *models.Token) error {
	token.UpdatedAt = time.Now()
	return tm.storage.UpdateToken(token)
}

// DeleteToken deletes token from storage
// Удаляет токен из storage
func (tm *TokenManager) DeleteToken(tokenID string) error {
	return tm.storage.DeleteToken(tokenID)
}

// GetTokensByProcessInstance gets all tokens for process instance
// Получает все токены для экземпляра процесса
func (tm *TokenManager) GetTokensByProcessInstance(processInstanceID string) ([]*models.Token, error) {
	return tm.storage.LoadTokensByProcessInstance(processInstanceID)
}

// GetActiveTokens gets all active tokens
// Получает все активные токены
func (tm *TokenManager) GetAllActiveTokens() ([]*models.Token, error) {
	return tm.storage.LoadActiveTokens()
}

// GetActiveTokens gets active tokens for specific process instance
// Получает активные токены для конкретного экземпляра процесса
func (tm *TokenManager) GetActiveTokens(instanceID string) ([]*models.Token, error) {
	tokens, err := tm.storage.LoadTokensByProcessInstance(instanceID)
	if err != nil {
		return nil, err
	}

	var activeTokens []*models.Token
	for _, token := range tokens {
		if token.IsActive() || token.IsWaiting() {
			activeTokens = append(activeTokens, token)
		}
	}

	return activeTokens, nil
}

// GetTokensByState gets tokens by state
// Получает токены по состоянию
func (tm *TokenManager) GetTokensByState(state models.TokenState) ([]*models.Token, error) {
	return tm.storage.LoadTokensByState(state)
}

// MoveToken moves token to next element
// Перемещает токен к следующему элементу
func (tm *TokenManager) MoveToken(tokenID, nextElementID string) error {
	token, err := tm.storage.LoadToken(tokenID)
	if err != nil {
		return fmt.Errorf("failed to load token: %w", err)
	}

	token.MoveTo(nextElementID)

	if err := tm.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	logger.Info("Token moved",
		logger.String("token_id", tokenID),
		logger.String("from", token.PreviousElementID),
		logger.String("to", nextElementID))

	return nil
}

// SetTokenState sets token state
// Устанавливает состояние токена
func (tm *TokenManager) SetTokenState(tokenID string, state models.TokenState) error {
	token, err := tm.storage.LoadToken(tokenID)
	if err != nil {
		return fmt.Errorf("failed to load token: %w", err)
	}

	oldState := token.State
	token.SetState(state)

	if err := tm.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to update token: %w", err)
	}

	logger.Info("Token state changed",
		logger.String("token_id", tokenID),
		logger.String("from", string(oldState)),
		logger.String("to", string(state)))

	return nil
}

// Delegate methods to token operations
// Делегирующие методы к операциям с токенами

// CloneToken creates a clone of token for parallel execution
// Создает клон токена для параллельного выполнения
func (tm *TokenManager) CloneToken(tokenID string) (*models.Token, error) {
	return tm.tokenOperations.CloneToken(tokenID)
}

// MergeTokens merges multiple tokens back into one (for joining parallel flows)
// Объединяет множественные токены обратно в один (для соединения параллельных потоков)
func (tm *TokenManager) MergeTokens(tokenIDs []string, targetElementID string) (*models.Token, error) {
	return tm.tokenOperations.MergeTokens(tokenIDs, targetElementID)
}

// GetTokenStatistics gets token statistics
// Получает статистику токенов
func (tm *TokenManager) GetTokenStatistics() (map[string]int, error) {
	return tm.tokenOperations.GetTokenStatistics()
}

// ExecuteToken executes token (delegated to engine in practice)
// Выполняет токен (делегируется в engine на практике)
func (tm *TokenManager) ExecuteToken(token *models.Token) error {
	// This method will be implemented by the component's engine
	// but interface requires it for TokenManagerInterface
	// Этот метод будет реализован engine компонента
	// но интерфейс требует его для TokenManagerInterface
	return fmt.Errorf("ExecuteToken should be called through component engine")
}
