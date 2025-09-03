/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"atom-engine/proto/process/processpb"
	"atom-engine/src/core/logger"
)

// TokenList lists tokens via gRPC
// Выводит список токенов через gRPC
func (d *DaemonCommand) TokenList() error {
	logger.Debug("Listing tokens")

	// Parse arguments for filtering
	var instanceID, state string

	args := os.Args[3:] // Skip "atomd token list"
	if len(args) > 0 && args[0] != "" {
		instanceID = args[0]
	}
	if len(args) > 1 && args[1] != "" {
		state = args[1]
	}

	logger.Debug("Token list request",
		logger.String("instance_id", instanceID),
		logger.String("state", state))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect to daemon for token list",
			logger.String("error", err.Error()))
		return fmt.Errorf("daemon is not running. Start daemon first with 'atomd start': %w", err)
	}
	defer conn.Close()

	// Create process gRPC client
	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call ListTokens gRPC method
	response, err := client.ListTokens(ctx, &processpb.ListTokensRequest{
		InstanceIdFilter: instanceID,
		StateFilter:      state,
		Limit:            0, // No limit
	})
	if err != nil {
		logger.Error("Failed to list tokens", logger.String("error", err.Error()))
		return fmt.Errorf("failed to list tokens: %w", err)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Message)
		return nil
	}

	fmt.Printf("Token List\n")
	fmt.Printf("==========\n")
	fmt.Printf("Found %d token(s):\n\n", len(response.Tokens))

	if len(response.Tokens) == 0 {
		fmt.Printf("No tokens found.\n")
		return nil
	}

	fmt.Printf("TOKEN ID                  PROCESS INSTANCE          CURRENT ELEMENT          STATE      WAITING FOR          CREATED              \n")
	fmt.Printf("------------------------- ------------------------- ------------------------ ---------- -------------------- --------------------\n")

	for _, token := range response.Tokens {
		waitingFor := token.WaitingFor
		if waitingFor == "" {
			waitingFor = "-"
		}

		createdAt := time.Unix(token.CreatedAt, 0).Format("2006-01-02 15:04:05")

		fmt.Printf("%-25s %-25s %-24s %-10s %-20s %s\n",
			token.TokenId,
			token.ProcessInstanceId,
			token.CurrentElementId,
			colorizeStatus(token.State),
			waitingFor,
			createdAt)
	}

	return nil
}

// TokenShow displays token details via gRPC
// Отображает детали токена через gRPC
func (d *DaemonCommand) TokenShow() error {
	logger.Debug("Showing token details")

	if len(os.Args) < 4 {
		logger.Error("Invalid token show arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd token show <token_id>")
	}

	tokenID := os.Args[3]
	logger.Debug("Token show request", logger.String("token_id", tokenID))

	fmt.Printf("Token Details:\n")
	fmt.Printf("Token ID: %s\n", tokenID)
	fmt.Printf("Note: Token details functionality needs to be implemented\n")

	return nil
}

// TokenTrace traces token execution path via gRPC
// Трассирует путь выполнения токена через gRPC
func (d *DaemonCommand) TokenTrace() error {
	logger.Debug("Tracing token execution")

	if len(os.Args) < 4 {
		logger.Error("Invalid token trace arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd token trace <instance_id>")
	}

	instanceID := os.Args[3]
	logger.Debug("Token trace request", logger.String("instance_id", instanceID))

	fmt.Printf("Token Trace:\n")
	fmt.Printf("Process Instance: %s\n", instanceID)
	fmt.Printf("Note: Token tracing functionality needs to be implemented\n")

	return nil
}
