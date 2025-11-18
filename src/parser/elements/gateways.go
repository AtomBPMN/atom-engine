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

// GatewayParser parses all BPMN gateway elements
// Парсер всех элементов шлюзов BPMN
type GatewayParser struct{}

// NewGatewayParser creates new gateway parser
// Создает новый парсер шлюзов
func NewGatewayParser() *GatewayParser {
	return &GatewayParser{}
}

// Parse parses any BPMN gateway element
// Парсит любой элемент шлюза BPMN
func (p *GatewayParser) Parse(
	element *parser.XMLElement,
	context *parser.ParseContext,
) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	gatewayType := element.XMLName.Local
	result["type"] = gatewayType
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
		case "gatewayDirection":
			result["gateway_direction"] = attr.Value
		case "default":
			result["default_flow"] = attr.Value
		case "instantiate":
			if instantiate, err := strconv.ParseBool(attr.Value); err == nil {
				result["instantiate"] = instantiate
			} else {
				result["instantiate"] = attr.Value
			}
		}
	}

	// Store all attributes
	// Сохранение всех атрибутов
	result["attributes"] = attributes

	// Parse incoming and outgoing flows
	// Парсинг входящих и исходящих потоков
	incoming := make([]string, 0)
	outgoing := make([]string, 0)

	for _, child := range element.Children {
		if child.XMLName.Local == "incoming" && child.Text != "" {
			incoming = append(incoming, child.Text)
		} else if child.XMLName.Local == "outgoing" && child.Text != "" {
			outgoing = append(outgoing, child.Text)
		}
	}

	if len(incoming) > 0 {
		result["incoming"] = incoming
	}
	if len(outgoing) > 0 {
		result["outgoing"] = outgoing
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

	// Parse gateway-specific configuration
	// Парсинг специфичной конфигурации шлюза
	gatewayConfig := p.parseGatewaySpecificConfig(element, gatewayType)
	if len(gatewayConfig) > 0 {
		result["gateway_config"] = gatewayConfig
	}

	// Count flow connections
	// Подсчет соединений потока
	result["incoming_count"] = len(incoming)
	result["outgoing_count"] = len(outgoing)

	return result, nil
}

// parseGatewaySpecificConfig parses gateway type specific configuration
// Парсинг специфичной для типа шлюза конфигурации
func (p *GatewayParser) parseGatewaySpecificConfig(
	element *parser.XMLElement,
	gatewayType string,
) map[string]interface{} {
	config := make(map[string]interface{})

	switch gatewayType {
	case "exclusiveGateway":
		config = p.parseExclusiveGateway(element)
	case "parallelGateway":
		config = p.parseParallelGateway(element)
	case "inclusiveGateway":
		config = p.parseInclusiveGateway(element)
	case "complexGateway":
		config = p.parseComplexGateway(element)
	case "eventBasedGateway":
		config = p.parseEventBasedGateway(element)
	}

	return config
}

// parseExclusiveGateway parses exclusive gateway specific elements
// Парсинг специфичных элементов исключающего шлюза
func (p *GatewayParser) parseExclusiveGateway(element *parser.XMLElement) map[string]interface{} {
	config := make(map[string]interface{})
	config["gateway_type"] = "exclusive"
	config["description"] = "Only one outgoing flow is activated"

	// Default flow is handled in main parsing
	// Поток по умолчанию обрабатывается в основном парсинге

	return config
}

// parseParallelGateway parses parallel gateway specific elements
// Парсинг специфичных элементов параллельного шлюза
func (p *GatewayParser) parseParallelGateway(element *parser.XMLElement) map[string]interface{} {
	config := make(map[string]interface{})
	config["gateway_type"] = "parallel"
	config["description"] = "All outgoing flows are activated simultaneously"

	return config
}

// parseInclusiveGateway parses inclusive gateway specific elements
// Парсинг специфичных элементов включающего шлюза
func (p *GatewayParser) parseInclusiveGateway(element *parser.XMLElement) map[string]interface{} {
	config := make(map[string]interface{})
	config["gateway_type"] = "inclusive"
	config["description"] = "One or more outgoing flows are activated based on conditions"

	return config
}

// parseComplexGateway parses complex gateway specific elements
// Парсинг специфичных элементов сложного шлюза
func (p *GatewayParser) parseComplexGateway(element *parser.XMLElement) map[string]interface{} {
	config := make(map[string]interface{})
	config["gateway_type"] = "complex"
	config["description"] = "Complex activation condition defined"

	// Parse activation condition if present
	// Парсинг условия активации если есть
	for _, child := range element.Children {
		if child.XMLName.Local == "activationCondition" {
			config["activation_condition"] = child.Text
		}
	}

	return config
}

// parseEventBasedGateway parses event-based gateway specific elements
// Парсинг специфичных элементов шлюза основанного на событиях
func (p *GatewayParser) parseEventBasedGateway(element *parser.XMLElement) map[string]interface{} {
	config := make(map[string]interface{})
	config["gateway_type"] = "event_based"
	config["description"] = "Waits for events to determine which path to follow"

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "eventGatewayType":
			config["event_gateway_type"] = attr.Value
		case "instantiate":
			if instantiate, err := strconv.ParseBool(attr.Value); err == nil {
				config["instantiate"] = instantiate
			} else {
				config["instantiate"] = attr.Value
			}
		}
	}

	return config
}

// parseExtensionElements parses extension elements for gateways
// Парсинг элементов расширения для шлюзов
func (p *GatewayParser) parseExtensionElements(element *parser.XMLElement) map[string]interface{} {
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

// GetElementType returns element type (generic for all gateways)
// Возвращает тип элемента (общий для всех шлюзов)
func (p *GatewayParser) GetElementType() string {
	return "gateway"
}
