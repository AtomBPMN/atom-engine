/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"fmt"
	"os"

	"atom-engine/src/core/logger"
)

// CLI handles command line interface
// Обработчик интерфейса командной строки
type CLI struct {
	daemon *DaemonCommand
}

// NewCLI creates new CLI instance
// Создает новый экземпляр CLI
func NewCLI() *CLI {
	return &CLI{
		daemon: NewDaemonCommand(),
	}
}

// Execute processes command line arguments
// Обрабатывает аргументы командной строки
func (c *CLI) Execute() error {
	if len(os.Args) < 2 {
		showHelp()
		return nil
	}

	command := os.Args[1]
	logger.Debug("Executing CLI command",
		logger.String("command", command),
		logger.Any("args", os.Args[2:]))

	switch command {
	case "start":
		return c.daemon.Start()
	case "run":
		return c.daemon.Run()
	case "stop":
		return c.daemon.Stop()
	case "status":
		return c.daemon.Status()
	case "events":
		return c.daemon.ShowEvents()
	case "storage":
		return c.handleStorageCommand()
	case "timer":
		return c.handleTimerCommand()
	case "process":
		return c.handleProcessCommand()
	case "token":
		return c.handleTokenCommand()
	case "job":
		return c.handleJobCommand()
	case "message":
		return c.handleMessageCommand()
	case "expression":
		return c.handleExpressionCommand()
	case "bpmn":
		return c.handleBPMNCommand()
	case "incident":
		return c.handleIncidentCommand()
	case "help", "--help", "-h":
		showHelp()
		return nil
	default:
		logger.Error("Unknown command", logger.String("command", command))
		return fmt.Errorf("unknown command: %s", command)
	}
}

// handleTimerCommand processes timer sub-commands
// Обрабатывает под-команды timer
func (c *CLI) handleTimerCommand() error {
	if len(os.Args) < 3 {
		showTimerHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing timer command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "add":
		return c.daemon.TimerAdd()
	case "remove":
		return c.daemon.TimerRemove()
	case "status":
		return c.daemon.TimerStatus()
	case "stats":
		return c.daemon.TimerStats()
	case "list":
		return c.daemon.TimerList()
	case "help", "--help", "-h":
		showTimerHelp()
		return nil
	default:
		logger.Error("Unknown timer command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown timer command: %s", subCommand)
	}
}

// handleStorageCommand processes storage sub-commands
// Обрабатывает под-команды storage
func (c *CLI) handleStorageCommand() error {
	if len(os.Args) < 3 {
		showStorageHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing storage command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "status":
		return c.daemon.StorageStatus()
	case "info":
		return c.daemon.StorageInfo()
	case "help", "--help", "-h":
		showStorageHelp()
		return nil
	default:
		logger.Error("Unknown storage command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown storage command: %s", subCommand)
	}
}

// handleProcessCommand processes process sub-commands
// Обрабатывает под-команды process
func (c *CLI) handleProcessCommand() error {
	if len(os.Args) < 3 {
		showProcessHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing process command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "start":
		return c.daemon.ProcessStart()
	case "status":
		return c.daemon.ProcessStatus()
	case "info":
		return c.daemon.ProcessInfo()
	case "cancel":
		return c.daemon.ProcessCancel()
	case "list":
		return c.daemon.ProcessList()
	case "help", "--help", "-h":
		showProcessHelp()
		return nil
	default:
		logger.Error("Unknown process command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown process command: %s", subCommand)
	}
}

// handleTokenCommand processes token sub-commands
// Обрабатывает под-команды token
func (c *CLI) handleTokenCommand() error {
	if len(os.Args) < 3 {
		showTokenHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing token command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "list":
		return c.daemon.TokenList()
	case "show":
		return c.daemon.TokenShow()
	case "trace":
		return c.daemon.TokenTrace()
	case "help", "--help", "-h":
		showTokenHelp()
		return nil
	default:
		logger.Error("Unknown token command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown token command: %s", subCommand)
	}
}

// handleJobCommand processes job sub-commands
// Обрабатывает под-команды job
func (c *CLI) handleJobCommand() error {
	if len(os.Args) < 3 {
		showJobHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing job command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "list":
		return c.daemon.JobList()
	case "show":
		return c.daemon.JobShow()
	case "activate":
		return c.daemon.JobActivate()
	case "complete":
		return c.daemon.JobComplete()
	case "fail":
		return c.daemon.JobFail()
	case "throw-error":
		return c.daemon.JobThrowError()
	case "cancel":
		return c.daemon.JobCancel()
	case "create":
		return c.daemon.JobCreate()
	case "stats":
		return c.daemon.JobStats()
	case "help", "--help", "-h":
		showJobHelp()
		return nil
	default:
		logger.Error("Unknown job command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown job command: %s", subCommand)
	}
}

// handleMessageCommand processes message sub-commands
// Обрабатывает под-команды message
func (c *CLI) handleMessageCommand() error {
	if len(os.Args) < 3 {
		showMessageHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing message command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "publish":
		return c.daemon.MessagePublish()
	case "list":
		return c.daemon.MessageList()
	case "subscriptions":
		return c.daemon.MessageSubscriptions()
	case "buffered":
		return c.daemon.MessageBuffered()
	case "cleanup":
		return c.daemon.MessageCleanup()
	case "stats":
		return c.daemon.MessageStats()
	case "test":
		return c.daemon.MessageTest()
	case "help", "--help", "-h":
		showMessageHelp()
		return nil
	default:
		logger.Error("Unknown message command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown message command: %s", subCommand)
	}
}

// handleExpressionCommand processes expression sub-commands
// Обрабатывает под-команды expression
func (c *CLI) handleExpressionCommand() error {
	if len(os.Args) < 3 {
		showExpressionHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing expression command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "eval":
		return c.daemon.ExpressionEvaluate()
	case "validate":
		return c.daemon.ExpressionValidate()
	case "parse":
		return c.daemon.ExpressionParse()
	case "functions":
		return c.daemon.ExpressionFunctions()
	case "test":
		return c.daemon.ExpressionTest()
	case "help", "--help", "-h":
		showExpressionHelp()
		return nil
	default:
		logger.Error("Unknown expression command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown expression command: %s", subCommand)
	}
}

// handleBPMNCommand processes BPMN sub-commands
// Обрабатывает под-команды bpmn
func (c *CLI) handleBPMNCommand() error {
	if len(os.Args) < 3 {
		showBPMNHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing BPMN command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "parse":
		return c.daemon.BPMNParse()
	case "list":
		return c.daemon.BPMNList()
	case "show":
		return c.daemon.BPMNShow()
	case "delete":
		return c.daemon.BPMNDelete()
	case "stats":
		return c.daemon.BPMNStats()
	case "json":
		return c.daemon.BPMNJson()
	case "xml":
		return c.daemon.BPMNXml()
	case "help", "--help", "-h":
		showBPMNHelp()
		return nil
	default:
		logger.Error("Unknown BPMN command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown bpmn command: %s", subCommand)
	}
}

// handleIncidentCommand processes incident sub-commands
// Обрабатывает под-команды incident
func (c *CLI) handleIncidentCommand() error {
	if len(os.Args) < 3 {
		showIncidentHelp()
		return nil
	}

	subCommand := os.Args[2]
	logger.Debug("Executing incident command", logger.String("subcommand", subCommand))

	switch subCommand {
	case "list":
		return c.daemon.IncidentList()
	case "show":
		return c.daemon.IncidentShow()
	case "resolve":
		return c.daemon.IncidentResolve()
	case "stats":
		return c.daemon.IncidentStats()
	case "help", "--help", "-h":
		showIncidentHelp()
		return nil
	default:
		logger.Error("Unknown incident command", logger.String("subcommand", subCommand))
		return fmt.Errorf("unknown incident command: %s", subCommand)
	}
}
