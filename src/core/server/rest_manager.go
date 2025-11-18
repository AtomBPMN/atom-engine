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
	"atom-engine/src/core/restapi"
)

// startRESTServer starts REST API server
// Запускает REST API сервер
func (c *Core) startRESTServer() error {
	restConfig := &restapi.Config{
		Host: c.config.RestAPI.Host,
		Port: c.config.RestAPI.Port,
	}

	if restConfig.Port == 0 {
		restConfig.Port = 27555 // Default port
	}

	if restConfig.Host == "" {
		restConfig.Host = "localhost"
	}

	server := restapi.NewServer(restConfig, c)
	err := server.Start()
	if err != nil {
		return fmt.Errorf("failed to start REST API server: %w", err)
	}

	c.restServer = server
	
	logger.Info("REST API server started",
		logger.String("host", restConfig.Host),
		logger.Any("port", restConfig.Port))
	
	return nil
}

// stopRESTServer stops REST API server
// Останавливает REST API сервер
func (c *Core) stopRESTServer() {
	if c.restServer != nil {
		err := c.restServer.Stop()
		if err != nil {
			logger.Error("Failed to stop REST API server", logger.String("error", err.Error()))
		} else {
			logger.Info("REST API server stopped")
		}
		c.restServer = nil
	}
}
