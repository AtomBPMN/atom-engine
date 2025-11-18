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

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"

	"github.com/dgraph-io/badger/v3"
)

// BatchOperation represents a single batch operation
// Представляет одну операцию в батче
type BatchOperation struct {
	Key   []byte
	Value []byte
	Type  BatchOperationType
}

// BatchOperationType defines types of batch operations
// Определяет типы батчевых операций
type BatchOperationType int

const (
	BatchSet BatchOperationType = iota
	BatchDelete
)

// ExecuteBatch executes multiple operations in a single transaction
// Выполняет множественные операции в одной транзакции
func (s *BadgerStorage) ExecuteBatch(operations []BatchOperation) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	if len(operations) == 0 {
		return nil
	}

	return s.db.Update(func(txn *badger.Txn) error {
		for _, op := range operations {
			switch op.Type {
			case BatchSet:
				if err := txn.Set(op.Key, op.Value); err != nil {
					return fmt.Errorf("failed to set key %s: %w", string(op.Key), err)
				}
			case BatchDelete:
				if err := txn.Delete(op.Key); err != nil {
					return fmt.Errorf("failed to delete key %s: %w", string(op.Key), err)
				}
			}
		}
		return nil
	})
}

// SaveBufferedMessagesBatch saves multiple buffered messages in a single transaction
// Сохраняет множественные буферизованные сообщения в одной транзакции
func (s *BadgerStorage) SaveBufferedMessagesBatch(ctx context.Context, messages []*models.BufferedMessage) error {
	if len(messages) == 0 {
		return nil
	}

	logger.Debug("Saving batch of buffered messages", logger.Int("count", len(messages)))

	operations := make([]BatchOperation, 0, len(messages))

	for _, message := range messages {
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to serialize buffered message %s: %w", message.ID, err)
		}

		key := fmt.Sprintf("messages:buffered:%s", message.ID)
		operations = append(operations, BatchOperation{
			Key:   []byte(key),
			Value: data,
			Type:  BatchSet,
		})
	}

	if err := s.ExecuteBatch(operations); err != nil {
		return fmt.Errorf("failed to execute batch save of buffered messages: %w", err)
	}

	logger.Info("Successfully saved batch of buffered messages", logger.Int("count", len(messages)))
	return nil
}

// SaveTokensBatch saves multiple tokens in a single transaction
// Сохраняет множественные токены в одной транзакции
func (s *BadgerStorage) SaveTokensBatch(tokens []*models.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	logger.Debug("Saving batch of tokens", logger.Int("count", len(tokens)))

	operations := make([]BatchOperation, 0, len(tokens))

	for _, token := range tokens {
		data, err := token.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to serialize token %s: %w", token.TokenID, err)
		}

		key := TokenPrefix + token.TokenID
		operations = append(operations, BatchOperation{
			Key:   []byte(key),
			Value: data,
			Type:  BatchSet,
		})
	}

	if err := s.ExecuteBatch(operations); err != nil {
		return fmt.Errorf("failed to execute batch save of tokens: %w", err)
	}

	logger.Info("Successfully saved batch of tokens", logger.Int("count", len(tokens)))
	return nil
}

// DeleteMessagesBatch deletes multiple messages in a single transaction
// Удаляет множественные сообщения в одной транзакции
func (s *BadgerStorage) DeleteMessagesBatch(ctx context.Context, messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	logger.Debug("Deleting batch of messages", logger.Int("count", len(messageIDs)))

	operations := make([]BatchOperation, 0, len(messageIDs))

	for _, messageID := range messageIDs {
		key := fmt.Sprintf("messages:buffered:%s", messageID)
		operations = append(operations, BatchOperation{
			Key:  []byte(key),
			Type: BatchDelete,
		})
	}

	if err := s.ExecuteBatch(operations); err != nil {
		return fmt.Errorf("failed to execute batch delete of messages: %w", err)
	}

	logger.Info("Successfully deleted batch of messages", logger.Int("count", len(messageIDs)))
	return nil
}

// CleanupExpiredMessagesBatch efficiently removes expired messages using batches
// Эффективно удаляет просроченные сообщения используя батчи
func (s *BadgerStorage) CleanupExpiredMessagesBatch(ctx context.Context, batchSize int) (int, error) {
	if !s.ready {
		return 0, fmt.Errorf("storage not ready")
	}

	var cleanedCount int
	prefix := []byte("messages:buffered:")

	for {
		var toDelete []string

		// Find expired messages in batches
		err := s.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = prefix
			opts.PrefetchSize = batchSize
			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid() && len(toDelete) < batchSize; it.Next() {
				item := it.Item()

				var data []byte
				err := item.Value(func(val []byte) error {
					data = append([]byte(nil), val...)
					return nil
				})
				if err != nil {
					continue
				}

				var message models.BufferedMessage
				if err := json.Unmarshal(data, &message); err != nil {
					continue
				}

				if message.IsExpired() {
					toDelete = append(toDelete, message.ID)
				}
			}
			return nil
		})

		if err != nil {
			return cleanedCount, fmt.Errorf("failed to find expired messages: %w", err)
		}

		if len(toDelete) == 0 {
			break
		}

		// Delete batch
		if err := s.DeleteMessagesBatch(ctx, toDelete); err != nil {
			return cleanedCount, fmt.Errorf("failed to delete batch of expired messages: %w", err)
		}

		cleanedCount += len(toDelete)

		// If we got less than batch size, we're done
		if len(toDelete) < batchSize {
			break
		}
	}

	if cleanedCount > 0 {
		logger.Info("Cleaned up expired messages in batches", logger.Int("cleaned_count", cleanedCount))
	}

	return cleanedCount, nil
}

// GetBatchConfig returns current batch configuration
// Возвращает текущую конфигурацию батчей
func (s *BadgerStorage) GetBatchConfig() (int, int64) {
	defaultCount := 128
	defaultSize := int64(16 * 1024 * 1024) // 16MB

	if s.config.Options != nil && s.config.Options.Performance != nil {
		if s.config.Options.Performance.MaxBatchCount != nil {
			defaultCount = *s.config.Options.Performance.MaxBatchCount
		}
		if s.config.Options.Performance.MaxBatchSize != nil {
			defaultSize = *s.config.Options.Performance.MaxBatchSize
		}
	}

	return defaultCount, defaultSize
}
