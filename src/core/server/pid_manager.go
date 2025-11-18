/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"
	"os"
	"strconv"

	"atom-engine/src/core/logger"
)

// createPIDFile creates a PID file for the current process
// Создает PID файл для текущего процесса
func (c *Core) createPIDFile() error {
	pidPath := c.config.GetPIDFilePath()

	// Check if PID file already exists
	if _, err := os.Stat(pidPath); err == nil {
		// Read existing PID
		if data, readErr := os.ReadFile(pidPath); readErr == nil {
			if existingPID, parseErr := strconv.Atoi(string(data)); parseErr == nil {
				// Check if process is still running
				if process, findErr := os.FindProcess(existingPID); findErr == nil {
					if signalErr := process.Signal(os.Signal(nil)); signalErr == nil {
						return fmt.Errorf("instance '%s' is already running with PID %d", c.config.InstanceName, existingPID)
					}
				}
			}
		}
		// Remove stale PID file
		os.Remove(pidPath)
	}

	// Create new PID file
	pid := os.Getpid()
	err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	logger.Info("Created PID file", logger.String("path", pidPath), logger.Int("pid", pid))
	return nil
}

// removePIDFile removes the PID file
// Удаляет PID файл
func (c *Core) removePIDFile() error {
	pidPath := c.config.GetPIDFilePath()

	if _, err := os.Stat(pidPath); err == nil {
		if err := os.Remove(pidPath); err != nil {
			return fmt.Errorf("failed to remove PID file: %w", err)
		}
		logger.Info("Removed PID file", logger.String("path", pidPath))
	}

	return nil
}
