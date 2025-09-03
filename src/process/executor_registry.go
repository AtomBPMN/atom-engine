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
	logger.Info("Registering StartEventExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewStartEventExecutor(er.component))
	logger.Info("Registering EndEventExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewEndEventExecutor(er.component))
	er.RegisterExecutor(&TaskExecutor{})
	er.RegisterExecutor(&UserTaskExecutor{})

	// Register service task executor with process component access
	logger.Info("Registering ServiceTaskExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewServiceTaskExecutor(er.component))

	// Register gateway executors with process component access
	logger.Info("Registering ExclusiveGatewayExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewExclusiveGatewayExecutor(er.component))
	er.RegisterExecutor(&ParallelGatewayExecutor{})
	er.RegisterExecutor(&InclusiveGatewayExecutor{})
	er.RegisterExecutor(&EventBasedGatewayExecutor{})

	// Register event executors with process component access
	logger.Info("Registering IntermediateCatchEventExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewIntermediateCatchEventExecutor(er.component))
	logger.Info("Registering IntermediateThrowEventExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewIntermediateThrowEventExecutor(er.component))

	// Register boundary event executors with process component access
	logger.Info("Registering BoundaryEventExecutor with process component", logger.Bool("hasComponentInterface", er.component != nil))
	er.RegisterExecutor(NewBoundaryEventExecutor(er.component))

	// Register task executors
	er.RegisterExecutor(&ScriptTaskExecutor{})
	er.RegisterExecutor(&CallActivityExecutor{})
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
