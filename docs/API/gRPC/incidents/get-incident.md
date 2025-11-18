# GetIncident

Получение детальной информации о конкретном инциденте.

## Информация о методе

- **Сервис**: `IncidentsService`
- **Метод**: `GetIncident`
- **Тип**: Unary RPC
- **Авторизация**: Требуется

## Описание

Возвращает полную информацию об инциденте включая:
- Детали ошибки и стек трейс
- Связанные объекты (процесс, задание, токен)
- История попыток решения
- Временные метки

## Протокол

### Запрос (GetIncidentRequest)

```proto
message GetIncidentRequest {
  string incident_id = 1;        // ID инцидента
}
```

### Ответ (GetIncidentResponse)

```proto
message GetIncidentResponse {
  string incident_id = 1;        // ID инцидента
  string type = 2;               // Тип: JOB, BPMN, EXPRESSION, etc.
  string status = 3;             // OPEN, RESOLVED, DISMISSED
  string error_message = 4;      // Сообщение об ошибке
  string stack_trace = 5;        // Стек трейс ошибки
  string process_instance_id = 6; // ID экземпляра процесса
  string job_key = 7;            // Ключ задания (если применимо)
  string created_at = 8;         // Время создания
  string resolved_at = 9;        // Время решения
  string resolution_comment = 10; // Комментарий к решению
  int32 retry_count = 11;        // Количество попыток
}
```

## Примеры использования

### Go Client

```go
req := &incidentspb.GetIncidentRequest{
    IncidentId: "srv1-abc123def456",
}

resp, err := client.GetIncident(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Incident Type: %s\n", resp.Type)
fmt.Printf("Status: %s\n", resp.Status)
```

### gRPCurl

```bash
grpcurl -plaintext \
  -d '{"incident_id":"srv1-abc123def456"}' \
  localhost:27500 \
  incidents.IncidentsService/GetIncident
```

## Возможные ошибки

- `NOT_FOUND` - Инцидент не найден
- `INVALID_ARGUMENT` - Некорректный ID инцидента
- `INTERNAL` - Внутренняя ошибка сервера

## Дополнительная информация

- Возвращает полную историю инцидента
- Включает техническую информацию для диагностики
- Поддерживает все типы инцидентов системы
- Безопасен для повторных вызовов
