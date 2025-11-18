# GET /api/v1/bpmn/processes/:key/json

## Описание
Получение JSON представления BPMN процесса. Возвращает полное определение процесса в формате, используемом для выполнения в движке.

## URL
```
GET /api/v1/bpmn/processes/{process_key}/json
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры пути
- `process_key` (string): Уникальный ключ процесса

## Параметры запроса (Query Parameters)
- `format` (string): Формат ответа (`compact`, `pretty`, по умолчанию: `pretty`)
- `include_metadata` (boolean): Включить метаданные (по умолчанию: `true`)

## Примеры запросов

### Базовый запрос
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2-3/json" \
  -H "X-API-Key: your-api-key-here"
```

### Компактный формат
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2-3/json?format=compact" \
  -H "X-API-Key: your-api-key-here"
```

### Без метаданных
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes/order-processing-v2-3/json?include_metadata=false" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const processKey = 'order-processing-v2-3';
const response = await fetch(`/api/v1/bpmn/processes/${processKey}/json?format=pretty`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const processJson = await response.json();
console.log('Process definition:', processJson.data);
```

## Ответы

### 200 OK - JSON процесса получен
```json
{
  "success": true,
  "data": {
    "process_definition": {
      "process_id": "order-processing-v2",
      "process_key": "order-processing-v2-3",
      "version": 3,
      "name": "Order Processing Workflow",
      "description": "Complete order processing from validation to fulfillment",
      "tenant_id": "production",
      "is_executable": true,
      "elements": {
        "start_order": {
          "id": "start_order",
          "type": "startEvent",
          "name": "Order Received",
          "outgoing": ["flow_1"],
          "position": {
            "x": 100,
            "y": 100
          }
        },
        "validate_order": {
          "id": "validate_order",
          "type": "serviceTask",
          "name": "Validate Order",
          "incoming": ["flow_1"],
          "outgoing": ["flow_2"],
          "task_definition": {
            "type": "order-validator",
            "retries": 3,
            "timeout": "PT30S"
          },
          "position": {
            "x": 250,
            "y": 100
          }
        },
        "payment_gateway": {
          "id": "payment_gateway",
          "type": "exclusiveGateway",
          "name": "Payment Method?",
          "incoming": ["flow_2"],
          "outgoing": ["flow_3_credit", "flow_3_debit"],
          "conditions": {
            "flow_3_credit": "paymentMethod = 'credit'",
            "flow_3_debit": "paymentMethod = 'debit'"
          },
          "position": {
            "x": 400,
            "y": 100
          }
        },
        "process_credit": {
          "id": "process_credit",
          "type": "serviceTask", 
          "name": "Process Credit Payment",
          "incoming": ["flow_3_credit"],
          "outgoing": ["flow_4"],
          "task_definition": {
            "type": "credit-processor",
            "retries": 2,
            "timeout": "PT60S"
          },
          "position": {
            "x": 550,
            "y": 50
          }
        },
        "process_debit": {
          "id": "process_debit",
          "type": "serviceTask",
          "name": "Process Debit Payment", 
          "incoming": ["flow_3_debit"],
          "outgoing": ["flow_5"],
          "task_definition": {
            "type": "debit-processor",
            "retries": 2,
            "timeout": "PT60S"
          },
          "position": {
            "x": 550,
            "y": 150
          }
        },
        "join_gateway": {
          "id": "join_gateway",
          "type": "exclusiveGateway",
          "name": "Join",
          "incoming": ["flow_4", "flow_5"],
          "outgoing": ["flow_6"],
          "position": {
            "x": 700,
            "y": 100
          }
        },
        "send_confirmation": {
          "id": "send_confirmation",
          "type": "serviceTask",
          "name": "Send Confirmation",
          "incoming": ["flow_6"],
          "outgoing": ["flow_7"],
          "task_definition": {
            "type": "email-sender",
            "retries": 3,
            "timeout": "PT30S"
          },
          "position": {
            "x": 850,
            "y": 100
          }
        },
        "end_success": {
          "id": "end_success",
          "type": "endEvent",
          "name": "Order Completed",
          "incoming": ["flow_7"],
          "position": {
            "x": 1000,
            "y": 100
          }
        }
      },
      "flows": {
        "flow_1": {
          "id": "flow_1",
          "source": "start_order",
          "target": "validate_order"
        },
        "flow_2": {
          "id": "flow_2", 
          "source": "validate_order",
          "target": "payment_gateway"
        },
        "flow_3_credit": {
          "id": "flow_3_credit",
          "source": "payment_gateway",
          "target": "process_credit",
          "condition": "paymentMethod = 'credit'"
        },
        "flow_3_debit": {
          "id": "flow_3_debit",
          "source": "payment_gateway", 
          "target": "process_debit",
          "condition": "paymentMethod = 'debit'"
        },
        "flow_4": {
          "id": "flow_4",
          "source": "process_credit",
          "target": "join_gateway"
        },
        "flow_5": {
          "id": "flow_5",
          "source": "process_debit",
          "target": "join_gateway"
        },
        "flow_6": {
          "id": "flow_6",
          "source": "join_gateway",
          "target": "send_confirmation"
        },
        "flow_7": {
          "id": "flow_7",
          "source": "send_confirmation",
          "target": "end_success"
        }
      },
      "boundary_events": {
        "payment_timeout": {
          "id": "payment_timeout",
          "type": "boundaryEvent",
          "name": "Payment Timeout",
          "attached_to": "process_credit",
          "interrupting": true,
          "timer_definition": {
            "duration": "PT5M"
          },
          "outgoing": ["flow_timeout"]
        }
      },
      "message_definitions": {
        "order_confirmation": {
          "name": "order_confirmation",
          "correlation_key": "orderId"
        }
      },
      "variables": {
        "input_variables": [
          {
            "name": "orderId",
            "type": "string",
            "required": true
          },
          {
            "name": "paymentMethod",
            "type": "string",
            "required": true,
            "allowed_values": ["credit", "debit"]
          },
          {
            "name": "amount",
            "type": "number",
            "required": true,
            "min": 0.01
          }
        ],
        "output_variables": [
          {
            "name": "transactionId",
            "type": "string"
          },
          {
            "name": "status",
            "type": "string"
          }
        ]
      }
    },
    "metadata": {
      "generated_at": "2025-01-11T10:30:00.000Z",
      "generated_by": "atom-engine-v1.0.0",
      "original_file": "order-processing-v2.bpmn",
      "file_hash": "sha256:abc123def456...",
      "parsing_version": "1.0.0",
      "bpmn_version": "2.0"
    },
    "statistics": {
      "total_elements": 8,
      "total_flows": 8,
      "total_boundary_events": 1,
      "complexity_score": 15,
      "estimated_paths": 2
    }
  },
  "request_id": "req_1641998401800"
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
      "process_key": "non-existent-process-key"
    }
  },
  "request_id": "req_1641998401801"
}
```

## Форматы ответа

### Pretty Format (по умолчанию)
- Человекочитаемый JSON с отступами
- Включает все метаданные
- Подходит для отладки и анализа

### Compact Format
- Минимизированный JSON без отступов
- Меньший размер для передачи по сети
- Подходит для production интеграций

### Без метаданных
```json
{
  "success": true,
  "data": {
    "process_definition": {
      // Только определение процесса без metadata и statistics
    }
  }
}
```

## Структура JSON процесса

### Process Definition
- **Основная информация**: ID, версия, название, описание
- **Elements**: Все элементы процесса (tasks, events, gateways)
- **Flows**: Последовательные потоки между элементами
- **Boundary Events**: События, привязанные к активностям
- **Message Definitions**: Определения сообщений
- **Variables**: Входные и выходные переменные

### Element Structure
```json
{
  "element_id": {
    "id": "string",
    "type": "elementType",
    "name": "string",
    "incoming": ["flow_ids"],
    "outgoing": ["flow_ids"],
    "position": { "x": number, "y": number },
    // Специфичные поля для типа элемента
  }
}
```

### Task Definition
```json
{
  "task_definition": {
    "type": "worker-type",
    "retries": 3,
    "timeout": "PT30S",
    "custom_headers": {
      "priority": "high"
    }
  }
}
```

## Использование

### Process Visualization
```javascript
async function visualizeProcess(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}/json`);
  const data = await response.json();
  
  const definition = data.data.process_definition;
  
  // Создание диаграммы
  const nodes = Object.values(definition.elements).map(element => ({
    id: element.id,
    label: element.name,
    type: element.type,
    position: element.position
  }));
  
  const edges = Object.values(definition.flows).map(flow => ({
    id: flow.id,
    source: flow.source,
    target: flow.target,
    label: flow.condition || ''
  }));
  
  return { nodes, edges };
}
```

### Process Validation
```javascript
async function validateProcessStructure(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}/json`);
  const data = await response.json();
  
  const definition = data.data.process_definition;
  const issues = [];
  
  // Проверка start events
  const startEvents = Object.values(definition.elements)
    .filter(e => e.type === 'startEvent');
  
  if (startEvents.length === 0) {
    issues.push('No start events found');
  }
  
  // Проверка end events  
  const endEvents = Object.values(definition.elements)
    .filter(e => e.type === 'endEvent');
    
  if (endEvents.length === 0) {
    issues.push('No end events found');
  }
  
  // Проверка orphaned elements
  Object.values(definition.elements).forEach(element => {
    if (element.type !== 'startEvent' && (!element.incoming || element.incoming.length === 0)) {
      issues.push(`Element ${element.id} has no incoming flows`);
    }
    
    if (element.type !== 'endEvent' && (!element.outgoing || element.outgoing.length === 0)) {
      issues.push(`Element ${element.id} has no outgoing flows`);
    }
  });
  
  return issues;
}
```

### Process Migration
```javascript
// Миграция процесса между версиями движка
async function migrateProcessDefinition(processKey, targetVersion) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}/json`);
  const data = await response.json();
  
  const definition = data.data.process_definition;
  
  // Применяем миграционные правила
  const migrated = applyMigrationRules(definition, targetVersion);
  
  // Валидируем мигрированный процесс
  const validation = await validateMigratedProcess(migrated);
  
  if (validation.isValid) {
    // Сохраняем как новую версию
    return await deployMigratedProcess(migrated);
  } else {
    throw new Error(`Migration failed: ${validation.errors.join(', ')}`);
  }
}
```

### Process Comparison
```javascript
async function compareProcessVersions(processId, version1, version2) {
  const [v1Response, v2Response] = await Promise.all([
    fetch(`/api/v1/bpmn/processes/${processId}-${version1}/json`),
    fetch(`/api/v1/bpmn/processes/${processId}-${version2}/json`)
  ]);
  
  const v1Data = await v1Response.json();
  const v2Data = await v2Response.json();
  
  const v1Def = v1Data.data.process_definition;
  const v2Def = v2Data.data.process_definition;
  
  const differences = {
    added_elements: [],
    removed_elements: [],
    modified_elements: [],
    added_flows: [],
    removed_flows: []
  };
  
  // Сравнение элементов
  const v1Elements = Object.keys(v1Def.elements);
  const v2Elements = Object.keys(v2Def.elements);
  
  differences.added_elements = v2Elements.filter(id => !v1Elements.includes(id));
  differences.removed_elements = v1Elements.filter(id => !v2Elements.includes(id));
  
  // Сравнение потоков
  const v1Flows = Object.keys(v1Def.flows);
  const v2Flows = Object.keys(v2Def.flows);
  
  differences.added_flows = v2Flows.filter(id => !v1Flows.includes(id));
  differences.removed_flows = v1Flows.filter(id => !v2Flows.includes(id));
  
  return differences;
}
```

### Code Generation
```javascript
// Генерация кода на основе процесса
async function generateWorkerCode(processKey, language = 'javascript') {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}/json`);
  const data = await response.json();
  
  const definition = data.data.process_definition;
  
  // Находим все service tasks
  const serviceTasks = Object.values(definition.elements)
    .filter(e => e.type === 'serviceTask')
    .map(task => ({
      id: task.id,
      name: task.name,
      type: task.task_definition.type,
      timeout: task.task_definition.timeout,
      retries: task.task_definition.retries
    }));
  
  // Генерируем код для каждого task
  const workers = serviceTasks.map(task => 
    generateWorkerCode(task, language)
  );
  
  return workers;
}

function generateWorkerCodeForTask(task, language) {
  if (language === 'javascript') {
    return `
// Worker for ${task.name}
class ${toPascalCase(task.type)}Worker {
  async execute(job) {
    const variables = job.variables;
    
    try {
      // TODO: Implement ${task.name} logic
      const result = await this.process${toPascalCase(task.name)}(variables);
      
      await job.complete(result);
    } catch (error) {
      await job.fail(error.message);
    }
  }
  
  async process${toPascalCase(task.name)}(variables) {
    // Implementation here
    throw new Error('Implementation required - connect to AtomBPMN engine');
  }
}
`;
  }
}
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes/:key`](./get-process.md) - Метаданные процесса
- [`GET /api/v1/bpmn/processes`](./list-processes.md) - Список всех процессов
- [`POST /api/v1/bpmn/parse`](./parse-bpmn.md) - Парсинг BPMN в JSON
- [`POST /api/v1/processes`](../processes/start-process.md) - Запуск процесса по JSON
