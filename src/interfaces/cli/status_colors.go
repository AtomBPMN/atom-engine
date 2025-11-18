/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

// StandardStatusColors provides centralized status colorization for all CLI commands
// Стандартизированные цвета статусов для всех CLI команд

// ColorizeOperationStatus colors operation result statuses
// Окрашивает статусы результатов операций
func ColorizeOperationStatus(status string) string {
	switch status {
	case "SUCCESS":
		return colorize(status, ColorBoldGreen)
	case "FAILED":
		return colorize(status, ColorBoldRed)
	default:
		return colorizeStatus(status) // Fallback to general status coloring
	}
}

// ColorizeDaemonStatus colors daemon-specific status messages
// Окрашивает статусы демона
func ColorizeDaemonStatus(status string) string {
	switch status {
	case "running":
		return colorize(status, ColorBoldGreen)
	case "stopped", "not running":
		return colorize(status, ColorGray)
	case "started":
		return colorize(status, ColorBoldGreen)
	default:
		return colorizeStatus(status)
	}
}

// ColorizeSystemStatus colors system component statuses
// Окрашивает статусы системных компонентов
func ColorizeSystemStatus(status string) string {
	switch status {
	case "ready", "READY":
		return colorize(status, ColorBoldGreen)
	case "starting", "STARTING":
		return colorize(status, ColorBoldYellow)
	case "stopping", "STOPPING":
		return colorize(status, ColorBoldYellow)
	case "stopped", "STOPPED":
		return colorize(status, ColorGray)
	case "error", "ERROR":
		return colorize(status, ColorBoldRed)
	default:
		return colorizeStatus(status)
	}
}

// ColorizeEventStatus colors event status in system events
// Окрашивает статусы событий в системных событиях
func ColorizeEventStatus(status string) string {
	switch status {
	case "INFO":
		return colorize(status, ColorBoldBlue)
	case "WARNING", "WARN":
		return colorize(status, ColorBoldYellow)
	case "ERROR":
		return colorize(status, ColorBoldRed)
	case "DEBUG":
		return colorize(status, ColorGray)
	case "SUCCESS":
		return colorize(status, ColorBoldGreen)
	default:
		return colorizeStatus(status)
	}
}

// ColorizeMessage applies appropriate color to status messages
// Применяет подходящий цвет к сообщениям о статусе
func ColorizeMessage(message string) string {
	// Simple heuristic for common message patterns
	// Простая эвристика для общих паттернов сообщений
	switch {
	case message == "Atom Engine daemon is running":
		return colorize(message, ColorBoldGreen)
	case message == "Atom Engine daemon stopped":
		return colorize(message, ColorGray)
	case message == "Core system started":
		return colorize(message, ColorBoldGreen)
	case message == "Core system stopped":
		return colorize(message, ColorGray)
	case message == "Daemon is not running":
		return colorize(message, ColorGray)
	default:
		return message
	}
}

// ColorizeConnectionStatus colors connection/health statuses
// Окрашивает статусы подключения/здоровья
func ColorizeConnectionStatus(connected, healthy bool) (string, string) {
	var connectedStr, healthyStr string

	if connected {
		connectedStr = colorize("true", ColorBoldGreen)
	} else {
		connectedStr = colorize("false", ColorBoldRed)
	}

	if healthy {
		healthyStr = colorize("true", ColorBoldGreen)
	} else {
		healthyStr = colorize("false", ColorBoldRed)
	}

	return connectedStr, healthyStr
}
