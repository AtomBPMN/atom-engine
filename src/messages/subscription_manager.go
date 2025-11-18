/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package messages

import (
	"context"
	"fmt"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// SubscriptionManager manages message subscriptions
type SubscriptionManager struct {
	storage   storage.Storage
	logger    logger.ComponentLogger
	isRunning bool
}

// NewSubscriptionManager creates new subscription manager
func NewSubscriptionManager(storage storage.Storage, logger logger.ComponentLogger) *SubscriptionManager {
	return &SubscriptionManager{
		storage: storage,
		logger:  logger,
	}
}

// Start starts the subscription manager
func (sm *SubscriptionManager) Start() error {
	sm.logger.Info("Starting subscription manager")
	sm.isRunning = true
	sm.logger.Info("Subscription manager started")
	return nil
}

// Stop stops the subscription manager
func (sm *SubscriptionManager) Stop() {
	sm.logger.Info("Stopping subscription manager")
	sm.isRunning = false
	sm.logger.Info("Subscription manager stopped")
}

// CreateSubscription creates a new message subscription
func (sm *SubscriptionManager) CreateSubscription(
	ctx context.Context,
	subscription *models.ProcessMessageSubscription,
) error {
	sm.logger.Info("Creating message subscription",
		logger.String("messageName", subscription.MessageName),
		logger.String("startEventID", subscription.StartEventID),
	)

	// Check if subscription already exists
	existing, err := sm.storage.GetProcessMessageSubscription(
		ctx,
		subscription.TenantID,
		subscription.ProcessDefinitionKey,
		subscription.StartEventID,
	)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}

	if existing != nil {
		sm.logger.Info("Subscription already exists - will reuse existing",
			logger.String("existing_id", existing.ID),
			logger.String("process_key", subscription.ProcessDefinitionKey),
			logger.String("event_id", subscription.StartEventID))
		// Return success but indicate that subscription already exists
		// Возвращаем успех но указываем что подписка уже существует
		return fmt.Errorf(
			"subscription already exists for process %s, event %s message_name=%s",
			subscription.ProcessDefinitionKey,
			subscription.StartEventID,
			subscription.MessageName,
		)
	}

	// Save subscription
	if err := sm.storage.SaveProcessMessageSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	sm.logger.Info("Message subscription created", logger.String("id", subscription.ID))
	return nil
}

// DeleteSubscription deletes a message subscription
func (sm *SubscriptionManager) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	sm.logger.Info("Deleting message subscription", logger.String("subscriptionID", subscriptionID))

	// Get subscription first to log details
	subscription, err := sm.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	// Delete subscription
	if err := sm.storage.DeleteProcessMessageSubscription(ctx, subscriptionID); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	sm.logger.Info("Message subscription deleted",
		logger.String("processKey", subscription.ProcessDefinitionKey),
		logger.String("messageName", subscription.MessageName),
	)

	return nil
}

// ListSubscriptions lists message subscriptions
func (sm *SubscriptionManager) ListSubscriptions(
	ctx context.Context,
	tenantID string,
	limit, offset int,
) ([]*models.ProcessMessageSubscription, error) {
	sm.logger.Debug("Listing message subscriptions", logger.Int("limit", limit), logger.Int("offset", offset))

	subscriptions, err := sm.storage.ListProcessMessageSubscriptions(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	sm.logger.Debug("Listed message subscriptions", logger.Int("count", len(subscriptions)))
	return subscriptions, nil
}

// GetSubscription gets message subscription by process key and event ID
func (sm *SubscriptionManager) GetSubscription(
	ctx context.Context,
	tenantID, processKey, startEventID string,
) (*models.ProcessMessageSubscription, error) {
	sm.logger.Debug("Getting message subscription",
		logger.String("processKey", processKey),
		logger.String("startEventID", startEventID),
	)

	subscription, err := sm.storage.GetProcessMessageSubscription(ctx, tenantID, processKey, startEventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription != nil {
		sm.logger.Debug("Message subscription found", logger.String("messageName", subscription.MessageName))
	} else {
		sm.logger.Debug("Message subscription not found")
	}

	return subscription, nil
}

// GetSubscriptionByID gets message subscription by ID
func (sm *SubscriptionManager) GetSubscriptionByID(
	ctx context.Context,
	subscriptionID string,
) (*models.ProcessMessageSubscription, error) {
	sm.logger.Debug("Getting message subscription by ID", logger.String("subscriptionID", subscriptionID))

	// Get all subscriptions and find by ID (storage interface limitation)
	subscriptions, err := sm.storage.ListProcessMessageSubscriptions(ctx, "", 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	for _, subscription := range subscriptions {
		if subscription.ID == subscriptionID {
			sm.logger.Debug("Message subscription found by ID", logger.String("messageName", subscription.MessageName))
			return subscription, nil
		}
	}

	sm.logger.Debug("Message subscription not found by ID", logger.String("subscriptionID", subscriptionID))
	return nil, nil
}

// UpdateSubscription updates a message subscription
func (sm *SubscriptionManager) UpdateSubscription(
	ctx context.Context,
	subscription *models.ProcessMessageSubscription,
) error {
	sm.logger.Info("Updating message subscription", logger.String("subscriptionID", subscription.ID))

	// Check if subscription exists
	existing, err := sm.GetSubscriptionByID(ctx, subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing subscription: %w", err)
	}

	if existing == nil {
		return fmt.Errorf("subscription not found: %s", subscription.ID)
	}

	// Update subscription
	if err := sm.storage.SaveProcessMessageSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	sm.logger.Info("Message subscription updated")
	return nil
}

// ActivateSubscription activates a message subscription
func (sm *SubscriptionManager) ActivateSubscription(ctx context.Context, subscriptionID string) error {
	sm.logger.Info("Activating message subscription", logger.String("subscriptionID", subscriptionID))

	subscription, err := sm.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	if subscription.IsActive {
		sm.logger.Info("Message subscription already active")
		return nil
	}

	subscription.IsActive = true
	subscription.UpdatedAt = time.Now()

	if err := sm.storage.SaveProcessMessageSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}

	sm.logger.Info("Message subscription activated")
	return nil
}

// DeactivateSubscription deactivates a message subscription
func (sm *SubscriptionManager) DeactivateSubscription(ctx context.Context, subscriptionID string) error {
	sm.logger.Info("Deactivating message subscription", logger.String("subscriptionID", subscriptionID))

	subscription, err := sm.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if subscription == nil {
		return fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	if !subscription.IsActive {
		sm.logger.Info("Message subscription already inactive")
		return nil
	}

	subscription.IsActive = false
	subscription.UpdatedAt = time.Now()

	if err := sm.storage.SaveProcessMessageSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to deactivate subscription: %w", err)
	}

	sm.logger.Info("Message subscription deactivated")
	return nil
}

// GetSubscriptionsByMessageName gets subscriptions by message name
func (sm *SubscriptionManager) GetSubscriptionsByMessageName(
	ctx context.Context,
	tenantID, messageName string,
) ([]*models.ProcessMessageSubscription, error) {
	sm.logger.Debug("Getting subscriptions by message name", logger.String("messageName", messageName))

	subscriptions, err := sm.storage.ListProcessMessageSubscriptions(ctx, tenantID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var matchingSubscriptions []*models.ProcessMessageSubscription
	for _, subscription := range subscriptions {
		if subscription.MessageName == messageName && subscription.IsActive {
			matchingSubscriptions = append(matchingSubscriptions, subscription)
		}
	}

	sm.logger.Debug("Found subscriptions by message name",
		logger.String("messageName", messageName),
		logger.Int("count", len(subscriptions)),
	)
	return matchingSubscriptions, nil
}
