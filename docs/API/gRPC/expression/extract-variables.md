# ExtractVariables

## Описание
Анализирует FEEL выражение и извлекает все используемые переменные, включая вложенные свойства объектов. Полезно для анализа зависимостей и валидации контекста.

## Синтаксис
```protobuf
rpc ExtractVariables(ExtractVariablesRequest) returns (ExtractVariablesResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### ExtractVariablesRequest
```protobuf
message ExtractVariablesRequest {
  string expression = 1;  // FEEL выражение для анализа
  bool include_paths = 2; // Включить полные пути свойств
  string tenant_id = 3;   // ID тенанта
}
```

## Параметры ответа

### ExtractVariablesResponse
```protobuf
message ExtractVariablesResponse {
  repeated string variables = 1;           // Список переменных
  repeated VariableInfo variable_info = 2; // Детальная информация
  bool success = 3;                        // Успешность извлечения
  string error_message = 4;                // Сообщение об ошибке
}

message VariableInfo {
  string name = 1;         // Имя переменной
  string full_path = 2;    // Полный путь (например: user.profile.name)
  string type_hint = 3;    // Предполагаемый тип
  repeated int32 positions = 4; // Позиции в выражении
  bool is_nested = 5;      // Является ли вложенным свойством
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
    
    // Анализ сложного выражения
    expression := `if user.age >= 18 and user.profile.verified = true then
        calculate_discount(order.total, user.membership.level)
    else 0`
    
    response, err := client.ExtractVariables(ctx, &pb.ExtractVariablesRequest{
        Expression:   expression,
        IncludePaths: true,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        fmt.Printf("🔍 Анализ выражения:\n%s\n\n", expression)
        fmt.Printf("📊 Найдено переменных: %d\n\n", len(response.Variables))
        
        // Простой список переменных
        fmt.Println("📝 Переменные:")
        for _, variable := range response.Variables {
            fmt.Printf("  • %s\n", variable)
        }
        
        fmt.Println("\n📋 Детальная информация:")
        for _, info := range response.VariableInfo {
            fmt.Printf("  🔗 %s\n", info.Name)
            if info.FullPath != info.Name {
                fmt.Printf("     📍 Путь: %s\n", info.FullPath)
            }
            if info.TypeHint != "" {
                fmt.Printf("     📊 Тип: %s\n", info.TypeHint)
            }
            if info.IsNested {
                fmt.Printf("     🏗️ Вложенное свойство\n")
            }
            fmt.Printf("     📍 Позиции: %v\n", info.Positions)
            fmt.Println()
        }
    } else {
        fmt.Printf("❌ Ошибка: %s\n", response.ErrorMessage)
    }
}
```

### Python
```python
import grpc
import json

import expression_pb2
import expression_pb2_grpc

def extract_variables(expression, include_paths=True):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = expression_pb2.ExtractVariablesRequest(
        expression=expression,
        include_paths=include_paths
    )
    
    try:
        response = stub.ExtractVariables(request, metadata=metadata)
        
        if response.success:
            print(f"🔍 Выражение: {expression}")
            print(f"📊 Переменных найдено: {len(response.variables)}")
            print(f"📝 Переменные: {list(response.variables)}")
            
            # Группировка по типам
            root_vars = []
            nested_vars = []
            
            for info in response.variable_info:
                if info.is_nested:
                    nested_vars.append(info)
                else:
                    root_vars.append(info)
            
            if root_vars:
                print(f"\n🌳 Корневые переменные ({len(root_vars)}):")
                for var in root_vars:
                    type_info = f" ({var.type_hint})" if var.type_hint else ""
                    print(f"  • {var.name}{type_info}")
            
            if nested_vars:
                print(f"\n🏗️ Вложенные свойства ({len(nested_vars)}):")
                for var in nested_vars:
                    print(f"  • {var.full_path}")
            
            return {
                'variables': list(response.variables),
                'variable_info': response.variable_info,
                'root_vars': [v.name for v in root_vars],
                'nested_vars': [v.full_path for v in nested_vars]
            }
        else:
            print(f"❌ Ошибка: {response.error_message}")
            return None
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Анализатор зависимостей для BPMN процессов
class DependencyAnalyzer:
    def __init__(self):
        self.expressions = {}
        self.dependencies = {}
    
    def add_expression(self, name, expression):
        """Добавить выражение для анализа"""
        self.expressions[name] = expression
        
        print(f"➕ Добавлено выражение: {name}")
        result = extract_variables(expression)
        
        if result:
            self.dependencies[name] = result
        print()
    
    def analyze_dependencies(self):
        """Анализ всех зависимостей"""
        print("🔬 АНАЛИЗ ЗАВИСИМОСТЕЙ ПРОЦЕССА")
        print("=" * 50)
        
        all_variables = set()
        root_variables = set() 
        nested_properties = set()
        
        for name, deps in self.dependencies.items():
            all_variables.update(deps['variables'])
            root_variables.update(deps['root_vars'])
            nested_properties.update(deps['nested_vars'])
        
        print(f"\n📊 СВОДНАЯ СТАТИСТИКА:")
        print(f"   Всего уникальных переменных: {len(all_variables)}")
        print(f"   Корневые переменные: {len(root_variables)}")
        print(f"   Вложенные свойства: {len(nested_properties)}")
        
        print(f"\n🌳 ТРЕБУЕМЫЕ КОРНЕВЫЕ ПЕРЕМЕННЫЕ:")
        for var in sorted(root_variables):
            print(f"   • {var}")
        
        if nested_properties:
            print(f"\n🏗️ СТРУКТУРА ОБЪЕКТОВ:")
            # Группируем по корневым объектам
            object_structure = {}
            for path in nested_properties:
                root = path.split('.')[0]
                if root not in object_structure:
                    object_structure[root] = []
                object_structure[root].append(path)
            
            for obj, paths in object_structure.items():
                print(f"   📦 {obj}:")
                for path in sorted(paths):
                    property_path = '.'.join(path.split('.')[1:])
                    print(f"      └── {property_path}")
        
        return {
            'all_variables': all_variables,
            'root_variables': root_variables,
            'nested_properties': nested_properties
        }

# Пример использования анализатора
if __name__ == "__main__":
    analyzer = DependencyAnalyzer()
    
    # BPMN процесс: заявка на кредит
    analyzer.add_expression(
        "Проверка возраста",
        "applicant.age >= 18 and applicant.age <= 80"
    )
    
    analyzer.add_expression(
        "Оценка дохода", 
        "applicant.income.monthly >= 30000 and applicant.income.stable = true"
    )
    
    analyzer.add_expression(
        "Кредитная история",
        "credit.score >= 650 and credit.defaults = 0"
    )
    
    analyzer.add_expression(
        "Расчет суммы кредита",
        "min(requested.amount, applicant.income.monthly * 24)"
    )
    
    analyzer.add_expression(
        "Финальное решение",
        "if approved.age and approved.income and approved.credit then 'APPROVED' else 'REJECTED'"
    )
    
    # Анализ всех зависимостей
    summary = analyzer.analyze_dependencies()
    
    # Генерация примера контекста
    print(f"\n💡 ПРИМЕР КОНТЕКСТА ДЛЯ ТЕСТИРОВАНИЯ:")
    print("{")
    for var in sorted(summary['root_variables']):
        if var in ['applicant', 'credit', 'requested', 'approved']:
            print(f'  "{var}": {{ ... }},')
        else:
            print(f'  "{var}": "значение",')
    print("}")
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function extractVariables(expression, includePaths = true) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            expression: expression,
            include_paths: includePaths
        };
        
        client.extractVariables(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                console.log(`🔍 Анализ: ${expression}`);
                console.log(`📊 Найдено переменных: ${response.variables.length}`);
                console.log(`📝 Переменные: ${response.variables.join(', ')}`);
                
                // Классификация переменных
                const rootVars = response.variable_info.filter(v => !v.is_nested);
                const nestedVars = response.variable_info.filter(v => v.is_nested);
                
                if (rootVars.length > 0) {
                    console.log(`\n🌳 Корневые (${rootVars.length}):`);
                    rootVars.forEach(v => {
                        console.log(`  • ${v.name}${v.type_hint ? ` (${v.type_hint})` : ''}`);
                    });
                }
                
                if (nestedVars.length > 0) {
                    console.log(`\n🏗️ Вложенные (${nestedVars.length}):`);
                    nestedVars.forEach(v => {
                        console.log(`  • ${v.full_path}`);
                    });
                }
                
                console.log(); // Пустая строка
                
                resolve({
                    variables: response.variables,
                    variableInfo: response.variable_info,
                    rootVars: rootVars.map(v => v.name),
                    nestedVars: nestedVars.map(v => v.full_path)
                });
            } else {
                console.log(`❌ Ошибка: ${response.error_message}`);
                resolve(null);
            }
        });
    });
}

// Генератор схем данных на основе переменных
class SchemaGenerator {
    constructor() {
        this.schemas = new Map();
    }
    
    async analyzeExpression(name, expression) {
        console.log(`📋 Анализ выражения: ${name}`);
        console.log(`   ${expression}`);
        
        const result = await extractVariables(expression);
        
        if (result) {
            this.schemas.set(name, result);
            return result;
        }
        
        return null;
    }
    
    generateJSONSchema() {
        console.log('\n📐 ГЕНЕРАЦИЯ JSON SCHEMA');
        console.log('='.repeat(50));
        
        // Собираем все уникальные переменные
        const allVars = new Set();
        const nestedPaths = new Set();
        
        this.schemas.forEach((schema, name) => {
            schema.rootVars.forEach(v => allVars.add(v));
            schema.nestedVars.forEach(v => nestedPaths.add(v));
        });
        
        // Строим структуру схемы
        const schemaStructure = {
            type: "object",
            properties: {}
        };
        
        // Добавляем корневые переменные
        allVars.forEach(varName => {
            // Проверяем, есть ли вложенные свойства для этой переменной
            const hasNested = Array.from(nestedPaths).some(path => path.startsWith(varName + '.'));
            
            if (hasNested) {
                schemaStructure.properties[varName] = {
                    type: "object",
                    properties: {}
                };
                
                // Добавляем вложенные свойства
                Array.from(nestedPaths)
                    .filter(path => path.startsWith(varName + '.'))
                    .forEach(path => {
                        const parts = path.split('.');
                        let current = schemaStructure.properties[varName];
                        
                        for (let i = 1; i < parts.length; i++) {
                            const part = parts[i];
                            
                            if (i === parts.length - 1) {
                                // Последний элемент - добавляем свойство
                                current.properties[part] = {
                                    type: "string", // По умолчанию строка
                                    description: `Значение для ${path}`
                                };
                            } else {
                                // Промежуточный объект
                                if (!current.properties[part]) {
                                    current.properties[part] = {
                                        type: "object",
                                        properties: {}
                                    };
                                }
                                current = current.properties[part];
                            }
                        }
                    });
            } else {
                // Простая переменная
                schemaStructure.properties[varName] = {
                    type: "string",
                    description: `Значение для переменной ${varName}`
                };
            }
        });
        
        console.log('\n📄 JSON Schema:');
        console.log(JSON.stringify(schemaStructure, null, 2));
        
        return schemaStructure;
    }
    
    generateSampleData() {
        console.log('\n💡 ПРИМЕР ДАННЫХ:');
        console.log('='.repeat(30));
        
        const sampleData = {};
        const allVars = new Set();
        const nestedPaths = new Set();
        
        this.schemas.forEach((schema, name) => {
            schema.rootVars.forEach(v => allVars.add(v));
            schema.nestedVars.forEach(v => nestedPaths.add(v));
        });
        
        // Генерируем примеры данных
        allVars.forEach(varName => {
            const hasNested = Array.from(nestedPaths).some(path => path.startsWith(varName + '.'));
            
            if (hasNested) {
                sampleData[varName] = {};
                
                // Добавляем вложенные свойства с примерами
                Array.from(nestedPaths)
                    .filter(path => path.startsWith(varName + '.'))
                    .forEach(path => {
                        const parts = path.split('.');
                        let current = sampleData[varName];
                        
                        for (let i = 1; i < parts.length; i++) {
                            const part = parts[i];
                            
                            if (i === parts.length - 1) {
                                // Генерируем примеры на основе имени
                                current[part] = generateSampleValue(part);
                            } else {
                                if (!current[part]) {
                                    current[part] = {};
                                }
                                current = current[part];
                            }
                        }
                    });
            } else {
                sampleData[varName] = generateSampleValue(varName);
            }
        });
        
        console.log(JSON.stringify(sampleData, null, 2));
        
        return sampleData;
    }
}

// Генератор примеров значений на основе имени переменной
function generateSampleValue(name) {
    const lowerName = name.toLowerCase();
    
    if (lowerName.includes('age')) return 25;
    if (lowerName.includes('price') || lowerName.includes('cost') || lowerName.includes('amount')) return 1000;
    if (lowerName.includes('email')) return 'user@example.com';
    if (lowerName.includes('name')) return 'John Doe';
    if (lowerName.includes('status')) return 'active';
    if (lowerName.includes('verified') || lowerName.includes('enabled')) return true;
    if (lowerName.includes('count') || lowerName.includes('number')) return 5;
    if (lowerName.includes('date')) return '2023-12-01';
    
    return `sample_${name}`;
}

// Демонстрация анализа выражений
async function demonstrateVariableExtraction() {
    console.log('🧪 ДЕМОНСТРАЦИЯ ИЗВЛЕЧЕНИЯ ПЕРЕМЕННЫХ');
    console.log('='.repeat(60));
    
    const generator = new SchemaGenerator();
    
    // Анализируем различные типы выражений
    const expressions = [
        ['Простое условие', 'age >= 18'],
        ['Вложенные объекты', 'user.profile.name = "John" and user.settings.notifications = true'],
        ['Списки и функции', 'count(order.items) > 0 and sum(order.items[item.price]) > 100'],
        ['Сложная логика', 'if customer.vip and order.total > 1000 then discount.premium else discount.standard'],
        ['Математика', 'sqrt(position.x * position.x + position.y * position.y) < radius']
    ];
    
    for (const [name, expression] of expressions) {
        await generator.analyzeExpression(name, expression);
        console.log('-'.repeat(40));
    }
    
    // Генерируем схему и примеры
    const schema = generator.generateJSONSchema();
    const sampleData = generator.generateSampleData();
    
    return { schema, sampleData };
}

demonstrateVariableExtraction().catch(console.error);
```

## Типы анализа

### Поверхностный анализ
```protobuf
ExtractVariablesRequest {
  expression: "user.name + order.total"
  include_paths: false
}
// Результат: ["user", "order"]
```

### Глубокий анализ
```protobuf
ExtractVariablesRequest {
  expression: "user.name + order.total"
  include_paths: true
}
// Результат: ["user", "user.name", "order", "order.total"]
```

## Применение

### Валидация контекста
- Проверка наличия всех требуемых переменных
- Генерация схем данных для валидации
- Построение минимального контекста

### BPMN анализ
- Анализ зависимостей между задачами
- Оптимизация передачи данных
- Документирование процессов

### IDE поддержка
- Автодополнение переменных
- Предупреждения о неиспользуемых переменных
- Рефакторинг выражений

## Структура VariableInfo

### Поля информации
- **name**: Базовое имя переменной  
- **full_path**: Полный путь (user.profile.name)
- **type_hint**: Предполагаемый тип данных
- **positions**: Позиции в исходном выражении
- **is_nested**: Флаг вложенности

## Связанные методы
- [ValidateExpression](validate-expression.md) - Валидация с проверкой переменных
- [ParseExpression](parse-expression.md) - Структурный анализ выражений
- [EvaluateExpression](evaluate-expression.md) - Использование извлеченных переменных
