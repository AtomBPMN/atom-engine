# GET /api/v1/jobs/stats

## Описание
Получение статистики по заданиям: производительность, распределение по статусам и типам, анализ worker'ов.

## URL
```
GET /api/v1/jobs/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры запроса (Query Parameters)
- `period` (string): Период для статистики (`1h`, `24h`, `7d`, `30d`, `all`)
- `type` (string): Статистика для конкретного типа заданий
- `worker` (string): Статистика для конкретного worker'а
- `tenant_id` (string): Статистика для конкретного тенанта

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/jobs/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за неделю
```bash
curl -X GET "http://localhost:27555/api/v1/jobs/stats?period=7d" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Статистика заданий
```json
{
  "success": true,
  "data": {
    "period": "all",
    "generated_at": "2025-01-11T10:32:15.456Z",
    "overview": {
      "total_jobs": 67834,
      "active_jobs": 34,
      "completed_jobs": 67800,
      "failed_jobs": 0,
      "cancelled_jobs": 0,
      "success_rate_percent": 100.0,
      "avg_duration_seconds": 12.5
    },
    "job_types": {
      "email-service": {
        "count": 25600,
        "success_rate": 99.8,
        "avg_duration_ms": 2500
      },
      "payment-processor": {
        "count": 18900,
        "success_rate": 98.5,
        "avg_duration_ms": 5000
      },
      "pdf-generator": {
        "count": 12400,
        "success_rate": 97.2,
        "avg_duration_ms": 8500
      }
    },
    "worker_performance": {
      "email-worker-01": {
        "jobs_processed": 8934,
        "success_rate": 99.9,
        "avg_processing_time_ms": 2400
      },
      "payment-worker-02": {
        "jobs_processed": 6745,
        "success_rate": 98.8,
        "avg_processing_time_ms": 4800
      }
    },
    "performance_metrics": {
      "jobs_per_second": 15.8,
      "peak_throughput": 45.2,
      "avg_queue_time_ms": 1200,
      "p95_processing_time_ms": 15000
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/jobs`](./list-jobs.md) - Детальный список заданий
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
