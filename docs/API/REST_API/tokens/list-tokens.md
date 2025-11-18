# GET /api/v1/tokens

## Описание
Получение списка токенов выполнения BPMN процессов с фильтрацией и пагинацией. Токены представляют текущее состояние выполнения процесса.

## URL
```
GET /api/v1/tokens
```

## Авторизация
✅ **Требуется API ключ** с разрешением `token`

## Параметры запроса (Query Parameters)

### Фильтрация
- `process_instance_id` (string): ID экземпляра процесса
- `state` (string): Состояние токена (`active`, `completed`, `cancelled`)
- `element_type` (string): Тип BPMN элемента (`start_event`, `service_task`, `user_task`, `gateway`, `end_event`)
- `tenant_id` (string): ID тенанта

### Пагинация
- `limit` (integer): Количество записей (по умолчанию: 10, максимум: 100)
- `offset` (integer): Смещение для пагинации (по умолчанию: 0)

## Примеры запросов

### Все токены
```bash
curl -X GET "http://localhost:27555/api/v1/tokens" \
  -H "X-API-Key: your-api-key-here"
```

### Активные токены процесса
```bash
curl -X GET "http://localhost:27555/api/v1/tokens?process_instance_id=srv1-aB3dEf9hK2mN5pQ8uV&state=active" \
  -H "X-API-Key: your-api-key-here"
```

### Токены по типу элемента
```bash
curl -X GET "http://localhost:27555/api/v1/tokens?element_type=service_task&limit=50" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/tokens?state=active&limit=20', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const tokens = await response.json();
```

## Ответы

### 200 OK - Список токенов
```json
{
  "success": true,
  "data": {
    "tokens": [
      {
        "id": "srv1-token123abc456def789",
        "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "process_definition_key": "Order_Processing",
        "process_definition_version": 3,
        "element_id": "ServiceTask_PaymentProcess",
        "element_name": "Process Payment",
        "element_type": "SERVICE_TASK",
        "state": "ACTIVE",
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:15:30.123Z",
        "variables": {
          "orderId": "ORD-2025-001234",
          "amount": 299.99,
          "currency": "USD",
          "customerId": "CUST-789456"
        },
        "execution_context": {
          "current_step": "ServiceTask_PaymentProcess",
          "previous_step": "ServiceTask_ValidateOrder",
          "next_possible_steps": ["ExclusiveGateway_PaymentResult"],
          "path_taken": ["StartEvent_1", "ServiceTask_ValidateOrder", "ServiceTask_PaymentProcess"],
          "execution_started_at": "2025-01-11T10:10:15.123Z",
          "waiting_for": "job_completion"
        },
        "tenant_id": "tenant-001",
        "parent_token_id": null,
        "child_tokens": [],
        "is_root_token": true
      },
      {
        "id": "srv1-token456def789ghi012",
        "process_instance_id": "srv1-cD6fGh9iJ3kL7mN0pR",
        "process_definition_key": "Invoice_Approval",
        "process_definition_version": 2,
        "element_id": "UserTask_ManagerReview",
        "element_name": "Manager Review",
        "element_type": "USER_TASK",
        "state": "ACTIVE",
        "created_at": "2025-01-11T09:30:45.456Z",
        "updated_at": "2025-01-11T10:00:12.789Z",
        "variables": {
          "invoiceId": "INV-2025-005678",
          "amount": 1500.00,
          "submittedBy": "john.doe@company.com",
          "category": "office_supplies"
        },
        "execution_context": {
          "current_step": "UserTask_ManagerReview",
          "previous_step": "ExclusiveGateway_AmountCheck",
          "next_possible_steps": ["ExclusiveGateway_ApprovalResult"],
          "path_taken": ["StartEvent_1", "UserTask_SubmitInvoice", "ExclusiveGateway_AmountCheck", "UserTask_ManagerReview"],
          "execution_started_at": "2025-01-11T09:25:30.456Z",
          "waiting_for": "user_task_completion",
          "assigned_user": "manager@company.com",
          "due_date": "2025-01-13T17:00:00.000Z"
        },
        "tenant_id": "tenant-001",
        "parent_token_id": null,
        "child_tokens": [],
        "is_root_token": true
      },
      {
        "id": "srv1-token789ghi012jkl345",
        "process_instance_id": "srv1-eF8gHi1jK4lM7nO9pS",
        "process_definition_key": "Parallel_Processing",
        "process_definition_version": 1,
        "element_id": "ParallelGateway_Split",
        "element_name": "Split Tasks",
        "element_type": "PARALLEL_GATEWAY",
        "state": "COMPLETED",
        "created_at": "2025-01-11T08:45:20.789Z",
        "updated_at": "2025-01-11T08:45:21.123Z",
        "completed_at": "2025-01-11T08:45:21.123Z",
        "variables": {
          "requestId": "REQ-2025-003456",
          "priority": "HIGH",
          "branches": ["task_a", "task_b", "task_c"]
        },
        "execution_context": {
          "current_step": "ParallelGateway_Split",
          "previous_step": "StartEvent_1",
          "execution_started_at": "2025-01-11T08:45:20.789Z",
          "execution_completed_at": "2025-01-11T08:45:21.123Z",
          "branches_created": 3
        },
        "tenant_id": "tenant-001",
        "parent_token_id": null,
        "child_tokens": [
          "srv1-token890jkl345mno678",
          "srv1-token901kml456nop789",
          "srv1-token012lmn567opq890"
        ],
        "is_root_token": true
      }
    ],
    "pagination": {
      "total": 156,
      "limit": 10,
      "offset": 0,
      "has_next": true,
      "has_previous": false
    },
    "summary": {
      "total_tokens": 156,
      "active_tokens": 89,
      "completed_tokens": 62,
      "cancelled_tokens": 5,
      "by_element_type": {
        "SERVICE_TASK": 45,
        "USER_TASK": 28,
        "PARALLEL_GATEWAY": 23,
        "EXCLUSIVE_GATEWAY": 19,
        "START_EVENT": 15,
        "END_EVENT": 12,
        "INTERMEDIATE_EVENT": 8,
        "SCRIPT_TASK": 6
      },
      "by_process": {
        "Order_Processing": 67,
        "Invoice_Approval": 34,
        "Customer_Onboarding": 28,
        "Parallel_Processing": 27
      }
    }
  },
  "request_id": "req_1641998404800"
}
```

## Token States

### ACTIVE
Токен активен и ожидает выполнения или завершения текущего элемента

### COMPLETED
Токен завершил выполнение элемента и передан дальше

### CANCELLED
Токен был отменен (например, при отмене процесса или граничном событии)

## Использование

### Token Monitor
```javascript
class TokenMonitor {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getTokens(filters = {}) {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        params.append(key, value);
      }
    });
    
    const response = await fetch(`/api/v1/tokens?${params}`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  async getActiveTokens(processInstanceId = null) {
    const filters = { state: 'active', limit: 100 };
    if (processInstanceId) {
      filters.process_instance_id = processInstanceId;
    }
    
    return await this.getTokens(filters);
  }
  
  async getProcessTokens(processInstanceId) {
    return await this.getTokens({
      process_instance_id: processInstanceId,
      limit: 100
    });
  }
  
  async getTokensByElementType(elementType) {
    return await this.getTokens({
      element_type: elementType,
      limit: 100
    });
  }
  
  async analyzeTokenDistribution() {
    const result = await this.getTokens({ limit: 1000 });
    const { summary, tokens } = result.data;
    
    return {
      state_distribution: {
        active: summary.active_tokens,
        completed: summary.completed_tokens,
        cancelled: summary.cancelled_tokens,
        total: summary.total_tokens
      },
      element_type_distribution: summary.by_element_type,
      process_distribution: summary.by_process,
      average_execution_time: this.calculateAverageExecutionTime(tokens),
      longest_running_tokens: this.findLongestRunningTokens(tokens),
      bottlenecks: this.identifyBottlenecks(summary.by_element_type)
    };
  }
  
  calculateAverageExecutionTime(tokens) {
    const completedTokens = tokens.filter(token => 
      token.state === 'COMPLETED' && token.completed_at
    );
    
    if (completedTokens.length === 0) return 0;
    
    const totalTime = completedTokens.reduce((sum, token) => {
      const start = new Date(token.created_at);
      const end = new Date(token.completed_at);
      return sum + (end - start);
    }, 0);
    
    return totalTime / completedTokens.length / 1000; // seconds
  }
  
  findLongestRunningTokens(tokens, limit = 10) {
    const activeTokens = tokens.filter(token => token.state === 'ACTIVE');
    
    return activeTokens
      .map(token => ({
        ...token,
        running_time_seconds: (Date.now() - new Date(token.created_at)) / 1000
      }))
      .sort((a, b) => b.running_time_seconds - a.running_time_seconds)
      .slice(0, limit);
  }
  
  identifyBottlenecks(elementTypeDistribution) {
    const total = Object.values(elementTypeDistribution).reduce((sum, count) => sum + count, 0);
    
    return Object.entries(elementTypeDistribution)
      .map(([type, count]) => ({
        element_type: type,
        count,
        percentage: (count / total) * 100
      }))
      .filter(item => item.percentage > 20) // Consider bottleneck if >20% of tokens
      .sort((a, b) => b.percentage - a.percentage);
  }
  
  async startRealTimeMonitoring(callback, interval = 5000) {
    const monitor = async () => {
      try {
        const analysis = await this.analyzeTokenDistribution();
        callback(analysis);
      } catch (error) {
        console.error('Token monitoring error:', error);
      }
    };
    
    // Initial call
    await monitor();
    
    // Set up interval
    return setInterval(monitor, interval);
  }
  
  async generateTokenReport(processInstanceId = null) {
    const filters = processInstanceId ? { process_instance_id: processInstanceId } : {};
    const result = await this.getTokens({ ...filters, limit: 1000 });
    
    const report = {
      generated_at: new Date().toISOString(),
      scope: processInstanceId ? `Process ${processInstanceId}` : 'All Processes',
      summary: result.data.summary,
      analysis: await this.analyzeTokenDistribution(),
      recommendations: this.generateRecommendations(result.data)
    };
    
    return report;
  }
  
  generateRecommendations(tokenData) {
    const recommendations = [];
    const { summary } = tokenData;
    
    // Check for high number of active tokens
    const activePercentage = (summary.active_tokens / summary.total_tokens) * 100;
    if (activePercentage > 70) {
      recommendations.push({
        type: 'performance',
        priority: 'high',
        message: `High percentage of active tokens (${activePercentage.toFixed(1)}%)`,
        suggestion: 'Review process execution and potential bottlenecks'
      });
    }
    
    // Check for cancelled tokens
    if (summary.cancelled_tokens > 0) {
      const cancelledPercentage = (summary.cancelled_tokens / summary.total_tokens) * 100;
      recommendations.push({
        type: 'quality',
        priority: cancelledPercentage > 10 ? 'high' : 'medium',
        message: `${summary.cancelled_tokens} cancelled tokens (${cancelledPercentage.toFixed(1)}%)`,
        suggestion: 'Investigate causes of token cancellation'
      });
    }
    
    // Check for element type imbalances
    const serviceTaskPercentage = (summary.by_element_type.SERVICE_TASK || 0) / summary.total_tokens * 100;
    if (serviceTaskPercentage > 50) {
      recommendations.push({
        type: 'architecture',
        priority: 'medium',
        message: `High concentration of Service Task tokens (${serviceTaskPercentage.toFixed(1)}%)`,
        suggestion: 'Consider breaking down complex service tasks or optimizing external service calls'
      });
    }
    
    return recommendations;
  }
}

// Использование
const monitor = new TokenMonitor('your-api-key');

// Получение активных токенов
const activeTokens = await monitor.getActiveTokens();
console.log('Active tokens:', activeTokens);

// Анализ распределения токенов
const analysis = await monitor.analyzeTokenDistribution();
console.log('Token analysis:', analysis);

// Токены конкретного процесса
const processTokens = await monitor.getProcessTokens('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Process tokens:', processTokens);

// Генерация отчета
const report = await monitor.generateTokenReport();
console.log('Token report:', report);

// Мониторинг в реальном времени
const monitoringInterval = await monitor.startRealTimeMonitoring((analysis) => {
  console.log('Current token distribution:', analysis.state_distribution);
  
  if (analysis.bottlenecks.length > 0) {
    console.warn('Bottlenecks detected:', analysis.bottlenecks);
  }
}, 10000); // Every 10 seconds

// Остановка мониторинга через 5 минут
setTimeout(() => {
  clearInterval(monitoringInterval);
  console.log('Token monitoring stopped');
}, 5 * 60 * 1000);
```

### Process Flow Tracker
```javascript
class ProcessFlowTracker {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.tokenMonitor = new TokenMonitor(apiKey);
  }
  
  async traceProcessExecution(processInstanceId) {
    const tokens = await this.tokenMonitor.getProcessTokens(processInstanceId);
    
    if (!tokens.success || tokens.data.tokens.length === 0) {
      return { error: 'No tokens found for process instance' };
    }
    
    return this.buildExecutionTrace(tokens.data.tokens);
  }
  
  buildExecutionTrace(tokens) {
    // Sort tokens by creation time to build chronological flow
    const sortedTokens = tokens.sort((a, b) => 
      new Date(a.created_at) - new Date(b.created_at)
    );
    
    const trace = {
      process_instance_id: sortedTokens[0].process_instance_id,
      process_definition_key: sortedTokens[0].process_definition_key,
      execution_flow: [],
      current_state: this.analyzeCurrentState(tokens),
      execution_statistics: this.calculateExecutionStats(tokens),
      parallel_branches: this.identifyParallelBranches(tokens)
    };
    
    // Build execution flow
    sortedTokens.forEach(token => {
      const flowStep = {
        step_number: trace.execution_flow.length + 1,
        token_id: token.id,
        element_id: token.element_id,
        element_name: token.element_name,
        element_type: token.element_type,
        started_at: token.created_at,
        completed_at: token.completed_at,
        duration_ms: token.completed_at ? 
          new Date(token.completed_at) - new Date(token.created_at) : null,
        state: token.state,
        variables_at_step: token.variables,
        waiting_for: token.execution_context?.waiting_for,
        is_parallel_branch: token.parent_token_id !== null
      };
      
      trace.execution_flow.push(flowStep);
    });
    
    return trace;
  }
  
  analyzeCurrentState(tokens) {
    const activeTokens = tokens.filter(token => token.state === 'ACTIVE');
    
    return {
      active_tokens_count: activeTokens.length,
      current_activities: activeTokens.map(token => ({
        element_id: token.element_id,
        element_name: token.element_name,
        element_type: token.element_type,
        waiting_for: token.execution_context?.waiting_for,
        running_since: token.created_at
      })),
      is_waiting: activeTokens.some(token => 
        token.execution_context?.waiting_for !== undefined
      ),
      next_possible_steps: this.predictNextSteps(activeTokens)
    };
  }
  
  calculateExecutionStats(tokens) {
    const completedTokens = tokens.filter(token => token.state === 'COMPLETED');
    const activeTokens = tokens.filter(token => token.state === 'ACTIVE');
    
    const stats = {
      total_steps: tokens.length,
      completed_steps: completedTokens.length,
      active_steps: activeTokens.length,
      progress_percentage: tokens.length > 0 ? 
        (completedTokens.length / tokens.length) * 100 : 0
    };
    
    if (completedTokens.length > 0) {
      const durations = completedTokens
        .filter(token => token.completed_at)
        .map(token => new Date(token.completed_at) - new Date(token.created_at));
      
      stats.total_execution_time_ms = durations.reduce((sum, duration) => sum + duration, 0);
      stats.average_step_duration_ms = stats.total_execution_time_ms / durations.length;
      stats.fastest_step_ms = Math.min(...durations);
      stats.slowest_step_ms = Math.max(...durations);
    }
    
    return stats;
  }
  
  identifyParallelBranches(tokens) {
    const branches = {};
    
    tokens.forEach(token => {
      if (token.child_tokens && token.child_tokens.length > 0) {
        branches[token.id] = {
          parent_element: {
            id: token.element_id,
            name: token.element_name,
            type: token.element_type
          },
          child_tokens: token.child_tokens,
          branch_count: token.child_tokens.length,
          created_at: token.created_at
        };
      }
    });
    
    return branches;
  }
  
  predictNextSteps(activeTokens) {
    return activeTokens.map(token => ({
      current_element: token.element_id,
      possible_next_elements: token.execution_context?.next_possible_steps || [],
      depends_on: token.execution_context?.waiting_for
    }));
  }
  
  async compareProcessExecutions(processInstanceIds) {
    const traces = await Promise.all(
      processInstanceIds.map(id => this.traceProcessExecution(id))
    );
    
    return {
      processes_compared: processInstanceIds.length,
      traces,
      comparison: this.generateExecutionComparison(traces)
    };
  }
  
  generateExecutionComparison(traces) {
    const validTraces = traces.filter(trace => !trace.error);
    
    if (validTraces.length < 2) {
      return { error: 'Need at least 2 valid traces for comparison' };
    }
    
    return {
      common_elements: this.findCommonElements(validTraces),
      execution_time_comparison: this.compareExecutionTimes(validTraces),
      path_differences: this.identifyPathDifferences(validTraces),
      performance_insights: this.generatePerformanceInsights(validTraces)
    };
  }
  
  findCommonElements(traces) {
    const elementSets = traces.map(trace => 
      new Set(trace.execution_flow.map(step => step.element_id))
    );
    
    const commonElements = elementSets.reduce((common, currentSet) => 
      new Set([...common].filter(element => currentSet.has(element)))
    );
    
    return Array.from(commonElements);
  }
  
  compareExecutionTimes(traces) {
    const timeComparisons = {};
    
    traces.forEach((trace, index) => {
      timeComparisons[`process_${index + 1}`] = {
        total_time_ms: trace.execution_statistics.total_execution_time_ms || 0,
        average_step_time_ms: trace.execution_statistics.average_step_duration_ms || 0,
        step_count: trace.execution_statistics.total_steps
      };
    });
    
    return timeComparisons;
  }
}

// Использование
const tracker = new ProcessFlowTracker('your-api-key');

// Трассировка выполнения процесса
const trace = await tracker.traceProcessExecution('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Process execution trace:', trace);

// Сравнение выполнения нескольких процессов
const comparison = await tracker.compareProcessExecutions([
  'srv1-aB3dEf9hK2mN5pQ8uV',
  'srv1-cD6fGh9iJ3kL7mN0pR'
]);
console.log('Process comparison:', comparison);
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/tokens`](../processes/get-process-tokens.md) - Токены конкретного процесса
- [`GET /api/v1/processes/:id/tokens/:token_id/trace`](../processes/get-token-trace.md) - Трассировка токена
- [`GET /api/v1/processes`](../processes/list-processes.md) - Связанные процессы
