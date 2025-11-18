# ListBPMNProcesses

## Описание
Получает список всех загруженных BPMN процессов с поддержкой фильтрации и пагинации.

## Синтаксис
```protobuf
rpc ListBPMNProcesses(ListBPMNProcessesRequest) returns (ListBPMNProcessesResponse);
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

### ListBPMNProcessesRequest
```protobuf
message ListBPMNProcessesRequest {
  int32 limit = 1;           // Максимальное количество записей
  int32 offset = 2;          // Смещение для пагинации
  string status = 3;         // Фильтр по статусу процесса
  string process_id = 4;     // Фильтр по ID процесса
}
```

#### Поля:
- **limit** (int32, optional): Максимальное количество процессов в ответе (по умолчанию: 50, максимум: 1000)
- **offset** (int32, optional): Количество записей для пропуска (для пагинации)
- **status** (string, optional): Фильтр по статусу процесса (`ACTIVE`, `DEPLOYED`, `INACTIVE`)
- **process_id** (string, optional): Фильтр по ID или префиксу ID процесса

## Параметры ответа

### ListBPMNProcessesResponse
```protobuf
message ListBPMNProcessesResponse {
  repeated BPMNProcessInfo processes = 1;  // Список процессов
  int32 total_count = 2;                   // Общее количество процессов
  bool has_more = 3;                       // Есть ли еще записи
  bool success = 4;                        // Статус успешности
  string message = 5;                      // Сообщение о результате
}

message BPMNProcessInfo {
  string process_key = 1;      // Уникальный ключ процесса
  string process_id = 2;       // ID процесса
  int32 version = 3;           // Версия процесса
  string status = 4;           // Статус процесса
  int32 elements_count = 5;    // Количество элементов
  string created_at = 6;       // Время создания
  string updated_at = 7;       // Время обновления
  string file_path = 8;        // Путь к оригинальному файлу
  int64 file_size = 9;         // Размер файла в байтах
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
    
    // Получение всех процессов
    response, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
        Limit:  50,
        Offset: 0,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Найдено процессов: %d (всего: %d)\n", 
            len(response.Processes), response.TotalCount)
        
        for _, process := range response.Processes {
            fmt.Printf("- %s (v%d): %s [%d элементов]\n",
                process.ProcessId, process.Version, 
                process.Status, process.ElementsCount)
        }
        
        if response.HasMore {
            fmt.Println("Есть еще записи...")
        }
    }
}

// Пагинация через все процессы
func listAllProcesses(client pb.ParserServiceClient, ctx context.Context) {
    const pageSize = 20
    offset := 0
    
    for {
        response, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
            Limit:  pageSize,
            Offset: int32(offset),
        })
        
        if err != nil {
            log.Printf("Ошибка загрузки страницы: %v", err)
            break
        }
        
        // Обработка текущей страницы
        for _, process := range response.Processes {
            fmt.Printf("Process: %s\n", process.ProcessId)
        }
        
        if !response.HasMore {
            break
        }
        
        offset += pageSize
    }
}
```

### Python
```python
import grpc
import parser_pb2
import parser_pb2_grpc

def list_bpmn_processes():
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Фильтрация по активным процессам
    request = parser_pb2.ListBPMNProcessesRequest(
        limit=50,
        offset=0,
        status="ACTIVE"
    )
    
    try:
        response = stub.ListBPMNProcesses(request, metadata=metadata)
        
        if response.success:
            print(f"Найдено активных процессов: {len(response.processes)}")
            print(f"Всего процессов: {response.total_count}")
            
            for process in response.processes:
                print(f"- {process.process_id} v{process.version}")
                print(f"  Ключ: {process.process_key}")
                print(f"  Статус: {process.status}")
                print(f"  Элементов: {process.elements_count}")
                print(f"  Создан: {process.created_at}")
                print()
                
        else:
            print(f"Ошибка: {response.message}")
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")

def search_processes_by_prefix(prefix):
    """Поиск процессов по префиксу ID"""
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = parser_pb2.ListBPMNProcessesRequest(
        process_id=prefix,
        limit=100
    )
    
    response = stub.ListBPMNProcesses(request, metadata=metadata)
    
    if response.success:
        matching_processes = [p for p in response.processes 
                            if p.process_id.startswith(prefix)]
        print(f"Найдено процессов с префиксом '{prefix}': {len(matching_processes)}")
        return matching_processes
    
    return []

if __name__ == "__main__":
    list_bpmn_processes()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'parser.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

function listBPMNProcesses() {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    const request = {
        limit: 50,
        offset: 0,
        status: 'ACTIVE'
    };
    
    client.listBPMNProcesses(request, metadata, (error, response) => {
        if (error) {
            console.error('gRPC Error:', error.details);
            return;
        }
        
        if (response.success) {
            console.log(`Найдено процессов: ${response.processes.length}`);
            console.log(`Всего: ${response.total_count}`);
            
            response.processes.forEach(process => {
                console.log(`- ${process.process_id} v${process.version}`);
                console.log(`  Статус: ${process.status}`);
                console.log(`  Элементов: ${process.elements_count}`);
                console.log(`  Размер файла: ${process.file_size} байт`);
            });
            
            if (response.has_more) {
                console.log('Есть еще записи для загрузки');
            }
        } else {
            console.log(`Ошибка: ${response.message}`);
        }
    });
}

// Асинхронная загрузка всех страниц
async function getAllProcesses() {
    return new Promise((resolve, reject) => {
        const client = new parserProto.ParserService('localhost:27500',
            grpc.credentials.createInsecure());
        
        const metadata = new grpc.Metadata();
        metadata.add('x-api-key', 'your-api-key-here');
        
        const allProcesses = [];
        let offset = 0;
        const pageSize = 20;
        
        function loadPage() {
            const request = {
                limit: pageSize,
                offset: offset
            };
            
            client.listBPMNProcesses(request, metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (!response.success) {
                    reject(new Error(response.message));
                    return;
                }
                
                allProcesses.push(...response.processes);
                
                if (response.has_more) {
                    offset += pageSize;
                    loadPage(); // Загрузка следующей страницы
                } else {
                    resolve(allProcesses);
                }
            });
        }
        
        loadPage();
    });
}

// Использование
listBPMNProcesses();

getAllProcesses()
    .then(processes => {
        console.log(`Всего загружено процессов: ${processes.length}`);
    })
    .catch(error => {
        console.error('Ошибка загрузки:', error.message);
    });
```

## Фильтрация и поиск

### Фильтр по статусу
```go
// Только активные процессы
response, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
    Status: "ACTIVE",
    Limit:  100,
})
```

### Поиск по ID
```go
// Процессы начинающиеся с "order-"
response, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
    ProcessId: "order-",
    Limit:     50,
})
```

### Пагинация
```go
// Вторая страница по 25 записей
response, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
    Limit:  25,
    Offset: 25,
})
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры пагинации
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Ошибка базы данных

### Примеры ошибок
```json
{
  "success": false,
  "message": "Invalid limit: must be between 1 and 1000"
}
```

## Связанные методы
- [ParseBPMNFile](parse-bpmn-file.md) - Загрузка нового процесса
- [GetBPMNProcess](get-bpmn-process.md) - Детали конкретного процесса
- [DeleteBPMNProcess](delete-bpmn-process.md) - Удаление процесса
- [GetBPMNStats](get-bpmn-stats.md) - Статистика парсера

## Статусы процессов

- **ACTIVE** - Процесс активен и готов к выполнению
- **DEPLOYED** - Процесс развернут в системе
- **INACTIVE** - Процесс неактивен (возможно, устаревшая версия)

