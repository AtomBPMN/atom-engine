# EvaluateBatch

## Описание
Выполняет пакетное вычисление нескольких FEEL выражений с общим контекстом переменных. Оптимизирован для массовой обработки выражений.

## Синтаксис
```protobuf
rpc EvaluateBatch(EvaluateBatchRequest) returns (EvaluateBatchResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### EvaluateBatchRequest
```protobuf
message EvaluateBatchRequest {
  repeated string expressions = 1;  // Список FEEL выражений
  string context = 2;              // JSON с общими переменными
  string tenant_id = 3;            // ID тенанта
}
```

## Параметры ответа

### EvaluateBatchResponse
```protobuf
message EvaluateBatchResponse {
  repeated ExpressionResult results = 1;  // Результаты для каждого выражения
  bool overall_success = 2;               // Общий успех операции
  int32 successful_count = 3;             // Количество успешных вычислений
  int32 failed_count = 4;                // Количество неудачных вычислений
}

message ExpressionResult {
  string result = 1;        // JSON результат
  bool success = 2;         // Успешность конкретного вычисления
  string error_message = 3; // Сообщение об ошибке
  string result_type = 4;   // Тип результата
  int32 expression_index = 5; // Индекс выражения в запросе
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
    
    // Пакетная оценка правил
    expressions := []string{
        "age >= 18",
        "score > 85",
        "category = 'premium'",
        "balance > 1000",
        `if age >= 65 then discount * 1.2 else discount`,
    }
    
    response, err := client.EvaluateBatch(ctx, &pb.EvaluateBatchRequest{
        Expressions: expressions,
        Context:     `{"age": 30, "score": 92, "category": "premium", "balance": 1500, "discount": 0.1}`,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("📊 Обработано: %d успешно, %d с ошибками\n", 
        response.SuccessfulCount, response.FailedCount)
    
    for _, result := range response.Results {
        if result.Success {
            fmt.Printf("✅ [%d]: %s (%s)\n", 
                result.ExpressionIndex, result.Result, result.ResultType)
        } else {
            fmt.Printf("❌ [%d]: %s\n", 
                result.ExpressionIndex, result.ErrorMessage)
        }
    }
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def evaluate_batch(expressions, context=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    context_json = json.dumps(context or {})
    
    request = expression_pb2.EvaluateBatchRequest(
        expressions=expressions,
        context=context_json
    )
    
    try:
        response = stub.EvaluateBatch(request, metadata=metadata)
        
        print(f"📊 Результаты: {response.successful_count} успешно, {response.failed_count} ошибок")
        
        results = []
        for result in response.results:
            if result.success:
                value = json.loads(result.result) if result.result_type != "string" else result.result.strip('"')
                print(f"✅ [{result.expression_index}]: {value}")
                results.append(value)
            else:
                print(f"❌ [{result.expression_index}]: {result.error_message}")
                results.append(None)
        
        return results
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Пример использования для валидации данных
if __name__ == "__main__":
    # Правила валидации формы
    validation_rules = [
        "length(email) > 5",
        "contains(email, '@')",
        "age >= 18",
        "age <= 100",
        "length(password) >= 8",
        "phone matches '^\\+\\d{10,15}$'"
    ]
    
    user_data = {
        "email": "user@example.com",
        "age": 25,
        "password": "securepass123",
        "phone": "+1234567890"
    }
    
    results = evaluate_batch(validation_rules, user_data)
    
    # Проверяем все ли правила прошли
    if results and all(results):
        print("🎉 Все правила валидации прошли успешно!")
    else:
        print("⚠️  Некоторые правила не прошли валидацию")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function evaluateBatch(expressions, context = {}) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            expressions: expressions,
            context: JSON.stringify(context)
        };
        
        client.evaluateBatch(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log(`📊 Результаты: ${response.successful_count} успешно, ${response.failed_count} ошибок`);
            
            const results = [];
            response.results.forEach((result) => {
                if (result.success) {
                    let value;
                    try {
                        value = JSON.parse(result.result);
                    } catch {
                        value = result.result.replace(/^"|"$/g, '');
                    }
                    
                    console.log(`✅ [${result.expression_index}]: ${value}`);
                    results.push(value);
                } else {
                    console.log(`❌ [${result.expression_index}]: ${result.error_message}`);
                    results.push(null);
                }
            });
            
            resolve(results);
        });
    });
}

// Пример: расчет скидок для разных категорий товаров
async function calculateDiscounts() {
    const discountExpressions = [
        "if category = 'electronics' then price * 0.1 else 0",
        "if category = 'clothing' then price * 0.15 else 0", 
        "if category = 'books' then price * 0.05 else 0",
        "if quantity > 5 then price * 0.02 else 0", // Скидка за объем
        "if is_member then price * 0.05 else 0"    // Скидка для членов
    ];
    
    const productContext = {
        category: 'electronics',
        price: 1000,
        quantity: 3,
        is_member: true
    };
    
    const discounts = await evaluateBatch(discountExpressions, productContext);
    
    const totalDiscount = discounts
        .filter(d => d !== null && typeof d === 'number')
        .reduce((sum, discount) => sum + discount, 0);
    
    console.log(`💰 Общая скидка: $${totalDiscount.toFixed(2)}`);
    console.log(`💵 Финальная цена: $${(productContext.price - totalDiscount).toFixed(2)}`);
}

calculateDiscounts().catch(console.error);
```

## Применение

### Сценарии использования
- **Валидация данных**: Проверка множественных правил
- **Бизнес-правила**: Оценка условий и ограничений  
- **Расчет показателей**: Вычисление метрик и KPI
- **Условная логика**: Обработка сложных условий

### Преимущества пакетной обработки
- **Производительность**: Одно соединение для множественных вычислений
- **Атомарность**: Общий контекст для всех выражений
- **Отчетность**: Детальная статистика успешности

## Ограничения

### Технические ограничения
- Максимум **1000 выражений** за запрос
- Общий размер контекста до **10MB**
- Таймаут обработки **30 секунд**

### Обработка ошибок
- Ошибка в одном выражении не останавливает обработку других
- Подробная информация об ошибках для каждого выражения
- Индексация результатов соответствует порядку запроса

## Связанные методы
- [EvaluateExpression](evaluate-expression.md) - Одиночные выражения
- [EvaluateCondition](evaluate-condition.md) - Булевы результаты
- [ValidateExpression](validate-expression.md) - Предварительная валидация
