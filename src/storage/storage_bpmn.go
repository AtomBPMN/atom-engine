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
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
)

// BPMN storage key prefixes
// Префиксы ключей для BPMN storage
const (
	BPMNProcessPrefix = "bpmn:process:"
	BPMNFilePrefix    = "bpmn:file:"
)

// SaveBPMNProcess saves BPMN process data to storage
// Сохраняет данные BPMN процесса в storage
func (bs *BadgerStorage) SaveBPMNProcess(processID string, data []byte) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := BPMNProcessPrefix + processID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// LoadBPMNProcess loads BPMN process data from storage
// Загружает данные BPMN процесса из storage
func (bs *BadgerStorage) LoadBPMNProcess(processID string) ([]byte, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := BPMNProcessPrefix + processID
	var data []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			data = append([]byte(nil), val...)
			return nil
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("BPMN process not found: %s", processID)
		}
		return nil, fmt.Errorf("failed to load BPMN process: %w", err)
	}

	return data, nil
}

// LoadBPMNProcessByProcessID loads BPMN process by process_id and version
// Загружает BPMN процесс по process_id и версии
func (bs *BadgerStorage) LoadBPMNProcessByProcessID(processID string, version int) ([]byte, string, error) {
	if bs.db == nil {
		return nil, "", fmt.Errorf("database not initialized")
	}

	var foundData []byte
	var foundKey string
	var maxVersion int = -1
	var latestData []byte
	var latestKey string

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(BPMNProcessPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			// Read process data
			err := item.Value(func(val []byte) error {
				// Parse JSON to check process_id and version
				var processData map[string]interface{}
				if err := json.Unmarshal(val, &processData); err != nil {
					return nil // Skip invalid JSON, continue iteration
				}

				// Check process_id
				if procID, exists := processData["process_id"]; exists {
					if procIDStr, ok := procID.(string); ok && procIDStr == processID {
						// Check version
						if procVer, exists := processData["process_version"]; exists {
							if procVerFloat, ok := procVer.(float64); ok {
								currentVersion := int(procVerFloat)

								// If looking for specific version
								if version != -1 && currentVersion == version {
									// Found exact match
									foundData = append([]byte(nil), val...)
									foundKey = strings.TrimPrefix(string(key), BPMNProcessPrefix)
									return fmt.Errorf("found") // Use error to break iteration
								}

								// If looking for latest version (version = -1), track the highest version
								if version == -1 && currentVersion > maxVersion {
									maxVersion = currentVersion
									latestData = append([]byte(nil), val...)
									latestKey = strings.TrimPrefix(string(key), BPMNProcessPrefix)
								}
							}
						}
					}
				}
				return nil
			})
			if err != nil && err.Error() == "found" {
				return err
			}
		}
		return nil
	})

	// If exact version match found
	if err != nil && err.Error() == "found" {
		return foundData, foundKey, nil
	}

	// If looking for latest version and found at least one
	if version == -1 && maxVersion > -1 {
		return latestData, latestKey, nil
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to search BPMN processes: %w", err)
	}

	return nil, "", fmt.Errorf("BPMN process not found: process_id=%s, version=%d", processID, version)
}

// LoadBPMNProcessByBPMNID loads BPMN process by BPMN ID
// Загружает BPMN процесс по BPMN ID
func (bs *BadgerStorage) LoadBPMNProcessByBPMNID(bpmnID string) ([]byte, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var foundData []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(BPMNProcessPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// Read process data
			err := item.Value(func(val []byte) error {
				// Parse JSON to check BPMN ID
				var processData map[string]interface{}
				if err := json.Unmarshal(val, &processData); err != nil {
					return nil // Skip invalid JSON, continue iteration
				}

				// Check BPMN ID
				if procBPMNID, exists := processData["bpmn_id"]; exists {
					if procBPMNIDStr, ok := procBPMNID.(string); ok && procBPMNIDStr == bpmnID {
						// Found matching process
						foundData = append([]byte(nil), val...)
						return fmt.Errorf("found") // Use error to break iteration
					}
				}
				return nil
			})
			if err != nil && err.Error() == "found" {
				return err
			}
		}
		return nil
	})

	if err != nil && err.Error() == "found" {
		return foundData, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search BPMN processes: %w", err)
	}

	return nil, fmt.Errorf("BPMN process not found: %s", bpmnID)
}

// LoadAllBPMNProcesses loads all BPMN processes from storage
// Загружает все BPMN процессы из storage
func (bs *BadgerStorage) LoadAllBPMNProcesses() (map[string][]byte, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	processes := make(map[string][]byte)

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(BPMNProcessPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			// Extract process ID from key
			// Извлекаем ID процесса из ключа
			processID := strings.TrimPrefix(string(key), BPMNProcessPrefix)

			err := item.Value(func(val []byte) error {
				processes[processID] = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read BPMN process %s: %w", processID, err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load BPMN processes: %w", err)
	}

	return processes, nil
}

// DeleteBPMNProcess deletes BPMN process from storage
// Удаляет BPMN процесс из storage
func (bs *BadgerStorage) DeleteBPMNProcess(processID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := BPMNProcessPrefix + processID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// SaveBPMNFile saves original BPMN file content to storage
// Сохраняет содержимое оригинального BPMN файла в storage
func (bs *BadgerStorage) SaveBPMNFile(processID, filename string, content []byte) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create file record with proper timestamp
	// Создаем запись файла с правильным timestamp
	now := time.Now()
	fileRecord := map[string]interface{}{
		"process_id":    processID,
		"filename":      filename,
		"content":       content,
		"size":          len(content),
		"saved_at":      now.Format(time.RFC3339), // ISO 8601 timestamp
		"saved_at_unix": now.Unix(),               // Unix timestamp for easy querying
		"saved_at_nano": now.UnixNano(),           // Nanosecond precision for uniqueness
	}

	data, err := json.Marshal(fileRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal BPMN file record: %w", err)
	}

	key := BPMNFilePrefix + processID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// LoadBPMNFile loads original BPMN file content from storage
// Загружает содержимое оригинального BPMN файла из storage
func (bs *BadgerStorage) LoadBPMNFile(processID string) ([]byte, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := BPMNFilePrefix + processID
	var data []byte

	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			data = append([]byte(nil), val...)
			return nil
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("BPMN file not found: %s", processID)
		}
		return nil, fmt.Errorf("failed to load BPMN file: %w", err)
	}

	// Unmarshal file record and extract content
	// Демаршалим запись файла и извлекаем содержимое
	var fileRecord map[string]interface{}
	if err := json.Unmarshal(data, &fileRecord); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BPMN file record: %w", err)
	}

	contentInterface, exists := fileRecord["content"]
	if !exists {
		return nil, fmt.Errorf("no content found in BPMN file record")
	}

	// Handle different content formats
	// Обрабатываем разные форматы содержимого
	switch content := contentInterface.(type) {
	case []byte:
		return content, nil
	case string:
		return []byte(content), nil
	case []interface{}:
		// JSON array of bytes
		// JSON массив байтов
		bytes := make([]byte, len(content))
		for i, b := range content {
			if byteVal, ok := b.(float64); ok {
				bytes[i] = byte(byteVal)
			} else {
				return nil, fmt.Errorf("invalid byte value in content array")
			}
		}
		return bytes, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %T", content)
	}
}

// GetBPMNProcessStats returns statistics about BPMN processes in storage
// Возвращает статистику о BPMN процессах в storage
func (bs *BadgerStorage) GetBPMNProcessStats() (map[string]interface{}, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	stats := make(map[string]interface{})
	processCount := 0
	fileCount := 0

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // Only count keys
		it := txn.NewIterator(opts)
		defer it.Close()

		// Count BPMN processes
		// Считаем BPMN процессы
		processPrefix := []byte(BPMNProcessPrefix)
		for it.Seek(processPrefix); it.ValidForPrefix(processPrefix); it.Next() {
			processCount++
		}

		// Count BPMN files
		// Считаем BPMN файлы
		filePrefix := []byte(BPMNFilePrefix)
		for it.Seek(filePrefix); it.ValidForPrefix(filePrefix); it.Next() {
			fileCount++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get BPMN statistics: %w", err)
	}

	stats["process_count"] = processCount
	stats["file_count"] = fileCount
	stats["storage_ready"] = bs.ready

	return stats, nil
}

// GetMaxProcessVersionByProcessID finds highest version number for given ProcessID
// Находит максимальный номер версии для указанного ProcessID
func (bs *BadgerStorage) GetMaxProcessVersionByProcessID(processID string) (int, error) {
	if bs.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	maxVersion := 0

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(BPMNProcessPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				// Parse JSON to extract ProcessID and ProcessVersion
				// Парсим JSON чтобы извлечь ProcessID и ProcessVersion
				var process map[string]interface{}
				if err := json.Unmarshal(val, &process); err != nil {
					return nil // Skip invalid JSON
				}

				// Check if this process matches the ProcessID we're looking for
				// Проверяем совпадает ли ProcessID с тем что мы ищем
				if procID, exists := process["process_id"]; exists {
					if procIDStr, ok := procID.(string); ok && procIDStr == processID {
						// Extract ProcessVersion
						// Извлекаем ProcessVersion
						if version, exists := process["process_version"]; exists {
							if versionFloat, ok := version.(float64); ok {
								versionInt := int(versionFloat)
								if versionInt > maxVersion {
									maxVersion = versionInt
								}
							}
						}
					}
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
		return 0, fmt.Errorf("failed to search for process versions: %w", err)
	}

	return maxVersion, nil
}
