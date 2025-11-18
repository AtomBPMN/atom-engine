# TimeWheel Service

Высокопроизводительная система управления таймерами с иерархической структурой timewheel в Atom Engine.

## Обзор

TimeWheel Service обеспечивает точное и масштабируемое управление таймерами с производительностью O(1) для операций добавления и удаления. Система использует 5-уровневую иерархическую структуру для охвата временных интервалов от секунд до 100+ лет.

## Методы сервиса

### Управление таймерами
- **[AddTimer](add-timer.md)** - Создание новых таймеров
- **[RemoveTimer](remove-timer.md)** - Удаление таймеров
- **[GetTimerStatus](get-timer-status.md)** - Проверка состояния таймера

### Мониторинг и аналитика
- **[ListTimers](list-timers.md)** - Пагинированный список с фильтрацией
- **[GetTimeWheelStats](get-timewheel-stats.md)** - Статистика производительности

## Быстрый старт

### Go
```go
conn, _ := grpc.Dial("localhost:27500", grpc.WithInsecure())
client := timewheelpb.NewTimeWheelServiceClient(conn)

ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key")

// Создание таймера на 30 секунд
_, err := client.AddTimer(ctx, &timewheelpb.AddTimerRequest{
    TimerId:      "my-timer",
    Duration:     "PT30S",
    CallbackData: `{"action": "notify", "message": "Timer fired!"}`,
})
```

### Python
```python
channel = grpc.insecure_channel('localhost:27500')
stub = timewheel_pb2_grpc.TimeWheelServiceStub(channel)

response = stub.AddTimer(
    timewheel_pb2.AddTimerRequest(
        timer_id='my-timer',
        duration='PT1M',
        callback_data='{"type": "reminder"}'
    ),
    metadata=[('x-api-key', 'your-key')]
)
```

### JavaScript
```javascript
const client = new timewheelProto.TimeWheelService('localhost:27500',
    grpc.credentials.createInsecure());

client.addTimer({
    timer_id: 'my-timer',
    duration: 'PT5M',
    callback_data: JSON.stringify({action: 'cleanup'})
}, metadata, callback);
```

## Архитектура TimeWheel

### Иерархическая структура
```
Уровень 4: Годы    [0-99]    │ Максимальный диапазон: ~100 лет
Уровень 3: Дни     [0-30]    │ Обработка долгосрочных таймеров
Уровень 2: Часы    [0-23]    │ Ежедневные циклы
Уровень 1: Минуты  [0-59]    │ Почасовые операции
Уровень 0: Секунды [0-59]    │ Точность до секунды
```

### Производительность
- **O(1)** добавление и удаление таймеров
- **Масштабируемость** до миллионов одновременных таймеров  
- **Точность** до секунды
- **Диапазон** от секунд до 100+ лет

## ISO 8601 Поддержка

### Duration Format (PT...)
```
PT30S         - 30 секунд
PT5M          - 5 минут  
PT1H          - 1 час
PT2H30M       - 2 часа 30 минут
P1D           - 1 день
P1DT12H       - 1 день 12 часов
P1W           - 1 неделя
P1M           - 1 месяц
P1Y           - 1 год
```

### Repeating Interval Format (R...)
```
R5/PT30S      - 5 повторов каждые 30 секунд
R10/PT1M      - 10 повторов каждую минуту
R/PT15M       - Бесконечно каждые 15 минут
R3/P1D        - 3 повтора каждый день
R/P1W         - Еженедельно, бесконечно
```

## BPMN Интеграция

### Timer Start Events
```javascript
// Запуск процесса каждые 10 минут
await addTimer('process-starter', {
    interval: 'R/PT10M',
    callbackData: {
        type: 'start_event',
        process_definition_key: 'daily-report'
    }
});
```

### Boundary Timer Events
```javascript
// Таймаут для пользовательской задачи (5 минут)
await addTimer('task-timeout', {
    duration: 'PT5M',
    callbackData: {
        type: 'boundary_event',
        process_instance_id: 'proc-123',
        activity_id: 'user-task-approval',
        interrupting: true
    }
});
```

### Intermediate Timer Events
```javascript
// Пауза в процессе на 30 секунд
await addTimer('process-delay', {
    duration: 'PT30S',
    callbackData: {
        type: 'intermediate_event',
        process_instance_id: 'proc-123',
        element_id: 'intermediate-timer-1'
    }
});
```

## Типы таймеров

### BPMN Таймеры
- **`start_event`** - Timer Start Events
- **`boundary_event`** - Boundary Timer Events  
- **`intermediate_event`** - Intermediate Timer Events

### Системные таймеры
- **`reminder`** - Напоминания пользователей
- **`cleanup`** - Очистка ресурсов
- **`monitoring`** - Мониторинг системы
- **`maintenance`** - Техническое обслуживание

### Пользовательские
- Любые кастомные типы для специфических нужд

## Статусы таймера

### Жизненный цикл
```
SCHEDULED ──┐
            ├──→ FIRED      (успешное срабатывание)
            └──→ CANCELLED  (отменен до срабатывания)
```

### Описание статусов
- **`SCHEDULED`** - Активный, ожидает срабатывания
- **`FIRED`** - Сработал и выполнил callback
- **`CANCELLED`** - Отменен вручную

## Мониторинг и метрики

### Основные метрики
```javascript
const stats = await getTimeWheelStats();

// Счетчики таймеров
stats.totalTimers     // Общее количество
stats.pendingTimers   // Активных
stats.firedTimers     // Сработавших  
stats.cancelledTimers // Отмененных

// Производительность
stats.currentTick     // Текущий системный тик
stats.slotsCount     // Количество слотов
stats.timerTypes     // Распределение по типам
```

### Расчетные показатели
```javascript
const loadFactor = stats.pendingTimers / stats.slotsCount;
const successRate = stats.firedTimers / stats.totalTimers * 100;
const cancelRate = stats.cancelledTimers / stats.totalTimers * 100;
```

### Пороговые значения для алертов
- **Load Factor > 0.8**: Критическая загрузка
- **Cancel Rate > 25%**: Высокий уровень отмен
- **Success Rate < 70%**: Проблемы с надежностью

## Фильтрация и поиск

### Фильтры ListTimers
```javascript
// По статусу
await listTimers({ statusFilter: 'SCHEDULED' });

// По времени создания (новые первыми)
await listTimers({ 
    sortBy: 'created_at', 
    sortOrder: 'DESC' 
});

// Ближайшие к срабатыванию
await listTimers({ 
    statusFilter: 'SCHEDULED',
    sortBy: 'remaining_seconds', 
    sortOrder: 'ASC' 
});

// С пагинацией
await listTimers({ 
    pageSize: 50, 
    page: 2 
});
```

### Поиск по критериям
```python
# Поиск по процессу
process_timers = [t for t in all_timers 
                 if t.process_instance_id == 'proc-123']

# Поиск по типу
boundary_timers = [t for t in all_timers 
                  if t.timer_type == 'boundary_event']

# Ближайшие к срабатыванию
upcoming = [t for t in scheduled_timers 
           if t.remaining_seconds <= 300]  # 5 минут
```

## Авторизация

Все методы требуют API ключ с разрешением `timer` или `*`:

```
Headers:
x-api-key: your-api-key-here
```

## Управление жизненным циклом

### Создание таймера
1. Валидация параметров (ID, duration/interval)
2. Парсинг ISO 8601 формата
3. Размещение в соответствующем wheel уровне
4. Возврат подтверждения с scheduled_at

### Мониторинг
1. Периодическая проверка статуса
2. Отслеживание remaining_seconds
3. Обработка событий срабатывания

### Отмена
1. Поиск таймера по ID
2. Удаление из wheel структуры  
3. Обновление статуса на CANCELLED

## Рекомендации по использованию

### Лучшие практики
```javascript
// ✅ Хорошо: Описательные ID
const timerId = 'bpmn-boundary-user-task-approval-123';

// ❌ Плохо: Непонятные ID
const timerId = 'timer1';

// ✅ Хорошо: Структурированные callback данные
const callbackData = {
    type: 'boundary_event',
    process_instance_id: 'proc-123',
    activity_id: 'user-task',
    action: 'timeout'
};

// ❌ Плохо: Неструктурированные данные
const callbackData = 'timeout user task';
```

### Производительность
- Используйте пакетные операции для массовых изменений
- Мониторьте Load Factor (должен быть < 0.8)
- Очищайте завершенные таймеры периодически
- Используйте подходящие page_size для листинга

### Отладка
```javascript
// Диагностика проблемного таймера
const status = await getTimerStatus('problematic-timer');
if (status.status === 'CANCELLED') {
    console.log('Таймер был отменен неожиданно');
}

// Проверка загрузки системы
const stats = await getTimeWheelStats();
const loadFactor = stats.pendingTimers / stats.slotsCount;
if (loadFactor > 0.75) {
    console.log('⚠️ Система перегружена!');
}
```

## Интеграции

### Process Engine
TimeWheel тесно интегрирован с Process Engine для обработки BPMN таймеров:

```javascript
// Process Engine автоматически создает таймеры
// для Timer Start Events, Boundary Events, etc.
```

### Job System
Таймеры могут запускать задания:

```javascript
const callbackData = {
    type: 'create_job',
    job_type: 'cleanup',
    worker: 'maintenance-worker'
};
```

### Message System
Таймеры могут публиковать сообщения:

```javascript
const callbackData = {
    type: 'publish_message',
    message_name: 'timeout_occurred',
    correlation_key: 'proc-123'
};
```

## Примеры использования

### Система напоминаний
```python
# Напоминание через 1 час
add_timer('meeting-reminder', 'PT1H', {
    'type': 'reminder',
    'user_id': 'user-123',
    'message': 'Meeting in 5 minutes'
})

# Повторяющиеся напоминания о воде
add_timer('water-reminder', 'R8/PT30M', {
    'type': 'health_reminder', 
    'action': 'drink_water'
})
```

### Очистка ресурсов
```go
// Очистка временных файлов каждые 6 часов
addTimer("cleanup-temp", "R/PT6H", map[string]interface{}{
    "type":   "cleanup",
    "target": "temp_files",
    "older_than": "24h",
})
```

### Мониторинг системы
```javascript
// Проверка здоровья сервисов каждые 30 секунд
await addTimer('health-check', {
    interval: 'R/PT30S',
    callbackData: {
        type: 'monitoring',
        action: 'health_check',
        services: ['api', 'database', 'redis']
    }
});
```

## Связанные компоненты

- **[Process Service](../process/README.md)** - Использует таймеры для BPMN событий
- **[Jobs Service](../jobs/README.md)** - Таймеры могут создавать задания
- **[Messages Service](../messages/README.md)** - Интеграция с системой сообщений

## Миграция и совместимость

### Устаревшие поля
- `delay_ms` → используйте `duration`
- `interval_ms` → используйте `interval`  
- `limit` → используйте `page_size`

### Обратная совместимость
Система поддерживает устаревшие поля, но рекомендуется переход на новые.

## Troubleshooting

### Частые проблемы

**Таймер не срабатывает**
1. Проверьте статус: `getTimerStatus(id)`
2. Убедитесь что демон запущен
3. Проверьте загрузку системы

**Высокая загрузка**
1. Мониторьте Load Factor
2. Очищайте завершенные таймеры
3. Рассмотрите горизонтальное масштабирование

**Неточность срабатывания**
1. Проверьте системные ресурсы
2. Анализируйте метрики производительности
3. Оптимизируйте количество одновременных таймеров
