# CleanupExpiredMessages

## Описание
Очищает истекшие сообщения из системы. Удаляет сообщения, у которых истек TTL (Time To Live). Операция освобождает место и улучшает производительность.

## Синтаксис
```protobuf
rpc CleanupExpiredMessages(CleanupExpiredMessagesRequest) returns (CleanupExpiredMessagesResponse);
```

## Package
```protobuf
package messages;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `messages` или `*`

## Параметры запроса

### CleanupExpiredMessagesRequest
```protobuf
message CleanupExpiredMessagesRequest {
  string tenant_id = 1;    // ID тенанта
}
```

## Параметры ответа

### CleanupExpiredMessagesResponse
```protobuf
message CleanupExpiredMessagesResponse {
  int32 cleaned_count = 1;  // Количество удаленных сообщений
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
    
    // Очистка истекших сообщений
    response, err := client.CleanupExpiredMessages(ctx, &pb.CleanupExpiredMessagesRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("🧹 Очистка завершена\n")
        fmt.Printf("   Удалено сообщений: %d\n", response.CleanedCount)
        
        if response.CleanedCount > 0 {
            fmt.Printf("   ✅ Система очищена от истекших сообщений\n")
        } else {
            fmt.Printf("   ℹ️ Истекшие сообщения не найдены\n")
        }
    } else {
        fmt.Printf("❌ Ошибка очистки: %s\n", response.Message)
    }
}

// Автоматическая очистка по расписанию
func scheduleCleanup() {
    // ... client setup ...
    
    ticker := time.NewTicker(1 * time.Hour) // Каждый час
    defer ticker.Stop()
    
    fmt.Printf("⏰ Автоматическая очистка запущена (каждый час)\n")
    
    for range ticker.C {
        response, err := client.CleanupExpiredMessages(ctx, &pb.CleanupExpiredMessagesRequest{})
        if err != nil {
            fmt.Printf("❌ Ошибка автоочистки: %v\n", err)
            continue
        }
        
        if response.Success && response.CleanedCount > 0 {
            now := time.Now().Format("2006-01-02 15:04:05")
            fmt.Printf("[%s] 🧹 Автоочистка: удалено %d сообщений\n", 
                       now, response.CleanedCount)
        }
    }
}

// Очистка с предварительной проверкой
func cleanupWithStats() {
    // ... client setup ...
    
    // Получаем статистику перед очисткой
    statsResponse, err := client.GetMessageStats(ctx, &pb.GetMessageStatsRequest{})
    if err != nil {
        fmt.Printf("❌ Ошибка получения статистики: %v\n", err)
        return
    }
    
    if !statsResponse.Success {
        fmt.Printf("❌ Ошибка статистики: %s\n", statsResponse.Message)
        return
    }
    
    beforeCount := statsResponse.Stats.ExpiredMessages
    fmt.Printf("📊 Истекших сообщений перед очисткой: %d\n", beforeCount)
    
    if beforeCount == 0 {
        fmt.Printf("ℹ️ Нет сообщений для очистки\n")
        return
    }
    
    // Выполняем очистку
    cleanupResponse, err := client.CleanupExpiredMessages(ctx, &pb.CleanupExpiredMessagesRequest{})
    if err != nil {
        fmt.Printf("❌ Ошибка очистки: %v\n", err)
        return
    }
    
    if cleanupResponse.Success {
        fmt.Printf("✅ Очистка завершена: удалено %d сообщений\n", cleanupResponse.CleanedCount)
        
        // Проверяем результат
        if cleanupResponse.CleanedCount != beforeCount {
            fmt.Printf("⚠️ Удалено %d из %d ожидаемых\n", 
                       cleanupResponse.CleanedCount, beforeCount)
        }
    }
}
```

### Python
```python
import grpc
import time
import threading
from datetime import datetime

import messages_pb2
import messages_pb2_grpc

def cleanup_expired_messages():
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = messages_pb2.CleanupExpiredMessagesRequest()
    
    try:
        response = stub.CleanupExpiredMessages(request, metadata=metadata)
        
        if response.success:
            print("🧹 Очистка завершена")
            print(f"   Удалено сообщений: {response.cleaned_count}")
            
            if response.cleaned_count > 0:
                print("   ✅ Система очищена от истекших сообщений")
            else:
                print("   ℹ️ Истекшие сообщения не найдены")
            
            return response.cleaned_count
        else:
            print(f"❌ Ошибка очистки: {response.message}")
            return 0
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return 0

def cleanup_with_stats():
    """Очистка с предварительной проверкой статистики"""
    channel = grpc.insecure_channel('localhost:27500')
    stub = messages_pb2_grpc.MessagesServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Получаем статистику перед очисткой
    try:
        stats_response = stub.GetMessageStats(
            messages_pb2.GetMessageStatsRequest(), 
            metadata=metadata
        )
        
        if stats_response.success:
            before_count = stats_response.stats.expired_messages
            print(f"📊 Истекших сообщений перед очисткой: {before_count}")
            
            if before_count == 0:
                print("ℹ️ Нет сообщений для очистки")
                return 0
        else:
            print(f"❌ Ошибка статистики: {stats_response.message}")
            
    except grpc.RpcError as e:
        print(f"Ошибка получения статистики: {e.details()}")
        before_count = None
    
    # Выполняем очистку
    cleaned_count = cleanup_expired_messages()
    
    # Проверяем результат
    if before_count is not None and cleaned_count != before_count:
        print(f"⚠️ Удалено {cleaned_count} из {before_count} ожидаемых")
    
    return cleaned_count

class MessageCleaner:
    def __init__(self, interval_hours=1):
        self.interval_hours = interval_hours
        self.running = False
        self.thread = None
    
    def start(self):
        """Запускает автоматическую очистку"""
        if self.running:
            return
        
        self.running = True
        self.thread = threading.Thread(target=self._run_cleanup)
        self.thread.daemon = True
        self.thread.start()
        
        print(f"⏰ Автоматическая очистка запущена (каждые {self.interval_hours}ч)")
    
    def stop(self):
        """Останавливает автоматическую очистку"""
        self.running = False
        if self.thread:
            self.thread.join(timeout=5)
        print("⏰ Автоматическая очистка остановлена")
    
    def _run_cleanup(self):
        """Основной цикл очистки"""
        while self.running:
            try:
                cleaned = cleanup_expired_messages()
                if cleaned > 0:
                    now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                    print(f"[{now}] 🧹 Автоочистка: удалено {cleaned} сообщений")
            except Exception as e:
                print(f"❌ Ошибка автоочистки: {e}")
            
            # Ждем до следующей очистки
            time.sleep(self.interval_hours * 3600)

# Пример использования
if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1:
        command = sys.argv[1]
        
        if command == "auto":
            # Автоматическая очистка
            interval = int(sys.argv[2]) if len(sys.argv) > 2 else 1
            cleaner = MessageCleaner(interval)
            
            try:
                cleaner.start()
                print("Нажмите Ctrl+C для остановки")
                while True:
                    time.sleep(1)
            except KeyboardInterrupt:
                cleaner.stop()
                
        elif command == "stats":
            # Очистка со статистикой
            cleanup_with_stats()
        else:
            print("Неизвестная команда. Используйте: auto, stats")
    else:
        # Обычная очистка
        cleanup_expired_messages()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'messages.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const messagesProto = grpc.loadPackageDefinition(packageDefinition).messages;

async function cleanupExpiredMessages() {
    const client = new messagesProto.MessagesService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.cleanupExpiredMessages(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log("🧹 Очистка завершена");
                console.log(`   Удалено сообщений: ${response.cleaned_count}`);
                
                if (response.cleaned_count > 0) {
                    console.log("   ✅ Система очищена от истекших сообщений");
                } else {
                    console.log("   ℹ️ Истекшие сообщения не найдены");
                }
                
                resolve(response.cleaned_count);
            } else {
                console.log(`❌ Ошибка очистки: ${response.message}`);
                resolve(0);
            }
        });
    });
}

async function cleanupWithStats() {
    const client = new messagesProto.MessagesService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    // Получаем статистику перед очисткой
    try {
        const statsResponse = await new Promise((resolve, reject) => {
            client.getMessageStats({}, metadata, (error, response) => {
                if (error) reject(error);
                else resolve(response);
            });
        });
        
        if (statsResponse.success) {
            const beforeCount = statsResponse.stats.expired_messages;
            console.log(`📊 Истекших сообщений перед очисткой: ${beforeCount}`);
            
            if (beforeCount === 0) {
                console.log("ℹ️ Нет сообщений для очистки");
                return 0;
            }
            
            // Выполняем очистку
            const cleanedCount = await cleanupExpiredMessages();
            
            // Проверяем результат
            if (cleanedCount !== beforeCount) {
                console.log(`⚠️ Удалено ${cleanedCount} из ${beforeCount} ожидаемых`);
            }
            
            return cleanedCount;
        } else {
            console.log(`❌ Ошибка статистики: ${statsResponse.message}`);
        }
    } catch (error) {
        console.log(`Ошибка получения статистики: ${error.message}`);
    }
    
    // Выполняем очистку без статистики
    return await cleanupExpiredMessages();
}

class MessageCleaner {
    constructor(intervalHours = 1) {
        this.intervalHours = intervalHours;
        this.running = false;
        this.intervalId = null;
    }
    
    start() {
        if (this.running) return;
        
        this.running = true;
        console.log(`⏰ Автоматическая очистка запущена (каждые ${this.intervalHours}ч)`);
        
        this.intervalId = setInterval(async () => {
            try {
                const cleaned = await cleanupExpiredMessages();
                if (cleaned > 0) {
                    const now = new Date().toLocaleString();
                    console.log(`[${now}] 🧹 Автоочистка: удалено ${cleaned} сообщений`);
                }
            } catch (error) {
                console.log(`❌ Ошибка автоочистки: ${error.message}`);
            }
        }, this.intervalHours * 3600 * 1000);
    }
    
    stop() {
        if (!this.running) return;
        
        this.running = false;
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
        console.log("⏰ Автоматическая очистка остановлена");
    }
}

// Примеры использования
async function examples() {
    // Обычная очистка
    await cleanupExpiredMessages();
    
    // Очистка со статистикой
    await cleanupWithStats();
}

if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args[0] === 'auto') {
        // Автоматическая очистка
        const interval = args[1] ? parseInt(args[1]) : 1;
        const cleaner = new MessageCleaner(interval);
        
        cleaner.start();
        console.log("Нажмите Ctrl+C для остановки");
        
        process.on('SIGINT', () => {
            cleaner.stop();
            process.exit(0);
        });
        
    } else if (args[0] === 'stats') {
        // Очистка со статистикой
        cleanupWithStats().catch(console.error);
        
    } else {
        // Обычная очистка
        examples().catch(console.error);
    }
}

module.exports = { cleanupExpiredMessages, cleanupWithStats, MessageCleaner };
```

## Автоматическая очистка

### Рекомендуемое расписание
- **Каждый час**: Для систем с высокой нагрузкой
- **Каждые 6 часов**: Для обычных систем
- **Ежедневно**: Для систем с низкой нагрузкой

### Мониторинг очистки
Рекомендуется логировать результаты очистки для мониторинга состояния системы.

## Влияние на производительность

### Преимущества очистки
- Освобождение места в базе данных
- Улучшение производительности запросов
- Снижение использования памяти

### Оптимальное время
Выполняйте очистку в периоды низкой нагрузки для минимального влияния на производительность.

## Возможные ошибки

### gRPC Status Codes
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Ошибка базы данных

### Примеры ошибок
```json
{
  "cleaned_count": 0,
  "success": false,
  "message": "Database cleanup operation failed"
}
```

## Связанные методы
- [GetMessageStats](get-message-stats.md) - Проверка количества истекших сообщений
- [ListBufferedMessages](list-buffered-messages.md) - Просмотр сообщений перед очисткой
- [PublishMessage](publish-message.md) - Настройка TTL для новых сообщений
