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

// LoadSystemEvents loads recent system events from database
// Загружает последние системные события из базы данных
func (s *BadgerStorage) LoadSystemEvents(limit int) ([]*SystemEventRecord, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var events []*SystemEventRecord
	prefix := []byte("system_events:")

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true // Load newest events first
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix) && count < limit; it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var event SystemEventRecord
				if err := json.Unmarshal(val, &event); err != nil {
					logger.Warn("Failed to unmarshal system event",
						logger.String("key", string(item.Key())),
						logger.String("error", err.Error()))
					return nil // Continue processing other events
				}
				events = append(events, &event)
				return nil
			})

			if err != nil {
				return err
			}
			count++
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to load system events", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to load system events from database: %w", err)
	}

	logger.Debug("Loaded system events from storage",
		logger.Int("count", len(events)),
		logger.Int("limit", limit))
	return events, nil
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

// SaveSystemMetrics saves system metrics to storage
// Сохраняет системные метрики в storage
func (s *BadgerStorage) SaveSystemMetrics(metrics *SystemMetrics) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	metrics.LastUpdated = time.Now()
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal system metrics: %w", err)
	}

	key := "system_metrics:current"
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to save system metrics: %w", err)
	}

	logger.Debug("System metrics saved to storage",
		logger.Int64("requests", metrics.TotalRequests),
		logger.Int64("errors", metrics.TotalErrors),
		logger.Float64("cpu", metrics.CPUUsage))
	return nil
}

// LoadSystemMetrics loads system metrics from storage
// Загружает системные метрики из storage
func (s *BadgerStorage) LoadSystemMetrics() (*SystemMetrics, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var metrics SystemMetrics
	key := "system_metrics:current"

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// Return default empty metrics if not found
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &metrics)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Return default metrics if no data found
			return &SystemMetrics{
				LastUpdated: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to load system metrics: %w", err)
	}

	return &metrics, nil
}

// IncrementRequestCount increments total request count
// Увеличивает счетчик общих запросов
func (s *BadgerStorage) IncrementRequestCount() error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := "system_metrics:request_count"

		// Get current value
		var count int64
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if err == nil {
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &count)
			})
			if err != nil {
				return err
			}
		}

		// Increment and save
		count++
		data, err := json.Marshal(count)
		if err != nil {
			return err
		}

		return txn.Set([]byte(key), data)
	})
}

// IncrementErrorCount increments total error count
// Увеличивает счетчик общих ошибок
func (s *BadgerStorage) IncrementErrorCount() error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := "system_metrics:error_count"

		// Get current value
		var count int64
		item, err := txn.Get([]byte(key))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if err == nil {
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &count)
			})
			if err != nil {
				return err
			}
		}

		// Increment and save
		count++
		data, err := json.Marshal(count)
		if err != nil {
			return err
		}

		return txn.Set([]byte(key), data)
	})
}

// UpdateCPUUsage updates CPU usage metric
// Обновляет метрику использования CPU
func (s *BadgerStorage) UpdateCPUUsage(usage float64) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	key := "system_metrics:cpu_usage"
	data, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("failed to marshal CPU usage: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// UpdateMemoryUsage updates memory usage metric
// Обновляет метрику использования памяти
func (s *BadgerStorage) UpdateMemoryUsage(usage int64) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	key := "system_metrics:memory_usage"
	data, err := json.Marshal(usage)
	if err != nil {
		return fmt.Errorf("failed to marshal memory usage: %w", err)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// SaveRateLimitInfo saves rate limit information to storage
// Сохраняет информацию о rate limit в storage
func (s *BadgerStorage) SaveRateLimitInfo(identifier string, info *RateLimitInfo) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	info.LastAccess = time.Now()
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit info: %w", err)
	}

	key := fmt.Sprintf("rate_limit:%s", identifier)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to save rate limit info: %w", err)
	}

	return nil
}

// LoadRateLimitInfo loads rate limit information from storage
// Загружает информацию о rate limit из storage
func (s *BadgerStorage) LoadRateLimitInfo(identifier string) (*RateLimitInfo, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var info RateLimitInfo
	key := fmt.Sprintf("rate_limit:%s", identifier)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &info)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil // Not found, return nil
		}
		return nil, fmt.Errorf("failed to load rate limit info: %w", err)
	}

	return &info, nil
}

// LoadAllRateLimitInfo loads all rate limit information from storage
// Загружает всю информацию о rate limit из storage
func (s *BadgerStorage) LoadAllRateLimitInfo() (map[string]*RateLimitInfo, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	rateLimits := make(map[string]*RateLimitInfo)

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("rate_limit:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())

			// Extract identifier from key
			identifier := key[len("rate_limit:"):]

			err := item.Value(func(val []byte) error {
				var info RateLimitInfo
				if err := json.Unmarshal(val, &info); err != nil {
					return err
				}
				rateLimits[identifier] = &info
				return nil
			})

			if err != nil {
				logger.Warn("Failed to unmarshal rate limit info", logger.String("key", key), logger.String("error", err.Error()))
				continue
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load rate limit info: %w", err)
	}

	return rateLimits, nil
}

// DeleteRateLimitInfo deletes rate limit information from storage
// Удаляет информацию о rate limit из storage
func (s *BadgerStorage) DeleteRateLimitInfo(identifier string) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	key := fmt.Sprintf("rate_limit:%s", identifier)

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// CleanupExpiredRateLimitInfo removes expired rate limit entries
// Удаляет истекшие записи rate limit
func (s *BadgerStorage) CleanupExpiredRateLimitInfo() error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	now := time.Now()
	var keysToDelete []string

	// First, collect expired keys
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("rate_limit:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				var info RateLimitInfo
				if err := json.Unmarshal(val, &info); err != nil {
					return err
				}

				// Mark for deletion if expired (both reset time and last access are old)
				if now.After(info.ResetTime) && now.Sub(info.LastAccess) > 5*time.Minute {
					keysToDelete = append(keysToDelete, key)
				}
				return nil
			})

			if err != nil {
				logger.Warn("Failed to unmarshal rate limit info during cleanup",
					logger.String("key", key),
					logger.String("error", err.Error()),
				)
				continue
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan rate limit info for cleanup: %w", err)
	}

	// Then, delete expired keys
	if len(keysToDelete) > 0 {
		err = s.db.Update(func(txn *badger.Txn) error {
			for _, key := range keysToDelete {
				if err := txn.Delete([]byte(key)); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to delete expired rate limit info: %w", err)
		}

		logger.Debug("Rate limit cleanup completed", logger.Int("deleted_entries", len(keysToDelete)))
	}

	return nil
}

// SaveMultiplexerState saves multiplexer state to storage
// Сохраняет состояние мультиплексера в storage
func (s *BadgerStorage) SaveMultiplexerState(componentName string, state *MultiplexerState) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	state.LastUpdated = time.Now()
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal multiplexer state: %w", err)
	}

	key := fmt.Sprintf("multiplexer_state:%s", componentName)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to save multiplexer state: %w", err)
	}

	logger.Debug("Multiplexer state saved to storage", logger.String("component", componentName))
	return nil
}

// LoadMultiplexerState loads multiplexer state from storage
// Загружает состояние мультиплексера из storage
func (s *BadgerStorage) LoadMultiplexerState(componentName string) (*MultiplexerState, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var state MultiplexerState
	key := fmt.Sprintf("multiplexer_state:%s", componentName)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &state)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Return default state if not found
			return &MultiplexerState{
				ComponentName: componentName,
				IsRunning:     false,
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to load multiplexer state: %w", err)
	}

	return &state, nil
}

// SaveChannelStats saves channel statistics to storage
// Сохраняет статистику каналов в storage
func (s *BadgerStorage) SaveChannelStats(componentName string, stats *ChannelStatistics) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	stats.LastUpdated = time.Now()
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal channel stats: %w", err)
	}

	key := fmt.Sprintf("channel_stats:%s", componentName)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to save channel stats: %w", err)
	}

	logger.Debug("Channel stats saved to storage", logger.String("component", componentName))
	return nil
}

// LoadChannelStats loads channel statistics from storage
// Загружает статистику каналов из storage
func (s *BadgerStorage) LoadChannelStats(componentName string) (*ChannelStatistics, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var stats ChannelStatistics
	key := fmt.Sprintf("channel_stats:%s", componentName)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &stats)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Return default stats if not found
			return &ChannelStatistics{
				ComponentName: componentName,
				ChannelStats:  make(map[string]*ChannelStat),
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to load channel stats: %w", err)
	}

	return &stats, nil
}

// SaveRoutingMetrics saves routing metrics to storage
// Сохраняет метрики маршрутизации в storage
func (s *BadgerStorage) SaveRoutingMetrics(componentName string, metrics *RoutingMetrics) error {
	if !s.ready {
		return fmt.Errorf("storage not ready")
	}

	metrics.LastUpdated = time.Now()
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal routing metrics: %w", err)
	}

	key := fmt.Sprintf("routing_metrics:%s", componentName)
	err = s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})

	if err != nil {
		return fmt.Errorf("failed to save routing metrics: %w", err)
	}

	logger.Debug("Routing metrics saved to storage", logger.String("component", componentName))
	return nil
}

// LoadRoutingMetrics loads routing metrics from storage
// Загружает метрики маршрутизации из storage
func (s *BadgerStorage) LoadRoutingMetrics(componentName string) (*RoutingMetrics, error) {
	if !s.ready {
		return nil, fmt.Errorf("storage not ready")
	}

	var metrics RoutingMetrics
	key := fmt.Sprintf("routing_metrics:%s", componentName)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &metrics)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Return default metrics if not found
			return &RoutingMetrics{
				ComponentName: componentName,
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to load routing metrics: %w", err)
	}

	return &metrics, nil
}
