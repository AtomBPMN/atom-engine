# GET /api/v1/processes/:id/typed/status

## Описание
Получение типизированного статуса процесса с расширенной информацией о прогрессе, временных метриках и бизнес-контексте.

## URL
```
GET /api/v1/processes/:id/typed/status
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры запроса (Query Parameters)
- `include_timeline` (boolean): Включить временную шкалу выполнения (по умолчанию: true)
- `include_metrics` (boolean): Включить метрики производительности (по умолчанию: true)

## Примеры запросов

### Полный типизированный статус
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/status" \
  -H "X-API-Key: your-api-key-here"
```

### Статус без метрик
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/status?include_metrics=false" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/status', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const typedStatus = await response.json();
```

## Ответы

### 200 OK - Типизированный статус процесса
```json
{
  "success": true,
  "data": {
    "process_instance": {
      "id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "process_definition_key": "Order_Processing",
      "process_definition_version": 3,
      "status": "ACTIVE",
      "started_at": "2025-01-11T10:15:30.123Z",
      "updated_at": "2025-01-11T10:30:45.456Z",
      "tenant_id": "tenant-001"
    },
    "execution_status": {
      "state": "RUNNING",
      "state_description": "Process is actively executing",
      "health": "HEALTHY",
      "last_activity_at": "2025-01-11T10:30:45.456Z",
      "is_waiting": true,
      "waiting_reason": "External service call in progress",
      "can_be_cancelled": true,
      "estimated_completion": "2025-01-11T10:45:00.000Z"
    },
    "progress_info": {
      "current_phase": "Payment Processing",
      "completion_percentage": 42.5,
      "steps_completed": 3,
      "total_estimated_steps": 7,
      "milestones": [
        {
          "name": "Order Validation",
          "status": "COMPLETED",
          "completed_at": "2025-01-11T10:20:15.789Z"
        },
        {
          "name": "Inventory Check",
          "status": "COMPLETED", 
          "completed_at": "2025-01-11T10:25:30.456Z"
        },
        {
          "name": "Payment Processing",
          "status": "IN_PROGRESS",
          "started_at": "2025-01-11T10:30:15.123Z",
          "estimated_completion": "2025-01-11T10:35:00.000Z"
        },
        {
          "name": "Order Fulfillment",
          "status": "PENDING",
          "estimated_start": "2025-01-11T10:35:00.000Z"
        },
        {
          "name": "Shipping",
          "status": "PENDING",
          "estimated_start": "2025-01-11T10:40:00.000Z"
        }
      ]
    },
    "current_activities": [
      {
        "element_id": "ServiceTask_PaymentProcess",
        "element_name": "Process Payment",
        "element_type": "SERVICE_TASK",
        "activity_state": "ACTIVE",
        "started_at": "2025-01-11T10:30:15.123Z",
        "duration_so_far_ms": 30333,
        "estimated_remaining_ms": 270000,
        "assigned_worker": "payment-worker-01",
        "retry_count": 0,
        "max_retries": 3,
        "timeout_at": "2025-01-11T10:35:15.123Z",
        "business_context": {
          "operation": "Credit Card Processing",
          "amount": 1050.99,
          "currency": "USD",
          "merchant_id": "MERCH-123456"
        }
      }
    ],
    "recent_activities": [
      {
        "element_id": "ServiceTask_ValidateOrder",
        "element_name": "Validate Order",
        "element_type": "SERVICE_TASK",
        "completed_at": "2025-01-11T10:30:10.456Z",
        "duration_ms": 14876,
        "result": "SUCCESS",
        "output_variables": {
          "validationResult": "PASSED",
          "validationScore": 98.5
        }
      },
      {
        "element_id": "ExclusiveGateway_OrderCheck",
        "element_name": "Order Validation Gateway",
        "element_type": "EXCLUSIVE_GATEWAY",
        "completed_at": "2025-01-11T10:30:15.123Z",
        "duration_ms": 124,
        "result": "SUCCESS",
        "decision_taken": {
          "condition": "validationScore > 90",
          "path_chosen": "ServiceTask_PaymentProcess",
          "reason": "Order validation passed with high score"
        }
      }
    ],
    "timeline": [
      {
        "timestamp": "2025-01-11T10:15:30.123Z",
        "event_type": "PROCESS_STARTED",
        "title": "Order Processing Started",
        "description": "New order ORD-2025-001234 received and process initiated",
        "icon": "play",
        "color": "green"
      },
      {
        "timestamp": "2025-01-11T10:15:35.456Z",
        "event_type": "ACTIVITY_STARTED",
        "title": "Order Validation Started",
        "description": "Validating order items and customer information",
        "element_id": "ServiceTask_ValidateOrder",
        "icon": "check-circle",
        "color": "blue"
      },
      {
        "timestamp": "2025-01-11T10:30:10.456Z",
        "event_type": "ACTIVITY_COMPLETED",
        "title": "Order Validation Completed",
        "description": "Order validation passed with score 98.5",
        "element_id": "ServiceTask_ValidateOrder",
        "duration_ms": 14876,
        "icon": "check",
        "color": "green"
      },
      {
        "timestamp": "2025-01-11T10:30:15.123Z",
        "event_type": "ACTIVITY_STARTED",
        "title": "Payment Processing Started",
        "description": "Processing credit card payment for $1,050.99",
        "element_id": "ServiceTask_PaymentProcess",
        "icon": "credit-card",
        "color": "blue"
      }
    ],
    "business_metrics": {
      "priority": "HIGH",
      "business_value": 1050.99,
      "currency": "USD",
      "customer_tier": "PREMIUM",
      "sla_status": "ON_TRACK",
      "sla_deadline": "2025-01-11T18:00:00.000Z",
      "time_until_sla_breach": "7h 29m 15s",
      "escalation_level": 0,
      "business_impact": "MEDIUM",
      "department": "Sales",
      "cost_center": "CC-001-SALES"
    },
    "performance_metrics": {
      "total_execution_time_ms": 916333,
      "average_activity_duration_ms": 7431,
      "slowest_activity": {
        "element_id": "ServiceTask_ValidateOrder",
        "duration_ms": 14876,
        "performance_category": "SLOW"
      },
      "fastest_activity": {
        "element_id": "ExclusiveGateway_OrderCheck", 
        "duration_ms": 124,
        "performance_category": "FAST"
      },
      "throughput_score": 7.8,
      "efficiency_rating": "GOOD",
      "compared_to_average": {
        "faster_by_percent": 15.3,
        "status": "ABOVE_AVERAGE"
      }
    },
    "health_indicators": {
      "overall_health": "HEALTHY",
      "indicators": [
        {
          "name": "Execution Flow",
          "status": "HEALTHY",
          "description": "Process is progressing normally",
          "last_check": "2025-01-11T10:30:45.456Z"
        },
        {
          "name": "Resource Usage",
          "status": "HEALTHY", 
          "description": "Memory and CPU usage within normal limits",
          "last_check": "2025-01-11T10:30:45.456Z"
        },
        {
          "name": "External Dependencies",
          "status": "WARNING",
          "description": "Payment service response time elevated",
          "last_check": "2025-01-11T10:30:45.456Z",
          "details": {
            "payment_service_latency_ms": 2500,
            "normal_latency_ms": 800,
            "threshold_exceeded": true
          }
        },
        {
          "name": "SLA Compliance",
          "status": "HEALTHY",
          "description": "Process is on track to meet SLA",
          "last_check": "2025-01-11T10:30:45.456Z"
        }
      ]
    },
    "next_steps": [
      {
        "element_id": "ServiceTask_PaymentProcess",
        "description": "Complete payment processing",
        "estimated_start": "immediate",
        "estimated_duration_ms": 270000,
        "depends_on": "payment service response"
      },
      {
        "element_id": "ExclusiveGateway_PaymentResult",
        "description": "Route based on payment result",
        "estimated_start": "after payment completion",
        "estimated_duration_ms": 100
      },
      {
        "element_id": "ServiceTask_OrderFulfillment",
        "description": "Initiate order fulfillment",
        "estimated_start": "if payment successful",
        "estimated_duration_ms": 180000
      }
    ]
  },
  "request_id": "req_1641998405000"
}
```

## Status Types

### Process States
- **RUNNING** - Процесс активно выполняется
- **WAITING** - Процесс ожидает внешнего события
- **SUSPENDED** - Процесс приостановлен
- **COMPLETED** - Процесс завершен успешно
- **CANCELLED** - Процесс отменен
- **FAILED** - Процесс завершен с ошибкой

### Health Statuses
- **HEALTHY** - Нормальное выполнение
- **WARNING** - Потенциальные проблемы
- **CRITICAL** - Серьезные проблемы
- **UNKNOWN** - Статус неизвестен

## Использование

### Process Status Monitor
```javascript
class ProcessStatusMonitor {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.monitors = new Map();
  }
  
  async getTypedStatus(processInstanceId, options = {}) {
    const params = new URLSearchParams();
    if (options.includeTimeline !== undefined) {
      params.append('include_timeline', options.includeTimeline);
    }
    if (options.includeMetrics !== undefined) {
      params.append('include_metrics', options.includeMetrics);
    }
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/status?${params}`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch status: ${response.statusText}`);
    }
    
    return await response.json();
  }
  
  async getExecutionProgress(processInstanceId) {
    const status = await this.getTypedStatus(processInstanceId);
    return {
      percentage: status.data.progress_info.completion_percentage,
      current_phase: status.data.progress_info.current_phase,
      milestones: status.data.progress_info.milestones,
      estimated_completion: status.data.execution_status.estimated_completion
    };
  }
  
  async getHealthStatus(processInstanceId) {
    const status = await this.getTypedStatus(processInstanceId);
    return status.data.health_indicators;
  }
  
  async analyzePerformance(processInstanceId) {
    const status = await this.getTypedStatus(processInstanceId);
    const metrics = status.data.performance_metrics;
    
    return {
      efficiency: metrics.efficiency_rating,
      throughput_score: metrics.throughput_score,
      execution_time: metrics.total_execution_time_ms,
      bottlenecks: this.identifyBottlenecks(status.data),
      recommendations: this.generatePerformanceRecommendations(metrics)
    };
  }
  
  identifyBottlenecks(statusData) {
    const bottlenecks = [];
    
    // Check slow activities
    if (statusData.performance_metrics.slowest_activity?.performance_category === 'SLOW') {
      bottlenecks.push({
        type: 'SLOW_ACTIVITY',
        element_id: statusData.performance_metrics.slowest_activity.element_id,
        duration_ms: statusData.performance_metrics.slowest_activity.duration_ms,
        severity: 'MEDIUM'
      });
    }
    
    // Check waiting activities
    statusData.current_activities?.forEach(activity => {
      if (activity.duration_so_far_ms > 60000) { // More than 1 minute
        bottlenecks.push({
          type: 'LONG_RUNNING_ACTIVITY',
          element_id: activity.element_id,
          duration_ms: activity.duration_so_far_ms,
          severity: 'HIGH'
        });
      }
    });
    
    // Check health warnings
    statusData.health_indicators.indicators?.forEach(indicator => {
      if (indicator.status === 'WARNING' || indicator.status === 'CRITICAL') {
        bottlenecks.push({
          type: 'HEALTH_ISSUE',
          name: indicator.name,
          description: indicator.description,
          severity: indicator.status === 'CRITICAL' ? 'HIGH' : 'MEDIUM'
        });
      }
    });
    
    return bottlenecks;
  }
  
  generatePerformanceRecommendations(metrics) {
    const recommendations = [];
    
    if (metrics.efficiency_rating === 'POOR') {
      recommendations.push({
        type: 'EFFICIENCY',
        priority: 'HIGH',
        message: 'Process efficiency is below acceptable levels',
        actions: [
          'Review process design for optimization opportunities',
          'Check for unnecessary wait times',
          'Optimize external service calls'
        ]
      });
    }
    
    if (metrics.throughput_score < 5) {
      recommendations.push({
        type: 'THROUGHPUT',
        priority: 'MEDIUM',
        message: 'Process throughput is below target',
        actions: [
          'Consider parallel execution where possible',
          'Optimize database queries',
          'Review resource allocation'
        ]
      });
    }
    
    return recommendations;
  }
  
  startRealTimeMonitoring(processInstanceId, callback, interval = 5000) {
    if (this.monitors.has(processInstanceId)) {
      this.stopMonitoring(processInstanceId);
    }
    
    const monitor = setInterval(async () => {
      try {
        const status = await this.getTypedStatus(processInstanceId);
        callback(status.data);
        
        // Stop monitoring if process is completed or cancelled
        if (['COMPLETED', 'CANCELLED', 'FAILED'].includes(status.data.execution_status.state)) {
          this.stopMonitoring(processInstanceId);
        }
      } catch (error) {
        console.error(`Error monitoring process ${processInstanceId}:`, error);
        callback({ error: error.message });
      }
    }, interval);
    
    this.monitors.set(processInstanceId, monitor);
    
    return monitor;
  }
  
  stopMonitoring(processInstanceId) {
    const monitor = this.monitors.get(processInstanceId);
    if (monitor) {
      clearInterval(monitor);
      this.monitors.delete(processInstanceId);
    }
  }
  
  stopAllMonitoring() {
    for (const [processInstanceId, monitor] of this.monitors) {
      clearInterval(monitor);
    }
    this.monitors.clear();
  }
  
  async createStatusDashboard(processInstanceIds, containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container with ID '${containerId}' not found`);
    }
    
    container.innerHTML = ''; // Clear existing content
    
    const dashboard = document.createElement('div');
    dashboard.className = 'process-status-dashboard';
    
    for (const processInstanceId of processInstanceIds) {
      try {
        const status = await this.getTypedStatus(processInstanceId);
        const processCard = this.createProcessStatusCard(status.data);
        dashboard.appendChild(processCard);
        
        // Start monitoring for this process
        this.startRealTimeMonitoring(processInstanceId, (statusData) => {
          this.updateProcessStatusCard(processCard, statusData);
        });
        
      } catch (error) {
        const errorCard = this.createErrorCard(processInstanceId, error.message);
        dashboard.appendChild(errorCard);
      }
    }
    
    container.appendChild(dashboard);
    
    return dashboard;
  }
  
  createProcessStatusCard(statusData) {
    const card = document.createElement('div');
    card.className = 'process-status-card';
    card.dataset.processId = statusData.process_instance.id;
    
    const header = document.createElement('div');
    header.className = 'card-header';
    
    const title = document.createElement('h3');
    title.textContent = `${statusData.process_instance.process_definition_key} v${statusData.process_instance.process_definition_version}`;
    header.appendChild(title);
    
    const statusBadge = document.createElement('span');
    statusBadge.className = `status-badge ${statusData.execution_status.state.toLowerCase()}`;
    statusBadge.textContent = statusData.execution_status.state;
    header.appendChild(statusBadge);
    
    card.appendChild(header);
    
    // Progress section
    const progressSection = document.createElement('div');
    progressSection.className = 'progress-section';
    
    const progressBar = document.createElement('div');
    progressBar.className = 'progress-bar';
    
    const progressFill = document.createElement('div');
    progressFill.className = 'progress-fill';
    progressFill.style.width = `${statusData.progress_info.completion_percentage}%`;
    progressBar.appendChild(progressFill);
    
    const progressText = document.createElement('div');
    progressText.className = 'progress-text';
    progressText.textContent = `${statusData.progress_info.completion_percentage.toFixed(1)}% - ${statusData.progress_info.current_phase}`;
    
    progressSection.appendChild(progressBar);
    progressSection.appendChild(progressText);
    card.appendChild(progressSection);
    
    // Health indicators
    const healthSection = document.createElement('div');
    healthSection.className = 'health-section';
    
    statusData.health_indicators.indicators.forEach(indicator => {
      const healthItem = document.createElement('div');
      healthItem.className = `health-item ${indicator.status.toLowerCase()}`;
      healthItem.innerHTML = `
        <span class="health-name">${indicator.name}</span>
        <span class="health-status">${indicator.status}</span>
      `;
      healthSection.appendChild(healthItem);
    });
    
    card.appendChild(healthSection);
    
    // Current activities
    if (statusData.current_activities.length > 0) {
      const activitiesSection = document.createElement('div');
      activitiesSection.className = 'activities-section';
      
      const activitiesTitle = document.createElement('h4');
      activitiesTitle.textContent = 'Current Activities';
      activitiesSection.appendChild(activitiesTitle);
      
      statusData.current_activities.forEach(activity => {
        const activityItem = document.createElement('div');
        activityItem.className = 'activity-item';
        activityItem.innerHTML = `
          <div class="activity-name">${activity.element_name}</div>
          <div class="activity-duration">Running for ${Math.round(activity.duration_so_far_ms / 1000)}s</div>
        `;
        activitiesSection.appendChild(activityItem);
      });
      
      card.appendChild(activitiesSection);
    }
    
    return card;
  }
  
  updateProcessStatusCard(card, statusData) {
    // Update progress
    const progressFill = card.querySelector('.progress-fill');
    const progressText = card.querySelector('.progress-text');
    
    if (progressFill && progressText) {
      progressFill.style.width = `${statusData.progress_info.completion_percentage}%`;
      progressText.textContent = `${statusData.progress_info.completion_percentage.toFixed(1)}% - ${statusData.progress_info.current_phase}`;
    }
    
    // Update status badge
    const statusBadge = card.querySelector('.status-badge');
    if (statusBadge) {
      statusBadge.className = `status-badge ${statusData.execution_status.state.toLowerCase()}`;
      statusBadge.textContent = statusData.execution_status.state;
    }
    
    // Update health indicators
    const healthItems = card.querySelectorAll('.health-item');
    statusData.health_indicators.indicators.forEach((indicator, index) => {
      if (healthItems[index]) {
        healthItems[index].className = `health-item ${indicator.status.toLowerCase()}`;
        healthItems[index].querySelector('.health-status').textContent = indicator.status;
      }
    });
  }
}

// Использование
const monitor = new ProcessStatusMonitor('your-api-key');

// Получение типизированного статуса
const status = await monitor.getTypedStatus('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Process status:', status);

// Анализ производительности
const performance = await monitor.analyzePerformance('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Performance analysis:', performance);

// Мониторинг в реальном времени
monitor.startRealTimeMonitoring('srv1-aB3dEf9hK2mN5pQ8uV', (statusData) => {
  console.log(`Progress: ${statusData.progress_info.completion_percentage}%`);
  console.log(`Current phase: ${statusData.progress_info.current_phase}`);
  console.log(`Health: ${statusData.health_indicators.overall_health}`);
});

// Создание дашборда статуса
const processIds = ['srv1-aB3dEf9hK2mN5pQ8uV', 'srv1-cD6fGh9iJ3kL7mN0pR'];
await monitor.createStatusDashboard(processIds, 'status-dashboard');
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/typed/info`](./get-process-typed-info.md) - Типизированная информация о процессе
- [`GET /api/v1/processes/:id/status`](../processes/get-process-status.md) - Базовый статус процесса
- [`GET /api/v1/processes/:id/typed/activities`](./get-process-typed-activities.md) - Типизированные активности
