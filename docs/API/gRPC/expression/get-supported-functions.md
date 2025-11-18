# GetSupportedFunctions

## Описание
Возвращает список всех поддерживаемых встроенных функций FEEL с их описаниями, синтаксисом и примерами использования. Поддерживает фильтрацию по категориям.

## Синтаксис
```protobuf
rpc GetSupportedFunctions(GetSupportedFunctionsRequest) returns (GetSupportedFunctionsResponse);
```

## Package
```protobuf
package expression;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `expression` или `*`

## Параметры запроса

### GetSupportedFunctionsRequest
```protobuf
message GetSupportedFunctionsRequest {
  string category = 1;    // Категория функций (опционально)
  string tenant_id = 2;   // ID тенанта
}
```

## Параметры ответа

### GetSupportedFunctionsResponse
```protobuf
message GetSupportedFunctionsResponse {
  repeated FunctionInfo functions = 1;  // Список функций
  repeated string categories = 2;       // Доступные категории
  int32 total_count = 3;               // Общее количество функций
}

message FunctionInfo {
  string name = 1;          // Название функции
  string description = 2;   // Описание функции
  string syntax = 3;        // Синтаксис вызова
  string category = 4;      // Категория функции
  repeated string examples = 5; // Примеры использования
  repeated string parameters = 6; // Описание параметров
  string return_type = 7;   // Тип возвращаемого значения
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
    
    // Получить все функции
    response, err := client.GetSupportedFunctions(ctx, &pb.GetSupportedFunctionsRequest{})
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("📚 Всего функций: %d\n", response.TotalCount)
    fmt.Printf("📂 Категории: %v\n\n", response.Categories)
    
    // Группировка по категориям
    categoryMap := make(map[string][]*pb.FunctionInfo)
    for _, fn := range response.Functions {
        categoryMap[fn.Category] = append(categoryMap[fn.Category], fn)
    }
    
    for category, functions := range categoryMap {
        fmt.Printf("📁 %s (%d функций):\n", category, len(functions))
        for _, fn := range functions {
            fmt.Printf("  🔧 %s - %s\n", fn.Name, fn.Description)
            fmt.Printf("     📝 %s\n", fn.Syntax)
            if len(fn.Examples) > 0 {
                fmt.Printf("     💡 %s\n", fn.Examples[0])
            }
        }
        fmt.Println()
    }
    
    // Получить функции определенной категории
    fmt.Println("🔍 Функции работы со строками:")
    stringResponse, err := client.GetSupportedFunctions(ctx, &pb.GetSupportedFunctionsRequest{
        Category: "string",
    })
    
    if err == nil {
        for _, fn := range stringResponse.Functions {
            fmt.Printf("  📄 %s: %s\n", fn.Name, fn.Description)
        }
    }
}
```

### Python
```python
import grpc

import expression_pb2
import expression_pb2_grpc

def get_functions(category=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = expression_pb2_grpc.ExpressionServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = expression_pb2.GetSupportedFunctionsRequest(
        category=category or ""
    )
    
    try:
        response = stub.GetSupportedFunctions(request, metadata=metadata)
        
        if category:
            print(f"🔍 Функции категории '{category}': {response.total_count}")
        else:
            print(f"📚 Всего функций: {response.total_count}")
            print(f"📂 Категории: {list(response.categories)}")
        
        # Группировка и отображение
        for func in response.functions:
            print(f"\n🔧 {func.name}")
            print(f"   📖 {func.description}")
            print(f"   📝 Синтаксис: {func.syntax}")
            print(f"   📊 Возвращает: {func.return_type}")
            
            if func.parameters:
                print(f"   📋 Параметры:")
                for param in func.parameters:
                    print(f"     • {param}")
            
            if func.examples:
                print(f"   💡 Примеры:")
                for example in func.examples[:2]:  # Первые 2 примера
                    print(f"     ▶ {example}")
        
        return response.functions
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return []

# Интерактивный справочник функций
def interactive_help():
    print("🏠 FEEL Function Reference")
    print("=" * 50)
    
    # Получаем список категорий
    all_functions = get_functions()
    categories = set(func.category for func in all_functions)
    
    while True:
        print(f"\n📂 Доступные категории:")
        for i, category in enumerate(sorted(categories), 1):
            count = len([f for f in all_functions if f.category == category])
            print(f"  {i}. {category} ({count} функций)")
        
        print(f"  {len(categories) + 1}. Все функции")
        print(f"  0. Выход")
        
        try:
            choice = int(input("\nВыберите категорию: "))
            if choice == 0:
                break
            elif choice <= len(categories):
                selected_category = sorted(categories)[choice - 1]
                get_functions(selected_category)
            elif choice == len(categories) + 1:
                get_functions()
        except (ValueError, IndexError):
            print("❌ Неверный выбор")

if __name__ == "__main__":
    # Показать все функции
    print("📚 Демонстрация всех функций FEEL:\n")
    get_functions()
    
    # Интерактивный режим (закомментировано для демо)
    # interactive_help()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'expression.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const expressionProto = grpc.loadPackageDefinition(packageDefinition).expression;

async function getSupportedFunctions(category = null) {
    const client = new expressionProto.ExpressionService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = { category: category || '' };
        
        client.getSupportedFunctions(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log(category 
                ? `🔍 Функции категории '${category}': ${response.total_count}`
                : `📚 Всего функций: ${response.total_count}`);
                
            if (!category) {
                console.log(`📂 Категории: ${response.categories.join(', ')}`);
            }
            
            // Группировка по категориям
            const groupedFunctions = {};
            response.functions.forEach(func => {
                if (!groupedFunctions[func.category]) {
                    groupedFunctions[func.category] = [];
                }
                groupedFunctions[func.category].push(func);
            });
            
            // Отображение функций
            Object.keys(groupedFunctions).sort().forEach(cat => {
                if (category && cat !== category) return;
                
                console.log(`\n📁 ${cat.toUpperCase()} (${groupedFunctions[cat].length} функций):`);
                
                groupedFunctions[cat].forEach(func => {
                    console.log(`\n  🔧 ${func.name}`);
                    console.log(`     📖 ${func.description}`);
                    console.log(`     📝 ${func.syntax}`);
                    console.log(`     📊 → ${func.return_type}`);
                    
                    if (func.examples.length > 0) {
                        console.log(`     💡 ${func.examples[0]}`);
                    }
                });
            });
            
            resolve(response.functions);
        });
    });
}

// Создание интерактивной документации
async function generateFunctionDocs() {
    console.log('📋 Генерация справочника FEEL функций\n');
    
    try {
        const functions = await getSupportedFunctions();
        
        // Создаем чит-лист для быстрого доступа
        console.log('\n' + '='.repeat(60));
        console.log('📖 FEEL FUNCTIONS CHEAT SHEET');
        console.log('='.repeat(60));
        
        const categories = ['string', 'number', 'list', 'date', 'context'];
        
        for (const category of categories) {
            const categoryFunctions = functions.filter(f => f.category === category);
            
            if (categoryFunctions.length > 0) {
                console.log(`\n🏷️  ${category.toUpperCase()}:`);
                categoryFunctions.forEach(func => {
                    const shortExample = func.examples[0] || func.syntax;
                    console.log(`   ${func.name.padEnd(15)} | ${shortExample}`);
                });
            }
        }
        
        // Топ наиболее используемых функций
        const commonFunctions = ['upper', 'lower', 'substring', 'length', 'sum', 'count', 'max', 'min', 'now', 'if'];
        const topFunctions = functions.filter(f => commonFunctions.includes(f.name));
        
        console.log('\n' + '='.repeat(60));
        console.log('⭐ ТОП-10 НАИБОЛЕЕ ИСПОЛЬЗУЕМЫХ ФУНКЦИЙ');
        console.log('='.repeat(60));
        
        topFunctions.forEach((func, index) => {
            console.log(`\n${index + 1}. ${func.name}`);
            console.log(`   📖 ${func.description}`);
            console.log(`   💡 ${func.examples[0] || func.syntax}`);
        });
        
    } catch (error) {
        console.error(`❌ Ошибка: ${error.message}`);
    }
}

// Демонстрация работы с конкретными категориями
async function demonstrateCategories() {
    const categories = ['string', 'number', 'list', 'date'];
    
    console.log('🎯 Демонстрация функций по категориям:\n');
    
    for (const category of categories) {
        console.log(`\n${'='.repeat(30)}`);
        console.log(`📂 КАТЕГОРИЯ: ${category.toUpperCase()}`);
        console.log('='.repeat(30));
        
        try {
            await getSupportedFunctions(category);
        } catch (error) {
            console.log(`❌ Ошибка получения функций категории ${category}: ${error.message}`);
        }
        
        // Небольшая пауза для читаемости
        await new Promise(resolve => setTimeout(resolve, 1000));
    }
}

// Запуск демонстрации
async function main() {
    await generateFunctionDocs();
    // await demonstrateCategories(); // Раскомментировать для полной демо
}

main().catch(console.error);
```

## Категории функций

### String Functions
- **upper()**, **lower()** - Преобразование регистра
- **substring()** - Извлечение подстроки
- **length()** - Длина строки
- **contains()** - Поиск подстроки
- **matches()** - Регулярные выражения

### Number Functions
- **abs()** - Абсолютное значение
- **round()**, **floor()**, **ceil()** - Округление
- **min()**, **max()** - Минимум/максимум
- **sum()** - Сумма значений

### List Functions
- **count()** - Количество элементов
- **append()** - Добавление элемента
- **reverse()** - Обращение списка
- **sort()** - Сортировка
- **filter()** - Фильтрация

### Date Functions
- **now()** - Текущие дата/время
- **today()** - Текущая дата
- **date()** - Создание даты
- **date and time()** - Создание даты и времени

### Context Functions
- **get entries()** - Получение записей контекста
- **get value()** - Получение значения по ключу

## Применение

### Автодополнение IDE
```javascript
// Получение функций для автодополнения
const functions = await getSupportedFunctions();
const functionNames = functions.map(f => f.name);
```

### Валидация выражений
```javascript
// Проверка существования функции в выражении
const unknownFunctions = extractFunctions(expression)
  .filter(name => !functionNames.includes(name));
```

### Генерация документации
```javascript
// Создание справочника для пользователей
functions.forEach(func => {
  generateDocPage(func.name, func.description, func.examples);
});
```

## Фильтрация

### По категории
```protobuf
// Только строковые функции
GetSupportedFunctionsRequest {
  category: "string"
}
```

### Поддерживаемые категории
- **string** - Работа со строками
- **number** - Математические функции  
- **list** - Операции со списками
- **date** - Работа с датами
- **context** - Контекстные функции
- **conversion** - Преобразование типов
- **logical** - Логические функции

## Связанные методы
- [EvaluateExpression](evaluate-expression.md) - Использование функций
- [ValidateExpression](validate-expression.md) - Проверка существования функций
- [TestExpression](test-expression.md) - Тестирование функций
