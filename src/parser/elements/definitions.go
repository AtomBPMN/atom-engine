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

// DefinitionsParser parses bpmn:definitions element
// Парсер элемента bpmn:definitions
type DefinitionsParser struct{}

// NewDefinitionsParser creates new definitions parser
// Создает новый парсер definitions
func NewDefinitionsParser() *DefinitionsParser {
	return &DefinitionsParser{}
}

// Parse parses bpmn:definitions element
// Парсит элемент bpmn:definitions
func (p *DefinitionsParser) Parse(
	element *parser.XMLElement,
	context *parser.ParseContext,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	result["type"] = "definitions"
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
		case "targetNamespace":
			result["target_namespace"] = attr.Value
		case "exporter":
			result["exporter"] = attr.Value
		case "exporterVersion":
			result["exporter_version"] = attr.Value
		}
	}

	// Store all attributes
	// Сохранение всех атрибутов
	result["attributes"] = attributes

	// Parse namespace declarations
	// Парсинг объявлений пространств имен
	namespaces := make(map[string]string)
	for _, attr := range element.Attributes {
		if attr.Name.Space == "xmlns" ||
			(attr.Name.Local == "xmlns" && attr.Name.Space == "") ||
			(attr.Name.Space == "" && len(attr.Name.Local) > 5 && attr.Name.Local[:5] == "xmlns") {
			namespaceKey := attr.Name.Local
			if attr.Name.Local == "xmlns" {
				namespaceKey = "default"
			}
			namespaces[namespaceKey] = attr.Value
		}
	}
	result["namespaces"] = namespaces

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

	// Count child elements
	// Подсчет дочерних элементов
	childCounts := make(map[string]int)
	for _, child := range element.Children {
		childCounts[child.XMLName.Local]++
	}
	result["child_counts"] = childCounts

	return result, nil
}

// parseExtensionElements parses extension elements
// Парсинг элементов расширения
func (p *DefinitionsParser) parseExtensionElements(element *parser.XMLElement) map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = "extensionElements"

	// Parse all child extension elements
	// Парсинг всех дочерних элементов расширения
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

// GetElementType returns element type
// Возвращает тип элемента
func (p *DefinitionsParser) GetElementType() string {
	return "definitions"
}
