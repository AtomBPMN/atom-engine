/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package expression

import (
	"context"
	"fmt"

	"atom-engine/src/core/logger"
)

// Component represents the expression evaluation component
// Представляет компонент оценки выражений
type Component struct {
	evaluator        *ExpressionEvaluator
	evaluationHelper *EvaluationHelper
	logger           logger.ComponentLogger
	ready            bool
	ctx              context.Context
	cancel           context.CancelFunc
}

// ComponentInterface defines expression component interface
// Определяет интерфейс компонента выражений
type ComponentInterface interface {
	Init() error
	Start() error
	Stop() error
	IsReady() bool

	// Main evaluation methods
	// Основные методы оценки
	EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error)
	EvaluateCondition(variables map[string]interface{}, condition string) (bool, error)
	EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	ParseRetries(retriesStr string) (int, error)

	// Helper access
	// Доступ к хелперам
	GetEvaluationHelper() *EvaluationHelper
}

// NewComponent creates new expression component
// Создает новый компонент выражений
func NewComponent() *Component {
	ctx, cancel := context.WithCancel(context.Background())

	return &Component{
		logger: logger.NewComponentLogger("expression"),
		ctx:    ctx,
		cancel: cancel,
		ready:  false,
	}
}

// Init initializes expression component
// Инициализирует компонент выражений
func (c *Component) Init() error {
	c.logger.Info("Initializing expression component")

	// Initialize main evaluator
	// Инициализируем главный оценщик
	c.evaluator = NewExpressionEvaluator()
	if c.evaluator == nil {
		return fmt.Errorf("failed to create expression evaluator")
	}

	// Initialize evaluation helper
	// Инициализируем хелпер оценки
	c.evaluationHelper = NewEvaluationHelper(c.evaluator)
	if c.evaluationHelper == nil {
		return fmt.Errorf("failed to create evaluation helper")
	}

	c.logger.Info("Expression component initialized successfully")
	return nil
}

// Start starts expression component
// Запускает компонент выражений
func (c *Component) Start() error {
	c.logger.Info("Starting expression component")

	if c.evaluator == nil {
		return fmt.Errorf("expression evaluator not initialized")
	}

	c.ready = true
	c.logger.Info("Expression component started successfully")
	return nil
}

// Stop stops expression component
// Останавливает компонент выражений
func (c *Component) Stop() error {
	c.logger.Info("Stopping expression component")

	c.ready = false

	if c.cancel != nil {
		c.cancel()
	}

	c.logger.Info("Expression component stopped successfully")
	return nil
}

// IsReady returns whether component is ready
// Возвращает готовность компонента
func (c *Component) IsReady() bool {
	return c.ready && c.evaluator != nil
}

// EvaluateExpression evaluates expression in parameters
// Вычисляет выражение в параметрах
func (c *Component) EvaluateExpression(expression string, variables map[string]interface{}) (interface{}, error) {
	if !c.IsReady() {
		return nil, fmt.Errorf("expression component not ready")
	}

	return c.evaluator.EvaluateExpression(expression, variables)
}

// EvaluateCondition evaluates conditional expression
// Вычисляет условное выражение
func (c *Component) EvaluateCondition(variables map[string]interface{}, condition string) (bool, error) {
	if !c.IsReady() {
		return false, fmt.Errorf("expression component not ready")
	}

	return c.evaluator.EvaluateCondition(variables, condition)
}

// EvaluateExpressionEngine full expression engine
// Полноценный движок выражений
func (c *Component) EvaluateExpressionEngine(
	expression interface{},
	variables map[string]interface{},
) (interface{}, error) {
	if !c.IsReady() {
		return nil, fmt.Errorf("expression component not ready")
	}

	return c.evaluator.EvaluateExpressionEngine(expression, variables)
}

// ParseRetries parses retries count from string
// Парсит количество повторов из строки
func (c *Component) ParseRetries(retriesStr string) (int, error) {
	if !c.IsReady() {
		return 3, fmt.Errorf("expression component not ready")
	}

	return c.evaluator.ParseRetries(retriesStr)
}

// GetEvaluationHelper returns evaluation helper
// Возвращает хелпер оценки
func (c *Component) GetEvaluationHelper() *EvaluationHelper {
	return c.evaluationHelper
}
