# Expression Service

Сервис вычисления и анализа FEEL (Friendly Enough Expression Language) выражений в Atom Engine.

## Обзор

Expression Service предоставляет полный набор инструментов для работы с FEEL выражениями:
- Вычисление выражений с контекстом
- Массовая обработка выражений
- Парсинг в AST
- Валидация синтаксиса и семантики
- Извлечение переменных
- Тестирование с набором данных
- Справочник встроенных функций

## Методы сервиса

### Основные методы вычисления
- **[EvaluateExpression](evaluate-expression.md)** - Вычисление FEEL выражения
- **[EvaluateBatch](evaluate-batch.md)** - Пакетное вычисление выражений
- **[EvaluateCondition](evaluate-condition.md)** - Вычисление булевых условий

### Анализ и валидация
- **[ParseExpression](parse-expression.md)** - Парсинг в AST
- **[ValidateExpression](validate-expression.md)** - Валидация выражений
- **[ExtractVariables](extract-variables.md)** - Извлечение переменных

### Поддержка разработки
- **[GetSupportedFunctions](get-supported-functions.md)** - Список встроенных функций
- **[TestExpression](test-expression.md)** - Тестирование выражений

## Быстрый старт

### Go
```go
conn, _ := grpc.Dial("localhost:27500", grpc.WithInsecure())
client := expressionpb.NewExpressionServiceClient(conn)

ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key")

// Простое вычисление
result, _ := client.EvaluateExpression(ctx, &expressionpb.EvaluateExpressionRequest{
    Expression: "x + y * 2",
    Context:    `{"x": 10, "y": 5}`,
})

// Результат: "20"
```

### Python
```python
channel = grpc.insecure_channel('localhost:27500')
stub = expression_pb2_grpc.ExpressionServiceStub(channel)

response = stub.EvaluateExpression(
    expression_pb2.EvaluateExpressionRequest(
        expression="age >= 18",
        context='{"age": 25}'
    ),
    metadata=[('x-api-key', 'your-key')]
)

# response.result: "true"
```

### JavaScript
```javascript
const client = new expressionProto.ExpressionService('localhost:27500',
    grpc.credentials.createInsecure());

client.evaluateExpression({
    expression: 'upper(name)',
    context: JSON.stringify({name: 'john'})
}, metadata, callback);

// Результат: "JOHN"
```

## FEEL Синтаксис

### Базовые операторы
```feel
// Арифметика
x + y - z * 2 / 3

// Сравнения  
age >= 18 and status = "active"

// Логика
not (expired or suspended)
```

### Условия
```feel
if age >= 65 then "senior"
else if age >= 18 then "adult"  
else "minor"
```

### Функции
```feel
// Строки
upper("hello")           // "HELLO"
substring("hello", 2, 3) // "ell"

// Числа
abs(-5)                  // 5
round(3.14, 1)          // 3.1

// Списки
sum([1, 2, 3, 4])       // 10
count(items)            // количество элементов

// Даты
now()                   // текущая дата/время
```

### Работа с объектами
```feel
// Доступ к свойствам
user.profile.name
order.items[1].price

// Фильтрация
users[age > 18]
orders[status = "completed"]
```

## Категории функций

### String Functions
- `upper()`, `lower()` - регистр
- `substring()`, `length()` - работа с подстроками
- `contains()`, `matches()` - поиск и регулярные выражения

### Number Functions  
- `abs()`, `round()`, `floor()`, `ceil()` - математика
- `min()`, `max()`, `sum()` - агрегация

### List Functions
- `count()`, `append()` - размер и добавление
- `reverse()`, `sort()` - порядок
- `filter()` - фильтрация

### Date Functions
- `now()`, `today()` - текущие значения
- `date()`, `date and time()` - создание дат

## Применение в BPMN

### Gateway условия
```feel
// Exclusive Gateway
order.total > 1000

// Inclusive Gateway  
user.vip = true or order.priority = "high"
```

### Service Task выражения
```feel
// Входные переменные
{
  "customerId": user.id,
  "amount": order.total * 0.9
}

// Условия выполнения
user.verified = true and balance >= order.total
```

### Timer выражения
```feel
// Длительность на основе приоритета
if priority = "urgent" then "PT15M" 
else if priority = "high" then "PT1H"
else "P1D"
```

## Авторизация

Все методы требуют API ключ с разрешением `expression` или `*`:

```
Headers:
x-api-key: your-api-key-here
```

## Ошибки

### Частые ошибки
- **Синтаксическая ошибка**: `Syntax error at position 5`
- **Неизвестная переменная**: `Variable 'x' is not defined`
- **Неизвестная функция**: `Function 'unknownFunc' does not exist`
- **Несовместимость типов**: `Cannot compare number with string`

### Отладка
1. Используйте `ValidateExpression` для проверки синтаксиса
2. Проверьте переменные через `ExtractVariables`
3. Тестируйте с `TestExpression`

## Производительность

### Рекомендации
- Используйте `EvaluateBatch` для множественных вычислений
- Кэшируйте результаты `ParseExpression` для повторного использования
- Минимизируйте сложность выражений для критических путей

### Ограничения
- Максимальная длина выражения: 10,000 символов
- Максимальный размер контекста: 10MB
- Таймаут вычисления: 30 секунд

## Связанные компоненты

- **[Process Service](../process/README.md)** - использует Expression для BPMN логики
- **[Jobs Service](../jobs/README.md)** - вычисляет переменные заданий
- **[Messages Service](../messages/README.md)** - корреляционные ключи

## Примеры интеграции

### Валидация данных
```go
// Проверка бизнес-правил
rules := []string{
    "age >= 18",
    "email contains '@'", 
    "length(password) >= 8"
}

// Пакетная проверка
response, _ := client.EvaluateBatch(ctx, &expressionpb.EvaluateBatchRequest{
    Expressions: rules,
    Context: userData,
})
```

### Динамическое поведение
```python
# Расчет скидки на основе правил
discount_rule = """
if customer.vip_level = 'gold' then 0.2
else if customer.vip_level = 'silver' then 0.15  
else if order.total > 1000 then 0.1
else 0.05
"""

result = evaluate_expression(discount_rule, customer_context)
```

### IDE поддержка
```javascript
// Автодополнение функций
const functions = await getSupportedFunctions();
const functionNames = functions.map(f => f.name);

// Валидация в реальном времени
const validation = await validateExpression(userInput);
if (!validation.isValid) {
    showErrors(validation.errors);
}
```
