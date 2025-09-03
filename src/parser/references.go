/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"atom-engine/src/core/logger"
)

// ReferenceParser parses all BPMN reference elements (error, signal, message)
// Парсер всех ссылочных элементов BPMN (error, signal, message)
type ReferenceParser struct{}

// NewReferenceParser creates new reference parser
// Создает новый парсер ссылок
func NewReferenceParser() *ReferenceParser {
	return &ReferenceParser{}
}

// Parse parses any BPMN reference element
// Парсит любой ссылочный элемент BPMN
func (p *ReferenceParser) Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error) {
	logger.Debug("Parsing reference element",
		logger.String("element_type", element.XMLName.Local),
		logger.String("process_id", context.ProcessID))

	result := make(map[string]interface{})

	// Set element type and namespace
	// Установка типа элемента и пространства имен
	refType := element.XMLName.Local
	result["type"] = refType
	result["namespace"] = element.XMLName.Space
	result["parsed_with"] = "reference_parser"

	// Parse all attributes
	// Парсинг всех атрибутов
	attributes := make(map[string]string)
	for _, attr := range element.Attributes {
		attributes[attr.Name.Local] = attr.Value

		// Store key attributes directly for easy access
		// Сохранение ключевых атрибутов напрямую для удобного доступа
		switch attr.Name.Local {
		case "id":
			result["id"] = attr.Value
		case "name":
			result["name"] = attr.Value
		case "errorCode":
			result["error_code"] = attr.Value
		case "escalationCode":
			result["escalation_code"] = attr.Value
		}
	}
	result["attributes"] = attributes

	// Parse specific reference content based on type
	// Парсинг специфичного содержимого ссылки по типу
	switch refType {
	case "error":
		errorData := p.parseError(element)
		result["error_data"] = errorData
		logger.Debug("Parsed error element",
			logger.String("element_id", getStringValueSafe(result["id"])),
			logger.String("error_code", getStringValueSafe(errorData["error_code"])),
			logger.String("error_name", getStringValueSafe(errorData["name"])))

	case "signal":
		signalData := p.parseSignal(element)
		result["signal_data"] = signalData
		logger.Debug("Parsed signal element",
			logger.String("element_id", getStringValueSafe(result["id"])),
			logger.String("signal_name", getStringValueSafe(signalData["name"])))

	case "message":
		messageData := p.parseMessage(element)
		result["message_data"] = messageData
		logger.Debug("Parsed message element",
			logger.String("element_id", getStringValueSafe(result["id"])),
			logger.String("message_name", getStringValueSafe(messageData["name"])))

	case "escalation":
		escalationData := p.parseEscalation(element)
		result["escalation_data"] = escalationData
		logger.Debug("Parsed escalation element",
			logger.String("element_id", getStringValueSafe(result["id"])),
			logger.String("escalation_code", getStringValueSafe(escalationData["escalation_code"])),
			logger.String("escalation_name", getStringValueSafe(escalationData["name"])))

	default:
		logger.Warn("Unknown reference element type",
			logger.String("element_type", refType),
			logger.String("element_id", getStringValueSafe(result["id"])))
	}

	// Add text content if present
	// Добавление текстового содержимого если есть
	if element.Text != "" {
		result["text"] = element.Text
	}

	// Parse documentation if present
	// Парсинг документации если есть
	documentations := make([]string, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "documentation" && child.Text != "" {
			documentations = append(documentations, child.Text)
		}
	}
	if len(documentations) > 0 {
		result["documentation"] = documentations
	}

	// Parse extension elements if present
	// Парсинг элементов расширения если есть
	extensionElements := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "extensionElements" {
			extData := p.parseExtensionElements(child)
			extensionElements = append(extensionElements, extData)
		}
	}
	if len(extensionElements) > 0 {
		result["extension_elements"] = extensionElements
	}

	logger.Info("Successfully parsed reference element",
		logger.String("element_type", refType),
		logger.String("element_id", getStringValueSafe(result["id"])),
		logger.String("process_id", context.ProcessID))

	return result, nil
}

// parseError parses bpmn:error element
// Парсинг элемента bpmn:error
func (p *ReferenceParser) parseError(element *XMLElement) map[string]interface{} {
	errorData := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			errorData["id"] = attr.Value
		case "name":
			errorData["name"] = attr.Value
		case "errorCode":
			errorData["error_code"] = attr.Value
			logger.Debug("Found error code",
				logger.String("error_code", attr.Value))
		case "errorMessage":
			errorData["error_message"] = attr.Value
		}
	}

	return errorData
}

// parseSignal parses bpmn:signal element
// Парсинг элемента bpmn:signal
func (p *ReferenceParser) parseSignal(element *XMLElement) map[string]interface{} {
	signalData := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			signalData["id"] = attr.Value
		case "name":
			signalData["name"] = attr.Value
			logger.Debug("Found signal name",
				logger.String("signal_name", attr.Value))
		}
	}

	return signalData
}

// parseMessage parses bpmn:message element
// Парсинг элемента bpmn:message
func (p *ReferenceParser) parseMessage(element *XMLElement) map[string]interface{} {
	messageData := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			messageData["id"] = attr.Value
		case "name":
			messageData["name"] = attr.Value
			logger.Debug("Found message name",
				logger.String("message_name", attr.Value))
		}
	}

	// Parse correlationProperty elements if present
	// Парсинг элементов correlationProperty если есть
	correlationProperties := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "correlationProperty" {
			corrProp := p.parseCorrelationProperty(child)
			correlationProperties = append(correlationProperties, corrProp)
		}
	}

	if len(correlationProperties) > 0 {
		messageData["correlation_properties"] = correlationProperties
		logger.Debug("Found correlation properties",
			logger.Int("count", len(correlationProperties)))
	}

	return messageData
}

// parseEscalation parses bpmn:escalation element
// Парсинг элемента bpmn:escalation
func (p *ReferenceParser) parseEscalation(element *XMLElement) map[string]interface{} {
	escalationData := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			escalationData["id"] = attr.Value
		case "name":
			escalationData["name"] = attr.Value
		case "escalationCode":
			escalationData["escalation_code"] = attr.Value
			logger.Debug("Found escalation code",
				logger.String("escalation_code", attr.Value))
		}
	}

	return escalationData
}

// parseCorrelationProperty parses correlationProperty element
// Парсинг элемента correlationProperty
func (p *ReferenceParser) parseCorrelationProperty(element *XMLElement) map[string]interface{} {
	corrProp := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			corrProp["id"] = attr.Value
		case "name":
			corrProp["name"] = attr.Value
		case "type":
			corrProp["type"] = attr.Value
		}
	}

	// Parse correlationPropertyRetrievalExpression elements if present
	// Парсинг элементов correlationPropertyRetrievalExpression если есть
	retrievalExpressions := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "correlationPropertyRetrievalExpression" {
			retrieval := p.parseCorrelationPropertyRetrievalExpression(child)
			retrievalExpressions = append(retrievalExpressions, retrieval)
		}
	}

	if len(retrievalExpressions) > 0 {
		corrProp["retrieval_expressions"] = retrievalExpressions
	}

	return corrProp
}

// parseCorrelationPropertyRetrievalExpression parses correlationPropertyRetrievalExpression element
// Парсинг элемента correlationPropertyRetrievalExpression
func (p *ReferenceParser) parseCorrelationPropertyRetrievalExpression(element *XMLElement) map[string]interface{} {
	retrieval := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "messageRef":
			retrieval["message_ref"] = attr.Value
		}
	}

	// Parse formalExpression if present
	// Парсинг formalExpression если есть
	for _, child := range element.Children {
		if child.XMLName.Local == "formalExpression" {
			retrieval["expression"] = child.Text
			for _, attr := range child.Attributes {
				if attr.Name.Local == "language" {
					retrieval["expression_language"] = attr.Value
				}
			}
		}
	}

	return retrieval
}

// parseExtensionElements parses extension elements for references
// Парсинг элементов расширения для ссылок
func (p *ReferenceParser) parseExtensionElements(element *XMLElement) map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = "extensionElements"

	extensions := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		extElement := make(map[string]interface{})
		extElement["type"] = child.XMLName.Local
		extElement["namespace"] = child.XMLName.Space

		// Parse attributes
		// Парсинг атрибутов
		attributes := make(map[string]string)
		for _, attr := range child.Attributes {
			attributes[attr.Name.Local] = attr.Value
		}
		extElement["attributes"] = attributes

		// Parse text content
		// Парсинг текстового содержимого
		if child.Text != "" {
			extElement["text"] = child.Text
		}

		extensions = append(extensions, extElement)
	}

	result["extensions"] = extensions
	return result
}

// GetElementType returns element type for references
// Возвращает тип элемента для ссылок
func (p *ReferenceParser) GetElementType() string {
	return "reference"
}

// getStringValueSafe safely gets string value from interface{}
// Безопасно получает строковое значение из interface{}
func getStringValueSafe(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
