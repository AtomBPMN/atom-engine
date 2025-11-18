# GET /api/v1/system/metrics

## Описание
Получение детальных метрик производительности системы для мониторинга и анализа.

## URL
```
GET /api/v1/system/metrics
```

## Авторизация
✅ **Требуется API ключ** с разрешением `system`

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/system/metrics" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/system/metrics', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});
const metrics = await response.json();
```

## Ответы

### 200 OK - Метрики получены
```json
{
  "success": true,
  "data": {
    "timestamp": "2025-01-11T10:30:00.000Z",
    "collection_duration_ms": 15,
    "system": {
      "uptime_seconds": 86400,
      "cpu_usage_percent": 15.5,
      "memory_usage_mb": 512,
      "memory_total_mb": 2048,
      "disk_usage_gb": 8.5,
      "disk_total_gb": 50.0,
      "load_average": [0.8, 0.9, 1.1],
      "goroutines": 245,
      "gc_cycles": 1247,
      "heap_objects": 125432
    },
    "process_engine": {
      "total_processes": 15847,
      "active_processes": 142,
      "completed_processes": 15705,
      "cancelled_processes": 0,
      "processes_per_minute": 25.4,
      "avg_process_duration_ms": 45000,
      "tokens_per_second": 45.2,
      "gateway_evaluations_per_second": 12.3
    },
    "jobs": {
      "total_jobs": 67834,
      "active_jobs": 34,
      "completed_jobs": 67800,
      "failed_jobs": 0,
      "jobs_per_minute": 150,
      "avg_job_duration_ms": 2500,
      "active_workers": 12,
      "job_types": {
        "email-service": 15,
        "payment-processor": 8,
        "pdf-generator": 11
      }
    },
    "timers": {
      "total_timers": 5632,
      "active_timers": 89,
      "fired_timers": 5543,
      "cancelled_timers": 0,
      "timers_per_minute": 8.7,
      "avg_timer_accuracy_ms": 12,
      "wheel_levels_usage": [45, 23, 15, 5, 1]
    },
    "messages": {
      "total_messages": 23456,
      "correlated_messages": 23449,
      "buffered_messages": 7,
      "expired_messages": 0,
      "correlation_rate_percent": 99.97,
      "avg_correlation_time_ms": 5,
      "active_subscriptions": 23
    },
    "storage": {
      "reads_per_second": 245,
      "writes_per_second": 67,
      "avg_read_latency_ms": 2.1,
      "avg_write_latency_ms": 8.5,
      "cache_hit_ratio_percent": 94.2,
      "database_size_mb": 1024,
      "compaction_count": 12,
      "last_backup_age_hours": 4
    },
    "network": {
      "grpc_requests_per_second": 89,
      "rest_requests_per_second": 156,
      "active_connections": 45,
      "bytes_sent_per_second": 125432,
      "bytes_received_per_second": 87654,
      "avg_request_duration_ms": 25
    },
    "errors": {
      "total_errors": 23,
      "errors_per_minute": 0.1,
      "error_rate_percent": 0.01,
      "by_type": {
        "validation_error": 15,
        "timeout_error": 5,
        "connection_error": 3
      }
    }
  },
  "request_id": "req_1641998400700"
}
```

## Использование в мониторинге

### Prometheus Integration
```go
// Экспорт метрик в Prometheus формате
func exportPrometheusMetrics(metrics SystemMetrics) string {
    return fmt.Sprintf(`
# HELP atom_processes_total Total number of processes
# TYPE atom_processes_total counter
atom_processes_total %d

# HELP atom_processes_active Currently active processes  
# TYPE atom_processes_active gauge
atom_processes_active %d

# HELP atom_cpu_usage_percent CPU usage percentage
# TYPE atom_cpu_usage_percent gauge
atom_cpu_usage_percent %f
`, metrics.ProcessEngine.TotalProcesses, 
   metrics.ProcessEngine.ActiveProcesses,
   metrics.System.CPUUsagePercent)
}
```

### Grafana Dashboard Query
```promql
# Процессы в минуту
rate(atom_processes_total[5m]) * 60

# Использование памяти
atom_memory_usage_mb / atom_memory_total_mb * 100

# Latency percentiles
histogram_quantile(0.95, rate(atom_request_duration_ms_bucket[5m]))
```

## Связанные endpoints
- [`GET /api/v1/system/status`](./system-status.md) - Статус системы
- [`GET /api/v1/system/info`](./system-info.md) - Системная информация
