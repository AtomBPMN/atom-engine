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

// StructureParser parses all BPMN structural elements (collaboration, participant, etc.)
// Парсер всех структурных элементов BPMN (collaboration, participant и т.д.)
type StructureParser struct{}

// NewStructureParser creates new structure parser
// Создает новый парсер структуры
func NewStructureParser() *StructureParser {
	return &StructureParser{}
}

// Parse parses any BPMN structural element
// Парсит любой структурный элемент BPMN
func (p *StructureParser) Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error) {
	logger.Debug("Parsing structure element",
		logger.String("element_type", element.XMLName.Local),
		logger.String("process_id", context.ProcessID))

	result := make(map[string]interface{})

	// Set element type and namespace
	// Установка типа элемента и пространства имен
	structType := element.XMLName.Local
	result["type"] = structType
	result["namespace"] = element.XMLName.Space
	result["parsed_with"] = "structure_parser"

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
		case "name":
			result["name"] = attr.Value
		case "processRef":
			result["process_ref"] = attr.Value
		case "isClosed":
			if closed, err := strconv.ParseBool(attr.Value); err == nil {
				result["is_closed"] = closed
			} else {
				result["is_closed"] = attr.Value
			}
		}
	}
	result["attributes"] = attributes

	// Parse specific structure types
	// Парсинг специфических типов структур
	switch structType {
	case "collaboration":
		collaboration := p.parseCollaboration(element)
		for key, value := range collaboration {
			result[key] = value
		}
	case "participant":
		participant := p.parseParticipant(element)
		for key, value := range participant {
			result[key] = value
		}
	case "messageFlow":
		messageFlow := p.parseMessageFlow(element)
		for key, value := range messageFlow {
			result[key] = value
		}
	case "conversation", "callConversation", "subConversation":
		conversation := p.parseConversation(element)
		for key, value := range conversation {
			result[key] = value
		}
	case "conversationNode", "callActivity":
		conversationNode := p.parseConversationNode(element)
		for key, value := range conversationNode {
			result[key] = value
		}
	case "participantAssociation":
		association := p.parseParticipantAssociation(element)
		for key, value := range association {
			result[key] = value
		}
	case "participantMultiplicity":
		multiplicity := p.parseParticipantMultiplicity(element)
		for key, value := range multiplicity {
			result[key] = value
		}
	case "choreography", "globalChoreographyTask":
		choreography := p.parseChoreography(element)
		for key, value := range choreography {
			result[key] = value
		}
	case "textAnnotation", "association", "group":
		artifact := p.parseArtifact(element)
		for key, value := range artifact {
			result[key] = value
		}
	}

	// Parse child elements
	// Парсинг дочерних элементов
	children := make([]map[string]interface{}, 0)
	for _, child := range element.Children {
		childResult := make(map[string]interface{})
		childResult["type"] = child.XMLName.Local
		childResult["namespace"] = child.XMLName.Space

		// Parse attributes for child elements
		// Парсинг атрибутов для дочерних элементов
		childAttributes := make(map[string]string)
		for _, attr := range child.Attributes {
			childAttributes[attr.Name.Local] = attr.Value
		}
		childResult["attributes"] = childAttributes

		// Parse text content
		// Парсинг текстового содержимого
		if child.Text != "" {
			childResult["text"] = child.Text
		}

		children = append(children, childResult)
	}
	if len(children) > 0 {
		result["children"] = children
	}

	// Parse text content at the root level
	// Парсинг текстового содержимого на корневом уровне
	if element.Text != "" {
		result["text"] = element.Text
	}

	// Parse extension elements
	// Парсинг элементов расширения
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

	logger.Info("Successfully parsed structure element",
		logger.String("element_type", structType),
		logger.String("element_id", getStringValueStructure(result["id"])),
		logger.String("process_id", context.ProcessID))

	return result, nil
}

// parseCollaboration parses bpmn:collaboration element
// Парсинг элемента bpmn:collaboration
func (p *StructureParser) parseCollaboration(element *XMLElement) map[string]interface{} {
	collaboration := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			collaboration["id"] = attr.Value
		case "name":
			collaboration["name"] = attr.Value
		case "isClosed":
			if closed, err := strconv.ParseBool(attr.Value); err == nil {
				collaboration["is_closed"] = closed
			} else {
				collaboration["is_closed"] = attr.Value
			}
		}
	}

	// Parse participants
	// Парсинг участников
	participants := make([]map[string]interface{}, 0)
	messageFlows := make([]map[string]interface{}, 0)
	artifacts := make([]map[string]interface{}, 0)
	conversations := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "participant":
			participant := p.parseParticipant(child)
			participants = append(participants, participant)
		case "messageFlow":
			messageFlow := p.parseMessageFlow(child)
			messageFlows = append(messageFlows, messageFlow)
		case "textAnnotation", "association", "group":
			artifact := p.parseArtifact(child)
			artifacts = append(artifacts, artifact)
		case "conversation", "callConversation", "subConversation":
			conversation := p.parseConversation(child)
			conversations = append(conversations, conversation)
		}
	}

	if len(participants) > 0 {
		collaboration["participants"] = participants
	}
	if len(messageFlows) > 0 {
		collaboration["message_flows"] = messageFlows
	}
	if len(artifacts) > 0 {
		collaboration["artifacts"] = artifacts
	}
	if len(conversations) > 0 {
		collaboration["conversations"] = conversations
	}

	return collaboration
}

// parseParticipant parses bpmn:participant element
// Парсинг элемента bpmn:participant
func (p *StructureParser) parseParticipant(element *XMLElement) map[string]interface{} {
	participant := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			participant["id"] = attr.Value
		case "name":
			participant["name"] = attr.Value
		case "processRef":
			participant["process_ref"] = attr.Value
		case "multiplicity":
			if multiplicity, err := strconv.Atoi(attr.Value); err == nil {
				participant["multiplicity"] = multiplicity
			} else {
				participant["multiplicity"] = attr.Value
			}
		}
	}

	// Parse participant multiplicity elements
	// Парсинг элементов multiplicity участника
	multiplicities := make([]map[string]interface{}, 0)
	associations := make([]map[string]interface{}, 0)

	for _, child := range element.Children {
		switch child.XMLName.Local {
		case "participantMultiplicity":
			multiplicity := p.parseParticipantMultiplicity(child)
			multiplicities = append(multiplicities, multiplicity)
		case "participantAssociation":
			association := p.parseParticipantAssociation(child)
			associations = append(associations, association)
		}
	}

	if len(multiplicities) > 0 {
		participant["multiplicity_elements"] = multiplicities
	}
	if len(associations) > 0 {
		participant["associations"] = associations
	}

	return participant
}

// parseMessageFlow parses bpmn:messageFlow element
// Парсинг элемента bpmn:messageFlow
func (p *StructureParser) parseMessageFlow(element *XMLElement) map[string]interface{} {
	messageFlow := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			messageFlow["id"] = attr.Value
		case "name":
			messageFlow["name"] = attr.Value
		case "sourceRef":
			messageFlow["source_ref"] = attr.Value
		case "targetRef":
			messageFlow["target_ref"] = attr.Value
		case "messageRef":
			messageFlow["message_ref"] = attr.Value
		}
	}

	return messageFlow
}
