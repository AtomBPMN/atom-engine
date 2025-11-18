/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"sync"
	"time"

	"atom-engine/src/core/logger"
)

// SignalSubscription represents a signal subscription
// Представляет подписку на сигнал
type SignalSubscription struct {
	SignalName     string                 `json:"signal_name"`
	TokenID        string                 `json:"token_id"`
	ElementID      string                 `json:"element_id"`
	CancelActivity bool                   `json:"cancel_activity"`
	Variables      map[string]interface{} `json:"variables"`
	CreatedAt      time.Time              `json:"created_at"`
}

// SignalManager manages signal subscriptions and broadcasting
// Управляет подписками на сигналы и их broadcasting
type SignalManager struct {
	subscriptions map[string][]*SignalSubscription // map[signalName]subscriptions
	mutex         sync.RWMutex
	component     ComponentInterface
}

// NewSignalManager creates a new signal manager
// Создает новый менеджер сигналов
func NewSignalManager(component ComponentInterface) *SignalManager {
	return &SignalManager{
		subscriptions: make(map[string][]*SignalSubscription),
		component:     component,
	}
}

// Subscribe adds a signal subscription
// Добавляет подписку на сигнал
func (sm *SignalManager) Subscribe(
	signalName, tokenID, elementID string,
	cancelActivity bool,
	variables map[string]interface{},
) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	subscription := &SignalSubscription{
		SignalName:     signalName,
		TokenID:        tokenID,
		ElementID:      elementID,
		CancelActivity: cancelActivity,
		Variables:      variables,
		CreatedAt:      time.Now(),
	}

	sm.subscriptions[signalName] = append(sm.subscriptions[signalName], subscription)

	logger.Info("Signal subscription added",
		logger.String("signal_name", signalName),
		logger.String("token_id", tokenID),
		logger.String("element_id", elementID),
		logger.Bool("cancel_activity", cancelActivity))

	return nil
}

// BroadcastSignal sends a signal to all subscribers
// Отправляет сигнал всем подписчикам
func (sm *SignalManager) BroadcastSignal(signalName string, variables map[string]interface{}) error {
	sm.mutex.Lock()
	subscriptions := sm.subscriptions[signalName]
	// Clear subscriptions after broadcasting (signals are consumed)
	delete(sm.subscriptions, signalName)
	sm.mutex.Unlock()

	if len(subscriptions) == 0 {
		logger.Info("No subscribers for signal", logger.String("signal_name", signalName))
		return nil
	}

	logger.Info("Broadcasting signal to subscribers",
		logger.String("signal_name", signalName),
		logger.Int("subscriber_count", len(subscriptions)))

	// Process each subscription
	for _, subscription := range subscriptions {
		if err := sm.processSignalSubscription(subscription, variables); err != nil {
			logger.Error("Failed to process signal subscription",
				logger.String("signal_name", signalName),
				logger.String("token_id", subscription.TokenID),
				logger.String("error", err.Error()))
			continue
		}
	}

	return nil
}

// processSignalSubscription processes a single signal subscription
// Обрабатывает одну подписку на сигнал
func (sm *SignalManager) processSignalSubscription(
	subscription *SignalSubscription,
	signalVariables map[string]interface{},
) error {
	logger.Info("Processing signal subscription",
		logger.String("signal_name", subscription.SignalName),
		logger.String("token_id", subscription.TokenID),
		logger.String("element_id", subscription.ElementID))

	// Merge subscription variables with signal variables
	mergedVariables := make(map[string]interface{})
	for k, v := range subscription.Variables {
		mergedVariables[k] = v
	}
	for k, v := range signalVariables {
		mergedVariables[k] = v
	}

	// Use message callback mechanism to trigger boundary event
	// Используем механизм message callback для активации boundary события
	return sm.component.HandleMessageCallback(
		subscription.ElementID,  // messageID (using elementID)
		subscription.SignalName, // messageName (signal name)
		subscription.TokenID,    // correlationKey (using tokenID)
		subscription.TokenID,    // tokenID
		mergedVariables,         // variables
	)
}

// UnsubscribeByToken removes all subscriptions for a token
// Удаляет все подписки для токена
func (sm *SignalManager) UnsubscribeByToken(tokenID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	count := 0
	for signalName, subscriptions := range sm.subscriptions {
		filtered := make([]*SignalSubscription, 0)
		for _, sub := range subscriptions {
			if sub.TokenID != tokenID {
				filtered = append(filtered, sub)
			} else {
				count++
			}
		}
		if len(filtered) == 0 {
			delete(sm.subscriptions, signalName)
		} else {
			sm.subscriptions[signalName] = filtered
		}
	}

	if count > 0 {
		logger.Info("Signal subscriptions removed for token",
			logger.String("token_id", tokenID),
			logger.Int("removed_count", count))
	}

	return nil
}

// GetSubscriptions returns current subscriptions (for debugging)
// Возвращает текущие подписки (для отладки)
func (sm *SignalManager) GetSubscriptions() map[string][]*SignalSubscription {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Deep copy to avoid race conditions
	result := make(map[string][]*SignalSubscription)
	for signalName, subscriptions := range sm.subscriptions {
		copySubscriptions := make([]*SignalSubscription, len(subscriptions))
		copy(copySubscriptions, subscriptions)
		result[signalName] = copySubscriptions
	}

	return result
}
