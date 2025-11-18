# GetStorageStatus

## Описание
Проверяет состояние подключения к базе данных и общее здоровье системы хранения. Возвращает информацию о подключении, работоспособности и времени работы.

## Синтаксис
```protobuf
rpc GetStorageStatus(GetStorageStatusRequest) returns (GetStorageStatusResponse);
```

## Package
```protobuf
package atom.storage.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `storage` или `*`

## Параметры запроса

### GetStorageStatusRequest
```protobuf
message GetStorageStatusRequest {}
```

## Параметры ответа

### GetStorageStatusResponse
```protobuf
message GetStorageStatusResponse {
  bool is_connected = 1;        // Подключена ли база данных
  bool is_healthy = 2;          // Работоспособность системы
  string status = 3;            // Текстовое описание статуса
  int64 uptime_seconds = 4;     // Время работы в секундах
}
```

## Примеры использования

### Go
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    
    pb "atom-engine/proto/storage/storagepb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewStorageServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Проверяем статус хранилища
    response, err := client.GetStorageStatus(ctx, &pb.GetStorageStatusRequest{})
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("🗄️ СТАТУС ХРАНИЛИЩА")
    fmt.Println("=" * 30)
    
    // Основная информация
    connStatus := "❌"
    if response.IsConnected {
        connStatus = "✅"
    }
    
    healthStatus := "❌"  
    if response.IsHealthy {
        healthStatus = "✅"
    }
    
    fmt.Printf("%s Подключение: %s\n", connStatus, boolToString(response.IsConnected))
    fmt.Printf("%s Работоспособность: %s\n", healthStatus, boolToString(response.IsHealthy))
    fmt.Printf("📊 Статус: %s\n", response.Status)
    
    // Время работы
    uptime := time.Duration(response.UptimeSeconds) * time.Second
    fmt.Printf("⏱️ Время работы: %s\n", formatUptime(uptime))
    
    // Общая оценка
    if response.IsConnected && response.IsHealthy {
        fmt.Println("🟢 Система хранения работает нормально")
    } else if response.IsConnected && !response.IsHealthy {
        fmt.Println("🟡 База подключена, но есть проблемы с работоспособностью")
    } else {
        fmt.Println("🔴 Критическая проблема: нет подключения к базе данных")
    }
}

func boolToString(b bool) string {
    if b {
        return "подключена"
    }
    return "отключена"
}

func formatUptime(d time.Duration) string {
    days := int(d.Hours()) / 24
    hours := int(d.Hours()) % 24
    minutes := int(d.Minutes()) % 60
    
    if days > 0 {
        return fmt.Sprintf("%dд %dч %dм", days, hours, minutes)
    } else if hours > 0 {
        return fmt.Sprintf("%dч %dм", hours, minutes)
    } else {
        return fmt.Sprintf("%dм", minutes)
    }
}

// Мониторинг состояния хранилища
func monitorStorageHealth(client pb.StorageServiceClient, ctx context.Context, interval time.Duration) {
    fmt.Printf("🔍 Мониторинг состояния хранилища каждые %v\n", interval)
    fmt.Printf("%-12s | %-10s | %-8s | %s\n", "Время", "Подключение", "Здоровье", "Статус")
    fmt.Printf("%s\n", strings.Repeat("-", 50))
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            response, err := client.GetStorageStatus(ctx, &pb.GetStorageStatusRequest{})
            if err != nil {
                fmt.Printf("%-12s | ❌ ОШИБКА: %v\n", time.Now().Format("15:04:05"), err)
                continue
            }
            
            connIcon := "❌"
            if response.IsConnected {
                connIcon = "✅"
            }
            
            healthIcon := "❌"
            if response.IsHealthy {
                healthIcon = "✅"
            }
            
            fmt.Printf("%-12s | %-10s | %-8s | %s\n",
                time.Now().Format("15:04:05"),
                connIcon,
                healthIcon,
                response.Status)
        }
    }
}

// Проверка готовности системы
func waitForStorageReady(client pb.StorageServiceClient, ctx context.Context, timeout time.Duration) error {
    fmt.Printf("⏳ Ожидание готовности хранилища (таймаут: %v)...\n", timeout)
    
    timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-timeoutCtx.Done():
            return fmt.Errorf("таймаут ожидания готовности хранилища")
            
        case <-ticker.C:
            response, err := client.GetStorageStatus(ctx, &pb.GetStorageStatusRequest{})
            if err != nil {
                fmt.Printf("⚠️ Ошибка проверки: %v\n", err)
                continue
            }
            
            if response.IsConnected && response.IsHealthy {
                fmt.Printf("✅ Хранилище готово! Статус: %s\n", response.Status)
                return nil
            }
            
            fmt.Printf("⏳ Ожидание... Подключение: %v, Здоровье: %v\n", 
                response.IsConnected, response.IsHealthy)
        }
    }
}
```

### Python
```python
import grpc
import time
from datetime import timedelta

import storage_pb2
import storage_pb2_grpc

def get_storage_status():
    channel = grpc.insecure_channel('localhost:27500')
    stub = storage_pb2_grpc.StorageServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = storage_pb2.GetStorageStatusRequest()
    
    try:
        response = stub.GetStorageStatus(request, metadata=metadata)
        
        print("🗄️ СТАТУС ХРАНИЛИЩА")
        print("=" * 30)
        
        # Основная информация
        conn_icon = "✅" if response.is_connected else "❌"
        health_icon = "✅" if response.is_healthy else "❌"
        
        print(f"{conn_icon} Подключение: {'подключена' if response.is_connected else 'отключена'}")
        print(f"{health_icon} Работоспособность: {'здорова' if response.is_healthy else 'проблемы'}")
        print(f"📊 Статус: {response.status}")
        
        # Время работы
        uptime_td = timedelta(seconds=response.uptime_seconds)
        print(f"⏱️ Время работы: {format_uptime(uptime_td)}")
        
        # Общая оценка
        if response.is_connected and response.is_healthy:
            print("🟢 Система хранения работает нормально")
        elif response.is_connected and not response.is_healthy:
            print("🟡 База подключена, но есть проблемы")
        else:
            print("🔴 Критическая проблема: нет подключения к БД")
        
        return {
            'is_connected': response.is_connected,
            'is_healthy': response.is_healthy,
            'status': response.status,
            'uptime_seconds': response.uptime_seconds
        }
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

def format_uptime(td):
    days = td.days
    hours, remainder = divmod(td.seconds, 3600)
    minutes, _ = divmod(remainder, 60)
    
    if days > 0:
        return f"{days}д {hours}ч {minutes}м"
    elif hours > 0:
        return f"{hours}ч {minutes}м"
    else:
        return f"{minutes}м"

# Класс для мониторинга хранилища
class StorageMonitor:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = storage_pb2_grpc.StorageServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
        self.alerts_sent = set()
    
    def check_status(self):
        """Проверяет статус хранилища"""
        try:
            request = storage_pb2.GetStorageStatusRequest()
            response = self.stub.GetStorageStatus(request, metadata=self.metadata)
            
            return {
                'is_connected': response.is_connected,
                'is_healthy': response.is_healthy,
                'status': response.status,
                'uptime_seconds': response.uptime_seconds,
                'timestamp': time.time()
            }
        except grpc.RpcError as e:
            return {
                'error': f"{e.code()} - {e.details()}",
                'timestamp': time.time()
            }
    
    def check_and_alert(self):
        """Проверяет статус и отправляет алерты при проблемах"""
        status = self.check_status()
        
        if 'error' in status:
            print(f"❌ Ошибка мониторинга: {status['error']}")
            return status
        
        # Проверяем критические проблемы
        if not status['is_connected']:
            alert_key = 'db_disconnected'
            if alert_key not in self.alerts_sent:
                self.send_alert("🚨 КРИТИЧНО: База данных отключена!", status)
                self.alerts_sent.add(alert_key)
        else:
            # Убираем алерт если проблема решена
            self.alerts_sent.discard('db_disconnected')
        
        if status['is_connected'] and not status['is_healthy']:
            alert_key = 'db_unhealthy'
            if alert_key not in self.alerts_sent:
                self.send_alert("⚠️ ВНИМАНИЕ: Проблемы с работоспособностью БД", status)
                self.alerts_sent.add(alert_key)
        else:
            self.alerts_sent.discard('db_unhealthy')
        
        return status
    
    def send_alert(self, message, status):
        """Отправляет алерт (здесь просто выводит в консоль)"""
        timestamp = time.strftime('%Y-%m-%d %H:%M:%S')
        print(f"\n[{timestamp}] {message}")
        print(f"   Статус: {status.get('status', 'неизвестен')}")
        print(f"   Время работы: {status.get('uptime_seconds', 0)}с")
        print()
    
    def continuous_monitoring(self, interval_seconds=30):
        """Непрерывный мониторинг с заданным интервалом"""
        print(f"🚀 Запуск мониторинга хранилища каждые {interval_seconds} секунд")
        print("Время       | Связь | Здоровье | Статус")
        print("-" * 45)
        
        try:
            while True:
                status = self.check_and_alert()
                
                if 'error' not in status:
                    conn_icon = "✅" if status['is_connected'] else "❌"
                    health_icon = "✅" if status['is_healthy'] else "❌"
                    
                    current_time = time.strftime('%H:%M:%S')
                    print(f"{current_time} | {conn_icon:^5} | {health_icon:^8} | {status['status']}")
                else:
                    current_time = time.strftime('%H:%M:%S')
                    print(f"{current_time} | {'❌':^5} | {'❌':^8} | ERROR")
                
                time.sleep(interval_seconds)
                
        except KeyboardInterrupt:
            print("\n🛑 Мониторинг остановлен")
    
    def wait_for_ready(self, timeout_seconds=60):
        """Ждет пока хранилище станет готовым"""
        print(f"⏳ Ожидание готовности хранилища (таймаут: {timeout_seconds}с)...")
        
        start_time = time.time()
        
        while time.time() - start_time < timeout_seconds:
            status = self.check_status()
            
            if 'error' not in status and status['is_connected'] and status['is_healthy']:
                print(f"✅ Хранилище готово! Статус: {status['status']}")
                return True
            
            if 'error' not in status:
                print(f"⏳ Ожидание... Связь: {status['is_connected']}, "
                      f"Здоровье: {status['is_healthy']}")
            else:
                print(f"⏳ Ожидание... Ошибка: {status['error']}")
            
            time.sleep(2)
        
        print("❌ Таймаут: хранилище не готово")
        return False

# Функция для проверки готовности к запуску
def pre_start_check():
    """Проверки перед запуском системы"""
    print("🔍 ПРОВЕРКА ГОТОВНОСТИ СИСТЕМЫ")
    print("=" * 40)
    
    status = get_storage_status()
    
    if not status:
        print("❌ Не удалось подключиться к системе хранения")
        return False
    
    if not status['is_connected']:
        print("❌ База данных не подключена")
        return False
    
    if not status['is_healthy']:
        print("⚠️ Проблемы с работоспособностью БД")
        print("🔧 Рекомендуется проверить логи и исправить проблемы")
        return False
    
    print("✅ Система хранения готова к работе")
    return True

# Диагностические утилиты
def diagnose_storage_issues():
    """Диагностирует проблемы с хранилищем"""
    print("🏥 ДИАГНОСТИКА ХРАНИЛИЩА")
    print("=" * 30)
    
    status = get_storage_status()
    
    if not status:
        print("❌ Критическая проблема: нет связи с сервисом хранилища")
        print("\n💡 Рекомендации:")
        print("   1. Проверьте что демон запущен")
        print("   2. Проверьте сетевое соединение")
        print("   3. Проверьте правильность API ключа")
        return
    
    issues = []
    recommendations = []
    
    if not status['is_connected']:
        issues.append("🔴 База данных отключена")
        recommendations.extend([
            "• Проверьте путь к файлу базы данных",
            "• Убедитесь в наличии прав на запись",
            "• Проверьте место на диске"
        ])
    
    if status['is_connected'] and not status['is_healthy']:
        issues.append("🟡 Проблемы с работоспособностью")
        recommendations.extend([
            "• Проверьте логи на ошибки",
            "• Проверьте целостность базы данных",
            "• Проверьте производительность диска"
        ])
    
    # Проверяем время работы
    if status['uptime_seconds'] < 60:
        issues.append("🟡 Система недавно перезапускалась")
        recommendations.append("• Проверьте логи на причины перезапуска")
    
    # Вывод результатов
    if issues:
        print("⚠️ ОБНАРУЖЕННЫЕ ПРОБЛЕМЫ:")
        for issue in issues:
            print(f"   {issue}")
    else:
        print("✅ ПРОБЛЕМ НЕ ОБНАРУЖЕНО")
    
    if recommendations:
        print(f"\n💡 РЕКОМЕНДАЦИИ:")
        for rec in recommendations:
            print(f"   {rec}")
    
    # Общее состояние
    uptime_hours = status['uptime_seconds'] / 3600
    print(f"\n📊 ОБЩАЯ ИНФОРМАЦИЯ:")
    print(f"   Время работы: {format_uptime(timedelta(seconds=status['uptime_seconds']))}")
    print(f"   Стабильность: {'🟢 Высокая' if uptime_hours > 24 else '🟡 Средняя' if uptime_hours > 1 else '🔴 Низкая'}")

if __name__ == "__main__":
    # Простая проверка статуса
    get_storage_status()
    
    print("\n" + "="*50)
    
    # Проверка готовности
    pre_start_check()
    
    print("\n" + "="*50)
    
    # Диагностика
    diagnose_storage_issues()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'storage.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const storageProto = grpc.loadPackageDefinition(packageDefinition).atom.storage.v1;

async function getStorageStatus() {
    const client = new storageProto.StorageService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {};
        
        client.getStorageStatus(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            console.log('🗄️ СТАТУС ХРАНИЛИЩА');
            console.log('='.repeat(30));
            
            // Основная информация
            const connIcon = response.is_connected ? '✅' : '❌';
            const healthIcon = response.is_healthy ? '✅' : '❌';
            
            console.log(`${connIcon} Подключение: ${response.is_connected ? 'подключена' : 'отключена'}`);
            console.log(`${healthIcon} Работоспособность: ${response.is_healthy ? 'здорова' : 'проблемы'}`);
            console.log(`📊 Статус: ${response.status}`);
            
            // Время работы
            const uptime = formatUptime(response.uptime_seconds);
            console.log(`⏱️ Время работы: ${uptime}`);
            
            // Общая оценка
            if (response.is_connected && response.is_healthy) {
                console.log('🟢 Система хранения работает нормально');
            } else if (response.is_connected && !response.is_healthy) {
                console.log('🟡 База подключена, но есть проблемы');
            } else {
                console.log('🔴 Критическая проблема: нет подключения к БД');
            }
            
            resolve({
                isConnected: response.is_connected,
                isHealthy: response.is_healthy,
                status: response.status,
                uptimeSeconds: response.uptime_seconds
            });
        });
    });
}

function formatUptime(seconds) {
    const days = Math.floor(seconds / (24 * 3600));
    const hours = Math.floor((seconds % (24 * 3600)) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
        return `${days}д ${hours}ч ${minutes}м`;
    } else if (hours > 0) {
        return `${hours}ч ${minutes}м`;
    } else {
        return `${minutes}м`;
    }
}

// Класс для мониторинга хранилища
class StorageHealthMonitor {
    constructor() {
        this.client = new storageProto.StorageService('localhost:27500',
            grpc.credentials.createInsecure());
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
        this.alertsSent = new Set();
        this.isMonitoring = false;
        this.monitoringInterval = null;
    }
    
    async checkStatus() {
        return new Promise((resolve, reject) => {
            this.client.getStorageStatus({}, this.metadata, (error, response) => {
                if (error) {
                    resolve({ error: error.message, timestamp: Date.now() });
                } else {
                    resolve({
                        isConnected: response.is_connected,
                        isHealthy: response.is_healthy,
                        status: response.status,
                        uptimeSeconds: response.uptime_seconds,
                        timestamp: Date.now()
                    });
                }
            });
        });
    }
    
    async checkAndAlert() {
        const status = await this.checkStatus();
        
        if (status.error) {
            console.log(`❌ Ошибка мониторинга: ${status.error}`);
            return status;
        }
        
        // Проверяем критические проблемы
        if (!status.isConnected) {
            const alertKey = 'db_disconnected';
            if (!this.alertsSent.has(alertKey)) {
                this.sendAlert('🚨 КРИТИЧНО: База данных отключена!', status);
                this.alertsSent.add(alertKey);
            }
        } else {
            this.alertsSent.delete('db_disconnected');
        }
        
        if (status.isConnected && !status.isHealthy) {
            const alertKey = 'db_unhealthy';
            if (!this.alertsSent.has(alertKey)) {
                this.sendAlert('⚠️ ВНИМАНИЕ: Проблемы с работоспособностью БД', status);
                this.alertsSent.add(alertKey);
            }
        } else {
            this.alertsSent.delete('db_unhealthy');
        }
        
        return status;
    }
    
    sendAlert(message, status) {
        const timestamp = new Date().toLocaleString();
        console.log(`\n[${timestamp}] ${message}`);
        console.log(`   Статус: ${status.status}`);
        console.log(`   Время работы: ${status.uptimeSeconds}с`);
        console.log();
    }
    
    startMonitoring(intervalMs = 30000) {
        if (this.isMonitoring) {
            console.log('⚠️ Мониторинг уже запущен');
            return;
        }
        
        this.isMonitoring = true;
        console.log(`🚀 Запуск мониторинга хранилища каждые ${intervalMs / 1000} секунд`);
        console.log('Время    | Связь | Здоровье | Статус');
        console.log('-'.repeat(40));
        
        const monitor = async () => {
            if (!this.isMonitoring) return;
            
            const status = await this.checkAndAlert();
            
            const currentTime = new Date().toLocaleTimeString();
            
            if (!status.error) {
                const connIcon = status.isConnected ? '✅' : '❌';
                const healthIcon = status.isHealthy ? '✅' : '❌';
                
                console.log(`${currentTime} | ${connIcon}     | ${healthIcon}        | ${status.status}`);
            } else {
                console.log(`${currentTime} | ${'❌'}     | ${'❌'}        | ERROR`);
            }
        };
        
        // Первая проверка сразу
        monitor();
        
        // Запуск периодических проверок
        this.monitoringInterval = setInterval(monitor, intervalMs);
    }
    
    stopMonitoring() {
        if (!this.isMonitoring) return;
        
        this.isMonitoring = false;
        if (this.monitoringInterval) {
            clearInterval(this.monitoringInterval);
            this.monitoringInterval = null;
        }
        console.log('🛑 Мониторинг остановлен');
    }
    
    async waitForReady(timeoutMs = 60000) {
        console.log(`⏳ Ожидание готовности хранилища (таймаут: ${timeoutMs / 1000}с)...`);
        
        const startTime = Date.now();
        
        while (Date.now() - startTime < timeoutMs) {
            const status = await this.checkStatus();
            
            if (!status.error && status.isConnected && status.isHealthy) {
                console.log(`✅ Хранилище готово! Статус: ${status.status}`);
                return true;
            }
            
            if (!status.error) {
                console.log(`⏳ Ожидание... Связь: ${status.isConnected}, Здоровье: ${status.isHealthy}`);
            } else {
                console.log(`⏳ Ожидание... Ошибка: ${status.error}`);
            }
            
            await new Promise(resolve => setTimeout(resolve, 2000));
        }
        
        console.log('❌ Таймаут: хранилище не готово');
        return false;
    }
}

// Функции диагностики
async function diagnoseStorageIssues() {
    console.log('🏥 ДИАГНОСТИКА ХРАНИЛИЩА');
    console.log('='.repeat(30));
    
    try {
        const status = await getStorageStatus();
        
        const issues = [];
        const recommendations = [];
        
        if (!status.isConnected) {
            issues.push('🔴 База данных отключена');
            recommendations.push(
                '• Проверьте путь к файлу базы данных',
                '• Убедитесь в наличии прав на запись',
                '• Проверьте место на диске'
            );
        }
        
        if (status.isConnected && !status.isHealthy) {
            issues.push('🟡 Проблемы с работоспособностью');
            recommendations.push(
                '• Проверьте логи на ошибки',
                '• Проверьте целостность базы данных',
                '• Проверьте производительность диска'
            );
        }
        
        // Проверяем время работы
        if (status.uptimeSeconds < 60) {
            issues.push('🟡 Система недавно перезапускалась');
            recommendations.push('• Проверьте логи на причины перезапуска');
        }
        
        // Вывод результатов
        console.log('\n⚠️ РЕЗУЛЬТАТЫ ДИАГНОСТИКИ:');
        
        if (issues.length > 0) {
            console.log('Обнаруженные проблемы:');
            issues.forEach(issue => console.log(`   ${issue}`));
        } else {
            console.log('✅ Проблем не обнаружено');
        }
        
        if (recommendations.length > 0) {
            console.log('\n💡 Рекомендации:');
            recommendations.forEach(rec => console.log(`   ${rec}`));
        }
        
        // Общее состояние
        const uptimeHours = status.uptimeSeconds / 3600;
        console.log('\n📊 ОБЩАЯ ИНФОРМАЦИЯ:');
        console.log(`   Время работы: ${formatUptime(status.uptimeSeconds)}`);
        
        let stability;
        if (uptimeHours > 24) stability = '🟢 Высокая';
        else if (uptimeHours > 1) stability = '🟡 Средняя';
        else stability = '🔴 Низкая';
        
        console.log(`   Стабильность: ${stability}`);
        
    } catch (error) {
        console.log('❌ Критическая проблема: нет связи с сервисом хранилища');
        console.log('\n💡 Рекомендации:');
        console.log('   1. Проверьте что демон запущен');
        console.log('   2. Проверьте сетевое соединение');
        console.log('   3. Проверьте правильность API ключа');
    }
}

async function preStartCheck() {
    console.log('🔍 ПРОВЕРКА ГОТОВНОСТИ СИСТЕМЫ');
    console.log('='.repeat(40));
    
    try {
        const status = await getStorageStatus();
        
        if (!status.isConnected) {
            console.log('❌ База данных не подключена');
            return false;
        }
        
        if (!status.isHealthy) {
            console.log('⚠️ Проблемы с работоспособностью БД');
            console.log('🔧 Рекомендуется проверить логи и исправить проблемы');
            return false;
        }
        
        console.log('✅ Система хранения готова к работе');
        return true;
        
    } catch (error) {
        console.log('❌ Не удалось подключиться к системе хранения');
        return false;
    }
}

// Демонстрация всех возможностей
async function demonstrateStorageMonitoring() {
    console.log('🚀 Демонстрация мониторинга хранилища\n');
    
    // Простая проверка статуса
    console.log('📊 Текущий статус:');
    try {
        await getStorageStatus();
    } catch (error) {
        console.log(`❌ Ошибка: ${error.message}`);
    }
    
    console.log('\n' + '='.repeat(50));
    
    // Проверка готовности
    await preStartCheck();
    
    console.log('\n' + '='.repeat(50));
    
    // Диагностика
    await diagnoseStorageIssues();
    
    console.log('\n' + '='.repeat(50));
    
    // Демонстрация мониторинга (кратковременного)
    const monitor = new StorageHealthMonitor();
    
    console.log('\n📈 Демонстрация мониторинга (30 секунд):');
    monitor.startMonitoring(5000); // Каждые 5 секунд
    
    setTimeout(() => {
        monitor.stopMonitoring();
        console.log('\n✅ Демонстрация завершена');
    }, 30000);
}

// Основная демонстрация
async function main() {
    try {
        await demonstrateStorageMonitoring();
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Состояния хранилища

### Комбинации статусов
- **✅ Connected + ✅ Healthy**: Нормальная работа
- **✅ Connected + ❌ Unhealthy**: Проблемы с производительностью
- **❌ Disconnected + ❌ Unhealthy**: Критическая ситуация

### Возможные значения status
- **`ready`** - Система готова к работе
- **`connecting`** - Подключение к базе данных  
- **`maintenance`** - Режим обслуживания
- **`degraded`** - Сниженная производительность
- **`error`** - Критические ошибки

## Применение

### Health Checks
```javascript
// Проверка перед запуском приложения
const isReady = await preStartCheck();
if (!isReady) {
    process.exit(1);
}
```

### Мониторинг
```python
# Непрерывный мониторинг в production
monitor = StorageMonitor()
monitor.continuous_monitoring(interval_seconds=60)
```

### DevOps интеграция
```bash
# Использование в скриптах развертывания
atomd storage status
if [ $? -ne 0 ]; then
    echo "Storage not ready"
    exit 1
fi
```

### Container готовность
```go
// Kubernetes readiness probe
if !waitForStorageReady(client, ctx, 30*time.Second) {
    return errors.New("storage not ready")
}
```

## Метрики мониторинга

### Основные показатели
- **Uptime**: Время непрерывной работы
- **Connection Status**: Состояние подключения
- **Health Status**: Общая работоспособность

### Алерты
- **Connection Lost**: Потеря подключения к БД
- **Health Issues**: Проблемы с производительностью
- **Frequent Restarts**: Частые перезапуски

## Связанные методы
- [GetStorageInfo](get-storage-info.md) - Подробная информация о хранилище
