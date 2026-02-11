# Contact API Validation Rules

> **For Frontend Team** - Complete validation rules for contact creation and CSV import

---

## Single Contact Creation

### Endpoint
`POST /agent/contacts`

### Required Fields
All fields below are **REQUIRED** and cannot be empty:

| Field | Type | Validation |
|-------|------|------------|
| `first_name` | string | Cannot be empty |
| `last_name` | string | Cannot be empty |
| `email` | string | Cannot be empty, must be valid email format |
| `phone` | string | **Exactly 10 digits** (non-digit characters like `-`, `()`, `+`, spaces are allowed but stripped) |
| `budget_min` | number | Cannot be empty, must be ≥ 0 |
| `budget_max` | number | Cannot be empty, must be ≥ 0, must be ≥ budget_min |
| `property_type` | string | Cannot be empty |
| `bedrooms` | integer | Cannot be empty, must be ≥ 0 |
| `bathrooms` | integer | Cannot be empty, must be ≥ 0 |
| `square_feet` | integer | Cannot be empty, must be ≥ 0 |
| `preferred_location` | string | Cannot be empty |

### Phone Number Validation
- Must be **exactly 10 digits**
- Non-digit characters (`-`, `()`, `+`, spaces) are removed during validation
- After removing non-digit characters, must contain only 0-9
- Must be exactly 10 digits after cleaning

**Valid Examples:**
```
✅ 9876543210
✅ (987) 654-3210
✅ 987-654-3210
✅ +91 9876543210
```

**Invalid Examples:**
```
❌ 12345         (too few digits)
❌ 123456789012  (too many digits)
❌ 98765abcde    (contains letters)
❌ ""            (empty)
```

### Duplicate Detection
- **Email**: Contact with same email in organization = **ERROR**
- **Name**: Contact with same first_name + last_name in organization = **ERROR**

### Error Responses

#### Required Fields
```json
{"error": "first_name is required"}
{"error": "last_name is required"}
{"error": "email is required"}
{"error": "phone is required"}
{"error": "property_type is required"}
{"error": "preferred_location is required"}
{"error": "budget_min is required"}
{"error": "budget_max is required"}
{"error": "bedrooms is required"}
{"error": "bathrooms is required"}
{"error": "square_feet is required"}
```

#### Phone Validation
```json
{"error": "phone number must contain only digits"}
{"error": "phone number must be exactly 10 digits"}
```

#### Value Validation
```json
{"error": "minimum budget cannot be negative"}
{"error": "maximum budget cannot be negative"}
{"error": "budget_min cannot be greater than budget_max"}
{"error": "bedrooms cannot be negative"}
{"error": "bathrooms cannot be negative"}
{"error": "square_feet cannot be negative"}
```

#### Duplicate Errors
```json
{"error": "contact with this email already exists in your organization"}
{"error": "contact with this name already exists in your organization"}
```

---

## CSV Import

### Endpoint
`POST /agent/contacts/import`

### CSV Requirements

#### File Validation
- Must be CSV file (`.csv` extension or `text/csv` content-type)

#### Required Headers (case-insensitive)
The CSV **MUST** contain all these columns:
- `first_name`
- `last_name`
- `email`
- `phone`
- `budget_min`
- `budget_max`
- `property_type`
- `bedrooms`
- `bathrooms`
- `square_feet`
- `preferred_location`

❌ **Removed Fields**: `notes` column is NO LONGER supported

#### Sample CSV Template
```csv
first_name,last_name,email,phone,budget_min,budget_max,property_type,bedrooms,bathrooms,square_feet,preferred_location
John,Doe,john@example.com,9876543210,300000,500000,Apartment,3,2,1500,Downtown
Jane,Smith,jane@example.com,9123456789,400000,600000,House,4,3,2000,Suburbs
```

### Row-Level Validation

#### All Fields Required
Every row must have **all fields filled**. Empty fields = validation error for that row.

**❌ CRITICAL: Rows with ANY empty field will be REJECTED**

If a CSV row has 1 or more empty fields, that row will:
- ❌ **NOT be imported** (skipped)
- 📝 **Show in error message** with row number and missing fields
- ⚠️ **Not affect other rows** - valid rows still import

**Example:**
```csv
first_name,last_name,email,phone,budget_min,budget_max,property_type,bedrooms,bathrooms,square_feet,preferred_location
John,Doe,john@example.com,9876543210,300000,500000,Apartment,3,2,1500,Downtown
Jane,Smith,,9123456789,400000,600000,House,4,3,2000,Suburbs
Bob,Johnson,bob@example.com,,350000,450000,Condo,2,1,1200,
```

**Result:**
```json
{
  "error": "CSV contains validation errors:
Row 3: email is required
Row 4: phone is required; preferred_location is required"
}
```

- ✅ Row 2 (John) - Imported successfully
- ❌ Row 3 (Jane) - Skipped (email empty)
- ❌ Row 4 (Bob) - Skipped (phone and location empty)

**Final Result:** 1 contact imported, 2 rows skipped

#### Phone Number Validation
- Same rules as single contact creation
- Must be exactly 10 digits after removing non-digit characters

#### Value Validation
- All numeric fields must be valid numbers
- `budget_min`, `budget_max`, `bedrooms`, `bathrooms`, `square_feet` cannot be negative
- `budget_min` ≤ `budget_max`

### Error Handling

#### CSV Structure Errors
```json
{"error": "CSV must contain 'first_name' column"}
{"error": "CSV must contain 'phone' column"}
```

#### Row-Level Errors
When validation fails, you'll get detailed error messages:

```json
{
  "error": "CSV contains validation errors:\nRow 2: first_name is required; phone is required\nRow 3: phone must be exactly 10 digits\nRow 5: budget_min cannot be greater than budget_max"
}
```

- Row numbers start from 2 (Row 1 is the header)
- Multiple errors per row are separated by `;`
- Max 10 error rows shown to avoid overwhelming response

### Duplicate Handling During Import

Contacts are checked for duplicates **before** creation:

| Duplicate Type | Action |
|----------------|--------|
| **Email exists** | Skip row, log message |
| **Name exists (first + last)** | Skip row, log message |

**Console Output:**
```
Row 3: Skipped - duplicate email: john@example.com
Row 5: Skipped - duplicate name: Jane Smith
Created 10 contacts so far...
Bulk create completed: 8 success, 2 skipped
```

### Import Process

1. **Upload** → Returns `202 Accepted` with `job_id`
2. **Background Processing** → Validates all rows
3. **Duplicate Check** → Skips duplicates (email OR name)
4. **Creation** → Creates valid contacts
5. **Notification** → User gets notification with success/skip counts

---

## Frontend Implementation

### ⚠️ Important: CSV Validation Behavior

**Before upload, warn users:**
- All 11 fields are **mandatory** in CSV
- Rows with **any empty field** will be **rejected**
- Only fully complete rows will be imported
- Users will see detailed error messages for invalid rows

**Recommended UI:**
```
⚠️ CSV Import Requirements:
• All 11 fields must be filled for each contact
• Rows with missing data will be skipped
• You'll receive a detailed error report for invalid rows
```

### Contact Creation Form Validation

```javascript
function validateContactForm(data) {
  const errors = {};
  
  // Required fields
  if (!data.first_name?.trim()) errors.first_name = 'First name is required';
  if (!data.last_name?.trim()) errors.last_name = 'Last name is required';
  if (!data.email?.trim()) errors.email = 'Email is required';
  if (!data.phone?.trim()) errors.phone = 'Phone is required';
  if (!data.property_type?.trim()) errors.property_type = 'Property type is required';
  if (!data.preferred_location?.trim()) errors.preferred_location = 'Preferred location is required';
  
  // Numeric required fields
  if (data.budget_min === undefined || data.budget_min === null || data.budget_min === '') {
    errors.budget_min = 'Minimum budget is required';
  }
  if (data.budget_max === undefined || data.budget_max === null || data.budget_max === '') {
    errors.budget_max = 'Maximum budget is required';
  }
  if (data.bedrooms === undefined || data.bedrooms === null || data.bedrooms === '') {
    errors.bedrooms = 'Bedrooms is required';
  }
  if (data.bathrooms === undefined || data.bathrooms === null || data.bathrooms === '') {
    errors.bathrooms = 'Bathrooms is required';
  }
  if (data.square_feet === undefined || data.square_feet === null || data.square_feet === '') {
    errors.square_feet = 'Square feet is required';
  }
  
  // Phone validation (10 digits)
  if (data.phone) {
    const cleaned = data.phone.replace(/[\s\-()+ ]/g, '');
    if (!/^\d+$/.test(cleaned)) {
      errors.phone = 'Phone must contain only digits';
    } else if (cleaned.length !== 10) {
      errors.phone = 'Phone must be exactly 10 digits';
    }
  }
  
  // Negative validation
  if (data.budget_min < 0) errors.budget_min = 'Cannot be negative';
  if (data.budget_max < 0) errors.budget_max = 'Cannot be negative';
  if (data.bedrooms < 0) errors.bedrooms = 'Cannot be negative';
  if (data.bathrooms < 0) errors.bathrooms = 'Cannot be negative';
  if (data.square_feet < 0) errors.square_feet = 'Cannot be negative';
  
  // Budget range
  if (data.budget_min > data.budget_max) {
    errors.budget_min = 'Min budget cannot exceed max budget';
  }
  
  return Object.keys(errors).length > 0 ? errors : null;
}
```

### CSV File Validation

```javascript
function validateCSVBeforeUpload(file) {
  // Check file type
  if (!file.name.endsWith('.csv') && file.type !== 'text/csv') {
    return 'Please select a CSV file';
  }
  
  // Check file size (max 5MB)
  if (file.size > 5 * 1024 * 1024) {
    return 'File size must be less than 5MB';
  }
  
  return null;
}
```

### Phone Input Formatting

```javascript
function formatPhoneInput(value) {
  // Remove all non-digits
  const cleaned = value.replace(/\D/g, '');
  
  // Limit to 10 digits
  const limited = cleaned.substring(0, 10);
  
  // Format as (XXX) XXX-XXXX
  if (limited.length >= 6) {
    return `(${limited.slice(0, 3)}) ${limited.slice(3, 6)}-${limited.slice(6)}`;
  } else if (limited.length >= 3) {
    return `(${limited.slice(0, 3)}) ${limited.slice(3)}`;
  } else {
    return limited;
  }
}
```

---

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| `201` | Contact created successfully |
| `202` | CSV import started (async) |
| `400` | Validation error or bad request |
| `404` | Contact not found |
| `500` | Server error |

---

## Summary of Changes

### ✅ What Changed
1. **All fields are now required** (cannot be empty)
2. **Phone must be exactly 10 digits** (non-digit chars are stripped)
3. **Duplicate check for email OR name** (not just email)
4. **Notes field removed** (no longer supported)
5. **CSV validation is strict** - all fields required per row
6. **Detailed CSV error messages** - shows row numbers and specific errors

### ⚠️ Breaking Changes
- **Notes field removed**: Remove from your forms and CSV templates
- **All fields required**: Update forms to mark all fields as required
- **Phone format**: Must be exactly 10 digits (not 11 or country code)
