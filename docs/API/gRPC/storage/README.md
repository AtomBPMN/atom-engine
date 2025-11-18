# Storage Service

Служба мониторинга и управления системой хранения данных в Atom Engine на базе BadgerDB.

## Обзор

Storage Service предоставляет инструменты для мониторинга состояния базы данных, анализа использования дискового пространства и диагностики производительности системы хранения.

## Методы сервиса

### Мониторинг состояния
- **[GetStorageStatus](get-storage-status.md)** - Статус подключения и работоспособности
- **[GetStorageInfo](get-storage-info.md)** - Подробная информация и статистика

## Быстрый старт

### Go
```go
conn, _ := grpc.Dial("localhost:27500", grpc.WithInsecure())
client := storagepb.NewStorageServiceClient(conn)

ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key")

// Проверка статуса хранилища
statusResp, _ := client.GetStorageStatus(ctx, &storagepb.GetStorageStatusRequest{})
fmt.Printf("База подключена: %v\n", statusResp.IsConnected)

// Получение информации о размере
infoResp, _ := client.GetStorageInfo(ctx, &storagepb.GetStorageInfoRequest{})
fmt.Printf("Использовано: %d байт\n", infoResp.UsedSizeBytes)
```

### Python
```python
channel = grpc.insecure_channel('localhost:27500')
stub = storage_pb2_grpc.StorageServiceStub(channel)

# Проверка статуса
status = stub.GetStorageStatus(
    storage_pb2.GetStorageStatusRequest(),
    metadata=[('x-api-key', 'your-key')]
)

if status.is_connected and status.is_healthy:
    print("✅ Хранилище готово")
else:
    print("❌ Проблемы с хранилищем")
```

### JavaScript
```javascript
const client = new storageProto.StorageService('localhost:27500',
    grpc.credentials.createInsecure());

// Мониторинг размера БД
client.getStorageInfo({}, metadata, (error, response) => {
    const usagePercent = (response.used_size_bytes / response.total_size_bytes) * 100;
    console.log(`Использование БД: ${usagePercent.toFixed(1)}%`);
});
```

## BadgerDB Architecture

### Компоненты хранилища
```
BadgerDB
├── Value Log       │ Хранение значений
├── LSM Tree        │ Индексы и метаданные  
├── WAL             │ Write-Ahead Log
└── Bloom Filters   │ Быстрый поиск
```

### Характеристики
- **Тип**: Встраиваемая key-value БД
- **Производительность**: Высокая скорость записи/чтения
- **ACID**: Поддержка транзакций
- **Компрессия**: Автоматическое сжатие данных

## Мониторинг состояния

### Проверка готовности системы
```python
def wait_for_storage_ready(timeout_seconds=60):
    monitor = StorageMonitor()
    return monitor.wait_for_ready(timeout_seconds)

# Использование в CI/CD
if not wait_for_storage_ready(30):
    print("❌ Storage not ready")
    exit(1)
```

### Health Check для Kubernetes
```yaml
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - atomd storage status | grep -q "healthy.*true"
  initialDelaySeconds: 10
  periodSeconds: 30
```

### Docker Compose пример
```yaml
healthcheck:
  test: ["CMD", "atomd", "storage", "status"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

## Анализ использования дискового пространства

### Базовые метрики
```javascript
const info = await getStorageInfo();

// Основные показатели
const usagePercent = (info.usedSizeBytes / info.totalSizeBytes) * 100;
const avgRecordSize = info.usedSizeBytes / info.totalKeys;
const efficiency = info.usedSizeBytes / (info.usedSizeBytes + info.freeSizeBytes);

console.log(`Использование: ${usagePercent.toFixed(1)}%`);
console.log(`Средний размер записи: ${avgRecordSize} байт`);
console.log(`Эффективность: ${efficiency.toFixed(3)}`);
```

### Прогнозирование роста
```python
class GrowthPredictor:
    def __init__(self, monitor):
        self.monitor = monitor
        self.history = []
    
    def predict_full_date(self):
        """Прогнозирует когда закончится место"""
        if len(self.history) < 2:
            return None
            
        # Линейная регрессия по росту размера
        growth_rate = self.calculate_growth_rate()
        current_info = self.history[-1]
        
        remaining_space = current_info['free_size_bytes']
        if growth_rate <= 0:
            return None
            
        hours_until_full = remaining_space / growth_rate
        return datetime.now() + timedelta(hours=hours_until_full)
```

## Статистика производительности

### BadgerDB метрики
```json
{
  "statistics": {
    "compactions": "15",           // Количество компактификаций
    "level0_files": "3",           // Файлы уровня 0
    "level1_files": "12",          // Файлы уровня 1  
    "bloom_filter_memory": "1MB",  // Память bloom фильтров
    "index_memory": "2MB",         // Память индексов
    "read_operations": "152436",   // Операции чтения
    "write_operations": "89234"    // Операции записи
  }
}
```

### Интерпретация метрик
- **Частые компактификации**: Высокая нагрузка на запись
- **Много файлов L0**: Возможна необходимость тюнинга
- **Большая память индексов**: Много ключей в БД
- **Соотношение read/write**: Паттерн использования

## Алерты и мониторинг

### Критические пороги
```javascript
const THRESHOLDS = {
    CRITICAL_USAGE: 90,      // % использования диска
    HIGH_USAGE: 75,          // % предупреждение
    MAX_KEYS: 10_000_000,    // Максимум ключей
    MAX_RECORD_SIZE: 100_1024 // Максимальный размер записи
};

function checkThresholds(info) {
    const alerts = [];
    
    const usagePercent = (info.usedSizeBytes / info.totalSizeBytes) * 100;
    if (usagePercent > THRESHOLDS.CRITICAL_USAGE) {
        alerts.push({
            level: 'CRITICAL',
            message: `Использование диска: ${usagePercent.toFixed(1)}%`
        });
    }
    
    if (info.totalKeys > THRESHOLDS.MAX_KEYS) {
        alerts.push({
            level: 'WARNING', 
            message: `Слишком много ключей: ${info.totalKeys}`
        });
    }
    
    return alerts;
}
```

### Системы мониторинга
```bash
# Prometheus метрики
curl -s http://localhost:27555/metrics | grep storage_

# Grafana Dashboard
- Использование диска по времени
- Рост количества ключей
- Производительность операций
- Время отклика хранилища
```

## Оптимизация производительности

### Рекомендации по настройке
```go
// Оптимальные настройки BadgerDB
opts := badger.DefaultOptions(dataDir).
    WithNumMemtables(2).              // Количество мемтаблиц
    WithNumLevelZeroTables(1).        // L0 таблиц
    WithNumLevelZeroTablesStall(2).   // Порог остановки записи
    WithValueThreshold(1024).         // Порог для value log
    WithNumCompactors(2)              // Количество компакторов
```

### Стратегии очистки данных
```python
def cleanup_old_data(days_to_keep=30):
    """Очистка данных старше указанного периода"""
    cutoff_date = datetime.now() - timedelta(days=days_to_keep)
    
    # Получаем статистику до очистки
    before_info = get_storage_info()
    
    # Выполняем очистку (логика зависит от схемы данных)
    cleanup_count = perform_cleanup(cutoff_date)
    
    # Статистика после очистки
    after_info = get_storage_info()
    
    freed_space = before_info['used_size_bytes'] - after_info['used_size_bytes']
    print(f"Очищено {cleanup_count} записей")
    print(f"Освобождено {format_bytes(freed_space)}")
```

## Резервное копирование

### Мониторинг для бэкапов
```bash
#!/bin/bash
# Скрипт проверки готовности к бэкапу

# Проверяем статус хранилища
if ! atomd storage status | grep -q "healthy.*true"; then
    echo "❌ Storage not healthy, skipping backup"
    exit 1
fi

# Проверяем размер БД
USED_SIZE=$(atomd storage info | grep "used_size" | cut -d: -f2)
if [ "$USED_SIZE" -gt 10737418240 ]; then # > 10GB
    echo "⚠️ Large database size: $USED_SIZE bytes"
fi

echo "✅ Ready for backup"
```

### Снэпшоты состояния
```go
func createStorageSnapshot() (*StorageSnapshot, error) {
    status, err := client.GetStorageStatus(ctx, &pb.GetStorageStatusRequest{})
    if err != nil {
        return nil, err
    }
    
    info, err := client.GetStorageInfo(ctx, &pb.GetStorageInfoRequest{})
    if err != nil {
        return nil, err
    }
    
    return &StorageSnapshot{
        Timestamp:       time.Now(),
        IsHealthy:       status.IsHealthy,
        UsedSizeBytes:   info.UsedSizeBytes,
        TotalKeys:       info.TotalKeys,
        DatabasePath:    info.DatabasePath,
        UptimeSeconds:   status.UptimeSeconds,
    }, nil
}
```

## Troubleshooting

### Частые проблемы

**База данных не подключается**
```bash
# Проверка прав доступа
ls -la /path/to/database/
# Проверка места на диске
df -h /path/to/database/
# Проверка процесса демона
ps aux | grep atomd
```

**Медленная производительность**
```python
def diagnose_performance():
    info = get_storage_info()
    
    # Проверяем размер записей
    avg_size = info['used_size_bytes'] / info['total_keys']
    if avg_size > 50 * 1024:  # > 50KB
        print("⚠️ Большие записи могут замедлять работу")
    
    # Проверяем фрагментацию
    efficiency = info['used_size_bytes'] / info['total_size_bytes']
    if efficiency < 0.7:  # < 70%
        print("⚠️ Возможна фрагментация, рассмотрите компактификацию")
```

**Нехватка места**
```javascript
async function handleLowSpace(info) {
    const usagePercent = (info.usedSizeBytes / info.totalSizeBytes) * 100;
    
    if (usagePercent > 90) {
        // Экстренные меры
        console.log('🚨 Критически мало места!');
        await emergencyCleanup();
        await compactDatabase();
    } else if (usagePercent > 75) {
        // Плановая очистка
        console.log('⚠️ Планируем очистку данных');
        await scheduleCleanup();
    }
}
```

## Авторизация

Все методы требуют API ключ с разрешением `storage` или `*`:

```
Headers:
x-api-key: your-api-key-here
```

## Интеграции

### Мониторинг системы
Storage Service интегрируется с системой мониторинга для:
- Отправки алертов о критическом состоянии
- Сбора метрик производительности
- Автоматического масштабирования

### CI/CD Pipeline
```yaml
# .github/workflows/deploy.yml
- name: Check Storage Health
  run: |
    if ! atomd storage status | grep -q "healthy.*true"; then
      echo "Storage not ready for deployment"
      exit 1
    fi
```

### Container Orchestration
```yaml
# Kubernetes Deployment
spec:
  containers:
  - name: atom-engine
    readinessProbe:
      exec:
        command: ["/bin/atomd", "storage", "status"]
      initialDelaySeconds: 5
      periodSeconds: 10
```

## Связанные компоненты

Все остальные сервисы зависят от Storage Service:
- **[Process Service](../process/README.md)** - Хранение состояний процессов
- **[Jobs Service](../jobs/README.md)** - Персистентность заданий  
- **[TimeWheel Service](../timewheel/README.md)** - Сохранение таймеров
- **[Messages Service](../messages/README.md)** - Буферизация сообщений

## Лучшие практики

### Производство
1. **Мониторинг**: Настройте алерты на критические метрики
2. **Резервирование**: Регулярные бэкапы БД
3. **Планирование**: Мониторьте рост и планируйте расширение
4. **Очистка**: Автоматизируйте удаление старых данных

### Разработка
1. **Тестирование**: Проверяйте готовность хранилища в тестах
2. **Логирование**: Логируйте операции с БД для отладки  
3. **Профилирование**: Анализируйте производительность запросов
4. **Миграции**: Версионируйте изменения схемы данных

### DevOps
1. **Автоматизация**: Включите проверки хранилища в пайплайны
2. **Масштабирование**: Мониторьте потребность в ресурсах
3. **Восстановление**: Тестируйте процедуры восстановления
4. **Документирование**: Ведите документацию по обслуживанию
