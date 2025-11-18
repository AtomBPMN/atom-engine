# Авторизация и Аутентификация

## Обзор

Atom Engine REST API использует API ключи для аутентификации и ролевую модель для авторизации.

## Методы аутентификации

### API Key Authentication

**Заголовок**: `X-API-Key`

```http
X-API-Key: your-api-key-here
```

### Формат API ключа

- **Длина**: 32 символа
- **Алфавит**: `a-zA-Z0-9_-` (URL-safe)
- **Пример**: `AbC123_xyz-789DefGhi456JklMno01`

## Разрешения (Permissions)

### Системные разрешения
- `system` - Доступ к системной информации и метрикам
- `storage` - Операции с хранилищем данных

### Бизнес-логика разрешения
- `bpmn` - Парсинг и управление BPMN процессами
- `process` - Управление экземплярами процессов
- `token` - Просмотр токенов выполнения
- `timer` - Управление таймерами
- `job` - Управление заданиями
- `message` - Публикация и управление сообщениями
- `expression` - Вычисление выражений
- `incident` - Управление инцидентами

## Уровни доступа

### Public endpoints (без авторизации)
- `GET /health` - Проверка состояния системы

### Authenticated endpoints (требуют API ключ)
Все остальные endpoints требуют валидный API ключ.

## Обработка ошибок

### 401 Unauthorized
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key",
    "details": null
  },
  "request_id": "req_123456789"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN", 
    "message": "Insufficient permissions for this operation",
    "details": {
      "required_permission": "process",
      "provided_permissions": ["bpmn", "storage"]
    }
  },
  "request_id": "req_123456789"
}
```

## Rate Limiting

### Лимиты по умолчанию
- **Requests per minute**: 60
- **Burst limit**: 100

### Заголовки ответа
```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 58
X-RateLimit-Reset: 1640995200
```

### 429 Too Many Requests
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests",
    "details": {
      "retry_after": 60
    }
  },
  "request_id": "req_123456789"
}
```

## IP Validation

### Localhost доступ
- `127.0.0.1`, `::1` - полный доступ без ограничений
- Все разрешения автоматически предоставляются

### Внешние IP
- Требуют валидный API ключ
- Проверка разрешений согласно конфигурации

## Конфигурация

### Пример конфигурации auth
```yaml
auth:
  enabled: true
  api_keys:
    - key: "your-api-key-here"
      permissions: ["system", "process", "job"]
      description: "Process management key"
  rate_limit:
    enabled: true
    requests_per_minute: 60
    burst_limit: 100
  ip_whitelist:
    - "127.0.0.1"
    - "192.168.1.0/24"
```

## Примеры использования

### cURL
```bash
curl -H "X-API-Key: your-api-key-here" \
     -X GET \
     https://atom-engine.example.com/api/v1/processes
```

### JavaScript/Fetch
```javascript
fetch('/api/v1/processes', {
  headers: {
    'X-API-Key': 'your-api-key-here',
    'Content-Type': 'application/json'
  }
})
```

### Go
```go
req, _ := http.NewRequest("GET", "/api/v1/processes", nil)
req.Header.Set("X-API-Key", "your-api-key-here")
```

## Security Best Practices

1. **Хранение ключей**: Используйте переменные окружения
2. **Ротация**: Регулярно обновляйте API ключи  
3. **Минимальные права**: Предоставляйте только необходимые разрешения
4. **Мониторинг**: Отслеживайте использование API через логи
5. **HTTPS**: Всегда используйте HTTPS в production

## Troubleshooting

### Отладка авторизации
1. Проверьте формат API ключа
2. Убедитесь в наличии требуемых разрешений
3. Проверьте IP whitelist
4. Проверьте rate limits
5. Проверьте логи системы для деталей
