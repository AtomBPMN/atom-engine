# Atom Engine API Overview

Общий обзор API архитектуры Atom Engine BPMN процессора.

## Общая статистика

### gRPC API
- **Сервисов**: 8
- **Методов**: 55 
- **Совместимость**: Zeebe 8

### REST API  
- **Endpoints**: 84
- **Групп**: 11
- **Версий**: v1

### Общий итог
- **Всего endpoints**: 139 (55 gRPC + 84 REST)

## Архитектурные особенности

### Двойной интерфейс
- gRPC для высокопроизводительного взаимодействия (55 методов)
- REST API для web интеграций + административные функции (84 endpoints)  
- 100% покрытие BPMN бизнес-логики через gRPC
- Дополнительные web/системные функции только в REST

### Сервисная архитектура
- **Parser Service** - BPMN парсинг и управление процессами
- **Process Service** - Выполнение процессов и управление экземплярами  
- **Jobs Service** - Управление заданиями service tasks
- **Messages Service** - Сообщения и корреляция
- **Expression Service** - FEEL выражения
- **TimeWheel Service** - Система таймеров
- **Storage Service** - Операции с БД
- **Incidents Service** - Управление инцидентами

### Безопасность и качество
- Авторизация через API ключи
- Rate limiting
- CORS поддержка  
- Structured logging
- Валидация данных
- Типизированные ответы

### Интеграционные возможности
- Zeebe 8 совместимость
- ISO 8601 форматы времени
- JSON/XML обработка
- Streaming для jobs
- Webhook поддержка
- Swagger документация

## Структура документации

```
docs/API/
├── api_overview.md          # Этот файл - общий обзор
├── gRPC/
│   └── grpc_endpoints.md    # gRPC методы по сервисам  
└── REST_API/
    └── rest_endpoints.md    # REST endpoints по группам
```

## Использование

### Для разработчиков
- REST API для web приложений
- gRPC для микросервисной архитектуры  
- CLI для администрирования

### Для интеграций
- Zeebe clients совместимость
- Standard BPMN 2.0
- FEEL expression engine
- Event-driven архитектура

---
