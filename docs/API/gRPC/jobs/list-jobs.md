# ListJobs

## Описание
Возвращает список заданий с возможностью фильтрации по типу, состоянию, воркеру и другим параметрам. Поддерживает пагинацию для работы с большими объемами данных.

## Синтаксис
```protobuf
rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
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

### ListJobsRequest
```protobuf
message ListJobsRequest {
  string job_type = 1;     // Фильтр по типу задания
  string state = 2;        // Фильтр по состоянию
  string worker = 3;       // Фильтр по воркеру
  int32 limit = 4;         // Количество записей (максимум 1000)
  int32 offset = 5;        // Смещение для пагинации
  string sort_by = 6;      // Поле для сортировки
  bool sort_desc = 7;      // Сортировка по убыванию
}
```

#### Поля:
- **job_type** (string, optional): Фильтр по типу задания (например, "service-task")
- **state** (string, optional): Фильтр по состоянию ("PENDING", "ACTIVATABLE", "ACTIVATED", "RUNNING", "COMPLETED", "FAILED", "CANCELLED")
- **worker** (string, optional): Фильтр по имени воркера
- **limit** (int32, optional): Количество записей на страницу (по умолчанию 10, максимум 1000)
- **offset** (int32, optional): Смещение для пагинации (по умолчанию 0)
- **sort_by** (string, optional): Поле сортировки ("created_at", "activated_at", "retries", "deadline")
- **sort_desc** (bool, optional): Сортировка по убыванию (по умолчанию false)

## Параметры ответа

### ListJobsResponse
```protobuf
message ListJobsResponse {
  bool success = 1;           // Статус успешности операции
  string message = 2;         // Сообщение о результате
  repeated Job jobs = 3;      // Список заданий
  int32 total = 4;           // Общее количество заданий
  bool has_more = 5;         // Есть ли еще записи
}

message Job {
  string job_key = 1;        // Ключ задания
  string job_type = 2;       // Тип задания
  string state = 3;          // Состояние задания
  string worker = 4;         // Имя воркера
  int32 retries = 5;         // Количество попыток
  string created_at = 6;     // Время создания
  string activated_at = 7;   // Время активации
  string deadline = 8;       // Крайний срок
  map<string, string> variables = 9; // Переменные задания
  map<string, string> custom_headers = 10; // Пользовательские заголовки
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
    
    // Простой список всех заданий
    response, err := client.ListJobs(ctx, &pb.ListJobsRequest{
        Limit: 10,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("📋 Найдено заданий: %d из %d\n", len(response.Jobs), response.Total)
        
        for _, job := range response.Jobs {
            fmt.Printf("• %s [%s] - %s (попыток: %d)\n", 
                       job.JobKey, job.JobType, job.State, job.Retries)
        }
        
        if response.HasMore {
            fmt.Printf("... и еще %d заданий\n", response.Total - len(response.Jobs))
        }
    } else {
        fmt.Printf("❌ Ошибка получения списка: %s\n", response.Message)
    }
}

// Менеджер для работы со списками заданий
type JobListManager struct {
    client pb.JobsServiceClient
    ctx    context.Context
}

func NewJobListManager(client pb.JobsServiceClient, ctx context.Context) *JobListManager {
    return &JobListManager{
        client: client,
        ctx:    ctx,
    }
}

func (jlm *JobListManager) GetAllJobs() ([]*pb.Job, error) {
    var allJobs []*pb.Job
    offset := int32(0)
    limit := int32(100)
    
    for {
        response, err := jlm.client.ListJobs(jlm.ctx, &pb.ListJobsRequest{
            Limit:  limit,
            Offset: offset,
        })
        
        if err != nil {
            return nil, fmt.Errorf("ошибка получения заданий: %v", err)
        }
        
        if !response.Success {
            return nil, fmt.Errorf("ошибка API: %s", response.Message)
        }
        
        allJobs = append(allJobs, response.Jobs...)
        
        if !response.HasMore {
            break
        }
        
        offset += limit
    }
    
    return allJobs, nil
}

func (jlm *JobListManager) GetJobsByType(jobType string) ([]*pb.Job, error) {
    response, err := jlm.client.ListJobs(jlm.ctx, &pb.ListJobsRequest{
        JobType: jobType,
        Limit:   1000,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка получения заданий по типу: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

func (jlm *JobListManager) GetActiveJobs() ([]*pb.Job, error) {
    response, err := jlm.client.ListJobs(jlm.ctx, &pb.ListJobsRequest{
        State: "ACTIVATABLE",
        Limit: 1000,
        SortBy: "created_at",
        SortDesc: true, // Новые сначала
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка получения активных заданий: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

func (jlm *JobListManager) GetFailedJobs() ([]*pb.Job, error) {
    response, err := jlm.client.ListJobs(jlm.ctx, &pb.ListJobsRequest{
        State: "FAILED",
        Limit: 1000,
        SortBy: "retries",
        SortDesc: true, // Больше попыток сначала
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка получения провалившихся заданий: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

func (jlm *JobListManager) GetJobsByWorker(workerName string) ([]*pb.Job, error) {
    response, err := jlm.client.ListJobs(jlm.ctx, &pb.ListJobsRequest{
        Worker: workerName,
        Limit:  1000,
        SortBy: "activated_at",
        SortDesc: true,
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка получения заданий воркера: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

func (jlm *JobListManager) SearchJobs(filters JobFilters) ([]*pb.Job, error) {
    request := &pb.ListJobsRequest{
        JobType:  filters.JobType,
        State:    filters.State,
        Worker:   filters.Worker,
        Limit:    filters.Limit,
        Offset:   filters.Offset,
        SortBy:   filters.SortBy,
        SortDesc: filters.SortDesc,
    }
    
    if request.Limit == 0 {
        request.Limit = 10
    }
    
    response, err := jlm.client.ListJobs(jlm.ctx, request)
    
    if err != nil {
        return nil, fmt.Errorf("ошибка поиска заданий: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

type JobFilters struct {
    JobType  string
    State    string
    Worker   string
    Limit    int32
    Offset   int32
    SortBy   string
    SortDesc bool
}

// Аналитические функции
type JobAnalytics struct {
    listManager *JobListManager
}

func NewJobAnalytics(listManager *JobListManager) *JobAnalytics {
    return &JobAnalytics{
        listManager: listManager,
    }
}

func (ja *JobAnalytics) GetJobStatsByType() (map[string]int, error) {
    jobs, err := ja.listManager.GetAllJobs()
    if err != nil {
        return nil, err
    }
    
    stats := make(map[string]int)
    
    for _, job := range jobs {
        stats[job.JobType]++
    }
    
    return stats, nil
}

func (ja *JobAnalytics) GetJobStatsByState() (map[string]int, error) {
    jobs, err := ja.listManager.GetAllJobs()
    if err != nil {
        return nil, err
    }
    
    stats := make(map[string]int)
    
    for _, job := range jobs {
        stats[job.State]++
    }
    
    return stats, nil
}

func (ja *JobAnalytics) GetWorkerLoad() (map[string]int, error) {
    jobs, err := ja.listManager.GetAllJobs()
    if err != nil {
        return nil, err
    }
    
    load := make(map[string]int)
    
    for _, job := range jobs {
        if job.Worker != "" && job.State == "ACTIVATED" {
            load[job.Worker]++
        }
    }
    
    return load, nil
}

func (ja *JobAnalytics) GetRetryDistribution() (map[int32]int, error) {
    jobs, err := ja.listManager.GetAllJobs()
    if err != nil {
        return nil, err
    }
    
    distribution := make(map[int32]int)
    
    for _, job := range jobs {
        distribution[job.Retries]++
    }
    
    return distribution, nil
}

func (ja *JobAnalytics) GetOldestJobs(limit int) ([]*pb.Job, error) {
    response, err := ja.listManager.client.ListJobs(ja.listManager.ctx, &pb.ListJobsRequest{
        State:    "ACTIVATABLE",
        Limit:    int32(limit),
        SortBy:   "created_at",
        SortDesc: false, // Старые сначала
    })
    
    if err != nil {
        return nil, fmt.Errorf("ошибка получения старых заданий: %v", err)
    }
    
    if !response.Success {
        return nil, fmt.Errorf("ошибка API: %s", response.Message)
    }
    
    return response.Jobs, nil
}

func (ja *JobAnalytics) PrintJobsSummary() error {
    fmt.Printf("📊 Анализ заданий:\n\n")
    
    // Статистика по типам
    typeStats, err := ja.GetJobStatsByType()
    if err != nil {
        return err
    }
    
    fmt.Printf("По типам:\n")
    for jobType, count := range typeStats {
        fmt.Printf("  %s: %d\n", jobType, count)
    }
    
    // Статистика по состояниям
    stateStats, err := ja.GetJobStatsByState()
    if err != nil {
        return err
    }
    
    fmt.Printf("\nПо состояниям:\n")
    for state, count := range stateStats {
        fmt.Printf("  %s: %d\n", state, count)
    }
    
    // Нагрузка воркеров
    workerLoad, err := ja.GetWorkerLoad()
    if err != nil {
        return err
    }
    
    if len(workerLoad) > 0 {
        fmt.Printf("\nНагрузка воркеров:\n")
        for worker, count := range workerLoad {
            fmt.Printf("  %s: %d активных заданий\n", worker, count)
        }
    }
    
    // Старые задания
    oldJobs, err := ja.GetOldestJobs(5)
    if err != nil {
        return err
    }
    
    if len(oldJobs) > 0 {
        fmt.Printf("\nСамые старые задания:\n")
        for _, job := range oldJobs {
            createdAt, _ := time.Parse(time.RFC3339, job.CreatedAt)
            age := time.Since(createdAt)
            fmt.Printf("  %s [%s] - возраст: %s\n", job.JobKey, job.JobType, age.String())
        }
    }
    
    return nil
}

// Пагинированный итератор
type JobIterator struct {
    listManager *JobListManager
    filters     JobFilters
    currentPage []*pb.Job
    pageIndex   int
    offset      int32
    hasMore     bool
}

func NewJobIterator(listManager *JobListManager, filters JobFilters) *JobIterator {
    if filters.Limit == 0 {
        filters.Limit = 50 // Размер страницы по умолчанию
    }
    
    return &JobIterator{
        listManager: listManager,
        filters:     filters,
        hasMore:     true,
    }
}

func (ji *JobIterator) Next() bool {
    if ji.pageIndex >= len(ji.currentPage) {
        // Загружаем следующую страницу
        if !ji.hasMore {
            return false
        }
        
        if err := ji.loadNextPage(); err != nil {
            fmt.Printf("Ошибка загрузки страницы: %v\n", err)
            return false
        }
        
        if len(ji.currentPage) == 0 {
            return false
        }
        
        ji.pageIndex = 0
    }
    
    return ji.pageIndex < len(ji.currentPage)
}

func (ji *JobIterator) Job() *pb.Job {
    if ji.pageIndex >= len(ji.currentPage) {
        return nil
    }
    
    job := ji.currentPage[ji.pageIndex]
    ji.pageIndex++
    return job
}

func (ji *JobIterator) loadNextPage() error {
    filters := ji.filters
    filters.Offset = ji.offset
    
    jobs, err := ji.listManager.SearchJobs(filters)
    if err != nil {
        return err
    }
    
    ji.currentPage = jobs
    ji.pageIndex = 0
    ji.offset += ji.filters.Limit
    ji.hasMore = len(jobs) == int(ji.filters.Limit)
    
    return nil
}

// Пример использования итератора
func processAllJobsOfType(listManager *JobListManager, jobType string) {
    iterator := NewJobIterator(listManager, JobFilters{
        JobType: jobType,
        Limit:   100,
    })
    
    processedCount := 0
    
    for iterator.Next() {
        job := iterator.Job()
        if job == nil {
            break
        }
        
        // Обрабатываем задание
        fmt.Printf("Обрабатываем задание: %s [%s]\n", job.JobKey, job.State)
        processedCount++
        
        // Можно добавить дополнительную логику обработки
        if job.State == "FAILED" && job.Retries > 5 {
            fmt.Printf("  ⚠️ Задание с большим количеством неудачных попыток: %d\n", job.Retries)
        }
    }
    
    fmt.Printf("📋 Обработано заданий типа '%s': %d\n", jobType, processedCount)
}
```

### Python
```python
import grpc
import time
from typing import List, Dict, Optional, Iterator
from dataclasses import dataclass
from datetime import datetime, timedelta

import jobs_pb2
import jobs_pb2_grpc

def list_jobs(job_type="", state="", worker="", limit=10, offset=0, sort_by="", sort_desc=False):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.ListJobsRequest(
        job_type=job_type,
        state=state,
        worker=worker,
        limit=limit,
        offset=offset,
        sort_by=sort_by,
        sort_desc=sort_desc
    )
    
    try:
        response = stub.ListJobs(request, metadata=metadata)
        
        if response.success:
            print(f"📋 Найдено заданий: {len(response.jobs)} из {response.total}")
            
            for job in response.jobs:
                print(f"• {job.job_key} [{job.job_type}] - {job.state} (попыток: {job.retries})")
            
            if response.has_more:
                print(f"... и еще {response.total - len(response.jobs)} заданий")
            
            return response.jobs
        else:
            print(f"❌ Ошибка получения списка: {response.message}")
            return []
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return []

@dataclass
class JobFilters:
    job_type: str = ""
    state: str = ""
    worker: str = ""
    limit: int = 10
    offset: int = 0
    sort_by: str = ""
    sort_desc: bool = False

class JobListManager:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def get_all_jobs(self):
        """Получает все задания с автоматической пагинацией"""
        all_jobs = []
        offset = 0
        limit = 100
        
        while True:
            try:
                request = jobs_pb2.ListJobsRequest(
                    limit=limit,
                    offset=offset
                )
                
                response = self.stub.ListJobs(request, metadata=self.metadata)
                
                if not response.success:
                    print(f"❌ Ошибка API: {response.message}")
                    break
                
                all_jobs.extend(response.jobs)
                
                if not response.has_more:
                    break
                
                offset += limit
                
            except grpc.RpcError as e:
                print(f"gRPC Error: {e.details()}")
                break
        
        return all_jobs
    
    def get_jobs_by_type(self, job_type):
        """Получает задания определенного типа"""
        try:
            request = jobs_pb2.ListJobsRequest(
                job_type=job_type,
                limit=1000
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []
    
    def get_active_jobs(self):
        """Получает все активные задания"""
        try:
            request = jobs_pb2.ListJobsRequest(
                state="ACTIVATABLE",
                limit=1000,
                sort_by="created_at",
                sort_desc=True  # Новые сначала
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []
    
    def get_failed_jobs(self):
        """Получает провалившиеся задания"""
        try:
            request = jobs_pb2.ListJobsRequest(
                state="FAILED",
                limit=1000,
                sort_by="retries",
                sort_desc=True  # Больше попыток сначала
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []
    
    def get_jobs_by_worker(self, worker_name):
        """Получает задания конкретного воркера"""
        try:
            request = jobs_pb2.ListJobsRequest(
                worker=worker_name,
                limit=1000,
                sort_by="activated_at",
                sort_desc=True
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []
    
    def search_jobs(self, filters: JobFilters):
        """Поиск заданий с фильтрами"""
        try:
            request = jobs_pb2.ListJobsRequest(
                job_type=filters.job_type,
                state=filters.state,
                worker=filters.worker,
                limit=filters.limit or 10,
                offset=filters.offset,
                sort_by=filters.sort_by,
                sort_desc=filters.sort_desc
            )
            
            response = self.stub.ListJobs(request, metadata=self.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []

class JobAnalytics:
    def __init__(self, list_manager: JobListManager):
        self.list_manager = list_manager
    
    def get_job_stats_by_type(self):
        """Статистика заданий по типам"""
        jobs = self.list_manager.get_all_jobs()
        stats = {}
        
        for job in jobs:
            stats[job.job_type] = stats.get(job.job_type, 0) + 1
        
        return stats
    
    def get_job_stats_by_state(self):
        """Статистика заданий по состояниям"""
        jobs = self.list_manager.get_all_jobs()
        stats = {}
        
        for job in jobs:
            stats[job.state] = stats.get(job.state, 0) + 1
        
        return stats
    
    def get_worker_load(self):
        """Нагрузка воркеров"""
        jobs = self.list_manager.get_all_jobs()
        load = {}
        
        for job in jobs:
            if job.worker and job.state == "ACTIVATED":
                load[job.worker] = load.get(job.worker, 0) + 1
        
        return load
    
    def get_retry_distribution(self):
        """Распределение количества попыток"""
        jobs = self.list_manager.get_all_jobs()
        distribution = {}
        
        for job in jobs:
            distribution[job.retries] = distribution.get(job.retries, 0) + 1
        
        return distribution
    
    def get_oldest_jobs(self, limit=10):
        """Получает самые старые задания"""
        try:
            request = jobs_pb2.ListJobsRequest(
                state="ACTIVATABLE",
                limit=limit,
                sort_by="created_at",
                sort_desc=False  # Старые сначала
            )
            
            response = self.list_manager.stub.ListJobs(request, metadata=self.list_manager.metadata)
            
            if response.success:
                return list(response.jobs)
            else:
                print(f"❌ Ошибка API: {response.message}")
                return []
                
        except grpc.RpcError as e:
            print(f"gRPC Error: {e.details()}")
            return []
    
    def print_jobs_summary(self):
        """Выводит сводку по заданиям"""
        print("📊 Анализ заданий:\n")
        
        # Статистика по типам
        type_stats = self.get_job_stats_by_type()
        print("По типам:")
        for job_type, count in type_stats.items():
            print(f"  {job_type}: {count}")
        
        # Статистика по состояниям
        state_stats = self.get_job_stats_by_state()
        print("\nПо состояниям:")
        for state, count in state_stats.items():
            print(f"  {state}: {count}")
        
        # Нагрузка воркеров
        worker_load = self.get_worker_load()
        if worker_load:
            print("\nНагрузка воркеров:")
            for worker, count in worker_load.items():
                print(f"  {worker}: {count} активных заданий")
        
        # Старые задания
        old_jobs = self.get_oldest_jobs(5)
        if old_jobs:
            print("\nСамые старые задания:")
            for job in old_jobs:
                try:
                    created_at = datetime.fromisoformat(job.created_at.replace('Z', '+00:00'))
                    age = datetime.now() - created_at
                    print(f"  {job.job_key} [{job.job_type}] - возраст: {age}")
                except:
                    print(f"  {job.job_key} [{job.job_type}] - возраст: неизвестен")

class JobIterator:
    def __init__(self, list_manager: JobListManager, filters: JobFilters):
        self.list_manager = list_manager
        self.filters = filters
        if self.filters.limit == 0:
            self.filters.limit = 50  # Размер страницы по умолчанию
        
        self.current_page = []
        self.page_index = 0
        self.offset = 0
        self.has_more = True
    
    def __iter__(self):
        return self
    
    def __next__(self):
        if self.page_index >= len(self.current_page):
            # Загружаем следующую страницу
            if not self.has_more:
                raise StopIteration
            
            self._load_next_page()
            
            if len(self.current_page) == 0:
                raise StopIteration
            
            self.page_index = 0
        
        if self.page_index >= len(self.current_page):
            raise StopIteration
        
        job = self.current_page[self.page_index]
        self.page_index += 1
        return job
    
    def _load_next_page(self):
        filters = JobFilters(
            job_type=self.filters.job_type,
            state=self.filters.state,
            worker=self.filters.worker,
            limit=self.filters.limit,
            offset=self.offset,
            sort_by=self.filters.sort_by,
            sort_desc=self.filters.sort_desc
        )
        
        jobs = self.list_manager.search_jobs(filters)
        
        self.current_page = jobs
        self.page_index = 0
        self.offset += self.filters.limit
        self.has_more = len(jobs) == self.filters.limit

def process_all_jobs_of_type(list_manager: JobListManager, job_type: str):
    """Пример обработки всех заданий определенного типа"""
    iterator = JobIterator(list_manager, JobFilters(
        job_type=job_type,
        limit=100
    ))
    
    processed_count = 0
    
    for job in iterator:
        # Обрабатываем задание
        print(f"Обрабатываем задание: {job.job_key} [{job.state}]")
        processed_count += 1
        
        # Дополнительная логика
        if job.state == "FAILED" and job.retries > 5:
            print(f"  ⚠️ Задание с большим количеством неудачных попыток: {job.retries}")
    
    print(f"📋 Обработано заданий типа '{job_type}': {processed_count}")

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Использование:")
        print("  python list_jobs.py [filters...]")
        print("  python list_jobs.py --type service-task")
        print("  python list_jobs.py --state ACTIVATABLE")
        print("  python list_jobs.py --worker worker1")
        print("  python list_jobs.py --analytics")
        print("  python list_jobs.py --process-type service-task")
        sys.exit(1)
    
    list_manager = JobListManager()
    
    if "--analytics" in sys.argv:
        analytics = JobAnalytics(list_manager)
        analytics.print_jobs_summary()
    elif "--process-type" in sys.argv:
        idx = sys.argv.index("--process-type")
        if idx + 1 < len(sys.argv):
            job_type = sys.argv[idx + 1]
            process_all_jobs_of_type(list_manager, job_type)
        else:
            print("❌ Укажите тип задания после --process-type")
    else:
        # Парсим фильтры
        filters = JobFilters()
        
        for i in range(1, len(sys.argv), 2):
            if i + 1 >= len(sys.argv):
                break
            
            arg = sys.argv[i]
            value = sys.argv[i + 1]
            
            if arg == "--type":
                filters.job_type = value
            elif arg == "--state":
                filters.state = value
            elif arg == "--worker":
                filters.worker = value
            elif arg == "--limit":
                filters.limit = int(value)
            elif arg == "--sort":
                filters.sort_by = value
        
        jobs = list_manager.search_jobs(filters)
        
        if jobs:
            print(f"📋 Найдено заданий: {len(jobs)}")
            for job in jobs:
                print(f"• {job.job_key} [{job.job_type}] - {job.state} (попыток: {job.retries})")
        else:
            print("📋 Задания не найдены")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

async function listJobs(filters = {}) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_type: filters.jobType || "",
            state: filters.state || "",
            worker: filters.worker || "",
            limit: filters.limit || 10,
            offset: filters.offset || 0,
            sort_by: filters.sortBy || "",
            sort_desc: filters.sortDesc || false
        };
        
        client.listJobs(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`📋 Найдено заданий: ${response.jobs.length} из ${response.total}`);
                
                response.jobs.forEach(job => {
                    console.log(`• ${job.job_key} [${job.job_type}] - ${job.state} (попыток: ${job.retries})`);
                });
                
                if (response.has_more) {
                    console.log(`... и еще ${response.total - response.jobs.length} заданий`);
                }
                
                resolve(response.jobs);
            } else {
                console.log(`❌ Ошибка получения списка: ${response.message}`);
                resolve([]);
            }
        });
    });
}

class JobListManager {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async getAllJobs() {
        const allJobs = [];
        let offset = 0;
        const limit = 100;
        let hasMore = true;
        
        while (hasMore) {
            try {
                const jobs = await this._getJobsPage(limit, offset);
                allJobs.push(...jobs.jobs);
                
                hasMore = jobs.hasMore;
                offset += limit;
                
            } catch (error) {
                console.error(`Ошибка получения заданий: ${error.message}`);
                break;
            }
        }
        
        return allJobs;
    }
    
    async getJobsByType(jobType) {
        return await this._searchJobs({ jobType: jobType, limit: 1000 });
    }
    
    async getActiveJobs() {
        return await this._searchJobs({ 
            state: "ACTIVATABLE", 
            limit: 1000, 
            sortBy: "created_at", 
            sortDesc: true 
        });
    }
    
    async getFailedJobs() {
        return await this._searchJobs({ 
            state: "FAILED", 
            limit: 1000, 
            sortBy: "retries", 
            sortDesc: true 
        });
    }
    
    async getJobsByWorker(workerName) {
        return await this._searchJobs({ 
            worker: workerName, 
            limit: 1000, 
            sortBy: "activated_at", 
            sortDesc: true 
        });
    }
    
    async searchJobs(filters) {
        return await this._searchJobs(filters);
    }
    
    async _searchJobs(filters) {
        return new Promise((resolve, reject) => {
            const request = {
                job_type: filters.jobType || "",
                state: filters.state || "",
                worker: filters.worker || "",
                limit: filters.limit || 10,
                offset: filters.offset || 0,
                sort_by: filters.sortBy || "",
                sort_desc: filters.sortDesc || false
            };
            
            this.client.listJobs(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error: ${error.message}`);
                    resolve([]);
                    return;
                }
                
                if (response.success) {
                    resolve(response.jobs);
                } else {
                    console.log(`❌ Ошибка API: ${response.message}`);
                    resolve([]);
                }
            });
        });
    }
    
    async _getJobsPage(limit, offset) {
        return new Promise((resolve, reject) => {
            const request = { limit: limit, offset: offset };
            
            this.client.listJobs(request, this.metadata, (error, response) => {
                if (error) {
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    resolve({
                        jobs: response.jobs,
                        hasMore: response.has_more,
                        total: response.total
                    });
                } else {
                    reject(new Error(response.message));
                }
            });
        });
    }
}

class JobAnalytics {
    constructor(listManager) {
        this.listManager = listManager;
    }
    
    async getJobStatsByType() {
        const jobs = await this.listManager.getAllJobs();
        const stats = {};
        
        jobs.forEach(job => {
            stats[job.job_type] = (stats[job.job_type] || 0) + 1;
        });
        
        return stats;
    }
    
    async getJobStatsByState() {
        const jobs = await this.listManager.getAllJobs();
        const stats = {};
        
        jobs.forEach(job => {
            stats[job.state] = (stats[job.state] || 0) + 1;
        });
        
        return stats;
    }
    
    async getWorkerLoad() {
        const jobs = await this.listManager.getAllJobs();
        const load = {};
        
        jobs.forEach(job => {
            if (job.worker && job.state === "ACTIVATED") {
                load[job.worker] = (load[job.worker] || 0) + 1;
            }
        });
        
        return load;
    }
    
    async getRetryDistribution() {
        const jobs = await this.listManager.getAllJobs();
        const distribution = {};
        
        jobs.forEach(job => {
            distribution[job.retries] = (distribution[job.retries] || 0) + 1;
        });
        
        return distribution;
    }
    
    async getOldestJobs(limit = 10) {
        return await this.listManager._searchJobs({
            state: "ACTIVATABLE",
            limit: limit,
            sortBy: "created_at",
            sortDesc: false
        });
    }
    
    async printJobsSummary() {
        console.log("📊 Анализ заданий:\n");
        
        // Статистика по типам
        const typeStats = await this.getJobStatsByType();
        console.log("По типам:");
        Object.entries(typeStats).forEach(([jobType, count]) => {
            console.log(`  ${jobType}: ${count}`);
        });
        
        // Статистика по состояниям
        const stateStats = await this.getJobStatsByState();
        console.log("\nПо состояниям:");
        Object.entries(stateStats).forEach(([state, count]) => {
            console.log(`  ${state}: ${count}`);
        });
        
        // Нагрузка воркеров
        const workerLoad = await this.getWorkerLoad();
        if (Object.keys(workerLoad).length > 0) {
            console.log("\nНагрузка воркеров:");
            Object.entries(workerLoad).forEach(([worker, count]) => {
                console.log(`  ${worker}: ${count} активных заданий`);
            });
        }
        
        // Старые задания
        const oldJobs = await this.getOldestJobs(5);
        if (oldJobs.length > 0) {
            console.log("\nСамые старые задания:");
            oldJobs.forEach(job => {
                try {
                    const createdAt = new Date(job.created_at);
                    const age = Math.floor((Date.now() - createdAt.getTime()) / (1000 * 60 * 60));
                    console.log(`  ${job.job_key} [${job.job_type}] - возраст: ${age}ч`);
                } catch {
                    console.log(`  ${job.job_key} [${job.job_type}] - возраст: неизвестен`);
                }
            });
        }
    }
}

class JobIterator {
    constructor(listManager, filters = {}) {
        this.listManager = listManager;
        this.filters = { ...filters };
        if (!this.filters.limit) {
            this.filters.limit = 50; // Размер страницы по умолчанию
        }
        
        this.currentPage = [];
        this.pageIndex = 0;
        this.offset = 0;
        this.hasMore = true;
    }
    
    async *[Symbol.asyncIterator]() {
        while (true) {
            if (this.pageIndex >= this.currentPage.length) {
                // Загружаем следующую страницу
                if (!this.hasMore) {
                    break;
                }
                
                await this._loadNextPage();
                
                if (this.currentPage.length === 0) {
                    break;
                }
                
                this.pageIndex = 0;
            }
            
            if (this.pageIndex >= this.currentPage.length) {
                break;
            }
            
            const job = this.currentPage[this.pageIndex];
            this.pageIndex++;
            yield job;
        }
    }
    
    async _loadNextPage() {
        const filters = {
            ...this.filters,
            offset: this.offset
        };
        
        const jobs = await this.listManager.searchJobs(filters);
        
        this.currentPage = jobs;
        this.pageIndex = 0;
        this.offset += this.filters.limit;
        this.hasMore = jobs.length === this.filters.limit;
    }
}

async function processAllJobsOfType(listManager, jobType) {
    const iterator = new JobIterator(listManager, {
        jobType: jobType,
        limit: 100
    });
    
    let processedCount = 0;
    
    for await (const job of iterator) {
        // Обрабатываем задание
        console.log(`Обрабатываем задание: ${job.job_key} [${job.state}]`);
        processedCount++;
        
        // Дополнительная логика
        if (job.state === "FAILED" && job.retries > 5) {
            console.log(`  ⚠️ Задание с большим количеством неудачных попыток: ${job.retries}`);
        }
    }
    
    console.log(`📋 Обработано заданий типа '${jobType}': ${processedCount}`);
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node list-jobs.js [filters...]');
        console.log('  node list-jobs.js --type service-task');
        console.log('  node list-jobs.js --state ACTIVATABLE');
        console.log('  node list-jobs.js --worker worker1');
        console.log('  node list-jobs.js --analytics');
        console.log('  node list-jobs.js --process-type service-task');
        process.exit(1);
    }
    
    (async () => {
        const listManager = new JobListManager();
        
        if (args.includes('--analytics')) {
            const analytics = new JobAnalytics(listManager);
            await analytics.printJobsSummary();
        } else if (args.includes('--process-type')) {
            const idx = args.indexOf('--process-type');
            if (idx + 1 < args.length) {
                const jobType = args[idx + 1];
                await processAllJobsOfType(listManager, jobType);
            } else {
                console.log('❌ Укажите тип задания после --process-type');
            }
        } else {
            // Парсим фильтры
            const filters = {};
            
            for (let i = 0; i < args.length; i += 2) {
                if (i + 1 >= args.length) break;
                
                const arg = args[i];
                const value = args[i + 1];
                
                switch (arg) {
                    case '--type':
                        filters.jobType = value;
                        break;
                    case '--state':
                        filters.state = value;
                        break;
                    case '--worker':
                        filters.worker = value;
                        break;
                    case '--limit':
                        filters.limit = parseInt(value);
                        break;
                    case '--sort':
                        filters.sortBy = value;
                        break;
                }
            }
            
            const jobs = await listManager.searchJobs(filters);
            
            if (jobs.length > 0) {
                console.log(`📋 Найдено заданий: ${jobs.length}`);
                jobs.forEach(job => {
                    console.log(`• ${job.job_key} [${job.job_type}] - ${job.state} (попыток: ${job.retries})`);
                });
            } else {
                console.log('📋 Задания не найдены');
            }
        }
    })().catch(error => {
        console.error(`Ошибка: ${error.message}`);
        process.exit(1);
    });
}

module.exports = {
    listJobs,
    JobListManager,
    JobAnalytics,
    JobIterator,
    processAllJobsOfType
};
```

## Фильтры и сортировка

### Доступные состояния
- **PENDING**: Задания созданы и ожидают активации
- **ACTIVATABLE**: Задания готовые к активации (синоним для PENDING)
- **ACTIVATED**: Задания активированы и назначены воркерам (синоним для RUNNING)
- **RUNNING**: Задания выполняются воркерами
- **COMPLETED**: Успешно завершенные задания
- **FAILED**: Провалившиеся задания
- **CANCELLED**: Отмененные задания

### Поля сортировки
- **created_at**: По времени создания
- **activated_at**: По времени активации
- **retries**: По количеству попыток
- **deadline**: По крайнему сроку

### Пагинация
- **limit**: Максимум 1000 записей на запрос
- **offset**: Смещение для получения следующих страниц
- **has_more**: Указывает на наличие дополнительных записей

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверные параметры фильтрации
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Invalid limit value: must be between 1 and 1000",
  "jobs": [],
  "total": 0,
  "has_more": false
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Активация заданий для воркеров
- [GetJob](get-job.md) - Получение деталей конкретного задания
- [CancelJob](cancel-job.md) - Отмена заданий из списка
- [GetJobStats](get-job-stats.md) - Агрегированная статистика заданий
