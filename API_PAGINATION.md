# API Pagination Documentation

> **For Frontend Team**  
> All paginated API endpoints with request/response details

---

## Pagination Standard

### Query Parameters

| Parameter | Type | Required | Default | Max | Description |
|-----------|------|----------|---------|-----|-------------|
| `page` | integer | No | `1` | - | Current page number (1-indexed) |
| `limit` | integer | No | `20` | `100` | Number of items per page |

### Response Format

All paginated endpoints return the following structure:

```json
{
  "data": [...],           // Array of items (field name varies)
  "total_count": 150,      // Total number of items
  "page": 1,               // Current page number
  "limit": 20,             // Items per page
  "total_pages": 8,        // Total number of pages
  "offset_start": 1,       // Starting item number (e.g., 1)
  "offset_end": 20         // Ending item number (e.g., 20)
}
```

---

## 1. Get Campaigns (Paginated)

### Endpoint
```
GET /api/campaigns
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Org Admin or Org User

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |
| `status` | string | No | `""` | Filter by status: `draft`, `scheduled`, `running`, `paused`, `completed` |
| `sort_by` | string | No | `""` | Field to sort by |
| `sort_order` | string | No | `DESC` | Sort order: `ASC` or `DESC` |

### Request Example

```bash
GET /api/campaigns?page=1&limit=20&status=scheduled&sort_by=created_at&sort_order=DESC
Authorization: Bearer <token>
```

### Response

```json
{
  "campaigns": [
    {
      "id": "uuid",
      "organization_id": "uuid",
      "name": "Summer Sale Campaign",
      "template_id": "uuid",
      "audience_ids": ["uuid1", "uuid2"],
      "contact_id": null,
      "schedule_type": "once",
      "scheduled_at": "2026-02-15T10:00:00Z",
      "status": "scheduled",
      "created_by": "uuid",
      "created_at": "2026-02-11T06:00:00Z",
      "updated_at": "2026-02-11T06:00:00Z"
    }
  ],
  "total_count": 45,
  "page": 1,
  "limit": 20,
  "total_pages": 3,
  "offset_start": 1,
  "offset_end": 20
}
```

---

## 2. Get Contacts (Paginated)

### Endpoint
```
GET /api/contacts
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Org Admin or Org User

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |
| `search` | string | No | `""` | Search by name, email, or phone |
| `sort_by` | string | No | `""` | Field to sort by |
| `sort_order` | string | No | `DESC` | Sort order: `ASC` or `DESC` |

### Request Example

```bash
GET /api/contacts?page=1&limit=20&search=john&sort_by=created_at&sort_order=DESC
Authorization: Bearer <token>
```

### Response

```json
{
  "contacts": [
    {
      "id": "uuid",
      "organization_id": "uuid",
      "created_by": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "phone": "+1234567890",
      "budget_min": 300000,
      "budget_max": 500000,
      "property_type": "Apartment",
      "bedrooms": 3,
      "bathrooms": 2,
      "square_feet": 1500,
      "preferred_location": "Downtown",
      "notes": "Looking for apartment near metro",
      "created_at": "2026-02-10T12:00:00Z",
      "updated_at": "2026-02-10T12:00:00Z"
    }
  ],
  "total_count": 150,
  "page": 1,
  "limit": 20,
  "total_pages": 8,
  "offset_start": 1,
  "offset_end": 20
}
```

---

## 3. Get Notifications (Paginated)

### Endpoint
```
GET /api/notifications
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Any authenticated user

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |

### Request Example

```bash
GET /api/notifications?page=1&limit=20
Authorization: Bearer <token>
```

### Response

```json
{
  "notifications": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "title": "CSV Import Complete",
      "message": "Your contact import has finished. 150 contacts added successfully.",
      "type": "info",
      "is_read": false,
      "created_at": "2026-02-11T10:00:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "limit": 20
}
```

**Note**: This endpoint uses `total` instead of `total_count` and doesn't include `total_pages`, `offset_start`, or `offset_end`.

---

## 4. Get Audience Contacts (Paginated)

### Endpoint
```
GET /api/audiences/:id/contacts
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Org Admin or Org User

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Audience ID |

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |

### Request Example

```bash
GET /api/audiences/abc123-def456/contacts?page=1&limit=20
Authorization: Bearer <token>
```

### Response

```json
{
  "contacts": [
    {
      "id": "uuid",
      "organization_id": "uuid",
      "first_name": "Jane",
      "last_name": "Smith",
      "email": "jane@example.com",
      "phone": "+1234567890",
      "budget_min": 200000,
      "budget_max": 350000,
      "property_type": "Condo",
      "bedrooms": 2,
      "bathrooms": 2,
      "preferred_location": "Suburbs",
      "created_at": "2026-02-05T08:00:00Z"
    }
  ],
  "total": 75,
  "page": 1,
  "limit": 20
}
```

**Note**: This endpoint uses `total` instead of `total_count` and doesn't include `total_pages`, `offset_start`, or `offset_end`.

---

## 5. Get Campaign Logs (Paginated)

### Endpoint
```
GET /api/campaigns/:id/logs
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Org Admin or Org User

### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string (UUID) | Campaign ID |

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `50` | Items per page (max: 100) |

### Request Example

```bash
GET /api/campaigns/abc123-def456/logs?page=1&limit=50
Authorization: Bearer <token>
```

### Response

```json
{
  "logs": [
    {
      "id": "uuid",
      "campaign_id": "uuid",
      "contact_id": "uuid",
      "contact_email": "user@example.com",
      "status": "sent",
      "error_message": null,
      "sent_at": "2026-02-11T09:00:00Z"
    }
  ],
  "total": 250,
  "page": 1,
  "limit": 50,
  "stats": {
    "total_sent": 250,
    "total_failed": 5,
    "total_pending": 0
  }
}
```

**Note**: This endpoint uses `total` instead of `total_count` and doesn't include `total_pages`, `offset_start`, or `offset_end`. It also includes a `stats` object with campaign statistics.

---

## 6. Get Organizations (Paginated)

### Endpoint
```
GET /superadmin/orgs
```

### Authentication
- **Required**: Yes (Bearer token)
- **Role**: Super Admin only

### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |

### Request Example

```bash
GET /superadmin/orgs?page=1&limit=20
Authorization: Bearer <super_admin_token>
```

### Response

```json
{
  "organizations": [
    {
      "id": "uuid",
      "name": "Amar Business Group",
      "is_active": true,
      "created_by": "super_admin_uuid",
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-02-10T15:30:00Z"
    },
    {
      "id": "uuid",
      "name": "Tech Solutions Inc",
      "is_active": true,
      "created_by": "super_admin_uuid",
      "created_at": "2026-02-01T08:00:00Z",
      "updated_at": "2026-02-01T08:00:00Z"
    }
  ],
  "total_count": 45,
  "page": 1,
  "limit": 20,
  "total_pages": 3,
  "offset_start": 1,
  "offset_end": 20
}
```

---

### 7. Get Agents (Paginated)

**Two endpoints available:**

#### A. Org Admin View

**Endpoint:**
```
GET /orgadmin/agents
```

**Authentication:** Org Admin only

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |

**Request Example:**
```bash
GET /orgadmin/agents?page=1&limit=20
Authorization: Bearer <org_admin_token>
```

**Response:**
```json
{
  "agents": [
    {
      "id": "uuid",
      "organization_id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "org_user",
      "is_active": true,
      "is_password_set": true,
      "invite_token": null,
      "token_expires_at": null,
      "created_at": "2026-02-05T10:00:00Z",
      "updated_at": "2026-02-06T12:00:00Z"
    },
    {
      "id": "uuid",
      "organization_id": "uuid",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "role": "org_admin",
      "is_active": true,
      "is_password_set": true,
      "created_at": "2026-02-01T09:00:00Z"
    }
  ],
  "total_count": 25,
  "page": 1,
  "limit": 20,
  "total_pages": 2,
  "offset_start": 1,
  "offset_end": 20
}
```

#### B. Super Admin View

**Endpoint:**
```
GET /superadmin/orgs/:org_id/agents
```

**Authentication:** Super Admin only

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `org_id` | string (UUID) | Organization ID |

**Query Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `page` | integer | No | `1` | Page number |
| `limit` | integer | No | `20` | Items per page (max: 100) |

**Request Example:**
```bash
GET /superadmin/orgs/abc123-def456/agents?page=1&limit=20
Authorization: Bearer <super_admin_token>
```

**Response:** Same format as Org Admin view

---

## Summary Table

| Endpoint | Default Limit | Max Limit | Has Search | Has Sorting | Has Status Filter |
|----------|---------------|-----------|------------|-------------|-------------------|
| GET /agent/campaigns | 20 | 100 | ❌ | ✅ | ✅ (status) |
| GET /agent/contacts | 20 | 100 | ✅ | ✅ | ❌ |
| GET /agent/notifications | 20 | 100 | ❌ | ❌ | ❌ |
| GET /agent/audiences/:id/contacts | 20 | 100 | ❌ | ❌ | ❌ |
| GET /agent/campaigns/:id/logs | 50 | 100 | ❌ | ❌ | ❌ |
| GET /superadmin/orgs | 20 | 100 | ❌ | ❌ | ❌ |
| GET /orgadmin/agents | 20 | 100 | ❌ | ❌ | ❌ |
| GET /superadmin/orgs/:org_id/agents | 20 | 100 | ❌ | ❌ | ❌ |

---

## Response Variations

### Full Pagination Metadata
**Endpoints**: `/agent/campaigns`, `/agent/contacts`, `/superadmin/orgs`, `/orgadmin/agents`, `/superadmin/orgs/:org_id/agents`

```json
{
  "data": [...],
  "total_count": 100,
  "page": 1,
  "limit": 20,
  "total_pages": 5,
  "offset_start": 1,
  "offset_end": 20
}
```

### Minimal Pagination Metadata
**Endpoints**: `/agent/notifications`, `/agent/audiences/:id/contacts`, `/agent/campaigns/:id/logs`

```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

---

## Important Notes

### Pagination Limits
- Minimum `page`: **1** (defaults to 1 if less)
- Minimum `limit`: **1** (defaults to endpoint default if less)
- Maximum `limit`: **100** (automatically capped)

### Empty Results
When there are no results:
```json
{
  "contacts": [],
  "total_count": 0,
  "page": 1,
  "limit": 20,
  "total_pages": 0,
  "offset_start": 0,
  "offset_end": 0
}
```

### Error Responses

**401 Unauthorized** - No/invalid token:
```json
{
  "error": "Unauthorized"
}
```

**403 Forbidden** - Wrong role/organization:
```json
{
  "error": "You don't have permission to access this resource"
}
```

**404 Not Found** - Resource doesn't exist:
```json
{
  "error": "Campaign not found"
}
```

**500 Server Error**:
```json
{
  "error": "Failed to fetch contacts"
}
```

---

## Frontend Implementation Example

### React/TypeScript Example

```typescript
interface PaginationParams {
  page: number;
  limit: number;
  search?: string;
  sortBy?: string;
  sortOrder?: 'ASC' | 'DESC';
  status?: string;
}

interface PaginatedResponse<T> {
  data: T[];
  total_count: number;
  page: number;
  limit: number;
  total_pages: number;
  offset_start: number;
  offset_end: number;
}

async function fetchContacts(params: PaginationParams) {
  const queryString = new URLSearchParams({
    page: params.page.toString(),
    limit: params.limit.toString(),
    ...(params.search && { search: params.search }),
    ...(params.sortBy && { sort_by: params.sortBy }),
    ...(params.sortOrder && { sort_order: params.sortOrder }),
  }).toString();

  const response = await fetch(`/api/contacts?${queryString}`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });

  return response.json();
}
```

### Usage
```typescript
const result = await fetchContacts({
  page: 1,
  limit: 20,
  search: 'john',
  sortBy: 'created_at',
  sortOrder: 'DESC'
});

console.log(`Showing ${result.offset_start}-${result.offset_end} of ${result.total_count}`);
// Output: "Showing 1-20 of 150"
```
