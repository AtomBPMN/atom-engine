# GET /api/v1/expressions/stats

## Описание
Получение статистики и метрик использования Expression Engine. Включает информацию о производительности, кэше компиляций и общем использовании.

## URL
```
GET /api/v1/expressions/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression`

## Параметры запроса (Query Parameters)

### Фильтрация данных
- `period` (string): Период статистики (`hour`, `day`, `week`, `month`) (по умолчанию: `day`)
- `include_cache` (boolean): Включить статистику кэша (по умолчанию: true)
- `include_performance` (boolean): Включить метрики производительности (по умолчанию: true)

## Примеры запросов

### Общая статистика
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за последний час
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/stats?period=hour" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/expressions/stats?period=week&include_performance=true', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const stats = await response.json();
console.log('Expression stats:', stats.data);
```

## Ответы

### 200 OK - Статистика Expression Engine
```json
{
  "success": true,
  "data": {
    "period": "day",
    "timestamp": "2025-01-11T10:30:00Z",
    "engine_info": {
      "version": "1.1.0",
      "uptime_seconds": 86400,
      "memory_usage_mb": 128.5,
      "cpu_usage_percent": 15.2
    },
    "execution_stats": {
      "total_evaluations": 15420,
      "successful_evaluations": 15398,
      "failed_evaluations": 22,
      "average_execution_time_ms": 12.5,
      "p95_execution_time_ms": 45.2,
      "p99_execution_time_ms": 120.8,
      "evaluations_per_second": 178.5
    },
    "compilation_stats": {
      "total_compilations": 1250,
      "successful_compilations": 1248,
      "failed_compilations": 2,
      "cache_hits": 14170,
      "cache_misses": 1250,
      "cache_hit_rate": 91.9,
      "average_compilation_time_ms": 85.3,
      "total_compiled_expressions": 892
    },
    "expression_complexity": {
      "simple_expressions": {
        "count": 8500,
        "percentage": 55.1,
        "average_time_ms": 5.2
      },
      "medium_expressions": {
        "count": 5200,
        "percentage": 33.7,
        "average_time_ms": 18.7
      },
      "complex_expressions": {
        "count": 1720,
        "percentage": 11.2,
        "average_time_ms": 42.1
      }
    },
    "function_usage": {
      "most_used": [
        {"name": "sum", "count": 3420, "percentage": 22.2},
        {"name": "if", "count": 2890, "percentage": 18.7},
        {"name": "count", "count": 1650, "percentage": 10.7},
        {"name": "mean", "count": 980, "percentage": 6.4},
        {"name": "upper", "count": 750, "percentage": 4.9}
      ],
      "total_function_calls": 15420,
      "unique_functions_used": 45
    },
    "error_analysis": {
      "syntax_errors": 12,
      "runtime_errors": 8,
      "type_errors": 2,
      "most_common_errors": [
        {
          "type": "UNDEFINED_VARIABLE",
          "count": 8,
          "percentage": 36.4
        },
        {
          "type": "TYPE_MISMATCH", 
          "count": 6,
          "percentage": 27.3
        }
      ]
    },
    "performance_metrics": {
      "optimization_impact": {
        "expressions_optimized": 1248,
        "average_speedup": 2.8,
        "total_time_saved_ms": 125400
      },
      "memory_efficiency": {
        "peak_memory_mb": 156.8,
        "average_memory_mb": 128.5,
        "memory_saved_by_compilation": "22.3%"
      }
    }
  },
  "request_id": "req_1641998404300"
}
```

## Использование

### Stats Dashboard
```javascript
class ExpressionStatsDashboard {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.refreshInterval = null;
  }
  
  async loadStats(period = 'day') {
    const response = await fetch(`/api/v1/expressions/stats?period=${period}`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  startAutoRefresh(interval = 30000) {
    this.refreshInterval = setInterval(async () => {
      const stats = await this.loadStats();
      this.updateDashboard(stats.data);
    }, interval);
  }
  
  stopAutoRefresh() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
      this.refreshInterval = null;
    }
  }
  
  updateDashboard(stats) {
    // Update dashboard elements
    this.updateExecutionMetrics(stats.execution_stats);
    this.updateCompilationMetrics(stats.compilation_stats);
    this.updateErrorAnalysis(stats.error_analysis);
  }
  
  updateExecutionMetrics(execStats) {
    document.getElementById('total-evaluations').textContent = execStats.total_evaluations.toLocaleString();
    document.getElementById('success-rate').textContent = 
      ((execStats.successful_evaluations / execStats.total_evaluations) * 100).toFixed(2) + '%';
    document.getElementById('avg-execution-time').textContent = execStats.average_execution_time_ms + 'ms';
    document.getElementById('evaluations-per-second').textContent = execStats.evaluations_per_second.toFixed(1);
  }
  
  updateCompilationMetrics(compStats) {
    document.getElementById('cache-hit-rate').textContent = compStats.cache_hit_rate.toFixed(1) + '%';
    document.getElementById('total-compilations').textContent = compStats.total_compilations.toLocaleString();
    document.getElementById('compiled-expressions').textContent = compStats.total_compiled_expressions.toLocaleString();
  }
  
  updateErrorAnalysis(errorStats) {
    const errorList = document.getElementById('common-errors');
    errorList.innerHTML = '';
    
    errorStats.most_common_errors.forEach(error => {
      const errorItem = document.createElement('li');
      errorItem.innerHTML = `${error.type}: ${error.count} (${error.percentage.toFixed(1)}%)`;
      errorList.appendChild(errorItem);
    });
  }
}

// Использование
const dashboard = new ExpressionStatsDashboard('your-api-key');

// Загрузка статистики
const stats = await dashboard.loadStats('day');
console.log('Expression statistics:', stats);

// Автообновление каждые 30 секунд
dashboard.startAutoRefresh(30000);
```

### Performance Monitor
```javascript
class ExpressionPerformanceMonitor {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.alerts = [];
    this.thresholds = {
      execution_time_ms: 100,
      error_rate_percent: 5,
      cache_hit_rate_percent: 80
    };
  }
  
  async checkPerformance() {
    const stats = await this.loadStats();
    const alerts = this.analyzePerformance(stats.data);
    
    if (alerts.length > 0) {
      this.handleAlerts(alerts);
    }
    
    return {
      status: alerts.length === 0 ? 'healthy' : 'warning',
      alerts,
      metrics: this.extractKeyMetrics(stats.data)
    };
  }
  
  analyzePerformance(stats) {
    const alerts = [];
    
    // Check execution time
    if (stats.execution_stats.average_execution_time_ms > this.thresholds.execution_time_ms) {
      alerts.push({
        type: 'SLOW_EXECUTION',
        severity: 'WARNING',
        message: `Average execution time (${stats.execution_stats.average_execution_time_ms}ms) exceeds threshold (${this.thresholds.execution_time_ms}ms)`,
        suggestion: 'Consider optimizing complex expressions or increasing compilation usage'
      });
    }
    
    // Check error rate
    const errorRate = (stats.execution_stats.failed_evaluations / stats.execution_stats.total_evaluations) * 100;
    if (errorRate > this.thresholds.error_rate_percent) {
      alerts.push({
        type: 'HIGH_ERROR_RATE',
        severity: 'CRITICAL',
        message: `Error rate (${errorRate.toFixed(2)}%) exceeds threshold (${this.thresholds.error_rate_percent}%)`,
        suggestion: 'Review and fix common expression errors'
      });
    }
    
    // Check cache performance
    if (stats.compilation_stats.cache_hit_rate < this.thresholds.cache_hit_rate_percent) {
      alerts.push({
        type: 'LOW_CACHE_HIT_RATE',
        severity: 'WARNING',
        message: `Cache hit rate (${stats.compilation_stats.cache_hit_rate}%) is below threshold (${this.thresholds.cache_hit_rate_percent}%)`,
        suggestion: 'Increase cache size or review expression patterns'
      });
    }
    
    return alerts;
  }
  
  async loadStats() {
    const response = await fetch('/api/v1/expressions/stats', {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  extractKeyMetrics(stats) {
    return {
      evaluations_per_second: stats.execution_stats.evaluations_per_second,
      average_execution_time: stats.execution_stats.average_execution_time_ms,
      error_rate: (stats.execution_stats.failed_evaluations / stats.execution_stats.total_evaluations) * 100,
      cache_hit_rate: stats.compilation_stats.cache_hit_rate,
      memory_usage: stats.engine_info.memory_usage_mb,
      cpu_usage: stats.engine_info.cpu_usage_percent
    };
  }
  
  handleAlerts(alerts) {
    alerts.forEach(alert => {
      console.warn(`[${alert.severity}] ${alert.type}: ${alert.message}`);
      
      if (alert.severity === 'CRITICAL') {
        // Send critical alerts to monitoring system
        this.sendCriticalAlert(alert);
      }
    });
  }
  
  sendCriticalAlert(alert) {
    // Implement integration with monitoring system
    console.error('CRITICAL ALERT:', alert);
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Влияет на статистику выполнения
- [`POST /api/v1/expressions/compile`](./compile-expression.md) - Влияет на статистику компиляции
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Влияет на общую статистику
