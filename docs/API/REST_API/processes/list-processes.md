# GET /api/v1/processes

## Описание
Получение списка экземпляров процессов с фильтрацией и пагинацией.

## URL
```
GET /api/v1/processes
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

```http
X-API-Key: your-api-key-here
```

## Параметры запроса (Query Parameters)

### Фильтрация
- `status` (string): Фильтр по статусу (`ACTIVE`, `COMPLETED`, `CANCELLED`)
- `process_id` (string): Фильтр по ID процесса
- `tenant_id` (string): Фильтр по тенанту  
- `started_after` (string): Процессы запущенные после даты (ISO 8601)
- `started_before` (string): Процессы запущенные до даты (ISO 8601)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)
- `sort_by` (string): Поле сортировки (по умолчанию: "started_at")
- `sort_order` (string): Порядок сортировки (`ASC`, `DESC`, по умолчанию: "DESC")

## Примеры запросов

### Базовый запрос
```bash
curl -X GET "http://localhost:27555/api/v1/processes" \
  -H "X-API-Key: your-api-key-here"
```

### Фильтрация по статусу
```bash
curl -X GET "http://localhost:27555/api/v1/processes?status=ACTIVE" \
  -H "X-API-Key: your-api-key-here"
```

### Фильтрация по процессу с пагинацией
```bash
curl -X GET "http://localhost:27555/api/v1/processes?process_id=order-processing&page=2&page_size=10" \
  -H "X-API-Key: your-api-key-here"
```

### Фильтрация по времени
```bash
curl -X GET "http://localhost:27555/api/v1/processes?started_after=2025-01-01T00:00:00Z&started_before=2025-01-31T23:59:59Z" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const params = new URLSearchParams({
  status: 'ACTIVE',
  process_id: 'order-processing',
  page: '1',
  page_size: '20'
});

const response = await fetch(`/api/v1/processes?${params}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const result = await response.json();
```

### Go
```go
params := url.Values{}
params.Add("status", "ACTIVE")
params.Add("process_id", "order-processing")
params.Add("page", "1")
params.Add("page_size", "20")

req, _ := http.NewRequest("GET", "/api/v1/processes?"+params.Encode(), nil)
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Успешное получение списка
```json
{
  "success": true,
  "data": {
    "processes": [
      {
        "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "process_id": "order-processing",
        "process_key": "order-processing-v2",
        "version": 2,
        "status": "ACTIVE",
        "tenant_id": "production",
        "started_at": "2025-01-11T10:30:00.123Z",
        "updated_at": "2025-01-11T10:32:15.456Z",
        "current_activity": "validate-payment",
        "variables": {
          "orderId": "ORD-12345",
          "amount": 299.99,
          "status": "processing"
        }
      },
      {
        "instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
        "process_id": "order-processing", 
        "process_key": "order-processing-v2",
        "version": 2,
        "status": "COMPLETED",
        "tenant_id": "production",
        "started_at": "2025-01-11T09:15:30.789Z",
        "updated_at": "2025-01-11T09:18:45.012Z",
        "completed_at": "2025-01-11T09:18:45.012Z",
        "current_activity": null,
        "variables": {
          "orderId": "ORD-12340",
          "amount": 150.00,
          "status": "completed"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 142,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    }
  },
  "request_id": "req_1641998400200"
}
```

### 400 Bad Request - Неверные параметры
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "parameter_errors": {
        "page_size": "Page size must be between 1 and 100",
        "sort_by": "Invalid sort field. Allowed: started_at, updated_at, status",
        "started_after": "Invalid date format. Use ISO 8601"
      }
    }
  },
  "request_id": "req_1641998400201"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key",
    "details": null
  },
  "request_id": "req_1641998400202"
}
```

## Поля ответа

### Process Object
- `instance_id` (string): Уникальный ID экземпляра
- `process_id` (string): ID определения процесса
- `process_key` (string): Ключ процесса с версией
- `version` (integer): Версия процесса
- `status` (string): Статус процесса
- `tenant_id` (string): ID тенанта
- `started_at` (string): Время запуска (ISO 8601 UTC)
- `updated_at` (string): Время последнего обновления
- `completed_at` (string, nullable): Время завершения
- `cancelled_at` (string, nullable): Время отмены
- `current_activity` (string, nullable): Текущая активность
- `variables` (object): Переменные процесса

### Pagination Object
- `page` (integer): Текущая страница
- `page_size` (integer): Размер страницы
- `total_count` (integer): Общее количество записей
- `total_pages` (integer): Общее количество страниц
- `has_next` (boolean): Есть ли следующая страница
- `has_prev` (boolean): Есть ли предыдущая страница

## Валидация параметров

### page
- Минимум: 1
- Максимум: 10000

### page_size
- Минимум: 1
- Максимум: 100
- По умолчанию: 20

### sort_by
Допустимые значения:
- `started_at` (по умолчанию)
- `updated_at`
- `status`
- `process_id`

### sort_order
- `ASC` - по возрастанию
- `DESC` - по убыванию (по умолчанию)

### Даты
- Формат: ISO 8601 UTC (`2025-01-11T10:30:00Z`)
- Временные зоны поддерживаются (`2025-01-11T10:30:00+03:00`)

## Производительность

### Оптимизация запросов
- **Индексы**: на `status`, `process_id`, `started_at`, `tenant_id`
- **Кэширование**: результаты кэшируются на 30 секунд
- **Лимиты**: максимум 100 записей на запрос

### Рекомендации
```bash
# Быстрые запросы (используют индексы)
GET /api/v1/processes?status=ACTIVE
GET /api/v1/processes?process_id=order-processing
GET /api/v1/processes?started_after=2025-01-01T00:00:00Z

# Медленные запросы (полное сканирование)
GET /api/v1/processes?sort_by=variables  # не поддерживается
```

## Использование

### Мониторинг процессов
```javascript
// Получить активные процессы
const activeProcesses = await fetch('/api/v1/processes?status=ACTIVE');

// Получить процессы за последний час
const recentProcesses = await fetch(
  `/api/v1/processes?started_after=${new Date(Date.now() - 3600000).toISOString()}`
);
```

### Пагинация через все страницы
```javascript
async function getAllProcesses() {
  let allProcesses = [];
  let page = 1;
  let hasNext = true;
  
  while (hasNext) {
    const response = await fetch(`/api/v1/processes?page=${page}&page_size=100`);
    const result = await response.json();
    
    allProcesses.push(...result.data.processes);
    hasNext = result.data.pagination.has_next;
    page++;
  }
  
  return allProcesses;
}
```

## Связанные endpoints
- [`POST /api/v1/processes`](./start-process.md) - Запуск процесса
- [`GET /api/v1/processes/:id`](./get-process-status.md) - Детали процесса
- [`GET /api/v1/processes/stats`](./get-process-stats.md) - Статистика процессов
