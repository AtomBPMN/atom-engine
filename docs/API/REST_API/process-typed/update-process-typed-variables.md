# PUT /api/v1/processes/:id/typed/variables

## Описание
Обновление типизированных переменных процесса с валидацией по JSON Schema и автоматическим преобразованием типов.

## URL
```
PUT /api/v1/processes/:id/typed/variables
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры тела запроса

### Обязательные поля
- `variables` (object): Объект с переменными для обновления

### Опциональные поля
- `validate` (boolean): Выполнить валидацию перед обновлением (по умолчанию: true)
- `transform` (boolean): Автоматически преобразовать типы (по умолчанию: true)
- `partial_update` (boolean): Частичное обновление объектов (по умолчанию: true)
- `update_reason` (string): Причина обновления для аудита

## Примеры запросов

### Обновление нескольких переменных
```bash
curl -X PUT "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "variables": {
      "paymentStatus": "COMPLETED",
      "totalAmount": 1150.99,
      "customer": {
        "type": "VIP",
        "loyaltyPoints": 1250
      }
    },
    "update_reason": "Payment processed successfully"
  }'
```

### Частичное обновление объекта
```bash
curl -X PUT "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key-here" \
  -d '{
    "variables": {
      "customer.address.city": "New York",
      "customer.address.zipCode": "10001"
    },
    "partial_update": true,
    "update_reason": "Customer updated shipping address"
  }'
```

### JavaScript
```javascript
const updateRequest = {
  variables: {
    orderItems: [
      {
        productId: "PROD-001",
        name: "Laptop Computer",
        quantity: 1,
        unitPrice: 999.99,
        totalPrice: 999.99
      },
      {
        productId: "PROD-003",
        name: "Monitor",
        quantity: 1,
        unitPrice: 299.99,
        totalPrice: 299.99
      }
    ],
    totalAmount: 1299.98,
    paymentStatus: "PENDING"
  },
  validate: true,
  transform: true,
  update_reason: "Customer added monitor to order"
};

const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your-api-key-here'
  },
  body: JSON.stringify(updateRequest)
});

const result = await response.json();
```

## Ответы

### 200 OK - Переменные обновлены
```json
{
  "success": true,
  "data": {
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "updated_at": "2025-01-11T11:15:30.456Z",
    "updated_variables": {
      "paymentStatus": {
        "old_value": "PROCESSING",
        "new_value": "COMPLETED",
        "type": "string",
        "validation_passed": true,
        "transformed": false
      },
      "totalAmount": {
        "old_value": 1050.99,
        "new_value": 1150.99,
        "type": "number",
        "validation_passed": true,
        "transformed": false
      },
      "customer": {
        "old_value": {
          "id": "CUST-789456",
          "name": "John Doe",
          "email": "john.doe@example.com",
          "type": "PREMIUM"
        },
        "new_value": {
          "id": "CUST-789456",
          "name": "John Doe",
          "email": "john.doe@example.com",
          "type": "VIP",
          "loyaltyPoints": 1250
        },
        "type": "object",
        "validation_passed": true,
        "transformed": false,
        "partial_update": true,
        "fields_updated": ["type", "loyaltyPoints"]
      }
    },
    "validation_results": {
      "all_valid": true,
      "validation_count": 3,
      "errors": [],
      "warnings": [
        {
          "field": "customer.loyaltyPoints",
          "message": "New field added - not in original schema",
          "severity": "INFO"
        }
      ]
    },
    "transformations_applied": [],
    "business_rules_triggered": [
      {
        "rule": "VIP_CUSTOMER_BENEFITS",
        "description": "Customer upgraded to VIP status",
        "actions": ["apply_vip_discount", "send_welcome_email"]
      }
    ],
    "audit_info": {
      "updated_by": "api_user",
      "update_reason": "Payment processed successfully",
      "change_log_id": "srv1-change789abc123def456",
      "affected_variables_count": 3
    }
  },
  "request_id": "req_1641998405200"
}
```

### 400 Bad Request - Валидационные ошибки
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "One or more variables failed validation",
    "details": {
      "validation_errors": [
        {
          "field": "paymentStatus",
          "value": "INVALID_STATUS",
          "errors": [
            "Value must be one of: PENDING, PROCESSING, COMPLETED, FAILED, REFUNDED"
          ],
          "schema_rule": "enum"
        },
        {
          "field": "totalAmount",
          "value": -100,
          "errors": [
            "Minimum value is 0"
          ],
          "schema_rule": "minimum"
        },
        {
          "field": "customer.email",
          "value": "invalid-email",
          "errors": [
            "Invalid email format"
          ],
          "schema_rule": "format"
        }
      ],
      "variables_processed": 3,
      "variables_valid": 0,
      "can_retry": true,
      "suggestions": [
        "Check enum values for paymentStatus",
        "Ensure totalAmount is positive",
        "Provide valid email format"
      ]
    }
  },
  "request_id": "req_1641998405201"
}
```

### 409 Conflict - Переменная только для чтения
```json
{
  "success": false,
  "error": {
    "code": "READONLY_VARIABLE",
    "message": "Attempt to update readonly variable",
    "details": {
      "readonly_variables": ["orderId", "createdAt"],
      "attempted_updates": {
        "orderId": "ORD-2025-999999"
      },
      "allowed_variables": ["customer", "orderItems", "totalAmount", "paymentStatus"]
    }
  },
  "request_id": "req_1641998405202"
}
```

## Использование

### Typed Variables Updater
```javascript
class TypedVariablesUpdater {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async updateVariables(processInstanceId, variables, options = {}) {
    const requestBody = {
      variables,
      validate: options.validate !== false,
      transform: options.transform !== false,
      partial_update: options.partialUpdate !== false,
      update_reason: options.reason
    };
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/variables`,
      {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey
        },
        body: JSON.stringify(requestBody)
      }
    );
    
    const result = await response.json();
    
    if (!response.ok) {
      throw new Error(`Update failed: ${result.error?.message || response.statusText}`);
    }
    
    return result;
  }
  
  async updateSingleVariable(processInstanceId, variableName, value, reason) {
    return await this.updateVariables(processInstanceId, {
      [variableName]: value
    }, { reason });
  }
  
  async updateObjectProperty(processInstanceId, objectName, propertyPath, value, reason) {
    const variables = {};
    variables[`${objectName}.${propertyPath}`] = value;
    
    return await this.updateVariables(processInstanceId, variables, {
      partialUpdate: true,
      reason
    });
  }
  
  async batchUpdateVariables(processInstanceId, updates) {
    const results = [];
    
    for (const update of updates) {
      try {
        const result = await this.updateVariables(
          processInstanceId,
          update.variables,
          update.options
        );
        
        results.push({
          success: true,
          update,
          result
        });
        
      } catch (error) {
        results.push({
          success: false,
          update,
          error: error.message
        });
      }
    }
    
    return results;
  }
  
  async safeUpdateVariables(processInstanceId, variables, options = {}) {
    try {
      // First, try without transformation to preserve exact values
      return await this.updateVariables(processInstanceId, variables, {
        ...options,
        transform: false
      });
      
    } catch (error) {
      // If it fails, try with transformation
      console.warn('Update failed without transformation, retrying with transformation');
      
      return await this.updateVariables(processInstanceId, variables, {
        ...options,
        transform: true
      });
    }
  }
  
  async validateBeforeUpdate(processInstanceId, variables) {
    // Use GET endpoint to get current schema
    const schemaResponse = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/variables?include_schema=true`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    if (!schemaResponse.ok) {
      throw new Error('Failed to fetch variable schema for validation');
    }
    
    const schemaData = await schemaResponse.json();
    const { schema } = schemaData.data;
    
    const validationResults = {};
    
    Object.entries(variables).forEach(([name, value]) => {
      const fieldSchema = schema.properties[name];
      if (fieldSchema) {
        validationResults[name] = this.validateValue(value, fieldSchema);
      } else {
        validationResults[name] = {
          valid: false,
          errors: ['Variable not found in schema']
        };
      }
    });
    
    const allValid = Object.values(validationResults).every(result => result.valid);
    
    return {
      valid: allValid,
      results: validationResults,
      canProceed: allValid
    };
  }
  
  validateValue(value, schema) {
    const errors = [];
    
    // Type validation
    const expectedType = schema.type;
    const actualType = Array.isArray(value) ? 'array' : typeof value;
    
    if (actualType !== expectedType) {
      errors.push(`Expected ${expectedType}, got ${actualType}`);
    }
    
    // String validations
    if (expectedType === 'string' && typeof value === 'string') {
      if (schema.minLength && value.length < schema.minLength) {
        errors.push(`Minimum length is ${schema.minLength}`);
      }
      
      if (schema.maxLength && value.length > schema.maxLength) {
        errors.push(`Maximum length is ${schema.maxLength}`);
      }
      
      if (schema.pattern && !new RegExp(schema.pattern).test(value)) {
        errors.push('Value does not match required pattern');
      }
      
      if (schema.enum && !schema.enum.includes(value)) {
        errors.push(`Value must be one of: ${schema.enum.join(', ')}`);
      }
    }
    
    // Number validations
    if (expectedType === 'number' && typeof value === 'number') {
      if (schema.minimum !== undefined && value < schema.minimum) {
        errors.push(`Minimum value is ${schema.minimum}`);
      }
      
      if (schema.maximum !== undefined && value > schema.maximum) {
        errors.push(`Maximum value is ${schema.maximum}`);
      }
    }
    
    return {
      valid: errors.length === 0,
      errors
    };
  }
  
  async createVariableUpdateForm(processInstanceId, variableNames, containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container with ID '${containerId}' not found`);
    }
    
    // Get current variables and schema
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/variables?variables=${variableNames.join(',')}&include_schema=true`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    const data = await response.json();
    const { variables, schema } = data.data;
    
    const form = document.createElement('form');
    form.className = 'variable-update-form';
    
    variableNames.forEach(name => {
      if (variables[name]) {
        const fieldDiv = this.createVariableField(name, variables[name], schema.properties[name]);
        form.appendChild(fieldDiv);
      }
    });
    
    // Add update reason field
    const reasonDiv = document.createElement('div');
    reasonDiv.className = 'form-field';
    
    const reasonLabel = document.createElement('label');
    reasonLabel.textContent = 'Update Reason';
    reasonDiv.appendChild(reasonLabel);
    
    const reasonInput = document.createElement('input');
    reasonInput.type = 'text';
    reasonInput.name = 'update_reason';
    reasonInput.placeholder = 'Reason for this update';
    reasonDiv.appendChild(reasonInput);
    
    form.appendChild(reasonDiv);
    
    // Add submit button
    const submitButton = document.createElement('button');
    submitButton.type = 'submit';
    submitButton.textContent = 'Update Variables';
    submitButton.className = 'btn btn-primary';
    form.appendChild(submitButton);
    
    // Add form submission handler
    form.onsubmit = async (e) => {
      e.preventDefault();
      
      const formData = new FormData(form);
      const updateVariables = {};
      const updateReason = formData.get('update_reason');
      
      variableNames.forEach(name => {
        const value = formData.get(name);
        if (value !== null && value !== '') {
          // Convert based on original type
          const originalType = variables[name].type;
          updateVariables[name] = this.convertFormValue(value, originalType);
        }
      });
      
      try {
        const result = await this.updateVariables(processInstanceId, updateVariables, {
          reason: updateReason
        });
        
        this.showUpdateSuccess(form, result);
        
      } catch (error) {
        this.showUpdateError(form, error.message);
      }
    };
    
    container.appendChild(form);
    
    return form;
  }
  
  createVariableField(name, variable, schema) {
    const fieldDiv = document.createElement('div');
    fieldDiv.className = 'form-field';
    
    const label = document.createElement('label');
    label.textContent = schema?.title || name;
    if (variable.required) {
      label.innerHTML += ' <span class="required">*</span>';
    }
    fieldDiv.appendChild(label);
    
    let input;
    
    switch (variable.type) {
      case 'string':
        if (schema?.enum) {
          input = document.createElement('select');
          schema.enum.forEach(option => {
            const optionElement = document.createElement('option');
            optionElement.value = option;
            optionElement.textContent = option;
            optionElement.selected = option === variable.value;
            input.appendChild(optionElement);
          });
        } else {
          input = document.createElement('input');
          input.type = 'text';
          input.value = variable.value || '';
        }
        break;
        
      case 'number':
        input = document.createElement('input');
        input.type = 'number';
        input.value = variable.value || '';
        if (schema?.minimum !== undefined) input.min = schema.minimum;
        if (schema?.maximum !== undefined) input.max = schema.maximum;
        break;
        
      case 'boolean':
        input = document.createElement('input');
        input.type = 'checkbox';
        input.checked = variable.value || false;
        break;
        
      default:
        input = document.createElement('textarea');
        input.value = JSON.stringify(variable.value, null, 2);
        break;
    }
    
    input.name = name;
    input.disabled = variable.readonly;
    
    fieldDiv.appendChild(input);
    
    if (variable.description) {
      const description = document.createElement('small');
      description.className = 'field-description';
      description.textContent = variable.description;
      fieldDiv.appendChild(description);
    }
    
    return fieldDiv;
  }
  
  convertFormValue(value, type) {
    switch (type) {
      case 'number':
        return parseFloat(value);
      case 'boolean':
        return value === 'on' || value === 'true';
      case 'object':
      case 'array':
        try {
          return JSON.parse(value);
        } catch {
          return value;
        }
      default:
        return value;
    }
  }
  
  showUpdateSuccess(form, result) {
    const message = document.createElement('div');
    message.className = 'update-success';
    message.textContent = `Successfully updated ${Object.keys(result.data.updated_variables).length} variables`;
    
    form.insertBefore(message, form.firstChild);
    
    setTimeout(() => {
      message.remove();
    }, 5000);
  }
  
  showUpdateError(form, errorMessage) {
    const message = document.createElement('div');
    message.className = 'update-error';
    message.textContent = `Update failed: ${errorMessage}`;
    
    form.insertBefore(message, form.firstChild);
    
    setTimeout(() => {
      message.remove();
    }, 10000);
  }
}

// Использование
const updater = new TypedVariablesUpdater('your-api-key');

// Простое обновление переменной
const result = await updater.updateSingleVariable(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  'paymentStatus',
  'COMPLETED',
  'Payment processed by gateway'
);

// Обновление нескольких переменных
const multiResult = await updater.updateVariables(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  {
    totalAmount: 1299.98,
    paymentStatus: 'PENDING',
    'customer.type': 'VIP'
  },
  { reason: 'Order modified by customer' }
);

// Валидация перед обновлением
const validation = await updater.validateBeforeUpdate(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  {
    paymentStatus: 'COMPLETED',
    totalAmount: 1500
  }
);

if (validation.valid) {
  console.log('Validation passed, proceeding with update');
} else {
  console.log('Validation failed:', validation.results);
}

// Создание формы обновления
await updater.createVariableUpdateForm(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  ['paymentStatus', 'totalAmount', 'customer'],
  'update-form-container'
);
```

## Связанные endpoints
- [`GET /api/v1/processes/:id/typed/variables`](./get-process-typed-variables.md) - Получение текущих переменных
- [`GET /api/v1/processes/:id/typed/info`](./get-process-typed-info.md) - Схема и метаданные переменных
- [`PUT /api/v1/processes/:id/variables`](../processes/update-process-variables.md) - Базовое обновление переменных
