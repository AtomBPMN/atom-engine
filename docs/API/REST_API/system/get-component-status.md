# GET /api/v1/system/components/:name

## Описание
Получение детального статуса конкретного системного компонента.

## URL
```
GET /api/v1/system/components/{component_name}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Параметры пути
- `component_name` (string): Имя компонента (process_engine, timewheel, storage, etc.)

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/components/process_engine" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Статус компонента
```json
{
  "success": true,
  "data": {
    "name": "process_engine",
    "status": "HEALTHY",
    "version": "1.0.0",
    "uptime_seconds": 86400,
    "last_heartbeat": "2025-01-11T10:30:00.000Z",
    "metrics": {
      "active_processes": 142,
      "total_processes": 15847,
      "processes_per_minute": 25.4,
      "memory_usage_mb": 256
    },
    "configuration": {
      "max_concurrent_processes": 10000,
      "default_process_timeout": 3600
    }
  }
}
```

### 404 Not Found - Компонент не найден
```json
{
  "success": false,
  "error": {
    "code": "COMPONENT_NOT_FOUND",
    "message": "Component not found",
    "details": {
      "component_name": "unknown_component",
      "available_components": ["process_engine", "timewheel", "storage"]
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/system/components`](./list-components.md) - Список всех компонентов
- [`GET /api/v1/system/components/:name/health`](./get-component-health.md) - Проверка здоровья компонента
