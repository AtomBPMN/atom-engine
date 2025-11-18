/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"sync"

	"atom-engine/src/core/logger"
)

// ErrorBoundarySubscription represents an error boundary event subscription
// Подписка на граничное событие ошибки
type ErrorBoundarySubscription struct {
	TokenID       string `json:"token_id"`
	ElementID     string `json:"element_id"`        // Error boundary event ID
	AttachedToRef string `json:"attached_to_ref"`   // Service task ID that error is attached to
	ErrorRef      string `json:"error_ref"`         // Error definition reference
	ErrorCode     string `json:"error_code"`        // Error code to match (e.g., "404")
	ErrorName     string `json:"error_name"`        // Error name
	CancelActivity bool  `json:"cancel_activity"`   // Whether this is interrupting
	OutgoingFlows []string `json:"outgoing_flows"`  // Sequence flows to activate on error
}

// ErrorBoundaryRegistry manages error boundary event subscriptions
// Реестр для управления подписками на граничные события ошибок
type ErrorBoundaryRegistry struct {
	mutex         sync.RWMutex
	subscriptions map[string][]*ErrorBoundarySubscription // Key: tokenID, Value: list of subscriptions
}

// NewErrorBoundaryRegistry creates new error boundary registry
// Создает новый реестр граничных событий ошибок
func NewErrorBoundaryRegistry() *ErrorBoundaryRegistry {
	return &ErrorBoundaryRegistry{
		subscriptions: make(map[string][]*ErrorBoundarySubscription),
	}
}

// RegisterErrorBoundary registers error boundary event for token
// Регистрирует граничное событие ошибки для токена
func (ebr *ErrorBoundaryRegistry) RegisterErrorBoundary(subscription *ErrorBoundarySubscription) {
	ebr.mutex.Lock()
	defer ebr.mutex.Unlock()

	logger.Info("Registering error boundary subscription",
		logger.String("token_id", subscription.TokenID),
		logger.String("element_id", subscription.ElementID),
		logger.String("attached_to", subscription.AttachedToRef),
		logger.String("error_code", subscription.ErrorCode))

	if ebr.subscriptions[subscription.TokenID] == nil {
		ebr.subscriptions[subscription.TokenID] = make([]*ErrorBoundarySubscription, 0)
	}

	ebr.subscriptions[subscription.TokenID] = append(ebr.subscriptions[subscription.TokenID], subscription)
}

// GetErrorBoundariesForToken gets all error boundary subscriptions for token
// Получает все подписки на граничные события ошибок для токена
func (ebr *ErrorBoundaryRegistry) GetErrorBoundariesForToken(tokenID string) []*ErrorBoundarySubscription {
	ebr.mutex.RLock()
	defer ebr.mutex.RUnlock()

	subscriptions, exists := ebr.subscriptions[tokenID]
	if !exists {
		return nil
	}

	// Return copy to avoid concurrent modification
	result := make([]*ErrorBoundarySubscription, len(subscriptions))
	copy(result, subscriptions)
	return result
}

// FindMatchingErrorBoundary finds error boundary that matches error code
// Находит граничное событие ошибки которое соответствует коду ошибки
func (ebr *ErrorBoundaryRegistry) FindMatchingErrorBoundary(tokenID, errorCode string) *ErrorBoundarySubscription {
	subscriptions := ebr.GetErrorBoundariesForToken(tokenID)
	if subscriptions == nil {
		return nil
	}

	for _, subscription := range subscriptions {
		if subscription.ErrorCode == errorCode {
			logger.Info("Found matching error boundary for error code",
				logger.String("token_id", tokenID),
				logger.String("error_code", errorCode),
				logger.String("boundary_element_id", subscription.ElementID))
			return subscription
		}
	}

	return nil
}

// RemoveErrorBoundariesForToken removes all error boundary subscriptions for token
// Удаляет все подписки на граничные события ошибок для токена
func (ebr *ErrorBoundaryRegistry) RemoveErrorBoundariesForToken(tokenID string) {
	ebr.mutex.Lock()
	defer ebr.mutex.Unlock()

	if _, exists := ebr.subscriptions[tokenID]; exists {
		logger.Info("Removing error boundary subscriptions for token",
			logger.String("token_id", tokenID))
		delete(ebr.subscriptions, tokenID)
	}
}

// GetAllSubscriptions returns all active subscriptions for debugging
// Возвращает все активные подписки для отладки
func (ebr *ErrorBoundaryRegistry) GetAllSubscriptions() map[string][]*ErrorBoundarySubscription {
	ebr.mutex.RLock()
	defer ebr.mutex.RUnlock()

	result := make(map[string][]*ErrorBoundarySubscription)
	for tokenID, subscriptions := range ebr.subscriptions {
		result[tokenID] = make([]*ErrorBoundarySubscription, len(subscriptions))
		copy(result[tokenID], subscriptions)
	}
	return result
}
