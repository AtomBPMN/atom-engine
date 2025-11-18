# DELETE /api/v1/messages/cleanup

## Описание
Очистка просроченных сообщений из буфера. Удаляет сообщения, время жизни которых истекло.

## URL
```
DELETE /api/v1/messages/cleanup
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Параметры запроса (Query Parameters)
- `tenant_id` (string): Очистка для конкретного тенанта
- `force` (boolean): Принудительная очистка всех буферизованных сообщений
- `older_than` (string): Удалить сообщения старше указанного времени (ISO 8601)

## Примеры запросов

### Очистка просроченных
```bash
curl -X DELETE "http://localhost:27555/api/v1/messages/cleanup" \
  -H "X-API-Key: your-api-key-here"
```

### Принудительная очистка
```bash
curl -X DELETE "http://localhost:27555/api/v1/messages/cleanup?force=true" \
  -H "X-API-Key: your-api-key-here"
```

### Очистка старых сообщений
```bash
curl -X DELETE "http://localhost:27555/api/v1/messages/cleanup?older_than=P7D" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Очистка выполнена
```json
{
  "success": true,
  "data": {
    "cleanup_performed_at": "2025-01-11T10:30:00.000Z",
    "messages_deleted": 15,
    "cleanup_criteria": {
      "expired_messages": 12,
      "forced_cleanup": 0,
      "older_than_threshold": 3
    },
    "deleted_by_tenant": {
      "production": 10,
      "development": 3,
      "staging": 2
    },
    "storage_freed_bytes": 45620,
    "remaining_buffered": 7
  },
  "request_id": "req_1641998403500"
}
```

## Связанные endpoints
- [`GET /api/v1/messages/buffered`](./list-buffered.md) - Просмотр буферизованных сообщений
- [`GET /api/v1/messages/stats`](./get-message-stats.md) - Статистика сообщений
