# PUT /api/v1/jobs/:key/cancel

## Описание
Отмена задания. Применимо к заданиям в статусе CREATED или ACTIVE.

## URL
```
PUT /api/v1/jobs/{job_key}/cancel
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры пути
- `job_key` (string): Ключ задания для отмены

## Примеры запросов

### cURL
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/cancel" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const jobKey = 'srv1-job-xyz789';
const response = await fetch(`/api/v1/jobs/${jobKey}/cancel`, {
  method: 'PUT',
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const result = await response.json();
```

## Ответы

### 200 OK - Задание отменено
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "status": "CANCELLED",
    "cancelled_at": "2025-01-11T10:35:30.789Z",
    "cancelled_by": "api-admin",
    "was_active": true,
    "worker_id": "payment-worker-02",
    "cancellation_reason": "Manual cancellation via API"
  },
  "request_id": "req_1641998403300"
}
```

### 409 Conflict - Задание нельзя отменить
```json
{
  "success": false,
  "error": {
    "code": "JOB_NOT_CANCELLABLE",
    "message": "Job cannot be cancelled in current state",
    "details": {
      "job_key": "srv1-job-xyz789",
      "current_status": "COMPLETED"
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/jobs/:key`](./get-job.md) - Проверить статус перед отменой
- [`GET /api/v1/jobs`](./list-jobs.md) - Список заданий для отмены
