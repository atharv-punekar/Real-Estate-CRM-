# Frontend Changelog

> **For Frontend Team**  
> Track of all API and validation changes that impact the frontend

---

## 2026-02-11 (Update 8)

### 🚀 CSV Import is now SYNCHRONOUS

**Affected Endpoints:**
- `POST /agent/contacts/import`

**What Changed:**
- The API now waits for the import to finish before sending a response.
- **NO POLLING REQUIRED**: You no longer need to call a separate status API.
- The response directly returns the counts of imported and skipped records.

**Response Example:**
```json
{
  "message": "CSV import completed",
  "total_records": 100,
  "imported_records": 85,
  "skipped_records": 15
}
```

---

### 🔒 Removed: Background Job Status API
- `GET /agent/jobs/:id` has been **removed** as it is no longer needed for CSV imports.

---

### 🛡️ Security Check: Stale Organizations

**Affected Endpoints:**
- `POST /agent/contacts/import`

**What Changed:**
- The backend now strictly verifies that the `organization_id` in your JWT token actually exists in the database.
- If your database was reset/wiped but your browser kept an old token, you will now get a clear error instead of a generic database failure.

**New Error Message:**
```json
{"error": "Invalid organization session. Please log in again."}
```

**Action Required:** If you see this error, simply log out and log back in.

---

## 2026-02-11 (Update 9)

### ✅ Relaxed Contact Name Uniqueness

**Affected Endpoints:**
- `POST /agent/contacts`
- `POST /agent/contacts/import`

**What Changed:**
- You can now create multiple contacts with the **same name** (e.g., "John Doe") within an organization.
- **Uniqueness is still enforced on Email and Phone Number.**
- The error `contact with this name already exists in your organization` has been removed.

---

## 2026-02-11 (Update 7)

### ⚠️ Campaign Validation Updates

**Affected Endpoints:**
- `POST /agent/campaigns` - Create campaign
- `PUT /agent/campaigns/:id` - Update campaign

**New Validation Rules:**

#### 1. Campaign Name
- **Alphabets and spaces only** (a-z, A-Z, space)
- **Min 2 characters**
- **No multiple consecutive spaces**
- **Unique within organization** (case-insensitive)

**Error Messages:**
```json
{"error": "name must contain only alphabetic characters and spaces"}
{"error": "a campaign with this name already exists in your organization"}
```

#### 2. Scheduled Date/Time
- **Cannot schedule campaigns in the past** (applies to both create and update)
- Server validates against current time

**Error Message:**
```json
{"error": "Cannot schedule campaign in the past. Please select a future date and time"}
```

---

## 2026-02-11 (Update 6)

### ⚠️ Email Template Validation Updates

**Affected Endpoints:**
- `POST /agent/templates` - Create template
- `PUT /agent/templates/:id` - Update template

**Major Changes:**

#### 1. From Name & Reply To - Now Optional
- ✅ `from_name` is **optional** (can be omitted or null)
- ✅ `reply_to` is **optional** (can be omitted or null)
- Frontend doesn't need to send these fields

**Request Example:**
```json
{
  "name": "Welcome Email",
  "subject": "Welcome!",
  "html_body": "<html>...</html>"
  // from_name and reply_to omitted - no error
}
```

#### 2. Template Name Validation
- **Alphabets and spaces only** (a-z, A-Z, space)
- **Min 2 characters**
- **No multiple consecutive spaces**
- **Unique within organization** (case-insensitive)

**Error Messages:**
```json
{"error": "name must contain only alphabetic characters and spaces"}
{"error": "a template with this name already exists in your organization"}
```

#### 3. Body Validation - Either/Or Required
- **Either** `html_body` **OR** `plain_text_body` must be provided
- Not both can be empty
- Can provide both if needed

**Error Message:**
```json
{"error": "either html_body or plain_text_body must be provided"}
```

---

## 2026-02-11 (Update 5)

### ✅ Cascading Contact Deletion

**Affected Endpoint:**
- `DELETE /agent/contacts/:id`

**Major Change:**
When a contact is deleted, it's now **automatically removed** from:
1. **All audiences** (`audience_contact` entries deleted)
2. **All campaigns** (`contact_id` set to NULL - campaign remains active)

**What This Means:**
- ✅ **No manual cleanup required** - deletion is atomic
- ✅ **Transactional** - all operations succeed or all fail
- ✅ **No orphaned data** - maintains data integrity

**Campaign Behavior:**
```json
// Before contact deletion
{
  "id": "campaign-123",
  "contact_id": "contact-456",
  "audience_ids": ["aud-1"]
}

// After contact deleted
{
  "id": "campaign-123",
  "contact_id": null,  // ← Contact reference removed
  "audience_ids": ["aud-1"]  // ← Audiences unchanged
}
```

**Frontend Changes:**
- ✅ Single `DELETE` call handles everything
- ✅ Refresh audience contact counts after deletion
- ✅ Refresh campaign lists if showing contact_id

**See:** `CASCADING_DELETE.md` for complete details

---

## 2026-02-11 (Update 4)

### ✅ Audience Name Validation

**Affected Endpoints:**
- `POST /agent/audiences` - Create audience
- `PUT /agent/audiences/:id` - Update audience

**New Validation Rules:**
- **Alphabets and spaces only** (a-z, A-Z, space)
- **Min 2 characters**
- **No multiple consecutive spaces**
- **Unique within organization** (case-insensitive)

**Error Messages:**
```json
{"error": "name is required"}
{"error": "name must be at least 2 characters long"}
{"error": "name must contain only alphabetic characters and spaces"}
{"error": "name cannot contain multiple consecutive spaces"}
{"error": "an audience with this name already exists in your organization"}
```

**JavaScript Validation:**
```javascript
function validateAudienceName(name) {
  const errors = [];
  
  if (!name || !name.trim()) {
    return 'Name is required';
  }
  
  const trimmed = name.trim();
  
  if (trimmed.length < 2) {
    return 'Name must be at least 2 characters long';
  }
  
  // Alphabets and spaces only
  if (!/^[a-zA-Z ]+$/.test(trimmed)) {
    return 'Name must contain only alphabetic characters and spaces';
  }
  
  // No multiple consecutive spaces
  if (/  /.test(trimmed)) {
    return 'Name cannot contain multiple consecutive spaces';
  }
  
  return null;
}
```

---

## 2026-02-11 (Update 3)

### ⚠️ STRICT VALIDATION RULES FOR CONTACT FORMS

**Affected Endpoints:**
- `POST /agent/contacts` - Create contact
- `PUT /agent/contacts/:id` - Update contact
- `POST /agent/contacts/import` - CSV import

---

#### 1. Name Fields (BREAKING)
**First Name & Last Name:**
- **Alphabets ONLY** (a-z, A-Z)
- **Min 2 characters**
- ❌ **NO** numbers, spaces, hyphens, apostrophes, or special characters

**What Changed:**
```
❌ John-Smith  → Must use JohnSmith
❌ O'Brien     → Must use OBrien
❌ Mary Anne   → Must use MaryAnne or Mary (pick one)
❌ José        → Must use Jose
```

**Error Messages:**
```json
{"error": "first_name must contain only alphabetic characters"}
{"error": "last_name must be at least 2 characters long"}
```

---

#### 2. Email Validation (ENHANCED)
**New:** Regex validation required
- Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- **Exactly one @ symbol**
- **Must have domain extension** (.com, .org, etc.)

**Error Messages:**
```json
{"error": "email must be a valid format (e.g., user@example.com)"}
{"error": "email must contain exactly one @ symbol"}
```

---

#### 3. Phone Number (STRICTER)
**Changed:** NO special characters allowed in input
- **Exactly 10 digits** (digits only)
- ❌ Cannot accept: `-`, `()`, `+`, spaces

**Before:** `(987) 654-3210` was accepted  
**Now:** Only `9876543210` accepted

**Frontend Must:**
- Strip all non-digit characters before sending to API
- Or block non-digit input entirely

---

#### 4. Bedrooms & Bathrooms (LIMITS ADDED)
**New:** Max value is 5

**Values allowed:** 0, 1, 2, 3, 4, 5

**Error Messages:**
```json
{"error": "bedrooms cannot exceed 5"}
{"error": "bathrooms cannot exceed 5"}
```

**⚠️ REQUIRED:** Use dropdown/select input with options 0-5

---

#### 5. Budget Validation (ENHANCED)
**New:** Max must be >= Min

**Error Messages:**
```json
{"error": "maximum budget must be greater than or equal to minimum budget"}
```

---

## Required Frontend Changes

### 1. **Update Input Fields**
- First/Last Name: Pattern `[a-zA-Z]{2,}`, block non-alphabetic input
- Email: Regex validation with @ and domain
- Phone: Input type="tel", maxLength="10", pattern="\d{10}"

### 2. **Add Dropdowns**
```html
<!-- Bedrooms -->
<select name="bedrooms" required>
  <option value="0">Studio (0)</option>
  <option value="1">1 Bedroom</option>
  <option value="2">2 Bedrooms</option>
  <option value="3">3 Bedrooms</option>
  <option value="4">4 Bedrooms</option>
  <option value="5">5+ Bedrooms</option>
</select>

<!-- Bathrooms -->
<select name="bathrooms" required>
  <option value="0">0 Bathrooms</option>
  <option value="1">1 Bathroom</option>
  <option value="2">2 Bathrooms</option>
  <option value="3">3 Bathrooms</option>
  <option value="4">4 Bathrooms</option>
  <option value="5">5+ Bathrooms</option>
</select>
```

### 3. **Input Validation**
```javascript
// Name - Block non-alphabetic
function validateNameInput(e) {
  e.target.value = e.target.value.replace(/[^a-zA-Z]/g, '');
}

// Phone - Digits only
function validatePhoneInput(e) {
  const digits = e.target.value.replace(/\D/g, '');
  e.target.value = digits.substring(0, 10);
}

// Email - Regex
function validateEmail(email) {
  const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  return regex.test(email) && (email.match(/@/g) || []).length === 1;
}
```

---

**See:** `CONTACT_FORM_VALIDATION.md` for complete JavaScript validation code

---

## 2026-02-11 (Update 2)

### ⚠️ BREAKING CHANGES: Contact Validation

**Affected Endpoints:**
- `POST /agent/contacts` - Create contact
- `PUT /agent/contacts/:id` - Update contact  
- `POST /agent/contacts/import` - CSV import

**Major Changes:**

#### 1. All Fields Now Required
All contact fields are **required** and cannot be empty:
- `first_name`, `last_name`, `email`, `phone`
- `property_type`, `preferred_location`
- `budget_min`, `budget_max`
- `bedrooms`, `bathrooms`, `square_feet`

**Error Examples:**
```json
{"error": "first_name is required"}
{"error": "phone is required"}
```

#### 2. Phone Number Validation (BREAKING)
- Must be **exactly 10 digits** 
- Non-digit characters (`-`, `()`, `+`, spaces) are removed during validation
- After cleaning, must contain only digits 0-9

**Valid:**
```
✅ 9876543210
✅ (987) 654-3210
✅ +91 9876543210
```

**Invalid:**
```
❌ 12345 (too short)
❌ 123456789012 (too long)
```

**Error Examples:**
```json
{"error": "phone number must contain only digits"}
{"error": "phone number must be exactly 10 digits"}
```

#### 3. Duplicate Detection Enhanced
Checks for duplicates using **email OR name**:
- Email exists → Error
- Full name (first + last) exists → Error

**Error Examples:**
```json
{"error": "contact with this email already exists in your organization"}
{"error": "contact with this name already exists in your organization"}
```

#### 4. Notes Field Removed (BREAKING)
- ❌ `notes` field has been **removed** from contact model
- Update forms to remove this field
- CSV templates should not include `notes` column

---

### CSV Import Changes

**Required CSV Headers:**
All these columns are **required** (notes removed):
- `first_name`, `last_name`, `email`, `phone`
- `budget_min`, `budget_max`, `property_type`
- `bedrooms`, `bathrooms`, `square_feet`, `preferred_location`

**Row-Level Validation:**
- All fields must be filled (no empty values)
- Phone validation same as single contact
- Detailed error messages with row numbers

**Error Example:**
```json
{
  "error": "CSV contains validation errors:\nRow 2: phone is required\nRow 3: phone must be exactly 10 digits\nRow 5: budget_min cannot be greater than budget_max"
}
```

**Duplicate Handling:**
Rows with duplicate email OR name are **skipped** (not error):
```
Row 3: Skipped - duplicate email: john@example.com
Row 5: Skipped - duplicate name: Jane Smith
```

---

**See:** `CONTACT_VALIDATION.md` for complete documentation

---

## 2026-02-11

### ✅ Pagination Added to All List Endpoints

**Affected Endpoints:**
- `GET /superadmin/orgs` - Organization list
- `GET /orgadmin/agents` - Agent list (org admin)
- `GET /superadmin/orgs/:org_id/agents` - Agent list (super admin)

**Changes:**
- All endpoints now support `page` and `limit` query parameters
- Default: `page=1`, `limit=20`
- Max limit: 100
- Response includes full pagination metadata

**Response Format:**
```json
{
  "organizations": [...],  // or "agents"
  "total_count": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5,
  "offset_start": 1,
  "offset_end": 20
}
```

**See:** `API_PAGINATION.md` for complete documentation

---

### ✅ Validation Rules Added

#### Organization Name
**Endpoint:** `POST /superadmin/orgs`

**New Rules:**
- Only alphabetic characters (a-z, A-Z) and spaces
- No numbers or special characters
- No multiple consecutive spaces
- Must be unique (case-insensitive)

**Error Examples:**
```json
{"error": "organization name can only contain alphabetic characters and spaces"}
{"error": "organization name cannot contain multiple consecutive spaces"}
{"error": "An organization with this name already exists"}
```

---

#### Agent Name
**Endpoints:** 
- `POST /orgadmin/agents`
- `PUT /orgadmin/agents/:id`
- `POST /superadmin/orgs/:org_id/agents`
- `PUT /superadmin/orgs/:org_id/agents/:id`

**Rules:**
- Only alphabetic characters (a-z, A-Z)
- Exactly one space (between first and last name)
- Min 2 characters for each part

**Error Examples:**
```json
{"error": "name must contain exactly one space between first name and surname"}
{"error": "first name must be at least 2 characters long"}
{"error": "name must contain only alphabetic characters"}
```

---

#### Agent Email
**Endpoints:** Same as Agent Name

**Rules:**
- Local part: only lowercase letters (a-z) and numbers (0-9)
- Must have exactly one @ symbol
- Valid domain with extension

**Error Examples:**
```json
{"error": "email local part must contain only lowercase letters and numbers"}
{"error": "email domain must be valid (e.g., example.com)"}
```

---

#### Password
**Endpoint:** `POST /auth/activate`

**Rules:**
- Min 6, Max 20 characters
- Must contain: lowercase letter, digit, special character
- Special chars: `!@#$%^&*()_+-=[]{}|;:,.<>?/~\``

**Error Examples:**
```json
{"error": "password must be at least 6 characters long"}
{"error": "password must not exceed 20 characters"}
{"error": "password must contain at least: one lowercase letter, one numeric digit"}
```

**See:** `API_VALIDATION_CHANGES.md` for complete details

---

## Change Log Format

Future changes will be added in this format:

### Date: YYYY-MM-DD

**Summary:** Brief description

**Affected Endpoints:**
- List of endpoints

**Changes:**
- What changed
- Breaking changes (if any)

**Migration Required:**
- [ ] Yes - Instructions below
- [x] No

---

## Notes

- **Breaking Changes** will be marked with ⚠️
- **New Features** will be marked with ✅
- **Bug Fixes** will be marked with 🐛
- **Deprecations** will be marked with 🔔
