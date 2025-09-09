/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"atom-engine/src/core/logger"

	"github.com/dgraph-io/badger/v3"
)

// SaveTimer saves timer to database
// Сохраняет таймер в базу данных
func (s *BadgerStorage) SaveTimer(timer *TimerRecord) error {
	timer.UpdatedAt = time.Now()
	if timer.CreatedAt.IsZero() {
		timer.CreatedAt = time.Now()
	}

	key := fmt.Sprintf("timer_%s", timer.ID)
	err := s.saveJSON(key, timer)
	if err != nil {
		logger.Error("Failed to save timer",
			logger.String("timer_id", timer.ID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to save timer to database: %w", err)
	}

	logger.Debug("Timer saved to storage",
		logger.String("timer_id", timer.ID),
		logger.String("element_id", timer.ElementID),
		logger.String("state", timer.State))

	return nil
}

// LoadTimer loads timer from database
// Загружает таймер из базы данных
func (s *BadgerStorage) LoadTimer(timerID string) (*TimerRecord, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var timer TimerRecord
	key := fmt.Sprintf("timer_%s", timerID)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &timer)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, fmt.Errorf("timer not found: %s", timerID)
	}

	if err != nil {
		logger.Error("Failed to load timer",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load timer from database: %w", err)
	}

	logger.Debug("Timer loaded from storage",
		logger.String("timer_id", timer.ID),
		logger.String("element_id", timer.ElementID),
		logger.String("state", timer.State))

	return &timer, nil
}

// LoadAllTimers loads all timers from database
// Загружает все таймеры из базы данных
func (s *BadgerStorage) LoadAllTimers() ([]*TimerRecord, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var timers []*TimerRecord
	prefix := []byte("timer_")

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var timer TimerRecord
				if err := json.Unmarshal(val, &timer); err != nil {
					logger.Warn("Failed to unmarshal timer",
						logger.String("key", string(item.Key())),
						logger.String("error", err.Error()))
					return nil // Continue processing other timers
				}
				timers = append(timers, &timer)
				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to load timers", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load timers from database: %w", err)
	}

	logger.Info("Loaded timers from storage", logger.Int("count", len(timers)))
	return timers, nil
}

// DeleteTimer deletes timer from database
// Удаляет таймер из базы данных
func (s *BadgerStorage) DeleteTimer(timerID string) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	key := fmt.Sprintf("timer_%s", timerID)
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err == badger.ErrKeyNotFound {
		logger.Warn("Timer not found for deletion", logger.String("timer_id", timerID))
		return fmt.Errorf("timer not found: %s", timerID)
	}

	if err != nil {
		logger.Error("Failed to delete timer",
			logger.String("timer_id", timerID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to delete timer from database: %w", err)
	}

	logger.Debug("Timer deleted from storage", logger.String("timer_id", timerID))
	return nil
}

// UpdateTimer updates timer in database
// Обновляет таймер в базе данных
func (s *BadgerStorage) UpdateTimer(timer *TimerRecord) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	timer.UpdatedAt = time.Now()

	data, err := json.Marshal(timer)
	if err != nil {
		return fmt.Errorf("failed to marshal timer: %w", err)
	}

	key := fmt.Sprintf("timer_%s", timer.ID)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		logger.Error("Failed to update timer",
			logger.String("timer_id", timer.ID),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to update timer in database: %w", err)
	}

	logger.Debug("Timer updated in storage",
		logger.String("timer_id", timer.ID),
		logger.String("state", timer.State))

	return nil
}
