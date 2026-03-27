# API Reference

> REST API documentation for OctAi Web Backend

## Overview

The OctAi web backend provides a RESTful API for managing configuration, sessions, models, and gateway lifecycle.

**Base URL**: `http://localhost:18800`

---

## Authentication

### Access Control

The API supports IP-based access control. By default, only loopback addresses (127.0.0.0/8, ::1) are allowed.

**Configure allowed CIDRs:**

```json
{
  "launcher": {
    "public": true,
    "allowed_cidrs": ["192.168.1.0/24", "10.0.0.0/8"]
  }
}
```

### Token Authentication (Future)

> Planned feature: Bearer token authentication for API access

---

## Rate Limiting

Currently, no rate limiting is enforced. Consider implementing reverse proxy rate limiting for production:

```nginx
limit_req_zone $api_limit 10r/s;
```

---

## Endpoints

### Configuration

#### GET /api/config

Get the complete system configuration.

**Response:**

```json
{
  "model_list": [...],
  "agents": {...},
  "channels": {...},
  "tools": {...}
}
```

**Example:**

```bash
curl http://localhost:18800/api/config
```

---

#### PUT /api/config

Update the complete system configuration.

**Request Body:**

```json
{
  "model_list": [...],
  "agents": {...}
}
```

**Response:**

```json
{
  "status": "ok"
}
```

**Example:**

```bash
curl -X PUT http://localhost:18800/api/config \
  -H "Content-Type: application/json" \
  -d '{"model_list": [...]}'
```

---

#### PATCH /api/config

Partially update the configuration using JSON Merge Patch (RFC 7396).

**Request Body:**

```json
{
  "agents": {
    "defaults": {
      "model_name": "gpt-4o-mini"
    }
  }
}
```

**Response:**

```json
{
  "status": "ok"
}
```

**Example:**

```bash
curl -X PATCH http://localhost:18800/api/config \
  -H "Content-Type: application/json" \
  -d '{"agents": {"defaults": {"model_name": "gpt-4o-mini"}}}'
```

---

#### POST /api/config/test-command-patterns

Test a command against whitelist and blacklist patterns.

**Request Body:**

```json
{
  "allow_patterns": ["^git\\s+.*"],
  "deny_patterns": ["^rm\\s+-rf"],
  "command": "git status"
}
```

**Response:**

```json
{
  "allowed": true,
  "blocked": false,
  "matched_whitelist": "^git\\s+.*"
}
```

---

### Models

#### GET /api/models

List all configured models with masked API keys.

**Response:**

```json
{
  "models": [
    {
      "index": 0,
      "model_name": "gpt-4o-mini",
      "model": "openai/gpt-4o-mini",
      "api_key": "sk-****abcd",
      "configured": true,
      "is_default": true,
      "is_virtual": false
    }
  ],
  "total": 1,
  "default_model": "gpt-4o-mini"
}
```

---

#### POST /api/models

Add a new model configuration.

**Request Body:**

```json
{
  "model_name": "claude-sonnet-4",
  "model": "anthropic/claude-sonnet-4",
  "api_key": "sk-ant-your-key"
}
```

**Response:**

```json
{
  "status": "ok",
  "index": 1
}
```

---

#### POST /api/models/default

Set the default model for all agents.

**Request Body:**

```json
{
  "model_name": "claude-sonnet-4"
}
```

**Response:**

```json
{
  "status": "ok",
  "default_model": "claude-sonnet-4"
}
```

---

#### PUT /api/models/{index}

Update a model configuration by index.

**Path Parameters:**

| Name | Type | Description |
|------|------|-------------|
| index | integer | Model index in the list |

**Request Body:**

```json
{
  "model_name": "gpt-4o-mini",
  "model": "openai/gpt-4o-mini",
  "api_key": "sk-new-key"
}
```

---

#### DELETE /api/models/{index}

Delete a model configuration by index.

**Response:**

```json
{
  "status": "ok"
}
```

---

### Sessions

#### GET /api/sessions

List conversation sessions.

**Query Parameters:**

| Name | Type | Default | Description |
|------|------|---------|-------------|
| offset | integer | 0 | Pagination offset |
| limit | integer | 20 | Number of results |

**Response:**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Help with coding",
    "preview": "Can you help me debug this function?",
    "message_count": 4,
    "created": "2024-01-15T10:30:00Z",
    "updated": "2024-01-15T10:45:00Z"
  }
]
```

---

#### GET /api/sessions/{id}

Get a specific session with full message history.

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    },
    {
      "role": "assistant",
      "content": "Hi! How can I help you today?"
    }
  ],
  "summary": "Greeting conversation",
  "created": "2024-01-15T10:30:00Z",
  "updated": "2024-01-15T10:45:00Z"
}
```

---

#### DELETE /api/sessions/{id}

Delete a session.

**Response:** HTTP 204 No Content

---

### Gateway

#### GET /api/gateway/status

Get gateway process status and health.

**Response:**

```json
{
  "gateway_status": "running",
  "pid": 12345,
  "config_default_model": "gpt-4o-mini",
  "boot_default_model": "gpt-4o-mini",
  "gateway_restart_required": false,
  "gateway_start_allowed": true
}
```

**Status Values:**

| Value | Description |
|-------|-------------|
| `stopped` | Gateway is not running |
| `starting` | Gateway is starting up |
| `running` | Gateway is operational |
| `restarting` | Gateway is restarting |
| `error` | Gateway failed to start or crashed |

---

#### POST /api/gateway/start

Start the gateway process.

**Response:**

```json
{
  "status": "ok",
  "pid": 12345
}
```

**Error Response:**

```json
{
  "status": "precondition_failed",
  "message": "no default model configured"
}
```

---

#### POST /api/gateway/stop

Stop the gateway process gracefully.

**Response:**

```json
{
  "status": "ok",
  "pid": 12345
}
```

---

#### POST /api/gateway/restart

Restart the gateway process.

**Response:**

```json
{
  "status": "ok",
  "pid": 12346
}
```

---

#### GET /api/gateway/logs

Get buffered gateway logs.

**Query Parameters:**

| Name | Type | Description |
|------|------|-------------|
| log_offset | integer | Line offset for incremental updates |
| log_run_id | integer | Run ID to detect restarts |

**Response:**

```json
{
  "logs": [
    "[INFO] Gateway starting...",
    "[INFO] Connected to Telegram"
  ],
  "log_total": 50,
  "log_run_id": 1
}
```

---

#### POST /api/gateway/logs/clear

Clear the log buffer.

**Response:**

```json
{
  "status": "cleared",
  "log_total": 0,
  "log_run_id": 1
}
```

---

### Pico Channel (WebSocket)

#### GET /api/pico/status

Get Pico channel status and WebSocket URL.

**Response:**

```json
{
  "token": "abc123...",
  "ws_url": "ws://localhost:18790/ws",
  "enabled": true
}
```

---

#### POST /api/pico/token

Generate a new WebSocket token.

**Response:**

```json
{
  "token": "xyz789...",
  "ws_url": "ws://localhost:18790/ws"
}
```

---

#### POST /api/pico/setup

Auto-configure Pico channel for immediate use.

**Response:**

```json
{
  "token": "abc123...",
  "ws_url": "ws://localhost:18790/ws",
  "enabled": true,
  "changed": true
}
```

---

### Channels Catalog

#### GET /api/channels

List available communication channels.

**Response:**

```json
[
  {"name": "telegram", "config_key": "telegram"},
  {"name": "discord", "config_key": "discord"},
  {"name": "slack", "config_key": "slack"},
  {"name": "whatsapp", "config_key": "whatsapp", "variant": "bridge"},
  {"name": "whatsapp_native", "config_key": "whatsapp", "variant": "native"},
  {"name": "matrix", "config_key": "matrix"},
  {"name": "feishu", "config_key": "feishu"},
  {"name": "dingtalk", "config_key": "dingtalk"},
  {"name": "wecom", "config_key": "wecom"},
  {"name": "line", "config_key": "line"},
  {"name": "qq", "config_key": "qq"},
  {"name": "onebot", "config_key": "onebot"},
  {"name": "irc", "config_key": "irc"},
  {"name": "maixcam", "config_key": "maixcam"},
  {"name": "pico", "config_key": "pico"}
]
```

---

### Skills

#### GET /api/skills

List installed skills.

**Response:**

```json
{
  "skills": [
    {
      "name": "git-expert",
      "description": "Git and GitHub workflow assistance",
      "enabled": true
    }
  ]
}
```

---

### OAuth

#### GET /api/oauth/providers

List available OAuth providers.

**Response:**

```json
{
  "providers": [
    {
      "name": "anthropic",
      "display_name": "Anthropic",
      "auth_url": "/api/oauth/anthropic"
    }
  ]
}
```

---

#### GET /api/oauth/{provider}/login

Initiate OAuth login flow.

**Response:** Redirect to provider's OAuth page

---

#### GET /api/oauth/{provider}/callback

Handle OAuth callback.

**Response:**

```json
{
  "status": "success",
  "provider": "anthropic"
}
```

---

## Error Responses

### Validation Error

```json
{
  "status": "validation_error",
  "errors": [
    "channels.telegram.token is required when telegram channel is enabled"
  ]
}
```

### Not Found

```json
{
  "error": "session not found"
}
```

### Access Denied

```json
{
  "error": "access denied by network policy"
}
```

---

## WebSocket Protocol
The Pico channel uses WebSocket for real-time chat communication.

### Connection

```
ws://localhost:18790/ws?token=YOUR_TOKEN
```

### Message Format
Messages are JSON objects:
```json
{
  "type": "message",
  "content": "Hello!"
}
```

### Response Format
```json
{
  "type": "response",
  "content": "Hi! How can I help?"
}
```

---

## OpenAPI Specification
> Coming soon: Full OpenAPI 3.0 specification
