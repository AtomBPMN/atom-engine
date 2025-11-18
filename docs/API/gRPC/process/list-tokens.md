# ListTokens

## Описание
Получает список токенов выполнения с поддержкой фильтрации по экземпляру процесса, состоянию и пагинации.

## Синтаксис
```protobuf
rpc ListTokens(ListTokensRequest) returns (ListTokensResponse);
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

### ListTokensRequest
```protobuf
message ListTokensRequest {
  string instance_id_filter = 1;   // Фильтр по ID экземпляра процесса
  string state_filter = 2;         // Фильтр по состоянию токена
  int32 limit = 3;                 // Лимит записей (устарел, используйте page_size)
  int32 page_size = 4;             // Размер страницы (по умолчанию: 20)
  int32 page = 5;                  // Номер страницы (начиная с 1)
  string sort_by = 6;              // Поле сортировки (по умолчанию: "created_at")
  string sort_order = 7;           // Порядок сортировки: "ASC" или "DESC"
}
```

#### Поля:
- **instance_id_filter** (string, optional): Фильтр по ID экземпляра процесса
- **state_filter** (string, optional): Фильтр по состоянию (`ACTIVE`, `COMPLETED`, `CANCELLED`)
- **limit** (int32, deprecated): Используйте `page_size`
- **page_size** (int32, optional): Количество записей на странице (1-1000, по умолчанию: 20)
- **page** (int32, optional): Номер страницы (начиная с 1)
- **sort_by** (string, optional): Поле сортировки (`created_at`, `updated_at`, `state`, `element_id`)
- **sort_order** (string, optional): Порядок сортировки (`ASC`, `DESC`)

## Параметры ответа

### ListTokensResponse
```protobuf
message ListTokensResponse {
  repeated TokenInfo tokens = 1;   // Список токенов
  bool success = 2;                // Статус успешности
  string message = 3;              // Сообщение о результате
  int32 total_count = 4;           // Общее количество токенов
  int32 page = 5;                  // Текущая страница
  int32 page_size = 6;             // Размер страницы
  int32 total_pages = 7;           // Общее количество страниц
}

message TokenInfo {
  string token_id = 1;                    // ID токена
  string process_instance_id = 2;         // ID экземпляра процесса
  string process_key = 3;                 // Ключ процесса
  string current_element_id = 4;          // ID текущего элемента BPMN
  string state = 5;                       // Состояние токена
  string waiting_for = 6;                 // Что ожидает токен
  int64 created_at = 7;                   // Время создания (Unix timestamp)
  int64 updated_at = 8;                   // Время обновления (Unix timestamp)
  map<string, string> variables = 9;      // Переменные токена
}
```

#### Состояния токена:
- **ACTIVE** - Токен активен, выполняется
- **COMPLETED** - Токен завершен
- **CANCELLED** - Токен отменен
- **WAITING** - Токен ожидает события/условия

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
    
    // Получение всех токенов
    response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
        PageSize:  50,
        Page:      1,
        SortBy:    "created_at",
        SortOrder: "DESC",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("=== Токены выполнения (страница %d/%d) ===\n", 
            response.Page, response.TotalPages)
        fmt.Printf("Всего найдено: %d\n\n", response.TotalCount)
        
        for i, token := range response.Tokens {
            fmt.Printf("%d. %s\n", i+1, token.TokenId)
            fmt.Printf("   Экземпляр: %s\n", token.ProcessInstanceId)
            fmt.Printf("   Процесс: %s\n", token.ProcessKey)
            fmt.Printf("   Элемент: %s\n", token.CurrentElementId)
            fmt.Printf("   Состояние: %s\n", token.State)
            if token.WaitingFor != "" {
                fmt.Printf("   Ожидает: %s\n", token.WaitingFor)
            }
            fmt.Printf("   Создан: %s\n", formatTimestamp(token.CreatedAt))
            fmt.Printf("   Обновлен: %s\n", formatTimestamp(token.UpdatedAt))
            
            if len(token.Variables) > 0 {
                fmt.Printf("   Переменных: %d\n", len(token.Variables))
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

// Получение активных токенов для экземпляра процесса
func getActiveTokensForInstance(client pb.ProcessServiceClient, ctx context.Context, instanceId string) ([]*pb.TokenInfo, error) {
    response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
        InstanceIdFilter: instanceId,
        StateFilter:      "ACTIVE",
        PageSize:        1000,
        SortBy:          "created_at",
        SortOrder:       "ASC",
    })
    
    if err != nil {
        return nil, err
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка получения токенов: %s", response.Message)
    }
    
    return response.Tokens, nil
}

// Анализ токенов процесса
func analyzeProcessTokens(client pb.ProcessServiceClient, ctx context.Context, instanceId string) {
    fmt.Printf("=== Анализ токенов процесса %s ===\n", instanceId)
    
    // Получаем все токены для процесса
    response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
        InstanceIdFilter: instanceId,
        PageSize:        1000,
    })
    
    if err != nil {
        log.Printf("Ошибка: %v", err)
        return
    }
    
    if !response.Success {
        log.Printf("Ошибка получения токенов: %s", response.Message)
        return
    }
    
    tokens := response.Tokens
    fmt.Printf("Всего токенов: %d\n", len(tokens))
    
    // Статистика по состояниям
    stateCounts := make(map[string]int)
    elementCounts := make(map[string]int)
    
    var oldestToken, newestToken *pb.TokenInfo
    
    for _, token := range tokens {
        stateCounts[token.State]++
        elementCounts[token.CurrentElementId]++
        
        if oldestToken == nil || token.CreatedAt < oldestToken.CreatedAt {
            oldestToken = token
        }
        
        if newestToken == nil || token.CreatedAt > newestToken.CreatedAt {
            newestToken = token
        }
    }
    
    fmt.Println("\nСтатистика по состояниям:")
    for state, count := range stateCounts {
        fmt.Printf("  %s: %d\n", state, count)
    }
    
    fmt.Println("\nТоп элементов по количеству токенов:")
    type elementStat struct {
        elementId string
        count     int
    }
    
    var sortedElements []elementStat
    for elementId, count := range elementCounts {
        sortedElements = append(sortedElements, elementStat{elementId, count})
    }
    
    // Простая сортировка
    for i := 0; i < len(sortedElements)-1; i++ {
        for j := i + 1; j < len(sortedElements); j++ {
            if sortedElements[i].count < sortedElements[j].count {
                sortedElements[i], sortedElements[j] = sortedElements[j], sortedElements[i]
            }
        }
    }
    
    for i, element := range sortedElements {
        if i >= 10 { // Топ 10
            break
        }
        fmt.Printf("  %s: %d токенов\n", element.elementId, element.count)
    }
    
    if oldestToken != nil && newestToken != nil {
        fmt.Printf("\nВременной диапазон:\n")
        fmt.Printf("  Самый старый токен: %s (%s)\n", 
            oldestToken.TokenId, formatTimestamp(oldestToken.CreatedAt))
        fmt.Printf("  Самый новый токен: %s (%s)\n", 
            newestToken.TokenId, formatTimestamp(newestToken.CreatedAt))
        
        duration := time.Unix(newestToken.CreatedAt, 0).Sub(time.Unix(oldestToken.CreatedAt, 0))
        fmt.Printf("  Длительность выполнения: %v\n", duration)
    }
}

// Трассировка выполнения процесса
func traceProcessExecution(client pb.ProcessServiceClient, ctx context.Context, instanceId string) {
    fmt.Printf("=== Трассировка выполнения процесса %s ===\n", instanceId)
    
    // Получаем все токены в хронологическом порядке
    response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
        InstanceIdFilter: instanceId,
        PageSize:        1000,
        SortBy:          "created_at",
        SortOrder:       "ASC",
    })
    
    if err != nil {
        log.Printf("Ошибка: %v", err)
        return
    }
    
    if !response.Success {
        log.Printf("Ошибка получения токенов: %s", response.Message)
        return
    }
    
    fmt.Printf("Найдено %d токенов\n\n", len(response.Tokens))
    
    // Группировка токенов по элементам для показа потока
    elementFlow := make(map[string][]*pb.TokenInfo)
    
    for _, token := range response.Tokens {
        elementFlow[token.CurrentElementId] = append(elementFlow[token.CurrentElementId], token)
    }
    
    fmt.Println("Поток выполнения:")
    
    // Отображение в хронологическом порядке
    for i, token := range response.Tokens {
        created := time.Unix(token.CreatedAt, 0)
        updated := time.Unix(token.UpdatedAt, 0)
        
        fmt.Printf("%d. [%s] %s\n", i+1, created.Format("15:04:05"), token.CurrentElementId)
        fmt.Printf("   Токен: %s (%s)\n", token.TokenId, token.State)
        
        if token.WaitingFor != "" {
            fmt.Printf("   Ожидает: %s\n", token.WaitingFor)
        }
        
        if !created.Equal(updated) {
            duration := updated.Sub(created)
            fmt.Printf("   Длительность: %v\n", duration)
        }
        
        // Показываем ключевые переменные
        if len(token.Variables) > 0 {
            fmt.Printf("   Переменные: ")
            count := 0
            for key, value := range token.Variables {
                if count < 3 { // Показываем первые 3
                    fmt.Printf("%s=%s ", key, value)
                    count++
                } else {
                    fmt.Printf("... (+%d)", len(token.Variables)-3)
                    break
                }
            }
            fmt.Println()
        }
        
        fmt.Println()
    }
}

// Мониторинг активных токенов
func monitorActiveTokens(client pb.ProcessServiceClient, ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    fmt.Printf("Мониторинг активных токенов (интервал: %v)\n", interval)
    
    for {
        select {
        case <-ticker.C:
            response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
                StateFilter: "ACTIVE",
                PageSize:   1000,
            })
            
            if err != nil {
                log.Printf("Ошибка получения токенов: %v", err)
                continue
            }
            
            if !response.Success {
                log.Printf("Ошибка: %s", response.Message)
                continue
            }
            
            fmt.Printf("[%s] Активных токенов: %d\n", 
                time.Now().Format("15:04:05"), len(response.Tokens))
            
            // Группировка по процессам
            processTokens := make(map[string]int)
            elementTokens := make(map[string]int)
            
            for _, token := range response.Tokens {
                processTokens[token.ProcessKey]++
                elementTokens[token.CurrentElementId]++
            }
            
            if len(processTokens) > 0 {
                fmt.Println("  По процессам:")
                for processKey, count := range processTokens {
                    fmt.Printf("    %s: %d токенов\n", processKey, count)
                }
            }
            
            if len(elementTokens) > 5 { // Показываем только если много разных элементов
                fmt.Printf("  Активных элементов: %d\n", len(elementTokens))
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
from datetime import datetime
from collections import defaultdict

import process_pb2
import process_pb2_grpc

def list_tokens(filters=None, pagination=None, sorting=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Значения по умолчанию
    if filters is None:
        filters = {}
    if pagination is None:
        pagination = {'page_size': 20, 'page': 1}
    if sorting is None:
        sorting = {'sort_by': 'created_at', 'sort_order': 'DESC'}
    
    request = process_pb2.ListTokensRequest(
        instance_id_filter=filters.get('instance_id', ''),
        state_filter=filters.get('state', ''),
        page_size=pagination.get('page_size', 20),
        page=pagination.get('page', 1),
        sort_by=sorting.get('sort_by', 'created_at'),
        sort_order=sorting.get('sort_order', 'DESC')
    )
    
    try:
        response = stub.ListTokens(request, metadata=metadata)
        
        if response.success:
            tokens = []
            for token in response.tokens:
                tokens.append({
                    'token_id': token.token_id,
                    'process_instance_id': token.process_instance_id,
                    'process_key': token.process_key,
                    'current_element_id': token.current_element_id,
                    'state': token.state,
                    'waiting_for': token.waiting_for,
                    'created_at': token.created_at,
                    'updated_at': token.updated_at,
                    'variables': dict(token.variables)
                })
            
            return {
                'tokens': tokens,
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

def display_tokens(tokens_data):
    """Отображение токенов в читаемом формате"""
    if not tokens_data:
        return
    
    tokens = tokens_data['tokens']
    
    print(f"=== Токены выполнения (страница {tokens_data['page']}/{tokens_data['total_pages']}) ===")
    print(f"Всего найдено: {tokens_data['total_count']}\n")
    
    for i, token in enumerate(tokens, 1):
        print(f"{i}. {token['token_id']}")
        print(f"   Экземпляр: {token['process_instance_id']}")
        print(f"   Процесс: {token['process_key']}")
        print(f"   Элемент: {token['current_element_id']}")
        print(f"   Состояние: {get_state_emoji(token['state'])} {token['state']}")
        
        if token['waiting_for']:
            print(f"   Ожидает: {token['waiting_for']}")
        
        print(f"   Создан: {format_timestamp(token['created_at'])}")
        print(f"   Обновлен: {format_timestamp(token['updated_at'])}")
        
        if token['variables']:
            print(f"   Переменных: {len(token['variables'])}")
        print()

def get_state_emoji(state):
    """Emoji для состояния токена"""
    emoji_map = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'CANCELLED': '⏹️',
        'WAITING': '⏳'
    }
    return emoji_map.get(state, '❓')

def format_timestamp(timestamp):
    """Форматирование Unix timestamp"""
    return datetime.fromtimestamp(timestamp).strftime('%Y-%m-%d %H:%M:%S')

def analyze_process_tokens(instance_id):
    """Анализ токенов конкретного процесса"""
    print(f"=== Анализ токенов процесса {instance_id} ===")
    
    # Получаем все токены для процесса
    tokens_data = list_tokens(
        filters={'instance_id': instance_id},
        pagination={'page_size': 1000}
    )
    
    if not tokens_data:
        print("Токены не найдены")
        return
    
    tokens = tokens_data['tokens']
    print(f"Всего токенов: {len(tokens)}")
    
    # Статистика по состояниям
    state_counts = defaultdict(int)
    element_counts = defaultdict(int)
    
    oldest_token = None
    newest_token = None
    
    for token in tokens:
        state_counts[token['state']] += 1
        element_counts[token['current_element_id']] += 1
        
        if not oldest_token or token['created_at'] < oldest_token['created_at']:
            oldest_token = token
        
        if not newest_token or token['created_at'] > newest_token['created_at']:
            newest_token = token
    
    print("\nСтатистика по состояниям:")
    for state, count in state_counts.items():
        emoji = get_state_emoji(state)
        print(f"  {emoji} {state}: {count}")
    
    print("\nТоп элементов по количеству токенов:")
    sorted_elements = sorted(element_counts.items(), key=lambda x: x[1], reverse=True)
    for element_id, count in sorted_elements[:10]:
        print(f"  {element_id}: {count} токенов")
    
    if oldest_token and newest_token:
        print(f"\nВременной диапазон:")
        print(f"  Самый старый токен: {oldest_token['token_id']} ({format_timestamp(oldest_token['created_at'])})")
        print(f"  Самый новый токен: {newest_token['token_id']} ({format_timestamp(newest_token['created_at'])})")
        
        duration_seconds = newest_token['created_at'] - oldest_token['created_at']
        duration = format_duration(duration_seconds)
        print(f"  Длительность выполнения: {duration}")

def format_duration(seconds):
    """Форматирование длительности в секундах"""
    if seconds < 60:
        return f"{seconds:.1f}с"
    elif seconds < 3600:
        return f"{seconds/60:.1f}м"
    elif seconds < 86400:
        return f"{seconds/3600:.1f}ч"
    else:
        return f"{seconds/86400:.1f}д"

def trace_process_execution(instance_id):
    """Трассировка выполнения процесса"""
    print(f"=== Трассировка выполнения процесса {instance_id} ===")
    
    # Получаем все токены в хронологическом порядке
    tokens_data = list_tokens(
        filters={'instance_id': instance_id},
        pagination={'page_size': 1000},
        sorting={'sort_by': 'created_at', 'sort_order': 'ASC'}
    )
    
    if not tokens_data:
        print("Токены не найдены")
        return
    
    tokens = tokens_data['tokens']
    print(f"Найдено {len(tokens)} токенов\n")
    
    print("Поток выполнения:")
    
    for i, token in enumerate(tokens, 1):
        created = datetime.fromtimestamp(token['created_at'])
        updated = datetime.fromtimestamp(token['updated_at'])
        
        print(f"{i}. [{created.strftime('%H:%M:%S')}] {token['current_element_id']}")
        print(f"   Токен: {token['token_id']} ({token['state']})")
        
        if token['waiting_for']:
            print(f"   Ожидает: {token['waiting_for']}")
        
        if created != updated:
            duration = updated - created
            print(f"   Длительность: {duration}")
        
        # Показываем ключевые переменные
        if token['variables']:
            vars_preview = list(token['variables'].items())[:3]
            vars_str = ', '.join([f"{k}={v}" for k, v in vars_preview])
            if len(token['variables']) > 3:
                vars_str += f" ... (+{len(token['variables']) - 3})"
            print(f"   Переменные: {vars_str}")
        
        print()

def get_active_tokens_summary():
    """Сводка по активным токенам"""
    print("=== Сводка по активным токенам ===")
    
    tokens_data = list_tokens(
        filters={'state': 'ACTIVE'},
        pagination={'page_size': 1000}
    )
    
    if not tokens_data:
        print("Активные токены не найдены")
        return
    
    tokens = tokens_data['tokens']
    print(f"Всего активных токенов: {len(tokens)}")
    
    # Группировка по процессам
    process_tokens = defaultdict(list)
    element_tokens = defaultdict(int)
    
    for token in tokens:
        process_tokens[token['process_key']].append(token)
        element_tokens[token['current_element_id']] += 1
    
    print(f"\nПо процессам:")
    for process_key, process_token_list in process_tokens.items():
        print(f"  {process_key}: {len(process_token_list)} токенов")
        
        # Группировка по экземплярам
        instances = defaultdict(int)
        for token in process_token_list:
            instances[token['process_instance_id']] += 1
        
        if len(instances) > 1:
            print(f"    Экземпляров: {len(instances)}")
    
    print(f"\nТоп элементов с активными токенами:")
    sorted_elements = sorted(element_tokens.items(), key=lambda x: x[1], reverse=True)
    for element_id, count in sorted_elements[:10]:
        print(f"  {element_id}: {count} токенов")

def monitor_token_activity(interval=30):
    """Мониторинг активности токенов"""
    print(f"Мониторинг активности токенов (интервал: {interval}с)")
    
    previous_counts = {}
    
    try:
        while True:
            # Получаем счетчики по состояниям
            current_counts = {}
            
            for state in ['ACTIVE', 'COMPLETED', 'CANCELLED', 'WAITING']:
                tokens_data = list_tokens(
                    filters={'state': state},
                    pagination={'page_size': 1}  # Нужен только total_count
                )
                
                if tokens_data:
                    current_counts[state] = tokens_data['total_count']
                else:
                    current_counts[state] = 0
            
            # Отображение состояния
            timestamp = datetime.now().strftime('%H:%M:%S')
            print(f"\n[{timestamp}] === Активность токенов ===")
            
            total = sum(current_counts.values())
            print(f"Всего токенов: {total}")
            
            for state, count in current_counts.items():
                emoji = get_state_emoji(state)
                change = ""
                
                if state in previous_counts:
                    delta = count - previous_counts[state]
                    if delta > 0:
                        change = f" (+{delta})"
                    elif delta < 0:
                        change = f" ({delta})"
                
                print(f"  {emoji} {state}: {count}{change}")
            
            # Дополнительная информация об активных токенах
            if current_counts['ACTIVE'] > 0:
                active_tokens_data = list_tokens(
                    filters={'state': 'ACTIVE'},
                    pagination={'page_size': 100}
                )
                
                if active_tokens_data:
                    processes = defaultdict(int)
                    for token in active_tokens_data['tokens']:
                        processes[token['process_key']] += 1
                    
                    if len(processes) > 1:
                        print(f"  Активные процессы: {len(processes)}")
            
            previous_counts = current_counts.copy()
            time.sleep(interval)
            
    except KeyboardInterrupt:
        print("\nМониторинг остановлен")

def find_stuck_tokens(max_age_hours=24):
    """Поиск зависших токенов"""
    print(f"Поиск токенов, активных более {max_age_hours} часов...")
    
    tokens_data = list_tokens(
        filters={'state': 'ACTIVE'},
        pagination={'page_size': 1000}
    )
    
    if not tokens_data:
        print("Активные токены не найдены")
        return []
    
    current_time = time.time()
    cutoff_time = current_time - (max_age_hours * 3600)
    
    stuck_tokens = []
    for token in tokens_data['tokens']:
        if token['created_at'] < cutoff_time:
            age_hours = (current_time - token['created_at']) / 3600
            stuck_tokens.append({
                'token': token,
                'age_hours': age_hours
            })
    
    if stuck_tokens:
        print(f"Найдено {len(stuck_tokens)} зависших токенов:")
        
        # Сортируем по возрасту
        stuck_tokens.sort(key=lambda x: x['age_hours'], reverse=True)
        
        for item in stuck_tokens:
            token = item['token']
            age = item['age_hours']
            
            print(f"\n🔴 {token['token_id']} (возраст: {age:.1f}ч)")
            print(f"   Процесс: {token['process_key']}")
            print(f"   Экземпляр: {token['process_instance_id']}")
            print(f"   Элемент: {token['current_element_id']}")
            print(f"   Ожидает: {token['waiting_for'] or 'неизвестно'}")
    else:
        print("Зависших токенов не найдено")
    
    return stuck_tokens

if __name__ == "__main__":
    # Примеры использования
    
    # Простое получение списка токенов
    result = list_tokens()
    if result:
        display_tokens(result)
    
    # Анализ токенов конкретного процесса
    # analyze_process_tokens("srv1-aB3dEf9hK2mN5pQ8uV")
    
    # Трассировка выполнения
    # trace_process_execution("srv1-aB3dEf9hK2mN5pQ8uV")
    
    # Сводка по активным токенам
    # get_active_tokens_summary()
    
    # Поиск зависших токенов
    # stuck = find_stuck_tokens(max_age_hours=1)
    
    # Мониторинг активности
    # monitor_token_activity(interval=60)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'process.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

async function listTokens(options = {}) {
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    const {
        filters = {},
        pagination = { pageSize: 20, page: 1 },
        sorting = { sortBy: 'created_at', sortOrder: 'DESC' }
    } = options;
    
    return new Promise((resolve, reject) => {
        const request = {
            instance_id_filter: filters.instanceId || '',
            state_filter: filters.state || '',
            page_size: pagination.pageSize || 20,
            page: pagination.page || 1,
            sort_by: sorting.sortBy || 'created_at',
            sort_order: sorting.sortOrder || 'DESC'
        };
        
        client.listTokens(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            const tokens = response.tokens.map(token => ({
                tokenId: token.token_id,
                processInstanceId: token.process_instance_id,
                processKey: token.process_key,
                currentElementId: token.current_element_id,
                state: token.state,
                waitingFor: token.waiting_for,
                createdAt: Number(token.created_at) * 1000, // Convert to JS timestamp
                updatedAt: Number(token.updated_at) * 1000,
                variables: token.variables
            }));
            
            resolve({
                tokens,
                totalCount: response.total_count,
                page: response.page,
                pageSize: response.page_size,
                totalPages: response.total_pages
            });
        });
    });
}

function getStateEmoji(state) {
    const emojiMap = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'CANCELLED': '⏹️',
        'WAITING': '⏳'
    };
    return emojiMap[state] || '❓';
}

function formatTimestamp(timestamp) {
    return new Date(timestamp).toLocaleString();
}

function formatDuration(milliseconds) {
    const seconds = milliseconds / 1000;
    
    if (seconds < 60) {
        return `${seconds.toFixed(1)}с`;
    } else if (seconds < 3600) {
        return `${(seconds / 60).toFixed(1)}м`;
    } else if (seconds < 86400) {
        return `${(seconds / 3600).toFixed(1)}ч`;
    } else {
        return `${(seconds / 86400).toFixed(1)}д`;
    }
}

async function displayTokens(tokensData) {
    if (!tokensData) {
        console.log('Нет данных для отображения');
        return;
    }
    
    const { tokens, page, totalPages, totalCount } = tokensData;
    
    console.log(`=== Токены выполнения (страница ${page}/${totalPages}) ===`);
    console.log(`Всего найдено: ${totalCount}\n`);
    
    tokens.forEach((token, index) => {
        console.log(`${index + 1}. ${token.tokenId}`);
        console.log(`   Экземпляр: ${token.processInstanceId}`);
        console.log(`   Процесс: ${token.processKey}`);
        console.log(`   Элемент: ${token.currentElementId}`);
        console.log(`   Состояние: ${getStateEmoji(token.state)} ${token.state}`);
        
        if (token.waitingFor) {
            console.log(`   Ожидает: ${token.waitingFor}`);
        }
        
        console.log(`   Создан: ${formatTimestamp(token.createdAt)}`);
        console.log(`   Обновлен: ${formatTimestamp(token.updatedAt)}`);
        
        if (token.variables && Object.keys(token.variables).length > 0) {
            console.log(`   Переменных: ${Object.keys(token.variables).length}`);
        }
        console.log();
    });
}

async function analyzeProcessTokens(instanceId) {
    console.log(`=== Анализ токенов процесса ${instanceId} ===`);
    
    try {
        // Получаем все токены для процесса
        const tokensData = await listTokens({
            filters: { instanceId },
            pagination: { pageSize: 1000 }
        });
        
        if (!tokensData || !tokensData.tokens.length) {
            console.log('Токены не найдены');
            return;
        }
        
        const tokens = tokensData.tokens;
        console.log(`Всего токенов: ${tokens.length}`);
        
        // Статистика по состояниям
        const stateCounts = {};
        const elementCounts = {};
        
        let oldestToken = null;
        let newestToken = null;
        
        tokens.forEach(token => {
            stateCounts[token.state] = (stateCounts[token.state] || 0) + 1;
            elementCounts[token.currentElementId] = (elementCounts[token.currentElementId] || 0) + 1;
            
            if (!oldestToken || token.createdAt < oldestToken.createdAt) {
                oldestToken = token;
            }
            
            if (!newestToken || token.createdAt > newestToken.createdAt) {
                newestToken = token;
            }
        });
        
        console.log('\nСтатистика по состояниям:');
        Object.entries(stateCounts).forEach(([state, count]) => {
            const emoji = getStateEmoji(state);
            console.log(`  ${emoji} ${state}: ${count}`);
        });
        
        console.log('\nТоп элементов по количеству токенов:');
        const sortedElements = Object.entries(elementCounts)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 10);
        
        sortedElements.forEach(([elementId, count]) => {
            console.log(`  ${elementId}: ${count} токенов`);
        });
        
        if (oldestToken && newestToken) {
            console.log('\nВременной диапазон:');
            console.log(`  Самый старый токен: ${oldestToken.tokenId} (${formatTimestamp(oldestToken.createdAt)})`);
            console.log(`  Самый новый токен: ${newestToken.tokenId} (${formatTimestamp(newestToken.createdAt)})`);
            
            const duration = newestToken.createdAt - oldestToken.createdAt;
            console.log(`  Длительность выполнения: ${formatDuration(duration)}`);
        }
        
    } catch (error) {
        console.error('Ошибка анализа токенов:', error.message);
    }
}

async function traceProcessExecution(instanceId) {
    console.log(`=== Трассировка выполнения процесса ${instanceId} ===`);
    
    try {
        // Получаем все токены в хронологическом порядке
        const tokensData = await listTokens({
            filters: { instanceId },
            pagination: { pageSize: 1000 },
            sorting: { sortBy: 'created_at', sortOrder: 'ASC' }
        });
        
        if (!tokensData || !tokensData.tokens.length) {
            console.log('Токены не найдены');
            return;
        }
        
        const tokens = tokensData.tokens;
        console.log(`Найдено ${tokens.length} токенов\n`);
        
        console.log('Поток выполнения:');
        
        tokens.forEach((token, index) => {
            const created = new Date(token.createdAt);
            const updated = new Date(token.updatedAt);
            
            console.log(`${index + 1}. [${created.toLocaleTimeString()}] ${token.currentElementId}`);
            console.log(`   Токен: ${token.tokenId} (${token.state})`);
            
            if (token.waitingFor) {
                console.log(`   Ожидает: ${token.waitingFor}`);
            }
            
            if (created.getTime() !== updated.getTime()) {
                const duration = updated.getTime() - created.getTime();
                console.log(`   Длительность: ${formatDuration(duration)}`);
            }
            
            // Показываем ключевые переменные
            if (token.variables && Object.keys(token.variables).length > 0) {
                const varEntries = Object.entries(token.variables).slice(0, 3);
                let varsStr = varEntries.map(([k, v]) => `${k}=${v}`).join(', ');
                
                if (Object.keys(token.variables).length > 3) {
                    varsStr += ` ... (+${Object.keys(token.variables).length - 3})`;
                }
                
                console.log(`   Переменные: ${varsStr}`);
            }
            
            console.log();
        });
        
    } catch (error) {
        console.error('Ошибка трассировки:', error.message);
    }
}

async function getActiveTokensSummary() {
    console.log('=== Сводка по активным токенам ===');
    
    try {
        const tokensData = await listTokens({
            filters: { state: 'ACTIVE' },
            pagination: { pageSize: 1000 }
        });
        
        if (!tokensData || !tokensData.tokens.length) {
            console.log('Активные токены не найдены');
            return;
        }
        
        const tokens = tokensData.tokens;
        console.log(`Всего активных токенов: ${tokens.length}`);
        
        // Группировка по процессам
        const processTokens = {};
        const elementTokens = {};
        
        tokens.forEach(token => {
            if (!processTokens[token.processKey]) {
                processTokens[token.processKey] = [];
            }
            processTokens[token.processKey].push(token);
            
            elementTokens[token.currentElementId] = (elementTokens[token.currentElementId] || 0) + 1;
        });
        
        console.log('\nПо процессам:');
        Object.entries(processTokens).forEach(([processKey, processTokensList]) => {
            console.log(`  ${processKey}: ${processTokensList.length} токенов`);
            
            // Группировка по экземплярам
            const instances = {};
            processTokensList.forEach(token => {
                instances[token.processInstanceId] = (instances[token.processInstanceId] || 0) + 1;
            });
            
            if (Object.keys(instances).length > 1) {
                console.log(`    Экземпляров: ${Object.keys(instances).length}`);
            }
        });
        
        console.log('\nТоп элементов с активными токенами:');
        const sortedElements = Object.entries(elementTokens)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 10);
        
        sortedElements.forEach(([elementId, count]) => {
            console.log(`  ${elementId}: ${count} токенов`);
        });
        
    } catch (error) {
        console.error('Ошибка получения сводки:', error.message);
    }
}

async function findStuckTokens(maxAgeHours = 24) {
    console.log(`Поиск токенов, активных более ${maxAgeHours} часов...`);
    
    try {
        const tokensData = await listTokens({
            filters: { state: 'ACTIVE' },
            pagination: { pageSize: 1000 }
        });
        
        if (!tokensData || !tokensData.tokens.length) {
            console.log('Активные токены не найдены');
            return [];
        }
        
        const currentTime = Date.now();
        const cutoffTime = currentTime - (maxAgeHours * 60 * 60 * 1000);
        
        const stuckTokens = tokensData.tokens
            .filter(token => token.createdAt < cutoffTime)
            .map(token => ({
                token,
                ageHours: (currentTime - token.createdAt) / (1000 * 60 * 60)
            }))
            .sort((a, b) => b.ageHours - a.ageHours);
        
        if (stuckTokens.length > 0) {
            console.log(`Найдено ${stuckTokens.length} зависших токенов:`);
            
            stuckTokens.forEach(({ token, ageHours }) => {
                console.log(`\n🔴 ${token.tokenId} (возраст: ${ageHours.toFixed(1)}ч)`);
                console.log(`   Процесс: ${token.processKey}`);
                console.log(`   Экземпляр: ${token.processInstanceId}`);
                console.log(`   Элемент: ${token.currentElementId}`);
                console.log(`   Ожидает: ${token.waitingFor || 'неизвестно'}`);
            });
        } else {
            console.log('Зависших токенов не найдено');
        }
        
        return stuckTokens;
        
    } catch (error) {
        console.error('Ошибка поиска зависших токенов:', error.message);
        return [];
    }
}

async function monitorTokenActivity(intervalSeconds = 30) {
    console.log(`Мониторинг активности токенов (интервал: ${intervalSeconds}с)`);
    
    let previousCounts = {};
    
    const monitor = async () => {
        try {
            const currentCounts = {};
            
            // Получаем счетчики по состояниям
            for (const state of ['ACTIVE', 'COMPLETED', 'CANCELLED', 'WAITING']) {
                try {
                    const tokensData = await listTokens({
                        filters: { state },
                        pagination: { pageSize: 1 } // Нужен только total_count
                    });
                    
                    currentCounts[state] = tokensData ? tokensData.totalCount : 0;
                } catch (error) {
                    currentCounts[state] = 0;
                }
            }
            
            // Отображение состояния
            const timestamp = new Date().toLocaleTimeString();
            console.log(`\n[${timestamp}] === Активность токенов ===`);
            
            const total = Object.values(currentCounts).reduce((sum, count) => sum + count, 0);
            console.log(`Всего токенов: ${total}`);
            
            Object.entries(currentCounts).forEach(([state, count]) => {
                const emoji = getStateEmoji(state);
                let change = '';
                
                if (previousCounts[state] !== undefined) {
                    const delta = count - previousCounts[state];
                    if (delta > 0) {
                        change = ` (+${delta})`;
                    } else if (delta < 0) {
                        change = ` (${delta})`;
                    }
                }
                
                console.log(`  ${emoji} ${state}: ${count}${change}`);
            });
            
            // Дополнительная информация об активных токенах
            if (currentCounts.ACTIVE > 0) {
                try {
                    const activeTokensData = await listTokens({
                        filters: { state: 'ACTIVE' },
                        pagination: { pageSize: 100 }
                    });
                    
                    if (activeTokensData && activeTokensData.tokens.length > 0) {
                        const processes = {};
                        activeTokensData.tokens.forEach(token => {
                            processes[token.processKey] = (processes[token.processKey] || 0) + 1;
                        });
                        
                        if (Object.keys(processes).length > 1) {
                            console.log(`  Активные процессы: ${Object.keys(processes).length}`);
                        }
                    }
                } catch (error) {
                    // Игнорируем ошибки получения дополнительной информации
                }
            }
            
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
    const arg = process.argv[3];
    
    switch (command) {
        case 'list':
            listTokens().then(displayTokens);
            break;
            
        case 'active':
            listTokens({ 
                filters: { state: 'ACTIVE' } 
            }).then(displayTokens);
            break;
            
        case 'analyze':
            if (!arg) {
                console.log('Требуется instance_id');
                process.exit(1);
            }
            analyzeProcessTokens(arg);
            break;
            
        case 'trace':
            if (!arg) {
                console.log('Требуется instance_id');
                process.exit(1);
            }
            traceProcessExecution(arg);
            break;
            
        case 'summary':
            getActiveTokensSummary();
            break;
            
        case 'stuck':
            const hours = parseInt(arg) || 24;
            findStuckTokens(hours);
            break;
            
        case 'monitor':
            const interval = parseInt(arg) || 30;
            monitorTokenActivity(interval);
            break;
            
        default:
            console.log('Использование:');
            console.log('  node tokens.js list                     - список токенов');
            console.log('  node tokens.js active                   - только активные');
            console.log('  node tokens.js analyze <instance_id>    - анализ токенов процесса');
            console.log('  node tokens.js trace <instance_id>      - трассировка выполнения');
            console.log('  node tokens.js summary                  - сводка по активным');
            console.log('  node tokens.js stuck [hours]            - поиск зависших токенов');
            console.log('  node tokens.js monitor [interval]       - мониторинг активности');
            break;
    }
}

module.exports = {
    listTokens,
    displayTokens,
    analyzeProcessTokens,
    traceProcessExecution,
    getActiveTokensSummary,
    findStuckTokens,
    monitorTokenActivity
};
```

## Диагностика и отладка

### Поиск проблемных токенов
```python
def diagnose_process_issues(instance_id):
    """Диагностика проблем в процессе"""
    tokens = get_all_tokens_for_process(instance_id)
    
    issues = []
    
    # Поиск зависших токенов
    stuck_tokens = [t for t in tokens if is_token_stuck(t)]
    if stuck_tokens:
        issues.append(f"Зависшие токены: {len(stuck_tokens)}")
    
    # Поиск дублирующихся токенов
    element_tokens = defaultdict(list)
    for token in tokens:
        if token['state'] == 'ACTIVE':
            element_tokens[token['current_element_id']].append(token)
    
    duplicate_elements = {k: v for k, v in element_tokens.items() if len(v) > 1}
    if duplicate_elements:
        issues.append(f"Дублирующиеся токены на элементах: {list(duplicate_elements.keys())}")
    
    return issues
```

### Performance анализ
```go
func analyzeTokenPerformance(client pb.ProcessServiceClient, ctx context.Context, processKey string) {
    // Анализ производительности токенов по типу процесса
    response, err := client.ListTokens(ctx, &pb.ListTokensRequest{
        // Фильтр по process_key здесь нет, нужно фильтровать вручную
        PageSize: 1000,
        SortBy:   "created_at",
    })
    
    // Фильтрация и анализ времени выполнения элементов
    elementDurations := make(map[string][]time.Duration)
    
    for _, token := range response.Tokens {
        if token.ProcessKey == processKey && token.State == "COMPLETED" {
            duration := time.Unix(token.UpdatedAt, 0).Sub(time.Unix(token.CreatedAt, 0))
            elementDurations[token.CurrentElementId] = append(elementDurations[token.CurrentElementId], duration)
        }
    }
    
    // Расчет статистики
    for elementId, durations := range elementDurations {
        avg := calculateAverage(durations)
        fmt.Printf("Элемент %s: среднее время %v (%d выполнений)\n", 
            elementId, avg, len(durations))
    }
}
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры фильтрации или пагинации
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
- [GetTokenStatus](get-token-status.md) - Статус конкретного токена
- [ListProcessInstances](list-process-instances.md) - Экземпляры процессов
- [GetProcessInstanceStatus](get-process-instance-status.md) - Статус экземпляра
