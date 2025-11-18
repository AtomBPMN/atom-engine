/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import (
	"encoding/json"
	"time"
)

// GatewaySyncState tracks token synchronization state for parallel gateways
// Отслеживает состояние синхронизации токенов для параллельных шлюзов
type GatewaySyncState struct {
	ID                 string    `json:"id"`
	GatewayID          string    `json:"gateway_id"`
	ProcessInstanceID  string    `json:"process_instance_id"`
	ExpectedTokenCount int       `json:"expected_token_count"`
	ArrivedTokens      []string  `json:"arrived_tokens"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// NewGatewaySyncState creates new gateway synchronization state
// Создает новое состояние синхронизации шлюза
func NewGatewaySyncState(gatewayID, processInstanceID string, expectedCount int) *GatewaySyncState {
	now := time.Now()
	return &GatewaySyncState{
		ID:                 GenerateID(),
		GatewayID:          gatewayID,
		ProcessInstanceID:  processInstanceID,
		ExpectedTokenCount: expectedCount,
		ArrivedTokens:      make([]string, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// AddToken adds arrived token to synchronization state
// Добавляет пришедший токен в состояние синхронизации
func (gss *GatewaySyncState) AddToken(tokenID string) {
	gss.ArrivedTokens = append(gss.ArrivedTokens, tokenID)
	gss.UpdatedAt = time.Now()
}

// IsComplete checks if all expected tokens have arrived
// Проверяет, пришли ли все ожидаемые токены
func (gss *GatewaySyncState) IsComplete() bool {
	return len(gss.ArrivedTokens) >= gss.ExpectedTokenCount
}

// HasToken checks if token already arrived
// Проверяет, уже ли пришел токен
func (gss *GatewaySyncState) HasToken(tokenID string) bool {
	for _, arrivedToken := range gss.ArrivedTokens {
		if arrivedToken == tokenID {
			return true
		}
	}
	return false
}

// ToJSON serializes gateway sync state to JSON
// Сериализует состояние синхронизации шлюза в JSON
func (gss *GatewaySyncState) ToJSON() ([]byte, error) {
	return json.Marshal(gss)
}

// FromJSON deserializes gateway sync state from JSON
// Десериализует состояние синхронизации шлюза из JSON
func (gss *GatewaySyncState) FromJSON(data []byte) error {
	return json.Unmarshal(data, gss)
}
