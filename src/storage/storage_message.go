/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"atom-engine/src/core/models"

	"github.com/dgraph-io/badger/v3"
)

// Message storage methods

// SaveProcessMessageSubscription saves process message subscription
func (bs *BadgerStorage) SaveProcessMessageSubscription(ctx context.Context, subscription *models.ProcessMessageSubscription) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	key := fmt.Sprintf("msg_sub:%s", subscription.ID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetProcessMessageSubscription gets process message subscription
func (bs *BadgerStorage) GetProcessMessageSubscription(ctx context.Context, tenantID, processKey, startEventID string) (*models.ProcessMessageSubscription, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var subscription *models.ProcessMessageSubscription
	prefix := []byte("msg_sub:")

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var sub models.ProcessMessageSubscription
				if err := json.Unmarshal(val, &sub); err != nil {
					return err
				}

				if sub.TenantID == tenantID && sub.ProcessDefinitionKey == processKey && sub.StartEventID == startEventID {
					subscription = &sub
				}
				return nil
			})
			if err != nil {
				return err
			}
			if subscription != nil {
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

// ListProcessMessageSubscriptions lists process message subscriptions
func (bs *BadgerStorage) ListProcessMessageSubscriptions(ctx context.Context, tenantID string, limit, offset int) ([]*models.ProcessMessageSubscription, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var subscriptions []*models.ProcessMessageSubscription
	prefix := []byte("msg_sub:")

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var sub models.ProcessMessageSubscription
				if err := json.Unmarshal(val, &sub); err != nil {
					return err
				}

				// Filter by tenant if specified
				if tenantID != "" && sub.TenantID != tenantID {
					return nil
				}

				// Apply offset
				if skipped < offset {
					skipped++
					return nil
				}

				subscriptions = append(subscriptions, &sub)
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return subscriptions, nil
}

// DeleteProcessMessageSubscription deletes process message subscription
func (bs *BadgerStorage) DeleteProcessMessageSubscription(ctx context.Context, subscriptionID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := fmt.Sprintf("msg_sub:%s", subscriptionID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// SaveBufferedMessage saves buffered message
func (bs *BadgerStorage) SaveBufferedMessage(ctx context.Context, message *models.BufferedMessage) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	key := fmt.Sprintf("buf_msg:%s", message.ID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetBufferedMessage gets buffered message
func (bs *BadgerStorage) GetBufferedMessage(ctx context.Context, messageID string) (*models.BufferedMessage, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := fmt.Sprintf("buf_msg:%s", messageID)
	var message *models.BufferedMessage

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			message = &models.BufferedMessage{}
			return json.Unmarshal(val, message)
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// ListBufferedMessages lists buffered messages with performance optimizations
func (bs *BadgerStorage) ListBufferedMessages(ctx context.Context, tenantID string, limit, offset int) ([]*models.BufferedMessage, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var messages []*models.BufferedMessage
	prefix := []byte("buf_msg:")

	// Calculate optimal prefetch size based on batch configuration
	maxBatchCount, _ := bs.GetBatchConfig()
	prefetchSize := maxBatchCount
	if limit > 0 && limit < prefetchSize {
		prefetchSize = limit * 2 // Prefetch a bit more than needed
	}

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = prefetchSize
		opts.PrefetchValues = true // Prefetch values for better performance
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0

		// Pre-allocate slice for better memory performance
		if limit > 0 {
			messages = make([]*models.BufferedMessage, 0, limit)
		} else {
			messages = make([]*models.BufferedMessage, 0, 100) // Default capacity
		}

		for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...) // Copy value for safety
				return nil
			})
			if err != nil {
				continue // Skip corrupted entries
			}

			var msg models.BufferedMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				continue // Skip corrupted entries
			}

			// Filter by tenant if specified (early filtering for performance)
			if tenantID != "" && msg.TenantID != tenantID {
				continue
			}

			// Apply offset
			if skipped < offset {
				skipped++
				continue
			}

			messages = append(messages, &msg)
			count++
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return messages, nil
}

// DeleteBufferedMessage deletes buffered message
func (bs *BadgerStorage) DeleteBufferedMessage(ctx context.Context, messageID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := fmt.Sprintf("buf_msg:%s", messageID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// SaveMessageCorrelationResult saves message correlation result
func (bs *BadgerStorage) SaveMessageCorrelationResult(ctx context.Context, result *models.MessageCorrelationResult) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	key := fmt.Sprintf("msg_corr:%s", result.ID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// ListMessageCorrelationResults lists message correlation results
func (bs *BadgerStorage) ListMessageCorrelationResults(ctx context.Context, tenantID, messageName, processKey string, limit, offset int) ([]*models.MessageCorrelationResult, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var results []*models.MessageCorrelationResult
	prefix := []byte("msg_corr:")

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		skipped := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var result models.MessageCorrelationResult
				if err := json.Unmarshal(val, &result); err != nil {
					return err
				}

				// Apply filters
				if tenantID != "" && result.TenantID != tenantID {
					return nil
				}
				if messageName != "" && result.MessageName != messageName {
					return nil
				}
				// processKey filter would need process definition lookup

				// Apply offset
				if skipped < offset {
					skipped++
					return nil
				}

				results = append(results, &result)
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list results: %w", err)
	}

	return results, nil
}

// DeleteMessageCorrelationResult deletes message correlation result
func (bs *BadgerStorage) DeleteMessageCorrelationResult(ctx context.Context, resultID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := fmt.Sprintf("msg_corr:%s", resultID)
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
