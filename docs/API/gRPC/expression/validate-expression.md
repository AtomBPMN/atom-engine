# ValidateExpression

## Описание
Проверяет синтаксическую и семантическую корректность FEEL выражения без его выполнения. Возвращает подробную информацию об ошибках и предупреждениях.

## Синтаксис
```protobuf
rpc ValidateExpression(ValidateExpressionRequest) returns (ValidateExpressionResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### ValidateExpressionRequest
```protobuf
message ValidateExpressionRequest {
  string expression = 1;          // FEEL выражение для валидации
  string context_schema = 2;      // JSON схема для контекста (опционально)
  string tenant_id = 3;           // ID тенанта
}
```

## Параметры ответа

### ValidateExpressionResponse
```protobuf
message ValidateExpressionResponse {
  bool is_valid = 1;                     // Валидность выражения
  repeated ValidationError errors = 2;    // Список ошибок
  repeated ValidationWarning warnings = 3; // Список предупреждений
  repeated string used_variables = 4;     // Используемые переменные
  string result_type = 5;                // Ожидаемый тип результата
}

message ValidationError {
  string message = 1;    // Сообщение об ошибке
  int32 line = 2;        // Номер строки
  int32 column = 3;      // Номер колонки
  string error_code = 4; // Код ошибки
}

message ValidationWarning {
  string message = 1;      // Сообщение предупреждения
  int32 line = 2;          // Номер строки
  int32 column = 3;        // Номер колонки  
  string warning_code = 4; // Код предупреждения
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
    
    // Валидация выражений
    expressions := []string{
        "x + y",                          // Корректное
        "if age >= then 'adult'",        // Ошибка синтаксиса
        "unknownFunc(123)",              // Неизвестная функция
        "user.age > 18 and status = 1",  // Корректное
    }
    
    for _, expr := range expressions {
        fmt.Printf("\n🔍 Валидация: %s\n", expr)
        
        response, err := client.ValidateExpression(ctx, &pb.ValidateExpressionRequest{
            Expression: expr,
        })
        
        if err != nil {
            fmt.Printf("❌ gRPC Error: %v\n", err)
            continue
        }
        
        if response.IsValid {
            fmt.Printf("✅ Выражение корректно\n")
            fmt.Printf("📊 Тип результата: %s\n", response.ResultType)
            
            if len(response.UsedVariables) > 0 {
                fmt.Printf("🔗 Используемые переменные: %v\n", response.UsedVariables)
            }
        } else {
            fmt.Printf("❌ Выражение содержит ошибки:\n")
            for _, errMsg := range response.Errors {
                fmt.Printf("  💥 [%d:%d] %s (код: %s)\n", 
                    errMsg.Line, errMsg.Column, errMsg.Message, errMsg.ErrorCode)
            }
        }
        
        if len(response.Warnings) > 0 {
            fmt.Printf("⚠️ Предупреждения:\n")
            for _, warning := range response.Warnings {
                fmt.Printf("  ⚡ [%d:%d] %s (код: %s)\n",
                    warning.Line, warning.Column, warning.Message, warning.WarningCode)
            }
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

def validate_expression(expression, context_schema=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = expression_pb2.ValidateExpressionRequest(
        expression=expression,
        context_schema=context_schema or ""
    )
    
    try:
        response = stub.ValidateExpression(request, metadata=metadata)
        
        print(f"🔍 Валидация: {expression}")
        
        if response.is_valid:
            print("✅ Выражение корректно")
            print(f"📊 Тип результата: {response.result_type}")
            
            if response.used_variables:
                print(f"🔗 Переменные: {list(response.used_variables)}")
        else:
            print("❌ Ошибки в выражении:")
            for error in response.errors:
                print(f"  💥 [{error.line}:{error.column}] {error.message}")
        
        if response.warnings:
            print("⚠️ Предупреждения:")
            for warning in response.warnings:
                print(f"  ⚡ [{warning.line}:{warning.column}] {warning.message}")
        
        return response.is_valid
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

# Валидация выражений с контекстной схемой  
if __name__ == "__main__":
    # Схема контекста для валидации
    user_schema = json.dumps({
        "type": "object",
        "properties": {
            "age": {"type": "number", "minimum": 0},
            "name": {"type": "string"},
            "status": {"type": "string", "enum": ["active", "inactive"]},
            "balance": {"type": "number"}
        },
        "required": ["age", "name"]
    })
    
    test_expressions = [
        ("age >= 18", "✅ Простое условие"),
        ("name + ' is ' + status", "✅ Конкатенация строк"),
        ("balance > 1000 and status = 'active'", "✅ Сложное условие"),
        ("age + unknownField", "❌ Неизвестное поле"),
        ("if age >= then 'adult'", "❌ Синтаксическая ошибка")
    ]
    
    for expression, description in test_expressions:
        print(f"\n{'-'*50}")
        print(f"📝 {description}")
        validate_expression(expression, user_schema)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function validateExpression(expression, contextSchema = null) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            expression: expression,
            context_schema: contextSchema || ''
        };
        
        client.validateExpression(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log(`🔍 Валидация: ${expression}`);
            
            if (response.is_valid) {
                console.log('✅ Выражение корректно');
                console.log(`📊 Тип результата: ${response.result_type}`);
                
                if (response.used_variables.length > 0) {
                    console.log(`🔗 Переменные: ${response.used_variables.join(', ')}`);
                }
            } else {
                console.log('❌ Ошибки:');
                response.errors.forEach(error => {
                    console.log(`  💥 [${error.line}:${error.column}] ${error.message} (${error.error_code})`);
                });
            }
            
            if (response.warnings.length > 0) {
                console.log('⚠️ Предупреждения:');
                response.warnings.forEach(warning => {
                    console.log(`  ⚡ [${warning.line}:${warning.column}] ${warning.message} (${warning.warning_code})`);
                });
            }
            
            resolve({
                isValid: response.is_valid,
                errors: response.errors,
                warnings: response.warnings,
                variables: response.used_variables,
                resultType: response.result_type
            });
        });
    });
}

// Пакетная валидация выражений
async function validateBusinessRules() {
    console.log('🏢 Валидация бизнес-правил:\n');
    
    const businessRules = [
        {
            name: 'Проверка возраста',
            expression: 'age >= 18 and age <= 120'
        },
        {
            name: 'Расчет скидки',
            expression: 'if customer.level = "premium" then order.total * 0.1 else 0'
        },
        {
            name: 'Валидация email',
            expression: 'contains(email, "@") and length(email) > 5'
        },
        {
            name: 'Некорректное правило',
            expression: 'if price >= then "expensive"' // Ошибка
        }
    ];
    
    for (const rule of businessRules) {
        console.log(`\n📋 Правило: ${rule.name}`);
        try {
            const result = await validateExpression(rule.expression);
            
            if (!result.isValid) {
                console.log('🚨 Правило требует исправления!');
            }
        } catch (error) {
            console.log(`❌ Ошибка валидации: ${error.message}`);
        }
    }
}

validateBusinessRules().catch(console.error);
```

## Типы ошибок

### Синтаксические ошибки
- **SYNTAX_ERROR**: Неверный синтаксис FEEL
- **UNEXPECTED_TOKEN**: Неожиданный токен
- **MISSING_OPERAND**: Отсутствует операнд
- **UNCLOSED_PARENTHESES**: Незакрытые скобки

### Семантические ошибки  
- **UNKNOWN_FUNCTION**: Неизвестная функция
- **WRONG_ARGUMENT_COUNT**: Неверное количество аргументов
- **TYPE_MISMATCH**: Несовместимость типов
- **UNDEFINED_VARIABLE**: Неопределенная переменная

## Типы предупреждений

### Качество кода
- **UNUSED_VARIABLE**: Неиспользуемая переменная
- **REDUNDANT_CONDITION**: Избыточное условие
- **POTENTIAL_NULL**: Потенциальное значение null
- **PERFORMANCE_WARNING**: Неоптимальная конструкция

## Схема контекста

### JSON Schema поддержка
```json
{
  "type": "object",
  "properties": {
    "user": {
      "type": "object", 
      "properties": {
        "age": {"type": "number"},
        "name": {"type": "string"}
      }
    }
  }
}
```

## Применение

### CI/CD интеграция
- Валидация выражений в конвейере развертывания
- Проверка качества бизнес-правил
- Предотвращение деплоя с ошибками

### IDE поддержка
- Подсветка ошибок в реальном времени
- Автодополнение на основе схемы
- Рефакторинг выражений

## Связанные методы
- [ParseExpression](parse-expression.md) - Структурный анализ
- [ExtractVariables](extract-variables.md) - Анализ зависимостей
- [EvaluateExpression](evaluate-expression.md) - Выполнение валидных выражений
