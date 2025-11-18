/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"encoding/json"
	"fmt"

	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// BPMNHelper provides common BPMN process loading and parsing functionality
// Предоставляет общие функции загрузки и парсинга BPMN процессов
type BPMNHelper struct {
	storage storage.Storage
}

// NewBPMNHelper creates new BPMN helper
// Создает новый BPMN helper
func NewBPMNHelper(storage storage.Storage) *BPMNHelper {
	return &BPMNHelper{
		storage: storage,
	}
}

// LoadProcessElements loads and parses BPMN process, returns elements map
// Загружает и парсит BPMN процесс, возвращает карту элементов
func (bh *BPMNHelper) LoadProcessElements(processKey string) (map[string]interface{}, error) {
	processData, err := bh.storage.LoadBPMNProcess(processKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN process: %w", err)
	}

	var processJSON map[string]interface{}
	if err := json.Unmarshal(processData, &processJSON); err != nil {
		return nil, fmt.Errorf("failed to parse process JSON: %w", err)
	}

	elements, exists := processJSON["elements"].(map[string]interface{})
	if !exists {
		return nil, fmt.Errorf("no elements found in process definition")
	}

	return elements, nil
}

// LoadBPMNProcess loads and parses full BPMN process structure
// Загружает и парсит полную структуру BPMN процесса
func (bh *BPMNHelper) LoadBPMNProcess(processKey string) (*models.BPMNProcess, error) {
	processData, err := bh.storage.LoadBPMNProcess(processKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN process: %w", err)
	}

	var bpmnProcess models.BPMNProcess
	if err := json.Unmarshal(processData, &bpmnProcess); err != nil {
		return nil, fmt.Errorf("failed to parse BPMN process: %w", err)
	}

	return &bpmnProcess, nil
}

// GetElementOutgoingFlows gets outgoing sequence flows for element
// Получает исходящие sequence flows для элемента
func (bh *BPMNHelper) GetElementOutgoingFlows(elements map[string]interface{}, elementID string) ([]string, error) {
	element, exists := elements[elementID].(map[string]interface{})
	if !exists {
		return nil, fmt.Errorf("element %s not found in process definition", elementID)
	}

	outgoing, exists := element["outgoing"]
	if !exists {
		return []string{}, nil // No outgoing flows - element can be completed
	}

	var outgoingFlows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				outgoingFlows = append(outgoingFlows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		outgoingFlows = append(outgoingFlows, outgoingStr)
	}

	return outgoingFlows, nil
}

// GetBPMNProcessForToken loads BPMN process and returns interface map for token
// Compatible with existing Component.GetBPMNProcessForToken interface
// Загружает BPMN процесс и возвращает interface карту для токена
// Совместим с существующим интерфейсом Component.GetBPMNProcessForToken
func (bh *BPMNHelper) GetBPMNProcessForToken(processKey string) (map[string]interface{}, error) {
	bpmnProcess, err := bh.LoadBPMNProcess(processKey)
	if err != nil {
		return nil, err
	}

	// Return as map interface - same format as Component.GetBPMNProcessForToken
	processMap := map[string]interface{}{
		"elements":        bpmnProcess.Elements,
		"process_id":      bpmnProcess.ProcessID,
		"process_name":    bpmnProcess.ProcessName,
		"process_version": bpmnProcess.ProcessVersion,
		"is_executable":   bpmnProcess.IsExecutable,
	}

	return processMap, nil
}
