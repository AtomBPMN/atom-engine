# GetTimerStatus

## Описание
Получает подробную информацию о статусе таймера, включая время до срабатывания, расписание и информацию о повторах.

## Синтаксис
```protobuf
rpc GetTimerStatus(GetTimerStatusRequest) returns (GetTimerStatusResponse);
```

## Package
```protobuf
package atom.timewheel.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer` или `*`

## Параметры запроса

### GetTimerStatusRequest
```protobuf
message GetTimerStatusRequest {
  string timer_id = 1;      // ID таймера для проверки
}
```

## Параметры ответа

### GetTimerStatusResponse
```protobuf
message GetTimerStatusResponse {
  string timer_id = 1;        // ID таймера
  string status = 2;          // Статус: "pending", "fired", "cancelled"
  int64 scheduled_at = 3;     // Unix timestamp планового срабатывания
  int64 remaining_ms = 4;     // Миллисекунды до срабатывания
  bool is_repeating = 5;      // Является ли повторяющимся
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
    
    // Проверяем статус таймера
    response, err := client.GetTimerStatus(ctx, &pb.GetTimerStatusRequest{
        TimerId: "timer-simple-30s",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("🔍 Статус таймера: %s\n", response.TimerId)
    fmt.Printf("📊 Статус: %s\n", response.Status)
    
    if response.Status == "pending" {
        scheduledTime := time.Unix(response.ScheduledAt, 0)
        fmt.Printf("📅 Запланирован на: %s\n", scheduledTime.Format("15:04:05"))
        
        remainingDuration := time.Duration(response.RemainingMs) * time.Millisecond
        fmt.Printf("⏱️ Осталось: %s\n", remainingDuration.String())
        
        if response.IsRepeating {
            fmt.Printf("🔄 Повторяющийся таймер\n")
        } else {
            fmt.Printf("1️⃣ Одноразовый таймер\n")
        }
    } else if response.Status == "fired" {
        fmt.Printf("✅ Таймер уже сработал\n")
    } else if response.Status == "cancelled" {
        fmt.Printf("❌ Таймер отменен\n")
    }
}

// Мониторинг группы таймеров
func monitorTimers(client pb.TimeWheelServiceClient, ctx context.Context, timerIds []string) {
    fmt.Printf("📊 Мониторинг %d таймеров...\n\n", len(timerIds))
    
    for _, timerId := range timerIds {
        response, err := client.GetTimerStatus(ctx, &pb.GetTimerStatusRequest{
            TimerId: timerId,
        })
        
        if err != nil {
            fmt.Printf("❌ Ошибка для %s: %v\n", timerId, err)
            continue
        }
        
        status := "❓"
        switch response.Status {
        case "pending":
            status = "⏳"
        case "fired":
            status = "✅"
        case "cancelled":
            status = "❌"
        }
        
        fmt.Printf("%s %s - %s", status, timerId, response.Status)
        
        if response.Status == "pending" {
            remainingDuration := time.Duration(response.RemainingMs) * time.Millisecond
            fmt.Printf(" (осталось: %s)", remainingDuration.String())
            
            if response.IsRepeating {
                fmt.Printf(" [повторяющийся]")
            }
        }
        
        fmt.Println()
    }
}

// Ожидание срабатывания таймера
func waitForTimer(client pb.TimeWheelServiceClient, ctx context.Context, timerId string, pollInterval time.Duration) {
    fmt.Printf("⏱️ Ожидание срабатывания таймера: %s\n", timerId)
    
    ticker := time.NewTicker(pollInterval)
    defer ticker.Stop()
    
    for {
        response, err := client.GetTimerStatus(ctx, &pb.GetTimerStatusRequest{
            TimerId: timerId,
        })
        
        if err != nil {
            fmt.Printf("❌ Ошибка проверки статуса: %v\n", err)
            break
        }
        
        if response.Status == "fired" {
            fmt.Printf("🎯 Таймер %s сработал!\n", timerId)
            break
        } else if response.Status == "cancelled" {
            fmt.Printf("❌ Таймер %s был отменен\n", timerId)
            break
        } else if response.Status == "pending" {
            remainingDuration := time.Duration(response.RemainingMs) * time.Millisecond
            fmt.Printf("⏳ Осталось: %s\n", remainingDuration.String())
        }
        
        <-ticker.C
    }
}
```

### Python
```python
import grpc
import time
from datetime import datetime, timedelta
import threading

import timewheel_pb2
import timewheel_pb2_grpc

def get_timer_status(timer_id):
    channel = grpc.insecure_channel('localhost:27500')
    stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = timewheel_pb2.GetTimerStatusRequest(timer_id=timer_id)
    
    try:
        response = stub.GetTimerStatus(request, metadata=metadata)
        
        print(f"🔍 Таймер: {response.timer_id}")
        
        status_icons = {
            'pending': '⏳',
            'fired': '✅', 
            'cancelled': '❌'
        }
        
        icon = status_icons.get(response.status, '❓')
        print(f"{icon} Статус: {response.status}")
        
        if response.status == 'pending':
            scheduled_time = datetime.fromtimestamp(response.scheduled_at)
            print(f"📅 Запланирован: {scheduled_time.strftime('%H:%M:%S')}")
            
            remaining_seconds = response.remaining_ms / 1000
            remaining_td = timedelta(seconds=remaining_seconds)
            print(f"⏱️ Осталось: {remaining_td}")
            
            if response.is_repeating:
                print("🔄 Повторяющийся таймер")
            else:
                print("1️⃣ Одноразовый таймер")
        
        return {
            'timer_id': response.timer_id,
            'status': response.status,
            'scheduled_at': response.scheduled_at,
            'remaining_ms': response.remaining_ms,
            'is_repeating': response.is_repeating
        }
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Система мониторинга таймеров
class TimerMonitor:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = timewheel_pb2_grpc.TimeWheelServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
        self.monitoring = False
        self.monitored_timers = {}
    
    def add_timer_to_monitor(self, timer_id, callback=None):
        """Добавляет таймер для мониторинга"""
        self.monitored_timers[timer_id] = {
            'callback': callback,
            'last_status': None,
            'first_check': True
        }
        print(f"👀 Добавлен в мониторинг: {timer_id}")
    
    def check_timer_status(self, timer_id):
        """Проверяет статус конкретного таймера"""
        try:
            request = timewheel_pb2.GetTimerStatusRequest(timer_id=timer_id)
            response = self.stub.GetTimerStatus(request, metadata=self.metadata)
            
            return {
                'timer_id': response.timer_id,
                'status': response.status,
                'scheduled_at': response.scheduled_at,
                'remaining_ms': response.remaining_ms,
                'is_repeating': response.is_repeating
            }
        except grpc.RpcError as e:
            print(f"❌ Ошибка проверки {timer_id}: {e.details()}")
            return None
    
    def start_monitoring(self, interval=5):
        """Запускает мониторинг с заданным интервалом (секунды)"""
        if self.monitoring:
            print("⚠️ Мониторинг уже запущен")
            return
        
        self.monitoring = True
        print(f"🚀 Запуск мониторинга каждые {interval} секунд")
        
        def monitor_loop():
            while self.monitoring:
                self._check_all_timers()
                time.sleep(interval)
        
        monitor_thread = threading.Thread(target=monitor_loop, daemon=True)
        monitor_thread.start()
    
    def _check_all_timers(self):
        """Проверяет все таймеры под мониторингом"""
        for timer_id, timer_info in list(self.monitored_timers.items()):
            status_data = self.check_timer_status(timer_id)
            
            if status_data is None:
                continue
            
            current_status = status_data['status']
            previous_status = timer_info['last_status']
            
            # Показываем изменения статуса или первую проверку
            if timer_info['first_check'] or current_status != previous_status:
                self._report_status_change(timer_id, current_status, status_data)
                
                # Вызываем callback если статус изменился
                if timer_info['callback'] and current_status != previous_status:
                    try:
                        timer_info['callback'](timer_id, current_status, status_data)
                    except Exception as e:
                        print(f"❌ Ошибка в callback для {timer_id}: {e}")
                
                timer_info['last_status'] = current_status
                timer_info['first_check'] = False
            
            # Убираем из мониторинга завершенные таймеры
            if current_status in ['fired', 'cancelled']:
                print(f"🏁 Убираем из мониторинга: {timer_id} (статус: {current_status})")
                del self.monitored_timers[timer_id]
    
    def _report_status_change(self, timer_id, status, status_data):
        """Выводит информацию о изменении статуса"""
        icons = {'pending': '⏳', 'fired': '🎯', 'cancelled': '❌'}
        icon = icons.get(status, '❓')
        
        timestamp = datetime.now().strftime('%H:%M:%S')
        print(f"[{timestamp}] {icon} {timer_id}: {status}")
        
        if status == 'pending' and status_data['remaining_ms'] > 0:
            remaining_sec = status_data['remaining_ms'] / 1000
            print(f"           ⏱️ Осталось: {remaining_sec:.1f}s")
    
    def stop_monitoring(self):
        """Останавливает мониторинг"""
        self.monitoring = False
        print("🛑 Мониторинг остановлен")
    
    def list_monitored(self):
        """Показывает список отслеживаемых таймеров"""
        if not self.monitored_timers:
            print("📭 Нет таймеров под мониторингом")
            return
        
        print(f"📋 Таймеры под мониторингом ({len(self.monitored_timers)}):")
        for timer_id, info in self.monitored_timers.items():
            status = info['last_status'] or 'не проверялся'
            print(f"  • {timer_id} - последний статус: {status}")

# Демонстрация системы мониторинга
def demonstrate_monitoring():
    print("👁️ Демонстрация мониторинга таймеров\n")
    
    monitor = TimerMonitor()
    
    # Callback функция для уведомлений
    def timer_callback(timer_id, status, data):
        if status == 'fired':
            print(f"🔔 УВЕДОМЛЕНИЕ: Таймер {timer_id} сработал!")
        elif status == 'cancelled':
            print(f"🚫 УВЕДОМЛЕНИЕ: Таймер {timer_id} отменен!")
    
    # Добавляем таймеры для мониторинга (предполагаем что они созданы)
    test_timers = [
        'reminder-meeting-1',
        'bpmn-boundary-task-123', 
        'monitoring-health-check'
    ]
    
    for timer_id in test_timers:
        monitor.add_timer_to_monitor(timer_id, timer_callback)
    
    print()
    monitor.list_monitored()
    
    # Запускаем мониторинг на 30 секунд
    print(f"\n🚀 Запускаем мониторинг на 30 секунд...")
    monitor.start_monitoring(interval=3)
    
    # Симулируем работу
    time.sleep(30)
    
    monitor.stop_monitoring()
    print("✅ Демонстрация мониторинга завершена")

# Утилиты для анализа таймеров
def analyze_timer_performance(timer_ids):
    """Анализирует производительность группы таймеров"""
    print(f"📈 Анализ производительности {len(timer_ids)} таймеров\n")
    
    statuses = {'pending': 0, 'fired': 0, 'cancelled': 0, 'error': 0}
    remaining_times = []
    
    for timer_id in timer_ids:
        status_data = get_timer_status(timer_id)
        
        if status_data:
            statuses[status_data['status']] += 1
            
            if status_data['status'] == 'pending':
                remaining_times.append(status_data['remaining_ms'] / 1000)
        else:
            statuses['error'] += 1
        
        print("-" * 30)
    
    # Сводная статистика
    print(f"\n📊 СВОДКА АНАЛИЗА:")
    print(f"   ⏳ Активных: {statuses['pending']}")
    print(f"   ✅ Сработавших: {statuses['fired']}")  
    print(f"   ❌ Отмененных: {statuses['cancelled']}")
    print(f"   💥 Ошибок: {statuses['error']}")
    
    if remaining_times:
        avg_remaining = sum(remaining_times) / len(remaining_times)
        min_remaining = min(remaining_times)
        max_remaining = max(remaining_times)
        
        print(f"\n⏱️ ВРЕМЕНА ОЖИДАНИЯ (секунды):")
        print(f"   Среднее: {avg_remaining:.1f}")
        print(f"   Минимум: {min_remaining:.1f}")
        print(f"   Максимум: {max_remaining:.1f}")

if __name__ == "__main__":
    # Простая проверка
    get_timer_status("test-timer-1")
    
    print("\n" + "="*50)
    
    # Демонстрация мониторинга
    demonstrate_monitoring()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'timewheel.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const timewheelProto = grpc.loadPackageDefinition(packageDefinition).atom.timewheel.v1;

async function getTimerStatus(timerId) {
    const client = new timewheelProto.TimeWheelService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { timer_id: timerId };
        
        client.getTimerStatus(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            const statusIcons = {
                'pending': '⏳',
                'fired': '✅',
                'cancelled': '❌'
            };
            
            const icon = statusIcons[response.status] || '❓';
            console.log(`🔍 Таймер: ${response.timer_id}`);
            console.log(`${icon} Статус: ${response.status}`);
            
            if (response.status === 'pending') {
                const scheduledTime = new Date(response.scheduled_at * 1000);
                console.log(`📅 Запланирован: ${scheduledTime.toLocaleTimeString()}`);
                
                const remainingMs = response.remaining_ms;
                const remainingSeconds = Math.floor(remainingMs / 1000);
                const remainingMinutes = Math.floor(remainingSeconds / 60);
                const remainingSecs = remainingSeconds % 60;
                
                if (remainingMinutes > 0) {
                    console.log(`⏱️ Осталось: ${remainingMinutes}м ${remainingSecs}с`);
                } else {
                    console.log(`⏱️ Осталось: ${remainingSecs}с`);
                }
                
                console.log(`🔄 ${response.is_repeating ? 'Повторяющийся' : 'Одноразовый'}`);
            }
            
            resolve({
                timerId: response.timer_id,
                status: response.status,
                scheduledAt: response.scheduled_at,
                remainingMs: response.remaining_ms,
                isRepeating: response.is_repeating
            });
        });
    });
}

// Класс для реактивного мониторинга таймеров
class ReactiveTimerMonitor {
    constructor() {
        this.timers = new Map();
        this.eventHandlers = new Map();
        this.intervalId = null;
        this.isMonitoring = false;
    }
    
    // Добавляет таймер для мониторинга
    watch(timerId, options = {}) {
        const config = {
            onStatusChange: options.onStatusChange || null,
            onFire: options.onFire || null,
            onCancel: options.onCancel || null,
            checkInterval: options.checkInterval || 5000,
            autoRemove: options.autoRemove !== false // По умолчанию true
        };
        
        this.timers.set(timerId, {
            ...config,
            lastStatus: null,
            firstCheck: true
        });
        
        console.log(`👀 Добавлен в мониторинг: ${timerId}`);
        
        if (!this.isMonitoring) {
            this.startMonitoring();
        }
    }
    
    // Убирает таймер из мониторинга
    unwatch(timerId) {
        if (this.timers.delete(timerId)) {
            console.log(`👁️‍🗨️ Убран из мониторинга: ${timerId}`);
            
            if (this.timers.size === 0) {
                this.stopMonitoring();
            }
        }
    }
    
    // Запускает мониторинг
    startMonitoring() {
        if (this.isMonitoring) return;
        
        this.isMonitoring = true;
        console.log('🚀 Мониторинг запущен');
        
        const checkAll = async () => {
            if (!this.isMonitoring) return;
            
            for (const [timerId, config] of this.timers.entries()) {
                await this.checkTimer(timerId, config);
            }
            
            // Планируем следующую проверку
            if (this.isMonitoring && this.timers.size > 0) {
                this.intervalId = setTimeout(checkAll, 5000); // 5 секунд
            }
        };
        
        checkAll();
    }
    
    // Останавливает мониторинг
    stopMonitoring() {
        if (!this.isMonitoring) return;
        
        this.isMonitoring = false;
        
        if (this.intervalId) {
            clearTimeout(this.intervalId);
            this.intervalId = null;
        }
        
        console.log('🛑 Мониторинг остановлен');
    }
    
    // Проверяет конкретный таймер
    async checkTimer(timerId, config) {
        try {
            const status = await getTimerStatus(timerId);
            
            const currentStatus = status.status;
            const previousStatus = config.lastStatus;
            
            // Обрабатываем изменения статуса
            if (config.firstCheck || currentStatus !== previousStatus) {
                this.handleStatusChange(timerId, currentStatus, previousStatus, status);
                
                // Вызываем пользовательские обработчики
                if (config.onStatusChange) {
                    config.onStatusChange(timerId, currentStatus, previousStatus, status);
                }
                
                if (currentStatus === 'fired' && config.onFire) {
                    config.onFire(timerId, status);
                }
                
                if (currentStatus === 'cancelled' && config.onCancel) {
                    config.onCancel(timerId, status);
                }
                
                config.lastStatus = currentStatus;
                config.firstCheck = false;
            }
            
            // Автоматически убираем завершенные таймеры
            if (config.autoRemove && ['fired', 'cancelled'].includes(currentStatus)) {
                this.unwatch(timerId);
            }
            
        } catch (error) {
            console.log(`❌ Ошибка проверки ${timerId}: ${error.message}`);
        }
    }
    
    // Обрабатывает изменения статуса
    handleStatusChange(timerId, currentStatus, previousStatus, statusData) {
        const timestamp = new Date().toLocaleTimeString();
        const icons = { 'pending': '⏳', 'fired': '🎯', 'cancelled': '❌' };
        const icon = icons[currentStatus] || '❓';
        
        console.log(`[${timestamp}] ${icon} ${timerId}: ${previousStatus || 'new'} → ${currentStatus}`);
        
        if (currentStatus === 'pending' && statusData.remainingMs > 0) {
            const remainingSeconds = Math.floor(statusData.remainingMs / 1000);
            console.log(`           ⏱️ Осталось: ${remainingSeconds}с`);
        }
    }
    
    // Показывает список отслеживаемых таймеров
    listWatched() {
        if (this.timers.size === 0) {
            console.log('📭 Нет таймеров под мониторингом');
            return;
        }
        
        console.log(`📋 Отслеживаемые таймеры (${this.timers.size}):`);
        for (const [timerId, config] of this.timers.entries()) {
            const status = config.lastStatus || 'не проверялся';
            console.log(`  • ${timerId} - ${status}`);
        }
    }
    
    // Очищает все таймеры
    clear() {
        this.timers.clear();
        this.stopMonitoring();
        console.log('🧹 Мониторинг очищен');
    }
}

// Демонстрация реактивного мониторинга
async function demonstrateReactiveMonitoring() {
    console.log('⚡ Демонстрация реактивного мониторинга\n');
    
    const monitor = new ReactiveTimerMonitor();
    
    // Настраиваем обработчики событий
    const handlers = {
        onStatusChange: (timerId, current, previous) => {
            if (previous) {
                console.log(`🔄 ${timerId}: статус изменился с ${previous} на ${current}`);
            }
        },
        
        onFire: (timerId, statusData) => {
            console.log(`🔔 СРАБАТЫВАНИЕ: ${timerId} сработал в ${new Date().toLocaleTimeString()}`);
            
            // Здесь можно добавить логику обработки срабатывания
            if (statusData.isRepeating) {
                console.log(`   🔄 Это повторяющийся таймер, ждем следующего срабатывания`);
            }
        },
        
        onCancel: (timerId, statusData) => {
            console.log(`🚫 ОТМЕНА: ${timerId} был отменен`);
        }
    };
    
    // Добавляем таймеры для мониторинга
    const testTimers = [
        'demo-timer-short',   // Короткий таймер для быстрой демонстрации
        'demo-timer-medium',  // Средний таймер
        'demo-timer-repeat'   // Повторяющийся таймер
    ];
    
    testTimers.forEach(timerId => {
        monitor.watch(timerId, handlers);
    });
    
    console.log('\n📋 Состояние мониторинга:');
    monitor.listWatched();
    
    console.log('\n⏱️ Мониторинг будет работать 60 секунд...');
    console.log('(В реальности таймеры должны быть созданы заранее)');
    
    // Симулируем работу мониторинга
    setTimeout(() => {
        console.log('\n🏁 Завершаем демонстрацию мониторинга');
        monitor.clear();
    }, 60000);
}

// Утилиты для анализа состояния таймеров
class TimerAnalytics {
    static async analyzeTimerGroup(timerIds) {
        console.log(`📊 Анализ группы из ${timerIds.length} таймеров\n`);
        
        const results = {
            pending: [],
            fired: [],
            cancelled: [],
            errors: []
        };
        
        const remainingTimes = [];
        
        for (const timerId of timerIds) {
            try {
                const status = await getTimerStatus(timerId);
                results[status.status].push(timerId);
                
                if (status.status === 'pending') {
                    remainingTimes.push(status.remainingMs / 1000);
                }
                
                console.log(`✓ ${timerId}: ${status.status}`);
                
            } catch (error) {
                results.errors.push({ timerId, error: error.message });
                console.log(`✗ ${timerId}: ошибка - ${error.message}`);
            }
        }
        
        // Сводная статистика
        console.log('\n📈 СВОДКА АНАЛИЗА:');
        console.log(`   ⏳ Активных: ${results.pending.length}`);
        console.log(`   ✅ Сработавших: ${results.fired.length}`);
        console.log(`   ❌ Отмененных: ${results.cancelled.length}`);
        console.log(`   💥 Ошибок: ${results.errors.length}`);
        
        if (remainingTimes.length > 0) {
            const avgRemaining = remainingTimes.reduce((a, b) => a + b, 0) / remainingTimes.length;
            const minRemaining = Math.min(...remainingTimes);
            const maxRemaining = Math.max(...remainingTimes);
            
            console.log('\n⏱️ ВРЕМЕНА ОЖИДАНИЯ:');
            console.log(`   Среднее: ${avgRemaining.toFixed(1)}с`);
            console.log(`   Минимум: ${minRemaining.toFixed(1)}с`);
            console.log(`   Максимум: ${maxRemaining.toFixed(1)}с`);
        }
        
        return results;
    }
    
    static async monitorUntilCompletion(timerIds, options = {}) {
        const pollInterval = options.pollInterval || 2000;
        const timeout = options.timeout || 300000; // 5 минут
        
        console.log(`🎯 Мониторинг до завершения (${timerIds.length} таймеров)`);
        console.log(`   Интервал проверки: ${pollInterval}мс`);
        console.log(`   Таймаут: ${timeout}мс\n`);
        
        const startTime = Date.now();
        const completed = new Set();
        
        const checkInterval = setInterval(async () => {
            const elapsed = Date.now() - startTime;
            
            if (elapsed >= timeout) {
                console.log('⏰ Таймаут мониторинга достигнут');
                clearInterval(checkInterval);
                return;
            }
            
            for (const timerId of timerIds) {
                if (completed.has(timerId)) continue;
                
                try {
                    const status = await getTimerStatus(timerId);
                    
                    if (['fired', 'cancelled'].includes(status.status)) {
                        completed.add(timerId);
                        const icon = status.status === 'fired' ? '🎯' : '❌';
                        console.log(`${icon} ${timerId} завершен (${status.status})`);
                    }
                } catch (error) {
                    completed.add(timerId);
                    console.log(`💥 ${timerId} ошибка: ${error.message}`);
                }
            }
            
            if (completed.size === timerIds.length) {
                console.log(`\n🏆 Все таймеры завершены за ${elapsed}мс`);
                clearInterval(checkInterval);
            }
        }, pollInterval);
    }
}

// Основная демонстрация
async function main() {
    try {
        // Простая проверка статуса
        console.log('🔍 Простая проверка статуса:\n');
        await getTimerStatus('test-timer-example');
        
        console.log('\n' + '='.repeat(60));
        
        // Реактивный мониторинг
        await demonstrateReactiveMonitoring();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Статусы таймера

### Возможные статусы
- **`pending`** - Ожидает срабатывания
- **`fired`** - Сработал (выполнен)
- **`cancelled`** - Отменен

### Информация для pending таймеров
- **scheduled_at** - Unix timestamp планового срабатывания
- **remaining_ms** - Миллисекунды до срабатывания
- **is_repeating** - Повторяющийся ли таймер

## Применение

### BPMN Мониторинг
```javascript
// Мониторинг boundary таймера
const status = await getTimerStatus(`boundary-${activityId}`);
if (status.status === 'pending' && status.remainingMs < 60000) {
    console.log('Boundary таймер скоро сработает!');
}
```

### Система напоминаний
```python
# Проверка до отправки напоминания
status = get_timer_status('reminder-meeting')
if status and status['status'] == 'pending':
    remaining_min = status['remaining_ms'] / 60000
    print(f"До напоминания осталось {remaining_min:.1f} минут")
```

### Отладка и мониторинг
```go
// Диагностика проблемных таймеров
for _, timerId := range problematicTimers {
    status := getTimerStatus(timerId)
    if status.Status == "cancelled" {
        log.Printf("Таймер %s был отменен неожиданно", timerId)
    }
}
```

### Производительность
- **O(1)** поиск в timewheel структуре
- **Минимальная нагрузка** на систему
- **Точность** до миллисекунды

## Связанные методы
- [AddTimer](add-timer.md) - Создание таймеров для мониторинга
- [RemoveTimer](remove-timer.md) - Отмена при необходимости
- [ListTimers](list-timers.md) - Массовый мониторинг
