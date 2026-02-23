package api

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type ConnectParams struct {
	Token     string                 `json:"token,omitempty"`
	DeviceID  string                 `json:"deviceId,omitempty"`
	Role      string                 `json:"role,omitempty"`
	Scopes    []string               `json:"scopes,omitempty"`
	ClientInfo map[string]interface{} `json:"clientInfo,omitempty"`
}

type ChatSendParams struct {
	SessionKey   string                 `json:"sessionKey"`
	Message      string                 `json:"message"`
	Deliver      bool                   `json:"deliver"`
	IdempotencyKey string               `json:"idempotencyKey"`
	Attachments  []ChatAttachment       `json:"attachments,omitempty"`
}

type ChatAttachment struct {
	Type     string `json:"type"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
}

type ChatHistoryParams struct {
	SessionKey string `json:"sessionKey"`
	Limit      int    `json:"limit"`
}

type ChatHistoryResponse struct {
	Messages      []ChatMessage `json:"messages"`
	ThinkingLevel string        `json:"thinkingLevel,omitempty"`
}

type ChatMessage struct {
	Role      string      `json:"role"`
	Content   []ChatBlock `json:"content"`
	Timestamp int64       `json:"timestamp"`
}

type ChatBlock struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Source interface{} `json:"source,omitempty"`
}

type StatusResponse struct {
	Status    string                 `json:"status"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	Model     string                 `json:"model"`
	Tools     map[string]interface{} `json:"tools"`
	Skills    map[string]interface{} `json:"skills"`
	Channels  []string               `json:"channels"`
	Workspace string                 `json:"workspace"`
}

type ModelsListResponse struct {
	Models []ModelInfo `json:"models"`
}

type ModelInfo struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	Provider     string `json:"provider"`
	APIKeySet    bool   `json:"apiKeySet"`
}

func (s *Server) handleConnect(params interface{}) (interface{}, *ErrorInfo) {
	// For now, accept all connections (no auth required)
	return map[string]interface{}{
		"type":     "hello-ok",
		"protocol": 1,
		"auth": map[string]interface{}{
			"role":   "operator",
			"scopes": []string{"operator.admin"},
		},
	}, nil
}

func (s *Server) handleChatSend(params interface{}) (interface{}, *ErrorInfo) {
	p, ok := params.(map[string]interface{})
	if !ok {
		return nil, &ErrorInfo{
			Code:    "INVALID_PARAMS",
			Message: "Invalid parameters",
		}
	}

	sessionKey, _ := p["sessionKey"].(string)
	message, _ := p["message"].(string)
	
	if sessionKey == "" {
		sessionKey = "default"
	}

	if message == "" {
		return nil, &ErrorInfo{
			Code:    "EMPTY_MESSAGE",
			Message: "Message cannot be empty",
		}
	}

	// Create user message event first
	userEvent := map[string]interface{}{
		"runId": sessionKey,
		"sessionKey": sessionKey,
		"state":  "delta",
		"message": ChatMessage{
			Role:    "user",
			Content: []ChatBlock{{Type: "text", Text: message}},
			Timestamp: time.Now().UnixMilli(),
		},
	}
	
	// Add user event to buffer for HTTP polling
	s.addEvent("event", "chat", userEvent)

	// Process the message using the agent loop
	ctx := context.Background()
	fmt.Printf("[DEBUG] Processing message: %s, session: %s\n", message, sessionKey)
	response, err := s.agentLoop.ProcessDirectWithChannel(ctx, message, sessionKey, "cli", "direct")
	if err != nil {
		fmt.Printf("[DEBUG] Error processing message: %v\n", err)
		return nil, &ErrorInfo{
			Code:    "PROCESSING_ERROR",
			Message: fmt.Sprintf("Error processing message: %v", err),
		}
	}
	fmt.Printf("[DEBUG] Got response: %s\n", response)

	// Create chat response event
	fmt.Printf("[DEBUG] Creating assistant response event\n")
	chatEvent := map[string]interface{}{
		"runId": sessionKey,
		"sessionKey": sessionKey,
		"state":  "final",
		"message": ChatMessage{
			Role:    "assistant",
			Content: []ChatBlock{{Type: "text", Text: response}},
			Timestamp: time.Now().UnixMilli(),
		},
	}
	fmt.Printf("[DEBUG] Assistant event created: state=%s, role=%s, content_length=%d\n", 
		chatEvent["state"], chatEvent["message"].(ChatMessage).Role, len(response))
	
	// Add assistant event to buffer for HTTP polling
	fmt.Printf("[DEBUG] Adding assistant event to buffer...\n")
	s.addEvent("event", "chat", chatEvent)
	fmt.Printf("[DEBUG] Assistant event added to buffer successfully\n")
	
	// Broadcast the response to WebSocket clients
	s.broadcastToClients(map[string]interface{}{
		"type":    "event",
		"event":   "chat",
		"payload": chatEvent,
	})

	return map[string]interface{}{
		"runId": sessionKey,
		"status": "sent",
	}, nil
}

func (s *Server) handleChatHistory(params interface{}) (interface{}, *ErrorInfo) {
	p, ok := params.(map[string]interface{})
	if !ok {
		return nil, &ErrorInfo{
			Code:    "INVALID_PARAMS",
			Message: "Invalid parameters",
		}
	}

	sessionKey, _ := p["sessionKey"].(string)
	if sessionKey == "" {
		sessionKey = "default"
	}

	// Try to construct a simple chat history from recent events
	// Since we don't have persistent session storage in this implementation,
	// we'll build the history from the event buffer
	bufferMutex.RLock()
	defer bufferMutex.RUnlock()

	var messages []ChatMessage
	var latestTimestamp int64
	
	// Collect all chat events for this session, in order
	for _, event := range eventBuffer {
		if event.Event == "chat" && event.Payload != nil {
			if payload, ok := event.Payload.(map[string]interface{}); ok {
				if payloadSessionKey, exists := payload["sessionKey"].(string); exists && payloadSessionKey == sessionKey {
					if message, exists := payload["message"]; exists {
						if chatMsg, ok := message.(ChatMessage); ok {
							messages = append(messages, chatMsg)
							if chatMsg.Timestamp > latestTimestamp {
								latestTimestamp = chatMsg.Timestamp
							}
						}
					}
				}
			}
		}
	}
	
	// Sort messages by timestamp
	if len(messages) > 1 {
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp < messages[j].Timestamp
		})
	}

	return ChatHistoryResponse{
		Messages:      messages,
		ThinkingLevel: "",
	}, nil
}

func (s *Server) handleStatus(params interface{}) (interface{}, *ErrorInfo) {
	startupInfo := s.agentLoop.GetStartupInfo()
	toolsInfo, _ := startupInfo["tools"].(map[string]interface{})
	skillsInfo, _ := startupInfo["skills"].(map[string]interface{})

	return StatusResponse{
		Status:    "running",
		Uptime:    "running",
		Version:   "4cc8b90",
		Model:     s.config.Agents.Defaults.Model,
		Tools:     toolsInfo,
		Skills:    skillsInfo,
		Channels:  []string{}, // No channels enabled by default
		Workspace: s.config.WorkspacePath(),
	}, nil
}

func (s *Server) handleHealth(params interface{}) (interface{}, *ErrorInfo) {
	return map[string]interface{}{
		"status": "ok",
		"uptime": "running",
	}, nil
}

func (s *Server) handleModelsList(params interface{}) (interface{}, *ErrorInfo) {
	var models []ModelInfo
	
	// Add models from model_list configuration
	for _, mc := range s.config.ModelList {
		provider := "unknown"
		if len(mc.Model) > 0 && mc.Model[0] != '/' {
			// Extract provider from model format like "openai/gpt-4"
			for i, char := range mc.Model {
				if char == '/' {
					provider = mc.Model[:i]
					break
				}
			}
		}
		
		models = append(models, ModelInfo{
			Name:      mc.ModelName,
			Model:     mc.Model,
			Provider:  provider,
			APIKeySet: mc.APIKey != "",
		})
	}

	return ModelsListResponse{
		Models: models,
	}, nil
}

func (s *Server) handleConfigGet(params interface{}) (interface{}, *ErrorInfo) {
	// Return a simplified config for UI
	return map[string]interface{}{
		"agents": s.config.Agents,
		"model_list": s.config.ModelList,
		"tools": s.config.Tools,
		"heartbeat": s.config.Heartbeat,
	}, nil
}

func (s *Server) handleConfigSet(params interface{}) (interface{}, *ErrorInfo) {
	// TODO: Implement config setting
	return map[string]interface{}{
		"status": "ok",
		"message": "Config updated (placeholder)",
	}, nil
}

type NodeListResponse struct {
	Nodes []NodeInfo `json:"nodes"`
}

type EventsPollParams struct {
	LastSeq    int64  `json:"lastSeq,omitempty"`
	SessionID  string `json:"sessionId,omitempty"`
}

type EventInfo struct {
	Seq      int64       `json:"seq"`
	Type     string      `json:"type"`
	Event    string      `json:"event"`
	Payload  interface{} `json:"payload,omitempty"`
}

type EventsPollResponse struct {
	Events []EventInfo `json:"events"`
	HasMore bool       `json:"hasMore"`
}

var (
	eventSequence int64 = 0
	eventBuffer   []EventInfo
	bufferMutex   sync.RWMutex
)

type NodeInfo struct {
	NodeID      string   `json:"nodeId"`
	DisplayName string   `json:"displayName"`
	Connected   bool     `json:"connected"`
	Paired      bool     `json:"paired"`
	RemoteIP    string   `json:"remoteIp"`
	Version     string   `json:"version"`
	Caps        []string `json:"caps"`
	Commands    []string `json:"commands"`
}

func (s *Server) handleNodeList(params interface{}) (interface{}, *ErrorInfo) {
	// For now, return the current gateway as a single node
	// In a distributed setup, this would return multiple nodes
	nodes := []NodeInfo{
		{
			NodeID:      "gateway-1",
			DisplayName: "PicoClaw Gateway",
			Connected:   true,
			Paired:      true,
			RemoteIP:    "localhost",
			Version:     "4cc8b90",
			Caps: []string{
				"chat.send",
				"chat.history", 
				"status",
				"health",
				"models.list",
				"config.get",
				"config.set",
				"node.list",
			},
			Commands: []string{
				"execute",
				"read_file",
				"write_file",
				"list_files",
				"web_fetch",
			},
		},
	}

	return NodeListResponse{
		Nodes: nodes,
	}, nil
}

func (s *Server) addEvent(eventType, event string, payload interface{}) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	
	eventSequence++
	newEvent := EventInfo{
		Seq:     eventSequence,
		Type:    eventType,
		Event:   event,
		Payload: payload,
	}
	
	eventBuffer = append(eventBuffer, newEvent)
	
	// Keep only the last 1000 events to prevent memory issues
	if len(eventBuffer) > 1000 {
		eventBuffer = eventBuffer[len(eventBuffer)-1000:]
	}
}

func (s *Server) handleEventsPoll(params interface{}) (interface{}, *ErrorInfo) {
	p, ok := params.(map[string]interface{})
	if !ok {
		return nil, &ErrorInfo{
			Code:    "INVALID_PARAMS",
			Message: "Invalid parameters",
		}
	}
	
	lastSeq := int64(0)
	if seq, exists := p["lastSeq"].(float64); exists {
		lastSeq = int64(seq)
	}
	
	bufferMutex.RLock()
	defer bufferMutex.RUnlock()
	
	var events []EventInfo
	for _, event := range eventBuffer {
		if event.Seq > lastSeq {
			events = append(events, event)
		}
	}
	
	return EventsPollResponse{
		Events:  events,
		HasMore: false,
	}, nil
}

type AgentIdentityResponse struct {
	AgentID    string `json:"agentId"`
	DisplayName string `json:"displayName"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	Provider   string `json:"provider"`
}

func (s *Server) handleAgentIdentityGet(params interface{}) (interface{}, *ErrorInfo) {
	// Return the main agent identity
	model := s.config.Agents.Defaults.Model
	provider := "unknown"
	
	// Extract provider from model format like "openai/gpt-4"
	if len(model) > 0 && model[0] != '/' {
		for i, char := range model {
			if char == '/' {
				provider = model[:i]
				break
			}
		}
	}
	
	return AgentIdentityResponse{
		AgentID:    "main",
		DisplayName: "PicoClaw Assistant",
		Type:       "assistant",
		Status:     "active",
		Model:      model,
		Provider:   provider,
	}, nil
}

type AgentInfo struct {
	AgentID    string `json:"agentId"`
	DisplayName string `json:"displayName"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Model      string `json:"model"`
	Provider   string `json:"provider"`
	Capabilities []string `json:"capabilities"`
}

type AgentsListResponse struct {
	Agents []AgentInfo `json:"agents"`
}

func (s *Server) handleAgentsList(params interface{}) (interface{}, *ErrorInfo) {
	// Get capabilities from startup info
	startupInfo := s.agentLoop.GetStartupInfo()
	var capabilities []string
	
	if toolsInfo, ok := startupInfo["tools"].(map[string]interface{}); ok {
		for toolName := range toolsInfo {
			if toolName != "count" {
				capabilities = append(capabilities, toolName)
			}
		}
	}
	
	if skillsInfo, ok := startupInfo["skills"].(map[string]interface{}); ok {
		if available, ok := skillsInfo["available"].(float64); ok && available > 0 {
			capabilities = append(capabilities, "skills")
		}
	}
	
	// Create the main agent info
	model := s.config.Agents.Defaults.Model
	provider := "unknown"
	
	if len(model) > 0 && model[0] != '/' {
		for i, char := range model {
			if char == '/' {
				provider = model[:i]
				break
			}
		}
	}
	
	mainAgent := AgentInfo{
		AgentID:     "main",
		DisplayName: "PicoClaw Assistant",
		Type:        "assistant",
		Status:      "active",
		Model:       model,
		Provider:    provider,
		Capabilities: capabilities,
	}
	
	return AgentsListResponse{
		Agents: []AgentInfo{mainAgent},
	}, nil
}

type DevicePairListResponse struct {
	Devices []DevicePairInfo `json:"devices"`
}

type DevicePairInfo struct {
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Status     string `json:"status"`
	Paired     bool   `json:"paired"`
}

func (s *Server) handleDevicePairList(params interface{}) (interface{}, *ErrorInfo) {
	// For now, return empty device list
	return DevicePairListResponse{
		Devices: []DevicePairInfo{},
	}, nil
}

type SessionInfo struct {
	SessionKey   string `json:"sessionKey"`
	AgentID      string `json:"agentId"`
	LastActivity int64  `json:"lastActivity"`
	Active       bool   `json:"active"`
}

type SessionsListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}

func (s *Server) handleSessionsList(params interface{}) (interface{}, *ErrorInfo) {
	// Return the main session
	now := time.Now().UnixMilli()
	return SessionsListResponse{
		Sessions: []SessionInfo{
			{
				SessionKey:   "main",
				AgentID:      "main",
				LastActivity: now,
				Active:       true,
			},
		},
	}, nil
}