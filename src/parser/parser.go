/*
This file is part of the AtomBPMN (R) project.
Copyright (c) 2025 Matreska Market LLC (ООО «Matreska Market»).
Authors: Matreska Team.

This project is dual-licensed under AGPL-3.0 and AtomBPMN Commercial License.
*/

package parser

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"atom-engine/src/core/logger"
	"atom-engine/src/core/models"
)

// BPMNParser main BPMN parser coordinator
// Главный координатор BPMN парсера
type BPMNParser struct {
	elementParsers map[string]ElementParser
	processData    *models.BPMNProcess
}

// ElementParser interface for all element parsers
// Интерфейс для всех парсеров элементов
type ElementParser interface {
	Parse(element *XMLElement, context *ParseContext) (map[string]interface{}, error)
	GetElementType() string
}

// ParseContext provides context during parsing
// Контекст парсинга
type ParseContext struct {
	ProcessID      string
	NamespaceMap   map[string]string
	ElementCounts  map[string]int
	AllElements    map[string]interface{}
	CurrentElement string
}

// XMLElement generic XML element representation
// Представление общего XML элемента
type XMLElement struct {
	XMLName    xml.Name
	Attributes []xml.Attr    `xml:",any,attr"`
	Content    []byte        `xml:",innerxml"`
	Children   []*XMLElement `xml:",any"`
	Text       string        `xml:",chardata"`
}

// NewBPMNParser creates new BPMN parser
// Создает новый BPMN парсер
func NewBPMNParser() *BPMNParser {
	parser := &BPMNParser{
		elementParsers: make(map[string]ElementParser),
	}

	// Register all element parsers
	// Регистрация всех парсеров элементов
	parser.registerElementParsers()

	return parser
}

// registerElementParsers registers all available element parsers
// Регистрирует все доступные парсеры элементов
func (p *BPMNParser) registerElementParsers() {
	// Import elements package
	// Импорт пакета elements

	// Register core elements parsers
	// Регистрация парсеров основных элементов
	definitionsParser := NewDefinitionsParser()
	processParser := NewProcessParser()
	eventParser := NewEventParser()
	taskParser := NewTaskParser()
	gatewayParser := NewGatewayParser()
	flowParser := NewFlowParser()

	// Register definitions parser
	// Регистрация парсера definitions
	p.elementParsers["definitions"] = definitionsParser

	// Register process parser
	// Регистрация парсера process
	p.elementParsers["process"] = processParser

	// Register event parsers for all event types
	// Регистрация парсеров событий для всех типов событий
	eventTypes := []string{
		"startEvent", "endEvent", "intermediateCatchEvent", "intermediateThrowEvent", "boundaryEvent",
	}
	for _, eventType := range eventTypes {
		p.elementParsers[eventType] = eventParser
	}

	// Register task parsers for all task types
	// Регистрация парсеров задач для всех типов задач
	taskTypes := []string{
		"task", "userTask", "serviceTask", "scriptTask", "sendTask", "receiveTask",
		"manualTask", "businessRuleTask", "callActivity", "subProcess",
	}
	for _, taskType := range taskTypes {
		p.elementParsers[taskType] = taskParser
	}

	// Register gateway parsers for all gateway types
	// Регистрация парсеров шлюзов для всех типов шлюзов
	gatewayTypes := []string{
		"exclusiveGateway", "parallelGateway", "inclusiveGateway",
		"complexGateway", "eventBasedGateway",
	}
	for _, gatewayType := range gatewayTypes {
		p.elementParsers[gatewayType] = gatewayParser
	}

	// Register flow parsers for all flow types
	// Регистрация парсеров потоков для всех типов потоков
	flowTypes := []string{
		"sequenceFlow", "messageFlow", "association",
	}
	for _, flowType := range flowTypes {
		p.elementParsers[flowType] = flowParser
	}

	// Register new specialized parsers
	// Регистрация новых специализированных парсеров

	// Event definition parser for all event definition types
	// Парсер определений событий для всех типов определений событий
	eventDefParser := NewEventDefinitionParser()
	eventDefinitionTypes := []string{
		"timerEventDefinition", "messageEventDefinition", "signalEventDefinition",
		"conditionalEventDefinition", "errorEventDefinition", "escalationEventDefinition",
		"compensateEventDefinition", "linkEventDefinition", "terminateEventDefinition",
		"cancelEventDefinition",
	}
	for _, eventDefType := range eventDefinitionTypes {
		p.elementParsers[eventDefType] = eventDefParser
	}

	// Metadata parser for all zeebe extension and metadata elements
	// Парсер метаданных для всех элементов расширения zeebe и метаданных
	metadataParser := NewMetadataParser()
	metadataTypes := []string{
		"properties", "property", "taskDefinition", "subscription", "formDefinition",
		"calledElement", "ioMapping", "input", "output", "header", "script",
		"assignmentDefinition", "userTask",
	}
	for _, metadataType := range metadataTypes {
		p.elementParsers[metadataType] = metadataParser
	}

	// Reference parser for error, signal, message and escalation definitions
	// Парсер ссылок для определений error, signal, message и escalation
	referenceParser := NewReferenceParser()
	referenceTypes := []string{
		"error", "signal", "message", "escalation",
	}
	for _, referenceType := range referenceTypes {
		p.elementParsers[referenceType] = referenceParser
	}

	// Structure parser for collaboration and choreography elements
	// Парсер структуры для элементов collaboration и choreography
	structureParser := NewStructureParser()
	structureTypes := []string{
		"collaboration", "participant", "messageFlow", "conversation",
		"conversationNode", "participantAssociation", "participantMultiplicity",
		"choreography", "textAnnotation", "group",
	}
	for _, structureType := range structureTypes {
		p.elementParsers[structureType] = structureParser
	}

	// Additional time-related elements for timer parsing
	// Дополнительные элементы времени для парсинга таймеров
	timeDurationTypes := []string{
		"timeDuration", "timeDate", "timeCycle",
	}
	for _, timeDurationType := range timeDurationTypes {
		p.elementParsers[timeDurationType] = eventDefParser
	}
}

// ParseBPMNContent parses BPMN XML content and returns JSON
// Парсит содержимое BPMN XML и возвращает JSON
func (p *BPMNParser) ParseBPMNContent(bpmnContent, processID string, force bool) (*models.BPMNProcess, error) {
	logger.Info("Starting BPMN content parsing",
		logger.String("content_length", fmt.Sprintf("%d", len(bpmnContent))),
		logger.String("process_id", processID),
		logger.Bool("force", force))

	content := []byte(bpmnContent)

	// Parse XML structure
	xmlRoot, err := p.parseXMLStructure(content)
	if err != nil {
		logger.Error("Failed to parse XML structure",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse XML structure: %w", err)
	}

	// Create process data model
	if processID == "" {
		processID = p.extractProcessIDFromXML(xmlRoot)
		logger.Info("Extracted process ID from XML", logger.String("process_id", processID))
	}

	processName := p.extractProcessNameFromXML(xmlRoot)
	processVersion := p.extractProcessVersionFromXML(xmlRoot)
	bpmnProcess := models.NewBPMNProcess(processID, processName)
	bpmnProcess.BPMNID = models.GenerateBPMNID()
	bpmnProcess.OriginalFile = "uploaded_content.bpmn" // No file path for content
	bpmnProcess.ContentHash = models.GenerateContentHash(content)
	bpmnProcess.ProcessVersion = processVersion

	logger.Info("Created BPMN process model",
		logger.String("bpmn_id", bpmnProcess.BPMNID),
		logger.String("process_id", processID),
		logger.String("process_name", processName),
		logger.Int("process_version", processVersion))

	// Create parse context
	context := &ParseContext{
		ProcessID:     processID,
		NamespaceMap:  p.extractNamespaces(xmlRoot),
		ElementCounts: make(map[string]int),
		AllElements:   make(map[string]interface{}),
	}

	logger.Info("Starting element parsing",
		logger.Int("namespace_count", len(context.NamespaceMap)))

	// Parse all elements
	err = p.parseAllElements(xmlRoot, context, bpmnProcess)
	if err != nil {
		logger.Error("Failed to parse elements",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse elements: %w", err)
	}

	// Set final data
	bpmnProcess.ElementCounts = context.ElementCounts
	bpmnProcess.ParsedAt = time.Now()

	// Calculate total elements
	totalElements := 0
	for _, count := range context.ElementCounts {
		totalElements += count
	}

	logger.Info("Successfully parsed BPMN content",
		logger.String("bpmn_id", bpmnProcess.BPMNID),
		logger.String("process_id", bpmnProcess.ProcessID),
		logger.Int("total_elements", totalElements))

	return bpmnProcess, nil
}

// ParseBPMNFile parses BPMN XML file and returns JSON
// Парсит BPMN XML файл и возвращает JSON
func (p *BPMNParser) ParseBPMNFile(filePath, processID string, force bool) (*models.BPMNProcess, error) {
	logger.Info("Starting BPMN file parsing",
		logger.String("file", filePath),
		logger.String("process_id", processID),
		logger.Bool("force", force))

	// Read file content
	// Чтение содержимого файла
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read BPMN file",
			logger.String("file", filePath),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to read BPMN file: %w", err)
	}

	logger.Info("Successfully read BPMN file",
		logger.String("file", filePath),
		logger.Int("size_bytes", len(content)))

	// Parse XML structure
	// Парсинг XML структуры
	xmlRoot, err := p.parseXMLStructure(content)
	if err != nil {
		logger.Error("Failed to parse XML structure",
			logger.String("file", filePath),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse XML structure: %w", err)
	}

	// Create process data model
	// Создание модели данных процесса
	if processID == "" {
		processID = p.extractProcessIDFromXML(xmlRoot)
		logger.Info("Extracted process ID from XML", logger.String("process_id", processID))
	}

	processName := p.extractProcessNameFromXML(xmlRoot)
	processVersion := p.extractProcessVersionFromXML(xmlRoot)
	bpmnProcess := models.NewBPMNProcess(processID, processName)
	bpmnProcess.BPMNID = models.GenerateBPMNID()
	bpmnProcess.OriginalFile = filepath.Base(filePath)
	bpmnProcess.ContentHash = models.GenerateContentHash(content)
	bpmnProcess.ProcessVersion = processVersion

	logger.Info("Created BPMN process model",
		logger.String("bpmn_id", bpmnProcess.BPMNID),
		logger.String("process_id", processID),
		logger.String("process_name", processName),
		logger.Int("extracted_process_version", processVersion))

	// Create parse context
	// Создание контекста парсинга
	context := &ParseContext{
		ProcessID:     processID,
		NamespaceMap:  p.extractNamespaces(xmlRoot),
		ElementCounts: make(map[string]int),
		AllElements:   make(map[string]interface{}),
	}

	logger.Info("Starting element parsing",
		logger.Int("namespace_count", len(context.NamespaceMap)))

	// Parse all elements
	// Парсинг всех элементов
	err = p.parseAllElements(xmlRoot, context, bpmnProcess)
	if err != nil {
		logger.Error("Failed to parse elements",
			logger.String("file", filePath),
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to parse elements: %w", err)
	}

	// Update element counts in process
	// Обновление счетчиков элементов в процессе
	for elementType, count := range context.ElementCounts {
		bpmnProcess.UpdateElementCount(elementType, count)
	}

	// Extract is_executable from parsed process element
	// Извлечение is_executable из спарсенного элемента process
	for elementID, element := range bpmnProcess.Elements {
		if elementMap, ok := element.(map[string]interface{}); ok {
			if elementType, exists := elementMap["type"]; exists && elementType == "process" {
				if isExecutable, exists := elementMap["is_executable"]; exists {
					if executable, ok := isExecutable.(bool); ok {
						bpmnProcess.IsExecutable = executable
						logger.Debug("Set is_executable from process element",
							logger.String("element_id", elementID),
							logger.Bool("is_executable", executable))
						break
					}
				}
			}
		}
	}

	logger.Info("Successfully completed BPMN parsing",
		logger.String("bpmn_id", bpmnProcess.BPMNID),
		logger.Int("total_elements", bpmnProcess.GetTotalElements()),
		logger.Any("element_counts", context.ElementCounts))

	p.processData = bpmnProcess
	return bpmnProcess, nil
}

// parseXMLStructure parses XML into generic structure
// Парсинг XML в общую структуру
func (p *BPMNParser) parseXMLStructure(content []byte) (*XMLElement, error) {
	var root XMLElement
	err := xml.Unmarshal(content, &root)
	if err != nil {
		return nil, fmt.Errorf("XML unmarshal failed: %w", err)
	}
	return &root, nil
}

// parseAllElements recursively parses all XML elements
// Рекурсивно парсит все XML элементы
func (p *BPMNParser) parseAllElements(
	element *XMLElement,
	context *ParseContext,
	bpmnProcess *models.BPMNProcess,
) error {
	// Skip empty elements and text-only elements
	// Пропускаем пустые элементы и элементы только с текстом
	if element.XMLName.Local == "" {
		return nil
	}

	// Get element type
	// Получение типа элемента
	elementType := element.XMLName.Local
	context.CurrentElement = elementType

	// Skip diagram elements (they are not part of process logic)
	// Пропускаем элементы диаграммы (они не часть логики процесса)
	if p.isDiagramElement(elementType) {
		logger.Debug("Skipping diagram element",
			logger.String("element_type", elementType))
		return nil
	}

	// Count element only if it's not a diagram element
	// Подсчет элемента только если это не элемент диаграммы
	context.ElementCounts[elementType]++

	// Find appropriate parser
	// Поиск подходящего парсера
	if parser, exists := p.elementParsers[elementType]; exists {
		// Parse element with specific parser
		// Парсинг элемента с определенным парсером
		parsedData, err := parser.Parse(element, context)
		if err != nil {
			elementID := p.getElementID(element)
			logger.Info("Failed to parse element with specific parser, falling back to generic",
				logger.String("element_type", elementType),
				logger.String("element_id", elementID),
				logger.String("error", err.Error()))
			fmt.Printf("⚠️  Failed to parse: '%s' (ID: %s) - %s\n", elementType, elementID, err.Error())

			// Fall back to generic parsing
			// Откат к общему парсингу
			parsedData = p.parseGenericElement(element)
			if elementID != "" {
				context.AllElements[elementID] = parsedData
				bpmnProcess.AddElement(elementID, parsedData)
			}
		} else {
			// Store parsed data
			// Сохранение спарсенных данных
			elementID := p.getElementID(element)
			if elementID != "" {
				context.AllElements[elementID] = parsedData
				bpmnProcess.AddElement(elementID, parsedData)
			}
		}
	} else {
		// Generic parsing for unknown elements
		// Общий парсинг для неизвестных элементов
		elementID := p.getElementID(element)

		// Only log unknown business elements, not diagram elements
		// Логируем только неизвестные бизнес-элементы, не элементы диаграммы
		if !p.isDiagramElement(elementType) && !p.isKnownMetadataElement(elementType) {
			logger.Info("Parsing unknown business element with generic parser",
				logger.String("element_type", elementType),
				logger.String("element_id", elementID),
				logger.String("namespace", element.XMLName.Space))
			fmt.Printf("ℹ️  Using generic parser: '%s' (ID: %s) - no specific parser available\n", elementType, elementID)
		} else if p.isDiagramElement(elementType) {
			fmt.Printf("🎨 Diagram element: '%s' - processed for visualization\n", elementType)
		}

		parsedData := p.parseGenericElement(element)
		if elementID != "" {
			context.AllElements[elementID] = parsedData
			bpmnProcess.AddElement(elementID, parsedData)
		}
	}

	// Parse child elements recursively
	// Рекурсивный парсинг дочерних элементов
	for _, child := range element.Children {
		err := p.parseAllElements(child, context, bpmnProcess)
		if err != nil {
			return err
		}
	}

	return nil
}
