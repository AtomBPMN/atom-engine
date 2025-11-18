/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package elements

import (
	"strconv"

	"atom-engine/src/parser"
)

// ProcessParser parses bpmn:process element
// Парсер элемента bpmn:process
type ProcessParser struct{}

// NewProcessParser creates new process parser
// Создает новый парсер process
func NewProcessParser() *ProcessParser {
	return &ProcessParser{}
}

// Parse parses bpmn:process element
// Парсит элемент bpmn:process
func (p *ProcessParser) Parse(
	element *parser.XMLElement,
	context *parser.ParseContext,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	result["type"] = "process"
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
		case "isExecutable":
			if executable, err := strconv.ParseBool(attr.Value); err == nil {
				result["is_executable"] = executable
			} else {
				result["is_executable"] = attr.Value
			}
		case "processType":
			result["process_type"] = attr.Value
		case "isClosed":
			if closed, err := strconv.ParseBool(attr.Value); err == nil {
				result["is_closed"] = closed
			} else {
				result["is_closed"] = attr.Value
			}
		}
	}

	// Store all attributes
	// Сохранение всех атрибутов
	result["attributes"] = attributes

	// Parse documentation if present
	// Парсинг документации если есть
	documentations := make([]string, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "documentation" {
			if child.Text != "" {
				documentations = append(documentations, child.Text)
			}
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

	// Count child flow elements
	// Подсчет дочерних элементов потока
	flowElements := make(map[string]int)
	for _, child := range element.Children {
		elementType := child.XMLName.Local
		switch elementType {
		case "startEvent", "endEvent", "intermediateCatchEvent", "intermediateThrowEvent", "boundaryEvent":
			flowElements["events"]++
		case "task", "userTask", "serviceTask", "scriptTask", "sendTask",
			"receiveTask", "manualTask", "businessRuleTask", "callActivity", "subProcess":
			flowElements["activities"]++
		case "exclusiveGateway", "parallelGateway", "inclusiveGateway", "complexGateway", "eventBasedGateway":
			flowElements["gateways"]++
		case "sequenceFlow":
			flowElements["flows"]++
		case "dataObject", "dataStore", "dataStoreReference":
			flowElements["data"]++
		default:
			flowElements["other"]++
		}

		// Also count individual element types
		// Также подсчитываем отдельные типы элементов
		flowElements[elementType]++
	}
	result["flow_elements_count"] = flowElements

	// Extract flow node IDs for later reference
	// Извлечение ID узлов потока для последующих ссылок
	flowNodeIDs := make([]string, 0)
	for _, child := range element.Children {
		for _, attr := range child.Attributes {
			if attr.Name.Local == "id" {
				flowNodeIDs = append(flowNodeIDs, attr.Value)
				break
			}
		}
	}
	result["flow_node_ids"] = flowNodeIDs

	return result, nil
}

// parseExtensionElements parses extension elements specific to process
// Парсинг элементов расширения специфичных для процесса
func (p *ProcessParser) parseExtensionElements(element *parser.XMLElement) map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = "extensionElements"

	// Parse process-specific extension elements
	// Парсинг специфичных для процесса элементов расширения
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

		// Handle special Zeebe elements
		// Обработка специальных Zeebe элементов
		switch child.XMLName.Local {
		case "properties":
			properties := p.parseZeebeProperties(child)
			extElement["properties"] = properties
		case "taskDefinition":
			taskDef := p.parseZeebeTaskDefinition(child)
			extElement["task_definition"] = taskDef
		}

		extensions = append(extensions, extElement)
	}

	result["extensions"] = extensions
	return result
}

// parseZeebeProperties parses Zeebe properties
// Парсинг свойств Zeebe
func (p *ProcessParser) parseZeebeProperties(element *parser.XMLElement) []map[string]string {
	properties := make([]map[string]string, 0)

	for _, child := range element.Children {
		if child.XMLName.Local == "property" {
			property := make(map[string]string)
			for _, attr := range child.Attributes {
				property[attr.Name.Local] = attr.Value
			}
			if child.Text != "" {
				property["text"] = child.Text
			}
			properties = append(properties, property)
		}
	}

	return properties
}

// parseZeebeTaskDefinition parses Zeebe task definition
// Парсинг определения задачи Zeebe
func (p *ProcessParser) parseZeebeTaskDefinition(element *parser.XMLElement) map[string]string {
	taskDef := make(map[string]string)

	for _, attr := range element.Attributes {
		taskDef[attr.Name.Local] = attr.Value
	}

	return taskDef
}

// GetElementType returns element type
// Возвращает тип элемента
func (p *ProcessParser) GetElementType() string {
	return "process"
}
