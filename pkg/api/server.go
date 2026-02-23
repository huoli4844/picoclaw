package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
)

type Server struct {
	server      *http.Server
	agentLoop   *agent.AgentLoop
	msgBus      *bus.MessageBus
	config      *config.Config
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
}

type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     string      `json:"id"`
}

type Response struct {
	ID      string      `json:"id"`
	OK      bool        `json:"ok"`
	Payload interface{} `json:"payload,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type HelloOK struct {
	Type      string      `json:"type"`
	Protocol  int         `json:"protocol"`
	Features  interface{} `json:"features,omitempty"`
	Snapshot  interface{} `json:"snapshot,omitempty"`
	Auth      interface{} `json:"auth,omitempty"`
	Policy    interface{} `json:"policy,omitempty"`
}

func NewServer(host string, port int, agentLoop *agent.AgentLoop, msgBus *bus.MessageBus, cfg *config.Config) *Server {
	mux := http.NewServeMux()
	s := &Server{
		agentLoop: agentLoop,
		msgBus:    msgBus,
		config:    cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}

	// Health endpoints
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readyHandler)
	
	// WebSocket endpoint for OpenClaw compatibility
	mux.HandleFunc("/gateway", s.websocketHandler)
	
	// REST API endpoints  
	mux.HandleFunc("/api", s.apiHandler)

	addr := fmt.Sprintf("%s:%d", host, port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("🚀 PicoClaw API Server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"uptime": "running",
	})
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready", 
		"uptime": "running",
	})
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()
	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
	}()

	log.Printf("WebSocket client connected from %s", r.RemoteAddr)

	// Send hello response
	hello := HelloOK{
		Type:     "hello-ok",
		Protocol: 1,
		Features: map[string]interface{}{
			"methods": []string{"chat.send", "chat.history", "status", "health"},
			"events":  []string{"chat.delta", "chat.done"},
		},
	}
	
	if err := conn.WriteJSON(hello); err != nil {
		log.Printf("Error sending hello: %v", err)
		return
	}

	// Handle WebSocket messages
	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		resp := s.handleRequest(req)
		if err := conn.WriteJSON(resp); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

func (s *Server) apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := s.handleRequest(req)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRequest(req Request) Response {
	var payload interface{}
	var err *ErrorInfo

	switch req.Method {
	case "connect":
		payload, err = s.handleConnect(req.Params)
	case "chat.send":
		payload, err = s.handleChatSend(req.Params)
	case "chat.history":
		payload, err = s.handleChatHistory(req.Params)
	case "status":
		payload, err = s.handleStatus(req.Params)
	case "health":
		payload, err = s.handleHealth(req.Params)
	case "models.list":
		payload, err = s.handleModelsList(req.Params)
	case "config.get":
		payload, err = s.handleConfigGet(req.Params)
	case "config.set":
		payload, err = s.handleConfigSet(req.Params)
	case "node.list":
		payload, err = s.handleNodeList(req.Params)
	case "events.poll":
		payload, err = s.handleEventsPoll(req.Params)
	case "agent.identity.get":
		payload, err = s.handleAgentIdentityGet(req.Params)
	case "agents.list":
		payload, err = s.handleAgentsList(req.Params)
	case "device.pair.list":
		payload, err = s.handleDevicePairList(req.Params)
	case "sessions.list":
		payload, err = s.handleSessionsList(req.Params)
	default:
		err = &ErrorInfo{
			Code:    "METHOD_NOT_FOUND",
			Message: fmt.Sprintf("Method '%s' not found", req.Method),
		}
	}

	if err != nil {
		return Response{
			ID:    req.ID,
			OK:    false,
			Error: err,
		}
	}

	return Response{
		ID:      req.ID,
		OK:      true,
		Payload: payload,
	}
}

func (s *Server) broadcastToClients(message interface{}) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	
	for conn := range s.clients {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// corsMiddleware adds CORS headers to all responses
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}