/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"encoding/json"
	"fmt"

	"atom-engine/src/core/models"
)

// isDiagramElement checks if element is a diagram element
// Проверяет является ли элемент элементом диаграммы
func (p *BPMNParser) isDiagramElement(elementType string) bool {
	diagramElements := []string{
		"BPMNDiagram", "BPMNPlane", "BPMNShape", "BPMNEdge", "BPMNLabel",
		"Bounds", "waypoint", "label",
	}

	for _, diagElement := range diagramElements {
		if elementType == diagElement {
			return true
		}
	}
	return false
}

// isKnownMetadataElement checks if element is a known metadata element
// Проверяет является ли элемент известным метаданным элементом
func (p *BPMNParser) isKnownMetadataElement(elementType string) bool {
	metadataElements := []string{
		"documentation", "extensionElements", "incoming", "outgoing",
		"conditionExpression", "script", "condition", "property",
	}

	for _, metaElement := range metadataElements {
		if elementType == metaElement {
			return true
		}
	}
	return false
}

// parseGenericElement parses any element generically
// Общий парсинг любого элемента
func (p *BPMNParser) parseGenericElement(element *XMLElement) map[string]interface{} {
	result := make(map[string]interface{})

	// Add element type
	// Добавление типа элемента
	result["type"] = element.XMLName.Local
	result["namespace"] = element.XMLName.Space
	result["parsed_with"] = "generic_parser"

	// Add all attributes
	// Добавление всех атрибутов
	attributes := make(map[string]string)
	for _, attr := range element.Attributes {
		attributes[attr.Name.Local] = attr.Value
	}
	if len(attributes) > 0 {
		result["attributes"] = attributes
	}

	// Add text content if present
	// Добавление текстового содержимого если есть
	if element.Text != "" {
		result["text"] = element.Text
	}

	// Parse child elements as raw data
	// Парсинг дочерних элементов как необработанные данные
	if len(element.Children) > 0 {
		children := make([]map[string]interface{}, 0)
		for _, child := range element.Children {
			childData := make(map[string]interface{})
			childData["tag"] = child.XMLName.Local
			childData["namespace"] = child.XMLName.Space
			childData["text"] = child.Text

			// Child attributes
			// Атрибуты дочернего элемента
			childAttrs := make(map[string]string)
			for _, attr := range child.Attributes {
				childAttrs[attr.Name.Local] = attr.Value
			}
			if len(childAttrs) > 0 {
				childData["attributes"] = childAttrs
			}

			children = append(children, childData)
		}
		result["children"] = children
	}

	return result
}

// getElementID extracts ID attribute from element
// Извлекает ID атрибут из элемента
func (p *BPMNParser) getElementID(element *XMLElement) string {
	for _, attr := range element.Attributes {
		if attr.Name.Local == "id" {
			return attr.Value
		}
	}
	return ""
}

// extractProcessIDFromXML extracts process ID from XML root
// Извлекает ID процесса из корня XML
func (p *BPMNParser) extractProcessIDFromXML(root *XMLElement) string {
	// Look for bpmn:process element and extract id
	// Поиск элемента bpmn:process и извлечение id
	return p.findElementAttribute(root, "process", "id")
}

// extractProcessNameFromXML extracts process name from XML root
// Извлекает имя процесса из корня XML
func (p *BPMNParser) extractProcessNameFromXML(root *XMLElement) string {
	// Look for bpmn:process element and extract name
	// Поиск элемента bpmn:process и извлечение name
	return p.findElementAttribute(root, "process", "name")
}

// extractNamespaces extracts all namespace declarations
// Извлекает все объявления пространств имен
func (p *BPMNParser) extractNamespaces(root *XMLElement) map[string]string {
	namespaces := make(map[string]string)
	for _, attr := range root.Attributes {
		if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" {
			namespaces[attr.Name.Local] = attr.Value
		}
	}
	return namespaces
}

// findElementAttribute recursively finds element and returns attribute value
// Рекурсивно находит элемент и возвращает значение атрибута
func (p *BPMNParser) findElementAttribute(element *XMLElement, elementType, attrName string) string {
	if element.XMLName.Local == elementType {
		for _, attr := range element.Attributes {
			if attr.Name.Local == attrName {
				return attr.Value
			}
		}
	}

	for _, child := range element.Children {
		if result := p.findElementAttribute(child, elementType, attrName); result != "" {
			return result
		}
	}

	return ""
}

// ToJSON converts parsed BPMN to JSON
// Конвертирует спарсенный BPMN в JSON
func (p *BPMNParser) ToJSON() ([]byte, error) {
	if p.processData == nil {
		return nil, fmt.Errorf("no process data to convert")
	}

	return json.MarshalIndent(p.processData, "", "  ")
}

// GetProcessData returns parsed process data
// Возвращает данные спарсенного процесса
func (p *BPMNParser) GetProcessData() *models.BPMNProcess {
	return p.processData
}
