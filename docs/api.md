---
layout: default
title: API Reference
---

# API Reference

MailCleaner provides a REST API for managing accounts, rules, and previewing email matches.

## Base URL

```
http://localhost:8080/api
```

## Authentication

Currently, the API does not require authentication. Run behind a reverse proxy with authentication in production.

## Endpoints

### Accounts

#### List Accounts

```http
GET /api/accounts
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "Work Email",
    "server": "imap.gmail.com",
    "port": 993,
    "username": "user@gmail.com",
    "tls": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Create Account

```http
POST /api/accounts
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Work Email",
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@gmail.com",
  "password": "app-password",
  "tls": true
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Work Email",
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@gmail.com",
  "tls": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Get Account

```http
GET /api/accounts/:id
```

#### Update Account

```http
PUT /api/accounts/:id
Content-Type: application/json
```

**Request:**
```json
{
  "name": "Updated Name",
  "server": "imap.gmail.com",
  "port": 993,
  "username": "user@gmail.com",
  "password": "new-password",
  "tls": true
}
```

#### Delete Account

```http
DELETE /api/accounts/:id
```

**Response:** `204 No Content`

#### Test Connection

```http
POST /api/accounts/:id/test
```

**Response:**
```json
{
  "success": true,
  "message": "Connection successful",
  "folders": [
    { "name": "INBOX", "delimiter": "/", "attributes": [] },
    { "name": "Sent", "delimiter": "/", "attributes": [] }
  ],
  "total_emails": 1523
}
```

#### List Folders

```http
GET /api/accounts/:id/folders
```

**Response:**
```json
[
  { "name": "INBOX", "delimiter": "/", "attributes": [] },
  { "name": "Sent", "delimiter": "/", "attributes": ["\\Sent"] },
  { "name": "Drafts", "delimiter": "/", "attributes": ["\\Drafts"] }
]
```

### Rules

#### List Rules for Account

```http
GET /api/accounts/:id/rules
```

**Response:**
```json
[
  {
    "id": 1,
    "account_id": 1,
    "name": "GitHub Notifications",
    "pattern": "github.com",
    "pattern_type": "from_domain",
    "move_to_folder": "GitHub",
    "enabled": true,
    "priority": 10,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

#### Create Rule

```http
POST /api/accounts/:id/rules
Content-Type: application/json
```

**Request:**
```json
{
  "name": "GitHub Notifications",
  "pattern": "github.com",
  "pattern_type": "from_domain",
  "move_to_folder": "GitHub",
  "enabled": true,
  "priority": 10
}
```

#### Get Rule

```http
GET /api/rules/:id
```

#### Update Rule

```http
PUT /api/rules/:id
Content-Type: application/json
```

#### Delete Rule

```http
DELETE /api/rules/:id
```

### Preview

#### Preview Rule Matches

```http
GET /api/accounts/:id/preview?folder=INBOX&limit=100
```

**Query Parameters:**
- `folder` - IMAP folder to scan (default: INBOX)
- `limit` - Maximum messages to fetch (default: 100)

**Response:**
```json
{
  "total_messages": 1523,
  "matched_messages": 45,
  "messages": [
    {
      "uid": 12345,
      "seq_num": 100,
      "from": "notifications@github.com",
      "to": "user@gmail.com",
      "subject": "Your PR was merged",
      "date": "2024-01-15T10:30:00Z",
      "flags": ["\\Seen"],
      "matched_rule": {
        "id": 1,
        "name": "GitHub Notifications",
        "move_to_folder": "GitHub"
      }
    }
  ],
  "rule_matches": {
    "1": 45
  }
}
```

#### Apply Rules

```http
POST /api/accounts/:id/apply
Content-Type: application/json
```

**Request:**
```json
{
  "folder": "INBOX",
  "dry_run": false
}
```

**Response:**
```json
{
  "success": true,
  "moved": 45,
  "errors": []
}
```

## WebSocket API

### Live Preview

Connect to the WebSocket endpoint for real-time preview updates:

```
WS /ws/preview
```

#### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/preview');

ws.onopen = () => {
  // Start preview for an account
  ws.send(JSON.stringify({
    type: 'start_preview',
    account_id: 1,
    folder: 'INBOX'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Preview update:', data);
};
```

#### Message Types

**Client → Server:**

```json
{
  "type": "start_preview",
  "account_id": 1,
  "folder": "INBOX"
}
```

```json
{
  "type": "stop_preview"
}
```

**Server → Client:**

```json
{
  "type": "preview_started",
  "account_id": 1,
  "folder": "INBOX"
}
```

```json
{
  "type": "message",
  "message": {
    "uid": 12345,
    "from": "user@example.com",
    "subject": "Test email",
    "matched_rule": { ... }
  }
}
```

```json
{
  "type": "preview_complete",
  "total": 100,
  "matched": 25
}
```

```json
{
  "type": "error",
  "message": "Connection failed"
}
```

## Error Responses

All endpoints return errors in this format:

```json
{
  "error": "Error message here"
}
```

Common HTTP status codes:
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Next Steps

- [Usage Examples](usage) - See the API in action
- [Configuration Reference](configuration) - Full configuration options
- [Getting Started](getting-started) - Installation guide
