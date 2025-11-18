# GET /api/v1/messages

## Описание
Получение списка результатов корреляции сообщений с фильтрацией по статусу и тенанту.

## URL
```
GET /api/v1/messages
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Параметры запроса (Query Parameters)

### Фильтрация
- `tenant_id` (string): Фильтр по тенанту
- `name` (string): Фильтр по имени сообщения
- `correlation_key` (string): Фильтр по ключу корреляции
- `correlated` (boolean): Фильтр по статусу корреляции
- `period` (string): Период (`1h`, `24h`, `7d`, `30d`)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)

## Примеры запросов

### Все сообщения
```bash
curl -X GET "http://localhost:27555/api/v1/messages" \
  -H "X-API-Key: your-api-key-here"
```

### Коррелированные сообщения
```bash
curl -X GET "http://localhost:27555/api/v1/messages?correlated=true" \
  -H "X-API-Key: your-api-key-here"
```

### Сообщения за последний час
```bash
curl -X GET "http://localhost:27555/api/v1/messages?period=1h" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Список сообщений
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "message_id": "srv1-msg-abc123def456",
        "name": "payment_completed",
        "correlation_key": "order-123",
        "published_at": "2025-01-11T10:30:00.000Z",
        "tenant_id": "production",
        "correlated": true,
        "correlation_result": {
          "processes_triggered": 1,
          "processes_started": 0,
          "subscriptions_matched": 1
        },
        "variables": {
          "paymentId": "pay_abc123",
          "amount": 299.99
        }
      },
      {
        "message_id": "srv1-msg-def456ghi789", 
        "name": "inventory_updated",
        "correlation_key": "product-456",
        "published_at": "2025-01-11T10:25:00.000Z",
        "tenant_id": "production",
        "correlated": false,
        "buffered": true,
        "expires_at": "2025-01-12T10:25:00.000Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 156,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    },
    "summary": {
      "total_messages": 156,
      "correlated_messages": 143,
      "buffered_messages": 13,
      "correlation_rate_percent": 91.7
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/messages/publish`](./publish-message.md) - Публикация сообщения
- [`GET /api/v1/messages/subscriptions`](./list-subscriptions.md) - Подписки на сообщения
