# Date Functions Usage Guide

## Описание

Руководство по использованию функций для работы с датами и временем в FEEL expressions. Включает функции `duration()`, `subtract()` и `add()` для выполнения арифметических операций с датами.

## Доступные функции

### duration(string)

Парсит строку длительности в формате ISO 8601 и возвращает её для использования в других функциях.

**Синтаксис:**
```feel
duration("P3D")
duration("PT1H30M")
duration("P1DT2H30M15S")
```

**Параметры:**
- `string` (string, обязательный): Строка длительности в формате ISO 8601

**Возвращает:**
- `string`: Строка длительности в формате ISO 8601

**Примеры:**

```feel
duration("P3D")           // 3 дня
duration("PT1H")          // 1 час
duration("PT30M")         // 30 минут
duration("PT1H30M")       // 1 час 30 минут
duration("P1DT2H30M")     // 1 день 2 часа 30 минут
duration("P1W")           // 1 неделя (7 дней)
```

**Формат ISO 8601:**
- `P` - префикс периода
- `Y` - годы
- `M` - месяцы (в части даты)
- `D` - дни
- `T` - разделитель времени
- `H` - часы
- `M` - минуты (в части времени)
- `S` - секунды

**Примеры форматов:**
- `P3D` - 3 дня
- `PT5M` - 5 минут
- `PT1H30M` - 1 час 30 минут
- `P1DT2H30M15S` - 1 день 2 часа 30 минут 15 секунд
- `P1Y2M3DT4H5M6S` - 1 год 2 месяца 3 дня 4 часа 5 минут 6 секунд

---

### subtract(datetime, duration)

Вычитает длительность из даты-времени и возвращает новую дату-время.

**Синтаксис:**
```feel
subtract(datetime, duration)
subtract("2025-12-13T12:18:19.675Z", duration("P3D"))
subtract(datetime, duration("P3D"))
```

**Параметры:**
- `datetime` (string, обязательный): Дата-время в формате ISO 8601 (например, `2025-12-13T12:18:19.675Z`)
- `duration` (string, обязательный): Длительность, полученная через функцию `duration()`

**Возвращает:**
- `string`: Новая дата-время в формате ISO 8601

**Примеры:**

```feel
// Вычитание 3 дней
subtract("2025-12-13T12:18:19.675Z", duration("P3D"))
// Результат: "2025-12-10T12:18:19.675Z"

// Вычитание с переходом через месяц
subtract("2025-12-02T12:18:19.675Z", duration("P3D"))
// Результат: "2025-11-29T12:18:19.675Z"

// Вычитание часов
subtract("2025-12-13T15:30:00.000Z", duration("PT2H"))
// Результат: "2025-12-13T13:30:00.000Z"

// Вычитание дней и часов
subtract("2025-12-13T15:30:00.000Z", duration("P1DT2H"))
// Результат: "2025-12-12T13:30:00.000Z"

// Использование переменной
subtract(datetime, duration("P7D"))
// Если datetime = "2025-12-13T12:18:19.675Z"
// Результат: "2025-12-06T12:18:19.675Z"
```

**Особенности:**
- Автоматически учитывает переходы через месяцы и годы
- Сохраняет время (часы, минуты, секунды, миллисекунды)
- Поддерживает все форматы ISO 8601 для длительности

---

### add(datetime, duration)

Добавляет длительность к дате-времени и возвращает новую дату-время.

**Синтаксис:**
```feel
add(datetime, duration)
add("2025-12-13T12:18:19.675Z", duration("P1D"))
add(datetime, duration("P1D"))
```

**Параметры:**
- `datetime` (string, обязательный): Дата-время в формате ISO 8601 (например, `2025-12-13T12:18:19.675Z`)
- `duration` (string, обязательный): Длительность, полученная через функцию `duration()`

**Возвращает:**
- `string`: Новая дата-время в формате ISO 8601

**Примеры:**

```feel
// Добавление 1 дня
add("2025-12-13T12:18:19.675Z", duration("P1D"))
// Результат: "2025-12-14T12:18:19.675Z"

// Добавление с переходом через месяц
add("2025-12-30T12:18:19.675Z", duration("P3D"))
// Результат: "2026-01-02T12:18:19.675Z"

// Добавление часов
add("2025-12-13T15:30:00.000Z", duration("PT2H"))
// Результат: "2025-12-13T17:30:00.000Z"

// Добавление дней и часов
add("2025-12-13T15:30:00.000Z", duration("P1DT2H"))
// Результат: "2025-12-14T17:30:00.000Z"

// Использование переменной
add(datetime, duration("P7D"))
// Если datetime = "2025-12-13T12:18:19.675Z"
// Результат: "2025-12-20T12:18:19.675Z"
```

**Особенности:**
- Автоматически учитывает переходы через месяцы и годы
- Сохраняет время (часы, минуты, секунды, миллисекунды)
- Поддерживает все форматы ISO 8601 для длительности

---

## Вложенные вызовы функций

Функции поддерживают вложенные вызовы, что позволяет создавать сложные выражения:

```feel
// Вычитание 3 дней из даты
subtract("2025-12-13T12:18:19.675Z", duration("P3D"))

// Вычитание 1 недели (7 дней)
subtract("2025-12-13T12:18:19.675Z", duration("P7D"))

// Добавление 1 месяца и 2 дней
add("2025-12-13T12:18:19.675Z", duration("P1M2D"))
```

## Использование с переменными

Функции могут работать с переменными из контекста:

```json
{
  "datetime": "2025-12-13T12:18:19.675Z",
  "days_to_subtract": 3
}
```

```feel
// Использование переменной datetime
subtract(datetime, duration("P3D"))

// Можно комбинировать с другими выражениями
subtract(datetime, duration("P" + string(days_to_subtract) + "D"))
```

## Примеры использования в BPMN

### Timer выражения

```xml
<timerEventDefinition>
  <timeDuration>subtract(datetime, duration("P3D"))</timeDuration>
</timerEventDefinition>
```

### Service Task переменные

```json
{
  "dueDate": "subtract(now(), duration(\"P7D\"))",
  "startDate": "add(createdDate, duration(\"P1D\"))"
}
```

### Gateway условия

```feel
// Проверка, прошло ли 3 дня с момента создания
subtract(now(), duration("P3D")) > createdDate
```

## Формат дат

Все функции работают с датами в формате ISO 8601:

**Поддерживаемые форматы:**
- `2025-12-13T12:18:19.675Z` - с миллисекундами и UTC
- `2025-12-13T12:18:19Z` - без миллисекунд, с UTC
- `2025-12-13T12:18:19+03:00` - с часовым поясом

**Выходной формат:**
- Всегда возвращается в формате `RFC3339Nano` с миллисекундами и UTC: `2025-12-13T12:18:19.675Z`

## Примеры через REST API

### Вычисление выражения с вычитанием дат

```bash
curl -X POST "http://localhost:27555/api/v1/expressions/evaluate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "subtract(\"2025-12-13T12:18:19.675Z\", duration(\"P3D\"))",
    "context": {}
  }'
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "result": "2025-12-10T12:18:19.675Z",
    "result_type": "string"
  }
}
```

### Вычисление с переменными

```bash
curl -X POST "http://localhost:27555/api/v1/expressions/evaluate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "subtract(datetime, duration(\"P3D\"))",
    "context": {
      "datetime": "2025-12-13T12:18:19.675Z"
    }
  }'
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "result": "2025-12-10T12:18:19.675Z",
    "result_type": "string"
  }
}
```

### Добавление длительности

```bash
curl -X POST "http://localhost:27555/api/v1/expressions/evaluate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "add(\"2025-12-13T12:18:19.675Z\", duration(\"P1D\"))",
    "context": {}
  }'
```

**Ответ:**
```json
{
  "success": true,
  "data": {
    "result": "2025-12-14T12:18:19.675Z",
    "result_type": "string"
  }
}
```

## Типичные сценарии использования

### 1. Вычисление даты окончания срока

```feel
// Дата создания + 30 дней
add(createdDate, duration("P30D"))

// Дата создания + 1 месяц
add(createdDate, duration("P1M"))
```

### 2. Вычисление даты начала периода

```feel
// Текущая дата - 7 дней (неделя назад)
subtract(now(), duration("P7D"))

// Дата события - 3 дня
subtract(eventDate, duration("P3D"))
```

### 3. Вычисление интервалов

```feel
// Дата начала + 2 часа
add(startTime, duration("PT2H"))

// Дата окончания - 30 минут
subtract(endTime, duration("PT30M"))
```

### 4. Работа с дедлайнами

```feel
// Дедлайн = дата создания + срок выполнения
add(createdDate, duration("P" + string(deadlineDays) + "D"))

// Проверка, осталось ли менее 3 дней
subtract(deadline, duration("P3D")) < now()
```

## Ошибки и валидация

### Типичные ошибки

1. **Неверный формат даты:**
   ```
   Error: invalid datetime format: invalid ISO 8601 datetime format: 2025-12-13
   ```
   Решение: Используйте полный формат ISO 8601 с временем: `2025-12-13T12:18:19.675Z`

2. **Неверный формат длительности:**
   ```
   Error: invalid duration format: invalid ISO8601 duration format: 3D
   ```
   Решение: Используйте формат ISO 8601: `P3D` вместо `3D`

3. **Неверное количество аргументов:**
   ```
   Error: subtract() requires exactly 2 arguments, got 1
   ```
   Решение: Убедитесь, что передаете оба аргумента: `subtract(datetime, duration)`

### Валидация выражений

Перед использованием выражений рекомендуется валидировать их:

```bash
curl -X POST "http://localhost:27555/api/v1/expressions/validate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "subtract(datetime, duration(\"P3D\"))"
  }'
```

## Связанные функции

- **[EvaluateExpression](eval-expression.md)** - Вычисление выражений
- **[ListFunctions](list-functions.md)** - Список всех доступных функций
- **[ValidateExpression](validate-expression.md)** - Валидация выражений

## Дополнительные ресурсы

- [ISO 8601 Duration Format](https://en.wikipedia.org/wiki/ISO_8601#Durations)
- [FEEL Expression Language Documentation](https://camunda.github.io/feel-scala/)

