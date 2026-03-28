package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/raynaythegreat/ai-business-hq/pkg/config"
	"github.com/raynaythegreat/ai-business-hq/pkg/logger"
)

// registerModelRoutes binds model list management endpoints to the ServeMux.
func (h *Handler) registerModelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models", h.handleListModels)
	mux.HandleFunc("POST /api/models", h.handleAddModel)
	mux.HandleFunc("POST /api/models/default", h.handleSetDefaultModel)
	mux.HandleFunc("POST /api/models/{index}/test", h.handleTestModelKey)
	mux.HandleFunc("POST /api/models/{index}/rotate-key", h.handleRotateModelKey)
	mux.HandleFunc("PUT /api/models/{index}", h.handleUpdateModel)
	mux.HandleFunc("DELETE /api/models/{index}", h.handleDeleteModel)
	mux.HandleFunc("GET /api/models/auto", h.handleGetAutoRouting)
	mux.HandleFunc("POST /api/models/auto", h.handleSetAutoRouting)
}

// modelResponse is the JSON structure returned for each model in the list.
// All ModelConfig fields are included so the frontend can display and edit them.
type modelResponse struct {
	Index      int    `json:"index"`
	ModelName  string `json:"model_name"`
	Model      string `json:"model"`
	APIBase    string `json:"api_base,omitempty"`
	APIKey     string `json:"api_key"`
	Proxy      string `json:"proxy,omitempty"`
	AuthMethod string `json:"auth_method,omitempty"`
	// Advanced fields
	ConnectMode    string         `json:"connect_mode,omitempty"`
	Workspace      string         `json:"workspace,omitempty"`
	RPM            int            `json:"rpm,omitempty"`
	MaxTokensField string         `json:"max_tokens_field,omitempty"`
	RequestTimeout int            `json:"request_timeout,omitempty"`
	ThinkingLevel  string         `json:"thinking_level,omitempty"`
	ExtraBody      map[string]any `json:"extra_body,omitempty"`
	// Meta
	Configured bool `json:"configured"`
	IsDefault  bool `json:"is_default"`
	IsVirtual  bool `json:"is_virtual"`
}

// handleListModels returns all model_list entries with masked API keys.
//
//	GET /api/models
func (h *Handler) handleListModels(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	defaultModel := cfg.Agents.Defaults.GetModelName()
	configured := make([]bool, len(cfg.ModelList))

	var wg sync.WaitGroup
	wg.Add(len(cfg.ModelList))
	for i, m := range cfg.ModelList {
		go func(i int, m *config.ModelConfig) {
			defer wg.Done()
			configured[i] = isModelConfigured(m)
		}(i, m)
	}
	wg.Wait()

	configuredOnly := r.URL.Query().Get("configured_only") == "true"
	models := make([]modelResponse, 0, len(cfg.ModelList))
	for i, m := range cfg.ModelList {
		if configuredOnly && !configured[i] {
			continue
		}
		models = append(models, modelResponse{
			Index:          i,
			ModelName:      m.ModelName,
			Model:          m.Model,
			APIBase:        m.APIBase,
			APIKey:         maskAPIKey(m.APIKey()),
			Proxy:          m.Proxy,
			AuthMethod:     m.AuthMethod,
			ConnectMode:    m.ConnectMode,
			Workspace:      m.Workspace,
			RPM:            m.RPM,
			MaxTokensField: m.MaxTokensField,
			RequestTimeout: m.RequestTimeout,
			ThinkingLevel:  m.ThinkingLevel,
			ExtraBody:      m.ExtraBody,
			Configured:     configured[i],
			IsDefault:      m.ModelName == defaultModel,
			IsVirtual:      m.IsVirtual(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"models":        models,
		"total":         len(models),
		"default_model": defaultModel,
	})
}

// handleAddModel appends a new model configuration entry.
//
//	POST /api/models
func (h *Handler) handleAddModel(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	if mc.APIKey != "" {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	cfg.ModelList = append(cfg.ModelList, &mc.ModelConfig)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"index":  len(cfg.ModelList) - 1,
	})
}

// handleUpdateModel replaces a model configuration entry at the given index.
// If the request body omits api_key (or sends an empty string), the existing
// stored key is preserved so callers can update only api_base / proxy without
// exposing or clearing the secret.
//
//	PUT /api/models/{index}
func (h *Handler) handleUpdateModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ModelList)-1), http.StatusNotFound)
		return
	}

	// Preserve the existing API key when the caller omits it (empty string).
	// This lets the UI update api_base / proxy without clearing the stored secret.
	if mc.APIKey == "" {
		mc.ModelConfig.SetAPIKey(cfg.ModelList[idx].APIKey())
	} else {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}
	// Preserve existing ExtraBody when omitted (nil), but clear it when
	// the frontend sends an empty object {} to indicate the field should
	// be removed.
	if mc.ExtraBody == nil {
		mc.ExtraBody = cfg.ModelList[idx].ExtraBody
	} else if len(mc.ExtraBody) == 0 {
		mc.ExtraBody = nil
	}

	cfg.ModelList[idx] = &mc.ModelConfig

	logger.Debugf("update model config: %#v", mc.ModelConfig)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDeleteModel removes a model configuration entry at the given index.
//
//	DELETE /api/models/{index}
func (h *Handler) handleDeleteModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ModelList)-1), http.StatusNotFound)
		return
	}

	deletedModelName := cfg.ModelList[idx].ModelName

	cfg.ModelList = append(cfg.ModelList[:idx], cfg.ModelList[idx+1:]...)

	// If the deleted model was the default, clear it.
	if cfg.Agents.Defaults.ModelName == deletedModelName {
		cfg.Agents.Defaults.ModelName = ""
	}

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSetDefaultModel sets the default model for all agents.
//
//	POST /api/models/default
func (h *Handler) handleSetDefaultModel(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		ModelName string `json:"model_name"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.ModelName == "" {
		http.Error(w, "model_name is required", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify the model_name exists in model_list and is not a virtual model
	found := false
	isVirtual := false
	for _, m := range cfg.ModelList {
		if m.ModelName == req.ModelName {
			found = true
			isVirtual = m.IsVirtual()
			break
		}
	}
	if !found {
		http.Error(w, fmt.Sprintf("Model %q not found in model_list", req.ModelName), http.StatusNotFound)
		return
	}
	if isVirtual {
		http.Error(w, fmt.Sprintf("Cannot set virtual model %q as default", req.ModelName), http.StatusBadRequest)
		return
	}

	cfg.Agents.Defaults.ModelName = req.ModelName

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":        "ok",
		"default_model": req.ModelName,
	})
}

// handleTestModelKey verifies the API key for a model at the given index by
// making a live request to the provider's /models endpoint.
//
//	POST /api/models/{index}/test
func (h *Handler) handleTestModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range", idx), http.StatusNotFound)
		return
	}

	m := cfg.ModelList[idx]

	w.Header().Set("Content-Type", "application/json")

	// OAuth models — auth is managed via the credentials page, not API keys.
	if strings.EqualFold(strings.TrimSpace(m.AuthMethod), "oauth") {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"note":    "OAuth model — authentication managed via credentials page",
		})
		return
	}

	// Resolve the API base before checking the key so local models are
	// detected first (they don't need or use an API key).
	apiBase := resolveProviderAPIBase(m)

	// Skip local models — they don't support the /models endpoint.
	if apiBase != "" && hasLocalAPIBase(apiBase) {
		avail := probeLocalModelAvailability(m)
		json.NewEncoder(w).Encode(map[string]any{
			"success": avail,
			"models":  []string{},
			"note":    "Local model — connectivity tested via runtime probe",
		})
		return
	}

	apiKey := m.APIKey()
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "No API key configured for this model",
		})
		return
	}

	if apiBase == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Cannot determine API base URL for this provider",
		})
		return
	}

	models, err := listProviderModels(apiBase, apiKey)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"models":  models,
	})
}

// handleRotateModelKey replaces the API key for a chat model at the given index.
//
//	POST /api/models/{index}/rotate-key
func (h *Handler) handleRotateModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		NewAPIKey string `json:"new_api_key"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if req.NewAPIKey == "" {
		http.Error(w, "new_api_key is required", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ModelList)-1), http.StatusNotFound)
		return
	}

	cfg.ModelList[idx].SetAPIKey(req.NewAPIKey)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// maskAPIKey returns a masked version of an API key for safe display.
// Keys longer than 12 chars show prefix + last 4 chars: "sk-****abcd".
// Keys 9-12 chars show prefix + last 2 chars: "sk-****cd".
// Shorter keys are fully masked as "****".
// Empty keys return empty string.
// Ensure at least 40% of the key will not be displayed.
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}

	if len(key) <= 8 {
		return "****"
	}

	// Show first 3 chars and last 2 chars
	if len(key) <= 12 {
		return key[:3] + "****" + key[len(key)-2:]
	}

	// Show first 3 chars and last 4 chars
	return key[:3] + "****" + key[len(key)-4:]
}

// handleGetAutoRouting returns the current intelligent model routing configuration.
//
//	GET /api/models/auto
func (h *Handler) handleGetAutoRouting(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	routing := cfg.Agents.Defaults.Routing
	var enabled bool
	var lightModel string
	var threshold float64
	if routing != nil {
		enabled = routing.Enabled
		lightModel = routing.LightModel
		threshold = routing.Threshold
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"enabled":     enabled,
		"light_model": lightModel,
		"threshold":   threshold,
	})
}

// handleSetAutoRouting updates the intelligent model routing configuration.
//
//	POST /api/models/auto
func (h *Handler) handleSetAutoRouting(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Enabled    bool    `json:"enabled"`
		LightModel string  `json:"light_model"`
		Threshold  float64 `json:"threshold"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if cfg.Agents.Defaults.Routing == nil {
		cfg.Agents.Defaults.Routing = &config.RoutingConfig{}
	}
	cfg.Agents.Defaults.Routing.Enabled = req.Enabled
	cfg.Agents.Defaults.Routing.LightModel = req.LightModel
	cfg.Agents.Defaults.Routing.Threshold = req.Threshold

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":                   "ok",
		"gateway_restart_required": true,
	})
}

// ─── Image Model Routes ───────────────────────────────────────────────────────

// registerImageModelRoutes binds image model CRUD endpoints.
func (h *Handler) registerImageModelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/image-models", h.handleListImageModels)
	mux.HandleFunc("POST /api/image-models", h.handleAddImageModel)
	mux.HandleFunc("POST /api/image-models/default", h.handleSetDefaultImageModel)
	mux.HandleFunc("POST /api/image-models/{index}/test", h.handleTestImageModelKey)
	mux.HandleFunc("POST /api/image-models/{index}/rotate-key", h.handleRotateImageModelKey)
	mux.HandleFunc("PUT /api/image-models/{index}", h.handleUpdateImageModel)
	mux.HandleFunc("DELETE /api/image-models/{index}", h.handleDeleteImageModel)
}

// handleListImageModels returns all image_model_list entries with masked API keys.
//
//	GET /api/image-models
func (h *Handler) handleListImageModels(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	configured := make([]bool, len(cfg.ImageModelList))

	var wg sync.WaitGroup
	wg.Add(len(cfg.ImageModelList))
	for i, m := range cfg.ImageModelList {
		go func(i int, m *config.ModelConfig) {
			defer wg.Done()
			configured[i] = isModelConfigured(m)
		}(i, m)
	}
	wg.Wait()

	configuredOnly := r.URL.Query().Get("configured_only") == "true"
	models := make([]modelResponse, 0, len(cfg.ImageModelList))
	for i, m := range cfg.ImageModelList {
		if configuredOnly && !configured[i] {
			continue
		}
		models = append(models, modelResponse{
			Index:          i,
			ModelName:      m.ModelName,
			Model:          m.Model,
			APIBase:        m.APIBase,
			APIKey:         maskAPIKey(m.APIKey()),
			Proxy:          m.Proxy,
			AuthMethod:     m.AuthMethod,
			ConnectMode:    m.ConnectMode,
			Workspace:      m.Workspace,
			RPM:            m.RPM,
			MaxTokensField: m.MaxTokensField,
			RequestTimeout: m.RequestTimeout,
			ThinkingLevel:  m.ThinkingLevel,
			ExtraBody:      m.ExtraBody,
			Configured:     configured[i],
			IsDefault:      false,
			IsVirtual:      m.IsVirtual(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"models": models,
		"total":  len(models),
	})
}

// handleRotateImageModelKey replaces the API key for an image model at the given index.
//
//	POST /api/image-models/{index}/rotate-key
func (h *Handler) handleRotateImageModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		NewAPIKey string `json:"new_api_key"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if req.NewAPIKey == "" {
		http.Error(w, "new_api_key is required", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ImageModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ImageModelList)-1), http.StatusNotFound)
		return
	}

	cfg.ImageModelList[idx].SetAPIKey(req.NewAPIKey)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleAddImageModel appends a new image model configuration entry.
//
//	POST /api/image-models
func (h *Handler) handleAddImageModel(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	if mc.APIKey != "" {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	cfg.ImageModelList = append(cfg.ImageModelList, &mc.ModelConfig)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"index":  len(cfg.ImageModelList) - 1,
	})
}

// handleUpdateImageModel replaces an image model configuration entry at the given index.
//
//	PUT /api/image-models/{index}
func (h *Handler) handleUpdateImageModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ImageModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ImageModelList)-1), http.StatusNotFound)
		return
	}

	if mc.APIKey == "" {
		mc.ModelConfig.SetAPIKey(cfg.ImageModelList[idx].APIKey())
	} else {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}
	if mc.ExtraBody == nil {
		mc.ExtraBody = cfg.ImageModelList[idx].ExtraBody
	} else if len(mc.ExtraBody) == 0 {
		mc.ExtraBody = nil
	}

	cfg.ImageModelList[idx] = &mc.ModelConfig

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDeleteImageModel removes an image model configuration entry at the given index.
//
//	DELETE /api/image-models/{index}
func (h *Handler) handleDeleteImageModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ImageModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.ImageModelList)-1), http.StatusNotFound)
		return
	}

	cfg.ImageModelList = append(cfg.ImageModelList[:idx], cfg.ImageModelList[idx+1:]...)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSetDefaultImageModel acknowledges a set-default request for image models.
//
//	POST /api/image-models/default
func (h *Handler) handleSetDefaultImageModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleTestImageModelKey verifies the API key for an image model at the given index.
//
//	POST /api/image-models/{index}/test
func (h *Handler) handleTestImageModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.ImageModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range", idx), http.StatusNotFound)
		return
	}

	m := cfg.ImageModelList[idx]

	w.Header().Set("Content-Type", "application/json")

	if strings.EqualFold(strings.TrimSpace(m.AuthMethod), "oauth") {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"note":    "OAuth model — authentication managed via credentials page",
		})
		return
	}

	apiBase := resolveProviderAPIBase(m)

	if apiBase != "" && hasLocalAPIBase(apiBase) {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"models":  []string{},
			"note":    "Local model — skipped remote test",
		})
		return
	}

	apiKey := m.APIKey()
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "No API key configured for this model",
		})
		return
	}

	if apiBase == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Cannot determine API base URL for this provider",
		})
		return
	}

	models, err := listProviderModels(apiBase, apiKey)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"models":  models,
	})
}

// ─── Video Model Routes ───────────────────────────────────────────────────────

// registerVideoModelRoutes binds video model CRUD endpoints.
func (h *Handler) registerVideoModelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/video-models", h.handleListVideoModels)
	mux.HandleFunc("POST /api/video-models", h.handleAddVideoModel)
	mux.HandleFunc("POST /api/video-models/default", h.handleSetDefaultVideoModel)
	mux.HandleFunc("POST /api/video-models/{index}/test", h.handleTestVideoModelKey)
	mux.HandleFunc("POST /api/video-models/{index}/rotate-key", h.handleRotateVideoModelKey)
	mux.HandleFunc("PUT /api/video-models/{index}", h.handleUpdateVideoModel)
	mux.HandleFunc("DELETE /api/video-models/{index}", h.handleDeleteVideoModel)
}

// handleListVideoModels returns all video_model_list entries with masked API keys.
//
//	GET /api/video-models
func (h *Handler) handleListVideoModels(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	configured := make([]bool, len(cfg.VideoModelList))

	var wg sync.WaitGroup
	wg.Add(len(cfg.VideoModelList))
	for i, m := range cfg.VideoModelList {
		go func(i int, m *config.ModelConfig) {
			defer wg.Done()
			configured[i] = isModelConfigured(m)
		}(i, m)
	}
	wg.Wait()

	configuredOnly := r.URL.Query().Get("configured_only") == "true"
	models := make([]modelResponse, 0, len(cfg.VideoModelList))
	for i, m := range cfg.VideoModelList {
		if configuredOnly && !configured[i] {
			continue
		}
		models = append(models, modelResponse{
			Index:          i,
			ModelName:      m.ModelName,
			Model:          m.Model,
			APIBase:        m.APIBase,
			APIKey:         maskAPIKey(m.APIKey()),
			Proxy:          m.Proxy,
			AuthMethod:     m.AuthMethod,
			ConnectMode:    m.ConnectMode,
			Workspace:      m.Workspace,
			RPM:            m.RPM,
			MaxTokensField: m.MaxTokensField,
			RequestTimeout: m.RequestTimeout,
			ThinkingLevel:  m.ThinkingLevel,
			ExtraBody:      m.ExtraBody,
			Configured:     configured[i],
			IsDefault:      false,
			IsVirtual:      m.IsVirtual(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"models": models,
		"total":  len(models),
	})
}

// handleRotateVideoModelKey replaces the API key for a video model at the given index.
//
//	POST /api/video-models/{index}/rotate-key
func (h *Handler) handleRotateVideoModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		NewAPIKey string `json:"new_api_key"`
	}
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if req.NewAPIKey == "" {
		http.Error(w, "new_api_key is required", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.VideoModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.VideoModelList)-1), http.StatusNotFound)
		return
	}

	cfg.VideoModelList[idx].SetAPIKey(req.NewAPIKey)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleAddVideoModel appends a new video model configuration entry.
//
//	POST /api/video-models
func (h *Handler) handleAddVideoModel(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	if mc.APIKey != "" {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	cfg.VideoModelList = append(cfg.VideoModelList, &mc.ModelConfig)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"index":  len(cfg.VideoModelList) - 1,
	})
}

// handleUpdateVideoModel replaces a video model configuration entry at the given index.
//
//	PUT /api/video-models/{index}
func (h *Handler) handleUpdateVideoModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	type custom struct {
		config.ModelConfig
		APIKey string `json:"api_key"`
	}

	var mc custom
	if err = json.Unmarshal(body, &mc); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err = mc.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.VideoModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.VideoModelList)-1), http.StatusNotFound)
		return
	}

	if mc.APIKey == "" {
		mc.ModelConfig.SetAPIKey(cfg.VideoModelList[idx].APIKey())
	} else {
		mc.ModelConfig.SetAPIKey(mc.APIKey)
	}
	if mc.ExtraBody == nil {
		mc.ExtraBody = cfg.VideoModelList[idx].ExtraBody
	} else if len(mc.ExtraBody) == 0 {
		mc.ExtraBody = nil
	}

	cfg.VideoModelList[idx] = &mc.ModelConfig

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleDeleteVideoModel removes a video model configuration entry at the given index.
//
//	DELETE /api/video-models/{index}
func (h *Handler) handleDeleteVideoModel(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.VideoModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range (0-%d)", idx, len(cfg.VideoModelList)-1), http.StatusNotFound)
		return
	}

	cfg.VideoModelList = append(cfg.VideoModelList[:idx], cfg.VideoModelList[idx+1:]...)

	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleSetDefaultVideoModel acknowledges a set-default request for video models.
//
//	POST /api/video-models/default
func (h *Handler) handleSetDefaultVideoModel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleTestVideoModelKey verifies the API key for a video model at the given index.
//
//	POST /api/video-models/{index}/test
func (h *Handler) handleTestVideoModelKey(w http.ResponseWriter, r *http.Request) {
	idx, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	if idx < 0 || idx >= len(cfg.VideoModelList) {
		http.Error(w, fmt.Sprintf("Index %d out of range", idx), http.StatusNotFound)
		return
	}

	m := cfg.VideoModelList[idx]

	w.Header().Set("Content-Type", "application/json")

	if strings.EqualFold(strings.TrimSpace(m.AuthMethod), "oauth") {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"note":    "OAuth model — authentication managed via credentials page",
		})
		return
	}

	apiBase := resolveProviderAPIBase(m)

	if apiBase != "" && hasLocalAPIBase(apiBase) {
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"models":  []string{},
			"note":    "Local model — skipped remote test",
		})
		return
	}

	apiKey := m.APIKey()
	if apiKey == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "No API key configured for this model",
		})
		return
	}

	if apiBase == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   "Cannot determine API base URL for this provider",
		})
		return
	}

	models, err := listProviderModels(apiBase, apiKey)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"models":  models,
	})
}
