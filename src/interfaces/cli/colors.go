/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import "fmt"

// ANSI color codes for terminal output
// ANSI коды цветов для вывода в терминал
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// Bold colors
	ColorBoldRed    = "\033[1;31m"
	ColorBoldGreen  = "\033[1;32m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
	ColorBoldPurple = "\033[1;35m"
	ColorBoldCyan   = "\033[1;36m"
	ColorBoldWhite  = "\033[1;37m"
)

// colorize applies color to text and resets after
// Применяет цвет к тексту и сбрасывает после
func colorize(text, color string) string {
	return color + text + ColorReset
}

// colorizeStatus applies appropriate color based on status type
// Применяет соответствующий цвет в зависимости от типа статуса
func colorizeStatus(status string) string {
	switch status {
	// Active/Running states - green
	case "ACTIVE", "RUNNING", "Yes":
		return colorize(status, ColorBoldGreen)

	// Completed/Success states - blue
	case "COMPLETED", "SUCCESS", "DEPLOYED":
		return colorize(status, ColorBoldBlue)

	// Scheduled/Pending states - yellow
	case "SCHEDULED", "PENDING", "WAITING":
		return colorize(status, ColorBoldYellow)

	// Cancelled/Failed/Error states - red
	case "CANCELLED", "FAILED", "ERROR", "TIMEOUT":
		return colorize(status, ColorBoldRed)

	// Fired/Triggered states - purple
	case "FIRED", "TRIGGERED":
		return colorize(status, ColorBoldPurple)

	// Inactive/No states - gray
	case "INACTIVE", "No", "DISABLED":
		return colorize(status, ColorGray)

	// Default - no color
	default:
		return status
	}
}

// colorizeJobStatus applies color for job-specific statuses
// Применяет цвет для статусов заданий
func colorizeJobStatus(status string) string {
	switch status {
	case "AVAILABLE":
		return colorize(status, ColorBoldGreen)
	case "ACTIVATED":
		return colorize(status, ColorBoldCyan)
	case "COMPLETE":
		return colorize(status, ColorBoldBlue)
	case "FAILED":
		return colorize(status, ColorBoldRed)
	case "CANCELED":
		return colorize(status, ColorBoldRed)
	case "ERROR":
		return colorize(status, ColorRed)
	default:
		return colorizeStatus(status) // Fallback to general status coloring
	}
}

// colorizeTimerType applies color for timer types
// Применяет цвет для типов таймеров
func colorizeTimerType(timerType string) string {
	switch timerType {
	case "DURATION":
		return colorize(timerType, ColorCyan)
	case "CYCLE":
		return colorize(timerType, ColorPurple)
	case "DATE":
		return colorize(timerType, ColorBlue)
	case "BOUNDARY":
		return colorize(timerType, ColorYellow)
	default:
		return timerType
	}
}

// colorizeBool applies color for boolean values
// Применяет цвет для логических значений
func colorizeBool(value bool) string {
	if value {
		return colorize("Yes", ColorBoldGreen)
	}
	return colorize("No", ColorGray)
}

// colorizeRetries applies color for retry counts
// Применяет цвет для счетчиков повторов
func colorizeRetries(current, max int32) string {
	retriesInfo := fmt.Sprintf("%d/%d", current, max)

	if current == 0 {
		return colorize(retriesInfo, ColorGreen)
	} else if current < max {
		return colorize(retriesInfo, ColorYellow)
	} else {
		return colorize(retriesInfo, ColorRed)
	}
}

// colorizeLevel applies color for wheel levels
// Применяет цвет для уровней колеса
func colorizeLevel(level string) string {
	if level == "-1" || level == "N/A" {
		return colorize(level, ColorGray)
	}

	// Color by level type
	switch {
	case level[0:2] == "L0": // Seconds
		return colorize(level, ColorGreen)
	case level[0:2] == "L1": // Minutes
		return colorize(level, ColorCyan)
	case level[0:2] == "L2": // Hours
		return colorize(level, ColorBlue)
	case level[0:2] == "L3": // Days
		return colorize(level, ColorPurple)
	case level[0:2] == "L4": // Years
		return colorize(level, ColorRed)
	default:
		return level
	}
}
