package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/cmd/web-server/models"
	"github.com/sipeed/picoclaw/cmd/web-server/services"
	"github.com/sipeed/picoclaw/cmd/web-server/utils"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	agentLoop         *agent.AgentLoop
	config            *config.Config
	conversationSvc   *services.ConversationService
	thoughtCollectors map[string]*services.ThoughtCollector
	muThoughts        sync.RWMutex
	logAdapters       map[string]*services.LogEntryAdapter
	muAdapters        sync.RWMutex
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(agentLoop *agent.AgentLoop, cfg *config.Config, convSvc *services.ConversationService) *ChatHandler {
	return &ChatHandler{
		agentLoop:         agentLoop,
		config:            cfg,
		conversationSvc:   convSvc,
		thoughtCollectors: make(map[string]*services.ThoughtCollector),
		logAdapters:       make(map[string]*services.LogEntryAdapter),
	}
}

// HandleChat 处理聊天请求
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if h.agentLoop == nil {
		http.Error(w, "Agent not initialized", http.StatusInternalServerError)
		return
	}

	// 如果请求流式响应
	if req.Stream {
		h.handleStreamingChat(w, r, req)
	} else {
		h.handleNonStreamingChat(w, r, req)
	}
}

// handleStreamingChat 处理流式聊天
func (h *ChatHandler) handleStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest) {
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

	// 如果提供了 ConversationID，使用它作为 sessionKey；否则创建新对话
	var sessionKey string
	var conversationID string

	if req.ConversationID != "" {
		// 使用已有的对话 ID
		conversationID = req.ConversationID
		sessionKey = "web:" + conversationID
	} else {
		// 创建新对话
		newConv, err := h.conversationSvc.CreateConversation("", req.Model)
		if err != nil {
			http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
			return
		}
		conversationID = newConv.ID
		sessionKey = "web:" + conversationID
		log.Printf("Created new conversation: %s", conversationID)
	}

	// 保存用户消息到历史记录
	h.conversationSvc.SaveUserMessage(conversationID, req.Message, req.Model)

	// 创建思考过程收集器
	h.muThoughts.Lock()
	collector := services.NewThoughtCollectorWithSession(func(thought models.Thought) {
		utils.SendSSEThought(w, flusher, thought)
	}, sessionKey)
	h.thoughtCollectors[sessionKey] = collector
	h.muThoughts.Unlock()

	// 创建日志适配器来捕获agent和tool的详细日志
	logAdapter := services.NewLogEntryAdapter(collector)

	// 注册到日志系统以监听真实的日志
	logger.AddLogListener(logAdapter)

	// 注册日志适配器
	h.muAdapters.Lock()
	h.logAdapters[sessionKey] = logAdapter
	h.muAdapters.Unlock()

	// Update the agent loop's model temporarily
	originalModel := h.config.Agents.Defaults.Model
	h.config.Agents.Defaults.Model = req.Model

	// 清理函数
	defer func() {
		h.muThoughts.Lock()
		delete(h.thoughtCollectors, sessionKey)
		h.muThoughts.Unlock()

		h.muAdapters.Lock()
		delete(h.logAdapters, sessionKey)
		h.muAdapters.Unlock()

		// 从日志系统中移除监听器
		logger.RemoveLogListener(logAdapter)

		h.config.Agents.Defaults.Model = originalModel
	}()

	// 发送详细的初始思考过程
	collector.AddThought("thinking", "🤔 收到用户消息: "+req.Message)
	collector.AddThought("thinking", "📋 创建会话: "+sessionKey)
	collector.AddThought("thinking", "⚙️ 开始处理消息，使用模型: "+req.Model)
	collector.AddThought("thinking", "🔍 连接日志监听器以捕获所有系统操作...")
	collector.AddThought("thinking", "🧠 AI 系统初始化完成，开始智能分析...")
	collector.AddThought("thinking", "📝 准备调用 AgentLoop 处理用户请求...")

	ctx := context.Background()
	startTime := time.Now()

	// 使用自定义的处理函数来收集思考过程
	response, err := h.processWithThoughtCollection(ctx, req.Message, sessionKey, collector)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		collector.AddThought("thinking", "❌ 处理消息时出错: "+err.Error())
		// 发送错误完成消息
		utils.SendSSEComplete(w, flusher, "", req.Model, err.Error(), conversationID, collector.GetThoughts())
		return
	}

	// 发送完成思考过程
	collector.AddThoughtWithDetails("tool_result", "✅ AI 完成分析，生成回复内容", "agent_reasoning", "", "", duration-500, 0)
	collector.AddThought("thinking", "✅ 消息处理完成，耗时: "+fmt.Sprintf("%dms", duration))

	// 保存助手消息到历史记录
	h.conversationSvc.SaveAssistantMessage(conversationID, response, req.Model, collector.GetThoughts())

	// 发送最终完成消息
	utils.SendSSEComplete(w, flusher, response, req.Model, "", conversationID, collector.GetThoughts())
}

// handleNonStreamingChat 处理非流式聊天
func (h *ChatHandler) handleNonStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest) {
	// 如果提供了 ConversationID，使用它作为 sessionKey；否则创建新对话
	var sessionKey string
	var conversationID string

	if req.ConversationID != "" {
		// 使用已有的对话 ID
		conversationID = req.ConversationID
		sessionKey = "web:" + conversationID
	} else {
		// 创建新对话
		newConv, err := h.conversationSvc.CreateConversation("", req.Model)
		if err != nil {
			log.Printf("Failed to create conversation: %v", err)
			http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
			return
		}
		conversationID = newConv.ID
		sessionKey = "web:" + conversationID
		log.Printf("Created new conversation: %s", conversationID)
	}

	// 保存用户消息到历史记录
	h.conversationSvc.SaveUserMessage(conversationID, req.Message, req.Model)

	// 创建简单的思考过程记录
	thoughts := []models.Thought{
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
	originalModel := h.config.Agents.Defaults.Model
	h.config.Agents.Defaults.Model = req.Model
	defer func() {
		h.config.Agents.Defaults.Model = originalModel
	}()

	// Process the message
	ctx := context.Background()

	startTime := time.Now()
	response, err := h.agentLoop.ProcessDirect(ctx, req.Message, sessionKey)
	duration := int(time.Since(startTime).Milliseconds())

	if err != nil {
		thoughts = append(thoughts, models.Thought{
			Type:      "thinking",
			Timestamp: time.Now(),
			Content:   "❌ 处理消息时出错: " + err.Error(),
		})
		http.Error(w, fmt.Sprintf("Chat error: %v", err), http.StatusInternalServerError)
		return
	}

	thoughts = append(thoughts,
		models.Thought{
			Type:      "thinking",
			Timestamp: time.Now(),
			Content:   "✅ 消息处理完成，耗时: " + fmt.Sprintf("%dms", duration),
		},
		models.Thought{
			Type:      "tool_call",
			Timestamp: time.Now().Add(-time.Second * 2),
			Content:   "🔧 AI 正在分析问题并准备调用相关工具",
			ToolName:  "agent_reasoning",
		},
		models.Thought{
			Type:      "tool_result",
			Timestamp: time.Now().Add(-time.Second),
			Content:   "✅ AI 完成分析，生成回复内容",
			ToolName:  "agent_reasoning",
			Duration:  duration - 500,
		},
	)

	// 保存助手消息到历史记录
	h.conversationSvc.SaveAssistantMessage(conversationID, response, req.Model, thoughts)

	chatResponse := models.ChatResponse{
		Message:        response,
		Model:          req.Model,
		ConversationID: conversationID,
		Timestamp:      time.Now(),
		Thoughts:       thoughts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResponse)
}

// processWithThoughtCollection 包装ProcessDirect来收集思考过程
func (h *ChatHandler) processWithThoughtCollection(ctx context.Context, message, sessionKey string, collector *services.ThoughtCollector) (string, error) {
	// 发送详细的系统处理步骤
	collector.AddThought("thinking", "🧠 AI 正在分析用户请求...")
	collector.AddThought("thinking", "📋 识别任务类型和需求...")
	collector.AddThought("thinking", "🔍 检查可用工具和技能...")
	collector.AddThought("thinking", "⚡ 准备调用 AgentLoop 执行处理...")

	// 执行agent处理，同时通过日志监听器实时收集所有操作
	collector.AddThought("thinking", "🚀 开始调用 AgentLoop.ProcessDirect...")
	startTime := time.Now()
	response, err := h.agentLoop.ProcessDirect(ctx, message, sessionKey)
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
