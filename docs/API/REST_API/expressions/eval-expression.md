# POST /api/v1/expressions/eval

## Описание
Вычисление FEEL выражения с предоставленным контекстом данных. Поддерживает полный синтаксис FEEL 1.1.

## URL
```
POST /api/v1/expressions/eval
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
- `expression` (string): FEEL выражение для вычисления

### Опциональные поля
- `context` (object): Контекст данных для вычисления (по умолчанию: {})
- `return_type` (string): Ожидаемый тип результата (`auto`, `boolean`, `number`, `string`, `object`)

## Примеры запросов

### Арифметическое выражение
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/eval" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "x + y * 2",
    "context": {
      "x": 10,
      "y": 5
    }
  }'
```

### Условное выражение
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/eval" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "if age >= 18 then \"adult\" else \"minor\"",
    "context": {
      "age": 25
    },
    "return_type": "string"
  }'
```

### Работа с объектами
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/eval" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "order.amount > 100 and order.status = \"pending\"",
    "context": {
      "order": {
        "amount": 299.99,
        "status": "pending",
        "customerId": "CUST-123"
      }
    }
  }'
```

### JavaScript
```javascript
const expression = {
  expression: 'sum(items[amount > 50].price)',
  context: {
    items: [
      { name: 'Product A', price: 25.99, amount: 30 },
      { name: 'Product B', price: 75.50, amount: 100 },
      { name: 'Product C', price: 120.00, amount: 75 }
    ]
  }
};

const response = await fetch('/api/v1/expressions/eval', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(expression)
});

const result = await response.json();
console.log('Result:', result.data.result);
```

## Ответы

### 200 OK - Выражение вычислено
```json
{
  "success": true,
  "data": {
    "expression": "x + y * 2",
    "result": 20,
    "result_type": "number",
    "context_used": {
      "x": 10,
      "y": 5
    },
    "evaluation_time_ms": 12,
    "variables_accessed": ["x", "y"],
    "functions_used": [],
    "is_valid": true
  },
  "request_id": "req_1641998403600"
}
```

### 200 OK - Сложное выражение
```json
{
  "success": true,
  "data": {
    "expression": "if order.amount > 100 then order.amount * 0.1 else 0",
    "result": 29.999,
    "result_type": "number",
    "context_used": {
      "order": {
        "amount": 299.99,
        "status": "pending"
      }
    },
    "evaluation_time_ms": 8,
    "variables_accessed": ["order.amount"],
    "functions_used": [],
    "is_valid": true,
    "description": "Conditional discount calculation: 10% for orders over $100"
  },
  "request_id": "req_1641998403601"
}
```

### 400 Bad Request - Синтаксическая ошибка
```json
{
  "success": false,
  "error": {
    "code": "EXPRESSION_SYNTAX_ERROR",
    "message": "Syntax error in FEEL expression",
    "details": {
      "expression": "x + y *",
      "error_position": 7,
      "error_description": "Unexpected end of expression",
      "expected": "operand",
      "suggestions": [
        "Complete the multiplication: x + y * z",
        "Remove the trailing operator: x + y"
      ]
    }
  },
  "request_id": "req_1641998403602"
}
```

### 400 Bad Request - Ошибка выполнения
```json
{
  "success": false,
  "error": {
    "code": "EXPRESSION_RUNTIME_ERROR",
    "message": "Runtime error during expression evaluation",
    "details": {
      "expression": "order.customer.name",
      "error_description": "Property 'customer' is null or undefined",
      "context_path": "order.customer",
      "available_properties": ["amount", "status", "customerId"],
      "suggestion": "Check if 'order.customer' exists or use safe navigation: order.customer?.name"
    }
  },
  "request_id": "req_1641998403603"
}
```

## Поля ответа

### Successful Evaluation
- `expression` (string): Исходное выражение
- `result` (any): Результат вычисления
- `result_type` (string): Тип результата
- `context_used` (object): Использованный контекст
- `evaluation_time_ms` (integer): Время вычисления в миллисекундах

### Analysis Information
- `variables_accessed` (array): Переменные, к которым был доступ
- `functions_used` (array): Использованные функции
- `is_valid` (boolean): Валидность выражения

### Error Information (при ошибках)
- `error_position` (integer): Позиция ошибки в выражении
- `error_description` (string): Описание ошибки
- `suggestions` (array): Предложения по исправлению

## FEEL Syntax Support

### Supported Operations
```yaml
Arithmetic:
  - "+", "-", "*", "/", "**" (power)
  - "%" (modulo)

Comparison:
  - "=", "!=", "<", "<=", ">", ">="
  - "between X and Y"
  - "in [list]"

Logical:
  - "and", "or", "not"

String:
  - "contains", "starts with", "ends with"
  - "matches" (regex)

List:
  - "[item1, item2, ...]"
  - "list[index]", "list[condition]"
  - "count(list)", "sum(list)", "mean(list)"

Date/Time:
  - "date(\"2025-01-11\")"
  - "time(\"10:30:00\")"
  - "duration(\"PT5M\")"
```

### Built-in Functions
```yaml
Math Functions:
  - abs(number)
  - ceiling(number)
  - floor(number)
  - round(number, digits)

String Functions:
  - upper(string)
  - lower(string)
  - substring(string, start, length)
  - replace(string, pattern, replacement)

List Functions:
  - count(list)
  - sum(list)
  - mean(list)
  - min(list)
  - max(list)
  - sort(list)
  - reverse(list)

Date Functions:
  - now()
  - today()
  - year(date)
  - month(date)
  - day(date)
```

## Использование

### Business Rules Engine
```javascript
class BusinessRulesEngine {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async evaluateOrderDiscount(orderData) {
    const rules = [
      {
        name: 'VIP Customer Discount',
        expression: 'customer.type = "VIP" and order.amount > 500',
        discount: 0.15
      },
      {
        name: 'Large Order Discount',
        expression: 'order.amount > 1000',
        discount: 0.10
      },
      {
        name: 'First Time Customer',
        expression: 'customer.orders_count = 0 and order.amount > 100',
        discount: 0.05
      }
    ];
    
    const context = { order: orderData.order, customer: orderData.customer };
    
    for (const rule of rules) {
      const evaluation = await this.evaluateExpression(rule.expression, context);
      
      if (evaluation.result === true) {
        return {
          applied_rule: rule.name,
          discount_percent: rule.discount * 100,
          discount_amount: orderData.order.amount * rule.discount,
          final_amount: orderData.order.amount * (1 - rule.discount)
        };
      }
    }
    
    return { applied_rule: null, discount_percent: 0 };
  }
  
  async evaluateExpression(expression, context) {
    const response = await fetch('/api/v1/expressions/eval', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({ expression, context })
    });
    
    return await response.json();
  }
}

// Использование
const engine = new BusinessRulesEngine('your-api-key');

const orderData = {
  order: { amount: 750, currency: 'USD' },
  customer: { type: 'VIP', orders_count: 5 }
};

const discount = await engine.evaluateOrderDiscount(orderData);
console.log('Applied discount:', discount);
```

### Dynamic Form Validation
```javascript
class DynamicValidator {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async validateForm(formData, validationRules) {
    const results = [];
    
    for (const rule of validationRules) {
      try {
        const evaluation = await fetch('/api/v1/expressions/eval', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': this.apiKey
          },
          body: JSON.stringify({
            expression: rule.expression,
            context: formData
          })
        });
        
        const result = await evaluation.json();
        
        results.push({
          field: rule.field,
          rule_name: rule.name,
          is_valid: result.data.result,
          error_message: result.data.result ? null : rule.error_message
        });
        
      } catch (error) {
        results.push({
          field: rule.field,
          rule_name: rule.name,
          is_valid: false,
          error_message: `Validation error: ${error.message}`
        });
      }
    }
    
    return {
      is_valid: results.every(r => r.is_valid),
      results
    };
  }
}

// Использование
const validator = new DynamicValidator('your-api-key');

const formData = {
  email: 'user@example.com',
  age: 25,
  password: 'secretpassword',
  confirmPassword: 'secretpassword'
};

const validationRules = [
  {
    field: 'email',
    name: 'Email Format',
    expression: 'matches(email, "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")',
    error_message: 'Invalid email format'
  },
  {
    field: 'age',
    name: 'Age Requirement',
    expression: 'age >= 18 and age <= 120',
    error_message: 'Age must be between 18 and 120'
  },
  {
    field: 'password',
    name: 'Password Strength',
    expression: 'string length(password) >= 8 and matches(password, ".*[A-Z].*") and matches(password, ".*[0-9].*")',
    error_message: 'Password must be at least 8 characters with uppercase and number'
  },
  {
    field: 'confirmPassword',
    name: 'Password Confirmation',
    expression: 'password = confirmPassword',
    error_message: 'Passwords do not match'
  }
];

const validation = await validator.validateForm(formData, validationRules);
console.log('Validation result:', validation);
```

### Decision Engine
```javascript
class DecisionEngine {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async makeCreditDecision(applicationData) {
    const decisionTable = [
      {
        condition: 'credit_score >= 750 and income >= 50000 and debt_ratio < 0.3',
        decision: 'APPROVED',
        interest_rate: 3.5,
        max_amount: 100000
      },
      {
        condition: 'credit_score >= 650 and income >= 30000 and debt_ratio < 0.4',
        decision: 'APPROVED',
        interest_rate: 5.5,
        max_amount: 50000
      },
      {
        condition: 'credit_score >= 600 and income >= 25000 and debt_ratio < 0.5',
        decision: 'CONDITIONAL',
        interest_rate: 8.5,
        max_amount: 25000,
        conditions: ['Require co-signer', 'Additional documentation needed']
      },
      {
        condition: 'credit_score < 600 or income < 25000 or debt_ratio >= 0.5',
        decision: 'DECLINED',
        reason: 'Does not meet minimum requirements'
      }
    ];
    
    for (const row of decisionTable) {
      const evaluation = await this.evaluateExpression(row.condition, applicationData);
      
      if (evaluation.data.result === true) {
        return {
          decision: row.decision,
          interest_rate: row.interest_rate,
          max_amount: row.max_amount,
          conditions: row.conditions || [],
          reason: row.reason || null,
          applied_rule: row.condition
        };
      }
    }
    
    return {
      decision: 'DECLINED',
      reason: 'No matching decision rule found'
    };
  }
  
  async evaluateExpression(expression, context) {
    const response = await fetch('/api/v1/expressions/eval', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({ expression, context })
    });
    
    return await response.json();
  }
}

// Использование
const decisionEngine = new DecisionEngine('your-api-key');

const application = {
  credit_score: 720,
  income: 65000,
  debt_ratio: 0.25,
  employment_years: 3,
  requested_amount: 40000
};

const decision = await decisionEngine.makeCreditDecision(application);
console.log('Credit decision:', decision);
```

### Expression Caching
```javascript
class CachedExpressionEvaluator {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.cache = new Map();
    this.cacheMaxSize = 1000;
    this.cacheTTL = 300000; // 5 minutes
  }
  
  async evaluate(expression, context) {
    const cacheKey = this.generateCacheKey(expression, context);
    const cached = this.cache.get(cacheKey);
    
    if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
      console.log('Cache hit for expression');
      return cached.result;
    }
    
    const result = await this.evaluateExpression(expression, context);
    
    // Cache successful results
    if (result.success) {
      this.cache.set(cacheKey, {
        result,
        timestamp: Date.now()
      });
      
      // Maintain cache size
      if (this.cache.size > this.cacheMaxSize) {
        const firstKey = this.cache.keys().next().value;
        this.cache.delete(firstKey);
      }
    }
    
    return result;
  }
  
  generateCacheKey(expression, context) {
    return `${expression}:${JSON.stringify(context)}`;
  }
  
  async evaluateExpression(expression, context) {
    const response = await fetch('/api/v1/expressions/eval', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({ expression, context })
    });
    
    return await response.json();
  }
  
  clearCache() {
    this.cache.clear();
  }
  
  getCacheStats() {
    return {
      size: this.cache.size,
      maxSize: this.cacheMaxSize,
      utilization: (this.cache.size / this.cacheMaxSize) * 100
    };
  }
}
```

## Производительность

### Optimization Tips
1. **Кэширование**: Кэшируйте результаты для повторяющихся выражений
2. **Контекст**: Передавайте только необходимые данные в контексте
3. **Простота**: Избегайте излишне сложных выражений
4. **Предварительная валидация**: Используйте `/validate` для проверки синтаксиса

### Performance Benchmarks
- **Простые арифметические**: < 5ms
- **Условные выражения**: < 10ms
- **Работа с объектами**: < 15ms
- **Сложные вычисления**: < 50ms

## Связанные endpoints
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация синтаксиса
- [`POST /api/v1/expressions/parse`](./parse-expression.md) - Парсинг в AST
- [`GET /api/v1/expressions/functions`](./list-functions.md) - Список функций
- [`POST /api/v1/expressions/test`](./test-expression.md) - Тестирование выражений
