/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"strconv"
)

// TaskParser parses all BPMN task elements
// Парсер всех элементов задач BPMN
type TaskParser struct{}

// NewTaskParser creates new task parser
// Создает новый парсер задач
func NewTaskParser() *TaskParser {
	return &TaskParser{}
}

// Parse parses any BPMN task element
// Парсит любой элемент задачи BPMN
func (p *TaskParser) Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Set element type
	// Установка типа элемента
	taskType := element.XMLName.Local
	result["type"] = taskType
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
		case "isForCompensation":
			if compensation, err := strconv.ParseBool(attr.Value); err == nil {
				result["is_for_compensation"] = compensation
			} else {
				result["is_for_compensation"] = attr.Value
			}
		case "startQuantity":
			if quantity, err := strconv.Atoi(attr.Value); err == nil {
				result["start_quantity"] = quantity
			} else {
				result["start_quantity"] = attr.Value
			}
		case "completionQuantity":
			if quantity, err := strconv.Atoi(attr.Value); err == nil {
				result["completion_quantity"] = quantity
			} else {
				result["completion_quantity"] = attr.Value
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

	// Parse task-specific elements based on task type
	// Парсинг специфичных для задачи элементов основываясь на типе задачи
	switch taskType {
	case "scriptTask":
		script := p.parseScriptTask(element)
		result["script"] = script
	case "serviceTask":
		service := p.parseServiceTask(element)
		result["service"] = service
	case "sendTask":
		send := p.parseSendTask(element)
		result["send_task"] = send
	case "receiveTask":
		receive := p.parseReceiveTask(element)
		result["receive_task"] = receive
	case "userTask":
		user := p.parseUserTask(element)
		result["user_task"] = user
	case "callActivity":
		call := p.parseCallActivity(element)
		result["call_activity"] = call
	case "subProcess":
		subprocess := p.parseSubProcess(element)
		result["subprocess"] = subprocess
	}

	// Parse I/O specifications
	// Парсинг спецификаций ввода/вывода
	ioSpec := p.parseIOSpecification(element)
	if len(ioSpec) > 0 {
		result["io_specification"] = ioSpec
	}

	return result, nil
}

// parseScriptTask parses script task specific elements
// Парсинг специфичных элементов скриптовой задачи
func (p *TaskParser) parseScriptTask(element *XMLElement) map[string]interface{} {
	script := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "scriptFormat":
			script["format"] = attr.Value
		}
	}

	for _, child := range element.Children {
		if child.XMLName.Local == "script" && child.Text != "" {
			script["content"] = child.Text
		}
	}

	return script
}

// parseServiceTask parses service task specific elements
// Парсинг специфичных элементов сервисной задачи
func (p *TaskParser) parseServiceTask(element *XMLElement) map[string]interface{} {
	service := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "operationRef":
			service["operation_ref"] = attr.Value
		case "implementation":
			service["implementation"] = attr.Value
		}
	}

	return service
}

// parseUserTask parses user task specific elements
// Парсинг специфичных элементов пользовательской задачи
func (p *TaskParser) parseUserTask(element *XMLElement) map[string]interface{} {
	user := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "implementation":
			user["implementation"] = attr.Value
		}
	}

	return user
}

// parseSendTask parses send task specific elements
// Парсинг специфичных элементов задачи отправки
func (p *TaskParser) parseSendTask(element *XMLElement) map[string]interface{} {
	send := make(map[string]interface{})

	// Parse direct attributes
	// Парсинг прямых атрибутов
	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "messageRef":
			send["message_ref"] = attr.Value
		case "operationRef":
			send["operation_ref"] = attr.Value
		case "implementation":
			send["implementation"] = attr.Value
		}
	}

	// Extract task type from extension elements taskDefinition
	// Извлечение типа задачи из taskDefinition в элементах расширения
	for _, child := range element.Children {
		if child.XMLName.Local == "extensionElements" {
			for _, extChild := range child.Children {
				if extChild.XMLName.Local == "taskDefinition" {
					for _, attr := range extChild.Attributes {
						if attr.Name.Local == "type" {
							send["task_type"] = attr.Value
						}
					}
				}
			}
		}
	}

	return send
}

// parseReceiveTask parses receive task specific elements
// Парсинг специфичных элементов задачи получения
func (p *TaskParser) parseReceiveTask(element *XMLElement) map[string]interface{} {
	receive := make(map[string]interface{})

	// Parse direct attributes
	// Парсинг прямых атрибутов
	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "messageRef":
			receive["message_ref"] = attr.Value
		case "operationRef":
			receive["operation_ref"] = attr.Value
		case "implementation":
			receive["implementation"] = attr.Value
		case "instantiate":
			if instantiate, err := strconv.ParseBool(attr.Value); err == nil {
				receive["instantiate"] = instantiate
			} else {
				receive["instantiate"] = attr.Value
			}
		}
	}

	return receive
}

// parseCallActivity parses call activity specific elements
// Парсинг специфичных элементов вызова активности
func (p *TaskParser) parseCallActivity(element *XMLElement) map[string]interface{} {
	call := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "calledElement":
			call["called_element"] = attr.Value
		}
	}

	return call
}

// parseSubProcess parses subprocess specific elements
// Парсинг специфичных элементов подпроцесса
func (p *TaskParser) parseSubProcess(element *XMLElement) map[string]interface{} {
	subprocess := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "triggeredByEvent":
			if triggered, err := strconv.ParseBool(attr.Value); err == nil {
				subprocess["triggered_by_event"] = triggered
			} else {
				subprocess["triggered_by_event"] = attr.Value
			}
		}
	}

	// Count child flow elements in subprocess
	// Подсчет дочерних элементов потока в подпроцессе
	childElements := make(map[string]int)
	for _, child := range element.Children {
		childElements[child.XMLName.Local]++
	}
	subprocess["child_elements"] = childElements

	return subprocess
}

// parseIOSpecification parses I/O specification
// Парсинг спецификации ввода/вывода
func (p *TaskParser) parseIOSpecification(element *XMLElement) map[string]interface{} {
	ioSpec := make(map[string]interface{})

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "ioSpecification":
			// Parse input/output sets
			// Парсинг наборов входов/выходов
			inputs := make([]map[string]interface{}, 0)
			outputs := make([]map[string]interface{}, 0)

			for _, grandChild := range child.Children {
				switch grandChild.XMLName.Local {
				case "dataInput":
					input := p.parseDataInput(grandChild)
					inputs = append(inputs, input)
				case "dataOutput":
					output := p.parseDataOutput(grandChild)
					outputs = append(outputs, output)
				}
			}

			if len(inputs) > 0 {
				ioSpec["inputs"] = inputs
			}
			if len(outputs) > 0 {
				ioSpec["outputs"] = outputs
			}
		}
	}

	return ioSpec
}

// parseDataInput parses data input element
// Парсинг элемента входных данных
func (p *TaskParser) parseDataInput(element *XMLElement) map[string]interface{} {
	input := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			input["id"] = attr.Value
		case "name":
			input["name"] = attr.Value
		case "isCollection":
			if collection, err := strconv.ParseBool(attr.Value); err == nil {
				input["is_collection"] = collection
			}
		}
	}

	return input
}

// parseDataOutput parses data output element
// Парсинг элемента выходных данных
func (p *TaskParser) parseDataOutput(element *XMLElement) map[string]interface{} {
	output := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			output["id"] = attr.Value
		case "name":
			output["name"] = attr.Value
		case "isCollection":
			if collection, err := strconv.ParseBool(attr.Value); err == nil {
				output["is_collection"] = collection
			}
		}
	}

	return output
}

// parseExtensionElements parses extension elements for tasks
// Парсинг элементов расширения для задач
func (p *TaskParser) parseExtensionElements(element *XMLElement) map[string]interface{} {
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

		// Handle specific Zeebe task extensions
		// Обработка специфичных расширений задач Zeebe
		switch child.XMLName.Local {
		case "taskDefinition":
			taskDef := p.parseZeebeTaskDefinition(child)
			extElement["task_definition"] = taskDef
		case "script":
			script := p.parseZeebeScript(child)
			extElement["script"] = script
		case "calledElement":
			calledElement := p.parseZeebeCalledElement(child)
			extElement["called_element"] = calledElement
		case "formDefinition":
			formDef := p.parseZeebeFormDefinition(child)
			extElement["form_definition"] = formDef
		case "ioMapping":
			ioMapping := p.parseZeebeIOMapping(child)
			extElement["io_mapping"] = ioMapping
		case "taskHeaders":
			taskHeaders := p.parseZeebeTaskHeaders(child)
			extElement["task_headers"] = taskHeaders
		}

		extensions = append(extensions, extElement)
	}

	result["extensions"] = extensions
	return result
}

// parseZeebeTaskDefinition parses Zeebe task definition
// Парсинг определения задачи Zeebe
func (p *TaskParser) parseZeebeTaskDefinition(element *XMLElement) map[string]interface{} {
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

	return taskDef
}

// parseZeebeScript parses Zeebe script element
// Парсинг элемента скрипта Zeebe
func (p *TaskParser) parseZeebeScript(element *XMLElement) map[string]interface{} {
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

// parseZeebeCalledElement parses Zeebe called element
// Парсинг вызываемого элемента Zeebe
func (p *TaskParser) parseZeebeCalledElement(element *XMLElement) map[string]interface{} {
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
		}
	}

	return calledElement
}

// parseZeebeFormDefinition parses Zeebe form definition
// Парсинг определения формы Zeebe
func (p *TaskParser) parseZeebeFormDefinition(element *XMLElement) map[string]interface{} {
	formDef := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "formKey":
			formDef["form_key"] = attr.Value
		case "externalReference":
			formDef["external_reference"] = attr.Value
		}
	}

	return formDef
}

// parseZeebeIOMapping parses Zeebe ioMapping element
// Парсинг ioMapping элемента Zeebe
func (p *TaskParser) parseZeebeIOMapping(element *XMLElement) map[string]interface{} {
	ioMapping := make(map[string]interface{})

	inputs := make([]map[string]interface{}, 0)
	outputs := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "input":
			input := p.parseZeebeInput(child)
			inputs = append(inputs, input)
		case "output":
			output := p.parseZeebeOutput(child)
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

// parseZeebeInput parses Zeebe input element
// Парсинг input элемента Zeebe
func (p *TaskParser) parseZeebeInput(element *XMLElement) map[string]interface{} {
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

// parseZeebeOutput parses Zeebe output element
// Парсинг output элемента Zeebe
func (p *TaskParser) parseZeebeOutput(element *XMLElement) map[string]interface{} {
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

// parseZeebeTaskHeaders parses Zeebe taskHeaders element
// Парсинг taskHeaders элемента Zeebe
func (p *TaskParser) parseZeebeTaskHeaders(element *XMLElement) map[string]interface{} {
	taskHeaders := make(map[string]interface{})

	headers := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		if child.XMLName.Local == "header" {
			header := p.parseZeebeHeader(child)
			headers = append(headers, header)
		}
	}

	if len(headers) > 0 {
		taskHeaders["headers"] = headers
		taskHeaders["header_count"] = len(headers)
	}

	return taskHeaders
}

// parseZeebeHeader parses Zeebe header element
// Парсинг header элемента Zeebe
func (p *TaskParser) parseZeebeHeader(element *XMLElement) map[string]interface{} {
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

// GetElementType returns element type (generic for all tasks)
// Возвращает тип элемента (общий для всех задач)
func (p *TaskParser) GetElementType() string {
	return "task"
}
