# POST /api/v1/expressions/validate

## Описание
Валидация синтаксиса FEEL выражения без его выполнения. Проверяет корректность синтаксиса и доступность функций.

## URL
```
POST /api/v1/expressions/validate
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression`

## Заголовки запроса
```http
Content-Type: application/json
Accept: application/json
X-API-Key: your-api-key-here
```

## Параметры тела запроса

### Обязательные поля
- `expression` (string): FEEL выражение для валидации

### Опциональные поля
- `schema` (object): Схема контекста для валидации переменных
- `strict_mode` (boolean): Строгий режим валидации (по умолчанию: false)

## Примеры запросов

### Простая валидация
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/validate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "x + y > 100"
  }'
```

### Валидация со схемой
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/validate" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "order.amount * tax_rate",
    "schema": {
      "order": {
        "type": "object",
        "properties": {
          "amount": {"type": "number"},
          "currency": {"type": "string"}
        }
      },
      "tax_rate": {"type": "number"}
    },
    "strict_mode": true
  }'
```

### JavaScript
```javascript
const validation = {
  expression: 'if customer.type = "VIP" then discount_rate * 2 else discount_rate',
  schema: {
    customer: {
      type: 'object',
      properties: {
        type: { type: 'string', enum: ['REGULAR', 'VIP', 'PREMIUM'] },
        id: { type: 'string' }
      }
    },
    discount_rate: { type: 'number', minimum: 0, maximum: 1 }
  },
  strict_mode: true
};

const response = await fetch('/api/v1/expressions/validate', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(validation)
});

const result = await response.json();
```

## Ответы

### 200 OK - Выражение валидно
```json
{
  "success": true,
  "data": {
    "expression": "x + y > 100",
    "is_valid": true,
    "syntax_correct": true,
    "variables_found": ["x", "y"],
    "functions_used": [],
    "return_type": "boolean",
    "complexity_score": 3,
    "validation_time_ms": 5,
    "warnings": [],
    "suggestions": []
  },
  "request_id": "req_1641998403700"
}
```

### 200 OK - Валидация со схемой
```json
{
  "success": true,
  "data": {
    "expression": "order.amount * tax_rate",
    "is_valid": true,
    "syntax_correct": true,
    "schema_valid": true,
    "variables_found": ["order.amount", "tax_rate"],
    "functions_used": [],
    "return_type": "number",
    "complexity_score": 2,
    "validation_time_ms": 8,
    "schema_validation": {
      "all_variables_defined": true,
      "type_compatibility": "valid",
      "missing_variables": [],
      "undefined_variables": []
    },
    "warnings": [],
    "suggestions": [
      "Consider adding null checks for object properties"
    ]
  },
  "request_id": "req_1641998403701"
}
```

### 200 OK - Невалидное выражение
```json
{
  "success": true,
  "data": {
    "expression": "x + y *",
    "is_valid": false,
    "syntax_correct": false,
    "errors": [
      {
        "type": "SYNTAX_ERROR",
        "message": "Unexpected end of expression",
        "position": 7,
        "severity": "ERROR",
        "suggestion": "Complete the multiplication operation"
      }
    ],
    "variables_found": ["x", "y"],
    "functions_used": [],
    "complexity_score": 0,
    "validation_time_ms": 3
  },
  "request_id": "req_1641998403702"
}
```

### 200 OK - Ошибки схемы
```json
{
  "success": true,
  "data": {
    "expression": "order.price + unknown_variable",
    "is_valid": false,
    "syntax_correct": true,
    "schema_valid": false,
    "variables_found": ["order.price", "unknown_variable"],
    "functions_used": [],
    "errors": [
      {
        "type": "SCHEMA_ERROR",
        "message": "Variable 'unknown_variable' is not defined in schema",
        "variable": "unknown_variable",
        "severity": "ERROR"
      },
      {
        "type": "SCHEMA_ERROR",
        "message": "Property 'price' not found in object 'order'",
        "variable": "order.price",
        "available_properties": ["amount", "currency"],
        "severity": "ERROR",
        "suggestion": "Use 'order.amount' instead of 'order.price'"
      }
    ],
    "schema_validation": {
      "all_variables_defined": false,
      "missing_variables": ["unknown_variable"],
      "undefined_variables": ["order.price"]
    },
    "validation_time_ms": 12
  },
  "request_id": "req_1641998403703"
}
```

## Поля ответа

### Validation Result
- `expression` (string): Исходное выражение
- `is_valid` (boolean): Общая валидность
- `syntax_correct` (boolean): Корректность синтаксиса
- `schema_valid` (boolean): Соответствие схеме (если предоставлена)

### Analysis Information
- `variables_found` (array): Найденные переменные
- `functions_used` (array): Использованные функции
- `return_type` (string): Предполагаемый тип возвращаемого значения
- `complexity_score` (integer): Оценка сложности (0-10)

### Error Information
- `errors` (array): Список ошибок валидации
- `warnings` (array): Предупреждения
- `suggestions` (array): Предложения по улучшению

### Schema Validation (если применимо)
- `all_variables_defined` (boolean): Все переменные определены в схеме
- `missing_variables` (array): Переменные не найденные в схеме
- `undefined_variables` (array): Неопределенные свойства объектов

## Типы ошибок

### Syntax Errors
- `SYNTAX_ERROR` - Синтаксическая ошибка
- `UNEXPECTED_TOKEN` - Неожиданный токен
- `MISSING_OPERAND` - Отсутствует операнд
- `MISSING_OPERATOR` - Отсутствует оператор
- `UNMATCHED_PARENTHESES` - Несовпадающие скобки

### Schema Errors
- `SCHEMA_ERROR` - Ошибка схемы
- `UNDEFINED_VARIABLE` - Неопределенная переменная
- `TYPE_MISMATCH` - Несоответствие типов
- `MISSING_PROPERTY` - Отсутствующее свойство

### Function Errors
- `UNKNOWN_FUNCTION` - Неизвестная функция
- `INVALID_ARGUMENTS` - Неверные аргументы функции
- `ARGUMENT_COUNT_MISMATCH` - Неверное количество аргументов

## Использование

### Expression Builder Validation
```javascript
class ExpressionBuilder {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.schema = null;
  }
  
  setSchema(schema) {
    this.schema = schema;
  }
  
  async validateExpression(expression) {
    const request = {
      expression,
      strict_mode: true
    };
    
    if (this.schema) {
      request.schema = this.schema;
    }
    
    const response = await fetch('/api/v1/expressions/validate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify(request)
    });
    
    const result = await response.json();
    
    return {
      isValid: result.data.is_valid,
      errors: result.data.errors || [],
      warnings: result.data.warnings || [],
      suggestions: result.data.suggestions || [],
      complexity: result.data.complexity_score,
      returnType: result.data.return_type
    };
  }
  
  async validateRealTime(expression, onValidation) {
    if (this.validationTimer) {
      clearTimeout(this.validationTimer);
    }
    
    // Debounce validation
    this.validationTimer = setTimeout(async () => {
      try {
        const validation = await this.validateExpression(expression);
        onValidation(validation);
      } catch (error) {
        onValidation({
          isValid: false,
          errors: [{ message: error.message, type: 'VALIDATION_ERROR' }]
        });
      }
    }, 300);
  }
}

// Использование
const builder = new ExpressionBuilder('your-api-key');

builder.setSchema({
  order: {
    type: 'object',
    properties: {
      amount: { type: 'number' },
      currency: { type: 'string' },
      customer: {
        type: 'object',
        properties: {
          type: { type: 'string', enum: ['REGULAR', 'VIP'] }
        }
      }
    }
  }
});

// Валидация в реальном времени
builder.validateRealTime('order.amount > 100', (validation) => {
  if (validation.isValid) {
    console.log('Expression is valid');
  } else {
    console.log('Errors:', validation.errors);
  }
});
```

### Form Validation Helper
```javascript
class ExpressionFormValidator {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async validateFormExpressions(formExpressions, schema) {
    const validations = await Promise.all(
      formExpressions.map(expr => this.validateSingle(expr, schema))
    );
    
    return {
      allValid: validations.every(v => v.isValid),
      validations,
      summary: this.createValidationSummary(validations)
    };
  }
  
  async validateSingle(expression, schema) {
    try {
      const response = await fetch('/api/v1/expressions/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify({
          expression: expression.text,
          schema,
          strict_mode: true
        })
      });
      
      const result = await response.json();
      
      return {
        id: expression.id,
        expression: expression.text,
        isValid: result.data.is_valid,
        errors: result.data.errors || [],
        warnings: result.data.warnings || [],
        returnType: result.data.return_type,
        complexity: result.data.complexity_score
      };
      
    } catch (error) {
      return {
        id: expression.id,
        expression: expression.text,
        isValid: false,
        errors: [{ message: error.message, type: 'NETWORK_ERROR' }]
      };
    }
  }
  
  createValidationSummary(validations) {
    return {
      total: validations.length,
      valid: validations.filter(v => v.isValid).length,
      invalid: validations.filter(v => !v.isValid).length,
      warnings: validations.reduce((sum, v) => sum + (v.warnings?.length || 0), 0),
      avgComplexity: validations.reduce((sum, v) => sum + (v.complexity || 0), 0) / validations.length
    };
  }
}
```

### Expression IDE Integration
```javascript
class ExpressionIDE {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.editor = null;
    this.validationDelay = 500;
  }
  
  initializeEditor(editorElement, schema) {
    this.editor = editorElement;
    this.schema = schema;
    
    // Real-time validation
    this.editor.addEventListener('input', (event) => {
      this.scheduleValidation(event.target.value);
    });
    
    // Syntax highlighting
    this.setupSyntaxHighlighting();
  }
  
  scheduleValidation(expression) {
    if (this.validationTimer) {
      clearTimeout(this.validationTimer);
    }
    
    this.validationTimer = setTimeout(() => {
      this.validateAndShowResults(expression);
    }, this.validationDelay);
  }
  
  async validateAndShowResults(expression) {
    if (!expression.trim()) {
      this.clearValidationResults();
      return;
    }
    
    try {
      const response = await fetch('/api/v1/expressions/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify({
          expression,
          schema: this.schema,
          strict_mode: true
        })
      });
      
      const result = await response.json();
      this.displayValidationResults(result.data);
      
    } catch (error) {
      this.displayError(error.message);
    }
  }
  
  displayValidationResults(validation) {
    // Clear previous results
    this.clearValidationResults();
    
    // Show validation status
    const statusElement = document.getElementById('validation-status');
    statusElement.className = validation.is_valid ? 'valid' : 'invalid';
    statusElement.textContent = validation.is_valid ? '✓ Valid' : '✗ Invalid';
    
    // Show errors
    if (validation.errors && validation.errors.length > 0) {
      const errorsElement = document.getElementById('validation-errors');
      errorsElement.innerHTML = validation.errors
        .map(error => `<div class="error">${error.message}</div>`)
        .join('');
    }
    
    // Show warnings
    if (validation.warnings && validation.warnings.length > 0) {
      const warningsElement = document.getElementById('validation-warnings');
      warningsElement.innerHTML = validation.warnings
        .map(warning => `<div class="warning">${warning}</div>`)
        .join('');
    }
    
    // Show suggestions
    if (validation.suggestions && validation.suggestions.length > 0) {
      const suggestionsElement = document.getElementById('validation-suggestions');
      suggestionsElement.innerHTML = validation.suggestions
        .map(suggestion => `<div class="suggestion">${suggestion}</div>`)
        .join('');
    }
    
    // Show metadata
    const metadataElement = document.getElementById('expression-metadata');
    metadataElement.innerHTML = `
      <div>Return Type: ${validation.return_type || 'unknown'}</div>
      <div>Complexity: ${validation.complexity_score || 0}/10</div>
      <div>Variables: ${validation.variables_found?.join(', ') || 'none'}</div>
      <div>Functions: ${validation.functions_used?.join(', ') || 'none'}</div>
    `;
  }
  
  clearValidationResults() {
    document.getElementById('validation-errors').innerHTML = '';
    document.getElementById('validation-warnings').innerHTML = '';
    document.getElementById('validation-suggestions').innerHTML = '';
  }
  
  setupSyntaxHighlighting() {
    // Implement syntax highlighting for FEEL expressions
    // This would involve parsing tokens and applying CSS classes
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Выполнение валидного выражения
- [`POST /api/v1/expressions/parse`](./parse-expression.md) - Парсинг в AST
- [`GET /api/v1/expressions/functions`](./list-functions.md) - Список доступных функций
- [`POST /api/v1/expressions/test`](./test-expression.md) - Тестирование с тест-кейсами
