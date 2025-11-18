# CreateIncident

## Описание
Создает новый инцидент в системе для отслеживания ошибок, сбоев и проблем в выполнении BPMN процессов, заданий, таймеров и других компонентов.

## Синтаксис
```protobuf
rpc CreateIncident(CreateIncidentRequest) returns (CreateIncidentResponse);
```

## Package
```protobuf
package incidents;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `incidents` или `*`

## Параметры запроса

### CreateIncidentRequest
```protobuf
message CreateIncidentRequest {
  IncidentType type = 1;               // Тип инцидента
  string message = 2;                  // Описание проблемы
  string error_code = 3;               // Код ошибки
  string process_instance_id = 4;      // ID экземпляра процесса
  string process_key = 5;              // Ключ процесса
  string element_id = 6;               // ID BPMN элемента
  string element_type = 7;             // Тип BPMN элемента
  string job_key = 8;                  // Ключ задания
  string job_type = 9;                 // Тип задания
  string worker_id = 10;               // ID воркера
  string timer_id = 11;                // ID таймера
  string message_name = 12;            // Имя сообщения
  string correlation_key = 13;         // Ключ корреляции
  int32 original_retries = 14;         // Исходное количество попыток
  map<string, string> metadata = 15;  // Дополнительные метаданные
}
```

## Типы инцидентов

### IncidentType
```protobuf
enum IncidentType {
  INCIDENT_TYPE_UNSPECIFIED = 0;     // Не указан
  INCIDENT_TYPE_JOB_FAILURE = 1;     // Ошибка задания
  INCIDENT_TYPE_BPMN_ERROR = 2;      // Ошибка BPMN
  INCIDENT_TYPE_EXPRESSION_ERROR = 3; // Ошибка выражения
  INCIDENT_TYPE_PROCESS_ERROR = 4;   // Ошибка процесса
  INCIDENT_TYPE_TIMER_ERROR = 5;     // Ошибка таймера
  INCIDENT_TYPE_MESSAGE_ERROR = 6;   // Ошибка сообщения
  INCIDENT_TYPE_SYSTEM_ERROR = 7;    // Системная ошибка
}
```

## Параметры ответа

### CreateIncidentResponse
```protobuf
message CreateIncidentResponse {
  Incident incident = 1;  // Созданный инцидент
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
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    
    pb "atom-engine/proto/incidents/incidentspb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewIncidentsServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    // Создание инцидента для ошибки задания
    response, err := client.CreateIncident(ctx, &pb.CreateIncidentRequest{
        Type:                pb.IncidentType_INCIDENT_TYPE_JOB_FAILURE,
        Message:            "Connection timeout to external service",
        ErrorCode:          "CONN_TIMEOUT_001",
        ProcessInstanceId:  "proc-12345",
        ProcessKey:         "payment-process",
        ElementId:          "service-task-payment",
        ElementType:        "serviceTask",
        JobKey:            "job-67890",
        JobType:           "payment-processor",
        WorkerId:          "worker-001",
        OriginalRetries:   3,
        Metadata: map[string]string{
            "service_url":     "https://payments.example.com/api",
            "timeout_ms":      "5000",
            "response_code":   "500",
            "worker_version":  "1.2.3",
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    incident := response.Incident
    fmt.Printf("✅ Инцидент создан: %s\n", incident.Id)
    fmt.Printf("📊 Тип: %s\n", incident.Type.String())
    fmt.Printf("💬 Сообщение: %s\n", incident.Message)
    fmt.Printf("🔢 Код ошибки: %s\n", incident.ErrorCode)
    fmt.Printf("🔄 Процесс: %s (ID: %s)\n", incident.ProcessKey, incident.ProcessInstanceId)
    fmt.Printf("📅 Создан: %s\n", incident.CreatedAt.AsTime().Format("2006-01-02 15:04:05"))
}

// Создание различных типов инцидентов
func createSampleIncidents(client pb.IncidentsServiceClient, ctx context.Context) {
    fmt.Println("🚨 Создание примеров различных инцидентов...")
    
    incidents := []struct {
        name    string
        request *pb.CreateIncidentRequest
    }{
        {
            name: "Job Failure",
            request: &pb.CreateIncidentRequest{
                Type:               pb.IncidentType_INCIDENT_TYPE_JOB_FAILURE,
                Message:           "Database connection failed",
                ErrorCode:         "DB_CONN_FAILED",
                ProcessInstanceId: "proc-001",
                JobKey:           "job-001",
                JobType:          "database-task",
                WorkerId:         "worker-db-01",
                OriginalRetries:  3,
                Metadata: map[string]string{
                    "database": "postgresql",
                    "host":     "db.example.com",
                },
            },
        },
        {
            name: "BPMN Error", 
            request: &pb.CreateIncidentRequest{
                Type:               pb.IncidentType_INCIDENT_TYPE_BPMN_ERROR,
                Message:           "Invalid gateway condition",
                ErrorCode:         "GATEWAY_CONDITION_ERROR",
                ProcessInstanceId: "proc-002",
                ProcessKey:        "approval-process",
                ElementId:         "gateway-approval",
                ElementType:       "exclusiveGateway",
                Metadata: map[string]string{
                    "condition": "amount > undefined_variable",
                    "line":      "45",
                },
            },
        },
        {
            name: "Expression Error",
            request: &pb.CreateIncidentRequest{
                Type:               pb.IncidentType_INCIDENT_TYPE_EXPRESSION_ERROR,
                Message:           "Division by zero in expression",
                ErrorCode:         "EXPR_DIV_BY_ZERO",
                ProcessInstanceId: "proc-003",
                ElementId:         "script-task-calc",
                Metadata: map[string]string{
                    "expression": "total / count",
                    "variables":  `{"total": 100, "count": 0}`,
                },
            },
        },
        {
            name: "Timer Error",
            request: &pb.CreateIncidentRequest{
                Type:      pb.IncidentType_INCIDENT_TYPE_TIMER_ERROR,
                Message:   "Timer scheduling failed",
                ErrorCode: "TIMER_SCHEDULE_ERROR",
                TimerId:   "timer-boundary-001",
                Metadata: map[string]string{
                    "duration":    "PT30M",
                    "wheel_level": "2",
                    "error_type":  "overflow",
                },
            },
        },
        {
            name: "Message Error",
            request: &pb.CreateIncidentRequest{
                Type:           pb.IncidentType_INCIDENT_TYPE_MESSAGE_ERROR,
                Message:        "Message correlation failed",
                ErrorCode:      "MSG_CORRELATION_FAILED",
                MessageName:    "payment-confirmed",
                CorrelationKey: "order-12345",
                Metadata: map[string]string{
                    "subscriptions": "3",
                    "ttl_expired":   "true",
                },
            },
        },
    }
    
    for _, incident := range incidents {
        fmt.Printf("\n📋 Создание инцидента: %s\n", incident.name)
        
        response, err := client.CreateIncident(ctx, incident.request)
        if err != nil {
            fmt.Printf("❌ Ошибка: %v\n", err)
            continue
        }
        
        created := response.Incident
        fmt.Printf("   ✅ ID: %s\n", created.Id)
        fmt.Printf("   📊 Статус: %s\n", created.Status.String())
        fmt.Printf("   💬 Сообщение: %s\n", created.Message)
        
        if len(created.Metadata) > 0 {
            fmt.Printf("   🏷️ Метаданные:\n")
            for key, value := range created.Metadata {
                fmt.Printf("      %s: %s\n", key, value)
            }
        }
    }
}
```

### Python
```python
import grpc
from datetime import datetime

import incidents_pb2
import incidents_pb2_grpc

def create_incident(incident_type, message, error_code, **kwargs):
    channel = grpc.insecure_channel('localhost:27500')
    stub = incidents_pb2_grpc.IncidentsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = incidents_pb2.CreateIncidentRequest(
        type=incident_type,
        message=message,
        error_code=error_code,
        **kwargs
    )
    
    try:
        response = stub.CreateIncident(request, metadata=metadata)
        
        incident = response.incident
        print(f"✅ Инцидент создан: {incident.id}")
        print(f"📊 Тип: {incidents_pb2.IncidentType.Name(incident.type)}")
        print(f"💬 Сообщение: {incident.message}")
        print(f"🔢 Код ошибки: {incident.error_code}")
        print(f"📊 Статус: {incidents_pb2.IncidentStatus.Name(incident.status)}")
        
        created_time = incident.created_at.ToDatetime()
        print(f"📅 Создан: {created_time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        if incident.process_instance_id:
            print(f"🔄 Процесс: {incident.process_instance_id}")
        
        if incident.job_key:
            print(f"🔧 Задание: {incident.job_key} (тип: {incident.job_type})")
            
        if incident.metadata:
            print("🏷️ Метаданные:")
            for key, value in incident.metadata.items():
                print(f"   {key}: {value}")
        
        return incident.id
        
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return None

# Класс для создания типизированных инцидентов
class IncidentCreator:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = incidents_pb2_grpc.IncidentsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def create_job_failure(self, job_key, job_type, worker_id, error_message, 
                          process_instance_id=None, retries=0, metadata=None):
        """Создает инцидент для ошибки задания"""
        request = incidents_pb2.CreateIncidentRequest(
            type=incidents_pb2.IncidentType.INCIDENT_TYPE_JOB_FAILURE,
            message=error_message,
            error_code="JOB_EXECUTION_FAILED",
            job_key=job_key,
            job_type=job_type,
            worker_id=worker_id,
            process_instance_id=process_instance_id or "",
            original_retries=retries,
            metadata=metadata or {}
        )
        
        try:
            response = self.stub.CreateIncident(request, metadata=self.metadata)
            print(f"🚨 Инцидент задания создан: {response.incident.id}")
            return response.incident.id
        except grpc.RpcError as e:
            print(f"❌ Ошибка создания инцидента задания: {e.details()}")
            return None
    
    def create_bpmn_error(self, process_instance_id, element_id, error_message,
                         process_key=None, element_type=None, metadata=None):
        """Создает инцидент для BPMN ошибки"""
        request = incidents_pb2.CreateIncidentRequest(
            type=incidents_pb2.IncidentType.INCIDENT_TYPE_BPMN_ERROR,
            message=error_message,
            error_code="BPMN_EXECUTION_ERROR", 
            process_instance_id=process_instance_id,
            process_key=process_key or "",
            element_id=element_id,
            element_type=element_type or "",
            metadata=metadata or {}
        )
        
        try:
            response = self.stub.CreateIncident(request, metadata=self.metadata)
            print(f"🔄 Инцидент BPMN создан: {response.incident.id}")
            return response.incident.id
        except grpc.RpcError as e:
            print(f"❌ Ошибка создания BPMN инцидента: {e.details()}")
            return None
    
    def create_expression_error(self, expression, variables, error_message,
                              process_instance_id=None, element_id=None):
        """Создает инцидент для ошибки выражения"""
        metadata = {
            "expression": expression,
            "variables": str(variables),
            "evaluation_engine": "FEEL"
        }
        
        request = incidents_pb2.CreateIncidentRequest(
            type=incidents_pb2.IncidentType.INCIDENT_TYPE_EXPRESSION_ERROR,
            message=error_message,
            error_code="EXPRESSION_EVALUATION_ERROR",
            process_instance_id=process_instance_id or "",
            element_id=element_id or "",
            metadata=metadata
        )
        
        try:
            response = self.stub.CreateIncident(request, metadata=self.metadata)
            print(f"🧮 Инцидент выражения создан: {response.incident.id}")
            return response.incident.id
        except grpc.RpcError as e:
            print(f"❌ Ошибка создания инцидента выражения: {e.details()}")
            return None
    
    def create_timer_error(self, timer_id, error_message, metadata=None):
        """Создает инцидент для ошибки таймера"""
        request = incidents_pb2.CreateIncidentRequest(
            type=incidents_pb2.IncidentType.INCIDENT_TYPE_TIMER_ERROR,
            message=error_message,
            error_code="TIMER_ERROR",
            timer_id=timer_id,
            metadata=metadata or {}
        )
        
        try:
            response = self.stub.CreateIncident(request, metadata=self.metadata)
            print(f"⏰ Инцидент таймера создан: {response.incident.id}")
            return response.incident.id
        except grpc.RpcError as e:
            print(f"❌ Ошибка создания инцидента таймера: {e.details()}")
            return None

# Демонстрация создания инцидентов
def demonstrate_incident_creation():
    print("🚨 Демонстрация создания различных инцидентов\n")
    
    creator = IncidentCreator()
    
    # 1. Инцидент задания
    print("1. 🔧 Создание инцидента задания:")
    job_incident_id = creator.create_job_failure(
        job_key="payment-job-123",
        job_type="payment-processor",
        worker_id="payment-worker-01",
        error_message="Payment gateway timeout after 30 seconds",
        process_instance_id="proc-payment-456",
        retries=2,
        metadata={
            "gateway": "stripe",
            "amount": "99.99",
            "currency": "USD",
            "timeout_ms": "30000"
        }
    )
    
    print(f"\n2. 🔄 Создание BPMN инцидента:")
    bpmn_incident_id = creator.create_bpmn_error(
        process_instance_id="proc-approval-789",
        process_key="document-approval",
        element_id="gateway-decision",
        element_type="exclusiveGateway",
        error_message="Gateway condition evaluation failed: variable 'approver' not found",
        metadata={
            "condition": "approver.role = 'manager'",
            "available_variables": "document, user, timestamp"
        }
    )
    
    print(f"\n3. 🧮 Создание инцидента выражения:")
    expr_incident_id = creator.create_expression_error(
        expression="total / count * 100",
        variables={"total": 150, "count": 0},
        error_message="Division by zero in percentage calculation",
        process_instance_id="proc-report-321",
        element_id="script-task-calculate"
    )
    
    print(f"\n4. ⏰ Создание инцидента таймера:")
    timer_incident_id = creator.create_timer_error(
        timer_id="boundary-timer-001",
        error_message="Timer wheel overflow - duration too large",
        metadata={
            "duration": "P999Y",
            "wheel_level": "4",
            "max_supported": "P100Y"
        }
    )
    
    # Сводка созданных инцидентов
    created_incidents = [
        ("Job Failure", job_incident_id),
        ("BPMN Error", bpmn_incident_id), 
        ("Expression Error", expr_incident_id),
        ("Timer Error", timer_incident_id)
    ]
    
    print(f"\n📋 СВОДКА СОЗДАННЫХ ИНЦИДЕНТОВ:")
    print("=" * 40)
    for incident_type, incident_id in created_incidents:
        status = "✅" if incident_id else "❌"
        print(f"{status} {incident_type}: {incident_id or 'Не создан'}")

# Автоматическое создание инцидентов на основе исключений
class AutoIncidentReporter:
    def __init__(self):
        self.creator = IncidentCreator()
    
    def report_job_exception(self, job_context, exception):
        """Автоматически создает инцидент для исключения задания"""
        error_metadata = {
            "exception_type": type(exception).__name__,
            "stack_trace": str(exception),
            "job_duration_ms": str(job_context.get('duration_ms', 0)),
            "worker_version": job_context.get('worker_version', 'unknown')
        }
        
        return self.creator.create_job_failure(
            job_key=job_context['job_key'],
            job_type=job_context['job_type'],
            worker_id=job_context['worker_id'],
            error_message=f"Job failed with {type(exception).__name__}: {str(exception)}",
            process_instance_id=job_context.get('process_instance_id'),
            retries=job_context.get('retries', 0),
            metadata=error_metadata
        )
    
    def report_process_exception(self, process_context, exception):
        """Автоматически создает инцидент для исключения процесса"""
        error_metadata = {
            "exception_type": type(exception).__name__,
            "stack_trace": str(exception),
            "process_version": str(process_context.get('version', 1)),
            "tenant_id": process_context.get('tenant_id', 'default')
        }
        
        return self.creator.create_bpmn_error(
            process_instance_id=process_context['process_instance_id'],
            process_key=process_context.get('process_key', ''),
            element_id=process_context.get('element_id', ''),
            element_type=process_context.get('element_type', ''),
            error_message=f"Process execution failed: {str(exception)}",
            metadata=error_metadata
        )

if __name__ == "__main__":
    # Простое создание инцидента
    create_incident(
        incident_type=incidents_pb2.IncidentType.INCIDENT_TYPE_JOB_FAILURE,
        message="Service unavailable",
        error_code="SVC_UNAVAILABLE_503",
        job_key="test-job-1",
        job_type="http-request",
        worker_id="worker-http-01",
        metadata={
            "url": "https://api.example.com/data",
            "response_code": "503"
        }
    )
    
    print("\n" + "="*60)
    
    # Демонстрация различных типов
    demonstrate_incident_creation()
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'incidents.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const incidentsProto = grpc.loadPackageDefinition(packageDefinition).incidents;

async function createIncident(incidentData) {
    const client = new incidentsProto.IncidentsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        client.createIncident(incidentData, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            const incident = response.incident;
            console.log(`✅ Инцидент создан: ${incident.id}`);
            console.log(`📊 Тип: ${getIncidentTypeName(incident.type)}`);
            console.log(`💬 Сообщение: ${incident.message}`);
            console.log(`🔢 Код ошибки: ${incident.error_code}`);
            console.log(`📊 Статус: ${getIncidentStatusName(incident.status)}`);
            
            const createdTime = new Date(incident.created_at.seconds * 1000);
            console.log(`📅 Создан: ${createdTime.toLocaleString()}`);
            
            if (incident.process_instance_id) {
                console.log(`🔄 Процесс: ${incident.process_instance_id}`);
            }
            
            if (incident.job_key) {
                console.log(`🔧 Задание: ${incident.job_key} (тип: ${incident.job_type})`);
            }
            
            if (Object.keys(incident.metadata).length > 0) {
                console.log('🏷️ Метаданные:');
                Object.entries(incident.metadata).forEach(([key, value]) => {
                    console.log(`   ${key}: ${value}`);
                });
            }
            
            resolve(incident.id);
        });
    });
}

// Утилиты для конвертации enum значений в строки
function getIncidentTypeName(type) {
    const types = {
        0: 'UNSPECIFIED',
        1: 'JOB_FAILURE',
        2: 'BPMN_ERROR', 
        3: 'EXPRESSION_ERROR',
        4: 'PROCESS_ERROR',
        5: 'TIMER_ERROR',
        6: 'MESSAGE_ERROR',
        7: 'SYSTEM_ERROR'
    };
    return types[type] || 'UNKNOWN';
}

function getIncidentStatusName(status) {
    const statuses = {
        0: 'UNSPECIFIED',
        1: 'OPEN',
        2: 'RESOLVED',
        3: 'DISMISSED'
    };
    return statuses[status] || 'UNKNOWN';
}

// Класс для управления инцидентами
class IncidentManager {
    constructor() {
        this.client = new incidentsProto.IncidentsService('localhost:27500',
            grpc.credentials.createInsecure());
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async createJobFailureIncident(jobContext, error) {
        const incidentData = {
            type: 1, // INCIDENT_TYPE_JOB_FAILURE
            message: `Job ${jobContext.jobType} failed: ${error.message}`,
            error_code: error.code || 'JOB_EXECUTION_ERROR',
            job_key: jobContext.jobKey,
            job_type: jobContext.jobType,
            worker_id: jobContext.workerId,
            process_instance_id: jobContext.processInstanceId || '',
            original_retries: jobContext.retries || 0,
            metadata: {
                error_name: error.name,
                stack_trace: error.stack || '',
                job_duration_ms: (jobContext.durationMs || 0).toString(),
                worker_version: jobContext.workerVersion || 'unknown',
                ...jobContext.customMetadata
            }
        };
        
        try {
            const incidentId = await createIncident(incidentData);
            console.log(`🚨 Инцидент задания создан: ${incidentId}`);
            return incidentId;
        } catch (error) {
            console.error(`❌ Ошибка создания инцидента: ${error.message}`);
            return null;
        }
    }
    
    async createBPMNErrorIncident(processContext, error) {
        const incidentData = {
            type: 2, // INCIDENT_TYPE_BPMN_ERROR
            message: `BPMN execution error in ${processContext.elementType}: ${error.message}`,
            error_code: error.code || 'BPMN_EXECUTION_ERROR',
            process_instance_id: processContext.processInstanceId,
            process_key: processContext.processKey || '',
            element_id: processContext.elementId,
            element_type: processContext.elementType,
            metadata: {
                error_name: error.name,
                process_version: (processContext.version || 1).toString(),
                tenant_id: processContext.tenantId || 'default',
                execution_path: processContext.executionPath || '',
                ...processContext.customMetadata
            }
        };
        
        try {
            const incidentId = await createIncident(incidentData);
            console.log(`🔄 BPMN инцидент создан: ${incidentId}`);
            return incidentId;
        } catch (error) {
            console.error(`❌ Ошибка создания BPMN инцидента: ${error.message}`);
            return null;
        }
    }
    
    async createExpressionErrorIncident(expressionContext, error) {
        const incidentData = {
            type: 3, // INCIDENT_TYPE_EXPRESSION_ERROR
            message: `Expression evaluation failed: ${error.message}`,
            error_code: 'EXPRESSION_EVALUATION_ERROR',
            process_instance_id: expressionContext.processInstanceId || '',
            element_id: expressionContext.elementId || '',
            metadata: {
                expression: expressionContext.expression,
                variables: JSON.stringify(expressionContext.variables || {}),
                evaluation_engine: 'FEEL',
                error_position: (expressionContext.errorPosition || 0).toString(),
                ...expressionContext.customMetadata
            }
        };
        
        try {
            const incidentId = await createIncident(incidentData);
            console.log(`🧮 Инцидент выражения создан: ${incidentId}`);
            return incidentId;
        } catch (error) {
            console.error(`❌ Ошибка создания инцидента выражения: ${error.message}`);
            return null;
        }
    }
    
    async createTimerErrorIncident(timerContext, error) {
        const incidentData = {
            type: 5, // INCIDENT_TYPE_TIMER_ERROR
            message: `Timer error: ${error.message}`,
            error_code: 'TIMER_ERROR',
            timer_id: timerContext.timerId,
            metadata: {
                duration: timerContext.duration || '',
                wheel_level: (timerContext.wheelLevel || 0).toString(),
                timer_type: timerContext.timerType || '',
                error_type: error.name,
                ...timerContext.customMetadata
            }
        };
        
        try {
            const incidentId = await createIncident(incidentData);
            console.log(`⏰ Инцидент таймера создан: ${incidentId}`);
            return incidentId;
        } catch (error) {
            console.error(`❌ Ошибка создания инцидента таймера: ${error.message}`);
            return null;
        }
    }
    
    async createSystemErrorIncident(systemContext, error) {
        const incidentData = {
            type: 7, // INCIDENT_TYPE_SYSTEM_ERROR
            message: `System error: ${error.message}`,
            error_code: error.code || 'SYSTEM_ERROR',
            metadata: {
                component: systemContext.component || 'unknown',
                version: systemContext.version || '1.0.0',
                environment: systemContext.environment || 'production',
                error_name: error.name,
                stack_trace: error.stack || '',
                memory_usage: (systemContext.memoryUsage || 0).toString(),
                ...systemContext.customMetadata
            }
        };
        
        try {
            const incidentId = await createIncident(incidentData);
            console.log(`⚙️ Системный инцидент создан: ${incidentId}`);
            return incidentId;
        } catch (error) {
            console.error(`❌ Ошибка создания системного инцидента: ${error.message}`);
            return null;
        }
    }
}

// Декоратор для автоматического создания инцидентов
function withIncidentReporting(incidentManager, context) {
    return function(target, propertyName, descriptor) {
        const method = descriptor.value;
        
        descriptor.value = async function(...args) {
            try {
                return await method.apply(this, args);
            } catch (error) {
                // Автоматически создаем инцидент при ошибке
                await incidentManager.createJobFailureIncident(context, error);
                throw error; // Пере-выбрасываем ошибку
            }
        };
        
        return descriptor;
    };
}

// Демонстрация создания различных инцидентов
async function demonstrateIncidentCreation() {
    console.log('🚨 Демонстрация создания инцидентов\n');
    
    const manager = new IncidentManager();
    
    try {
        // 1. Инцидент задания
        console.log('1. 🔧 Создание инцидента задания:');
        await manager.createJobFailureIncident(
            {
                jobKey: 'email-job-456',
                jobType: 'send-email',
                workerId: 'email-worker-02',
                processInstanceId: 'proc-notification-789',
                retries: 1,
                durationMs: 15000,
                workerVersion: '2.1.0',
                customMetadata: {
                    recipient: 'user@example.com',
                    template: 'welcome-email',
                    smtp_server: 'mail.example.com'
                }
            },
            {
                name: 'SmtpError',
                message: 'SMTP server connection timeout',
                code: 'SMTP_TIMEOUT',
                stack: 'Error: SMTP timeout\n    at SmtpClient.connect(...)'
            }
        );
        
        // 2. BPMN инцидент
        console.log('\n2. 🔄 Создание BPMN инцидента:');
        await manager.createBPMNErrorIncident(
            {
                processInstanceId: 'proc-order-321',
                processKey: 'order-fulfillment',
                elementId: 'gateway-payment',
                elementType: 'exclusiveGateway',
                version: 2,
                tenantId: 'tenant-retail',
                executionPath: 'start -> gateway-payment',
                customMetadata: {
                    order_id: 'ORD-12345',
                    payment_method: 'credit_card'
                }
            },
            {
                name: 'GatewayConditionError',
                message: 'Variable paymentStatus is undefined',
                code: 'GATEWAY_CONDITION_ERROR'
            }
        );
        
        // 3. Инцидент выражения
        console.log('\n3. 🧮 Создание инцидента выражения:');
        await manager.createExpressionErrorIncident(
            {
                processInstanceId: 'proc-calc-654',
                elementId: 'script-task-discount',
                expression: 'basePrice * (1 - discountRate)',
                variables: { basePrice: 100, discountRate: null },
                errorPosition: 15,
                customMetadata: {
                    customer_tier: 'premium',
                    promotion_id: 'SUMMER2023'
                }
            },
            {
                name: 'ExpressionError',
                message: 'Cannot multiply by null value'
            }
        );
        
        // 4. Инцидент таймера
        console.log('\n4. ⏰ Создание инцидента таймера:');
        await manager.createTimerErrorIncident(
            {
                timerId: 'reminder-timer-001',
                duration: 'PT2H',
                wheelLevel: 2,
                timerType: 'boundary_event',
                customMetadata: {
                    process_element: 'user-task-approval',
                    timeout_reason: 'user_inactivity'
                }
            },
            {
                name: 'TimerSchedulingError',
                message: 'Timer wheel capacity exceeded'
            }
        );
        
        // 5. Системный инцидент
        console.log('\n5. ⚙️ Создание системного инцидента:');
        await manager.createSystemErrorIncident(
            {
                component: 'process-engine',
                version: '3.2.1',
                environment: 'production',
                memoryUsage: 85,
                customMetadata: {
                    active_processes: '1250',
                    worker_nodes: '5',
                    database_connections: '15'
                }
            },
            {
                name: 'OutOfMemoryError',
                message: 'JVM heap space exceeded',
                code: 'MEMORY_EXHAUSTED',
                stack: 'java.lang.OutOfMemoryError: Java heap space\n    at ProcessEngine.execute(...)'
            }
        );
        
        console.log('\n✅ Демонстрация создания инцидентов завершена');
        
    } catch (error) {
        console.error(`❌ Ошибка в демонстрации: ${error.message}`);
    }
}

// Пример использования с интеграцией в обработчик ошибок
class ErrorHandler {
    constructor() {
        this.incidentManager = new IncidentManager();
    }
    
    async handleJobError(jobContext, error) {
        console.log(`🚨 Обработка ошибки задания: ${jobContext.jobKey}`);
        
        // Создаем инцидент
        const incidentId = await this.incidentManager.createJobFailureIncident(jobContext, error);
        
        // Дополнительная логика обработки
        if (incidentId) {
            console.log(`📊 Инцидент зарегистрирован, продолжаем обработку...`);
            
            // Можем добавить уведомления, логирование и т.д.
            await this.notifyOperations(incidentId, 'JOB_FAILURE', jobContext);
        }
        
        return incidentId;
    }
    
    async notifyOperations(incidentId, type, context) {
        console.log(`📧 Уведомление операционной команде об инциденте ${incidentId}`);
        // Здесь может быть отправка email, Slack уведомления и т.д.
    }
}

// Основная демонстрация
async function main() {
    try {
        // Простое создание инцидента
        console.log('📋 Простое создание инцидента:');
        await createIncident({
            type: 1, // JOB_FAILURE
            message: 'Database connection timeout',
            error_code: 'DB_TIMEOUT_001',
            job_key: 'db-sync-job',
            job_type: 'database-sync',
            worker_id: 'sync-worker-01',
            original_retries: 3,
            metadata: {
                database: 'user_data',
                timeout_ms: '30000',
                connection_pool: 'primary'
            }
        });
        
        console.log('\n' + '='.repeat(60));
        
        // Комплексная демонстрация
        await demonstrateIncidentCreation();
        
    } catch (error) {
        console.error('❌ Ошибка:', error.message);
    }
}

main();
```

## Контекстная информация

### Process Context
- **process_instance_id**: ID экземпляра процесса
- **process_key**: Ключ определения процесса
- **element_id**: ID BPMN элемента
- **element_type**: Тип элемента (serviceTask, gateway, etc.)

### Job Context  
- **job_key**: Уникальный ключ задания
- **job_type**: Тип задания
- **worker_id**: ID воркера
- **original_retries**: Количество попыток

### Timer Context
- **timer_id**: ID таймера
- Контекст процесса если таймер связан с BPMN

### Message Context
- **message_name**: Имя сообщения
- **correlation_key**: Ключ корреляции

## Применение

### Автоматическое создание
```javascript
// В обработчиках ошибок
try {
    await executeJob(job);
} catch (error) {
    await createJobFailureIncident(job, error);
    throw error;
}
```

### Интеграция с логированием
```python
# Совмещение с системой логирования
import logging

def handle_error(error, context):
    logging.error(f"Error: {error}", extra=context)
    create_incident(error_type, str(error), context)
```

### Мониторинг системы
```go
// Создание инцидентов для системных метрик
if memoryUsage > criticalThreshold {
    createSystemIncident("High memory usage", memoryData)
}
```

## Связанные методы
- [ResolveIncident](resolve-incident.md) - Решение созданных инцидентов
- [ListIncidents](list-incidents.md) - Просмотр всех инцидентов
- [GetIncident](get-incident.md) - Детали конкретного инцидента
