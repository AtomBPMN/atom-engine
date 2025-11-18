# GET /api/v1/storage/status

## Описание
Получение статуса системы хранения данных Atom Engine: подключение, производительность и состояние базы данных.

## URL
```
GET /api/v1/storage/status
```

## Авторизация
✅ **Требуется API ключ** с разрешением `storage`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/storage/status" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/storage/status', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const storageStatus = await response.json();
```

## Ответы

### 200 OK - Хранилище работает
```json
{
  "success": true,
  "data": {
    "is_connected": true,
    "is_healthy": true,
    "status": "HEALTHY",
    "uptime_seconds": 86400,
    "database_type": "badger",
    "database_version": "4.2.0",
    "connection_pool": {
      "active_connections": 5,
      "idle_connections": 15,
      "max_connections": 20,
      "total_connections": 20
    },
    "performance": {
      "reads_per_second": 245,
      "writes_per_second": 67,
      "avg_read_latency_ms": 2.1,
      "avg_write_latency_ms": 8.5,
      "cache_hit_ratio_percent": 94.2
    },
    "health_checks": {
      "connectivity": {
        "status": "PASS",
        "response_time_ms": 3,
        "last_check": "2025-01-11T10:30:00.000Z"
      },
      "write_test": {
        "status": "PASS",
        "response_time_ms": 12,
        "last_check": "2025-01-11T10:29:00.000Z"
      },
      "read_test": {
        "status": "PASS", 
        "response_time_ms": 2,
        "last_check": "2025-01-11T10:29:00.000Z"
      }
    },
    "last_backup": "2025-01-11T06:00:00.000Z",
    "next_backup": "2025-01-12T06:00:00.000Z"
  },
  "request_id": "req_1641998401200"
}
```

### 503 Service Unavailable - Проблемы с хранилищем
```json
{
  "success": false,
  "data": {
    "is_connected": false,
    "is_healthy": false,
    "status": "UNHEALTHY",
    "error": "Database connection timeout",
    "last_successful_connection": "2025-01-11T10:25:00.000Z",
    "health_checks": {
      "connectivity": {
        "status": "FAIL",
        "error": "Connection timeout after 5000ms",
        "last_check": "2025-01-11T10:30:00.000Z"
      }
    }
  },
  "error": {
    "code": "STORAGE_UNHEALTHY",
    "message": "Storage system is not healthy",
    "details": {
      "failed_checks": ["connectivity", "write_test"]
    }
  },
  "request_id": "req_1641998401201"
}
```

## Поля ответа

### Connection Status
- `is_connected` (boolean): Статус подключения к БД
- `is_healthy` (boolean): Общее здоровье хранилища
- `status` (string): Статус (`HEALTHY`, `DEGRADED`, `UNHEALTHY`)
- `uptime_seconds` (integer): Время работы соединения

### Database Information
- `database_type` (string): Тип базы данных (badger)
- `database_version` (string): Версия БД

### Connection Pool
- `active_connections` (integer): Активные соединения
- `idle_connections` (integer): Простаивающие соединения
- `max_connections` (integer): Максимум соединений
- `total_connections` (integer): Общее количество

### Performance Metrics
- `reads_per_second` (float): Операции чтения в секунду
- `writes_per_second` (float): Операции записи в секунду
- `avg_read_latency_ms` (float): Средняя задержка чтения
- `avg_write_latency_ms` (float): Средняя задержка записи
- `cache_hit_ratio_percent` (float): Процент попаданий в кэш

### Health Checks
Каждая проверка содержит:
- `status` (string): Результат (`PASS`, `FAIL`)
- `response_time_ms` (float): Время ответа
- `last_check` (string): Время последней проверки
- `error` (string): Описание ошибки (если есть)

## Статусы хранилища

### Возможные статусы
- `HEALTHY` - Все функции работают нормально
- `DEGRADED` - Частичная функциональность, снижена производительность
- `UNHEALTHY` - Критические проблемы, система не работает

### Критерии статусов
```yaml
HEALTHY:
  - Все health checks проходят
  - Задержка < 10ms для чтения
  - Задержка < 50ms для записи
  - Cache hit ratio > 85%

DEGRADED:
  - Некоторые checks не проходят
  - Задержка 10-100ms
  - Cache hit ratio 70-85%

UNHEALTHY:
  - Критические checks не проходят
  - Задержка > 100ms
  - Connection pool исчерпан
```

## Мониторинг

### Health Check Script
```bash
#!/bin/bash
# Проверка состояния хранилища
STORAGE_STATUS=$(curl -s -H "X-API-Key: $API_KEY" \
  /api/v1/storage/status | jq -r '.data.status')

case $STORAGE_STATUS in
  "HEALTHY")
    echo "Storage is healthy"
    exit 0
    ;;
  "DEGRADED")
    echo "Storage is degraded"
    exit 1
    ;;
  "UNHEALTHY")
    echo "Storage is unhealthy"
    exit 2
    ;;
esac
```

### JavaScript мониторинг
```javascript
async function monitorStorageHealth() {
  const response = await fetch('/api/v1/storage/status');
  const status = await response.json();
  
  const metrics = status.data.performance;
  
  // Проверка производительности
  if (metrics.avg_read_latency_ms > 10) {
    console.warn(`High read latency: ${metrics.avg_read_latency_ms}ms`);
  }
  
  if (metrics.avg_write_latency_ms > 50) {
    console.warn(`High write latency: ${metrics.avg_write_latency_ms}ms`);
  }
  
  if (metrics.cache_hit_ratio_percent < 85) {
    console.warn(`Low cache hit ratio: ${metrics.cache_hit_ratio_percent}%`);
  }
  
  return status.data;
}
```

### Prometheus Metrics
```prometheus
# Экспорт метрик для Prometheus
atom_storage_connected{type="badger"} 1
atom_storage_read_latency_ms 2.1
atom_storage_write_latency_ms 8.5
atom_storage_cache_hit_ratio 0.942
atom_storage_reads_per_second 245
atom_storage_writes_per_second 67
```

## Troubleshooting

### Частые проблемы

#### Высокая задержка
```bash
# Проверка I/O нагрузки
iostat -x 1

# Проверка места на диске
df -h /data/base

# Анализ slow queries
tail -f /opt/atom-engine/logs/app.log | grep "slow"
```

#### Connection Pool исчерпан
```javascript
// Мониторинг connection pool
async function checkConnectionPool() {
  const status = await getStorageStatus();
  const pool = status.connection_pool;
  
  const utilization = pool.active_connections / pool.max_connections;
  
  if (utilization > 0.8) {
    console.warn(`High connection pool utilization: ${(utilization * 100).toFixed(1)}%`);
  }
}
```

#### Low cache hit ratio
```go
// Анализ cache performance
func analyzeCache(status StorageStatus) {
    hitRatio := status.Performance.CacheHitRatioPercent
    
    if hitRatio < 85 {
        log.Warnf("Low cache hit ratio: %.1f%%", hitRatio)
        
        // Возможные причины:
        // 1. Недостаточно памяти для cache
        // 2. Random access patterns
        // 3. Cache size слишком мал
    }
}
```

## Оптимизация производительности

### Configuration Tuning
```yaml
# config/config.yaml
storage:
  badger:
    cache_size_mb: 512      # Увеличить для лучшего hit ratio
    sync_writes: false      # Async writes для производительности
    compression: true       # Сжатие для экономии места
    background_compact: true # Background compaction
```

### Мониторинг тенденций
```javascript
// Отслеживание тенденций производительности
class StorageMonitor {
  constructor() {
    this.metrics = [];
  }
  
  async collectMetrics() {
    const status = await getStorageStatus();
    this.metrics.push({
      timestamp: Date.now(),
      readLatency: status.performance.avg_read_latency_ms,
      writeLatency: status.performance.avg_write_latency_ms,
      cacheHitRatio: status.performance.cache_hit_ratio_percent
    });
    
    // Keep only last 100 measurements
    if (this.metrics.length > 100) {
      this.metrics.shift();
    }
  }
  
  analyzeTrends() {
    const recent = this.metrics.slice(-10);
    const avgReadLatency = recent.reduce((sum, m) => sum + m.readLatency, 0) / recent.length;
    
    if (avgReadLatency > 5) {
      console.warn('Read latency trending up');
    }
  }
}
```

## Связанные endpoints
- [`GET /api/v1/storage/info`](./storage-info.md) - Детальная информация о хранилище
- [`GET /api/v1/system/status`](../system/system-status.md) - Общий статус системы
