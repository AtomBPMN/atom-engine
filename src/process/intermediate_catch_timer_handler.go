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
func (icth *IntermediateCatchTimerHandler) HandleTimerEvent(
	token *models.Token,
	element map[string]interface{},
	eventDef map[string]interface{},
) (*ExecutionResult, error) {
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

		logger.Info("DEBUG: Timer event handler creating execution result",
			logger.String("element_id", token.CurrentElementID),
			logger.Int("next_elements_count", len(nextElements)),
			logger.Any("next_elements", nextElements))

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
func (icth *IntermediateCatchTimerHandler) createTimerRequest(
	token *models.Token,
	eventDef map[string]interface{},
) *TimerRequest {
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

	// Extract timer definition based on type with FEEL expression evaluation
	timerType, _ := timerMap["type"].(string)
	switch timerType {
	case "duration":
		if duration, exists := timerMap["duration"].(string); exists {
			evaluatedDuration, err := icth.evaluateTimerExpression(duration, token)
			if err != nil {
				logger.Error("Failed to evaluate timer duration expression",
					logger.String("token_id", token.TokenID),
					logger.String("expression", duration),
					logger.String("error", err.Error()))
				return nil
			}
			durationStr := fmt.Sprintf("%v", evaluatedDuration)
			request.TimeDuration = &durationStr
			logger.Debug("Timer duration evaluated and extracted",
				logger.String("original", duration),
				logger.String("evaluated", durationStr))
		}
	case "date":
		if date, exists := timerMap["date"].(string); exists {
			evaluatedDate, err := icth.evaluateTimerExpression(date, token)
			if err != nil {
				logger.Error("Failed to evaluate timer date expression",
					logger.String("token_id", token.TokenID),
					logger.String("expression", date),
					logger.String("error", err.Error()))
				return nil
			}
			dateStr := fmt.Sprintf("%v", evaluatedDate)
			request.TimeDate = &dateStr
			logger.Debug("Timer date evaluated and extracted",
				logger.String("original", date),
				logger.String("evaluated", dateStr))
		}
	case "cycle":
		if cycle, exists := timerMap["cycle"].(string); exists {
			evaluatedCycle, err := icth.evaluateTimerExpression(cycle, token)
			if err != nil {
				logger.Error("Failed to evaluate timer cycle expression",
					logger.String("token_id", token.TokenID),
					logger.String("expression", cycle),
					logger.String("error", err.Error()))
				return nil
			}
			cycleStr := fmt.Sprintf("%v", evaluatedCycle)
			request.TimeCycle = &cycleStr
			logger.Debug("Timer cycle evaluated and extracted",
				logger.String("original", cycle),
				logger.String("evaluated", cycleStr))
		}
	default:
		// Fallback - try to extract any available timer definition
		if duration, exists := timerMap["duration"].(string); exists {
			evaluatedDuration, err := icth.evaluateTimerExpression(duration, token)
			if err != nil {
				logger.Error("Failed to evaluate timer duration expression (fallback)",
					logger.String("token_id", token.TokenID),
					logger.String("expression", duration),
					logger.String("error", err.Error()))
				return nil
			}
			durationStr := fmt.Sprintf("%v", evaluatedDuration)
			request.TimeDuration = &durationStr
			logger.Debug("Timer duration evaluated and extracted (fallback)",
				logger.String("original", duration),
				logger.String("evaluated", durationStr))
		} else if date, exists := timerMap["date"].(string); exists {
			evaluatedDate, err := icth.evaluateTimerExpression(date, token)
			if err != nil {
				logger.Error("Failed to evaluate timer date expression (fallback)",
					logger.String("token_id", token.TokenID),
					logger.String("expression", date),
					logger.String("error", err.Error()))
				return nil
			}
			dateStr := fmt.Sprintf("%v", evaluatedDate)
			request.TimeDate = &dateStr
			logger.Debug("Timer date evaluated and extracted (fallback)",
				logger.String("original", date),
				logger.String("evaluated", dateStr))
		} else if cycle, exists := timerMap["cycle"].(string); exists {
			evaluatedCycle, err := icth.evaluateTimerExpression(cycle, token)
			if err != nil {
				logger.Error("Failed to evaluate timer cycle expression (fallback)",
					logger.String("token_id", token.TokenID),
					logger.String("expression", cycle),
					logger.String("error", err.Error()))
				return nil
			}
			cycleStr := fmt.Sprintf("%v", evaluatedCycle)
			request.TimeCycle = &cycleStr
			logger.Debug("Timer cycle evaluated and extracted (fallback)",
				logger.String("original", cycle),
				logger.String("evaluated", cycleStr))
		} else {
			logger.Warn("No timer definition found in timer_data",
				logger.String("token_id", token.TokenID),
				logger.Any("timer_data", timerMap))
			return nil
		}
	}

	return request
}

// evaluateTimerExpression evaluates timer expressions using expression component
// Вычисляет timer expressions используя expression компонент
func (icth *IntermediateCatchTimerHandler) evaluateTimerExpression(
	expression string,
	token *models.Token,
) (interface{}, error) {
	// If not a FEEL expression (doesn't start with =), return as is
	// Если не FEEL expression (не начинается с =), возвращаем как есть
	if expression == "" || len(expression) == 0 || expression[0] != '=' {
		return expression, nil
	}

	// Get expression component through process component
	// Получаем expression компонент через process компонент
	if icth.processComponent == nil {
		return nil, fmt.Errorf("process component not available for expression evaluation")
	}

	// Get core interface
	core := icth.processComponent.GetCore()
	if core == nil {
		return nil, fmt.Errorf("core interface not available for expression evaluation")
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		return nil, fmt.Errorf("expression component not available")
	}

	// Cast to expression evaluator interface with EvaluateExpressionEngine method
	// Приводим к интерфейсу expression evaluator с методом EvaluateExpressionEngine
	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		return nil, fmt.Errorf("failed to cast expression component to ExpressionEvaluator interface")
	}

	// Evaluate FEEL expression using expression engine
	// Вычисляем FEEL expression используя expression engine
	result, err := expressionComp.EvaluateExpressionEngine(expression, token.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate FEEL expression '%s': %w", expression, err)
	}

	logger.Debug("Timer expression evaluated successfully",
		logger.String("token_id", token.TokenID),
		logger.String("original_expression", expression),
		logger.Any("evaluated_result", result),
		logger.Any("token_variables", token.Variables))

	return result, nil
}
