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
	GRPC         GRPCConfig     `yaml:"grpc"`
	RestAPI      RestAPIConfig  `yaml:"rest_api"`
	Logger       LoggerConfig   `yaml:"logger"`
	Storage      StorageConfig  `yaml:"storage"`
	BPMN         BPMNConfig     `yaml:"bpmn"`
	Auth         AuthConfig     `yaml:"auth"`
}

// DatabaseConfig holds database configuration
// Конфигурация базы данных
type DatabaseConfig struct {
	Path string `yaml:"path"`
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
	Directory string               `yaml:"directory"`
	Type      string               `yaml:"type"` // badger, leveldb, etc
	Options   StorageOptionsConfig `yaml:"options"`
}

// StorageOptionsConfig holds storage options
// Настройки опций хранилища
type StorageOptionsConfig struct {
	SyncWrites       *bool                    `yaml:"sync_writes,omitempty"`
	ValueLogFileSize *int64                   `yaml:"value_log_file_size,omitempty"`
	Performance      *BadgerPerformanceConfig `yaml:"performance,omitempty"`
}

// BadgerPerformanceConfig holds BadgerDB performance settings
// Настройки производительности BadgerDB
type BadgerPerformanceConfig struct {
	// Memory settings
	MemTableSize            *int64 `yaml:"mem_table_size,omitempty"`
	NumMemtables            *int   `yaml:"num_memtables,omitempty"`
	NumLevelZeroTables      *int   `yaml:"num_level_zero_tables,omitempty"`
	NumLevelZeroTablesStall *int   `yaml:"num_level_zero_tables_stall,omitempty"`

	// Cache settings
	ValueCacheSize *int64 `yaml:"value_cache_size,omitempty"`
	BlockCacheSize *int64 `yaml:"block_cache_size,omitempty"`
	IndexCacheSize *int64 `yaml:"index_cache_size,omitempty"`

	// Table and file settings
	BaseTableSize       *int64 `yaml:"base_table_size,omitempty"`
	MaxTableSize        *int64 `yaml:"max_table_size,omitempty"`
	LevelSizeMultiplier *int   `yaml:"level_size_multiplier,omitempty"`

	// Compaction settings
	NumCompactors    *int  `yaml:"num_compactors,omitempty"`
	CompactL0OnClose *bool `yaml:"compact_l0_on_close,omitempty"`

	// I/O settings
	TableLoadingMode    *string `yaml:"table_loading_mode,omitempty"`
	ValueLogLoadingMode *string `yaml:"value_log_loading_mode,omitempty"`

	// Advanced settings
	BloomFalsePositive *float64 `yaml:"bloom_false_positive,omitempty"`
	DetectConflicts    *bool    `yaml:"detect_conflicts,omitempty"`
	ManageTxns         *bool    `yaml:"manage_txns,omitempty"`

	// Batch processing
	MaxBatchCount *int   `yaml:"max_batch_count,omitempty"`
	MaxBatchSize  *int64 `yaml:"max_batch_size,omitempty"`
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

// AuthConfig holds auth configuration
// Конфигурация авторизации
type AuthConfig struct {
	Enabled      bool            `yaml:"enabled"`
	AllowedHosts []string        `yaml:"allowed_hosts"`
	APIKeys      []APIKeyConfig  `yaml:"api_keys"`
	RateLimit    RateLimitConfig `yaml:"rate_limiting"`
	Audit        AuditConfig     `yaml:"audit"`
}

// APIKeyConfig represents an API key configuration
type APIKeyConfig struct {
	Key          string   `yaml:"key"`
	Name         string   `yaml:"name"`
	Permissions  []string `yaml:"permissions"`
	AllowedHosts []string `yaml:"allowed_hosts,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
}

// AuditConfig represents audit logging configuration
type AuditConfig struct {
	Enabled           bool `yaml:"enabled"`
	LogFailedAttempts bool `yaml:"log_failed_attempts"`
	LogSuccessfulAuth bool `yaml:"log_successful_auth"`
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

	// gRPC defaults
	if config.GRPC.Host == "" {
		config.GRPC.Host = "localhost"
	}
	if config.GRPC.Port == 0 {
		config.GRPC.Port = 27500
	}

	// REST API defaults
	if config.RestAPI.Host == "" {
		config.RestAPI.Host = "localhost"
	}
	if config.RestAPI.Port == 0 {
		config.RestAPI.Port = 27555
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

	// Auth defaults
	// Auth is disabled by default for backward compatibility
	// Rate limiting defaults
	if config.Auth.RateLimit.RequestsPerMinute == 0 {
		config.Auth.RateLimit.RequestsPerMinute = 100 // Default 100 requests per minute
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
