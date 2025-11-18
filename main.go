/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package main

import (
	"log"
	"os"

	"atom-engine/src/interfaces/cli"
)

func main() {
	// Create CLI instance
	cliHandler := cli.NewCLI()

	// Execute command
	err := cliHandler.Execute()
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
