# EvaluateExpression

## Описание
Вычисляет FEEL (Friendly Enough Expression Language) выражение с заданным контекстом переменных. Возвращает результат в формате JSON.

## Синтаксис
```protobuf
rpc EvaluateExpression(EvaluateExpressionRequest) returns (EvaluateExpressionResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### EvaluateExpressionRequest
```protobuf
message EvaluateExpressionRequest {
  string expression = 1;  // FEEL выражение
  string context = 2;     // JSON с переменными
  string tenant_id = 3;   // ID тенанта
}
```

## Параметры ответа

### EvaluateExpressionResponse
```protobuf
message EvaluateExpressionResponse {
  string result = 1;        // JSON результат
  bool success = 2;         // Успешность вычисления
  string error_message = 3; // Сообщение об ошибке
  string result_type = 4;   // Тип результата
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
    
    // Простое арифметическое выражение
    response, err := client.EvaluateExpression(ctx, &pb.EvaluateExpressionRequest{
        Expression: "x + y",
        Context:    `{"x": 10, "y": 5}`,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("✅ Результат: %s (тип: %s)\n", response.Result, response.ResultType)
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.ErrorMessage)
    }
    
    // Условное выражение
    response, err = client.EvaluateExpression(ctx, &pb.EvaluateExpressionRequest{
        Expression: `if age >= 18 then "adult" else "minor"`,
        Context:    `{"age": 25}`,
    })
    
    if err == nil && response.Success {
        fmt.Printf("✅ Категория: %s\n", response.Result)
    }
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def evaluate_expression(expression, context=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    context_json = json.dumps(context or {})
    
    request = expression_pb2.EvaluateExpressionRequest(
        expression=expression,
        context=context_json
    )
    
    try:
        response = stub.EvaluateExpression(request, metadata=metadata)
        
        if response.success:
            result = json.loads(response.result) if response.result_type != "string" else response.result.strip('"')
            print(f"✅ Результат: {result} (тип: {response.result_type})")
            return result
        else:
            print(f"❌ Ошибка: {response.error_message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Примеры использования
if __name__ == "__main__":
    # Арифметика
    evaluate_expression("2 + 3 * 4")
    
    # Работа с переменными
    evaluate_expression("price * quantity", {"price": 10.5, "quantity": 3})
    
    # Условия
    evaluate_expression('if score > 90 then "A" else "B"', {"score": 95})
    
    # Строки
    evaluate_expression('upper("hello world")')
    
    # Списки
    evaluate_expression("sum([1, 2, 3, 4, 5])")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function evaluateExpression(expression, context = {}) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            expression: expression,
            context: JSON.stringify(context)
        };
        
        client.evaluateExpression(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                let result;
                try {
                    result = JSON.parse(response.result);
                } catch {
                    result = response.result.replace(/^"|"$/g, ''); // Убираем кавычки для строк
                }
                
                console.log(`✅ Результат: ${result} (тип: ${response.result_type})`);
                resolve(result);
            } else {
                console.log(`❌ Ошибка: ${response.error_message}`);
                resolve(null);
            }
        });
    });
}

// Примеры использования
async function examples() {
    // Арифметика
    await evaluateExpression("10 / 2 + 3");
    
    // Условия
    await evaluateExpression("if temperature > 30 then 'hot' else 'normal'", 
        { temperature: 35 });
    
    // Функции над списками
    await evaluateExpression("count(items)", 
        { items: ["apple", "banana", "orange"] });
    
    // Работа с датами
    await evaluateExpression("now()");
    
    // Сложные объекты
    await evaluateExpression("user.name", 
        { user: { name: "John", age: 30 } });
}

examples().catch(console.error);
```

## FEEL Выражения

### Арифметические операторы
- `+`, `-`, `*`, `/` - базовые операции
- `**` - возведение в степень
- `%` - остаток от деления

### Операторы сравнения
- `=`, `!=` - равенство/неравенство
- `<`, `<=`, `>`, `>=` - сравнение
- `in` - принадлежность списку
- `instance of` - проверка типа

### Логические операторы
- `and`, `or`, `not`
- `if-then-else`

### Встроенные функции
- **Строковые**: `upper()`, `lower()`, `substring()`, `length()`
- **Числовые**: `abs()`, `round()`, `floor()`, `ceil()`
- **Списки**: `count()`, `sum()`, `max()`, `min()`
- **Даты**: `now()`, `today()`, `date()`

## Типы результатов

### Поддерживаемые типы
- **string**: Строка
- **number**: Число
- **boolean**: Булево значение
- **object**: JSON объект
- **array**: Массив
- **null**: Пустое значение

## Возможные ошибки

### Синтаксические ошибки
- Неверный синтаксис FEEL
- Неопределенные переменные
- Неверные типы операндов

### Примеры ошибок
```json
{
  "result": "",
  "success": false,
  "error_message": "Variable 'x' is not defined",
  "result_type": ""
}
```

## Связанные методы
- [EvaluateCondition](evaluate-condition.md) - Для булевых результатов
- [ValidateExpression](validate-expression.md) - Валидация синтаксиса
- [ExtractVariables](extract-variables.md) - Извлечение переменных
