# ThrowError

## Описание
Выбрасывает BPMN ошибку для задания, которая может быть перехвачена boundary event в процессе. Используется для сигнализации о бизнес-ошибках.

## Синтаксис
```protobuf
rpc ThrowError(ThrowErrorRequest) returns (ThrowErrorResponse);
```

## Package
```protobuf
package atom.jobs.v1;
```

## Авторизация
✅ **Требуется API ключ** с разрешением `jobs` или `*`

```go
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "x-api-key", "your-api-key-here")
```

## Параметры запроса

### ThrowErrorRequest
```protobuf
message ThrowErrorRequest {
  string job_key = 1;       // Ключ задания
  string error_code = 2;    // Код BPMN ошибки
  string error_message = 3; // Сообщение об ошибке
  map<string, string> variables = 4; // Переменные для передачи
}
```

#### Поля:
- **job_key** (string, required): Уникальный ключ задания
- **error_code** (string, required): Код BPMN ошибки для корреляции с boundary event
- **error_message** (string, optional): Описание ошибки
- **variables** (map, optional): Дополнительные переменные для процесса

## Параметры ответа

### ThrowErrorResponse
```protobuf
message ThrowErrorResponse {
  bool success = 1;         // Статус успешности операции
  string message = 2;       // Сообщение о результате
  bool error_caught = 3;    // Была ли ошибка перехвачена boundary event
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
    
    pb "atom-engine/proto/jobs/jobspb"
)

func main() {
    conn, err := grpc.Dial("localhost:27500", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewJobsServiceClient(conn)
    ctx := metadata.AppendToOutgoingContext(context.Background(), 
        "x-api-key", "your-api-key-here")
    
    jobKey := "atom-jobkey12345"
    
    // Простое выбрасывание ошибки
    response, err := client.ThrowError(ctx, &pb.ThrowErrorRequest{
        JobKey:       jobKey,
        ErrorCode:    "VALIDATION_ERROR",
        ErrorMessage: "Invalid data format",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if response.Success {
        if response.ErrorCaught {
            fmt.Printf("✅ BPMN ошибка %s выброшена и перехвачена boundary event\n", "VALIDATION_ERROR")
        } else {
            fmt.Printf("⚠️ BPMN ошибка %s выброшена, но не перехвачена\n", "VALIDATION_ERROR")
        }
    } else {
        fmt.Printf("❌ Ошибка выброса: %s\n", response.Message)
    }
}

// Обработчик бизнес-ошибок с различными стратегиями
type BusinessErrorHandler struct {
    client pb.JobsServiceClient
}

func NewBusinessErrorHandler(client pb.JobsServiceClient) *BusinessErrorHandler {
    return &BusinessErrorHandler{client: client}
}

func (h *BusinessErrorHandler) HandleValidationError(ctx context.Context, jobKey string, 
    field string, value interface{}, rule string) error {
    
    variables := map[string]string{
        "validation_field": field,
        "validation_value": fmt.Sprintf("%v", value),
        "validation_rule":  rule,
        "error_timestamp":  time.Now().Format(time.RFC3339),
    }
    
    response, err := h.client.ThrowError(ctx, &pb.ThrowErrorRequest{
        JobKey:       jobKey,
        ErrorCode:    "VALIDATION_ERROR",
        ErrorMessage: fmt.Sprintf("Validation failed for field '%s': %s", field, rule),
        Variables:    variables,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось выбросить ошибку валидации: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка выброса валидации: %s", response.Message)
    }
    
    if response.ErrorCaught {
        fmt.Printf("🔍 Ошибка валидации поля '%s' обработана процессом\n", field)
    } else {
        fmt.Printf("⚠️ Ошибка валидации поля '%s' не перехвачена\n", field)
    }
    
    return nil
}

func (h *BusinessErrorHandler) HandleBusinessRuleViolation(ctx context.Context, jobKey string, 
    ruleName string, details map[string]interface{}) error {
    
    variables := make(map[string]string)
    variables["rule_name"] = ruleName
    variables["error_type"] = "BUSINESS_RULE"
    variables["error_timestamp"] = time.Now().Format(time.RFC3339)
    
    // Конвертируем детали в строки
    for key, value := range details {
        variables[fmt.Sprintf("detail_%s", key)] = fmt.Sprintf("%v", value)
    }
    
    response, err := h.client.ThrowError(ctx, &pb.ThrowErrorRequest{
        JobKey:       jobKey,
        ErrorCode:    "BUSINESS_RULE_VIOLATION",
        ErrorMessage: fmt.Sprintf("Business rule violation: %s", ruleName),
        Variables:    variables,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось выбросить бизнес-ошибку: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка выброса бизнес-правила: %s", response.Message)
    }
    
    if response.ErrorCaught {
        fmt.Printf("📋 Нарушение бизнес-правила '%s' обработано процессом\n", ruleName)
    } else {
        fmt.Printf("⚠️ Нарушение бизнес-правила '%s' не перехвачено\n", ruleName)
    }
    
    return nil
}

func (h *BusinessErrorHandler) HandleExternalServiceError(ctx context.Context, jobKey string, 
    serviceName string, statusCode int, serviceResponse string) error {
    
    variables := map[string]string{
        "service_name":     serviceName,
        "status_code":      fmt.Sprintf("%d", statusCode),
        "service_response": serviceResponse,
        "error_type":       "EXTERNAL_SERVICE",
        "error_timestamp":  time.Now().Format(time.RFC3339),
    }
    
    errorCode := "EXTERNAL_SERVICE_ERROR"
    if statusCode >= 500 {
        errorCode = "EXTERNAL_SERVICE_UNAVAILABLE"
    } else if statusCode == 401 || statusCode == 403 {
        errorCode = "EXTERNAL_SERVICE_AUTH_ERROR"
    } else if statusCode == 404 {
        errorCode = "EXTERNAL_SERVICE_NOT_FOUND"
    }
    
    response, err := h.client.ThrowError(ctx, &pb.ThrowErrorRequest{
        JobKey:       jobKey,
        ErrorCode:    errorCode,
        ErrorMessage: fmt.Sprintf("External service %s returned %d: %s", serviceName, statusCode, serviceResponse),
        Variables:    variables,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось выбросить ошибку внешнего сервиса: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка выброса внешнего сервиса: %s", response.Message)
    }
    
    if response.ErrorCaught {
        fmt.Printf("🔗 Ошибка внешнего сервиса '%s' (%d) обработана процессом\n", serviceName, statusCode)
    } else {
        fmt.Printf("⚠️ Ошибка внешнего сервиса '%s' (%d) не перехвачена\n", serviceName, statusCode)
    }
    
    return nil
}

func (h *BusinessErrorHandler) HandleInsufficientFundsError(ctx context.Context, jobKey string, 
    accountID string, requestedAmount, availableAmount float64) error {
    
    variables := map[string]string{
        "account_id":        accountID,
        "requested_amount":  fmt.Sprintf("%.2f", requestedAmount),
        "available_amount":  fmt.Sprintf("%.2f", availableAmount),
        "shortage_amount":   fmt.Sprintf("%.2f", requestedAmount-availableAmount),
        "error_type":        "INSUFFICIENT_FUNDS",
        "error_timestamp":   time.Now().Format(time.RFC3339),
    }
    
    response, err := h.client.ThrowError(ctx, &pb.ThrowErrorRequest{
        JobKey:       jobKey,
        ErrorCode:    "INSUFFICIENT_FUNDS",
        ErrorMessage: fmt.Sprintf("Insufficient funds in account %s: requested %.2f, available %.2f", 
                                 accountID, requestedAmount, availableAmount),
        Variables:    variables,
    })
    
    if err != nil {
        return fmt.Errorf("не удалось выбросить ошибку недостатка средств: %v", err)
    }
    
    if !response.Success {
        return fmt.Errorf("ошибка выброса недостатка средств: %s", response.Message)
    }
    
    if response.ErrorCaught {
        fmt.Printf("💰 Ошибка недостатка средств для счета '%s' обработана процессом\n", accountID)
    } else {
        fmt.Printf("⚠️ Ошибка недостатка средств для счета '%s' не перехвачена\n", accountID)
    }
    
    return nil
}

// Пример комплексного обработчика заданий с бизнес-ошибками
type PaymentProcessor struct {
    errorHandler *BusinessErrorHandler
    // другие зависимости...
}

func NewPaymentProcessor(client pb.JobsServiceClient) *PaymentProcessor {
    return &PaymentProcessor{
        errorHandler: NewBusinessErrorHandler(client),
    }
}

func (p *PaymentProcessor) ProcessPayment(ctx context.Context, jobKey string, 
    paymentData map[string]interface{}) error {
    
    // Валидация данных платежа
    if err := p.validatePaymentData(paymentData); err != nil {
        return p.errorHandler.HandleValidationError(ctx, jobKey, 
            err.Field, err.Value, err.Rule)
    }
    
    // Проверка лимитов
    amount := paymentData["amount"].(float64)
    if amount > 10000 {
        return p.errorHandler.HandleBusinessRuleViolation(ctx, jobKey,
            "MAX_PAYMENT_LIMIT", map[string]interface{}{
                "amount": amount,
                "limit":  10000,
            })
    }
    
    // Проверка баланса
    accountID := paymentData["account_id"].(string)
    balance, err := p.getAccountBalance(accountID)
    if err != nil {
        return p.errorHandler.HandleExternalServiceError(ctx, jobKey,
            "account-service", 500, err.Error())
    }
    
    if balance < amount {
        return p.errorHandler.HandleInsufficientFundsError(ctx, jobKey,
            accountID, amount, balance)
    }
    
    // Выполнение платежа
    paymentResult, err := p.executePayment(paymentData)
    if err != nil {
        if isExternalServiceError(err) {
            statusCode := extractStatusCode(err)
            return p.errorHandler.HandleExternalServiceError(ctx, jobKey,
                "payment-gateway", statusCode, err.Error())
        }
        
        // Общая ошибка
        return p.errorHandler.HandleBusinessRuleViolation(ctx, jobKey,
            "PAYMENT_PROCESSING_ERROR", map[string]interface{}{
                "error": err.Error(),
            })
    }
    
    fmt.Printf("✅ Платеж успешно обработан: %s\n", paymentResult.TransactionID)
    return nil
}

type ValidationError struct {
    Field string
    Value interface{}
    Rule  string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Rule)
}

func (p *PaymentProcessor) validatePaymentData(data map[string]interface{}) *ValidationError {
    // Проверка обязательных полей
    requiredFields := []string{"amount", "account_id", "currency"}
    for _, field := range requiredFields {
        if _, exists := data[field]; !exists {
            return &ValidationError{
                Field: field,
                Value: nil,
                Rule:  "required field missing",
            }
        }
    }
    
    // Проверка типов
    if amount, ok := data["amount"].(float64); !ok {
        return &ValidationError{
            Field: "amount",
            Value: data["amount"],
            Rule:  "must be a number",
        }
    } else if amount <= 0 {
        return &ValidationError{
            Field: "amount",
            Value: amount,
            Rule:  "must be positive",
        }
    }
    
    // Проверка валюты
    if currency, ok := data["currency"].(string); !ok {
        return &ValidationError{
            Field: "currency",
            Value: data["currency"],
            Rule:  "must be a string",
        }
    } else if !isValidCurrency(currency) {
        return &ValidationError{
            Field: "currency",
            Value: currency,
            Rule:  "must be valid ISO currency code",
        }
    }
    
    return nil
}

func (p *PaymentProcessor) getAccountBalance(accountID string) (float64, error) {
    // Имитация вызова внешнего сервиса
    // В реальном коде здесь будет HTTP/gRPC вызов
    if accountID == "invalid-account" {
        return 0, fmt.Errorf("account not found")
    }
    
    // Имитация баланса
    balances := map[string]float64{
        "acc-123": 5000.00,
        "acc-456": 150.00,
        "acc-789": 25000.00,
    }
    
    balance, exists := balances[accountID]
    if !exists {
        return 0, fmt.Errorf("account %s not found", accountID)
    }
    
    return balance, nil
}

type PaymentResult struct {
    TransactionID string
    Status        string
}

func (p *PaymentProcessor) executePayment(data map[string]interface{}) (*PaymentResult, error) {
    // Имитация обработки платежа
    accountID := data["account_id"].(string)
    
    if accountID == "blocked-account" {
        return nil, &ExternalServiceError{
            StatusCode: 403,
            Message:    "Account is blocked",
        }
    }
    
    if accountID == "gateway-error" {
        return nil, &ExternalServiceError{
            StatusCode: 502,
            Message:    "Payment gateway unavailable",
        }
    }
    
    return &PaymentResult{
        TransactionID: fmt.Sprintf("txn-%d", time.Now().Unix()),
        Status:        "completed",
    }, nil
}

type ExternalServiceError struct {
    StatusCode int
    Message    string
}

func (e *ExternalServiceError) Error() string {
    return e.Message
}

func isExternalServiceError(err error) bool {
    _, ok := err.(*ExternalServiceError)
    return ok
}

func extractStatusCode(err error) int {
    if extErr, ok := err.(*ExternalServiceError); ok {
        return extErr.StatusCode
    }
    return 500
}

func isValidCurrency(currency string) bool {
    validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "RUB"}
    for _, valid := range validCurrencies {
        if currency == valid {
            return true
        }
    }
    return false
}
```

### Python
```python
import grpc
import time
from enum import Enum
from typing import Dict, Any, Optional

import jobs_pb2
import jobs_pb2_grpc

class ErrorType(Enum):
    VALIDATION = "VALIDATION_ERROR"
    BUSINESS_RULE = "BUSINESS_RULE_VIOLATION"
    EXTERNAL_SERVICE = "EXTERNAL_SERVICE_ERROR"
    INSUFFICIENT_FUNDS = "INSUFFICIENT_FUNDS"
    AUTH_ERROR = "AUTHENTICATION_ERROR"

def throw_error(job_key, error_code, error_message, variables=None):
    channel = grpc.insecure_channel('localhost:27500')
    stub = jobs_pb2_grpc.JobsServiceStub(channel)
    metadata = [('x-api-key', 'your-api-key-here')]
    
    request = jobs_pb2.ThrowErrorRequest(
        job_key=job_key,
        error_code=error_code,
        error_message=error_message,
        variables=variables or {}
    )
    
    try:
        response = stub.ThrowError(request, metadata=metadata)
        
        if response.success:
            if response.error_caught:
                print(f"✅ BPMN ошибка {error_code} выброшена и перехвачена boundary event")
            else:
                print(f"⚠️ BPMN ошибка {error_code} выброшена, но не перехвачена")
            return True
        else:
            print(f"❌ Ошибка выброса: {response.message}")
            return False
            
    except grpc.RpcError as e:
        print(f"gRPC Error: {e.code()} - {e.details()}")
        return False

class BusinessErrorHandler:
    def __init__(self):
        self.channel = grpc.insecure_channel('localhost:27500')
        self.stub = jobs_pb2_grpc.JobsServiceStub(self.channel)
        self.metadata = [('x-api-key', 'your-api-key-here')]
    
    def handle_validation_error(self, job_key, field, value, rule):
        """Обработка ошибки валидации"""
        variables = {
            'validation_field': field,
            'validation_value': str(value),
            'validation_rule': rule,
            'error_timestamp': time.strftime('%Y-%m-%dT%H:%M:%SZ'),
        }
        
        return self._throw_error(
            job_key,
            ErrorType.VALIDATION.value,
            f"Validation failed for field '{field}': {rule}",
            variables
        )
    
    def handle_business_rule_violation(self, job_key, rule_name, details):
        """Обработка нарушения бизнес-правила"""
        variables = {
            'rule_name': rule_name,
            'error_type': 'BUSINESS_RULE',
            'error_timestamp': time.strftime('%Y-%m-%dT%H:%M:%SZ'),
        }
        
        # Добавляем детали
        for key, value in details.items():
            variables[f'detail_{key}'] = str(value)
        
        return self._throw_error(
            job_key,
            ErrorType.BUSINESS_RULE.value,
            f"Business rule violation: {rule_name}",
            variables
        )
    
    def handle_external_service_error(self, job_key, service_name, status_code, service_response):
        """Обработка ошибки внешнего сервиса"""
        variables = {
            'service_name': service_name,
            'status_code': str(status_code),
            'service_response': service_response,
            'error_type': 'EXTERNAL_SERVICE',
            'error_timestamp': time.strftime('%Y-%m-%dT%H:%M:%SZ'),
        }
        
        # Определяем тип ошибки по статус коду
        if status_code >= 500:
            error_code = "EXTERNAL_SERVICE_UNAVAILABLE"
        elif status_code in [401, 403]:
            error_code = "EXTERNAL_SERVICE_AUTH_ERROR"
        elif status_code == 404:
            error_code = "EXTERNAL_SERVICE_NOT_FOUND"
        else:
            error_code = "EXTERNAL_SERVICE_ERROR"
        
        return self._throw_error(
            job_key,
            error_code,
            f"External service {service_name} returned {status_code}: {service_response}",
            variables
        )
    
    def handle_insufficient_funds_error(self, job_key, account_id, requested_amount, available_amount):
        """Обработка ошибки недостатка средств"""
        variables = {
            'account_id': account_id,
            'requested_amount': f"{requested_amount:.2f}",
            'available_amount': f"{available_amount:.2f}",
            'shortage_amount': f"{requested_amount - available_amount:.2f}",
            'error_type': 'INSUFFICIENT_FUNDS',
            'error_timestamp': time.strftime('%Y-%m-%dT%H:%M:%SZ'),
        }
        
        return self._throw_error(
            job_key,
            ErrorType.INSUFFICIENT_FUNDS.value,
            f"Insufficient funds in account {account_id}: requested {requested_amount:.2f}, available {available_amount:.2f}",
            variables
        )
    
    def _throw_error(self, job_key, error_code, error_message, variables):
        """Внутренний метод для выброса ошибки"""
        request = jobs_pb2.ThrowErrorRequest(
            job_key=job_key,
            error_code=error_code,
            error_message=error_message,
            variables=variables
        )
        
        try:
            response = self.stub.ThrowError(request, metadata=self.metadata)
            
            if response.success:
                if response.error_caught:
                    print(f"✅ Ошибка {error_code} обработана процессом")
                else:
                    print(f"⚠️ Ошибка {error_code} не перехвачена")
                return True
            else:
                print(f"❌ Не удалось выбросить ошибку: {response.message}")
                return False
                
        except grpc.RpcError as e:
            print(f"gRPC Error при выбросе ошибки: {e.details()}")
            return False

class PaymentProcessor:
    def __init__(self):
        self.error_handler = BusinessErrorHandler()
        # другие зависимости...
    
    def process_payment(self, job_key, payment_data):
        """Обработка платежа с бизнес-ошибками"""
        try:
            # Валидация данных платежа
            validation_error = self.validate_payment_data(payment_data)
            if validation_error:
                return self.error_handler.handle_validation_error(
                    job_key,
                    validation_error['field'],
                    validation_error['value'],
                    validation_error['rule']
                )
            
            # Проверка лимитов
            amount = payment_data['amount']
            if amount > 10000:
                return self.error_handler.handle_business_rule_violation(
                    job_key,
                    "MAX_PAYMENT_LIMIT",
                    {"amount": amount, "limit": 10000}
                )
            
            # Проверка баланса
            account_id = payment_data['account_id']
            try:
                balance = self.get_account_balance(account_id)
            except Exception as e:
                return self.error_handler.handle_external_service_error(
                    job_key, "account-service", 500, str(e)
                )
            
            if balance < amount:
                return self.error_handler.handle_insufficient_funds_error(
                    job_key, account_id, amount, balance
                )
            
            # Выполнение платежа
            try:
                payment_result = self.execute_payment(payment_data)
                print(f"✅ Платеж успешно обработан: {payment_result['transaction_id']}")
                return True
                
            except ExternalServiceError as e:
                return self.error_handler.handle_external_service_error(
                    job_key, "payment-gateway", e.status_code, e.message
                )
            except Exception as e:
                return self.error_handler.handle_business_rule_violation(
                    job_key,
                    "PAYMENT_PROCESSING_ERROR",
                    {"error": str(e)}
                )
                
        except Exception as e:
            print(f"⚠️ Критическая ошибка обработки платежа: {e}")
            return False
    
    def validate_payment_data(self, data):
        """Валидация данных платежа"""
        # Проверка обязательных полей
        required_fields = ['amount', 'account_id', 'currency']
        for field in required_fields:
            if field not in data:
                return {
                    'field': field,
                    'value': None,
                    'rule': 'required field missing'
                }
        
        # Проверка типов и значений
        amount = data.get('amount')
        if not isinstance(amount, (int, float)):
            return {
                'field': 'amount',
                'value': amount,
                'rule': 'must be a number'
            }
        
        if amount <= 0:
            return {
                'field': 'amount',
                'value': amount,
                'rule': 'must be positive'
            }
        
        # Проверка валюты
        currency = data.get('currency')
        if not isinstance(currency, str):
            return {
                'field': 'currency',
                'value': currency,
                'rule': 'must be a string'
            }
        
        if not self.is_valid_currency(currency):
            return {
                'field': 'currency',
                'value': currency,
                'rule': 'must be valid ISO currency code'
            }
        
        return None
    
    def get_account_balance(self, account_id):
        """Получение баланса счета (имитация внешнего сервиса)"""
        if account_id == "invalid-account":
            raise Exception("Account not found")
        
        # Имитация баланса
        balances = {
            "acc-123": 5000.00,
            "acc-456": 150.00,
            "acc-789": 25000.00,
        }
        
        if account_id not in balances:
            raise Exception(f"Account {account_id} not found")
        
        return balances[account_id]
    
    def execute_payment(self, data):
        """Выполнение платежа (имитация)"""
        account_id = data['account_id']
        
        if account_id == "blocked-account":
            raise ExternalServiceError(403, "Account is blocked")
        
        if account_id == "gateway-error":
            raise ExternalServiceError(502, "Payment gateway unavailable")
        
        return {
            'transaction_id': f"txn-{int(time.time())}",
            'status': 'completed'
        }
    
    def is_valid_currency(self, currency):
        """Проверка валидности валюты"""
        valid_currencies = ['USD', 'EUR', 'GBP', 'JPY', 'RUB']
        return currency in valid_currencies

class ExternalServiceError(Exception):
    def __init__(self, status_code, message):
        self.status_code = status_code
        self.message = message
        super().__init__(message)

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 4:
        print("Использование:")
        print("  python throw_error.py <job_key> <error_code> <error_message>")
        print("  python throw_error.py test")
        sys.exit(1)
    
    if sys.argv[1] == "test":
        # Тестирование различных типов бизнес-ошибок
        processor = PaymentProcessor()
        
        test_cases = [
            {
                'name': 'Успешный платеж',
                'data': {
                    'amount': 100.0,
                    'account_id': 'acc-123',
                    'currency': 'USD'
                }
            },
            {
                'name': 'Ошибка валидации - отсутствует amount',
                'data': {
                    'account_id': 'acc-123',
                    'currency': 'USD'
                }
            },
            {
                'name': 'Ошибка валидации - неверная валюта',
                'data': {
                    'amount': 100.0,
                    'account_id': 'acc-123',
                    'currency': 'INVALID'
                }
            },
            {
                'name': 'Превышение лимита',
                'data': {
                    'amount': 15000.0,
                    'account_id': 'acc-789',
                    'currency': 'USD'
                }
            },
            {
                'name': 'Недостаток средств',
                'data': {
                    'amount': 1000.0,
                    'account_id': 'acc-456',
                    'currency': 'USD'
                }
            },
            {
                'name': 'Заблокированный счет',
                'data': {
                    'amount': 100.0,
                    'account_id': 'blocked-account',
                    'currency': 'USD'
                }
            },
        ]
        
        for i, test_case in enumerate(test_cases):
            job_key = f"test-job-{i+1}"
            print(f"\n--- Тест: {test_case['name']} ---")
            processor.process_payment(job_key, test_case['data'])
    else:
        job_key = sys.argv[1]
        error_code = sys.argv[2]
        error_message = sys.argv[3]
        
        throw_error(job_key, error_code, error_message)
```

### JavaScript/Node.js
```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = 'jobs.proto';
const packageDefinition = protoLoader.loadSync(PROTO_PATH);
const jobsProto = grpc.loadPackageDefinition(packageDefinition).atom.jobs.v1;

const ErrorTypes = {
    VALIDATION: 'VALIDATION_ERROR',
    BUSINESS_RULE: 'BUSINESS_RULE_VIOLATION',
    EXTERNAL_SERVICE: 'EXTERNAL_SERVICE_ERROR',
    INSUFFICIENT_FUNDS: 'INSUFFICIENT_FUNDS',
    AUTH_ERROR: 'AUTHENTICATION_ERROR'
};

async function throwError(jobKey, errorCode, errorMessage, variables = {}) {
    const client = new jobsProto.JobsService('localhost:27500',
        grpc.credentials.createInsecure());
    
    const metadata = new grpc.Metadata();
    metadata.add('x-api-key', 'your-api-key-here');
    
    return new Promise((resolve, reject) => {
        const request = {
            job_key: jobKey,
            error_code: errorCode,
            error_message: errorMessage,
            variables: variables
        };
        
        client.throwError(request, metadata, (error, response) => {
            if (error) {
                reject(error);
                return;
            }
            
            if (response.success) {
                if (response.error_caught) {
                    console.log(`✅ BPMN ошибка ${errorCode} выброшена и перехвачена boundary event`);
                } else {
                    console.log(`⚠️ BPMN ошибка ${errorCode} выброшена, но не перехвачена`);
                }
                resolve(true);
            } else {
                console.log(`❌ Ошибка выброса: ${response.message}`);
                resolve(false);
            }
        });
    });
}

class BusinessErrorHandler {
    constructor() {
        this.client = new jobsProto.JobsService('localhost:27500',
            grpc.credentials.createInsecure());
        
        this.metadata = new grpc.Metadata();
        this.metadata.add('x-api-key', 'your-api-key-here');
    }
    
    async handleValidationError(jobKey, field, value, rule) {
        const variables = {
            validation_field: field,
            validation_value: String(value),
            validation_rule: rule,
            error_timestamp: new Date().toISOString(),
        };
        
        return await this._throwError(
            jobKey,
            ErrorTypes.VALIDATION,
            `Validation failed for field '${field}': ${rule}`,
            variables
        );
    }
    
    async handleBusinessRuleViolation(jobKey, ruleName, details) {
        const variables = {
            rule_name: ruleName,
            error_type: 'BUSINESS_RULE',
            error_timestamp: new Date().toISOString(),
        };
        
        // Добавляем детали
        Object.entries(details).forEach(([key, value]) => {
            variables[`detail_${key}`] = String(value);
        });
        
        return await this._throwError(
            jobKey,
            ErrorTypes.BUSINESS_RULE,
            `Business rule violation: ${ruleName}`,
            variables
        );
    }
    
    async handleExternalServiceError(jobKey, serviceName, statusCode, serviceResponse) {
        const variables = {
            service_name: serviceName,
            status_code: String(statusCode),
            service_response: serviceResponse,
            error_type: 'EXTERNAL_SERVICE',
            error_timestamp: new Date().toISOString(),
        };
        
        // Определяем тип ошибки по статус коду
        let errorCode;
        if (statusCode >= 500) {
            errorCode = 'EXTERNAL_SERVICE_UNAVAILABLE';
        } else if ([401, 403].includes(statusCode)) {
            errorCode = 'EXTERNAL_SERVICE_AUTH_ERROR';
        } else if (statusCode === 404) {
            errorCode = 'EXTERNAL_SERVICE_NOT_FOUND';
        } else {
            errorCode = 'EXTERNAL_SERVICE_ERROR';
        }
        
        return await this._throwError(
            jobKey,
            errorCode,
            `External service ${serviceName} returned ${statusCode}: ${serviceResponse}`,
            variables
        );
    }
    
    async handleInsufficientFundsError(jobKey, accountId, requestedAmount, availableAmount) {
        const variables = {
            account_id: accountId,
            requested_amount: requestedAmount.toFixed(2),
            available_amount: availableAmount.toFixed(2),
            shortage_amount: (requestedAmount - availableAmount).toFixed(2),
            error_type: 'INSUFFICIENT_FUNDS',
            error_timestamp: new Date().toISOString(),
        };
        
        return await this._throwError(
            jobKey,
            ErrorTypes.INSUFFICIENT_FUNDS,
            `Insufficient funds in account ${accountId}: requested ${requestedAmount.toFixed(2)}, available ${availableAmount.toFixed(2)}`,
            variables
        );
    }
    
    async _throwError(jobKey, errorCode, errorMessage, variables) {
        return new Promise((resolve, reject) => {
            const request = {
                job_key: jobKey,
                error_code: errorCode,
                error_message: errorMessage,
                variables: variables
            };
            
            this.client.throwError(request, this.metadata, (error, response) => {
                if (error) {
                    console.error(`gRPC Error при выбросе ошибки: ${error.message}`);
                    reject(error);
                    return;
                }
                
                if (response.success) {
                    if (response.error_caught) {
                        console.log(`✅ Ошибка ${errorCode} обработана процессом`);
                    } else {
                        console.log(`⚠️ Ошибка ${errorCode} не перехвачена`);
                    }
                    resolve(true);
                } else {
                    console.log(`❌ Не удалось выбросить ошибку: ${response.message}`);
                    resolve(false);
                }
            });
        });
    }
}

class PaymentProcessor {
    constructor() {
        this.errorHandler = new BusinessErrorHandler();
        // другие зависимости...
    }
    
    async processPayment(jobKey, paymentData) {
        try {
            // Валидация данных платежа
            const validationError = this.validatePaymentData(paymentData);
            if (validationError) {
                return await this.errorHandler.handleValidationError(
                    jobKey,
                    validationError.field,
                    validationError.value,
                    validationError.rule
                );
            }
            
            // Проверка лимитов
            const amount = paymentData.amount;
            if (amount > 10000) {
                return await this.errorHandler.handleBusinessRuleViolation(
                    jobKey,
                    "MAX_PAYMENT_LIMIT",
                    { amount: amount, limit: 10000 }
                );
            }
            
            // Проверка баланса
            const accountId = paymentData.account_id;
            let balance;
            try {
                balance = await this.getAccountBalance(accountId);
            } catch (error) {
                return await this.errorHandler.handleExternalServiceError(
                    jobKey, "account-service", 500, error.message
                );
            }
            
            if (balance < amount) {
                return await this.errorHandler.handleInsufficientFundsError(
                    jobKey, accountId, amount, balance
                );
            }
            
            // Выполнение платежа
            try {
                const paymentResult = await this.executePayment(paymentData);
                console.log(`✅ Платеж успешно обработан: ${paymentResult.transaction_id}`);
                return true;
                
            } catch (error) {
                if (error.statusCode) {
                    return await this.errorHandler.handleExternalServiceError(
                        jobKey, "payment-gateway", error.statusCode, error.message
                    );
                } else {
                    return await this.errorHandler.handleBusinessRuleViolation(
                        jobKey,
                        "PAYMENT_PROCESSING_ERROR",
                        { error: error.message }
                    );
                }
            }
            
        } catch (error) {
            console.log(`⚠️ Критическая ошибка обработки платежа: ${error.message}`);
            return false;
        }
    }
    
    validatePaymentData(data) {
        // Проверка обязательных полей
        const requiredFields = ['amount', 'account_id', 'currency'];
        for (const field of requiredFields) {
            if (!(field in data)) {
                return {
                    field: field,
                    value: null,
                    rule: 'required field missing'
                };
            }
        }
        
        // Проверка типов и значений
        const amount = data.amount;
        if (typeof amount !== 'number') {
            return {
                field: 'amount',
                value: amount,
                rule: 'must be a number'
            };
        }
        
        if (amount <= 0) {
            return {
                field: 'amount',
                value: amount,
                rule: 'must be positive'
            };
        }
        
        // Проверка валюты
        const currency = data.currency;
        if (typeof currency !== 'string') {
            return {
                field: 'currency',
                value: currency,
                rule: 'must be a string'
            };
        }
        
        if (!this.isValidCurrency(currency)) {
            return {
                field: 'currency',
                value: currency,
                rule: 'must be valid ISO currency code'
            };
        }
        
        return null;
    }
    
    async getAccountBalance(accountId) {
        // Имитация вызова внешнего сервиса
        if (accountId === "invalid-account") {
            throw new Error("Account not found");
        }
        
        // Имитация баланса
        const balances = {
            "acc-123": 5000.00,
            "acc-456": 150.00,
            "acc-789": 25000.00,
        };
        
        if (!(accountId in balances)) {
            throw new Error(`Account ${accountId} not found`);
        }
        
        return balances[accountId];
    }
    
    async executePayment(data) {
        // Имитация обработки платежа
        const accountId = data.account_id;
        
        if (accountId === "blocked-account") {
            const error = new Error("Account is blocked");
            error.statusCode = 403;
            throw error;
        }
        
        if (accountId === "gateway-error") {
            const error = new Error("Payment gateway unavailable");
            error.statusCode = 502;
            throw error;
        }
        
        return {
            transaction_id: `txn-${Date.now()}`,
            status: 'completed'
        };
    }
    
    isValidCurrency(currency) {
        const validCurrencies = ['USD', 'EUR', 'GBP', 'JPY', 'RUB'];
        return validCurrencies.includes(currency);
    }
}

// Примеры использования
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.length === 0) {
        console.log('Использование:');
        console.log('  node throw-error.js <job_key> <error_code> <error_message>');
        console.log('  node throw-error.js test');
        process.exit(1);
    }
    
    if (args[0] === 'test') {
        // Тестирование различных типов бизнес-ошибок
        const processor = new PaymentProcessor();
        
        const testCases = [
            {
                name: 'Успешный платеж',
                data: {
                    amount: 100.0,
                    account_id: 'acc-123',
                    currency: 'USD'
                }
            },
            {
                name: 'Ошибка валидации - отсутствует amount',
                data: {
                    account_id: 'acc-123',
                    currency: 'USD'
                }
            },
            {
                name: 'Ошибка валидации - неверная валюта',
                data: {
                    amount: 100.0,
                    account_id: 'acc-123',
                    currency: 'INVALID'
                }
            },
            {
                name: 'Превышение лимита',
                data: {
                    amount: 15000.0,
                    account_id: 'acc-789',
                    currency: 'USD'
                }
            },
            {
                name: 'Недостаток средств',
                data: {
                    amount: 1000.0,
                    account_id: 'acc-456',
                    currency: 'USD'
                }
            },
            {
                name: 'Заблокированный счет',
                data: {
                    amount: 100.0,
                    account_id: 'blocked-account',
                    currency: 'USD'
                }
            },
        ];
        
        (async () => {
            for (let i = 0; i < testCases.length; i++) {
                const testCase = testCases[i];
                const jobKey = `test-job-${i + 1}`;
                console.log(`\n--- Тест: ${testCase.name} ---`);
                await processor.processPayment(jobKey, testCase.data);
            }
        })();
    } else {
        const jobKey = args[0];
        const errorCode = args[1];
        const errorMessage = args[2];
        
        throwError(jobKey, errorCode, errorMessage).catch(error => {
            console.error(`Ошибка: ${error.message}`);
            process.exit(1);
        });
    }
}

module.exports = {
    throwError,
    BusinessErrorHandler,
    PaymentProcessor,
    ErrorTypes
};
```

## BPMN Error Event Integration

### Boundary Error Event
```xml
<bpmn:boundaryEvent id="ValidationErrorBoundary" attachedToRef="ProcessPaymentTask">
  <bpmn:errorEventDefinition errorRef="ValidationError" />
</bpmn:boundaryEvent>

<bpmn:error id="ValidationError" errorCode="VALIDATION_ERROR" />
```

### Error End Event
```xml
<bpmn:endEvent id="InsufficientFundsEnd">
  <bpmn:errorEventDefinition errorRef="InsufficientFundsError" />
</bpmn:endEvent>

<bpmn:error id="InsufficientFundsError" errorCode="INSUFFICIENT_FUNDS" />
```

## Возможные ошибки

### gRPC Status Codes
- `INVALID_ARGUMENT` (3): Неверный job_key или параметры
- `NOT_FOUND` (5): Задание не найдено
- `PERMISSION_DENIED` (7): Недостаточно прав доступа
- `UNAUTHENTICATED` (16): Отсутствует или неверный API ключ

### Примеры ошибок
```json
{
  "success": false,
  "message": "Job 'atom-jobkey12345' not found or already completed",
  "error_caught": false
}
```

## Связанные методы
- [ActivateJobs](activate-jobs.md) - Получение заданий для выполнения
- [CompleteJob](complete-job.md) - Успешное завершение задания
- [FailJob](fail-job.md) - Провал задания с повтором
- [GetJob](get-job.md) - Получение деталей задания
