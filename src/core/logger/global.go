/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

import (
	"sync"

	"atom-engine/src/core/config"
)

var (
	globalLogger *Logger
	once         sync.Once
)

// Init initializes global logger
// Инициализирует глобальный логгер
func Init(cfg *config.LoggerConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = New(cfg)
	})
	return err
}

// GetGlobal returns global logger instance
// Возвращает экземпляр глобального логгера
func GetGlobal() *Logger {
	return globalLogger
}

// Debug logs debug message using global logger
// Логирует debug сообщение через глобальный логгер
func Debug(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

// Info logs info message using global logger
// Логирует info сообщение через глобальный логгер
func Info(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

// Warn logs warning message using global logger
// Логирует предупреждающее сообщение через глобальный логгер
func Warn(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

// Error logs error message using global logger
// Логирует сообщение об ошибке через глобальный логгер
func Error(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// Fatal logs fatal message using global logger and exits
// Логирует критическое сообщение через глобальный логгер и завершает работу
func Fatal(msg string, fields ...Field) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	}
}

// Close closes global logger
// Закрывает глобальный логгер
func Close() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

// Helper functions for creating fields
// Вспомогательные функции для создания полей

// String creates string field
// Создает строковое поле
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates int field
// Создает числовое поле
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates int64 field
// Создает поле int64
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates float64 field
// Создает поле float64
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates bool field
// Создает булево поле
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates field with any value
// Создает поле с любым значением
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
