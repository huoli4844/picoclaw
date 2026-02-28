package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/channels"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/mcp"
	"github.com/sipeed/picoclaw/pkg/media"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/skills"

	// 导入重构后的包
	"github.com/sipeed/picoclaw/cmd/web-server/handlers"
	"github.com/sipeed/picoclaw/cmd/web-server/services"
)

var (
	cfg             *config.Config
	agentLoop       *agent.AgentLoop
	skillsLoader    *skills.SkillsLoader
	skillsWorkspace string
	mcpRegistry     *mcp.Registry
	conversationSvc *services.ConversationService
	channelsManager *channels.Manager
)

// 全局日志处理器，用于捕获日志并发送到适配器
func init() {
	// 设置日志级别为DEBUG以捕获所有日志
	logger.SetLevel(logger.DEBUG)
}

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
	if cfg != nil && cfg.Agents.Defaults.Workspace != "" {
		skillsWorkspace = cfg.WorkspacePath()
		log.Printf("Using configured workspace: %s", skillsWorkspace)
	} else {
		// 回退到默认工作空间
		home, _ := os.UserHomeDir()
		skillsWorkspace = filepath.Join(home, ".picoclaw", "workspace")
		log.Printf("Using default workspace: %s", skillsWorkspace)
	}

	globalSkills := filepath.Join(home, ".picoclaw", "skills")
	builtinSkills := "./skills"
	skillsLoader = skills.NewSkillsLoader(skillsWorkspace, globalSkills, builtinSkills)

	return nil
}

func initServices() error {
	var err error

	// 初始化对话服务
	conversationSvc, err = services.NewConversationService()
	if err != nil {
		return fmt.Errorf("failed to initialize conversation service: %v", err)
	}

	// 初始化MCP注册表
	home, _ := os.UserHomeDir()
	mcpStoragePath := filepath.Join(home, ".picoclaw", "mcp")
	mcpRegistry, err = mcp.NewRegistry("", mcpStoragePath)
	if err != nil {
		log.Printf("Warning: Failed to initialize MCP registry: %v", err)
	}

	// 初始化Channel Manager
	if cfg != nil {
		msgBus := bus.NewMessageBus()
		mediaStore := media.NewFileMediaStore()
		channelsManager, err = channels.NewManager(cfg, msgBus, mediaStore)
		if err != nil {
			log.Printf("Warning: Failed to initialize channels manager: %v", err)
		} else {
			// 启动Channel Manager
			ctx := context.Background()
			if err := channelsManager.StartAll(ctx); err != nil {
				log.Printf("Warning: Failed to start channels manager: %v", err)
			} else {
				log.Println("✅ Channels manager started successfully")
			}
		}
	}

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

func setupRoutes() http.Handler {
	r := mux.NewRouter()

	// 创建处理器
	configHandler := handlers.NewConfigHandler(cfg)
	chatHandler := handlers.NewChatHandler(agentLoop, cfg, conversationSvc)
	conversationHandler := handlers.NewConversationHandler(conversationSvc)
	fileHandler := handlers.NewFileHandler()
	skillHandler := handlers.NewSkillHandler(skillsLoader, skillsWorkspace, cfg)
	mcpHandler := handlers.NewMCPHandler(mcpRegistry)

	// API 路由
	api := r.PathPrefix("/api").Subrouter()

	// 配置和聊天相关路由
	api.HandleFunc("/config", configHandler.GetConfig).Methods("GET")
	api.HandleFunc("/config", configHandler.UpdateConfig).Methods("PUT")
	api.HandleFunc("/chat", chatHandler.HandleChat).Methods("POST")
	api.HandleFunc("/models", configHandler.GetModels).Methods("GET")

	// 对话历史相关路由
	api.HandleFunc("/conversations", conversationHandler.GetConversations).Methods("GET")
	api.HandleFunc("/conversations", conversationHandler.CreateConversation).Methods("POST")
	api.HandleFunc("/conversations/{id}", conversationHandler.GetConversation).Methods("GET")
	api.HandleFunc("/conversations/{id}", conversationHandler.UpdateConversation).Methods("PUT")
	api.HandleFunc("/conversations/{id}", conversationHandler.DeleteConversation).Methods("DELETE")

	// 文件浏览器相关路由
	api.HandleFunc("/files", fileHandler.ListFiles).Methods("GET")
	api.HandleFunc("/file-content", fileHandler.GetFileContent).Methods("GET")
	api.HandleFunc("/file-delete", fileHandler.DeleteFile).Methods("DELETE")

	// 技能相关路由
	api.HandleFunc("/skills", skillHandler.GetSkills).Methods("GET")
	api.HandleFunc("/skills/search", skillHandler.SearchSkills).Methods("POST")
	api.HandleFunc("/skills/install", skillHandler.InstallSkill).Methods("POST")
	api.HandleFunc("/skills/{name}", skillHandler.GetSkillDetail).Methods("GET")
	api.HandleFunc("/skills/{name}", skillHandler.UninstallSkill).Methods("DELETE")

	// MCP相关路由
	api.HandleFunc("/mcp/servers", mcpHandler.GetServers).Methods("GET")
	api.HandleFunc("/mcp/servers/{id:.+}", mcpHandler.GetServerDetail).Methods("GET")
	api.HandleFunc("/mcp/servers/{id:.+}", mcpHandler.UninstallServer).Methods("DELETE")
	api.HandleFunc("/mcp/servers/{id:.+}/validate", mcpHandler.ValidateServer).Methods("POST")
	api.HandleFunc("/mcp/servers/{id:.+}/call", mcpHandler.CallTool).Methods("POST")
	api.HandleFunc("/mcp/search", mcpHandler.SearchServers).Methods("POST")
	api.HandleFunc("/mcp/sources", mcpHandler.GetSources).Methods("GET")
	api.HandleFunc("/mcp/install", mcpHandler.InstallServer).Methods("POST")

	// 静态文件服务（用于前端构建文件）
	// 首先尝试服务前端文件，如果不存在则回退到 API
	distPath := "./web/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Printf("Warning: Frontend build directory not found at %s", distPath)
	}

	fs := http.FileServer(http.Dir(distPath))
	r.PathPrefix("/").Handler(fs)

	return r
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		// 继续运行，但功能可能受限
	}

	// 初始化服务
	if err := initServices(); err != nil {
		log.Printf("Warning: Failed to initialize services: %v", err)
		// 继续运行，但功能可能受限
	}

	// 设置路由
	router := setupRoutes()

	// 应用 CORS 中间件
	handler := corsMiddleware(router)

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	fmt.Printf("PicoClaw Web Server starting on port %s\n", port)
	fmt.Printf("Frontend: http://localhost:%s\n", port)
	fmt.Printf("API: http://localhost:%s/api\n", port)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
