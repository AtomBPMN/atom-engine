# PUT /api/v1/jobs/:key/fail

## Описание
Провал задания с указанием ошибки и количества оставшихся попыток. Используется worker'ами при невозможности выполнить задание.

## URL
```
PUT /api/v1/jobs/{job_key}/fail
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры пути
- `job_key` (string): Ключ задания для провала

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `retries` (integer): Количество оставшихся попыток

### Опциональные поля
- `error_message` (string): Описание ошибки
- `backoff_duration` (string): Время задержки перед повтором (ISO 8601, по умолчанию: экспоненциальная задержка)

### Пример тела запроса
```json
{
  "retries": 2,
  "error_message": "Payment gateway timeout after 30 seconds",
  "backoff_duration": "PT1M"
}
```

## Примеры запросов

### Базовый провал
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/fail" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "retries": 2,
    "error_message": "Connection timeout to external service"
  }'
```

### Провал с задержкой
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/fail" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "retries": 1,
    "error_message": "Rate limit exceeded",
    "backoff_duration": "PT5M"
  }'
```

### Финальный провал
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/fail" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "retries": 0,
    "error_message": "Invalid payment method - cannot retry"
  }'
```

### JavaScript
```javascript
const jobKey = 'srv1-job-xyz789';
const failure = {
  retries: 2,
  error_message: 'Temporary service unavailable',
  backoff_duration: 'PT2M'
};

const response = await fetch(`/api/v1/jobs/${jobKey}/fail`, {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(failure)
});

const result = await response.json();
```

## Ответы

### 200 OK - Задание провалено (с повторами)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "status": "FAILED_RETRYABLE",
    "failed_at": "2025-01-11T10:33:45.123Z",
    "failed_by": "payment-worker-02",
    "error_message": "Payment gateway timeout after 30 seconds",
    "retries_remaining": 2,
    "original_retries": 3,
    "attempt_number": 2,
    "retry_strategy": {
      "backoff_duration": "PT1M",
      "backoff_type": "FIXED",
      "next_retry_at": "2025-01-11T10:34:45.123Z"
    },
    "failure_history": [
      {
        "attempt": 1,
        "failed_at": "2025-01-11T10:32:30.456Z",
        "error": "Connection refused",
        "worker_id": "payment-worker-02"
      },
      {
        "attempt": 2,
        "failed_at": "2025-01-11T10:33:45.123Z",
        "error": "Payment gateway timeout after 30 seconds",
        "worker_id": "payment-worker-02"
      }
    ],
    "process_info": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "process-payment",
      "will_retry": true
    }
  },
  "request_id": "req_1641998402700"
}
```

### 200 OK - Задание окончательно провалено
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "status": "FAILED",
    "failed_at": "2025-01-11T10:35:15.789Z",
    "failed_by": "payment-worker-02",
    "error_message": "Invalid payment method - cannot retry",
    "retries_remaining": 0,
    "original_retries": 3,
    "attempt_number": 4,
    "total_attempts": 4,
    "failure_history": [
      {
        "attempt": 1,
        "failed_at": "2025-01-11T10:32:30.456Z",
        "error": "Connection refused"
      },
      {
        "attempt": 2,
        "failed_at": "2025-01-11T10:33:45.123Z", 
        "error": "Payment gateway timeout"
      },
      {
        "attempt": 3,
        "failed_at": "2025-01-11T10:34:50.234Z",
        "error": "Service temporarily unavailable"
      },
      {
        "attempt": 4,
        "failed_at": "2025-01-11T10:35:15.789Z",
        "error": "Invalid payment method - cannot retry"
      }
    ],
    "incident_created": {
      "incident_id": "srv1-incident-job-failure",
      "type": "JOB_FAILURE",
      "status": "OPEN",
      "created_at": "2025-01-11T10:35:15.790Z"
    },
    "process_info": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "process-payment",
      "will_retry": false,
      "process_blocked": true
    }
  },
  "request_id": "req_1641998402701"
}
```

### 404 Not Found - Задание не найдено
```json
{
  "success": false,
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Job not found or not accessible",
    "details": {
      "job_key": "srv1-job-nonexistent"
    }
  },
  "request_id": "req_1641998402702"
}
```

### 400 Bad Request - Неверные параметры
```json
{
  "success": false,
  "error": {
    "code": "INVALID_RETRY_COUNT",
    "message": "Invalid retry count",
    "details": {
      "provided_retries": -1,
      "valid_range": "0 to current_retries",
      "current_retries": 2
    }
  },
  "request_id": "req_1641998402703"
}
```

## Поля ответа

### Job Failure Information
- `job_key` (string): Ключ провалившегося задания
- `status` (string): Новый статус (`FAILED_RETRYABLE`, `FAILED`)
- `failed_at` (string): Время провала
- `failed_by` (string): ID worker'а
- `error_message` (string): Сообщение об ошибке

### Retry Information
- `retries_remaining` (integer): Оставшиеся попытки
- `original_retries` (integer): Изначальное количество попыток
- `attempt_number` (integer): Номер текущей попытки
- `total_attempts` (integer): Общее количество попыток (для завершенных)

### Retry Strategy
- `backoff_duration` (string): Время задержки перед повтором
- `backoff_type` (string): Тип задержки (`FIXED`, `EXPONENTIAL`)
- `next_retry_at` (string): Время следующей попытки

### Failure History
- `failure_history` (array): История всех провалов
- Каждая запись содержит: attempt, failed_at, error, worker_id

### Incident Information (для окончательных провалов)
- `incident_created` (object): Созданный инцидент
- `incident_id`, `type`, `status`, `created_at`

## Стратегии повторов

### Fixed Backoff
```javascript
// Фиксированная задержка
{
  "retries": 2,
  "backoff_duration": "PT30S"  // Всегда 30 секунд
}
```

### Exponential Backoff (по умолчанию)
```javascript
// Экспоненциальная задержка (автоматически рассчитывается)
{
  "retries": 3,
  "error_message": "Temporary failure"
  // backoff_duration не указан - система рассчитает автоматически
}

// Результат:
// 1-я попытка: немедленно
// 2-я попытка: через 2^1 = 2 секунды  
// 3-я попытка: через 2^2 = 4 секунды
// 4-я попытка: через 2^3 = 8 секунд
```

## Использование

### Worker Error Handling
```javascript
class RobustWorker {
  async processJob(job) {
    try {
      const result = await this.executeJobLogic(job);
      return await this.completeJob(job.key, result);
      
    } catch (error) {
      return await this.handleJobFailure(job, error);
    }
  }
  
  async handleJobFailure(job, error) {
    const errorType = this.classifyError(error);
    
    switch (errorType) {
      case 'RETRYABLE':
        return await this.failJobWithRetry(job.key, error.message, job.retries - 1);
        
      case 'RATE_LIMITED':
        return await this.failJobWithBackoff(job.key, error.message, job.retries - 1, 'PT5M');
        
      case 'PERMANENT':
        return await this.failJobPermanently(job.key, error.message);
        
      default:
        return await this.failJobWithRetry(job.key, error.message, Math.max(0, job.retries - 1));
    }
  }
  
  classifyError(error) {
    if (error.code === 'RATE_LIMITED') return 'RATE_LIMITED';
    if (error.code === 'INVALID_INPUT') return 'PERMANENT';
    if (error.code === 'UNAUTHORIZED') return 'PERMANENT';
    if (error.code === 'TIMEOUT') return 'RETRYABLE';
    if (error.code === 'CONNECTION_ERROR') return 'RETRYABLE';
    
    return 'RETRYABLE'; // По умолчанию пытаемся повторить
  }
  
  async failJobWithRetry(jobKey, errorMessage, retries) {
    return await fetch(`/api/v1/jobs/${jobKey}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        retries: Math.max(0, retries),
        error_message: errorMessage
      })
    });
  }
  
  async failJobWithBackoff(jobKey, errorMessage, retries, backoffDuration) {
    return await fetch(`/api/v1/jobs/${jobKey}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        retries: Math.max(0, retries),
        error_message: errorMessage,
        backoff_duration: backoffDuration
      })
    });
  }
  
  async failJobPermanently(jobKey, errorMessage) {
    return await fetch(`/api/v1/jobs/${jobKey}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        retries: 0,
        error_message: `PERMANENT FAILURE: ${errorMessage}`
      })
    });
  }
}
```

### Retry Strategy Helper
```javascript
class RetryStrategyHelper {
  static calculateBackoff(attempt, baseDelay = 1000, maxDelay = 300000) {
    // Экспоненциальная задержка с jitter
    const exponentialDelay = Math.min(baseDelay * Math.pow(2, attempt - 1), maxDelay);
    const jitter = Math.random() * 0.1 * exponentialDelay; // ±10% jitter
    const totalDelay = exponentialDelay + jitter;
    
    return this.millisecondsToISO8601(totalDelay);
  }
  
  static millisecondsToISO8601(ms) {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `PT${hours}H${minutes % 60}M${seconds % 60}S`;
    } else if (minutes > 0) {
      return `PT${minutes}M${seconds % 60}S`;
    } else {
      return `PT${seconds}S`;
    }
  }
  
  static getRetryStrategy(errorType, attempt) {
    const strategies = {
      NETWORK_ERROR: {
        maxRetries: 5,
        baseDelay: 1000,
        maxDelay: 60000
      },
      RATE_LIMITED: {
        maxRetries: 3,
        baseDelay: 60000, // Начинаем с 1 минуты
        maxDelay: 300000  // Максимум 5 минут
      },
      SERVICE_UNAVAILABLE: {
        maxRetries: 4,
        baseDelay: 5000,
        maxDelay: 120000
      },
      TIMEOUT: {
        maxRetries: 3,
        baseDelay: 2000,
        maxDelay: 30000
      }
    };
    
    const strategy = strategies[errorType] || strategies.NETWORK_ERROR;
    
    return {
      shouldRetry: attempt <= strategy.maxRetries,
      backoffDuration: this.calculateBackoff(attempt, strategy.baseDelay, strategy.maxDelay),
      retriesRemaining: Math.max(0, strategy.maxRetries - attempt)
    };
  }
}

// Использование
async function handleJobError(job, error) {
  const errorType = classifyErrorType(error);
  const strategy = RetryStrategyHelper.getRetryStrategy(errorType, job.attempt || 1);
  
  if (strategy.shouldRetry) {
    return await fetch(`/api/v1/jobs/${job.key}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
      },
      body: JSON.stringify({
        retries: strategy.retriesRemaining,
        error_message: `${errorType}: ${error.message}`,
        backoff_duration: strategy.backoffDuration
      })
    });
  } else {
    // Окончательный провал
    return await fetch(`/api/v1/jobs/${job.key}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
      },
      body: JSON.stringify({
        retries: 0,
        error_message: `MAX_RETRIES_EXCEEDED: ${error.message}`
      })
    });
  }
}
```

### Job Failure Monitoring
```javascript
async function monitorJobFailures() {
  const failures = await fetch('/api/v1/jobs?status=FAILED&period=1h');
  const data = await failures.json();
  
  const analysis = {
    totalFailures: data.data.jobs.length,
    failuresByType: {},
    failuresByWorker: {},
    averageAttempts: 0
  };
  
  data.data.jobs.forEach(job => {
    // Группировка по типам ошибок
    const errorType = extractErrorType(job.error_message);
    analysis.failuresByType[errorType] = (analysis.failuresByType[errorType] || 0) + 1;
    
    // Группировка по worker
    const workerId = job.failed_by;
    analysis.failuresByWorker[workerId] = (analysis.failuresByWorker[workerId] || 0) + 1;
    
    // Подсчет попыток
    analysis.averageAttempts += job.total_attempts || 1;
  });
  
  analysis.averageAttempts = analysis.averageAttempts / data.data.jobs.length;
  
  return analysis;
}

function extractErrorType(errorMessage) {
  if (errorMessage.includes('timeout')) return 'TIMEOUT';
  if (errorMessage.includes('connection')) return 'CONNECTION';
  if (errorMessage.includes('rate limit')) return 'RATE_LIMITED';
  if (errorMessage.includes('unauthorized')) return 'AUTH_ERROR';
  return 'OTHER';
}
```

## Связанные endpoints
- [`POST /api/v1/jobs/activate`](./activate-jobs.md) - Активация заданий
- [`PUT /api/v1/jobs/:key/complete`](./complete-job.md) - Завершить задание
- [`GET /api/v1/jobs/:key`](./get-job.md) - Детали задания
- [`GET /api/v1/incidents`](../incidents/list-incidents.md) - Инциденты от провалов заданий
