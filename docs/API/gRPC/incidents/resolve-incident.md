# ResolveIncident

Решение системного инцидента через повторное выполнение или отклонение.

## Информация о методе

- **Сервис**: `IncidentsService`
- **Метод**: `ResolveIncident`
- **Тип**: Unary RPC
- **Авторизация**: Требуется

## Описание

Решает инцидент одним из способов:
- **retry** - повторное выполнение с указанием количества попыток
- **dismiss** - отклонение инцидента как решенного

## Протокол

### Запрос (ResolveIncidentRequest)

```proto
message ResolveIncidentRequest {
  string incident_id = 1;        // ID инцидента
  string action = 2;             // "retry" или "dismiss" 
  int32 retries = 3;             // Количество повторов (для retry)
  string comment = 4;            // Комментарий к решению
}
```

### Ответ (ResolveIncidentResponse)

```proto
message ResolveIncidentResponse {
  bool success = 1;              // Успешность операции
  string message = 2;            // Сообщение о результате
  string resolved_at = 3;        // Время решения (ISO 8601)
}
```

## Примеры использования

### Go Client

```go
req := &incidentspb.ResolveIncidentRequest{
    IncidentId: "srv1-abc123def456",
    Action:     "retry",
    Retries:    3,
    Comment:    "Fixed configuration issue",
}

resp, err := client.ResolveIncident(ctx, req)
```

### gRPCurl

```bash
grpcurl -plaintext \
  -d '{"incident_id":"srv1-abc123def456","action":"retry","retries":3}' \
  localhost:27500 \
  incidents.IncidentsService/ResolveIncident
```

## Возможные ошибки

- `NOT_FOUND` - Инцидент не найден
- `INVALID_ARGUMENT` - Некорректные параметры
- `ALREADY_EXISTS` - Инцидент уже решен
- `INTERNAL` - Внутренняя ошибка сервера

## Дополнительная информация

- Инцидент переходит в статус `RESOLVED`
- При retry создается новое задание или токен
- При dismiss инцидент помечается как решенный без действий
- Комментарий сохраняется в истории инцидента
