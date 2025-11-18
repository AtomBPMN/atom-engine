/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package config

import (
	"fmt"
	"os"
	"strings"
)

// Validate validates the configuration
// Валидирует конфигурацию
func (c *Config) Validate() error {
	if err := c.validateBasePath(); err != nil {
		return fmt.Errorf("base_path validation failed: %w", err)
	}

	if err := c.validateGRPC(); err != nil {
		return fmt.Errorf("grpc validation failed: %w", err)
	}

	if err := c.validateRestAPI(); err != nil {
		return fmt.Errorf("rest_api validation failed: %w", err)
	}

	if err := c.validateDatabase(); err != nil {
		return fmt.Errorf("database validation failed: %w", err)
	}

	if err := c.validateStorage(); err != nil {
		return fmt.Errorf("storage validation failed: %w", err)
	}

	if err := c.validateLogger(); err != nil {
		return fmt.Errorf("logger validation failed: %w", err)
	}

	if err := c.validatePortConflicts(); err != nil {
		return fmt.Errorf("port conflicts detected: %w", err)
	}

	return nil
}

// validateBasePath validates base path
// Валидирует базовый путь
func (c *Config) validateBasePath() error {
	if c.BasePath == "" {
		return fmt.Errorf("base_path cannot be empty")
	}

	// Check if base path exists or can be created
	if _, err := os.Stat(c.BasePath); os.IsNotExist(err) {
		if err := os.MkdirAll(c.BasePath, 0755); err != nil {
			return fmt.Errorf("cannot create base path %s: %w", c.BasePath, err)
		}
	}

	return nil
}

// validateGRPC validates gRPC configuration
// Валидирует конфигурацию gRPC
func (c *Config) validateGRPC() error {
	if c.GRPC.Port < 1024 || c.GRPC.Port > 65535 {
		return fmt.Errorf("grpc port must be between 1024 and 65535, got %d", c.GRPC.Port)
	}

	if c.GRPC.Host == "" {
		return fmt.Errorf("grpc host cannot be empty")
	}

	return nil
}

// validateRestAPI validates REST API configuration
// Валидирует конфигурацию REST API
func (c *Config) validateRestAPI() error {
	if c.RestAPI.Port < 1024 || c.RestAPI.Port > 65535 {
		return fmt.Errorf("rest_api port must be between 1024 and 65535, got %d", c.RestAPI.Port)
	}

	if c.RestAPI.Host == "" {
		return fmt.Errorf("rest_api host cannot be empty")
	}

	return nil
}

// validateDatabase validates database configuration
// Валидирует конфигурацию базы данных
func (c *Config) validateDatabase() error {
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	return nil
}

// validateStorage validates storage configuration
// Валидирует конфигурацию хранилища
func (c *Config) validateStorage() error {
	if c.Storage.Directory == "" {
		return fmt.Errorf("storage directory cannot be empty")
	}

	validTypes := []string{"badger", "leveldb", "memory"}
	valid := false
	for _, vt := range validTypes {
		if c.Storage.Type == vt {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("storage type must be one of %v, got %s", validTypes, c.Storage.Type)
	}

	return nil
}

// validateLogger validates logger configuration
// Валидирует конфигурацию логгера
func (c *Config) validateLogger() error {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	valid := false
	for _, level := range validLevels {
		if strings.ToLower(c.Logger.Level) == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("logger level must be one of %v, got %s", validLevels, c.Logger.Level)
	}

	validFormats := []string{"json", "text"}
	valid = false
	for _, format := range validFormats {
		if strings.ToLower(c.Logger.Format) == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("logger format must be one of %v, got %s", validFormats, c.Logger.Format)
	}

	if c.Logger.Directory == "" {
		return fmt.Errorf("logger directory cannot be empty")
	}

	if c.Logger.MaxSize <= 0 {
		return fmt.Errorf("logger max_size must be positive, got %d", c.Logger.MaxSize)
	}

	if c.Logger.MaxAge <= 0 {
		return fmt.Errorf("logger max_age must be positive, got %d", c.Logger.MaxAge)
	}

	if c.Logger.MaxBackups <= 0 {
		return fmt.Errorf("logger max_backups must be positive, got %d", c.Logger.MaxBackups)
	}

	return nil
}

// validatePortConflicts checks for port conflicts
// Проверяет конфликты портов
func (c *Config) validatePortConflicts() error {
	ports := map[int]string{
		c.GRPC.Port:    "grpc",
		c.RestAPI.Port: "rest_api",
	}

	usedPorts := make(map[int][]string)
	for port, service := range ports {
		usedPorts[port] = append(usedPorts[port], service)
	}

	for port, services := range usedPorts {
		if len(services) > 1 {
			return fmt.Errorf("port %d is used by multiple services: %v", port, services)
		}
	}

	return nil
}
