# GET /api/v1/bpmn/processes

## Описание
Получение списка всех BPMN процессов с их метаданными, статистикой и информацией о версиях.

## URL
```
GET /api/v1/bpmn/processes
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры запроса (Query Parameters)

### Фильтрация
- `tenant_id` (string): Фильтр по тенанту
- `name` (string): Поиск по названию процесса (частичное совпадение)
- `executable` (boolean): Только исполняемые процессы
- `created_after` (string): Процессы созданные после даты (ISO 8601)
- `created_before` (string): Процессы созданные до даты (ISO 8601)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 20, максимум: 100)
- `sort_by` (string): Поле сортировки (по умолчанию: "created_at")
- `sort_order` (string): Порядок сортировки (`ASC`, `DESC`, по умолчанию: "DESC")

## Примеры запросов

### Базовый запрос
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes" \
  -H "X-API-Key: your-api-key-here"
```

### Поиск по названию
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes?name=order" \
  -H "X-API-Key: your-api-key-here"
```

### Только исполняемые процессы
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes?executable=true&page_size=50" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const params = new URLSearchParams({
  name: 'order',
  executable: 'true',
  page: '1',
  page_size: '20'
});

const response = await fetch(`/api/v1/bpmn/processes?${params}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processes = await response.json();
```

## Ответы

### 200 OK - Список процессов
```json
{
  "success": true,
  "data": {
    "processes": [
      {
        "process_id": "order-processing-v2",
        "process_key": "order-processing-v2-3",
        "version": 3,
        "name": "Order Processing Workflow",
        "description": "Complete order processing from validation to fulfillment",
        "tenant_id": "production",
        "is_executable": true,
        "created_at": "2025-01-11T10:30:00.000Z",
        "updated_at": "2025-01-11T10:30:00.000Z",
        "file_info": {
          "file_name": "order-processing-v2.bpmn",
          "file_size_bytes": 15420,
          "file_hash": "sha256:abc123def456..."
        },
        "elements": {
          "start_events": 1,
          "end_events": 2,
          "tasks": 8,
          "gateways": 3,
          "flows": 15,
          "boundary_events": 2,
          "total": 31
        },
        "versions": {
          "latest_version": 3,
          "total_versions": 3,
          "versions_list": [1, 2, 3]
        },
        "usage_stats": {
          "total_instances": 1247,
          "active_instances": 23,
          "completed_instances": 1224,
          "last_execution": "2025-01-11T10:25:00.000Z"
        }
      },
      {
        "process_id": "user-registration",
        "process_key": "user-registration-1",
        "version": 1,
        "name": "User Registration Process",
        "description": "New user registration and verification workflow",
        "tenant_id": "production",
        "is_executable": true,
        "created_at": "2025-01-10T15:45:00.000Z",
        "updated_at": "2025-01-10T15:45:00.000Z",
        "file_info": {
          "file_name": "user-registration.bpmn",
          "file_size_bytes": 8920,
          "file_hash": "sha256:def456ghi789..."
        },
        "elements": {
          "start_events": 1,
          "end_events": 1,
          "tasks": 5,
          "gateways": 1,
          "flows": 7,
          "boundary_events": 0,
          "total": 15
        },
        "versions": {
          "latest_version": 1,
          "total_versions": 1,
          "versions_list": [1]
        },
        "usage_stats": {
          "total_instances": 456,
          "active_instances": 12,
          "completed_instances": 444,
          "last_execution": "2025-01-11T10:20:00.000Z"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 47,
      "total_pages": 3,
      "has_next": true,
      "has_prev": false
    },
    "summary": {
      "total_processes": 47,
      "executable_processes": 45,
      "non_executable_processes": 2,
      "total_versions": 89,
      "avg_elements_per_process": 22.3
    }
  },
  "request_id": "req_1641998401500"
}
```

## Поля ответа

### Process Object
- `process_id` (string): ID процесса
- `process_key` (string): Уникальный ключ с версией
- `version` (integer): Версия процесса
- `name` (string): Название процесса
- `description` (string): Описание
- `tenant_id` (string): ID тенанта
- `is_executable` (boolean): Можно ли выполнять

### File Information
- `file_name` (string): Имя исходного BPMN файла
- `file_size_bytes` (integer): Размер файла
- `file_hash` (string): SHA256 хеш содержимого

### Element Statistics
- `start_events` (integer): Количество start events
- `end_events` (integer): Количество end events
- `tasks` (integer): Количество задач
- `gateways` (integer): Количество шлюзов
- `flows` (integer): Количество потоков
- `boundary_events` (integer): Количество boundary events
- `total` (integer): Общее количество элементов

### Version Information
- `latest_version` (integer): Последняя версия
- `total_versions` (integer): Общее количество версий
- `versions_list` (array): Список всех версий

### Usage Statistics
- `total_instances` (integer): Общее количество запусков
- `active_instances` (integer): Активные экземпляры
- `completed_instances` (integer): Завершенные экземпляры
- `last_execution` (string): Время последнего запуска

## Фильтрация и поиск

### Поиск по названию
```bash
# Поиск процессов содержащих "order"
GET /api/v1/bpmn/processes?name=order

# Case-insensitive поиск
GET /api/v1/bpmn/processes?name=ORDER
```

### Фильтрация по исполняемости
```bash
# Только исполняемые процессы
GET /api/v1/bpmn/processes?executable=true

# Только не исполняемые (например, для отладки)
GET /api/v1/bpmn/processes?executable=false
```

### Временная фильтрация
```bash
# Процессы созданные за последние 24 часа
SINCE=$(date -d '24 hours ago' --iso-8601=seconds)
GET /api/v1/bpmn/processes?created_after=$SINCE

# Процессы созданные в определенном диапазоне
GET /api/v1/bpmn/processes?created_after=2025-01-01T00:00:00Z&created_before=2025-01-31T23:59:59Z
```

## Сортировка

### Доступные поля для сортировки
- `created_at` - По дате создания (по умолчанию)
- `updated_at` - По дате обновления
- `name` - По названию процесса
- `process_id` - По ID процесса
- `version` - По версии
- `total_instances` - По количеству запусков

### Примеры сортировки
```bash
# Сортировка по названию (A-Z)
GET /api/v1/bpmn/processes?sort_by=name&sort_order=ASC

# Самые популярные процессы
GET /api/v1/bpmn/processes?sort_by=total_instances&sort_order=DESC

# Последние обновленные
GET /api/v1/bpmn/processes?sort_by=updated_at&sort_order=DESC
```

## Использование

### Process Catalog
```javascript
// Создание каталога процессов
async function buildProcessCatalog() {
  const response = await fetch('/api/v1/bpmn/processes?page_size=100');
  const data = await response.json();
  
  const catalog = data.data.processes.map(process => ({
    id: process.process_id,
    name: process.name,
    description: process.description,
    complexity: process.elements.total,
    popularity: process.usage_stats.total_instances,
    isActive: process.usage_stats.active_instances > 0
  }));
  
  return catalog.sort((a, b) => b.popularity - a.popularity);
}
```

### Process Health Check
```javascript
async function checkProcessHealth() {
  const response = await fetch('/api/v1/bpmn/processes?executable=false');
  const data = await response.json();
  
  const brokenProcesses = data.data.processes;
  
  if (brokenProcesses.length > 0) {
    console.warn(`Found ${brokenProcesses.length} non-executable processes:`);
    brokenProcesses.forEach(process => {
      console.log(`- ${process.process_id}: ${process.name}`);
    });
  }
  
  return brokenProcesses;
}
```

### Version Analysis
```javascript
async function analyzeProcessVersions() {
  const response = await fetch('/api/v1/bpmn/processes?page_size=100');
  const data = await response.json();
  
  const versionStats = data.data.processes.map(process => ({
    process_id: process.process_id,
    versions: process.versions.total_versions,
    latest: process.versions.latest_version,
    outdated: process.versions.latest_version > 1
  }));
  
  const outdatedCount = versionStats.filter(p => p.outdated).length;
  console.log(`${outdatedCount} processes have multiple versions`);
  
  return versionStats;
}
```

### Deployment Dashboard
```javascript
// Данные для deployment dashboard
async function getDeploymentOverview() {
  const [processes, stats] = await Promise.all([
    fetch('/api/v1/bpmn/processes?page_size=100'),
    fetch('/api/v1/bpmn/stats')
  ]);
  
  const processData = await processes.json();
  const statsData = await stats.json();
  
  return {
    totalProcesses: processData.data.summary.total_processes,
    executableProcesses: processData.data.summary.executable_processes,
    recentDeployments: processData.data.processes
      .filter(p => {
        const createdAt = new Date(p.created_at);
        const dayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);
        return createdAt > dayAgo;
      }),
    popularProcesses: processData.data.processes
      .sort((a, b) => b.usage_stats.total_instances - a.usage_stats.total_instances)
      .slice(0, 5)
  };
}
```

## Мониторинг

### Process Usage Tracking
```bash
#!/bin/bash
# Отслеживание использования процессов
curl -s -H "X-API-Key: $API_KEY" \
  "/api/v1/bpmn/processes?sort_by=total_instances&sort_order=DESC&page_size=10" | \
  jq -r '.data.processes[] | "\(.process_id): \(.usage_stats.total_instances) instances"'
```

### Unused Processes Detection
```javascript
async function findUnusedProcesses() {
  const response = await fetch('/api/v1/bpmn/processes?page_size=100');
  const data = await response.json();
  
  const unusedProcesses = data.data.processes.filter(process => {
    const lastExecution = new Date(process.usage_stats.last_execution);
    const monthAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    
    return process.usage_stats.total_instances === 0 || lastExecution < monthAgo;
  });
  
  console.log(`Found ${unusedProcesses.length} unused processes`);
  return unusedProcesses;
}
```

### Process Complexity Analysis
```javascript
async function analyzeComplexity() {
  const response = await fetch('/api/v1/bpmn/processes?page_size=100');
  const data = await response.json();
  
  const complexityStats = data.data.processes.map(process => ({
    process_id: process.process_id,
    complexity: process.elements.total,
    category: process.elements.total < 10 ? 'simple' :
              process.elements.total < 30 ? 'medium' : 'complex'
  }));
  
  const categories = complexityStats.reduce((acc, p) => {
    acc[p.category] = (acc[p.category] || 0) + 1;
    return acc;
  }, {});
  
  console.log('Process complexity distribution:', categories);
  return complexityStats;
}
```

## Связанные endpoints
- [`POST /api/v1/bpmn/parse`](./parse-bpmn.md) - Добавление нового процесса
- [`GET /api/v1/bpmn/processes/:key`](./get-process.md) - Детали конкретного процесса
- [`GET /api/v1/bpmn/stats`](./get-bpmn-stats.md) - Общая статистика BPMN
- [`POST /api/v1/processes`](../processes/start-process.md) - Запуск процесса
