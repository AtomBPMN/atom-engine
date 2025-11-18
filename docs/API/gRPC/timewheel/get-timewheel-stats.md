# GetTimeWheelStats

## Описание
Получает подробную статистику работы системы timewheel, включая количество таймеров по статусам, производительность и распределение по типам.

## Синтаксис
```protobuf
rpc GetTimeWheelStats(GetTimeWheelStatsRequest) returns (GetTimeWheelStatsResponse);
```

## Package
```protobuf
package atom.timewheel.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer` или `*`

## Параметры запроса

### GetTimeWheelStatsRequest
```protobuf
message GetTimeWheelStatsRequest {}
```

## Параметры ответа

### GetTimeWheelStatsResponse
```protobuf
message GetTimeWheelStatsResponse {
  int32 total_timers = 1;           // Общее количество таймеров
  int32 pending_timers = 2;         // Активных таймеров
  int32 fired_timers = 3;           // Сработавших таймеров
  int32 cancelled_timers = 4;       // Отмененных таймеров
  int64 current_tick = 5;           // Текущий тик системы
  int32 slots_count = 6;            // Количество слотов в wheel
  map<string, int32> timer_types = 7; // Типы таймеров и их количество
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
    "sort"
    
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
    
    // Получаем статистику timewheel
    response, err := client.GetTimeWheelStats(ctx, &pb.GetTimeWheelStatsRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("📊 СТАТИСТИКА TIMEWHEEL")
    fmt.Println("=" * 40)
    
    // Общая информация
    fmt.Printf("🎯 Всего таймеров: %d\n", response.TotalTimers)
    fmt.Printf("⏳ Активных: %d\n", response.PendingTimers)
    fmt.Printf("✅ Сработало: %d\n", response.FiredTimers)
    fmt.Printf("❌ Отменено: %d\n", response.CancelledTimers)
    
    // Системная информация
    fmt.Printf("\n🔧 СИСТЕМНАЯ ИНФОРМАЦИЯ:\n")
    fmt.Printf("   Текущий тик: %d\n", response.CurrentTick)
    fmt.Printf("   Слотов в wheel: %d\n", response.SlotsCount)
    
    // Статистика по типам
    if len(response.TimerTypes) > 0 {
        fmt.Printf("\n📋 ТИПЫ ТАЙМЕРОВ:\n")
        
        // Сортируем типы по количеству (по убыванию)
        type timerTypeStat struct {
            name  string
            count int32
        }
        
        var typeStats []timerTypeStat
        for timerType, count := range response.TimerTypes {
            typeStats = append(typeStats, timerTypeStat{name: timerType, count: count})
        }
        
        sort.Slice(typeStats, func(i, j int) bool {
            return typeStats[i].count > typeStats[j].count
        })
        
        for _, stat := range typeStats {
            percentage := float64(stat.count) / float64(response.TotalTimers) * 100
            fmt.Printf("   📌 %-20s: %3d (%4.1f%%)\n", stat.name, stat.count, percentage)
        }
    }
    
    // Анализ эффективности
    fmt.Printf("\n⚡ АНАЛИЗ ЭФФЕКТИВНОСТИ:\n")
    if response.TotalTimers > 0 {
        successRate := float64(response.FiredTimers) / float64(response.TotalTimers) * 100
        cancelRate := float64(response.CancelledTimers) / float64(response.TotalTimers) * 100
        
        fmt.Printf("   Успешность: %.1f%%\n", successRate)
        fmt.Printf("   Отмены: %.1f%%\n", cancelRate)
        
        if response.PendingTimers > 0 {
            loadFactor := float64(response.PendingTimers) / float64(response.SlotsCount)
            fmt.Printf("   Загрузка слотов: %.2f\n", loadFactor)
            
            if loadFactor > 0.75 {
                fmt.Printf("   ⚠️ Высокая загрузка! Рекомендуется масштабирование\n")
            } else if loadFactor < 0.1 {
                fmt.Printf("   💡 Низкая загрузка, система работает эффективно\n")
            }
        }
    } else {
        fmt.Printf("   📭 Система пуста\n")
    }
}

// Мониторинг производительности в реальном времени
func monitorPerformance(client pb.TimeWheelServiceClient, ctx context.Context, duration time.Duration) {
    fmt.Printf("📈 Мониторинг производительности на %v\n", duration)
    fmt.Printf("%-10s | %-8s | %-8s | %-8s | %-10s\n", "Время", "Всего", "Активных", "Успешно", "Загрузка")
    fmt.Printf("%s\n", strings.Repeat("-", 50))
    
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(duration)
    startTime := time.Now()
    
    for {
        select {
        case <-timeout:
            fmt.Printf("✅ Мониторинг завершен\n")
            return
            
        case <-ticker.C:
            response, err := client.GetTimeWheelStats(ctx, &pb.GetTimeWheelStatsRequest{})
            if err != nil {
                fmt.Printf("❌ Ошибка получения статистики: %v\n", err)
                continue
            }
            
            elapsed := time.Since(startTime)
            loadFactor := float64(response.PendingTimers) / float64(response.SlotsCount)
            
            fmt.Printf("%-10s | %-8d | %-8d | %-8d | %-10.3f\n",
                elapsed.Truncate(time.Second).String(),
                response.TotalTimers,
                response.PendingTimers,
                response.FiredTimers,
                loadFactor)
        }
    }
}
```

### Python
```python
import grpc
import time
from datetime import datetime
import threading
import json

import timewheel_pb2
import timewheel_pb2_grpc

def get_timewheel_stats():
    channel = grpc.insecure_channel('localhost:27500')
    stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = timewheel_pb2.GetTimeWheelStatsRequest()
    
    try:
        response = stub.GetTimeWheelStats(request, metadata=metadata)
        
        print("📊 СТАТИСТИКА TIMEWHEEL")
        print("=" * 40)
        
        # Основные счетчики
        print(f"🎯 Всего таймеров: {response.total_timers}")
        print(f"⏳ Активных: {response.pending_timers}")
        print(f"✅ Сработало: {response.fired_timers}")
        print(f"❌ Отменено: {response.cancelled_timers}")
        
        # Системная информация
        print(f"\n🔧 СИСТЕМНАЯ ИНФОРМАЦИЯ:")
        print(f"   Текущий тик: {response.current_tick}")
        print(f"   Слотов в wheel: {response.slots_count}")
        
        # Распределение по типам
        if response.timer_types:
            print(f"\n📋 ТИПЫ ТАЙМЕРОВ:")
            
            # Сортируем по убыванию количества
            sorted_types = sorted(response.timer_types.items(), 
                                key=lambda x: x[1], reverse=True)
            
            for timer_type, count in sorted_types:
                if response.total_timers > 0:
                    percentage = (count / response.total_timers) * 100
                    print(f"   📌 {timer_type:<20}: {count:3d} ({percentage:4.1f}%)")
                else:
                    print(f"   📌 {timer_type:<20}: {count:3d}")
        
        # Анализ производительности
        print(f"\n⚡ АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ:")
        if response.total_timers > 0:
            success_rate = (response.fired_timers / response.total_timers) * 100
            cancel_rate = (response.cancelled_timers / response.total_timers) * 100
            
            print(f"   Успешность: {success_rate:.1f}%")
            print(f"   Отмены: {cancel_rate:.1f}%")
            
            if response.pending_timers > 0:
                load_factor = response.pending_timers / response.slots_count
                print(f"   Загрузка слотов: {load_factor:.3f}")
                
                if load_factor > 0.75:
                    print("   ⚠️ Высокая загрузка! Требуется оптимизация")
                elif load_factor < 0.1:
                    print("   💡 Низкая загрузка, система работает оптимально")
                else:
                    print("   ✅ Нормальная загрузка")
        else:
            print("   📭 Система пуста")
        
        return {
            'total_timers': response.total_timers,
            'pending_timers': response.pending_timers,
            'fired_timers': response.fired_timers,
            'cancelled_timers': response.cancelled_timers,
            'current_tick': response.current_tick,
            'slots_count': response.slots_count,
            'timer_types': dict(response.timer_types)
        }
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Система мониторинга метрик
class TimeWheelMetrics:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = timewheel_pb2_grpc.TimeWheelServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
        self.history = []
        self.monitoring = False
        self.monitor_thread = None
    
    def collect_metrics(self):
        """Собирает текущие метрики"""
        try:
            request = timewheel_pb2.GetTimeWheelStatsRequest()
            response = self.stub.GetTimeWheelStats(request, metadata=self.metadata)
            
            metrics = {
                'timestamp': datetime.now(),
                'total_timers': response.total_timers,
                'pending_timers': response.pending_timers,
                'fired_timers': response.fired_timers,
                'cancelled_timers': response.cancelled_timers,
                'current_tick': response.current_tick,
                'slots_count': response.slots_count,
                'timer_types': dict(response.timer_types),
                'load_factor': response.pending_timers / response.slots_count if response.slots_count > 0 else 0
            }
            
            self.history.append(metrics)
            return metrics
            
        except grpc.RpcError as e:
            print(f"❌ Ошибка сбора метрик: {e.details()}")
            return None
    
    def start_monitoring(self, interval=10):
        """Запускает непрерывный мониторинг"""
        if self.monitoring:
            print("⚠️ Мониторинг уже запущен")
            return
        
        self.monitoring = True
        print(f"🚀 Запуск мониторинга каждые {interval} секунд")
        
        def monitor_loop():
            while self.monitoring:
                metrics = self.collect_metrics()
                if metrics:
                    self._log_metrics(metrics)
                time.sleep(interval)
        
        self.monitor_thread = threading.Thread(target=monitor_loop, daemon=True)
        self.monitor_thread.start()
    
    def stop_monitoring(self):
        """Останавливает мониторинг"""
        if self.monitoring:
            self.monitoring = False
            print("🛑 Мониторинг остановлен")
    
    def _log_metrics(self, metrics):
        """Логирует метрики с временной меткой"""
        timestamp = metrics['timestamp'].strftime('%H:%M:%S')
        print(f"[{timestamp}] "
              f"📊 Всего: {metrics['total_timers']}, "
              f"⏳ Активных: {metrics['pending_timers']}, "
              f"📈 Загрузка: {metrics['load_factor']:.3f}")
    
    def generate_report(self, last_minutes=None):
        """Генерирует отчет за указанный период"""
        if not self.history:
            print("📭 Нет данных для отчета")
            return
        
        # Фильтруем данные по времени
        if last_minutes:
            cutoff_time = datetime.now() - timedelta(minutes=last_minutes)
            data = [m for m in self.history if m['timestamp'] >= cutoff_time]
        else:
            data = self.history
        
        if not data:
            print(f"📭 Нет данных за последние {last_minutes} минут")
            return
        
        print(f"\n📋 ОТЧЕТ ЗА ПЕРИОД")
        print(f"   Период: {len(data)} точек данных")
        print(f"   С: {data[0]['timestamp'].strftime('%H:%M:%S')}")
        print(f"   По: {data[-1]['timestamp'].strftime('%H:%M:%S')}")
        
        # Статистика по активности
        total_start = data[0]['total_timers']
        total_end = data[-1]['total_timers']
        print(f"\n📊 ИЗМЕНЕНИЯ:")
        print(f"   Всего таймеров: {total_start} → {total_end} ({total_end - total_start:+d})")
        
        # Пиковые значения
        max_pending = max(d['pending_timers'] for d in data)
        max_load = max(d['load_factor'] for d in data)
        avg_load = sum(d['load_factor'] for d in data) / len(data)
        
        print(f"\n🏔️ ПИКОВЫЕ ЗНАЧЕНИЯ:")
        print(f"   Максимум активных: {max_pending}")
        print(f"   Пиковая загрузка: {max_load:.3f}")
        print(f"   Средняя загрузка: {avg_load:.3f}")
        
        # Анализ типов таймеров
        all_types = set()
        for d in data:
            all_types.update(d['timer_types'].keys())
        
        if all_types:
            print(f"\n📈 ДИНАМИКА ТИПОВ ТАЙМЕРОВ:")
            for timer_type in sorted(all_types):
                counts = [d['timer_types'].get(timer_type, 0) for d in data]
                if any(counts):
                    print(f"   {timer_type}: {counts[0]} → {counts[-1]} "
                          f"(пик: {max(counts)})")
    
    def export_metrics(self, filename=None):
        """Экспортирует метрики в JSON"""
        if not self.history:
            print("📭 Нет данных для экспорта")
            return
        
        if not filename:
            filename = f"timewheel_metrics_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        
        # Подготовка данных для экспорта
        export_data = []
        for metrics in self.history:
            export_metrics = dict(metrics)
            export_metrics['timestamp'] = metrics['timestamp'].isoformat()
            export_data.append(export_metrics)
        
        try:
            with open(filename, 'w', encoding='utf-8') as f:
                json.dump(export_data, f, indent=2, ensure_ascii=False)
            
            print(f"💾 Метрики экспортированы в {filename}")
            print(f"   Записей: {len(export_data)}")
            
        except Exception as e:
            print(f"❌ Ошибка экспорта: {e}")
    
    def clear_history(self):
        """Очищает историю метрик"""
        self.history.clear()
        print("🧹 История метрик очищена")

# Демонстрация системы мониторинга
def demonstrate_metrics_monitoring():
    print("📈 Демонстрация мониторинга метрик TimeWheel\n")
    
    metrics = TimeWheelMetrics()
    
    # Собираем базовые метрики
    print("📊 Текущие метрики:")
    current = metrics.collect_metrics()
    
    if current:
        print(f"   Общее состояние на {current['timestamp'].strftime('%H:%M:%S')}")
        print(f"   Активных таймеров: {current['pending_timers']}")
        print(f"   Загрузка системы: {current['load_factor']:.3f}")
    
    # Запускаем кратковременный мониторинг
    print(f"\n🚀 Запуск мониторинга на 30 секунд...")
    metrics.start_monitoring(interval=5)
    
    # Даем поработать мониторингу
    time.sleep(30)
    
    metrics.stop_monitoring()
    
    # Генерируем отчет
    print(f"\n📋 Генерация отчета:")
    metrics.generate_report()
    
    # Опционально экспортируем данные
    # metrics.export_metrics()

# Диагностические утилиты
def diagnose_timewheel_health():
    """Диагностика здоровья системы TimeWheel"""
    print("🏥 ДИАГНОСТИКА СИСТЕМЫ TIMEWHEEL")
    print("=" * 50)
    
    stats = get_timewheel_stats()
    if not stats:
        print("❌ Не удалось получить статистику")
        return
    
    issues = []
    recommendations = []
    
    # Проверка загрузки
    if stats['slots_count'] > 0:
        load_factor = stats['pending_timers'] / stats['slots_count']
        
        if load_factor > 0.8:
            issues.append("🔴 Критически высокая загрузка слотов")
            recommendations.append("• Увеличить количество слотов")
            recommendations.append("• Оптимизировать количество одновременных таймеров")
        elif load_factor > 0.5:
            issues.append("🟡 Повышенная загрузка слотов")
            recommendations.append("• Мониторить рост загрузки")
    
    # Проверка соотношения отмен
    if stats['total_timers'] > 0:
        cancel_rate = stats['cancelled_timers'] / stats['total_timers']
        
        if cancel_rate > 0.3:
            issues.append("🔴 Высокий уровень отмен таймеров")
            recommendations.append("• Проверить логику отмены таймеров")
            recommendations.append("• Оптимизировать жизненный цикл процессов")
        elif cancel_rate > 0.15:
            issues.append("🟡 Повышенный уровень отмен")
    
    # Проверка активности
    if stats['pending_timers'] == 0 and stats['total_timers'] > 0:
        issues.append("🔵 Нет активных таймеров")
        recommendations.append("• Проверить, правильно ли завершились все процессы")
    
    # Вывод результатов
    if issues:
        print("⚠️ ОБНАРУЖЕННЫЕ ПРОБЛЕМЫ:")
        for issue in issues:
            print(f"   {issue}")
    else:
        print("✅ СИСТЕМА РАБОТАЕТ НОРМАЛЬНО")
    
    if recommendations:
        print(f"\n💡 РЕКОМЕНДАЦИИ:")
        for rec in recommendations:
            print(f"   {rec}")
    
    print(f"\n📊 КРАТКАЯ СВОДКА:")
    print(f"   Общая производительность: {'🟢 Хорошо' if not issues else '🟡 Требует внимания' if len(issues) < 3 else '🔴 Критично'}")
    print(f"   Активных таймеров: {stats['pending_timers']}")
    print(f"   Эффективность: {(stats['fired_timers'] / max(stats['total_timers'], 1)) * 100:.1f}%")

if __name__ == "__main__":
    # Простая статистика
    get_timewheel_stats()
    
    print("\n" + "="*60)
    
    # Диагностика
    diagnose_timewheel_health()
    
    print("\n" + "="*60)
    
    # Демонстрация мониторинга
    demonstrate_metrics_monitoring()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'timewheel.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const timewheelProto = grpc.loadPackageDefinition(packageDefinition).atom.timewheel.v1;

async function getTimeWheelStats() {
    const client = new timewheelProto.TimeWheelService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.getTimeWheelStats(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log('📊 СТАТИСТИКА TIMEWHEEL');
            console.log('='.repeat(40));
            
            // Основные счетчики
            console.log(`🎯 Всего таймеров: ${response.total_timers}`);
            console.log(`⏳ Активных: ${response.pending_timers}`);
            console.log(`✅ Сработало: ${response.fired_timers}`);
            console.log(`❌ Отменено: ${response.cancelled_timers}`);
            
            // Системная информация
            console.log('\n🔧 СИСТЕМНАЯ ИНФОРМАЦИЯ:');
            console.log(`   Текущий тик: ${response.current_tick}`);
            console.log(`   Слотов в wheel: ${response.slots_count}`);
            
            // Типы таймеров
            if (Object.keys(response.timer_types).length > 0) {
                console.log('\n📋 ТИПЫ ТАЙМЕРОВ:');
                
                // Сортируем по убыванию
                const sortedTypes = Object.entries(response.timer_types)
                    .sort(([,a], [,b]) => b - a);
                
                sortedTypes.forEach(([type, count]) => {
                    const percentage = response.total_timers > 0 
                        ? (count / response.total_timers * 100).toFixed(1)
                        : '0.0';
                    console.log(`   📌 ${type.padEnd(20)}: ${count.toString().padStart(3)} (${percentage}%)`);
                });
            }
            
            // Анализ производительности
            console.log('\n⚡ АНАЛИЗ ПРОИЗВОДИТЕЛЬНОСТИ:');
            if (response.total_timers > 0) {
                const successRate = (response.fired_timers / response.total_timers * 100).toFixed(1);
                const cancelRate = (response.cancelled_timers / response.total_timers * 100).toFixed(1);
                
                console.log(`   Успешность: ${successRate}%`);
                console.log(`   Отмены: ${cancelRate}%`);
                
                if (response.pending_timers > 0) {
                    const loadFactor = response.pending_timers / response.slots_count;
                    console.log(`   Загрузка слотов: ${loadFactor.toFixed(3)}`);
                    
                    if (loadFactor > 0.75) {
                        console.log('   ⚠️ Высокая загрузка! Требуется масштабирование');
                    } else if (loadFactor < 0.1) {
                        console.log('   💡 Низкая загрузка, система эффективна');
                    } else {
                        console.log('   ✅ Нормальная загрузка');
                    }
                }
            } else {
                console.log('   📭 Система пуста');
            }
            
            resolve({
                totalTimers: response.total_timers,
                pendingTimers: response.pending_timers,
                firedTimers: response.fired_timers,
                cancelledTimers: response.cancelled_timers,
                currentTick: response.current_tick,
                slotsCount: response.slots_count,
                timerTypes: response.timer_types
            });
        });
    });
}

// Класс для продвинутой аналитики TimeWheel
class TimeWheelAnalytics {
    constructor() {
        this.history = [];
        this.monitoringInterval = null;
        this.isMonitoring = false;
    }
    
    async collectSnapshot() {
        try {
            const stats = await getTimeWheelStats();
            const snapshot = {
                ...stats,
                timestamp: new Date(),
                loadFactor: stats.pendingTimers / stats.slotsCount
            };
            
            this.history.push(snapshot);
            
            // Ограничиваем историю последними 1000 записями
            if (this.history.length > 1000) {
                this.history = this.history.slice(-1000);
            }
            
            return snapshot;
        } catch (error) {
            console.log(`❌ Ошибка сбора данных: ${error.message}`);
            return null;
        }
    }
    
    startContinuousMonitoring(intervalMs = 10000) {
        if (this.isMonitoring) {
            console.log('⚠️ Мониторинг уже запущен');
            return;
        }
        
        this.isMonitoring = true;
        console.log(`🚀 Запуск непрерывного мониторинга (интервал: ${intervalMs}мс)`);
        
        console.log('Время      | Всего | Актив | Успеш | Отмен | Загрузка | Тенденция');
        console.log('-'.repeat(70));
        
        this.monitoringInterval = setInterval(async () => {
            const snapshot = await this.collectSnapshot();
            if (snapshot) {
                this.logSnapshot(snapshot);
            }
        }, intervalMs);
    }
    
    stopContinuousMonitoring() {
        if (!this.isMonitoring) return;
        
        clearInterval(this.monitoringInterval);
        this.isMonitoring = false;
        console.log('🛑 Мониторинг остановлен');
    }
    
    logSnapshot(snapshot) {
        const time = snapshot.timestamp.toLocaleTimeString();
        const trend = this.calculateTrend();
        const trendIcon = trend > 0 ? '📈' : trend < 0 ? '📉' : '➡️';
        
        console.log(`${time} | ${snapshot.totalTimers.toString().padStart(5)} | ` +
                   `${snapshot.pendingTimers.toString().padStart(5)} | ` +
                   `${snapshot.firedTimers.toString().padStart(5)} | ` +
                   `${snapshot.cancelledTimers.toString().padStart(5)} | ` +
                   `${snapshot.loadFactor.toFixed(3).padStart(8)} | ${trendIcon}`);
    }
    
    calculateTrend() {
        if (this.history.length < 3) return 0;
        
        const recent = this.history.slice(-3);
        const oldLoad = recent[0].loadFactor;
        const newLoad = recent[2].loadFactor;
        
        return newLoad - oldLoad;
    }
    
    generatePerformanceReport() {
        if (this.history.length < 2) {
            console.log('📭 Недостаточно данных для отчета');
            return;
        }
        
        console.log('\n📋 ОТЧЕТ О ПРОИЗВОДИТЕЛЬНОСТИ');
        console.log('='.repeat(50));
        
        const first = this.history[0];
        const last = this.history[this.history.length - 1];
        const duration = (last.timestamp - first.timestamp) / 1000 / 60; // минуты
        
        console.log(`📊 Период анализа: ${duration.toFixed(1)} минут`);
        console.log(`📈 Точек данных: ${this.history.length}`);
        
        // Изменения за период
        const totalChange = last.totalTimers - first.totalTimers;
        const firedChange = last.firedTimers - first.firedTimers;
        const cancelledChange = last.cancelledTimers - first.cancelledTimers;
        
        console.log('\n📊 ИЗМЕНЕНИЯ ЗА ПЕРИОД:');
        console.log(`   Всего таймеров: ${first.totalTimers} → ${last.totalTimers} (${totalChange >= 0 ? '+' : ''}${totalChange})`);
        console.log(`   Сработало: +${firedChange}`);
        console.log(`   Отменено: +${cancelledChange}`);
        
        // Статистика загрузки
        const loadFactors = this.history.map(h => h.loadFactor);
        const avgLoad = loadFactors.reduce((a, b) => a + b) / loadFactors.length;
        const maxLoad = Math.max(...loadFactors);
        const minLoad = Math.min(...loadFactors);
        
        console.log('\n⚡ СТАТИСТИКА ЗАГРУЗКИ:');
        console.log(`   Средняя: ${avgLoad.toFixed(3)}`);
        console.log(`   Максимум: ${maxLoad.toFixed(3)}`);
        console.log(`   Минимум: ${minLoad.toFixed(3)}`);
        
        // Рекомендации
        this.generateRecommendations(avgLoad, maxLoad, last);
    }
    
    generateRecommendations(avgLoad, maxLoad, currentStats) {
        console.log('\n💡 РЕКОМЕНДАЦИИ:');
        
        const recommendations = [];
        
        if (maxLoad > 0.8) {
            recommendations.push('🔴 Критическая загрузка! Увеличить количество слотов');
        } else if (avgLoad > 0.6) {
            recommendations.push('🟡 Высокая средняя загрузка, подготовить масштабирование');
        }
        
        if (currentStats.totalTimers > 0) {
            const cancelRate = currentStats.cancelledTimers / currentStats.totalTimers;
            if (cancelRate > 0.2) {
                recommendations.push('⚠️ Высокий уровень отмен таймеров, проверить логику');
            }
        }
        
        if (avgLoad < 0.1 && currentStats.totalTimers > 100) {
            recommendations.push('💡 Низкая загрузка, можно оптимизировать ресурсы');
        }
        
        if (recommendations.length === 0) {
            recommendations.push('✅ Система работает оптимально');
        }
        
        recommendations.forEach((rec, index) => {
            console.log(`   ${index + 1}. ${rec}`);
        });
    }
    
    exportToCSV(filename) {
        if (this.history.length === 0) {
            console.log('📭 Нет данных для экспорта');
            return;
        }
        
        const fs = require('fs');
        
        const csvHeader = 'timestamp,total_timers,pending_timers,fired_timers,cancelled_timers,load_factor\n';
        const csvRows = this.history.map(h => 
            `${h.timestamp.toISOString()},${h.totalTimers},${h.pendingTimers},${h.firedTimers},${h.cancelledTimers},${h.loadFactor.toFixed(6)}`
        );
        
        const csvContent = csvHeader + csvRows.join('\n');
        
        try {
            fs.writeFileSync(filename || 'timewheel_stats.csv', csvContent);
            console.log(`💾 Данные экспортированы в ${filename || 'timewheel_stats.csv'}`);
            console.log(`   Записей: ${this.history.length}`);
        } catch (error) {
            console.log(`❌ Ошибка экспорта: ${error.message}`);
        }
    }
}

// Система алертов на основе метрик
class TimeWheelAlerts {
    constructor(analytics) {
        this.analytics = analytics;
        this.thresholds = {
            highLoad: 0.75,
            criticalLoad: 0.9,
            highCancelRate: 0.25,
            lowActivity: 0.01
        };
        this.alertHistory = [];
    }
    
    checkAlerts() {
        if (this.analytics.history.length === 0) return;
        
        const current = this.analytics.history[this.analytics.history.length - 1];
        const alerts = [];
        
        // Проверка загрузки
        if (current.loadFactor >= this.thresholds.criticalLoad) {
            alerts.push({
                level: 'CRITICAL',
                message: `Критическая загрузка: ${(current.loadFactor * 100).toFixed(1)}%`,
                metric: 'load_factor',
                value: current.loadFactor
            });
        } else if (current.loadFactor >= this.thresholds.highLoad) {
            alerts.push({
                level: 'WARNING',
                message: `Высокая загрузка: ${(current.loadFactor * 100).toFixed(1)}%`,
                metric: 'load_factor', 
                value: current.loadFactor
            });
        }
        
        // Проверка отмен
        if (current.totalTimers > 0) {
            const cancelRate = current.cancelledTimers / current.totalTimers;
            if (cancelRate >= this.thresholds.highCancelRate) {
                alerts.push({
                    level: 'WARNING',
                    message: `Высокий уровень отмен: ${(cancelRate * 100).toFixed(1)}%`,
                    metric: 'cancel_rate',
                    value: cancelRate
                });
            }
        }
        
        // Проверка низкой активности
        if (current.loadFactor <= this.thresholds.lowActivity && current.totalTimers > 0) {
            alerts.push({
                level: 'INFO',
                message: `Низкая активность системы: ${(current.loadFactor * 100).toFixed(1)}%`,
                metric: 'load_factor',
                value: current.loadFactor
            });
        }
        
        // Логируем новые алерты
        alerts.forEach(alert => {
            const alertKey = `${alert.metric}_${alert.level}`;
            const lastAlert = this.alertHistory.find(a => a.key === alertKey);
            
            // Показываем алерт только если прошло достаточно времени с последнего
            if (!lastAlert || (Date.now() - lastAlert.timestamp) > 60000) { // 1 минута
                this.logAlert(alert);
                this.alertHistory.push({
                    key: alertKey,
                    timestamp: Date.now(),
                    ...alert
                });
            }
        });
        
        return alerts;
    }
    
    logAlert(alert) {
        const icons = {
            'CRITICAL': '🚨',
            'WARNING': '⚠️',
            'INFO': 'ℹ️'
        };
        
        const icon = icons[alert.level] || '📊';
        const timestamp = new Date().toLocaleTimeString();
        
        console.log(`[${timestamp}] ${icon} ${alert.level}: ${alert.message}`);
    }
}

// Демонстрация полной аналитики
async function demonstrateFullAnalytics() {
    console.log('🔬 Демонстрация полной аналитики TimeWheel\n');
    
    const analytics = new TimeWheelAnalytics();
    const alerts = new TimeWheelAlerts(analytics);
    
    console.log('📊 Сбор базовой статистики...');
    await analytics.collectSnapshot();
    
    console.log('\n🚀 Запуск мониторинга с проверкой алертов на 30 секунд...');
    
    // Запускаем мониторинг
    analytics.startContinuousMonitoring(5000);
    
    // Проверяем алерты каждые 10 секунд
    const alertInterval = setInterval(() => {
        alerts.checkAlerts();
    }, 10000);
    
    // Останавливаем через 30 секунд
    setTimeout(() => {
        analytics.stopContinuousMonitoring();
        clearInterval(alertInterval);
        
        console.log('\n📋 Генерация финального отчета...');
        analytics.generatePerformanceReport();
        
        // Опционально экспортируем данные
        // analytics.exportToCSV('timewheel_demo.csv');
        
    }, 30000);
}

// Основная демонстрация
async function main() {
    try {
        // Простая статистика
        console.log('📊 Получение базовой статистики:\n');
        await getTimeWheelStats();
        
        console.log('\n' + '='.repeat(60));
        
        // Полная аналитика
        await demonstrateFullAnalytics();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Метрики TimeWheel

### Основные счетчики
- **total_timers**: Общее количество таймеров
- **pending_timers**: Активных (ожидающих) таймеров
- **fired_timers**: Сработавших таймеров  
- **cancelled_timers**: Отмененных таймеров

### Системные метрики
- **current_tick**: Текущий тик системы
- **slots_count**: Количество слотов в wheel
- **timer_types**: Распределение по типам

### Расчетные показатели
- **Load Factor**: `pending_timers / slots_count`
- **Success Rate**: `fired_timers / total_timers * 100%`
- **Cancel Rate**: `cancelled_timers / total_timers * 100%`

## Применение

### Мониторинг производительности
```javascript
// Проверка загрузки системы
const stats = await getTimeWheelStats();
const loadFactor = stats.pendingTimers / stats.slotsCount;

if (loadFactor > 0.8) {
    console.log('⚠️ Система перегружена!');
}
```

### Анализ типов таймеров
```python
# Определение наиболее используемых типов
stats = get_timewheel_stats()
most_used = max(stats['timer_types'].items(), key=lambda x: x[1])
print(f"Самый частый тип: {most_used[0]} ({most_used[1]} таймеров)")
```

### Диагностика проблем
```go
// Поиск аномалий в статистике
stats := getTimeWheelStats()
cancelRate := float64(stats.CancelledTimers) / float64(stats.TotalTimers)

if cancelRate > 0.3 {
    log.Println("Высокий уровень отмен таймеров!")
}
```

### Планирование ресурсов
```javascript
// Прогнозирование нагрузки
const trend = analytics.calculateTrend();
if (trend > 0.1) {
    console.log('📈 Рост нагрузки, планируем масштабирование');
}
```

## Рекомендации по мониторингу

### Критические пороги
- **Load Factor > 0.8**: Критическая загрузка
- **Cancel Rate > 25%**: Проблемы с отменами
- **Pending = 0**: Возможные проблемы с созданием

### Частота проверки
- **Производство**: каждые 30-60 секунд
- **Разработка**: каждые 5-10 секунд
- **Отладка**: каждые 1-2 секунды

## Связанные методы
- [ListTimers](list-timers.md) - Детальный анализ отдельных таймеров
- [AddTimer](add-timer.md) - Понимание нагрузки при создании
- [GetTimerStatus](get-timer-status.md) - Диагностика конкретных таймеров
