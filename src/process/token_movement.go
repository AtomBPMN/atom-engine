/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/storage"
)

// TokenMovement handles token movement operations
// Обрабатывает операции перемещения токенов
type TokenMovement struct {
	storage            storage.Storage
	component          ComponentInterface
	bpmnHelper         *BPMNHelper
	executionProcessor *ExecutionProcessor
}

// NewTokenMovement creates new token movement helper
// Создает новый helper перемещения токенов
func NewTokenMovement(storage storage.Storage, component ComponentInterface) *TokenMovement {
	return &TokenMovement{
		storage:            storage,
		component:          component,
		bpmnHelper:         NewBPMNHelper(storage),
		executionProcessor: NewExecutionProcessor(storage, component),
	}
}

// MoveTokenToNextElements moves token to next elements using outgoing flows
// Перемещает токен к следующим элементам используя outgoing flows
func (tm *TokenMovement) MoveTokenToNextElements(token *models.Token, currentElementID string) error {
	// Load process elements
	elements, err := tm.bpmnHelper.LoadProcessElements(token.ProcessKey)
	if err != nil {
		return fmt.Errorf("failed to load process elements: %w", err)
	}

	// Get outgoing flows for current element
	outgoingFlows, err := tm.bpmnHelper.GetElementOutgoingFlows(elements, currentElementID)
	if err != nil {
		return fmt.Errorf("failed to get outgoing flows: %w", err)
	}

	if len(outgoingFlows) == 0 {
		// No outgoing flows - complete the token
		return tm.CompleteToken(token)
	}

	// Load full BPMN process for moveTokenToNextElements
	bpmnProcess, err := tm.bpmnHelper.LoadBPMNProcess(token.ProcessKey)
	if err != nil {
		return fmt.Errorf("failed to load BPMN process: %w", err)
	}

	// Use existing ExecutionProcessor logic for moving token
	return tm.executionProcessor.moveTokenToNextElements(token, outgoingFlows, bpmnProcess)
}

// CompleteToken completes token and cancels its boundary timers
// Завершает токен и отменяет его boundary таймеры
func (tm *TokenMovement) CompleteToken(token *models.Token) error {
	logger.Info("Completing token - no outgoing flows",
		logger.String("token_id", token.TokenID))

	// Cancel boundary timers for completing token
	if err := tm.component.CancelBoundaryTimersForToken(token.TokenID); err != nil {
		logger.Error("Failed to cancel boundary timers for completing token",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		// Continue - boundary timer cancellation is not critical
	}

	token.SetState(models.TokenStateCompleted)
	if err := tm.storage.UpdateToken(token); err != nil {
		return fmt.Errorf("failed to complete token: %w", err)
	}

	return nil
}
