/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package models

import "context"

// JSONMessageProcessor interface for components that can process JSON messages
// Интерфейс для компонентов, которые могут обрабатывать JSON сообщения
type JSONMessageProcessor interface {
	ProcessMessage(ctx context.Context, messageJSON string) error
}

// ComponentResponse represents a response from component JSON processing
// Представляет ответ от обработки JSON компонентом
type ComponentResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONMessageRouter interface for routing JSON messages to components
// Интерфейс для маршрутизации JSON сообщений к компонентам
type JSONMessageRouter interface {
	SendMessage(componentName, messageJSON string) error
	SendMessageWithResponse(ctx context.Context, componentName, messageJSON string) (*ComponentResponse, error)
}
