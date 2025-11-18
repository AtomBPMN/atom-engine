# FailJob

## Описание
Сигнализирует о неудачном выполнении задания. Может уменьшить количество попыток или установить обратное ожидание (backoff) перед повторной попыткой.

## Синтаксис
```protobuf
rpc FailJob(FailJobRequest) returns (FailJobResponse);
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

### FailJobRequest
```protobuf
message FailJobRequest {
  string job_key = 1;           // Ключ задания
  int32 retries = 2;            // Новое количество попыток
  string error_message = 3;     // Сообщение об ошибке
  int64 backoff_timeout = 4;    // Таймаут перед повтором (мс)
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания
- **retries** (int32, required): Новое количество оставшихся попыток (обычно текущее значение - 1)
- **error_message** (string, optional): Описание ошибки для диагностики
- **backoff_timeout** (int64, optional): Время ожидания перед повтором в миллисекундах

## Параметры ответа

### FailJobResponse
```protobuf
message FailJobResponse {
  bool success = 1;         // Статус успешности операции
  string message = 2;       // Сообщение о результате
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
    
    // Простое провал задания
    response, err := client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:       jobKey,
        Retries:      2,  // Уменьшаем на 1
        ErrorMessage: "Connection timeout",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Задание %s провалено успешно\n", jobKey)
    } else {
        fmt.Printf("Ошибка провала: %s\n", response.Message)
    }
}

// Провал с экспоненциальным backoff
func failJobWithBackoff(client pb.JobsServiceClient, ctx context.Context, 
    jobKey string, currentRetries int32, attempt int, baseDelay time.Duration) error {
    
    // Экспоненциальный backoff: baseDelay * 2^attempt
    backoffMs := int64(baseDelay.Milliseconds()) * (1 << attempt)
    
    // Максимальный backoff 5 минут
    maxBackoffMs := int64(5 * 60 * 1000)
    if backoffMs > maxBackoffMs {
        backoffMs = maxBackoffMs
    }
    
    response, err := client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:         jobKey,
        Retries:        currentRetries - 1,
        ErrorMessage:   fmt.Sprintf("Attempt %d failed, will retry after %dms", attempt, backoffMs),
        BackoffTimeout: backoffMs,
    })
    
    if err != nil {
        return fmt.Errorf("ошибка провала задания: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("провал не выполнен: %s", response.Message)
    }
    
    fmt.Printf("⏰ Задание %s будет повторено через %s\n", 
        jobKey, time.Duration(backoffMs)*time.Millisecond)
    
    return nil
}

// Обработчик ошибок с различными стратегиями
type ErrorHandler struct {
    client pb.JobsServiceClient
}

func (h *ErrorHandler) HandleJobError(ctx context.Context, jobKey string, 
    retries int32, err error) error {
    
    errorMsg := err.Error()
    
    switch {
    case isRetryableError(err):
        return h.retryableError(ctx, jobKey, retries, errorMsg)
    case isRateLimitError(err):
        return h.rateLimitError(ctx, jobKey, retries, errorMsg)
    case isTemporaryError(err):
        return h.temporaryError(ctx, jobKey, retries, errorMsg)
    default:
        return h.permanentError(ctx, jobKey, errorMsg)
    }
}

func (h *ErrorHandler) retryableError(ctx context.Context, jobKey string, 
    retries int32, errorMsg string) error {
    
    if retries <= 0 {
        return h.permanentError(ctx, jobKey, "Max retries exceeded: "+errorMsg)
    }
    
    // Быстрый повтор для обычных ошибок
    response, err := h.client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:         jobKey,
        Retries:        retries - 1,
        ErrorMessage:   errorMsg,
        BackoffTimeout: 1000, // 1 секунда
    })
    
    if err != nil || !response.Success {
        return fmt.Errorf("не удалось запланировать повтор: %v", err)
    }
    
    fmt.Printf("🔄 Задание %s будет повторено через 1с (осталось попыток: %d)\n", 
        jobKey, retries-1)
    return nil
}

func (h *ErrorHandler) rateLimitError(ctx context.Context, jobKey string, 
    retries int32, errorMsg string) error {
    
    if retries <= 0 {
        return h.permanentError(ctx, jobKey, "Rate limit exceeded, no retries left")
    }
    
    // Длительный backoff для rate limit
    response, err := h.client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:         jobKey,
        Retries:        retries - 1,
        ErrorMessage:   "Rate limit: " + errorMsg,
        BackoffTimeout: 60000, // 1 минута
    })
    
    if err != nil || !response.Success {
        return fmt.Errorf("не удалось обработать rate limit: %v", err)
    }
    
    fmt.Printf("⏸️ Rate limit: задание %s будет повторено через 1 мин\n", jobKey)
    return nil
}

func (h *ErrorHandler) temporaryError(ctx context.Context, jobKey string, 
    retries int32, errorMsg string) error {
    
    if retries <= 0 {
        return h.permanentError(ctx, jobKey, "Temporary error, no retries left")
    }
    
    // Средний backoff для временных ошибок
    response, err := h.client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:         jobKey,
        Retries:        retries - 1,
        ErrorMessage:   "Temporary: " + errorMsg,
        BackoffTimeout: 10000, // 10 секунд
    })
    
    if err != nil || !response.Success {
        return fmt.Errorf("не удалось обработать временную ошибку: %v", err)
    }
    
    fmt.Printf("⏳ Временная ошибка: задание %s повтор через 10с\n", jobKey)
    return nil
}

func (h *ErrorHandler) permanentError(ctx context.Context, jobKey, errorMsg string) error {
    // Для постоянных ошибок устанавливаем retries = 0
    response, err := h.client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:       jobKey,
        Retries:      0,
        ErrorMessage: "Permanent error: " + errorMsg,
    })
    
    if err != nil || !response.Success {
        return fmt.Errorf("не удалось зафиксировать постоянную ошибку: %v", err)
    }
    
    fmt.Printf("❌ Постоянная ошибка: задание %s не будет повторено\n", jobKey)
    return nil
}

// Классификация ошибок
func isRetryableError(err error) bool {
    errMsg := err.Error()
    retryableErrors := []string{
        "connection refused",
        "timeout",
        "network unreachable",
        "temporary failure",
    }
    
    for _, retryable := range retryableErrors {
        if contains(errMsg, retryable) {
            return true
        }
    }
    return false
}

func isRateLimitError(err error) bool {
    errMsg := err.Error()
    return contains(errMsg, "rate limit") || contains(errMsg, "too many requests")
}

func isTemporaryError(err error) bool {
    errMsg := err.Error()
    return contains(errMsg, "service unavailable") || contains(errMsg, "server error")
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && 
           (s == substr || 
            (len(s) > len(substr) && 
             (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
              strings.Contains(s, substr))))
}

// Пример использования в воркере
func processJobWithErrorHandling(client pb.JobsServiceClient, ctx context.Context, job *pb.ActivatedJob) {
    errorHandler := &ErrorHandler{client: client}
    
    // Попытка выполнить задание
    err := performJobWork(job)
    
    if err != nil {
        fmt.Printf("⚠️ Ошибка выполнения задания %s: %v\n", job.JobKey, err)
        
        // Обрабатываем ошибку
        handleErr := errorHandler.HandleJobError(ctx, job.JobKey, job.Retries, err)
        if handleErr != nil {
            log.Printf("Критическая ошибка обработки: %v", handleErr)
        }
        return
    }
    
    // Успешное завершение
    variables := map[string]string{
        "result": "success",
        "completedAt": time.Now().Format(time.RFC3339),
    }
    
    _, err = client.CompleteJob(ctx, &pb.CompleteJobRequest{
        JobKey:    job.JobKey,
        Variables: variables,
    })
    
    if err != nil {
        log.Printf("Ошибка завершения задания: %v", err)
    }
}

func performJobWork(job *pb.ActivatedJob) error {
    // Имитация работы с возможными ошибками
    time.Sleep(100 * time.Millisecond)
    
    // Различные типы ошибок для демонстрации
    switch job.JobKey {
    case "fail-retryable":
        return fmt.Errorf("connection refused")
    case "fail-rate-limit":
        return fmt.Errorf("rate limit exceeded")
    case "fail-temporary":
        return fmt.Errorf("service unavailable")
    case "fail-permanent":
        return fmt.Errorf("invalid configuration")
    default:
        return nil // Успех
    }
}
```

### Python
```python
import grpc
import time
import math
from enum import Enum

import jobs_pb2
import jobs_pb2_grpc

class ErrorType(Enum):
    RETRYABLE = "retryable"
    RATE_LIMIT = "rate_limit"
    TEMPORARY = "temporary"
    PERMANENT = "permanent"

def fail_job(job_key, retries, error_message, backoff_timeout=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.FailJobRequest(
        job_key=job_key,
        retries=retries,
        error_message=error_message
    )
    
    if backoff_timeout is not None:
        request.backoff_timeout = backoff_timeout
    
    try:
        response = stub.FailJob(request, metadata=metadata)
        
        if response.success:
            print(f"⚠️ Задание {job_key} провалено: {error_message}")
            if backoff_timeout:
                print(f"   Повтор через {backoff_timeout/1000:.1f}с")
            return True
        else:
            print(f"❌ Ошибка провала: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

class ErrorHandler:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def handle_job_error(self, job_key, retries, error, attempt=1):
        """Обработка ошибки задания с соответствующей стратегией"""
        error_type = self.classify_error(error)
        error_message = str(error)
        
        if error_type == ErrorType.RETRYABLE:
            return self.handle_retryable_error(job_key, retries, error_message, attempt)
        elif error_type == ErrorType.RATE_LIMIT:
            return self.handle_rate_limit_error(job_key, retries, error_message)
        elif error_type == ErrorType.TEMPORARY:
            return self.handle_temporary_error(job_key, retries, error_message)
        else:  # PERMANENT
            return self.handle_permanent_error(job_key, error_message)
    
    def classify_error(self, error):
        """Классификация ошибки по типу"""
        error_str = str(error).lower()
        
        if any(keyword in error_str for keyword in 
               ['connection refused', 'timeout', 'network unreachable']):
            return ErrorType.RETRYABLE
        elif any(keyword in error_str for keyword in 
                ['rate limit', 'too many requests']):
            return ErrorType.RATE_LIMIT
        elif any(keyword in error_str for keyword in 
                ['service unavailable', 'server error']):
            return ErrorType.TEMPORARY
        else:
            return ErrorType.PERMANENT
    
    def handle_retryable_error(self, job_key, retries, error_message, attempt):
        """Обработка повторяемой ошибки с экспоненциальным backoff"""
        if retries <= 0:
            return self.handle_permanent_error(job_key, f"Max retries exceeded: {error_message}")
        
        # Экспоненциальный backoff: 1s * 2^attempt
        backoff_ms = min(1000 * (2 ** attempt), 300000)  # Максимум 5 минут
        
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=retries - 1,
            error_message=f"Attempt {attempt}: {error_message}",
            backoff_timeout=backoff_ms
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"🔄 Задание {job_key} будет повторено через {backoff_ms/1000:.1f}с "
                      f"(попытка {attempt}, осталось: {retries-1})")
                return True
            else:
                print(f"❌ Не удалось запланировать повтор: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при обработке повторяемой ошибки: {e.details()}")
            return False
    
    def handle_rate_limit_error(self, job_key, retries, error_message):
        """Обработка ошибки rate limit"""
        if retries <= 0:
            return self.handle_permanent_error(job_key, "Rate limit exceeded, no retries left")
        
        # Длительное ожидание для rate limit
        backoff_ms = 60000  # 1 минута
        
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=retries - 1,
            error_message=f"Rate limit: {error_message}",
            backoff_timeout=backoff_ms
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"⏸️ Rate limit: задание {job_key} будет повторено через 1 минуту")
                return True
            else:
                print(f"❌ Не удалось обработать rate limit: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при обработке rate limit: {e.details()}")
            return False
    
    def handle_temporary_error(self, job_key, retries, error_message):
        """Обработка временной ошибки"""
        if retries <= 0:
            return self.handle_permanent_error(job_key, "Temporary error, no retries left")
        
        # Средний backoff для временных ошибок
        backoff_ms = 10000  # 10 секунд
        
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=retries - 1,
            error_message=f"Temporary: {error_message}",
            backoff_timeout=backoff_ms
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"⏳ Временная ошибка: задание {job_key} повтор через 10с")
                return True
            else:
                print(f"❌ Не удалось обработать временную ошибку: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при обработке временной ошибки: {e.details()}")
            return False
    
    def handle_permanent_error(self, job_key, error_message):
        """Обработка постоянной ошибки (без повторов)"""
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=0,
            error_message=f"Permanent error: {error_message}"
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"❌ Постоянная ошибка: задание {job_key} не будет повторено")
                return True
            else:
                print(f"❌ Не удалось зафиксировать постоянную ошибку: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при обработке постоянной ошибки: {e.details()}")
            return False

def process_job_with_error_handling(job):
    """Обработка задания с автоматической обработкой ошибок"""
    error_handler = ErrorHandler()
    
    try:
        # Имитация выполнения задания
        result = perform_job_work(job)
        
        if result['success']:
            # Завершаем задание успешно
            complete_job(job['job_key'], result['variables'])
        else:
            # Обрабатываем ошибку
            error_handler.handle_job_error(
                job['job_key'], 
                job['retries'], 
                Exception(result['error'])
            )
            
    except Exception as e:
        print(f"⚠️ Исключение при выполнении задания {job['job_key']}: {e}")
        error_handler.handle_job_error(job['job_key'], job['retries'], e)

def perform_job_work(job):
    """Имитация выполнения задания с возможными ошибками"""
    import random
    
    # Имитация времени выполнения
    time.sleep(0.1)
    
    # Различные исходы для демонстрации
    if job['job_key'] == 'fail-retryable':
        return {'success': False, 'error': 'Connection refused'}
    elif job['job_key'] == 'fail-rate-limit':
        return {'success': False, 'error': 'Rate limit exceeded'}
    elif job['job_key'] == 'fail-temporary':
        return {'success': False, 'error': 'Service unavailable'}
    elif job['job_key'] == 'fail-permanent':
        return {'success': False, 'error': 'Invalid configuration'}
    else:
        # Случайный исход
        if random.random() < 0.8:  # 80% успех
            return {
                'success': True, 
                'variables': {
                    'result': 'success',
                    'processed_at': time.strftime('%Y-%m-%dT%H:%M:%SZ')
                }
            }
        else:
            # 20% ошибка
            return {'success': False, 'error': 'Random failure for testing'}

def complete_job(job_key, variables):
    """Завершение задания"""
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.CompleteJobRequest(
        job_key=job_key,
        variables=variables
    )
    
    try:
        response = stub.CompleteJob(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Задание {job_key} успешно завершено")
        else:
            print(f"❌ Ошибка завершения: {response.message}")
            
    except grpc.RpcError as e:
        print(f"gRPC Error при завершении: {e.details()}")

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 4:
        print("Использование:")
        print("  python fail_job.py <job_key> <retries> <error_message> [backoff_ms]")
        print("  python fail_job.py test")
        sys.exit(1)
    
    if sys.argv[1] == "test":
        # Тестирование различных типов ошибок
        test_jobs = [
            {'job_key': 'success-job', 'retries': 3},
            {'job_key': 'fail-retryable', 'retries': 3},
            {'job_key': 'fail-rate-limit', 'retries': 2},
            {'job_key': 'fail-temporary', 'retries': 1},
            {'job_key': 'fail-permanent', 'retries': 3},
        ]
        
        for job in test_jobs:
            print(f"\n--- Тестирование задания {job['job_key']} ---")
            process_job_with_error_handling(job)
    else:
        job_key = sys.argv[1]
        retries = int(sys.argv[2])
        error_message = sys.argv[3]
        backoff_ms = int(sys.argv[4]) if len(sys.argv) > 4 else None
        
        fail_job(job_key, retries, error_message, backoff_ms)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function failJob(jobKey, retries, errorMessage, backoffTimeout = null) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_key: jobKey,
            retries: retries,
            error_message: errorMessage
        };
        
        if (backoffTimeout !== null) {
            request.backoff_timeout = backoffTimeout;
        }
        
        client.failJob(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`⚠️ Задание ${jobKey} провалено: ${errorMessage}`);
                if (backoffTimeout) {
                    console.log(`   Повтор через ${backoffTimeout/1000}с`);
                }
                resolve(true);
            } else {
                console.log(`❌ Ошибка провала: ${response.message}`);
                resolve(false);
            }
        });
    });
}

class ErrorHandler {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async handleJobError(jobKey, retries, error, attempt = 1) {
        const errorType = this.classifyError(error);
        const errorMessage = error.message || error.toString();
        
        switch (errorType) {
            case 'retryable':
                return await this.handleRetryableError(jobKey, retries, errorMessage, attempt);
            case 'rate_limit':
                return await this.handleRateLimitError(jobKey, retries, errorMessage);
            case 'temporary':
                return await this.handleTemporaryError(jobKey, retries, errorMessage);
            default: // permanent
                return await this.handlePermanentError(jobKey, errorMessage);
        }
    }
    
    classifyError(error) {
        const errorStr = (error.message || error.toString()).toLowerCase();
        
        if (['connection refused', 'timeout', 'network unreachable'].some(keyword => 
            errorStr.includes(keyword))) {
            return 'retryable';
        }
        
        if (['rate limit', 'too many requests'].some(keyword => 
            errorStr.includes(keyword))) {
            return 'rate_limit';
        }
        
        if (['service unavailable', 'server error'].some(keyword => 
            errorStr.includes(keyword))) {
            return 'temporary';
        }
        
        return 'permanent';
    }
    
    async handleRetryableError(jobKey, retries, errorMessage, attempt) {
        if (retries <= 0) {
            return await this.handlePermanentError(jobKey, `Max retries exceeded: ${errorMessage}`);
        }
        
        // Экспоненциальный backoff: 1s * 2^attempt, максимум 5 минут
        const backoffMs = Math.min(1000 * Math.pow(2, attempt), 300000);
        
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries - 1,
                error_message: `Attempt ${attempt}: ${errorMessage}`,
                backoff_timeout: backoffMs
            };
            
            this.client.failJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при обработке повторяемой ошибки: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`🔄 Задание ${jobKey} будет повторено через ${backoffMs/1000}с ` +
                               `(попытка ${attempt}, осталось: ${retries-1})`);
                    resolve(true);
                } else {
                    console.log(`❌ Не удалось запланировать повтор: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    async handleRateLimitError(jobKey, retries, errorMessage) {
        if (retries <= 0) {
            return await this.handlePermanentError(jobKey, "Rate limit exceeded, no retries left");
        }
        
        const backoffMs = 60000; // 1 минута
        
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries - 1,
                error_message: `Rate limit: ${errorMessage}`,
                backoff_timeout: backoffMs
            };
            
            this.client.failJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при обработке rate limit: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`⏸️ Rate limit: задание ${jobKey} будет повторено через 1 минуту`);
                    resolve(true);
                } else {
                    console.log(`❌ Не удалось обработать rate limit: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    async handleTemporaryError(jobKey, retries, errorMessage) {
        if (retries <= 0) {
            return await this.handlePermanentError(jobKey, "Temporary error, no retries left");
        }
        
        const backoffMs = 10000; // 10 секунд
        
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries - 1,
                error_message: `Temporary: ${errorMessage}`,
                backoff_timeout: backoffMs
            };
            
            this.client.failJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при обработке временной ошибки: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`⏳ Временная ошибка: задание ${jobKey} повтор через 10с`);
                    resolve(true);
                } else {
                    console.log(`❌ Не удалось обработать временную ошибку: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    async handlePermanentError(jobKey, errorMessage) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: 0,
                error_message: `Permanent error: ${errorMessage}`
            };
            
            this.client.failJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при обработке постоянной ошибки: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`❌ Постоянная ошибка: задание ${jobKey} не будет повторено`);
                    resolve(true);
                } else {
                    console.log(`❌ Не удалось зафиксировать постоянную ошибку: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
}

async function processJobWithErrorHandling(job) {
    const errorHandler = new ErrorHandler();
    
    try {
        // Имитация выполнения задания
        const result = await performJobWork(job);
        
        if (result.success) {
            // Завершаем задание успешно
            await completeJob(job.job_key, result.variables);
        } else {
            // Обрабатываем ошибку
            await errorHandler.handleJobError(
                job.job_key, 
                job.retries, 
                new Error(result.error)
            );
        }
        
    } catch (error) {
        console.log(`⚠️ Исключение при выполнении задания ${job.job_key}: ${error.message}`);
        await errorHandler.handleJobError(job.job_key, job.retries, error);
    }
}

async function performJobWork(job) {
    // Имитация времени выполнения
    await new Promise(resolve => setTimeout(resolve, 100));
    
    // Различные исходы для демонстрации
    switch (job.job_key) {
        case 'fail-retryable':
            return { success: false, error: 'Connection refused' };
        case 'fail-rate-limit':
            return { success: false, error: 'Rate limit exceeded' };
        case 'fail-temporary':
            return { success: false, error: 'Service unavailable' };
        case 'fail-permanent':
            return { success: false, error: 'Invalid configuration' };
        default:
            // Случайный исход (80% успех)
            if (Math.random() < 0.8) {
                return {
                    success: true,
                    variables: {
                        result: 'success',
                        processed_at: new Date().toISOString()
                    }
                };
            } else {
                return { success: false, error: 'Random failure for testing' };
            }
    }
}

async function completeJob(jobKey, variables) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_key: jobKey,
            variables: variables
        };
        
        client.completeJob(request, metadata, (error, response) => {
            if (error) {
                console.error(`gRPC Error при завершении: ${error.message}`);
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Задание ${jobKey} успешно завершено`);
                resolve(true);
            } else {
                console.log(`❌ Ошибка завершения: ${response.message}`);
                resolve(false);
            }
        });
    });
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node fail-job.js <job_key> <retries> <error_message> [backoff_ms]');
        console.log('  node fail-job.js test');
        process.exit(1);
    }
    
    if (args[0] === 'test') {
        // Тестирование различных типов ошибок
        const testJobs = [
            { job_key: 'success-job', retries: 3 },
            { job_key: 'fail-retryable', retries: 3 },
            { job_key: 'fail-rate-limit', retries: 2 },
            { job_key: 'fail-temporary', retries: 1 },
            { job_key: 'fail-permanent', retries: 3 },
        ];
        
        (async () => {
            for (const job of testJobs) {
                console.log(`\n--- Тестирование задания ${job.job_key} ---`);
                await processJobWithErrorHandling(job);
            }
        })();
    } else {
        const jobKey = args[0];
        const retries = parseInt(args[1]);
        const errorMessage = args[2];
        const backoffMs = args[3] ? parseInt(args[3]) : null;
        
        failJob(jobKey, retries, errorMessage, backoffMs).catch(error => {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        });
    }
}

module.exports = {
    failJob,
    ErrorHandler,
    processJobWithErrorHandling
};
```

## Стратегии обработки ошибок

### Экспоненциальный Backoff
```go
backoffMs := int64(baseDelay.Milliseconds()) * (1 << attempt)
if backoffMs > maxBackoffMs {
    backoffMs = maxBackoffMs
}
```

### Circuit Breaker Pattern
```python
class CircuitBreaker:
    def __init__(self, failure_threshold=5, recovery_timeout=60):
        self.failure_count = 0
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.last_failure_time = None
        self.state = "CLOSED"  # CLOSED, OPEN, HALF_OPEN
    
    def should_fail_fast(self):
        if self.state == "OPEN":
            if time.time() - self.last_failure_time > self.recovery_timeout:
                self.state = "HALF_OPEN"
                return False
            return True
        return False
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный job_key или параметры
- `NOT_FOUND` (5): Задание не найдено
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Job 'atom-jobkey12345' not found or already completed"
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Получение заданий для выполнения
- [CompleteJob](complete-job.md) - Успешное завершение задания
- [UpdateJobRetries](update-job-retries.md) - Обновление количества попыток
- [GetJob](get-job.md) - Получение деталей задания
