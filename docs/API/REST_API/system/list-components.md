# GET /api/v1/system/components

## Описание
Получение списка всех системных компонентов с их статусами и характеристиками.

## URL
```
GET /api/v1/system/components
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/components" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Список компонентов
```json
{
  "success": true,
  "data": {
    "components": [
      {
        "name": "process_engine",
        "status": "HEALTHY",
        "version": "1.0.0",
        "uptime_seconds": 86400,
        "description": "BPMN process execution engine"
      },
      {
        "name": "timewheel",
        "status": "HEALTHY", 
        "version": "1.0.0",
        "uptime_seconds": 86400,
        "description": "Hierarchical timer management system"
      }
    ],
    "total_count": 8,
    "healthy_count": 8,
    "degraded_count": 0,
    "unhealthy_count": 0
  }
}
```

## Связанные endpoints
- [`GET /api/v1/system/components/:name`](./get-component-status.md) - Статус конкретного компонента
