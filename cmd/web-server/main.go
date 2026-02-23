package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
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

var (
	cfg       *config.Config
	agentLoop *agent.AgentLoop
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
