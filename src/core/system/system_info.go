/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package system

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// GetTotalMemory returns total system memory in bytes using platform-specific syscalls
// Возвращает общий объем системной памяти в байтах используя platform-specific syscalls
func GetTotalMemory() int64 {
	// Strategy 1: Try Linux /proc/meminfo (most accurate)
	// Стратегия 1: Пробуем Linux /proc/meminfo (наиболее точно)
	if memTotal := getLinuxTotalMemory(); memTotal > 0 {
		return memTotal
	}

	// Strategy 2: Try syscall sysinfo on Linux
	// Стратегия 2: Пробуем syscall sysinfo на Linux
	if memTotal := getSysInfoTotalMemory(); memTotal > 0 {
		return memTotal
	}

	// Strategy 3: Try generic Unix syscall
	// Стратегия 3: Пробуем generic Unix syscall
	if memTotal := getUnixTotalMemory(); memTotal > 0 {
		return memTotal
	}

	// Fallback: Use Go runtime stats (least accurate but cross-platform)
	// Fallback: Используем Go runtime статистику (менее точно но cross-platform)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Sys)
}

// getLinuxTotalMemory reads total memory from /proc/meminfo on Linux
// Читает общую память из /proc/meminfo на Linux
func getLinuxTotalMemory() int64 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0 // Not Linux or /proc not available
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if kbValue, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					// Convert from KB to bytes
					return kbValue * 1024
				}
			}
		}
	}
	return 0
}

// getSysInfoTotalMemory uses Linux sysinfo syscall
// Использует Linux sysinfo syscall
func getSysInfoTotalMemory() int64 {
	type sysinfo struct {
		Uptime    int64
		Loads     [3]uint64
		Totalram  uint64
		Freeram   uint64
		Sharedram uint64
		Bufferram uint64
		Totalswap uint64
		Freeswap  uint64
		Procs     uint16
		Pad       uint16
		Totalhigh uint64
		Freehigh  uint64
		Unit      uint32
		_         [0]uint8 // Padding
	}

	var info sysinfo
	_, _, errno := syscall.Syscall(syscall.SYS_SYSINFO,
		uintptr(unsafe.Pointer(&info)), 0, 0)

	if errno != 0 {
		return 0 // syscall failed
	}

	// Convert to bytes (info.Unit is the memory unit)
	memBytes := int64(info.Totalram) * int64(info.Unit)
	if memBytes <= 0 {
		// Fallback if Unit is 0 or invalid - assume bytes
		memBytes = int64(info.Totalram)
	}

	return memBytes
}

// getUnixTotalMemory tries generic Unix approaches
// Пробует generic Unix подходы
func getUnixTotalMemory() int64 {
	// Try to read from /proc/sys/kernel/shmmax (sometimes available)
	if data, err := os.ReadFile("/proc/sys/kernel/shmmax"); err == nil {
		if value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil && value > 0 {
			// This is shared memory max, rough approximation of total memory
			return value
		}
	}

	// Other Unix-specific methods could be added here (BSD, macOS, etc.)
	return 0
}

// GetDiskSpace returns disk space information for given path
// Возвращает информацию о дисковом пространстве для указанного пути
func GetDiskSpace(path string) (total int64, free int64, err error) {
	var stat syscall.Statfs_t

	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// Calculate total and free space
	total = int64(stat.Blocks) * int64(stat.Bsize)
	free = int64(stat.Bavail) * int64(stat.Bsize)

	return total, free, nil
}

// GetSystemDiskSpace returns disk space for the system root
// Возвращает дисковое пространство для системного корня
func GetSystemDiskSpace() int64 {
	// Try to get disk space for root directory
	if total, _, err := GetDiskSpace("/"); err == nil {
		return total
	}

	// Fallback: try current working directory
	if wd, err := os.Getwd(); err == nil {
		if total, _, err := GetDiskSpace(wd); err == nil {
			return total
		}
	}

	// If all fails, return 0
	return 0
}

// GetMemoryInfo returns detailed memory information
// Возвращает детальную информацию о памяти
func GetMemoryInfo() (total int64, used int64, free int64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Use runtime memory stats as approximation
	total = int64(m.Sys)
	used = int64(m.Alloc)
	free = total - used

	return total, used, free
}
