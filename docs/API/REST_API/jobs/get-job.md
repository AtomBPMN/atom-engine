# GET /api/v1/jobs/:key

## Описание
Получение детальной информации о конкретном задании, включая историю выполнения, переменные и метаданные.

## URL
```
GET /api/v1/jobs/{job_key}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры пути
- `job_key` (string): Ключ задания

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/jobs/srv1-job-xyz789" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const jobKey = 'srv1-job-xyz789';
const response = await fetch(`/api/v1/jobs/${jobKey}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const jobDetails = await response.json();
```

## Ответы

### 200 OK - Детали задания (Активное)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "type": "payment-processor",
    "status": "ACTIVE",
    "created_at": "2025-01-11T10:30:00.000Z",
    "activated_at": "2025-01-11T10:31:30.789Z",
    "deadline": "2025-01-11T10:33:30.789Z",
    "timeout_ms": 120000,
    "elapsed_time_ms": 90000,
    "remaining_time_ms": 30000,
    "process_context": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "process_id": "order-processing",
      "process_version": 3,
      "element_id": "process-payment",
      "element_type": "serviceTask",
      "element_name": "Process Payment",
      "element_instance_id": "srv1-elem-payment01",
      "token_id": "srv1-token-payment01"
    },
    "worker_assignment": {
      "worker_id": "payment-worker-02",
      "worker_type": "payment-processor",
      "assigned_at": "2025-01-11T10:31:30.789Z",
      "max_jobs_for_worker": 5,
      "current_jobs_for_worker": 3
    },
    "retry_configuration": {
      "retries": 3,
      "original_retries": 3,
      "retries_used": 0,
      "retry_backoff": "EXPONENTIAL",
      "max_retry_delay_ms": 60000
    },
    "variables": {
      "orderId": "ORD-12345",
      "customerId": "CUST-67890",
      "amount": 299.99,
      "paymentMethod": "credit",
      "currency": "USD",
      "billingAddress": {
        "street": "123 Main St",
        "city": "Anytown",
        "country": "US"
      }
    },
    "custom_headers": {
      "priority": "high",
      "payment_provider": "stripe",
      "correlation_id": "corr_abc123",
      "source_system": "web_checkout"
    },
    "tenant_id": "production",
    "execution_history": [
      {
        "timestamp": "2025-01-11T10:30:00.000Z",
        "action": "JOB_CREATED",
        "details": "Job created for payment processing"
      },
      {
        "timestamp": "2025-01-11T10:31:30.789Z",
        "action": "JOB_ACTIVATED",
        "worker_id": "payment-worker-02",
        "details": "Job activated by worker"
      }
    ],
    "performance_metrics": {
      "creation_to_activation_ms": 90789,
      "estimated_completion_time": "2025-01-11T10:32:30.000Z",
      "timeout_risk": "LOW"
    }
  },
  "request_id": "req_1641998402900"
}
```

### 200 OK - Детали задания (Завершенное)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-abc123",
    "type": "email-sender",
    "status": "COMPLETED",
    "created_at": "2025-01-11T10:25:00.000Z",
    "activated_at": "2025-01-11T10:25:30.123Z",
    "completed_at": "2025-01-11T10:26:15.456Z",
    "deadline": "2025-01-11T10:27:30.123Z",
    "timeout_ms": 120000,
    "process_context": {
      "process_instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
      "process_id": "order-processing",
      "element_id": "send-confirmation",
      "element_name": "Send Confirmation Email",
      "token_id": "srv1-token-email01"
    },
    "worker_assignment": {
      "worker_id": "email-worker-01",
      "worker_type": "email-sender",
      "assigned_at": "2025-01-11T10:25:30.123Z"
    },
    "retry_configuration": {
      "retries": 3,
      "original_retries": 3,
      "retries_used": 0
    },
    "variables": {
      "recipient": "customer@example.com",
      "subject": "Order Confirmation #ORD-12340",
      "body": "Thank you for your order...",
      "orderId": "ORD-12340",
      "template": "order-confirmation"
    },
    "custom_headers": {
      "priority": "normal",
      "template": "order-confirmation",
      "locale": "en-US"
    },
    "execution_summary": {
      "total_duration_ms": 45333,
      "activation_duration_ms": 30123,
      "processing_duration_ms": 42800,
      "system_overhead_ms": 2533,
      "completed_by": "email-worker-01",
      "result_variables": {
        "emailSent": true,
        "messageId": "msg_abc123def456",
        "sentAt": "2025-01-11T10:26:15.400Z",
        "provider": "sendgrid"
      }
    },
    "execution_history": [
      {
        "timestamp": "2025-01-11T10:25:00.000Z",
        "action": "JOB_CREATED",
        "details": "Job created for email sending"
      },
      {
        "timestamp": "2025-01-11T10:25:30.123Z",
        "action": "JOB_ACTIVATED",
        "worker_id": "email-worker-01",
        "details": "Job activated by worker"
      },
      {
        "timestamp": "2025-01-11T10:26:15.456Z",
        "action": "JOB_COMPLETED",
        "worker_id": "email-worker-01",
        "details": "Job completed successfully",
        "result_summary": {
          "variables_added": 4,
          "processing_time_ms": 42800
        }
      }
    ],
    "performance_metrics": {
      "creation_to_activation_ms": 30123,
      "activation_to_completion_ms": 45333,
      "completion_efficiency": 95.2
    }
  },
  "request_id": "req_1641998403000"
}
```

### 200 OK - Детали задания (Провалившееся)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-def456",
    "type": "inventory-checker",
    "status": "FAILED",
    "created_at": "2025-01-11T10:20:00.000Z",
    "activated_at": "2025-01-11T10:20:15.789Z",
    "failed_at": "2025-01-11T10:22:30.123Z",
    "deadline": "2025-01-11T10:22:15.789Z",
    "process_context": {
      "process_instance_id": "srv1-eF5gH9iJ2kL4mN7pR0",
      "element_id": "check-inventory",
      "element_name": "Check Inventory Availability"
    },
    "worker_assignment": {
      "worker_id": "inventory-worker-03",
      "assigned_at": "2025-01-11T10:20:15.789Z"
    },
    "retry_configuration": {
      "retries": 0,
      "original_retries": 3,
      "retries_used": 3,
      "retry_backoff": "EXPONENTIAL"
    },
    "variables": {
      "productId": "PROD-001",
      "quantity": 2,
      "warehouseId": "WH-MAIN"
    },
    "failure_information": {
      "error_message": "Inventory service unavailable after 3 attempts",
      "error_code": "SERVICE_UNAVAILABLE",
      "final_attempt": true,
      "total_attempts": 4,
      "failure_history": [
        {
          "attempt": 1,
          "failed_at": "2025-01-11T10:20:45.000Z",
          "error": "Connection timeout",
          "backoff_ms": 2000
        },
        {
          "attempt": 2,
          "failed_at": "2025-01-11T10:21:30.000Z",
          "error": "Service temporarily unavailable",
          "backoff_ms": 4000
        },
        {
          "attempt": 3,
          "failed_at": "2025-01-11T10:22:15.000Z",
          "error": "HTTP 503 Service Unavailable",
          "backoff_ms": 8000
        },
        {
          "attempt": 4,
          "failed_at": "2025-01-11T10:22:30.123Z",
          "error": "Max retries exceeded",
          "final_attempt": true
        }
      ]
    },
    "incident_information": {
      "incident_created": true,
      "incident_id": "srv1-incident-inventory-failure",
      "incident_type": "JOB_FAILURE",
      "created_at": "2025-01-11T10:22:30.124Z"
    },
    "execution_history": [
      {
        "timestamp": "2025-01-11T10:20:00.000Z",
        "action": "JOB_CREATED"
      },
      {
        "timestamp": "2025-01-11T10:20:15.789Z",
        "action": "JOB_ACTIVATED",
        "worker_id": "inventory-worker-03"
      },
      {
        "timestamp": "2025-01-11T10:20:45.000Z",
        "action": "JOB_FAILED_RETRY",
        "attempt": 1,
        "error": "Connection timeout"
      },
      {
        "timestamp": "2025-01-11T10:21:30.000Z",
        "action": "JOB_FAILED_RETRY",
        "attempt": 2,
        "error": "Service temporarily unavailable"
      },
      {
        "timestamp": "2025-01-11T10:22:15.000Z",
        "action": "JOB_FAILED_RETRY",
        "attempt": 3,
        "error": "HTTP 503 Service Unavailable"
      },
      {
        "timestamp": "2025-01-11T10:22:30.123Z",
        "action": "JOB_FAILED_FINAL",
        "error": "Max retries exceeded"
      }
    ]
  },
  "request_id": "req_1641998403001"
}
```

### 404 Not Found - Задание не найдено
```json
{
  "success": false,
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Job not found",
    "details": {
      "job_key": "srv1-job-nonexistent"
    }
  },
  "request_id": "req_1641998403002"
}
```

## Поля ответа

### Basic Information
- `job_key` (string): Уникальный ключ задания
- `type` (string): Тип задания
- `status` (string): Текущий статус
- `timeout_ms` (integer): Таймаут в миллисекундах

### Timing Information
- `created_at` (string): Время создания
- `activated_at` (string): Время активации
- `completed_at` (string): Время завершения (если применимо)
- `failed_at` (string): Время провала (если применимо)
- `deadline` (string): Крайний срок выполнения

### Process Context
- `process_instance_id` (string): ID экземпляра процесса
- `process_id` (string): ID процесса
- `element_id` (string): ID элемента BPMN
- `element_name` (string): Название элемента
- `token_id` (string): ID токена

### Worker Assignment
- `worker_id` (string): ID назначенного worker'а
- `worker_type` (string): Тип worker'а
- `assigned_at` (string): Время назначения

### Retry Configuration
- `retries` (integer): Оставшиеся попытки
- `original_retries` (integer): Изначальное количество попыток
- `retries_used` (integer): Использованные попытки

### Variables and Headers
- `variables` (object): Переменные задания
- `custom_headers` (object): Пользовательские заголовки

### Execution Summary (для завершенных)
- `total_duration_ms` (integer): Общее время выполнения
- `completed_by` (string): ID worker'а, завершившего задание
- `result_variables` (object): Переменные результата

### Failure Information (для провалившихся)
- `error_message` (string): Сообщение об ошибке
- `failure_history` (array): История всех попыток
- `incident_information` (object): Информация о созданном инциденте

## Использование

### Job Details Inspector
```javascript
async function inspectJob(jobKey) {
  const response = await fetch(`/api/v1/jobs/${jobKey}`);
  const data = await response.json();
  
  if (!data.success) {
    throw new Error(`Job not found: ${jobKey}`);
  }
  
  const job = data.data;
  
  const inspection = {
    basicInfo: {
      key: job.job_key,
      type: job.type,
      status: job.status,
      processInstance: job.process_context.process_instance_id
    },
    
    timing: {
      created: new Date(job.created_at),
      activated: job.activated_at ? new Date(job.activated_at) : null,
      completed: job.completed_at ? new Date(job.completed_at) : null,
      deadline: new Date(job.deadline),
      elapsed: job.elapsed_time_ms,
      remaining: job.remaining_time_ms
    },
    
    worker: {
      id: job.worker_assignment?.worker_id,
      type: job.worker_assignment?.worker_type,
      assignedAt: job.worker_assignment?.assigned_at
    },
    
    variables: job.variables,
    headers: job.custom_headers,
    
    performance: analyzeJobPerformance(job),
    issues: identifyJobIssues(job)
  };
  
  return inspection;
}

function analyzeJobPerformance(job) {
  const performance = {
    status: 'GOOD',
    metrics: {},
    warnings: []
  };
  
  if (job.status === 'ACTIVE') {
    // Анализ активного задания
    const timeoutRisk = (job.elapsed_time_ms / job.timeout_ms) * 100;
    
    performance.metrics.timeoutRisk = timeoutRisk;
    
    if (timeoutRisk > 80) {
      performance.status = 'CRITICAL';
      performance.warnings.push('Job near timeout');
    } else if (timeoutRisk > 60) {
      performance.status = 'WARNING';
      performance.warnings.push('Job running longer than expected');
    }
  }
  
  if (job.status === 'COMPLETED' && job.execution_summary) {
    // Анализ завершенного задания
    const duration = job.execution_summary.total_duration_ms;
    const efficiency = job.execution_summary.completion_efficiency || 0;
    
    performance.metrics.duration = duration;
    performance.metrics.efficiency = efficiency;
    
    if (duration > 60000) { // > 1 минуты
      performance.warnings.push('Long execution time');
    }
    
    if (efficiency < 80) {
      performance.warnings.push('Low completion efficiency');
    }
  }
  
  if (job.retry_configuration.retries_used > 0) {
    performance.warnings.push(`Used ${job.retry_configuration.retries_used} retries`);
  }
  
  return performance;
}

function identifyJobIssues(job) {
  const issues = [];
  
  // Проверка статуса
  if (job.status === 'FAILED') {
    issues.push({
      type: 'FAILURE',
      severity: 'HIGH',
      message: job.failure_information?.error_message || 'Job failed',
      attempts: job.failure_information?.total_attempts
    });
  }
  
  // Проверка таймаута
  if (job.status === 'ACTIVE' && job.remaining_time_ms < 30000) {
    issues.push({
      type: 'TIMEOUT_RISK',
      severity: 'MEDIUM',
      message: `Only ${Math.round(job.remaining_time_ms / 1000)}s remaining`,
      deadline: job.deadline
    });
  }
  
  // Проверка повторов
  if (job.retry_configuration.retries_used > 1) {
    issues.push({
      type: 'MULTIPLE_RETRIES',
      severity: 'LOW',
      message: `Job required ${job.retry_configuration.retries_used} retries`,
      reliability: 'QUESTIONABLE'
    });
  }
  
  return issues;
}
```

### Job Debugging Helper
```javascript
async function debugJob(jobKey) {
  const job = await inspectJob(jobKey);
  
  console.log(`=== Job Debug Report: ${jobKey} ===`);
  console.log(`Status: ${job.basicInfo.status}`);
  console.log(`Type: ${job.basicInfo.type}`);
  console.log(`Process: ${job.basicInfo.processInstance}`);
  
  if (job.worker.id) {
    console.log(`Worker: ${job.worker.id} (${job.worker.type})`);
  }
  
  console.log('\n--- Timing ---');
  console.log(`Created: ${job.timing.created.toISOString()}`);
  if (job.timing.activated) {
    console.log(`Activated: ${job.timing.activated.toISOString()}`);
    const activationDelay = job.timing.activated.getTime() - job.timing.created.getTime();
    console.log(`Activation delay: ${activationDelay}ms`);
  }
  
  if (job.timing.remaining !== undefined) {
    console.log(`Remaining time: ${Math.round(job.timing.remaining / 1000)}s`);
  }
  
  console.log('\n--- Variables ---');
  console.log(JSON.stringify(job.variables, null, 2));
  
  if (job.headers && Object.keys(job.headers).length > 0) {
    console.log('\n--- Custom Headers ---');
    console.log(JSON.stringify(job.headers, null, 2));
  }
  
  console.log('\n--- Performance ---');
  console.log(`Status: ${job.performance.status}`);
  if (job.performance.warnings.length > 0) {
    console.log('Warnings:');
    job.performance.warnings.forEach(warning => console.log(`  - ${warning}`));
  }
  
  if (job.issues.length > 0) {
    console.log('\n--- Issues ---');
    job.issues.forEach(issue => {
      console.log(`[${issue.severity}] ${issue.type}: ${issue.message}`);
    });
  }
  
  return job;
}
```

### Job Comparison
```javascript
async function compareJobs(jobKeys) {
  const jobs = await Promise.all(
    jobKeys.map(key => inspectJob(key).catch(err => ({ error: err.message, key })))
  );
  
  const comparison = {
    successful: jobs.filter(j => !j.error),
    failed: jobs.filter(j => j.error),
    summary: {
      totalJobs: jobs.length,
      avgDuration: 0,
      statusDistribution: {},
      workerDistribution: {},
      typeDistribution: {}
    }
  };
  
  comparison.successful.forEach(job => {
    // Статистика по статусам
    const status = job.basicInfo.status;
    comparison.summary.statusDistribution[status] = 
      (comparison.summary.statusDistribution[status] || 0) + 1;
    
    // Статистика по workers
    const worker = job.worker.type || 'unknown';
    comparison.summary.workerDistribution[worker] = 
      (comparison.summary.workerDistribution[worker] || 0) + 1;
    
    // Статистика по типам
    const type = job.basicInfo.type;
    comparison.summary.typeDistribution[type] = 
      (comparison.summary.typeDistribution[type] || 0) + 1;
    
    // Длительность
    if (job.timing.elapsed) {
      comparison.summary.avgDuration += job.timing.elapsed;
    }
  });
  
  if (comparison.successful.length > 0) {
    comparison.summary.avgDuration = 
      comparison.summary.avgDuration / comparison.successful.length;
  }
  
  return comparison;
}
```

### Job History Tracker
```javascript
class JobHistoryTracker {
  constructor() {
    this.jobStates = new Map();
  }
  
  async trackJob(jobKey) {
    const currentState = await this.getJobState(jobKey);
    const previousState = this.jobStates.get(jobKey);
    
    if (previousState && currentState.status !== previousState.status) {
      this.onJobStatusChanged(jobKey, previousState.status, currentState.status);
    }
    
    this.jobStates.set(jobKey, currentState);
    
    return currentState;
  }
  
  async getJobState(jobKey) {
    try {
      const response = await fetch(`/api/v1/jobs/${jobKey}`);
      const data = await response.json();
      
      if (data.success) {
        return {
          status: data.data.status,
          lastUpdate: new Date(),
          worker: data.data.worker_assignment?.worker_id,
          remaining: data.data.remaining_time_ms
        };
      }
    } catch (error) {
      return { status: 'UNKNOWN', error: error.message };
    }
  }
  
  onJobStatusChanged(jobKey, oldStatus, newStatus) {
    console.log(`Job ${jobKey}: ${oldStatus} → ${newStatus}`);
    
    if (newStatus === 'COMPLETED') {
      console.log(`✅ Job ${jobKey} completed successfully`);
    } else if (newStatus === 'FAILED') {
      console.log(`❌ Job ${jobKey} failed`);
    }
  }
  
  getJobHistory(jobKey) {
    return this.jobStates.get(jobKey);
  }
}
```

## Связанные endpoints
- [`GET /api/v1/jobs`](./list-jobs.md) - Список заданий
- [`PUT /api/v1/jobs/:key/complete`](./complete-job.md) - Завершить задание
- [`PUT /api/v1/jobs/:key/fail`](./fail-job.md) - Провалить задание
- [`POST /api/v1/jobs/activate`](./activate-jobs.md) - Активировать задания
