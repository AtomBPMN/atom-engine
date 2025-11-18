# ListMessageSubscriptions

## Описание
Возвращает список активных подписок на сообщения. Подписки создаются автоматически при парсинге BPMN процессов с Message Events.

## Синтаксис
```protobuf
rpc ListMessageSubscriptions(ListMessageSubscriptionsRequest) returns (ListMessageSubscriptionsResponse);
```

## Package
```protobuf
package messages;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `messages` или `*`

## Параметры запроса

### ListMessageSubscriptionsRequest
```protobuf
message ListMessageSubscriptionsRequest {
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

### ListMessageSubscriptionsResponse
```protobuf
message ListMessageSubscriptionsResponse {
  repeated MessageSubscription subscriptions = 1;  // Список подписок
  int32 total_count = 2;                           // Общее количество
  bool success = 3;                                // Статус успешности
  string message = 4;                              // Сообщение о результате
  int32 page = 5;                                  // Текущая страница
  int32 page_size = 6;                             // Размер страницы
  int32 total_pages = 7;                           // Общее количество страниц
}

message MessageSubscription {
  string id = 1;                      // ID подписки
  string tenant_id = 2;               // ID тенанта
  string process_definition_key = 3;  // Ключ определения процесса
  int32 process_version = 4;          // Версия процесса
  string start_event_id = 5;          // ID стартового события
  string message_name = 6;            // Имя сообщения
  string message_ref = 7;             // Ссылка на сообщение
  string correlation_key = 8;         // Ключ корреляции
  bool is_active = 9;                 // Активна ли подписка
  int64 created_at = 10;              // Время создания
  int64 updated_at = 11;              // Время обновления
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
    
    // Список всех подписок
    response, err := client.ListMessageSubscriptions(ctx, &pb.ListMessageSubscriptionsRequest{
        PageSize: 10,
        Page:     1,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("📋 Найдено %d подписок на сообщения\n", response.TotalCount)
        
        for _, sub := range response.Subscriptions {
            status := "🟢"
            if !sub.IsActive {
                status = "🔴"
            }
            
            createdAt := time.Unix(sub.CreatedAt, 0)
            
            fmt.Printf("%s %s [%s]\n", status, sub.MessageName, sub.ProcessDefinitionKey)
            fmt.Printf("   Версия: %d | Событие: %s\n", sub.ProcessVersion, sub.StartEventId)
            fmt.Printf("   Создано: %s\n", createdAt.Format("2006-01-02 15:04:05"))
            
            if sub.CorrelationKey != "" {
                fmt.Printf("   Корреляция: %s\n", sub.CorrelationKey)
            }
            fmt.Println()
        }
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.Message)
    }
}

// Список только активных подписок
func listActiveSubscriptions() {
    // ... client setup ...
    
    response, err := client.ListMessageSubscriptions(ctx, &pb.ListMessageSubscriptionsRequest{
        PageSize:  50,
        Page:      1,
        SortBy:    "message_name",
        SortOrder: "ASC",
    })
    
    if err == nil && response.Success {
        activeCount := 0
        for _, sub := range response.Subscriptions {
            if sub.IsActive {
                activeCount++
                fmt.Printf("✅ %s → %s (v%d)\n", 
                           sub.MessageName, sub.ProcessDefinitionKey, sub.ProcessVersion)
            }
        }
        fmt.Printf("Активных подписок: %d из %d\n", activeCount, len(response.Subscriptions))
    }
}
```

### Python
```python
import grpc
from datetime import datetime

import messages_pb2
import messages_pb2_grpc

def list_message_subscriptions(page=1, page_size=20, sort_by="created_at", sort_order="DESC"):
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = messages_pb2.ListMessageSubscriptionsRequest(
        page=page,
        page_size=page_size,
        sort_by=sort_by,
        sort_order=sort_order
    )
    
    try:
        response = stub.ListMessageSubscriptions(request, metadata=metadata)
        
        if response.success:
            print(f"📋 Найдено {response.total_count} подписок на сообщения")
            
            active_count = 0
            for sub in response.subscriptions:
                status = "🟢" if sub.is_active else "🔴"
                if sub.is_active:
                    active_count += 1
                
                created_at = datetime.fromtimestamp(sub.created_at)
                
                print(f"{status} {sub.message_name} [{sub.process_definition_key}]")
                print(f"   Версия: {sub.process_version} | Событие: {sub.start_event_id}")
                print(f"   Создано: {created_at.strftime('%Y-%m-%d %H:%M:%S')}")
                
                if sub.correlation_key:
                    print(f"   Корреляция: {sub.correlation_key}")
                print()
            
            print(f"Активных: {active_count} из {len(response.subscriptions)}")
            return response.subscriptions
        else:
            print(f"❌ Ошибка: {response.message}")
            return []
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return []

def get_active_subscriptions():
    """Возвращает только активные подписки"""
    all_subs = list_message_subscriptions(page_size=100)
    return [sub for sub in all_subs if sub.is_active]

# Пример использования
if __name__ == "__main__":
    # Все подписки
    subscriptions = list_message_subscriptions()
    
    # Только активные
    active_subs = get_active_subscriptions()
    print(f"Активных подписок: {len(active_subs)}")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'messages.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const messagesProto = grpc.loadPackageDefinition(packageDefinition).messages;

async function listMessageSubscriptions(page = 1, pageSize = 20, sortBy = "created_at", sortOrder = "DESC") {
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
        
        client.listMessageSubscriptions(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`📋 Найдено ${response.total_count} подписок на сообщения`);
                
                let activeCount = 0;
                response.subscriptions.forEach(sub => {
                    const status = sub.is_active ? "🟢" : "🔴";
                    if (sub.is_active) activeCount++;
                    
                    const createdAt = new Date(sub.created_at * 1000);
                    
                    console.log(`${status} ${sub.message_name} [${sub.process_definition_key}]`);
                    console.log(`   Версия: ${sub.process_version} | Событие: ${sub.start_event_id}`);
                    console.log(`   Создано: ${createdAt.toLocaleString()}`);
                    
                    if (sub.correlation_key) {
                        console.log(`   Корреляция: ${sub.correlation_key}`);
                    }
                    console.log();
                });
                
                console.log(`Активных: ${activeCount} из ${response.subscriptions.length}`);
                resolve(response.subscriptions);
            } else {
                console.log(`❌ Ошибка: ${response.message}`);
                resolve([]);
            }
        });
    });
}

async function getActiveSubscriptions() {
    const allSubs = await listMessageSubscriptions(1, 100);
    return allSubs.filter(sub => sub.is_active);
}

// Примеры использования
async function examples() {
    // Все подписки
    const subscriptions = await listMessageSubscriptions();
    
    // Только активные
    const activeSubs = await getActiveSubscriptions();
    console.log(`Активных подписок: ${activeSubs.length}`);
    
    // Сортировка по имени сообщения
    const sortedByName = await listMessageSubscriptions(1, 20, "message_name", "ASC");
}

examples().catch(console.error);
```

## Поля сортировки

### Доступные поля
- **created_at**: Время создания (по умолчанию)
- **updated_at**: Время обновления
- **message_name**: Имя сообщения
- **process_definition_key**: Ключ процесса
- **process_version**: Версия процесса

## Статусы подписок

### is_active
- **true**: Подписка активна, может принимать сообщения
- **false**: Подписка неактивна (процесс удален или деактивирован)

## Типы подписок

### Start Message Events
Подписки для запуска новых экземпляров процессов

### Intermediate Catch Message Events
Подписки для корреляции с ожидающими токенами в процессах

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры пагинации
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

## Связанные методы
- [PublishMessage](publish-message.md) - Публикация сообщений
- [ListBufferedMessages](list-buffered-messages.md) - Буферизованные сообщения
- [GetMessageStats](get-message-stats.md) - Статистика сообщений
