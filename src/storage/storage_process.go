/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package storage

import (
	"fmt"

	"atom-engine/src/core/models"

	"github.com/dgraph-io/badger/v3"
)

// Process Instance storage key prefixes
// Префиксы ключей для Process Instance storage
const (
	ProcessInstancePrefix = "process:instance:"
	TokenPrefix           = "process:token:"
)

// SaveProcessInstance saves process instance to storage
// Сохраняет экземпляр процесса в storage
func (bs *BadgerStorage) SaveProcessInstance(instance *models.ProcessInstance) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := instance.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize process instance: %w", err)
	}

	key := ProcessInstancePrefix + instance.InstanceID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// LoadProcessInstance loads process instance from storage
// Загружает экземпляр процесса из storage
func (bs *BadgerStorage) LoadProcessInstance(instanceID string) (*models.ProcessInstance, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := ProcessInstancePrefix + instanceID
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
			return nil, fmt.Errorf("process instance not found: %s", instanceID)
		}
		return nil, fmt.Errorf("failed to load process instance: %w", err)
	}

	var instance models.ProcessInstance
	if err := instance.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to deserialize process instance: %w", err)
	}

	return &instance, nil
}

// LoadProcessInstancesByProcessKey loads all process instances for specific process key
// Загружает все экземпляры процессов для определенного ключа процесса
func (bs *BadgerStorage) LoadProcessInstancesByProcessKey(processKey string) ([]*models.ProcessInstance, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var instances []*models.ProcessInstance

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(ProcessInstancePrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read process instance data: %w", err)
			}

			var instance models.ProcessInstance
			if err := instance.FromJSON(data); err != nil {
				continue // Skip invalid entries
			}

			// Filter by process key
			if instance.ProcessKey == processKey {
				instances = append(instances, &instance)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load process instances by process key: %w", err)
	}

	return instances, nil
}

// LoadAllProcessInstances loads all process instances from storage
// Загружает все экземпляры процессов из storage
func (bs *BadgerStorage) LoadAllProcessInstances() ([]*models.ProcessInstance, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var instances []*models.ProcessInstance

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(ProcessInstancePrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read process instance data: %w", err)
			}

			var instance models.ProcessInstance
			if err := instance.FromJSON(data); err != nil {
				continue // Skip invalid entries
			}

			instances = append(instances, &instance)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load all process instances: %w", err)
	}

	return instances, nil
}

// UpdateProcessInstance updates existing process instance in storage
// Обновляет существующий экземпляр процесса в storage
func (bs *BadgerStorage) UpdateProcessInstance(instance *models.ProcessInstance) error {
	// Same as SaveProcessInstance for BadgerDB
	return bs.SaveProcessInstance(instance)
}

// DeleteProcessInstance deletes process instance from storage
// Удаляет экземпляр процесса из storage
func (bs *BadgerStorage) DeleteProcessInstance(instanceID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := ProcessInstancePrefix + instanceID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// SaveToken saves token to storage
// Сохраняет токен в storage
func (bs *BadgerStorage) SaveToken(token *models.Token) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := token.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize token: %w", err)
	}

	key := TokenPrefix + token.TokenID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// LoadToken loads token from storage
// Загружает токен из storage
func (bs *BadgerStorage) LoadToken(tokenID string) (*models.Token, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := TokenPrefix + tokenID
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
			return nil, fmt.Errorf("token not found: %s", tokenID)
		}
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	var token models.Token
	if err := token.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to deserialize token: %w", err)
	}

	return &token, nil
}

// LoadTokensByProcessInstance loads all tokens for specific process instance
// Загружает все токены для определенного экземпляра процесса
func (bs *BadgerStorage) LoadTokensByProcessInstance(processInstanceID string) ([]*models.Token, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var tokens []*models.Token

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(TokenPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read token data: %w", err)
			}

			var token models.Token
			if err := token.FromJSON(data); err != nil {
				continue // Skip invalid entries
			}

			// Filter by process instance ID
			if token.ProcessInstanceID == processInstanceID {
				tokens = append(tokens, &token)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load tokens by process instance: %w", err)
	}

	return tokens, nil
}

// LoadActiveTokens loads all active tokens from storage
// Загружает все активные токены из storage
func (bs *BadgerStorage) LoadActiveTokens() ([]*models.Token, error) {
	return bs.LoadTokensByState(models.TokenStateActive)
}

// LoadAllTokens loads all tokens from storage
// Загружает все токены из storage
func (bs *BadgerStorage) LoadAllTokens() ([]*models.Token, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var tokens []*models.Token

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(TokenPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read token data: %w", err)
			}

			var token models.Token
			if err := token.FromJSON(data); err != nil {
				continue // Skip invalid entries
			}

			tokens = append(tokens, &token)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load all tokens: %w", err)
	}

	return tokens, nil
}

// LoadTokensByState loads all tokens with specific state
// Загружает все токены с определенным состоянием
func (bs *BadgerStorage) LoadTokensByState(state models.TokenState) ([]*models.Token, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var tokens []*models.Token

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(TokenPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			var data []byte
			err := item.Value(func(val []byte) error {
				data = append([]byte(nil), val...)
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read token data: %w", err)
			}

			var token models.Token
			if err := token.FromJSON(data); err != nil {
				continue // Skip invalid entries
			}

			// Filter by state
			if token.State == state {
				tokens = append(tokens, &token)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load tokens by state: %w", err)
	}

	return tokens, nil
}

// UpdateToken updates existing token in storage
// Обновляет существующий токен в storage
func (bs *BadgerStorage) UpdateToken(token *models.Token) error {
	// Same as SaveToken for BadgerDB
	return bs.SaveToken(token)
}

// DeleteToken deletes token from storage
// Удаляет токен из storage
func (bs *BadgerStorage) DeleteToken(tokenID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := TokenPrefix + tokenID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
