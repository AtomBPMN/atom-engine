# CompleteJob

## Описание
Завершает выполнение задания и передает результат обратно в процесс. Используется воркерами для сигнализации успешного выполнения задания.

## Синтаксис
```protobuf
rpc CompleteJob(CompleteJobRequest) returns (CompleteJobResponse);
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

### CompleteJobRequest
```protobuf
message CompleteJobRequest {
  string job_key = 1;                    // Ключ задания
  map<string, string> variables = 2;     // Переменные результата
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания, полученный при активации
- **variables** (map<string, string>, optional): Переменные результата выполнения задания

## Параметры ответа

### CompleteJobResponse
```protobuf
message CompleteJobResponse {
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
    "encoding/json"
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
    
    // Завершение задания с результатом
    variables := map[string]string{
        "result":        "success",
        "responseCode":  "200",
        "responseData":  `{"status": "processed", "id": "12345"}`,
        "executionTime": "1.5s",
    }
    
    response, err := client.CompleteJob(ctx, &pb.CompleteJobRequest{
        JobKey:    jobKey,
        Variables: variables,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("Задание %s успешно завершено\n", jobKey)
        fmt.Printf("Сообщение: %s\n", response.Message)
    } else {
        fmt.Printf("Ошибка завершения: %s\n", response.Message)
    }
}

// Воркер для обработки HTTP заданий
func httpWorker(client pb.JobsServiceClient, ctx context.Context) {
    for {
        // Активируем задания
        activateResponse, err := client.ActivateJobs(ctx, &pb.ActivateJobsRequest{
            Type:         "http-request",
            Worker:       "http-worker-1",
            MaxJobs:      5,
            Timeout:      30000, // 30 секунд
        })
        
        if err != nil {
            log.Printf("Ошибка активации: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        if !activateResponse.Success {
            log.Printf("Активация не удалась: %s", activateResponse.Message)
            time.Sleep(5 * time.Second)
            continue
        }
        
        // Обрабатываем каждое задание
        for _, job := range activateResponse.Jobs {
            go processHTTPJob(client, ctx, job)
        }
        
        // Если нет заданий, ждем
        if len(activateResponse.Jobs) == 0 {
            time.Sleep(1 * time.Second)
        }
    }
}

func processHTTPJob(client pb.JobsServiceClient, ctx context.Context, job *pb.JobInfo) {
    fmt.Printf("Обработка HTTP задания: %s\n", job.JobKey)
    
    // Извлекаем параметры задания
    url := job.Variables["url"]
    method := job.Variables["method"]
    if method == "" {
        method = "GET"
    }
    
    // Выполняем HTTP запрос (имитация)
    result, err := executeHTTPRequest(url, method, job.Variables)
    
    if err != nil {
        // Задание провалилось
        failResponse, failErr := client.FailJob(ctx, &pb.FailJobRequest{
            JobKey:   job.JobKey,
            Retries:  job.Retries - 1,
            ErrorMessage: err.Error(),
        })
        
        if failErr != nil {
            log.Printf("Ошибка провала задания: %v", failErr)
        } else if failResponse.Success {
            fmt.Printf("Задание %s провалено: %s\n", job.JobKey, err.Error())
        }
        return
    }
    
    // Задание выполнено успешно
    variables := map[string]string{
        "result":        "success",
        "responseCode":  fmt.Sprintf("%d", result.StatusCode),
        "responseData":  result.Body,
        "executionTime": result.Duration.String(),
        "completedAt":   time.Now().Format(time.RFC3339),
    }
    
    completeResponse, err := client.CompleteJob(ctx, &pb.CompleteJobRequest{
        JobKey:    job.JobKey,
        Variables: variables,
    })
    
    if err != nil {
        log.Printf("Ошибка завершения задания: %v", err)
        return
    }
    
    if completeResponse.Success {
        fmt.Printf("✅ Задание %s успешно завершено\n", job.JobKey)
    } else {
        fmt.Printf("❌ Ошибка завершения: %s\n", completeResponse.Message)
    }
}

type HTTPResult struct {
    StatusCode int
    Body       string
    Duration   time.Duration
}

func executeHTTPRequest(url, method string, params map[string]string) (*HTTPResult, error) {
    start := time.Now()
    
    // Имитация HTTP запроса
    time.Sleep(100 * time.Millisecond) // Симуляция обработки
    
    // Простая проверка URL
    if url == "" {
        return nil, fmt.Errorf("URL не указан")
    }
    
    if url == "http://error.example.com" {
        return nil, fmt.Errorf("connection refused")
    }
    
    duration := time.Since(start)
    
    // Имитация успешного ответа
    result := &HTTPResult{
        StatusCode: 200,
        Body:       `{"message": "Request processed successfully", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`,
        Duration:   duration,
    }
    
    return result, nil
}

// Воркер для обработки заданий валидации данных
func validationWorker(client pb.JobsServiceClient, ctx context.Context) {
    for {
        activateResponse, err := client.ActivateJobs(ctx, &pb.ActivateJobsRequest{
            Type:         "data-validation",
            Worker:       "validation-worker-1",
            MaxJobs:      10,
            Timeout:      15000, // 15 секунд
        })
        
        if err != nil {
            log.Printf("Ошибка активации валидации: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        if !activateResponse.Success || len(activateResponse.Jobs) == 0 {
            time.Sleep(2 * time.Second)
            continue
        }
        
        for _, job := range activateResponse.Jobs {
            go processValidationJob(client, ctx, job)
        }
    }
}

func processValidationJob(client pb.JobsServiceClient, ctx context.Context, job *pb.JobInfo) {
    fmt.Printf("Валидация данных: %s\n", job.JobKey)
    
    // Извлекаем данные для валидации
    dataToValidate := job.Variables["data"]
    schema := job.Variables["schema"]
    
    // Выполняем валидацию
    validationResult, err := validateData(dataToValidate, schema)
    
    if err != nil {
        // Ошибка валидации
        failResponse, _ := client.FailJob(ctx, &pb.FailJobRequest{
            JobKey:       job.JobKey,
            Retries:      job.Retries - 1,
            ErrorMessage: fmt.Sprintf("Validation error: %v", err),
        })
        
        if failResponse != nil && failResponse.Success {
            fmt.Printf("❌ Валидация провалена: %s\n", err.Error())
        }
        return
    }
    
    // Валидация успешна
    variables := map[string]string{
        "validationResult": validationResult.Status,
        "isValid":          fmt.Sprintf("%t", validationResult.IsValid),
        "errors":           strings.Join(validationResult.Errors, ","),
        "warnings":         strings.Join(validationResult.Warnings, ","),
        "validatedAt":      time.Now().Format(time.RFC3339),
    }
    
    completeResponse, err := client.CompleteJob(ctx, &pb.CompleteJobRequest{
        JobKey:    job.JobKey,
        Variables: variables,
    })
    
    if err != nil {
        log.Printf("Ошибка завершения валидации: %v", err)
        return
    }
    
    if completeResponse.Success {
        if validationResult.IsValid {
            fmt.Printf("✅ Валидация %s прошла успешно\n", job.JobKey)
        } else {
            fmt.Printf("⚠️ Валидация %s завершена с ошибками\n", job.JobKey)
        }
    }
}

type ValidationResult struct {
    Status    string
    IsValid   bool
    Errors    []string
    Warnings  []string
}

func validateData(data, schema string) (*ValidationResult, error) {
    if data == "" {
        return nil, fmt.Errorf("no data provided")
    }
    
    // Простая валидация JSON
    var jsonData interface{}
    err := json.Unmarshal([]byte(data), &jsonData)
    
    result := &ValidationResult{
        Status:   "completed",
        IsValid:  err == nil,
        Errors:   []string{},
        Warnings: []string{},
    }
    
    if err != nil {
        result.Errors = append(result.Errors, "Invalid JSON format")
    }
    
    // Дополнительные проверки на основе схемы
    if schema != "" && err == nil {
        // Здесь была бы реальная валидация по схеме
        result.Warnings = append(result.Warnings, "Schema validation available via expression component")
    }
    
    return result, nil
}
```

### Python
```python
import grpc
import json
import time
import requests
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime

import jobs_pb2
import jobs_pb2_grpc

def complete_job(job_key, variables=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    if variables is None:
        variables = {}
    
    request = jobs_pb2.CompleteJobRequest(
        job_key=job_key,
        variables=variables
    )
    
    try:
        response = stub.CompleteJob(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Задание {job_key} успешно завершено")
            return True
        else:
            print(f"❌ Ошибка завершения: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

class JobWorker:
    def __init__(self, worker_name, job_type, max_jobs=5):
        self.worker_name = worker_name
        self.job_type = job_type
        self.max_jobs = max_jobs
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
        self.running = False
    
    def start(self):
        """Запуск воркера"""
        print(f"Запуск воркера {self.worker_name} для заданий типа {self.job_type}")
        self.running = True
        
        while self.running:
            try:
                # Активируем задания
                jobs = self.activate_jobs()
                
                if jobs:
                    print(f"Получено {len(jobs)} заданий")
                    
                    # Обрабатываем задания параллельно
                    with ThreadPoolExecutor(max_workers=self.max_jobs) as executor:
                        futures = [executor.submit(self.process_job, job) for job in jobs]
                        
                        # Ждем завершения всех заданий
                        for future in futures:
                            try:
                                future.result()
                            except Exception as e:
                                print(f"Ошибка обработки задания: {e}")
                else:
                    # Нет заданий, ждем
                    time.sleep(1)
                    
            except KeyboardInterrupt:
                print("Остановка воркера...")
                break
            except Exception as e:
                print(f"Ошибка воркера: {e}")
                time.sleep(5)
        
        self.running = False
        print(f"Воркер {self.worker_name} остановлен")
    
    def stop(self):
        """Остановка воркера"""
        self.running = False
    
    def activate_jobs(self):
        """Активация заданий"""
        request = jobs_pb2.ActivateJobsRequest(
            type=self.job_type,
            worker=self.worker_name,
            max_jobs=self.max_jobs,
            timeout=30000  # 30 секунд
        )
        
        try:
            response = self.stub.ActivateJobs(request, metadata=self.metadata)
            
            if response.success:
                return response.jobs
            else:
                print(f"Активация не удалась: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"Ошибка активации: {e.details()}")
            return []
    
    def process_job(self, job):
        """Обработка одного задания (переопределяется в подклассах)"""
        print(f"Обработка задания {job.job_key}")
        
        # Базовая обработка - просто завершаем задание
        variables = {
            'result': 'completed',
            'worker': self.worker_name,
            'processed_at': datetime.now().isoformat()
        }
        
        return self.complete_job(job.job_key, variables)
    
    def complete_job(self, job_key, variables):
        """Завершение задания"""
        request = jobs_pb2.CompleteJobRequest(
            job_key=job_key,
            variables=variables
        )
        
        try:
            response = self.stub.CompleteJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"✅ Задание {job_key} завершено")
                return True
            else:
                print(f"❌ Ошибка завершения {job_key}: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC ошибка завершения {job_key}: {e.details()}")
            return False
    
    def fail_job(self, job_key, retries, error_message):
        """Провал задания"""
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=retries,
            error_message=error_message
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"⚠️ Задание {job_key} провалено: {error_message}")
                return True
            else:
                print(f"❌ Ошибка провала {job_key}: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC ошибка провала {job_key}: {e.details()}")
            return False

class HTTPWorker(JobWorker):
    def __init__(self, worker_name="http-worker"):
        super().__init__(worker_name, "http-request", max_jobs=3)
    
    def process_job(self, job):
        """Обработка HTTP задания"""
        print(f"HTTP запрос: {job.job_key}")
        
        try:
            # Извлекаем параметры
            url = job.variables.get('url', '')
            method = job.variables.get('method', 'GET')
            headers = self.parse_headers(job.variables.get('headers', '{}'))
            timeout = int(job.variables.get('timeout', '30'))
            
            # Выполняем HTTP запрос
            start_time = time.time()
            
            if method.upper() == 'GET':
                response = requests.get(url, headers=headers, timeout=timeout)
            elif method.upper() == 'POST':
                data = job.variables.get('body', '')
                response = requests.post(url, headers=headers, data=data, timeout=timeout)
            else:
                raise ValueError(f"Неподдерживаемый HTTP метод: {method}")
            
            execution_time = time.time() - start_time
            
            # Готовим результат
            variables = {
                'result': 'success',
                'status_code': str(response.status_code),
                'response_body': response.text[:1000],  # Ограничиваем размер
                'execution_time': f"{execution_time:.3f}s",
                'response_headers': json.dumps(dict(response.headers)),
                'completed_at': datetime.now().isoformat()
            }
            
            return self.complete_job(job.job_key, variables)
            
        except requests.exceptions.Timeout:
            return self.fail_job(job.job_key, job.retries - 1, "Request timeout")
        except requests.exceptions.ConnectionError:
            return self.fail_job(job.job_key, job.retries - 1, "Connection error")
        except Exception as e:
            return self.fail_job(job.job_key, job.retries - 1, str(e))
    
    def parse_headers(self, headers_str):
        """Парсинг заголовков из строки JSON"""
        try:
            return json.loads(headers_str)
        except:
            return {}

class ValidationWorker(JobWorker):
    def __init__(self, worker_name="validation-worker"):
        super().__init__(worker_name, "data-validation", max_jobs=10)
    
    def process_job(self, job):
        """Обработка задания валидации данных"""
        print(f"Валидация данных: {job.job_key}")
        
        try:
            # Извлекаем данные
            data = job.variables.get('data', '')
            schema = job.variables.get('schema', '')
            validation_type = job.variables.get('validation_type', 'json')
            
            # Выполняем валидацию
            result = self.validate_data(data, schema, validation_type)
            
            # Готовим результат
            variables = {
                'validation_result': result['status'],
                'is_valid': str(result['is_valid']).lower(),
                'errors': json.dumps(result['errors']),
                'warnings': json.dumps(result['warnings']),
                'validated_at': datetime.now().isoformat()
            }
            
            return self.complete_job(job.job_key, variables)
            
        except Exception as e:
            return self.fail_job(job.job_key, job.retries - 1, f"Validation error: {str(e)}")
    
    def validate_data(self, data, schema, validation_type):
        """Валидация данных"""
        result = {
            'status': 'completed',
            'is_valid': True,
            'errors': [],
            'warnings': []
        }
        
        if not data:
            result['is_valid'] = False
            result['errors'].append('No data provided')
            return result
        
        if validation_type == 'json':
            try:
                json.loads(data)
            except json.JSONDecodeError as e:
                result['is_valid'] = False
                result['errors'].append(f'Invalid JSON: {str(e)}')
        
        elif validation_type == 'email':
            import re
            email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
            if not re.match(email_pattern, data):
                result['is_valid'] = False
                result['errors'].append('Invalid email format')
        
        elif validation_type == 'number':
            try:
                float(data)
            except ValueError:
                result['is_valid'] = False
                result['errors'].append('Not a valid number')
        
        # Если есть схема, можно добавить дополнительную валидацию
        if schema and result['is_valid']:
            result['warnings'].append('Schema validation available via expression component')
        
        return result

class FileProcessingWorker(JobWorker):
    def __init__(self, worker_name="file-worker"):
        super().__init__(worker_name, "file-processing", max_jobs=2)
    
    def process_job(self, job):
        """Обработка задания работы с файлами"""
        print(f"Обработка файла: {job.job_key}")
        
        try:
            # Извлекаем параметры
            file_path = job.variables.get('file_path', '')
            operation = job.variables.get('operation', 'read')
            
            if operation == 'read':
                result = self.read_file(file_path)
            elif operation == 'process':
                result = self.process_file(file_path)
            elif operation == 'analyze':
                result = self.analyze_file(file_path)
            else:
                raise ValueError(f"Неизвестная операция: {operation}")
            
            # Готовим результат
            variables = {
                'result': 'success',
                'operation': operation,
                'file_path': file_path,
                'processed_at': datetime.now().isoformat(),
                **result  # Добавляем результаты операции
            }
            
            return self.complete_job(job.job_key, variables)
            
        except Exception as e:
            return self.fail_job(job.job_key, job.retries - 1, str(e))
    
    def read_file(self, file_path):
        """Чтение файла"""
        # Имитация чтения файла
        time.sleep(0.5)  # Симуляция времени обработки
        
        return {
            'file_size': '1024',
            'lines_count': '50',
            'encoding': 'utf-8'
        }
    
    def process_file(self, file_path):
        """Обработка файла"""
        # Имитация обработки файла
        time.sleep(2)  # Симуляция длительной обработки
        
        return {
            'processed_lines': '45',
            'skipped_lines': '5',
            'output_file': f"{file_path}.processed"
        }
    
    def analyze_file(self, file_path):
        """Анализ файла"""
        # Имитация анализа файла
        time.sleep(1)
        
        return {
            'file_type': 'text/csv',
            'columns_count': '10',
            'rows_count': '1000',
            'has_headers': 'true'
        }

if __name__ == "__main__":
    import sys
    import threading
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python complete_job.py <job_key> [variables_json]")
        print("  python complete_job.py worker <worker_type>")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "worker":
        worker_type = sys.argv[2] if len(sys.argv) > 2 else "http"
        
        if worker_type == "http":
            worker = HTTPWorker("http-worker-python")
        elif worker_type == "validation":
            worker = ValidationWorker("validation-worker-python")
        elif worker_type == "file":
            worker = FileProcessingWorker("file-worker-python")
        else:
            print(f"Неизвестный тип воркера: {worker_type}")
            sys.exit(1)
        
        try:
            worker.start()
        except KeyboardInterrupt:
            print("\nОстановка воркера...")
            worker.stop()
    
    else:
        # Простое завершение задания
        job_key = command
        variables = {}
        
        if len(sys.argv) > 2:
            try:
                variables = json.loads(sys.argv[2])
            except json.JSONDecodeError:
                print("Неверный формат JSON для переменных")
                sys.exit(1)
        
        complete_job(job_key, variables)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const axios = require('axios');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function completeJob(jobKey, variables = {}) {
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

class JobWorker {
    constructor(workerName, jobType, maxJobs = 5) {
        this.workerName = workerName;
        this.jobType = jobType;
        this.maxJobs = maxJobs;
        this.running = false;
        
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async start() {
        console.log(`Запуск воркера ${this.workerName} для заданий типа ${this.jobType}`);
        this.running = true;
        
        while (this.running) {
            try {
                // Активируем задания
                const jobs = await this.activateJobs();
                
                if (jobs.length > 0) {
                    console.log(`Получено ${jobs.length} заданий`);
                    
                    // Обрабатываем задания параллельно
                    const promises = jobs.map(job => this.processJob(job));
                    await Promise.allSettled(promises);
                } else {
                    // Нет заданий, ждем
                    await this.sleep(1000);
                }
                
            } catch (error) {
                console.error(`Ошибка воркера: ${error.message}`);
                await this.sleep(5000);
            }
        }
        
        console.log(`Воркер ${this.workerName} остановлен`);
    }
    
    stop() {
        this.running = false;
    }
    
    async activateJobs() {
        return new Promise((resolve, reject) => {
            const request = {
                type: this.jobType,
                worker: this.workerName,
                max_jobs: this.maxJobs,
                timeout: 30000 // 30 секунд
            };
            
            this.client.activateJobs(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    resolve(response.jobs || []);
                } else {
                    console.log(`Активация не удалась: ${response.message}`);
                    resolve([]);
                }
            });
        });
    }
    
    async processJob(job) {
        console.log(`Обработка задания ${job.job_key}`);
        
        try {
            // Базовая обработка - переопределяется в подклассах
            const variables = {
                result: 'completed',
                worker: this.workerName,
                processed_at: new Date().toISOString()
            };
            
            return await this.completeJob(job.job_key, variables);
            
        } catch (error) {
            console.error(`Ошибка обработки задания ${job.job_key}: ${error.message}`);
            return await this.failJob(job.job_key, job.retries - 1, error.message);
        }
    }
    
    async completeJob(jobKey, variables) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                variables: variables
            };
            
            this.client.completeJob(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`✅ Задание ${jobKey} завершено`);
                    resolve(true);
                } else {
                    console.log(`❌ Ошибка завершения ${jobKey}: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    async failJob(jobKey, retries, errorMessage) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                retries: retries,
                error_message: errorMessage
            };
            
            this.client.failJob(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`⚠️ Задание ${jobKey} провалено: ${errorMessage}`);
                    resolve(true);
                } else {
                    console.log(`❌ Ошибка провала ${jobKey}: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

class HTTPWorker extends JobWorker {
    constructor(workerName = 'http-worker-js') {
        super(workerName, 'http-request', 3);
    }
    
    async processJob(job) {
        console.log(`HTTP запрос: ${job.job_key}`);
        
        try {
            // Извлекаем параметры
            const url = job.variables.url || '';
            const method = (job.variables.method || 'GET').toUpperCase();
            const timeout = parseInt(job.variables.timeout || '30') * 1000;
            
            let headers = {};
            try {
                headers = JSON.parse(job.variables.headers || '{}');
            } catch (e) {
                // Игнорируем ошибки парсинга заголовков
            }
            
            // Выполняем HTTP запрос
            const startTime = Date.now();
            
            const config = {
                method: method,
                url: url,
                headers: headers,
                timeout: timeout,
                validateStatus: () => true // Принимаем любой статус код
            };
            
            if (method === 'POST' || method === 'PUT') {
                config.data = job.variables.body || '';
            }
            
            const response = await axios(config);
            const executionTime = Date.now() - startTime;
            
            // Готовим результат
            const variables = {
                result: 'success',
                status_code: response.status.toString(),
                response_body: JSON.stringify(response.data).substring(0, 1000), // Ограничиваем размер
                execution_time: `${executionTime}ms`,
                response_headers: JSON.stringify(response.headers),
                completed_at: new Date().toISOString()
            };
            
            return await this.completeJob(job.job_key, variables);
            
        } catch (error) {
            let errorMessage = error.message;
            
            if (error.code === 'ECONNABORTED') {
                errorMessage = 'Request timeout';
            } else if (error.code === 'ECONNREFUSED') {
                errorMessage = 'Connection refused';
            }
            
            return await this.failJob(job.job_key, job.retries - 1, errorMessage);
        }
    }
}

class ValidationWorker extends JobWorker {
    constructor(workerName = 'validation-worker-js') {
        super(workerName, 'data-validation', 10);
    }
    
    async processJob(job) {
        console.log(`Валидация данных: ${job.job_key}`);
        
        try {
            // Извлекаем данные
            const data = job.variables.data || '';
            const schema = job.variables.schema || '';
            const validationType = job.variables.validation_type || 'json';
            
            // Выполняем валидацию
            const result = this.validateData(data, schema, validationType);
            
            // Готовим результат
            const variables = {
                validation_result: result.status,
                is_valid: result.isValid.toString(),
                errors: JSON.stringify(result.errors),
                warnings: JSON.stringify(result.warnings),
                validated_at: new Date().toISOString()
            };
            
            return await this.completeJob(job.job_key, variables);
            
        } catch (error) {
            return await this.failJob(job.job_key, job.retries - 1, `Validation error: ${error.message}`);
        }
    }
    
    validateData(data, schema, validationType) {
        const result = {
            status: 'completed',
            isValid: true,
            errors: [],
            warnings: []
        };
        
        if (!data) {
            result.isValid = false;
            result.errors.push('No data provided');
            return result;
        }
        
        switch (validationType) {
            case 'json':
                try {
                    JSON.parse(data);
                } catch (e) {
                    result.isValid = false;
                    result.errors.push(`Invalid JSON: ${e.message}`);
                }
                break;
                
            case 'email':
                const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                if (!emailRegex.test(data)) {
                    result.isValid = false;
                    result.errors.push('Invalid email format');
                }
                break;
                
            case 'number':
                if (isNaN(parseFloat(data))) {
                    result.isValid = false;
                    result.errors.push('Not a valid number');
                }
                break;
                
            case 'url':
                try {
                    new URL(data);
                } catch (e) {
                    result.isValid = false;
                    result.errors.push('Invalid URL format');
                }
                break;
                
            default:
                result.warnings.push(`Unknown validation type: ${validationType}`);
        }
        
        // Если есть схема, можно добавить дополнительную валидацию
        if (schema && result.isValid) {
            result.warnings.push('Schema validation available via expression component');
        }
        
        return result;
    }
}

class DataProcessingWorker extends JobWorker {
    constructor(workerName = 'data-worker-js') {
        super(workerName, 'data-processing', 5);
    }
    
    async processJob(job) {
        console.log(`Обработка данных: ${job.job_key}`);
        
        try {
            // Извлекаем параметры
            const data = job.variables.data || '';
            const operation = job.variables.operation || 'transform';
            
            let result;
            switch (operation) {
                case 'transform':
                    result = await this.transformData(data);
                    break;
                case 'aggregate':
                    result = await this.aggregateData(data);
                    break;
                case 'filter':
                    result = await this.filterData(data, job.variables.filter_criteria || '');
                    break;
                default:
                    throw new Error(`Unknown operation: ${operation}`);
            }
            
            // Готовим результат
            const variables = {
                result: 'success',
                operation: operation,
                processed_records: result.recordsProcessed.toString(),
                output_data: JSON.stringify(result.data).substring(0, 1000),
                processing_time: result.processingTime,
                completed_at: new Date().toISOString()
            };
            
            return await this.completeJob(job.job_key, variables);
            
        } catch (error) {
            return await this.failJob(job.job_key, job.retries - 1, error.message);
        }
    }
    
    async transformData(data) {
        const startTime = Date.now();
        
        // Имитация обработки данных
        await this.sleep(500);
        
        let parsedData;
        try {
            parsedData = JSON.parse(data);
        } catch (e) {
            throw new Error('Invalid JSON data for transformation');
        }
        
        // Простая трансформация - добавляем timestamp
        if (Array.isArray(parsedData)) {
            parsedData = parsedData.map(item => ({
                ...item,
                processed_at: new Date().toISOString()
            }));
        } else {
            parsedData.processed_at = new Date().toISOString();
        }
        
        const processingTime = Date.now() - startTime;
        
        return {
            data: parsedData,
            recordsProcessed: Array.isArray(parsedData) ? parsedData.length : 1,
            processingTime: `${processingTime}ms`
        };
    }
    
    async aggregateData(data) {
        const startTime = Date.now();
        
        // Имитация агрегации данных
        await this.sleep(300);
        
        let parsedData;
        try {
            parsedData = JSON.parse(data);
        } catch (e) {
            throw new Error('Invalid JSON data for aggregation');
        }
        
        if (!Array.isArray(parsedData)) {
            throw new Error('Data must be an array for aggregation');
        }
        
        // Простая агрегация - подсчет и группировка
        const aggregated = {
            total_records: parsedData.length,
            aggregated_at: new Date().toISOString()
        };
        
        // Если есть числовые поля, вычисляем суммы
        if (parsedData.length > 0) {
            const firstRecord = parsedData[0];
            Object.keys(firstRecord).forEach(key => {
                const values = parsedData.map(record => record[key]).filter(val => !isNaN(val));
                if (values.length > 0) {
                    aggregated[`${key}_sum`] = values.reduce((sum, val) => sum + parseFloat(val), 0);
                    aggregated[`${key}_avg`] = aggregated[`${key}_sum`] / values.length;
                    aggregated[`${key}_count`] = values.length;
                }
            });
        }
        
        const processingTime = Date.now() - startTime;
        
        return {
            data: aggregated,
            recordsProcessed: parsedData.length,
            processingTime: `${processingTime}ms`
        };
    }
    
    async filterData(data, filterCriteria) {
        const startTime = Date.now();
        
        // Имитация фильтрации данных
        await this.sleep(200);
        
        let parsedData;
        try {
            parsedData = JSON.parse(data);
        } catch (e) {
            throw new Error('Invalid JSON data for filtering');
        }
        
        if (!Array.isArray(parsedData)) {
            throw new Error('Data must be an array for filtering');
        }
        
        let filteredData = parsedData;
        
        if (filterCriteria) {
            try {
                const criteria = JSON.parse(filterCriteria);
                
                // Простая фильтрация по критериям
                filteredData = parsedData.filter(record => {
                    return Object.entries(criteria).every(([key, value]) => {
                        return record[key] === value;
                    });
                });
            } catch (e) {
                // Если критерии не JSON, пропускаем фильтрацию
                console.log('Invalid filter criteria, returning all data');
            }
        }
        
        const processingTime = Date.now() - startTime;
        
        return {
            data: filteredData,
            recordsProcessed: filteredData.length,
            processingTime: `${processingTime}ms`
        };
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node complete-job.js <job_key> [variables_json]');
        console.log('  node complete-job.js worker <worker_type>');
        console.log('');
        console.log('Типы воркеров: http, validation, data');
        process.exit(1);
    }
    
    const command = args[0];
    
    if (command === 'worker') {
        const workerType = args[1] || 'http';
        
        let worker;
        switch (workerType) {
            case 'http':
                worker = new HTTPWorker();
                break;
            case 'validation':
                worker = new ValidationWorker();
                break;
            case 'data':
                worker = new DataProcessingWorker();
                break;
            default:
                console.log(`Неизвестный тип воркера: ${workerType}`);
                process.exit(1);
        }
        
        // Обработка сигнала завершения
        process.on('SIGINT', () => {
            console.log('\nОстановка воркера...');
            worker.stop();
            process.exit(0);
        });
        
        worker.start().catch(error => {
            console.error(`Ошибка воркера: ${error.message}`);
            process.exit(1);
        });
        
    } else {
        // Простое завершение задания
        const jobKey = command;
        let variables = {};
        
        if (args.length > 1) {
            try {
                variables = JSON.parse(args[1]);
            } catch (e) {
                console.log('Неверный формат JSON для переменных');
                process.exit(1);
            }
        }
        
        completeJob(jobKey, variables).catch(error => {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        });
    }
}

module.exports = {
    completeJob,
    JobWorker,
    HTTPWorker,
    ValidationWorker,
    DataProcessingWorker
};
```

## Лучшие практики

### Обработка результатов
```go
// Структурированные результаты
variables := map[string]string{
    "result":     "success",           // Статус выполнения
    "returnCode": "0",                 // Код возврата
    "output":     jsonResult,          // Основной результат
    "metadata":   jsonMetadata,        // Метаданные выполнения
    "duration":   "1.5s",              // Время выполнения
    "worker":     "worker-instance-1", // Идентификатор воркера
}
```

### Идемпотентность
```python
def complete_job_idempotent(job_key, variables):
    """Идемпотентное завершение задания"""
    # Проверяем, не завершено ли уже задание
    job_status = get_job_status(job_key)
    
    if job_status and job_status['state'] == 'COMPLETED':
        print(f"Задание {job_key} уже завершено")
        return True
    
    return complete_job(job_key, variables)
```

### Обработка ошибок
```javascript
async function safeCompleteJob(jobKey, variables) {
    const maxRetries = 3;
    let attempt = 0;
    
    while (attempt < maxRetries) {
        try {
            return await completeJob(jobKey, variables);
        } catch (error) {
            attempt++;
            console.log(`Попытка ${attempt}/${maxRetries} завершения ${jobKey} не удалась: ${error.message}`);
            
            if (attempt >= maxRetries) {
                throw error;
            }
            
            // Экспоненциальная задержка
            await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));
        }
    }
}
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный job_key или переменные
- `NOT_FOUND` (5): Задание не найдено или уже завершено
- `DEADLINE_EXCEEDED` (4): Задание устарело (истек timeout)
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
- [FailJob](fail-job.md) - Сигнализация о неудачном выполнении
- [GetJob](get-job.md) - Получение деталей задания
- [ListJobs](list-jobs.md) - Список всех заданий
