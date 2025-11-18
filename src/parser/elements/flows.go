/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package elements

import (
	"atom-engine/src/parser"
)

// FlowParser parses all BPMN flow elements
// Парсер всех элементов потоков BPMN
type FlowParser struct{}

// NewFlowParser creates new flow parser
// Создает новый парсер потоков
func NewFlowParser() *FlowParser {
	return &FlowParser{}
}

// Parse parses any BPMN flow element
// Парсит любой элемент потока BPMN
func (p *FlowParser) Parse(element *parser.XMLElement, context *parser.ParseContext) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	flowType := element.XMLName.Local
	result["type"] = flowType
	result["namespace"] = element.XMLName.Space

	// Parse all attributes
	// Парсинг всех атрибутов
	attributes := make(map[string]string)
	for _, attr := range element.Attributes {
		attributes[attr.Name.Local] = attr.Value

		// Store key attributes in result
		// Сохранение ключевых атрибутов в результате
		switch attr.Name.Local {
		case "id":
			result["id"] = attr.Value
		case "name":
			result["name"] = attr.Value
		case "sourceRef":
			result["source_ref"] = attr.Value
		case "targetRef":
			result["target_ref"] = attr.Value
		case "isImmediate":
			result["is_immediate"] = attr.Value
		case "associationDirection":
			result["association_direction"] = attr.Value
		}
	}

	// Store all attributes
	// Сохранение всех атрибутов
	result["attributes"] = attributes

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

	// Parse flow-specific elements
	// Парсинг специфичных для потока элементов
	switch flowType {
	case "sequenceFlow":
		seqFlow := p.parseSequenceFlow(element)
		if len(seqFlow) > 0 {
			result["sequence_flow"] = seqFlow
		}
	case "messageFlow":
		msgFlow := p.parseMessageFlow(element)
		if len(msgFlow) > 0 {
			result["message_flow"] = msgFlow
		}
	case "association":
		association := p.parseAssociation(element)
		if len(association) > 0 {
			result["association"] = association
		}
	}

	return result, nil
}

// parseSequenceFlow parses sequence flow specific elements
// Парсинг специфичных элементов последовательного потока
func (p *FlowParser) parseSequenceFlow(element *parser.XMLElement) map[string]interface{} {
	seqFlow := make(map[string]interface{})

	// Parse condition expression if present
	// Парсинг выражения условия если есть
	for _, child := range element.Children {
		if child.XMLName.Local == "conditionExpression" {
			condition := make(map[string]interface{})
			condition["expression"] = child.Text

			// Parse condition attributes
			// Парсинг атрибутов условия
			conditionAttrs := make(map[string]string)
			for _, attr := range child.Attributes {
				conditionAttrs[attr.Name.Local] = attr.Value

				if attr.Name.Local == "type" {
					condition["expression_type"] = attr.Value
				}
			}
			condition["attributes"] = conditionAttrs

			seqFlow["condition"] = condition
		}
	}

	// Check if this is a conditional flow
	// Проверка является ли это условным потоком
	hasCondition := false
	for _, child := range element.Children {
		if child.XMLName.Local == "conditionExpression" {
			hasCondition = true
			break
		}
	}
	seqFlow["has_condition"] = hasCondition

	return seqFlow
}

// parseMessageFlow parses message flow specific elements
// Парсинг специфичных элементов потока сообщений
func (p *FlowParser) parseMessageFlow(element *parser.XMLElement) map[string]interface{} {
	msgFlow := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "messageRef":
			msgFlow["message_ref"] = attr.Value
		}
	}

	return msgFlow
}

// parseAssociation parses association specific elements
// Парсинг специфичных элементов ассоциации
func (p *FlowParser) parseAssociation(element *parser.XMLElement) map[string]interface{} {
	association := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "associationDirection":
			association["direction"] = attr.Value
		}
	}

	return association
}

// parseExtensionElements parses extension elements for flows
// Парсинг элементов расширения для потоков
func (p *FlowParser) parseExtensionElements(element *parser.XMLElement) map[string]interface{} {
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

// GetElementType returns element type (generic for all flows)
// Возвращает тип элемента (общий для всех потоков)
func (p *FlowParser) GetElementType() string {
	return "flow"
}
