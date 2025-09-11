/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"fmt"
	"strings"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
	"atom-engine/src/incidents"
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
func (jc *JobCallbacks) HandleJobCallback(jobID, elementID, tokenID, status, errorMessage string, variables map[string]interface{}) error {
	if !jc.component.IsReady() {
		return fmt.Errorf("process component not ready")
	}

	logger.Info("Handling job callback",
		logger.String("job_id", jobID),
		logger.String("element_id", elementID),
		logger.String("token_id", tokenID),
		logger.String("status", status),
		logger.String("error_message", errorMessage))

	// Handle ERROR_THROWN differently - load token directly without state validation
	if status == "ERROR_THROWN" {
		// Load token directly for BPMN error processing
		token, err := jc.storage.LoadToken(tokenID)
		if err != nil {
			logger.Error("Failed to load token for BPMN error callback",
				logger.String("token_id", tokenID),
				logger.String("error", err.Error()))
			return fmt.Errorf("failed to load token %s: %w", tokenID, err)
		}

		logger.Info("Token loaded for BPMN error processing",
			logger.String("token_id", tokenID),
			logger.String("token_state", string(token.State)),
			logger.String("waiting_for", token.WaitingFor))

		// BPMN business error - activate error boundary events
		return jc.handleJobBPMNError(token, jobID, elementID, errorMessage, variables)
	}

	// For other statuses, load and validate token using helper
	expectedWaitingFor := fmt.Sprintf("job:%s", jobID)
	token, err := jc.callbackHelper.LoadAndValidateToken(tokenID, expectedWaitingFor)
	if err != nil {
		return err
	}

	// Handle different job statuses
	switch status {
	case "FAILED":
		// Job technical failure - check for error boundary events
		if errorMessage != "" {
			return jc.handleJobFailure(token, jobID, elementID, errorMessage, variables)
		}
	}

	// Process successful completion callback and continue execution using helper
	return jc.callbackHelper.ProcessCallbackAndContinue(token, elementID, variables)
}

// handleJobFailure handles job failure and checks for error boundary events
// Обрабатывает провал job'а и проверяет граничные события ошибок
func (jc *JobCallbacks) handleJobFailure(token *models.Token, jobID, elementID, errorMessage string, variables map[string]interface{}) error {
	logger.Info("Handling job failure",
		logger.String("job_id", jobID),
		logger.String("element_id", elementID),
		logger.String("token_id", token.TokenID),
		logger.String("error_message", errorMessage))

	// Extract error code from error message (simple pattern matching)
	// In real BPMN errors, error code might be passed differently
	errorCode := extractErrorCodeFromMessage(errorMessage)

	logger.Info("Extracted error code from job failure",
		logger.String("token_id", token.TokenID),
		logger.String("error_code", errorCode),
		logger.String("error_message", errorMessage))

	// Check if there are error boundary events registered for this token
	if errorBoundary := jc.component.FindMatchingErrorBoundary(token.TokenID, errorCode); errorBoundary != nil {
		logger.Info("Found matching error boundary event for job failure",
			logger.String("token_id", token.TokenID),
			logger.String("error_code", errorCode),
			logger.String("boundary_element_id", errorBoundary.ElementID))

		// Remove error boundary subscriptions for this token
		jc.component.RemoveErrorBoundariesForToken(token.TokenID)

		// Cancel the current token if this is an interrupting boundary event
		if errorBoundary.CancelActivity {
			logger.Info("Cancelling activity due to interrupting error boundary event",
				logger.String("token_id", token.TokenID),
				logger.String("attached_to", errorBoundary.AttachedToRef))

			token.SetState(models.TokenStateCanceled)
			if err := jc.storage.UpdateToken(token); err != nil {
				logger.Error("Failed to cancel token",
					logger.String("token_id", token.TokenID),
					logger.String("error", err.Error()))
			}
		}

		// Create new token for error boundary event continuation
		return jc.activateErrorBoundaryFlow(token, errorBoundary, variables)
	}

	// No error boundary found - handle as regular job failure
	logger.Info("No matching error boundary found for job failure, handling as regular failure",
		logger.String("token_id", token.TokenID),
		logger.String("error_code", errorCode))

	// Create incident for unhandled job failure
	err := jc.createJobFailureIncident(token, jobID, elementID, errorMessage)
	if err != nil {
		logger.Error("Failed to create job failure incident",
			logger.String("token_id", token.TokenID),
			logger.String("job_id", jobID),
			logger.String("element_id", elementID),
			logger.String("error", err.Error()))
	}

	// Mark token as failed
	token.SetState(models.TokenStateFailed)
	if err := jc.storage.UpdateToken(token); err != nil {
		logger.Error("Failed to update failed token",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
	}

	return fmt.Errorf("job failed: %s", errorMessage)
}

// extractErrorCodeFromMessage extracts error code from error message
// Simple implementation - in production might need more sophisticated parsing
func extractErrorCodeFromMessage(errorMessage string) string {
	// For now, check if message contains common HTTP error codes
	if contains(errorMessage, "404") {
		return "404"
	}
	if contains(errorMessage, "500") {
		return "500"
	}
	if contains(errorMessage, "403") {
		return "403"
	}
	// Default error code if no specific code found
	return "GENERAL_ERROR"
}

// contains checks if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(strings.ToLower(s), strings.ToLower(substr)))))
}

// activateErrorBoundaryFlow activates the error boundary event flow
// Активирует поток граничного события ошибки
func (jc *JobCallbacks) activateErrorBoundaryFlow(originalToken *models.Token, errorBoundary *ErrorBoundarySubscription, variables map[string]interface{}) error {
	logger.Info("Activating error boundary flow",
		logger.String("original_token_id", originalToken.TokenID),
		logger.String("boundary_element_id", errorBoundary.ElementID),
		logger.Int("outgoing_flows_count", len(errorBoundary.OutgoingFlows)))

	// Continue execution with outgoing flows from error boundary event
	if len(errorBoundary.OutgoingFlows) > 0 {
		// Use execution processor to continue with next elements
		if jc.component != nil {
			// Get the engine from component to access execution processor
			// Use the callback helper to continue execution
			return jc.callbackHelper.ProcessCallbackAndContinue(originalToken, errorBoundary.ElementID, variables)
		}
	}

	logger.Info("Error boundary event has no outgoing flows - process ends",
		logger.String("boundary_element_id", errorBoundary.ElementID))
	return nil
}

// handleJobBPMNError handles BPMN error thrown by job and activates error boundary events
// Обрабатывает BPMN ошибку выброшенную job'ом и активирует граничные события ошибок
func (jc *JobCallbacks) handleJobBPMNError(token *models.Token, jobID, elementID, errorMessage string, variables map[string]interface{}) error {
	// Extract errorCode from variables
	errorCode := "GENERAL_ERROR" // default
	if variables != nil {
		if errCode, exists := variables["errorCode"]; exists {
			if errCodeStr, ok := errCode.(string); ok {
				errorCode = errCodeStr
			}
		}
	}

	logger.Info("Handling BPMN error from job",
		logger.String("job_id", jobID),
		logger.String("element_id", elementID),
		logger.String("token_id", token.TokenID),
		logger.String("error_code", errorCode),
		logger.String("error_message", errorMessage))

	// Look for matching error boundary event
	errorBoundary := jc.component.FindMatchingErrorBoundary(token.TokenID, errorCode)
	if errorBoundary == nil {
		logger.Info("No matching error boundary found for BPMN error, creating incident",
			logger.String("token_id", token.TokenID),
			logger.String("error_code", errorCode))

		// No error boundary found - close job as ERROR_THROWN and create UNHANDLED_BPMN_ERROR incident
		// Граничного события ошибки не найдено - закрываем job как ERROR_THROWN и создаем UNHANDLED_BPMN_ERROR инцидент

		// Close job as ERROR_THROWN (no retries, no JOB_FAILURE incident)
		if err := jc.completeJobWithBPMNError(jobID, errorCode, errorMessage); err != nil {
			logger.Error("Failed to close job as ERROR_THROWN after unhandled BPMN error",
				logger.String("job_id", jobID),
				logger.String("error_code", errorCode),
				logger.String("error", err.Error()))
		}

		// Create UNHANDLED_BPMN_ERROR incident instead of JOB_FAILURE
		err := jc.createUnhandledBPMNErrorIncident(token, elementID, errorCode, errorMessage)
		if err != nil {
			logger.Error("Failed to create unhandled BPMN error incident",
				logger.String("token_id", token.TokenID),
				logger.String("element_id", elementID),
				logger.String("error_code", errorCode),
				logger.String("error", err.Error()))
		}

		token.SetState(models.TokenStateCanceled)
		token.SetVariables(variables)

		if err := jc.storage.SaveToken(token); err != nil {
			logger.Error("Failed to save token after BPMN error",
				logger.String("token_id", token.TokenID),
				logger.String("error", err.Error()))
		}

		return nil
	}

	logger.Info("Found matching error boundary for BPMN error, activating flow",
		logger.String("token_id", token.TokenID),
		logger.String("error_boundary_id", errorBoundary.ElementID),
		logger.String("error_code", errorCode))

	// Remove error boundary subscription
	jc.component.RemoveErrorBoundariesForToken(token.TokenID)

	// Complete job with BPMN error status (job is now closed)
	// Завершаем job со статусом BPMN error (job теперь закрыт)
	if err := jc.completeJobWithBPMNError(jobID, errorCode, errorMessage); err != nil {
		logger.Error("Failed to complete job after BPMN error boundary activation",
			logger.String("job_id", jobID),
			logger.String("error_code", errorCode),
			logger.String("error", err.Error()))
		// Continue processing despite job completion error
		// Продолжаем обработку несмотря на ошибку завершения job
	}

	// Cancel the original token
	originalToken := token
	originalToken.SetState(models.TokenStateCanceled)
	originalToken.SetVariables(variables)

	if err := jc.storage.SaveToken(originalToken); err != nil {
		logger.Error("Failed to save original token after BPMN error",
			logger.String("token_id", originalToken.TokenID),
			logger.String("error", err.Error()))
	}

	// Continue execution with outgoing flows from error boundary event
	if len(errorBoundary.OutgoingFlows) > 0 {
		// Use callback helper to continue execution from error boundary
		// For BPMN errors, we pass the error boundary element ID and variables with error info
		errorVariables := make(map[string]interface{})
		for k, v := range variables {
			errorVariables[k] = v
		}
		errorVariables["errorCode"] = errorCode
		errorVariables["errorMessage"] = errorMessage

		return jc.callbackHelper.ProcessCallbackAndContinue(originalToken, errorBoundary.ElementID, errorVariables)
	}

	logger.Info("Error boundary event has no outgoing flows, process ends",
		logger.String("error_boundary_id", errorBoundary.ElementID))

	return nil
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

// failJob marks specific job as failed via jobs component
// Помечает конкретный job как failed через jobs компонент
func (jc *JobCallbacks) failJob(jobID, errorMessage string) error {
	if jc.core == nil {
		return fmt.Errorf("core interface not set")
	}

	jobsComp := jc.core.GetJobsComponent()
	if jobsComp == nil {
		return fmt.Errorf("jobs component not available")
	}

	// Try to fail job via jobs component interface with retries and error message
	// Пытаемся пометить job как failed через интерфейс jobs компонента с retries и сообщением об ошибке
	if jobsFailMethod, ok := jobsComp.(interface {
		FailJob(jobKey string, retries int, errorMessage string) error
	}); ok {
		if err := jobsFailMethod.FailJob(jobID, 0, errorMessage); err != nil {
			return fmt.Errorf("failed to mark job as failed via jobs component: %w", err)
		}
		logger.Info("Job marked as failed via jobs component",
			logger.String("job_id", jobID),
			logger.String("error_message", errorMessage))
		return nil
	}

	// Fallback: log warning
	// Фоллбэк: логируем предупреждение
	logger.Warn("Jobs component doesn't support FailJob method",
		logger.String("job_id", jobID))

	return nil
}

// completeJobWithBPMNError marks job as completed with BPMN error status via jobs component
// Помечает job как completed с статусом BPMN error через jobs компонент
func (jc *JobCallbacks) completeJobWithBPMNError(jobID, errorCode, errorMessage string) error {
	if jc.core == nil {
		return fmt.Errorf("core interface not set")
	}

	jobsComp := jc.core.GetJobsComponent()
	if jobsComp == nil {
		return fmt.Errorf("jobs component not available")
	}

	// Try to complete job with BPMN error via jobs component interface
	// Пытаемся завершить job с BPMN error через интерфейс jobs компонента
	if jobsCompleteMethod, ok := jobsComp.(interface {
		CompleteJobWithBPMNError(jobKey, errorCode, errorMessage string) error
	}); ok {
		if err := jobsCompleteMethod.CompleteJobWithBPMNError(jobID, errorCode, errorMessage); err != nil {
			return fmt.Errorf("failed to complete job with BPMN error via jobs component: %w", err)
		}
		logger.Info("Job completed as ERROR_THROWN via jobs component",
			logger.String("job_id", jobID),
			logger.String("error_code", errorCode))
		return nil
	}

	// Fallback: log warning
	// Фоллбэк: логируем предупреждение
	logger.Warn("Jobs component doesn't support CompleteJobWithBPMNError method",
		logger.String("job_id", jobID))

	return nil
}

// createJobFailureIncident creates incident for unhandled job failure
func (jc *JobCallbacks) createJobFailureIncident(token *models.Token, jobID, elementID, errorMessage string) error {
	if jc.component == nil {
		return fmt.Errorf("component not available")
	}

	if jc.core == nil {
		return fmt.Errorf("core interface not available")
	}

	incidentsComp := jc.core.GetIncidentsComponent()
	if incidentsComp == nil {
		return fmt.Errorf("incidents component not available")
	}

	payload := incidents.CreateIncidentPayload{
		Type:              "job_failure",
		Message:           errorMessage,
		ProcessInstanceID: token.ProcessInstanceID,
		ElementID:         elementID,
		ElementType:       "serviceTask", // Service task element type
		JobKey:            jobID,
		OriginalRetries:   0,
	}

	message, err := incidents.CreateIncidentMessage(payload)
	if err != nil {
		return fmt.Errorf("failed to create incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	err = jc.core.SendMessage("incidents", message)
	if err != nil {
		return fmt.Errorf("failed to create job failure incident: %w", err)
	}

	logger.Info("Job failure incident created successfully",
		logger.String("token_id", token.TokenID),
		logger.String("job_id", jobID),
		logger.String("element_id", elementID),
		logger.String("process_instance_id", token.ProcessInstanceID))

	return nil
}

// createBPMNErrorIncident creates incident for unhandled BPMN error
func (jc *JobCallbacks) createBPMNErrorIncident(token *models.Token, elementID, errorCode, errorMessage string) error {
	if jc.component == nil {
		return fmt.Errorf("component not available")
	}

	if jc.core == nil {
		return fmt.Errorf("core interface not available")
	}

	incidentsComp := jc.core.GetIncidentsComponent()
	if incidentsComp == nil {
		return fmt.Errorf("incidents component not available")
	}

	payload := incidents.CreateIncidentPayload{
		Type:              "bpmn_error",
		Message:           fmt.Sprintf("%s: %s", errorCode, errorMessage),
		ErrorCode:         errorCode,
		ProcessInstanceID: token.ProcessInstanceID,
		ElementID:         elementID,
		ElementType:       "serviceTask", // Service task element type
	}

	message, err := incidents.CreateIncidentMessage(payload)
	if err != nil {
		return fmt.Errorf("failed to create incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	err = jc.core.SendMessage("incidents", message)
	if err != nil {
		return fmt.Errorf("failed to create BPMN error incident: %w", err)
	}

	logger.Info("BPMN error incident created successfully",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID),
		logger.String("error_code", errorCode),
		logger.String("process_instance_id", token.ProcessInstanceID))

	return nil
}

// createUnhandledBPMNErrorIncident creates incident for unhandled BPMN error (when no boundary event found)
func (jc *JobCallbacks) createUnhandledBPMNErrorIncident(token *models.Token, elementID, errorCode, errorMessage string) error {
	if jc.component == nil {
		return fmt.Errorf("component not available")
	}

	if jc.core == nil {
		return fmt.Errorf("core interface not available")
	}

	incidentsComp := jc.core.GetIncidentsComponent()
	if incidentsComp == nil {
		return fmt.Errorf("incidents component not available")
	}

	payload := incidents.CreateIncidentPayload{
		Type:              "bpmn_error",
		Message:           fmt.Sprintf("UNHANDLED_BPMN_ERROR %s: %s", errorCode, errorMessage),
		ErrorCode:         errorCode,
		ProcessInstanceID: token.ProcessInstanceID,
		ElementID:         elementID,
		ElementType:       "serviceTask", // Service task element type
	}

	message, err := incidents.CreateIncidentMessage(payload)
	if err != nil {
		return fmt.Errorf("failed to create incident message: %w", err)
	}

	// Send JSON message to incidents component through Core
	err = jc.core.SendMessage("incidents", message)
	if err != nil {
		return fmt.Errorf("failed to create unhandled BPMN error incident: %w", err)
	}

	logger.Info("Unhandled BPMN error incident created successfully",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", elementID),
		logger.String("error_code", errorCode),
		logger.String("process_instance_id", token.ProcessInstanceID))

	return nil
}
