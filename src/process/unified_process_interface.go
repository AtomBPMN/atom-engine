/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

// UnifiedProcessInterface combines all specialized manager interfaces
// Объединенный интерфейс всех специализированных менеджеров
type UnifiedProcessInterface interface {
	// Component lifecycle
	ComponentLifecycleInterface

	// Specialized managers
	ProcessManagerInterface
	TokenManagerInterface
	TimerCallbackManagerInterface
	JobCallbackManagerInterface
	MessageCallbackManagerInterface
}
