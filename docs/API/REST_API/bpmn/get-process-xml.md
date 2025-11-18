# GET /api/v1/bpmn/processes/:key/xml

## Описание
Получение оригинального BPMN XML содержимого процесса из файловой системы.

## URL
```
GET /api/v1/bpmn/processes/{process_key}/xml
```

## Авторизация
✅ **Требуется API ключ** с разрешением `bpmn`

## Параметры пути
- `process_key` (string): Ключ процесса (PROCESS KEY)

## Примеры запросов

### cURL
```bash
curl -X GET "http://localhost:27555/api/v1/bpmn/processes/atom-7-1k2-PVn4Y9j-CF5M/xml" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const processKey = 'atom-7-1k2-PVn4Y9j-CF5M';
const response = await fetch(`/api/v1/bpmn/processes/${processKey}/xml`, {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const xmlContent = await response.text();
console.log('BPMN XML:', xmlContent);
```

## Ответы

### 200 OK - XML получен
**Headers:**
```
Content-Type: application/xml; charset=utf-8
Content-Disposition: inline; filename="process.bpmn"
Content-Length: 3845
```

**Body:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="Process_1" name="Example Process" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1" name="Start">
      <bpmn:outgoing>Flow_1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:endEvent id="EndEvent_1" name="End">
      <bpmn:incoming>Flow_1</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="StartEvent_1" targetRef="EndEvent_1" />
  </bpmn:process>
</bpmn:definitions>
```

### 404 Not Found - Процесс не найден
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "BPMN process XML not found"
  }
}
```

## Использование

### Сохранение файла
```bash
curl -O -J "http://localhost:27555/api/v1/bpmn/processes/atom-7-1k2-PVn4Y9j-CF5M/xml" \
  -H "X-API-Key: your-api-key-here"
```

### Валидация XML
```javascript
async function validateBPMNXML(processKey) {
  const response = await fetch(`/api/v1/bpmn/processes/${processKey}/xml`);
  const xmlContent = await response.text();
  
  const parser = new DOMParser();
  const doc = parser.parseFromString(xmlContent, 'application/xml');
  
  const parseErrors = doc.querySelectorAll('parsererror');
  if (parseErrors.length > 0) {
    console.error('XML parsing errors:', parseErrors);
    return false;
  }
  
  console.log('✅ Valid BPMN XML');
  return true;
}
```

## Связанные endpoints
- [`GET /api/v1/bpmn/processes/:key/json`](./get-process-json.md) - JSON процесса
- [`GET /api/v1/bpmn/processes/:key`](./get-process.md) - Метаданные процесса
