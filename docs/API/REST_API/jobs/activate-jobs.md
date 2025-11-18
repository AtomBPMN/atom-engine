# POST /api/v1/jobs/activate

## Описание
Активация заданий для worker. Worker "забирает" задания определенного типа для выполнения. Используется для реализации job polling pattern.

## URL
```
POST /api/v1/jobs/activate
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

```http
X-API-Key: your-api-key-here
```

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `type` (string): Тип заданий для активации
- `worker` (string): Идентификатор worker

### Опциональные поля
- `max_jobs` (integer): Максимальное количество заданий (по умолчанию: 10, максимум: 100)
- `timeout` (integer): Таймаут в миллисекундах (по умолчанию: 300000 = 5 минут)
- `fetch_variables` (array): Список переменных для получения (пустой = все переменные)

### Пример тела запроса
```json
{
  "type": "email-service",
  "worker": "email-worker-01",
  "max_jobs": 5,
  "timeout": 60000,
  "fetch_variables": ["recipient", "subject", "body", "attachments"]
}
```

## Примеры запросов

### Базовая активация
```bash
curl -X POST http://localhost:27555/api/v1/jobs/activate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "type": "email-service",
    "worker": "email-worker-01"
  }'
```

### Активация с ограниченными переменными
```bash
curl -X POST http://localhost:27555/api/v1/jobs/activate \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "type": "payment-processor",
    "worker": "payment-worker-02",
    "max_jobs": 3,
    "timeout": 120000,
    "fetch_variables": ["amount", "currency", "paymentMethod"]
  }'
```

### JavaScript
```javascript
const response = await fetch('/api/v1/jobs/activate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify({
    type: 'email-service',
    worker: 'email-worker-01',
    max_jobs: 5,
    timeout: 60000
  })
});

const jobs = await response.json();
```

### Go
```go
requestData := map[string]interface{}{
    "type":     "email-service",
    "worker":   "email-worker-01",
    "max_jobs": 5,
    "timeout":  60000,
}

jsonData, _ := json.Marshal(requestData)
req, _ := http.NewRequest("POST", "/api/v1/jobs/activate", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Задания успешно активированы
```json
{
  "success": true,
  "data": {
    "jobs": [
      {
        "key": "srv1-job-aB3dEf9hK2mN5pQ8",
        "type": "email-service",
        "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "element_id": "send-confirmation-email",
        "element_instance_id": "srv1-elem-aB3dEf9h",
        "worker": "email-worker-01",
        "retries": 3,
        "deadline": "2025-01-11T10:36:00.000Z",
        "variables": {
          "recipient": "customer@example.com",
          "subject": "Order Confirmation #12345",
          "body": "Thank you for your order...",
          "orderId": "ORD-12345",
          "customerName": "John Doe"
        },
        "custom_headers": {
          "priority": "high",
          "template": "order-confirmation",
          "locale": "en-US"
        },
        "created_at": "2025-01-11T10:30:00.123Z",
        "activated_at": "2025-01-11T10:31:00.000Z"
      },
      {
        "key": "srv1-job-cD4eF8gH1jK3mN6p",
        "type": "email-service", 
        "process_instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
        "element_id": "send-welcome-email",
        "element_instance_id": "srv1-elem-cD4eF8gH",
        "worker": "email-worker-01",
        "retries": 3,
        "deadline": "2025-01-11T10:36:00.000Z",
        "variables": {
          "recipient": "newuser@example.com",
          "subject": "Welcome to our service!",
          "body": "Welcome to our platform...",
          "userId": "USER-67890",
          "userName": "Jane Smith"
        },
        "custom_headers": {
          "priority": "normal",
          "template": "welcome-email",
          "locale": "en-US"
        },
        "created_at": "2025-01-11T10:29:30.456Z",
        "activated_at": "2025-01-11T10:31:00.000Z"
      }
    ],
    "worker": "email-worker-01",
    "activated_count": 2,
    "timeout": 60000
  },
  "request_id": "req_1641998400400"
}
```

### 200 OK - Нет доступных заданий
```json
{
  "success": true,
  "data": {
    "jobs": [],
    "worker": "email-worker-01",
    "activated_count": 0,
    "timeout": 60000
  },
  "request_id": "req_1641998400401"
}
```

### 400 Bad Request - Неверные параметры
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field_errors": {
        "type": "Job type is required",
        "worker": "Worker ID is required",
        "max_jobs": "Max jobs must be between 1 and 100",
        "timeout": "Timeout must be between 1000 and 3600000 milliseconds"
      }
    }
  },
  "request_id": "req_1641998400402"
}
```

## Поля ответа (Job Object)

### Основная информация
- `key` (string): Уникальный ключ задания
- `type` (string): Тип задания
- `worker` (string): ID назначенного worker
- `retries` (integer): Оставшееся количество попыток

### Контекст процесса
- `process_instance_id` (string): ID экземпляра процесса
- `element_id` (string): ID элемента BPMN
- `element_instance_id` (string): ID экземпляра элемента

### Данные и конфигурация
- `variables` (object): Переменные для обработки
- `custom_headers` (object): Пользовательские заголовки
- `deadline` (string): Крайний срок выполнения (ISO 8601 UTC)

### Временные метки
- `created_at` (string): Время создания задания
- `activated_at` (string): Время активации

## Worker Lifecycle

### 1. Активация заданий
```javascript
const jobs = await activateJobs({
  type: 'email-service',
  worker: 'email-worker-01',
  max_jobs: 10
});
```

### 2. Обработка заданий
```javascript
for (const job of jobs.data.jobs) {
  try {
    await processJob(job);
    await completeJob(job.key, { status: 'sent' });
  } catch (error) {
    await failJob(job.key, 2, error.message);
  }
}
```

### 3. Polling loop
```javascript
async function workerLoop() {
  while (true) {
    try {
      const jobs = await activateJobs({
        type: 'email-service',
        worker: 'email-worker-01',
        timeout: 30000
      });
      
      if (jobs.data.jobs.length > 0) {
        await processJobs(jobs.data.jobs);
      } else {
        await sleep(5000); // Пауза если нет заданий
      }
    } catch (error) {
      console.error('Worker error:', error);
      await sleep(10000); // Пауза при ошибке
    }
  }
}
```

## Валидация параметров

### type
- Формат: 1-100 символов, буквы, цифры, дефисы, подчеркивания
- Примеры: `email-service`, `payment_processor`, `pdf-generator`

### worker
- Формат: 1-100 символов, буквы, цифры, дефисы, подчеркивания  
- Рекомендуется: включать hostname/pod ID для уникальности
- Примеры: `email-worker-pod-123`, `payment-processor-node-01`

### max_jobs
- Минимум: 1
- Максимум: 100
- По умолчанию: 10

### timeout
- Минимум: 1000ms (1 секунда)
- Максимум: 3600000ms (1 час)
- По умолчанию: 300000ms (5 минут)

## Long Polling

### Поведение
- Если заданий нет, запрос "висит" до timeout или появления заданий
- Позволяет снизить нагрузку на систему по сравнению с частым polling
- Worker должен обрабатывать timeout как нормальное поведение

### Пример long polling worker
```javascript
async function longPollingWorker() {
  while (true) {
    try {
      const response = await fetch('/api/v1/jobs/activate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'your-api-key-here'
        },
        body: JSON.stringify({
          type: 'email-service',
          worker: 'email-worker-01',
          timeout: 30000  // 30 секунд long polling
        })
      });
      
      const result = await response.json();
      
      if (result.data.jobs.length > 0) {
        await processJobs(result.data.jobs);
      }
      // Сразу запрашиваем следующую порцию
    } catch (error) {
      console.error('Polling error:', error);
      await sleep(5000); // Пауза при ошибке
    }
  }
}
```

## Производительность

### Рекомендации
- **Batch размер**: 5-20 заданий для баланса latency/throughput
- **Timeout**: 30-60 секунд для long polling
- **Retry**: Экспоненциальный backoff при ошибках
- **Мониторинг**: Отслеживайте queue depth и worker utilization

### Метрики
- Время активации заданий
- Количество активных workers
- Queue depth по типам заданий
- Worker throughput

## Связанные endpoints
- [`PUT /api/v1/jobs/:key/complete`](./complete-job.md) - Завершить задание
- [`PUT /api/v1/jobs/:key/fail`](./fail-job.md) - Провалить задание
- [`POST /api/v1/jobs/:key/throw-error`](./throw-error.md) - Выбросить BPMN ошибку
- [`GET /api/v1/jobs`](./list-jobs.md) - Список заданий
- [`GET /api/v1/jobs/stats`](./get-job-stats.md) - Статистика заданий
