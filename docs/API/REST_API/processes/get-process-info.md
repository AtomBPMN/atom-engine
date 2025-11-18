# GET /api/v1/processes/:id/info

## Описание
Получение полной детальной информации об экземпляре процесса, включая историю выполнения, метрики производительности и диагностическую информацию.

## URL
```
GET /api/v1/processes/{instance_id}/info
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `instance_id` (string): ID экземпляра процесса

## Параметры запроса (Query Parameters)
- `include_history` (boolean): Включить историю выполнения (по умолчанию: `true`)
- `include_variables` (boolean): Включить переменные (по умолчанию: `true`)
- `include_incidents` (boolean): Включить инциденты (по умолчанию: `true`)

## Примеры запросов

### Полная информация
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/info" \
  -H "X-API-Key: your-api-key-here"
```

### Без истории
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/info?include_history=false" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const instanceId = 'srv1-aB3dEf9hK2mN5pQ8uV';
const response = await fetch(`/api/v1/processes/${instanceId}/info?include_history=true`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processInfo = await response.json();
```

## Ответы

### 200 OK - Детальная информация получена
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_definition": {
      "process_id": "order-processing",
      "process_key": "order-processing-v2-3",
      "version": 3,
      "name": "Order Processing Workflow",
      "description": "Complete order processing from validation to fulfillment",
      "tenant_id": "production"
    },
    "status": {
      "current_status": "ACTIVE",
      "started_at": "2025-01-11T10:30:00.123Z",
      "updated_at": "2025-01-11T10:32:15.456Z",
      "completed_at": null,
      "cancelled_at": null,
      "cancellation_reason": null
    },
    "execution_context": {
      "started_by": "api-key-order-service",
      "parent_process_id": null,
      "child_processes": [],
      "correlation_keys": {
        "orderId": "ORD-12345",
        "customerId": "CUST-67890"
      },
      "business_key": "order-ORD-12345"
    },
    "current_state": {
      "current_activities": [
        {
          "element_id": "process-payment",
          "element_type": "serviceTask",
          "element_name": "Process Payment",
          "tokens": 1,
          "started_at": "2025-01-11T10:31:30.789Z",
          "worker_id": "payment-worker-02",
          "job_key": "srv1-job-xyz789",
          "retries_left": 3,
          "timeout_at": "2025-01-11T10:33:30.789Z"
        }
      ],
      "waiting_activities": [],
      "suspended_activities": []
    },
    "variables": {
      "current_variables": {
        "orderId": "ORD-12345",
        "customerId": "CUST-67890",
        "amount": 299.99,
        "paymentMethod": "credit",
        "status": "processing",
        "priority": "high",
        "validationResult": {
          "isValid": true,
          "validatedAt": "2025-01-11T10:30:45.123Z",
          "validator": "validation-service-v2"
        },
        "inventoryCheck": {
          "available": true,
          "reserved": 2,
          "warehouse": "MAIN-01"
        }
      },
      "variable_history": [
        {
          "timestamp": "2025-01-11T10:30:00.123Z",
          "activity": "start_process",
          "changes": {
            "added": ["orderId", "customerId", "amount", "paymentMethod"],
            "modified": [],
            "removed": []
          }
        },
        {
          "timestamp": "2025-01-11T10:30:45.123Z", 
          "activity": "validate-order",
          "changes": {
            "added": ["validationResult"],
            "modified": ["status"],
            "removed": []
          }
        },
        {
          "timestamp": "2025-01-11T10:31:15.234Z",
          "activity": "check-inventory",
          "changes": {
            "added": ["inventoryCheck"],
            "modified": [],
            "removed": []
          }
        }
      ]
    },
    "execution_history": [
      {
        "activity_id": "start_order",
        "activity_type": "startEvent",
        "activity_name": "Order Received",
        "started_at": "2025-01-11T10:30:00.123Z",
        "completed_at": "2025-01-11T10:30:00.124Z",
        "duration_ms": 1,
        "status": "COMPLETED",
        "tokens_consumed": 0,
        "tokens_produced": 1
      },
      {
        "activity_id": "validate-order",
        "activity_type": "serviceTask",
        "activity_name": "Validate Order",
        "started_at": "2025-01-11T10:30:15.234Z",
        "completed_at": "2025-01-11T10:30:45.123Z",
        "duration_ms": 29889,
        "status": "COMPLETED",
        "worker_id": "validation-worker-01",
        "job_key": "srv1-job-abc123",
        "retries_used": 0,
        "tokens_consumed": 1,
        "tokens_produced": 1
      },
      {
        "activity_id": "check-inventory",
        "activity_type": "serviceTask",
        "activity_name": "Check Inventory", 
        "started_at": "2025-01-11T10:30:45.234Z",
        "completed_at": "2025-01-11T10:31:15.234Z",
        "duration_ms": 30000,
        "status": "COMPLETED",
        "worker_id": "inventory-worker-03",
        "job_key": "srv1-job-def456",
        "retries_used": 1,
        "tokens_consumed": 1,
        "tokens_produced": 1
      }
    ],
    "active_jobs": [
      {
        "job_key": "srv1-job-xyz789",
        "type": "payment-processor",
        "element_id": "process-payment",
        "worker_id": "payment-worker-02",
        "activated_at": "2025-01-11T10:31:30.789Z",
        "timeout_at": "2025-01-11T10:33:30.789Z",
        "retries": 3,
        "custom_headers": {
          "priority": "high",
          "payment_provider": "stripe"
        }
      }
    ],
    "active_timers": [
      {
        "timer_id": "srv1-timer-payment-timeout",
        "element_id": "payment-timeout-boundary",
        "timer_type": "boundary",
        "duration": "PT5M",
        "scheduled_at": "2025-01-11T10:36:30.789Z",
        "remaining_ms": 240000
      }
    ],
    "incidents": [
      {
        "incident_id": "srv1-incident-retry",
        "type": "JOB_FAILURE",
        "status": "RESOLVED",
        "element_id": "check-inventory",
        "job_key": "srv1-job-def456",
        "error_message": "Temporary connection timeout",
        "created_at": "2025-01-11T10:30:50.000Z",
        "resolved_at": "2025-01-11T10:31:00.000Z",
        "resolution": "RETRY",
        "retries_after_incident": 1
      }
    ],
    "performance_metrics": {
      "total_duration_ms": 135333,
      "active_duration_ms": 135333,
      "waiting_time_ms": 0,
      "avg_activity_duration_ms": 20296,
      "longest_activity": {
        "activity_id": "check-inventory",
        "duration_ms": 30000
      },
      "shortest_activity": {
        "activity_id": "start_order", 
        "duration_ms": 1
      },
      "throughput": {
        "activities_per_minute": 1.33,
        "completion_rate": 0.75
      }
    },
    "resource_usage": {
      "memory_usage_bytes": 2048576,
      "cpu_time_ms": 1250,
      "database_operations": {
        "reads": 15,
        "writes": 8,
        "total_latency_ms": 245
      },
      "network_calls": 4,
      "cache_hits": 12,
      "cache_misses": 3
    },
    "metadata": {
      "generated_at": "2025-01-11T10:32:15.456Z",
      "data_freshness_seconds": 0,
      "included_sections": ["history", "variables", "incidents"],
      "process_version_at_start": 3,
      "engine_version": "1.0.0"
    }
  },
  "request_id": "req_1641998402100"
}
```

## Поля ответа

### Process Definition
- Информация об определении процесса, которое выполняется
- ID, версия, название, описание

### Status Information
- Детальная информация о статусе и временных метках
- Причина отмены (если применимо)

### Execution Context
- Кто запустил процесс
- Родительские/дочерние процессы
- Ключи корреляции и бизнес-ключ

### Current State
- Активные, ожидающие и приостановленные активности
- Детали для каждой активности включая worker, job, timeouts

### Variables
- Текущие переменные процесса
- История изменений переменных по активностям

### Execution History
- Полная история выполненных активностей
- Детали производительности для каждой активности
- Информация о workers, jobs, retries

### Active Resources
- Активные задания (jobs)
- Активные таймеры
- Timeout информация

### Incidents
- История инцидентов
- Резолюции и повторы
- Связь с активностями и заданиями

### Performance Metrics
- Общие метрики производительности
- Самые быстрые/медленные активности
- Throughput метрики

### Resource Usage
- Использование памяти и CPU
- Database операции
- Network вызовы и кэширование

## Использование

### Detailed Process Analysis
```javascript
async function analyzeProcessPerformance(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}/info`);
  const data = await response.json();
  
  const process = data.data;
  
  const analysis = {
    // Общая производительность
    totalDuration: process.performance_metrics.total_duration_ms,
    avgActivityDuration: process.performance_metrics.avg_activity_duration_ms,
    
    // Анализ активностей
    activities: process.execution_history.map(activity => ({
      name: activity.activity_name,
      type: activity.activity_type,
      duration: activity.duration_ms,
      retries: activity.retries_used || 0,
      efficiency: activity.duration_ms < process.performance_metrics.avg_activity_duration_ms ? 'good' : 'poor'
    })),
    
    // Узкие места
    bottlenecks: process.execution_history
      .filter(activity => activity.duration_ms > process.performance_metrics.avg_activity_duration_ms * 2)
      .sort((a, b) => b.duration_ms - a.duration_ms),
    
    // Проблемы
    issues: {
      incidents: process.incidents.length,
      activeRetries: process.active_jobs.filter(job => job.retries < 3).length,
      longRunningActivities: process.current_state.current_activities.filter(activity => {
        const startTime = new Date(activity.started_at).getTime();
        return Date.now() - startTime > 300000; // > 5 минут
      }).length
    },
    
    // Использование ресурсов
    resourceEfficiency: {
      memoryUsage: process.resource_usage.memory_usage_bytes,
      cacheHitRatio: process.resource_usage.cache_hits / 
                     (process.resource_usage.cache_hits + process.resource_usage.cache_misses),
      dbEfficiency: process.resource_usage.database_operations.total_latency_ms / 
                    process.resource_usage.database_operations.reads
    }
  };
  
  return analysis;
}
```

### Process Debugging
```javascript
async function debugProcess(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}/info?include_incidents=true`);
  const data = await response.json();
  
  const process = data.data;
  
  const debug = {
    currentIssues: [],
    recommendations: [],
    healthScore: 100
  };
  
  // Проверка активных проблем
  if (process.incidents.filter(i => i.status === 'OPEN').length > 0) {
    debug.currentIssues.push('Has open incidents');
    debug.healthScore -= 30;
  }
  
  // Проверка долго выполняющихся активностей
  process.current_state.current_activities.forEach(activity => {
    const startTime = new Date(activity.started_at).getTime();
    const runningTime = Date.now() - startTime;
    
    if (runningTime > 600000) { // > 10 минут
      debug.currentIssues.push(`Activity ${activity.element_name} running for ${Math.round(runningTime/60000)} minutes`);
      debug.healthScore -= 20;
    }
  });
  
  // Проверка использования retry
  const highRetryJobs = process.active_jobs.filter(job => job.retries < 2);
  if (highRetryJobs.length > 0) {
    debug.currentIssues.push(`${highRetryJobs.length} jobs with high retry usage`);
    debug.healthScore -= 15;
  }
  
  // Рекомендации
  if (process.performance_metrics.avg_activity_duration_ms > 30000) {
    debug.recommendations.push('Consider optimizing activity performance');
  }
  
  if (process.resource_usage.cache_hits / (process.resource_usage.cache_hits + process.resource_usage.cache_misses) < 0.8) {
    debug.recommendations.push('Low cache hit ratio - consider cache optimization');
  }
  
  return debug;
}
```

### Variable Change Tracking
```javascript
async function trackVariableChanges(instanceId, variableName) {
  const response = await fetch(`/api/v1/processes/${instanceId}/info?include_variables=true`);
  const data = await response.json();
  
  const variableHistory = data.data.variables.variable_history;
  const changes = [];
  
  variableHistory.forEach(entry => {
    if (entry.changes.added && entry.changes.added.includes(variableName)) {
      changes.push({
        type: 'ADDED',
        timestamp: entry.timestamp,
        activity: entry.activity,
        value: getCurrentVariableValue(data.data.variables.current_variables, variableName, entry.timestamp)
      });
    }
    
    if (entry.changes.modified && entry.changes.modified.includes(variableName)) {
      changes.push({
        type: 'MODIFIED',
        timestamp: entry.timestamp,
        activity: entry.activity,
        value: getCurrentVariableValue(data.data.variables.current_variables, variableName, entry.timestamp)
      });
    }
    
    if (entry.changes.removed && entry.changes.removed.includes(variableName)) {
      changes.push({
        type: 'REMOVED',
        timestamp: entry.timestamp,
        activity: entry.activity,
        value: null
      });
    }
  });
  
  return changes;
}

function getCurrentVariableValue(variables, variableName, timestamp) {
  // В реальной реализации здесь была бы логика получения значения на момент timestamp
  return variables[variableName];
}
```

### Process Health Dashboard
```javascript
class ProcessHealthDashboard {
  async getProcessHealth(instanceId) {
    const response = await fetch(`/api/v1/processes/${instanceId}/info`);
    const data = await response.json();
    
    const process = data.data;
    
    return {
      instanceId,
      status: process.status.current_status,
      health: this.calculateHealthScore(process),
      performance: this.analyzePerformance(process),
      issues: this.identifyIssues(process),
      recommendations: this.generateRecommendations(process)
    };
  }
  
  calculateHealthScore(process) {
    let score = 100;
    
    // Вычеты за проблемы
    score -= process.incidents.filter(i => i.status === 'OPEN').length * 20;
    score -= process.current_state.current_activities.filter(a => 
      Date.now() - new Date(a.started_at).getTime() > 600000
    ).length * 15;
    
    // Вычеты за плохую производительность
    if (process.performance_metrics.avg_activity_duration_ms > 60000) {
      score -= 25;
    }
    
    return Math.max(0, score);
  }
  
  analyzePerformance(process) {
    const metrics = process.performance_metrics;
    
    return {
      duration: metrics.total_duration_ms,
      avgActivity: metrics.avg_activity_duration_ms,
      throughput: metrics.throughput.activities_per_minute,
      efficiency: metrics.throughput.completion_rate,
      bottleneck: metrics.longest_activity
    };
  }
  
  identifyIssues(process) {
    const issues = [];
    
    // Открытые инциденты
    const openIncidents = process.incidents.filter(i => i.status === 'OPEN');
    if (openIncidents.length > 0) {
      issues.push({
        type: 'INCIDENT',
        severity: 'HIGH',
        count: openIncidents.length,
        description: `${openIncidents.length} open incidents`
      });
    }
    
    // Долго выполняющиеся активности
    const longRunning = process.current_state.current_activities.filter(a =>
      Date.now() - new Date(a.started_at).getTime() > 600000
    );
    
    if (longRunning.length > 0) {
      issues.push({
        type: 'PERFORMANCE',
        severity: 'MEDIUM',
        count: longRunning.length,
        description: `${longRunning.length} long-running activities`
      });
    }
    
    return issues;
  }
  
  generateRecommendations(process) {
    const recommendations = [];
    
    if (process.performance_metrics.avg_activity_duration_ms > 30000) {
      recommendations.push('Optimize activity performance');
    }
    
    if (process.resource_usage.cache_hits / (process.resource_usage.cache_hits + process.resource_usage.cache_misses) < 0.8) {
      recommendations.push('Improve cache utilization');
    }
    
    if (process.incidents.length > 5) {
      recommendations.push('Review and address recurring incidents');
    }
    
    return recommendations;
  }
}
```

## Связанные endpoints
- [`GET /api/v1/processes/:id`](./get-process-status.md) - Базовый статус процесса
- [`GET /api/v1/processes/:id/tokens`](./get-process-tokens.md) - Информация о токенах
- [`GET /api/v1/processes/:id/tokens/trace`](./get-token-trace.md) - Трассировка выполнения
- [`GET /api/v1/incidents`](../incidents/list-incidents.md) - Детали инцидентов
