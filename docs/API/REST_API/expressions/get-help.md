# GET /api/v1/expressions/help

## Описание
Получение справочной информации по FEEL expressions. Включает документацию по синтаксису, примеры и руководство по использованию.

## URL
```
GET /api/v1/expressions/help
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression`

## Параметры запроса (Query Parameters)

### Фильтрация контента
- `topic` (string): Конкретная тема справки (`syntax`, `functions`, `examples`, `operators`, `types`)
- `format` (string): Формат ответа (`json`, `markdown`, `html`) (по умолчанию: `json`)
- `language` (string): Язык документации (`en`, `ru`) (по умолчанию: `en`)

## Примеры запросов

### Полная справка
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/help" \
  -H "X-API-Key: your-api-key-here"
```

### Справка по синтаксису
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/help?topic=syntax" \
  -H "X-API-Key: your-api-key-here"
```

### Справка в формате Markdown
```bash
curl -X GET "http://localhost:27555/api/v1/expressions/help?format=markdown" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/expressions/help?topic=functions&language=ru', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const help = await response.json();
console.log('Functions help:', help.data);
```

## Ответы

### 200 OK - Полная справка
```json
{
  "success": true,
  "data": {
    "title": "FEEL Expressions Help",
    "version": "1.1",
    "last_updated": "2025-01-11T10:30:00Z",
    "sections": {
      "overview": {
        "title": "Overview",
        "content": "FEEL (Friendly Enough Expression Language) is a simple expression language for decision models. It provides a way to define and evaluate expressions in business contexts.",
        "subsections": [
          {
            "title": "Key Features",
            "items": [
              "Simple and readable syntax",
              "Type-safe evaluation",
              "Rich set of built-in functions",
              "Support for complex data structures",
              "Date and time handling"
            ]
          }
        ]
      },
      "syntax": {
        "title": "Syntax Reference",
        "content": "FEEL syntax is designed to be intuitive and business-friendly.",
        "subsections": [
          {
            "title": "Basic Expressions",
            "examples": [
              {
                "expression": "x + y",
                "description": "Arithmetic addition",
                "result": "Sum of x and y"
              },
              {
                "expression": "name = \"John\"",
                "description": "String comparison",
                "result": "Boolean result"
              },
              {
                "expression": "age > 18",
                "description": "Numeric comparison",
                "result": "Boolean result"
              }
            ]
          },
          {
            "title": "Conditional Expressions",
            "examples": [
              {
                "expression": "if condition then value1 else value2",
                "description": "If-then-else expression",
                "result": "value1 if condition is true, value2 otherwise"
              }
            ]
          },
          {
            "title": "List Operations",
            "examples": [
              {
                "expression": "[1, 2, 3, 4]",
                "description": "List literal",
                "result": "List of numbers"
              },
              {
                "expression": "list[1]",
                "description": "List indexing (1-based)",
                "result": "First element of list"
              },
              {
                "expression": "list[condition]",
                "description": "List filtering",
                "result": "Filtered list elements"
              }
            ]
          }
        ]
      },
      "operators": {
        "title": "Operators",
        "content": "FEEL supports various types of operators for different operations.",
        "subsections": [
          {
            "title": "Arithmetic Operators",
            "operators": [
              {
                "symbol": "+",
                "name": "Addition",
                "description": "Adds two numbers",
                "example": "5 + 3 = 8"
              },
              {
                "symbol": "-",
                "name": "Subtraction", 
                "description": "Subtracts second number from first",
                "example": "5 - 3 = 2"
              },
              {
                "symbol": "*",
                "name": "Multiplication",
                "description": "Multiplies two numbers",
                "example": "5 * 3 = 15"
              },
              {
                "symbol": "/",
                "name": "Division",
                "description": "Divides first number by second",
                "example": "6 / 3 = 2"
              },
              {
                "symbol": "**",
                "name": "Exponentiation",
                "description": "Raises first number to power of second",
                "example": "2 ** 3 = 8"
              }
            ]
          },
          {
            "title": "Comparison Operators",
            "operators": [
              {
                "symbol": "=",
                "name": "Equal",
                "description": "Tests for equality",
                "example": "5 = 5 → true"
              },
              {
                "symbol": "!=",
                "name": "Not Equal",
                "description": "Tests for inequality",
                "example": "5 != 3 → true"
              },
              {
                "symbol": "<",
                "name": "Less Than",
                "description": "Tests if first is less than second",
                "example": "3 < 5 → true"
              },
              {
                "symbol": "<=",
                "name": "Less Than or Equal",
                "description": "Tests if first is less than or equal to second",
                "example": "5 <= 5 → true"
              },
              {
                "symbol": ">",
                "name": "Greater Than",
                "description": "Tests if first is greater than second",
                "example": "5 > 3 → true"
              },
              {
                "symbol": ">=",
                "name": "Greater Than or Equal",
                "description": "Tests if first is greater than or equal to second",
                "example": "5 >= 5 → true"
              }
            ]
          },
          {
            "title": "Logical Operators",
            "operators": [
              {
                "symbol": "and",
                "name": "Logical AND",
                "description": "Returns true if both operands are true",
                "example": "true and false → false"
              },
              {
                "symbol": "or",
                "name": "Logical OR",
                "description": "Returns true if at least one operand is true",
                "example": "true or false → true"
              },
              {
                "symbol": "not",
                "name": "Logical NOT",
                "description": "Returns opposite boolean value",
                "example": "not true → false"
              }
            ]
          }
        ]
      },
      "functions": {
        "title": "Built-in Functions",
        "content": "FEEL provides a comprehensive set of built-in functions for various operations.",
        "categories": [
          {
            "name": "Math Functions",
            "functions": [
              {
                "name": "abs",
                "signature": "abs(number)",
                "description": "Returns absolute value",
                "example": "abs(-5) → 5"
              },
              {
                "name": "sum",
                "signature": "sum(list)",
                "description": "Sums all numbers in list",
                "example": "sum([1, 2, 3]) → 6"
              },
              {
                "name": "mean",
                "signature": "mean(list)",
                "description": "Calculates average of numbers",
                "example": "mean([1, 2, 3]) → 2"
              }
            ]
          },
          {
            "name": "String Functions",
            "functions": [
              {
                "name": "upper",
                "signature": "upper(string)",
                "description": "Converts to uppercase",
                "example": "upper(\"hello\") → \"HELLO\""
              },
              {
                "name": "substring",
                "signature": "substring(string, start, length?)",
                "description": "Extracts substring",
                "example": "substring(\"hello\", 2, 3) → \"ell\""
              },
              {
                "name": "contains",
                "signature": "contains(string, substring)",
                "description": "Checks if string contains substring",
                "example": "contains(\"hello\", \"ell\") → true"
              }
            ]
          }
        ]
      },
      "types": {
        "title": "Data Types",
        "content": "FEEL supports various data types for different kinds of values.",
        "types": [
          {
            "name": "number",
            "description": "Numeric values including integers and decimals",
            "examples": ["42", "3.14", "-17.5"],
            "operations": ["Arithmetic operations", "Comparisons", "Math functions"]
          },
          {
            "name": "string",
            "description": "Text values enclosed in double quotes",
            "examples": ["\"hello\"", "\"world\"", "\"Hello, World!\""],
            "operations": ["String functions", "Comparisons", "Concatenation"]
          },
          {
            "name": "boolean",
            "description": "Logical values representing true or false",
            "examples": ["true", "false"],
            "operations": ["Logical operations", "Conditional expressions"]
          },
          {
            "name": "list",
            "description": "Ordered collection of values",
            "examples": ["[1, 2, 3]", "[\"a\", \"b\", \"c\"]", "[true, false]"],
            "operations": ["Indexing", "Filtering", "List functions"]
          },
          {
            "name": "object",
            "description": "Key-value pairs representing structured data",
            "examples": ["{name: \"John\", age: 30}", "{x: 1, y: 2}"],
            "operations": ["Property access", "Object construction"]
          },
          {
            "name": "date",
            "description": "Date values",
            "examples": ["date(\"2025-01-11\")", "today()"],
            "operations": ["Date functions", "Date arithmetic", "Formatting"]
          },
          {
            "name": "time",
            "description": "Time values",
            "examples": ["time(\"10:30:00\")", "time(10, 30, 0)"],
            "operations": ["Time functions", "Time arithmetic"]
          },
          {
            "name": "duration",
            "description": "Time duration values",
            "examples": ["duration(\"PT5M\")", "duration(\"P1D\")"],
            "operations": ["Duration arithmetic", "Date/time calculations"]
          }
        ]
      },
      "examples": {
        "title": "Common Examples",
        "content": "Practical examples of FEEL expressions in real-world scenarios.",
        "categories": [
          {
            "name": "Business Rules",
            "examples": [
              {
                "title": "Customer Discount",
                "expression": "if customer.type = \"VIP\" and order.amount > 100 then order.amount * 0.1 else 0",
                "description": "Calculate 10% discount for VIP customers with orders over $100",
                "context": {
                  "customer": {"type": "VIP"},
                  "order": {"amount": 150}
                },
                "result": 15
              },
              {
                "title": "Age Verification",
                "expression": "age >= 18 and age <= 65",
                "description": "Check if person is of working age",
                "context": {"age": 25},
                "result": true
              }
            ]
          },
          {
            "name": "Data Processing",
            "examples": [
              {
                "title": "Order Total",
                "expression": "sum(items[].price * items[].quantity)",
                "description": "Calculate total price for all items",
                "context": {
                  "items": [
                    {"price": 10, "quantity": 2},
                    {"price": 5, "quantity": 3}
                  ]
                },
                "result": 35
              },
              {
                "title": "Text Processing",
                "expression": "upper(substring(name, 1, 1)) + lower(substring(name, 2))",
                "description": "Capitalize first letter of name",
                "context": {"name": "john"},
                "result": "John"
              }
            ]
          }
        ]
      }
    },
    "quick_reference": {
      "common_patterns": [
        {
          "pattern": "if-then-else",
          "syntax": "if condition then value1 else value2",
          "use_case": "Conditional logic"
        },
        {
          "pattern": "list filtering",
          "syntax": "list[condition]",
          "use_case": "Filter list elements"
        },
        {
          "pattern": "object property access",
          "syntax": "object.property",
          "use_case": "Access object properties"
        },
        {
          "pattern": "function call",
          "syntax": "function_name(arg1, arg2)",
          "use_case": "Call built-in functions"
        }
      ],
      "best_practices": [
        "Use clear variable names",
        "Break complex expressions into simpler parts",
        "Use parentheses for clarity",
        "Test expressions with various inputs",
        "Consider null values in conditions"
      ]
    }
  },
  "request_id": "req_1641998404100"
}
```

### 200 OK - Справка по синтаксису
```json
{
  "success": true,
  "data": {
    "topic": "syntax",
    "title": "FEEL Syntax Reference",
    "sections": [
      {
        "title": "Basic Syntax Rules",
        "rules": [
          {
            "rule": "Expressions are case-sensitive",
            "example": "name != Name"
          },
          {
            "rule": "Strings must be enclosed in double quotes",
            "example": "\"Hello World\""
          },
          {
            "rule": "List indexing is 1-based",
            "example": "list[1] // first element"
          },
          {
            "rule": "Property access uses dot notation",
            "example": "customer.name"
          }
        ]
      },
      {
        "title": "Expression Types",
        "types": [
          {
            "type": "Literal Expressions",
            "description": "Direct values",
            "examples": ["42", "\"hello\"", "true", "[1, 2, 3]"]
          },
          {
            "type": "Variable Expressions",
            "description": "Reference to variables",
            "examples": ["x", "customer", "order.amount"]
          },
          {
            "type": "Binary Expressions",
            "description": "Operations with two operands",
            "examples": ["x + y", "age > 18", "name = \"John\""]
          },
          {
            "type": "Function Expressions", 
            "description": "Function calls",
            "examples": ["sum([1, 2, 3])", "upper(name)", "now()"]
          }
        ]
      }
    ]
  }
}
```

## Использование

### Interactive Help System
```javascript
class FEELHelpSystem {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.helpCache = new Map();
  }
  
  async getHelp(topic = null, format = 'json') {
    const cacheKey = `${topic || 'all'}-${format}`;
    
    if (this.helpCache.has(cacheKey)) {
      return this.helpCache.get(cacheKey);
    }
    
    const params = new URLSearchParams();
    if (topic) params.append('topic', topic);
    if (format !== 'json') params.append('format', format);
    
    const response = await fetch(`/api/v1/expressions/help?${params}`, {
      headers: { 'X-API-Key': this.apiKey }
    });
    
    const result = await response.json();
    this.helpCache.set(cacheKey, result.data);
    
    return result.data;
  }
  
  async searchHelp(query) {
    const help = await this.getHelp();
    const results = [];
    
    // Search in all sections
    this.searchInObject(help, query.toLowerCase(), '', results);
    
    return results.sort((a, b) => b.relevance - a.relevance);
  }
  
  searchInObject(obj, query, path, results, depth = 0) {
    if (depth > 10) return; // Prevent infinite recursion
    
    for (const [key, value] of Object.entries(obj)) {
      const currentPath = path ? `${path}.${key}` : key;
      
      if (typeof value === 'string') {
        const relevance = this.calculateRelevance(value, query);
        if (relevance > 0) {
          results.push({
            path: currentPath,
            content: value,
            relevance,
            type: this.determineContentType(currentPath)
          });
        }
      } else if (Array.isArray(value)) {
        value.forEach((item, index) => {
          if (typeof item === 'object') {
            this.searchInObject(item, query, `${currentPath}[${index}]`, results, depth + 1);
          } else if (typeof item === 'string') {
            const relevance = this.calculateRelevance(item, query);
            if (relevance > 0) {
              results.push({
                path: `${currentPath}[${index}]`,
                content: item,
                relevance,
                type: 'array_item'
              });
            }
          }
        });
      } else if (typeof value === 'object' && value !== null) {
        this.searchInObject(value, query, currentPath, results, depth + 1);
      }
    }
  }
  
  calculateRelevance(text, query) {
    const lowerText = text.toLowerCase();
    
    if (lowerText.includes(query)) {
      // Exact match gets highest score
      if (lowerText === query) return 100;
      
      // Word boundary match gets high score
      const wordRegex = new RegExp(`\\b${query}\\b`);
      if (wordRegex.test(lowerText)) return 80;
      
      // Partial match gets medium score
      return 40;
    }
    
    // Check for partial word matches
    const words = query.split(' ');
    let matchCount = 0;
    
    words.forEach(word => {
      if (lowerText.includes(word)) {
        matchCount++;
      }
    });
    
    return matchCount > 0 ? (matchCount / words.length) * 30 : 0;
  }
  
  determineContentType(path) {
    if (path.includes('example')) return 'example';
    if (path.includes('function')) return 'function';
    if (path.includes('operator')) return 'operator';
    if (path.includes('syntax')) return 'syntax';
    return 'general';
  }
  
  async getContextualHelp(expression, cursorPosition) {
    // Analyze expression context and provide relevant help
    const context = this.analyzeExpressionContext(expression, cursorPosition);
    const help = await this.getHelp();
    
    const suggestions = [];
    
    switch (context.type) {
      case 'function_call':
        suggestions.push(...this.getFunctionHelp(context.functionName, help));
        break;
      case 'operator':
        suggestions.push(...this.getOperatorHelp(context.operator, help));
        break;
      case 'property_access':
        suggestions.push(...this.getPropertyHelp(help));
        break;
      default:
        suggestions.push(...this.getGeneralHelp(help));
    }
    
    return suggestions;
  }
  
  analyzeExpressionContext(expression, cursorPosition) {
    const beforeCursor = expression.substring(0, cursorPosition);
    const afterCursor = expression.substring(cursorPosition);
    
    // Check if we're in a function call
    const functionMatch = beforeCursor.match(/(\w+)\s*\(\s*$/);
    if (functionMatch) {
      return {
        type: 'function_call',
        functionName: functionMatch[1]
      };
    }
    
    // Check if we're after an operator
    const operatorMatch = beforeCursor.match(/\s*([\+\-\*\/\>\<\=]+)\s*$/);
    if (operatorMatch) {
      return {
        type: 'operator',
        operator: operatorMatch[1]
      };
    }
    
    // Check if we're accessing a property
    if (beforeCursor.endsWith('.')) {
      return {
        type: 'property_access'
      };
    }
    
    return { type: 'general' };
  }
  
  getFunctionHelp(functionName, help) {
    const suggestions = [];
    
    // Find function in help data
    if (help.sections && help.sections.functions) {
      help.sections.functions.categories?.forEach(category => {
        const func = category.functions?.find(f => f.name === functionName);
        if (func) {
          suggestions.push({
            type: 'function',
            title: `${func.name}()`,
            description: func.description,
            signature: func.signature,
            example: func.example
          });
        }
      });
    }
    
    return suggestions;
  }
  
  getOperatorHelp(operator, help) {
    const suggestions = [];
    
    if (help.sections && help.sections.operators) {
      help.sections.operators.subsections?.forEach(subsection => {
        const op = subsection.operators?.find(o => o.symbol === operator);
        if (op) {
          suggestions.push({
            type: 'operator',
            title: `${op.symbol} (${op.name})`,
            description: op.description,
            example: op.example
          });
        }
      });
    }
    
    return suggestions;
  }
  
  generateHelpMarkdown(helpData) {
    let markdown = `# ${helpData.title}\n\n`;
    
    if (helpData.sections) {
      Object.entries(helpData.sections).forEach(([sectionKey, section]) => {
        markdown += `## ${section.title}\n\n`;
        
        if (section.content) {
          markdown += `${section.content}\n\n`;
        }
        
        if (section.subsections) {
          section.subsections.forEach(subsection => {
            markdown += `### ${subsection.title}\n\n`;
            
            if (subsection.examples) {
              subsection.examples.forEach(example => {
                markdown += `**${example.expression}**\n`;
                markdown += `${example.description}\n\n`;
              });
            }
            
            if (subsection.operators) {
              subsection.operators.forEach(op => {
                markdown += `- \`${op.symbol}\` - ${op.description}\n`;
                markdown += `  Example: \`${op.example}\`\n\n`;
              });
            }
          });
        }
      });
    }
    
    return markdown;
  }
}

// Использование
const helpSystem = new FEELHelpSystem('your-api-key');

// Получение полной справки
const fullHelp = await helpSystem.getHelp();
console.log('Full help:', fullHelp);

// Поиск в справке
const searchResults = await helpSystem.searchHelp('sum function');
console.log('Search results:', searchResults);

// Контекстная справка
const contextHelp = await helpSystem.getContextualHelp('sum(', 4);
console.log('Context help:', contextHelp);

// Генерация Markdown документации
const markdown = helpSystem.generateHelpMarkdown(fullHelp);
console.log('Markdown help:', markdown);
```

### Help Widget Integration
```javascript
class HelpWidget {
  constructor(apiKey, container) {
    this.apiKey = apiKey;
    this.container = container;
    this.helpSystem = new FEELHelpSystem(apiKey);
    this.isVisible = false;
  }
  
  async initialize() {
    await this.helpSystem.getHelp(); // Preload help data
    this.createWidget();
    this.attachEventListeners();
  }
  
  createWidget() {
    this.widget = document.createElement('div');
    this.widget.className = 'feel-help-widget';
    this.widget.innerHTML = `
      <div class="help-header">
        <h3>FEEL Help</h3>
        <button class="close-btn">&times;</button>
      </div>
      <div class="help-search">
        <input type="text" placeholder="Search help..." class="search-input">
      </div>
      <div class="help-content">
        <div class="help-tabs">
          <button class="tab-btn active" data-tab="syntax">Syntax</button>
          <button class="tab-btn" data-tab="functions">Functions</button>
          <button class="tab-btn" data-tab="examples">Examples</button>
        </div>
        <div class="tab-content" id="syntax-tab">
          <!-- Syntax help content -->
        </div>
        <div class="tab-content" id="functions-tab" style="display: none;">
          <!-- Functions help content -->
        </div>
        <div class="tab-content" id="examples-tab" style="display: none;">
          <!-- Examples help content -->
        </div>
      </div>
    `;
    
    this.container.appendChild(this.widget);
  }
  
  attachEventListeners() {
    // Close button
    this.widget.querySelector('.close-btn').addEventListener('click', () => {
      this.hide();
    });
    
    // Search input
    const searchInput = this.widget.querySelector('.search-input');
    searchInput.addEventListener('input', async (e) => {
      if (e.target.value.length > 2) {
        const results = await this.helpSystem.searchHelp(e.target.value);
        this.displaySearchResults(results);
      }
    });
    
    // Tab buttons
    this.widget.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        this.switchTab(e.target.dataset.tab);
      });
    });
  }
  
  async show(topic = null) {
    this.isVisible = true;
    this.widget.style.display = 'block';
    
    if (topic) {
      const helpData = await this.helpSystem.getHelp(topic);
      this.displayTopicHelp(topic, helpData);
    }
  }
  
  hide() {
    this.isVisible = false;
    this.widget.style.display = 'none';
  }
  
  toggle() {
    if (this.isVisible) {
      this.hide();
    } else {
      this.show();
    }
  }
  
  switchTab(tabName) {
    // Update tab buttons
    this.widget.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.remove('active');
    });
    this.widget.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    
    // Show tab content
    this.widget.querySelectorAll('.tab-content').forEach(content => {
      content.style.display = 'none';
    });
    this.widget.querySelector(`#${tabName}-tab`).style.display = 'block';
  }
  
  displaySearchResults(results) {
    const contentArea = this.widget.querySelector('.help-content');
    contentArea.innerHTML = '<h4>Search Results</h4>';
    
    if (results.length === 0) {
      contentArea.innerHTML += '<p>No results found.</p>';
      return;
    }
    
    results.forEach(result => {
      const resultElement = document.createElement('div');
      resultElement.className = 'search-result';
      resultElement.innerHTML = `
        <div class="result-type">${result.type}</div>
        <div class="result-content">${result.content}</div>
        <div class="result-path">${result.path}</div>
      `;
      contentArea.appendChild(resultElement);
    });
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Применение полученных знаний
- [`GET /api/v1/expressions/functions`](./list-functions.md) - Подробная информация о функциях
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация синтаксиса
