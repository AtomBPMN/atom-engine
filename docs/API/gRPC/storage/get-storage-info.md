# GetStorageInfo

## Описание
Получает подробную информацию о хранилище данных включая размер, статистику использования, путь к базе данных и дополнительные метрики производительности.

## Синтаксис
```protobuf
rpc GetStorageInfo(GetStorageInfoRequest) returns (GetStorageInfoResponse);
```

## Package
```protobuf
package atom.storage.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `storage` или `*`

## Параметры запроса

### GetStorageInfoRequest
```protobuf
message GetStorageInfoRequest {}
```

## Параметры ответа

### GetStorageInfoResponse
```protobuf
message GetStorageInfoResponse {
  int64 total_size_bytes = 1;      // Общий размер базы данных в байтах
  int64 used_size_bytes = 2;       // Использованное место в байтах
  int64 free_size_bytes = 3;       // Свободное место в байтах
  int64 total_keys = 4;            // Общее количество ключей
  string database_path = 5;        // Путь к файлу базы данных
  map<string, string> statistics = 6; // Дополнительная статистика
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
    
    pb "atom-engine/proto/storage/storagepb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewStorageServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Получаем информацию о хранилище
    response, err := client.GetStorageInfo(ctx, &pb.GetStorageInfoRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("💾 ИНФОРМАЦИЯ О ХРАНИЛИЩЕ")
    fmt.Println("=" * 40)
    
    // Размеры
    fmt.Printf("📊 РАЗМЕР ДАННЫХ:\n")
    fmt.Printf("   Общий размер: %s\n", formatBytes(response.TotalSizeBytes))
    fmt.Printf("   Использовано: %s\n", formatBytes(response.UsedSizeBytes))
    fmt.Printf("   Свободно: %s\n", formatBytes(response.FreeSizeBytes))
    
    // Процент использования
    var usagePercent float64
    if response.TotalSizeBytes > 0 {
        usagePercent = float64(response.UsedSizeBytes) / float64(response.TotalSizeBytes) * 100
    }
    fmt.Printf("   Использование: %.1f%%\n", usagePercent)
    
    // Количество ключей
    fmt.Printf("\n🗃️ ДАННЫЕ:\n")
    fmt.Printf("   Общее количество ключей: %s\n", formatNumber(response.TotalKeys))
    
    // Путь к базе
    fmt.Printf("\n📁 РАСПОЛОЖЕНИЕ:\n")
    fmt.Printf("   Путь к базе: %s\n", response.DatabasePath)
    
    // Дополнительная статистика
    if len(response.Statistics) > 0 {
        fmt.Printf("\n📈 ДОПОЛНИТЕЛЬНАЯ СТАТИСТИКА:\n")
        for key, value := range response.Statistics {
            fmt.Printf("   %s: %s\n", key, value)
        }
    }
    
    // Анализ использования
    analyzeStorageUsage(response)
}

func formatBytes(bytes int64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
        TB = GB * 1024
    )
    
    switch {
    case bytes >= TB:
        return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
    default:
        return fmt.Sprintf("%d байт", bytes)
    }
}

func formatNumber(num int64) string {
    if num >= 1_000_000_000 {
        return fmt.Sprintf("%.1fB", float64(num)/1_000_000_000)
    } else if num >= 1_000_000 {
        return fmt.Sprintf("%.1fM", float64(num)/1_000_000)
    } else if num >= 1_000 {
        return fmt.Sprintf("%.1fK", float64(num)/1_000)
    }
    return fmt.Sprintf("%d", num)
}

func analyzeStorageUsage(info *pb.GetStorageInfoResponse) {
    fmt.Printf("\n🔍 АНАЛИЗ ИСПОЛЬЗОВАНИЯ:\n")
    
    // Анализ заполненности
    if info.TotalSizeBytes > 0 {
        usagePercent := float64(info.UsedSizeBytes) / float64(info.TotalSizeBytes) * 100
        
        if usagePercent > 90 {
            fmt.Printf("   🔴 КРИТИЧНО: Использование %.1f%% - требуется очистка\n", usagePercent)
        } else if usagePercent > 75 {
            fmt.Printf("   🟡 ВНИМАНИЕ: Использование %.1f%% - планируйте очистку\n", usagePercent)
        } else {
            fmt.Printf("   🟢 НОРМА: Использование %.1f%% - достаточно места\n", usagePercent)
        }
    }
    
    // Анализ количества ключей
    if info.TotalKeys > 1_000_000 {
        fmt.Printf("   📊 Большое количество ключей (%s) - рассмотрите архивирование\n", 
            formatNumber(info.TotalKeys))
    } else if info.TotalKeys > 100_000 {
        fmt.Printf("   📊 Среднее количество ключей (%s) - мониторьте рост\n", 
            formatNumber(info.TotalKeys))
    } else {
        fmt.Printf("   📊 Нормальное количество ключей (%s)\n", 
            formatNumber(info.TotalKeys))
    }
    
    // Средний размер записи
    if info.TotalKeys > 0 && info.UsedSizeBytes > 0 {
        avgRecordSize := info.UsedSizeBytes / info.TotalKeys
        fmt.Printf("   📏 Средний размер записи: %s\n", formatBytes(avgRecordSize))
        
        if avgRecordSize > 10*1024 { // > 10KB
            fmt.Printf("   ⚠️  Записи довольно большие - оптимизируйте структуру данных\n")
        }
    }
}

// Мониторинг роста базы данных
func monitorGrowth(client pb.StorageServiceClient, ctx context.Context, intervalSeconds int) {
    fmt.Printf("📈 Мониторинг роста базы данных каждые %d секунд\n", intervalSeconds)
    fmt.Printf("%-12s | %-10s | %-8s | %-12s | %s\n", 
        "Время", "Размер", "Ключи", "Использование", "Изменения")
    fmt.Printf("%s\n", strings.Repeat("-", 65))
    
    var prevSize, prevKeys int64
    ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            response, err := client.GetStorageInfo(ctx, &pb.GetStorageInfoRequest{})
            if err != nil {
                fmt.Printf("%-12s | ❌ ОШИБКА: %v\n", time.Now().Format("15:04:05"), err)
                continue
            }
            
            var usagePercent float64
            if response.TotalSizeBytes > 0 {
                usagePercent = float64(response.UsedSizeBytes) / float64(response.TotalSizeBytes) * 100
            }
            
            // Изменения с предыдущей проверки
            sizeChange := response.UsedSizeBytes - prevSize
            keysChange := response.TotalKeys - prevKeys
            
            changeStr := ""
            if prevSize > 0 {
                if sizeChange > 0 {
                    changeStr = fmt.Sprintf("+%s", formatBytes(sizeChange))
                } else if sizeChange < 0 {
                    changeStr = fmt.Sprintf("-%s", formatBytes(-sizeChange))
                } else {
                    changeStr = "нет изменений"
                }
                
                if keysChange != 0 {
                    changeStr += fmt.Sprintf(" (%+d ключей)", keysChange)
                }
            }
            
            fmt.Printf("%-12s | %-10s | %-8s | %-12.1f%% | %s\n",
                time.Now().Format("15:04:05"),
                formatBytes(response.UsedSizeBytes),
                formatNumber(response.TotalKeys),
                usagePercent,
                changeStr)
            
            prevSize = response.UsedSizeBytes
            prevKeys = response.TotalKeys
        }
    }
}

// Проверка места на диске
func checkDiskSpace(client pb.StorageServiceClient, ctx context.Context) {
    response, err := client.GetStorageInfo(ctx, &pb.GetStorageInfoRequest{})
    if err != nil {
        fmt.Printf("❌ Ошибка получения информации: %v\n", err)
        return
    }
    
    fmt.Printf("💽 ПРОВЕРКА МЕСТА НА ДИСКЕ\n")
    fmt.Printf("=" * 35)
    fmt.Printf("\n📁 База данных: %s\n", response.DatabasePath)
    
    // Проверяем свободное место
    if response.FreeSizeBytes < 100*1024*1024 { // < 100MB
        fmt.Printf("🔴 КРИТИЧНО: Мало свободного места (%s)\n", formatBytes(response.FreeSizeBytes))
        fmt.Printf("💡 Рекомендации:\n")
        fmt.Printf("   • Освободите место на диске\n")
        fmt.Printf("   • Выполните очистку старых данных\n")
        fmt.Printf("   • Рассмотрите перенос на больший диск\n")
    } else if response.FreeSizeBytes < 1024*1024*1024 { // < 1GB
        fmt.Printf("🟡 ВНИМАНИЕ: Ограниченное свободное место (%s)\n", formatBytes(response.FreeSizeBytes))
        fmt.Printf("💡 Планируйте очистку или расширение\n")
    } else {
        fmt.Printf("✅ Достаточно свободного места (%s)\n", formatBytes(response.FreeSizeBytes))
    }
}
```

### Python
```python
import grpc
import time
from datetime import datetime
import os

import storage_pb2
import storage_pb2_grpc

def get_storage_info():
    channel = grpc.insecure_channel('localhost:27500')
    stub = storage_pb2_grpc.StorageServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = storage_pb2.GetStorageInfoRequest()
    
    try:
        response = stub.GetStorageInfo(request, metadata=metadata)
        
        print("💾 ИНФОРМАЦИЯ О ХРАНИЛИЩЕ")
        print("=" * 40)
        
        # Размеры
        print("📊 РАЗМЕР ДАННЫХ:")
        print(f"   Общий размер: {format_bytes(response.total_size_bytes)}")
        print(f"   Использовано: {format_bytes(response.used_size_bytes)}")
        print(f"   Свободно: {format_bytes(response.free_size_bytes)}")
        
        # Процент использования
        usage_percent = 0
        if response.total_size_bytes > 0:
            usage_percent = (response.used_size_bytes / response.total_size_bytes) * 100
        print(f"   Использование: {usage_percent:.1f}%")
        
        # Количество ключей
        print(f"\n🗃️ ДАННЫЕ:")
        print(f"   Общее количество ключей: {format_number(response.total_keys)}")
        
        # Путь к базе
        print(f"\n📁 РАСПОЛОЖЕНИЕ:")
        print(f"   Путь к базе: {response.database_path}")
        
        # Дополнительная статистика
        if response.statistics:
            print(f"\n📈 ДОПОЛНИТЕЛЬНАЯ СТАТИСТИКА:")
            for key, value in response.statistics.items():
                print(f"   {key}: {value}")
        
        # Анализ использования
        analyze_storage_usage(response)
        
        return {
            'total_size_bytes': response.total_size_bytes,
            'used_size_bytes': response.used_size_bytes,
            'free_size_bytes': response.free_size_bytes,
            'total_keys': response.total_keys,
            'database_path': response.database_path,
            'statistics': dict(response.statistics)
        }
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def format_bytes(bytes_count):
    """Форматирует байты в читаемый вид"""
    for unit in ['байт', 'KB', 'MB', 'GB', 'TB']:
        if bytes_count < 1024:
            return f"{bytes_count:.2f} {unit}"
        bytes_count /= 1024
    return f"{bytes_count:.2f} PB"

def format_number(num):
    """Форматирует большие числа"""
    if num >= 1_000_000_000:
        return f"{num/1_000_000_000:.1f}B"
    elif num >= 1_000_000:
        return f"{num/1_000_000:.1f}M"
    elif num >= 1_000:
        return f"{num/1_000:.1f}K"
    return str(num)

def analyze_storage_usage(info):
    """Анализирует использование хранилища"""
    print(f"\n🔍 АНАЛИЗ ИСПОЛЬЗОВАНИЯ:")
    
    # Анализ заполненности
    if info.total_size_bytes > 0:
        usage_percent = (info.used_size_bytes / info.total_size_bytes) * 100
        
        if usage_percent > 90:
            print(f"   🔴 КРИТИЧНО: Использование {usage_percent:.1f}% - требуется очистка")
        elif usage_percent > 75:
            print(f"   🟡 ВНИМАНИЕ: Использование {usage_percent:.1f}% - планируйте очистку")
        else:
            print(f"   🟢 НОРМА: Использование {usage_percent:.1f}% - достаточно места")
    
    # Анализ количества ключей
    if info.total_keys > 1_000_000:
        print(f"   📊 Большое количество ключей ({format_number(info.total_keys)}) - рассмотрите архивирование")
    elif info.total_keys > 100_000:
        print(f"   📊 Среднее количество ключей ({format_number(info.total_keys)}) - мониторьте рост")
    else:
        print(f"   📊 Нормальное количество ключей ({format_number(info.total_keys)})")
    
    # Средний размер записи
    if info.total_keys > 0 and info.used_size_bytes > 0:
        avg_record_size = info.used_size_bytes / info.total_keys
        print(f"   📏 Средний размер записи: {format_bytes(avg_record_size)}")
        
        if avg_record_size > 10 * 1024:  # > 10KB
            print("   ⚠️  Записи довольно большие - оптимизируйте структуру данных")

# Класс для мониторинга хранилища
class StorageMonitor:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = storage_pb2_grpc.StorageServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
        self.history = []
    
    def get_info(self):
        """Получает информацию о хранилище"""
        try:
            request = storage_pb2.GetStorageInfoRequest()
            response = self.stub.GetStorageInfo(request, metadata=self.metadata)
            
            return {
                'timestamp': time.time(),
                'total_size_bytes': response.total_size_bytes,
                'used_size_bytes': response.used_size_bytes,
                'free_size_bytes': response.free_size_bytes,
                'total_keys': response.total_keys,
                'database_path': response.database_path,
                'statistics': dict(response.statistics)
            }
        except grpc.RpcError as e:
            return {'error': f"{e.code()} - {e.details()}", 'timestamp': time.time()}
    
    def monitor_growth(self, interval_seconds=60, duration_minutes=None):
        """Мониторинг роста базы данных"""
        print(f"📈 Мониторинг роста базы данных каждые {interval_seconds} секунд")
        print(f"{'Время':<12} | {'Размер':<10} | {'Ключи':<8} | {'Использование':<12} | Изменения")
        print("-" * 70)
        
        prev_info = None
        start_time = time.time()
        
        try:
            while True:
                if duration_minutes and (time.time() - start_time) > duration_minutes * 60:
                    break
                
                info = self.get_info()
                current_time = datetime.now().strftime('%H:%M:%S')
                
                if 'error' in info:
                    print(f"{current_time:<12} | ❌ ОШИБКА: {info['error']}")
                    time.sleep(interval_seconds)
                    continue
                
                # Вычисляем использование
                usage_percent = 0
                if info['total_size_bytes'] > 0:
                    usage_percent = (info['used_size_bytes'] / info['total_size_bytes']) * 100
                
                # Изменения с предыдущей проверки
                change_str = ""
                if prev_info:
                    size_change = info['used_size_bytes'] - prev_info['used_size_bytes']
                    keys_change = info['total_keys'] - prev_info['total_keys']
                    
                    if size_change > 0:
                        change_str = f"+{format_bytes(size_change)}"
                    elif size_change < 0:
                        change_str = f"-{format_bytes(-size_change)}"
                    else:
                        change_str = "нет изменений"
                    
                    if keys_change != 0:
                        change_str += f" ({keys_change:+d} ключей)"
                
                print(f"{current_time:<12} | {format_bytes(info['used_size_bytes']):<10} | "
                      f"{format_number(info['total_keys']):<8} | {usage_percent:<12.1f}% | {change_str}")
                
                # Сохраняем в историю
                self.history.append(info)
                if len(self.history) > 1000:  # Ограничиваем размер истории
                    self.history = self.history[-1000:]
                
                prev_info = info
                time.sleep(interval_seconds)
                
        except KeyboardInterrupt:
            print("\n🛑 Мониторинг остановлен")
    
    def generate_report(self):
        """Генерирует отчет на основе истории"""
        if len(self.history) < 2:
            print("📭 Недостаточно данных для отчета")
            return
        
        print("📋 ОТЧЕТ ПО ХРАНИЛИЩУ")
        print("=" * 30)
        
        first = self.history[0]
        last = self.history[-1]
        duration_hours = (last['timestamp'] - first['timestamp']) / 3600
        
        print(f"📊 Период анализа: {duration_hours:.1f} часов")
        print(f"📈 Точек данных: {len(self.history)}")
        
        # Изменения за период
        size_change = last['used_size_bytes'] - first['used_size_bytes']
        keys_change = last['total_keys'] - first['total_keys']
        
        print(f"\n📊 ИЗМЕНЕНИЯ ЗА ПЕРИОД:")
        print(f"   Размер: {format_bytes(first['used_size_bytes'])} → {format_bytes(last['used_size_bytes'])}")
        if size_change > 0:
            print(f"   Рост: +{format_bytes(size_change)}")
        elif size_change < 0:
            print(f"   Уменьшение: {format_bytes(-size_change)}")
        
        print(f"   Ключи: {format_number(first['total_keys'])} → {format_number(last['total_keys'])}")
        if keys_change != 0:
            print(f"   Изменение: {keys_change:+d}")
        
        # Скорость роста
        if duration_hours > 0 and size_change > 0:
            growth_rate_per_hour = size_change / duration_hours
            print(f"\n📈 СКОРОСТЬ РОСТА:")
            print(f"   {format_bytes(growth_rate_per_hour)}/час")
            
            # Прогноз
            if growth_rate_per_hour > 0:
                remaining_space = last['free_size_bytes']
                hours_until_full = remaining_space / growth_rate_per_hour
                if hours_until_full < 24 * 7:  # Меньше недели
                    print(f"   ⚠️  При текущей скорости место закончится через {hours_until_full:.1f} часов")
    
    def check_disk_space(self):
        """Проверяет место на диске"""
        info = self.get_info()
        
        if 'error' in info:
            print(f"❌ Ошибка получения информации: {info['error']}")
            return
        
        print("💽 ПРОВЕРКА МЕСТА НА ДИСКЕ")
        print("=" * 35)
        print(f"📁 База данных: {info['database_path']}")
        
        # Проверяем свободное место
        free_space = info['free_size_bytes']
        
        if free_space < 100 * 1024 * 1024:  # < 100MB
            print(f"🔴 КРИТИЧНО: Мало свободного места ({format_bytes(free_space)})")
            print("💡 Рекомендации:")
            print("   • Освободите место на диске")
            print("   • Выполните очистку старых данных")  
            print("   • Рассмотрите перенос на больший диск")
        elif free_space < 1024 * 1024 * 1024:  # < 1GB
            print(f"🟡 ВНИМАНИЕ: Ограниченное свободное место ({format_bytes(free_space)})")
            print("💡 Планируйте очистку или расширение")
        else:
            print(f"✅ Достаточно свободного места ({format_bytes(free_space)})")
        
        # Дополнительная проверка файловой системы
        if os.path.exists(info['database_path']):
            dir_path = os.path.dirname(info['database_path'])
            try:
                stat_result = os.statvfs(dir_path)
                fs_free = stat_result.f_bavail * stat_result.f_frsize
                print(f"💿 Свободно на файловой системе: {format_bytes(fs_free)}")
                
                if fs_free != free_space:
                    print(f"ℹ️  Расхождение с данными БД: {format_bytes(abs(fs_free - free_space))}")
            except OSError:
                print("⚠️  Не удалось получить информацию о файловой системе")

# Утилиты для администрирования
def storage_health_check():
    """Комплексная проверка здоровья хранилища"""
    print("🏥 КОМПЛЕКСНАЯ ПРОВЕРКА ХРАНИЛИЩА")
    print("=" * 45)
    
    monitor = StorageMonitor()
    info = monitor.get_info()
    
    if 'error' in info:
        print(f"❌ Не удалось получить информацию: {info['error']}")
        return
    
    issues = []
    recommendations = []
    
    # Проверка места
    if info['total_size_bytes'] > 0:
        usage_percent = (info['used_size_bytes'] / info['total_size_bytes']) * 100
        if usage_percent > 90:
            issues.append("🔴 Критически мало места")
            recommendations.append("• Немедленно освободите место")
        elif usage_percent > 75:
            issues.append("🟡 Ограниченное место")
            recommendations.append("• Планируйте очистку данных")
    
    # Проверка количества ключей
    if info['total_keys'] > 10_000_000:
        issues.append("🟡 Очень много ключей")
        recommendations.append("• Рассмотрите архивирование старых данных")
    
    # Проверка среднего размера записи
    if info['total_keys'] > 0:
        avg_record_size = info['used_size_bytes'] / info['total_keys']
        if avg_record_size > 50 * 1024:  # > 50KB
            issues.append("🟡 Большие записи")
            recommendations.append("• Оптимизируйте структуру данных")
    
    # Результаты
    if issues:
        print("⚠️ ОБНАРУЖЕННЫЕ ПРОБЛЕМЫ:")
        for issue in issues:
            print(f"   {issue}")
    else:
        print("✅ ПРОБЛЕМ НЕ ОБНАРУЖЕНО")
    
    if recommendations:
        print(f"\n💡 РЕКОМЕНДАЦИИ:")
        for rec in recommendations:
            print(f"   {rec}")
    
    # Общая оценка
    health_score = max(0, 100 - len(issues) * 25)
    print(f"\n📊 ОБЩАЯ ОЦЕНКА: {health_score}%")
    
    if health_score >= 90:
        print("🟢 Отличное состояние")
    elif health_score >= 70:
        print("🟡 Хорошее состояние")
    elif health_score >= 50:
        print("🟠 Требует внимания")
    else:
        print("🔴 Критическое состояние")

if __name__ == "__main__":
    # Простая проверка информации
    get_storage_info()
    
    print("\n" + "="*60)
    
    # Комплексная проверка здоровья
    storage_health_check()
    
    print("\n" + "="*60)
    
    # Демонстрация мониторинга (кратковременного)
    print("\n📈 Демонстрация мониторинга (30 секунд):")
    monitor = StorageMonitor()
    monitor.monitor_growth(interval_seconds=5, duration_minutes=0.5)
    
    # Генерация отчета
    monitor.generate_report()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const fs = require('fs');
const path = require('path');

const PROTO_PATH = 'storage.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const storageProto = grpc.loadPackageDefinition(packageDefinition).atom.storage.v1;

async function getStorageInfo() {
    const client = new storageProto.StorageService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.getStorageInfo(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log('💾 ИНФОРМАЦИЯ О ХРАНИЛИЩЕ');
            console.log('='.repeat(40));
            
            // Размеры
            console.log('📊 РАЗМЕР ДАННЫХ:');
            console.log(`   Общий размер: ${formatBytes(response.total_size_bytes)}`);
            console.log(`   Использовано: ${formatBytes(response.used_size_bytes)}`);
            console.log(`   Свободно: ${formatBytes(response.free_size_bytes)}`);
            
            // Процент использования
            let usagePercent = 0;
            if (response.total_size_bytes > 0) {
                usagePercent = (response.used_size_bytes / response.total_size_bytes) * 100;
            }
            console.log(`   Использование: ${usagePercent.toFixed(1)}%`);
            
            // Количество ключей
            console.log('\n🗃️ ДАННЫЕ:');
            console.log(`   Общее количество ключей: ${formatNumber(response.total_keys)}`);
            
            // Путь к базе
            console.log('\n📁 РАСПОЛОЖЕНИЕ:');
            console.log(`   Путь к базе: ${response.database_path}`);
            
            // Дополнительная статистика
            if (Object.keys(response.statistics).length > 0) {
                console.log('\n📈 ДОПОЛНИТЕЛЬНАЯ СТАТИСТИКА:');
                Object.entries(response.statistics).forEach(([key, value]) => {
                    console.log(`   ${key}: ${value}`);
                });
            }
            
            // Анализ использования
            analyzeStorageUsage(response);
            
            resolve({
                totalSizeBytes: response.total_size_bytes,
                usedSizeBytes: response.used_size_bytes,
                freeSizeBytes: response.free_size_bytes,
                totalKeys: response.total_keys,
                databasePath: response.database_path,
                statistics: response.statistics
            });
        });
    });
}

function formatBytes(bytes) {
    const sizes = ['байт', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 байт';
    
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
}

function formatNumber(num) {
    if (num >= 1e9) return `${(num / 1e9).toFixed(1)}B`;
    if (num >= 1e6) return `${(num / 1e6).toFixed(1)}M`;
    if (num >= 1e3) return `${(num / 1e3).toFixed(1)}K`;
    return num.toString();
}

function analyzeStorageUsage(info) {
    console.log('\n🔍 АНАЛИЗ ИСПОЛЬЗОВАНИЯ:');
    
    // Анализ заполненности
    if (info.total_size_bytes > 0) {
        const usagePercent = (info.used_size_bytes / info.total_size_bytes) * 100;
        
        if (usagePercent > 90) {
            console.log(`   🔴 КРИТИЧНО: Использование ${usagePercent.toFixed(1)}% - требуется очистка`);
        } else if (usagePercent > 75) {
            console.log(`   🟡 ВНИМАНИЕ: Использование ${usagePercent.toFixed(1)}% - планируйте очистку`);
        } else {
            console.log(`   🟢 НОРМА: Использование ${usagePercent.toFixed(1)}% - достаточно места`);
        }
    }
    
    // Анализ количества ключей
    if (info.total_keys > 1000000) {
        console.log(`   📊 Большое количество ключей (${formatNumber(info.total_keys)}) - рассмотрите архивирование`);
    } else if (info.total_keys > 100000) {
        console.log(`   📊 Среднее количество ключей (${formatNumber(info.total_keys)}) - мониторьте рост`);
    } else {
        console.log(`   📊 Нормальное количество ключей (${formatNumber(info.total_keys)})`);
    }
    
    // Средний размер записи
    if (info.total_keys > 0 && info.used_size_bytes > 0) {
        const avgRecordSize = info.used_size_bytes / info.total_keys;
        console.log(`   📏 Средний размер записи: ${formatBytes(avgRecordSize)}`);
        
        if (avgRecordSize > 10 * 1024) { // > 10KB
            console.log('   ⚠️  Записи довольно большие - оптимизируйте структуру данных');
        }
    }
}

// Класс для продвинутого мониторинга хранилища
class StorageAnalytics {
    constructor() {
        this.client = new storageProto.StorageService('localhost:27500',
            grpc.credentials.createInsecure());
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
        this.history = [];
        this.isMonitoring = false;
        this.monitoringInterval = null;
    }
    
    async getInfo() {
        return new Promise((resolve, reject) => {
            this.client.getStorageInfo({}, this.metadata, (error, response) => {
                if (error) {
                    resolve({ error: error.message, timestamp: Date.now() });
                } else {
                    resolve({
                        totalSizeBytes: response.total_size_bytes,
                        usedSizeBytes: response.used_size_bytes,
                        freeSizeBytes: response.free_size_bytes,
                        totalKeys: response.total_keys,
                        databasePath: response.database_path,
                        statistics: response.statistics,
                        timestamp: Date.now()
                    });
                }
            });
        });
    }
    
    startGrowthMonitoring(intervalMs = 60000) {
        if (this.isMonitoring) {
            console.log('⚠️ Мониторинг уже запущен');
            return;
        }
        
        this.isMonitoring = true;
        console.log(`📈 Запуск мониторинга роста каждые ${intervalMs / 1000} секунд`);
        console.log('Время    | Размер    | Ключи   | Использование | Изменения');
        console.log('-'.repeat(65));
        
        let previousInfo = null;
        
        const monitor = async () => {
            if (!this.isMonitoring) return;
            
            const info = await this.getInfo();
            const currentTime = new Date().toLocaleTimeString();
            
            if (info.error) {
                console.log(`${currentTime} | ❌ ОШИБКА: ${info.error}`);
            } else {
                // Вычисляем использование
                const usagePercent = info.totalSizeBytes > 0 
                    ? (info.usedSizeBytes / info.totalSizeBytes) * 100 
                    : 0;
                
                // Изменения с предыдущей проверки
                let changeStr = '';
                if (previousInfo && !previousInfo.error) {
                    const sizeChange = info.usedSizeBytes - previousInfo.usedSizeBytes;
                    const keysChange = info.totalKeys - previousInfo.totalKeys;
                    
                    if (sizeChange > 0) {
                        changeStr = `+${formatBytes(sizeChange)}`;
                    } else if (sizeChange < 0) {
                        changeStr = `-${formatBytes(-sizeChange)}`;
                    } else {
                        changeStr = 'нет изменений';
                    }
                    
                    if (keysChange !== 0) {
                        changeStr += ` (${keysChange > 0 ? '+' : ''}${keysChange} ключей)`;
                    }
                }
                
                console.log(`${currentTime} | ${formatBytes(info.usedSizeBytes).padEnd(9)} | ` +
                           `${formatNumber(info.totalKeys).padEnd(7)} | ${usagePercent.toFixed(1)}%`.padEnd(13) +
                           ` | ${changeStr}`);
                
                // Сохраняем в историю
                this.history.push(info);
                if (this.history.length > 1000) {
                    this.history = this.history.slice(-1000);
                }
            }
            
            previousInfo = info;
        };
        
        // Первая проверка сразу
        monitor();
        
        // Запуск периодических проверок
        this.monitoringInterval = setInterval(monitor, intervalMs);
    }
    
    stopMonitoring() {
        if (!this.isMonitoring) return;
        
        this.isMonitoring = false;
        if (this.monitoringInterval) {
            clearInterval(this.monitoringInterval);
            this.monitoringInterval = null;
        }
        console.log('🛑 Мониторинг остановлен');
    }
    
    generateReport() {
        if (this.history.length < 2) {
            console.log('📭 Недостаточно данных для отчета');
            return;
        }
        
        console.log('\n📋 ОТЧЕТ ПО ХРАНИЛИЩУ');
        console.log('='.repeat(30));
        
        const first = this.history[0];
        const last = this.history[this.history.length - 1];
        const durationHours = (last.timestamp - first.timestamp) / (1000 * 60 * 60);
        
        console.log(`📊 Период анализа: ${durationHours.toFixed(1)} часов`);
        console.log(`📈 Точек данных: ${this.history.length}`);
        
        // Изменения за период
        const sizeChange = last.usedSizeBytes - first.usedSizeBytes;
        const keysChange = last.totalKeys - first.totalKeys;
        
        console.log('\n📊 ИЗМЕНЕНИЯ ЗА ПЕРИОД:');
        console.log(`   Размер: ${formatBytes(first.usedSizeBytes)} → ${formatBytes(last.usedSizeBytes)}`);
        if (sizeChange > 0) {
            console.log(`   Рост: +${formatBytes(sizeChange)}`);
        } else if (sizeChange < 0) {
            console.log(`   Уменьшение: ${formatBytes(-sizeChange)}`);
        }
        
        console.log(`   Ключи: ${formatNumber(first.totalKeys)} → ${formatNumber(last.totalKeys)}`);
        if (keysChange !== 0) {
            console.log(`   Изменение: ${keysChange > 0 ? '+' : ''}${keysChange}`);
        }
        
        // Скорость роста
        if (durationHours > 0 && sizeChange > 0) {
            const growthRatePerHour = sizeChange / durationHours;
            console.log('\n📈 СКОРОСТЬ РОСТА:');
            console.log(`   ${formatBytes(growthRatePerHour)}/час`);
            
            // Прогноз
            if (growthRatePerHour > 0) {
                const remainingSpace = last.freeSizeBytes;
                const hoursUntilFull = remainingSpace / growthRatePerHour;
                if (hoursUntilFull < 24 * 7) { // Меньше недели
                    console.log(`   ⚠️  При текущей скорости место закончится через ${hoursUntilFull.toFixed(1)} часов`);
                }
            }
        }
        
        // Статистика производительности
        const sizes = this.history.map(h => h.usedSizeBytes);
        const avgSize = sizes.reduce((a, b) => a + b) / sizes.length;
        const maxSize = Math.max(...sizes);
        const minSize = Math.min(...sizes);
        
        console.log('\n📊 СТАТИСТИКА РАЗМЕРА:');
        console.log(`   Средний: ${formatBytes(avgSize)}`);
        console.log(`   Максимум: ${formatBytes(maxSize)}`);
        console.log(`   Минимум: ${formatBytes(minSize)}`);
    }
    
    async checkDiskSpace() {
        console.log('💽 ПРОВЕРКА МЕСТА НА ДИСКЕ');
        console.log('='.repeat(35));
        
        const info = await this.getInfo();
        
        if (info.error) {
            console.log(`❌ Ошибка получения информации: ${info.error}`);
            return;
        }
        
        console.log(`📁 База данных: ${info.databasePath}`);
        
        // Проверяем свободное место
        const freeSpace = info.freeSizeBytes;
        
        if (freeSpace < 100 * 1024 * 1024) { // < 100MB
            console.log(`🔴 КРИТИЧНО: Мало свободного места (${formatBytes(freeSpace)})`);
            console.log('💡 Рекомендации:');
            console.log('   • Освободите место на диске');
            console.log('   • Выполните очистку старых данных');
            console.log('   • Рассмотрите перенос на больший диск');
        } else if (freeSpace < 1024 * 1024 * 1024) { // < 1GB
            console.log(`🟡 ВНИМАНИЕ: Ограниченное свободное место (${formatBytes(freeSpace)})`);
            console.log('💡 Планируйте очистку или расширение');
        } else {
            console.log(`✅ Достаточно свободного места (${formatBytes(freeSpace)})`);
        }
        
        // Дополнительная проверка файловой системы (если возможно)
        try {
            const dirPath = path.dirname(info.databasePath);
            const stats = fs.statSync(dirPath);
            console.log(`📊 Директория существует: ${stats.isDirectory() ? '✅' : '❌'}`);
        } catch (error) {
            console.log('⚠️  Не удалось проверить файловую систему');
        }
    }
    
    async performHealthCheck() {
        console.log('🏥 КОМПЛЕКСНАЯ ПРОВЕРКА ХРАНИЛИЩА');
        console.log('='.repeat(45));
        
        const info = await this.getInfo();
        
        if (info.error) {
            console.log(`❌ Не удалось получить информацию: ${info.error}`);
            return;
        }
        
        const issues = [];
        const recommendations = [];
        
        // Проверка места
        if (info.totalSizeBytes > 0) {
            const usagePercent = (info.usedSizeBytes / info.totalSizeBytes) * 100;
            if (usagePercent > 90) {
                issues.push('🔴 Критически мало места');
                recommendations.push('• Немедленно освободите место');
            } else if (usagePercent > 75) {
                issues.push('🟡 Ограниченное место');
                recommendations.push('• Планируйте очистку данных');
            }
        }
        
        // Проверка количества ключей
        if (info.totalKeys > 10000000) {
            issues.push('🟡 Очень много ключей');
            recommendations.push('• Рассмотрите архивирование старых данных');
        }
        
        // Проверка среднего размера записи
        if (info.totalKeys > 0) {
            const avgRecordSize = info.usedSizeBytes / info.totalKeys;
            if (avgRecordSize > 50 * 1024) { // > 50KB
                issues.push('🟡 Большие записи');
                recommendations.push('• Оптимизируйте структуру данных');
            }
        }
        
        // Результаты
        if (issues.length > 0) {
            console.log('⚠️ ОБНАРУЖЕННЫЕ ПРОБЛЕМЫ:');
            issues.forEach(issue => console.log(`   ${issue}`));
        } else {
            console.log('✅ ПРОБЛЕМ НЕ ОБНАРУЖЕНО');
        }
        
        if (recommendations.length > 0) {
            console.log('\n💡 РЕКОМЕНДАЦИИ:');
            recommendations.forEach(rec => console.log(`   ${rec}`));
        }
        
        // Общая оценка
        const healthScore = Math.max(0, 100 - issues.length * 25);
        console.log(`\n📊 ОБЩАЯ ОЦЕНКА: ${healthScore}%`);
        
        if (healthScore >= 90) {
            console.log('🟢 Отличное состояние');
        } else if (healthScore >= 70) {
            console.log('🟡 Хорошее состояние');
        } else if (healthScore >= 50) {
            console.log('🟠 Требует внимания');
        } else {
            console.log('🔴 Критическое состояние');
        }
        
        return healthScore;
    }
}

// Демонстрация всех возможностей
async function demonstrateStorageAnalytics() {
    console.log('🚀 Демонстрация аналитики хранилища\n');
    
    // Базовая информация
    try {
        await getStorageInfo();
    } catch (error) {
        console.log(`❌ Ошибка получения информации: ${error.message}`);
        return;
    }
    
    console.log('\n' + '='.repeat(60));
    
    // Создаем объект аналитики
    const analytics = new StorageAnalytics();
    
    // Проверка здоровья
    const healthScore = await analytics.performHealthCheck();
    
    console.log('\n' + '='.repeat(60));
    
    // Проверка места на диске
    await analytics.checkDiskSpace();
    
    console.log('\n' + '='.repeat(60));
    
    // Демонстрация мониторинга (кратковременного)
    console.log('\n📈 Демонстрация мониторинга роста (30 секунд):');
    analytics.startGrowthMonitoring(5000); // Каждые 5 секунд
    
    setTimeout(() => {
        analytics.stopMonitoring();
        
        // Генерация отчета
        console.log('\n📋 Генерация отчета:');
        analytics.generateReport();
        
        console.log('\n✅ Демонстрация завершена');
    }, 30000);
}

// Основная демонстрация
async function main() {
    try {
        await demonstrateStorageAnalytics();
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Информация о размере

### Основные поля
- **total_size_bytes**: Общий размер БД файла
- **used_size_bytes**: Фактически используемое место
- **free_size_bytes**: Свободное место в БД
- **total_keys**: Количество записей

### Расчетные показатели
- **Процент использования**: `used_size_bytes / total_size_bytes * 100%`
- **Средний размер записи**: `used_size_bytes / total_keys`
- **Коэффициент заполнения**: Эффективность использования места

## Статистические данные

### Поле statistics
```json
{
  "compactions": "15",
  "level0_files": "3",
  "level1_files": "12",
  "bloom_filter_memory": "1048576",
  "index_memory": "2097152",
  "read_operations": "152436",
  "write_operations": "89234"
}
```

## Применение

### Мониторинг производительности
```javascript
// Отслеживание роста БД
const info = await getStorageInfo();
if (info.usedSizeBytes > threshold) {
    sendAlert('Database size exceeded threshold');
}
```

### Планирование емкости
```python
# Прогнозирование необходимости очистки
growth_rate = calculate_growth_rate(history)
time_until_full = remaining_space / growth_rate
```

### Диагностика проблем
```go
// Выявление неэффективного использования места
avgRecordSize := info.UsedSizeBytes / info.TotalKeys
if avgRecordSize > 100*1024 { // > 100KB
    log.Println("Записи слишком большие")
}
```

### Автоматическая очистка
```javascript
// Триггер для архивирования
if (info.totalKeys > 10_000_000) {
    scheduleArchiving();
}
```

## Алерты и пороги

### Критические пороги
- **Использование > 90%**: Критически мало места
- **Средняя запись > 100KB**: Неэффективная структура
- **Ключей > 10M**: Требуется архивирование

### Мониторинг тенденций
- Скорость роста БД
- Изменение размера записей
- Частота операций чтения/записи

## Связанные методы
- [GetStorageStatus](get-storage-status.md) - Проверка подключения перед получением информации
