# GET /api/v1/incidents/:id

## Описание
Получение детальной информации о конкретном инциденте, включая полную диагностическую информацию и историю решений.

## URL
```
GET /api/v1/incidents/:id
```

## Авторизация
✅ **Требуется API ключ** с разрешением `incident`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор инцидента

## Примеры запросов

### Получение инцидента
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/srv1-inc123abc456def789" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/incidents/srv1-inc123abc456def789', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const incident = await response.json();
```

## Ответы

### 200 OK - Детали инцидента (Job инцидент)
```json
{
  "success": true,
  "data": {
    "id": "srv1-inc123abc456def789",
    "type": "JOB",
    "status": "OPEN",
    "message": "Job execution failed after maximum retries",
    "created_at": "2025-01-11T10:15:30.123Z",
    "updated_at": "2025-01-11T10:15:30.123Z",
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "process_definition_key": "Order_Processing",
    "process_definition_version": 3,
    "element_id": "ServiceTask_PaymentProcess",
    "element_name": "Process Payment",
    "element_type": "SERVICE_TASK",
    "job_key": "srv1-job789xyz123abc456",
    "job_type": "payment-service",
    "worker": "payment-worker-01",
    "retries_left": 0,
    "max_retries": 3,
    "severity": "HIGH",
    "tenant_id": "tenant-001",
    "error_details": {
      "error_code": "PAYMENT_SERVICE_UNAVAILABLE",
      "error_message": "Payment service timeout after 30 seconds",
      "stack_trace": "PaymentException: Connection timeout\n  at PaymentService.process(PaymentRequest.java:45)\n  at PaymentWorker.execute(PaymentWorker.java:123)",
      "custom_headers": {
        "x-correlation-id": "corr-123456789",
        "x-tenant-id": "tenant-001",
        "x-retry-count": "3"
      },
      "variables_at_failure": {
        "orderId": "ORD-2025-001234",
        "amount": 299.99,
        "currency": "USD",
        "customerId": "CUST-789456",
        "paymentMethod": "credit_card"
      }
    },
    "execution_context": {
      "bpmn_element_path": "StartEvent_1 -> ServiceTask_ValidateOrder -> ServiceTask_PaymentProcess",
      "active_tokens": ["srv1-token456def789ghi012"],
      "failed_at_step": "ServiceTask_PaymentProcess",
      "execution_attempts": [
        {
          "attempt": 1,
          "timestamp": "2025-01-11T10:10:15.123Z",
          "duration_ms": 30000,
          "error": "Connection timeout"
        },
        {
          "attempt": 2,
          "timestamp": "2025-01-11T10:12:30.456Z",
          "duration_ms": 30000,
          "error": "Connection timeout"
        },
        {
          "attempt": 3,
          "timestamp": "2025-01-11T10:15:00.789Z",
          "duration_ms": 30000,
          "error": "Connection timeout"
        }
      ]
    },
    "resolution_attempts": 0,
    "related_incidents": [
      {
        "id": "srv1-inc456def789ghi012",
        "type": "JOB",
        "message": "Similar payment service failure",
        "created_at": "2025-01-11T09:30:22.456Z",
        "status": "RESOLVED"
      }
    ],
    "diagnostics": {
      "impact_assessment": {
        "blocked_process_instances": 1,
        "affected_customers": 1,
        "business_impact": "HIGH",
        "estimated_revenue_impact": 299.99
      },
      "system_health": {
        "service_availability": "DEGRADED",
        "error_rate_spike": true,
        "similar_failures_last_hour": 3
      },
      "recommended_actions": [
        "Check payment service health",
        "Verify network connectivity",
        "Consider increasing timeout configuration",
        "Retry incident with corrected configuration"
      ]
    }
  },
  "request_id": "req_1641998404500"
}
```

### 200 OK - Expression инцидент
```json
{
  "success": true,
  "data": {
    "id": "srv1-inc456def789ghi012",
    "type": "EXPRESSION",
    "status": "RESOLVED",
    "message": "Invalid expression syntax in gateway condition",
    "created_at": "2025-01-11T09:45:15.456Z",
    "updated_at": "2025-01-11T10:30:22.789Z",
    "resolved_at": "2025-01-11T10:30:22.789Z",
    "process_instance_id": "srv1-cD6fGh9iJ3kL7mN0pR",
    "process_definition_key": "Invoice_Approval",
    "process_definition_version": 2,
    "element_id": "ExclusiveGateway_AmountCheck",
    "element_name": "Amount Decision Gateway",
    "element_type": "EXCLUSIVE_GATEWAY",
    "severity": "MEDIUM",
    "tenant_id": "tenant-001",
    "error_details": {
      "error_code": "EXPRESSION_SYNTAX_ERROR",
      "error_message": "Unexpected token ')' at position 15",
      "expression": "amount > 1000 and )",
      "expected_expression": "amount > 1000 and approved = true",
      "position": 15,
      "context_variables": {
        "amount": 1500,
        "approved": true,
        "submittedBy": "john.doe@company.com"
      },
      "suggestion": "Complete the expression or remove the extra parenthesis"
    },
    "execution_context": {
      "bpmn_element_path": "StartEvent_1 -> UserTask_ReviewInvoice -> ExclusiveGateway_AmountCheck",
      "gateway_conditions": [
        {
          "sequence_flow_id": "SequenceFlow_ToApproval",
          "condition": "amount > 1000 and )",
          "valid": false
        },
        {
          "sequence_flow_id": "SequenceFlow_ToAutoApprove", 
          "condition": "amount <= 1000",
          "valid": true
        }
      ],
      "failed_evaluation": {
        "expression": "amount > 1000 and )",
        "variables": {
          "amount": 1500
        },
        "error_position": 15
      }
    },
    "resolution": {
      "type": "RETRY",
      "resolved_by": "system",
      "resolved_at": "2025-01-11T10:30:22.789Z",
      "resolution_comment": "Expression corrected automatically by removing extra parenthesis",
      "retries_performed": 1,
      "corrected_expression": "amount > 1000 and approved = true",
      "resolution_details": {
        "method": "AUTO_CORRECTION",
        "confidence": 0.95,
        "manual_intervention": false
      }
    },
    "resolution_attempts": 1,
    "diagnostics": {
      "impact_assessment": {
        "blocked_process_instances": 1,
        "affected_users": 1,
        "business_impact": "MEDIUM"
      },
      "fix_effectiveness": {
        "auto_fix_applied": true,
        "fix_success_rate": 1.0,
        "similar_issues_prevented": 2
      }
    }
  },
  "request_id": "req_1641998404501"
}
```

### 404 Not Found - Инцидент не найден
```json
{
  "success": false,
  "error": {
    "code": "INCIDENT_NOT_FOUND",
    "message": "Incident with id 'invalid-incident-id' not found",
    "details": {
      "incident_id": "invalid-incident-id",
      "suggestions": [
        "Verify the incident ID is correct",
        "Check if the incident was deleted",
        "Ensure you have permission to access this incident"
      ]
    }
  },
  "request_id": "req_1641998404502"
}
```

## Использование

### Incident Detail Viewer
```javascript
class IncidentDetailViewer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getIncident(incidentId) {
    const response = await fetch(`/api/v1/incidents/${incidentId}`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch incident: ${response.statusText}`);
    }
    
    return await response.json();
  }
  
  async getIncidentWithContext(incidentId) {
    const incident = await this.getIncident(incidentId);
    
    // Enrich with additional context
    const enrichedIncident = {
      ...incident.data,
      formatted_details: this.formatIncidentDetails(incident.data),
      troubleshooting_guide: this.generateTroubleshootingGuide(incident.data),
      similar_patterns: await this.findSimilarIncidents(incident.data)
    };
    
    return enrichedIncident;
  }
  
  formatIncidentDetails(incident) {
    const formatted = {
      summary: this.createSummary(incident),
      timeline: this.createTimeline(incident),
      technical_details: this.formatTechnicalDetails(incident),
      impact_analysis: this.formatImpactAnalysis(incident)
    };
    
    return formatted;
  }
  
  createSummary(incident) {
    return {
      title: `${incident.type} Incident: ${incident.element_name || incident.element_id}`,
      description: incident.message,
      severity: incident.severity,
      status: incident.status,
      duration: incident.resolved_at ? 
        this.calculateDuration(incident.created_at, incident.resolved_at) : 
        this.calculateDuration(incident.created_at, new Date().toISOString()),
      affected_process: incident.process_definition_key,
      affected_instance: incident.process_instance_id
    };
  }
  
  createTimeline(incident) {
    const timeline = [{
      timestamp: incident.created_at,
      event: 'Incident Created',
      description: incident.message,
      type: 'error'
    }];
    
    // Add execution attempts for job incidents
    if (incident.execution_context?.execution_attempts) {
      incident.execution_context.execution_attempts.forEach((attempt, index) => {
        timeline.push({
          timestamp: attempt.timestamp,
          event: `Execution Attempt ${attempt.attempt}`,
          description: `Failed after ${attempt.duration_ms}ms: ${attempt.error}`,
          type: 'retry'
        });
      });
    }
    
    // Add resolution if exists
    if (incident.resolution) {
      timeline.push({
        timestamp: incident.resolved_at,
        event: 'Incident Resolved',
        description: incident.resolution.resolution_comment,
        type: 'success'
      });
    }
    
    return timeline.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
  }
  
  formatTechnicalDetails(incident) {
    const details = {
      element_info: {
        id: incident.element_id,
        name: incident.element_name,
        type: incident.element_type,
        process_path: incident.execution_context?.bpmn_element_path
      },
      error_info: incident.error_details,
      system_context: {
        process_instance: incident.process_instance_id,
        process_version: incident.process_definition_version,
        tenant: incident.tenant_id
      }
    };
    
    if (incident.type === 'JOB') {
      details.job_info = {
        key: incident.job_key,
        type: incident.job_type,
        worker: incident.worker,
        retries: `${incident.max_retries - incident.retries_left}/${incident.max_retries}`
      };
    }
    
    return details;
  }
  
  formatImpactAnalysis(incident) {
    const impact = incident.diagnostics?.impact_assessment || {};
    
    return {
      business_impact: impact.business_impact || 'UNKNOWN',
      affected_instances: impact.blocked_process_instances || 0,
      affected_users: impact.affected_users || impact.affected_customers || 0,
      estimated_cost: impact.estimated_revenue_impact || 0,
      service_health: incident.diagnostics?.system_health?.service_availability || 'UNKNOWN'
    };
  }
  
  generateTroubleshootingGuide(incident) {
    const guides = {
      JOB: this.generateJobTroubleshootingGuide(incident),
      EXPRESSION: this.generateExpressionTroubleshootingGuide(incident),
      BPMN: this.generateBPMNTroubleshootingGuide(incident),
      PROCESS: this.generateProcessTroubleshootingGuide(incident),
      TIMER: this.generateTimerTroubleshootingGuide(incident),
      MESSAGE: this.generateMessageTroubleshootingGuide(incident),
      SYSTEM: this.generateSystemTroubleshootingGuide(incident)
    };
    
    return guides[incident.type] || this.generateGenericTroubleshootingGuide(incident);
  }
  
  generateJobTroubleshootingGuide(incident) {
    const steps = [
      'Check external service availability',
      'Verify network connectivity',
      'Review job worker logs',
      'Check authentication credentials',
      'Validate input data format'
    ];
    
    if (incident.error_details?.error_code) {
      switch (incident.error_details.error_code) {
        case 'PAYMENT_SERVICE_UNAVAILABLE':
          steps.unshift('Check payment service status dashboard');
          steps.push('Consider fallback payment method');
          break;
        case 'AUTHENTICATION_FAILED':
          steps.unshift('Verify API credentials');
          steps.push('Check token expiration');
          break;
      }
    }
    
    return {
      immediate_actions: steps,
      diagnostic_commands: [
        `curl -I ${incident.error_details?.service_url || 'service-url'}`,
        'Check worker logs for correlation ID: ' + incident.error_details?.custom_headers?.['x-correlation-id']
      ],
      escalation_criteria: [
        'Service down for > 5 minutes',
        'Multiple similar incidents',
        'High-value transaction affected'
      ]
    };
  }
  
  generateExpressionTroubleshootingGuide(incident) {
    return {
      immediate_actions: [
        'Review expression syntax',
        'Check variable availability',
        'Validate data types',
        'Test expression in isolation'
      ],
      diagnostic_steps: [
        'Use expression validation endpoint',
        'Check context variables',
        'Verify FEEL syntax compliance'
      ],
      common_fixes: [
        'Add null checks: variable?.property',
        'Ensure proper parentheses matching',
        'Verify function names and parameters'
      ]
    };
  }
  
  calculateDuration(startTime, endTime) {
    const start = new Date(startTime);
    const end = new Date(endTime);
    const durationMs = end - start;
    
    const hours = Math.floor(durationMs / (1000 * 60 * 60));
    const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((durationMs % (1000 * 60)) / 1000);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  }
  
  async findSimilarIncidents(incident) {
    // This would call the incidents list endpoint to find similar patterns
    try {
      const response = await fetch(`/api/v1/incidents?type=${incident.type}&limit=10`, {
        headers: { 'X-API-Key': this.apiKey }
      });
      
      const result = await response.json();
      
      return result.data.incidents
        .filter(other => other.id !== incident.id)
        .filter(other => this.calculateSimilarity(incident, other) > 0.7)
        .slice(0, 5);
        
    } catch (error) {
      console.error('Failed to fetch similar incidents:', error);
      return [];
    }
  }
  
  calculateSimilarity(incident1, incident2) {
    let score = 0;
    let factors = 0;
    
    // Same element
    if (incident1.element_id === incident2.element_id) {
      score += 0.4;
    }
    factors++;
    
    // Same error code
    if (incident1.error_details?.error_code === incident2.error_details?.error_code) {
      score += 0.3;
    }
    factors++;
    
    // Same process
    if (incident1.process_definition_key === incident2.process_definition_key) {
      score += 0.2;
    }
    factors++;
    
    // Same type
    if (incident1.type === incident2.type) {
      score += 0.1;
    }
    factors++;
    
    return score;
  }
}

// Использование
const viewer = new IncidentDetailViewer('your-api-key');

// Получение детальной информации об инциденте
const incident = await viewer.getIncidentWithContext('srv1-inc123abc456def789');
console.log('Incident details:', incident);

// Форматированная информация
console.log('Summary:', incident.formatted_details.summary);
console.log('Timeline:', incident.formatted_details.timeline);
console.log('Troubleshooting:', incident.troubleshooting_guide);
```

## Связанные endpoints
- [`GET /api/v1/incidents`](./list-incidents.md) - Список всех инцидентов
- [`POST /api/v1/incidents/:id/resolve`](./resolve-incident.md) - Решение инцидента
- [`GET /api/v1/incidents/stats`](./get-incident-stats.md) - Статистика инцидентов
