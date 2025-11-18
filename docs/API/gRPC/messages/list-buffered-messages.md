# ListBufferedMessages

## Описание
Возвращает список буферизованных сообщений, которые не были сопоставлены с процессами. Поддерживает пагинацию и сортировку.

## Синтаксис
```protobuf
rpc ListBufferedMessages(ListBufferedMessagesRequest) returns (ListBufferedMessagesResponse);
```

## Package
```protobuf
package messages;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `messages` или `*`

## Параметры запроса

### ListBufferedMessagesRequest
```protobuf
message ListBufferedMessagesRequest {
  string tenant_id = 1;     // ID тенанта
  int32 limit = 2;          // Лимит (deprecated)
  int32 offset = 3;         // Смещение (deprecated)
  int32 page_size = 4;      // Размер страницы (по умолчанию 20)
  int32 page = 5;           // Номер страницы (начиная с 1)
  string sort_by = 6;       // Поле сортировки
  string sort_order = 7;    // Порядок: ASC/DESC
}
```

## Параметры ответа

### ListBufferedMessagesResponse
```protobuf
message ListBufferedMessagesResponse {
  repeated BufferedMessage messages = 1;  // Список сообщений
  int32 total_count = 2;                  // Общее количество
  bool success = 3;                       // Статус успешности
  string message = 4;                     // Сообщение о результате
  int32 page = 5;                         // Текущая страница
  int32 page_size = 6;                    // Размер страницы
  int32 total_pages = 7;                  // Общее количество страниц
}

message BufferedMessage {
  string id = 1;                          // ID сообщения
  string tenant_id = 2;                   // ID тенанта
  string name = 3;                        // Имя сообщения
  string correlation_key = 4;             // Ключ корреляции
  map<string, string> variables = 5;      // Переменные
  int64 published_at = 6;                 // Время публикации
  int64 buffered_at = 7;                  // Время буферизации
  int64 expires_at = 8;                   // Время истечения
  string reason = 9;                      // Причина буферизации
  string element_id = 10;                 // ID элемента
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
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    
    pb "atom-engine/proto/messages/messagespb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewMessagesServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Список всех буферизованных сообщений
    response, err := client.ListBufferedMessages(ctx, &pb.ListBufferedMessagesRequest{
        PageSize: 10,
        Page:     1,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("📋 Найдено %d буферизованных сообщений (страница %d из %d)\n", 
                   response.TotalCount, response.Page, response.TotalPages)
        
        for _, msg := range response.Messages {
            publishedAt := time.Unix(msg.PublishedAt, 0)
            expiresAt := time.Unix(msg.ExpiresAt, 0)
            
            fmt.Printf("• %s [%s] - опубликовано: %s, истекает: %s\n",
                       msg.Id, msg.Name, 
                       publishedAt.Format("2006-01-02 15:04:05"),
                       expiresAt.Format("2006-01-02 15:04:05"))
            
            if msg.Reason != "" {
                fmt.Printf("  Причина: %s\n", msg.Reason)
            }
        }
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.Message)
    }
}

// Список с сортировкой
func listSorted() {
    // ... client setup ...
    
    response, err := client.ListBufferedMessages(ctx, &pb.ListBufferedMessagesRequest{
        PageSize:  20,
        Page:      1,
        SortBy:    "published_at",
        SortOrder: "DESC", // Новые сначала
    })
    
    // ... обработка ответа ...
}
```

### Python
```python
import grpc
from datetime import datetime

import messages_pb2
import messages_pb2_grpc

def list_buffered_messages(page=1, page_size=20, sort_by="published_at", sort_order="DESC"):
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = messages_pb2.ListBufferedMessagesRequest(
        page=page,
        page_size=page_size,
        sort_by=sort_by,
        sort_order=sort_order
    )
    
    try:
        response = stub.ListBufferedMessages(request, metadata=metadata)
        
        if response.success:
            print(f"📋 Найдено {response.total_count} буферизованных сообщений")
            print(f"   Страница {response.page} из {response.total_pages}")
            
            for msg in response.messages:
                published_at = datetime.fromtimestamp(msg.published_at)
                expires_at = datetime.fromtimestamp(msg.expires_at)
                
                print(f"• {msg.id} [{msg.name}]")
                print(f"  Опубликовано: {published_at.strftime('%Y-%m-%d %H:%M:%S')}")
                print(f"  Истекает: {expires_at.strftime('%Y-%m-%d %H:%M:%S')}")
                
                if msg.correlation_key:
                    print(f"  Ключ корреляции: {msg.correlation_key}")
                
                if msg.reason:
                    print(f"  Причина: {msg.reason}")
                print()
            
            return response.messages
        else:
            print(f"❌ Ошибка: {response.message}")
            return []
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return []

# Пример использования
if __name__ == "__main__":
    # Первая страница
    messages = list_buffered_messages()
    
    # Сортировка по времени истечения
    messages = list_buffered_messages(
        sort_by="expires_at",
        sort_order="ASC"  # Скоро истекающие сначала
    )
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'messages.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const messagesProto = grpc.loadPackageDefinition(packageDefinition).messages;

async function listBufferedMessages(page = 1, pageSize = 20, sortBy = "published_at", sortOrder = "DESC") {
    const client = new messagesProto.MessagesService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            page: page,
            page_size: pageSize,
            sort_by: sortBy,
            sort_order: sortOrder
        };
        
        client.listBufferedMessages(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`📋 Найдено ${response.total_count} буферизованных сообщений`);
                console.log(`   Страница ${response.page} из ${response.total_pages}`);
                
                response.messages.forEach(msg => {
                    const publishedAt = new Date(msg.published_at * 1000);
                    const expiresAt = new Date(msg.expires_at * 1000);
                    
                    console.log(`• ${msg.id} [${msg.name}]`);
                    console.log(`  Опубликовано: ${publishedAt.toLocaleString()}`);
                    console.log(`  Истекает: ${expiresAt.toLocaleString()}`);
                    
                    if (msg.correlation_key) {
                        console.log(`  Ключ корреляции: ${msg.correlation_key}`);
                    }
                    
                    if (msg.reason) {
                        console.log(`  Причина: ${msg.reason}`);
                    }
                    console.log();
                });
                
                resolve(response.messages);
            } else {
                console.log(`❌ Ошибка: ${response.message}`);
                resolve([]);
            }
        });
    });
}

// Примеры использования
async function examples() {
    // Все буферизованные сообщения
    const messages = await listBufferedMessages();
    
    // С пагинацией
    const page2 = await listBufferedMessages(2, 10);
    
    // Сортировка по имени
    const sortedByName = await listBufferedMessages(1, 20, "name", "ASC");
}

examples().catch(console.error);
```

## Поля сортировки

### Доступные поля
- **published_at**: Время публикации (по умолчанию)
- **buffered_at**: Время буферизации
- **expires_at**: Время истечения
- **name**: Имя сообщения
- **correlation_key**: Ключ корреляции

### Порядок сортировки
- **ASC**: По возрастанию
- **DESC**: По убыванию (по умолчанию)

## Причины буферизации
- **no_subscription**: Нет подписки на сообщение
- **no_correlation**: Не найден процесс для корреляции
- **multiple_matches**: Несколько подходящих процессов

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры пагинации
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

## Связанные методы
- [PublishMessage](publish-message.md) - Публикация сообщений
- [CleanupExpiredMessages](cleanup-expired-messages.md) - Очистка истекших сообщений
- [GetMessageStats](get-message-stats.md) - Статистика сообщений
