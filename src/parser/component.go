/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"atom-engine/src/core/config"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// Component represents BPMN parser component
// Компонент парсера BPMN
type Component struct {
	config          *config.Config
	storage         storage.Storage
	parser          *BPMNParser
	ready           bool
	responseChannel chan string
}

// NewComponent creates new parser component
// Создает новый компонент парсера
func NewComponent(cfg *config.Config, storage storage.Storage) *Component {
	return &Component{
		config:          cfg,
		storage:         storage,
		parser:          NewBPMNParser(),
		ready:           false,
		responseChannel: make(chan string, 100), // Buffered channel for parser responses
	}
}

// Init initializes parser component
// Инициализирует компонент парсера
func (c *Component) Init() error {
	logger.Info("Initializing BPMN parser component...")

	// Component is ready after initialization
	// Компонент готов после инициализации
	c.ready = true
	logger.Info("BPMN parser component initialized")
	return nil
}

// Start starts parser component
// Запускает компонент парсера
func (c *Component) Start() error {
	if !c.ready {
		return fmt.Errorf("parser component not initialized")
	}

	logger.Info("Starting BPMN parser component...")
	logger.Info("BPMN parser component is ready")
	return nil
}

// Stop stops parser component
// Останавливает компонент парсера
func (c *Component) Stop() error {
	logger.Info("Stopping BPMN parser component...")
	c.ready = false
	logger.Info("BPMN parser component stopped")
	return nil
}

// IsReady returns parser component ready status
// Возвращает статус готовности компонента парсера
func (c *Component) IsReady() bool {
	return c.ready
}

// ParseBPMNFile parses BPMN file and saves to storage
// Парсит BPMN файл и сохраняет в storage
func (c *Component) ParseBPMNFile(filePath, processID string, force bool) (*ParseResult, error) {
	if !c.ready {
		return nil, fmt.Errorf("parser component not ready")
	}

	// Check if file exists
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	logger.Info("Parsing BPMN file",
		logger.String("file", filePath),
		logger.String("process_id", processID),
		logger.Bool("force", force))

	// Parse BPMN file
	// Парсинг BPMN файла
	bpmnProcess, err := c.parser.ParseBPMNFile(filePath, processID, force)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BPMN file: %w", err)
	}

	// Read original file content for storage
	// Чтение оригинального содержимого файла для хранения
	originalContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file content: %w", err)
	}

	// Set additional metadata
	// Установка дополнительных метаданных
	bpmnProcess.ParsedAt = time.Now()
	bpmnProcess.Status = "active"

	// Determine correct version number based on existing processes
	// Определяем правильный номер версии на основе существующих процессов
	maxVersion, err := c.storage.GetMaxProcessVersionByProcessID(bpmnProcess.ProcessID)
	if err != nil {
		logger.Warn("Failed to get max version for process",
			logger.String("process_id", bpmnProcess.ProcessID),
			logger.String("error", err.Error()))
		bpmnProcess.ProcessVersion = 1 // Fallback to version 1
	} else {
		bpmnProcess.ProcessVersion = maxVersion + 1 // Increment version
	}

	logger.Info("Set process version",
		logger.String("process_id", bpmnProcess.ProcessID),
		logger.Int("version", bpmnProcess.ProcessVersion),
		logger.Int("previous_max_version", maxVersion))

	// Convert to JSON for storage
	// Конвертация в JSON для хранения
	jsonData, err := bpmnProcess.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Save to storage using processID:v{version} format
	// Сохранение в storage с форматом processID:v{version}
	storageKey := fmt.Sprintf("%s:v%d", bpmnProcess.ProcessID, bpmnProcess.ProcessVersion)
	err = c.storage.SaveBPMNProcess(storageKey, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to save BPMN process to storage: %w", err)
	}

	// Save original file to filesystem (configured directory)
	// Сохранение оригинального файла в файловую систему (настроенная директория)
	err = c.saveOriginalFile(bpmnProcess, originalContent)
	if err != nil {
		logger.Warn("Failed to save original file", logger.String("error", err.Error()))
		// Continue execution, this is not critical
		// Продолжаем выполнение, это не критично
	}

	// Save JSON file to filesystem (configured directory)
	// Сохранение JSON файла в файловую систему (настроенная директория)
	err = c.saveJSONFile(bpmnProcess, jsonData)
	if err != nil {
		logger.Warn("Failed to save JSON file", logger.String("error", err.Error()))
		// Continue execution, this is not critical
		// Продолжаем выполнение, это не критично
	}

	// Log successful parsing
	// Логирование успешного парсинга
	err = c.storage.LogSystemEvent(models.EventTypeBPMNParse, models.StatusSuccess,
		fmt.Sprintf("Successfully parsed BPMN file: %s -> %s", filePath, bpmnProcess.BPMNID))
	if err != nil {
		logger.Warn("Failed to log parse event", logger.String("error", err.Error()))
	}

	logger.Info("Successfully parsed and saved BPMN file",
		logger.String("bpmn_id", bpmnProcess.BPMNID),
		logger.String("process_id", bpmnProcess.ProcessID),
		logger.Int("total_elements", bpmnProcess.GetTotalElements()))

	return &ParseResult{
		BPMNID:        bpmnProcess.BPMNID,
		ProcessID:     bpmnProcess.ProcessID,
		ProcessName:   bpmnProcess.ProcessName,
		TotalElements: bpmnProcess.GetTotalElements(),
		ElementCounts: bpmnProcess.ElementCounts,
		ParsedAt:      bpmnProcess.ParsedAt,
		Success:       true,
	}, nil
}

// ListBPMNProcesses returns list of all BPMN processes
// Возвращает список всех BPMN процессов
func (c *Component) ListBPMNProcesses(limit int) ([]*ProcessInfo, error) {
	if !c.ready {
		return nil, fmt.Errorf("parser component not ready")
	}

	// Load all processes from storage
	// Загрузка всех процессов из storage
	allProcesses, err := c.storage.LoadAllBPMNProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN processes: %w", err)
	}

	processes := make([]*ProcessInfo, 0)
	count := 0

	for processKey, jsonData := range allProcesses {
		// Apply limit if specified
		// Применение лимита если указан
		if limit > 0 && count >= limit {
			break
		}

		// Parse JSON data to get process info
		// Парсинг JSON данных для получения информации о процессе
		var bpmnProcess models.BPMNProcess
		err := bpmnProcess.FromJSON(jsonData)
		if err != nil {
			logger.Warn("Failed to parse BPMN process data",
				logger.String("process_key", processKey),
				logger.String("error", err.Error()))
			continue
		}

		processes = append(processes, &ProcessInfo{
			BPMNID:         bpmnProcess.BPMNID,
			ProcessID:      bpmnProcess.ProcessID,
			ProcessName:    bpmnProcess.ProcessName,
			Version:        bpmnProcess.Version,
			ProcessVersion: bpmnProcess.ProcessVersion,
			Status:         bpmnProcess.Status,
			TotalElements:  bpmnProcess.GetTotalElements(),
			ParsedAt:       bpmnProcess.ParsedAt,
			CreatedAt:      bpmnProcess.CreatedAt,
		})

		count++
	}

	return processes, nil
}

// GetBPMNProcessDetails returns detailed information about BPMN process
// Возвращает подробную информацию о BPMN процессе
func (c *Component) GetBPMNProcessDetails(processKey string) (*models.BPMNProcess, error) {
	if !c.ready {
		return nil, fmt.Errorf("parser component not ready")
	}

	// Try to load process by BPMN ID first
	// Пытаемся сначала загрузить процесс по BPMN ID
	jsonData, err := c.storage.LoadBPMNProcessByBPMNID(processKey)
	if err != nil {
		// If not found by BPMN ID, try loading by storage key (backward compatibility)
		// Если не найден по BPMN ID, пытаемся загрузить по ключу storage (обратная совместимость)
		jsonData, err = c.storage.LoadBPMNProcess(processKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load BPMN process: %w", err)
		}
	}

	// Parse JSON data
	// Парсинг JSON данных
	var bpmnProcess models.BPMNProcess
	err = bpmnProcess.FromJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BPMN process data: %w", err)
	}

	return &bpmnProcess, nil
}

// GetBPMNProcessJSON returns JSON data for BPMN process
// Возвращает JSON данные для BPMN процесса
func (c *Component) GetBPMNProcessJSON(processKey string) ([]byte, error) {
	if !c.ready {
		return nil, fmt.Errorf("parser component not ready")
	}

	// Try to load process by BPMN ID first
	// Пытаемся сначала загрузить процесс по BPMN ID
	jsonData, err := c.storage.LoadBPMNProcessByBPMNID(processKey)
	if err != nil {
		// If not found by BPMN ID, try loading by storage key (backward compatibility)
		// Если не найден по BPMN ID, пытаемся загрузить по ключу storage (обратная совместимость)
		return c.storage.LoadBPMNProcess(processKey)
	}

	return jsonData, nil
}

// DeleteBPMNProcess deletes BPMN process
// Удаляет BPMN процесс
func (c *Component) DeleteBPMNProcess(processID string) error {
	if !c.ready {
		return fmt.Errorf("parser component not ready")
	}

	// Delete from storage
	// Удаление из storage
	err := c.storage.DeleteBPMNProcess(processID)
	if err != nil {
		return fmt.Errorf("failed to delete BPMN process: %w", err)
	}

	// Log deletion
	// Логирование удаления
	err = c.storage.LogSystemEvent(models.EventTypeBPMNDelete, models.StatusSuccess,
		fmt.Sprintf("Successfully deleted BPMN process: %s", processID))
	if err != nil {
		logger.Warn("Failed to log delete event", logger.String("error", err.Error()))
	}

	logger.Info("Successfully deleted BPMN process", logger.String("process_id", processID))
	return nil
}

// GetBPMNStats returns BPMN parser statistics
// Возвращает статистику парсера BPMN
func (c *Component) GetBPMNStats() (*BPMNStats, error) {
	if !c.ready {
		return nil, fmt.Errorf("parser component not ready")
	}

	// Load all processes to calculate stats
	// Загрузка всех процессов для подсчета статистики
	allProcesses, err := c.storage.LoadAllBPMNProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN processes for stats: %w", err)
	}

	stats := &BPMNStats{
		TotalProcesses: len(allProcesses),
		ElementCounts:  make(map[string]int),
		StatusCounts:   make(map[string]int),
	}

	// Calculate detailed statistics
	// Подсчет детальной статистики
	for _, jsonData := range allProcesses {
		var bpmnProcess models.BPMNProcess
		err := bpmnProcess.FromJSON(jsonData)
		if err != nil {
			continue // Skip corrupted data
		}

		// Count by status
		// Подсчет по статусу
		stats.StatusCounts[bpmnProcess.Status]++

		// Count elements by type
		// Подсчет элементов по типу
		for elementType, count := range bpmnProcess.ElementCounts {
			stats.ElementCounts[elementType] += count
		}

		// Track total elements
		// Отслеживание общего количества элементов
		stats.TotalElements += bpmnProcess.GetTotalElements()
	}

	return stats, nil
}

// getBPMNPath returns BPMN storage directory from configuration
// Возвращает директорию для хранения BPMN из конфигурации
func (c *Component) getBPMNPath() string {
	if c.config != nil && c.config.BPMN.Path != "" {
		// Use bpmn path from config relative to current working directory
		// Используем BPMN путь из конфигурации относительно текущей рабочей директории
		return c.config.BPMN.Path
	}
	// Fallback to bpmn_test for backward compatibility
	// Fallback на bpmn_test для обратной совместимости
	return "bpmn_test"
}

// saveOriginalFile saves original BPMN file to configured directory
// Сохраняет оригинальный BPMN файл в настроенную директорию
func (c *Component) saveOriginalFile(bpmnProcess *models.BPMNProcess, content []byte) error {
	bpmnPath := c.getBPMNPath()

	// Ensure BPMN directory exists
	// Убеждаемся что BPMN директория существует
	err := os.MkdirAll(bpmnPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create BPMN directory %s: %w", bpmnPath, err)
	}

	// Generate filename
	// Генерация имени файла
	filename := fmt.Sprintf("%s_v%d.bpmn", bpmnProcess.ProcessID, bpmnProcess.ProcessVersion)
	filePath := filepath.Join(bpmnPath, filename)

	// Write file
	// Запись файла
	err = ioutil.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to save original file: %w", err)
	}

	logger.Debug("Saved original BPMN file", logger.String("path", filePath))
	return nil
}

// saveJSONFile saves parsed JSON to configured directory
// Сохраняет спарсенный JSON в настроенную директорию
func (c *Component) saveJSONFile(bpmnProcess *models.BPMNProcess, jsonData []byte) error {
	bpmnPath := c.getBPMNPath()

	// Ensure BPMN directory exists
	// Убеждаемся что BPMN директория существует
	err := os.MkdirAll(bpmnPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create BPMN directory %s: %w", bpmnPath, err)
	}

	// Generate JSON filename
	// Генерация имени JSON файла
	filename := fmt.Sprintf("%s_v%d.json", bpmnProcess.ProcessID, bpmnProcess.ProcessVersion)
	filePath := filepath.Join(bpmnPath, filename)

	// Write JSON file
	// Запись JSON файла
	err = ioutil.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to save JSON file: %w", err)
	}

	logger.Debug("Saved parsed JSON file", logger.String("path", filePath))
	return nil
}

// ParseResult represents result of BPMN parsing operation
// Результат операции парсинга BPMN
type ParseResult struct {
	BPMNID        string         `json:"bpmn_id"`
	ProcessID     string         `json:"process_id"`
	ProcessName   string         `json:"process_name"`
	TotalElements int            `json:"total_elements"`
	ElementCounts map[string]int `json:"element_counts"`
	ParsedAt      time.Time      `json:"parsed_at"`
	Success       bool           `json:"success"`
}

// ProcessInfo represents brief information about BPMN process
// Краткая информация о BPMN процессе
type ProcessInfo struct {
	BPMNID         string    `json:"bpmn_id"`
	ProcessID      string    `json:"process_id"`
	ProcessName    string    `json:"process_name"`
	Version        string    `json:"version"`
	ProcessVersion int       `json:"process_version"`
	Status         string    `json:"status"`
	TotalElements  int       `json:"total_elements"`
	ParsedAt       time.Time `json:"parsed_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// BPMNStats represents BPMN parser statistics
// Статистика парсера BPMN
type BPMNStats struct {
	TotalProcesses int            `json:"total_processes"`
	TotalElements  int            `json:"total_elements"`
	ElementCounts  map[string]int `json:"element_counts"`
	StatusCounts   map[string]int `json:"status_counts"`
}

// ProcessMessage processes JSON message from core engine
// Обрабатывает JSON сообщение от core engine
func (c *Component) ProcessMessage(ctx context.Context, messageJSON string) error {
	if !c.IsReady() {
		return fmt.Errorf("parser component not ready")
	}

	// Parse message to determine type
	// Парсим сообщение для определения типа
	var request ParserRequest
	if err := json.Unmarshal([]byte(messageJSON), &request); err != nil {
		return fmt.Errorf("failed to parse parser request: %w", err)
	}

	logger.Debug("Processing parser request", logger.String("type", request.Type), logger.String("request_id", request.RequestID))

	switch request.Type {
	case "parse_bpmn_file":
		return c.handleParseBPMNFile(ctx, request)
	case "parse_bpmn_content":
		return c.handleParseBPMNContent(ctx, request)
	case "validate_bpmn":
		return c.handleValidateBPMN(ctx, request)
	case "get_process_info":
		return c.handleGetProcessInfo(ctx, request)
	case "list_processes":
		return c.handleListProcesses(ctx, request)
	case "delete_process":
		return c.handleDeleteProcess(ctx, request)
	case "get_stats":
		return c.handleGetStats(ctx, request)
	default:
		return fmt.Errorf("unknown parser request type: %s", request.Type)
	}
}

// handleParseBPMNFile handles BPMN file parsing request
// Обрабатывает запрос парсинга BPMN файла
func (c *Component) handleParseBPMNFile(ctx context.Context, request ParserRequest) error {
	var payload ParseBPMNFilePayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("parse_bpmn_file_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	result, err := c.ParseBPMNFile(payload.FilePath, payload.ProcessID, payload.Force)

	var response ParserResponse
	if err != nil {
		response = CreateParserErrorResponse("parse_bpmn_file_response", request.RequestID, err.Error())
	} else {
		parseResult := JSONParseResult{
			ProcessKey:     result.BPMNID,
			ProcessID:      result.ProcessID,
			ProcessVersion: 1, // TODO: Get actual version
			ElementsCount:  result.TotalElements,
			Success:        result.Success,
			Message:        "BPMN file parsed successfully",
			ProcessData:    map[string]interface{}{"element_counts": result.ElementCounts},
			Timestamp:      result.ParsedAt.Unix(),
		}
		response = CreateParserResponse("parse_bpmn_file_response", request.RequestID, parseResult)
	}

	return c.sendResponse(response)
}

// handleParseBPMNContent handles BPMN content parsing request
// Обрабатывает запрос парсинга содержимого BPMN
func (c *Component) handleParseBPMNContent(ctx context.Context, request ParserRequest) error {
	var payload ParseBPMNContentPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("parse_bpmn_content_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	// TODO: Implement ParseBPMNContent method in component
	// For now, return not implemented error
	response := CreateParserErrorResponse("parse_bpmn_content_response", request.RequestID, "parse_bpmn_content not implemented yet")
	return c.sendResponse(response)
}

// handleValidateBPMN handles BPMN validation request
// Обрабатывает запрос валидации BPMN
func (c *Component) handleValidateBPMN(ctx context.Context, request ParserRequest) error {
	var payload ValidateBPMNPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("validate_bpmn_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	// Simple validation - try to parse
	var err error
	if payload.FilePath != "" {
		_, err = c.ParseBPMNFile(payload.FilePath, "", false)
	} else {
		// TODO: Implement content validation
		err = fmt.Errorf("content validation not implemented yet")
	}

	validationResult := ValidationResult{
		Valid:   err == nil,
		Message: "BPMN validation completed",
	}

	if err != nil {
		validationResult.Errors = []string{err.Error()}
	}

	response := CreateParserResponse("validate_bpmn_response", request.RequestID, validationResult)
	return c.sendResponse(response)
}

// handleGetProcessInfo handles process info request
// Обрабатывает запрос информации о процессе
func (c *Component) handleGetProcessInfo(ctx context.Context, request ParserRequest) error {
	var payload GetProcessInfoPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("get_process_info_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	bpmnProcess, err := c.GetBPMNProcessDetails(payload.ProcessKey)

	var response ParserResponse
	if err != nil {
		response = CreateParserErrorResponse("get_process_info_response", request.RequestID, err.Error())
	} else {
		processInfo := ProcessInfoResult{
			ProcessKey:     bpmnProcess.BPMNID,
			ProcessID:      bpmnProcess.ProcessID,
			ProcessVersion: bpmnProcess.ProcessVersion,
			Name:           bpmnProcess.ProcessName,
			ElementsCount:  bpmnProcess.GetTotalElements(),
			Status:         bpmnProcess.Status,
			CreatedAt:      bpmnProcess.CreatedAt.Unix(),
			ParsedAt:       bpmnProcess.ParsedAt.Unix(),
			ProcessData:    map[string]interface{}{"element_counts": bpmnProcess.ElementCounts},
		}
		response = CreateParserResponse("get_process_info_response", request.RequestID, processInfo)
	}

	return c.sendResponse(response)
}

// handleListProcesses handles process listing request
// Обрабатывает запрос списка процессов
func (c *Component) handleListProcesses(ctx context.Context, request ParserRequest) error {
	var payload ListProcessesPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("list_processes_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	processes, err := c.ListBPMNProcesses(payload.Limit)

	var response ParserResponse
	if err != nil {
		response = CreateParserErrorResponse("list_processes_response", request.RequestID, err.Error())
	} else {
		// Convert to ProcessInfoResult slice
		processInfoResults := make([]ProcessInfoResult, len(processes))
		for i, p := range processes {
			processInfoResults[i] = ProcessInfoResult{
				ProcessKey:     p.BPMNID,
				ProcessID:      p.ProcessID,
				ProcessVersion: p.ProcessVersion,
				Name:           p.ProcessName,
				ElementsCount:  p.TotalElements,
				Status:         p.Status,
				CreatedAt:      p.CreatedAt.Unix(),
				ParsedAt:       p.ParsedAt.Unix(),
			}
		}

		listResult := ProcessListResult{
			Processes: processInfoResults,
			Total:     len(processInfoResults),
			Limit:     payload.Limit,
			Offset:    payload.Offset,
		}
		response = CreateParserResponse("list_processes_response", request.RequestID, listResult)
	}

	return c.sendResponse(response)
}

// handleDeleteProcess handles process deletion request
// Обрабатывает запрос удаления процесса
func (c *Component) handleDeleteProcess(ctx context.Context, request ParserRequest) error {
	var payload DeleteProcessPayload
	if err := mapToStruct(request.Payload, &payload); err != nil {
		response := CreateParserErrorResponse("delete_process_response", request.RequestID, fmt.Sprintf("invalid payload: %v", err))
		return c.sendResponse(response)
	}

	err := c.DeleteBPMNProcess(payload.ProcessID)

	var response ParserResponse
	if err != nil {
		response = CreateParserErrorResponse("delete_process_response", request.RequestID, err.Error())
	} else {
		deleteResult := DeleteResult{
			ProcessID: payload.ProcessID,
			Success:   true,
			Message:   "Process deleted successfully",
			Timestamp: time.Now().Unix(),
		}
		response = CreateParserResponse("delete_process_response", request.RequestID, deleteResult)
	}

	return c.sendResponse(response)
}

// handleGetStats handles get statistics request
// Обрабатывает запрос получения статистики
func (c *Component) handleGetStats(ctx context.Context, request ParserRequest) error {
	stats, err := c.GetBPMNStats()

	var response ParserResponse
	if err != nil {
		response = CreateParserErrorResponse("get_stats_response", request.RequestID, err.Error())
	} else {
		statsResult := ParserStatsResult{
			TotalProcesses:       stats.TotalProcesses,
			ActiveProcesses:      stats.StatusCounts["active"],
			ParsedToday:          0, // TODO: Implement today count
			LastParseTime:        0, // TODO: Implement last parse time
			AverageElementsCount: 0, // TODO: Calculate average
			ParseErrors:          0, // TODO: Track parse errors
		}
		if stats.TotalProcesses > 0 {
			statsResult.AverageElementsCount = stats.TotalElements / stats.TotalProcesses
		}
		response = CreateParserResponse("get_stats_response", request.RequestID, statsResult)
	}

	return c.sendResponse(response)
}

// sendResponse sends parser response through response channel
// Отправляет ответ парсера через канал ответов
func (c *Component) sendResponse(response ParserResponse) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal parser response", logger.String("error", err.Error()))
		return fmt.Errorf("failed to marshal parser response: %w", err)
	}

	logger.Info("Parser response",
		logger.String("type", response.Type),
		logger.String("request_id", response.RequestID),
		logger.Bool("success", response.Success))

	logger.Debug("Parser response JSON", logger.String("json", string(responseJSON)))

	if c.responseChannel != nil {
		select {
		case c.responseChannel <- string(responseJSON):
		default:
			logger.Warn("Parser response channel full, response dropped")
			return fmt.Errorf("parser response channel full")
		}
	}

	return nil
}

// GetResponseChannel returns the response channel for reading parser responses
// Возвращает канал ответов для чтения ответов парсера
func (c *Component) GetResponseChannel() <-chan string {
	return c.responseChannel
}
