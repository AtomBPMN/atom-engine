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

	// Create boundary timers when token enters activity
	if err := hce.createBoundaryTimers(token, element); err != nil {
		logger.Error("Failed to create boundary timers",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
	}

	// Create error boundary subscriptions when token enters activity
	logger.Info("About to create error boundary subscriptions",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	if err := hce.createErrorBoundaries(token, element); err != nil {
		logger.Error("Failed to create error boundary subscriptions",
			logger.String("token_id", token.TokenID),
			logger.String("element_id", token.CurrentElementID),
			logger.String("error", err.Error()))
	}

	logger.Info("Completed error boundary subscriptions processing",
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

	// Apply output mapping (Camunda 8 standard)
	// Применяем output mapping (стандарт Camunda 8)
	err = hce.applyOutputMapping(element, token)
	if err != nil {
		logger.Warn("Failed to apply output mapping",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		// Continue execution even if output mapping fails
		// Продолжаем выполнение даже если output mapping не сработал
	}

	// Log all available variables for debugging
	// Логируем все доступные переменные для отладки
	hce.logAvailableVariables(token)

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
	err = hce.applyAuthentication(req, config, parsedURL)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	// Log request details after authentication (so we can see all headers including auth)
	// Логируем детали запроса после аутентификации (чтобы видеть все заголовки включая auth)
	hce.logHttpRequest(req, config, parsedURL)

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

	// Log response details
	hce.logHttpResponse(resp, respBody, bodyData, responseHeaders)

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
		// Try to parse as JSON first
		// Пытаемся сначала распарсить как JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(v), &jsonData); err == nil {
			// Valid JSON - marshal it back to ensure proper formatting
			// Валидный JSON - маршалим обратно для правильного форматирования
			return json.Marshal(jsonData)
		}
		// If not valid JSON, check if it looks like JSON object without braces
		// Если не валидный JSON, проверяем похож ли на JSON объект без фигурных скобок
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, `"`) && strings.Contains(trimmed, ":") {
			// Looks like JSON object property, wrap in braces
			// Похоже на свойство JSON объекта, оборачиваем в фигурные скобки
			wrapped := "{" + trimmed + "}"
			if err := json.Unmarshal([]byte(wrapped), &jsonData); err == nil {
				// Successfully wrapped, return as JSON
				// Успешно обернули, возвращаем как JSON
				return json.Marshal(jsonData)
			}
		}
		// Return as is if can't parse
		// Возвращаем как есть если не можем распарсить
		return []byte(v), nil
	case []byte:
		// Try to parse as JSON first
		// Пытаемся сначала распарсить как JSON
		var jsonData interface{}
		if err := json.Unmarshal(v, &jsonData); err == nil {
			// Valid JSON - marshal it back to ensure proper formatting
			// Валидный JSON - маршалим обратно для правильного форматирования
			return json.Marshal(jsonData)
		}
		// Return as is if not valid JSON
		// Возвращаем как есть если не валидный JSON
		return v, nil
	default:
		// Marshal as JSON
		return json.Marshal(body)
	}
}

// applyAuthentication applies authentication to the HTTP request
func (hce *HttpConnectorExecutor) applyAuthentication(req *http.Request, config *HttpConnectorConfig, parsedURL *url.URL) error {
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

	case "apiKey":
		if config.AuthenticationAPIKeyName == "" || config.AuthenticationAPIKeyValue == "" {
			return fmt.Errorf("apiKey authentication requires name and value")
		}
		// Support apiKey in headers, query, or cookie
		// Поддерживаем apiKey в заголовках, query параметрах или cookies
		switch strings.ToLower(config.AuthenticationAPIKeyIn) {
		case "header", "headers":
			req.Header.Set(config.AuthenticationAPIKeyName, config.AuthenticationAPIKeyValue)
		case "query":
			if parsedURL.RawQuery == "" {
				parsedURL.RawQuery = fmt.Sprintf("%s=%s", config.AuthenticationAPIKeyName, url.QueryEscape(config.AuthenticationAPIKeyValue))
			} else {
				parsedURL.RawQuery += fmt.Sprintf("&%s=%s", config.AuthenticationAPIKeyName, url.QueryEscape(config.AuthenticationAPIKeyValue))
			}
			req.URL = parsedURL
		case "cookie":
			req.AddCookie(&http.Cookie{
				Name:  config.AuthenticationAPIKeyName,
				Value: config.AuthenticationAPIKeyValue,
			})
		default:
			// Default to header if location not specified
			// По умолчанию используем заголовок если местоположение не указано
			req.Header.Set(config.AuthenticationAPIKeyName, config.AuthenticationAPIKeyValue)
		}

	case "none", "":
		// No authentication needed

	default:
		return fmt.Errorf("unsupported authentication type: %s", config.AuthenticationType)
	}

	return nil
}

// logHttpRequest logs detailed HTTP request information
func (hce *HttpConnectorExecutor) logHttpRequest(
	req *http.Request,
	config *HttpConnectorConfig,
	parsedURL *url.URL,
) {
	var logLines []string
	logLines = append(logLines, "======REQUEST========")
	logLines = append(logLines, fmt.Sprintf("METHOD: %s", req.Method))
	logLines = append(logLines, fmt.Sprintf("URL: %s", parsedURL.String()))

	// Query parameters
	if parsedURL.RawQuery != "" {
		logLines = append(logLines, "PARAM:")
		queryParams := parsedURL.Query()
		for key, values := range queryParams {
			for _, value := range values {
				logLines = append(logLines, fmt.Sprintf("  %s = %s", key, value))
			}
		}
	} else {
		logLines = append(logLines, "PARAM: (none)")
	}

	// Headers
	logLines = append(logLines, "HEADER:")
	if len(req.Header) > 0 {
		for key, values := range req.Header {
			for _, value := range values {
				// Mask sensitive headers
				if strings.ToLower(key) == "authorization" {
					if len(value) > 20 {
						logLines = append(logLines, fmt.Sprintf("  %s = %s...", key, value[:20]))
					} else {
						logLines = append(logLines, fmt.Sprintf("  %s = ***", key))
					}
				} else {
					logLines = append(logLines, fmt.Sprintf("  %s = %s", key, value))
				}
			}
		}
	} else {
		logLines = append(logLines, "  (none)")
	}

	// Body
	logLines = append(logLines, "PAYLOAD:")
	if config.Body != nil {
		bodyBytes, err := hce.prepareRequestBody(config.Body)
		if err == nil {
			bodyStr := string(bodyBytes)
			// Try to format as JSON if possible
			var jsonData interface{}
			if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
				if formatted, err := json.MarshalIndent(jsonData, "  ", "  "); err == nil {
					bodyStr = string(formatted)
				}
			}
			logLines = append(logLines, fmt.Sprintf("  %s", bodyStr))
		} else {
			logLines = append(logLines, fmt.Sprintf("  (error preparing body: %v)", err))
		}
	} else {
		logLines = append(logLines, "  (empty)")
	}

	logLines = append(logLines, "====================")

	// Log each line separately for better readability
	// Логируем каждую строку отдельно для лучшей читаемости
	for _, line := range logLines {
		logger.Debug(line)
	}
}

// logHttpResponse logs detailed HTTP response information
func (hce *HttpConnectorExecutor) logHttpResponse(
	resp *http.Response,
	respBody []byte,
	bodyData interface{},
	responseHeaders map[string]interface{},
) {
	var logLines []string
	logLines = append(logLines, "======RESPONSE==============")
	logLines = append(logLines, fmt.Sprintf("STATUS: %d %s", resp.StatusCode, resp.Status))

	// Response headers
	logLines = append(logLines, "HEADER:")
	if len(responseHeaders) > 0 {
		for key, value := range responseHeaders {
			logLines = append(logLines, fmt.Sprintf("  %s = %v", key, value))
		}
	} else {
		logLines = append(logLines, "  (none)")
	}

	// Response body
	logLines = append(logLines, "BODY:")
	if bodyData != nil {
		var bodyStr string
		if bodyBytes, ok := bodyData.([]byte); ok {
			bodyStr = string(bodyBytes)
		} else if bodyStrVal, ok := bodyData.(string); ok {
			bodyStr = bodyStrVal
		} else {
			// Try to format as JSON
			if formatted, err := json.MarshalIndent(bodyData, "  ", "  "); err == nil {
				bodyStr = string(formatted)
			} else {
				bodyStr = fmt.Sprintf("%v", bodyData)
			}
		}

		// Try to format as JSON if it's a string that looks like JSON
		if bodyStr != "" {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(bodyStr), &jsonData); err == nil {
				if formatted, err := json.MarshalIndent(jsonData, "  ", "  "); err == nil {
					bodyStr = string(formatted)
				}
			}
			logLines = append(logLines, fmt.Sprintf("  %s", bodyStr))
		} else {
			logLines = append(logLines, "  (empty)")
		}
	} else {
		logLines = append(logLines, "  (empty)")
	}

	logLines = append(logLines, "=============================")

	// Log each line separately for better readability
	// Логируем каждую строку отдельно для лучшей читаемости
	for _, line := range logLines {
		logger.Debug(line)
	}
}

// logAvailableVariables logs all available variables from HTTP response
// Логирует все доступные переменные из HTTP ответа
func (hce *HttpConnectorExecutor) logAvailableVariables(
	token *models.Token,
) {
	if token.Variables == nil {
		return
	}

	var logLines []string
	logLines = append(logLines, "")
	logLines = append(logLines, "======AVAILABLE VARIABLES========")
	logLines = append(logLines, "Variables you can use in process:")
	logLines = append(logLines, "")

	// Get response object from variables
	responseObj, hasResponse := token.Variables["response"]
	if !hasResponse {
		logLines = append(logLines, "  (no response variable found)")
	} else {
		// Log response.status
		if responseMap, ok := responseObj.(map[string]interface{}); ok {
			if status, ok := responseMap["status"]; ok {
				logLines = append(logLines, fmt.Sprintf("  response.status = %v", status))
			}

			// Log response.body fields
			if body, ok := responseMap["body"]; ok {
				if bodyMap, ok := body.(map[string]interface{}); ok {
					hce.logMapVariables("response.body", bodyMap, &logLines, 1)
				} else {
					logLines = append(logLines, fmt.Sprintf("  response.body = %v", body))
				}
			}

			// Log response.headers fields
			if headers, ok := responseMap["headers"]; ok {
				if headersMap, ok := headers.(map[string]interface{}); ok {
					logLines = append(logLines, "")
					logLines = append(logLines, "  Response headers:")
					for key, value := range headersMap {
						logLines = append(logLines, fmt.Sprintf("    response.headers.%s = %v", key, value))
					}
				}
			}
		}
	}

	// Log other process variables (except response)
	hasOtherVars := false
	for key, value := range token.Variables {
		if key == "response" {
			continue
		}
		if !hasOtherVars {
			logLines = append(logLines, "")
			logLines = append(logLines, "  Other process variables:")
			hasOtherVars = true
		}
		logLines = append(logLines, fmt.Sprintf("    %s = %v", key, value))
	}

	logLines = append(logLines, "")
	logLines = append(logLines, "=================================")

	// Log each line separately
	for _, line := range logLines {
		logger.Debug(line)
	}
}

// logMapVariables recursively logs map variables with proper nesting
// Рекурсивно логирует переменные map с правильной вложенностью
func (hce *HttpConnectorExecutor) logMapVariables(
	prefix string,
	data map[string]interface{},
	logLines *[]string,
	depth int,
) {
	if depth > 5 {
		// Prevent infinite recursion
		// Предотвращаем бесконечную рекурсию
		return
	}

	for key, value := range data {
		fullKey := fmt.Sprintf("%s.%s", prefix, key)
		
		switch v := value.(type) {
		case map[string]interface{}:
			// Nested map
			*logLines = append(*logLines, fmt.Sprintf("  %s = {object}", fullKey))
			hce.logMapVariables(fullKey, v, logLines, depth+1)
		case []interface{}:
			// Array
			*logLines = append(*logLines, fmt.Sprintf("  %s = [array with %d items]", fullKey, len(v)))
			// Log array items
			for i, item := range v {
				itemKey := fmt.Sprintf("%s[%d]", fullKey, i)
				if itemMap, ok := item.(map[string]interface{}); ok {
					*logLines = append(*logLines, fmt.Sprintf("  %s = {object}", itemKey))
					hce.logMapVariables(itemKey, itemMap, logLines, depth+1)
				} else {
					*logLines = append(*logLines, fmt.Sprintf("  %s = %v", itemKey, item))
				}
			}
		case string:
			// String value - show with quotes
			*logLines = append(*logLines, fmt.Sprintf("  %s = \"%s\"", fullKey, v))
		default:
			// Other types
			*logLines = append(*logLines, fmt.Sprintf("  %s = %v", fullKey, v))
		}
	}
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

// applyOutputMapping applies output mapping to token variables (Camunda 8 standard)
// Применяет output mapping к переменным токена (стандарт Camunda 8)
func (hce *HttpConnectorExecutor) applyOutputMapping(
	element map[string]interface{},
	token *models.Token,
) error {
	logger.Debug("Applying output mapping",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	// Extract output mappings from element
	outputs := hce.extractOutputMappings(element)
	if len(outputs) == 0 {
		logger.Debug("No output mappings found",
			logger.String("element_id", token.CurrentElementID))
		return nil
	}

	logger.Info("Found output mappings",
		logger.String("element_id", token.CurrentElementID),
		logger.Int("count", len(outputs)))

	// Ensure token variables are initialized
	if token.Variables == nil {
		token.Variables = make(map[string]interface{})
	}

	// Apply each output mapping
	for i, output := range outputs {
		sourceRaw, sourceExists := output["source"]
		targetRaw, targetExists := output["target"]

		if !sourceExists || !targetExists {
			logger.Warn("Output mapping missing source or target",
				logger.Int("index", i),
				logger.Any("output", output))
			continue
		}

		source, sourceOK := sourceRaw.(string)
		target, targetOK := targetRaw.(string)

		if !sourceOK || !targetOK {
			logger.Warn("Output mapping source or target is not a string",
				logger.Int("index", i),
				logger.Any("source", sourceRaw),
				logger.Any("target", targetRaw))
			continue
		}

		logger.Debug("Processing output mapping",
			logger.Int("index", i),
			logger.String("source", source),
			logger.String("target", target))

		// Evaluate source expression using token variables
		value := hce.evaluateInputValue(source, token.Variables)

		// Set target variable
		token.Variables[target] = value

		logger.Info("Output mapping applied",
			logger.String("source", source),
			logger.String("target", target),
			logger.Any("value", value))
	}

	return nil
}

// extractOutputMappings extracts output mappings from element
// Извлекает output mappings из элемента
func (hce *HttpConnectorExecutor) extractOutputMappings(
	element map[string]interface{},
) []map[string]interface{} {
	outputs := make([]map[string]interface{}, 0)

	// Get extension elements
	extensionElements, exists := element["extension_elements"]
	if !exists {
		return outputs
	}

	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		return outputs
	}

	// Find ioMapping in extension elements
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
			if !exists || extType != "ioMapping" {
				continue
			}

			// Found ioMapping - look for structured data first
			if ioMapping, exists := extMap["io_mapping"]; exists {
				if ioMappingMap, ok := ioMapping.(map[string]interface{}); ok {
					if outputsData, exists := ioMappingMap["outputs"]; exists {
						if outputsList, ok := outputsData.([]interface{}); ok {
							for _, output := range outputsList {
								if outputMap, ok := output.(map[string]interface{}); ok {
									outputs = append(outputs, outputMap)
								}
							}
						}
					}
				}
			}

			// Also check io_mapping_data (legacy structure)
			if ioMappingData, exists := extMap["io_mapping_data"]; exists {
				if ioMappingMap, ok := ioMappingData.(map[string]interface{}); ok {
					if outputsData, exists := ioMappingMap["outputs"]; exists {
						if outputsList, ok := outputsData.([]interface{}); ok {
							for _, output := range outputsList {
								if outputMap, ok := output.(map[string]interface{}); ok {
									outputs = append(outputs, outputMap)
								}
							}
						}
					}
				}
			}
		}
	}

	logger.Debug("Extracted output mappings",
		logger.Int("count", len(outputs)),
		logger.Any("outputs", outputs))

	return outputs
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
		case "authentication.apiKeyLocation":
			// Support both apiKeyLocation and apiKey.in
			// Поддерживаем и apiKeyLocation и apiKey.in
			config.AuthenticationAPIKeyIn = fmt.Sprintf("%v", value)
		case "authentication.name":
			// Support both authentication.name and authentication.apiKey.name
			// Поддерживаем и authentication.name и authentication.apiKey.name
			config.AuthenticationAPIKeyName = fmt.Sprintf("%v", value)
		case "authentication.value":
			// Support both authentication.value and authentication.apiKey.value
			// Поддерживаем и authentication.value и authentication.apiKey.value
			config.AuthenticationAPIKeyValue = fmt.Sprintf("%v", value)
		case "body":
			config.Body = value
		case "queryParameters":
			// Handle queryParameters - can be a map or object
			// Обрабатываем queryParameters - может быть map или объект
			if queryMap, ok := value.(map[string]interface{}); ok {
				// Direct map - add all key-value pairs to QueryParameters
				// Прямой map - добавляем все пары ключ-значение в QueryParameters
				for key, val := range queryMap {
					config.QueryParameters[key] = val
				}
				logger.Debug("Added query parameters from map",
					logger.Int("count", len(queryMap)),
					logger.Any("params", queryMap))
			} else {
				// Try to convert to map
				// Пытаемся преобразовать в map
				logger.Debug("Query parameters value is not a map",
					logger.String("type", fmt.Sprintf("%T", value)),
					logger.Any("value", value))
			}
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

// createBoundaryTimers creates boundary timers for activity
func (hce *HttpConnectorExecutor) createBoundaryTimers(token *models.Token, element map[string]interface{}) error {
	if hce.processComponent == nil {
		return nil
	}

	bpmnProcess, err := hce.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	boundaryEvents := hce.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil
	}

	logger.Info("Found boundary events for activity",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	for eventID, boundaryEvent := range boundaryEvents {
		if err := hce.createBoundaryTimerForEvent(token, eventID, boundaryEvent); err != nil {
			logger.Error("Failed to create boundary timer",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue
		}
	}

	return nil
}

// findBoundaryEventsForActivity finds boundary events attached to activity
func (hce *HttpConnectorExecutor) findBoundaryEventsForActivity(
	activityID string,
	bpmnProcess map[string]interface{},
) map[string]map[string]interface{} {
	boundaryEvents := make(map[string]map[string]interface{})

	elements, exists := bpmnProcess["elements"]
	if !exists {
		return boundaryEvents
	}

	elementsMap, ok := elements.(map[string]interface{})
	if !ok {
		return boundaryEvents
	}

	for elementID, element := range elementsMap {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			continue
		}

		elementType, exists := elementMap["type"]
		if !exists || elementType != "boundaryEvent" {
			continue
		}

		attachedToRef, exists := elementMap["attached_to_ref"]
		if exists && attachedToRef == activityID {
			boundaryEvents[elementID] = elementMap
		}
	}

	return boundaryEvents
}

// createBoundaryTimerForEvent creates timer for boundary event if it has timer definition
func (hce *HttpConnectorExecutor) createBoundaryTimerForEvent(
	token *models.Token,
	eventID string,
	boundaryEvent map[string]interface{},
) error {
	eventDefinitions, exists := boundaryEvent["event_definitions"]
	if !exists {
		return nil
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil
	}

	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "timerEventDefinition" {
			continue
		}

		timerData, exists := eventDefMap["timer"]
		if !exists {
			continue
		}

		timerMap, ok := timerData.(map[string]interface{})
		if !ok {
			continue
		}

		timerRequest := &TimerRequest{
			ElementID:         eventID,
			TokenID:           token.TokenID,
			ProcessInstanceID: token.ProcessInstanceID,
			ProcessKey:        token.ProcessKey,
		}

		if attachedToRef, exists := boundaryEvent["attached_to_ref"]; exists {
			if attachedStr, ok := attachedToRef.(string); ok {
				timerRequest.AttachedToRef = &attachedStr
			}
		}

		if cancelActivity, exists := boundaryEvent["cancel_activity"]; exists {
			if cancelBool, ok := cancelActivity.(bool); ok {
				timerRequest.CancelActivity = &cancelBool
			}
		}

		if duration, exists := timerMap["duration"]; exists {
			if durationStr, ok := duration.(string); ok {
				evaluatedDuration, err := hce.evaluateTimerExpression(durationStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer duration expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", durationStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer duration: %w", err)
				}
				evaluatedDurationStr := fmt.Sprintf("%v", evaluatedDuration)
				timerRequest.TimeDuration = &evaluatedDurationStr
				logger.Debug("Boundary timer duration evaluated",
					logger.String("original", durationStr),
					logger.String("evaluated", evaluatedDurationStr))
			}
		} else if cycle, exists := timerMap["cycle"]; exists {
			if cycleStr, ok := cycle.(string); ok {
				evaluatedCycle, err := hce.evaluateTimerExpression(cycleStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer cycle expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", cycleStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer cycle: %w", err)
				}
				evaluatedCycleStr := fmt.Sprintf("%v", evaluatedCycle)
				timerRequest.TimeCycle = &evaluatedCycleStr
				logger.Debug("Boundary timer cycle evaluated",
					logger.String("original", cycleStr),
					logger.String("evaluated", evaluatedCycleStr))
			}
		} else if date, exists := timerMap["date"]; exists {
			if dateStr, ok := date.(string); ok {
				evaluatedDate, err := hce.evaluateTimerExpression(dateStr, token)
				if err != nil {
					logger.Error("Failed to evaluate boundary timer date expression",
						logger.String("token_id", token.TokenID),
						logger.String("expression", dateStr),
						logger.String("error", err.Error()))
					return fmt.Errorf("failed to evaluate boundary timer date: %w", err)
				}
				evaluatedDateStr := fmt.Sprintf("%v", evaluatedDate)
				timerRequest.TimeDate = &evaluatedDateStr
				logger.Debug("Boundary timer date evaluated",
					logger.String("original", dateStr),
					logger.String("evaluated", evaluatedDateStr))
			}
		}

		timerID, err := hce.processComponent.CreateBoundaryTimerWithID(timerRequest)
		if err != nil {
			return fmt.Errorf("failed to create boundary timer: %w", err)
		}

		logger.Info("Boundary timer created",
			logger.String("parent_token_id", token.TokenID),
			logger.String("timer_id", timerID),
			logger.String("event_id", eventID),
			logger.String("activity_id", token.CurrentElementID))

		if err := hce.processComponent.LinkBoundaryTimerToToken(token.TokenID, timerID); err != nil {
			logger.Error("Failed to link boundary timer to token",
				logger.String("parent_token_id", token.TokenID),
				logger.String("timer_id", timerID),
				logger.String("error", err.Error()))
		}
	}

	return nil
}

// createErrorBoundaries creates error boundary subscriptions for activity
func (hce *HttpConnectorExecutor) createErrorBoundaries(token *models.Token, element map[string]interface{}) error {
	if hce.processComponent == nil {
		return nil
	}

	bpmnProcess, err := hce.processComponent.GetBPMNProcessForToken(token)
	if err != nil {
		return fmt.Errorf("failed to get BPMN process: %w", err)
	}

	boundaryEvents := hce.findBoundaryEventsForActivity(token.CurrentElementID, bpmnProcess)
	if len(boundaryEvents) == 0 {
		return nil
	}

	logger.Info("Found boundary events for error boundary registration",
		logger.String("activity_id", token.CurrentElementID),
		logger.Int("boundary_events_count", len(boundaryEvents)))

	for eventID, boundaryEvent := range boundaryEvents {
		if err := hce.createErrorBoundaryForEvent(token, eventID, boundaryEvent, bpmnProcess); err != nil {
			logger.Error("Failed to create error boundary subscription",
				logger.String("token_id", token.TokenID),
				logger.String("event_id", eventID),
				logger.String("error", err.Error()))
			continue
		}
	}

	return nil
}

// createErrorBoundaryForEvent creates error boundary subscription for specific event
func (hce *HttpConnectorExecutor) createErrorBoundaryForEvent(
	token *models.Token,
	eventID string,
	boundaryEvent interface{},
	bpmnProcess interface{},
) error {
	boundaryEventMap, ok := boundaryEvent.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid boundary event structure")
	}

	eventDefinitions, exists := boundaryEventMap["event_definitions"]
	if !exists {
		return nil
	}

	eventDefList, ok := eventDefinitions.([]interface{})
	if !ok {
		return nil
	}

	for _, eventDef := range eventDefList {
		eventDefMap, ok := eventDef.(map[string]interface{})
		if !ok {
			continue
		}

		eventType, exists := eventDefMap["type"]
		if !exists || eventType != "errorEventDefinition" {
			continue
		}

		errorCode, errorName := hce.extractErrorInfo(eventDefMap, bpmnProcess)

		cancelActivity := true
		if cancelActivityAttr, exists := boundaryEventMap["cancel_activity"]; exists {
			if cancelActivityBool, ok := cancelActivityAttr.(bool); ok {
				cancelActivity = cancelActivityBool
			} else if cancelActivityStr, ok := cancelActivityAttr.(string); ok {
				cancelActivity = cancelActivityStr != "false"
			}
		}

		outgoingFlows := hce.getOutgoingFlows(boundaryEventMap)

		subscription := &ErrorBoundarySubscription{
			TokenID:        token.TokenID,
			ElementID:      eventID,
			AttachedToRef:  token.CurrentElementID,
			ErrorCode:      errorCode,
			ErrorName:      errorName,
			CancelActivity: cancelActivity,
			OutgoingFlows:  outgoingFlows,
		}

		hce.processComponent.RegisterErrorBoundary(subscription)

		logger.Info("Error boundary subscription created",
			logger.String("token_id", token.TokenID),
			logger.String("event_id", eventID),
			logger.String("error_code", errorCode),
			logger.Bool("cancel_activity", cancelActivity))

		return nil
	}

	return nil
}

// extractErrorInfo extracts error code and name from error event definition
func (hce *HttpConnectorExecutor) extractErrorInfo(
	eventDef map[string]interface{},
	bpmnProcess interface{},
) (string, string) {
	errorRef, exists := eventDef["reference"]
	if !exists {
		return "GENERAL_ERROR", "General Error"
	}

	errorRefStr, ok := errorRef.(string)
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	bpmnProcessMap, ok := bpmnProcess.(map[string]interface{})
	if !ok {
		return "GENERAL_ERROR", "General Error"
	}

	if elements, exists := bpmnProcessMap["elements"]; exists {
		if elementsMap, ok := elements.(map[string]interface{}); ok {
			if errorElement, exists := elementsMap[errorRefStr]; exists {
				if errorDefMap, ok := errorElement.(map[string]interface{}); ok {
					errorCode := "GENERAL_ERROR"
					errorName := "General Error"

					if code, exists := errorDefMap["error_code"]; exists {
						if codeStr, ok := code.(string); ok {
							errorCode = codeStr
						}
					}

					if name, exists := errorDefMap["name"]; exists {
						if nameStr, ok := name.(string); ok {
							errorName = nameStr
						}
					}

					logger.Info("Resolved error definition from elements",
						logger.String("error_ref", errorRefStr),
						logger.String("error_code", errorCode),
						logger.String("error_name", errorName))

					return errorCode, errorName
				}
			}
		}
	}

	logger.Warn("Could not resolve error definition, using default",
		logger.String("error_ref", errorRefStr))
	return "GENERAL_ERROR", "General Error"
}

// getOutgoingFlows extracts outgoing sequence flows from boundary event
func (hce *HttpConnectorExecutor) getOutgoingFlows(boundaryEvent map[string]interface{}) []string {
	outgoing, exists := boundaryEvent["outgoing"]
	if !exists {
		return []string{}
	}

	var flows []string
	if outgoingList, ok := outgoing.([]interface{}); ok {
		for _, item := range outgoingList {
			if flowID, ok := item.(string); ok {
				flows = append(flows, flowID)
			}
		}
	} else if outgoingStr, ok := outgoing.(string); ok {
		flows = append(flows, outgoingStr)
	}

	return flows
}

// evaluateTimerExpression evaluates timer expressions using expression component
func (hce *HttpConnectorExecutor) evaluateTimerExpression(expression string, token *models.Token) (interface{}, error) {
	if expression == "" || len(expression) == 0 || expression[0] != '=' {
		return expression, nil
	}

	if hce.processComponent == nil {
		return nil, fmt.Errorf("process component not available for expression evaluation")
	}

	core := hce.processComponent.GetCore()
	if core == nil {
		return nil, fmt.Errorf("core interface not available for expression evaluation")
	}

	expressionCompInterface := core.GetExpressionComponent()
	if expressionCompInterface == nil {
		return nil, fmt.Errorf("expression component not available")
	}

	type ExpressionEvaluator interface {
		EvaluateExpressionEngine(expression interface{}, variables map[string]interface{}) (interface{}, error)
	}

	expressionComp, ok := expressionCompInterface.(ExpressionEvaluator)
	if !ok {
		return nil, fmt.Errorf("failed to cast expression component to ExpressionEvaluator interface")
	}

	result, err := expressionComp.EvaluateExpressionEngine(expression, token.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate FEEL expression '%s': %w", expression, err)
	}

	logger.Debug("Boundary timer expression evaluated successfully",
		logger.String("token_id", token.TokenID),
		logger.String("original_expression", expression),
		logger.Any("evaluated_result", result),
		logger.Any("token_variables", token.Variables))

	return result, nil
}
