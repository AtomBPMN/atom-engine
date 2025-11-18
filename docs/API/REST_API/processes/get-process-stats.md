# GET /api/v1/processes/stats

## Описание
Получение общей статистики по всем экземплярам процессов: производительность, использование ресурсов, распределение по статусам и тенденции.

## URL
```
GET /api/v1/processes/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры запроса (Query Parameters)
- `period` (string): Период для статистики (`1h`, `24h`, `7d`, `30d`, `all`)
- `process_id` (string): Статистика для конкретного процесса
- `tenant_id` (string): Статистика для конкретного тенанта
- `group_by` (string): Группировка данных (`process_id`, `status`, `tenant_id`)

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/processes/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за неделю
```bash
curl -X GET "http://localhost:27555/api/v1/processes/stats?period=7d" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика конкретного процесса
```bash
curl -X GET "http://localhost:27555/api/v1/processes/stats?process_id=order-processing" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/stats?period=24h&group_by=process_id', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processStats = await response.json();
```

## Ответы

### 200 OK - Статистика получена
```json
{
  "success": true,
  "data": {
    "period": "all",
    "generated_at": "2025-01-11T10:32:15.456Z",
    "overview": {
      "total_instances": 45678,
      "active_instances": 234,
      "completed_instances": 44987,
      "cancelled_instances": 457,
      "failed_instances": 0,
      "success_rate_percent": 98.0,
      "completion_rate_percent": 98.5
    },
    "status_distribution": {
      "ACTIVE": {
        "count": 234,
        "percentage": 0.51
      },
      "COMPLETED": {
        "count": 44987,
        "percentage": 98.49
      },
      "CANCELLED": {
        "count": 457,
        "percentage": 1.00
      }
    },
    "performance_metrics": {
      "avg_duration_seconds": 42,
      "median_duration_seconds": 35,
      "p95_duration_seconds": 89,
      "p99_duration_seconds": 156,
      "fastest_completion_seconds": 5,
      "slowest_completion_seconds": 3600,
      "throughput": {
        "instances_per_hour": 125,
        "instances_per_day": 3000,
        "peak_throughput_per_hour": 245
      }
    },
    "process_breakdown": [
      {
        "process_id": "order-processing",
        "instances": 15847,
        "success_rate": 98.2,
        "avg_duration_seconds": 45,
        "active_instances": 67,
        "latest_version": 3
      },
      {
        "process_id": "user-registration",
        "instances": 8934,
        "success_rate": 99.1,
        "avg_duration_seconds": 15,
        "active_instances": 23,
        "latest_version": 2
      },
      {
        "process_id": "payment-workflow",
        "instances": 6745,
        "success_rate": 97.8,
        "avg_duration_seconds": 89,
        "active_instances": 45,
        "latest_version": 4
      },
      {
        "process_id": "inventory-check",
        "instances": 4567,
        "success_rate": 99.8,
        "avg_duration_seconds": 8,
        "active_instances": 12,
        "latest_version": 1
      }
    ],
    "temporal_trends": {
      "hourly_distribution": {
        "00": 45, "01": 23, "02": 15, "03": 12,
        "04": 18, "05": 34, "06": 67, "07": 89,
        "08": 145, "09": 189, "10": 234, "11": 212,
        "12": 198, "13": 187, "14": 201, "15": 178,
        "16": 156, "17": 134, "18": 98, "19": 76,
        "20": 65, "21": 54, "22": 43, "23": 38
      },
      "daily_trend": {
        "2025-01-05": 2890,
        "2025-01-06": 3124,
        "2025-01-07": 2987,
        "2025-01-08": 3456,
        "2025-01-09": 3234,
        "2025-01-10": 3567,
        "2025-01-11": 1247
      },
      "growth_rate": {
        "daily_percent": 2.3,
        "weekly_percent": 8.7,
        "monthly_percent": 15.2
      }
    },
    "resource_utilization": {
      "avg_memory_per_instance_mb": 4.5,
      "total_memory_usage_gb": 0.2,
      "avg_cpu_time_per_instance_ms": 1250,
      "total_cpu_time_hours": 15.8,
      "database_operations": {
        "total_reads": 2456789,
        "total_writes": 567890,
        "avg_latency_ms": 12.5
      },
      "storage_usage": {
        "total_size_mb": 1024,
        "avg_size_per_instance_kb": 23
      }
    },
    "activity_statistics": {
      "total_activities_executed": 567890,
      "avg_activities_per_instance": 12.4,
      "most_used_activity_types": {
        "serviceTask": 345678,
        "exclusiveGateway": 123456,
        "userTask": 87654,
        "parallelGateway": 45678
      },
      "activity_performance": {
        "avg_activity_duration_ms": 3400,
        "slowest_activity_type": "userTask",
        "fastest_activity_type": "exclusiveGateway"
      }
    },
    "error_analysis": {
      "incident_rate_percent": 2.1,
      "total_incidents": 956,
      "resolved_incidents": 889,
      "open_incidents": 67,
      "incident_types": {
        "JOB_FAILURE": 567,
        "TIMEOUT": 234,
        "EXPRESSION_ERROR": 123,
        "SYSTEM_ERROR": 32
      },
      "mttr_hours": 2.3,
      "mtbf_hours": 168.5
    },
    "worker_statistics": {
      "total_workers_engaged": 45,
      "avg_workers_per_instance": 2.3,
      "worker_utilization_percent": 67.8,
      "top_worker_types": [
        {
          "type": "payment-processor",
          "jobs_processed": 15847,
          "avg_processing_time_ms": 5000
        },
        {
          "type": "email-sender",
          "jobs_processed": 12456,
          "avg_processing_time_ms": 2000
        },
        {
          "type": "inventory-checker",
          "jobs_processed": 8934,
          "avg_processing_time_ms": 1500
        }
      ]
    },
    "capacity_analysis": {
      "current_load_percent": 23.4,
      "peak_load_percent": 89.7,
      "estimated_max_capacity": 1000,
      "bottleneck_indicators": [
        {
          "component": "worker_pool",
          "utilization_percent": 67.8,
          "recommendation": "Consider scaling payment-processor workers"
        }
      ]
    },
    "quality_metrics": {
      "process_quality_score": 87.5,
      "factors": {
        "success_rate_weight": 40,
        "performance_weight": 30,
        "reliability_weight": 20,
        "efficiency_weight": 10
      },
      "improvement_areas": [
        "Reduce average duration for payment-workflow",
        "Improve incident resolution time",
        "Optimize worker allocation"
      ]
    }
  },
  "request_id": "req_1641998402400"
}
```

## Поля ответа

### Overview Statistics
- Общие счетчики экземпляров процессов
- Показатели успешности и завершенности

### Status Distribution
- Распределение экземпляров по статусам
- Процентное соотношение

### Performance Metrics
- Временные характеристики выполнения
- Throughput метрики
- Percentiles производительности

### Process Breakdown
- Статистика по отдельным процессам
- Сравнительный анализ производительности

### Temporal Trends
- Почасовое и ежедневное распределение
- Темпы роста

### Resource Utilization
- Использование памяти, CPU, БД
- Эффективность использования ресурсов

### Activity Statistics
- Статистика по типам активностей
- Производительность различных элементов

### Error Analysis
- Анализ инцидентов и ошибок
- MTTR (Mean Time To Repair) и MTBF (Mean Time Between Failures)

## Использование

### Dashboard Metrics
```javascript
async function getProcessDashboardData() {
  const response = await fetch('/api/v1/processes/stats?period=24h');
  const stats = await response.json();
  
  const data = stats.data;
  
  return {
    // KPI Cards
    totalInstances: data.overview.total_instances,
    activeInstances: data.overview.active_instances,
    successRate: data.overview.success_rate_percent,
    avgDuration: data.performance_metrics.avg_duration_seconds,
    
    // Charts
    statusChart: data.status_distribution,
    throughputChart: data.temporal_trends.hourly_distribution,
    processBreakdownChart: data.process_breakdown,
    
    // Performance indicators
    p95Duration: data.performance_metrics.p95_duration_seconds,
    incidentRate: data.error_analysis.incident_rate_percent,
    workerUtilization: data.worker_statistics.worker_utilization_percent,
    
    // Alerts
    bottlenecks: data.capacity_analysis.bottleneck_indicators,
    improvementAreas: data.quality_metrics.improvement_areas
  };
}
```

### Performance Analysis
```javascript
async function analyzeProcessPerformance() {
  const response = await fetch('/api/v1/processes/stats?period=7d&group_by=process_id');
  const stats = await response.json();
  
  const processes = stats.data.process_breakdown;
  
  const analysis = {
    fastestProcesses: processes
      .sort((a, b) => a.avg_duration_seconds - b.avg_duration_seconds)
      .slice(0, 5),
      
    slowestProcesses: processes
      .sort((a, b) => b.avg_duration_seconds - a.avg_duration_seconds)
      .slice(0, 5),
      
    mostReliableProcesses: processes
      .sort((a, b) => b.success_rate - a.success_rate)
      .slice(0, 5),
      
    performanceIssues: processes.filter(p => 
      p.success_rate < 95 || p.avg_duration_seconds > 120
    ),
    
    capacityRecommendations: generateCapacityRecommendations(processes)
  };
  
  return analysis;
}

function generateCapacityRecommendations(processes) {
  const recommendations = [];
  
  processes.forEach(process => {
    if (process.active_instances > 100) {
      recommendations.push({
        type: 'SCALE_UP',
        process_id: process.process_id,
        message: `High active instance count: ${process.active_instances}`,
        suggestion: 'Consider adding more workers'
      });
    }
    
    if (process.success_rate < 95) {
      recommendations.push({
        type: 'RELIABILITY',
        process_id: process.process_id,
        message: `Low success rate: ${process.success_rate}%`,
        suggestion: 'Review error patterns and improve error handling'
      });
    }
  });
  
  return recommendations;
}
```

### Trend Analysis
```javascript
async function analyzeTrends() {
  const [current, previous] = await Promise.all([
    fetch('/api/v1/processes/stats?period=7d'),
    fetch('/api/v1/processes/stats?period=14d')
  ]);
  
  const currentStats = await current.json();
  const previousStats = await previous.json();
  
  const currentData = currentStats.data;
  const previousData = previousStats.data;
  
  const trends = {
    volume: {
      current: currentData.overview.total_instances,
      previous: previousData.overview.total_instances,
      change_percent: ((currentData.overview.total_instances - previousData.overview.total_instances) / previousData.overview.total_instances) * 100
    },
    
    performance: {
      current_avg_duration: currentData.performance_metrics.avg_duration_seconds,
      previous_avg_duration: previousData.performance_metrics.avg_duration_seconds,
      performance_change_percent: ((previousData.performance_metrics.avg_duration_seconds - currentData.performance_metrics.avg_duration_seconds) / previousData.performance_metrics.avg_duration_seconds) * 100
    },
    
    quality: {
      current_success_rate: currentData.overview.success_rate_percent,
      previous_success_rate: previousData.overview.success_rate_percent,
      quality_change: currentData.overview.success_rate_percent - previousData.overview.success_rate_percent
    }
  };
  
  return trends;
}
```

### Capacity Planning
```javascript
async function performCapacityPlanning() {
  const response = await fetch('/api/v1/processes/stats?period=30d');
  const stats = await response.json();
  
  const data = stats.data;
  
  // Анализ роста
  const growthRate = data.temporal_trends.growth_rate.monthly_percent / 100;
  const currentThroughput = data.performance_metrics.throughput.instances_per_day;
  
  // Прогноз на следующие 3 месяца
  const forecast = {
    month1: Math.round(currentThroughput * (1 + growthRate)),
    month2: Math.round(currentThroughput * Math.pow(1 + growthRate, 2)),
    month3: Math.round(currentThroughput * Math.pow(1 + growthRate, 3))
  };
  
  // Анализ текущей нагрузки
  const currentLoad = data.capacity_analysis.current_load_percent;
  const peakLoad = data.capacity_analysis.peak_load_percent;
  
  const recommendations = [];
  
  if (peakLoad > 80) {
    recommendations.push({
      type: 'IMMEDIATE',
      priority: 'HIGH',
      message: 'Peak load exceeds 80%',
      action: 'Scale up worker pools immediately'
    });
  }
  
  if (forecast.month3 > currentThroughput * 2) {
    recommendations.push({
      type: 'PLANNING',
      priority: 'MEDIUM',
      message: `Expected 200% growth in 3 months`,
      action: 'Plan infrastructure scaling'
    });
  }
  
  return {
    currentMetrics: {
      throughput: currentThroughput,
      load: currentLoad,
      peakLoad: peakLoad
    },
    forecast,
    recommendations,
    resourceNeeds: calculateResourceNeeds(forecast, data.resource_utilization)
  };
}

function calculateResourceNeeds(forecast, currentUtilization) {
  const currentDailyThroughput = 3000; // Из данных
  const growthMultiplier = forecast.month3 / currentDailyThroughput;
  
  return {
    memory_gb: (currentUtilization.total_memory_usage_gb * growthMultiplier).toFixed(1),
    cpu_hours: (currentUtilization.total_cpu_time_hours * growthMultiplier).toFixed(1),
    storage_mb: (currentUtilization.storage_usage.total_size_mb * growthMultiplier).toFixed(0),
    workers_needed: Math.ceil(45 * growthMultiplier) // Текущие workers * рост
  };
}
```

### SLA Monitoring
```javascript
async function monitorSLA() {
  const response = await fetch('/api/v1/processes/stats?period=24h');
  const stats = await response.json();
  
  const data = stats.data;
  
  // Определяем SLA пороги
  const SLA_THRESHOLDS = {
    success_rate: 95,      // 95% success rate
    avg_duration: 60,      // 60 seconds average
    p95_duration: 120,     // 120 seconds 95th percentile
    incident_rate: 5       // 5% incident rate
  };
  
  const slaStatus = {
    success_rate: {
      current: data.overview.success_rate_percent,
      threshold: SLA_THRESHOLDS.success_rate,
      status: data.overview.success_rate_percent >= SLA_THRESHOLDS.success_rate ? 'PASS' : 'FAIL'
    },
    
    avg_duration: {
      current: data.performance_metrics.avg_duration_seconds,
      threshold: SLA_THRESHOLDS.avg_duration,
      status: data.performance_metrics.avg_duration_seconds <= SLA_THRESHOLDS.avg_duration ? 'PASS' : 'FAIL'
    },
    
    p95_duration: {
      current: data.performance_metrics.p95_duration_seconds,
      threshold: SLA_THRESHOLDS.p95_duration,
      status: data.performance_metrics.p95_duration_seconds <= SLA_THRESHOLDS.p95_duration ? 'PASS' : 'FAIL'
    },
    
    incident_rate: {
      current: data.error_analysis.incident_rate_percent,
      threshold: SLA_THRESHOLDS.incident_rate,
      status: data.error_analysis.incident_rate_percent <= SLA_THRESHOLDS.incident_rate ? 'PASS' : 'FAIL'
    }
  };
  
  const overallSLA = Object.values(slaStatus).every(metric => metric.status === 'PASS');
  
  return {
    overall_status: overallSLA ? 'PASS' : 'FAIL',
    metrics: slaStatus,
    violations: Object.entries(slaStatus)
      .filter(([key, metric]) => metric.status === 'FAIL')
      .map(([key, metric]) => ({
        metric: key,
        current: metric.current,
        threshold: metric.threshold,
        severity: calculateViolationSeverity(metric.current, metric.threshold, key)
      }))
  };
}

function calculateViolationSeverity(current, threshold, metric) {
  const deviation = Math.abs(current - threshold) / threshold;
  
  if (deviation > 0.2) return 'HIGH';    // Отклонение > 20%
  if (deviation > 0.1) return 'MEDIUM';  // Отклонение > 10%
  return 'LOW';
}
```

## Связанные endpoints
- [`GET /api/v1/processes`](./list-processes.md) - Список экземпляров процессов
- [`GET /api/v1/bpmn/stats`](../bpmn/get-bpmn-stats.md) - Статистика BPMN процессов
- [`GET /api/v1/system/metrics`](../system/system-metrics.md) - Системные метрики
- [`GET /api/v1/jobs/stats`](../jobs/get-job-stats.md) - Статистика заданий
