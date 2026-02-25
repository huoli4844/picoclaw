package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sipeed/picoclaw/pkg/mcp"
)

// MCPHandler MCP处理器
type MCPHandler struct {
	mcpRegistry *mcp.Registry
}

// NewMCPHandler 创建MCP处理器
func NewMCPHandler(registry *mcp.Registry) *MCPHandler {
	return &MCPHandler{
		mcpRegistry: registry,
	}
}

// GetServers 获取MCP服务器列表
func (h *MCPHandler) GetServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	servers, err := h.mcpRegistry.GetInstalledServers()
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

// GetServerDetail 获取服务器详情
func (h *MCPHandler) GetServerDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	server, err := h.mcpRegistry.GetServer(serverID)
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

// ValidateServer 验证服务器
func (h *MCPHandler) ValidateServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	response, err := h.mcpRegistry.ValidateInstallation(serverID)
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

// UninstallServer 卸载服务器
func (h *MCPHandler) UninstallServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	serverID := vars["id"]

	err := h.mcpRegistry.UninstallServer(serverID)
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

// CallTool 调用工具
func (h *MCPHandler) CallTool(w http.ResponseWriter, r *http.Request) {
	log.Printf("MCP工具调用请求开始")
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
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
	server, err := h.mcpRegistry.GetServer(serverID)
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
				"timestamp":    h.getCurrentTimestamp(),
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
				"timestamp":    h.getCurrentTimestamp(),
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
				"timestamp":    h.getCurrentTimestamp(),
				"isSimulation": false,
			}
			json.NewEncoder(w).Encode(result)
			return
		}

		log.Printf("工具调用成功，结果: %+v", toolResult)

		// 格式化结果并发送响应
		result := h.formatToolResult(serverID, req.ToolName, req.Arguments, toolResult, true)
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
			"timestamp":    h.getCurrentTimestamp(),
			"isSimulation": false,
		}
		json.NewEncoder(w).Encode(result)
		return
	}
	log.Printf("工具调用成功，结果: %+v", toolResult)

	// 格式化结果并发送响应
	result := h.formatToolResult(serverID, req.ToolName, req.Arguments, toolResult, false)
	json.NewEncoder(w).Encode(result)
}

// SearchServers 搜索服务器
func (h *MCPHandler) SearchServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
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

	response, err := h.mcpRegistry.SearchServers(req)
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

// GetSources 获取MCP来源
func (h *MCPHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sources, err := mcp.GetAvailableMCPSources()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": true,
		"data":    sources,
	}

	json.NewEncoder(w).Encode(result)
}

// InstallServer 安装服务器
func (h *MCPHandler) InstallServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.mcpRegistry == nil {
		http.Error(w, "MCP registry not initialized", http.StatusServiceUnavailable)
		return
	}

	var req mcp.MCPInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.mcpRegistry.InstallServer(req)
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

// formatToolResult 格式化工具结果
func (h *MCPHandler) formatToolResult(serverID, toolName string, arguments map[string]interface{}, toolResult *mcp.ToolCallResult, isSimulation bool) map[string]interface{} {
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

	return map[string]interface{}{
		"success":      !toolResult.IsError,
		"serverID":     serverID,
		"toolName":     toolName,
		"arguments":    arguments,
		"result":       strings.TrimSpace(resultText),
		"timestamp":    fmt.Sprintf("%v", h.getCurrentTimestamp()),
		"isSimulation": isSimulation,
		"content":      toolResult.Content,
	}
}

// getCurrentTimestamp 获取当前时间戳
func (h *MCPHandler) getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
