# GetBPMNProcessXML

## Описание
Получает оригинальное XML содержимое BPMN процесса из файловой системы.

## Синтаксис
```protobuf
rpc GetBPMNProcessXML(GetBPMNProcessXMLRequest) returns (GetBPMNProcessXMLResponse);
```

## Авторизация
✅ **Требуется API ключ** с разрешением `parser`, `read` или `*`

## Параметры запроса

### GetBPMNProcessXMLRequest
```protobuf
message GetBPMNProcessXMLRequest {
  string process_key = 1;        // Ключ процесса (PROCESS KEY)
}
```

## Параметры ответа

### GetBPMNProcessXMLResponse
```protobuf
message GetBPMNProcessXMLResponse {
  bool success = 1;              // Статус успешности
  string message = 2;            // Сообщение о результате
  string xml_data = 3;           // Оригинальное BPMN XML содержимое
  string filename = 4;           // Имя файла
  int32 file_size = 5;           // Размер файла в байтах
}
```

## Пример использования

### Go
```go
client := parserpb.NewParserServiceClient(conn)
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")

response, err := client.GetBPMNProcessXML(ctx, &parserpb.GetBPMNProcessXMLRequest{
    ProcessKey: "atom-7-1k2-PVn4Y9j-CF5M",
})

if err != nil {
    log.Fatal(err)
}

if response.Success {
    fmt.Printf("Файл: %s (%d байт)\n", response.Filename, response.FileSize)
    fmt.Printf("XML:\n%s\n", response.XmlData)
} else {
    fmt.Printf("Ошибка: %s\n", response.Message)
}
```

## Возможные ошибки
- `NOT_FOUND` (5): Процесс не найден
- `PERMISSION_DENIED` (7): Недостаточно прав
- `INTERNAL` (13): Ошибка чтения файла

## Связанные методы
- [GetBPMNProcessJSON](get-bpmn-process-json.md) - JSON данные процесса
- [GetBPMNProcess](get-bpmn-process.md) - Метаданные процесса
