# GET /api/v1/timers

## Описание
Получение списка таймеров с фильтрацией по статусу, типу и метаданным.

## URL
```
GET /api/v1/timers
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer`

## Параметры запроса (Query Parameters)

### Фильтрация
- `status` (string): Фильтр по статусу (`SCHEDULED`, `ACTIVE`, `COMPLETED`, `CANCELLED`)
- `type` (string): Фильтр по типу (`DURATION`, `CYCLE`)
- `tenant_id` (string): Фильтр по тенанту
- `metadata_key` (string): Фильтр по ключу метаданных
- `metadata_value` (string): Фильтр по значению метаданных

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)
- `sort_by` (string): Поле сортировки (`created_at`, `scheduled_at`, `remaining_time`)
- `sort_order` (string): Порядок сортировки (`ASC`, `DESC`)

## Примеры запросов

### Все таймеры
```bash
curl -X GET "http://localhost:27555/api/v1/timers" \
  -H "X-API-Key: your-api-key-here"
```

### Активные таймеры
```bash
curl -X GET "http://localhost:27555/api/v1/timers?status=SCHEDULED" \
  -H "X-API-Key: your-api-key-here"
```

### Циклические таймеры
```bash
curl -X GET "http://localhost:27555/api/v1/timers?type=CYCLE&page_size=50" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Список таймеров
```json
{
  "success": true,
  "data": {
    "timers": [
      {
        "timer_id": "payment-timeout-ORD-12345",
        "duration_or_cycle": "PT5M",
        "status": "SCHEDULED",
        "timer_type": "DURATION",
        "created_at": "2025-01-11T10:30:00.000Z",
        "scheduled_at": "2025-01-11T10:35:00.000Z",
        "remaining_time_ms": 240000,
        "tenant_id": "production",
        "metadata": {
          "order_id": "ORD-12345",
          "timeout_type": "payment"
        }
      },
      {
        "timer_id": "hourly-cleanup",
        "duration_or_cycle": "R/PT1H",
        "status": "SCHEDULED",
        "timer_type": "CYCLE",
        "created_at": "2025-01-11T10:00:00.000Z",
        "scheduled_at": "2025-01-11T11:00:00.000Z",
        "remaining_time_ms": 1800000,
        "tenant_id": "default",
        "cycle_info": {
          "current_iteration": 5,
          "total_iterations": "infinite"
        },
        "metadata": {
          "task_type": "cleanup"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 89,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    },
    "summary": {
      "total_timers": 89,
      "scheduled_timers": 76,
      "completed_timers": 8,
      "cancelled_timers": 5,
      "duration_timers": 67,
      "cycle_timers": 22
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/timers`](./create-timer.md) - Создать таймер
- [`GET /api/v1/timers/:id`](./get-timer.md) - Детали таймера
