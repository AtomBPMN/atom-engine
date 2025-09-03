/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/models"
)

// TokenManagerInterface handles token operations
// Интерфейс менеджера токенов
type TokenManagerInterface interface {
	// Token operations
	GetActiveTokens(instanceID string) ([]*models.Token, error)
	ExecuteToken(token *models.Token) error
}
