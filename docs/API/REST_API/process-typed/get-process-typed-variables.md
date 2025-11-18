# GET /api/v1/processes/:id/typed/variables

## Описание
Получение типизированных переменных процесса с JSON Schema и метаданными для валидации и рендеринга форм.

## URL
```
GET /api/v1/processes/:id/typed/variables
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры запроса (Query Parameters)
- `include_schema` (boolean): Включить JSON Schema (по умолчанию: true)
- `include_history` (boolean): Включить историю изменений (по умолчанию: false)
- `variables` (string): Список переменных через запятую (если не указан, возвращаются все)

## Примеры запросов

### Все типизированные переменные
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables" \
  -H "X-API-Key: your-api-key-here"
```

### Конкретные переменные с историей
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables?variables=orderId,customer&include_history=true" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/variables', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const typedVariables = await response.json();
```

## Ответы

### 200 OK - Типизированные переменные
```json
{
  "success": true,
  "data": {
    "process_instance_id": "srv1-aB3dEf9hK2mN5pQ8uV",
    "last_updated": "2025-01-11T10:30:45.456Z",
    "variables": {
      "orderId": {
        "value": "ORD-2025-001234",
        "type": "string",
        "required": true,
        "readonly": true,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:15:30.123Z",
        "updated_by": "system",
        "source": "process_start",
        "description": "Unique order identifier",
        "validation_rules": {
          "pattern": "^ORD-\\d{4}-\\d{6}$",
          "min_length": 15,
          "max_length": 15
        },
        "ui_hints": {
          "display_name": "Order ID",
          "input_type": "text",
          "placeholder": "ORD-YYYY-NNNNNN"
        }
      },
      "customer": {
        "value": {
          "id": "CUST-789456",
          "name": "John Doe",
          "email": "john.doe@example.com",
          "type": "PREMIUM",
          "address": {
            "street": "123 Main St",
            "city": "Boston",
            "state": "MA",
            "zipCode": "02101",
            "country": "USA"
          }
        },
        "type": "object",
        "required": true,
        "readonly": false,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:25:15.789Z",
        "updated_by": "validation_service",
        "source": "external_api",
        "description": "Complete customer information",
        "validation_rules": {
          "required_fields": ["id", "name", "email"],
          "custom_validation": "customer_exists_check"
        },
        "ui_hints": {
          "display_name": "Customer",
          "input_type": "object_form",
          "sections": [
            {
              "title": "Basic Info",
              "fields": ["id", "name", "email", "type"]
            },
            {
              "title": "Address",
              "fields": ["address"]
            }
          ]
        }
      },
      "orderItems": {
        "value": [
          {
            "productId": "PROD-001",
            "name": "Laptop Computer",
            "category": "Electronics",
            "quantity": 1,
            "unitPrice": 999.99,
            "discount": 0,
            "totalPrice": 999.99
          },
          {
            "productId": "PROD-002",
            "name": "Wireless Mouse",
            "category": "Electronics",
            "quantity": 2,
            "unitPrice": 25.50,
            "discount": 0.1,
            "totalPrice": 45.90
          }
        ],
        "type": "array",
        "required": true,
        "readonly": false,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:20:10.456Z",
        "updated_by": "inventory_service",
        "source": "cart_calculation",
        "description": "List of ordered items with pricing",
        "validation_rules": {
          "min_items": 1,
          "max_items": 50,
          "item_validation": "product_exists_check"
        },
        "ui_hints": {
          "display_name": "Order Items",
          "input_type": "dynamic_table",
          "columns": [
            {"field": "productId", "title": "Product ID", "width": "15%"},
            {"field": "name", "title": "Product Name", "width": "30%"},
            {"field": "quantity", "title": "Qty", "width": "10%"},
            {"field": "unitPrice", "title": "Unit Price", "width": "15%"},
            {"field": "totalPrice", "title": "Total", "width": "15%"}
          ],
          "add_button": "Add Item",
          "remove_button": "Remove"
        }
      },
      "totalAmount": {
        "value": 1050.99,
        "type": "number",
        "required": true,
        "readonly": false,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:20:10.456Z",
        "updated_by": "calculation_service",
        "source": "automatic_calculation",
        "description": "Total order amount including tax",
        "validation_rules": {
          "minimum": 0,
          "decimal_places": 2,
          "currency": "USD"
        },
        "ui_hints": {
          "display_name": "Total Amount",
          "input_type": "currency",
          "currency_symbol": "$",
          "decimal_places": 2,
          "readonly_calculated": true
        }
      },
      "paymentStatus": {
        "value": "PROCESSING",
        "type": "string",
        "required": true,
        "readonly": false,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:30:45.456Z",
        "updated_by": "payment_service",
        "source": "payment_gateway",
        "description": "Current payment processing status",
        "validation_rules": {
          "enum": ["PENDING", "PROCESSING", "COMPLETED", "FAILED", "REFUNDED"]
        },
        "ui_hints": {
          "display_name": "Payment Status",
          "input_type": "select",
          "options": [
            {"value": "PENDING", "label": "Pending", "color": "gray"},
            {"value": "PROCESSING", "label": "Processing", "color": "blue"},
            {"value": "COMPLETED", "label": "Completed", "color": "green"},
            {"value": "FAILED", "label": "Failed", "color": "red"},
            {"value": "REFUNDED", "label": "Refunded", "color": "orange"}
          ]
        }
      },
      "approvalRequired": {
        "value": true,
        "type": "boolean",
        "required": false,
        "readonly": false,
        "created_at": "2025-01-11T10:15:30.123Z",
        "updated_at": "2025-01-11T10:25:15.789Z",
        "updated_by": "business_rules",
        "source": "rule_evaluation",
        "description": "Whether manager approval is required",
        "validation_rules": {},
        "ui_hints": {
          "display_name": "Requires Approval",
          "input_type": "checkbox",
          "help_text": "Check if this order requires manager approval"
        }
      }
    },
    "schema": {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "title": "Order Processing Variables",
      "properties": {
        "orderId": {
          "type": "string",
          "title": "Order ID",
          "description": "Unique order identifier",
          "pattern": "^ORD-\\d{4}-\\d{6}$",
          "minLength": 15,
          "maxLength": 15,
          "readOnly": true
        },
        "customer": {
          "type": "object",
          "title": "Customer",
          "description": "Complete customer information",
          "properties": {
            "id": {
              "type": "string",
              "title": "Customer ID",
              "pattern": "^CUST-\\d{6}$"
            },
            "name": {
              "type": "string",
              "title": "Full Name",
              "minLength": 1,
              "maxLength": 100
            },
            "email": {
              "type": "string",
              "title": "Email Address",
              "format": "email"
            },
            "type": {
              "type": "string",
              "title": "Customer Type",
              "enum": ["REGULAR", "PREMIUM", "VIP"],
              "default": "REGULAR"
            },
            "address": {
              "type": "object",
              "title": "Address",
              "properties": {
                "street": {"type": "string", "title": "Street"},
                "city": {"type": "string", "title": "City"},
                "state": {"type": "string", "title": "State"},
                "zipCode": {"type": "string", "title": "Zip Code"},
                "country": {"type": "string", "title": "Country"}
              }
            }
          },
          "required": ["id", "name", "email"]
        },
        "orderItems": {
          "type": "array",
          "title": "Order Items",
          "description": "List of ordered items with pricing",
          "items": {
            "type": "object",
            "properties": {
              "productId": {"type": "string", "title": "Product ID"},
              "name": {"type": "string", "title": "Product Name"},
              "category": {"type": "string", "title": "Category"},
              "quantity": {"type": "integer", "title": "Quantity", "minimum": 1},
              "unitPrice": {"type": "number", "title": "Unit Price", "minimum": 0},
              "discount": {"type": "number", "title": "Discount", "minimum": 0, "maximum": 1},
              "totalPrice": {"type": "number", "title": "Total Price", "minimum": 0}
            },
            "required": ["productId", "name", "quantity", "unitPrice", "totalPrice"]
          },
          "minItems": 1,
          "maxItems": 50
        },
        "totalAmount": {
          "type": "number",
          "title": "Total Amount",
          "description": "Total order amount including tax",
          "minimum": 0,
          "multipleOf": 0.01
        },
        "paymentStatus": {
          "type": "string",
          "title": "Payment Status",
          "description": "Current payment processing status",
          "enum": ["PENDING", "PROCESSING", "COMPLETED", "FAILED", "REFUNDED"],
          "default": "PENDING"
        },
        "approvalRequired": {
          "type": "boolean",
          "title": "Requires Approval",
          "description": "Whether manager approval is required",
          "default": false
        }
      },
      "required": ["orderId", "customer", "orderItems", "totalAmount", "paymentStatus"]
    },
    "variable_history": [
      {
        "variable_name": "paymentStatus",
        "old_value": "PENDING",
        "new_value": "PROCESSING",
        "changed_at": "2025-01-11T10:30:45.456Z",
        "changed_by": "payment_service",
        "change_reason": "Payment processing initiated",
        "source_activity": "ServiceTask_PaymentProcess"
      },
      {
        "variable_name": "customer",
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
          "type": "PREMIUM",
          "address": {
            "street": "123 Main St",
            "city": "Boston",
            "state": "MA",
            "zipCode": "02101",
            "country": "USA"
          }
        },
        "changed_at": "2025-01-11T10:25:15.789Z",
        "changed_by": "validation_service",
        "change_reason": "Address information added from customer database",
        "source_activity": "ServiceTask_ValidateOrder"
      }
    ],
    "metadata": {
      "total_variables": 6,
      "readonly_variables": 1,
      "last_modification": "2025-01-11T10:30:45.456Z",
      "modification_count": 8,
      "data_size_bytes": 2048
    }
  },
  "request_id": "req_1641998405100"
}
```

## Использование

### Typed Variables Client
```javascript
class TypedVariablesClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getTypedVariables(processInstanceId, options = {}) {
    const params = new URLSearchParams();
    if (options.includeSchema !== undefined) {
      params.append('include_schema', options.includeSchema);
    }
    if (options.includeHistory !== undefined) {
      params.append('include_history', options.includeHistory);
    }
    if (options.variables) {
      params.append('variables', Array.isArray(options.variables) ? 
        options.variables.join(',') : options.variables);
    }
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/variables?${params}`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch typed variables: ${response.statusText}`);
    }
    
    return await response.json();
  }
  
  async getVariableValue(processInstanceId, variableName) {
    const result = await this.getTypedVariables(processInstanceId, {
      variables: [variableName],
      includeSchema: false,
      includeHistory: false
    });
    
    return result.data.variables[variableName]?.value;
  }
  
  async getVariableWithMetadata(processInstanceId, variableName) {
    const result = await this.getTypedVariables(processInstanceId, {
      variables: [variableName],
      includeSchema: true,
      includeHistory: true
    });
    
    return result.data.variables[variableName];
  }
  
  async validateVariable(processInstanceId, variableName, value) {
    const result = await this.getTypedVariables(processInstanceId, {
      variables: [variableName],
      includeSchema: true
    });
    
    const variable = result.data.variables[variableName];
    const schema = result.data.schema.properties[variableName];
    
    if (!variable || !schema) {
      return {
        valid: false,
        errors: [`Variable '${variableName}' not found`]
      };
    }
    
    return this.validateAgainstSchema(value, schema, variable.validation_rules);
  }
  
  validateAgainstSchema(value, schema, validationRules = {}) {
    const validation = {
      valid: true,
      errors: [],
      warnings: []
    };
    
    // Type validation
    if (schema.type && typeof value !== schema.type && 
        !(schema.type === 'array' && Array.isArray(value))) {
      validation.valid = false;
      validation.errors.push(`Expected type ${schema.type}, got ${typeof value}`);
    }
    
    // Required validation (if value is null/undefined)
    if ((value === null || value === undefined) && schema.required) {
      validation.valid = false;
      validation.errors.push('This field is required');
    }
    
    // String validations
    if (schema.type === 'string' && typeof value === 'string') {
      if (schema.minLength && value.length < schema.minLength) {
        validation.valid = false;
        validation.errors.push(`Minimum length is ${schema.minLength}`);
      }
      
      if (schema.maxLength && value.length > schema.maxLength) {
        validation.valid = false;
        validation.errors.push(`Maximum length is ${schema.maxLength}`);
      }
      
      if (schema.pattern && !new RegExp(schema.pattern).test(value)) {
        validation.valid = false;
        validation.errors.push('Value does not match required pattern');
      }
      
      if (schema.format === 'email' && !this.isValidEmail(value)) {
        validation.valid = false;
        validation.errors.push('Invalid email format');
      }
    }
    
    // Number validations
    if (schema.type === 'number' && typeof value === 'number') {
      if (schema.minimum !== undefined && value < schema.minimum) {
        validation.valid = false;
        validation.errors.push(`Minimum value is ${schema.minimum}`);
      }
      
      if (schema.maximum !== undefined && value > schema.maximum) {
        validation.valid = false;
        validation.errors.push(`Maximum value is ${schema.maximum}`);
      }
      
      if (schema.multipleOf && value % schema.multipleOf !== 0) {
        validation.valid = false;
        validation.errors.push(`Value must be a multiple of ${schema.multipleOf}`);
      }
    }
    
    // Array validations
    if (schema.type === 'array' && Array.isArray(value)) {
      if (schema.minItems && value.length < schema.minItems) {
        validation.valid = false;
        validation.errors.push(`Minimum ${schema.minItems} items required`);
      }
      
      if (schema.maxItems && value.length > schema.maxItems) {
        validation.valid = false;
        validation.errors.push(`Maximum ${schema.maxItems} items allowed`);
      }
      
      // Validate each item against schema.items if present
      if (schema.items) {
        value.forEach((item, index) => {
          const itemValidation = this.validateAgainstSchema(item, schema.items);
          if (!itemValidation.valid) {
            validation.valid = false;
            itemValidation.errors.forEach(error => {
              validation.errors.push(`Item ${index}: ${error}`);
            });
          }
        });
      }
    }
    
    // Enum validation
    if (schema.enum && !schema.enum.includes(value)) {
      validation.valid = false;
      validation.errors.push(`Value must be one of: ${schema.enum.join(', ')}`);
    }
    
    // Object validation
    if (schema.type === 'object' && typeof value === 'object' && value !== null) {
      if (schema.required) {
        schema.required.forEach(requiredField => {
          if (!(requiredField in value) || value[requiredField] === null || value[requiredField] === undefined) {
            validation.valid = false;
            validation.errors.push(`Required field '${requiredField}' is missing`);
          }
        });
      }
      
      // Validate properties if schema.properties exists
      if (schema.properties) {
        Object.entries(value).forEach(([key, fieldValue]) => {
          if (schema.properties[key]) {
            const fieldValidation = this.validateAgainstSchema(fieldValue, schema.properties[key]);
            if (!fieldValidation.valid) {
              validation.valid = false;
              fieldValidation.errors.forEach(error => {
                validation.errors.push(`${key}: ${error}`);
              });
            }
          }
        });
      }
    }
    
    return validation;
  }
  
  isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
  
  async getVariableHistory(processInstanceId, variableName) {
    const result = await this.getTypedVariables(processInstanceId, {
      variables: [variableName],
      includeHistory: true,
      includeSchema: false
    });
    
    return result.data.variable_history?.filter(h => h.variable_name === variableName) || [];
  }
  
  async generateFormConfig(processInstanceId, variableNames = null) {
    const options = { includeSchema: true };
    if (variableNames) {
      options.variables = variableNames;
    }
    
    const result = await this.getTypedVariables(processInstanceId, options);
    const { variables, schema } = result.data;
    
    const formConfig = {
      fields: [],
      validation: {},
      sections: []
    };
    
    Object.entries(variables).forEach(([name, variable]) => {
      const fieldSchema = schema.properties[name];
      
      const fieldConfig = {
        name,
        label: fieldSchema?.title || variable.ui_hints?.display_name || name,
        type: this.mapTypeToInputType(variable.type, variable.ui_hints),
        value: variable.value,
        required: variable.required,
        readonly: variable.readonly,
        description: variable.description,
        placeholder: variable.ui_hints?.placeholder,
        options: variable.ui_hints?.options || (fieldSchema?.enum ? 
          fieldSchema.enum.map(val => ({ value: val, label: val })) : null)
      };
      
      formConfig.fields.push(fieldConfig);
      
      // Add validation rules
      if (fieldSchema) {
        formConfig.validation[name] = this.extractValidationRules(fieldSchema);
      }
    });
    
    return formConfig;
  }
  
  mapTypeToInputType(type, uiHints = {}) {
    if (uiHints.input_type) {
      return uiHints.input_type;
    }
    
    switch (type) {
      case 'string': return 'text';
      case 'number': return 'number';
      case 'boolean': return 'checkbox';
      case 'array': return 'list';
      case 'object': return 'object';
      default: return 'text';
    }
  }
  
  extractValidationRules(schema) {
    const rules = {};
    
    if (schema.required) rules.required = true;
    if (schema.minLength) rules.minLength = schema.minLength;
    if (schema.maxLength) rules.maxLength = schema.maxLength;
    if (schema.minimum) rules.min = schema.minimum;
    if (schema.maximum) rules.max = schema.maximum;
    if (schema.pattern) rules.pattern = schema.pattern;
    if (schema.format) rules.format = schema.format;
    if (schema.enum) rules.enum = schema.enum;
    
    return rules;
  }
  
  async watchVariableChanges(processInstanceId, variableNames, callback, interval = 2000) {
    let lastValues = {};
    
    // Get initial values
    try {
      const initial = await this.getTypedVariables(processInstanceId, {
        variables: variableNames,
        includeSchema: false,
        includeHistory: false
      });
      
      Object.entries(initial.data.variables).forEach(([name, variable]) => {
        lastValues[name] = JSON.stringify(variable.value);
      });
    } catch (error) {
      console.error('Error getting initial variable values:', error);
    }
    
    const pollVariables = async () => {
      try {
        const current = await this.getTypedVariables(processInstanceId, {
          variables: variableNames,
          includeSchema: false,
          includeHistory: false
        });
        
        const changes = [];
        
        Object.entries(current.data.variables).forEach(([name, variable]) => {
          const currentValue = JSON.stringify(variable.value);
          const lastValue = lastValues[name];
          
          if (currentValue !== lastValue) {
            changes.push({
              name,
              oldValue: lastValue ? JSON.parse(lastValue) : undefined,
              newValue: variable.value,
              updatedAt: variable.updated_at,
              updatedBy: variable.updated_by
            });
            
            lastValues[name] = currentValue;
          }
        });
        
        if (changes.length > 0) {
          callback(changes);
        }
        
      } catch (error) {
        console.error('Error polling variable changes:', error);
      }
    };
    
    const intervalId = setInterval(pollVariables, interval);
    
    return () => clearInterval(intervalId);
  }
}

// Использование
const client = new TypedVariablesClient('your-api-key');

// Получение всех типизированных переменных
const allVariables = await client.getTypedVariables('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('All variables:', allVariables);

// Получение конкретной переменной с метаданными
const customerVar = await client.getVariableWithMetadata('srv1-aB3dEf9hK2mN5pQ8uV', 'customer');
console.log('Customer variable:', customerVar);

// Валидация значения переменной
const validation = await client.validateVariable(
  'srv1-aB3dEf9hK2mN5pQ8uV', 
  'customer', 
  {
    id: 'CUST-123456',
    name: 'John Doe',
    email: 'invalid-email',
    type: 'PREMIUM'
  }
);
console.log('Validation result:', validation);

// Генерация конфигурации формы
const formConfig = await client.generateFormConfig('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Form configuration:', formConfig);

// Отслеживание изменений переменных
const stopWatching = await client.watchVariableChanges(
  'srv1-aB3dEf9hK2mN5pQ8uV',
  ['paymentStatus', 'totalAmount'],
  (changes) => {
    console.log('Variable changes detected:', changes);
  }
);

// Остановка отслеживания через 60 секунд
setTimeout(stopWatching, 60000);
```

## Связанные endpoints
- [`PUT /api/v1/processes/:id/typed/variables`](./update-process-typed-variables.md) - Обновление типизированных переменных
- [`GET /api/v1/processes/:id/typed/info`](./get-process-typed-info.md) - Типизированная информация процесса
- [`GET /api/v1/processes/:id/variables`](../processes/get-process-variables.md) - Базовые переменные процесса
