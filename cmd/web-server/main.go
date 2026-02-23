package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/skills"
	"github.com/sipeed/picoclaw/pkg/tools"
)

type ChatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"`
	Stream  bool   `json:"stream"`
}

type ChatResponse struct {
	Message   string    `json:"message"`
	Model     string    `json:"model"`
	Timestamp time.Time `json:"timestamp"`
}

type ConfigResponse struct {
	ModelList []ModelConfig `json:"model_list"`
	Agents    struct {
		Defaults struct {
			Model string `json:"model"`
		} `json:"defaults"`
	} `json:"agents"`
}

type ModelConfig struct {
	ModelName string `json:"model_name"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key,omitempty"`
	APIBase   string `json:"api_base,omitempty"`
}

type SkillInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type SkillDetail struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Source      string            `json:"source"`
	Description string            `json:"description"`
	Content     string            `json:"content"`
	Metadata    map[string]string `json:"metadata"`
}

type SearchSkillsRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type InstallSkillRequest struct {
	Slug     string `json:"slug"`
	Registry string `json:"registry"`
	Version  string `json:"version,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

var (
	cfg          *config.Config
	agentLoop    *agent.AgentLoop
	skillsLoader *skills.SkillsLoader
)

func loadConfig() error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".picoclaw", "config.json")

	var err error
	cfg, err = config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create provider and agent loop
	provider, modelID, err := providers.CreateProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Use the resolved model ID from provider creation
	if modelID != "" {
		cfg.Agents.Defaults.Model = modelID
	}

	msgBus := bus.NewMessageBus()
	agentLoop = agent.NewAgentLoop(cfg, msgBus, provider)

	// Initialize skills loader - 使用与主程序相同的工作空间配置
	var skillsWorkspace string
	if cfg != nil && cfg.Agents.Defaults.Workspace != "" {
		skillsWorkspace = cfg.WorkspacePath()
		log.Printf("Using configured workspace: %s", skillsWorkspace)
	} else {
		// 回退到默认工作空间
		skillsWorkspace = filepath.Join(home, ".picoclaw", "workspace")
		log.Printf("Using default workspace: %s", skillsWorkspace)
	}

	globalSkills := filepath.Join(home, ".picoclaw", "skills")
	builtinSkills := "./skills"
	skillsLoader = skills.NewSkillsLoader(skillsWorkspace, globalSkills, builtinSkills)

	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := ConfigResponse{
		ModelList: make([]ModelConfig, 0),
	}

	if cfg != nil && cfg.ModelList != nil {
		for _, model := range cfg.ModelList {
			modelConfig := ModelConfig{
				ModelName: model.ModelName,
				Model:     model.Model,
				// 不返回 API Key 到前端
			}
			if model.APIBase != "" {
				modelConfig.APIBase = model.APIBase
			}
			response.ModelList = append(response.ModelList, modelConfig)
		}
	}

	if cfg != nil {
		response.Agents.Defaults.Model = cfg.Agents.Defaults.Model
	}

	json.NewEncoder(w).Encode(response)
}

func updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 更新配置
	if cfg == nil {
		cfg = &config.Config{}
	}

	cfg.ModelList = make([]config.ModelConfig, 0)
	for _, model := range req.ModelList {
		modelConfig := config.ModelConfig{
			ModelName: model.ModelName,
			Model:     model.Model,
			APIKey:    model.APIKey,
		}
		if model.APIBase != "" {
			modelConfig.APIBase = model.APIBase
		}
		cfg.ModelList = append(cfg.ModelList, modelConfig)
	}

	// 更新默认模型
	if len(req.Agents.Defaults.Model) > 0 {
		cfg.Agents.Defaults.Model = req.Agents.Defaults.Model
	}

	// 保存配置
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".picoclaw", "config.json")

	configData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal config: %v", err), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save config: %v", err), http.StatusInternalServerError)
		return
	}

	// 重新加载配置
	if err := loadConfig(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload config: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if agentLoop == nil {
		http.Error(w, "Agent not initialized", http.StatusInternalServerError)
		return
	}

	// Update the agent loop's model temporarily
	originalModel := cfg.Agents.Defaults.Model
	cfg.Agents.Defaults.Model = req.Model
	defer func() {
		cfg.Agents.Defaults.Model = originalModel
	}()

	// Process the message
	ctx := context.Background()
	sessionKey := fmt.Sprintf("web:%d", time.Now().Unix())
	response, err := agentLoop.ProcessDirect(ctx, req.Message, sessionKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Chat error: %v", err), http.StatusInternalServerError)
		return
	}

	chatResponse := ChatResponse{
		Message:   response,
		Model:     req.Model,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse)
}

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	models := make([]ModelConfig, 0)

	if cfg != nil && cfg.ModelList != nil {
		for _, model := range cfg.ModelList {
			modelConfig := ModelConfig{
				ModelName: model.ModelName,
				Model:     model.Model,
			}
			if model.APIBase != "" {
				modelConfig.APIBase = model.APIBase
			}
			models = append(models, modelConfig)
		}
	}

	json.NewEncoder(w).Encode(models)
}

// 技能相关的处理函数
func skillsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if skillsLoader == nil {
		http.Error(w, "Skills loader not initialized", http.StatusInternalServerError)
		return
	}

	skillsList := skillsLoader.ListSkills()
	result := make([]SkillInfo, 0, len(skillsList))

	for _, skill := range skillsList {
		result = append(result, SkillInfo{
			Name:        skill.Name,
			Path:        skill.Path,
			Source:      skill.Source,
			Description: skill.Description,
		})
	}

	json.NewEncoder(w).Encode(result)
}

func skillDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	skillName := vars["name"]

	if skillsLoader == nil {
		http.Error(w, "Skills loader not initialized", http.StatusInternalServerError)
		return
	}

	skillContent, exists := skillsLoader.LoadSkill(skillName)
	if !exists {
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	// 获取技能信息
	skillsList := skillsLoader.ListSkills()
	var skillInfo *skills.SkillInfo
	for _, skill := range skillsList {
		if skill.Name == skillName {
			skillInfo = &skill
			break
		}
	}

	if skillInfo == nil {
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	// 解析元数据
	metadata := make(map[string]string)
	if skillContent != "" {
		lines := strings.Split(skillContent, "\n")
		inFrontmatter := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "---" {
				if !inFrontmatter {
					inFrontmatter = true
					continue
				} else {
					break
				}
			}
			if inFrontmatter && strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					metadata[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	detail := SkillDetail{
		Name:        skillInfo.Name,
		Path:        skillInfo.Path,
		Source:      skillInfo.Source,
		Description: skillInfo.Description,
		Content:     skillContent,
		Metadata:    metadata,
	}

	json.NewEncoder(w).Encode(detail)
}

func searchSkillsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SearchSkillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if cfg == nil || cfg.Tools.Skills.Registries.ClawHub.Enabled == false {
		http.Error(w, "Skills registry not configured", http.StatusNotFound)
		return
	}

	// 使用工具搜索技能
	clawHubConfig := cfg.Tools.Skills.Registries.ClawHub
	registryConfig := skills.RegistryConfig{
		ClawHub: skills.ClawHubConfig{
			Enabled:         clawHubConfig.Enabled,
			BaseURL:         clawHubConfig.BaseURL,
			AuthToken:       clawHubConfig.AuthToken,
			SearchPath:      clawHubConfig.SearchPath,
			SkillsPath:      clawHubConfig.SkillsPath,
			DownloadPath:    clawHubConfig.DownloadPath,
			Timeout:         clawHubConfig.Timeout,
			MaxZipSize:      clawHubConfig.MaxZipSize,
			MaxResponseSize: clawHubConfig.MaxResponseSize,
		},
		MaxConcurrentSearches: cfg.Tools.Skills.MaxConcurrentSearches,
	}
	registryMgr := skills.NewRegistryManagerFromConfig(registryConfig)
	cache := skills.NewSearchCache(cfg.Tools.Skills.SearchCache.MaxSize, time.Duration(cfg.Tools.Skills.SearchCache.TTLSeconds)*time.Second)
	findSkillTool := tools.NewFindSkillsTool(registryMgr, cache)

	if req.Limit == 0 {
		req.Limit = 10
	}

	ctx := context.Background()
	result := findSkillTool.Execute(ctx, map[string]interface{}{
		"query": req.Query,
		"limit": req.Limit,
	})

	// 解析搜索结果
	var searchResults []interface{}
	if !result.IsError {
		searchResults = []interface{}{result.ForLLM}
	} else {
		searchResults = []interface{}{}
	}

	response := map[string]interface{}{
		"query":   req.Query,
		"results": searchResults,
	}

	json.NewEncoder(w).Encode(response)
}

func installSkillHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InstallSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 使用工具安装技能
	clawHubConfig := cfg.Tools.Skills.Registries.ClawHub
	registryConfig := skills.RegistryConfig{
		ClawHub: skills.ClawHubConfig{
			Enabled:         clawHubConfig.Enabled,
			BaseURL:         clawHubConfig.BaseURL,
			AuthToken:       clawHubConfig.AuthToken,
			SearchPath:      clawHubConfig.SearchPath,
			SkillsPath:      clawHubConfig.SkillsPath,
			DownloadPath:    clawHubConfig.DownloadPath,
			Timeout:         clawHubConfig.Timeout,
			MaxZipSize:      clawHubConfig.MaxZipSize,
			MaxResponseSize: clawHubConfig.MaxResponseSize,
		},
		MaxConcurrentSearches: cfg.Tools.Skills.MaxConcurrentSearches,
	}
	registryMgr := skills.NewRegistryManagerFromConfig(registryConfig)
	installSkillTool := tools.NewInstallSkillTool(registryMgr, ".")

	installParams := map[string]interface{}{
		"slug":     req.Slug,
		"registry": req.Registry,
	}

	if req.Version != "" {
		installParams["version"] = req.Version
	}
	if req.Force {
		installParams["force"] = req.Force
	}

	ctx := context.Background()
	result := installSkillTool.Execute(ctx, installParams)

	var status string
	var message string
	if !result.IsError {
		status = "success"
		message = "Skill installed successfully"
	} else {
		status = "error"
		message = result.ForLLM
	}

	response := map[string]interface{}{
		"status":  status,
		"message": message,
		"result":  result.ForLLM,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		// 继续运行，但功能可能受限
	}

	r := mux.NewRouter()

	// API 路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/config", getConfigHandler).Methods("GET")
	api.HandleFunc("/config", updateConfigHandler).Methods("PUT")
	api.HandleFunc("/chat", chatHandler).Methods("POST")
	api.HandleFunc("/models", modelsHandler).Methods("GET")

	// 技能相关路由
	api.HandleFunc("/skills", skillsHandler).Methods("GET")
	api.HandleFunc("/skills/search", searchSkillsHandler).Methods("POST")
	api.HandleFunc("/skills/install", installSkillHandler).Methods("POST")
	api.HandleFunc("/skills/{name}", skillDetailHandler).Methods("GET")

	// 静态文件服务（用于前端构建文件）
	// 首先尝试服务前端文件，如果不存在则回退到 API
	distPath := "./web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Printf("Warning: Frontend build directory not found at %s", distPath)
	}

	fs := http.FileServer(http.Dir(distPath))
	r.PathPrefix("/").Handler(fs)

	// 应用 CORS 中间件
	handler := corsMiddleware(r)

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	fmt.Printf("PicoClaw Web Server starting on port %s\n", port)
	fmt.Printf("Frontend: http://localhost:%s\n", port)
	fmt.Printf("API: http://localhost:%s/api\n", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
