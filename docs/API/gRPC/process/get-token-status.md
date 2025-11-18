# GetTokenStatus

## Описание
Получает детальную информацию о конкретном токене выполнения, включая его состояние, переменные и историю изменений.

## Синтаксис
```protobuf
rpc GetTokenStatus(GetTokenStatusRequest) returns (GetTokenStatusResponse);
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

### GetTokenStatusRequest
```protobuf
message GetTokenStatusRequest {
  string token_id = 1;      // ID токена
}
```

#### Поля:
- **token_id** (string, required): Уникальный идентификатор токена

## Параметры ответа

### GetTokenStatusResponse
```protobuf
message GetTokenStatusResponse {
  TokenDetails token = 1;   // Детальная информация о токене
  bool success = 2;         // Статус успешности
  string message = 3;       // Сообщение о результате
}

message TokenDetails {
  string token_id = 1;                    // ID токена
  string process_instance_id = 2;         // ID экземпляра процесса
  string process_key = 3;                 // Ключ процесса
  string current_element_id = 4;          // ID текущего элемента BPMN
  string element_type = 5;                // Тип элемента BPMN
  string state = 6;                       // Состояние токена
  string waiting_for = 7;                 // Что ожидает токен
  int64 created_at = 8;                   // Время создания (Unix timestamp)
  int64 updated_at = 9;                   // Время обновления (Unix timestamp)
  map<string, string> variables = 10;     // Переменные токена
  repeated string execution_path = 11;     // Путь выполнения (история элементов)
  string parent_token_id = 12;            // ID родительского токена (для параллельных потоков)
  repeated string child_token_ids = 13;    // ID дочерних токенов
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
    
    tokenId := "srv1-tokenid12345"
    
    // Получение статуса токена
    response, err := client.GetTokenStatus(ctx, &pb.GetTokenStatusRequest{
        TokenId: tokenId,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        token := response.Token
        
        fmt.Printf("=== Статус токена %s ===\n", token.TokenId)
        fmt.Printf("Экземпляр процесса: %s\n", token.ProcessInstanceId)
        fmt.Printf("Процесс: %s\n", token.ProcessKey)
        fmt.Printf("Текущий элемент: %s (%s)\n", token.CurrentElementId, token.ElementType)
        fmt.Printf("Состояние: %s\n", token.State)
        
        if token.WaitingFor != "" {
            fmt.Printf("Ожидает: %s\n", token.WaitingFor)
        }
        
        fmt.Printf("Создан: %s\n", formatTimestamp(token.CreatedAt))
        fmt.Printf("Обновлен: %s\n", formatTimestamp(token.UpdatedAt))
        
        // Показать путь выполнения
        if len(token.ExecutionPath) > 0 {
            fmt.Printf("\nПуть выполнения (%d элементов):\n", len(token.ExecutionPath))
            for i, elementId := range token.ExecutionPath {
                fmt.Printf("  %d. %s\n", i+1, elementId)
            }
        }
        
        // Показать иерархию токенов
        if token.ParentTokenId != "" {
            fmt.Printf("\nРодительский токен: %s\n", token.ParentTokenId)
        }
        
        if len(token.ChildTokenIds) > 0 {
            fmt.Printf("Дочерние токены (%d):\n", len(token.ChildTokenIds))
            for i, childId := range token.ChildTokenIds {
                fmt.Printf("  %d. %s\n", i+1, childId)
            }
        }
        
        // Показать переменные
        if len(token.Variables) > 0 {
            fmt.Printf("\nПеременные токена (%d):\n", len(token.Variables))
            for key, value := range token.Variables {
                fmt.Printf("  %s: %s\n", key, value)
            }
        }
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

func formatTimestamp(timestamp int64) string {
    return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// Анализ производительности токена
func analyzeTokenPerformance(client pb.ProcessServiceClient, ctx context.Context, tokenId string) {
    response, err := client.GetTokenStatus(ctx, &pb.GetTokenStatusRequest{
        TokenId: tokenId,
    })
    
    if err != nil {
        log.Printf("Ошибка получения токена: %v", err)
        return
    }
    
    if !response.Success {
        log.Printf("Токен не найден: %s", response.Message)
        return
    }
    
    token := response.Token
    
    fmt.Printf("=== Анализ производительности токена %s ===\n", tokenId)
    
    // Вычисление времени жизни токена
    created := time.Unix(token.CreatedAt, 0)
    updated := time.Unix(token.UpdatedAt, 0)
    
    if token.State == "COMPLETED" || token.State == "CANCELLED" {
        lifetime := updated.Sub(created)
        fmt.Printf("Время жизни токена: %v\n", lifetime)
        
        if len(token.ExecutionPath) > 1 {
            avgTimePerElement := lifetime / time.Duration(len(token.ExecutionPath))
            fmt.Printf("Среднее время на элемент: %v\n", avgTimePerElement)
        }
    } else {
        currentAge := time.Since(created)
        fmt.Printf("Текущий возраст токена: %v\n", currentAge)
        
        if currentAge > 24*time.Hour {
            fmt.Printf("⚠️  Токен выполняется более 24 часов\n")
        }
    }
    
    // Анализ пути выполнения
    if len(token.ExecutionPath) > 0 {
        fmt.Printf("\nАнализ пути выполнения:\n")
        fmt.Printf("Пройдено элементов: %d\n", len(token.ExecutionPath))
        fmt.Printf("Текущий элемент: %s\n", token.CurrentElementId)
        
        // Проверка на циклы
        elementCounts := make(map[string]int)
        for _, elementId := range token.ExecutionPath {
            elementCounts[elementId]++
        }
        
        for elementId, count := range elementCounts {
            if count > 1 {
                fmt.Printf("🔄 Элемент %s выполнялся %d раз (возможный цикл)\n", elementId, count)
            }
        }
    }
    
    // Анализ переменных
    if len(token.Variables) > 0 {
        fmt.Printf("\nАнализ переменных:\n")
        fmt.Printf("Количество переменных: %d\n", len(token.Variables))
        
        // Проверка размера переменных
        totalSize := 0
        for key, value := range token.Variables {
            size := len(key) + len(value)
            totalSize += size
            
            if len(value) > 1000 {
                fmt.Printf("📊 Большая переменная %s: %d символов\n", key, len(value))
            }
        }
        
        fmt.Printf("Общий размер переменных: %d символов\n", totalSize)
        
        if totalSize > 10000 {
            fmt.Printf("⚠️  Большой объем переменных может влиять на производительность\n")
        }
    }
}

// Построение дерева токенов
func buildTokenTree(client pb.ProcessServiceClient, ctx context.Context, rootTokenId string) {
    fmt.Printf("=== Дерево токенов (корень: %s) ===\n", rootTokenId)
    
    visited := make(map[string]bool)
    printTokenTree(client, ctx, rootTokenId, 0, visited)
}

func printTokenTree(client pb.ProcessServiceClient, ctx context.Context, tokenId string, depth int, visited map[string]bool) {
    if visited[tokenId] {
        fmt.Printf("%s🔄 %s (уже посещен)\n", getIndent(depth), tokenId)
        return
    }
    
    visited[tokenId] = true
    
    response, err := client.GetTokenStatus(ctx, &pb.GetTokenStatusRequest{
        TokenId: tokenId,
    })
    
    if err != nil {
        fmt.Printf("%s❌ %s (ошибка: %v)\n", getIndent(depth), tokenId, err)
        return
    }
    
    if !response.Success {
        fmt.Printf("%s❌ %s (не найден)\n", getIndent(depth), tokenId)
        return
    }
    
    token := response.Token
    stateIcon := getStateIcon(token.State)
    
    fmt.Printf("%s%s %s (%s) - %s\n", 
        getIndent(depth), stateIcon, token.TokenId, token.State, token.CurrentElementId)
    
    // Рекурсивно обрабатываем дочерние токены
    for _, childId := range token.ChildTokenIds {
        printTokenTree(client, ctx, childId, depth+1, visited)
    }
}

func getIndent(depth int) string {
    indent := ""
    for i := 0; i < depth; i++ {
        indent += "  "
    }
    return indent
}

func getStateIcon(state string) string {
    switch state {
    case "ACTIVE":
        return "🟢"
    case "COMPLETED":
        return "✅"
    case "CANCELLED":
        return "⏹️"
    case "WAITING":
        return "⏳"
    default:
        return "❓"
    }
}

// Сравнение двух токенов
func compareTokens(client pb.ProcessServiceClient, ctx context.Context, tokenId1, tokenId2 string) {
    fmt.Printf("=== Сравнение токенов ===\n")
    
    // Получаем оба токена
    response1, err1 := client.GetTokenStatus(ctx, &pb.GetTokenStatusRequest{TokenId: tokenId1})
    response2, err2 := client.GetTokenStatus(ctx, &pb.GetTokenStatusRequest{TokenId: tokenId2})
    
    if err1 != nil || !response1.Success {
        fmt.Printf("Ошибка получения токена 1: %v\n", err1)
        return
    }
    
    if err2 != nil || !response2.Success {
        fmt.Printf("Ошибка получения токена 2: %v\n", err2)
        return
    }
    
    token1 := response1.Token
    token2 := response2.Token
    
    fmt.Printf("Токен 1: %s\n", token1.TokenId)
    fmt.Printf("Токен 2: %s\n", token2.TokenId)
    fmt.Println()
    
    // Сравнение основных параметров
    compareField("Процесс", token1.ProcessKey, token2.ProcessKey)
    compareField("Экземпляр", token1.ProcessInstanceId, token2.ProcessInstanceId)
    compareField("Текущий элемент", token1.CurrentElementId, token2.CurrentElementId)
    compareField("Состояние", token1.State, token2.State)
    compareField("Ожидает", token1.WaitingFor, token2.WaitingFor)
    
    // Сравнение времени
    created1 := time.Unix(token1.CreatedAt, 0)
    created2 := time.Unix(token2.CreatedAt, 0)
    timeDiff := created2.Sub(created1)
    
    fmt.Printf("Время создания:\n")
    fmt.Printf("  Токен 1: %s\n", created1.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Токен 2: %s\n", created2.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Разница: %v\n", timeDiff)
    
    // Сравнение переменных
    fmt.Printf("\nПеременные:\n")
    fmt.Printf("  Токен 1: %d переменных\n", len(token1.Variables))
    fmt.Printf("  Токен 2: %d переменных\n", len(token2.Variables))
    
    // Общие переменные
    commonVars := 0
    for key := range token1.Variables {
        if _, exists := token2.Variables[key]; exists {
            commonVars++
        }
    }
    fmt.Printf("  Общих переменных: %d\n", commonVars)
    
    // Сравнение путей выполнения
    fmt.Printf("\nПуть выполнения:\n")
    fmt.Printf("  Токен 1: %d элементов\n", len(token1.ExecutionPath))
    fmt.Printf("  Токен 2: %d элементов\n", len(token2.ExecutionPath))
    
    // Общие элементы в пути
    commonElements := 0
    elementMap1 := make(map[string]bool)
    for _, elementId := range token1.ExecutionPath {
        elementMap1[elementId] = true
    }
    
    for _, elementId := range token2.ExecutionPath {
        if elementMap1[elementId] {
            commonElements++
        }
    }
    fmt.Printf("  Общих элементов в пути: %d\n", commonElements)
}

func compareField(fieldName, value1, value2 string) {
    if value1 == value2 {
        fmt.Printf("%s: %s ✅\n", fieldName, value1)
    } else {
        fmt.Printf("%s:\n", fieldName)
        fmt.Printf("  Токен 1: %s\n", value1)
        fmt.Printf("  Токен 2: %s\n", value2)
        fmt.Printf("  ❌ Различаются\n")
    }
}
```

### Python
```python
import grpc
import json
from datetime import datetime, timedelta
from collections import Counter

import process_pb2
import process_pb2_grpc

def get_token_status(token_id):
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = process_pb2.GetTokenStatusRequest(
        token_id=token_id
    )
    
    try:
        response = stub.GetTokenStatus(request, metadata=metadata)
        
        if response.success:
            token = response.token
            return {
                'token_id': token.token_id,
                'process_instance_id': token.process_instance_id,
                'process_key': token.process_key,
                'current_element_id': token.current_element_id,
                'element_type': token.element_type,
                'state': token.state,
                'waiting_for': token.waiting_for,
                'created_at': token.created_at,
                'updated_at': token.updated_at,
                'variables': dict(token.variables),
                'execution_path': list(token.execution_path),
                'parent_token_id': token.parent_token_id,
                'child_token_ids': list(token.child_token_ids)
            }
        else:
            print(f"Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def display_token_status(token_id):
    """Отображение статуса токена в читаемом формате"""
    token = get_token_status(token_id)
    
    if not token:
        return
    
    print(f"=== Статус токена {token['token_id']} ===")
    print(f"Экземпляр процесса: {token['process_instance_id']}")
    print(f"Процесс: {token['process_key']}")
    print(f"Текущий элемент: {token['current_element_id']} ({token['element_type']})")
    print(f"Состояние: {get_state_emoji(token['state'])} {token['state']}")
    
    if token['waiting_for']:
        print(f"Ожидает: {token['waiting_for']}")
    
    print(f"Создан: {format_timestamp(token['created_at'])}")
    print(f"Обновлен: {format_timestamp(token['updated_at'])}")
    
    # Время жизни токена
    if token['state'] in ['COMPLETED', 'CANCELLED']:
        lifetime_seconds = token['updated_at'] - token['created_at']
        lifetime = format_duration(lifetime_seconds)
        print(f"Время жизни: {lifetime}")
    else:
        age_seconds = datetime.now().timestamp() - token['created_at']
        age = format_duration(age_seconds)
        print(f"Возраст: {age}")
    
    # Путь выполнения
    if token['execution_path']:
        print(f"\nПуть выполнения ({len(token['execution_path'])} элементов):")
        for i, element_id in enumerate(token['execution_path'], 1):
            current_marker = " ← текущий" if element_id == token['current_element_id'] else ""
            print(f"  {i}. {element_id}{current_marker}")
    
    # Иерархия токенов
    if token['parent_token_id']:
        print(f"\nРодительский токен: {token['parent_token_id']}")
    
    if token['child_token_ids']:
        print(f"Дочерние токены ({len(token['child_token_ids'])}):")
        for i, child_id in enumerate(token['child_token_ids'], 1):
            print(f"  {i}. {child_id}")
    
    # Переменные
    if token['variables']:
        print(f"\nПеременные токена ({len(token['variables'])}):")
        for key, value in token['variables'].items():
            # Попытка красиво отформатировать JSON
            try:
                parsed_value = json.loads(value)
                if isinstance(parsed_value, dict):
                    print(f"  {key}:")
                    print(f"    {json.dumps(parsed_value, indent=4, ensure_ascii=False)}")
                else:
                    print(f"  {key}: {value}")
            except:
                # Обрезаем длинные значения
                display_value = value if len(value) <= 100 else value[:97] + "..."
                print(f"  {key}: {display_value}")

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

def analyze_token_performance(token_id):
    """Анализ производительности токена"""
    token = get_token_status(token_id)
    
    if not token:
        return
    
    print(f"=== Анализ производительности токена {token_id} ===")
    
    # Анализ времени выполнения
    created = datetime.fromtimestamp(token['created_at'])
    updated = datetime.fromtimestamp(token['updated_at'])
    
    if token['state'] in ['COMPLETED', 'CANCELLED']:
        lifetime = updated - created
        print(f"Время жизни токена: {lifetime}")
        
        if token['execution_path'] and len(token['execution_path']) > 1:
            avg_time_per_element = lifetime / len(token['execution_path'])
            print(f"Среднее время на элемент: {avg_time_per_element}")
    else:
        current_age = datetime.now() - created
        print(f"Текущий возраст токена: {current_age}")
        
        if current_age > timedelta(hours=24):
            print("⚠️  Токен выполняется более 24 часов")
    
    # Анализ пути выполнения
    if token['execution_path']:
        print(f"\nАнализ пути выполнения:")
        print(f"Пройдено элементов: {len(token['execution_path'])}")
        print(f"Текущий элемент: {token['current_element_id']}")
        
        # Проверка на циклы
        element_counts = Counter(token['execution_path'])
        cycles = {element: count for element, count in element_counts.items() if count > 1}
        
        if cycles:
            print("🔄 Обнаружены возможные циклы:")
            for element_id, count in cycles.items():
                print(f"  {element_id}: выполнялся {count} раз")
        else:
            print("✅ Циклы не обнаружены")
    
    # Анализ переменных
    if token['variables']:
        print(f"\nАнализ переменных:")
        print(f"Количество переменных: {len(token['variables'])}")
        
        total_size = 0
        large_vars = []
        
        for key, value in token['variables'].items():
            size = len(key) + len(value)
            total_size += size
            
            if len(value) > 1000:
                large_vars.append((key, len(value)))
        
        print(f"Общий размер переменных: {total_size} символов")
        
        if large_vars:
            print("📊 Большие переменные:")
            for var_name, var_size in large_vars:
                print(f"  {var_name}: {var_size} символов")
        
        if total_size > 10000:
            print("⚠️  Большой объем переменных может влиять на производительность")

def build_token_tree(root_token_id):
    """Построение дерева токенов"""
    print(f"=== Дерево токенов (корень: {root_token_id}) ===")
    
    visited = set()
    print_token_tree(root_token_id, 0, visited)

def print_token_tree(token_id, depth, visited):
    """Рекурсивная печать дерева токенов"""
    if token_id in visited:
        print(f"{'  ' * depth}🔄 {token_id} (уже посещен)")
        return
    
    visited.add(token_id)
    
    token = get_token_status(token_id)
    
    if not token:
        print(f"{'  ' * depth}❌ {token_id} (не найден)")
        return
    
    state_icon = get_state_emoji(token['state'])
    print(f"{'  ' * depth}{state_icon} {token['token_id']} ({token['state']}) - {token['current_element_id']}")
    
    # Рекурсивно обрабатываем дочерние токены
    for child_id in token['child_token_ids']:
        print_token_tree(child_id, depth + 1, visited)

def compare_tokens(token_id1, token_id2):
    """Сравнение двух токенов"""
    print("=== Сравнение токенов ===")
    
    token1 = get_token_status(token_id1)
    token2 = get_token_status(token_id2)
    
    if not token1:
        print(f"Токен 1 ({token_id1}) не найден")
        return
    
    if not token2:
        print(f"Токен 2 ({token_id2}) не найден")
        return
    
    print(f"Токен 1: {token1['token_id']}")
    print(f"Токен 2: {token2['token_id']}")
    print()
    
    # Сравнение основных параметров
    def compare_field(field_name, value1, value2):
        if value1 == value2:
            print(f"{field_name}: {value1} ✅")
        else:
            print(f"{field_name}:")
            print(f"  Токен 1: {value1}")
            print(f"  Токен 2: {value2}")
            print("  ❌ Различаются")
    
    compare_field("Процесс", token1['process_key'], token2['process_key'])
    compare_field("Экземпляр", token1['process_instance_id'], token2['process_instance_id'])
    compare_field("Текущий элемент", token1['current_element_id'], token2['current_element_id'])
    compare_field("Состояние", token1['state'], token2['state'])
    compare_field("Ожидает", token1['waiting_for'], token2['waiting_for'])
    
    # Сравнение времени
    created1 = datetime.fromtimestamp(token1['created_at'])
    created2 = datetime.fromtimestamp(token2['created_at'])
    time_diff = created2 - created1
    
    print(f"\nВремя создания:")
    print(f"  Токен 1: {created1.strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"  Токен 2: {created2.strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"  Разница: {time_diff}")
    
    # Сравнение переменных
    print(f"\nПеременные:")
    print(f"  Токен 1: {len(token1['variables'])} переменных")
    print(f"  Токен 2: {len(token2['variables'])} переменных")
    
    common_vars = set(token1['variables'].keys()) & set(token2['variables'].keys())
    print(f"  Общих переменных: {len(common_vars)}")
    
    # Сравнение путей выполнения
    print(f"\nПуть выполнения:")
    print(f"  Токен 1: {len(token1['execution_path'])} элементов")
    print(f"  Токен 2: {len(token2['execution_path'])} элементов")
    
    common_elements = set(token1['execution_path']) & set(token2['execution_path'])
    print(f"  Общих элементов в пути: {len(common_elements)}")

def find_related_tokens(token_id):
    """Поиск связанных токенов"""
    token = get_token_status(token_id)
    
    if not token:
        return
    
    print(f"=== Связанные токены для {token_id} ===")
    
    related_tokens = set()
    
    # Добавляем родительский токен
    if token['parent_token_id']:
        related_tokens.add(token['parent_token_id'])
        print(f"Родительский токен: {token['parent_token_id']}")
    
    # Добавляем дочерние токены
    if token['child_token_ids']:
        related_tokens.update(token['child_token_ids'])
        print(f"Дочерние токены: {', '.join(token['child_token_ids'])}")
    
    # Добавляем родственные токены (дети того же родителя)
    if token['parent_token_id']:
        parent_token = get_token_status(token['parent_token_id'])
        if parent_token and parent_token['child_token_ids']:
            siblings = [tid for tid in parent_token['child_token_ids'] if tid != token_id]
            if siblings:
                related_tokens.update(siblings)
                print(f"Родственные токены: {', '.join(siblings)}")
    
    # Получаем токены того же экземпляра процесса
    # Здесь нужно было бы использовать ListTokens с фильтром по instance_id
    # Но для простоты покажем только прямые связи
    
    print(f"\nВсего связанных токенов: {len(related_tokens)}")
    
    return list(related_tokens)

def export_token_details(token_id, filename=None):
    """Экспорт деталей токена в JSON"""
    token = get_token_status(token_id)
    
    if not token:
        return
    
    if not filename:
        filename = f"token_{token_id}_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    
    # Добавляем метаданные экспорта
    export_data = {
        'export_metadata': {
            'exported_at': datetime.now().isoformat(),
            'token_id': token_id,
            'export_version': '1.0'
        },
        'token_details': token,
        'analysis': {
            'age_seconds': datetime.now().timestamp() - token['created_at'],
            'is_completed': token['state'] in ['COMPLETED', 'CANCELLED'],
            'has_cycles': len(token['execution_path']) != len(set(token['execution_path'])),
            'variables_count': len(token['variables']),
            'path_length': len(token['execution_path'])
        }
    }
    
    try:
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(export_data, f, indent=2, ensure_ascii=False)
        
        print(f"Детали токена экспортированы в: {filename}")
        return filename
        
    except Exception as e:
        print(f"Ошибка экспорта: {e}")
        return None

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python token_status.py <token_id>")
        print("  python token_status.py analyze <token_id>")
        print("  python token_status.py tree <token_id>")
        print("  python token_status.py compare <token_id1> <token_id2>")
        print("  python token_status.py export <token_id>")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "analyze":
        if len(sys.argv) < 3:
            print("Требуется token_id")
            sys.exit(1)
        analyze_token_performance(sys.argv[2])
    elif command == "tree":
        if len(sys.argv) < 3:
            print("Требуется token_id")
            sys.exit(1)
        build_token_tree(sys.argv[2])
    elif command == "compare":
        if len(sys.argv) < 4:
            print("Требуется два token_id")
            sys.exit(1)
        compare_tokens(sys.argv[2], sys.argv[3])
    elif command == "export":
        if len(sys.argv) < 3:
            print("Требуется token_id")
            sys.exit(1)
        export_token_details(sys.argv[2])
    else:
        # Простое отображение статуса
        token_id = command
        display_token_status(token_id)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const fs = require('fs').promises;

const PROTO_PATH = 'process.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

async function getTokenStatus(tokenId) {
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { token_id: tokenId };
        
        client.getTokenStatus(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            const token = response.token;
            resolve({
                tokenId: token.token_id,
                processInstanceId: token.process_instance_id,
                processKey: token.process_key,
                currentElementId: token.current_element_id,
                elementType: token.element_type,
                state: token.state,
                waitingFor: token.waiting_for,
                createdAt: Number(token.created_at) * 1000, // Convert to JS timestamp
                updatedAt: Number(token.updated_at) * 1000,
                variables: token.variables,
                executionPath: token.execution_path,
                parentTokenId: token.parent_token_id,
                childTokenIds: token.child_token_ids
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

async function displayTokenStatus(tokenId) {
    try {
        const token = await getTokenStatus(tokenId);
        
        console.log(`=== Статус токена ${token.tokenId} ===`);
        console.log(`Экземпляр процесса: ${token.processInstanceId}`);
        console.log(`Процесс: ${token.processKey}`);
        console.log(`Текущий элемент: ${token.currentElementId} (${token.elementType})`);
        console.log(`Состояние: ${getStateEmoji(token.state)} ${token.state}`);
        
        if (token.waitingFor) {
            console.log(`Ожидает: ${token.waitingFor}`);
        }
        
        console.log(`Создан: ${formatTimestamp(token.createdAt)}`);
        console.log(`Обновлен: ${formatTimestamp(token.updatedAt)}`);
        
        // Время жизни токена
        if (['COMPLETED', 'CANCELLED'].includes(token.state)) {
            const lifetime = token.updatedAt - token.createdAt;
            console.log(`Время жизни: ${formatDuration(lifetime)}`);
        } else {
            const age = Date.now() - token.createdAt;
            console.log(`Возраст: ${formatDuration(age)}`);
        }
        
        // Путь выполнения
        if (token.executionPath && token.executionPath.length > 0) {
            console.log(`\nПуть выполнения (${token.executionPath.length} элементов):`);
            token.executionPath.forEach((elementId, index) => {
                const currentMarker = elementId === token.currentElementId ? ' ← текущий' : '';
                console.log(`  ${index + 1}. ${elementId}${currentMarker}`);
            });
        }
        
        // Иерархия токенов
        if (token.parentTokenId) {
            console.log(`\nРодительский токен: ${token.parentTokenId}`);
        }
        
        if (token.childTokenIds && token.childTokenIds.length > 0) {
            console.log(`Дочерние токены (${token.childTokenIds.length}):`);
            token.childTokenIds.forEach((childId, index) => {
                console.log(`  ${index + 1}. ${childId}`);
            });
        }
        
        // Переменные
        if (token.variables && Object.keys(token.variables).length > 0) {
            console.log(`\nПеременные токена (${Object.keys(token.variables).length}):`);
            Object.entries(token.variables).forEach(([key, value]) => {
                try {
                    const parsed = JSON.parse(value);
                    if (typeof parsed === 'object' && parsed !== null) {
                        console.log(`  ${key}:`);
                        console.log(`    ${JSON.stringify(parsed, null, 4)}`);
                    } else {
                        console.log(`  ${key}: ${value}`);
                    }
                } catch {
                    // Обрезаем длинные значения
                    const displayValue = value.length <= 100 ? value : value.substring(0, 97) + '...';
                    console.log(`  ${key}: ${displayValue}`);
                }
            });
        }
        
        return token;
        
    } catch (error) {
        console.error(`Ошибка получения статуса токена: ${error.message}`);
        return null;
    }
}

async function analyzeTokenPerformance(tokenId) {
    try {
        const token = await getTokenStatus(tokenId);
        
        console.log(`=== Анализ производительности токена ${tokenId} ===`);
        
        // Анализ времени выполнения
        const created = new Date(token.createdAt);
        const updated = new Date(token.updatedAt);
        
        if (['COMPLETED', 'CANCELLED'].includes(token.state)) {
            const lifetime = updated.getTime() - created.getTime();
            console.log(`Время жизни токена: ${formatDuration(lifetime)}`);
            
            if (token.executionPath && token.executionPath.length > 1) {
                const avgTimePerElement = lifetime / token.executionPath.length;
                console.log(`Среднее время на элемент: ${formatDuration(avgTimePerElement)}`);
            }
        } else {
            const currentAge = Date.now() - token.createdAt;
            console.log(`Текущий возраст токена: ${formatDuration(currentAge)}`);
            
            if (currentAge > 24 * 60 * 60 * 1000) { // 24 hours
                console.log('⚠️  Токен выполняется более 24 часов');
            }
        }
        
        // Анализ пути выполнения
        if (token.executionPath && token.executionPath.length > 0) {
            console.log('\nАнализ пути выполнения:');
            console.log(`Пройдено элементов: ${token.executionPath.length}`);
            console.log(`Текущий элемент: ${token.currentElementId}`);
            
            // Проверка на циклы
            const elementCounts = {};
            token.executionPath.forEach(elementId => {
                elementCounts[elementId] = (elementCounts[elementId] || 0) + 1;
            });
            
            const cycles = Object.entries(elementCounts).filter(([, count]) => count > 1);
            
            if (cycles.length > 0) {
                console.log('🔄 Обнаружены возможные циклы:');
                cycles.forEach(([elementId, count]) => {
                    console.log(`  ${elementId}: выполнялся ${count} раз`);
                });
            } else {
                console.log('✅ Циклы не обнаружены');
            }
        }
        
        // Анализ переменных
        if (token.variables && Object.keys(token.variables).length > 0) {
            console.log('\nАнализ переменных:');
            console.log(`Количество переменных: ${Object.keys(token.variables).length}`);
            
            let totalSize = 0;
            const largeVars = [];
            
            Object.entries(token.variables).forEach(([key, value]) => {
                const size = key.length + value.length;
                totalSize += size;
                
                if (value.length > 1000) {
                    largeVars.push({ name: key, size: value.length });
                }
            });
            
            console.log(`Общий размер переменных: ${totalSize} символов`);
            
            if (largeVars.length > 0) {
                console.log('📊 Большие переменные:');
                largeVars.forEach(({ name, size }) => {
                    console.log(`  ${name}: ${size} символов`);
                });
            }
            
            if (totalSize > 10000) {
                console.log('⚠️  Большой объем переменных может влиять на производительность');
            }
        }
        
    } catch (error) {
        console.error(`Ошибка анализа производительности: ${error.message}`);
    }
}

async function buildTokenTree(rootTokenId) {
    console.log(`=== Дерево токенов (корень: ${rootTokenId}) ===`);
    
    const visited = new Set();
    await printTokenTree(rootTokenId, 0, visited);
}

async function printTokenTree(tokenId, depth, visited) {
    if (visited.has(tokenId)) {
        console.log(`${'  '.repeat(depth)}🔄 ${tokenId} (уже посещен)`);
        return;
    }
    
    visited.add(tokenId);
    
    try {
        const token = await getTokenStatus(tokenId);
        
        const stateIcon = getStateEmoji(token.state);
        console.log(`${'  '.repeat(depth)}${stateIcon} ${token.tokenId} (${token.state}) - ${token.currentElementId}`);
        
        // Рекурсивно обрабатываем дочерние токены
        if (token.childTokenIds && token.childTokenIds.length > 0) {
            for (const childId of token.childTokenIds) {
                await printTokenTree(childId, depth + 1, visited);
            }
        }
        
    } catch (error) {
        console.log(`${'  '.repeat(depth)}❌ ${tokenId} (ошибка: ${error.message})`);
    }
}

async function compareTokens(tokenId1, tokenId2) {
    console.log('=== Сравнение токенов ===');
    
    try {
        const [token1, token2] = await Promise.all([
            getTokenStatus(tokenId1),
            getTokenStatus(tokenId2)
        ]);
        
        console.log(`Токен 1: ${token1.tokenId}`);
        console.log(`Токен 2: ${token2.tokenId}`);
        console.log();
        
        // Функция сравнения полей
        const compareField = (fieldName, value1, value2) => {
            if (value1 === value2) {
                console.log(`${fieldName}: ${value1} ✅`);
            } else {
                console.log(`${fieldName}:`);
                console.log(`  Токен 1: ${value1}`);
                console.log(`  Токен 2: ${value2}`);
                console.log('  ❌ Различаются');
            }
        };
        
        compareField('Процесс', token1.processKey, token2.processKey);
        compareField('Экземпляр', token1.processInstanceId, token2.processInstanceId);
        compareField('Текущий элемент', token1.currentElementId, token2.currentElementId);
        compareField('Состояние', token1.state, token2.state);
        compareField('Ожидает', token1.waitingFor, token2.waitingFor);
        
        // Сравнение времени
        const created1 = new Date(token1.createdAt);
        const created2 = new Date(token2.createdAt);
        const timeDiff = created2.getTime() - created1.getTime();
        
        console.log('\nВремя создания:');
        console.log(`  Токен 1: ${created1.toLocaleString()}`);
        console.log(`  Токен 2: ${created2.toLocaleString()}`);
        console.log(`  Разница: ${formatDuration(Math.abs(timeDiff))}`);
        
        // Сравнение переменных
        const vars1 = Object.keys(token1.variables || {});
        const vars2 = Object.keys(token2.variables || {});
        const commonVars = vars1.filter(key => vars2.includes(key));
        
        console.log('\nПеременные:');
        console.log(`  Токен 1: ${vars1.length} переменных`);
        console.log(`  Токен 2: ${vars2.length} переменных`);
        console.log(`  Общих переменных: ${commonVars.length}`);
        
        // Сравнение путей выполнения
        const path1 = token1.executionPath || [];
        const path2 = token2.executionPath || [];
        const commonElements = path1.filter(element => path2.includes(element));
        
        console.log('\nПуть выполнения:');
        console.log(`  Токен 1: ${path1.length} элементов`);
        console.log(`  Токен 2: ${path2.length} элементов`);
        console.log(`  Общих элементов в пути: ${commonElements.length}`);
        
    } catch (error) {
        console.error(`Ошибка сравнения токенов: ${error.message}`);
    }
}

async function exportTokenDetails(tokenId, filename) {
    try {
        const token = await getTokenStatus(tokenId);
        
        if (!filename) {
            const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
            filename = `token_${tokenId}_${timestamp}.json`;
        }
        
        // Добавляем метаданные экспорта
        const exportData = {
            export_metadata: {
                exported_at: new Date().toISOString(),
                token_id: tokenId,
                export_version: '1.0'
            },
            token_details: token,
            analysis: {
                age_milliseconds: Date.now() - token.createdAt,
                is_completed: ['COMPLETED', 'CANCELLED'].includes(token.state),
                has_cycles: token.executionPath ? 
                    token.executionPath.length !== new Set(token.executionPath).size : false,
                variables_count: Object.keys(token.variables || {}).length,
                path_length: token.executionPath ? token.executionPath.length : 0
            }
        };
        
        await fs.writeFile(filename, JSON.stringify(exportData, null, 2));
        console.log(`Детали токена экспортированы в: ${filename}`);
        
        return filename;
        
    } catch (error) {
        console.error(`Ошибка экспорта: ${error.message}`);
        return null;
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node token-status.js <token_id>                    - показать статус');
        console.log('  node token-status.js analyze <token_id>            - анализ производительности');
        console.log('  node token-status.js tree <token_id>               - дерево токенов');
        console.log('  node token-status.js compare <token_id1> <token_id2> - сравнение токенов');
        console.log('  node token-status.js export <token_id> [filename]  - экспорт в JSON');
        process.exit(1);
    }
    
    const command = args[0];
    
    (async () => {
        try {
            switch (command) {
                case 'analyze':
                    if (args.length < 2) {
                        console.log('Требуется token_id');
                        process.exit(1);
                    }
                    await analyzeTokenPerformance(args[1]);
                    break;
                    
                case 'tree':
                    if (args.length < 2) {
                        console.log('Требуется token_id');
                        process.exit(1);
                    }
                    await buildTokenTree(args[1]);
                    break;
                    
                case 'compare':
                    if (args.length < 3) {
                        console.log('Требуется два token_id');
                        process.exit(1);
                    }
                    await compareTokens(args[1], args[2]);
                    break;
                    
                case 'export':
                    if (args.length < 2) {
                        console.log('Требуется token_id');
                        process.exit(1);
                    }
                    await exportTokenDetails(args[1], args[2]);
                    break;
                    
                default:
                    // Простое отображение статуса
                    await displayTokenStatus(command);
                    break;
            }
        } catch (error) {
            console.error('Ошибка:', error.message);
        }
    })();
}

module.exports = {
    getTokenStatus,
    displayTokenStatus,
    analyzeTokenPerformance,
    buildTokenTree,
    compareTokens,
    exportTokenDetails
};
```

## Диагностика проблем

### Обнаружение зависших токенов
```go
func detectStuckTokens(client pb.ProcessServiceClient, ctx context.Context) {
    // Получаем все активные токены старше 24 часов
    cutoffTime := time.Now().Add(-24 * time.Hour).Unix()
    
    // Здесь нужно было бы использовать ListTokens с фильтром по времени
    // Но для демонстрации покажем логику
}
```

### Анализ производительности
```python
def analyze_element_performance(token_id):
    """Анализ производительности выполнения элементов"""
    token = get_token_status(token_id)
    
    if not token or not token['execution_path']:
        return
    
    # Анализ времени на каждый элемент (требует дополнительных данных)
    print("Анализ требует расширенной информации о времени выполнения элементов")
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный token_id
- `NOT_FOUND` (5): Токен не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Token 'invalid-token-id' not found"
}
```

## Связанные методы
- [ListTokens](list-tokens.md) - Список токенов для фильтрации
- [GetProcessInstanceStatus](get-process-instance-status.md) - Статус экземпляра процесса
- [ListProcessInstances](list-process-instances.md) - Экземпляры процессов
