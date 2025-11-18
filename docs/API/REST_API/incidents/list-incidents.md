# GET /api/v1/incidents

## Описание
Получение списка инцидентов в системе с фильтрацией и пагинацией. Инциденты представляют собой проблемы, возникшие во время выполнения процессов.

## URL
```
GET /api/v1/incidents
```

## Авторизация
✅ **Требуется API ключ** с разрешением `incident`

## Параметры запроса (Query Parameters)

### Фильтрация
- `status` (string): Статус инцидента (`open`, `resolved`, `dismissed`)
- `type` (string): Тип инцидента (`job`, `bpmn`, `expression`, `process`, `timer`, `message`, `system`)
- `process_instance_id` (string): ID экземпляра процесса
- `created_after` (string): Дата создания после (ISO 8601)
- `created_before` (string): Дата создания до (ISO 8601)

### Пагинация
- `limit` (integer): Количество записей (по умолчанию: 10, максимум: 100)
- `offset` (integer): Смещение для пагинации (по умолчанию: 0)

## Примеры запросов

### Все инциденты
```bash
curl -X GET "http://localhost:27555/api/v1/incidents" \
  -H "X-API-Key: your-api-key-here"
```

### Открытые инциденты задач
```bash
curl -X GET "http://localhost:27555/api/v1/incidents?status=open&type=job" \
  -H "X-API-Key: your-api-key-here"
```

### Инциденты конкретного процесса
```bash
curl -X GET "http://localhost:27555/api/v1/incidents?process_instance_id=srv1-aB3dEf9hK2mN5pQ8uV&limit=20" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/incidents?status=open&limit=50', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const incidents = await response.json();
```

## Ответы

### 200 OK - Список инцидентов
```json
{
  "success": true,
  "data": {
    "incidents": [
      {
        "id": "srv1-inc123abc456def789",
        "type": "JOB",
        "status": "OPEN",
        "message": "Job execution failed after maximum retries",
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:15:30.123Z",
        "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
        "process_definition_key": "Order_Processing",
        "element_id": "ServiceTask_PaymentProcess",
        "element_name": "Process Payment",
        "job_key": "srv1-job789xyz123abc456",
        "retries_left": 0,
        "error_details": {
          "error_code": "PAYMENT_SERVICE_UNAVAILABLE",
          "error_message": "Payment service timeout after 30 seconds",
          "stack_trace": "PaymentException: Connection timeout\n  at PaymentService.process()",
          "custom_headers": {
            "x-correlation-id": "corr-123456789",
            "x-tenant-id": "tenant-001"
          }
        },
        "resolution_attempts": 0,
        "severity": "HIGH",
        "tenant_id": "tenant-001"
      },
      {
        "id": "srv1-inc456def789ghi012",
        "type": "EXPRESSION",
        "status": "RESOLVED",
        "message": "Invalid expression syntax in gateway condition",
        "created_at": "2025-01-11T09:45:15.456Z",
        "updated_at": "2025-01-11T10:30:22.789Z",
        "resolved_at": "2025-01-11T10:30:22.789Z",
        "process_instance_id": "srv1-cD6fGh9iJ3kL7mN0pR",
        "process_definition_key": "Invoice_Approval",
        "element_id": "ExclusiveGateway_AmountCheck", 
        "element_name": "Amount Decision Gateway",
        "error_details": {
          "error_code": "EXPRESSION_SYNTAX_ERROR",
          "error_message": "Unexpected token ')' at position 15",
          "expression": "amount > 1000 and )",
          "suggestion": "Complete the expression or remove the extra parenthesis"
        },
        "resolution": {
          "type": "RETRY",
          "resolved_by": "system",
          "resolution_comment": "Expression corrected automatically",
          "retries_performed": 1
        },
        "severity": "MEDIUM",
        "tenant_id": "tenant-001"
      }
    ],
    "pagination": {
      "total": 87,
      "limit": 10,
      "offset": 0,
      "has_next": true,
      "has_previous": false
    },
    "summary": {
      "total_incidents": 87,
      "open_incidents": 42,
      "resolved_incidents": 38,
      "dismissed_incidents": 7,
      "by_type": {
        "JOB": 35,
        "EXPRESSION": 18,
        "BPMN": 12,
        "PROCESS": 8,
        "TIMER": 6,
        "MESSAGE": 5,
        "SYSTEM": 3
      },
      "by_severity": {
        "LOW": 25,
        "MEDIUM": 38,
        "HIGH": 19,
        "CRITICAL": 5
      }
    }
  },
  "request_id": "req_1641998404400"
}
```

## Incident Types

### JOB
Ошибки выполнения заданий (service tasks, send tasks)

### EXPRESSION
Ошибки в FEEL выражениях и условиях

### BPMN
Ошибки структуры процесса или недостающие элементы

### PROCESS
Ошибки экземпляра процесса или его жизненного цикла

### TIMER
Ошибки таймеров и временных событий

### MESSAGE
Ошибки обработки сообщений и корреляции

### SYSTEM
Системные ошибки и ошибки инфраструктуры

## Incident Status

### OPEN
Инцидент активен и требует внимания

### RESOLVED
Инцидент решен успешно

### DISMISSED
Инцидент отклонен без решения

## Использование

### Incident Monitor
```javascript
class IncidentMonitor {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.pollInterval = null;
  }
  
  async getIncidents(filters = {}) {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        params.append(key, value);
      }
    });
    
    const response = await fetch(`/api/v1/incidents?${params}`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  async getCriticalIncidents() {
    // Get open incidents with high/critical severity
    const filters = {
      status: 'open',
      limit: 100
    };
    
    const result = await this.getIncidents(filters);
    
    return result.data.incidents.filter(incident => 
      ['HIGH', 'CRITICAL'].includes(incident.severity)
    );
  }
  
  async getIncidentsByProcess(processInstanceId) {
    return await this.getIncidents({
      process_instance_id: processInstanceId,
      limit: 50
    });
  }
  
  async getIncidentTrends(hours = 24) {
    const since = new Date(Date.now() - hours * 60 * 60 * 1000).toISOString();
    
    const result = await this.getIncidents({
      created_after: since,
      limit: 1000
    });
    
    return this.analyzeIncidentTrends(result.data.incidents, hours);
  }
  
  analyzeIncidentTrends(incidents, hours) {
    const hourlyBuckets = new Array(hours).fill(0).map(() => ({
      hour: 0,
      count: 0,
      by_type: {},
      by_severity: {}
    }));
    
    incidents.forEach(incident => {
      const incidentHour = Math.floor(
        (Date.now() - new Date(incident.created_at).getTime()) / (60 * 60 * 1000)
      );
      
      if (incidentHour < hours) {
        const bucket = hourlyBuckets[incidentHour];
        bucket.count++;
        
        bucket.by_type[incident.type] = (bucket.by_type[incident.type] || 0) + 1;
        bucket.by_severity[incident.severity] = (bucket.by_severity[incident.severity] || 0) + 1;
      }
    });
    
    return {
      hourly_counts: hourlyBuckets,
      total_incidents: incidents.length,
      peak_hour: hourlyBuckets.reduce((max, bucket, index) => 
        bucket.count > hourlyBuckets[max].count ? index : max, 0
      ),
      average_per_hour: incidents.length / hours
    };
  }
  
  startPolling(callback, interval = 30000) {
    this.pollInterval = setInterval(async () => {
      try {
        const criticalIncidents = await this.getCriticalIncidents();
        callback(criticalIncidents);
      } catch (error) {
        console.error('Error polling incidents:', error);
      }
    }, interval);
  }
  
  stopPolling() {
    if (this.pollInterval) {
      clearInterval(this.pollInterval);
      this.pollInterval = null;
    }
  }
}

// Использование
const monitor = new IncidentMonitor('your-api-key');

// Получение всех инцидентов
const allIncidents = await monitor.getIncidents({ limit: 50 });
console.log('All incidents:', allIncidents);

// Критические инциденты
const criticalIncidents = await monitor.getCriticalIncidents();
console.log('Critical incidents:', criticalIncidents);

// Инциденты по процессу
const processIncidents = await monitor.getIncidentsByProcess('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Process incidents:', processIncidents);

// Анализ трендов
const trends = await monitor.getIncidentTrends(24);
console.log('Incident trends:', trends);

// Мониторинг в реальном времени
monitor.startPolling((criticalIncidents) => {
  if (criticalIncidents.length > 0) {
    console.warn('Critical incidents detected:', criticalIncidents.length);
    // Send alerts
  }
}, 30000);
```

### Incident Dashboard
```javascript
class IncidentDashboard {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.monitor = new IncidentMonitor(apiKey);
  }
  
  async generateDashboardData() {
    const [
      allIncidents,
      criticalIncidents,
      trends
    ] = await Promise.all([
      this.monitor.getIncidents({ limit: 1000 }),
      this.monitor.getCriticalIncidents(),
      this.monitor.getIncidentTrends(24)
    ]);
    
    return {
      overview: this.createOverview(allIncidents.data),
      critical_alerts: criticalIncidents,
      trends: trends,
      charts: this.prepareChartData(allIncidents.data)
    };
  }
  
  createOverview(incidentData) {
    const { summary } = incidentData;
    
    return {
      total: summary.total_incidents,
      open: summary.open_incidents,
      resolved: summary.resolved_incidents,
      dismissed: summary.dismissed_incidents,
      resolution_rate: summary.total_incidents > 0 ? 
        ((summary.resolved_incidents / summary.total_incidents) * 100).toFixed(1) : 0,
      most_common_type: Object.entries(summary.by_type)
        .sort(([,a], [,b]) => b - a)[0]?.[0] || 'N/A',
      severity_distribution: summary.by_severity
    };
  }
  
  prepareChartData(incidentData) {
    return {
      by_type: Object.entries(incidentData.summary.by_type).map(([type, count]) => ({
        label: type,
        value: count
      })),
      by_severity: Object.entries(incidentData.summary.by_severity).map(([severity, count]) => ({
        label: severity,
        value: count
      })),
      status_distribution: [
        { label: 'Open', value: incidentData.summary.open_incidents },
        { label: 'Resolved', value: incidentData.summary.resolved_incidents },
        { label: 'Dismissed', value: incidentData.summary.dismissed_incidents }
      ]
    };
  }
  
  async exportIncidents(filters = {}, format = 'json') {
    const incidents = await this.monitor.getIncidents({
      ...filters,
      limit: 10000 // Large limit for export
    });
    
    switch (format) {
      case 'csv':
        return this.convertToCSV(incidents.data.incidents);
      case 'xlsx':
        return this.convertToExcel(incidents.data.incidents);
      default:
        return incidents.data.incidents;
    }
  }
  
  convertToCSV(incidents) {
    const headers = [
      'ID', 'Type', 'Status', 'Message', 'Created At', 
      'Process Instance', 'Element ID', 'Severity', 'Tenant ID'
    ];
    
    const rows = incidents.map(incident => [
      incident.id,
      incident.type,
      incident.status,
      incident.message,
      incident.created_at,
      incident.process_instance_id || '',
      incident.element_id || '',
      incident.severity,
      incident.tenant_id || ''
    ]);
    
    return [headers, ...rows]
      .map(row => row.map(cell => `"${cell}"`).join(','))
      .join('\n');
  }
}
```

## Связанные endpoints
- [`GET /api/v1/incidents/:id`](./get-incident.md) - Детали конкретного инцидента
- [`POST /api/v1/incidents/:id/resolve`](./resolve-incident.md) - Решение инцидента
- [`GET /api/v1/incidents/stats`](./get-incident-stats.md) - Статистика инцидентов
