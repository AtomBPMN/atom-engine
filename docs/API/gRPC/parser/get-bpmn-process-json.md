# GetBPMNProcessJSON

## Описание
Получает JSON представление BPMN процесса - структурированные данные процесса, готовые для выполнения движком.

## Синтаксис
```protobuf
rpc GetBPMNProcessJSON(GetBPMNProcessJSONRequest) returns (GetBPMNProcessJSONResponse);
```

## Package
```protobuf
package parser;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `parser`, `read` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### GetBPMNProcessJSONRequest
```protobuf
message GetBPMNProcessJSONRequest {
  string process_key = 1;        // Уникальный ключ процесса
  bool pretty_format = 2;        // Форматированный JSON (с отступами)
  bool include_metadata = 3;     // Включить метаданные парсинга
}
```

#### Поля:
- **process_key** (string, required): Уникальный ключ процесса
- **pretty_format** (bool, optional): Если `true`, JSON форматируется с отступами для читаемости
- **include_metadata** (bool, optional): Включить дополнительные метаданные парсинга

## Параметры ответа

### GetBPMNProcessJSONResponse
```protobuf
message GetBPMNProcessJSONResponse {
  bool success = 1;              // Статус успешности
  string message = 2;            // Сообщение о результате
  string json_data = 3;          // JSON данные процесса
  int32 json_size = 4;           // Размер JSON в байтах
  string content_hash = 5;       // Хеш содержимого
}
```

#### Поля ответа:
- **success** (bool): `true` если данные получены успешно
- **message** (string): Описание результата операции
- **json_data** (string): JSON представление процесса
- **json_size** (int32): Размер JSON данных в байтах
- **content_hash** (string): MD5 хеш содержимого для проверки целостности

## Примеры использования

### Go
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    
    pb "atom-engine/proto/parser/parserpb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewParserServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    processKey := "atom-7-1k2-PVn4Y9j-CF5M"
    
    // Получение JSON данных процесса
    response, err := client.GetBPMNProcessJSON(ctx, &pb.GetBPMNProcessJSONRequest{
        ProcessKey:      processKey,
        PrettyFormat:    true,
        IncludeMetadata: true,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("JSON размер: %d байт\n", response.JsonSize)
        fmt.Printf("Хеш содержимого: %s\n", response.ContentHash)
        
        // Сохранение в файл
        filename := fmt.Sprintf("process_%s.json", processKey)
        err := os.WriteFile(filename, []byte(response.JsonData), 0644)
        if err != nil {
            log.Printf("Ошибка сохранения: %v", err)
        } else {
            fmt.Printf("JSON сохранен в: %s\n", filename)
        }
        
        // Парсинг и анализ JSON
        analyzeProcessJSON(response.JsonData)
        
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

func analyzeProcessJSON(jsonData string) {
    var process map[string]interface{}
    
    if err := json.Unmarshal([]byte(jsonData), &process); err != nil {
        log.Printf("Ошибка парсинга JSON: %v", err)
        return
    }
    
    fmt.Println("\n=== Анализ процесса ===")
    
    // Основная информация
    if processId, ok := process["processId"].(string); ok {
        fmt.Printf("Process ID: %s\n", processId)
    }
    
    if processName, ok := process["processName"].(string); ok {
        fmt.Printf("Process Name: %s\n", processName)
    }
    
    // Анализ элементов
    if elements, ok := process["elements"].(map[string]interface{}); ok {
        fmt.Printf("Элементов в процессе: %d\n", len(elements))
        
        // Группировка по типам
        elementTypes := make(map[string]int)
        for _, element := range elements {
            if elementMap, ok := element.(map[string]interface{}); ok {
                if elementType, ok := elementMap["type"].(string); ok {
                    elementTypes[elementType]++
                }
            }
        }
        
        fmt.Println("Типы элементов:")
        for elementType, count := range elementTypes {
            fmt.Printf("  %s: %d\n", elementType, count)
        }
    }
    
    // Анализ связей
    if flows, ok := process["flows"].(map[string]interface{}); ok {
        fmt.Printf("Связей (flows): %d\n", len(flows))
    }
}

// Сравнение двух версий процесса
func compareProcessVersions(client pb.ParserServiceClient, ctx context.Context, key1, key2 string) {
    // Получение первой версии
    response1, err := client.GetBPMNProcessJSON(ctx, &pb.GetBPMNProcessJSONRequest{
        ProcessKey:      key1,
        PrettyFormat:    false,
        IncludeMetadata: false,
    })
    
    if err != nil || !response1.Success {
        log.Printf("Ошибка получения процесса %s: %v", key1, err)
        return
    }
    
    // Получение второй версии
    response2, err := client.GetBPMNProcessJSON(ctx, &pb.GetBPMNProcessJSONRequest{
        ProcessKey:      key2,
        PrettyFormat:    false,
        IncludeMetadata: false,
    })
    
    if err != nil || !response2.Success {
        log.Printf("Ошибка получения процесса %s: %v", key2, err)
        return
    }
    
    fmt.Printf("Сравнение процессов:\n")
    fmt.Printf("  %s: %d байт (хеш: %s)\n", key1, response1.JsonSize, response1.ContentHash)
    fmt.Printf("  %s: %d байт (хеш: %s)\n", key2, response2.JsonSize, response2.ContentHash)
    
    if response1.ContentHash == response2.ContentHash {
        fmt.Println("✅ Процессы идентичны")
    } else {
        fmt.Println("❌ Процессы различаются")
        
        // Детальное сравнение
        var process1, process2 map[string]interface{}
        json.Unmarshal([]byte(response1.JsonData), &process1)
        json.Unmarshal([]byte(response2.JsonData), &process2)
        
        // Сравнение количества элементов
        elements1, _ := process1["elements"].(map[string]interface{})
        elements2, _ := process2["elements"].(map[string]interface{})
        
        fmt.Printf("  Элементов: %d vs %d\n", len(elements1), len(elements2))
    }
}
```

### Python
```python
import grpc
import json
import hashlib
from datetime import datetime

import parser_pb2
import parser_pb2_grpc

def get_bpmn_process_json(process_key, pretty=True, include_metadata=True):
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = parser_pb2.GetBPMNProcessJSONRequest(
        process_key=process_key,
        pretty_format=pretty,
        include_metadata=include_metadata
    )
    
    try:
        response = stub.GetBPMNProcessJSON(request, metadata=metadata)
        
        if response.success:
            print(f"JSON размер: {response.json_size} байт")
            print(f"Хеш содержимого: {response.content_hash}")
            
            # Парсинг JSON
            process_data = json.loads(response.json_data)
            
            return {
                'data': process_data,
                'size': response.json_size,
                'hash': response.content_hash,
                'raw_json': response.json_data
            }
            
        else:
            print(f"Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def analyze_process_structure(process_data):
    """Детальный анализ структуры процесса"""
    data = process_data['data']
    
    print("\n=== Анализ структуры процесса ===")
    
    # Основная информация
    print(f"Process ID: {data.get('processId', 'N/A')}")
    print(f"Process Name: {data.get('processName', 'N/A')}")
    print(f"Version: {data.get('version', 'N/A')}")
    
    # Элементы
    elements = data.get('elements', {})
    flows = data.get('flows', {})
    
    print(f"\nСтатистика:")
    print(f"  Элементов: {len(elements)}")
    print(f"  Связей: {len(flows)}")
    
    # Анализ элементов по типам
    element_types = {}
    important_elements = {
        'startEvents': [],
        'endEvents': [],
        'serviceTasks': [],
        'userTasks': [],
        'gateways': []
    }
    
    for element_id, element in elements.items():
        element_type = element.get('type', 'unknown')
        element_types[element_type] = element_types.get(element_type, 0) + 1
        
        # Сбор важных элементов
        if element_type == 'startEvent':
            important_elements['startEvents'].append(element_id)
        elif element_type == 'endEvent':
            important_elements['endEvents'].append(element_id)
        elif element_type == 'serviceTask':
            important_elements['serviceTasks'].append(element_id)
        elif element_type == 'userTask':
            important_elements['userTasks'].append(element_id)
        elif 'gateway' in element_type.lower():
            important_elements['gateways'].append(element_id)
    
    print("\nТипы элементов:")
    for element_type, count in sorted(element_types.items()):
        print(f"  {element_type}: {count}")
    
    # Важные элементы
    print("\nВажные элементы:")
    for category, element_list in important_elements.items():
        if element_list:
            print(f"  {category}: {len(element_list)}")
            for element_id in element_list[:3]:  # Показываем первые 3
                element = elements[element_id]
                name = element.get('name', '<без имени>')
                print(f"    - {element_id}: {name}")
            if len(element_list) > 3:
                print(f"    ... и еще {len(element_list) - 3}")
    
    # Проверка целостности
    print("\n=== Проверка целостности ===")
    
    # Проверка start/end events
    start_count = len(important_elements['startEvents'])
    end_count = len(important_elements['endEvents'])
    
    if start_count == 0:
        print("❌ Отсутствуют startEvent")
    elif start_count == 1:
        print("✅ Один startEvent")
    else:
        print(f"⚠️  Множественные startEvent: {start_count}")
    
    if end_count == 0:
        print("❌ Отсутствуют endEvent")
    else:
        print(f"✅ EndEvent найдены: {end_count}")
    
    # Проверка связей
    orphaned_elements = []
    for element_id, element in elements.items():
        element_type = element.get('type', '')
        
        # Пропускаем startEvent (у них нет входящих связей)
        if element_type == 'startEvent':
            continue
            
        # Ищем входящие связи
        has_incoming = any(
            flow.get('targetRef') == element_id 
            for flow in flows.values()
        )
        
        if not has_incoming:
            orphaned_elements.append(element_id)
    
    if orphaned_elements:
        print(f"⚠️  Элементы без входящих связей: {len(orphaned_elements)}")
        for element_id in orphaned_elements[:3]:
            print(f"    - {element_id}")
    else:
        print("✅ Все элементы связаны")

def save_process_backup(process_key):
    """Создание резервной копии процесса"""
    result = get_bpmn_process_json(process_key, pretty=True, include_metadata=True)
    
    if not result:
        return None
    
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    filename = f"backup_{process_key}_{timestamp}.json"
    
    try:
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(result['raw_json'])
        
        print(f"Резервная копия сохранена: {filename}")
        print(f"Размер: {result['size']} байт")
        print(f"Хеш: {result['hash']}")
        
        return filename
        
    except Exception as e:
        print(f"Ошибка сохранения: {e}")
        return None

def validate_process_integrity(process_key):
    """Валидация целостности процесса"""
    result = get_bpmn_process_json(process_key, pretty=False, include_metadata=False)
    
    if not result:
        return False
    
    # Проверка хеша
    calculated_hash = hashlib.md5(result['raw_json'].encode()).hexdigest()
    
    if calculated_hash != result['hash']:
        print("❌ Ошибка целостности: хеш не совпадает")
        return False
    
    print("✅ Целостность данных подтверждена")
    
    # Структурная валидация
    try:
        data = result['data']
        
        # Обязательные поля
        required_fields = ['processId', 'elements', 'flows']
        missing_fields = [field for field in required_fields if field not in data]
        
        if missing_fields:
            print(f"❌ Отсутствуют обязательные поля: {missing_fields}")
            return False
        
        # Валидация элементов
        elements = data['elements']
        flows = data['flows']
        
        # Проверка ссылок в flows
        for flow_id, flow in flows.items():
            source_ref = flow.get('sourceRef')
            target_ref = flow.get('targetRef')
            
            if source_ref and source_ref not in elements:
                print(f"❌ Связь {flow_id}: sourceRef '{source_ref}' не найден")
                return False
                
            if target_ref and target_ref not in elements:
                print(f"❌ Связь {flow_id}: targetRef '{target_ref}' не найден")
                return False
        
        print("✅ Структурная валидация пройдена")
        return True
        
    except Exception as e:
        print(f"❌ Ошибка валидации: {e}")
        return False

def extract_service_tasks_config(process_key):
    """Извлечение конфигурации service tasks"""
    result = get_bpmn_process_json(process_key, pretty=False, include_metadata=False)
    
    if not result:
        return {}
    
    data = result['data']
    elements = data.get('elements', {})
    
    service_tasks = {}
    
    for element_id, element in elements.items():
        if element.get('type') == 'serviceTask':
            task_config = {
                'id': element_id,
                'name': element.get('name', ''),
                'type': element.get('taskDefinition', {}).get('type', ''),
                'retries': element.get('taskDefinition', {}).get('retries', 3),
                'headers': element.get('taskHeaders', {}),
                'ioMapping': element.get('ioMapping', {})
            }
            
            service_tasks[element_id] = task_config
    
    print(f"Найдено Service Tasks: {len(service_tasks)}")
    for task_id, config in service_tasks.items():
        print(f"  - {task_id}: {config['name']} (тип: {config['type']})")
    
    return service_tasks

if __name__ == "__main__":
    process_key = "atom-7-1k2-PVn4Y9j-CF5M"
    
    # Получение и анализ
    result = get_bpmn_process_json(process_key)
    if result:
        analyze_process_structure(result)
        
        # Валидация
        validate_process_integrity(process_key)
        
        # Извлечение конфигурации задач
        extract_service_tasks_config(process_key)
        
        # Создание резервной копии
        save_process_backup(process_key)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const fs = require('fs').promises;
const crypto = require('crypto');

const PROTO_PATH = 'parser.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

async function getBPMNProcessJSON(processKey, options = {}) {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    const {
        prettyFormat = true,
        includeMetadata = true
    } = options;
    
    return new Promise((resolve, reject) => {
        const request = {
            process_key: processKey,
            pretty_format: prettyFormat,
            include_metadata: includeMetadata
        };
        
        client.getBPMNProcessJSON(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            try {
                const processData = JSON.parse(response.json_data);
                
                resolve({
                    data: processData,
                    size: response.json_size,
                    hash: response.content_hash,
                    rawJson: response.json_data
                });
            } catch (parseError) {
                reject(new Error(`JSON parse error: ${parseError.message}`));
            }
        });
    });
}

async function analyzeProcessComplexity(processKey) {
    try {
        const result = await getBPMNProcessJSON(processKey, { prettyFormat: false });
        const { data } = result;
        
        console.log('=== Анализ сложности процесса ===');
        console.log(`Process Key: ${processKey}`);
        console.log(`JSON размер: ${result.size} байт`);
        
        const elements = data.elements || {};
        const flows = data.flows || {};
        
        // Основные метрики
        const metrics = {
            totalElements: Object.keys(elements).length,
            totalFlows: Object.keys(flows).length,
            cyclomaticComplexity: calculateCyclomaticComplexity(elements, flows),
            maxDepth: calculateMaxDepth(elements, flows),
            parallelPaths: countParallelPaths(elements, flows)
        };
        
        console.log('\nМетрики сложности:');
        console.log(`  Элементов: ${metrics.totalElements}`);
        console.log(`  Связей: ${metrics.totalFlows}`);
        console.log(`  Цикломатическая сложность: ${metrics.cyclomaticComplexity}`);
        console.log(`  Максимальная глубина: ${metrics.maxDepth}`);
        console.log(`  Параллельных путей: ${metrics.parallelPaths}`);
        
        // Оценка сложности
        const complexityScore = calculateComplexityScore(metrics);
        console.log(`\nОценка сложности: ${complexityScore.score}/100 (${complexityScore.level})`);
        
        if (complexityScore.recommendations.length > 0) {
            console.log('\nРекомендации:');
            complexityScore.recommendations.forEach(rec => {
                console.log(`  - ${rec}`);
            });
        }
        
        return metrics;
        
    } catch (error) {
        console.error('Ошибка анализа сложности:', error.message);
        return null;
    }
}

function calculateCyclomaticComplexity(elements, flows) {
    // Упрощенный расчет цикломатической сложности для BPMN
    let gateways = 0;
    let decisions = 0;
    
    for (const element of Object.values(elements)) {
        const type = element.type;
        
        if (type.includes('Gateway')) {
            gateways++;
            
            // Эксклюзивные и инклюзивные шлюзы добавляют решения
            if (type === 'exclusiveGateway' || type === 'inclusiveGateway') {
                // Считаем исходящие потоки
                const outgoing = Object.values(flows).filter(
                    flow => flow.sourceRef === element.id
                ).length;
                
                if (outgoing > 1) {
                    decisions += outgoing - 1;
                }
            }
        }
    }
    
    // Базовая формула: M = E - N + 2P (для связного графа)
    // Упрощаем для BPMN: 1 + количество точек принятия решений
    return 1 + decisions;
}

function calculateMaxDepth(elements, flows) {
    // Поиск максимальной глубины от startEvent до endEvent
    const startEvents = Object.keys(elements).filter(
        id => elements[id].type === 'startEvent'
    );
    
    if (startEvents.length === 0) return 0;
    
    let maxDepth = 0;
    
    for (const startEvent of startEvents) {
        const depth = findMaxDepthFromElement(startEvent, elements, flows, new Set());
        maxDepth = Math.max(maxDepth, depth);
    }
    
    return maxDepth;
}

function findMaxDepthFromElement(elementId, elements, flows, visited) {
    if (visited.has(elementId)) return 0; // Предотвращение циклов
    
    visited.add(elementId);
    
    const element = elements[elementId];
    if (!element) return 0;
    
    // Если это endEvent, возвращаем 1
    if (element.type === 'endEvent') {
        visited.delete(elementId);
        return 1;
    }
    
    // Находим исходящие потоки
    const outgoingFlows = Object.values(flows).filter(
        flow => flow.sourceRef === elementId
    );
    
    if (outgoingFlows.length === 0) {
        visited.delete(elementId);
        return 1;
    }
    
    let maxChildDepth = 0;
    for (const flow of outgoingFlows) {
        const childDepth = findMaxDepthFromElement(flow.targetRef, elements, flows, visited);
        maxChildDepth = Math.max(maxChildDepth, childDepth);
    }
    
    visited.delete(elementId);
    return 1 + maxChildDepth;
}

function countParallelPaths(elements, flows) {
    // Подсчет параллельных шлюзов (fork)
    let parallelCount = 0;
    
    for (const element of Object.values(elements)) {
        if (element.type === 'parallelGateway') {
            // Проверяем, является ли это fork (разветвление)
            const outgoing = Object.values(flows).filter(
                flow => flow.sourceRef === element.id
            ).length;
            
            if (outgoing > 1) {
                parallelCount += outgoing - 1;
            }
        }
    }
    
    return parallelCount;
}

function calculateComplexityScore(metrics) {
    let score = 100;
    const recommendations = [];
    
    // Штрафы за сложность
    if (metrics.totalElements > 50) {
        score -= 20;
        recommendations.push('Процесс содержит много элементов - рассмотрите разделение на подпроцессы');
    }
    
    if (metrics.cyclomaticComplexity > 10) {
        score -= 30;
        recommendations.push('Высокая цикломатическая сложность - упростите логику принятия решений');
    }
    
    if (metrics.maxDepth > 15) {
        score -= 15;
        recommendations.push('Процесс слишком глубокий - рассмотрите использование подпроцессов');
    }
    
    if (metrics.parallelPaths > 5) {
        score -= 10;
        recommendations.push('Много параллельных путей - убедитесь в необходимости');
    }
    
    score = Math.max(0, score);
    
    let level;
    if (score >= 80) level = 'Низкая';
    else if (score >= 60) level = 'Средняя';
    else if (score >= 40) level = 'Высокая';
    else level = 'Очень высокая';
    
    return { score, level, recommendations };
}

async function exportProcessDocumentation(processKey) {
    try {
        const result = await getBPMNProcessJSON(processKey, { 
            prettyFormat: true, 
            includeMetadata: true 
        });
        
        const { data } = result;
        const elements = data.elements || {};
        const flows = data.flows || {};
        
        // Создание документации
        const doc = {
            metadata: {
                processKey: processKey,
                processId: data.processId,
                processName: data.processName,
                version: data.version,
                generatedAt: new Date().toISOString(),
                jsonSize: result.size,
                contentHash: result.hash
            },
            summary: {
                totalElements: Object.keys(elements).length,
                totalFlows: Object.keys(flows).length,
                elementTypes: {}
            },
            elements: {},
            flows: flows,
            rawProcess: data
        };
        
        // Анализ элементов
        for (const [elementId, element] of Object.entries(elements)) {
            const type = element.type;
            doc.summary.elementTypes[type] = (doc.summary.elementTypes[type] || 0) + 1;
            
            doc.elements[elementId] = {
                id: elementId,
                type: type,
                name: element.name || '',
                description: element.documentation || '',
                properties: element
            };
        }
        
        // Сохранение документации
        const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
        const filename = `process_documentation_${processKey}_${timestamp}.json`;
        
        await fs.writeFile(filename, JSON.stringify(doc, null, 2));
        console.log(`Документация экспортирована: ${filename}`);
        
        return doc;
        
    } catch (error) {
        console.error('Ошибка экспорта документации:', error.message);
        return null;
    }
}

async function validateProcessJSON(processKey) {
    try {
        const result = await getBPMNProcessJSON(processKey, { prettyFormat: false });
        
        console.log('=== Валидация JSON процесса ===');
        
        // Проверка хеша
        const calculatedHash = crypto.createHash('md5')
            .update(result.rawJson)
            .digest('hex');
        
        if (calculatedHash !== result.hash) {
            console.log('❌ Ошибка целостности: хеш не совпадает');
            return false;
        }
        console.log('✅ Целостность данных подтверждена');
        
        const { data } = result;
        const elements = data.elements || {};
        const flows = data.flows || {};
        
        // Структурные проверки
        const issues = [];
        
        // Проверка обязательных полей
        if (!data.processId) issues.push('Отсутствует processId');
        if (Object.keys(elements).length === 0) issues.push('Нет элементов в процессе');
        
        // Проверка startEvent
        const startEvents = Object.values(elements).filter(e => e.type === 'startEvent');
        if (startEvents.length === 0) {
            issues.push('Отсутствует startEvent');
        } else if (startEvents.length > 1) {
            issues.push(`Множественные startEvent: ${startEvents.length}`);
        }
        
        // Проверка endEvent
        const endEvents = Object.values(elements).filter(e => e.type === 'endEvent');
        if (endEvents.length === 0) {
            issues.push('Отсутствует endEvent');
        }
        
        // Проверка ссылок в flows
        for (const [flowId, flow] of Object.entries(flows)) {
            if (flow.sourceRef && !elements[flow.sourceRef]) {
                issues.push(`Flow ${flowId}: sourceRef '${flow.sourceRef}' не найден`);
            }
            if (flow.targetRef && !elements[flow.targetRef]) {
                issues.push(`Flow ${flowId}: targetRef '${flow.targetRef}' не найден`);
            }
        }
        
        // Результат валидации
        if (issues.length === 0) {
            console.log('✅ Все проверки пройдены');
            return true;
        } else {
            console.log(`❌ Найдено проблем: ${issues.length}`);
            issues.forEach(issue => console.log(`  - ${issue}`));
            return false;
        }
        
    } catch (error) {
        console.error('Ошибка валидации:', error.message);
        return false;
    }
}

// Примеры использования
if (require.main === module) {
    const processKey = process.argv[2];
    
    if (!processKey) {
        console.log('Использование: node get-json.js <process_key>');
        process.exit(1);
    }
    
    (async () => {
        try {
            // Базовое получение JSON
            const result = await getBPMNProcessJSON(processKey);
            console.log(`Процесс загружен: ${result.size} байт`);
            
            // Анализ сложности
            await analyzeProcessComplexity(processKey);
            
            // Валидация
            await validateProcessJSON(processKey);
            
            // Экспорт документации
            await exportProcessDocumentation(processKey);
            
        } catch (error) {
            console.error('Ошибка:', error.message);
        }
    })();
}

module.exports = {
    getBPMNProcessJSON,
    analyzeProcessComplexity,
    validateProcessJSON,
    exportProcessDocumentation
};
```

## Структура JSON процесса

### Основные разделы
```json
{
  "processId": "order-process",
  "processName": "Order Processing",
  "version": "1.0",
  "isExecutable": true,
  "elements": {
    "start1": {
      "id": "start1",
      "type": "startEvent",
      "name": "Order Received",
      "outgoing": ["flow1"]
    },
    "task1": {
      "id": "task1", 
      "type": "serviceTask",
      "name": "Validate Order",
      "incoming": ["flow1"],
      "outgoing": ["flow2"],
      "taskDefinition": {
        "type": "validation-service",
        "retries": 3
      }
    }
  },
  "flows": {
    "flow1": {
      "id": "flow1",
      "type": "sequenceFlow",
      "sourceRef": "start1",
      "targetRef": "task1"
    }
  }
}
```

## Использование JSON

### Импорт в другие системы
```go
// Загрузка JSON для выполнения
processJSON := response.JsonData
var processData map[string]interface{}
json.Unmarshal([]byte(processJSON), &processData)

// Запуск экземпляра процесса
instanceResponse, err := processClient.StartProcessInstance(ctx, &pb.StartProcessInstanceRequest{
    ProcessId: processData["processId"].(string),
    Variables: variables,
})
```

### Миграция между средами
```bash
# Экспорт из dev
curl -H "X-API-Key: dev-key" \
  http://localhost:27555/api/v1/bpmn/process-key/json > process.json

# Импорт в prod
curl -H "X-API-Key: prod-key" \
  -X POST -d @process.json \
  http://prod:27555/api/v1/bpmn/parse
```

## Возможные ошибки

### gRPC Status Codes
- `NOT_FOUND` (5): Процесс с указанным ключом не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Ошибка генерации JSON

### Примеры ошибок
```json
{
  "success": false,
  "message": "Process with key 'invalid-key' not found"
}
```

## Связанные методы
- [GetBPMNProcess](get-bpmn-process.md) - Метаданные процесса
- [ParseBPMNFile](parse-bpmn-file.md) - Загрузка процесса
- [ListBPMNProcesses](list-bpmn-processes.md) - Список доступных процессов

