# POST /api/v1/timers

## Описание
Создание нового таймера в системе с поддержкой однократных и циклических таймеров в формате ISO 8601.

## URL
```
POST /api/v1/timers
```

## Авторизация
✅ **Требуется API ключ** с разрешением `timer`

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `timer_id` (string): Уникальный ID таймера
- `duration_or_cycle` (string): Длительность или цикл в формате ISO 8601

### Опциональные поля
- `metadata` (object): Дополнительные метаданные таймера
- `callback_url` (string): URL для callback уведомлений
- `tenant_id` (string): ID тенанта (по умолчанию: "default")

### Пример тела запроса
```json
{
  "timer_id": "payment-timeout-ORD-12345",
  "duration_or_cycle": "PT5M",
  "metadata": {
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "element_id": "payment-timeout-boundary",
    "order_id": "ORD-12345",
    "timeout_reason": "payment_processing"
  },
  "callback_url": "http://localhost:27555/internal/timer/callback",
  "tenant_id": "production"
}
```

## Примеры запросов

### Простой таймер на 5 минут
```bash
curl -X POST "http://localhost:27555/api/v1/timers" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "timer_id": "simple-timer-001",
    "duration_or_cycle": "PT5M"
  }'
```

### Циклический таймер
```bash
curl -X POST "http://localhost:27555/api/v1/timers" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "timer_id": "hourly-cleanup",
    "duration_or_cycle": "R/PT1H",
    "metadata": {
      "task_type": "cleanup",
      "priority": "low"
    }
  }'
```

### Таймер с ограниченными повторами
```bash
curl -X POST "http://localhost:27555/api/v1/timers" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "timer_id": "reminder-timer",
    "duration_or_cycle": "R3/PT10M",
    "metadata": {
      "reminder_type": "payment_due",
      "user_id": "USER-67890"
    },
    "callback_url": "https://notifications.example.com/webhook"
  }'
```

### JavaScript
```javascript
const timer = {
  timer_id: 'session-timeout-' + Date.now(),
  duration_or_cycle: 'PT30M',
  metadata: {
    session_id: 'sess_abc123',
    user_id: 'user_456'
  }
};

const response = await fetch('/api/v1/timers', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(timer)
});

const result = await response.json();
```

## Ответы

### 201 Created - Таймер создан
```json
{
  "success": true,
  "data": {
    "timer_id": "payment-timeout-ORD-12345",
    "duration_or_cycle": "PT5M",
    "status": "SCHEDULED",
    "created_at": "2025-01-11T10:30:00.000Z",
    "scheduled_at": "2025-01-11T10:35:00.000Z",
    "tenant_id": "production",
    "timer_type": "DURATION",
    "cycle_info": null,
    "remaining_time_ms": 300000,
    "metadata": {
      "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "element_id": "payment-timeout-boundary",
      "order_id": "ORD-12345",
      "timeout_reason": "payment_processing"
    },
    "callback_url": "http://localhost:27555/internal/timer/callback",
    "timewheel_level": 2,
    "slot_position": 156
  },
  "request_id": "req_1641998402500"
}
```

### 201 Created - Циклический таймер создан
```json
{
  "success": true,
  "data": {
    "timer_id": "hourly-cleanup",
    "duration_or_cycle": "R/PT1H",
    "status": "SCHEDULED",
    "created_at": "2025-01-11T10:30:00.000Z",
    "scheduled_at": "2025-01-11T11:30:00.000Z",
    "tenant_id": "default",
    "timer_type": "CYCLE",
    "cycle_info": {
      "repetitions": "infinite",
      "interval_duration": "PT1H",
      "current_iteration": 1,
      "next_iterations": [
        "2025-01-11T11:30:00.000Z",
        "2025-01-11T12:30:00.000Z",
        "2025-01-11T13:30:00.000Z"
      ]
    },
    "remaining_time_ms": 3600000,
    "metadata": {
      "task_type": "cleanup",
      "priority": "low"
    },
    "callback_url": null,
    "timewheel_level": 3,
    "slot_position": 89
  },
  "request_id": "req_1641998402501"
}
```

### 400 Bad Request - Неверный формат
```json
{
  "success": false,
  "error": {
    "code": "INVALID_DURATION_FORMAT",
    "message": "Invalid ISO 8601 duration format",
    "details": {
      "provided": "5M",
      "expected_format": "PT5M",
      "valid_examples": [
        "PT30S - 30 seconds",
        "PT5M - 5 minutes", 
        "PT1H - 1 hour",
        "P1D - 1 day",
        "R5/PT30S - 5 repetitions every 30 seconds",
        "R/PT1H - infinite repetitions every hour"
      ]
    }
  },
  "request_id": "req_1641998402502"
}
```

### 409 Conflict - Таймер уже существует
```json
{
  "success": false,
  "error": {
    "code": "TIMER_ALREADY_EXISTS",
    "message": "Timer with this ID already exists",
    "details": {
      "timer_id": "payment-timeout-ORD-12345",
      "existing_status": "SCHEDULED",
      "created_at": "2025-01-11T10:25:00.000Z",
      "suggestion": "Use different timer_id or cancel existing timer first"
    }
  },
  "request_id": "req_1641998402503"
}
```

## Поля ответа

### Basic Information
- `timer_id` (string): ID таймера
- `duration_or_cycle` (string): Исходный формат длительности
- `status` (string): Статус таймера
- `created_at` (string): Время создания
- `scheduled_at` (string): Время срабатывания

### Timer Configuration
- `timer_type` (string): Тип (`DURATION`, `CYCLE`)
- `tenant_id` (string): ID тенанта
- `metadata` (object): Метаданные
- `callback_url` (string): URL для callback

### Cycle Information (для циклических таймеров)
- `repetitions` (string): Количество повторов
- `interval_duration` (string): Интервал между повторами
- `current_iteration` (integer): Текущая итерация
- `next_iterations` (array): Следующие срабатывания

### System Information
- `remaining_time_ms` (integer): Оставшееся время в миллисекундах
- `timewheel_level` (integer): Уровень в иерархии timewheel
- `slot_position` (integer): Позиция в слоте timewheel

## Форматы ISO 8601

### Duration Format (Длительность)
```yaml
Seconds:
  - PT10S    # 10 секунд
  - PT30S    # 30 секунд

Minutes:
  - PT1M     # 1 минута
  - PT5M     # 5 минут
  - PT30M    # 30 минут

Hours:
  - PT1H     # 1 час
  - PT2H30M  # 2 часа 30 минут

Days:
  - P1D      # 1 день
  - P7D      # 7 дней
  - P1DT2H   # 1 день 2 часа

Combined:
  - P1DT2H30M45S  # 1 день 2 часа 30 минут 45 секунд
```

### Cycle Format (Циклы)
```yaml
Infinite Repetition:
  - R/PT30S    # Каждые 30 секунд бесконечно
  - R/PT5M     # Каждые 5 минут бесконечно
  - R/PT1H     # Каждый час бесконечно

Limited Repetition:
  - R3/PT10S   # 3 раза каждые 10 секунд
  - R5/PT1M    # 5 раз каждую минуту
  - R10/PT1H   # 10 раз каждый час

Complex Cycles:
  - R/P1D      # Каждый день бесконечно
  - R12/PT2H   # 12 раз каждые 2 часа
```

## Типы таймеров

### DURATION (Однократные)
- Срабатывают один раз через указанное время
- Автоматически удаляются после срабатывания
- Подходят для timeouts, deadlines

### CYCLE (Циклические)
- Срабатывают периодически
- Могут быть ограничены количеством повторов
- Подходят для периодических задач

## Использование

### Process Timeout Timer
```javascript
async function createProcessTimeout(processInstanceId, timeoutMinutes) {
  const timer = {
    timer_id: `process-timeout-${processInstanceId}`,
    duration_or_cycle: `PT${timeoutMinutes}M`,
    metadata: {
      process_instance_id: processInstanceId,
      timeout_type: 'process_execution',
      action: 'cancel_process'
    },
    callback_url: '/internal/process/timeout-handler'
  };
  
  const response = await fetch('/api/v1/timers', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your-api-key'
    },
    body: JSON.stringify(timer)
  });
  
  return await response.json();
}
```

### Reminder System
```javascript
class ReminderService {
  async scheduleReminder(userId, message, delayMinutes, maxReminders = 3) {
    const timer = {
      timer_id: `reminder-${userId}-${Date.now()}`,
      duration_or_cycle: `R${maxReminders}/PT${delayMinutes}M`,
      metadata: {
        user_id: userId,
        message: message,
        reminder_type: 'user_notification'
      },
      callback_url: '/api/notifications/reminder-callback'
    };
    
    return await this.createTimer(timer);
  }
  
  async schedulePaymentReminder(orderId, amount) {
    // Напоминания через 1 день, 3 дня, и 7 дней
    const intervals = ['P1D', 'P3D', 'P7D'];
    const timers = [];
    
    for (let i = 0; i < intervals.length; i++) {
      const timer = {
        timer_id: `payment-reminder-${orderId}-${i + 1}`,
        duration_or_cycle: intervals[i],
        metadata: {
          order_id: orderId,
          amount: amount,
          reminder_sequence: i + 1,
          total_reminders: intervals.length
        }
      };
      
      timers.push(await this.createTimer(timer));
    }
    
    return timers;
  }
  
  async createTimer(timer) {
    const response = await fetch('/api/v1/timers', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
      },
      body: JSON.stringify(timer)
    });
    
    return await response.json();
  }
}
```

### Cleanup Scheduler
```javascript
async function scheduleCleanupTasks() {
  const cleanupTasks = [
    {
      timer_id: 'cleanup-temp-files',
      duration_or_cycle: 'R/PT6H',  // Каждые 6 часов
      metadata: {
        task: 'cleanup_temp_files',
        priority: 'low'
      }
    },
    {
      timer_id: 'cleanup-old-logs',
      duration_or_cycle: 'R/P1D',   // Каждый день
      metadata: {
        task: 'cleanup_old_logs',
        retention_days: 30,
        priority: 'medium'
      }
    },
    {
      timer_id: 'backup-database',
      duration_or_cycle: 'R/PT12H', // Каждые 12 часов
      metadata: {
        task: 'backup_database',
        priority: 'high'
      }
    }
  ];
  
  const results = [];
  for (const task of cleanupTasks) {
    const response = await fetch('/api/v1/timers', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'your-api-key'
      },
      body: JSON.stringify(task)
    });
    
    results.push(await response.json());
  }
  
  return results;
}
```

### Session Management
```javascript
class SessionManager {
  async createSessionTimer(sessionId, timeoutMinutes = 30) {
    const timer = {
      timer_id: `session-${sessionId}`,
      duration_or_cycle: `PT${timeoutMinutes}M`,
      metadata: {
        session_id: sessionId,
        timeout_action: 'destroy_session',
        created_at: new Date().toISOString()
      },
      callback_url: '/internal/session/timeout'
    };
    
    return await fetch('/api/v1/timers', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'session-manager-key'
      },
      body: JSON.stringify(timer)
    });
  }
  
  async extendSession(sessionId, additionalMinutes = 30) {
    // Сначала отменяем старый таймер
    await fetch(`/api/v1/timers/${sessionId}`, {
      method: 'DELETE',
      headers: { 'X-API-Key': 'session-manager-key' }
    });
    
    // Создаем новый с обновленным временем
    return await this.createSessionTimer(sessionId, additionalMinutes);
  }
}
```

## Валидация

### Timer ID
- Формат: буквы, цифры, дефисы, подчеркивания
- Длина: 1-100 символов
- Уникальность в рамках тенанта

### Duration Format
- Должен соответствовать ISO 8601
- Минимальная длительность: 1 секунда
- Максимальная длительность: 100 лет

### Cycle Format
- R[count]/duration или R/duration
- Count: 1-1000 (для ограниченных циклов)
- Duration: согласно правилам выше

## Связанные endpoints
- [`GET /api/v1/timers`](./list-timers.md) - Список таймеров
- [`GET /api/v1/timers/:id`](./get-timer.md) - Статус таймера
- [`DELETE /api/v1/timers/:id`](./delete-timer.md) - Удалить таймер
- [`GET /api/v1/timers/stats`](./get-timer-stats.md) - Статистика таймеров
