# GET /api/v1/daemon/status

## Описание
Получение статуса демона Atom Engine и информации о его работе.

## URL
```
GET /api/v1/daemon/status
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/daemon/status" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/daemon/status', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const daemonStatus = await response.json();
```

## Ответы

### 200 OK - Демон работает
```json
{
  "success": true,
  "data": {
    "status": "RUNNING",
    "pid": 12345,
    "started_at": "2025-01-10T10:30:00.000Z",
    "uptime_seconds": 86400,
    "version": "1.0.0",
    "config_file": "/opt/atom-engine/config/config.yaml",
    "working_directory": "/opt/atom-engine",
    "log_file": "/opt/atom-engine/logs/app.log",
    "components": {
      "initialized": 8,
      "running": 8,
      "failed": 0
    },
    "ports": {
      "grpc": 27500,
      "rest": 27555
    },
    "memory": {
      "used_mb": 512,
      "resident_mb": 256
    }
  },
  "request_id": "req_1641998400800"
}
```

### 503 Service Unavailable - Демон не работает
```json
{
  "success": false,
  "error": {
    "code": "DAEMON_NOT_RUNNING",
    "message": "Daemon is not running",
    "details": {
      "last_known_status": "STOPPED",
      "stopped_at": "2025-01-11T09:00:00.000Z"
    }
  },
  "request_id": "req_1641998400801"
}
```

## Поля ответа

### Daemon Status
- `status` (string): Статус демона (`STARTING`, `RUNNING`, `STOPPING`, `STOPPED`)
- `pid` (integer): Process ID демона
- `started_at` (string): Время запуска
- `uptime_seconds` (integer): Время работы в секундах
- `version` (string): Версия демона

### Configuration
- `config_file` (string): Путь к файлу конфигурации
- `working_directory` (string): Рабочая директория
- `log_file` (string): Путь к файлу логов

### Component Status
- `initialized` (integer): Количество инициализированных компонентов
- `running` (integer): Количество работающих компонентов
- `failed` (integer): Количество сбойных компонентов

### Resource Usage
- `memory` (object): Использование памяти
- `ports` (object): Используемые порты

## Статусы демона

### Возможные статусы
- `STARTING` - Демон запускается
- `RUNNING` - Демон работает нормально
- `STOPPING` - Демон останавливается
- `STOPPED` - Демон остановлен
- `FAILED` - Демон завершился с ошибкой

## Использование

### Проверка готовности
```bash
#!/bin/bash
# Ожидание готовности демона
while true; do
  STATUS=$(curl -s -H "X-API-Key: $API_KEY" /api/v1/daemon/status | jq -r '.data.status')
  if [ "$STATUS" = "RUNNING" ]; then
    echo "Daemon is ready"
    break
  fi
  echo "Waiting for daemon... Status: $STATUS"
  sleep 2
done
```

### Мониторинг процесса
```javascript
async function monitorDaemon() {
  const response = await fetch('/api/v1/daemon/status');
  const status = await response.json();
  
  if (status.data.components.failed > 0) {
    console.error(`${status.data.components.failed} components failed`);
  }
  
  const uptimeHours = status.data.uptime_seconds / 3600;
  console.log(`Daemon running for ${uptimeHours.toFixed(1)} hours`);
}
```

## Связанные endpoints
- [`POST /api/v1/daemon/start`](./daemon-start.md) - Запуск демона
- [`POST /api/v1/daemon/stop`](./daemon-stop.md) - Остановка демона
- [`GET /api/v1/daemon/events`](./daemon-events.md) - События демона
