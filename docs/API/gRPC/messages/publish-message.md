# PublishMessage

## Описание
Публикует сообщение в системе для корреляции с процессами BPMN. Сообщения могут запускать новые процессы или коррелироваться с ожидающими элементами.

## Синтаксис
```protobuf
rpc PublishMessage(PublishMessageRequest) returns (PublishMessageResponse);
```

## Package
```protobuf
package messages;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `messages` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### PublishMessageRequest
```protobuf
message PublishMessageRequest {
  string tenant_id = 1;              // ID тенанта
  string message_name = 2;           // Имя сообщения
  string correlation_key = 3;        // Ключ корреляции
  map<string, string> variables = 4; // Переменные сообщения
  int64 ttl_seconds = 5;             // Время жизни в секундах
}
```

#### Поля:
- **tenant_id** (string, optional): ID тенанта для мультитенантности
- **message_name** (string, required): Имя сообщения для корреляции
- **correlation_key** (string, optional): Ключ для корреляции с процессами
- **variables** (map, optional): Переменные, передаваемые в процесс
- **ttl_seconds** (int64, optional): Время жизни сообщения (по умолчанию 3600)

## Параметры ответа

### PublishMessageResponse
```protobuf
message PublishMessageResponse {
  string message_id = 1;    // ID созданного сообщения
  bool success = 2;         // Статус успешности
  string message = 3;       // Сообщение о результате
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
    
    // Публикация простого сообщения
    response, err := client.PublishMessage(ctx, &pb.PublishMessageRequest{
        MessageName: "order_created",
        Variables: map[string]string{
            "orderId": "12345",
            "amount":  "100.50",
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Сообщение опубликовано: %s\n", response.MessageId)
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.Message)
    }
}

// Публикация с ключом корреляции
func publishWithCorrelation() {
    // ... client setup ...
    
    response, err := client.PublishMessage(ctx, &pb.PublishMessageRequest{
        MessageName:    "payment_completed",
        CorrelationKey: "order-12345",
        Variables: map[string]string{
            "status":    "success",
            "paymentId": "pay-67890",
        },
        TtlSeconds: 1800, // 30 минут
    })
    
    // ... обработка ответа ...
}
```

### Python
```python
import grpc

import messages_pb2
import messages_pb2_grpc

def publish_message(message_name, variables=None, correlation_key="", ttl=3600):
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = messages_pb2.PublishMessageRequest(
        message_name=message_name,
        correlation_key=correlation_key,
        variables=variables or {},
        ttl_seconds=ttl
    )
    
    try:
        response = stub.PublishMessage(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Сообщение опубликовано: {response.message_id}")
            return response.message_id
        else:
            print(f"❌ Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Пример использования
if __name__ == "__main__":
    # Простое сообщение
    publish_message("user_registered", {
        "userId": "123",
        "email": "user@example.com"
    })
    
    # С корреляцией
    publish_message("order_updated", {
        "status": "shipped",
        "trackingNumber": "TRK123"
    }, correlation_key="order-456")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'messages.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const messagesProto = grpc.loadPackageDefinition(packageDefinition).messages;

async function publishMessage(messageName, variables = {}, correlationKey = "", ttl = 3600) {
    const client = new messagesProto.MessagesService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            message_name: messageName,
            correlation_key: correlationKey,
            variables: variables,
            ttl_seconds: ttl
        };
        
        client.publishMessage(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Сообщение опубликовано: ${response.message_id}`);
                resolve(response.message_id);
            } else {
                console.log(`❌ Ошибка: ${response.message}`);
                resolve(null);
            }
        });
    });
}

// Примеры использования
async function examples() {
    // Простое сообщение
    await publishMessage("file_uploaded", {
        fileName: "document.pdf",
        size: "1024000"
    });
    
    // С корреляцией и TTL
    await publishMessage("task_completed", {
        taskId: "task-789",
        result: "success"
    }, "process-instance-123", 7200); // 2 часа
}

examples().catch(console.error);
```

## Корреляция сообщений

### Запуск процессов
Сообщения могут запускать новые экземпляры процессов, если есть подходящие Start Message Events.

### Корреляция с ожидающими элементами
Сообщения коррелируются с Intermediate Catch Message Events по имени сообщения и ключу корреляции.

## TTL (Time To Live)
Сообщения автоматически удаляются после истечения TTL. По умолчанию: 1 час.

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Пустое имя сообщения
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "message_id": "",
  "success": false,
  "message": "Message name cannot be empty"
}
```

## Связанные методы
- [ListBufferedMessages](list-buffered-messages.md) - Просмотр буферизованных сообщений
- [ListMessageSubscriptions](list-message-subscriptions.md) - Просмотр подписок
- [GetMessageStats](get-message-stats.md) - Статистика сообщений
