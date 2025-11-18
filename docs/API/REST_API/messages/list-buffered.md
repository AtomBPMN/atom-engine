# GET /api/v1/messages/buffered

## Описание
Получение списка буферизованных сообщений, которые ожидают корреляции с процессами.

## URL
```
GET /api/v1/messages/buffered
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Параметры запроса (Query Parameters)

### Фильтрация
- `tenant_id` (string): Фильтр по тенанту
- `message_name` (string): Фильтр по имени сообщения
- `correlation_key` (string): Фильтр по ключу корреляции
- `expires_before` (string): Сообщения истекающие до даты (ISO 8601)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)

## Примеры запросов

### Все буферизованные сообщения
```bash
curl -X GET "http://localhost:27555/api/v1/messages/buffered" \
  -H "X-API-Key: your-api-key-here"
```

### Истекающие сообщения
```bash
curl -X GET "http://localhost:27555/api/v1/messages/buffered?expires_before=2025-01-12T00:00:00Z" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Буферизованные сообщения
```json
{
  "success": true,
  "data": {
    "buffered_messages": [
      {
        "message_id": "srv1-msg-buf123",
        "name": "inventory_updated",
        "correlation_key": "product-789",
        "published_at": "2025-01-11T10:15:00.000Z",
        "expires_at": "2025-01-12T10:15:00.000Z",
        "time_to_expiry_ms": 79200000,
        "variables": {
          "productId": "PROD-789",
          "newQuantity": 50
        },
        "tenant_id": "production"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 7,
      "total_pages": 1
    },
    "summary": {
      "total_buffered": 7,
      "expiring_soon": 2,
      "oldest_message_age_hours": 18
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/messages/publish`](./publish-message.md) - Публикация сообщений
- [`DELETE /api/v1/messages/cleanup`](./cleanup-messages.md) - Очистка просроченных
