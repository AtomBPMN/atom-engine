/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package server

import (
	"fmt"

	"atom-engine/src/core/logger"
)

// initializeJobsMultiplexer initializes and starts the Jobs Message Multiplexer
func (c *Core) initializeJobsMultiplexer() error {
	if c.jobsComp == nil {
		return fmt.Errorf("jobs component not available")
	}

	// Get the source channel from jobs component
	sourceChannel := c.jobsComp.GetResponseChannel()
	if sourceChannel == nil {
		return fmt.Errorf("jobs response channel not available")
	}

	// Create logger for multiplexer
	multiplexerLogger := logger.NewComponentLogger("jobs-multiplexer")

	// Create the multiplexer
	c.jobsMultiplexer = NewMessageMultiplexer(
		sourceChannel,
		"jobs",
		multiplexerLogger,
	)

	// Start the multiplexer
	err := c.jobsMultiplexer.Start()
	if err != nil {
		return fmt.Errorf("failed to start jobs message multiplexer: %w", err)
	}

	logger.Info("Jobs Message Multiplexer initialized and started successfully")
	return nil
}

// processJobCallbacks processes job callbacks from the multiplexer
func (c *Core) processJobCallbacks() {
	if c.jobsMultiplexer == nil {
		logger.Warn("Jobs multiplexer not available, cannot process job callbacks")
		return
	}

	callbackChannel := c.jobsMultiplexer.GetJobCallbackChannel()
	if callbackChannel == nil {
		logger.Warn("Job callback channel not available")
		return
	}

	logger.Info("Started processing job callbacks via message multiplexer")

	for {
		select {
		case callback, ok := <-callbackChannel:
			if !ok {
				logger.Info("Job callback channel closed, stopping callback processor")
				return
			}

			// Process the job callback
			c.handleJobCallback(callback)

		case bpmnError, ok := <-c.jobsMultiplexer.GetBPMNErrorChannel():
			if !ok {
				logger.Info("BPMN error channel closed")
				continue
			}

			// Process BPMN error callback
			c.handleBPMNErrorCallback(bpmnError)
		}
	}
}

// handleJobCallback processes a single job callback
func (c *Core) handleJobCallback(callback string) {
	logger.Debug("Processing job callback via multiplexer",
		logger.Int("callback_length", len(callback)))

	// Use the existing job callback handling logic
	c.handleJobsResponse(callback)
}

// handleBPMNErrorCallback processes a BPMN error callback
func (c *Core) handleBPMNErrorCallback(bpmnError string) {
	logger.Debug("Processing BPMN error callback via multiplexer",
		logger.Int("error_length", len(bpmnError)))

	// BPMN errors are also job callbacks, so we use the same handling logic
	// BPMN ошибки также являются job callback'ами, поэтому используем ту же логику обработки
	c.handleJobsResponse(bpmnError)
}

// getJobsMultiplexerMetrics returns metrics from the jobs multiplexer
func (c *Core) GetJobsMultiplexerMetrics() (MultiplexerMetrics, error) {
	if c.jobsMultiplexer == nil {
		return MultiplexerMetrics{}, fmt.Errorf("jobs multiplexer not available")
	}

	return c.jobsMultiplexer.GetMetrics(), nil
}

// stopJobsMultiplexer gracefully stops the jobs message multiplexer
func (c *Core) stopJobsMultiplexer() error {
	if c.jobsMultiplexer == nil {
		return nil // Already stopped or never started
	}

	logger.Info("Stopping jobs message multiplexer")

	err := c.jobsMultiplexer.Stop()
	if err != nil {
		logger.Error("Error stopping jobs message multiplexer", logger.String("error", err.Error()))
		return err
	}

	logger.Info("Jobs message multiplexer stopped successfully")
	return nil
}
