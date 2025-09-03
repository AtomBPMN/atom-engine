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
	logger.Info("Initializing BadgerDB", logger.String("path", s.config.Path))

	opts := badger.DefaultOptions(s.config.Path)
	opts.Logger = nil // Disable badger logs

	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	s.db = db
	logger.Info("BadgerDB initialized successfully")
	return nil
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
