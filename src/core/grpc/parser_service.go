/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"atom-engine/proto/parser/parserpb"
	"atom-engine/src/core/logger"
	"atom-engine/src/parser"
)

// ParserService implements BPMN parser gRPC service
// Реализует gRPC сервис парсера BPMN
type ParserService struct {
	parserpb.UnimplementedParserServiceServer
	core CoreInterface
}

// NewParserService creates new parser service
// Создает новый сервис парсера
func NewParserService(core CoreInterface) *ParserService {
	return &ParserService{
		core: core,
	}
}

// ParseBPMNFile parses BPMN file and saves to storage
// Парсит BPMN файл и сохраняет в хранилище
func (s *ParserService) ParseBPMNFile(ctx context.Context, req *parserpb.ParseBPMNFileRequest) (*parserpb.ParseBPMNFileResponse, error) {
	logger.Info("Received ParseBPMNFile request",
		logger.String("file_path", req.FilePath),
		logger.String("process_id", req.ProcessId),
		logger.Bool("force", req.Force))

	// Create JSON message for parser component
	payload := parser.ParseBPMNFilePayload{
		FilePath:  req.FilePath,
		ProcessID: req.ProcessId,
		Force:     req.Force,
	}

	message, err := parser.CreateParseBPMNFileMessage(payload)
	if err != nil {
		logger.Error("Failed to create parse BPMN file message", logger.String("error", err.Error()))
		return &parserpb.ParseBPMNFileResponse{
			Success: false,
			Message: fmt.Sprintf("failed to create parse message: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Send JSON message to parser component through Core
	if err := s.core.SendMessage("parser", message); err != nil {
		logger.Error("Failed to send parse BPMN file message", logger.String("error", err.Error()))
		return &parserpb.ParseBPMNFileResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Parse BPMN file request sent successfully")

	// Wait for response from parser component
	// Ожидаем ответ от компонента парсера
	responseJSON, err := s.core.WaitForParserResponse(10000) // 10 second timeout
	if err != nil {
		logger.Error("Failed to get parser response", logger.String("error", err.Error()))
		return &parserpb.ParseBPMNFileResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get parser response: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Parse JSON response
	// Парсим JSON ответ
	var parserResponse parser.ParserResponse
	if err := json.Unmarshal([]byte(responseJSON), &parserResponse); err != nil {
		logger.Error("Failed to parse parser response", logger.String("error", err.Error()))
		return &parserpb.ParseBPMNFileResponse{
			Success: false,
			Message: fmt.Sprintf("failed to parse response JSON: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Build response from parser data
	// Строим ответ из данных парсера
	response := &parserpb.ParseBPMNFileResponse{
		Success: parserResponse.Success,
		Message: "BPMN file processed successfully",
	}

	if !parserResponse.Success {
		response.Message = parserResponse.Error
		return response, nil
	}

	// Extract result data if parsing was successful
	// Извлекаем данные результата если парсинг был успешным
	if resultData, ok := parserResponse.Result.(map[string]interface{}); ok {
		if processKey, ok := resultData["process_key"].(string); ok {
			response.BpmnId = processKey
		}
		if processID, ok := resultData["process_id"].(string); ok {
			response.ProcessId = processID
		}
		if processName, ok := resultData["process_name"].(string); ok {
			response.ProcessName = processName
		}
		if elementsCount, ok := resultData["elements_count"].(float64); ok {
			response.TotalElements = int32(elementsCount)
			response.SuccessfulElements = int32(elementsCount) // Parser only saves successfully parsed elements
		}
	}

	return response, nil
}

// ListBPMNProcesses lists all BPMN processes
// Выводит список всех BPMN процессов
func (s *ParserService) ListBPMNProcesses(ctx context.Context, req *parserpb.ListBPMNProcessesRequest) (*parserpb.ListBPMNProcessesResponse, error) {
	// Set defaults for pagination and sorting parameters
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}
	page := req.Page
	if page <= 0 {
		page = 1 // Default page
	}
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at" // Default sort field
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC" // Default sort order
	}

	logger.Info("Received ListBPMNProcesses request",
		logger.Int("limit", int(req.Limit)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder))

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.ListBPMNProcessesResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.ListBPMNProcessesResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	// Get all processes for sorting/pagination
	processList, err := parserComp.ListBPMNProcesses(0)
	if err != nil {
		logger.Error("Failed to list BPMN processes", logger.String("error", err.Error()))
		return &parserpb.ListBPMNProcessesResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to list BPMN processes: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to protobuf format
	processes := make([]*parserpb.BPMNProcessSummary, len(processList))
	for i, process := range processList {
		processes[i] = &parserpb.BPMNProcessSummary{
			ProcessKey:    process.BPMNID,
			ProcessId:     process.ProcessID,
			ProcessName:   process.ProcessName,
			Version:       fmt.Sprintf("v%d", process.ProcessVersion),
			Status:        process.Status,
			TotalElements: int32(process.TotalElements),
			CreatedAt:     process.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     process.CreatedAt.Format(time.RFC3339),
		}
	}

	// Store total count before pagination
	totalCount := len(processes)

	// Apply sorting
	sort.Slice(processes, func(i, j int) bool {
		switch sortBy {
		case "created_at":
			if sortOrder == "ASC" {
				return processes[i].CreatedAt < processes[j].CreatedAt
			}
			return processes[i].CreatedAt > processes[j].CreatedAt
		case "updated_at":
			if sortOrder == "ASC" {
				return processes[i].UpdatedAt < processes[j].UpdatedAt
			}
			return processes[i].UpdatedAt > processes[j].UpdatedAt
		case "process_key":
			if sortOrder == "ASC" {
				return processes[i].ProcessKey < processes[j].ProcessKey
			}
			return processes[i].ProcessKey > processes[j].ProcessKey
		case "process_name":
			if sortOrder == "ASC" {
				return processes[i].ProcessName < processes[j].ProcessName
			}
			return processes[i].ProcessName > processes[j].ProcessName
		default:
			// Default to created_at DESC
			return processes[i].CreatedAt > processes[j].CreatedAt
		}
	})

	// Calculate pagination
	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)
	offset := (int(page) - 1) * int(pageSize)

	// Apply pagination
	var paginatedProcesses []*parserpb.BPMNProcessSummary
	if offset < len(processes) {
		end := offset + int(pageSize)
		if end > len(processes) {
			end = len(processes)
		}
		paginatedProcesses = processes[offset:end]
	}

	// Use paginated processes for new pagination system or legacy limit for old system
	var finalProcesses []*parserpb.BPMNProcessSummary
	if req.PageSize > 0 || (req.PageSize == 0 && req.Limit == 0) {
		// New pagination system (also default when no parameters specified)
		finalProcesses = paginatedProcesses
	} else if req.Limit > 0 && req.PageSize <= 0 {
		// Legacy limit system for backward compatibility
		if len(processes) > int(req.Limit) {
			finalProcesses = processes[:req.Limit]
			totalCount = len(finalProcesses)
			totalPages = 1
		} else {
			finalProcesses = processes
		}
	} else {
		finalProcesses = paginatedProcesses
	}

	logger.Info("BPMN processes listed successfully",
		logger.Int("count", len(finalProcesses)),
		logger.Int("total_count", totalCount),
		logger.Int("page", int(page)),
		logger.Int("page_size", int(pageSize)),
		logger.Int("total_pages", totalPages))

	response := &parserpb.ListBPMNProcessesResponse{
		Success:    true,
		Message:    "Successfully retrieved BPMN processes",
		Processes:  finalProcesses,
		TotalCount: int32(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int32(totalPages),
	}

	return response, nil
}

// GetBPMNProcess retrieves a specific BPMN process
// Получает конкретный BPMN процесс
func (s *ParserService) GetBPMNProcess(ctx context.Context, req *parserpb.GetBPMNProcessRequest) (*parserpb.GetBPMNProcessResponse, error) {
	logger.Info("Received GetBPMNProcess request",
		logger.String("process_key", req.ProcessKey))

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.GetBPMNProcessResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.GetBPMNProcessResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	processInfo, err := parserComp.GetBPMNProcessDetails(req.ProcessKey)
	if err != nil {
		logger.Error("Failed to get BPMN process",
			logger.String("process_key", req.ProcessKey),
			logger.String("error", err.Error()))
		return &parserpb.GetBPMNProcessResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get BPMN process: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	details := &parserpb.BPMNProcessDetails{
		ProcessKey:     processInfo.BPMNID,
		ProcessId:      processInfo.ProcessID,
		ProcessName:    processInfo.ProcessName,
		Version:        fmt.Sprintf("v%d", processInfo.ProcessVersion),
		ProcessVersion: int32(processInfo.ProcessVersion),
		Status:         processInfo.Status,
		TotalElements:  int32(processInfo.GetTotalElements()),
		ContentHash:    processInfo.ContentHash,
		OriginalFile:   processInfo.OriginalFile,
		CreatedAt:      processInfo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      processInfo.UpdatedAt.Format(time.RFC3339),
		ParsedAt:       processInfo.ParsedAt.Format(time.RFC3339),
		ElementCounts:  map[string]int32{},
	}

	// Element counts populated from process data
	// Счетчики элементов заполняются из данных процесса
	for elementType, count := range processInfo.ElementCounts {
		details.ElementCounts[elementType] = int32(count)
	}

	return &parserpb.GetBPMNProcessResponse{
		Success: true,
		Message: "Successfully retrieved BPMN process",
		Process: details,
	}, nil
}

// DeleteBPMNProcess deletes a BPMN process
// Удаляет BPMN процесс
func (s *ParserService) DeleteBPMNProcess(ctx context.Context, req *parserpb.DeleteBPMNProcessRequest) (*parserpb.DeleteBPMNProcessResponse, error) {
	logger.Info("Received DeleteBPMNProcess request",
		logger.String("process_id", req.ProcessId))

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.DeleteBPMNProcessResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.DeleteBPMNProcessResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	err := parserComp.DeleteBPMNProcess(req.ProcessId)
	if err != nil {
		logger.Error("Failed to delete BPMN process",
			logger.String("process_id", req.ProcessId),
			logger.String("error", err.Error()))
		return &parserpb.DeleteBPMNProcessResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete BPMN process: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	return &parserpb.DeleteBPMNProcessResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted BPMN process: %s", req.ProcessId),
	}, nil
}

// GetBPMNStats returns BPMN parsing statistics
// Возвращает статистику парсинга BPMN
func (s *ParserService) GetBPMNStats(ctx context.Context, req *parserpb.GetBPMNStatsRequest) (*parserpb.GetBPMNStatsResponse, error) {
	logger.Info("Received GetBPMNStats request")

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.GetBPMNStatsResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.GetBPMNStatsResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	stats, err := parserComp.GetBPMNStats()
	if err != nil {
		logger.Error("Failed to get BPMN stats", logger.String("error", err.Error()))
		return &parserpb.GetBPMNStatsResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get BPMN stats: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Get active processes count from status counts
	// Получаем количество активных процессов из счетчиков статусов
	activeProcesses := int32(0)
	if activeCount, exists := stats.StatusCounts["active"]; exists {
		activeProcesses = int32(activeCount)
	}

	response := &parserpb.GetBPMNStatsResponse{
		Success:             true,
		Message:             "Successfully retrieved BPMN statistics",
		TotalProcesses:      int32(stats.TotalProcesses),
		ActiveProcesses:     activeProcesses,
		TotalElementsParsed: int32(stats.TotalElements),
		SuccessfulElements:  int32(stats.TotalElements), // Parser tracks only successful parsing
		GenericElements:     0,                          // Not tracked separately
		FailedElements:      0,                          // Failed processes are not saved to storage
		ElementTypeCounts:   make(map[string]int32),
		LastParsedAt:        time.Now().Format(time.RFC3339),
	}

	// Add real element type counts from parser statistics
	// Добавляем реальные счетчики типов элементов из статистики парсера
	for elementType, count := range stats.ElementCounts {
		response.ElementTypeCounts[elementType] = int32(count)
	}

	return response, nil
}

// GetBPMNProcessJSON retrieves JSON representation of BPMN process
// Получает JSON представление BPMN процесса
func (s *ParserService) GetBPMNProcessJSON(ctx context.Context, req *parserpb.GetBPMNProcessJSONRequest) (*parserpb.GetBPMNProcessJSONResponse, error) {
	logger.Info("Received GetBPMNProcessJSON request",
		logger.String("process_key", req.ProcessKey))

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.GetBPMNProcessJSONResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.GetBPMNProcessJSONResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	jsonData, err := parserComp.GetBPMNProcessJSON(req.ProcessKey)
	if err != nil {
		logger.Error("Failed to get BPMN process JSON",
			logger.String("process_key", req.ProcessKey),
			logger.String("error", err.Error()))
		return &parserpb.GetBPMNProcessJSONResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get BPMN process JSON: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	return &parserpb.GetBPMNProcessJSONResponse{
		Success:  true,
		Message:  "Successfully retrieved BPMN process JSON",
		JsonData: string(jsonData),
	}, nil
}

// GetBPMNProcessXML retrieves original XML content of BPMN process
// Получает оригинальное XML содержимое BPMN процесса
func (s *ParserService) GetBPMNProcessXML(ctx context.Context, req *parserpb.GetBPMNProcessXMLRequest) (*parserpb.GetBPMNProcessXMLResponse, error) {
	logger.Info("Received GetBPMNProcessXML request",
		logger.String("process_key", req.ProcessKey))

	parserCompInterface := s.core.GetParserComponent()
	if parserCompInterface == nil {
		return &parserpb.GetBPMNProcessXMLResponse{
			Success: false,
			Message: "Parser component not available",
		}, status.Error(codes.Internal, "Parser component not available")
	}

	parserComp, ok := parserCompInterface.(*parser.Component)
	if !ok {
		return &parserpb.GetBPMNProcessXMLResponse{
			Success: false,
			Message: "Invalid parser component type",
		}, status.Error(codes.Internal, "Invalid parser component type")
	}

	xmlData, err := parserComp.GetBPMNProcessXML(req.ProcessKey)
	if err != nil {
		logger.Error("Failed to get BPMN process XML",
			logger.String("process_key", req.ProcessKey),
			logger.String("error", err.Error()))
		return &parserpb.GetBPMNProcessXMLResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get BPMN process XML: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	// Extract filename from process key for response
	// Извлекаем имя файла из ключа процесса для ответа
	filename := fmt.Sprintf("%s.bpmn", req.ProcessKey)

	logger.Info("Successfully retrieved BPMN process XML",
		logger.String("process_key", req.ProcessKey),
		logger.Int("file_size", len(xmlData)))

	return &parserpb.GetBPMNProcessXMLResponse{
		Success:  true,
		Message:  "Successfully retrieved BPMN process XML",
		XmlData:  string(xmlData),
		Filename: filename,
		FileSize: int32(len(xmlData)),
	}, nil
}
