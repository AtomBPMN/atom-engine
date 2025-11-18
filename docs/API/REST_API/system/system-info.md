# GET /api/v1/system/info

## Описание
Получение системной информации Atom Engine: конфигурация, версии, окружение и технические характеристики.

## URL
```
GET /api/v1/system/info
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

```http
X-API-Key: your-api-key-here
```

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/info" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/system/info', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const systemInfo = await response.json();
console.log('Version:', systemInfo.data.version);
```

### Go
```go
req, _ := http.NewRequest("GET", "/api/v1/system/info", nil)
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 200 OK - Информация получена
```json
{
  "success": true,
  "data": {
    "version": {
      "atom_engine": "1.0.0",
      "build_date": "2025-01-11T10:00:00Z",
      "commit_hash": "abc123def456",
      "go_version": "go1.21.5",
      "compiler": "gc",
      "platform": "linux/amd64"
    },
    "configuration": {
      "instance_name": "atom-engine-prod-01",
      "node_id": "srv1",
      "environment": "production",
      "log_level": "INFO",
      "grpc_port": 27500,
      "rest_port": 8080,
      "auth_enabled": true,
      "rate_limiting_enabled": true
    },
    "capabilities": {
      "max_concurrent_processes": 10000,
      "max_timers": 100000,
      "max_jobs_per_worker": 100,
      "supported_bpmn_version": "2.0",
      "zeebe_compatibility": "8.x",
      "expression_engine": "FEEL-1.1"
    },
    "storage": {
      "type": "badger",
      "version": "4.2.0",
      "path": "/data/base",
      "encryption": false,
      "compression": true,
      "backup_enabled": true
    },
    "components": {
      "enabled": [
        "process_engine",
        "timewheel", 
        "message_system",
        "job_manager",
        "expression_engine",
        "incident_manager",
        "grpc_server",
        "rest_server"
      ],
      "disabled": [],
      "versions": {
        "badger": "4.2.0",
        "gin": "1.9.1",
        "grpc": "1.59.0",
        "protobuf": "1.31.0"
      }
    },
    "limits": {
      "max_process_variables_size_mb": 5,
      "max_job_timeout_seconds": 3600,
      "max_timer_duration_years": 100,
      "max_message_ttl_hours": 24,
      "max_api_requests_per_minute": 1000
    },
    "features": {
      "clustering": false,
      "high_availability": false,
      "auto_scaling": false,
      "metrics_export": true,
      "health_checks": true,
      "audit_logging": true,
      "backup_restore": true
    },
    "runtime": {
      "started_at": "2025-01-10T10:30:00.000Z",
      "uptime_seconds": 86400,
      "pid": 12345,
      "working_directory": "/opt/atom-engine",
      "config_file": "/opt/atom-engine/config/config.yaml",
      "timezone": "UTC"
    }
  },
  "request_id": "req_1641998400600"
}
```

## Поля ответа

### Version Information
- `atom_engine` (string): Версия Atom Engine
- `build_date` (string): Дата сборки
- `commit_hash` (string): Hash git commit
- `go_version` (string): Версия Go
- `compiler` (string): Компилятор Go
- `platform` (string): Целевая платформа

### Configuration
- `instance_name` (string): Имя экземпляра
- `node_id` (string): ID узла
- `environment` (string): Окружение
- `log_level` (string): Уровень логирования
- `*_port` (integer): Порты сервисов
- `*_enabled` (boolean): Включенные функции

### Capabilities
- `max_*` (integer): Максимальные лимиты системы
- `supported_*` (string): Поддерживаемые версии/стандарты
- `*_compatibility` (string): Совместимость с другими системами

### Storage Information
- `type` (string): Тип хранилища
- `version` (string): Версия БД
- `path` (string): Путь к данным
- `encryption` (boolean): Шифрование включено
- `compression` (boolean): Сжатие включено

### Component Versions
- `enabled` (array): Включенные компоненты
- `disabled` (array): Отключенные компоненты  
- `versions` (object): Версии зависимостей

### System Limits
- `max_*` (integer): Системные ограничения
- Лимиты на размеры данных, таймауты, скорости

### Runtime Information
- `started_at` (string): Время запуска
- `uptime_seconds` (integer): Время работы
- `pid` (integer): Process ID
- `working_directory` (string): Рабочая директория

## Использование

### Version Check
```javascript
async function checkVersion() {
  const response = await fetch('/api/v1/system/info');
  const info = await response.json();
  
  const currentVersion = info.data.version.atom_engine;
  const latestVersion = await getLatestVersion();
  
  if (currentVersion !== latestVersion) {
    console.warn(`Update available: ${currentVersion} -> ${latestVersion}`);
  }
}
```

### Configuration Audit
```bash
# Проверка конфигурации для аудита
curl -s -H "X-API-Key: $API_KEY" /api/v1/system/info | \
  jq '.data.configuration' > current-config.json

# Сравнение с требуемой конфигурацией
diff expected-config.json current-config.json
```

### Capacity Planning
```javascript
async function checkCapacityLimits() {
  const info = await fetch('/api/v1/system/info').then(r => r.json());
  const status = await fetch('/api/v1/system/status').then(r => r.json());
  
  const limits = info.data.limits;
  const current = status.data.components.process_engine;
  
  const processUsage = (current.active_processes / limits.max_concurrent_processes) * 100;
  
  if (processUsage > 80) {
    console.warn(`Process capacity at ${processUsage.toFixed(1)}%`);
  }
}
```

### Compatibility Check
```go
func checkZeebeCompatibility(info SystemInfo) bool {
    requiredVersion := "8.x"
    return info.Capabilities.ZeebeCompatibility == requiredVersion
}
```

## Security Considerations

### Information Disclosure
- Версии компонентов могут помочь атакующим
- В production скрывайте детальную информацию
- Используйте whitelist для доступа к этому endpoint

### Конфигурация для production
```yaml
# config/config.yaml
system_info:
  show_versions: false      # Скрыть версии зависимостей
  show_paths: false        # Скрыть пути файловой системы
  show_internal_ports: false # Скрыть внутренние порты
```

### Filtered Response (Production)
```json
{
  "success": true,
  "data": {
    "version": {
      "atom_engine": "1.0.0",
      "platform": "linux/amd64"
    },
    "configuration": {
      "instance_name": "atom-engine-prod-01",
      "environment": "production"
    },
    "capabilities": {
      "supported_bpmn_version": "2.0",
      "zeebe_compatibility": "8.x"
    }
  }
}
```

## Мониторинг и Alerting

### Version Drift Detection
```prometheus
# Prometheus правило для version drift
- alert: AtomEngineVersionDrift
  expr: |
    count by (version) (atom_engine_version_info) > 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Multiple Atom Engine versions detected"
    description: "{{ $labels.version }} version drift detected"
```

### Configuration Compliance
```javascript
// Проверка соответствия конфигурации
async function validateConfiguration() {
  const info = await getSystemInfo();
  const compliance = {
    authEnabled: info.configuration.auth_enabled,
    rateLimitingEnabled: info.configuration.rate_limiting_enabled,
    encryptionEnabled: info.storage.encryption,
    logLevel: info.configuration.log_level === 'INFO'
  };
  
  const violations = Object.entries(compliance)
    .filter(([key, value]) => !value)
    .map(([key]) => key);
    
  if (violations.length > 0) {
    await sendComplianceAlert(violations);
  }
}
```

## Performance

### Caching
- Информация кэшируется на 5 минут
- Инвалидация при перезапуске компонентов
- Separate cache per API key

### Response Time
- **Типичное время**: < 20ms
- **Максимальное время**: < 100ms
- **Источник данных**: In-memory конфигурация

## Связанные endpoints
- [`GET /api/v1/system/status`](./system-status.md) - Статус системы
- [`GET /api/v1/system/metrics`](./system-metrics.md) - Метрики производительности
- [`GET /api/v1/system/components`](./list-components.md) - Информация о компонентах
