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
	"atom-engine/src/storage"
)

// JobCallbacks handles job-related callbacks
// Обрабатывает callbacks связанные с jobs
type JobCallbacks struct {
	storage        storage.Storage
	component      ComponentInterface
	core           CoreInterface
	callbackHelper *CallbackHelper
}

// NewJobCallbacks creates new job callbacks handler
// Создает новый обработчик callbacks jobs
func NewJobCallbacks(storage storage.Storage, component ComponentInterface) *JobCallbacks {
	return &JobCallbacks{
		storage:        storage,
		component:      component,
		callbackHelper: NewCallbackHelper(storage, component),
	}
}

// SetCore sets core interface for job management
// Устанавливает интерфейс core для управления jobs
func (jc *JobCallbacks) SetCore(core CoreInterface) {
	jc.core = core
}

// Init initializes job callbacks handler
// Инициализирует обработчик callbacks jobs
func (jc *JobCallbacks) Init() error {
	logger.Info("Initializing job callbacks handler")
	return nil
}

// HandleJobCallback handles job completion callback
// Обрабатывает callback завершения job
func (jc *JobCallbacks) HandleJobCallback(jobID, elementID, tokenID string, variables map[string]interface{}) error {
	if !jc.component.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	logger.Info("Handling job callback",
		logger.String("job_id", jobID),
		logger.String("element_id", elementID),
		logger.String("token_id", tokenID))

	// Load and validate token using helper
	expectedWaitingFor := fmt.Sprintf("job:%s", jobID)
	token, err := jc.callbackHelper.LoadAndValidateToken(tokenID, expectedWaitingFor)
	if err != nil {
		return err
	}

	// Process callback and continue execution using helper
	return jc.callbackHelper.ProcessCallbackAndContinue(token, elementID, variables)
}

// cancelJob cancels specific job via jobs component
// Отменяет конкретный job через jobs компонент
func (jc *JobCallbacks) cancelJob(jobID string) error {
	if jc.core == nil {
		return fmt.Errorf("core interface not set")
	}

	jobsComp := jc.core.GetJobsComponent()
	if jobsComp == nil {
		return fmt.Errorf("jobs component not available")
	}

	// Try to cancel job via jobs component interface with reason parameter
	// Пытаемся отменить job через интерфейс jobs компонента с параметром reason
	if jobsCancelMethod, ok := jobsComp.(interface {
		CancelJob(jobID, reason string) error
	}); ok {
		if err := jobsCancelMethod.CancelJob(jobID, "interrupted by boundary timer"); err != nil {
			return fmt.Errorf("failed to cancel job via jobs component: %w", err)
		}
		logger.Info("Job canceled via jobs component",
			logger.String("job_id", jobID))
		return nil
	}

	// Fallback: try without reason parameter (older interface)
	// Фоллбэк: пробуем без параметра reason (старый интерфейс)
	if jobsCancelMethod, ok := jobsComp.(interface {
		CancelJob(jobID string) error
	}); ok {
		if err := jobsCancelMethod.CancelJob(jobID); err != nil {
			return fmt.Errorf("failed to cancel job via jobs component: %w", err)
		}
		logger.Info("Job canceled via jobs component (fallback interface)",
			logger.String("job_id", jobID))
		return nil
	}

	// Last fallback: log warning
	// Последний фоллбэк: логируем предупреждение
	logger.Warn("Jobs component doesn't support cancellation",
		logger.String("job_id", jobID))

	return nil
}
