# GET /api/v1/bpmn/stats

## Описание
Получение общей статистики по всем BPMN процессам в системе: количество процессов, элементов, статистика парсинга и производительности.

## URL
```
GET /api/v1/bpmn/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры запроса (Query Parameters)
- `tenant_id` (string): Статистика для конкретного тенанта
- `period` (string): Период для временной статистики (`24h`, `7d`, `30d`, `all`)

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за неделю
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/stats?period=7d" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика по тенанту
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/stats?tenant_id=production" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/bpmn/stats?period=30d', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const bpmnStats = await response.json();
console.log('BPMN Statistics:', bpmnStats.data);
```

## Ответы

### 200 OK - Статистика получена
```json
{
  "success": true,
  "data": {
    "overview": {
      "total_processes": 47,
      "active_processes": 45,
      "inactive_processes": 2,
      "total_versions": 89,
      "avg_versions_per_process": 1.89,
      "newest_process_age_days": 2,
      "oldest_process_age_days": 186
    },
    "process_complexity": {
      "total_elements": 1045,
      "avg_elements_per_process": 22.3,
      "max_elements_in_process": 87,
      "min_elements_in_process": 3,
      "complexity_distribution": {
        "simple": 15,
        "medium": 23,
        "complex": 7,
        "very_complex": 2
      }
    },
    "element_statistics": {
      "total_elements": 1045,
      "elements_by_type": {
        "start_events": 47,
        "end_events": 52,
        "service_tasks": 312,
        "user_tasks": 145,
        "script_tasks": 23,
        "send_tasks": 18,
        "receive_tasks": 12,
        "call_activities": 8,
        "exclusive_gateways": 167,
        "parallel_gateways": 89,
        "inclusive_gateways": 34,
        "event_based_gateways": 12,
        "boundary_events": 67,
        "intermediate_events": 45,
        "sequence_flows": 524,
        "message_flows": 8
      },
      "most_used_elements": [
        {
          "type": "sequence_flows",
          "count": 524,
          "percentage": 50.1
        },
        {
          "type": "service_tasks", 
          "count": 312,
          "percentage": 29.9
        },
        {
          "type": "exclusive_gateways",
          "count": 167,
          "percentage": 16.0
        }
      ]
    },
    "parsing_statistics": {
      "total_parsing_operations": 156,
      "successful_parses": 149,
      "failed_parses": 7,
      "success_rate_percent": 95.5,
      "avg_parsing_time_ms": 247,
      "max_parsing_time_ms": 2456,
      "min_parsing_time_ms": 89,
      "total_parsing_time_seconds": 38.5,
      "parsing_errors": {
        "validation_errors": 4,
        "xml_syntax_errors": 2,
        "schema_violations": 1
      }
    },
    "usage_patterns": {
      "most_deployed_processes": [
        {
          "process_id": "order-processing",
          "deployments": 12,
          "last_deployment": "2025-01-11T10:30:00.000Z"
        },
        {
          "process_id": "user-registration",
          "deployments": 8,
          "last_deployment": "2025-01-10T15:45:00.000Z"
        },
        {
          "process_id": "payment-workflow",
          "deployments": 6,
          "last_deployment": "2025-01-09T09:20:00.000Z"
        }
      ],
      "most_executed_processes": [
        {
          "process_id": "order-processing",
          "total_executions": 15847,
          "success_rate": 98.2
        },
        {
          "process_id": "user-registration", 
          "total_executions": 8934,
          "success_rate": 99.1
        },
        {
          "process_id": "inventory-check",
          "total_executions": 6745,
          "success_rate": 97.8
        }
      ],
      "unused_processes": [
        {
          "process_id": "legacy-workflow",
          "last_execution": "2024-11-15T10:00:00.000Z",
          "days_since_execution": 57
        },
        {
          "process_id": "test-process-v1",
          "last_execution": null,
          "days_since_execution": null
        }
      ]
    },
    "performance_metrics": {
      "total_process_instances": 45678,
      "active_process_instances": 234,
      "avg_process_duration_seconds": 42,
      "p50_process_duration_seconds": 35,
      "p95_process_duration_seconds": 89,
      "p99_process_duration_seconds": 156,
      "fastest_process": {
        "process_id": "simple-approval",
        "avg_duration_seconds": 5
      },
      "slowest_process": {
        "process_id": "complex-workflow",
        "avg_duration_seconds": 245
      }
    },
    "file_statistics": {
      "total_file_size_bytes": 2456789,
      "avg_file_size_bytes": 52270,
      "largest_file": {
        "process_id": "mega-workflow",
        "size_bytes": 156789,
        "size_mb": 0.15
      },
      "smallest_file": {
        "process_id": "simple-task",
        "size_bytes": 2345,
        "size_kb": 2.3
      },
      "compression_ratio": 0.23,
      "storage_efficiency": "77% compressed"
    },
    "trends": {
      "deployments_last_24h": 5,
      "deployments_last_7d": 23,
      "deployments_last_30d": 67,
      "execution_trend": {
        "today": 1247,
        "yesterday": 1189,
        "change_percent": 4.9
      },
      "complexity_trend": {
        "avg_elements_last_month": 22.3,
        "avg_elements_previous_month": 19.8,
        "change_percent": 12.6
      }
    },
    "health_indicators": {
      "parsing_health": "GOOD",
      "execution_health": "EXCELLENT", 
      "storage_health": "GOOD",
      "overall_health": "GOOD",
      "issues": [
        {
          "type": "WARNING",
          "message": "7 unused processes detected",
          "recommendation": "Consider archiving unused processes"
        },
        {
          "type": "INFO",
          "message": "Parsing success rate below 100%",
          "recommendation": "Review failed parsing operations"
        }
      ]
    },
    "generated_at": "2025-01-11T10:30:00.000Z",
    "period": "all",
    "tenant_id": null
  },
  "request_id": "req_1641998401900"
}
```

## Поля ответа

### Overview Statistics
- `total_processes` (integer): Общее количество процессов
- `active_processes` (integer): Активные процессы
- `inactive_processes` (integer): Неактивные процессы
- `total_versions` (integer): Общее количество версий
- `avg_versions_per_process` (float): Среднее количество версий на процесс

### Complexity Analysis
- `total_elements` (integer): Общее количество элементов
- `avg_elements_per_process` (float): Среднее количество элементов на процесс
- `complexity_distribution` (object): Распределение по сложности
  - `simple` (< 10 элементов)
  - `medium` (10-30 элементов)
  - `complex` (30-50 элементов)
  - `very_complex` (> 50 элементов)

### Element Statistics
- `elements_by_type` (object): Количество элементов по типам
- `most_used_elements` (array): Самые используемые типы элементов

### Parsing Statistics
- `successful_parses` (integer): Успешные парсинги
- `failed_parses` (integer): Неудачные парсинги
- `success_rate_percent` (float): Процент успешности
- `avg_parsing_time_ms` (integer): Среднее время парсинга

### Usage Patterns
- `most_deployed_processes` (array): Самые развертываемые процессы
- `most_executed_processes` (array): Самые выполняемые процессы
- `unused_processes` (array): Неиспользуемые процессы

### Performance Metrics
- `total_process_instances` (integer): Общее количество экземпляров
- `avg_process_duration_seconds` (integer): Средняя длительность
- Percentiles: p50, p95, p99
- `fastest_process` и `slowest_process`: Самые быстрые/медленные процессы

## Использование

### Dashboard Metrics
```javascript
async function getBpmnDashboardData() {
  const response = await fetch('/api/v1/bpmn/stats?period=7d');
  const stats = await response.json();
  
  const data = stats.data;
  
  return {
    // KPI карточки
    totalProcesses: data.overview.total_processes,
    activeProcesses: data.overview.active_processes,
    successRate: data.parsing_statistics.success_rate_percent,
    avgComplexity: data.process_complexity.avg_elements_per_process,
    
    // Графики
    complexityChart: data.process_complexity.complexity_distribution,
    elementTypesChart: data.element_statistics.elements_by_type,
    trendsChart: data.trends,
    
    // Таблицы
    topProcesses: data.usage_patterns.most_executed_processes,
    unusedProcesses: data.usage_patterns.unused_processes,
    
    // Алерты
    healthIssues: data.health_indicators.issues
  };
}
```

### Health Monitoring
```javascript
async function monitorBpmnHealth() {
  const response = await fetch('/api/v1/bpmn/stats');
  const stats = await response.json();
  
  const health = stats.data.health_indicators;
  
  // Проверка критических показателей
  const issues = [];
  
  if (stats.data.parsing_statistics.success_rate_percent < 95) {
    issues.push({
      severity: 'HIGH',
      message: `Parsing success rate is ${stats.data.parsing_statistics.success_rate_percent}%`,
      recommendation: 'Review failed parsing operations'
    });
  }
  
  if (stats.data.usage_patterns.unused_processes.length > 10) {
    issues.push({
      severity: 'MEDIUM',
      message: `${stats.data.usage_patterns.unused_processes.length} unused processes`,
      recommendation: 'Consider archiving unused processes'
    });
  }
  
  if (stats.data.process_complexity.avg_elements_per_process > 50) {
    issues.push({
      severity: 'LOW',
      message: 'High average process complexity',
      recommendation: 'Consider process simplification'
    });
  }
  
  return {
    overall_health: health.overall_health,
    issues: [...health.issues, ...issues]
  };
}
```

### Capacity Planning
```javascript
async function analyzeCapacityTrends() {
  const [current, lastMonth] = await Promise.all([
    fetch('/api/v1/bpmn/stats?period=30d'),
    fetch('/api/v1/bpmn/stats?period=60d')
  ]);
  
  const currentStats = await current.json();
  const lastMonthStats = await lastMonth.json();
  
  const currentData = currentStats.data;
  const lastMonthData = lastMonthStats.data;
  
  const growth = {
    processes: {
      current: currentData.overview.total_processes,
      previous: lastMonthData.overview.total_processes,
      growth_rate: ((currentData.overview.total_processes - lastMonthData.overview.total_processes) / lastMonthData.overview.total_processes) * 100
    },
    executions: {
      current: currentData.performance_metrics.total_process_instances,
      growth_trend: currentData.trends.execution_trend.change_percent
    },
    storage: {
      current_mb: currentData.file_statistics.total_file_size_bytes / (1024 * 1024),
      avg_file_size_kb: currentData.file_statistics.avg_file_size_bytes / 1024
    }
  };
  
  // Прогноз на следующий месяц
  const forecast = {
    estimated_processes: Math.round(growth.processes.current * (1 + growth.processes.growth_rate / 100)),
    estimated_storage_mb: Math.round(growth.storage.current_mb * 1.2), // Учитываем рост
    capacity_warning: growth.processes.growth_rate > 20 // Рост более 20%
  };
  
  return { growth, forecast };
}
```

### Quality Analysis
```javascript
async function analyzeProcessQuality() {
  const response = await fetch('/api/v1/bpmn/stats');
  const stats = await response.json();
  
  const data = stats.data;
  
  const qualityMetrics = {
    // Показатели качества
    parsing_quality: data.parsing_statistics.success_rate_percent,
    execution_quality: data.performance_metrics.total_process_instances > 0 ? 
      (data.performance_metrics.total_process_instances - data.performance_metrics.active_process_instances) / 
      data.performance_metrics.total_process_instances * 100 : 0,
    
    // Показатели сложности
    complexity_score: calculateComplexityScore(data.process_complexity),
    
    // Показатели использования
    utilization_score: calculateUtilizationScore(data.usage_patterns),
    
    // Общий балл качества (0-100)
    overall_score: 0
  };
  
  qualityMetrics.overall_score = (
    qualityMetrics.parsing_quality * 0.3 +
    qualityMetrics.execution_quality * 0.4 +
    qualityMetrics.complexity_score * 0.2 +
    qualityMetrics.utilization_score * 0.1
  );
  
  return qualityMetrics;
}

function calculateComplexityScore(complexity) {
  const dist = complexity.complexity_distribution;
  const total = dist.simple + dist.medium + dist.complex + dist.very_complex;
  
  // Чем больше простых процессов, тем лучше
  return (dist.simple * 100 + dist.medium * 70 + dist.complex * 40 + dist.very_complex * 10) / total;
}

function calculateUtilizationScore(usage) {
  const total = usage.most_executed_processes.length + usage.unused_processes.length;
  const used = usage.most_executed_processes.length;
  
  return (used / total) * 100;
}
```

### Report Generation
```javascript
async function generateBpmnReport() {
  const response = await fetch('/api/v1/bpmn/stats?period=30d');
  const stats = await response.json();
  
  const data = stats.data;
  
  const report = `
# BPMN Statistics Report
Generated: ${new Date().toISOString()}

## Summary
- **Total Processes**: ${data.overview.total_processes}
- **Active Processes**: ${data.overview.active_processes}
- **Success Rate**: ${data.parsing_statistics.success_rate_percent}%
- **Average Complexity**: ${data.process_complexity.avg_elements_per_process} elements

## Top Performing Processes
${data.usage_patterns.most_executed_processes.map(p => 
  `- ${p.process_id}: ${p.total_executions} executions (${p.success_rate}% success)`
).join('\n')}

## Issues & Recommendations
${data.health_indicators.issues.map(issue => 
  `- **${issue.type}**: ${issue.message}\n  *Recommendation*: ${issue.recommendation}`
).join('\n\n')}

## Performance Trends
- **Executions Today**: ${data.trends.execution_trend.today}
- **Change from Yesterday**: ${data.trends.execution_trend.change_percent > 0 ? '+' : ''}${data.trends.execution_trend.change_percent}%
- **Deployments Last 7 Days**: ${data.trends.deployments_last_7d}
  `;
  
  return report;
}
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes`](./list-processes.md) - Детальный список процессов
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
- [`GET /api/v1/processes/stats`](../processes/get-process-stats.md) - Статистика выполнения процессов
