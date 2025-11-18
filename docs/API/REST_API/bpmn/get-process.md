# GET /api/v1/bpmn/processes/:key

## Описание
Получение детальной информации о конкретном BPMN процессе по его ключу (process key).

## URL
```
GET /api/v1/bpmn/processes/{process_key}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры пути
- `process_key` (string): Уникальный ключ процесса (например: "order-processing-v2-3")

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2-3" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const processKey = 'order-processing-v2-3';
const response = await fetch(`/api/v1/bpmn/processes/${processKey}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processDetails = await response.json();
```

### Go
```go
processKey := "order-processing-v2-3"
req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/bpmn/processes/%s", processKey), nil)
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Процесс найден
```json
{
  "success": true,
  "data": {
    "process_id": "order-processing-v2",
    "process_key": "order-processing-v2-3",
    "version": 3,
    "name": "Order Processing Workflow",
    "description": "Complete order processing from validation to fulfillment",
    "tenant_id": "production",
    "is_executable": true,
    "created_at": "2025-01-11T10:30:00.000Z",
    "updated_at": "2025-01-11T10:30:00.000Z",
    "created_by": "api-key-deployment-manager",
    "file_info": {
      "file_name": "order-processing-v2.bpmn",
      "file_size_bytes": 15420,
      "file_hash": "sha256:abc123def456789...",
      "original_path": "/deployments/2025-01-11/order-processing-v2.bpmn"
    },
    "elements": {
      "start_events": 1,
      "end_events": 2,
      "service_tasks": 6,
      "user_tasks": 2,
      "exclusive_gateways": 2,
      "parallel_gateways": 1,
      "sequence_flows": 15,
      "boundary_events": 2,
      "total": 31,
      "details": [
        {
          "id": "start_order",
          "type": "startEvent",
          "name": "Order Received"
        },
        {
          "id": "validate_order",
          "type": "serviceTask", 
          "name": "Validate Order",
          "task_definition": "order-validator"
        },
        {
          "id": "payment_gateway",
          "type": "exclusiveGateway",
          "name": "Payment Method?"
        }
      ]
    },
    "versions": {
      "current_version": 3,
      "latest_version": 3,
      "total_versions": 3,
      "version_history": [
        {
          "version": 1,
          "created_at": "2025-01-09T10:00:00.000Z",
          "created_by": "api-key-initial-deployment",
          "changes": "Initial version"
        },
        {
          "version": 2,
          "created_at": "2025-01-10T15:30:00.000Z", 
          "created_by": "api-key-dev-team",
          "changes": "Added payment validation step"
        },
        {
          "version": 3,
          "created_at": "2025-01-11T10:30:00.000Z",
          "created_by": "api-key-deployment-manager",
          "changes": "Optimized parallel processing"
        }
      ]
    },
    "usage_stats": {
      "total_instances": 1247,
      "active_instances": 23,
      "completed_instances": 1204,
      "cancelled_instances": 20,
      "failed_instances": 0,
      "success_rate_percent": 98.4,
      "avg_duration_seconds": 45,
      "last_execution": "2025-01-11T10:25:00.000Z",
      "executions_last_24h": 156,
      "executions_last_7d": 987
    },
    "performance": {
      "avg_processing_time_ms": 45000,
      "p50_processing_time_ms": 38000,
      "p95_processing_time_ms": 78000,
      "p99_processing_time_ms": 125000,
      "bottleneck_activities": [
        {
          "activity_id": "payment_processing",
          "avg_duration_ms": 15000,
          "percentage_of_total": 33.3
        }
      ]
    },
    "dependencies": {
      "message_definitions": [
        {
          "name": "order_confirmation",
          "correlation_key": "orderId"
        }
      ],
      "task_definitions": [
        {
          "type": "order-validator",
          "retries": 3,
          "timeout": "PT30S"
        },
        {
          "type": "payment-processor",
          "retries": 2,
          "timeout": "PT60S"
        }
      ],
      "timer_definitions": [
        {
          "id": "payment_timeout",
          "duration": "PT5M",
          "type": "boundary"
        }
      ]
    },
    "validation": {
      "is_valid": true,
      "last_validation": "2025-01-11T10:30:00.000Z",
      "errors": [],
      "warnings": [
        {
          "element_id": "send_email_task",
          "message": "Task has no timeout configuration",
          "severity": "LOW"
        }
      ]
    }
  },
  "request_id": "req_1641998401600"
}
```

### 404 Not Found - Процесс не найден
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_NOT_FOUND",
    "message": "BPMN process not found",
    "details": {
      "process_key": "non-existent-process-key",
      "suggestions": [
        "order-processing-v2-3",
        "user-registration-1",
        "payment-workflow-2"
      ]
    }
  },
  "request_id": "req_1641998401601"
}
```

## Поля ответа

### Basic Information
- `process_id` (string): ID процесса
- `process_key` (string): Уникальный ключ с версией
- `version` (integer): Версия процесса
- `name` (string): Название процесса
- `description` (string): Описание процесса
- `tenant_id` (string): ID тенанта
- `is_executable` (boolean): Можно ли выполнять процесс

### File Information
- `file_name` (string): Имя исходного файла
- `file_size_bytes` (integer): Размер файла
- `file_hash` (string): SHA256 хеш
- `original_path` (string): Исходный путь файла

### Element Details
Подробная информация о всех элементах процесса:
- `id` (string): ID элемента
- `type` (string): Тип элемента
- `name` (string): Название элемента
- `task_definition` (string): Определение задачи (для service tasks)

### Version History
- `version` (integer): Номер версии
- `created_at` (string): Время создания версии
- `created_by` (string): Кто создал версию
- `changes` (string): Описание изменений

### Usage Statistics
- `total_instances` (integer): Общее количество запусков
- `active_instances` (integer): Активные экземпляры
- `completed_instances` (integer): Завершенные экземпляры
- `success_rate_percent` (float): Процент успешных выполнений
- `avg_duration_seconds` (integer): Средняя длительность

### Performance Metrics
- `avg_processing_time_ms` (integer): Среднее время обработки
- `p50_processing_time_ms` (integer): Медианное время
- `p95_processing_time_ms` (integer): 95-й перцентиль
- `bottleneck_activities` (array): Узкие места процесса

## Использование

### Process Analysis
```javascript
async function analyzeProcess(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}`);
  const process = await response.json();
  
  const data = process.data;
  
  console.log(`Process: ${data.name}`);
  console.log(`Success rate: ${data.usage_stats.success_rate_percent}%`);
  console.log(`Average duration: ${data.usage_stats.avg_duration_seconds}s`);
  
  // Анализ производительности
  if (data.performance.p95_processing_time_ms > 60000) {
    console.warn('Process has high 95th percentile latency');
  }
  
  // Анализ узких мест
  if (data.performance.bottleneck_activities.length > 0) {
    console.log('Bottleneck activities:');
    data.performance.bottleneck_activities.forEach(activity => {
      console.log(`- ${activity.activity_id}: ${activity.avg_duration_ms}ms (${activity.percentage_of_total}%)`);
    });
  }
  
  return data;
}
```

### Version Comparison
```javascript
async function compareVersions(processId) {
  // Получаем все версии процесса
  const allVersions = await getAllVersions(processId);
  
  const comparison = allVersions.map(version => ({
    version: version.version,
    elements: version.elements.total,
    performance: version.performance.avg_processing_time_ms,
    success_rate: version.usage_stats.success_rate_percent
  }));
  
  console.table(comparison);
  return comparison;
}

async function getAllVersions(processId) {
  const response = await fetch(`/api/v1/bpmn/processes?name=${processId}`);
  const data = await response.json();
  
  // Получаем детали для каждой версии
  const versions = [];
  for (const process of data.data.processes) {
    const details = await fetch(`/api/v1/bpmn/processes/${process.process_key}`);
    versions.push(await details.json());
  }
  
  return versions.map(v => v.data);
}
```

### Dependency Analysis
```javascript
async function analyzeDependencies(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}`);
  const process = await response.json();
  
  const deps = process.data.dependencies;
  
  console.log('Process Dependencies:');
  console.log(`- Task types: ${deps.task_definitions.map(t => t.type).join(', ')}`);
  console.log(`- Messages: ${deps.message_definitions.map(m => m.name).join(', ')}`);
  console.log(`- Timers: ${deps.timer_definitions.length} timers`);
  
  // Проверка доступности воркеров
  for (const taskDef of deps.task_definitions) {
    const workers = await checkWorkerAvailability(taskDef.type);
    if (workers.length === 0) {
      console.warn(`No workers available for task type: ${taskDef.type}`);
    }
  }
  
  return deps;
}
```

### Performance Monitoring
```javascript
class ProcessMonitor {
  constructor(processKey) {
    this.processKey = processKey;
    this.baseline = null;
  }
  
  async setBaseline() {
    const response = await fetch(`/api/v1/bpmn/processes/${this.processKey}`);
    const process = await response.json();
    
    this.baseline = {
      avg_duration: process.data.usage_stats.avg_duration_seconds,
      success_rate: process.data.usage_stats.success_rate_percent,
      p95_latency: process.data.performance.p95_processing_time_ms
    };
  }
  
  async checkRegression() {
    const response = await fetch(`/api/v1/bpmn/processes/${this.processKey}`);
    const process = await response.json();
    
    const current = {
      avg_duration: process.data.usage_stats.avg_duration_seconds,
      success_rate: process.data.usage_stats.success_rate_percent,
      p95_latency: process.data.performance.p95_processing_time_ms
    };
    
    const regressions = [];
    
    if (current.avg_duration > this.baseline.avg_duration * 1.2) {
      regressions.push('Duration increased by >20%');
    }
    
    if (current.success_rate < this.baseline.success_rate - 5) {
      regressions.push('Success rate decreased by >5%');
    }
    
    if (current.p95_latency > this.baseline.p95_latency * 1.5) {
      regressions.push('P95 latency increased by >50%');
    }
    
    return regressions;
  }
}
```

### Documentation Generation
```javascript
async function generateProcessDocumentation(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}`);
  const process = await response.json();
  
  const data = process.data;
  
  const doc = `
# ${data.name}

## Overview
${data.description}

## Statistics
- **Total Executions**: ${data.usage_stats.total_instances}
- **Success Rate**: ${data.usage_stats.success_rate_percent}%
- **Average Duration**: ${data.usage_stats.avg_duration_seconds}s

## Elements
- **Tasks**: ${data.elements.service_tasks + data.elements.user_tasks}
- **Gateways**: ${data.elements.exclusive_gateways + data.elements.parallel_gateways}
- **Events**: ${data.elements.start_events + data.elements.end_events + data.elements.boundary_events}

## Dependencies
### Task Types
${data.dependencies.task_definitions.map(t => `- ${t.type} (timeout: ${t.timeout})`).join('\n')}

### Messages
${data.dependencies.message_definitions.map(m => `- ${m.name}`).join('\n')}

## Performance
- **P50**: ${data.performance.p50_processing_time_ms}ms
- **P95**: ${data.performance.p95_processing_time_ms}ms
- **P99**: ${data.performance.p99_processing_time_ms}ms
  `;
  
  return doc;
}
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes`](./list-processes.md) - Список всех процессов
- [`GET /api/v1/bpmn/processes/:key/json`](./get-process-json.md) - JSON представление процесса
- [`DELETE /api/v1/bpmn/processes/:id`](./delete-process.md) - Удаление процесса
- [`POST /api/v1/processes`](../processes/start-process.md) - Запуск экземпляра процесса
