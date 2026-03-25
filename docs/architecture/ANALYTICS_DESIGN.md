# Analytics & Insights System Design

> Design document for AI Business HQ analytics and business intelligence features

## Overview

This document outlines the analytics and insights system for AI Business HQ, designed to provide business users with actionable intelligence about AI usage, costs, and performance.

---

## 1. Metrics to Track

### 1.1 Usage Metrics

| Metric | Description | Granularity |
|--------|-------------|-------------|
| `messages_total` | Total messages processed (user + assistant) | Per session, per channel, per agent |
| `messages_user` | User-originated messages | Per session, per channel |
| `messages_assistant` | AI-generated responses | Per session, per channel |
| `tokens_input` | Input tokens consumed | Per model, per agent |
| `tokens_output` | Output tokens generated | Per model, per agent |
| `tokens_total` | Sum of input + output tokens | Per model, per agent |
| `api_calls` | LLM API requests made | Per provider, per model |
| `api_calls_streaming` | Streaming API calls | Per provider |
| `api_calls_batch` | Non-streaming API calls | Per provider |
| `tool_invocations` | Tool executions | Per tool, per agent |
| `tool_errors` | Failed tool executions | Per tool |
| `sessions_created` | New conversation sessions | Per channel |
| `sessions_active` | Sessions with activity in period | Per channel |
| `media_processed` | Images/audio files processed | Per type |

### 1.2 Performance Metrics

| Metric | Description | Type |
|--------|-------------|------|
| `response_time_p50` | Median response latency | Histogram |
| `response_time_p95` | 95th percentile latency | Histogram |
| `response_time_p99` | 99th percentile latency | Histogram |
| `time_to_first_token` | Streaming TTFT | Histogram |
| `tokens_per_second` | Streaming generation speed | Gauge |
| `tool_execution_time` | Tool call duration | Histogram |
| `context_compression_time` | History compression duration | Histogram |
| `uptime_seconds` | Gateway uptime | Counter |
| `error_rate` | Errors per 1000 requests | Gauge |
| `retry_rate` | Retry attempts per request | Gauge |
| `cooldown_activations` | Provider cooldown events | Counter |

### 1.3 Business Metrics

| Metric | Description | Calculation |
|--------|-------------|-------------|
| `cost_total` | Total estimated API costs | Sum of (tokens × price) |
| `cost_per_model` | Costs by model | Per-model aggregation |
| `cost_per_channel` | Costs by channel | Per-channel aggregation |
| `cost_per_user` | Costs per user (if tracked) | User aggregation |
| `savings_fallback` | Savings from fallback to cheaper models | Price differential |
| `savings_caching` | Savings from response caching | Cached requests × avg cost |
| `roi_indicator` | Basic ROI proxy | Tasks completed / cost |
| `automation_rate` | Automated vs manual interactions | Tool calls / total interactions |

### 1.4 User Engagement

| Metric | Description | Calculation |
|--------|-------------|-------------|
| `dau` | Daily Active Users | Unique users with activity |
| `mau` | Monthly Active Users | Unique users over 30 days |
| `session_length_avg` | Average conversation length | Messages / session |
| `retention_1d` | 1-day retention | Returned users / total |
| `retention_7d` | 7-day retention | Returned users / total |
| `messages_per_user` | Messages per active user | Total / DAU or MAU |
| `peak_concurrent` | Peak concurrent sessions | Max sessions in time window |
| `channel_distribution` | Usage by channel | Per-channel breakdown |

---

## 2. Data Collection

### 2.1 Event Logging Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Event Flow                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │  Agent   │───▶│   Event      │───▶│   Analytics          │  │
│  │  Loop    │    │   Bus        │    │   Collector          │  │
│  └──────────┘    └──────────────┘    └──────────┬───────────┘  │
│                                                │                │
│                                                ▼                │
│                                    ┌──────────────────────┐    │
│                                    │   Event Processor    │    │
│                                    │   (async)            │    │
│                                    └──────────┬───────────┘    │
│                                               │                 │
│                          ┌────────────────────┼────────────┐   │
│                          ▼                    ▼            ▼   │
│                   ┌─────────────┐    ┌─────────────┐  ┌──────┐ │
│                   │  Real-time  │    │   Batch     │  │ Log  │ │
│                   │  Aggregator │    │   Queue     │  │ File │ │
│                   └──────┬──────┘    └──────┬──────┘  └──────┘ │
│                          │                  │                   │
│                          ▼                  ▼                   │
│                   ┌─────────────┐    ┌─────────────┐           │
│                   │   Redis     │    │   SQLite    │           │
│                   │   (hot)     │    │   (cold)    │           │
│                   └─────────────┘    └─────────────┘           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Event Types

Leverage existing `pkg/agent/events.go` event system:

```go
type AnalyticsEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"`
    AgentID     string                 `json:"agent_id"`
    SessionID   string                 `json:"session_id"`
    Channel     string                 `json:"channel,omitempty"`
    ChatID      string                 `json:"chat_id,omitempty"`
    Model       string                 `json:"model,omitempty"`
    Provider    string                 `json:"provider,omitempty"`
    Metadata    map[string]any         `json:"metadata,omitempty"`
    Metrics     EventMetrics           `json:"metrics"`
}

type EventMetrics struct {
    TokensIn        int     `json:"tokens_in,omitempty"`
    TokensOut       int     `json:"tokens_out,omitempty"`
    DurationMs      int     `json:"duration_ms,omitempty"`
    CostEstimate    float64 `json:"cost_estimate,omitempty"`
    Success         bool    `json:"success"`
    ErrorCode       string  `json:"error_code,omitempty"`
}
```

### 2.3 Real-time vs Batch Processing

| Data Type | Processing | Retention | Storage |
|-----------|------------|-----------|---------|
| Live metrics | Real-time (stream) | 24 hours | Redis |
| Hourly rollups | Batch (every hour) | 30 days | SQLite |
| Daily aggregates | Batch (daily) | 1 year | SQLite |
| Raw events | Batch (async write) | 7 days | SQLite + log files |

#### Real-time Processing (MVP)

```go
type RealtimeAggregator struct {
    redis       *redis.Client
    tickWindow  time.Duration  // 1 minute
    metrics     *sync.Map
}

func (a *RealtimeAggregator) Record(event AnalyticsEvent) {
    key := fmt.Sprintf("metrics:%s:%s", event.EventType, time.Now().Format("2006-01-02:15:04"))
    a.redis.HIncrBy(ctx, key, "count", 1)
    a.redis.Expire(ctx, key, 24*time.Hour)
}
```

#### Batch Processing

```go
type BatchProcessor struct {
    db          *sql.DB
    interval    time.Duration  // 1 hour
    eventBuffer []AnalyticsEvent
}

func (p *BatchProcessor) ProcessHourlyAggregates() error {
    query := `
        INSERT INTO hourly_metrics (
            timestamp, agent_id, channel, model,
            messages, tokens_in, tokens_out, api_calls,
            avg_latency_ms, error_count, cost_estimate
        )
        SELECT 
            date_trunc('hour', timestamp),
            agent_id, channel, model,
            COUNT(*) FILTER (WHERE event_type = 'llm_response'),
            SUM(tokens_in), SUM(tokens_out),
            COUNT(*) FILTER (WHERE event_type = 'llm_request'),
            AVG(duration_ms),
            COUNT(*) FILTER (WHERE NOT success),
            SUM(cost_estimate)
        FROM analytics_events
        WHERE timestamp >= NOW() - INTERVAL '1 hour'
        GROUP BY 1, 2, 3, 4
        ON CONFLICT (timestamp, agent_id, channel, model) 
        DO UPDATE SET 
            messages = EXCLUDED.messages + hourly_metrics.messages,
            tokens_in = EXCLUDED.tokens_in + hourly_metrics.tokens_in,
            ...
    `
    return p.db.Exec(query)
}
```

### 2.4 Storage Strategy

#### SQLite Schema (Primary Storage)

```sql
-- Raw events (hot storage, 7-day retention)
CREATE TABLE analytics_events (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp    DATETIME NOT NULL,
    event_type   TEXT NOT NULL,
    agent_id     TEXT NOT NULL,
    session_id   TEXT,
    channel      TEXT,
    chat_id      TEXT,
    model        TEXT,
    provider     TEXT,
    tokens_in    INTEGER DEFAULT 0,
    tokens_out   INTEGER DEFAULT 0,
    duration_ms  INTEGER DEFAULT 0,
    cost_estimate REAL DEFAULT 0,
    success      BOOLEAN DEFAULT true,
    error_code   TEXT,
    metadata     TEXT,  -- JSON
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_timestamp ON analytics_events(timestamp);
CREATE INDEX idx_events_agent ON analytics_events(agent_id);
CREATE INDEX idx_events_type ON analytics_events(event_type);

-- Hourly aggregates
CREATE TABLE hourly_metrics (
    timestamp      DATETIME NOT NULL,
    agent_id       TEXT NOT NULL,
    channel        TEXT,
    model          TEXT,
    messages       INTEGER DEFAULT 0,
    tokens_in      INTEGER DEFAULT 0,
    tokens_out     INTEGER DEFAULT 0,
    api_calls      INTEGER DEFAULT 0,
    avg_latency_ms REAL DEFAULT 0,
    p95_latency_ms REAL DEFAULT 0,
    error_count    INTEGER DEFAULT 0,
    cost_estimate  REAL DEFAULT 0,
    PRIMARY KEY (timestamp, agent_id, channel, model)
);

-- Daily aggregates
CREATE TABLE daily_metrics (
    date           DATE NOT NULL,
    agent_id       TEXT NOT NULL,
    channel        TEXT,
    model          TEXT,
    messages       INTEGER DEFAULT 0,
    tokens_in      INTEGER DEFAULT 0,
    tokens_out     INTEGER DEFAULT 0,
    api_calls      INTEGER DEFAULT 0,
    avg_latency_ms REAL DEFAULT 0,
    error_count    INTEGER DEFAULT 0,
    cost_estimate  REAL DEFAULT 0,
    dau            INTEGER DEFAULT 0,
    sessions       INTEGER DEFAULT 0,
    PRIMARY KEY (date, agent_id, channel, model)
);

-- Model pricing (for cost calculations)
CREATE TABLE model_pricing (
    model          TEXT PRIMARY KEY,
    provider       TEXT NOT NULL,
    input_price    REAL NOT NULL,   -- per 1M tokens
    output_price   REAL NOT NULL,   -- per 1M tokens
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Retention policy tracking
CREATE TABLE retention_runs (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    run_at         DATETIME NOT NULL,
    events_deleted INTEGER,
    oldest_kept    DATETIME
);
```

#### Redis Keys (Real-time Cache)

```
metrics:minute:{timestamp}          -> Hash: event_type -> count
metrics:hour:{timestamp}            -> Hash: aggregated metrics
counter:messages:total              -> Integer
counter:tokens:total                -> Integer
gauge:latency:p95:{timestamp}       -> Float
set:active_users:{date}             -> Set of user IDs
```

---

## 3. Dashboard Design

### 3.1 Key Visualizations

#### Overview Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  AI Business HQ - Analytics Dashboard                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │  Messages   │ │   Tokens    │ │ API Calls   │ │   Cost      │   │
│  │   12,453    │ │  2.4M       │ │   4,521     │ │   $47.82    │   │
│  │  ↑ 12%      │ │  ↑ 8%       │ │  ↑ 15%      │ │  ↓ 3%       │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │              Message Volume (24h)                               │ │
│  │  ▁▂▃▅▆▇█▇▆▅▄▃▂▁▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▁▂▃▄▅▆▇█        │ │
│  │  00:00                    12:00                    23:00       │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────┐  ┌─────────────────────────┐          │
│  │  Usage by Channel       │  │  Response Latency       │          │
│  │  ┌──────────────────┐   │  │  ┌──────────────────┐   │          │
│  │  │ Telegram  45%    │   │  │  │ p50: 1.2s        │   │          │
│  │  │ Slack     30%    │   │  │  │ p95: 3.4s        │   │          │
│  │  │ Discord   15%    │   │  │  │ p99: 5.8s        │   │          │
│  │  │ Web       10%    │   │  │  │                  │   │          │
│  │  └──────────────────┘   │  │  │ ▁▂▃▄▅▆▇█▇▆▅▄     │   │          │
│  └─────────────────────────┘  └─────────────────────────┘          │
│                                                                      │
│  ┌─────────────────────────┐  ┌─────────────────────────┐          │
│  │  Top Models by Usage    │  │  Cost Trend (7d)        │          │
│  │  gpt-4o-mini    52%     │  │  $12 ────────────────── │          │
│  │  claude-sonnet  28%     │  │     ╱╲    ╱╲            │          │
│  │  gpt-4o         15%     │  │    ╱  ╲  ╱  ╲           │          │
│  │  Other           5%     │  │   ╱    ╲╱    ╲          │          │
│  └─────────────────────────┘  └─────────────────────────┘          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### Performance Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  Performance Metrics                                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  Response Time Distribution (ms)                              │ │
│  │  0-500ms    ████████████████████  45%                        │ │
│  │  500-1s     ██████████████        28%                        │ │
│  │  1-2s       ████████              15%                        │ │
│  │  2-5s       ████                   8%                        │ │
│  │  >5s        ██                     4%                        │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  Error Rate Over Time                                         │ │
│  │  5% ┤                                                        │ │
│  │  4% ┤    *                                                   │ │
│  │  3% ┤  *   *     *                                           │ │
│  │  2% ┤*       *  *  *   *                                     │ │
│  │  1% ┤          *     *   * *                                 │ │
│  │  0% └────────────────────────────────────────────────────────│ │
│  │     Mon  Tue  Wed  Thu  Fri  Sat  Sun                        │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌──────────────────────────┐  ┌──────────────────────────┐       │
│  │  Tool Performance        │  │  Provider Health         │       │
│  │  shell     avg: 2.3s     │  │  OpenAI     ● Healthy    │       │
│  │  web       avg: 1.8s     │  │  Anthropic  ● Healthy    │       │
│  │  file      avg: 0.4s     │  │  Azure      ● Degraded   │       │
│  │  spawn     avg: 15.2s    │  │  Ollama     ● Offline    │       │
│  └──────────────────────────┘  └──────────────────────────┘       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

#### Cost Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  Cost Analysis                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │  Today      │ │  This Week  │ │ This Month  │ │  Projected  │   │
│  │   $12.45    │ │   $78.32    │ │   $245.67   │ │   $312.00   │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  Cost by Model (30d)                                          │ │
│  │  gpt-4o        ████████████████████  $124.50                 │ │
│  │  claude-sonnet ████████████         $78.20                    │ │
│  │  gpt-4o-mini   ████                 $28.90                    │ │
│  │  ollama        ░░░░                 $0.00 (local)             │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │  Cost Savings                                                 │ │
│  │  ┌─────────────────────────────────────────────────────────┐ │ │
│  │  │ Fallback to cheaper models:    $45.20 saved             │ │ │
│  │  │ Context caching:               $12.80 saved             │ │ │
│  │  │ Local model usage:             $89.00 saved             │ │ │
│  │  │ ───────────────────────────────────────────────────     │ │ │
│  │  │ Total estimated savings:       $147.00                  │ │ │
│  │  └─────────────────────────────────────────────────────────┘ │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Real-time vs Historical Data

| Component | Data Source | Refresh Rate |
|-----------|-------------|--------------|
| Live counters | Redis | 5 seconds |
| Minute charts | Redis + SQLite | 1 minute |
| Hourly trends | SQLite | 5 minutes |
| Daily summaries | SQLite | 1 hour |
| Historical comparisons | SQLite | On-demand |

### 3.3 Export Capabilities

#### Supported Formats

```go
type ExportFormat string

const (
    ExportFormatCSV    ExportFormat = "csv"
    ExportFormatJSON   ExportFormat = "json"
    ExportFormatExcel  ExportFormat = "xlsx"
)

type ExportRequest struct {
    Format      ExportFormat
    StartDate   time.Time
    EndDate     time.Time
    Metrics     []string
    GroupBy     []string  // e.g., ["channel", "model"]
    Resolution  string    // "hour", "day", "week"
}
```

#### Export API

```
POST /api/v1/analytics/export
Content-Type: application/json

{
  "format": "csv",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "metrics": ["messages", "tokens_in", "tokens_out", "cost"],
  "group_by": ["channel", "model"],
  "resolution": "day"
}

Response: 200 OK
Content-Type: text/csv
Content-Disposition: attachment; filename="analytics_2024-01-01_2024-01-31.csv"
```

---

## 4. API Endpoints

### 4.1 GET /api/v1/analytics/usage

Retrieve usage metrics with optional filtering.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start` | string | -24h | Start timestamp (ISO 8601 or relative) |
| `end` | string | now | End timestamp |
| `resolution` | string | hour | Aggregation: minute, hour, day |
| `channel` | string | all | Filter by channel |
| `agent_id` | string | all | Filter by agent |
| `model` | string | all | Filter by model |

**Response:**

```json
{
  "period": {
    "start": "2024-01-15T00:00:00Z",
    "end": "2024-01-15T23:59:59Z",
    "resolution": "hour"
  },
  "summary": {
    "messages_total": 12453,
    "messages_user": 6234,
    "messages_assistant": 6219,
    "tokens_input": 1245678,
    "tokens_output": 876543,
    "tokens_total": 2122221,
    "api_calls": 4521,
    "unique_sessions": 892
  },
  "breakdown": {
    "by_channel": [
      {"channel": "telegram", "messages": 5604, "tokens": 950000},
      {"channel": "slack", "messages": 3736, "tokens": 620000},
      {"channel": "discord", "messages": 1868, "tokens": 350000},
      {"channel": "web", "messages": 1245, "tokens": 202221}
    ],
    "by_model": [
      {"model": "gpt-4o-mini", "messages": 6476, "tokens": 1100000, "pct": 52},
      {"model": "claude-sonnet-4", "messages": 3487, "tokens": 600000, "pct": 28},
      {"model": "gpt-4o", "messages": 1868, "tokens": 380000, "pct": 15},
      {"model": "other", "messages": 622, "tokens": 42221, "pct": 5}
    ]
  },
  "timeline": [
    {
      "timestamp": "2024-01-15T00:00:00Z",
      "messages": 234,
      "tokens_in": 23456,
      "tokens_out": 12345,
      "api_calls": 112
    },
    {
      "timestamp": "2024-01-15T01:00:00Z",
      "messages": 189,
      "tokens_in": 19876,
      "tokens_out": 10234,
      "api_calls": 89
    }
  ]
}
```

### 4.2 GET /api/v1/analytics/performance

Retrieve performance metrics.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start` | string | -24h | Start timestamp |
| `end` | string | now | End timestamp |
| `resolution` | string | hour | Aggregation level |
| `percentiles` | string | 50,95,99 | Latency percentiles |

**Response:**

```json
{
  "period": {
    "start": "2024-01-15T00:00:00Z",
    "end": "2024-01-15T23:59:59Z"
  },
  "summary": {
    "total_requests": 4521,
    "successful_requests": 4412,
    "failed_requests": 109,
    "success_rate": 0.9759,
    "avg_latency_ms": 1823,
    "latency_percentiles": {
      "p50": 1234,
      "p95": 3456,
      "p99": 5678
    },
    "avg_ttft_ms": 234,
    "avg_tokens_per_second": 45.6,
    "retry_count": 89,
    "cooldown_count": 3
  },
  "by_provider": [
    {
      "provider": "openai",
      "requests": 2345,
      "success_rate": 0.982,
      "avg_latency_ms": 1654,
      "p95_latency_ms": 3123,
      "error_count": 42
    },
    {
      "provider": "anthropic",
      "requests": 1567,
      "success_rate": 0.989,
      "avg_latency_ms": 1876,
      "p95_latency_ms": 3456,
      "error_count": 17
    }
  ],
  "by_tool": [
    {
      "tool": "shell",
      "invocations": 456,
      "avg_duration_ms": 2345,
      "error_count": 12
    },
    {
      "tool": "web",
      "invocations": 234,
      "avg_duration_ms": 1876,
      "error_count": 8
    }
  ],
  "timeline": [
    {
      "timestamp": "2024-01-15T00:00:00Z",
      "requests": 112,
      "avg_latency_ms": 1654,
      "p95_latency_ms": 3234,
      "error_count": 3
    }
  ],
  "errors": [
    {"code": "rate_limit", "count": 45, "pct": 41.3},
    {"code": "timeout", "count": 32, "pct": 29.4},
    {"code": "context_length", "count": 21, "pct": 19.3},
    {"code": "other", "count": 11, "pct": 10.0}
  ]
}
```

### 4.3 GET /api/v1/analytics/costs

Retrieve cost analysis and projections.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start` | string | -30d | Start timestamp |
| `end` | string | now | End timestamp |
| `resolution` | string | day | Aggregation level |
| `currency` | string | USD | Currency code |

**Response:**

```json
{
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "currency": "USD",
  "summary": {
    "total_cost": 245.67,
    "total_tokens_in": 12456789,
    "total_tokens_out": 8765432,
    "avg_cost_per_message": 0.0197,
    "avg_cost_per_session": 0.2756,
    "projected_monthly": 312.00
  },
  "by_model": [
    {
      "model": "gpt-4o",
      "provider": "openai",
      "tokens_in": 4567890,
      "tokens_out": 2345678,
      "cost": 124.50,
      "pct_of_total": 50.7,
      "input_rate": 5.00,
      "output_rate": 15.00
    },
    {
      "model": "claude-sonnet-4",
      "provider": "anthropic",
      "tokens_in": 3456789,
      "tokens_out": 3456789,
      "cost": 78.20,
      "pct_of_total": 31.8,
      "input_rate": 3.00,
      "output_rate": 15.00
    }
  ],
  "by_channel": [
    {"channel": "telegram", "cost": 110.55, "pct": 45.0},
    {"channel": "slack", "cost": 73.70, "pct": 30.0},
    {"channel": "discord", "cost": 36.85, "pct": 15.0},
    {"channel": "web", "cost": 24.57, "pct": 10.0}
  ],
  "savings": {
    "fallback_savings": 45.20,
    "caching_savings": 12.80,
    "local_model_savings": 89.00,
    "total_savings": 147.00,
    "savings_pct": 37.5
  },
  "timeline": [
    {
      "date": "2024-01-01",
      "cost": 8.45,
      "tokens_in": 401234,
      "tokens_out": 287654
    },
    {
      "date": "2024-01-02",
      "cost": 9.12,
      "tokens_in": 432123,
      "tokens_out": 298765
    }
  ],
  "pricing": {
    "last_updated": "2024-01-10T00:00:00Z",
    "models": [
      {
        "model": "gpt-4o",
        "input_per_million": 5.00,
        "output_per_million": 15.00
      },
      {
        "model": "gpt-4o-mini",
        "input_per_million": 0.15,
        "output_per_million": 0.60
      },
      {
        "model": "claude-sonnet-4",
        "input_per_million": 3.00,
        "output_per_million": 15.00
      }
    ]
  }
}
```

### 4.4 GET /api/v1/analytics/engagement

Retrieve user engagement metrics.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `start` | string | -30d | Start timestamp |
| `end` | string | now | End timestamp |

**Response:**

```json
{
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "summary": {
    "dau": 156,
    "mau": 892,
    "dau_mau_ratio": 0.175,
    "total_sessions": 2345,
    "avg_session_length": 5.2,
    "avg_messages_per_user": 13.9,
    "peak_concurrent": 45
  },
  "retention": {
    "day_1": 0.45,
    "day_7": 0.28,
    "day_30": 0.15
  },
  "by_channel": [
    {
      "channel": "telegram",
      "dau": 67,
      "mau": 312,
      "sessions": 1023,
      "avg_session_length": 4.8
    },
    {
      "channel": "slack",
      "dau": 45,
      "mau": 234,
      "sessions": 678,
      "avg_session_length": 5.6
    }
  ],
  "timeline": [
    {
      "date": "2024-01-01",
      "dau": 123,
      "new_users": 12,
      "returning_users": 111,
      "sessions": 89
    }
  ]
}
```

### 4.5 GET /api/v1/analytics/realtime

Retrieve real-time metrics (WebSocket or polling).

**Response:**

```json
{
  "timestamp": "2024-01-15T14:32:45Z",
  "metrics": {
    "active_sessions": 23,
    "messages_last_minute": 12,
    "messages_last_hour": 456,
    "tokens_last_minute": 4567,
    "tokens_last_hour": 234567,
    "avg_latency_last_5m_ms": 1234,
    "error_rate_last_5m": 0.012,
    "requests_in_flight": 3
  },
  "channels": {
    "telegram": {"active": 12, "queued": 2},
    "slack": {"active": 8, "queued": 0},
    "discord": {"active": 3, "queued": 1}
  },
  "providers": {
    "openai": {"status": "healthy", "latency_ms": 1234},
    "anthropic": {"status": "healthy", "latency_ms": 1456}
  }
}
```

---

## 5. Implementation Approach

### 5.1 Go Packages

#### Package Structure

```
pkg/
├── analytics/
│   ├── collector.go       # Event collection from EventBus
│   ├── aggregator.go      # Real-time aggregation
│   ├── processor.go       # Batch processing
│   ├── storage.go         # SQLite operations
│   ├── cache.go           # Redis operations
│   ├── calculator.go      # Cost calculations
│   ├── exporter.go        # CSV/JSON export
│   └── models.go          # Data structures
├── pricing/
│   ├── models.go          # Pricing model definitions
│   ├── fetcher.go         # Fetch pricing from providers
│   └── calculator.go      # Cost estimation
└── dashboard/
    └── handlers.go        # HTTP handlers for analytics API
```

#### Core Package: `pkg/analytics/models.go`

```go
package analytics

import "time"

type MetricType string

const (
    MetricTypeCounter MetricType = "counter"
    MetricTypeGauge   MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
)

type Event struct {
    Timestamp   time.Time          `json:"timestamp"`
    EventType   string             `json:"event_type"`
    AgentID     string             `json:"agent_id"`
    SessionID   string             `json:"session_id,omitempty"`
    Channel     string             `json:"channel,omitempty"`
    ChatID      string             `json:"chat_id,omitempty"`
    Model       string             `json:"model,omitempty"`
    Provider    string             `json:"provider,omitempty"`
    TokensIn    int                `json:"tokens_in,omitempty"`
    TokensOut   int                `json:"tokens_out,omitempty"`
    DurationMs  int                `json:"duration_ms,omitempty"`
    Cost        float64            `json:"cost,omitempty"`
    Success     bool               `json:"success"`
    ErrorCode   string             `json:"error_code,omitempty"`
    Metadata    map[string]any     `json:"metadata,omitempty"`
}

type UsageSummary struct {
    MessagesTotal   int64   `json:"messages_total"`
    MessagesUser    int64   `json:"messages_user"`
    MessagesAssistant int64 `json:"messages_assistant"`
    TokensInput     int64   `json:"tokens_input"`
    TokensOutput    int64   `json:"tokens_output"`
    TokensTotal     int64   `json:"tokens_total"`
    APICalls        int64   `json:"api_calls"`
    UniqueSessions  int64   `json:"unique_sessions"`
}

type PerformanceSummary struct {
    TotalRequests      int64              `json:"total_requests"`
    SuccessfulRequests int64              `json:"successful_requests"`
    FailedRequests     int64              `json:"failed_requests"`
    SuccessRate        float64            `json:"success_rate"`
    AvgLatencyMs       int64              `json:"avg_latency_ms"`
    LatencyPercentiles map[string]int64   `json:"latency_percentiles"`
    AvgTTFTMs          int64              `json:"avg_ttft_ms"`
    AvgTPS             float64            `json:"avg_tokens_per_second"`
    RetryCount         int64              `json:"retry_count"`
    CooldownCount      int64              `json:"cooldown_count"`
}

type CostSummary struct {
    TotalCost        float64            `json:"total_cost"`
    TotalTokensIn    int64              `json:"total_tokens_in"`
    TotalTokensOut   int64              `json:"total_tokens_out"`
    AvgCostPerMsg    float64            `json:"avg_cost_per_message"`
    AvgCostPerSession float64           `json:"avg_cost_per_session"`
    ProjectedMonthly float64            `json:"projected_monthly"`
}

type EngagementSummary struct {
    DAU                 int64     `json:"dau"`
    MAU                 int64     `json:"mau"`
    DAUMAURatio         float64   `json:"dau_mau_ratio"`
    TotalSessions       int64     `json:"total_sessions"`
    AvgSessionLength    float64   `json:"avg_session_length"`
    AvgMessagesPerUser  float64   `json:"avg_messages_per_user"`
    PeakConcurrent      int64     `json:"peak_concurrent"`
}

type QueryParams struct {
    Start       time.Time
    End         time.Time
    Resolution  string
    Channel     string
    AgentID     string
    Model       string
    Percentiles []int
}
```

#### Collector: `pkg/analytics/collector.go`

```go
package analytics

import (
    "context"
    "github.com/raynaythegreat/ai-business-hq/pkg/agent"
)

type Collector struct {
    eventBus   *agent.EventBus
    processor  *Processor
    enabled    bool
}

func NewCollector(eventBus *agent.EventBus, processor *Processor) *Collector {
    return &Collector{
        eventBus:  eventBus,
        processor: processor,
        enabled:   true,
    }
}

func (c *Collector) Start(ctx context.Context) error {
    ch := c.eventBus.Subscribe()
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case event := <-ch:
                if c.enabled {
                    c.processEvent(event)
                }
            }
        }
    }()
    return nil
}

func (c *Collector) processEvent(event agent.Event) {
    analyticsEvent := c.convertEvent(event)
    c.processor.Record(analyticsEvent)
}

func (c *Collector) convertEvent(e agent.Event) Event {
    evt := Event{
        Timestamp: e.Time,
        EventType: e.Kind.String(),
        AgentID:   e.Meta.AgentID,
        SessionID: e.Meta.SessionKey,
    }
    
    switch p := e.Payload.(type) {
    case agent.LLMRequestPayload:
        evt.Model = p.Model
        evt.EventType = "llm_request"
    case agent.LLMResponsePayload:
        evt.EventType = "llm_response"
        evt.TokensOut = p.ContentLen
    case agent.ToolExecStartPayload:
        evt.EventType = "tool_exec_start"
        evt.Metadata = map[string]any{"tool": p.Tool}
    case agent.ToolExecEndPayload:
        evt.EventType = "tool_exec_end"
        evt.DurationMs = int(p.Duration.Milliseconds())
        evt.Success = !p.IsError
    case agent.TurnEndPayload:
        evt.EventType = "turn_end"
        evt.DurationMs = int(p.Duration.Milliseconds())
        evt.Success = p.Status == agent.TurnEndStatusCompleted
    }
    
    if e.Meta.Source != "" {
        evt.Channel = e.Meta.Source
    }
    
    return evt
}
```

### 5.2 Frontend Chart Library Recommendations

| Library | Size | Pros | Cons | Recommendation |
|---------|------|------|------|----------------|
| **Recharts** | ~45KB | React-native, composable, simple API | Less customizable | **Primary** |
| **Chart.js** | ~65KB | Mature, lots of chart types | Imperative API | Alternative |
| **Apache ECharts** | ~300KB | Feature-rich, beautiful defaults | Large bundle | For complex needs |
| **Visx** | ~30KB | Low-level, highly customizable | More code needed | Advanced use |

#### Recommended Stack (MVP)

```typescript
import {
  LineChart, Line, AreaChart, Area, BarChart, Bar,
  PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid,
  Tooltip, Legend, ResponsiveContainer
} from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

const UsageChart = ({ data }) => (
  <ResponsiveContainer width="100%" height={300}>
    <AreaChart data={data}>
      <CartesianGrid strokeDasharray="3 3" />
      <XAxis dataKey="timestamp" />
      <YAxis />
      <Tooltip />
      <Area type="monotone" dataKey="messages" stroke="#8884d8" fill="#8884d8" />
    </AreaChart>
  </ResponsiveContainer>
);
```

### 5.3 Caching Strategy

#### Multi-Layer Cache

```
┌─────────────────────────────────────────────────────────────────┐
│                      Cache Layers                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐     │
│  │  Browser     │    │  API Server  │    │  Redis       │     │
│  │  Cache       │───▶│  In-Memory   │───▶│  (shared)    │     │
│  │  (5min)      │    │  (1min)      │    │  (1hour)     │     │
│  └──────────────┘    └──────────────┘    └──────────────┘     │
│                                                │                │
│                                                ▼                │
│                                        ┌──────────────┐        │
│                                        │  SQLite      │        │
│                                        │  (source)    │        │
│                                        └──────────────┘        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Cache Keys

```go
const (
    CacheKeyUsagePrefix    = "analytics:usage"
    CacheKeyPerfPrefix     = "analytics:perf"
    CacheKeyCostPrefix     = "analytics:cost"
    CacheKeyRealtimePrefix = "analytics:realtime"
)

func buildCacheKey(prefix, resolution string, start, end time.Time, filters map[string]string) string {
    hash := sha256.Sum256([]byte(fmt.Sprintf("%v", filters)))
    return fmt.Sprintf("%s:%s:%s:%s:%x",
        prefix,
        resolution,
        start.Format("20060102"),
        end.Format("20060102"),
        hash[:8],
    )
}
```

#### Cache TTLs

| Data Type | TTL | Invalidation |
|-----------|-----|--------------|
| Real-time metrics | 5 seconds | Time-based |
| Hourly aggregates | 5 minutes | Time-based |
| Daily aggregates | 1 hour | On new data |
| Historical data | 24 hours | On-demand |
| Export files | 1 hour | On generation |

#### Implementation: `pkg/analytics/cache.go`

```go
package analytics

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewCache(addr string) *Cache {
    return &Cache{
        client: redis.NewClient(&redis.Options{Addr: addr}),
        ttl:    5 * time.Minute,
    }
}

func (c *Cache) Get(ctx context.Context, key string, dest any) bool {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return false
    }
    return json.Unmarshal([]byte(val), dest) == nil
}

func (c *Cache) Set(ctx context.Context, key string, val any) error {
    data, err := json.Marshal(val)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Increment(ctx context.Context, key string, field string, delta int64) error {
    return c.client.HIncrBy(ctx, key, field, delta).Err()
}
```

### 5.4 MVP Implementation Phases

#### Phase 1: Foundation (Week 1-2)

- [ ] Create `pkg/analytics` package structure
- [ ] Implement event collector with EventBus integration
- [ ] SQLite schema migration
- [ ] Basic storage layer

#### Phase 2: Core Metrics (Week 3-4)

- [ ] Usage metrics collection and API
- [ ] Performance metrics collection and API
- [ ] Cost calculation engine
- [ ] Basic caching

#### Phase 3: Dashboard (Week 5-6)

- [ ] Frontend dashboard scaffolding
- [ ] Usage charts and visualizations
- [ ] Performance monitoring views
- [ ] Cost analysis views

#### Phase 4: Advanced Features (Week 7-8)

- [ ] Real-time metrics WebSocket
- [ ] Export functionality
- [ ] Retention policies
- [ ] Alerting thresholds

---

## 6. Configuration

### 6.1 Analytics Configuration Schema

```json
{
  "analytics": {
    "enabled": true,
    "storage": {
      "type": "sqlite",
      "path": "./data/analytics.db",
      "retention_days": 90
    },
    "cache": {
      "enabled": true,
      "redis_url": "redis://localhost:6379/1",
      "ttl_seconds": 300
    },
    "collection": {
      "sample_rate": 1.0,
      "exclude_channels": [],
      "exclude_events": []
    },
    "pricing": {
      "auto_update": true,
      "update_interval_hours": 24,
      "custom_rates": {
        "custom-model": {
          "input_per_million": 1.00,
          "output_per_million": 2.00
        }
      }
    },
    "dashboard": {
      "default_period": "7d",
      "refresh_interval_seconds": 30
    }
  }
}
```

### 6.2 Environment Variables

```bash
AIBHQ_ANALYTICS_ENABLED=true
AIBHQ_ANALYTICS_DB_PATH=./data/analytics.db
AIBHQ_ANALYTICS_REDIS_URL=redis://localhost:6379/1
AIBHQ_ANALYTICS_RETENTION_DAYS=90
AIBHQ_ANALYTICS_SAMPLE_RATE=1.0
```

---

## 7. Future Enhancements

- **Custom Dashboards**: User-configurable widgets and layouts
- **Alerting**: Threshold-based notifications via channels
- **Anomaly Detection**: ML-based usage pattern detection
- **Team Analytics**: Multi-tenant usage breakdown
- **API Usage Quotas**: Per-user/team rate limiting
- **Budget Controls**: Cost alerts and automatic throttling
- **A/B Testing**: Model comparison analytics
- **Data Warehouse Export**: BigQuery/Snowflake integration
