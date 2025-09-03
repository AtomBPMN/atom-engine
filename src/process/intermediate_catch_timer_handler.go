/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// IntermediateCatchTimerHandler handles timer events for intermediate catch events
// Обработчик timer событий для промежуточных catch событий
type IntermediateCatchTimerHandler struct {
	processComponent ComponentInterface
}

// NewIntermediateCatchTimerHandler creates new timer handler
// Создает новый обработчик timer событий
func NewIntermediateCatchTimerHandler(processComponent ComponentInterface) *IntermediateCatchTimerHandler {
	return &IntermediateCatchTimerHandler{
		processComponent: processComponent,
	}
}

// HandleTimerEvent handles timer intermediate catch events
// Обрабатывает timer промежуточные catch события
func (icth *IntermediateCatchTimerHandler) HandleTimerEvent(token *models.Token, element map[string]interface{}, eventDef map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Handling timer intermediate catch event",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Create timer and set token to waiting state
	// Создаем таймер и устанавливаем токен в состояние ожидания
	timerRequest := icth.createTimerRequest(token, eventDef)
	if timerRequest != nil {

		// Get outgoing flows for later continuation
		var nextElements []string
		if outgoing, exists := element["outgoing"]; exists {
			if outgoingList, ok := outgoing.([]interface{}); ok {
				for _, item := range outgoingList {
					if flowID, ok := item.(string); ok {
						nextElements = append(nextElements, flowID)
					}
				}
			} else if outgoingStr, ok := outgoing.(string); ok {
				nextElements = append(nextElements, outgoingStr)
			}
		}

		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: nextElements,
			WaitingFor:   "timer:" + token.CurrentElementID,
			Completed:    false,
			TimerRequest: timerRequest,
		}, nil
	}

	return &ExecutionResult{
		Success:   false,
		Error:     "failed to create timer request",
		Completed: false,
	}, nil
}

// createTimerRequest creates timer request from timer event definition
// Создает запрос таймера из определения timer event
func (icth *IntermediateCatchTimerHandler) createTimerRequest(token *models.Token, eventDef map[string]interface{}) *TimerRequest {
	timerData, exists := eventDef["timer_data"]
	if !exists {
		// Try legacy "timer" field
		timerData, exists = eventDef["timer"]
		if !exists {
			logger.Warn("Timer event definition missing timer_data",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", token.CurrentElementID))
			return nil
		}
	}

	timerMap, ok := timerData.(map[string]interface{})
	if !ok {
		logger.Warn("Invalid timer_data structure",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID))
		return nil
	}

	request := &TimerRequest{
		ElementID:         token.CurrentElementID,
		TokenID:           token.TokenID,
		ProcessInstanceID: token.ProcessInstanceID,
		ProcessKey:        token.ProcessKey,
	}

	// Extract timer definition based on type
	timerType, _ := timerMap["type"].(string)
	switch timerType {
	case "duration":
		if duration, exists := timerMap["duration"].(string); exists {
			request.TimeDuration = &duration
			logger.Debug("Timer duration extracted",
				logger.String("duration", duration))
		}
	case "date":
		if date, exists := timerMap["date"].(string); exists {
			request.TimeDate = &date
			logger.Debug("Timer date extracted",
				logger.String("date", date))
		}
	case "cycle":
		if cycle, exists := timerMap["cycle"].(string); exists {
			request.TimeCycle = &cycle
			logger.Debug("Timer cycle extracted",
				logger.String("cycle", cycle))
		}
	default:
		// Fallback - try to extract any available timer definition
		if duration, exists := timerMap["duration"].(string); exists {
			request.TimeDuration = &duration
			logger.Debug("Timer duration extracted (fallback)",
				logger.String("duration", duration))
		} else if date, exists := timerMap["date"].(string); exists {
			request.TimeDate = &date
			logger.Debug("Timer date extracted (fallback)",
				logger.String("date", date))
		} else if cycle, exists := timerMap["cycle"].(string); exists {
			request.TimeCycle = &cycle
			logger.Debug("Timer cycle extracted (fallback)",
				logger.String("cycle", cycle))
		} else {
			logger.Warn("No timer definition found in timer_data",
				logger.String("token_id", token.TokenID),
				logger.Any("timer_data", timerMap))
			return nil
		}
	}

	return request
}
