package api

import (
	"net/http"
	"sync"

	"github.com/raynaythegreat/ai-business-hq/pkg/marketplace"
	"github.com/raynaythegreat/ai-business-hq/pkg/tenant"
	"github.com/raynaythegreat/ai-business-hq/web/backend/launcherconfig"
)

var postInitFuncs []func(h *Handler)

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
	marketplaceStore     marketplace.MarketplaceStore
	analyticsCache       map[string]interface{}
	membershipIDs        map[string]string
}

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

func (h *Handler) SetServerOptions(port int, public bool, publicExplicit bool, allowedCIDRs []string) {
	h.serverPort = port
	h.serverPublic = public
	h.serverPublicExplicit = publicExplicit
	h.serverCIDRs = append([]string(nil), allowedCIDRs...)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	h.registerConfigRoutes(mux)
	h.registerPicoRoutes(mux)
	h.registerGatewayRoutes(mux)
	h.registerSessionRoutes(mux)
	h.registerOAuthRoutes(mux)
	h.registerModelRoutes(mux)
	h.registerChannelRoutes(mux)
	h.registerSkillRoutes(mux)
	h.registerToolRoutes(mux)
	h.registerStartupRoutes(mux)
	h.registerLauncherConfigRoutes(mux)
	h.registerWeixinRoutes(mux)
	h.registerWecomRoutes(mux)
	h.registerOrganizationRoutes(mux)
	h.registerMembershipRoutes(mux)
	h.registerSubscriptionRoutes(mux)
	h.registerAnalyticsRoutes(mux)
	h.registerMarketplaceRoutes(mux)
}

func (h *Handler) Shutdown() {
	h.StopGateway()
}
