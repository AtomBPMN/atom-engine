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

// Gateway synchronization storage key prefixes
// Префиксы ключей для Gateway synchronization storage
const (
	GatewaySyncPrefix = "gateway:sync:"
)

// SaveGatewaySyncState saves gateway synchronization state to storage
// Сохраняет состояние синхронизации шлюза в storage
func (bs *BadgerStorage) SaveGatewaySyncState(state *models.GatewaySyncState) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	data, err := state.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize gateway sync state: %w", err)
	}

	key := GatewaySyncPrefix + state.GatewayID + ":" + state.ProcessInstanceID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// LoadGatewaySyncState loads gateway synchronization state from storage
// Загружает состояние синхронизации шлюза из storage
func (bs *BadgerStorage) LoadGatewaySyncState(gatewayID, processInstanceID string) (*models.GatewaySyncState, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	key := GatewaySyncPrefix + gatewayID + ":" + processInstanceID
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
			return nil, nil // Not found is not an error for gateway sync state
		}
		return nil, fmt.Errorf("failed to load gateway sync state: %w", err)
	}

	var state models.GatewaySyncState
	if err := state.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to deserialize gateway sync state: %w", err)
	}

	return &state, nil
}

// DeleteGatewaySyncState deletes gateway synchronization state from storage
// Удаляет состояние синхронизации шлюза из storage
func (bs *BadgerStorage) DeleteGatewaySyncState(gatewayID, processInstanceID string) error {
	if bs.db == nil {
		return fmt.Errorf("database not initialized")
	}

	key := GatewaySyncPrefix + gatewayID + ":" + processInstanceID

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}
