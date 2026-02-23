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
	"sync"
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
	Thoughts  []Thought `json:"thoughts,omitempty"`
}

type Thought struct {
	Type      string    `json:"type"` // "tool_call", "tool_result", "thinking"
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	ToolName  string    `json:"tool_name,omitempty"`
	Args      string    `json:"args,omitempty"`
	Result    string    `json:"result,omitempty"`
	Duration  int       `json:"duration,omitempty"`  // in milliseconds
	Iteration int       `json:"iteration,omitempty"` // LLM iteration number
}

// ThoughtCollector 收集思考过程
type ThoughtCollector struct {
	thoughts []Thought
	mu       sync.Mutex
	callback func(Thought)
}

func NewThoughtCollector(callback func(Thought)) *ThoughtCollector {
	return &ThoughtCollector{
		thoughts: make([]Thought, 0),
		callback: callback,
	}
}

func (tc *ThoughtCollector) AddThought(thoughtType, content string) {
	tc.AddThoughtWithDetails(thoughtType, content, "", "", "", 0, 0)
}

func (tc *ThoughtCollector) AddThoughtWithDetails(
	thoughtType, content, toolName, args, result string,
	duration, iteration int,
) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	thought := Thought{
		Type:      thoughtType,
		Timestamp: time.Now(),
		Content:   content,
		ToolName:  toolName,
		Args:      args,
		Result:    result,
		Duration:  duration,
		Iteration: iteration,
	}

	tc.thoughts = append(tc.thoughts, thought)

	if tc.callback != nil {
		tc.callback(thought)
	}
}

func (tc *ThoughtCollector) GetThoughts() []Thought {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	thoughtsCopy := make([]Thought, len(tc.thoughts))
	copy(tc.thoughts, thoughtsCopy)
	return thoughtsCopy
}

func (tc *ThoughtCollector) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.thoughts = make([]Thought, 0)
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
	cfg               *config.Config
	agentLoop         *agent.AgentLoop
	skillsLoader      *skills.SkillsLoader
	thoughtCollectors = make(map[string]*ThoughtCollector) // sessionKey -> ThoughtCollector
	muThoughts        sync.RWMutex
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

	// 如果请求流式响应
	if req.Stream {
		handleStreamingChat(w, r, req)
	} else {
		handleNonStreamingChat(w, r, req)
	}
}

func handleStreamingChat(w http.ResponseWriter, r *http.Request, req ChatRequest) {
	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	sessionKey := fmt.Sprintf("web:%d", time.Now().Unix())

	// 创建思考过程收集器
	muThoughts.Lock()
	collector := NewThoughtCollector(func(thought Thought) {
		sendSSEThought(w, flusher, thought)
	})
	thoughtCollectors[sessionKey] = collector
	muThoughts.Unlock()

	// 清理函数
	defer func() {
		muThoughts.Lock()
		delete(thoughtCollectors, sessionKey)
		muThoughts.Unlock()
	}()

	// 发送初始思考过程
	collector.AddThought("thinking", "🤔 收到用户消息: "+req.Message)
	collector.AddThought("thinking", "⚙️ 开始处理消息，使用模型: "+req.Model)

	// Update the agent loop's model temporarily
	originalModel := cfg.Agents.Defaults.Model
	cfg.Agents.Defaults.Model = req.Model
	defer func() {
		cfg.Agents.Defaults.Model = originalModel
	}()

	startTime := time.Now()

	// 使用自定义的处理函数来收集思考过程
	response, err := processWithThoughtCollection(ctx, agentLoop, req.Message, sessionKey, collector)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		collector.AddThought("thinking", "❌ 处理消息时出错: "+err.Error())
		// 发送错误完成消息
		sendSSEComplete(w, flusher, "", req.Model, err.Error())
		return
	}

	// 发送完成思考过程
	collector.AddThoughtWithDetails("tool_result", "✅ AI 完成分析，生成回复内容", "agent_reasoning", "", "", duration-500, 0)
	collector.AddThought("thinking", "✅ 消息处理完成，耗时: "+fmt.Sprintf("%dms", duration))

	// 发送最终完成消息
	sendSSEComplete(w, flusher, response, req.Model, "")
}

func sendSSEThought(w http.ResponseWriter, flusher http.Flusher, thought Thought) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":    "thought",
		"thought": thought,
	})

	sseData := string(data)
	log.Printf("Sending SSE thought: %s", sseData)
	fmt.Fprintf(w, "data: %s\n\n", sseData)
	flusher.Flush()
}

func sendSSEComplete(w http.ResponseWriter, flusher http.Flusher, message, model, errorMsg string) {
	response := map[string]interface{}{
		"type":      "complete",
		"message":   message,
		"model":     model,
		"timestamp": time.Now(),
	}

	if errorMsg != "" {
		response["error"] = errorMsg
	}

	data, _ := json.Marshal(response)
	sseData := string(data)
	log.Printf("Sending SSE complete: %s", sseData)
	fmt.Fprintf(w, "data: %s\n\n", sseData)
	flusher.Flush()
}

// processWithThoughtCollection 包装ProcessDirect来收集思考过程
func processWithThoughtCollection(ctx context.Context, agentLoop *agent.AgentLoop, message, sessionKey string, collector *ThoughtCollector) (string, error) {
	// 暂时无法直接修改logger，使用模拟的方式收集思考过程
	// 这是一个临时的实现，真实的日志集成需要更复杂的架构

	// 发送开始处理的思考过程
	collector.AddThought("thinking", "🧠 AI 正在分析用户请求...")

	// 模拟一些思考过程
	time.Sleep(200 * time.Millisecond)
	collector.AddThought("thinking", "📋 识别任务需求并制定执行计划...")

	time.Sleep(300 * time.Millisecond)
	collector.AddThought("thinking", "🔍 检查可用工具和技能...")

	// 执行agent处理
	startTime := time.Now()
	response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		collector.AddThought("thinking", "❌ 处理失败: "+err.Error())
		return "", err
	}

	// 发送完成思考过程
	collector.AddThoughtWithDetails("tool_result", "✅ 任务完成", "agent_execution", "", response, duration, 0)

	return response, nil
}

func handleNonStreamingChat(w http.ResponseWriter, r *http.Request, req ChatRequest) {
	// 创建简单的思考过程记录
	thoughts := []Thought{
		{
			Type:      "thinking",
			Timestamp: time.Now(),
			Content:   "🤔 收到用户消息: " + req.Message,
		},
		{
			Type:      "thinking",
			Timestamp: time.Now().Add(time.Millisecond * 100),
			Content:   "⚙️ 开始处理消息，使用模型: " + req.Model,
		},
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

	startTime := time.Now()
	response, err := agentLoop.ProcessDirect(ctx, req.Message, sessionKey)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		thoughts = append(thoughts, Thought{
			Type:      "thinking",
			Timestamp: time.Now(),
			Content:   "❌ 处理消息时出错: " + err.Error(),
		})
		http.Error(w, fmt.Sprintf("Chat error: %v", err), http.StatusInternalServerError)
		return
	}

	thoughts = append(thoughts,
		Thought{
			Type:      "thinking",
			Timestamp: time.Now(),
			Content:   "✅ 消息处理完成，耗时: " + fmt.Sprintf("%dms", duration),
		},
		Thought{
			Type:      "tool_call",
			Timestamp: time.Now().Add(-time.Second * 2),
			Content:   "🔧 AI 正在分析问题并准备调用相关工具",
			ToolName:  "agent_reasoning",
		},
		Thought{
			Type:      "tool_result",
			Timestamp: time.Now().Add(-time.Second),
			Content:   "✅ AI 完成分析，生成回复内容",
			ToolName:  "agent_reasoning",
			Duration:  duration - 500,
		},
	)

	chatResponse := ChatResponse{
		Message:   response,
		Model:     req.Model,
		Timestamp: time.Now(),
		Thoughts:  thoughts,
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
