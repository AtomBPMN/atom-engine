/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package timewheel

import "fmt"

// Timer errors
// Ошибки таймера
var (
	ErrTimerTooFar         = fmt.Errorf("timer delay exceeds level horizon")
	ErrInvalidAnchor       = fmt.Errorf("invalid timer anchor")
	ErrTimerNotFound       = fmt.Errorf("timer not found")
	ErrTimerAlreadyExists  = fmt.Errorf("timer already exists")
	ErrInvalidConfig       = fmt.Errorf("invalid timing wheel configuration")
	ErrWheelNotRunning     = fmt.Errorf("timing wheel is not running")
	ErrWheelAlreadyRunning = fmt.Errorf("timing wheel is already running")
)

// ErrInvalidTimerRequest creates error for invalid timer request
// Создает ошибку для неверного запроса таймера
func ErrInvalidTimerRequest(msg string) error {
	return fmt.Errorf("invalid timer request: %s", msg)
}

// ErrTimerParsingFailed creates error for timer parsing failure
// Создает ошибку для неудачного парсинга таймера
func ErrTimerParsingFailed(timerType, value string, err error) error {
	return fmt.Errorf("failed to parse timer %s '%s': %w", timerType, value, err)
}

// ErrTimerSchedulingFailed creates error for timer scheduling failure
// Создает ошибку для неудачного планирования таймера
func ErrTimerSchedulingFailed(timerID string, err error) error {
	return fmt.Errorf("failed to schedule timer %s: %w", timerID, err)
}
