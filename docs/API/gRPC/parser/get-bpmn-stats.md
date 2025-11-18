# GetBPMNStats

## Описание
Получает статистику парсера BPMN, включая общую информацию о процессах, элементах и производительности системы.

## Синтаксис
```protobuf
rpc GetBPMNStats(GetBPMNStatsRequest) returns (GetBPMNStatsResponse);
```

## Package
```protobuf
package parser;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `parser`, `read` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### GetBPMNStatsRequest
```protobuf
message GetBPMNStatsRequest {
  // Пустое сообщение - статистика не требует параметров
}
```

## Параметры ответа

### GetBPMNStatsResponse
```protobuf
message GetBPMNStatsResponse {
  bool success = 1;                    // Статус успешности
  string message = 2;                  // Сообщение о результате
  BPMNStats stats = 3;                 // Статистика парсера
}

message BPMNStats {
  int32 total_processes = 1;           // Общее количество процессов
  int32 active_processes = 2;          // Количество активных процессов
  int32 total_elements = 3;            // Общее количество элементов
  map<string, int32> element_types = 4; // Статистика по типам элементов
  int64 total_file_size = 5;           // Общий размер BPMN файлов (байты)
  string last_parsed_at = 6;           // Время последнего парсинга
  int32 parse_errors = 7;              // Количество ошибок парсинга
  repeated ProcessStats top_processes = 8; // Топ процессов по размеру
}

message ProcessStats {
  string process_id = 1;               // ID процесса
  string process_key = 2;              // Ключ процесса
  int32 elements_count = 3;            // Количество элементов
  int64 file_size = 4;                 // Размер файла
  string status = 5;                   // Статус процесса
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
    
    pb "atom-engine/proto/parser/parserpb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewParserServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Получение статистики
    response, err := client.GetBPMNStats(ctx, &pb.GetBPMNStatsRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        stats := response.Stats
        
        fmt.Println("=== BPMN Parser Statistics ===")
        fmt.Printf("Всего процессов: %d\n", stats.TotalProcesses)
        fmt.Printf("Активных процессов: %d\n", stats.ActiveProcesses)
        fmt.Printf("Всего элементов: %d\n", stats.TotalElements)
        fmt.Printf("Общий размер файлов: %s\n", formatBytes(stats.TotalFileSize))
        fmt.Printf("Последний парсинг: %s\n", stats.LastParsedAt)
        fmt.Printf("Ошибок парсинга: %d\n", stats.ParseErrors)
        
        // Статистика по типам элементов
        if len(stats.ElementTypes) > 0 {
            fmt.Println("\nСтатистика по типам элементов:")
            for elementType, count := range stats.ElementTypes {
                fmt.Printf("  %-20s: %d\n", elementType, count)
            }
        }
        
        // Топ процессов
        if len(stats.TopProcesses) > 0 {
            fmt.Println("\nТоп процессов по размеру:")
            for i, process := range stats.TopProcesses {
                fmt.Printf("%d. %s (%d элементов, %s)\n", 
                    i+1, process.ProcessId, process.ElementsCount, 
                    formatBytes(process.FileSize))
            }
        }
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Мониторинг статистики с интервалом
func monitorStats(client pb.ParserServiceClient, ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    fmt.Println("Начинаем мониторинг статистики BPMN парсера...")
    
    for {
        select {
        case <-ticker.C:
            response, err := client.GetBPMNStats(ctx, &pb.GetBPMNStatsRequest{})
            if err != nil {
                log.Printf("Ошибка получения статистики: %v", err)
                continue
            }
            
            if response.Success {
                stats := response.Stats
                fmt.Printf("[%s] Процессов: %d, Элементов: %d, Ошибок: %d\n",
                    time.Now().Format("15:04:05"), 
                    stats.TotalProcesses, stats.TotalElements, stats.ParseErrors)
            }
            
        case <-ctx.Done():
            fmt.Println("Мониторинг остановлен")
            return
        }
    }
}
```

### Python
```python
import grpc
import parser_pb2
import parser_pb2_grpc
from datetime import datetime
import time

def get_bpmn_stats():
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    try:
        request = parser_pb2.GetBPMNStatsRequest()
        response = stub.GetBPMNStats(request, metadata=metadata)
        
        if response.success:
            stats = response.stats
            
            print("=== BPMN Parser Statistics ===")
            print(f"Всего процессов: {stats.total_processes}")
            print(f"Активных процессов: {stats.active_processes}")
            print(f"Всего элементов: {stats.total_elements}")
            print(f"Общий размер файлов: {format_bytes(stats.total_file_size)}")
            print(f"Последний парсинг: {stats.last_parsed_at}")
            print(f"Ошибок парсинга: {stats.parse_errors}")
            
            # Статистика по типам элементов
            if stats.element_types:
                print("\nСтатистика по типам элементов:")
                # Сортируем по количеству
                sorted_types = sorted(stats.element_types.items(), 
                                    key=lambda x: x[1], reverse=True)
                for element_type, count in sorted_types:
                    print(f"  {element_type:<20}: {count}")
            
            # Топ процессов
            if stats.top_processes:
                print("\nТоп процессов по размеру:")
                for i, process in enumerate(stats.top_processes, 1):
                    print(f"{i}. {process.process_id}")
                    print(f"   Элементов: {process.elements_count}")
                    print(f"   Размер: {format_bytes(process.file_size)}")
                    print(f"   Статус: {process.status}")
                    print()
            
            return stats
            
        else:
            print(f"Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def format_bytes(bytes_value):
    """Форматирование размера в читаемый вид"""
    for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
        if bytes_value < 1024.0:
            return f"{bytes_value:.1f} {unit}"
        bytes_value /= 1024.0
    return f"{bytes_value:.1f} PB"

def analyze_parser_health():
    """Анализ здоровья парсера"""
    stats = get_bpmn_stats()
    if not stats:
        return
    
    print("\n=== Анализ здоровья парсера ===")
    
    # Проверка на ошибки
    if stats.parse_errors > 0:
        error_rate = (stats.parse_errors / stats.total_processes) * 100
        print(f"⚠️  Найдены ошибки парсинга: {stats.parse_errors}")
        print(f"   Процент ошибок: {error_rate:.1f}%")
        
        if error_rate > 10:
            print("   🔴 Высокий уровень ошибок! Требуется внимание.")
        elif error_rate > 5:
            print("   🟡 Умеренный уровень ошибок.")
        else:
            print("   🟢 Приемлемый уровень ошибок.")
    else:
        print("✅ Ошибок парсинга не обнаружено")
    
    # Проверка активности
    if stats.total_processes > 0:
        active_rate = (stats.active_processes / stats.total_processes) * 100
        print(f"📊 Процент активных процессов: {active_rate:.1f}%")
        
        if active_rate < 50:
            print("   💡 Много неактивных процессов - возможно, стоит провести очистку")
    
    # Проверка размера
    avg_file_size = stats.total_file_size / stats.total_processes if stats.total_processes > 0 else 0
    print(f"📏 Средний размер файла: {format_bytes(avg_file_size)}")
    
    if avg_file_size > 1024 * 1024:  # 1MB
        print("   ⚠️  Большие BPMN файлы могут замедлять парсинг")

def compare_stats_over_time():
    """Сравнение статистики с течением времени"""
    print("Сбор статистики для сравнения (интервал 10 секунд)...")
    
    # Первый снимок
    stats1 = get_bpmn_stats()
    if not stats1:
        return
    
    time.sleep(10)
    
    # Второй снимок
    stats2 = get_bpmn_stats()
    if not stats2:
        return
    
    print("\n=== Изменения за 10 секунд ===")
    
    # Сравнение
    delta_processes = stats2.total_processes - stats1.total_processes
    delta_elements = stats2.total_elements - stats1.total_elements
    delta_errors = stats2.parse_errors - stats1.parse_errors
    
    if delta_processes != 0:
        print(f"Процессов: {stats1.total_processes} → {stats2.total_processes} ({delta_processes:+d})")
    
    if delta_elements != 0:
        print(f"Элементов: {stats1.total_elements} → {stats2.total_elements} ({delta_elements:+d})")
    
    if delta_errors != 0:
        print(f"Ошибок: {stats1.parse_errors} → {stats2.parse_errors} ({delta_errors:+d})")
    
    if delta_processes == 0 and delta_elements == 0 and delta_errors == 0:
        print("Изменений не обнаружено")

def export_stats_to_json():
    """Экспорт статистики в JSON"""
    import json
    
    stats = get_bpmn_stats()
    if not stats:
        return
    
    # Преобразование в словарь
    stats_dict = {
        'timestamp': datetime.now().isoformat(),
        'total_processes': stats.total_processes,
        'active_processes': stats.active_processes,
        'total_elements': stats.total_elements,
        'total_file_size': stats.total_file_size,
        'last_parsed_at': stats.last_parsed_at,
        'parse_errors': stats.parse_errors,
        'element_types': dict(stats.element_types),
        'top_processes': [
            {
                'process_id': p.process_id,
                'process_key': p.process_key,
                'elements_count': p.elements_count,
                'file_size': p.file_size,
                'status': p.status
            }
            for p in stats.top_processes
        ]
    }
    
    filename = f"bpmn_stats_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    
    try:
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(stats_dict, f, indent=2, ensure_ascii=False)
        print(f"Статистика экспортирована в: {filename}")
    except Exception as e:
        print(f"Ошибка экспорта: {e}")

if __name__ == "__main__":
    # Базовая статистика
    get_bpmn_stats()
    
    # Анализ здоровья
    analyze_parser_health()
    
    # Экспорт в JSON
    # export_stats_to_json()
    
    # Сравнение во времени
    # compare_stats_over_time()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const fs = require('fs').promises;

const PROTO_PATH = 'parser.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

async function getBPMNStats() {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.getBPMNStats(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            resolve(response.stats);
        });
    });
}

function formatBytes(bytes) {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let value = bytes;
    let unitIndex = 0;
    
    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex++;
    }
    
    return `${value.toFixed(1)} ${units[unitIndex]}`;
}

async function displayStats() {
    try {
        const stats = await getBPMNStats();
        
        console.log('=== BPMN Parser Statistics ===');
        console.log(`Всего процессов: ${stats.total_processes}`);
        console.log(`Активных процессов: ${stats.active_processes}`);
        console.log(`Всего элементов: ${stats.total_elements}`);
        console.log(`Общий размер файлов: ${formatBytes(stats.total_file_size)}`);
        console.log(`Последний парсинг: ${stats.last_parsed_at}`);
        console.log(`Ошибок парсинга: ${stats.parse_errors}`);
        
        // Статистика по типам элементов
        if (stats.element_types && Object.keys(stats.element_types).length > 0) {
            console.log('\nСтатистика по типам элементов:');
            
            // Сортируем по количеству
            const sortedTypes = Object.entries(stats.element_types)
                .sort(([,a], [,b]) => b - a);
            
            sortedTypes.forEach(([type, count]) => {
                console.log(`  ${type.padEnd(20)}: ${count}`);
            });
        }
        
        // Топ процессов
        if (stats.top_processes && stats.top_processes.length > 0) {
            console.log('\nТоп процессов по размеру:');
            stats.top_processes.forEach((process, index) => {
                console.log(`${index + 1}. ${process.process_id}`);
                console.log(`   Элементов: ${process.elements_count}`);
                console.log(`   Размер: ${formatBytes(process.file_size)}`);
                console.log(`   Статус: ${process.status}`);
                console.log();
            });
        }
        
        return stats;
        
    } catch (error) {
        console.error('Ошибка получения статистики:', error.message);
        return null;
    }
}

async function monitorStats(intervalSeconds = 30) {
    console.log(`Мониторинг статистики каждые ${intervalSeconds} секунд...`);
    console.log('Нажмите Ctrl+C для остановки\n');
    
    let previousStats = null;
    
    const monitor = async () => {
        try {
            const stats = await getBPMNStats();
            const timestamp = new Date().toLocaleTimeString();
            
            console.log(`[${timestamp}] Процессов: ${stats.total_processes}, Элементов: ${stats.total_elements}, Ошибок: ${stats.parse_errors}`);
            
            // Показываем изменения
            if (previousStats) {
                const deltaProcesses = stats.total_processes - previousStats.total_processes;
                const deltaElements = stats.total_elements - previousStats.total_elements;
                const deltaErrors = stats.parse_errors - previousStats.parse_errors;
                
                if (deltaProcesses !== 0 || deltaElements !== 0 || deltaErrors !== 0) {
                    const changes = [];
                    if (deltaProcesses !== 0) changes.push(`процессов: ${deltaProcesses > 0 ? '+' : ''}${deltaProcesses}`);
                    if (deltaElements !== 0) changes.push(`элементов: ${deltaElements > 0 ? '+' : ''}${deltaElements}`);
                    if (deltaErrors !== 0) changes.push(`ошибок: ${deltaErrors > 0 ? '+' : ''}${deltaErrors}`);
                    
                    console.log(`   Изменения: ${changes.join(', ')}`);
                }
            }
            
            previousStats = stats;
            
        } catch (error) {
            console.error(`[${new Date().toLocaleTimeString()}] Ошибка: ${error.message}`);
        }
    };
    
    // Первый запуск
    await monitor();
    
    // Периодический запуск
    const interval = setInterval(monitor, intervalSeconds * 1000);
    
    // Обработка сигнала завершения
    process.on('SIGINT', () => {
        clearInterval(interval);
        console.log('\nМониторинг остановлен');
        process.exit(0);
    });
}

async function generateReport() {
    try {
        const stats = await getBPMNStats();
        
        const report = {
            generated_at: new Date().toISOString(),
            summary: {
                total_processes: stats.total_processes,
                active_processes: stats.active_processes,
                total_elements: stats.total_elements,
                total_file_size: stats.total_file_size,
                parse_errors: stats.parse_errors
            },
            health_indicators: {
                error_rate: stats.total_processes > 0 ? 
                    (stats.parse_errors / stats.total_processes * 100).toFixed(2) + '%' : '0%',
                active_rate: stats.total_processes > 0 ? 
                    (stats.active_processes / stats.total_processes * 100).toFixed(2) + '%' : '0%',
                avg_file_size: stats.total_processes > 0 ? 
                    Math.round(stats.total_file_size / stats.total_processes) : 0
            },
            element_distribution: stats.element_types || {},
            top_processes: stats.top_processes || []
        };
        
        const filename = `bpmn_report_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.json`;
        
        await fs.writeFile(filename, JSON.stringify(report, null, 2));
        console.log(`Отчет сохранен: ${filename}`);
        
        return report;
        
    } catch (error) {
        console.error('Ошибка генерации отчета:', error.message);
        return null;
    }
}

// Примеры использования
if (require.main === module) {
    const command = process.argv[2];
    
    switch (command) {
        case 'monitor':
            const interval = parseInt(process.argv[3]) || 30;
            monitorStats(interval);
            break;
            
        case 'report':
            generateReport();
            break;
            
        default:
            displayStats();
            break;
    }
}

module.exports = {
    getBPMNStats,
    displayStats,
    monitorStats,
    generateReport
};
```

## Интерпретация статистики

### Ключевые метрики
- **total_processes**: Общее количество загруженных процессов
- **active_processes**: Процессы готовые к выполнению
- **total_elements**: Сумма всех BPMN элементов
- **element_types**: Распределение элементов по типам
- **parse_errors**: Количество ошибок при парсинге

### Показатели здоровья системы
- **Процент ошибок** < 5% - нормально
- **Процент активных процессов** > 80% - хорошо
- **Средний размер файла** < 100KB - оптимально

## Возможные ошибки

### gRPC Status Codes
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `INTERNAL` (13): Ошибка сбора статистики

## Связанные методы
- [ListBPMNProcesses](list-bpmn-processes.md) - Детальный список процессов
- [ParseBPMNFile](parse-bpmn-file.md) - Загрузка новых процессов
- [DeleteBPMNProcess](delete-bpmn-process.md) - Очистка неиспользуемых процессов

