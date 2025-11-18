# POST /api/v1/incidents/:id/resolve

## Описание
Решение инцидента путем повтора выполнения или отклонения. Позволяет исправить проблему и продолжить выполнение процесса.

## URL
```
POST /api/v1/incidents/:id/resolve
```

## Авторизация
✅ **Требуется API ключ** с разрешением `incident`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор инцидента

## Параметры тела запроса

### Обязательные поля
- `action` (string): Действие для решения (`retry`, `dismiss`)

### Для action = "retry"
- `retries` (integer, optional): Количество повторов (по умолчанию: 1)
- `variables` (object, optional): Обновленные переменные процесса
- `job_retries` (integer, optional): Количество повторов для задания (только для JOB инцидентов)

### Для action = "dismiss"
- `reason` (string, optional): Причина отклонения

### Общие поля
- `comment` (string, optional): Комментарий к решению

## Примеры запросов

### Повтор инцидента задания с обновленными переменными
```bash
curl -X POST "http://localhost:27555/api/v1/incidents/srv1-inc123abc456def789/resolve" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "action": "retry",
    "retries": 3,
    "job_retries": 5,
    "variables": {
      "paymentServiceUrl": "https://backup-payment.service.com",
      "timeout": 60000
    },
    "comment": "Using backup payment service with increased timeout"
  }'
```

### Отклонение инцидента
```bash
curl -X POST "http://localhost:27555/api/v1/incidents/srv1-inc456def789ghi012/resolve" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "action": "dismiss",
    "reason": "Known issue - payment service maintenance",
    "comment": "Payment service is under maintenance, customer will retry later"
  }'
```

### JavaScript
```javascript
const resolveRequest = {
  action: 'retry',
  retries: 2,
  variables: {
    retryAttempt: 4,
    lastError: null
  },
  comment: 'Resolved by correcting service configuration'
};

const response = await fetch('/api/v1/incidents/srv1-inc123abc456def789/resolve', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(resolveRequest)
});

const result = await response.json();
```

## Ответы

### 200 OK - Инцидент решен повтором
```json
{
  "success": true,
  "data": {
    "incident_id": "srv1-inc123abc456def789",
    "resolution_id": "srv1-res789abc123def456",
    "action": "RETRY",
    "status": "RESOLVED",
    "resolved_at": "2025-01-11T11:30:45.123Z",
    "resolved_by": "api_user",
    "resolution_details": {
      "retries_granted": 3,
      "job_retries_set": 5,
      "variables_updated": {
        "paymentServiceUrl": "https://backup-payment.service.com",
        "timeout": 60000
      },
      "comment": "Using backup payment service with increased timeout"
    },
    "process_continuation": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "resumed": true,
      "next_activity": "ServiceTask_PaymentProcess",
      "new_job_key": "srv1-job123def456ghi789",
      "execution_started": true
    },
    "impact_resolved": {
      "blocked_instances_freed": 1,
      "affected_customers_notified": 1,
      "estimated_recovery_time": "2-5 minutes"
    }
  },
  "request_id": "req_1641998404600"
}
```

### 200 OK - Инцидент отклонен
```json
{
  "success": true,
  "data": {
    "incident_id": "srv1-inc456def789ghi012",
    "resolution_id": "srv1-res456ghi789abc123",
    "action": "DISMISS",
    "status": "DISMISSED",
    "resolved_at": "2025-01-11T11:45:30.456Z",
    "resolved_by": "api_user",
    "resolution_details": {
      "reason": "Known issue - payment service maintenance",
      "comment": "Payment service is under maintenance, customer will retry later"
    },
    "process_impact": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "process_status": "TERMINATED",
      "termination_reason": "Incident dismissed",
      "cleanup_performed": true
    },
    "business_impact": {
      "customer_notification_sent": true,
      "refund_initiated": false,
      "escalation_required": false
    }
  },
  "request_id": "req_1641998404601"
}
```

### 400 Bad Request - Неверное действие
```json
{
  "success": false,
  "error": {
    "code": "INVALID_RESOLUTION_ACTION",
    "message": "Invalid resolution action for incident type",
    "details": {
      "incident_id": "srv1-inc123abc456def789",
      "incident_type": "EXPRESSION",
      "requested_action": "retry",
      "validation_errors": [
        {
          "field": "job_retries",
          "message": "job_retries not applicable for EXPRESSION incidents"
        }
      ],
      "supported_actions": ["retry", "dismiss"],
      "suggestions": [
        "Remove job_retries field for EXPRESSION incidents",
        "Use 'variables' to fix expression context"
      ]
    }
  },
  "request_id": "req_1641998404602"
}
```

### 404 Not Found - Инцидент не найден
```json
{
  "success": false,
  "error": {
    "code": "INCIDENT_NOT_FOUND",
    "message": "Incident not found or already resolved",
    "details": {
      "incident_id": "invalid-incident-id",
      "possible_reasons": [
        "Incident ID is incorrect",
        "Incident was already resolved",
        "Incident was deleted",
        "Insufficient permissions"
      ]
    }
  },
  "request_id": "req_1641998404603"
}
```

### 409 Conflict - Инцидент уже решен
```json
{
  "success": false,
  "error": {
    "code": "INCIDENT_ALREADY_RESOLVED", 
    "message": "Incident is already resolved",
    "details": {
      "incident_id": "srv1-inc123abc456def789",
      "current_status": "RESOLVED",
      "resolved_at": "2025-01-11T10:30:22.789Z",
      "resolved_by": "previous_user",
      "resolution_action": "RETRY"
    }
  },
  "request_id": "req_1641998404604"
}
```

## Resolution Actions

### RETRY
Повторить выполнение после исправления проблемы

**Применимо для:**
- JOB incidents - повторить выполнение задания
- EXPRESSION incidents - повторить вычисление выражения
- TIMER incidents - перезапустить таймер
- MESSAGE incidents - повторить обработку сообщения

### DISMISS
Отклонить инцидент без исправления

**Применимо для:**
- Всех типов инцидентов
- Когда проблема не может быть исправлена
- Когда бизнес-процесс должен быть прерван

## Использование

### Incident Resolver
```javascript
class IncidentResolver {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async resolveIncident(incidentId, action, options = {}) {
    const requestBody = {
      action,
      ...options
    };
    
    const response = await fetch(`/api/v1/incidents/${incidentId}/resolve`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(requestBody)
    });
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to resolve incident: ${error.error.message}`);
    }
    
    return await response.json();
  }
  
  async retryIncident(incidentId, options = {}) {
    return await this.resolveIncident(incidentId, 'retry', options);
  }
  
  async dismissIncident(incidentId, reason, comment) {
    return await this.resolveIncident(incidentId, 'dismiss', {
      reason,
      comment
    });
  }
  
  async retryJobIncident(incidentId, fixes = {}) {
    const options = {
      job_retries: fixes.retries || 3,
      comment: fixes.comment || 'Retrying with corrected configuration'
    };
    
    if (fixes.variables) {
      options.variables = fixes.variables;
    }
    
    return await this.retryIncident(incidentId, options);
  }
  
  async retryExpressionIncident(incidentId, correctedVariables, comment) {
    return await this.retryIncident(incidentId, {
      variables: correctedVariables,
      comment: comment || 'Retrying with corrected variables'
    });
  }
  
  async bulkResolveIncidents(incidentIds, action, options = {}) {
    const results = await Promise.allSettled(
      incidentIds.map(id => this.resolveIncident(id, action, options))
    );
    
    return {
      successful: results.filter(r => r.status === 'fulfilled').map(r => r.value),
      failed: results.filter(r => r.status === 'rejected').map(r => ({
        error: r.reason.message,
        incident_id: r.reason.incident_id
      })),
      summary: {
        total: incidentIds.length,
        successful: results.filter(r => r.status === 'fulfilled').length,
        failed: results.filter(r => r.status === 'rejected').length
      }
    };
  }
  
  async autoResolveCommonIssues(incidents) {
    const resolutions = [];
    
    for (const incident of incidents) {
      try {
        const resolution = await this.determineAutoResolution(incident);
        
        if (resolution) {
          const result = await this.resolveIncident(
            incident.id, 
            resolution.action, 
            resolution.options
          );
          
          resolutions.push({
            incident_id: incident.id,
            status: 'resolved',
            action: resolution.action,
            result
          });
        } else {
          resolutions.push({
            incident_id: incident.id,
            status: 'manual_intervention_required',
            reason: 'No auto-resolution strategy available'
          });
        }
      } catch (error) {
        resolutions.push({
          incident_id: incident.id,
          status: 'failed',
          error: error.message
        });
      }
    }
    
    return resolutions;
  }
  
  async determineAutoResolution(incident) {
    // Auto-resolution logic based on incident type and patterns
    switch (incident.type) {
      case 'JOB':
        return this.determineJobAutoResolution(incident);
      case 'EXPRESSION':
        return this.determineExpressionAutoResolution(incident);
      case 'TIMER':
        return this.determineTimerAutoResolution(incident);
      default:
        return null;
    }
  }
  
  determineJobAutoResolution(incident) {
    const errorCode = incident.error_details?.error_code;
    
    switch (errorCode) {
      case 'TIMEOUT':
      case 'CONNECTION_TIMEOUT':
        return {
          action: 'retry',
          options: {
            job_retries: 2,
            variables: {
              timeout: (incident.error_details?.timeout || 30000) * 2
            },
            comment: 'Auto-retry with increased timeout'
          }
        };
        
      case 'RATE_LIMIT_EXCEEDED':
        return {
          action: 'retry',
          options: {
            job_retries: 1,
            comment: 'Auto-retry after rate limit cooldown'
          }
        };
        
      case 'TEMPORARY_SERVICE_UNAVAILABLE':
        if (incident.severity === 'LOW' || incident.severity === 'MEDIUM') {
          return {
            action: 'retry',
            options: {
              job_retries: 3,
              comment: 'Auto-retry for temporary service issue'
            }
          };
        }
        break;
    }
    
    return null;
  }
  
  determineExpressionAutoResolution(incident) {
    const errorCode = incident.error_details?.error_code;
    
    if (errorCode === 'UNDEFINED_VARIABLE' && incident.severity === 'LOW') {
      // Try to provide default values for common variables
      const commonDefaults = {
        approved: false,
        priority: 'NORMAL',
        amount: 0
      };
      
      const missingVar = incident.error_details?.variable;
      if (missingVar && commonDefaults.hasOwnProperty(missingVar)) {
        return {
          action: 'retry',
          options: {
            variables: {
              [missingVar]: commonDefaults[missingVar]
            },
            comment: `Auto-retry with default value for ${missingVar}`
          }
        };
      }
    }
    
    return null;
  }
  
  determineTimerAutoResolution(incident) {
    if (incident.error_details?.error_code === 'INVALID_DURATION') {
      // Auto-fix common duration format issues
      return {
        action: 'retry',
        options: {
          comment: 'Auto-retry with corrected duration format'
        }
      };
    }
    
    return null;
  }
}

// Использование
const resolver = new IncidentResolver('your-api-key');

// Простое решение инцидента
const result = await resolver.retryIncident('srv1-inc123abc456def789', {
  retries: 3,
  comment: 'Resolved after service restart'
});

// Решение инцидента задания с исправлениями
const jobFixes = {
  retries: 5,
  variables: {
    serviceUrl: 'https://backup-service.com',
    timeout: 60000
  },
  comment: 'Using backup service with increased timeout'
};

const jobResult = await resolver.retryJobIncident('srv1-inc123abc456def789', jobFixes);

// Отклонение инцидента
const dismissResult = await resolver.dismissIncident(
  'srv1-inc456def789ghi012',
  'Service maintenance window',
  'Payment service is in scheduled maintenance'
);

// Массовое решение инцидентов
const incidentIds = ['srv1-inc123', 'srv1-inc456', 'srv1-inc789'];
const bulkResult = await resolver.bulkResolveIncidents(incidentIds, 'retry', {
  retries: 2,
  comment: 'Bulk retry after system fixes'
});

console.log('Bulk resolution result:', bulkResult);

// Автоматическое решение распространенных проблем
const incidents = [
  { id: 'srv1-inc123', type: 'JOB', error_details: { error_code: 'TIMEOUT' } },
  { id: 'srv1-inc456', type: 'EXPRESSION', error_details: { error_code: 'UNDEFINED_VARIABLE', variable: 'approved' } }
];

const autoResolutions = await resolver.autoResolveCommonIssues(incidents);
console.log('Auto-resolution results:', autoResolutions);
```

### Resolution Workflow Manager
```javascript
class ResolutionWorkflowManager {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.resolver = new IncidentResolver(apiKey);
    this.escalationRules = new Map();
  }
  
  addEscalationRule(incidentType, criteria, action) {
    if (!this.escalationRules.has(incidentType)) {
      this.escalationRules.set(incidentType, []);
    }
    
    this.escalationRules.get(incidentType).push({
      criteria,
      action
    });
  }
  
  async processIncidentResolution(incident, userAction) {
    const workflow = {
      incident_id: incident.id,
      steps: [],
      final_status: null,
      escalated: false
    };
    
    // Step 1: Pre-resolution validation
    workflow.steps.push({
      step: 'validation',
      timestamp: new Date().toISOString(),
      status: 'completed',
      details: await this.validateResolution(incident, userAction)
    });
    
    // Step 2: Apply resolution
    try {
      const resolution = await this.resolver.resolveIncident(
        incident.id,
        userAction.action,
        userAction.options
      );
      
      workflow.steps.push({
        step: 'resolution',
        timestamp: new Date().toISOString(),
        status: 'completed',
        details: resolution.data
      });
      
      workflow.final_status = 'resolved';
      
    } catch (error) {
      workflow.steps.push({
        step: 'resolution',
        timestamp: new Date().toISOString(),
        status: 'failed',
        error: error.message
      });
      
      // Step 3: Handle failure - check escalation
      const escalation = await this.checkEscalation(incident, error);
      if (escalation.required) {
        workflow.steps.push({
          step: 'escalation',
          timestamp: new Date().toISOString(),
          status: 'initiated',
          details: escalation
        });
        
        workflow.escalated = true;
        workflow.final_status = 'escalated';
      } else {
        workflow.final_status = 'failed';
      }
    }
    
    // Step 4: Post-resolution verification
    if (workflow.final_status === 'resolved') {
      const verification = await this.verifyResolution(incident.id);
      workflow.steps.push({
        step: 'verification',
        timestamp: new Date().toISOString(),
        status: verification.passed ? 'completed' : 'failed',
        details: verification
      });
    }
    
    return workflow;
  }
  
  async validateResolution(incident, userAction) {
    const validation = {
      valid: true,
      warnings: [],
      recommendations: []
    };
    
    // Check if action is appropriate for incident type
    if (userAction.action === 'retry' && incident.type === 'SYSTEM') {
      validation.warnings.push('Retry may not be effective for SYSTEM incidents');
      validation.recommendations.push('Consider dismiss action if this is a infrastructure issue');
    }
    
    // Check retry limits
    if (userAction.action === 'retry' && incident.resolution_attempts >= 3) {
      validation.warnings.push('Incident has already been retried multiple times');
      validation.recommendations.push('Consider escalation or dismiss action');
    }
    
    return validation;
  }
  
  async checkEscalation(incident, error) {
    const rules = this.escalationRules.get(incident.type) || [];
    
    for (const rule of rules) {
      if (rule.criteria(incident, error)) {
        return {
          required: true,
          rule: rule.action,
          reason: 'Escalation criteria met',
          next_steps: rule.action.steps
        };
      }
    }
    
    return { required: false };
  }
  
  async verifyResolution(incidentId) {
    // Wait a moment for resolution to take effect
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    try {
      const response = await fetch(`/api/v1/incidents/${incidentId}`, {
        headers: { 'X-API-Key': this.apiKey }
      });
      
      if (!response.ok) {
        return { passed: false, reason: 'Could not fetch incident for verification' };
      }
      
      const incident = await response.json();
      const isResolved = ['RESOLVED', 'DISMISSED'].includes(incident.data.status);
      
      return {
        passed: isResolved,
        current_status: incident.data.status,
        reason: isResolved ? 'Incident successfully resolved' : 'Incident still open'
      };
      
    } catch (error) {
      return {
        passed: false,
        reason: `Verification failed: ${error.message}`
      };
    }
  }
}

// Настройка правил эскалации
const workflowManager = new ResolutionWorkflowManager('your-api-key');

// Правило эскалации для критических JOB инцидентов
workflowManager.addEscalationRule('JOB', 
  (incident, error) => incident.severity === 'CRITICAL' && incident.resolution_attempts >= 2,
  {
    type: 'IMMEDIATE_ESCALATION',
    steps: ['Notify on-call engineer', 'Create high-priority ticket', 'Alert management']
  }
);

// Правило эскалации для долгоживущих инцидентов
workflowManager.addEscalationRule('EXPRESSION',
  (incident, error) => {
    const age = Date.now() - new Date(incident.created_at).getTime();
    return age > 30 * 60 * 1000; // 30 minutes
  },
  {
    type: 'TIME_BASED_ESCALATION',
    steps: ['Review by senior developer', 'Check process definition']
  }
);
```

## Связанные endpoints
- [`GET /api/v1/incidents/:id`](./get-incident.md) - Просмотр деталей перед решением
- [`GET /api/v1/incidents`](./list-incidents.md) - Поиск инцидентов для решения
- [`GET /api/v1/incidents/stats`](./get-incident-stats.md) - Анализ эффективности решений
