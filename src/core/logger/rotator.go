/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"atom-engine/src/core/config"
)

// Rotator handles log file rotation
// Обрабатывает ротацию файлов логов
type Rotator struct {
	config   *config.LoggerConfig
	file     *os.File
	size     int64
	filename string
	cleaner  *Cleaner
	mu       sync.Mutex
}

// NewRotator creates new rotator instance
// Создает новый экземпляр ротатора
func NewRotator(cfg *config.LoggerConfig) (*Rotator, error) {
	filename := filepath.Join(cfg.Directory, "app.log")

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	rotator := &Rotator{
		config:   cfg,
		file:     file,
		size:     stat.Size(),
		filename: filename,
		cleaner:  NewCleaner(cfg),
	}

	// Clean old files on startup
	go rotator.cleaner.CleanOldFiles()

	return rotator, nil
}

// Write implements io.Writer interface
// Реализует интерфейс io.Writer
func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if rotation is needed by size
	if r.shouldRotateBySize(len(p)) {
		if err := r.rotate(); err != nil {
			return 0, fmt.Errorf("failed to rotate log: %w", err)
		}
	}

	n, err = r.file.Write(p)
	if err != nil {
		return n, err
	}

	r.size += int64(n)
	return n, nil
}

// shouldRotateBySize checks if file should be rotated based on size
// Проверяет, нужна ли ротация по размеру
func (r *Rotator) shouldRotateBySize(writeSize int) bool {
	maxSize := r.config.MaxSize * 1024 * 1024 // Convert MB to bytes
	return r.size+int64(writeSize) > maxSize
}

// rotate performs log file rotation
// Выполняет ротацию файла логов
func (r *Rotator) rotate() error {
	// Close current file
	if err := r.file.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("app-%s.log", timestamp)
	backupPath := filepath.Join(r.config.Directory, backupName)

	// Rename current file to backup
	if err := os.Rename(r.filename, backupPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Create new log file
	file, err := os.OpenFile(r.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	r.file = file
	r.size = 0

	// Clean old backup files
	go r.cleaner.CleanOldFiles()

	return nil
}

// Close closes the rotator
// Закрывает ротатор
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		return r.file.Close()
	}
	return nil
}
