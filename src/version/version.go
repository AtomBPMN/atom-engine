/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package version

import (
	"runtime"
	"time"
)

// Build information variables
// These are set via ldflags during build process
// Переменные информации о сборке
// Устанавливаются через ldflags во время процесса сборки
var (
	Version   = "dev"     // Application version
	GitCommit = "unknown" // Git commit hash
	BuildTime = "unknown" // Build timestamp
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH
)

// GetBuildInfo returns build information
// Возвращает информацию о сборке
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_time": BuildTime,
		"go_version": GoVersion,
		"platform":   Platform,
	}
}

// GetBuildTime returns build time as time.Time
// Возвращает время сборки как time.Time
func GetBuildTime() time.Time {
	if BuildTime == "unknown" {
		return time.Now() // Fallback to current time
	}

	// Try to parse build time
	if t, err := time.Parse(time.RFC3339, BuildTime); err == nil {
		return t
	}

	// Fallback if parsing fails
	return time.Now()
}
