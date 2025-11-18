# GET /api/v1/processes/:id/tokens/trace

## Описание
Получение полной трассировки выполнения токенов для экземпляра процесса. Показывает путь выполнения, ветвления, объединения и временные характеристики.

## URL
```
GET /api/v1/processes/{instance_id}/tokens/trace
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `instance_id` (string): ID экземпляра процесса

## Параметры запроса (Query Parameters)
- `format` (string): Формат вывода (`detailed`, `summary`, по умолчанию: `detailed`)
- `include_variables` (boolean): Включить изменения переменных (по умолчанию: `false`)
- `start_time` (string): Начало периода трассировки (ISO 8601)
- `end_time` (string): Конец периода трассировки (ISO 8601)

## Примеры запросов

### Полная трассировка
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/tokens/trace" \
  -H "X-API-Key: your-api-key-here"
```

### Краткая трассировка
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/tokens/trace?format=summary" \
  -H "X-API-Key: your-api-key-here"
```

### С переменными
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/tokens/trace?include_variables=true" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const instanceId = 'srv1-aB3dEf9hK2mN5pQ8uV';
const response = await fetch(`/api/v1/processes/${instanceId}/tokens/trace?format=detailed`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const trace = await response.json();
```

## Ответы

### 200 OK - Трассировка получена (Detailed Format)
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_id": "order-processing",
    "trace_generated_at": "2025-01-11T10:32:15.456Z",
    "execution_timeline": [
      {
        "sequence": 1,
        "timestamp": "2025-01-11T10:30:00.123Z",
        "action": "TOKEN_CREATED",
        "token_id": "srv1-token-start001",
        "element_id": "start_order",
        "element_type": "startEvent",
        "element_name": "Order Received",
        "duration_ms": 0,
        "variables_snapshot": {
          "orderId": "ORD-12345",
          "customerId": "CUST-67890",
          "amount": 299.99
        },
        "flow_details": {
          "incoming_flow": null,
          "outgoing_flows": ["flow_1"]
        }
      },
      {
        "sequence": 2,
        "timestamp": "2025-01-11T10:30:00.124Z",
        "action": "TOKEN_MOVED",
        "token_id": "srv1-token-start001",
        "from_element": "start_order",
        "to_element": "validate-order",
        "via_flow": "flow_1",
        "duration_ms": 1,
        "flow_details": {
          "condition": null,
          "probability": 1.0
        }
      },
      {
        "sequence": 3,
        "timestamp": "2025-01-11T10:30:15.234Z",
        "action": "TOKEN_ACTIVATED",
        "token_id": "srv1-token-validate01",
        "element_id": "validate-order",
        "element_type": "serviceTask",
        "element_name": "Validate Order",
        "job_details": {
          "job_key": "srv1-job-abc123",
          "worker_type": "order-validator",
          "worker_id": "validation-worker-01",
          "timeout": "PT30S",
          "retries": 3
        }
      },
      {
        "sequence": 4,
        "timestamp": "2025-01-11T10:30:45.123Z",
        "action": "TOKEN_COMPLETED",
        "token_id": "srv1-token-validate01",
        "element_id": "validate-order",
        "duration_ms": 29889,
        "result": "SUCCESS",
        "variables_changed": {
          "added": ["validationResult"],
          "modified": ["status"],
          "removed": []
        },
        "execution_details": {
          "retries_used": 0,
          "worker_response_time_ms": 28500,
          "system_overhead_ms": 1389
        }
      },
      {
        "sequence": 5,
        "timestamp": "2025-01-11T10:31:25.678Z",
        "action": "TOKEN_ACTIVATED",
        "token_id": "srv1-token-gateway01",
        "element_id": "payment-gateway",
        "element_type": "exclusiveGateway",
        "element_name": "Payment Method?",
        "gateway_details": {
          "conditions_to_evaluate": [
            {
              "flow_id": "flow_3_credit",
              "condition": "paymentMethod = 'credit'"
            },
            {
              "flow_id": "flow_3_debit",
              "condition": "paymentMethod = 'debit'"
            }
          ]
        }
      },
      {
        "sequence": 6,
        "timestamp": "2025-01-11T10:31:30.789Z",
        "action": "GATEWAY_EVALUATED",
        "token_id": "srv1-token-gateway01",
        "element_id": "payment-gateway",
        "evaluation_result": {
          "selected_flow": "flow_3_credit",
          "condition_matched": "paymentMethod = 'credit'",
          "evaluation_time_ms": 5,
          "variable_values": {
            "paymentMethod": "credit"
          }
        }
      },
      {
        "sequence": 7,
        "timestamp": "2025-01-11T10:31:30.794Z",
        "action": "TOKEN_SPLIT",
        "parent_token_id": "srv1-token-gateway01",
        "child_token_id": "srv1-token-payment01",
        "to_element": "process-payment",
        "via_flow": "flow_3_credit",
        "split_reason": "GATEWAY_DECISION"
      },
      {
        "sequence": 8,
        "timestamp": "2025-01-11T10:31:30.795Z",
        "action": "TOKEN_ACTIVATED",
        "token_id": "srv1-token-payment01",
        "element_id": "process-payment",
        "element_type": "serviceTask",
        "element_name": "Process Payment",
        "job_details": {
          "job_key": "srv1-job-xyz789",
          "worker_type": "payment-processor",
          "worker_id": "payment-worker-02",
          "timeout": "PT2M",
          "retries": 3
        },
        "current_status": "ACTIVE"
      }
    ],
    "execution_paths": [
      {
        "path_id": "main-path-01",
        "start_element": "start_order",
        "current_element": "process-payment",
        "elements_traversed": [
          "start_order",
          "validate-order", 
          "payment-gateway",
          "process-payment"
        ],
        "flows_taken": [
          "flow_1",
          "flow_2",
          "flow_3_credit"
        ],
        "total_duration_ms": 90671,
        "status": "ACTIVE",
        "progress_percent": 50.0
      }
    ],
    "token_relationships": {
      "token_tree": [
        {
          "token_id": "srv1-token-start001",
          "element": "start_order",
          "state": "COMPLETED",
          "children": [
            {
              "token_id": "srv1-token-validate01",
              "element": "validate-order",
              "state": "COMPLETED",
              "children": [
                {
                  "token_id": "srv1-token-gateway01",
                  "element": "payment-gateway",
                  "state": "COMPLETED",
                  "children": [
                    {
                      "token_id": "srv1-token-payment01",
                      "element": "process-payment",
                      "state": "ACTIVE",
                      "children": []
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    },
    "performance_analysis": {
      "total_execution_time_ms": 90671,
      "time_breakdown": {
        "active_processing_ms": 29889,
        "gateway_evaluation_ms": 5,
        "system_overhead_ms": 1777,
        "waiting_time_ms": 59000
      },
      "bottlenecks": [
        {
          "element_id": "validate-order",
          "duration_ms": 29889,
          "percentage_of_total": 33.0,
          "bottleneck_type": "PROCESSING_TIME"
        }
      ],
      "parallel_execution": {
        "max_concurrent_tokens": 1,
        "parallelism_opportunities": [
          {
            "element": "payment-gateway",
            "note": "Could be optimized for parallel payment processing"
          }
        ]
      }
    },
    "variable_changes": [
      {
        "timestamp": "2025-01-11T10:30:00.123Z",
        "element": "start_order",
        "action": "PROCESS_STARTED",
        "variables": {
          "orderId": "ORD-12345",
          "customerId": "CUST-67890",
          "amount": 299.99,
          "paymentMethod": "credit"
        }
      },
      {
        "timestamp": "2025-01-11T10:30:45.123Z",
        "element": "validate-order",
        "action": "TASK_COMPLETED",
        "variables_added": {
          "validationResult": {
            "isValid": true,
            "validatedAt": "2025-01-11T10:30:45.123Z"
          }
        },
        "variables_modified": {
          "status": {
            "from": "new",
            "to": "validated"
          }
        }
      }
    ],
    "summary": {
      "total_tokens_created": 4,
      "total_tokens_completed": 3,
      "total_tokens_active": 1,
      "total_flows_traversed": 3,
      "total_elements_visited": 4,
      "execution_status": "IN_PROGRESS",
      "estimated_completion_time": "2025-01-11T10:34:00.000Z"
    }
  },
  "request_id": "req_1641998402300"
}
```

### 200 OK - Краткая трассировка (Summary Format)
```json
{
  "success": true,
  "data": {
    "instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "execution_path": [
      {
        "element": "start_order",
        "type": "startEvent",
        "status": "COMPLETED",
        "duration_ms": 1,
        "timestamp": "2025-01-11T10:30:00.123Z"
      },
      {
        "element": "validate-order",
        "type": "serviceTask",
        "status": "COMPLETED",
        "duration_ms": 29889,
        "timestamp": "2025-01-11T10:30:15.234Z"
      },
      {
        "element": "payment-gateway",
        "type": "exclusiveGateway",
        "status": "COMPLETED",
        "duration_ms": 5,
        "decision": "flow_3_credit",
        "timestamp": "2025-01-11T10:31:25.678Z"
      },
      {
        "element": "process-payment",
        "type": "serviceTask",
        "status": "ACTIVE",
        "duration_ms": null,
        "timestamp": "2025-01-11T10:31:30.795Z"
      }
    ],
    "current_position": "process-payment",
    "progress_percent": 50.0,
    "total_duration_ms": 90671,
    "estimated_remaining_ms": 90000
  }
}
```

## Использование

### Process Visualization
```javascript
async function visualizeProcessExecution(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}/tokens/trace?format=detailed`);
  const data = await response.json();
  
  const trace = data.data;
  
  // Создание данных для визуализации
  const visualization = {
    nodes: [],
    edges: [],
    timeline: []
  };
  
  // Обработка элементов процесса
  const visitedElements = new Set();
  
  trace.execution_timeline.forEach(entry => {
    if (entry.element_id && !visitedElements.has(entry.element_id)) {
      visualization.nodes.push({
        id: entry.element_id,
        label: entry.element_name,
        type: entry.element_type,
        status: getElementStatus(entry.element_id, trace),
        duration: getElementDuration(entry.element_id, trace)
      });
      visitedElements.add(entry.element_id);
    }
  });
  
  // Обработка потоков
  trace.execution_timeline
    .filter(entry => entry.action === 'TOKEN_MOVED')
    .forEach(entry => {
      visualization.edges.push({
        id: entry.via_flow,
        source: entry.from_element,
        target: entry.to_element,
        timestamp: entry.timestamp
      });
    });
  
  // Создание временной шкалы
  visualization.timeline = trace.execution_timeline.map(entry => ({
    time: entry.timestamp,
    event: entry.action,
    element: entry.element_name || entry.to_element,
    duration: entry.duration_ms
  }));
  
  return visualization;
}

function getElementStatus(elementId, trace) {
  const completedEntry = trace.execution_timeline.find(
    entry => entry.element_id === elementId && entry.action === 'TOKEN_COMPLETED'
  );
  
  if (completedEntry) return 'COMPLETED';
  
  const activeEntry = trace.execution_timeline.find(
    entry => entry.element_id === elementId && entry.current_status === 'ACTIVE'
  );
  
  return activeEntry ? 'ACTIVE' : 'PENDING';
}
```

### Performance Analysis
```javascript
async function analyzeExecutionPerformance(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}/tokens/trace?format=detailed`);
  const data = await response.json();
  
  const trace = data.data;
  
  const analysis = {
    bottlenecks: trace.performance_analysis.bottlenecks,
    timeDistribution: analyzeTimeDistribution(trace),
    pathEfficiency: analyzePathEfficiency(trace),
    recommendations: generateOptimizationRecommendations(trace)
  };
  
  return analysis;
}

function analyzeTimeDistribution(trace) {
  const breakdown = trace.performance_analysis.time_breakdown;
  const total = trace.performance_analysis.total_execution_time_ms;
  
  return {
    processing: {
      time_ms: breakdown.active_processing_ms,
      percentage: (breakdown.active_processing_ms / total) * 100
    },
    waiting: {
      time_ms: breakdown.waiting_time_ms,
      percentage: (breakdown.waiting_time_ms / total) * 100
    },
    overhead: {
      time_ms: breakdown.system_overhead_ms,
      percentage: (breakdown.system_overhead_ms / total) * 100
    }
  };
}

function generateOptimizationRecommendations(trace) {
  const recommendations = [];
  
  // Анализ узких мест
  trace.performance_analysis.bottlenecks.forEach(bottleneck => {
    if (bottleneck.percentage_of_total > 30) {
      recommendations.push({
        type: 'PERFORMANCE',
        priority: 'HIGH',
        element: bottleneck.element_id,
        message: `Element consumes ${bottleneck.percentage_of_total}% of total execution time`,
        suggestion: 'Consider optimizing this task or adding parallel processing'
      });
    }
  });
  
  // Анализ ожидания
  const timeBreakdown = trace.performance_analysis.time_breakdown;
  if (timeBreakdown.waiting_time_ms > timeBreakdown.active_processing_ms) {
    recommendations.push({
      type: 'EFFICIENCY',
      priority: 'MEDIUM',
      message: 'High waiting time detected',
      suggestion: 'Review worker availability and job queue processing'
    });
  }
  
  return recommendations;
}
```

### Path Comparison
```javascript
async function compareExecutionPaths(instanceIds) {
  const traces = await Promise.all(
    instanceIds.map(async id => {
      const response = await fetch(`/api/v1/processes/${id}/tokens/trace?format=summary`);
      const data = await response.json();
      return { instanceId: id, trace: data.data };
    })
  );
  
  const comparison = {
    commonPath: findCommonPath(traces),
    deviations: findPathDeviations(traces),
    performanceComparison: comparePerformance(traces)
  };
  
  return comparison;
}

function findCommonPath(traces) {
  if (traces.length === 0) return [];
  
  const firstPath = traces[0].trace.execution_path.map(e => e.element);
  
  return firstPath.filter(element => 
    traces.every(trace => 
      trace.trace.execution_path.some(e => e.element === element)
    )
  );
}

function findPathDeviations(traces) {
  const deviations = [];
  
  traces.forEach((trace, index) => {
    const otherTraces = traces.filter((_, i) => i !== index);
    const uniqueElements = trace.trace.execution_path
      .map(e => e.element)
      .filter(element => 
        !otherTraces.every(other => 
          other.trace.execution_path.some(e => e.element === element)
        )
      );
    
    if (uniqueElements.length > 0) {
      deviations.push({
        instanceId: trace.instanceId,
        uniqueElements,
        reason: 'Different execution path'
      });
    }
  });
  
  return deviations;
}
```

### Debug Helper
```javascript
async function debugProcessExecution(instanceId) {
  const response = await fetch(`/api/v1/processes/${instanceId}/tokens/trace?include_variables=true`);
  const data = await response.json();
  
  const trace = data.data;
  
  const debug = {
    executionIssues: [],
    performanceIssues: [],
    dataIssues: []
  };
  
  // Поиск проблем выполнения
  const activeTokens = trace.execution_timeline.filter(
    entry => entry.current_status === 'ACTIVE'
  );
  
  activeTokens.forEach(token => {
    const startTime = new Date(token.timestamp).getTime();
    const runningTime = Date.now() - startTime;
    
    if (runningTime > 600000) { // > 10 минут
      debug.executionIssues.push({
        type: 'LONG_RUNNING_TOKEN',
        token_id: token.token_id,
        element: token.element_name,
        running_time_minutes: Math.round(runningTime / 60000)
      });
    }
  });
  
  // Поиск проблем производительности
  trace.performance_analysis.bottlenecks.forEach(bottleneck => {
    if (bottleneck.percentage_of_total > 50) {
      debug.performanceIssues.push({
        type: 'MAJOR_BOTTLENECK',
        element: bottleneck.element_id,
        impact_percent: bottleneck.percentage_of_total
      });
    }
  });
  
  // Поиск проблем с данными
  if (trace.variable_changes) {
    const suspiciousChanges = trace.variable_changes.filter(change => {
      return change.variables_modified && 
             Object.keys(change.variables_modified).length > 10; // Много изменений за раз
    });
    
    if (suspiciousChanges.length > 0) {
      debug.dataIssues.push({
        type: 'BULK_VARIABLE_CHANGES',
        occurrences: suspiciousChanges.length
      });
    }
  }
  
  return debug;
}
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/tokens`](./get-process-tokens.md) - Информация о токенах
- [`GET /api/v1/processes/:id`](./get-process-status.md) - Статус процесса
- [`GET /api/v1/processes/:id/info`](./get-process-info.md) - Детальная информация о процессе
