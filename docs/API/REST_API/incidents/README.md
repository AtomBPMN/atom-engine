# Incident Management API

Управление инцидентами в Atom Engine - система отслеживания и решения проблем, возникающих во время выполнения BPMN процессов.

## Обзор

Incident Management API предоставляет полный набор инструментов для работы с инцидентами системы:
- Мониторинг и отслеживание инцидентов
- Детальная диагностика проблем
- Решение инцидентов различными способами
- Аналитика и статистика

## Типы инцидентов

### JOB
Ошибки выполнения заданий (Service Tasks, Send Tasks)
- Таймауты подключения к внешним сервисам
- Ошибки аутентификации
- Превышение лимитов повторов

### EXPRESSION
Ошибки в FEEL выражениях и условиях
- Синтаксические ошибки
- Неопределенные переменные
- Ошибки типов данных

### BPMN
Ошибки структуры процесса
- Недостающие элементы
- Некорректные ссылки
- Нарушения схемы

### PROCESS
Ошибки экземпляра процесса
- Ошибки жизненного цикла
- Проблемы токенов
- Конфликты состояния

### TIMER
Ошибки таймеров и временных событий
- Некорректные форматы длительности
- Ошибки планирования
- Превышение времени ожидания

### MESSAGE
Ошибки обработки сообщений
- Проблемы корреляции
- Недоставленные сообщения
- Ошибки подписок

### SYSTEM
Системные ошибки
- Ошибки инфраструктуры
- Проблемы ресурсов
- Сбои компонентов

## Статусы инцидентов

- **OPEN** - Инцидент активен и требует внимания
- **RESOLVED** - Инцидент решен успешно
- **DISMISSED** - Инцидент отклонен без решения

## Уровни серьезности

- **LOW** - Минимальное влияние на бизнес-процессы
- **MEDIUM** - Умеренное влияние, требует внимания
- **HIGH** - Значительное влияние, требует быстрого решения
- **CRITICAL** - Критическое влияние, требует немедленного вмешательства

## Endpoints

### Список и поиск инцидентов
- [`GET /api/v1/incidents`](./list-incidents.md) - Получение списка инцидентов с фильтрацией

### Детальная информация
- [`GET /api/v1/incidents/:id`](./get-incident.md) - Подробная информация об инциденте

### Решение инцидентов
- [`POST /api/v1/incidents/:id/resolve`](./resolve-incident.md) - Решение инцидента

### Аналитика и статистика
- [`GET /api/v1/incidents/stats`](./get-incident-stats.md) - Статистика и метрики инцидентов

## Быстрый старт

### Получение списка открытых инцидентов
```bash
curl -X GET "http://localhost:27555/api/v1/incidents?status=open" \
  -H "X-API-Key: your-api-key-here"
```

### Просмотр деталей инцидента
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/srv1-inc123abc456def789" \
  -H "X-API-Key: your-api-key-here"
```

### Решение инцидента повтором
```bash
curl -X POST "http://localhost:27555/api/v1/incidents/srv1-inc123abc456def789/resolve" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "action": "retry",
    "retries": 3,
    "comment": "Resolved after service restart"
  }'
```

### Получение статистики
```bash
curl -X GET "http://localhost:27555/api/v1/incidents/stats?period=week" \
  -H "X-API-Key: your-api-key-here"
```

## Workflow решения инцидентов

### 1. Обнаружение и создание
- Система автоматически создает инцидент при ошибке
- Инцидент получает уникальный ID и классификацию
- Устанавливается уровень серьезности

### 2. Диагностика
- Сбор контекстной информации
- Анализ причин возникновения
- Определение влияния на бизнес-процессы

### 3. Решение
- **Retry** - повторить выполнение после исправления
- **Dismiss** - отклонить без исправления
- Обновление переменных процесса при необходимости

### 4. Верификация
- Проверка успешности решения
- Мониторинг повторного возникновения
- Обновление статистики

## Примеры использования

### Мониторинг критических инцидентов
```javascript
const response = await fetch('/api/v1/incidents?status=open&severity=critical', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});

const criticalIncidents = await response.json();
```

### Массовое решение однотипных инцидентов
```javascript
const incidents = ['inc1', 'inc2', 'inc3'];

for (const incidentId of incidents) {
  await fetch(`/api/v1/incidents/${incidentId}/resolve`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your-api-key-here'
    },
    body: JSON.stringify({
      action: 'retry',
      retries: 2,
      comment: 'Bulk resolution after system maintenance'
    })
  });
}
```

### Анализ трендов инцидентов
```javascript
const stats = await fetch('/api/v1/incidents/stats?period=month&group_by=type', {
  headers: { 'X-API-Key': 'your-api-key-here' }
});

const monthlyStats = await stats.json();
console.log('Most common incident type:', 
  Object.entries(monthlyStats.data.by_type)
    .sort(([,a], [,b]) => b.total - a.total)[0][0]
);
```

## Best Practices

### Мониторинг
1. Настройте уведомления для критических инцидентов
2. Регулярно проверяйте статистику трендов
3. Используйте фильтры для фокуса на важных инцидентах

### Решение
1. Анализируйте причину перед решением
2. Обновляйте переменные процесса при необходимости
3. Документируйте решения через комментарии
4. Мониторьте эффективность решений

### Предотвращение
1. Анализируйте повторяющиеся паттерны
2. Улучшайте процессы на основе статистики
3. Настройте проактивный мониторинг
4. Обучайте команду на основе инцидентов

## Авторизация

Все endpoints требуют API ключ с разрешением `incident`:

```http
X-API-Key: your-api-key-here
```

## Лимиты и пагинация

- Максимальный лимит записей: 100
- По умолчанию возвращается 10 записей
- Используйте параметры `limit` и `offset` для пагинации

## Связанные компоненты

- **Process Engine** - источник большинства инцидентов
- **Job Management** - JOB инциденты
- **Expression Engine** - EXPRESSION инциденты  
- **Timer System** - TIMER инциденты
- **Message System** - MESSAGE инциденты
