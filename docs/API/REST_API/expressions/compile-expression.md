# POST /api/v1/expressions/compile

## Описание
Компиляция FEEL выражения в оптимизированное представление для быстрого выполнения. Полезно для выражений, которые будут выполняться многократно.

## URL
```
POST /api/v1/expressions/compile
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
- `expression` (string): FEEL выражение для компиляции

### Опциональные поля
- `optimization_level` (string): Уровень оптимизации (`none`, `basic`, `aggressive`) (по умолчанию: `basic`)
- `target_context` (object): Ожидаемый контекст для оптимизации
- `cache_compiled` (boolean): Кэшировать скомпилированное выражение (по умолчанию: true)

## Примеры запросов

### Базовая компиляция
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/compile" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "order.amount * (1 - discount) + shipping_cost"
  }'
```

### Компиляция с агрессивной оптимизацией
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/compile" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "sum(items[price > threshold].price * quantity)",
    "optimization_level": "aggressive",
    "target_context": {
      "items": [
        {"price": 100, "quantity": 2},
        {"price": 50, "quantity": 1}
      ],
      "threshold": 75
    }
  }'
```

### JavaScript
```javascript
const compileRequest = {
  expression: 'if customer.type = "VIP" and order.amount > 500 then order.amount * 0.15 else 0',
  optimization_level: 'basic',
  target_context: {
    customer: { type: 'string' },
    order: { amount: 'number' }
  },
  cache_compiled: true
};

const response = await fetch('/api/v1/expressions/compile', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(compileRequest)
});

const result = await response.json();
console.log('Compiled expression:', result.data);
```

## Ответы

### 200 OK - Успешная компиляция
```json
{
  "success": true,
  "data": {
    "expression": "order.amount * (1 - discount) + shipping_cost",
    "compiled_id": "comp_1641998404200_abc123",
    "compilation_successful": true,
    "optimization_level": "basic",
    "compilation_time_ms": 45,
    "optimizations_applied": [
      "constant_folding",
      "dead_code_elimination",
      "common_subexpression_elimination"
    ],
    "performance_metrics": {
      "original_complexity": 8,
      "optimized_complexity": 5,
      "expected_speedup": "2.3x",
      "memory_usage_reduction": "30%"
    },
    "compiled_bytecode": {
      "instructions": [
        {"op": "LOAD_VAR", "var": "order.amount"},
        {"op": "LOAD_CONST", "value": 1},
        {"op": "LOAD_VAR", "var": "discount"},
        {"op": "SUB"},
        {"op": "MUL"},
        {"op": "LOAD_VAR", "var": "shipping_cost"},
        {"op": "ADD"}
      ],
      "constants": [1],
      "variables": ["order.amount", "discount", "shipping_cost"]
    },
    "cache_info": {
      "cached": true,
      "cache_key": "expr_hash_789xyz",
      "ttl_seconds": 3600
    },
    "metadata": {
      "return_type": "number",
      "pure_function": true,
      "deterministic": true,
      "side_effects": false
    }
  },
  "request_id": "req_1641998404200"
}
```

### 200 OK - Агрессивная оптимизация
```json
{
  "success": true,
  "data": {
    "expression": "sum(items[price > threshold].price * quantity)",
    "compiled_id": "comp_1641998404201_def456",
    "compilation_successful": true,
    "optimization_level": "aggressive",
    "compilation_time_ms": 120,
    "optimizations_applied": [
      "constant_folding",
      "loop_unrolling",
      "vectorization", 
      "filter_pushdown",
      "expression_inlining"
    ],
    "performance_metrics": {
      "original_complexity": 15,
      "optimized_complexity": 6,
      "expected_speedup": "4.7x",
      "memory_usage_reduction": "45%"
    },
    "compiled_bytecode": {
      "instructions": [
        {"op": "LOAD_VAR", "var": "items"},
        {"op": "FILTER_MAP", "filter_expr": "price > threshold", "map_expr": "price * quantity"},
        {"op": "SUM"}
      ],
      "optimized_filters": [
        {
          "original": "items[price > threshold]",
          "optimized": "vectorized_filter(items, threshold)",
          "speedup": "3.2x"
        }
      ]
    },
    "compilation_warnings": [
      {
        "type": "AGGRESSIVE_OPTIMIZATION",
        "message": "Loop unrolling applied - may increase memory usage for large lists",
        "suggestion": "Consider using 'basic' optimization for very large datasets"
      }
    ]
  },
  "request_id": "req_1641998404201"
}
```

### 400 Bad Request - Ошибка компиляции
```json
{
  "success": false,
  "error": {
    "code": "COMPILATION_ERROR",
    "message": "Expression cannot be compiled",
    "details": {
      "expression": "x + y *",
      "compilation_stage": "PARSING",
      "error_position": 7,
      "error_description": "Unexpected end of expression during compilation",
      "suggestions": [
        "Fix syntax errors before compilation",
        "Use validate endpoint to check syntax first"
      ],
      "partial_compilation": {
        "compiled_until": "x + y",
        "remaining": "*"
      }
    }
  },
  "request_id": "req_1641998404202"
}
```

## Compilation Optimization Levels

### None
```yaml
Description: No optimizations applied
Use Case: Debug mode, preserve original structure
Performance: No improvement
Compilation Time: Fastest
```

### Basic (Default)
```yaml
Optimizations:
  - Constant folding
  - Dead code elimination
  - Common subexpression elimination
  - Basic type inference

Use Case: General purpose optimization
Performance: 1.5x - 3x speedup
Compilation Time: Fast
```

### Aggressive
```yaml
Optimizations:
  - All basic optimizations
  - Loop unrolling
  - Vectorization
  - Filter pushdown
  - Expression inlining
  - Advanced constant propagation

Use Case: High-performance scenarios
Performance: 3x - 10x speedup
Compilation Time: Slower
Memory Usage: May be higher
```

## Использование

### Expression Compiler Manager
```javascript
class ExpressionCompiler {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.compiledCache = new Map();
  }
  
  async compile(expression, options = {}) {
    const cacheKey = this.generateCacheKey(expression, options);
    
    if (this.compiledCache.has(cacheKey)) {
      return this.compiledCache.get(cacheKey);
    }
    
    const response = await fetch('/api/v1/expressions/compile', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        expression,
        ...options
      })
    });
    
    const result = await response.json();
    
    if (result.success) {
      this.compiledCache.set(cacheKey, result.data);
      
      // Set up cache expiration
      setTimeout(() => {
        this.compiledCache.delete(cacheKey);
      }, (result.data.cache_info?.ttl_seconds || 3600) * 1000);
    }
    
    return result.data;
  }
  
  async compileForPerformance(expression, targetContext) {
    // Analyze expression complexity and choose optimization level
    const complexity = this.analyzeComplexity(expression);
    
    let optimizationLevel = 'basic';
    if (complexity > 10) {
      optimizationLevel = 'aggressive';
    } else if (complexity < 3) {
      optimizationLevel = 'none';
    }
    
    return await this.compile(expression, {
      optimization_level: optimizationLevel,
      target_context: targetContext,
      cache_compiled: true
    });
  }
  
  analyzeComplexity(expression) {
    // Simple complexity analysis
    let complexity = 0;
    
    // Count operators
    complexity += (expression.match(/[\+\-\*\/\>\<\=]/g) || []).length;
    
    // Count function calls
    complexity += (expression.match(/\w+\(/g) || []).length * 2;
    
    // Count conditional expressions
    complexity += (expression.match(/if\s+.*\s+then/g) || []).length * 3;
    
    // Count list operations
    complexity += (expression.match(/\[.*\]/g) || []).length * 2;
    
    return complexity;
  }
  
  generateCacheKey(expression, options) {
    return `${expression}_${JSON.stringify(options)}`;
  }
  
  async batchCompile(expressions, options = {}) {
    const results = await Promise.all(
      expressions.map(expr => this.compile(expr, options))
    );
    
    return results.map((result, index) => ({
      expression: expressions[index],
      compilation: result
    }));
  }
  
  getCompilationStats() {
    const stats = {
      total_compiled: this.compiledCache.size,
      cache_utilization: 0,
      average_speedup: 0,
      memory_savings: 0
    };
    
    const compilations = Array.from(this.compiledCache.values());
    
    if (compilations.length > 0) {
      const speedups = compilations
        .filter(c => c.performance_metrics?.expected_speedup)
        .map(c => parseFloat(c.performance_metrics.expected_speedup));
      
      if (speedups.length > 0) {
        stats.average_speedup = speedups.reduce((a, b) => a + b, 0) / speedups.length;
      }
      
      const memorySavings = compilations
        .filter(c => c.performance_metrics?.memory_usage_reduction)
        .map(c => parseFloat(c.performance_metrics.memory_usage_reduction));
      
      if (memorySavings.length > 0) {
        stats.memory_savings = memorySavings.reduce((a, b) => a + b, 0) / memorySavings.length;
      }
    }
    
    return stats;
  }
}

// Использование
const compiler = new ExpressionCompiler('your-api-key');

// Компиляция одного выражения
const compiled = await compiler.compile(
  'sum(items[price > 100].price * quantity)',
  { optimization_level: 'aggressive' }
);

console.log('Compilation result:', compiled);
console.log('Expected speedup:', compiled.performance_metrics.expected_speedup);

// Компиляция для производительности
const performanceCompiled = await compiler.compileForPerformance(
  'complex_calculation(a, b, c) + if d > 10 then e * f else g',
  { a: 'number', b: 'number', c: 'number', d: 'number', e: 'number', f: 'number', g: 'number' }
);

// Пакетная компиляция
const expressions = [
  'x + y',
  'if condition then value1 else value2',
  'sum(numbers)'
];

const batchResults = await compiler.batchCompile(expressions, {
  optimization_level: 'basic'
});

// Статистика компиляций
const stats = compiler.getCompilationStats();
console.log('Compilation stats:', stats);
```

### High-Performance Expression Engine
```javascript
class HighPerformanceExpressionEngine {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.compiler = new ExpressionCompiler(apiKey);
    this.executionCache = new Map();
    this.metrics = {
      compilations: 0,
      executions: 0,
      cache_hits: 0,
      total_speedup: 0
    };
  }
  
  async prepareExpression(expression, targetContext) {
    // Compile expression for optimal performance
    const compiled = await this.compiler.compileForPerformance(expression, targetContext);
    
    this.metrics.compilations++;
    
    return {
      compiled_id: compiled.compiled_id,
      expression,
      optimizations: compiled.optimizations_applied,
      expected_speedup: compiled.performance_metrics.expected_speedup
    };
  }
  
  async executeCompiled(compiledId, context) {
    // In a real implementation, this would execute the compiled bytecode
    // For this example, we'll simulate fast execution
    
    const startTime = Date.now();
    
    // Simulate compiled execution (much faster than interpretation)
    const result = await this.simulateCompiledExecution(compiledId, context);
    
    const executionTime = Date.now() - startTime;
    
    this.metrics.executions++;
    this.updatePerformanceMetrics(executionTime, compiledId);
    
    return {
      result,
      execution_time_ms: executionTime,
      compiled_execution: true
    };
  }
  
  async simulateCompiledExecution(compiledId, context) {
    // This would be replaced with actual bytecode execution
    // Simulating faster execution with compiled code
    
    await new Promise(resolve => setTimeout(resolve, 1)); // Simulated fast execution
    
    // Return simulated result
    return Math.random() * 100;
  }
  
  async prepareBatchExpressions(expressions, commonContext) {
    const preparations = await Promise.all(
      expressions.map(expr => this.prepareExpression(expr, commonContext))
    );
    
    return {
      prepared_count: preparations.length,
      preparations,
      total_expected_speedup: preparations.reduce((sum, prep) => {
        const speedup = parseFloat(prep.expected_speedup) || 1;
        return sum + speedup;
      }, 0) / preparations.length
    };
  }
  
  async executeBatch(compiledIds, contexts) {
    const results = await Promise.all(
      compiledIds.map((id, index) => this.executeCompiled(id, contexts[index]))
    );
    
    return {
      results,
      batch_execution_time_ms: results.reduce((sum, r) => sum + r.execution_time_ms, 0),
      average_execution_time_ms: results.reduce((sum, r) => sum + r.execution_time_ms, 0) / results.length
    };
  }
  
  updatePerformanceMetrics(executionTime, compiledId) {
    // Update performance tracking
    const compilation = this.compiler.compiledCache.get(compiledId);
    if (compilation && compilation.performance_metrics.expected_speedup) {
      const expectedSpeedup = parseFloat(compilation.performance_metrics.expected_speedup);
      this.metrics.total_speedup += expectedSpeedup;
    }
  }
  
  getPerformanceReport() {
    return {
      total_compilations: this.metrics.compilations,
      total_executions: this.metrics.executions,
      cache_hit_rate: this.metrics.executions > 0 ? 
        (this.metrics.cache_hits / this.metrics.executions) * 100 : 0,
      average_speedup: this.metrics.executions > 0 ? 
        this.metrics.total_speedup / this.metrics.executions : 0,
      compilation_cache_size: this.compiler.compiledCache.size,
      execution_cache_size: this.executionCache.size
    };
  }
  
  async optimizeForWorkload(expressions, sampleContexts) {
    // Analyze workload and optimize compilation strategy
    const analysis = {
      expression_count: expressions.length,
      unique_expressions: new Set(expressions).size,
      complexity_distribution: {},
      recommended_strategy: null
    };
    
    // Analyze complexity distribution
    const complexities = expressions.map(expr => this.compiler.analyzeComplexity(expr));
    analysis.complexity_distribution = {
      simple: complexities.filter(c => c < 5).length,
      medium: complexities.filter(c => c >= 5 && c < 15).length,
      complex: complexities.filter(c => c >= 15).length
    };
    
    // Recommend strategy
    if (analysis.complexity_distribution.complex > expressions.length * 0.3) {
      analysis.recommended_strategy = 'aggressive_compilation';
    } else if (analysis.unique_expressions < expressions.length * 0.5) {
      analysis.recommended_strategy = 'heavy_caching';
    } else {
      analysis.recommended_strategy = 'balanced_approach';
    }
    
    return analysis;
  }
}

// Использование
const engine = new HighPerformanceExpressionEngine('your-api-key');

// Подготовка выражений для высокопроизводительного выполнения
const expressions = [
  'sum(orders[amount > 100].amount * tax_rate)',
  'if customer.type = "VIP" then discount_rate * 2 else discount_rate',
  'mean(sales[date >= start_date and date <= end_date].amount)'
];

const commonContext = {
  orders: [{ amount: 150 }, { amount: 75 }, { amount: 200 }],
  tax_rate: 0.08,
  customer: { type: 'VIP' },
  discount_rate: 0.1,
  sales: [{ amount: 100, date: '2025-01-01' }],
  start_date: '2025-01-01',
  end_date: '2025-01-31'
};

// Анализ рабочей нагрузки
const workloadAnalysis = await engine.optimizeForWorkload(expressions, [commonContext]);
console.log('Workload analysis:', workloadAnalysis);

// Подготовка пакета выражений
const batchPreparation = await engine.prepareBatchExpressions(expressions, commonContext);
console.log('Batch preparation:', batchPreparation);

// Выполнение подготовленных выражений
const contexts = [commonContext, commonContext, commonContext];
const compiledIds = batchPreparation.preparations.map(p => p.compiled_id);
const batchResults = await engine.executeBatch(compiledIds, contexts);

console.log('Batch execution results:', batchResults);

// Отчет о производительности
const performanceReport = engine.getPerformanceReport();
console.log('Performance report:', performanceReport);
```

## Преимущества компиляции

### Performance Benefits
1. **Faster Execution**: 2x-10x speedup depending on expression complexity
2. **Memory Efficiency**: Reduced memory allocations during execution
3. **CPU Optimization**: Better instruction-level optimization

### Use Cases
1. **Batch Processing**: Processing thousands of records with same expression
2. **Real-time Systems**: Low-latency expression evaluation
3. **Repeated Execution**: Expressions executed multiple times
4. **Complex Calculations**: Heavy computational expressions

### Best Practices
1. Compile expressions that will be used repeatedly
2. Use appropriate optimization level for your use case
3. Profile before and after compilation
4. Cache compiled expressions
5. Monitor compilation overhead vs execution gains

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Выполнение скомпилированных выражений
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация перед компиляцией
- [`GET /api/v1/expressions/stats`](./get-expression-stats.md) - Статистика компиляций
