# Contact Form Validation - Frontend Guide

> **UPDATED 2026-02-11** - New validation rules for contact forms

---

## ⚠️ CRITICAL VALIDATION RULES

### 1. First Name & Last Name
- **Alphabets ONLY** (a-z, A-Z)
- **Min 2 characters**
- ❌ No numbers, spaces, or special characters

**Valid:**
```
✅ John
✅ Sarah
✅ Mohammed
```

**Invalid:**
```
❌ J (too short)
❌ John123 (contains numbers)
❌ O'Brien (contains special character)
❌ Mary-Anne (contains hyphen)
```

**Error Messages:**
```json
{"error": "first_name must be at least 2 characters long"}
{"error": "first_name must contain only alphabetic characters"}
{"error": "last_name must contain only alphabetic characters"}
```

---

### 2. Email
- **Must use regex validation**: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- **Exactly one @ symbol**
- **Must have domain with extension** (e.g., .com, .org)

**Valid:**
```
✅ john@example.com
✅ user.name@company.co.uk
✅ test123@domain.org
```

**Invalid:**
```
❌ john@example (no extension)
❌ @example.com (no local part)
❌ john@@example.com (multiple @)
❌ john (no @)
```

**Error Messages:**
```json
{"error": "email must be a valid format (e.g., user@example.com)"}
{"error": "email must contain exactly one @ symbol"}
```

---

### 3. Phone Number
- **Exactly 10 digits**
- ❌ **NO special characters** (no `-`, `()`, `+`, spaces)
- Only digits 0-9 allowed

**Valid:**
```
✅ 9876543210
```

**Invalid:**
```
❌ (987) 654-3210 (has special chars)
❌ +91 9876543210 (has + and space)
❌ 987-654-3210 (has hyphens)
❌ 12345 (too short)
```

**Error Messages:**
```json
{"error": "phone number must contain only digits"}
{"error": "phone number must be exactly 10 digits"}
```

---

### 4. Budget Min & Max
- **Cannot be negative**
-  **Max must be >= Min**

**Error Messages:**
```json
{"error": "minimum budget cannot be negative"}
{"error": "maximum budget must be greater than or equal to minimum budget"}
```

**Recommended:** Use number input or slider UI

---

### 5. Bedrooms (DROPDOWN: 0-5)
- **Values:** 0, 1, 2, 3, 4, 5
- Use dropdown or select input

**HTML Example:**
```html
<select name="bedrooms" required>
  <option value="">Select bedrooms</option>
  <option value="0">Studio (0)</option>
  <option value="1">1 Bedroom</option>
  <option value="2">2 Bedrooms</option>
  <option value="3">3 Bedrooms</option>
  <option value="4">4 Bedrooms</option>
  <option value="5">5+ Bedrooms</option>
</select>
```

**Error Messages:**
```json
{"error": "bedrooms cannot exceed 5"}
```

---

### 6. Bathrooms (DROPDOWN: 0-5)
- **Values:** 0, 1, 2, 3, 4, 5
- Use dropdown or select input

**HTML Example:**
```html
<select name="bathrooms" required>
  <option value="">Select bathrooms</option>
  <option value="0">0 Bathrooms</option>
  <option value="1">1 Bathroom</option>
  <option value="2">2 Bathrooms</option>
  <option value="3">3 Bathrooms</option>
  <option value="4">4 Bathrooms</option>
  <option value="5">5+ Bathrooms</option>
</select>
```

---

### 7. Square Feet (OPTIONAL DROPDOWN)
- **Cannot be negative**
- **Recommended:** Provide preset ranges in dropdown

**HTML Example:**
```html
<select name="square_feet" required>
  <option value="">Select size</option>
  <option value="500">Under 500 sqft</option>
  <option value="750">500-1000 sqft</option>
  <option value="1250">1000-1500 sqft</option>
  <option value="1750">1500-2000 sqft</option>
  <option value="2500">2000-3000 sqft</option>
  <option value="4000">3000+ sqft</option>
</select>
```

Or use number input:
```html
<input type="number" name="square_feet" min="0" required>
```

---

## JavaScript Validation

### Complete Form Validation Function

```javascript
function validateContactForm(data) {
  const errors = {};
  
  // First Name - alphabets only, min 2 chars
  if (!data.first_name?.trim()) {
    errors.first_name = 'First name is required';
  } else if (data.first_name.trim().length < 2) {
    errors.first_name = 'First name must be at least 2 characters long';
  } else if (!/^[a-zA-Z]+$/.test(data.first_name.trim())) {
    errors.first_name = 'First name must contain only alphabetic characters';
  }
  
  // Last Name - alphabets only, min 2 chars
  if (!data.last_name?.trim()) {
    errors.last_name = 'Last name is required';
  } else if (data.last_name.trim().length < 2) {
    errors.last_name = 'Last name must be at least 2 characters long';
  } else if (!/^[a-zA-Z]+$/.test(data.last_name.trim())) {
    errors.last_name = 'Last name must contain only alphabetic characters';
  }
  
  // Email - regex validation with @ and domain
  if (!data.email?.trim()) {
    errors.email = 'Email is required';
  } else {
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(data.email.trim())) {
      errors.email = 'Email must be a valid format (e.g., user@example.com)';
    } else if ((data.email.match(/@/g) || []).length !== 1) {
      errors.email = 'Email must contain exactly one @ symbol';
    }
  }
  
  // Phone - exactly 10 digits, no special chars
  if (!data.phone?.trim()) {
    errors.phone = 'Phone is required';
  } else {
    const phoneDigits = data.phone.replace(/\D/g, '');
    if (phoneDigits.length !== 10) {
      errors.phone = 'Phone must be exactly 10 digits';
    } else if (!/^\d+$/.test(phoneDigits)) {
      errors.phone = 'Phone must contain only digits';
    }
  }
  
  // Property Type
  if (!data.property_type?.trim()) {
    errors.property_type = 'Property type is required';
  }
  
  // Preferred Location
  if (!data.preferred_location?.trim()) {
    errors.preferred_location = 'Preferred location is required';
  }
  
  // Budget validation
  if (data.budget_min === undefined || data.budget_min === null || data.budget_min === '') {
    errors.budget_min = 'Minimum budget is required';
  } else if (data.budget_min < 0) {
    errors.budget_min = 'Minimum budget cannot be negative';
  }
  
  if (data.budget_max === undefined || data.budget_max === null || data.budget_max === '') {
    errors.budget_max = 'Maximum budget is required';
  } else if (data.budget_max < 0) {
    errors.budget_max = 'Maximum budget cannot be negative';
  }
  
  // Budget range check
  if (data.budget_min > 0 && data.budget_max > 0 && data.budget_min > data.budget_max) {
    errors.budget_min = 'Min budget cannot exceed max budget';
  }
  
  // Bedrooms (0-5)
  if (data.bedrooms === undefined || data.bedrooms === null || data.bedrooms === '') {
    errors.bedrooms = 'Bedrooms is required';
  } else if (data.bedrooms < 0 || data.bedrooms > 5) {
    errors.bedrooms = 'Bedrooms must be between 0 and 5';
  }
  
  // Bathrooms (0-5)
  if (data.bathrooms === undefined || data.bathrooms === null || data.bathrooms === '') {
    errors.bathrooms = 'Bathrooms is required';
  } else if (data.bathrooms < 0 || data.bathrooms > 5) {
    errors.bathrooms = 'Bathrooms must be between 0 and 5';
  }
  
  // Square Feet
  if (data.square_feet === undefined || data.square_feet === null || data.square_feet === '') {
    errors.square_feet = 'Square feet is required';
  } else if (data.square_feet < 0) {
    errors.square_feet = 'Square feet cannot be negative';
  }
  
  return Object.keys(errors).length > 0 ? errors : null;
}
```

---

## Real-Time Input Validation

### Name Input (Alphabets Only)

```javascript
function validateNameInput(event) {
  // Allow only alphabetic characters
  event.target.value = event.target.value.replace(/[^a-zA-Z]/g, '');
}

// Usage
<input 
  type="text" 
  name="first_name" 
  onInput={validateNameInput}
  pattern="[a-zA-Z]{2,}"
  required 
/>
```

### Phone Input (Digits Only)

```javascript
function validatePhoneInput(event) {
  // Remove all non-digits
  const digits = event.target.value.replace(/\D/g, '');
  // Limit to 10 digits
  event.target.value = digits.substring(0, 10);
}

// Usage
<input 
  type="tel" 
  name="phone" 
  onInput={validatePhoneInput}
  maxLength="10"
  pattern="\d{10}"
  required 
/>
```

---

## Summary of Changes

| Field | Old Rule | New Rule |
|-------|----------|----------|
| **First/Last Name** | Any characters | **Alphabets ONLY (a-z, A-Z), min 2 chars** |
| **Email** | Basic check | **Regex with @ and domain required** |
| **Phone** | 10 digits (with formatting) | **Exactly 10 digits, NO special chars** |
| **Bedrooms** | Any number | **0-5 only (use dropdown)** |
| **Bathrooms** | Any number | **0-5 only (use dropdown)** |
| **Budget Max** | No check | **Must be >= Budget Min** |

---

## ⚠️ BREAKING CHANGES

1. **Names**: Users can NO LONGER enter:
   - Hyphens (Mary-Anne) → Must use MaryAnne
   - Apostrophes (O'Brien) → Must use OBrien
   - Spaces in compound names → Must use single word

2. **Phone**: Users MUST enter plain digits:
   - ❌ (987) 654-3210
   - ✅ 9876543210

3. **Bedrooms/Bathrooms**: Max value is 5
   - For 6+ bedrooms/bathrooms, use value 5

---

## UI Recommendations

### 1. **Use Dropdowns**
- Bedrooms: 0-5 dropdown
- Bathrooms: 0-5 dropdown
- Square Feet: Preset ranges (optional)

### 2. **Input Masks**
- Phone: Auto-strip non-digits as user types
- Name: Block non-alphabetic characters

### 3. **Real-Time Validation**
- Show error messages as user types
- Green checkmark when valid
- Red border when invalid

### 4. **Clear Instructions**
- "Enter 10-digit phone number (digits only)"
- "Name must contain only letters"
- "Email must include @ and domain (.com, .org, etc.)"
