# EvaluateCondition

## Описание
Вычисляет FEEL выражение и возвращает булевый результат. Специализированный метод для условий, гарантирующий возврат true/false значений.

## Синтаксис
```protobuf
rpc EvaluateCondition(EvaluateConditionRequest) returns (EvaluateConditionResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### EvaluateConditionRequest
```protobuf
message EvaluateConditionRequest {
  string condition = 1;   // FEEL условие для вычисления
  string context = 2;     // JSON с переменными
  string tenant_id = 3;   // ID тенанта
}
```

## Параметры ответа

### EvaluateConditionResponse
```protobuf
message EvaluateConditionResponse {
  bool result = 1;          // Булевый результат условия
  bool success = 2;         // Успешность вычисления
  string error_message = 3; // Сообщение об ошибке
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
    
    // Проверка условий доступа
    conditions := []struct {
        condition string
        context   string
        desc      string
    }{
        {
            condition: "age >= 18",
            context:   `{"age": 25}`,
            desc:      "Проверка совершеннолетия",
        },
        {
            condition: "balance > 1000 and status = 'active'",
            context:   `{"balance": 1500, "status": "active"}`,
            desc:      "Проверка баланса и статуса",
        },
        {
            condition: "role = 'admin' or permissions contains 'write'",
            context:   `{"role": "user", "permissions": ["read", "write"]}`,
            desc:      "Проверка прав доступа",
        },
    }
    
    for _, test := range conditions {
        fmt.Printf("🔍 %s\n", test.desc)
        fmt.Printf("   Условие: %s\n", test.condition)
        
        response, err := client.EvaluateCondition(ctx, &pb.EvaluateConditionRequest{
            Condition: test.condition,
            Context:   test.context,
        })
        
        if err != nil {
            fmt.Printf("❌ gRPC Error: %v\n\n", err)
            continue
        }
        
        if response.Success {
            if response.Result {
                fmt.Printf("✅ Условие выполнено: TRUE\n")
            } else {
                fmt.Printf("❌ Условие не выполнено: FALSE\n")
            }
        } else {
            fmt.Printf("💥 Ошибка: %s\n", response.ErrorMessage)
        }
        fmt.Println()
    }
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def evaluate_condition(condition, context=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    context_json = json.dumps(context or {})
    
    request = expression_pb2.EvaluateConditionRequest(
        condition=condition,
        context=context_json
    )
    
    try:
        response = stub.EvaluateCondition(request, metadata=metadata)
        
        if response.success:
            print(f"🔍 {condition}")
            print(f"   ✅ Результат: {'TRUE' if response.result else 'FALSE'}")
            return response.result
        else:
            print(f"❌ Ошибка в условии: {response.error_message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Система валидации данных
class DataValidator:
    def __init__(self):
        self.rules = []
    
    def add_rule(self, name, condition, context=None):
        self.rules.append({
            'name': name,
            'condition': condition,
            'context': context or {}
        })
    
    def validate(self, data):
        results = {}
        all_passed = True
        
        print("🛡️ Запуск валидации данных...\n")
        
        for rule in self.rules:
            # Объединяем контекст правила с данными
            full_context = {**rule['context'], **data}
            
            print(f"📋 Правило: {rule['name']}")
            result = evaluate_condition(rule['condition'], full_context)
            
            results[rule['name']] = result
            if not result:
                all_passed = False
            print()
        
        return all_passed, results

# Пример использования валидатора
if __name__ == "__main__":
    validator = DataValidator()
    
    # Правила валидации пользователя
    validator.add_rule("Возраст", "age >= 18 and age <= 120")
    validator.add_rule("Email", "contains(email, '@') and length(email) >= 5")
    validator.add_rule("Пароль", "length(password) >= 8")
    validator.add_rule("Статус", "status in ['active', 'pending', 'suspended']")
    validator.add_rule("Баланс", "balance >= 0")
    
    # Тестовые данные
    test_users = [
        {
            "name": "John Doe",
            "age": 30,
            "email": "john@example.com",
            "password": "securepass123",
            "status": "active",
            "balance": 1000
        },
        {
            "name": "Jane Smith", 
            "age": 16,  # Невалидный возраст
            "email": "invalid-email",  # Невалидный email
            "password": "123",  # Слабый пароль
            "status": "unknown",  # Невалидный статус
            "balance": -100  # Отрицательный баланс
        }
    ]
    
    for i, user in enumerate(test_users, 1):
        print(f"{'='*50}")
        print(f"👤 Пользователь {i}: {user['name']}")
        print('='*50)
        
        passed, results = validator.validate(user)
        
        if passed:
            print("🎉 Все проверки пройдены! Пользователь валиден.")
        else:
            failed_rules = [rule for rule, result in results.items() if not result]
            print(f"⚠️  Провалено правил: {len(failed_rules)}")
            print(f"   Не прошли: {', '.join(failed_rules)}")
        print()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function evaluateCondition(condition, context = {}) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            condition: condition,
            context: JSON.stringify(context)
        };
        
        client.evaluateCondition(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`🔍 ${condition}`);
                console.log(`   ${response.result ? '✅ TRUE' : '❌ FALSE'}`);
                resolve(response.result);
            } else {
                console.log(`💥 Ошибка: ${response.error_message}`);
                resolve(null);
            }
        });
    });
}

// Система авторизации на основе условий
class AuthorizationEngine {
    constructor() {
        this.policies = new Map();
    }
    
    addPolicy(name, condition) {
        this.policies.set(name, condition);
    }
    
    async checkAccess(user, resource, action) {
        console.log(`🔐 Проверка доступа для ${user.name}`);
        console.log(`   Ресурс: ${resource}, Действие: ${action}\n`);
        
        const context = {
            user: user,
            resource: resource,
            action: action,
            timestamp: new Date().toISOString()
        };
        
        let accessGranted = true;
        const results = [];
        
        for (const [policyName, condition] of this.policies) {
            console.log(`📜 Политика: ${policyName}`);
            
            try {
                const result = await evaluateCondition(condition, context);
                results.push({ policy: policyName, result });
                
                if (!result) {
                    accessGranted = false;
                }
            } catch (error) {
                console.log(`❌ Ошибка в политике ${policyName}: ${error.message}`);
                accessGranted = false;
            }
            
            console.log(); // Пустая строка для читаемости
        }
        
        return { accessGranted, results };
    }
}

// Пример использования системы авторизации
async function demonstrateAuthorization() {
    const authEngine = new AuthorizationEngine();
    
    // Определяем политики безопасности
    authEngine.addPolicy(
        'AdminOnly', 
        'user.role = "admin"'
    );
    
    authEngine.addPolicy(
        'BusinessHours',
        'hour(timestamp) >= 9 and hour(timestamp) <= 17'
    );
    
    authEngine.addPolicy(
        'ResourceOwner',
        'user.id = resource.owner_id or user.role = "admin"'
    );
    
    authEngine.addPolicy(
        'ActiveUser',
        'user.status = "active" and user.verified = true'
    );
    
    // Тестовые пользователи
    const users = [
        {
            id: 1,
            name: "Admin User",
            role: "admin",
            status: "active",
            verified: true
        },
        {
            id: 2, 
            name: "Regular User",
            role: "user",
            status: "active",
            verified: true
        },
        {
            id: 3,
            name: "Inactive User",
            role: "user", 
            status: "suspended",
            verified: false
        }
    ];
    
    const resource = {
        id: 'doc-123',
        owner_id: 2,
        type: 'document'
    };
    
    console.log('🏛️ ДЕМОНСТРАЦИЯ СИСТЕМЫ АВТОРИЗАЦИИ');
    console.log('='.repeat(50));
    
    for (const user of users) {
        console.log(`\n${'─'.repeat(40)}`);
        
        const { accessGranted, results } = await authEngine.checkAccess(
            user, 
            resource, 
            'delete'
        );
        
        console.log(`🎯 ИТОГ для ${user.name}:`);
        console.log(`   ${accessGranted ? '✅ ДОСТУП РАЗРЕШЕН' : '🚫 ДОСТУП ЗАПРЕЩЕН'}`);
        
        const failedPolicies = results.filter(r => !r.result);
        if (failedPolicies.length > 0) {
            console.log(`   💥 Нарушенные политики: ${failedPolicies.map(p => p.policy).join(', ')}`);
        }
        
        console.log();
    }
}

// Простые проверки условий
async function simpleConditionTests() {
    console.log('🎯 Простые проверки условий:\n');
    
    const tests = [
        { condition: '5 > 3', context: {}, desc: 'Арифметическое сравнение' },
        { condition: 'age >= 18', context: { age: 21 }, desc: 'Проверка возраста' },
        { condition: 'name = "John"', context: { name: "John" }, desc: 'Строковое равенство' },
        { condition: 'items contains "apple"', context: { items: ["apple", "banana"] }, desc: 'Содержимое списка' },
        { condition: 'score > 90 and grade = "A"', context: { score: 95, grade: "A" }, desc: 'Сложное условие' }
    ];
    
    for (const test of tests) {
        console.log(`📝 ${test.desc}:`);
        await evaluateCondition(test.condition, test.context);
        console.log();
    }
}

// Основная демонстрация
async function main() {
    try {
        await simpleConditionTests();
        console.log('='.repeat(60));
        await demonstrateAuthorization();
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Применение

### BPMN Gateway Условия
```javascript
// Эксклюзивные шлюзы
await evaluateCondition('order.total > 1000', orderData);
await evaluateCondition('user.vip = true', userData);
```

### Бизнес-правила
```javascript
// Правила одобрения кредита
await evaluateCondition('income > 50000 and credit_score >= 700', applicantData);
```

### Фильтрация данных
```javascript
// Фильтры для отчетов
await evaluateCondition('date >= "2023-01-01" and status = "completed"', recordData);
```

### Авторизация
```javascript
// Проверка прав доступа
await evaluateCondition('user.role = "admin" or resource.owner = user.id', authContext);
```

## Преобразования типов

### Автоматическое приведение
- **Числа**: `"123" → 123`
- **Булевы**: `"true" → true`, `1 → true`, `0 → false`  
- **null/undefined**: `null → false`, `undefined → false`

### Ошибки типов
Не-булевые результаты приводятся к false при ошибках вычисления.

## Связанные методы
- [EvaluateExpression](evaluate-expression.md) - Для не-булевых результатов  
- [ValidateExpression](validate-expression.md) - Проверка синтаксиса условий
- [EvaluateBatch](evaluate-batch.md) - Множественные условия
