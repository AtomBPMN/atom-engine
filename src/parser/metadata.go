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

// MetadataParser parses all BPMN metadata and extension elements
// Парсер всех метаданных и элементов расширения BPMN
type MetadataParser struct{}

// NewMetadataParser creates new metadata parser
// Создает новый парсер метаданных
func NewMetadataParser() *MetadataParser {
	return &MetadataParser{}
}

// Parse parses any BPMN metadata element
// Парсит любой элемент метаданных BPMN
func (p *MetadataParser) Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error) {
	logger.Debug("Parsing metadata element",
		logger.String("element_type", element.XMLName.Local),
		logger.String("process_id", context.ProcessID))

	result := make(map[string]interface{})

	// Set element type and namespace
	// Установка типа элемента и пространства имен
	metaType := element.XMLName.Local
	result["type"] = metaType
	result["namespace"] = element.XMLName.Space
	result["parsed_with"] = "metadata_parser"

	// Parse all attributes
	// Парсинг всех атрибутов
	attributes := make(map[string]string)
	for _, attr := range element.Attributes {
		attributes[attr.Name.Local] = attr.Value
	}
	result["attributes"] = attributes

	// Parse specific metadata content based on type
	// Парсинг специфичного содержимого метаданных по типу
	switch metaType {
	case "properties":
		properties := p.parseProperties(element)
		result["properties_data"] = properties
		logger.Debug("Parsed properties element",
			logger.Int("property_count", len(properties)))

	case "property":
		property := p.parseProperty(element)
		result["property_data"] = property
		logger.Debug("Parsed property element",
			logger.String("name", getStringValue(property["name"])),
			logger.String("value", getStringValue(property["value"])))

	case "taskDefinition":
		taskDef := p.parseTaskDefinition(element)
		result["task_definition_data"] = taskDef
		logger.Debug("Parsed task definition element",
			logger.String("type", getStringValue(taskDef["type"])),
			logger.Int("header_count", getIntValue(taskDef["header_count"])))

	case "subscription":
		subscription := p.parseSubscription(element)
		result["subscription_data"] = subscription
		logger.Debug("Parsed subscription element",
			logger.String("correlation_key", getStringValue(subscription["correlation_key"])))

	case "formDefinition":
		formDef := p.parseFormDefinition(element)
		result["form_definition_data"] = formDef
		logger.Debug("Parsed form definition element",
			logger.String("form_id", getStringValue(formDef["form_id"])))

	case "calledElement":
		calledElement := p.parseCalledElement(element)
		result["called_element_data"] = calledElement
		logger.Debug("Parsed called element",
			logger.String("process_id", getStringValue(calledElement["process_id"])))

	case "ioMapping":
		ioMapping := p.parseIOMapping(element)
		result["io_mapping_data"] = ioMapping
		logger.Debug("Parsed IO mapping element",
			logger.Int("input_count", getIntValue(ioMapping["input_count"])),
			logger.Int("output_count", getIntValue(ioMapping["output_count"])))

	case "input":
		input := p.parseInput(element)
		result["input_data"] = input
		logger.Debug("Parsed input element",
			logger.String("source", getStringValue(input["source"])),
			logger.String("target", getStringValue(input["target"])))

	case "output":
		output := p.parseOutput(element)
		result["output_data"] = output
		logger.Debug("Parsed output element",
			logger.String("source", getStringValue(output["source"])),
			logger.String("target", getStringValue(output["target"])))

	case "header":
		header := p.parseHeader(element)
		result["header_data"] = header
		logger.Debug("Parsed header element",
			logger.String("key", getStringValue(header["key"])),
			logger.String("value", getStringValue(header["value"])))

	case "script":
		script := p.parseScript(element)
		result["script_data"] = script
		logger.Debug("Parsed script element",
			logger.String("expression", getStringValue(script["expression"])),
			logger.String("result_variable", getStringValue(script["result_variable"])))

	case "assignmentDefinition":
		assignment := p.parseAssignmentDefinition(element)
		result["assignment_data"] = assignment
		logger.Debug("Parsed assignment definition element",
			logger.String("assignee", getStringValue(assignment["assignee"])),
			logger.String("candidate_groups", getStringValue(assignment["candidate_groups"])))

	case "userTask":
		userTask := p.parseUserTask(element)
		result["user_task_data"] = userTask
		logger.Debug("Parsed user task element")

	default:
		logger.Warn("Unknown metadata element type",
			logger.String("element_type", metaType))
	}

	// Add text content if present
	// Добавление текстового содержимого если есть
	if element.Text != "" {
		result["text"] = element.Text
	}

	logger.Info("Successfully parsed metadata element",
		logger.String("element_type", metaType),
		logger.String("process_id", context.ProcessID))

	return result, nil
}

// parseProperties parses zeebe:properties element
// Парсинг элемента zeebe:properties
func (p *MetadataParser) parseProperties(element *XMLElement) []map[string]interface{} {
	properties := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		if child.XMLName.Local == "property" {
			property := p.parseProperty(child)
			properties = append(properties, property)
		}
	}

	logger.Debug("Parsed properties container",
		logger.Int("property_count", len(properties)))

	return properties
}

// parseProperty parses zeebe:property element
// Парсинг элемента zeebe:property
func (p *MetadataParser) parseProperty(element *XMLElement) map[string]interface{} {
	property := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "name":
			property["name"] = attr.Value
		case "value":
			property["value"] = attr.Value
		}
	}

	return property
}

// parseTaskDefinition parses zeebe:taskDefinition element
// Парсинг элемента zeebe:taskDefinition
func (p *MetadataParser) parseTaskDefinition(element *XMLElement) map[string]interface{} {
	taskDef := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "type":
			taskDef["type"] = attr.Value
		case "retries":
			if retries, err := strconv.Atoi(attr.Value); err == nil {
				taskDef["retries"] = retries
			} else {
				taskDef["retries"] = attr.Value
			}
		}
	}

	// Parse headers
	// Парсинг заголовков
	headers := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		if child.XMLName.Local == "header" {
			header := p.parseHeader(child)
			headers = append(headers, header)
		}
	}

	if len(headers) > 0 {
		taskDef["headers"] = headers
		taskDef["header_count"] = len(headers)
	}

	return taskDef
}

// parseSubscription parses zeebe:subscription element
// Парсинг элемента zeebe:subscription
func (p *MetadataParser) parseSubscription(element *XMLElement) map[string]interface{} {
	subscription := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "correlationKey":
			subscription["correlation_key"] = attr.Value
		case "messageName":
			subscription["message_name"] = attr.Value
		}
	}

	return subscription
}

// parseFormDefinition parses zeebe:formDefinition element
// Парсинг элемента zeebe:formDefinition
func (p *MetadataParser) parseFormDefinition(element *XMLElement) map[string]interface{} {
	formDef := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "formId":
			formDef["form_id"] = attr.Value
		case "formKey":
			formDef["form_key"] = attr.Value
		case "externalReference":
			formDef["external_reference"] = attr.Value
		}
	}

	return formDef
}

// parseCalledElement parses zeebe:calledElement element
// Парсинг элемента zeebe:calledElement
func (p *MetadataParser) parseCalledElement(element *XMLElement) map[string]interface{} {
	calledElement := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "processId":
			calledElement["process_id"] = attr.Value
		case "propagateAllChildVariables":
			if propagate, err := strconv.ParseBool(attr.Value); err == nil {
				calledElement["propagate_all_child_variables"] = propagate
			} else {
				calledElement["propagate_all_child_variables"] = attr.Value
			}
		case "propagateAllParentVariables":
			if propagate, err := strconv.ParseBool(attr.Value); err == nil {
				calledElement["propagate_all_parent_variables"] = propagate
			} else {
				calledElement["propagate_all_parent_variables"] = attr.Value
			}
		}
	}

	return calledElement
}

// parseIOMapping parses zeebe:ioMapping element
// Парсинг элемента zeebe:ioMapping
func (p *MetadataParser) parseIOMapping(element *XMLElement) map[string]interface{} {
	ioMapping := make(map[string]interface{})

	inputs := make([]map[string]interface{}, 0)
	outputs := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "input":
			input := p.parseInput(child)
			inputs = append(inputs, input)
		case "output":
			output := p.parseOutput(child)
			outputs = append(outputs, output)
		}
	}

	if len(inputs) > 0 {
		ioMapping["inputs"] = inputs
		ioMapping["input_count"] = len(inputs)
	}
	if len(outputs) > 0 {
		ioMapping["outputs"] = outputs
		ioMapping["output_count"] = len(outputs)
	}

	return ioMapping
}

// parseInput parses zeebe:input element
// Парсинг элемента zeebe:input
func (p *MetadataParser) parseInput(element *XMLElement) map[string]interface{} {
	input := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "source":
			input["source"] = attr.Value
		case "target":
			input["target"] = attr.Value
		}
	}

	return input
}

// parseOutput parses zeebe:output element
// Парсинг элемента zeebe:output
func (p *MetadataParser) parseOutput(element *XMLElement) map[string]interface{} {
	output := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "source":
			output["source"] = attr.Value
		case "target":
			output["target"] = attr.Value
		}
	}

	return output
}

// parseHeader parses zeebe:header element
// Парсинг элемента zeebe:header
func (p *MetadataParser) parseHeader(element *XMLElement) map[string]interface{} {
	header := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "key":
			header["key"] = attr.Value
		case "value":
			header["value"] = attr.Value
		}
	}

	return header
}

// parseScript parses zeebe:script element
// Парсинг элемента zeebe:script
func (p *MetadataParser) parseScript(element *XMLElement) map[string]interface{} {
	script := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "expression":
			script["expression"] = attr.Value
		case "resultVariable":
			script["result_variable"] = attr.Value
		}
	}

	return script
}

// parseAssignmentDefinition parses zeebe:assignmentDefinition element
// Парсинг элемента zeebe:assignmentDefinition
func (p *MetadataParser) parseAssignmentDefinition(element *XMLElement) map[string]interface{} {
	assignment := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "assignee":
			assignment["assignee"] = attr.Value
		case "candidateGroups":
			assignment["candidate_groups"] = attr.Value
		case "candidateUsers":
			assignment["candidate_users"] = attr.Value
		case "dueDate":
			assignment["due_date"] = attr.Value
		case "followUpDate":
			assignment["follow_up_date"] = attr.Value
		}
	}

	return assignment
}

// parseUserTask parses zeebe:userTask element
// Парсинг элемента zeebe:userTask
func (p *MetadataParser) parseUserTask(element *XMLElement) map[string]interface{} {
	userTask := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "type":
			userTask["type"] = attr.Value
		}
	}

	// UserTask is usually just a marker element
	// userTask обычно просто элемент-маркер
	userTask["marker"] = true

	return userTask
}

// GetElementType returns element type for metadata
// Возвращает тип элемента для метаданных
func (p *MetadataParser) GetElementType() string {
	return "metadata"
}

// getIntValue safely gets int value from interface{}
// Безопасно получает целое значение из interface{}
func getIntValue(value interface{}) int {
	if value == nil {
		return 0
	}
	if intVal, ok := value.(int); ok {
		return intVal
	}
	return 0
}
