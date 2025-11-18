# ParseBPMNFile

## Описание
Парсит BPMN файл, преобразует в JSON структуру и сохраняет в хранилище для последующего выполнения.

## Синтаксис
```protobuf
rpc ParseBPMNFile(ParseBPMNFileRequest) returns (ParseBPMNFileResponse);
```

## Package
```protobuf
package parser;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `parser` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### ParseBPMNFileRequest
```protobuf
message ParseBPMNFileRequest {
  string file_path = 1;      // Путь к BPMN файлу
  string process_id = 2;     // Опциональный ID процесса (если не указан, извлекается из файла)
  bool force = 3;            // Принудительная перезаписка существующего процесса
}
```

#### Поля:
- **file_path** (string, required): Путь к BPMN файлу относительно рабочей директории
- **process_id** (string, optional): Пользовательский ID процесса. Если не указан, используется ID из BPMN файла
- **force** (bool, optional): Если `true`, перезаписывает существующий процесс с таким же ID

## Параметры ответа

### ParseBPMNFileResponse  
```protobuf
message ParseBPMNFileResponse {
  bool success = 1;          // Статус успешности операции
  string message = 2;        // Сообщение о результате
  string process_key = 3;    // Уникальный ключ процесса в системе
  string process_id = 4;     // ID процесса
  int32 version = 5;         // Версия процесса
  int32 elements_count = 6;  // Количество элементов в процессе
  string created_at = 7;     // Время создания (ISO 8601)
}
```

#### Поля ответа:
- **success** (bool): `true` если парсинг прошел успешно
- **message** (string): Описание результата операции
- **process_key** (string): Сгенерированный уникальный ключ процесса для использования в API
- **process_id** (string): ID процесса (из запроса или извлеченный из BPMN)
- **version** (int32): Номер версии процесса (автоинкремент)
- **elements_count** (int32): Количество BPMN элементов в процессе
- **created_at** (string): Timestamp создания в формате ISO 8601

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
    // Подключение к серверу
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewParserServiceClient(conn)
    
    // Добавление API ключа в metadata
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Парсинг BPMN файла
    response, err := client.ParseBPMNFile(ctx, &pb.ParseBPMNFileRequest{
        FilePath:  "processes/order-process.bpmn",
        ProcessId: "order-process-v1",
        Force:     false,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Процесс успешно загружен!\n")
        fmt.Printf("Process Key: %s\n", response.ProcessKey)
        fmt.Printf("Version: %d\n", response.Version)
        fmt.Printf("Elements: %d\n", response.ElementsCount)
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}
```

### Python
```python
import grpc
import parser_pb2
import parser_pb2_grpc

def parse_bpmn_file():
    # Создание канала
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    
    # API ключ в metadata
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Создание запроса
    request = parser_pb2.ParseBPMNFileRequest(
        file_path="processes/order-process.bpmn",
        process_id="order-process-v1",
        force=False
    )
    
    try:
        # Вызов метода
        response = stub.ParseBPMNFile(request, metadata=metadata)
        
        if response.success:
            print(f"Процесс успешно загружен!")
            print(f"Process Key: {response.process_key}")
            print(f"Version: {response.version}")
            print(f"Elements: {response.elements_count}")
        else:
            print(f"Ошибка: {response.message}")
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")

if __name__ == "__main__":
    parse_bpmn_file()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

// Загрузка proto файла
const PROTO_PATH = path.join(__dirname, 'parser.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

function parseBPMNFile() {
    // Создание клиента
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    // Metadata с API ключом
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    // Создание запроса
    const request = {
        file_path: 'processes/order-process.bpmn',
        process_id: 'order-process-v1',
        force: false
    };
    
    // Вызов метода
    client.parseBPMNFile(request, metadata, (error, response) => {
        if (error) {
            console.error('gRPC Error:', error.code, error.details);
            return;
        }
        
        if (response.success) {
            console.log('Процесс успешно загружен!');
            console.log(`Process Key: ${response.process_key}`);
            console.log(`Version: ${response.version}`);
            console.log(`Elements: ${response.elements_count}`);
        } else {
            console.log(`Ошибка: ${response.message}`);
        }
    });
}

parseBPMNFile();
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный путь к файлу или некорректные параметры
- `NOT_FOUND` (5): BPMN файл не найден по указанному пути
- `ALREADY_EXISTS` (6): Процесс с таким ID уже существует (используйте force=true)
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Ошибка парсинга BPMN или внутренняя ошибка сервера

### Примеры ошибок
```json
{
  "success": false,
  "message": "BPMN file not found: processes/missing-file.bpmn"
}
```

```json
{
  "success": false,
  "message": "Process with ID 'order-process-v1' already exists. Use force=true to overwrite."
}
```

```json
{
  "success": false,
  "message": "Invalid BPMN: Missing start event in process definition"
}
```

## Валидация BPMN

Парсер выполняет следующие проверки:
- ✅ Валидность XML структуры
- ✅ Соответствие BPMN 2.0 схеме
- ✅ Наличие обязательных элементов (StartEvent)
- ✅ Корректность flow sequence connections
- ✅ Уникальность ID элементов
- ✅ Валидность FEEL выражений

## Связанные методы
- [ListBPMNProcesses](list-bpmn-processes.md) - Список загруженных процессов
- [GetBPMNProcess](get-bpmn-process.md) - Получение деталей процесса
- [DeleteBPMNProcess](delete-bpmn-process.md) - Удаление процесса
- [GetBPMNProcessJSON](get-bpmn-process-json.md) - JSON данные процесса

## Формат BPMN файла

### Поддерживаемые элементы
- **События**: StartEvent, EndEvent, IntermediateCatchEvent, IntermediateThrowEvent, BoundaryEvent
- **Задачи**: ServiceTask, UserTask, ScriptTask, SendTask, ReceiveTask, CallActivity
- **Шлюзы**: ExclusiveGateway, ParallelGateway, InclusiveGateway, EventBasedGateway
- **Потоки**: SequenceFlow, MessageFlow
- **Данные**: DataObject, DataStore

### Пример BPMN файла
```xml
<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:atom="http://atom.org/bpmn">
  <bpmn:process id="order-process" isExecutable="true">
    <bpmn:startEvent id="start"/>
    <bpmn:serviceTask id="validate-order" name="Validate Order">
      <bpmn:extensionElements>
        <atom:taskDefinition type="validation-service"/>
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:endEvent id="end"/>
    <bpmn:sequenceFlow sourceRef="start" targetRef="validate-order"/>
    <bpmn:sequenceFlow sourceRef="validate-order" targetRef="end"/>
  </bpmn:process>
</bpmn:definitions>
```
