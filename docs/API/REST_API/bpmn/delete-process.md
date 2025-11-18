# DELETE /api/v1/bpmn/processes/:id

## Описание
Удаление определения BPMN процесса из системы. Внимание: это операция удаляет процесс полностью, включая все его версии.

## URL
```
DELETE /api/v1/bpmn/processes/{process_id}
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры пути
- `process_id` (string): ID процесса для удаления (не process_key!)

## Параметры запроса (Query Parameters)
- `force` (boolean): Принудительное удаление даже если есть активные экземпляры
- `cascade` (boolean): Удалить связанные данные (экземпляры, историю)

## Примеры запросов

### Базовое удаление
```bash
curl -X DELETE "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2" \
  -H "X-API-Key: your-api-key-here"
```

### Принудительное удаление
```bash
curl -X DELETE "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2?force=true&cascade=true" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const processId = 'order-processing-v2';

const response = await fetch(`/api/v1/bpmn/processes/${processId}?force=true`, {
  method: 'DELETE',
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

if (response.ok) {
  console.log('Process deleted successfully');
} else {
  const error = await response.json();
  console.error('Deletion failed:', error.error.message);
}
```

## Ответы

### 200 OK - Процесс удален
```json
{
  "success": true,
  "data": {
    "process_id": "order-processing-v2",
    "deleted_at": "2025-01-11T10:30:00.000Z",
    "deleted_by": "api-key-admin",
    "versions_deleted": 3,
    "instances_affected": {
      "active_instances": 0,
      "completed_instances": 1247,
      "cancelled_instances": 20,
      "total_affected": 1267
    },
    "cleanup_summary": {
      "process_definitions_deleted": 3,
      "process_instances_archived": 1267,
      "timers_cancelled": 0,
      "jobs_cancelled": 0,
      "messages_cleaned": 15,
      "storage_freed_bytes": 2048576
    }
  },
  "request_id": "req_1641998401700"
}
```

### 409 Conflict - Есть активные экземпляры
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_HAS_ACTIVE_INSTANCES",
    "message": "Cannot delete process with active instances",
    "details": {
      "process_id": "order-processing-v2",
      "active_instances": 23,
      "active_instance_ids": [
        "srv1-aB3dEf9hK2mN5pQ8uV",
        "srv1-cD4eF8gH1jK3mN6pQ9",
        "srv1-eF5gH9iJ2kL4mN7pR0"
      ],
      "suggestion": "Use force=true to delete anyway or cancel active instances first"
    }
  },
  "request_id": "req_1641998401701"
}
```

### 404 Not Found - Процесс не найден
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_NOT_FOUND",
    "message": "BPMN process not found",
    "details": {
      "process_id": "non-existent-process",
      "available_processes": [
        "order-processing-v2",
        "user-registration",
        "payment-workflow"
      ]
    }
  },
  "request_id": "req_1641998401702"
}
```

### 400 Bad Request - Нельзя удалить системный процесс
```json
{
  "success": false,
  "error": {
    "code": "SYSTEM_PROCESS_DELETION_FORBIDDEN",
    "message": "Cannot delete system processes",
    "details": {
      "process_id": "system-health-check",
      "reason": "Process is marked as system-critical"
    }
  },
  "request_id": "req_1641998401703"
}
```

## Поведение при удалении

### Что удаляется
1. **Определения процесса** - Все версии процесса
2. **Метаданные** - Статистика, файловая информация
3. **Зависимости** - Связанные task definitions, messages

### Что происходит с экземплярами

#### Без параметра `cascade`
- **Активные экземпляры**: Остаются работать, но нельзя запускать новые
- **Завершенные экземпляры**: Остаются в истории
- **Архивация**: Данные помечаются как orphaned

#### С параметром `cascade=true`
- **Активные экземпляры**: Отменяются (если `force=true`)
- **Завершенные экземпляры**: Архивируются или удаляются
- **История**: Очищается полностью

### Связанные операции
```javascript
// Что происходит при удалении процесса
const deletionEffects = {
  processDefinitions: 'Полное удаление всех версий',
  activeInstances: 'Отмена (если force=true) или блокировка удаления',
  completedInstances: 'Архивация (если cascade=true)',
  activeJobs: 'Отмена заданий',
  activeTimers: 'Отмена таймеров',
  messageSubscriptions: 'Удаление подписок',
  incidentRecords: 'Архивация инцидентов',
  auditLog: 'Запись события удаления'
};
```

## Предупреждения и безопасность

### ⚠️ Критические предупреждения

**1. Необратимая операция**
```bash
# ОПАСНО: Удаляет процесс навсегда
DELETE /api/v1/bpmn/processes/important-process?cascade=true&force=true
```

**2. Влияние на активные экземпляры**
```javascript
// Проверка перед удалением
async function safeProcessDeletion(processId) {
  // 1. Проверяем активные экземпляры
  const activeInstances = await getActiveInstances(processId);
  
  if (activeInstances.length > 0) {
    console.warn(`Found ${activeInstances.length} active instances`);
    
    // 2. Предлагаем отмену экземпляров
    const confirmCancel = confirm('Cancel all active instances?');
    if (confirmCancel) {
      await cancelAllInstances(activeInstances);
    } else {
      return false;
    }
  }
  
  // 3. Создаем backup
  await backupProcessDefinition(processId);
  
  // 4. Удаляем процесс
  return await deleteProcess(processId, { force: true, cascade: true });
}
```

**3. Системные процессы**
```yaml
# Защищенные процессы (нельзя удалить)
system_processes:
  - system-health-check
  - audit-logging
  - backup-process
  - monitoring-workflow
```

## Использование

### Batch Cleanup
```bash
#!/bin/bash
# Массовое удаление устаревших процессов
UNUSED_PROCESSES=$(curl -s -H "X-API-Key: $API_KEY" \
  "/api/v1/bpmn/processes" | \
  jq -r '.data.processes[] | select(.usage_stats.total_instances == 0) | .process_id')

for process_id in $UNUSED_PROCESSES; do
  echo "Deleting unused process: $process_id"
  curl -X DELETE -H "X-API-Key: $API_KEY" \
    "/api/v1/bpmn/processes/$process_id?cascade=true"
done
```

### Version Cleanup
```javascript
// Удаление старых версий процесса
async function cleanupOldVersions(processId, keepVersions = 3) {
  const processes = await fetch(`/api/v1/bpmn/processes?name=${processId}`);
  const data = await processes.json();
  
  const allVersions = data.data.processes
    .sort((a, b) => b.version - a.version);
  
  const versionsToDelete = allVersions.slice(keepVersions);
  
  for (const version of versionsToDelete) {
    console.log(`Deleting old version: ${version.process_key}`);
    
    await fetch(`/api/v1/bpmn/processes/${version.process_id}`, {
      method: 'DELETE',
      headers: { 'X-API-Key': 'your-api-key' }
    });
  }
  
  return versionsToDelete.length;
}
```

### Safe Deletion Workflow
```javascript
class ProcessDeletionManager {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async safeDeletion(processId, options = {}) {
    const steps = [];
    
    try {
      // 1. Получаем информацию о процессе
      const processInfo = await this.getProcessInfo(processId);
      steps.push('Process info retrieved');
      
      // 2. Проверяем активные экземпляры
      const activeCount = processInfo.usage_stats.active_instances;
      if (activeCount > 0 && !options.force) {
        throw new Error(`Process has ${activeCount} active instances`);
      }
      steps.push('Active instances checked');
      
      // 3. Создаем backup
      if (options.backup) {
        await this.backupProcess(processId);
        steps.push('Backup created');
      }
      
      // 4. Отменяем активные экземпляры
      if (activeCount > 0 && options.force) {
        await this.cancelActiveInstances(processId);
        steps.push('Active instances cancelled');
      }
      
      // 5. Удаляем процесс
      const result = await this.deleteProcess(processId, options);
      steps.push('Process deleted');
      
      return {
        success: true,
        steps,
        result
      };
      
    } catch (error) {
      return {
        success: false,
        steps,
        error: error.message
      };
    }
  }
  
  async backupProcess(processId) {
    // Создание backup определения процесса
    const processJson = await fetch(`/api/v1/bpmn/processes/${processId}/json`);
    const backup = await processJson.json();
    
    // Сохранение в backup storage
    await this.saveBackup(processId, backup);
  }
}
```

### Rollback Capability
```javascript
// Восстановление удаленного процесса из backup
async function restoreProcessFromBackup(processId, backupTimestamp) {
  const backup = await loadBackup(processId, backupTimestamp);
  
  // Восстанавливаем через BPMN parse
  const formData = new FormData();
  formData.append('file', new Blob([backup.bpmnXml], { type: 'application/xml' }));
  formData.append('process_id', processId);
  formData.append('force', 'true');
  
  const response = await fetch('/api/v1/bpmn/parse', {
    method: 'POST',
    headers: { 'X-API-Key': 'your-api-key' },
    body: formData
  });
  
  return response.json();
}
```

## Мониторинг и аудит

### Audit Logging
```json
{
  "event": "PROCESS_DELETED",
  "timestamp": "2025-01-11T10:30:00.000Z",
  "user": "api-key-admin",
  "details": {
    "process_id": "order-processing-v2",
    "versions_deleted": 3,
    "instances_affected": 1267,
    "forced": true,
    "cascade": true
  }
}
```

### Recovery Monitoring
```javascript
// Мониторинг orphaned instances после удаления
async function monitorOrphanedInstances() {
  const instances = await fetch('/api/v1/processes?status=ACTIVE');
  const data = await instances.json();
  
  const orphaned = data.data.processes.filter(instance => {
    // Проверяем существование процесса
    return !processExists(instance.process_id);
  });
  
  if (orphaned.length > 0) {
    console.warn(`Found ${orphaned.length} orphaned instances`);
    // Отправляем alert
    await sendAlert('ORPHANED_INSTANCES_DETECTED', orphaned);
  }
}
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes`](./list-processes.md) - Список процессов перед удалением
- [`GET /api/v1/bpmn/processes/:key`](./get-process.md) - Проверка деталей перед удалением
- [`GET /api/v1/processes`](../processes/list-processes.md) - Проверка активных экземпляров
- [`POST /api/v1/bpmn/parse`](./parse-bpmn.md) - Восстановление из backup
