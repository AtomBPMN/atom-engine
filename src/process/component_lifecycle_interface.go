/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

// ComponentLifecycleInterface handles component lifecycle operations
// Интерфейс жизненного цикла компонента
type ComponentLifecycleInterface interface {
	// Lifecycle operations
	Init() error
	Start() error
	Stop() error
	IsReady() bool
}
