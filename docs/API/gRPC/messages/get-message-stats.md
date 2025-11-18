# GetMessageStats

## Описание
Возвращает статистику по сообщениям в системе, включая общее количество, буферизованные сообщения и активность за день.

## Синтаксис
```protobuf
rpc GetMessageStats(GetMessageStatsRequest) returns (GetMessageStatsResponse);
```

## Package
```protobuf
package messages;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `messages` или `*`

## Параметры запроса

### GetMessageStatsRequest
```protobuf
message GetMessageStatsRequest {
  string tenant_id = 1;    // ID тенанта
}
```

## Параметры ответа

### GetMessageStatsResponse
```protobuf
message GetMessageStatsResponse {
  MessageStats stats = 1;  // Статистика сообщений
  bool success = 2;        // Статус успешности
  string message = 3;      // Сообщение о результате
}

message MessageStats {
  int32 total_messages = 1;           // Общее количество сообщений
  int32 buffered_messages = 2;        // Буферизованные сообщения
  int32 expired_messages = 3;         // Истекшие сообщения
  int32 published_today = 4;          // Опубликовано сегодня
  int32 instances_created_today = 5;  // Экземпляров создано сегодня
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
    
    // Получение статистики
    response, err := client.GetMessageStats(ctx, &pb.GetMessageStatsRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        stats := response.Stats
        
        fmt.Printf("📊 Статистика сообщений:\n")
        fmt.Printf("   Всего сообщений: %d\n", stats.TotalMessages)
        fmt.Printf("   Буферизованных: %d\n", stats.BufferedMessages)
        fmt.Printf("   Истекших: %d\n", stats.ExpiredMessages)
        fmt.Printf("   Опубликовано сегодня: %d\n", stats.PublishedToday)
        fmt.Printf("   Экземпляров создано сегодня: %d\n", stats.InstancesCreatedToday)
        
        // Анализ эффективности
        if stats.TotalMessages > 0 {
            bufferedRate := float64(stats.BufferedMessages) / float64(stats.TotalMessages) * 100
            fmt.Printf("   Процент буферизации: %.1f%%\n", bufferedRate)
            
            if bufferedRate > 20 {
                fmt.Printf("   ⚠️ Высокий процент буферизации\n")
            }
        }
        
        if stats.PublishedToday > 0 {
            successRate := float64(stats.InstancesCreatedToday) / float64(stats.PublishedToday) * 100
            fmt.Printf("   Успешность корреляции: %.1f%%\n", successRate)
        }
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.Message)
    }
}

// Мониторинг статистики
func monitorStats() {
    // ... client setup ...
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    fmt.Printf("📊 Мониторинг статистики сообщений (каждые 30с)\n")
    
    for range ticker.C {
        response, err := client.GetMessageStats(ctx, &pb.GetMessageStatsRequest{})
        if err != nil {
            fmt.Printf("❌ Ошибка получения статистики: %v\n", err)
            continue
        }
        
        if response.Success {
            stats := response.Stats
            now := time.Now().Format("15:04:05")
            
            fmt.Printf("[%s] Всего: %d | Буферизовано: %d | Сегодня: %d\n",
                       now, stats.TotalMessages, stats.BufferedMessages, stats.PublishedToday)
        }
    }
}
```

### Python
```python
import grpc
import time

import messages_pb2
import messages_pb2_grpc

def get_message_stats():
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = messages_pb2.GetMessageStatsRequest()
    
    try:
        response = stub.GetMessageStats(request, metadata=metadata)
        
        if response.success:
            stats = response.stats
            
            print("📊 Статистика сообщений:")
            print(f"   Всего сообщений: {stats.total_messages}")
            print(f"   Буферизованных: {stats.buffered_messages}")
            print(f"   Истекших: {stats.expired_messages}")
            print(f"   Опубликовано сегодня: {stats.published_today}")
            print(f"   Экземпляров создано сегодня: {stats.instances_created_today}")
            
            # Анализ эффективности
            if stats.total_messages > 0:
                buffered_rate = (stats.buffered_messages / stats.total_messages) * 100
                print(f"   Процент буферизации: {buffered_rate:.1f}%")
                
                if buffered_rate > 20:
                    print("   ⚠️ Высокий процент буферизации")
            
            if stats.published_today > 0:
                success_rate = (stats.instances_created_today / stats.published_today) * 100
                print(f"   Успешность корреляции: {success_rate:.1f}%")
            
            return stats
        else:
            print(f"❌ Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def monitor_stats(interval=30):
    """Мониторинг статистики в реальном времени"""
    print(f"📊 Мониторинг статистики сообщений (каждые {interval}с)")
    print("Нажмите Ctrl+C для остановки")
    
    try:
        while True:
            stats = get_message_stats()
            if stats:
                now = time.strftime("%H:%M:%S")
                print(f"[{now}] Всего: {stats.total_messages} | "
                      f"Буферизовано: {stats.buffered_messages} | "
                      f"Сегодня: {stats.published_today}")
            
            time.sleep(interval)
    except KeyboardInterrupt:
        print("\n🛑 Мониторинг остановлен")

# Пример использования
if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1 and sys.argv[1] == "monitor":
        interval = int(sys.argv[2]) if len(sys.argv) > 2 else 30
        monitor_stats(interval)
    else:
        get_message_stats()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'messages.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const messagesProto = grpc.loadPackageDefinition(packageDefinition).messages;

async function getMessageStats() {
    const client = new messagesProto.MessagesService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.getMessageStats(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                const stats = response.stats;
                
                console.log("📊 Статистика сообщений:");
                console.log(`   Всего сообщений: ${stats.total_messages}`);
                console.log(`   Буферизованных: ${stats.buffered_messages}`);
                console.log(`   Истекших: ${stats.expired_messages}`);
                console.log(`   Опубликовано сегодня: ${stats.published_today}`);
                console.log(`   Экземпляров создано сегодня: ${stats.instances_created_today}`);
                
                // Анализ эффективности
                if (stats.total_messages > 0) {
                    const bufferedRate = (stats.buffered_messages / stats.total_messages) * 100;
                    console.log(`   Процент буферизации: ${bufferedRate.toFixed(1)}%`);
                    
                    if (bufferedRate > 20) {
                        console.log("   ⚠️ Высокий процент буферизации");
                    }
                }
                
                if (stats.published_today > 0) {
                    const successRate = (stats.instances_created_today / stats.published_today) * 100;
                    console.log(`   Успешность корреляции: ${successRate.toFixed(1)}%`);
                }
                
                resolve(stats);
            } else {
                console.log(`❌ Ошибка: ${response.message}`);
                resolve(null);
            }
        });
    });
}

async function monitorStats(interval = 30000) {
    console.log(`📊 Мониторинг статистики сообщений (каждые ${interval/1000}с)`);
    console.log("Нажмите Ctrl+C для остановки");
    
    const monitor = setInterval(async () => {
        try {
            const stats = await getMessageStats();
            if (stats) {
                const now = new Date().toLocaleTimeString();
                console.log(`[${now}] Всего: ${stats.total_messages} | ` +
                           `Буферизовано: ${stats.buffered_messages} | ` +
                           `Сегодня: ${stats.published_today}`);
            }
        } catch (error) {
            console.log(`❌ Ошибка получения статистики: ${error.message}`);
        }
    }, interval);
    
    // Обработка Ctrl+C
    process.on('SIGINT', () => {
        console.log('\n🛑 Мониторинг остановлен');
        clearInterval(monitor);
        process.exit(0);
    });
}

// Примеры использования
async function examples() {
    // Получение текущей статистики
    const stats = await getMessageStats();
    
    // Мониторинг (раскомментировать для запуска)
    // await monitorStats(30000); // каждые 30 секунд
}

if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args[0] === 'monitor') {
        const interval = args[1] ? parseInt(args[1]) * 1000 : 30000;
        monitorStats(interval);
    } else {
        examples().catch(console.error);
    }
}

module.exports = { getMessageStats, monitorStats };
```

## Метрики статистики

### Основные счетчики
- **total_messages**: Общее количество сообщений в системе
- **buffered_messages**: Сообщения, ожидающие корреляции
- **expired_messages**: Истекшие сообщения (TTL)

### Дневная активность
- **published_today**: Сообщения, опубликованные сегодня
- **instances_created_today**: Экземпляры процессов, созданные сегодня

## Анализ эффективности

### Процент буферизации
Показывает, какая часть сообщений не может быть немедленно обработана:
- **< 10%**: Хорошая эффективность
- **10-20%**: Приемлемый уровень
- **> 20%**: Высокий процент буферизации, требует внимания

### Успешность корреляции
Отношение созданных экземпляров к опубликованным сообщениям за день.

## Возможные ошибки

### gRPC Status Codes
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

## Связанные методы
- [ListBufferedMessages](list-buffered-messages.md) - Детали буферизованных сообщений
- [ListMessageSubscriptions](list-message-subscriptions.md) - Активные подписки
- [CleanupExpiredMessages](cleanup-expired-messages.md) - Очистка истекших сообщений
