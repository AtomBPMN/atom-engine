# GET /api/v1/storage/info

## Описание
Получение детальной информации о системе хранения данных: размер, статистика, конфигурация и метрики использования.

## URL
```
GET /api/v1/storage/info
```

## Авторизация
✅ **Требуется API ключ** с разрешением `storage`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/storage/info" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/storage/info', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const storageInfo = await response.json();
```

## Ответы

### 200 OK - Информация получена
```json
{
  "success": true,
  "data": {
    "total_size_bytes": 1073741824,
    "used_size_bytes": 134217728,
    "free_size_bytes": 939524096,
    "total_keys": 125847,
    "database_path": "/opt/atom-engine/data/base",
    "database_version": "4.2.0",
    "created_at": "2025-01-01T10:00:00.000Z",
    "last_compaction": "2025-01-11T06:00:00.000Z",
    "storage_config": {
      "compression_enabled": true,
      "encryption_enabled": false,
      "sync_writes": false,
      "cache_size_mb": 256,
      "max_table_size_mb": 64,
      "background_compaction": true
    },
    "statistics": {
      "total_reads": 15648923,
      "total_writes": 2847561,
      "total_deletes": 125478,
      "compaction_count": 24,
      "last_backup_size_bytes": 98765432,
      "backup_frequency_hours": 24,
      "retention_days": 90
    },
    "tables": {
      "processes": {
        "key_count": 15847,
        "size_bytes": 67108864,
        "last_accessed": "2025-01-11T10:30:00.000Z"
      },
      "tokens": {
        "key_count": 45231,
        "size_bytes": 23456789,
        "last_accessed": "2025-01-11T10:29:55.000Z"
      },
      "timers": {
        "key_count": 5632,
        "size_bytes": 8912345,
        "last_accessed": "2025-01-11T10:29:50.000Z"
      },
      "jobs": {
        "key_count": 67834,
        "size_bytes": 34567890,
        "last_accessed": "2025-01-11T10:29:58.000Z"
      },
      "messages": {
        "key_count": 23456,
        "size_bytes": 12345678,
        "last_accessed": "2025-01-11T10:29:45.000Z"
      }
    },
    "disk_usage": {
      "partition": "/dev/sda1",
      "filesystem": "ext4",
      "total_disk_gb": 50.0,
      "used_disk_gb": 8.5,
      "available_disk_gb": 40.2,
      "usage_percent": 17.4
    },
    "performance_stats": {
      "avg_read_time_us": 2100,
      "avg_write_time_us": 8500,
      "p95_read_time_us": 5000,
      "p95_write_time_us": 15000,
      "cache_hit_ratio": 0.942,
      "bloom_filter_false_positives": 0.001
    }
  },
  "request_id": "req_1641998401300"
}
```

## Поля ответа

### Storage Size Information
- `total_size_bytes` (integer): Общий размер БД в байтах
- `used_size_bytes` (integer): Используемый размер
- `free_size_bytes` (integer): Свободное место
- `total_keys` (integer): Общее количество ключей

### Database Information
- `database_path` (string): Путь к файлам БД
- `database_version` (string): Версия БД
- `created_at` (string): Время создания БД
- `last_compaction` (string): Последняя компактификация

### Configuration
- `compression_enabled` (boolean): Сжатие включено
- `encryption_enabled` (boolean): Шифрование включено
- `sync_writes` (boolean): Синхронные записи
- `cache_size_mb` (integer): Размер кэша в MB
- `background_compaction` (boolean): Фоновая компактификация

### Statistics
- `total_reads` (integer): Общее количество чтений
- `total_writes` (integer): Общее количество записей
- `total_deletes` (integer): Общее количество удалений
- `compaction_count` (integer): Количество компактификаций

### Table Information
Для каждой таблицы:
- `key_count` (integer): Количество ключей
- `size_bytes` (integer): Размер в байтах
- `last_accessed` (string): Последний доступ

### Performance Statistics
- `avg_read_time_us` (integer): Среднее время чтения (микросекунды)
- `avg_write_time_us` (integer): Среднее время записи
- `p95_read_time_us` (integer): 95-й перцентиль чтения
- `cache_hit_ratio` (float): Коэффициент попадания в кэш

## Анализ использования

### Capacity Planning
```javascript
async function analyzeStorageCapacity() {
  const info = await getStorageInfo();
  
  const usagePercent = (info.used_size_bytes / info.total_size_bytes) * 100;
  const diskUsagePercent = info.disk_usage.usage_percent;
  
  console.log(`Database usage: ${usagePercent.toFixed(1)}%`);
  console.log(`Disk usage: ${diskUsagePercent}%`);
  
  // Прогноз заполнения
  const growthRate = calculateGrowthRate(info);
  const daysUntilFull = calculateDaysUntilFull(info, growthRate);
  
  if (daysUntilFull < 30) {
    console.warn(`Storage will be full in ${daysUntilFull} days`);
  }
  
  return {
    usagePercent,
    diskUsagePercent,
    daysUntilFull
  };
}

function calculateGrowthRate(info) {
  // Анализ роста за последние дни
  // Возвращает байт/день
  return info.used_size_bytes / 30; // упрощенная формула
}
```

### Table Analysis
```javascript
async function analyzeTableUsage() {
  const info = await getStorageInfo();
  const tables = info.tables;
  
  // Сортировка таблиц по размеру
  const sortedTables = Object.entries(tables)
    .sort(([,a], [,b]) => b.size_bytes - a.size_bytes)
    .map(([name, data]) => ({
      name,
      size_mb: (data.size_bytes / 1024 / 1024).toFixed(1),
      key_count: data.key_count,
      avg_key_size: (data.size_bytes / data.key_count).toFixed(0)
    }));
  
  console.table(sortedTables);
  
  return sortedTables;
}
```

### Performance Analysis
```go
// Анализ производительности хранилища
func analyzeStoragePerformance(info StorageInfo) {
    stats := info.PerformanceStats
    
    // Проверка времени чтения
    if stats.AvgReadTimeUs > 5000 { // > 5ms
        log.Warnf("High read latency: %d μs", stats.AvgReadTimeUs)
    }
    
    // Проверка времени записи
    if stats.AvgWriteTimeUs > 20000 { // > 20ms
        log.Warnf("High write latency: %d μs", stats.AvgWriteTimeUs)
    }
    
    // Проверка cache hit ratio
    if stats.CacheHitRatio < 0.85 {
        log.Warnf("Low cache hit ratio: %.3f", stats.CacheHitRatio)
    }
}
```

## Мониторинг и алертинг

### Disk Space Monitoring
```bash
#!/bin/bash
# Мониторинг места на диске
USAGE=$(curl -s -H "X-API-Key: $API_KEY" /api/v1/storage/info | \
  jq -r '.data.disk_usage.usage_percent')

if (( $(echo "$USAGE > 80" | bc -l) )); then
  echo "WARNING: Disk usage is ${USAGE}%"
  exit 1
elif (( $(echo "$USAGE > 90" | bc -l) )); then
  echo "CRITICAL: Disk usage is ${USAGE}%"
  exit 2
fi

echo "Disk usage is ${USAGE}% - OK"
```

### Growth Rate Alert
```javascript
async function checkGrowthRate() {
  const info = await getStorageInfo();
  const currentSize = info.used_size_bytes;
  
  // Сравнение с размером 24 часа назад
  const previousSize = await getPreviousDaySize();
  const growthBytes = currentSize - previousSize;
  const growthPercent = (growthBytes / previousSize) * 100;
  
  if (growthPercent > 10) { // Рост более 10% в день
    await sendAlert({
      type: 'HIGH_GROWTH_RATE',
      message: `Storage growing at ${growthPercent.toFixed(1)}% per day`,
      current_size_gb: (currentSize / 1024 / 1024 / 1024).toFixed(2),
      growth_gb: (growthBytes / 1024 / 1024 / 1024).toFixed(2)
    });
  }
}
```

### Compaction Monitoring
```javascript
async function monitorCompaction() {
  const info = await getStorageInfo();
  const lastCompaction = new Date(info.last_compaction);
  const hoursAgo = (Date.now() - lastCompaction.getTime()) / (1000 * 60 * 60);
  
  if (hoursAgo > 48) { // Нет компактификации более 48 часов
    console.warn(`Last compaction was ${hoursAgo.toFixed(1)} hours ago`);
  }
  
  // Анализ эффективности компактификации
  const compressionRatio = info.used_size_bytes / info.total_size_bytes;
  if (compressionRatio < 0.7) {
    console.info('Storage may benefit from compaction');
  }
}
```

## Оптимизация

### Storage Optimization Recommendations
```javascript
async function getOptimizationRecommendations() {
  const info = await getStorageInfo();
  const recommendations = [];
  
  // Анализ конфигурации
  if (!info.storage_config.compression_enabled) {
    recommendations.push({
      type: 'CONFIG',
      message: 'Enable compression to reduce storage size',
      estimated_savings: '20-40%'
    });
  }
  
  if (info.storage_config.cache_size_mb < 512) {
    recommendations.push({
      type: 'CONFIG',
      message: 'Increase cache size for better performance',
      current: `${info.storage_config.cache_size_mb}MB`,
      recommended: '512MB+'
    });
  }
  
  // Анализ производительности
  if (info.performance_stats.cache_hit_ratio < 0.9) {
    recommendations.push({
      type: 'PERFORMANCE',
      message: 'Consider increasing cache size or optimizing access patterns',
      current_hit_ratio: info.performance_stats.cache_hit_ratio
    });
  }
  
  return recommendations;
}
```

### Backup Size Trend
```javascript
async function analyzeBackupTrend() {
  const info = await getStorageInfo();
  
  // Сравнение размера бэкапа с текущим размером БД
  const backupSizeGB = info.statistics.last_backup_size_bytes / (1024 ** 3);
  const currentSizeGB = info.used_size_bytes / (1024 ** 3);
  
  const compressionRatio = backupSizeGB / currentSizeGB;
  
  console.log(`Backup compression ratio: ${(compressionRatio * 100).toFixed(1)}%`);
  
  if (compressionRatio > 0.9) {
    console.warn('Low backup compression - consider enabling compression');
  }
}
```

## Связанные endpoints
- [`GET /api/v1/storage/status`](./storage-status.md) - Статус хранилища
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
