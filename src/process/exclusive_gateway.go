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
	"atom-engine/src/expression"
)

// ExclusiveGatewayExecutor executes exclusive gateways
// Исполнитель эксклюзивных шлюзов
type ExclusiveGatewayExecutor struct {
	processComponent ComponentInterface
}

// NewExclusiveGatewayExecutor creates new exclusive gateway executor
// Создает новый исполнитель эксклюзивного шлюза
func NewExclusiveGatewayExecutor(processComponent ComponentInterface) *ExclusiveGatewayExecutor {
	return &ExclusiveGatewayExecutor{
		processComponent: processComponent,
	}
}

// Execute executes exclusive gateway
// Выполняет эксклюзивный шлюз
func (ege *ExclusiveGatewayExecutor) Execute(token *models.Token, element map[string]interface{}) (*ExecutionResult, error) {
	logger.Info("Executing exclusive gateway",
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
			Error:     "exclusive gateway has no outgoing sequence flows",
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
			Error:     "exclusive gateway has no valid outgoing sequence flows",
			Completed: false,
		}, nil
	}

	// Load BPMN process to get sequence flow conditions
	// Загружаем BPMN процесс чтобы получить условия sequence flows
	var selectedFlow string
	var err error

	if ege.processComponent != nil {
		selectedFlow, err = ege.evaluateGatewayConditions(token, outgoingFlows)
		if err != nil {
			logger.Error("Failed to evaluate gateway conditions",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
			// Fallback to first flow if condition evaluation fails
			// Возврат к первому потоку если оценка условий не удалась
			selectedFlow = outgoingFlows[0]
		}
	} else {
		// No process component available, use first flow as fallback
		// Нет доступного process component, используем первый поток как fallback
		selectedFlow = outgoingFlows[0]
		logger.Warn("Process component not available, using first flow",
			logger.String("token_id", token.TokenID))
	}

	logger.Info("Exclusive gateway executed",
		logger.String("token_id", token.TokenID),
		logger.String("gateway_name", gatewayName),
		logger.String("selected_flow", selectedFlow),
		logger.Int("total_flows", len(outgoingFlows)))

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: false,
		NextElements: []string{selectedFlow},
		Completed:    false,
	}, nil
}

// evaluateGatewayConditions evaluates sequence flow conditions for gateway
// Оценивает условия sequence flows для шлюза
func (ege *ExclusiveGatewayExecutor) evaluateGatewayConditions(token *models.Token, outgoingFlows []string) (string, error) {
	// Get BPMN process data
	// Получаем данные BPMN процесса
	processData, err := ege.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to get BPMN process: %w", err)
	}

	elements, ok := processData["elements"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid process elements structure")
	}

	// Prepare token variables as context for expression evaluation
	// Подготавливаем переменные токена как контекст для оценки выражений
	evaluationContext := make(map[string]interface{})
	for key, value := range token.Variables {
		evaluationContext[key] = value
	}

	logger.Info("Evaluating gateway conditions",
		logger.String("token_id", token.TokenID),
		logger.Int("outgoing_flows_count", len(outgoingFlows)),
		logger.Any("token_variables", token.Variables))

	var defaultFlow string

	// Evaluate each outgoing flow condition
	// Оцениваем условие каждого исходящего потока
	for _, flowID := range outgoingFlows {
		flowElement, exists := elements[flowID]
		if !exists {
			logger.Warn("Outgoing flow not found in elements",
				logger.String("flow_id", flowID))
			continue
		}

		flowMap, ok := flowElement.(map[string]interface{})
		if !ok {
			logger.Warn("Invalid flow element structure",
				logger.String("flow_id", flowID))
			continue
		}

		// Debug: Check sequence_flow structure
		// Отладка: Проверяем структуру sequence_flow
		if sequenceFlow, hasSeqFlow := flowMap["sequence_flow"]; hasSeqFlow {
			logger.Info("Flow sequence_flow structure",
				logger.String("flow_id", flowID),
				logger.Any("sequence_flow", sequenceFlow))
		}

		// Check if flow has condition
		// Проверяем есть ли у потока условие
		var conditionData interface{}
		hasCondition := false

		// First check direct condition field
		if cond, exists := flowMap["condition"]; exists {
			conditionData = cond
			hasCondition = true
			logger.Info("Found direct condition",
				logger.String("flow_id", flowID),
				logger.Any("condition", cond))
		}

		// Also check sequence_flow.condition
		if seqFlow, exists := flowMap["sequence_flow"]; exists {
			if seqFlowMap, ok := seqFlow.(map[string]interface{}); ok {
				if cond, exists := seqFlowMap["condition"]; exists {
					conditionData = cond
					hasCondition = true
					logger.Info("Found sequence_flow condition",
						logger.String("flow_id", flowID),
						logger.Any("condition", cond))
				}
			}
		}

		if hasCondition {
			conditionMap, ok := conditionData.(map[string]interface{})
			if !ok {
				logger.Warn("Invalid condition structure",
					logger.String("flow_id", flowID))
				continue
			}

			expression, ok := conditionMap["expression"].(string)
			if !ok || expression == "" {
				logger.Warn("Empty or invalid condition expression",
					logger.String("flow_id", flowID))
				continue
			}

			logger.Info("Evaluating flow condition",
				logger.String("flow_id", flowID),
				logger.String("expression", expression))

			// Get expression component through core and evaluate condition
			// Получаем expression компонент через core и оцениваем условие
			result := ege.evaluateConditionWithExpressionEngine(expression, evaluationContext)

			logger.Info("Condition evaluation result",
				logger.String("flow_id", flowID),
				logger.String("expression", expression),
				logger.Bool("result", result))

			if result {
				return flowID, nil
			}
		} else {
			// Flow without condition - potential default flow
			// Поток без условия - потенциальный default поток
			if defaultFlow == "" {
				defaultFlow = flowID
				logger.Info("Found potential default flow",
					logger.String("flow_id", flowID))
			}
		}
	}

	// If no condition evaluated to true, use default flow or first flow
	// Если ни одно условие не истинно, используем default поток или первый поток
	if defaultFlow != "" {
		logger.Info("Using default flow (no condition)",
			logger.String("flow_id", defaultFlow))
		return defaultFlow, nil
	}

	// Fallback to first flow if no conditions match and no default found
	// Возврат к первому потоку если никакие условия не подошли и default не найден
	if len(outgoingFlows) > 0 {
		logger.Warn("No conditions matched and no default flow, using first flow",
			logger.String("flow_id", outgoingFlows[0]))
		return outgoingFlows[0], nil
	}

	return "", fmt.Errorf("no valid outgoing flows found")
}

// evaluateConditionWithExpressionEngine evaluates condition using full expression engine
// Оценивает условие используя полноценный expression engine
func (ege *ExclusiveGatewayExecutor) evaluateConditionWithExpressionEngine(condition string, variables map[string]interface{}) bool {
	// Get expression component through core
	// Получаем expression компонент через core
	if ege.processComponent == nil {
		logger.Warn("Process component not available for expression evaluation")
		return false
	}

	// Get core interface
	core := ege.processComponent.GetCore()
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
func (ege *ExclusiveGatewayExecutor) GetElementType() string {
	return "exclusiveGateway"
}
