# gRPC API Documentation

Полная документация gRPC API для Atom Engine - BPMN Process Engine с поддержкой Zeebe 8.x протокола.

## Обзор архитектуры

Atom Engine предоставляет **8 gRPC сервисов** с **47 методами** для полного управления BPMN процессами:

### Основные сервисы (Core BPMN)
- [**Parser Service**](parser/) - Парсинг и валидация BPMN (6 методов)
- [**Process Service**](process/) - Управление процессами (6 методов) 
- [**Jobs Service**](jobs/) - Управление заданиями (10 методов)
- [**Messages Service**](messages/) - Система сообщений (5 методов)

### Вспомогательные сервисы (Support)
- [**Expression Service**](expression/) - Вычисление выражений (8 методов)
- [**TimeWheel Service**](timewheel/) - Система таймеров (5 методов)
- [**Storage Service**](storage/) - Управление хранилищем (2 метода)
- [**Incidents Service**](incidents/) - Управление инцидентами (5 методов)

## Подключение к gRPC

### Базовая информация
- **Адрес**: `localhost:27500` (по умолчанию)
- **Протокол**: gRPC с Protocol Buffers
- **Авторизация**: API ключи (см. [авторизация](auth/))
- **Совместимость**: Zeebe 8.x

### Настройка клиента

**Go Client:**
```go
conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Создание клиентов сервисов
processClient := processpb.NewProcessServiceClient(conn)
jobsClient := jobspb.NewJobsServiceClient(conn)
```

**gRPCurl:**
```bash
# Список сервисов
grpcurl -plaintext localhost:27500 list

# Вызов метода
grpcurl -plaintext -d '{"process_key":"my-process"}' \
  localhost:27500 process.ProcessService/StartProcessInstance
```

## Архитектурные принципы

### Автономные компоненты
- Каждый сервис работает независимо
- Обмен данными через JSON между компонентами
- Восстановление состояния после перезапуска

### Zeebe-совместимость
- 100% совместимость с Zeebe 8.x API
- Стандартные типы данных и статусы
- Поддержка всех BPMN 2.0 элементов

### Производительность
- Persistent storage с BadgerDB
- Hierarchical timewheel для таймеров  
- Оптимизированная корреляция сообщений
- Масштабируемая система заданий

## Типичные сценарии использования

### 1. Развертывание и запуск процесса
```bash
# 1. Парсинг BPMN
grpcurl -plaintext -d @process.json localhost:27500 parser.ParserService/ParseBPMNFile

# 2. Запуск экземпляра
grpcurl -plaintext -d '{"process_key":"my-process"}' localhost:27500 process.ProcessService/StartProcessInstance
```

### 2. Обработка заданий воркером
```bash
# 1. Активация заданий
grpcurl -plaintext -d '{"type":"service-task","worker":"worker1","max_jobs":5}' localhost:27500 jobs.JobsService/ActivateJobs

# 2. Завершение задания  
grpcurl -plaintext -d '{"job_key":"job123","variables":"{}"}' localhost:27500 jobs.JobsService/CompleteJob
```

### 3. Корреляция сообщений
```bash
# 1. Публикация сообщения
grpcurl -plaintext -d '{"name":"order_created","correlation_key":"order123"}' localhost:27500 messages.MessagesService/PublishMessage

# 2. Проверка корреляции
grpcurl -plaintext -d '{"message_name":"order_created"}' localhost:27500 messages.MessagesService/ListMessageSubscriptions
```

## Обработка ошибок

Все gRPC методы используют стандартные коды ошибок:
- `OK` - Успешное выполнение
- `INVALID_ARGUMENT` - Некорректные параметры
- `NOT_FOUND` - Ресурс не найден
- `ALREADY_EXISTS` - Ресурс уже существует
- `INTERNAL` - Внутренняя ошибка сервера

## Мониторинг

- **Storage Service** - Статус и метрики хранилища
- **Incidents Service** - Отслеживание системных ошибок
- **Jobs Service** - Статистика заданий и воркеров
- **TimeWheel Service** - Метрики системы таймеров

## Разработка

### Proto файлы
Все `.proto` файлы находятся в `proto/` директории:
```
proto/
├── parser/parser.proto
├── process/process.proto  
├── jobs/jobs.proto
├── messages/messages.proto
├── expression/expression.proto
├── timewheel/timewheel.proto
├── storage/storage.proto
└── incidents/incidents.proto
```

### Генерация кода
```bash
make proto          # Генерация protobuf файлов
make build-full     # Полная сборка с proto
```

## Поддержка

- **Документация**: Каждый метод имеет детальное описание
- **Примеры**: Go, gRPCurl примеры для всех методов
- **Совместимость**: Zeebe 8.x клиенты работают без изменений