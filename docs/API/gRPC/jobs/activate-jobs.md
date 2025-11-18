# ActivateJobs

## Описание
Активирует задания для воркера (polling). Воркеры используют этот метод для получения новых заданий для выполнения. Поддерживает потоковую передачу для длительного polling.

## Синтаксис
```protobuf
rpc ActivateJobs(ActivateJobsRequest) returns (stream ActivateJobsResponse);
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

### ActivateJobsRequest
```protobuf
message ActivateJobsRequest {
  string type = 1;                      // Тип заданий для активации
  string worker = 2;                    // Идентификатор воркера
  int32 timeout = 3;                    // Timeout в миллисекундах
  int32 max_jobs_to_activate = 4;       // Максимальное количество заданий
  repeated string fetch_variable = 5;   // Переменные для загрузки
  string tenant_ids = 6;                // ID тенантов (разделенные запятыми)
}
```

#### Поля:
- **type** (string, required): Тип заданий для активации (например, `"http-request"`, `"email-send"`)
- **worker** (string, required): Уникальный идентификатор воркера
- **timeout** (int32, optional): Timeout активации в миллисекундах (по умолчанию: 30000)
- **max_jobs_to_activate** (int32, optional): Максимальное количество заданий для активации (по умолчанию: 10, максимум: 100)
- **fetch_variable** (repeated string, optional): Список переменных для загрузки с заданием
- **tenant_ids** (string, optional): ID тенантов, разделенные запятыми (для мультитенантности)

## Параметры ответа

### ActivateJobsResponse (stream)
```protobuf
message ActivateJobsResponse {
  repeated ActivatedJob jobs = 1;       // Список активированных заданий
  bool success = 2;                     // Статус успешности
  string error_message = 3;             // Сообщение об ошибке
}

message ActivatedJob {
  string job_key = 1;                   // Уникальный ключ задания
  string type = 2;                      // Тип задания
  string process_instance_key = 3;      // Ключ экземпляра процесса
  string bpmn_process_id = 4;           // ID BPMN процесса
  string process_definition_version = 5; // Версия определения процесса
  string process_definition_key = 6;    // Ключ определения процесса
  string element_id = 7;                // ID элемента BPMN
  string element_instance_key = 8;      // Ключ экземпляра элемента
  map<string, string> custom_headers = 9; // Пользовательские заголовки
  string worker = 10;                   // Идентификатор воркера
  int32 retries = 11;                   // Количество попыток
  int64 deadline = 12;                  // Deadline задания (Unix timestamp)
  string variables = 13;                // Переменные в формате JSON
  string tenant_id = 14;                // ID тенанта
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
    "io"
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
    
    // Создаем воркер для HTTP заданий
    worker := NewJobWorker(client, "http-worker-1", "http-request")
    worker.Start(ctx)
}

type JobWorker struct {
    client     pb.JobsServiceClient
    workerName string
    jobType    string
    maxJobs    int32
    timeout    int32
    running    bool
}

func NewJobWorker(client pb.JobsServiceClient, workerName, jobType string) *JobWorker {
    return &JobWorker{
        client:     client,
        workerName: workerName,
        jobType:    jobType,
        maxJobs:    10,
        timeout:    30000, // 30 секунд
        running:    false,
    }
}

func (w *JobWorker) Start(ctx context.Context) {
    w.running = true
    fmt.Printf("🚀 Запуск воркера %s для заданий типа %s\n", w.workerName, w.jobType)
    
    for w.running {
        err := w.activateAndProcessJobs(ctx)
        if err != nil {
            log.Printf("Ошибка активации заданий: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        time.Sleep(1 * time.Second)
    }
    
    fmt.Printf("⏹️ Воркер %s остановлен\n", w.workerName)
}

func (w *JobWorker) activateAndProcessJobs(ctx context.Context) error {
    request := &pb.ActivateJobsRequest{
        Type:               w.jobType,
        Worker:             w.workerName,
        Timeout:            w.timeout,
        MaxJobsToActivate:  w.maxJobs,
        FetchVariable:      []string{"url", "method", "headers", "body"},
    }
    
    stream, err := w.client.ActivateJobs(ctx, request)
    if err != nil {
        return fmt.Errorf("ошибка создания stream: %v", err)
    }
    defer stream.CloseSend()
    
    for {
        response, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("ошибка получения ответа: %v", err)
        }
        
        if !response.Success {
            log.Printf("⚠️ Активация не удалась: %s", response.ErrorMessage)
            continue
        }
        
        if len(response.Jobs) > 0 {
            fmt.Printf("📥 Получено %d заданий\n", len(response.Jobs))
            
            for _, job := range response.Jobs {
                go w.processJob(ctx, job)
            }
        }
    }
    
    return nil
}

func (w *JobWorker) processJob(ctx context.Context, job *pb.ActivatedJob) {
    fmt.Printf("⚙️ Обработка задания %s (тип: %s)\n", job.JobKey, job.Type)
    
    // Парсим переменные
    var variables map[string]interface{}
    if err := json.Unmarshal([]byte(job.Variables), &variables); err != nil {
        log.Printf("❌ Ошибка парсинга переменных для %s: %v", job.JobKey, err)
        w.failJob(ctx, job.JobKey, job.Retries-1, "Invalid variables JSON")
        return
    }
    
    // Выполняем работу
    switch job.Type {
    case "http-request":
        w.processHTTPJob(ctx, job, variables)
    default:
        log.Printf("⚠️ Неизвестный тип задания: %s", job.Type)
        w.failJob(ctx, job.JobKey, job.Retries-1, "Unknown job type")
    }
}

func (w *JobWorker) processHTTPJob(ctx context.Context, job *pb.ActivatedJob, variables map[string]interface{}) {
    url, _ := variables["url"].(string)
    method, _ := variables["method"].(string)
    if method == "" {
        method = "GET"
    }
    
    fmt.Printf("🌐 HTTP %s запрос к %s\n", method, url)
    
    // Имитация HTTP запроса
    time.Sleep(100 * time.Millisecond)
    
    if url == "" {
        w.failJob(ctx, job.JobKey, job.Retries-1, "URL is required")
        return
    }
    
    // Успешное выполнение
    resultVariables := map[string]string{
        "httpStatus":     "200",
        "responseBody":   `{"result": "success"}`,
        "executionTime":  "150ms",
        "completedAt":    time.Now().Format(time.RFC3339),
    }
    
    w.completeJob(ctx, job.JobKey, resultVariables)
}

func (w *JobWorker) completeJob(ctx context.Context, jobKey string, variables map[string]string) {
    response, err := w.client.CompleteJob(ctx, &pb.CompleteJobRequest{
        JobKey:    jobKey,
        Variables: variables,
    })
    
    if err != nil {
        log.Printf("❌ Ошибка завершения задания %s: %v", jobKey, err)
        return
    }
    
    if response.Success {
        fmt.Printf("✅ Задание %s успешно завершено\n", jobKey)
    }
}

func (w *JobWorker) failJob(ctx context.Context, jobKey string, retries int32, errorMessage string) {
    response, err := w.client.FailJob(ctx, &pb.FailJobRequest{
        JobKey:       jobKey,
        Retries:      retries,
        ErrorMessage: errorMessage,
    })
    
    if err != nil {
        log.Printf("❌ Ошибка провала задания %s: %v", jobKey, err)
        return
    }
    
    if response.Success {
        fmt.Printf("⚠️ Задание %s провалено: %s\n", jobKey, errorMessage)
    }
}
```

### Python
```python
import grpc
import json
import time
import threading
from concurrent.futures import ThreadPoolExecutor

import jobs_pb2
import jobs_pb2_grpc

class JobWorker:
    def __init__(self, worker_name, job_type, max_jobs=10, timeout=30000):
        self.worker_name = worker_name
        self.job_type = job_type
        self.max_jobs = max_jobs
        self.timeout = timeout
        self.running = False
        
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def start(self):
        print(f"🚀 Запуск воркера {self.worker_name} для заданий типа {self.job_type}")
        self.running = True
        
        while self.running:
            try:
                self.activate_and_process_jobs()
            except KeyboardInterrupt:
                break
            except Exception as e:
                print(f"Ошибка воркера: {e}")
                time.sleep(5)
        
        self.running = False
        print(f"⏹️ Воркер {self.worker_name} остановлен")
    
    def activate_and_process_jobs(self):
        request = jobs_pb2.ActivateJobsRequest(
            type=self.job_type,
            worker=self.worker_name,
            timeout=self.timeout,
            max_jobs_to_activate=self.max_jobs,
            fetch_variable=['url', 'method', 'headers', 'body']
        )
        
        try:
            stream = self.stub.ActivateJobs(request, metadata=self.metadata)
            
            for response in stream:
                if not response.success:
                    print(f"⚠️ Активация не удалась: {response.error_message}")
                    continue
                
                if response.jobs:
                    print(f"📥 Получено {len(response.jobs)} заданий")
                    
                    with ThreadPoolExecutor(max_workers=self.max_jobs) as executor:
                        futures = [executor.submit(self.process_job, job) for job in response.jobs]
                        
                        for future in futures:
                            try:
                                future.result()
                            except Exception as e:
                                print(f"Ошибка обработки задания: {e}")
                
                if not self.running:
                    break
                    
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.code()} - {e.details()}")
            time.sleep(5)
    
    def process_job(self, job):
        print(f"⚙️ Обработка задания {job.job_key} (тип: {job.type})")
        
        try:
            variables = json.loads(job.variables) if job.variables else {}
            
            if job.type == "http-request":
                self.process_http_job(job, variables)
            else:
                print(f"⚠️ Неизвестный тип задания: {job.type}")
                self.fail_job(job.job_key, job.retries - 1, "Unknown job type")
                
        except Exception as e:
            print(f"❌ Ошибка обработки задания {job.job_key}: {e}")
            self.fail_job(job.job_key, job.retries - 1, str(e))
    
    def process_http_job(self, job, variables):
        url = variables.get('url', '')
        method = variables.get('method', 'GET')
        
        print(f"🌐 HTTP {method} запрос к {url}")
        time.sleep(0.1)  # Имитация HTTP запроса
        
        if not url:
            self.fail_job(job.job_key, job.retries - 1, "URL is required")
            return
        
        result_variables = {
            'httpStatus': '200',
            'responseBody': '{"result": "success"}',
            'executionTime': '150ms',
            'completedAt': time.strftime('%Y-%m-%dT%H:%M:%SZ')
        }
        
        self.complete_job(job.job_key, result_variables)
    
    def complete_job(self, job_key, variables):
        request = jobs_pb2.CompleteJobRequest(
            job_key=job_key,
            variables=variables
        )
        
        try:
            response = self.stub.CompleteJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"✅ Задание {job_key} успешно завершено")
            else:
                print(f"❌ Завершение задания {job_key} не удалось: {response.message}")
                
        except grpc.RpcError as e:
            print(f"❌ Ошибка завершения задания {job_key}: {e.details()}")
    
    def fail_job(self, job_key, retries, error_message):
        request = jobs_pb2.FailJobRequest(
            job_key=job_key,
            retries=retries,
            error_message=error_message
        )
        
        try:
            response = self.stub.FailJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"⚠️ Задание {job_key} провалено: {error_message}")
                
        except grpc.RpcError as e:
            print(f"❌ Ошибка провала задания {job_key}: {e.details()}")

if __name__ == "__main__":
    worker = JobWorker("python-worker", "http-request")
    try:
        worker.start()
    except KeyboardInterrupt:
        worker.stop()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

class JobWorker {
    constructor(workerName, jobType, maxJobs = 10, timeout = 30000) {
        this.workerName = workerName;
        this.jobType = jobType;
        this.maxJobs = maxJobs;
        this.timeout = timeout;
        this.running = false;
        
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async start() {
        console.log(`🚀 Запуск воркера ${this.workerName} для заданий типа ${this.jobType}`);
        this.running = true;
        
        while (this.running) {
            try {
                await this.activateAndProcessJobs();
            } catch (error) {
                console.error(`Ошибка воркера: ${error.message}`);
                await this.sleep(5000);
            }
        }
        
        console.log(`⏹️ Воркер ${this.workerName} остановлен`);
    }
    
    async activateAndProcessJobs() {
        const request = {
            type: this.jobType,
            worker: this.workerName,
            timeout: this.timeout,
            max_jobs_to_activate: this.maxJobs,
            fetch_variable: ['url', 'method', 'headers', 'body']
        };
        
        return new Promise((resolve, reject) => {
            const stream = this.client.activateJobs(request, this.metadata);
            
            stream.on('data', (response) => {
                if (!response.success) {
                    console.log(`⚠️ Активация не удалась: ${response.error_message}`);
                    return;
                }
                
                if (response.jobs && response.jobs.length > 0) {
                    console.log(`📥 Получено ${response.jobs.length} заданий`);
                    
                    const promises = response.jobs.map(job => this.processJob(job));
                    Promise.allSettled(promises);
                }
            });
            
            stream.on('end', resolve);
            stream.on('error', reject);
        });
    }
    
    async processJob(job) {
        console.log(`⚙️ Обработка задания ${job.job_key} (тип: ${job.type})`);
        
        try {
            const variables = job.variables ? JSON.parse(job.variables) : {};
            
            if (job.type === 'http-request') {
                await this.processHttpJob(job, variables);
            } else {
                console.log(`⚠️ Неизвестный тип задания: ${job.type}`);
                await this.failJob(job.job_key, job.retries - 1, 'Unknown job type');
            }
            
        } catch (error) {
            console.error(`❌ Ошибка обработки задания ${job.job_key}: ${error.message}`);
            await this.failJob(job.job_key, job.retries - 1, error.message);
        }
    }
    
    async processHttpJob(job, variables) {
        const url = variables.url || '';
        const method = variables.method || 'GET';
        
        console.log(`🌐 HTTP ${method} запрос к ${url}`);
        await this.sleep(100); // Имитация HTTP запроса
        
        if (!url) {
            await this.failJob(job.job_key, job.retries - 1, 'URL is required');
            return;
        }
        
        const resultVariables = {
            httpStatus: '200',
            responseBody: '{"result": "success"}',
            executionTime: '150ms',
            completedAt: new Date().toISOString()
        };
        
        await this.completeJob(job.job_key, resultVariables);
    }
    
    async completeJob(jobKey, variables) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                variables: variables
            };
            
            this.client.completeJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`❌ Ошибка завершения задания ${jobKey}: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`✅ Задание ${jobKey} успешно завершено`);
                    resolve(true);
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
                    console.error(`❌ Ошибка провала задания ${jobKey}: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    console.log(`⚠️ Задание ${jobKey} провалено: ${errorMessage}`);
                    resolve(true);
                }
            });
        });
    }
    
    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
    
    stop() {
        this.running = false;
    }
}

// Пример использования
if (require.main === module) {
    const worker = new JobWorker('js-worker', 'http-request');
    
    process.on('SIGINT', () => {
        console.log('\nОстановка воркера...');
        worker.stop();
        process.exit(0);
    });
    
    worker.start().catch(error => {
        console.error(`Ошибка воркера: ${error.message}`);
        process.exit(1);
    });
}

module.exports = { JobWorker };
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры активации
- `DEADLINE_EXCEEDED` (4): Timeout активации
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ
- `RESOURCE_EXHAUSTED` (8): Превышено количество воркеров

## Связанные методы
- [CompleteJob](complete-job.md) - Завершение полученного задания
- [FailJob](fail-job.md) - Сигнализация о неудачном выполнении
- [ListJobs](list-jobs.md) - Просмотр доступных заданий
- [GetJob](get-job.md) - Получение деталей конкретного задания