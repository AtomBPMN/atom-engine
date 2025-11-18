# UpdateJobRetries

## Описание
Обновляет количество оставшихся попыток для задания без его завершения или провала. Полезно для динамического управления retry логикой.

## Синтаксис
```protobuf
rpc UpdateJobRetries(UpdateJobRetriesRequest) returns (UpdateJobRetriesResponse);
```

## Package
```protobuf
package atom.jobs.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `jobs` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### UpdateJobRetriesRequest
```protobuf
message UpdateJobRetriesRequest {
  string job_key = 1;    // Ключ задания
  int32 retries = 2;     // Новое количество попыток
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания
- **retries** (int32, required): Новое количество оставшихся попыток (0-100)

## Параметры ответа

### UpdateJobRetriesResponse
```protobuf
message UpdateJobRetriesResponse {
  bool success = 1;         // Статус успешности операции
  string message = 2;       // Сообщение о результате
  int32 previous_retries = 3; // Предыдущее количество попыток
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
    
    pb "atom-engine/proto/jobs/jobspb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewJobsServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    jobKey := "atom-jobkey12345"
    
    // Простое обновление попыток
    response, err := client.UpdateJobRetries(ctx, &pb.UpdateJobRetriesRequest{
        JobKey:  jobKey,
        Retries: 5,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Попытки обновлены для задания %s: %d → %d\n", 
                   jobKey, response.PreviousRetries, 5)
    } else {
        fmt.Printf("❌ Ошибка обновления: %s\n", response.Message)
    }
}

// Менеджер динамических попыток
type RetryManager struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewRetryManager(client pb.JobsServiceClient, ctx context.Context) *RetryManager {
    return &RetryManager{
        client: client,
        ctx:    ctx,
    }
}

func (rm *RetryManager) IncreaseRetries(jobKey string, additionalRetries int32) error {
    // Получаем текущую информацию о задании
    jobInfo, err := rm.client.GetJob(rm.ctx, &pb.GetJobRequest{
        JobKey: jobKey,
    })
    if err != nil {
        return fmt.Errorf("не удалось получить информацию о задании: %v", err)
    }
    
    if !jobInfo.Success {
        return fmt.Errorf("задание не найдено: %s", jobInfo.Message)
    }
    
    // Вычисляем новое количество попыток
    newRetries := jobInfo.Job.Retries + additionalRetries
    if newRetries > 100 {
        newRetries = 100 // Максимальный лимит
    }
    
    response, err := rm.client.UpdateJobRetries(rm.ctx, &pb.UpdateJobRetriesRequest{
        JobKey:  jobKey,
        Retries: newRetries,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось обновить попытки: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка обновления попыток: %s", response.Message)
    }
    
    fmt.Printf("🔄 Задание %s: попытки увеличены %d → %d (+%d)\n", 
               jobKey, response.PreviousRetries, newRetries, additionalRetries)
    
    return nil
}

func (rm *RetryManager) DecreaseRetries(jobKey string, removedRetries int32) error {
    // Получаем текущую информацию о задании
    jobInfo, err := rm.client.GetJob(rm.ctx, &pb.GetJobRequest{
        JobKey: jobKey,
    })
    if err != nil {
        return fmt.Errorf("не удалось получить информацию о задании: %v", err)
    }
    
    if !jobInfo.Success {
        return fmt.Errorf("задание не найдено: %s", jobInfo.Message)
    }
    
    // Вычисляем новое количество попыток
    newRetries := jobInfo.Job.Retries - removedRetries
    if newRetries < 0 {
        newRetries = 0 // Минимальный лимит
    }
    
    response, err := rm.client.UpdateJobRetries(rm.ctx, &pb.UpdateJobRetriesRequest{
        JobKey:  jobKey,
        Retries: newRetries,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось обновить попытки: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка обновления попыток: %s", response.Message)
    }
    
    fmt.Printf("⬇️ Задание %s: попытки уменьшены %d → %d (-%d)\n", 
               jobKey, response.PreviousRetries, newRetries, removedRetries)
    
    return nil
}

func (rm *RetryManager) SetRetries(jobKey string, retries int32) error {
    if retries < 0 {
        retries = 0
    }
    if retries > 100 {
        retries = 100
    }
    
    response, err := rm.client.UpdateJobRetries(rm.ctx, &pb.UpdateJobRetriesRequest{
        JobKey:  jobKey,
        Retries: retries,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось установить попытки: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка установки попыток: %s", response.Message)
    }
    
    fmt.Printf("📝 Задание %s: попытки установлены %d → %d\n", 
               jobKey, response.PreviousRetries, retries)
    
    return nil
}

func (rm *RetryManager) ResetRetries(jobKey string) error {
    return rm.SetRetries(jobKey, 3) // Стандартное значение
}

// Адаптивная стратегия попыток на основе типа ошибки
type AdaptiveRetryStrategy struct {
    retryManager *RetryManager
}

func NewAdaptiveRetryStrategy(client pb.JobsServiceClient, ctx context.Context) *AdaptiveRetryStrategy {
    return &AdaptiveRetryStrategy{
        retryManager: NewRetryManager(client, ctx),
    }
}

func (ars *AdaptiveRetryStrategy) HandleErrorAndAdjustRetries(jobKey string, errorType string) error {
    switch errorType {
    case "CONNECTION_ERROR":
        // Для ошибок соединения даем больше попыток
        return ars.retryManager.IncreaseRetries(jobKey, 2)
        
    case "RATE_LIMIT":
        // Для rate limit уменьшаем попытки, чтобы не усугублять ситуацию
        return ars.retryManager.DecreaseRetries(jobKey, 1)
        
    case "AUTH_ERROR":
        // Для ошибок авторизации обычно не стоит повторять
        return ars.retryManager.SetRetries(jobKey, 0)
        
    case "VALIDATION_ERROR":
        // Для ошибок валидации повторы бесполезны
        return ars.retryManager.SetRetries(jobKey, 0)
        
    case "TEMPORARY_ERROR":
        // Для временных ошибок сохраняем текущее количество попыток
        return nil
        
    case "SERVICE_UNAVAILABLE":
        // Для недоступности сервиса даем еще шансы
        return ars.retryManager.IncreaseRetries(jobKey, 3)
        
    default:
        // Для неизвестных ошибок используем стандартное количество
        return ars.retryManager.ResetRetries(jobKey)
    }
}

// Групповое обновление попыток
type BatchRetryUpdater struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewBatchRetryUpdater(client pb.JobsServiceClient, ctx context.Context) *BatchRetryUpdater {
    return &BatchRetryUpdater{
        client: client,
        ctx:    ctx,
    }
}

func (bru *BatchRetryUpdater) UpdateMultipleJobs(updates []JobRetryUpdate) ([]UpdateResult, error) {
    results := make([]UpdateResult, 0, len(updates))
    
    for _, update := range updates {
        result := UpdateResult{
            JobKey: update.JobKey,
        }
        
        response, err := bru.client.UpdateJobRetries(bru.ctx, &pb.UpdateJobRetriesRequest{
            JobKey:  update.JobKey,
            Retries: update.NewRetries,
        })
        
        if err != nil {
            result.Error = err.Error()
            result.Success = false
        } else {
            result.Success = response.Success
            result.Message = response.Message
            result.PreviousRetries = response.PreviousRetries
            result.NewRetries = update.NewRetries
        }
        
        results = append(results, result)
        
        if result.Success {
            fmt.Printf("✅ %s: %d → %d попыток\n", 
                       update.JobKey, result.PreviousRetries, update.NewRetries)
        } else {
            fmt.Printf("❌ %s: ошибка - %s\n", update.JobKey, result.Message)
        }
    }
    
    return results, nil
}

func (bru *BatchRetryUpdater) ResetAllToDefault(jobKeys []string, defaultRetries int32) error {
    updates := make([]JobRetryUpdate, 0, len(jobKeys))
    
    for _, jobKey := range jobKeys {
        updates = append(updates, JobRetryUpdate{
            JobKey:     jobKey,
            NewRetries: defaultRetries,
        })
    }
    
    results, err := bru.UpdateMultipleJobs(updates)
    if err != nil {
        return err
    }
    
    successCount := 0
    for _, result := range results {
        if result.Success {
            successCount++
        }
    }
    
    fmt.Printf("📊 Сброс попыток завершен: %d из %d заданий обновлено\n", 
               successCount, len(jobKeys))
    
    return nil
}

type JobRetryUpdate struct {
    JobKey     string
    NewRetries int32
}

type UpdateResult struct {
    JobKey          string
    Success         bool
    Message         string
    PreviousRetries int32
    NewRetries      int32
    Error           string
}

// Мониторинг и автоматическое управление попытками
type RetryMonitor struct {
    retryManager *RetryManager
    strategy     *AdaptiveRetryStrategy
}

func NewRetryMonitor(client pb.JobsServiceClient, ctx context.Context) *RetryMonitor {
    rm := NewRetryManager(client, ctx)
    return &RetryMonitor{
        retryManager: rm,
        strategy:     NewAdaptiveRetryStrategy(client, ctx),
    }
}

func (monitor *RetryMonitor) MonitorAndOptimize(jobKeys []string) error {
    for _, jobKey := range jobKeys {
        // Получаем информацию о задании
        jobInfo, err := monitor.retryManager.client.GetJob(monitor.retryManager.ctx, &pb.GetJobRequest{
            JobKey: jobKey,
        })
        
        if err != nil {
            fmt.Printf("⚠️ Не удалось получить информацию о задании %s: %v\n", jobKey, err)
            continue
        }
        
        if !jobInfo.Success {
            fmt.Printf("⚠️ Задание %s не найдено\n", jobKey)
            continue
        }
        
        job := jobInfo.Job
        
        // Анализируем статус и историю ошибок
        if job.Retries == 0 && job.State == "FAILED" {
            // Задание исчерпало попытки, возможно стоит дать еще один шанс
            fmt.Printf("🔄 Задание %s исчерпало попытки, даем еще один шанс\n", jobKey)
            monitor.retryManager.SetRetries(jobKey, 1)
        } else if job.Retries > 10 {
            // Слишком много попыток, возможно есть системная проблема
            fmt.Printf("⚠️ Задание %s имеет %d попыток, уменьшаем до разумного лимита\n", 
                       jobKey, job.Retries)
            monitor.retryManager.SetRetries(jobKey, 3)
        }
        
        // Можно добавить дополнительную логику на основе:
        // - Времени создания задания
        // - Типа задания
        // - Истории ошибок
        // - Текущей нагрузки на систему
    }
    
    return nil
}
```

### Python
```python
import grpc
import time
from typing import List, Dict, Optional
from dataclasses import dataclass

import jobs_pb2
import jobs_pb2_grpc

def update_job_retries(job_key, retries):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.UpdateJobRetriesRequest(
        job_key=job_key,
        retries=retries
    )
    
    try:
        response = stub.UpdateJobRetries(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Попытки обновлены для задания {job_key}: {response.previous_retries} → {retries}")
            return True
        else:
            print(f"❌ Ошибка обновления: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

class RetryManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def increase_retries(self, job_key, additional_retries):
        """Увеличивает количество попыток для задания"""
        # Получаем текущую информацию о задании
        job_info = self._get_job_info(job_key)
        if not job_info:
            return False
        
        new_retries = min(job_info['retries'] + additional_retries, 100)
        
        if self._update_retries(job_key, new_retries):
            print(f"🔄 Задание {job_key}: попытки увеличены {job_info['retries']} → {new_retries} (+{additional_retries})")
            return True
        return False
    
    def decrease_retries(self, job_key, removed_retries):
        """Уменьшает количество попыток для задания"""
        job_info = self._get_job_info(job_key)
        if not job_info:
            return False
        
        new_retries = max(job_info['retries'] - removed_retries, 0)
        
        if self._update_retries(job_key, new_retries):
            print(f"⬇️ Задание {job_key}: попытки уменьшены {job_info['retries']} → {new_retries} (-{removed_retries})")
            return True
        return False
    
    def set_retries(self, job_key, retries):
        """Устанавливает конкретное количество попыток"""
        retries = max(0, min(retries, 100))  # Ограничиваем диапазон
        
        job_info = self._get_job_info(job_key)
        if not job_info:
            return False
        
        if self._update_retries(job_key, retries):
            print(f"📝 Задание {job_key}: попытки установлены {job_info['retries']} → {retries}")
            return True
        return False
    
    def reset_retries(self, job_key):
        """Сбрасывает попытки к стандартному значению"""
        return self.set_retries(job_key, 3)
    
    def _get_job_info(self, job_key):
        """Получает информацию о задании"""
        try:
            request = jobs_pb2.GetJobRequest(job_key=job_key)
            response = self.stub.GetJob(request, metadata=self.metadata)
            
            if response.success:
                return {
                    'retries': response.job.retries,
                    'state': response.job.state,
                    'type': response.job.type
                }
            else:
                print(f"❌ Задание не найдено: {response.message}")
                return None
                
        except grpc.RpcError as e:
            print(f"gRPC Error при получении информации о задании: {e.details()}")
            return None
    
    def _update_retries(self, job_key, retries):
        """Внутренний метод для обновления попыток"""
        try:
            request = jobs_pb2.UpdateJobRetriesRequest(
                job_key=job_key,
                retries=retries
            )
            
            response = self.stub.UpdateJobRetries(request, metadata=self.metadata)
            return response.success
            
        except grpc.RpcError as e:
            print(f"gRPC Error при обновлении попыток: {e.details()}")
            return False

class AdaptiveRetryStrategy:
    def __init__(self):
        self.retry_manager = RetryManager()
    
    def handle_error_and_adjust_retries(self, job_key, error_type):
        """Адаптивная стратегия обработки ошибок"""
        strategies = {
            'CONNECTION_ERROR': lambda: self.retry_manager.increase_retries(job_key, 2),
            'RATE_LIMIT': lambda: self.retry_manager.decrease_retries(job_key, 1),
            'AUTH_ERROR': lambda: self.retry_manager.set_retries(job_key, 0),
            'VALIDATION_ERROR': lambda: self.retry_manager.set_retries(job_key, 0),
            'TEMPORARY_ERROR': lambda: True,  # Не изменяем попытки
            'SERVICE_UNAVAILABLE': lambda: self.retry_manager.increase_retries(job_key, 3),
        }
        
        strategy = strategies.get(error_type, lambda: self.retry_manager.reset_retries(job_key))
        return strategy()

@dataclass
class JobRetryUpdate:
    job_key: str
    new_retries: int

@dataclass 
class UpdateResult:
    job_key: str
    success: bool
    message: str = ""
    previous_retries: int = 0
    new_retries: int = 0
    error: str = ""

class BatchRetryUpdater:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def update_multiple_jobs(self, updates: List[JobRetryUpdate]) -> List[UpdateResult]:
        """Обновляет попытки для нескольких заданий"""
        results = []
        
        for update in updates:
            result = UpdateResult(job_key=update.job_key, success=False)
            
            try:
                request = jobs_pb2.UpdateJobRetriesRequest(
                    job_key=update.job_key,
                    retries=update.new_retries
                )
                
                response = self.stub.UpdateJobRetries(request, metadata=self.metadata)
                
                result.success = response.success
                result.message = response.message
                result.previous_retries = response.previous_retries
                result.new_retries = update.new_retries
                
            except grpc.RpcError as e:
                result.error = str(e.details())
            
            results.append(result)
            
            if result.success:
                print(f"✅ {update.job_key}: {result.previous_retries} → {update.new_retries} попыток")
            else:
                print(f"❌ {update.job_key}: ошибка - {result.message or result.error}")
        
        return results
    
    def reset_all_to_default(self, job_keys: List[str], default_retries: int = 3):
        """Сбрасывает все задания к значению по умолчанию"""
        updates = [JobRetryUpdate(job_key=job_key, new_retries=default_retries) 
                  for job_key in job_keys]
        
        results = self.update_multiple_jobs(updates)
        
        success_count = sum(1 for result in results if result.success)
        print(f"📊 Сброс попыток завершен: {success_count} из {len(job_keys)} заданий обновлено")
        
        return results

class RetryMonitor:
    def __init__(self):
        self.retry_manager = RetryManager()
        self.strategy = AdaptiveRetryStrategy()
    
    def monitor_and_optimize(self, job_keys: List[str]):
        """Мониторинг и оптимизация попыток"""
        for job_key in job_keys:
            job_info = self.retry_manager._get_job_info(job_key)
            
            if not job_info:
                print(f"⚠️ Не удалось получить информацию о задании {job_key}")
                continue
            
            retries = job_info['retries']
            state = job_info['state']
            
            # Анализируем и оптимизируем
            if retries == 0 and state == "FAILED":
                print(f"🔄 Задание {job_key} исчерпало попытки, даем еще один шанс")
                self.retry_manager.set_retries(job_key, 1)
            elif retries > 10:
                print(f"⚠️ Задание {job_key} имеет {retries} попыток, уменьшаем до разумного лимита")
                self.retry_manager.set_retries(job_key, 3)

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Использование:")
        print("  python update_job_retries.py <job_key> <retries>")
        print("  python update_job_retries.py test")
        sys.exit(1)
    
    if sys.argv[1] == "test":
        # Тестирование различных сценариев
        retry_manager = RetryManager()
        batch_updater = BatchRetryUpdater()
        monitor = RetryMonitor()
        
        # Тестовые задания
        test_jobs = ["test-job-1", "test-job-2", "test-job-3"]
        
        print("--- Тест увеличения попыток ---")
        for job_key in test_jobs:
            retry_manager.increase_retries(job_key, 2)
        
        print("\n--- Тест группового сброса ---")
        batch_updater.reset_all_to_default(test_jobs)
        
        print("\n--- Тест мониторинга ---")
        monitor.monitor_and_optimize(test_jobs)
    else:
        job_key = sys.argv[1]
        retries = int(sys.argv[2])
        
        update_job_retries(job_key, retries)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function updateJobRetries(jobKey, retries) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_key: jobKey,
            retries: retries
        };
        
        client.updateJobRetries(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Попытки обновлены для задания ${jobKey}: ${response.previous_retries} → ${retries}`);
                resolve(true);
            } else {
                console.log(`❌ Ошибка обновления: ${response.message}`);
                resolve(false);
            }
        });
    });
}

class RetryManager {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async increaseRetries(jobKey, additionalRetries) {
        const jobInfo = await this._getJobInfo(jobKey);
        if (!jobInfo) return false;
        
        const newRetries = Math.min(jobInfo.retries + additionalRetries, 100);
        
        if (await this._updateRetries(jobKey, newRetries)) {
            console.log(`🔄 Задание ${jobKey}: попытки увеличены ${jobInfo.retries} → ${newRetries} (+${additionalRetries})`);
            return true;
        }
        return false;
    }
    
    async decreaseRetries(jobKey, removedRetries) {
        const jobInfo = await this._getJobInfo(jobKey);
        if (!jobInfo) return false;
        
        const newRetries = Math.max(jobInfo.retries - removedRetries, 0);
        
        if (await this._updateRetries(jobKey, newRetries)) {
            console.log(`⬇️ Задание ${jobKey}: попытки уменьшены ${jobInfo.retries} → ${newRetries} (-${removedRetries})`);
            return true;
        }
        return false;
    }
    
    async setRetries(jobKey, retries) {
        retries = Math.max(0, Math.min(retries, 100)); // Ограничиваем диапазон
        
        const jobInfo = await this._getJobInfo(jobKey);
        if (!jobInfo) return false;
        
        if (await this._updateRetries(jobKey, retries)) {
            console.log(`📝 Задание ${jobKey}: попытки установлены ${jobInfo.retries} → ${retries}`);
            return true;
        }
        return false;
    }
    
    async resetRetries(jobKey) {
        return await this.setRetries(jobKey, 3);
    }
    
    async _getJobInfo(jobKey) {
        return new Promise((resolve, reject) => {
            const request = { job_key: jobKey };
            
            this.client.getJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при получении информации о задании: ${error.message}`);
                    resolve(null);
                    return;
                }
                
                if (response.success) {
                    resolve({
                        retries: response.job.retries,
                        state: response.job.state,
                        type: response.job.type
                    });
                } else {
                    console.log(`❌ Задание не найдено: ${response.message}`);
                    resolve(null);
                }
            });
        });
    }
    
    async _updateRetries(jobKey, retries) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries
            };
            
            this.client.updateJobRetries(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при обновлении попыток: ${error.message}`);
                    resolve(false);
                    return;
                }
                
                resolve(response.success);
            });
        });
    }
}

class AdaptiveRetryStrategy {
    constructor() {
        this.retryManager = new RetryManager();
    }
    
    async handleErrorAndAdjustRetries(jobKey, errorType) {
        const strategies = {
            'CONNECTION_ERROR': () => this.retryManager.increaseRetries(jobKey, 2),
            'RATE_LIMIT': () => this.retryManager.decreaseRetries(jobKey, 1),
            'AUTH_ERROR': () => this.retryManager.setRetries(jobKey, 0),
            'VALIDATION_ERROR': () => this.retryManager.setRetries(jobKey, 0),
            'TEMPORARY_ERROR': () => Promise.resolve(true), // Не изменяем попытки
            'SERVICE_UNAVAILABLE': () => this.retryManager.increaseRetries(jobKey, 3),
        };
        
        const strategy = strategies[errorType] || (() => this.retryManager.resetRetries(jobKey));
        return await strategy();
    }
}

class BatchRetryUpdater {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async updateMultipleJobs(updates) {
        const results = [];
        
        for (const update of updates) {
            const result = {
                jobKey: update.jobKey,
                success: false
            };
            
            try {
                const response = await this._updateSingleJob(update.jobKey, update.newRetries);
                
                result.success = response.success;
                result.message = response.message;
                result.previousRetries = response.previous_retries;
                result.newRetries = update.newRetries;
                
            } catch (error) {
                result.error = error.message;
            }
            
            results.push(result);
            
            if (result.success) {
                console.log(`✅ ${update.jobKey}: ${result.previousRetries} → ${update.newRetries} попыток`);
            } else {
                console.log(`❌ ${update.jobKey}: ошибка - ${result.message || result.error}`);
            }
        }
        
        return results;
    }
    
    async resetAllToDefault(jobKeys, defaultRetries = 3) {
        const updates = jobKeys.map(jobKey => ({
            jobKey: jobKey,
            newRetries: defaultRetries
        }));
        
        const results = await this.updateMultipleJobs(updates);
        
        const successCount = results.filter(result => result.success).length;
        console.log(`📊 Сброс попыток завершен: ${successCount} из ${jobKeys.length} заданий обновлено`);
        
        return results;
    }
    
    async _updateSingleJob(jobKey, retries) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries
            };
            
            this.client.updateJobRetries(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                resolve(response);
            });
        });
    }
}

class RetryMonitor {
    constructor() {
        this.retryManager = new RetryManager();
        this.strategy = new AdaptiveRetryStrategy();
    }
    
    async monitorAndOptimize(jobKeys) {
        for (const jobKey of jobKeys) {
            const jobInfo = await this.retryManager._getJobInfo(jobKey);
            
            if (!jobInfo) {
                console.log(`⚠️ Не удалось получить информацию о задании ${jobKey}`);
                continue;
            }
            
            const retries = jobInfo.retries;
            const state = jobInfo.state;
            
            // Анализируем и оптимизируем
            if (retries === 0 && state === "FAILED") {
                console.log(`🔄 Задание ${jobKey} исчерпало попытки, даем еще один шанс`);
                await this.retryManager.setRetries(jobKey, 1);
            } else if (retries > 10) {
                console.log(`⚠️ Задание ${jobKey} имеет ${retries} попыток, уменьшаем до разумного лимита`);
                await this.retryManager.setRetries(jobKey, 3);
            }
        }
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node update-job-retries.js <job_key> <retries>');
        console.log('  node update-job-retries.js test');
        process.exit(1);
    }
    
    if (args[0] === 'test') {
        // Тестирование различных сценариев
        (async () => {
            const retryManager = new RetryManager();
            const batchUpdater = new BatchRetryUpdater();
            const monitor = new RetryMonitor();
            
            // Тестовые задания
            const testJobs = ["test-job-1", "test-job-2", "test-job-3"];
            
            console.log("--- Тест увеличения попыток ---");
            for (const jobKey of testJobs) {
                await retryManager.increaseRetries(jobKey, 2);
            }
            
            console.log("\n--- Тест группового сброса ---");
            await batchUpdater.resetAllToDefault(testJobs);
            
            console.log("\n--- Тест мониторинга ---");
            await monitor.monitorAndOptimize(testJobs);
        })();
    } else {
        const jobKey = args[0];
        const retries = parseInt(args[1]);
        
        updateJobRetries(jobKey, retries).catch(error => {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        });
    }
}

module.exports = {
    updateJobRetries,
    RetryManager,
    AdaptiveRetryStrategy,
    BatchRetryUpdater,
    RetryMonitor
};
```

## Стратегии управления попытками

### Базовые операции
- **Увеличение**: Добавляет дополнительные попытки при необходимости
- **Уменьшение**: Снижает попытки для предотвращения спама
- **Установка**: Точно задает количество попыток
- **Сброс**: Возвращает к стандартному значению

### Адаптивные стратегии
- **CONNECTION_ERROR**: +2 попытки
- **RATE_LIMIT**: -1 попытка
- **AUTH_ERROR**: 0 попыток (бесполезно повторять)
- **VALIDATION_ERROR**: 0 попыток (данные некорректны)
- **SERVICE_UNAVAILABLE**: +3 попытки

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный job_key или значение retries
- `NOT_FOUND` (5): Задание не найдено
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Job 'atom-jobkey12345' not found or already completed",
  "previous_retries": 0
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Получение заданий для выполнения
- [FailJob](fail-job.md) - Провал задания с настройкой попыток
- [GetJob](get-job.md) - Получение информации о задании
- [ListJobs](list-jobs.md) - Список заданий с фильтрацией
