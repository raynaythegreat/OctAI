package api

import (
	"net/http"
	"sync"

	"github.com/raynaythegreat/ai-business-hq/pkg/tenant"
	"github.com/raynaythegreat/ai-business-hq/web/backend/launcherconfig"
)

var postInitFuncs []func(h *Handler)

// Handler serves HTTP API requests.
type Handler struct {
	configPath           string
	serverPort           int
	serverPublic         bool
	serverPublicExplicit bool
	serverCIDRs          []string
	oauthMu              sync.Mutex
	oauthFlows           map[string]*oauthFlow
	oauthState           map[string]string
	weixinMu             sync.Mutex
	weixinFlows          map[string]*weixinFlow
	wecomMu              sync.Mutex
	wecomFlows           map[string]*wecomFlow
	tenantStore          tenant.TenantStore
	analyticsCache       map[string]interface{}
	membershipIDs        map[string]string
}

// NewHandler creates an instance of the API handler.
func NewHandler(configPath string) *Handler {
	h := &Handler{
		configPath:     configPath,
		serverPort:     launcherconfig.DefaultPort,
		oauthFlows:     make(map[string]*oauthFlow),
		oauthState:     make(map[string]string),
		weixinFlows:    make(map[string]*weixinFlow),
		wecomFlows:     make(map[string]*wecomFlow),
		analyticsCache: make(map[string]interface{}),
		membershipIDs:  make(map[string]string),
	}
	for _, fn := range postInitFuncs {
		fn(h)
	}
	return h
}

// SetServerOptions stores current backend listen options for fallback behavior.
func (h *Handler) SetServerOptions(port int, public bool, publicExplicit bool, allowedCIDRs []string) {
	h.serverPort = port
	h.serverPublic = public
	h.serverPublicExplicit = publicExplicit
	h.serverCIDRs = append([]string(nil), allowedCIDRs...)
}

// RegisterRoutes binds all API endpoint handlers to the ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Config CRUD
	h.registerConfigRoutes(mux)

	// Pico Channel (WebSocket chat)
	h.registerPicoRoutes(mux)

	// Gateway process lifecycle
	h.registerGatewayRoutes(mux)

	// Session history
	h.registerSessionRoutes(mux)

	// OAuth login and credential management
	h.registerOAuthRoutes(mux)

	// Model list management
	h.registerModelRoutes(mux)

	// Channel catalog (for frontend navigation/config pages)
	h.registerChannelRoutes(mux)

	// Skills and tools support/actions
	h.registerSkillRoutes(mux)
	h.registerToolRoutes(mux)

	// OS startup / launch-at-login
	h.registerStartupRoutes(mux)

	// Launcher service parameters (port/public)
	h.registerLauncherConfigRoutes(mux)

	// WeChat QR login flow
	h.registerWeixinRoutes(mux)

	// WeCom QR login flow
	h.registerWecomRoutes(mux)

	// API v2 - SaaS features
	h.registerOrganizationRoutes(mux)
	h.registerMembershipRoutes(mux)
	h.registerSubscriptionRoutes(mux)
	h.registerAnalyticsRoutes(mux)
}

// Shutdown gracefully shuts down the handler, stopping the gateway if it was started by this handler.
func (h *Handler) Shutdown() {
	h.StopGateway()
}
