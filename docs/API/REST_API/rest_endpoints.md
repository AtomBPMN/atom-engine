# REST API Endpoints

Полный список всех REST API endpoints в Atom Engine.

**Базовый URL**: `/api/v1`

## Health & System

### Health Check
- `GET /health` - Проверка доступности системы

### System Management
- `GET /api/v1/system/status` - Статус системы
- `GET /api/v1/system/info` - Информация о системе  
- `GET /api/v1/system/metrics` - Метрики системы
- `GET /api/v1/system/health` - Системная проверка здоровья
- `GET /api/v1/system/components` - Список компонентов
- `GET /api/v1/system/components/:name` - Статус компонента
- `GET /api/v1/system/components/:name/health` - Здоровье компонента

### Daemon Management
- `GET /api/v1/daemon/status` - Статус демона
- `POST /api/v1/daemon/start` - Запуск демона
- `POST /api/v1/daemon/stop` - Остановка демона
- `GET /api/v1/daemon/events` - События демона

## Storage Operations

### Database Management
- `GET /api/v1/storage/status` - Статус хранилища
- `GET /api/v1/storage/info` - Информация о хранилище

## BPMN Parser

### Process Management
- `POST /api/v1/bpmn/parse` - Парсинг BPMN файла
- `GET /api/v1/bpmn/processes` - Список BPMN процессов
- `GET /api/v1/bpmn/processes/:key` - Детали BPMN процесса
- `DELETE /api/v1/bpmn/processes/:id` - Удалить BPMN процесс
- `GET /api/v1/bpmn/processes/:key/json` - JSON данные процесса
- `GET /api/v1/bpmn/processes/:key/xml` - Оригинальный BPMN XML
- `GET /api/v1/bpmn/stats` - Статистика BPMN

## Process Engine

### Process Instances
- `POST /api/v1/processes` - Запуск процесса
- `GET /api/v1/processes` - Список экземпляров процессов
- `GET /api/v1/processes/:id` - Статус экземпляра процесса
- `GET /api/v1/processes/:id/info` - Детальная информация о процессе
- `DELETE /api/v1/processes/:id` - Отмена экземпляра процесса
- `GET /api/v1/processes/:id/tokens` - Токены процесса
- `GET /api/v1/processes/:id/tokens/trace` - Трассировка токенов
- `GET /api/v1/processes/stats` - Статистика процессов

### Enhanced Process Endpoints (Typed)
- `POST /api/v1/processes/typed` - Запуск процесса (typed)
- `GET /api/v1/processes/typed` - Список процессов (typed)
- `GET /api/v1/processes/:id/typed` - Статус процесса (typed)
- `DELETE /api/v1/processes/:id/typed` - Отмена процесса (typed)
- `GET /api/v1/processes/:id/tokens/typed` - Токены процесса (typed)
- `GET /api/v1/processes/:id/trace/typed` - Трассировка процесса (typed)

## Timer Management

### Timer Operations
- `POST /api/v1/timers` - Создать таймер
- `GET /api/v1/timers` - Список таймеров
- `GET /api/v1/timers/:id` - Статус таймера
- `DELETE /api/v1/timers/:id` - Удалить таймер
- `GET /api/v1/timers/stats` - Статистика таймеров

## Job Management

### Job Operations
- `POST /api/v1/jobs` - Создать задание
- `GET /api/v1/jobs` - Список заданий
- `GET /api/v1/jobs/:key` - Детали задания
- `POST /api/v1/jobs/activate` - Активировать задания для worker
- `PUT /api/v1/jobs/:key/complete` - Завершить задание
- `PUT /api/v1/jobs/:key/fail` - Провалить задание
- `POST /api/v1/jobs/:key/throw-error` - Выбросить ошибку
- `PUT /api/v1/jobs/:key/retries` - Обновить повторы задания
- `DELETE /api/v1/jobs/:key` - Отменить задание
- `PUT /api/v1/jobs/:key/timeout` - Обновить таймаут задания
- `GET /api/v1/jobs/stats` - Статистика заданий

## Message System

### Message Operations
- `POST /api/v1/messages/publish` - Публиковать сообщение
- `GET /api/v1/messages` - Список буферизованных сообщений
- `GET /api/v1/messages/subscriptions` - Список подписок
- `GET /api/v1/messages/stats` - Статистика сообщений
- `DELETE /api/v1/messages/expired` - Очистка просроченных сообщений
- `POST /api/v1/messages/test` - Тест сообщений

## Expression Engine

### Expression Operations
- `POST /api/v1/expressions/evaluate` - Вычислить выражение
- `POST /api/v1/expressions/evaluate/batch` - Batch вычисление
- `POST /api/v1/expressions/evaluate/condition` - Вычислить условие
- `POST /api/v1/expressions/parse` - Парсить выражение в AST
- `POST /api/v1/expressions/validate` - Валидация выражения
- `POST /api/v1/expressions/test` - Тестирование выражения
- `POST /api/v1/expressions/extract-variables` - Извлечь переменные
- `GET /api/v1/expressions/functions` - Поддерживаемые функции

## Incident Management

### Incident Operations
- `POST /api/v1/incidents` - Создать инцидент
- `GET /api/v1/incidents` - Список инцидентов
- `GET /api/v1/incidents/:id` - Детали инцидента
- `PUT /api/v1/incidents/:id/resolve` - Решить инцидент
- `GET /api/v1/incidents/stats` - Статистика инцидентов

## Token Management

### Token Operations
- `GET /api/v1/tokens/:id` - Статус токена

---

**Всего REST endpoints**: 85

**Общие характеристики**:
- Все endpoints требуют авторизации (кроме /health)
- JSON формат запросов/ответов
- Стандартизованная структура ответов APIResponse
- Поддержка пагинации
- Rate limiting
- CORS поддержка
- Swagger документация
- Middleware для логирования и авторизации

**Форматы данных**:
- ISO 8601 даты/время
- JSON переменные
- FEEL expressions
- BPMN XML/JSON
- Multipart form-data для файлов
