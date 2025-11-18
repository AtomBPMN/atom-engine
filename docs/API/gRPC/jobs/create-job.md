# CreateJob

## Описание
Создает новое задание вручную для выполнения воркерами. Используется для программного создания заданий вне контекста BPMN процессов.

## Синтаксис
```protobuf
rpc CreateJob(CreateJobRequest) returns (CreateJobResponse);
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

### CreateJobRequest
```protobuf
message CreateJobRequest {
  string job_type = 1;                     // Тип задания
  int32 retries = 2;                       // Количество попыток
  map<string, string> variables = 3;       // Переменные задания
  map<string, string> custom_headers = 4;  // Пользовательские заголовки
  int64 timeout = 5;                       // Таймаут в миллисекундах
  string process_instance_key = 6;         // Ключ экземпляра процесса (опционально)
  string element_id = 7;                   // ID элемента BPMN (опционально)
}
```

#### Поля:
- **job_type** (string, required): Тип задания для сопоставления с воркерами
- **retries** (int32, optional): Количество попыток (по умолчанию 3)
- **variables** (map, optional): Переменные, доступные воркеру
- **custom_headers** (map, optional): Пользовательские заголовки для задания
- **timeout** (int64, optional): Таймаут выполнения в миллисекундах
- **process_instance_key** (string, optional): Связь с экземпляром процесса
- **element_id** (string, optional): ID элемента BPMN для контекста

## Параметры ответа

### CreateJobResponse
```protobuf
message CreateJobResponse {
  bool success = 1;         // Статус успешности операции
  string message = 2;       // Сообщение о результате
  string job_key = 3;       // Ключ созданного задания
  string created_at = 4;    // Время создания (RFC3339)
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
    
    // Простое создание задания
    response, err := client.CreateJob(ctx, &pb.CreateJobRequest{
        JobType: "data-processing",
        Retries: 3,
        Variables: map[string]string{
            "file_path": "/data/input.csv",
            "format":    "csv",
        },
        CustomHeaders: map[string]string{
            "priority": "high",
            "region":   "us-east-1",
        },
        Timeout: 300000, // 5 минут
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Задание создано: %s\n", response.JobKey)
        fmt.Printf("   Время создания: %s\n", response.CreatedAt)
    } else {
        fmt.Printf("❌ Ошибка создания: %s\n", response.Message)
    }
}

// Фабрика заданий для различных типов операций
type JobFactory struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewJobFactory(client pb.JobsServiceClient, ctx context.Context) *JobFactory {
    return &JobFactory{
        client: client,
        ctx:    ctx,
    }
}

func (jf *JobFactory) CreateDataProcessingJob(filePath, format string, priority string) (string, error) {
    variables := map[string]string{
        "file_path": filePath,
        "format":    format,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    headers := map[string]string{
        "priority":    priority,
        "job_category": "data-processing",
    }
    
    // Устанавливаем таймаут в зависимости от приоритета
    timeout := int64(300000) // 5 минут по умолчанию
    if priority == "high" {
        timeout = 180000 // 3 минуты для высокого приоритета
    } else if priority == "low" {
        timeout = 600000 // 10 минут для низкого приоритета
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "data-processing",
        Retries:       5, // Больше попыток для важных данных
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       timeout,
    })
    
    if err != nil {
        return "", fmt.Errorf("ошибка создания задания обработки данных: %v", err)
    }
    
    if !response.Success {
        return "", fmt.Errorf("не удалось создать задание: %s", response.Message)
    }
    
    fmt.Printf("📊 Создано задание обработки данных: %s\n", response.JobKey)
    fmt.Printf("   Файл: %s\n", filePath)
    fmt.Printf("   Приоритет: %s\n", priority)
    
    return response.JobKey, nil
}

func (jf *JobFactory) CreateEmailJob(recipient, subject, body string) (string, error) {
    variables := map[string]string{
        "recipient": recipient,
        "subject":   subject,
        "body":      body,
        "sender":    "system@company.com",
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    headers := map[string]string{
        "email_type": "notification",
        "priority":   "normal",
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "send-email",
        Retries:       3,
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       60000, // 1 минута
    })
    
    if err != nil {
        return "", fmt.Errorf("ошибка создания email задания: %v", err)
    }
    
    if !response.Success {
        return "", fmt.Errorf("не удалось создать email задание: %s", response.Message)
    }
    
    fmt.Printf("📧 Создано email задание: %s\n", response.JobKey)
    fmt.Printf("   Получатель: %s\n", recipient)
    fmt.Printf("   Тема: %s\n", subject)
    
    return response.JobKey, nil
}

func (jf *JobFactory) CreateAPICallJob(url, method string, payload map[string]string) (string, error) {
    variables := map[string]string{
        "url":       url,
        "method":    method,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    // Добавляем payload как переменные
    for key, value := range payload {
        variables["payload_"+key] = value
    }
    
    headers := map[string]string{
        "api_type":    "external",
        "retry_policy": "exponential_backoff",
    }
    
    // Больше попыток для внешних API
    retries := int32(5)
    if method == "GET" {
        retries = 3 // Меньше попыток для безопасных операций
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "api-call",
        Retries:       retries,
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       120000, // 2 минуты
    })
    
    if err != nil {
        return "", fmt.Errorf("ошибка создания API задания: %v", err)
    }
    
    if !response.Success {
        return "", fmt.Errorf("не удалось создать API задание: %s", response.Message)
    }
    
    fmt.Printf("🌐 Создано API задание: %s\n", response.JobKey)
    fmt.Printf("   URL: %s\n", url)
    fmt.Printf("   Метод: %s\n", method)
    
    return response.JobKey, nil
}

func (jf *JobFactory) CreateReportJob(reportType, format string, parameters map[string]string) (string, error) {
    variables := map[string]string{
        "report_type": reportType,
        "format":      format,
        "generated_at": time.Now().Format(time.RFC3339),
    }
    
    // Добавляем параметры отчета
    for key, value := range parameters {
        variables["param_"+key] = value
    }
    
    headers := map[string]string{
        "report_category": "analytics",
        "priority":        "normal",
    }
    
    // Больше времени для генерации отчетов
    timeout := int64(900000) // 15 минут
    if reportType == "complex_analytics" {
        timeout = 1800000 // 30 минут для сложных отчетов
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "generate-report",
        Retries:       2, // Меньше попыток для ресурсоемких задач
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       timeout,
    })
    
    if err != nil {
        return "", fmt.Errorf("ошибка создания задания отчета: %v", err)
    }
    
    if !response.Success {
        return "", fmt.Errorf("не удалось создать задание отчета: %s", response.Message)
    }
    
    fmt.Printf("📊 Создано задание отчета: %s\n", response.JobKey)
    fmt.Printf("   Тип: %s\n", reportType)
    fmt.Printf("   Формат: %s\n", format)
    
    return response.JobKey, nil
}

// Массовое создание заданий
func (jf *JobFactory) CreateBatchJobs(jobRequests []JobRequest) ([]string, error) {
    var jobKeys []string
    var errors []error
    
    for i, request := range jobRequests {
        jobKey, err := jf.createSingleJob(request)
        if err != nil {
            errors = append(errors, fmt.Errorf("задание %d: %v", i+1, err))
            continue
        }
        
        jobKeys = append(jobKeys, jobKey)
        
        // Небольшая задержка между созданием заданий
        time.Sleep(10 * time.Millisecond)
    }
    
    if len(errors) > 0 {
        fmt.Printf("⚠️ Создано %d из %d заданий\n", len(jobKeys), len(jobRequests))
        for _, err := range errors {
            fmt.Printf("   Ошибка: %v\n", err)
        }
    } else {
        fmt.Printf("✅ Успешно создано %d заданий\n", len(jobKeys))
    }
    
    return jobKeys, nil
}

type JobRequest struct {
    JobType       string
    Retries       int32
    Variables     map[string]string
    CustomHeaders map[string]string
    Timeout       int64
}

func (jf *JobFactory) createSingleJob(request JobRequest) (string, error) {
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       request.JobType,
        Retries:       request.Retries,
        Variables:     request.Variables,
        CustomHeaders: request.CustomHeaders,
        Timeout:       request.Timeout,
    })
    
    if err != nil {
        return "", err
    }
    
    if !response.Success {
        return "", fmt.Errorf(response.Message)
    }
    
    return response.JobKey, nil
}

// Планировщик заданий
type JobScheduler struct {
    factory *JobFactory
    running bool
    stopCh  chan struct{}
}

func NewJobScheduler(factory *JobFactory) *JobScheduler {
    return &JobScheduler{
        factory: factory,
        stopCh:  make(chan struct{}),
    }
}

func (js *JobScheduler) Start() {
    if js.running {
        return
    }
    
    js.running = true
    go js.run()
    fmt.Printf("⏰ Планировщик заданий запущен\n")
}

func (js *JobScheduler) Stop() {
    if !js.running {
        return
    }
    
    close(js.stopCh)
    js.running = false
    fmt.Printf("⏰ Планировщик заданий остановлен\n")
}

func (js *JobScheduler) run() {
    ticker := time.NewTicker(5 * time.Minute) // Каждые 5 минут
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            js.createScheduledJobs()
        case <-js.stopCh:
            return
        }
    }
}

func (js *JobScheduler) createScheduledJobs() {
    now := time.Now()
    
    // Создаем регулярные задания
    
    // Ежечасные задания
    if now.Minute() == 0 {
        js.createHourlyJobs()
    }
    
    // Ежедневные задания в 2:00
    if now.Hour() == 2 && now.Minute() == 0 {
        js.createDailyJobs()
    }
    
    // Еженедельные задания в воскресенье в 3:00
    if now.Weekday() == time.Sunday && now.Hour() == 3 && now.Minute() == 0 {
        js.createWeeklyJobs()
    }
}

func (js *JobScheduler) createHourlyJobs() {
    fmt.Printf("🕐 Создание ежечасных заданий...\n")
    
    // Мониторинг системы
    _, err := js.factory.CreateAPICallJob("http://monitoring/api/health", "GET", nil)
    if err != nil {
        fmt.Printf("⚠️ Ошибка создания задания мониторинга: %v\n", err)
    }
}

func (js *JobScheduler) createDailyJobs() {
    fmt.Printf("📅 Создание ежедневных заданий...\n")
    
    // Генерация дневного отчета
    params := map[string]string{
        "date":   time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
        "region": "all",
    }
    
    _, err := js.factory.CreateReportJob("daily_summary", "pdf", params)
    if err != nil {
        fmt.Printf("⚠️ Ошибка создания задания дневного отчета: %v\n", err)
    }
    
    // Очистка временных файлов
    _, err = js.factory.CreateDataProcessingJob("/tmp/cleanup", "directory", "low")
    if err != nil {
        fmt.Printf("⚠️ Ошибка создания задания очистки: %v\n", err)
    }
}

func (js *JobScheduler) createWeeklyJobs() {
    fmt.Printf("📊 Создание еженедельных заданий...\n")
    
    // Генерация недельного отчета
    params := map[string]string{
        "start_date": time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
        "end_date":   time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
        "type":       "comprehensive",
    }
    
    _, err := js.factory.CreateReportJob("weekly_analytics", "excel", params)
    if err != nil {
        fmt.Printf("⚠️ Ошибка создания задания недельного отчета: %v\n", err)
    }
}

// Создание заданий на основе событий
func (jf *JobFactory) CreateJobFromEvent(eventType string, eventData map[string]interface{}) (string, error) {
    switch eventType {
    case "user_registered":
        return jf.handleUserRegistered(eventData)
    case "order_placed":
        return jf.handleOrderPlaced(eventData)
    case "payment_failed":
        return jf.handlePaymentFailed(eventData)
    case "file_uploaded":
        return jf.handleFileUploaded(eventData)
    default:
        return "", fmt.Errorf("неизвестный тип события: %s", eventType)
    }
}

func (jf *JobFactory) handleUserRegistered(data map[string]interface{}) (string, error) {
    email, ok := data["email"].(string)
    if !ok {
        return "", fmt.Errorf("отсутствует email в данных события")
    }
    
    name, _ := data["name"].(string)
    if name == "" {
        name = "Пользователь"
    }
    
    return jf.CreateEmailJob(
        email,
        "Добро пожаловать!",
        fmt.Sprintf("Здравствуйте, %s! Спасибо за регистрацию.", name),
    )
}

func (jf *JobFactory) handleOrderPlaced(data map[string]interface{}) (string, error) {
    orderID, ok := data["order_id"].(string)
    if !ok {
        return "", fmt.Errorf("отсутствует order_id в данных события")
    }
    
    variables := map[string]string{
        "order_id": orderID,
        "action":   "process_payment",
    }
    
    if customerID, ok := data["customer_id"].(string); ok {
        variables["customer_id"] = customerID
    }
    
    headers := map[string]string{
        "priority":      "high",
        "order_context": "true",
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "process-payment",
        Retries:       3,
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       180000, // 3 минуты
    })
    
    if err != nil {
        return "", err
    }
    
    if !response.Success {
        return "", fmt.Errorf(response.Message)
    }
    
    fmt.Printf("💳 Создано задание обработки платежа для заказа %s: %s\n", orderID, response.JobKey)
    return response.JobKey, nil
}

func (jf *JobFactory) handlePaymentFailed(data map[string]interface{}) (string, error) {
    orderID, ok := data["order_id"].(string)
    if !ok {
        return "", fmt.Errorf("отсутствует order_id в данных события")
    }
    
    variables := map[string]string{
        "order_id": orderID,
        "action":   "notify_failure",
    }
    
    if reason, ok := data["failure_reason"].(string); ok {
        variables["failure_reason"] = reason
    }
    
    headers := map[string]string{
        "priority":        "high",
        "notification_type": "payment_failure",
    }
    
    response, err := jf.client.CreateJob(jf.ctx, &pb.CreateJobRequest{
        JobType:       "send-notification",
        Retries:       5, // Важно уведомить о проблеме
        Variables:     variables,
        CustomHeaders: headers,
        Timeout:       60000, // 1 минута
    })
    
    if err != nil {
        return "", err
    }
    
    if !response.Success {
        return "", fmt.Errorf(response.Message)
    }
    
    fmt.Printf("❌ Создано задание уведомления о неудачном платеже для заказа %s: %s\n", 
               orderID, response.JobKey)
    return response.JobKey, nil
}

func (jf *JobFactory) handleFileUploaded(data map[string]interface{}) (string, error) {
    filePath, ok := data["file_path"].(string)
    if !ok {
        return "", fmt.Errorf("отсутствует file_path в данных события")
    }
    
    fileType := "unknown"
    if ft, ok := data["file_type"].(string); ok {
        fileType = ft
    }
    
    // Выбираем приоритет на основе типа файла
    priority := "normal"
    if fileType == "image" || fileType == "video" {
        priority = "low" // Медиа файлы можно обрабатывать позже
    } else if fileType == "document" {
        priority = "high" // Документы важны
    }
    
    return jf.CreateDataProcessingJob(filePath, fileType, priority)
}
```

### Python
```python
import grpc
import time
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional

import jobs_pb2
import jobs_pb2_grpc

def create_job(job_type, retries=3, variables=None, custom_headers=None, timeout=None, 
               process_instance_key="", element_id=""):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.CreateJobRequest(
        job_type=job_type,
        retries=retries,
        variables=variables or {},
        custom_headers=custom_headers or {},
        timeout=timeout or 300000,  # 5 минут по умолчанию
        process_instance_key=process_instance_key,
        element_id=element_id
    )
    
    try:
        response = stub.CreateJob(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Задание создано: {response.job_key}")
            print(f"   Время создания: {response.created_at}")
            return response.job_key
        else:
            print(f"❌ Ошибка создания: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

class JobFactory:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def create_data_processing_job(self, file_path, format_type, priority="normal"):
        """Создает задание обработки данных"""
        variables = {
            'file_path': file_path,
            'format': format_type,
            'timestamp': datetime.now().isoformat(),
        }
        
        headers = {
            'priority': priority,
            'job_category': 'data-processing',
        }
        
        # Устанавливаем таймаут в зависимости от приоритета
        timeout_map = {
            'high': 180000,   # 3 минуты
            'normal': 300000, # 5 минут
            'low': 600000,    # 10 минут
        }
        timeout = timeout_map.get(priority, 300000)
        
        job_key = self._create_job(
            job_type="data-processing",
            retries=5,
            variables=variables,
            custom_headers=headers,
            timeout=timeout
        )
        
        if job_key:
            print(f"📊 Создано задание обработки данных: {job_key}")
            print(f"   Файл: {file_path}")
            print(f"   Приоритет: {priority}")
        
        return job_key
    
    def create_email_job(self, recipient, subject, body):
        """Создает задание отправки email"""
        variables = {
            'recipient': recipient,
            'subject': subject,
            'body': body,
            'sender': 'system@company.com',
            'timestamp': datetime.now().isoformat(),
        }
        
        headers = {
            'email_type': 'notification',
            'priority': 'normal',
        }
        
        job_key = self._create_job(
            job_type="send-email",
            retries=3,
            variables=variables,
            custom_headers=headers,
            timeout=60000  # 1 минута
        )
        
        if job_key:
            print(f"📧 Создано email задание: {job_key}")
            print(f"   Получатель: {recipient}")
            print(f"   Тема: {subject}")
        
        return job_key
    
    def create_api_call_job(self, url, method, payload=None):
        """Создает задание вызова API"""
        variables = {
            'url': url,
            'method': method,
            'timestamp': datetime.now().isoformat(),
        }
        
        # Добавляем payload как переменные
        if payload:
            for key, value in payload.items():
                variables[f'payload_{key}'] = str(value)
        
        headers = {
            'api_type': 'external',
            'retry_policy': 'exponential_backoff',
        }
        
        # Больше попыток для внешних API
        retries = 5 if method.upper() != 'GET' else 3
        
        job_key = self._create_job(
            job_type="api-call",
            retries=retries,
            variables=variables,
            custom_headers=headers,
            timeout=120000  # 2 минуты
        )
        
        if job_key:
            print(f"🌐 Создано API задание: {job_key}")
            print(f"   URL: {url}")
            print(f"   Метод: {method}")
        
        return job_key
    
    def create_report_job(self, report_type, format_type, parameters=None):
        """Создает задание генерации отчета"""
        variables = {
            'report_type': report_type,
            'format': format_type,
            'generated_at': datetime.now().isoformat(),
        }
        
        # Добавляем параметры отчета
        if parameters:
            for key, value in parameters.items():
                variables[f'param_{key}'] = str(value)
        
        headers = {
            'report_category': 'analytics',
            'priority': 'normal',
        }
        
        # Больше времени для генерации отчетов
        timeout = 900000  # 15 минут
        if report_type == 'complex_analytics':
            timeout = 1800000  # 30 минут
        
        job_key = self._create_job(
            job_type="generate-report",
            retries=2,
            variables=variables,
            custom_headers=headers,
            timeout=timeout
        )
        
        if job_key:
            print(f"📊 Создано задание отчета: {job_key}")
            print(f"   Тип: {report_type}")
            print(f"   Формат: {format_type}")
        
        return job_key
    
    def create_batch_jobs(self, job_requests):
        """Массовое создание заданий"""
        job_keys = []
        errors = []
        
        for i, request in enumerate(job_requests):
            try:
                job_key = self._create_job(**request)
                if job_key:
                    job_keys.append(job_key)
                else:
                    errors.append(f"Задание {i+1}: Не удалось создать")
                
                # Небольшая задержка между созданием заданий
                time.sleep(0.01)
                
            except Exception as e:
                errors.append(f"Задание {i+1}: {e}")
        
        if errors:
            print(f"⚠️ Создано {len(job_keys)} из {len(job_requests)} заданий")
            for error in errors:
                print(f"   Ошибка: {error}")
        else:
            print(f"✅ Успешно создано {len(job_keys)} заданий")
        
        return job_keys
    
    def _create_job(self, job_type, retries=3, variables=None, custom_headers=None, 
                   timeout=300000, process_instance_key="", element_id=""):
        """Внутренний метод создания задания"""
        try:
            request = jobs_pb2.CreateJobRequest(
                job_type=job_type,
                retries=retries,
                variables=variables or {},
                custom_headers=custom_headers or {},
                timeout=timeout,
                process_instance_key=process_instance_key,
                element_id=element_id
            )
            
            response = self.stub.CreateJob(request, metadata=self.metadata)
            
            if response.success:
                return response.job_key
            else:
                print(f"❌ Ошибка создания задания: {response.message}")
                return None
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return None
    
    def create_job_from_event(self, event_type, event_data):
        """Создание заданий на основе событий"""
        handlers = {
            'user_registered': self._handle_user_registered,
            'order_placed': self._handle_order_placed,
            'payment_failed': self._handle_payment_failed,
            'file_uploaded': self._handle_file_uploaded,
        }
        
        handler = handlers.get(event_type)
        if not handler:
            raise ValueError(f"Неизвестный тип события: {event_type}")
        
        return handler(event_data)
    
    def _handle_user_registered(self, data):
        email = data.get('email')
        if not email:
            raise ValueError("Отсутствует email в данных события")
        
        name = data.get('name', 'Пользователь')
        
        return self.create_email_job(
            recipient=email,
            subject="Добро пожаловать!",
            body=f"Здравствуйте, {name}! Спасибо за регистрацию."
        )
    
    def _handle_order_placed(self, data):
        order_id = data.get('order_id')
        if not order_id:
            raise ValueError("Отсутствует order_id в данных события")
        
        variables = {
            'order_id': order_id,
            'action': 'process_payment',
        }
        
        if 'customer_id' in data:
            variables['customer_id'] = data['customer_id']
        
        headers = {
            'priority': 'high',
            'order_context': 'true',
        }
        
        job_key = self._create_job(
            job_type="process-payment",
            retries=3,
            variables=variables,
            custom_headers=headers,
            timeout=180000  # 3 минуты
        )
        
        if job_key:
            print(f"💳 Создано задание обработки платежа для заказа {order_id}: {job_key}")
        
        return job_key
    
    def _handle_payment_failed(self, data):
        order_id = data.get('order_id')
        if not order_id:
            raise ValueError("Отсутствует order_id в данных события")
        
        variables = {
            'order_id': order_id,
            'action': 'notify_failure',
        }
        
        if 'failure_reason' in data:
            variables['failure_reason'] = data['failure_reason']
        
        headers = {
            'priority': 'high',
            'notification_type': 'payment_failure',
        }
        
        job_key = self._create_job(
            job_type="send-notification",
            retries=5,
            variables=variables,
            custom_headers=headers,
            timeout=60000
        )
        
        if job_key:
            print(f"❌ Создано задание уведомления о неудачном платеже для заказа {order_id}: {job_key}")
        
        return job_key
    
    def _handle_file_uploaded(self, data):
        file_path = data.get('file_path')
        if not file_path:
            raise ValueError("Отсутствует file_path в данных события")
        
        file_type = data.get('file_type', 'unknown')
        
        # Выбираем приоритет на основе типа файла
        priority = 'normal'
        if file_type in ['image', 'video']:
            priority = 'low'
        elif file_type == 'document':
            priority = 'high'
        
        return self.create_data_processing_job(file_path, file_type, priority)

import threading
import schedule

class JobScheduler:
    def __init__(self, factory: JobFactory):
        self.factory = factory
        self.running = False
        self.thread = None
    
    def start(self):
        """Запускает планировщик заданий"""
        if self.running:
            return
        
        self.running = True
        
        # Настройка расписания
        schedule.every(5).minutes.do(self._create_scheduled_jobs)
        schedule.every().hour.at(":00").do(self._create_hourly_jobs)
        schedule.every().day.at("02:00").do(self._create_daily_jobs)
        schedule.every().sunday.at("03:00").do(self._create_weekly_jobs)
        
        self.thread = threading.Thread(target=self._run_scheduler)
        self.thread.daemon = True
        self.thread.start()
        
        print("⏰ Планировщик заданий запущен")
    
    def stop(self):
        """Останавливает планировщик"""
        self.running = False
        schedule.clear()
        
        if self.thread:
            self.thread.join(timeout=5)
        
        print("⏰ Планировщик заданий остановлен")
    
    def _run_scheduler(self):
        """Основной цикл планировщика"""
        while self.running:
            schedule.run_pending()
            time.sleep(30)
    
    def _create_scheduled_jobs(self):
        """Создает регулярные задания"""
        # Здесь можно добавить логику создания заданий каждые 5 минут
        pass
    
    def _create_hourly_jobs(self):
        """Создает ежечасные задания"""
        print("🕐 Создание ежечасных заданий...")
        
        # Мониторинг системы
        try:
            self.factory.create_api_call_job("http://monitoring/api/health", "GET")
        except Exception as e:
            print(f"⚠️ Ошибка создания задания мониторинга: {e}")
    
    def _create_daily_jobs(self):
        """Создает ежедневные задания"""
        print("📅 Создание ежедневных заданий...")
        
        # Генерация дневного отчета
        yesterday = datetime.now() - timedelta(days=1)
        params = {
            'date': yesterday.strftime('%Y-%m-%d'),
            'region': 'all',
        }
        
        try:
            self.factory.create_report_job("daily_summary", "pdf", params)
        except Exception as e:
            print(f"⚠️ Ошибка создания задания дневного отчета: {e}")
        
        # Очистка временных файлов
        try:
            self.factory.create_data_processing_job("/tmp/cleanup", "directory", "low")
        except Exception as e:
            print(f"⚠️ Ошибка создания задания очистки: {e}")
    
    def _create_weekly_jobs(self):
        """Создает еженедельные задания"""
        print("📊 Создание еженедельных заданий...")
        
        # Генерация недельного отчета
        week_ago = datetime.now() - timedelta(days=7)
        yesterday = datetime.now() - timedelta(days=1)
        
        params = {
            'start_date': week_ago.strftime('%Y-%m-%d'),
            'end_date': yesterday.strftime('%Y-%m-%d'),
            'type': 'comprehensive',
        }
        
        try:
            self.factory.create_report_job("weekly_analytics", "excel", params)
        except Exception as e:
            print(f"⚠️ Ошибка создания задания недельного отчета: {e}")

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Использование:")
        print("  python create_job.py simple <job_type> [retries] [timeout]")
        print("  python create_job.py email <recipient> <subject> <body>")
        print("  python create_job.py api <url> <method> [payload_json]")
        print("  python create_job.py report <type> <format> [params_json]")
        print("  python create_job.py event <event_type> <event_data_json>")
        print("  python create_job.py batch <jobs_json_file>")
        sys.exit(1)
    
    factory = JobFactory()
    command = sys.argv[1]
    
    try:
        if command == "simple":
            job_type = sys.argv[2]
            retries = int(sys.argv[3]) if len(sys.argv) > 3 else 3
            timeout = int(sys.argv[4]) if len(sys.argv) > 4 else 300000
            
            job_key = create_job(job_type, retries, timeout=timeout)
            if job_key:
                print(f"Задание создано: {job_key}")
            
        elif command == "email":
            if len(sys.argv) < 5:
                print("❌ Укажите получателя, тему и текст сообщения")
                sys.exit(1)
            
            recipient = sys.argv[2]
            subject = sys.argv[3]
            body = sys.argv[4]
            
            factory.create_email_job(recipient, subject, body)
            
        elif command == "api":
            if len(sys.argv) < 4:
                print("❌ Укажите URL и метод")
                sys.exit(1)
            
            url = sys.argv[2]
            method = sys.argv[3]
            payload = None
            
            if len(sys.argv) > 4:
                import json
                payload = json.loads(sys.argv[4])
            
            factory.create_api_call_job(url, method, payload)
            
        elif command == "report":
            if len(sys.argv) < 4:
                print("❌ Укажите тип и формат отчета")
                sys.exit(1)
            
            report_type = sys.argv[2]
            format_type = sys.argv[3]
            params = None
            
            if len(sys.argv) > 4:
                import json
                params = json.loads(sys.argv[4])
            
            factory.create_report_job(report_type, format_type, params)
            
        elif command == "event":
            if len(sys.argv) < 4:
                print("❌ Укажите тип события и данные")
                sys.exit(1)
            
            event_type = sys.argv[2]
            import json
            event_data = json.loads(sys.argv[3])
            
            factory.create_job_from_event(event_type, event_data)
            
        elif command == "batch":
            if len(sys.argv) < 3:
                print("❌ Укажите файл с заданиями")
                sys.exit(1)
            
            import json
            with open(sys.argv[2], 'r') as f:
                job_requests = json.load(f)
            
            factory.create_batch_jobs(job_requests)
            
        else:
            print(f"❌ Неизвестная команда: {command}")
            sys.exit(1)
            
    except Exception as e:
        print(f"❌ Ошибка: {e}")
        sys.exit(1)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function createJob(jobType, retries = 3, variables = {}, customHeaders = {}, 
                        timeout = 300000, processInstanceKey = "", elementId = "") {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_type: jobType,
            retries: retries,
            variables: variables,
            custom_headers: customHeaders,
            timeout: timeout,
            process_instance_key: processInstanceKey,
            element_id: elementId
        };
        
        client.createJob(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Задание создано: ${response.job_key}`);
                console.log(`   Время создания: ${response.created_at}`);
                resolve(response.job_key);
            } else {
                console.log(`❌ Ошибка создания: ${response.message}`);
                resolve(null);
            }
        });
    });
}

class JobFactory {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async createDataProcessingJob(filePath, formatType, priority = "normal") {
        const variables = {
            file_path: filePath,
            format: formatType,
            timestamp: new Date().toISOString(),
        };
        
        const headers = {
            priority: priority,
            job_category: 'data-processing',
        };
        
        // Устанавливаем таймаут в зависимости от приоритета
        const timeoutMap = {
            high: 180000,   // 3 минуты
            normal: 300000, // 5 минут
            low: 600000,    // 10 минут
        };
        const timeout = timeoutMap[priority] || 300000;
        
        const jobKey = await this._createJob({
            jobType: "data-processing",
            retries: 5,
            variables: variables,
            customHeaders: headers,
            timeout: timeout
        });
        
        if (jobKey) {
            console.log(`📊 Создано задание обработки данных: ${jobKey}`);
            console.log(`   Файл: ${filePath}`);
            console.log(`   Приоритет: ${priority}`);
        }
        
        return jobKey;
    }
    
    async createEmailJob(recipient, subject, body) {
        const variables = {
            recipient: recipient,
            subject: subject,
            body: body,
            sender: 'system@company.com',
            timestamp: new Date().toISOString(),
        };
        
        const headers = {
            email_type: 'notification',
            priority: 'normal',
        };
        
        const jobKey = await this._createJob({
            jobType: "send-email",
            retries: 3,
            variables: variables,
            customHeaders: headers,
            timeout: 60000  // 1 минута
        });
        
        if (jobKey) {
            console.log(`📧 Создано email задание: ${jobKey}`);
            console.log(`   Получатель: ${recipient}`);
            console.log(`   Тема: ${subject}`);
        }
        
        return jobKey;
    }
    
    async createApiCallJob(url, method, payload = null) {
        const variables = {
            url: url,
            method: method,
            timestamp: new Date().toISOString(),
        };
        
        // Добавляем payload как переменные
        if (payload) {
            Object.entries(payload).forEach(([key, value]) => {
                variables[`payload_${key}`] = String(value);
            });
        }
        
        const headers = {
            api_type: 'external',
            retry_policy: 'exponential_backoff',
        };
        
        // Больше попыток для внешних API
        const retries = method.toUpperCase() !== 'GET' ? 5 : 3;
        
        const jobKey = await this._createJob({
            jobType: "api-call",
            retries: retries,
            variables: variables,
            customHeaders: headers,
            timeout: 120000  // 2 минуты
        });
        
        if (jobKey) {
            console.log(`🌐 Создано API задание: ${jobKey}`);
            console.log(`   URL: ${url}`);
            console.log(`   Метод: ${method}`);
        }
        
        return jobKey;
    }
    
    async createReportJob(reportType, formatType, parameters = null) {
        const variables = {
            report_type: reportType,
            format: formatType,
            generated_at: new Date().toISOString(),
        };
        
        // Добавляем параметры отчета
        if (parameters) {
            Object.entries(parameters).forEach(([key, value]) => {
                variables[`param_${key}`] = String(value);
            });
        }
        
        const headers = {
            report_category: 'analytics',
            priority: 'normal',
        };
        
        // Больше времени для генерации отчетов
        let timeout = 900000;  // 15 минут
        if (reportType === 'complex_analytics') {
            timeout = 1800000;  // 30 минут
        }
        
        const jobKey = await this._createJob({
            jobType: "generate-report",
            retries: 2,
            variables: variables,
            customHeaders: headers,
            timeout: timeout
        });
        
        if (jobKey) {
            console.log(`📊 Создано задание отчета: ${jobKey}`);
            console.log(`   Тип: ${reportType}`);
            console.log(`   Формат: ${formatType}`);
        }
        
        return jobKey;
    }
    
    async createBatchJobs(jobRequests) {
        const jobKeys = [];
        const errors = [];
        
        for (let i = 0; i < jobRequests.length; i++) {
            try {
                const jobKey = await this._createJob(jobRequests[i]);
                if (jobKey) {
                    jobKeys.push(jobKey);
                } else {
                    errors.push(`Задание ${i+1}: Не удалось создать`);
                }
                
                // Небольшая задержка между созданием заданий
                await new Promise(resolve => setTimeout(resolve, 10));
                
            } catch (error) {
                errors.push(`Задание ${i+1}: ${error.message}`);
            }
        }
        
        if (errors.length > 0) {
            console.log(`⚠️ Создано ${jobKeys.length} из ${jobRequests.length} заданий`);
            errors.forEach(error => {
                console.log(`   Ошибка: ${error}`);
            });
        } else {
            console.log(`✅ Успешно создано ${jobKeys.length} заданий`);
        }
        
        return jobKeys;
    }
    
    async _createJob({ jobType, retries = 3, variables = {}, customHeaders = {}, 
                     timeout = 300000, processInstanceKey = "", elementId = "" }) {
        return new Promise((resolve, reject) => {
            const request = {
                job_type: jobType,
                retries: retries,
                variables: variables,
                custom_headers: customHeaders,
                timeout: timeout,
                process_instance_key: processInstanceKey,
                element_id: elementId
            };
            
            this.client.createJob(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    resolve(response.job_key);
                } else {
                    console.log(`❌ Ошибка создания задания: ${response.message}`);
                    resolve(null);
                }
            });
        });
    }
    
    async createJobFromEvent(eventType, eventData) {
        const handlers = {
            'user_registered': this._handleUserRegistered.bind(this),
            'order_placed': this._handleOrderPlaced.bind(this),
            'payment_failed': this._handlePaymentFailed.bind(this),
            'file_uploaded': this._handleFileUploaded.bind(this),
        };
        
        const handler = handlers[eventType];
        if (!handler) {
            throw new Error(`Неизвестный тип события: ${eventType}`);
        }
        
        return await handler(eventData);
    }
    
    async _handleUserRegistered(data) {
        const email = data.email;
        if (!email) {
            throw new Error("Отсутствует email в данных события");
        }
        
        const name = data.name || 'Пользователь';
        
        return await this.createEmailJob(
            email,
            "Добро пожаловать!",
            `Здравствуйте, ${name}! Спасибо за регистрацию.`
        );
    }
    
    async _handleOrderPlaced(data) {
        const orderId = data.order_id;
        if (!orderId) {
            throw new Error("Отсутствует order_id в данных события");
        }
        
        const variables = {
            order_id: orderId,
            action: 'process_payment',
        };
        
        if (data.customer_id) {
            variables.customer_id = data.customer_id;
        }
        
        const headers = {
            priority: 'high',
            order_context: 'true',
        };
        
        const jobKey = await this._createJob({
            jobType: "process-payment",
            retries: 3,
            variables: variables,
            customHeaders: headers,
            timeout: 180000  // 3 минуты
        });
        
        if (jobKey) {
            console.log(`💳 Создано задание обработки платежа для заказа ${orderId}: ${jobKey}`);
        }
        
        return jobKey;
    }
    
    async _handlePaymentFailed(data) {
        const orderId = data.order_id;
        if (!orderId) {
            throw new Error("Отсутствует order_id в данных события");
        }
        
        const variables = {
            order_id: orderId,
            action: 'notify_failure',
        };
        
        if (data.failure_reason) {
            variables.failure_reason = data.failure_reason;
        }
        
        const headers = {
            priority: 'high',
            notification_type: 'payment_failure',
        };
        
        const jobKey = await this._createJob({
            jobType: "send-notification",
            retries: 5,
            variables: variables,
            customHeaders: headers,
            timeout: 60000
        });
        
        if (jobKey) {
            console.log(`❌ Создано задание уведомления о неудачном платеже для заказа ${orderId}: ${jobKey}`);
        }
        
        return jobKey;
    }
    
    async _handleFileUploaded(data) {
        const filePath = data.file_path;
        if (!filePath) {
            throw new Error("Отсутствует file_path в данных события");
        }
        
        const fileType = data.file_type || 'unknown';
        
        // Выбираем приоритет на основе типа файла
        let priority = 'normal';
        if (['image', 'video'].includes(fileType)) {
            priority = 'low';
        } else if (fileType === 'document') {
            priority = 'high';
        }
        
        return await this.createDataProcessingJob(filePath, fileType, priority);
    }
}

class JobScheduler {
    constructor(factory) {
        this.factory = factory;
        this.running = false;
        this.intervals = [];
    }
    
    start() {
        if (this.running) return;
        
        this.running = true;
        
        // Настройка расписания
        this.intervals.push(setInterval(() => this._createScheduledJobs(), 5 * 60 * 1000)); // Каждые 5 минут
        this.intervals.push(setInterval(() => this._createHourlyJobs(), 60 * 60 * 1000)); // Каждый час
        this.intervals.push(setInterval(() => this._createDailyJobs(), 24 * 60 * 60 * 1000)); // Каждый день
        
        console.log("⏰ Планировщик заданий запущен");
    }
    
    stop() {
        this.running = false;
        
        this.intervals.forEach(interval => clearInterval(interval));
        this.intervals = [];
        
        console.log("⏰ Планировщик заданий остановлен");
    }
    
    async _createScheduledJobs() {
        // Здесь можно добавить логику создания заданий каждые 5 минут
    }
    
    async _createHourlyJobs() {
        console.log("🕐 Создание ежечасных заданий...");
        
        // Мониторинг системы
        try {
            await this.factory.createApiCallJob("http://monitoring/api/health", "GET");
        } catch (error) {
            console.log(`⚠️ Ошибка создания задания мониторинга: ${error.message}`);
        }
    }
    
    async _createDailyJobs() {
        console.log("📅 Создание ежедневных заданий...");
        
        // Генерация дневного отчета
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        
        const params = {
            date: yesterday.toISOString().split('T')[0],
            region: 'all',
        };
        
        try {
            await this.factory.createReportJob("daily_summary", "pdf", params);
        } catch (error) {
            console.log(`⚠️ Ошибка создания задания дневного отчета: ${error.message}`);
        }
        
        // Очистка временных файлов
        try {
            await this.factory.createDataProcessingJob("/tmp/cleanup", "directory", "low");
        } catch (error) {
            console.log(`⚠️ Ошибка создания задания очистки: ${error.message}`);
        }
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node create-job.js simple <job_type> [retries] [timeout]');
        console.log('  node create-job.js email <recipient> <subject> <body>');
        console.log('  node create-job.js api <url> <method> [payload_json]');
        console.log('  node create-job.js report <type> <format> [params_json]');
        console.log('  node create-job.js event <event_type> <event_data_json>');
        console.log('  node create-job.js batch <jobs_json_file>');
        process.exit(1);
    }
    
    const factory = new JobFactory();
    const command = args[0];
    
    (async () => {
        try {
            switch (command) {
                case 'simple':
                    const jobType = args[1];
                    const retries = args[2] ? parseInt(args[2]) : 3;
                    const timeout = args[3] ? parseInt(args[3]) : 300000;
                    
                    const jobKey = await createJob(jobType, retries, {}, {}, timeout);
                    if (jobKey) {
                        console.log(`Задание создано: ${jobKey}`);
                    }
                    break;
                    
                case 'email':
                    if (args.length < 4) {
                        console.log('❌ Укажите получателя, тему и текст сообщения');
                        process.exit(1);
                    }
                    
                    await factory.createEmailJob(args[1], args[2], args[3]);
                    break;
                    
                case 'api':
                    if (args.length < 3) {
                        console.log('❌ Укажите URL и метод');
                        process.exit(1);
                    }
                    
                    const url = args[1];
                    const method = args[2];
                    const payload = args[3] ? JSON.parse(args[3]) : null;
                    
                    await factory.createApiCallJob(url, method, payload);
                    break;
                    
                case 'report':
                    if (args.length < 3) {
                        console.log('❌ Укажите тип и формат отчета');
                        process.exit(1);
                    }
                    
                    const reportType = args[1];
                    const formatType = args[2];
                    const params = args[3] ? JSON.parse(args[3]) : null;
                    
                    await factory.createReportJob(reportType, formatType, params);
                    break;
                    
                case 'event':
                    if (args.length < 3) {
                        console.log('❌ Укажите тип события и данные');
                        process.exit(1);
                    }
                    
                    const eventType = args[1];
                    const eventData = JSON.parse(args[2]);
                    
                    await factory.createJobFromEvent(eventType, eventData);
                    break;
                    
                case 'batch':
                    if (args.length < 2) {
                        console.log('❌ Укажите файл с заданиями');
                        process.exit(1);
                    }
                    
                    const fs = require('fs');
                    const jobRequests = JSON.parse(fs.readFileSync(args[1], 'utf8'));
                    
                    await factory.createBatchJobs(jobRequests);
                    break;
                    
                default:
                    console.log(`❌ Неизвестная команда: ${command}`);
                    process.exit(1);
            }
        } catch (error) {
            console.error(`❌ Ошибка: ${error.message}`);
            process.exit(1);
        }
    })();
}

module.exports = {
    createJob,
    JobFactory,
    JobScheduler
};
```

## Области применения

### Программное создание заданий
- **Обработка данных**: Создание заданий для batch processing
- **Уведомления**: Email, SMS, push-уведомления
- **Интеграции**: Вызовы внешних API
- **Отчеты**: Генерация аналитических отчетов

### Событийно-ориентированная архитектура
- **Реакция на события**: Автоматическое создание заданий
- **Асинхронная обработка**: Отложенное выполнение операций
- **Микросервисы**: Координация между сервисами

### Планирование задач
- **Регулярные операции**: Ежечасные, ежедневные, еженедельные задания
- **Мониторинг**: Проверки состояния системы
- **Обслуживание**: Очистка, архивирование, резервное копирование

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры задания
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Invalid job_type: must not be empty",
  "job_key": "",
  "created_at": ""
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Активация созданных заданий
- [ListJobs](list-jobs.md) - Просмотр созданных заданий
- [GetJob](get-job.md) - Получение деталей задания
- [CancelJob](cancel-job.md) - Отмена созданного задания
