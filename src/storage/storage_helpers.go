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

	"github.com/dgraph-io/badger/v3"
)

// Helper functions to reduce code duplication in storage operations
// Вспомогательные функции для уменьшения дублирования кода в storage операциях

// validateStorage checks if storage is ready and database is initialized
// Проверяет готовность storage и инициализацию базы данных
func (bs *BadgerStorage) validateStorage() error {
	if !bs.ready {
		return fmt.Errorf("storage is not ready")
	}
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return nil
}

// saveJSON saves JSON-serializable data with the given key
// Сохраняет JSON-сериализуемые данные с указанным ключом
func (bs *BadgerStorage) saveJSON(key string, data interface{}) error {
	if err := bs.validateStorage(); err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), jsonData)
	})
}

// loadJSON loads JSON data from storage and unmarshals into target
// Загружает JSON данные из storage и десериализует в target
func (bs *BadgerStorage) loadJSON(key string, target interface{}) error {
	if err := bs.validateStorage(); err != nil {
		return err
	}

	return bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("key not found: %s", key)
			}
			return fmt.Errorf("failed to get key %s: %w", key, err)
		}

		return item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, target); err != nil {
				return fmt.Errorf("failed to unmarshal data for key %s: %w", key, err)
			}
			return nil
		})
	})
}

// deleteKey deletes a key from storage
// Удаляет ключ из storage
func (bs *BadgerStorage) deleteKey(key string) error {
	if err := bs.validateStorage(); err != nil {
		return err
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// keyExists checks if a key exists in storage
// Проверяет существование ключа в storage
func (bs *BadgerStorage) keyExists(key string) (bool, error) {
	if err := bs.validateStorage(); err != nil {
		return false, err
	}

	exists := false
	err := bs.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Key doesn't exist, but no error
			}
			return err
		}
		exists = true
		return nil
	})

	return exists, err
}

// iterateWithPrefix iterates over all keys with the given prefix
// Итерирует по всем ключам с указанным префиксом
func (bs *BadgerStorage) iterateWithPrefix(prefix string, handler func(key []byte, value []byte) error) error {
	if err := bs.validateStorage(); err != nil {
		return err
	}

	return bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := item.Key()

			err := item.Value(func(val []byte) error {
				return handler(key, val)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
