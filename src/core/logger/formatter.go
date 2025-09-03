/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Formatter interface for log formatting
// Интерфейс для форматирования логов
type Formatter interface {
	Format(*LogEntry) string
}

// JSONFormatter formats logs as JSON
// Форматирует логи в формате JSON
type JSONFormatter struct{}

// TextFormatter formats logs as plain text
// Форматирует логи в текстовом формате
type TextFormatter struct{}

// NewFormatter creates formatter based on format type
// Создает форматтер на основе типа формата
func NewFormatter(format string) Formatter {
	switch strings.ToLower(format) {
	case "json":
		return &JSONFormatter{}
	case "text":
		return &TextFormatter{}
	default:
		return &JSONFormatter{}
	}
}

// Format implements Formatter interface for JSON
// Реализует интерфейс Formatter для JSON
func (f *JSONFormatter) Format(entry *LogEntry) string {
	data := map[string]interface{}{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	// Add fields
	for _, field := range entry.Fields {
		data[field.Key] = field.Value
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		// Fallback to simple format if JSON marshal fails
		return fmt.Sprintf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level.String(), entry.Message)
	}

	return string(bytes)
}

// Format implements Formatter interface for text
// Реализует интерфейс Formatter для текста
func (f *TextFormatter) Format(entry *LogEntry) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	level := fmt.Sprintf("%-5s", entry.Level.String())

	var fields strings.Builder
	for i, field := range entry.Fields {
		if i > 0 {
			fields.WriteString(" ")
		}
		fields.WriteString(fmt.Sprintf("%s=%v", field.Key, field.Value))
	}

	if fields.Len() > 0 {
		return fmt.Sprintf("%s [%s] %s | %s", timestamp, level, entry.Message, fields.String())
	}

	return fmt.Sprintf("%s [%s] %s", timestamp, level, entry.Message)
}
