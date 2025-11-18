# GetJob

## Описание
Получает детальную информацию о конкретном задании по его ключу. Возвращает полные данные включая переменные, заголовки и метаданные.

## Синтаксис
```protobuf
rpc GetJob(GetJobRequest) returns (GetJobResponse);
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

### GetJobRequest
```protobuf
message GetJobRequest {
  string job_key = 1;    // Ключ задания
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания

## Параметры ответа

### GetJobResponse
```protobuf
message GetJobResponse {
  bool success = 1;           // Статус успешности операции
  string message = 2;         // Сообщение о результате
  Job job = 3;               // Данные задания
}

message Job {
  string job_key = 1;                      // Ключ задания
  string job_type = 2;                     // Тип задания
  string state = 3;                        // Состояние задания
  string worker = 4;                       // Имя воркера
  int32 retries = 5;                       // Количество попыток
  string created_at = 6;                   // Время создания (RFC3339)
  string activated_at = 7;                 // Время активации (RFC3339)
  string completed_at = 8;                 // Время завершения (RFC3339)
  string deadline = 9;                     // Крайний срок (RFC3339)
  string process_instance_key = 10;        // Ключ экземпляра процесса
  string element_id = 11;                  // ID элемента BPMN
  map<string, string> variables = 12;      // Переменные задания
  map<string, string> custom_headers = 13; // Пользовательские заголовки
  string error_message = 14;               // Сообщение об ошибке (если есть)
  int64 timeout = 15;                      // Таймаут в миллисекундах
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
    
    // Получаем детали задания
    response, err := client.GetJob(ctx, &pb.GetJobRequest{
        JobKey: jobKey,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        job := response.Job
        fmt.Printf("📋 Информация о задании %s:\n\n", jobKey)
        
        // Основная информация
        fmt.Printf("Тип: %s\n", job.JobType)
        fmt.Printf("Состояние: %s\n", job.State)
        fmt.Printf("Воркер: %s\n", job.Worker)
        fmt.Printf("Попытки: %d\n", job.Retries)
        
        // Временные метки
        createdAt, _ := time.Parse(time.RFC3339, job.CreatedAt)
        fmt.Printf("Создано: %s\n", createdAt.Format("2006-01-02 15:04:05"))
        
        if job.ActivatedAt != "" {
            activatedAt, _ := time.Parse(time.RFC3339, job.ActivatedAt)
            fmt.Printf("Активировано: %s\n", activatedAt.Format("2006-01-02 15:04:05"))
        }
        
        if job.CompletedAt != "" {
            completedAt, _ := time.Parse(time.RFC3339, job.CompletedAt)
            fmt.Printf("Завершено: %s\n", completedAt.Format("2006-01-02 15:04:05"))
        }
        
        if job.Deadline != "" {
            deadline, _ := time.Parse(time.RFC3339, job.Deadline)
            fmt.Printf("Крайний срок: %s\n", deadline.Format("2006-01-02 15:04:05"))
        }
        
        // Процесс и элемент
        fmt.Printf("Экземпляр процесса: %s\n", job.ProcessInstanceKey)
        fmt.Printf("Элемент BPMN: %s\n", job.ElementId)
        
        // Переменные
        if len(job.Variables) > 0 {
            fmt.Printf("\nПеременные:\n")
            for key, value := range job.Variables {
                fmt.Printf("  %s: %s\n", key, value)
            }
        }
        
        // Заголовки
        if len(job.CustomHeaders) > 0 {
            fmt.Printf("\nЗаголовки:\n")
            for key, value := range job.CustomHeaders {
                fmt.Printf("  %s: %s\n", key, value)
            }
        }
        
        // Ошибка
        if job.ErrorMessage != "" {
            fmt.Printf("\nОшибка: %s\n", job.ErrorMessage)
        }
        
        // Таймаут
        if job.Timeout > 0 {
            timeout := time.Duration(job.Timeout) * time.Millisecond
            fmt.Printf("Таймаут: %s\n", timeout.String())
        }
    } else {
        fmt.Printf("❌ Ошибка получения задания: %s\n", response.Message)
    }
}

// Менеджер для работы с деталями заданий
type JobDetailsManager struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewJobDetailsManager(client pb.JobsServiceClient, ctx context.Context) *JobDetailsManager {
    return &JobDetailsManager{
        client: client,
        ctx:    ctx,
    }
}

func (jdm *JobDetailsManager) GetJob(jobKey string) (*pb.Job, error) {
    response, err := jdm.client.GetJob(jdm.ctx, &pb.GetJobRequest{
        JobKey: jobKey,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("задание не найдено: %s", response.Message)
    }
    
    return response.Job, nil
}

func (jdm *JobDetailsManager) PrintJobDetails(jobKey string) error {
    job, err := jdm.GetJob(jobKey)
    if err != nil {
        return err
    }
    
    fmt.Printf("📋 Детали задания %s:\n", jobKey)
    fmt.Printf("═══════════════════════════════════════\n")
    
    // Статус и тип
    fmt.Printf("🏷️  Тип: %s\n", job.JobType)
    fmt.Printf("📊 Состояние: %s\n", jdm.formatState(job.State))
    
    if job.Worker != "" {
        fmt.Printf("👤 Воркер: %s\n", job.Worker)
    }
    
    fmt.Printf("🔄 Попытки: %d\n", job.Retries)
    
    // Временная линия
    fmt.Printf("\n⏰ Временная линия:\n")
    if job.CreatedAt != "" {
        createdAt, _ := time.Parse(time.RFC3339, job.CreatedAt)
        age := time.Since(createdAt)
        fmt.Printf("   📅 Создано: %s (%s назад)\n", 
                   createdAt.Format("2006-01-02 15:04:05"), age.Round(time.Second))
    }
    
    if job.ActivatedAt != "" {
        activatedAt, _ := time.Parse(time.RFC3339, job.ActivatedAt)
        fmt.Printf("   ▶️  Активировано: %s\n", activatedAt.Format("2006-01-02 15:04:05"))
        
        if job.CreatedAt != "" {
            createdAt, _ := time.Parse(time.RFC3339, job.CreatedAt)
            waitTime := activatedAt.Sub(createdAt)
            fmt.Printf("       (время ожидания: %s)\n", waitTime.Round(time.Second))
        }
    }
    
    if job.CompletedAt != "" {
        completedAt, _ := time.Parse(time.RFC3339, job.CompletedAt)
        fmt.Printf("   ✅ Завершено: %s\n", completedAt.Format("2006-01-02 15:04:05"))
        
        if job.ActivatedAt != "" {
            activatedAt, _ := time.Parse(time.RFC3339, job.ActivatedAt)
            execTime := completedAt.Sub(activatedAt)
            fmt.Printf("       (время выполнения: %s)\n", execTime.Round(time.Second))
        }
    }
    
    if job.Deadline != "" {
        deadline, _ := time.Parse(time.RFC3339, job.Deadline)
        fmt.Printf("   ⏱️  Крайний срок: %s\n", deadline.Format("2006-01-02 15:04:05"))
        
        if time.Now().After(deadline) {
            fmt.Printf("       ⚠️ ПРОСРОЧЕНО на %s\n", time.Since(deadline).Round(time.Second))
        } else {
            fmt.Printf("       ⏳ Осталось: %s\n", time.Until(deadline).Round(time.Second))
        }
    }
    
    // Контекст процесса
    fmt.Printf("\n🔗 Контекст процесса:\n")
    fmt.Printf("   📍 Экземпляр: %s\n", job.ProcessInstanceKey)
    fmt.Printf("   🎯 Элемент BPMN: %s\n", job.ElementId)
    
    // Переменные
    if len(job.Variables) > 0 {
        fmt.Printf("\n📦 Переменные (%d):\n", len(job.Variables))
        for key, value := range job.Variables {
            fmt.Printf("   %s = %s\n", key, jdm.formatValue(value))
        }
    }
    
    // Заголовки
    if len(job.CustomHeaders) > 0 {
        fmt.Printf("\n📋 Заголовки (%d):\n", len(job.CustomHeaders))
        for key, value := range job.CustomHeaders {
            fmt.Printf("   %s: %s\n", key, value)
        }
    }
    
    // Настройки
    fmt.Printf("\n⚙️ Настройки:\n")
    if job.Timeout > 0 {
        timeout := time.Duration(job.Timeout) * time.Millisecond
        fmt.Printf("   ⏰ Таймаут: %s\n", timeout.String())
    } else {
        fmt.Printf("   ⏰ Таймаут: не установлен\n")
    }
    
    // Ошибка
    if job.ErrorMessage != "" {
        fmt.Printf("\n❌ Ошибка:\n")
        fmt.Printf("   %s\n", job.ErrorMessage)
    }
    
    return nil
}

func (jdm *JobDetailsManager) formatState(state string) string {
    stateEmojis := map[string]string{
        "ACTIVATABLE": "🟡 ACTIVATABLE (готово к активации)",
        "ACTIVATED":   "🔵 ACTIVATED (выполняется)",
        "COMPLETED":   "🟢 COMPLETED (завершено)",
        "FAILED":      "🔴 FAILED (провалено)",
        "CANCELLED":   "⚫ CANCELLED (отменено)",
    }
    
    if formatted, exists := stateEmojis[state]; exists {
        return formatted
    }
    return state
}

func (jdm *JobDetailsManager) formatValue(value string) string {
    if len(value) > 50 {
        return value[:47] + "..."
    }
    return value
}

func (jdm *JobDetailsManager) CompareJobs(jobKey1, jobKey2 string) error {
    job1, err := jdm.GetJob(jobKey1)
    if err != nil {
        return fmt.Errorf("ошибка получения первого задания: %v", err)
    }
    
    job2, err := jdm.GetJob(jobKey2)
    if err != nil {
        return fmt.Errorf("ошибка получения второго задания: %v", err)
    }
    
    fmt.Printf("📊 Сравнение заданий:\n")
    fmt.Printf("═══════════════════════════════════════\n")
    fmt.Printf("%-20s | %-30s | %-30s\n", "Параметр", jobKey1, jobKey2)
    fmt.Printf("═══════════════════════════════════════\n")
    
    // Сравниваем основные параметры
    jdm.compareField("Тип", job1.JobType, job2.JobType)
    jdm.compareField("Состояние", job1.State, job2.State)
    jdm.compareField("Воркер", job1.Worker, job2.Worker)
    jdm.compareField("Попытки", fmt.Sprintf("%d", job1.Retries), fmt.Sprintf("%d", job2.Retries))
    jdm.compareField("Процесс", job1.ProcessInstanceKey, job2.ProcessInstanceKey)
    jdm.compareField("Элемент", job1.ElementId, job2.ElementId)
    
    // Сравниваем времена
    fmt.Printf("───────────────────────────────────────\n")
    jdm.compareTimeField("Создано", job1.CreatedAt, job2.CreatedAt)
    jdm.compareTimeField("Активировано", job1.ActivatedAt, job2.ActivatedAt)
    jdm.compareTimeField("Завершено", job1.CompletedAt, job2.CompletedAt)
    
    return nil
}

func (jdm *JobDetailsManager) compareField(name, val1, val2 string) {
    equal := "✅"
    if val1 != val2 {
        equal = "❌"
    }
    
    fmt.Printf("%-20s | %-30s | %-30s %s\n", name, jdm.truncate(val1, 30), jdm.truncate(val2, 30), equal)
}

func (jdm *JobDetailsManager) compareTimeField(name, time1, time2 string) {
    format1 := jdm.formatTime(time1)
    format2 := jdm.formatTime(time2)
    
    equal := "✅"
    if time1 != time2 {
        equal = "❌"
    }
    
    fmt.Printf("%-20s | %-30s | %-30s %s\n", name, format1, format2, equal)
}

func (jdm *JobDetailsManager) formatTime(timeStr string) string {
    if timeStr == "" {
        return "-"
    }
    
    t, err := time.Parse(time.RFC3339, timeStr)
    if err != nil {
        return timeStr
    }
    
    return t.Format("2006-01-02 15:04:05")
}

func (jdm *JobDetailsManager) truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

// Мониторинг задания в реальном времени
func (jdm *JobDetailsManager) MonitorJob(jobKey string, interval time.Duration) {
    fmt.Printf("🔍 Мониторинг задания %s (обновление каждые %s)\n", jobKey, interval)
    fmt.Printf("Нажмите Ctrl+C для остановки\n\n")
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    lastState := ""
    lastRetries := int32(-1)
    
    for range ticker.C {
        job, err := jdm.GetJob(jobKey)
        if err != nil {
            fmt.Printf("❌ Ошибка получения задания: %v\n", err)
            continue
        }
        
        now := time.Now().Format("15:04:05")
        
        // Отслеживаем изменения состояния
        if job.State != lastState {
            fmt.Printf("[%s] 📊 Состояние: %s → %s\n", now, lastState, job.State)
            lastState = job.State
        }
        
        // Отслеживаем изменения количества попыток
        if job.Retries != lastRetries {
            if lastRetries != -1 {
                fmt.Printf("[%s] 🔄 Попытки: %d → %d\n", now, lastRetries, job.Retries)
            }
            lastRetries = job.Retries
        }
        
        // Показываем текущий статус
        fmt.Printf("[%s] %s | Попытки: %d | Воркер: %s\n", 
                   now, jdm.formatState(job.State), job.Retries, job.Worker)
        
        // Если задание завершено, останавливаем мониторинг
        if job.State == "COMPLETED" || job.State == "FAILED" || job.State == "CANCELLED" {
            fmt.Printf("\n✅ Задание завершено со статусом: %s\n", job.State)
            break
        }
    }
}

// Экспорт данных задания
func (jdm *JobDetailsManager) ExportJobToJSON(jobKey string) ([]byte, error) {
    job, err := jdm.GetJob(jobKey)
    if err != nil {
        return nil, err
    }
    
    // Преобразуем в более удобный формат для JSON
    export := map[string]interface{}{
        "job_key":              job.JobKey,
        "job_type":             job.JobType,
        "state":                job.State,
        "worker":               job.Worker,
        "retries":              job.Retries,
        "created_at":           job.CreatedAt,
        "activated_at":         job.ActivatedAt,
        "completed_at":         job.CompletedAt,
        "deadline":             job.Deadline,
        "process_instance_key": job.ProcessInstanceKey,
        "element_id":           job.ElementId,
        "variables":            job.Variables,
        "custom_headers":       job.CustomHeaders,
        "error_message":        job.ErrorMessage,
        "timeout":              job.Timeout,
    }
    
    return json.Marshal(export)
}
```

### Python
```python
import grpc
import json
import time
from datetime import datetime, timezone
from typing import Optional, Dict, Any

import jobs_pb2
import jobs_pb2_grpc

def get_job(job_key):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.GetJobRequest(job_key=job_key)
    
    try:
        response = stub.GetJob(request, metadata=metadata)
        
        if response.success:
            job = response.job
            print(f"📋 Информация о задании {job_key}:\n")
            
            # Основная информация
            print(f"Тип: {job.job_type}")
            print(f"Состояние: {job.state}")
            print(f"Воркер: {job.worker}")
            print(f"Попытки: {job.retries}")
            
            # Временные метки
            if job.created_at:
                created_at = datetime.fromisoformat(job.created_at.replace('Z', '+00:00'))
                print(f"Создано: {created_at.strftime('%Y-%m-%d %H:%M:%S')}")
            
            if job.activated_at:
                activated_at = datetime.fromisoformat(job.activated_at.replace('Z', '+00:00'))
                print(f"Активировано: {activated_at.strftime('%Y-%m-%d %H:%M:%S')}")
            
            if job.completed_at:
                completed_at = datetime.fromisoformat(job.completed_at.replace('Z', '+00:00'))
                print(f"Завершено: {completed_at.strftime('%Y-%m-%d %H:%M:%S')}")
            
            if job.deadline:
                deadline = datetime.fromisoformat(job.deadline.replace('Z', '+00:00'))
                print(f"Крайний срок: {deadline.strftime('%Y-%m-%d %H:%M:%S')}")
            
            # Процесс и элемент
            print(f"Экземпляр процесса: {job.process_instance_key}")
            print(f"Элемент BPMN: {job.element_id}")
            
            # Переменные
            if job.variables:
                print("\nПеременные:")
                for key, value in job.variables.items():
                    print(f"  {key}: {value}")
            
            # Заголовки
            if job.custom_headers:
                print("\nЗаголовки:")
                for key, value in job.custom_headers.items():
                    print(f"  {key}: {value}")
            
            # Ошибка
            if job.error_message:
                print(f"\nОшибка: {job.error_message}")
            
            # Таймаут
            if job.timeout > 0:
                timeout_sec = job.timeout / 1000
                print(f"Таймаут: {timeout_sec}с")
            
            return job
        else:
            print(f"❌ Ошибка получения задания: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

class JobDetailsManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def get_job(self, job_key):
        """Получает детали задания"""
        try:
            request = jobs_pb2.GetJobRequest(job_key=job_key)
            response = self.stub.GetJob(request, metadata=self.metadata)
            
            if response.success:
                return response.job
            else:
                raise Exception(f"Задание не найдено: {response.message}")
                
        except grpc.RpcError as e:
            raise Exception(f"gRPC Error: {e.details()}")
    
    def print_job_details(self, job_key):
        """Выводит детальную информацию о задании"""
        job = self.get_job(job_key)
        
        print(f"📋 Детали задания {job_key}:")
        print("═" * 40)
        
        # Статус и тип
        print(f"🏷️  Тип: {job.job_type}")
        print(f"📊 Состояние: {self._format_state(job.state)}")
        
        if job.worker:
            print(f"👤 Воркер: {job.worker}")
        
        print(f"🔄 Попытки: {job.retries}")
        
        # Временная линия
        print("\n⏰ Временная линия:")
        if job.created_at:
            created_at = datetime.fromisoformat(job.created_at.replace('Z', '+00:00'))
            age = datetime.now(timezone.utc) - created_at
            print(f"   📅 Создано: {created_at.strftime('%Y-%m-%d %H:%M:%S')} ({self._format_duration(age)} назад)")
        
        if job.activated_at:
            activated_at = datetime.fromisoformat(job.activated_at.replace('Z', '+00:00'))
            print(f"   ▶️  Активировано: {activated_at.strftime('%Y-%m-%d %H:%M:%S')}")
            
            if job.created_at:
                created_at = datetime.fromisoformat(job.created_at.replace('Z', '+00:00'))
                wait_time = activated_at - created_at
                print(f"       (время ожидания: {self._format_duration(wait_time)})")
        
        if job.completed_at:
            completed_at = datetime.fromisoformat(job.completed_at.replace('Z', '+00:00'))
            print(f"   ✅ Завершено: {completed_at.strftime('%Y-%m-%d %H:%M:%S')}")
            
            if job.activated_at:
                activated_at = datetime.fromisoformat(job.activated_at.replace('Z', '+00:00'))
                exec_time = completed_at - activated_at
                print(f"       (время выполнения: {self._format_duration(exec_time)})")
        
        if job.deadline:
            deadline = datetime.fromisoformat(job.deadline.replace('Z', '+00:00'))
            print(f"   ⏱️  Крайний срок: {deadline.strftime('%Y-%m-%d %H:%M:%S')}")
            
            now = datetime.now(timezone.utc)
            if now > deadline:
                overdue = now - deadline
                print(f"       ⚠️ ПРОСРОЧЕНО на {self._format_duration(overdue)}")
            else:
                remaining = deadline - now
                print(f"       ⏳ Осталось: {self._format_duration(remaining)}")
        
        # Контекст процесса
        print("\n🔗 Контекст процесса:")
        print(f"   📍 Экземпляр: {job.process_instance_key}")
        print(f"   🎯 Элемент BPMN: {job.element_id}")
        
        # Переменные
        if job.variables:
            print(f"\n📦 Переменные ({len(job.variables)}):")
            for key, value in job.variables.items():
                print(f"   {key} = {self._format_value(value)}")
        
        # Заголовки
        if job.custom_headers:
            print(f"\n📋 Заголовки ({len(job.custom_headers)}):")
            for key, value in job.custom_headers.items():
                print(f"   {key}: {value}")
        
        # Настройки
        print("\n⚙️ Настройки:")
        if job.timeout > 0:
            timeout_sec = job.timeout / 1000
            print(f"   ⏰ Таймаут: {timeout_sec}с")
        else:
            print("   ⏰ Таймаут: не установлен")
        
        # Ошибка
        if job.error_message:
            print(f"\n❌ Ошибка:")
            print(f"   {job.error_message}")
    
    def _format_state(self, state):
        state_emojis = {
            "ACTIVATABLE": "🟡 ACTIVATABLE (готово к активации)",
            "ACTIVATED": "🔵 ACTIVATED (выполняется)",
            "COMPLETED": "🟢 COMPLETED (завершено)",
            "FAILED": "🔴 FAILED (провалено)",
            "CANCELLED": "⚫ CANCELLED (отменено)",
        }
        return state_emojis.get(state, state)
    
    def _format_value(self, value):
        if len(value) > 50:
            return value[:47] + "..."
        return value
    
    def _format_duration(self, delta):
        total_seconds = int(delta.total_seconds())
        
        if total_seconds < 60:
            return f"{total_seconds}с"
        elif total_seconds < 3600:
            minutes = total_seconds // 60
            seconds = total_seconds % 60
            return f"{minutes}м {seconds}с"
        else:
            hours = total_seconds // 3600
            minutes = (total_seconds % 3600) // 60
            return f"{hours}ч {minutes}м"
    
    def compare_jobs(self, job_key1, job_key2):
        """Сравнивает два задания"""
        job1 = self.get_job(job_key1)
        job2 = self.get_job(job_key2)
        
        print("📊 Сравнение заданий:")
        print("═" * 40)
        print(f"{'Параметр':<20} | {job_key1:<30} | {job_key2:<30}")
        print("═" * 40)
        
        # Сравниваем основные параметры
        self._compare_field("Тип", job1.job_type, job2.job_type)
        self._compare_field("Состояние", job1.state, job2.state)
        self._compare_field("Воркер", job1.worker, job2.worker)
        self._compare_field("Попытки", str(job1.retries), str(job2.retries))
        self._compare_field("Процесс", job1.process_instance_key, job2.process_instance_key)
        self._compare_field("Элемент", job1.element_id, job2.element_id)
        
        # Сравниваем времена
        print("─" * 40)
        self._compare_time_field("Создано", job1.created_at, job2.created_at)
        self._compare_time_field("Активировано", job1.activated_at, job2.activated_at)
        self._compare_time_field("Завершено", job1.completed_at, job2.completed_at)
    
    def _compare_field(self, name, val1, val2):
        equal = "✅" if val1 == val2 else "❌"
        print(f"{name:<20} | {self._truncate(val1, 30):<30} | {self._truncate(val2, 30):<30} {equal}")
    
    def _compare_time_field(self, name, time1, time2):
        format1 = self._format_time(time1)
        format2 = self._format_time(time2)
        equal = "✅" if time1 == time2 else "❌"
        print(f"{name:<20} | {format1:<30} | {format2:<30} {equal}")
    
    def _format_time(self, time_str):
        if not time_str:
            return "-"
        
        try:
            t = datetime.fromisoformat(time_str.replace('Z', '+00:00'))
            return t.strftime('%Y-%m-%d %H:%M:%S')
        except:
            return time_str
    
    def _truncate(self, s, max_len):
        if len(s) <= max_len:
            return s
        return s[:max_len-3] + "..."
    
    def monitor_job(self, job_key, interval=5):
        """Мониторинг задания в реальном времени"""
        print(f"🔍 Мониторинг задания {job_key} (обновление каждые {interval}с)")
        print("Нажмите Ctrl+C для остановки\n")
        
        last_state = ""
        last_retries = -1
        
        try:
            while True:
                try:
                    job = self.get_job(job_key)
                    now = datetime.now().strftime("%H:%M:%S")
                    
                    # Отслеживаем изменения состояния
                    if job.state != last_state:
                        print(f"[{now}] 📊 Состояние: {last_state} → {job.state}")
                        last_state = job.state
                    
                    # Отслеживаем изменения количества попыток
                    if job.retries != last_retries:
                        if last_retries != -1:
                            print(f"[{now}] 🔄 Попытки: {last_retries} → {job.retries}")
                        last_retries = job.retries
                    
                    # Показываем текущий статус
                    print(f"[{now}] {self._format_state(job.state)} | Попытки: {job.retries} | Воркер: {job.worker}")
                    
                    # Если задание завершено, останавливаем мониторинг
                    if job.state in ["COMPLETED", "FAILED", "CANCELLED"]:
                        print(f"\n✅ Задание завершено со статусом: {job.state}")
                        break
                    
                except Exception as e:
                    print(f"❌ Ошибка получения задания: {e}")
                
                time.sleep(interval)
                
        except KeyboardInterrupt:
            print("\n🛑 Мониторинг остановлен пользователем")
    
    def export_job_to_json(self, job_key):
        """Экспорт данных задания в JSON"""
        job = self.get_job(job_key)
        
        export = {
            "job_key": job.job_key,
            "job_type": job.job_type,
            "state": job.state,
            "worker": job.worker,
            "retries": job.retries,
            "created_at": job.created_at,
            "activated_at": job.activated_at,
            "completed_at": job.completed_at,
            "deadline": job.deadline,
            "process_instance_key": job.process_instance_key,
            "element_id": job.element_id,
            "variables": dict(job.variables),
            "custom_headers": dict(job.custom_headers),
            "error_message": job.error_message,
            "timeout": job.timeout,
        }
        
        return json.dumps(export, indent=2, ensure_ascii=False)

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Использование:")
        print("  python get_job.py show <job_key>")
        print("  python get_job.py compare <job_key1> <job_key2>")
        print("  python get_job.py monitor <job_key> [interval]")
        print("  python get_job.py export <job_key>")
        sys.exit(1)
    
    manager = JobDetailsManager()
    command = sys.argv[1]
    
    if command == "show":
        job_key = sys.argv[2]
        manager.print_job_details(job_key)
        
    elif command == "compare":
        if len(sys.argv) < 4:
            print("❌ Укажите два ключа заданий для сравнения")
            sys.exit(1)
        
        job_key1 = sys.argv[2]
        job_key2 = sys.argv[3]
        manager.compare_jobs(job_key1, job_key2)
        
    elif command == "monitor":
        job_key = sys.argv[2]
        interval = int(sys.argv[3]) if len(sys.argv) > 3 else 5
        manager.monitor_job(job_key, interval)
        
    elif command == "export":
        job_key = sys.argv[2]
        json_data = manager.export_job_to_json(job_key)
        print(json_data)
        
    else:
        print(f"❌ Неизвестная команда: {command}")
        sys.exit(1)
```

### JavaScript/Node.js  
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function getJob(jobKey) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { job_key: jobKey };
        
        client.getJob(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                const job = response.job;
                console.log(`📋 Информация о задании ${jobKey}:\n`);
                
                // Основная информация
                console.log(`Тип: ${job.job_type}`);
                console.log(`Состояние: ${job.state}`);
                console.log(`Воркер: ${job.worker}`);
                console.log(`Попытки: ${job.retries}`);
                
                // Временные метки
                if (job.created_at) {
                    const createdAt = new Date(job.created_at);
                    console.log(`Создано: ${createdAt.toLocaleString()}`);
                }
                
                if (job.activated_at) {
                    const activatedAt = new Date(job.activated_at);
                    console.log(`Активировано: ${activatedAt.toLocaleString()}`);
                }
                
                if (job.completed_at) {
                    const completedAt = new Date(job.completed_at);
                    console.log(`Завершено: ${completedAt.toLocaleString()}`);
                }
                
                if (job.deadline) {
                    const deadline = new Date(job.deadline);
                    console.log(`Крайний срок: ${deadline.toLocaleString()}`);
                }
                
                // Процесс и элемент
                console.log(`Экземпляр процесса: ${job.process_instance_key}`);
                console.log(`Элемент BPMN: ${job.element_id}`);
                
                // Переменные
                if (Object.keys(job.variables).length > 0) {
                    console.log('\nПеременные:');
                    Object.entries(job.variables).forEach(([key, value]) => {
                        console.log(`  ${key}: ${value}`);
                    });
                }
                
                // Заголовки
                if (Object.keys(job.custom_headers).length > 0) {
                    console.log('\nЗаголовки:');
                    Object.entries(job.custom_headers).forEach(([key, value]) => {
                        console.log(`  ${key}: ${value}`);
                    });
                }
                
                // Ошибка
                if (job.error_message) {
                    console.log(`\nОшибка: ${job.error_message}`);
                }
                
                // Таймаут
                if (job.timeout > 0) {
                    const timeoutSec = job.timeout / 1000;
                    console.log(`Таймаут: ${timeoutSec}с`);
                }
                
                resolve(job);
            } else {
                console.log(`❌ Ошибка получения задания: ${response.message}`);
                resolve(null);
            }
        });
    });
}

class JobDetailsManager {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async getJob(jobKey) {
        return new Promise((resolve, reject) => {
            const request = { job_key: jobKey };
            
            this.client.getJob(request, this.metadata, (error, response) => {
                if (error) {
                    reject(new Error(`gRPC Error: ${error.message}`));
                    return;
                }
                
                if (response.success) {
                    resolve(response.job);
                } else {
                    reject(new Error(`Задание не найдено: ${response.message}`));
                }
            });
        });
    }
    
    async printJobDetails(jobKey) {
        const job = await this.getJob(jobKey);
        
        console.log(`📋 Детали задания ${jobKey}:`);
        console.log('═'.repeat(40));
        
        // Статус и тип
        console.log(`🏷️  Тип: ${job.job_type}`);
        console.log(`📊 Состояние: ${this._formatState(job.state)}`);
        
        if (job.worker) {
            console.log(`👤 Воркер: ${job.worker}`);
        }
        
        console.log(`🔄 Попытки: ${job.retries}`);
        
        // Временная линия
        console.log('\n⏰ Временная линия:');
        if (job.created_at) {
            const createdAt = new Date(job.created_at);
            const age = Date.now() - createdAt.getTime();
            console.log(`   📅 Создано: ${createdAt.toLocaleString()} (${this._formatDuration(age)} назад)`);
        }
        
        if (job.activated_at) {
            const activatedAt = new Date(job.activated_at);
            console.log(`   ▶️  Активировано: ${activatedAt.toLocaleString()}`);
            
            if (job.created_at) {
                const createdAt = new Date(job.created_at);
                const waitTime = activatedAt.getTime() - createdAt.getTime();
                console.log(`       (время ожидания: ${this._formatDuration(waitTime)})`);
            }
        }
        
        if (job.completed_at) {
            const completedAt = new Date(job.completed_at);
            console.log(`   ✅ Завершено: ${completedAt.toLocaleString()}`);
            
            if (job.activated_at) {
                const activatedAt = new Date(job.activated_at);
                const execTime = completedAt.getTime() - activatedAt.getTime();
                console.log(`       (время выполнения: ${this._formatDuration(execTime)})`);
            }
        }
        
        if (job.deadline) {
            const deadline = new Date(job.deadline);
            console.log(`   ⏱️  Крайний срок: ${deadline.toLocaleString()}`);
            
            const now = Date.now();
            if (now > deadline.getTime()) {
                const overdue = now - deadline.getTime();
                console.log(`       ⚠️ ПРОСРОЧЕНО на ${this._formatDuration(overdue)}`);
            } else {
                const remaining = deadline.getTime() - now;
                console.log(`       ⏳ Осталось: ${this._formatDuration(remaining)}`);
            }
        }
        
        // Контекст процесса
        console.log('\n🔗 Контекст процесса:');
        console.log(`   📍 Экземпляр: ${job.process_instance_key}`);
        console.log(`   🎯 Элемент BPMN: ${job.element_id}`);
        
        // Переменные
        const variablesCount = Object.keys(job.variables).length;
        if (variablesCount > 0) {
            console.log(`\n📦 Переменные (${variablesCount}):`);
            Object.entries(job.variables).forEach(([key, value]) => {
                console.log(`   ${key} = ${this._formatValue(value)}`);
            });
        }
        
        // Заголовки
        const headersCount = Object.keys(job.custom_headers).length;
        if (headersCount > 0) {
            console.log(`\n📋 Заголовки (${headersCount}):`);
            Object.entries(job.custom_headers).forEach(([key, value]) => {
                console.log(`   ${key}: ${value}`);
            });
        }
        
        // Настройки
        console.log('\n⚙️ Настройки:');
        if (job.timeout > 0) {
            const timeoutSec = job.timeout / 1000;
            console.log(`   ⏰ Таймаут: ${timeoutSec}с`);
        } else {
            console.log('   ⏰ Таймаут: не установлен');
        }
        
        // Ошибка
        if (job.error_message) {
            console.log('\n❌ Ошибка:');
            console.log(`   ${job.error_message}`);
        }
    }
    
    _formatState(state) {
        const stateEmojis = {
            'ACTIVATABLE': '🟡 ACTIVATABLE (готово к активации)',
            'ACTIVATED': '🔵 ACTIVATED (выполняется)',
            'COMPLETED': '🟢 COMPLETED (завершено)',
            'FAILED': '🔴 FAILED (провалено)',
            'CANCELLED': '⚫ CANCELLED (отменено)',
        };
        return stateEmojis[state] || state;
    }
    
    _formatValue(value) {
        if (value.length > 50) {
            return value.substring(0, 47) + '...';
        }
        return value;
    }
    
    _formatDuration(milliseconds) {
        const seconds = Math.floor(milliseconds / 1000);
        
        if (seconds < 60) {
            return `${seconds}с`;
        } else if (seconds < 3600) {
            const minutes = Math.floor(seconds / 60);
            const remainingSeconds = seconds % 60;
            return `${minutes}м ${remainingSeconds}с`;
        } else {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return `${hours}ч ${minutes}м`;
        }
    }
    
    async compareJobs(jobKey1, jobKey2) {
        const job1 = await this.getJob(jobKey1);
        const job2 = await this.getJob(jobKey2);
        
        console.log('📊 Сравнение заданий:');
        console.log('═'.repeat(40));
        console.log(`${'Параметр'.padEnd(20)} | ${jobKey1.padEnd(30)} | ${jobKey2.padEnd(30)}`);
        console.log('═'.repeat(40));
        
        // Сравниваем основные параметры
        this._compareField('Тип', job1.job_type, job2.job_type);
        this._compareField('Состояние', job1.state, job2.state);
        this._compareField('Воркер', job1.worker, job2.worker);
        this._compareField('Попытки', job1.retries.toString(), job2.retries.toString());
        this._compareField('Процесс', job1.process_instance_key, job2.process_instance_key);
        this._compareField('Элемент', job1.element_id, job2.element_id);
        
        // Сравниваем времена
        console.log('─'.repeat(40));
        this._compareTimeField('Создано', job1.created_at, job2.created_at);
        this._compareTimeField('Активировано', job1.activated_at, job2.activated_at);
        this._compareTimeField('Завершено', job1.completed_at, job2.completed_at);
    }
    
    _compareField(name, val1, val2) {
        const equal = val1 === val2 ? '✅' : '❌';
        console.log(`${name.padEnd(20)} | ${this._truncate(val1, 30).padEnd(30)} | ${this._truncate(val2, 30).padEnd(30)} ${equal}`);
    }
    
    _compareTimeField(name, time1, time2) {
        const format1 = this._formatTime(time1);
        const format2 = this._formatTime(time2);
        const equal = time1 === time2 ? '✅' : '❌';
        console.log(`${name.padEnd(20)} | ${format1.padEnd(30)} | ${format2.padEnd(30)} ${equal}`);
    }
    
    _formatTime(timeStr) {
        if (!timeStr) return '-';
        
        try {
            const date = new Date(timeStr);
            return date.toLocaleString();
        } catch {
            return timeStr;
        }
    }
    
    _truncate(s, maxLen) {
        if (s.length <= maxLen) return s;
        return s.substring(0, maxLen - 3) + '...';
    }
    
    async monitorJob(jobKey, interval = 5000) {
        console.log(`🔍 Мониторинг задания ${jobKey} (обновление каждые ${interval/1000}с)`);
        console.log('Нажмите Ctrl+C для остановки\n');
        
        let lastState = '';
        let lastRetries = -1;
        
        const monitor = setInterval(async () => {
            try {
                const job = await this.getJob(jobKey);
                const now = new Date().toLocaleTimeString();
                
                // Отслеживаем изменения состояния
                if (job.state !== lastState) {
                    console.log(`[${now}] 📊 Состояние: ${lastState} → ${job.state}`);
                    lastState = job.state;
                }
                
                // Отслеживаем изменения количества попыток
                if (job.retries !== lastRetries) {
                    if (lastRetries !== -1) {
                        console.log(`[${now}] 🔄 Попытки: ${lastRetries} → ${job.retries}`);
                    }
                    lastRetries = job.retries;
                }
                
                // Показываем текущий статус
                console.log(`[${now}] ${this._formatState(job.state)} | Попытки: ${job.retries} | Воркер: ${job.worker}`);
                
                // Если задание завершено, останавливаем мониторинг
                if (['COMPLETED', 'FAILED', 'CANCELLED'].includes(job.state)) {
                    console.log(`\n✅ Задание завершено со статусом: ${job.state}`);
                    clearInterval(monitor);
                }
                
            } catch (error) {
                console.log(`❌ Ошибка получения задания: ${error.message}`);
            }
        }, interval);
        
        // Обработка Ctrl+C
        process.on('SIGINT', () => {
            console.log('\n🛑 Мониторинг остановлен пользователем');
            clearInterval(monitor);
            process.exit(0);
        });
    }
    
    async exportJobToJSON(jobKey) {
        const job = await this.getJob(jobKey);
        
        const exportData = {
            job_key: job.job_key,
            job_type: job.job_type,
            state: job.state,
            worker: job.worker,
            retries: job.retries,
            created_at: job.created_at,
            activated_at: job.activated_at,
            completed_at: job.completed_at,
            deadline: job.deadline,
            process_instance_key: job.process_instance_key,
            element_id: job.element_id,
            variables: job.variables,
            custom_headers: job.custom_headers,
            error_message: job.error_message,
            timeout: job.timeout,
        };
        
        return JSON.stringify(exportData, null, 2);
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node get-job.js show <job_key>');
        console.log('  node get-job.js compare <job_key1> <job_key2>');
        console.log('  node get-job.js monitor <job_key> [interval_ms]');
        console.log('  node get-job.js export <job_key>');
        process.exit(1);
    }
    
    const manager = new JobDetailsManager();
    const command = args[0];
    
    (async () => {
        try {
            switch (command) {
                case 'show':
                    if (args.length < 2) {
                        console.log('❌ Укажите ключ задания');
                        process.exit(1);
                    }
                    await manager.printJobDetails(args[1]);
                    break;
                    
                case 'compare':
                    if (args.length < 3) {
                        console.log('❌ Укажите два ключа заданий для сравнения');
                        process.exit(1);
                    }
                    await manager.compareJobs(args[1], args[2]);
                    break;
                    
                case 'monitor':
                    if (args.length < 2) {
                        console.log('❌ Укажите ключ задания для мониторинга');
                        process.exit(1);
                    }
                    const interval = args.length > 2 ? parseInt(args[2]) : 5000;
                    await manager.monitorJob(args[1], interval);
                    break;
                    
                case 'export':
                    if (args.length < 2) {
                        console.log('❌ Укажите ключ задания для экспорта');
                        process.exit(1);
                    }
                    const jsonData = await manager.exportJobToJSON(args[1]);
                    console.log(jsonData);
                    break;
                    
                default:
                    console.log(`❌ Неизвестная команда: ${command}`);
                    process.exit(1);
            }
        } catch (error) {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        }
    })();
}

module.exports = {
    getJob,
    JobDetailsManager
};
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный job_key
- `NOT_FOUND` (5): Задание не найдено
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Job 'atom-jobkey12345' not found",
  "job": null
}
```

## Связанные методы
- [ListJobs](list-jobs.md) - Получение списка заданий
- [ActivateJobs](activate-jobs.md) - Активация заданий для выполнения
- [CompleteJob](complete-job.md) - Завершение задания
- [FailJob](fail-job.md) - Провал задания
- [CancelJob](cancel-job.md) - Отмена задания
