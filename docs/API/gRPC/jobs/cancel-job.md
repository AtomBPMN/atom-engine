# CancelJob

## Описание
Отменяет активное задание, делая его недоступным для активации воркерами. Полезно для остановки выполнения заданий при изменении бизнес-логики.

## Синтаксис
```protobuf
rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
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

### CancelJobRequest
```protobuf
message CancelJobRequest {
  string job_key = 1;    // Ключ задания
  string reason = 2;     // Причина отмены (опционально)
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания
- **reason** (string, optional): Причина отмены для аудита и логирования

## Параметры ответа

### CancelJobResponse
```protobuf
message CancelJobResponse {
  bool success = 1;         // Статус успешности операции
  string message = 2;       // Сообщение о результате
  string previous_state = 3; // Предыдущее состояние задания
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
    
    // Простая отмена задания
    response, err := client.CancelJob(ctx, &pb.CancelJobRequest{
        JobKey: jobKey,
        Reason: "Business logic changed",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Задание %s отменено (было: %s)\n", jobKey, response.PreviousState)
        fmt.Printf("   Причина: Business logic changed\n")
    } else {
        fmt.Printf("❌ Ошибка отмены: %s\n", response.Message)
    }
}

// Менеджер отмены заданий
type JobCancellationManager struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewJobCancellationManager(client pb.JobsServiceClient, ctx context.Context) *JobCancellationManager {
    return &JobCancellationManager{
        client: client,
        ctx:    ctx,
    }
}

func (jcm *JobCancellationManager) CancelJob(jobKey, reason string) error {
    response, err := jcm.client.CancelJob(jcm.ctx, &pb.CancelJobRequest{
        JobKey: jobKey,
        Reason: reason,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось отменить задание: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка отмены: %s", response.Message)
    }
    
    fmt.Printf("✅ Задание %s отменено (было: %s)\n", jobKey, response.PreviousState)
    if reason != "" {
        fmt.Printf("   Причина: %s\n", reason)
    }
    
    return nil
}

func (jcm *JobCancellationManager) CancelJobsByType(jobType, reason string) (int, error) {
    // Получаем список заданий по типу
    listResponse, err := jcm.client.ListJobs(jcm.ctx, &pb.ListJobsRequest{
        JobType: jobType,
        State:   "ACTIVATABLE", // Только активные задания
        Limit:   1000,
    })
    
    if err != nil {
        return 0, fmt.Errorf("не удалось получить список заданий: %v", err)
    }
    
    if !listResponse.Success {
        return 0, fmt.Errorf("ошибка получения списка: %s", listResponse.Message)
    }
    
    cancelledCount := 0
    
    for _, job := range listResponse.Jobs {
        err := jcm.CancelJob(job.JobKey, reason)
        if err != nil {
            fmt.Printf("⚠️ Не удалось отменить задание %s: %v\n", job.JobKey, err)
        } else {
            cancelledCount++
        }
    }
    
    fmt.Printf("📊 Отменено заданий типа '%s': %d из %d\n", 
               jobType, cancelledCount, len(listResponse.Jobs))
    
    return cancelledCount, nil
}

func (jcm *JobCancellationManager) CancelJobsByWorker(workerName, reason string) (int, error) {
    // Получаем список заданий, назначенных конкретному воркеру
    listResponse, err := jcm.client.ListJobs(jcm.ctx, &pb.ListJobsRequest{
        Worker: workerName,
        State:  "ACTIVATED", // Задания, назначенные воркеру
        Limit:  1000,
    })
    
    if err != nil {
        return 0, fmt.Errorf("не удалось получить список заданий воркера: %v", err)
    }
    
    if !listResponse.Success {
        return 0, fmt.Errorf("ошибка получения списка: %s", listResponse.Message)
    }
    
    cancelledCount := 0
    
    for _, job := range listResponse.Jobs {
        err := jcm.CancelJob(job.JobKey, reason)
        if err != nil {
            fmt.Printf("⚠️ Не удалось отменить задание %s воркера %s: %v\n", 
                       job.JobKey, workerName, err)
        } else {
            cancelledCount++
        }
    }
    
    fmt.Printf("📊 Отменено заданий воркера '%s': %d из %d\n", 
               workerName, cancelledCount, len(listResponse.Jobs))
    
    return cancelledCount, nil
}

func (jcm *JobCancellationManager) CancelOldJobs(olderThan time.Duration, reason string) (int, error) {
    // Получаем все активные задания
    listResponse, err := jcm.client.ListJobs(jcm.ctx, &pb.ListJobsRequest{
        State: "ACTIVATABLE",
        Limit: 1000,
    })
    
    if err != nil {
        return 0, fmt.Errorf("не удалось получить список заданий: %v", err)
    }
    
    if !listResponse.Success {
        return 0, fmt.Errorf("ошибка получения списка: %s", listResponse.Message)
    }
    
    cutoffTime := time.Now().Add(-olderThan)
    cancelledCount := 0
    
    for _, job := range listResponse.Jobs {
        // Парсим время создания задания
        jobCreatedAt, err := time.Parse(time.RFC3339, job.CreatedAt)
        if err != nil {
            fmt.Printf("⚠️ Не удалось парсить время создания задания %s: %v\n", job.JobKey, err)
            continue
        }
        
        if jobCreatedAt.Before(cutoffTime) {
            err := jcm.CancelJob(job.JobKey, fmt.Sprintf("%s (created %s ago)", 
                reason, time.Since(jobCreatedAt).String()))
            if err != nil {
                fmt.Printf("⚠️ Не удалось отменить старое задание %s: %v\n", job.JobKey, err)
            } else {
                cancelledCount++
            }
        }
    }
    
    fmt.Printf("📊 Отменено старых заданий (старше %s): %d\n", 
               olderThan.String(), cancelledCount)
    
    return cancelledCount, nil
}

func (jcm *JobCancellationManager) EmergencyStopAllJobs(reason string) (int, error) {
    // Экстренная остановка всех активных заданий
    fmt.Printf("🚨 ЭКСТРЕННАЯ ОСТАНОВКА: %s\n", reason)
    
    states := []string{"ACTIVATABLE", "ACTIVATED"}
    totalCancelled := 0
    
    for _, state := range states {
        listResponse, err := jcm.client.ListJobs(jcm.ctx, &pb.ListJobsRequest{
            State: state,
            Limit: 1000,
        })
        
        if err != nil {
            fmt.Printf("⚠️ Не удалось получить список заданий в состоянии %s: %v\n", state, err)
            continue
        }
        
        if !listResponse.Success {
            fmt.Printf("⚠️ Ошибка получения списка в состоянии %s: %s\n", state, listResponse.Message)
            continue
        }
        
        for _, job := range listResponse.Jobs {
            err := jcm.CancelJob(job.JobKey, fmt.Sprintf("EMERGENCY STOP: %s", reason))
            if err != nil {
                fmt.Printf("⚠️ Не удалось отменить задание %s: %v\n", job.JobKey, err)
            } else {
                totalCancelled++
            }
        }
    }
    
    fmt.Printf("🚨 Экстренная остановка завершена: отменено %d заданий\n", totalCancelled)
    return totalCancelled, nil
}

// Планировщик отмены заданий
type JobCancellationScheduler struct {
    manager *JobCancellationManager
    running bool
    stopCh  chan struct{}
}

func NewJobCancellationScheduler(manager *JobCancellationManager) *JobCancellationScheduler {
    return &JobCancellationScheduler{
        manager: manager,
        stopCh:  make(chan struct{}),
    }
}

func (jcs *JobCancellationScheduler) Start() {
    if jcs.running {
        return
    }
    
    jcs.running = true
    go jcs.run()
    fmt.Printf("📅 Планировщик отмены заданий запущен\n")
}

func (jcs *JobCancellationScheduler) Stop() {
    if !jcs.running {
        return
    }
    
    close(jcs.stopCh)
    jcs.running = false
    fmt.Printf("📅 Планировщик отмены заданий остановлен\n")
}

func (jcs *JobCancellationScheduler) run() {
    ticker := time.NewTicker(5 * time.Minute) // Проверяем каждые 5 минут
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            jcs.performScheduledCleanup()
        case <-jcs.stopCh:
            return
        }
    }
}

func (jcs *JobCancellationScheduler) performScheduledCleanup() {
    fmt.Printf("🧹 Выполняем запланированную очистку заданий...\n")
    
    // Отменяем задания старше 24 часов
    count, err := jcs.manager.CancelOldJobs(24*time.Hour, "Automatic cleanup - too old")
    if err != nil {
        fmt.Printf("⚠️ Ошибка очистки старых заданий: %v\n", err)
    } else if count > 0 {
        fmt.Printf("🧹 Очищено старых заданий: %d\n", count)
    }
    
    // Можно добавить другие правила очистки:
    // - Задания с большим количеством неудачных попыток
    // - Задания определенного типа в пиковые часы
    // - Задания с истекшим SLA
}

// Система мониторинга для отмены проблемных заданий
type JobHealthMonitor struct {
    manager *JobCancellationManager
    rules   []CancellationRule
}

type CancellationRule struct {
    Name      string
    Condition func(*pb.Job) bool
    Reason    string
}

func NewJobHealthMonitor(manager *JobCancellationManager) *JobHealthMonitor {
    monitor := &JobHealthMonitor{
        manager: manager,
    }
    
    // Добавляем стандартные правила
    monitor.AddRule(CancellationRule{
        Name: "TooManyRetries",
        Condition: func(job *pb.Job) bool {
            return job.Retries > 10
        },
        Reason: "Too many retries - possible infinite loop",
    })
    
    monitor.AddRule(CancellationRule{
        Name: "StuckJob",
        Condition: func(job *pb.Job) bool {
            if job.State != "ACTIVATED" {
                return false
            }
            
            activatedAt, err := time.Parse(time.RFC3339, job.ActivatedAt)
            if err != nil {
                return false
            }
            
            // Если задание активировано больше часа назад
            return time.Since(activatedAt) > time.Hour
        },
        Reason: "Job stuck in ACTIVATED state for too long",
    })
    
    return monitor
}

func (jhm *JobHealthMonitor) AddRule(rule CancellationRule) {
    jhm.rules = append(jhm.rules, rule)
}

func (jhm *JobHealthMonitor) MonitorAndCancel() error {
    // Получаем все активные задания
    listResponse, err := jhm.manager.client.ListJobs(jhm.manager.ctx, &pb.ListJobsRequest{
        Limit: 1000, // Получаем все задания
    })
    
    if err != nil {
        return fmt.Errorf("не удалось получить список заданий для мониторинга: %v", err)
    }
    
    if !listResponse.Success {
        return fmt.Errorf("ошибка получения списка заданий: %s", listResponse.Message)
    }
    
    cancelledByRule := make(map[string]int)
    
    for _, job := range listResponse.Jobs {
        for _, rule := range jhm.rules {
            if rule.Condition(job) {
                err := jhm.manager.CancelJob(job.JobKey, 
                    fmt.Sprintf("Health monitor: %s", rule.Reason))
                if err != nil {
                    fmt.Printf("⚠️ Не удалось отменить задание %s по правилу %s: %v\n", 
                               job.JobKey, rule.Name, err)
                } else {
                    cancelledByRule[rule.Name]++
                    fmt.Printf("🏥 Задание %s отменено по правилу: %s\n", job.JobKey, rule.Name)
                }
                break // Применяем только первое подходящее правило
            }
        }
    }
    
    // Выводим статистику
    for ruleName, count := range cancelledByRule {
        fmt.Printf("📊 Правило '%s': отменено %d заданий\n", ruleName, count)
    }
    
    return nil
}
```

### Python
```python
import grpc
import time
from datetime import datetime, timedelta
from typing import List, Callable, Dict
from dataclasses import dataclass

import jobs_pb2
import jobs_pb2_grpc

def cancel_job(job_key, reason=""):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.CancelJobRequest(
        job_key=job_key,
        reason=reason
    )
    
    try:
        response = stub.CancelJob(request, metadata=metadata)
        
        if response.success:
            print(f"✅ Задание {job_key} отменено (было: {response.previous_state})")
            if reason:
                print(f"   Причина: {reason}")
            return True
        else:
            print(f"❌ Ошибка отмены: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

class JobCancellationManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def cancel_job(self, job_key, reason=""):
        """Отменяет конкретное задание"""
        try:
            request = jobs_pb2.CancelJobRequest(
                job_key=job_key,
                reason=reason
            )
            
            response = self.stub.CancelJob(request, metadata=self.metadata)
            
            if response.success:
                print(f"✅ Задание {job_key} отменено (было: {response.previous_state})")
                if reason:
                    print(f"   Причина: {reason}")
                return True
            else:
                print(f"❌ Ошибка отмены: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при отмене задания: {e.details()}")
            return False
    
    def cancel_jobs_by_type(self, job_type, reason=""):
        """Отменяет все задания определенного типа"""
        jobs = self._get_jobs_list(job_type=job_type, state="ACTIVATABLE")
        
        cancelled_count = 0
        for job in jobs:
            if self.cancel_job(job['job_key'], reason):
                cancelled_count += 1
        
        print(f"📊 Отменено заданий типа '{job_type}': {cancelled_count} из {len(jobs)}")
        return cancelled_count
    
    def cancel_jobs_by_worker(self, worker_name, reason=""):
        """Отменяет все задания конкретного воркера"""
        jobs = self._get_jobs_list(worker=worker_name, state="ACTIVATED")
        
        cancelled_count = 0
        for job in jobs:
            if self.cancel_job(job['job_key'], reason):
                cancelled_count += 1
        
        print(f"📊 Отменено заданий воркера '{worker_name}': {cancelled_count} из {len(jobs)}")
        return cancelled_count
    
    def cancel_old_jobs(self, older_than_hours, reason=""):
        """Отменяет задания старше указанного времени"""
        jobs = self._get_jobs_list(state="ACTIVATABLE")
        
        cutoff_time = datetime.now() - timedelta(hours=older_than_hours)
        cancelled_count = 0
        
        for job in jobs:
            try:
                job_created_at = datetime.fromisoformat(job['created_at'].replace('Z', '+00:00'))
                if job_created_at < cutoff_time:
                    age = datetime.now() - job_created_at
                    full_reason = f"{reason} (created {age} ago)" if reason else f"Too old (created {age} ago)"
                    if self.cancel_job(job['job_key'], full_reason):
                        cancelled_count += 1
            except Exception as e:
                print(f"⚠️ Не удалось обработать задание {job['job_key']}: {e}")
        
        print(f"📊 Отменено старых заданий (старше {older_than_hours}ч): {cancelled_count}")
        return cancelled_count
    
    def emergency_stop_all_jobs(self, reason=""):
        """Экстренная остановка всех активных заданий"""
        print(f"🚨 ЭКСТРЕННАЯ ОСТАНОВКА: {reason}")
        
        states = ["ACTIVATABLE", "ACTIVATED"]
        total_cancelled = 0
        
        for state in states:
            jobs = self._get_jobs_list(state=state)
            
            for job in jobs:
                if self.cancel_job(job['job_key'], f"EMERGENCY STOP: {reason}"):
                    total_cancelled += 1
        
        print(f"🚨 Экстренная остановка завершена: отменено {total_cancelled} заданий")
        return total_cancelled
    
    def _get_jobs_list(self, job_type=None, worker=None, state=None):
        """Получает список заданий с фильтрацией"""
        try:
            request = jobs_pb2.ListJobsRequest(
                job_type=job_type or "",
                worker=worker or "",
                state=state or "",
                limit=1000
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return [
                    {
                        'job_key': job.job_key,
                        'job_type': job.job_type,
                        'state': job.state,
                        'worker': job.worker,
                        'created_at': job.created_at,
                        'activated_at': job.activated_at,
                        'retries': job.retries
                    }
                    for job in response.jobs
                ]
            else:
                print(f"❌ Ошибка получения списка заданий: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error при получении списка заданий: {e.details()}")
            return []

import threading
import schedule

class JobCancellationScheduler:
    def __init__(self, manager: JobCancellationManager):
        self.manager = manager
        self.running = False
        self.thread = None
    
    def start(self):
        """Запускает планировщик"""
        if self.running:
            return
        
        self.running = True
        
        # Настройка расписания
        schedule.every(5).minutes.do(self._perform_scheduled_cleanup)
        schedule.every().day.at("02:00").do(self._daily_cleanup)
        schedule.every().hour.do(self._health_check)
        
        self.thread = threading.Thread(target=self._run_scheduler)
        self.thread.daemon = True
        self.thread.start()
        
        print("📅 Планировщик отмены заданий запущен")
    
    def stop(self):
        """Останавливает планировщик"""
        self.running = False
        schedule.clear()
        
        if self.thread:
            self.thread.join(timeout=5)
        
        print("📅 Планировщик отмены заданий остановлен")
    
    def _run_scheduler(self):
        """Основной цикл планировщика"""
        while self.running:
            schedule.run_pending()
            time.sleep(30)  # Проверяем каждые 30 секунд
    
    def _perform_scheduled_cleanup(self):
        """Выполняет запланированную очистку"""
        print("🧹 Выполняем запланированную очистку заданий...")
        
        # Отменяем задания старше 24 часов
        count = self.manager.cancel_old_jobs(24, "Automatic cleanup - too old")
        if count > 0:
            print(f"🧹 Очищено старых заданий: {count}")
    
    def _daily_cleanup(self):
        """Ежедневная глубокая очистка"""
        print("🌙 Выполняем ежедневную очистку...")
        
        # Отменяем очень старые задания
        self.manager.cancel_old_jobs(72, "Daily cleanup - expired")
    
    def _health_check(self):
        """Проверка здоровья заданий"""
        monitor = JobHealthMonitor(self.manager)
        monitor.monitor_and_cancel()

@dataclass
class CancellationRule:
    name: str
    condition: Callable
    reason: str

class JobHealthMonitor:
    def __init__(self, manager: JobCancellationManager):
        self.manager = manager
        self.rules = []
        
        # Добавляем стандартные правила
        self.add_rule(CancellationRule(
            name="TooManyRetries",
            condition=lambda job: job['retries'] > 10,
            reason="Too many retries - possible infinite loop"
        ))
        
        self.add_rule(CancellationRule(
            name="StuckJob",
            condition=self._is_stuck_job,
            reason="Job stuck in ACTIVATED state for too long"
        ))
    
    def add_rule(self, rule: CancellationRule):
        """Добавляет правило для отмены заданий"""
        self.rules.append(rule)
    
    def monitor_and_cancel(self):
        """Мониторинг и отмена проблемных заданий"""
        jobs = self.manager._get_jobs_list()
        
        cancelled_by_rule = {}
        
        for job in jobs:
            for rule in self.rules:
                try:
                    if rule.condition(job):
                        if self.manager.cancel_job(job['job_key'], f"Health monitor: {rule.reason}"):
                            cancelled_by_rule[rule.name] = cancelled_by_rule.get(rule.name, 0) + 1
                            print(f"🏥 Задание {job['job_key']} отменено по правилу: {rule.name}")
                        break  # Применяем только первое подходящее правило
                except Exception as e:
                    print(f"⚠️ Ошибка применения правила {rule.name} к заданию {job['job_key']}: {e}")
        
        # Выводим статистику
        for rule_name, count in cancelled_by_rule.items():
            print(f"📊 Правило '{rule_name}': отменено {count} заданий")
    
    def _is_stuck_job(self, job):
        """Проверяет, заблокировано ли задание"""
        if job['state'] != "ACTIVATED":
            return False
        
        try:
            activated_at = datetime.fromisoformat(job['activated_at'].replace('Z', '+00:00'))
            return datetime.now() - activated_at > timedelta(hours=1)
        except:
            return False

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python cancel_job.py <job_key> [reason]")
        print("  python cancel_job.py test")
        print("  python cancel_job.py emergency <reason>")
        sys.exit(1)
    
    if sys.argv[1] == "test":
        # Тестирование различных сценариев отмены
        manager = JobCancellationManager()
        
        print("--- Тест отмены по типу ---")
        manager.cancel_jobs_by_type("test-job-type", "Testing cancellation")
        
        print("\n--- Тест отмены старых заданий ---")
        manager.cancel_old_jobs(1, "Test cleanup")
        
        print("\n--- Тест мониторинга здоровья ---")
        monitor = JobHealthMonitor(manager)
        monitor.monitor_and_cancel()
        
    elif sys.argv[1] == "emergency":
        reason = sys.argv[2] if len(sys.argv) > 2 else "Emergency stop requested"
        manager = JobCancellationManager()
        manager.emergency_stop_all_jobs(reason)
        
    else:
        job_key = sys.argv[1]
        reason = sys.argv[2] if len(sys.argv) > 2 else ""
        
        cancel_job(job_key, reason)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function cancelJob(jobKey, reason = "") {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_key: jobKey,
            reason: reason
        };
        
        client.cancelJob(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`✅ Задание ${jobKey} отменено (было: ${response.previous_state})`);
                if (reason) {
                    console.log(`   Причина: ${reason}`);
                }
                resolve(true);
            } else {
                console.log(`❌ Ошибка отмены: ${response.message}`);
                resolve(false);
            }
        });
    });
}

class JobCancellationManager {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async cancelJob(jobKey, reason = "") {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                reason: reason
            };
            
            this.client.cancelJob(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при отмене задания: ${error.message}`);
                    resolve(false);
                    return;
                }
                
                if (response.success) {
                    console.log(`✅ Задание ${jobKey} отменено (было: ${response.previous_state})`);
                    if (reason) {
                        console.log(`   Причина: ${reason}`);
                    }
                    resolve(true);
                } else {
                    console.log(`❌ Ошибка отмены: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
    
    async cancelJobsByType(jobType, reason = "") {
        const jobs = await this._getJobsList({ jobType: jobType, state: "ACTIVATABLE" });
        
        let cancelledCount = 0;
        for (const job of jobs) {
            if (await this.cancelJob(job.job_key, reason)) {
                cancelledCount++;
            }
        }
        
        console.log(`📊 Отменено заданий типа '${jobType}': ${cancelledCount} из ${jobs.length}`);
        return cancelledCount;
    }
    
    async cancelJobsByWorker(workerName, reason = "") {
        const jobs = await this._getJobsList({ worker: workerName, state: "ACTIVATED" });
        
        let cancelledCount = 0;
        for (const job of jobs) {
            if (await this.cancelJob(job.job_key, reason)) {
                cancelledCount++;
            }
        }
        
        console.log(`📊 Отменено заданий воркера '${workerName}': ${cancelledCount} из ${jobs.length}`);
        return cancelledCount;
    }
    
    async cancelOldJobs(olderThanHours, reason = "") {
        const jobs = await this._getJobsList({ state: "ACTIVATABLE" });
        
        const cutoffTime = new Date(Date.now() - olderThanHours * 60 * 60 * 1000);
        let cancelledCount = 0;
        
        for (const job of jobs) {
            try {
                const jobCreatedAt = new Date(job.created_at);
                if (jobCreatedAt < cutoffTime) {
                    const age = Math.floor((Date.now() - jobCreatedAt.getTime()) / (1000 * 60 * 60));
                    const fullReason = reason ? `${reason} (created ${age}h ago)` : `Too old (created ${age}h ago)`;
                    if (await this.cancelJob(job.job_key, fullReason)) {
                        cancelledCount++;
                    }
                }
            } catch (error) {
                console.log(`⚠️ Не удалось обработать задание ${job.job_key}: ${error.message}`);
            }
        }
        
        console.log(`📊 Отменено старых заданий (старше ${olderThanHours}ч): ${cancelledCount}`);
        return cancelledCount;
    }
    
    async emergencyStopAllJobs(reason = "") {
        console.log(`🚨 ЭКСТРЕННАЯ ОСТАНОВКА: ${reason}`);
        
        const states = ["ACTIVATABLE", "ACTIVATED"];
        let totalCancelled = 0;
        
        for (const state of states) {
            const jobs = await this._getJobsList({ state: state });
            
            for (const job of jobs) {
                if (await this.cancelJob(job.job_key, `EMERGENCY STOP: ${reason}`)) {
                    totalCancelled++;
                }
            }
        }
        
        console.log(`🚨 Экстренная остановка завершена: отменено ${totalCancelled} заданий`);
        return totalCancelled;
    }
    
    async _getJobsList(filters = {}) {
        return new Promise((resolve, reject) => {
            const request = {
                job_type: filters.jobType || "",
                worker: filters.worker || "",
                state: filters.state || "",
                limit: 1000
            };
            
            this.client.listJobs(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при получении списка заданий: ${error.message}`);
                    resolve([]);
                    return;
                }
                
                if (response.success) {
                    resolve(response.jobs.map(job => ({
                        job_key: job.job_key,
                        job_type: job.job_type,
                        state: job.state,
                        worker: job.worker,
                        created_at: job.created_at,
                        activated_at: job.activated_at,
                        retries: job.retries
                    })));
                } else {
                    console.log(`❌ Ошибка получения списка заданий: ${response.message}`);
                    resolve([]);
                }
            });
        });
    }
}

class JobCancellationScheduler {
    constructor(manager) {
        this.manager = manager;
        this.running = false;
        this.intervals = [];
    }
    
    start() {
        if (this.running) return;
        
        this.running = true;
        
        // Настройка расписания
        this.intervals.push(setInterval(() => this._performScheduledCleanup(), 5 * 60 * 1000)); // Каждые 5 минут
        this.intervals.push(setInterval(() => this._dailyCleanup(), 24 * 60 * 60 * 1000)); // Каждый день
        this.intervals.push(setInterval(() => this._healthCheck(), 60 * 60 * 1000)); // Каждый час
        
        console.log("📅 Планировщик отмены заданий запущен");
    }
    
    stop() {
        this.running = false;
        
        this.intervals.forEach(interval => clearInterval(interval));
        this.intervals = [];
        
        console.log("📅 Планировщик отмены заданий остановлен");
    }
    
    async _performScheduledCleanup() {
        console.log("🧹 Выполняем запланированную очистку заданий...");
        
        const count = await this.manager.cancelOldJobs(24, "Automatic cleanup - too old");
        if (count > 0) {
            console.log(`🧹 Очищено старых заданий: ${count}`);
        }
    }
    
    async _dailyCleanup() {
        console.log("🌙 Выполняем ежедневную очистку...");
        await this.manager.cancelOldJobs(72, "Daily cleanup - expired");
    }
    
    async _healthCheck() {
        const monitor = new JobHealthMonitor(this.manager);
        await monitor.monitorAndCancel();
    }
}

class JobHealthMonitor {
    constructor(manager) {
        this.manager = manager;
        this.rules = [];
        
        // Добавляем стандартные правила
        this.addRule({
            name: "TooManyRetries",
            condition: (job) => job.retries > 10,
            reason: "Too many retries - possible infinite loop"
        });
        
        this.addRule({
            name: "StuckJob",
            condition: (job) => this._isStuckJob(job),
            reason: "Job stuck in ACTIVATED state for too long"
        });
    }
    
    addRule(rule) {
        this.rules.push(rule);
    }
    
    async monitorAndCancel() {
        const jobs = await this.manager._getJobsList();
        
        const cancelledByRule = {};
        
        for (const job of jobs) {
            for (const rule of this.rules) {
                try {
                    if (rule.condition(job)) {
                        if (await this.manager.cancelJob(job.job_key, `Health monitor: ${rule.reason}`)) {
                            cancelledByRule[rule.name] = (cancelledByRule[rule.name] || 0) + 1;
                            console.log(`🏥 Задание ${job.job_key} отменено по правилу: ${rule.name}`);
                        }
                        break; // Применяем только первое подходящее правило
                    }
                } catch (error) {
                    console.log(`⚠️ Ошибка применения правила ${rule.name} к заданию ${job.job_key}: ${error.message}`);
                }
            }
        }
        
        // Выводим статистику
        Object.entries(cancelledByRule).forEach(([ruleName, count]) => {
            console.log(`📊 Правило '${ruleName}': отменено ${count} заданий`);
        });
    }
    
    _isStuckJob(job) {
        if (job.state !== "ACTIVATED") return false;
        
        try {
            const activatedAt = new Date(job.activated_at);
            return Date.now() - activatedAt.getTime() > 60 * 60 * 1000; // 1 час
        } catch {
            return false;
        }
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node cancel-job.js <job_key> [reason]');
        console.log('  node cancel-job.js test');
        console.log('  node cancel-job.js emergency <reason>');
        process.exit(1);
    }
    
    if (args[0] === 'test') {
        // Тестирование различных сценариев отмены
        (async () => {
            const manager = new JobCancellationManager();
            
            console.log("--- Тест отмены по типу ---");
            await manager.cancelJobsByType("test-job-type", "Testing cancellation");
            
            console.log("\n--- Тест отмены старых заданий ---");
            await manager.cancelOldJobs(1, "Test cleanup");
            
            console.log("\n--- Тест мониторинга здоровья ---");
            const monitor = new JobHealthMonitor(manager);
            await monitor.monitorAndCancel();
        })();
    } else if (args[0] === 'emergency') {
        const reason = args[1] || "Emergency stop requested";
        const manager = new JobCancellationManager();
        manager.emergencyStopAllJobs(reason);
    } else {
        const jobKey = args[0];
        const reason = args[1] || "";
        
        cancelJob(jobKey, reason).catch(error => {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        });
    }
}

module.exports = {
    cancelJob,
    JobCancellationManager,
    JobCancellationScheduler,
    JobHealthMonitor
};
```

## Стратегии отмены заданий

### Базовые операции
- **Прямая отмена**: Отмена конкретного задания по ключу
- **Групповая отмена**: Отмена по типу, воркеру или состоянию
- **Временная отмена**: Отмена заданий старше определенного времени
- **Экстренная остановка**: Отмена всех активных заданий

### Правила автоматической отмены
- **TooManyRetries**: Слишком много попыток (>10)
- **StuckJob**: Задание зависло в состоянии ACTIVATED (>1 часа)
- **OldJob**: Задание создано слишком давно (>24 часов)
- **ErrorPattern**: Задания с определенными паттернами ошибок

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
  "message": "Job 'atom-jobkey12345' not found or already completed",
  "previous_state": ""
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Получение заданий для выполнения
- [ListJobs](list-jobs.md) - Список заданий для массовой отмены
- [GetJob](get-job.md) - Получение информации о задании
- [GetJobStats](get-job-stats.md) - Статистика заданий для анализа
