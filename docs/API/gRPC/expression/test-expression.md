# TestExpression

## Описание
Тестирует FEEL выражение с набором тестовых случаев, позволяя проверить корректность работы выражения на различных входных данных.

## Синтаксис
```protobuf
rpc TestExpression(TestExpressionRequest) returns (TestExpressionResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### TestExpressionRequest
```protobuf
message TestExpressionRequest {
  string expression = 1;           // FEEL выражение для тестирования
  repeated TestCase test_cases = 2; // Набор тестовых случаев
  string tenant_id = 3;            // ID тенанта
}

message TestCase {
  string name = 1;             // Название тестового случая
  string context = 2;          // JSON контекст для теста
  string expected_result = 3;  // Ожидаемый результат
  string expected_type = 4;    // Ожидаемый тип результата
}
```

## Параметры ответа

### TestExpressionResponse
```protobuf
message TestExpressionResponse {
  repeated TestResult results = 1;  // Результаты тестов
  int32 passed_count = 2;          // Количество пройденных тестов
  int32 failed_count = 3;          // Количество провалившихся тестов
  bool all_passed = 4;             // Все тесты прошли успешно
  string summary = 5;              // Сводка тестирования
}

message TestResult {
  string test_name = 1;        // Название теста
  bool passed = 2;             // Прошел ли тест
  string actual_result = 3;    // Фактический результат
  string expected_result = 4;  // Ожидаемый результат
  string error_message = 5;    // Сообщение об ошибке
  string actual_type = 6;      // Фактический тип результата
  string expected_type = 7;    // Ожидаемый тип результата
}
```

## Примеры использования

### Go
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    
    pb "atom-engine/proto/expression/expressionpb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewExpressionServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Тестирование выражения расчета скидки
    expression := `if age >= 65 then price * 0.2 else if age >= 18 then price * 0.1 else 0`
    
    testCases := []*pb.TestCase{
        {
            Name:           "Пенсионная скидка",
            Context:        `{"age": 70, "price": 1000}`,
            ExpectedResult: "200.0",
            ExpectedType:   "number",
        },
        {
            Name:           "Взрослая скидка",
            Context:        `{"age": 30, "price": 1000}`,
            ExpectedResult: "100.0",
            ExpectedType:   "number",
        },
        {
            Name:           "Детский случай",
            Context:        `{"age": 15, "price": 1000}`,
            ExpectedResult: "0",
            ExpectedType:   "number",
        },
        {
            Name:           "Граничный случай 18",
            Context:        `{"age": 18, "price": 500}`,
            ExpectedResult: "50.0",
            ExpectedType:   "number",
        },
        {
            Name:           "Граничный случай 65",
            Context:        `{"age": 65, "price": 1200}`,
            ExpectedResult: "240.0",
            ExpectedType:   "number",
        },
    }
    
    response, err := client.TestExpression(ctx, &pb.TestExpressionRequest{
        Expression: expression,
        TestCases:  testCases,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("🧪 Тестирование выражения:\n%s\n\n", expression)
    fmt.Printf("📊 Результаты: %d прошло, %d провалилось\n", 
        response.PassedCount, response.FailedCount)
    
    if response.AllPassed {
        fmt.Println("✅ Все тесты прошли успешно!")
    } else {
        fmt.Println("❌ Некоторые тесты провалились")
    }
    
    fmt.Println("\n📋 Детальные результаты:")
    for _, result := range response.Results {
        status := "✅"
        if !result.Passed {
            status = "❌"
        }
        
        fmt.Printf("%s %s\n", status, result.TestName)
        fmt.Printf("   Ожидалось: %s (%s)\n", result.ExpectedResult, result.ExpectedType)
        fmt.Printf("   Получено:  %s (%s)\n", result.ActualResult, result.ActualType)
        
        if !result.Passed && result.ErrorMessage != "" {
            fmt.Printf("   Ошибка: %s\n", result.ErrorMessage)
        }
        fmt.Println()
    }
    
    fmt.Printf("📝 Сводка: %s\n", response.Summary)
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def test_expression(expression, test_cases):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    # Преобразуем тестовые случаи
    proto_test_cases = []
    for case in test_cases:
        proto_case = expression_pb2.TestCase(
            name=case['name'],
            context=json.dumps(case['context']),
            expected_result=str(case['expected_result']),
            expected_type=case.get('expected_type', 'string')
        )
        proto_test_cases.append(proto_case)
    
    request = expression_pb2.TestExpressionRequest(
        expression=expression,
        test_cases=proto_test_cases
    )
    
    try:
        response = stub.TestExpression(request, metadata=metadata)
        
        print(f"🧪 Тестирование: {expression}")
        print(f"📊 Результаты: {response.passed_count} ✅ / {response.failed_count} ❌")
        
        if response.all_passed:
            print("🎉 Все тесты прошли!")
        else:
            print("⚠️  Есть проваленные тесты")
        
        print("\n📋 Детали:")
        for result in response.results:
            status = "✅" if result.passed else "❌"
            print(f"{status} {result.test_name}")
            
            if not result.passed:
                print(f"   Ожидалось: {result.expected_result}")
                print(f"   Получено:  {result.actual_result}")
                if result.error_message:
                    print(f"   Ошибка: {result.error_message}")
            print()
        
        print(f"📝 {response.summary}")
        
        return response.all_passed
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

# Набор тестов для валидации email
def test_email_validation():
    print("📧 Тестирование валидации email\n")
    
    email_expression = 'contains(email, "@") and length(email) >= 5 and not contains(email, " ")'
    
    test_cases = [
        {
            'name': 'Корректный email',
            'context': {'email': 'user@example.com'},
            'expected_result': True,
            'expected_type': 'boolean'
        },
        {
            'name': 'Без собаки',
            'context': {'email': 'userexample.com'},
            'expected_result': False,
            'expected_type': 'boolean'
        },
        {
            'name': 'Слишком короткий',
            'context': {'email': 'a@b'},
            'expected_result': False,
            'expected_type': 'boolean'
        },
        {
            'name': 'С пробелом',
            'context': {'email': 'user name@example.com'},
            'expected_result': False,
            'expected_type': 'boolean'
        },
        {
            'name': 'Пустая строка',
            'context': {'email': ''},
            'expected_result': False,
            'expected_type': 'boolean'
        }
    ]
    
    return test_expression(email_expression, test_cases)

# Тестирование математических функций
def test_math_functions():
    print("🔢 Тестирование математических функций\n")
    
    expressions_and_tests = [
        {
            'expression': 'abs(x)',
            'cases': [
                {'name': 'Положительное', 'context': {'x': 5}, 'expected_result': 5},
                {'name': 'Отрицательное', 'context': {'x': -5}, 'expected_result': 5},
                {'name': 'Ноль', 'context': {'x': 0}, 'expected_result': 0},
            ]
        },
        {
            'expression': 'max(a, b, c)',
            'cases': [
                {'name': 'Первый максимум', 'context': {'a': 10, 'b': 5, 'c': 7}, 'expected_result': 10},
                {'name': 'Последний максимум', 'context': {'a': 3, 'b': 1, 'c': 8}, 'expected_result': 8},
                {'name': 'Равные значения', 'context': {'a': 5, 'b': 5, 'c': 5}, 'expected_result': 5},
            ]
        }
    ]
    
    all_passed = True
    for test_suite in expressions_and_tests:
        print(f"🎯 Тестируем: {test_suite['expression']}")
        passed = test_expression(test_suite['expression'], test_suite['cases'])
        all_passed = all_passed and passed
        print("-" * 40)
    
    return all_passed

# Комплексное тестирование бизнес-правил
class BusinessRulesTester:
    def __init__(self):
        self.test_suites = {}
    
    def add_rule(self, name, expression, test_cases):
        self.test_suites[name] = {
            'expression': expression,
            'test_cases': test_cases
        }
    
    def run_all_tests(self):
        print("🏢 КОМПЛЕКСНОЕ ТЕСТИРОВАНИЕ БИЗНЕС-ПРАВИЛ")
        print("=" * 60)
        
        total_passed = 0
        total_failed = 0
        
        for rule_name, rule_data in self.test_suites.items():
            print(f"\n📋 Правило: {rule_name}")
            print(f"📝 Выражение: {rule_data['expression']}")
            print("-" * 40)
            
            passed = test_expression(rule_data['expression'], rule_data['test_cases'])
            
            if passed:
                total_passed += 1
                print("✅ ПРАВИЛО ПРОШЛО ВСЕ ТЕСТЫ")
            else:
                total_failed += 1
                print("❌ ПРАВИЛО ПРОВАЛИЛО ТЕСТЫ")
            
            print("=" * 40)
        
        print(f"\n📊 ИТОГО: {total_passed} правил прошли, {total_failed} провалили")
        
        return total_failed == 0

if __name__ == "__main__":
    # Создаем тестер бизнес-правил
    tester = BusinessRulesTester()
    
    # Правило одобрения кредита
    tester.add_rule(
        "Одобрение кредита",
        "age >= 21 and age <= 65 and income >= 30000 and credit_score >= 650",
        [
            {'name': 'Идеальный клиент', 'context': {'age': 35, 'income': 50000, 'credit_score': 750}, 'expected_result': True},
            {'name': 'Молодой клиент', 'context': {'age': 20, 'income': 50000, 'credit_score': 750}, 'expected_result': False},
            {'name': 'Низкий доход', 'context': {'age': 35, 'income': 25000, 'credit_score': 750}, 'expected_result': False},
            {'name': 'Плохая кредитная история', 'context': {'age': 35, 'income': 50000, 'credit_score': 600}, 'expected_result': False},
        ]
    )
    
    # Правило расчета скидки VIP
    tester.add_rule(
        "VIP скидка",
        "if vip_level = 'gold' then 0.15 else if vip_level = 'silver' then 0.10 else if vip_level = 'bronze' then 0.05 else 0",
        [
            {'name': 'Gold VIP', 'context': {'vip_level': 'gold'}, 'expected_result': 0.15},
            {'name': 'Silver VIP', 'context': {'vip_level': 'silver'}, 'expected_result': 0.10},
            {'name': 'Bronze VIP', 'context': {'vip_level': 'bronze'}, 'expected_result': 0.05},
            {'name': 'Обычный клиент', 'context': {'vip_level': 'none'}, 'expected_result': 0},
        ]
    )
    
    # Запускаем все тесты
    all_passed = tester.run_all_tests()
    
    if all_passed:
        print("\n🎉 ВСЕ БИЗНЕС-ПРАВИЛА РАБОТАЮТ КОРРЕКТНО!")
    else:
        print("\n🚨 НЕКОТОРЫЕ ПРАВИЛА ТРЕБУЮТ ИСПРАВЛЕНИЯ!")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function testExpression(expression, testCases) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        // Преобразуем тестовые случаи в формат protobuf
        const protoTestCases = testCases.map(testCase => ({
            name: testCase.name,
            context: JSON.stringify(testCase.context),
            expected_result: String(testCase.expected_result),
            expected_type: testCase.expected_type || 'string'
        }));
        
        const request = {
            expression: expression,
            test_cases: protoTestCases
        };
        
        client.testExpression(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log(`🧪 Тестирование: ${expression}`);
            console.log(`📊 ${response.passed_count} ✅ / ${response.failed_count} ❌`);
            
            if (response.all_passed) {
                console.log('🎉 Все тесты прошли!');
            } else {
                console.log('⚠️  Есть проваленные тесты');
            }
            
            console.log('\n📋 Результаты:');
            response.results.forEach(result => {
                const status = result.passed ? '✅' : '❌';
                console.log(`${status} ${result.test_name}`);
                
                if (!result.passed) {
                    console.log(`   Ожидалось: ${result.expected_result} (${result.expected_type})`);
                    console.log(`   Получено:  ${result.actual_result} (${result.actual_type})`);
                    if (result.error_message) {
                        console.log(`   Ошибка: ${result.error_message}`);
                    }
                }
            });
            
            console.log(`\n📝 ${response.summary}\n`);
            
            resolve({
                allPassed: response.all_passed,
                passed: response.passed_count,
                failed: response.failed_count,
                results: response.results
            });
        });
    });
}

// Класс для автоматического тестирования выражений
class ExpressionTestSuite {
    constructor(name) {
        this.name = name;
        this.tests = [];
    }
    
    addTest(expression, testCases) {
        this.tests.push({
            expression,
            testCases
        });
    }
    
    async runAllTests() {
        console.log(`🏗️  Запуск тест-сьюта: ${this.name}`);
        console.log('='.repeat(50));
        
        let totalPassed = 0;
        let totalFailed = 0;
        const failedExpressions = [];
        
        for (let i = 0; i < this.tests.length; i++) {
            const test = this.tests[i];
            console.log(`\n${i + 1}. Тест выражения:`);
            
            try {
                const result = await testExpression(test.expression, test.testCases);
                
                totalPassed += result.passed;
                totalFailed += result.failed;
                
                if (!result.allPassed) {
                    failedExpressions.push(test.expression);
                }
            } catch (error) {
                console.log(`❌ Ошибка тестирования: ${error.message}`);
                totalFailed += test.testCases.length;
                failedExpressions.push(test.expression);
            }
            
            console.log('-'.repeat(40));
        }
        
        console.log(`\n📈 ИТОГИ ТЕСТ-СЬЮТА "${this.name}":`);
        console.log(`   Всего тестов: ${totalPassed + totalFailed}`);
        console.log(`   Прошло: ${totalPassed}`);
        console.log(`   Провалилось: ${totalFailed}`);
        console.log(`   Успешность: ${((totalPassed / (totalPassed + totalFailed)) * 100).toFixed(1)}%`);
        
        if (failedExpressions.length > 0) {
            console.log(`\n🚨 Выражения с ошибками:`);
            failedExpressions.forEach((expr, index) => {
                console.log(`   ${index + 1}. ${expr}`);
            });
        }
        
        return totalFailed === 0;
    }
}

// Демонстрация тестирования различных типов выражений
async function demonstrateExpressionTesting() {
    // Математический тест-сьют
    const mathSuite = new ExpressionTestSuite('Математические операции');
    
    mathSuite.addTest(
        'round(x, n)',
        [
            { name: 'Округление до целого', context: { x: 3.14159 }, expected_result: 3, expected_type: 'number' },
            { name: 'Округление до 2 знаков', context: { x: 3.14159, n: 2 }, expected_result: 3.14, expected_type: 'number' },
            { name: 'Отрицательное число', context: { x: -2.7 }, expected_result: -3, expected_type: 'number' },
        ]
    );
    
    mathSuite.addTest(
        'x * y + z',
        [
            { name: 'Простая арифметика', context: { x: 2, y: 3, z: 4 }, expected_result: 10, expected_type: 'number' },
            { name: 'С нулем', context: { x: 5, y: 0, z: 7 }, expected_result: 7, expected_type: 'number' },
            { name: 'Отрицательные числа', context: { x: -2, y: 3, z: 1 }, expected_result: -5, expected_type: 'number' },
        ]
    );
    
    // Строковый тест-сьют
    const stringSuite = new ExpressionTestSuite('Работа со строками');
    
    stringSuite.addTest(
        'upper(substring(name, 1, 3))',
        [
            { name: 'Обычная строка', context: { name: 'hello world' }, expected_result: 'HEL', expected_type: 'string' },
            { name: 'Короткая строка', context: { name: 'hi' }, expected_result: 'HI', expected_type: 'string' },
            { name: 'Пустая строка', context: { name: '' }, expected_result: '', expected_type: 'string' },
        ]
    );
    
    // Условный тест-сьют
    const conditionalSuite = new ExpressionTestSuite('Условная логика');
    
    conditionalSuite.addTest(
        'if score >= 90 then "A" else if score >= 80 then "B" else if score >= 70 then "C" else "F"',
        [
            { name: 'Отличная оценка', context: { score: 95 }, expected_result: 'A', expected_type: 'string' },
            { name: 'Хорошая оценка', context: { score: 85 }, expected_result: 'B', expected_type: 'string' },
            { name: 'Удовлетворительно', context: { score: 75 }, expected_result: 'C', expected_type: 'string' },
            { name: 'Неудовлетворительно', context: { score: 65 }, expected_result: 'F', expected_type: 'string' },
            { name: 'Граничный случай', context: { score: 90 }, expected_result: 'A', expected_type: 'string' },
        ]
    );
    
    // Запуск всех тест-сьютов
    const suites = [mathSuite, stringSuite, conditionalSuite];
    let allSuitesPass = true;
    
    console.log('🎯 КОМПЛЕКСНОЕ ТЕСТИРОВАНИЕ FEEL ВЫРАЖЕНИЙ');
    console.log('='.repeat(60));
    
    for (const suite of suites) {
        const passed = await suite.runAllTests();
        allSuitesPass = allSuitesPass && passed;
        
        console.log('\n' + '═'.repeat(60));
    }
    
    console.log('\n🏆 ФИНАЛЬНЫЙ РЕЗУЛЬТАТ:');
    if (allSuitesPass) {
        console.log('🎉 ВСЕ ТЕСТ-СЬЮТЫ ПРОШЛИ УСПЕШНО!');
        console.log('   Все FEEL выражения работают корректно.');
    } else {
        console.log('⚠️  НЕКОТОРЫЕ ТЕСТ-СЬЮТЫ ПРОВАЛИЛИСЬ!');
        console.log('   Требуется исправление выражений.');
    }
}

// Быстрое тестирование одного выражения
async function quickTest() {
    console.log('⚡ БЫСТРЫЙ ТЕСТ:\n');
    
    const result = await testExpression(
        'length(name) > 0 and contains(email, "@")',
        [
            { name: 'Валидные данные', context: { name: 'John', email: 'john@test.com' }, expected_result: true },
            { name: 'Пустое имя', context: { name: '', email: 'john@test.com' }, expected_result: false },
            { name: 'Некорректный email', context: { name: 'John', email: 'invalid' }, expected_result: false },
        ]
    );
    
    return result.allPassed;
}

// Основная демонстрация
async function main() {
    try {
        // Быстрый тест
        await quickTest();
        
        console.log('\n' + '█'.repeat(80));
        
        // Полная демонстрация
        await demonstrateExpressionTesting();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Применение

### Unit Testing
```javascript
// Автоматические тесты для BPMN выражений
const testSuite = new ExpressionTestSuite('BPMN Gateway Conditions');
testSuite.addTest('order.total > 1000', testCases);
await testSuite.runAllTests();
```

### Regression Testing
```javascript  
// Проверка после изменений в движке
const regressionTests = loadExistingTests();
await testExpression(expression, regressionTests);
```

### Quality Assurance
```javascript
// Валидация бизнес-правил перед деплоем
const businessRules = loadBusinessRules();
await validateAllRules(businessRules);
```

## Преимущества

### Автоматизация тестирования
- **Пакетное тестирование** множественных сценариев
- **Детальная отчетность** по каждому случаю
- **Статистика прохождения** тестов

### Качество кода
- **Проверка граничных случаев**
- **Валидация типов результатов**
- **Регрессионное тестирование**

### Документирование
- **Примеры использования** выражений
- **Ожидаемое поведение** в различных условиях
- **Покрытие тестами** бизнес-логики

## Связанные методы
- [EvaluateExpression](evaluate-expression.md) - Базовое вычисление
- [ValidateExpression](validate-expression.md) - Валидация синтаксиса  
- [EvaluateBatch](evaluate-batch.md) - Массовое вычисление
