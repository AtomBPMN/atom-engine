# GET /api/v1/messages/stats

## Описание
Получение статистики по системе сообщений: корреляция, буферизация, производительность.

## URL
```
GET /api/v1/messages/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Параметры запроса (Query Parameters)
- `period` (string): Период для статистики (`1h`, `24h`, `7d`, `30d`, `all`)
- `tenant_id` (string): Статистика для конкретного тенанта

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/messages/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за день
```bash
curl -X GET "http://localhost:27555/api/v1/messages/stats?period=24h" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Статистика сообщений
```json
{
  "success": true,
  "data": {
    "period": "all",
    "generated_at": "2025-01-11T10:32:15.456Z",
    "overview": {
      "total_messages": 23456,
      "correlated_messages": 23449,
      "buffered_messages": 7,
      "expired_messages": 0,
      "correlation_rate_percent": 99.97,
      "avg_correlation_time_ms": 5
    },
    "message_types": {
      "order_created": {
        "count": 8934,
        "correlation_rate": 100.0,
        "avg_correlation_time_ms": 3
      },
      "payment_completed": {
        "count": 6745,
        "correlation_rate": 99.8,
        "avg_correlation_time_ms": 8
      },
      "inventory_updated": {
        "count": 4567,
        "correlation_rate": 95.2,
        "avg_correlation_time_ms": 12
      }
    },
    "correlation_performance": {
      "avg_correlation_time_ms": 5,
      "p95_correlation_time_ms": 15,
      "p99_correlation_time_ms": 35,
      "fastest_correlation_ms": 1,
      "slowest_correlation_ms": 89
    },
    "subscriptions": {
      "active_subscriptions": 23,
      "avg_subscriptions_per_message": 1.2,
      "most_subscribed_messages": [
        {
          "message_name": "order_created",
          "subscription_count": 5
        },
        {
          "message_name": "payment_completed", 
          "subscription_count": 3
        }
      ]
    },
    "buffering_analysis": {
      "current_buffered": 7,
      "max_buffered": 15,
      "avg_buffer_time_hours": 2.3,
      "buffer_hit_rate_percent": 85.7
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/messages/publish`](./publish-message.md) - Публикация сообщений
- [`GET /api/v1/messages`](./list-messages.md) - Список сообщений
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
