# GetJobStats

## Описание
Получает агрегированную статистику по всем заданиям в системе. Предоставляет информацию о количестве заданий по состояниям, типам, воркерам и другим метрикам.

## Синтаксис
```protobuf
rpc GetJobStats(GetJobStatsRequest) returns (GetJobStatsResponse);
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

### GetJobStatsRequest
```protobuf
message GetJobStatsRequest {
  string job_type = 1;      // Фильтр по типу задания (опционально)
  string worker = 2;        // Фильтр по воркеру (опционально)
  string time_range = 3;    // Временной диапазон ("1h", "24h", "7d", "30d")
}
```

#### Поля:
- **job_type** (string, optional): Фильтр статистики по конкретному типу задания
- **worker** (string, optional): Фильтр статистики по конкретному воркеру
- **time_range** (string, optional): Временной диапазон для анализа ("1h", "24h", "7d", "30d")

## Параметры ответа

### GetJobStatsResponse
```protobuf
message GetJobStatsResponse {
  bool success = 1;                     // Статус успешности операции
  string message = 2;                   // Сообщение о результате
  JobStats stats = 3;                   // Статистика заданий
}

message JobStats {
  int32 total_jobs = 1;                 // Общее количество заданий
  map<string, int32> by_state = 2;      // Статистика по состояниям
  map<string, int32> by_type = 3;       // Статистика по типам
  map<string, int32> by_worker = 4;     // Статистика по воркерам
  PerformanceStats performance = 5;     // Метрики производительности
  string generated_at = 6;              // Время генерации статистики
}

message PerformanceStats {
  double avg_execution_time = 1;        // Среднее время выполнения (мс)
  double avg_wait_time = 2;             // Среднее время ожидания (мс)
  int32 successful_jobs = 3;            // Количество успешных заданий
  int32 failed_jobs = 4;                // Количество провалившихся заданий
  double success_rate = 5;              // Процент успешности (0-100)
  int32 retry_count = 6;                // Общее количество повторов
  double avg_retries_per_job = 7;       // Среднее количество повторов на задание
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
    
    // Получаем общую статистику
    response, err := client.GetJobStats(ctx, &pb.GetJobStatsRequest{
        TimeRange: "24h",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        stats := response.Stats
        fmt.Printf("📊 Статистика заданий за последние 24 часа:\n\n")
        
        // Общая информация
        fmt.Printf("Общее количество заданий: %d\n", stats.TotalJobs)
        fmt.Printf("Время генерации: %s\n\n", stats.GeneratedAt)
        
        // Статистика по состояниям
        fmt.Printf("По состояниям:\n")
        for state, count := range stats.ByState {
            percentage := float64(count) / float64(stats.TotalJobs) * 100
            fmt.Printf("  %s: %d (%.1f%%)\n", state, count, percentage)
        }
        
        // Статистика по типам
        if len(stats.ByType) > 0 {
            fmt.Printf("\nПо типам:\n")
            for jobType, count := range stats.ByType {
                percentage := float64(count) / float64(stats.TotalJobs) * 100
                fmt.Printf("  %s: %d (%.1f%%)\n", jobType, count, percentage)
            }
        }
        
        // Статистика по воркерам
        if len(stats.ByWorker) > 0 {
            fmt.Printf("\nПо воркерам:\n")
            for worker, count := range stats.ByWorker {
                fmt.Printf("  %s: %d заданий\n", worker, count)
            }
        }
        
        // Метрики производительности
        if stats.Performance != nil {
            perf := stats.Performance
            fmt.Printf("\nПроизводительность:\n")
            fmt.Printf("  Среднее время выполнения: %.1fмс\n", perf.AvgExecutionTime)
            fmt.Printf("  Среднее время ожидания: %.1fмс\n", perf.AvgWaitTime)
            fmt.Printf("  Успешных заданий: %d\n", perf.SuccessfulJobs)
            fmt.Printf("  Провалившихся заданий: %d\n", perf.FailedJobs)
            fmt.Printf("  Процент успешности: %.1f%%\n", perf.SuccessRate)
            fmt.Printf("  Всего повторов: %d\n", perf.RetryCount)
            fmt.Printf("  Среднее повторов на задание: %.2f\n", perf.AvgRetriesPerJob)
        }
    } else {
        fmt.Printf("❌ Ошибка получения статистики: %s\n", response.Message)
    }
}

// Менеджер статистики заданий
type JobStatsManager struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewJobStatsManager(client pb.JobsServiceClient, ctx context.Context) *JobStatsManager {
    return &JobStatsManager{
        client: client,
        ctx:    ctx,
    }
}

func (jsm *JobStatsManager) GetOverallStats(timeRange string) (*pb.JobStats, error) {
    response, err := jsm.client.GetJobStats(jsm.ctx, &pb.GetJobStatsRequest{
        TimeRange: timeRange,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка получения статистики: %s", response.Message)
    }
    
    return response.Stats, nil
}

func (jsm *JobStatsManager) GetStatsForJobType(jobType, timeRange string) (*pb.JobStats, error) {
    response, err := jsm.client.GetJobStats(jsm.ctx, &pb.GetJobStatsRequest{
        JobType:   jobType,
        TimeRange: timeRange,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка получения статистики: %s", response.Message)
    }
    
    return response.Stats, nil
}

func (jsm *JobStatsManager) GetStatsForWorker(worker, timeRange string) (*pb.JobStats, error) {
    response, err := jsm.client.GetJobStats(jsm.ctx, &pb.GetJobStatsRequest{
        Worker:    worker,
        TimeRange: timeRange,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка получения статистики: %s", response.Message)
    }
    
    return response.Stats, nil
}

func (jsm *JobStatsManager) PrintDetailedReport(timeRange string) error {
    stats, err := jsm.GetOverallStats(timeRange)
    if err != nil {
        return err
    }
    
    fmt.Printf("📊 Детальный отчет по заданиям (%s)\n", timeRange)
    fmt.Printf("═══════════════════════════════════════════\n")
    fmt.Printf("Время генерации: %s\n", stats.GeneratedAt)
    fmt.Printf("Общее количество заданий: %d\n\n", stats.TotalJobs)
    
    // Анализ по состояниям
    fmt.Printf("🔍 Анализ по состояниям:\n")
    activatable := stats.ByState["ACTIVATABLE"]
    activated := stats.ByState["ACTIVATED"]
    completed := stats.ByState["COMPLETED"]
    failed := stats.ByState["FAILED"]
    cancelled := stats.ByState["CANCELLED"]
    
    fmt.Printf("  🟡 Готовые к активации: %d\n", activatable)
    fmt.Printf("  🔵 В процессе выполнения: %d\n", activated)
    fmt.Printf("  🟢 Завершенные: %d\n", completed)
    fmt.Printf("  🔴 Провалившиеся: %d\n", failed)
    fmt.Printf("  ⚫ Отмененные: %d\n", cancelled)
    
    // Анализ нагрузки
    if activatable > 0 || activated > 0 {
        fmt.Printf("\n⚡ Анализ нагрузки:\n")
        
        if activatable > 50 {
            fmt.Printf("  ⚠️ Много заданий ожидают активации (%d)\n", activatable)
        }
        
        if activated > 100 {
            fmt.Printf("  ⚠️ Высокая нагрузка выполнения (%d)\n", activated)
        }
        
        pendingTotal := activatable + activated
        if pendingTotal > 0 {
            fmt.Printf("  📋 Всего в обработке: %d заданий\n", pendingTotal)
        }
    }
    
    // Анализ производительности
    if stats.Performance != nil {
        perf := stats.Performance
        fmt.Printf("\n📈 Анализ производительности:\n")
        
        // Время выполнения
        avgExecSec := perf.AvgExecutionTime / 1000
        fmt.Printf("  ⏱️ Среднее время выполнения: %.1fс\n", avgExecSec)
        
        if avgExecSec > 60 {
            fmt.Printf("     ⚠️ Медленное выполнение (>1 минуты)\n")
        } else if avgExecSec < 1 {
            fmt.Printf("     ✅ Быстрое выполнение (<1 секунды)\n")
        }
        
        // Время ожидания
        avgWaitSec := perf.AvgWaitTime / 1000
        fmt.Printf("  ⏳ Среднее время ожидания: %.1fс\n", avgWaitSec)
        
        if avgWaitSec > 300 {
            fmt.Printf("     ⚠️ Долгое ожидание (>5 минут)\n")
        }
        
        // Успешность
        fmt.Printf("  📊 Процент успешности: %.1f%%\n", perf.SuccessRate)
        
        if perf.SuccessRate < 80 {
            fmt.Printf("     ❌ Низкая успешность (<80%%)\n")
        } else if perf.SuccessRate > 95 {
            fmt.Printf("     ✅ Отличная успешность (>95%%)\n")
        }
        
        // Повторы
        fmt.Printf("  🔄 Среднее повторов: %.2f\n", perf.AvgRetriesPerJob)
        
        if perf.AvgRetriesPerJob > 2 {
            fmt.Printf("     ⚠️ Много повторов (>2 в среднем)\n")
        }
    }
    
    // Топ типов заданий
    if len(stats.ByType) > 0 {
        fmt.Printf("\n🏷️ Топ типов заданий:\n")
        
        // Сортируем по количеству
        type TypeCount struct {
            Type  string
            Count int32
        }
        
        var types []TypeCount
        for jobType, count := range stats.ByType {
            types = append(types, TypeCount{Type: jobType, Count: count})
        }
        
        // Простая сортировка по убыванию
        for i := 0; i < len(types); i++ {
            for j := i + 1; j < len(types); j++ {
                if types[j].Count > types[i].Count {
                    types[i], types[j] = types[j], types[i]
                }
            }
        }
        
        // Показываем топ-5
        for i, tc := range types {
            if i >= 5 {
                break
            }
            percentage := float64(tc.Count) / float64(stats.TotalJobs) * 100
            fmt.Printf("  %d. %s: %d (%.1f%%)\n", i+1, tc.Type, tc.Count, percentage)
        }
    }
    
    // Активные воркеры
    if len(stats.ByWorker) > 0 {
        fmt.Printf("\n👥 Активные воркеры:\n")
        for worker, count := range stats.ByWorker {
            fmt.Printf("  %s: %d заданий\n", worker, count)
        }
    }
    
    return nil
}

func (jsm *JobStatsManager) CompareTimeRanges(range1, range2 string) error {
    stats1, err := jsm.GetOverallStats(range1)
    if err != nil {
        return fmt.Errorf("ошибка получения статистики для %s: %v", range1, err)
    }
    
    stats2, err := jsm.GetOverallStats(range2)
    if err != nil {
        return fmt.Errorf("ошибка получения статистики для %s: %v", range2, err)
    }
    
    fmt.Printf("📊 Сравнение периодов: %s vs %s\n", range1, range2)
    fmt.Printf("═══════════════════════════════════════\n")
    
    // Сравнение общих метрик
    fmt.Printf("%-20s | %-15s | %-15s | Изменение\n", "Метрика", range1, range2)
    fmt.Printf("─────────────────────────────────────────────────────────────\n")
    
    jsm.compareMetric("Всего заданий", stats1.TotalJobs, stats2.TotalJobs)
    jsm.compareMetric("Завершенные", stats1.ByState["COMPLETED"], stats2.ByState["COMPLETED"])
    jsm.compareMetric("Провалившиеся", stats1.ByState["FAILED"], stats2.ByState["FAILED"])
    
    if stats1.Performance != nil && stats2.Performance != nil {
        perf1 := stats1.Performance
        perf2 := stats2.Performance
        
        fmt.Printf("─────────────────────────────────────────────────────────────\n")
        jsm.compareFloatMetric("Успешность (%)", perf1.SuccessRate, perf2.SuccessRate)
        jsm.compareFloatMetric("Ср. выполнение (с)", perf1.AvgExecutionTime/1000, perf2.AvgExecutionTime/1000)
        jsm.compareFloatMetric("Ср. ожидание (с)", perf1.AvgWaitTime/1000, perf2.AvgWaitTime/1000)
        jsm.compareFloatMetric("Ср. повторов", perf1.AvgRetriesPerJob, perf2.AvgRetriesPerJob)
    }
    
    return nil
}

func (jsm *JobStatsManager) compareMetric(name string, val1, val2 int32) {
    diff := val2 - val1
    var change string
    
    if diff > 0 {
        change = fmt.Sprintf("↗️ +%d", diff)
    } else if diff < 0 {
        change = fmt.Sprintf("↘️ %d", diff)
    } else {
        change = "➡️ 0"
    }
    
    fmt.Printf("%-20s | %-15d | %-15d | %s\n", name, val1, val2, change)
}

func (jsm *JobStatsManager) compareFloatMetric(name string, val1, val2 float64) {
    diff := val2 - val1
    var change string
    
    if diff > 0.1 {
        change = fmt.Sprintf("↗️ +%.2f", diff)
    } else if diff < -0.1 {
        change = fmt.Sprintf("↘️ %.2f", diff)
    } else {
        change = "➡️ ~0"
    }
    
    fmt.Printf("%-20s | %-15.2f | %-15.2f | %s\n", name, val1, val2, change)
}

// Мониторинг статистики в реальном времени
func (jsm *JobStatsManager) MonitorStats(interval time.Duration) {
    fmt.Printf("📊 Мониторинг статистики заданий (обновление каждые %s)\n", interval)
    fmt.Printf("Нажмите Ctrl+C для остановки\n\n")
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    var prevStats *pb.JobStats
    
    for range ticker.C {
        stats, err := jsm.GetOverallStats("1h")
        if err != nil {
            fmt.Printf("❌ Ошибка получения статистики: %v\n", err)
            continue
        }
        
        now := time.Now().Format("15:04:05")
        
        // Отображаем текущие метрики
        activatable := stats.ByState["ACTIVATABLE"]
        activated := stats.ByState["ACTIVATED"]
        completed := stats.ByState["COMPLETED"]
        failed := stats.ByState["FAILED"]
        
        fmt.Printf("[%s] Всего: %d | Ожидает: %d | Выполняется: %d | Завершено: %d | Провалено: %d\n",
                   now, stats.TotalJobs, activatable, activated, completed, failed)
        
        // Показываем изменения
        if prevStats != nil {
            changes := jsm.detectChanges(prevStats, stats)
            if len(changes) > 0 {
                fmt.Printf("         Изменения: %s\n", changes)
            }
        }
        
        prevStats = stats
    }
}

func (jsm *JobStatsManager) detectChanges(prev, curr *pb.JobStats) string {
    var changes []string
    
    // Проверяем изменения в состояниях
    for state, currCount := range curr.ByState {
        prevCount := prev.ByState[state]
        if currCount != prevCount {
            diff := currCount - prevCount
            if diff > 0 {
                changes = append(changes, fmt.Sprintf("%s +%d", state, diff))
            } else {
                changes = append(changes, fmt.Sprintf("%s %d", state, diff))
            }
        }
    }
    
    if len(changes) == 0 {
        return ""
    }
    
    if len(changes) > 3 {
        return fmt.Sprintf("%s и еще %d", changes[0], len(changes)-1)
    }
    
    result := ""
    for i, change := range changes {
        if i > 0 {
            result += ", "
        }
        result += change
    }
    
    return result
}

// Экспорт статистики в различные форматы
func (jsm *JobStatsManager) ExportStatsToCSV(timeRange string) (string, error) {
    stats, err := jsm.GetOverallStats(timeRange)
    if err != nil {
        return "", err
    }
    
    csv := "Type,Metric,Value\n"
    
    // Общие метрики
    csv += fmt.Sprintf("General,Total Jobs,%d\n", stats.TotalJobs)
    csv += fmt.Sprintf("General,Generated At,%s\n", stats.GeneratedAt)
    
    // По состояниям
    for state, count := range stats.ByState {
        csv += fmt.Sprintf("State,%s,%d\n", state, count)
    }
    
    // По типам
    for jobType, count := range stats.ByType {
        csv += fmt.Sprintf("JobType,%s,%d\n", jobType, count)
    }
    
    // По воркерам
    for worker, count := range stats.ByWorker {
        csv += fmt.Sprintf("Worker,%s,%d\n", worker, count)
    }
    
    // Производительность
    if stats.Performance != nil {
        perf := stats.Performance
        csv += fmt.Sprintf("Performance,Avg Execution Time,%.2f\n", perf.AvgExecutionTime)
        csv += fmt.Sprintf("Performance,Avg Wait Time,%.2f\n", perf.AvgWaitTime)
        csv += fmt.Sprintf("Performance,Success Rate,%.2f\n", perf.SuccessRate)
        csv += fmt.Sprintf("Performance,Successful Jobs,%d\n", perf.SuccessfulJobs)
        csv += fmt.Sprintf("Performance,Failed Jobs,%d\n", perf.FailedJobs)
        csv += fmt.Sprintf("Performance,Retry Count,%d\n", perf.RetryCount)
        csv += fmt.Sprintf("Performance,Avg Retries Per Job,%.2f\n", perf.AvgRetriesPerJob)
    }
    
    return csv, nil
}
```

### Python
```python
import grpc
import time
import csv
import io
from datetime import datetime
from typing import Dict, List, Optional

import jobs_pb2
import jobs_pb2_grpc

def get_job_stats(job_type="", worker="", time_range="24h"):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.GetJobStatsRequest(
        job_type=job_type,
        worker=worker,
        time_range=time_range
    )
    
    try:
        response = stub.GetJobStats(request, metadata=metadata)
        
        if response.success:
            stats = response.stats
            print(f"📊 Статистика заданий за {time_range}:\n")
            
            # Общая информация
            print(f"Общее количество заданий: {stats.total_jobs}")
            print(f"Время генерации: {stats.generated_at}\n")
            
            # Статистика по состояниям
            print("По состояниям:")
            for state, count in stats.by_state.items():
                percentage = (count / stats.total_jobs * 100) if stats.total_jobs > 0 else 0
                print(f"  {state}: {count} ({percentage:.1f}%)")
            
            # Статистика по типам
            if stats.by_type:
                print("\nПо типам:")
                for job_type_name, count in stats.by_type.items():
                    percentage = (count / stats.total_jobs * 100) if stats.total_jobs > 0 else 0
                    print(f"  {job_type_name}: {count} ({percentage:.1f}%)")
            
            # Статистика по воркерам
            if stats.by_worker:
                print("\nПо воркерам:")
                for worker_name, count in stats.by_worker.items():
                    print(f"  {worker_name}: {count} заданий")
            
            # Метрики производительности
            if stats.performance:
                perf = stats.performance
                print("\nПроизводительность:")
                print(f"  Среднее время выполнения: {perf.avg_execution_time:.1f}мс")
                print(f"  Среднее время ожидания: {perf.avg_wait_time:.1f}мс")
                print(f"  Успешных заданий: {perf.successful_jobs}")
                print(f"  Провалившихся заданий: {perf.failed_jobs}")
                print(f"  Процент успешности: {perf.success_rate:.1f}%")
                print(f"  Всего повторов: {perf.retry_count}")
                print(f"  Среднее повторов на задание: {perf.avg_retries_per_job:.2f}")
            
            return stats
        else:
            print(f"❌ Ошибка получения статистики: {response.message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

class JobStatsManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def get_overall_stats(self, time_range="24h"):
        """Получает общую статистику заданий"""
        try:
            request = jobs_pb2.GetJobStatsRequest(time_range=time_range)
            response = self.stub.GetJobStats(request, metadata=self.metadata)
            
            if response.success:
                return response.stats
            else:
                raise Exception(f"Ошибка получения статистики: {response.message}")
                
        except grpc.RpcError as e:
            raise Exception(f"gRPC Error: {e.details()}")
    
    def get_stats_for_job_type(self, job_type, time_range="24h"):
        """Получает статистику для конкретного типа задания"""
        try:
            request = jobs_pb2.GetJobStatsRequest(
                job_type=job_type,
                time_range=time_range
            )
            response = self.stub.GetJobStats(request, metadata=self.metadata)
            
            if response.success:
                return response.stats
            else:
                raise Exception(f"Ошибка получения статистики: {response.message}")
                
        except grpc.RpcError as e:
            raise Exception(f"gRPC Error: {e.details()}")
    
    def get_stats_for_worker(self, worker, time_range="24h"):
        """Получает статистику для конкретного воркера"""
        try:
            request = jobs_pb2.GetJobStatsRequest(
                worker=worker,
                time_range=time_range
            )
            response = self.stub.GetJobStats(request, metadata=self.metadata)
            
            if response.success:
                return response.stats
            else:
                raise Exception(f"Ошибка получения статистики: {response.message}")
                
        except grpc.RpcError as e:
            raise Exception(f"gRPC Error: {e.details()}")
    
    def print_detailed_report(self, time_range="24h"):
        """Выводит детальный отчет"""
        stats = self.get_overall_stats(time_range)
        
        print(f"📊 Детальный отчет по заданиям ({time_range})")
        print("═" * 43)
        print(f"Время генерации: {stats.generated_at}")
        print(f"Общее количество заданий: {stats.total_jobs}\n")
        
        # Анализ по состояниям
        print("🔍 Анализ по состояниям:")
        activatable = stats.by_state.get("ACTIVATABLE", 0)
        activated = stats.by_state.get("ACTIVATED", 0)
        completed = stats.by_state.get("COMPLETED", 0)
        failed = stats.by_state.get("FAILED", 0)
        cancelled = stats.by_state.get("CANCELLED", 0)
        
        print(f"  🟡 Готовые к активации: {activatable}")
        print(f"  🔵 В процессе выполнения: {activated}")
        print(f"  🟢 Завершенные: {completed}")
        print(f"  🔴 Провалившиеся: {failed}")
        print(f"  ⚫ Отмененные: {cancelled}")
        
        # Анализ нагрузки
        if activatable > 0 or activated > 0:
            print("\n⚡ Анализ нагрузки:")
            
            if activatable > 50:
                print(f"  ⚠️ Много заданий ожидают активации ({activatable})")
            
            if activated > 100:
                print(f"  ⚠️ Высокая нагрузка выполнения ({activated})")
            
            pending_total = activatable + activated
            if pending_total > 0:
                print(f"  📋 Всего в обработке: {pending_total} заданий")
        
        # Анализ производительности
        if stats.performance:
            perf = stats.performance
            print("\n📈 Анализ производительности:")
            
            # Время выполнения
            avg_exec_sec = perf.avg_execution_time / 1000
            print(f"  ⏱️ Среднее время выполнения: {avg_exec_sec:.1f}с")
            
            if avg_exec_sec > 60:
                print("     ⚠️ Медленное выполнение (>1 минуты)")
            elif avg_exec_sec < 1:
                print("     ✅ Быстрое выполнение (<1 секунды)")
            
            # Время ожидания
            avg_wait_sec = perf.avg_wait_time / 1000
            print(f"  ⏳ Среднее время ожидания: {avg_wait_sec:.1f}с")
            
            if avg_wait_sec > 300:
                print("     ⚠️ Долгое ожидание (>5 минут)")
            
            # Успешность
            print(f"  📊 Процент успешности: {perf.success_rate:.1f}%")
            
            if perf.success_rate < 80:
                print("     ❌ Низкая успешность (<80%)")
            elif perf.success_rate > 95:
                print("     ✅ Отличная успешность (>95%)")
            
            # Повторы
            print(f"  🔄 Среднее повторов: {perf.avg_retries_per_job:.2f}")
            
            if perf.avg_retries_per_job > 2:
                print("     ⚠️ Много повторов (>2 в среднем)")
        
        # Топ типов заданий
        if stats.by_type:
            print("\n🏷️ Топ типов заданий:")
            
            # Сортируем по количеству
            sorted_types = sorted(stats.by_type.items(), key=lambda x: x[1], reverse=True)
            
            # Показываем топ-5
            for i, (job_type, count) in enumerate(sorted_types[:5]):
                percentage = (count / stats.total_jobs * 100) if stats.total_jobs > 0 else 0
                print(f"  {i+1}. {job_type}: {count} ({percentage:.1f}%)")
        
        # Активные воркеры
        if stats.by_worker:
            print("\n👥 Активные воркеры:")
            for worker, count in stats.by_worker.items():
                print(f"  {worker}: {count} заданий")
    
    def compare_time_ranges(self, range1, range2):
        """Сравнивает статистику между двумя периодами"""
        stats1 = self.get_overall_stats(range1)
        stats2 = self.get_overall_stats(range2)
        
        print(f"📊 Сравнение периодов: {range1} vs {range2}")
        print("═" * 39)
        
        # Сравнение общих метрик
        print(f"{'Метрика':<20} | {range1:<15} | {range2:<15} | Изменение")
        print("─" * 62)
        
        self._compare_metric("Всего заданий", stats1.total_jobs, stats2.total_jobs)
        self._compare_metric("Завершенные", stats1.by_state.get("COMPLETED", 0), stats2.by_state.get("COMPLETED", 0))
        self._compare_metric("Провалившиеся", stats1.by_state.get("FAILED", 0), stats2.by_state.get("FAILED", 0))
        
        if stats1.performance and stats2.performance:
            perf1 = stats1.performance
            perf2 = stats2.performance
            
            print("─" * 62)
            self._compare_float_metric("Успешность (%)", perf1.success_rate, perf2.success_rate)
            self._compare_float_metric("Ср. выполнение (с)", perf1.avg_execution_time/1000, perf2.avg_execution_time/1000)
            self._compare_float_metric("Ср. ожидание (с)", perf1.avg_wait_time/1000, perf2.avg_wait_time/1000)
            self._compare_float_metric("Ср. повторов", perf1.avg_retries_per_job, perf2.avg_retries_per_job)
    
    def _compare_metric(self, name, val1, val2):
        diff = val2 - val1
        
        if diff > 0:
            change = f"↗️ +{diff}"
        elif diff < 0:
            change = f"↘️ {diff}"
        else:
            change = "➡️ 0"
        
        print(f"{name:<20} | {val1:<15} | {val2:<15} | {change}")
    
    def _compare_float_metric(self, name, val1, val2):
        diff = val2 - val1
        
        if diff > 0.1:
            change = f"↗️ +{diff:.2f}"
        elif diff < -0.1:
            change = f"↘️ {diff:.2f}"
        else:
            change = "➡️ ~0"
        
        print(f"{name:<20} | {val1:<15.2f} | {val2:<15.2f} | {change}")
    
    def monitor_stats(self, interval=30):
        """Мониторинг статистики в реальном времени"""
        print(f"📊 Мониторинг статистики заданий (обновление каждые {interval}с)")
        print("Нажмите Ctrl+C для остановки\n")
        
        prev_stats = None
        
        try:
            while True:
                try:
                    stats = self.get_overall_stats("1h")
                    now = datetime.now().strftime("%H:%M:%S")
                    
                    # Отображаем текущие метрики
                    activatable = stats.by_state.get("ACTIVATABLE", 0)
                    activated = stats.by_state.get("ACTIVATED", 0)
                    completed = stats.by_state.get("COMPLETED", 0)
                    failed = stats.by_state.get("FAILED", 0)
                    
                    print(f"[{now}] Всего: {stats.total_jobs} | Ожидает: {activatable} | "
                          f"Выполняется: {activated} | Завершено: {completed} | Провалено: {failed}")
                    
                    # Показываем изменения
                    if prev_stats:
                        changes = self._detect_changes(prev_stats, stats)
                        if changes:
                            print(f"         Изменения: {changes}")
                    
                    prev_stats = stats
                    
                except Exception as e:
                    print(f"❌ Ошибка получения статистики: {e}")
                
                time.sleep(interval)
                
        except KeyboardInterrupt:
            print("\n🛑 Мониторинг остановлен пользователем")
    
    def _detect_changes(self, prev_stats, curr_stats):
        changes = []
        
        # Проверяем изменения в состояниях
        for state, curr_count in curr_stats.by_state.items():
            prev_count = prev_stats.by_state.get(state, 0)
            if curr_count != prev_count:
                diff = curr_count - prev_count
                if diff > 0:
                    changes.append(f"{state} +{diff}")
                else:
                    changes.append(f"{state} {diff}")
        
        if not changes:
            return ""
        
        if len(changes) > 3:
            return f"{changes[0]} и еще {len(changes)-1}"
        
        return ", ".join(changes)
    
    def export_stats_to_csv(self, time_range="24h"):
        """Экспорт статистики в CSV"""
        stats = self.get_overall_stats(time_range)
        
        output = io.StringIO()
        writer = csv.writer(output)
        
        # Заголовок
        writer.writerow(["Type", "Metric", "Value"])
        
        # Общие метрики
        writer.writerow(["General", "Total Jobs", stats.total_jobs])
        writer.writerow(["General", "Generated At", stats.generated_at])
        
        # По состояниям
        for state, count in stats.by_state.items():
            writer.writerow(["State", state, count])
        
        # По типам
        for job_type, count in stats.by_type.items():
            writer.writerow(["JobType", job_type, count])
        
        # По воркерам
        for worker, count in stats.by_worker.items():
            writer.writerow(["Worker", worker, count])
        
        # Производительность
        if stats.performance:
            perf = stats.performance
            writer.writerow(["Performance", "Avg Execution Time", f"{perf.avg_execution_time:.2f}"])
            writer.writerow(["Performance", "Avg Wait Time", f"{perf.avg_wait_time:.2f}"])
            writer.writerow(["Performance", "Success Rate", f"{perf.success_rate:.2f}"])
            writer.writerow(["Performance", "Successful Jobs", perf.successful_jobs])
            writer.writerow(["Performance", "Failed Jobs", perf.failed_jobs])
            writer.writerow(["Performance", "Retry Count", perf.retry_count])
            writer.writerow(["Performance", "Avg Retries Per Job", f"{perf.avg_retries_per_job:.2f}"])
        
        return output.getvalue()

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python get_job_stats.py show [time_range]")
        print("  python get_job_stats.py type <job_type> [time_range]")
        print("  python get_job_stats.py worker <worker> [time_range]")
        print("  python get_job_stats.py compare <range1> <range2>")
        print("  python get_job_stats.py monitor [interval]")
        print("  python get_job_stats.py export [time_range]")
        sys.exit(1)
    
    manager = JobStatsManager()
    command = sys.argv[1]
    
    try:
        if command == "show":
            time_range = sys.argv[2] if len(sys.argv) > 2 else "24h"
            manager.print_detailed_report(time_range)
            
        elif command == "type":
            if len(sys.argv) < 3:
                print("❌ Укажите тип задания")
                sys.exit(1)
            job_type = sys.argv[2]
            time_range = sys.argv[3] if len(sys.argv) > 3 else "24h"
            stats = manager.get_stats_for_job_type(job_type, time_range)
            get_job_stats(job_type, "", time_range)
            
        elif command == "worker":
            if len(sys.argv) < 3:
                print("❌ Укажите имя воркера")
                sys.exit(1)
            worker = sys.argv[2]
            time_range = sys.argv[3] if len(sys.argv) > 3 else "24h"
            stats = manager.get_stats_for_worker(worker, time_range)
            get_job_stats("", worker, time_range)
            
        elif command == "compare":
            if len(sys.argv) < 4:
                print("❌ Укажите два временных диапазона для сравнения")
                sys.exit(1)
            range1 = sys.argv[2]
            range2 = sys.argv[3]
            manager.compare_time_ranges(range1, range2)
            
        elif command == "monitor":
            interval = int(sys.argv[2]) if len(sys.argv) > 2 else 30
            manager.monitor_stats(interval)
            
        elif command == "export":
            time_range = sys.argv[2] if len(sys.argv) > 2 else "24h"
            csv_data = manager.export_stats_to_csv(time_range)
            print(csv_data)
            
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

async function getJobStats(jobType = "", worker = "", timeRange = "24h") {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_type: jobType,
            worker: worker,
            time_range: timeRange
        };
        
        client.getJobStats(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                const stats = response.stats;
                console.log(`📊 Статистика заданий за ${timeRange}:\n`);
                
                // Общая информация
                console.log(`Общее количество заданий: ${stats.total_jobs}`);
                console.log(`Время генерации: ${stats.generated_at}\n`);
                
                // Статистика по состояниям
                console.log("По состояниям:");
                Object.entries(stats.by_state).forEach(([state, count]) => {
                    const percentage = stats.total_jobs > 0 ? (count / stats.total_jobs * 100) : 0;
                    console.log(`  ${state}: ${count} (${percentage.toFixed(1)}%)`);
                });
                
                // Статистика по типам
                if (Object.keys(stats.by_type).length > 0) {
                    console.log("\nПо типам:");
                    Object.entries(stats.by_type).forEach(([jobTypeName, count]) => {
                        const percentage = stats.total_jobs > 0 ? (count / stats.total_jobs * 100) : 0;
                        console.log(`  ${jobTypeName}: ${count} (${percentage.toFixed(1)}%)`);
                    });
                }
                
                // Статистика по воркерам
                if (Object.keys(stats.by_worker).length > 0) {
                    console.log("\nПо воркерам:");
                    Object.entries(stats.by_worker).forEach(([workerName, count]) => {
                        console.log(`  ${workerName}: ${count} заданий`);
                    });
                }
                
                // Метрики производительности
                if (stats.performance) {
                    const perf = stats.performance;
                    console.log("\nПроизводительность:");
                    console.log(`  Среднее время выполнения: ${perf.avg_execution_time.toFixed(1)}мс`);
                    console.log(`  Среднее время ожидания: ${perf.avg_wait_time.toFixed(1)}мс`);
                    console.log(`  Успешных заданий: ${perf.successful_jobs}`);
                    console.log(`  Провалившихся заданий: ${perf.failed_jobs}`);
                    console.log(`  Процент успешности: ${perf.success_rate.toFixed(1)}%`);
                    console.log(`  Всего повторов: ${perf.retry_count}`);
                    console.log(`  Среднее повторов на задание: ${perf.avg_retries_per_job.toFixed(2)}`);
                }
                
                resolve(stats);
            } else {
                console.log(`❌ Ошибка получения статистики: ${response.message}`);
                resolve(null);
            }
        });
    });
}

class JobStatsManager {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async getOverallStats(timeRange = "24h") {
        return new Promise((resolve, reject) => {
            const request = { time_range: timeRange };
            
            this.client.getJobStats(request, this.metadata, (error, response) => {
                if (error) {
                    reject(new Error(`gRPC Error: ${error.message}`));
                    return;
                }
                
                if (response.success) {
                    resolve(response.stats);
                } else {
                    reject(new Error(`Ошибка получения статистики: ${response.message}`));
                }
            });
        });
    }
    
    async getStatsForJobType(jobType, timeRange = "24h") {
        return new Promise((resolve, reject) => {
            const request = {
                job_type: jobType,
                time_range: timeRange
            };
            
            this.client.getJobStats(request, this.metadata, (error, response) => {
                if (error) {
                    reject(new Error(`gRPC Error: ${error.message}`));
                    return;
                }
                
                if (response.success) {
                    resolve(response.stats);
                } else {
                    reject(new Error(`Ошибка получения статистики: ${response.message}`));
                }
            });
        });
    }
    
    async getStatsForWorker(worker, timeRange = "24h") {
        return new Promise((resolve, reject) => {
            const request = {
                worker: worker,
                time_range: timeRange
            };
            
            this.client.getJobStats(request, this.metadata, (error, response) => {
                if (error) {
                    reject(new Error(`gRPC Error: ${error.message}`));
                    return;
                }
                
                if (response.success) {
                    resolve(response.stats);
                } else {
                    reject(new Error(`Ошибка получения статистики: ${response.message}`));
                }
            });
        });
    }
    
    async printDetailedReport(timeRange = "24h") {
        const stats = await this.getOverallStats(timeRange);
        
        console.log(`📊 Детальный отчет по заданиям (${timeRange})`);
        console.log('═'.repeat(43));
        console.log(`Время генерации: ${stats.generated_at}`);
        console.log(`Общее количество заданий: ${stats.total_jobs}\n`);
        
        // Анализ по состояниям
        console.log("🔍 Анализ по состояниям:");
        const activatable = stats.by_state["ACTIVATABLE"] || 0;
        const activated = stats.by_state["ACTIVATED"] || 0;
        const completed = stats.by_state["COMPLETED"] || 0;
        const failed = stats.by_state["FAILED"] || 0;
        const cancelled = stats.by_state["CANCELLED"] || 0;
        
        console.log(`  🟡 Готовые к активации: ${activatable}`);
        console.log(`  🔵 В процессе выполнения: ${activated}`);
        console.log(`  🟢 Завершенные: ${completed}`);
        console.log(`  🔴 Провалившиеся: ${failed}`);
        console.log(`  ⚫ Отмененные: ${cancelled}`);
        
        // Анализ нагрузки
        if (activatable > 0 || activated > 0) {
            console.log("\n⚡ Анализ нагрузки:");
            
            if (activatable > 50) {
                console.log(`  ⚠️ Много заданий ожидают активации (${activatable})`);
            }
            
            if (activated > 100) {
                console.log(`  ⚠️ Высокая нагрузка выполнения (${activated})`);
            }
            
            const pendingTotal = activatable + activated;
            if (pendingTotal > 0) {
                console.log(`  📋 Всего в обработке: ${pendingTotal} заданий`);
            }
        }
        
        // Анализ производительности
        if (stats.performance) {
            const perf = stats.performance;
            console.log("\n📈 Анализ производительности:");
            
            // Время выполнения
            const avgExecSec = perf.avg_execution_time / 1000;
            console.log(`  ⏱️ Среднее время выполнения: ${avgExecSec.toFixed(1)}с`);
            
            if (avgExecSec > 60) {
                console.log("     ⚠️ Медленное выполнение (>1 минуты)");
            } else if (avgExecSec < 1) {
                console.log("     ✅ Быстрое выполнение (<1 секунды)");
            }
            
            // Время ожидания
            const avgWaitSec = perf.avg_wait_time / 1000;
            console.log(`  ⏳ Среднее время ожидания: ${avgWaitSec.toFixed(1)}с`);
            
            if (avgWaitSec > 300) {
                console.log("     ⚠️ Долгое ожидание (>5 минут)");
            }
            
            // Успешность
            console.log(`  📊 Процент успешности: ${perf.success_rate.toFixed(1)}%`);
            
            if (perf.success_rate < 80) {
                console.log("     ❌ Низкая успешность (<80%)");
            } else if (perf.success_rate > 95) {
                console.log("     ✅ Отличная успешность (>95%)");
            }
            
            // Повторы
            console.log(`  🔄 Среднее повторов: ${perf.avg_retries_per_job.toFixed(2)}`);
            
            if (perf.avg_retries_per_job > 2) {
                console.log("     ⚠️ Много повторов (>2 в среднем)");
            }
        }
        
        // Топ типов заданий
        if (Object.keys(stats.by_type).length > 0) {
            console.log("\n🏷️ Топ типов заданий:");
            
            // Сортируем по количеству
            const sortedTypes = Object.entries(stats.by_type)
                .sort(([,a], [,b]) => b - a);
            
            // Показываем топ-5
            for (let i = 0; i < Math.min(5, sortedTypes.length); i++) {
                const [jobType, count] = sortedTypes[i];
                const percentage = stats.total_jobs > 0 ? (count / stats.total_jobs * 100) : 0;
                console.log(`  ${i+1}. ${jobType}: ${count} (${percentage.toFixed(1)}%)`);
            }
        }
        
        // Активные воркеры
        if (Object.keys(stats.by_worker).length > 0) {
            console.log("\n👥 Активные воркеры:");
            Object.entries(stats.by_worker).forEach(([worker, count]) => {
                console.log(`  ${worker}: ${count} заданий`);
            });
        }
    }
    
    async compareTimeRanges(range1, range2) {
        const stats1 = await this.getOverallStats(range1);
        const stats2 = await this.getOverallStats(range2);
        
        console.log(`📊 Сравнение периодов: ${range1} vs ${range2}`);
        console.log('═'.repeat(39));
        
        // Сравнение общих метрик
        console.log(`${'Метрика'.padEnd(20)} | ${range1.padEnd(15)} | ${range2.padEnd(15)} | Изменение`);
        console.log('─'.repeat(62));
        
        this._compareMetric("Всего заданий", stats1.total_jobs, stats2.total_jobs);
        this._compareMetric("Завершенные", stats1.by_state["COMPLETED"] || 0, stats2.by_state["COMPLETED"] || 0);
        this._compareMetric("Провалившиеся", stats1.by_state["FAILED"] || 0, stats2.by_state["FAILED"] || 0);
        
        if (stats1.performance && stats2.performance) {
            const perf1 = stats1.performance;
            const perf2 = stats2.performance;
            
            console.log('─'.repeat(62));
            this._compareFloatMetric("Успешность (%)", perf1.success_rate, perf2.success_rate);
            this._compareFloatMetric("Ср. выполнение (с)", perf1.avg_execution_time/1000, perf2.avg_execution_time/1000);
            this._compareFloatMetric("Ср. ожидание (с)", perf1.avg_wait_time/1000, perf2.avg_wait_time/1000);
            this._compareFloatMetric("Ср. повторов", perf1.avg_retries_per_job, perf2.avg_retries_per_job);
        }
    }
    
    _compareMetric(name, val1, val2) {
        const diff = val2 - val1;
        let change;
        
        if (diff > 0) {
            change = `↗️ +${diff}`;
        } else if (diff < 0) {
            change = `↘️ ${diff}`;
        } else {
            change = "➡️ 0";
        }
        
        console.log(`${name.padEnd(20)} | ${val1.toString().padEnd(15)} | ${val2.toString().padEnd(15)} | ${change}`);
    }
    
    _compareFloatMetric(name, val1, val2) {
        const diff = val2 - val1;
        let change;
        
        if (diff > 0.1) {
            change = `↗️ +${diff.toFixed(2)}`;
        } else if (diff < -0.1) {
            change = `↘️ ${diff.toFixed(2)}`;
        } else {
            change = "➡️ ~0";
        }
        
        console.log(`${name.padEnd(20)} | ${val1.toFixed(2).padEnd(15)} | ${val2.toFixed(2).padEnd(15)} | ${change}`);
    }
    
    async monitorStats(interval = 30000) {
        console.log(`📊 Мониторинг статистики заданий (обновление каждые ${interval/1000}с)`);
        console.log('Нажмите Ctrl+C для остановки\n');
        
        let prevStats = null;
        
        const monitor = setInterval(async () => {
            try {
                const stats = await this.getOverallStats("1h");
                const now = new Date().toLocaleTimeString();
                
                // Отображаем текущие метрики
                const activatable = stats.by_state["ACTIVATABLE"] || 0;
                const activated = stats.by_state["ACTIVATED"] || 0;
                const completed = stats.by_state["COMPLETED"] || 0;
                const failed = stats.by_state["FAILED"] || 0;
                
                console.log(`[${now}] Всего: ${stats.total_jobs} | Ожидает: ${activatable} | ` +
                           `Выполняется: ${activated} | Завершено: ${completed} | Провалено: ${failed}`);
                
                // Показываем изменения
                if (prevStats) {
                    const changes = this._detectChanges(prevStats, stats);
                    if (changes) {
                        console.log(`         Изменения: ${changes}`);
                    }
                }
                
                prevStats = stats;
                
            } catch (error) {
                console.log(`❌ Ошибка получения статистики: ${error.message}`);
            }
        }, interval);
        
        // Обработка Ctrl+C
        process.on('SIGINT', () => {
            console.log('\n🛑 Мониторинг остановлен пользователем');
            clearInterval(monitor);
            process.exit(0);
        });
    }
    
    _detectChanges(prevStats, currStats) {
        const changes = [];
        
        // Проверяем изменения в состояниях
        Object.entries(currStats.by_state).forEach(([state, currCount]) => {
            const prevCount = prevStats.by_state[state] || 0;
            if (currCount !== prevCount) {
                const diff = currCount - prevCount;
                if (diff > 0) {
                    changes.push(`${state} +${diff}`);
                } else {
                    changes.push(`${state} ${diff}`);
                }
            }
        });
        
        if (changes.length === 0) {
            return "";
        }
        
        if (changes.length > 3) {
            return `${changes[0]} и еще ${changes.length - 1}`;
        }
        
        return changes.join(", ");
    }
    
    async exportStatsToCSV(timeRange = "24h") {
        const stats = await this.getOverallStats(timeRange);
        
        let csv = "Type,Metric,Value\n";
        
        // Общие метрики
        csv += `General,Total Jobs,${stats.total_jobs}\n`;
        csv += `General,Generated At,${stats.generated_at}\n`;
        
        // По состояниям
        Object.entries(stats.by_state).forEach(([state, count]) => {
            csv += `State,${state},${count}\n`;
        });
        
        // По типам
        Object.entries(stats.by_type).forEach(([jobType, count]) => {
            csv += `JobType,${jobType},${count}\n`;
        });
        
        // По воркерам
        Object.entries(stats.by_worker).forEach(([worker, count]) => {
            csv += `Worker,${worker},${count}\n`;
        });
        
        // Производительность
        if (stats.performance) {
            const perf = stats.performance;
            csv += `Performance,Avg Execution Time,${perf.avg_execution_time.toFixed(2)}\n`;
            csv += `Performance,Avg Wait Time,${perf.avg_wait_time.toFixed(2)}\n`;
            csv += `Performance,Success Rate,${perf.success_rate.toFixed(2)}\n`;
            csv += `Performance,Successful Jobs,${perf.successful_jobs}\n`;
            csv += `Performance,Failed Jobs,${perf.failed_jobs}\n`;
            csv += `Performance,Retry Count,${perf.retry_count}\n`;
            csv += `Performance,Avg Retries Per Job,${perf.avg_retries_per_job.toFixed(2)}\n`;
        }
        
        return csv;
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node get-job-stats.js show [time_range]');
        console.log('  node get-job-stats.js type <job_type> [time_range]');
        console.log('  node get-job-stats.js worker <worker> [time_range]');
        console.log('  node get-job-stats.js compare <range1> <range2>');
        console.log('  node get-job-stats.js monitor [interval_ms]');
        console.log('  node get-job-stats.js export [time_range]');
        process.exit(1);
    }
    
    const manager = new JobStatsManager();
    const command = args[0];
    
    (async () => {
        try {
            switch (command) {
                case 'show':
                    const timeRange = args[1] || "24h";
                    await manager.printDetailedReport(timeRange);
                    break;
                    
                case 'type':
                    if (args.length < 2) {
                        console.log('❌ Укажите тип задания');
                        process.exit(1);
                    }
                    await getJobStats(args[1], "", args[2] || "24h");
                    break;
                    
                case 'worker':
                    if (args.length < 2) {
                        console.log('❌ Укажите имя воркера');
                        process.exit(1);
                    }
                    await getJobStats("", args[1], args[2] || "24h");
                    break;
                    
                case 'compare':
                    if (args.length < 3) {
                        console.log('❌ Укажите два временных диапазона для сравнения');
                        process.exit(1);
                    }
                    await manager.compareTimeRanges(args[1], args[2]);
                    break;
                    
                case 'monitor':
                    const interval = args[1] ? parseInt(args[1]) : 30000;
                    await manager.monitorStats(interval);
                    break;
                    
                case 'export':
                    const exportRange = args[1] || "24h";
                    const csvData = await manager.exportStatsToCSV(exportRange);
                    console.log(csvData);
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
    getJobStats,
    JobStatsManager
};
```

## Временные диапазоны

### Поддерживаемые форматы
- **1h**: Последний час
- **24h**: Последние 24 часа  
- **7d**: Последние 7 дней
- **30d**: Последние 30 дней

### Метрики производительности
- **avg_execution_time**: Среднее время выполнения в миллисекундах
- **avg_wait_time**: Среднее время ожидания активации
- **success_rate**: Процент успешно завершенных заданий
- **avg_retries_per_job**: Среднее количество повторов на задание

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный временной диапазон
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Invalid time range format: must be 1h, 24h, 7d, or 30d",
  "stats": null
}
```

## Связанные методы
- [ListJobs](list-jobs.md) - Получение списка заданий для детального анализа
- [GetJob](get-job.md) - Детали конкретных заданий
- [ActivateJobs](activate-jobs.md) - Активация заданий
- [CompleteJob](complete-job.md) - Завершение заданий
