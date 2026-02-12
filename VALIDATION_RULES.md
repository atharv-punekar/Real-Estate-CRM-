# Validation Rules Documentation

> **For Frontend Team**  
> All validation rules and error messages for form fields

---

## Table of Contents
- [Agent Name Validation](#agent-name-validation)
- [Agent Email Validation](#agent-email-validation)
- [Password Validation](#password-validation)
- [Quick Reference](#quick-reference)

---

## Agent Name Validation

**Applied to:** Agent creation and updates (both org admin and super admin)

### Rules

| Rule | Description |
|------|-------------|
| **Characters** | Only alphabetic characters (a-z, A-Z) allowed |
| **Spaces** | Exactly ONE space allowed (between first and last name) |
| **First Name** | Minimum 2 characters |
| **Last Name** | Minimum 2 characters |
| **Format** | `FirstName LastName` (e.g., "John Doe") |

### Valid Examples
```
✅ John Doe
✅ Jane Smith
✅ Ab Cd
✅ Muhammad Ali
```

### Invalid Examples
```
❌ John           (no last name)
❌ JohnDoe        (no space)
❌ John  Doe      (multiple spaces)
❌ John-Doe       (special character)
❌ John123        (contains numbers)
❌ J Doe          (first name too short)
❌ John D         (last name too short)
```

### Error Messages

| Validation | Error Message |
|------------|---------------|
| Empty name | "name cannot be empty" |
| Not exactly 2 parts | "name must contain exactly one space (first name and last name)" |
| Invalid characters | "name can only contain alphabetic characters and one space" |
| First name too short | "first name must be at least 2 characters long" |
| Last name too short | "last name must be at least 2 characters long" |

### Frontend Implementation

```javascript
function validateAgentName(name) {
  if (!name || name.trim() === '') {
    return 'Name cannot be empty';
  }
  
  const parts = name.trim().split(' ');
  if (parts.length !== 2) {
    return 'Name must contain exactly one space (first name and last name)';
  }
  
  const [firstName, lastName] = parts;
  
  // Check alphabetic only
  const alphaOnly = /^[a-zA-Z]+$/;
  if (!alphaOnly.test(firstName) || !alphaOnly.test(lastName)) {
    return 'Name can only contain alphabetic characters and one space';
  }
  
  // Check minimum lengths
  if (firstName.length < 2) {
    return 'First name must be at least 2 characters long';
  }
  if (lastName.length < 2) {
    return 'Last name must be at least 2 characters long';
  }
  
  return null; // Valid
}
```

---

## Agent Email Validation

**Applied to:** Agent creation and updates (both org admin and super admin)

### Rules

| Rule | Description |
|------|-------------|
| **Format** | Standard email format (e.g., user@example.com) |
| **Local Part** | Letters, numbers, dots, underscores, plus/minus signs allowed |
| **@ Symbol** | Exactly ONE @ symbol required |
| **Domain** | Valid domain format (e.g., example.com) |
| **Domain Dots** | Must contain at least one dot |

### Valid Examples
```
✅ john@example.com
✅ john.doe@company.org
✅ user+tag@mail.example.com
✅ agent_name@real-estate.com
✅ 123user@domain.net
```

### Invalid Examples
```
❌ john@@example.com      (multiple @ symbols)
❌ @example.com           (missing local part)
❌ john@                  (missing domain)
❌ john@example           (no domain extension)
```

### Error Messages

| Validation | Error Message |
|------------|---------------|
| Empty email | "email cannot be empty" |
| Invalid format | "email must be a valid format (e.g., user@example.com)" |
| Invalid @ count | "email must contain exactly one '@' symbol" |

### Frontend Implementation

```javascript
function validateAgentEmail(email) {
  if (!email || email.trim() === '') {
    return 'Email cannot be empty';
  }
  
  const trimmed = email.trim();
  
  // Standard email regex
  const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
  if (!emailRegex.test(trimmed)) {
    return 'Email must be a valid format (e.g., user@example.com)';
  }
  
  // Check for exactly one @
  const atCount = (trimmed.match(/@/g) || []).length;
  if (atCount !== 1) {
    return 'Email must contain exactly one @ symbol';
  }
  
  return null; // Valid
}
```

---

## Password Validation

**Applied to:** Password activation/set password endpoint

### Rules

| Rule | Description |
|------|-------------|
| **Minimum Length** | 6 characters |
| **Maximum Length** | 20 characters |
| **Lowercase** | At least 1 lowercase letter (a-z) |
| **Digit** | At least 1 numeric digit (0-9) |
| **Special Character** | At least 1 special character |

### Allowed Special Characters
```
! @ # $ % ^ & * ( ) _ + - = [ ] { } | ; : , . < > ? / ~ `
```

### Valid Examples
```
✅ Pass123!
✅ myP@ss1
✅ secureP@ssw0rd
✅ a1!bcd
```

### Invalid Examples
```
❌ abc12!        (too short - only 6 chars minimum)
❌ Pass123       (no special character)
❌ Password!     (no digit)
❌ PASS123!      (no lowercase letter)
❌ abcdefghijklmnopqrst1! (too long - exceeds 20 chars)
```

### Error Messages

| Validation | Error Message |
|------------|---------------|
| Empty password | "password cannot be empty" |
| Too short | "password must be at least 6 characters long" |
| Too long | "password must not exceed 20 characters" |
| Missing requirements | "password must contain at least: [missing items]" |

**Combined Error Example:**
```
"password must contain at least: one lowercase letter, one numeric digit, one special character"
```

### Frontend Implementation

```javascript
function validatePassword(password) {
  if (!password) {
    return 'Password cannot be empty';
  }
  
  if (password.length < 6) {
    return 'Password must be at least 6 characters long';
  }
  
  if (password.length > 20) {
    return 'Password must not exceed 20 characters';
  }
  
  const hasLower = /[a-z]/.test(password);
  const hasDigit = /[0-9]/.test(password);
  const hasSpecial = /[!@#$%^&*()_+\-=\[\]{}|;:,.<>?/~`]/.test(password);
  
  const missing = [];
  if (!hasLower) missing.push('one lowercase letter');
  if (!hasDigit) missing.push('one numeric digit');
  if (!hasSpecial) missing.push('one special character');
  
  if (missing.length > 0) {
    return `Password must contain at least: ${missing.join(', ')}`;
  }
  
  return null; // Valid
}
```

### Password Strength Indicator

**Recommended UI Component:**

```javascript
function getPasswordStrength(password) {
  if (!password) return { strength: 0, label: 'Empty', color: 'gray' };
  
  let strength = 0;
  
  // Length
  if (password.length >= 6) strength++;
  if (password.length >= 10) strength++;
  if (password.length >= 15) strength++;
  
  // Character types
  if (/[a-z]/.test(password)) strength++;
  if (/[A-Z]/.test(password)) strength++;
  if (/[0-9]/.test(password)) strength++;
  if (/[!@#$%^&*()_+\-=\[\]{}|;:,.<>?/~`]/.test(password)) strength++;
  
  if (strength <= 2) return { strength, label: 'Weak', color: 'red' };
  if (strength <= 4) return { strength, label: 'Fair', color: 'orange' };
  if (strength <= 6) return { strength, label: 'Good', color: 'yellow' };
  return { strength, label: 'Strong', color: 'green' };
}
```

---

## Quick Reference

### API Endpoints Using These Validations

| Endpoint | Name | Email | Password |
|----------|------|-------|----------|
| `POST /auth/activate` | ❌ | ❌ | ✅ |
| `POST /orgadmin/agents` | ✅ | ✅ | ❌ |
| `PUT /orgadmin/agents/:id` | ✅ | ✅ | ❌ |
| `POST /superadmin/orgs/:org_id/agents` | ✅ | ✅ | ❌ |
| `PUT /superadmin/orgs/:org_id/agents/:id` | ✅ | ✅ | ❌ |

### Validation Summary Table

| Field | Min Length | Max Length | Allowed Characters | Special Rules |
|-------|------------|------------|-------------------|---------------|
| **Name** | 2 (each part) | - | a-z, A-Z, 1 space | Exactly 2 parts (first + last) |
| **Email (local)** | 1 | - | a-z, 0-9 | Before @ symbol |
| **Email (domain)** | 3+ | - | a-z, 0-9, ., - | Must have valid domain format |
| **Password** | 6 | 20 | Any | Must have: lowercase, digit, special char |

### Expected HTTP Error Codes

| Status Code | Scenario |
|-------------|----------|
| **400 Bad Request** | Validation error (invalid format) |
| **409 Conflict** | Email already exists |
| **409 Conflict** | Name already exists in organization |

### Sample Error Response

```json
{
  "error": "password must contain at least: one lowercase letter, one special character"
}
```

---

## Testing Checklist

### For Name Field
- [ ] Test with valid two-part name
- [ ] Test with single word (should fail)
- [ ] Test with three+ words (should fail)
- [ ] Test with numbers (should fail)
- [ ] Test with special characters (should fail)
- [ ] Test with too-short first name (should fail)
- [ ] Test with too-short last name (should fail)
- [ ] Test with multiple spaces (should fail)

### For Email Field
- [ ] Test with valid lowercase email
- [ ] Test with uppercase letters (should fail)
- [ ] Test with special chars in local part (should fail)
- [ ] Test with missing @ (should fail)
- [ ] Test with multiple @ (should fail)
- [ ] Test without domain extension (should fail)
- [ ] Test with numbers (should succeed)
- [ ] Test with hyphens in domain (should succeed)

### For Password Field
- [ ] Test with valid password (6-20 chars, has lower, digit, special)
- [ ] Test with <6 characters (should fail)
- [ ] Test with >20 characters (should fail)
- [ ] Test without lowercase (should fail)
- [ ] Test without digit (should fail)
- [ ] Test without special char (should fail)
- [ ] Test with all requirements met (should succeed)
- [ ] Test combined missing requirements error message

---

## Notes for Frontend Team

1. **Real-time Validation**: Implement client-side validation for better UX before API calls
2. **Error Display**: Show validation errors inline below each field
3. **Password Visibility Toggle**: Add show/hide password button
4. **Password Strength Meter**: Display visual strength indicator
5. **Character Counter**: Show remaining characters for password (x/20)
6. **Autocomplete**: Set `autocomplete="new-password"` for password fields
7. **Email Lowercase**: Auto-convert email input to lowercase on blur
8. **Name Formatting**: Consider auto-capitalizing first letter of each name part
9. **Unique Email**: Handle 409 Conflict for duplicate emails gracefully
10. **Unique Name**: Handle 409 Conflict for duplicate names within org

---

## Change Log

- **2026-02-11**: Initial validation rules documentation
  - Agent name validation (alphabetic only, 2 parts, min 2 chars each)
  - Agent email validation (lowercase + numbers, valid domain)
  - Password validation (6-20 chars, lower + digit + special)
  - Improved error messages for combined password requirements
