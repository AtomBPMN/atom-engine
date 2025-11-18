# REST API Детальная Документация

Детальная документация для каждого REST API endpoint Atom Engine.

## Структура документации

### 🔐 Авторизация
- [Методы авторизации и аутентификации](auth/README.md)

### 💓 Health & System  
- [GET /health](health/health-check.md) - Проверка доступности системы
- [GET /api/v1/system/status](system/system-status.md) - Статус системы
- [GET /api/v1/system/info](system/system-info.md) - Информация о системе
- [GET /api/v1/system/metrics](system/system-metrics.md) - Метрики системы
- [GET /api/v1/system/health](system/system-health.md) - Системная проверка здоровья
- [GET /api/v1/system/components](system/list-components.md) - Список компонентов
- [GET /api/v1/system/components/:name](system/get-component-status.md) - Статус компонента
- [GET /api/v1/system/components/:name/health](system/get-component-health.md) - Здоровье компонента

### ⚙️ Daemon Management
- [GET /api/v1/daemon/status](daemon/daemon-status.md) - Статус демона
- [POST /api/v1/daemon/start](daemon/daemon-start.md) - Запуск демона
- [POST /api/v1/daemon/stop](daemon/daemon-stop.md) - Остановка демона
- [GET /api/v1/daemon/events](daemon/daemon-events.md) - События демона

### 💾 Storage Operations
- [GET /api/v1/storage/status](storage/storage-status.md) - Статус хранилища
- [GET /api/v1/storage/info](storage/storage-info.md) - Информация о хранилище

### 📋 BPMN Parser
- [POST /api/v1/bpmn/parse](bpmn/parse-bpmn.md) - Парсинг BPMN файла
- [GET /api/v1/bpmn/processes](bpmn/list-processes.md) - Список BPMN процессов
- [GET /api/v1/bpmn/processes/:key](bpmn/get-process.md) - Детали BPMN процесса
- [DELETE /api/v1/bpmn/processes/:id](bpmn/delete-process.md) - Удалить BPMN процесс
- [GET /api/v1/bpmn/processes/:key/json](bpmn/get-process-json.md) - JSON данные процесса
- [GET /api/v1/bpmn/stats](bpmn/get-bpmn-stats.md) - Статистика BPMN

### 🔄 Process Engine
- [POST /api/v1/processes](processes/start-process.md) - Запуск процесса
- [GET /api/v1/processes](processes/list-processes.md) - Список экземпляров процессов
- [GET /api/v1/processes/:id](processes/get-process-status.md) - Статус экземпляра процесса
- [GET /api/v1/processes/:id/info](processes/get-process-info.md) - Детальная информация о процессе
- [DELETE /api/v1/processes/:id](processes/cancel-process.md) - Отмена экземпляра процесса
- [GET /api/v1/processes/:id/tokens](processes/get-process-tokens.md) - Токены процесса
- [GET /api/v1/processes/:id/tokens/trace](processes/get-token-trace.md) - Трассировка токенов
- [GET /api/v1/processes/stats](processes/get-process-stats.md) - Статистика процессов

#### Enhanced Process Endpoints (Typed)
- [POST /api/v1/processes/typed](processes/start-process-typed.md) - Запуск процесса (typed)
- [GET /api/v1/processes/typed](processes/list-processes-typed.md) - Список процессов (typed)
- [GET /api/v1/processes/:id/typed](processes/get-process-status-typed.md) - Статус процесса (typed)
- [DELETE /api/v1/processes/:id/typed](processes/cancel-process-typed.md) - Отмена процесса (typed)
- [GET /api/v1/processes/:id/tokens/typed](processes/get-process-tokens-typed.md) - Токены процесса (typed)
- [GET /api/v1/processes/:id/trace/typed](processes/trace-process-execution-typed.md) - Трассировка процесса (typed)

### ⏰ Timer Management
- [POST /api/v1/timers](timers/create-timer.md) - Создать таймер
- [GET /api/v1/timers](timers/list-timers.md) - Список таймеров
- [GET /api/v1/timers/:id](timers/get-timer.md) - Статус таймера
- [DELETE /api/v1/timers/:id](timers/delete-timer.md) - Удалить таймер
- [GET /api/v1/timers/stats](timers/get-timer-stats.md) - Статистика таймеров

### 🔧 Job Management
- [POST /api/v1/jobs](jobs/create-job.md) - Создать задание
- [GET /api/v1/jobs](jobs/list-jobs.md) - Список заданий
- [GET /api/v1/jobs/:key](jobs/get-job.md) - Детали задания
- [POST /api/v1/jobs/activate](jobs/activate-jobs.md) - Активировать задания для worker
- [PUT /api/v1/jobs/:key/complete](jobs/complete-job.md) - Завершить задание
- [PUT /api/v1/jobs/:key/fail](jobs/fail-job.md) - Провалить задание
- [POST /api/v1/jobs/:key/throw-error](jobs/throw-error.md) - Выбросить ошибку
- [PUT /api/v1/jobs/:key/retries](jobs/update-job-retries.md) - Обновить повторы задания
- [DELETE /api/v1/jobs/:key](jobs/cancel-job.md) - Отменить задание
- [PUT /api/v1/jobs/:key/timeout](jobs/update-job-timeout.md) - Обновить таймаут задания
- [GET /api/v1/jobs/stats](jobs/get-job-stats.md) - Статистика заданий

### 💬 Message System
- [POST /api/v1/messages/publish](messages/publish-message.md) - Публиковать сообщение
- [GET /api/v1/messages](messages/list-buffered-messages.md) - Список буферизованных сообщений
- [GET /api/v1/messages/subscriptions](messages/list-subscriptions.md) - Список подписок
- [GET /api/v1/messages/stats](messages/get-message-stats.md) - Статистика сообщений
- [DELETE /api/v1/messages/expired](messages/cleanup-expired.md) - Очистка просроченных сообщений
- [POST /api/v1/messages/test](messages/test-message.md) - Тест сообщений

### 🧮 Expression Engine
- [POST /api/v1/expressions/evaluate](expressions/evaluate-expression.md) - Вычислить выражение
- [POST /api/v1/expressions/evaluate/batch](expressions/evaluate-batch.md) - Batch вычисление
- [POST /api/v1/expressions/evaluate/condition](expressions/evaluate-condition.md) - Вычислить условие
- [POST /api/v1/expressions/parse](expressions/parse-expression.md) - Парсить выражение в AST
- [POST /api/v1/expressions/validate](expressions/validate-expression.md) - Валидация выражения
- [POST /api/v1/expressions/test](expressions/test-expression.md) - Тестирование выражения
- [POST /api/v1/expressions/extract-variables](expressions/extract-variables.md) - Извлечь переменные
- [GET /api/v1/expressions/functions](expressions/get-supported-functions.md) - Поддерживаемые функции

### 🚨 Incident Management
- [POST /api/v1/incidents](incidents/create-incident.md) - Создать инцидент
- [GET /api/v1/incidents](incidents/list-incidents.md) - Список инцидентов
- [GET /api/v1/incidents/:id](incidents/get-incident.md) - Детали инцидента
- [PUT /api/v1/incidents/:id/resolve](incidents/resolve-incident.md) - Решить инцидент
- [GET /api/v1/incidents/stats](incidents/get-incident-stats.md) - Статистика инцидентов

### 🎯 Token Management
- [GET /api/v1/tokens/:id](tokens/get-token-status.md) - Статус токена

## Формат документации

Каждый endpoint содержит:
- **Описание** - назначение и функциональность
- **URL и методы** - точный путь и HTTP метод
- **Авторизация** - требования к доступу
- **Параметры** - path, query, body параметры
- **Примеры запросов** - cURL, JavaScript, Go
- **Ответы** - все возможные ответы с примерами
- **Валидация** - правила валидации параметров
- **Ограничения** - лимиты и ограничения
- **Использование** - практические примеры
- **Связанные endpoints** - ссылки на связанные API

## Общие принципы

### Стандартизованные ответы
Все endpoints возвращают стандартизованную структуру:
```json
{
  "success": true/false,
  "data": { ... },           // При success: true
  "error": { ... },          // При success: false
  "request_id": "req_..."
}
```

### Коды ошибок
- `UNAUTHORIZED` - Неверный или отсутствующий API ключ
- `FORBIDDEN` - Недостаточно прав доступа
- `VALIDATION_ERROR` - Ошибки валидации данных
- `NOT_FOUND` - Ресурс не найден
- `CONFLICT` - Конфликт состояния
- `RATE_LIMITED` - Превышен лимит запросов
- `INTERNAL_ERROR` - Внутренняя ошибка сервера

### HTTP статус коды
- `200` - Успешный запрос
- `201` - Ресурс создан
- `400` - Неверный запрос
- `401` - Не авторизован
- `403` - Доступ запрещен
- `404` - Не найдено
- `409` - Конфликт
- `429` - Слишком много запросов
- `500` - Внутренняя ошибка сервера

## Быстрый старт

### 1. Получение API ключа
```bash
# Настройка в config/config.yaml
auth:
  api_keys:
    - key: "your-api-key-here"
      permissions: ["process", "job", "message"]
```

### 2. Проверка доступности
```bash
curl http://localhost:27555/health
```

### 3. Первый запрос с авторизацией
```bash
curl -H "X-API-Key: your-api-key-here" \
     http://localhost:27555/api/v1/system/status
```

### 4. Запуск процесса
```bash
curl -X POST http://localhost:27555/api/v1/processes \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{"process_id": "my-process", "variables": {"key": "value"}}'
```

---

**Всего документировано**: 84 endpoints  
**Статус**: Детальная документация создана  
**Обновлено**: 2025-01-11
