# GET /api/v1/expressions/functions

## Описание
Получение списка поддерживаемых функций в FEEL engine с их описаниями, параметрами и примерами использования.

## URL
```
GET /api/v1/expressions/functions
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression`

## Параметры запроса (Query Parameters)

### Фильтрация
- `category` (string): Категория функций (`math`, `string`, `list`, `date`, `logical`, `conversion`)
- `search` (string): Поиск по имени или описанию функции
- `include_examples` (boolean): Включить примеры использования (по умолчанию: true)

## Примеры запросов

### Все функции
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/functions" \
  -H "X-API-Key: your-api-key-here"
```

### Математические функции
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/functions?category=math" \
  -H "X-API-Key: your-api-key-here"
```

### Поиск функций
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/functions?search=sum" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/expressions/functions?category=string&include_examples=true', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const functions = await response.json();
console.log('String functions:', functions.data);
```

## Ответы

### 200 OK - Список функций
```json
{
  "success": true,
  "data": {
    "functions": [
      {
        "name": "sum",
        "category": "math",
        "description": "Returns the sum of all numbers in a list",
        "signature": "sum(list)",
        "parameters": [
          {
            "name": "list",
            "type": "list<number>",
            "description": "List of numbers to sum",
            "required": true
          }
        ],
        "return_type": "number",
        "examples": [
          {
            "expression": "sum([1, 2, 3, 4])",
            "result": 10,
            "description": "Sum of list elements"
          },
          {
            "expression": "sum(order.items[].price)",
            "context": {
              "order": {
                "items": [
                  {"price": 25.99},
                  {"price": 15.50},
                  {"price": 8.75}
                ]
              }
            },
            "result": 50.24,
            "description": "Sum of order item prices"
          }
        ]
      },
      {
        "name": "substring",
        "category": "string",
        "description": "Extracts a substring from a string",
        "signature": "substring(string, start, length?)",
        "parameters": [
          {
            "name": "string",
            "type": "string",
            "description": "The source string",
            "required": true
          },
          {
            "name": "start",
            "type": "number",
            "description": "Starting position (1-based)",
            "required": true
          },
          {
            "name": "length",
            "type": "number",
            "description": "Length of substring (optional)",
            "required": false
          }
        ],
        "return_type": "string",
        "examples": [
          {
            "expression": "substring(\"Hello World\", 7)",
            "result": "World",
            "description": "Substring from position 7 to end"
          },
          {
            "expression": "substring(\"Hello World\", 1, 5)",
            "result": "Hello",
            "description": "Substring with specified length"
          }
        ]
      },
      {
        "name": "count",
        "category": "list",
        "description": "Returns the number of elements in a list",
        "signature": "count(list)",
        "parameters": [
          {
            "name": "list",
            "type": "list",
            "description": "List to count elements",
            "required": true
          }
        ],
        "return_type": "number",
        "examples": [
          {
            "expression": "count([1, 2, 3])",
            "result": 3,
            "description": "Count list elements"
          },
          {
            "expression": "count(customers[age > 18])",
            "context": {
              "customers": [
                {"age": 25}, {"age": 17}, {"age": 30}
              ]
            },
            "result": 2,
            "description": "Count filtered elements"
          }
        ]
      },
      {
        "name": "now",
        "category": "date",
        "description": "Returns the current date and time",
        "signature": "now()",
        "parameters": [],
        "return_type": "date-time",
        "examples": [
          {
            "expression": "now()",
            "result": "2025-01-11T10:30:00.000Z",
            "description": "Current timestamp"
          },
          {
            "expression": "now() > order.created_at",
            "context": {
              "order": {
                "created_at": "2025-01-10T15:30:00.000Z"
              }
            },
            "result": true,
            "description": "Compare with order creation time"
          }
        ]
      }
    ],
    "categories": [
      {
        "name": "math",
        "description": "Mathematical operations and calculations",
        "function_count": 12,
        "functions": ["abs", "ceiling", "floor", "round", "sum", "mean", "min", "max", "sqrt", "power", "mod", "random"]
      },
      {
        "name": "string",
        "description": "String manipulation and processing",
        "function_count": 10,
        "functions": ["upper", "lower", "substring", "contains", "starts with", "ends with", "replace", "matches", "split", "join"]
      },
      {
        "name": "list",
        "description": "List operations and transformations",
        "function_count": 8,
        "functions": ["count", "sum", "mean", "min", "max", "sort", "reverse", "distinct"]
      },
      {
        "name": "date",
        "description": "Date and time operations",
        "function_count": 15,
        "functions": ["now", "today", "date", "time", "duration", "year", "month", "day", "hour", "minute", "second", "add", "subtract", "format", "parse"]
      },
      {
        "name": "logical",
        "description": "Logical operations and conditions",
        "function_count": 5,
        "functions": ["and", "or", "not", "if", "all"]
      },
      {
        "name": "conversion",
        "description": "Type conversion functions",
        "function_count": 6,
        "functions": ["string", "number", "boolean", "date", "time", "duration"]
      }
    ],
    "summary": {
      "total_functions": 56,
      "categories_count": 6,
      "most_used_category": "string",
      "newest_functions": ["random", "format", "parse"]
    }
  },
  "request_id": "req_1641998403900"
}
```

### 200 OK - Фильтрованный список (математические функции)
```json
{
  "success": true,
  "data": {
    "functions": [
      {
        "name": "abs",
        "category": "math",
        "description": "Returns the absolute value of a number",
        "signature": "abs(number)",
        "parameters": [
          {
            "name": "number",
            "type": "number",
            "description": "The number to get absolute value",
            "required": true
          }
        ],
        "return_type": "number",
        "examples": [
          {
            "expression": "abs(-5)",
            "result": 5,
            "description": "Absolute value of negative number"
          },
          {
            "expression": "abs(balance)",
            "context": {"balance": -150.50},
            "result": 150.5,
            "description": "Absolute value of account balance"
          }
        ]
      },
      {
        "name": "round",
        "category": "math",
        "description": "Rounds a number to specified decimal places",
        "signature": "round(number, digits?)",
        "parameters": [
          {
            "name": "number",
            "type": "number",
            "description": "The number to round",
            "required": true
          },
          {
            "name": "digits",
            "type": "number",
            "description": "Number of decimal places (default: 0)",
            "required": false
          }
        ],
        "return_type": "number",
        "examples": [
          {
            "expression": "round(3.14159)",
            "result": 3,
            "description": "Round to nearest integer"
          },
          {
            "expression": "round(3.14159, 2)",
            "result": 3.14,
            "description": "Round to 2 decimal places"
          }
        ]
      }
    ],
    "filter_applied": {
      "category": "math",
      "total_matching": 12
    }
  }
}
```

## Function Categories

### Math Functions
```yaml
Basic Operations:
  - abs(number) - Absolute value
  - ceiling(number) - Round up to integer
  - floor(number) - Round down to integer
  - round(number, digits?) - Round to decimal places
  - mod(dividend, divisor) - Modulo operation

Aggregation:
  - sum(list) - Sum of numbers
  - mean(list) - Average of numbers  
  - min(list) - Minimum value
  - max(list) - Maximum value

Advanced:
  - sqrt(number) - Square root
  - power(base, exponent) - Exponentiation
  - random() - Random number 0-1
```

### String Functions
```yaml
Case Conversion:
  - upper(string) - Convert to uppercase
  - lower(string) - Convert to lowercase

Substring Operations:
  - substring(string, start, length?) - Extract substring
  - contains(string, substring) - Check if contains
  - starts with(string, prefix) - Check if starts with
  - ends with(string, suffix) - Check if ends with

Pattern Matching:
  - matches(string, pattern) - Regex match
  - replace(string, pattern, replacement) - Replace text

List Operations:
  - split(string, delimiter) - Split into list
  - join(list, delimiter) - Join list into string
```

### List Functions
```yaml
Aggregation:
  - count(list) - Number of elements
  - sum(list) - Sum of numbers
  - mean(list) - Average value
  - min(list) - Minimum value
  - max(list) - Maximum value

Transformation:
  - sort(list) - Sort elements
  - reverse(list) - Reverse order
  - distinct(list) - Remove duplicates
```

### Date Functions
```yaml
Current Time:
  - now() - Current date and time
  - today() - Current date

Constructors:
  - date(year, month, day) - Create date
  - time(hour, minute, second?) - Create time
  - duration(string) - Parse ISO 8601 duration

Extractors:
  - year(date) - Extract year
  - month(date) - Extract month
  - day(date) - Extract day
  - hour(datetime) - Extract hour
  - minute(datetime) - Extract minute
  - second(datetime) - Extract second

Operations:
  - add(datetime, duration) - Add duration
  - subtract(datetime, duration) - Subtract duration
```

## Использование

### Function Browser
```javascript
class FunctionBrowser {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.functions = null;
  }
  
  async loadFunctions() {
    const response = await fetch('/api/v1/expressions/functions', {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    const result = await response.json();
    this.functions = result.data.functions;
    
    return this.functions;
  }
  
  searchFunctions(query) {
    if (!this.functions) {
      throw new Error('Functions not loaded. Call loadFunctions() first.');
    }
    
    const searchTerm = query.toLowerCase();
    
    return this.functions.filter(func => 
      func.name.toLowerCase().includes(searchTerm) ||
      func.description.toLowerCase().includes(searchTerm) ||
      func.category.toLowerCase().includes(searchTerm)
    );
  }
  
  getFunctionsByCategory(category) {
    if (!this.functions) {
      throw new Error('Functions not loaded. Call loadFunctions() first.');
    }
    
    return this.functions.filter(func => func.category === category);
  }
  
  getFunctionDetails(functionName) {
    if (!this.functions) {
      throw new Error('Functions not loaded. Call loadFunctions() first.');
    }
    
    return this.functions.find(func => func.name === functionName);
  }
  
  validateFunctionCall(functionName, args) {
    const func = this.getFunctionDetails(functionName);
    
    if (!func) {
      return {
        valid: false,
        error: `Function '${functionName}' not found`
      };
    }
    
    const requiredParams = func.parameters.filter(p => p.required);
    
    if (args.length < requiredParams.length) {
      return {
        valid: false,
        error: `Function '${functionName}' requires ${requiredParams.length} arguments, got ${args.length}`
      };
    }
    
    if (args.length > func.parameters.length) {
      return {
        valid: false,
        error: `Function '${functionName}' accepts maximum ${func.parameters.length} arguments, got ${args.length}`
      };
    }
    
    return { valid: true };
  }
  
  generateDocumentation(category) {
    const functions = this.getFunctionsByCategory(category);
    
    let doc = `# ${category.toUpperCase()} Functions\n\n`;
    
    functions.forEach(func => {
      doc += `## ${func.name}\n\n`;
      doc += `**Description:** ${func.description}\n\n`;
      doc += `**Signature:** \`${func.signature}\`\n\n`;
      doc += `**Returns:** ${func.return_type}\n\n`;
      
      if (func.parameters.length > 0) {
        doc += `**Parameters:**\n`;
        func.parameters.forEach(param => {
          const required = param.required ? ' (required)' : ' (optional)';
          doc += `- \`${param.name}\` (${param.type})${required}: ${param.description}\n`;
        });
        doc += '\n';
      }
      
      if (func.examples && func.examples.length > 0) {
        doc += `**Examples:**\n`;
        func.examples.forEach((example, index) => {
          doc += `${index + 1}. \`${example.expression}\` → \`${JSON.stringify(example.result)}\`\n`;
          if (example.description) {
            doc += `   ${example.description}\n`;
          }
        });
        doc += '\n';
      }
      
      doc += '---\n\n';
    });
    
    return doc;
  }
}

// Использование
const browser = new FunctionBrowser('your-api-key');

// Загрузка всех функций
await browser.loadFunctions();

// Поиск функций
const mathFunctions = browser.searchFunctions('sum');
console.log('Math functions:', mathFunctions);

// Функции по категории
const stringFunctions = browser.getFunctionsByCategory('string');
console.log('String functions:', stringFunctions);

// Детали функции
const sumFunction = browser.getFunctionDetails('sum');
console.log('Sum function details:', sumFunction);

// Валидация вызова функции
const validation = browser.validateFunctionCall('sum', [['1', '2', '3']]);
console.log('Validation result:', validation);
```

### Expression Builder with Function Help
```javascript
class ExpressionBuilderWithHelp {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.functions = null;
  }
  
  async initialize() {
    await this.loadFunctions();
    this.setupAutocompletion();
  }
  
  async loadFunctions() {
    const response = await fetch('/api/v1/expressions/functions', {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    const result = await response.json();
    this.functions = result.data.functions;
  }
  
  setupAutocompletion() {
    // Setup autocompletion for expression editor
    const functionNames = this.functions.map(f => f.name);
    
    // This would integrate with your editor's autocompletion
    this.autocompletionItems = functionNames.map(name => ({
      label: name,
      kind: 'function',
      detail: this.getFunctionSignature(name),
      documentation: this.getFunctionDocumentation(name)
    }));
  }
  
  getFunctionSignature(functionName) {
    const func = this.functions.find(f => f.name === functionName);
    return func ? func.signature : '';
  }
  
  getFunctionDocumentation(functionName) {
    const func = this.functions.find(f => f.name === functionName);
    if (!func) return '';
    
    let doc = func.description + '\n\n';
    
    if (func.parameters.length > 0) {
      doc += 'Parameters:\n';
      func.parameters.forEach(param => {
        doc += `- ${param.name} (${param.type}): ${param.description}\n`;
      });
    }
    
    if (func.examples && func.examples.length > 0) {
      doc += '\nExamples:\n';
      func.examples.forEach(example => {
        doc += `${example.expression} → ${JSON.stringify(example.result)}\n`;
      });
    }
    
    return doc;
  }
  
  getContextualHelp(cursorPosition, expression) {
    // Analyze expression at cursor position and provide contextual help
    const beforeCursor = expression.substring(0, cursorPosition);
    const currentWord = this.getCurrentWord(beforeCursor);
    
    if (this.isTypingFunction(beforeCursor)) {
      return this.getFunctionSuggestions(currentWord);
    }
    
    return this.getGeneralSuggestions();
  }
  
  getFunctionSuggestions(partial) {
    return this.functions
      .filter(func => func.name.startsWith(partial.toLowerCase()))
      .map(func => ({
        name: func.name,
        signature: func.signature,
        description: func.description,
        category: func.category
      }));
  }
  
  getCurrentWord(text) {
    const match = text.match(/\w+$/);
    return match ? match[0] : '';
  }
  
  isTypingFunction(text) {
    // Simple heuristic: check if we're after a letter and before a parenthesis
    return /\w+$/.test(text);
  }
}
```

### Function Usage Analytics
```javascript
class FunctionUsageAnalytics {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.usageStats = new Map();
  }
  
  async loadFunctions() {
    const response = await fetch('/api/v1/expressions/functions', {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    return await response.json();
  }
  
  recordFunctionUsage(functionName, context = {}) {
    const current = this.usageStats.get(functionName) || {
      count: 0,
      contexts: [],
      firstUsed: Date.now(),
      lastUsed: Date.now()
    };
    
    current.count++;
    current.lastUsed = Date.now();
    current.contexts.push({
      timestamp: Date.now(),
      context: context
    });
    
    // Keep only last 100 contexts to avoid memory bloat
    if (current.contexts.length > 100) {
      current.contexts.shift();
    }
    
    this.usageStats.set(functionName, current);
  }
  
  getUsageReport() {
    const functions = Array.from(this.usageStats.entries())
      .map(([name, stats]) => ({ name, ...stats }))
      .sort((a, b) => b.count - a.count);
    
    return {
      totalFunctions: this.usageStats.size,
      mostUsed: functions.slice(0, 10),
      leastUsed: functions.slice(-10),
      usageByCategory: this.groupByCategory(functions),
      totalUsages: functions.reduce((sum, f) => sum + f.count, 0)
    };
  }
  
  groupByCategory(functions) {
    // This would require mapping function names to categories from the API
    return {};
  }
  
  recommendFunctions(currentExpression) {
    // Analyze current expression and recommend useful functions
    const recommendations = [];
    
    if (currentExpression.includes('list') || currentExpression.includes('[')) {
      recommendations.push({
        function: 'count',
        reason: 'For counting list elements',
        example: 'count(items)'
      });
    }
    
    if (currentExpression.includes('string') || currentExpression.includes('"')) {
      recommendations.push({
        function: 'contains',
        reason: 'For string search operations',
        example: 'contains(text, "keyword")'
      });
    }
    
    return recommendations;
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Использование функций в выражениях
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация вызовов функций
- [`POST /api/v1/expressions/test`](./test-expression.md) - Тестирование функций
