/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"encoding/json"
	"fmt"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// processJobsResponses processes jobs responses in background
// Обрабатывает ответы jobs в фоне
func (c *Core) processJobsResponses() {
	if c.jobsComp == nil {
		return
	}

	responseChannel := c.jobsComp.GetResponseChannel()

	for {
		select {
		case response := <-responseChannel:
			// Process all job responses - simplified logic
			// Обрабатываем все ответы jobs - упрощенная логика
			c.handleJobsResponse(response)
		}
	}
}

// handleJobsResponse handles single jobs response
// Обрабатывает один ответ jobs
func (c *Core) handleJobsResponse(response string) {
	// Parse job callback response for readable logging
	// Парсим ответ job callback для читаемого логирования
	var jobResp struct {
		JobID             string `json:"job_id"`
		ElementID         string `json:"element_id"`
		TokenID           string `json:"token_id"`
		ProcessInstanceID string `json:"process_instance_id"`
		Status            string `json:"status"`
		CompletedAt       string `json:"completed_at"`
	}

	if err := json.Unmarshal([]byte(response), &jobResp); err == nil {
		logger.Info("CLI Job Callback",
			logger.String("element_id", jobResp.ElementID),
			logger.String("job_id", jobResp.JobID),
			logger.String("token_id", jobResp.TokenID),
			logger.String("process_instance_id", jobResp.ProcessInstanceID),
			logger.String("status", jobResp.Status),
			logger.String("completed_at", jobResp.CompletedAt))

		// Parse full callback for variables
		var fullCallback struct {
			JobID             string                 `json:"job_id"`
			ElementID         string                 `json:"element_id"`
			TokenID           string                 `json:"token_id"`
			ProcessInstanceID string                 `json:"process_instance_id"`
			Status            string                 `json:"status"`
			Variables         map[string]interface{} `json:"variables"`
			ErrorMessage      string                 `json:"error_message"`
		}

		json.Unmarshal([]byte(response), &fullCallback)

		// Forward job callback to process component
		// Передаем job callback в process component
		if c.processComp != nil {
			if err := c.processComp.HandleJobCallback(
				fullCallback.JobID,
				fullCallback.ElementID,
				fullCallback.TokenID,
				fullCallback.Status,
				fullCallback.ErrorMessage,
				fullCallback.Variables,
			); err != nil {
				logger.Error("Failed to handle job callback in process component",
					logger.String("job_id", fullCallback.JobID),
					logger.String("element_id", fullCallback.ElementID),
					logger.String("token_id", fullCallback.TokenID),
					logger.String("status", fullCallback.Status),
					logger.String("error", err.Error()))
			} else {
				logger.Info("Job callback processed successfully",
					logger.String("job_id", fullCallback.JobID),
					logger.String("element_id", fullCallback.ElementID),
					logger.String("token_id", fullCallback.TokenID),
					logger.String("status", fullCallback.Status))
			}
		}
	}

	// Also log full JSON for debugging
	// Также логируем полный JSON для отладки
	logger.Debug("Job completed", logger.String("response", response))

	// Log job response to storage
	// Логируем ответ job'а в storage
	err := c.storage.LogSystemEvent(
		models.EventTypeReady,
		models.StatusSuccess,
		fmt.Sprintf("Job completed: %s", response),
	)
	if err != nil {
		logger.Warn("Failed to log job response to storage", logger.String("error", err.Error()))
	}
}
