/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"atom-engine/src/core/logger"
	"strconv"
)

// EventDefinitionParser parses all BPMN event definition elements
// Парсер всех элементов определений событий BPMN
type EventDefinitionParser struct{}

// NewEventDefinitionParser creates new event definition parser
// Создает новый парсер определений событий
func NewEventDefinitionParser() *EventDefinitionParser {
	return &EventDefinitionParser{}
}

// Parse parses any BPMN event definition element
// Парсит любой элемент определения события BPMN
func (p *EventDefinitionParser) Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error) {
	logger.Debug("Parsing event definition element",
		logger.String("element_type", element.XMLName.Local),
		logger.String("process_id", context.ProcessID))

	result := make(map[string]interface{})

	// Set element type and namespace
	// Установка типа элемента и пространства имен
	defType := element.XMLName.Local
	result["type"] = defType
	result["namespace"] = element.XMLName.Space
	result["parsed_with"] = "event_definition_parser"

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
		case "messageRef":
			result["message_ref"] = attr.Value
		case "signalRef":
			result["signal_ref"] = attr.Value
		case "errorRef":
			result["error_ref"] = attr.Value
		case "escalationRef":
			result["escalation_ref"] = attr.Value
		case "name":
			result["name"] = attr.Value
		}
	}
	result["attributes"] = attributes

	// Parse specific event definition content based on type
	// Парсинг специфичного содержимого определения события по типу
	switch defType {
	case "timerEventDefinition":
		timer := p.parseTimerEventDefinition(element)
		result["timer_data"] = timer
		logger.Debug("Parsed timer event definition",
			logger.String("element_id", result["id"].(string)),
			logger.Any("timer_data", timer))

	case "messageEventDefinition":
		message := p.parseMessageEventDefinition(element)
		result["message_data"] = message
		logger.Debug("Parsed message event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("message_data", message))

	case "signalEventDefinition":
		signal := p.parseSignalEventDefinition(element)
		result["signal_data"] = signal
		logger.Debug("Parsed signal event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("signal_data", signal))

	case "conditionalEventDefinition":
		condition := p.parseConditionalEventDefinition(element)
		result["condition_data"] = condition
		logger.Debug("Parsed conditional event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("condition_data", condition))

	case "errorEventDefinition":
		errorData := p.parseErrorEventDefinition(element)
		result["error_data"] = errorData
		logger.Debug("Parsed error event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("error_data", errorData))

	case "escalationEventDefinition":
		escalation := p.parseEscalationEventDefinition(element)
		result["escalation_data"] = escalation
		logger.Debug("Parsed escalation event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("escalation_data", escalation))

	case "compensateEventDefinition":
		compensate := p.parseCompensateEventDefinition(element)
		result["compensate_data"] = compensate
		logger.Debug("Parsed compensate event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("compensate_data", compensate))

	case "linkEventDefinition":
		link := p.parseLinkEventDefinition(element)
		result["link_data"] = link
		logger.Debug("Parsed link event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("link_data", link))

	case "terminateEventDefinition":
		terminate := p.parseTerminateEventDefinition(element)
		result["terminate_data"] = terminate
		logger.Debug("Parsed terminate event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("terminate_data", terminate))

	case "cancelEventDefinition":
		cancel := p.parseCancelEventDefinition(element)
		result["cancel_data"] = cancel
		logger.Debug("Parsed cancel event definition",
			logger.String("element_id", getStringValue(result["id"])),
			logger.Any("cancel_data", cancel))

	default:
		logger.Warn("Unknown event definition type",
			logger.String("element_type", defType),
			logger.String("element_id", getStringValue(result["id"])))
	}

	logger.Info("Successfully parsed event definition element",
		logger.String("element_type", defType),
		logger.String("element_id", getStringValue(result["id"])),
		logger.String("process_id", context.ProcessID))

	return result, nil
}

// parseTimerEventDefinition parses timer event definition with duration, cycle or date
// Парсинг определения события таймера с длительностью, циклом или датой
func (p *EventDefinitionParser) parseTimerEventDefinition(element *XMLElement) map[string]interface{} {
	timer := make(map[string]interface{})

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "timeDuration":
			timer["duration"] = child.Text
			timer["type"] = "duration"
			// Parse duration attributes
			// Парсинг атрибутов длительности
			for _, attr := range child.Attributes {
				if attr.Name.Local == "type" {
					timer["duration_type"] = attr.Value
				}
			}
			logger.Debug("Found timer duration",
				logger.String("duration", child.Text))

		case "timeDate":
			timer["date"] = child.Text
			timer["type"] = "date"
			logger.Debug("Found timer date",
				logger.String("date", child.Text))

		case "timeCycle":
			timer["cycle"] = child.Text
			timer["type"] = "cycle"
			logger.Debug("Found timer cycle",
				logger.String("cycle", child.Text))
		}
	}

	return timer
}

// parseMessageEventDefinition parses message event definition
// Парсинг определения события сообщения
func (p *EventDefinitionParser) parseMessageEventDefinition(element *XMLElement) map[string]interface{} {
	message := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "messageRef" {
			message["message_ref"] = attr.Value
			logger.Debug("Found message reference",
				logger.String("message_ref", attr.Value))
		}
	}

	return message
}

// parseSignalEventDefinition parses signal event definition
// Парсинг определения события сигнала
func (p *EventDefinitionParser) parseSignalEventDefinition(element *XMLElement) map[string]interface{} {
	signal := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "signalRef" {
			signal["signal_ref"] = attr.Value
			logger.Debug("Found signal reference",
				logger.String("signal_ref", attr.Value))
		}
	}

	return signal
}

// parseConditionalEventDefinition parses conditional event definition
// Парсинг определения условного события
func (p *EventDefinitionParser) parseConditionalEventDefinition(element *XMLElement) map[string]interface{} {
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
			logger.Debug("Found condition expression",
				logger.String("expression", child.Text))
		}
	}

	return condition
}

// parseErrorEventDefinition parses error event definition
// Парсинг определения события ошибки
func (p *EventDefinitionParser) parseErrorEventDefinition(element *XMLElement) map[string]interface{} {
	errorData := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "errorRef" {
			errorData["error_ref"] = attr.Value
			logger.Debug("Found error reference",
				logger.String("error_ref", attr.Value))
		}
	}

	return errorData
}

// parseEscalationEventDefinition parses escalation event definition
// Парсинг определения события эскалации
func (p *EventDefinitionParser) parseEscalationEventDefinition(element *XMLElement) map[string]interface{} {
	escalation := make(map[string]interface{})

	for _, attr := range element.Attributes {
		if attr.Name.Local == "escalationRef" {
			escalation["escalation_ref"] = attr.Value
			logger.Debug("Found escalation reference",
				logger.String("escalation_ref", attr.Value))
		}
	}

	return escalation
}

// parseCompensateEventDefinition parses compensate event definition
// Парсинг определения события компенсации
func (p *EventDefinitionParser) parseCompensateEventDefinition(element *XMLElement) map[string]interface{} {
	compensate := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "activityRef":
			compensate["activity_ref"] = attr.Value
			logger.Debug("Found compensate activity reference",
				logger.String("activity_ref", attr.Value))
		case "waitForCompletion":
			if wait, err := strconv.ParseBool(attr.Value); err == nil {
				compensate["wait_for_completion"] = wait
			} else {
				compensate["wait_for_completion"] = attr.Value
			}
			logger.Debug("Found compensate wait for completion",
				logger.String("wait_for_completion", attr.Value))
		}
	}

	return compensate
}

// parseLinkEventDefinition parses link event definition
// Парсинг определения события связи
func (p *EventDefinitionParser) parseLinkEventDefinition(element *XMLElement) map[string]interface{} {
	link := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "name":
			link["name"] = attr.Value
			logger.Debug("Found link name",
				logger.String("name", attr.Value))
		case "source":
			link["source"] = attr.Value
			logger.Debug("Found link source",
				logger.String("source", attr.Value))
		case "target":
			link["target"] = attr.Value
			logger.Debug("Found link target",
				logger.String("target", attr.Value))
		}
	}

	// Parse source and target elements
	// Парсинг элементов источника и цели
	sources := make([]string, 0)
	targets := make([]string, 0)

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "source":
			if child.Text != "" {
				sources = append(sources, child.Text)
			}
		case "target":
			if child.Text != "" {
				targets = append(targets, child.Text)
			}
		}
	}

	if len(sources) > 0 {
		link["sources"] = sources
	}
	if len(targets) > 0 {
		link["targets"] = targets
	}

	return link
}

// parseTerminateEventDefinition parses terminate event definition
// Парсинг определения события завершения
func (p *EventDefinitionParser) parseTerminateEventDefinition(element *XMLElement) map[string]interface{} {
	terminate := make(map[string]interface{})

	// Terminate event definition usually has no special attributes or children
	// Определение события завершения обычно не имеет специальных атрибутов или дочерних элементов
	terminate["type"] = "terminate"

	logger.Debug("Parsed terminate event definition")

	return terminate
}

// parseCancelEventDefinition parses cancel event definition
// Парсинг определения события отмены
func (p *EventDefinitionParser) parseCancelEventDefinition(element *XMLElement) map[string]interface{} {
	cancel := make(map[string]interface{})

	// Cancel event definition usually has no special attributes or children
	// Определение события отмены обычно не имеет специальных атрибутов или дочерних элементов
	cancel["type"] = "cancel"

	logger.Debug("Parsed cancel event definition")

	return cancel
}

// GetElementType returns element type for event definitions
// Возвращает тип элемента для определений событий
func (p *EventDefinitionParser) GetElementType() string {
	return "event_definition"
}

// getStringValue safely gets string value from interface{}
// Безопасно получает строковое значение из interface{}
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
