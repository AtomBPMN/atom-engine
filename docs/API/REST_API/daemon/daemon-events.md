# GET /api/v1/daemon/events

## Описание
Получение списка системных событий демона: запуск, остановка, ошибки компонентов и другие важные события.

## URL
```
GET /api/v1/daemon/events
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Параметры запроса (Query Parameters)

### Фильтрация
- `level` (string): Уровень событий (`INFO`, `WARN`, `ERROR`, `FATAL`)
- `component` (string): Фильтр по компоненту
- `since` (string): События после даты (ISO 8601)
- `until` (string): События до даты (ISO 8601)

### Пагинация
- `page` (integer): Номер страницы (по умолчанию: 1)
- `page_size` (integer): Размер страницы (по умолчанию: 50, максимум: 200)

## Примеры запросов

### Последние события
```bash
curl -X GET "http://localhost:27555/api/v1/daemon/events" \
  -H "X-API-Key: your-api-key-here"
```

### События ошибок
```bash
curl -X GET "http://localhost:27555/api/v1/daemon/events?level=ERROR" \
  -H "X-API-Key: your-api-key-here"
```

### События за последний час
```bash
curl -X GET "http://localhost:27555/api/v1/daemon/events?since=2025-01-11T09:30:00Z" \
  -H "X-API-Key: your-api-key-here"
```

## Ответы

### 200 OK - События получены
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "evt_1641998401100",
        "timestamp": "2025-01-11T10:30:00.123Z",
        "level": "INFO",
        "component": "daemon",
        "event_type": "DAEMON_STARTED",
        "message": "Atom Engine daemon started successfully",
        "details": {
          "pid": 12345,
          "version": "1.0.0",
          "startup_time_ms": 15420
        }
      },
      {
        "id": "evt_1641998401101", 
        "timestamp": "2025-01-11T10:29:45.456Z",
        "level": "INFO",
        "component": "process_engine",
        "event_type": "COMPONENT_INITIALIZED",
        "message": "Process engine initialized",
        "details": {
          "max_processes": 10000,
          "initialization_time_ms": 2340
        }
      },
      {
        "id": "evt_1641998401102",
        "timestamp": "2025-01-11T10:25:30.789Z", 
        "level": "WARN",
        "component": "storage",
        "event_type": "CONNECTION_RECOVERED",
        "message": "Database connection recovered after temporary failure",
        "details": {
          "downtime_seconds": 45,
          "error": "connection timeout",
          "recovery_attempts": 3
        }
      },
      {
        "id": "evt_1641998401103",
        "timestamp": "2025-01-11T10:20:15.012Z",
        "level": "ERROR", 
        "component": "job_manager",
        "event_type": "WORKER_DISCONNECTED",
        "message": "Worker unexpectedly disconnected",
        "details": {
          "worker_id": "email-worker-03",
          "active_jobs": 5,
          "last_heartbeat": "2025-01-11T10:19:45.000Z"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 50,
      "total_count": 1247,
      "total_pages": 25,
      "has_next": true,
      "has_prev": false
    },
    "summary": {
      "info_count": 845,
      "warn_count": 302,
      "error_count": 95,
      "fatal_count": 5
    }
  },
  "request_id": "req_1641998401100"
}
```

## Поля события

### Основная информация
- `id` (string): Уникальный ID события
- `timestamp` (string): Время события (ISO 8601 UTC)
- `level` (string): Уровень события
- `component` (string): Компонент источник
- `event_type` (string): Тип события
- `message` (string): Человекочитаемое описание

### Детали события
- `details` (object): Дополнительная информация
- Специфичные поля в зависимости от типа события

## Типы событий

### Daemon Events
- `DAEMON_STARTED` - Демон запущен
- `DAEMON_STOPPING` - Демон останавливается
- `DAEMON_STOPPED` - Демон остановлен
- `CONFIG_RELOADED` - Конфигурация перезагружена

### Component Events
- `COMPONENT_INITIALIZED` - Компонент инициализирован
- `COMPONENT_STARTED` - Компонент запущен
- `COMPONENT_STOPPED` - Компонент остановлен
- `COMPONENT_FAILED` - Компонент сбойнул
- `COMPONENT_RECOVERED` - Компонент восстановлен

### Storage Events
- `DATABASE_CONNECTED` - Подключение к БД
- `DATABASE_DISCONNECTED` - Отключение от БД
- `CONNECTION_RECOVERED` - Восстановление соединения
- `BACKUP_COMPLETED` - Резервное копирование завершено

### Performance Events
- `HIGH_MEMORY_USAGE` - Высокое использование памяти
- `HIGH_CPU_USAGE` - Высокая загрузка CPU
- `SLOW_QUERY_DETECTED` - Обнаружен медленный запрос

## Использование

### Мониторинг в реальном времени
```javascript
async function monitorEvents() {
  let lastEventId = null;
  
  while (true) {
    const params = new URLSearchParams({
      level: 'ERROR,WARN',
      page_size: '10'
    });
    
    if (lastEventId) {
      params.append('since_id', lastEventId);
    }
    
    const response = await fetch(`/api/v1/daemon/events?${params}`, {
      headers: { 'X-API-Key': 'your-api-key' }
    });
    
    const data = await response.json();
    
    for (const event of data.data.events) {
      console.log(`[${event.level}] ${event.component}: ${event.message}`);
      lastEventId = event.id;
      
      if (event.level === 'ERROR' || event.level === 'FATAL') {
        await sendAlert(event);
      }
    }
    
    await new Promise(resolve => setTimeout(resolve, 5000));
  }
}
```

### Анализ ошибок
```bash
#!/bin/bash
# Анализ ошибок за последние 24 часа
SINCE=$(date -d '24 hours ago' --iso-8601=seconds)

curl -s -H "X-API-Key: $API_KEY" \
  "/api/v1/daemon/events?level=ERROR&since=$SINCE&page_size=200" | \
  jq -r '.data.events[] | "\(.timestamp) [\(.component)] \(.message)"'
```

### Проверка здоровья системы
```javascript
async function checkSystemHealth() {
  const response = await fetch('/api/v1/daemon/events?level=ERROR,FATAL&since=' + 
    new Date(Date.now() - 3600000).toISOString(), {
    headers: { 'X-API-Key': 'your-api-key' }
  });
  
  const data = await response.json();
  const recentErrors = data.data.events;
  
  if (recentErrors.length > 10) {
    console.warn(`High error rate: ${recentErrors.length} errors in last hour`);
  }
  
  // Группировка по компонентам
  const errorsByComponent = recentErrors.reduce((acc, event) => {
    acc[event.component] = (acc[event.component] || 0) + 1;
    return acc;
  }, {});
  
  return errorsByComponent;
}
```

## Алертинг

### Настройка алертов
```yaml
# alerts.yaml
alerts:
  - name: "High Error Rate"
    condition: "error_count > 50 in 5m"
    action: "send_email"
    
  - name: "Component Failed"
    condition: "event_type = COMPONENT_FAILED"
    action: "send_slack"
    
  - name: "Database Issues"
    condition: "component = storage AND level = ERROR"
    action: "send_pagerduty"
```

### Webhook уведомления
```javascript
// Обработка критических событий
app.post('/webhook/events', (req, res) => {
  const event = req.body;
  
  if (event.level === 'FATAL' || event.event_type === 'COMPONENT_FAILED') {
    sendSlackAlert({
      channel: '#ops-alerts',
      message: `🚨 Critical event: ${event.message}`,
      details: event.details
    });
  }
  
  res.status(200).send('OK');
});
```

## Производительность

### Индексы
- По `timestamp` (для временной фильтрации)
- По `level` (для фильтрации по уровню)
- По `component` (для фильтрации по компоненту)

### Ретенция
- **INFO events**: 30 дней
- **WARN events**: 90 дней  
- **ERROR events**: 1 год
- **FATAL events**: Постоянно

## Связанные endpoints
- [`GET /api/v1/daemon/status`](./daemon-status.md) - Текущий статус демона
- [`GET /api/v1/system/status`](../system/system-status.md) - Статус всей системы
- [`GET /api/v1/incidents`](../incidents/list-incidents.md) - Инциденты системы
