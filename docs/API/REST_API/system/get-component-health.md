# GET /api/v1/system/components/:name/health

## Описание
Проверка здоровья конкретного компонента системы с диагностическими тестами.

## URL
```
GET /api/v1/system/components/{component_name}/health
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Параметры пути
- `component_name` (string): Имя компонента для проверки

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/components/storage/health" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Компонент здоров
```json
{
  "success": true,
  "data": {
    "component": "storage",
    "health": "HEALTHY",
    "checks": {
      "connectivity": {
        "status": "PASS",
        "response_time_ms": 3,
        "details": "Database connection active"
      },
      "performance": {
        "status": "PASS", 
        "read_latency_ms": 2.1,
        "write_latency_ms": 8.5
      },
      "capacity": {
        "status": "PASS",
        "used_percent": 17.4,
        "available_gb": 40.2
      }
    }
  }
}
```

### 503 Service Unavailable - Компонент нездоров
```json
{
  "success": false,
  "data": {
    "component": "storage",
    "health": "UNHEALTHY",
    "checks": {
      "connectivity": {
        "status": "FAIL",
        "error": "Connection timeout",
        "last_success": "2025-01-11T10:25:00.000Z"
      }
    }
  },
  "error": {
    "code": "COMPONENT_UNHEALTHY",
    "message": "Component health check failed"
  }
}
```

## Связанные endpoints
- [`GET /api/v1/system/components/:name`](./get-component-status.md) - Статус компонента
- [`GET /api/v1/system/health`](./system-health.md) - Общее здоровье системы
