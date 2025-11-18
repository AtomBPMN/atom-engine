# GET /api/v1/processes/:id/typed/activities

## Описание
Получение типизированной информации об активностях процесса с детальными метаданными, временными характеристиками и бизнес-контекстом.

## URL
```
GET /api/v1/processes/:id/typed/activities
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры запроса (Query Parameters)
- `state` (string): Фильтр по состоянию (`active`, `completed`, `cancelled`, `all`) (по умолчанию: `all`)
- `include_timeline` (boolean): Включить временную шкалу (по умолчанию: true)
- `include_metrics` (boolean): Включить метрики производительности (по умолчанию: true)

## Примеры запросов

### Все активности с метриками
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/activities" \
  -H "X-API-Key: your-api-key-here"
```

### Только активные активности
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/activities?state=active" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/activities?include_metrics=true', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const activities = await response.json();
```

## Ответы

### 200 OK - Типизированные активности
```json
{
  "success": true,
  "data": {
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_definition_key": "Order_Processing",
    "process_version": 3,
    "retrieved_at": "2025-01-11T11:30:00.123Z",
    "activities": [
      {
        "activity_id": "start_001",
        "element_id": "StartEvent_1",
        "element_name": "Order Received",
        "element_type": "START_EVENT",
        "activity_state": "COMPLETED",
        "started_at": "2025-01-11T10:15:30.123Z",
        "completed_at": "2025-01-11T10:15:30.173Z",
        "duration_ms": 50,
        "token_id": "srv1-token123abc456def789",
        "sequence_counter": 1,
        "business_context": {
          "operation": "Process Initiation",
          "trigger": "Order Submission",
          "initiated_by": "customer_portal",
          "correlation_keys": {
            "orderId": "ORD-2025-001234",
            "customerId": "CUST-789456"
          }
        },
        "input_variables": {},
        "output_variables": {
          "orderId": "ORD-2025-001234",
          "orderDate": "2025-01-11T10:15:30.123Z",
          "source": "online"
        },
        "performance_metrics": {
          "execution_efficiency": 100,
          "compared_to_average": "NORMAL",
          "performance_category": "FAST"
        },
        "next_activities": ["ServiceTask_ValidateOrder"],
        "previous_activities": []
      },
      {
        "activity_id": "validate_001",
        "element_id": "ServiceTask_ValidateOrder",
        "element_name": "Validate Order",
        "element_type": "SERVICE_TASK",
        "activity_state": "COMPLETED",
        "started_at": "2025-01-11T10:15:35.456Z",
        "completed_at": "2025-01-11T10:30:10.456Z",
        "duration_ms": 875000,
        "token_id": "srv1-token123abc456def789",
        "sequence_counter": 2,
        "job_info": {
          "job_key": "srv1-job456def789ghi012",
          "job_type": "order-validation",
          "worker": "validation-worker-01",
          "retries_performed": 0,
          "max_retries": 3
        },
        "business_context": {
          "operation": "Order Validation",
          "validation_rules": ["inventory_check", "customer_verification", "payment_method_validation"],
          "external_services": ["inventory_service", "customer_service"],
          "compliance_checks": ["fraud_detection", "sanctions_screening"]
        },
        "input_variables": {
          "orderId": "ORD-2025-001234",
          "customer": {
            "id": "CUST-789456",
            "type": "PREMIUM"
          },
          "orderItems": [
            {
              "productId": "PROD-001",
              "quantity": 1
            }
          ]
        },
        "output_variables": {
          "validationResult": "PASSED",
          "validationScore": 98.5,
          "inventoryStatus": "AVAILABLE",
          "fraudScore": 12.3,
          "customer": {
            "id": "CUST-789456",
            "name": "John Doe",
            "email": "john.doe@example.com",
            "type": "PREMIUM",
            "verified": true
          }
        },
        "performance_metrics": {
          "execution_efficiency": 45,
          "compared_to_average": "SLOWER_THAN_AVERAGE",
          "performance_category": "SLOW",
          "bottlenecks": [
            {
              "service": "customer_service",
              "delay_ms": 450000,
              "reason": "High response time"
            }
          ]
        },
        "external_interactions": [
          {
            "service": "inventory_service",
            "operation": "check_availability",
            "duration_ms": 234,
            "status": "SUCCESS"
          },
          {
            "service": "customer_service",
            "operation": "get_customer_details",
            "duration_ms": 450123,
            "status": "SUCCESS",
            "retries": 1
          },
          {
            "service": "fraud_detection",
            "operation": "screen_order",
            "duration_ms": 1205,
            "status": "SUCCESS"
          }
        ],
        "next_activities": ["ExclusiveGateway_ValidationResult"],
        "previous_activities": ["StartEvent_1"]
      },
      {
        "activity_id": "gateway_001",
        "element_id": "ExclusiveGateway_ValidationResult",
        "element_name": "Validation Result Gateway",
        "element_type": "EXCLUSIVE_GATEWAY",
        "activity_state": "COMPLETED",
        "started_at": "2025-01-11T10:30:10.456Z",
        "completed_at": "2025-01-11T10:30:10.580Z",
        "duration_ms": 124,
        "token_id": "srv1-token123abc456def789",
        "sequence_counter": 3,
        "gateway_evaluation": {
          "conditions_evaluated": [
            {
              "sequence_flow_id": "flow_to_payment",
              "condition": "validationResult = \"PASSED\" and validationScore > 80",
              "result": true,
              "evaluation_time_ms": 45
            },
            {
              "sequence_flow_id": "flow_to_rejection",
              "condition": "validationResult = \"FAILED\" or validationScore <= 80",
              "result": false,
              "evaluation_time_ms": 32
            }
          ],
          "path_taken": "flow_to_payment",
          "decision_reason": "Validation passed with high score (98.5)"
        },
        "business_context": {
          "operation": "Flow Control",
          "decision_point": "Order Validation Result",
          "business_rule": "Orders with validation score > 80 proceed to payment"
        },
        "input_variables": {
          "validationResult": "PASSED",
          "validationScore": 98.5
        },
        "output_variables": {},
        "performance_metrics": {
          "execution_efficiency": 100,
          "compared_to_average": "NORMAL",
          "performance_category": "FAST"
        },
        "next_activities": ["ServiceTask_PaymentProcess"],
        "previous_activities": ["ServiceTask_ValidateOrder"]
      },
      {
        "activity_id": "payment_001",
        "element_id": "ServiceTask_PaymentProcess",
        "element_name": "Process Payment",
        "element_type": "SERVICE_TASK",
        "activity_state": "ACTIVE",
        "started_at": "2025-01-11T10:30:15.123Z",
        "duration_so_far_ms": 3584877,
        "estimated_remaining_ms": 180000,
        "estimated_completion_at": "2025-01-11T11:33:00.000Z",
        "token_id": "srv1-token123abc456def789",
        "sequence_counter": 4,
        "job_info": {
          "job_key": "srv1-job789ghi012jkl345",
          "job_type": "payment-processing",
          "worker": "payment-worker-01",
          "retries_performed": 1,
          "max_retries": 3,
          "last_retry_at": "2025-01-11T11:15:30.456Z",
          "retry_reason": "Timeout waiting for payment gateway response"
        },
        "business_context": {
          "operation": "Payment Processing",
          "payment_method": "credit_card",
          "amount": 1050.99,
          "currency": "USD",
          "payment_gateway": "stripe",
          "merchant_id": "MERCH-123456",
          "transaction_id": "txn_1234567890"
        },
        "input_variables": {
          "totalAmount": 1050.99,
          "customer": {
            "id": "CUST-789456",
            "paymentMethod": "credit_card_ending_1234"
          },
          "paymentGateway": "stripe"
        },
        "current_state": {
          "waiting_for": "payment_gateway_response",
          "last_heartbeat": "2025-01-11T11:29:45.789Z",
          "timeout_at": "2025-01-11T11:35:15.123Z",
          "can_be_cancelled": true
        },
        "external_interactions": [
          {
            "service": "stripe_gateway",
            "operation": "create_payment_intent",
            "started_at": "2025-01-11T10:30:20.456Z",
            "status": "IN_PROGRESS",
            "timeout_at": "2025-01-11T11:35:15.123Z"
          }
        ],
        "performance_metrics": {
          "execution_efficiency": 25,
          "compared_to_average": "MUCH_SLOWER_THAN_AVERAGE",
          "performance_category": "VERY_SLOW",
          "bottlenecks": [
            {
              "service": "stripe_gateway",
              "delay_ms": 3500000,
              "reason": "Extended payment processing time"
            }
          ],
          "alerts": [
            {
              "type": "LONG_RUNNING_ACTIVITY",
              "severity": "WARNING",
              "message": "Activity running longer than expected",
              "threshold_exceeded": "60 minutes"
            }
          ]
        },
        "escalation_info": {
          "escalation_required": false,
          "escalation_threshold": "90 minutes",
          "time_until_escalation": "30 minutes 23 seconds"
        },
        "next_activities": ["ExclusiveGateway_PaymentResult"],
        "previous_activities": ["ExclusiveGateway_ValidationResult"]
      }
    ],
    "execution_timeline": [
      {
        "timestamp": "2025-01-11T10:15:30.123Z",
        "event": "PROCESS_STARTED",
        "activity_id": "start_001",
        "description": "Order processing initiated",
        "icon": "play-circle",
        "color": "green"
      },
      {
        "timestamp": "2025-01-11T10:15:35.456Z",
        "event": "ACTIVITY_STARTED",
        "activity_id": "validate_001",
        "description": "Order validation started",
        "icon": "check-circle",
        "color": "blue"
      },
      {
        "timestamp": "2025-01-11T10:30:10.456Z",
        "event": "ACTIVITY_COMPLETED",
        "activity_id": "validate_001",
        "description": "Order validation completed successfully",
        "icon": "check",
        "color": "green"
      },
      {
        "timestamp": "2025-01-11T10:30:10.580Z",
        "event": "GATEWAY_EVALUATED",
        "activity_id": "gateway_001",
        "description": "Validation result gateway - proceeding to payment",
        "icon": "git-branch",
        "color": "purple"
      },
      {
        "timestamp": "2025-01-11T10:30:15.123Z",
        "event": "ACTIVITY_STARTED",
        "activity_id": "payment_001",
        "description": "Payment processing started",
        "icon": "credit-card",
        "color": "blue"
      },
      {
        "timestamp": "2025-01-11T11:15:30.456Z",
        "event": "ACTIVITY_RETRIED",
        "activity_id": "payment_001",
        "description": "Payment processing retried due to timeout",
        "icon": "refresh-cw",
        "color": "orange"
      }
    ],
    "process_metrics": {
      "total_activities": 4,
      "completed_activities": 3,
      "active_activities": 1,
      "cancelled_activities": 0,
      "total_execution_time_ms": 3659877,
      "average_activity_duration_ms": 1219959,
      "longest_activity": {
        "activity_id": "payment_001",
        "duration_ms": 3584877,
        "still_running": true
      },
      "shortest_activity": {
        "activity_id": "start_001", 
        "duration_ms": 50
      },
      "bottlenecks": [
        {
          "activity_id": "payment_001",
          "issue": "Long running payment processing",
          "impact": "HIGH"
        },
        {
          "activity_id": "validate_001",
          "issue": "Slow customer service response",
          "impact": "MEDIUM"
        }
      ],
      "efficiency_score": 52.3,
      "sla_status": "AT_RISK",
      "estimated_completion": "2025-01-11T11:33:00.000Z"
    }
  },
  "request_id": "req_1641998405300"
}
```

## Использование

### Process Activities Analyzer
```javascript
class ProcessActivitiesAnalyzer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getTypedActivities(processInstanceId, options = {}) {
    const params = new URLSearchParams();
    if (options.state) params.append('state', options.state);
    if (options.includeTimeline !== undefined) {
      params.append('include_timeline', options.includeTimeline);
    }
    if (options.includeMetrics !== undefined) {
      params.append('include_metrics', options.includeMetrics);
    }
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/activities?${params}`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch activities: ${response.statusText}`);
    }
    
    return await response.json();
  }
  
  async getActiveActivities(processInstanceId) {
    const result = await this.getTypedActivities(processInstanceId, { state: 'active' });
    return result.data.activities.filter(activity => activity.activity_state === 'ACTIVE');
  }
  
  async getCompletedActivities(processInstanceId) {
    const result = await this.getTypedActivities(processInstanceId, { state: 'completed' });
    return result.data.activities.filter(activity => activity.activity_state === 'COMPLETED');
  }
  
  async analyzePerformance(processInstanceId) {
    const result = await this.getTypedActivities(processInstanceId, {
      includeMetrics: true
    });
    
    const { activities, process_metrics } = result.data;
    
    return {
      overall_efficiency: process_metrics.efficiency_score,
      sla_status: process_metrics.sla_status,
      bottlenecks: this.identifyBottlenecks(activities),
      performance_distribution: this.analyzePerformanceDistribution(activities),
      recommendations: this.generatePerformanceRecommendations(activities, process_metrics)
    };
  }
  
  identifyBottlenecks(activities) {
    const bottlenecks = [];
    
    activities.forEach(activity => {
      // Long running activities
      if (activity.activity_state === 'ACTIVE' && activity.duration_so_far_ms > 300000) { // > 5 minutes
        bottlenecks.push({
          type: 'LONG_RUNNING',
          activity_id: activity.activity_id,
          element_name: activity.element_name,
          duration_ms: activity.duration_so_far_ms,
          severity: 'HIGH'
        });
      }
      
      // Slow completed activities
      if (activity.activity_state === 'COMPLETED' && 
          activity.performance_metrics?.performance_category === 'SLOW') {
        bottlenecks.push({
          type: 'SLOW_EXECUTION',
          activity_id: activity.activity_id,
          element_name: activity.element_name,
          duration_ms: activity.duration_ms,
          severity: 'MEDIUM'
        });
      }
      
      // External service delays
      if (activity.external_interactions) {
        activity.external_interactions.forEach(interaction => {
          if (interaction.duration_ms > 30000) { // > 30 seconds
            bottlenecks.push({
              type: 'EXTERNAL_SERVICE_DELAY',
              activity_id: activity.activity_id,
              service: interaction.service,
              operation: interaction.operation,
              duration_ms: interaction.duration_ms,
              severity: 'MEDIUM'
            });
          }
        });
      }
      
      // Retry indicators
      if (activity.job_info?.retries_performed > 0) {
        bottlenecks.push({
          type: 'RETRY_REQUIRED',
          activity_id: activity.activity_id,
          element_name: activity.element_name,
          retries: activity.job_info.retries_performed,
          severity: 'LOW'
        });
      }
    });
    
    return bottlenecks.sort((a, b) => {
      const severityOrder = { 'HIGH': 3, 'MEDIUM': 2, 'LOW': 1 };
      return severityOrder[b.severity] - severityOrder[a.severity];
    });
  }
  
  analyzePerformanceDistribution(activities) {
    const completed = activities.filter(a => a.activity_state === 'COMPLETED');
    
    if (completed.length === 0) {
      return { message: 'No completed activities to analyze' };
    }
    
    const durations = completed.map(a => a.duration_ms);
    const categories = { FAST: 0, NORMAL: 0, SLOW: 0, VERY_SLOW: 0 };
    
    completed.forEach(activity => {
      const category = activity.performance_metrics?.performance_category || 'NORMAL';
      categories[category] = (categories[category] || 0) + 1;
    });
    
    return {
      total_activities: completed.length,
      average_duration_ms: durations.reduce((sum, d) => sum + d, 0) / durations.length,
      min_duration_ms: Math.min(...durations),
      max_duration_ms: Math.max(...durations),
      categories,
      performance_spread: {
        fast_percentage: (categories.FAST / completed.length) * 100,
        slow_percentage: ((categories.SLOW + categories.VERY_SLOW) / completed.length) * 100
      }
    };
  }
  
  generatePerformanceRecommendations(activities, processMetrics) {
    const recommendations = [];
    
    // Overall efficiency
    if (processMetrics.efficiency_score < 70) {
      recommendations.push({
        type: 'EFFICIENCY',
        priority: 'HIGH',
        message: `Process efficiency (${processMetrics.efficiency_score}%) is below target`,
        actions: [
          'Review process design for optimization opportunities',
          'Identify and address bottlenecks',
          'Consider parallel execution where possible'
        ]
      });
    }
    
    // SLA risk
    if (processMetrics.sla_status === 'AT_RISK') {
      recommendations.push({
        type: 'SLA_RISK',
        priority: 'HIGH',
        message: 'Process is at risk of missing SLA',
        actions: [
          'Monitor closely for escalation',
          'Prepare contingency plans',
          'Notify stakeholders of potential delay'
        ]
      });
    }
    
    // Long running activities
    const longRunning = activities.filter(a => 
      a.activity_state === 'ACTIVE' && a.duration_so_far_ms > 600000 // > 10 minutes
    );
    
    if (longRunning.length > 0) {
      recommendations.push({
        type: 'LONG_RUNNING_ACTIVITIES',
        priority: 'MEDIUM',
        message: `${longRunning.length} activities running longer than expected`,
        actions: [
          'Check external service health',
          'Consider timeout adjustments',
          'Review retry strategies'
        ],
        activities: longRunning.map(a => a.element_name)
      });
    }
    
    // External service issues
    const externalIssues = activities.filter(a => 
      a.performance_metrics?.bottlenecks?.some(b => b.service)
    );
    
    if (externalIssues.length > 0) {
      recommendations.push({
        type: 'EXTERNAL_SERVICES',
        priority: 'MEDIUM',
        message: 'External service performance issues detected',
        actions: [
          'Monitor external service health',
          'Consider circuit breaker patterns',
          'Implement fallback mechanisms'
        ]
      });
    }
    
    return recommendations;
  }
  
  async createExecutionVisualization(processInstanceId, containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container with ID '${containerId}' not found`);
    }
    
    const result = await this.getTypedActivities(processInstanceId, {
      includeTimeline: true,
      includeMetrics: true
    });
    
    const { activities, execution_timeline } = result.data;
    
    // Create timeline visualization
    const timeline = document.createElement('div');
    timeline.className = 'execution-timeline';
    
    execution_timeline.forEach(event => {
      const eventElement = document.createElement('div');
      eventElement.className = `timeline-event ${event.color}`;
      
      eventElement.innerHTML = `
        <div class="event-time">${new Date(event.timestamp).toLocaleTimeString()}</div>
        <div class="event-icon">${this.getEventIcon(event.icon)}</div>
        <div class="event-description">${event.description}</div>
      `;
      
      timeline.appendChild(eventElement);
    });
    
    // Create activities table
    const activitiesTable = document.createElement('table');
    activitiesTable.className = 'activities-table';
    
    const headerRow = document.createElement('tr');
    headerRow.innerHTML = `
      <th>Activity</th>
      <th>Type</th>
      <th>State</th>
      <th>Duration</th>
      <th>Performance</th>
      <th>Issues</th>
    `;
    activitiesTable.appendChild(headerRow);
    
    activities.forEach(activity => {
      const row = document.createElement('tr');
      row.className = activity.activity_state.toLowerCase();
      
      const duration = activity.activity_state === 'ACTIVE' ? 
        activity.duration_so_far_ms : activity.duration_ms;
      
      const performanceClass = activity.performance_metrics?.performance_category?.toLowerCase() || 'normal';
      
      const issues = [];
      if (activity.job_info?.retries_performed > 0) {
        issues.push(`${activity.job_info.retries_performed} retries`);
      }
      if (activity.performance_metrics?.alerts) {
        issues.push(...activity.performance_metrics.alerts.map(alert => alert.message));
      }
      
      row.innerHTML = `
        <td>${activity.element_name}</td>
        <td>${activity.element_type.replace('_', ' ')}</td>
        <td><span class="state-badge ${activity.activity_state.toLowerCase()}">${activity.activity_state}</span></td>
        <td>${this.formatDuration(duration)}</td>
        <td><span class="performance-badge ${performanceClass}">${activity.performance_metrics?.performance_category || 'NORMAL'}</span></td>
        <td>${issues.join(', ') || '-'}</td>
      `;
      
      activitiesTable.appendChild(row);
    });
    
    // Combine elements
    const wrapper = document.createElement('div');
    wrapper.className = 'process-execution-view';
    
    const timelineSection = document.createElement('div');
    timelineSection.className = 'timeline-section';
    timelineSection.innerHTML = '<h3>Execution Timeline</h3>';
    timelineSection.appendChild(timeline);
    
    const activitiesSection = document.createElement('div');
    activitiesSection.className = 'activities-section';
    activitiesSection.innerHTML = '<h3>Activities</h3>';
    activitiesSection.appendChild(activitiesTable);
    
    wrapper.appendChild(timelineSection);
    wrapper.appendChild(activitiesSection);
    
    container.appendChild(wrapper);
    
    return wrapper;
  }
  
  getEventIcon(iconName) {
    const icons = {
      'play-circle': '▶',
      'check-circle': '🔄',
      'check': '✅',
      'git-branch': '🔀',
      'credit-card': '💳',
      'refresh-cw': '🔄',
      'alert-triangle': '⚠️',
      'x-circle': '❌'
    };
    
    return icons[iconName] || '●';
  }
  
  formatDuration(ms) {
    if (ms < 1000) return `${ms}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    if (ms < 3600000) return `${(ms / 60000).toFixed(1)}m`;
    return `${(ms / 3600000).toFixed(1)}h`;
  }
  
  async monitorProcessProgress(processInstanceId, callback, interval = 10000) {
    let lastUpdateTime = null;
    
    const monitor = async () => {
      try {
        const result = await this.getTypedActivities(processInstanceId, {
          includeMetrics: true
        });
        
        const currentUpdateTime = result.data.retrieved_at;
        
        // Only call callback if there are changes
        if (currentUpdateTime !== lastUpdateTime) {
          const analysis = {
            activities: result.data.activities,
            metrics: result.data.process_metrics,
            active_count: result.data.activities.filter(a => a.activity_state === 'ACTIVE').length,
            completed_count: result.data.activities.filter(a => a.activity_state === 'COMPLETED').length,
            bottlenecks: this.identifyBottlenecks(result.data.activities)
          };
          
          callback(analysis);
          lastUpdateTime = currentUpdateTime;
        }
        
      } catch (error) {
        console.error('Error monitoring process progress:', error);
        callback({ error: error.message });
      }
    };
    
    // Initial call
    await monitor();
    
    // Set up interval
    const intervalId = setInterval(monitor, interval);
    
    return () => clearInterval(intervalId);
  }
}

// Использование
const analyzer = new ProcessActivitiesAnalyzer('your-api-key');

// Получение всех активностей
const activities = await analyzer.getTypedActivities('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Process activities:', activities);

// Анализ производительности
const performance = await analyzer.analyzePerformance('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Performance analysis:', performance);

// Создание визуализации выполнения
await analyzer.createExecutionVisualization('srv1-aB3dEf9hK2mN5pQ8uV', 'execution-view');

// Мониторинг прогресса
const stopMonitoring = await analyzer.monitorProcessProgress(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  (analysis) => {
    console.log(`Active activities: ${analysis.active_count}`);
    console.log(`Completed activities: ${analysis.completed_count}`);
    
    if (analysis.bottlenecks.length > 0) {
      console.warn('Bottlenecks detected:', analysis.bottlenecks);
    }
  }
);

// Остановка мониторинга через 5 минут
setTimeout(stopMonitoring, 5 * 60 * 1000);
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/typed/status`](./get-process-typed-status.md) - Типизированный статус процесса
- [`GET /api/v1/processes/:id/typed/info`](./get-process-typed-info.md) - Типизированная информация процесса
- [`GET /api/v1/tokens`](../tokens/list-tokens.md) - Токены выполнения
