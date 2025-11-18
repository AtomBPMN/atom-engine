# POST /api/v1/messages/publish

## Описание
Публикация сообщения в системе для корреляции с BPMN процессами. Сообщения могут активировать message start events или intermediate catch events.

## URL
```
POST /api/v1/messages/publish
```

## Авторизация
✅ **Требуется API ключ** с разрешением `message`

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `name` (string): Имя сообщения (должно соответствовать определению в BPMN)

### Опциональные поля
- `correlation_key` (string): Ключ корреляции для связи с экземпляром процесса
- `variables` (object): Переменные сообщения
- `ttl` (string): Время жизни сообщения в формате ISO 8601 (по умолчанию: "PT24H")
- `tenant_id` (string): ID тенанта (по умолчанию: "default")

## Примеры запросов

### Простое сообщение
```bash
curl -X POST "http://localhost:27555/api/v1/messages/publish" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "name": "order_created"
  }'
```

### Сообщение с корреляцией
```bash
curl -X POST "http://localhost:27555/api/v1/messages/publish" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "name": "payment_completed",
    "correlation_key": "order-123",
    "variables": {
      "paymentId": "pay_abc123",
      "amount": 299.99,
      "status": "completed"
    }
  }'
```

### Сообщение с TTL
```bash
curl -X POST "http://localhost:27555/api/v1/messages/publish" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "name": "urgent_notification",
    "correlation_key": "user-456",
    "variables": {
      "priority": "high",
      "message": "Action required"
    },
    "ttl": "PT1H"
  }'
```

### JavaScript
```javascript
const message = {
  name: 'inventory_updated',
  correlation_key: 'product-789',
  variables: {
    productId: 'PROD-789',
    newQuantity: 150,
    warehouseId: 'WH-01',
    updatedBy: 'inventory-system'
  },
  ttl: 'PT6H'
};

const response = await fetch('/api/v1/messages/publish', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(message)
});

const result = await response.json();
```

## Ответы

### 200 OK - Сообщение опубликовано и коррелировано
```json
{
  "success": true,
  "data": {
    "message_id": "srv1-msg-abc123def456",
    "name": "payment_completed",
    "correlation_key": "order-123",
    "published_at": "2025-01-11T10:30:00.000Z",
    "ttl": "PT24H",
    "expires_at": "2025-01-12T10:30:00.000Z",
    "tenant_id": "production",
    "correlation_result": {
      "correlated": true,
      "subscriptions_matched": 2,
      "processes_triggered": [
        {
          "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
          "element_id": "wait-payment-confirmation",
          "element_type": "intermediateCatchEvent",
          "continued_at": "2025-01-11T10:30:00.123Z"
        }
      ],
      "new_processes_started": [
        {
          "process_instance_id": "srv1-cD4eF8gH1jK3mN6pQ9",
          "process_id": "payment-reconciliation",
          "started_at": "2025-01-11T10:30:00.456Z"
        }
      ]
    },
    "variables_propagated": {
      "paymentId": "pay_abc123",
      "amount": 299.99,
      "status": "completed"
    }
  },
  "request_id": "req_1641998403400"
}
```

### 200 OK - Сообщение опубликовано, но не коррелировано
```json
{
  "success": true,
  "data": {
    "message_id": "srv1-msg-def456ghi789",
    "name": "customer_updated",
    "correlation_key": "customer-999",
    "published_at": "2025-01-11T10:30:00.000Z",
    "ttl": "PT24H",
    "expires_at": "2025-01-12T10:30:00.000Z",
    "tenant_id": "default",
    "correlation_result": {
      "correlated": false,
      "reason": "NO_MATCHING_SUBSCRIPTIONS",
      "subscriptions_checked": 0,
      "buffered": true,
      "buffer_until": "2025-01-12T10:30:00.000Z"
    },
    "variables": {
      "customerId": "customer-999",
      "field": "email",
      "newValue": "newemail@example.com"
    }
  },
  "request_id": "req_1641998403401"
}
```

### 400 Bad Request - Неверное имя сообщения
```json
{
  "success": false,
  "error": {
    "code": "INVALID_MESSAGE_NAME",
    "message": "Invalid message name format",
    "details": {
      "provided_name": "",
      "requirements": [
        "Message name must not be empty",
        "Name should be alphanumeric with underscores",
        "Must match BPMN message definition"
      ]
    }
  },
  "request_id": "req_1641998403402"
}
```

## Поля ответа

### Message Information
- `message_id` (string): Уникальный ID сообщения
- `name` (string): Имя сообщения
- `correlation_key` (string): Ключ корреляции
- `published_at` (string): Время публикации
- `ttl` (string): Время жизни
- `expires_at` (string): Время истечения

### Correlation Result
- `correlated` (boolean): Было ли сообщение коррелировано
- `subscriptions_matched` (integer): Количество совпавших подписок
- `processes_triggered` (array): Продолженные процессы
- `new_processes_started` (array): Новые запущенные процессы

### Buffering Information (если не коррелировано)
- `buffered` (boolean): Буферизовано ли сообщение
- `buffer_until` (string): До какого времени буферизовано

## Message Correlation

### Correlation Rules
1. **По имени сообщения**: Должно совпадать с BPMN определением
2. **По ключу корреляции**: Должен совпадать с выражением в процессе
3. **По тенанту**: Сообщения коррелируются только в рамках тенанта

### BPMN Message Definitions
```xml
<!-- В BPMN файле -->
<bpmn:message id="PaymentCompleted" name="payment_completed" />

<bpmn:intermediateCatchEvent id="wait-payment">
  <bpmn:messageEventDefinition messageRef="PaymentCompleted" />
  <bpmn:extensionElements>
    <zeebe:subscription correlationKey="orderId" />
  </bpmn:extensionElements>
</bpmn:intermediateCatchEvent>
```

## Использование

### Order Processing Messages
```javascript
class OrderMessagePublisher {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async publishOrderCreated(orderData) {
    const message = {
      name: 'order_created',
      correlation_key: orderData.orderId,
      variables: {
        orderId: orderData.orderId,
        customerId: orderData.customerId,
        amount: orderData.amount,
        items: orderData.items,
        createdAt: new Date().toISOString()
      },
      ttl: 'PT24H'
    };
    
    return await this.publishMessage(message);
  }
  
  async publishPaymentCompleted(paymentData) {
    const message = {
      name: 'payment_completed',
      correlation_key: paymentData.orderId,
      variables: {
        paymentId: paymentData.paymentId,
        orderId: paymentData.orderId,
        amount: paymentData.amount,
        method: paymentData.method,
        transactionId: paymentData.transactionId,
        completedAt: new Date().toISOString()
      }
    };
    
    return await this.publishMessage(message);
  }
  
  async publishShipmentDispatched(shipmentData) {
    const message = {
      name: 'shipment_dispatched',
      correlation_key: shipmentData.orderId,
      variables: {
        trackingNumber: shipmentData.trackingNumber,
        orderId: shipmentData.orderId,
        carrier: shipmentData.carrier,
        estimatedDelivery: shipmentData.estimatedDelivery,
        dispatchedAt: new Date().toISOString()
      }
    };
    
    return await this.publishMessage(message);
  }
  
  async publishMessage(message) {
    const response = await fetch('/api/v1/messages/publish', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(message)
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log(`Message ${message.name} published:`, result.data.message_id);
      
      if (result.data.correlation_result.correlated) {
        console.log(`Correlated with ${result.data.correlation_result.processes_triggered.length} processes`);
      } else {
        console.log('Message buffered for future correlation');
      }
    }
    
    return result;
  }
}

// Использование
const publisher = new OrderMessagePublisher('your-api-key');

// Публикация сообщения о создании заказа
await publisher.publishOrderCreated({
  orderId: 'ORD-12345',
  customerId: 'CUST-67890',
  amount: 299.99,
  items: [{ sku: 'PROD-001', quantity: 2 }]
});

// Публикация сообщения о завершении платежа
await publisher.publishPaymentCompleted({
  paymentId: 'pay_abc123',
  orderId: 'ORD-12345',
  amount: 299.99,
  method: 'credit_card',
  transactionId: 'txn_def456'
});
```

### Event-Driven Architecture
```javascript
class EventPublisher {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async publishDomainEvent(eventType, aggregateId, eventData, ttl = 'PT24H') {
    const message = {
      name: eventType,
      correlation_key: aggregateId,
      variables: {
        eventType,
        aggregateId,
        aggregateVersion: eventData.version || 1,
        eventData,
        eventId: this.generateEventId(),
        occurredAt: new Date().toISOString()
      },
      ttl
    };
    
    const result = await this.publishMessage(message);
    
    // Логирование для event sourcing
    this.logEvent(eventType, aggregateId, result);
    
    return result;
  }
  
  async publishCustomerEvents(customerId, eventType, data) {
    const eventMap = {
      'customer_registered': {
        name: 'customer_registered',
        ttl: 'P7D' // 7 дней для важных событий
      },
      'customer_updated': {
        name: 'customer_updated',
        ttl: 'PT12H'
      },
      'customer_deleted': {
        name: 'customer_deleted',
        ttl: 'P30D' // 30 дней для audit
      }
    };
    
    const eventConfig = eventMap[eventType];
    if (!eventConfig) {
      throw new Error(`Unknown customer event type: ${eventType}`);
    }
    
    return await this.publishDomainEvent(
      eventConfig.name,
      customerId,
      data,
      eventConfig.ttl
    );
  }
  
  generateEventId() {
    return 'evt_' + Math.random().toString(36).substr(2, 9);
  }
  
  logEvent(eventType, aggregateId, result) {
    const logEntry = {
      timestamp: new Date().toISOString(),
      eventType,
      aggregateId,
      messageId: result.data?.message_id,
      correlated: result.data?.correlation_result?.correlated,
      processesAffected: result.data?.correlation_result?.processes_triggered?.length || 0
    };
    
    console.log('Event published:', logEntry);
  }
}
```

### Message Broadcasting
```javascript
class NotificationBroadcaster {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async broadcastSystemNotification(notificationType, data) {
    // Системные уведомления отправляются без корреляции
    const message = {
      name: 'system_notification',
      variables: {
        notificationType,
        severity: data.severity || 'INFO',
        title: data.title,
        message: data.message,
        timestamp: new Date().toISOString(),
        source: 'notification-service'
      },
      ttl: 'PT2H' // Краткое время жизни для уведомлений
    };
    
    return await this.publishMessage(message);
  }
  
  async broadcastToUser(userId, notificationType, data) {
    // Уведомления для конкретного пользователя
    const message = {
      name: 'user_notification',
      correlation_key: userId,
      variables: {
        userId,
        notificationType,
        title: data.title,
        message: data.message,
        actionUrl: data.actionUrl,
        timestamp: new Date().toISOString()
      },
      ttl: 'P1D'
    };
    
    return await this.publishMessage(message);
  }
  
  async publishMessage(message) {
    const response = await fetch('/api/v1/messages/publish', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(message)
    });
    
    return await response.json();
  }
}

// Использование
const broadcaster = new NotificationBroadcaster('your-api-key');

// Системное уведомление
await broadcaster.broadcastSystemNotification('maintenance_scheduled', {
  severity: 'WARNING',
  title: 'Scheduled Maintenance',
  message: 'System will be unavailable on Sunday 2-4 AM'
});

// Персональное уведомление
await broadcaster.broadcastToUser('user-123', 'order_shipped', {
  title: 'Order Shipped',
  message: 'Your order #ORD-12345 has been shipped',
  actionUrl: '/orders/ORD-12345/tracking'
});
```

### Message Reliability
```javascript
class ReliableMessagePublisher {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.retryAttempts = 3;
    this.retryDelay = 1000; // 1 second
  }
  
  async publishWithRetry(message, maxRetries = this.retryAttempts) {
    let lastError;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const result = await this.publishMessage(message);
        
        if (result.success) {
          if (attempt > 1) {
            console.log(`Message published successfully on attempt ${attempt}`);
          }
          return result;
        } else {
          lastError = new Error(result.error.message);
        }
        
      } catch (error) {
        lastError = error;
        console.warn(`Publish attempt ${attempt} failed:`, error.message);
        
        if (attempt < maxRetries) {
          await this.sleep(this.retryDelay * attempt); // Exponential backoff
        }
      }
    }
    
    throw new Error(`Failed to publish message after ${maxRetries} attempts: ${lastError.message}`);
  }
  
  async publishMessage(message) {
    const response = await fetch('/api/v1/messages/publish', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(message)
    });
    
    return await response.json();
  }
  
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## TTL (Time To Live)

### Стандартные TTL значения
- **Системные события**: PT2H (2 часа)
- **Бизнес события**: PT24H (24 часа) 
- **Критические события**: P7D (7 дней)
- **Audit события**: P30D (30 дней)

### TTL формат (ISO 8601)
```yaml
Seconds:  PT30S   # 30 секунд
Minutes:  PT5M    # 5 минут  
Hours:    PT2H    # 2 часа
Days:     P1D     # 1 день
Weeks:    P1W     # 1 неделя
Months:   P1M     # 1 месяц
Combined: P1DT2H  # 1 день 2 часа
```

## Связанные endpoints
- [`GET /api/v1/messages`](./list-messages.md) - Список результатов корреляции
- [`GET /api/v1/messages/subscriptions`](./list-subscriptions.md) - Активные подписки
- [`GET /api/v1/messages/buffered`](./list-buffered.md) - Буферизованные сообщения
- [`GET /api/v1/messages/stats`](./get-message-stats.md) - Статистика сообщений
