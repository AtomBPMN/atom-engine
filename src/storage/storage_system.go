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
	"os"
	"path/filepath"
	"time"

	"atom-engine/src/core/logger"

	"github.com/dgraph-io/badger/v3"
)

// LogSystemEvent logs system events to database
// Логирует системные события в базу данных
func (s *BadgerStorage) LogSystemEvent(eventType, status, message string) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	event := SystemEventRecord{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: eventType,
		Status:    status,
		Message:   message,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := fmt.Sprintf("system_events:%s", event.ID)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to log system event: %w", err)
	}

	// Note: Removed forced sync for better performance
	// BadgerDB handles sync according to SyncWrites configuration

	logger.Debug("System event logged to DB",
		logger.String("event_type", eventType),
		logger.String("status", status),
		logger.String("message", message))
	return nil
}

// GetStatus returns current storage status
// Возвращает текущий статус storage
func (s *BadgerStorage) GetStatus() (*StorageStatus, error) {
	status := &StorageStatus{
		IsConnected: s.db != nil,
		IsHealthy:   s.ready && s.db != nil,
		Status:      "unknown",
	}

	if s.ready && s.db != nil {
		status.Status = "ready"
		status.UptimeSeconds = int64(time.Since(s.startTime).Seconds())
	} else if s.db != nil {
		status.Status = "initializing"
	} else {
		status.Status = "disconnected"
	}

	return status, nil
}

// GetInfo returns storage information and statistics
// Возвращает информацию и статистику storage
func (s *BadgerStorage) GetInfo() (*StorageInfo, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	info := &StorageInfo{
		DatabasePath: s.config.Path,
		Statistics:   make(map[string]string),
	}

	// Get database size
	if stat, err := os.Stat(s.config.Path); err == nil && stat.IsDir() {
		size, err := getDirSize(s.config.Path)
		if err == nil {
			info.UsedSizeBytes = size
		}
	}

	// Get key count and other statistics
	var keyCount int64
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			keyCount++
		}
		return nil
	})

	if err == nil {
		info.TotalKeys = keyCount
	}

	// Add basic statistics
	info.Statistics["db_type"] = "badger"
	info.Statistics["key_count"] = fmt.Sprintf("%d", keyCount)
	info.Statistics["db_path"] = s.config.Path

	return info, nil
}

// getDirSize calculates directory size recursively
// Вычисляет размер директории рекурсивно
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
