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
	"strings"
	"time"

	"atom-engine/proto/process/processpb"
	"atom-engine/src/core/logger"
)

// TokenList lists tokens via gRPC
// Выводит список токенов через gRPC
func (d *DaemonCommand) TokenList() error {
	logger.Debug("Listing tokens")

	// Parse arguments for filtering and pagination
	var instanceID, state string
	var pageSize, page int32 = 20, 1 // Default values

	args := os.Args[3:] // Skip "atomd token list"

	// Parse arguments: handle flags and positional arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--page" || arg == "-p" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &page); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if arg == "--page-size" || arg == "-s" {
			if i+1 < len(args) {
				if p, err := fmt.Sscanf(args[i+1], "%d", &pageSize); err == nil && p == 1 {
					i++ // Skip the next argument as it's the value
					continue
				}
			}
		} else if !strings.HasPrefix(arg, "--") && !strings.HasPrefix(arg, "-") {
			// Positional arguments
			if instanceID == "" {
				instanceID = arg
			} else if state == "" {
				state = arg
			}
		}
	}

	logger.Debug("Token list request",
		logger.String("instance_id", instanceID),
		logger.String("state", state),
		logger.Int("page_size", int(pageSize)),
		logger.Int("page", int(page)))

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
		Limit:            0, // Legacy field for backward compatibility
		PageSize:         pageSize,
		Page:             page,
		SortBy:           "created_at",
		SortOrder:        "DESC",
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

	// Display pagination information
	if response.TotalPages > 1 {
		fmt.Printf("Page %d of %d (Total: %d tokens, Showing: %d)\n\n",
			response.Page, response.TotalPages, response.TotalCount, len(response.Tokens))
	} else {
		fmt.Printf("Found %d token(s):\n\n", response.TotalCount)
	}

	if len(response.Tokens) == 0 {
		fmt.Printf("No tokens found.\n")
		if response.TotalPages > 1 {
			fmt.Printf("Try a different page number with --page <N>\n")
		}
		return nil
	}

	fmt.Printf("TOKEN ID                  PROCESS INSTANCE          CURRENT ELEMENT          ")
	fmt.Printf("STATE      WAITING FOR          CREATED              \n")
	fmt.Printf("------------------------- ------------------------- ------------------------ ")
	fmt.Printf("---------- -------------------- --------------------\n")

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

	// Display navigation hints for pagination
	if response.TotalPages > 1 {
		fmt.Printf("\n")
		if response.Page > 1 {
			fmt.Printf("Previous page: atomd token list")
			if instanceID != "" {
				fmt.Printf(" %s", instanceID)
			}
			if state != "" {
				fmt.Printf(" %s", state)
			}
			fmt.Printf(" --page %d", response.Page-1)
			if pageSize != 20 {
				fmt.Printf(" --page-size %d", pageSize)
			}
			fmt.Printf("\n")
		}
		if response.Page < response.TotalPages {
			fmt.Printf("Next page: atomd token list")
			if instanceID != "" {
				fmt.Printf(" %s", instanceID)
			}
			if state != "" {
				fmt.Printf(" %s", state)
			}
			fmt.Printf(" --page %d", response.Page+1)
			if pageSize != 20 {
				fmt.Printf(" --page-size %d", pageSize)
			}
			fmt.Printf("\n")
		}
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

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for token show", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create process gRPC client
	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request
	request := &processpb.GetTokenStatusRequest{
		TokenId: tokenID,
	}

	response, err := client.GetTokenStatus(ctx, request)
	if err != nil {
		logger.Error("Get token status failed", logger.String("error", err.Error()))
		return fmt.Errorf("get token status failed: %w", err)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Message)
		return nil
	}

	token := response.Token

	fmt.Printf("Token Details\n")
	fmt.Printf("=============\n")
	fmt.Printf("Token ID:             %s\n", token.TokenId)
	fmt.Printf("Process Instance ID:  %s\n", token.ProcessInstanceId)
	fmt.Printf("Process Key:          %s\n", token.ProcessKey)
	fmt.Printf("Current Element ID:   %s\n", token.CurrentElementId)
	fmt.Printf("State:                %s\n", colorizeStatus(token.State))
	fmt.Printf("Waiting For:          %s\n", token.WaitingFor)
	fmt.Printf("Created At:           %s\n", time.Unix(token.CreatedAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:           %s\n", time.Unix(token.UpdatedAt, 0).Format("2006-01-02 15:04:05"))

	if len(token.Variables) > 0 {
		fmt.Printf("\nVariables:\n")
		for key, value := range token.Variables {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

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

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for token trace", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create process gRPC client
	client := processpb.NewProcessServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request to list tokens for this process instance
	request := &processpb.ListTokensRequest{
		InstanceIdFilter: instanceID,
		StateFilter:      "",  // Get all tokens regardless of state
		PageSize:         100, // Get more tokens for tracing
		Page:             1,
		SortBy:           "created_at",
		SortOrder:        "ASC", // Sort by creation time ascending
	}

	response, err := client.ListTokens(ctx, request)
	if err != nil {
		logger.Error("List tokens failed", logger.String("error", err.Error()))
		return fmt.Errorf("list tokens failed: %w", err)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Message)
		return nil
	}

	fmt.Printf("Token Trace\n")
	fmt.Printf("===========\n")
	fmt.Printf("Process Instance: %s\n", instanceID)
	fmt.Printf("Total Tokens: %d\n\n", response.TotalCount)

	if len(response.Tokens) == 0 {
		fmt.Printf("No tokens found for process instance\n")
		return nil
	}

	fmt.Printf("Execution Trace (chronological order):\n")
	fmt.Printf("======================================\n")

	for i, token := range response.Tokens {
		fmt.Printf("%d. Token %s\n", i+1, token.TokenId)
		fmt.Printf("   Element:    %s\n", token.CurrentElementId)
		fmt.Printf("   State:      %s\n", colorizeStatus(token.State))
		fmt.Printf("   Created:    %s\n", time.Unix(token.CreatedAt, 0).Format("2006-01-02 15:04:05"))
		if token.UpdatedAt != token.CreatedAt {
			fmt.Printf("   Updated:    %s\n", time.Unix(token.UpdatedAt, 0).Format("2006-01-02 15:04:05"))
		}
		if token.WaitingFor != "" {
			fmt.Printf("   Waiting:    %s\n", token.WaitingFor)
		}
		fmt.Printf("\n")
	}

	// Show flow summary
	fmt.Printf("Flow Summary:\n")
	fmt.Printf("=============\n")
	elements := make([]string, 0)
	for _, token := range response.Tokens {
		if token.CurrentElementId != "" {
			// Avoid duplicates in summary
			found := false
			for _, elem := range elements {
				if elem == token.CurrentElementId {
					found = true
					break
				}
			}
			if !found {
				elements = append(elements, token.CurrentElementId)
			}
		}
	}

	if len(elements) > 0 {
		fmt.Printf("Elements visited: %s\n", strings.Join(elements, " → "))
	}

	return nil
}
