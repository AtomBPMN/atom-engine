# GET /api/v1/system/status

## Описание
Получение детального статуса системы и всех компонентов. Предоставляет полную диагностическую информацию о состоянии Atom Engine.

## URL
```
GET /api/v1/system/status
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

```http
X-API-Key: your-api-key-here
```

## Параметры запроса
Отсутствуют.

## Заголовки запроса
```http
Accept: application/json
X-API-Key: your-api-key-here
```

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/status" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/system/status', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const systemStatus = await response.json();
console.log('System health:', systemStatus.data.health);
```

### Go
```go
req, _ := http.NewRequest("GET", "/api/v1/system/status", nil)
req.Header.Set("X-API-Key", "your-api-key-here")

client := &http.Client{}
resp, err := client.Do(req)
```

## Ответы

### 200 OK - Система работает
```json
{
  "success": true,
  "data": {
    "status": "HEALTHY",
    "health": "GOOD",
    "version": "1.0.0",
    "build": "2025.01.11-abc123",
    "uptime_seconds": 86400,
    "started_at": "2025-01-10T10:30:00.000Z",
    "node_id": "srv1",
    "environment": "production",
    "components": {
      "core": {
        "status": "HEALTHY",
        "uptime_seconds": 86400,
        "last_heartbeat": "2025-01-11T10:30:00.000Z"
      },
      "storage": {
        "status": "HEALTHY",
        "connection_pool": {
          "active": 5,
          "idle": 15,
          "max": 20
        },
        "database_size_mb": 1024,
        "last_backup": "2025-01-11T06:00:00.000Z"
      },
      "process_engine": {
        "status": "HEALTHY",
        "active_processes": 142,
        "total_processes": 15847,
        "tokens_per_second": 45.2
      },
      "timewheel": {
        "status": "HEALTHY",
        "active_timers": 89,
        "wheel_levels": 5,
        "current_tick": 1641998400000
      },
      "message_system": {
        "status": "HEALTHY",
        "active_subscriptions": 23,
        "buffered_messages": 7,
        "correlation_rate": 98.5
      },
      "job_manager": {
        "status": "HEALTHY",
        "active_jobs": 34,
        "active_workers": 12,
        "jobs_per_minute": 150
      },
      "expression_engine": {
        "status": "HEALTHY",
        "cache_hit_ratio": 92.3,
        "evaluations_per_second": 250
      },
      "incident_manager": {
        "status": "HEALTHY",
        "open_incidents": 2,
        "resolved_today": 15
      }
    },
    "resources": {
      "memory": {
        "used_mb": 512,
        "available_mb": 2048,
        "usage_percent": 25.0
      },
      "cpu": {
        "usage_percent": 15.5,
        "load_average": [0.8, 0.9, 1.1]
      },
      "disk": {
        "used_gb": 8.5,
        "available_gb": 40.2,
        "usage_percent": 17.4
      }
    },
    "network": {
      "grpc_port": 27500,
      "rest_port": 8080,
      "active_connections": 45
    }
  },
  "request_id": "req_1641998400500"
}
```

### 503 Service Unavailable - Проблемы в системе
```json
{
  "success": false,
  "data": {
    "status": "DEGRADED",
    "health": "POOR",
    "version": "1.0.0", 
    "uptime_seconds": 86400,
    "components": {
      "storage": {
        "status": "UNHEALTHY",
        "error": "Connection timeout to database",
        "last_error_at": "2025-01-11T10:29:30.000Z"
      },
      "process_engine": {
        "status": "DEGRADED",
        "active_processes": 142,
        "error": "High memory usage detected"
      }
    },
    "critical_errors": [
      {
        "component": "storage",
        "message": "Database connection pool exhausted",
        "occurred_at": "2025-01-11T10:29:30.000Z"
      }
    ]
  },
  "error": {
    "code": "SYSTEM_UNHEALTHY",
    "message": "System is in degraded state",
    "details": {
      "unhealthy_components": ["storage"],
      "degraded_components": ["process_engine"]
    }
  },
  "request_id": "req_1641998400501"
}
```

## Поля ответа

### System Status
- `status` (string): Общий статус системы (`HEALTHY`, `DEGRADED`, `UNHEALTHY`)
- `health` (string): Здоровье системы (`GOOD`, `FAIR`, `POOR`, `CRITICAL`)
- `version` (string): Версия Atom Engine
- `build` (string): Build информация
- `uptime_seconds` (integer): Время работы в секундах
- `started_at` (string): Время запуска системы
- `node_id` (string): Идентификатор узла
- `environment` (string): Окружение (development, staging, production)

### Component Status
Каждый компонент содержит:
- `status` (string): Статус компонента
- `uptime_seconds` (integer): Время работы компонента
- `last_heartbeat` (string): Последний heartbeat
- Специфичные метрики для каждого компонента

### Resource Information
- `memory` - Использование памяти
- `cpu` - Загрузка процессора  
- `disk` - Использование диска
- `network` - Сетевая информация

## Статусы компонентов

### Возможные значения
- `HEALTHY` - Компонент работает нормально
- `DEGRADED` - Частичная функциональность
- `UNHEALTHY` - Компонент не работает
- `UNKNOWN` - Статус неизвестен
- `STARTING` - Компонент запускается
- `STOPPING` - Компонент останавливается

### Критерии статусов
```yaml
HEALTHY:
  - Все функции работают
  - Производительность в норме
  - Нет ошибок

DEGRADED:
  - Основные функции работают
  - Снижена производительность
  - Есть предупреждения

UNHEALTHY:
  - Критические функции не работают
  - Много ошибок
  - Требует вмешательства
```

## Мониторинг

### Health Check Pipeline
```javascript
async function systemHealthCheck() {
  const response = await fetch('/api/v1/system/status');
  const status = await response.json();
  
  if (status.data.health === 'POOR' || status.data.health === 'CRITICAL') {
    await alerting.sendAlert({
      severity: 'HIGH',
      message: `System health is ${status.data.health}`,
      components: status.data.components
    });
  }
  
  return status.data;
}
```

### Prometheus Metrics
```yaml
# Экспорт метрик для Prometheus
atom_system_status{status="HEALTHY"} 1
atom_component_status{component="storage",status="HEALTHY"} 1
atom_uptime_seconds 86400
atom_active_processes 142
atom_memory_usage_percent 25.0
```

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Atom Engine System Status",
    "panels": [
      {
        "title": "System Health",
        "targets": ["atom_system_status"]
      },
      {
        "title": "Component Status",
        "targets": ["atom_component_status"]
      }
    ]
  }
}
```

## Использование

### Operations Dashboard
```javascript
// Получение данных для ops dashboard
async function getSystemOverview() {
  const [status, metrics] = await Promise.all([
    fetch('/api/v1/system/status'),
    fetch('/api/v1/system/metrics')
  ]);
  
  return {
    status: await status.json(),
    metrics: await metrics.json()
  };
}
```

### Automated Health Checks
```bash
#!/bin/bash
# Скрипт проверки здоровья системы
HEALTH=$(curl -s -H "X-API-Key: $API_KEY" /api/v1/system/status | jq -r '.data.health')

case $HEALTH in
  "GOOD"|"FAIR")
    echo "System is healthy: $HEALTH"
    exit 0
    ;;
  "POOR")
    echo "System degraded: $HEALTH"
    exit 1
    ;;
  "CRITICAL")
    echo "System critical: $HEALTH"
    exit 2
    ;;
esac
```

### Load Balancer Integration
```nginx
# Nginx health check configuration
upstream atom_backend {
    server 127.0.0.1:8080;
    health_check uri=/api/v1/system/status
                 match=atom_healthy;
}

match atom_healthy {
    status 200;
    header Content-Type ~ "application/json";
    body ~ '"health":"GOOD"';
}
```

## Производительность

### Время ответа
- **Нормальные условия**: < 50ms
- **Высокая нагрузка**: < 200ms
- **Таймаут**: 5 секунд

### Кэширование
- Результат кэшируется на 10 секунд
- Инвалидация при изменении статуса компонентов
- Отдельный кэш для каждого API ключа

### Оптимизация
```go
// Параллельный сбор статусов компонентов
func collectComponentStatuses() map[string]ComponentStatus {
    components := []string{"storage", "process_engine", "timewheel"}
    results := make(chan ComponentStatusResult, len(components))
    
    for _, component := range components {
        go func(comp string) {
            results <- ComponentStatusResult{
                Name: comp,
                Status: getComponentStatus(comp),
            }
        }(component)
    }
    
    // Сбор результатов с таймаутом
    statusMap := make(map[string]ComponentStatus)
    timeout := time.After(2 * time.Second)
    
    for i := 0; i < len(components); i++ {
        select {
        case result := <-results:
            statusMap[result.Name] = result.Status
        case <-timeout:
            log.Warn("Component status collection timeout")
            return statusMap
        }
    }
    
    return statusMap
}
```

## Troubleshooting

### Частые проблемы
1. **Медленный ответ**: Проверьте загрузку системы
2. **503 ошибка**: Проверьте статус компонентов
3. **Timeout**: Проверьте сетевую связность

### Диагностика
```bash
# Проверка статуса системы
curl -w "%{time_total}\n" -H "X-API-Key: $API_KEY" /api/v1/system/status

# Проверка конкретного компонента
curl -H "X-API-Key: $API_KEY" /api/v1/system/components/storage

# Мониторинг в реальном времени
watch -n 5 'curl -s -H "X-API-Key: $API_KEY" /api/v1/system/status | jq .data.health'
```

## Связанные endpoints
- [`GET /health`](../health/health-check.md) - Простая проверка доступности
- [`GET /api/v1/system/info`](./system-info.md) - Системная информация
- [`GET /api/v1/system/metrics`](./system-metrics.md) - Детальные метрики
- [`GET /api/v1/system/components`](./list-components.md) - Список всех компонентов
