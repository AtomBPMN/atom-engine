# GET /api/v1/system/health

## Описание
Расширенная проверка здоровья системы с детальной диагностикой компонентов.

## URL
```
GET /api/v1/system/health
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/health" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Система здорова
```json
{
  "success": true,
  "data": {
    "overall_health": "HEALTHY",
    "checks": {
      "database_connectivity": {
        "status": "PASS",
        "response_time_ms": 5,
        "details": "Connection pool healthy"
      },
      "process_engine": {
        "status": "PASS", 
        "active_processes": 142,
        "memory_usage_mb": 256
      },
      "timewheel": {
        "status": "PASS",
        "active_timers": 89,
        "accuracy_ms": 12
      }
    }
  }
}
```

## Связанные endpoints
- [`GET /health`](../health/health-check.md) - Базовая проверка здоровья
- [`GET /api/v1/system/status`](./system-status.md) - Детальный статус
