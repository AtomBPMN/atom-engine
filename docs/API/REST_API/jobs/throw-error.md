# POST /api/v1/jobs/:key/throw-error

## Описание
Выброс BPMN ошибки из задания. Используется для инициации обработки ошибок в BPMN процессе через error boundary events или error handling.

## URL
```
POST /api/v1/jobs/{job_key}/throw-error
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры пути
- `job_key` (string): Ключ задания для выброса ошибки

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `error_code` (string): Код BPMN ошибки

### Опциональные поля
- `error_message` (string): Сообщение об ошибке
- `variables` (object): Переменные контекста ошибки

### Пример тела запроса
```json
{
  "error_code": "PAYMENT_DECLINED",
  "error_message": "Credit card payment was declined by the bank",
  "variables": {
    "errorType": "PAYMENT_DECLINED",
    "declineReason": "insufficient_funds",
    "cardLastFour": "4242",
    "attemptedAmount": 299.99,
    "bankResponseCode": "51",
    "retryable": false,
    "errorTimestamp": "2025-01-11T10:35:30.789Z"
  }
}
```

## Примеры запросов

### Ошибка платежа
```bash
curl -X POST "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/throw-error" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "error_code": "PAYMENT_DECLINED",
    "error_message": "Credit card payment was declined",
    "variables": {
      "errorType": "PAYMENT_DECLINED",
      "declineReason": "insufficient_funds",
      "retryable": false
    }
  }'
```

### Ошибка валидации
```bash
curl -X POST "http://localhost:27555/api/v1/jobs/srv1-job-abc123/throw-error" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "error_code": "VALIDATION_ERROR",
    "error_message": "Invalid customer data provided",
    "variables": {
      "errorType": "VALIDATION_ERROR",
      "invalidFields": ["email", "phoneNumber"],
      "correctionRequired": true
    }
  }'
```

### Системная ошибка
```bash
curl -X POST "http://localhost:27555/api/v1/jobs/srv1-job-def456/throw-error" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "error_code": "EXTERNAL_SERVICE_ERROR",
    "error_message": "External API is temporarily unavailable",
    "variables": {
      "errorType": "EXTERNAL_SERVICE_ERROR",
      "serviceName": "inventory-api",
      "httpStatus": 503,
      "retryAfter": "PT5M"
    }
  }'
```

### JavaScript
```javascript
const jobKey = 'srv1-job-xyz789';
const error = {
  error_code: 'BUSINESS_RULE_VIOLATION',
  error_message: 'Order amount exceeds customer credit limit',
  variables: {
    errorType: 'BUSINESS_RULE_VIOLATION',
    customerCreditLimit: 1000.00,
    requestedAmount: 1500.00,
    exceedsBy: 500.00,
    requiresApproval: true
  }
};

const response = await fetch(`/api/v1/jobs/${jobKey}/throw-error`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(error)
});

const result = await response.json();
```

### Go
```go
errorData := map[string]interface{}{
    "error_code":    "INVENTORY_UNAVAILABLE",
    "error_message": "Requested item is out of stock",
    "variables": map[string]interface{}{
        "errorType":     "INVENTORY_UNAVAILABLE",
        "requestedQty":  5,
        "availableQty":  0,
        "restockDate":   "2025-01-15",
        "alternativeProducts": []string{"PROD-002", "PROD-003"},
    },
}

jsonData, _ := json.Marshal(errorData)
req, _ := http.NewRequest("POST", "/api/v1/jobs/srv1-job-xyz789/throw-error", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Ошибка выброшена (с обработчиком)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "status": "ERROR_THROWN",
    "error_thrown_at": "2025-01-11T10:35:30.789Z",
    "thrown_by": "payment-worker-02",
    "error_details": {
      "error_code": "PAYMENT_DECLINED",
      "error_message": "Credit card payment was declined by the bank",
      "variables": {
        "errorType": "PAYMENT_DECLINED",
        "declineReason": "insufficient_funds",
        "cardLastFour": "4242",
        "attemptedAmount": 299.99,
        "retryable": false
      }
    },
    "process_continuation": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "process-payment",
      "error_handled": true,
      "error_boundary_triggered": {
        "boundary_event_id": "payment-error-boundary",
        "boundary_event_name": "Payment Error Handler",
        "error_code_matched": "PAYMENT_DECLINED",
        "triggered_at": "2025-01-11T10:35:30.790Z"
      },
      "next_activities": [
        {
          "element_id": "handle-payment-error",
          "element_type": "serviceTask",
          "element_name": "Handle Payment Error"
        }
      ]
    },
    "job_completion": {
      "completed_with_error": true,
      "execution_duration_ms": 240000,
      "error_variables_propagated": [
        "errorType",
        "declineReason", 
        "retryable"
      ]
    }
  },
  "request_id": "req_1641998403100"
}
```

### 200 OK - Ошибка выброшена (без обработчика)
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-abc123",
    "status": "ERROR_THROWN",
    "error_thrown_at": "2025-01-11T10:35:30.789Z",
    "thrown_by": "validation-worker-01",
    "error_details": {
      "error_code": "UNKNOWN_ERROR",
      "error_message": "Unexpected validation failure"
    },
    "process_continuation": {
      "process_instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
      "element_id": "validate-data",
      "error_handled": false,
      "no_error_boundary": true,
      "process_terminated": true,
      "termination_reason": "Unhandled BPMN error"
    },
    "incident_created": {
      "incident_id": "srv1-incident-unhandled-error",
      "incident_type": "UNHANDLED_BPMN_ERROR",
      "status": "OPEN",
      "created_at": "2025-01-11T10:35:30.791Z"
    },
    "job_completion": {
      "completed_with_error": true,
      "execution_duration_ms": 15000,
      "process_blocked": true
    }
  },
  "request_id": "req_1641998403101"
}
```

### 404 Not Found - Задание не найдено
```json
{
  "success": false,
  "error": {
    "code": "JOB_NOT_FOUND",
    "message": "Job not found or not accessible",
    "details": {
      "job_key": "srv1-job-nonexistent"
    }
  },
  "request_id": "req_1641998403102"
}
```

### 400 Bad Request - Неверный код ошибки
```json
{
  "success": false,
  "error": {
    "code": "INVALID_ERROR_CODE",
    "message": "Invalid BPMN error code format",
    "details": {
      "provided_code": "",
      "requirements": [
        "Error code must not be empty",
        "Error code should be alphanumeric with underscores",
        "Recommended format: CATEGORY_SPECIFIC_ERROR"
      ],
      "examples": [
        "PAYMENT_DECLINED",
        "VALIDATION_ERROR",
        "EXTERNAL_SERVICE_ERROR"
      ]
    }
  },
  "request_id": "req_1641998403103"
}
```

## Поля ответа

### Error Information
- `job_key` (string): Ключ задания
- `status` (string): Новый статус (`ERROR_THROWN`)
- `error_thrown_at` (string): Время выброса ошибки
- `thrown_by` (string): ID worker'а

### Error Details
- `error_code` (string): Код выброшенной ошибки
- `error_message` (string): Сообщение об ошибке
- `variables` (object): Переменные контекста ошибки

### Process Continuation
- `error_handled` (boolean): Была ли ошибка обработана
- `error_boundary_triggered` (object): Информация о сработавшем boundary event
- `next_activities` (array): Следующие активности процесса
- `process_terminated` (boolean): Был ли процесс завершен

### Incident Information (если ошибка не обработана)
- `incident_created` (object): Созданный инцидент
- `incident_id`, `incident_type`, `status`

## BPMN Error Handling

### Error Boundary Events
BPMN ошибки могут быть перехвачены error boundary events, прикрепленными к активностям:

```xml
<bpmn:serviceTask id="process-payment" name="Process Payment">
  <bpmn:extensionElements>
    <zeebe:taskDefinition type="payment-processor" />
  </bpmn:extensionElements>
</bpmn:serviceTask>

<bpmn:boundaryEvent id="payment-error-boundary" 
                    attachedToRef="process-payment">
  <bpmn:errorEventDefinition errorRef="PAYMENT_DECLINED" />
</bpmn:boundaryEvent>
```

### Error Propagation
Если error boundary event не найден:
1. Ошибка пропагируется на уровень выше
2. Если обработчика нет на верхнем уровне - создается инцидент
3. Процесс блокируется до решения инцидента

## Использование

### Payment Error Handling
```javascript
class PaymentWorker {
  async processPayment(job) {
    try {
      const paymentResult = await this.callPaymentGateway(job.variables);
      
      if (paymentResult.status === 'declined') {
        // Выбрасываем BPMN ошибку вместо обычного провала
        return await this.throwBPMNError(job.key, {
          error_code: 'PAYMENT_DECLINED',
          error_message: `Payment declined: ${paymentResult.decline_reason}`,
          variables: {
            errorType: 'PAYMENT_DECLINED',
            declineReason: paymentResult.decline_reason,
            bankCode: paymentResult.bank_response_code,
            retryable: this.isRetryableDecline(paymentResult.decline_reason),
            attemptedAmount: job.variables.amount,
            cardLastFour: job.variables.cardNumber?.slice(-4)
          }
        });
      }
      
      // Успешная обработка
      return await this.completeJob(job.key, paymentResult);
      
    } catch (error) {
      // Системная ошибка - выбрасываем соответствующую BPMN ошибку
      return await this.throwBPMNError(job.key, {
        error_code: 'PAYMENT_SYSTEM_ERROR',
        error_message: `Payment system error: ${error.message}`,
        variables: {
          errorType: 'PAYMENT_SYSTEM_ERROR',
          systemError: error.message,
          retryable: true,
          retryAfter: 'PT5M'
        }
      });
    }
  }
  
  async throwBPMNError(jobKey, errorInfo) {
    const response = await fetch(`/api/v1/jobs/${jobKey}/throw-error`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(errorInfo)
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log(`BPMN error thrown: ${errorInfo.error_code}`);
      
      if (result.data.process_continuation.error_handled) {
        console.log('Error handled by boundary event');
      } else {
        console.warn('Error not handled - incident created');
      }
    }
    
    return result;
  }
  
  isRetryableDecline(declineReason) {
    const retryableReasons = [
      'insufficient_funds',
      'temporary_hold',
      'try_again_later'
    ];
    return retryableReasons.includes(declineReason);
  }
}
```

### Validation Error Handler
```javascript
class ValidationWorker {
  async validateData(job) {
    const validation = this.validateInput(job.variables);
    
    if (!validation.isValid) {
      // Выбрасываем ошибку валидации
      return await fetch(`/api/v1/jobs/${job.key}/throw-error`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify({
          error_code: 'VALIDATION_ERROR',
          error_message: `Validation failed: ${validation.errors.join(', ')}`,
          variables: {
            errorType: 'VALIDATION_ERROR',
            invalidFields: validation.invalidFields,
            errors: validation.errors,
            correctionRequired: true,
            userFriendlyMessage: this.createUserMessage(validation.errors)
          }
        })
      });
    }
    
    // Валидация прошла успешно
    return await this.completeJob(job.key, {
      validationPassed: true,
      validatedAt: new Date().toISOString()
    });
  }
  
  validateInput(variables) {
    const errors = [];
    const invalidFields = [];
    
    if (!variables.email || !this.isValidEmail(variables.email)) {
      errors.push('Invalid email address');
      invalidFields.push('email');
    }
    
    if (!variables.phoneNumber || !this.isValidPhone(variables.phoneNumber)) {
      errors.push('Invalid phone number');
      invalidFields.push('phoneNumber');
    }
    
    if (!variables.age || variables.age < 18) {
      errors.push('Age must be 18 or older');
      invalidFields.push('age');
    }
    
    return {
      isValid: errors.length === 0,
      errors,
      invalidFields
    };
  }
  
  createUserMessage(errors) {
    if (errors.length === 1) {
      return `Please correct the following: ${errors[0]}`;
    } else {
      return `Please correct the following issues: ${errors.join(', ')}`;
    }
  }
}
```

### Generic Error Handler
```javascript
class ErrorHandlingWorker {
  async executeWithErrorHandling(job, businessLogic) {
    try {
      return await businessLogic(job);
      
    } catch (error) {
      const errorType = this.classifyError(error);
      
      switch (errorType.category) {
        case 'BUSINESS':
          return await this.throwBusinessError(job.key, error, errorType);
          
        case 'VALIDATION':
          return await this.throwValidationError(job.key, error, errorType);
          
        case 'EXTERNAL':
          return await this.throwExternalError(job.key, error, errorType);
          
        case 'SYSTEM':
          return await this.throwSystemError(job.key, error, errorType);
          
        default:
          return await this.throwGenericError(job.key, error);
      }
    }
  }
  
  classifyError(error) {
    if (error.code?.startsWith('BUSINESS_')) {
      return { category: 'BUSINESS', code: error.code };
    }
    
    if (error.code?.startsWith('VALIDATION_')) {
      return { category: 'VALIDATION', code: error.code };
    }
    
    if (error.code?.includes('TIMEOUT') || error.code?.includes('CONNECTION')) {
      return { category: 'EXTERNAL', code: 'EXTERNAL_SERVICE_ERROR' };
    }
    
    return { category: 'SYSTEM', code: 'SYSTEM_ERROR' };
  }
  
  async throwBusinessError(jobKey, error, errorType) {
    return await this.throwError(jobKey, {
      error_code: errorType.code,
      error_message: error.message,
      variables: {
        errorType: errorType.code,
        businessContext: error.context,
        userAction: error.userAction || 'Contact support',
        retryable: false
      }
    });
  }
  
  async throwExternalError(jobKey, error, errorType) {
    return await this.throwError(jobKey, {
      error_code: 'EXTERNAL_SERVICE_ERROR',
      error_message: `External service error: ${error.message}`,
      variables: {
        errorType: 'EXTERNAL_SERVICE_ERROR',
        serviceName: error.service,
        retryable: true,
        retryAfter: this.calculateRetryDelay(error),
        fallbackAction: 'Use alternative service'
      }
    });
  }
  
  async throwError(jobKey, errorInfo) {
    const response = await fetch(`/api/v1/jobs/${jobKey}/throw-error`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(errorInfo)
    });
    
    return await response.json();
  }
  
  calculateRetryDelay(error) {
    if (error.retryAfter) return error.retryAfter;
    if (error.code?.includes('RATE_LIMIT')) return 'PT5M';
    return 'PT1M';
  }
}
```

## Стандартные коды ошибок

### Business Errors
- `PAYMENT_DECLINED` - Платеж отклонен
- `INSUFFICIENT_FUNDS` - Недостаточно средств
- `BUSINESS_RULE_VIOLATION` - Нарушение бизнес-правил
- `APPROVAL_REQUIRED` - Требуется одобрение

### Validation Errors
- `VALIDATION_ERROR` - Общая ошибка валидации
- `INVALID_INPUT` - Неверные входные данные
- `MISSING_REQUIRED_FIELD` - Отсутствует обязательное поле
- `FORMAT_ERROR` - Неверный формат данных

### System Errors
- `EXTERNAL_SERVICE_ERROR` - Ошибка внешнего сервиса
- `TIMEOUT_ERROR` - Таймаут операции
- `SYSTEM_ERROR` - Системная ошибка
- `DATABASE_ERROR` - Ошибка базы данных

### Integration Errors
- `API_ERROR` - Ошибка API
- `NETWORK_ERROR` - Сетевая ошибка
- `AUTHENTICATION_ERROR` - Ошибка аутентификации
- `AUTHORIZATION_ERROR` - Ошибка авторизации

## Связанные endpoints
- [`PUT /api/v1/jobs/:key/complete`](./complete-job.md) - Завершить задание успешно
- [`PUT /api/v1/jobs/:key/fail`](./fail-job.md) - Провалить задание (технические ошибки)
- [`GET /api/v1/jobs/:key`](./get-job.md) - Детали задания
- [`GET /api/v1/incidents`](../incidents/list-incidents.md) - Инциденты от необработанных ошибок
