# ListIncidents

Получение списка инцидентов с фильтрацией и пагинацией.

## Информация о методе

- **Сервис**: `IncidentsService`
- **Метод**: `ListIncidents`
- **Тип**: Unary RPC
- **Авторизация**: Требуется

## Описание

Возвращает список инцидентов с возможностью фильтрации по:
- Статусу (OPEN, RESOLVED, DISMISSED)
- Типу (JOB, BPMN, EXPRESSION, PROCESS, TIMER, MESSAGE, SYSTEM)
- Экземпляру процесса
- Временному диапазону

## Протокол

### Запрос (ListIncidentsRequest)

```proto
message ListIncidentsRequest {
  string status = 1;             // Фильтр по статусу
  string type = 2;               // Фильтр по типу
  string process_instance_id = 3; // Фильтр по процессу
  int32 limit = 4;               // Лимит записей (по умолчанию 10)
  int32 offset = 5;              // Смещение для пагинации
  string sort_by = 6;            // Поле сортировки
  string sort_order = 7;         // asc/desc
}
```

### Ответ (ListIncidentsResponse)

```proto
message ListIncidentsResponse {
  repeated IncidentSummary incidents = 1; // Список инцидентов
  int32 total_count = 2;         // Общее количество
  bool has_more = 3;             // Есть ли еще записи
}

message IncidentSummary {
  string incident_id = 1;        // ID инцидента
  string type = 2;               // Тип инцидента
  string status = 3;             // Статус
  string error_message = 4;      // Краткое сообщение об ошибке
  string process_instance_id = 5; // ID процесса
  string created_at = 6;         // Время создания
}
```

## Примеры использования

### Go Client

```go
req := &incidentspb.ListIncidentsRequest{
    Status: "OPEN",
    Type:   "JOB",
    Limit:  20,
}

resp, err := client.ListIncidents(ctx, req)
```

### gRPCurl

```bash
grpcurl -plaintext \
  -d '{"status":"OPEN","limit":10}' \
  localhost:27500 \
  incidents.IncidentsService/ListIncidents
```

## Возможные ошибки

- `INVALID_ARGUMENT` - Некорректные параметры фильтрации
- `INTERNAL` - Внутренняя ошибка сервера

## Дополнительная информация

- Поддерживает пагинацию для больших наборов данных
- Результаты сортируются по времени создания (убывание)
- Пустые фильтры возвращают все инциденты
- Максимальный лимит: 1000 записей за запрос
