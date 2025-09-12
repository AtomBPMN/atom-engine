# gRPC API Endpoints

Полный список всех gRPC сервисов и методов в Atom Engine.

## Parser Service

**Назначение**: Парсинг и управление BPMN процессами

- `ParseBPMNFile` - Парсит BPMN файл и сохраняет в хранилище
- `ListBPMNProcesses` - Список всех BPMN процессов  
- `GetBPMNProcess` - Получить детали BPMN процесса
- `DeleteBPMNProcess` - Удалить BPMN процесс
- `GetBPMNStats` - Получить статистику парсинга BPMN
- `GetBPMNProcessJSON` - Получить JSON данные BPMN процесса
- `GetBPMNProcessXML` - Получить оригинальный BPMN XML процесса

## Process Service

**Назначение**: Управление экземплярами процессов

- `StartProcessInstance` - Запуск экземпляра процесса
- `GetProcessInstanceStatus` - Получить статус экземпляра процесса
- `CancelProcessInstance` - Отменить экземпляр процесса  
- `ListProcessInstances` - Список экземпляров процессов
- `ListTokens` - Список токенов
- `GetTokenStatus` - Получить детали токена
- `GetProcessInstanceInfo` - Получить полную информацию об экземпляре процесса

## Jobs Service

**Назначение**: Управление заданиями для service tasks

- `CreateJob` - Создать новое задание
- `ActivateJobs` - Активировать задания для worker (streaming)
- `CompleteJob` - Завершить задание
- `FailJob` - Провалить задание
- `ThrowError` - Выбросить ошибку для задания
- `UpdateJobRetries` - Обновить количество повторов задания
- `CancelJob` - Отменить задание
- `ListJobs` - Список заданий
- `GetJobStats` - Получить статистику заданий
- `GetJob` - Получить детали задания
- `UpdateJobTimeout` - Обновить таймаут задания

## Messages Service

**Назначение**: Управление сообщениями и корреляцией

- `PublishMessage` - Публикация сообщения
- `ListBufferedMessages` - Список буферизованных сообщений
- `ListMessageSubscriptions` - Список подписок на сообщения
- `GetMessageStats` - Получить статистику сообщений
- `CleanupExpiredMessages` - Очистка просроченных сообщений

## Expression Service

**Назначение**: Вычисление FEEL выражений

- `EvaluateExpression` - Вычислить FEEL выражение
- `EvaluateBatch` - Вычислить несколько выражений
- `ParseExpression` - Парсить выражение в AST
- `ValidateExpression` - Валидация синтаксиса выражения
- `GetSupportedFunctions` - Список поддерживаемых функций
- `EvaluateCondition` - Вычислить условие (boolean результат)
- `ExtractVariables` - Извлечь переменные из выражения
- `TestExpression` - Тестирование выражения с образцами данных

## TimeWheel Service

**Назначение**: Управление таймерами

- `AddTimer` - Добавить таймер в time wheel
- `RemoveTimer` - Удалить таймер из time wheel
- `GetTimerStatus` - Получить статус таймера
- `GetTimeWheelStats` - Получить статистику time wheel
- `ListTimers` - Список всех таймеров

## Storage Service

**Назначение**: Операции с базой данных

- `GetStorageStatus` - Получить статус базы данных
- `GetStorageInfo` - Получить информацию о БД (размер, статистика)

## Incidents Service

**Назначение**: Управление инцидентами

- `CreateIncident` - Создать новый инцидент
- `ResolveIncident` - Решить инцидент
- `GetIncident` - Получить инцидент по ID
- `ListIncidents` - Список инцидентов с фильтрацией
- `GetIncidentStats` - Получить статистику инцидентов

---

**Всего gRPC методов**: 56

**Поддерживаемые форматы**:
- ISO 8601 duration (PT30S, PT1H, P1D)
- ISO 8601 cycles (R5/PT30S, R/PT1M)
- JSON variables
- FEEL expressions
- Zeebe 8 compatibility
