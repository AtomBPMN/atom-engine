# POST /api/v1/daemon/stop

## Описание
Graceful остановка демона Atom Engine с сохранением состояния и завершением активных операций.

## URL
```
POST /api/v1/daemon/stop
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Параметры тела запроса (опциональные)
```json
{
  "timeout_seconds": 30,
  "force": false,
  "reason": "Maintenance shutdown"
}
```

## Примеры запросов

### Graceful остановка
```bash
curl -X POST "http://localhost:27555/api/v1/daemon/stop" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here"
```

### Принудительная остановка
```bash
curl -X POST "http://localhost:27555/api/v1/daemon/stop" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "timeout_seconds": 10,
    "force": true,
    "reason": "Emergency shutdown"
  }'
```

## Ответы

### 200 OK - Демон останавливается
```json
{
  "success": true,
  "data": {
    "status": "STOPPING",
    "shutdown_initiated_at": "2025-01-11T10:30:00.000Z",
    "estimated_shutdown_time_seconds": 15,
    "active_operations": {
      "processes": 5,
      "jobs": 12,
      "timers": 89
    },
    "shutdown_sequence": [
      "Stop accepting new requests",
      "Complete active jobs",
      "Save process states", 
      "Shutdown components",
      "Close database connections"
    ]
  },
  "request_id": "req_1641998401000"
}
```

### 404 Not Found - Демон не работает
```json
{
  "success": false,
  "error": {
    "code": "DAEMON_NOT_RUNNING",
    "message": "Daemon is not running",
    "details": {
      "current_status": "STOPPED"
    }
  },
  "request_id": "req_1641998401001"
}
```

## Параметры остановки

### timeout_seconds (опционально)
- Максимальное время ожидания graceful shutdown
- По умолчанию: 30 секунд
- После таймаута выполняется принудительная остановка

### force (опционально)
- `true` - Принудительная остановка без ожидания
- `false` - Graceful остановка (по умолчанию)

### reason (опционально)
- Причина остановки для логирования
- Записывается в audit log

## Процесс остановки

### Graceful Shutdown Sequence
1. **Stop accepting requests** - Новые запросы отклоняются
2. **Complete active operations** - Завершение текущих операций
3. **Save state** - Сохранение состояния процессов
4. **Shutdown components** - Последовательная остановка
5. **Close connections** - Закрытие сетевых соединений
6. **Exit** - Завершение процесса

### Время остановки
- **Graceful shutdown**: 10-30 секунд
- **Force shutdown**: 1-5 секунд
- **Emergency shutdown**: немедленно

## Мониторинг остановки

### Ожидание завершения
```bash
#!/bin/bash
# Остановка и ожидание завершения
curl -X POST -H "X-API-Key: $API_KEY" /api/v1/daemon/stop

# Ожидание полной остановки
while true; do
  STATUS=$(curl -s -H "X-API-Key: $API_KEY" /api/v1/daemon/status | jq -r '.data.status' 2>/dev/null)
  if [ "$STATUS" = "STOPPED" ] || [ -z "$STATUS" ]; then
    echo "Daemon stopped successfully"
    break
  fi
  echo "Stopping... Status: $STATUS"
  sleep 2
done
```

### JavaScript мониторинг
```javascript
async function stopDaemonAndWait() {
  // Инициация остановки
  const stopResponse = await fetch('/api/v1/daemon/stop', {
    method: 'POST',
    headers: { 
      'Content-Type': 'application/json',
      'X-API-Key': 'your-api-key' 
    },
    body: JSON.stringify({
      reason: 'Controlled shutdown',
      timeout_seconds: 30
    })
  });
  
  if (!stopResponse.ok) {
    throw new Error('Failed to stop daemon');
  }
  
  // Мониторинг процесса остановки
  while (true) {
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    try {
      const statusResponse = await fetch('/api/v1/daemon/status');
      const status = await statusResponse.json();
      
      if (status.data.status === 'STOPPED') {
        console.log('Daemon stopped successfully');
        break;
      }
      
      console.log(`Stopping... Status: ${status.data.status}`);
    } catch (error) {
      // Daemon полностью остановлен
      console.log('Daemon stopped completely');
      break;
    }
  }
}
```

## Сохранение состояния

### Что сохраняется
- **Активные процессы** - Состояние и переменные
- **Таймеры** - Remaining time и метаданные
- **Задания** - Queue state и worker assignments
- **Сообщения** - Буферизованные сообщения
- **Конфигурация** - Runtime конфигурация

### Recovery после перезапуска
```javascript
// Проверка восстановленного состояния
async function checkRecoveryState() {
  const status = await fetch('/api/v1/system/status').then(r => r.json());
  
  console.log(`Recovered processes: ${status.data.components.process_engine.active_processes}`);
  console.log(`Recovered timers: ${status.data.components.timewheel.active_timers}`);
  console.log(`Recovered jobs: ${status.data.components.job_manager.active_jobs}`);
}
```

## Высокая доступность

### Rolling Restart
```bash
#!/bin/bash
# Rolling restart для кластера
NODES=("node1" "node2" "node3")

for node in "${NODES[@]}"; do
  echo "Stopping node: $node"
  curl -X POST -H "X-API-Key: $API_KEY" \
    "http://$node:8080/api/v1/daemon/stop"
  
  # Ожидание остановки
  while curl -s "http://$node:8080/health" >/dev/null 2>&1; do
    sleep 2
  done
  
  echo "Starting node: $node"
  ssh $node "systemctl start atom-engine"
  
  # Ожидание готовности
  while ! curl -s "http://$node:8080/health" >/dev/null 2>&1; do
    sleep 2
  done
  
  echo "Node $node restarted successfully"
done
```

## Связанные endpoints
- [`GET /api/v1/daemon/status`](./daemon-status.md) - Проверка статуса
- [`POST /api/v1/daemon/start`](./daemon-start.md) - Запуск демона
- [`GET /api/v1/daemon/events`](./daemon-events.md) - События остановки
