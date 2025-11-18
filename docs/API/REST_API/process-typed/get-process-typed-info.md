# GET /api/v1/processes/:id/typed/info

## Описание
Получение типизированной информации о процессе с JSON Schema и метаданными для веб-клиентов. Предоставляет структурированные данные с типами полей.

## URL
```
GET /api/v1/processes/:id/typed/info
```

## Авторизация
✅ **Требуется API ключ** с разрешением `process`

## Параметры пути
- `id` (string, обязательный): Уникальный идентификатор экземпляра процесса

## Параметры запроса (Query Parameters)
- `include_schema` (boolean): Включить JSON Schema для полей (по умолчанию: true)
- `include_metadata` (boolean): Включить метаданные процесса (по умолчанию: true)

## Примеры запросов

### Типизированная информация о процессе
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/info" \
  -H "X-API-Key: your-api-key-here"
```

### Без схемы, только данные
```bash
curl -X GET "http://localhost:27555/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/info?include_schema=false" \
  -H "X-API-Key: your-api-key-here"
```

### JavaScript
```javascript
const response = await fetch('/api/v1/processes/srv1-aB3dEf9hK2mN5pQ8uV/typed/info', {
  headers: {
    'X-API-Key': 'your-api-key-here'
  }
});

const typedInfo = await response.json();
```

## Ответы

### 200 OK - Типизированная информация процесса
```json
{
  "success": true,
  "data": {
    "process_instance": {
      "id": "srv1-aB3dEf9hK2mN5pQ8uV",
      "process_definition_key": "Order_Processing",
      "process_definition_version": 3,
      "status": "ACTIVE",
      "started_at": "2025-01-11T10:15:30.123Z",
      "updated_at": "2025-01-11T10:30:45.456Z",
      "tenant_id": "tenant-001"
    },
    "typed_variables": {
      "orderId": {
        "value": "ORD-2025-001234",
        "type": "string",
        "required": true,
        "readonly": true,
        "description": "Unique order identifier"
      },
      "customer": {
        "value": {
          "id": "CUST-789456",
          "name": "John Doe",
          "email": "john.doe@example.com",
          "type": "PREMIUM"
        },
        "type": "object",
        "required": true,
        "readonly": false,
        "description": "Customer information"
      },
      "orderItems": {
        "value": [
          {
            "productId": "PROD-001",
            "name": "Laptop",
            "quantity": 1,
            "price": 999.99
          },
          {
            "productId": "PROD-002", 
            "name": "Mouse",
            "quantity": 2,
            "price": 25.50
          }
        ],
        "type": "array",
        "required": true,
        "readonly": false,
        "description": "List of ordered items"
      },
      "totalAmount": {
        "value": 1050.99,
        "type": "number",
        "required": true,
        "readonly": false,
        "description": "Total order amount including tax",
        "format": "currency",
        "currency": "USD"
      },
      "paymentStatus": {
        "value": "PENDING",
        "type": "string",
        "required": true,
        "readonly": false,
        "description": "Current payment status",
        "enum": ["PENDING", "PROCESSING", "COMPLETED", "FAILED"]
      },
      "orderDate": {
        "value": "2025-01-11T10:15:30.123Z",
        "type": "string",
        "required": true,
        "readonly": true,
        "description": "Order creation date",
        "format": "date-time"
      },
      "approved": {
        "value": true,
        "type": "boolean",
        "required": false,
        "readonly": false,
        "description": "Whether the order is approved"
      }
    },
    "process_metadata": {
      "definition": {
        "key": "Order_Processing",
        "name": "Order Processing Workflow",
        "version": 3,
        "description": "Complete order processing from submission to fulfillment",
        "category": "E-Commerce",
        "tags": ["order", "payment", "fulfillment"]
      },
      "execution_info": {
        "current_activities": [
          {
            "element_id": "ServiceTask_PaymentProcess",
            "element_name": "Process Payment",
            "element_type": "SERVICE_TASK",
            "started_at": "2025-01-11T10:30:15.789Z",
            "estimated_completion": "2025-01-11T10:32:00.000Z"
          }
        ],
        "completed_activities": [
          {
            "element_id": "StartEvent_1",
            "element_name": "Order Received",
            "completed_at": "2025-01-11T10:15:30.123Z",
            "duration_ms": 50
          },
          {
            "element_id": "ServiceTask_ValidateOrder",
            "element_name": "Validate Order",
            "completed_at": "2025-01-11T10:30:10.456Z",
            "duration_ms": 4500
          }
        ],
        "progress": {
          "percentage": 35.5,
          "current_step": 3,
          "total_estimated_steps": 8,
          "estimated_completion": "2025-01-11T10:45:00.000Z"
        }
      },
      "business_context": {
        "department": "Sales",
        "priority": "HIGH",
        "sla_deadline": "2025-01-11T18:00:00.000Z",
        "escalation_required": false
      }
    },
    "json_schema": {
      "$schema": "http://json-schema.org/draft-07/schema#",
      "type": "object",
      "title": "Order Processing Variables",
      "properties": {
        "orderId": {
          "type": "string",
          "title": "Order ID",
          "description": "Unique order identifier",
          "pattern": "^ORD-\\d{4}-\\d{6}$",
          "readOnly": true
        },
        "customer": {
          "type": "object",
          "title": "Customer",
          "description": "Customer information",
          "properties": {
            "id": {
              "type": "string",
              "title": "Customer ID",
              "pattern": "^CUST-\\d{6}$"
            },
            "name": {
              "type": "string",
              "title": "Customer Name",
              "minLength": 1,
              "maxLength": 100
            },
            "email": {
              "type": "string",
              "title": "Email",
              "format": "email"
            },
            "type": {
              "type": "string",
              "title": "Customer Type",
              "enum": ["REGULAR", "PREMIUM", "VIP"],
              "default": "REGULAR"
            }
          },
          "required": ["id", "name", "email"],
          "additionalProperties": false
        },
        "orderItems": {
          "type": "array",
          "title": "Order Items",
          "description": "List of ordered items",
          "items": {
            "type": "object",
            "properties": {
              "productId": {
                "type": "string",
                "title": "Product ID",
                "pattern": "^PROD-\\d{3}$"
              },
              "name": {
                "type": "string",
                "title": "Product Name",
                "minLength": 1
              },
              "quantity": {
                "type": "integer",
                "title": "Quantity",
                "minimum": 1,
                "maximum": 100
              },
              "price": {
                "type": "number",
                "title": "Unit Price",
                "minimum": 0,
                "multipleOf": 0.01
              }
            },
            "required": ["productId", "name", "quantity", "price"]
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
          "description": "Current payment status",
          "enum": ["PENDING", "PROCESSING", "COMPLETED", "FAILED"],
          "default": "PENDING"
        },
        "orderDate": {
          "type": "string",
          "title": "Order Date",
          "description": "Order creation date",
          "format": "date-time",
          "readOnly": true
        },
        "approved": {
          "type": "boolean",
          "title": "Approved",
          "description": "Whether the order is approved",
          "default": false
        }
      },
      "required": ["orderId", "customer", "orderItems", "totalAmount", "paymentStatus", "orderDate"]
    },
    "ui_metadata": {
      "form_layout": {
        "sections": [
          {
            "title": "Order Information",
            "fields": ["orderId", "orderDate", "totalAmount"],
            "readonly": true
          },
          {
            "title": "Customer Details",
            "fields": ["customer"],
            "collapsible": true
          },
          {
            "title": "Order Items",
            "fields": ["orderItems"],
            "display_type": "table"
          },
          {
            "title": "Processing Status",
            "fields": ["paymentStatus", "approved"],
            "conditional": {
              "show_when": "paymentStatus !== 'PENDING'"
            }
          }
        ]
      },
      "field_hints": {
        "customer.email": {
          "input_type": "email",
          "placeholder": "customer@example.com"
        },
        "totalAmount": {
          "input_type": "currency",
          "currency_symbol": "$",
          "decimal_places": 2
        },
        "orderItems": {
          "display_type": "dynamic_table",
          "add_button_text": "Add Item",
          "remove_button_text": "Remove"
        }
      },
      "validation_messages": {
        "customer.email": "Please enter a valid email address",
        "orderItems": "At least one item is required",
        "totalAmount": "Total amount must be greater than 0"
      }
    }
  },
  "request_id": "req_1641998404900"
}
```

## Использование

### Typed Process Info Client
```javascript
class TypedProcessInfoClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
  }
  
  async getTypedProcessInfo(processInstanceId, options = {}) {
    const params = new URLSearchParams();
    if (options.includeSchema !== undefined) {
      params.append('include_schema', options.includeSchema);
    }
    if (options.includeMetadata !== undefined) {
      params.append('include_metadata', options.includeMetadata);
    }
    
    const response = await fetch(
      `/api/v1/processes/${processInstanceId}/typed/info?${params}`,
      {
        headers: { 'X-API-Key': this.apiKey }
      }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch typed process info: ${response.statusText}`);
    }
    
    return await response.json();
  }
  
  async getVariableSchema(processInstanceId) {
    const info = await this.getTypedProcessInfo(processInstanceId, {
      includeSchema: true,
      includeMetadata: false
    });
    
    return info.data.json_schema;
  }
  
  async getTypedVariables(processInstanceId) {
    const info = await this.getTypedProcessInfo(processInstanceId, {
      includeSchema: false,
      includeMetadata: false
    });
    
    return info.data.typed_variables;
  }
  
  async getProcessMetadata(processInstanceId) {
    const info = await this.getTypedProcessInfo(processInstanceId, {
      includeSchema: false,
      includeMetadata: true
    });
    
    return info.data.process_metadata;
  }
  
  validateVariableValue(variableSchema, value) {
    const validation = {
      valid: true,
      errors: []
    };
    
    // Type validation
    const expectedType = variableSchema.type;
    const actualType = Array.isArray(value) ? 'array' : typeof value;
    
    if (actualType !== expectedType) {
      validation.valid = false;
      validation.errors.push({
        field: 'type',
        message: `Expected ${expectedType}, got ${actualType}`
      });
    }
    
    // Required validation
    if (variableSchema.required && (value === null || value === undefined)) {
      validation.valid = false;
      validation.errors.push({
        field: 'required',
        message: 'This field is required'
      });
    }
    
    // Enum validation
    if (variableSchema.enum && !variableSchema.enum.includes(value)) {
      validation.valid = false;
      validation.errors.push({
        field: 'enum',
        message: `Value must be one of: ${variableSchema.enum.join(', ')}`
      });
    }
    
    // String validations
    if (expectedType === 'string' && typeof value === 'string') {
      if (variableSchema.minLength && value.length < variableSchema.minLength) {
        validation.valid = false;
        validation.errors.push({
          field: 'minLength',
          message: `Minimum length is ${variableSchema.minLength}`
        });
      }
      
      if (variableSchema.maxLength && value.length > variableSchema.maxLength) {
        validation.valid = false;
        validation.errors.push({
          field: 'maxLength',
          message: `Maximum length is ${variableSchema.maxLength}`
        });
      }
      
      if (variableSchema.pattern && !new RegExp(variableSchema.pattern).test(value)) {
        validation.valid = false;
        validation.errors.push({
          field: 'pattern',
          message: 'Value does not match required pattern'
        });
      }
    }
    
    // Number validations
    if (expectedType === 'number' && typeof value === 'number') {
      if (variableSchema.minimum !== undefined && value < variableSchema.minimum) {
        validation.valid = false;
        validation.errors.push({
          field: 'minimum',
          message: `Minimum value is ${variableSchema.minimum}`
        });
      }
      
      if (variableSchema.maximum !== undefined && value > variableSchema.maximum) {
        validation.valid = false;
        validation.errors.push({
          field: 'maximum',
          message: `Maximum value is ${variableSchema.maximum}`
        });
      }
    }
    
    return validation;
  }
  
  generateFormFields(typedVariables, jsonSchema) {
    const formFields = [];
    
    Object.entries(typedVariables).forEach(([fieldName, fieldInfo]) => {
      const schemaProperty = jsonSchema.properties[fieldName];
      
      const formField = {
        name: fieldName,
        label: schemaProperty?.title || fieldName,
        description: fieldInfo.description,
        type: fieldInfo.type,
        value: fieldInfo.value,
        required: fieldInfo.required,
        readonly: fieldInfo.readonly,
        validation: this.extractValidationRules(schemaProperty),
        ui_hints: this.extractUIHints(fieldInfo, schemaProperty)
      };
      
      formFields.push(formField);
    });
    
    return formFields;
  }
  
  extractValidationRules(schemaProperty) {
    if (!schemaProperty) return {};
    
    const rules = {};
    
    if (schemaProperty.minLength) rules.minLength = schemaProperty.minLength;
    if (schemaProperty.maxLength) rules.maxLength = schemaProperty.maxLength;
    if (schemaProperty.minimum) rules.minimum = schemaProperty.minimum;
    if (schemaProperty.maximum) rules.maximum = schemaProperty.maximum;
    if (schemaProperty.pattern) rules.pattern = schemaProperty.pattern;
    if (schemaProperty.enum) rules.enum = schemaProperty.enum;
    if (schemaProperty.format) rules.format = schemaProperty.format;
    
    return rules;
  }
  
  extractUIHints(fieldInfo, schemaProperty) {
    const hints = {};
    
    if (fieldInfo.format === 'currency') {
      hints.inputType = 'currency';
      hints.currency = fieldInfo.currency || 'USD';
    } else if (schemaProperty?.format === 'email') {
      hints.inputType = 'email';
    } else if (schemaProperty?.format === 'date-time') {
      hints.inputType = 'datetime-local';
    } else if (fieldInfo.type === 'boolean') {
      hints.inputType = 'checkbox';
    } else if (fieldInfo.enum) {
      hints.inputType = 'select';
      hints.options = fieldInfo.enum;
    }
    
    return hints;
  }
}

// Использование
const client = new TypedProcessInfoClient('your-api-key');

// Получение полной типизированной информации
const typedInfo = await client.getTypedProcessInfo('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Typed process info:', typedInfo);

// Получение только схемы переменных
const schema = await client.getVariableSchema('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Variable schema:', schema);

// Получение типизированных переменных
const variables = await client.getTypedVariables('srv1-aB3dEf9hK2mN5pQ8uV');
console.log('Typed variables:', variables);

// Генерация полей формы
const formFields = client.generateFormFields(
  typedInfo.data.typed_variables,
  typedInfo.data.json_schema
);
console.log('Form fields:', formFields);

// Валидация значения переменной
const validation = client.validateVariableValue(
  typedInfo.data.json_schema.properties.customer,
  {
    id: 'CUST-123456',
    name: 'John Doe',
    email: 'invalid-email',
    type: 'PREMIUM'
  }
);
console.log('Validation result:', validation);
```

### Dynamic Form Generator
```javascript
class DynamicFormGenerator {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.client = new TypedProcessInfoClient(apiKey);
  }
  
  async generateForm(processInstanceId, containerId) {
    const typedInfo = await this.client.getTypedProcessInfo(processInstanceId);
    const { typed_variables, json_schema, ui_metadata } = typedInfo.data;
    
    const container = document.getElementById(containerId);
    if (!container) {
      throw new Error(`Container with ID '${containerId}' not found`);
    }
    
    container.innerHTML = ''; // Clear existing content
    
    const form = document.createElement('form');
    form.className = 'typed-process-form';
    
    // Generate form sections based on UI metadata
    if (ui_metadata?.form_layout?.sections) {
      ui_metadata.form_layout.sections.forEach(section => {
        const sectionElement = this.createFormSection(
          section,
          typed_variables,
          json_schema,
          ui_metadata
        );
        form.appendChild(sectionElement);
      });
    } else {
      // Fallback: create single section with all fields
      const allFields = Object.keys(typed_variables);
      const defaultSection = {
        title: 'Process Variables',
        fields: allFields
      };
      
      const sectionElement = this.createFormSection(
        defaultSection,
        typed_variables,
        json_schema,
        ui_metadata
      );
      form.appendChild(sectionElement);
    }
    
    // Add form controls
    const controlsDiv = document.createElement('div');
    controlsDiv.className = 'form-controls';
    
    const updateButton = document.createElement('button');
    updateButton.type = 'submit';
    updateButton.textContent = 'Update Variables';
    updateButton.className = 'btn btn-primary';
    
    const resetButton = document.createElement('button');
    resetButton.type = 'button';
    resetButton.textContent = 'Reset';
    resetButton.className = 'btn btn-secondary';
    resetButton.onclick = () => this.resetForm(form, typed_variables);
    
    controlsDiv.appendChild(updateButton);
    controlsDiv.appendChild(resetButton);
    form.appendChild(controlsDiv);
    
    // Add form submission handler
    form.onsubmit = (e) => this.handleFormSubmit(e, processInstanceId);
    
    container.appendChild(form);
    
    return {
      form,
      validate: () => this.validateForm(form, json_schema),
      getValues: () => this.getFormValues(form),
      setValues: (values) => this.setFormValues(form, values)
    };
  }
  
  createFormSection(section, typedVariables, jsonSchema, uiMetadata) {
    const sectionDiv = document.createElement('div');
    sectionDiv.className = 'form-section';
    
    // Section header
    const headerDiv = document.createElement('div');
    headerDiv.className = 'section-header';
    
    const title = document.createElement('h3');
    title.textContent = section.title;
    headerDiv.appendChild(title);
    
    if (section.collapsible) {
      const toggleButton = document.createElement('button');
      toggleButton.type = 'button';
      toggleButton.textContent = '▼';
      toggleButton.className = 'collapse-toggle';
      toggleButton.onclick = () => this.toggleSection(sectionDiv);
      headerDiv.appendChild(toggleButton);
    }
    
    sectionDiv.appendChild(headerDiv);
    
    // Section content
    const contentDiv = document.createElement('div');
    contentDiv.className = 'section-content';
    
    if (section.readonly) {
      contentDiv.classList.add('readonly');
    }
    
    section.fields.forEach(fieldName => {
      if (typedVariables[fieldName]) {
        const fieldElement = this.createFormField(
          fieldName,
          typedVariables[fieldName],
          jsonSchema.properties[fieldName],
          uiMetadata
        );
        contentDiv.appendChild(fieldElement);
      }
    });
    
    sectionDiv.appendChild(contentDiv);
    
    return sectionDiv;
  }
  
  createFormField(fieldName, fieldInfo, schemaProperty, uiMetadata) {
    const fieldDiv = document.createElement('div');
    fieldDiv.className = 'form-field';
    fieldDiv.dataset.fieldName = fieldName;
    
    // Label
    const label = document.createElement('label');
    label.textContent = schemaProperty?.title || fieldName;
    if (fieldInfo.required) {
      label.innerHTML += ' <span class="required">*</span>';
    }
    fieldDiv.appendChild(label);
    
    // Input element
    const inputElement = this.createInputElement(
      fieldName,
      fieldInfo,
      schemaProperty,
      uiMetadata
    );
    fieldDiv.appendChild(inputElement);
    
    // Description
    if (fieldInfo.description) {
      const description = document.createElement('small');
      description.className = 'field-description';
      description.textContent = fieldInfo.description;
      fieldDiv.appendChild(description);
    }
    
    // Validation message container
    const validationDiv = document.createElement('div');
    validationDiv.className = 'validation-message';
    fieldDiv.appendChild(validationDiv);
    
    return fieldDiv;
  }
  
  createInputElement(fieldName, fieldInfo, schemaProperty, uiMetadata) {
    const hints = uiMetadata?.field_hints?.[fieldName] || {};
    
    switch (fieldInfo.type) {
      case 'string':
        if (fieldInfo.enum) {
          return this.createSelectElement(fieldName, fieldInfo, hints);
        } else if (hints.input_type === 'email') {
          return this.createEmailInput(fieldName, fieldInfo, hints);
        } else {
          return this.createTextInput(fieldName, fieldInfo, hints);
        }
        
      case 'number':
        return this.createNumberInput(fieldName, fieldInfo, hints);
        
      case 'boolean':
        return this.createCheckboxInput(fieldName, fieldInfo, hints);
        
      case 'array':
        return this.createArrayInput(fieldName, fieldInfo, hints);
        
      case 'object':
        return this.createObjectInput(fieldName, fieldInfo, hints);
        
      default:
        return this.createTextInput(fieldName, fieldInfo, hints);
    }
  }
  
  createTextInput(fieldName, fieldInfo, hints) {
    const input = document.createElement('input');
    input.type = 'text';
    input.name = fieldName;
    input.value = fieldInfo.value || '';
    input.placeholder = hints.placeholder || '';
    input.readonly = fieldInfo.readonly;
    
    return input;
  }
  
  createEmailInput(fieldName, fieldInfo, hints) {
    const input = document.createElement('input');
    input.type = 'email';
    input.name = fieldName;
    input.value = fieldInfo.value || '';
    input.placeholder = hints.placeholder || 'example@domain.com';
    input.readonly = fieldInfo.readonly;
    
    return input;
  }
  
  createNumberInput(fieldName, fieldInfo, hints) {
    const input = document.createElement('input');
    input.type = 'number';
    input.name = fieldName;
    input.value = fieldInfo.value || '';
    input.readonly = fieldInfo.readonly;
    
    if (hints.input_type === 'currency') {
      input.step = '0.01';
      input.min = '0';
    }
    
    return input;
  }
  
  createCheckboxInput(fieldName, fieldInfo, hints) {
    const input = document.createElement('input');
    input.type = 'checkbox';
    input.name = fieldName;
    input.checked = fieldInfo.value || false;
    input.disabled = fieldInfo.readonly;
    
    return input;
  }
  
  createSelectElement(fieldName, fieldInfo, hints) {
    const select = document.createElement('select');
    select.name = fieldName;
    select.disabled = fieldInfo.readonly;
    
    fieldInfo.enum.forEach(option => {
      const optionElement = document.createElement('option');
      optionElement.value = option;
      optionElement.textContent = option;
      optionElement.selected = option === fieldInfo.value;
      select.appendChild(optionElement);
    });
    
    return select;
  }
  
  async handleFormSubmit(event, processInstanceId) {
    event.preventDefault();
    
    const form = event.target;
    const values = this.getFormValues(form);
    
    // Validate form
    const validation = this.validateForm(form);
    if (!validation.valid) {
      this.displayValidationErrors(form, validation.errors);
      return;
    }
    
    try {
      // Update process variables using typed endpoint
      const response = await fetch(
        `/api/v1/processes/${processInstanceId}/typed/variables`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': this.apiKey
          },
          body: JSON.stringify({ variables: values })
        }
      );
      
      if (response.ok) {
        this.showSuccessMessage(form, 'Variables updated successfully');
      } else {
        this.showErrorMessage(form, 'Failed to update variables');
      }
    } catch (error) {
      this.showErrorMessage(form, `Error: ${error.message}`);
    }
  }
}
```

## Связанные endpoints
- [`PUT /api/v1/processes/:id/typed/variables`](./update-process-typed-variables.md) - Обновление типизированных переменных
- [`GET /api/v1/processes/:id/typed/status`](./get-process-typed-status.md) - Типизированный статус процесса
- [`GET /api/v1/processes/:id/info`](../processes/get-process-info.md) - Базовая информация о процессе
