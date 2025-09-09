/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"fmt"
	"time"

	"atom-engine/src/core/logger"

	"github.com/dgraph-io/badger/v3"
)

// NewStorage creates new storage instance
// Создает новый экземпляр storage
func NewStorage(config *Config) Storage {
	return &BadgerStorage{
		config: config,
		ready:  false,
	}
}

// Init initializes database connection
// Инициализирует подключение к базе данных
func (s *BadgerStorage) Init() error {
	logger.Info("Initializing BadgerDB with performance optimizations", logger.String("path", s.config.Path))

	opts := badger.DefaultOptions(s.config.Path)
	opts.Logger = nil // Disable badger logs

	// Apply configuration options
	s.applyPerformanceOptions(&opts)

	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	s.db = db
	logger.Info("BadgerDB initialized successfully with performance optimizations")
	return nil
}

// applyPerformanceOptions applies performance configuration to BadgerDB options
// Применяет настройки производительности к опциям BadgerDB
func (s *BadgerStorage) applyPerformanceOptions(opts *badger.Options) {
	if s.config.Options == nil {
		logger.Debug("No performance options configured, using defaults")
		return
	}

	// Apply basic options
	if s.config.Options.SyncWrites != nil {
		opts.SyncWrites = *s.config.Options.SyncWrites
		logger.Debug("Applied SyncWrites", logger.Bool("value", *s.config.Options.SyncWrites))
	}

	if s.config.Options.ValueLogFileSize != nil {
		opts.ValueLogFileSize = *s.config.Options.ValueLogFileSize
		logger.Debug("Applied ValueLogFileSize", logger.Int64("value", *s.config.Options.ValueLogFileSize))
	}

	// Apply performance options
	if s.config.Options.Performance != nil {
		perf := s.config.Options.Performance

		// Memory settings
		if perf.MemTableSize != nil {
			opts.MemTableSize = *perf.MemTableSize
			logger.Debug("Applied MemTableSize", logger.Int64("value", *perf.MemTableSize))
		}

		if perf.NumMemtables != nil {
			opts.NumMemtables = *perf.NumMemtables
			logger.Debug("Applied NumMemtables", logger.Int("value", *perf.NumMemtables))
		}

		if perf.NumLevelZeroTables != nil {
			opts.NumLevelZeroTables = *perf.NumLevelZeroTables
			logger.Debug("Applied NumLevelZeroTables", logger.Int("value", *perf.NumLevelZeroTables))
		}

		if perf.NumLevelZeroTablesStall != nil {
			opts.NumLevelZeroTablesStall = *perf.NumLevelZeroTablesStall
			logger.Debug("Applied NumLevelZeroTablesStall", logger.Int("value", *perf.NumLevelZeroTablesStall))
		}

		// Cache settings
		if perf.ValueCacheSize != nil {
			// Note: BadgerDB uses BlockCacheSize for general caching
			opts.BlockCacheSize = *perf.ValueCacheSize
			logger.Debug("Applied ValueCacheSize as BlockCacheSize", logger.Int64("value", *perf.ValueCacheSize))
		}

		if perf.BlockCacheSize != nil {
			opts.BlockCacheSize = *perf.BlockCacheSize
			logger.Debug("Applied BlockCacheSize", logger.Int64("value", *perf.BlockCacheSize))
		}

		if perf.IndexCacheSize != nil {
			opts.IndexCacheSize = *perf.IndexCacheSize
			logger.Debug("Applied IndexCacheSize", logger.Int64("value", *perf.IndexCacheSize))
		}

		// Table settings
		if perf.BaseTableSize != nil {
			opts.BaseTableSize = *perf.BaseTableSize
			logger.Debug("Applied BaseTableSize", logger.Int64("value", *perf.BaseTableSize))
		}

		// Note: MaxTableSize not available in BadgerDB v3, using BaseTableSize instead
		if perf.MaxTableSize != nil {
			logger.Debug("MaxTableSize not supported in BadgerDB v3, consider using BaseTableSize", logger.Int64("value", *perf.MaxTableSize))
		}

		if perf.LevelSizeMultiplier != nil {
			opts.LevelSizeMultiplier = *perf.LevelSizeMultiplier
			logger.Debug("Applied LevelSizeMultiplier", logger.Int("value", *perf.LevelSizeMultiplier))
		}

		// Compaction settings
		if perf.NumCompactors != nil {
			opts.NumCompactors = *perf.NumCompactors
			logger.Debug("Applied NumCompactors", logger.Int("value", *perf.NumCompactors))
		}

		if perf.CompactL0OnClose != nil {
			opts.CompactL0OnClose = *perf.CompactL0OnClose
			logger.Debug("Applied CompactL0OnClose", logger.Bool("value", *perf.CompactL0OnClose))
		}

		// Loading mode settings - Note: BadgerDB v3 may have different field names
		if perf.TableLoadingMode != nil {
			logger.Debug("TableLoadingMode setting noted (implementation depends on BadgerDB version)", logger.String("mode", *perf.TableLoadingMode))
			// NOTE: Field name may vary in different BadgerDB versions
		}

		if perf.ValueLogLoadingMode != nil {
			logger.Debug("ValueLogLoadingMode setting noted (implementation depends on BadgerDB version)", logger.String("mode", *perf.ValueLogLoadingMode))
			// NOTE: Field name may vary in different BadgerDB versions
		}

		// Advanced settings
		if perf.BloomFalsePositive != nil {
			opts.BloomFalsePositive = *perf.BloomFalsePositive
			logger.Debug("Applied BloomFalsePositive", logger.Float64("value", *perf.BloomFalsePositive))
		}

		if perf.DetectConflicts != nil {
			opts.DetectConflicts = *perf.DetectConflicts
			logger.Debug("Applied DetectConflicts", logger.Bool("value", *perf.DetectConflicts))
		}

		if perf.ManageTxns != nil {
			// Note: ManageTxns is typically handled at application level in BadgerDB v3
			logger.Debug("ManageTxns setting noted", logger.Bool("value", *perf.ManageTxns))
		}
	}

	logger.Info("Performance options applied to BadgerDB")
}

// Start starts database
// Запускает базу данных
func (s *BadgerStorage) Start() error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	logger.Info("Starting BadgerDB storage...")
	s.ready = true
	s.startTime = time.Now()
	logger.Info("BadgerDB storage is ready")
	return nil
}

// Stop closes database connection
// Закрывает подключение к базе данных
func (s *BadgerStorage) Stop() error {
	s.ready = false
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// IsReady returns storage ready status
// Возвращает статус готовности storage
func (s *BadgerStorage) IsReady() bool {
	return s.ready
}
