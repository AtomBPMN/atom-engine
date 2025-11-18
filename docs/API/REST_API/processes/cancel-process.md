# DELETE /api/v1/processes/:id

## Описание
Отмена активного экземпляра процесса. Процесс переводится в статус CANCELLED и все активные токены завершаются.

## URL
```
DELETE /api/v1/processes/{instance_id}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

```http
X-API-Key: your-api-key-here
```

## Параметры пути
- `instance_id` (string, обязательный): ID экземпляра процесса для отмены

## Параметры запроса (Query Parameters)
- `reason` (string, опциональный): Причина отмены процесса (максимум 500 символов)

## Заголовки запроса
```http
Accept: application/json
X-API-Key: your-api-key-here
```

## Примеры запросов

### Базовая отмена
```bash
curl -X DELETE "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV" \
  -H "X-API-Key: your-api-key-here"
```

### Отмена с указанием причины
```bash
curl -X DELETE "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV?reason=Customer%20requested%20cancellation" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV?reason=Order cancelled by customer', {
  method: 'DELETE',
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const result = await response.json();
```

### Go
```go
instanceID := "srv1-aB3dEf9hK2mN5pQ8uV"
reason := "Order cancelled by customer"

req, _ := http.NewRequest("DELETE", 
  fmt.Sprintf("/api/v1/processes/%s?reason=%s", instanceID, url.QueryEscape(reason)), 
  nil)
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Процесс успешно отменен
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_id": "order-processing",
    "process_key": "order-processing-v2", 
    "version": 2,
    "status": "CANCELLED",
    "tenant_id": "production",
    "started_at": "2025-01-11T10:30:00.123Z",
    "cancelled_at": "2025-01-11T10:45:30.789Z",
    "cancellation_reason": "Customer requested cancellation",
    "cancelled_by": "api-key-process-manager",
    "active_tokens_cancelled": 3,
    "completed_activities": [
      "validate-order",
      "check-inventory"
    ],
    "cancelled_activities": [
      "process-payment",
      "send-confirmation-email",
      "update-inventory"
    ]
  },
  "request_id": "req_1641998400300"
}
```

### 404 Not Found - Процесс не найден
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_INSTANCE_NOT_FOUND",
    "message": "Process instance not found",
    "details": {
      "instance_id": "srv1-nonexistent123",
      "possible_reasons": [
        "Instance ID does not exist",
        "Instance already completed or cancelled",
        "Instance belongs to different tenant"
      ]
    }
  },
  "request_id": "req_1641998400301"
}
```

### 409 Conflict - Процесс нельзя отменить
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_NOT_CANCELLABLE",
    "message": "Process instance cannot be cancelled",
    "details": {
      "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "current_status": "COMPLETED",
      "cancellable_statuses": ["ACTIVE"],
      "completed_at": "2025-01-11T10:40:00.000Z"
    }
  },
  "request_id": "req_1641998400302"
}
```

### 400 Bad Request - Неверные параметры
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "parameter_errors": {
        "instance_id": "Invalid instance ID format",
        "reason": "Reason must not exceed 500 characters"
      }
    }
  },
  "request_id": "req_1641998400303"
}
```

## Поля ответа (успешная отмена)

### Основная информация
- `instance_id` (string): ID отмененного экземпляра
- `process_id` (string): ID определения процесса
- `process_key` (string): Ключ процесса с версией
- `version` (integer): Версия процесса
- `status` (string): Новый статус (`CANCELLED`)
- `tenant_id` (string): ID тенанта

### Информация об отмене
- `cancelled_at` (string): Время отмены (ISO 8601 UTC)
- `cancellation_reason` (string): Причина отмены
- `cancelled_by` (string): Идентификатор того, кто отменил процесс
- `active_tokens_cancelled` (integer): Количество отмененных активных токенов

### Состояние выполнения
- `completed_activities` (array): Активности, которые успели завершиться
- `cancelled_activities` (array): Активности, которые были отменены

## Поведение при отмене

### Что происходит
1. **Активные токены**: Все активные токены завершаются
2. **Задания**: Активные задания отменяются
3. **Таймеры**: Все таймеры процесса отменяются
4. **Подписки**: Подписки на сообщения удаляются
5. **Состояние**: Процесс переводится в статус CANCELLED

### Что НЕ происходит  
- **Откат**: Завершенные активности НЕ откатываются
- **Внешние системы**: Уже выполненные внешние вызовы НЕ отменяются
- **Данные**: Переменные процесса сохраняются

## Ограничения

### Статусы процессов
Отменить можно только процессы со статусом:
- `ACTIVE` ✅

Нельзя отменить процессы со статусом:
- `COMPLETED` ❌
- `CANCELLED` ❌

### Права доступа
- Требуется разрешение `process`
- В multi-tenant окружении можно отменять только процессы своего тенанта

### Время выполнения
- **Простые процессы**: < 100ms
- **Сложные процессы**: < 2s
- **Таймаут**: 30s

## Использование

### Отмена по бизнес-событию
```javascript
// Отмена заказа при возврате
async function cancelOrderProcess(orderId, reason) {
  const processes = await fetch(`/api/v1/processes?process_id=order-processing&status=ACTIVE`);
  const processData = await processes.json();
  
  for (const process of processData.data.processes) {
    if (process.variables.orderId === orderId) {
      await fetch(`/api/v1/processes/${process.instance_id}?reason=${encodeURIComponent(reason)}`, {
        method: 'DELETE',
        headers: { 'X-API-Key': 'your-api-key-here' }
      });
    }
  }
}
```

### Массовая отмена
```bash
# Отмена всех процессов определенного типа
for instance_id in $(curl -s "/api/v1/processes?process_id=old-process&status=ACTIVE" | jq -r '.data.processes[].instance_id'); do
  curl -X DELETE "/api/v1/processes/$instance_id?reason=Process%20deprecated"
done
```

### Graceful shutdown
```javascript
// Отмена всех активных процессов при остановке системы
async function gracefulShutdown() {
  const activeProcesses = await fetch('/api/v1/processes?status=ACTIVE');
  const processes = await activeProcesses.json();
  
  const cancelPromises = processes.data.processes.map(process => 
    fetch(`/api/v1/processes/${process.instance_id}?reason=System shutdown`, {
      method: 'DELETE'
    })
  );
  
  await Promise.all(cancelPromises);
}
```

## Мониторинг отмен

### Метрики
- Количество отмененных процессов
- Распределение причин отмены
- Время выполнения отмены

### Логирование
```json
{
  "level": "INFO",
  "message": "Process instance cancelled",
  "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
  "process_id": "order-processing",
  "reason": "Customer requested cancellation",
  "cancelled_by": "api-key-process-manager",
  "duration_ms": 150
}
```

## Связанные endpoints
- [`GET /api/v1/processes/:id`](./get-process-status.md) - Проверить статус процесса
- [`GET /api/v1/processes`](./list-processes.md) - Найти процессы для отмены
- [`POST /api/v1/processes`](./start-process.md) - Запустить новый процесс взамен
- [`GET /api/v1/incidents`](../incidents/list-incidents.md) - Проверить инциденты после отмены
