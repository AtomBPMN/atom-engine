# DeleteBPMNProcess

## Описание
Удаляет BPMN процесс из системы по его ID. Удаляются все версии процесса и связанные данные.

## Синтаксис
```protobuf
rpc DeleteBPMNProcess(DeleteBPMNProcessRequest) returns (DeleteBPMNProcessResponse);
```

## Package
```protobuf
package parser;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `parser` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### DeleteBPMNProcessRequest
```protobuf
message DeleteBPMNProcessRequest {
  string process_id = 1;      // ID процесса для удаления
}
```

#### Поля:
- **process_id** (string, required): ID процесса для удаления (не ключ, а именно process_id)

## Параметры ответа

### DeleteBPMNProcessResponse
```protobuf
message DeleteBPMNProcessResponse {
  bool success = 1;           // Статус успешности операции
  string message = 2;         // Сообщение о результате
  int32 deleted_versions = 3; // Количество удаленных версий
  repeated string deleted_keys = 4; // Список удаленных ключей процессов
}
```

#### Поля ответа:
- **success** (bool): `true` если процесс успешно удален
- **message** (string): Описание результата операции
- **deleted_versions** (int32): Количество удаленных версий процесса
- **deleted_keys** (repeated string): Список ключей удаленных процессов

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
    
    // Удаление процесса
    response, err := client.DeleteBPMNProcess(ctx, &pb.DeleteBPMNProcessRequest{
        ProcessId: "order-process-v1",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Процесс успешно удален!\n")
        fmt.Printf("Удалено версий: %d\n", response.DeletedVersions)
        fmt.Printf("Удаленные ключи:\n")
        for _, key := range response.DeletedKeys {
            fmt.Printf("  - %s\n", key)
        }
    } else {
        fmt.Printf("Ошибка удаления: %s\n", response.Message)
    }
}

// Безопасное удаление с подтверждением
func safeDeleteProcess(client pb.ParserServiceClient, ctx context.Context, processId string) error {
    // Сначала получаем информацию о процессе
    listResponse, err := client.ListBPMNProcesses(ctx, &pb.ListBPMNProcessesRequest{
        ProcessId: processId,
        Limit:     100,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка получения информации о процессе: %v", err)
    }
    
    if !listResponse.Success || len(listResponse.Processes) == 0 {
        return fmt.Errorf("процесс '%s' не найден", processId)
    }
    
    // Показываем что будет удалено
    fmt.Printf("Будет удалено процессов с ID '%s': %d\n", processId, len(listResponse.Processes))
    for _, process := range listResponse.Processes {
        fmt.Printf("  - %s (v%s) - %s\n", 
            process.ProcessKey, process.Version, process.Status)
    }
    
    // Запрос подтверждения
    fmt.Print("Продолжить удаление? (y/N): ")
    var confirm string
    fmt.Scanln(&confirm)
    
    if confirm != "y" && confirm != "Y" {
        return fmt.Errorf("удаление отменено пользователем")
    }
    
    // Выполнение удаления
    deleteResponse, err := client.DeleteBPMNProcess(ctx, &pb.DeleteBPMNProcessRequest{
        ProcessId: processId,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка удаления: %v", err)
    }
    
    if !deleteResponse.Success {
        return fmt.Errorf("удаление не выполнено: %s", deleteResponse.Message)
    }
    
    fmt.Printf("Успешно удалено %d версий процесса\n", deleteResponse.DeletedVersions)
    return nil
}
```

### Python
```python
import grpc
import parser_pb2
import parser_pb2_grpc

def delete_bpmn_process(process_id, confirm=False):
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    try:
        # Если нужно подтверждение, сначала показываем информацию
        if not confirm:
            # Получаем список процессов для показа
            list_request = parser_pb2.ListBPMNProcessesRequest(
                process_id=process_id,
                limit=100
            )
            list_response = stub.ListBPMNProcesses(list_request, metadata=metadata)
            
            if list_response.success and list_response.processes:
                print(f"Найдено процессов с ID '{process_id}': {len(list_response.processes)}")
                for process in list_response.processes:
                    print(f"  - {process.process_key} (v{process.version}) - {process.status}")
                
                user_input = input("Продолжить удаление? (y/N): ").strip().lower()
                if user_input not in ['y', 'yes']:
                    print("Удаление отменено")
                    return False
            else:
                print(f"Процесс '{process_id}' не найден")
                return False
        
        # Выполнение удаления
        delete_request = parser_pb2.DeleteBPMNProcessRequest(
            process_id=process_id
        )
        
        response = stub.DeleteBPMNProcess(delete_request, metadata=metadata)
        
        if response.success:
            print(f"Процесс '{process_id}' успешно удален!")
            print(f"Удалено версий: {response.deleted_versions}")
            if response.deleted_keys:
                print("Удаленные ключи:")
                for key in response.deleted_keys:
                    print(f"  - {key}")
            return True
        else:
            print(f"Ошибка удаления: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

def bulk_delete_processes(process_ids):
    """Массовое удаление процессов"""
    print(f"Удаление {len(process_ids)} процессов...")
    
    results = {}
    for process_id in process_ids:
        print(f"\nУдаление {process_id}...")
        success = delete_bpmn_process(process_id, confirm=True)  # Без подтверждения
        results[process_id] = success
        
        if success:
            print(f"✅ {process_id} удален")
        else:
            print(f"❌ {process_id} не удален")
    
    # Итоговый отчет
    print(f"\n=== Итоги удаления ===")
    successful = sum(1 for success in results.values() if success)
    print(f"Успешно удалено: {successful}/{len(process_ids)}")
    
    failed = [pid for pid, success in results.items() if not success]
    if failed:
        print("Не удалось удалить:")
        for pid in failed:
            print(f"  - {pid}")
    
    return results

def cleanup_old_processes(days_old=30):
    """Очистка старых процессов"""
    from datetime import datetime, timedelta
    
    channel = grpc.insecure_channel('localhost:27500')
    stub = parser_pb2_grpc.ParserServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Получаем все процессы
    list_request = parser_pb2.ListBPMNProcessesRequest(limit=1000)
    list_response = stub.ListBPMNProcesses(list_request, metadata=metadata)
    
    if not list_response.success:
        print(f"Ошибка получения списка процессов: {list_response.message}")
        return
    
    # Фильтрация старых процессов
    cutoff_date = datetime.now() - timedelta(days=days_old)
    old_processes = []
    
    for process in list_response.processes:
        try:
            created_at = datetime.fromisoformat(process.created_at.replace('Z', '+00:00'))
            if created_at < cutoff_date:
                old_processes.append(process.process_id)
        except ValueError:
            continue  # Пропускаем процессы с некорректной датой
    
    if not old_processes:
        print(f"Не найдено процессов старше {days_old} дней")
        return
    
    print(f"Найдено {len(old_processes)} процессов старше {days_old} дней")
    
    # Группируем по уникальным ID (может быть несколько версий)
    unique_ids = list(set(old_processes))
    print(f"Уникальных процессов для удаления: {len(unique_ids)}")
    
    user_input = input("Продолжить очистку? (y/N): ").strip().lower()
    if user_input not in ['y', 'yes']:
        print("Очистка отменена")
        return
    
    # Удаление
    bulk_delete_processes(unique_ids)

if __name__ == "__main__":
    # Примеры использования
    
    # Простое удаление
    delete_bpmn_process("test-process")
    
    # Массовое удаление
    # bulk_delete_processes(["process1", "process2", "process3"])
    
    # Очистка старых процессов
    # cleanup_old_processes(days_old=60)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const readline = require('readline');

const PROTO_PATH = 'parser.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const parserProto = grpc.loadPackageDefinition(packageDefinition).parser;

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

async function deleteBPMNProcess(processId, options = {}) {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    try {
        // Если нужно подтверждение, показываем информацию о процессе
        if (!options.skipConfirmation) {
            const processes = await listProcessesByID(processId);
            
            if (processes.length === 0) {
                console.log(`Процесс '${processId}' не найден`);
                return false;
            }
            
            console.log(`Найдено процессов с ID '${processId}': ${processes.length}`);
            processes.forEach(process => {
                console.log(`  - ${process.process_key} (v${process.version}) - ${process.status}`);
            });
            
            const confirm = await askQuestion('Продолжить удаление? (y/N): ');
            if (confirm.toLowerCase() !== 'y' && confirm.toLowerCase() !== 'yes') {
                console.log('Удаление отменено');
                return false;
            }
        }
        
        // Выполнение удаления
        const response = await new Promise((resolve, reject) => {
            const request = { process_id: processId };
            
            client.deleteBPMNProcess(request, metadata, (error, response) => {
                if (error) {
                    reject(error);
                } else {
                    resolve(response);
                }
            });
        });
        
        if (response.success) {
            console.log(`Процесс '${processId}' успешно удален!`);
            console.log(`Удалено версий: ${response.deleted_versions}`);
            
            if (response.deleted_keys && response.deleted_keys.length > 0) {
                console.log('Удаленные ключи:');
                response.deleted_keys.forEach(key => {
                    console.log(`  - ${key}`);
                });
            }
            return true;
        } else {
            console.log(`Ошибка удаления: ${response.message}`);
            return false;
        }
        
    } catch (error) {
        console.error('gRPC Error:', error.message);
        return false;
    }
}

async function listProcessesByID(processId) {
    const client = new parserProto.ParserService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            process_id: processId,
            limit: 100
        };
        
        client.listBPMNProcesses(request, metadata, (error, response) => {
            if (error) {
                reject(error);
            } else if (response.success) {
                resolve(response.processes);
            } else {
                resolve([]);
            }
        });
    });
}

async function bulkDeleteProcesses(processIds) {
    console.log(`Удаление ${processIds.length} процессов...`);
    
    const results = {};
    
    for (const processId of processIds) {
        console.log(`\nУдаление ${processId}...`);
        
        const success = await deleteBPMNProcess(processId, { skipConfirmation: true });
        results[processId] = success;
        
        if (success) {
            console.log(`✅ ${processId} удален`);
        } else {
            console.log(`❌ ${processId} не удален`);
        }
    }
    
    // Итоговый отчет
    console.log('\n=== Итоги удаления ===');
    const successful = Object.values(results).filter(success => success).length;
    console.log(`Успешно удалено: ${successful}/${processIds.length}`);
    
    const failed = Object.keys(results).filter(processId => !results[processId]);
    if (failed.length > 0) {
        console.log('Не удалось удалить:');
        failed.forEach(processId => {
            console.log(`  - ${processId}`);
        });
    }
    
    return results;
}

async function interactiveDelete() {
    try {
        const processId = await askQuestion('Введите ID процесса для удаления: ');
        
        if (!processId.trim()) {
            console.log('ID процесса не может быть пустым');
            return;
        }
        
        await deleteBPMNProcess(processId.trim());
        
    } catch (error) {
        console.error('Ошибка:', error.message);
    } finally {
        rl.close();
    }
}

// Функция для удаления процессов по паттерну
async function deleteProcessesByPattern(pattern) {
    try {
        // Получаем все процессы
        const client = new parserProto.ParserService('localhost:27500',
            grpc.credentials.createInsecure());
        
        const metadata = new grpc.Metadata();
        metadata.add('x-api-key', 'your-api-key-here');
        
        const allProcesses = await new Promise((resolve, reject) => {
            client.listBPMNProcesses({ limit: 1000 }, metadata, (error, response) => {
                if (error) {
                    reject(error);
                } else if (response.success) {
                    resolve(response.processes);
                } else {
                    resolve([]);
                }
            });
        });
        
        // Фильтрация по паттерну
        const regex = new RegExp(pattern);
        const matchingProcesses = allProcesses.filter(process => 
            regex.test(process.process_id));
        
        if (matchingProcesses.length === 0) {
            console.log(`Не найдено процессов, соответствующих паттерну: ${pattern}`);
            return;
        }
        
        console.log(`Найдено процессов по паттерну '${pattern}': ${matchingProcesses.length}`);
        matchingProcesses.forEach(process => {
            console.log(`  - ${process.process_id} (v${process.version})`);
        });
        
        const confirm = await askQuestion('Удалить все найденные процессы? (y/N): ');
        if (confirm.toLowerCase() !== 'y' && confirm.toLowerCase() !== 'yes') {
            console.log('Удаление отменено');
            return;
        }
        
        // Получаем уникальные ID
        const uniqueIds = [...new Set(matchingProcesses.map(p => p.process_id))];
        await bulkDeleteProcesses(uniqueIds);
        
    } catch (error) {
        console.error('Ошибка:', error.message);
    } finally {
        rl.close();
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        // Интерактивный режим
        interactiveDelete();
    } else if (args[0] === '--pattern') {
        // Удаление по паттерну
        const pattern = args[1];
        if (!pattern) {
            console.log('Использование: node delete.js --pattern <regex_pattern>');
            process.exit(1);
        }
        deleteProcessesByPattern(pattern);
    } else if (args[0] === '--bulk') {
        // Массовое удаление
        const processIds = args.slice(1);
        if (processIds.length === 0) {
            console.log('Использование: node delete.js --bulk <process_id1> <process_id2> ...');
            process.exit(1);
        }
        bulkDeleteProcesses(processIds).then(() => rl.close());
    } else {
        // Простое удаление
        const processId = args[0];
        deleteBPMNProcess(processId).then(() => rl.close());
    }
}

module.exports = {
    deleteBPMNProcess,
    bulkDeleteProcesses,
    deleteProcessesByPattern
};
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Пустой или неверный process_id
- `NOT_FOUND` (5): Процесс с указанным ID не найден
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `FAILED_PRECONDITION` (9): Процесс используется активными экземплярами

### Примеры ошибок
```json
{
  "success": false,
  "message": "Process 'unknown-process' not found"
}
```

```json
{
  "success": false,
  "message": "Cannot delete process: 3 active instances are running"
}
```

## Безопасность и ограничения

### Проверка активных экземпляров
Перед удалением система проверяет наличие активных экземпляров процесса. Если есть запущенные экземпляры, удаление блокируется.

### Каскадное удаление
При удалении процесса также удаляются:
- ✅ Все версии процесса
- ✅ Метаданные процесса
- ✅ Оригинальные BPMN файлы
- ❌ Активные экземпляры (блокируют удаление)
- ❌ История выполненных экземпляров (остается)

### Восстановление
После удаления процесс невозможно восстановить. Создайте резервную копию перед удалением:

```bash
# Сохранение процесса перед удалением
atomd bpmn json process-key > backup_process.json
```

## Связанные методы
- [ListBPMNProcesses](list-bpmn-processes.md) - Список процессов перед удалением
- [GetBPMNProcess](get-bpmn-process.md) - Информация о процессе
- [ParseBPMNFile](parse-bpmn-file.md) - Загрузка нового процесса
- [GetBPMNStats](get-bpmn-stats.md) - Статистика после удаления

