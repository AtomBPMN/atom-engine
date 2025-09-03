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
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds application configuration
// Содержит конфигурацию приложения
type Config struct {
	InstanceName string         `yaml:"instance_name"` // Instance/deployment name
	BasePath     string         `yaml:"base_path"`     // Base path for all relative paths
	Database     DatabaseConfig `yaml:"database"`
	Server       ServerConfig   `yaml:"server"`
	GRPC         GRPCConfig     `yaml:"grpc"`
	RestAPI      RestAPIConfig  `yaml:"rest_api"`
	Logger       LoggerConfig   `yaml:"logger"`
	Storage      StorageConfig  `yaml:"storage"`
	BPMN         BPMNConfig     `yaml:"bpmn"`
}

// DatabaseConfig holds database configuration
// Конфигурация базы данных
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// ServerConfig holds server configuration
// Конфигурация сервера
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// GRPCConfig holds gRPC server configuration
// Конфигурация gRPC сервера
type GRPCConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// RestAPIConfig holds REST API server configuration
// Конфигурация REST API сервера
type RestAPIConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// StorageConfig holds storage configuration
// Конфигурация хранилища
type StorageConfig struct {
	Directory string                 `yaml:"directory"`
	Type      string                 `yaml:"type"` // badger, leveldb, etc
	Options   map[string]interface{} `yaml:"options"`
}

// LoggerConfig holds logger configuration
// Конфигурация логгера
type LoggerConfig struct {
	Level         string `yaml:"level"`
	Format        string `yaml:"format"`
	Directory     string `yaml:"directory"`
	MaxSize       int64  `yaml:"max_size"`       // Maximum size in MB
	MaxAge        int    `yaml:"max_age"`        // Maximum age in days
	MaxBackups    int    `yaml:"max_backups"`    // Maximum number of backup files
	EnableConsole bool   `yaml:"enable_console"` // Enable console output
}

// BPMNConfig holds BPMN parser configuration
// Конфигурация BPMN парсера
type BPMNConfig struct {
	Path            string `yaml:"path"`
	StorageOriginal bool   `yaml:"storage_original"`
	Validation      bool   `yaml:"validation"`
}

// LoadConfig loads configuration from YAML file
// Загружает конфигурацию из YAML файла
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set base path
	if config.BasePath == "" {
		config.BasePath = "."
	}

	// Apply defaults and resolve paths
	setDefaults(&config)

	// Override with environment variables
	config.LoadFromEnv()

	resolvePaths(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// GetPIDFilePath returns the path to the PID file
// Возвращает путь к PID файлу
func (c *Config) GetPIDFilePath() string {
	return filepath.Join(c.BasePath, c.InstanceName+".pid")
}

// setDefaults sets default values for configuration
// Устанавливает значения по умолчанию для конфигурации
func setDefaults(config *Config) {
	// Instance name default
	if config.InstanceName == "" {
		config.InstanceName = "atom-engine"
	}

	// Server defaults
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	// gRPC defaults
	if config.GRPC.Host == "" {
		config.GRPC.Host = "localhost"
	}
	if config.GRPC.Port == 0 {
		config.GRPC.Port = 9090
	}

	// REST API defaults
	if config.RestAPI.Host == "" {
		config.RestAPI.Host = "localhost"
	}
	if config.RestAPI.Port == 0 {
		config.RestAPI.Port = 8080
	}

	// Database defaults
	if config.Database.Path == "" {
		config.Database.Path = "data/badger"
	}

	// Storage defaults
	if config.Storage.Directory == "" {
		config.Storage.Directory = "storage"
	}
	if config.Storage.Type == "" {
		config.Storage.Type = "badger"
	}

	// Logger defaults
	if config.Logger.Level == "" {
		config.Logger.Level = "info"
	}
	if config.Logger.Format == "" {
		config.Logger.Format = "json"
	}
	if config.Logger.Directory == "" {
		config.Logger.Directory = "logs"
	}
	if config.Logger.MaxSize == 0 {
		config.Logger.MaxSize = 100 // 100MB default
	}
	if config.Logger.MaxAge == 0 {
		config.Logger.MaxAge = 30 // 30 days default
	}
	if config.Logger.MaxBackups == 0 {
		config.Logger.MaxBackups = 10 // 10 backup files default
	}

	// BPMN defaults
	if config.BPMN.Path == "" {
		config.BPMN.Path = "bpmn/"
	}
	// Set default values for BPMN config if not specified
	if !config.BPMN.StorageOriginal {
		config.BPMN.StorageOriginal = true // Default to true
	}
	if !config.BPMN.Validation {
		config.BPMN.Validation = true // Default to true
	}
}

// resolvePaths resolves relative paths based on base path
// Разрешает относительные пути на основе базового пути
func resolvePaths(config *Config) {
	// Resolve database path
	if !filepath.IsAbs(config.Database.Path) {
		config.Database.Path = filepath.Join(config.BasePath, config.Database.Path)
	}

	// Resolve storage directory
	if !filepath.IsAbs(config.Storage.Directory) {
		config.Storage.Directory = filepath.Join(config.BasePath, config.Storage.Directory)
	}

	// Resolve logger directory
	if !filepath.IsAbs(config.Logger.Directory) {
		config.Logger.Directory = filepath.Join(config.BasePath, config.Logger.Directory)
	}

	// Resolve BPMN directory
	if !filepath.IsAbs(config.BPMN.Path) {
		config.BPMN.Path = filepath.Join(config.BasePath, config.BPMN.Path)
	}
}
