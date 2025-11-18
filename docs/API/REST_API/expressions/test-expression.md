# POST /api/v1/expressions/test

## Описание
Тестирование FEEL выражения с набором тестовых случаев. Позволяет проверить корректность выражения на различных входных данных.

## URL
```
POST /api/v1/expressions/test
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
- `expression` (string): FEEL выражение для тестирования
- `test_cases` (array): Массив тестовых случаев

### Структура test_case
- `name` (string): Название тестового случая
- `context` (object): Контекст данных для тестирования
- `expected_result` (any): Ожидаемый результат
- `description` (string, optional): Описание тест-кейса

### Опциональные поля
- `tolerance` (number): Допустимая погрешность для числовых сравнений (по умолчанию: 0.001)
- `strict_comparison` (boolean): Строгое сравнение результатов (по умолчанию: false)

## Примеры запросов

### Тестирование арифметического выражения
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/test" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "price * quantity * (1 - discount)",
    "test_cases": [
      {
        "name": "Standard calculation",
        "context": {
          "price": 100,
          "quantity": 2,
          "discount": 0.1
        },
        "expected_result": 180,
        "description": "Basic price calculation with 10% discount"
      },
      {
        "name": "Zero discount",
        "context": {
          "price": 50,
          "quantity": 3,
          "discount": 0
        },
        "expected_result": 150,
        "description": "Calculation without discount"
      }
    ]
  }'
```

### Тестирование условного выражения
```bash
curl -X POST "http://localhost:27555/api/v1/expressions/test" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "expression": "if age >= 18 then \"adult\" else \"minor\"",
    "test_cases": [
      {
        "name": "Adult case",
        "context": { "age": 25 },
        "expected_result": "adult"
      },
      {
        "name": "Minor case", 
        "context": { "age": 16 },
        "expected_result": "minor"
      },
      {
        "name": "Boundary case",
        "context": { "age": 18 },
        "expected_result": "adult"
      }
    ],
    "strict_comparison": true
  }'
```

### JavaScript
```javascript
const testRequest = {
  expression: 'order.total > 100 and customer.type = "VIP"',
  test_cases: [
    {
      name: 'VIP with large order',
      context: {
        order: { total: 150 },
        customer: { type: 'VIP' }
      },
      expected_result: true,
      description: 'VIP customer with qualifying order amount'
    },
    {
      name: 'VIP with small order',
      context: {
        order: { total: 50 },
        customer: { type: 'VIP' }
      },
      expected_result: false,
      description: 'VIP customer with non-qualifying order amount'
    },
    {
      name: 'Regular customer',
      context: {
        order: { total: 200 },
        customer: { type: 'REGULAR' }
      },
      expected_result: false,
      description: 'Non-VIP customer regardless of order amount'
    }
  ]
};

const response = await fetch('/api/v1/expressions/test', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(testRequest)
});

const result = await response.json();
```

## Ответы

### 200 OK - Все тесты прошли
```json
{
  "success": true,
  "data": {
    "expression": "price * quantity * (1 - discount)",
    "test_results": [
      {
        "name": "Standard calculation",
        "status": "PASSED",
        "expected": 180,
        "actual": 180,
        "execution_time_ms": 3,
        "context_used": {
          "price": 100,
          "quantity": 2,
          "discount": 0.1
        }
      },
      {
        "name": "Zero discount",
        "status": "PASSED",
        "expected": 150,
        "actual": 150,
        "execution_time_ms": 2,
        "context_used": {
          "price": 50,
          "quantity": 3,
          "discount": 0
        }
      }
    ],
    "summary": {
      "total_tests": 2,
      "passed": 2,
      "failed": 0,
      "skipped": 0,
      "success_rate": 100,
      "total_execution_time_ms": 5,
      "average_execution_time_ms": 2.5
    },
    "overall_status": "PASSED"
  },
  "request_id": "req_1641998404000"
}
```

### 200 OK - Тесты с ошибками
```json
{
  "success": true,
  "data": {
    "expression": "if age >= 18 then \"adult\" else \"minor\"",
    "test_results": [
      {
        "name": "Adult case",
        "status": "PASSED",
        "expected": "adult",
        "actual": "adult",
        "execution_time_ms": 4
      },
      {
        "name": "Minor case",
        "status": "PASSED",
        "expected": "minor",
        "actual": "minor",
        "execution_time_ms": 3
      },
      {
        "name": "Edge case",
        "status": "FAILED",
        "expected": "minor",
        "actual": "adult",
        "execution_time_ms": 4,
        "context_used": {
          "age": 18
        },
        "failure_reason": "Expected 'minor' but got 'adult'",
        "difference": {
          "type": "VALUE_MISMATCH",
          "expected_type": "string",
          "actual_type": "string",
          "message": "Values do not match"
        }
      }
    ],
    "summary": {
      "total_tests": 3,
      "passed": 2,
      "failed": 1,
      "success_rate": 66.67,
      "total_execution_time_ms": 11,
      "average_execution_time_ms": 3.67
    },
    "overall_status": "FAILED",
    "failed_tests": [
      {
        "name": "Edge case",
        "reason": "Expected 'minor' but got 'adult'",
        "suggestion": "Check boundary condition for age >= 18"
      }
    ]
  },
  "request_id": "req_1641998404001"
}
```

### 200 OK - Тесты с ошибками выполнения
```json
{
  "success": true,
  "data": {
    "expression": "order.customer.name",
    "test_results": [
      {
        "name": "Valid customer",
        "status": "PASSED",
        "expected": "John Doe",
        "actual": "John Doe",
        "execution_time_ms": 5
      },
      {
        "name": "Missing customer",
        "status": "ERROR",
        "expected": "Unknown",
        "execution_time_ms": 2,
        "error": {
          "type": "RUNTIME_ERROR",
          "message": "Cannot read property 'name' of null",
          "context_path": "order.customer",
          "suggestion": "Add null check: order.customer?.name"
        }
      }
    ],
    "summary": {
      "total_tests": 2,
      "passed": 1,
      "failed": 0,
      "errors": 1,
      "success_rate": 50,
      "total_execution_time_ms": 7
    },
    "overall_status": "ERROR"
  },
  "request_id": "req_1641998404002"
}
```

## Test Status Types

### PASSED
Тест выполнен успешно, результат соответствует ожиданию

### FAILED
Тест выполнен, но результат не соответствует ожиданию

### ERROR
Ошибка выполнения во время тестирования

### SKIPPED
Тест пропущен (например, из-за зависимостей)

## Использование

### Automated Test Suite
```javascript
class ExpressionTestSuite {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.testSuites = new Map();
  }
  
  addTestSuite(name, expression, testCases) {
    this.testSuites.set(name, {
      expression,
      test_cases: testCases,
      last_run: null,
      results: null
    });
  }
  
  async runTestSuite(name) {
    const suite = this.testSuites.get(name);
    if (!suite) {
      throw new Error(`Test suite '${name}' not found`);
    }
    
    const response = await fetch('/api/v1/expressions/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        expression: suite.expression,
        test_cases: suite.test_cases
      })
    });
    
    const result = await response.json();
    
    suite.last_run = Date.now();
    suite.results = result.data;
    
    return result.data;
  }
  
  async runAllTestSuites() {
    const results = new Map();
    
    for (const [name, suite] of this.testSuites) {
      try {
        const result = await this.runTestSuite(name);
        results.set(name, result);
      } catch (error) {
        results.set(name, {
          overall_status: 'ERROR',
          error: error.message
        });
      }
    }
    
    return this.generateTestReport(results);
  }
  
  generateTestReport(results) {
    let totalTests = 0;
    let totalPassed = 0;
    let totalFailed = 0;
    let totalErrors = 0;
    
    const suiteResults = [];
    
    for (const [name, result] of results) {
      if (result.summary) {
        totalTests += result.summary.total_tests;
        totalPassed += result.summary.passed || 0;
        totalFailed += result.summary.failed || 0;
        totalErrors += result.summary.errors || 0;
      }
      
      suiteResults.push({
        suite_name: name,
        status: result.overall_status,
        tests: result.summary?.total_tests || 0,
        passed: result.summary?.passed || 0,
        failed: result.summary?.failed || 0,
        errors: result.summary?.errors || 0,
        success_rate: result.summary?.success_rate || 0
      });
    }
    
    return {
      summary: {
        total_suites: results.size,
        total_tests: totalTests,
        total_passed: totalPassed,
        total_failed: totalFailed,
        total_errors: totalErrors,
        overall_success_rate: totalTests > 0 ? (totalPassed / totalTests) * 100 : 0
      },
      suites: suiteResults,
      timestamp: Date.now()
    };
  }
  
  getFailedTests() {
    const failedTests = [];
    
    for (const [suiteName, suite] of this.testSuites) {
      if (suite.results && suite.results.failed_tests) {
        suite.results.failed_tests.forEach(test => {
          failedTests.push({
            suite: suiteName,
            test: test.name,
            reason: test.reason,
            suggestion: test.suggestion
          });
        });
      }
    }
    
    return failedTests;
  }
}

// Использование
const testSuite = new ExpressionTestSuite('your-api-key');

// Добавление тестовых наборов
testSuite.addTestSuite('Price Calculator', 
  'price * quantity * (1 - discount)', 
  [
    {
      name: 'Basic calculation',
      context: { price: 100, quantity: 2, discount: 0.1 },
      expected_result: 180
    },
    {
      name: 'No discount',
      context: { price: 50, quantity: 1, discount: 0 },
      expected_result: 50
    }
  ]
);

testSuite.addTestSuite('Age Verification',
  'if age >= 18 then "adult" else "minor"',
  [
    {
      name: 'Adult',
      context: { age: 25 },
      expected_result: 'adult'
    },
    {
      name: 'Minor',
      context: { age: 16 },
      expected_result: 'minor'
    }
  ]
);

// Запуск всех тестов
const report = await testSuite.runAllTestSuites();
console.log('Test Report:', report);

// Получение неудачных тестов
const failedTests = testSuite.getFailedTests();
console.log('Failed Tests:', failedTests);
```

### Business Rules Testing
```javascript
class BusinessRulesTester {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async testDiscountRules() {
    const rules = {
      'VIP Discount': {
        expression: 'if customer.type = "VIP" and order.amount > 500 then order.amount * 0.15 else 0',
        test_cases: [
          {
            name: 'VIP with large order',
            context: {
              customer: { type: 'VIP' },
              order: { amount: 1000 }
            },
            expected_result: 150
          },
          {
            name: 'VIP with small order',
            context: {
              customer: { type: 'VIP' },
              order: { amount: 300 }
            },
            expected_result: 0
          },
          {
            name: 'Regular customer',
            context: {
              customer: { type: 'REGULAR' },
              order: { amount: 1000 }
            },
            expected_result: 0
          }
        ]
      },
      'Bulk Discount': {
        expression: 'if order.quantity >= 10 then order.amount * 0.1 else 0',
        test_cases: [
          {
            name: 'Bulk order',
            context: {
              order: { quantity: 15, amount: 500 }
            },
            expected_result: 50
          },
          {
            name: 'Small order',
            context: {
              order: { quantity: 5, amount: 500 }
            },
            expected_result: 0
          }
        ]
      }
    };
    
    const results = {};
    
    for (const [ruleName, rule] of Object.entries(rules)) {
      const response = await fetch('/api/v1/expressions/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify(rule)
      });
      
      results[ruleName] = await response.json();
    }
    
    return this.analyzeBusinessRulesResults(results);
  }
  
  analyzeBusinessRulesResults(results) {
    const analysis = {
      total_rules: Object.keys(results).length,
      passed_rules: 0,
      failed_rules: 0,
      rule_details: []
    };
    
    for (const [ruleName, result] of Object.entries(results)) {
      const ruleStatus = result.data.overall_status === 'PASSED' ? 'PASSED' : 'FAILED';
      
      if (ruleStatus === 'PASSED') {
        analysis.passed_rules++;
      } else {
        analysis.failed_rules++;
      }
      
      analysis.rule_details.push({
        rule_name: ruleName,
        status: ruleStatus,
        success_rate: result.data.summary.success_rate,
        failed_tests: result.data.failed_tests || []
      });
    }
    
    analysis.overall_health = (analysis.passed_rules / analysis.total_rules) * 100;
    
    return analysis;
  }
}
```

### Regression Testing
```javascript
class RegressionTester {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baselineResults = new Map();
  }
  
  async createBaseline(expressions) {
    for (const [name, config] of Object.entries(expressions)) {
      const result = await this.runTest(config.expression, config.test_cases);
      this.baselineResults.set(name, result);
    }
    
    console.log(`Created baseline with ${this.baselineResults.size} expressions`);
  }
  
  async runRegressionTest(expressions) {
    const regressionResults = {
      total_expressions: Object.keys(expressions).length,
      regressions: [],
      improvements: [],
      unchanged: [],
      new_expressions: []
    };
    
    for (const [name, config] of Object.entries(expressions)) {
      const currentResult = await this.runTest(config.expression, config.test_cases);
      const baseline = this.baselineResults.get(name);
      
      if (!baseline) {
        regressionResults.new_expressions.push({
          name,
          status: currentResult.overall_status,
          success_rate: currentResult.summary.success_rate
        });
        continue;
      }
      
      const comparison = this.compareResults(baseline, currentResult);
      
      switch (comparison.type) {
        case 'REGRESSION':
          regressionResults.regressions.push({
            name,
            baseline_success_rate: baseline.summary.success_rate,
            current_success_rate: currentResult.summary.success_rate,
            difference: comparison.difference,
            failing_tests: comparison.failing_tests
          });
          break;
          
        case 'IMPROVEMENT':
          regressionResults.improvements.push({
            name,
            baseline_success_rate: baseline.summary.success_rate,
            current_success_rate: currentResult.summary.success_rate,
            difference: comparison.difference
          });
          break;
          
        case 'UNCHANGED':
          regressionResults.unchanged.push({
            name,
            success_rate: currentResult.summary.success_rate
          });
          break;
      }
    }
    
    return regressionResults;
  }
  
  compareResults(baseline, current) {
    const baselineRate = baseline.summary.success_rate;
    const currentRate = current.summary.success_rate;
    const difference = currentRate - baselineRate;
    
    if (difference < -5) { // 5% tolerance
      return {
        type: 'REGRESSION',
        difference,
        failing_tests: this.findNewFailures(baseline, current)
      };
    } else if (difference > 5) {
      return {
        type: 'IMPROVEMENT',
        difference
      };
    } else {
      return {
        type: 'UNCHANGED',
        difference
      };
    }
  }
  
  findNewFailures(baseline, current) {
    const baselineFailures = new Set(
      baseline.failed_tests?.map(t => t.name) || []
    );
    const currentFailures = new Set(
      current.failed_tests?.map(t => t.name) || []
    );
    
    return Array.from(currentFailures).filter(name => !baselineFailures.has(name));
  }
  
  async runTest(expression, testCases) {
    const response = await fetch('/api/v1/expressions/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        expression,
        test_cases: testCases
      })
    });
    
    const result = await response.json();
    return result.data;
  }
}
```

## Связанные endpoints
- [`POST /api/v1/expressions/eval`](./eval-expression.md) - Выполнение тестируемых выражений
- [`POST /api/v1/expressions/validate`](./validate-expression.md) - Валидация перед тестированием
- [`GET /api/v1/expressions/functions`](./list-functions.md) - Функции для тестирования
