# ListProcessInstances

## Описание
Получает список экземпляров процессов с поддержкой фильтрации, сортировки и пагинации.

## Синтаксис
```protobuf
rpc ListProcessInstances(ListProcessInstancesRequest) returns (ListProcessInstancesResponse);
```

## Package
```protobuf
package atom.process.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`, `read` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### ListProcessInstancesRequest
```protobuf
message ListProcessInstancesRequest {
  string status_filter = 1;        // Фильтр по статусу (ACTIVE, COMPLETED, CANCELLED)
  int32 limit = 2;                 // Лимит записей (устарел, используйте page_size)
  string process_key_filter = 3;   // Фильтр по ключу процесса
  int32 page_size = 4;             // Размер страницы (по умолчанию: 20)
  int32 page = 5;                  // Номер страницы (начиная с 1)
  string sort_by = 6;              // Поле сортировки (по умолчанию: "started_at")
  string sort_order = 7;           // Порядок сортировки: "ASC" или "DESC" (по умолчанию: "DESC")
}
```

#### Поля:
- **status_filter** (string, optional): Фильтр по статусу (`ACTIVE`, `COMPLETED`, `CANCELLED`, `FAILED`)
- **limit** (int32, deprecated): Используйте `page_size`
- **process_key_filter** (string, optional): Фильтр по ключу/ID процесса
- **page_size** (int32, optional): Количество записей на странице (1-1000, по умолчанию: 20)
- **page** (int32, optional): Номер страницы (начиная с 1)
- **sort_by** (string, optional): Поле сортировки (`started_at`, `updated_at`, `status`, `process_key`)
- **sort_order** (string, optional): Порядок сортировки (`ASC`, `DESC`)

## Параметры ответа

### ListProcessInstancesResponse
```protobuf
message ListProcessInstancesResponse {
  repeated ProcessInstanceInfo instances = 1; // Список экземпляров процессов
  bool success = 2;                           // Статус успешности
  string message = 3;                         // Сообщение о результате
  int32 total_count = 4;                      // Общее количество записей
  int32 page = 5;                             // Текущая страница
  int32 page_size = 6;                        // Размер страницы
  int32 total_pages = 7;                      // Общее количество страниц
}

message ProcessInstanceInfo {
  string instance_id = 1;                     // ID экземпляра
  string process_key = 2;                     // Ключ процесса
  string status = 3;                          // Статус экземпляра
  string current_activity = 4;                // Текущая активность
  int64 started_at = 5;                       // Время запуска (Unix timestamp)
  int64 updated_at = 6;                       // Время обновления (Unix timestamp)
  map<string, string> variables = 7;          // Переменные процесса
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
    
    pb "atom-engine/proto/process/processpb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewProcessServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Простой запрос всех процессов
    response, err := client.ListProcessInstances(ctx, &pb.ListProcessInstancesRequest{
        PageSize:  20,
        Page:      1,
        SortBy:    "started_at",
        SortOrder: "DESC",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("=== Экземпляры процессов (страница %d/%d) ===\n", 
            response.Page, response.TotalPages)
        fmt.Printf("Всего найдено: %d\n\n", response.TotalCount)
        
        for i, instance := range response.Instances {
            fmt.Printf("%d. %s\n", i+1, instance.InstanceId)
            fmt.Printf("   Процесс: %s\n", instance.ProcessKey)
            fmt.Printf("   Статус: %s\n", instance.Status)
            fmt.Printf("   Текущая активность: %s\n", instance.CurrentActivity)
            fmt.Printf("   Запущен: %s\n", formatTimestamp(instance.StartedAt))
            fmt.Printf("   Обновлен: %s\n", formatTimestamp(instance.UpdatedAt))
            
            if len(instance.Variables) > 0 {
                fmt.Printf("   Переменные: %d\n", len(instance.Variables))
            }
            fmt.Println()
        }
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

func formatTimestamp(timestamp int64) string {
    return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// Получение активных процессов
func getActiveProcesses(client pb.ProcessServiceClient, ctx context.Context) ([]*pb.ProcessInstanceInfo, error) {
    response, err := client.ListProcessInstances(ctx, &pb.ListProcessInstancesRequest{
        StatusFilter: "ACTIVE",
        PageSize:     1000,
        SortBy:      "started_at",
        SortOrder:   "ASC", // Сначала старые
    })
    
    if err != nil {
        return nil, err
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка получения списка: %s", response.Message)
    }
    
    return response.Instances, nil
}

// Пагинация через все процессы
func getAllProcesses(client pb.ProcessServiceClient, ctx context.Context) ([]*pb.ProcessInstanceInfo, error) {
    var allInstances []*pb.ProcessInstanceInfo
    page := int32(1)
    pageSize := int32(100)
    
    for {
        response, err := client.ListProcessInstances(ctx, &pb.ListProcessInstancesRequest{
            PageSize: pageSize,
            Page:     page,
        })
        
        if err != nil {
            return nil, err
        }
        
        if !response.Success {
            return nil, fmt.Errorf("ошибка на странице %d: %s", page, response.Message)
        }
        
        allInstances = append(allInstances, response.Instances...)
        
        // Проверяем, есть ли еще страницы
        if page >= response.TotalPages {
            break
        }
        
        page++
        
        fmt.Printf("Загружено страница %d/%d (%d процессов)\n", 
            page-1, response.TotalPages, len(response.Instances))
    }
    
    return allInstances, nil
}

// Поиск процессов по критериям
func findProcessesByCriteria(client pb.ProcessServiceClient, ctx context.Context, criteria ProcessSearchCriteria) ([]*pb.ProcessInstanceInfo, error) {
    request := &pb.ListProcessInstancesRequest{
        PageSize:  1000,
        SortBy:    "started_at",
        SortOrder: "DESC",
    }
    
    if criteria.Status != "" {
        request.StatusFilter = criteria.Status
    }
    
    if criteria.ProcessKey != "" {
        request.ProcessKeyFilter = criteria.ProcessKey
    }
    
    response, err := client.ListProcessInstances(ctx, request)
    if err != nil {
        return nil, err
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка поиска: %s", response.Message)
    }
    
    // Дополнительная фильтрация по времени
    var filtered []*pb.ProcessInstanceInfo
    for _, instance := range response.Instances {
        if criteria.StartedAfter > 0 && instance.StartedAt < criteria.StartedAfter {
            continue
        }
        if criteria.StartedBefore > 0 && instance.StartedAt > criteria.StartedBefore {
            continue
        }
        filtered = append(filtered, instance)
    }
    
    return filtered, nil
}

type ProcessSearchCriteria struct {
    Status        string
    ProcessKey    string
    StartedAfter  int64
    StartedBefore int64
}

// Мониторинг активных процессов
func monitorActiveProcesses(client pb.ProcessServiceClient, ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    fmt.Printf("Мониторинг активных процессов (интервал: %v)\n", interval)
    
    for {
        select {
        case <-ticker.C:
            activeProcesses, err := getActiveProcesses(client, ctx)
            if err != nil {
                log.Printf("Ошибка получения активных процессов: %v", err)
                continue
            }
            
            fmt.Printf("[%s] Активных процессов: %d\n", 
                time.Now().Format("15:04:05"), len(activeProcesses))
            
            // Группировка по типам процессов
            processTypes := make(map[string]int)
            for _, process := range activeProcesses {
                processTypes[process.ProcessKey]++
            }
            
            if len(processTypes) > 0 {
                fmt.Println("  По типам:")
                for processKey, count := range processTypes {
                    fmt.Printf("    %s: %d\n", processKey, count)
                }
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
import time
from datetime import datetime, timedelta
from collections import defaultdict

import process_pb2
import process_pb2_grpc

def list_process_instances(filters=None, pagination=None, sorting=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Значения по умолчанию
    if filters is None:
        filters = {}
    if pagination is None:
        pagination = {'page_size': 20, 'page': 1}
    if sorting is None:
        sorting = {'sort_by': 'started_at', 'sort_order': 'DESC'}
    
    request = process_pb2.ListProcessInstancesRequest(
        status_filter=filters.get('status', ''),
        process_key_filter=filters.get('process_key', ''),
        page_size=pagination.get('page_size', 20),
        page=pagination.get('page', 1),
        sort_by=sorting.get('sort_by', 'started_at'),
        sort_order=sorting.get('sort_order', 'DESC')
    )
    
    try:
        response = stub.ListProcessInstances(request, metadata=metadata)
        
        if response.success:
            instances = []
            for instance in response.instances:
                instances.append({
                    'instance_id': instance.instance_id,
                    'process_key': instance.process_key,
                    'status': instance.status,
                    'current_activity': instance.current_activity,
                    'started_at': instance.started_at,
                    'updated_at': instance.updated_at,
                    'variables': dict(instance.variables)
                })
            
            return {
                'instances': instances,
                'total_count': response.total_count,
                'page': response.page,
                'page_size': response.page_size,
                'total_pages': response.total_pages
            }
        else:
            print(f"Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def display_process_instances(instances_data):
    """Отображение списка процессов в читаемом формате"""
    if not instances_data:
        return
    
    instances = instances_data['instances']
    
    print(f"=== Экземпляры процессов (страница {instances_data['page']}/{instances_data['total_pages']}) ===")
    print(f"Всего найдено: {instances_data['total_count']}\n")
    
    for i, instance in enumerate(instances, 1):
        print(f"{i}. {instance['instance_id']}")
        print(f"   Процесс: {instance['process_key']}")
        print(f"   Статус: {get_status_emoji(instance['status'])} {instance['status']}")
        print(f"   Текущая активность: {instance['current_activity']}")
        print(f"   Запущен: {format_timestamp(instance['started_at'])}")
        print(f"   Обновлен: {format_timestamp(instance['updated_at'])}")
        
        if instance['variables']:
            print(f"   Переменных: {len(instance['variables'])}")
        print()

def get_status_emoji(status):
    """Emoji для статуса процесса"""
    emoji_map = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'FAILED': '❌',
        'CANCELLED': '⏹️'
    }
    return emoji_map.get(status, '❓')

def format_timestamp(timestamp):
    """Форматирование Unix timestamp"""
    return datetime.fromtimestamp(timestamp).strftime('%Y-%m-%d %H:%M:%S')

def get_all_processes(max_pages=None):
    """Получение всех процессов с пагинацией"""
    all_instances = []
    page = 1
    page_size = 100
    
    print("Загрузка всех процессов...")
    
    while True:
        result = list_process_instances(
            pagination={'page_size': page_size, 'page': page}
        )
        
        if not result:
            break
        
        all_instances.extend(result['instances'])
        
        print(f"Загружена страница {page}/{result['total_pages']} ({len(result['instances'])} процессов)")
        
        if page >= result['total_pages']:
            break
        
        if max_pages and page >= max_pages:
            print(f"Достигнут лимит страниц: {max_pages}")
            break
        
        page += 1
    
    print(f"Всего загружено: {len(all_instances)} процессов")
    return all_instances

def search_processes_advanced(**criteria):
    """Расширенный поиск процессов"""
    print("Расширенный поиск процессов...")
    
    # Базовые фильтры
    filters = {}
    if 'status' in criteria:
        filters['status'] = criteria['status']
    if 'process_key' in criteria:
        filters['process_key'] = criteria['process_key']
    
    # Получаем все подходящие процессы
    all_processes = []
    page = 1
    
    while True:
        result = list_process_instances(
            filters=filters,
            pagination={'page_size': 100, 'page': page}
        )
        
        if not result or not result['instances']:
            break
        
        all_processes.extend(result['instances'])
        
        if page >= result['total_pages']:
            break
        page += 1
    
    # Дополнительная фильтрация
    filtered_processes = []
    
    for process in all_processes:
        # Фильтр по времени запуска
        if 'started_after' in criteria:
            if process['started_at'] < criteria['started_after']:
                continue
        
        if 'started_before' in criteria:
            if process['started_at'] > criteria['started_before']:
                continue
        
        # Фильтр по переменным
        if 'has_variable' in criteria:
            var_name = criteria['has_variable']
            if var_name not in process['variables']:
                continue
        
        if 'variable_equals' in criteria:
            var_name, var_value = criteria['variable_equals']
            if process['variables'].get(var_name) != var_value:
                continue
        
        # Фильтр по возрасту
        if 'max_age_hours' in criteria:
            max_age = criteria['max_age_hours']
            age_hours = (time.time() - process['started_at']) / 3600
            if age_hours > max_age:
                continue
        
        filtered_processes.append(process)
    
    print(f"Найдено {len(filtered_processes)} процессов после фильтрации")
    return filtered_processes

def get_process_statistics():
    """Статистика по процессам"""
    print("Сбор статистики по процессам...")
    
    # Получаем все процессы
    all_processes = get_all_processes(max_pages=50)  # Ограничиваем для производительности
    
    if not all_processes:
        print("Процессы не найдены")
        return
    
    # Статистика по статусам
    status_counts = defaultdict(int)
    process_key_counts = defaultdict(int)
    
    oldest_start = float('inf')
    newest_start = 0
    
    for process in all_processes:
        status_counts[process['status']] += 1
        process_key_counts[process['process_key']] += 1
        
        oldest_start = min(oldest_start, process['started_at'])
        newest_start = max(newest_start, process['started_at'])
    
    print(f"\n=== Статистика по {len(all_processes)} процессам ===")
    
    print("\nПо статусам:")
    for status, count in sorted(status_counts.items()):
        emoji = get_status_emoji(status)
        percentage = (count / len(all_processes)) * 100
        print(f"  {emoji} {status}: {count} ({percentage:.1f}%)")
    
    print("\nТоп-10 типов процессов:")
    sorted_types = sorted(process_key_counts.items(), key=lambda x: x[1], reverse=True)
    for process_key, count in sorted_types[:10]:
        percentage = (count / len(all_processes)) * 100
        print(f"  {process_key}: {count} ({percentage:.1f}%)")
    
    print(f"\nВременной диапазон:")
    print(f"  Самый старый: {format_timestamp(oldest_start)}")
    print(f"  Самый новый: {format_timestamp(newest_start)}")
    
    # Статистика по возрасту
    current_time = time.time()
    age_ranges = {
        'Менее 1 часа': 0,
        '1-24 часа': 0,
        '1-7 дней': 0,
        '1-30 дней': 0,
        'Более 30 дней': 0
    }
    
    for process in all_processes:
        age_hours = (current_time - process['started_at']) / 3600
        
        if age_hours < 1:
            age_ranges['Менее 1 часа'] += 1
        elif age_hours < 24:
            age_ranges['1-24 часа'] += 1
        elif age_hours < 24 * 7:
            age_ranges['1-7 дней'] += 1
        elif age_hours < 24 * 30:
            age_ranges['1-30 дней'] += 1
        else:
            age_ranges['Более 30 дней'] += 1
    
    print("\nПо возрасту:")
    for range_name, count in age_ranges.items():
        if count > 0:
            percentage = (count / len(all_processes)) * 100
            print(f"  {range_name}: {count} ({percentage:.1f}%)")

def monitor_process_count(interval=30):
    """Мониторинг количества процессов"""
    print(f"Мониторинг количества процессов (интервал: {interval}с)")
    
    previous_counts = {}
    
    try:
        while True:
            # Получаем текущие счетчики
            current_counts = {}
            
            for status in ['ACTIVE', 'COMPLETED', 'FAILED', 'CANCELLED']:
                result = list_process_instances(
                    filters={'status': status},
                    pagination={'page_size': 1}  # Нам нужен только total_count
                )
                
                if result:
                    current_counts[status] = result['total_count']
                else:
                    current_counts[status] = 0
            
            # Отображение текущего состояния
            timestamp = datetime.now().strftime('%H:%M:%S')
            print(f"\n[{timestamp}] === Статистика процессов ===")
            
            total = sum(current_counts.values())
            print(f"Всего процессов: {total}")
            
            for status, count in current_counts.items():
                emoji = get_status_emoji(status)
                change = ""
                
                if status in previous_counts:
                    delta = count - previous_counts[status]
                    if delta > 0:
                        change = f" (+{delta})"
                    elif delta < 0:
                        change = f" ({delta})"
                
                print(f"  {emoji} {status}: {count}{change}")
            
            previous_counts = current_counts.copy()
            time.sleep(interval)
            
    except KeyboardInterrupt:
        print("\nМониторинг остановлен")

if __name__ == "__main__":
    # Примеры использования
    
    # Простое получение списка
    result = list_process_instances()
    if result:
        display_process_instances(result)
    
    # Поиск активных процессов
    active_result = list_process_instances(filters={'status': 'ACTIVE'})
    if active_result:
        print(f"\nАктивных процессов: {active_result['total_count']}")
    
    # Расширенный поиск
    # yesterday = time.time() - 24 * 3600
    # recent_processes = search_processes_advanced(
    #     status='COMPLETED',
    #     started_after=yesterday
    # )
    
    # Статистика
    # get_process_statistics()
    
    # Мониторинг
    # monitor_process_count(interval=60)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'process.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

async function listProcessInstances(options = {}) {
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    const {
        filters = {},
        pagination = { pageSize: 20, page: 1 },
        sorting = { sortBy: 'started_at', sortOrder: 'DESC' }
    } = options;
    
    return new Promise((resolve, reject) => {
        const request = {
            status_filter: filters.status || '',
            process_key_filter: filters.processKey || '',
            page_size: pagination.pageSize || 20,
            page: pagination.page || 1,
            sort_by: sorting.sortBy || 'started_at',
            sort_order: sorting.sortOrder || 'DESC'
        };
        
        client.listProcessInstances(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            const instances = response.instances.map(instance => ({
                instanceId: instance.instance_id,
                processKey: instance.process_key,
                status: instance.status,
                currentActivity: instance.current_activity,
                startedAt: Number(instance.started_at) * 1000, // Convert to JS timestamp
                updatedAt: Number(instance.updated_at) * 1000,
                variables: instance.variables
            }));
            
            resolve({
                instances,
                totalCount: response.total_count,
                page: response.page,
                pageSize: response.page_size,
                totalPages: response.total_pages
            });
        });
    });
}

function getStatusEmoji(status) {
    const emojiMap = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'FAILED': '❌',
        'CANCELLED': '⏹️'
    };
    return emojiMap[status] || '❓';
}

function formatTimestamp(timestamp) {
    return new Date(timestamp).toLocaleString();
}

async function displayProcessInstances(instancesData) {
    if (!instancesData) {
        console.log('Нет данных для отображения');
        return;
    }
    
    const { instances, page, totalPages, totalCount } = instancesData;
    
    console.log(`=== Экземпляры процессов (страница ${page}/${totalPages}) ===`);
    console.log(`Всего найдено: ${totalCount}\n`);
    
    instances.forEach((instance, index) => {
        console.log(`${index + 1}. ${instance.instanceId}`);
        console.log(`   Процесс: ${instance.processKey}`);
        console.log(`   Статус: ${getStatusEmoji(instance.status)} ${instance.status}`);
        console.log(`   Текущая активность: ${instance.currentActivity}`);
        console.log(`   Запущен: ${formatTimestamp(instance.startedAt)}`);
        console.log(`   Обновлен: ${formatTimestamp(instance.updatedAt)}`);
        
        if (instance.variables && Object.keys(instance.variables).length > 0) {
            console.log(`   Переменных: ${Object.keys(instance.variables).length}`);
        }
        console.log();
    });
}

async function getAllProcesses(maxPages = null) {
    console.log('Загрузка всех процессов...');
    
    const allInstances = [];
    let page = 1;
    const pageSize = 100;
    
    while (true) {
        try {
            const result = await listProcessInstances({
                pagination: { pageSize, page }
            });
            
            allInstances.push(...result.instances);
            
            console.log(`Загружена страница ${page}/${result.totalPages} (${result.instances.length} процессов)`);
            
            if (page >= result.totalPages) {
                break;
            }
            
            if (maxPages && page >= maxPages) {
                console.log(`Достигнут лимит страниц: ${maxPages}`);
                break;
            }
            
            page++;
            
        } catch (error) {
            console.error(`Ошибка на странице ${page}:`, error.message);
            break;
        }
    }
    
    console.log(`Всего загружено: ${allInstances.length} процессов`);
    return allInstances;
}

async function searchProcessesAdvanced(criteria) {
    console.log('Расширенный поиск процессов...');
    
    // Базовые фильтры
    const filters = {};
    if (criteria.status) filters.status = criteria.status;
    if (criteria.processKey) filters.processKey = criteria.processKey;
    
    // Получаем все подходящие процессы
    let allProcesses = [];
    let page = 1;
    
    while (true) {
        try {
            const result = await listProcessInstances({
                filters,
                pagination: { pageSize: 100, page }
            });
            
            if (!result.instances.length) break;
            
            allProcesses.push(...result.instances);
            
            if (page >= result.totalPages) break;
            page++;
            
        } catch (error) {
            console.error('Ошибка поиска:', error.message);
            break;
        }
    }
    
    // Дополнительная фильтрация
    const filteredProcesses = allProcesses.filter(process => {
        // Фильтр по времени запуска
        if (criteria.startedAfter && process.startedAt < criteria.startedAfter) {
            return false;
        }
        
        if (criteria.startedBefore && process.startedAt > criteria.startedBefore) {
            return false;
        }
        
        // Фильтр по переменным
        if (criteria.hasVariable && !process.variables[criteria.hasVariable]) {
            return false;
        }
        
        if (criteria.variableEquals) {
            const [varName, varValue] = criteria.variableEquals;
            if (process.variables[varName] !== varValue) {
                return false;
            }
        }
        
        // Фильтр по возрасту
        if (criteria.maxAgeHours) {
            const ageHours = (Date.now() - process.startedAt) / (1000 * 60 * 60);
            if (ageHours > criteria.maxAgeHours) {
                return false;
            }
        }
        
        return true;
    });
    
    console.log(`Найдено ${filteredProcesses.length} процессов после фильтрации`);
    return filteredProcesses;
}

async function getProcessStatistics() {
    console.log('Сбор статистики по процессам...');
    
    try {
        // Получаем все процессы (ограничиваем для производительности)
        const allProcesses = await getAllProcesses(50);
        
        if (!allProcesses.length) {
            console.log('Процессы не найдены');
            return;
        }
        
        // Статистика по статусам
        const statusCounts = {};
        const processKeyCounts = {};
        
        let oldestStart = Date.now();
        let newestStart = 0;
        
        allProcesses.forEach(process => {
            statusCounts[process.status] = (statusCounts[process.status] || 0) + 1;
            processKeyCounts[process.processKey] = (processKeyCounts[process.processKey] || 0) + 1;
            
            oldestStart = Math.min(oldestStart, process.startedAt);
            newestStart = Math.max(newestStart, process.startedAt);
        });
        
        console.log(`\n=== Статистика по ${allProcesses.length} процессам ===`);
        
        console.log('\nПо статусам:');
        Object.entries(statusCounts)
            .sort(([,a], [,b]) => b - a)
            .forEach(([status, count]) => {
                const emoji = getStatusEmoji(status);
                const percentage = ((count / allProcesses.length) * 100).toFixed(1);
                console.log(`  ${emoji} ${status}: ${count} (${percentage}%)`);
            });
        
        console.log('\nТоп-10 типов процессов:');
        Object.entries(processKeyCounts)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 10)
            .forEach(([processKey, count]) => {
                const percentage = ((count / allProcesses.length) * 100).toFixed(1);
                console.log(`  ${processKey}: ${count} (${percentage}%)`);
            });
        
        console.log('\nВременной диапазон:');
        console.log(`  Самый старый: ${formatTimestamp(oldestStart)}`);
        console.log(`  Самый новый: ${formatTimestamp(newestStart)}`);
        
        // Статистика по возрасту
        const currentTime = Date.now();
        const ageRanges = {
            'Менее 1 часа': 0,
            '1-24 часа': 0,
            '1-7 дней': 0,
            '1-30 дней': 0,
            'Более 30 дней': 0
        };
        
        allProcesses.forEach(process => {
            const ageHours = (currentTime - process.startedAt) / (1000 * 60 * 60);
            
            if (ageHours < 1) {
                ageRanges['Менее 1 часа']++;
            } else if (ageHours < 24) {
                ageRanges['1-24 часа']++;
            } else if (ageHours < 24 * 7) {
                ageRanges['1-7 дней']++;
            } else if (ageHours < 24 * 30) {
                ageRanges['1-30 дней']++;
            } else {
                ageRanges['Более 30 дней']++;
            }
        });
        
        console.log('\nПо возрасту:');
        Object.entries(ageRanges).forEach(([rangeName, count]) => {
            if (count > 0) {
                const percentage = ((count / allProcesses.length) * 100).toFixed(1);
                console.log(`  ${rangeName}: ${count} (${percentage}%)`);
            }
        });
        
    } catch (error) {
        console.error('Ошибка сбора статистики:', error.message);
    }
}

async function monitorProcessCount(intervalSeconds = 30) {
    console.log(`Мониторинг количества процессов (интервал: ${intervalSeconds}с)`);
    
    let previousCounts = {};
    
    const monitor = async () => {
        try {
            const currentCounts = {};
            
            // Получаем счетчики для каждого статуса
            for (const status of ['ACTIVE', 'COMPLETED', 'FAILED', 'CANCELLED']) {
                try {
                    const result = await listProcessInstances({
                        filters: { status },
                        pagination: { pageSize: 1 } // Нам нужен только total_count
                    });
                    
                    currentCounts[status] = result.totalCount;
                } catch (error) {
                    currentCounts[status] = 0;
                }
            }
            
            // Отображение текущего состояния
            const timestamp = new Date().toLocaleTimeString();
            console.log(`\n[${timestamp}] === Статистика процессов ===`);
            
            const total = Object.values(currentCounts).reduce((sum, count) => sum + count, 0);
            console.log(`Всего процессов: ${total}`);
            
            Object.entries(currentCounts).forEach(([status, count]) => {
                const emoji = getStatusEmoji(status);
                let change = '';
                
                if (previousCounts[status] !== undefined) {
                    const delta = count - previousCounts[status];
                    if (delta > 0) {
                        change = ` (+${delta})`;
                    } else if (delta < 0) {
                        change = ` (${delta})`;
                    }
                }
                
                console.log(`  ${emoji} ${status}: ${count}${change}`);
            });
            
            previousCounts = { ...currentCounts };
            
        } catch (error) {
            console.error('Ошибка мониторинга:', error.message);
        }
    };
    
    // Первый запуск
    await monitor();
    
    // Периодический мониторинг
    const interval = setInterval(monitor, intervalSeconds * 1000);
    
    // Обработка сигнала завершения
    process.on('SIGINT', () => {
        clearInterval(interval);
        console.log('\nМониторинг остановлен');
        process.exit(0);
    });
}

// Примеры использования
if (require.main === module) {
    const command = process.argv[2];
    
    switch (command) {
        case 'list':
            listProcessInstances().then(displayProcessInstances);
            break;
            
        case 'active':
            listProcessInstances({ 
                filters: { status: 'ACTIVE' } 
            }).then(displayProcessInstances);
            break;
            
        case 'stats':
            getProcessStatistics();
            break;
            
        case 'monitor':
            const interval = parseInt(process.argv[3]) || 30;
            monitorProcessCount(interval);
            break;
            
        case 'search':
            // Пример: node list.js search --status=COMPLETED --hours=24
            const criteria = {};
            process.argv.slice(3).forEach(arg => {
                if (arg.startsWith('--status=')) {
                    criteria.status = arg.split('=')[1];
                }
                if (arg.startsWith('--hours=')) {
                    criteria.maxAgeHours = parseInt(arg.split('=')[1]);
                }
                if (arg.startsWith('--key=')) {
                    criteria.processKey = arg.split('=')[1];
                }
            });
            
            searchProcessesAdvanced(criteria)
                .then(processes => {
                    console.log(`Найдено ${processes.length} процессов`);
                    processes.slice(0, 10).forEach((process, index) => {
                        console.log(`${index + 1}. ${process.instanceId} (${process.status})`);
                    });
                });
            break;
            
        default:
            console.log('Использование:');
            console.log('  node list.js list                    - список процессов');
            console.log('  node list.js active                  - только активные');
            console.log('  node list.js stats                   - статистика');
            console.log('  node list.js monitor [interval]      - мониторинг');
            console.log('  node list.js search [options]        - поиск');
            break;
    }
}

module.exports = {
    listProcessInstances,
    displayProcessInstances,
    getAllProcesses,
    searchProcessesAdvanced,
    getProcessStatistics,
    monitorProcessCount
};
```

## Фильтрация и сортировка

### Доступные фильтры
```json
{
  "status_filter": "ACTIVE|COMPLETED|CANCELLED|FAILED",
  "process_key_filter": "order-process"
}
```

### Поля сортировки
- `started_at` - время запуска (по умолчанию)
- `updated_at` - время обновления  
- `status` - статус процесса
- `process_key` - ключ процесса

### Пагинация
```json
{
  "page_size": 50,     // 1-1000, по умолчанию 20
  "page": 2,           // начиная с 1
  "sort_order": "ASC"  // ASC или DESC
}
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры пагинации или сортировки
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Invalid page_size: must be between 1 and 1000"
}
```

## Связанные методы
- [GetProcessInstanceStatus](get-process-instance-status.md) - Статус конкретного экземпляра
- [StartProcessInstance](start-process-instance.md) - Запуск нового экземпляра
- [CancelProcessInstance](cancel-process-instance.md) - Отмена экземпляра
- [ListTokens](list-tokens.md) - Токены экземпляров
