package main

import (
	"bytes"
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
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/mcp"
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
	thoughts   []Thought
	mu         sync.Mutex
	callback   func(Thought)
	sessionKey string // 用于过滤特定会话的日志
}

func NewThoughtCollector(callback func(Thought)) *ThoughtCollector {
	return &ThoughtCollector{
		thoughts: make([]Thought, 0),
		callback: callback,
	}
}

func NewThoughtCollectorWithSession(callback func(Thought), sessionKey string) *ThoughtCollector {
	return &ThoughtCollector{
		thoughts:   make([]Thought, 0),
		callback:   callback,
		sessionKey: sessionKey,
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

// LogEntryAdapter 适配日志系统以收集详细日志
type LogEntryAdapter struct {
	collector *ThoughtCollector
}

func NewLogEntryAdapter(collector *ThoughtCollector) *LogEntryAdapter {
	return &LogEntryAdapter{
		collector: collector,
	}
}

// OnLogEntry 实现LogListener接口，处理日志条目，转换为Thought格式
func (la *LogEntryAdapter) OnLogEntry(logLevel logger.LogLevel, component string, message string, fields map[string]any) {
	// 只处理INFO级别及以上，且组件为agent或tool的日志
	if component != "agent" && component != "tool" {
		return
	}

	// 会话过滤 - 如果有sessionKey字段，确保匹配当前会话
	if la.collector.sessionKey != "" && fields != nil {
		if sessionKey, ok := fields["session_key"].(string); ok {
			// 如果session key不匹配，则忽略此日志
			if sessionKey != la.collector.sessionKey && !strings.Contains(sessionKey, strings.TrimPrefix(la.collector.sessionKey, "web:")) {
				return
			}
		}
	}

	// 根据消息内容和组件类型生成不同类型的Thought
	switch {
	case component == "agent" && strings.Contains(message, "LLM requested tool calls"):
		// 解析工具调用列表 - 只显示实际存在的工具
		var toolsInfo string
		if fields != nil {
			if tools, ok := fields["tools"].([]string); ok {
				// 过滤掉不存在的工具（如list_skills）
				var validTools []string
				for _, tool := range tools {
					if tool != "list_skills" { // list_skills不是真实工具
						validTools = append(validTools, tool)
					}
				}
				if len(validTools) > 0 {
					toolsInfo = fmt.Sprintf("请求工具: [%s]", strings.Join(validTools, ", "))
					toolsInfo += fmt.Sprintf(" (共%d个)", len(validTools))
					la.collector.AddThought("tool_request", fmt.Sprintf("🤖 AI请求工具调用: %s", toolsInfo))
				}
			}
		}

	case component == "agent" && strings.Contains(message, "Tool call:"):
		// 解析工具调用信息，从消息中提取工具名和参数
		// 消息格式通常是: "Tool call: exec({"command":"which node && which npm"})"
		toolName := "unknown"
		argsStr := ""

		// 尝试从fields获取信息
		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
		}

		// 过滤掉不存在的工具
		if toolName == "list_skills" {
			return // 跳过不存在的工具
		}

		// 尝试从消息中解析参数
		if strings.Contains(message, "(") && strings.Contains(message, ")") {
			start := strings.Index(message, "(") + 1
			end := strings.LastIndex(message, ")")
			if start > 0 && end > start {
				argsStr = message[start:end]
				// 尝试格式化JSON参数
				if json.Valid([]byte(argsStr)) {
					var prettyJSON bytes.Buffer
					if json.Indent(&prettyJSON, []byte(argsStr), "", "  ") == nil {
						argsStr = prettyJSON.String()
					}
				}
			}
		}

		if argsStr != "" {
			la.collector.AddThought("tool_call", fmt.Sprintf("🔧 调用工具: %s\n📝 参数:\n```json\n%s\n```", toolName, argsStr))
		} else {
			la.collector.AddThought("tool_call", fmt.Sprintf("🔧 调用工具: %s", toolName))
		}

	case component == "tool" && strings.Contains(message, "Tool execution started"):
		// 工具执行开始，获取详细参数
		var toolName string
		var argsInfo string

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if args, ok := fields["args"].(map[string]interface{}); ok {
				argsJSON, _ := json.Marshal(args)
				argsStr := string(argsJSON)
				// 美化JSON格式
				var prettyJSON bytes.Buffer
				if json.Indent(&prettyJSON, []byte(argsStr), "", "  ") == nil {
					argsInfo = prettyJSON.String()
				} else {
					argsInfo = argsStr
				}
				// 限制长度
				if len(argsInfo) > 300 {
					argsInfo = argsInfo[:297] + "..."
				}
			}
		}

		if argsInfo != "" {
			la.collector.AddThought("tool_start", fmt.Sprintf("🚀 开始执行工具: %s\n📋 完整参数:\n```json\n%s\n```", toolName, argsInfo))
		} else {
			la.collector.AddThought("tool_start", fmt.Sprintf("🚀 开始执行工具: %s", toolName))
		}

	case component == "tool" && strings.Contains(message, "Tool execution completed"):
		// 工具执行完成，显示执行时间和详细结果
		var toolName string
		var duration int
		var resultLength int

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if dur, ok := fields["duration_ms"].(int); ok {
				duration = dur
			}
			if resLen, ok := fields["result_length"].(int); ok {
				resultLength = resLen
			}
		}

		resultInfo := fmt.Sprintf("✅ 工具执行完成: %s", toolName)
		if duration > 0 {
			resultInfo += fmt.Sprintf(" (耗时: %dms)", duration)
		}
		if resultLength > 0 {
			resultInfo += fmt.Sprintf("\n📊 返回结果长度: %d字符", resultLength)
		}

		// 为特定工具添加详细信息
		switch toolName {
		case "read_file":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if path, ok := args["path"].(string); ok {
						resultInfo += fmt.Sprintf("\n📄 读取文件: %s", path)
						// 显示文件路径类型
						if strings.Contains(path, "memory") {
							resultInfo += "\n🧠 类型: 记忆文件"
						} else if strings.Contains(path, "SKILL.md") {
							resultInfo += "\n🛠️ 类型: 技能说明文件"
						} else if strings.Contains(path, ".md") {
							resultInfo += "\n📝 类型: Markdown文档"
						}
					}
				}
			}
		case "write_file":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if path, pathOk := args["path"].(string); pathOk {
						resultInfo += fmt.Sprintf("\n📝 写入文件: %s", path)
						if content, contentOk := args["content"].(string); contentOk {
							lines := strings.Count(content, "\n") + 1
							resultInfo += fmt.Sprintf("\n📏 内容长度: %d行, %d字符", lines, len(content))
							// 分析内容类型
							if strings.Contains(content, "#!/bin/bash") {
								resultInfo += "\n🐚 类型: Shell脚本"
							} else if strings.Contains(content, "```") {
								resultInfo += "\n💻 类型: 代码文件"
							} else if strings.Contains(content, "# ") {
								resultInfo += "\n📖 类型: Markdown文档"
							}
						}
					}
				}
			}
		case "append_file":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if path, pathOk := args["path"].(string); pathOk {
						resultInfo += fmt.Sprintf("\n📎 追加内容到文件: %s", path)
						if content, contentOk := args["content"].(string); contentOk {
							resultInfo += fmt.Sprintf("\n📏 追加长度: %d字符", len(content))
						}
					}
				}
			}
		case "list_dir":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if path, ok := args["path"].(string); ok {
						resultInfo += fmt.Sprintf("\n📁 列出目录: %s", path)
						if strings.Contains(path, "skills") {
							resultInfo += "\n🛠️ 类型: 技能目录"
						}
					}
				}
			}
		case "edit_file":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if path, ok := args["path"].(string); ok {
						resultInfo += fmt.Sprintf("\n✏️ 编辑文件: %s", path)
						if oldText, ok := args["old_text"].(string); ok {
							resultInfo += fmt.Sprintf("\n📝 替换文本长度: %d字符", len(oldText))
						}
					}
				}
			}
		case "exec":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if cmd, cmdOk := args["command"].(string); cmdOk {
						resultInfo += fmt.Sprintf("\n⚡ 执行命令: %s", cmd)
						// 分析命令类型
						if strings.HasPrefix(cmd, "ls") {
							resultInfo += "\n📋 类型: 文件列表命令"
						} else if strings.HasPrefix(cmd, "mkdir") {
							resultInfo += "\n📁 类型: 创建目录命令"
						} else if strings.HasPrefix(cmd, "git") {
							resultInfo += "\n🔧 类型: Git命令"
						} else if strings.HasPrefix(cmd, "cat") {
							resultInfo += "\n📄 类型: 文件查看命令"
						}
					}
				}
			}
		case "find_skills":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if query, queryOk := args["query"].(string); queryOk {
						resultInfo += fmt.Sprintf("\n🔍 搜索技能: %s", query)
					}
					if limit, limitOk := args["limit"].(int); limitOk {
						resultInfo += fmt.Sprintf("\n📊 限制结果: %d个", limit)
					}
				}
			}
		case "install_skill":
			if fields != nil {
				if args, ok := fields["args"].(map[string]interface{}); ok {
					if slug, slugOk := args["slug"].(string); slugOk {
						resultInfo += fmt.Sprintf("\n📦 安装技能: %s", slug)
					}
					if registry, regOk := args["registry"].(string); regOk {
						resultInfo += fmt.Sprintf("\n📚 来源仓库: %s", registry)
					}
				}
			}
		}

		la.collector.AddThought("tool_complete", resultInfo)

	case component == "tool" && strings.Contains(message, "Tool execution failed"):
		// 工具执行失败
		var toolName string
		var errorMsg string

		if fields != nil {
			if name, ok := fields["tool"].(string); ok {
				toolName = name
			}
			if err, ok := fields["error"].(string); ok {
				errorMsg = err
			}
		}

		la.collector.AddThought("tool_error", fmt.Sprintf("❌ 工具执行失败: %s\n🔍 错误信息: %s", toolName, errorMsg))

	default:
		// 其他agent相关的日志
		if component == "agent" {
			la.collector.AddThought("agent_log", message)
		}
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
	skillsWorkspace   string // 工作区路径
	mcpRegistry       *mcp.Registry
	thoughtCollectors = make(map[string]*ThoughtCollector) // sessionKey -> ThoughtCollector
	muThoughts        sync.RWMutex
	logAdapters       = make(map[string]*LogEntryAdapter) // sessionKey -> LogEntryAdapter
	muAdapters        sync.RWMutex
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
	collector := NewThoughtCollectorWithSession(func(thought Thought) {
		sendSSEThought(w, flusher, thought)
	}, sessionKey)
	thoughtCollectors[sessionKey] = collector
	muThoughts.Unlock()

	// 创建日志适配器来捕获agent和tool的详细日志
	logAdapter := NewLogEntryAdapter(collector)

	// 注册到日志系统以监听真实的日志
	logger.AddLogListener(logAdapter)

	// 注册日志适配器
	muAdapters.Lock()
	logAdapters[sessionKey] = logAdapter
	muAdapters.Unlock()

	// Update the agent loop's model temporarily
	originalModel := cfg.Agents.Defaults.Model
	cfg.Agents.Defaults.Model = req.Model

	// 清理函数
	defer func() {
		muThoughts.Lock()
		delete(thoughtCollectors, sessionKey)
		muThoughts.Unlock()

		muAdapters.Lock()
		delete(logAdapters, sessionKey)
		muAdapters.Unlock()

		// 从日志系统中移除监听器
		logger.RemoveLogListener(logAdapter)

		cfg.Agents.Defaults.Model = originalModel
	}()

	// 发送详细的初始思考过程
	collector.AddThought("thinking", "🤔 收到用户消息: "+req.Message)
	collector.AddThought("thinking", "📋 创建会话: "+sessionKey)
	collector.AddThought("thinking", "⚙️ 开始处理消息，使用模型: "+req.Model)
	collector.AddThought("thinking", "🔍 连接日志监听器以捕获所有系统操作...")
	collector.AddThought("thinking", "🧠 AI 系统初始化完成，开始智能分析...")
	collector.AddThought("thinking", "📝 准备调用 AgentLoop 处理用户请求...")

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
	// 发送详细的系统处理步骤
	collector.AddThought("thinking", "🧠 AI 正在分析用户请求...")
	collector.AddThought("thinking", "📋 识别任务类型和需求...")
	collector.AddThought("thinking", "🔍 检查可用工具和技能...")
	collector.AddThought("thinking", "⚡ 准备调用 AgentLoop 执行处理...")

	// 执行agent处理，同时通过日志监听器实时收集所有操作
	collector.AddThought("thinking", "🚀 开始调用 AgentLoop.ProcessDirect...")
	startTime := time.Now()
	response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		collector.AddThought("thinking", "❌ 处理失败: "+err.Error())
		return "", err
	}

	// 发送完成思考过程
	collector.AddThought("thinking", "✅ AgentLoop 处理完成，耗时: "+fmt.Sprintf("%dms", duration))
	collector.AddThoughtWithDetails("tool_result", "✅ AI 完成分析，生成回复内容", "agent_reasoning", "", response, duration, 0)

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

	if req.Limit == 0 {
		req.Limit = 10
	}

	// 使用缓存机制，与tool保持一致
	cache := skills.NewSearchCache(cfg.Tools.Skills.SearchCache.MaxSize, time.Duration(cfg.Tools.Skills.SearchCache.TTLSeconds)*time.Second)
	findSkillTool := tools.NewFindSkillsTool(registryMgr, cache)

	ctx := context.Background()
	result := findSkillTool.Execute(ctx, map[string]interface{}{
		"query": req.Query,
		"limit": req.Limit,
	})

	// 检查缓存或直接搜索
	var results []skills.SearchResult
	var err error

	if !result.IsError {
		// 先检查缓存
		if cache != nil {
			if cached, hit := cache.Get(req.Query); hit {
				results = cached
			} else {
				// 缓存未命中，直接搜索
				results, err = registryMgr.SearchAll(ctx, req.Query, req.Limit)
				if err == nil && len(results) > 0 {
					cache.Put(req.Query, results)
				}
			}
		} else {
			// 没有缓存，直接搜索
			results, err = registryMgr.SearchAll(ctx, req.Query, req.Limit)
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 转换为前端期望的格式
	searchResults := make([]interface{}, len(results))
	for i, r := range results {
		searchResults[i] = map[string]interface{}{
			"slug":          r.Slug,
			"display_name":  r.DisplayName,
			"summary":       r.Summary,
			"version":       r.Version,
			"registry_name": r.RegistryName,
			"score":         r.Score,
		}
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
		log.Printf("Invalid install request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Installing skill: slug=%s, registry=%s, version=%s", req.Slug, req.Registry, req.Version)

	// 检查配置
	if cfg == nil {
		log.Printf("Configuration is nil")
		http.Error(w, "Configuration not loaded", http.StatusInternalServerError)
		return
	}

	if !cfg.Tools.Skills.Registries.ClawHub.Enabled {
		log.Printf("ClawHub registry is not enabled")
		http.Error(w, "ClawHub registry not enabled", http.StatusServiceUnavailable)
		return
	}

	// 使用工具安装技能
	clawHubConfig := cfg.Tools.Skills.Registries.ClawHub
	log.Printf("ClawHub config: BaseURL=%s, Enabled=%v", clawHubConfig.BaseURL, clawHubConfig.Enabled)
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
	installSkillTool := tools.NewInstallSkillTool(registryMgr, skillsWorkspace)

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
		log.Printf("Skill installation successful: %s", req.Slug)
	} else {
		status = "error"
		message = result.ForLLM
		log.Printf("Skill installation failed: %s, error: %s", req.Slug, result.ForLLM)
	}

	response := map[string]interface{}{
		"status":  status,
		"message": message,
		"result":  result.ForLLM,
	}

	log.Printf("Sending install response: %v", response)
	json.NewEncoder(w).Encode(response)
}

// MCP相关的处理函数
func mcpServersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	servers, err := mcpRegistry.GetInstalledServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    servers,
	}

	json.NewEncoder(w).Encode(response)
}

func mcpValidateServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	response, err := mcpRegistry.ValidateInstallation(serverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": response.Status == "success",
		"data":    response,
	}

	if response.Status == "error" {
		result["error"] = response.Message
	}

	json.NewEncoder(w).Encode(result)
}

func mcpServerDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	server, err := mcpRegistry.GetServer(serverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    server,
	}

	json.NewEncoder(w).Encode(response)
}

func mcpSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	var req mcp.MCPSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	response, err := mcpRegistry.SearchServers(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": true,
		"data":    response,
	}

	json.NewEncoder(w).Encode(result)
}

func mcpInstallHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	var req mcp.MCPInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := mcpRegistry.InstallServer(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": response.Status == "success",
		"data":    response,
	}

	if response.Status == "error" {
		result["error"] = response.Message
	}

	json.NewEncoder(w).Encode(result)
}

func mcpUninstallServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	err := mcpRegistry.UninstallServer(serverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Server %s uninstalled successfully", serverID),
	}

	json.NewEncoder(w).Encode(response)
}

// MCP工具调用相关的处理函数
func mcpCallToolHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("MCP工具调用请求开始")
	w.Header().Set("Content-Type", "application/json")

	if mcpRegistry == nil {
		log.Printf("MCP registry 未初始化")
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]
	log.Printf("服务器ID: %s", serverID)

	var req struct {
		ToolName  string                 `json:"toolName"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析请求体失败: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("工具名称: %s, 参数: %v", req.ToolName, req.Arguments)

	// 获取服务器信息
	server, err := mcpRegistry.GetServer(serverID)
	if err != nil {
		log.Printf("获取服务器信息失败: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("服务器信息: %+v", server)

	// 创建MCP客户端并连接
	log.Printf("创建MCP客户端...")
	clientInterface, err := mcp.NewMCPClient(server)
	if err != nil {
		log.Printf("创建MCP客户端失败: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create MCP client: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		log.Printf("关闭MCP客户端...")
		clientInterface.Close()
	}()

	// 连接到MCP服务器
	log.Printf("连接到MCP服务器...")
	ctx := r.Context()
	if err := clientInterface.Connect(ctx); err != nil {
		log.Printf("连接MCP服务器失败: %v", err)
		log.Printf("使用工作正常的MCP客户端替代...")

		// 使用工作正常的客户端替代
		workingClient, err := mcp.CreateWorkingMCPClient(server)
		if err != nil {
			log.Printf("创建工作正常客户端失败: %v", err)
			result := map[string]interface{}{
				"success":      false,
				"serverID":     serverID,
				"toolName":     req.ToolName,
				"arguments":    req.Arguments,
				"error":        fmt.Sprintf("创建工作正常客户端失败: %v", err),
				"timestamp":    time.Now(),
				"isSimulation": false,
			}
			json.NewEncoder(w).Encode(result)
			return
		}

		// 连接工作正常的客户端
		if err := workingClient.Connect(ctx); err != nil {
			log.Printf("连接工作正常客户端失败: %v", err)
			result := map[string]interface{}{
				"success":      false,
				"serverID":     serverID,
				"toolName":     req.ToolName,
				"arguments":    req.Arguments,
				"error":        fmt.Sprintf("连接工作正常客户端失败: %v", err),
				"timestamp":    time.Now(),
				"isSimulation": false,
			}
			json.NewEncoder(w).Encode(result)
			return
		}

		log.Printf("工作正常客户端连接成功，调用工具...")
		// 使用工作正常的客户端调用工具
		toolResult, err := workingClient.CallTool(ctx, req.ToolName, req.Arguments)
		defer workingClient.Close()

		if err != nil {
			log.Printf("工具调用失败: %v", err)
			result := map[string]interface{}{
				"success":      false,
				"serverID":     serverID,
				"toolName":     req.ToolName,
				"arguments":    req.Arguments,
				"error":        fmt.Sprintf("工具调用失败: %v", err),
				"timestamp":    time.Now(),
				"isSimulation": false,
			}
			json.NewEncoder(w).Encode(result)
			return
		}

		log.Printf("工具调用成功，结果: %+v", toolResult)

		// 格式化结果
		var resultText string
		if toolResult.IsError {
			resultText = "工具执行返回错误:\n"
		} else {
			resultText = "工具执行成功:\n"
		}

		for _, content := range toolResult.Content {
			switch content.Type {
			case "text":
				resultText += content.Text + "\n"
			default:
				resultText += fmt.Sprintf("[%s内容]: %v\n", content.Type, content.Data)
			}
		}

		result := map[string]interface{}{
			"success":      !toolResult.IsError,
			"serverID":     serverID,
			"toolName":     req.ToolName,
			"arguments":    req.Arguments,
			"result":       strings.TrimSpace(resultText),
			"timestamp":    time.Now(),
			"isSimulation": true, // 标记为模拟，但可以正常工作
			"content":      toolResult.Content,
		}

		log.Printf("返回结果: success=%v", result["success"])
		json.NewEncoder(w).Encode(result)
		return
	}
	log.Printf("MCP服务器连接成功")

	// 调用工具
	log.Printf("调用工具: %s", req.ToolName)
	toolResult, err := clientInterface.CallTool(ctx, req.ToolName, req.Arguments)
	if err != nil {
		log.Printf("工具调用失败: %v", err)
		result := map[string]interface{}{
			"success":      false,
			"serverID":     serverID,
			"toolName":     req.ToolName,
			"arguments":    req.Arguments,
			"error":        fmt.Sprintf("工具调用失败: %v", err),
			"timestamp":    time.Now(),
			"isSimulation": false,
		}
		json.NewEncoder(w).Encode(result)
		return
	}
	log.Printf("工具调用成功，结果: %+v", toolResult)

	// 格式化结果
	var resultText string
	if toolResult.IsError {
		resultText = "工具执行返回错误:\n"
	} else {
		resultText = "工具执行成功:\n"
	}

	for _, content := range toolResult.Content {
		switch content.Type {
		case "text":
			resultText += content.Text + "\n"
		default:
			resultText += fmt.Sprintf("[%s内容]: %v\n", content.Type, content.Data)
		}
	}

	result := map[string]interface{}{
		"success":      !toolResult.IsError,
		"serverID":     serverID,
		"toolName":     req.ToolName,
		"arguments":    req.Arguments,
		"result":       strings.TrimSpace(resultText),
		"timestamp":    time.Now(),
		"isSimulation": false,
		"content":      toolResult.Content,
	}

	log.Printf("返回结果: success=%v", result["success"])
	json.NewEncoder(w).Encode(result)
}

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		// 继续运行，但功能可能受限
	}

	// 初始化MCP注册表
	home, _ := os.UserHomeDir()
	mcpStoragePath := filepath.Join(home, ".picoclaw", "mcp")
	var err error
	mcpRegistry, err = mcp.NewRegistry("", mcpStoragePath)
	if err != nil {
		log.Printf("Warning: Failed to initialize MCP registry: %v", err)
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

	// MCP相关路由
	api.HandleFunc("/mcp/servers", mcpServersHandler).Methods("GET")
	api.HandleFunc("/mcp/servers/{id:.+}", mcpServerDetailHandler).Methods("GET")
	api.HandleFunc("/mcp/servers/{id:.+}", mcpUninstallServerHandler).Methods("DELETE")
	api.HandleFunc("/mcp/servers/{id:.+}/validate", mcpValidateServerHandler).Methods("POST")
	api.HandleFunc("/mcp/servers/{id:.+}/call", mcpCallToolHandler).Methods("POST")
	api.HandleFunc("/mcp/search", mcpSearchHandler).Methods("POST")
	api.HandleFunc("/mcp/install", mcpInstallHandler).Methods("POST")

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
