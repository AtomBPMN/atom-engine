# RemoveTimer

## Описание
Удаляет таймер из системы timewheel по его уникальному ID. Поддерживает удаление как одноразовых, так и повторяющихся таймеров.

## Синтаксис
```protobuf
rpc RemoveTimer(RemoveTimerRequest) returns (RemoveTimerResponse);
```

## Package
```protobuf
package atom.timewheel.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer` или `*`

## Параметры запроса

### RemoveTimerRequest
```protobuf
message RemoveTimerRequest {
  string timer_id = 1;      // ID таймера для удаления
}
```

## Параметры ответа

### RemoveTimerResponse
```protobuf
message RemoveTimerResponse {
  string timer_id = 1;      // ID удаленного таймера
  bool success = 2;         // Успешность удаления
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
    
    // Удаление конкретного таймера
    response, err := client.RemoveTimer(ctx, &pb.RemoveTimerRequest{
        TimerId: "timer-simple-30s",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Таймер удален: %s\n", response.TimerId)
        fmt.Printf("💬 Сообщение: %s\n", response.Message)
    } else {
        fmt.Printf("❌ Ошибка удаления: %s\n", response.Message)
    }
}

// Пример массового удаления таймеров
func removeMultipleTimers(client pb.TimeWheelServiceClient, ctx context.Context, timerIds []string) {
    fmt.Printf("🗑️ Удаление %d таймеров...\n", len(timerIds))
    
    successCount := 0
    failCount := 0
    
    for _, timerId := range timerIds {
        response, err := client.RemoveTimer(ctx, &pb.RemoveTimerRequest{
            TimerId: timerId,
        })
        
        if err != nil {
            fmt.Printf("❌ gRPC ошибка для %s: %v\n", timerId, err)
            failCount++
            continue
        }
        
        if response.Success {
            fmt.Printf("✅ %s - удален\n", timerId)
            successCount++
        } else {
            fmt.Printf("❌ %s - ошибка: %s\n", timerId, response.Message)
            failCount++
        }
    }
    
    fmt.Printf("\n📊 Итого: %d удалено, %d ошибок\n", successCount, failCount)
}
```

### Python
```python
import grpc

import timewheel_pb2
import timewheel_pb2_grpc

def remove_timer(timer_id):
    channel = grpc.insecure_channel('localhost:27500')
    stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = timewheel_pb2.RemoveTimerRequest(timer_id=timer_id)
    
    try:
        response = stub.RemoveTimer(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Таймер '{timer_id}' успешно удален")
            print(f"💬 {response.message}")
            return True
        else:
            print(f"❌ Не удалось удалить '{timer_id}': {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

# Менеджер жизненного цикла таймеров
class TimerManager:
    def __init__(self):
        self.active_timers = set()
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = timewheel_pb2_grpc.TimeWheelServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def add_timer_id(self, timer_id):
        """Регистрирует созданный таймер"""
        self.active_timers.add(timer_id)
        print(f"📝 Зарегистрирован таймер: {timer_id}")
    
    def remove_timer(self, timer_id):
        """Удаляет таймер и убирает из реестра"""
        try:
            request = timewheel_pb2.RemoveTimerRequest(timer_id=timer_id)
            response = self.stub.RemoveTimer(request, metadata=self.metadata)
            
            if response.success:
                self.active_timers.discard(timer_id)
                print(f"✅ Таймер '{timer_id}' удален из системы")
                return True
            else:
                print(f"❌ Ошибка удаления '{timer_id}': {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error для '{timer_id}': {e.code()} - {e.details()}")
            return False
    
    def cleanup_all(self):
        """Удаляет все зарегистрированные таймеры"""
        if not self.active_timers:
            print("🎯 Нет активных таймеров для удаления")
            return
        
        print(f"🧹 Очистка {len(self.active_timers)} активных таймеров...")
        
        timers_to_remove = list(self.active_timers)  # Копия для безопасной итерации
        success_count = 0
        
        for timer_id in timers_to_remove:
            if self.remove_timer(timer_id):
                success_count += 1
        
        print(f"📊 Очистка завершена: {success_count}/{len(timers_to_remove)} удалено")
        
        # Обновляем список активных таймеров
        if success_count == len(timers_to_remove):
            self.active_timers.clear()
            print("✨ Все таймеры успешно удалены")
    
    def remove_by_pattern(self, pattern):
        """Удаляет таймеры по шаблону имени"""
        matching_timers = [t for t in self.active_timers if pattern in t]
        
        if not matching_timers:
            print(f"🔍 Таймеры с шаблоном '{pattern}' не найдены")
            return
        
        print(f"🎯 Найдено {len(matching_timers)} таймеров с шаблоном '{pattern}'")
        
        for timer_id in matching_timers:
            self.remove_timer(timer_id)
    
    def list_active(self):
        """Показывает список активных таймеров"""
        if self.active_timers:
            print(f"📋 Активные таймеры ({len(self.active_timers)}):")
            for timer_id in sorted(self.active_timers):
                print(f"  • {timer_id}")
        else:
            print("📭 Нет активных таймеров")

# Демонстрация использования менеджера
def demonstrate_timer_management():
    print("🎮 Демонстрация управления таймерами\n")
    
    manager = TimerManager()
    
    # Симулируем создание нескольких таймеров
    test_timers = [
        "reminder-meeting-1",
        "reminder-break-2", 
        "bpmn-timeout-task-123",
        "bpmn-boundary-event-456",
        "monitoring-health-check"
    ]
    
    # Регистрируем таймеры как созданные
    for timer_id in test_timers:
        manager.add_timer_id(timer_id)
    
    print()
    manager.list_active()
    
    # Удаляем конкретный таймер
    print(f"\n🎯 Удаление конкретного таймера:")
    manager.remove_timer("reminder-meeting-1")
    
    # Удаляем по шаблону
    print(f"\n🔍 Удаление таймеров с шаблоном 'reminder':")
    manager.remove_by_pattern("reminder")
    
    print()
    manager.list_active()
    
    # Полная очистка
    print(f"\n🧹 Полная очистка:")
    manager.cleanup_all()

# BPMN специфичные операции
def handle_bpmn_timer_cleanup():
    print("🔄 BPMN операции с таймерами\n")
    
    # Отмена граничного таймера при завершении активности
    def cancel_boundary_timer(activity_id):
        timer_id = f"bpmn-boundary-{activity_id}"
        print(f"🎯 Отмена boundary таймера для активности {activity_id}")
        return remove_timer(timer_id)
    
    # Очистка всех таймеров процесса при его отмене
    def cancel_process_timers(process_instance_id):
        print(f"🔄 Отмена всех таймеров для процесса {process_instance_id}")
        
        # В реальности здесь был бы запрос к ListTimers с фильтром
        # но для демо используем предопределенный список
        process_timers = [
            f"bpmn-boundary-{process_instance_id}-task1",
            f"bpmn-boundary-{process_instance_id}-task2", 
            f"bpmn-intermediate-{process_instance_id}-wait1"
        ]
        
        success_count = 0
        for timer_id in process_timers:
            if remove_timer(timer_id):
                success_count += 1
        
        print(f"📊 Отменено {success_count}/{len(process_timers)} таймеров процесса")
        return success_count == len(process_timers)
    
    # Демонстрация
    cancel_boundary_timer("user-task-approval-123")
    print()
    cancel_process_timers("proc-instance-456")

if __name__ == "__main__":
    # Простое удаление
    remove_timer("test-timer-1")
    
    print("\n" + "="*50)
    
    # Демонстрация менеджера
    demonstrate_timer_management()
    
    print("\n" + "="*50)
    
    # BPMN операции
    handle_bpmn_timer_cleanup()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'timewheel.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const timewheelProto = grpc.loadPackageDefinition(packageDefinition).atom.timewheel.v1;

async function removeTimer(timerId) {
    const client = new timewheelProto.TimeWheelService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { timer_id: timerId };
        
        client.removeTimer(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Таймер '${timerId}' удален`);
                console.log(`💬 ${response.message}`);
                resolve(true);
            } else {
                console.log(`❌ Ошибка удаления '${timerId}': ${response.message}`);
                resolve(false);
            }
        });
    });
}

// Класс для управления группами таймеров
class TimerGroup {
    constructor(name) {
        this.name = name;
        this.timers = new Map();
    }
    
    add(timerId, description) {
        this.timers.set(timerId, {
            description,
            createdAt: new Date()
        });
        console.log(`📝 Добавлен в группу '${this.name}': ${timerId}`);
    }
    
    async remove(timerId) {
        if (!this.timers.has(timerId)) {
            console.log(`⚠️ Таймер '${timerId}' не найден в группе '${this.name}'`);
            return false;
        }
        
        try {
            const success = await removeTimer(timerId);
            if (success) {
                this.timers.delete(timerId);
                console.log(`🗑️ Удален из группы '${this.name}': ${timerId}`);
            }
            return success;
        } catch (error) {
            console.log(`❌ Ошибка удаления '${timerId}': ${error.message}`);
            return false;
        }
    }
    
    async removeAll() {
        console.log(`🧹 Очистка группы '${this.name}' (${this.timers.size} таймеров)...`);
        
        const timerIds = Array.from(this.timers.keys());
        const results = await Promise.allSettled(
            timerIds.map(id => this.remove(id))
        );
        
        const successful = results.filter(r => r.status === 'fulfilled' && r.value === true).length;
        const failed = results.length - successful;
        
        console.log(`📊 Группа '${this.name}': ${successful} удалено, ${failed} ошибок`);
        return { successful, failed };
    }
    
    list() {
        console.log(`📋 Группа '${this.name}' (${this.timers.size} таймеров):`);
        
        for (const [timerId, info] of this.timers) {
            const age = Math.floor((Date.now() - info.createdAt.getTime()) / 1000);
            console.log(`  • ${timerId} - ${info.description} (${age}s назад)`);
        }
    }
    
    async removeByPattern(pattern) {
        const matching = Array.from(this.timers.keys()).filter(id => id.includes(pattern));
        
        if (matching.length === 0) {
            console.log(`🔍 В группе '${this.name}' нет таймеров с шаблоном '${pattern}'`);
            return { successful: 0, failed: 0 };
        }
        
        console.log(`🎯 Найдено ${matching.length} таймеров с шаблоном '${pattern}'`);
        
        const results = await Promise.allSettled(
            matching.map(id => this.remove(id))
        );
        
        const successful = results.filter(r => r.status === 'fulfilled' && r.value === true).length;
        const failed = results.length - successful;
        
        return { successful, failed };
    }
}

// Демонстрация групповых операций
async function demonstrateGroupOperations() {
    console.log('👥 Демонстрация групповых операций с таймерами\n');
    
    // Создаем группы таймеров
    const reminderGroup = new TimerGroup('Напоминания');
    const bpmnGroup = new TimerGroup('BPMN Процессы');
    const monitoringGroup = new TimerGroup('Мониторинг');
    
    // Симулируем добавление таймеров
    reminderGroup.add('reminder-meeting-1', 'Напоминание о встрече');
    reminderGroup.add('reminder-break-1', 'Напоминание о перерыве');
    reminderGroup.add('reminder-call-client', 'Звонок клиенту');
    
    bpmnGroup.add('bpmn-boundary-task-123', 'Boundary таймер для задачи 123');
    bpmnGroup.add('bpmn-boundary-task-456', 'Boundary таймер для задачи 456');
    bpmnGroup.add('bpmn-intermediate-wait', 'Промежуточное событие ожидания');
    
    monitoringGroup.add('monitoring-health', 'Проверка здоровья системы');
    monitoringGroup.add('monitoring-metrics', 'Сбор метрик');
    
    // Показываем все группы
    console.log('📊 Текущее состояние групп:');
    console.log('-'.repeat(40));
    reminderGroup.list();
    console.log();
    bpmnGroup.list();
    console.log();
    monitoringGroup.list();
    
    console.log('\n🎯 Удаление по шаблону "reminder" из группы напоминаний:');
    const reminderResult = await reminderGroup.removeByPattern('reminder');
    console.log(`Результат: ${reminderResult.successful} удалено, ${reminderResult.failed} ошибок`);
    
    console.log('\n🗑️ Полная очистка группы BPMN:');
    await bpmnGroup.removeAll();
    
    console.log('\n📋 Финальное состояние групп:');
    console.log('-'.repeat(40));
    reminderGroup.list();
    bpmnGroup.list();
    monitoringGroup.list();
}

// Специализированные функции для BPMN
const BPMNTimerManager = {
    // Отмена boundary таймера при завершении активности
    async cancelBoundaryTimer(processInstanceId, activityId) {
        const timerId = `bpmn-boundary-${processInstanceId}-${activityId}`;
        console.log(`🎯 Отмена boundary таймера: ${activityId}`);
        
        try {
            return await removeTimer(timerId);
        } catch (error) {
            console.log(`⚠️ Boundary таймер ${activityId} уже мог быть удален: ${error.message}`);
            return false;
        }
    },
    
    // Отмена всех таймеров процесса
    async cancelProcessTimers(processInstanceId) {
        console.log(`🔄 Отмена всех таймеров для процесса: ${processInstanceId}`);
        
        // В реальном приложении здесь был бы запрос к ListTimers с фильтром
        // Для демонстрации используем предопределенные ID
        const processTimers = [
            `bpmn-boundary-${processInstanceId}-task1`,
            `bpmn-boundary-${processInstanceId}-task2`,
            `bpmn-intermediate-${processInstanceId}-wait1`,
            `bpmn-start-${processInstanceId}`
        ];
        
        console.log(`📋 Найдено ${processTimers.length} таймеров для удаления`);
        
        const results = await Promise.allSettled(
            processTimers.map(timerId => removeTimer(timerId))
        );
        
        const successful = results.filter(r => r.status === 'fulfilled' && r.value === true).length;
        const failed = results.length - successful;
        
        console.log(`📊 Процесс ${processInstanceId}: ${successful} удалено, ${failed} не найдено`);
        
        return { successful, failed, total: processTimers.length };
    },
    
    // Очистка просроченных таймеров
    async cleanupExpiredTimers(olderThanHours = 24) {
        console.log(`🧹 Очистка таймеров старше ${olderThanHours} часов`);
        
        // В реальности здесь был бы запрос к ListTimers с временным фильтром
        const expiredTimers = [
            'expired-timer-1',
            'expired-timer-2',
            'old-boundary-timer'
        ];
        
        if (expiredTimers.length === 0) {
            console.log('✨ Просроченные таймеры не найдены');
            return { successful: 0, failed: 0 };
        }
        
        console.log(`🗑️ Найдено ${expiredTimers.length} просроченных таймеров`);
        
        const results = await Promise.allSettled(
            expiredTimers.map(timerId => removeTimer(timerId))
        );
        
        const successful = results.filter(r => r.status === 'fulfilled' && r.value === true).length;
        const failed = results.length - successful;
        
        return { successful, failed };
    }
};

// Демонстрация BPMN операций
async function demonstrateBPMNOperations() {
    console.log('🔄 BPMN операции с таймерами\n');
    
    // Отмена boundary таймера
    await BPMNTimerManager.cancelBoundaryTimer('proc-123', 'user-task-approval');
    
    console.log();
    
    // Отмена всех таймеров процесса
    await BPMNTimerManager.cancelProcessTimers('proc-456');
    
    console.log();
    
    // Очистка просроченных
    const cleanupResult = await BPMNTimerManager.cleanupExpiredTimers(12);
    console.log(`🧹 Очистка завершена: ${cleanupResult.successful} удалено`);
}

// Основная демонстрация
async function main() {
    try {
        // Простое удаление
        console.log('🎯 Простое удаление таймера:\n');
        await removeTimer('test-timer-example');
        
        console.log('\n' + '='.repeat(60));
        
        // Групповые операции
        await demonstrateGroupOperations();
        
        console.log('\n' + '='.repeat(60));
        
        // BPMN операции
        await demonstrateBPMNOperations();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Применение

### BPMN Process Cancellation
```javascript
// Отмена всех таймеров при отмене процесса
const processTimers = await listTimersByProcessId(processInstanceId);
await Promise.all(processTimers.map(t => removeTimer(t.timer_id)));
```

### Activity Completion
```javascript
// Отмена boundary таймера при завершении активности
await removeTimer(`boundary-${activityId}`);
```

### System Cleanup
```python
# Очистка просроченных таймеров
expired_timers = await listExpiredTimers(hours=24)
for timer in expired_timers:
    await removeTimer(timer.timer_id)
```

### Resource Management
```go
// Cleanup при завершении работы приложения
defer func() {
    for _, timerId := range activeTimers {
        removeTimer(timerId)
    }
}()
```

## Частые сценарии

### Отмена по завершении
- **BPMN Activity** завершилась → отмена boundary таймеров
- **Process** отменен → отмена всех таймеров процесса
- **User logout** → отмена персональных напоминаний

### Очистка ресурсов
- **Expired timers** → автоматическая очистка
- **System shutdown** → отмена всех активных таймеров
- **Memory pressure** → приоритетная очистка

### Обработка ошибок
- **Timer not found** → безопасное игнорирование
- **Already fired** → логирование для аудита
- **Network error** → повторная попытка

## Связанные методы
- [AddTimer](add-timer.md) - Создание таймеров
- [GetTimerStatus](get-timer-status.md) - Проверка перед удалением
- [ListTimers](list-timers.md) - Массовые операции
