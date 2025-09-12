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

// isBusinessElement checks if element is a core BPMN business element
// Проверяет является ли элемент основным бизнес-элементом BPMN
func (p *BPMNParser) isBusinessElement(elementType string) bool {
	businessElements := []string{
		// Events
		"startEvent", "endEvent", "intermediateCatchEvent", "intermediateThrowEvent", "boundaryEvent",
		// Activities
		"task", "userTask", "serviceTask", "scriptTask", "sendTask", "receiveTask",
		"manualTask", "businessRuleTask", "callActivity", "subProcess",
		// Gateways
		"exclusiveGateway", "parallelGateway", "inclusiveGateway", "complexGateway", "eventBasedGateway",
		// Flows
		"sequenceFlow", "messageFlow", "association",
		// Data
		"dataObject", "dataStore", "dataStoreReference",
		// Structural
		"process", "collaboration", "participant",
	}

	for _, bizElement := range businessElements {
		if elementType == bizElement {
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

// extractProcessVersionFromXML extracts process version from XML root
// Извлекает версию процесса из корня XML
func (p *BPMNParser) extractProcessVersionFromXML(root *XMLElement) int {
	// Priority 1: Look for <zeebe:versionTag> in process extensionElements
	// Приоритет 1: Поиск <zeebe:versionTag> в extensionElements процесса
	if versionTag := p.findZeebeVersionTag(root); versionTag != "" {
		if version := p.parseVersionString(versionTag); version > 0 {
			return version
		}
	}

	// Priority 2: Look for version attribute in bpmn:process element
	// Приоритет 2: Поиск атрибута version в элементе bpmn:process
	if processVersion := p.findElementAttribute(root, "process", "version"); processVersion != "" {
		if version := p.parseVersionString(processVersion); version > 0 {
			return version
		}
	}

	// Priority 3: Look for version attribute in bpmn:definitions element
	// Приоритет 3: Поиск атрибута version в элементе bpmn:definitions
	if root.XMLName.Local == "definitions" {
		for _, attr := range root.Attributes {
			if attr.Name.Local == "version" {
				if version := p.parseVersionString(attr.Value); version > 0 {
					return version
				}
			}
		}
	}

	// Priority 4: Look in definitions child if current element is not definitions
	// Приоритет 4: Поиск в дочернем definitions если текущий элемент не definitions
	if definitionsVersion := p.findElementAttribute(root, "definitions", "version"); definitionsVersion != "" {
		if version := p.parseVersionString(definitionsVersion); version > 0 {
			return version
		}
	}

	// Default fallback to version 1
	// Значение по умолчанию - версия 1
	return 1
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

// findZeebeVersionTag finds zeebe:versionTag in process extensionElements
// Находит zeebe:versionTag в extensionElements процесса
func (p *BPMNParser) findZeebeVersionTag(element *XMLElement) string {
	// Look for process element
	// Поиск элемента process
	if processElement := p.findProcessElement(element); processElement != nil {
		// Look for extensionElements in process
		// Поиск extensionElements в процессе
		for _, child := range processElement.Children {
			if child.XMLName.Local == "extensionElements" {
				// Look for versionTag in extensionElements
				// Поиск versionTag в extensionElements
				return p.findVersionTagInExtensions(child)
			}
		}
	}
	return ""
}

// findProcessElement recursively finds the process element
// Рекурсивно находит элемент process
func (p *BPMNParser) findProcessElement(element *XMLElement) *XMLElement {
	if element.XMLName.Local == "process" {
		return element
	}

	for _, child := range element.Children {
		if result := p.findProcessElement(child); result != nil {
			return result
		}
	}

	return nil
}

// findVersionTagInExtensions finds versionTag in extensionElements
// Находит versionTag в extensionElements
func (p *BPMNParser) findVersionTagInExtensions(extensionElements *XMLElement) string {
	for _, child := range extensionElements.Children {
		if child.XMLName.Local == "versionTag" {
			// Look for value attribute
			// Поиск атрибута value
			for _, attr := range child.Attributes {
				if attr.Name.Local == "value" {
					return attr.Value
				}
			}
		}
	}
	return ""
}

// parseVersionString parses version string into integer
// Парсит строку версии в целое число
func (p *BPMNParser) parseVersionString(versionStr string) int {
	if versionStr == "" {
		return 0
	}

	// Remove common prefixes like "v", "V", "version"
	// Удаление общих префиксов как "v", "V", "version"
	cleanVersion := versionStr
	if len(cleanVersion) > 1 {
		if cleanVersion[0] == 'v' || cleanVersion[0] == 'V' {
			cleanVersion = cleanVersion[1:]
		}
	}

	// Handle version formats like "1.0", "1.2.3" - take first number
	// Обработка форматов версий как "1.0", "1.2.3" - берем первое число
	var result int
	for i, char := range cleanVersion {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else if char == '.' && result > 0 {
			// Stop at first dot if we already have a number
			// Останавливаемся на первой точке если уже есть число
			break
		} else if result > 0 {
			// Stop at first non-digit if we already have a number
			// Останавливаемся на первом не-цифре если уже есть число
			break
		}
		// Continue if haven't found any digits yet
		// Продолжаем если еще не нашли цифры
		if i == 0 && (char < '0' || char > '9') {
			continue
		}
	}

	return result
}
