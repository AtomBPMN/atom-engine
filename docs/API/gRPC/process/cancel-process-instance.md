# CancelProcessInstance

## Описание
Отменяет выполнение экземпляра процесса, останавливая все активные токены и переводя процесс в статус CANCELLED.

## Синтаксис
```protobuf
rpc CancelProcessInstance(CancelProcessInstanceRequest) returns (CancelProcessInstanceResponse);
```

## Package
```protobuf
package atom.process.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### CancelProcessInstanceRequest
```protobuf
message CancelProcessInstanceRequest {
  string instance_id = 1;      // ID экземпляра процесса
  string reason = 2;           // Причина отмены (опционально)
}
```

#### Поля:
- **instance_id** (string, required): Уникальный идентификатор экземпляра процесса
- **reason** (string, optional): Причина отмены для аудита и логирования

## Параметры ответа

### CancelProcessInstanceResponse
```protobuf
message CancelProcessInstanceResponse {
  string instance_id = 1;      // ID отмененного экземпляра
  bool success = 2;            // Статус успешности операции
  string message = 3;          // Сообщение о результате
}
```

#### Поля ответа:
- **instance_id** (string): ID экземпляра процесса
- **success** (bool): `true` если отмена выполнена успешно
- **message** (string): Описание результата операции

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
    
    // Простая отмена процесса
    response, err := client.CancelProcessInstance(ctx, &pb.CancelProcessInstanceRequest{
        InstanceId: instanceId,
        Reason:     "User requested cancellation",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Процесс %s успешно отменен\n", response.InstanceId)
        fmt.Printf("Сообщение: %s\n", response.Message)
    } else {
        fmt.Printf("Ошибка отмены: %s\n", response.Message)
    }
}

// Безопасная отмена с подтверждением
func safeCancelProcess(client pb.ProcessServiceClient, ctx context.Context, instanceId, reason string) error {
    // Сначала проверяем статус процесса
    statusResponse, err := client.GetProcessInstanceStatus(ctx, &pb.GetProcessInstanceStatusRequest{
        InstanceId: instanceId,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка получения статуса: %v", err)
    }
    
    if !statusResponse.Success {
        return fmt.Errorf("не удалось получить статус: %s", statusResponse.Message)
    }
    
    // Проверяем, можно ли отменить процесс
    if statusResponse.Status != "ACTIVE" {
        return fmt.Errorf("процесс в статусе '%s' нельзя отменить", statusResponse.Status)
    }
    
    fmt.Printf("Процесс %s в статусе ACTIVE\n", instanceId)
    fmt.Printf("Активных токенов: %d\n", statusResponse.ActiveTokens)
    fmt.Printf("Причина отмены: %s\n", reason)
    
    // Запрос подтверждения
    fmt.Print("Подтвердите отмену процесса (y/N): ")
    var confirm string
    fmt.Scanln(&confirm)
    
    if confirm != "y" && confirm != "Y" {
        return fmt.Errorf("отмена отклонена пользователем")
    }
    
    // Выполнение отмены
    cancelResponse, err := client.CancelProcessInstance(ctx, &pb.CancelProcessInstanceRequest{
        InstanceId: instanceId,
        Reason:     reason,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка отмены: %v", err)
    }
    
    if !cancelResponse.Success {
        return fmt.Errorf("отмена не выполнена: %s", cancelResponse.Message)
    }
    
    fmt.Printf("Процесс успешно отменен: %s\n", cancelResponse.Message)
    return nil
}

// Массовая отмена процессов
func cancelMultipleProcesses(client pb.ProcessServiceClient, ctx context.Context, instanceIds []string, reason string) map[string]error {
    results := make(map[string]error)
    
    fmt.Printf("Отмена %d процессов...\n", len(instanceIds))
    
    for i, instanceId := range instanceIds {
        fmt.Printf("Отмена %d/%d: %s\n", i+1, len(instanceIds), instanceId)
        
        response, err := client.CancelProcessInstance(ctx, &pb.CancelProcessInstanceRequest{
            InstanceId: instanceId,
            Reason:     reason,
        })
        
        if err != nil {
            results[instanceId] = fmt.Errorf("gRPC ошибка: %v", err)
            fmt.Printf("  ❌ Ошибка: %v\n", err)
            continue
        }
        
        if response.Success {
            results[instanceId] = nil
            fmt.Printf("  ✅ Успешно отменен\n")
        } else {
            results[instanceId] = fmt.Errorf("отмена не выполнена: %s", response.Message)
            fmt.Printf("  ❌ Ошибка: %s\n", response.Message)
        }
        
        // Небольшая задержка между запросами
        time.Sleep(100 * time.Millisecond)
    }
    
    // Итоговый отчет
    successful := 0
    for _, err := range results {
        if err == nil {
            successful++
        }
    }
    
    fmt.Printf("\nИтого: %d/%d процессов отменено успешно\n", successful, len(instanceIds))
    
    return results
}

// Отмена процессов по фильтру
func cancelProcessesByFilter(client pb.ProcessServiceClient, ctx context.Context, processKey, reason string) error {
    // Получаем список активных процессов
    listResponse, err := client.ListProcessInstances(ctx, &pb.ListProcessInstancesRequest{
        StatusFilter:     "ACTIVE",
        ProcessKeyFilter: processKey,
        Limit:           1000,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка получения списка процессов: %v", err)
    }
    
    if !listResponse.Success {
        return fmt.Errorf("не удалось получить список: %s", listResponse.Message)
    }
    
    if len(listResponse.Instances) == 0 {
        fmt.Printf("Активные процессы '%s' не найдены\n", processKey)
        return nil
    }
    
    fmt.Printf("Найдено %d активных процессов '%s'\n", len(listResponse.Instances), processKey)
    
    // Собираем ID для отмены
    instanceIds := make([]string, len(listResponse.Instances))
    for i, instance := range listResponse.Instances {
        instanceIds[i] = instance.InstanceId
        fmt.Printf("  - %s (запущен: %s)\n", 
            instance.InstanceId, 
            time.Unix(instance.StartedAt, 0).Format("2006-01-02 15:04:05"))
    }
    
    fmt.Printf("\nПричина отмены: %s\n", reason)
    fmt.Print("Подтвердите массовую отмену (y/N): ")
    var confirm string
    fmt.Scanln(&confirm)
    
    if confirm != "y" && confirm != "Y" {
        return fmt.Errorf("массовая отмена отклонена")
    }
    
    // Выполнение массовой отмены
    results := cancelMultipleProcesses(client, ctx, instanceIds, reason)
    
    // Проверка результатов
    failed := []string{}
    for instanceId, err := range results {
        if err != nil {
            failed = append(failed, instanceId)
        }
    }
    
    if len(failed) > 0 {
        fmt.Printf("\nНе удалось отменить %d процессов:\n", len(failed))
        for _, instanceId := range failed {
            fmt.Printf("  - %s: %v\n", instanceId, results[instanceId])
        }
    }
    
    return nil
}
```

### Python
```python
import grpc
import time
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor, as_completed

import process_pb2
import process_pb2_grpc

def cancel_process_instance(instance_id, reason=""):
    channel = grpc.insecure_channel('localhost:27500')
    stub = process_pb2_grpc.ProcessServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = process_pb2.CancelProcessInstanceRequest(
        instance_id=instance_id,
        reason=reason
    )
    
    try:
        response = stub.CancelProcessInstance(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Процесс {response.instance_id} успешно отменен")
            print(f"   Сообщение: {response.message}")
            return True
        else:
            print(f"❌ Ошибка отмены: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

def safe_cancel_process(instance_id, reason=""):
    """Безопасная отмена с проверками"""
    print(f"Безопасная отмена процесса {instance_id}")
    
    # Проверяем статус процесса
    from get_process_instance_status import get_process_instance_status
    
    status = get_process_instance_status(instance_id)
    if not status:
        print("❌ Не удалось получить статус процесса")
        return False
    
    print(f"Текущий статус: {status['status']}")
    print(f"Активных токенов: {status['active_tokens']}")
    
    # Проверяем возможность отмены
    if status['status'] != 'ACTIVE':
        print(f"❌ Процесс в статусе '{status['status']}' нельзя отменить")
        return False
    
    if not reason:
        reason = input("Введите причину отмены: ").strip()
    
    print(f"Причина отмены: {reason}")
    
    # Подтверждение
    confirm = input("Подтвердите отмену (y/N): ").strip().lower()
    if confirm not in ['y', 'yes']:
        print("Отмена отклонена пользователем")
        return False
    
    # Выполнение отмены
    return cancel_process_instance(instance_id, reason)

def cancel_multiple_processes(instance_ids, reason="Batch cancellation", max_workers=5):
    """Параллельная отмена нескольких процессов"""
    print(f"Массовая отмена {len(instance_ids)} процессов...")
    
    if not reason:
        reason = input("Введите причину массовой отмены: ").strip()
    
    print(f"Причина: {reason}")
    confirm = input(f"Подтвердите отмену {len(instance_ids)} процессов (y/N): ").strip().lower()
    
    if confirm not in ['y', 'yes']:
        print("Массовая отмена отклонена")
        return {}
    
    results = {}
    
    # Параллельное выполнение отмены
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        # Создаем задачи
        future_to_id = {
            executor.submit(cancel_process_instance, instance_id, reason): instance_id
            for instance_id in instance_ids
        }
        
        # Обрабатываем результаты по мере выполнения
        for future in as_completed(future_to_id):
            instance_id = future_to_id[future]
            try:
                success = future.result()
                results[instance_id] = success
                
                if success:
                    print(f"✅ {instance_id}: отменен")
                else:
                    print(f"❌ {instance_id}: ошибка отмены")
                    
            except Exception as e:
                results[instance_id] = False
                print(f"❌ {instance_id}: исключение - {e}")
    
    # Итоговая статистика
    successful = sum(1 for success in results.values() if success)
    print(f"\n=== Результаты массовой отмены ===")
    print(f"Успешно отменено: {successful}/{len(instance_ids)}")
    
    failed = [iid for iid, success in results.items() if not success]
    if failed:
        print("Не удалось отменить:")
        for instance_id in failed:
            print(f"  - {instance_id}")
    
    return results

def cancel_processes_by_criteria(process_key=None, max_age_hours=None, reason=""):
    """Отмена процессов по критериям"""
    print("Поиск процессов для отмены...")
    
    # Получаем список активных процессов
    from list_process_instances import list_process_instances
    
    filters = {'status': 'ACTIVE'}
    if process_key:
        filters['process_key'] = process_key
    
    processes = list_process_instances(filters)
    if not processes:
        print("Активные процессы не найдены")
        return {}
    
    # Фильтрация по возрасту
    candidates = []
    if max_age_hours:
        cutoff_time = time.time() - (max_age_hours * 3600)
        for process in processes:
            if process['started_at'] < cutoff_time:
                candidates.append(process)
    else:
        candidates = processes
    
    if not candidates:
        print("Процессы, соответствующие критериям, не найдены")
        return {}
    
    print(f"Найдено {len(candidates)} процессов для отмены:")
    for process in candidates:
        started = datetime.fromtimestamp(process['started_at']).strftime('%Y-%m-%d %H:%M:%S')
        print(f"  - {process['instance_id']} ({process['process_key']}) - запущен {started}")
    
    if not reason:
        reason = f"Автоматическая отмена по критериям: process_key={process_key}, max_age_hours={max_age_hours}"
    
    instance_ids = [p['instance_id'] for p in candidates]
    return cancel_multiple_processes(instance_ids, reason)

def emergency_stop_all_processes(confirmation_phrase="EMERGENCY STOP"):
    """Экстренная остановка всех активных процессов"""
    print("🚨 ЭКСТРЕННАЯ ОСТАНОВКА ВСЕХ ПРОЦЕССОВ 🚨")
    print("Это действие отменит ВСЕ активные процессы в системе!")
    
    # Двойное подтверждение
    phrase = input(f"Для подтверждения введите '{confirmation_phrase}': ").strip()
    if phrase != confirmation_phrase:
        print("Экстренная остановка отменена - неверная фраза подтверждения")
        return {}
    
    final_confirm = input("Последнее подтверждение (YES/no): ").strip()
    if final_confirm != "YES":
        print("Экстренная остановка отменена")
        return {}
    
    # Получение всех активных процессов
    from list_process_instances import list_process_instances
    
    processes = list_process_instances({'status': 'ACTIVE', 'limit': 10000})
    if not processes:
        print("Активные процессы не найдены")
        return {}
    
    print(f"Найдено {len(processes)} активных процессов")
    
    instance_ids = [p['instance_id'] for p in processes]
    reason = f"Emergency stop at {datetime.now().isoformat()}"
    
    return cancel_multiple_processes(instance_ids, reason, max_workers=10)

def cancel_with_grace_period(instance_id, grace_seconds=30, reason=""):
    """Отмена с периодом ожидания"""
    print(f"Отмена процесса {instance_id} с периодом ожидания {grace_seconds}с")
    
    if not reason:
        reason = f"Graceful cancellation with {grace_seconds}s grace period"
    
    # Здесь можно было бы отправить сигнал процессу о предстоящей отмене
    # Например, опубликовать сообщение или установить переменную
    
    print(f"Период ожидания {grace_seconds} секунд...")
    time.sleep(grace_seconds)
    
    print("Выполнение отмены...")
    return cancel_process_instance(instance_id, reason)

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python cancel.py <instance_id> [reason]")
        print("  python cancel.py --batch <id1,id2,id3> [reason]")
        print("  python cancel.py --by-key <process_key> [reason]")
        print("  python cancel.py --emergency")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "--emergency":
        emergency_stop_all_processes()
    elif command == "--batch":
        if len(sys.argv) < 3:
            print("Требуется список ID процессов")
            sys.exit(1)
        instance_ids = sys.argv[2].split(',')
        reason = sys.argv[3] if len(sys.argv) > 3 else ""
        cancel_multiple_processes(instance_ids, reason)
    elif command == "--by-key":
        if len(sys.argv) < 3:
            print("Требуется process_key")
            sys.exit(1)
        process_key = sys.argv[2]
        reason = sys.argv[3] if len(sys.argv) > 3 else ""
        cancel_processes_by_criteria(process_key=process_key, reason=reason)
    else:
        # Простая отмена одного процесса
        instance_id = command
        reason = sys.argv[2] if len(sys.argv) > 2 else ""
        safe_cancel_process(instance_id, reason)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const readline = require('readline');

const PROTO_PATH = 'process.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const processProto = grpc.loadPackageDefinition(packageDefinition).atom.process.v1;

// Создание интерфейса для ввода пользователя
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

function askQuestion(question) {
    return new Promise(resolve => {
        rl.question(question, resolve);
    });
}

async function cancelProcessInstance(instanceId, reason = '') {
    const client = new processProto.ProcessService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            instance_id: instanceId,
            reason: reason
        };
        
        client.cancelProcessInstance(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Процесс ${response.instance_id} успешно отменен`);
                console.log(`   Сообщение: ${response.message}`);
                resolve(true);
            } else {
                console.log(`❌ Ошибка отмены: ${response.message}`);
                resolve(false);
            }
        });
    });
}

async function safeCancelProcess(instanceId, reason = '') {
    console.log(`Безопасная отмена процесса ${instanceId}`);
    
    try {
        // Проверяем статус процесса
        const { getProcessInstanceStatus } = require('./get-process-instance-status');
        const status = await getProcessInstanceStatus(instanceId);
        
        console.log(`Текущий статус: ${status.status}`);
        console.log(`Активных токенов: ${status.activeTokens}`);
        
        // Проверяем возможность отмены
        if (status.status !== 'ACTIVE') {
            console.log(`❌ Процесс в статусе '${status.status}' нельзя отменить`);
            return false;
        }
        
        // Запрашиваем причину если не указана
        if (!reason) {
            reason = await askQuestion('Введите причину отмены: ');
        }
        
        console.log(`Причина отмены: ${reason}`);
        
        // Подтверждение
        const confirm = await askQuestion('Подтвердите отмену (y/N): ');
        if (confirm.toLowerCase() !== 'y' && confirm.toLowerCase() !== 'yes') {
            console.log('Отмена отклонена пользователем');
            return false;
        }
        
        // Выполнение отмены
        return await cancelProcessInstance(instanceId, reason);
        
    } catch (error) {
        console.error(`Ошибка: ${error.message}`);
        return false;
    }
}

async function cancelMultipleProcesses(instanceIds, reason = 'Batch cancellation') {
    console.log(`Массовая отмена ${instanceIds.length} процессов...`);
    
    if (!reason || reason === 'Batch cancellation') {
        reason = await askQuestion('Введите причину массовой отмены: ');
    }
    
    console.log(`Причина: ${reason}`);
    const confirm = await askQuestion(`Подтвердите отмену ${instanceIds.length} процессов (y/N): `);
    
    if (confirm.toLowerCase() !== 'y' && confirm.toLowerCase() !== 'yes') {
        console.log('Массовая отмена отклонена');
        return {};
    }
    
    const results = {};
    const batchSize = 5; // Количество параллельных отмен
    
    // Обработка батчами для контроля нагрузки
    for (let i = 0; i < instanceIds.length; i += batchSize) {
        const batch = instanceIds.slice(i, i + batchSize);
        
        console.log(`Обработка батча ${Math.floor(i/batchSize) + 1}/${Math.ceil(instanceIds.length/batchSize)} (${batch.length} процессов)`);
        
        // Параллельная отмена в батче
        const batchPromises = batch.map(async (instanceId) => {
            try {
                const success = await cancelProcessInstance(instanceId, reason);
                results[instanceId] = success;
                
                if (success) {
                    console.log(`✅ ${instanceId}: отменен`);
                } else {
                    console.log(`❌ ${instanceId}: ошибка отмены`);
                }
                
                return { instanceId, success };
            } catch (error) {
                results[instanceId] = false;
                console.log(`❌ ${instanceId}: исключение - ${error.message}`);
                return { instanceId, success: false };
            }
        });
        
        await Promise.all(batchPromises);
        
        // Небольшая пауза между батчами
        if (i + batchSize < instanceIds.length) {
            await new Promise(resolve => setTimeout(resolve, 1000));
        }
    }
    
    // Итоговая статистика
    const successful = Object.values(results).filter(success => success).length;
    console.log(`\n=== Результаты массовой отмены ===`);
    console.log(`Успешно отменено: ${successful}/${instanceIds.length}`);
    
    const failed = Object.keys(results).filter(instanceId => !results[instanceId]);
    if (failed.length > 0) {
        console.log('Не удалось отменить:');
        failed.forEach(instanceId => {
            console.log(`  - ${instanceId}`);
        });
    }
    
    return results;
}

async function cancelProcessesByCriteria(options = {}) {
    const { processKey, maxAgeHours, reason = '' } = options;
    
    console.log('Поиск процессов для отмены...');
    
    try {
        // Получаем список активных процессов
        const { listProcessInstances } = require('./list-process-instances');
        
        const filters = { status: 'ACTIVE' };
        if (processKey) {
            filters.processKey = processKey;
        }
        
        const processes = await listProcessInstances(filters);
        if (!processes || processes.length === 0) {
            console.log('Активные процессы не найдены');
            return {};
        }
        
        // Фильтрация по возрасту
        let candidates = processes;
        if (maxAgeHours) {
            const cutoffTime = Date.now() - (maxAgeHours * 60 * 60 * 1000);
            candidates = processes.filter(process => {
                const startTime = new Date(process.startedAt).getTime();
                return startTime < cutoffTime;
            });
        }
        
        if (candidates.length === 0) {
            console.log('Процессы, соответствующие критериям, не найдены');
            return {};
        }
        
        console.log(`Найдено ${candidates.length} процессов для отмены:`);
        candidates.forEach(process => {
            const started = new Date(process.startedAt).toLocaleString();
            console.log(`  - ${process.instanceId} (${process.processKey}) - запущен ${started}`);
        });
        
        const finalReason = reason || `Автоматическая отмена по критериям: processKey=${processKey}, maxAgeHours=${maxAgeHours}`;
        
        const instanceIds = candidates.map(p => p.instanceId);
        return await cancelMultipleProcesses(instanceIds, finalReason);
        
    } catch (error) {
        console.error(`Ошибка поиска процессов: ${error.message}`);
        return {};
    }
}

async function emergencyStopAllProcesses(confirmationPhrase = 'EMERGENCY STOP') {
    console.log('🚨 ЭКСТРЕННАЯ ОСТАНОВКА ВСЕХ ПРОЦЕССОВ 🚨');
    console.log('Это действие отменит ВСЕ активные процессы в системе!');
    
    try {
        // Двойное подтверждение
        const phrase = await askQuestion(`Для подтверждения введите '${confirmationPhrase}': `);
        if (phrase !== confirmationPhrase) {
            console.log('Экстренная остановка отменена - неверная фраза подтверждения');
            return {};
        }
        
        const finalConfirm = await askQuestion('Последнее подтверждение (YES/no): ');
        if (finalConfirm !== 'YES') {
            console.log('Экстренная остановка отменена');
            return {};
        }
        
        // Получение всех активных процессов
        const { listProcessInstances } = require('./list-process-instances');
        
        const processes = await listProcessInstances({ 
            status: 'ACTIVE', 
            limit: 10000 
        });
        
        if (!processes || processes.length === 0) {
            console.log('Активные процессы не найдены');
            return {};
        }
        
        console.log(`Найдено ${processes.length} активных процессов`);
        
        const instanceIds = processes.map(p => p.instanceId);
        const reason = `Emergency stop at ${new Date().toISOString()}`;
        
        return await cancelMultipleProcesses(instanceIds, reason);
        
    } catch (error) {
        console.error(`Ошибка экстренной остановки: ${error.message}`);
        return {};
    }
}

async function cancelWithGracePeriod(instanceId, graceSeconds = 30, reason = '') {
    console.log(`Отмена процесса ${instanceId} с периодом ожидания ${graceSeconds}с`);
    
    if (!reason) {
        reason = `Graceful cancellation with ${graceSeconds}s grace period`;
    }
    
    // Здесь можно было бы отправить предупреждение процессу
    // Например, опубликовать сообщение или установить переменную
    
    console.log(`Период ожидания ${graceSeconds} секунд...`);
    
    // Обратный отсчет
    for (let i = graceSeconds; i > 0; i--) {
        if (i <= 10 || i % 10 === 0) {
            process.stdout.write(`${i}... `);
        }
        await new Promise(resolve => setTimeout(resolve, 1000));
    }
    
    console.log('\nВыполнение отмены...');
    return await cancelProcessInstance(instanceId, reason);
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node cancel.js <instance_id> [reason]');
        console.log('  node cancel.js --batch <id1,id2,id3> [reason]');
        console.log('  node cancel.js --by-key <process_key> [reason]');
        console.log('  node cancel.js --emergency');
        console.log('  node cancel.js --grace <instance_id> [seconds] [reason]');
        process.exit(1);
    }
    
    const command = args[0];
    
    (async () => {
        try {
            switch (command) {
                case '--emergency':
                    await emergencyStopAllProcesses();
                    break;
                    
                case '--batch':
                    if (args.length < 2) {
                        console.log('Требуется список ID процессов');
                        process.exit(1);
                    }
                    const instanceIds = args[1].split(',');
                    const batchReason = args[2] || '';
                    await cancelMultipleProcesses(instanceIds, batchReason);
                    break;
                    
                case '--by-key':
                    if (args.length < 2) {
                        console.log('Требуется process_key');
                        process.exit(1);
                    }
                    const processKey = args[1];
                    const keyReason = args[2] || '';
                    await cancelProcessesByCriteria({ processKey, reason: keyReason });
                    break;
                    
                case '--grace':
                    if (args.length < 2) {
                        console.log('Требуется instance_id');
                        process.exit(1);
                    }
                    const graceInstanceId = args[1];
                    const graceSeconds = parseInt(args[2]) || 30;
                    const graceReason = args[3] || '';
                    await cancelWithGracePeriod(graceInstanceId, graceSeconds, graceReason);
                    break;
                    
                default:
                    // Простая отмена одного процесса
                    const instanceId = command;
                    const reason = args[1] || '';
                    await safeCancelProcess(instanceId, reason);
                    break;
            }
        } catch (error) {
            console.error('Ошибка:', error.message);
        } finally {
            rl.close();
        }
    })();
}

module.exports = {
    cancelProcessInstance,
    safeCancelProcess,
    cancelMultipleProcesses,
    cancelProcessesByCriteria,
    emergencyStopAllProcesses,
    cancelWithGracePeriod
};
```

## Политики отмены

### Graceful Cancellation
```go
// Мягкая отмена с уведомлением
func gracefulCancel(client pb.ProcessServiceClient, ctx context.Context, instanceId string) error {
    // 1. Отправляем сигнал о предстоящей отмене
    variables := map[string]string{
        "cancellation_requested": "true",
        "cancellation_time": time.Now().Add(30 * time.Second).Format(time.RFC3339),
    }
    
    // Устанавливаем переменные процесса (если есть такой метод)
    // updateResponse, _ := client.UpdateProcessVariables(ctx, &pb.UpdateProcessVariablesRequest{
    //     InstanceId: instanceId,
    //     Variables:  variables,
    // })
    
    // 2. Ждем период grace
    time.Sleep(30 * time.Second)
    
    // 3. Выполняем отмену
    return cancel(client, ctx, instanceId, "Graceful cancellation after 30s grace period")
}
```

### Force Cancellation
```python
def force_cancel_process(instance_id, reason="Force cancellation"):
    """Принудительная отмена без проверок"""
    print(f"⚠️  ПРИНУДИТЕЛЬНАЯ отмена процесса {instance_id}")
    return cancel_process_instance(instance_id, f"FORCE: {reason}")
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный instance_id
- `NOT_FOUND` (5): Экземпляр процесса не найден  
- `FAILED_PRECONDITION` (9): Процесс нельзя отменить (уже завершен)
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Cannot cancel process instance 'srv1-abc123': already completed"
}
```

```json
{
  "success": false,
  "message": "Process instance 'invalid-id' not found"
}
```

## Лучшие практики

### Аудит отмен
- Всегда указывайте причину отмены
- Ведите лог всех отмен для аудита
- Уведомляйте заинтересованные стороны

### Безопасность
- Проверяйте статус перед отменой
- Используйте подтверждения для критичных процессов
- Ограничивайте права на массовую отмену

### Мониторинг
- Отслеживайте частоту отмен
- Анализируйте причины отмен
- Настройте алерты на аномальную активность

## Связанные методы
- [GetProcessInstanceStatus](get-process-instance-status.md) - Проверка статуса перед отменой
- [ListProcessInstances](list-process-instances.md) - Поиск процессов для отмены
- [StartProcessInstance](start-process-instance.md) - Запуск нового экземпляра