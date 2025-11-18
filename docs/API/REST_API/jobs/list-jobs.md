# GET /api/v1/jobs

## Описание
Получение списка заданий с фильтрацией по статусу, типу, worker и другим параметрам.

## URL
```
GET /api/v1/jobs
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры запроса (Query Parameters)

### Фильтрация
- `state` (string): Фильтр по состоянию (`pending`, `activatable`, `activated`, `running`, `completed`, `failed`, `cancelled`)
- `type` (string): Фильтр по типу задания
- `worker` (string): Фильтр по worker ID
- `process_instance_id` (string): Фильтр по экземпляру процесса
- `element_id` (string): Фильтр по элементу BPMN
- `tenant_id` (string): Фильтр по тенанту

### Временная фильтрация
- `created_after` (string): Задания созданные после даты (ISO 8601)
- `created_before` (string): Задания созданные до даты (ISO 8601)
- `period` (string): Предустановленный период (`1h`, `24h`, `7d`, `30d`)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)
- `sort_by` (string): Поле сортировки (`created_at`, `activated_at`, `deadline`)
- `sort_order` (string): Порядок сортировки (`ASC`, `DESC`)

## Примеры запросов

### Все задания
```bash
curl -X GET "http://localhost:27555/api/v1/jobs" \
  -H "X-API-Key: your-api-key-here"
```

### Активные задания
```bash
curl -X GET "http://localhost:27555/api/v1/jobs?state=activatable" \
  -H "X-API-Key: your-api-key-here"
```

### Задания конкретного типа
```bash
curl -X GET "http://localhost:27555/api/v1/jobs?type=payment-processor&page_size=50" \
  -H "X-API-Key: your-api-key-here"
```

### Задания за последний час
```bash
curl -X GET "http://localhost:27555/api/v1/jobs?period=1h&state=completed" \
  -H "X-API-Key: your-api-key-here"
```

### Pending задания конкретного типа
```bash
curl -X GET "http://localhost:27555/api/v1/jobs?state=pending&type=payment_notification" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const params = new URLSearchParams({
  state: 'activatable',
  type: 'email-service',
  page: '1',
  page_size: '20'
});

const response = await fetch(`/api/v1/jobs?${params}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const jobs = await response.json();
```

## Ответы

### 200 OK - Список заданий
```json
{
  "success": true,
  "data": {
    "jobs": [
      {
        "job_key": "srv1-job-xyz789",
        "type": "payment-processor",
        "status": "ACTIVE",
        "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "element_id": "process-payment",
        "element_instance_id": "srv1-elem-payment01",
        "worker_id": "payment-worker-02",
        "tenant_id": "production",
        "created_at": "2025-01-11T10:30:00.000Z",
        "activated_at": "2025-01-11T10:31:30.789Z",
        "deadline": "2025-01-11T10:33:30.789Z",
        "retries": 3,
        "variables": {
          "orderId": "ORD-12345",
          "amount": 299.99,
          "paymentMethod": "credit"
        },
        "custom_headers": {
          "priority": "high",
          "payment_provider": "stripe"
        },
        "timeout_ms": 120000,
        "elapsed_time_ms": 90000,
        "remaining_time_ms": 30000
      },
      {
        "job_key": "srv1-job-abc123",
        "type": "email-sender",
        "status": "COMPLETED",
        "process_instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
        "element_id": "send-confirmation",
        "element_instance_id": "srv1-elem-email01",
        "worker_id": "email-worker-01",
        "tenant_id": "production",
        "created_at": "2025-01-11T10:25:00.000Z",
        "activated_at": "2025-01-11T10:25:30.123Z",
        "completed_at": "2025-01-11T10:26:15.456Z",
        "deadline": "2025-01-11T10:27:30.123Z",
        "retries": 3,
        "variables": {
          "recipient": "customer@example.com",
          "subject": "Order Confirmation",
          "orderId": "ORD-12340"
        },
        "custom_headers": {
          "priority": "normal",
          "template": "order-confirmation"
        },
        "execution_summary": {
          "total_duration_ms": 45333,
          "retries_used": 0,
          "result_variables": {
            "emailSent": true,
            "messageId": "msg_abc123"
          }
        }
      },
      {
        "job_key": "srv1-job-def456",
        "type": "inventory-checker",
        "status": "FAILED",
        "process_instance_id": "srv1-eF5gH9iJ2kL4mN7pR0",
        "element_id": "check-inventory",
        "element_instance_id": "srv1-elem-inventory01",
        "worker_id": "inventory-worker-03",
        "tenant_id": "production",
        "created_at": "2025-01-11T10:20:00.000Z",
        "activated_at": "2025-01-11T10:20:15.789Z",
        "failed_at": "2025-01-11T10:22:30.123Z",
        "deadline": "2025-01-11T10:22:15.789Z",
        "retries": 0,
        "original_retries": 3,
        "error_message": "Inventory service unavailable after 3 attempts",
        "failure_history": [
          {
            "attempt": 1,
            "failed_at": "2025-01-11T10:20:45.000Z",
            "error": "Connection timeout"
          },
          {
            "attempt": 2,
            "failed_at": "2025-01-11T10:21:30.000Z",
            "error": "Service temporarily unavailable"
          },
          {
            "attempt": 3,
            "failed_at": "2025-01-11T10:22:30.123Z",
            "error": "Max retries exceeded"
          }
        ],
        "incident_id": "srv1-incident-inventory-failure"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 1247,
      "total_pages": 63,
      "has_next": true,
      "has_prev": false
    },
    "summary": {
      "total_jobs": 1247,
      "active_jobs": 34,
      "completed_jobs": 1189,
      "failed_jobs": 23,
      "cancelled_jobs": 1,
      "success_rate_percent": 95.4,
      "avg_duration_seconds": 42
    },
    "filters_applied": {
      "status": null,
      "type": null,
      "worker": null,
      "period": "all"
    }
  },
  "request_id": "req_1641998402800"
}
```

## Поля ответа

### Job Object
- `job_key` (string): Уникальный ключ задания
- `type` (string): Тип задания
- `status` (string): Статус задания
- `process_instance_id` (string): ID экземпляра процесса
- `element_id` (string): ID элемента BPMN
- `worker_id` (string): ID назначенного worker'а
- `tenant_id` (string): ID тенанта

### Timing Information
- `created_at` (string): Время создания
- `activated_at` (string): Время активации
- `completed_at` (string): Время завершения (если применимо)
- `failed_at` (string): Время провала (если применимо)
- `deadline` (string): Крайний срок выполнения

### Job Configuration
- `retries` (integer): Оставшиеся попытки
- `variables` (object): Переменные для обработки
- `custom_headers` (object): Пользовательские заголовки
- `timeout_ms` (integer): Таймаут в миллисекундах

### Execution Information
- `elapsed_time_ms` (integer): Прошедшее время
- `remaining_time_ms` (integer): Оставшееся время
- `execution_summary` (object): Сводка выполнения (для завершенных)
- `failure_history` (array): История провалов (для провалившихся)

### Summary Statistics
- Общие счетчики заданий
- Процент успешности
- Средняя длительность выполнения

## Статусы заданий

### Возможные статусы
- `pending` - Задание создано и ожидает активации
- `activatable` - Задание готово к активации (синоним для `pending`)
- `activated` - Задание активировано и назначено worker'у (синоним для `running`)
- `running` - Задание выполняется worker'ом
- `completed` - Задание успешно завершено
- `failed` - Задание провалено (нет попыток)
- `cancelled` - Задание отменено

## Использование

### Job Monitoring Dashboard
```javascript
async function getJobDashboardData() {
  const [activeJobs, recentJobs, failedJobs] = await Promise.all([
    fetch('/api/v1/jobs?state=activatable').then(r => r.json()),
    fetch('/api/v1/jobs?period=1h').then(r => r.json()),
    fetch('/api/v1/jobs?state=failed&period=24h').then(r => r.json())
  ]);
  
  return {
    activeJobs: activeJobs.data,
    recentActivity: recentJobs.data.summary,
    failedJobs: failedJobs.data.jobs,
    systemHealth: {
      successRate: recentJobs.data.summary.success_rate_percent,
      avgDuration: recentJobs.data.summary.avg_duration_seconds,
      totalProcessed: recentJobs.data.summary.total_jobs
    }
  };
}
```

### Worker Performance Analysis
```javascript
async function analyzeWorkerPerformance() {
  const response = await fetch('/api/v1/jobs?period=24h&page_size=1000');
  const data = await response.json();
  
  const workerStats = {};
  
  data.data.jobs.forEach(job => {
    if (!job.worker_id) return;
    
    if (!workerStats[job.worker_id]) {
      workerStats[job.worker_id] = {
        workerId: job.worker_id,
        totalJobs: 0,
        completedJobs: 0,
        failedJobs: 0,
        totalDuration: 0,
        jobTypes: new Set()
      };
    }
    
    const stats = workerStats[job.worker_id];
    stats.totalJobs++;
    stats.jobTypes.add(job.type);
    
    if (job.status === 'COMPLETED') {
      stats.completedJobs++;
      if (job.execution_summary) {
        stats.totalDuration += job.execution_summary.total_duration_ms;
      }
    } else if (job.status === 'FAILED') {
      stats.failedJobs++;
    }
  });
  
  // Рассчитываем производительность
  return Object.values(workerStats).map(stats => ({
    ...stats,
    successRate: (stats.completedJobs / stats.totalJobs) * 100,
    avgDuration: stats.completedJobs > 0 ? stats.totalDuration / stats.completedJobs : 0,
    jobTypes: Array.from(stats.jobTypes)
  })).sort((a, b) => b.successRate - a.successRate);
}
```

### Job Type Analysis
```javascript
async function analyzeJobTypes() {
  const response = await fetch('/api/v1/jobs?period=7d&page_size=1000');
  const data = await response.json();
  
  const typeStats = {};
  
  data.data.jobs.forEach(job => {
    if (!typeStats[job.type]) {
      typeStats[job.type] = {
        type: job.type,
        total: 0,
        completed: 0,
        failed: 0,
        active: 0,
        avgDuration: 0,
        totalDuration: 0
      };
    }
    
    const stats = typeStats[job.type];
    stats.total++;
    
    switch (job.status) {
      case 'COMPLETED':
        stats.completed++;
        if (job.execution_summary) {
          stats.totalDuration += job.execution_summary.total_duration_ms;
        }
        break;
      case 'FAILED':
        stats.failed++;
        break;
      case 'ACTIVE':
        stats.active++;
        break;
    }
  });
  
  // Финализируем расчеты
  Object.values(typeStats).forEach(stats => {
    stats.successRate = (stats.completed / stats.total) * 100;
    stats.avgDuration = stats.completed > 0 ? stats.totalDuration / stats.completed : 0;
    delete stats.totalDuration; // Убираем промежуточное поле
  });
  
  return Object.values(typeStats).sort((a, b) => b.total - a.total);
}
```

### Failed Jobs Investigation
```javascript
async function investigateFailures() {
  const response = await fetch('/api/v1/jobs?state=failed&period=24h&page_size=100');
  const data = await response.json();
  
  const failureAnalysis = {
    totalFailures: data.data.jobs.length,
    errorPatterns: {},
    workerIssues: {},
    typeIssues: {},
    timeDistribution: {}
  };
  
  data.data.jobs.forEach(job => {
    // Анализ ошибок
    const errorType = categorizeError(job.error_message);
    failureAnalysis.errorPatterns[errorType] = (failureAnalysis.errorPatterns[errorType] || 0) + 1;
    
    // Анализ по worker'ам
    if (job.worker_id) {
      failureAnalysis.workerIssues[job.worker_id] = (failureAnalysis.workerIssues[job.worker_id] || 0) + 1;
    }
    
    // Анализ по типам заданий
    failureAnalysis.typeIssues[job.type] = (failureAnalysis.typeIssues[job.type] || 0) + 1;
    
    // Временное распределение
    const hour = new Date(job.failed_at).getHours();
    failureAnalysis.timeDistribution[hour] = (failureAnalysis.timeDistribution[hour] || 0) + 1;
  });
  
  return failureAnalysis;
}

function categorizeError(errorMessage) {
  if (!errorMessage) return 'UNKNOWN';
  
  const message = errorMessage.toLowerCase();
  
  if (message.includes('timeout')) return 'TIMEOUT';
  if (message.includes('connection')) return 'CONNECTION_ERROR';
  if (message.includes('rate limit')) return 'RATE_LIMITED';
  if (message.includes('unauthorized') || message.includes('forbidden')) return 'AUTH_ERROR';
  if (message.includes('not found')) return 'NOT_FOUND';
  if (message.includes('unavailable') || message.includes('service down')) return 'SERVICE_UNAVAILABLE';
  
  return 'OTHER';
}
```

### Real-time Job Monitoring
```javascript
class JobMonitor {
  constructor() {
    this.lastCheck = new Date();
    this.activeJobs = new Map();
  }
  
  async startMonitoring(interval = 5000) {
    setInterval(async () => {
      await this.checkJobChanges();
    }, interval);
  }
  
  async checkJobChanges() {
    try {
      const response = await fetch(`/api/v1/jobs?state=activatable&created_after=${this.lastCheck.toISOString()}`);
      const data = await response.json();
      
      const currentActiveJobs = new Map();
      
      data.data.jobs.forEach(job => {
        currentActiveJobs.set(job.job_key, job);
        
        // Проверяем новые задания
        if (!this.activeJobs.has(job.job_key)) {
          this.onNewJob(job);
        }
        
        // Проверяем задания близкие к таймауту
        if (job.remaining_time_ms < 30000) { // < 30 секунд
          this.onJobNearTimeout(job);
        }
      });
      
      // Проверяем завершенные задания
      this.activeJobs.forEach((job, jobKey) => {
        if (!currentActiveJobs.has(jobKey)) {
          this.onJobCompleted(job);
        }
      });
      
      this.activeJobs = currentActiveJobs;
      this.lastCheck = new Date();
      
    } catch (error) {
      console.error('Job monitoring error:', error);
    }
  }
  
  onNewJob(job) {
    console.log(`New active job: ${job.job_key} (${job.type})`);
  }
  
  onJobNearTimeout(job) {
    console.warn(`Job ${job.job_key} near timeout: ${job.remaining_time_ms}ms remaining`);
  }
  
  onJobCompleted(job) {
    console.log(`Job completed or removed: ${job.job_key}`);
  }
}

// Использование
const monitor = new JobMonitor();
monitor.startMonitoring(3000); // Проверка каждые 3 секунды
```

## Связанные endpoints
- [`POST /api/v1/jobs/activate`](./activate-jobs.md) - Активация заданий
- [`GET /api/v1/jobs/:key`](./get-job.md) - Детали конкретного задания
- [`GET /api/v1/jobs/stats`](./get-job-stats.md) - Статистика заданий
- [`PUT /api/v1/jobs/:key/complete`](./complete-job.md) - Завершить задание
