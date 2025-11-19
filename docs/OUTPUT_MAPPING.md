# Output Mapping в Atom Engine

## Обзор

**Output Mapping** — это механизм автоматического маппинга выходных данных из результата выполнения задачи (HTTP request, Email, etc.) обратно в переменные процесса.

Реализовано в соответствии со стандартом **Camunda 8**.

---

## 🔄 Полный цикл IO Mapping

```
BPMN Process Variables
        ↓
  【 INPUT MAPPING 】
   Переменные → Параметры задачи
        ↓
   【 TASK EXECUTION 】
   Выполнение HTTP/Email/etc
        ↓
  【 OUTPUT MAPPING 】
   Результат → Переменные процесса
        ↓
BPMN Process Variables (обновлены)
```

---

## 📝 Синтаксис BPMN

### Структура ioMapping с Output

```xml
<zeebe:ioMapping>
  <!-- INPUT: Process Variables → Task Config -->
  <zeebe:input source="=orderId" target="customerId" />
  <zeebe:input source="POST" target="method" />
  <zeebe:input source="=apiUrl/users" target="url" />
  <zeebe:input source="={&#34;name&#34;:&#34;John&#34;}" target="body" />
  
  <!-- OUTPUT: Task Result → Process Variables -->
  <zeebe:output source="=response.body.id" target="userId" />
  <zeebe:output source="=response.body.email" target="userEmail" />
  <zeebe:output source="=response.status" target="httpStatus" />
  <zeebe:output source="=response.headers.Date" target="responseDate" />
</zeebe:ioMapping>
```

---

## 🎯 Примеры использования

### Пример 1: Извлечение данных из HTTP Response

#### BPMN:
```xml
<bpmn:serviceTask id="CreateUser" name="Create User">
  <bpmn:extensionElements>
    <zeebe:taskDefinition type="io.camunda:http-json:1" />
    <zeebe:ioMapping>
      <!-- Inputs -->
      <zeebe:input source="POST" target="method" />
      <zeebe:input source="https://api.example.com/users" target="url" />
      <zeebe:input source="={&#34;name&#34;:&#34;John Doe&#34;,&#34;email&#34;:&#34;john@example.com&#34;}" target="body" />
      
      <!-- Outputs - извлекаем нужные поля -->
      <zeebe:output source="=response.body.id" target="userId" />
      <zeebe:output source="=response.body.name" target="userName" />
      <zeebe:output source="=response.body.email" target="userEmail" />
      <zeebe:output source="=response.status" target="statusCode" />
    </zeebe:ioMapping>
  </bpmn:extensionElements>
</bpmn:serviceTask>
```

#### Переменные ДО выполнения:
```json
{}
```

#### HTTP Response:
```json
{
  "status": 201,
  "body": {
    "id": "user-12345",
    "name": "John Doe",
    "email": "john@example.com",
    "createdAt": "2025-11-19T10:00:00Z"
  },
  "headers": {
    "Content-Type": "application/json",
    "Date": "Wed, 19 Nov 2025 10:00:00 GMT"
  }
}
```

#### Переменные ПОСЛЕ output mapping:
```json
{
  "response": {
    "status": 201,
    "body": {
      "id": "user-12345",
      "name": "John Doe",
      "email": "john@example.com",
      "createdAt": "2025-11-19T10:00:00Z"
    },
    "headers": {
      "Content-Type": "application/json",
      "Date": "Wed, 19 Nov 2025 10:00:00 GMT"
    }
  },
  "userId": "user-12345",        ← Извлечено через output mapping
  "userName": "John Doe",         ← Извлечено через output mapping
  "userEmail": "john@example.com", ← Извлечено через output mapping
  "statusCode": 201               ← Извлечено через output mapping
}
```

---

### Пример 2: Proxmox VM Clone с Output Mapping

#### BPMN:
```xml
<bpmn:serviceTask id="PROXMOX_clone_vm" name="Clone VM">
  <bpmn:extensionElements>
    <zeebe:taskDefinition type="io.camunda:http-json:1" />
    <zeebe:ioMapping>
      <!-- Inputs -->
      <zeebe:input source="apiKey" target="authentication.type" />
      <zeebe:input source="headers" target="authentication.apiKeyLocation" />
      <zeebe:input source="Authorization" target="authentication.name" />
      <zeebe:input source="PVEAPIToken=root@pam!token=xxx" target="authentication.value" />
      <zeebe:input source="POST" target="method" />
      <zeebe:input source="=api_url/nodes/proxmox_vm_id/clone" target="url" />
      <zeebe:input source="=params" target="queryParameters" />
      
      <!-- Outputs - извлекаем taskId из ответа -->
      <zeebe:output source="=response.body.data" target="cloneTaskId" />
      <zeebe:output source="=response.status" target="cloneStatus" />
    </zeebe:ioMapping>
  </bpmn:extensionElements>
</bpmn:serviceTask>
```

#### Переменные ДО:
```json
{
  "api_url": "https://pve1.hlprod.ru:8006/api2/json",
  "proxmox_vm_id": "qemu/3013",
  "params": {
    "newid": "691699888",
    "name": "ru-test-vm",
    "target": "pve3",
    "full": "1",
    "storage": "ceph-pool"
  }
}
```

#### Proxmox Response:
```json
{
  "status": 200,
  "body": {
    "data": "UPID:pve3:0001F8B4:0018D5E9:67483C42:qmclone:691699888:root@pam:"
  }
}
```

#### Переменные ПОСЛЕ output mapping:
```json
{
  "api_url": "https://pve1.hlprod.ru:8006/api2/json",
  "proxmox_vm_id": "qemu/3013",
  "params": {...},
  "response": {
    "status": 200,
    "body": {
      "data": "UPID:pve3:0001F8B4:0018D5E9:67483C42:qmclone:691699888:root@pam:"
    }
  },
  "cloneTaskId": "UPID:pve3:0001F8B4:0018D5E9:67483C42:qmclone:691699888:root@pam:",
  "cloneStatus": 200
}
```

---

### Пример 3: Комплексные FEEL выражения

#### BPMN:
```xml
<zeebe:ioMapping>
  <zeebe:input source="GET" target="method" />
  <zeebe:input source="https://api.example.com/orders/123" target="url" />
  
  <!-- Комплексные FEEL выражения для извлечения данных -->
  <zeebe:output source="=response.body.order.id" target="orderId" />
  <zeebe:output source="=response.body.order.customer.name" target="customerName" />
  <zeebe:output source="=response.body.order.items[1].price" target="firstItemPrice" />
  <zeebe:output source="=response.body.order.total" target="orderTotal" />
</zeebe:ioMapping>
```

#### Response:
```json
{
  "status": 200,
  "body": {
    "order": {
      "id": "order-789",
      "customer": {
        "name": "Jane Smith",
        "email": "jane@example.com"
      },
      "items": [
        {"name": "Item 1", "price": 100},
        {"name": "Item 2", "price": 200}
      ],
      "total": 300
    }
  }
}
```

#### Результат:
```json
{
  "orderId": "order-789",
  "customerName": "Jane Smith",
  "firstItemPrice": 100,
  "orderTotal": 300
}
```

---

## 🔍 Как работает

### 1. Парсинг (src/parser/tasks.go)

Output mappings парсятся вместе с input mappings:

```go
func (p *TaskParser) parseZeebeIOMapping(element *XMLElement) map[string]interface{} {
    ioMapping := make(map[string]interface{})
    
    outputs := make([]map[string]interface{}, 0)
    for _, child := range element.Children {
        if child.XMLName.Local == "output" {
            output := p.parseZeebeOutput(child)
            outputs = append(outputs, output)
        }
    }
    
    ioMapping["outputs"] = outputs
    return ioMapping
}
```

### 2. Выполнение (src/process/http_connector.go)

После выполнения HTTP запроса применяется output mapping:

```go
func (hce *HttpConnectorExecutor) Execute(token, element) {
    // 1. Execute HTTP request
    response := executeHttpRequest(config)
    
    // 2. Store response in token.Variables["response"]
    updateTokenWithHttpResponse(token, response)
    
    // 3. Apply output mapping (NEW!)
    applyOutputMapping(element, token)
    
    // 4. Continue with next elements
    return nextElements
}
```

### 3. Применение Output Mapping

```go
func (hce *HttpConnectorExecutor) applyOutputMapping(element, token) {
    outputs := extractOutputMappings(element)
    
    for _, output := range outputs {
        source := output["source"]  // =response.body.id
        target := output["target"]  // userId
        
        // Evaluate FEEL expression
        value := evaluateInputValue(source, token.Variables)
        
        // Set target variable
        token.Variables[target] = value
    }
}
```

---

## 📊 Поддержка FEEL выражений

Output mapping поддерживает полноценные FEEL path expressions через встроенный PathNavigator:

| Expression | Описание | Пример результата |
|------------|----------|------------------|
| `=response.body.id` | Прямой доступ к полю | `"user-123"` |
| `=response.status` | HTTP status code | `200` |
| `=response.headers.Date` | Header значение | `"Wed, 19 Nov 2025..."` |
| `=response.body.items[0]` | Доступ к элементу массива | `{...}` |
| `=response.body.user.email` | Вложенные поля | `"user@example.com"` |
| `=users[0].emails[1]` | Множественные индексы | `"second@example.com"` |
| `=data[key]` | Доступ по переменной | `value` (где key - переменная) |

### FEEL Path Navigator

**Atom Engine** включает полноценный **PathNavigator** для обработки сложных FEEL path expressions:

#### Поддерживаемые паттерны

1. **Точечная нотация** (Dot notation)
   ```xml
   <zeebe:output source="=response.body.data" target="resultData" />
   ```
   Извлекает значение из вложенных полей объекта.

2. **Доступ к массивам** (Array access)
   ```xml
   <zeebe:output source="=items[0].name" target="firstName" />
   ```
   Доступ к элементу массива по числовому индексу.

3. **Доступ по переменной** (Variable-based access)
   ```xml
   <zeebe:output source="=data[fieldName]" target="fieldValue" />
   ```
   Использует значение переменной `fieldName` как ключ для доступа к map.

4. **Комплексные пути** (Complex paths)
   ```xml
   <zeebe:output source="=response.body.users[0].emails[1]" target="secondEmail" />
   ```
   Комбинация точечной нотации и множественных индексов массивов.

#### Примеры работы PathNavigator

```javascript
// Исходные данные в token.Variables
{
  "response": {
    "status": 200,
    "body": {
      "data": "UPID:pve3:000E0869...",
      "users": [
        {
          "id": 1,
          "emails": ["first@test.com", "second@test.com"]
        }
      ]
    },
    "headers": {
      "Date": "Wed, 19 Nov 2025 10:00:00 GMT"
    }
  }
}

// Output mappings
<zeebe:output source="=response.body.data" target="taskId" />
// Результат: taskId = "UPID:pve3:000E0869..."

<zeebe:output source="=response.status" target="httpStatus" />
// Результат: httpStatus = 200

<zeebe:output source="=response.body.users[0].emails[1]" target="email" />
// Результат: email = "second@test.com"

<zeebe:output source="=response.headers.Date" target="timestamp" />
// Результат: timestamp = "Wed, 19 Nov 2025 10:00:00 GMT"
```

#### Обработка ошибок

PathNavigator gracefully обрабатывает ошибки:
- Несуществующие поля → ошибка с fallback на старую логику
- Индекс за пределами массива → ошибка в логах
- Nil объекты → безопасная обработка с ошибкой
- Типы несовместимые с операцией → детальное сообщение об ошибке

#### Производительность

- O(1) доступ к полям map
- O(1) доступ к элементам массива
- Ленивая навигация (только по запрошенному пути)
- Нет парсинга до вызова (компиляция путей on-the-fly)

---

## 🚀 Преимущества

### До Output Mapping:
```javascript
// В следующей задаче приходилось писать:
=response.body.id
=response.body.name
=response.body.email
```

### С Output Mapping:
```javascript
// Можно просто использовать:
=userId
=userName
=userEmail
```

---

## ✅ Совместимость

- **Camunda 8**: ✅ Полная совместимость
- **Zeebe Protocol**: ✅ Стандарт Zeebe ioMapping
- **FEEL Path Expressions**: ✅ Полная поддержка через PathNavigator
- **HTTP Connector**: ✅ Реализовано с output mapping
- **Email Connector**: ⚠️ Требуется добавить output mapping (TODO)
- **Other Connectors**: ⚠️ Требуется добавить output mapping (TODO)

---

## 📝 Логирование

Output mapping логирует все операции включая PathNavigator:

```
2025-11-19 13:13:36 [DEBUG] Applying output mapping | element_id=PROXMOX_clone_vm
2025-11-19 13:13:36 [INFO ] Found output mappings | count=2
2025-11-19 13:13:36 [DEBUG] Processing output mapping | source==response.body.data target=cloneTaskId
2025-11-19 13:13:36 [DEBUG] Processing FEEL expression | expression==response.body.data
2025-11-19 13:13:36 [DEBUG] Path navigation successful | path=response.body.data result=UPID:pve3:... result_type=string
2025-11-19 13:13:36 [INFO ] Output mapping applied | source==response.body.data target=cloneTaskId value=UPID:pve3:...
2025-11-19 13:13:36 [DEBUG] Processing output mapping | source==response.status target=cloneStatus
2025-11-19 13:13:36 [DEBUG] Path navigation successful | path=response.status result=200 result_type=int
2025-11-19 13:13:36 [INFO ] Output mapping applied | source==response.status target=cloneStatus value=200
```

**Ключевые логи PathNavigator**:
- `Path navigation successful` - путь успешно навигирован
- `Path navigation failed` - ошибка навигации (с fallback)
- `Navigating path` - начало навигации
- `Path parsed into segments` - путь разобран на сегменты

---

## 🔧 Тестирование

### Unit Tests

PathNavigator покрыт полным набором unit tests:

```bash
# Запуск тестов PathNavigator
go test ./test/unit-test/expression/path_navigator_test.go -v

# Результат: 13/13 тестов прошли успешно
# Включая тесты:
# - Простой доступ к полям
# - Доступ к массивам
# - Вложенные массивы
# - Динамический доступ по переменным
# - Граничные случаи (nil, несуществующие поля)
```

### Интеграционное тестирование

Используйте реальный BPMN процесс:

```bash
# Парсинг BPMN с output mapping
./build/atomd bpmn parse bpmn_test/nocobase/create_clone_proxmox_with_output.bpmn

# Запуск процесса
./build/atomd process start Process_create_from_clone_output -d '{
  "api_url":"https://pve1.hlprod.ru:8006/api2/json",
  "proxmox_vm_id":"qemu/3013",
  "params":{"newid":"691699999","name":"test-vm","target":"pve3","full":"1","storage":"ceph-pool"}
}'

# Проверка переменных процесса
./build/atomd process info <instance_id>
# Ожидается: cloneTaskId и cloneStatus извлечены из response

# Проверка логов PathNavigator
tail -f build/logs/app.log | grep -i "path navigation"
```

---

## 📌 Важные замечания

1. **Output mapping выполняется ПОСЛЕ сохранения response**
   - Переменная `response` всегда доступна
   - Output mapping добавляет именованные переменные для удобства

2. **FEEL выражения обязательны**
   - Используйте `=response.body.field`, а не `response.body.field`
   - Без `=` значение будет воспринято как строка

3. **Ошибки не останавливают выполнение**
   - Если output mapping не сработал, логируется WARNING
   - Процесс продолжает выполнение

4. **Переменная response остается**
   - Output mapping НЕ удаляет переменную `response`
   - Это дополнительное удобство, не замена

---

## 🎯 Best Practices

1. **Извлекайте только нужные поля**
   ```xml
   <!-- Good -->
   <zeebe:output source="=response.body.id" target="userId" />
   
   <!-- Bad - слишком много -->
   <zeebe:output source="=response.body" target="allData" />
   ```

2. **Используйте понятные имена переменных**
   ```xml
   <!-- Good -->
   <zeebe:output source="=response.body.user.email" target="userEmail" />
   
   <!-- Bad - неясно -->
   <zeebe:output source="=response.body.user.email" target="x" />
   ```

3. **Проверяйте существование вложенных полей**
   ```xml
   <!-- Может вернуть null если user не существует -->
   <zeebe:output source="=response.body.user.email" target="userEmail" />
   ```

---

## 📚 Дополнительные ресурсы

- [Camunda 8 Documentation - Output Mapping](https://docs.camunda.io/docs/components/modeler/bpmn/service-tasks/#output-mappings)
- [FEEL Expressions Reference](https://docs.camunda.io/docs/components/modeler/feel/what-is-feel/)
- [HTTP Connector Documentation](docs/connectors/HTTP_CONNECTOR.md)

