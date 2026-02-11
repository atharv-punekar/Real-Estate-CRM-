# API Validation Changes

> **For Frontend Team** - New validation rules and error responses

---

## Changed Endpoints

### 1. Create Organization
**Endpoint:** `POST /superadmin/orgs`

**Name Validation:**
- Cannot be empty
- Only alphabetic characters (a-z, A-Z) and spaces allowed
- No numbers or special characters
- No multiple consecutive spaces
- Must be unique (case-insensitive)

**Error Response Examples:**
```json
{
  "error": "organization name cannot be empty"
}
```
```json
{
  "error": "organization name can only contain alphabetic characters and spaces"
}
```
```json
{
  "error": "organization name cannot contain multiple consecutive spaces"
}
```
```json
{
  "error": "An organization with this name already exists"
}
```

---

### 2. Password Activation
**Endpoint:** `POST /auth/activate`

**New Validation:**
- Min 6, Max 20 characters
- Must contain: lowercase letter, digit, special character

**Error Response Examples:**
```json
{
  "error": "password must be at least 6 characters long"
}
```
```json
{
  "error": "password must not exceed 20 characters"
}
```
```json
{
  "error": "password must contain at least: one lowercase letter, one numeric digit"
}
```

---

### 3. Create/Update Agent (Org Admin)

**Endpoints:** 
- `POST /orgadmin/agents`
- `PUT /orgadmin/agents/:id`

**Name Validation:**
- Only alphabetic characters (a-z, A-Z)
- Exactly one space (between first and last name)
- Min 2 characters for each part

**Email Validation:**
- Local part: only lowercase letters and numbers
- Must have valid domain with extension

**Error Response Examples:**
```json
{
  "error": "name must contain exactly one space between first name and surname"
}
```
```json
{
  "error": "first name must be at least 2 characters long"
}
```
```json
{
  "error": "name must contain only alphabetic characters"
}
```
```json
{
  "error": "email local part must contain only lowercase letters and numbers"
}
```
```json
{
  "error": "email domain must be valid (e.g., example.com)"
}
```

---

### 4. Create/Update Agent (Super Admin)
**Endpoints:** 
- `POST /superadmin/orgs/:org_id/agents`
- `PUT /superadmin/orgs/:org_id/agents/:id`

**Same validation rules as Org Admin endpoints above**

---

## Valid Examples

### Organization Name
```
✅ Amar Business Group
✅ Tech Solutions Inc
✅ Real Estate Company
❌ ""                    (empty)
❌ "  "                  (only spaces)
❌ Tech123 Solutions      (contains numbers)
❌ ABC@Company           (special character)
❌ Company  Name         (multiple consecutive spaces)
```

### Name
```
✅ John Doe
✅ Jane Smith
❌ john          (missing last name)
❌ John-Doe      (special character)
❌ J Smith       (first name too short)
```

### Email
```
✅ john@example.com
✅ user123@company.org
❌ John@example.com     (uppercase)
❌ john.doe@example.com (dot in local part)
❌ john@Example.com     (uppercase in domain)
```

### Password
```
✅ Pass123!
✅ myP@ss1
❌ abc12        (too short, no special char)
❌ Password!    (no digit)
```

---

## HTTP Status Codes

| Code | When |
|------|------|
| `400` | Validation error (format invalid) |
| `409` | Email already exists |
| `409` | Name already exists in organization |
