/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// HttpConnectorExecutor executes HTTP connector tasks
type HttpConnectorExecutor struct {
	processComponent ComponentInterface
}

// NewHttpConnectorExecutor creates new HTTP connector executor
func NewHttpConnectorExecutor(processComponent ComponentInterface) *HttpConnectorExecutor {
	return &HttpConnectorExecutor{
		processComponent: processComponent,
	}
}

// GetElementType returns element type
func (hce *HttpConnectorExecutor) GetElementType() string {
	return "serviceTask"
}

// HttpConnectorConfig contains HTTP connector configuration
type HttpConnectorConfig struct {
	Method                     string                 `json:"method"`
	URL                        string                 `json:"url"`
	AuthenticationType         string                 `json:"authentication.type"`
	AuthenticationUsername     string                 `json:"authentication.username"`
	AuthenticationPassword     string                 `json:"authentication.password"`
	AuthenticationBearerToken  string                 `json:"authentication.bearerToken"`
	AuthenticationAPIKeyName   string                 `json:"authentication.apiKey.name"`
	AuthenticationAPIKeyValue  string                 `json:"authentication.apiKey.value"`
	AuthenticationAPIKeyIn     string                 `json:"authentication.apiKey.in"`
	QueryParameters            map[string]interface{} `json:"queryParameters"`
	Headers                    map[string]interface{} `json:"headers"`
	Body                       interface{}            `json:"body"`
	ConnectionTimeoutInSeconds int                    `json:"connectionTimeoutInSeconds"`
	ReadTimeoutInSeconds       int                    `json:"readTimeoutInSeconds"`
	WriteTimeoutInSeconds      int                    `json:"writeTimeoutInSeconds"`
	StoreResponse              bool                   `json:"storeResponse"`
}

// HttpConnectorResponse represents HTTP response
type HttpConnectorResponse struct {
	Status  int                    `json:"status"`
	Body    interface{}            `json:"body"`
	Headers map[string]interface{} `json:"headers"`
}

// ZeebeInput represents a zeebe input element
type ZeebeInput struct {
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

// ZeebeIOMapping represents ioMapping element structure
type ZeebeIOMapping struct {
	XMLName xml.Name     `xml:"ioMapping"`
	Inputs  []ZeebeInput `xml:"input"`
}

// Execute executes HTTP connector request
func (hce *HttpConnectorExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing HTTP connector",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract HTTP configuration from ioMapping
	config, err := hce.extractHttpConnectorConfig(element, token.Variables)
	if err != nil {
		logger.Error("Failed to extract HTTP connector configuration",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("HTTP connector configuration error: %v", err),
			Completed: false,
		}, nil
	}

	logger.Info("HTTP connector configuration extracted",
		logger.String("token_id", token.TokenID),
		logger.String("method", config.Method),
		logger.String("url", config.URL),
		logger.String("auth_type", config.AuthenticationType))

	// Execute HTTP request
	response, err := hce.executeHttpRequest(config)
	if err != nil {
		logger.Error("HTTP request failed",
			logger.String("token_id", token.TokenID),
			logger.String("url", config.URL),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("HTTP request failed: %v", err),
			Completed: false,
		}, nil
	}

	logger.Info("HTTP request completed",
		logger.String("token_id", token.TokenID),
		logger.String("url", config.URL),
		logger.Int("status", response.Status))

	// Log full response for debugging
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err == nil {
		logger.Debug("HTTP response details",
			logger.String("token_id", token.TokenID),
			logger.String("response", string(responseJSON)))
	} else {
		logger.Debug("HTTP response details",
			logger.String("token_id", token.TokenID),
			logger.Int("status", response.Status),
			logger.Any("body", response.Body),
			logger.Any("headers", response.Headers))
	}

	// Update token variables with response
	err = hce.updateTokenWithHttpResponse(token, response)
	if err != nil {
		logger.Error("Failed to update token with HTTP response",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("Failed to process HTTP response: %v", err),
			Completed: false,
		}, nil
	}

	// Get outgoing sequence flows
	outgoing, exists := element["outgoing"]
	if !exists {
		return &ExecutionResult{
			Success:      true,
			TokenUpdated: true,
			NextElements: []string{},
			Completed:    true,
		}, nil
	}

	var nextElements []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				nextElements = append(nextElements, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		nextElements = append(nextElements, outgoingStr)
	}

	return &ExecutionResult{
		Success:      true,
		TokenUpdated: true,
		NextElements: nextElements,
		Completed:    false,
	}, nil
}

// extractHttpConnectorConfig extracts HTTP connector configuration from ioMapping
func (hce *HttpConnectorExecutor) extractHttpConnectorConfig(
	element map[string]interface{},
	tokenVariables map[string]interface{},
) (*HttpConnectorConfig, error) {
	logger.Debug("Starting HTTP connector config extraction")

	config := &HttpConnectorConfig{
		Method:                     "GET",
		ConnectionTimeoutInSeconds: 20,
		ReadTimeoutInSeconds:       20,
		WriteTimeoutInSeconds:      20,
		StoreResponse:              false,
	}

	// Extract ioMapping inputs
	extensionElements, exists := element["extension_elements"]
	if !exists {
		logger.Debug("No extension elements found")
		return nil, fmt.Errorf("no extension elements found")
	}

	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		logger.Debug("Extension elements is not an array")
		return nil, fmt.Errorf("extension elements is not an array")
	}

	logger.Debug("Extension elements found", logger.Int("count", len(extElementsList)))

	// Find ioMapping in extension elements
	for i, extElement := range extElementsList {
		logger.Debug("Processing extension element", logger.Int("index", i))

		extElementMap, ok := extElement.(map[string]interface{})
		if !ok {
			logger.Debug("Extension element is not a map")
			continue
		}

		extensions, exists := extElementMap["extensions"]
		if !exists {
			logger.Debug("No extensions found in element")
			continue
		}

		extensionsList, ok := extensions.([]interface{})
		if !ok {
			logger.Debug("Extensions is not an array")
			continue
		}

		logger.Debug("Found extensions", logger.Int("count", len(extensionsList)))

		for j, ext := range extensionsList {
			logger.Debug("Processing extension", logger.Int("ext_index", j))

			extMap, ok := ext.(map[string]interface{})
			if !ok {
				logger.Debug("Extension is not a map")
				continue
			}

			extType, exists := extMap["type"]
			if !exists {
				logger.Debug("Extension has no type")
				continue
			}

			logger.Debug("Extension type found", logger.String("type", fmt.Sprintf("%v", extType)))

			if extType != "ioMapping" {
				continue
			}

			logger.Info("=== ABOUT TO CALL parseIOMappingInputs ===")
			logger.Debug("Found ioMapping extension, calling parseIOMappingInputs")
			// Found ioMapping - extract inputs
			result, err := hce.parseIOMappingInputs(extMap, tokenVariables, config)
			logger.Info("=== parseIOMappingInputs RETURNED ===", logger.Bool("error", err != nil))
			if result != nil {
				logger.Info("Result config",
					logger.String("method", result.Method),
					logger.String("url", result.URL),
					logger.String("auth_type", result.AuthenticationType))
			}
			return result, err
		}
	}

	logger.Debug("No ioMapping found, returning default config")
	return config, nil // Return default config if no ioMapping found
}

// parseIOMappingInputs parses ioMapping inputs and populates config
func (hce *HttpConnectorExecutor) parseIOMappingInputs(
	ioMappingExt map[string]interface{},
	tokenVariables map[string]interface{},
	config *HttpConnectorConfig,
) (*HttpConnectorConfig, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC in parseIOMappingInputs", logger.Any("panic", r))
		}
	}()

	logger.Info("=== ENTERING parseIOMappingInputs ===")
	logger.Info("ioMappingExt map size", logger.Int("size", len(ioMappingExt)))

	if ioMappingExt == nil {
		logger.Info("ioMappingExt is nil")
		return config, nil
	}

	// Initialize maps if not already done
	if config.QueryParameters == nil {
		config.QueryParameters = make(map[string]interface{})
	}
	if config.Headers == nil {
		config.Headers = make(map[string]interface{})
	}

	// Look for structured io_mapping from parser first
	if ioMappingData, exists := ioMappingExt["io_mapping"]; exists {
		logger.Info("Found structured io_mapping from parser")
		if ioMappingMap, ok := ioMappingData.(map[string]interface{}); ok {
			if inputs, exists := ioMappingMap["inputs"]; exists {
				if inputsList, ok := inputs.([]interface{}); ok {
					return hce.processInputs(inputsList, tokenVariables, config)
				}
			}
		}
	}

	// Get io_mapping_data (structured data from metadata parser)
	ioMappingData, exists := ioMappingExt["io_mapping_data"]
	if !exists {
		logger.Info("No io_mapping_data found, checking for other structures")
		logger.Info("Available ioMapping keys", logger.Any("keys", getMapKeys(ioMappingExt)))

		// Try to find inputs directly
		if inputs, inputsExists := ioMappingExt["inputs"]; inputsExists {
			logger.Info("Found inputs directly")
			return hce.processInputs(inputs, tokenVariables, config)
		}

		logger.Warn("No ioMapping inputs found in parser data - element may not have ioMapping configured")
		return config, nil
	}

	ioMappingMap, ok := ioMappingData.(map[string]interface{})
	if !ok {
		logger.Info("io_mapping_data is not a map", logger.Any("type", fmt.Sprintf("%T", ioMappingData)))
		return config, nil
	}

	logger.Info("Found structured io_mapping_data", logger.Any("keys", getMapKeys(ioMappingMap)))

	// Look for inputs in structured data
	if inputs, inputsExists := ioMappingMap["inputs"]; inputsExists {
		logger.Info("Found structured inputs")
		return hce.processInputs(inputs, tokenVariables, config)
	}

	logger.Info("No inputs found in io_mapping_data")
	return config, nil
}

// evaluateInputValue evaluates input value using expression component
func (hce *HttpConnectorExecutor) evaluateInputValue(source string, variables map[string]interface{}) interface{} {
	logger.Debug("Evaluating input value", logger.String("source", source))

	// Handle HTML-encoded JSON (&#34; = ") first
	if strings.Contains(source, "&#34;") {
		decoded := strings.ReplaceAll(source, "&#34;", "\"")
		logger.Debug("Decoded HTML entities", logger.String("decoded", decoded))
		source = decoded
	}

	// Handle JSON strings (check if it starts with { or [)
	if (strings.HasPrefix(source, "{") && strings.HasSuffix(source, "}")) ||
		(strings.HasPrefix(source, "[") && strings.HasSuffix(source, "]")) {
		logger.Debug("Attempting to parse as JSON", logger.String("source", source))

		var jsonValue interface{}
		if err := json.Unmarshal([]byte(source), &jsonValue); err == nil {
			logger.Debug("Successfully parsed JSON", logger.Any("value", jsonValue))
			return jsonValue
		}
		logger.Debug("Failed to parse as JSON, treating as string")
	}

	// Handle FEEL expressions (starting with =) using expression component
	if strings.HasPrefix(source, "=") {
		logger.Debug("Processing FEEL expression", logger.String("expression", source))

		// Get expression component through process component
		if hce.processComponent == nil {
			logger.Warn("Process component not available for expression evaluation")
			return source
		}

		// Get core interface
		core := hce.processComponent.GetCore()
		if core == nil {
			logger.Warn("Core interface not available for expression evaluation")
			return source
		}

		// Get expression component
		expressionCompInterface := core.GetExpressionComponent()
		if expressionCompInterface == nil {
			logger.Warn("Expression component not available")
			return source
		}

		// Cast to expression evaluator interface
		type ExpressionEvaluator interface {
			EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
		}

		expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
		if !ok {
			logger.Warn("Failed to cast expression component to ExpressionEvaluator interface")
			return source
		}

		// Evaluate FEEL expression using expression engine
		result, err := expressionComp.EvaluateExpressionEngine(source, variables)
		if err != nil {
			logger.Warn("Failed to evaluate FEEL expression",
				logger.String("expression", source),
				logger.String("error", err.Error()))
			return source
		}

		logger.Debug("FEEL expression evaluated successfully",
			logger.String("expression", source),
			logger.Any("result", result),
			logger.String("result_type", fmt.Sprintf("%T", result)))
		return result
	}

	// Handle literal values
	logger.Debug("Returning literal value", logger.String("value", source))
	return source
}

// executeHttpRequest executes the HTTP request with the given configuration
func (hce *HttpConnectorExecutor) executeHttpRequest(config *HttpConnectorConfig) (*HttpConnectorResponse, error) {
	// Create HTTP client with timeouts
	client := &http.Client{
		Timeout: time.Duration(config.ConnectionTimeoutInSeconds) * time.Second,
	}

	// Parse URL and add query parameters
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Add query parameters
	if len(config.QueryParameters) > 0 {
		query := parsedURL.Query()
		for key, value := range config.QueryParameters {
			query.Add(key, fmt.Sprintf("%v", value))
		}
		parsedURL.RawQuery = query.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != nil && (config.Method == "POST" || config.Method == "PUT" || config.Method == "PATCH") {
		logger.Debug("Preparing request body",
			logger.Any("body_value", config.Body),
			logger.String("body_type", fmt.Sprintf("%T", config.Body)),
		)
		bodyBytes, err := hce.prepareRequestBody(config.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare request body: %v", err)
		}
		logger.Debug("Request body prepared",
			logger.String("body_json", string(bodyBytes)),
			logger.Int("body_length", len(bodyBytes)),
		)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	req, err := http.NewRequest(config.Method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Content-Type header for JSON body
	if config.Body != nil && (config.Method == "POST" || config.Method == "PUT" || config.Method == "PATCH") {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add custom headers
	for key, value := range config.Headers {
		req.Header.Set(key, fmt.Sprintf("%v", value))
	}

	// Apply authentication
	err = hce.applyAuthentication(req, config)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response body as JSON if possible
	var bodyData interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &bodyData); err != nil {
			// If not JSON, store as string
			bodyData = string(respBody)
		}
	}

	// Convert response headers
	responseHeaders := make(map[string]interface{})
	for key, values := range resp.Header {
		if len(values) == 1 {
			responseHeaders[key] = values[0]
		} else {
			responseHeaders[key] = values
		}
	}

	return &HttpConnectorResponse{
		Status:  resp.StatusCode,
		Body:    bodyData,
		Headers: responseHeaders,
	}, nil
}

// prepareRequestBody prepares request body for HTTP request
func (hce *HttpConnectorExecutor) prepareRequestBody(body interface{}) ([]byte, error) {
	logger.Debug("prepareRequestBody called", logger.Any("body", body), logger.String("type", fmt.Sprintf("%T", body)))
	switch v := body.(type) {
	case string:
		// If it's already a string, return as is
		return []byte(v), nil
	case []byte:
		// If it's bytes, return as is
		return v, nil
	default:
		// Marshal as JSON
		return json.Marshal(body)
	}
}

// applyAuthentication applies authentication to the HTTP request
func (hce *HttpConnectorExecutor) applyAuthentication(req *http.Request, config *HttpConnectorConfig) error {
	switch config.AuthenticationType {
	case "basic":
		if config.AuthenticationUsername == "" || config.AuthenticationPassword == "" {
			return fmt.Errorf("basic authentication requires username and password")
		}
		auth := base64.StdEncoding.EncodeToString([]byte(config.AuthenticationUsername + ":" + config.AuthenticationPassword))
		req.Header.Set("Authorization", "Basic "+auth)

	case "bearer":
		if config.AuthenticationBearerToken == "" {
			return fmt.Errorf("bearer authentication requires token")
		}
		req.Header.Set("Authorization", "Bearer "+config.AuthenticationBearerToken)

	case "none", "":
		// No authentication needed

	default:
		return fmt.Errorf("unsupported authentication type: %s", config.AuthenticationType)
	}

	return nil
}

// updateTokenWithHttpResponse updates token variables with HTTP response
func (hce *HttpConnectorExecutor) updateTokenWithHttpResponse(
	token *models.Token,
	response *HttpConnectorResponse,
) error {
	// Create response object
	responseObj := map[string]interface{}{
		"status":  response.Status,
		"body":    response.Body,
		"headers": response.Headers,
	}

	// Update token variables
	if token.Variables == nil {
		token.Variables = make(map[string]interface{})
	}

	// Store response in 'response' variable (Camunda 8 standard)
	token.Variables["response"] = responseObj

	logger.Info("Updated token variables with HTTP response",
		logger.String("token_id", token.TokenID),
		logger.Int("response_status", response.Status))

	return nil
}

// extractTaskDefinition extracts task definition from element
func (hce *HttpConnectorExecutor) extractTaskDefinition(element map[string]interface{}) (*TaskDefinition, error) {
	// Look for extension elements
	extensionElements, exists := element["extension_elements"]
	if !exists {
		return nil, fmt.Errorf("no extension elements found")
	}

	// Parse extension elements as array
	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		return nil, fmt.Errorf("extension elements is not an array")
	}

	// Find taskDefinition in extension elements
	for _, extElement := range extElementsList {
		extElementMap, ok := extElement.(map[string]interface{})
		if !ok {
			continue
		}

		extensions, exists := extElementMap["extensions"]
		if !exists {
			continue
		}

		extensionsList, ok := extensions.([]interface{})
		if !ok {
			continue
		}

		for _, ext := range extensionsList {
			extMap, ok := ext.(map[string]interface{})
			if !ok {
				continue
			}

			extType, exists := extMap["type"]
			if !exists || extType != "taskDefinition" {
				continue
			}

			// Found taskDefinition - extract data
			taskDef, exists := extMap["task_definition"]
			if !exists {
				continue
			}

			taskDefMap, ok := taskDef.(map[string]interface{})
			if !ok {
				continue
			}

			jobType, _ := taskDefMap["type"].(string)
			if jobType == "" {
				return nil, fmt.Errorf("task definition missing type")
			}

			retries := 3 // default retries
			if retriesVal, exists := taskDefMap["retries"]; exists {
				if retriesInt, ok := retriesVal.(int); ok {
					retries = retriesInt
				}
			}

			return &TaskDefinition{
				Type:    jobType,
				Retries: retries,
			}, nil
		}
	}

	return nil, fmt.Errorf("taskDefinition not found in extension elements")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getMapKeys returns keys of a map as slice
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// processInputs processes inputs array and populates config
func (hce *HttpConnectorExecutor) processInputs(
	inputs interface{},
	tokenVariables map[string]interface{},
	config *HttpConnectorConfig,
) (*HttpConnectorConfig, error) {
	logger.Debug("Processing inputs")

	inputsList, ok := inputs.([]interface{})
	if !ok {
		logger.Debug("Inputs is not an array", logger.Any("inputs_type", fmt.Sprintf("%T", inputs)))
		return config, nil
	}

	logger.Debug("Found inputs array", logger.Int("count", len(inputsList)))

	for i, input := range inputsList {
		inputMap, ok := input.(map[string]interface{})
		if !ok {
			logger.Debug("Input is not a map", logger.Int("index", i))
			continue
		}

		source, sourceExists := inputMap["source"].(string)
		target, targetExists := inputMap["target"].(string)

		logger.Debug("Processing input",
			logger.Int("index", i),
			logger.String("source", source),
			logger.String("target", target),
			logger.Bool("source_exists", sourceExists),
			logger.Bool("target_exists", targetExists))

		if !sourceExists || !targetExists {
			continue
		}

		// Evaluate source value
		value := hce.evaluateInputValue(source, tokenVariables)
		logger.Debug("Evaluated input value",
			logger.String("source", source),
			logger.String("target", target),
			logger.Any("value", value))

		// Map to config fields
		switch target {
		case "method":
			config.Method = fmt.Sprintf("%v", value)
		case "url":
			config.URL = fmt.Sprintf("%v", value)
		case "authentication.type":
			config.AuthenticationType = fmt.Sprintf("%v", value)
		case "authentication.username":
			config.AuthenticationUsername = fmt.Sprintf("%v", value)
		case "authentication.password":
			config.AuthenticationPassword = fmt.Sprintf("%v", value)
		case "authentication.bearerToken":
			config.AuthenticationBearerToken = fmt.Sprintf("%v", value)
		case "authentication.apiKey.name":
			config.AuthenticationAPIKeyName = fmt.Sprintf("%v", value)
		case "authentication.apiKey.value":
			config.AuthenticationAPIKeyValue = fmt.Sprintf("%v", value)
		case "authentication.apiKey.in":
			config.AuthenticationAPIKeyIn = fmt.Sprintf("%v", value)
		case "body":
			config.Body = value
		case "connectionTimeoutInSeconds":
			if intVal, ok := value.(int); ok {
				config.ConnectionTimeoutInSeconds = intVal
			} else if strVal, ok := value.(string); ok {
				if intVal, err := strconv.Atoi(strVal); err == nil {
					config.ConnectionTimeoutInSeconds = intVal
				}
			}
		case "readTimeoutInSeconds":
			if intVal, ok := value.(int); ok {
				config.ReadTimeoutInSeconds = intVal
			} else if strVal, ok := value.(string); ok {
				if intVal, err := strconv.Atoi(strVal); err == nil {
					config.ReadTimeoutInSeconds = intVal
				}
			}
		case "storeResponse":
			if boolVal, ok := value.(bool); ok {
				config.StoreResponse = boolVal
			} else if strVal, ok := value.(string); ok {
				config.StoreResponse = strVal == "true"
			}
		default:
			logger.Debug("Unmapped target", logger.String("target", target), logger.Any("value", value))
		}
	}

	logger.Debug("Config after processing inputs",
		logger.String("method", config.Method),
		logger.String("url", config.URL),
		logger.String("auth_type", config.AuthenticationType),
		logger.Any("body", config.Body),
		logger.Any("headers", config.Headers))

	return config, nil
}
