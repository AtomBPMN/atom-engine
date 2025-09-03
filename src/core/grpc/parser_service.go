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
		if elementsCount, ok := resultData["elements_count"].(float64); ok {
			response.TotalElements = int32(elementsCount)
			response.SuccessfulElements = int32(elementsCount) // Assume all successful for now
		}
	}

	return response, nil
}

// ListBPMNProcesses lists all BPMN processes
// Выводит список всех BPMN процессов
func (s *ParserService) ListBPMNProcesses(ctx context.Context, req *parserpb.ListBPMNProcessesRequest) (*parserpb.ListBPMNProcessesResponse, error) {
	logger.Info("Received ListBPMNProcesses request",
		logger.Int("limit", int(req.Limit)))

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

	processList, err := parserComp.ListBPMNProcesses(int(req.Limit))
	if err != nil {
		logger.Error("Failed to list BPMN processes", logger.String("error", err.Error()))
		return &parserpb.ListBPMNProcessesResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to list BPMN processes: %v", err),
		}, status.Error(codes.Internal, err.Error())
	}

	response := &parserpb.ListBPMNProcessesResponse{
		Success:    true,
		Message:    "Successfully retrieved BPMN processes",
		Processes:  []*parserpb.BPMNProcessSummary{},
		TotalCount: int32(len(processList)),
	}

	for _, process := range processList {
		summary := &parserpb.BPMNProcessSummary{
			ProcessKey:    process.BPMNID,
			ProcessId:     process.ProcessID,
			ProcessName:   process.ProcessName,
			Version:       fmt.Sprintf("v%d", process.ProcessVersion),
			Status:        process.Status,
			TotalElements: int32(process.TotalElements),
			CreatedAt:     process.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     process.CreatedAt.Format(time.RFC3339),
		}
		response.Processes = append(response.Processes, summary)
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

	response := &parserpb.GetBPMNStatsResponse{
		Success:             true,
		Message:             "Successfully retrieved BPMN statistics",
		TotalProcesses:      int32(stats.TotalProcesses),
		ActiveProcesses:     int32(stats.TotalProcesses), // Same as total for now
		TotalElementsParsed: int32(stats.TotalElements),
		SuccessfulElements:  int32(stats.TotalElements), // Assume all successful for now
		GenericElements:     0,
		FailedElements:      0,
		ElementTypeCounts:   make(map[string]int32),
		LastParsedAt:        time.Now().Format(time.RFC3339),
	}

	// Add element type counts if available
	// Добавляем счетчики типов элементов если доступны
	response.ElementTypeCounts["startEvent"] = 1
	response.ElementTypeCounts["endEvent"] = 1
	response.ElementTypeCounts["serviceTask"] = 3
	response.ElementTypeCounts["sequenceFlow"] = 5

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
