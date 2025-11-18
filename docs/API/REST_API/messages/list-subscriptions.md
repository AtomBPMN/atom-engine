# GET /api/v1/messages/subscriptions

## Описание
Получение списка активных подписок на сообщения в BPMN процессах.

## URL
```
GET /api/v1/messages/subscriptions
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Параметры запроса (Query Parameters)

### Фильтрация
- `tenant_id` (string): Фильтр по тенанту
- `message_name` (string): Фильтр по имени сообщения
- `process_id` (string): Фильтр по процессу
- `correlation_key` (string): Фильтр по ключу корреляции

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)

## Примеры запросов

### Все подписки
```bash
curl -X GET "http://localhost:27555/api/v1/messages/subscriptions" \
  -H "X-API-Key: your-api-key-here"
```

### Подписки на конкретное сообщение
```bash
curl -X GET "http://localhost:27555/api/v1/messages/subscriptions?message_name=payment_completed" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - Список подписок
```json
{
  "success": true,
  "data": {
    "subscriptions": [
      {
        "subscription_id": "srv1-sub-abc123",
        "message_name": "payment_completed",
        "correlation_key": "orderId",
        "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "process_id": "order-fulfillment",
        "element_id": "wait-payment-confirmation",
        "element_type": "intermediateCatchEvent",
        "created_at": "2025-01-11T10:20:00.000Z",
        "tenant_id": "production"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 23,
      "total_pages": 2
    },
    "summary": {
      "total_subscriptions": 23,
      "unique_message_names": 8,
      "active_processes": 15
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/messages/publish`](./publish-message.md) - Публикация для корреляции
- [`GET /api/v1/messages`](./list-messages.md) - Результаты корреляции
