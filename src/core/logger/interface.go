/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package logger

// Logger interface for component logging
type ComponentLogger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// NewComponentLogger creates a new component logger using global logger
func NewComponentLogger(component string) ComponentLogger {
	return &componentLogger{component: component}
}

type componentLogger struct {
	component string
}

func (cl *componentLogger) Debug(msg string, fields ...Field) {
	allFields := append([]Field{String("component", cl.component)}, fields...)
	Debug(msg, allFields...)
}

func (cl *componentLogger) Info(msg string, fields ...Field) {
	allFields := append([]Field{String("component", cl.component)}, fields...)
	Info(msg, allFields...)
}

func (cl *componentLogger) Warn(msg string, fields ...Field) {
	allFields := append([]Field{String("component", cl.component)}, fields...)
	Warn(msg, allFields...)
}

func (cl *componentLogger) Error(msg string, fields ...Field) {
	allFields := append([]Field{String("component", cl.component)}, fields...)
	Error(msg, allFields...)
}

func (cl *componentLogger) Fatal(msg string, fields ...Field) {
	allFields := append([]Field{String("component", cl.component)}, fields...)
	Fatal(msg, allFields...)
}
