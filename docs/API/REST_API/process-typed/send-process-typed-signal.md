# POST /api/v1/processes/:id/typed/signal

## Описание
Отправка типизированного сигнала процессу с валидацией данных и автоматической маршрутизацией к соответствующим элементам процесса.

## URL
```
POST /api/v1/processes/:id/typed/signal
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры тела запроса

### Обязательные поля
- `signal_name` (string): Имя сигнала для отправки
- `signal_type` (string): Тип сигнала (`message`, `signal`, `timer`, `error`)

### Опциональные поля
- `data` (object): Данные сигнала (будут валидированы)
- `correlation_key` (string): Ключ корреляции для точной маршрутизации
- `target_activity` (string): Целевая активность (element_id)
- `validate_data` (boolean): Валидировать данные сигнала (по умолчанию: true)
- `async_delivery` (boolean): Асинхронная доставка (по умолчанию: false)

## Примеры запросов

### Отправка сигнала сообщения
```bash
curl -X POST "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/signal" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "signal_name": "payment_completed",
    "signal_type": "message",
    "data": {
      "transactionId": "txn_1234567890",
      "amount": 1050.99,
      "currency": "USD",
      "status": "COMPLETED",
      "timestamp": "2025-01-11T11:45:30.123Z"
    },
    "correlation_key": "ORD-2025-001234"
  }'
```

### Отправка сигнала ошибки
```bash
curl -X POST "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/signal" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "signal_name": "payment_failed",
    "signal_type": "error",
    "data": {
      "errorCode": "PAYMENT_DECLINED",
      "errorMessage": "Insufficient funds",
      "retryable": true,
      "retryAfter": 300
    },
    "target_activity": "ServiceTask_PaymentProcess"
  }'
```

### JavaScript
```javascript
const signalRequest = {
  signal_name: 'approval_received',
  signal_type: 'message',
  data: {
    approver: 'manager@company.com',
    approved: true,
    approvalDate: '2025-01-11T12:00:00.123Z',
    comments: 'Order approved - customer is VIP'
  },
  correlation_key: 'ORD-2025-001234',
  target_activity: 'UserTask_ManagerApproval'
};

const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/signal', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(signalRequest)
});

const result = await response.json();
```

## Ответы

### 200 OK - Сигнал доставлен
```json
{
  "success": true,
  "data": {
    "signal_id": "srv1-sig789abc123def456",
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "signal_name": "payment_completed",
    "signal_type": "message",
    "sent_at": "2025-01-11T11:45:35.456Z",
    "delivery_status": "DELIVERED",
    "delivery_details": {
      "correlation_matched": true,
      "target_activities_found": ["IntermediateCatchEvent_PaymentResult"],
      "activities_triggered": [
        {
          "element_id": "IntermediateCatchEvent_PaymentResult",
          "element_name": "Payment Result",
          "triggered_at": "2025-01-11T11:45:35.456Z",
          "token_id": "srv1-token123abc456def789"
        }
      ],
      "variables_updated": {
        "paymentResult": {
          "transactionId": "txn_1234567890",
          "amount": 1050.99,
          "currency": "USD", 
          "status": "COMPLETED",
          "timestamp": "2025-01-11T11:45:30.123Z"
        },
        "paymentStatus": "COMPLETED",
        "lastUpdated": "2025-01-11T11:45:35.456Z"
      }
    },
    "validation_results": {
      "data_valid": true,
      "schema_matched": true,
      "warnings": [],
      "transformations_applied": [
        {
          "field": "timestamp",
          "from": "string",
          "to": "date-time",
          "reason": "Auto-conversion to proper date format"
        }
      ]
    },
    "routing_info": {
      "correlation_key_used": "ORD-2025-001234",
      "routing_strategy": "CORRELATION_KEY_MATCH",
      "matched_subscriptions": [
        {
          "subscription_id": "srv1-sub456def789ghi012",
          "element_id": "IntermediateCatchEvent_PaymentResult",
          "correlation_expression": "orderId"
        }
      ]
    },
    "business_context": {
      "operation": "Payment Completion Signal",
      "business_impact": "Order can proceed to fulfillment",
      "next_expected_activities": ["ServiceTask_OrderFulfillment"],
      "milestone_reached": "Payment Processing Complete"
    }
  },
  "request_id": "req_1641998405400"
}
```

### 200 OK - Асинхронная доставка
```json
{
  "success": true,
  "data": {
    "signal_id": "srv1-sig890def123ghi456",
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "signal_name": "inventory_update",
    "signal_type": "message",
    "sent_at": "2025-01-11T12:00:00.123Z",
    "delivery_status": "QUEUED",
    "delivery_details": {
      "async_delivery": true,
      "estimated_delivery_time": "2025-01-11T12:00:05.000Z",
      "queue_position": 3,
      "delivery_tracking_id": "srv1-track456def789abc123"
    },
    "validation_results": {
      "data_valid": true,
      "schema_matched": true,
      "warnings": []
    }
  },
  "request_id": "req_1641998405401"
}
```

### 404 Not Found - Нет подписчиков
```json
{
  "success": false,
  "error": {
    "code": "NO_SIGNAL_SUBSCRIBERS",
    "message": "No activities are waiting for this signal",
    "details": {
      "signal_name": "unknown_signal",
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "available_signals": [
        {
          "signal_name": "payment_completed",
          "element_id": "IntermediateCatchEvent_PaymentResult",
          "element_name": "Payment Result"
        },
        {
          "signal_name": "approval_received",
          "element_id": "UserTask_ManagerApproval", 
          "element_name": "Manager Approval"
        }
      ],
      "suggestions": [
        "Check signal name spelling",
        "Verify process is waiting for this signal",
        "Use one of the available signals listed above"
      ]
    }
  },
  "request_id": "req_1641998405402"
}
```

### 400 Bad Request - Валидационные ошибки
```json
{
  "success": false,
  "error": {
    "code": "SIGNAL_VALIDATION_FAILED",
    "message": "Signal data validation failed",
    "details": {
      "signal_name": "payment_completed",
      "validation_errors": [
        {
          "field": "amount",
          "value": -100,
          "error": "Amount must be positive",
          "expected_type": "number",
          "validation_rule": "minimum"
        },
        {
          "field": "currency",
          "value": "INVALID",
          "error": "Currency must be one of: USD, EUR, GBP",
          "validation_rule": "enum"
        },
        {
          "field": "transactionId",
          "value": "",
          "error": "Transaction ID is required",
          "validation_rule": "required"
        }
      ],
      "expected_schema": {
        "transactionId": {"type": "string", "required": true},
        "amount": {"type": "number", "minimum": 0},
        "currency": {"type": "string", "enum": ["USD", "EUR", "GBP"]},
        "status": {"type": "string", "enum": ["COMPLETED", "FAILED", "PENDING"]}
      }
    }
  },
  "request_id": "req_1641998405403"
}
```

## Signal Types

### MESSAGE
Доставка данных к промежуточным событиям перехвата сообщений

### SIGNAL
Широковещательный сигнал для событий сигналов

### TIMER
Активация таймерных событий (обычно используется системой)

### ERROR
Активация граничных событий ошибок

## Использование

### Typed Signal Client
```javascript
class TypedSignalClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async sendSignal(processInstanceId, signalName, signalType, options = {}) {
    const requestBody = {
      signal_name: signalName,
      signal_type: signalType,
      data: options.data,
      correlation_key: options.correlationKey,
      target_activity: options.targetActivity,
      validate_data: options.validateData !== false,
      async_delivery: options.asyncDelivery || false
    };
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/signal`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify(requestBody)
      }
    );
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(`Signal failed: ${result.error?.message || response.statusText}`);
    }
    
    return result;
  }
  
  async sendMessage(processInstanceId, messageName, data, correlationKey) {
    return await this.sendSignal(processInstanceId, messageName, 'message', {
      data,
      correlationKey
    });
  }
  
  async sendError(processInstanceId, errorCode, errorData, targetActivity) {
    return await this.sendSignal(processInstanceId, errorCode, 'error', {
      data: errorData,
      targetActivity
    });
  }
  
  async sendApproval(processInstanceId, approved, approverInfo, correlationKey) {
    const approvalData = {
      approved,
      approver: approverInfo.email,
      approverName: approverInfo.name,
      approvalDate: new Date().toISOString(),
      comments: approverInfo.comments || ''
    };
    
    return await this.sendMessage(processInstanceId, 'approval_received', approvalData, correlationKey);
  }
  
  async sendPaymentResult(processInstanceId, paymentResult, correlationKey) {
    const signalName = paymentResult.status === 'COMPLETED' ? 'payment_completed' : 'payment_failed';
    
    return await this.sendMessage(processInstanceId, signalName, paymentResult, correlationKey);
  }
  
  async getAvailableSignals(processInstanceId) {
    // This would typically be a separate GET endpoint, but we'll simulate it
    try {
      // Try sending a dummy signal to get available signals in error response
      await this.sendSignal(processInstanceId, 'dummy_signal_to_get_available', 'message');
    } catch (error) {
      // Parse error response to extract available signals
      if (error.message.includes('NO_SIGNAL_SUBSCRIBERS')) {
        // In a real implementation, this would be parsed from the error response
        return [];
      }
    }
    
    return [];
  }
  
  async batchSendSignals(processInstanceId, signals) {
    const results = [];
    
    for (const signal of signals) {
      try {
        const result = await this.sendSignal(
          processInstanceId,
          signal.name,
          signal.type,
          signal.options
        );
        
        results.push({
          success: true,
          signal: signal.name,
          result
        });
        
      } catch (error) {
        results.push({
          success: false,
          signal: signal.name,
          error: error.message
        });
      }
    }
    
    return results;
  }
  
  async validateSignalData(signalName, data, schema) {
    // Client-side validation before sending
    const validationResults = {
      valid: true,
      errors: []
    };
    
    if (!schema || !schema.properties) {
      return validationResults;
    }
    
    Object.entries(data || {}).forEach(([field, value]) => {
      const fieldSchema = schema.properties[field];
      if (fieldSchema) {
        const fieldValidation = this.validateFieldValue(value, fieldSchema);
        if (!fieldValidation.valid) {
          validationResults.valid = false;
          validationResults.errors.push({
            field,
            value,
            errors: fieldValidation.errors
          });
        }
      }
    });
    
    // Check required fields
    if (schema.required) {
      schema.required.forEach(requiredField => {
        if (!(requiredField in (data || {}))) {
          validationResults.valid = false;
          validationResults.errors.push({
            field: requiredField,
            error: 'Required field is missing'
          });
        }
      });
    }
    
    return validationResults;
  }
  
  validateFieldValue(value, schema) {
    const validation = { valid: true, errors: [] };
    
    // Type validation
    if (schema.type && typeof value !== schema.type) {
      validation.valid = false;
      validation.errors.push(`Expected ${schema.type}, got ${typeof value}`);
    }
    
    // Enum validation
    if (schema.enum && !schema.enum.includes(value)) {
      validation.valid = false;
      validation.errors.push(`Value must be one of: ${schema.enum.join(', ')}`);
    }
    
    // Number validations
    if (schema.type === 'number') {
      if (schema.minimum !== undefined && value < schema.minimum) {
        validation.valid = false;
        validation.errors.push(`Minimum value is ${schema.minimum}`);
      }
      
      if (schema.maximum !== undefined && value > schema.maximum) {
        validation.valid = false;
        validation.errors.push(`Maximum value is ${schema.maximum}`);
      }
    }
    
    // String validations
    if (schema.type === 'string') {
      if (schema.minLength && value.length < schema.minLength) {
        validation.valid = false;
        validation.errors.push(`Minimum length is ${schema.minLength}`);
      }
      
      if (schema.pattern && !new RegExp(schema.pattern).test(value)) {
        validation.valid = false;
        validation.errors.push('Value does not match required pattern');
      }
    }
    
    return validation;
  }
}

// Использование
const signalClient = new TypedSignalClient('your-api-key');

// Отправка сообщения о завершении платежа
const paymentResult = {
  transactionId: 'txn_1234567890',
  amount: 1050.99,
  currency: 'USD',
  status: 'COMPLETED',
  timestamp: new Date().toISOString()
};

const result = await signalClient.sendPaymentResult(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  paymentResult,
  'ORD-2025-001234'
);

console.log('Payment signal sent:', result);

// Отправка сигнала об одобрении
const approverInfo = {
  email: 'manager@company.com',
  name: 'John Manager',
  comments: 'Approved - customer is VIP'
};

const approvalResult = await signalClient.sendApproval(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  true,
  approverInfo,
  'ORD-2025-001234'
);

console.log('Approval signal sent:', approvalResult);

// Отправка сигнала об ошибке
const errorData = {
  errorCode: 'PAYMENT_DECLINED',
  errorMessage: 'Insufficient funds',
  retryable: true,
  retryAfter: 300
};

const errorResult = await signalClient.sendError(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  'PAYMENT_ERROR',
  errorData,
  'ServiceTask_PaymentProcess'
);

// Пакетная отправка сигналов
const signals = [
  {
    name: 'inventory_updated',
    type: 'message',
    options: {
      data: { productId: 'PROD-001', stock: 45 },
      correlationKey: 'ORD-2025-001234'
    }
  },
  {
    name: 'shipping_notification',
    type: 'message',
    options: {
      data: { trackingNumber: 'TRACK-123456', carrier: 'UPS' },
      correlationKey: 'ORD-2025-001234'
    }
  }
];

const batchResults = await signalClient.batchSendSignals('srv1-aB3dEf9hK2mN5pQ8uV', signals);
console.log('Batch signal results:', batchResults);
```

### Signal Automation Framework
```javascript
class SignalAutomationFramework {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.signalClient = new TypedSignalClient(apiKey);
    this.automationRules = new Map();
  }
  
  addAutomationRule(triggerCondition, signalConfig) {
    const ruleId = `rule_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    this.automationRules.set(ruleId, {
      trigger: triggerCondition,
      signal: signalConfig,
      created_at: new Date().toISOString(),
      executions: 0
    });
    
    return ruleId;
  }
  
  async processEvent(event) {
    const triggeredRules = [];
    
    for (const [ruleId, rule] of this.automationRules) {
      if (this.evaluateTrigger(rule.trigger, event)) {
        try {
          const result = await this.executeSignal(rule.signal, event);
          
          triggeredRules.push({
            ruleId,
            success: true,
            result
          });
          
          rule.executions++;
          
        } catch (error) {
          triggeredRules.push({
            ruleId,
            success: false,
            error: error.message
          });
        }
      }
    }
    
    return triggeredRules;
  }
  
  evaluateTrigger(trigger, event) {
    // Simple trigger evaluation - in real implementation would be more sophisticated
    if (trigger.eventType && event.type !== trigger.eventType) {
      return false;
    }
    
    if (trigger.conditions) {
      return trigger.conditions.every(condition => {
        const eventValue = this.getEventValue(event, condition.field);
        
        switch (condition.operator) {
          case 'equals':
            return eventValue === condition.value;
          case 'greater_than':
            return eventValue > condition.value;
          case 'less_than':
            return eventValue < condition.value;
          case 'contains':
            return String(eventValue).includes(condition.value);
          default:
            return false;
        }
      });
    }
    
    return true;
  }
  
  getEventValue(event, fieldPath) {
    return fieldPath.split('.').reduce((obj, key) => obj?.[key], event);
  }
  
  async executeSignal(signalConfig, event) {
    const processInstanceId = this.extractProcessInstanceId(event, signalConfig);
    
    if (!processInstanceId) {
      throw new Error('Cannot determine process instance ID from event');
    }
    
    const signalData = this.buildSignalData(signalConfig.dataTemplate, event);
    const correlationKey = this.buildCorrelationKey(signalConfig.correlationTemplate, event);
    
    return await this.signalClient.sendSignal(
      processInstanceId,
      signalConfig.signalName,
      signalConfig.signalType,
      {
        data: signalData,
        correlationKey,
        targetActivity: signalConfig.targetActivity
      }
    );
  }
  
  extractProcessInstanceId(event, signalConfig) {
    if (signalConfig.processInstanceIdField) {
      return this.getEventValue(event, signalConfig.processInstanceIdField);
    }
    
    // Default extraction logic
    return event.processInstanceId || event.process_instance_id;
  }
  
  buildSignalData(template, event) {
    if (!template) return event.data || {};
    
    const result = {};
    
    Object.entries(template).forEach(([key, valueTemplate]) => {
      if (typeof valueTemplate === 'string' && valueTemplate.startsWith('${') && valueTemplate.endsWith('}')) {
        // Template variable: ${event.field.path}
        const fieldPath = valueTemplate.slice(2, -1).replace('event.', '');
        result[key] = this.getEventValue(event, fieldPath);
      } else {
        result[key] = valueTemplate;
      }
    });
    
    return result;
  }
  
  buildCorrelationKey(template, event) {
    if (!template) return null;
    
    if (typeof template === 'string' && template.startsWith('${') && template.endsWith('}')) {
      const fieldPath = template.slice(2, -1).replace('event.', '');
      return this.getEventValue(event, fieldPath);
    }
    
    return template;
  }
}

// Настройка автоматизации
const automation = new SignalAutomationFramework('your-api-key');

// Правило: при завершении платежа отправить сигнал
automation.addAutomationRule(
  {
    eventType: 'payment_completed',
    conditions: [
      { field: 'status', operator: 'equals', value: 'SUCCESS' },
      { field: 'amount', operator: 'greater_than', value: 0 }
    ]
  },
  {
    signalName: 'payment_completed',
    signalType: 'message',
    processInstanceIdField: 'orderId',
    dataTemplate: {
      transactionId: '${transactionId}',
      amount: '${amount}',
      currency: '${currency}',
      status: 'COMPLETED'
    },
    correlationTemplate: '${orderId}'
  }
);

// Обработка события
const paymentEvent = {
  type: 'payment_completed',
  transactionId: 'txn_1234567890',
  orderId: 'srv1-aB3dEf9hK2mN5pQ8uV',
  amount: 1050.99,
  currency: 'USD',
  status: 'SUCCESS'
};

const automationResults = await automation.processEvent(paymentEvent);
console.log('Automation results:', automationResults);
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/typed/activities`](./get-process-typed-activities.md) - Активности, ожидающие сигналы
- [`GET /api/v1/messages`](../messages/list-messages.md) - Сообщения и подписки
- [`POST /api/v1/messages`](../messages/publish-message.md) - Альтернативная отправка сообщений
