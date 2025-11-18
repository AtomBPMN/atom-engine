# POST /api/v1/jobs

## Описание
Создание нового задания в системе. Используется для создания заданий вне контекста выполнения BPMN процессов.

## URL
```
POST /api/v1/jobs
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `type` (string): Тип задания
- `variables` (object): Переменные для обработки

### Опциональные поля
- `retries` (integer): Количество попыток (по умолчанию: 3)
- `timeout` (string): Таймаут в формате ISO 8601 (по умолчанию: "PT2M")
- `custom_headers` (object): Пользовательские заголовки
- `tenant_id` (string): ID тенанта (по умолчанию: "default")

## Примеры запросов

### Простое задание
```bash
curl -X POST "http://localhost:27555/api/v1/jobs" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "type": "email-sender",
    "variables": {
      "recipient": "user@example.com",
      "subject": "Welcome",
      "body": "Welcome to our service!"
    }
  }'
```

### Задание с настройками
```bash
curl -X POST "http://localhost:27555/api/v1/jobs" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "type": "data-processor",
    "variables": {
      "inputFile": "/data/input.csv",
      "outputFormat": "json"
    },
    "retries": 5,
    "timeout": "PT10M",
    "custom_headers": {
      "priority": "high",
      "department": "analytics"
    }
  }'
```

## Ответы

### 201 Created - Задание создано
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-manual123",
    "type": "email-sender",
    "status": "CREATED",
    "created_at": "2025-01-11T10:30:00.000Z",
    "retries": 3,
    "timeout_ms": 120000,
    "deadline": "2025-01-11T10:32:00.000Z",
    "variables": {
      "recipient": "user@example.com",
      "subject": "Welcome",
      "body": "Welcome to our service!"
    },
    "custom_headers": {},
    "tenant_id": "default",
    "manual_job": true,
    "available_for_activation": true
  },
  "request_id": "req_1641998403200"
}
```

## Связанные endpoints
- [`POST /api/v1/jobs/activate`](./activate-jobs.md) - Активация созданных заданий
- [`GET /api/v1/jobs/:key`](./get-job.md) - Статус созданного задания
