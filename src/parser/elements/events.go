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

// EventParser parses all BPMN event elements
// Парсер всех элементов событий BPMN
type EventParser struct{}

// NewEventParser creates new event parser
// Создает новый парсер событий
func NewEventParser() *EventParser {
	return &EventParser{}
}

// Parse parses any BPMN event element
// Парсит любой элемент события BPMN
func (p *EventParser) Parse(element *parser.XMLElement, context *parser.ParseContext) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	eventType := element.XMLName.Local
	result["type"] = eventType
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
		case "cancelActivity":
			if cancel, err := strconv.ParseBool(attr.Value); err == nil {
				result["cancel_activity"] = cancel
			} else {
				result["cancel_activity"] = attr.Value
			}
		case "attachedToRef":
			result["attached_to_ref"] = attr.Value
		case "parallelMultiple":
			if parallel, err := strconv.ParseBool(attr.Value); err == nil {
				result["parallel_multiple"] = parallel
			} else {
				result["parallel_multiple"] = attr.Value
			}
		case "isInterrupting":
			if interrupting, err := strconv.ParseBool(attr.Value); err == nil {
				result["is_interrupting"] = interrupting
			} else {
				result["is_interrupting"] = attr.Value
			}
		}
	}

	// Store all attributes
	// Сохранение всех атрибутов
	result["attributes"] = attributes

	// Set default cancelActivity for boundary events if not specified
	// Устанавливаем default cancelActivity для boundary events если не указано
	if eventType == "boundaryEvent" {
		if _, exists := result["cancel_activity"]; !exists {
			result["cancel_activity"] = true // Default is interrupting
		}
	}

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

	// Parse event definitions
	// Парсинг определений событий
	eventDefinitions := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		if p.isEventDefinition(child.XMLName.Local) {
			eventDef := p.parseEventDefinition(child)
			eventDefinitions = append(eventDefinitions, eventDef)
		}
	}
	if len(eventDefinitions) > 0 {
		result["event_definitions"] = eventDefinitions
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

	return result, nil
}

// isEventDefinition checks if element is an event definition
// Проверяет является ли элемент определением события
func (p *EventParser) isEventDefinition(elementType string) bool {
	eventDefinitions := []string{
		"messageEventDefinition",
		"timerEventDefinition",
		"errorEventDefinition",
		"escalationEventDefinition",
		"cancelEventDefinition",
		"compensateEventDefinition",
		"conditionalEventDefinition",
		"linkEventDefinition",
		"signalEventDefinition",
		"terminateEventDefinition",
	}

	for _, def := range eventDefinitions {
		if elementType == def {
			return true
		}
	}
	return false
}

// parseEventDefinition parses any event definition
// Парсинг любого определения события
func (p *EventParser) parseEventDefinition(element *parser.XMLElement) map[string]interface{} {
	result := make(map[string]interface{})

	// Set definition type
	// Установка типа определения
	defType := element.XMLName.Local
	result["type"] = defType
	result["namespace"] = element.XMLName.Space

	// Parse all attributes
	// Парсинг всех атрибутов
	attributes := make(map[string]string)
	for _, attr := range element.Attributes {
		attributes[attr.Name.Local] = attr.Value

		// Store key attributes directly
		// Сохранение ключевых атрибутов напрямую
		switch attr.Name.Local {
		case "id":
			result["id"] = attr.Value
		case "messageRef", "signalRef", "errorRef", "escalationRef":
			result["reference"] = attr.Value
			result["reference_type"] = attr.Name.Local
		}
	}
	result["attributes"] = attributes

	// Parse specific event definition content
	// Парсинг специфичного содержимого определения события
	switch defType {
	case "timerEventDefinition":
		timer := p.parseTimerEventDefinition(element)
		result["timer"] = timer
	case "messageEventDefinition":
		message := p.parseMessageEventDefinition(element)
		result["message"] = message
	case "conditionalEventDefinition":
		condition := p.parseConditionalEventDefinition(element)
		result["condition"] = condition
	case "escalationEventDefinition":
		escalation := p.parseEscalationEventDefinition(element)
		result["escalation"] = escalation
	}

	return result
}

// parseTimerEventDefinition parses timer event definition
// Парсинг определения события таймера
func (p *EventParser) parseTimerEventDefinition(element *parser.XMLElement) map[string]interface{} {
	timer := make(map[string]interface{})

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "timeDuration":
			timer["duration"] = child.Text
			// Parse duration attributes
			// Парсинг атрибутов длительности
			for _, attr := range child.Attributes {
				if attr.Name.Local == "type" {
					timer["duration_type"] = attr.Value
				}
			}
		case "timeDate":
			timer["date"] = child.Text
		case "timeCycle":
			timer["cycle"] = child.Text
		}
	}

	return timer
}

// parseMessageEventDefinition parses message event definition
// Парсинг определения события сообщения
func (p *EventParser) parseMessageEventDefinition(element *parser.XMLElement) map[string]interface{} {
	message := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "messageRef" {
			message["message_ref"] = attr.Value
		}
	}

	return message
}

// parseConditionalEventDefinition parses conditional event definition
// Парсинг определения условного события
func (p *EventParser) parseConditionalEventDefinition(element *parser.XMLElement) map[string]interface{} {
	condition := make(map[string]interface{})

	for _, child := range element.Children {
		if child.XMLName.Local == "condition" {
			condition["expression"] = child.Text
			// Parse condition attributes
			// Парсинг атрибутов условия
			for _, attr := range child.Attributes {
				if attr.Name.Local == "type" {
					condition["expression_type"] = attr.Value
				}
			}
		}
	}

	return condition
}

// parseEscalationEventDefinition parses escalation event definition
// Парсинг определения события эскалации
func (p *EventParser) parseEscalationEventDefinition(element *parser.XMLElement) map[string]interface{} {
	escalation := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "escalationRef" {
			escalation["escalation_ref"] = attr.Value
		}
	}

	return escalation
}

// parseExtensionElements parses extension elements for events
// Парсинг элементов расширения для событий
func (p *EventParser) parseExtensionElements(element *parser.XMLElement) map[string]interface{} {
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

// GetElementType returns element type (generic for all events)
// Возвращает тип элемента (общий для всех событий)
func (p *EventParser) GetElementType() string {
	return "event"
}
