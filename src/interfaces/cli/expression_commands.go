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

	"atom-engine/proto/expression/expressionpb"
	"atom-engine/src/core/logger"
)

// ExpressionEvaluate evaluates an expression via gRPC
// Вычисляет выражение через gRPC
func (d *DaemonCommand) ExpressionEvaluate() error {
	logger.Debug("Evaluating expression")

	if len(os.Args) < 4 {
		logger.Error("Invalid expression eval arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd expression eval <expression> [context]")
	}

	expression := os.Args[3]
	contextData := ""
	if len(os.Args) > 4 {
		contextData = os.Args[4]
	}

	logger.Debug("Expression eval request",
		logger.String("expression", expression),
		logger.String("context", contextData))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for expression eval", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create expression gRPC client
	client := expressionpb.NewExpressionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request
	request := &expressionpb.EvaluateExpressionRequest{
		Expression: expression,
		Context:    contextData,
		TenantId:   "", // Default tenant
	}

	response, err := client.EvaluateExpression(ctx, request)
	if err != nil {
		logger.Error("Expression evaluation failed", logger.String("error", err.Error()))
		return fmt.Errorf("expression evaluation failed: %w", err)
	}

	fmt.Printf("Expression Evaluation\n")
	fmt.Printf("====================\n")
	fmt.Printf("Expression: %s\n", expression)
	if contextData != "" {
		fmt.Printf("Context: %s\n", contextData)
	}
	fmt.Printf("Result: %s\n", response.Result)
	fmt.Printf("Type: %s\n", response.ResultType)
	fmt.Printf("Success: %t\n", response.Success)
	if !response.Success {
		fmt.Printf("Error: %s\n", response.ErrorMessage)
	}

	return nil
}

// ExpressionValidate validates an expression via gRPC
// Валидирует выражение через gRPC
func (d *DaemonCommand) ExpressionValidate() error {
	logger.Debug("Validating expression")

	if len(os.Args) < 4 {
		logger.Error("Invalid expression validate arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd expression validate <expression> [schema]")
	}

	expression := os.Args[3]
	schema := ""
	if len(os.Args) > 4 {
		schema = os.Args[4]
	}

	logger.Debug("Expression validate request",
		logger.String("expression", expression),
		logger.String("schema", schema))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for expression validate", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create expression gRPC client
	client := expressionpb.NewExpressionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request
	request := &expressionpb.ValidateExpressionRequest{
		Expression:    expression,
		ContextSchema: schema,
	}

	response, err := client.ValidateExpression(ctx, request)
	if err != nil {
		logger.Error("Expression validation failed", logger.String("error", err.Error()))
		return fmt.Errorf("expression validation failed: %w", err)
	}

	fmt.Printf("Expression Validation\n")
	fmt.Printf("====================\n")
	fmt.Printf("Expression: %s\n", expression)
	if schema != "" {
		fmt.Printf("Schema: %s\n", schema)
	}
	fmt.Printf("Valid: %t\n", response.Valid)
	if !response.Valid {
		fmt.Printf("Error: %s\n", response.ErrorMessage)
	}
	if len(response.Errors) > 0 {
		fmt.Printf("Validation Errors:\n")
		for i, validationError := range response.Errors {
			fmt.Printf("  %d. %s (line %d, column %d)\n", i+1, validationError.Message, validationError.Line, validationError.Column)
		}
	}
	if len(response.Warnings) > 0 {
		fmt.Printf("Warnings:\n")
		for i, warning := range response.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
	}

	return nil
}

// ExpressionParse parses an expression to AST via gRPC
// Парсит выражение в AST через gRPC
func (d *DaemonCommand) ExpressionParse() error {
	logger.Debug("Parsing expression to AST")

	if len(os.Args) < 4 {
		logger.Error("Invalid expression parse arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd expression parse <expression>")
	}

	expression := os.Args[3]
	logger.Debug("Expression parse request", logger.String("expression", expression))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for expression parse", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create expression gRPC client
	client := expressionpb.NewExpressionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request
	request := &expressionpb.ParseExpressionRequest{
		Expression: expression,
	}

	response, err := client.ParseExpression(ctx, request)
	if err != nil {
		logger.Error("Expression parsing failed", logger.String("error", err.Error()))
		return fmt.Errorf("expression parsing failed: %w", err)
	}

	fmt.Printf("Expression Parsing\n")
	fmt.Printf("=================\n")
	fmt.Printf("Expression: %s\n", expression)
	fmt.Printf("Success: %t\n", response.Success)
	if response.Success {
		fmt.Printf("AST: %s\n", response.Ast)
		if len(response.Variables) > 0 {
			fmt.Printf("Variables found: %v\n", response.Variables)
		}
	} else {
		fmt.Printf("Error: %s\n", response.ErrorMessage)
	}

	return nil
}

// ExpressionFunctions lists supported functions via gRPC
// Выводит список поддерживаемых функций через gRPC
func (d *DaemonCommand) ExpressionFunctions() error {
	logger.Debug("Listing expression functions")

	category := ""
	if len(os.Args) > 3 {
		category = os.Args[3]
	}

	logger.Debug("Expression functions request", logger.String("category", category))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for expression functions", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create expression gRPC client
	client := expressionpb.NewExpressionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Make gRPC request
	request := &expressionpb.GetSupportedFunctionsRequest{
		Category: category,
	}

	response, err := client.GetSupportedFunctions(ctx, request)
	if err != nil {
		logger.Error("Get supported functions failed", logger.String("error", err.Error()))
		return fmt.Errorf("get supported functions failed: %w", err)
	}

	fmt.Printf("Expression Functions\n")
	fmt.Printf("===================\n")
	if category != "" {
		fmt.Printf("Category: %s\n\n", category)
	}

	if len(response.Functions) == 0 {
		fmt.Printf("No functions found")
		if category != "" {
			fmt.Printf(" for category '%s'", category)
		}
		fmt.Printf("\n")
		return nil
	}

	fmt.Printf("Found %d function(s):\n\n", len(response.Functions))

	// Group functions by category
	functionsByCategory := make(map[string][]*expressionpb.FunctionInfo)
	for _, fn := range response.Functions {
		functionsByCategory[fn.Category] = append(functionsByCategory[fn.Category], fn)
	}

	// Display functions by category
	for cat, functions := range functionsByCategory {
		fmt.Printf("%s functions:\n", cat)
		for _, fn := range functions {
			fmt.Printf("  %s(%s) -> %s\n", fn.Name, getParameterList(fn.Parameters), fn.ReturnType)
			fmt.Printf("    %s\n", fn.Description)
			if len(fn.Examples) > 0 {
				fmt.Printf("    Examples: %v\n", fn.Examples)
			}
			fmt.Printf("\n")
		}
	}

	return nil
}

// ExpressionTest tests an expression with test cases via gRPC
// Тестирует выражение с тестовыми случаями через gRPC
func (d *DaemonCommand) ExpressionTest() error {
	logger.Debug("Testing expression")

	if len(os.Args) < 5 {
		logger.Error("Invalid expression test arguments", logger.Int("args_count", len(os.Args)))
		return fmt.Errorf("usage: atomd expression test <expression> <test_cases>")
	}

	expression := os.Args[3]
	testCasesJSON := os.Args[4]

	logger.Debug("Expression test request",
		logger.String("expression", expression),
		logger.String("test_cases", testCasesJSON))

	conn, err := d.grpcClient.Connect()
	if err != nil {
		logger.Error("Failed to connect for expression test", logger.String("error", err.Error()))
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Create expression gRPC client
	client := expressionpb.NewExpressionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// For now, create a simple test case from provided JSON context
	// In real usage, this would parse test_cases JSON properly
	testCases := []*expressionpb.TestCase{
		{
			Name:           "test1",
			Context:        testCasesJSON,
			ExpectedResult: "true", // FIXME: Parse expected result from test cases JSON
		},
	}

	// Make gRPC request
	request := &expressionpb.TestExpressionRequest{
		Expression: expression,
		TestCases:  testCases,
	}

	response, err := client.TestExpression(ctx, request)
	if err != nil {
		logger.Error("Expression testing failed", logger.String("error", err.Error()))
		return fmt.Errorf("expression testing failed: %w", err)
	}

	fmt.Printf("Expression Testing\n")
	fmt.Printf("=================\n")
	fmt.Printf("Expression: %s\n", expression)
	fmt.Printf("Test Cases: %s\n", testCasesJSON)
	fmt.Printf("All Passed: %t\n", response.AllPassed)
	fmt.Printf("Summary: %s\n", response.Summary)

	if len(response.Results) > 0 {
		fmt.Printf("\nTest Results:\n")
		for i, result := range response.Results {
			fmt.Printf("  %d. %s: ", i+1, result.TestName)
			if result.Passed {
				fmt.Printf("PASSED\n")
			} else {
				fmt.Printf("FAILED\n")
				fmt.Printf("     Expected: %s\n", result.ExpectedResult)
				fmt.Printf("     Actual: %s\n", result.ActualResult)
				if result.ErrorMessage != "" {
					fmt.Printf("     Error: %s\n", result.ErrorMessage)
				}
			}
		}
	}

	return nil
}
