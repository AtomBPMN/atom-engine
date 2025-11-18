/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
)

// IncidentStorageInterface represents incident-specific storage operations
// Представляет специфичные для инцидентов операции storage
type IncidentStorageInterface interface {
	SaveIncident(incident interface{}) error
	GetIncident(incidentID string) (interface{}, error)
	ListIncidents(filter interface{}) (interface{}, int, error)
}

// SaveIncident saves incident to storage
// Сохраняет инцидент в storage
func (bs *BadgerStorage) SaveIncident(incident interface{}) error {
	if !bs.ready {
		return fmt.Errorf("storage is not ready")
	}

	// Convert incident to map for generic handling
	incidentMap, ok := incident.(map[string]interface{})
	if !ok {
		// Try to marshal and unmarshal to convert
		data, err := json.Marshal(incident)
		if err != nil {
			return fmt.Errorf("failed to marshal incident: %w", err)
		}

		if err := json.Unmarshal(data, &incidentMap); err != nil {
			return fmt.Errorf("failed to unmarshal incident: %w", err)
		}
	}

	incidentID, exists := incidentMap["id"]
	if !exists {
		return fmt.Errorf("incident ID is required")
	}

	incidentIDStr, ok := incidentID.(string)
	if !ok {
		return fmt.Errorf("incident ID must be string")
	}

	// Generate key
	key := fmt.Sprintf("incident:%s", incidentIDStr)

	// Marshal incident data
	data, err := json.Marshal(incidentMap)
	if err != nil {
		return fmt.Errorf("failed to marshal incident data: %w", err)
	}

	// Save to database
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// GetIncident retrieves incident by ID
// Получает инцидент по ID
func (bs *BadgerStorage) GetIncident(incidentID string) (interface{}, error) {
	if !bs.ready {
		return nil, fmt.Errorf("storage is not ready")
	}

	if incidentID == "" {
		return nil, fmt.Errorf("incident ID is required")
	}

	key := fmt.Sprintf("incident:%s", incidentID)
	var incidentData map[string]interface{}

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Will return nil incident
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &incidentData)
		})
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}

	return incidentData, nil
}

// ListIncidents lists incidents with filtering
// Получает список инцидентов с фильтрацией
func (bs *BadgerStorage) ListIncidents(filter interface{}) (interface{}, int, error) {
	if !bs.ready {
		return nil, 0, fmt.Errorf("storage is not ready")
	}

	// Convert filter to map for generic handling
	var filterMap map[string]interface{}
	if filter != nil {
		var ok bool
		filterMap, ok = filter.(map[string]interface{})
		if !ok {
			// Try to marshal and unmarshal to convert
			data, err := json.Marshal(filter)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to marshal filter: %w", err)
			}

			if err := json.Unmarshal(data, &filterMap); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal filter: %w", err)
			}
		}
	}

	var incidents []map[string]interface{}
	prefix := []byte("incident:")

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var incident map[string]interface{}
				if err := json.Unmarshal(val, &incident); err != nil {
					return err
				}

				// Apply filters
				if bs.matchesIncidentFilter(incident, filterMap) {
					incidents = append(incidents, incident)
				}

				return nil
			})

			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list incidents: %w", err)
	}

	// Apply limit and offset
	total := len(incidents)
	limit := bs.getIntFromFilter(filterMap, "limit", 0)
	offset := bs.getIntFromFilter(filterMap, "offset", 0)

	if offset > 0 && offset < len(incidents) {
		incidents = incidents[offset:]
	} else if offset >= len(incidents) {
		incidents = []map[string]interface{}{}
	}

	if limit > 0 && limit < len(incidents) {
		incidents = incidents[:limit]
	}

	return incidents, total, nil
}

// matchesIncidentFilter checks if incident matches the filter
// Проверяет соответствует ли инцидент фильтру
func (bs *BadgerStorage) matchesIncidentFilter(incident map[string]interface{}, filter map[string]interface{}) bool {
	if filter == nil {
		return true
	}

	// Check status filter
	if statusFilter, exists := filter["status"]; exists {
		if statusArray, ok := statusFilter.([]interface{}); ok && len(statusArray) > 0 {
			incidentStatus, _ := incident["status"].(string)
			statusMatch := false
			for _, status := range statusArray {
				if statusStr, ok := status.(string); ok && statusStr == incidentStatus {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				return false
			}
		}
	}

	// Check type filter
	if typeFilter, exists := filter["type"]; exists {
		if typeArray, ok := typeFilter.([]interface{}); ok && len(typeArray) > 0 {
			incidentType, _ := incident["type"].(string)
			typeMatch := false
			for _, incType := range typeArray {
				if typeStr, ok := incType.(string); ok && typeStr == incidentType {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				return false
			}
		}
	}

	// Check process instance ID filter
	if processInstanceID, exists := filter["process_instance_id"]; exists {
		if processInstanceStr, ok := processInstanceID.(string); ok && processInstanceStr != "" {
			incidentProcessInstanceID, _ := incident["process_instance_id"].(string)
			if incidentProcessInstanceID != processInstanceStr {
				return false
			}
		}
	}

	// Check process key filter
	if processKey, exists := filter["process_key"]; exists {
		if processKeyStr, ok := processKey.(string); ok && processKeyStr != "" {
			incidentProcessKey, _ := incident["process_key"].(string)
			if incidentProcessKey != processKeyStr {
				return false
			}
		}
	}

	// Check element ID filter
	if elementID, exists := filter["element_id"]; exists {
		if elementIDStr, ok := elementID.(string); ok && elementIDStr != "" {
			incidentElementID, _ := incident["element_id"].(string)
			if incidentElementID != elementIDStr {
				return false
			}
		}
	}

	// Check job key filter
	if jobKey, exists := filter["job_key"]; exists {
		if jobKeyStr, ok := jobKey.(string); ok && jobKeyStr != "" {
			incidentJobKey, _ := incident["job_key"].(string)
			if incidentJobKey != jobKeyStr {
				return false
			}
		}
	}

	// Check worker ID filter
	if workerID, exists := filter["worker_id"]; exists {
		if workerIDStr, ok := workerID.(string); ok && workerIDStr != "" {
			incidentWorkerID, _ := incident["worker_id"].(string)
			if incidentWorkerID != workerIDStr {
				return false
			}
		}
	}

	// Check time filters
	if createdAfter, exists := filter["created_after"]; exists {
		if createdAfterStr, ok := createdAfter.(string); ok && createdAfterStr != "" {
			if afterTime, err := time.Parse(time.RFC3339, createdAfterStr); err == nil {
				if incidentCreatedAtStr, ok := incident["created_at"].(string); ok {
					if incidentCreatedAt, err := time.Parse(time.RFC3339, incidentCreatedAtStr); err == nil {
						if incidentCreatedAt.Before(afterTime) {
							return false
						}
					}
				}
			}
		}
	}

	if createdBefore, exists := filter["created_before"]; exists {
		if createdBeforeStr, ok := createdBefore.(string); ok && createdBeforeStr != "" {
			if beforeTime, err := time.Parse(time.RFC3339, createdBeforeStr); err == nil {
				if incidentCreatedAtStr, ok := incident["created_at"].(string); ok {
					if incidentCreatedAt, err := time.Parse(time.RFC3339, incidentCreatedAtStr); err == nil {
						if incidentCreatedAt.After(beforeTime) {
							return false
						}
					}
				}
			}
		}
	}

	return true
}

// getIntFromFilter safely gets int value from filter map
// Безопасно получает int значение из карты фильтра
func (bs *BadgerStorage) getIntFromFilter(filter map[string]interface{}, key string, defaultValue int) int {
	if filter == nil {
		return defaultValue
	}

	value, exists := filter[key]
	if !exists {
		return defaultValue
	}

	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}

	return defaultValue
}

// getStringFromFilter safely gets string value from filter map
// Безопасно получает string значение из карты фильтра
func (bs *BadgerStorage) getStringFromFilter(filter map[string]interface{}, key string, defaultValue string) string {
	if filter == nil {
		return defaultValue
	}

	value, exists := filter[key]
	if !exists {
		return defaultValue
	}

	if strVal, ok := value.(string); ok {
		return strVal
	}

	return defaultValue
}

// getStringArrayFromFilter safely gets string array from filter map
// Безопасно получает массив строк из карты фильтра
func (bs *BadgerStorage) getStringArrayFromFilter(filter map[string]interface{}, key string) []string {
	if filter == nil {
		return nil
	}

	value, exists := filter[key]
	if !exists {
		return nil
	}

	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case string:
		// Handle comma-separated values
		if strings.Contains(v, ",") {
			return strings.Split(v, ",")
		}
		return []string{v}
	}

	return nil
}
