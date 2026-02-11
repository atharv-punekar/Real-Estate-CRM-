# CSV Import Validation Rules

> **For Frontend Team** - Contact CSV import validation

---

## Endpoint
`POST /agent/contacts/import`

---

## File Requirements

### File Type
- Must be CSV file
- Accepts: `text/csv` or `application/vnd.ms-excel` content types
- OR filename must end with `.csv`

### Error Response
```json
{
  "error": "No file uploaded"
}
```
```json
{
  "error": "File must be a CSV"
}
```

---

## CSV Format

### Required Headers
The CSV must have specific column headers (case-sensitive):
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
- `notes`

### Sample CSV
```csv
first_name,last_name,email,phone,budget_min,budget_max,property_type,bedrooms,bathrooms,square_feet,preferred_location,notes
John,Doe,john@example.com,+1234567890,300000,500000,Apartment,3,2,1500,Downtown,Looking for modern apartment
Jane,Smith,jane@example.com,+1234567891,400000,600000,House,4,3,2000,Suburbs,Needs backyard
```

---

## Field Validations

### Per-Contact Validations

| Field | Validation | Error if Invalid |
|-------|------------|------------------|
| `budget_min` | Must be non-negative | "Minimum budget cannot be negative" |
| `budget_max` | Must be non-negative | "Maximum budget cannot be negative" |
| `bedrooms` | Must be non-negative integer | "Bedrooms cannot be negative" |
| `bathrooms` | Must be non-negative integer | "Bathrooms cannot be negative" |
| `square_feet` | Must be non-negative integer | "Square feet cannot be negative" |
| `budget_min` vs `budget_max` | Min must be ≤ Max | "minimum budget cannot be greater than maximum budget" |
| `email` | Must be valid email format | Email validation error |

### Duplicates Handling
- Contacts with **duplicate emails** within the **same organization** are **skipped**
- No error thrown, just skipped from import
- Import continues for remaining valid contacts

---

## Import Process

### 1. Upload Response (202 Accepted)
Import starts asynchronously in the background.

```json
{
  "message": "CSV import started",
  "job_id": "uuid"
}
```

### 2. Background Processing
- CSV is parsed
- Each contact is validated
- Valid contacts are created
- Invalid contacts are skipped
- Duplicates (by email) are skipped

### 3. Completion Notification
User receives a notification when import completes:
- Success count
- Skip count (duplicates + invalid)

---

## Response Codes

| Code | Meaning |
|------|---------|
| `202` | Import started successfully (async processing) |
| `400` | No file uploaded or invalid file type |
| `500` | Failed to process file or create job |

---

## Error Scenarios

### File Upload Errors
```json
{"error": "No file uploaded"}
{"error": "File must be a CSV"}
{"error": "Failed to read file"}
```

### CSV Parsing Errors
```json
{"error": "Failed to parse CSV: invalid header"}
{"error": "Failed to parse CSV: missing required columns"}
```

### Processing Errors
- Invalid data in CSV rows are **skipped**, not fail entire import
- User is notified of success count vs skip count
- Import job status can be tracked via job_id

---

## Frontend Implementation

### File Upload
```javascript
async function importContactsCSV(file) {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch('/agent/contacts/import', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
    },
    body: formData
  });
  
  return response.json();
}
```

### Validation Before Upload
```javascript
function validateCSVFile(file) {
  // Check file type
  if (!file.name.endsWith('.csv') && 
      file.type !== 'text/csv' && 
      file.type !== 'application/vnd.ms-excel') {
    return 'Please select a CSV file';
  }
  
  // Check file size (optional, e.g., max 5MB)
  if (file.size > 5 * 1024 * 1024) {
    return 'File size must be less than 5MB';
  }
  
  return null; // Valid
}
```

### Track Import Progress
```javascript
// Poll job status (if endpoint exists)
async function checkImportStatus(jobId) {
  const response = await fetch(`/agent/jobs/${jobId}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
}

// Or listen for notifications
// The backend sends notification when import completes
```

---

## Notes for Frontend

1. **File Validation**: Validate CSV file client-side before upload
2. **Async Process**: Import runs in background, show loading state
3. **Notifications**: Listen for completion notification
4. **Error Handling**: Handle both upload errors and processing errors
5. **Template Download**: Provide a sample CSV template for users
6. **Preview**: Consider showing CSV preview before upload
7. **Duplicate Warning**: Warn users that duplicate emails will be skipped

---

## Sample CSV Template

Provide this template for download:

```csv
first_name,last_name,email,phone,budget_min,budget_max,property_type,bedrooms,bathrooms,square_feet,preferred_location,notes
John,Doe,john@example.com,+1234567890,300000,500000,Apartment,3,2,1500,Downtown,Looking for modern apartment
```
