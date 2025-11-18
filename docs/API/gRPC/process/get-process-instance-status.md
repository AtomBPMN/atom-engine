# GetProcessInstanceStatus

## Описание
Получает текущий статус экземпляра процесса, включая состояние выполнения, активные токены и метрики.

## Синтаксис
```protobuf
rpc GetProcessInstanceStatus(GetProcessInstanceStatusRequest) returns (GetProcessInstanceStatusResponse);
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

### GetProcessInstanceStatusRequest
```protobuf
message GetProcessInstanceStatusRequest {
  string instance_id = 1;       // ID экземпляра процесса
}
```

#### Поля:
- **instance_id** (string, required): Уникальный идентификатор экземпляра процесса

## Параметры ответа

### GetProcessInstanceStatusResponse
```protobuf
message GetProcessInstanceStatusResponse {
  string instance_id = 1;       // ID экземпляра
  string process_id = 2;        // ID процесса
  string status = 3;            // Статус экземпляра
  string started_at = 4;        // Время запуска
  string completed_at = 5;      // Время завершения (если завершен)
  int32 active_tokens = 6;      // Количество активных токенов
  int32 completed_tokens = 7;   // Количество завершенных токенов
  map<string, string> variables = 8; // Текущие переменные процесса
  bool success = 9;             // Статус успешности запроса
  string message = 10;          // Сообщение о результате
}
```

#### Статусы экземпляра:
- **ACTIVE** - Процесс выполняется, есть активные токены
- **COMPLETED** - Процесс успешно завершен 
- **FAILED** - Процесс завершен с ошибкой
- **CANCELLED** - Процесс отменен пользователем
- **SUSPENDED** - Процесс приостановлен

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
    
    instanceId := "srv1-aB3dEf9hK2mN5pQ8uV"
    
    // Получение статуса экземпляра
    response, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
        InstanceId: instanceId,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("=== Статус экземпляра %s ===\n", response.InstanceId)
        fmt.Printf("Процесс: %s\n", response.ProcessId)
        fmt.Printf("Статус: %s\n", response.Status)
        fmt.Printf("Запущен: %s\n", response.StartedAt)
        
        if response.CompletedAt != "" {
            fmt.Printf("Завершен: %s\n", response.CompletedAt)
            
            // Вычисление времени выполнения
            startTime, _ := time.Parse(time.RFC3339, response.StartedAt)
            endTime, _ := time.Parse(time.RFC3339, response.CompletedAt)
            duration := endTime.Sub(startTime)
            fmt.Printf("Длительность: %v\n", duration)
        }
        
        fmt.Printf("Активных токенов: %d\n", response.ActiveTokens)
        fmt.Printf("Завершенных токенов: %d\n", response.CompletedTokens)
        
        // Вывод переменных
        if len(response.Variables) > 0 {
            fmt.Println("\nПеременные процесса:")
            for key, value := range response.Variables {
                fmt.Printf("  %s: %s\n", key, value)
            }
        }
    } else {
        fmt.Printf("Ошибка: %s\n", response.Message)
    }
}

// Ожидание завершения процесса
func waitForCompletion(client pb.ProcessServiceClient, ctx context.Context, instanceId string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    fmt.Printf("Ожидание завершения процесса %s (таймаут: %v)\n", instanceId, timeout)
    
    for {
        select {
        case <-ticker.C:
            response, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
                InstanceId: instanceId,
            })
            
            if err != nil {
                return fmt.Errorf("ошибка проверки статуса: %v", err)
            }
            
            if !response.Success {
                return fmt.Errorf("не удалось получить статус: %s", response.Message)
            }
            
            fmt.Printf("[%s] Статус: %s, Активных токенов: %d\n", 
                time.Now().Format("15:04:05"), response.Status, response.ActiveTokens)
            
            switch response.Status {
            case "COMPLETED":
                fmt.Println("✅ Процесс успешно завершен")
                return nil
            case "FAILED":
                return fmt.Errorf("❌ процесс завершен с ошибкой")
            case "CANCELLED":
                return fmt.Errorf("⏹️ процесс отменен")
            }
            
        case <-time.After(time.Until(deadline)):
            return fmt.Errorf("⏰ таймаут ожидания завершения процесса")
        }
    }
}

// Мониторинг нескольких процессов
func monitorMultipleProcesses(client pb.ProcessServiceClient, ctx context.Context, instanceIds []string) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    fmt.Printf("Мониторинг %d процессов...\n", len(instanceIds))
    
    for {
        select {
        case <-ticker.C:
            fmt.Printf("\n[%s] === Статус процессов ===\n", time.Now().Format("15:04:05"))
            
            activeCount := 0
            for _, instanceId := range instanceIds {
                response, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
                    InstanceId: instanceId,
                })
                
                if err != nil {
                    fmt.Printf("❌ %s: ошибка - %v\n", instanceId, err)
                    continue
                }
                
                if !response.Success {
                    fmt.Printf("❌ %s: %s\n", instanceId, response.Message)
                    continue
                }
                
                statusIcon := getStatusIcon(response.Status)
                fmt.Printf("%s %s: %s (%d активных токенов)\n", 
                    statusIcon, instanceId, response.Status, response.ActiveTokens)
                
                if response.Status == "ACTIVE" {
                    activeCount++
                }
            }
            
            if activeCount == 0 {
                fmt.Println("Все процессы завершены")
                return
            }
            
        case <-ctx.Done():
            fmt.Println("Мониторинг остановлен")
            return
        }
    }
}

func getStatusIcon(status string) string {
    switch status {
    case "ACTIVE":
        return "🟢"
    case "COMPLETED":
        return "✅"
    case "FAILED":
        return "❌"
    case "CANCELLED":
        return "⏹️"
    case "SUSPENDED":
        return "⏸️"
    default:
        return "❓"
    }
}
```

### Python
```python
import grpc
import time
import json
from datetime import datetime, timedelta
from concurrent.futures import ThreadPoolExecutor

import process_pb2
import process_pb2_grpc

def get_process_instance_status(instance_id):
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = process_pb2.GetProcessInstanceStatusRequest(
        instance_id=instance_id
    )
    
    try:
        response = stub.GetProcessInstanceStatus(request, metadata=metadata)
        
        if response.success:
            return {
                'instance_id': response.instance_id,
                'process_id': response.process_id,
                'status': response.status,
                'started_at': response.started_at,
                'completed_at': response.completed_at,
                'active_tokens': response.active_tokens,
                'completed_tokens': response.completed_tokens,
                'variables': dict(response.variables)
            }
        else:
            print(f"Ошибка: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def display_process_status(instance_id):
    """Отображение статуса процесса в читаемом формате"""
    status = get_process_instance_status(instance_id)
    
    if not status:
        return
    
    print(f"=== Статус экземпляра {status['instance_id']} ===")
    print(f"Процесс: {status['process_id']}")
    print(f"Статус: {get_status_emoji(status['status'])} {status['status']}")
    print(f"Запущен: {format_timestamp(status['started_at'])}")
    
    if status['completed_at']:
        print(f"Завершен: {format_timestamp(status['completed_at'])}")
        
        # Вычисление длительности
        start_time = datetime.fromisoformat(status['started_at'].replace('Z', '+00:00'))
        end_time = datetime.fromisoformat(status['completed_at'].replace('Z', '+00:00'))
        duration = end_time - start_time
        print(f"Длительность: {format_duration(duration)}")
    
    print(f"Активных токенов: {status['active_tokens']}")
    print(f"Завершенных токенов: {status['completed_tokens']}")
    
    # Переменные
    if status['variables']:
        print("\nПеременные процесса:")
        for key, value in status['variables'].items():
            # Попытка парсить JSON для красивого вывода
            try:
                parsed_value = json.loads(value)
                if isinstance(parsed_value, dict):
                    print(f"  {key}: {json.dumps(parsed_value, indent=4, ensure_ascii=False)}")
                else:
                    print(f"  {key}: {value}")
            except:
                print(f"  {key}: {value}")

def get_status_emoji(status):
    """Emoji для статуса процесса"""
    emoji_map = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'FAILED': '❌',
        'CANCELLED': '⏹️',
        'SUSPENDED': '⏸️'
    }
    return emoji_map.get(status, '❓')

def format_timestamp(timestamp_str):
    """Форматирование времени"""
    try:
        dt = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
        return dt.strftime('%Y-%m-%d %H:%M:%S UTC')
    except:
        return timestamp_str

def format_duration(duration):
    """Форматирование длительности"""
    total_seconds = int(duration.total_seconds())
    hours, remainder = divmod(total_seconds, 3600)
    minutes, seconds = divmod(remainder, 60)
    
    if hours > 0:
        return f"{hours}ч {minutes}м {seconds}с"
    elif minutes > 0:
        return f"{minutes}м {seconds}с"
    else:
        return f"{seconds}с"

def wait_for_completion(instance_id, timeout_minutes=30, check_interval=5):
    """Ожидание завершения процесса"""
    print(f"Ожидание завершения процесса {instance_id} (таймаут: {timeout_minutes} мин)")
    
    start_time = time.time()
    timeout_seconds = timeout_minutes * 60
    
    while time.time() - start_time < timeout_seconds:
        status = get_process_instance_status(instance_id)
        
        if not status:
            time.sleep(check_interval)
            continue
        
        current_time = datetime.now().strftime('%H:%M:%S')
        print(f"[{current_time}] Статус: {status['status']}, Активных токенов: {status['active_tokens']}")
        
        if status['status'] == 'COMPLETED':
            print("✅ Процесс успешно завершен")
            return True
        elif status['status'] in ['FAILED', 'CANCELLED']:
            print(f"❌ Процесс завершен со статусом: {status['status']}")
            return False
        
        time.sleep(check_interval)
    
    print("⏰ Таймаут ожидания завершения процесса")
    return False

def monitor_processes(instance_ids, interval=10):
    """Мониторинг нескольких процессов"""
    print(f"Мониторинг {len(instance_ids)} процессов с интервалом {interval}с")
    
    try:
        while True:
            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] === Статус процессов ===")
            
            active_count = 0
            statuses = []
            
            # Параллельное получение статусов
            with ThreadPoolExecutor(max_workers=10) as executor:
                futures = {executor.submit(get_process_instance_status, iid): iid 
                          for iid in instance_ids}
                
                for future in futures:
                    instance_id = futures[future]
                    try:
                        status = future.result(timeout=5)
                        if status:
                            statuses.append(status)
                            if status['status'] == 'ACTIVE':
                                active_count += 1
                    except Exception as e:
                        print(f"❌ {instance_id}: ошибка - {e}")
            
            # Отображение статусов
            for status in sorted(statuses, key=lambda x: x['instance_id']):
                emoji = get_status_emoji(status['status'])
                print(f"{emoji} {status['instance_id']}: {status['status']} "
                      f"({status['active_tokens']} активных токенов)")
            
            if active_count == 0:
                print("Все процессы завершены")
                break
            
            print(f"\nАктивных процессов: {active_count}/{len(instance_ids)}")
            time.sleep(interval)
            
    except KeyboardInterrupt:
        print("\nМониторинг остановлен пользователем")

def get_process_summary(instance_ids):
    """Сводка по процессам"""
    print(f"=== Сводка по {len(instance_ids)} процессам ===")
    
    statuses = []
    with ThreadPoolExecutor(max_workers=10) as executor:
        futures = {executor.submit(get_process_instance_status, iid): iid 
                  for iid in instance_ids}
        
        for future in futures:
            try:
                status = future.result(timeout=5)
                if status:
                    statuses.append(status)
            except Exception:
                pass
    
    # Статистика по статусам
    status_counts = {}
    total_active_tokens = 0
    total_completed_tokens = 0
    
    for status in statuses:
        status_name = status['status']
        status_counts[status_name] = status_counts.get(status_name, 0) + 1
        total_active_tokens += status['active_tokens']
        total_completed_tokens += status['completed_tokens']
    
    print("Статистика по статусам:")
    for status_name, count in sorted(status_counts.items()):
        emoji = get_status_emoji(status_name)
        print(f"  {emoji} {status_name}: {count}")
    
    print(f"\nОбщая статистика:")
    print(f"  Всего активных токенов: {total_active_tokens}")
    print(f"  Всего завершенных токенов: {total_completed_tokens}")
    print(f"  Загруженных статусов: {len(statuses)}/{len(instance_ids)}")

if __name__ == "__main__":
    # Пример использования
    instance_id = "srv1-aB3dEf9hK2mN5pQ8uV"
    
    # Простое получение статуса
    display_process_status(instance_id)
    
    # Ожидание завершения
    # wait_for_completion(instance_id, timeout_minutes=10)
    
    # Мониторинг нескольких процессов
    # instance_ids = ["srv1-aB3dEf9hK2mN5pQ8uV", "srv1-xY2zW8vA5rT3nM9p"]
    # monitor_processes(instance_ids)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'process.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

async function getProcessInstanceStatus(instanceId) {
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { instance_id: instanceId };
        
        client.getProcessInstanceStatus(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (!response.success) {
                reject(new Error(response.message));
                return;
            }
            
            resolve({
                instanceId: response.instance_id,
                processId: response.process_id,
                status: response.status,
                startedAt: response.started_at,
                completedAt: response.completed_at,
                activeTokens: response.active_tokens,
                completedTokens: response.completed_tokens,
                variables: response.variables
            });
        });
    });
}

function getStatusEmoji(status) {
    const emojiMap = {
        'ACTIVE': '🟢',
        'COMPLETED': '✅',
        'FAILED': '❌',
        'CANCELLED': '⏹️',
        'SUSPENDED': '⏸️'
    };
    return emojiMap[status] || '❓';
}

function formatDuration(startTime, endTime) {
    const start = new Date(startTime);
    const end = new Date(endTime);
    const diffMs = end - start;
    
    const hours = Math.floor(diffMs / (1000 * 60 * 60));
    const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((diffMs % (1000 * 60)) / 1000);
    
    if (hours > 0) {
        return `${hours}ч ${minutes}м ${seconds}с`;
    } else if (minutes > 0) {
        return `${minutes}м ${seconds}с`;
    } else {
        return `${seconds}с`;
    }
}

async function displayProcessStatus(instanceId) {
    try {
        const status = await getProcessInstanceStatus(instanceId);
        
        console.log(`=== Статус экземпляра ${status.instanceId} ===`);
        console.log(`Процесс: ${status.processId}`);
        console.log(`Статус: ${getStatusEmoji(status.status)} ${status.status}`);
        console.log(`Запущен: ${new Date(status.startedAt).toLocaleString()}`);
        
        if (status.completedAt) {
            console.log(`Завершен: ${new Date(status.completedAt).toLocaleString()}`);
            console.log(`Длительность: ${formatDuration(status.startedAt, status.completedAt)}`);
        }
        
        console.log(`Активных токенов: ${status.activeTokens}`);
        console.log(`Завершенных токенов: ${status.completedTokens}`);
        
        // Переменные
        if (status.variables && Object.keys(status.variables).length > 0) {
            console.log('\nПеременные процесса:');
            for (const [key, value] of Object.entries(status.variables)) {
                try {
                    const parsed = JSON.parse(value);
                    console.log(`  ${key}: ${JSON.stringify(parsed, null, 2)}`);
                } catch {
                    console.log(`  ${key}: ${value}`);
                }
            }
        }
        
        return status;
        
    } catch (error) {
        console.error(`Ошибка получения статуса: ${error.message}`);
        return null;
    }
}

async function waitForCompletion(instanceId, timeoutMinutes = 30, checkInterval = 5000) {
    console.log(`Ожидание завершения процесса ${instanceId} (таймаут: ${timeoutMinutes} мин)`);
    
    const startTime = Date.now();
    const timeoutMs = timeoutMinutes * 60 * 1000;
    
    return new Promise((resolve) => {
        const checkStatus = async () => {
            try {
                const status = await getProcessInstanceStatus(instanceId);
                
                const currentTime = new Date().toLocaleTimeString();
                console.log(`[${currentTime}] Статус: ${status.status}, Активных токенов: ${status.activeTokens}`);
                
                if (status.status === 'COMPLETED') {
                    console.log('✅ Процесс успешно завершен');
                    resolve(true);
                    return;
                } else if (['FAILED', 'CANCELLED'].includes(status.status)) {
                    console.log(`❌ Процесс завершен со статусом: ${status.status}`);
                    resolve(false);
                    return;
                }
                
                if (Date.now() - startTime < timeoutMs) {
                    setTimeout(checkStatus, checkInterval);
                } else {
                    console.log('⏰ Таймаут ожидания завершения процесса');
                    resolve(false);
                }
                
            } catch (error) {
                console.error(`Ошибка проверки статуса: ${error.message}`);
                setTimeout(checkStatus, checkInterval);
            }
        };
        
        checkStatus();
    });
}

async function monitorProcesses(instanceIds, intervalSeconds = 10) {
    console.log(`Мониторинг ${instanceIds.length} процессов с интервалом ${intervalSeconds}с`);
    
    const monitor = async () => {
        try {
            console.log(`\n[${new Date().toLocaleTimeString()}] === Статус процессов ===`);
            
            // Параллельное получение статусов
            const statusPromises = instanceIds.map(async (instanceId) => {
                try {
                    const status = await getProcessInstanceStatus(instanceId);
                    return { instanceId, status, error: null };
                } catch (error) {
                    return { instanceId, status: null, error: error.message };
                }
            });
            
            const results = await Promise.all(statusPromises);
            let activeCount = 0;
            
            // Отображение результатов
            results.forEach(({ instanceId, status, error }) => {
                if (error) {
                    console.log(`❌ ${instanceId}: ошибка - ${error}`);
                } else {
                    const emoji = getStatusEmoji(status.status);
                    console.log(`${emoji} ${instanceId}: ${status.status} (${status.activeTokens} активных токенов)`);
                    
                    if (status.status === 'ACTIVE') {
                        activeCount++;
                    }
                }
            });
            
            if (activeCount === 0) {
                console.log('Все процессы завершены');
                return false; // Остановить мониторинг
            }
            
            console.log(`\nАктивных процессов: ${activeCount}/${instanceIds.length}`);
            return true; // Продолжить мониторинг
            
        } catch (error) {
            console.error('Ошибка мониторинга:', error.message);
            return true; // Продолжить несмотря на ошибку
        }
    };
    
    // Первый запуск
    let shouldContinue = await monitor();
    
    // Периодический мониторинг
    const interval = setInterval(async () => {
        shouldContinue = await monitor();
        
        if (!shouldContinue) {
            clearInterval(interval);
        }
    }, intervalSeconds * 1000);
    
    // Обработка сигнала завершения
    process.on('SIGINT', () => {
        clearInterval(interval);
        console.log('\nМониторинг остановлен пользователем');
        process.exit(0);
    });
}

async function getProcessesSummary(instanceIds) {
    console.log(`=== Сводка по ${instanceIds.length} процессам ===`);
    
    try {
        // Параллельное получение статусов
        const statusPromises = instanceIds.map(async (instanceId) => {
            try {
                return await getProcessInstanceStatus(instanceId);
            } catch {
                return null;
            }
        });
        
        const statuses = (await Promise.all(statusPromises)).filter(Boolean);
        
        // Статистика по статусам
        const statusCounts = {};
        let totalActiveTokens = 0;
        let totalCompletedTokens = 0;
        
        statuses.forEach(status => {
            statusCounts[status.status] = (statusCounts[status.status] || 0) + 1;
            totalActiveTokens += status.activeTokens;
            totalCompletedTokens += status.completedTokens;
        });
        
        console.log('Статистика по статусам:');
        Object.entries(statusCounts)
            .sort(([a], [b]) => a.localeCompare(b))
            .forEach(([statusName, count]) => {
                const emoji = getStatusEmoji(statusName);
                console.log(`  ${emoji} ${statusName}: ${count}`);
            });
        
        console.log('\nОбщая статистика:');
        console.log(`  Всего активных токенов: ${totalActiveTokens}`);
        console.log(`  Всего завершенных токенов: ${totalCompletedTokens}`);
        console.log(`  Загруженных статусов: ${statuses.length}/${instanceIds.length}`);
        
        return {
            statusCounts,
            totalActiveTokens,
            totalCompletedTokens,
            loadedCount: statuses.length,
            totalCount: instanceIds.length
        };
        
    } catch (error) {
        console.error('Ошибка получения сводки:', error.message);
        return null;
    }
}

// Примеры использования
if (require.main === module) {
    const command = process.argv[2];
    const instanceId = process.argv[3];
    
    if (!instanceId) {
        console.log('Использование:');
        console.log('  node status.js show <instance_id>       - показать статус');
        console.log('  node status.js wait <instance_id>       - ждать завершения');
        console.log('  node status.js monitor <id1,id2,...>    - мониторинг нескольких');
        process.exit(1);
    }
    
    switch (command) {
        case 'show':
            displayProcessStatus(instanceId);
            break;
            
        case 'wait':
            waitForCompletion(instanceId);
            break;
            
        case 'monitor':
            const instanceIds = instanceId.split(',');
            monitorProcesses(instanceIds);
            break;
            
        case 'summary':
            const summaryIds = instanceId.split(',');
            getProcessesSummary(summaryIds);
            break;
            
        default:
            console.log('Неизвестная команда:', command);
            process.exit(1);
    }
}

module.exports = {
    getProcessInstanceStatus,
    displayProcessStatus,
    waitForCompletion,
    monitorProcesses,
    getProcessesSummary
};
```

## Мониторинг и отладка

### Health Check процесса
```go
func isProcessHealthy(client pb.ProcessServiceClient, ctx context.Context, instanceId string) bool {
    response, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
        InstanceId: instanceId,
    })
    
    if err != nil || !response.Success {
        return false
    }
    
    // Процесс здоров если он активен или корректно завершен
    return response.Status == "ACTIVE" || response.Status == "COMPLETED"
}
```

### Performance метрики
```python
def collect_performance_metrics(instance_ids):
    """Сбор метрик производительности"""
    metrics = {
        'total_processes': len(instance_ids),
        'active_processes': 0,
        'completed_processes': 0,
        'failed_processes': 0,
        'avg_active_tokens': 0,
        'total_active_tokens': 0
    }
    
    active_tokens_sum = 0
    
    for instance_id in instance_ids:
        status = get_process_instance_status(instance_id)
        if not status:
            continue
            
        if status['status'] == 'ACTIVE':
            metrics['active_processes'] += 1
            active_tokens_sum += status['active_tokens']
        elif status['status'] == 'COMPLETED':
            metrics['completed_processes'] += 1
        elif status['status'] == 'FAILED':
            metrics['failed_processes'] += 1
    
    metrics['total_active_tokens'] = active_tokens_sum
    if metrics['active_processes'] > 0:
        metrics['avg_active_tokens'] = active_tokens_sum / metrics['active_processes']
    
    return metrics
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный instance_id
- `NOT_FOUND` (5): Экземпляр процесса не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Process instance 'invalid-id' not found"
}
```

## Связанные методы
- [StartProcessInstance](start-process-instance.md) - Запуск нового экземпляра
- [CancelProcessInstance](cancel-process-instance.md) - Отмена экземпляра
- [ListProcessInstances](list-process-instances.md) - Список экземпляров
- [ListTokens](list-tokens.md) - Детали токенов процесса

