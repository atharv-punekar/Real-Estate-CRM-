# Cascading Contact Deletion

> **For Frontend Team** - Contact deletion behavior

---

## ⚠️ IMPORTANT: Automatic Cleanup on Contact Deletion

When a contact is deleted via `DELETE /agent/contacts/:id`, the system now **automatically removes** the contact from:

1. **All audiences** - Removed from `audience_contact` associations
2. **All campaigns** - Contact reference removed (contact_id set to NULL)

---

## Deletion Behavior

### What Happens

```
DELETE /agent/contacts/{contact_id}
↓
1. Remove from all audiences (audience_contact deleted)
2. Remove from all campaigns (contact_id → NULL)  
3. Soft delete contact (is_active = false)
```

### Transactional Guarantee

All three operations happen in a **single database transaction**:
- ✅ **All succeed** → Contact fully removed
- ❌ **Any fails** → Entire operation rolls back (nothing changes)

---

## Campaign Behavior

**When a contact is deleted:**
- ✅ Campaign **remains active**
- ✅ `contact_id` field set to `NULL`
- ✅ Campaign can still send to audiences

**Example:**
```json
// Before deletion
{
  "id": "campaign-123",
  "contact_id": "contact-456",
  "audience_ids": ["aud-1", "aud-2"]
}

// After contact deleted
{
  "id": "campaign-123",
  "contact_id": null,  // ← Contact reference removed
  "audience_ids": ["aud-1", "aud-2"]  // ← Audiences unchanged
}
```

---

## Audience Behavior

**When a contact is deleted:**
- ✅ Contact removed from **all audiences** automatically
- ✅ Audience count decreases
- ✅ No orphaned references

---

## Frontend Implications

### 1. **No Manual Cleanup Required**
You don't need to:
- Remove contact from audiences before deletion
- Update campaigns before deletion
- Handle orphaned references

### 2. **Single Delete Call**
```javascript
// Just delete the contact - cleanup is automatic
await deleteContact(contactId);
// ✅ Contact removed from audiences
// ✅ Contact removed from campaigns  
// ✅ Contact soft-deleted
```

### 3. **Error Handling**
```javascript
try {
  await deleteContact(contactId);
  // Success - all cleanup done
} catch (error) {
  // Failure - nothing changed (transaction rolled back)
  console.error('Delete failed:', error);
}
```

### 4. **UI Updates**
After successful deletion:
- ✅ Refresh audience contact counts
- ✅ Refresh campaign lists (if showing contact_id)
- ✅ Remove from contact list

---

## API Response

**Success:**
```json
{
  "message": "Contact deleted successfully"
}
```

**Error:**
```json
{
  "error": "Failed to delete contact"
}
```

---

## Summary

| Action | Old Behavior | New Behavior |
|--------|-------------|--------------|
| Delete contact | Contact soft-deleted only | **Cascading delete**: removed from audiences, campaigns, then soft-deleted |
| Audience associations | Remained in database | **Automatically removed** |
| Campaign references | `contact_id` remained | **Set to NULL** |
| Transaction safety | Not guaranteed | **Atomic operation** (all or nothing) |

---

## Breaking Changes

❌ **None** - This is purely backend improvement. The API endpoint remains the same.

---

**Implementation Date:** 2026-02-11
