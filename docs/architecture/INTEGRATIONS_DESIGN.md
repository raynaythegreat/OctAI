# Integrations & API Marketplace Design

> **Document Status**: Draft  
> **Version**: 1.0.0  
> **Last Updated**: March 2026  
> **Authors**: AI Business HQ Team

---

## Table of Contents

1. [Overview](#1-overview)
2. [Priority Integrations (Phase 1)](#2-priority-integrations-phase-1)
3. [Integration Architecture](#3-integration-architecture)
4. [Integration SDK Design](#4-integration-sdk-design)
5. [API Marketplace](#5-api-marketplace)
6. [Implementation Priority](#6-implementation-priority)
7. [Security Considerations](#7-security-considerations)
8. [Appendices](#appendices)

---

## 1. Overview

### 1.1 Purpose

This document defines the architecture for AI Business HQ's third-party integrations and API marketplace, enabling seamless connectivity with business-critical tools while providing a public API for developers.

### 1.2 Goals

- **Connectivity**: Integrate with 10+ business tools in Phase 1
- **Extensibility**: Plugin-based architecture for easy integration additions
- **Developer Experience**: Comprehensive SDK and documentation
- **Security**: Enterprise-grade OAuth, API key management, and rate limiting
- **Scalability**: Handle 1000+ requests/second per integration

### 1.3 Non-Goals

- Custom integration builder UI (Phase 2)
- White-label marketplace (Phase 3)
- Real-time bi-directional sync (all integrations)

---

## 2. Priority Integrations (Phase 1)

### 2.1 Integration Matrix

| Category | Integration | Priority | Auth Type | Sync Type | Est. Effort |
|----------|-------------|----------|-----------|-----------|-------------|
| **CRM** | Salesforce | P0 | OAuth 2.0 | Polling + Webhooks | 3 weeks |
| **CRM** | HubSpot | P0 | OAuth 2.0 | Webhooks | 2 weeks |
| **Project Mgmt** | Linear | P0 | OAuth 2.0 | Webhooks | 2 weeks |
| **Project Mgmt** | Jira | P1 | OAuth 2.0 | Webhooks | 3 weeks |
| **Project Mgmt** | Asana | P1 | OAuth 2.0 | Webhooks | 2 weeks |
| **Communication** | Slack | P0 | OAuth 2.0 | Webhooks | 2 weeks |
| **Communication** | Microsoft Teams | P1 | OAuth 2.0 | Webhooks | 3 weeks |
| **Finance** | Stripe | P0 | API Key | Webhooks | 2 weeks |
| **Finance** | QuickBooks | P1 | OAuth 2.0 | Polling | 3 weeks |

### 2.2 Detailed Integration Specs

#### 2.2.1 Salesforce

```yaml
integration:
  name: salesforce
  display_name: Salesforce CRM
  category: crm
  version: "1.0.0"
  
auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - api
    - refresh_token
    - offline_access
  endpoints:
    authorize: https://login.salesforce.com/services/oauth2/authorize
    token: https://login.salesforce.com/services/oauth2/token
    sandbox_authorize: https://test.salesforce.com/services/oauth2/authorize
    sandbox_token: https://test.salesforce.com/services/oauth2/token

capabilities:
  - contacts.read
  - contacts.write
  - leads.read
  - leads.write
  - opportunities.read
  - opportunities.write
  - accounts.read
  - accounts.write
  - cases.read
  - cases.write
  - custom_objects.read
  - custom_objects.write

webhooks:
  supported: true
  events:
    - create
    - update
    - delete
    - undelete

rate_limits:
  requests_per_day: 100000
  concurrent_requests: 25
```

#### 2.2.2 HubSpot

```yaml
integration:
  name: hubspot
  display_name: HubSpot CRM
  category: crm
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - crm.objects.contacts.read
    - crm.objects.contacts.write
    - crm.objects.companies.read
    - crm.objects.companies.write
    - crm.objects.deals.read
    - crm.objects.deals.write
    - crm.objects.owners.read
  endpoints:
    authorize: https://app.hubspot.com/oauth/authorize
    token: https://api.hubapi.com/oauth/v1/token

capabilities:
  - contacts.read
  - contacts.write
  - companies.read
  - companies.write
  - deals.read
  - deals.write
  - tickets.read
  - tickets.write
  - engagements.read
  - engagements.write

webhooks:
  supported: true
  subscription_url: https://api.hubapi.com/webhooks/v1/{appId}/subscriptions
  events:
    - contact.creation
    - contact.deletion
    - contact.propertyChange
    - deal.creation
    - deal.propertyChange
```

#### 2.2.3 Linear

```yaml
integration:
  name: linear
  display_name: Linear
  category: project_management
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - read
    - write
    - issues:create
    - comments:create
  endpoints:
    authorize: https://linear.app/oauth/authorize
    token: https://api.linear.app/oauth/token

api:
  type: graphql
  endpoint: https://api.linear.app/graphql

capabilities:
  - issues.read
  - issues.write
  - projects.read
  - projects.write
  - teams.read
  - cycles.read
  - comments.read
  - comments.write
  - labels.read
  - workflows.read

webhooks:
  supported: true
  events:
    - Issue.create
    - Issue.update
    - Issue.delete
    - Comment.create
    - Project.create
    - Project.update
```

#### 2.2.4 Jira

```yaml
integration:
  name: jira
  display_name: Jira
  category: project_management
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - read:jira-work
    - write:jira-work
    - offline_access
  endpoints:
    authorize: https://auth.atlassian.com/authorize
    token: https://auth.atlassian.com/oauth/token
  cloud_id_required: true

api:
  type: rest
  base_url: https://api.atlassian.com/ex/jira/{cloudId}/rest/api/3

capabilities:
  - issues.read
  - issues.write
  - projects.read
  - projects.write
  - boards.read
  - sprints.read
  - users.read
  - comments.read
  - comments.write
  - attachments.read
  - attachments.write

webhooks:
  supported: true
  events:
    - jira:issue_created
    - jira:issue_updated
    - jira:issue_deleted
    - comment_created
    - comment_updated
```

#### 2.2.5 Asana

```yaml
integration:
  name: asana
  display_name: Asana
  category: project_management
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - default
  endpoints:
    authorize: https://app.asana.com/-/oauth_authorize
    token: https://app.asana.com/-/oauth_token

api:
  type: rest
  base_url: https://app.asana.com/api/1.0

capabilities:
  - tasks.read
  - tasks.write
  - projects.read
  - projects.write
  - workspaces.read
  - teams.read
  - sections.read
  - stories.read
  - stories.write
  - attachments.read
  - attachments.write

webhooks:
  supported: true
  events:
    - task_added
    - task_changed
    - task_deleted
    - task_completed
```

#### 2.2.6 Slack

```yaml
integration:
  name: slack
  display_name: Slack
  category: communication
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    bot:
      - channels:read
      - channels:history
      - channels:manage
      - chat:write
      - chat:write.public
      - groups:read
      - groups:history
      - groups:write
      - im:read
      - im:history
      - im:write
      - mpim:read
      - mpim:history
      - mpim:write
      - users:read
      - users:read.email
      - team:read
      - files:read
      - files:write
      - reactions:read
      - reactions:write
  endpoints:
    authorize: https://slack.com/oauth/v2/authorize
    token: https://slack.com/api/oauth.v2.access

api:
  type: rest
  base_url: https://slack.com/api

capabilities:
  - messages.read
  - messages.write
  - channels.read
  - channels.write
  - users.read
  - files.read
  - files.write
  - reactions.read
  - reactions.write
  - search.read

webhooks:
  supported: true
  type: events_api
  events:
    - message.channels
    - message.groups
    - message.im
    - message.mpim
    - reaction_added
    - reaction_removed
    - channel_created
    - member_joined_channel
```

#### 2.2.7 Microsoft Teams

```yaml
integration:
  name: teams
  display_name: Microsoft Teams
  category: communication
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - Channel.ReadBasic.All
    - ChannelMessage.Read.All
    - ChannelMessage.Send
    - Chat.Read
    - Chat.ReadWrite
    - ChatMessage.Read
    - Team.ReadBasic.All
    - User.Read.All
    - Files.Read.All
    - Files.ReadWrite.All
  endpoints:
    authorize: https://login.microsoftonline.com/{tenantId}/oauth2/v2.0/authorize
    token: https://login.microsoftonline.com/{tenantId}/oauth2/v2.0/token

api:
  type: rest
  base_url: https://graph.microsoft.com/v1.0

capabilities:
  - messages.read
  - messages.write
  - channels.read
  - channels.write
  - teams.read
  - chats.read
  - chats.write
  - users.read
  - files.read
  - files.write

webhooks:
  supported: true
  type: change_notifications
  events:
    - channelCreated
    - channelDeleted
    - messageCreated
    - messageUpdated
    - messageDeleted
```

#### 2.2.8 Stripe

```yaml
integration:
  name: stripe
  display_name: Stripe
  category: finance
  version: "1.0.0"

auth:
  type: api_key
  header: Authorization
  prefix: Bearer
  modes:
    - secret_key
    - restricted_key

api:
  type: rest
  base_url: https://api.stripe.com/v1
  stripe_version: "2024-01-01"

capabilities:
  - customers.read
  - customers.write
  - charges.read
  - charges.write
  - invoices.read
  - invoices.write
  - subscriptions.read
  - subscriptions.write
  - products.read
  - products.write
  - prices.read
  - prices.write
  - payment_intents.read
  - payment_intents.write
  - refunds.read
  - refunds.write
  - balance.read
  - payouts.read

webhooks:
  supported: true
  signature_verification: true
  events:
    - customer.created
    - customer.updated
    - customer.deleted
    - invoice.paid
    - invoice.payment_failed
    - payment_intent.succeeded
    - payment_intent.payment_failed
    - charge.succeeded
    - charge.failed
    - subscription.created
    - subscription.updated
    - subscription.deleted
```

#### 2.2.9 QuickBooks

```yaml
integration:
  name: quickbooks
  display_name: QuickBooks Online
  category: finance
  version: "1.0.0"

auth:
  type: oauth2
  grant_type: authorization_code
  scopes:
    - com.intuit.quickbooks.accounting
  endpoints:
    authorize: https://appcenter.intuit.com/connect/oauth2
    token: https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer
    reconnect: https://appcenter.intuit.com/connect/oauth2

api:
  type: rest
  base_url: https://quickbooks.api.intuit.com/v3/company/{companyId}
  sandbox_url: https://sandbox-quickbooks.api.intuit.com/v3/company/{companyId}

capabilities:
  - customers.read
  - customers.write
  - invoices.read
  - invoices.write
  - payments.read
  - payments.write
  - accounts.read
  - items.read
  - items.write
  - vendors.read
  - vendors.write
  - bills.read
  - bills.write
  - purchase_orders.read
  - reports.read

webhooks:
  supported: true
  events:
    - CREATE
    - UPDATE
    - DELETE
    - MERGE
    - VOID

sync:
  type: polling
  recommended_interval: 300s
```

---

## 3. Integration Architecture

### 3.1 System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           AI Business HQ Core                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Agent     │    │   Skills    │    │   Memory    │    │   Tools     │  │
│  │   Engine    │    │   System    │    │   Store     │    │   Registry  │  │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘  │
│         │                  │                  │                  │         │
│         └──────────────────┴──────────────────┴──────────────────┘         │
│                                     │                                        │
│                          ┌──────────▼──────────┐                            │
│                          │  Integration Hub    │                            │
│                          │  (pkg/integration)  │                            │
│                          └──────────┬──────────┘                            │
│                                     │                                        │
│  ┌──────────────────────────────────┼──────────────────────────────────┐   │
│  │                                  │                                  │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │   │
│  │  │   OAuth     │  │   Webhook   │  │   API Key   │  │   Rate      │ │   │
│  │  │   Manager   │  │   Router    │  │   Manager   │  │   Limiter   │ │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘ │   │
│  │                                                                      │   │
│  │                     Integration Runtime Layer                        │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                     │                                        │
└─────────────────────────────────────┼────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
              ┌─────▼─────┐    ┌─────▼─────┐    ┌─────▼─────┐
              │  CRM      │    │  Project  │    │  Finance  │
              │  Adapters │    │  Mgmt     │    │  Adapters │
              │           │    │  Adapters │    │           │
              └─────┬─────┘    └─────┬─────┘    └─────┬─────┘
                    │                 │                 │
              ┌─────▼─────┐    ┌─────▼─────┐    ┌─────▼─────┐
              │Salesforce │    │  Linear   │    │  Stripe   │
              │  HubSpot  │    │   Jira    │    │QuickBooks │
              └───────────┘    │   Asana   │    └───────────┘
                               └───────────┘
```

### 3.2 Core Components

#### 3.2.1 Integration Hub

```go
// pkg/integration/hub.go
package integration

import (
    "context"
    "sync"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/auth"
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/webhook"
)

type Hub struct {
    mu          sync.RWMutex
    adapters    map[string]adapter.Adapter
    registry    *adapter.Registry
    oauthMgr    *auth.OAuthManager
    webhookMgr  *webhook.Manager
    rateLimiter *RateLimiter
    eventBus    chan IntegrationEvent
}

type Config struct {
    OAuthStore       auth.TokenStore
    WebhookSecretKey string
    RateLimitConfig  RateLimitConfig
    EventBusSize     int
}

func NewHub(cfg Config) *Hub {
    return &Hub{
        adapters:    make(map[string]adapter.Adapter),
        registry:    adapter.NewRegistry(),
        oauthMgr:    auth.NewOAuthManager(cfg.OAuthStore),
        webhookMgr:  webhook.NewManager(cfg.WebhookSecretKey),
        rateLimiter: NewRateLimiter(cfg.RateLimitConfig),
        eventBus:    make(chan IntegrationEvent, cfg.EventBusSize),
    }
}

func (h *Hub) Register(name string, a adapter.Adapter) error {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if _, exists := h.adapters[name]; exists {
        return ErrAdapterAlreadyRegistered
    }
    
    h.adapters[name] = a
    h.registry.Register(name, a.Schema())
    
    return nil
}

func (h *Hub) Get(name string) (adapter.Adapter, bool) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    a, ok := h.adapters[name]
    return a, ok
}

func (h *Hub) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    a, ok := h.Get(req.Integration)
    if !ok {
        return nil, ErrIntegrationNotFound
    }
    
    if err := h.rateLimiter.Wait(ctx, req.Integration, req.TenantID); err != nil {
        return nil, err
    }
    
    resp, err := a.Execute(ctx, req)
    if err != nil {
        return nil, err
    }
    
    h.eventBus <- IntegrationEvent{
        Type:        EventTypeExecute,
        Integration: req.Integration,
        TenantID:    req.TenantID,
        Action:      req.Action,
        Timestamp:   time.Now(),
    }
    
    return resp, nil
}

type ExecuteRequest struct {
    Integration string
    TenantID    string
    Action      string
    Params      map[string]any
    Options     ExecuteOptions
}

type ExecuteOptions struct {
    IdempotencyKey string
    Timeout        time.Duration
    RetryCount     int
}

type ExecuteResponse struct {
    Data       any
    Metadata   map[string]string
    Pagination *Pagination
}

type IntegrationEvent struct {
    Type        EventType
    Integration string
    TenantID    string
    Action      string
    Timestamp   time.Time
    Payload     any
}

type EventType string

const (
    EventTypeExecute  EventType = "execute"
    EventTypeWebhook  EventType = "webhook"
    EventTypeSync     EventType = "sync"
    EventTypeError    EventType = "error"
)
```

### 3.3 Webhook System Design

#### 3.3.1 Webhook Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Third-Party Service                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │ POST /webhooks/{integration}/{tenant}
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Webhook Receiver                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Signature  │  │   Rate      │  │   Request   │             │
│  │  Verify     │  │   Limit     │  │   Parser    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Webhook Queue                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐       │   │
│  │  │ W1  │ │ W2  │ │ W3  │ │ W4  │ │ W5  │ │ ... │       │   │
│  │  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘       │   │
│  └─────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Webhook Processor                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Transform  │─▶│  Validate   │─▶│  Dispatch   │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Event Handlers                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Skill     │  │   Agent     │  │   Custom    │             │
│  │   Trigger   │  │   Notify    │  │   Handler   │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

#### 3.3.2 Webhook Implementation

```go
// pkg/integration/webhook/manager.go
package webhook

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
)

type Manager struct {
    secretKey    string
    store        WebhookStore
    queue        chan *WebhookEvent
    processors   int
    handlers     map[string]EventHandler
}

type WebhookConfig struct {
    SecretKey  string
    Store      WebhookStore
    QueueSize  int
    Processors int
}

func NewManager(secretKey string) *Manager {
    return &Manager{
        secretKey:  secretKey,
        queue:      make(chan *WebhookEvent, 1000),
        processors: 10,
        handlers:   make(map[string]EventHandler),
    }
}

func (m *Manager) RegisterHandler(integration string, h EventHandler) {
    m.handlers[integration] = h
}

func (m *Manager) Handler(integration string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.PathValue("tenant")
        if tenantID == "" {
            http.Error(w, "missing tenant", http.StatusBadRequest)
            return
        }
        
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "failed to read body", http.StatusBadRequest)
            return
        }
        
        if !m.verifySignature(r, body, integration) {
            http.Error(w, "invalid signature", http.StatusUnauthorized)
            return
        }
        
        event := &WebhookEvent{
            ID:          generateID(),
            Integration: integration,
            TenantID:    tenantID,
            Headers:     r.Header.Clone(),
            Body:        body,
            ReceivedAt:  time.Now(),
        }
        
        select {
        case m.queue <- event:
            w.WriteHeader(http.StatusAccepted)
        default:
            http.Error(w, "queue full", http.StatusServiceUnavailable)
        }
    }
}

func (m *Manager) verifySignature(r *http.Request, body []byte, integration string) bool {
    signature := r.Header.Get("X-Signature-256")
    if signature == "" {
        signature = r.Header.Get("X-Hub-Signature-256")
    }
    if signature == "" {
        signature = r.Header.Get("Stripe-Signature")
    }
    
    if signature == "" {
        return false
    }
    
    expectedSig := m.computeSignature(body, integration)
    
    sig := strings.TrimPrefix(signature, "sha256=")
    return hmac.Equal([]byte(sig), []byte(expectedSig))
}

func (m *Manager) computeSignature(body []byte, integration string) string {
    h := hmac.New(sha256.New, []byte(m.secretKey))
    h.Write(body)
    return hex.EncodeToString(h.Sum(nil))
}

func (m *Manager) Start(ctx context.Context) {
    for i := 0; i < m.processors; i++ {
        go m.processor(ctx)
    }
}

func (m *Manager) processor(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-m.queue:
            m.processEvent(ctx, event)
        }
    }
}

func (m *Manager) processEvent(ctx context.Context, event *WebhookEvent) {
    handler, ok := m.handlers[event.Integration]
    if !ok {
        handler = m.defaultHandler
    }
    
    transformed, err := handler.Transform(ctx, event)
    if err != nil {
        m.store.RecordFailure(event.ID, err)
        return
    }
    
    if err := handler.Handle(ctx, transformed); err != nil {
        m.store.RecordFailure(event.ID, err)
        return
    }
    
    m.store.RecordSuccess(event.ID)
}

type WebhookEvent struct {
    ID          string
    Integration string
    TenantID    string
    Headers     http.Header
    Body        []byte
    ReceivedAt  time.Time
}

type TransformedEvent struct {
    ID          string
    Integration string
    TenantID    string
    EventType   string
    Resource    string
    Action      string
    Data        map[string]any
    Timestamp   time.Time
    Raw         *WebhookEvent
}

type EventHandler interface {
    Transform(ctx context.Context, event *WebhookEvent) (*TransformedEvent, error)
    Handle(ctx context.Context, event *TransformedEvent) error
}

type WebhookStore interface {
    RecordSuccess(eventID string)
    RecordFailure(eventID string, err error)
    GetStatus(eventID string) (*EventStatus, error)
}
```

### 3.4 OAuth Flow Implementation

#### 3.4.1 OAuth Manager

```go
// pkg/integration/auth/oauth.go
package auth

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
    
    "golang.org/x/oauth2"
)

type OAuthManager struct {
    store       TokenStore
    configs     map[string]*OAuthConfig
    http        *http.Client
}

type OAuthConfig struct {
    Provider         string
    ClientID         string
    ClientSecret     string
    AuthorizeURL     string
    TokenURL         string
    Scopes           []string
    RedirectURL      string
    PKCEEnabled      bool
    ExtraParams      map[string]string
}

type TokenStore interface {
    Store(ctx context.Context, tenantID, provider string, token *Token) error
    Get(ctx context.Context, tenantID, provider string) (*Token, error)
    Delete(ctx context.Context, tenantID, provider string) error
}

type Token struct    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    TokenType    string    `json:"token_type"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scope        string    `json:"scope"`
    Raw          json.RawMessage `json:"raw,omitempty"`
}

func NewOAuthManager(store TokenStore) *OAuthManager {
    return &OAuthManager{
        store:   store,
        configs: make(map[string]*OAuthConfig),
        http:    &http.Client{Timeout: 30 * time.Second},
    }
}

func (m *OAuthManager) Register(provider string, cfg *OAuthConfig) {
    m.configs[provider] = cfg
}

func (m *OAuthManager) GetAuthorizationURL(provider, tenantID, state string) (string, string, error) {
    cfg, ok := m.configs[provider]
    if !ok {
        return "", "", ErrProviderNotConfigured
    }
    
    pkceVerifier := ""
    if cfg.PKCEEnabled {
        pkceVerifier = generatePKCEVerifier()
    }
    
    stateToken := m.generateState(tenantID, provider, state, pkceVerifier)
    
    u, err := url.Parse(cfg.AuthorizeURL)
    if err != nil {
        return "", "", err
    }
    
    params := url.Values{
        "client_id":     {cfg.ClientID},
        "redirect_uri":  {cfg.RedirectURL},
        "response_type": {"code"},
        "scope":         {strings.Join(cfg.Scopes, " ")},
        "state":         {stateToken},
    }
    
    if cfg.PKCEEnabled {
        params.Set("code_challenge", computePKCEChallenge(pkceVerifier))
        params.Set("code_challenge_method", "S256")
    }
    
    for k, v := range cfg.ExtraParams {
        params.Set(k, v)
    }
    
    u.RawQuery = params.Encode()
    
    return u.String(), stateToken, nil
}

func (m *OAuthManager) HandleCallback(ctx context.Context, provider, code, state string) (*Token, error) {
    cfg, ok := m.configs[provider]
    if !ok {
        return nil, ErrProviderNotConfigured
    }
    
    stateData, err := m.verifyState(state)
    if err != nil {
        return nil, err
    }
    
    data := url.Values{
        "grant_type":   {"authorization_code"},
        "code":         {code},
        "redirect_uri": {cfg.RedirectURL},
        "client_id":    {cfg.ClientID},
    }
    
    if cfg.ClientSecret != "" {
        data.Set("client_secret", cfg.ClientSecret)
    }
    
    if cfg.PKCEEnabled && stateData.PKCEVerifier != "" {
        data.Set("code_verifier", stateData.PKCEVerifier)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    resp, err := m.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token exchange failed: %s", resp.Status)
    }
    
    var tokenResp struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        TokenType    string `json:"token_type"`
        ExpiresIn    int    `json:"expires_in"`
        Scope        string `json:"scope"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, err
    }
    
    token := &Token{
        AccessToken:  tokenResp.AccessToken,
        RefreshToken: tokenResp.RefreshToken,
        TokenType:    tokenResp.TokenType,
        Scope:        tokenResp.Scope,
    }
    
    if tokenResp.ExpiresIn > 0 {
        token.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
    }
    
    if err := m.store.Store(ctx, stateData.TenantID, provider, token); err != nil {
        return nil, err
    }
    
    return token, nil
}

func (m *OAuthManager) RefreshToken(ctx context.Context, tenantID, provider string) (*Token, error) {
    cfg, ok := m.configs[provider]
    if !ok {
        return nil, ErrProviderNotConfigured
    }
    
    token, err := m.store.Get(ctx, tenantID, provider)
    if err != nil {
        return nil, err
    }
    
    if token.RefreshToken == "" {
        return nil, ErrNoRefreshToken
    }
    
    data := url.Values{
        "grant_type":    {"refresh_token"},
        "refresh_token": {token.RefreshToken},
        "client_id":     {cfg.ClientID},
    }
    
    if cfg.ClientSecret != "" {
        data.Set("client_secret", cfg.ClientSecret)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    resp, err := m.http.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token refresh failed: %s", resp.Status)
    }
    
    var tokenResp struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        TokenType    string `json:"token_type"`
        ExpiresIn    int    `json:"expires_in"`
        Scope        string `json:"scope"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, err
    }
    
    newToken := &Token{
        AccessToken:  tokenResp.AccessToken,
        RefreshToken: tokenResp.RefreshToken,
        TokenType:    tokenResp.TokenType,
        Scope:        tokenResp.Scope,
    }
    
    if newToken.RefreshToken == "" {
        newToken.RefreshToken = token.RefreshToken
    }
    
    if tokenResp.ExpiresIn > 0 {
        newToken.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
    }
    
    if err := m.store.Store(ctx, tenantID, provider, newToken); err != nil {
        return nil, err
    }
    
    return newToken, nil
}

func (m *OAuthManager) GetValidToken(ctx context.Context, tenantID, provider string) (*Token, error) {
    token, err := m.store.Get(ctx, tenantID, provider)
    if err != nil {
        return nil, err
    }
    
    if time.Now().Add(5 * time.Minute).Before(token.ExpiresAt) {
        return token, nil
    }
    
    return m.RefreshToken(ctx, tenantID, provider)
}

func (m *OAuthManager) generateState(tenantID, provider, customState, pkce string) string {
    state := StateData{
        TenantID:     tenantID,
        Provider:     provider,
        CustomState:  customState,
        PKCEVerifier: pkce,
        Timestamp:    time.Now().Unix(),
        Nonce:        generateNonce(),
    }
    
    data, _ := json.Marshal(state)
    return base64.URLEncoding.EncodeToString(data)
}

func (m *OAuthManager) verifyState(state string) (*StateData, error) {
    data, err := base64.URLEncoding.DecodeString(state)
    if err != nil {
        return nil, err
    }
    
    var stateData StateData
    if err := json.Unmarshal(data, &stateData); err != nil {
        return nil, err
    }
    
    if time.Now().Unix()-stateData.Timestamp > 600 {
        return nil, ErrStateExpired
    }
    
    return &stateData, nil
}

type StateData struct {
    TenantID     string `json:"tenant_id"`
    Provider     string `json:"provider"`
    CustomState  string `json:"custom_state"`
    PKCEVerifier string `json:"pkce_verifier"`
    Timestamp    int64  `json:"timestamp"`
    Nonce        string `json:"nonce"`
}

func generatePKCEVerifier() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

func computePKCEChallenge(verifier string) string {
    h := sha256.Sum256([]byte(verifier))
    return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func generateNonce() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

### 3.5 API Key Management

```go
// pkg/integration/auth/apikey.go
package auth

import (
    "context"
    "crypto/rand"
    "crypto/subtle"
    "encoding/hex"
    "strings"
    "time"
)

type APIKeyManager struct {
    store APIKeyStore
}

type APIKey struct {
    ID          string
    TenantID    string
    Integration string
    KeyHash     string
    Prefix      string
    Name        string
    Scopes      []string
    ExpiresAt   *time.Time
    LastUsedAt  *time.Time
    CreatedAt   time.Time
    CreatedBy   string
    Metadata    map[string]string
}

type APIKeyStore interface {
    Store(ctx context.Context, key *APIKey) error
    Get(ctx context.Context, keyID string) (*APIKey, error)
    GetByPrefix(ctx context.Context, prefix string) (*APIKey, error)
    Delete(ctx context.Context, keyID string) error
    List(ctx context.Context, tenantID string) ([]*APIKey, error)
    UpdateLastUsed(ctx context.Context, keyID string, t time.Time) error
}

func NewAPIKeyManager(store APIKeyStore) *APIKeyManager {
    return &APIKeyManager{store: store}
}

func (m *APIKeyManager) Create(ctx context.Context, req *CreateAPIKeyRequest) (*APIKeyResponse, error) {
    keyBytes := make([]byte, 32)
    if _, err := rand.Read(keyBytes); err != nil {
        return nil, err
    }
    
    rawKey := hex.EncodeToString(keyBytes)
    prefix := "aibhq_" + req.Integration[:3] + "_"
    fullKey := prefix + rawKey
    
    keyHash := hashKey(fullKey)
    
    key := &APIKey{
        ID:          generateID(),
        TenantID:    req.TenantID,
        Integration: req.Integration,
        KeyHash:     keyHash,
        Prefix:      prefix + rawKey[:8],
        Name:        req.Name,
        Scopes:      req.Scopes,
        CreatedAt:   time.Now(),
        CreatedBy:   req.CreatedBy,
        Metadata:    req.Metadata,
    }
    
    if req.ExpiresIn > 0 {
        exp := time.Now().Add(req.ExpiresIn)
        key.ExpiresAt = &exp
    }
    
    if err := m.store.Store(ctx, key); err != nil {
        return nil, err
    }
    
    return &APIKeyResponse{
        APIKey: key,
        RawKey: fullKey,
    }, nil
}

func (m *APIKeyManager) Validate(ctx context.Context, key string) (*APIKey, error) {
    parts := strings.Split(key, "_")
    if len(parts) < 3 || parts[0] != "aibhq" {
        return nil, ErrInvalidAPIKey
    }
    
    prefix := parts[0] + "_" + parts[1] + "_" + parts[2] + "_"
    storedKey, err := m.store.GetByPrefix(ctx, prefix+parts[3][:8])
    if err != nil {
        return nil, ErrInvalidAPIKey
    }
    
    if storedKey.KeyHash != hashKey(key) {
        return nil, ErrInvalidAPIKey
    }
    
    if storedKey.ExpiresAt != nil && time.Now().After(*storedKey.ExpiresAt) {
        return nil, ErrAPIKeyExpired
    }
    
    now := time.Now()
    m.store.UpdateLastUsed(ctx, storedKey.ID, now)
    
    return storedKey, nil
}

func (m *APIKeyManager) Revoke(ctx context.Context, keyID string) error {
    return m.store.Delete(ctx, keyID)
}

func (m *APIKeyManager) List(ctx context.Context, tenantID string) ([]*APIKey, error) {
    return m.store.List(ctx, tenantID)
}

type CreateAPIKeyRequest struct {
    TenantID    string
    Integration string
    Name        string
    Scopes      []string
    ExpiresIn   time.Duration
    CreatedBy   string
    Metadata    map[string]string
}

type APIKeyResponse struct {
    *APIKey
    RawKey string
}

func hashKey(key string) string {
    h := sha256.Sum256([]byte(key))
    return hex.EncodeToString(h[:])
}
```

### 3.6 Rate Limiting and Quotas

```go
// pkg/integration/ratelimit/limiter.go
package ratelimit

import (
    "context"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    mu          sync.RWMutex
    limiters    map[string]*tenantLimiter
    config      RateLimitConfig
    store       RateLimitStore
}

type RateLimitConfig struct {
    DefaultRPS      int
    DefaultBurst    int
    TenantLimits    map[string]TenantLimit
    IntegrationLimits map[string]IntegrationLimit
}

type TenantLimit struct {
    RPS            int
    Burst          int
    DailyQuota     int
    MonthlyQuota   int
}

type IntegrationLimit struct {
    RPS            int
    Burst          int
    Concurrency    int
}

type tenantLimiter struct {
    limiter    *rate.Limiter
    quota      *QuotaTracker
    concurrent *ConcurrencyTracker
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*tenantLimiter),
        config:   config,
    }
}

func (r *RateLimiter) Wait(ctx context.Context, integration, tenantID string) error {
    key := tenantID + ":" + integration
    
    r.mu.RLock()
    tl, ok := r.limiters[key]
    r.mu.RUnlock()
    
    if !ok {
        tl = r.createLimiter(integration, tenantID)
        r.mu.Lock()
        r.limiters[key] = tl
        r.mu.Unlock()
    }
    
    if err := r.checkQuota(ctx, integration, tenantID, tl.quota); err != nil {
        return err
    }
    
    return tl.limiter.Wait(ctx)
}

func (r *RateLimiter) createLimiter(integration, tenantID string) *tenantLimiter {
    rps := r.config.DefaultRPS
    burst := r.config.DefaultBurst
    
    if tLimit, ok := r.config.TenantLimits[tenantID]; ok {
        rps = tLimit.RPS
        burst = tLimit.Burst
    }
    
    if iLimit, ok := r.config.IntegrationLimits[integration]; ok {
        if iLimit.RPS < rps {
            rps = iLimit.RPS
        }
        if iLimit.Burst < burst {
            burst = iLimit.Burst
        }
    }
    
    return &tenantLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), burst),
        quota:   NewQuotaTracker(tenantID, integration),
    }
}

func (r *RateLimiter) checkQuota(ctx context.Context, integration, tenantID string, tracker *QuotaTracker) error {
    tLimit, hasLimit := r.config.TenantLimits[tenantID]
    if !hasLimit {
        return nil
    }
    
    if tLimit.DailyQuota > 0 {
        if daily, err := tracker.GetDaily(ctx); err == nil && daily >= tLimit.DailyQuota {
            return ErrDailyQuotaExceeded
        }
    }
    
    if tLimit.MonthlyQuota > 0 {
        if monthly, err := tracker.GetMonthly(ctx); err == nil && monthly >= tLimit.MonthlyQuota {
            return ErrMonthlyQuotaExceeded
        }
    }
    
    return nil
}

func (r *RateLimiter) RecordUsage(ctx context.Context, integration, tenantID string, count int) error {
    key := tenantID + ":" + integration
    
    r.mu.RLock()
    tl, ok := r.limiters[key]
    r.mu.RUnlock()
    
    if !ok {
        return nil
    }
    
    return tl.quota.Increment(ctx, count)
}

type QuotaTracker struct {
    tenantID    string
    integration string
    store       QuotaStore
}

func NewQuotaTracker(tenantID, integration string) *QuotaTracker {
    return &QuotaTracker{
        tenantID:    tenantID,
        integration: integration,
    }
}

func (q *QuotaTracker) GetDaily(ctx context.Context) (int, error) {
    key := q.key("daily", time.Now().Format("2006-01-02"))
    return q.store.Get(ctx, key)
}

func (q *QuotaTracker) GetMonthly(ctx context.Context) (int, error) {
    key := q.key("monthly", time.Now().Format("2006-01"))
    return q.store.Get(ctx, key)
}

func (q *QuotaTracker) Increment(ctx context.Context, count int) error {
    dayKey := q.key("daily", time.Now().Format("2006-01-02"))
    monthKey := q.key("monthly", time.Now().Format("2006-01"))
    
    if err := q.store.Increment(ctx, dayKey, count, 25*time.Hour); err != nil {
        return err
    }
    
    return q.store.Increment(ctx, monthKey, count, 32*24*time.Hour)
}

func (q *QuotaTracker) key(quotaType, period string) string {
    return q.tenantID + ":" + q.integration + ":" + quotaType + ":" + period
}

type QuotaStore interface {
    Get(ctx context.Context, key string) (int, error)
    Increment(ctx context.Context, key string, count int, ttl time.Duration) error
}

type ConcurrencyTracker struct {
    mu        sync.Mutex
    active    int
    max       int
    semaphore chan struct{}
}

func NewConcurrencyTracker(max int) *ConcurrencyTracker {
    return &ConcurrencyTracker{
        max:       max,
        semaphore: make(chan struct{}, max),
    }
}

func (c *ConcurrencyTracker) Acquire(ctx context.Context) error {
    select {
    case c.semaphore <- struct{}{}:
        c.mu.Lock()
        c.active++
        c.mu.Unlock()
        return nil
    default:
        return ErrConcurrencyLimitExceeded
    }
}

func (c *ConcurrencyTracker) Release() {
    <-c.semaphore
    c.mu.Lock()
    c.active--
    c.mu.Unlock()
}
```

---

## 4. Integration SDK Design

### 4.1 Core Interfaces

```go
// pkg/integration/adapter/adapter.go
package adapter

import (
    "context"
)

type Adapter interface {
    Name() string
    Category() Category
    Schema() *Schema
    Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
    ValidateConfig(config map[string]any) error
}

type Schema struct {
    Name        string
    Version     string
    Category    Category
    Auth        AuthSchema
    Actions     []ActionSchema
    Webhooks    []WebhookSchema
    Capabilities []string
}

type Category string

const (
    CategoryCRM             Category = "crm"
    CategoryProjectMgmt     Category = "project_management"
    CategoryCommunication   Category = "communication"
    CategoryFinance         Category = "finance"
    CategoryMarketing       Category = "marketing"
    CategoryAnalytics       Category = "analytics"
    CategoryStorage         Category = "storage"
    CategoryCustom          Category = "custom"
)

type AuthSchema struct {
    Type        AuthType
    OAuth2      *OAuth2Schema
    APIKey      *APIKeySchema
    Basic       *BasicAuthSchema
}

type AuthType string

const (
    AuthTypeOAuth2   AuthType = "oauth2"
    AuthTypeAPIKey   AuthType = "api_key"
    AuthTypeBasic    AuthType = "basic"
    AuthTypeCustom   AuthType = "custom"
)

type OAuth2Schema struct {
    GrantType    string
    Scopes       []string
    AuthURL      string
    TokenURL     string
    RefreshURL   string
    PKCE         bool
}

type APIKeySchema struct {
    Header   string
    Prefix   string
    QueryParam string
}

type BasicAuthSchema struct {
    UsernameField string
    PasswordField string
}

type ActionSchema struct {
    Name        string
    Description string
    Category    string
    Input       InputSchema
    Output      OutputSchema
    Required    []string
    Examples    []Example
}

type InputSchema struct {
    Type       string
    Properties map[string]Property
    Required   []string
}

type Property struct {
    Type        string
    Description string
    Enum        []string
    Default     any
    Format      string
}

type OutputSchema struct {
    Type       string
    Properties map[string]Property
}

type Example struct {
    Input  map[string]any
    Output map[string]any
}

type WebhookSchema struct {
    Name        string
    Events      []string
    Transform   string
}

type ExecuteRequest struct {
    Action      string
    Params      map[string]any
    Credentials *Credentials
    Options     ExecuteOptions
}

type Credentials struct {
    OAuth2Token  *OAuth2Token
    APIKey       string
    BasicAuth    *BasicAuth
    Custom       map[string]any
}

type OAuth2Token struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    int64
    TokenType    string
}

type BasicAuth struct {
    Username string
    Password string
}

type ExecuteOptions struct {
    IdempotencyKey string
    Timeout        int
    RetryPolicy    RetryPolicy
}

type RetryPolicy struct {
    MaxRetries  int
    InitialWait int
    MaxWait     int
    Multiplier  float64
}

type ExecuteResponse struct {
    Success     bool
    Data        any
    Error       *ErrorDetail
    Metadata    ResponseMetadata
}

type ErrorDetail struct {
    Code       string
    Message    string
    Retryable  bool
    HTTPStatus int
}

type ResponseMetadata struct {
    RequestID   string
    Duration    int64
    RateLimit   *RateLimitInfo
    Pagination  *Pagination
}

type RateLimitInfo struct {
    Limit     int
    Remaining int
    Reset     int64
}

type Pagination struct {
    Cursor     string
    HasMore    bool
    TotalCount int
}
```

### 4.2 Base Adapter Implementation

```go
// pkg/integration/adapter/base.go
package adapter

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type BaseAdapter struct {
    name       string
    category   Category
    schema     *Schema
    httpClient *http.Client
}

type BaseAdapterOption func(*BaseAdapter)

func WithHTTPClient(client *http.Client) BaseAdapterOption {
    return func(ba *BaseAdapter) {
        ba.httpClient = client
    }
}

func NewBaseAdapter(name string, category Category, schema *Schema, opts ...BaseAdapterOption) *BaseAdapter {
    ba := &BaseAdapter{
        name:     name,
        category: category,
        schema:   schema,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    for _, opt := range opts {
        opt(ba)
    }
    
    return ba
}

func (ba *BaseAdapter) Name() string {
    return ba.name
}

func (ba *BaseAdapter) Category() Category {
    return ba.category
}

func (ba *BaseAdapter) Schema() *Schema {
    return ba.schema
}

func (ba *BaseAdapter) ValidateConfig(config map[string]any) error {
    return nil
}

func (ba *BaseAdapter) DoRequest(ctx context.Context, req *http.Request, creds *Credentials) (*http.Response, error) {
    switch {
    case creds.OAuth2Token != nil:
        req.Header.Set("Authorization", fmt.Sprintf("%s %s", 
            creds.OAuth2Token.TokenType, creds.OAuth2Token.AccessToken))
    case creds.APIKey != "":
        req.Header.Set(ba.schema.Auth.APIKey.Header, ba.schema.Auth.APIKey.Prefix+creds.APIKey)
    case creds.BasicAuth != nil:
        req.SetBasicAuth(creds.BasicAuth.Username, creds.BasicAuth.Password)
    }
    
    req = req.WithContext(ctx)
    
    return ba.httpClient.Do(req)
}

func (ba *BaseAdapter) ParseResponse(resp *http.Response, v any) error {
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        var errResp struct {
            Error   string `json:"error"`
            Message string `json:"message"`
        }
        json.NewDecoder(resp.Body).Decode(&errResp)
        return fmt.Errorf("API error: %s - %s", errResp.Error, errResp.Message)
    }
    
    return json.NewDecoder(resp.Body).Decode(v)
}
```

### 4.3 Salesforce Adapter Example

```go
// pkg/integration/adapters/salesforce/adapter.go
package salesforce

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
)

const (
    productionAuthURL = "https://login.salesforce.com/services/oauth2/authorize"
    productionTokenURL = "https://login.salesforce.com/services/oauth2/token"
    sandboxAuthURL    = "https://test.salesforce.com/services/oauth2/authorize"
    sandboxTokenURL   = "https://test.salesforce.com/services/oauth2/token"
)

type SalesforceAdapter struct {
    *adapter.BaseAdapter
    config Config
}

type Config struct {
    ClientID     string
    ClientSecret string
    Sandbox      bool
    APIVersion   string
}

func NewAdapter(cfg Config) *SalesforceAdapter {
    schema := &adapter.Schema{
        Name:     "salesforce",
        Version:  "1.0.0",
        Category: adapter.CategoryCRM,
        Auth: adapter.AuthSchema{
            Type: adapter.AuthTypeOAuth2,
            OAuth2: &adapter.OAuth2Schema{
                GrantType: "authorization_code",
                Scopes:    []string{"api", "refresh_token", "offline_access"},
                AuthURL:   productionAuthURL,
                TokenURL:  productionTokenURL,
                PKCE:      true,
            },
        },
        Actions: []adapter.ActionSchema{
            {
                Name:        "get_contact",
                Description: "Retrieve a contact by ID",
                Category:    "read",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "contact_id": {
                            Type:        "string",
                            Description: "Salesforce contact ID",
                        },
                    },
                    Required: []string{"contact_id"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":        {Type: "string"},
                        "firstName": {Type: "string"},
                        "lastName":  {Type: "string"},
                        "email":     {Type: "string"},
                        "phone":     {Type: "string"},
                    },
                },
            },
            {
                Name:        "create_contact",
                Description: "Create a new contact",
                Category:    "write",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "firstName": {Type: "string"},
                        "lastName":  {Type: "string"},
                        "email":     {Type: "string", Format: "email"},
                        "phone":     {Type: "string"},
                        "accountId": {Type: "string"},
                    },
                    Required: []string{"lastName"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":      {Type: "string"},
                        "success": {Type: "boolean"},
                    },
                },
            },
            {
                Name:        "query",
                Description: "Execute a SOQL query",
                Category:    "read",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "soql": {
                            Type:        "string",
                            Description: "SOQL query string",
                        },
                    },
                    Required: []string{"soql"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "totalSize": {Type: "integer"},
                        "done":      {Type: "boolean"},
                        "records":   {Type: "array"},
                    },
                },
            },
            {
                Name:        "create_lead",
                Description: "Create a new lead",
                Category:    "write",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "firstName":   {Type: "string"},
                        "lastName":    {Type: "string"},
                        "company":     {Type: "string"},
                        "email":       {Type: "string", Format: "email"},
                        "phone":       {Type: "string"},
                        "status": {
                            Type: "string",
                            Enum: []string{"Open", "Working", "Closed - Converted", "Closed - Not Converted"},
                        },
                    },
                    Required: []string{"lastName", "company"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":      {Type: "string"},
                        "success": {Type: "boolean"},
                    },
                },
            },
            {
                Name:        "get_opportunity",
                Description: "Retrieve an opportunity by ID",
                Category:    "read",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "opportunity_id": {Type: "string"},
                    },
                    Required: []string{"opportunity_id"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":             {Type: "string"},
                        "name":           {Type: "string"},
                        "amount":         {Type: "number"},
                        "stageName":      {Type: "string"},
                        "closeDate":      {Type: "string"},
                        "probability":    {Type: "number"},
                    },
                },
            },
        },
        Webhooks: []adapter.WebhookSchema{
            {
                Name:   "salesforce_events",
                Events: []string{"create", "update", "delete", "undelete"},
            },
        },
        Capabilities: []string{
            "contacts.read", "contacts.write",
            "leads.read", "leads.write",
            "opportunities.read", "opportunities.write",
            "accounts.read", "accounts.write",
            "cases.read", "cases.write",
        },
    }
    
    return &SalesforceAdapter{
        BaseAdapter: adapter.NewBaseAdapter("salesforce", adapter.CategoryCRM, schema),
        config:      cfg,
    }
}

func (a *SalesforceAdapter) Execute(ctx context.Context, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    instanceURL, ok := req.Credentials.Custom["instance_url"].(string)
    if !ok {
        return nil, fmt.Errorf("missing instance_url in credentials")
    }
    
    baseURL := fmt.Sprintf("%s/services/data/v%s", instanceURL, a.config.APIVersion)
    
    switch req.Action {
    case "get_contact":
        return a.getContact(ctx, baseURL, req)
    case "create_contact":
        return a.createContact(ctx, baseURL, req)
    case "query":
        return a.query(ctx, baseURL, req)
    case "create_lead":
        return a.createLead(ctx, baseURL, req)
    case "get_opportunity":
        return a.getOpportunity(ctx, baseURL, req)
    default:
        return nil, fmt.Errorf("unknown action: %s", req.Action)
    }
}

func (a *SalesforceAdapter) getContact(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    contactID, ok := req.Params["contact_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing contact_id")
    }
    
    endpoint := fmt.Sprintf("%s/sobjects/Contact/%s", baseURL, contactID)
    
    httpReq, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var contact Contact
    if err := a.ParseResponse(resp, &contact); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    contact,
    }, nil
}

func (a *SalesforceAdapter) createContact(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/sobjects/Contact", baseURL)
    
    body, err := json.Marshal(req.Params)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var result struct {
        ID      string `json:"id"`
        Success bool   `json:"success"`
        Errors  []struct {
            StatusCode string `json:"statusCode"`
            Message    string `json:"message"`
        } `json:"errors"`
    }
    
    if err := a.ParseResponse(resp, &result); err != nil {
        return nil, err
    }
    
    if !result.Success && len(result.Errors) > 0 {
        return &adapter.ExecuteResponse{
            Success: false,
            Error: &adapter.ErrorDetail{
                Code:    result.Errors[0].StatusCode,
                Message: result.Errors[0].Message,
            },
        }, nil
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    result,
    }, nil
}

func (a *SalesforceAdapter) query(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    soql, ok := req.Params["soql"].(string)
    if !ok {
        return nil, fmt.Errorf("missing soql query")
    }
    
    endpoint := fmt.Sprintf("%s/query?q=%s", baseURL, url.QueryEscape(soql))
    
    httpReq, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var result QueryResult
    if err := a.ParseResponse(resp, &result); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    result,
        Metadata: adapter.ResponseMetadata{
            Pagination: &adapter.Pagination{
                HasMore: !result.Done,
                Cursor:  result.NextRecordsURL,
            },
        },
    }, nil
}

func (a *SalesforceAdapter) createLead(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/sobjects/Lead", baseURL)
    
    body, err := json.Marshal(req.Params)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var result struct {
        ID      string `json:"id"`
        Success bool   `json:"success"`
        Errors  []struct {
            StatusCode string `json:"statusCode"`
            Message    string `json:"message"`
        } `json:"errors"`
    }
    
    if err := a.ParseResponse(resp, &result); err != nil {
        return nil, err
    }
    
    if !result.Success && len(result.Errors) > 0 {
        return &adapter.ExecuteResponse{
            Success: false,
            Error: &adapter.ErrorDetail{
                Code:    result.Errors[0].StatusCode,
                Message: result.Errors[0].Message,
            },
        }, nil
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    result,
    }, nil
}

func (a *SalesforceAdapter) getOpportunity(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    oppID, ok := req.Params["opportunity_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing opportunity_id")
    }
    
    endpoint := fmt.Sprintf("%s/sobjects/Opportunity/%s", baseURL, oppID)
    
    httpReq, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var opp Opportunity
    if err := a.ParseResponse(resp, &opp); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    opp,
    }, nil
}

type Contact struct {
    ID        string `json:"Id"`
    FirstName string `json:"FirstName"`
    LastName  string `json:"LastName"`
    Email     string `json:"Email"`
    Phone     string `json:"Phone"`
    AccountID string `json:"AccountId"`
}

type Lead struct {
    ID        string `json:"Id"`
    FirstName string `json:"FirstName"`
    LastName  string `json:"LastName"`
    Company   string `json:"Company"`
    Email     string `json:"Email"`
    Phone     string `json:"Phone"`
    Status    string `json:"Status"`
}

type Opportunity struct {
    ID           string  `json:"Id"`
    Name         string  `json:"Name"`
    Amount       float64 `json:"Amount"`
    StageName    string  `json:"StageName"`
    CloseDate    string  `json:"CloseDate"`
    Probability  float64 `json:"Probability"`
    AccountID    string  `json:"AccountId"`
}

type QueryResult struct {
    TotalSize      int                    `json:"totalSize"`
    Done           bool                   `json:"done"`
    Records        []map[string]any       `json:"records"`
    NextRecordsURL string                 `json:"nextRecordsUrl"`
}
```

### 4.4 Stripe Adapter Example

```go
// pkg/integration/adapters/stripe/adapter.go
package stripe

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
)

type StripeAdapter struct {
    *adapter.BaseAdapter
    config Config
}

type Config struct {
    APIVersion string
}

func NewAdapter(cfg Config) *StripeAdapter {
    schema := &adapter.Schema{
        Name:     "stripe",
        Version:  "1.0.0",
        Category: adapter.CategoryFinance,
        Auth: adapter.AuthSchema{
            Type: adapter.AuthTypeAPIKey,
            APIKey: &adapter.APIKeySchema{
                Header: "Authorization",
                Prefix: "Bearer ",
            },
        },
        Actions: []adapter.ActionSchema{
            {
                Name:        "create_customer",
                Description: "Create a new Stripe customer",
                Category:    "write",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "email":       {Type: "string", Format: "email"},
                        "name":        {Type: "string"},
                        "phone":       {Type: "string"},
                        "description": {Type: "string"},
                        "metadata":    {Type: "object"},
                    },
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":      {Type: "string"},
                        "object":  {Type: "string"},
                        "email":   {Type: "string"},
                        "name":    {Type: "string"},
                    },
                },
            },
            {
                Name:        "get_customer",
                Description: "Retrieve a customer by ID",
                Category:    "read",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "customer_id": {Type: "string"},
                    },
                    Required: []string{"customer_id"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":            {Type: "string"},
                        "email":         {Type: "string"},
                        "name":          {Type: "string"},
                        "balance":       {Type: "integer"},
                        "created":       {Type: "integer"},
                    },
                },
            },
            {
                Name:        "create_payment_intent",
                Description: "Create a payment intent",
                Category:    "write",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "amount":          {Type: "integer", Description: "Amount in cents"},
                        "currency":        {Type: "string", Description: "Three-letter ISO currency code"},
                        "customer":        {Type: "string"},
                        "payment_method":  {Type: "string"},
                        "description":     {Type: "string"},
                        "metadata":        {Type: "object"},
                        "confirm":         {Type: "boolean"},
                    },
                    Required: []string{"amount", "currency"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":              {Type: "string"},
                        "amount":          {Type: "integer"},
                        "currency":        {Type: "string"},
                        "status":          {Type: "string"},
                        "client_secret":   {Type: "string"},
                    },
                },
            },
            {
                Name:        "list_invoices",
                Description: "List invoices",
                Category:    "read",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "customer":    {Type: "string"},
                        "status":      {Type: "string", Enum: []string{"draft", "open", "paid", "uncollectible", "void"}},
                        "limit":       {Type: "integer"},
                        "starting_after": {Type: "string"},
                    },
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "object":      {Type: "string"},
                        "has_more":    {Type: "boolean"},
                        "data":        {Type: "array"},
                    },
                },
            },
            {
                Name:        "create_subscription",
                Description: "Create a subscription",
                Category:    "write",
                Input: adapter.InputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "customer": {Type: "string"},
                        "items": {
                            Type: "array",
                            Description: "Array of subscription items",
                        },
                        "trial_period_days": {Type: "integer"},
                        "metadata": {Type: "object"},
                    },
                    Required: []string{"customer", "items"},
                },
                Output: adapter.OutputSchema{
                    Type: "object",
                    Properties: map[string]adapter.Property{
                        "id":      {Type: "string"},
                        "status":  {Type: "string"},
                        "customer": {Type: "string"},
                    },
                },
            },
        },
        Webhooks: []adapter.WebhookSchema{
            {
                Name: "stripe_events",
                Events: []string{
                    "customer.created", "customer.updated", "customer.deleted",
                    "invoice.paid", "invoice.payment_failed",
                    "payment_intent.succeeded", "payment_intent.payment_failed",
                    "charge.succeeded", "charge.failed",
                    "subscription.created", "subscription.updated", "subscription.deleted",
                },
            },
        },
        Capabilities: []string{
            "customers.read", "customers.write",
            "charges.read", "charges.write",
            "invoices.read", "invoices.write",
            "subscriptions.read", "subscriptions.write",
            "payment_intents.read", "payment_intents.write",
        },
    }
    
    return &StripeAdapter{
        BaseAdapter: adapter.NewBaseAdapter("stripe", adapter.CategoryFinance, schema),
        config:      cfg,
    }
}

func (a *StripeAdapter) Execute(ctx context.Context, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    baseURL := "https://api.stripe.com/v1"
    
    switch req.Action {
    case "create_customer":
        return a.createCustomer(ctx, baseURL, req)
    case "get_customer":
        return a.getCustomer(ctx, baseURL, req)
    case "create_payment_intent":
        return a.createPaymentIntent(ctx, baseURL, req)
    case "list_invoices":
        return a.listInvoices(ctx, baseURL, req)
    case "create_subscription":
        return a.createSubscription(ctx, baseURL, req)
    default:
        return nil, fmt.Errorf("unknown action: %s", req.Action)
    }
}

func (a *StripeAdapter) createCustomer(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/customers", baseURL)
    
    formData := a.buildFormData(req.Params)
    
    httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(formData)))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    httpReq.Header.Set("Stripe-Version", a.config.APIVersion)
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var customer Customer
    if err := a.ParseResponse(resp, &customer); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    customer,
    }, nil
}

func (a *StripeAdapter) getCustomer(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    customerID, ok := req.Params["customer_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing customer_id")
    }
    
    endpoint := fmt.Sprintf("%s/customers/%s", baseURL, customerID)
    
    httpReq, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Stripe-Version", a.config.APIVersion)
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var customer Customer
    if err := a.ParseResponse(resp, &customer); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    customer,
    }, nil
}

func (a *StripeAdapter) createPaymentIntent(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/payment_intents", baseURL)
    
    formData := a.buildFormData(req.Params)
    
    httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(formData)))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    httpReq.Header.Set("Stripe-Version", a.config.APIVersion)
    
    if req.Options.IdempotencyKey != "" {
        httpReq.Header.Set("Idempotency-Key", req.Options.IdempotencyKey)
    }
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var pi PaymentIntent
    if err := a.ParseResponse(resp, &pi); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    pi,
    }, nil
}

func (a *StripeAdapter) listInvoices(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/invoices", baseURL)
    
    params := make([]string, 0)
    if customer, ok := req.Params["customer"].(string); ok {
        params = append(params, fmt.Sprintf("customer=%s", customer))
    }
    if status, ok := req.Params["status"].(string); ok {
        params = append(params, fmt.Sprintf("status=%s", status))
    }
    if limit, ok := req.Params["limit"].(int); ok {
        params = append(params, fmt.Sprintf("limit=%d", limit))
    }
    if startingAfter, ok := req.Params["starting_after"].(string); ok {
        params = append(params, fmt.Sprintf("starting_after=%s", startingAfter))
    }
    
    if len(params) > 0 {
        endpoint = endpoint + "?" + strings.Join(params, "&")
    }
    
    httpReq, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Stripe-Version", a.config.APIVersion)
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var result ListResponse
    if err := a.ParseResponse(resp, &result); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    result.Data,
        Metadata: adapter.ResponseMetadata{
            Pagination: &adapter.Pagination{
                HasMore: result.HasMore,
            },
        },
    }, nil
}

func (a *StripeAdapter) createSubscription(ctx context.Context, baseURL string, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    endpoint := fmt.Sprintf("%s/subscriptions", baseURL)
    
    formData := a.buildFormData(req.Params)
    
    httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(formData)))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    httpReq.Header.Set("Stripe-Version", a.config.APIVersion)
    
    resp, err := a.DoRequest(ctx, httpReq, req.Credentials)
    if err != nil {
        return nil, err
    }
    
    var sub Subscription
    if err := a.ParseResponse(resp, &sub); err != nil {
        return nil, err
    }
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    sub,
    }, nil
}

func (a *StripeAdapter) buildFormData(params map[string]any) string {
    values := make(url.Values)
    
    for key, value := range params {
        switch v := value.(type) {
        case string:
            values.Set(key, v)
        case int:
            values.Set(key, strconv.Itoa(v))
        case bool:
            values.Set(key, strconv.FormatBool(v))
        case map[string]any:
            for subKey, subVal := range v {
                values.Set(fmt.Sprintf("%s[%s]", key, subKey), fmt.Sprintf("%v", subVal))
            }
        case []any:
            for i, item := range v {
                if m, ok := item.(map[string]any); ok {
                    for subKey, subVal := range m {
                        values.Set(fmt.Sprintf("%s[%d][%s]", key, i, subKey), fmt.Sprintf("%v", subVal))
                    }
                }
            }
        }
    }
    
    return values.Encode()
}

type Customer struct {
    ID          string            `json:"id"`
    Object      string            `json:"object"`
    Email       string            `json:"email"`
    Name        string            `json:"name"`
    Phone       string            `json:"phone"`
    Balance     int64             `json:"balance"`
    Created     int64             `json:"created"`
    Metadata    map[string]string `json:"metadata"`
}

type PaymentIntent struct {
    ID            string `json:"id"`
    Object        string `json:"object"`
    Amount        int64  `json:"amount"`
    Currency      string `json:"currency"`
    Status        string `json:"status"`
    ClientSecret  string `json:"client_secret"`
    Customer      string `json:"customer"`
    Created       int64  `json:"created"`
}

type Subscription struct {
    ID       string `json:"id"`
    Object   string `json:"object"`
    Status   string `json:"status"`
    Customer string `json:"customer"`
    Created  int64  `json:"created"`
}

type ListResponse struct {
    Object   string `json:"object"`
    HasMore  bool   `json:"has_more"`
    Data     []any  `json:"data"`
}
```

### 4.5 Configuration Schema

```go
// pkg/integration/config/schema.go
package config

import (
    "encoding/json"
    "fmt"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
)

type IntegrationConfig struct {
    Name        string         `json:"name"`
    Enabled     bool           `json:"enabled"`
    Auth        AuthConfig     `json:"auth"`
    Webhooks    WebhookConfig  `json:"webhooks,omitempty"`
    RateLimit   RateLimitCfg   `json:"rateLimit,omitempty"`
    Options     map[string]any `json:"options,omitempty"`
}

type AuthConfig struct {
    Type         string         `json:"type"`
    OAuth2       *OAuth2Config  `json:"oauth2,omitempty"`
    APIKey       *APIKeyConfig  `json:"apiKey,omitempty"`
    Basic        *BasicConfig   `json:"basic,omitempty"`
}

type OAuth2Config struct {
    ClientID     string   `json:"clientId"`
    ClientSecret string   `json:"clientSecret"`
    Scopes       []string `json:"scopes"`
    Sandbox      bool     `json:"sandbox,omitempty"`
    ExtraParams  map[string]string `json:"extraParams,omitempty"`
}

type APIKeyConfig struct {
    Key        string `json:"key"`
    Header     string `json:"header,omitempty"`
    Prefix     string `json:"prefix,omitempty"`
    TestMode   bool   `json:"testMode,omitempty"`
}

type BasicConfig struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type WebhookConfig struct {
    Enabled    bool     `json:"enabled"`
    Secret     string   `json:"secret"`
    Events     []string `json:"events"`
    Endpoint   string   `json:"endpoint"`
}

type RateLimitCfg struct {
    RPS           int `json:"rps"`
    Burst         int `json:"burst"`
    DailyQuota    int `json:"dailyQuota,omitempty"`
    MonthlyQuota  int `json:"monthlyQuota,omitempty"`
}

func ParseConfig(data []byte) (*IntegrationConfig, error) {
    var cfg IntegrationConfig
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing integration config: %w", err)
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

func (c *IntegrationConfig) Validate() error {
    if c.Name == "" {
        return fmt.Errorf("integration name is required")
    }
    
    switch c.Auth.Type {
    case "oauth2":
        if c.Auth.OAuth2 == nil {
            return fmt.Errorf("oauth2 config is required for oauth2 auth type")
        }
        if c.Auth.OAuth2.ClientID == "" {
            return fmt.Errorf("oauth2 client ID is required")
        }
    case "api_key":
        if c.Auth.APIKey == nil || c.Auth.APIKey.Key == "" {
            return fmt.Errorf("API key is required for api_key auth type")
        }
    case "basic":
        if c.Auth.Basic == nil {
            return fmt.Errorf("basic auth config is required for basic auth type")
        }
    default:
        return fmt.Errorf("unsupported auth type: %s", c.Auth.Type)
    }
    
    return nil
}

func (c *IntegrationConfig) ToMap() map[string]any {
    data, _ := json.Marshal(c)
    var result map[string]any
    json.Unmarshal(data, &result)
    return result
}
```

### 4.6 Error Handling Standards

```go
// pkg/integration/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

type IntegrationError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    Provider   string `json:"provider,omitempty"`
    HTTPStatus int    `json:"-"`
    Retryable  bool   `json:"retryable"`
    Details    any    `json:"details,omitempty"`
}

func (e *IntegrationError) Error() string {
    if e.Provider != "" {
        return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *IntegrationError) IsRetryable() bool {
    return e.Retryable
}

var (
    ErrIntegrationNotFound = &IntegrationError{
        Code:       "INTEGRATION_NOT_FOUND",
        Message:    "The requested integration was not found",
        HTTPStatus: http.StatusNotFound,
        Retryable:  false,
    }
    
    ErrAdapterAlreadyRegistered = &IntegrationError{
        Code:       "ADAPTER_ALREADY_REGISTERED",
        Message:    "An adapter with this name is already registered",
        HTTPStatus: http.StatusConflict,
        Retryable:  false,
    }
    
    ErrProviderNotConfigured = &IntegrationError{
        Code:       "PROVIDER_NOT_CONFIGURED",
        Message:    "The OAuth provider is not configured",
        HTTPStatus: http.StatusInternalServerError,
        Retryable:  false,
    }
    
    ErrStateExpired = &IntegrationError{
        Code:       "STATE_EXPIRED",
        Message:    "OAuth state has expired",
        HTTPStatus: http.StatusBadRequest,
        Retryable:  false,
    }
    
    ErrNoRefreshToken = &IntegrationError{
        Code:       "NO_REFRESH_TOKEN",
        Message:    "No refresh token available",
        HTTPStatus: http.StatusBadRequest,
        Retryable:  false,
    }
    
    ErrInvalidAPIKey = &IntegrationError{
        Code:       "INVALID_API_KEY",
        Message:    "The provided API key is invalid",
        HTTPStatus: http.StatusUnauthorized,
        Retryable:  false,
    }
    
    ErrAPIKeyExpired = &IntegrationError{
        Code:       "API_KEY_EXPIRED",
        Message:    "The API key has expired",
        HTTPStatus: http.StatusUnauthorized,
        Retryable:  false,
    }
    
    ErrDailyQuotaExceeded = &IntegrationError{
        Code:       "DAILY_QUOTA_EXCEEDED",
        Message:    "Daily API quota has been exceeded",
        HTTPStatus: http.StatusTooManyRequests,
        Retryable:  true,
    }
    
    ErrMonthlyQuotaExceeded = &IntegrationError{
        Code:       "MONTHLY_QUOTA_EXCEEDED",
        Message:    "Monthly API quota has been exceeded",
        HTTPStatus: http.StatusTooManyRequests,
        Retryable:  true,
    }
    
    ErrConcurrencyLimitExceeded = &IntegrationError{
        Code:       "CONCURRENCY_LIMIT_EXCEEDED",
        Message:    "Maximum concurrent requests exceeded",
        HTTPStatus: http.StatusTooManyRequests,
        Retryable:  true,
    }
)

func NewProviderError(provider, code, message string, httpStatus int, retryable bool) *IntegrationError {
    return &IntegrationError{
        Code:       code,
        Message:    message,
        Provider:   provider,
        HTTPStatus: httpStatus,
        Retryable:  retryable,
    }
}

func IsRetryable(err error) bool {
    if ie, ok := err.(*IntegrationError); ok {
        return ie.Retryable
    }
    return false
}

func GetHTTPStatus(err error) int {
    if ie, ok := err.(*IntegrationError); ok {
        return ie.HTTPStatus
    }
    return http.StatusInternalServerError
}
```

### 4.7 Testing Approach

```go
// pkg/integration/adapter/adapter_test.go
package adapter_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/raynaythegreat/ai-business-hq/pkg/integration/adapter"
)

type MockAdapter struct {
    *adapter.BaseAdapter
    server *httptest.Server
}

func NewMockAdapter(server *httptest.Server) *MockAdapter {
    return &MockAdapter{
        BaseAdapter: adapter.NewBaseAdapter("mock", adapter.CategoryCRM, &adapter.Schema{
            Name:     "mock",
            Version:  "1.0.0",
            Category: adapter.CategoryCRM,
            Actions: []adapter.ActionSchema{
                {
                    Name: "test_action",
                    Input: adapter.InputSchema{
                        Type: "object",
                        Properties: map[string]adapter.Property{
                            "param1": {Type: "string"},
                        },
                    },
                },
            },
        }),
        server: server,
    }
}

func (m *MockAdapter) Execute(ctx context.Context, req *adapter.ExecuteRequest) (*adapter.ExecuteResponse, error) {
    httpReq, _ := http.NewRequest("GET", m.server.URL+"/test", nil)
    resp, err := m.DoRequest(ctx, httpReq, &adapter.Credentials{
        APIKey: "test-key",
    })
    if err != nil {
        return nil, err
    }
    
    var result map[string]any
    json.NewDecoder(resp.Body).Decode(&result)
    resp.Body.Close()
    
    return &adapter.ExecuteResponse{
        Success: true,
        Data:    result,
    }, nil
}

func TestAdapterExecute(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()
    
    a := NewMockAdapter(server)
    
    resp, err := a.Execute(context.Background(), &adapter.ExecuteRequest{
        Action: "test_action",
        Params: map[string]any{"param1": "value1"},
    })
    
    require.NoError(t, err)
    assert.True(t, resp.Success)
}

func TestAdapterRateLimit(t *testing.T) {
    calls := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        calls++
        if calls > 5 {
            w.WriteHeader(http.StatusTooManyRequests)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()
    
    a := NewMockAdapter(server)
    
    for i := 0; i < 6; i++ {
        _, err := a.Execute(context.Background(), &adapter.ExecuteRequest{
            Action: "test_action",
        })
        
        if i < 5 {
            assert.NoError(t, err)
        } else {
            assert.Error(t, err)
        }
    }
}

func TestAdapterTimeout(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        <-r.Context().Done()
    }))
    defer server.Close()
    
    a := NewMockAdapter(server)
    
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    _, err := a.Execute(ctx, &adapter.ExecuteRequest{
        Action: "test_action",
    })
    
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestAdapterRetry(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()
    
    a := NewMockAdapter(server)
    
    resp, err := a.Execute(context.Background(), &adapter.ExecuteRequest{
        Action: "test_action",
        Options: adapter.ExecuteOptions{
            RetryPolicy: adapter.RetryPolicy{
                MaxRetries:  3,
                InitialWait: 10,
                MaxWait:     100,
                Multiplier:  2,
            },
        },
    })
    
    require.NoError(t, err)
    assert.True(t, resp.Success)
    assert.Equal(t, 3, attempts)
}
```

---

## 5. API Marketplace

### 5.1 OpenAPI Specification

```yaml
openapi: 3.1.0
info:
  title: AI Business HQ API
  version: 1.0.0
  description: |
    The AI Business HQ API provides programmatic access to integrations,
    webhooks, and AI-powered business operations.
    
    ## Authentication
    All API requests require authentication using Bearer tokens or API keys.
    
    ## Rate Limits
    - Free tier: 100 requests/minute
    - Pro tier: 1000 requests/minute
    - Enterprise: Custom limits
    
  contact:
    name: AI Business HQ Support
    email: api@aibhq.io
    url: https://aibhq.io/docs/api
    
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: https://api.aibhq.io/v1
    description: Production
  - url: https://api-staging.aibhq.io/v1
    description: Staging

security:
  - BearerAuth: []
  - ApiKeyAuth: []

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key

  schemas:
    Integration:
      type: object
      required:
        - name
        - category
        - auth_type
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        display_name:
          type: string
        category:
          type: string
          enum: [crm, project_management, communication, finance, marketing]
        auth_type:
          type: string
          enum: [oauth2, api_key, basic]
        capabilities:
          type: array
          items:
            type: string
        status:
          type: string
          enum: [connected, disconnected, error]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    OAuthConfig:
      type: object
      required:
        - client_id
        - client_secret
      properties:
        client_id:
          type: string
        client_secret:
          type: string
        scopes:
          type: array
          items:
            type: string
        sandbox:
          type: boolean
          default: false

    Webhook:
      type: object
      required:
        - url
        - events
      properties:
        id:
          type: string
          format: uuid
        url:
          type: string
          format: uri
        events:
          type: array
          items:
            type: string
        secret:
          type: string
        active:
          type: boolean
        created_at:
          type: string
          format: date-time

    ExecuteRequest:
      type: object
      required:
        - integration
        - action
      properties:
        integration:
          type: string
        action:
          type: string
        params:
          type: object
          additionalProperties: true
        options:
          $ref: '#/components/schemas/ExecuteOptions'

    ExecuteOptions:
      type: object
      properties:
        idempotency_key:
          type: string
        timeout_ms:
          type: integer
          default: 30000
        retry_count:
          type: integer
          default: 3

    ExecuteResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          additionalProperties: true
        error:
          $ref: '#/components/schemas/Error'
        metadata:
          $ref: '#/components/schemas/ResponseMetadata'

    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
        retryable:
          type: boolean
        details:
          type: object
          additionalProperties: true

    ResponseMetadata:
      type: object
      properties:
        request_id:
          type: string
        duration_ms:
          type: integer
        rate_limit:
          $ref: '#/components/schemas/RateLimitInfo'
        pagination:
          $ref: '#/components/schemas/Pagination'

    RateLimitInfo:
      type: object
      properties:
        limit:
          type: integer
        remaining:
          type: integer
        reset:
          type: integer

    Pagination:
      type: object
      properties:
        cursor:
          type: string
        has_more:
          type: boolean
        total_count:
          type: integer

paths:
  /integrations:
    get:
      operationId: listIntegrations
      summary: List available integrations
      tags: [Integrations]
      parameters:
        - name: category
          in: query
          schema:
            type: string
        - name: status
          in: query
          schema:
            type: string
      responses:
        '200':
          description: List of integrations
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Integration'

  /integrations/{name}/connect:
    post:
      operationId: connectIntegration
      summary: Connect an integration
      tags: [Integrations]
      parameters:
        - name: name
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              oneOf:
                - $ref: '#/components/schemas/OAuthConfig'
      responses:
        '200':
          description: Connection initiated
          content:
            application/json:
              schema:
                type: object
                properties:
                  auth_url:
                    type: string
                    format: uri
                  state:
                    type: string

  /integrations/{name}/disconnect:
    post:
      operationId: disconnectIntegration
      summary: Disconnect an integration
      tags: [Integrations]
      parameters:
        - name: name
          in: path
          required: true
          schema:
            type: string
      responses:
        '204':
          description: Disconnected

  /execute:
    post:
      operationId: executeAction
      summary: Execute an integration action
      tags: [Actions]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ExecuteRequest'
      responses:
        '200':
          description: Execution result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ExecuteResponse'

  /webhooks:
    get:
      operationId: listWebhooks
      summary: List webhooks
      tags: [Webhooks]
      responses:
        '200':
          description: List of webhooks
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Webhook'

    post:
      operationId: createWebhook
      summary: Create a webhook
      tags: [Webhooks]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Webhook'
      responses:
        '201':
          description: Webhook created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Webhook'

  /webhooks/{id}:
    get:
      operationId: getWebhook
      summary: Get a webhook
      tags: [Webhooks]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Webhook details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Webhook'

    delete:
      operationId: deleteWebhook
      summary: Delete a webhook
      tags: [Webhooks]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '204':
          description: Webhook deleted

  /api-keys:
    get:
      operationId: listAPIKeys
      summary: List API keys
      tags: [API Keys]
      responses:
        '200':
          description: List of API keys

    post:
      operationId: createAPIKey
      summary: Create an API key
      tags: [API Keys]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - integration
              properties:
                name:
                  type: string
                integration:
                  type: string
                scopes:
                  type: array
                  items:
                    type: string
                expires_in_days:
                  type: integer
      responses:
        '201':
          description: API key created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  key:
                    type: string

  /api-keys/{id}:
    delete:
      operationId: revokeAPIKey
      summary: Revoke an API key
      tags: [API Keys]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '204':
          description: API key revoked

tags:
  - name: Integrations
    description: Integration management
  - name: Actions
    description: Execute integration actions
  - name: Webhooks
    description: Webhook management
  - name: API Keys
    description: API key management
```

### 5.2 Developer Portal Requirements

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Developer Portal Structure                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. Getting Started                                                         │
│     ├── Quick Start Guide                                                   │
│     ├── Authentication                                                      │
│     ├── Your First Integration                                              │
│     └── SDK Installation                                                    │
│                                                                              │
│  2. API Reference                                                           │
│     ├── OpenAPI Spec (interactive)                                          │
│     ├── Authentication & Security                                           │
│     ├── Rate Limits & Quotas                                                │
│     ├── Error Handling                                                      │
│     └── Pagination                                                          │
│                                                                              │
│  3. Integrations                                                            │
│     ├── CRM (Salesforce, HubSpot)                                           │
│     ├── Project Management (Linear, Jira, Asana)                           │
│     ├── Communication (Slack, Teams)                                        │
│     ├── Finance (Stripe, QuickBooks)                                        │
│     └── Building Custom Integrations                                        │
│                                                                              │
│  4. Webhooks                                                                │
│     ├── Overview                                                            │
│     ├── Event Types                                                         │
│     ├── Signature Verification                                              │
│     ├── Retry Policy                                                        │
│     └── Testing Webhooks                                                    │
│                                                                              │
│  5. SDKs & Tools                                                            │
│     ├── Go SDK                                                              │
│     ├── TypeScript SDK                                                      │
│     ├── Python SDK                                                          │
│     └── CLI Tool                                                            │
│                                                                              │
│  6. Guides                                                                  │
│     ├── OAuth Implementation                                                │
│     ├── Error Handling Best Practices                                       │
│     ├── Rate Limit Handling                                                 │
│     └── Idempotency Patterns                                                │
│                                                                              │
│  7. Changelog                                                               │
│     └── API Version History                                                 │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.3 API Versioning Strategy

```go
// pkg/api/versioning/middleware.go
package versioning

import (
    "net/http"
    "strings"
)

type Version struct {
    Major int
    Minor int
}

func ParseVersion(s string) (Version, bool) {
    s = strings.TrimPrefix(s, "v")
    parts := strings.Split(s, ".")
    
    var v Version
    if len(parts) >= 1 {
        if _, err := fmt.Sscanf(parts[0], "%d", &v.Major); err != nil {
            return Version{}, false
        }
    }
    if len(parts) >= 2 {
        fmt.Sscanf(parts[1], "%d", &v.Minor)
    }
    
    return v, true
}

func (v Version) String() string {
    return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

type VersioningMiddleware struct {
    defaultVersion Version
    supportedVersions []Version
    deprecatedVersions []Version
}

func NewVersioningMiddleware(defaultVersion Version) *VersioningMiddleware {
    return &VersioningMiddleware{
        defaultVersion: defaultVersion,
        supportedVersions: []Version{
            {Major: 1, Minor: 0},
        },
        deprecatedVersions: []Version{},
    }
}

func (vm *VersioningMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        version := vm.extractVersion(r)
        
        if !vm.isSupported(version) {
            http.Error(w, "Unsupported API version", http.StatusBadRequest)
            return
        }
        
        w.Header().Set("X-API-Version", version.String())
        
        if vm.isDeprecated(version) {
            w.Header().Set("X-API-Deprecated", "true")
            w.Header().Set("X-API-Sunset", "2026-06-01")
            w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="successor-version"`, vm.defaultVersion.String()))
        }
        
        ctx := context.WithValue(r.Context(), versionKey{}, version)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (vm *VersioningMiddleware) extractVersion(r *http.Request) Version {
    if v := r.Header.Get("X-API-Version"); v != "" {
        if version, ok := ParseVersion(v); ok {
            return version
        }
    }
    
    if v := r.URL.Query().Get("version"); v != "" {
        if version, ok := ParseVersion(v); ok {
            return version
        }
    }
    
    path := r.URL.Path
    if strings.HasPrefix(path, "/v1") || strings.HasPrefix(path, "/v1.") {
        return Version{Major: 1, Minor: 0}
    }
    
    return vm.defaultVersion
}

func (vm *VersioningMiddleware) isSupported(v Version) bool {
    for _, sv := range vm.supportedVersions {
        if sv.Major == v.Major && sv.Minor <= v.Minor {
            return true
        }
    }
    return false
}

func (vm *VersioningMiddleware) isDeprecated(v Version) bool {
    for _, dv := range vm.deprecatedVersions {
        if dv == v {
            return true
        }
    }
    return false
}

type versionKey struct{}

func GetVersion(ctx context.Context) Version {
    if v, ok := ctx.Value(versionKey{}).(Version); ok {
        return v
    }
    return Version{Major: 1, Minor: 0}
}
```

### 5.4 SDK Generation

```go
// scripts/generate-sdk/main.go
package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

type SDKConfig struct {
    Name       string
    Language   string
    Generator  string
    OutputDir  string
    PackageName string
}

var sdkConfigs = []SDKConfig{
    {
        Name:       "go",
        Language:   "Go",
        Generator:  "go",
        OutputDir:  "sdks/go",
        PackageName: "aibhq",
    },
    {
        Name:       "typescript",
        Language:   "TypeScript",
        Generator:  "typescript-axios",
        OutputDir:  "sdks/typescript",
        PackageName: "@aibhq/sdk",
    },
    {
        Name:       "python",
        Language:   "Python",
        Generator:  "python",
        OutputDir:  "sdks/python",
        PackageName: "aibhq",
    },
}

func main() {
    specPath := "docs/api/openapi.yaml"
    
    for _, cfg := range sdkConfigs {
        fmt.Printf("Generating %s SDK...\n", cfg.Language)
        
        cmd := exec.Command("openapi-generator-cli", "generate",
            "-i", specPath,
            "-g", cfg.Generator,
            "-o", cfg.OutputDir,
            "--additional-properties", fmt.Sprintf("packageName=%s", cfg.PackageName),
        )
        
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        
        if err := cmd.Run(); err != nil {
            fmt.Printf("Error generating %s SDK: %v\n", cfg.Language, err)
            os.Exit(1)
        }
        
        fmt.Printf("✓ %s SDK generated at %s\n", cfg.Language, cfg.OutputDir)
    }
}
```

---

## 6. Implementation Priority

### 6.1 Phase 1A: Foundation (Weeks 1-4)

| Week | Tasks | Deliverables |
|------|-------|--------------|
| 1 | Core interfaces, registry, base adapter | `pkg/integration/adapter/` package |
| 2 | OAuth manager, token store, API key manager | `pkg/integration/auth/` package |
| 3 | Webhook receiver, processor, queue | `pkg/integration/webhook/` package |
| 4 | Rate limiter, quota tracker | `pkg/integration/ratelimit/` package |

### 6.2 Phase 1B: Priority Integrations (Weeks 5-12)

| Week | Integration | Key Actions | Est. Effort |
|------|-------------|-------------|-------------|
| 5-6 | Slack | Messages, Channels, Reactions | 2 weeks |
| 7-8 | Salesforce | Contacts, Leads, Opportunities, Query | 3 weeks |
| 8-9 | HubSpot | Contacts, Companies, Deals | 2 weeks |
| 9-10 | Linear | Issues, Projects, Comments | 2 weeks |
| 11-12 | Stripe | Customers, Payment Intents, Subscriptions | 2 weeks |

### 6.3 Phase 1C: Secondary Integrations (Weeks 13-18)

| Week | Integration | Key Actions | Est. Effort |
|------|-------------|-------------|-------------|
| 13-15 | Jira | Issues, Projects, Sprints | 3 weeks |
| 15-16 | Asana | Tasks, Projects, Workspaces | 2 weeks |
| 17-18 | Microsoft Teams | Messages, Channels, Chats | 3 weeks |
| 18-19 | QuickBooks | Customers, Invoices, Payments | 3 weeks |

### 6.4 Phase 1D: API Marketplace (Weeks 19-24)

| Week | Tasks | Deliverables |
|------|-------|--------------|
| 19-20 | OpenAPI spec, API server | Public API endpoints |
| 21-22 | Developer portal | Documentation site |
| 23 | SDK generation | Go, TypeScript, Python SDKs |
| 24 | Testing, documentation | Release candidate |

### 6.5 Dependency Graph

```
┌─────────────┐
│   Core SDK  │
│  (Week 1-4) │
└──────┬──────┘
       │
       ├──────────────────┬──────────────────┐
       │                  │                  │
       ▼                  ▼                  ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Slack     │    │  Salesforce │    │   Stripe    │
│  (Week 5-6) │    │  (Week 7-8) │    │ (Week 11-12)│
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       │                  ▼                  │
       │           ┌─────────────┐          │
       │           │   HubSpot   │          │
       │           │  (Week 8-9) │          │
       │           └──────┬──────┘          │
       │                  │                  │
       │                  ▼                  │
       │           ┌─────────────┐          │
       │           │   Linear    │          │
       │           │  (Week 9-10)│          │
       │           └──────┬──────┘          │
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │ API Layer   │
                   │ (Week 19-24)│
                   └─────────────┘
```

---

## 7. Security Considerations

### 7.1 Authentication Security

- All OAuth tokens encrypted at rest using AES-256-GCM
- API keys hashed using bcrypt with cost 12
- PKCE required for all OAuth flows
- State parameter with 10-minute expiry
- Token rotation on refresh

### 7.2 Webhook Security

- HMAC-SHA256 signature verification
- Timestamp validation (5-minute window)
- IP allowlisting for supported providers
- Replay attack prevention via event ID deduplication

### 7.3 Rate Limiting

- Per-tenant rate limits
- Per-integration quotas
- Concurrent request limits
- Automatic backoff on 429 responses

### 7.4 Data Protection

- Sensitive fields encrypted in database
- Credential rotation support
- Audit logging for all operations
- PII masking in logs

---

## Appendices

### A. Integration Checklist

```markdown
## New Integration Checklist

### Phase 1: Research
- [ ] Review API documentation
- [ ] Identify auth requirements
- [ ] List required scopes/permissions
- [ ] Document rate limits
- [ ] Identify webhook capabilities

### Phase 2: Design
- [ ] Define action schemas
- [ ] Map data transformations
- [ ] Design error handling
- [ ] Plan pagination strategy

### Phase 3: Implementation
- [ ] Create adapter struct
- [ ] Implement auth flow
- [ ] Implement actions
- [ ] Add webhook handler
- [ ] Write unit tests
- [ ] Write integration tests

### Phase 4: Documentation
- [ ] Write API docs
- [ ] Add examples
- [ ] Update integration list
- [ ] Add to SDK

### Phase 5: Deployment
- [ ] Code review
- [ ] Security review
- [ ] Load testing
- [ ] Staging deployment
- [ ] Production deployment
```

### B. Error Codes Reference

| Code | HTTP Status | Description | Retryable |
|------|-------------|-------------|-----------|
| `INTEGRATION_NOT_FOUND` | 404 | Integration does not exist | No |
| `INVALID_CREDENTIALS` | 401 | Authentication failed | No |
| `TOKEN_EXPIRED` | 401 | OAuth token has expired | Yes |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit hit | Yes |
| `QUOTA_EXCEEDED` | 429 | Usage quota exceeded | Yes |
| `PROVIDER_ERROR` | 502 | Upstream provider error | Yes |
| `TIMEOUT` | 504 | Request timed out | Yes |
| `VALIDATION_ERROR` | 400 | Invalid request parameters | No |
| `ACTION_NOT_SUPPORTED` | 400 | Action not available | No |

### C. Configuration Example

```yaml
integrations:
  salesforce:
    enabled: true
    auth:
      type: oauth2
      client_id: "${SALESFORCE_CLIENT_ID}"
      client_secret: "${SALESFORCE_CLIENT_SECRET}"
      sandbox: false
      scopes:
        - api
        - refresh_token
        - offline_access
    rate_limit:
      rps: 25
      burst: 50
      daily_quota: 100000
    webhooks:
      enabled: true
      events:
        - create
        - update
        - delete

  stripe:
    enabled: true
    auth:
      type: api_key
      key: "${STRIPE_SECRET_KEY}"
      test_mode: false
    rate_limit:
      rps: 100
      burst: 200
    webhooks:
      enabled: true
      secret: "${STRIPE_WEBHOOK_SECRET}"
      events:
        - customer.created
        - invoice.paid
        - payment_intent.succeeded
```

---

*Document maintained by AI Business HQ Engineering Team*
