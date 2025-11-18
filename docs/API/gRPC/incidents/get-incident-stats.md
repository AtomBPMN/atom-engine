# GetIncidentStats

Получение статистики по инцидентам для мониторинга и аналитики.

## Информация о методе

- **Сервис**: `IncidentsService`
- **Метод**: `GetIncidentStats`
- **Тип**: Unary RPC
- **Авторизация**: Требуется

## Описание

Возвращает агрегированную статистику инцидентов:
- Распределение по статусам и типам
- Метрики времени решения
- Тренды создания инцидентов
- Топ проблемных процессов

## Протокол

### Запрос (GetIncidentStatsRequest)

```proto
message GetIncidentStatsRequest {
  string time_range = 1;         // Временной диапазон: 1h, 24h, 7d, 30d
  string group_by = 2;           // Группировка: type, status, process
}
```

### Ответ (GetIncidentStatsResponse)

```proto
message GetIncidentStatsResponse {
  int32 total_incidents = 1;     // Общее количество
  int32 open_incidents = 2;      // Открытые
  int32 resolved_incidents = 3;  // Решенные
  int32 dismissed_incidents = 4; // Отклоненные
  
  repeated TypeStats type_breakdown = 5; // По типам
  repeated ProcessStats process_breakdown = 6; // По процессам
  
  double avg_resolution_time_minutes = 7; // Среднее время решения
  int32 incidents_last_hour = 8;     // За последний час
  int32 incidents_last_day = 9;      // За последний день
}

message TypeStats {
  string type = 1;               // Тип инцидента
  int32 count = 2;               // Количество
  double percentage = 3;         // Процент от общего
}

message ProcessStats {
  string process_key = 1;        // Ключ процесса
  int32 incident_count = 2;      // Количество инцидентов
}
```

## Примеры использования

### Go Client

```go
req := &incidentspb.GetIncidentStatsRequest{
    TimeRange: "24h",
    GroupBy:   "type",
}

resp, err := client.GetIncidentStats(ctx, req)
```

### gRPCurl

```bash
grpcurl -plaintext \
  -d '{"time_range":"7d","group_by":"type"}' \
  localhost:27500 \
  incidents.IncidentsService/GetIncidentStats
```

## Возможные ошибки

- `INVALID_ARGUMENT` - Некорректный временной диапазон
- `INTERNAL` - Внутренняя ошибка сервера

## Дополнительная информация

- Обновляется в реальном времени
- Поддерживает различные временные диапазоны
- Данные кешируются для производительности
- Идеально для dashboard и мониторинга
