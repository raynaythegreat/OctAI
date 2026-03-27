# OctAi - Multi-Tenant SaaS Architecture Design

> **Version**: 1.0  
> **Status**: Draft for MVP  
> **Last Updated**: March 2026

---

## Table of Contents

1. [Overview](#1-overview)
2. [Multi-Tenancy Model](#2-multi-tenancy-model)
3. [Authentication & Authorization](#3-authentication--authorization)
4. [Billing & Subscription](#4-billing--subscription)
5. [Database Schema Changes](#5-database-schema-changes)
6. [API Changes](#6-api-changes)
7. [Implementation Phases](#7-implementation-phases)

---

## 1. Overview

### 1.1 Goals

- Transform OctAi from single-tenant CLI tool to multi-tenant SaaS platform
- Enable organization-based workspaces with team collaboration
- Implement usage-based billing with Stripe integration
- Maintain backward compatibility with existing CLI/single-user deployments

### 1.2 Design Principles

- **Shared Database with Tenant IDs**: Cost-effective for MVP, simpler migrations
- **Row-Level Security**: All queries filtered by `organization_id`
- **Soft Deletes**: Preserve data for audit/recovery
- **Async Operations**: Billing, usage metering via background jobs

### 1.3 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway / Load Balancer               │
└─────────────────────────────────────────────────────────────────┘
                                   │
┌──────────────────────────────────┼──────────────────────────────┐
│                          API Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Auth      │  │    Org      │  │  Billing    │              │
│  │  Handlers   │  │  Handlers   │  │  Handlers   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Agent     │  │  Channels   │  │   Usage     │              │
│  │  Handlers   │  │  Handlers   │  │  Handlers   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└──────────────────────────────────────────────────────────────────┘
                                   │
┌──────────────────────────────────┼──────────────────────────────┐
│                        Service Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ TenantCtx   │  │    RBAC     │  │   Stripe    │              │
│  │  Middleware │  │   Service   │  │   Client    │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Agent     │  │  Channels   │  │   Usage     │              │
│  │   Service   │  │   Service   │  │  Metering   │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└──────────────────────────────────────────────────────────────────┘
                                   │
┌──────────────────────────────────┼──────────────────────────────┐
│                        Data Layer                                │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    PostgreSQL (Shared)                     │  │
│  │   organizations │ users │ memberships │ subscriptions     │  │
│  │   usage_events │ agents │ sessions │ configs             │  │
│  └───────────────────────────────────────────────────────────┘  │
│  ┌─────────────────────┐  ┌─────────────────────┐               │
│  │       Redis         │  │     Object Store    │               │
│  │  (Sessions/Cache)   │  │  (Files/Media)      │               │
│  └─────────────────────┘  └─────────────────────┘               │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Multi-Tenancy Model

### 2.1 Organization/Workspace Structure

```
Organization (Tenant)
├── Settings (config, branding)
├── Members (users with roles)
├── Agents (AI configurations)
├── Channels (connected platforms)
├── Sessions (conversation history)
├── Skills (custom plugins)
└── Usage/Billing
```

### 2.2 Tenant Identification

Tenants are identified via:

1. **Subdomain**: `acme.aibhq.com` → extracts `acme` as organization slug
2. **Header**: `X-Organization-ID: org_abc123`
3. **JWT Claim**: `org_id` in access token

```go
type TenantContext struct {
    OrganizationID   string
    OrganizationSlug string
    UserID           string
    Role             Role
    SubscriptionTier string
}

func (t *TenantContext) CanAccessResource(resourceOrgID string) bool {
    return t.OrganizationID == resourceOrgID
}
```

### 2.3 Data Isolation Strategy

**Decision: Shared Database with Tenant IDs**

| Approach | Pros | Cons |
|----------|------|------|
| **Shared DB + Tenant IDs** ✓ | Simple migrations, cost-effective, easier queries | Requires careful query filtering |
| Separate Schemas | Better isolation | Complex migrations, more connections |
| Separate Databases | Maximum isolation | Expensive, complex ops |

**Implementation**:

```go
type TenantScopedModel struct {
    OrganizationID string `gorm:"index;not null"`
}

func TenantScope(db *gorm.DB, orgID string) *gorm.DB {
    return db.Where("organization_id = ?", orgID)
}

func (m *Agent) BeforeCreate(tx *gorm.DB) error {
    ctx := tx.Statement.Context
    if tenantCtx, ok := ctx.Value(TenantCtxKey).(*TenantContext); ok {
        m.OrganizationID = tenantCtx.OrganizationID
    }
    return nil
}
```

### 2.4 Tenant Provisioning Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Sign Up   │────▶│   Create    │────▶│   Create    │
│   Request   │     │    User     │     │    Org      │
└─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Send Invite│◀────│   Create    │◀────│   Create    │
│   Email     │     │ Subscription│     │  Membership │
└─────────────┘     └─────────────┘     └─────────────┘
```

```go
func (s *ProvisioningService) ProvisionOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
    tx := s.db.Begin()
    defer tx.Rollback()
    
    org := &Organization{
        ID:   uuid.New().String(),
        Slug: req.Slug,
        Name: req.Name,
    }
    if err := tx.Create(org).Error; err != nil {
        return nil, err
    }
    
    membership := &Membership{
        OrganizationID: org.ID,
        UserID:         req.OwnerUserID,
        Role:           RoleOwner,
    }
    if err := tx.Create(membership).Error; err != nil {
        return nil, err
    }
    
    subscription := &Subscription{
        OrganizationID: org.ID,
        Tier:           TierFree,
        Status:         StatusActive,
    }
    if err := tx.Create(subscription).Error; err != nil {
        return nil, err
    }
    
    tx.Commit()
    return org, nil
}
```

---

## 3. Authentication & Authorization

### 3.1 User Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| **Owner** | Organization creator | Full access, billing, delete org |
| **Admin** | Organization manager | Manage members, configure all, no billing |
| **Member** | Standard user | Use agents, manage own sessions |
| **Viewer** | Read-only access | View conversations, analytics |

### 3.2 Permission Matrix

| Resource | Owner | Admin | Member | Viewer |
|----------|-------|-------|--------|--------|
| **Organization** |
| View | ✓ | ✓ | ✓ | ✓ |
| Update settings | ✓ | ✓ | ✗ | ✗ |
| Delete org | ✓ | ✗ | ✗ | ✗ |
| Manage billing | ✓ | ✗ | ✗ | ✗ |
| **Members** |
| List members | ✓ | ✓ | ✓ | ✓ |
| Invite members | ✓ | ✓ | ✗ | ✗ |
| Remove members | ✓ | ✓ | ✗ | ✗ |
| Change roles | ✓ | ✓ | ✗ | ✗ |
| **Agents** |
| Create/edit/delete | ✓ | ✓ | ✓ | ✗ |
| Use agents | ✓ | ✓ | ✓ | ✓ |
| **Sessions** |
| View own sessions | ✓ | ✓ | ✓ | ✓ |
| View all sessions | ✓ | ✓ | ✗ | ✗ |
| Delete sessions | ✓ | ✓ | ✓ | ✗ |
| **Channels** |
| Configure channels | ✓ | ✓ | ✗ | ✗ |
| Use channels | ✓ | ✓ | ✓ | ✗ |
| **Usage & Analytics** |
| View usage | ✓ | ✓ | ✓ | ✓ |
| Export data | ✓ | ✓ | ✗ | ✗ |

### 3.3 Permission Implementation

```go
type Permission string

const (
    PermOrgRead       Permission = "org:read"
    PermOrgWrite      Permission = "org:write"
    PermOrgDelete     Permission = "org:delete"
    PermBillingManage Permission = "billing:manage"
    PermMembersManage Permission = "members:manage"
    PermAgentsManage  Permission = "agents:manage"
    PermAgentsUse     Permission = "agents:use"
    PermSessionsAll   Permission = "sessions:all"
    PermChannelsCfg   Permission = "channels:configure"
    PermAnalytics     Permission = "analytics:view"
)

var RolePermissions = map[Role][]Permission{
    RoleOwner: {
        PermOrgRead, PermOrgWrite, PermOrgDelete, PermBillingManage,
        PermMembersManage, PermAgentsManage, PermAgentsUse,
        PermSessionsAll, PermChannelsCfg, PermAnalytics,
    },
    RoleAdmin: {
        PermOrgRead, PermOrgWrite,
        PermMembersManage, PermAgentsManage, PermAgentsUse,
        PermSessionsAll, PermChannelsCfg, PermAnalytics,
    },
    RoleMember: {
        PermOrgRead, PermAgentsManage, PermAgentsUse, PermAnalytics,
    },
    RoleViewer: {
        PermOrgRead, PermAnalytics,
    },
}

func (r *RBACService) HasPermission(ctx context.Context, perm Permission) bool {
    tc := GetTenantContext(ctx)
    if tc == nil {
        return false
    }
    perms, ok := RolePermissions[tc.Role]
    if !ok {
        return false
    }
    for _, p := range perms {
        if p == perm {
            return true
        }
    }
    return false
}

func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        rbac := c.MustGet("rbac").(*RBACService)
        if !rbac.HasPermission(c.Request.Context(), perm) {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 3.4 SSO Integration Points

```go
type SSOProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(code string) (*SSOUser, error)
    GetUserInfo(token string) (*SSOUser, error)
}

type SSOUser struct {
    Email     string
    Name      string
    Provider  string
    ProviderID string
}

type SAMLProvider struct {
    sp *saml.ServiceProvider
}

func (s *SAMLProvider) GetAuthURL(state string) string {
    return s.sp.GetSSOBindingLocation(saml.HTTPRedirectBinding)
}

type OIDCProvider struct {
    config    *oauth2.Config
    provider  *oidc.Provider
    verifier  *oidc.IDTokenVerifier
}

func (o *OIDCProvider) ExchangeCode(code string) (*SSOUser, error) {
    token, err := o.config.Exchange(context.Background(), code)
    if err != nil {
        return nil, err
    }
    
    idToken, err := o.verifier.Verify(context.Background(), token.Extra("id_token").(string))
    if err != nil {
        return nil, err
    }
    
    var claims struct {
        Email string `json:"email"`
        Name  string `json:"name"`
        Sub   string `json:"sub"`
    }
    if err := idToken.Claims(&claims); err != nil {
        return nil, err
    }
    
    return &SSOUser{
        Email:      claims.Email,
        Name:       claims.Name,
        Provider:   "oidc",
        ProviderID: claims.Sub,
    }, nil
}
```

### 3.5 JWT Token Structure

```go
type AccessTokenClaims struct {
    jwt.RegisteredClaims
    UserID         string `json:"user_id"`
    OrganizationID string `json:"org_id"`
    Role           string `json:"role"`
    Tier           string `json:"tier"`
    Email          string `json:"email"`
}
```

---

## 4. Billing & Subscription

### 4.1 Pricing Tiers

| Feature | Free | Pro | Business | Enterprise |
|---------|------|-----|----------|------------|
| **Price** | $0 | $29/mo | $99/mo | Custom |
| **Users** | 1 | 5 | 25 | Unlimited |
| **Agents** | 1 | 5 | 20 | Unlimited |
| **Messages/mo** | 500 | 10,000 | 100,000 | Unlimited |
| **Channels** | 2 | 5 | 15 | All |
| **Storage** | 100MB | 5GB | 50GB | Custom |
| **Skills** | Built-in | + Custom | + Marketplace | + Private |
| **Support** | Community | Email | Priority | Dedicated |
| **SSO** | ✗ | ✗ | ✓ | ✓ |
| **Audit Logs** | ✗ | ✗ | ✓ | ✓ |
| **SLA** | ✗ | ✗ | 99.5% | 99.9% |

### 4.2 Feature Gates Implementation

```go
type TierLimits struct {
    MaxUsers        int
    MaxAgents       int
    MaxMessages     int
    MaxChannels     int
    MaxStorageBytes int64
    Features        []string
}

var TierConfig = map[string]TierLimits{
    TierFree: {
        MaxUsers:        1,
        MaxAgents:       1,
        MaxMessages:     500,
        MaxChannels:     2,
        MaxStorageBytes: 100 * 1024 * 1024,
        Features:        []string{"basic_skills"},
    },
    TierPro: {
        MaxUsers:        5,
        MaxAgents:       5,
        MaxMessages:     10000,
        MaxChannels:     5,
        MaxStorageBytes: 5 * 1024 * 1024 * 1024,
        Features:        []string{"basic_skills", "custom_skills"},
    },
    TierBusiness: {
        MaxUsers:        25,
        MaxAgents:       20,
        MaxMessages:     100000,
        MaxChannels:     15,
        MaxStorageBytes: 50 * 1024 * 1024 * 1024,
        Features:        []string{"basic_skills", "custom_skills", "marketplace", "sso", "audit_logs"},
    },
}

func (s *BillingService) CheckLimit(ctx context.Context, limitType string, current int) error {
    tc := GetTenantContext(ctx)
    limits := TierConfig[tc.SubscriptionTier]
    
    var max int
    switch limitType {
    case "users":
        max = limits.MaxUsers
    case "agents":
        max = limits.MaxAgents
    case "messages":
        max = limits.MaxMessages
    default:
        return nil
    }
    
    if current >= max {
        return ErrLimitExceeded{Limit: limitType, Max: max}
    }
    return nil
}

func (s *BillingService) HasFeature(ctx context.Context, feature string) bool {
    tc := GetTenantContext(ctx)
    limits := TierConfig[tc.SubscriptionTier]
    for _, f := range limits.Features {
        if f == feature {
            return true
        }
    }
    return false
}
```

### 4.3 Usage Metering

```go
type UsageEvent struct {
    ID             string    `gorm:"primaryKey"`
    OrganizationID string    `gorm:"index;not null"`
    EventType      string    `gorm:"index"` // message, token_input, token_output, storage
    Quantity       int64
    Metadata       JSONB
    CreatedAt      time.Time `gorm:"index"`
}

type UsageMeteringService struct {
    db       *gorm.DB
    redis    *redis.Client
    flushInt time.Duration
}

func (s *UsageMeteringService) Record(ctx context.Context, orgID, eventType string, qty int64, meta map[string]any) error {
    event := &UsageEvent{
        ID:             uuid.New().String(),
        OrganizationID: orgID,
        EventType:      eventType,
        Quantity:       qty,
        Metadata:       meta,
        CreatedAt:      time.Now(),
    }
    
    key := fmt.Sprintf("usage:buffer:%s", orgID)
    data, _ := json.Marshal(event)
    return s.redis.RPush(ctx, key, data).Err()
}

func (s *UsageMeteringService) Flush(ctx context.Context) error {
    keys, _ := s.redis.Keys(ctx, "usage:buffer:*").Result()
    
    for _, key := range keys {
        orgID := strings.TrimPrefix(key, "usage:buffer:")
        
        for {
            result, err := s.redis.LPop(ctx, key).Result()
            if err == redis.Nil {
                break
            }
            if err != nil {
                return err
            }
            
            var event UsageEvent
            if err := json.Unmarshal([]byte(result), &event); err != nil {
                continue
            }
            
            if err := s.db.Create(&event).Error; err != nil {
                s.redis.RPush(ctx, key, result)
                return err
            }
        }
    }
    return nil
}

func (s *UsageMeteringService) GetMonthlyUsage(ctx context.Context, orgID string) (*MonthlyUsage, error) {
    startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
    
    var usage []UsageAggregation
    err := s.db.Model(&UsageEvent{}).
        Select("event_type, SUM(quantity) as total").
        Where("organization_id = ? AND created_at >= ?", orgID, startOfMonth).
        Group("event_type").
        Scan(&usage).Error
    
    return &MonthlyUsage{ByType: usage}, err
}
```

### 4.4 Stripe Integration

```go
type StripeService struct {
    client *stripe.Client
    db     *gorm.DB
}

func (s *StripeService) CreateCustomer(ctx context.Context, org *Organization, email string) (string, error) {
    params := &stripe.CustomerParams{
        Email: stripe.String(email),
        Metadata: map[string]string{
            "organization_id": org.ID,
        },
    }
    customer, err := s.client.Customers.New(params)
    if err != nil {
        return "", err
    }
    return customer.ID, nil
}

func (s *StripeService) CreateSubscription(ctx context.Context, orgID, priceID string) (*Subscription, error) {
    var org Organization
    if err := s.db.First(&org, "id = ?", orgID).Error; err != nil {
        return nil, err
    }
    
    params := &stripe.SubscriptionParams{
        Customer: stripe.String(org.StripeCustomerID),
        Items: []*stripe.SubscriptionItemsParams{
            {Price: stripe.String(priceID)},
        },
        Metadata: map[string]string{
            "organization_id": orgID,
        },
    }
    
    sub, err := s.client.Subscriptions.New(params)
    if err != nil {
        return nil, err
    }
    
    subscription := &Subscription{
        ID:               uuid.New().String(),
        OrganizationID:   orgID,
        StripeSubID:      sub.ID,
        Status:           string(sub.Status),
        Tier:             priceToTier(priceID),
        CurrentPeriodEnd: time.Unix(sub.CurrentPeriodEnd, 0),
    }
    
    return subscription, s.db.Create(subscription).Error
}

func (s *StripeService) HandleWebhook(ctx context.Context, event stripe.Event) error {
    switch event.Type {
    case "customer.subscription.updated":
        var sub stripe.Subscription
        json.Unmarshal(event.Data.Raw, &sub)
        return s.updateSubscriptionStatus(ctx, sub)
        
    case "customer.subscription.deleted":
        var sub stripe.Subscription
        json.Unmarshal(event.Data.Raw, &sub)
        return s.handleCancellation(ctx, sub)
        
    case "invoice.payment_failed":
        var inv stripe.Invoice
        json.Unmarshal(event.Data.Raw, &inv)
        return s.handlePaymentFailure(ctx, inv)
    }
    return nil
}

func (s *StripeService) CreateCheckoutSession(ctx context.Context, orgID, priceID string) (string, error) {
    var org Organization
    if err := s.db.First(&org, "id = ?", orgID).Error; err != nil {
        return "", err
    }
    
    params := &stripe.CheckoutSessionParams{
        Customer: stripe.String(org.StripeCustomerID),
        Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
        },
        SuccessURL: stripe.String(fmt.Sprintf("https://%s.aibhq.com/settings/billing?success=true", org.Slug)),
        CancelURL:  stripe.String(fmt.Sprintf("https://%s.aibhq.com/settings/billing?canceled=true", org.Slug)),
    }
    
    session, err := s.client.CheckoutSessions.New(params)
    if err != nil {
        return "", err
    }
    return session.URL, nil
}
```

---

## 5. Database Schema Changes

### 5.1 Organizations Table

```sql
CREATE TABLE organizations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug              VARCHAR(63) UNIQUE NOT NULL,
    name              VARCHAR(255) NOT NULL,
    logo_url          VARCHAR(1024),
    settings          JSONB DEFAULT '{}',
    
    -- Billing
    stripe_customer_id VARCHAR(255),
    
    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    
    CONSTRAINT slug_format CHECK (slug ~ '^[a-z0-9](-?[a-z0-9])*$')
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_stripe_customer ON organizations(stripe_customer_id);
```

### 5.2 Users Table

```sql
CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             VARCHAR(255) UNIQUE NOT NULL,
    password_hash     VARCHAR(255),
    name              VARCHAR(255),
    avatar_url        VARCHAR(1024),
    
    -- Auth providers
    auth_provider     VARCHAR(50) DEFAULT 'local', -- local, google, github, saml, oidc
    auth_provider_id  VARCHAR(255),
    
    -- Status
    email_verified    BOOLEAN DEFAULT FALSE,
    status            VARCHAR(20) DEFAULT 'active', -- active, suspended, pending
    
    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    last_login_at     TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_auth_provider ON users(auth_provider, auth_provider_id);
```

### 5.3 Memberships Table

```sql
CREATE TYPE membership_role AS ENUM ('owner', 'admin', 'member', 'viewer');

CREATE TABLE memberships (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role              membership_role NOT NULL DEFAULT 'member',
    
    -- Invitation
    invited_by        UUID REFERENCES users(id),
    invited_at        TIMESTAMPTZ,
    accepted_at       TIMESTAMPTZ,
    
    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(organization_id, user_id)
);

CREATE INDEX idx_memberships_org ON memberships(organization_id);
CREATE INDEX idx_memberships_user ON memberships(user_id);
CREATE INDEX idx_memberships_role ON memberships(organization_id, role);
```

### 5.4 Subscriptions Table

```sql
CREATE TYPE subscription_status AS ENUM ('active', 'past_due', 'canceled', 'incomplete', 'trialing');
CREATE TYPE subscription_tier AS ENUM ('free', 'pro', 'business', 'enterprise');

CREATE TABLE subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Stripe
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_price_id       VARCHAR(255),
    
    -- Status
    status              subscription_status NOT NULL DEFAULT 'active',
    tier                subscription_tier NOT NULL DEFAULT 'free',
    
    -- Billing cycle
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    canceled_at         TIMESTAMPTZ
);

CREATE INDEX idx_subscriptions_org ON subscriptions(organization_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
```

### 5.5 Usage Tracking Tables

```sql
CREATE TABLE usage_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    event_type        VARCHAR(50) NOT NULL, -- message, token_input, token_output, storage_bytes, api_call
    quantity          BIGINT NOT NULL DEFAULT 1,
    
    -- Context
    agent_id          UUID,
    channel           VARCHAR(50),
    model             VARCHAR(100),
    
    metadata          JSONB DEFAULT '{}',
    
    created_at        TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE usage_events_2026_03 PARTITION OF usage_events
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE INDEX idx_usage_events_org_time ON usage_events(organization_id, created_at);
CREATE INDEX idx_usage_events_type ON usage_events(organization_id, event_type, created_at);

-- Aggregated usage for quick lookups
CREATE TABLE usage_summaries (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    period_start      DATE NOT NULL,
    period_end        DATE NOT NULL,
    
    message_count     BIGINT DEFAULT 0,
    token_input       BIGINT DEFAULT 0,
    token_output      BIGINT DEFAULT 0,
    storage_bytes     BIGINT DEFAULT 0,
    api_calls         BIGINT DEFAULT 0,
    
    updated_at        TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(organization_id, period_start)
);

CREATE INDEX idx_usage_summaries_org ON usage_summaries(organization_id, period_start);
```

### 5.6 Modified Existing Tables

```sql
-- Add organization_id to existing tables
ALTER TABLE agents ADD COLUMN organization_id UUID REFERENCES organizations(id);
ALTER TABLE agents ADD COLUMN created_by UUID REFERENCES users(id);
CREATE INDEX idx_agents_org ON agents(organization_id);

ALTER TABLE sessions ADD COLUMN organization_id UUID REFERENCES organizations(id);
ALTER TABLE sessions ADD COLUMN user_id UUID REFERENCES users(id);
CREATE INDEX idx_sessions_org ON sessions(organization_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);

ALTER TABLE skills ADD COLUMN organization_id UUID REFERENCES organizations(id);
ALTER TABLE skills ADD COLUMN is_shared BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_skills_org ON skills(organization_id);

ALTER TABLE channel_configs ADD COLUMN organization_id UUID REFERENCES organizations(id);
CREATE INDEX idx_channel_configs_org ON channel_configs(organization_id);

ALTER TABLE memory_entries ADD COLUMN organization_id UUID REFERENCES organizations(id);
CREATE INDEX idx_memory_entries_org ON memory_entries(organization_id);
```

### 5.7 GORM Models

```go
type Organization struct {
    ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Slug            string         `gorm:"uniqueIndex;size:63;not null"`
    Name            string         `gorm:"size:255;not null"`
    LogoURL         *string        `gorm:"size:1024"`
    Settings        JSONB          `gorm:"type:jsonb;default:'{}'"`
    StripeCustomerID *string       `gorm:"size:255"`
    
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
    
    // Relations
    Memberships     []Membership
    Subscription    *Subscription
    Agents          []Agent
}

type User struct {
    ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email           string         `gorm:"uniqueIndex;size:255;not null"`
    PasswordHash    *string        `gorm:"size:255"`
    Name            *string        `gorm:"size:255"`
    AvatarURL       *string        `gorm:"size:1024"`
    AuthProvider    string         `gorm:"size:50;default:'local'"`
    AuthProviderID  *string        `gorm:"size:255"`
    EmailVerified   bool           `gorm:"default:false"`
    Status          string         `gorm:"size:20;default:'active'"`
    
    CreatedAt       time.Time
    UpdatedAt       time.Time
    LastLoginAt     *time.Time
    DeletedAt       gorm.DeletedAt `gorm:"index"`
    
    // Relations
    Memberships     []Membership
}

type Membership struct {
    ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OrganizationID string    `gorm:"not null;index"`
    UserID         string    `gorm:"not null;index"`
    Role           Role      `gorm:"type:membership_role;default:'member'"`
    
    InvitedBy      *string
    InvitedAt      *time.Time
    AcceptedAt     *time.Time
    
    CreatedAt      time.Time
    UpdatedAt      time.Time
    
    Organization   Organization `gorm:"foreignKey:OrganizationID"`
    User           User         `gorm:"foreignKey:UserID"`
}

type Subscription struct {
    ID                   string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OrganizationID       string         `gorm:"not null;uniqueIndex"`
    StripeSubscriptionID *string        `gorm:"uniqueIndex;size:255"`
    StripePriceID        *string        `gorm:"size:255"`
    Status               SubStatus      `gorm:"type:subscription_status;default:'active'"`
    Tier                 Tier           `gorm:"type:subscription_tier;default:'free'"`
    CurrentPeriodStart   *time.Time
    CurrentPeriodEnd     *time.Time
    CancelAtPeriodEnd    bool           `gorm:"default:false"`
    
    CreatedAt            time.Time
    UpdatedAt            time.Time
    CanceledAt           *time.Time
    
    Organization         Organization `gorm:"foreignKey:OrganizationID"`
}

type UsageEvent struct {
    ID             string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    OrganizationID string    `gorm:"not null;index"`
    EventType      string    `gorm:"size:50;not null;index"`
    Quantity       int64     `gorm:"not null;default:1"`
    AgentID        *string
    Channel        *string   `gorm:"size:50"`
    Model          *string   `gorm:"size:100"`
    Metadata       JSONB     `gorm:"type:jsonb"`
    CreatedAt      time.Time `gorm:"index"`
}
```

---

## 6. API Changes

### 6.1 New Endpoints

#### Authentication

```
POST   /api/v1/auth/register          - Register new user
POST   /api/v1/auth/login             - Login (email/password)
POST   /api/v1/auth/logout            - Logout
POST   /api/v1/auth/refresh           - Refresh access token
POST   /api/v1/auth/forgot-password   - Request password reset
POST   /api/v1/auth/reset-password    - Reset password with token
GET    /api/v1/auth/me                - Get current user
GET    /api/v1/auth/sso/:provider     - Initiate SSO flow
GET    /api/v1/auth/sso/:provider/callback - SSO callback
```

#### Organizations

```
POST   /api/v1/organizations              - Create organization
GET    /api/v1/organizations              - List user's organizations
GET    /api/v1/organizations/:id          - Get organization
PATCH  /api/v1/organizations/:id          - Update organization
DELETE /api/v1/organizations/:id          - Delete organization
GET    /api/v1/organizations/:id/usage    - Get usage stats
```

#### Memberships

```
GET    /api/v1/organizations/:orgId/memberships        - List members
POST   /api/v1/organizations/:orgId/memberships/invite - Invite member
PATCH  /api/v1/organizations/:orgId/memberships/:id    - Update role
DELETE /api/v1/organizations/:orgId/memberships/:id    - Remove member
POST   /api/v1/organizations/:orgId/memberships/:id/accept - Accept invite
```

#### Billing

```
GET    /api/v1/organizations/:orgId/billing/subscription   - Get subscription
POST   /api/v1/organizations/:orgId/billing/checkout       - Create checkout session
POST   /api/v1/organizations/:orgId/billing/portal         - Create portal session
POST   /api/v1/organizations/:orgId/billing/cancel         - Cancel subscription
POST   /api/v1/webhooks/stripe                             - Stripe webhook
```

#### Tenanted Resources (existing, modified)

```
GET    /api/v1/agents                    - List agents (tenant-scoped)
POST   /api/v1/agents                    - Create agent
GET    /api/v1/agents/:id                - Get agent
PATCH  /api/v1/agents/:id                - Update agent
DELETE /api/v1/agents/:id                - Delete agent

GET    /api/v1/sessions                  - List sessions (tenant-scoped)
POST   /api/v1/sessions                  - Create session
GET    /api/v1/sessions/:id              - Get session
DELETE /api/v1/sessions/:id              - Delete session

GET    /api/v1/channels                  - List channel configs
POST   /api/v1/channels                  - Configure channel
DELETE /api/v1/channels/:id              - Remove channel config
```

### 6.2 Tenant Context Middleware

```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var orgID string
        
        // 1. Try header
        if id := c.GetHeader("X-Organization-ID"); id != "" {
            orgID = id
        }
        
        // 2. Try JWT claims
        if orgID == "" {
            if claims, ok := c.Get("claims"); ok {
                if ac, ok := claims.(*AccessTokenClaims); ok {
                    orgID = ac.OrganizationID
                }
            }
        }
        
        // 3. Try subdomain
        if orgID == "" {
            host := c.Request.Host
            if idx := strings.Index(host, ".aibhq.com"); idx > 0 {
                slug := host[:idx]
                org, _ := orgCache.GetBySlug(slug)
                if org != nil {
                    orgID = org.ID
                }
            }
        }
        
        if orgID == "" {
            c.JSON(400, gin.H{"error": "organization context required"})
            c.Abort()
            return
        }
        
        // Load tenant context
        tc, err := tenantService.GetContext(c.Request.Context(), orgID)
        if err != nil {
            c.JSON(404, gin.H{"error": "organization not found"})
            c.Abort()
            return
        }
        
        c.Set("tenant", tc)
        c.Next()
    }
}

func TenantDBMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        tc, exists := c.Get("tenant")
        if !exists {
            c.Next()
            return
        }
        
        tenantCtx := tc.(*TenantContext)
        
        // Attach tenant context to GORM
        ctx := context.WithValue(c.Request.Context(), TenantCtxKey, tenantCtx)
        c.Request = c.Request.WithContext(ctx)
        
        // Scope all queries
        scopedDB := db.Scopes(func(d *gorm.DB) *gorm.DB {
            return d.Where("organization_id = ?", tenantCtx.OrganizationID)
        })
        
        c.Set("db", scopedDB)
        c.Next()
    }
}
```

### 6.3 Rate Limiting Per Tenant

```go
type TenantRateLimiter struct {
    redis    *redis.Client
    limits   map[string]RateLimit
}

type RateLimit struct {
    Requests int
    Window   time.Duration
}

var TierRateLimits = map[string]RateLimit{
    TierFree:      {Requests: 100, Window: time.Minute},
    TierPro:       {Requests: 500, Window: time.Minute},
    TierBusiness:  {Requests: 2000, Window: time.Minute},
    TierEnterprise: {Requests: 10000, Window: time.Minute},
}

func (r *TenantRateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tc, ok := c.Get("tenant")
        if !ok {
            c.Next()
            return
        }
        
        tenant := tc.(*TenantContext)
        limit := TierRateLimits[tenant.SubscriptionTier]
        
        key := fmt.Sprintf("ratelimit:%s:%s", tenant.OrganizationID, c.FullPath())
        
        count, err := r.redis.Incr(c.Request.Context(), key).Result()
        if err != nil {
            c.Next()
            return
        }
        
        if count == 1 {
            r.redis.Expire(c.Request.Context(), key, limit.Window)
        }
        
        c.Header("X-RateLimit-Limit", strconv.Itoa(limit.Requests))
        c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, limit.Requests-int(count))))
        c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limit.Window).Unix(), 10))
        
        if count > int64(limit.Requests) {
            c.JSON(429, gin.H{
                "error": "rate limit exceeded",
                "retry_after": limit.Window.Seconds(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 6.4 API Response Format

```go
type APIResponse struct {
    Data   interface{} `json:"data,omitempty"`
    Error  *APIError   `json:"error,omitempty"`
    Meta   *Meta       `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

type Meta struct {
    Total      int64 `json:"total,omitempty"`
    Page       int   `json:"page,omitempty"`
    PageSize   int   `json:"page_size,omitempty"`
    TotalPages int   `json:"total_pages,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(200, APIResponse{Data: data})
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
    c.JSON(200, APIResponse{Data: data, Meta: meta})
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(201, APIResponse{Data: data})
}

func Error(c *gin.Context, status int, code, message string, details ...any) {
    resp := APIResponse{
        Error: &APIError{Code: code, Message: message},
    }
    if len(details) > 0 {
        resp.Error.Details = details[0]
    }
    c.JSON(status, resp)
}
```

### 6.5 Route Registration

```go
func RegisterRoutes(r *gin.Engine, services *Services) {
    api := r.Group("/api/v1")
    
    // Public routes
    auth := api.Group("/auth")
    {
        auth.POST("/register", services.Auth.Register)
        auth.POST("/login", services.Auth.Login)
        auth.POST("/refresh", services.Auth.Refresh)
        auth.POST("/forgot-password", services.Auth.ForgotPassword)
        auth.POST("/reset-password", services.Auth.ResetPassword)
        auth.GET("/sso/:provider", services.Auth.SSOBegin)
        auth.GET("/sso/:provider/callback", services.Auth.SSOCallback)
    }
    
    // Webhooks (public but verified)
    api.POST("/webhooks/stripe", services.Billing.HandleWebhook)
    
    // Authenticated routes
    protected := api.Group("")
    protected.Use(AuthMiddleware(services.Auth))
    {
        protected.GET("/auth/me", services.Auth.GetCurrentUser)
        protected.POST("/auth/logout", services.Auth.Logout)
        
        // Organizations
        orgs := protected.Group("/organizations")
        {
            orgs.GET("", services.Org.List)
            orgs.POST("", services.Org.Create)
        }
        
        org := protected.Group("/organizations/:orgId")
        org.Use(TenantMiddleware(), TenantDBMiddleware(services.DB), RBACMiddleware(services.RBAC))
        {
            org.GET("", RequirePermission(PermOrgRead), services.Org.Get)
            org.PATCH("", RequirePermission(PermOrgWrite), services.Org.Update)
            org.DELETE("", RequirePermission(PermOrgDelete), services.Org.Delete)
            org.GET("/usage", RequirePermission(PermAnalytics), services.Org.GetUsage)
            
            // Memberships
            org.GET("/memberships", RequirePermission(PermOrgRead), services.Membership.List)
            org.POST("/memberships/invite", RequirePermission(PermMembersManage), services.Membership.Invite)
            org.PATCH("/memberships/:id", RequirePermission(PermMembersManage), services.Membership.UpdateRole)
            org.DELETE("/memberships/:id", RequirePermission(PermMembersManage), services.Membership.Remove)
            
            // Billing
            org.GET("/billing/subscription", RequirePermission(PermBillingManage), services.Billing.GetSubscription)
            org.POST("/billing/checkout", RequirePermission(PermBillingManage), services.Billing.CreateCheckout)
            org.POST("/billing/portal", RequirePermission(PermBillingManage), services.Billing.CreatePortal)
            org.POST("/billing/cancel", RequirePermission(PermBillingManage), services.Billing.Cancel)
            
            // Agents
            org.GET("/agents", RequirePermission(PermAgentsUse), services.Agent.List)
            org.POST("/agents", RequirePermission(PermAgentsManage), services.Agent.Create)
            org.GET("/agents/:id", RequirePermission(PermAgentsUse), services.Agent.Get)
            org.PATCH("/agents/:id", RequirePermission(PermAgentsManage), services.Agent.Update)
            org.DELETE("/agents/:id", RequirePermission(PermAgentsManage), services.Agent.Delete)
            
            // Sessions
            org.GET("/sessions", RequirePermission(PermAgentsUse), services.Session.List)
            org.POST("/sessions", RequirePermission(PermAgentsUse), services.Session.Create)
            org.GET("/sessions/:id", RequirePermission(PermAgentsUse), services.Session.Get)
            org.DELETE("/sessions/:id", RequirePermission(PermAgentsUse), services.Session.Delete)
            
            // Channels
            org.GET("/channels", RequirePermission(PermOrgRead), services.Channel.List)
            org.POST("/channels", RequirePermission(PermChannelsCfg), services.Channel.Create)
            org.DELETE("/channels/:id", RequirePermission(PermChannelsCfg), services.Channel.Delete)
        }
    }
    
    // Rate limiting
    protected.Use((&TenantRateLimiter{Redis: services.Redis}).Middleware())
}
```

---

## 7. Implementation Phases

### Phase 1: Foundation (Week 1-2)

- [ ] Set up PostgreSQL database
- [ ] Create migration files for new tables
- [ ] Implement User model and authentication (email/password)
- [ ] Implement Organization model
- [ ] Implement basic JWT authentication

### Phase 2: Multi-Tenancy (Week 3-4)

- [ ] Implement TenantContext middleware
- [ ] Add organization_id to existing models
- [ ] Implement Membership model and CRUD
- [ ] Implement RBAC service
- [ ] Migrate existing auth to new system

### Phase 3: Billing (Week 5-6)

- [ ] Integrate Stripe SDK
- [ ] Implement Subscription model
- [ ] Create checkout session flow
- [ ] Handle Stripe webhooks
- [ ] Implement tier limits checking

### Phase 4: Usage & Metering (Week 7-8)

- [ ] Implement UsageEvent model
- [ ] Create usage metering service
- [ ] Build usage aggregation
- [ ] Create usage dashboard endpoints
- [ ] Implement rate limiting per tier

### Phase 5: SSO & Polish (Week 9-10)

- [ ] Implement OAuth2/OIDC providers (Google, GitHub)
- [ ] Add SAML support for Enterprise
- [ ] Build billing portal integration
- [ ] Performance optimization
- [ ] Security audit

---

## Appendix A: Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/aibhq?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Stripe
STRIPE_API_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# Pricing IDs
STRIPE_PRICE_PRO=price_xxx
STRIPE_PRICE_BUSINESS=price_xxx
STRIPE_PRICE_ENTERPRISE=price_xxx

# OAuth
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx

# App
APP_URL=https://app.aibhq.com
BASE_DOMAIN=aibhq.com
```

## Appendix B: Migration Strategy

For existing single-tenant deployments, provide migration path:

```sql
-- Create default organization for existing data
INSERT INTO organizations (id, slug, name)
SELECT 
    gen_random_uuid(),
    'default',
    'Default Organization'
WHERE NOT EXISTS (SELECT 1 FROM organizations);

-- Migrate existing agents to default org
UPDATE agents SET organization_id = (SELECT id FROM organizations WHERE slug = 'default')
WHERE organization_id IS NULL;

-- Create admin user from existing config
-- (Implementation depends on current auth method)
```
