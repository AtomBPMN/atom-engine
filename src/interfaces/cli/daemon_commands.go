/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"atom-engine/src/core/config"
	"atom-engine/src/core/logger"
	"atom-engine/src/core/server"
	"atom-engine/src/storage"

	"github.com/dgraph-io/badger/v4"
)

// Start runs daemon in background mode
// Запускает демон в фоновом режиме
func (d *DaemonCommand) Start() error {
	logger.Info("Starting daemon in background mode")

	// Check if already running
	if d.isRunning() {
		logger.Warn("Daemon already running")
		return fmt.Errorf("daemon is already running")
	}

	// Start daemon process in background
	cmd := exec.Command(d.executable, "run")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	err := cmd.Start()
	if err != nil {
		logger.Error("Failed to start daemon process", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	logger.Info("Daemon process started", logger.Int("pid", cmd.Process.Pid))
	fmt.Printf("Daemon started with PID: %d\n", cmd.Process.Pid)

	// Write PID file
	err = d.writePIDFile(cmd.Process.Pid)
	if err != nil {
		logger.Error("Failed to write PID file", logger.String("error", err.Error()))
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Detach from parent process
	cmd.Process.Release()

	return nil
}

// Run starts daemon in foreground mode
// Запускает демон в режиме foreground
func (d *DaemonCommand) Run() error {
	logger.Info("Starting daemon in foreground mode")
	fmt.Println("Starting Atom Engine daemon...")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start core system
	err := d.startCore()
	if err != nil {
		logger.Error("Failed to start core system", logger.String("error", err.Error()))
		return fmt.Errorf("failed to start core: %w", err)
	}

	fmt.Println(ColorizeMessage("Atom Engine daemon is running"))

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("Shutting down Atom Engine daemon...")

	// Stop core system
	err = d.stopCore()
	if err != nil {
		logger.Error("Error during shutdown", logger.String("error", err.Error()))
		fmt.Printf("Error during shutdown: %v\n", err)
		return err
	}

	logger.Info("Daemon stopped gracefully")
	fmt.Println(ColorizeMessage("Atom Engine daemon stopped"))
	return nil
}

// Stop stops running daemon
// Останавливает работающий демон
func (d *DaemonCommand) Stop() error {
	logger.Info("Stopping daemon")

	pid, err := d.readPIDFile()
	if err != nil {
		logger.Warn("Daemon not running or PID file not found", logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running or PID file not found")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		logger.Error("Failed to find process",
			logger.Int("pid", pid),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to find process: %w", err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		logger.Error("Failed to send SIGTERM",
			logger.Int("pid", pid),
			logger.String("error", err.Error()))
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Remove PID file
	d.removePIDFile()

	logger.Info("Daemon stopped", logger.Int("pid", pid))
	fmt.Printf("Daemon with PID %d stopped\n", pid)
	return nil
}

// Status shows daemon status
// Показывает статус демона
func (d *DaemonCommand) Status() error {
	if d.isRunning() {
		pid, _ := d.readPIDFile()
		logger.Debug("Daemon status checked",
			logger.String("status", "running"),
			logger.Int("pid", pid))
		fmt.Printf("Daemon is %s with PID: %d\n", ColorizeDaemonStatus("running"), pid)
	} else {
		logger.Debug("Daemon status checked", logger.String("status", "not running"))
		fmt.Printf("Daemon is %s\n", ColorizeDaemonStatus("not running"))
	}
	return nil
}

// ShowEvents displays system events from database
// Показывает системные события из базы данных
func (d *DaemonCommand) ShowEvents() error {
	logger.Debug("Showing system events")
	return d.listEvents(nil)
}

// isRunning checks if daemon is running
// Проверяет работает ли демон
func (d *DaemonCommand) isRunning() bool {
	pid, err := d.readPIDFile()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// writePIDFile writes process ID to file
// Записывает ID процесса в файл
func (d *DaemonCommand) writePIDFile(pid int) error {
	pidFile := "/tmp/atomd.pid"
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", pid)
	return err
}

// readPIDFile reads process ID from file
// Читает ID процесса из файла
func (d *DaemonCommand) readPIDFile() (int, error) {
	pidFile := "/tmp/atomd.pid"
	content, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(content), "%d", &pid)
	return pid, err
}

// removePIDFile removes PID file
// Удаляет файл с PID
func (d *DaemonCommand) removePIDFile() {
	pidFile := "/tmp/atomd.pid"
	os.Remove(pidFile)
}

// startCore starts the core system
// Запускает основную систему
func (d *DaemonCommand) startCore() error {
	// Load configuration with environment variables
	cfg, err := config.LoadConfigWithEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize core with loaded config
	core, err := server.NewCoreWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create core: %w", err)
	}

	d.core = core

	// Start core system
	err = d.core.Start()
	if err != nil {
		return fmt.Errorf("failed to start core: %w", err)
	}

	fmt.Println(ColorizeMessage("Core system started"))
	return nil
}

// stopCore stops the core system
// Останавливает основную систему
func (d *DaemonCommand) stopCore() error {
	if d.core == nil {
		return nil
	}

	err := d.core.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop core: %w", err)
	}

	fmt.Println(ColorizeMessage("Core system stopped"))
	return nil
}

// SystemEventRecord represents system event record
// Представляет запись системного события
type SystemEventRecord struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// listEvents lists all system events from storage
// Выводит все системные события из storage
func (d *DaemonCommand) listEvents(storage storage.Storage) error {
	// Open BadgerDB directly for reading in read-only mode
	opts := badger.DefaultOptions("./data/badger")
	opts.Logger = nil
	opts.ReadOnly = true

	db, err := badger.Open(opts)
	if err != nil {
		logger.Error("Failed to open database for events", logger.String("error", err.Error()))
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	fmt.Println("System Events:")
	fmt.Println("==============")

	count := 0
	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("system_events:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var event SystemEventRecord
				err := json.Unmarshal(v, &event)
				if err != nil {
					fmt.Printf("Error unmarshaling event: %v\n", err)
					return nil
				}
				fmt.Printf("[%s] %s | %s | %s\n",
					event.CreatedAt,
					event.EventType,
					ColorizeEventStatus(event.Status),
					event.Message)
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to read events", logger.String("error", err.Error()))
		return fmt.Errorf("failed to read events: %w", err)
	}

	if count == 0 {
		fmt.Println("No events found")
	} else {
		fmt.Printf("\nTotal events: %d\n", count)
	}

	logger.Debug("Events listed", logger.Int("count", count))
	return nil
}
