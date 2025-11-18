# GET /api/v1/timers/:id

## Описание
Получение детальной информации о конкретном таймере, включая статус, оставшееся время и историю выполнения.

## URL
```
GET /api/v1/timers/{timer_id}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer`

## Параметры пути
- `timer_id` (string): ID таймера

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/timers/payment-timeout-ORD-12345" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const timerId = 'payment-timeout-ORD-12345';
const response = await fetch(`/api/v1/timers/${timerId}`, {
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const timer = await response.json();
```

## Ответы

### 200 OK - Детали таймера
```json
{
  "success": true,
  "data": {
    "timer_id": "payment-timeout-ORD-12345",
    "duration_or_cycle": "PT5M",
    "status": "SCHEDULED",
    "timer_type": "DURATION",
    "created_at": "2025-01-11T10:30:00.000Z",
    "scheduled_at": "2025-01-11T10:35:00.000Z",
    "remaining_time_ms": 240000,
    "accuracy_ms": 12,
    "tenant_id": "production",
    "metadata": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "payment-timeout-boundary",
      "order_id": "ORD-12345"
    },
    "callback_url": "http://localhost:27555/internal/timer/callback",
    "execution_history": [
      {
        "timestamp": "2025-01-11T10:30:00.000Z",
        "action": "CREATED",
        "details": "Timer created and scheduled"
      }
    ],
    "timewheel_info": {
      "level": 2,
      "slot_position": 156,
      "estimated_accuracy_ms": 12
    }
  }
}
```

### 404 Not Found - Таймер не найден
```json
{
  "success": false,
  "error": {
    "code": "TIMER_NOT_FOUND",
    "message": "Timer not found",
    "details": {
      "timer_id": "non-existent-timer"
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/timers`](./create-timer.md) - Создать таймер
- [`DELETE /api/v1/timers/:id`](./delete-timer.md) - Удалить таймер
