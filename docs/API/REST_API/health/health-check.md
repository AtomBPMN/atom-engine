# GET /health

## Описание
Проверка доступности и состояния системы. Единственный endpoint, не требующий авторизации.

## URL
```
GET /health
```

## Авторизация
❌ **Не требуется** - Public endpoint

## Параметры запроса
Отсутствуют.

## Заголовки запроса
```http
Accept: application/json
```

## Тело запроса
Отсутствует.

## Пример запроса

### cURL
```bash
curl -X GET http://localhost:27555/health
```

### JavaScript
```javascript
fetch('/health')
  .then(response => response.json())
  .then(data => console.log(data));
```

### Go
```go
resp, err := http.Get("http://localhost:27555/health")
```

## Ответы

### 200 OK - Система доступна
```json
{
  "status": "ok",
  "timestamp": "2025-01-11T10:30:00Z",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "components": {
    "storage": "healthy",
    "timewheel": "healthy", 
    "process_engine": "healthy"
  }
}
```

### 503 Service Unavailable - Система недоступна
```json
{
  "status": "error",
  "timestamp": "2025-01-11T10:30:00Z",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "components": {
    "storage": "unhealthy",
    "timewheel": "healthy",
    "process_engine": "degraded"
  },
  "errors": [
    "Database connection failed",
    "Process engine running in degraded mode"
  ]
}
```

## Поля ответа

### Основные поля
- `status` (string): Общий статус системы (`ok`, `degraded`, `error`)
- `timestamp` (string): Время ответа в ISO 8601 UTC
- `version` (string): Версия Atom Engine
- `uptime_seconds` (number): Время работы в секундах

### Components
- `storage` (string): Состояние хранилища данных
- `timewheel` (string): Состояние системы таймеров  
- `process_engine` (string): Состояние процессного движка
- `message_system` (string): Состояние системы сообщений
- `expression_engine` (string): Состояние движка выражений

### Возможные значения состояний
- `healthy` - Компонент работает нормально
- `degraded` - Компонент работает с ограниченной функциональностью
- `unhealthy` - Компонент не работает
- `unknown` - Состояние компонента неизвестно

## Коды статуса HTTP
- `200` - Система полностью доступна
- `503` - Система недоступна или работает в ограниченном режиме

## Использование

### Мониторинг
```bash
# Проверка каждые 30 секунд
while true; do
  curl -f /health || echo "Service down!"
  sleep 30
done
```

### Load Balancer Health Check
```yaml
# Nginx upstream health check
upstream atom_engine {
    server 127.0.0.1:8080;
    health_check uri=/health;
}
```

### Kubernetes Readiness/Liveness Probe
```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: atom-engine
    readinessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 10
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 30
```

## Особенности

### Производительность
- **Время ответа**: < 10ms
- **Кэширование**: Не кэшируется
- **Нагрузка**: Минимальная нагрузка на систему

### Детализация ошибок
Health check не раскрывает детальную информацию об ошибках в production целях безопасности. Для детальной диагностики используйте:
- `GET /api/v1/system/status` (требует авторизацию)
- Логи системы
- Метрики мониторинга

## Связанные endpoints
- [`GET /api/v1/system/status`](../system/system-status.md) - Детальная информация о системе
- [`GET /api/v1/system/health`](../system/system-health.md) - Расширенная проверка здоровья
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Метрики системы
