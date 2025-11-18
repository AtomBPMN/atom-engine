# GET /api/v1/timers/stats

## Описание
Получение статистики по системе таймеров: производительность, точность, распределение по типам и состояниям.

## URL
```
GET /api/v1/timers/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer`

## Параметры запроса (Query Parameters)
- `period` (string): Период для статистики (`1h`, `24h`, `7d`, `30d`, `all`)
- `tenant_id` (string): Статистика для конкретного тенанта

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/timers/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за неделю
```bash
curl -X GET "http://localhost:27555/api/v1/timers/stats?period=7d" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Статистика таймеров
```json
{
  "success": true,
  "data": {
    "period": "all",
    "generated_at": "2025-01-11T10:32:15.456Z",
    "overview": {
      "total_timers": 5632,
      "active_timers": 89,
      "completed_timers": 5543,
      "cancelled_timers": 0,
      "duration_timers": 4567,
      "cycle_timers": 1065
    },
    "performance_metrics": {
      "avg_accuracy_ms": 12,
      "p95_accuracy_ms": 25,
      "p99_accuracy_ms": 50,
      "timers_per_second": 8.7,
      "total_firing_time_ms": 125000
    },
    "timewheel_statistics": {
      "levels_usage": [
        {"level": 0, "name": "seconds", "timers": 45},
        {"level": 1, "name": "minutes", "timers": 23},
        {"level": 2, "name": "hours", "timers": 15},
        {"level": 3, "name": "days", "timers": 5},
        {"level": 4, "name": "years", "timers": 1}
      ],
      "total_slots": 32768,
      "used_slots": 89,
      "utilization_percent": 0.27
    },
    "duration_distribution": {
      "under_1m": 1245,
      "1m_to_5m": 2890,
      "5m_to_30m": 987,
      "30m_to_1h": 345,
      "1h_to_24h": 123,
      "over_24h": 42
    },
    "cycle_analysis": {
      "infinite_cycles": 234,
      "limited_cycles": 831,
      "avg_iterations_per_cycle": 5.2,
      "most_common_intervals": {
        "PT1H": 156,
        "PT30M": 89,
        "PT5M": 67
      }
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/timers`](./list-timers.md) - Список таймеров
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
