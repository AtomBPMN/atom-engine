# GET /api/v1/processes/:id

## Описание
Получение текущего статуса экземпляра процесса с основной информацией о его выполнении.

## URL
```
GET /api/v1/processes/{instance_id}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `instance_id` (string): ID экземпляра процесса

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const instanceId = 'srv1-aB3dEf9hK2mN5pQ8uV';
const response = await fetch(`/api/v1/processes/${instanceId}`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processStatus = await response.json();
```

## Ответы

### 200 OK - Статус процесса
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_id": "order-processing",
    "process_key": "order-processing-v2-3",
    "version": 3,
    "status": "ACTIVE",
    "tenant_id": "production",
    "started_at": "2025-01-11T10:30:00.123Z",
    "updated_at": "2025-01-11T10:32:15.456Z",
    "current_activity": "process-payment",
    "variables": {
      "orderId": "ORD-12345",
      "customerId": "CUST-67890",
      "amount": 299.99,
      "paymentMethod": "credit",
      "status": "processing",
      "priority": "high",
      "items": [
        {
          "sku": "PROD-001",
          "quantity": 2,
          "price": 149.99
        }
      ]
    },
    "current_activities": [
      {
        "element_id": "process-payment",
        "element_type": "serviceTask",
        "element_name": "Process Payment",
        "tokens": 1,
        "started_at": "2025-01-11T10:31:30.789Z"
      }
    ],
    "completed_activities": [
      {
        "element_id": "validate-order",
        "element_type": "serviceTask",
        "element_name": "Validate Order",
        "completed_at": "2025-01-11T10:31:15.234Z",
        "duration_ms": 5000
      },
      {
        "element_id": "check-inventory",
        "element_type": "serviceTask", 
        "element_name": "Check Inventory",
        "completed_at": "2025-01-11T10:31:25.678Z",
        "duration_ms": 3000
      }
    ],
    "metrics": {
      "elapsed_time_seconds": 135,
      "completed_activities_count": 2,
      "active_activities_count": 1,
      "total_activities_count": 8,
      "progress_percent": 37.5
    }
  },
  "request_id": "req_1641998402000"
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
      "instance_id": "srv1-nonexistent123"
    }
  },
  "request_id": "req_1641998402001"
}
```

## Поля ответа

### Basic Information
- `instance_id` (string): ID экземпляра процесса
- `process_id` (string): ID определения процесса  
- `process_key` (string): Ключ процесса с версией
- `version` (integer): Версия процесса
- `status` (string): Текущий статус
- `tenant_id` (string): ID тенанта

### Timing Information
- `started_at` (string): Время запуска (ISO 8601 UTC)
- `updated_at` (string): Время последнего обновления
- `completed_at` (string, nullable): Время завершения
- `cancelled_at` (string, nullable): Время отмены

### Current State
- `current_activity` (string): Текущая активность
- `variables` (object): Переменные процесса
- `current_activities` (array): Активные элементы
- `completed_activities` (array): Завершенные элементы

### Progress Metrics
- `elapsed_time_seconds` (integer): Время выполнения
- `completed_activities_count` (integer): Количество завершенных активностей
- `active_activities_count` (integer): Количество активных активностей
- `progress_percent` (float): Процент выполнения

## Статусы процессов

### Возможные статусы
- `ACTIVE` - Процесс выполняется
- `COMPLETED` - Процесс успешно завершен
- `CANCELLED` - Процесс отменен

### Детали статусов
```yaml
ACTIVE:
  description: "Процесс активен и выполняется"
  characteristics:
    - Есть активные токены
    - Может выполнять операции
    - Переменные могут изменяться

COMPLETED:
  description: "Процесс успешно завершен"
  characteristics:
    - Достиг end event
    - Все токены завершены
    - Переменные зафиксированы

CANCELLED:
  description: "Процесс отменен"
  characteristics:
    - Принудительно остановлен
    - Активные токены отменены
    - Сохранено время отмены
```

## Использование

### Polling Status
```javascript
// Мониторинг статуса процесса
async function monitorProcessStatus(instanceId) {
  let status = 'ACTIVE';
  
  while (status === 'ACTIVE') {
    const response = await fetch(`/api/v1/processes/${instanceId}`);
    const data = await response.json();
    
    status = data.data.status;
    const progress = data.data.metrics.progress_percent;
    
    console.log(`Process ${instanceId}: ${status} (${progress}% complete)`);
    
    if (status === 'ACTIVE') {
      await new Promise(resolve => setTimeout(resolve, 5000)); // Ждем 5 секунд
    }
  }
  
  console.log(`Process finished with status: ${status}`);
  return status;
}
```

### Progress Tracking
```javascript
async function trackProcessProgress(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}`);
  const data = await response.json();
  
  const process = data.data;
  
  return {
    instanceId: process.instance_id,
    status: process.status,
    progress: {
      percent: process.metrics.progress_percent,
      completed: process.metrics.completed_activities_count,
      total: process.metrics.total_activities_count,
      current: process.current_activity
    },
    timing: {
      started: new Date(process.started_at),
      elapsed: process.metrics.elapsed_time_seconds,
      estimatedCompletion: estimateCompletion(process)
    }
  };
}

function estimateCompletion(process) {
  if (process.metrics.progress_percent === 0) return null;
  
  const elapsedSeconds = process.metrics.elapsed_time_seconds;
  const progressPercent = process.metrics.progress_percent;
  
  const estimatedTotalSeconds = (elapsedSeconds / progressPercent) * 100;
  const remainingSeconds = estimatedTotalSeconds - elapsedSeconds;
  
  return new Date(Date.now() + remainingSeconds * 1000);
}
```

### Activity Analysis
```javascript
async function analyzeProcessActivities(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}`);
  const data = await response.json();
  
  const process = data.data;
  
  const analysis = {
    currentActivities: process.current_activities.map(activity => ({
      name: activity.element_name,
      type: activity.element_type,
      waitingTime: Date.now() - new Date(activity.started_at).getTime()
    })),
    
    completedActivities: process.completed_activities.map(activity => ({
      name: activity.element_name,
      type: activity.element_type,
      duration: activity.duration_ms
    })),
    
    bottlenecks: process.completed_activities
      .filter(activity => activity.duration_ms > 10000) // > 10 секунд
      .sort((a, b) => b.duration_ms - a.duration_ms),
      
    averageDuration: process.completed_activities.length > 0 ?
      process.completed_activities.reduce((sum, activity) => sum + activity.duration_ms, 0) / 
      process.completed_activities.length : 0
  };
  
  return analysis;
}
```

### Variable Monitoring
```javascript
async function monitorProcessVariables(instanceId, variablesToWatch) {
  const response = await fetch(`/api/v1/processes/${instanceId}`);
  const data = await response.json();
  
  const variables = data.data.variables;
  const watched = {};
  
  variablesToWatch.forEach(varName => {
    if (variables.hasOwnProperty(varName)) {
      watched[varName] = {
        value: variables[varName],
        type: typeof variables[varName],
        lastUpdated: data.data.updated_at
      };
    }
  });
  
  return watched;
}

// Использование
const watchedVars = await monitorProcessVariables('srv1-abc123', ['status', 'amount', 'priority']);
console.log('Watched variables:', watchedVars);
```

### Status Dashboard
```javascript
class ProcessStatusDashboard {
  constructor() {
    this.processes = new Map();
  }
  
  async addProcess(instanceId) {
    const status = await this.getProcessStatus(instanceId);
    this.processes.set(instanceId, status);
    return status;
  }
  
  async refreshAll() {
    const updates = [];
    
    for (const [instanceId] of this.processes) {
      updates.push(this.getProcessStatus(instanceId));
    }
    
    const statuses = await Promise.all(updates);
    
    statuses.forEach((status, index) => {
      const instanceId = Array.from(this.processes.keys())[index];
      this.processes.set(instanceId, status);
    });
    
    return this.getSummary();
  }
  
  async getProcessStatus(instanceId) {
    const response = await fetch(`/api/v1/processes/${instanceId}`);
    return await response.json();
  }
  
  getSummary() {
    const summary = {
      total: this.processes.size,
      active: 0,
      completed: 0,
      cancelled: 0,
      avgProgress: 0
    };
    
    let totalProgress = 0;
    
    for (const [instanceId, statusData] of this.processes) {
      const status = statusData.data.status;
      const progress = statusData.data.metrics.progress_percent;
      
      summary[status.toLowerCase()]++;
      totalProgress += progress;
    }
    
    summary.avgProgress = this.processes.size > 0 ? totalProgress / this.processes.size : 0;
    
    return summary;
  }
  
  getSlowProcesses(thresholdSeconds = 300) {
    const slow = [];
    
    for (const [instanceId, statusData] of this.processes) {
      const process = statusData.data;
      
      if (process.status === 'ACTIVE' && 
          process.metrics.elapsed_time_seconds > thresholdSeconds) {
        slow.push({
          instanceId,
          elapsedTime: process.metrics.elapsed_time_seconds,
          currentActivity: process.current_activity,
          progress: process.metrics.progress_percent
        });
      }
    }
    
    return slow.sort((a, b) => b.elapsedTime - a.elapsedTime);
  }
}

// Использование
const dashboard = new ProcessStatusDashboard();
await dashboard.addProcess('srv1-abc123');
await dashboard.addProcess('srv1-def456');

const summary = await dashboard.refreshAll();
console.log('Dashboard summary:', summary);

const slowProcesses = dashboard.getSlowProcesses(600); // Процессы > 10 минут
console.log('Slow processes:', slowProcesses);
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/info`](./get-process-info.md) - Детальная информация о процессе
- [`GET /api/v1/processes/:id/tokens`](./get-process-tokens.md) - Токены процесса
- [`DELETE /api/v1/processes/:id`](./cancel-process.md) - Отмена процесса
- [`GET /api/v1/processes`](./list-processes.md) - Список всех процессов
