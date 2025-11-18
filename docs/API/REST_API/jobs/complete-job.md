# PUT /api/v1/jobs/:key/complete

## Описание
Завершение задания с передачей результата. Используется worker'ами для сообщения об успешном выполнении задания.

## URL
```
PUT /api/v1/jobs/{job_key}/complete
```

## Авторизация
✅ **Требуется API ключ** с разрешением `job`

## Параметры пути
- `job_key` (string): Ключ задания для завершения

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Опциональные поля
- `variables` (object): Переменные результата выполнения задания

### Пример тела запроса
```json
{
  "variables": {
    "paymentStatus": "completed",
    "transactionId": "txn_abc123def456",
    "amount": 299.99,
    "processingTime": 2500,
    "provider": "stripe",
    "fees": 8.97,
    "netAmount": 291.02,
    "receipt": {
      "url": "https://receipts.example.com/txn_abc123def456",
      "id": "rcpt_789xyz"
    }
  }
}
```

## Примеры запросов

### Завершение с результатом
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/complete" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "variables": {
      "paymentStatus": "completed",
      "transactionId": "txn_abc123def456",
      "amount": 299.99
    }
  }'
```

### Завершение без результата
```bash
curl -X PUT "http://localhost:27555/api/v1/jobs/srv1-job-xyz789/complete" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{}'
```

### JavaScript
```javascript
const jobKey = 'srv1-job-xyz789';
const result = {
  variables: {
    paymentStatus: 'completed',
    transactionId: 'txn_abc123def456',
    amount: 299.99,
    processingTime: 2500
  }
};

const response = await fetch(`/api/v1/jobs/${jobKey}/complete`, {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(result)
});

const completionResult = await response.json();
```

### Go
```go
result := map[string]interface{}{
    "variables": map[string]interface{}{
        "paymentStatus":  "completed",
        "transactionId":  "txn_abc123def456",
        "amount":         299.99,
        "processingTime": 2500,
    },
}

jsonData, _ := json.Marshal(result)
req, _ := http.NewRequest("PUT", "/api/v1/jobs/srv1-job-xyz789/complete", bytes.NewBuffer(jsonData))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Задание завершено
```json
{
  "success": true,
  "data": {
    "job_key": "srv1-job-xyz789",
    "status": "COMPLETED",
    "completed_at": "2025-01-11T10:35:30.789Z",
    "completed_by": "payment-worker-02",
    "execution_summary": {
      "started_at": "2025-01-11T10:31:30.789Z",
      "total_duration_ms": 240000,
      "processing_time_ms": 238500,
      "system_overhead_ms": 1500,
      "retries_used": 0
    },
    "process_continuation": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "process-payment",
      "token_id": "srv1-token-payment01",
      "next_activities": [
        {
          "element_id": "send-confirmation",
          "element_type": "serviceTask",
          "element_name": "Send Confirmation Email"
        }
      ]
    },
    "variables_updated": {
      "added": ["paymentStatus", "transactionId", "receipt"],
      "modified": ["amount"],
      "removed": []
    },
    "result_variables": {
      "paymentStatus": "completed",
      "transactionId": "txn_abc123def456",
      "amount": 299.99,
      "processingTime": 2500,
      "provider": "stripe",
      "fees": 8.97,
      "netAmount": 291.02
    }
  },
  "request_id": "req_1641998402600"
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
      "job_key": "srv1-job-nonexistent",
      "possible_reasons": [
        "Job key does not exist",
        "Job already completed or cancelled", 
        "Job belongs to different worker",
        "Job expired"
      ]
    }
  },
  "request_id": "req_1641998402601"
}
```

### 409 Conflict - Задание нельзя завершить
```json
{
  "success": false,
  "error": {
    "code": "JOB_NOT_COMPLETABLE",
    "message": "Job cannot be completed in current state",
    "details": {
      "job_key": "srv1-job-xyz789",
      "current_status": "CANCELLED",
      "worker_id": "payment-worker-02",
      "reason": "Job was cancelled before completion"
    }
  },
  "request_id": "req_1641998402602"
}
```

### 400 Bad Request - Неверные переменные
```json
{
  "success": false,
  "error": {
    "code": "INVALID_VARIABLES",
    "message": "Invalid result variables provided",
    "details": {
      "validation_errors": [
        {
          "field": "amount",
          "error": "Must be a positive number",
          "provided_value": -100
        },
        {
          "field": "transactionId",
          "error": "Must be a non-empty string",
          "provided_value": ""
        }
      ]
    }
  },
  "request_id": "req_1641998402603"
}
```

## Поля ответа

### Job Completion Information
- `job_key` (string): Ключ завершенного задания
- `status` (string): Новый статус (`COMPLETED`)
- `completed_at` (string): Время завершения
- `completed_by` (string): ID worker'а

### Execution Summary
- `started_at` (string): Время начала выполнения
- `total_duration_ms` (integer): Общее время выполнения
- `processing_time_ms` (integer): Время обработки
- `system_overhead_ms` (integer): Системные накладные расходы
- `retries_used` (integer): Количество использованных повторов

### Process Continuation
- `process_instance_id` (string): ID экземпляра процесса
- `element_id` (string): ID элемента, который выполнял задание
- `token_id` (string): ID токена
- `next_activities` (array): Следующие активности процесса

### Variables Information
- `variables_updated` (object): Информация об изменениях переменных
- `result_variables` (object): Переменные результата

## Использование

### Worker Implementation
```javascript
class PaymentWorker {
  async processJob(job) {
    try {
      // Обработка платежа
      const paymentResult = await this.processPayment(job.variables);
      
      // Завершение задания с результатом
      const completion = await fetch(`/api/v1/jobs/${job.key}/complete`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'payment-worker-key'
        },
        body: JSON.stringify({
          variables: {
            paymentStatus: paymentResult.status,
            transactionId: paymentResult.transactionId,
            amount: paymentResult.amount,
            fees: paymentResult.fees,
            receipt: paymentResult.receipt
          }
        })
      });
      
      const result = await completion.json();
      
      if (result.success) {
        console.log(`Job ${job.key} completed successfully`);
        return result.data;
      } else {
        throw new Error(`Failed to complete job: ${result.error.message}`);
      }
      
    } catch (error) {
      console.error(`Error processing job ${job.key}:`, error);
      throw error;
    }
  }
  
  async processPayment(variables) {
    // Симуляция обработки платежа
    const { amount, paymentMethod, orderId } = variables;
    
    // Валидация
    if (amount <= 0) {
      throw new Error('Invalid amount');
    }
    
    // Обработка платежа через внешний сервис
    const paymentResponse = await this.callPaymentService({
      amount,
      paymentMethod,
      orderId
    });
    
    return {
      status: 'completed',
      transactionId: paymentResponse.transactionId,
      amount: paymentResponse.amount,
      fees: paymentResponse.fees,
      receipt: {
        url: paymentResponse.receiptUrl,
        id: paymentResponse.receiptId
      }
    };
  }
  
  async callPaymentService(paymentData) {
    // Интеграция с платежным сервисом
    const response = await fetch('https://payments.example.com/api/process', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(paymentData)
    });
    
    return await response.json();
  }
}
```

### Generic Worker Pattern
```javascript
class GenericWorker {
  constructor(workerType, apiKey) {
    this.workerType = workerType;
    this.apiKey = apiKey;
    this.running = false;
  }
  
  async start() {
    this.running = true;
    
    while (this.running) {
      try {
        // Активация заданий
        const jobs = await this.activateJobs();
        
        if (jobs.length > 0) {
          // Обработка заданий параллельно
          await Promise.all(jobs.map(job => this.processJob(job)));
        } else {
          // Небольшая пауза если нет заданий
          await this.sleep(1000);
        }
        
      } catch (error) {
        console.error('Worker error:', error);
        await this.sleep(5000); // Пауза при ошибке
      }
    }
  }
  
  async activateJobs() {
    const response = await fetch('/api/v1/jobs/activate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        type: this.workerType,
        worker: `${this.workerType}-worker-${process.pid}`,
        max_jobs: 5,
        timeout: 30000
      })
    });
    
    const result = await response.json();
    return result.success ? result.data.jobs : [];
  }
  
  async processJob(job) {
    try {
      console.log(`Processing job ${job.key} of type ${job.type}`);
      
      // Пользовательская логика обработки
      const result = await this.executeJobLogic(job);
      
      // Завершение задания
      return await this.completeJob(job.key, result);
      
    } catch (error) {
      console.error(`Error processing job ${job.key}:`, error);
      
      // Провал задания
      return await this.failJob(job.key, error.message);
    }
  }
  
  async completeJob(jobKey, variables = {}) {
    const response = await fetch(`/api/v1/jobs/${jobKey}/complete`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({ variables })
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log(`Job ${jobKey} completed successfully`);
    } else {
      console.error(`Failed to complete job ${jobKey}:`, result.error.message);
    }
    
    return result;
  }
  
  async failJob(jobKey, errorMessage) {
    const response = await fetch(`/api/v1/jobs/${jobKey}/fail`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        retries: 2,
        error_message: errorMessage
      })
    });
    
    return await response.json();
  }
  
  // Переопределяется в наследниках
  async executeJobLogic(job) {
    throw new Error('executeJobLogic must be overridden by subclass');
  }
  
  stop() {
    this.running = false;
  }
  
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Использование
class EmailWorker extends GenericWorker {
  constructor(apiKey) {
    super('email-service', apiKey);
  }
  
  async executeJobLogic(job) {
    const { recipient, subject, body } = job.variables;
    
    // Отправка email
    const emailResult = await this.sendEmail(recipient, subject, body);
    
    return {
      emailSent: true,
      messageId: emailResult.messageId,
      sentAt: new Date().toISOString()
    };
  }
  
  async sendEmail(recipient, subject, body) {
    // Интеграция с email сервисом
    console.log(`Sending email to ${recipient}: ${subject}`);
    
    return {
      messageId: 'msg_' + Math.random().toString(36).substr(2, 9)
    };
  }
}
```

### Job Completion with Validation
```javascript
async function completeJobWithValidation(jobKey, resultVariables) {
  // Валидация результата
  const validation = validateJobResult(resultVariables);
  if (!validation.isValid) {
    throw new Error(`Invalid job result: ${validation.errors.join(', ')}`);
  }
  
  // Завершение задания
  const response = await fetch(`/api/v1/jobs/${jobKey}/complete`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your-api-key'
    },
    body: JSON.stringify({
      variables: resultVariables
    })
  });
  
  const result = await response.json();
  
  if (!result.success) {
    throw new Error(`Job completion failed: ${result.error.message}`);
  }
  
  return result.data;
}

function validateJobResult(variables) {
  const errors = [];
  
  // Проверка обязательных полей
  if (variables.status && !['completed', 'failed', 'cancelled'].includes(variables.status)) {
    errors.push('Invalid status value');
  }
  
  // Проверка типов данных
  if (variables.amount && (typeof variables.amount !== 'number' || variables.amount < 0)) {
    errors.push('Amount must be a positive number');
  }
  
  // Проверка формата ID
  if (variables.transactionId && !/^txn_[a-zA-Z0-9]+$/.test(variables.transactionId)) {
    errors.push('Invalid transaction ID format');
  }
  
  return {
    isValid: errors.length === 0,
    errors
  };
}
```

## Связанные endpoints
- [`POST /api/v1/jobs/activate`](./activate-jobs.md) - Активация заданий
- [`PUT /api/v1/jobs/:key/fail`](./fail-job.md) - Провалить задание
- [`GET /api/v1/jobs/:key`](./get-job.md) - Детали задания
- [`POST /api/v1/jobs/:key/throw-error`](./throw-error.md) - Выбросить BPMN ошибку
