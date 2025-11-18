# POST /api/v1/daemon/start

## Описание
Запуск демона Atom Engine. Инициализирует все компоненты системы.

## URL
```
POST /api/v1/daemon/start
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Параметры тела запроса (опциональные)
```json
{
  "config_file": "/custom/path/config.yaml",
  "force_restart": false,
  "components": ["process_engine", "timewheel"]
}
```

## Примеры запросов

### Базовый запуск
```bash
curl -X POST "http://localhost:27555/api/v1/daemon/start" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here"
```

### Запуск с параметрами
```bash
curl -X POST "http://localhost:27555/api/v1/daemon/start" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "config_file": "/etc/atom-engine/production.yaml",
    "force_restart": true
  }'
```

## Ответы

### 200 OK - Демон запущен
```json
{
  "success": true,
  "data": {
    "status": "STARTING",
    "pid": 12346,
    "started_at": "2025-01-11T10:30:00.000Z",
    "config_file": "/opt/atom-engine/config/config.yaml",
    "components_starting": [
      "core",
      "storage", 
      "process_engine",
      "timewheel",
      "message_system",
      "job_manager"
    ],
    "estimated_startup_time_seconds": 15
  },
  "request_id": "req_1641998400900"
}
```

### 409 Conflict - Демон уже работает
```json
{
  "success": false,
  "error": {
    "code": "DAEMON_ALREADY_RUNNING",
    "message": "Daemon is already running",
    "details": {
      "current_pid": 12345,
      "started_at": "2025-01-10T10:30:00.000Z",
      "uptime_seconds": 86400
    }
  },
  "request_id": "req_1641998400901"
}
```

### 400 Bad Request - Ошибка конфигурации
```json
{
  "success": false,
  "error": {
    "code": "CONFIG_ERROR",
    "message": "Configuration file error",
    "details": {
      "config_file": "/custom/path/config.yaml",
      "error": "File not found or invalid format"
    }
  },
  "request_id": "req_1641998400902"
}
```

## Параметры запроса

### config_file (опционально)
- Путь к альтернативному файлу конфигурации
- По умолчанию: `/opt/atom-engine/config/config.yaml`

### force_restart (опционально)
- `true` - Принудительно перезапустить если демон уже работает
- `false` - Вернуть ошибку если демон уже работает (по умолчанию)

### components (опционально)
- Массив компонентов для запуска
- По умолчанию: все доступные компоненты

## Процесс запуска

### Этапы инициализации
1. **Валидация конфигурации** - Проверка config файла
2. **Инициализация core** - Запуск ядра системы
3. **Запуск storage** - Подключение к базе данных
4. **Инициализация компонентов** - Последовательный запуск
5. **Готовность к работе** - Все компоненты готовы

### Время запуска
- **Холодный старт**: 15-30 секунд
- **Теплый старт**: 5-10 секунд
- **Восстановление**: 10-20 секунд

## Мониторинг запуска

### Ожидание готовности
```bash
#!/bin/bash
# Запуск и ожидание готовности
curl -X POST -H "X-API-Key: $API_KEY" /api/v1/daemon/start

# Ожидание готовности
while true; do
  STATUS=$(curl -s -H "X-API-Key: $API_KEY" /api/v1/daemon/status | jq -r '.data.status')
  if [ "$STATUS" = "RUNNING" ]; then
    echo "Daemon started successfully"
    break
  elif [ "$STATUS" = "FAILED" ]; then
    echo "Daemon startup failed"
    exit 1
  fi
  echo "Starting... Status: $STATUS"
  sleep 2
done
```

### JavaScript мониторинг
```javascript
async function startDaemonAndWait() {
  // Запуск демона
  const startResponse = await fetch('/api/v1/daemon/start', {
    method: 'POST',
    headers: { 'X-API-Key': 'your-api-key' }
  });
  
  if (!startResponse.ok) {
    throw new Error('Failed to start daemon');
  }
  
  // Ожидание готовности
  while (true) {
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    const statusResponse = await fetch('/api/v1/daemon/status');
    const status = await statusResponse.json();
    
    if (status.data.status === 'RUNNING') {
      console.log('Daemon is ready');
      break;
    } else if (status.data.status === 'FAILED') {
      throw new Error('Daemon startup failed');
    }
    
    console.log(`Starting... Status: ${status.data.status}`);
  }
}
```

## Troubleshooting

### Частые проблемы
1. **Config file not found** - Проверьте путь к конфигурации
2. **Port already in use** - Проверьте доступность портов
3. **Database connection failed** - Проверьте настройки БД
4. **Insufficient permissions** - Проверьте права доступа

### Логи запуска
```bash
# Просмотр логов запуска
tail -f /opt/atom-engine/logs/app.log | grep "startup"

# Фильтрация ошибок
tail -f /opt/atom-engine/logs/app.log | grep "ERROR"
```

## Связанные endpoints
- [`GET /api/v1/daemon/status`](./daemon-status.md) - Проверка статуса
- [`POST /api/v1/daemon/stop`](./daemon-stop.md) - Остановка демона
- [`GET /api/v1/daemon/events`](./daemon-events.md) - События запуска
