# POST /api/v1/bpmn/parse

## Описание
Парсинг BPMN файла и сохранение определения процесса в систему. Преобразует BPMN XML в JSON формат для выполнения.

## URL
```
POST /api/v1/bpmn/parse
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Content Type
```
multipart/form-data
```

## Параметры формы

### Обязательные поля
- `file` (file): BPMN XML файл для парсинга

### Опциональные поля
- `process_id` (string): Кастомный ID процесса (если не указан, берется из XML)
- `force` (boolean): Принудительная перезапись существующего процесса

## Примеры запросов

### cURL
```bash
curl -X POST "http://localhost:27555/api/v1/bpmn/parse" \
  -H "X-API-Key: your-api-key-here" \
  -F "file=@process.bpmn" \
  -F "process_id=order-processing-v2" \
  -F "force=true"
```

### JavaScript (FormData)
```javascript
const formData = new FormData();
formData.append('file', bpmnFile); // File object
formData.append('process_id', 'order-processing-v2');
formData.append('force', 'true');

const response = await fetch('/api/v1/bpmn/parse', {
  method: 'POST',
  headers: {
    'X-API-Key': 'your-api-key-here'
  },
  body: formData
});

const result = await response.json();
```

### Go (multipart)
```go
var buf bytes.Buffer
writer := multipart.NewWriter(&buf)

// Добавляем файл
fileWriter, _ := writer.CreateFormFile("file", "process.bpmn")
file, _ := os.Open("process.bpmn")
io.Copy(fileWriter, file)

// Добавляем параметры
writer.WriteField("process_id", "order-processing-v2")
writer.WriteField("force", "true")
writer.Close()

req, _ := http.NewRequest("POST", "/api/v1/bpmn/parse", &buf)
req.Header.Set("Content-Type", writer.FormDataContentType())
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Ответы

### 201 Created - Процесс успешно распарсен
```json
{
  "success": true,
  "data": {
    "process_id": "order-processing-v2",
    "process_key": "order-processing-v2-1", 
    "version": 1,
    "name": "Order Processing Workflow",
    "description": "Complete order processing from validation to fulfillment",
    "created_at": "2025-01-11T10:30:00.000Z",
    "tenant_id": "default",
    "file_size_bytes": 15420,
    "parsing_time_ms": 245,
    "elements": {
      "start_events": 1,
      "end_events": 1, 
      "tasks": 8,
      "gateways": 3,
      "flows": 12,
      "boundary_events": 2,
      "total": 27
    },
    "validation": {
      "errors": [],
      "warnings": [
        "Task 'send-email' has no outgoing flow timeout configured"
      ],
      "is_executable": true
    },
    "metadata": {
      "file_name": "process.bpmn",
      "file_hash": "sha256:abc123def456...",
      "bpmn_version": "2.0",
      "namespace": "http://bpmn.io/schema/bpmn"
    }
  },
  "request_id": "req_1641998401400"
}
```

### 400 Bad Request - Ошибка валидации BPMN
```json
{
  "success": false,
  "error": {
    "code": "BPMN_VALIDATION_ERROR",
    "message": "BPMN file contains validation errors",
    "details": {
      "file_name": "process.bpmn",
      "validation_errors": [
        {
          "element_id": "task_1",
          "element_type": "serviceTask",
          "error": "Missing task definition attribute",
          "line": 45,
          "column": 12
        },
        {
          "element_id": "gateway_1", 
          "element_type": "exclusiveGateway",
          "error": "Gateway has no outgoing flows",
          "line": 78,
          "column": 8
        }
      ],
      "warning_count": 3,
      "error_count": 2
    }
  },
  "request_id": "req_1641998401401"
}
```

### 409 Conflict - Процесс уже существует
```json
{
  "success": false,
  "error": {
    "code": "PROCESS_ALREADY_EXISTS",
    "message": "Process with this ID already exists",
    "details": {
      "process_id": "order-processing-v2",
      "existing_version": 1,
      "created_at": "2025-01-10T15:30:00.000Z",
      "suggestion": "Use force=true to overwrite or choose different process_id"
    }
  },
  "request_id": "req_1641998401402"
}
```

### 413 Request Entity Too Large - Файл слишком большой
```json
{
  "success": false,
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "BPMN file exceeds maximum size limit",
    "details": {
      "file_size_bytes": 10485760,
      "max_size_bytes": 5242880,
      "max_size_mb": 5
    }
  },
  "request_id": "req_1641998401403"
}
```

## Поля ответа (успешный парсинг)

### Process Information
- `process_id` (string): ID процесса
- `process_key` (string): Уникальный ключ с версией
- `version` (integer): Версия процесса
- `name` (string): Название процесса
- `description` (string): Описание процесса

### Parsing Results
- `created_at` (string): Время создания
- `file_size_bytes` (integer): Размер файла
- `parsing_time_ms` (integer): Время парсинга

### Element Statistics
- `start_events` (integer): Количество start events
- `end_events` (integer): Количество end events
- `tasks` (integer): Количество задач
- `gateways` (integer): Количество шлюзов
- `flows` (integer): Количество потоков
- `boundary_events` (integer): Количество boundary events
- `total` (integer): Общее количество элементов

### Validation Results
- `errors` (array): Ошибки валидации
- `warnings` (array): Предупреждения
- `is_executable` (boolean): Может ли процесс выполняться

## Валидация BPMN

### Поддерживаемые элементы
```yaml
Events:
  - startEvent
  - endEvent  
  - intermediateCatchEvent
  - intermediateThrowEvent
  - boundaryEvent

Tasks:
  - serviceTask
  - userTask
  - scriptTask
  - sendTask
  - receiveTask
  - callActivity

Gateways:
  - exclusiveGateway
  - parallelGateway
  - inclusiveGateway
  - eventBasedGateway

Other:
  - sequenceFlow
  - messageFlow
  - textAnnotation
```

### Правила валидации
1. **Структурная валидация**
   - Каждый процесс должен иметь минимум один start event
   - Start events не должны иметь входящих потоков
   - End events не должны иметь исходящих потоков

2. **Семантическая валидация**
   - Service tasks должны иметь task definition
   - Exclusive gateways должны иметь условия на исходящих потоках
   - Message events должны ссылаться на message definitions

3. **Ограничения**
   - Максимум 1000 элементов в процессе
   - Максимальная глубина вложенности: 10 уровней
   - Уникальность ID элементов

## Форматы файлов

### Поддерживаемые форматы
- `.bpmn` - BPMN XML files
- `.xml` - Generic XML files with BPMN content

### Максимальные размеры
- **Размер файла**: 5MB
- **Количество элементов**: 1000
- **Длина process_id**: 100 символов

### Пример BPMN файла
```xml
<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                   xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
                   id="Definitions_1">
  <bpmn:process id="order-processing" isExecutable="true">
    <bpmn:startEvent id="start" name="Order Received">
      <bpmn:outgoing>flow1</bpmn:outgoing>
    </bpmn:startEvent>
    
    <bpmn:serviceTask id="validate-order" name="Validate Order">
      <bpmn:incoming>flow1</bpmn:incoming>
      <bpmn:outgoing>flow2</bpmn:outgoing>
      <bpmn:extensionElements>
        <zeebe:taskDefinition type="order-validator" />
      </bpmn:extensionElements>
    </bpmn:serviceTask>
    
    <bpmn:endEvent id="end" name="Order Processed">
      <bpmn:incoming>flow2</bpmn:incoming>
    </bpmn:endEvent>
    
    <bpmn:sequenceFlow id="flow1" sourceRef="start" targetRef="validate-order" />
    <bpmn:sequenceFlow id="flow2" sourceRef="validate-order" targetRef="end" />
  </bpmn:process>
</bpmn:definitions>
```

## Использование

### Batch Import
```bash
#!/bin/bash
# Импорт множественных BPMN файлов
for file in *.bpmn; do
  echo "Importing $file..."
  curl -X POST /api/v1/bpmn/parse \
    -H "X-API-Key: $API_KEY" \
    -F "file=@$file" \
    -F "force=true"
  echo ""
done
```

### CI/CD Integration
```yaml
# .github/workflows/deploy-processes.yml
name: Deploy BPMN Processes
on:
  push:
    paths: ['processes/**/*.bpmn']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Deploy BPMN files
        run: |
          for file in processes/*.bpmn; do
            curl -X POST ${{ secrets.ATOM_ENGINE_URL }}/api/v1/bpmn/parse \
              -H "X-API-Key: ${{ secrets.API_KEY }}" \
              -F "file=@$file" \
              -F "force=true"
          done
```

### JavaScript Helper
```javascript
class BPMNParser {
  constructor(apiKey, baseUrl) {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }
  
  async parseFile(file, options = {}) {
    const formData = new FormData();
    formData.append('file', file);
    
    if (options.processId) {
      formData.append('process_id', options.processId);
    }
    
    if (options.force) {
      formData.append('force', 'true');
    }
    
    const response = await fetch(`${this.baseUrl}/api/v1/bpmn/parse`, {
      method: 'POST',
      headers: {
        'X-API-Key': this.apiKey
      },
      body: formData
    });
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(`BPMN parsing failed: ${result.error.message}`);
    }
    
    return result.data;
  }
  
  async validateFile(file) {
    try {
      const result = await this.parseFile(file, { dryRun: true });
      return {
        valid: result.validation.is_executable,
        errors: result.validation.errors,
        warnings: result.validation.warnings
      };
    } catch (error) {
      return {
        valid: false,
        errors: [error.message],
        warnings: []
      };
    }
  }
}
```

## Troubleshooting

### Частые ошибки

#### XML parsing error
```xml
<!-- Неправильно -->
<bpmn:startEvent id="start"
  <bpmn:outgoing>flow1</bpmn:outgoing>
</bpmn:startEvent>

<!-- Правильно -->
<bpmn:startEvent id="start">
  <bpmn:outgoing>flow1</bpmn:outgoing>
</bpmn:startEvent>
```

#### Missing task definition
```xml
<!-- Неправильно -->
<bpmn:serviceTask id="task1" name="Process Order" />

<!-- Правильно -->
<bpmn:serviceTask id="task1" name="Process Order">
  <bpmn:extensionElements>
    <zeebe:taskDefinition type="order-processor" />
  </bpmn:extensionElements>
</bpmn:serviceTask>
```

#### Validation debug
```bash
# Проверка конкретной ошибки
curl -X POST /api/v1/bpmn/parse \
  -H "X-API-Key: $API_KEY" \
  -F "file=@broken-process.bpmn" \
  --silent | jq '.error.details.validation_errors'
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes`](./list-processes.md) - Список процессов
- [`GET /api/v1/bpmn/processes/:key`](./get-process.md) - Детали процесса
- [`DELETE /api/v1/bpmn/processes/:id`](./delete-process.md) - Удаление процесса
- [`GET /api/v1/bpmn/stats`](./get-bpmn-stats.md) - Статистика парсинга
