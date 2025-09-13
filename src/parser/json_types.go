/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

// ParserRequest base structure for all parser requests
// Базовая структура для всех запросов парсера
type ParserRequest struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
}

// ParserResponse base structure for all parser responses
// Базовая структура для всех ответов парсера
type ParserResponse struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// ParseBPMNFilePayload payload for parsing BPMN file
// Payload для парсинга BPMN файла
type ParseBPMNFilePayload struct {
	FilePath  string `json:"file_path"`
	ProcessID string `json:"process_id,omitempty"`
	Force     bool   `json:"force,omitempty"`
}

// ParseBPMNContentPayload payload for parsing BPMN content
// Payload для парсинга содержимого BPMN
type ParseBPMNContentPayload struct {
	BPMNContent string `json:"bpmn_content"`
	ProcessID   string `json:"process_id,omitempty"`
	Force       bool   `json:"force,omitempty"`
}

// ValidateBPMNPayload payload for validating BPMN
// Payload для валидации BPMN
type ValidateBPMNPayload struct {
	BPMNContent string `json:"bpmn_content,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
}

// GetProcessInfoPayload payload for getting process info
// Payload для получения информации о процессе
type GetProcessInfoPayload struct {
	ProcessKey string `json:"process_key"`
	Version    int    `json:"version,omitempty"`
}

// ListProcessesPayload payload for listing processes
// Payload для списка процессов
type ListProcessesPayload struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// DeleteProcessPayload payload for deleting process
// Payload для удаления процесса
type DeleteProcessPayload struct {
	ProcessID string `json:"process_id"`
}

// JSONParseResult result structure for JSON parse operations
// Структура результата для JSON операций парсинга
type JSONParseResult struct {
	ProcessKey       string                 `json:"process_key"`
	ProcessID        string                 `json:"process_id"`
	ProcessName      string                 `json:"process_name"`
	ProcessVersion   int                    `json:"process_version"`
	ElementsCount    int                    `json:"elements_count"`
	Success          bool                   `json:"success"`
	Message          string                 `json:"message,omitempty"`
	Warnings         []string               `json:"warnings,omitempty"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	ProcessData      map[string]interface{} `json:"process_data,omitempty"`
	Timestamp        int64                  `json:"timestamp,omitempty"`
}

// ValidationResult result structure for validation operations
// Структура результата для операций валидации
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Message  string   `json:"message,omitempty"`
}

// ProcessInfoResult result structure for process info operations
// Структура результата для операций информации о процессе
type ProcessInfoResult struct {
	ProcessKey     string                 `json:"process_key"`
	ProcessID      string                 `json:"process_id"`
	ProcessVersion int                    `json:"process_version"`
	Name           string                 `json:"name,omitempty"`
	ElementsCount  int                    `json:"elements_count"`
	Status         string                 `json:"status"`
	CreatedAt      int64                  `json:"created_at"`
	ParsedAt       int64                  `json:"parsed_at"`
	ProcessData    map[string]interface{} `json:"process_data,omitempty"`
}

// ProcessListResult result structure for process list operations
// Структура результата для операций списка процессов
type ProcessListResult struct {
	Processes []ProcessInfoResult `json:"processes"`
	Total     int                 `json:"total"`
	Limit     int                 `json:"limit"`
	Offset    int                 `json:"offset"`
}

// ParserStatsResult result structure for parser statistics
// Структура результата для статистики парсера
type ParserStatsResult struct {
	TotalProcesses  int `json:"total_processes"`
	ActiveProcesses int `json:"active_processes"`
	ParsedToday     int `json:"parsed_today"`
}

// DeleteResult result structure for delete operations
// Структура результата для операций удаления
type DeleteResult struct {
	ProcessID string `json:"process_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}
