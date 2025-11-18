# GetBPMNProcess

## Описание
Получает детальную информацию о конкретном BPMN процессе по его ключу или ID.

## Синтаксис
```protobuf
rpc GetBPMNProcess(GetBPMNProcessRequest) returns (GetBPMNProcessResponse);
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

### GetBPMNProcessRequest
```protobuf
message GetBPMNProcessRequest {
  string process_key = 1;        // Уникальный ключ процесса
  bool include_elements = 2;     // Включить детали элементов
  bool include_original = 3;     // Включить оригинальный BPMN XML
}
```

#### Поля:
- **process_key** (string, required): Уникальный ключ процесса (возвращается при парсинге)
- **include_elements** (bool, optional): Включить детальную информацию об элементах процесса
- **include_original** (bool, optional): Включить оригинальный BPMN XML в ответ

## Параметры ответа

### GetBPMNProcessResponse
```protobuf
message GetBPMNProcessResponse {
  BPMNProcessDetails process = 1; // Детали процесса
  bool success = 2;               // Статус успешности
  string message = 3;             // Сообщение о результате
}

message BPMNProcessDetails {
  string process_key = 1;         // Уникальный ключ
  string process_id = 2;          // ID процесса
  int32 version = 3;              // Версия
  string status = 4;              // Статус
  int32 elements_count = 5;       // Количество элементов
  string created_at = 6;          // Время создания
  string updated_at = 7;          // Время обновления
  string file_path = 8;           // Путь к файлу
  int64 file_size = 9;            // Размер файла
  string original_xml = 10;       // Оригинальный BPMN XML
  repeated BPMNElement elements = 11; // Список элементов
  map<string, string> metadata = 12;  // Метаданные процесса
}

message BPMNElement {
  string id = 1;                  // ID элемента
  string type = 2;                // Тип элемента
  string name = 3;                // Имя элемента
  map<string, string> properties = 4; // Свойства элемента
}
```

## Примеры использования

### Go
```go
package main

import (
    "context"
    "fmt"
    "log"
    
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
    
    // Получение базовой информации о процессе
    response, err := client.GetBPMNProcess(ctx, &pb.GetBPMNProcessRequest{
        ProcessKey:      "atom-7-1k2-PVn4Y9j-CF5M",
        IncludeElements: false,
        IncludeOriginal: false,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        process := response.Process
        fmt.Printf("Процесс: %s (v%d)\n", process.ProcessId, process.Version)
        fmt.Printf("Ключ: %s\n", process.ProcessKey)
        fmt.Printf("Статус: %s\n", process.Status)
        fmt.Printf("Элементов: %d\n", process.ElementsCount)
        fmt.Printf("Создан: %s\n", process.CreatedAt)
        fmt.Printf("Размер файла: %d байт\n", process.FileSize)
        
        // Вывод метаданных
        if len(process.Metadata) > 0 {
            fmt.Println("Метаданные:")
            for key, value := range process.Metadata {
                fmt.Printf("  %s: %s\n", key, value)
            }
        }
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

// Получение процесса с элементами
func getProcessWithElements(client pb.ParserServiceClient, ctx context.Context, processKey string) {
    response, err := client.GetBPMNProcess(ctx, &pb.GetBPMNProcessRequest{
        ProcessKey:      processKey,
        IncludeElements: true,
        IncludeOriginal: false,
    })
    
    if err != nil {
        log.Printf("Ошибка: %v", err)
        return
    }
    
    if response.Success {
        process := response.Process
        fmt.Printf("Процесс %s содержит %d элементов:\n", 
            process.ProcessId, len(process.Elements))
        
        // Группировка элементов по типу
        elementsByType := make(map[string][]pb.BPMNElement)
        for _, element := range process.Elements {
            elementsByType[element.Type] = append(elementsByType[element.Type], element)
        }
        
        // Вывод по типам
        for elementType, elements := range elementsByType {
            fmt.Printf("\n%s (%d):\n", elementType, len(elements))
            for _, element := range elements {
                name := element.Name
                if name == "" {
                    name = "<без имени>"
                }
                fmt.Printf("  - %s: %s\n", element.Id, name)
            }
        }
    }
}
```

### Python
```python
import grpc
import parser_pb2
import parser_pb2_grpc
from collections import defaultdict

def get_bpmn_process_details(process_key):
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Получение полной информации о процессе
    request = parser_pb2.GetBPMNProcessRequest(
        process_key=process_key,
        include_elements=True,
        include_original=True
    )
    
    try:
        response = stub.GetBPMNProcess(request, metadata=metadata)
        
        if response.success:
            process = response.process
            
            print(f"=== Процесс {process.process_id} ===")
            print(f"Ключ: {process.process_key}")
            print(f"Версия: {process.version}")
            print(f"Статус: {process.status}")
            print(f"Элементов: {process.elements_count}")
            print(f"Файл: {process.file_path} ({process.file_size} байт)")
            print(f"Создан: {process.created_at}")
            
            # Анализ элементов
            if process.elements:
                analyze_process_elements(process.elements)
            
            # Сохранение оригинального XML
            if process.original_xml:
                save_original_xml(process.process_id, process.original_xml)
                
        else:
            print(f"Ошибка: {response.message}")
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")

def analyze_process_elements(elements):
    """Анализ элементов процесса"""
    print("\n=== Анализ элементов ===")
    
    # Группировка по типам
    by_type = defaultdict(list)
    for element in elements:
        by_type[element.type].append(element)
    
    # Статистика по типам
    print("Статистика по типам:")
    for element_type, element_list in by_type.items():
        print(f"  {element_type}: {len(element_list)}")
    
    # Детали важных элементов
    important_types = ['startEvent', 'endEvent', 'serviceTask', 'userTask']
    for element_type in important_types:
        if element_type in by_type:
            print(f"\n{element_type.upper()}:")
            for element in by_type[element_type]:
                name = element.name or "<без имени>"
                print(f"  - {element.id}: {name}")
                
                # Показываем важные свойства
                if element.properties:
                    for prop_key, prop_value in element.properties.items():
                        if prop_key in ['taskDefinition', 'conditionExpression', 'timerDefinition']:
                            print(f"    {prop_key}: {prop_value}")

def save_original_xml(process_id, xml_content):
    """Сохранение оригинального BPMN XML"""
    filename = f"{process_id}_original.bpmn"
    try:
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(xml_content)
        print(f"Оригинальный BPMN сохранен в: {filename}")
    except Exception as e:
        print(f"Ошибка сохранения XML: {e}")

def compare_process_versions(process_key1, process_key2):
    """Сравнение двух версий процесса"""
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Получение обеих версий
    processes = []
    for key in [process_key1, process_key2]:
        request = parser_pb2.GetBPMNProcessRequest(
            process_key=key,
            include_elements=True
        )
        response = stub.GetBPMNProcess(request, metadata=metadata)
        if response.success:
            processes.append(response.process)
    
    if len(processes) == 2:
        p1, p2 = processes
        print(f"Сравнение {p1.process_id} v{p1.version} и v{p2.version}")
        print(f"Элементов: {p1.elements_count} -> {p2.elements_count}")
        
        # Сравнение количества элементов по типам
        def count_by_type(elements):
            counts = defaultdict(int)
            for element in elements:
                counts[element.type] += 1
            return counts
        
        counts1 = count_by_type(p1.elements)
        counts2 = count_by_type(p2.elements)
        
        all_types = set(counts1.keys()) | set(counts2.keys())
        for element_type in sorted(all_types):
            c1, c2 = counts1[element_type], counts2[element_type]
            if c1 != c2:
                print(f"  {element_type}: {c1} -> {c2}")

if __name__ == "__main__":
    # Пример использования
    process_key = "atom-7-1k2-PVn4Y9j-CF5M"
    get_bpmn_process_details(process_key)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const fs = require('fs').promises;

const PROTO_PATH = 'parser.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

async function getBPMNProcessDetails(processKey) {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            process_key: processKey,
            include_elements: true,
            include_original: true
        };
        
        client.getBPMNProcess(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            resolve(response.process);
        });
    });
}

async function analyzeProcess(processKey) {
    try {
        const process = await getBPMNProcessDetails(processKey);
        
        console.log(`=== Процесс ${process.process_id} ===`);
        console.log(`Ключ: ${process.process_key}`);
        console.log(`Версия: ${process.version}`);
        console.log(`Статус: ${process.status}`);
        console.log(`Элементов: ${process.elements_count}`);
        console.log(`Размер: ${process.file_size} байт`);
        
        // Анализ элементов
        if (process.elements && process.elements.length > 0) {
            console.log('\n=== Анализ элементов ===');
            
            // Группировка по типам
            const byType = {};
            process.elements.forEach(element => {
                if (!byType[element.type]) {
                    byType[element.type] = [];
                }
                byType[element.type].push(element);
            });
            
            // Вывод статистики
            console.log('Элементы по типам:');
            Object.entries(byType).forEach(([type, elements]) => {
                console.log(`  ${type}: ${elements.length}`);
            });
            
            // Детали service tasks
            if (byType.serviceTask) {
                console.log('\nService Tasks:');
                byType.serviceTask.forEach(task => {
                    const name = task.name || '<без имени>';
                    console.log(`  - ${task.id}: ${name}`);
                    
                    // Показать task definition если есть
                    if (task.properties && task.properties.taskDefinition) {
                        console.log(`    Type: ${task.properties.taskDefinition}`);
                    }
                });
            }
        }
        
        // Сохранение оригинального XML
        if (process.original_xml) {
            const filename = `${process.process_id}_v${process.version}.bpmn`;
            await fs.writeFile(filename, process.original_xml, 'utf8');
            console.log(`Оригинальный BPMN сохранен: ${filename}`);
        }
        
        // Метаданные
        if (process.metadata && Object.keys(process.metadata).length > 0) {
            console.log('\nМетаданные:');
            Object.entries(process.metadata).forEach(([key, value]) => {
                console.log(`  ${key}: ${value}`);
            });
        }
        
    } catch (error) {
        console.error('Ошибка:', error.message);
    }
}

// Функция для проверки процесса на валидность
async function validateProcess(processKey) {
    try {
        const process = await getBPMNProcessDetails(processKey);
        const issues = [];
        
        // Группировка элементов
        const byType = {};
        process.elements.forEach(element => {
            if (!byType[element.type]) {
                byType[element.type] = [];
            }
            byType[element.type].push(element);
        });
        
        // Проверки
        if (!byType.startEvent || byType.startEvent.length === 0) {
            issues.push('Отсутствует startEvent');
        }
        
        if (!byType.endEvent || byType.endEvent.length === 0) {
            issues.push('Отсутствует endEvent');
        }
        
        if (byType.startEvent && byType.startEvent.length > 1) {
            issues.push('Несколько startEvent');
        }
        
        // Проверка service tasks на наличие task definition
        if (byType.serviceTask) {
            byType.serviceTask.forEach(task => {
                if (!task.properties || !task.properties.taskDefinition) {
                    issues.push(`ServiceTask ${task.id} без task definition`);
                }
            });
        }
        
        console.log(`\n=== Валидация процесса ${process.process_id} ===`);
        if (issues.length === 0) {
            console.log('✅ Процесс валиден');
        } else {
            console.log('❌ Найдены проблемы:');
            issues.forEach(issue => console.log(`  - ${issue}`));
        }
        
        return issues.length === 0;
        
    } catch (error) {
        console.error('Ошибка валидации:', error.message);
        return false;
    }
}

// Примеры использования
const processKey = 'atom-7-1k2-PVn4Y9j-CF5M';

analyzeProcess(processKey);
validateProcess(processKey);
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный process_key
- `NOT_FOUND` (5): Процесс с указанным ключом не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Process with key 'invalid-key' not found"
}
```

## Связанные методы
- [ListBPMNProcesses](list-bpmn-processes.md) - Список всех процессов
- [ParseBPMNFile](parse-bpmn-file.md) - Загрузка процесса
- [GetBPMNProcessJSON](get-bpmn-process-json.md) - JSON представление процесса
- [DeleteBPMNProcess](delete-bpmn-process.md) - Удаление процесса

