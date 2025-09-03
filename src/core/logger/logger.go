/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"atom-engine/src/core/config"
)

// LogLevel represents logging level
// Уровень логирования
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns string representation of log level
// Возвращает строковое представление уровня логирования
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses string to LogLevel
// Парсит строку в LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// Logger represents the logging system
// Система логирования
type Logger struct {
	level     LogLevel
	formatter Formatter
	writer    io.Writer
	rotator   *Rotator
	config    *config.LoggerConfig
	mu        sync.Mutex
}

// New creates new logger instance
// Создает новый экземпляр логгера
func New(cfg *config.LoggerConfig) (*Logger, error) {
	// Create logs directory if not exists
	if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	rotator, err := NewRotator(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create rotator: %w", err)
	}

	var writer io.Writer = rotator
	if cfg.EnableConsole {
		writer = io.MultiWriter(os.Stdout, rotator)
	}

	formatter := NewFormatter(cfg.Format)

	logger := &Logger{
		level:     ParseLogLevel(cfg.Level),
		formatter: formatter,
		writer:    writer,
		rotator:   rotator,
		config:    cfg,
	}

	return logger, nil
}

// Debug logs debug message
// Логирует debug сообщение
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

// Info logs info message
// Логирует info сообщение
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

// Warn logs warning message
// Логирует предупреждающее сообщение
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

// Error logs error message
// Логирует сообщение об ошибке
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

// Fatal logs fatal message and exits
// Логирует критическое сообщение и завершает работу
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(FATAL, msg, fields...)
	os.Exit(1)
}

// log writes log entry
// Записывает лог
func (l *Logger) log(level LogLevel, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}

	formatted := l.formatter.Format(entry)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.writer.Write([]byte(formatted + "\n"))
}

// SetLevel sets logging level
// Устанавливает уровень логирования
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close closes the logger
// Закрывает логгер
func (l *Logger) Close() error {
	if l.rotator != nil {
		return l.rotator.Close()
	}
	return nil
}

// Field represents log field
// Поле лога
type Field struct {
	Key   string
	Value interface{}
}

// LogEntry represents single log entry
// Запись лога
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Fields    []Field
}
