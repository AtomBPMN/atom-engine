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

// parseConversation parses bpmn:conversation element
// Парсинг элемента bpmn:conversation
func (p *StructureParser) parseConversation(element *XMLElement) map[string]interface{} {
	conversation := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			conversation["id"] = attr.Value
		case "name":
			conversation["name"] = attr.Value
		case "isClosed":
			if closed, err := strconv.ParseBool(attr.Value); err == nil {
				conversation["is_closed"] = closed
			} else {
				conversation["is_closed"] = attr.Value
			}
		}
	}

	conversation["conversation_type"] = element.XMLName.Local

	return conversation
}

// parseConversationNode parses conversation node elements
// Парсинг элементов conversation node
func (p *StructureParser) parseConversationNode(element *XMLElement) map[string]interface{} {
	conversationNode := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			conversationNode["id"] = attr.Value
		case "name":
			conversationNode["name"] = attr.Value
		}
	}

	return conversationNode
}

// parseParticipantAssociation parses participantAssociation element
// Парсинг элемента participantAssociation
func (p *StructureParser) parseParticipantAssociation(element *XMLElement) map[string]interface{} {
	association := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			association["id"] = attr.Value
		case "innerParticipantRef":
			association["inner_participant_ref"] = attr.Value
		case "outerParticipantRef":
			association["outer_participant_ref"] = attr.Value
		}
	}

	return association
}

// parseParticipantMultiplicity parses participantMultiplicity element
// Парсинг элемента participantMultiplicity
func (p *StructureParser) parseParticipantMultiplicity(element *XMLElement) map[string]interface{} {
	multiplicity := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "minimum":
			if min, err := strconv.Atoi(attr.Value); err == nil {
				multiplicity["minimum"] = min
			} else {
				multiplicity["minimum"] = attr.Value
			}
		case "maximum":
			if max, err := strconv.Atoi(attr.Value); err == nil {
				multiplicity["maximum"] = max
			} else {
				multiplicity["maximum"] = attr.Value
			}
		}
	}

	return multiplicity
}

// parseChoreography parses choreography element
// Парсинг элемента choreography
func (p *StructureParser) parseChoreography(element *XMLElement) map[string]interface{} {
	choreography := make(map[string]interface{})

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			choreography["id"] = attr.Value
		case "name":
			choreography["name"] = attr.Value
		case "isClosed":
			if closed, err := strconv.ParseBool(attr.Value); err == nil {
				choreography["is_closed"] = closed
			} else {
				choreography["is_closed"] = attr.Value
			}
		}
	}

	return choreography
}

// parseArtifact parses artifact elements (textAnnotation, association, group)
// Парсинг элементов артефактов (textAnnotation, association, group)
func (p *StructureParser) parseArtifact(element *XMLElement) map[string]interface{} {
	artifact := make(map[string]interface{})

	artifact["artifact_type"] = element.XMLName.Local

	for _, attr := range element.Attributes {
		switch attr.Name.Local {
		case "id":
			artifact["id"] = attr.Value
		case "name":
			artifact["name"] = attr.Value
		case "sourceRef":
			artifact["source_ref"] = attr.Value
		case "targetRef":
			artifact["target_ref"] = attr.Value
		case "associationDirection":
			artifact["association_direction"] = attr.Value
		}
	}

	// Parse text content for text annotations
	// Парсинг текстового содержимого для текстовых аннотаций
	if element.XMLName.Local == "textAnnotation" && element.Text != "" {
		artifact["text"] = element.Text
	}

	return artifact
}

// parseExtensionElements parses extension elements for structure
// Парсинг элементов расширения для структуры
func (p *StructureParser) parseExtensionElements(element *XMLElement) map[string]interface{} {
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

// GetElementType returns element type for structure
// Возвращает тип элемента для структуры
func (p *StructureParser) GetElementType() string {
	return "structure"
}

// getStringValueStructure safely gets string value from interface{}
// Безопасно получает строковое значение из interface{}
func getStringValueStructure(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// getIntValueStructure safely gets int value from interface{}
// Безопасно получает целое значение из interface{}
func getIntValueStructure(value interface{}) int {
	if value == nil {
		return 0
	}
	if intVal, ok := value.(int); ok {
		return intVal
	}
	return 0
}
