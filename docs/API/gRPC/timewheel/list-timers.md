# ListTimers

## Описание
Получает пагинированный список таймеров с фильтрацией по статусу и сортировкой. Включает подробную информацию о каждом таймере включая привязку к BPMN процессам.

## Синтаксис
```protobuf
rpc ListTimers(ListTimersRequest) returns (ListTimersResponse);
```

## Package
```protobuf
package atom.timewheel.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer` или `*`

## Параметры запроса

### ListTimersRequest
```protobuf
message ListTimersRequest {
  string status_filter = 1;    // Фильтр: "SCHEDULED", "FIRED", "CANCELLED"
  int32 limit = 2;             // ⚠️ Устарело: используйте page_size
  int32 page_size = 3;         // Размер страницы (по умолчанию: 20)
  int32 page = 4;              // Номер страницы (с 1, по умолчанию: 1)
  string sort_by = 5;          // Поле сортировки (по умолчанию: "created_at")
  string sort_order = 6;       // Порядок: "ASC" или "DESC" (по умолчанию: "DESC")
}
```

## Параметры ответа

### ListTimersResponse
```protobuf
message ListTimersResponse {
  repeated TimerInfo timers = 1;  // Список таймеров
  int32 total_count = 2;         // Общее количество
  int32 page = 3;                // Текущая страница  
  int32 page_size = 4;           // Размер страницы
  int32 total_pages = 5;         // Всего страниц
}

message TimerInfo {
  string timer_id = 1;              // ID таймера
  string element_id = 2;            // ID BPMN элемента
  string process_instance_id = 3;   // ID экземпляра процесса
  string timer_type = 4;            // Тип таймера
  string status = 5;                // Статус таймера
  int64 scheduled_at = 6;           // Unix timestamp срабатывания
  int64 created_at = 7;             // Unix timestamp создания
  string time_duration = 8;         // ISO 8601 длительность
  string time_cycle = 9;            // ISO 8601 цикл повтора
  int64 remaining_seconds = 10;     // Секунды до срабатывания
  int32 wheel_level = 11;           // Уровень timewheel (0-4)
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
    
    // Получаем список активных таймеров
    response, err := client.ListTimers(ctx, &pb.ListTimersRequest{
        StatusFilter: "SCHEDULED",
        PageSize:     10,
        Page:         1,
        SortBy:       "remaining_seconds",
        SortOrder:    "ASC", // Сначала те, что скоро сработают
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("📋 Список активных таймеров (страница %d из %d)\n", 
        response.Page, response.TotalPages)
    fmt.Printf("📊 Найдено: %d таймеров\n\n", response.TotalCount)
    
    if len(response.Timers) == 0 {
        fmt.Println("📭 Нет активных таймеров")
        return
    }
    
    // Выводим информацию о каждом таймере
    for i, timer := range response.Timers {
        fmt.Printf("🔸 %d. %s\n", i+1, timer.TimerId)
        fmt.Printf("   📌 Тип: %s\n", timer.TimerType)
        fmt.Printf("   📊 Статус: %s\n", timer.Status)
        
        if timer.ProcessInstanceId != "" {
            fmt.Printf("   🔄 Процесс: %s\n", timer.ProcessInstanceId)
        }
        
        if timer.ElementId != "" {
            fmt.Printf("   🎯 Элемент: %s\n", timer.ElementId)
        }
        
        // Времена
        createdTime := time.Unix(timer.CreatedAt, 0)
        scheduledTime := time.Unix(timer.ScheduledAt, 0)
        
        fmt.Printf("   📅 Создан: %s\n", createdTime.Format("15:04:05"))
        fmt.Printf("   ⏰ Сработает: %s\n", scheduledTime.Format("15:04:05"))
        
        if timer.RemainingSeconds > 0 {
            remaining := time.Duration(timer.RemainingSeconds) * time.Second
            fmt.Printf("   ⏱️ Осталось: %s\n", remaining.String())
        }
        
        // ISO 8601 информация
        if timer.TimeDuration != "" {
            fmt.Printf("   📏 Длительность: %s\n", timer.TimeDuration)
        }
        
        if timer.TimeCycle != "" {
            fmt.Printf("   🔄 Цикл: %s\n", timer.TimeCycle)
        }
        
        // Техническая информация
        fmt.Printf("   🏗️ Уровень wheel: %d\n", timer.WheelLevel)
        
        fmt.Println()
    }
}

// Поиск таймеров по процессу
func findTimersByProcess(client pb.TimeWheelServiceClient, ctx context.Context, processInstanceId string) {
    fmt.Printf("🔍 Поиск таймеров для процесса: %s\n", processInstanceId)
    
    page := 1
    foundTimers := []*pb.TimerInfo{}
    
    for {
        response, err := client.ListTimers(ctx, &pb.ListTimersRequest{
            PageSize:  100, // Большой размер страницы для эффективного поиска
            Page:      int32(page),
            SortBy:    "created_at",
            SortOrder: "DESC",
        })
        
        if err != nil {
            fmt.Printf("❌ Ошибка: %v\n", err)
            return
        }
        
        // Фильтруем таймеры по процессу
        for _, timer := range response.Timers {
            if timer.ProcessInstanceId == processInstanceId {
                foundTimers = append(foundTimers, timer)
            }
        }
        
        // Проверяем есть ли еще страницы
        if page >= int(response.TotalPages) {
            break
        }
        page++
    }
    
    fmt.Printf("📋 Найдено %d таймеров для процесса %s:\n", len(foundTimers), processInstanceId)
    
    for _, timer := range foundTimers {
        fmt.Printf("  • %s (%s) - %s\n", timer.TimerId, timer.TimerType, timer.Status)
    }
}

// Мониторинг ближайших таймеров
func monitorUpcomingTimers(client pb.TimeWheelServiceClient, ctx context.Context, withinMinutes int) {
    fmt.Printf("⏰ Мониторинг таймеров на ближайшие %d минут\n", withinMinutes)
    
    response, err := client.ListTimers(ctx, &pb.ListTimersRequest{
        StatusFilter: "SCHEDULED",
        PageSize:     50,
        Page:         1,
        SortBy:       "remaining_seconds",
        SortOrder:    "ASC",
    })
    
    if err != nil {
        fmt.Printf("❌ Ошибка: %v\n", err)
        return
    }
    
    upcomingTimers := []*pb.TimerInfo{}
    thresholdSeconds := int64(withinMinutes * 60)
    
    for _, timer := range response.Timers {
        if timer.RemainingSeconds <= thresholdSeconds && timer.RemainingSeconds > 0 {
            upcomingTimers = append(upcomingTimers, timer)
        }
    }
    
    if len(upcomingTimers) == 0 {
        fmt.Printf("✅ Нет таймеров на ближайшие %d минут\n", withinMinutes)
        return
    }
    
    fmt.Printf("🚨 Найдено %d таймеров на ближайшие %d минут:\n", len(upcomingTimers), withinMinutes)
    
    for _, timer := range upcomingTimers {
        remaining := time.Duration(timer.RemainingSeconds) * time.Second
        fmt.Printf("  ⏱️ %s - через %s", timer.TimerId, remaining.String())
        
        if timer.ProcessInstanceId != "" {
            fmt.Printf(" (процесс: %s)", timer.ProcessInstanceId)
        }
        
        fmt.Println()
    }
}
```

### Python
```python
import grpc
from datetime import datetime, timedelta
import time

import timewheel_pb2
import timewheel_pb2_grpc

def list_timers(status_filter=None, page_size=20, page=1, sort_by="created_at", sort_order="DESC"):
    channel = grpc.insecure_channel('localhost:27500')
    stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = timewheel_pb2.ListTimersRequest(
        status_filter=status_filter or "",
        page_size=page_size,
        page=page,
        sort_by=sort_by,
        sort_order=sort_order
    )
    
    try:
        response = stub.ListTimers(request, metadata=metadata)
        
        print(f"📋 Список таймеров (страница {response.page} из {response.total_pages})")
        print(f"📊 Всего найдено: {response.total_count}")
        
        if status_filter:
            print(f"🔍 Фильтр: {status_filter}")
        
        print(f"📄 На странице: {len(response.timers)} таймеров\n")
        
        if not response.timers:
            print("📭 Таймеры не найдены")
            return []
        
        # Выводим информацию о каждом таймере
        for i, timer in enumerate(response.timers, 1):
            print(f"🔸 {i}. {timer.timer_id}")
            print(f"   📌 Тип: {timer.timer_type}")
            print(f"   📊 Статус: {timer.status}")
            
            if timer.process_instance_id:
                print(f"   🔄 Процесс: {timer.process_instance_id}")
            
            if timer.element_id:
                print(f"   🎯 Элемент: {timer.element_id}")
            
            # Времена
            created_time = datetime.fromtimestamp(timer.created_at)
            scheduled_time = datetime.fromtimestamp(timer.scheduled_at)
            
            print(f"   📅 Создан: {created_time.strftime('%H:%M:%S')}")
            print(f"   ⏰ Сработает: {scheduled_time.strftime('%H:%M:%S')}")
            
            if timer.remaining_seconds > 0:
                remaining_td = timedelta(seconds=timer.remaining_seconds)
                print(f"   ⏱️ Осталось: {remaining_td}")
            
            # ISO 8601 информация
            if timer.time_duration:
                print(f"   📏 Длительность: {timer.time_duration}")
            
            if timer.time_cycle:
                print(f"   🔄 Цикл: {timer.time_cycle}")
            
            print(f"   🏗️ Уровень wheel: {timer.wheel_level}")
            print()
        
        return list(response.timers)
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return []

# Класс для работы с таймерами
class TimerManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = timewheel_pb2_grpc.TimeWheelServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def get_all_timers(self):
        """Получает все таймеры со всех страниц"""
        all_timers = []
        page = 1
        
        while True:
            request = timewheel_pb2.ListTimersRequest(
                page_size=100,  # Максимальный размер страницы
                page=page,
                sort_by="created_at",
                sort_order="DESC"
            )
            
            try:
                response = self.stub.ListTimers(request, metadata=self.metadata)
                
                all_timers.extend(response.timers)
                
                if page >= response.total_pages:
                    break
                    
                page += 1
                
            except grpc.RpcError as e:
                print(f"❌ Ошибка на странице {page}: {e.details()}")
                break
        
        return all_timers
    
    def find_timers_by_process(self, process_instance_id):
        """Находит все таймеры для указанного процесса"""
        all_timers = self.get_all_timers()
        
        process_timers = [
            timer for timer in all_timers 
            if timer.process_instance_id == process_instance_id
        ]
        
        print(f"🔍 Поиск таймеров для процесса: {process_instance_id}")
        print(f"📋 Найдено: {len(process_timers)} таймеров\n")
        
        for timer in process_timers:
            status_icon = {"SCHEDULED": "⏳", "FIRED": "✅", "CANCELLED": "❌"}.get(timer.status, "❓")
            print(f"  {status_icon} {timer.timer_id} ({timer.timer_type}) - {timer.status}")
            
            if timer.element_id:
                print(f"      🎯 Элемент: {timer.element_id}")
        
        return process_timers
    
    def find_timers_by_type(self, timer_type):
        """Находит все таймеры указанного типа"""
        all_timers = self.get_all_timers()
        
        type_timers = [
            timer for timer in all_timers 
            if timer.timer_type == timer_type
        ]
        
        print(f"🔍 Поиск таймеров типа: {timer_type}")
        print(f"📋 Найдено: {len(type_timers)} таймеров\n")
        
        # Группируем по статусам
        status_groups = {}
        for timer in type_timers:
            status = timer.status
            if status not in status_groups:
                status_groups[status] = []
            status_groups[status].append(timer)
        
        for status, timers in status_groups.items():
            status_icon = {"SCHEDULED": "⏳", "FIRED": "✅", "CANCELLED": "❌"}.get(status, "❓")
            print(f"  {status_icon} {status}: {len(timers)} таймеров")
        
        return type_timers
    
    def get_upcoming_timers(self, within_minutes=30):
        """Получает таймеры, которые сработают в ближайшее время"""
        request = timewheel_pb2.ListTimersRequest(
            status_filter="SCHEDULED",
            page_size=100,
            page=1,
            sort_by="remaining_seconds",
            sort_order="ASC"
        )
        
        try:
            response = self.stub.ListTimers(request, metadata=self.metadata)
            
            threshold_seconds = within_minutes * 60
            upcoming = [
                timer for timer in response.timers 
                if 0 < timer.remaining_seconds <= threshold_seconds
            ]
            
            print(f"⏰ Таймеры на ближайшие {within_minutes} минут:")
            print(f"📋 Найдено: {len(upcoming)} таймеров\n")
            
            for timer in upcoming:
                remaining_td = timedelta(seconds=timer.remaining_seconds)
                print(f"  ⏱️ {timer.timer_id} - через {remaining_td}")
                
                if timer.process_instance_id:
                    print(f"      🔄 Процесс: {timer.process_instance_id}")
                
                if timer.timer_type:
                    print(f"      📌 Тип: {timer.timer_type}")
            
            return upcoming
            
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.code()} - {e.details()}")
            return []
    
    def get_statistics(self):
        """Получает статистику по всем таймерам"""
        all_timers = self.get_all_timers()
        
        print("📊 СТАТИСТИКА ТАЙМЕРОВ")
        print("=" * 30)
        print(f"📋 Всего таймеров: {len(all_timers)}")
        
        if not all_timers:
            print("📭 Нет таймеров в системе")
            return
        
        # Статистика по статусам
        status_counts = {}
        for timer in all_timers:
            status = timer.status
            status_counts[status] = status_counts.get(status, 0) + 1
        
        print("\n📊 По статусам:")
        for status, count in status_counts.items():
            icon = {"SCHEDULED": "⏳", "FIRED": "✅", "CANCELLED": "❌"}.get(status, "❓")
            percentage = (count / len(all_timers)) * 100
            print(f"  {icon} {status}: {count} ({percentage:.1f}%)")
        
        # Статистика по типам
        type_counts = {}
        for timer in all_timers:
            timer_type = timer.timer_type or "unknown"
            type_counts[timer_type] = type_counts.get(timer_type, 0) + 1
        
        print("\n📋 По типам:")
        for timer_type, count in sorted(type_counts.items(), key=lambda x: x[1], reverse=True):
            percentage = (count / len(all_timers)) * 100
            print(f"  📌 {timer_type}: {count} ({percentage:.1f}%)")
        
        # Статистика по уровням wheel
        level_counts = {}
        for timer in all_timers:
            level = timer.wheel_level
            level_counts[level] = level_counts.get(level, 0) + 1
        
        print("\n🏗️ По уровням wheel:")
        level_names = {0: "секунды", 1: "минуты", 2: "часы", 3: "дни", 4: "годы"}
        for level in sorted(level_counts.keys()):
            count = level_counts[level]
            level_name = level_names.get(level, f"уровень {level}")
            percentage = (count / len(all_timers)) * 100
            print(f"  🔸 {level} ({level_name}): {count} ({percentage:.1f}%)")

# Демонстрация работы с таймерами
def demonstrate_timer_management():
    print("🎮 Демонстрация управления таймерами\n")
    
    manager = TimerManager()
    
    # Общая статистика
    print("📊 Получение общей статистики...")
    manager.get_statistics()
    
    print("\n" + "="*50)
    
    # Список активных таймеров
    print("\n⏳ Активные таймеры:")
    active_timers = list_timers(status_filter="SCHEDULED", page_size=5)
    
    print("\n" + "="*50)
    
    # Ближайшие таймеры
    print("\n⏰ Ближайшие таймеры:")
    upcoming = manager.get_upcoming_timers(within_minutes=60)
    
    print("\n" + "="*50)
    
    # Поиск по процессу (если есть таймеры с process_instance_id)
    if active_timers:
        for timer in active_timers:
            if timer.process_instance_id:
                print(f"\n🔍 Поиск таймеров для процесса: {timer.process_instance_id}")
                manager.find_timers_by_process(timer.process_instance_id)
                break
    
    print("\n✅ Демонстрация завершена")

# Функции для анализа производительности
def analyze_timer_performance():
    """Анализирует производительность таймеров"""
    print("📈 АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ ТАЙМЕРОВ")
    print("=" * 50)
    
    manager = TimerManager()
    all_timers = manager.get_all_timers()
    
    if not all_timers:
        print("📭 Нет таймеров для анализа")
        return
    
    # Анализ времени жизни таймеров
    now = datetime.now().timestamp()
    lifetimes = []
    
    for timer in all_timers:
        if timer.status in ['FIRED', 'CANCELLED']:
            # Для завершенных таймеров время жизни = scheduled_at - created_at
            lifetime = timer.scheduled_at - timer.created_at
            lifetimes.append(lifetime)
    
    if lifetimes:
        avg_lifetime = sum(lifetimes) / len(lifetimes)
        min_lifetime = min(lifetimes)
        max_lifetime = max(lifetimes)
        
        print(f"⏱️ ВРЕМЕНА ЖИЗНИ ТАЙМЕРОВ (завершенные: {len(lifetimes)}):")
        print(f"   Среднее: {timedelta(seconds=avg_lifetime)}")
        print(f"   Минимум: {timedelta(seconds=min_lifetime)}")
        print(f"   Максимум: {timedelta(seconds=max_lifetime)}")
    
    # Анализ точности срабатывания
    scheduled_timers = [t for t in all_timers if t.status == 'SCHEDULED']
    overdue_count = 0
    
    for timer in scheduled_timers:
        if timer.scheduled_at < now:
            overdue_count += 1
    
    if scheduled_timers:
        overdue_percentage = (overdue_count / len(scheduled_timers)) * 100
        print(f"\n⏰ ТОЧНОСТЬ СРАБАТЫВАНИЯ:")
        print(f"   Активных таймеров: {len(scheduled_timers)}")
        print(f"   Просроченных: {overdue_count} ({overdue_percentage:.1f}%)")
        
        if overdue_percentage > 5:
            print("   ⚠️ Высокий процент просроченных таймеров!")
        else:
            print("   ✅ Нормальная точность срабатывания")

if __name__ == "__main__":
    # Простой список
    list_timers(status_filter="SCHEDULED", page_size=5)
    
    print("\n" + "="*60)
    
    # Демонстрация менеджера
    demonstrate_timer_management()
    
    print("\n" + "="*60)
    
    # Анализ производительности
    analyze_timer_performance()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'timewheel.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const timewheelProto = grpc.loadPackageDefinition(packageDefinition).atom.timewheel.v1;

async function listTimers(options = {}) {
    const client = new timewheelProto.TimeWheelService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    const request = {
        status_filter: options.statusFilter || '',
        page_size: options.pageSize || 20,
        page: options.page || 1,
        sort_by: options.sortBy || 'created_at',
        sort_order: options.sortOrder || 'DESC'
    };
    
    return new Promise((resolve, reject) => {
        client.listTimers(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log(`📋 Список таймеров (страница ${response.page} из ${response.total_pages})`);
            console.log(`📊 Всего найдено: ${response.total_count}`);
            
            if (options.statusFilter) {
                console.log(`🔍 Фильтр: ${options.statusFilter}`);
            }
            
            console.log(`📄 На странице: ${response.timers.length} таймеров\n`);
            
            if (response.timers.length === 0) {
                console.log('📭 Таймеры не найдены');
                resolve([]);
                return;
            }
            
            // Выводим информацию о каждом таймере
            response.timers.forEach((timer, index) => {
                console.log(`🔸 ${index + 1}. ${timer.timer_id}`);
                console.log(`   📌 Тип: ${timer.timer_type}`);
                console.log(`   📊 Статус: ${timer.status}`);
                
                if (timer.process_instance_id) {
                    console.log(`   🔄 Процесс: ${timer.process_instance_id}`);
                }
                
                if (timer.element_id) {
                    console.log(`   🎯 Элемент: ${timer.element_id}`);
                }
                
                // Времена
                const createdTime = new Date(timer.created_at * 1000);
                const scheduledTime = new Date(timer.scheduled_at * 1000);
                
                console.log(`   📅 Создан: ${createdTime.toLocaleTimeString()}`);
                console.log(`   ⏰ Сработает: ${scheduledTime.toLocaleTimeString()}`);
                
                if (timer.remaining_seconds > 0) {
                    const remaining = formatDuration(timer.remaining_seconds);
                    console.log(`   ⏱️ Осталось: ${remaining}`);
                }
                
                // ISO 8601 информация
                if (timer.time_duration) {
                    console.log(`   📏 Длительность: ${timer.time_duration}`);
                }
                
                if (timer.time_cycle) {
                    console.log(`   🔄 Цикл: ${timer.time_cycle}`);
                }
                
                console.log(`   🏗️ Уровень wheel: ${timer.wheel_level}`);
                console.log();
            });
            
            resolve(response.timers);
        });
    });
}

function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    const parts = [];
    if (hours > 0) parts.push(`${hours}ч`);
    if (minutes > 0) parts.push(`${minutes}м`);
    if (secs > 0 || parts.length === 0) parts.push(`${secs}с`);
    
    return parts.join(' ');
}

// Продвинутый менеджер таймеров
class AdvancedTimerManager {
    constructor() {
        this.client = new timewheelProto.TimeWheelService('localhost:27500',
            grpc.credentials.createInsecure());
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async getAllTimers() {
        const allTimers = [];
        let page = 1;
        
        while (true) {
            try {
                const response = await this.listTimersPage(page, 100);
                allTimers.push(...response.timers);
                
                if (page >= response.total_pages) {
                    break;
                }
                page++;
            } catch (error) {
                console.log(`❌ Ошибка на странице ${page}: ${error.message}`);
                break;
            }
        }
        
        return allTimers;
    }
    
    async listTimersPage(page, pageSize) {
        return new Promise((resolve, reject) => {
            const request = {
                page_size: pageSize,
                page: page,
                sort_by: 'created_at',
                sort_order: 'DESC'
            };
            
            this.client.listTimers(request, this.metadata, (error, response) => {
                if (error) reject(error);
                else resolve(response);
            });
        });
    }
    
    async findByProcess(processInstanceId) {
        console.log(`🔍 Поиск таймеров для процесса: ${processInstanceId}`);
        
        const allTimers = await this.getAllTimers();
        const processTimers = allTimers.filter(timer => 
            timer.process_instance_id === processInstanceId
        );
        
        console.log(`📋 Найдено: ${processTimers.length} таймеров\n`);
        
        const statusIcons = {
            'SCHEDULED': '⏳',
            'FIRED': '✅', 
            'CANCELLED': '❌'
        };
        
        processTimers.forEach(timer => {
            const icon = statusIcons[timer.status] || '❓';
            console.log(`  ${icon} ${timer.timer_id} (${timer.timer_type}) - ${timer.status}`);
            
            if (timer.element_id) {
                console.log(`      🎯 Элемент: ${timer.element_id}`);
            }
        });
        
        return processTimers;
    }
    
    async findByType(timerType) {
        console.log(`🔍 Поиск таймеров типа: ${timerType}`);
        
        const allTimers = await this.getAllTimers();
        const typeTimers = allTimers.filter(timer => 
            timer.timer_type === timerType
        );
        
        console.log(`📋 Найдено: ${typeTimers.length} таймеров\n`);
        
        // Группируем по статусам
        const statusGroups = {};
        typeTimers.forEach(timer => {
            const status = timer.status;
            if (!statusGroups[status]) {
                statusGroups[status] = [];
            }
            statusGroups[status].push(timer);
        });
        
        const statusIcons = {
            'SCHEDULED': '⏳',
            'FIRED': '✅',
            'CANCELLED': '❌'
        };
        
        Object.entries(statusGroups).forEach(([status, timers]) => {
            const icon = statusIcons[status] || '❓';
            console.log(`  ${icon} ${status}: ${timers.length} таймеров`);
        });
        
        return typeTimers;
    }
    
    async getUpcomingTimers(withinMinutes = 30) {
        console.log(`⏰ Таймеры на ближайшие ${withinMinutes} минут:`);
        
        return new Promise((resolve, reject) => {
            const request = {
                status_filter: 'SCHEDULED',
                page_size: 100,
                page: 1,
                sort_by: 'remaining_seconds',
                sort_order: 'ASC'
            };
            
            this.client.listTimers(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                const thresholdSeconds = withinMinutes * 60;
                const upcoming = response.timers.filter(timer =>
                    timer.remaining_seconds > 0 && timer.remaining_seconds <= thresholdSeconds
                );
                
                console.log(`📋 Найдено: ${upcoming.length} таймеров\n`);
                
                upcoming.forEach(timer => {
                    const remaining = formatDuration(timer.remaining_seconds);
                    console.log(`  ⏱️ ${timer.timer_id} - через ${remaining}`);
                    
                    if (timer.process_instance_id) {
                        console.log(`      🔄 Процесс: ${timer.process_instance_id}`);
                    }
                    
                    if (timer.timer_type) {
                        console.log(`      📌 Тип: ${timer.timer_type}`);
                    }
                });
                
                resolve(upcoming);
            });
        });
    }
    
    async getStatistics() {
        console.log('📊 СТАТИСТИКА ТАЙМЕРОВ');
        console.log('='.repeat(30));
        
        const allTimers = await this.getAllTimers();
        console.log(`📋 Всего таймеров: ${allTimers.length}`);
        
        if (allTimers.length === 0) {
            console.log('📭 Нет таймеров в системе');
            return;
        }
        
        // Статистика по статусам
        const statusCounts = {};
        allTimers.forEach(timer => {
            const status = timer.status;
            statusCounts[status] = (statusCounts[status] || 0) + 1;
        });
        
        console.log('\n📊 По статусам:');
        const statusIcons = {
            'SCHEDULED': '⏳',
            'FIRED': '✅',
            'CANCELLED': '❌'
        };
        
        Object.entries(statusCounts).forEach(([status, count]) => {
            const icon = statusIcons[status] || '❓';
            const percentage = (count / allTimers.length * 100).toFixed(1);
            console.log(`  ${icon} ${status}: ${count} (${percentage}%)`);
        });
        
        // Статистика по типам
        const typeCounts = {};
        allTimers.forEach(timer => {
            const timerType = timer.timer_type || 'unknown';
            typeCounts[timerType] = (typeCounts[timerType] || 0) + 1;
        });
        
        console.log('\n📋 По типам:');
        const sortedTypes = Object.entries(typeCounts)
            .sort(([,a], [,b]) => b - a);
        
        sortedTypes.forEach(([timerType, count]) => {
            const percentage = (count / allTimers.length * 100).toFixed(1);
            console.log(`  📌 ${timerType}: ${count} (${percentage}%)`);
        });
        
        // Статистика по уровням wheel
        const levelCounts = {};
        allTimers.forEach(timer => {
            const level = timer.wheel_level;
            levelCounts[level] = (levelCounts[level] || 0) + 1;
        });
        
        console.log('\n🏗️ По уровням wheel:');
        const levelNames = {
            0: 'секунды',
            1: 'минуты', 
            2: 'часы',
            3: 'дни',
            4: 'годы'
        };
        
        Object.keys(levelCounts)
            .sort((a, b) => parseInt(a) - parseInt(b))
            .forEach(level => {
                const count = levelCounts[level];
                const levelName = levelNames[level] || `уровень ${level}`;
                const percentage = (count / allTimers.length * 100).toFixed(1);
                console.log(`  🔸 ${level} (${levelName}): ${count} (${percentage}%)`);
            });
    }
    
    async analyzePerformance() {
        console.log('📈 АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ ТАЙМЕРОВ');
        console.log('='.repeat(50));
        
        const allTimers = await this.getAllTimers();
        
        if (allTimers.length === 0) {
            console.log('📭 Нет таймеров для анализа');
            return;
        }
        
        // Анализ времени жизни таймеров
        const completedTimers = allTimers.filter(timer =>
            ['FIRED', 'CANCELLED'].includes(timer.status)
        );
        
        if (completedTimers.length > 0) {
            const lifetimes = completedTimers.map(timer =>
                timer.scheduled_at - timer.created_at
            );
            
            const avgLifetime = lifetimes.reduce((a, b) => a + b) / lifetimes.length;
            const minLifetime = Math.min(...lifetimes);
            const maxLifetime = Math.max(...lifetimes);
            
            console.log(`⏱️ ВРЕМЕНА ЖИЗНИ ТАЙМЕРОВ (завершенные: ${completedTimers.length}):`);
            console.log(`   Среднее: ${formatDuration(avgLifetime)}`);
            console.log(`   Минимум: ${formatDuration(minLifetime)}`);
            console.log(`   Максимум: ${formatDuration(maxLifetime)}`);
        }
        
        // Анализ точности срабатывания
        const scheduledTimers = allTimers.filter(timer => timer.status === 'SCHEDULED');
        const now = Date.now() / 1000;
        const overdueCount = scheduledTimers.filter(timer =>
            timer.scheduled_at < now
        ).length;
        
        if (scheduledTimers.length > 0) {
            const overduePercentage = (overdueCount / scheduledTimers.length * 100).toFixed(1);
            
            console.log('\n⏰ ТОЧНОСТЬ СРАБАТЫВАНИЯ:');
            console.log(`   Активных таймеров: ${scheduledTimers.length}`);
            console.log(`   Просроченных: ${overdueCount} (${overduePercentage}%)`);
            
            if (overduePercentage > 5) {
                console.log('   ⚠️ Высокий процент просроченных таймеров!');
            } else {
                console.log('   ✅ Нормальная точность срабатывания');
            }
        }
        
        // Анализ распределения по процессам
        const processTimers = {};
        allTimers.forEach(timer => {
            if (timer.process_instance_id) {
                const processId = timer.process_instance_id;
                processTimers[processId] = (processTimers[processId] || 0) + 1;
            }
        });
        
        if (Object.keys(processTimers).length > 0) {
            const processCount = Object.keys(processTimers).length;
            const avgTimersPerProcess = Object.values(processTimers).reduce((a, b) => a + b) / processCount;
            const maxTimersPerProcess = Math.max(...Object.values(processTimers));
            
            console.log('\n🔄 РАСПРЕДЕЛЕНИЕ ПО ПРОЦЕССАМ:');
            console.log(`   Процессов с таймерами: ${processCount}`);
            console.log(`   Среднее таймеров на процесс: ${avgTimersPerProcess.toFixed(1)}`);
            console.log(`   Максимум таймеров в одном процессе: ${maxTimersPerProcess}`);
        }
    }
}

// Демонстрация всех возможностей
async function demonstrateAdvancedTimerManagement() {
    console.log('🚀 Демонстрация продвинутого управления таймерами\n');
    
    const manager = new AdvancedTimerManager();
    
    try {
        // Общая статистика
        await manager.getStatistics();
        
        console.log('\n' + '='.repeat(60));
        
        // Список активных таймеров
        console.log('\n⏳ Активные таймеры (первые 5):');
        await listTimers({
            statusFilter: 'SCHEDULED',
            pageSize: 5,
            sortBy: 'remaining_seconds',
            sortOrder: 'ASC'
        });
        
        console.log('\n' + '='.repeat(60));
        
        // Ближайшие таймеры
        await manager.getUpcomingTimers(60);
        
        console.log('\n' + '='.repeat(60));
        
        // Анализ производительности
        await manager.analyzePerformance();
        
    } catch (error) {
        console.error('❌ Ошибка в демонстрации:', error.message);
    }
}

// Основная демонстрация
async function main() {
    try {
        // Простой список
        console.log('📋 Простой список таймеров:\n');
        await listTimers({ statusFilter: 'SCHEDULED', pageSize: 3 });
        
        console.log('\n' + '='.repeat(80));
        
        // Продвинутые функции
        await demonstrateAdvancedTimerManagement();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Поля сортировки

### Доступные поля
- **`created_at`** - Время создания (по умолчанию)
- **`scheduled_at`** - Время срабатывания
- **`remaining_seconds`** - Время до срабатывания
- **`timer_id`** - Алфавитная сортировка по ID
- **`timer_type`** - Сортировка по типу

### Порядок сортировки
- **`ASC`** - По возрастанию
- **`DESC`** - По убыванию (по умолчанию)

## Фильтры статуса

### Доступные статусы
- **`SCHEDULED`** - Активные таймеры (ожидают срабатывания)
- **`FIRED`** - Сработавшие таймеры
- **`CANCELLED`** - Отмененные таймеры

## Wheel уровни

### Иерархия уровней (0-4)
- **0**: Секунды (0-59)
- **1**: Минуты (0-59)
- **2**: Часы (0-23)
- **3**: Дни (0-30)
- **4**: Годы (0-99)

## Применение

### BPMN Process Monitoring
```javascript
// Мониторинг таймеров процесса
const processTimers = await manager.findByProcess('proc-123');
console.log(`Найдено ${processTimers.length} таймеров для процесса`);
```

### Alert System
```python
# Система предупреждений
upcoming = manager.get_upcoming_timers(within_minutes=5)
if upcoming:
    send_alert(f"Скоро сработает {len(upcoming)} таймеров!")
```

### Performance Analysis
```go
// Анализ производительности
timers := listTimers("FIRED", 100, 1)
avgLifetime := calculateAverageLifetime(timers)
```

### Resource Management
```javascript
// Управление ресурсами
const stats = await manager.getStatistics();
if (stats.scheduledCount > 1000) {
    console.log('⚠️ Высокая нагрузка на систему таймеров');
}
```

## Пагинация

### Параметры
- **page_size**: 1-1000 (по умолчанию: 20)
- **page**: ≥1 (по умолчанию: 1)
- **total_pages**: Автоматически рассчитывается

### Навигация
```javascript
// Получение всех страниц
let page = 1;
do {
    const response = await listTimers({ page, pageSize: 50 });
    processTimers(response.timers);
    page++;
} while (page <= response.total_pages);
```

## Связанные методы
- [AddTimer](add-timer.md) - Создание новых таймеров
- [GetTimerStatus](get-timer-status.md) - Подробности конкретного таймера
- [RemoveTimer](remove-timer.md) - Удаление найденных таймеров
- [GetTimeWheelStats](get-timewheel-stats.md) - Общая статистика системы
