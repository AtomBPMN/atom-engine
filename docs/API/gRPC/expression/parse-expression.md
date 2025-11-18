# ParseExpression

## Описание
Парсит FEEL выражение в абстрактное синтаксическое дерево (AST), возвращая структурированное представление выражения без его выполнения.

## Синтаксис
```protobuf
rpc ParseExpression(ParseExpressionRequest) returns (ParseExpressionResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### ParseExpressionRequest
```protobuf
message ParseExpressionRequest {
  string expression = 1;  // FEEL выражение для парсинга
  string tenant_id = 2;   // ID тенанта
}
```

## Параметры ответа

### ParseExpressionResponse
```protobuf
message ParseExpressionResponse {
  string ast_json = 1;      // JSON представление AST
  bool success = 2;         // Успешность парсинга
  string error_message = 3; // Сообщение об ошибке
  repeated string warnings = 4; // Предупреждения
}
```

## Примеры использования

### Go
```go
package main

import (
    "context"
    "encoding/json"
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
    
    // Парсинг простого выражения
    response, err := client.ParseExpression(ctx, &pb.ParseExpressionRequest{
        Expression: "x + y * 2",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        var ast interface{}
        json.Unmarshal([]byte(response.AstJson), &ast)
        
        fmt.Println("✅ AST успешно создано:")
        prettyJSON, _ := json.MarshalIndent(ast, "", "  ")
        fmt.Println(string(prettyJSON))
        
        if len(response.Warnings) > 0 {
            fmt.Println("\n⚠️ Предупреждения:")
            for _, warning := range response.Warnings {
                fmt.Printf("  - %s\n", warning)
            }
        }
    } else {
        fmt.Printf("❌ Ошибка парсинга: %s\n", response.ErrorMessage)
    }
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def parse_expression(expression):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = expression_pb2.ParseExpressionRequest(
        expression=expression
    )
    
    try:
        response = stub.ParseExpression(request, metadata=metadata)
        
        if response.success:
            ast = json.loads(response.ast_json)
            print(f"✅ AST для выражения: {expression}")
            print(json.dumps(ast, indent=2, ensure_ascii=False))
            
            if response.warnings:
                print("\n⚠️ Предупреждения:")
                for warning in response.warnings:
                    print(f"  - {warning}")
            
            return ast
        else:
            print(f"❌ Ошибка парсинга: {response.error_message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Примеры парсинга различных выражений
if __name__ == "__main__":
    expressions = [
        "x + y",
        "if age >= 18 then 'adult' else 'minor'",
        "sum([1, 2, 3, 4, 5])",
        "user.name",
        "count(items[price > 100])"
    ]
    
    for expr in expressions:
        print(f"\n{'='*50}")
        parse_expression(expr)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function parseExpression(expression) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { expression: expression };
        
        client.parseExpression(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                const ast = JSON.parse(response.ast_json);
                
                console.log(`✅ AST для: ${expression}`);
                console.log(JSON.stringify(ast, null, 2));
                
                if (response.warnings.length > 0) {
                    console.log('\n⚠️ Предупреждения:');
                    response.warnings.forEach(warning => {
                        console.log(`  - ${warning}`);
                    });
                }
                
                resolve(ast);
            } else {
                console.log(`❌ Ошибка: ${response.error_message}`);
                resolve(null);
            }
        });
    });
}

// Анализ структуры выражения
async function analyzeExpression() {
    const complexExpression = 'if user.age >= 18 and user.status = "active" then calculate_discount(user.category, order.total) else 0';
    
    console.log('🔍 Анализ сложного выражения:');
    console.log(`Выражение: ${complexExpression}\n`);
    
    const ast = await parseExpression(complexExpression);
    
    if (ast) {
        console.log('\n📋 Анализ структуры:');
        analyzeASTStructure(ast);
    }
}

function analyzeASTStructure(node, depth = 0) {
    const indent = '  '.repeat(depth);
    
    if (typeof node === 'object' && node !== null) {
        if (node.type) {
            console.log(`${indent}📦 Тип узла: ${node.type}`);
        }
        
        if (node.operator) {
            console.log(`${indent}⚙️ Оператор: ${node.operator}`);
        }
        
        if (node.value !== undefined) {
            console.log(`${indent}💎 Значение: ${node.value}`);
        }
        
        // Рекурсивно анализируем дочерние узлы
        Object.keys(node).forEach(key => {
            if (key !== 'type' && key !== 'operator' && key !== 'value') {
                if (Array.isArray(node[key])) {
                    console.log(`${indent}📋 ${key}:`);
                    node[key].forEach((item, index) => {
                        console.log(`${indent}  [${index}]:`);
                        analyzeASTStructure(item, depth + 2);
                    });
                } else if (typeof node[key] === 'object') {
                    console.log(`${indent}🔗 ${key}:`);
                    analyzeASTStructure(node[key], depth + 1);
                }
            }
        });
    }
}

analyzeExpression().catch(console.error);
```

## Структура AST

### Типы узлов
```json
{
  "type": "BinaryOperation",
  "operator": "+",
  "left": {
    "type": "Variable", 
    "name": "x"
  },
  "right": {
    "type": "BinaryOperation",
    "operator": "*",
    "left": {
      "type": "Variable",
      "name": "y" 
    },
    "right": {
      "type": "Literal",
      "value": 2,
      "dataType": "number"
    }
  }
}
```

### Основные типы узлов
- **Literal**: Литеральные значения (числа, строки, булевы)
- **Variable**: Переменные
- **BinaryOperation**: Бинарные операции (+, -, *, /, =, !=, etc.)
- **UnaryOperation**: Унарные операции (not, -)
- **FunctionCall**: Вызовы функций
- **ConditionalExpression**: if-then-else выражения
- **ListExpression**: Списки и фильтрация
- **PropertyAccess**: Доступ к свойствам объекта

## Применение

### Статический анализ
- **Валидация синтаксиса** без выполнения
- **Поиск используемых переменных**
- **Оптимизация выражений**
- **Генерация документации**

### IDE поддержка
- **Подсветка синтаксиса**
- **Автодополнение**
- **Поиск ошибок**
- **Рефакторинг**

## Предупреждения

### Типы предупреждений
- Неиспользуемые переменные
- Потенциальные деления на ноль
- Неоптимальные конструкции
- Устаревший синтаксис

## Связанные методы
- [ValidateExpression](validate-expression.md) - Валидация выражений
- [ExtractVariables](extract-variables.md) - Извлечение переменных
- [EvaluateExpression](evaluate-expression.md) - Выполнение выражений
