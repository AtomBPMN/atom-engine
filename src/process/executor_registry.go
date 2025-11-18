/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"atom-engine/src/core/logger"
)

// ExecutorRegistry manages element executor registration
// Управляет регистрацией исполнителей элементов
type ExecutorRegistry struct {
	executors map[string]ElementExecutor
	component ComponentInterface
}

// NewExecutorRegistry creates new executor registry
// Создает новый реестр исполнителей
func NewExecutorRegistry(component ComponentInterface) *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]ElementExecutor),
		component: component,
	}
}

// registerExecutors registers built-in element executors
// Регистрирует встроенные исполнители элементов
func (er *ExecutorRegistry) registerExecutors() {
	// Register basic element executors with process component access
	logger.Info("Registering StartEventExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewStartEventExecutor(er.component))
	logger.Info("Registering EndEventExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewEndEventExecutor(er.component))
	er.RegisterExecutor(&TaskExecutor{})
	er.RegisterExecutor(&UserTaskExecutor{})

	// Register service task executor with process component access
	logger.Info("Registering ServiceTaskExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewServiceTaskExecutor(er.component))

	// Note: HttpConnectorExecutor is handled dynamically in GetServiceTaskExecutor

	// Register gateway executors with process component access
	logger.Info("Registering ExclusiveGatewayExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewExclusiveGatewayExecutor(er.component))
	logger.Info("Registering ParallelGatewayExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewParallelGatewayExecutor(er.component))
	logger.Info("Registering InclusiveGatewayExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewInclusiveGatewayExecutor(er.component))
	logger.Info("Registering EventBasedGatewayExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewEventBasedGatewayExecutor(er.component))

	// Register event executors with process component access
	logger.Info("Registering IntermediateCatchEventExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewIntermediateCatchEventExecutor(er.component))
	logger.Info("Registering IntermediateThrowEventExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewIntermediateThrowEventExecutor(er.component))

	// Register boundary event executors with process component access
	logger.Info("Registering BoundaryEventExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewBoundaryEventExecutor(er.component))

	// Register task executors
	er.RegisterExecutor(&ScriptTaskExecutor{})
	logger.Info("Registering CallActivityExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewCallActivityExecutor(er.component))

	// Register Send Task and Receive Task executors
	logger.Info("Registering SendTaskExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewSendTaskExecutor(er.component))
	logger.Info("Registering ReceiveTaskExecutor with process component",
		logger.Bool("hasComponentInterface", er.component != nil),
	)
	er.RegisterExecutor(NewReceiveTaskExecutor(er.component))
}

// RegisterExecutor registers element executor
// Регистрирует исполнитель элемента
func (er *ExecutorRegistry) RegisterExecutor(executor ElementExecutor) {
	er.executors[executor.GetElementType()] = executor
	logger.Info("Registered element executor", logger.String("type", executor.GetElementType()))
}

// GetExecutor gets executor for element type
// Получает исполнитель для типа элемента
func (er *ExecutorRegistry) GetExecutor(elementType string) (ElementExecutor, bool) {
	executor, exists := er.executors[elementType]
	return executor, exists
}

// GetServiceTaskExecutor gets appropriate executor for service task
// Получает подходящий исполнитель для service task
func (er *ExecutorRegistry) GetServiceTaskExecutor(element map[string]interface{}) (ElementExecutor, bool) {
	// Try HTTP connector executor first
	httpExecutor := NewHttpConnectorExecutor(er.component)
	if er.isHttpConnector(element) {
		return httpExecutor, true
	}

	// Use regular service task executor
	return NewServiceTaskExecutor(er.component), true
}

// isHttpConnector checks if element is HTTP connector
// Проверяет, является ли элемент HTTP коннектором
func (er *ExecutorRegistry) isHttpConnector(element map[string]interface{}) bool {
	logger.Debug("Checking if element is HTTP connector", logger.String("element_id", getStringValue(element["id"])))

	// Look for extension elements
	extensionElements, exists := element["extension_elements"]
	if !exists {
		logger.Debug("No extension_elements found")
		return false
	}

	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		logger.Debug("extension_elements is not array")
		return false
	}

	// Find taskDefinition in extension elements
	for _, extElement := range extElementsList {
		extElementMap, ok := extElement.(map[string]interface{})
		if !ok {
			continue
		}

		extensions, exists := extElementMap["extensions"]
		if !exists {
			continue
		}

		extensionsList, ok := extensions.([]interface{})
		if !ok {
			continue
		}

		for _, ext := range extensionsList {
			extMap, ok := ext.(map[string]interface{})
			if !ok {
				continue
			}

			extType, exists := extMap["type"]
			if !exists || extType != "taskDefinition" {
				continue
			}

			logger.Debug("Found taskDefinition extension")

			// Found taskDefinition - check type
			taskDef, exists := extMap["task_definition"]
			if !exists {
				logger.Debug("No task_definition data found")
				continue
			}

			taskDefMap, ok := taskDef.(map[string]interface{})
			if !ok {
				logger.Debug("task_definition is not a map")
				continue
			}

			jobType, _ := taskDefMap["type"].(string)
			logger.Debug("Task definition type found", logger.String("type", jobType))
			isHttp := jobType == "io.camunda:http-json:1"
			logger.Debug("HTTP connector check result", logger.Bool("isHttpConnector", isHttp))
			return isHttp
		}
	}

	logger.Debug("No taskDefinition found")
	return false
}

// getStringValue safely extracts string value from interface{}
func getStringValue(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
