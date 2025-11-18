# StartProcessInstance

## Описание
Запускает новый экземпляр BPMN процесса с опциональными переменными для инициализации.

## Синтаксис
```protobuf
rpc StartProcessInstance(StartProcessInstanceRequest) returns (StartProcessInstanceResponse);
```

## Package
```protobuf
package atom.process.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### StartProcessInstanceRequest
```protobuf
message StartProcessInstanceRequest {
  string process_id = 1;              // ID процесса для запуска
  map<string, string> variables = 2;  // Переменные для инициализации
}
```

#### Поля:
- **process_id** (string, required): ID или ключ BPMN процесса для запуска
- **variables** (map<string, string>, optional): Переменные процесса в виде ключ-значение (JSON строки)

## Параметры ответа

### StartProcessInstanceResponse
```protobuf
message StartProcessInstanceResponse {
  string instance_id = 1;    // Уникальный ID экземпляра процесса
  string status = 2;         // Статус экземпляра (ACTIVE, COMPLETED, FAILED)
  bool success = 3;          // Статус успешности операции
  string message = 4;        // Сообщение о результате
}
```

#### Поля ответа:
- **instance_id** (string): Уникальный идентификатор созданного экземпляра процесса
- **status** (string): Текущий статус экземпляра процесса
  - `ACTIVE` - Процесс запущен и выполняется
  - `COMPLETED` - Процесс завершен успешно
  - `FAILED` - Процесс завершен с ошибкой
- **success** (bool): `true` если экземпляр успешно создан
- **message** (string): Описание результата операции

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
    
    pb "atom-engine/proto/process/processpb"
)

func main() {
    // Подключение к серверу
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewProcessServiceClient(conn)
    
    // Добавление API ключа в metadata
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Переменные для процесса
    variables := map[string]string{
        "orderId":    "12345",
        "customerId": "67890",
        "amount":     "199.99",
        "priority":   "high",
    }
    
    // Запуск экземпляра процесса
    response, err := client.StartProcessInstance(ctx, &pb.StartProcessInstanceRequest{
        ProcessId: "order-process-v1",
        Variables: variables,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Процесс успешно запущен!\n")
        fmt.Printf("Instance ID: %s\n", response.InstanceId)
        fmt.Printf("Status: %s\n", response.Status)
    } else {
        fmt.Printf("Ошибка запуска: %s\n", response.Message)
    }
}
```

### Python
```python
import grpc
import json
import process_pb2
import process_pb2_grpc

def start_process_instance():
    # Создание канала
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    
    # API ключ в metadata
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Переменные процесса (значения должны быть JSON строками)
    variables = {
        'orderId': '12345',
        'customerId': '67890',
        'amount': '199.99',
        'orderData': json.dumps({
            'items': [
                {'name': 'Product A', 'quantity': 2},
                {'name': 'Product B', 'quantity': 1}
            ],
            'shipping': 'express'
        })
    }
    
    # Создание запроса
    request = process_pb2.StartProcessInstanceRequest(
        process_id="order-process-v1",
        variables=variables
    )
    
    try:
        # Вызов метода
        response = stub.StartProcessInstance(request, metadata=metadata)
        
        if response.success:
            print(f"Процесс успешно запущен!")
            print(f"Instance ID: {response.instance_id}")
            print(f"Status: {response.status}")
        else:
            print(f"Ошибка запуска: {response.message}")
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")

if __name__ == "__main__":
    start_process_instance()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

// Загрузка proto файла
const PROTO_PATH = path.join(__dirname, 'process.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

function startProcessInstance() {
    // Создание клиента
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    // Metadata с API ключом
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    // Переменные процесса
    const variables = {
        'orderId': '12345',
        'customerId': '67890',
        'amount': '199.99',
        'orderData': JSON.stringify({
            items: [
                { name: 'Product A', quantity: 2 },
                { name: 'Product B', quantity: 1 }
            ],
            shipping: 'express'
        })
    };
    
    // Создание запроса
    const request = {
        process_id: 'order-process-v1',
        variables: variables
    };
    
    // Вызов метода
    client.startProcessInstance(request, metadata, (error, response) => {
        if (error) {
            console.error('gRPC Error:', error.code, error.details);
            return;
        }
        
        if (response.success) {
            console.log('Процесс успешно запущен!');
            console.log(`Instance ID: ${response.instance_id}`);
            console.log(`Status: ${response.status}`);
        } else {
            console.log(`Ошибка запуска: ${response.message}`);
        }
    });
}

startProcessInstance();
```

## Работа с переменными

### Форматы переменных
Переменные передаются как `map<string, string>`, где значения должны быть JSON строками:

```go
variables := map[string]string{
    // Простые типы
    "orderId":    "12345",
    "amount":     "199.99",
    "isUrgent":   "true",
    
    // Сложные объекты (JSON)
    "customer": `{
        "name": "John Doe",
        "email": "john@example.com",
        "address": {
            "street": "123 Main St",
            "city": "New York"
        }
    }`,
    
    // Массивы (JSON)
    "items": `[
        {"id": 1, "name": "Product A", "price": 99.99},
        {"id": 2, "name": "Product B", "price": 149.99}
    ]`
}
```

### Использование переменных в BPMN
Переменные доступны в FEEL выражениях:
```xml
<!-- Условие в Exclusive Gateway -->
<bpmn:conditionExpression>amount > 100</bpmn:conditionExpression>

<!-- Mapping в Service Task -->
<atom:ioMapping>
  <atom:input source="orderId" target="request.orderId"/>
  <atom:input source="customer.email" target="request.email"/>
</atom:ioMapping>
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный process_id или некорректные переменные
- `NOT_FOUND` (5): Процесс с указанным ID не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Внутренняя ошибка при создании экземпляра

### Примеры ошибок
```json
{
  "success": false,
  "message": "Process with ID 'unknown-process' not found"
}
```

```json
{
  "success": false,
  "message": "Invalid JSON in variable 'customer': unexpected token"
}
```

```json
{
  "success": false,
  "message": "Process definition has no start event"
}
```

## Статусы экземпляров

### ACTIVE
Процесс запущен и выполняется. Могут быть активные токены на различных элементах.

### COMPLETED  
Процесс успешно завершен. Достигнуто EndEvent без ошибок.

### FAILED
Процесс завершен с ошибкой. Возможные причины:
- Необработанное исключение в Service Task
- Timeout активности без boundary event
- Критическая ошибка системы

## Мониторинг выполнения

После запуска процесса можно отслеживать его состояние:

```go
// Получение статуса экземпляра
statusResponse, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
    InstanceId: response.InstanceId,
})

// Получение активных токенов
tokensResponse, err := client.ListTokens(ctx, &pb.ListTokensRequest{
    InstanceId: response.InstanceId,
    State: "ACTIVE",
})
```

## Связанные методы
- [GetProcessInstanceStatus](get-process-instance-status.md) - Получение статуса экземпляра
- [CancelProcessInstance](cancel-process-instance.md) - Отмена экземпляра
- [ListProcessInstances](list-process-instances.md) - Список экземпляров
- [ListTokens](list-tokens.md) - Список токенов экземпляра
- [GetProcessInstanceInfo](get-process-instance-info.md) - Полная информация об экземпляре

## Паттерны использования

### Синхронный запуск с ожиданием завершения
```go
func startAndWaitProcess(client pb.ProcessServiceClient, ctx context.Context) error {
    // Запуск процесса
    response, err := client.StartProcessInstance(ctx, &pb.StartProcessInstanceRequest{
        ProcessId: "sync-process",
        Variables: variables,
    })
    if err != nil {
        return err
    }
    
    // Ожидание завершения
    for {
        status, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
            InstanceId: response.InstanceId,
        })
        if err != nil {
            return err
        }
        
        if status.Status == "COMPLETED" || status.Status == "FAILED" {
            break
        }
        
        time.Sleep(time.Second)
    }
    
    return nil
}
```

### Массовый запуск процессов
```go
func startMultipleProcesses(client pb.ProcessServiceClient, ctx context.Context, count int) {
    var wg sync.WaitGroup
    
    for i := 0; i < count; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            
            variables := map[string]string{
                "batchId": fmt.Sprintf("batch-%d", index),
                "index":   fmt.Sprintf("%d", index),
            }
            
            _, err := client.StartProcessInstance(ctx, &pb.StartProcessInstanceRequest{
                ProcessId: "batch-process",
                Variables: variables,
            })
            if err != nil {
                log.Printf("Failed to start process %d: %v", index, err)
            }
        }(i)
    }
    
    wg.Wait()
}
```
