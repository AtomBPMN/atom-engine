# GET /api/v1/incidents/stats

## Описание
Получение статистики и аналитики по инцидентам системы. Включает метрики производительности, тренды и анализ эффективности решений.

## URL
```
GET /api/v1/incidents/stats
```

## Авторизация
✅ **Требуется API ключ** с разрешением `incident`

## Параметры запроса (Query Parameters)

### Фильтрация временного периода
- `period` (string): Период статистики (`hour`, `day`, `week`, `month`) (по умолчанию: `day`)
- `start_date` (string): Начальная дата (ISO 8601)
- `end_date` (string): Конечная дата (ISO 8601)

### Группировка данных
- `group_by` (string): Группировка статистики (`type`, `severity`, `status`, `process`) (по умолчанию: `type`)
- `include_trends` (boolean): Включить анализ трендов (по умолчанию: true)

## Примеры запросов

### Общая статистика за день
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/stats" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика за неделю с трендами
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/stats?period=week&include_trends=true" \
  -H "X-API-Key: your-api-key-here"
```

### Статистика по типам инцидентов
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/stats?group_by=type&period=month" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/incidents/stats?period=week&group_by=severity', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const stats = await response.json();
```

## Ответы

### 200 OK - Статистика инцидентов
```json
{
  "success": true,
  "data": {
    "period": "day",
    "start_time": "2025-01-10T00:00:00Z",
    "end_time": "2025-01-11T00:00:00Z",
    "summary": {
      "total_incidents": 127,
      "open_incidents": 23,
      "resolved_incidents": 89,
      "dismissed_incidents": 15,
      "resolution_rate": 70.1,
      "average_resolution_time_minutes": 45.3,
      "median_resolution_time_minutes": 22.0
    },
    "by_type": {
      "JOB": {
        "total": 48,
        "open": 12,
        "resolved": 31,
        "dismissed": 5,
        "resolution_rate": 64.6,
        "avg_resolution_time_minutes": 52.1
      },
      "EXPRESSION": {
        "total": 28,
        "open": 4,
        "resolved": 22,
        "dismissed": 2,
        "resolution_rate": 78.6,
        "avg_resolution_time_minutes": 18.4
      },
      "BPMN": {
        "total": 19,
        "open": 3,
        "resolved": 14,
        "dismissed": 2,
        "resolution_rate": 73.7,
        "avg_resolution_time_minutes": 65.2
      },
      "PROCESS": {
        "total": 12,
        "open": 2,
        "resolved": 8,
        "dismissed": 2,
        "resolution_rate": 66.7,
        "avg_resolution_time_minutes": 38.9
      },
      "TIMER": {
        "total": 8,
        "open": 1,
        "resolved": 6,
        "dismissed": 1,
        "resolution_rate": 75.0,
        "avg_resolution_time_minutes": 25.3
      },
      "MESSAGE": {
        "total": 7,
        "open": 1,
        "resolved": 5,
        "dismissed": 1,
        "resolution_rate": 71.4,
        "avg_resolution_time_minutes": 31.7
      },
      "SYSTEM": {
        "total": 5,
        "open": 0,
        "resolved": 3,
        "dismissed": 2,
        "resolution_rate": 60.0,
        "avg_resolution_time_minutes": 95.4
      }
    },
    "by_severity": {
      "LOW": {
        "total": 45,
        "resolved": 38,
        "resolution_rate": 84.4,
        "avg_resolution_time_minutes": 15.2
      },
      "MEDIUM": {
        "total": 52,
        "resolved": 39,
        "resolution_rate": 75.0,
        "avg_resolution_time_minutes": 35.8
      },
      "HIGH": {
        "total": 25,
        "resolved": 10,
        "resolution_rate": 40.0,
        "avg_resolution_time_minutes": 85.3
      },
      "CRITICAL": {
        "total": 5,
        "resolved": 2,
        "resolution_rate": 40.0,
        "avg_resolution_time_minutes": 142.7
      }
    },
    "top_processes_with_incidents": [
      {
        "process_key": "Order_Processing",
        "incidents": 34,
        "percentage": 26.8,
        "most_common_type": "JOB",
        "avg_resolution_time": 48.5
      },
      {
        "process_key": "Invoice_Approval",
        "incidents": 28,
        "percentage": 22.0,
        "most_common_type": "EXPRESSION",
        "avg_resolution_time": 22.1
      },
      {
        "process_key": "Customer_Onboarding",
        "incidents": 19,
        "percentage": 15.0,
        "most_common_type": "BPMN",
        "avg_resolution_time": 67.3
      }
    ],
    "resolution_patterns": {
      "retry_success_rate": 78.5,
      "dismiss_rate": 11.8,
      "escalation_rate": 9.7,
      "auto_resolution_rate": 23.6,
      "most_effective_resolution": "retry",
      "average_attempts_to_resolve": 1.4
    },
    "time_distribution": {
      "peak_hours": [
        {"hour": 9, "incidents": 18},
        {"hour": 14, "incidents": 16},
        {"hour": 16, "incidents": 14}
      ],
      "quiet_hours": [
        {"hour": 2, "incidents": 1},
        {"hour": 5, "incidents": 2},
        {"hour": 23, "incidents": 3}
      ],
      "weekend_vs_weekday": {
        "weekday_average": 15.2,
        "weekend_average": 8.7,
        "weekend_reduction": 42.8
      }
    },
    "trends": {
      "incident_count_trend": {
        "direction": "decreasing",
        "change_percent": -12.3,
        "confidence": 0.87
      },
      "resolution_time_trend": {
        "direction": "improving",
        "change_percent": -8.5,
        "confidence": 0.92
      },
      "severity_trend": {
        "critical_incidents": {
          "direction": "stable",
          "change_percent": 2.1
        },
        "high_incidents": {
          "direction": "decreasing",
          "change_percent": -15.8
        }
      }
    },
    "performance_metrics": {
      "mttr_minutes": 45.3,
      "mttd_minutes": 8.7,
      "availability_impact": 99.2,
      "incident_frequency_per_day": 127,
      "resolution_efficiency": 0.785
    }
  },
  "request_id": "req_1641998404700"
}
```

### 200 OK - Статистика с группировкой по процессам
```json
{
  "success": true,
  "data": {
    "period": "week",
    "group_by": "process",
    "summary": {
      "total_incidents": 486,
      "unique_processes": 12,
      "most_problematic_process": "Order_Processing",
      "least_problematic_process": "Simple_Approval"
    },
    "by_process": {
      "Order_Processing": {
        "total_incidents": 145,
        "resolution_rate": 68.3,
        "avg_resolution_time_minutes": 52.4,
        "incident_distribution": {
          "JOB": 89,
          "EXPRESSION": 32,
          "BPMN": 24
        },
        "problem_areas": [
          {"element_id": "ServiceTask_PaymentProcess", "incidents": 45},
          {"element_id": "ServiceTask_InventoryCheck", "incidents": 28},
          {"element_id": "ExclusiveGateway_PriorityCheck", "incidents": 18}
        ]
      },
      "Invoice_Approval": {
        "total_incidents": 98,
        "resolution_rate": 82.7,
        "avg_resolution_time_minutes": 28.1,
        "incident_distribution": {
          "EXPRESSION": 67,
          "JOB": 21,
          "BPMN": 10
        },
        "problem_areas": [
          {"element_id": "ExclusiveGateway_AmountCheck", "incidents": 34},
          {"element_id": "UserTask_ManagerApproval", "incidents": 22}
        ]
      }
    }
  }
}
```

## Использование

### Incident Analytics Dashboard
```javascript
class IncidentAnalyticsDashboard {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getStats(period = 'day', groupBy = 'type') {
    const response = await fetch(`/api/v1/incidents/stats?period=${period}&group_by=${groupBy}&include_trends=true`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  async generateExecutiveSummary(period = 'week') {
    const stats = await this.getStats(period);
    const data = stats.data;
    
    return {
      headline_metrics: {
        total_incidents: data.summary.total_incidents,
        resolution_rate: `${data.summary.resolution_rate}%`,
        avg_resolution_time: `${Math.round(data.summary.average_resolution_time_minutes)} minutes`,
        trend: data.trends.incident_count_trend.direction
      },
      key_insights: this.generateKeyInsights(data),
      action_items: this.generateActionItems(data),
      risk_assessment: this.assessRisk(data)
    };
  }
  
  generateKeyInsights(data) {
    const insights = [];
    
    // Resolution rate insight
    if (data.summary.resolution_rate > 80) {
      insights.push({
        type: 'positive',
        message: `High resolution rate of ${data.summary.resolution_rate}% indicates effective incident management`
      });
    } else if (data.summary.resolution_rate < 60) {
      insights.push({
        type: 'concern',
        message: `Low resolution rate of ${data.summary.resolution_rate}% needs attention`
      });
    }
    
    // Most problematic type
    const typesByCount = Object.entries(data.by_type)
      .sort(([,a], [,b]) => b.total - a.total);
    
    if (typesByCount.length > 0) {
      const [topType, topStats] = typesByCount[0];
      insights.push({
        type: 'info',
        message: `${topType} incidents are most common (${topStats.total} incidents, ${topStats.resolution_rate}% resolved)`
      });
    }
    
    // Severity distribution insight
    const criticalPercent = (data.by_severity.CRITICAL?.total || 0) / data.summary.total_incidents * 100;
    if (criticalPercent > 5) {
      insights.push({
        type: 'warning',
        message: `High percentage of critical incidents (${criticalPercent.toFixed(1)}%)`
      });
    }
    
    return insights;
  }
  
  generateActionItems(data) {
    const actions = [];
    
    // High resolution time
    if (data.summary.average_resolution_time_minutes > 60) {
      actions.push({
        priority: 'high',
        action: 'Investigate processes with long resolution times',
        target: `Reduce average resolution time from ${Math.round(data.summary.average_resolution_time_minutes)} to 45 minutes`
      });
    }
    
    // Low resolution rate for critical incidents
    const criticalResolutionRate = data.by_severity.CRITICAL?.resolution_rate || 100;
    if (criticalResolutionRate < 80) {
      actions.push({
        priority: 'urgent',
        action: 'Improve critical incident response procedures',
        target: `Increase critical incident resolution rate from ${criticalResolutionRate}% to 90%`
      });
    }
    
    // Process-specific issues
    if (data.top_processes_with_incidents) {
      const topProcess = data.top_processes_with_incidents[0];
      if (topProcess.incidents > data.summary.total_incidents * 0.2) {
        actions.push({
          priority: 'medium',
          action: `Review and optimize ${topProcess.process_key} process`,
          target: `Reduce incidents by 30% through process improvements`
        });
      }
    }
    
    return actions;
  }
  
  assessRisk(data) {
    let riskScore = 0;
    const riskFactors = [];
    
    // Resolution rate risk
    if (data.summary.resolution_rate < 70) {
      riskScore += 30;
      riskFactors.push('Low overall resolution rate');
    }
    
    // Critical incident risk
    const criticalCount = data.by_severity.CRITICAL?.total || 0;
    if (criticalCount > 5) {
      riskScore += 25;
      riskFactors.push('High number of critical incidents');
    }
    
    // Trend risk
    if (data.trends.incident_count_trend.direction === 'increasing') {
      riskScore += 20;
      riskFactors.push('Increasing incident trend');
    }
    
    // Open incident backlog risk
    const openPercent = (data.summary.open_incidents / data.summary.total_incidents) * 100;
    if (openPercent > 20) {
      riskScore += 15;
      riskFactors.push('Large backlog of open incidents');
    }
    
    let riskLevel;
    if (riskScore >= 70) riskLevel = 'HIGH';
    else if (riskScore >= 40) riskLevel = 'MEDIUM';
    else if (riskScore >= 20) riskLevel = 'LOW';
    else riskLevel = 'MINIMAL';
    
    return {
      score: riskScore,
      level: riskLevel,
      factors: riskFactors,
      recommendations: this.getRiskRecommendations(riskLevel, riskFactors)
    };
  }
  
  getRiskRecommendations(riskLevel, factors) {
    const recommendations = [];
    
    if (riskLevel === 'HIGH') {
      recommendations.push('Immediate escalation to management required');
      recommendations.push('Consider activating incident response team');
    }
    
    if (factors.includes('Low overall resolution rate')) {
      recommendations.push('Review incident resolution procedures');
      recommendations.push('Provide additional training to support staff');
    }
    
    if (factors.includes('High number of critical incidents')) {
      recommendations.push('Implement stricter change management');
      recommendations.push('Increase monitoring and alerting coverage');
    }
    
    return recommendations;
  }
  
  async compareTimeperiods(currentPeriod = 'week', comparisonPeriod = 'week') {
    const [current, previous] = await Promise.all([
      this.getStats(currentPeriod),
      this.getStats(comparisonPeriod) // In real implementation, would adjust date range
    ]);
    
    return {
      current: current.data.summary,
      previous: previous.data.summary,
      changes: {
        total_incidents: this.calculateChange(
          previous.data.summary.total_incidents,
          current.data.summary.total_incidents
        ),
        resolution_rate: this.calculateChange(
          previous.data.summary.resolution_rate,
          current.data.summary.resolution_rate
        ),
        resolution_time: this.calculateChange(
          previous.data.summary.average_resolution_time_minutes,
          current.data.summary.average_resolution_time_minutes,
          'lower_is_better'
        )
      }
    };
  }
  
  calculateChange(oldValue, newValue, direction = 'higher_is_better') {
    const change = ((newValue - oldValue) / oldValue) * 100;
    const isImprovement = direction === 'higher_is_better' ? change > 0 : change < 0;
    
    return {
      percent: Math.abs(change).toFixed(1),
      direction: change > 0 ? 'increase' : 'decrease',
      is_improvement: isImprovement
    };
  }
  
  async exportReport(period = 'month', format = 'json') {
    const stats = await this.getStats(period);
    const summary = await this.generateExecutiveSummary(period);
    
    const report = {
      generated_at: new Date().toISOString(),
      period,
      executive_summary: summary,
      detailed_stats: stats.data,
      metadata: {
        total_incidents_analyzed: stats.data.summary.total_incidents,
        data_quality: 'complete',
        confidence_level: 'high'
      }
    };
    
    switch (format) {
      case 'csv':
        return this.convertToCsv(report);
      case 'pdf':
        return this.generatePdfReport(report);
      default:
        return report;
    }
  }
}

// Использование
const dashboard = new IncidentAnalyticsDashboard('your-api-key');

// Получение статистики
const weeklyStats = await dashboard.getStats('week', 'type');
console.log('Weekly stats:', weeklyStats);

// Исполнительная сводка
const executiveSummary = await dashboard.generateExecutiveSummary('month');
console.log('Executive summary:', executiveSummary);

// Сравнение периодов
const comparison = await dashboard.compareTimeperiods('week', 'week');
console.log('Period comparison:', comparison);

// Экспорт отчета
const report = await dashboard.exportReport('month');
console.log('Monthly report:', report);
```

### Incident Trend Analyzer
```javascript
class IncidentTrendAnalyzer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async analyzeTrends(days = 30) {
    const trends = {
      daily_counts: [],
      patterns: {},
      predictions: {},
      anomalies: []
    };
    
    // Collect daily data for trend analysis
    for (let i = days; i >= 0; i--) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      
      const dayStats = await this.getDayStats(date);
      trends.daily_counts.push({
        date: date.toISOString().split('T')[0],
        count: dayStats.total_incidents,
        resolved: dayStats.resolved_incidents,
        resolution_rate: dayStats.resolution_rate
      });
    }
    
    // Analyze patterns
    trends.patterns = this.detectPatterns(trends.daily_counts);
    
    // Generate predictions
    trends.predictions = this.generatePredictions(trends.daily_counts);
    
    // Detect anomalies
    trends.anomalies = this.detectAnomalies(trends.daily_counts);
    
    return trends;
  }
  
  detectPatterns(dailyData) {
    return {
      weekly_pattern: this.detectWeeklyPattern(dailyData),
      trend_direction: this.calculateTrendDirection(dailyData),
      seasonality: this.detectSeasonality(dailyData),
      volatility: this.calculateVolatility(dailyData)
    };
  }
  
  detectWeeklyPattern(dailyData) {
    const byDayOfWeek = [0, 1, 2, 3, 4, 5, 6].map(day => ({
      day,
      average: 0,
      count: 0
    }));
    
    dailyData.forEach(data => {
      const dayOfWeek = new Date(data.date).getDay();
      byDayOfWeek[dayOfWeek].average += data.count;
      byDayOfWeek[dayOfWeek].count++;
    });
    
    byDayOfWeek.forEach(day => {
      day.average = day.count > 0 ? day.average / day.count : 0;
    });
    
    return byDayOfWeek;
  }
  
  generatePredictions(dailyData) {
    // Simple linear regression for trend prediction
    const x = dailyData.map((_, index) => index);
    const y = dailyData.map(data => data.count);
    
    const n = x.length;
    const sumX = x.reduce((sum, val) => sum + val, 0);
    const sumY = y.reduce((sum, val) => sum + val, 0);
    const sumXY = x.reduce((sum, val, i) => sum + val * y[i], 0);
    const sumXX = x.reduce((sum, val) => sum + val * val, 0);
    
    const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
    const intercept = (sumY - slope * sumX) / n;
    
    // Predict next 7 days
    const predictions = [];
    for (let i = 1; i <= 7; i++) {
      const predictedCount = Math.max(0, Math.round(slope * (n + i - 1) + intercept));
      const date = new Date();
      date.setDate(date.getDate() + i);
      
      predictions.push({
        date: date.toISOString().split('T')[0],
        predicted_count: predictedCount,
        confidence: this.calculatePredictionConfidence(slope, dailyData)
      });
    }
    
    return {
      next_7_days: predictions,
      trend_slope: slope,
      trend_direction: slope > 0.1 ? 'increasing' : slope < -0.1 ? 'decreasing' : 'stable'
    };
  }
  
  calculatePredictionConfidence(slope, data) {
    const variance = this.calculateVariance(data.map(d => d.count));
    const stability = Math.exp(-Math.abs(slope) - variance / 100);
    return Math.min(Math.max(stability, 0.1), 0.9);
  }
  
  calculateVariance(values) {
    const mean = values.reduce((sum, val) => sum + val, 0) / values.length;
    const squaredDiffs = values.map(val => Math.pow(val - mean, 2));
    return squaredDiffs.reduce((sum, val) => sum + val, 0) / values.length;
  }
}
```

## Связанные endpoints
- [`GET /api/v1/incidents`](./list-incidents.md) - Данные для детального анализа
- [`GET /api/v1/incidents/:id`](./get-incident.md) - Детали для анализа паттернов
- [`POST /api/v1/incidents/:id/resolve`](./resolve-incident.md) - Влияет на метрики решений
