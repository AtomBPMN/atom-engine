# DELETE /api/v1/timers/:id

## Описание
Отмена и удаление таймера из системы. Таймер переводится в статус CANCELLED и удаляется из timewheel.

## URL
```
DELETE /api/v1/timers/{timer_id}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer`

## Параметры пути
- `timer_id` (string): ID таймера для удаления

## Примеры запросов

### cURL
```bash
curl -X DELETE "http://localhost:27555/api/v1/timers/payment-timeout-ORD-12345" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const timerId = 'payment-timeout-ORD-12345';
const response = await fetch(`/api/v1/timers/${timerId}`, {
  method: 'DELETE',
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const result = await response.json();
```

## Ответы

### 200 OK - Таймер отменен
```json
{
  "success": true,
  "data": {
    "timer_id": "payment-timeout-ORD-12345",
    "status": "CANCELLED",
    "cancelled_at": "2025-01-11T10:32:15.456Z",
    "cancelled_by": "api-key-process-manager",
    "was_active": true,
    "remaining_time_ms": 180000,
    "cleanup_info": {
      "timewheel_level": 2,
      "slot_cleared": true,
      "callback_cancelled": true
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

### 409 Conflict - Таймер уже завершен
```json
{
  "success": false,
  "error": {
    "code": "TIMER_ALREADY_COMPLETED",
    "message": "Cannot cancel completed timer",
    "details": {
      "timer_id": "payment-timeout-ORD-12345",
      "current_status": "COMPLETED",
      "completed_at": "2025-01-11T10:31:00.000Z"
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/timers/:id`](./get-timer.md) - Проверить статус перед удалением
- [`POST /api/v1/timers`](./create-timer.md) - Создать новый таймер
