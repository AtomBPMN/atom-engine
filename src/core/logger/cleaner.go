/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"atom-engine/src/core/config"
)

// Cleaner handles cleanup of old log files
// Обрабатывает очистку старых файлов логов
type Cleaner struct {
	config *config.LoggerConfig
}

// NewCleaner creates new cleaner instance
// Создает новый экземпляр очистителя
func NewCleaner(cfg *config.LoggerConfig) *Cleaner {
	return &Cleaner{config: cfg}
}

// CleanOldFiles removes old backup files
// Удаляет старые файлы резервных копий
func (c *Cleaner) CleanOldFiles() {
	c.cleanByCount()
	c.cleanByAge()
}

// cleanByCount removes files exceeding max backup count
// Удаляет файлы, превышающие максимальное количество резервных копий
func (c *Cleaner) cleanByCount() {
	files, err := c.getLogFiles()
	if err != nil {
		return
	}

	if len(files) <= c.config.MaxBackups {
		return
	}

	// Sort by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	// Remove oldest files
	filesToRemove := len(files) - c.config.MaxBackups
	for i := 0; i < filesToRemove; i++ {
		filepath := filepath.Join(c.config.Directory, files[i].Name())
		os.Remove(filepath)
	}
}

// cleanByAge removes files older than max age
// Удаляет файлы старше максимального возраста
func (c *Cleaner) cleanByAge() {
	if c.config.MaxAge <= 0 {
		return
	}

	files, err := c.getLogFiles()
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -c.config.MaxAge)

	for _, file := range files {
		if file.ModTime().Before(cutoff) {
			filepath := filepath.Join(c.config.Directory, file.Name())
			os.Remove(filepath)
		}
	}
}

// getLogFiles returns list of log files in directory
// Возвращает список файлов логов в директории
func (c *Cleaner) getLogFiles() ([]os.FileInfo, error) {
	entries, err := os.ReadDir(c.config.Directory)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Match log files (app.log, app-*.log)
		if name == "app.log" || (strings.HasPrefix(name, "app-") && strings.HasSuffix(name, ".log")) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			// Skip current active log file
			if name != "app.log" {
				files = append(files, info)
			}
		}
	}

	return files, nil
}
