/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package process

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// EmailConnectorExecutor executes email connector tasks
type EmailConnectorExecutor struct {
	processComponent ComponentInterface
}

// NewEmailConnectorExecutor creates new email connector executor
func NewEmailConnectorExecutor(processComponent ComponentInterface) *EmailConnectorExecutor {
	return &EmailConnectorExecutor{
		processComponent: processComponent,
	}
}

// GetElementType returns element type
func (ece *EmailConnectorExecutor) GetElementType() string {
	return "serviceTask"
}

// EmailConnectorConfig contains email connector configuration
type EmailConnectorConfig struct {
	Authentication EmailAuthConfig
	Protocol       string
	SMTPConfig     SMTPConfig
	Action         EmailMessage
}

// EmailAuthConfig contains authentication configuration
type EmailAuthConfig struct {
	Type     string
	Username string
	Password string
}

// SMTPConfig contains SMTP server configuration
type SMTPConfig struct {
	Host                   string
	Port                   int
	CryptographicProtocol  string
}

// EmailMessage contains email message data
type EmailMessage struct {
	From        string
	To          string
	CC          string
	BCC         string
	Subject     string
	ContentType string
	Body        string
	Headers     map[string]interface{}
	Attachments []EmailAttachment
}

// EmailAttachment represents email attachment
type EmailAttachment struct {
	Filename    string
	Content     string
	ContentType string
}

// EmailResponse contains email send result
type EmailResponse struct {
	Status    string
	MessageID string
	Timestamp string
	Error     string
}

// Execute executes email connector request
func (ece *EmailConnectorExecutor) Execute(
	token *models.Token,
	element map[string]interface{},
) (*ExecutionResult, error) {
	logger.Info("Executing email connector",
		logger.String("token_id", token.TokenID),
		logger.String("element_id", token.CurrentElementID))

	config, err := ece.extractEmailConnectorConfig(element, token.Variables)
	if err != nil {
		logger.Error("Failed to extract email connector configuration",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("Email connector configuration error: %v", err),
			Completed: false,
		}, nil
	}

	logger.Info("Email connector configuration extracted",
		logger.String("token_id", token.TokenID),
		logger.String("to", config.Action.To),
		logger.String("subject", config.Action.Subject))

	response, err := ece.sendEmail(config)
	if err != nil {
		logger.Error("Email send failed",
			logger.String("token_id", token.TokenID),
			logger.String("to", config.Action.To),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("Email send failed: %v", err),
			Completed: false,
		}, nil
	}

	logger.Info("Email sent successfully",
		logger.String("token_id", token.TokenID),
		logger.String("to", config.Action.To),
		logger.String("message_id", response.MessageID))

	err = ece.updateTokenWithEmailResponse(token, response)
	if err != nil {
		logger.Error("Failed to update token with email response",
			logger.String("token_id", token.TokenID),
			logger.String("error", err.Error()))
		return &ExecutionResult{
			Success:   false,
			Error:     fmt.Sprintf("Failed to process email response: %v", err),
			Completed: false,
		}, nil
	}

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

// extractEmailConnectorConfig extracts email connector configuration from ioMapping
func (ece *EmailConnectorExecutor) extractEmailConnectorConfig(
	element map[string]interface{},
	tokenVariables map[string]interface{},
) (*EmailConnectorConfig, error) {
	config := &EmailConnectorConfig{
		Protocol: "smtp",
	}

	extensionElements, exists := element["extension_elements"]
	if !exists {
		return nil, fmt.Errorf("no extension elements found")
	}

	extElementsList, ok := extensionElements.([]interface{})
	if !ok {
		return nil, fmt.Errorf("extension elements is not an array")
	}

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
			if !exists {
				continue
			}

			if extType == "ioMapping" {
				ioMapping, exists := extMap["io_mapping"]
				if !exists {
					continue
				}

				ioMappingMap, ok := ioMapping.(map[string]interface{})
				if !ok {
					continue
				}

				inputs, exists := ioMappingMap["inputs"]
				if !exists {
					continue
				}

				inputsList, ok := inputs.([]interface{})
				if !ok {
					continue
				}

				for _, input := range inputsList {
					inputMap, ok := input.(map[string]interface{})
					if !ok {
						continue
					}

					source, _ := inputMap["source"].(string)
					target, _ := inputMap["target"].(string)

					value := ece.resolveInputValue(source, tokenVariables)

					ece.setConfigValue(config, target, value)
				}
			}
		}
	}

	if config.SMTPConfig.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if config.Action.To == "" {
		return nil, fmt.Errorf("recipient (to) is required")
	}
	if config.Action.From == "" {
		return nil, fmt.Errorf("sender (from) is required")
	}

	return config, nil
}

// resolveInputValue resolves input value from source
func (ece *EmailConnectorExecutor) resolveInputValue(source string, variables map[string]interface{}) interface{} {
	if source == "" {
		return ""
	}

	if strings.HasPrefix(source, "=") {
		expr := strings.TrimPrefix(source, "=")
		return expr
	}

	if val, exists := variables[source]; exists {
		return val
	}

	return source
}

// setConfigValue sets configuration value by target path
func (ece *EmailConnectorExecutor) setConfigValue(config *EmailConnectorConfig, target string, value interface{}) {
	valueStr := fmt.Sprintf("%v", value)

	parts := strings.Split(target, ".")
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "authentication":
		if len(parts) >= 2 {
			switch parts[1] {
			case "type":
				config.Authentication.Type = valueStr
			case "username":
				config.Authentication.Username = valueStr
			case "password":
				config.Authentication.Password = valueStr
			}
		}
	case "protocol":
		config.Protocol = valueStr
	case "data":
		if len(parts) >= 2 {
			switch parts[1] {
			case "smtpConfig":
				if len(parts) >= 3 {
					switch parts[2] {
					case "smtpHost":
						config.SMTPConfig.Host = valueStr
					case "smtpPort":
						if port, ok := value.(int); ok {
							config.SMTPConfig.Port = port
						} else if portStr, ok := value.(string); ok {
							fmt.Sscanf(portStr, "%d", &config.SMTPConfig.Port)
						}
					case "smtpCryptographicProtocol":
						config.SMTPConfig.CryptographicProtocol = valueStr
					}
				}
			case "smtpAction":
				if len(parts) >= 3 {
					switch parts[2] {
					case "from":
						config.Action.From = valueStr
					case "to":
						config.Action.To = valueStr
					case "cc":
						config.Action.CC = valueStr
					case "bcc":
						config.Action.BCC = valueStr
					case "subject":
						config.Action.Subject = valueStr
					case "contentType":
						config.Action.ContentType = valueStr
					case "body":
						config.Action.Body = valueStr
					case "headers":
						if headersMap, ok := value.(map[string]interface{}); ok {
							config.Action.Headers = headersMap
						}
					}
				}
			}
		}
	}
}

// sendEmail sends email via SMTP
func (ece *EmailConnectorExecutor) sendEmail(config *EmailConnectorConfig) (*EmailResponse, error) {
	client, err := ece.connectSMTP(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Close()

	if config.Authentication.Type == "simple" {
		auth := ece.createAuth(config)
		if err := client.Auth(auth); err != nil {
			return nil, fmt.Errorf("SMTP authentication failed: %v", err)
		}
	}

	if err := client.Mail(config.Action.From); err != nil {
		return nil, fmt.Errorf("failed to set sender: %v", err)
	}

	recipients := ece.parseRecipients(config.Action.To, config.Action.CC, config.Action.BCC)
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return nil, fmt.Errorf("failed to add recipient %s: %v", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return nil, fmt.Errorf("failed to start data transmission: %v", err)
	}

	message := ece.buildMessage(config)
	if _, err := writer.Write([]byte(message)); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to write message: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close data transmission: %v", err)
	}

	messageID := fmt.Sprintf("<%s@%s>", generateMessageID(), config.SMTPConfig.Host)

	return &EmailResponse{
		Status:    "success",
		MessageID: messageID,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// createAuth creates SMTP auth based on configuration
func (ece *EmailConnectorExecutor) createAuth(config *EmailConnectorConfig) smtp.Auth {
	if config.SMTPConfig.CryptographicProtocol == "NONE" || config.SMTPConfig.CryptographicProtocol == "" {
		return &insecurePlainAuth{
			username: config.Authentication.Username,
			password: config.Authentication.Password,
			host:     config.SMTPConfig.Host,
		}
	}
	
	return smtp.PlainAuth("", config.Authentication.Username, config.Authentication.Password, config.SMTPConfig.Host)
}

// insecurePlainAuth implements SMTP AUTH PLAIN without TLS requirement
type insecurePlainAuth struct {
	username string
	password string
	host     string
}

func (a *insecurePlainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *insecurePlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}

// connectSMTP establishes SMTP connection
func (ece *EmailConnectorExecutor) connectSMTP(config *EmailConnectorConfig) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", config.SMTPConfig.Host, config.SMTPConfig.Port)

	switch config.SMTPConfig.CryptographicProtocol {
	case "TLS", "SSL":
		tlsConfig := &tls.Config{
			ServerName: config.SMTPConfig.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		client, err := smtp.NewClient(conn, config.SMTPConfig.Host)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return client, nil

	case "STARTTLS":
		client, err := smtp.Dial(addr)
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			ServerName: config.SMTPConfig.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, err
		}
		return client, nil

	case "NONE", "":
		return smtp.Dial(addr)

	default:
		return nil, fmt.Errorf("unsupported cryptographic protocol: %s", config.SMTPConfig.CryptographicProtocol)
	}
}

// buildMessage builds MIME email message
func (ece *EmailConnectorExecutor) buildMessage(config *EmailConnectorConfig) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("From: %s\r\n", config.Action.From))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", config.Action.To))

	if config.Action.CC != "" {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", config.Action.CC))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", config.Action.Subject)))
	builder.WriteString("MIME-Version: 1.0\r\n")

	if len(config.Action.Attachments) > 0 {
		boundary := generateBoundary()
		builder.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		builder.WriteString("\r\n")

		builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		ece.writeMessageBody(&builder, config)

		for _, attachment := range config.Action.Attachments {
			builder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			ece.writeAttachment(&builder, attachment)
		}

		builder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		ece.writeMessageBody(&builder, config)
	}

	return builder.String()
}

// writeMessageBody writes message body to builder
func (ece *EmailConnectorExecutor) writeMessageBody(builder *strings.Builder, config *EmailConnectorConfig) {
	contentType := config.Action.ContentType
	if contentType == "" || contentType == "PLAIN" {
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(config.Action.Body)
		builder.WriteString("\r\n")
	} else if contentType == "HTML" {
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(config.Action.Body)
		builder.WriteString("\r\n")
	}
}

// writeAttachment writes attachment to builder
func (ece *EmailConnectorExecutor) writeAttachment(builder *strings.Builder, attachment EmailAttachment) {
	contentType := attachment.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	builder.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	builder.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Filename))
	builder.WriteString("Content-Transfer-Encoding: base64\r\n")
	builder.WriteString("\r\n")

	encoded := base64.StdEncoding.EncodeToString([]byte(attachment.Content))
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		builder.WriteString(encoded[i:end])
		builder.WriteString("\r\n")
	}
}

// parseRecipients parses and combines all recipients
func (ece *EmailConnectorExecutor) parseRecipients(to, cc, bcc string) []string {
	recipients := make([]string, 0)

	recipients = append(recipients, ece.splitEmails(to)...)
	recipients = append(recipients, ece.splitEmails(cc)...)
	recipients = append(recipients, ece.splitEmails(bcc)...)

	return recipients
}

// splitEmails splits email addresses by comma or semicolon
func (ece *EmailConnectorExecutor) splitEmails(emails string) []string {
	if emails == "" {
		return []string{}
	}

	emails = strings.ReplaceAll(emails, ";", ",")
	parts := strings.Split(emails, ",")

	result := make([]string, 0, len(parts))
	for _, email := range parts {
		email = strings.TrimSpace(email)
		if email != "" {
			result = append(result, email)
		}
	}

	return result
}

// updateTokenWithEmailResponse updates token variables with email response
func (ece *EmailConnectorExecutor) updateTokenWithEmailResponse(token *models.Token, response *EmailResponse) error {
	if token.Variables == nil {
		token.Variables = make(map[string]interface{})
	}

	responseData := map[string]interface{}{
		"status":    response.Status,
		"messageId": response.MessageID,
		"timestamp": response.Timestamp,
	}

	if response.Error != "" {
		responseData["error"] = response.Error
	}

	token.Variables["response"] = responseData

	return nil
}

// generateMessageID generates unique message ID
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// generateBoundary generates MIME boundary
func generateBoundary() string {
	return fmt.Sprintf("boundary_%d", time.Now().UnixNano())
}

// guessContentType guesses content type from filename
func guessContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

