# POST /api/v1/processes

## Описание
Запуск нового экземпляра BPMN процесса с опциональными переменными.

## URL
```
POST /api/v1/processes
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

```http
X-API-Key: your-api-key-here
```

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `process_id` (string): ID BPMN процесса для запуска

### Опциональные поля  
- `variables` (object): Переменные для инициализации процесса
- `version` (integer): Версия процесса (по умолчанию: последняя)
- `tenant_id` (string): ID тенанта (по умолчанию: "default")

### Пример тела запроса
```json
{
  "process_id": "order-fulfillment-v1",
  "variables": {
    "orderId": "ORD-12345",
    "customerId": "CUST-67890",
    "amount": 299.99,
    "priority": "high",
    "items": [
      {
        "sku": "PROD-001",
        "quantity": 2,
        "price": 149.99
      }
    ]
  },
  "version": 3,
  "tenant_id": "production"
}
```

## Примеры запросов

### cURL
```bash
curl -X POST http://localhost:27555/api/v1/processes \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "process_id": "order-fulfillment-v1",
    "variables": {
      "orderId": "ORD-12345",
      "amount": 299.99
    }
  }'
```

### JavaScript/Fetch
```javascript
const response = await fetch('/api/v1/processes', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify({
    process_id: 'order-fulfillment-v1',
    variables: {
      orderId: 'ORD-12345',
      amount: 299.99
    }
  })
});

const result = await response.json();
```

### Go
```go
data := map[string]interface{}{
    "process_id": "order-fulfillment-v1",
    "variables": map[string]interface{}{
        "orderId": "ORD-12345",
        "amount":  299.99,
    },
}

jsonData, _ := json.Marshal(data)
req, _ := http.NewRequest("POST", "/api/v1/processes", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 201 Created - Процесс успешно запущен
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_id": "order-fulfillment-v1",
    "process_key": "order-fulfillment-v1-3",
    "version": 3,
    "status": "ACTIVE",
    "tenant_id": "production",
    "started_at": "2025-01-11T10:30:00.123Z",
    "variables": {
      "orderId": "ORD-12345",
      "customerId": "CUST-67890", 
      "amount": 299.99,
      "priority": "high"
    },
    "current_activities": [
      {
        "element_id": "validate-order",
        "element_type": "serviceTask",
        "tokens": 1
      }
    ]
  },
  "request_id": "req_1641998400123"
}
```

### 400 Bad Request - Неверные данные запроса
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "field_errors": {
        "process_id": "Process ID is required",
        "variables.amount": "Amount must be a positive number"
      }
    }
  },
  "request_id": "req_1641998400124"
}
```

### 404 Not Found - Процесс не найден
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_NOT_FOUND",
    "message": "Process definition not found",
    "details": {
      "process_id": "non-existent-process",
      "version": 3,
      "available_versions": [1, 2]
    }
  },
  "request_id": "req_1641998400125"
}
```

### 401 Unauthorized - Неверный API ключ
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key",
    "details": null
  },
  "request_id": "req_1641998400126"
}
```

### 403 Forbidden - Недостаточно прав
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions for this operation",
    "details": {
      "required_permission": "process",
      "provided_permissions": ["bpmn", "storage"]
    }
  },
  "request_id": "req_1641998400127"
}
```

## Поля ответа (успешный запуск)

### Основная информация
- `instance_id` (string): Уникальный ID экземпляра процесса
- `process_id` (string): ID определения процесса
- `process_key` (string): Ключ процесса с версией
- `version` (integer): Версия процесса
- `status` (string): Текущий статус (`ACTIVE`, `COMPLETED`, `CANCELLED`)
- `tenant_id` (string): ID тенанта

### Временные метки
- `started_at` (string): Время запуска в ISO 8601 UTC

### Контекст выполнения  
- `variables` (object): Текущие переменные процесса
- `current_activities` (array): Активные элементы процесса

### Поля current_activities
- `element_id` (string): ID элемента BPMN
- `element_type` (string): Тип элемента (serviceTask, userTask, etc.)
- `tokens` (integer): Количество активных токенов

## Ограничения

### Размер данных
- **Максимальный размер запроса**: 10MB
- **Максимальный размер variables**: 5MB
- **Максимальная вложенность JSON**: 10 уровней

### Переменные
- Поддерживаемые типы: string, number, boolean, object, array, null
- Специальные символы в именах переменных экранируются
- Циклические ссылки не поддерживаются

## Валидация

### Process ID
- Формат: алфавитно-цифровой + дефисы/подчеркивания
- Длина: 1-100 символов  
- Не может начинаться с цифры

### Variables
- Имена переменных: 1-50 символов
- Значения строк: до 10KB
- Массивы: до 1000 элементов

## Использование в workflow

### Пример интеграции заказа
```javascript
// Запуск процесса обработки заказа
const processInstance = await startProcess({
  process_id: 'order-processing',
  variables: {
    order: orderData,
    customer: customerData
  }
});

// Мониторинг выполнения
const status = await getProcessStatus(processInstance.instance_id);
```

### Batch запуск процессов
```bash
# Запуск нескольких процессов
for order in $(cat orders.txt); do
  curl -X POST /api/v1/processes \
    -H "X-API-Key: $API_KEY" \
    -d "{\"process_id\":\"order-processing\",\"variables\":{\"orderId\":\"$order\"}}"
done
```

## Связанные endpoints
- [`GET /api/v1/processes/:id`](./get-process-status.md) - Получить статус процесса
- [`GET /api/v1/processes`](./list-processes.md) - Список процессов  
- [`DELETE /api/v1/processes/:id`](./cancel-process.md) - Отменить процесс
- [`GET /api/v1/bpmn/processes`](../bpmn/list-processes.md) - Список доступных процессов для запуска
