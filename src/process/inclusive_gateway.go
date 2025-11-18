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
	"atom-engine/src/expression"
)

// InclusiveGatewayExecutor executes inclusive gateways
// Исполнитель включающих шлюзов
type InclusiveGatewayExecutor struct {
	processComponent ComponentInterface
}

// NewInclusiveGatewayExecutor creates new inclusive gateway executor
// Создает новый исполнитель включающего шлюза
func NewInclusiveGatewayExecutor(processComponent ComponentInterface) *InclusiveGatewayExecutor {
	return &InclusiveGatewayExecutor{
		processComponent: processComponent,
	}
}

// Execute executes inclusive gateway
// Выполняет включающий шлюз
func (ige *InclusiveGatewayExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing inclusive gateway",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Get gateway name for logging
	gatewayName, _ := element["name"].(string)
	if gatewayName == "" {
		gatewayName = token.CurrentElementID
	}

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:   false,
			Error:     "inclusive gateway has no outgoing sequence flows",
			Completed: false,
		}, nil
	}

	var outgoingFlows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				outgoingFlows = append(outgoingFlows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		outgoingFlows = append(outgoingFlows, outgoingStr)
	}

	if len(outgoingFlows) == 0 {
		return &ExecutionResult{
			Success:   false,
			Error:     "inclusive gateway has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// For inclusive gateway, evaluate conditions and select matching flows
	// Для включающего шлюза оцениваем условия и выбираем подходящие потоки
	selectedFlows := ige.evaluateInclusiveGatewayConditions(outgoingFlows, token, element)

	logger.Info("Inclusive gateway executed",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.Int("selected_flows", len(selectedFlows)),
		logger.Int("total_flows", len(outgoingFlows)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: selectedFlows,
		Completed:    false,
	}, nil
}

// evaluateInclusiveGatewayConditions evaluates all conditions and returns matching flows
// Оценивает все условия и возвращает подходящие потоки
func (ige *InclusiveGatewayExecutor) evaluateInclusiveGatewayConditions(
	outgoingFlows []string,
	token *models.Token,
	element map[string]interface{},
) []string {
	var selectedFlows []string
	var defaultFlow string
	hasSelectedFlow := false

	// Create evaluation context from token variables
	// Создаем контекст оценки из переменных токена
	evaluationContext := make(map[string]interface{})
	if token.Variables != nil {
		for k, v := range token.Variables {
			evaluationContext[k] = v
		}
	}

	logger.Debug("Evaluating inclusive gateway conditions",
		logger.String("token_id", token.TokenID),
		logger.Int("outgoing_flows", len(outgoingFlows)),
		logger.Any("variables", evaluationContext))

	// Evaluate each outgoing flow
	// Оцениваем каждый исходящий поток
	for _, flowID := range outgoingFlows {
		logger.Debug("Evaluating flow for inclusive gateway",
			logger.String("token_id", token.TokenID),
			logger.String("flow_id", flowID))

		// Check if this flow has a condition
		// Проверяем есть ли у этого потока условие
		hasCondition, conditionResult := ige.evaluateFlowCondition(flowID, evaluationContext, element)

		if hasCondition {
			if conditionResult {
				selectedFlows = append(selectedFlows, flowID)
				hasSelectedFlow = true
				logger.Info("Flow condition evaluated to true - including in inclusive gateway",
					logger.String("token_id", token.TokenID),
					logger.String("flow_id", flowID))
			} else {
				logger.Debug("Flow condition evaluated to false - excluding from inclusive gateway",
					logger.String("token_id", token.TokenID),
					logger.String("flow_id", flowID))
			}
		} else {
			// Flow without condition - potential default flow
			// Поток без условия - потенциальный default поток
			if defaultFlow == "" {
				defaultFlow = flowID
				logger.Debug("Found potential default flow for inclusive gateway",
					logger.String("token_id", token.TokenID),
					logger.String("flow_id", flowID))
			}
		}
	}

	// For inclusive gateway: if no conditions are true, use default flow
	// Для включающего шлюза: если нет истинных условий, используем default поток
	if !hasSelectedFlow && defaultFlow != "" {
		selectedFlows = append(selectedFlows, defaultFlow)
		logger.Info("No conditions were true - using default flow for inclusive gateway",
			logger.String("token_id", token.TokenID),
			logger.String("default_flow", defaultFlow))
	}

	// Fallback: if no flows selected and no default, select all flows (traditional inclusive behavior)
	// Fallback: если никакие потоки не выбраны и нет default, выбираем все потоки (традиционное inclusive поведение)
	if len(selectedFlows) == 0 {
		selectedFlows = outgoingFlows
		logger.Warn("No conditions or default flow found - selecting all flows as fallback",
			logger.String("token_id", token.TokenID),
			logger.Int("selected_flows", len(selectedFlows)))
	}

	return selectedFlows
}

// evaluateFlowCondition evaluates condition for a specific flow
// Оценивает условие для определенного потока
func (ige *InclusiveGatewayExecutor) evaluateFlowCondition(
	flowID string,
	variables map[string]interface{},
	element map[string]interface{},
) (hasCondition bool, result bool) {
	// Get sequence flows from element
	// Получаем sequence flows из элемента
	sequenceFlows, exists := element["sequenceFlows"]
	if !exists {
		return false, false
	}

	sequenceFlowsMap, ok := sequenceFlows.(map[string]interface{})
	if !ok {
		return false, false
	}

	// Get specific flow
	// Получаем определенный поток
	flowData, exists := sequenceFlowsMap[flowID]
	if !exists {
		return false, false
	}

	flowMap, ok := flowData.(map[string]interface{})
	if !ok {
		return false, false
	}

	// Check for condition
	// Проверяем условие
	conditionData, hasCondition := flowMap["conditionExpression"]
	if !hasCondition {
		return false, false
	}

	conditionMap, ok := conditionData.(map[string]interface{})
	if !ok {
		return false, false
	}

	expression, ok := conditionMap["expression"].(string)
	if !ok || expression == "" {
		return false, false
	}

	// Evaluate condition using expression engine
	// Оцениваем условие используя expression engine
	result = ige.evaluateConditionWithExpressionEngine(expression, variables)
	return true, result
}

// evaluateConditionWithExpressionEngine evaluates condition using full expression engine
// Оценивает условие используя полноценный expression engine
func (ige *InclusiveGatewayExecutor) evaluateConditionWithExpressionEngine(
	condition string,
	variables map[string]interface{},
) bool {
	// Get expression component through core
	// Получаем expression компонент через core
	if ige.processComponent == nil {
		logger.Warn("Process component not available for expression evaluation")
		return false
	}

	// Get core interface
	core := ige.processComponent.GetCore()
	if core == nil {
		logger.Warn("Core interface not available for expression evaluation")
		return false
	}

	// Get expression component
	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		logger.Warn("Expression component not available")
		return false
	}

	// Cast to expression component interface
	expressionComp, ok := expressionCompInterface.(*expression.Component)
	if !ok {
		logger.Warn("Failed to cast expression component")
		return false
	}

	// Evaluate condition using expression engine
	// Оцениваем условие используя expression engine
	result, err := expressionComp.EvaluateCondition(variables, condition)
	if err != nil {
		logger.Warn("Failed to evaluate condition with expression engine",
			logger.String("condition", condition),
			logger.String("error", err.Error()))
		return false
	}

	logger.Debug("Condition evaluated with expression engine",
		logger.String("condition", condition),
		logger.Bool("result", result),
		logger.Any("variables", variables))

	return result
}

// GetElementType returns element type
// Возвращает тип элемента
func (ige *InclusiveGatewayExecutor) GetElementType() string {
	return "inclusiveGateway"
}
