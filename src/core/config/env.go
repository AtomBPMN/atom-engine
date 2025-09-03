/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadFromEnv loads configuration from environment variables
// Загружает конфигурацию из переменных окружения
func (c *Config) LoadFromEnv() {
	// Instance name
	if env := os.Getenv("ATOM_INSTANCE_NAME"); env != "" {
		c.InstanceName = env
	}

	// Base path
	if env := os.Getenv("ATOM_BASE_PATH"); env != "" {
		c.BasePath = env
	}

	// Server configuration
	if env := os.Getenv("ATOM_SERVER_HOST"); env != "" {
		c.Server.Host = env
	}
	if env := os.Getenv("ATOM_SERVER_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			c.Server.Port = port
		}
	}

	// gRPC configuration
	if env := os.Getenv("ATOM_GRPC_HOST"); env != "" {
		c.GRPC.Host = env
	}
	if env := os.Getenv("ATOM_GRPC_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			c.GRPC.Port = port
		}
	}

	// REST API configuration
	if env := os.Getenv("ATOM_REST_API_HOST"); env != "" {
		c.RestAPI.Host = env
	}
	if env := os.Getenv("ATOM_REST_API_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			c.RestAPI.Port = port
		}
	}

	// Database configuration
	if env := os.Getenv("ATOM_DATABASE_PATH"); env != "" {
		c.Database.Path = env
	}

	// Storage configuration
	if env := os.Getenv("ATOM_STORAGE_DIRECTORY"); env != "" {
		c.Storage.Directory = env
	}
	if env := os.Getenv("ATOM_STORAGE_TYPE"); env != "" {
		c.Storage.Type = env
	}

	// Logger configuration
	if env := os.Getenv("ATOM_LOGGER_LEVEL"); env != "" {
		c.Logger.Level = strings.ToLower(env)
	}
	if env := os.Getenv("ATOM_LOGGER_FORMAT"); env != "" {
		c.Logger.Format = strings.ToLower(env)
	}
	if env := os.Getenv("ATOM_LOGGER_DIRECTORY"); env != "" {
		c.Logger.Directory = env
	}
	if env := os.Getenv("ATOM_LOGGER_MAX_SIZE"); env != "" {
		if size, err := strconv.ParseInt(env, 10, 64); err == nil {
			c.Logger.MaxSize = size
		}
	}
	if env := os.Getenv("ATOM_LOGGER_MAX_AGE"); env != "" {
		if age, err := strconv.Atoi(env); err == nil {
			c.Logger.MaxAge = age
		}
	}
	if env := os.Getenv("ATOM_LOGGER_MAX_BACKUPS"); env != "" {
		if backups, err := strconv.Atoi(env); err == nil {
			c.Logger.MaxBackups = backups
		}
	}
	if env := os.Getenv("ATOM_LOGGER_ENABLE_CONSOLE"); env != "" {
		c.Logger.EnableConsole = strings.ToLower(env) == "true"
	}
}

// GetConfigPath returns configuration file path from environment or default
// Возвращает путь к файлу конфигурации из окружения или по умолчанию
func GetConfigPath() string {
	if env := os.Getenv("ATOM_CONFIG_PATH"); env != "" {
		return env
	}
	return "build/config/config.yaml"
}

// LoadConfigWithEnv loads configuration with proper environment handling
// Загружает конфигурацию с правильной обработкой переменных окружения
func LoadConfigWithEnv() (*Config, error) {
	// Load .env file if exists
	envPath := "build/config/.env"
	if _, err := os.Stat(envPath); err == nil {
		if err := loadEnvFile(envPath); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Get config path
	configPath := GetConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Load config
	return LoadConfig(configPath)
}

// loadEnvFile loads environment variables from file
// Загружает переменные окружения из файла
func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// GetEnvWithDefault returns environment variable value or default
// Возвращает значение переменной окружения или значение по умолчанию
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt returns environment variable as integer or default
// Возвращает переменную окружения как число или значение по умолчанию
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool returns environment variable as boolean or default
// Возвращает переменную окружения как булево значение или значение по умолчанию
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}
