# AddTimer

## Описание
Добавляет новый таймер в иерархическую систему timewheel. Поддерживает как одноразовые, так и повторяющиеся таймеры с ISO 8601 форматом длительности.

## Синтаксис
```protobuf
rpc AddTimer(AddTimerRequest) returns (AddTimerResponse);
```

## Package
```protobuf
package atom.timewheel.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer` или `*`

## Параметры запроса

### AddTimerRequest
```protobuf
message AddTimerRequest {
  string timer_id = 1;      // Уникальный ID таймера
  int64 delay_ms = 2;       // ⚠️ Устарело: используйте duration
  string callback_data = 3;  // JSON данные для callback
  bool repeating = 4;       // ⚠️ Устарело: используйте interval
  int64 interval_ms = 5;    // ⚠️ Устарело: используйте interval
  string duration = 6;      // ISO 8601 длительность (PT30S, PT1H, P1D)
  string interval = 7;      // ISO 8601 интервал повтора (R5/PT30S, R/PT1M)
}
```

## Параметры ответа

### AddTimerResponse
```protobuf
message AddTimerResponse {
  string timer_id = 1;      // ID созданного таймера
  bool success = 2;         // Успешность создания
  string message = 3;       // Сообщение о результате
  int64 scheduled_at = 4;   // Unix timestamp запланированного срабатывания
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
    
    pb "atom-engine/proto/timewheel/timewheelpb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewTimeWheelServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Одноразовый таймер на 30 секунд
    response, err := client.AddTimer(ctx, &pb.AddTimerRequest{
        TimerId:      "timer-simple-30s",
        Duration:     "PT30S",
        CallbackData: `{"type": "notification", "message": "30 seconds elapsed"}`,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Таймер создан: %s\n", response.TimerId)
        fmt.Printf("📅 Сработает в: %s\n", 
            time.Unix(response.ScheduledAt, 0).Format("15:04:05"))
        fmt.Printf("💬 Сообщение: %s\n", response.Message)
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.Message)
    }
    
    // Повторяющийся таймер: 5 раз каждые 10 секунд
    response2, err := client.AddTimer(ctx, &pb.AddTimerRequest{
        TimerId:      "timer-repeat-5x10s",
        Interval:     "R5/PT10S",
        CallbackData: `{"type": "heartbeat", "counter": 0}`,
    })
    
    if err == nil && response2.Success {
        fmt.Printf("🔄 Повторяющийся таймер создан: %s\n", response2.TimerId)
        fmt.Printf("   Повторов: 5, интервал: 10 секунд\n")
    }
    
    // Бесконечно повторяющийся таймер каждую минуту
    response3, err := client.AddTimer(ctx, &pb.AddTimerRequest{
        TimerId:      "timer-infinite-1m",
        Interval:     "R/PT1M",
        CallbackData: `{"type": "monitoring", "service": "health-check"}`,
    })
    
    if err == nil && response3.Success {
        fmt.Printf("♾️  Бесконечный таймер создан: %s\n", response3.TimerId)
    }
}
```

### Python
```python
import grpc
import json
from datetime import datetime, timedelta

import timewheel_pb2
import timewheel_pb2_grpc

def add_timer(timer_id, duration=None, interval=None, callback_data=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = timewheel_pb2.AddTimerRequest(
        timer_id=timer_id,
        callback_data=json.dumps(callback_data or {})
    )
    
    # Устанавливаем duration или interval
    if interval:
        request.interval = interval
    elif duration:
        request.duration = duration
    else:
        raise ValueError("Необходимо указать duration или interval")
    
    try:
        response = stub.AddTimer(request, metadata=metadata)
        
        if response.success:
            scheduled_time = datetime.fromtimestamp(response.scheduled_at)
            print(f"✅ Таймер '{timer_id}' создан")
            print(f"📅 Запланирован на: {scheduled_time.strftime('%H:%M:%S')}")
            print(f"💬 {response.message}")
            return response.timer_id
        else:
            print(f"❌ Ошибка создания таймера: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Создание различных типов таймеров
def create_sample_timers():
    print("⏰ Создание примеров таймеров\n")
    
    # 1. Простой таймер на 5 минут
    add_timer(
        "meeting-reminder", 
        duration="PT5M",
        callback_data={"type": "reminder", "message": "Meeting in 5 minutes"}
    )
    
    # 2. Уведомления каждые 30 секунд, 10 раз
    add_timer(
        "status-updates",
        interval="R10/PT30S", 
        callback_data={"type": "status", "service": "api-monitor"}
    )
    
    # 3. Ежечасная очистка кэша
    add_timer(
        "cache-cleanup",
        interval="R/PT1H",
        callback_data={"type": "maintenance", "action": "clear-cache"}
    )
    
    # 4. Таймер на завтра в 9:00 (24 часа)
    add_timer(
        "daily-report",
        duration="P1D",
        callback_data={"type": "report", "frequency": "daily"}
    )
    
    # 5. Граничные события BPMN (30 секунд таймаут)
    add_timer(
        "bpmn-timeout-activity-123",
        duration="PT30S",
        callback_data={
            "type": "boundary_event",
            "process_instance_id": "proc-456",
            "activity_id": "activity-123",
            "event_type": "timeout"
        }
    )

# Вспомогательные функции для работы с ISO 8601
class ISO8601Helper:
    @staticmethod
    def minutes_to_iso(minutes):
        return f"PT{minutes}M"
    
    @staticmethod
    def hours_to_iso(hours):
        return f"PT{hours}H"
    
    @staticmethod
    def days_to_iso(days):
        return f"P{days}D"
    
    @staticmethod
    def repeating_seconds(count, seconds):
        return f"R{count}/PT{seconds}S"
    
    @staticmethod
    def infinite_minutes(minutes):
        return f"R/PT{minutes}M"

# Демонстрация с помощником ISO 8601
def demo_with_iso_helper():
    print("🔧 Использование ISO 8601 Helper\n")
    
    helper = ISO8601Helper()
    
    # Таймер на 15 минут
    add_timer(
        "break-timer",
        duration=helper.minutes_to_iso(15),
        callback_data={"message": "Break time is over!"}
    )
    
    # 3 напоминания каждые 2 часа 
    add_timer(
        "medication-reminder", 
        interval=helper.repeating_seconds(3, 7200),  # 3 раза по 2 часа в секундах
        callback_data={"type": "health", "action": "take_medication"}
    )
    
    # Еженедельный бэкап (каждые 7 дней, бесконечно)
    add_timer(
        "weekly-backup",
        interval=f"R/{helper.days_to_iso(7)}",
        callback_data={"type": "backup", "frequency": "weekly"}
    )

if __name__ == "__main__":
    create_sample_timers()
    print("-" * 50)
    demo_with_iso_helper()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'timewheel.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const timewheelProto = grpc.loadPackageDefinition(packageDefinition).atom.timewheel.v1;

async function addTimer(timerId, options = {}) {
    const client = new timewheelProto.TimeWheelService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            timer_id: timerId,
            callback_data: JSON.stringify(options.callbackData || {})
        };
        
        // Устанавливаем duration или interval
        if (options.interval) {
            request.interval = options.interval;
        } else if (options.duration) {
            request.duration = options.duration;
        } else {
            reject(new Error('Необходимо указать duration или interval'));
            return;
        }
        
        client.addTimer(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                const scheduledTime = new Date(response.scheduled_at * 1000);
                console.log(`✅ Таймер '${timerId}' создан`);
                console.log(`📅 Запланирован на: ${scheduledTime.toLocaleTimeString()}`);
                console.log(`💬 ${response.message}`);
                resolve(response);
            } else {
                console.log(`❌ Ошибка: ${response.message}`);
                resolve(null);
            }
        });
    });
}

// Создание BPMN таймеров
async function createBPMNTimers() {
    console.log('🔄 Создание BPMN таймеров\n');
    
    // Timer Start Event - запуск процесса каждые 10 минут
    await addTimer('bpmn-start-timer-daily-report', {
        interval: 'R/PT10M',
        callbackData: {
            type: 'start_event',
            process_definition_key: 'daily-report-process',
            trigger_type: 'timer'
        }
    });
    
    // Boundary Timer Event - таймаут для активности (2 минуты)
    await addTimer('bpmn-boundary-user-task-timeout', {
        duration: 'PT2M',
        callbackData: {
            type: 'boundary_event',
            process_instance_id: 'pi-12345',
            activity_id: 'user-task-approval',
            interrupting: true
        }
    });
    
    // Intermediate Timer Event - пауза в процессе (30 секунд)
    await addTimer('bpmn-intermediate-wait', {
        duration: 'PT30S',
        callbackData: {
            type: 'intermediate_event',
            process_instance_id: 'pi-12345',
            element_id: 'wait-event-1'
        }
    });
    
    console.log();
}

// Система напоминаний
class ReminderSystem {
    constructor() {
        this.reminders = new Map();
    }
    
    async addReminder(name, when, message, repeat = null) {
        const timerId = `reminder-${name}-${Date.now()}`;
        
        const callbackData = {
            type: 'reminder',
            name: name,
            message: message,
            created_at: new Date().toISOString()
        };
        
        const options = { callbackData };
        
        if (repeat) {
            options.interval = repeat;
        } else {
            options.duration = when;
        }
        
        try {
            const result = await addTimer(timerId, options);
            
            if (result) {
                this.reminders.set(name, {
                    timerId: result.timer_id,
                    message: message,
                    scheduledAt: result.scheduled_at
                });
                
                console.log(`📝 Напоминание "${name}" установлено`);
                return timerId;
            }
        } catch (error) {
            console.log(`❌ Ошибка создания напоминания: ${error.message}`);
        }
        
        return null;
    }
    
    listReminders() {
        console.log('📋 Активные напоминания:');
        
        for (const [name, data] of this.reminders) {
            const time = new Date(data.scheduledAt * 1000);
            console.log(`  • ${name}: "${data.message}" в ${time.toLocaleString()}`);
        }
    }
}

// Демонстрация системы напоминаний
async function demonstrateReminderSystem() {
    console.log('📱 Система напоминаний\n');
    
    const reminders = new ReminderSystem();
    
    // Разовые напоминания
    await reminders.addReminder(
        'встреча',
        'PT15M',
        'Встреча с командой через 15 минут'
    );
    
    await reminders.addReminder(
        'обед',
        'PT1H',
        'Время обеда!'
    );
    
    // Повторяющиеся напоминания
    await reminders.addReminder(
        'вода',
        'R8/PT30M', // 8 раз каждые 30 минут
        'Не забудьте выпить воды'
    );
    
    await reminders.addReminder(
        'поза',
        'R/PT45M', // Каждые 45 минут, бесконечно
        'Время размяться и изменить позу'
    );
    
    console.log();
    reminders.listReminders();
}

// Главная демонстрация
async function main() {
    try {
        await createBPMNTimers();
        console.log('='.repeat(50));
        await demonstrateReminderSystem();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## ISO 8601 Форматы

### Duration (длительность)
```
PT30S         - 30 секунд
PT5M          - 5 минут  
PT1H          - 1 час
PT2H30M       - 2 часа 30 минут
P1D           - 1 день
P1DT12H       - 1 день 12 часов
P1W           - 1 неделя
P1M           - 1 месяц
P1Y           - 1 год
```

### Interval (повторяющийся интервал)
```
R5/PT30S      - 5 повторов каждые 30 секунд
R10/PT1M      - 10 повторов каждую минуту
R/PT15M       - Бесконечно каждые 15 минут
R3/PT2H       - 3 повтора каждые 2 часа
R/P1D         - Ежедневно, бесконечно
```

## Timewheel Уровни

### Иерархическая структура
- **Уровень 0**: Секунды (0-59)
- **Уровень 1**: Минуты (0-59)  
- **Уровень 2**: Часы (0-23)
- **Уровень 3**: Дни (0-30)
- **Уровень 4**: Годы (0-99)

### Производительность
- **O(1)** операции добавления/удаления
- **Масштабируемость** до 100+ лет
- **Точность** до секунды

## Применение в BPMN

### Timer Start Events
```javascript
// Запуск процесса каждый час
await addTimer('process-hourly', {
    interval: 'R/PT1H',
    callbackData: { type: 'start_event', process_key: 'hourly-report' }
});
```

### Boundary Timer Events  
```javascript
// Таймаут для пользовательской задачи
await addTimer('task-timeout', {
    duration: 'PT10M',
    callbackData: { 
        type: 'boundary_event', 
        process_instance_id: 'pi-123',
        interrupting: true 
    }
});
```

### Intermediate Timer Events
```javascript
// Пауза в процессе
await addTimer('process-delay', {
    duration: 'PT30S',
    callbackData: { type: 'intermediate_event', element_id: 'timer-1' }
});
```

## Связанные методы
- [RemoveTimer](remove-timer.md) - Удаление таймера
- [GetTimerStatus](get-timer-status.md) - Проверка статуса
- [ListTimers](list-timers.md) - Список всех таймеров
