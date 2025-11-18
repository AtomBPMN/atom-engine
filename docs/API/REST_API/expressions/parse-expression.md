# POST /api/v1/expressions/parse

## Описание
Парсинг FEEL выражения в Abstract Syntax Tree (AST). Возвращает структурированное представление выражения.

## URL
```
POST /api/v1/expressions/parse
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
- `expression` (string): FEEL выражение для парсинга

### Опциональные поля
- `include_positions` (boolean): Включить позиции в AST (по умолчанию: false)
- `simplify` (boolean): Упростить AST (по умолчанию: true)

## Примеры запросов

### Простое выражение
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/parse" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "x + y * 2"
  }'
```

### Сложное выражение с позициями
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/parse" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "if order.amount > 100 then \"discount\" else \"normal\"",
    "include_positions": true,
    "simplify": false
  }'
```

### JavaScript
```javascript
const parseRequest = {
  expression: 'sum(items[price > 50].price)',
  include_positions: true,
  simplify: true
};

const response = await fetch('/api/v1/expressions/parse', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(parseRequest)
});

const result = await response.json();
console.log('AST:', result.data.ast);
```

## Ответы

### 200 OK - Выражение распарсено
```json
{
  "success": true,
  "data": {
    "expression": "x + y * 2",
    "ast": {
      "type": "BinaryExpression",
      "operator": "+",
      "left": {
        "type": "Identifier",
        "name": "x"
      },
      "right": {
        "type": "BinaryExpression",
        "operator": "*",
        "left": {
          "type": "Identifier",
          "name": "y"
        },
        "right": {
          "type": "Literal",
          "value": 2,
          "raw": "2"
        }
      }
    },
    "tokens": [
      {"type": "IDENTIFIER", "value": "x", "position": 0},
      {"type": "OPERATOR", "value": "+", "position": 2},
      {"type": "IDENTIFIER", "value": "y", "position": 4},
      {"type": "OPERATOR", "value": "*", "position": 6},
      {"type": "NUMBER", "value": "2", "position": 8}
    ],
    "metadata": {
      "variables": ["x", "y"],
      "functions": [],
      "literals": [2],
      "operators": ["+", "*"],
      "complexity": 3,
      "depth": 2
    },
    "parse_time_ms": 5
  },
  "request_id": "req_1641998403800"
}
```

### 200 OK - Условное выражение с позициями
```json
{
  "success": true,
  "data": {
    "expression": "if order.amount > 100 then \"discount\" else \"normal\"",
    "ast": {
      "type": "ConditionalExpression",
      "test": {
        "type": "BinaryExpression",
        "operator": ">",
        "left": {
          "type": "MemberExpression",
          "object": {
            "type": "Identifier",
            "name": "order",
            "position": {"start": 3, "end": 8}
          },
          "property": {
            "type": "Identifier", 
            "name": "amount",
            "position": {"start": 9, "end": 15}
          },
          "computed": false,
          "position": {"start": 3, "end": 15}
        },
        "right": {
          "type": "Literal",
          "value": 100,
          "raw": "100",
          "position": {"start": 18, "end": 21}
        },
        "position": {"start": 3, "end": 21}
      },
      "consequent": {
        "type": "Literal",
        "value": "discount",
        "raw": "\"discount\"",
        "position": {"start": 27, "end": 37}
      },
      "alternate": {
        "type": "Literal",
        "value": "normal",
        "raw": "\"normal\"",
        "position": {"start": 43, "end": 51}
      },
      "position": {"start": 0, "end": 51}
    },
    "metadata": {
      "variables": ["order.amount"],
      "functions": [],
      "literals": [100, "discount", "normal"],
      "operators": [">"],
      "complexity": 4,
      "depth": 3
    },
    "parse_time_ms": 12
  },
  "request_id": "req_1641998403801"
}
```

### 400 Bad Request - Синтаксическая ошибка
```json
{
  "success": false,
  "error": {
    "code": "PARSE_ERROR",
    "message": "Unable to parse expression",
    "details": {
      "expression": "x + y *",
      "error_position": 7,
      "error_description": "Unexpected end of input",
      "expected_tokens": ["IDENTIFIER", "NUMBER", "LPAREN"],
      "partial_ast": {
        "type": "BinaryExpression",
        "operator": "+",
        "left": {
          "type": "Identifier",
          "name": "x"
        },
        "right": {
          "type": "IncompleteExpression",
          "operator": "*",
          "left": {
            "type": "Identifier",
            "name": "y"
          }
        }
      }
    }
  },
  "request_id": "req_1641998403802"
}
```

## AST Node Types

### Expression Types
```yaml
BinaryExpression:
  - operator: "+", "-", "*", "/", ">", "<", ">=", "<=", "=", "!="
  - left: Expression
  - right: Expression

UnaryExpression:
  - operator: "-", "not"
  - argument: Expression

ConditionalExpression:
  - test: Expression
  - consequent: Expression
  - alternate: Expression

CallExpression:
  - callee: Identifier
  - arguments: [Expression]

MemberExpression:
  - object: Expression
  - property: Identifier
  - computed: boolean
```

### Literal Types
```yaml
Literal:
  - value: any
  - raw: string

Identifier:
  - name: string

ArrayExpression:
  - elements: [Expression]

ObjectExpression:
  - properties: [Property]
```

### Advanced Types
```yaml
FunctionExpression:
  - params: [Identifier]
  - body: Expression

QuantifiedExpression:
  - quantifier: "some", "every"
  - variables: [Identifier]
  - condition: Expression

FilterExpression:
  - expression: Expression
  - filter: Expression
```

## Использование

### AST Analysis
```javascript
class ASTAnalyzer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async analyzeExpression(expression) {
    const response = await fetch('/api/v1/expressions/parse', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        expression,
        include_positions: true,
        simplify: false
      })
    });
    
    const result = await response.json();
    
    if (!result.success) {
      throw new Error(result.error.message);
    }
    
    return {
      ast: result.data.ast,
      variables: this.extractVariables(result.data.ast),
      functions: this.extractFunctions(result.data.ast),
      complexity: this.calculateComplexity(result.data.ast),
      dependencies: this.findDependencies(result.data.ast)
    };
  }
  
  extractVariables(ast) {
    const variables = new Set();
    
    this.traverseAST(ast, (node) => {
      if (node.type === 'Identifier' && this.isVariable(node)) {
        variables.add(node.name);
      } else if (node.type === 'MemberExpression') {
        variables.add(this.getMemberPath(node));
      }
    });
    
    return Array.from(variables);
  }
  
  extractFunctions(ast) {
    const functions = new Set();
    
    this.traverseAST(ast, (node) => {
      if (node.type === 'CallExpression') {
        functions.add(node.callee.name);
      }
    });
    
    return Array.from(functions);
  }
  
  calculateComplexity(ast) {
    let complexity = 0;
    
    this.traverseAST(ast, (node) => {
      switch (node.type) {
        case 'BinaryExpression':
        case 'UnaryExpression':
          complexity += 1;
          break;
        case 'ConditionalExpression':
          complexity += 2;
          break;
        case 'CallExpression':
          complexity += 1 + (node.arguments?.length || 0);
          break;
        case 'QuantifiedExpression':
          complexity += 3;
          break;
      }
    });
    
    return complexity;
  }
  
  findDependencies(ast) {
    const dependencies = {
      variables: [],
      functions: [],
      objects: []
    };
    
    this.traverseAST(ast, (node) => {
      if (node.type === 'Identifier') {
        dependencies.variables.push(node.name);
      } else if (node.type === 'CallExpression') {
        dependencies.functions.push(node.callee.name);
      } else if (node.type === 'MemberExpression') {
        const rootObject = this.getRootObject(node);
        dependencies.objects.push(rootObject);
      }
    });
    
    // Remove duplicates
    Object.keys(dependencies).forEach(key => {
      dependencies[key] = [...new Set(dependencies[key])];
    });
    
    return dependencies;
  }
  
  traverseAST(node, callback) {
    if (!node || typeof node !== 'object') return;
    
    callback(node);
    
    Object.values(node).forEach(value => {
      if (Array.isArray(value)) {
        value.forEach(item => this.traverseAST(item, callback));
      } else if (typeof value === 'object') {
        this.traverseAST(value, callback);
      }
    });
  }
  
  getMemberPath(node) {
    if (node.type === 'MemberExpression') {
      const object = this.getMemberPath(node.object);
      return `${object}.${node.property.name}`;
    } else if (node.type === 'Identifier') {
      return node.name;
    }
    return '';
  }
  
  getRootObject(node) {
    if (node.type === 'MemberExpression') {
      return this.getRootObject(node.object);
    } else if (node.type === 'Identifier') {
      return node.name;
    }
    return '';
  }
  
  isVariable(node) {
    // Determine if identifier is a variable vs function name
    return true; // Simplified for example
  }
}

// Использование
const analyzer = new ASTAnalyzer('your-api-key');

const analysis = await analyzer.analyzeExpression(
  'if order.amount > threshold and customer.type = "VIP" then calculateDiscount(order.amount) else 0'
);

console.log('Variables:', analysis.variables);
console.log('Functions:', analysis.functions);
console.log('Complexity:', analysis.complexity);
console.log('Dependencies:', analysis.dependencies);
```

### Expression Transformer
```javascript
class ExpressionTransformer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async transformExpression(expression, transformations) {
    const parseResult = await this.parseExpression(expression);
    let transformedAST = parseResult.ast;
    
    for (const transformation of transformations) {
      transformedAST = this.applyTransformation(transformedAST, transformation);
    }
    
    return this.astToExpression(transformedAST);
  }
  
  async parseExpression(expression) {
    const response = await fetch('/api/v1/expressions/parse', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({ expression })
    });
    
    return await response.json();
  }
  
  applyTransformation(ast, transformation) {
    switch (transformation.type) {
      case 'RENAME_VARIABLE':
        return this.renameVariable(ast, transformation.from, transformation.to);
      case 'REPLACE_FUNCTION':
        return this.replaceFunction(ast, transformation.from, transformation.to);
      case 'SIMPLIFY_CONSTANTS':
        return this.simplifyConstants(ast);
      default:
        return ast;
    }
  }
  
  renameVariable(ast, oldName, newName) {
    if (ast.type === 'Identifier' && ast.name === oldName) {
      return { ...ast, name: newName };
    }
    
    const transformed = { ...ast };
    Object.keys(transformed).forEach(key => {
      if (Array.isArray(transformed[key])) {
        transformed[key] = transformed[key].map(item => 
          this.renameVariable(item, oldName, newName)
        );
      } else if (typeof transformed[key] === 'object' && transformed[key] !== null) {
        transformed[key] = this.renameVariable(transformed[key], oldName, newName);
      }
    });
    
    return transformed;
  }
  
  replaceFunction(ast, oldFunction, newFunction) {
    if (ast.type === 'CallExpression' && ast.callee.name === oldFunction) {
      return {
        ...ast,
        callee: { ...ast.callee, name: newFunction }
      };
    }
    
    const transformed = { ...ast };
    Object.keys(transformed).forEach(key => {
      if (Array.isArray(transformed[key])) {
        transformed[key] = transformed[key].map(item => 
          this.replaceFunction(item, oldFunction, newFunction)
        );
      } else if (typeof transformed[key] === 'object' && transformed[key] !== null) {
        transformed[key] = this.replaceFunction(transformed[key], oldFunction, newFunction);
      }
    });
    
    return transformed;
  }
  
  astToExpression(ast) {
    // Convert AST back to expression string
    // This is a simplified implementation
    switch (ast.type) {
      case 'BinaryExpression':
        return `${this.astToExpression(ast.left)} ${ast.operator} ${this.astToExpression(ast.right)}`;
      case 'Identifier':
        return ast.name;
      case 'Literal':
        return typeof ast.value === 'string' ? `"${ast.value}"` : String(ast.value);
      default:
        return JSON.stringify(ast);
    }
  }
}
```

### Expression Optimizer
```javascript
class ExpressionOptimizer {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async optimizeExpression(expression) {
    const parseResult = await this.parseExpression(expression);
    const optimizedAST = this.optimizeAST(parseResult.data.ast);
    
    return {
      original: expression,
      optimized: this.astToExpression(optimizedAST),
      improvements: this.getOptimizations(parseResult.data.ast, optimizedAST)
    };
  }
  
  optimizeAST(ast) {
    // Apply various optimization techniques
    let optimized = ast;
    
    optimized = this.constantFolding(optimized);
    optimized = this.deadCodeElimination(optimized);
    optimized = this.algebraicSimplification(optimized);
    
    return optimized;
  }
  
  constantFolding(ast) {
    // Evaluate constant expressions at compile time
    if (ast.type === 'BinaryExpression') {
      const left = this.constantFolding(ast.left);
      const right = this.constantFolding(ast.right);
      
      if (left.type === 'Literal' && right.type === 'Literal') {
        const result = this.evaluateConstantOperation(left.value, ast.operator, right.value);
        return {
          type: 'Literal',
          value: result,
          raw: String(result)
        };
      }
      
      return { ...ast, left, right };
    }
    
    return ast;
  }
  
  algebraicSimplification(ast) {
    // Apply algebraic rules (x + 0 = x, x * 1 = x, etc.)
    if (ast.type === 'BinaryExpression') {
      const left = this.algebraicSimplification(ast.left);
      const right = this.algebraicSimplification(ast.right);
      
      // x + 0 = x
      if (ast.operator === '+' && right.type === 'Literal' && right.value === 0) {
        return left;
      }
      
      // 0 + x = x
      if (ast.operator === '+' && left.type === 'Literal' && left.value === 0) {
        return right;
      }
      
      // x * 1 = x
      if (ast.operator === '*' && right.type === 'Literal' && right.value === 1) {
        return left;
      }
      
      // 1 * x = x
      if (ast.operator === '*' && left.type === 'Literal' && left.value === 1) {
        return right;
      }
      
      return { ...ast, left, right };
    }
    
    return ast;
  }
  
  evaluateConstantOperation(left, operator, right) {
    switch (operator) {
      case '+': return left + right;
      case '-': return left - right;
      case '*': return left * right;
      case '/': return left / right;
      case '>': return left > right;
      case '<': return left < right;
      case '>=': return left >= right;
      case '<=': return left <= right;
      case '=': return left === right;
      case '!=': return left !== right;
      default: return null;
    }
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Выполнение выражения
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация выражения
- [`GET /api/v1/expressions/functions`](./list-functions.md) - Список функций
- [`POST /api/v1/expressions/test`](./test-expression.md) - Тестирование выражений
